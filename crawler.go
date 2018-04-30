package main

import (
	"context"
	crand "crypto/rand"
	"encoding/base64"
	"log"
	mrand "math/rand"
	"time"

	host "github.com/libp2p/go-libp2p-host"
	dht "github.com/libp2p/go-libp2p-kad-dht"
	peer "github.com/libp2p/go-libp2p-peer"
	pstore "github.com/libp2p/go-libp2p-peerstore"
	swarm "github.com/libp2p/go-libp2p-swarm"
)

const WORKERS = 16

type Crawler struct {
	ctx context.Context
	h   host.Host
	dht *dht.IpfsDHT
	out *CrawlLog

	peers map[peer.ID]struct{}
	work  chan pstore.PeerInfo
}

func NewCrawler(ctx context.Context, h host.Host, dht *dht.IpfsDHT, out *CrawlLog) *Crawler {
	c := &Crawler{ctx: ctx, h: h, dht: dht, out: out,
		peers: make(map[peer.ID]struct{}),
		work:  make(chan pstore.PeerInfo, WORKERS),
	}

	for i := 0; i < WORKERS; i++ {
		go c.worker()
	}

	return c
}

func (c *Crawler) Crawl() {
	anchor := make([]byte, 32)
	for {

		_, err := crand.Read(anchor)
		if err != nil {
			log.Fatal(err)
		}

		str := base64.RawStdEncoding.EncodeToString(anchor)
		c.crawlFromAnchor(str)

		select {
		case <-time.After(60 * time.Second):
		case <-c.ctx.Done():
			return
		}
	}
}

func (c *Crawler) crawlFromAnchor(key string) {
	log.Printf("Crawling from anchor %s\n", key)

	ctx, cancel := context.WithTimeout(c.ctx, 60*time.Second)
	pch, err := c.dht.GetClosestPeers(ctx, key)

	if err != nil {
		log.Fatal(err)
	}

	var ps []peer.ID
	for p := range pch {
		ps = append(ps, p)
	}
	cancel()

	log.Printf("Found %d peers", len(ps))
	for _, p := range ps {
		c.crawlPeer(p)
	}
}

func (c *Crawler) crawlPeer(p peer.ID) {
	_, ok := c.peers[p]
	if ok {
		return
	}

	log.Printf("Crawling peer %s\n", p.Pretty())

	ctx, cancel := context.WithTimeout(c.ctx, 60*time.Second)
	pi, err := c.dht.FindPeer(ctx, p)
	cancel()

	if err != nil {
		log.Printf("Peer not found %s: %s", p.Pretty(), err.Error())
		return
	}

	c.peers[p] = struct{}{}
	select {
	case c.work <- pi:
	case <-c.ctx.Done():
		return
	}

	ctx, cancel = context.WithTimeout(c.ctx, 60*time.Second)
	pch, err := c.dht.FindPeersConnectedToPeer(ctx, p)

	if err != nil {
		log.Printf("Can't find peers connected to peer %s: %s", p.Pretty(), err.Error())
		cancel()
		return
	}

	var ps []peer.ID
	for pip := range pch {
		ps = append(ps, pip.ID)
	}
	cancel()

	log.Printf("Peer %s is connected to %d peers", p.Pretty(), len(ps))

	for _, p := range ps {
		c.crawlPeer(p)
	}
}

func (c *Crawler) worker() {
	for {
		select {
		case pi, ok := <-c.work:
			if !ok {
				return
			}
			// add a bit of delay to avoid connection storms
			dt := mrand.Intn(60000)
			time.Sleep(time.Duration(dt) * time.Millisecond)
			c.tryConnect(pi)

		case <-c.ctx.Done():
			return
		}
	}
}

func (c *Crawler) tryConnect(pi pstore.PeerInfo) {
	backoff := 0
	var ctx context.Context
	var cancel func()

again:
	log.Printf("Connecting to %s (%d)", pi.ID.Pretty(), len(pi.Addrs))
	ctx, cancel = context.WithTimeout(c.ctx, 60*time.Second)

	err := c.h.Connect(ctx, pi)
	cancel()

	switch {
	case err == swarm.ErrDialBackoff:
		backoff++
		if backoff < 10 {
			dt := 1000 + mrand.Intn(backoff*10000)
			log.Printf("Backing off dialing %s", pi.ID.Pretty())
			time.Sleep(time.Duration(dt) * time.Millisecond)
			goto again
		} else {
			log.Printf("FAILED to connect to %s; giving up from dial backoff", pi.ID.Pretty())
			c.out.LogError(pi, err)
		}
	case err != nil:
		log.Printf("FAILED to connect to %s: %s", pi.ID.Pretty(), err.Error())
		c.out.LogError(pi, err)
	default:
		log.Printf("CONNECTED to %s", pi.ID.Pretty())
		conns := c.h.Network().ConnsToPeer(pi.ID)
		if len(conns) == 0 {
			log.Printf("ERROR: supposedly connected, but no conns to peer", pi.ID.Pretty())
		} else {
			c.out.LogConnect(conns[0].RemoteMultiaddr(), pi)
		}
	}
}
