package main

import (
	"flag"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	sysd "github.com/coreos/go-systemd/daemon"
	"github.com/ncthompson/ThingsWeather/configuration"
	"github.com/ncthompson/ThingsWeather/interfaces/influxif"
	"github.com/ncthompson/ThingsWeather/interfaces/thingsif"
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
		log.Fatalf("Failed to open configuration: %v.\n", err)
	}

	mqtt, err := thingsif.NewClient(config.MConfig)
	if err != nil {
		log.Fatalf("Failed to start MQTT client: %v\n", err)
	}
	inf, err := influxif.NewClient(config.DbConfig)
	if err != nil {
		log.Fatalf("Failed to start Influxdb client: %v\n", err)
	}

	hist, err := mqtt.GetLast7days()
	if err != nil {
		log.Printf("Could not sync old data: %v", err)
	} else {
		_ = inf.SyncDatabase(hist)
	}

	go updater(mqtt, inf)
	_, err = sysd.SdNotify(false, "READY=1")
	if err != nil {
		log.Printf("Could not signal systemd: %v", err)
	}

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

func updater(mqtt *thingsif.MQTTCli, inf *influxif.InfluxIf) {
	for {
		var nodeData *thingsif.Message
		var err error
		nodeData, err = mqtt.WaitForData()
		if err != nil {
			log.Printf("Error: %v", err)
			continue
		}
		if nodeData != nil && nodeData.PayloadFields != nil {
			if nodeData.PayloadFields.Valid {
				log.Printf("Node: %v\n", nodeData.DevID)
				log.Printf("Time: %v\n", nodeData.Metadata.Time)
				log.Printf("Temperature: %v\n", nodeData.PayloadFields.Temp)
				log.Printf("Humidity: %v\n", nodeData.PayloadFields.Humid)
				log.Printf("Battery: %v\n", nodeData.PayloadFields.Bat)
				log.Printf("Rain: %v\n", nodeData.PayloadFields.Rain)
				log.Printf("Pressure: %v\n", nodeData.PayloadFields.Pres)
				thingsif.PrintGatways(nodeData.Metadata.Gateways)
				err = inf.WriteToDatabase(nodeData)
				if err != nil {
					log.Printf("Batch point error: %v\n", err)
				}
			} else {
				log.Printf("Invalid Gateway")
				log.Printf("Time: %v\n", nodeData.Metadata.Time)
				log.Printf("Temperature: %v\n", nodeData.PayloadFields.Temp)
				log.Printf("Humidity: %v\n", nodeData.PayloadFields.Humid)
				log.Printf("Battery: %v\n", nodeData.PayloadFields.Bat)
				log.Printf("Rain: %v\n", nodeData.PayloadFields.Rain)
			}
		}
	}
}
