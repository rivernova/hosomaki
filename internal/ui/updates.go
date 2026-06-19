// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package ui

import (
	"fmt"
	"strings"

	"github.com/rivernova/hosomaki/internal/collector"
	"github.com/rivernova/hosomaki/internal/prompt"
)

func UpdatesHeader() string {
	return Title("pending updates")
}

func UpdatesNoPending() string {
	return Section("results", "No pending package updates found.\n")
}

func UpdatesNoPendingMsg(msg string) string {
	return Section("results", msg+"\n")
}

func UpdatesPendingList(updates []collector.PendingUpdate, securityOnly bool) string {
	var b strings.Builder

	label := "pending updates"
	if securityOnly {
		label = "security-related updates"
	}

	b.WriteString(Section(label, fmt.Sprintf("%d package(s)\n", len(updates))))

	for _, u := range updates {
		flags := ""
		if u.Security {
			flags += " [" + styleWarn + "SECURITY" + styleReset + "]"
		}
		if u.RebootRequired {
			flags += " [" + styleWarn + "REBOOT" + styleReset + "]"
		}

		inst := u.Installed
		if inst == "" {
			inst = "?"
		}
		avail := u.Available
		if avail == "" {
			avail = "?"
		}

		_, _ = fmt.Fprintf(&b, "  %s%s%s  %s → %s\n",
			u.Package, flags, styleReset, inst, avail)
	}

	return b.String()
}

func UpdatesResultSection(result *prompt.UpdatesResult) string {
	var b strings.Builder

	b.WriteString("\n")

	// Summary
	summary := strings.TrimSpace(result.Summary)
	if summary == "" {
		summary = "(no summary)"
	}
	b.WriteString(SectionSummary(summary))

	// Individual updates
	if len(result.Updates) > 0 {
		b.WriteString(Section("updates", ""))
		for i, u := range result.Updates {
			severity := styleOK
			switch u.Category {
			case "security":
				severity = styleFail
			case "major":
				severity = styleWarn
			}

			reboot := ""
			if u.RebootRequired {
				reboot = " [reboot required]"
			}

			_, _ = fmt.Fprintf(&b, "%s  %s%-20s%s  %s → %s  [%s]%s\n",
				glyphForCategory(u.Category),
				styleReset,
				u.Package,
				severity,
				u.Installed, u.Available,
				u.Category,
				reboot,
			)

			// Detail - only if there's interesting info
			// (the AI response doesn't have per-update detail in the schema,
			//  so this is just the structured overview)
			_ = i
		}
	}

	return b.String()
}

func glyphForCategory(cat string) string {
	switch cat {
	case "security":
		return glyphFail
	case "major":
		return glyphWarn
	default:
		return glyphOK
	}
}
