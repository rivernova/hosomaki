// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package ui

import (
	"fmt"
	"strings"
)

// primitives. the raw building blocks

const (
	kvKeyWidth    = 16
	separatorLen  = 46
	separatorRune = '─'
	summaryIndent = "   "

	glyphOK   = "✓"
	glyphWarn = "!"
	glyphFail = "✗"

	// Pastel palette
	styleReset       = "\x1b[0m"
	styleTitle       = "\x1b[38;5;146m" // soft lavender
	styleSection     = "\x1b[38;5;182m" // light lilac
	styleMuted       = "\x1b[38;5;250m" // soft grey
	styleOK          = "\x1b[38;5;121m" // mint pastel
	styleWarn        = "\x1b[38;5;222m" // soft amber pastel
	styleFail        = "\x1b[38;5;210m" // coral pastel
	styleSeparator   = "\x1b[38;5;183m" // muted lilac
	styleSummaryLine = "\x1b[38;5;139m" // deeper lavender

)

func Title(text string) string {
	return fmt.Sprintf("%s%s%s\n", styleTitle, text, styleReset)
}

func Section(title, body string) string {
	var b strings.Builder
	fmt.Fprintf(&b, "\n%s%s%s\n\n", styleSection, title, styleReset)
	if strings.TrimSpace(body) != "" {
		b.WriteString(body)
		if !strings.HasSuffix(body, "\n") {
			b.WriteByte('\n')
		}
	}
	return b.String()
}

func SectionCompact(title, body string) string {
	var b strings.Builder
	fmt.Fprintf(&b, "\n%s%s%s\n\n", styleSection, title, styleReset)
	if strings.TrimSpace(body) != "" {
		b.WriteString(body)
		if !strings.HasSuffix(body, "\n") {
			b.WriteByte('\n')
		}
	}
	return b.String()
}

func SectionSummary(body string) string {
	sep := strings.Repeat(string(separatorRune), separatorLen)
	var b strings.Builder
	fmt.Fprintf(&b, "\n%ssummary%s\n%s%s%s\n", styleSection, styleReset, styleSeparator, sep, styleReset)
	if strings.TrimSpace(body) != "" {
		b.WriteString(body)
		if !strings.HasSuffix(body, "\n") {
			b.WriteByte('\n')
		}
	}
	b.WriteByte('\n')
	return b.String()
}

func SummaryLine(text string) string {
	return fmt.Sprintf("%s%s%s%s\n", summaryIndent, styleSummaryLine, text, styleReset)
}

func KeyValue(key, value string) string {
	padded := fmt.Sprintf("%-*s", kvKeyWidth, key)
	return fmt.Sprintf("%s%s%s %s\n", styleMuted, padded, styleReset, value)
}

func BulletOK(text string) string {
	return fmt.Sprintf("%s%s%s %s\n", styleOK, glyphOK, styleReset, text)
}

func BulletWarn(text string) string {
	return fmt.Sprintf("%s%s%s %s\n", styleWarn, glyphWarn, styleReset, text)
}

func BulletFail(text string) string {
	return fmt.Sprintf("%s%s%s %s\n", styleFail, glyphFail, styleReset, text)
}

func Separator() string {
	return ""
}

func sectionHeader(title string) string {
	return fmt.Sprintf("\n%s%s%s\n\n", styleSection, title, styleReset)
}

func compactHeader(title string) string {
	return fmt.Sprintf("\n%s%s%s\n\n", styleSection, title, styleReset)
}
