package fssync

import (
	"context"
	"errors"
	"fmt"

	"github.com/Gleipnir-Technology/arcgis-go/fieldseeker"
	"github.com/Gleipnir-Technology/fieldseeker-sync/database"
)

var config *Config

func ensureConfig() error {
	config = ReadConfig()
	if len(config.Database.URL) == 0 {
		return errors.New("You must specify a database URL")
	}
	return nil
}

func InitDB() error {
	err := ensureConfig()
	if err != nil {
		return err
	}

	err = database.ConnectDB(context.Background(), config.Database.URL)
	if err != nil {
		return fmt.Errorf("Failed to initialize connection: %v", err)
	}
	return nil
}

func Initialize() error {
	err := ensureConfig()
	if err != nil {
		return err
	}

	err = InitDB()
	if err != nil {
		return err
	}
	err = fieldseeker.Initialize(
		config.Arcgis.ServiceRoot,
		config.Arcgis.TenantID,
		config.Arcgis.Token,
		config.Arcgis.FieldSeekerService)

	if err != nil {
		return fmt.Errorf("Failed to initialize fieldseeker: %v", err)
	}

	return nil
}
