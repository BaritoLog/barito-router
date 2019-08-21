package router

import (
	"bytes"
	"encoding/json"

	"github.com/golang/protobuf/jsonpb"
	"github.com/golang/protobuf/ptypes/struct"
	pb "github.com/vwidjaya/barito-proto/producer"
)

func ConvertBytesToTimber(b []byte, context pb.TimberContext) (timber pb.Timber, err error) {
	var content structpb.Struct

	err = jsonpb.Unmarshal(bytes.NewBuffer(b), &content)
	if err != nil {
		return
	}

	timber.Context = &context
	timber.Content = &content
	return
}

func ConvertBytesToTimberCollection(b []byte, context pb.TimberContext) (timberCol pb.TimberCollection, err error) {
	var timberColMap map[string]interface{}
	err = json.Unmarshal(b, &timberColMap)
	if err != nil {
		return
	}

	for _, timberMap := range timberColMap["items"].([]interface{}) {
		b, _ := json.Marshal(timberMap)
		timber, _ := ConvertBytesToTimber(b, pb.TimberContext{})
		timberCol.Items = append(timberCol.Items, &timber)
	}

	timberCol.Context = &context
	return
}

func TimberContextFromProfile(profile *Profile) pb.TimberContext {
	return pb.TimberContext{
		KafkaTopic:             profile.Meta.Kafka.TopicName,
		KafkaPartition:         profile.Meta.Kafka.Partition,
		KafkaReplicationFactor: int32(profile.Meta.Kafka.ReplicationFactor),
		EsIndexPrefix:          profile.Meta.Elasticsearch.IndexPrefix,
		EsDocumentType:         profile.Meta.Elasticsearch.DocumentType,
		AppMaxTps:              int32(profile.MaxTps),
		AppSecret:              profile.AppSecret,
	}
}
