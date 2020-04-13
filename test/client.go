package test

import (
	"log"
	"net"
	"net/url"
	"strings"

	"github.com/SpeedVan/go-gesclient"
	"github.com/SpeedVan/go-gesclient/client"
)

// NewClient todo
func NewClient(
	name string,
	debug bool,
	endpoint string,
	sslHost string,
	sslSkipVerify bool,
	verbose bool,
) (client.Connection, error) {

	if debug {
		gesclient.Debug()
	}
	settingsBuilder := client.CreateConnectionSettings()

	var uri *url.URL
	var err error
	if !strings.Contains(endpoint, "://") {
		gossipSeeds := strings.Split(endpoint, ",")
		endpoints := make([]*net.TCPAddr, len(gossipSeeds))
		for i, gossipSeed := range gossipSeeds {
			endpoints[i], err = net.ResolveTCPAddr("tcp", gossipSeed)
			if err != nil {
				log.Fatalf("Error resolving: %v", gossipSeed)
			}
		}
		settingsBuilder.SetGossipSeedEndPoints(endpoints)
	} else {
		uri, err = url.Parse(endpoint)
		if err != nil {
			log.Fatalf("Error parsing address: %v", err)
		}

		if uri.User != nil {
			username := uri.User.Username()
			password, _ := uri.User.Password()
			settingsBuilder.SetDefaultUserCredentials(client.NewUserCredentials(username, password))
		}
	}

	if sslHost != "" {
		settingsBuilder.UseSslConnection(sslHost, !sslSkipVerify)
	}

	if verbose {
		settingsBuilder.EnableVerboseLogging()
	}

	cli, err := gesclient.Create(settingsBuilder.Build(), uri, name)
	if err != nil {
		return nil, err
	}

	if err := cli.ConnectAsync().Wait(); err != nil {
		return nil, err
	}

	return cli, err
}
