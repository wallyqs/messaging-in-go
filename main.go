package main

import (
	"context"
	"log"
	"time"
)

func main() {
	nc := NewClient()
	err := nc.Connect("127.0.0.1:4222")
	if err != nil {
		log.Fatalf("Error: %s", err)
	}
	defer nc.Close()

	recv := 0
	doneCh := make(chan struct{})
	nc.Subscribe("hello", func(subject, reply string, payload []byte) {
		log.Printf("[Received] %s", string(payload))
		recv++
		if recv == 20 {
			close(doneCh)
		}
	})

	nc.Subscribe(">", func(subject, reply string, payload []byte) {
		log.Printf("[Wildcard] %s", string(payload))
		recv++
		if recv == 20 {
			close(doneCh)
		}
		time.Sleep(1 * time.Second)
	})

	nc.Subscribe("hi", func(subject, reply string, payload []byte) {
		log.Printf("[Request:%s] %s", reply, string(payload))
		nc.Publish(reply, []byte("hi!"))
	})

	// Send 10 messages
	for i := 0; i < 10; i++ {
		err = nc.Publish("hello", []byte("world"))
		if err != nil {
			log.Fatalf("Error: %s", err)
		}
	}

	// Send 10 requests using context
	ctx := context.Background()
	for i := 0; i < 10; i++ {
		ctx, cancel := context.WithTimeout(ctx, 1*time.Second)
		defer cancel()
		response, err := nc.Request(ctx, "hi", []byte("hello!"))
		if err != nil {
			log.Fatalf("Error: %s", err)
		}
		log.Printf("[Response] %s", string(response))
	}

	<-doneCh
}
