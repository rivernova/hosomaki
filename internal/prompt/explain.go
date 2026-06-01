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

func Explain(input, cmd string, env collector.Environment) string {
	var cmdContext string
	if c := strings.TrimSpace(cmd); c != "" {
		cmdContext = fmt.Sprintf("\nThe output below was produced by running: %s\n", c)
	}

	return fmt.Sprintf(`You are a Linux system expert. You will be given log output or an error message.

%sRULES — follow every one without exception:
- Plain prose only. No markdown. No bullet points. No numbered lists. No headers. No bold. No italics.
- Do not suggest fixes, solutions, commands to run, or next steps of any kind.
- Do not open with a preamble. Do not close with a summary or offer further help.
- Write between two and four sentences. Never exceed five sentences under any circumstances.
- If multiple distinct errors are present, address each one within the same paragraph.
- State what is happening and why. Focus on root cause and system behaviour.
- If a command is provided, use it to inform your understanding of the context.
- Your explanation must be correct for the host environment described above (distribution, kernel, init system, security model). Do not guess based on a different distro.
%sInput:
%s

REQUIRED — you MUST include this block at the very end of your response, after all prose, no exceptions:
---JSON---
{"patterns": <integer: count of distinct error patterns or issues you identified>, "causes": <integer: count of distinct root causes you identified>}
---END---
Example of a valid block: ---JSON---
{"patterns": 2, "causes": 1}
---END---
Do not skip this block. Do not add any text after ---END---.`, EnvironmentSection(env), cmdContext, input)
}
