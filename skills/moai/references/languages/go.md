# Language: Go

When to read: `internal/quality` sent you here because the project's detected languages include `go`. The project's own declarations (CI steps, `Makefile`, `magefile.go`, `AGENTS.md`) win over every default below.

## Evidence and defaults

| Kind | Default command | Use it only when |
| --- | --- | --- |
| format | `gofmt -l .` | `go.mod`; any file it lists is unformatted (the command exits 0 either way, so read its output) |
| lint | `go vet ./...` | `go.mod` |
| lint | `golangci-lint run` | `.golangci.yml`, `.golangci.yaml` or `.golangci.toml` exists |
| type check | covered by `go vet ./...` and the build | |
| test | `go test ./...` | `go.mod`; add `-count=1` so a cached result is not reported as a new run |
| build | `go build ./...` | `go.mod` |

## Dependency existence

- `go list -m -versions <module>` lists the versions of a module path; an unknown path ends with an error. Module paths are repository paths (`github.com/<owner>/<repo>`), so also check that the repository is the one the documentation names.
- Add with `go get <module>@<version>`; show the `go.mod` and `go.sum` lines it wrote.

## Traps

- `go test` prints `(cached)` for an unchanged package: that is an earlier result. Use `-count=1` at VERIFY.
- `[no test files]` for a package means nothing ran there.
- `go mod tidy` rewrites `go.mod` and `go.sum`; a VERIFY run never runs it.
- Build tags select files (`//go:build`): run the tags the CI runs.
- `t.Skip` passes silently; a new skip is a finding.
