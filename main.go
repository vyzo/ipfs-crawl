package main

import (
	"context"
	"io"
	"log"
	"time"

	ds "github.com/ipfs/go-datastore"
	ds_sync "github.com/ipfs/go-datastore/sync"
	logging "github.com/ipfs/go-log"
	libp2p "github.com/libp2p/go-libp2p"
	dht "github.com/libp2p/go-libp2p-kad-dht"
)

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	r, w := io.Pipe()
	go handleEvents(r)
	logging.WriterGroup.AddWriter(w)

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
