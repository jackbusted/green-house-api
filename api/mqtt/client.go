package mqtt

import (
	"encoding/json"
	"fmt"
	"green-house-api/api/response"

	mqtt "github.com/eclipse/paho.mqtt.golang"
)

type MqttClient struct {
	client mqtt.Client
}

func NewClient(brokerURL string, clientID string) (*MqttClient, error) {
	opts := mqtt.NewClientOptions()

	opts.AddBroker(brokerURL)
	opts.SetClientID(clientID)

	client := mqtt.NewClient(opts)
	token := client.Connect()

	if token.Wait() && token.Error() != nil {
		return nil, fmt.Errorf("failed to connect MQTT broker: %w", token.Error())
	}

	opts.SetAutoReconnect(true)

	return &MqttClient{client: client}, nil
}

func (c *MqttClient) Publish(topic string, payload response.DeviceReportResp) error {
	// 0 At most once
	// 1 At least once
	// 2 Exactly once

	jsonPayload, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal response : %s", err.Error())
	}

	token := c.client.Publish(
		topic,
		0,
		false,
		jsonPayload,
	)

	if token.Wait() && token.Error() != nil {
		return fmt.Errorf("failed to publish MQTT message : %w", token.Error())
	}

	return nil
}

func (c *MqttClient) IsConnected() bool {
	if c == nil || c.client == nil {
		return false
	}

	return c.client.IsConnected()
}

func (c *MqttClient) Disconnect() {
	c.client.Disconnect(250)
}
