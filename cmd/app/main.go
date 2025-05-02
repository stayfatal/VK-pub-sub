package main

import (
	"context"
	"fmt"
	"runtime"

	"github.com/stayfatal/VK-pub-sub/pkg/pubsub"
)

func main() {
	var (
		subsAmount = 5
		subject    = "test subject"
		sp         = pubsub.NewPubSub()
	)

	for i := 0; i < subsAmount; i++ {
		sp.Subscribe(subject, func(interface{}) {})
	}

	sp.Close(context.Background())

	fmt.Println(runtime.NumGoroutine())
}
