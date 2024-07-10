package main

import (
	"auth/server"
	"auth/storage"
	"auth/tokenizer"
	"auth/utilities"
	"os"
	"time"

	"github.com/sirupsen/logrus"
)

func init() {
	res := utilities.CheckEnv([]string{"PORT", "STORAGE_HOST", "STORAGE_PORT", "JWT_SECRET"})
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

	logrus.Info("Starting auth service...")

	jwt_secret := os.Getenv("JWT_SECRET")
	port := ":" + os.Getenv("PORT")
	storage_addr := os.Getenv("STORAGE_HOST") + ":" + os.Getenv("STORAGE_PORT")
	logrus.Debug("Storage address: ", storage_addr)

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
