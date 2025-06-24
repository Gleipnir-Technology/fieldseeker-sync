package main

import (
	"context"
	"flag"
	"log"
	"os"

	"github.com/Gleipnir-Technology/arcgis-go"
	"github.com/Gleipnir-Technology/arcgis-go/fieldseeker"
	"gleipnir.technology/fieldseeker-sync"
)

func main() {
	offset := flag.Int("offset", 0, "where to start in the set of records")
	flag.Parse()
	layers := flag.Args()

	err := fssync.Initialize()
	if err != nil {
		log.Println("Failed to initialize: ", err)
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
			log.Println("Layer is not valid", l)
			os.Exit(2)
		}
	}
	inserts := 0
	updates := 0
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
		i, u, err := downloadAllRecords(layer, *offset)
		if err != nil {
			log.Println("Failed: ", err)
			os.Exit(5)
		}
		inserts += i
		updates += u
	}
	log.Printf("Run complete. Total inserts %d updates %d\n", inserts, updates)
}

func downloadAllRecords(layer arcgis.Layer, offset int) (int, int, error) {
	//log.Printf("%v %v\n", layer.ID, layer.Name)
	inserts := 0
	updates := 0
	count, err := fieldseeker.QueryCount(layer.ID)
	if err != nil {
		return inserts, updates, err
	}
	log.Printf("Need to get %v records for layer '%v'\n", count.Count, layer.Name)
	if count.Count == 0 {
		//log.Printf("No records available\n")
		return inserts, updates, nil
	}
	for {
		if offset >= count.Count {
			//log.Printf("Offset is at %v/%v records. Stopping.\n", offset, count.Count)
			break
		}
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
			log.Println("Failure:", err)
			os.Exit(6)
		}
		i, u, err := fssync.SaveOrUpdateDBRecords(context.Background(), "FS_"+layer.Name, qr)
		if err != nil {
			log.Println("Failed to save records:", err)
			saveRawQuery(layer, query, "temp/failure.json")
			os.Exit(7)
		}
		inserts += i
		updates += u
		offset += len(qr.Features)
		//log.Printf("Handled %v %v records. Offset %v. %v remain\n", len(qr.Features), layer.Name, offset, count.Count-offset)
	}
	log.Printf("%d inserts, %d updates, %d no change\n", inserts, updates, count.Count-inserts-updates)
	return inserts, updates, nil
}

func saveRawQuery(layer arcgis.Layer, query *arcgis.Query, filename string) {
	output, err := os.Create(filename)
	if err != nil {
		log.Println("Failed to open", filename, err)
		return
	}
	qr, err := fieldseeker.DoQueryRaw(
		layer.ID,
		query)
	if err != nil {
		log.Println("Failed to do query", err)
		return
	}

	_, err = output.Write(qr)
	if err != nil {
		log.Println("Failed to write results", err)
		return
	}
	log.Println("Wrote failed query to", filename)
}
