package main

import (
	"log"
	"testing"
	"time"

	"github.com/nats-io/go-nats"
)

func Benchmark_Subscribe_Basic(b *testing.B) {
	b.StopTimer()
	b.StartTimer()

	nc := NewClient()
	err := nc.Connect("127.0.0.1:4222")
	if err != nil {
		log.Fatalf("Error: %s", err)
	}
	defer nc.Close()

	recv := 0
	total := 1000000
	doneCh := make(chan struct{})
	nc.Subscribe("a", func(subject, reply string, payload []byte) {
		recv++
		if recv == total {
			close(doneCh)
		}
	})

	go func() {
		nc2, err := nats.Connect("nats://127.0.0.1:4222")
		if err != nil {
			log.Fatalf("Error: %s", err)
		}
		payload := []byte("a")
		for i := 0; i < total; i++ {
			nc2.Publish("a", payload)
			if i%1000 == 0 {
				time.Sleep(1 * time.Nanosecond)
			}
		}
		nc2.Close()
	}()

	<-doneCh
	b.StopTimer()
}

func Benchmark_Publish_Basic(b *testing.B) {
	b.StopTimer()
	b.StartTimer()

	nc := NewClient()
	err := nc.Connect("127.0.0.1:4222")
	if err != nil {
		log.Fatalf("Error: %s", err)
	}
	defer nc.Close()

	total := 1000000
	payload := []byte("a")
	for i := 0; i < total; i++ {
		nc.Publish("a", payload)
		if i%1000 == 0 {
			time.Sleep(1 * time.Nanosecond)
		}
	}

	b.StopTimer()
}

func Benchmark_Subscribe_Pure(b *testing.B) {
	b.StopTimer()
	b.StartTimer()

	nc, err := nats.Connect("nats://127.0.0.1:4222")
	if err != nil {
		log.Fatalf("Error: %s", err)
	}
	defer nc.Close()

	recv := 0
	total := 1000000
	doneCh := make(chan struct{})
	nc.Subscribe("a", func(*nats.Msg) {
		recv++
		if recv == total {
			close(doneCh)
		}
	})

	go func() {
		nc2, err := nats.Connect("nats://127.0.0.1:4222")
		if err != nil {
			log.Fatalf("Error: %s", err)
		}
		payload := []byte("a")
		for i := 0; i < total; i++ {
			nc2.Publish("a", payload)
			if i%1000 == 0 {
				time.Sleep(1 * time.Nanosecond)
			}
		}
		nc2.Close()
	}()

	<-doneCh
	b.StopTimer()
}

func Benchmark_Publish_Pure(b *testing.B) {
	b.StopTimer()
	b.StartTimer()

	nc, err := nats.Connect("nats://127.0.0.1:4222")
	if err != nil {
		log.Fatalf("Error: %s", err)
	}
	defer nc.Close()

	total := 1000000
	payload := []byte("a")
	for i := 0; i < total; i++ {
		nc.Publish("a", payload)
		if i%1000 == 0 {
			time.Sleep(1 * time.Nanosecond)
		}
	}

	b.StopTimer()
}
