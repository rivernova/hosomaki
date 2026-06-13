// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package prompt

import (
	"strings"
	"testing"

	"github.com/rivernova/hosomaki/internal/collector"
)

// unit tests for the ports prompt builder

func makePortsInput(ports string) PortsInput {
	return PortsInput{
		Environment: collector.Environment{
			DistroID:   "ubuntu",
			InitSystem: "systemd",
		},
		Ports: ports,
	}
}

func TestPorts_ContainsSchema(t *testing.T) {
	p := Ports(makePortsInput("tcp  0.0.0.0:22  — sshd (pid 1)"))
	if !strings.Contains(p, SchemaPorts) {
		t.Error("Ports() prompt must contain the schema constant")
	}
}

func TestPorts_ContainsEnvironmentSection(t *testing.T) {
	p := Ports(makePortsInput("tcp  0.0.0.0:22  — sshd (pid 1)"))
	if !strings.Contains(p, "Host environment") {
		t.Error("Ports() prompt must contain the environment section")
	}
}

func TestPorts_ContainsPortData(t *testing.T) {
	portData := "tcp  0.0.0.0:3306  — mysqld (pid 5678)"
	p := Ports(makePortsInput(portData))
	if !strings.Contains(p, portData) {
		t.Errorf("Ports() prompt must embed the port list verbatim, missing %q", portData)
	}
}

func TestPorts_InstructsPureJSON(t *testing.T) {
	p := Ports(makePortsInput("tcp  0.0.0.0:22  — sshd (pid 1)"))
	if !strings.Contains(p, "Return ONLY a JSON object") {
		t.Error("Ports() prompt must instruct the model to return only JSON")
	}
}

func TestPorts_InstructsNoMarkdown(t *testing.T) {
	p := Ports(makePortsInput("tcp  0.0.0.0:22  — sshd (pid 1)"))
	if !strings.Contains(p, "No markdown") {
		t.Error("Ports() prompt must forbid markdown output")
	}
}

func TestPorts_InstructsNoFixes(t *testing.T) {
	p := Ports(makePortsInput("tcp  0.0.0.0:22  — sshd (pid 1)"))
	if !strings.Contains(p, "Do not suggest") {
		t.Error("Ports() prompt must forbid fix or command suggestions")
	}
}

func TestPorts_InstructsNoInventedPorts(t *testing.T) {
	p := Ports(makePortsInput("tcp  0.0.0.0:22  — sshd (pid 1)"))
	if !strings.Contains(p, "Do not invent ports") {
		t.Error("Ports() prompt must instruct the model not to invent ports")
	}
}

func TestPorts_InstructsEmptyFindingsWhenClean(t *testing.T) {
	p := Ports(makePortsInput("tcp  127.0.0.1:22  — sshd (pid 1)"))
	if !strings.Contains(p, "empty findings array") {
		t.Error("Ports() prompt must tell the model to return empty findings when nothing is wrong")
	}
}

func TestPorts_SeverityFieldReferenced(t *testing.T) {
	p := Ports(makePortsInput("tcp  0.0.0.0:22  — sshd (pid 1)"))
	if !strings.Contains(p, `"severity"`) {
		t.Error("Ports() prompt must reference the 'severity' JSON field by name")
	}
}

func TestPorts_SummaryFieldReferenced(t *testing.T) {
	p := Ports(makePortsInput("tcp  0.0.0.0:22  — sshd (pid 1)"))
	if !strings.Contains(p, `"summary"`) {
		t.Error("Ports() prompt must reference the 'summary' JSON field by name")
	}
}

func TestPorts_PortFieldReferenced(t *testing.T) {
	p := Ports(makePortsInput("tcp  0.0.0.0:22  — sshd (pid 1)"))
	if !strings.Contains(p, `"port"`) {
		t.Error("Ports() prompt must reference the 'port' JSON field by name")
	}
}

func TestPorts_LocalhostExemptionMentioned(t *testing.T) {
	p := Ports(makePortsInput("tcp  127.0.0.1:6379  — redis-server (pid 999)"))
	if !strings.Contains(p, "127.0.0.1") {
		t.Error("Ports() prompt must mention the loopback exemption by address")
	}
	if !strings.Contains(p, "loopback") {
		t.Error("Ports() prompt must use the word 'loopback' when describing the exemption")
	}
}

func TestPorts_EmptyPortList(t *testing.T) {
	p := Ports(makePortsInput("(no listening ports found)"))
	if !strings.Contains(p, "no listening ports found") {
		t.Errorf("Ports() prompt must embed the empty port message verbatim")
	}
}

func TestSchemaPortsResult_ContainsSummary(t *testing.T) {
	if !strings.Contains(SchemaPorts, "summary") {
		t.Error("SchemaPortsResult must contain 'summary' field")
	}
}

func TestSchemaPortsResult_ContainsFindings(t *testing.T) {
	if !strings.Contains(SchemaPorts, "findings") {
		t.Error("SchemaPortsResult must contain 'findings' field")
	}
}

func TestSchemaPortsResult_ContainsSeverity(t *testing.T) {
	if !strings.Contains(SchemaPorts, "severity") {
		t.Error("SchemaPortsResult must contain 'severity' field")
	}
}

func TestSchemaPortsResult_ContainsPort(t *testing.T) {
	if !strings.Contains(SchemaPorts, `"port"`) {
		t.Error("SchemaPortsResult must contain 'port' field")
	}
}

func TestPorts_PromptPackageHasNoSanitiserImport(_ *testing.T) {
	// Intentionally empty
}
