// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package watcher

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"os/exec"
	"strconv"
	"strings"
	"time"
)

// manages the follow lifecycle

const (
	defaultSeedLines = 20
	shutdownGrace    = 2 * time.Second
	tickInterval     = 500 * time.Millisecond
)

type FlushFunc func(ctx context.Context, text string) error

type Config struct {
	Service string

	SeedLines int

	Buffer BufferConfig

	Sanitise func(raw string) string

	OnFlush FlushFunc

	OnLine func(raw string)
}

type Watcher struct {
	cfg Config
	buf *lineBuffer
}

func New(cfg Config) (*Watcher, error) {
	if strings.TrimSpace(cfg.Service) == "" {
		return nil, fmt.Errorf("watcher: service name must not be empty")
	}
	if cfg.Sanitise == nil {
		return nil, fmt.Errorf("watcher: Sanitise function must not be nil")
	}
	if cfg.OnFlush == nil {
		return nil, fmt.Errorf("watcher: OnFlush function must not be nil")
	}
	if cfg.Buffer.SilenceWindow <= 0 {
		cfg.Buffer.SilenceWindow = DefaultBufferConfig().SilenceWindow
	}
	if cfg.Buffer.MaxLines <= 0 {
		cfg.Buffer.MaxLines = DefaultBufferConfig().MaxLines
	}
	if cfg.SeedLines < 0 {
		cfg.SeedLines = 0
	}
	return &Watcher{cfg: cfg, buf: newLineBuffer(cfg.Buffer)}, nil
}

func (w *Watcher) Run(ctx context.Context) error {
	cmd, stdout, err := w.startJournalctl()
	if err != nil {
		return fmt.Errorf("watcher: start journalctl: %w", err)
	}

	cmdDone := make(chan error, 1)
	go func() { cmdDone <- cmd.Wait() }()

	lineCh := make(chan string, 256)
	scanDone := make(chan struct{})
	go w.scanLines(stdout, lineCh, scanDone)

	ticker := time.NewTicker(tickInterval)
	defer ticker.Stop()

	var runErr error

loop:
	for {
		select {
		case <-ctx.Done():
			break loop

		case line, ok := <-lineCh:
			if !ok {
				break loop
			}
			w.ingestLine(line)
			if reason, ok := w.buf.shouldFlush(); ok {
				if err := w.flush(ctx, reason); err != nil && ctx.Err() == nil {
					runErr = err
					break loop
				}
			}

		case <-ticker.C:
			if _, ok := w.buf.shouldFlush(); ok {
				if err := w.flush(ctx, flushReasonSilence); err != nil && ctx.Err() == nil {
					runErr = err
					break loop
				}
			}
		}
	}

	// Drain remaining lines if the context was cancelled
	if ctx.Err() != nil && w.buf.len() > 0 {
		_ = w.flush(context.Background(), flushReasonShutdown)
	}

	w.stopProcess(cmd, cmdDone)
	<-scanDone

	if ctx.Err() != nil {
		return nil
	}
	return runErr
}

func (w *Watcher) startJournalctl() (*exec.Cmd, io.ReadCloser, error) {
	seedN := w.cfg.SeedLines
	if seedN == 0 {
		seedN = defaultSeedLines
	}

	args := []string{
		"-u", w.cfg.Service,
		"--follow",
		"--lines", strconv.Itoa(seedN),
		"--no-pager",
		"--no-hostname",
		"-o", "short-monotonic",
	}

	cmd := exec.Command("journalctl", args...)

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return nil, nil, fmt.Errorf("stdout pipe: %w", err)
	}

	if err := cmd.Start(); err != nil {
		return nil, nil, fmt.Errorf("start: %w", err)
	}

	return cmd, stdout, nil
}

func (w *Watcher) scanLines(r io.Reader, lineCh chan<- string, done chan<- struct{}) {
	defer close(lineCh)
	defer close(done)

	scanner := bufio.NewScanner(r)
	scanner.Buffer(make([]byte, 256*1024), 256*1024)

	for scanner.Scan() {
		line := scanner.Text()
		if strings.TrimSpace(line) == "" {
			continue
		}
		lineCh <- line
	}
}

func (w *Watcher) ingestLine(raw string) {
	if w.cfg.OnLine != nil {
		w.cfg.OnLine(raw)
	}
	classified := w.cfg.Sanitise(raw)
	if strings.TrimSpace(classified) == "" {
		return
	}
	w.buf.add(classified)
}

func (w *Watcher) flush(ctx context.Context, _ flushReason) error {
	text, actionable := w.buf.drain()
	if !actionable || strings.TrimSpace(text) == "" {
		return nil
	}
	return w.cfg.OnFlush(ctx, text)
}

func (w *Watcher) stopProcess(cmd *exec.Cmd, cmdDone <-chan error) {
	if cmd.Process == nil {
		return
	}
	_ = cmd.Process.Signal(interruptSignal())

	select {
	case <-cmdDone:
		// exited cleanly
	case <-time.After(shutdownGrace):
		_ = cmd.Process.Kill()
		<-cmdDone
	}
}
