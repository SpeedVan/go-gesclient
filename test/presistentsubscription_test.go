package test

import (
	"fmt"
	"log"
	"strconv"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/SpeedVan/go-gesclient/client"
)

// Test todo
func Test(t *testing.T) {
	cli, err := NewClient(
		"test",
		false,
		"tcp://faas:123456@10.121.117.207:1113",
		"",
		false,
		false,
	)
	if err != nil {
		fmt.Println(err)
	}
	cli.Connected().Add(func(evt client.Event) error { log.Printf("Connected: %+v", evt); return nil })
	cli.Disconnected().Add(func(evt client.Event) error { log.Printf("Disconnected: %+v", evt); return nil })
	cli.Reconnecting().Add(func(evt client.Event) error { log.Printf("Reconnecting: %+v", evt); return nil })
	cli.Closed().Add(func(evt client.Event) error { log.Fatalf("Connection closed: %+v", evt); return nil })
	cli.ErrorOccurred().Add(func(evt client.Event) error { log.Printf("Error: %+v", evt); return nil })
	cli.AuthenticationFailed().Add(func(evt client.Event) error { log.Printf("Auth failed: %+v", evt); return nil })
	task, err := cli.ConnectToPersistentSubscriptionAsync(
		"params-015f6744abefd47b44132f1ee2092b30aeb3a474649deaeb1f3ad9c324aff823-3",
		"Computer",
		eventAppearedHandler,
		subscriptionDropped,
		nil,
		500,
		false,
	)
	if err != nil {
		fmt.Printf("Error occured while subscribing to stream: %v\n", err)
	} else if err := task.Error(); err != nil {
		fmt.Printf("Error occured while waiting for result of subscribing to stream: %v\n", err)
	} else {
		sub := task.Result().(client.PersistentSubscription)
		fmt.Printf("SubscribeToStream result: %+v\n", sub)
		defer func() { sub.Stop() }()
	}
	var str string
	fmt.Scanln(&str)
}

func eventAppearedHandler(s client.PersistentSubscription, r *client.ResolvedEvent) error {
	fmt.Println("获得事件：", r.Event())

	time.Sleep(10 * time.Second)
	return nil
}

func subscriptionDropped(s client.PersistentSubscription, dr client.SubscriptionDropReason, err error) error {
	return nil
}

func Test2(t *testing.T) {
	pool := NewRoutinePool(10)

	for i := 0; i <= 20; i++ {
		if i > 10 {
			time.Sleep(time.Second)
		}
		f := func(n int) func() {

			return func() {
				fmt.Println("[" + strconv.Itoa(n) + "]" + "start")
				time.Sleep(5 * time.Second)
				fmt.Println("[" + strconv.Itoa(n) + "]" + "end")
			}
		}(i)

		pool.Go(f, func() {
			fmt.Println("[" + strconv.Itoa(i) + "]" + "reachSize")
		})
	}

	var str string
	fmt.Scanln(&str)
}

type RoutinePool struct {
	size int32
	num  int32
	lock *sync.Mutex
}

func NewRoutinePool(size int32) *RoutinePool {
	return &RoutinePool{
		size: size,
		num:  0,
		lock: &sync.Mutex{},
	}
}

func (s *RoutinePool) Go(f func(), reachSize func()) {
	s.lock.Lock()
	atomic.AddInt32(&s.num, 1)
	if s.num <= s.size {
		doneChan := make(chan string)
		go func() {
			select {
			case <-doneChan: //拿到结果
				atomic.AddInt32(&s.num, -1)
			case <-time.After(30 * time.Second):
				atomic.AddInt32(&s.num, -1)
			}
			defer close(doneChan)
		}()
		go func() {
			f()
			doneChan <- "done"
		}()
	} else {
		reachSize()
	}
	s.lock.Unlock()
}
