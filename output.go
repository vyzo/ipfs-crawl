package main

import (
	"encoding/json"
	"log"
	"os"
	"sync"

	pstore "github.com/libp2p/go-libp2p-peerstore"
	ma "github.com/multiformats/go-multiaddr"
)

type CrawlLog struct {
	f *os.File
	w *json.Encoder

	mx sync.Mutex
}

type CrawlRecord struct {
	ID      string   `json:"id"`
	ConAddr string   `json:"conaddr,omitempty"`
	Addrs   []string `json:"addrs"`
	Status  string   `json:"status"`
	Error   string   `json:"error,omitempty"`
}

func NewCrawlLog(path string) (*CrawlLog, error) {
	f, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0644)
	if err != nil {
		return nil, err
	}

	enc := json.NewEncoder(f)

	return &CrawlLog{f: f, w: enc}, nil
}

func (o *CrawlLog) LogConnect(ca ma.Multiaddr, pi pstore.PeerInfo) {
	o.mx.Lock()
	defer o.mx.Unlock()

	err := o.w.Encode(peerInfoToCrawlRecord(pi, ca, "OK", ""))
	if err != nil {
		log.Fatal(err)
	}
}

func (o *CrawlLog) LogError(pi pstore.PeerInfo, e error) {
	o.mx.Lock()
	defer o.mx.Unlock()

	err := o.w.Encode(peerInfoToCrawlRecord(pi, nil, "ERROR", e.Error()))
	if err != nil {
		log.Fatal(err)
	}
}

func peerInfoToCrawlRecord(pi pstore.PeerInfo, ca ma.Multiaddr, status, e string) CrawlRecord {
	addrs := make([]string, len(pi.Addrs))
	for i, a := range pi.Addrs {
		addrs[i] = a.String()
	}

	var cas string
	if ca != nil {
		cas = ca.String()
	}

	return CrawlRecord{ID: pi.ID.Pretty(), Addrs: addrs, Status: status, Error: e, ConAddr: cas}
}
