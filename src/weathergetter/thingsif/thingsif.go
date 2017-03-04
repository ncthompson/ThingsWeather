package thingsif

import (
	"encoding/json"
	"fmt"
	MQTT "github.com/eclipse/paho.mqtt.golang"
	"os"
)

/*
   Payload Structure
*/

type GwMetadata struct {
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
	Valid bool    `json:"valid"`
}

/*
   Configuration
*/

type MqttConfig struct {
	Region   string
	Username string
	Password string
}

type mqttCli struct {
	choke chan [2]string
	cli   MQTT.Client
}

func InitialiseMqttClient(conf MqttConfig) (*mqttCli, error) {
	mqtt := &mqttCli{}

	opts := MQTT.NewClientOptions()
	opts.SetAutoReconnect(true)
	opts.SetMessageChannelDepth(1024)
	opts.SetPassword(conf.Password)
	opts.SetUsername(conf.Username)
	broker := "tcp://" + conf.Region + ".thethings.network:1883"
	opts.AddBroker(broker)
	topic := "+/devices/+/up"
	opts.OnConnect = func(c MQTT.Client) {
		fmt.Print("Connected\n")
		if token := c.Subscribe(topic, byte(0), nil); token.Wait() && token.Error() != nil {
			fmt.Println(token.Error())
			os.Exit(1)
		}
		fmt.Print("Subscribed\n")
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
		fmt.Printf("GW Id:%v\n", gw[i].GtwId)
		fmt.Printf("RSSI:%v\n", gw[i].Rssi)
		fmt.Printf("SNR:%v\n", gw[i].Snr)
	}
}
