# progress.md — SPEC-HOOK-SYNC-ASYNC-001

## §E.1 Plan-phase Audit-Ready Signal

audit_ready: true
plan_status: audit-ready
plan_complete_at: 2026-10-16T22:30:39+09:00
plan_audit: .moai/reports/t1780/plan-audit.md, PASS, overall_score 0.9375, must_pass_failed 0, blocking_count 0, audited_sha 7ec170cca15d3a2976b41f915cf38ff96a4e2581

## §E.2 Run-phase Evidence

### M3 — codex-review-entry gate-ON shell fast path (2026-10-21, commits 772b9a1590 + 3483d65726)

TDD order observed: the M3 suite (internal/cli/codex_review_entry_fastpath_test.go) ran RED
before the wrapper implementation existed — the verbatim pre-GREEN failures showed the
current forward behavior on every ALLOW state (`argv counts (forward=1 kick=0) != want
(forward=0 kick=1)`, `hook codex-review-entry` recorded on no-receipt / fresh-pass /
config-only states) and `the wrapper does not declare CODEX_ENTRY_RUNTIME_PREFIXES` /
`the live wrapper carries no "snapshots"` for the golden and mirror pins; the 6 fall-through
preservation cases (fresh FAIL, missing git, both sha tools absent, WT- branch, tree_scope
skip, primary skip) were green pre-M3 exactly as their regression-guard classification
predicts.

| AC | Status | Verification command | Actual output (this run, tree 3483d65726) |
|----|--------|----------------------|--------------------------------------------|
| AC-HSA-006 | PASS | `go test ./internal/cli/ -run 'TestCodexEntryFastpath' -count=1` | `ok github.com/modu-ai/moai-adk/internal/cli 17.224s` (exit 0; BothSurfacesCarryProbe asserts the probe strings and byte equality in BOTH wrapper surfaces) |
| AC-HSA-007 | PASS | same run — ProbeMatrix cases no-receipt / fresh-FAIL / missing-git / missing-shasum / both-sha-absent | no-receipt: `forward=0 kick=1`; fresh FAIL: `forward=1 kick=0`; probe failures: `forward=1 kick=0` — all exit 0 |
| AC-HSA-008 | PASS | same run — KeyParity | shell digest == `verify.Key` digest on the `shasum` AND the `openssl` pipeline (64-hex, first 16 compared) |
| AC-HSA-015 | PASS | same run — SingleKickPerBucket | two entries one bucket: `kick count = 1`; pre-created Stop-arm marker: `kick count = 0`; fresh pass: `forward=0 kick=0` |
| AC-HSA-016 | PASS | same run — ProbeMatrix WT- branch case | `forward=1 kick=0`, exit 0 — the Go handler reached exactly once, no shell probe state consulted |
| AC-HSA-017 | PASS | same run — ProbeMatrix gate-off case | exit 0, stub never invoked (argv file absent) |

Probe matrix: 18/18 green (gate off, no receipt, stale receipt, unparseable snapshot, fresh
pass, fresh inconclusive, fresh FAIL, clean tree, config-only-only, behind binary, HEAD-match
binary, no-commit-metadata binary, missing git, missing shasum, both sha tools absent, WT-
branch, tree_scope skip, primary-checkout session). Golden test pins the shell
runtime-config prefix lists to `reviewGateRuntimePrefixes` /
`reviewGateTreeConfigExactPaths` / `reviewGateTreeConfigDirPrefixes`.

Companion evidence: `go build ./...` exit 0; `GOOS=windows GOARCH=amd64 go build ./...`
exit 0; `go vet ./internal/cli/...` exit 0; `gofmt -l` empty on touched dirs;
`golangci-lint run --timeout=2m` = the 2 pre-existing staticcheck findings in
internal/cli/paths.go (ST1008), 0 new; `go test ./internal/template/ -run 'ReviewGate|
Registration|Codex|HookSync' -count=1` = `ok ... 0.455s`; `go test ./internal/cli/ -run
'Codex' -count=1` = `ok ... 226.440s`; `grep -c 'AskUserQuestion\|mcp__askuser'` over the
new test and both wrappers = 0.

## §E.3 Run-phase Audit-Ready Signal

_<pending run-phase>_

## §E.4 Sync-phase Audit-Ready Signal

_<pending sync-phase>_

## §F Phase 4 Mode Selection

Logged 2026-10-21 by lane-5 (run tn8rpj) before the first run-phase Agent() spawn.

**Input parameters**: tier=M · scope≈10 files (hook shell + template mirror + settings.json/.tmpl
entries + hooks-system.md + internal/hook Go handler + tests) · domain count=4 (shell hooks,
Go hook handler, template/settings, SPEC docs) · file language mix=shell/Go/JSON/YAML/Markdown ·
concurrency benefit=LOW (coding-heavy, milestone-ordered; M3 leads design reversibility) ·
Agent Teams prereqs=not requested (no operator --team).

**Mode evaluation**:

| Mode | Selected | Rationale |
|---|---|---|
| direct | not selected | semantic multi-file change, no trivial case |
| serial | **selected** | coding-heavy per-milestone implementation, one writer per tree |
| fanout | not selected | coding-task parallelism caveat — no research-heavy fan-out in scope |
| sweep | not selected | <30 files, non-mechanical, interdependent edits |

**Decision: serial** — one manager-develop spawn per milestone group (M3, then M4+M5, then M7
verification close), sequential, same worktree.

**Justification**: Anthropic's coding-task parallelism caveat applies — the milestones are
ordered by decision reversibility (plan §F) and share files (settings.json, hooks-system.md
appear in multiple milestones), so parallel spawns would need write serialization anyway. The
M2 gate observation runs as a lane-local verification, not a spawn. Kickoff decision record:
`.moai/reports/t1913/progress.md` § Kickoff (autonomous form — plan-audit PASS 0.9375 ≥ Tier M
0.80, artifact hash unchanged, leader tn8rpj dispatch).
