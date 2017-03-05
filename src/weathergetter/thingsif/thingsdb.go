package thingsif

import (
	"encoding/json"
	"fmt"
)

type DbMessage struct {
	Time  string  `json:"time"`
	DevId string  `json:"device_id"`
	Raw   string  `json:"raw"`
	Bat   float64 `json:"bat"`
	Humd  float64 `json:"humd"`
	Temp  float64 `json:"temp"`
	Valid bool    `json:"valid"`
}

type DbList struct {
	List []*DbMessage
}

func (mq *mqttCli) GetLast7days() ([]DbMessage, error) {
	url := "https://" + mq.conf.Username + ".data.thethingsnetwork.org/api/v2/query?last=7d"
	messages := make([]DbMessage, 0)
	body, err := getHttpBody(url, mq.conf.Password)
	if err != nil {
		return messages, fmt.Errorf("Failed to get history: %v", err)
	}
	err = json.Unmarshal(body, &messages)
	if err != nil {
		return messages, fmt.Errorf("Failed to get history: %v", err)
	}
	return messages, nil
}