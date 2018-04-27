package main

import (
	"flag"
	"log"
	"time"
	"weathergetter/configuration"
	"weathergetter/influxif"
	"weathergetter/stbsource"
)

func main() {
	configFile := flag.String("config", "config.json", "Configuration file location.")
	flag.Parse()

	config, err := configuration.OpenConfig(*configFile)
	if err != nil {
		log.Fatalf("Failed to open configuraion: %v.\n", err)
	}

	inf, err := influxif.InitialiseInfluxClient(config.DbConfig)
	if err != nil {
		log.Fatalf("Failed to start Influxdb client: %v\n", err)
	}

	for {
		measure, err := stbsource.GetStbSource()
		if err != nil {
			log.Printf("ERROR: %v\n", err)
		} else {
			err = inf.WriteStbToDatabase(measure)
			if err != nil {
				log.Printf("ERROR: %v\n", err)
			}
		}
		time.Sleep(30 * time.Second)
	}
}
