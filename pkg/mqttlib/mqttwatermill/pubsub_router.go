package mqttwatermill

import (
	"context"
	"fmt"

	"log/slog"

	"github.com/ThreeDotsLabs/watermill"
	"github.com/ThreeDotsLabs/watermill-googlecloud/pkg/googlecloud"
	"github.com/ThreeDotsLabs/watermill/message"
	"github.com/ThreeDotsLabs/watermill/message/router/plugin"

	"github.com/sandrolain/gomsvc/pkg/mqttlib"
)

// PubSubRouter is a Watermill Router implementation that routes messages to GCP Pub/Sub
type PubSubRouter struct {
	Logger *slog.Logger

	// Subscriber is the Watermill Subscriber to receive messages from MQTT
	Subscriber message.Subscriber

	// Publisher is the Watermill Publisher to send messages to
	Publisher message.Publisher

	Router  *message.Router
	handler *message.Handler
}

type PubSubRouterConfig struct {
	ProjectID       string
	SubscriberTopic string
	PublisherTopic  string
	MqttClient      *mqttlib.Client
	HandlerName     string
}

// NewPubSubRouter creates a new PubSubRouter
func NewPubSubRouter(cfg PubSubRouterConfig, logger *slog.Logger) (*PubSubRouter, error) {
	if logger == nil {
		logger = slog.Default()
	}

	subscriber, err := NewSubscriber(cfg.MqttClient, logger)
	if err != nil {
		return nil, fmt.Errorf("failed to create MQTT subscriber: %w", err)
	}

	pubCfg := googlecloud.PublisherConfig{
		ProjectID: cfg.ProjectID,
	}

	wmLogger := watermill.NewSlogLogger(logger)

	publisher, err := googlecloud.NewPublisher(pubCfg, wmLogger)
	if err != nil {
		return nil, fmt.Errorf("failed to create Google Cloud publisher: %w", err)
	}

	router, err := message.NewRouter(message.RouterConfig{}, wmLogger)
	if err != nil {
		panic(err)
	}

	router.AddPlugin(plugin.SignalsHandler)

	handlerName := cfg.HandlerName
	if handlerName == "" {
		handlerName = "handler_" + cfg.SubscriberTopic
	}

	handler := router.AddHandler(
		handlerName,
		cfg.SubscriberTopic, // topic from which we will read events
		subscriber,
		cfg.PublisherTopic, // topic to which we will publish events
		publisher,
		message.PassthroughHandler,
	)

	return &PubSubRouter{
		Logger:     logger,
		Subscriber: subscriber,
		Publisher:  publisher,
		Router:     router,
		handler:    handler,
	}, nil
}

func (r *PubSubRouter) Start(ctx context.Context) error {
	return r.Router.Run(ctx)
}

func (r *PubSubRouter) Close() error {
	return r.Subscriber.Close()
}
