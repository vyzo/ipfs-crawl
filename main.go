package main

import (
	"context"
	"log"
	"time"

	ds "github.com/ipfs/go-datastore"
	ds_sync "github.com/ipfs/go-datastore/sync"
	libp2p "github.com/libp2p/go-libp2p"
	dht "github.com/libp2p/go-libp2p-kad-dht"
)

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	_, err := NewEventsLogger("crawl-events.json")
	if err != nil {
		log.Fatal(err)
	}

	out, err := NewCrawlLog("ipfs-crawl.out")
	if err != nil {
		log.Fatal(err)
	}

	h, err := libp2p.New(ctx)
	if err != nil {
		log.Fatal(err)
	}

	log.Printf("I am %s\n", h.ID().Pretty())

	dstore := ds_sync.MutexWrap(ds.NewMapDatastore())
	dht := dht.NewDHTClient(ctx, h, dstore)

	c := NewCrawler(ctx, h, dht, out)

	err = c.Bootstrap()
	if err != nil {
		log.Fatal(err)
	}

	// wait a bit for the bootstrap
	log.Printf("Waiting a minute for DHT bootstrap...")
	time.Sleep(1 * time.Minute)
	log.Printf("GO crawler GO!")

	c.Crawl()
}
