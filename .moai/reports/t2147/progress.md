# t2147 progress - bump version to v3.2.1

Template: previous bump a85be12449 (v3.2.0; also eba919e44b for v3.1.3). They touched version.go, system.yaml, 4 READMEs, docs-site/hugo.toml (a85be12449 also edited CHANGELOG + tests; CHANGELOG NOT edited here, per card).

## Files changed (commit 801eba9d92)
- pkg/version/version.go: v3.2.0 -> v3.2.1
- .moai/config/sections/system.yaml: template_version and version v3.2.0 -> v3.2.1
- README.md, README.ko.md, README.ja.md, README.zh.md line 24: Release-v3.1.2 -> Release-v3.2.1
- docs-site/hugo.toml: version v3.2.0 -> v3.2.1; releaseDate 2026-10-23 -> 2026-11-07 (not in lane pre-read; previous bump changed it)
- system.yaml.tmpl uses {{.Version}}; no literal, unchanged. No make build needed (no template edit).

## Verification
- go build ./... : no output (empty = success; exit status masked by `| tail`)
- go test -count=1 ./pkg/version/ : ok  github.com/modu-ai/moai-adk/pkg/version 0.256s
- go test -count=1 -run TestVersionStamp ./internal/cli/ : ok  github.com/modu-ai/moai-adk/internal/cli 8.722s

## Gaps
- releaseDate 2026-11-07 is an assumption (today); confirm against the actual release date.
- TestReleasePreflight* and other tests not run; no full suite.
