// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package spinner

import (
	"fmt"
	"os"
	"sync"
	"time"
)

// simple spinner logic with label feature

var frames = []string{"⠋", "⠙", "⠹", "⠸", "⠼", "⠴", "⠦", "⠧", "⠇", "⠏"}

const (
	clearLine      = "\r\033[K"
	tickerDuration = 80 * time.Millisecond
	lavenderColor  = "\x1b[38;2;214;201;240m"
	resetColor     = "\x1b[0m"
)

type Spinner struct {
	mu    sync.Mutex
	label string
	stop  chan struct{}
	done  chan struct{}
	once  sync.Once
}

func Start(label string) *Spinner {
	s := &Spinner{
		label: label,
		stop:  make(chan struct{}),
		done:  make(chan struct{}),
	}
	go s.run()
	return s
}

func (s *Spinner) run() {
	defer close(s.done)

	ticker := time.NewTicker(tickerDuration)
	defer ticker.Stop()

	var frameIdx int

	for {
		select {
		case <-s.stop:
			s.mu.Lock()
			_, _ = fmt.Fprint(os.Stderr, clearLine)
			s.mu.Unlock()
			return
		case <-ticker.C:
			s.mu.Lock()
			label := s.label
			frame := frames[frameIdx%len(frames)]
			_, _ = fmt.Fprintf(os.Stderr, "\r%s%s %s%s", lavenderColor, frame, label, resetColor)
			s.mu.Unlock()
			frameIdx++
		}
	}
}

func (s *Spinner) SetLabel(label string) {
	s.mu.Lock()
	s.label = label
	s.mu.Unlock()
}

func (s *Spinner) ClearLine() {
	s.mu.Lock()
	_, _ = fmt.Fprint(os.Stderr, clearLine)
	s.mu.Unlock()
}

func (s *Spinner) Stop() {
	s.once.Do(func() {
		close(s.stop)
		<-s.done
	})
}
