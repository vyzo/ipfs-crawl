package main

import (
	"encoding/json"
	"fmt"
	"io"
	"time"
)

type PeerDialLog struct {
	Dials   []DialAttempt
	Success bool
}

type DialAttempt struct {
	TargetAddr string
	Result     string
	Error      string
	Duration   string
}

type dialLog struct {
	peers map[string]*PeerDialLog
}

func handleEvents(r io.Reader) {
	dialLog := &dialLog{
		peers: make(map[string]*PeerDialLog),
	}

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

			pdl, ok := dialLog.peers[rpeer]
			if !ok {
				pdl = new(PeerDialLog)
				dialLog.peers[rpeer] = pdl
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

			/*// TEMP
			data, err := json.Marshal(datt)
			if err != nil {
				panic(err)
			}

			fmt.Println(string(data))
			// TEMP */
		case "swarmDialAttemptSync":
			// this event tracks the entire life of a dial to a peer
			// It looks something like:
			/*
				map[duration:1.396019259e+09 event:swarmDialAttemptSync peerID:QmaCpDMGvV2BGHeYERUEnRQAwe3N8SzbUtfsmvsqQLuvuJ system:swarm2 time:2018-04-30T08:59:24.772706819Z
			*/
			// we should use this as a trigger to write out to the logfile this particular dial event

			fmt.Println(ev)
		}
	}
}
