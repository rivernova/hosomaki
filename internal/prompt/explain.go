// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at [https://mozilla.org/MPL/2.0/](https://mozilla.org/MPL/2.0/).

package prompt

import (
	"fmt"
	"strings"

	"github.com/rivernova/hosomaki/internal/collector"
)

// this file contains the prompt template and builder for the "explain" command

const explainBase = `<system>You are a Linux sysadmin expert. Analyze the provided logs and output exactly one <issue> block per distinct component error. 

CRITICAL FORMATTING CONTRAINTS (CLI PARSER SAFE):
- Your entire response must conform strictly to the XML schema below.
- DO NOT wrap the output in markdown code blocks.
- DO NOT use markdown characters inside the XML tags (no asterisks, no backticks, no text tables, no custom markdown formatting).
- DO NOT output any mitigation commands, fixes, or recommendations. Focus strictly on understanding the logs.
- DO NOT copy-paste raw logs into the symptom or cause tags. Translate the log context into clear, well-elaborated plain narrative.

<issues>
  <issue>
    <component>[The system service, package, identifier, or kernel element affected]</component>
    <symptom>[A highly detailed, well-elaborated plain narrative describing what is misbehaving and the immediate system impact of this failure.]</symptom>
    <cause>[A comprehensive plain narrative explaining the background mechanics of why this error occurred based on the log pattern.]</cause>
  </issue>
</issues></system>
%s
%s
%s=== LOG INPUT ===
%s`

func Explain(input, command string, env collector.Environment, language string) string {
	lang := ""
	if l := languageLine(language); l != "" {
		lang = "\n" + l
	}

	envBlock := EnvironmentSection(env)

	cmdCtx := ""
	if c := strings.TrimSpace(command); c != "" {
		cmdCtx = fmt.Sprintf("These logs were produced by: %s\n\n", c)
	}

	body := strings.TrimSpace(input)
	if body == "" {
		body = "(no data)"
	}

	return fmt.Sprintf(explainBase, lang, envBlock, cmdCtx, body)
}
