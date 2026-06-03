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

// simple terminal spinner implementation with support for dynamic labels and RGB colors

var frames = []string{"⠋", "⠙", "⠹", "⠸", "⠼", "⠴", "⠦", "⠧", "⠇", "⠏"}

const (
	clearLine      = "\r\033[K"
	resetColor     = "\x1b[0m"
	tickerDuration = 80 * time.Millisecond
)

type Spinner struct {
	mu          sync.RWMutex
	label       string
	colorEscSeq string
	stop        chan struct{}
	done        chan struct{}
	once        sync.Once
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

func StartWithRGB(label string, r, g, b int) *Spinner {
	s := Start(label)
	s.SetRGB(r, g, b)
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
			fmt.Fprint(os.Stderr, clearLine)
			return
		case <-ticker.C:
			s.mu.RLock()
			currentLabel := s.label
			currentColor := s.colorEscSeq
			s.mu.RUnlock()

			frame := frames[frameIdx%len(frames)]
			frameIdx++

			if currentColor == "" {
				fmt.Fprintf(os.Stderr, "\r%s %s", frame, currentLabel)
			} else {
				fmt.Fprintf(os.Stderr, "\r%s%s %s%s", currentColor, frame, currentLabel, resetColor)
			}
		}
	}
}

func (s *Spinner) SetLabel(label string) {
	s.mu.Lock()
	s.label = label
	s.mu.Unlock()
}

func (s *Spinner) SetRGB(r, g, b int) {
	s.mu.Lock()
	s.colorEscSeq = fmt.Sprintf("\x1b[38;2;%d;%d;%dm", r, g, b)
	s.mu.Unlock()
}

func (s *Spinner) Stop() {
	s.once.Do(func() {
		close(s.stop)
		<-s.done
	})
}
