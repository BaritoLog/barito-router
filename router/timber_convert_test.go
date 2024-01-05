package router

import (
	"testing"

	. "github.com/BaritoLog/go-boilerplate/testkit"
	"github.com/golang/protobuf/proto"
	pb "github.com/vwidjaya/barito-proto/producer"
)

func TestConvertBytesToTimber(t *testing.T) {
	rawTimber := sampleRawTimber()
	timber, _ := ConvertBytesToTimber(rawTimber, pb.TimberContext{})
	ok := proto.Equal(&timber, sampleContextlessTimber())
	FatalIf(t, !ok, "wrong timber proto generated")
}

func TestConvertBytesToTimberCollection(t *testing.T) {
	rawTimberCol := sampleRawTimberCollection()
	timberCol, _ := ConvertBytesToTimberCollection(rawTimberCol, pb.TimberContext{})
	ok := proto.Equal(&timberCol, sampleContextlessTimberCollection())
	if !ok {
		t.Log("expected: ", sampleContextlessTimberCollection())
		t.Log("actual: ", &timberCol)
		t.Fatal("wrong timber collection proto generated")
	}
}

func sampleRawTimber() []byte {
	return []byte(`{
		"location": "some-location",
		"message": "some-message"
	}`)
}

func sampleRawTimberCollection() []byte {
	return []byte(`{
		"items": [
			{
				"location": "some-location",
				"message": "some-message"
			},
			{
				"location": "some-location",
				"message": "some-message"
			}
		]
	}`)
}

func sampleContextlessTimber() *pb.Timber {
	timber := pb.SampleTimberProto()
	timber.Context = &pb.TimberContext{}
	return timber
}

func sampleContextlessTimberCollection() *pb.TimberCollection {
	return &pb.TimberCollection{
		Context: &pb.TimberContext{},
		Items: []*pb.Timber{
			sampleContextlessTimber(),
			sampleContextlessTimber(),
		},
	}
}

func BenchmarkConvertBytesToTimberCollection(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_, err := ConvertBytesToTimberCollection(sampleRawTimberCollection(), pb.TimberContext{})
		if err != nil {
			panic(err)
		}
	}
}
