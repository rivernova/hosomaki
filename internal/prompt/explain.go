// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package prompt

import (
	"fmt"
	"strings"

	"github.com/rivernova/hosomaki/internal/collector"
)

// this file contains the prompt template for the "explain" command

const explainBase = `You are the explain engine inside hosomaki, a Linux CLI tool.
Read the input below and explain in plain language what it means and what, if
anything, the user should do about it.

Answer in RAW prose only. Do NOT use markdown, headings, bullet points, code
fences, colours, icons, indentation or any layout. Write a few clear sentences.
hosomaki handles all formatting.
%s
%s
%s=== INPUT ===
%s`

func Explain(input, command string, env collector.Environment, language string) string {
	lang := ""
	if l := languageLine(language); l != "" {
		lang = "\n" + l
	}

	envBlock := EnvironmentSection(env)

	cmdCtx := ""
	if c := strings.TrimSpace(command); c != "" {
		cmdCtx = fmt.Sprintf("The input was produced by running: %s\n\n", c)
	}

	body := strings.TrimSpace(input)
	if body == "" {
		body = "(no data)"
	}

	return fmt.Sprintf(explainBase, lang, envBlock, cmdCtx, body)
}
