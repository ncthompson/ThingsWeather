package thingsif

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
)

/*
 * Handler structure
 */

type HandMessage struct {
	Services []HandServ `json:"services"`
}

type HandServ struct {
	MqttAddr string     `json:"mqtt_address"`
	Metadata []HandMeta `json:"metadata"`
}

type HandMeta struct {
	AppId string `json:"app_id"`
}

func getHttpBody(url, password string) ([]byte, error) {
	client := &http.Client{}
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("Failed to get body: %v", err)
	}
	if password != "" {
		req.Header.Add("Authorization", "key "+password)
	}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("Failed to get body: %v", err)
	}
	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("Failed to get body: %v", resp.Status)
	}
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("Failed to get body: %v", err)
	}
	return body, nil

}

func (mq *mqttCli) getBroker() (string, error) {
	url := "http://discovery.thethingsnetwork.org:8080/announcements/handler"
	body, err := getHttpBody(url, "")
	if err != nil {
		return "", fmt.Errorf("Failed to get broker: %v", err)
	}
	services := &HandMessage{}
	err = json.Unmarshal(body, services)
	if err != nil {
		return "", fmt.Errorf("Failed to get broker: %v", err)
	}
	for i := 0; i < len(services.Services); i++ {
		serv := services.Services[i]
		for j := 0; j < len(serv.Metadata); j++ {
			met := serv.Metadata[j]
			if met.AppId == mq.conf.Username {
				return serv.MqttAddr, nil
			}
		}
	}
	return "", errors.New("Application username not found.")
}
