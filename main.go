package main

import (
	"context"
	"fmt"
	"os"

	"github.com/Gleipnir-Technology/arcgis-go"
	"github.com/Gleipnir-Technology/arcgis-go/fieldseeker"
)

func main() {
	config, err := ReadConfig()
	if err != nil {
		fmt.Println("Failed to read config: ", err)
		os.Exit(1)
	}

	if len(config.Database.URL) == 0 {
		fmt.Println("You must specify a database URL")
		os.Exit(2)
	}
	err = ConnectDB(context.Background(), config.Database.URL)
	if err != nil {
		fmt.Println("Failed to initialize connection: ", err)
	}

	ag := arcgis.ArcGIS{
		config.Arcgis.ServiceRoot,
		config.Arcgis.TenantID,
		config.Arcgis.Token}
	fs := fieldseeker.NewFieldSeeker(&ag, "foo")
	fs.EnsureHasServiceInfo()
	for _, layer := range fs.FeatureServer.Layers {
		err := downloadAllRecords(fs, layer)
		if err != nil {
			fmt.Println("Failed: %v", err)
			os.Exit(3)
		}
	}
}

func downloadAllRecords(fs *fieldseeker.FieldSeeker, layer arcgis.Layer) error {
	fmt.Printf("%v %v\n", layer.ID, layer.Name)
	count, err := fs.Arcgis.QueryCount(fs.ServiceName, layer.ID)
	if err != nil {
		return err
	}
	fmt.Printf("Need to get %v records\n", count.Count)

	if err != nil {
		return err
	}
	offset := 0
	for {
		query := arcgis.NewQuery()
		query.ResultRecordCount = fs.FeatureServer.MaxRecordCount
		query.ResultOffset = offset
		query.OutFields = "*"
		query.Where = "1=1"
		qr, err := fs.Arcgis.QueryRaw(
			fs.ServiceName,
			layer.ID,
			query)
		if err != nil {
			fmt.Printf("Failure: %v", err)
			os.Exit(1)
		}

		err = saveOrUpdateDBRecords(qr)
		if err != nil {
			os.Exit(2)
		}
		os.Exit(0)
		offset += query.ResultRecordCount
		if offset > count.Count {
			break
		}
	}

	return nil
}
