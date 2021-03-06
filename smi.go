// Copyright 2013-2015, Homin Lee. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package subtitle

import (
	"io"
	"strconv"
	"strings"
	"time"

	"golang.org/x/net/html"
)

var nextTs time.Duration
var lastBr bool

// State represents current parsing state
type State uint8

// StateMachine to parse SAMI subtitle
const (
	StateFindTag State = iota
	StateIdle
	StateSync
	StateText
	StateBr
)

func (s State) String() string {
	switch s {
	case StateFindTag:
		return "StateFindTag"
	case StateIdle:
		return "StateIdle"
	case StateSync:
		return "StateSync"
	case StateText:
		return "StateText"
	case StateBr:
		return "StateBr"
	}
	return "StateUnknown"
}

// ReadSmi read smi scripts from data stream
func ReadSmi(r io.Reader) (book Book, err error) {
	z := html.NewTokenizer(r)

	var state State
	var raw string
	var t html.Token

stateLoop:
	for {
		switch state {
		case StateFindTag:
			tt := z.Next()
			if tt == html.ErrorToken {
				ttErr := z.Err()
				if ttErr == io.EOF {
					break stateLoop
				}
				return nil, ttErr
			}

			raw = string(z.Raw())
			t = z.Token()

			// log.Printf("RAW: \"%s\"\n", raw)
			// log.Printf("TKN: %v, \"%v\"\n", t.Type, t.Data)
			// for _, v := range t.Attr {
			// 	log.Printf("  %v: %v, ", v.Key, v.Val)
			// }

			if strings.TrimSpace(raw) == "" {
				continue
			}

			// select state
			switch {
			case t.Type == html.StartTagToken && t.Data == "sync":
				state = StateSync
			case t.Type == html.TextToken:
				state = StateText
			case t.Type == html.StartTagToken && t.Data == "br":
				state = StateBr
			}
			continue

		case StateSync:
			if len(t.Attr) < 1 {
				panic("sync tag should have start attr")
			}
			ts, err := strconv.Atoi(t.Attr[0].Val)
			if err != nil {
				panic(err)
			}
			nextTs = time.Duration(ts) * time.Millisecond
			state = StateFindTag
			continue

		case StateText:
			s := strings.TrimSpace(t.Data)

			// remove html comment
			if strings.HasPrefix(s, "<!--") && strings.HasSuffix(s, "-->") {
				state = StateFindTag
				continue
			}

			if lastBr {
				lastBr = false
				if len(book) > 0 {
					ls := &book[len(book)-1]
					ls.Text += "\n"
					ls.Text += s
				}

				state = StateFindTag
				continue
			}

			// Blank &nbsp; to erase screen
			// Or last End of script is empty
			if len(book) > 0 {
				ls := &book[len(book)-1]
				if ls.End == 0 {
					ls.End = nextTs
				}
			}

			// Text with contents
			if s != "" {
				scr := Script{
					Text:  s,
					Start: nextTs,
				}
				book = append(book, scr)
			}
			state = StateFindTag
			continue

		case StateBr:
			lastBr = true
			state = StateFindTag
			continue
		}
	}

	return
}
