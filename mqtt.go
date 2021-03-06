package main

import (
	"encoding/json"
	"fmt"
	"net/url"
	"os"
	"time"

	log "github.com/sirupsen/logrus"

	mqtt "github.com/eclipse/paho.mqtt.golang"
)

type mqttConfig struct {
	Name              string      `json:"name"`
	UnitOfMeasurement string      `json:"unit_of_measurement"`
	StateTopic        string      `json:"state_topic"`
	ConfigTopic       string      `json:"-"`
	UniqueID          string      `json:"unique_id"`
	Client            mqtt.Client `json:"-"`
}

func (m *mqttConfig) Connect(uri *url.URL) {
	opts := createClientOptions(m.Name, uri)
	client := mqtt.NewClient(opts)
	token := client.Connect()
	log.WithFields(log.Fields{"mqtt_broker": uri.Host}).Debug("Connecting to mqtt broker")
	for !token.WaitTimeout(10 * time.Second) {
	}

	if err := token.Error(); err != nil {
		log.Fatalln(err)
	}
	log.WithFields(log.Fields{"mqtt_broker": uri.Host}).Info("Connected to mqtt broker")
	m.Client = client
}

func createClientOptions(clientID string, uri *url.URL) *mqtt.ClientOptions {
	opts := mqtt.NewClientOptions()
	opts.AddBroker(fmt.Sprintf("tcp://%s", uri.Host))
	opts.SetClientID(clientID)
	if username := os.Getenv("MQTT_USERNAME"); len(username) > 0 {
		opts.SetUsername(username)
		log.WithFields(log.Fields{"mqtt_username": username}).Debug("Found mqtt username")
	}
	if password := os.Getenv("MQTT_PASSWORD"); len(password) > 0 {
		opts.SetPassword(password)
		log.Debug("Found mqtt password")
	}
	return opts
}

// SendConfig send the home assistant auto discovery config to mqtt
func (m *mqttConfig) SendConfig() error {
	mqttPayload, err := preparePayload(m)
	if err != nil {
		return err
	}
	log.Debug("Sending mqtt config")
	err = sendMessage(m.Client, m.ConfigTopic, mqttPayload)
	return err
}

// Send any data to home assistant
func (m *mqttConfig) Send(data string) error {
	log.Debug("Sending mqtt message")
	err := sendMessage(m.Client, m.StateTopic, data)
	return err
}

func sendMessage(client mqtt.Client, topic string, data string) error {
	token := client.Publish(topic, 0, false, data)
	if token.Error() != nil {
		log.Fatalln(token.Error())
		return token.Error()
	}
	return nil
}

func preparePayload(data interface{}) (string, error) {
	mqttPayload, err := json.Marshal(data)
	if err != nil {
		return "", err
	}
	return string(mqttPayload), nil
}
