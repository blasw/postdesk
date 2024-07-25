package main

import (
	"os"
	"users-service/server"
	"users-service/storage"
	"users-service/utilities"

	"github.com/sirupsen/logrus"
)

func init() {
	check := utilities.CheckEnv([]string{"PORT", "DB_ADDR"})
	if check != "" {
		logrus.Fatal(check)
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

	dbAddr := os.Getenv("DB_ADDR")
	port := ":" + os.Getenv("PORT")

	store, err := storage.NewPostgreStore(dbAddr)
	if err != nil {
		logrus.WithError(err).Fatal("Unable to connect to the database")
	}
	logrus.Info("Storage connected")

	serv := server.NewServer(store, port)

	logrus.Info("Running...")

	if err := serv.Run(); err != nil {
		logrus.WithError(err).Fatal("Unable to start server")
	}
}
