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

func (s *ProducerStore) read(attr producerAttributes) (value *grpcParts, ok bool) {
	s.producerStoreMutex.RLock()
	defer s.producerStoreMutex.RUnlock()

	value, ok = s.producerStoreMap[attr]
	return
}

func (s *ProducerStore) GetClient(attr producerAttributes) pb.ProducerClient {
	value, ok := s.read(attr)
	if !ok {
		s.producerStoreMutex.Lock()
		defer s.producerStoreMutex.Unlock()

		if value, ok = s.producerStoreMap[attr]; !ok {
			conn, _ := grpc.Dial(attr.producerAddr, grpc.WithInsecure())

			value = &grpcParts{
				conn:   conn,
				client: pb.NewProducerClient(conn),
			}
			s.producerStoreMap[attr] = value
		}
	}

	return value.client
}

func (s *ProducerStore) CloseConns() {
	for attr, parts := range s.producerStoreMap {
		delete(s.producerStoreMap, attr)
		parts.conn.Close()
	}
}
