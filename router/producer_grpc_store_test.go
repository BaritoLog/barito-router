package router

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"log"
	"os"
	"testing"

	pb "github.com/vwidjaya/barito-proto/producer"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
)

func TestBeni(t *testing.T) {
	caCert, err := os.ReadFile("/Users/beni.harmadi/jek/data/grpc_producer/ca2.crt")
	if err != nil {
		log.Fatalf("failed to read CA certificate: %v", err)
	}
	caCertPool := x509.NewCertPool()
	if !caCertPool.AppendCertsFromPEM(caCert) {
		log.Fatalf("failed to add CA certificate to pool: %v", err)
	}

	// Load client certificate and key
	clientCert, err := tls.LoadX509KeyPair("/Users/beni.harmadi/jek/data/grpc_producer/tls2.crt", "/Users/beni.harmadi/jek/data/grpc_producer/tls2.key")
	if err != nil {
		log.Fatalf("failed to load client certificate and key: %v", err)
	}

	// Create the credentials and dial options
	creds := credentials.NewTLS(&tls.Config{
		ServerName:         "producer.rada.s-go-sy-primary-gke-01.internal.barito.gtflabs.io",
		Certificates:       []tls.Certificate{clientCert},
		RootCAs:            caCertPool,
		InsecureSkipVerify: true,
	})

	url := "producer.rada.s-go-sy-primary-gke-01.internal.barito.gtflabs.io:443"
	// url = "localhost:8082"
	conn, err := grpc.Dial(url, grpc.WithTransportCredentials(creds), grpc.WithAuthority("producer.rada.s-go-sy-primary-gke-01.internal.barito.gtflabs.io"))
	if err != nil {
		panic(err)
	}

	// conn, err = grpc.Dial("localhost:8082", grpc.WithInsecure())
	// if err != nil {
	// 	panic(err)
	// }

	fmt.Println("Connection", conn)
	client := pb.NewProducerClient(conn)
	payload, err := ConvertBytesToTimberCollection(sampleRawTimberCollection(), pb.TimberContext{})
	if err != nil {
		t.Fatal(err)
	}
	result, err := client.ProduceBatch(context.Background(), &payload)
	fmt.Println("err", err)
	fmt.Println("result", result)

	t.Fail()
}
