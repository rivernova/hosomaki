// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package prompt

import (
	"fmt"

	"github.com/rivernova/hosomaki/internal/collector"
)

// prompt logic for the mounts command

type MountFinding struct {
	Severity   string `json:"severity"`
	MountPoint string `json:"mount_point"` // verbatim from input
	Title      string `json:"title"`
	Detail     string `json:"detail"`
}

type MountsResult struct {
	Summary  string         `json:"summary"`
	Findings []MountFinding `json:"findings"`
}

type MountsInput struct {
	Environment collector.Environment
	Mounts      string
}

func Mounts(in MountsInput) string {
	return fmt.Sprintf(`You are a Linux storage and filesystem expert reviewing the active mounts on a live system.

%s
TASK
Analyse the mount list below. Identify mount points that are unhealthy, approaching
capacity limits, stale (for NFS), misconfigured, or otherwise worth the operator's
attention.

A mount point is worth flagging when any of the following apply:
- NFS mount is marked STALE — this is a critical finding requiring immediate attention.
- Disk usage is at or above 85%% — warning; at or above 95%% — critical.
- A filesystem is mounted read-only when it is expected to be writable (e.g. /usr, /var, /home).
- A mount has the "errors=remount-ro" option active (indicates a past filesystem error).
- A root filesystem (/) or /var mount has unusually low available space.
- A mount uses an unexpected or legacy filesystem type for its role (e.g. FAT32 on /var).

Do NOT flag:
- Pseudo-filesystems (proc, sysfs, tmpfs, devtmpfs, cgroup, etc.) — these are normal.
- NFS mounts that are marked "responsive".
- Disk usage below 85%%.
- Read-only mounts that are expected to be read-only (e.g. /boot/efi with vfat, /sys).
- Empty findings — if everything looks healthy, return an empty findings array.

OUTPUT
Return ONLY a JSON object — no prose, no markdown fences, no text outside the JSON.
The JSON must use exactly these field names. Do not rename, abbreviate, or add fields.

SCHEMA
%s

FIELD RULES
- "summary": one to two sentences. State how many real mount points are present,
  overall disk health, and whether anything requires attention. Maximum 40 words.
- "findings": one entry per distinct concern. Do not invent mount points not present
  in the input.
- "severity": exactly one of "critical" (data loss risk or immediate action needed),
  "warning" (should be investigated soon), or "info" (worth noting).
- "mount_point": copy the mountpoint verbatim from the input.
- "title": a concise plain-text label. No trailing punctuation.
- "detail": 2–4 sentences. Describe precisely what is wrong, why it matters, what
  the likely cause is, and what the operator should do. Reference specific values
  (percentages, NFS server names) from the input. Do not suggest commands that
  modify the system.

OUTPUT FORMAT
No markdown. No bullet points. No numbered lists. No headers.
All string values are plain prose.

Active mounts:
%s`, EnvironmentSection(in.Environment), SchemaMounts, in.Mounts)
}
