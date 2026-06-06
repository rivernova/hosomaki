// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package ui

import (
	"bytes"
	"encoding/json"
	"strings"
)

// costum writer to capture JSON

const (
	sentinelStart = "---JSON---"
	sentinelEnd   = "---END---"
)

type SentinelWriter struct {
	w       interface{ Write([]byte) (int, error) }
	buf     bytes.Buffer
	cutting bool
}

func NewSentinelWriter(w interface{ Write([]byte) (int, error) }) *SentinelWriter {
	return &SentinelWriter{w: w}
}

func (s *SentinelWriter) Write(p []byte) (int, error) {
	if s.cutting {
		s.buf.Write(p)
		return len(p), nil
	}

	combined := s.buf.String() + string(p)

	if idx := strings.Index(combined, sentinelStart); idx >= 0 {
		visible := combined[:idx]
		if _, err := s.w.Write([]byte(visible)); err != nil {
			return 0, err
		}
		s.buf.Reset()
		s.buf.WriteString(combined[idx:])
		s.cutting = true
		return len(p), nil
	}

	safeLen := len(combined) - (len(sentinelStart) - 1)
	if safeLen <= 0 {
		s.buf.Reset()
		s.buf.WriteString(combined)
		return len(p), nil
	}

	if _, err := s.w.Write([]byte(combined[:safeLen])); err != nil {
		return 0, err
	}
	s.buf.Reset()
	s.buf.WriteString(combined[safeLen:])
	return len(p), nil
}

func (s *SentinelWriter) Flush() {
	if !s.cutting && s.buf.Len() > 0 {
		_, err := s.w.Write(s.buf.Bytes())
		if err != nil {
			return
		}
		s.buf.Reset()
	}
}

func (s *SentinelWriter) ExtractJSON() string {
	raw := s.buf.String()
	start := strings.Index(raw, sentinelStart)
	if start < 0 {
		return ""
	}
	raw = raw[start+len(sentinelStart):]
	end := strings.Index(raw, sentinelEnd)
	if end >= 0 {
		raw = raw[:end]
	}
	return strings.TrimSpace(raw)
}

type DoctorCounts struct {
	Anomalies int `json:"anomalies"`
	Actions   int `json:"actions"`
}

type StatusCounts struct {
	FailedServices   int `json:"failed_services"`
	WarnServices     int `json:"warn_services"`
	PatternsDetected int `json:"patterns_detected"`
}

type ExplainCounts struct {
	Patterns int `json:"patterns"`
	Causes   int `json:"causes"`
}

func ParseDoctorCounts(sw *SentinelWriter) DoctorCounts {
	var c DoctorCounts
	err := json.Unmarshal([]byte(sw.ExtractJSON()), &c)
	if err != nil {
		return DoctorCounts{}
	}
	return c
}

func ParseStatusCounts(sw *SentinelWriter) StatusCounts {
	var c StatusCounts
	err := json.Unmarshal([]byte(sw.ExtractJSON()), &c)
	if err != nil {
		return StatusCounts{}
	}
	return c
}

func ParseExplainCounts(sw *SentinelWriter) ExplainCounts {
	var c ExplainCounts
	err := json.Unmarshal([]byte(sw.ExtractJSON()), &c)
	if err != nil {
		return ExplainCounts{}
	}
	return c
}
