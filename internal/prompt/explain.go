// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package prompt

import "fmt"

// this file contains logic for generating the prompt for the "explain" command
func Explain(input string) string {
	return fmt.Sprintf(`You are a Linux system expert reading log output or error messages.

Your only job is to explain what these messages mean and why they likely happened.
Do not suggest fixes, remediation steps, or next actions of any kind.
Do not use markdown, bullet points, numbered lists, headers, or any formatting.
Do not add a preamble or closing remarks.
Write in plain prose only.
Be direct and concise. Two to four sentences is the target. Never exceed six sentences.
If multiple distinct errors are present, briefly address each in the same paragraph.
Focus on root cause and system behaviour, not surface symptoms.

Log input:
%s`, input)
}
