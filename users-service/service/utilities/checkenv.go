package utilities

import (
	"os"
	"strings"
)

func CheckEnv(keys []string) string {
	msgs := []string{}

	for _, key := range keys {
		if _, ok := os.LookupEnv(key); !ok {
			msgs = append(msgs, "missing environment variable: "+key)
		}
	}

	return strings.Join(msgs, "\n")
}
