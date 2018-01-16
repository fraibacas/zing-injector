package zpubsub

import (
	"fmt"
	"time"

	"cloud.google.com/go/pubsub"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	"golang.org/x/net/context"
)

func getGoogleClient(ctx context.Context, projectID string) (*pubsub.Client, error) {
	c, err := pubsub.NewClient(ctx, projectID)
	if err != nil {
		return nil, errors.Wrap(err, "Failed to create pubsub client")
	}
	return c, nil
}

// GetOrCreateTopicFromClient creates or retrieves a topic for the given pubsub client
func GetOrCreateTopicFromClient(ctx context.Context, client *pubsub.Client, topicName string) (*pubsub.Topic, error) {
	topic := client.Topic(topicName)
	exists, err := topic.Exists(ctx)
	if err != nil {
		return nil, err
	}

	if !exists {
		log.WithField("topicName", topicName).Info("Created new topic")
		topic, err = client.CreateTopic(ctx, topicName)
		if err != nil {
			return nil, err
		}
	}
	return topic, nil
}

// GetOrCreateSubscriptionFromClient creates or retrieves a subscription for the given pubsub client and topic
func GetOrCreateSubscriptionFromClient(ctx context.Context, client *pubsub.Client, topic *pubsub.Topic, subName string) (*pubsub.Subscription, error) {
	sub := client.Subscription(subName)
	exists, err := sub.Exists(ctx)
	if err != nil {
		return nil, err
	}
	if !exists {
		fmt.Println("Creating Subscription ", subName)
		sub, err = client.CreateSubscription(ctx, subName, pubsub.SubscriptionConfig{
			Topic:             topic,
			RetentionDuration: 168 * time.Hour, // FOR SOME REASON THIS IS NEEDED
			AckDeadline:       30 * time.Second,
		})
		if err != nil {
			return nil, err
		}
	}
	return sub, nil
}

// GetOrCreateTopic creates or retrieves a topic from projectId and topic name
func GetOrCreateTopic(ctx context.Context, projectID, topicName string) (*pubsub.Topic, error) {
	gc, err := getGoogleClient(ctx, projectID)
	if err != nil {
		return nil, err
	}
	return GetOrCreateTopicFromClient(ctx, gc, topicName)
}

// GetOrCreateSubscription creates or retrieves a topic given projectId and topic name and subscription name
func GetOrCreateSubscription(ctx context.Context, projectID, topicName, subName string) (*pubsub.Subscription, error) {
	gc, err := getGoogleClient(ctx, projectID)
	if err != nil {
		return nil, err
	}
	topic, err := GetOrCreateTopicFromClient(ctx, gc, topicName)
	if err != nil {
		return nil, err
	}
	return GetOrCreateSubscriptionFromClient(ctx, gc, topic, subName)
}

// GetTopics returns all pubsub topics within the project
func GetTopics(ctx context.Context, projectID string) (*pubsub.TopicIterator, error) {
	gc, err := getGoogleClient(ctx, projectID)
	if err != nil {
		return nil, err
	}
	return gc.Topics(ctx), nil
}

// GetSubscriptions returns all pubsub subscriptions within the project
func GetSubscriptions(ctx context.Context, projectID string) (*pubsub.SubscriptionIterator, error) {
	gc, err := getGoogleClient(ctx, projectID)
	if err != nil {
		return nil, err
	}
	return gc.Subscriptions(ctx), nil
}

// Config has all the config needed to connect/create pubsub clients, topics and subscriptions
type Config struct {
	ProjectID        string
	TopicName        string
	SubscriptionName string
}

// NewConfig creates PubSubConfig from the passed info
func NewConfig(projectID, topicName, subName string) Config {
	return Config{ProjectID: projectID, TopicName: topicName, SubscriptionName: subName}
}

// Client wraps all the necessary components to send/receive messages from pubsub
type Client struct {
	ctx          context.Context
	client       *pubsub.Client
	topic        *pubsub.Topic
	subscription *pubsub.Subscription
	config       Config
}

// NewClient creates a PubSubClient from the passed configuration
func NewClient(ctx context.Context, cfg Config) (*Client, error) {
	gc, err := getGoogleClient(ctx, cfg.ProjectID)
	if err != nil {
		return nil, err
	}

	topic, err := GetOrCreateTopicFromClient(ctx, gc, cfg.TopicName)
	if err != nil {
		return nil, errors.Wrap(err, "Could not create/get topic")
	}

	// create the subscription if required
	var sub *pubsub.Subscription
	if cfg.SubscriptionName != "" {
		sub, err = GetOrCreateSubscriptionFromClient(ctx, gc, topic, cfg.SubscriptionName)
		if err != nil {
			return nil, errors.Wrap(err, "Could not create/get subscription")
		}
	}

	return &Client{ctx: ctx, client: gc, topic: topic, subscription: sub, config: cfg}, nil
}

// Message data received from pubsub
type Message struct {
	ID          string
	Data        []byte
	PublishTime time.Time
}

// MessageHandler func signature of message processors
type MessageHandler interface {
	Handle(m *Message) error
}

// Publish sends data to the topic the client was created for
func (c *Client) Publish(data []byte) (string, error) {
	res := c.topic.Publish(c.ctx, &pubsub.Message{Data: data})
	msgID, err := res.Get(c.ctx)
	if err != nil {
		return "", errors.Wrap(err, "Error sending msg")
	}
	fmt.Println("Queued message id ", msgID)
	return msgID, nil
}

// Receive listens for data from the topic and subscription the client was created for
func (c *Client) Receive(ctx context.Context, mh MessageHandler) error {
	if c.subscription == nil {
		return errors.New("Cannot receive on a nil subscription")
	}
	err := c.subscription.Receive(ctx, func(ctx context.Context, m *pubsub.Message) {
		zm := &Message{ID: m.ID, Data: m.Data, PublishTime: m.PublishTime}
		herr := mh.Handle(zm)
		if herr == nil {
			m.Ack()
		} else {
			m.Nack()
		}
	})
	if err != nil {
		return err
	}
	return nil
}

/*  SAMPLE CODE TO ITERATE THROUGH TOPICS AND SUBSCRIPTIONS

import "zing-injector/vendor/google.golang.org/api/iterator"


// Get topics just for fun
topics, err := zpubsub.GetTopics(ctx, cfg.ProjectID)
if err != nil {
	log.WithError(err).WithField("projectID", cfg.ProjectID).Error("Could not retrieve topics")
} else {
	for {
		t, terr := topics.Next()
		if terr != nil && terr.Error() == iterator.Done.Error() {
			break
		}
		if terr != nil {
			log.WithError(err).Error("Error iterating topics")
		} else {
			fmt.Println(t)
		}
	}
}
// Get subscriptions just for fun
subs, err := zpubsub.GetSubscriptions(ctx, cfg.ProjectID)
if err != nil {
	log.WithError(err).WithField("projectID", cfg.ProjectID).Error("Could not retrieve Subscriptions")
} else {
	for {
		s, serr := subs.Next()
		if serr != nil && serr.Error() == iterator.Done.Error() {
			break
		}
		if serr != nil {
			log.WithError(err).Error("Error iterating Subscriptions")
		} else {
			fmt.Println(s)
		}
	}
}
*/
