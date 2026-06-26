// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package collector

import (
	"context"
	"fmt"
	"os/exec"
	"strings"
	"time"
)

const firewallCmdTimeout = 5 * time.Second

type FirewallBackend string

const (
	BackendFirewalld FirewallBackend = "firewalld"
	BackendUfw       FirewallBackend = "ufw"
	BackendNftables  FirewallBackend = "nftables"
	BackendIptables  FirewallBackend = "iptables"
	BackendNone      FirewallBackend = ""
)

type FirewallReadStatus string

const (
	ReadNone    FirewallReadStatus = "none"
	ReadOK      FirewallReadStatus = "ok"
	ReadPartial FirewallReadStatus = "partial"
	ReadEmpty   FirewallReadStatus = "empty"
	ReadFailed  FirewallReadStatus = "failed"
)

type FirewallRule struct {
	Backend  FirewallBackend
	Chain    string
	Action   string
	Protocol string
	Port     string
	Source   string
	Comment  string
}

type FirewallResult struct {
	Backend    FirewallBackend
	Rules      []FirewallRule
	Zones      []string
	Warning    string
	ReadStatus FirewallReadStatus
}

func runFirewallCmd(name string, args ...string) ([]byte, error) {
	ctx, cancel := context.WithTimeout(context.Background(), firewallCmdTimeout)
	defer cancel()
	return exec.CommandContext(ctx, name, args...).Output()
}

func ufwState() (active, installed bool) {
	out, err := runFirewallCmd(binUfw, "status")
	if err != nil {
		return false, false
	}
	text := string(out)
	return strings.Contains(text, "Status: active"), strings.Contains(text, "Status:")
}

// DetectFirewallBackend probes active backends: firewalld → ufw → nftables → iptables.
func DetectFirewallBackend() FirewallBackend {
	out, err := runFirewallCmd(binFirewallCmd, "--state")
	if err == nil && strings.TrimSpace(string(out)) == "running" {
		return BackendFirewalld
	}

	_, installed := ufwState()
	if installed {
		return BackendUfw
	}

	out, err = runFirewallCmd(binNft, "list", "ruleset")
	if err == nil && strings.Contains(string(out), "chain ") {
		return BackendNftables
	}

	if _, err := runFirewallCmd(binIptables, "-L", "-n"); err == nil {
		return BackendIptables
	}

	return BackendNone
}

// FirewallRules reads rules from the detected backend.
func FirewallRules() FirewallResult {
	backend := DetectFirewallBackend()

	switch backend {
	case BackendFirewalld:
		return finalizeFirewallResult(collectFirewalld())
	case BackendUfw:
		if active, installed := ufwState(); installed && !active {
			return FirewallResult{
				Backend:    BackendUfw,
				ReadStatus: ReadFailed,
				Warning:    "ufw is installed but inactive — raw netfilter fallback skipped to avoid misreading managed rules",
			}
		}
		return finalizeFirewallResult(collectUfw())
	case BackendNftables:
		return finalizeFirewallResult(collectNftables())
	case BackendIptables:
		return finalizeFirewallResult(collectIptables())
	default:
		return FirewallResult{
			Backend:    BackendNone,
			ReadStatus: ReadNone,
			Warning:    "no firewall backend detected (tried firewalld, ufw, nftables, iptables)",
		}
	}
}

func finalizeFirewallResult(result FirewallResult) FirewallResult {
	if result.ReadStatus != "" {
		return result
	}
	switch {
	case result.Warning != "" && len(result.Rules) == 0:
		result.ReadStatus = ReadFailed
	case result.Warning != "":
		result.ReadStatus = ReadPartial
	case len(result.Rules) == 0:
		result.ReadStatus = ReadEmpty
	default:
		result.ReadStatus = ReadOK
	}
	return result
}

func collectFirewalld() FirewallResult {
	result := FirewallResult{Backend: BackendFirewalld}

	zoneOut, err := runFirewallCmd(binFirewallCmd, "--get-default-zone")
	zoneName := ""
	if err == nil {
		zoneName = strings.TrimSpace(string(zoneOut))
	} else {
		result.Warning = fmt.Sprintf("firewall-cmd --get-default-zone: %v", err)
		return result
	}

	// Read rules from default zone
	if zoneName != "" {
		out, err := runFirewallCmd(binFirewallCmd, "--zone="+zoneName, "--list-all")
		if err == nil {
			result.Rules = append(result.Rules, parseFirewalldOutput(zoneName, string(out))...)
		} else {
			result.Warning = fmt.Sprintf("firewall-cmd --zone=%s --list-all: %v", zoneName, err)
			return result
		}
	}

	result.Zones = append(result.Zones, zoneName)

	// Read rules from all active (interface-bound) zones
	activeOut, err := runFirewallCmd(binFirewallCmd, "--get-active-zones")
	if err == nil {
		for _, line := range strings.Split(string(activeOut), "\n") {
			line = strings.TrimSpace(line)
			if line == "" || strings.HasPrefix(line, "interfaces:") {
				continue
			}
			az := strings.Fields(line)[0]
			if az == "" || az == zoneName {
				continue
			}
			out, azErr := runFirewallCmd(binFirewallCmd, "--zone="+az, "--list-all")
			if azErr == nil {
				result.Rules = append(result.Rules, parseFirewalldOutput(az, string(out))...)
				result.Zones = append(result.Zones, az)
			} else if result.Warning == "" {
				result.Warning = fmt.Sprintf("firewall-cmd --zone=%s --list-all: %v", az, azErr)
			}
		}
	}

	return result
}

func parseFirewalldOutput(zone, output string) []FirewallRule {
	var rules []FirewallRule
	for _, line := range nonEmptyLines(output) {
		line = strings.TrimSpace(line)
		switch {
		case strings.HasPrefix(line, "ports:"):
			for _, p := range strings.Fields(strings.TrimPrefix(line, "ports:")) {
				proto, port := "tcp", p
				if parts := strings.SplitN(p, "/", 2); len(parts) == 2 {
					port, proto = parts[0], parts[1]
				}
				rules = append(rules, FirewallRule{
					Backend: BackendFirewalld, Chain: zone, Action: "ALLOW",
					Protocol: proto, Port: port, Source: "any",
				})
			}
		case strings.HasPrefix(line, "services:"):
			for _, s := range strings.Fields(strings.TrimPrefix(line, "services:")) {
				rules = append(rules, FirewallRule{
					Backend: BackendFirewalld, Chain: zone, Action: "ALLOW",
					Comment: "service: " + s,
				})
			}
		case strings.HasPrefix(line, "sources:"):
			for _, s := range strings.Fields(strings.TrimPrefix(line, "sources:")) {
				rules = append(rules, FirewallRule{
					Backend: BackendFirewalld, Chain: zone, Action: "ALLOW",
					Source: s, Comment: "trusted source",
				})
			}
		case strings.HasPrefix(line, "rich-rules:"):
			if rich := strings.TrimSpace(strings.TrimPrefix(line, "rich-rules:")); rich != "" {
				rules = append(rules, FirewallRule{
					Backend: BackendFirewalld, Chain: zone, Action: "RULE",
					Comment: "rich rule: " + rich,
				})
			}
		}
	}
	return rules
}

func collectUfw() FirewallResult {
	result := FirewallResult{Backend: BackendUfw}
	out, err := runFirewallCmd(binUfw, "status", "numbered")
	if err != nil {
		result.Warning = fmt.Sprintf("ufw status numbered: %v", err)
		return result
	}
	result.Rules = parseUfwOutput(string(out))
	return result
}

func parseUfwOutput(output string) []FirewallRule {
	var rules []FirewallRule
	for _, line := range nonEmptyLines(output) {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "Status:") || strings.HasPrefix(line, "To") || line == "" {
			continue
		}
		fields := strings.Fields(line)
		if len(fields) < 4 {
			continue
		}

		portIdx := -1
		var port, protocol, action, source, comment string
		for i, f := range fields {
			if strings.Contains(f, "/") {
				portIdx = i
				parts := strings.SplitN(f, "/", 2)
				port, protocol = parts[0], parts[1]
				break
			}
		}
		if portIdx < 0 || portIdx+1 >= len(fields) {
			continue
		}

		action = fields[portIdx+1]
		if portIdx+2 < len(fields) && (fields[portIdx+2] == "IN" || fields[portIdx+2] == "OUT") {
			action += " " + fields[portIdx+2]
		}
		for _, f := range fields {
			if f == "Anywhere" || strings.Contains(f, ".") || strings.Contains(f, ":") {
				source = f
			}
		}
		for i, f := range fields {
			if strings.HasPrefix(f, "(") && i > portIdx {
				comment = strings.Trim(strings.Join(fields[i:], " "), "()")
				break
			}
		}

		rules = append(rules, FirewallRule{
			Backend: BackendUfw, Action: action, Protocol: protocol,
			Port: port, Source: source, Comment: comment,
		})
	}
	return rules
}

func collectNftables() FirewallResult {
	result := FirewallResult{Backend: BackendNftables}
	out, err := runFirewallCmd(binNft, "list", "ruleset")
	if err != nil {
		result.Warning = fmt.Sprintf("nft list ruleset: %v", err)
		return result
	}
	result.Rules = parseNftOutput(string(out))
	return result
}

func parseNftOutput(output string) []FirewallRule {
	var rules []FirewallRule
	var currentChain string
	for _, line := range strings.Split(output, "\n") {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "chain ") {
			if parts := strings.Fields(line); len(parts) >= 2 {
				currentChain = parts[1]
			}
			continue
		}
		if !strings.HasPrefix(line, "tcp dport") && !strings.HasPrefix(line, "udp dport") &&
			!strings.HasPrefix(line, "tcp sport") && !strings.HasPrefix(line, "ip protocol") &&
			!strings.HasPrefix(line, "iif") && !strings.HasPrefix(line, "oif") &&
			!strings.HasPrefix(line, "ct state") && line != "drop" && line != "accept" {
			continue
		}

		rule := FirewallRule{Backend: BackendNftables, Chain: currentChain}
		fields := strings.Fields(line)
		if len(fields) > 0 {
			switch fields[len(fields)-1] {
			case "accept":
				rule.Action = "ACCEPT"
			case "drop":
				rule.Action = "DROP"
			case "reject":
				rule.Action = "REJECT"
			default:
				rule.Action = strings.ToUpper(fields[len(fields)-1])
			}
		}
		for i, f := range fields {
			if f == "dport" && i+1 < len(fields) {
				if fields[i+1] == "{" {
					var ports []string
					for j := i + 2; j < len(fields); j++ {
						if fields[j] == "}" {
							break
						}
						ports = append(ports, strings.TrimRight(fields[j], ","))
					}
					rule.Port = strings.Join(ports, ",")
				} else {
					rule.Port = fields[i+1]
				}
			}
		}
		switch {
		case strings.Contains(line, "tcp "):
			rule.Protocol = "tcp"
		case strings.Contains(line, "udp "):
			rule.Protocol = "udp"
		default:
			rule.Protocol = "all"
		}
		rules = append(rules, rule)
	}
	return rules
}

func collectIptables() FirewallResult {
	result := FirewallResult{Backend: BackendIptables}
	out, err := runFirewallCmd(binIptables, "-L", "-n")
	if err != nil {
		result.Warning = fmt.Sprintf("iptables -L -n: %v", err)
		return result
	}
	result.Rules = parseIptablesOutput(string(out))
	return result
}

func parseIptablesOutput(output string) []FirewallRule {
	var rules []FirewallRule
	var currentChain string
	inChain := false

	for _, line := range strings.Split(output, "\n") {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "Chain ") {
			inChain = false
			if parts := strings.Fields(line); len(parts) >= 2 {
				currentChain = parts[1]
				inChain = true
				if strings.Contains(line, "(policy ") {
					policy := strings.TrimSuffix(strings.Split(line, "(policy ")[1], ")")
					rules = append(rules, FirewallRule{
						Backend: BackendIptables, Chain: currentChain,
						Action: strings.ToUpper(policy), Comment: "default policy",
					})
				}
			}
			continue
		}
		if strings.HasPrefix(line, "target") || strings.HasPrefix(line, "pkts") || line == "" || !inChain {
			continue
		}
		fields := strings.Fields(line)
		if len(fields) < 4 {
			continue
		}
		port := ""
		for _, f := range fields {
			if strings.HasPrefix(f, "dpt:") {
				port = strings.TrimPrefix(f, "dpt:")
			}
		}
		protocol := fields[1]
		if protocol == "0" {
			protocol = "all"
		}
		rules = append(rules, FirewallRule{
			Backend: BackendIptables, Chain: currentChain, Action: fields[0],
			Protocol: protocol, Port: port, Source: fields[3],
		})
	}
	return rules
}

// FormatFirewallForPrompt formats firewall data for the AI prompt.
func FormatFirewallForPrompt(result FirewallResult) string {
	if result.Backend == BackendNone {
		return "read_status: none\n(no firewall backend detected)"
	}

	var b strings.Builder
	_, _ = fmt.Fprintf(&b, "backend: %s\n", result.Backend)
	_, _ = fmt.Fprintf(&b, "read_status: %s\n", result.ReadStatus)
	if result.Warning != "" {
		_, _ = fmt.Fprintf(&b, "collection_warning: %s\n", result.Warning)
	}
	if len(result.Zones) > 0 {
		_, _ = fmt.Fprintf(&b, "zones: %s\n", strings.Join(result.Zones, ", "))
	}
	if len(result.Rules) == 0 {
		b.WriteString("rules: (none collected)\n")
		return strings.TrimRight(b.String(), "\n")
	}

	b.WriteString("rules:\n")
	for i, r := range result.Rules {
		_, _ = fmt.Fprintf(&b, "  rule_%d:\n", i+1)
		if r.Chain != "" {
			_, _ = fmt.Fprintf(&b, "    chain:    %s\n", r.Chain)
		}
		if r.Action != "" {
			_, _ = fmt.Fprintf(&b, "    action:   %s\n", r.Action)
		}
		if r.Protocol != "" {
			_, _ = fmt.Fprintf(&b, "    protocol: %s\n", r.Protocol)
		}
		if r.Port != "" {
			_, _ = fmt.Fprintf(&b, "    port:     %s\n", r.Port)
		}
		if r.Source != "" {
			_, _ = fmt.Fprintf(&b, "    source:   %s\n", r.Source)
		}
		if r.Comment != "" {
			_, _ = fmt.Fprintf(&b, "    comment:  %s\n", r.Comment)
		}
	}
	return strings.TrimRight(b.String(), "\n")
}
