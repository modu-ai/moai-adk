# t661 window delta judgment (absorb c005ec1f1: ^1 e09900ef5, ^2 local develop 84e5666d9)

origin/develop ed71054d3 is an ancestor of local develop (left-right 0 31), so local develop was current.
Absorbed files: 255 (delta-files.txt); 33 distinct dirs.
internal/cli test dependency dirs: `go list -deps -test ./internal/cli/` filtered to module github.com/modu-ai/moai-adk -> 117 dirs.
Intersection: 7 dirs: internal/cli, internal/cli/update/backup, internal/cli/update/report, internal/config, internal/config/toolpolicy, internal/spec, internal/template.
Controls on the same matcher: internal/cli matched (positive), not/a/real/pkg not matched (negative).
go.mod / go.sum: not in the delta.
Sink files touched by the absorb: update.go, update_clean_install.go, update_template_sync.go (135 diff lines). Added/removed lines naming userHomeDirFn, paths.Home, UserHomeDir, ensureGlobalSettingsEnv, RemoveAll or WriteFile: 0; only comment lines mention .claude/skills paths. Sink now at update.go:931, call site update_template_sync.go:590.
Verdict: delta reaches internal/cli tests -> re-measure with the lead-approved card-test selector plus home fingerprint.
