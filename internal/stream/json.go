// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package stream

import (
	"strings"
)

// json stream parsing utilities for extracting objects from a stream of JSON text

type ArrayItemScanner struct {
	onItem func(key, rawValue string)

	buf         strings.Builder
	depth       int
	inString    bool
	escape      bool
	currentKey  string
	keyBuf      strings.Builder
	scanningKey bool
	expectValue bool

	objStart int
	active   bool

	strValueStart int
	scanningStr   bool
}

func NewArrayItemScanner(onItem func(key, rawValue string)) *ArrayItemScanner {
	return &ArrayItemScanner{onItem: onItem}
}

func (s *ArrayItemScanner) Write(p []byte) (int, error) {
	for _, b := range p {
		s.buf.WriteByte(b)
		s.advance(b)
	}
	return len(p), nil
}

func (s *ArrayItemScanner) Raw() string {
	return s.buf.String()
}

func (s *ArrayItemScanner) advance(b byte) {
	if s.escape {
		s.escape = false
		if s.scanningKey {
			s.keyBuf.WriteByte(b)
		}
		return
	}
	if b == '\\' && s.inString {
		s.escape = true
		return
	}

	if b == '"' {
		s.inString = !s.inString
		if s.depth == 1 {
			if s.inString {
				if s.scanningKey || !s.expectValue {
					s.scanningKey = true
					s.expectValue = false
					s.keyBuf.Reset()
				} else {
					s.strValueStart = s.buf.Len() - 1
					s.scanningStr = true
					s.expectValue = false
				}
			} else if s.scanningKey {
				s.currentKey = s.keyBuf.String()
				s.scanningKey = false
			} else if s.scanningStr {
				raw := sanitise(s.buf.String()[s.strValueStart:])
				s.onItem(s.currentKey, raw)
				s.scanningStr = false
			}
		}
		return
	}
	if s.inString {
		if s.scanningKey {
			s.keyBuf.WriteByte(b)
		}
		return
	}

	switch b {
	case ':':
		if s.depth == 1 {
			s.scanningKey = false
			s.expectValue = true
		}
	case '{':
		s.depth++
		if s.depth == 2 {
			s.objStart = s.buf.Len() - 1
			s.active = true
			s.expectValue = false
		}
	case '}':
		if s.active && s.depth == 2 {
			raw := sanitise(s.buf.String()[s.objStart:])
			s.onItem(s.currentKey, raw)
			s.active = false
		}
		s.depth--
	case '[':
		if s.depth == 1 {
			s.scanningKey = false
			s.expectValue = false
		}
	case ',':
		if s.depth == 1 {
			s.scanningKey = true
			s.expectValue = false
		}
	}
}

func sanitise(s string) string {
	var b strings.Builder
	b.Grow(len(s))
	i := 0
	for i < len(s) {
		if s[i] != '\\' || i+1 >= len(s) {
			b.WriteByte(s[i])
			i++
			continue
		}
		next := s[i+1]
		switch next {
		case '"', '\\', '/', 'b', 'f', 'n', 'r', 't':
			b.WriteByte(s[i])
			b.WriteByte(next)
			i += 2
		case 'u':
			if i+5 < len(s) && isHex(s[i+2]) && isHex(s[i+3]) && isHex(s[i+4]) && isHex(s[i+5]) {
				b.Write([]byte(s[i : i+6]))
				i += 6
			} else {
				b.WriteString(`\\`)
				i++
			}
		default:
			b.WriteString(`\\`)
			i++
		}
	}
	return b.String()
}

func isHex(c byte) bool {
	return (c >= '0' && c <= '9') || (c >= 'a' && c <= 'f') || (c >= 'A' && c <= 'F')
}
