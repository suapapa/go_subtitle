// Copyright 2013, Homin Lee. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package subtitle

import (
	"bufio"
	"fmt"
	"io"
	"log"
	"time"
)

// ReadSrt read srt format subtitle from data slice
func ReadSrt(r io.Reader) (Book, error) {
	var book Book
	var script Script

	const (
		StateIdx = iota
		StateTs
		StateScript
	)
	state := StateIdx

	scanner := bufio.NewScanner(r)
	for scanner.Scan() {
		line := scanner.Text()

		switch state {
		case StateIdx:
			/* log.Println("StateIdx") */
			_, err := fmt.Sscanln(line, &script.Idx)
			if err != nil {
				log.Fatalf("failed to parse index! in \"%s\" : %s",
					line, err)
			}
			state = StateTs

		case StateTs:
			/* log.Println("StateTs") */
			var sH, sM, sS, sMs int
			var eH, eM, eS, eMs int
			_, err := fmt.Sscanf(line,
				"%d:%d:%d,%d --> %d:%d:%d,%d",
				&sH, &sM, &sS, &sMs,
				&eH, &eM, &eS, &eMs)
			if err != nil {
				log.Fatalln("failed to parse timestamp!")
			}

			startMs := sMs + sS*1000 + sM*60*1000 + sH*60*60*1000
			script.Start = time.Duration(startMs) * time.Millisecond

			endMs := eMs + eS*1000 + eM*60*1000 + eH*60*60*1000
			script.End = time.Duration(endMs) * time.Millisecond

			script.Text = ""
			/* log.Println("script = ", script) */
			state = StateScript

		case StateScript:
			/* log.Println("StateScript") */
			if line == "" {
				/* log.Println("script = ", script) */
				book = append(book, script)
				state = StateIdx
			} else {
				if script.Text != "" {
					script.Text += "\n"
				}
				script.Text += line
			}
		}

	}
	if err := scanner.Err(); err != nil {
		return nil, err
	}

	/* log.Println("book = ", book) */
	return book, nil
}

// ExportToSrtFile export script book in SRT format
func ExportToSrtFile(b Book, w io.Writer) error {
	for i, s := range b {
		fmt.Fprintln(w, i+1)

		srtTime := func(d time.Duration) (h, m, s, ms int64) {
			n := d.Nanoseconds()
			// hours
			if n > 60*60*1000000000 {
				h = n / (60 * 60 * 1000000000)
				n -= h * 60 * 60 * 1000000000
			}
			// minutes
			if n > 60*1000000000 {
				m = n / (60 * 1000000000)
				n -= m * 60 * 1000000000
			}
			// seconds
			if n > 1000000000 {
				s = n / 1000000000
				n -= s * 1000000000
			}
			// milliseconds
			if n > 1000000 {
				ms = n / 1000000
			}
			return
		}

		sH, sM, sS, sMs := srtTime(s.Start)
		eH, eM, eS, eMs := srtTime(s.End)

		fmt.Fprintf(w, "%02d:%02d:%02d,%03d --> %02d:%02d:%02d,%03d\n",
			sH, sM, sS, sMs,
			eH, eM, eS, eMs,
		)
		fmt.Fprintln(w, s.Text)
		fmt.Fprintln(w, "")
	}
	return nil
}
