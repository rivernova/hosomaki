// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package ai

import (
	"fmt"
	"strings"
)

// repair prompts for structural and semantic failures

type Repairer struct{}

func (Repairer) BuildElementRepairPrompt(itemSchema Schema, invalidItem string, errors []string) string {
	return fmt.Sprintf(
		"One item in a list did not match its required schema. Correct that single item.\n"+
			"\n"+
			"Rules:\n"+
			"  - Return ONLY the corrected item as one JSON object.\n"+
			"  - No prose. No markdown fences. No surrounding array. No extra keys.\n"+
			"  - Preserve every value that is already correct.\n"+
			"  - Change only what the ERRORS require.\n"+
			"\n"+
			"REQUIRED ITEM SCHEMA:\n%s\n"+
			"\n"+
			"ITEM:\n%s\n"+
			"\n"+
			"ERRORS:\n  - %s",
		itemSchema.String(),
		invalidItem,
		strings.Join(errors, "\n  - "),
	)
}

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

func (Repairer) BuildStructuralRepairPromptWithContext(schema Schema, invalidJSON string, errors []string, originalPrompt string) string {
	return fmt.Sprintf(
		"Your previous response did not match the required schema and contains no usable content.\n"+
			"Repeat the original task below and return ONLY a JSON object matching the schema.\n"+
			"No prose, no markdown fences, no extra keys.\n"+
			"\n"+
			"REQUIRED SCHEMA:\n%s\n"+
			"\n"+
			"PREVIOUS INVALID RESPONSE:\n%s\n"+
			"\n"+
			"ERRORS:\n  - %s\n"+
			"\n"+
			"ORIGINAL TASK:\n%s",
		schema.String(),
		invalidJSON,
		strings.Join(errors, "\n  - "),
		originalPrompt,
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
