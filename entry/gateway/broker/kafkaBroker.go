package broker

import (
	"errors"
	"time"

	"github.com/IBM/sarama"
	"github.com/sirupsen/logrus"
)

type Broker interface {
	Publish(msg any) error
}

type KafkaBroker struct {
	client *sarama.SyncProducer
}

type KafkaMessage struct {
	Topic string
	Data  []byte
}

func NewKafkaBroker(addr string) (*KafkaBroker, error) {
	var err error

	for i := 0; i < 10; i++ {
		producer, err := sarama.NewSyncProducer([]string{addr}, nil)
		if err == nil {
			return &KafkaBroker{client: &producer}, nil
		}

		time.Sleep(time.Second)
	}

	return nil, err
}

// WARNING: Accepts only broker.KafkaMessage
func (k *KafkaBroker) Publish(msg any) error {
	var data KafkaMessage
	if _, ok := msg.(KafkaMessage); !ok {
		return errors.New("Invalid message type")
	}
	data = msg.(KafkaMessage)

	message := sarama.ProducerMessage{
		Topic: data.Topic,
		Value: sarama.ByteEncoder(data.Data),
	}

	_, _, err := (*k.client).SendMessage(&message)
	if err != nil {
		logrus.WithError(err).Error("Error accured while sending message to kafka")
		return err
	}

	return nil
}
