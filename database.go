package fssync

import (
	"context"
	"database/sql"
	"embed"
	"errors"
	"fmt"
	"log"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/Gleipnir-Technology/arcgis-go"
	"github.com/georgysavva/scany/v2/pgxscan"
	"github.com/google/uuid"
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

func MosquitoSourceQuery(query DBQuery) ([]*MosquitoSource, error) {
	results := make([]*MosquitoSource, 0)
	if pgInstance == nil {
		return results, errors.New("You must initialize the DB first")
	}

	args := pgx.NamedArgs{
		"east":  query.Bounds.East,
		"north": query.Bounds.North,
		"south": query.Bounds.South,
		"west":  query.Bounds.West,
	}
	q := "SELECT GEOMETRY_X AS \"geometry.X\",GEOMETRY_Y AS \"geometry.Y\",name,habitat,usetype,waterorigin,description,accessdesc,comments,globalid FROM FS_PointLocation WHERE GEOMETRY_X > @west AND GEOMETRY_X < @east AND GEOMETRY_Y > @south AND GEOMETRY_Y < @north"
	if query.Limit > 0 {
		args["limit"] = query.Limit
		q = q + " LIMIT @limit"
	}
	log.Println("Searching mosquito source bounds west: ", query.Bounds.West, " east:", query.Bounds.East, " south:", query.Bounds.South, " north:", query.Bounds.North)

	rows, _ := pgInstance.db.Query(context.Background(), q, args)
	var locations []*FS_PointLocation

	if err := pgxscan.ScanAll(&locations, rows); err != nil {
		log.Println("CollectRows on FS_PointLocation error:", err)
		return results, err
	}

	globalids := make([]string, len(locations))
	for _, l := range locations {
		globalids = append(globalids, l.GlobalID)
	}
	args = pgx.NamedArgs{
		"globalids": globalids,
	}
	rows, _ = pgInstance.db.Query(context.Background(), "SELECT comments,enddatetime,sitecond,pointlocid FROM FS_MosquitoInspection WHERE pointlocid=ANY(@globalids)", args)
	var inspections []*FS_MosquitoInspection

	if err := pgxscan.ScanAll(&inspections, rows); err != nil {
		log.Println("CollectRows on FS_MosquitoInspection error:", err)
		return results, err
	}

	// Collect all the data into our final result structure
	inspection_by_id := make(map[string][]MosquitoInspection, len(locations))
	for _, mi := range inspections {
		group := inspection_by_id[mi.PointLocationID]
		created_epoch, err := strconv.ParseInt(mi.EndDateTime, 10, 64)
		if err != nil {
			log.Println("Unable to convert timestamp", mi.EndDateTime, err)
			continue
		}
		created := time.UnixMilli(created_epoch)
		group = append(group, MosquitoInspection{
			Comments:  mi.Comments,
			Condition: mi.Condition,
			Created:   created,
		})
		inspection_by_id[mi.PointLocationID] = group
	}
	for _, pl := range locations {
		results = append(results, &MosquitoSource{
			Access:      pl.Access,
			Comments:    pl.Comments,
			Description: pl.Description,
			Location:    pl.Geometry.asLatLong(),
			Habitat:     pl.Habitat,
			Inspections: inspection_by_id[pl.GlobalID],
			Name:        pl.Name,
			UseType:     pl.UseType,
			WaterOrigin: pl.WaterOrigin,
		})
	}
	return results, nil
}

func NoteQuery() ([]Note, error) {
	return []Note{
		{
			Category: "entry",
			Content:  "Gate code 123",
			Created:  time.Date(2025, time.March, 10, 23, 0, 0, 0, time.UTC),
			ID:       uuid.MustParse("2012b322-b753-41e7-b5fb-2d6556a162d0"),
			Location: LatLong{
				Latitude:  33.0687195,
				Longitude: -110.8019039,
			},
		},
		{
			Category: "info",
			Content:  "Just so you know",
			ID:       uuid.MustParse("55a52c09-67d9-4c0d-9dc7-f492f08a60ed"),
			Created:  time.Date(2025, time.April, 17, 3, 4, 5, 0, time.UTC),
			Location: LatLong{
				Latitude:  33.2667195,
				Longitude: -111.8209039,
			},
		},
		{
			Category: "todo",
			Content:  "Check water",
			ID:       uuid.MustParse("9187d1f1-a9d3-48d4-b654-98c8e933df28"),
			Location: LatLong{
				Latitude:  33.8012,
				Longitude: -112.031,
			},
		},
		{
			Category: "todo",
			Content:  "Spray treatment",
			ID:       uuid.MustParse("a1ee1d92-f783-4303-857f-a152163d6e98"),
			Location: LatLong{
				Latitude:  33.4420,
				Longitude: -111.613,
			},
		},
	}, nil
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

func ServiceRequestQuery(query *DBQuery) ([]*ServiceRequest, error) {
	if pgInstance == nil {
		return make([]*ServiceRequest, 0), errors.New("You must initialize the DB first")
	}

	args := pgx.NamedArgs{
		"east":  query.Bounds.East,
		"north": query.Bounds.North,
		"south": query.Bounds.South,
		"west":  query.Bounds.West,
	}
	rows, _ := pgInstance.db.Query(context.Background(), "SELECT GEOMETRY_X AS \"geometry.X\",GEOMETRY_Y AS \"geometry.Y\",PRIORITY,REQADDR1,REQCITY,REQTARGET,REQZIP,STATUS,SOURCE FROM FS_ServiceRequest WHERE GEOMETRY_X > @west AND GEOMETRY_X < @east AND GEOMETRY_Y > @south AND GEOMETRY_Y < @north", args)
	var requests []*ServiceRequest

	if err := pgxscan.ScanAll(&requests, rows); err != nil {
		log.Println("CollectRows error:", err)
		return make([]*ServiceRequest, 0), err
	}

	return requests, nil
}

func TrapDataQuery(query *DBQuery) ([]*TrapData, error) {
	if pgInstance == nil {
		return make([]*TrapData, 0), errors.New("You must initialize the DB first")
	}

	args := pgx.NamedArgs{
		"east":  query.Bounds.East,
		"north": query.Bounds.North,
		"south": query.Bounds.South,
		"west":  query.Bounds.West,
	}
	rows, _ := pgInstance.db.Query(context.Background(), "SELECT geometry_x AS \"geometry.X\",geometry_y AS \"geometry.Y\",name,description,accessdesc,objectid,globalid FROM FS_TrapLocation WHERE geometry_x > @west AND geometry_x < @east AND geometry_y > @south AND geometry_y < @north", args)
	var fs_trap_locations []*FS_TrapLocation

	if err := pgxscan.ScanAll(&fs_trap_locations, rows); err != nil {
		log.Println("CollectRows error:", err)
		return make([]*TrapData, 0), err
	}
	var traps []*TrapData
	for _, l := range fs_trap_locations {
		traps = append(traps, &TrapData{
			Geometry:    l.Geometry,
			Access:      l.Access,
			Description: l.Description,
			Name:        l.Name,
		})
	}

	return traps, nil
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
