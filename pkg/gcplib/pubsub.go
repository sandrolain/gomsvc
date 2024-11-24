package gcplib

import (
	"context"
	"fmt"
	"time"

	"cloud.google.com/go/pubsub"

	"log/slog"
)

// PubSub represents a GCP Pub/Sub client
type PubSub struct {
	client *pubsub.Client
}

// NewPubSub creates a new PubSub client
func NewPubSub(ctx context.Context, projectID string) (*PubSub, error) {
	client, err := pubsub.NewClient(ctx, projectID)
	if err != nil {
		return nil, fmt.Errorf("error creating Pub/Sub client: %v", err)
	}
	return &PubSub{client}, nil
}

// Topic creates a new topic
func (p *PubSub) Topic(ctx context.Context, topicID string) (*pubsub.Topic, error) {
	topic := p.client.Topic(topicID)
	exists, err := topic.Exists(ctx)
	if err != nil {
		return nil, fmt.Errorf("error checking if topic exists: %v", err)
	}
	if !exists {
		topic, err = p.client.CreateTopic(ctx, topicID)
		if err != nil {
			return nil, fmt.Errorf("error creating topic: %v", err)
		}
	}
	return topic, nil
}

// Subscription creates a new subscription
func (p *PubSub) Subscription(ctx context.Context, topicID, subscriptionID string) (*pubsub.Subscription, error) {
	topic, err := p.Topic(ctx, topicID)
	if err != nil {
		return nil, fmt.Errorf("error getting topic: %v", err)
	}
	subscription := p.client.Subscription(subscriptionID)
	exists, err := subscription.Exists(ctx)
	if err != nil {
		return nil, fmt.Errorf("error checking if subscription exists: %v", err)
	}
	if !exists {
		subscription, err = p.client.CreateSubscription(ctx, subscriptionID, pubsub.SubscriptionConfig{
			Topic:       topic,
			AckDeadline: 20 * time.Second,
		})
		if err != nil {
			return nil, fmt.Errorf("error creating subscription: %v", err)
		}
	}
	return subscription, nil
}

// Publish publishes a message to a topic
func (p *PubSub) Publish(ctx context.Context, topicID string, data []byte) (string, error) {
	topic, err := p.Topic(ctx, topicID)
	if err != nil {
		return "", fmt.Errorf("error getting topic: %v", err)
	}
	msg := &pubsub.Message{
		Data: data,
	}
	result := topic.Publish(ctx, msg)
	// Block until the result is resolved
	_, err = result.Get(ctx)
	if err != nil {
		return "", fmt.Errorf("error publishing message: %v", err)
	}
	return result.Get(ctx)
}

// Pull pulls messages from a subscription until the context is cancelled
// It returns a cancel function that can be called to stop pulling messages
func (p *PubSub) Pull(ctx context.Context, subscriptionID string, callback func(*pubsub.Message)) (func(), error) {
	subscription, err := p.Subscription(ctx, "", subscriptionID)
	if err != nil {
		return nil, fmt.Errorf("error getting subscription: %v", err)
	}

	cancelCtx, cancel := context.WithCancel(ctx)
	go func() {
		err = subscription.Receive(cancelCtx, func(ctx context.Context, msg *pubsub.Message) {
			callback(msg)
			msg.Ack()
		})
		if err != nil && err != context.Canceled {
			slog.Error("error pulling messages", "error", err)
		}
	}()
	return cancel, nil
}
