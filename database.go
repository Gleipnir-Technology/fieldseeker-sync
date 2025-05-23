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
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/pressly/goose/v3"
)

type postgres struct {
	db *pgxpool.Pool
}

type Bounds struct {
	East  float64
	North float64
	South float64
	West  float64
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

	var version string
	query := `SELECT version()`
	err = pgInstance.db.QueryRow(context.Background(), query).Scan(&version)
	if err != nil {
		return fmt.Errorf("Failed to get database version: %w", err)
	}
	log.Println("Connected to", version)
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
		batch.Queue(query, args).Exec(func(ct pgconn.CommandTag) error {
			if ct.Update() {
				// log.Println("Update", f.Attributes[qr.UniqueIdField.Name])
			} else if ct.Insert() {
				// log.Println("Insert", f.Attributes[qr.UniqueIdField.Name])
			} else {
				log.Println("No idea what happened here")
			}
			return nil
		})
	}
	results := pgInstance.db.SendBatch(ctx, batch)

	return results.Close()
}

func SaveUser(displayname string, hash string, username string) error {
	log.Println("Saving new user")
	query := `INSERT INTO user_ (display_name, password_hash_type, password_hash, username) VALUES (@display_name, @password_hash_type, @password_hash, @username)`
	args := pgx.NamedArgs{
		"display_name":       displayname,
		"password_hash_type": "bcrypt-14",
		"password_hash":      hash,
		"username":           username,
	}
	row, err := pgInstance.db.Exec(context.Background(), query, args)
	if err != nil {
		return fmt.Errorf("Unable to insert row into user_", err)
	}
	log.Println("Saved user", username, row)
	return nil
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

func ServiceRequests(b Bounds) ([]*ServiceRequest, error) {
	if pgInstance == nil {
		return make([]*ServiceRequest, 0), errors.New("You must initialize the DB first")
	}

	args := pgx.NamedArgs{
		"east":  b.East,
		"north": b.North,
		"south": b.South,
		"west":  b.West,
	}
	rows, _ := pgInstance.db.Query(context.Background(), "SELECT GEOMETRY_X AS \"geometry.X\",GEOMETRY_Y AS \"geometry.Y\",PRIORITY,REQADDR1,REQCITY,REQTARGET,REQZIP,STATUS,SOURCE FROM FS_ServiceRequest WHERE GEOMETRY_X > @west AND GEOMETRY_X < @east AND GEOMETRY_Y > @south AND GEOMETRY_Y < @north", args)
	var requests []*ServiceRequest

	if err := pgxscan.ScanAll(&requests, rows); err != nil {
		log.Println("CollectRows error:", err)
		return make([]*ServiceRequest, 0), err
	}

	return requests, nil
}

func ValidateUser(username string, password string) (*User, error) {
	var (
		display_name string
		hash         string
	)
	query := `SELECT display_name,password_hash FROM user_ WHERE username=$1`
	err := pgInstance.db.QueryRow(context.Background(), query, username).Scan(&display_name, &hash)
	if err != nil {
		return nil, err
	}
	if !VerifyPassword(password, hash) {
		return nil, nil
	}
	return &User{
		DisplayName: display_name,
		Username:    username,
	}, nil
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
