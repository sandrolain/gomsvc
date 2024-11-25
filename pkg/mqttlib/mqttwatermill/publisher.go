package mqttwatermill

import (
	"context"
	"fmt"
	"log/slog"
	"strconv"
	"sync"

	"github.com/ThreeDotsLabs/watermill"
	"github.com/ThreeDotsLabs/watermill/message"
	"github.com/sandrolain/gomsvc/pkg/mqttlib"
)

const (
	MetadataTopic     = "mqtt_topic"
	MetadataQoS       = "mqtt_qos"
	MetadataRetained  = "mqtt_retained"
	MetadataMessageID = "mqtt_message_id"
	MetadataDuplicate = "mqtt_duplicate"
)

// Publisher is a Watermill Publisher implementation for MQTT
type Publisher struct {
	client     *mqttlib.Client
	closed     bool
	closedLock sync.Mutex
	logger     watermill.LoggerAdapter
}

// NewPublisher creates a new MQTT Publisher
func NewPublisher(client *mqttlib.Client, logger *slog.Logger) (message.Publisher, error) {
	if logger == nil {
		logger = slog.Default()
	}

	return &Publisher{
		client: client,
		logger: watermill.NewSlogLogger(logger),
	}, nil
}

// Publish publishes messages to MQTT
func (p *Publisher) Publish(topic string, messages ...*message.Message) error {
	p.closedLock.Lock()
	if p.closed {
		p.closedLock.Unlock()
		return fmt.Errorf("publisher is closed")
	}
	p.closedLock.Unlock()

	for _, msg := range messages {
		qos := byte(0)
		retained := false

		if qosStr := msg.Metadata.Get(MetadataQoS); qosStr != "" {
			if qosInt, err := strconv.Atoi(qosStr); err == nil && qosInt >= 0 && qosInt <= 2 {
				qos = byte(qosInt)
			}
		}

		if retainedStr := msg.Metadata.Get(MetadataRetained); retainedStr != "" {
			if retainedBool, err := strconv.ParseBool(retainedStr); err == nil {
				retained = retainedBool
			}
		}

		p.logger.Trace("Publishing message", watermill.LogFields{
			"topic":     topic,
			"messageID": msg.UUID,
			"qos":       qos,
			"retained":  retained,
		})

		if err := p.client.Publish(context.Background(), topic, qos, retained, msg.Payload); err != nil {
			return err
		}
	}

	return nil
}

// Close closes the publisher
func (p *Publisher) Close() error {
	p.closedLock.Lock()
	defer p.closedLock.Unlock()

	if p.closed {
		return nil
	}

	p.closed = true
	return nil
}
