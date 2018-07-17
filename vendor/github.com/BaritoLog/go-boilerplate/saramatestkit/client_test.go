package saramatestkit

import (
	"testing"

	. "github.com/BaritoLog/go-boilerplate/testkit"
	"github.com/Shopify/sarama"
)

func TestClient(t *testing.T) {
	var c interface{} = NewClient()
	client, ok := c.(sarama.Client)
	FatalIf(t, !ok, "client must implement the sarama.Client")

	conf := client.Config()
	FatalIf(t, conf != nil, "wrong client.Config()")

	broker, err := client.Controller()
	FatalIf(t, broker != nil || err != nil, "wrong client.Controller()")

	brokers := client.Brokers()
	FatalIf(t, len(brokers) > 0, "wrong client.Brokers()")

	topics, err := client.Topics()
	FatalIf(t, len(topics) > 0 || err != nil, "wrong client.Topics()")

	partitions, err := client.Partitions("topic")
	FatalIf(t, len(partitions) > 0 || err != nil, "wrong client.Partitions()")

	partitions, err = client.WritablePartitions("topic")
	FatalIf(t, len(partitions) > 0 || err != nil, "wrong client.WritablePartitions()")

	broker, err = client.Leader("topic", 0)
	FatalIf(t, broker != nil || err != nil, "wrong client.Leader()")

	replicationIDs, err := client.Replicas("topic", 0)
	FatalIf(t, len(replicationIDs) > 0 || err != nil, "wrong client.Replicas()")

	replicationIDs, err = client.InSyncReplicas("topic", 0)
	FatalIf(t, len(replicationIDs) > 0 || err != nil, "wrong client.Replicas()")

	err = client.RefreshMetadata("topics")
	FatalIf(t, err != nil, "wrong client.RefreshMetadata()")

	offset, err := client.GetOffset("topic", 0, 0)
	FatalIf(t, offset != 0 || err != nil, "wrong client.GetOffset()")

	broker, err = client.Coordinator("consumerGroup")
	FatalIf(t, broker != nil || err != nil, "wrong client.Coordinator()")

	err = client.RefreshCoordinator("consumerGroup")
	FatalIf(t, err != nil, "wrong client.RefreshCoordinator()")

	err = client.Close()
	FatalIf(t, err != nil, "wrong client.Close()")

	closed := client.Closed()
	FatalIf(t, closed != false, "wrong client.Closed()")
}
