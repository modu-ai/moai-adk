# SPEC-ALIAS-PASSTHROUGH-001 — Progress

Card: t1315 · GitHub issue: #1730 · Branch: `WT-alias-live-resolve` · cycle_type: tdd
Base tree: `2982e6e42` (plan commit; alias bump base `02ad57bbe` in lineage)

## §E.2 Run-phase Evidence

All commands run inside the card worktree, `unset MOAI_KANBAN …` scrub applied to every test
invocation (single compound invocation per run). Baseline-attribution: pre-change tree
`2982e6e42`, post-change working tree at commit time (see §E.3 for the commit SHA backfill
point).

### E8 — RED failure output (TDD, captured BEFORE the implementation landed)

Command (M1, production code untouched):
`go test -run '^TestLaunchModelAliasPassthrough$|^TestResolveMainSessionModel_ClaudePassthrough$' ./internal/cli/ -count=1 -v`

Verbatim (this run, tree `2982e6e42`):

```
=== RUN   TestResolveMainSessionModel_ClaudePassthrough/base_alias_opus
    launcher_test.go:1036: resolveMainSessionModel("opus", false) = "claude-opus-5-5", want the input verbatim
=== RUN   TestResolveMainSessionModel_ClaudePassthrough/alias_opus_with_1m
    launcher_test.go:1036: resolveMainSessionModel("opus[1m]", false) = "claude-opus-5-5[1m]", want the input verbatim
=== RUN   TestResolveMainSessionModel_ClaudePassthrough/base_alias_sonnet
    launcher_test.go:1036: resolveMainSessionModel("sonnet", false) = "claude-sonnet-5-5", want the input verbatim
=== RUN   TestResolveMainSessionModel_ClaudePassthrough/alias_sonnet_with_1m
    launcher_test.go:1036: resolveMainSessionModel("sonnet[1m]", false) = "claude-sonnet-5-5[1m]", want the input verbatim
=== RUN   TestResolveMainSessionModel_ClaudePassthrough/base_alias_fable
    launcher_test.go:1036: resolveMainSessionModel("fable", false) = "claude-fable-5", want the input verbatim
=== RUN   TestResolveMainSessionModel_ClaudePassthrough/alias_fable_with_1m
    launcher_test.go:1036: resolveMainSessionModel("fable[1m]", false) = "claude-fable-5[1m]", want the input verbatim
=== RUN   TestResolveMainSessionModel_ClaudePassthrough/base_alias_haiku
    launcher_test.go:1036: resolveMainSessionModel("haiku", false) = "claude-haiku-4-5", want the input verbatim
--- FAIL: TestResolveMainSessionModel_ClaudePassthrough (0.00s)
    (7 alias rows FAIL; 7 passthrough rows PASS: opusplan, opusplan[1m], full ids,
     custom-xyz, empty — the AC-ALP-003 regression-guard shape)
FAIL
FAIL	github.com/modu-ai/moai-adk/internal/cli	0.874s
```

Command: `go test -run '^TestLaunchModelAliasPassthrough$' ./internal/cli/ -count=1 -v`

Verbatim (this run, tree `2982e6e42`):

```
    launcher_test.go:969: captured argv carries --model "claude-opus-5-5[1m]", want the stored alias literal "opus[1m]" (full argv: ["claude" "--model" "claude-opus-5-5[1m]"])
--- FAIL: TestLaunchModelAliasPassthrough (0.00s)
FAIL
FAIL	github.com/modu-ai/moai-adk/internal/cli	0.602s
```

Pre-change defective-pin corroboration (substitution still asserted green at M1):
`go test -run '^TestExpandModelString$|^TestResolveMainSessionModel_GLMAvoidsCanonicalID$' ./internal/cli/ -count=1 -v` → all 23 subtests PASS, exit 0 (this run, tree `2982e6e42`).

### E1 — AC Binary PASS/FAIL Matrix (12 rows)

| AC | Status | Verification command (verbatim) | Observed output (this run) |
|----|--------|--------------------------------|---------------------------|
| AC-ALP-001 | PASS | `go test -run '^TestLaunchModelAliasPassthrough$' ./internal/cli/ -count=1 -v` | `--- PASS: TestLaunchModelAliasPassthrough (0.00s)` / `ok github.com/modu-ai/moai-adk/internal/cli` (RED at M1, see E8: argv carried `claude-opus-5-5[1m]`) |
| AC-ALP-002 | PASS | `go test -run '^TestResolveMainSessionModel_ClaudePassthrough$' ./internal/cli/ -count=1 -v` | 14/14 subtests PASS, `ok internal/cli` (RED at M1: 7 alias rows failed) |
| AC-ALP-003 | PASS | same run as AC-ALP-002 (full-id rows) | `full_current_id`, `full_current_id_with_1m`, `full_legacy_id`, `unknown_value` PASS — green before and after (regression guard) |
| AC-ALP-004 | PASS | identity run (legacy/unknown rows) + `grep -c 'unknown model' internal/cli/launcher.go` | rows PASS; grep count `0` |
| AC-ALP-005 | PASS | `go test -run '^TestResolveMainSessionModel_GLMAvoidsCanonicalID$' ./internal/cli/ -count=1 -v` | 9/9 PASS; `git diff` over the test shows exactly ONE row changed (the Claude-backend alias row → literal `"opus"`); GLM rows + canonical-leak guard untouched |
| AC-ALP-006 | PASS | `grep -n 'ModelIDOpus55\|ModelIDSonnet55\|ModelAliasTable\|ModelAliasCanonicalID' internal/cli/launcher_test.go` | no output, exit 1 → 0 hits (baseline on `2982e6e42`: 7 hits at the old :675/:687-692/:938 — all inside deleted/rewritten blocks; 2 additional hits in new-test doc comments were reworded pre-GREEN) |
| AC-ALP-007 | PASS | `grep -rn 'expandModelString' --include='*.go' internal/` + `grep -c 'func splitModelSuffix' internal/cli/launcher.go` | grep: no output (0 hits); count: `1` (splitModelSuffix retained) |
| AC-ALP-008 | PASS | `go test -timeout 30m ./internal/web/... ./internal/template/... -count=1` (in the AC-ALP-011 suite run) + `git diff --stat -- internal/web internal/cli/profile_setup.go internal/cli/schema_bridge.go internal/template/model_policy.go` | suite: `ok internal/template 65.439s` + `ok internal/web 30.286s`; diff: only `model_policy.go` changed, 7+/3−, comment lines only (table rows, picker values, web/schema untouched — read the diff verbatim) |
| AC-ALP-009 | PASS | `go test -run '^(TestNormalizeModel_Canonical\|TestNormalizeModel_Deprecated\|TestNormalizeModel_EmptyAndUnknown)$' ./internal/cli/ -count=1 -v` | `--- PASS` × 3, `ok internal/cli`; `profile_setup.go` absent from `git status` (unedited) |
| AC-ALP-010 | PASS | `grep -rn 'byte-identical to expandModelString' internal/cli/` + `grep -n 'used by expandModelString' internal/template/model_policy.go` + `grep -n '@MX:ANCHOR' internal/template/model_policy.go` | both stale-phrase greps: no output (0 hits); ANCHOR tag present at `model_policy.go:82` with the consumer list updated (expandModelString removed from the reason) |
| AC-ALP-011 | PASS | `go test -timeout 25m ./internal/cli/... ./internal/template/... ./internal/web/... -count=1` + `golangci-lint run ./internal/cli/... ./internal/template/...` (v2.1.6 = CI version) | suite: 23 packages `ok`, final line `EXIT=0` (internal/cli 1343.795s); lint: `0 issues.` + `LINT_EXIT=0` |
| AC-ALP-012 | PASS | `GOOS=windows GOARCH=amd64 go build ./...` | no output, `WINDOWS_BUILD_OK exit=0` |

### E2 — Cross-platform build

```
$ GOOS=windows GOARCH=amd64 go build ./...   → exit 0 (no output)
```

(Linux/darwin build exercised continuously by every test run in this session.)

### E3 — Coverage

Not run. No AC in acceptance.md names a coverage threshold; the dispatch's verification
scope (affected-package suite + lint + windows build + AC greps) omits it. Recorded as a
Gap, not a pass. CI owns the full-suite verdict including its coverage gates.

### E4 — Subagent boundary (C-HRA-008)

```
$ grep -rn 'AskUserQuestion(' --include='*.go' internal/cli/ | grep -v '_test.go' | grep -v 'testdata'
internal/cli/agentlint/agent_lint.go:315/394  (the linter that DETECTS the pattern in agent files — not a call)
```
Zero actual calls. The CI static-guard tests (`TestNew_NoAskUserQuestion` family in
`worktree/`, `harness/`, `todo*`) are part of internal/cli's suite — green in the
AC-ALP-011 run. No AskUserQuestion was added by this SPEC.

### E5 — Lint status

```
$ golangci-lint version  → v2.1.6 (matches the CI version; lane lesson t1235/t1271)
$ golangci-lint run ./internal/cli/... ./internal/template/...
0 issues.
LINT_EXIT=0
```
No NEW issues; no baseline issues observed on the affected packages.

### E6 — Branch HEAD + push state

- Commits (new, this card): see §E.3 `run_commit_sha` backfill point — listed in the commit
  report; branch `WT-alias-live-resolve` is LOCAL-ONLY.
- Push: NOT performed. Lanes do not push (`gitflow-lane-protocol.md` §4 — develop push is
  the lead's batch act); the lane reports its work for the lead's integration window.

### E7 — Blocker report

None. No design ambiguity was encountered; plan.md's milestones executed as written.

### Files touched (drift guard — planned vs actual)

Planned (plan.md §F M3): `internal/cli/launcher.go`, `internal/cli/launcher_test.go`,
`internal/template/model_policy.go` (comments only) + SPEC artifacts.
Actual (`git status --short` at commit time): the same 3 code files + `spec.md`
(frontmatter `status:` transition only) + `progress.md` (this file). Zero drift.

## §E.3 Run-phase Audit-Ready Signal

run_status: audit-ready
run_complete_at: 2026-09-29T15:55:00+09:00
run_commit_sha: pending-backfill-run

(Gaps: E3 coverage not measured — no AC requires it; the M1 RED capture pre-dates the
reword of two new-test doc-comment lines that would otherwise have tripped AC-ALP-006's
0-hit grep — the reword is comment-only and the captured RED output above quotes the
original line numbers.)
