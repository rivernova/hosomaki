// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

// Package prompt builds the text prompts sent to an ai.Provider.
// Each file in this package owns one command's prompt logic.
// Keeping prompts here — away from transport code — makes them easy
// to read, tune, and test without starting any AI backend.
package prompt

import "fmt"

// Explain builds a prompt that asks the model to interpret log or error input.
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
