package main

import (
	"auth/server"
	"auth/storage"
	"auth/tokenizer"
	"auth/utilities"
	"os"
	"time"

	"github.com/elastic/go-elasticsearch/v8"
	"github.com/sirupsen/logrus"
	"gopkg.in/go-extras/elogrus.v8"
)

func init() {
	res := utilities.CheckEnv([]string{"PORT", "STORAGE_HOST", "STORAGE_PORT", "JWT_SECRET", "ELASTIC"})
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
		hook, err = elogrus.NewAsyncElasticHook(esClient, eHost, logrus.DebugLevel, "auth_logs")
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

	logrus.Info("Starting auth service...")

	jwt_secret := os.Getenv("JWT_SECRET")
	port := ":" + os.Getenv("PORT")
	storage_addr := os.Getenv("STORAGE_HOST") + ":" + os.Getenv("STORAGE_PORT")

	tokenizer := tokenizer.NewJwtTokenizer(jwt_secret, time.Hour*24*7, time.Hour*24*30)
	redisStorage, err := storage.NewRedisStorage(storage_addr)
	if err != nil {
		logrus.Fatal(err)
	}

	logrus.Info("Storage connected")

	serv := server.NewServer(port, redisStorage, tokenizer)

	logrus.Info("Running...")
	if err := serv.Run(); err != nil {
		logrus.Fatal(err)
	}
}
