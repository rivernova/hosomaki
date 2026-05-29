// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package render

import "fmt"

// this file defines the pastel truecolor palette

type RGB struct{ R, G, B uint8 }

type Palette struct {
	Heading RGB
	Rule    RGB
	Label   RGB
	Value   RGB
	Text    RGB
	OK      RGB
	Warn    RGB
	Crit    RGB
	Accent  RGB
	Dim     RGB
}

func DefaultPalette() Palette {
	return Palette{
		Heading: RGB{198, 182, 255},
		Rule:    RGB{74, 78, 94},
		Label:   RGB{138, 146, 168},
		Value:   RGB{222, 224, 232},
		Text:    RGB{198, 202, 214},
		OK:      RGB{167, 227, 193},
		Warn:    RGB{244, 211, 148},
		Crit:    RGB{240, 160, 165},
		Accent:  RGB{150, 210, 224},
		Dim:     RGB{110, 116, 134},
	}
}

const (
	ansiReset = "\x1b[0m"
)

func fgSeq(c RGB) string {
	return fmt.Sprintf("\x1b[38;2;%d;%d;%dm", c.R, c.G, c.B)
}
