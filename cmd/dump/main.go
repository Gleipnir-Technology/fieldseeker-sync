package main

import (
	"flag"
	"log"
	"os"

	"gleipnir.technology/fieldseeker-sync"
)

func main() {
	limit := flag.Int("limit", 10, "limit the number of objects")
	flag.Parse()
	types := flag.Args()

	err := fssync.InitDB()
	if err != nil {
		log.Println("Failed to initialize DB: ", err)
		os.Exit(1)
	}
	if len(types) == 0 {
		log.Println("Please include at least one type")
		os.Exit(2)
	}

	for _, type_ := range types {
		if type_ == "trapdata" {
			query := fssync.NewQuery()
			trapdata, err := fssync.TrapDataQuery(&query)
			if err != nil {
				log.Println(err)
				os.Exit(3)
			}
			log.Println("Total trap datas", len(trapdata))
			for _, trap := range trapdata {
				log.Println(trap)
			}
		} else if type_ == "mosquitosource" {
			query := fssync.NewQuery()
			query.Limit = *limit
			sources, err := fssync.MosquitoSourceQuery(&query)
			if err != nil {
				log.Println(err)
				os.Exit(4)
			}
			log.Println("Total sources", len(sources))
			for _, s := range sources {
				log.Println("Access: ", s.Access())
				log.Println("Comments: ", s.Comments())
				log.Println("Description: ", s.Description())
				log.Println("Location: ", s.Location().Latitude(), s.Location().Longitude())
				log.Println("Habitat: ", s.Habitat())
				log.Println("Name: ", s.Name())
				log.Println("UseType: ", s.UseType())
				log.Println("WaterOrigin: ", s.WaterOrigin())
				for _, i := range s.Inspections {
					log.Println("  Condition: ", i.Condition(), " Created: ", i.Created().String())
				}
				log.Println("========")
			}
		} else {
			log.Println("Unrecognized type", type_)
			continue
		}
	}
}
