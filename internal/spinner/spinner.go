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

// spinner is a simple terminal spinner that runs in a separate goroutine

var frames = []string{"⠋", "⠙", "⠹", "⠸", "⠼", "⠴", "⠦", "⠧", "⠇", "⠏"}

type Spinner struct {
	label chan string
	stop  chan struct{}
	done  chan struct{}
	once  sync.Once
}

func Start(label string) *Spinner {
	s := &Spinner{
		label: make(chan string, 1),
		stop:  make(chan struct{}),
		done:  make(chan struct{}),
	}
	go func() {
		defer close(s.done)
		current := label
		i := 0
		for {
			select {
			case <-s.stop:
				fmt.Fprintf(os.Stderr, "\r\033[K")
				return
			case l := <-s.label:
				current = l
			default:
				fmt.Fprintf(os.Stderr, "\r%s %s", frames[i%len(frames)], current)
				i++
				time.Sleep(80 * time.Millisecond)
			}
		}
	}()
	return s
}

func (s *Spinner) SetLabel(label string) {
	select {
	case s.label <- label:
	default:
		select {
		case <-s.label:
		default:
		}
		s.label <- label
	}
}

func (s *Spinner) Stop() {
	s.once.Do(func() {
		close(s.stop)
		<-s.done
	})
}
