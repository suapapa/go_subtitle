package subtitle

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"strconv"
	"strings"
	"time"

	"golang.org/x/net/html"
)

var lastTs time.Duration
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
func ReadSmi(data []byte) (book Book) {
	b := bytes.NewBuffer(data)
	z := html.NewTokenizer(b)

	var state State
	var r string
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
				// TODO: return error
				fmt.Print("Error while tokenize:", ttErr)
				return nil
			}

			r = string(z.Raw())
			t = z.Token()

			//log.Printf("Raw: \"%s\"\n", r)
			//log.Printf("TKN: %v, \"%v\"\n", t.Type, t.Data)
			// for _, v := range t.Attr {
			// 	log.Printf("  %v: %v, ", v.Key, v.Val)
			// }

			// Don't care following tag tokens
			switch {
			case strings.TrimSpace(r) == "":
				continue
			case t.Type == html.EndTagToken:
				continue
			case t.Type == html.StartTagToken:
				if t.Data == "font" || t.Data == "p" {
					continue
				}
			case t.Type == html.CommentToken:
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
			lastTs = nextTs
			nextTs = time.Duration(ts) * time.Millisecond
			state = StateFindTag
			continue

		case StateText:
			s := strings.TrimSpace(t.Data)

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

// ReadSmiFile read smi scripts from a file
func ReadSmiFile(filename string) Book {
	data, err := ioutil.ReadFile(filename)
	if err != nil {
		//log.Fatalln("faile to read file, ", filename)
	}

	// skip UTF-8 BOM if exists
	if bytes.Equal(data[:3], []byte{0xEF, 0xBB, 0xBF}) {
		data = data[3:]
	}

	return ReadSmi(data)
}
