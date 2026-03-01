---
description: Go style and conventions for Neptune
globs: "**/*.go"
alwaysApply: false
---

# Go Standards

- **Format**: Use `gofmt -s`. Run `make check-fmt` before committing. Use `git commit -s` for DCO sign-off (see [CONTRIBUTING.md](CONTRIBUTING.md)).
- **Lint**: Follow `.golangci.yml`; do not introduce new linter violations. Run `make lint` when available.
- **Packages**: Code in `internal/` must not be imported from outside this module.
- **Errors**: Always handle errors; add context with `fmt.Errorf("...: %w", err)` where useful.
- **Tests**: Place `*_test.go` in the same package as the code. Use table-driven tests when testing multiple cases.
- **Exports**: Public functions and types must have doc comments starting with the name (e.g. `// NewClient creates ...`).
