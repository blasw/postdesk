package broker

import (
	"context"
	"time"

	"github.com/IBM/sarama"
	"github.com/sirupsen/logrus"
)

type Broker interface {
	//listens forever
	Listen(topic string, handler func(*sarama.ConsumerMessage))
	//creates a message
	Produce(uuid string, topic string, message []byte) error
	//waits for message
	WaitForMessage(ctx context.Context, topic string, uuid string) ([]byte, error)
}

type KafkaClient struct {
	consumer sarama.Consumer
	producer sarama.SyncProducer
}

func NewKafkaClient(addr string) (*KafkaClient, error) {
	var err error
	var consumer sarama.Consumer

	for i := 0; i < 10; i++ {
		consumer, err = sarama.NewConsumer([]string{addr}, nil)
		if err == nil {
			break
		}

		time.Sleep(time.Second)
	}

	if err != nil {
		return nil, err
	}

	producer, err := sarama.NewSyncProducer([]string{addr}, nil)
	if err != nil {
		return nil, err
	}

	config := sarama.NewConfig()
	config.Version = sarama.V3_6_0_0

	admin, err := sarama.NewClusterAdmin([]string{addr}, config)
	if err != nil {
		return nil, err
	}
	defer func() {
		if err := admin.Close(); err != nil {
			logrus.Fatal(err)
		}
	}()

	//CREATING TOPICS
	topics := []string{"create_topic", "create_topic_response", "delete_post", "like_post", "unlike_post", "get_posts", "get_posts_response", "sign_up", "sign_up_response", "sign_in", "sign_in_response"}

	for _, topic := range topics {
		_ = admin.CreateTopic(topic, &sarama.TopicDetail{
			NumPartitions:     1,
			ReplicationFactor: 1,
		}, false)
	}

	return &KafkaClient{consumer, producer}, nil
}

func (k *KafkaClient) Listen(topic string, handler func(*sarama.ConsumerMessage)) {
	parts, err := k.consumer.Partitions(topic)
	if err != nil {
		logrus.WithError(err).Fatal("Error accured while getting partitions")
	}

	for _, part := range parts {
		cons, err := k.consumer.ConsumePartition(topic, part, sarama.OffsetNewest)
		if err != nil {
			logrus.WithError(err).Fatal("Error accured while consuming partition")
		}

		go func(consumer sarama.PartitionConsumer) {
			defer consumer.AsyncClose()
			for msg := range cons.Messages() {
				handler(msg)
			}
		}(cons)
	}
}

func (k *KafkaClient) Produce(uuid string, topic string, msg []byte) error {
	_, _, err := k.producer.SendMessage(&sarama.ProducerMessage{
		Topic: topic,
		Value: sarama.ByteEncoder(msg),
		Headers: []sarama.RecordHeader{
			{
				Key:   []byte("UUID"),
				Value: []byte(uuid),
			},
		},
	})

	return err
}

func (k *KafkaClient) WaitForMessage(ctx context.Context, topic string, uuid string) ([]byte, error) {
	partConsumer, err := k.consumer.ConsumePartition(topic, 0, sarama.OffsetNewest)
	if err != nil {
		return nil, err
	}
	defer partConsumer.Close()

	for {
		select {
		case msg := <-partConsumer.Messages():
			for _, header := range msg.Headers {
				if string(header.Key) == "UUID" && string(header.Value) == uuid {
					return msg.Value, nil
				}
			}

		case err := <-partConsumer.Errors():
			return nil, err

		case <-ctx.Done():
			return nil, ctx.Err()
		}
	}
}
