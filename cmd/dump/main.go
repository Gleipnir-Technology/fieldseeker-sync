package main

import (
	"flag"
	"log"
	"os"

	"gleipnir.technology/fieldseeker-sync-bridge"
)

func main() {
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
			bounds := fssync.NewBounds()
			trapdata, err := fssync.TrapDataQuery(&bounds)
			if err != nil {
				log.Println(err)
				os.Exit(2)
			}
			log.Println("Total trap datas", len(trapdata))
			for _, trap := range trapdata {
				log.Println(trap)
			}
		} else {
			log.Println("Unrecognized type", type_)
			continue
		}
	}
}
