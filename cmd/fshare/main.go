package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"sync"

	util "github.com/blinkspark/go-blink-util"
	"github.com/libp2p/go-libp2p"
	"github.com/libp2p/go-libp2p-core/host"
	"github.com/libp2p/go-libp2p-core/peer"
	discovery "github.com/libp2p/go-libp2p-discovery"
	dht "github.com/libp2p/go-libp2p-kad-dht"
)

var (
	port int

	listenString string
)

func init() {
	flag.IntVar(&port, "p", 22333, "-p PORT")
	flag.Parse()

	listenString = fmt.Sprintf("/ip4/0.0.0.0/tcp/%d", port)
}

func main() {
	client, dhtDiscover, err := newClient()
	util.CheckErr(err)
	util.Ignore(client, dhtDiscover)
}

func newClient() (host.Host, *discovery.RoutingDiscovery, error) {
	h, err := libp2p.New(context.Background(), libp2p.ListenAddrStrings(listenString))
	if err != nil {
		return nil, nil, err
	}

	ipfsDHT, err := dht.New(context.Background(), h)
	if err != nil {
		log.Panic(err)
	}

	var wg sync.WaitGroup
	for _, peerAddr := range dht.DefaultBootstrapPeers {
		pi, _ := peer.AddrInfoFromP2pAddr(peerAddr)
		wg.Add(1)
		go func() {
			defer wg.Done()
			if err := h.Connect(context.Background(), *pi); err != nil {
				log.Println(err)
			}
			log.Println("Connection established with bootstrap node:", *pi)
		}()
	}
	wg.Wait()

	dhtDiscover := discovery.NewRoutingDiscovery(ipfsDHT)
	return h, dhtDiscover, nil
}
