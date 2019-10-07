package router

import (
	pb "github.com/vwidjaya/barito-proto/producer"
	"google.golang.org/grpc"
)

type producerAttributes struct {
	consulAddr   string
	producerAddr string
	producerName string
	appSecret    string
}

type grpcParts struct {
	conn   *grpc.ClientConn
	client pb.ProducerClient
}

type ProducerStore map[producerAttributes]*grpcParts

func NewProducerStore() ProducerStore {
	return make(map[producerAttributes]*grpcParts)
}

func (s ProducerStore) GetClient(attr producerAttributes) pb.ProducerClient {
	if _, ok := s[attr]; !ok {
		conn, _ := grpc.Dial(attr.producerAddr, grpc.WithInsecure())

		s[attr] = &grpcParts{
			conn:   conn,
			client: pb.NewProducerClient(conn),
		}
	}

	return s[attr].client
}

func (s ProducerStore) CloseConns() {
	for attr, parts := range s {
		delete(s, attr)
		parts.conn.Close()
	}
}
