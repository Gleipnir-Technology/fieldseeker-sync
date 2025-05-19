package fssync

import (
	"context"
	"database/sql"
	"embed"
	"errors"
	"fmt"
	"log"
	"sort"
	"strings"
	"sync"

	"github.com/Gleipnir-Technology/arcgis-go"
	"github.com/georgysavva/scany/v2/pgxscan"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/pressly/goose/v3"
)

type postgres struct {
	db *pgxpool.Pool
}

var (
	pgInstance *postgres
	pgOnce     sync.Once
)

//go:embed migrations/*.sql
var embedMigrations embed.FS

func ConnectDB(ctx context.Context, connection_string string) error {
	err := doMigrations(connection_string)
	if err != nil {
		return err
	}

	err = nil
	pgOnce.Do(func() {
		db, e := pgxpool.New(ctx, connection_string)
		pgInstance = &postgres{db}
		err = e
	})
	if err != nil {
		return fmt.Errorf("unable to create connection pool: %w", err)
	}

	return nil
}

func SaveOrUpdateDBRecords(ctx context.Context, table string, qr *arcgis.QueryResult) error {
	query := upsertFromQueryResult(table, qr)
	batch := &pgx.Batch{}
	for _, f := range qr.Features {
		args := pgx.NamedArgs{}
		for k, v := range f.Attributes {
			args[k] = v
		}
		// specially add geometry since it isn't in the list of attributes
		args["geometry_x"] = f.Geometry.X
		args["geometry_y"] = f.Geometry.Y
		batch.Queue(query, args)
	}
	results := pgInstance.db.SendBatch(ctx, batch)
	defer results.Close()

	for _, f := range qr.Features {
		_, err := results.Exec()
		if err != nil {
			/*var pgErr *pgconn.PgError
			if errors.As(err, &pgErr) && pgErr.Code == pgerrcode.UniqueViolation {
				fmt.Printf("Object %s already exists\n", f.Attributes["OBJECTID"])
				continue
			} else {
				fmt.Println("Failed to upsert: ", err)
			}*/
			fmt.Println("Error on exec: ", err)
			fmt.Println("Bad row: ", f)
		}
	}
	return results.Close()
}

func ServiceRequestCount() (int, error) {
	if pgInstance == nil {
		return 0, errors.New("You must initialize the DB first")
	}

	var count int
	err := pgInstance.db.QueryRow(context.Background(), "SELECT COUNT(*) FROM FS_SERVICEREQUEST").Scan(&count)
	if err != nil {
		return 0, err
	}
	return count, nil
}

func ServiceRequests() ([]*ServiceRequest, error) {

	if pgInstance == nil {
		return make([]*ServiceRequest, 0), errors.New("You must initialize the DB first")
	}

	rows, _ := pgInstance.db.Query(context.Background(), "SELECT GEOMETRY_X AS \"geometry.X\",GEOMETRY_Y AS \"geometry.Y\",PRIORITY,REQADDR1,REQCITY,REQTARGET,REQZIP,STATUS,SOURCE FROM FS_ServiceRequest")
	var requests []*ServiceRequest

	if err := pgxscan.ScanAll(&requests, rows); err != nil {
		log.Println("CollectRows error:", err)
		return make([]*ServiceRequest, 0), err
	}

	return requests, nil
}

func doMigrations(connection_string string) error {
	log.Println("Connecting to database at", connection_string)
	db, err := sql.Open("pgx", connection_string)
	if err != nil {
		return fmt.Errorf("Failed to open database connection: %w", err)
	}
	defer db.Close()

	goose.SetBaseFS(embedMigrations)

	if err := goose.SetDialect("postgres"); err != nil {
		return fmt.Errorf("Failed to select dialect: %w", err)
	}

	if err := goose.Up(db, "migrations"); err != nil {
		return fmt.Errorf("Failed to run migrations: %w", err)
	}
	return nil
}

// Generate a query for upsert from a QueryResult
func upsertFromQueryResult(table string, qr *arcgis.QueryResult) string {
	// Make the rows appear in a deterministic order
	sorted_columns := make([]string, 0, len(qr.Fields))
	for _, f := range qr.Fields {
		sorted_columns = append(sorted_columns, f.Name)
	}
	sort.Strings(sorted_columns)

	var sb strings.Builder
	sb.WriteString("INSERT INTO ")
	sb.WriteString(table)
	sb.WriteString(" (")
	for _, field := range sorted_columns {
		sb.WriteString(field)
		sb.WriteString(",")
	}
	// Specially add the geometry values since they aren't in the fields
	sb.WriteString("geometry_x,geometry_y")
	sb.WriteString(")\nVALUES (")
	for _, field := range sorted_columns {
		sb.WriteString("@")
		sb.WriteString(field)
		sb.WriteString(",")
	}
	// Specially add the geometry values since they aren't in the fields
	sb.WriteString("@geometry_x,@geometry_y")
	sb.WriteString(")\nON CONFLICT(")
	sb.WriteString(qr.UniqueIdField.Name)
	sb.WriteString(")\nDO UPDATE SET\n")
	for _, field := range qr.Fields {
		// skip the unique field since we can't set it again
		if field.Name == qr.UniqueIdField.Name {
			continue
		}
		sb.WriteString(" ")
		sb.WriteString(field.Name)
		sb.WriteString(" = EXCLUDED.")
		sb.WriteString(field.Name)
		sb.WriteString(",\n")
	}
	// Specially add the geometry values since they aren't in the fields
	sb.WriteString(" geometry_x = EXCLUDED.geometry_x,\n geometry_y = EXCLUDED.geometry_y\n;")
	return sb.String()
}
