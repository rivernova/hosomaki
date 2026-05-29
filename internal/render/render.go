// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package render

import (
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"
	"unicode/utf8"
)

// this file contains the renderer that turns analysis reports into formatted terminal output

type Status int

const (
	Neutral Status = iota
	OK
	Info
	Warn
	Crit
)

const (
	indentStep   = 2
	maxLabelCol  = 16
	minContent   = 50
	maxContent   = 96
	ruleWidth    = 48
	defaultWidth = 80
)

type Renderer struct {
	w     io.Writer
	pal   Palette
	color bool
	width int
}

type Option func(*Renderer)

func WithColor(on bool) Option { return func(r *Renderer) { r.color = on } }

func WithWidth(cols int) Option {
	return func(r *Renderer) {
		if cols > 0 {
			r.width = clamp(cols, minContent, maxContent)
		}
	}
}

func WithPalette(p Palette) Option { return func(r *Renderer) { r.pal = p } }

func New(w io.Writer, opts ...Option) *Renderer {
	r := &Renderer{
		w:     w,
		pal:   DefaultPalette(),
		color: detectColor(w),
		width: detectWidth(),
	}
	for _, o := range opts {
		o(r)
	}
	return r
}

func detectColor(w io.Writer) bool {
	if _, ok := os.LookupEnv("NO_COLOR"); ok {
		return false
	}
	f, ok := w.(*os.File)
	if !ok {
		return false
	}
	info, err := f.Stat()
	if err != nil {
		return false
	}
	return info.Mode()&os.ModeCharDevice != 0
}

func detectWidth() int {
	if c, err := strconv.Atoi(strings.TrimSpace(os.Getenv("COLUMNS"))); err == nil && c > 0 {
		return clamp(c, minContent, maxContent)
	}
	return defaultWidth
}

func (r *Renderer) paint(c RGB, s string) string {
	if !r.color {
		return s
	}
	return fgSeq(c) + s + ansiReset
}

func (r *Renderer) line(s string) { fmt.Fprintln(r.w, s) }

func (r *Renderer) Blank() { r.line("") }

func (r *Renderer) Title(s string) {
	r.Blank()
	r.line(indent(1) + r.paint(r.pal.Heading, s))
}

func (r *Renderer) Section(title string) {
	r.Blank()
	r.line(indent(1) + r.paint(r.pal.Heading, title))
	r.line(indent(1) + r.paint(r.pal.Rule, strings.Repeat("─", r.ruleLen())))
}

func (r *Renderer) ruleLen() int {
	if w := r.width - indentStep; w < ruleWidth {
		return max(w, 8)
	}
	return ruleWidth
}

func (r *Renderer) Metric(label, value string, st Status) {
	lbl := pad(label, maxLabelCol)
	row := indent(1) + r.paint(r.pal.Label, lbl) + "  " + r.paint(r.statusColor(st), value)
	if tag := statusTag(st); tag != "" {
		gap := r.tagGap(lbl, value)
		row += strings.Repeat(" ", gap) + r.paint(r.statusColor(st), tag)
	}
	r.line(row)
}

func (r *Renderer) tagGap(paddedLabel, value string) int {
	used := indentStep + dispWidth(paddedLabel) + 2 + dispWidth(value)
	target := r.width - 8
	if gap := target - used; gap > 1 {
		return gap
	}
	return 2
}

func (r *Renderer) Finding(st Status, text string) {
	icon := statusIcon(st)
	c := r.statusColor(st)
	prefix := indent(1) + r.paint(c, icon) + " "
	for i, ln := range r.wrapBody(text, indentStep+2) {
		if i == 0 {
			r.line(prefix + r.paint(r.pal.Text, ln))
		} else {
			r.line(indent(1) + "  " + r.paint(r.pal.Text, ln))
		}
	}
}

func (r *Renderer) Process(text string) {
	r.line(indent(1) + r.paint(r.pal.Dim, "> "+text))
}

func (r *Renderer) Paragraph(text string) {
	for _, block := range strings.Split(strings.TrimRight(text, "\n"), "\n") {
		block = strings.TrimSpace(block)
		if block == "" {
			r.Blank()
			continue
		}
		for _, ln := range r.wrapBody(block, indentStep) {
			r.line(indent(1) + r.paint(r.pal.Text, ln))
		}
	}
}

func (r *Renderer) Subject(name string, st Status) {
	r.line(indent(1) + r.paint(r.statusColor(st), name))
}

func (r *Renderer) Detail(key, value string) {
	var body string
	switch {
	case key != "" && value != "":
		body = key + ": " + value
	case key != "":
		body = key
	default:
		body = value
	}
	bullet := r.paint(r.pal.Dim, "•") + " "
	lines := r.wrapBody(body, indentStep*2+2)
	for i, ln := range lines {
		if i == 0 {
			r.line(indent(2) + bullet + r.paint(r.pal.Text, ln))
		} else {
			r.line(indent(2) + "  " + r.paint(r.pal.Text, ln))
		}
	}
}

func (r *Renderer) Command(cmd string, disruptive bool) {
	r.line(indent(3) + r.paint(r.pal.Accent, cmd))
	if disruptive {
		r.line(indent(3) + r.paint(r.pal.Warn, "potentially disruptive — review before running"))
	}
}

func (r *Renderer) SummaryLine(text string, st Status) {
	r.line(indent(1) + r.paint(r.statusColor(st), text))
}

func (r *Renderer) Done() {
	r.Blank()
	r.line(indent(1) + r.paint(r.pal.OK, "done ✓"))
	r.Blank()
}

func (r *Renderer) Error(err error) {
	r.Blank()
	first := true
	for _, ln := range strings.Split(strings.TrimRight(err.Error(), "\n"), "\n") {
		if first {
			r.line(indent(1) + r.paint(r.pal.Crit, "! ") + r.paint(r.pal.Text, ln))
			first = false
			continue
		}
		r.line(indent(1) + "  " + r.paint(r.pal.Text, ln))
	}
	r.Blank()
}

func (r *Renderer) StreamStart(section string) io.Writer {
	r.Section(section)
	r.Blank()
	fmt.Fprint(r.w, indent(1))
	return &streamWriter{r: r}
}

func (r *Renderer) StreamEnd() {
	fmt.Fprintln(r.w)
	r.Blank()
}

type streamWriter struct {
	r *Renderer
}

func (sw *streamWriter) Write(p []byte) (int, error) {
	s := string(p)
	for len(s) > 0 {
		nl := strings.IndexByte(s, '\n')
		if nl < 0 {
			sw.writeChunk(s)
			s = ""
		} else {
			sw.writeChunk(s[:nl])
			fmt.Fprintln(sw.r.w)
			fmt.Fprint(sw.r.w, indent(1))
			s = s[nl+1:]
		}
	}
	return len(p), nil
}

func (sw *streamWriter) writeChunk(chunk string) {
	if sw.r.color {
		fmt.Fprint(sw.r.w, fgSeq(sw.r.pal.Text)+chunk+ansiReset)
	} else {
		fmt.Fprint(sw.r.w, chunk)
	}
}

func (r *Renderer) statusColor(st Status) RGB {
	switch st {
	case OK:
		return r.pal.OK
	case Warn:
		return r.pal.Warn
	case Crit:
		return r.pal.Crit
	case Info:
		return r.pal.Accent
	default:
		return r.pal.Value
	}
}

func (r *Renderer) wrapBody(text string, leading int) []string {
	return wrap(text, r.width-leading)
}

func statusIcon(st Status) string {
	switch st {
	case OK:
		return "✓"
	case Warn, Crit:
		return "!"
	default:
		return ">"
	}
}

func statusTag(st Status) string {
	switch st {
	case OK:
		return "[ OK ]"
	case Warn:
		return "[ WARN ]"
	case Crit:
		return "[ CRIT ]"
	case Info:
		return "[ INFO ]"
	default:
		return ""
	}
}

func indent(steps int) string { return strings.Repeat(" ", steps*indentStep) }

func dispWidth(s string) int { return utf8.RuneCountInString(s) }

func pad(s string, n int) string {
	if d := dispWidth(s); d < n {
		return s + strings.Repeat(" ", n-d)
	}
	return s
}

func wrap(text string, width int) []string {
	text = strings.TrimSpace(text)
	if text == "" {
		return []string{""}
	}
	if width <= 0 {
		return []string{text}
	}

	var lines []string
	var cur strings.Builder
	curLen := 0

	flush := func() {
		lines = append(lines, cur.String())
		cur.Reset()
		curLen = 0
	}

	for _, word := range strings.Fields(text) {
		wl := dispWidth(word)
		for wl > width {
			if curLen > 0 {
				flush()
			}
			head, tail := splitAt(word, width)
			lines = append(lines, head)
			word, wl = tail, dispWidth(tail)
		}
		switch {
		case curLen == 0:
			cur.WriteString(word)
			curLen = wl
		case curLen+1+wl <= width:
			cur.WriteByte(' ')
			cur.WriteString(word)
			curLen += 1 + wl
		default:
			flush()
			cur.WriteString(word)
			curLen = wl
		}
	}
	if curLen > 0 || len(lines) == 0 {
		flush()
	}
	return lines
}

func splitAt(s string, n int) (head, tail string) {
	i := 0
	for idx := range s {
		if i == n {
			return s[:idx], s[idx:]
		}
		i++
	}
	return s, ""
}

func clamp(v, lo, hi int) int {
	if v < lo {
		return lo
	}
	if v > hi {
		return hi
	}
	return v
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}
