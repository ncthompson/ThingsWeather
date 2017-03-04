package main

import (
	"flag"
	"fmt"
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
			fmt.Printf("Configuration creation error: %v.\n", err)
		}
		return
	}

	config, err := configuration.OpenConfig(*configFile)
	if err != nil {
		fmt.Printf("Failed to open configuraion: %v.\n", err)
		return
	}

	mqtt, err := thingsif.InitialiseMqttClient(config.MConfig)
	if err != nil {
		panic(err)
	}
	inf, err := influxif.InitialiseInfluxClient(config.DbConfig)
	if err != nil {
		panic(err)
	}

	for {
		nodeData, err := mqtt.WaitForData()
		if err != nil {
			fmt.Printf("Error: %v", err)
		} else {
			fmt.Printf("Time: %v\n", nodeData.Metadata.Time)
			fmt.Printf("Temperature: %v\n", nodeData.Payload_fields.Temp)
			fmt.Printf("Humidity: %v\n", nodeData.Payload_fields.Humd)
			fmt.Printf("Battery: %v\n", nodeData.Payload_fields.Bat)
			thingsif.PrintGatways(nodeData.Metadata.Gateways)
			err := inf.WriteToDatabase(nodeData)
			if err != nil {
				fmt.Printf("Batch point error: %v\n", err)
			}
		}
	}
}
