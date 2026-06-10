// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package prompt

import (
	"fmt"

	"github.com/rivernova/hosomaki/internal/collector"
)

// prompt logic for the audit command

type WatchIssue struct {
	What string `json:"what"`
	Why  string `json:"why"`
}

type WatchResult struct {
	Issues []WatchIssue `json:"issues"`
}

type WatchInput struct {
	Service     string
	Batch       string
	Environment collector.Environment
}

func Watch(in WatchInput) string {
	return fmt.Sprintf(`You are a Linux system expert monitoring a live journal stream.
 
%s
SERVICE
%s
 
INPUT FORMAT
The batch below is a fragment of the live journal for the service above.
It has been sanitised. You will see:
  - Line categories: <ERROR>, <WARN>, <INFO>, <DEBUG>, <TRANSACTION>, <SCRIPTLET>.
  - Placeholders: <URL>, <PATH>, <CONFIG_PATH>, <LOG_PATH>, <CACHE_PATH>,
    <LIB_PATH>, <HOME_PATH>, <HEX>, <UUID>, <IPV4>, <IPV6>, <MAC>,
    <EMAIL>, <REPO_CACHE>, <VERSION>.
Treat placeholders as opaque identifiers. Do not invent real values.
 
TASK
Examine the batch. Identify only lines tagged <ERROR> or <WARN>, plus any
<INFO> lines that are clearly part of the same incident. Group related lines
into distinct issues.
 
If the batch contains NO <ERROR> and NO <WARN> lines, return EXACTLY
{"issues": []} and nothing else. Do not invent issues.
 
Be concise. This output will be displayed inline in a live terminal session.
One entry per distinct issue. Do not repeat the same issue multiple times.
 
OUTPUT
Return ONLY a JSON object — no prose, no markdown fences, no text outside the JSON.
 
SCHEMA
%s
 
FIELD RULES
- "what": 2–4 sentences. Describe precisely what is happening.
  Reference specific line categories or placeholders. State the observable
  behaviour and its immediate effect on the service.
- "why": 2–4 sentences. Explain the root cause. If the cause cannot be
  determined from this batch alone, state what is most likely and why.
- Both fields are plain strings. No nested objects or arrays.
- Do not suggest fixes, commands, or remediation steps.
- Group related log lines into a single entry. Do not create one entry per line.
 
Batch:
%s`, EnvironmentSection(in.Environment), in.Service, SchemaExplain, in.Batch)
}
