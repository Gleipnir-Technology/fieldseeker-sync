package main

import (
	"context"
	"fmt"
	"os"

	"github.com/Gleipnir-Technology/arcgis-go"
	"github.com/Gleipnir-Technology/arcgis-go/fieldseeker"
	"gleipnir.technology/fieldseeker-sync-bridge"
)

func main() {
	config, err := fssync.ReadConfig()
	if err != nil {
		fmt.Println("Failed to read config: ", err)
		os.Exit(1)
	}

	if len(config.Database.URL) == 0 {
		fmt.Println("You must specify a database URL")
		os.Exit(2)
	}
	err = fssync.ConnectDB(context.Background(), config.Database.URL)
	if err != nil {
		fmt.Println("Failed to initialize connection: ", err)
		os.Exit(3)
	}

	ag := arcgis.ArcGIS{
		config.Arcgis.ServiceRoot,
		config.Arcgis.TenantID,
		config.Arcgis.Token}
	fmt.Println("Connecting to FieldSeeker at ", ag)
	fs := fieldseeker.NewFieldSeeker(&ag, config.Arcgis.FieldSeekerService)
	err = fs.EnsureHasServiceInfo()
	if err != nil {
		fmt.Println("Failed to get FieldSeeker service info:", err)
		os.Exit(4)
	}
	for _, layer := range fs.FeatureServer.Layers {
		err := saveSchema(fs, layer)
		if err != nil {
			fmt.Println("Failed: ", err)
			os.Exit(5)
		}
	}
}

func saveSchema(fs *fieldseeker.FieldSeeker, layer arcgis.Layer) error {
	fmt.Printf("Layer %v named '%v'\n", layer.ID, layer.Name)
	output, err := os.Create(fmt.Sprintf("schema/%v.json", layer.Name))
	if err != nil {
		return err
	}

	query := arcgis.NewQuery()
	query.ResultRecordCount = 1
	query.ResultOffset = 0
	query.OutFields = "*"
	query.Where = "1=1"
	qr, err := fs.Arcgis.QueryRaw(
		fs.ServiceName,
		layer.ID,
		query)
	if err != nil {
		return err
	}

	_, err = output.Write(qr)
	if err != nil {
		return err
	}

	return nil
}
