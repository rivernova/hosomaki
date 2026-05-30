// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package ai

import (
	"context"
	"io"
)

// this file contains the definition of an abstraction for ai providers

type Provider interface {
	Generate(ctx context.Context, prompt string) (string, error)
	GenerateStream(ctx context.Context, prompt string, onFirstToken func(), w io.Writer) (string, error)
}
