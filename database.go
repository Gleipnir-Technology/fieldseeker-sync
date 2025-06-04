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

type DBQuery struct {
	Bounds Bounds
	Limit  int
}

func NewQuery() DBQuery {
	return DBQuery{
		Bounds: NewBounds(),
		Limit:  0,
	}
}

func MosquitoSourceQuery(q *DBQuery) ([]MosquitoSource, error) {
	results := make([]MosquitoSource, 0)
	if pgInstance == nil {
		return results, errors.New("You must initialize the DB first")
	}
	args, query := prepQuery(q, "SELECT GEOMETRY_X AS \"geometry.X\",GEOMETRY_Y AS \"geometry.Y\",creationdate,name,habitat,usetype,waterorigin,description,accessdesc,comments,globalid FROM FS_PointLocation WHERE GEOMETRY_X > @west AND GEOMETRY_X < @east AND GEOMETRY_Y > @south AND GEOMETRY_Y < @north")

	rows, _ := pgInstance.db.Query(context.Background(), query, args)
	var locations []*FS_PointLocation

	if err := pgxscan.ScanAll(&locations, rows); err != nil {
		log.Println("CollectRows on FS_PointLocation error:", err)
		return results, err
	}

	location_by_globalid := make(map[string]*FS_PointLocation, len(locations))
	globalids := make([]string, 0)
	for _, l := range locations {
		location_by_globalid[l.GlobalID] = l
		globalids = append(globalids, l.GlobalID)
	}
	args = pgx.NamedArgs{
		"globalids": globalids,
	}
	rows, _ = pgInstance.db.Query(context.Background(), "SELECT comments,enddatetime,globalid,sitecond,pointlocid FROM FS_MosquitoInspection WHERE pointlocid=ANY(@globalids)", args)
	var inspections []*FS_MosquitoInspection

	if err := pgxscan.ScanAll(&inspections, rows); err != nil {
		log.Println("CollectRows on FS_MosquitoInspection error:", err)
		return results, err
	}

	inspections_by_locid := make(map[string][]*FS_MosquitoInspection)
	for _, i := range inspections {
		x := inspections_by_locid[i.PointLocationID]
		x = append(x, i)
		inspections_by_locid[i.PointLocationID] = x
	}

	rows, _ = pgInstance.db.Query(context.Background(), "SELECT comments,enddatetime,globalid,habitat,product,qty,qtyunit,sitecond,treatacres,treathectares,pointlocid FROM FS_Treatment WHERE pointlocid=ANY(@globalids)", args)
	var treatments []*FS_Treatment

	if err := pgxscan.ScanAll(&treatments, rows); err != nil {
		log.Println("CollectRows on FS_Treatment error:", err)
		return results, err
	}
	treatments_by_locid := make(map[string][]*FS_Treatment)
	for _, i := range treatments {
		x := treatments_by_locid[i.PointLocationID]
		x = append(x, i)
		treatments_by_locid[i.PointLocationID] = x
	}
	for _, pl := range locations {
		results = append(results, NewMosquitoSource(pl, inspections_by_locid[pl.GlobalID], treatments_by_locid[pl.GlobalID]))
	}
	return results, nil
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
		return fmt.Errorf("Unable to insert row into user_: %v", err)
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

func ServiceRequestQuery(q *DBQuery) ([]ServiceRequest, error) {
	results := make([]ServiceRequest, 0)
	if pgInstance == nil {
		return results, errors.New("You must initialize the DB first")
	}

	args, query := prepQuery(q, "SELECT GEOMETRY_X AS \"geometry.X\",GEOMETRY_Y AS \"geometry.Y\",CreationDate,globalid,PRIORITY,REQADDR1,REQCITY,REQTARGET,REQZIP,STATUS,SOURCE FROM FS_ServiceRequest WHERE GEOMETRY_X > @west AND GEOMETRY_X < @east AND GEOMETRY_Y > @south AND GEOMETRY_Y < @north")
	rows, _ := pgInstance.db.Query(context.Background(), query, args)
	var fs_service_requests []*FS_ServiceRequest

	if err := pgxscan.ScanAll(&fs_service_requests, rows); err != nil {
		log.Println("CollectRows error:", err)
		return results, err
	}
	for _, r := range fs_service_requests {
		results = append(results, ServiceRequest{data: r})
	}

	return results, nil
}

func TrapDataQuery(q *DBQuery) ([]TrapData, error) {
	results := make([]TrapData, 0)
	if pgInstance == nil {
		return results, errors.New("You must initialize the DB first")
	}

	log.Println("Getting FS_TrapLocation")
	args, query := prepQuery(q, "SELECT geometry_x AS \"geometry.X\",geometry_y AS \"geometry.Y\",creationdate,globalid,name,description,accessdesc,objectid FROM FS_TrapLocation WHERE geometry_x > @west AND geometry_x < @east AND geometry_y > @south AND geometry_y < @north")
	rows, _ := pgInstance.db.Query(context.Background(), query, args)
	var fs_trap_locations []*FS_TrapLocation

	if err := pgxscan.ScanAll(&fs_trap_locations, rows); err != nil {
		log.Println("CollectRows error:", err)
		return results, err
	}
	log.Println("Found FS_TrapLocation", len(fs_trap_locations))
	for _, l := range fs_trap_locations {
		results = append(results, TrapData{data: l})
	}

	return results, nil
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

// Given a database query and predicate produce named args and a full text DB query
func prepQuery(q *DBQuery, predicate string) (pgx.NamedArgs, string) {
	args := pgx.NamedArgs{
		"east":  q.Bounds.East,
		"north": q.Bounds.North,
		"south": q.Bounds.South,
		"west":  q.Bounds.West,
	}
	query := predicate
	if q.Limit > 0 {
		args["limit"] = q.Limit
		query = query + " LIMIT @limit"
	}
	return args, query
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
