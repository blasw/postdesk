package main

import (
	"context"
	"loadbalancer/broker"
	"loadbalancer/pb"
	"loadbalancer/users"
	"os"
	"strings"
	"time"

	"github.com/IBM/sarama"
	"github.com/sirupsen/logrus"
)

var currentInstance pb.UsersServiceClient

var availableInstances = []pb.UsersServiceClient{}

func init() {
	required_envs := []string{"PORT", "BROKER_HOST", "BROKER_ADDR", "SERVICES_ADDRS"}
	for _, env := range required_envs {
		if _, ok := os.LookupEnv(env); !ok {
			logrus.Fatal("Missing environment variable: " + env)
		}
	}

	addrs := strings.Split(os.Getenv("SERVICES_ADDRS"), ",")

	for _, addr := range addrs {
		client, err := users.NewUsersService(addr)
		if err != nil {
			logrus.WithError(err).WithField("addr", addr).Error("Unable to connect to users service")
		}
		availableInstances = append(availableInstances, client)
	}
}

func loadbalancingDaemon() {
	if len(availableInstances) == 0 {
		logrus.Fatal("No available instances")
	}

	var minLoad float64

	for {
		for _, instance := range availableInstances {
			health, err := instance.CheckHealth(context.Background(), nil)
			if err != nil {
				logrus.WithError(err).Error("Error accured while trying to check health in users service")
				continue
			}

			if minLoad == 0 || float64(health.Health) < minLoad {
				minLoad = float64(health.Health)
				currentInstance = instance
			}
		}

		time.Sleep(time.Second)
	}
}

func signInReq(msg *sarama.ConsumerMessage) {
	resp, err := currentInstance.SignIn(context.Background(), &pb.SignInRequest{})

	if err != nil {
		logrus.WithError(err).Error("Error accured while trying to sign in in users service")
		return
	}

	logrus.WithField("id", resp.UserId).Info("User signed in")
}

func signUpReq(msg *sarama.ConsumerMessage) {
	resp, err := currentInstance.SignUp(context.Background(), &pb.SignUpRequest{})

	if err != nil {
		logrus.WithError(err).Error("Error accured while trying to sign up in users service")
		return
	}

	logrus.WithField("id", resp.UserId).Info("User signed up")
}

func main() {
	go loadbalancingDaemon()

	brokerAddr := os.Getenv("BROKER_HOST") + ":" + os.Getenv("BROKER_ADDR")

	broker, err := broker.NewKafkaConsumer(brokerAddr)
	if err != nil {
		logrus.WithError(err).Fatal("Error accured while connecting to kafka broker")
	}

	logrus.Info("Kafka message broker connected")

	logrus.Info("Running kafka listeners...")
	broker.Listen("sign_in", signInReq)
	broker.Listen("sign_up", signUpReq)

	select {}
}
