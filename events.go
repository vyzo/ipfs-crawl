package main

import (
	"encoding/json"
	"io"
	"log"
	"time"
)

type PeerDialLog struct {
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
		case "swarmDialAttemptSync":
			// this event tracks the entire life of a dial to a peer
			// It looks something like:
			/*
				map[duration:1.396019259e+09 event:swarmDialAttemptSync peerID:QmaCpDMGvV2BGHeYERUEnRQAwe3N8SzbUtfsmvsqQLuvuJ system:swarm2 time:2018-04-30T08:59:24.772706819Z
			*/
			// we should use this as a trigger to write out to the logfile this particular dial event
			p := ev["peerID"].(string)
			pdl, ok := dialLog.peers[p]
			if !ok {
				break
			}
			delete(dialLog.peers, p)
			dur := ev["duration"].(float64)
			pdl.Duration = time.Duration(dur).String()

			data, err := json.Marshal(pdl)
			if err != nil {
				panic(err)
			}

			// print to log file
			log.Println(string(data))
		}
	}
}
