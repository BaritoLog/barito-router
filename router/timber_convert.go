package router

import (
	"encoding/json"

	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/types/known/structpb"

	pb "github.com/bentol/barito-proto/producer"
)

func ConvertBytesToTimber(b []byte, context pb.TimberContext) (timber pb.Timber, err error) {
	var content structpb.Struct

	err = protojson.Unmarshal(b, &content)
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
		timberContent, _ := structpb.NewValue(timberMap)
		timberCol.Items = append(timberCol.Items, &pb.Timber{
			Context: &pb.TimberContext{},
			Content: timberContent.GetStructValue(),
		})
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
		DisableAppTps:          profile.DisableAppTps,
		AppGroupMaxTps:         int32(profile.AppGroupMaxTps),
	}
}
