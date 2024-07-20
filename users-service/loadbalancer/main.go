package main

import (
	"os"

	"github.com/sirupsen/logrus"
)

func init() {
	required_envs := []string{"PORT", "BROKER_HOST", "BROKER_ADDR", "SERVICE1"}
	for _, env := range required_envs {
		if _, ok := os.LookupEnv(env); !ok {
			logrus.Fatal("Missing environment variable: " + env)
		}
	}
}

func main() {

}
