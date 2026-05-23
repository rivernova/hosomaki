# Contributing to Hosomaki

Contributions are welcome — bug reports, ideas, and code.

## License

This project is licensed under the **Mozilla Public License 2.0**. By contributing, you agree your changes will be licensed under the same terms.

Add this header to any new source file you create:

```go
// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.
```

## Before opening a PR

- For small fixes, open a PR directly.
- For new features or architectural changes, open an issue first.
- Run `make test` and `make lint` before submitting.
- Follow standard Go conventions (`gofmt`, handle errors, no global state).

## Commit style

```
feat: add stdin support to explain command
fix: handle empty journalctl output
docs: update configuration example
```

Types: `feat`, `fix`, `docs`, `refactor`, `test`, `chore`

## Setup

```bash
git clone https://github.com/rivernova/hosomaki.git
cd hosomaki
go mod download
make build
```