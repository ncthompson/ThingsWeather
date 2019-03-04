package influxif

import (
	"fmt"
	"log"
	"strconv"
	"time"

	"github.com/influxdata/influxdb/client/v2"
	"github.com/ncthompson/ThingsWeather/interfaces/stbsource"
	"github.com/ncthompson/ThingsWeather/interfaces/thingsif"
)

const precision = "ns"

type InfluxConfig struct {
	HostAddress string
	Database    string
	Username    string
	Password    string
}

type InfluxIf struct {
	conf InfluxConfig
	cli  client.Client
}

func NewClient(conf InfluxConfig) (*InfluxIf, error) {
	inf := &InfluxIf{}
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

func (inf *InfluxIf) dataToBatch(data *thingsif.Message) (client.BatchPoints, error) {

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
		"device-id":       data.DevID,
		"hardware-serial": data.HWSerial,
		"port":            strconv.Itoa(data.Port),
	}
	payload := data.PayloadFields
	meta := data.Metadata
	if payload.Valid {
		err = setPayload(payload, timeStamp, tags, bp)
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
		err = setGateway(gw, meta.Frequency, timeStamp, tags, bp)
		if err != nil {
			return bp, err
		}
	}

	return bp, nil
}

func (inf *InfluxIf) stbDataToBatch(data *stbsource.StbWeather) (client.BatchPoints, error) {
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

func (inf *InfluxIf) WriteStbToDatabase(data *stbsource.StbWeather) error {
	batch, err := inf.stbDataToBatch(data)
	if err != nil {
		return err
	}
	return inf.cli.Write(batch)
}

func (inf *InfluxIf) WriteToDatabase(data *thingsif.Message) error {
	batch, err := inf.dataToBatch(data)
	if err != nil {
		return err
	}
	return inf.cli.Write(batch)
}

func (inf *InfluxIf) SyncDatabase(data []thingsif.DbMessage) error {

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
			payload := data[i]
			if payload.Valid {
				added++
				tags := map[string]string{
					"device-id": data[i].DevID,
				}
				err = setPayload(payload.NodeEntry(), ts, tags, bp)
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

func setGateway(g *thingsif.GwMetadata, freq float64, t time.Time, tags map[string]string, bp client.BatchPoints) error {
	tagsGW := tags
	tagsGW["gtw_id"] = g.GtwID
	tagsGW["channel"] = strconv.Itoa(g.Channel)
	tagsGW["frequency"] = strconv.FormatFloat(freq, 'f', 1, 64)
	err := addDataPoint("rssi", g.RSSI, t, tagsGW, bp)
	if err != nil {
		return err
	}
	err = addDataPoint("snr", g.SNR, t, tagsGW, bp)
	if err != nil {
		return err
	}
	err = addDataPoint("altitude", g.Altitude, t, tagsGW, bp)
	if err != nil {
		return err
	}
	err = addDataPoint("latitude", g.Latitude, t, tagsGW, bp)
	if err != nil {
		return err
	}
	err = addDataPoint("longitude", g.Longitude, t, tagsGW, bp)
	if err != nil {
		return err
	}
	return nil
}

func setPayload(p *thingsif.NodeEntry, t time.Time, tags map[string]string, bp client.BatchPoints) error {
	err := addDataPoint("temperature", p.Temp, t, tags, bp)
	if err != nil {
		return err
	}
	err = addDataPoint("humidity", p.Humid, t, tags, bp)
	if err != nil {
		return err
	}
	err = addDataPoint("battery-voltage", p.Bat, t, tags, bp)
	if err != nil {
		return err
	}
	err = addDataPoint("rain-tips", p.Rain, t, tags, bp)
	if err != nil {
		return err
	}
	err = addDataPoint("pressure", p.Pres, t, tags, bp)
	if err != nil {
		return err
	}
	return nil
}

func (inf *InfluxIf) Close() {
	inf.cli.Close()
}
