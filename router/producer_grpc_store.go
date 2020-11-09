package router

import (
	"sync"

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

type ProducerStore struct {
	producerStoreMap   map[producerAttributes]*grpcParts
	producerStoreMutex *sync.RWMutex
}

func NewProducerStore() *ProducerStore {
	return &ProducerStore{
		producerStoreMap:   make(map[producerAttributes]*grpcParts),
		producerStoreMutex: &sync.RWMutex{},
	}
}

func (s *ProducerStore) GetClient(attr producerAttributes) pb.ProducerClient {
	if _, ok := s.producerStoreMap[attr]; !ok {
		conn, _ := grpc.Dial(attr.producerAddr, grpc.WithInsecure())

		s.producerStoreMutex.Lock()
		s.producerStoreMap[attr] = &grpcParts{
			conn:   conn,
			client: pb.NewProducerClient(conn),
		}
		s.producerStoreMutex.Unlock()
	}

	return s.producerStoreMap[attr].client
}

func (s *ProducerStore) CloseConns() {
	for attr, parts := range s.producerStoreMap {
		delete(s.producerStoreMap, attr)
		parts.conn.Close()
	}
}
