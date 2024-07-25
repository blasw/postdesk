package broker

import (
	"time"

	"github.com/IBM/sarama"
	"github.com/sirupsen/logrus"
)

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

func (k *KafkaClient) Produce(topic string, msg []byte, headers ...sarama.RecordHeader) error {
	_, _, err := k.producer.SendMessage(&sarama.ProducerMessage{
		Topic:   topic,
		Value:   sarama.ByteEncoder(msg),
		Headers: headers,
	})

	return err
}
