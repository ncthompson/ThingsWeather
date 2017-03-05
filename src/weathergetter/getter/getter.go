package main

import (
	"flag"
	"log"
	"os"
	"weathergetter/configuration"
	"weathergetter/influxif"
	"weathergetter/thingsif"
)

func main() {
	createTemplate := flag.Bool("template", false, "Create sample configuration template.")
	configFile := flag.String("config", "config.json", "Configuration file location.")
	flag.Parse()
	if *createTemplate {
		err := configuration.CreateConfigTemplate()
		if err != nil {
			log.Fatalf("Configuration creation error: %v.\n", err)
		}
		os.Exit(0)
	}

	config, err := configuration.OpenConfig(*configFile)
	if err != nil {
		log.Fatalf("Failed to open configuraion: %v.\n", err)
	}

	mqtt, err := thingsif.InitialiseMqttClient(config.MConfig)
	if err != nil {
		log.Fatalf("Failed to start MQTT client: %v\n", err)
	}
	inf, err := influxif.InitialiseInfluxClient(config.DbConfig)
	if err != nil {
		log.Fatalf("Failed to start Influxdb client: %v\n", err)
	}

	hist, err := mqtt.GetLast7days()

	inf.SyncDatabase(hist)

	for {
		nodeData, err := mqtt.WaitForData()
		if err != nil {
			log.Printf("Error: %v", err)
		} else {
			log.Printf("Time: %v\n", nodeData.Metadata.Time)
			log.Printf("Temperature: %v\n", nodeData.Payload_fields.Temp)
			log.Printf("Humidity: %v\n", nodeData.Payload_fields.Humd)
			log.Printf("Battery: %v\n", nodeData.Payload_fields.Bat)
			thingsif.PrintGatways(nodeData.Metadata.Gateways)
			err := inf.WriteToDatabase(nodeData)
			if err != nil {
				log.Printf("Batch point error: %v\n", err)
			}
		}
	}
}
