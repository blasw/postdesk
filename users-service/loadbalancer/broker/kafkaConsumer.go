package broker

import (
	"time"

	"github.com/IBM/sarama"
	"github.com/sirupsen/logrus"
)

type KafkaConsumer struct {
	client sarama.Consumer
}

func NewKafkaConsumer(addr string) (*KafkaConsumer, error) {
	var err error

	for i := 0; i < 10; i++ {
		client, err := sarama.NewConsumer([]string{addr}, nil)
		if err == nil {
			return &KafkaConsumer{client}, nil
		}

		time.Sleep(time.Second)
	}

	return nil, err
}

func (k *KafkaConsumer) Listen(topic string, handler func(*sarama.ConsumerMessage)) {
	parts, err := k.client.Partitions(topic)
	if err != nil {
		logrus.WithError(err).Fatal("Error accured while getting partitions")
	}

	for _, part := range parts {
		cons, err := k.client.ConsumePartition(topic, part, sarama.OffsetNewest)
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
