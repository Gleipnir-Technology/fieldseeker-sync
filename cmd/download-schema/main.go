package main

import (
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/Gleipnir-Technology/arcgis-go"
	"github.com/Gleipnir-Technology/arcgis-go/fieldseeker"
	"github.com/Gleipnir-Technology/fieldseeker-sync"
)

func main() {
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
		err := saveSchema(layer)
		if err != nil {
			fmt.Println("Failed: ", err)
			os.Exit(5)
		}
	}
}

func saveSchema(layer arcgis.Layer) error {
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
	qr, err := fieldseeker.DoQueryRaw(
		layer.ID, query)
	if err != nil {
		return err
	}

	_, err = output.Write(qr)
	if err != nil {
		return err
	}

	return nil
}
