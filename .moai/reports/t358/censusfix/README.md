# censusfix — provenance of the census fixture

Card **t358** · SPEC-CI-TEST-OBSERVABILITY-001 · captured 2026-08-29 in worktree
`.claude/worktrees/t358` at HEAD `1a635aea8`.

`scripts/ci-census/testdata/fixture.jsonl` is not hand-written JSON: every event
shape in it was produced by a real `go test -json` run of the throwaway module
kept here. This directory is that module plus the raw streams, so a later reader
can re-derive the fixture rather than trust it.

Sources carry a `.txt` suffix (`go.mod.txt`, `*.go.txt`) so this tree is inert:
it is not a nested Go module and its files are not Go sources. To reproduce,
copy the tree somewhere outside the repo and strip the suffixes.

| File | Command that produced it | Shape it contributes |
|---|---|---|
| `stream.jsonl` | `go test -count=1 -json ./...` (alpha + beta) | pass / `t.Skip` / failing test with captured output / package with no test files |
| `broken.jsonl` | same, with a syntax error planted in `beta` | `build-output` + `build-fail`, and a package fail carrying `FailedBuild` |
| `bad.jsonl` | `go test -count=1 -json ./nonexistent/...` | the `[setup failed]` path |
| `gamma.jsonl` | `go test -count=1 -json ./gamma/...` (`TestMain` calls `os.Exit(1)`) | a package-level fail with NO failing test and NO build failure |

## The finding these captures produced

`*-stderr.txt` are all **0 bytes**. Under `-json`, `go test` writes compiler
diagnostics into the event stream as `build-output` / `build-fail` events on
**stdout**; nothing reaches stderr. So a CI step that redirects stdout to a file
does not leave the build error console-visible — it swallows it — unless the
census prints those events back. That is why `test-census.sh` emits `BUILD
FAILED` rows, and it corrects the assumption recorded in `spec.md` §F.2 and
`plan.md` §B.2.
