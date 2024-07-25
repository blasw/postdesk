package main

import (
	"gateway/auth"
	"gateway/broker"
	"gateway/server"
	"gateway/utilities"
	"os"

	"github.com/sirupsen/logrus"
)

func init() {
	res := utilities.CheckEnv([]string{"PORT", "AUTH_HOST", "AUTH_PORT", "BROKER_HOST", "BROKER_PORT"})
	if res != "" {
		logrus.Fatal(res)
	}
}

func initLogger() {
	logrus.SetLevel(logrus.DebugLevel)
	logrus.SetFormatter(&logrus.TextFormatter{
		ForceColors:   true,
		FullTimestamp: true,
	})
	logrus.SetReportCaller(true)
	logrus.SetOutput(os.Stdout)
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

	kafkaBroker, err := broker.NewKafkaBroker(brokerAddr)
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
