package saramatestkit

import "github.com/Shopify/sarama"

type SyncProducer struct {
	SendMessageFunc  func(msg *sarama.ProducerMessage) (partition int32, offset int64, err error)
	SendMessagesFunc func(msgs []*sarama.ProducerMessage) error
	CloseFunc        func() error
}

func NewSyncProducer() *SyncProducer {
	return &SyncProducer{
		SendMessageFunc: func(msg *sarama.ProducerMessage) (partition int32, offset int64, err error) {
			return
		},
		SendMessagesFunc: func(msgs []*sarama.ProducerMessage) error {
			return nil
		},
		CloseFunc: func() error {
			return nil
		},
	}
}

func (p SyncProducer) SendMessage(msg *sarama.ProducerMessage) (int32, int64, error) {
	return p.SendMessageFunc(msg)
}
func (p SyncProducer) SendMessages(msgs []*sarama.ProducerMessage) error {
	return p.SendMessagesFunc(msgs)
}

func (p SyncProducer) Close() error {
	return p.CloseFunc()
}
