package configuration

import (
	"encoding/json"
	"github.com/ncthompson/ThingsWeather/interfaces/influxif"
	"github.com/ncthompson/ThingsWeather/interfaces/thingsif"
	"os"
)

type GetterConfig struct {
	DbConfig influxif.InfluxConfig
	MConfig  thingsif.MqttConfig
}

func sampleConfig() GetterConfig {
	db := influxif.InfluxConfig{
		HostAddress: "http://localhost:8086",
		Database:    "database_name",
		Username:    "database_username",
		Password:    "database_password",
	}

	mq := thingsif.MqttConfig{
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
	defer f.Close()
	if err != nil {
		return err
	}
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
	defer f.Close()
	if err != nil {
		return nil, err
	}
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
