// Copyright 2013-2015, Homin Lee. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package subtitle

import (
	"fmt"
	"regexp"
	"sort"
	"time"
)

// Book is collection of scripts
type Book []Script

// Find a script on given timestamp.
// If not hit, it returns next script.
// So, caller should re-check the script is hit on the timestamp.
func (b Book) Find(ts time.Duration) *Script {
	si := sort.Search(len(b), func(i int) bool {
		return b[i].Start >= ts
	})

	if si >= len(b) {
		return nil
	}

	return &b[si]
}

// Script represents a script with index and start/end time
type Script struct {
	Idx        int
	Start, End time.Duration
	Text       string
}

// Duration returns how long the script should be shown
func (s *Script) Duration() time.Duration {
	return s.End - s.Start
}

// TextWithoutMarkup strips HTML markup from script
func (s *Script) TextWithoutMarkup() string {
	return reMakrup.ReplaceAllString(s.Text, "")
}

// CheckHit checks the script with given timestamp
func (s *Script) CheckHit(ts time.Duration) HitStatus {
	switch {
	case ts < s.Start:
		return ScrEARLY
	case ts >= s.Start && s.End >= s.End:
		return ScrHIT
	case s.End < ts:
		return ScrLATE
	}
	return ScrINVALID
}

func (s *Script) String() string {
	return fmt.Sprintf("%d:%s(%s-%s)", s.Idx, s.Text, s.Start, s.End)
}

// HitStatus is type for timestamp check
type HitStatus uint8

const (
	ScrINVALID HitStatus = iota
	ScrEARLY             // Not yet
	ScrHIT               // Now
	ScrLATE              // Gone
)

var reMakrup = regexp.MustCompile("</?[^<>]+?>")
