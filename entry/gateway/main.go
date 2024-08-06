package main

import (
	"gateway/auth"
	"gateway/broker"
	"gateway/server"
	"gateway/utilities"
	"os"
	"time"

	"github.com/elastic/go-elasticsearch/v8"
	"github.com/sirupsen/logrus"
	"gopkg.in/go-extras/elogrus.v8"
)

func init() {
	res := utilities.CheckEnv([]string{"PORT", "AUTH_HOST", "AUTH_PORT", "BROKER_HOST", "BROKER_PORT", "ELASTIC"})
	if res != "" {
		logrus.Fatal(res)
	}
}

func initLogger() {
	var err error
	var hook *elogrus.ElasticHook

	logrus.SetLevel(logrus.DebugLevel)
	logrus.SetFormatter(&logrus.TextFormatter{
		ForceColors:   true,
		FullTimestamp: true,
	})
	logrus.SetReportCaller(true)
	logrus.SetOutput(os.Stdout)

	if os.Getenv("ELASTIC") == "false" {
		return
	}

	eHost := os.Getenv("ELASTIC_HOST")
	ePORT := os.Getenv("ELASTIC_PORT")

	cfg := elasticsearch.Config{
		Addresses: []string{"http://" + eHost + ":" + ePORT},
	}

	esClient, err := elasticsearch.NewClient(cfg)
	if err != nil {
		logrus.Fatal(err)
	}

	for i := 0; i < 40; i++ {
		hook, err = elogrus.NewAsyncElasticHook(esClient, eHost, logrus.DebugLevel, "gateway_logs")
		if err == nil {
			break
		}
		time.Sleep(time.Second)
	}

	if err != nil {
		logrus.Fatal(err)
	}

	logrus.AddHook(hook)
}

func main() {
	initLogger()

	logrus.Info("Gateway server starting...")

	port := ":" + os.Getenv("PORT")
	authAddr := os.Getenv("AUTH_HOST") + ":" + os.Getenv("AUTH_PORT")
	brokerAddr := os.Getenv("BROKER_HOST") + ":" + os.Getenv("BROKER_PORT")

	authService, err := auth.Connect(authAddr)
	if err != nil {
		logrus.WithError(err).Fatal("Error accured while connecting to auth service")
	}

	logrus.Info("Auth service connected")

	kafkaBroker, err := broker.NewKafkaClient(brokerAddr)
	if err != nil {
		logrus.WithError(err).Fatal("Error accured while connecting to kafka broker")
	}

	logrus.Info("Kafka message broker connected")

	serv, err := server.New(port, authService, kafkaBroker)
	if err != nil {
		logrus.WithError(err).Error("Error accured while initiating the server")
	}

	logrus.Info("Running...")
	if err := serv.Run(); err != nil {
		logrus.WithError(err).Error("Error accured while running the server")
		return
	}
}
