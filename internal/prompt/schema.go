// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package prompt

// schema definitions for the JSON output expected from the LLM in each prompt

const SchemaExplain = `{"issues":[{"what":"string","why":"string"}]}`

const SchemaDoctorFull = `{"issues":[{"severity":"string","title":"string","detail":"string"}],"actions":[{"description":"string","disruptive":false}]}`

const SchemaDoctorBrief = `{"summary":"string"}`

const SchemaStatusFull = `{"summary":"string","details":[{"title":"string","value":"string"}]}`

const SchemaStatusBrief = `{"summary":"string"}`
