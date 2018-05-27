package router

import (
	"context"
	"log"

	"github.com/BaritoLog/barito-router/pb"
	"google.golang.org/grpc"
	"google.golang.org/grpc/status"
)

type kafkaPixy struct {
	addr            string
	topic           string
	group           string
	grpcConn        *grpc.ClientConn
	kafkaPixyClient pb.KafkaPixyClient
	consNAckRq      *pb.ConsNAckRq
}

type KafkaPixy interface {
	GrpcConn() *grpc.ClientConn
	KafkaPixyClient() pb.KafkaPixyClient
	Consume() (message []byte, err error)
}

func NewKafkaPixy(addr string, topic string, group string) KafkaPixy {
	k := &kafkaPixy{
		addr:  addr,
		topic: topic,
		group: group,
	}

	conn, err := grpc.Dial(addr, grpc.WithInsecure())
	if err != nil {
		log.Fatalf("did not connect: %s", err)
	}
	// defer conn.Close()

	k.grpcConn = conn

	c := pb.NewKafkaPixyClient(conn)
	k.kafkaPixyClient = c

	req := &pb.ConsNAckRq{
		Topic:   topic,
		Group:   group,
		NoAck:   true,
		AutoAck: false,
	}
	k.consNAckRq = req

	return k
}

func (k *kafkaPixy) GrpcConn() *grpc.ClientConn {
	return k.grpcConn
}

func (k *kafkaPixy) KafkaPixyClient() pb.KafkaPixyClient {
	return k.kafkaPixyClient
}

func (k *kafkaPixy) Consume() (message []byte, err error) {
	result, err := k.kafkaPixyClient.ConsumeNAck(context.Background(), k.consNAckRq)

	if err != nil {
		st, _ := status.FromError(err)
		return nil, st.Err()
	}

	message = result.Message

	return
}
