package thingsif

import (
	"encoding/json"
	"fmt"
	MQTT "github.com/eclipse/paho.mqtt.golang"
	"log"
	"os"
)

/*
   Payload Structure
*/

type GwMetadata struct {
	Timestamp  int64   `json:"timestamp"`
	GtwId      string  `json:"gtw_id"`
	GtwTrusted bool    `json:"gtw_trusted"`
	Channel    int     `json:"channel"`
	Rssi       float64 `json:"rssi"`
	Snr        float64 `json:"snr"`
	RfChain    int     `json:"rf_chain"`
	Latitude   float64 `json:"latitude"`
	Longitude  float64 `json:"longitude"`
	Altitude   float64 `json:"altitude"`
}

type Message struct {
	Payload_fields *NodeEntry    `json:"payload_fields"`
	DevId          string        `json:"dev_id"`
	AppId          string        `json:"app_id"`
	HwSerial       string        `json:"hardware_serial"`
	Port           int           `json:"port"`
	Counter        int           `json:"counter"`
	PayloadRaw     string        `json:"payload_raw"`
	Metadata       *NodeMetadata `json:"metadata"`
}

type NodeMetadata struct {
	Time       string        `json:"time"`
	Frequency  float64       `json:"frequency"`
	Modulation string        `json:"modulation"`
	DataRate   string        `json:"data_rate"`
	CodingRate string        `json:"coding_rate"`
	Gateways   []*GwMetadata `json:"gateways"`
}

type NodeEntry struct {
	Bat   float64 `json:"bat"`
	Humd  float64 `json:"humd"`
	Temp  float64 `json:"temp"`
	Rain  float64 `json:"rain,omitempty"`
	Pres  float64 `json:"pres,omitempty"`
	Valid bool    `json:"valid"`
}

/*
   Configuration
*/

type MqttConfig struct {
	Username string
	Password string
}

type mqttCli struct {
	choke chan [2]string
	cli   MQTT.Client
	conf  MqttConfig
}

func InitialiseMqttClient(conf MqttConfig) (*mqttCli, error) {
	mqtt := &mqttCli{}
	mqtt.conf = conf

	opts := MQTT.NewClientOptions()
	opts.SetAutoReconnect(true)
	opts.SetMessageChannelDepth(1024)
	opts.SetPassword(conf.Password)
	opts.SetUsername(conf.Username)
	url, err := mqtt.getBroker()
	if err != nil {
		return nil, err
	}
	broker := "tcp://" + url
	opts.AddBroker(broker)
	topic := "+/devices/+/up"
	opts.OnConnect = func(c MQTT.Client) {
		log.Print("Connected\n")
		if token := c.Subscribe(topic, byte(0), nil); token.Wait() && token.Error() != nil {
			fmt.Println(token.Error())
			os.Exit(1)
		}
		log.Print("Subscribed\n")
	}
	mqtt.choke = make(chan [2]string)
	opts.SetDefaultPublishHandler(func(client MQTT.Client, msg MQTT.Message) {
		mqtt.choke <- [2]string{msg.Topic(), string(msg.Payload())}
	})

	mqttCli := MQTT.NewClient(opts)

	if token := mqttCli.Connect(); token.Wait() && token.Error() != nil {
		return nil, token.Error()
	}
	mqtt.cli = mqttCli

	return mqtt, nil
}

/*
 * Blocks on incoming message.
 */
func (mq *mqttCli) WaitForData() (*Message, error) {
	msg := &Message{}
	incoming := <-mq.choke
	err := json.Unmarshal([]byte(incoming[1]), msg)

	if err != nil {
		return nil, err
	}
	return msg, nil
}

/*
 *
 */
func PrintGatways(gw []*GwMetadata) {
	for i := 0; i < len(gw); i++ {
		log.Printf("GW Id:%v\n", gw[i].GtwId)
		log.Printf("RSSI:%v\n", gw[i].Rssi)
		log.Printf("SNR:%v\n", gw[i].Snr)
	}
}

func (mq *mqttCli) Close() {
	mq.cli.Disconnect(1000)
}
