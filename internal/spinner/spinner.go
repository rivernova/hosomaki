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

// simple terminal spinner implementation with label support

var frames = []string{"⠋", "⠙", "⠹", "⠸", "⠼", "⠴", "⠦", "⠧", "⠇", "⠏"}

const (
	clearLine      = "\r\033[K"
	tickerDuration = 80 * time.Millisecond
	lavenderColor  = "\x1b[38;2;214;201;240m"
	resetColor     = "\x1b[0m"
)

type Spinner struct {
	mu    sync.RWMutex
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
			_, err := fmt.Fprint(os.Stderr, clearLine)
			if err != nil {
				return
			}
			return
		case <-ticker.C:
			s.mu.RLock()
			currentLabel := s.label
			s.mu.RUnlock()

			frame := frames[frameIdx%len(frames)]
			frameIdx++

			_, err := fmt.Fprintf(os.Stderr, "\r%s%s %s%s", lavenderColor, frame, currentLabel, resetColor)
			if err != nil {
				return
			}
		}
	}
}

func (s *Spinner) SetLabel(label string) {
	s.mu.Lock()
	s.label = label
	s.mu.Unlock()
}

func (s *Spinner) Stop() {
	s.once.Do(func() {
		close(s.stop)
		<-s.done
	})
}
