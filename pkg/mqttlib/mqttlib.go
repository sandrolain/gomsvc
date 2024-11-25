package mqttlib

import (
	"context"
	"crypto/sha256"
	"crypto/tls"
	"fmt"
	"sync"
	"time"

	mqtt "github.com/eclipse/paho.mqtt.golang"
)

// EnvClientConfig represents the environment configuration for MQTT client
type EnvClientConfig struct {
	Broker     string `env:"MQTT_BROKER" validate:"required,hostname_port"`
	ClientID   string `env:"MQTT_CLIENT_ID" validate:"required"`
	Username   string `env:"MQTT_USERNAME"`
	Password   string `env:"MQTT_PASSWORD"`
	CACertPath string `env:"MQTT_CA_CERT_PATH"`
}

// ClientOptions represents the configuration options for MQTT client
type ClientOptions struct {
	Broker               string
	ClientID             string
	Username             string
	Password             string
	KeepAlive            time.Duration
	PingTimeout          time.Duration
	ConnectTimeout       time.Duration
	MaxReconnectInterval time.Duration
	AutoReconnect        bool
	CleanSession         bool
	Store                mqtt.Store
	TLSConfig            *tls.Config
	OnConnect            func()
	OnConnectionLost     func(error)
}

// DefaultClientOptions returns default client options
func DefaultClientOptions() ClientOptions {
	return ClientOptions{
		KeepAlive:            60 * time.Second,
		PingTimeout:          1 * time.Second,
		ConnectTimeout:       30 * time.Second,
		MaxReconnectInterval: 10 * time.Minute,
		AutoReconnect:        true,
		CleanSession:         false,
	}
}

// ClientOptionsFromEnvConfig creates ClientOptions from environment configuration
func ClientOptionsFromEnvConfig(cfg EnvClientConfig) ClientOptions {
	opts := DefaultClientOptions()
	opts.Broker = cfg.Broker
	opts.ClientID = cfg.ClientID
	opts.Username = cfg.Username
	opts.Password = cfg.Password

	if cfg.CACertPath != "" {
		tlsConfig := &tls.Config{
			MinVersion:         tls.VersionTLS12,
			InsecureSkipVerify: false,
		}
		opts.TLSConfig = tlsConfig
	}

	return opts
}

// NewClient creates a new MQTT client with the given options
func NewClient(co ClientOptions) (*Client, error) {
	opts := mqtt.NewClientOptions()
	opts.AddBroker(co.Broker)
	opts.SetClientID(co.ClientID)
	opts.SetKeepAlive(co.KeepAlive)
	opts.SetPingTimeout(co.PingTimeout)
	opts.SetCleanSession(co.CleanSession)
	opts.SetAutoReconnect(co.AutoReconnect)
	opts.SetMaxReconnectInterval(co.MaxReconnectInterval)
	opts.SetConnectTimeout(co.ConnectTimeout)

	if co.Username != "" {
		opts.SetUsername(co.Username)
		opts.SetPassword(co.Password)
	}

	if co.TLSConfig != nil {
		opts.SetTLSConfig(co.TLSConfig)
	}

	if co.Store != nil {
		opts.SetStore(co.Store)
	}

	if co.OnConnect != nil {
		opts.SetOnConnectHandler(func(c mqtt.Client) {
			co.OnConnect()
		})
	}

	if co.OnConnectionLost != nil {
		opts.SetConnectionLostHandler(func(c mqtt.Client, err error) {
			co.OnConnectionLost(err)
		})
	}

	client := mqtt.NewClient(opts)
	token := client.Connect()
	if token.Wait() && token.Error() != nil {
		return nil, fmt.Errorf("cannot create mqtt client: %w", token.Error())
	}

	return &Client{
		client: &client,
		subs:   make(map[string]mqtt.MessageHandler),
		mu:     &sync.RWMutex{},
	}, nil
}

// Client represents an MQTT client
type Client struct {
	client *mqtt.Client
	subs   map[string]mqtt.MessageHandler
	mu     *sync.RWMutex
}

// SubscribeHandler is a function type for handling incoming messages
type SubscribeHandler func(context.Context, IncomingMessage)

// Subscribe subscribes to a topic with the given QoS and handler
func (c *Client) Subscribe(ctx context.Context, topic string, qos byte, h SubscribeHandler) error {
	handler := func(c mqtt.Client, m mqtt.Message) {
		h(ctx, IncomingMessage{
			Message: m,
		})
	}

	c.mu.Lock()
	c.subs[topic] = handler
	c.mu.Unlock()

	token := (*c.client).Subscribe(topic, qos, handler)
	token.Wait()
	if err := token.Error(); err != nil {
		c.mu.Lock()
		delete(c.subs, topic)
		c.mu.Unlock()
		return fmt.Errorf("failed to subscribe to topic %s: %w", topic, err)
	}
	return nil
}

// Unsubscribe unsubscribes from the given topics
func (c *Client) Unsubscribe(topics ...string) error {
	token := (*c.client).Unsubscribe(topics...)
	token.Wait()
	if err := token.Error(); err != nil {
		return fmt.Errorf("failed to unsubscribe from topics: %w", err)
	}

	c.mu.Lock()
	for _, topic := range topics {
		delete(c.subs, topic)
	}
	c.mu.Unlock()

	return nil
}

// Publish publishes a message to the given topic with the specified QoS
func (c *Client) Publish(ctx context.Context, topic string, qos byte, retained bool, payload interface{}) error {
	token := (*c.client).Publish(topic, qos, retained, payload)
	token.Wait()
	if err := token.Error(); err != nil {
		return fmt.Errorf("failed to publish message to topic %s: %w", topic, err)
	}
	return nil
}

// Close disconnects the client and cleans up resources
func (c *Client) Close() {
	if c.client != nil {
		(*c.client).Disconnect(250)
	}
}

// IsConnected returns true if the client is currently connected
func (c *Client) IsConnected() bool {
	return (*c.client).IsConnected()
}

// IncomingMessage represents a received MQTT message
type IncomingMessage struct {
	Message mqtt.Message
}

// Topic returns the topic of the message
func (m *IncomingMessage) Topic() string {
	return m.Message.Topic()
}

// Payload returns the payload of the message
func (m *IncomingMessage) Payload() []byte {
	return m.Message.Payload()
}

// Hash returns a SHA-256 hash of the message payload and topic
func (m *IncomingMessage) Hash() []byte {
	h := sha256.New()
	d := append(m.Message.Payload(), []byte(m.Message.Topic())...)
	h.Write(d)
	return h.Sum(nil)
}
