package router

type TimberContext struct {
	KafkaTopic             string `json:"kafka_topic"`
	KafkaPartition         int32  `json:"kafka_partition"`
	KafkaReplicationFactor int16  `json:"kafka_replication_factor"`
	ESIndexPrefix          string `json:"es_index_prefix"`
	ESDocumentType         string `json:"es_document_type"`
	AppMaxTPS              int    `json:"app_max_tps"`
}
