# Progress — SPEC-V3R6-HOOK-INPUT-SCHEMA-001

> Tier S. Run-phase implementation record (TDD, RED-GREEN-REFACTOR). cycle_type=tdd.

## §A. Run-phase Summary

Two confirmed `internal/hook` stdin-parse defects fixed, both verified against the live `~/.moai/logs/hook-stderr.log` evidence and resolved through the real CLI binary:

1. **Defect 1 — `globs` type mismatch** (REQ-HIS-001): `HookInput.Globs` changed `string` → `[]string` (types.go:255). Claude Code v2.1.69+ emits `globs` as a JSON array; the prior `string` typing failed every `instructions-loaded` load with `cannot unmarshal array into Go struct field HookInput.globs of type string`.
2. **Defect 2 — empty-stdin robustness gap** (REQ-HIS-002): `ReadInput` (protocol.go) now treats empty / blank / whitespace-only stdin as a graceful no-op success — returns a default `*HookInput` (validateInput defaults applied) + nil error instead of `ErrHookInvalidInput`. Resolves the user-visible `PostToolUseFailure: Bash UnknownFailure` on PreToolUse(Bash) with empty stdin.

Mode Selection: Mode 5 (sub-agent), single manager-develop delegation, Tier S minimal scope. cycle_type=tdd.

## §B. D2 Gate Determination (empty-vs-truncated)

The plan §E Risk D2 row required confirming whether the live `pre-tool` failure payload was strictly-empty (`""`/whitespace) or truncated-non-empty (partial JSON), since the Go error `unexpected end of JSON input` is identical for both.

**Determination: empty/blank stdin (NOT truncated-non-empty).** Evidence:

1. **Wrapper script analysis** — `.claude/hooks/moai/handle-pre-tool.sh` pipes `head -c 65536 | moai hook pre-tool`. `head -c 65536` truncates only when input EXCEEDS 64KB. A PreToolUse(Bash) payload (a small command string) never approaches 64KB, so truncation-by-head is not a realistic cause.
2. **Failure-chain trace** — for empty stdin, `io.ReadAll` returns `[]byte{}`, then `normalizeHookInput` calls `json.Unmarshal([]byte{}, ...)` which returns exactly `unexpected end of JSON input`. The empty case reproduces the observed error string precisely.
3. **No truncated-non-empty evidence** — the log records only the error string, not the payload; no partial-JSON artifact was found. The empty-stdin no-op is a strict improvement that resolves the documented symptom.

Per plan §E D2 guidance, no BLOCKER is warranted — the empty/blank/whitespace no-op scope (AC-2) was implemented as scoped. **Boundary note**: non-empty malformed JSON STILL returns `ErrHookInvalidInput` (intentionally NOT broadened); if a future truncated-non-empty payload is ever confirmed, that is a separate scope decision.

## §E.2 Run-phase Evidence

| Item | Status | Verification Command | Actual Output |
|------|--------|----------------------|---------------|
| AC-1 (REQ-HIS-001) — globs array decodes | PASS | `go test ./internal/hook/ -run 'TestReadInput/instructions-loaded'` | `--- PASS: TestReadInput/instructions-loaded_globs_array_decodes` — Globs populated `[**/*.go **/*.md]` |
| AC-1 end-to-end (CLI) | PASS | `echo '{...,"globs":["**/*.go","**/*.md"]}' \| go run ./cmd/moai hook instructions-loaded; echo $?` | `{}` then `exit=0` (was: `cannot unmarshal array ... exit non-zero`) |
| AC-2 (REQ-HIS-002) — empty stdin no-op | PASS | `go test ./internal/hook/ -run 'TestReadInput/empty_stdin'` | `--- PASS: TestReadInput/empty_stdin` — non-nil `*HookInput`, HookEventName=`unknown`, SessionID=`unknown`, nil err |
| AC-2 — whitespace-only stdin no-op | PASS | `go test ./internal/hook/ -run 'TestReadInput/whitespace_only_stdin'` | `--- PASS: TestReadInput/whitespace_only_stdin` (input `"   \n"`) |
| AC-3 (REQ-HIS-003, regression) — existing tests green | PASS | `go test ./internal/hook/...` | all packages `ok` (incl. the 2 INTENTIONALLY inverted cases; `malformed_JSON` still errors — scope not broadened) |
| AC-4 (REQ-HIS-002, mandatory smoke) | PASS | `go run ./cmd/moai hook pre-tool </dev/null; echo "exit=$?"` | `{"hookSpecificOutput":{"hookEventName":"PreToolUse","permissionDecision":"allow"}}` then `exit=0` (was: `unexpected end of JSON input`, non-zero) |
| RED evidence (AC-1) | CAPTURED | `go test ./internal/hook/ -run TestReadInput` (pre-M2) | `invalid operation: got.Globs[0] != "**/*.go" (mismatched types byte and untyped string)` — compile-level RED from `Globs string` |
| RED evidence (AC-2) | CAPTURED | `go test ./internal/hook/ -run TestReadInput` (pre-M3, post-M2) | `FAIL: empty_stdin / whitespace_only_stdin — unexpected error: hook: invalid JSON input: unexpected end of JSON input` |
| Cross-platform build | PASS | `go build ./... && GOOS=windows GOARCH=amd64 go build ./internal/hook/...` | both exit 0 |
| go vet | PASS | `go vet ./internal/hook/...` | exit 0 |
| golangci-lint | PASS | `golangci-lint run ./internal/hook/... --timeout=2m` | `0 issues.` |
| Coverage (internal/hook) | PASS | `go test -cover ./internal/hook/` | `coverage: 81.8% of statements` (net non-decreasing; test added before fix) |
| Full cascade | PASS (1 pre-existing unrelated) | `go test ./...` | only `internal/template TestOutputStylesTemplateLiveParity` (einstein.md template/live drift) fails — PRE-EXISTING, confirmed identical with hook changes stashed; out-of-scope, owned by parallel session |

## §E.3 Run-phase Audit-Ready Signal

```yaml
run_complete_at: 2026-06-03
run_commit_sha: "27745a7fa"   # cherry-picked from worktree 0258f4f6c onto docs/glm-webtool-routing-m1-m5
run_status: implemented
ac_pass_count: 4          # AC-1, AC-2, AC-3, AC-4 — all 4 mandatory AC PASS
ac_fail_count: 0
preserve_list_post_run_count: 0   # no PRESERVE-list files modified outside scope
l44_pre_commit_fetch: n/a-worktree-isolated   # orchestrator owns pre-spawn fetch + push
l44_post_push_fetch: n/a-worktree-isolated
new_warnings_or_lints_introduced: 0
cross_platform_build:
  native: pass
  windows_amd64: pass
total_run_phase_files: 4   # types.go + protocol.go + protocol_test.go + progress.md
m1_to_mN_commit_strategy: single-commit   # Tier S, M1-M5 in one atomic commit
```

## §F. Files Modified (scope envelope)

- `internal/hook/types.go` — Globs `string` → `[]string` (1 field line) + gofmt struct-tag realignment
- `internal/hook/protocol.go` — empty/blank/whitespace short-circuit in `ReadInput` + `bytes` import
- `internal/hook/protocol_test.go` — added array-form globs decode case (AC-1) + inverted 2 empty/whitespace cases (AC-2)
- `.moai/specs/SPEC-V3R6-HOOK-INPUT-SCHEMA-001/progress.md` — this file

## §E.4 Sync-phase Audit-Ready Signal

```yaml
sync_complete_at: 2026-06-03
sync_commit_sha: "c255d31a0"
sync_status: implemented
spec_frontmatter_transitions:
  - field: status
    old_value: in-progress
    new_value: implemented
changelog_entry_position: "[Unreleased]/Fixed section"
ac_final: "4/4 mandatory AC PASS (AC-1 globs array, AC-2 empty-stdin no-op + 2 inverted cases, AC-3 regression, AC-4 pre-tool smoke exit-0)"
fix_outcome: "both internal/hook defects resolved — instructions-loaded array globs decodes; pre-tool empty stdin → exit 0 (PostToolUseFailure: Bash symptom resolved)"
d2_determination: "empty/blank stdin (NOT truncated) — handle-pre-tool.sh head -c 65536 truncates only >64KB; small Bash payloads never reach that"
sync_method: orchestrator-direct (Tier S, active parallel-session race — worktree/cherry-pick overhead avoided)
```

## §G. Status Transition Note (orchestrator action)

This is the M1 commit (`draft → in-progress`). The spec.md `status:` frontmatter transition `draft → in-progress` and `updated:` refresh MUST be applied by the orchestrator when reconciling, because `spec.md` is uncommitted in the shared checkout and absent from this isolated worktree (created from base `ad974fe5b` which predates the SPEC). The implementation source files + this progress.md are committed in the worktree for cherry-pick.

## §E.5 Mx-phase Audit-Ready Signal

```yaml
mx_complete_at: 2026-06-03
mx_commit_sha: "1c9c0a790"
mx_status: completed
spec_frontmatter_transitions:
  - field: status
    old_value: implemented
    new_value: completed
  - field: version
    old_value: "0.1.0"
    new_value: "0.2.0"
four_phase_close:
  plan_artifacts: "authored by manager-spec (uncommitted in shared checkout), committed in reconcile 00f6c8c91"
  run: "27745a7fa (cherry-pick of worktree 0258f4f6c)"
  reconcile: 00f6c8c91
  sync: c255d31a0
  mx: "1c9c0a790"
ac_final: "4/4 mandatory AC PASS"
close_subject_doctrine_dogfood: "this close commit uses the full SPEC-ID per REQ-DLC-011"
einstein_md_failure: "pre-existing uncommitted template drift (out of scope; not introduced by this SPEC)"
```
