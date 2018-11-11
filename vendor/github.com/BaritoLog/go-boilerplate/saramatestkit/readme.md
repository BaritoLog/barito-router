# Sarama Test Kit

[Sarama](https://github.com/Shopify/sarama) [Monkey patching](https://github.com/bouk/monkey) collection to help to test your code 

### NewClient

```go
func MyFunction() {
  // create dummy client
  dummy := saramatestkit.NewClient()
  dummy.TopicsFunc = func() ([]string, error) {
  	return []string{"topic-01", "topic-02"}, nil 
  }
  
  // monkey patch 
  patch := saramatestkit.PatchNewClient(dummy, nil)
  defer patch.Unpatch()

  // business as usual
  client, err := sarama.NewClient(nil, nil)
  if err != nil {
    log.Fatal(log.Error())
  }

  topics, err := client.Topics()
  if err != nil {
    log.Fatal(log.Error())
  }
  
  fmt.Println(topics) // {"topic-01", "topic-02"}
	
}
```

### NewSyncProducer

```go
func MyFunction(){
  var topic string

  // mocks.NewSyncProducer() is good but sometimes is not enough.
  // saramatestkit.NewSyncProducer() is a alternative for producer mock
  dummy := saramatestkit.NewSyncProducer()
  dummy.SendMessageFunc = func(msg *sarama.ProducerMessage) (partition int32, offset int64, err error) {
  	topic = msg.Topic
  	return
  }
  
  // monkey patch
  patch := PatchNewSyncProducer(dummy, err)
  defer patch.Unpatch()
  
  // business as usual
  sarama.SendMessage(&sarama.ProducerMessage{
    Topic: "some-topic",
    Value: sarama.ByteEncoder([]byte("some-message")),
  })
  
  fmt.Println(topic) // "some-topic"
  
}

```
