// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package ai

import (
	"context"
)

// interface for ai providers

type Provider interface {
	Generate(ctx context.Context, prompt string) (string, error)
	GenerateJSON(ctx context.Context, prompt string, onFirstToken func()) (string, error)
}
