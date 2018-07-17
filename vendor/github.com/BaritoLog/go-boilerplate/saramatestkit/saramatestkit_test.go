package saramatestkit

import (
	"fmt"
	"testing"

	. "github.com/BaritoLog/go-boilerplate/testkit"
	"github.com/Shopify/sarama"
	"github.com/Shopify/sarama/mocks"
)

func TestPatchNewClient(t *testing.T) {
	client := NewClient()
	err := fmt.Errorf("some-error")

	patch := PatchNewClient(client, err)
	defer patch.Unpatch()

	client2, err2 := sarama.NewClient(nil, nil)
	FatalIf(t, client2 != client || err2 != err, "sarama.NewClient is not patched")
}

func TestPatchNewSyncProducer(t *testing.T) {
	producer := mocks.NewSyncProducer(t, sarama.NewConfig())
	err := fmt.Errorf("some-error")

	patch := PatchNewSyncProducer(producer, err)
	defer patch.Unpatch()

	producer2, err2 := sarama.NewSyncProducer(nil, nil)
	FatalIf(t, producer2 != producer || err2 != err, "sarama.NewSyncProducer is not patched")

}
