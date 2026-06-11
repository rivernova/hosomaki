// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package collector

import (
	"fmt"
	"os/exec"
	"strconv"
	"strings"
)

// collects context, the why command needs to reconstruct the full failure chain,
// which requires the INFO and DEBUG lines before the crash,
// not just the error lines that record it

func WhyLogs(service string, opts LogOptions) (string, error) {
	n := lines(opts.Lines, defaultServiceLines)

	args := []string{"-u", service, "-n", strconv.Itoa(n)}
	if opts.Since != "" {
		args = append(args, "--since", opts.Since)
	}
	args = append(args, journalctl.format...)

	out, err := exec.Command(binJournalctl, args...).Output() // #nosec G204
	if err != nil {
		return "", fmt.Errorf(
			"could not collect journal for service %q: %w", service, err,
		)
	}

	text := strings.TrimSpace(string(out))
	if !isJournalContent(text) {
		return "", fmt.Errorf(
			"no journal entries found for service %q — "+
				"is the service name correct and has it run recently?",
			service,
		)
	}

	return text, nil
}
