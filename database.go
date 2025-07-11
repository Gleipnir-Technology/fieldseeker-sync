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
	args, query := prepQuery(q, "SELECT GEOMETRY_X AS \"geometry.X\",GEOMETRY_Y AS \"geometry.Y\",accessdesc,active,comments,creationdate,description,habitat,lastinspectdate,name,nextactiondatescheduled,usetype,waterorigin,zone,globalid FROM FS_PointLocation WHERE GEOMETRY_X > @west AND GEOMETRY_X < @east AND GEOMETRY_Y > @south AND GEOMETRY_Y < @north")

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
	rows, _ = pgInstance.db.Query(context.Background(), "SELECT actiontaken,comments,enddatetime,fieldtech,globalid,locationname,pointlocid,sitecond,zone FROM FS_MosquitoInspection WHERE pointlocid=ANY(@globalids)", args)
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

	rows, _ = pgInstance.db.Query(context.Background(), "SELECT comments,enddatetime,fieldtech,globalid,habitat,product,qty,qtyunit,sitecond,treatacres,treathectares,pointlocid FROM FS_Treatment WHERE pointlocid=ANY(@globalids)", args)
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

// Pretty big function. Given a set of query results we're going to iterate over each of them.
// For each one, if the row doesn't exist, we create a row. If it does exist, we check to see
// if its already correctly represented. If it isn't, we add a new version.
func SaveOrUpdateDBRecords(ctx context.Context, table string, qr *arcgis.QueryResult) (int, int, error) {
	inserts, updates := 0, 0
	// Get the current state of every row for our current query result
	sorted_columns := make([]string, 0, len(qr.Fields))
	for _, f := range qr.Fields {
		sorted_columns = append(sorted_columns, f.Name)
	}
	sort.Strings(sorted_columns)

	objectids := make([]int, 0)
	for _, l := range qr.Features {
		oid := l.Attributes["OBJECTID"].(float64)
		objectids = append(objectids, int(oid))
	}

	rows_by_objectid, err := rowmapViaQuery(ctx, table, sorted_columns, objectids)
	if err != nil {
		return inserts, updates, fmt.Errorf("Failed to get existing rows: %v", err)
	}
	// log.Println("Rows from query", len(rows_by_objectid))

	for _, feature := range qr.Features {
		oid := feature.Attributes["OBJECTID"].(float64)
		row := rows_by_objectid[int(oid)]
		// If we have no matching row we'll need to create it
		if len(row) == 0 {

			if err := insertRowFromFeature(ctx, table, sorted_columns, &feature); err != nil {
				return inserts, updates, fmt.Errorf("Failed to insert row: %v", err)
			}
			inserts += 1
		} else if hasUpdates(row, feature) {
			if err := updateRowFromFeature(ctx, table, sorted_columns, &feature); err != nil {
				return inserts, updates, fmt.Errorf("Failed to update row: %v", err)
			}
			updates += 1
		}
	}
	return inserts, updates, nil
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

	args, query := prepQuery(q, "SELECT GEOMETRY_X AS \"geometry.X\",GEOMETRY_Y AS \"geometry.Y\",ASSIGNEDTECH,CreationDate,DOG,globalid,PRIORITY,REQADDR1,REQCITY,RECDATETIME,REQTARGET,REQZIP,SOURCE,Spanish,STATUS FROM FS_ServiceRequest WHERE GEOMETRY_X > @west AND GEOMETRY_X < @east AND GEOMETRY_Y > @south AND GEOMETRY_Y < @north")
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

func hasUpdates(row map[string]string, feature arcgis.Feature) bool {
	for key, value := range feature.Attributes {
		rowdata := row[strings.ToLower(key)]
		// We'll accept any 'nil' as represented by the empty string in the database
		if value == nil && rowdata == "" {
			continue
		}
		// check strings first, their simplest
		if featureAsString, ok := value.(string); ok {
			if featureAsString != rowdata {
				return true
			}
			continue
		} else if featureAsInt, ok := value.(int); ok {
			// Previously had a nil value, now we have a real value
			if rowdata == "" {
				return true
			}
			rowAsInt, err := strconv.Atoi(rowdata)
			if err != nil {
				log.Fatal(fmt.Sprintf("Failed to convert '%s' to an int to compare against %v for %v", rowdata, featureAsInt, key))
			}
			if rowAsInt != featureAsInt {
				return true
			} else {
				continue
			}
		} else if featureAsFloat, ok := value.(float64); ok {
			// Previously had a nil value, now we have a real value
			if rowdata == "" {
				return true
			}
			rowAsFloat, err := strconv.ParseFloat(rowdata, 64)
			if err != nil {
				log.Fatal(fmt.Sprintf("Failed to convert '%s' to a float64 to compare against %v for %v", rowdata, featureAsFloat, key))
			}
			if rowAsFloat != featureAsFloat {
				return true
			} else {
				continue
			}
		}
		log.Printf("Type: %T\tkey: %s\tvalue: %v\trow: %s\n", value, key, value, rowdata)
		log.Fatal("Need type update.")
	}
	return false
}

func insertRowFromFeatureFS(ctx context.Context, transaction pgx.Tx, table string, sorted_columns []string, feature *arcgis.Feature) error {
	// Create the query to produce the main row
	var sb strings.Builder
	sb.WriteString("INSERT INTO ")
	sb.WriteString(table)
	sb.WriteString(" (")
	for _, field := range sorted_columns {
		sb.WriteString(field)
		sb.WriteString(",")
	}
	// Specially add the geometry values since they aren't in the fields
	sb.WriteString("geometry_x,geometry_y,updated")
	sb.WriteString(")\nVALUES (")
	for _, field := range sorted_columns {
		sb.WriteString("@")
		sb.WriteString(field)
		sb.WriteString(",")
	}
	// Specially add the geometry values since they aren't in the fields
	sb.WriteString("@geometry_x,@geometry_y,@updated)")

	args := pgx.NamedArgs{}
	for k, v := range feature.Attributes {
		args[k] = v
	}
	// specially add geometry since it isn't in the list of attributes
	args["geometry_x"] = feature.Geometry.X
	args["geometry_y"] = feature.Geometry.Y
	args["updated"] = time.Now()

	_, err := transaction.Exec(ctx, sb.String(), args)
	if err != nil {
		return fmt.Errorf("Failed to insert row into %s: %v", table, err)
	}
	return nil
}

func insertRowFromFeatureHistory(ctx context.Context, transaction pgx.Tx, table string, sorted_columns []string, feature *arcgis.Feature, version int) error {
	history_table := toHistoryTable(table)
	var sb strings.Builder
	sb.WriteString("INSERT INTO ")
	sb.WriteString(history_table)
	sb.WriteString(" (")
	for _, field := range sorted_columns {
		sb.WriteString(field)
		sb.WriteString(",")
	}
	// Specially add the geometry values since they aren't in the fields
	sb.WriteString("created,geometry_x,geometry_y,version")
	sb.WriteString(")\nVALUES (")
	for _, field := range sorted_columns {
		sb.WriteString("@")
		sb.WriteString(field)
		sb.WriteString(",")
	}
	// Specially add the geometry values since they aren't in the fields
	sb.WriteString("@created,@geometry_x,@geometry_y,@version)")
	args := pgx.NamedArgs{}
	for k, v := range feature.Attributes {
		args[k] = v
	}
	args["created"] = time.Now()
	args["version"] = version
	if _, err := transaction.Exec(ctx, sb.String(), args); err != nil {
		return fmt.Errorf("Failed to insert history row into %s: %v", table, err)
	}
	return nil
}

func insertRowFromFeature(ctx context.Context, table string, sorted_columns []string, feature *arcgis.Feature) error {
	var options pgx.TxOptions
	transaction, err := pgInstance.db.BeginTx(ctx, options)
	if err != nil {
		return fmt.Errorf("Unable to start transaction")
	}

	err = insertRowFromFeatureFS(ctx, transaction, table, sorted_columns, feature)
	if err != nil {
		return fmt.Errorf("Unable to insert FS: %v", err)
	}

	err = insertRowFromFeatureHistory(ctx, transaction, table, sorted_columns, feature, 1)
	if err != nil {
		return fmt.Errorf("Failed to insert history: %v", err)
	}

	err = transaction.Commit(ctx)
	if err != nil {
		return fmt.Errorf("Failed to commit transaction: %v", err)
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

type StringScanner struct {
}

// Produces a map of OBJECTID to a 'row' which is in turn a map of column names to their values as strings
func rowmapViaQuery(ctx context.Context, table string, sorted_columns []string, objectids []int) (map[int]map[string]string, error) {
	result := make(map[int]map[string]string)

	query := selectAllFromQueryResult(table, sorted_columns)

	args := pgx.NamedArgs{
		"objectids": objectids,
	}
	rows, err := pgInstance.db.Query(ctx, query, args)
	if err != nil {
		return result, fmt.Errorf("Failed to query rows: %v", err)
	}
	defer rows.Close()

	// +2 for geometry x and geometry x
	columnNames := make([]string, len(sorted_columns)+2)
	for i, c := range sorted_columns {
		columnNames[i] = c
	}
	columnNames[len(sorted_columns)] = "geometry_x"
	columnNames[len(sorted_columns)+1] = "geometry_y"

	rowSlice, err := pgx.CollectRows(rows, func(row pgx.CollectableRow) (map[string]string, error) {
		fieldDescriptions := row.FieldDescriptions()
		values := make([]interface{}, len(fieldDescriptions))
		valuePtrs := make([]interface{}, len(fieldDescriptions))

		for i := range values {
			valuePtrs[i] = &values[i]
		}

		if err := row.Scan(valuePtrs...); err != nil {
			return nil, err
		}

		result := make(map[string]string)
		for i, fd := range fieldDescriptions {
			if values[i] != nil {
				result[fd.Name] = fmt.Sprintf("%v", values[i])
				//log.Printf("col %v type %T val %v", fd.Name, values[i], values[i])
			} else {
				result[fd.Name] = ""
			}
		}

		return result, nil
	})
	if err != nil {
		return result, fmt.Errorf("Failed to collect rows: %v", err)
	}
	for _, row := range rowSlice {
		o := row["objectid"]
		objectid, err := strconv.Atoi(o)
		if err != nil {
			return result, fmt.Errorf("Failed to parse objectid %s: %v", o, err)
		}
		result[objectid] = row
	}
	return result, nil
}

// Generate a query to get all columns from a QueryResult
func selectAllFromQueryResult(table string, sorted_columns []string) string {
	var sb strings.Builder
	sb.WriteString("SELECT * FROM ")
	sb.WriteString(table)
	sb.WriteString(" WHERE OBJECTID=ANY(@objectids)")
	return sb.String()
}

func toHistoryTable(table string) string {
	return "History_" + table[3:len(table)]
}

func updateRowFromFeature(ctx context.Context, table string, sorted_columns []string, feature *arcgis.Feature) error {
	// Get the current highest version for the row in question
	history_table := toHistoryTable(table)
	var sb strings.Builder
	sb.WriteString("SELECT MAX(version) FROM ")
	sb.WriteString(history_table)
	sb.WriteString(" WHERE OBJECTID=@objectid")

	args := pgx.NamedArgs{}
	o := feature.Attributes["OBJECTID"].(float64)
	args["objectid"] = int(o)

	var version int
	if err := pgInstance.db.QueryRow(ctx, sb.String(), args).Scan(&version); err != nil {
		return fmt.Errorf("Failed to query for version: %v", err)
	}

	var options pgx.TxOptions
	transaction, err := pgInstance.db.BeginTx(ctx, options)
	if err != nil {
		return fmt.Errorf("Unable to start transaction")
	}

	err = insertRowFromFeatureHistory(ctx, transaction, table, sorted_columns, feature, version+1)
	if err != nil {
		return fmt.Errorf("Failed to insert history: %v", err)
	}
	err = updateRowFromFeatureFS(ctx, transaction, table, sorted_columns, feature)
	if err != nil {
		return fmt.Errorf("Failed to update row from feature: %v", err)
	}

	err = transaction.Commit(ctx)
	if err != nil {
		return fmt.Errorf("Failed to commit transaction: %v", err)
	}
	return nil
}

func updateRowFromFeatureFS(ctx context.Context, transaction pgx.Tx, table string, sorted_columns []string, feature *arcgis.Feature) error {
	// Create the query to produce the main row
	var sb strings.Builder
	sb.WriteString("UPDATE ")
	sb.WriteString(table)
	sb.WriteString(" SET ")
	for _, field := range sorted_columns {
		// OBJECTID is special as our primary key, so skip it
		if field == "OBJECTID" {
			continue
		}
		sb.WriteString(field)
		sb.WriteString("=@")
		sb.WriteString(field)
		sb.WriteString(",")
	}
	// Specially add the geometry values since they aren't in the fields
	sb.WriteString("geometry_x=@geometry_x,geometry_y=@geometry_y,updated=@updated WHERE OBJECTID=@OBJECTID")

	args := pgx.NamedArgs{}
	for k, v := range feature.Attributes {
		args[k] = v
	}
	// specially add geometry since it isn't in the list of attributes
	args["geometry_x"] = feature.Geometry.X
	args["geometry_y"] = feature.Geometry.Y
	args["updated"] = time.Now()

	_, err := transaction.Exec(ctx, sb.String(), args)
	if err != nil {
		return fmt.Errorf("Failed to update row into %s: %v", table, err)
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
