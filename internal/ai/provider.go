// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package ai

import "context"

// Provider is the contract every AI backend must satisfy.
// Adding a new backend (Anthropic, OpenAI, …) means implementing this
// interface — nothing else in the codebase needs to change.
type Provider interface {
	// Generate sends a fully-formed prompt and returns the model's response.
	// Callers are responsible for building the prompt; the provider only
	// handles transport and parsing.
	Generate(ctx context.Context, prompt string) (string, error)
}
