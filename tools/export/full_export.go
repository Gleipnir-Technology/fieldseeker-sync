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
	err := fssync.Initialize()
	if err != nil {
		fmt.Println("Failed to initialize: ", err)
		os.Exit(1)
	}
	for _, layer := range fieldseeker.FeatureServerLayers() {
		err := downloadAllRecords(layer)
		if err != nil {
			fmt.Println("Failed: ", err)
			os.Exit(5)
		}
	}
}

func downloadAllRecords(layer arcgis.Layer) error {
	fmt.Printf("%v %v\n", layer.ID, layer.Name)
	count, err := fieldseeker.QueryCount(layer.ID)
	if err != nil {
		return err
	}
	fmt.Printf("Need to get %v records\n", count.Count)
	if count.Count == 0 {
		return nil
	}
	if err != nil {
		return err
	}
	offset := 0
	for {
		query := arcgis.NewQuery()
		query.ResultRecordCount = fieldseeker.MaxRecordCount()
		query.ResultOffset = offset
		query.OutFields = "*"
		query.Where = "1=1"
		qr, err := fieldseeker.DoQuery(
			layer.ID,
			query)
		if err != nil {
			fmt.Printf("Failure: %v", err)
			os.Exit(6)
		}

		err = fssync.SaveOrUpdateDBRecords(context.Background(), "FS_"+layer.Name, qr)
		if err != nil {
			os.Exit(7)
		}
		offset += len(qr.Features)
		fmt.Printf("Got %v %v records. %v remain\n", len(qr.Features), layer.Name, count.Count-offset)
		if offset > count.Count {
			break
		}
	}

	return nil
}
