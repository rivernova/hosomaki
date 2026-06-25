// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package ai

import (
	"context"
	"fmt"
)

// CheckProviderHealth verifies that a provider is reachable before routing requests.
func CheckProviderHealth(ctx context.Context, p Provider, name string) error {
	if err := p.HealthCheck(ctx); err != nil {
		return fmt.Errorf("%s: %w", name, err)
	}
	return nil
}
