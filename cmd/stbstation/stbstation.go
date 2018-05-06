package main

import (
	"configuration"
	"flag"
	"interfaces/influxif"
	"interfaces/stbsource"
	"log"
	"time"
)

func main() {
	configFile := flag.String("config", "config.json", "Configuration file location.")
	updateRate := flag.Int("rate", 30, "Set the update rate in seconds")
	flag.Parse()

	config, err := configuration.OpenConfig(*configFile)
	if err != nil {
		log.Fatalf("Failed to open configuraion: %v.\n", err)
	}

	inf, err := influxif.InitialiseInfluxClient(config.DbConfig)
	if err != nil {
		log.Fatalf("Failed to start Influxdb client: %v\n", err)
	}
	ticker := time.NewTicker(time.Duration(*updateRate) * time.Second)
	for range ticker.C {
		measure, err := stbsource.GetStbSource()
		if err != nil {
			log.Printf("ERROR: %v\n", err)
		} else {
			err = inf.WriteStbToDatabase(measure)
			if err != nil {
				log.Printf("ERROR: %v\n", err)
			}
		}
	}
}
