// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package prompt

import (
	"fmt"
	"strings"

	"github.com/rivernova/hosomaki/internal/collector"
)

// template for prompt for explain command

type ExplainResult struct {
	What string `json:"what"`
	Why  string `json:"why"`
}

func Explain(input, cmd string, env collector.Environment) string {
	var cmdContext string
	if c := strings.TrimSpace(cmd); c != "" {
		cmdContext = fmt.Sprintf("\nThe output below was produced by running: %s\n", c)
	}

	return fmt.Sprintf(`You are a Linux system expert. You will be given log output or an error message.

%s
TASK
Analyse the input below. Respond with ONLY a JSON object — no other text, no markdown, no explanation.

The JSON object must have exactly these two keys, both with string values:
{"what":"<prose string>","why":"<prose string>"}

"what": one continuous prose string describing every distinct error or event in the input.
"why": one continuous prose string explaining every root cause. Do not suggest fixes.

Both values must be plain strings, not arrays.
%s
Input:
%s`, EnvironmentSection(env), cmdContext, input)
}
