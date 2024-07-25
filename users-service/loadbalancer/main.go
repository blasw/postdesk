package main

import (
	"context"
	"encoding/json"
	"loadbalancer/broker"
	"loadbalancer/pb"
	"loadbalancer/users"
	"net"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/IBM/sarama"
	"github.com/sirupsen/logrus"
	"google.golang.org/protobuf/types/known/emptypb"
)

var currentInstance pb.UsersServiceClient

var availableInstances = []pb.UsersServiceClient{}

var mu = &sync.Mutex{}

func initialize() {
	required_envs := []string{"BROKER_HOST", "BROKER_PORT", "SERVICES_ADDRS"}
	for _, env := range required_envs {
		if _, ok := os.LookupEnv(env); !ok {
			logrus.Fatal("Missing environment variable: " + env)
		}
	}

	addrs := strings.Split(os.Getenv("SERVICES_ADDRS"), ",")

	wg := sync.WaitGroup{}

	for _, addr := range addrs {
		wg.Add(1)
		go func(addr string) {
			if err := tryConnect(addr); err != nil {
				logrus.WithError(err).WithField("addr", addr).Error("Unable to connect to users service")
			} else {
				logrus.WithField("addr", addr).Info("Connected to users service")
			}

			wg.Done()
		}(addr)
	}

	wg.Wait()
}

func tryConnect(addr string) error {
	var err error

	for i := 0; i < 10; i++ {
		if err = ping(addr); err == nil {
			break
		}

		time.Sleep(time.Second)
	}

	if err != nil {
		return err
	}

	instance, err := users.NewUsersService(addr)
	if err != nil {
		return err
	}

	mu.Lock()
	availableInstances = append(availableInstances, instance)
	mu.Unlock()

	logrus.Debug(len(availableInstances))

	return nil
}

func loadbalancingDaemon() {
	if len(availableInstances) == 0 {
		logrus.Fatal("No available instances")
	}

	var minLoad float32 = 0

	for {
		for _, instance := range availableInstances {
			health, err := instance.CheckHealth(context.Background(), &emptypb.Empty{})
			if err != nil {
				logrus.WithError(err).Error("Error accured while trying to check health in users service")
				continue
			}

			if minLoad == 0 || health.Health < minLoad {
				minLoad = health.Health
				currentInstance = instance
			}
		}

		time.Sleep(5 * time.Second)
	}
}

type SignUpResponse struct {
	UserId int64  `json:"user_id"`
	Error  string `json:"error"`
}

func signUpReq(kafka *broker.KafkaClient) func(msg *sarama.ConsumerMessage) {
	return func(msg *sarama.ConsumerMessage) {
		logrus.WithField("topic", msg.Topic).Debug("Message received")

		uniqueID := ""

		for _, header := range msg.Headers {
			if string(header.Key) == "UUID" {
				uniqueID = string(header.Value)
			}
		}

		if uniqueID == "" {
			logrus.Error("Missing UUID header in kafka message")
			return
		}

		jsonBytes := msg.Value

		logrus.Debug("Forwarding message to users service")
		result, err := currentInstance.SignUp(context.Background(), &pb.SignUpRequest{Json: jsonBytes})
		logrus.Debug("Received response from users service")

		if err != nil {
			logrus.WithError(err).Error("Error accured while trying to sign up in users service")
			resp := SignUpResponse{Error: err.Error()}
			respBytes, err := json.Marshal(resp)
			if err != nil {
				logrus.WithError(err).Error("Error accured while marshaling response")
				return
			}

			err = kafka.Produce("sign_up_response", respBytes, sarama.RecordHeader{Key: []byte("UUID"), Value: []byte(uniqueID)})
			if err != nil {
				logrus.WithError(err).Error("Error accured while producing response")
			}

			return
		}

		logrus.Debug("Generating response")
		resp := SignUpResponse{UserId: result.UserId}
		respBytes, err := json.Marshal(resp)
		if err != nil {
			logrus.WithError(err).Error("Error accured while marshaling response")
			return
		}

		logrus.Debug("Producing response")
		err = kafka.Produce("sign_up_response", respBytes, sarama.RecordHeader{Key: []byte("UUID"), Value: []byte(uniqueID)})
		if err != nil {
			logrus.WithError(err).Error("Error accured while producing response")
		}
	}
}

type SignInResponse struct {
	UserId int64  `json:"user_id"`
	Error  string `json:"error"`
}

func signInReq(kafka *broker.KafkaClient) func(msg *sarama.ConsumerMessage) {
	return func(msg *sarama.ConsumerMessage) {
		uniqueID := ""

		for _, header := range msg.Headers {
			if string(header.Key) == "UUID" {
				uniqueID = string(header.Value)
			}
		}

		if uniqueID == "" {
			logrus.Error("Missing UUID header in kafka message")
			return
		}

		jsonBytes := msg.Value

		result, err := currentInstance.SignIn(context.Background(), &pb.SignInRequest{Json: jsonBytes})

		if err != nil {
			logrus.WithError(err).Error("Error accured while trying to sign in in users service")
			resp := SignInResponse{Error: err.Error()}
			respBytes, err := json.Marshal(resp)
			if err != nil {
				logrus.WithError(err).Error("Error accured while marshaling response")
				return
			}

			err = kafka.Produce("sign_in_response", respBytes, sarama.RecordHeader{Key: []byte("UUID"), Value: []byte(uniqueID)})
			if err != nil {
				logrus.WithError(err).Error("Error accured while producing response")
			}

			return
		}

		resp := SignInResponse{UserId: result.UserId}
		respBytes, err := json.Marshal(resp)
		if err != nil {
			logrus.WithError(err).Error("Error accured while marshaling response")
			return
		}

		err = kafka.Produce("sign_in_response", respBytes, sarama.RecordHeader{Key: []byte("UUID"), Value: []byte(uniqueID)})
		if err != nil {
			logrus.WithError(err).Error("Error accured while producing response")
		}
	}
}

func ping(addr string) error {
	timeout := 5 * time.Second
	conn, err := net.DialTimeout("tcp", addr, timeout)
	if err != nil {
		return err
	}
	conn.Close()

	return nil
}

func main() {
	initialize()

	go loadbalancingDaemon()

	brokerAddr := os.Getenv("BROKER_HOST") + ":" + os.Getenv("BROKER_PORT")

	broker, err := broker.NewKafkaClient(brokerAddr)
	if err != nil {
		logrus.WithError(err).Fatal("Error accured while connecting to kafka broker")
	}

	logrus.Info("Kafka message broker connected")

	logrus.Info("Running kafka listeners...")
	broker.Listen("sign_in", signInReq(broker))
	broker.Listen("sign_up", signUpReq(broker))

	select {}
}
