package router

import (
	"encoding/json"
)

type Profile struct {
	ID                   int         `json:"id"`
	Name                 string      `json:"name"`
	AppSecret            string      `json:"app_secret"`
	AppGroup             string      `json:"app_group_name"`
	AppGroupSecret       string      `json:"app_group_secret"`
	MaxTps               int         `json:"max_tps"`
	AppGroupMaxTps       int         `json:"app_group_max_tps"`
	DisableAppTps        bool        `json:"disable_app_tps"`
	ClusterName          string      `json:"cluster_name"`
	ConsulHost           string      `json:"consul_host"`
	ConsulHosts          []string    `json:"consul_hosts"`
	ProducerAddress      string      `json:"producer_address"`
	ProducerMtlsEnabled  bool        `json:"producer_mtls_enabled"`
	ProducerLocation     string      `json:"producer_location"`
	KibanaAddress        string      `json:"kibana_address"`
	KibanaMtlsEnabled    bool        `json:"kibana_mtls_enabled"`
	KibanaLocation       string      `json:"kibana_location"`
	ElasticsearchAddress string      `json:"elasticsearch_address"`
	ElasticsearchStatus  string      `json:"elasticsearch_status"`
	AppStatus            string      `json:"status"`
	Meta                 ProfileMeta `json:"meta"`
}

type ProfileMeta struct {
	ServiceNames  map[string]string `json:"service_names"`
	Kafka         KafkaMeta         `json:"kafka"`
	Elasticsearch ElasticsearchMeta `json:"elasticsearch"`
}

type KafkaMeta struct {
	TopicName         string `json:"topic_name"`
	Partition         int32  `json:"partition"`
	ReplicationFactor int16  `json:"replication_factor"`
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

	if profile.ConsulHosts == nil {
		profile.ConsulHosts = []string{profile.ConsulHost}
	}
	return &profile, nil
}

func (p Profile) MetaServiceName(name string) (val string, ok bool) {
	val, ok = p.Meta.ServiceNames[name]
	return
}
