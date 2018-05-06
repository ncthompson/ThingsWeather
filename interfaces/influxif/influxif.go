package influxif

import (
	"fmt"
	"github.com/influxdata/influxdb/client/v2"
	"interfaces/stbsource"
	"interfaces/thingsif"
	"log"
	"strconv"
	"time"
)

const precision = "ns"

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

func addDataPoint(name string, value float64, ts time.Time, tags map[string]string, bp client.BatchPoints) error {

	field := map[string]interface{}{
		"value": value,
	}
	pt, err := client.NewPoint(
		name,
		tags,
		field,
		ts,
	)
	if err != nil {
		return err
	}
	bp.AddPoint(pt)
	return nil
}

func (inf *influxIf) dataToBatch(data *thingsif.Message) (client.BatchPoints, error) {

	timeStamp, err := time.Parse(time.RFC3339Nano, data.Metadata.Time)
	if err != nil {
		return nil, err
	}

	bp, err := client.NewBatchPoints(client.BatchPointsConfig{
		Database:  inf.conf.Database,
		Precision: precision,
	})
	if err != nil {
		return bp, err
	}

	tags := map[string]string{
		"device-id":       data.DevId,
		"hardware-serial": data.HwSerial,
		"port":            strconv.Itoa(data.Port),
	}
	pld := data.Payload_fields
	meta := data.Metadata
	if pld.Valid {

		err = addDataPoint("temperature", pld.Temp, timeStamp, tags, bp)
		if err != nil {
			return bp, err
		}

		err = addDataPoint("humidity", pld.Humd, timeStamp, tags, bp)
		if err != nil {
			return bp, err
		}

		err = addDataPoint("battery-voltage", pld.Bat, timeStamp, tags, bp)
		if err != nil {
			return bp, err
		}

		err = addDataPoint("rain-tips", pld.Rain, timeStamp, tags, bp)
		if err != nil {
			return bp, err
		}

		err = addDataPoint("pressure", pld.Pres, timeStamp, tags, bp)
		if err != nil {
			return bp, err
		}
	}
	tags["modulation"] = meta.Modulation
	tags["data_rate"] = meta.DataRate
	tags["coding_rate"] = meta.CodingRate

	err = addDataPoint("frequency", meta.Frequency, timeStamp, tags, bp)
	if err != nil {
		return bp, err
	}

	for i := 0; i < len(meta.Gateways); i++ {
		gw := meta.Gateways[i]
		tagsGw := tags
		tagsGw["gtw_id"] = gw.GtwId
		tagsGw["channel"] = strconv.Itoa(gw.Channel)
		tagsGw["frequency"] = strconv.FormatFloat(meta.Frequency, 'f', 1, 64)
		err = addDataPoint("rssi", gw.Rssi, timeStamp, tagsGw, bp)
		if err != nil {
			return bp, err
		}
		err = addDataPoint("snr", gw.Snr, timeStamp, tagsGw, bp)
		if err != nil {
			return bp, err
		}
		err = addDataPoint("altitude", gw.Altitude, timeStamp, tagsGw, bp)
		if err != nil {
			return bp, err
		}
		err = addDataPoint("latitude", gw.Latitude, timeStamp, tagsGw, bp)
		if err != nil {
			return bp, err
		}
		err = addDataPoint("longitude", gw.Longitude, timeStamp, tagsGw, bp)
		if err != nil {
			return bp, err
		}
	}

	return bp, nil
}

func (inf *influxIf) stbDataToBatch(data *stbsource.StbWeather) (client.BatchPoints, error) {
	bp, err := client.NewBatchPoints(client.BatchPointsConfig{
		Database:  inf.conf.Database,
		Precision: precision,
	})
	if err != nil {
		return bp, err
	}
	tags := map[string]string{
		"device-id": data.Station,
	}

	err = addDataPoint("temperature", data.Temperature, data.Timestamp.UTC(), tags, bp)
	if err != nil {
		return bp, err
	}

	err = addDataPoint("humidity", data.Humidity, data.Timestamp.UTC(), tags, bp)
	if err != nil {
		return bp, err
	}
	return bp, nil
}

func (inf *influxIf) WriteStbToDatabase(data *stbsource.StbWeather) error {
	batch, err := inf.stbDataToBatch(data)
	if err != nil {
		return err
	}
	return inf.cli.Write(batch)
}

func (inf *influxIf) WriteToDatabase(data *thingsif.Message) error {
	batch, err := inf.dataToBatch(data)
	if err != nil {
		return err
	}
	return inf.cli.Write(batch)
}

func (inf *influxIf) SyncDatabase(data []thingsif.DbMessage) error {

	log.Printf("Entries: %v\n", len(data))
	added := 0
	bp, err := client.NewBatchPoints(client.BatchPointsConfig{
		Database:  inf.conf.Database,
		Precision: precision,
	})
	if err != nil {
		return err
	}
	for i := 0; i < len(data); i++ {
		ts, err := time.Parse(time.RFC3339Nano, data[i].Time)
		if err != nil {
			return err
		}
		// Assume all point will be logged together.
		query := fmt.Sprintf("select * from temperature where time=%v;", ts.UnixNano())
		q := client.NewQuery(query, inf.conf.Database, precision)

		response, err := inf.cli.Query(q)
		if err != nil {
			log.Println(err.Error())
			return err

		}
		if len(response.Results[0].Series) == 0 {
			pld := data[i]
			if pld.Valid {
				added++
				tags := map[string]string{
					"device-id": data[i].DevId,
				}
				err = addDataPoint("temperature", pld.Temp, ts, tags, bp)
				if err != nil {
					return err
				}

				err = addDataPoint("humidity", pld.Humd, ts, tags, bp)
				if err != nil {
					return err
				}

				err = addDataPoint("battery-voltage", pld.Bat, ts, tags, bp)
				if err != nil {
					return err
				}
			}
		}
	}
	log.Printf("Points synced: %v\n", added)
	if len(bp.Points()) > 0 {
		return inf.cli.Write(bp)
	}
	return nil
}

func (inf *influxIf) Close() {
	inf.cli.Close()
}
