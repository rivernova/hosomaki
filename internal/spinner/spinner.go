// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package spinner

import (
	"fmt"
	"os"
	"time"
)

var frames = []string{"⠋", "⠙", "⠹", "⠸", "⠼", "⠴", "⠦", "⠧", "⠇", "⠏"}

type Spinner struct {
	stop chan struct{}
	done chan struct{}
}

func Start(label string) *Spinner {
	s := &Spinner{
		stop: make(chan struct{}),
		done: make(chan struct{}),
	}
	go func() {
		defer close(s.done)
		i := 0
		for {
			select {
			case <-s.stop:
				fmt.Fprintf(os.Stderr, "\r\033[K") // clear the line
				return
			default:
				fmt.Fprintf(os.Stderr, "\r%s %s", frames[i%len(frames)], label)
				i++
				time.Sleep(80 * time.Millisecond)
			}
		}
	}()
	return s
}

func (s *Spinner) Stop() {
	close(s.stop)
	<-s.done
}
