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

// Subscriber is a Watermill Subscriber implementation for MQTT
type Subscriber struct {
	client     *mqttlib.Client
	closed     bool
	closedLock sync.Mutex
	logger     watermill.LoggerAdapter

	outputChannels     map[string]chan *message.Message
	outputChannelsLock sync.RWMutex
}

// NewSubscriber creates a new MQTT Subscriber
func NewSubscriber(client *mqttlib.Client, logger *slog.Logger) (message.Subscriber, error) {
	if logger == nil {
		logger = slog.Default()
	}

	return &Subscriber{
		client:         client,
		logger:         watermill.NewSlogLogger(logger),
		outputChannels: make(map[string]chan *message.Message),
	}, nil
}

// Subscribe subscribes to MQTT topics
func (s *Subscriber) Subscribe(ctx context.Context, topic string) (<-chan *message.Message, error) {
	s.closedLock.Lock()
	if s.closed {
		s.closedLock.Unlock()
		return nil, fmt.Errorf("subscriber is closed")
	}
	s.closedLock.Unlock()

	s.logger.Info("Subscribing to MQTT topic", watermill.LogFields{
		"topic": topic,
	})

	// Create output channel for messages
	output := make(chan *message.Message)

	// Store the output channel
	s.outputChannelsLock.Lock()
	s.outputChannels[topic] = output
	s.outputChannelsLock.Unlock()

	// Subscribe to MQTT topic
	err := s.client.Subscribe(ctx, topic, 1, func(ctx context.Context, msg mqttlib.IncomingMessage) {
		message := message.NewMessage(watermill.NewUUID(), msg.Payload())
		message.Metadata.Set(MetadataTopic, msg.Topic())
		message.Metadata.Set(MetadataMessageID, strconv.Itoa(int(msg.Message.MessageID())))
		message.Metadata.Set(MetadataQoS, strconv.Itoa(int(msg.Message.Qos())))
		message.Metadata.Set(MetadataDuplicate, strconv.FormatBool((msg.Message.Duplicate())))
		message.Metadata.Set(MetadataRetained, strconv.FormatBool(msg.Message.Retained()))

		s.logger.Trace("Received message", watermill.LogFields{
			"topic":     topic,
			"messageID": message.UUID,
		})

		output <- message

		// Send message to output channel
		select {
		case <-message.Acked():
			msg.Message.Ack()
		case <-message.Nacked():
			// TODO: Handle NACK
		case <-ctx.Done():
		}
	})

	if err != nil {
		s.outputChannelsLock.Lock()
		delete(s.outputChannels, topic)
		s.outputChannelsLock.Unlock()
		close(output)
		return nil, err
	}

	// Handle cleanup when context is done
	go func() {
		<-ctx.Done()
		s.outputChannelsLock.Lock()
		if ch, exists := s.outputChannels[topic]; exists {
			delete(s.outputChannels, topic)
			close(ch)
		}
		s.outputChannelsLock.Unlock()

		if err := s.client.Unsubscribe(topic); err != nil {
			s.logger.Error("Failed to unsubscribe from topic", err, watermill.LogFields{
				"topic": topic,
			})
		}
	}()

	return output, nil
}

// Close closes the subscriber
func (s *Subscriber) Close() error {
	s.closedLock.Lock()
	defer s.closedLock.Unlock()

	if s.closed {
		return nil
	}

	s.closed = true

	s.outputChannelsLock.Lock()
	for _, ch := range s.outputChannels {
		close(ch)
	}
	s.outputChannels = make(map[string]chan *message.Message)
	s.outputChannelsLock.Unlock()

	return nil
}
