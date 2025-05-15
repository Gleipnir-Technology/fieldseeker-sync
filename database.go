package main

import (
	"context"
	"database/sql"
	"embed"
	"fmt"
	"os"
	"sync"

	//"github.com/Gleipnir-Technology/arcgis-go"
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

func doMigrations(connection_string string) error {
	fmt.Println("Connecting to database at", connection_string)
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

// func saveOrUpdateDBRecords(qr *arcgis.QueryResult) (error) {
func saveOrUpdateDBRecords(qr []byte) error {
	output, err := os.Create("records/service-records.json")
	if err != nil {
		return err
	}
	defer output.Close()
	b, err := output.Write(qr)
	if err != nil {
		return err
	}
	fmt.Printf("Wrote %v bytes\n", b)
	return nil
}

/*
func (pg *postgres) InsertServiceRequest(ctx context.Context, feature ) error {
	query := `INSERT INTO service_request (address, city, notes_field, notes_tech, target) VALUES (@address, @city, @notes_field, @notes_tech, @target)`
	args := pgx.NamedArgs{
		"address": feature.Attributes["REQADDR1"],
		"city": feature.Attributes["REQCITY"],
		"notes_field": feature.Attributes["REQFLDNOTES"],
		"notes_tech": feature.Attributes["REQNOTESFORTECH"],
		"target": feature.Attributes["REQTARGET"],
	}
	_, err := pg.db.Exec(context.Background(), query, args)
	if err != nil {
		fmt.Println("Failed insert: %w", err)
	}
}*/
