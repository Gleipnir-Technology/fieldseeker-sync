package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/Gleipnir-Technology/arcgis-go"
	"github.com/Gleipnir-Technology/arcgis-go/fieldseeker"
	"gleipnir.technology/fieldseeker-sync-bridge"
)

func main() {
	flag.Parse()
	layers := flag.Args()

	err := fssync.Initialize()
	if err != nil {
		fmt.Println("Failed to initialize: ", err)
		os.Exit(1)
	}
	// Check that we specified the layers correctly
	for _, l := range layers {
		is_valid := false
		for _, layer := range fieldseeker.FeatureServerLayers() {
			if l == layer.Name {
				is_valid = true
				break
			}
		}
		if !is_valid {
			fmt.Println("Layer is not valid", l)
			os.Exit(2)
		}
	}
	for _, layer := range fieldseeker.FeatureServerLayers() {
		if len(layers) > 0 {
			is_selected := false
			for _, l := range layers {
				if layer.Name == l {
					is_selected = true
					break
				}
			}
			if !is_selected {
				continue
			}
		}
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
	log.Printf("Need to get %v records\n", count.Count)
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
		query.SpatialReference = "4326"
		query.OutFields = "*"
		query.Where = "1=1"
		qr, err := fieldseeker.DoQuery(
			layer.ID,
			query)
		if err != nil {
			fmt.Printf("Failure: %v", err)
			os.Exit(6)
		}
		//for _, r := range qr.Features {
		//log.Println(r.Attributes["OBJECTID"])
		//}
		err = fssync.SaveOrUpdateDBRecords(context.Background(), "FS_"+layer.Name, qr)
		if err != nil {
			os.Exit(7)
		}
		offset += len(qr.Features)
		log.Printf("Got %v %v records. %v remain\n", len(qr.Features), layer.Name, count.Count-offset)
		if offset >= count.Count {
			break
		}
	}

	return nil
}
