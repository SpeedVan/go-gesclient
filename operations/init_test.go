package operations_test

import (
	"net/url"
	"time"

	"github.com/SpeedVan/go-gesclient"
	"github.com/SpeedVan/go-gesclient/client"
)

var es client.Connection

func init() {
	ensureConnection()
}

func ensureConnection() {
	if es != nil {
		return
	}

	var err error

	uri, _ := url.Parse("tcp://faas:123456@10.121.117.207:1113/")
	es, err = gesclient.Create(client.DefaultConnectionSettings, uri, "benchmark")
	if err != nil {
		panic(err)
	}
	es.Disconnected().Add(func(event client.Event) error { panic("disconnected") })
	es.ConnectAsync().Wait()
	time.Sleep(100 * time.Millisecond)
}
