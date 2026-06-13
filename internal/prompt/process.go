// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package prompt

import (
	"fmt"

	"github.com/rivernova/hosomaki/internal/collector"
)

// prompt logic for the --pid flag for the explain command

func ExplainProcess(sanitisedSnapshot string, pid int, env collector.Environment) string {
	return fmt.Sprintf(`You are a Linux system expert. You will be given a live procfs snapshot of a running process.

%s
INPUT FORMAT
The snapshot has been sanitised. You will see:
  - Placeholders: <PATH>, <CONFIG_PATH>, <LOG_PATH>, <CACHE_PATH>, <LIB_PATH>,
    <HOME_PATH>, <HEX>, <UUID>, <IPV4>, <IPV6>, <MAC>, <URL>.
Treat placeholders as opaque identifiers. Do not invent real values.

The snapshot contains:
  - Process metadata from /proc/<pid>/status: name, state, memory usage, thread count, UID.
  - Open file descriptors from /proc/<pid>/fd: resolved symlink targets, deduplicated.
  - TCP socket state from /proc/<pid>/net/tcp and /proc/<pid>/net/tcp6.
  - "Collection warnings" lines (if present) indicate that some data was
    inaccessible due to permissions. Account for this uncertainty in your answer.

TASK
Explain what the process (PID %d) is doing right now, in plain language accessible
to a system administrator. Use the open files, socket state, and process metadata
to infer what role this process is playing on the system.

Identify every distinct aspect of the process's current activity that can be
observed from the snapshot. Group tightly related observations into a single entry.

If the snapshot contains no meaningful information (empty file list, no sockets,
unknown state), return EXACTLY {"issues": []} and nothing else.

OUTPUT
Return ONLY a JSON object — no prose, no markdown fences, no text outside the JSON.

SCHEMA
%s

FIELD RULES
- "what": 2–4 sentences. Describe precisely what this process aspect is doing
  right now. Reference specific open files, socket states, or status fields
  visible in the snapshot. Explain the observable behaviour and what it means.
- "why": 2–4 sentences. Explain what this behaviour indicates about the role
  or current activity of the process. Reference the process name, UID, or
  file access patterns where they help clarify intent.
- Both values must be plain strings. Do not use arrays or nested objects.
- Do not suggest actions, fixes, or remediation steps.
- Group closely related observations into a single entry.

Process snapshot:
%s`, EnvironmentSection(env), pid, SchemaExplain, sanitisedSnapshot)
}
