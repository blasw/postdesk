package logs

import (
	"context"
	"encoding/json"
	"errors"
	"strings"
	"time"

	"github.com/elastic/go-elasticsearch/v8"
	"github.com/elastic/go-elasticsearch/v8/esapi"
	"github.com/sirupsen/logrus"
)

type ElasticHook struct {
	client *elasticsearch.Client
	index  string
	ctx    context.Context
}

func NewElasticHook(client *elasticsearch.Client, index string) *ElasticHook {
	return &ElasticHook{
		client: client,
		index:  index,
		ctx:    context.Background(),
	}
}

func (hook *ElasticHook) Fire(entry *logrus.Entry) error {
	entry.Data["timestamp"] = entry.Time.Format(time.RFC3339)
	logData, err := json.Marshal(entry.Data)
	if err != nil {
		return err
	}

	req := esapi.IndexRequest{
		Index:   hook.index,
		Body:    strings.NewReader(string(logData)),
		Refresh: "true",
	}
	res, err := req.Do(hook.ctx, hook.client)
	if err != nil {
		return err
	}
	defer res.Body.Close()

	if res.IsError() {
		return errors.New("error indexing log entry")
	}

	return nil
}

func (hook *ElasticHook) Levels() []logrus.Level {
	return logrus.AllLevels
}
