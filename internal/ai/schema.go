// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package ai

import "fmt"

// required structure for JSON outputs

type Schema struct {
	raw string
}

func NewSchema(s string) Schema {
	if s == "" {
		panic("ai.NewSchema: schema string must not be empty")
	}
	return Schema{raw: s}
}

func (s Schema) String() string { return s.raw }

func (s Schema) Format(f fmt.State, verb rune) {
	_, err := fmt.Fprint(f, s.raw)
	if err != nil {
		return
	}
}
