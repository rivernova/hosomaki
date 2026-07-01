// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package hosomaki

import (
	"context"
	"io"
)

// reusable fake provider for driving command runners with scripted LLM output

type fakeProvider struct {
	stream    string
	repairs   []string
	repairIdx int
	jsonCall  int
}

func (f *fakeProvider) Generate(_ context.Context, _ string) (string, error) {
	return f.stream, nil
}

func (f *fakeProvider) GenerateStream(_ context.Context, _ string, onFirstToken func(), w io.Writer) (string, error) {
	if onFirstToken != nil {
		onFirstToken()
	}
	if w != nil {
		if _, err := io.WriteString(w, f.stream); err != nil {
			return "", err
		}
	}
	return f.stream, nil
}

func (f *fakeProvider) GenerateJSON(_ context.Context, _ string, onFirstToken func()) (string, error) {
	f.jsonCall++
	if onFirstToken != nil {
		onFirstToken()
	}
	if len(f.repairs) == 0 {
		return f.stream, nil
	}
	idx := f.repairIdx
	f.repairIdx++
	if idx >= len(f.repairs) {
		return f.repairs[len(f.repairs)-1], nil
	}
	return f.repairs[idx], nil
}
