package main

import (
	"encoding/json"
	"fmt"
	"io"
)

func handleEvents(r io.Reader) {
	dec := json.NewDecoder(r)
	for {
		var ev map[string]interface{}
		if err := dec.Decode(&ev); err != nil {
			panic(err)
		}

		// deal with this thing
		fmt.Println(ev)
	}

}
