package spinner

import (
	"fmt"
	"os"
	"sync"
	"time"
)

const spinnerColor = "\x1b[38;5;151m"
const resetColor = "\x1b[0m"

var frames = []string{"⠋", "⠙", "⠹", "⠸", "⠼", "⠴", "⠦", "⠧", "⠇", "⠏"}

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
		i := 0
		for {
			select {
			case <-s.stop:
				fmt.Fprintf(os.Stderr, "\r\033[K")
				return
			default:
				fmt.Fprintf(
					os.Stderr,
					"\r%s%s%s %s",
					spinnerColor,
					frames[i%len(frames)],
					resetColor,
					label,
				)
				i++
				time.Sleep(80 * time.Millisecond)
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
