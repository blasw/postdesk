package broker

import (
	"context"
	"time"

	"github.com/IBM/sarama"
	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
)

type Broker interface {
	//listens forever
	Listen(topic string, handler func(*sarama.ConsumerMessage))
	//creates a message
	Produce(topic string, message []byte) (string, error)
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
				logrus.WithFields(logrus.Fields{
					"topic":   msg.Topic,
					"message": string(msg.Value),
				}).Debug("Message received")

				handler(msg)
			}
		}(cons)
	}
}

func (k *KafkaClient) Produce(topic string, msg []byte) (string, error) {
	uniqueID := uuid.New().String()

	_, _, err := k.producer.SendMessage(&sarama.ProducerMessage{
		Topic: topic,
		Value: sarama.ByteEncoder(msg),
		Headers: []sarama.RecordHeader{
			{
				Key:   []byte("UUID"),
				Value: []byte(uniqueID),
			},
		},
	})

	return uniqueID, err
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
