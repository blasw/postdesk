package main

import (
	"os"
	"time"
	"users-service/server"
	"users-service/storage"
	"users-service/utilities"

	"github.com/elastic/go-elasticsearch/v8"
	"github.com/sirupsen/logrus"
	"gopkg.in/go-extras/elogrus.v8"
)

func init() {
	check := utilities.CheckEnv([]string{"PORT", "DB_ADDR", "ELASTIC"})
	if check != "" {
		logrus.Fatal(check)
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
		hook, err = elogrus.NewAsyncElasticHook(esClient, eHost, logrus.DebugLevel, "users_service_logs")
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
