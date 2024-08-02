package main

import (
	"context"
	"encoding/json"
	"loadbalancer/broker"
	"loadbalancer/pb"

	"github.com/IBM/sarama"
	"github.com/sirupsen/logrus"
)

type SignUpResponse struct {
	UserId int64 `json:"user_id"`
	Error  int32 `json:"error"`
}

func signUpReq(kafka *broker.KafkaClient) func(msg *sarama.ConsumerMessage) {
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

		result, err := currentInstance.SignUp(context.Background(), &pb.SignUpRequest{Json: jsonBytes})
		if err != nil {
			logrus.WithError(err).Error("Error accured while trying to sign up in users service")
			return
		}
		resp := SignUpResponse{UserId: result.UserId, Error: result.Status}
		respBytes, err := json.Marshal(resp)
		if err != nil {
			logrus.WithError(err).Error("Error accured while marshaling response")
			return
		}

		err = kafka.Produce("sign_up_response", respBytes, sarama.RecordHeader{Key: []byte("UUID"), Value: []byte(uniqueID)})
		if err != nil {
			logrus.WithError(err).Error("Error accured while producing response")
		}
	}
}

type SignInResponse struct {
	UserId int64 `json:"user_id"`
	Error  int32 `json:"error"`
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
			return
		}
		resp := SignInResponse{UserId: result.UserId, Error: result.Status}

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
