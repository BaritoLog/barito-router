package saramatestkit

import (
	"testing"

	. "github.com/BaritoLog/go-boilerplate/testkit"
	"github.com/Shopify/sarama"
)

func TestNewSyncProduer(t *testing.T) {
	var p interface{} = NewSyncProducer()
	producer, ok := p.(sarama.SyncProducer)
	FatalIf(t, !ok, "producer must implement the sarama.SyncProducer")

	partition, offset, err := producer.SendMessage(nil)
	FatalIf(t, partition != 0 || offset != 0 || err != nil, "wrong producer.SendMessage()")
	FatalIf(t, producer.SendMessages(nil) != nil, "wrong producer.SendMessages()")
	FatalIf(t, producer.Close() != nil, "wrong producer.Close()")
}
