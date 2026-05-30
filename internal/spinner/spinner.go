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

// this file contains two-phase terminal spinner

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

var typewriter = []string{
	cDim + "▏" + rst,
	cAccent + "▎" + rst,
	cAccent + "▍" + rst,
	cValue + "▌" + rst,
	cAccent + "▍" + rst,
	cAccent + "▎" + rst,
	cDim + "▏" + rst,
	cDim + "·" + rst,
}

type Spinner struct {
	stop    chan struct{}
	done    chan struct{}
	writing chan string
	once    sync.Once
}

func Start(label string) *Spinner {
	s := &Spinner{
		stop:    make(chan struct{}),
		done:    make(chan struct{}),
		writing: make(chan string, 1),
	}
	go s.run(label)
	return s
}

func (s *Spinner) run(label string) {
	defer close(s.done)

	for _, frame := range assemble {
		select {
		case <-s.stop:
			fmt.Fprintf(os.Stderr, "\r\033[K")
			return
		case newLabel := <-s.writing:
			s.runWriting(newLabel)
			return
		default:
			fmt.Fprintf(os.Stderr, "\r  %s  %s%s%s", frame, cLabel, label, rst)
			time.Sleep(70 * time.Millisecond)
		}
	}

	i := 0
	for {
		select {
		case <-s.stop:
			fmt.Fprintf(os.Stderr, "\r\033[K")
			return
		case newLabel := <-s.writing:
			s.runWriting(newLabel)
			return
		default:
			fmt.Fprintf(os.Stderr, "\r  %s  %s%s%s",
				pulse[i%len(pulse)], cLabel, label, rst)
			i++
			time.Sleep(90 * time.Millisecond)
		}
	}
}

func (s *Spinner) runWriting(label string) {
	i := 0
	for {
		select {
		case <-s.stop:
			fmt.Fprintf(os.Stderr, "\r\033[K")
			return
		default:
			fmt.Fprintf(os.Stderr, "\r  %s  %s%s%s",
				typewriter[i%len(typewriter)], cLabel, label, rst)
			i++
			time.Sleep(80 * time.Millisecond)
		}
	}
}

func (s *Spinner) Writing(label string) {
	select {
	case s.writing <- label:
	default:
		// ignore if already writing
	}
}

func (s *Spinner) Stop() {
	s.once.Do(func() {
		close(s.stop)
		<-s.done
	})
}
