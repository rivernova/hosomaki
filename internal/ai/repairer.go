// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package ai

import (
	"fmt"
	"strings"
)

// templates for repair prompts

type Repairer struct{}

func (Repairer) BuildStructuralRepairPrompt(schema Schema, invalidJSON string, errors []string) string {
	return fmt.Sprintf(
		"The JSON below is structurally invalid. Correct it to match the required schema.\n"+
			"\n"+
			"Rules:\n"+
			"  - Return ONLY the corrected JSON object.\n"+
			"  - No prose. No markdown fences. No extra keys.\n"+
			"  - Preserve every value that is already correct.\n"+
			"  - Fix only the fields listed under ERRORS.\n"+
			"\n"+
			"REQUIRED SCHEMA:\n%s\n"+
			"\n"+
			"INVALID JSON:\n%s\n"+
			"\n"+
			"ERRORS:\n  - %s",
		schema.String(),
		invalidJSON,
		strings.Join(errors, "\n  - "),
	)
}

func (Repairer) BuildSemanticRepairPrompt(schema Schema, originalPrompt string, errors []string) string {
	return fmt.Sprintf(
		"Your previous response was structurally valid JSON but failed these requirements:\n"+
			"  - %s\n"+
			"\n"+
			"You must retry the task below and produce a response that satisfies all requirements.\n"+
			"Return ONLY a JSON object matching this schema — no prose, no markdown fences:\n"+
			"%s\n"+
			"\n"+
			"TASK:\n%s",
		strings.Join(errors, "\n  - "),
		schema.String(),
		originalPrompt,
	)
}
