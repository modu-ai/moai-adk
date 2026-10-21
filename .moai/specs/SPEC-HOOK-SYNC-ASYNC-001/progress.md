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

### Repair round 1 — R1-R6 (2026-10-22, commits d09e0d6dbb + f823d6a953)

Sources: `.moai/reports/t1913/card-review.md` (codex review 1, fail) + `.moai/reports/t1913/sync-audit.md`
(FAIL 7.33/10, rcpt-7ba78473422dd10fc70c31d8) consolidated repair list. Every repair was
RED-first — the new fixtures/cases were observed failing on the pre-repair wrapper before the
fix landed (RED run: `10 FAIL` lines over KeyParity, KeyFallsThroughOnUntrackedDir → renamed
KeyDeclinesUnreproducibleEntries, GoldenConfigDigest, and the 8 new/changed probe subtests).

| R | Repair | RED observation (pre-repair) | GREEN observation (this run, tree f823d6a953) |
|---|--------|------------------------------|------------------------------------------------|
| R1 | payload cwd/project_dir precedence — diverging stdin session dir falls through | `payload_cwd_diverges... (forward=0 kick=1) != want (forward=1 kick=0)` | `forward=1 kick=0`, exit 0; same-root payloads still proceed |
| R2 | untracked directory entries decline (embedded repo = ls-files dir entry; Go WalkDirs descendants) | KeyParity digest mismatch + `codex_entry_key succeeded on a tree whose untracked set contains a directory` | KeyDeclinesUnreproducibleEntries/untracked_directory_entry PASS (non-zero exit) |
| R3 | symlink entries decline (os.Readlink bytes not portably readable; ln -s strips trailing newlines — fixtures built with os.Symlink) | KeyParity digest mismatch on the newline-symlink shape | plain + newline-target decline subtests PASS |
| R4 | log file probed openable pre-spawn (/dev/null arm); pre-exec spawn death releases the marker | `unopenable_log_path (forward=0 kick=0) != want (forward=0 kick=1)` | `forward=0 kick=1`, exit 0 — the kick fires via the /dev/null arm |
| R5 | stamp without a -g component falls through; Go owns the leniency | `no_commit_component... (forward=0 kick=1) != want (forward=1 kick=0)` | `forward=1 kick=0`, exit 0 |
| R6 | receipt binding adds config_digest + tool_version (CheckReceipt :73/:75): unbound → fall through, mismatch → miss arm | all three new cases RED (`tool_version_mismatch (forward=0 kick=0) != want kick=1`, `config_digest_mismatch ...`, `receipt_without_bound_fields (forward=0 kick=0) != want (forward=1 kick=0)`) | mismatch cases kick (`forward=0 kick=1`); unbound falls through (`forward=1 kick=0`); fresh pass/inconclusive/fail retain their M3 dispositions with bound-field-matched fixture receipts |

Golden pin: `TestCodexEntryFastpath_GoldenConfigDigest` — the shell's length-prefixed
emission (gate/codex-review, method/review/start, target/uncommittedChanges) hashes to
`verify.ConfigDigest`'s output byte-for-byte. Probe matrix now 24 subtests, all green:
`go test ./internal/cli/ -run 'TestCodexEntryFastpath' -count=1` → `ok ... 27.687s`, exit 0
(this run, tree f823d6a953). Companion: builds (native + GOOS=windows) exit 0; vet exit 0;
gofmt clean; golangci-lint = the same 2 pre-existing paths.go findings, 0 new; template
registration suite `ok ... 0.582s`; wrapper pair `cmp` byte-identical.

### M4 + M5 — async conversions (2026-10-21, commits 603ad65bb9 + a29a17cbeb + 4a0fa991ed)

TDD order observed: RED captured before each implementation commit. M4 RED — the
updated shape tests failed with the old emission shape (`SystemMessage = "TimeoutError: ..."
want empty (async hooks deliver additionalContext only)` + `HookSpecificOutput is nil`),
three registry cases failed (`HookSpecificOutput is nil, want non-nil`), and the new
registration test failed (`PostToolUseFailure handle-post-tool-failure.sh carries no async
key` on BOTH surfaces). M5 RED — the registration test failed for ConfigChange and
chain-event on both surfaces, and TestHookEntryParity_ChainEventRegistered failed with
`chain-event.sh must carry async=true`.

Commit 603ad65bb9 (M3 follow-up): the identity-guard list gained the `t@fixture` literal
the M3 commit (3483d65726) introduced — `TestCommitIdentityGuard_BuiltinListCoversFixtureEnumeration`
failed on the M3 tip (baseline evidence: the fixture test file is identical to
7addd36d15 and `git show 7addd36d15:internal/hook/commit_identity_guard.go | grep -c
'"t@fixture"'` = 0); list count 62 -> 63, test and predicate unweakened.

| AC | Status | Verification command | Actual output (this run, tree 4a0fa991ed) |
|----|--------|----------------------|--------------------------------------------|
| AC-HSA-001 | PASS | `go test ./internal/template/ -run 'TestHookSyncAsyncRegistrationFlags' -count=1` | `ok github.com/modu-ai/moai-adk/internal/template 0.443s` (ConfigChange row async=true in live + rendered template) |
| AC-HSA-002 | PASS | same run | live settings.json ConfigChange row: `grep -A 3 'handle-config-change.sh"'` then `grep -c '"async": true'` = 1 |
| AC-HSA-003 | PASS | same run + `go test ./internal/cli/ -run 'TestHookEntryParity' -count=1` | `ok ... internal/template 0.443s`; `ok ... internal/cli 1.369s` (3/3 parity subtests; chain-event async=true both surfaces, timeout 5 unchanged) |
| AC-HSA-004 | PASS | `TestHookSyncAsyncRegistrationFlags` (M4 commit run) | `ok ... internal/template 0.361s` (PostToolUseFailure row async=true both surfaces; per-entry grep = 1/1) |
| AC-HSA-005 | PASS | `go test ./internal/hook/ -run 'TestPostToolUseFailureHandler_AdditionalContextNotSystemMessage' -count=1 -v` | `--- PASS: TestPostToolUseFailureHandler_AdditionalContextNotSystemMessage (0.00s)` (additionalContext present with category prefix, systemMessage absent, hookEventName PostToolUseFailure) |

Async-count ledger (AC-HSA-009 progression): baseline 11/11 (E-L09) -> 12/12 at the M4
commit (12 = `git show a29a17cbeb:.claude/settings.json | grep -c '"async": true'`; the
same form on the template also prints 12) -> **14/14** at the M5 tip (`grep -c
'"async": true'` = 14 live, 14 template) — exactly baseline + 3, no keep-sync
registration touched.

Companion evidence: `go build ./...` exit 0; `GOOS=windows GOARCH=amd64 go build ./...`
exit 0; `go vet ./internal/hook/... ./internal/template/... ./internal/cli/...` exit 0;
`gofmt -l` empty on touched dirs; `golangci-lint run --timeout=2m` = the same 2
pre-existing staticcheck findings in internal/cli/paths.go (ST1008, file untouched by
this SPEC), 0 new; full `go test ./internal/hook/ -count=1` = `ok ... 282.669s`;
`go test ./internal/template/ -run 'HookSyncAsync|HookEntries|SettingsTemplate|Registration|ReviewGate'
-count=1` = `ok ... 0.320s`; boundary grep over every touched Go file = 0 matches; the
hooks-system.md PostToolUseFailure stdout row edit is byte-identical across the live and
template mirrors (`cmp` exit 0).

### M7 — verification close (2026-10-21/22, commit 703654daca + the ledger evidence commit)

M6 re-pin: AC-HSA-010/011/012 re-observed on the landing tree (703654daca). Base archaeology —
the M6 content rode the inherited base: `git show 8249bff8c3:...` already carries both UPS
entries at `timeout: 30`, the corrected row-369 doctrine wording, and async count 11 (all
observed this run), so the only settings deltas this SPEC adds are the three async flag pairs.

Deferred doctrine touch-up landed as commit 703654daca: the Timeout Configuration table's
"All other registered events" parenthetical now names all four async entry kinds (PostToolUse
harness-observe, PostToolUseFailure, ConfigChange, SubagentStop chain-event); live + template
pair, `cmp` byte-parity exit 0, TestRuleTemplateMirrorDrift green (0.499s).

AC-HSA-013 finalization: the M1 table (gitignored path, cited not committed) carries 29 rows x
exactly 10 per-run timings (awk count: 29 rows of 10) and gained the fixed prescription-order
statement (fast-path → timeout raise → async) in an M7 finalization section.

AC-HSA-014 observation (regression-guard, not a release-blocking pass): this tree's hook trace
log shows 119 UserPromptSubmit executions, max duration 281ms vs the 30s registration, 0
timeout/error records; coverage boundary (this worktree's sessions only) recorded as a Gap in
the ledger green cell.

| AC | Status | Verification command | Actual output (this run, tree 703654daca) |
|----|--------|----------------------|--------------------------------------------|
| AC-HSA-009 | PASS | `grep -c "async.: true"` both surfaces + `git diff 8249bff8c3..HEAD -- .claude/settings.json` | 14/14 (live/template) = baseline 11 + 3 exactly; diff = 3 hunks, 6+/3-, every added line an async pair — 0 keep-sync changes (ledger E-L09 GREEN) |
| AC-HSA-010 | PASS | `sed -n '369p'` + `'373,375p'` live doctrine | corrected wording present, ceiling bullet restated as default (E-L10 GREEN) |
| AC-HSA-011 | PASS | same on template mirror | byte-identical wording; `cmp` live↔template exit 0 (E-L11 GREEN) |
| AC-HSA-012 | PASS | `awk NR==268 \|\| NR==274` both surfaces | `timeout: 30` both UPS entries, both surfaces (E-L12 GREEN) |
| AC-HSA-013 | PASS | `ls` + awk run-count + grep prescription over the gitignored table | 29 rows x 10 runs; prescription-order statement present once (E-L13 GREEN) |
| AC-HSA-014 | OBSERVED (regression-guard) | grep over `.moai/logs/trace-*.jsonl` | 119 executions, max 281ms, 0 timeout records; coverage Gap recorded (E-L14 GREEN) |

Ledger sweep: E-L01..E-L16 + E-L18 GREEN cells closed in acceptance.md, each citing command +
verbatim output + exit code + tree 703654daca. All AC-HSA-001..017 green; AC-HSA-014 per its
regression-guard closure rule (§D.4). The ledger RED cells are untouched.

## §E.3 Run-phase Audit-Ready Signal

run_complete_at: 2026-10-22T00:20:00+09:00
run_commit_sha: 703654daca (M7 observation tree; the ledger evidence commit follows this signal)
run_status: complete
ac_pass_count: 17 (AC-HSA-001..017; AC-HSA-014 per its regression-guard closure rule)
ac_fail_count: 0
preserve_list_post_run_count: 0 (keep-sync table untouched — E-L09 diff evidence; no PRESERVE violations)
l44_pre_commit_fetch: not-applicable (card worktree; branch-local work, no push in this phase)
l44_post_push_fetch: not-applicable (no push — dispatch constraint; landing is the leader's batch flow)
new_warnings_or_lints_introduced: 0 on SPEC-touched files (golangci-lint = 2 staticcheck findings, both on files absent from the run-phase set: internal/cli/paths.go ST1008 pre-existing per M3/M4+M5 baselines; internal/cli/factory_card.go QF1001 first observed this M7 run, file byte-unchanged since f04b507365 — see acceptance.md Gaps)
cross_platform_build.darwin: pass (`go build ./...` exit 0, this run)
cross_platform_build.windows: pass (`GOOS=windows GOARCH=amd64 go build ./...` exit 0, this run)
total_run_phase_files: 16 (git diff --name-only 410791e91c..HEAD — the absorb merge to the M7 tip)
m1_to_mN_commit_strategy: per-milestone commits — M3 (772b9a1590 feat + 3483d65726 test + 7addd36d15 evidence), M4 (a29a17cbeb feat + 603ad65bb9 fix), M5 (4a0fa991ed feat + f04b507365 evidence), M7 (703654daca docs + ledger evidence commit)

## §E.4 Sync-phase Audit-Ready Signal

```yaml
sync_phase: complete
sync_complete_at: 2026-10-22T00:00:00+09:00
sync_commit_sha: 8b48246f072c482ab7e7dac2f1f13247ebe861dc
sync_status: complete
ac_source: .moai/specs/SPEC-HOOK-SYNC-ASYNC-001/acceptance.md (tier: M)
b12_self_test_a: pass (grep -c 'SPEC-HOOK-SYNC-ASYNC-001' CHANGELOG.md = 0 before emission — no duplicate entry)
b12_self_test_b: pass (MOAI-AC-COUNTER on acceptance.md: live=17 excluded=0 ambiguous=0 — matches the 17 ACs, AC-HSA-014 per its regression-guard closure rule)
b12_self_test_c: pass (all paths named in the CHANGELOG entry verified against the tree: internal/hook/post_tool_failure.go, .claude/hooks/moai/handle-codex-review-entry.sh, internal/template/templates/.claude/hooks/moai/handle-codex-review-entry.sh (byte-identical to the wrapper), .claude/settings.json and internal/template/templates/.claude/settings.json.tmpl (14 async flags each), both hooks-system.md surfaces)
changelog_entry_position: [Unreleased] > Changed (last entry of the subsection)
frontmatter_status_transitions.spec_md: in-progress → implemented → completed (merged, rides the single sync commit)
frontmatter_status_transitions.plan_acceptance: stateless — no status field; updated: 2026-10-22 refresh only
canary_compliance_check.mx_tags: sync sub-step — MX Tag validation folded into this close per the 3-phase close doctrine (no separate Mx-phase commit)
```

> Backfill note: `sync_commit_sha` carried the canonical `pending-backfill-sync` placeholder in
> the sync commit itself (a commit cannot cite its own hash); the real SHA `8b48246f07` was
> backfilled in the follow-up lane commit (D3 placeholder completion, spec-frontmatter-schema.md
> § SHA placeholder backfill exemption).

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
