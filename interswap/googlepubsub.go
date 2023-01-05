package interswap

import (
	"context"
	"log"

	"cloud.google.com/go/pubsub"
	"google.golang.org/api/option"
)

var (
	psclient *pubsub.Client

	interSwapTxTopic *pubsub.Topic
)

func startPubsubClient(ggc_project string, ggc_acc string) error {
	ctx := context.Background()

	client, err := pubsub.NewClient(ctx, ggc_project, option.WithCredentialsFile(ggc_acc))
	if err != nil {
		log.Fatal(err)
	}
	psclient = client
	return nil
}

func startPubsubTopic(topicName string) (*pubsub.Topic, error) {
	ctx := context.Background()

	topic := psclient.Topic(topicName)

	// Create the topic if it doesn't exist.
	exists, err := topic.Exists(ctx)
	if err != nil {
		log.Println(err)
	}
	if !exists {
		log.Printf("Topic %v doesn't exist - creating it", topicName)
		topic, err = psclient.CreateTopic(ctx, topicName)
		if err != nil {
			log.Fatal(err)
		}
	}
	return topic, nil
}
