package router

import (
	"encoding/json"
)

type Profile struct {
	ID          string       `json:"id"`
	Name        string       `json:"name"`
	AppGroup    string       `json:"app_group_name"`
	MaxTps      int          `json:"max_tps"`
	ClusterName string       `json:"cluster_name"`
	ConsulHost  string       `json:"consul_host"`
	AppStatus   string       `json:"status"`
	Meta        *ProfileMeta `json:"meta"`
}

type ProfileMeta struct {
	ServiceNames  map[string]string  `json:"service_names"`
	Kafka         *KafkaMeta         `json:"kafka"`
	Elasticsearch *ElasticsearchMeta `json:"elasticsearch"`
}

type KafkaMeta struct {
	TopicName         string `json:"topic_name"`
	Partition         string `json:"partition"`
	ReplicationFactor string `json:"replication_factor"`
	ConsumerGroup     string `json:"consumer_group"`
}

type ElasticsearchMeta struct {
	IndexPrefix  string `json:"index_prefix"`
	DocumentType string `json:"document_type"`
}

func NewProfileFromBytes(b []byte) (*Profile, error) {
	var profile Profile
	err := json.Unmarshal(b, &profile)
	if err != nil {
		return nil, err
	}

	return &profile, nil
}

func (p Profile) MetaServiceName(name string) (val string, ok bool) {
	if p.Meta == nil {
		ok = false
		return
	}

	val, ok = p.Meta.ServiceNames[name]
	return
}
