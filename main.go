package main

import (
	"context"
	"log"

	ds "github.com/ipfs/go-datastore"
	ds_sync "github.com/ipfs/go-datastore/sync"
	libp2p "github.com/libp2p/go-libp2p"
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

	c := NewCrawler(ctx, h, dht)

	err = c.Bootstrap()
	if err != nil {
		log.Fatal(err)
	}

	c.Crawl()
}
