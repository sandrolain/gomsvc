package mqttlib

import (
	"crypto/sha256"
	"fmt"
	"time"

	mqtt "github.com/eclipse/paho.mqtt.golang"
)

type EnvClientConfig struct {
	Broker   string `env:"MQTT_BROKER" validate:"required,hostname_port"`
	ClientID string `env:"MQTT_CLIENT_ID" validate:"required"`
}

type ClientOptions struct {
	Broker   string
	ClientID string
}

func ClientOptionsFromEnvConfig(cfg EnvClientConfig) ClientOptions {
	return ClientOptions{
		Broker:   cfg.Broker,
		ClientID: cfg.ClientID,
	}
}

func NewClient(co ClientOptions) (res *Client, err error) {
	opts := mqtt.NewClientOptions()
	opts.AddBroker(co.Broker)
	opts.SetClientID(co.ClientID)
	opts.SetKeepAlive(60 * time.Second)
	opts.SetPingTimeout(1 * time.Second)
	opts.SetCleanSession(false)

	client := mqtt.NewClient(opts)
	token := client.Connect()
	if token.Wait() && token.Error() != nil {
		err = fmt.Errorf("cannot create mqtt client: %w", token.Error())
		return
	}
	res = &Client{
		client: &client,
	}
	return
}

type Client struct {
	client *mqtt.Client
}

type SubscribeHandler func(IncomingMessage)

func (c *Client) Subscribe(topic string, qos byte, h SubscribeHandler) error {
	token := (*c.client).Subscribe(topic, qos, func(c mqtt.Client, m mqtt.Message) {
		h(IncomingMessage{
			Message: m,
		})
	})
	token.Wait()
	return token.Error()
}

type IncomingMessage struct {
	Message mqtt.Message
}

func (m *IncomingMessage) Hash() []byte {
	h := sha256.New()
	d := append(m.Message.Payload(), []byte(m.Message.Topic())...)
	h.Write(d)
	return h.Sum(nil)
}
