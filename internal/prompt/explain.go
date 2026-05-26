// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package prompt

import (
	"fmt"
	"strings"
)

// this file contains logic for constructing the prompt for the "explain" command

func Explain(input, cmd string) string {
	var cmdContext string
	if c := strings.TrimSpace(cmd); c != "" {
		cmdContext = fmt.Sprintf("\nThe output below was produced by running: %s\n", c)
	}

	return fmt.Sprintf(`You are a Linux system expert. You will be given log output or an error message.

RULES — follow every one without exception:
- Plain prose only. No markdown. No bullet points. No numbered lists. No headers. No bold. No italics.
- Do not suggest fixes, solutions, commands to run, or next steps of any kind.
- Do not open with a preamble. Do not close with a summary or offer further help.
- Write between two and four sentences. Never exceed five sentences under any circumstances.
- If multiple distinct errors are present, address each one within the same paragraph.
- State what is happening and why. Focus on root cause and system behaviour.
- If a command is provided, use it to inform your understanding of the context.
%sInput:
%s`, cmdContext, input)
}
