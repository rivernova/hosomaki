// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package prompt

import "fmt"

// this file contains logic for constructing the prompt sent to the AI provider for explanations

func Explain(input string) string {
	return fmt.Sprintf(`You are a Linux system expert. A user has piped log output or an error message to you.

Explain clearly and concisely:
1. What it means.
2. Why it likely happened.
3. What the user should do about it (if anything).

Rules: plain text only, no markdown, no bullet points, max 5 sentences, be direct.

Input:
%s`, input)
}
