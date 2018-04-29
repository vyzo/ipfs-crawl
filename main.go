package main

import (
	"context"
	"log"

	ds "github.com/ipfs/go-datastore"
	ds_sync "github.com/ipfs/go-datastore/sync"
	libp2p "github.com/libp2p/go-libp2p"
	host "github.com/libp2p/go-libp2p-host"
	dht "github.com/libp2p/go-libp2p-kad-dht"
)

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	h, err := libp2p.New(ctx)
	if err != nil {
		log.Fatal(err)
	}

	log.Printf("I am %s\n", h.ID().Pretty())

	dstore := ds_sync.MutexWrap(ds.NewMapDatastore())
	dht := dht.NewDHTClient(ctx, h, dstore)

	c := &crawler{ctx: ctx, h: h, dht: dht}

	err = c.bootstrap()
	if err != nil {
		log.Fatal(err)
	}

	c.crawl()
}

type crawler struct {
	ctx context.Context
	h   host.Host
	dht *dht.IpfsDHT
}

func (c *crawler) bootstrap() error {
	return nil
}

func (c *crawler) crawl() {

}
