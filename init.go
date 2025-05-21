package fssync

import (
	"context"
	"errors"
	"fmt"

	"github.com/Gleipnir-Technology/arcgis-go/fieldseeker"
)

var config *Config

func ensureConfig() error {
	var err error
	config, err = ReadConfig()
	if err != nil {
		return fmt.Errorf("Failed to read config: %v", err)
	}
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

	err = ConnectDB(context.Background(), config.Database.URL)
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
	fieldseeker.Initialize(
		config.Arcgis.ServiceRoot,
		config.Arcgis.TenantID,
		config.Arcgis.Token,
		config.Arcgis.FieldSeekerService)

	if err != nil {
		return fmt.Errorf("Failed to initialize fieldseeker: %v", err)
	}

	return nil
}
