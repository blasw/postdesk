package main

import (
	"context"
	"errors"
	"fmt"
	"loadbalancer/broker"
	"loadbalancer/pb"
	"loadbalancer/users"
	"net"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/elastic/go-elasticsearch/v8"
	"github.com/sirupsen/logrus"
	"google.golang.org/protobuf/types/known/emptypb"
	"gopkg.in/go-extras/elogrus.v8"
)

var currentInstance pb.UsersServiceClient

var availableInstances = []pb.UsersServiceClient{}

var mu = &sync.Mutex{}

var (
	INTERNAL_ERR = errors.New(fmt.Sprint(http.StatusInternalServerError))
	BAD_REQ_ERR  = errors.New(fmt.Sprint(http.StatusBadRequest))
)

func initialize() {
	required_envs := []string{"BROKER_HOST", "BROKER_PORT", "SERVICES_ADDRS", "ELASTIC"}
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

func ping(addr string) error {
	timeout := 5 * time.Second
	conn, err := net.DialTimeout("tcp", addr, timeout)
	if err != nil {
		return err
	}
	conn.Close()

	return nil
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
		hook, err = elogrus.NewAsyncElasticHook(esClient, eHost, logrus.DebugLevel, "users_loadbalancer_logs")
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
