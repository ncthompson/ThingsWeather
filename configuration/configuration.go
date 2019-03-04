package configuration

import (
	"encoding/json"
	"os"

	"github.com/ncthompson/ThingsWeather/interfaces/influxif"
	"github.com/ncthompson/ThingsWeather/interfaces/thingsif"
)

type GetterConfig struct {
	DbConfig influxif.InfluxConfig
	MConfig  thingsif.MQTTConfig
}

func sampleConfig() GetterConfig {
	db := influxif.InfluxConfig{
		HostAddress: "http://localhost:8086",
		Database:    "database_name",
		Username:    "database_username",
		Password:    "database_password",
	}

	mq := thingsif.MQTTConfig{
		Username: "application_id",
		Password: "access_key",
	}

	conf := GetterConfig{
		MConfig:  mq,
		DbConfig: db,
	}
	return conf
}

func CreateConfigTemplate() error {
	f, err := os.Create("sampleConfig.json")
	if err != nil {
		return err
	}
	defer f.Close()

	sample := sampleConfig()
	conf, err := json.MarshalIndent(sample, "", "\t")
	if err != nil {
		return err
	}
	_, err = f.Write(conf)
	return err
}

func OpenConfig(file string) (*GetterConfig, error) {
	f, err := os.Open(file)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	conf := &GetterConfig{}
	data := make([]byte, 4096)
	n, err := f.Read(data)
	if err != nil {
		return nil, err
	}
	err = json.Unmarshal(data[0:n], conf)
	if err != nil {
		return nil, err
	}
	return conf, nil
}
