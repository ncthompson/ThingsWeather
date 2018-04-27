package main

import (
	"flag"
	sysd "github.com/coreos/go-systemd/daemon"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"
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
	go func() {
		for {
			nodeData, err := mqtt.WaitForData()
			if err != nil {
				log.Printf("Error: %v", err)
			} else {
				if nodeData != nil && nodeData.Payload_fields != nil {
					if nodeData.Payload_fields.Valid {
						log.Printf("Node: %v\n", nodeData.DevId)
						log.Printf("Time: %v\n", nodeData.Metadata.Time)
						log.Printf("Temperature: %v\n", nodeData.Payload_fields.Temp)
						log.Printf("Humidity: %v\n", nodeData.Payload_fields.Humd)
						log.Printf("Battery: %v\n", nodeData.Payload_fields.Bat)
						log.Printf("Rain: %v\n", nodeData.Payload_fields.Rain)
						log.Printf("Pressure: %v\n", nodeData.Payload_fields.Pres)
						thingsif.PrintGatways(nodeData.Metadata.Gateways)
						err := inf.WriteToDatabase(nodeData)
						if err != nil {
							log.Printf("Batch point error: %v\n", err)
						}
					} else {
						log.Printf("Invalid Gateway")
						log.Printf("Time: %v\n", nodeData.Metadata.Time)
						log.Printf("Temperature: %v\n", nodeData.Payload_fields.Temp)
						log.Printf("Humidity: %v\n", nodeData.Payload_fields.Humd)
						log.Printf("Battery: %v\n", nodeData.Payload_fields.Bat)
						log.Printf("Rain: %v\n", nodeData.Payload_fields.Rain)
					}
				}
			}
		}
	}()
	sysd.SdNotify(false, "READY=1")

	termChan := make(chan os.Signal, 1)
	signal.Notify(termChan, syscall.SIGTERM, syscall.SIGINT)
	<-termChan
	go func() {
		time.Sleep(10 * time.Second)
		panic("Unclean shutdown.")
	}()
	mqtt.Close()
	inf.Close()
	log.Print("Graceful shutdown.")
	os.Exit(0)
}
