package main

import (
	"context"

	host "github.com/libp2p/go-libp2p-host"
	dht "github.com/libp2p/go-libp2p-kad-dht"
)

type Crawler struct {
	ctx context.Context
	h   host.Host
	dht *dht.IpfsDHT
}

func NewCrawler(ctx context.Context, h host.Host, dht *dht.IpfsDHT) *Crawler {
	return &Crawler{ctx: ctx, h: h, dht: dht}
}

func (c *Crawler) Crawl() {

}
