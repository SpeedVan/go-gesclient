package gesclient

import (
	"fmt"
	"log"
	"net"
	"net/url"
	"strings"

	"github.com/SpeedVan/go-gesclient/client"
	"github.com/SpeedVan/go-gesclient/internal"
)

func New(name string, debug bool, endpoint string, sslHost string, sslSkipVerify bool, verbose bool) (client.Connection, error) {
	if debug {
		Debug()
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

	cli, err := Create(settingsBuilder.Build(), uri, name)
	if err != nil {
		return nil, err
	}

	if err := cli.ConnectAsync().Wait(); err != nil {
		return nil, err
	}

	return cli, nil
}

func Create(settings *client.ConnectionSettings, uri *url.URL, name string) (client.Connection, error) {
	var scheme string
	var connectionSettings *client.ConnectionSettings

	if uri == nil {
		scheme = ""
	} else {
		scheme = uri.Scheme
	}

	if settings == nil {
		connectionSettings = client.DefaultConnectionSettings
	} else {
		connectionSettings = settings
	}

	credentials := getCredentialsFromUri(uri)
	if credentials != nil {
		connectionSettings.DefaultUserCredentials = credentials
	}

	var endPointDiscoverer internal.EndpointDiscoverer
	if scheme == "discover" {
		clusterSettings := client.NewClusterSettings(getUrlHostname(uri), connectionSettings.MaxDiscoverAttempts(),
			getUrlPort(uri), nil, connectionSettings.GossipTimeout())

		endPointDiscoverer = internal.NewClusterDnsEndPointDiscoverer(
			clusterSettings.ClusterDns(),
			clusterSettings.MaxDiscoverAttempts(),
			clusterSettings.ExternalGossipPort(),
			clusterSettings.GossipSeeds(),
			clusterSettings.GossipTimeout())
	} else if scheme == "tcp" || scheme == "ssl" {
		if scheme == "ssl" {
			connectionSettings = client.ConnectionSettingsBuilderFrom(connectionSettings).
				UseSslConnection(getUrlHostname(uri), true).
				Build()
		}

		tcpEndpoint, err := net.ResolveTCPAddr("tcp", uri.Host)
		if err != nil {
			return nil, err
		}
		endPointDiscoverer = internal.NewStaticEndpointDiscoverer(tcpEndpoint, connectionSettings.UseSslConnection())
	} else if connectionSettings.GossipSeeds() != nil && len(connectionSettings.GossipSeeds()) > 0 {
		clusterSettings := client.NewClusterSettings("", connectionSettings.MaxDiscoverAttempts(), 0,
			connectionSettings.GossipSeeds(), connectionSettings.GossipTimeout())

		endPointDiscoverer = internal.NewClusterDnsEndPointDiscoverer(
			clusterSettings.ClusterDns(),
			clusterSettings.MaxDiscoverAttempts(),
			clusterSettings.ExternalGossipPort(),
			clusterSettings.GossipSeeds(),
			clusterSettings.GossipTimeout())
	} else {
		return nil, fmt.Errorf("Invalid scheme for connection '%s'", scheme)
	}
	return internal.NewConnection(connectionSettings, nil, endPointDiscoverer, name), nil
}

func getCredentialsFromUri(uri *url.URL) *client.UserCredentials {
	if uri == nil || uri.User == nil {
		return nil
	}
	password, _ := uri.User.Password()
	return client.NewUserCredentials(uri.User.Username(), password)
}
