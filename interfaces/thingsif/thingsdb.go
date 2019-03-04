package thingsif

import (
	"encoding/json"
	"fmt"
	"time"
)

type DbMessage struct {
	Time  string  `json:"time"`
	DevID string  `json:"device_id"`
	Raw   string  `json:"raw"`
	Bat   float64 `json:"bat"`
	Humid float64 `json:"humd"`
	Temp  float64 `json:"temp"`
	Rain  float64 `json:"rain,omitempty"`
	Pres  float64 `json:"pres,omitempty"`
	Valid bool    `json:"valid"`
}

type DbList struct {
	List []*DbMessage
}

func (mq *MQTTCli) GetLast7days() ([]DbMessage, error) {
	fmt.Printf("Start get %v\n", time.Now())
	url := "https://" + mq.conf.Username + ".data.thethingsnetwork.org/api/v2/query?last=7d"
	messages := make([]DbMessage, 0)
	body, err := getHTTPBody(url, mq.conf.Password)
	if err != nil {
		return messages, fmt.Errorf("failed to get history: %v", err)
	}
	err = json.Unmarshal(body, &messages)
	if err != nil {
		return messages, fmt.Errorf("failed to get history: %v", err)
	}
	fmt.Printf("done get %v\n", time.Now())
	return messages, nil
}

func (d *DbMessage) NodeEntry() *NodeEntry {
	return &NodeEntry{
		Bat:   d.Bat,
		Humid: d.Humid,
		Temp:  d.Temp,
		Rain:  d.Rain,
		Pres:  d.Pres,
		Valid: d.Valid,
	}
}
