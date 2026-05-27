// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package collector

import (
	"bufio"
	"os"
	"os/exec"
	"strings"
)

// this file contains environment detection
type Environment struct {
	// Distro
	DistroID         string
	DistroLike       string
	DistroVersion    string
	DistroPrettyName string

	// Kernel & arch
	Kernel       string
	KernelFull   string
	Architecture string

	// Userspace
	InitSystem     string
	PackageManager string
	Shell          string
	Hostname       string

	// Security / isolation
	SELinux        string
	AppArmor       string
	Virtualisation string
}

func Env() Environment {
	e := Environment{}

	readOSRelease(&e)

	if v, err := exec.Command("uname", "-r").Output(); err == nil {
		e.Kernel = strings.TrimSpace(string(v))
	}
	if v, err := exec.Command("uname", "-srm").Output(); err == nil {
		e.KernelFull = strings.TrimSpace(string(v))
	}
	if v, err := exec.Command("uname", "-m").Output(); err == nil {
		e.Architecture = strings.TrimSpace(string(v))
	}

	e.InitSystem = detectInitSystem()
	e.PackageManager = detectPackageManager(e.DistroID, e.DistroLike)
	e.Shell = detectShell()
	e.Hostname, _ = os.Hostname()

	e.SELinux = detectSELinux()
	e.AppArmor = detectAppArmor()
	e.Virtualisation = detectVirtualisation()

	return e
}

func readOSRelease(e *Environment) {
	f, err := os.Open("/etc/os-release")
	if err != nil {
		return
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		key, val, ok := splitKeyValue(scanner.Text())
		if !ok {
			continue
		}
		switch key {
		case "ID":
			e.DistroID = val
		case "ID_LIKE":
			e.DistroLike = val
		case "VERSION_ID":
			e.DistroVersion = val
		case "PRETTY_NAME":
			e.DistroPrettyName = val
		}
	}
}

func splitKeyValue(line string) (key, value string, ok bool) {
	line = strings.TrimSpace(line)
	if line == "" || strings.HasPrefix(line, "#") {
		return "", "", false
	}
	idx := strings.IndexByte(line, '=')
	if idx <= 0 {
		return "", "", false
	}
	key = line[:idx]
	value = strings.TrimSpace(line[idx+1:])
	if len(value) >= 2 {
		first, last := value[0], value[len(value)-1]
		if (first == '"' && last == '"') || (first == '\'' && last == '\'') {
			value = value[1 : len(value)-1]
		}
	}
	return key, value, true
}

func detectInitSystem() string {
	data, err := os.ReadFile("/proc/1/comm")
	if err != nil {
		return ""
	}
	name := strings.TrimSpace(string(data))
	switch name {
	case "systemd":
		return "systemd"
	case "runit":
		return "runit"
	case "openrc-init", "openrc":
		return "openrc"
	case "init":
		if _, err := os.Stat("/run/openrc"); err == nil {
			return "openrc"
		}
		return "sysvinit"
	}
	return name
}

func detectPackageManager(id, idLike string) string {
	fields := append([]string{id}, strings.Fields(idLike)...)
	for _, name := range fields {
		switch name {
		case "debian", "ubuntu", "linuxmint", "pop", "elementary", "kali", "raspbian":
			return "apt"
		case "fedora", "rhel", "centos", "rocky", "almalinux", "ol":
			return "dnf"
		case "arch", "manjaro", "endeavouros", "garuda", "cachyos":
			return "pacman"
		case "opensuse", "opensuse-leap", "opensuse-tumbleweed", "sles", "suse":
			return "zypper"
		case "alpine":
			return "apk"
		case "void":
			return "xbps"
		case "gentoo":
			return "emerge"
		case "nixos":
			return "nix"
		}
	}

	for _, p := range []struct{ bin, name string }{
		{"/usr/bin/apt", "apt"},
		{"/usr/bin/dnf", "dnf"},
		{"/usr/bin/pacman", "pacman"},
		{"/usr/bin/zypper", "zypper"},
		{"/sbin/apk", "apk"},
		{"/usr/bin/xbps-install", "xbps"},
		{"/usr/bin/emerge", "emerge"},
	} {
		if _, err := os.Stat(p.bin); err == nil {
			return p.name
		}
	}
	return ""
}

func detectShell() string {
	s := os.Getenv("SHELL")
	if s == "" {
		return ""
	}
	if i := strings.LastIndexByte(s, '/'); i >= 0 {
		return s[i+1:]
	}
	return s
}

func detectSELinux() string {
	data, err := os.ReadFile("/sys/fs/selinux/enforce")
	if err != nil {
		return ""
	}
	switch strings.TrimSpace(string(data)) {
	case "1":
		return "Enforcing"
	case "0":
		return "Permissive"
	default:
		return "Disabled"
	}
}

func detectAppArmor() string {
	if _, err := os.Stat("/sys/kernel/security/apparmor"); err == nil {
		return "enabled"
	}
	return ""
}

func detectVirtualisation() string {
	out, err := exec.Command("systemd-detect-virt").Output()
	if err != nil {
		if len(out) > 0 {
			return strings.TrimSpace(string(out))
		}
		return ""
	}
	return strings.TrimSpace(string(out))
}
