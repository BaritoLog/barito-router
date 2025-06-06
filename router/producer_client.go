package router

import (
	"bytes"
	"errors"
	"net/http"
	"strconv"
	"time"
)

const (
	HEADER_APP_MAX_TPS       = "X-App-Max-TPS"
	HEADER_APP_GROUP_MAX_TPS = "X-App-Group-Max-TPS"
	HEADER_DISABLE_APP_TPS   = "X-Disable-App-TPS"
	HEADER_KAFKA_TOPIC_NAME  = "X-Kafka-Topic-Name"
)

type ProducerClient interface {
	Send(*http.Request, []byte, *Profile) (*http.Response, error)
}

type producerClient struct {
	client *http.Client
}

func NewProducerClientFromEnv() ProducerClient {
	return &producerClient{
		client: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

func (p *producerClient) Send(req *http.Request, reqBody []byte, profile *Profile) (*http.Response, error) {
	if profile == nil {
		return nil, errors.New("nil profile provided to producer client")
	}

	// create request
	reader := bytes.NewReader(reqBody)
	newReq, err := http.NewRequest(req.Method, profile.ProducerAddress+req.URL.Path, reader)
	if err != nil {
		return nil, errors.New("failed to create new request: " + err.Error())
	}

	// copy headers from original request
	for key, values := range req.Header {
		for _, value := range values {
			newReq.Header.Add(key, value)
		}
	}

	// add headers from profile for context in the producer
	newReq.Header.Set(HEADER_KAFKA_TOPIC_NAME, profile.Meta.Kafka.TopicName)
	newReq.Header.Set(HEADER_APP_MAX_TPS, strconv.Itoa(profile.MaxTps))
	newReq.Header.Set(HEADER_APP_GROUP_MAX_TPS, strconv.Itoa(profile.AppGroupMaxTps))
	if profile.DisableAppTps {
		newReq.Header.Set(HEADER_DISABLE_APP_TPS, "true")
	}

	// execute request
	resp, err := p.client.Do(newReq)
	if err != nil {
		return nil, errors.New("failed to send request to producer: " + err.Error())
	}
	if resp.StatusCode != http.StatusOK {
		return nil, errors.New("failed to send request, status code: " + resp.Status)
	}

	return resp, nil
}
