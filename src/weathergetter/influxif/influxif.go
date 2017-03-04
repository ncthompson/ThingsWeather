package influxif

import (
	"github.com/influxdata/influxdb/client/v2"
	"time"
	"weathergetter/thingsif"
)

type InfluxConfig struct {
	HostAddress string
	Database    string
	Username    string
	Password    string
}

type influxIf struct {
	conf InfluxConfig
	cli  client.Client
}

func InitialiseInfluxClient(conf InfluxConfig) (*influxIf, error) {
	inf := &influxIf{}
	inf.conf = conf
	idb, err := client.NewHTTPClient(client.HTTPConfig{
		Addr:     conf.HostAddress,
		Username: conf.Username,
		Password: conf.Password,
	})
	inf.cli = idb

	return inf, err
}

func (inf *influxIf) dataToBatch(data *thingsif.Message) (client.BatchPoints, error) {

	timeStamp, err := time.Parse(time.RFC3339Nano, data.Metadata.Time)
	if err != nil {
		return nil, err
	}

	bp, err := client.NewBatchPoints(client.BatchPointsConfig{
		Database:  inf.conf.Database,
		Precision: "us",
	})
	if err != nil {
		return bp, err
	}

	//Temperature
	tags := map[string]string{
		"device-id": data.DevId,
	}

	fieldTemp := map[string]interface{}{
		"value": data.Payload_fields.Temp,
	}
	ptTemp, err := client.NewPoint(
		"temperature",
		tags,
		fieldTemp,
		timeStamp,
	)
	fieldHumd := map[string]interface{}{
		"value": data.Payload_fields.Humd,
	}
	if err != nil {
		return bp, err
	}
	ptHumd, err := client.NewPoint(
		"humidity",
		tags,
		fieldHumd,
		timeStamp,
	)
	if err != nil {
		return bp, err
	}
	fieldBat := map[string]interface{}{
		"value": data.Payload_fields.Bat,
	}
	ptBat, err := client.NewPoint(
		"battery-voltage",
		tags,
		fieldBat,
		timeStamp,
	)

	if err != nil {
		return bp, err
	}

	bp.AddPoint(ptTemp)
	bp.AddPoint(ptHumd)
	bp.AddPoint(ptBat)
	return bp, nil
}

func (inf *influxIf) WriteToDatabase(data *thingsif.Message) error {
	batch, err := inf.dataToBatch(data)
	if err != nil {
		return err
	}
	return inf.cli.Write(batch)
}
