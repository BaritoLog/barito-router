package router

import (
	"crypto/tls"
	"crypto/x509"
	"os"
	"sync"

	pb "github.com/vwidjaya/barito-proto/producer"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
)

type producerAttributes struct {
	consulAddr          string
	producerAddr        string
	producerMtlsEnabled bool
	producerName        string
	appSecret           string
}

type grpcParts struct {
	conn   *grpc.ClientConn
	client pb.ProducerClient
}

type ProducerStore struct {
	producerStoreMap    map[producerAttributes]*grpcParts
	producerStoreMutex  *sync.RWMutex
	mtlsCertsPathPrefix string
}

func NewProducerStore() *ProducerStore {
	return &ProducerStore{
		producerStoreMap:    make(map[producerAttributes]*grpcParts),
		producerStoreMutex:  &sync.RWMutex{},
		mtlsCertsPathPrefix: os.Getenv("MTLS_CERTS_PATH_PREFIX"),
	}
}

func (s *ProducerStore) read(attr producerAttributes) (value *grpcParts, ok bool) {
	s.producerStoreMutex.RLock()
	defer s.producerStoreMutex.RUnlock()

	value, ok = s.producerStoreMap[attr]
	return
}

func (s *ProducerStore) createGrpcConnection(attr producerAttributes) (conn *grpc.ClientConn, err error) {
	if attr.producerMtlsEnabled {
		caCert, err := os.ReadFile(s.mtlsCertsPathPrefix + "/ca.crt")
		if err != nil {
			return nil, err
		}
		caCertPool := x509.NewCertPool()
		if !caCertPool.AppendCertsFromPEM(caCert) {
			return nil, err
		}

		// Load client certificate and key
		clientCert, err := tls.LoadX509KeyPair(s.mtlsCertsPathPrefix+"/tls.crt", s.mtlsCertsPathPrefix+"/tls.key")
		if err != nil {
			return nil, err
		}

		// Create the credentials and dial options
		creds := credentials.NewTLS(&tls.Config{
			ServerName:         attr.producerAddr,
			Certificates:       []tls.Certificate{clientCert},
			RootCAs:            caCertPool,
			InsecureSkipVerify: true,
		})

		conn, err = grpc.Dial(
			attr.producerAddr,
			grpc.WithTransportCredentials(creds),
			grpc.WithAuthority(attr.producerAddr),
		)
		return conn, err
	} else {
		conn, err = grpc.Dial(attr.producerAddr, grpc.WithInsecure())
	}
	return
}

func (s *ProducerStore) GetClient(attr producerAttributes) pb.ProducerClient {
	value, ok := s.read(attr)
	if !ok {
		s.producerStoreMutex.Lock()
		defer s.producerStoreMutex.Unlock()

		if value, ok = s.producerStoreMap[attr]; !ok {
			conn, _ := s.createGrpcConnection(attr)

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
