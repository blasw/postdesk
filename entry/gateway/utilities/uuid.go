package utilities

import "github.com/google/uuid"

func GetUUID() string {
	return uuid.New().String()
}