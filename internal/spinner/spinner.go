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

// this file contains a custom spinner implementation

const (
	cDim     = "\x1b[38;2;110;116;134m"
	cValue   = "\x1b[38;2;222;224;232m"
	cAccent  = "\x1b[38;2;150;210;224m"
	cHeading = "\x1b[38;2;198;182;255m"
	cLabel   = "\x1b[38;2;138;146;168m"
	rst      = "\x1b[0m"
)

var assemble = []string{
	cDim + "◌" + rst,
	cDim + "○" + rst,
	cValue + "◎" + rst,
	cDim + "◉" + rst,
	cAccent + "◉" + rst,
	cAccent + "●" + rst,
	cHeading + "●" + rst,
	cAccent + "●" + rst,
}

var pulse = []string{
	cAccent + "●" + rst,
	cAccent + "◉" + rst,
	cDim + "◉" + rst,
	cAccent + "◉" + rst,
	cAccent + "●" + rst,
	cHeading + "●" + rst,
	cAccent + "●" + rst,
	cAccent + "◉" + rst,
}

type Spinner struct {
	stop chan struct{}
	done chan struct{}
	once sync.Once
}

func Start(label string) *Spinner {
	s := &Spinner{
		stop: make(chan struct{}),
		done: make(chan struct{}),
	}
	go func() {
		defer close(s.done)

		for _, frame := range assemble {
			select {
			case <-s.stop:
				fmt.Fprintf(os.Stderr, "\r\033[K")
				return
			default:
				fmt.Fprintf(os.Stderr, "\r  %s  %s%s%s",
					frame, cLabel, label, rst)
				time.Sleep(500 * time.Millisecond)
			}
		}

		i := 0
		for {
			select {
			case <-s.stop:
				fmt.Fprintf(os.Stderr, "\r\033[K")
				return
			default:
				fmt.Fprintf(os.Stderr, "\r  %s  %s%s%s",
					pulse[i%len(pulse)], cLabel, label, rst)
				i++
				time.Sleep(500 * time.Millisecond)
			}
		}
	}()
	return s
}

func (s *Spinner) Stop() {
	s.once.Do(func() {
		close(s.stop)
		<-s.done
	})
}
