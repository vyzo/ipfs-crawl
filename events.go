package main

import (
	"encoding/json"
	"io"
	"os"
	"time"

	logging "github.com/ipfs/go-log"
)

type PeerDialLog struct {
	Peer     string        `json:"peer"`
	Dials    []DialAttempt `json:"dials"`
	Success  bool          `json:"success"`
	Duration string        `json:"duration"`
}

type DialAttempt struct {
	TargetAddr string `json:"targetAddr"`
	Result     string `json:"result"`
	Error      string `json:"error,omitempty"`
	Duration   string `json:"duration"`
}

type EventsLogger struct {
	fi    *os.File
	enc   *json.Encoder
	peers map[string]*PeerDialLog
}

func NewEventsLogger(path string) (*EventsLogger, error) {
	f, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0644)
	if err != nil {
		return nil, err
	}

	enc := json.NewEncoder(f)

	el := &EventsLogger{
		fi:    f,
		enc:   enc,
		peers: make(map[string]*PeerDialLog),
	}

	r, w := io.Pipe()
	go el.handleEvents(r)
	logging.WriterGroup.AddWriter(w)

	return el, nil
}

func (el *EventsLogger) handleEvents(r io.Reader) {

	dec := json.NewDecoder(r)
	for {
		var ev map[string]interface{}
		if err := dec.Decode(&ev); err != nil {
			panic(err)
		}

		switch ev["event"] {
		case "connDial":
			rpeer, ok := ev["remotePeer"].(string)
			if !ok {
				panic("remotePeer field not set")
			}

			pdl, ok := el.peers[rpeer]
			if !ok {
				pdl = &PeerDialLog{Peer: rpeer}
				el.peers[rpeer] = pdl
			}

			if ev["dial"] == "success" {
				pdl.Success = true
			}

			dur := ev["duration"].(float64)
			durv := time.Duration(dur)

			var errstr string
			if erri, ok := ev["error"]; ok {
				errstr = erri.(string)
			}

			datt := DialAttempt{
				TargetAddr: ev["remoteAddr"].(string),
				Result:     ev["dial"].(string),
				Error:      errstr,
				Duration:   durv.String(),
			}

			pdl.Dials = append(pdl.Dials, datt)
		case "swarmDialAttemptSync":
			// this event tracks the entire life of a dial to a peer
			// It looks something like:
			/*
				map[duration:1.396019259e+09 event:swarmDialAttemptSync peerID:QmaCpDMGvV2BGHeYERUEnRQAwe3N8SzbUtfsmvsqQLuvuJ system:swarm2 time:2018-04-30T08:59:24.772706819Z
			*/
			// we should use this as a trigger to write out to the logfile this particular dial event
			p := ev["peerID"].(string)
			pdl, ok := el.peers[p]
			if !ok {
				break
			}
			delete(el.peers, p)
			dur := ev["duration"].(float64)
			pdl.Duration = time.Duration(dur).String()

			if err := el.enc.Encode(pdl); err != nil {
				panic(err)
			}
		}
	}
}
