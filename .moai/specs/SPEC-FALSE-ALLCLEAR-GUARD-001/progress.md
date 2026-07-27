# SPEC-FALSE-ALLCLEAR-GUARD-001 — Progress

## §E.1 Plan-phase Audit-Ready Signal

```yaml
plan_status: audit-ready
plan_complete_at: 2026-07-27
tier: M
artifacts: [spec.md, plan.md, acceptance.md]
milestones: [M1 slog scoping, M2 ast-grep sentinel + gate + doctor + docs]
```

## §E.2 Run-phase Evidence

M1 commit: `a5c735aed` — `fix(SPEC-FALSE-ALLCLEAR-GUARD-001): M1 scope slog suppression to the moai hook path`
M2 commit: `49994712b` — `feat(SPEC-FALSE-ALLCLEAR-GUARD-001): M2 report an unavailable ast-grep scanner instead of a clean scan`

Merge-base: `6763aff3b` (origin/main at plan-phase entry). `git diff <merge-base>..HEAD --stat` → 32 files, 1845 insertions, 109 deletions.

Verbatim per-AC command output (independently reproduced by sync-auditor; cited per the sync-audit report's Commands section):

```
### AC-FAG-001 — exactly one slog.SetDefault site
grep -rn 'slog.SetDefault' internal/ cmd/ --include='*.go' | grep -v _test → 1 (logging.go:73)

### AC-FAG-002 — level resolution
TestResolveLogLevel: list=1, subtest PASS=8
grep -rln 'config.EnvLogLevel' internal/cli/ → 2 files
grep -rn '"MOAI_LOG_LEVEL"' internal/cli/ | grep -v _test → 0
MOAI_LOG_LEVEL end-to-end: unset→4 WARN stderr; error→0 WARN (148 bytes M2 guidance only);
  info→3 INFO incl. "gopls bridge disabled (lsp.yaml: enabled=false)"; bogus-not-a-level→falls back to warn (4 WARN, 0 INFO)

### AC-FAG-003 — hook discards, non-hook does not
TestLoggingHandlerSelection: list=1, subtest PASS=7
moai hook pre-tool (MOAI_LOG_LEVEL=debug) → stderr 0 bytes (carve-out unconditional)

### AC-FAG-004 — falsification: scanner warning observable, revert removes it
env PATH=$EMPTY moai ast-grep ./internal/astgrep → exit=1, stderr 1050 bytes
  "ast-grep: scan did not run — the ast-grep (sg) CLI was not found. Install it from
   https://ast-grep.github.io/guide/quick-start.html, then re-run."
M1-only build (a5c735aed) reproduces pre-M2 shape; pre-change binary at merge-base emits 0 stderr bytes.

### AC-FAG-005 — no slog record reaches stdout
--format=text|json|sarif → stdout_bytes 0/0/0, level= 0/0/0, msg= 0/0/0 (sg absent)
Positive control (sg present, 38,004 stdout lines): grep -c 'level=' stdout = 0; grep -c 'msg=' stdout = 0
Constructive falsification (dest: os.Stdout) → grep -c 'level=' stdout: 0 → 4 (discriminates); TestLoggingHandlerSelection fails 2 subtests

### AC-FAG-006 — trivial fast path bypasses full init
moai --version → exit=0, stdout 1 line, stderr 0 bytes
TestTrivialPathSkipsFullInit: list=1, PASS=1

### AC-FAG-007 — sentinel returned, matchable, carries guidance
TestScannerScan_UnavailableSentinel: list=1, subtest PASS=3 (unavailable / non-sentinel / clean_scan — subtest 3 executed live, sg present)

### AC-FAG-008 — moai ast-grep exits non-zero with stderr guidance
env PATH=$EMPTY moai ast-grep ./internal/astgrep → exit=1; grep -ci install stderr = 2; grep -c 'no findings' stdout = 0
Positive control (sg present): moai ast-grep ./internal → exit=1 (19006 findings); grep -c 'scan did not run' stderr = 0

### AC-FAG-009 — message never reaches stdout under any format
text/json/sarif → stdout install-mentions 0/0/0

### AC-FAG-010 — ast-edit behavior unchanged
env PATH=$EMPTY moai ast-edit --pattern x --rewrite y --lang go ./internal/astgrep
  → exit=0, "ast-grep (sg) is not installed; nothing to apply."
git diff --stat 6763aff3b -- internal/cli/astedit.go | wc -l → 0

### AC-FAG-011 — skip reason survives all three frames
TestPreTool_AstGrepSkipReasonSurfaces: list=1, PASS=1
Form A (delete out.SystemMessage = gateNotice in pre_tool.go) → FAIL: TestPreTool_AstGrepSkipReasonSurfaces
Form B (restore gate.go discarding pass branch + return true, "") → FAIL: TestPreTool_AstGrepSkipReasonSurfaces
End-to-end on the wire:
  echo '{"session_id":"t","hook_event_name":"PreToolUse","tool_name":"Bash","tool_input":{"command":"git commit -m \"x\""}}' \
   | env PATH=$EMPTY CLAUDE_PROJECT_DIR=$PD moai hook pre-tool
  → exit=0, stderr 0 bytes, stdout:
  {"systemMessage":"ast-grep scan skipped: the sg CLI was not found, so no rules ran
    (install from https://ast-grep.github.io/guide/quick-start.html)",
   "hookSpecificOutput":{"hookEventName":"PreToolUse","permissionDecision":"allow"}}

### AC-FAG-012 — unavailable and other-error reasons distinct
TestAstGrepGateReasons: list=1, subtest PASS=2 (both reasons package-level named constants)

### AC-FAG-013 — gate never denies because sg is absent
-list 2, --- PASS 2, --- FAIL 0

### AC-FAG-014 — docs updated in all four locales
per-locale anchored diff: en→2 ja→2 ko→2 zh→2
grep -l 'guide/quick-start' docs-site/content/*/cli-reference/ast-grep.md | wc -l → 4
grep -c 'exits without an error' docs-site/content/en/cli-reference/ast-grep.md → 0

### AC-FAG-015 — moai doctor reports sg
JSON grep 0→2 (exactly 1 matching check row); grep -c 'exec.LookPath' internal/cli/doctor.go → 3

### AC-FAG-016 — sg check warns with guidance
sg absent: {"name":"ast-grep CLI","status":"warn","message":"sg not found — 'moai ast-grep' cannot scan
  and the commit gate skips its rules; install from https://ast-grep.github.io/guide/quick-start.html"}
sg present: {"name":"ast-grep CLI","status":"ok","message":"ast-grep 0.40.5"}

### AC-FAG-017 — full suite green, no new skips
go test -count=1 ./... → exit=0, 105 ok, 0 FAIL
Skip baseline (measured, not recorded pre-M1 as required — see F1/F2 below):
  non-verbose grep -c '^--- SKIP' → 0 (structurally pinned, cannot discriminate — F2 known debt)
  verbose '--- SKIP' baseline=56 (merge-base) HEAD=56; comm -13 / comm -23 named-skip diff → empty both directions (no new skip, none removed)

### AC-FAG-018 — vet, format, cross-platform build clean
go vet ./... → exit=0
GOOS=windows GOARCH=amd64 go build ./... → exit=0
GOOS=linux GOARCH=amd64 go build ./... → exit=0
gofmt -l <SPEC file list + new test files> → 0
git diff --name-only <merge-base>..HEAD -- '*.go' | xargs gofmt -l | wc -l → 0
```

Coverage (merge-base `6763aff3b` → HEAD):

| Package | Baseline | HEAD | Delta |
|---|---:|---:|---:|
| `internal/astgrep` | 88.7% | 89.0% | +0.3 |
| `internal/hook/quality` | 87.5% | 87.6% | +0.1 |
| `internal/hook` | 83.3% | 83.4% | +0.1 |
| `internal/cli` | 75.3% | 75.4% | +0.1 |
| `internal/cli/uikit` | 98.8% | 98.8% | 0.0 |

Golden files: only the sg row + column widening + `4 ok`→`5 ok` + `Pass 11`→`Pass 12`, no unrelated drift (`git diff <mb>..HEAD -- internal/cli/testdata/`).

## §E.3 Run-phase Audit-Ready Signal

```yaml
run_status: audit-ready
run_complete_at: 2026-07-27
run_commit_sha_m1: a5c735aed
run_commit_sha_m2: 49994712b
ac_pass_count: 18
ac_total: 18
suite_status: "go test -count=1 ./... exit=0, 105 ok, 0 FAIL"
cross_platform: "darwin (native), windows (cross-build exit 0), linux (cross-build exit 0)"
known_debt_at_run_close:
  - "AC-FAG-017 skip-count sub-assertion is structurally vacuous on non-verbose go test output (grep -c '^--- SKIP' pinned at 0); underlying REQ-FAG-031 verified directly instead (56=56 skip count, empty named-skip diff both directions)."
  - "AC-FAG-008 install sub-assertion passes with M1 alone (zero M2 change); other two sub-assertions (exit code, 'no findings' absence) discriminate correctly."
```

## §E.4 Sync-phase Audit-Ready Signal

```yaml
sync_status: audit-ready
sync_complete_at: 2026-07-27
sync_commit_sha: 4e3275d62
sync_audit_verdict: "CONDITIONAL PASS 89.9/100 (Functionality 90 must-pass PASS, Security 95 must-pass PASS, Craft 82 PASS-with-debt, Consistency 93 PASS)"
sync_audit_report: .moai/reports/sync-audit/SPEC-FALSE-ALLCLEAR-GUARD-001-2026-07-27.md
f1_resolution: "progress.md §E.2/§E.3 populated at sync-phase (this commit) — closes the sole reason the verdict was CONDITIONAL rather than PASS."
residual_debt_deferred_to_followup:
  - "AC-FAG-017 skip-count command re-anchor to 'go test -v' form + recorded baseline (56) — deferred, acceptance.md body not modified this sync per manager-docs' frontmatter-only mandate."
  - "AC-FAG-008 install sub-assertion re-anchor to 'scan did not run' literal — deferred, same reason."
  - "AC-FAG-015 prose mislabel of which command is the registration guard — deferred, same reason."
  - "Gap 3 (largest residual risk): whether Claude Code's PreToolUse allow path actually surfaces HookOutput.SystemMessage to the user was not verified against a live session — mechanism confirmed on the wire, user-visibility unconfirmed."
  - "deps.go:179 gopls-bridge-disabled severity asymmetry (Warn vs sibling Info at deps.go:131) — follow-up candidate, out of this SPEC's declared scope (§C)."
  - "MOAI_LOG_FORMAT remains unwired and is now visibly misleading rather than invisibly inert — explicitly out of scope per spec.md §C."
```
