// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package prompt

// just schemas for the JSONs the command need

const SchemaExplain = `{"issues":[{"what":"string","why":"string"}]}`

const SchemaDoctorFull = `{"issues":[{"severity":"string","title":"string","detail":"string"}],"actions":[{"description":"string","disruptive":bool}]}`

const SchemaDoctorBrief = `{"summary":"string"}`

const SchemaStatusFull = `{"overview":"string","anomalies":[{"severity":"string","title":"string","detail":"string"}]}`

const SchemaStatusBrief = `{"summary":"string"}`

const SchemaAudit = `{"summary":"string","findings":[{"severity":"string","category":"string","title":"string","detail":"string"}]}`

const SchemaWatch = `{"issues":[{"what":"string","why":"string"}]}`

const SchemaWhy = `{"summary":"string","chain":[{"event":"string","detail":"string"}],"next_steps":["string"]}`

const SchemaPorts = `{"summary":"string","findings":[{"severity":"string","port":"string","title":"string","detail":"string"}]}`

const SchemaTimers = `{"summary":"string","timers":[{"name":"string","schedule":"string","last_run":"string","next_run":"string","status":"string","detail":"string"}]}`

const SchemaCrons = `{"summary":"string","jobs":[{"source":"string","schedule":"string","command":"string","what_it_does":"string","last_run":"string","status":"string","detail":"string"}]}`

const SchemaMounts = `{"summary":"string","findings":[{"severity":"string","mount_point":"string","title":"string","detail":"string"}]}`

const SchemaUpdates = `{"summary":"string","updates":[{"package":"string","installed":"string","available":"string","category":"string","reboot_required":false,"detail":"string"}]}`

const SchemaHistory = `{"summary":"string","entries":[{"timestamp":"string","command":"string","summary":"string"}]}`
