// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package prompt

import (
	"fmt"

	"github.com/rivernova/hosomaki/internal/collector"
)

// prompt logic for the why command

type WhyStep struct {
	Event  string `json:"event"`
	Detail string `json:"detail"`
}

type WhyResult struct {
	Summary   string    `json:"summary"`
	Chain     []WhyStep `json:"chain"`
	NextSteps []string  `json:"next_steps"`
}

type WhyInput struct {
	Service     string
	ExitCode    int
	Logs        string
	Environment collector.Environment
}

func Why(in WhyInput) string {
	return fmt.Sprintf(`You are a Linux system expert performing a failure-chain analysis.

%s
CONTEXT
Service  : %s
Exit code: %d (%s)

INPUT FORMAT
The journal excerpt below covers the period leading up to and including the failure.
It has been sanitised. You will see:
  - Line categories: <ERROR>, <WARN>, <INFO>, <DEBUG>, <TRANSACTION>, <SCRIPTLET>.
  - Placeholders: <URL>, <PATH>, <CONFIG_PATH>, <LOG_PATH>, <CACHE_PATH>,
    <LIB_PATH>, <HOME_PATH>, <HEX>, <UUID>, <IPV4>, <IPV6>, <MAC>,
    <EMAIL>, <REPO_CACHE>, <VERSION>.
Treat placeholders as opaque identifiers. Do not invent real values for them.

TASK
Reconstruct the failure chain that caused the service to exit with code %d.

A failure chain is an ordered sequence of events where each event either caused
or enabled the next. The first entry must be the root cause (or the earliest
observable precursor visible in the logs). The last entry must be the proximate
cause that produced the nonzero exit.

If the logs are insufficient to establish a full chain, state what can be
determined and what remains unknown.

Exit-code semantics to consider:
  1   — generic error (examine the last <ERROR> line)
  2   — misuse of shell built-in or bad argument
  126 — permission denied or command not executable
  127 — command not found
  128 — invalid exit argument
  130 — terminated by SIGINT (Ctrl-C)
  137 — killed by SIGKILL (OOM killer or explicit kill)
  139 — segmentation fault (SIGSEGV)
  143 — terminated by SIGTERM
  Other values above 128 indicate termination by signal (value − 128 = signal number).

OUTPUT
Return ONLY a JSON object — no prose, no markdown fences, no text outside the JSON.
The JSON must use exactly these field names. Do not rename, abbreviate, or add fields.

SCHEMA
%s

FIELD RULES
- "summary": exactly ONE sentence. Answer "why did %s exit with code %d?" in plain
  language. Name the root cause. Maximum 35 words.
- "chain": two to six entries ordered root cause → proximate cause.
  Do not exceed six entries. Do not create one entry per log line.
- "event": a terse plain-text label, 3–8 words. No trailing punctuation.
- "detail": 2–4 sentences. Explain what happened at this step, cite the log
  evidence (reference specific line categories or sanitised placeholders), and
  describe how this step led to the next one in the chain.
- "next_steps": two to five plain-prose sentences, each describing one concrete
  remediation step. Name the exact command, file, or configuration value to
  inspect or change. Write in imperative mood.
- All values are plain strings. No markdown, no nested objects, no arrays inside strings.

No markdown. No bullet points. No numbered lists. No headers outside the JSON.

Journal excerpt:
%s`,
		EnvironmentSection(in.Environment),
		in.Service,
		in.ExitCode,
		exitCodeLabel(in.ExitCode),
		in.ExitCode,
		SchemaWhy,
		in.Service,
		in.ExitCode,
		in.Logs,
	)
}

func exitCodeLabel(code int) string {
	switch code {
	case 1:
		return "generic error"
	case 2:
		return "misuse of shell built-in or bad argument"
	case 126:
		return "permission denied or command not executable"
	case 127:
		return "command not found"
	case 128:
		return "invalid exit argument"
	case 130:
		return "terminated by SIGINT (Ctrl-C)"
	case 137:
		return "killed by SIGKILL — likely OOM killer"
	case 139:
		return "segmentation fault (SIGSEGV)"
	case 143:
		return "terminated by SIGTERM"
	}
	if code > 128 {
		return fmt.Sprintf("terminated by signal %d", code-128)
	}
	return "nonzero exit"
}
