// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package collector

import (
	"bufio"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// procfs collection for the --pid flag of the explain command

type ProcessSnapshot struct {
	PID              int
	Name             string
	State            string
	VmRSS            string
	Threads          string
	UID              string
	OpenFiles        []string
	Sockets          []string
	CollectionErrors []string
}

var ErrProcessNotFound = errors.New("process not found")

var ErrProcessPermission = errors.New("permission denied reading process information")

func ProcessInfo(pid int) (*ProcessSnapshot, error) {
	dir := fmt.Sprintf("/proc/%d", pid)

	if err := checkProcDir(dir); err != nil {
		return nil, err
	}

	snap := &ProcessSnapshot{PID: pid}

	if err := collectStatus(snap, dir); err != nil {
		snap.CollectionErrors = append(snap.CollectionErrors, fmt.Sprintf("status: %v", err))
	}

	if err := collectFDs(snap, dir); err != nil {
		snap.CollectionErrors = append(snap.CollectionErrors, fmt.Sprintf("file descriptors: %v", err))
	}

	collectSockets(snap, dir)

	return snap, nil
}

func checkProcDir(dir string) error {
	_, err := os.Lstat(dir)
	if err != nil {
		if os.IsNotExist(err) {
			return ErrProcessNotFound
		}
		if os.IsPermission(err) {
			return ErrProcessPermission
		}
		return ErrProcessNotFound
	}
	return nil
}

func collectStatus(snap *ProcessSnapshot, dir string) error {
	path := filepath.Join(dir, "status")
	f, err := os.Open(path)
	if err != nil {
		if os.IsPermission(err) {
			return ErrProcessPermission
		}
		return fmt.Errorf("open %s: %w", path, err)
	}
	defer func() { _ = f.Close() }()

	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := scanner.Text()
		key, val, ok := strings.Cut(line, ":")
		if !ok {
			continue
		}
		val = strings.TrimSpace(val)
		switch key {
		case "Name":
			snap.Name = val
		case "State":
			snap.State = val
		case "VmRSS":
			snap.VmRSS = val
		case "Threads":
			snap.Threads = val
		case "Uid":
			if fields := strings.Fields(val); len(fields) > 0 {
				snap.UID = fields[0]
			}
		}
	}
	return scanner.Err()
}

func collectFDs(snap *ProcessSnapshot, dir string) error {
	fdDir := filepath.Join(dir, "fd")
	entries, err := os.ReadDir(fdDir)
	if err != nil {
		if os.IsPermission(err) {
			return ErrProcessPermission
		}
		return fmt.Errorf("readdir %s: %w", fdDir, err)
	}

	seen := make(map[string]struct{})
	for _, entry := range entries {
		target, linkErr := os.Readlink(filepath.Join(fdDir, entry.Name()))
		if linkErr != nil {
			// The FD may have been closed between ReadDir and Readlink
			continue
		}

		label := normaliseFDTarget(target)
		if _, dup := seen[label]; dup {
			continue
		}
		seen[label] = struct{}{}
		snap.OpenFiles = append(snap.OpenFiles, label)
	}
	return nil
}

func normaliseFDTarget(target string) string {
	switch {
	case strings.HasPrefix(target, "socket:["):
		return "socket (anonymous)"
	case strings.HasPrefix(target, "pipe:["):
		return "pipe (anonymous)"
	default:
		return target
	}
}

func collectSockets(snap *ProcessSnapshot, dir string) {
	for _, name := range []string{"tcp", "tcp6"} {
		path := filepath.Join(dir, "net", name)
		lines, err := readNetTCP(path)
		if err != nil {
			snap.CollectionErrors = append(snap.CollectionErrors,
				fmt.Sprintf("net/%s: %v", name, err))
			continue
		}
		snap.Sockets = append(snap.Sockets, lines...)
	}
}

var tcpState = map[string]string{
	"01": "ESTABLISHED",
	"02": "SYN_SENT",
	"03": "SYN_RECV",
	"04": "FIN_WAIT1",
	"05": "FIN_WAIT2",
	"06": "TIME_WAIT",
	"07": "CLOSE",
	"08": "CLOSE_WAIT",
	"09": "LAST_ACK",
	"0A": "LISTEN",
	"0B": "CLOSING",
}

func readNetTCP(path string) ([]string, error) {
	f, err := os.Open(path)
	if err != nil {
		if os.IsPermission(err) {
			return nil, ErrProcessPermission
		}
		return nil, fmt.Errorf("open: %w", err)
	}
	defer func() { _ = f.Close() }()

	var results []string
	scanner := bufio.NewScanner(f)
	first := true
	for scanner.Scan() {
		if first {
			first = false
			continue // skip header line
		}
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}
		if summary, ok := parseTCPLine(line); ok {
			results = append(results, summary)
		}
	}
	return results, scanner.Err()
}

func parseTCPLine(line string) (string, bool) {
	fields := strings.Fields(line)
	if len(fields) < 4 {
		return "", false
	}

	local := decodeHexAddr(fields[1])
	remote := decodeHexAddr(fields[2])
	state := tcpState[strings.ToUpper(fields[3])]
	if state == "" {
		state = "state-" + fields[3]
	}

	if remote == "0.0.0.0:0" || remote == "[::]:0" {
		return fmt.Sprintf("%s  %s  local=%s", "tcp", state, local), true
	}
	return fmt.Sprintf("%s  %s  local=%s  remote=%s", "tcp", state, local, remote), true
}

func decodeHexAddr(raw string) string {
	addrPart, portPart, ok := strings.Cut(raw, ":")
	if !ok {
		return raw
	}

	port, portErr := parseHexUint16(portPart)

	switch len(addrPart) {
	case 8: // IPv4 — 4 bytes
		b := make([]byte, 4)
		if n, _ := fmt.Sscanf(addrPart, "%02X%02X%02X%02X", &b[3], &b[2], &b[1], &b[0]); n != 4 {
			return raw
		}
		ip := fmt.Sprintf("%d.%d.%d.%d", b[0], b[1], b[2], b[3])
		if portErr != nil {
			return ip
		}
		return fmt.Sprintf("%s:%d", ip, port)

	case 32: // IPv6 — 16 bytes
		var w [4]uint32
		if n, _ := fmt.Sscanf(addrPart, "%08X%08X%08X%08X", &w[0], &w[1], &w[2], &w[3]); n != 4 {
			return raw
		}
		ip := fmt.Sprintf("%08x%08x%08x%08x", reverseBytes32(w[0]), reverseBytes32(w[1]),
			reverseBytes32(w[2]), reverseBytes32(w[3]))
		formatted := fmtIPv6Hex(ip)
		if portErr != nil {
			return "[" + formatted + "]"
		}
		return fmt.Sprintf("[%s]:%d", formatted, port)

	default:
		return raw
	}
}

func parseHexUint16(s string) (uint16, error) {
	var v uint64
	_, err := fmt.Sscanf(s, "%X", &v)
	return uint16(v), err
}

func reverseBytes32(v uint32) uint32 {
	return (v>>24)&0xFF | (v>>8)&0xFF00 | (v<<8)&0xFF0000 | (v<<24)&0xFF000000
}

func fmtIPv6Hex(s string) string {
	if len(s) != 32 {
		return s
	}
	var b strings.Builder
	for i := 0; i < 8; i++ {
		if i > 0 {
			b.WriteByte(':')
		}
		b.WriteString(s[i*4 : i*4+4])
	}
	return b.String()
}

func FormatProcessSnapshotForPrompt(snap *ProcessSnapshot) string {
	var b strings.Builder

	_, err := fmt.Fprintf(&b, "PID: %d\n", snap.PID)
	if err != nil {
		return ""
	}
	if snap.Name != "" {
		_, err := fmt.Fprintf(&b, "Process name: %s\n", snap.Name)
		if err != nil {
			return ""
		}
	}
	if snap.State != "" {
		_, err := fmt.Fprintf(&b, "State: %s\n", snap.State)
		if err != nil {
			return ""
		}
	}
	if snap.VmRSS != "" {
		_, err := fmt.Fprintf(&b, "Memory (RSS): %s\n", snap.VmRSS)
		if err != nil {
			return ""
		}
	}
	if snap.Threads != "" {
		_, err := fmt.Fprintf(&b, "Threads: %s\n", snap.Threads)
		if err != nil {
			return ""
		}
	}
	if snap.UID != "" {
		_, err := fmt.Fprintf(&b, "UID: %s\n", snap.UID)
		if err != nil {
			return ""
		}
	}

	b.WriteString("\n--- Open files and sockets ---\n")
	if len(snap.OpenFiles) == 0 {
		b.WriteString("(none visible)\n")
	} else {
		for _, f := range snap.OpenFiles {
			_, err := fmt.Fprintf(&b, "  %s\n", f)
			if err != nil {
				return ""
			}
		}
	}

	b.WriteString("\n--- TCP socket state ---\n")
	if len(snap.Sockets) == 0 {
		b.WriteString("(none visible)\n")
	} else {
		for _, s := range snap.Sockets {
			_, err := fmt.Fprintf(&b, "  %s\n", s)
			if err != nil {
				return ""
			}
		}
	}

	if len(snap.CollectionErrors) > 0 {
		b.WriteString("\n--- Collection warnings ---\n")
		for _, e := range snap.CollectionErrors {
			_, err := fmt.Fprintf(&b, "  warning: %s\n", e)
			if err != nil {
				return ""
			}
		}
	}

	return strings.TrimSpace(b.String())
}
