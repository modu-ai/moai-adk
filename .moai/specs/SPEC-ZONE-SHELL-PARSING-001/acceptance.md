---
id: SPEC-ZONE-SHELL-PARSING-001
title: "Acceptance criteria — evidence ledger, AC matrix, scenarios, gates"
version: "0.5.0"
created: 2026-10-07
updated: 2026-10-07
author: manager-spec
tier: M
---

# SPEC-ZONE-SHELL-PARSING-001 — acceptance

## §A Method

Every criterion adopts TWO cells (verification-completeness §2): a RED-now cell observed on the pre-implementation tree and pinned to it — or, for preservation criteria (AC-ZSP-005/006), a preserved-baseline cell recording the green state that must not regress — and a green-path cell naming the milestone that flips (or keeps) it. All RED-now observations below were measured at tree `b9ef003808da2ac1dfe42e5374d5bde4302f46b3` (branch `WT-shell-parsing-guard`), 2026-10-07, in this run. The RED is red for the stated reason: the guard's current parsing admits (or crashes on) the exercised input, and M1/M2's fixes are the only changes this SPEC can make that flip it.

## §B Evidence Ledger

Lint note: `moai spec lint` on this SPEC reports 0 errors and 7 `VacuousTestAssertion` warnings proposing `^…$`-anchored `-run` selectors. The ledger keeps the selectors exactly as executed (the verbatim outputs are attributed to THESE commands; re-anchoring would decouple command from output). Card t1421 additionally recorded a case where the wrapped-anchor suggestion empties the selection. The countable-swept-set obligations in AC-ZSP-004 (per-cell `--- PASS` lines, `[no tests to run]` prohibition) carry the assertion strength the lint asks for.


### EV-1 — three allow-class subtests RED

- Command: `go test ./internal/hook/ -run 'TestProtectedZoneShellParsingBypasses/(dash_dash_operand|conditional_function_declaration|command_env_prefix)' -count=1 -v -timeout 2m`
- Verbatim stdout (this run):
```
=== RUN   TestProtectedZoneShellParsingBypasses
=== RUN   TestProtectedZoneShellParsingBypasses/dash_dash_operand
    protected_zone_shell_parsing_test.go:34: rm -- -zone/secret.md: decision="allow" reason="", want deny
=== RUN   TestProtectedZoneShellParsingBypasses/conditional_function_declaration
    protected_zone_shell_parsing_test.go:49: false && rm() { :; }; rm zone_dir/secret.md: decision="allow" reason="", want deny
=== RUN   TestProtectedZoneShellParsingBypasses/command_env_prefix
    protected_zone_shell_parsing_test.go:66: command rm zone_dir/a.log: decision="allow" reason="", want deny
    protected_zone_shell_parsing_test.go:66: env rm zone_dir/a.log: decision="allow" reason="", want deny
    protected_zone_shell_parsing_test.go:66: command -p rm zone_dir/a.log: decision="allow" reason="", want deny
--- FAIL: TestProtectedZoneShellParsingBypasses (0.03s)
    --- FAIL: TestProtectedZoneShellParsingBypasses/dash_dash_operand (0.01s)
    --- FAIL: TestProtectedZoneShellParsingBypasses/conditional_function_declaration (0.01s)
    --- FAIL: TestProtectedZoneShellParsingBypasses/command_env_prefix (0.01s)
FAIL
FAIL	github.com/modu-ai/moai-adk/internal/hook	0.894s
FAIL
```
- Exit code: 1
- Tree: `b9ef003808da2ac1dfe42e5374d5bde4302f46b3` (corroborates the lane-1 record in `.moai/reports/t1574/red-reproduction.md`)

### EV-2 — recursion subtest crashes the test binary

- Command: `go test ./internal/hook/ -run 'TestProtectedZoneShellParsingBypasses/recursion_state_reset' -count=1 -timeout 3m`
- Verbatim stdout (head, this run):
```
runtime: goroutine stack exceeds 1000000000-byte limit
runtime: sp=0x4750721014d0 stack=[0x475072100000, 0x475092100000]
fatal error: stack overflow

runtime stack:
runtime.throw({0x100c1f5ec?, 0x1001e70c4?})
```
- Tail (same run): `FAIL	github.com/modu-ai/moai-adk/internal/hook	0.733s`
- Exit code: 1
- Tree: `b9ef003808da2ac1dfe42e5374d5bde4302f46b3` (corroborates lane-1's independent observation, 2.536s, same fatal line)

### EV-3 — executable path absorbed as a declared function (plan-phase probe)

- Command: an in-package probe test (`internal/hook/zz_probe_pathfunc_test.go`, DELETED after observation) running the handler with command `rm() { :; }; /bin/rm zone_dir/secret.md`; the probe call mirrors the regression helpers (`newZoneRoot` + `zoneShippedDoc("  probe_zone:\n    paths: [\"zone_dir/\"]\n")` + `zoneTestHandler` + `zoneCall`).
- Verbatim stdout:
```
=== RUN   TestZZProbePathFuncAbsorption
    zz_probe_pathfunc_test.go:20: decision="allow" reason=""
--- PASS: TestZZProbePathFuncAbsorption (0.00s)
PASS
ok  	github.com/modu-ai/moai-adk/internal/hook	0.649s
```
- Exit code: 0
- Tree: `b9ef003808da2ac1dfe42e5374d5bde4302f46b3`
- Reading: real bash dispatches functions by bare name only — `/bin/rm` runs the external binary and the protected file is deleted for real; the guard answers `allow`. Independently listed by the merge-gate ledger round 1 (`.moai/reports/t1574/red-reproduction.md` § 게이트 원장).

### EV-4 — sweep selector matches nothing today

- Command: `go test ./internal/hook/ -run 'TestProtectedZoneShellParsingMatrix' -count=1 -timeout 2m`
- Verbatim stdout:
```
ok  	github.com/modu-ai/moai-adk/internal/hook	0.869s [no tests to run]
```
- Exit code: 0
- Tree: `b9ef003808da2ac1dfe42e5374d5bde4302f46b3`
- Reading: exit 0 with an EMPTY swept set — the pass asserts nothing (verification-completeness §1.1); the criterion is red because no matrix exists. M2 creates it and makes an empty cell list fail loudly.

### EV-5 — preserved-behavior and quality baselines (same tree, same run)

| Check | Command | Verbatim stdout | Exit |
|-------|---------|-----------------|------|
| Existing family | `go test ./internal/hook/ -run 'TestProtectedZone$' -count=1 -timeout 10m` | `ok  	github.com/modu-ai/moai-adk/internal/hook	4.515s` | 0 |
| go vet | `go vet ./internal/hook/` | (no output) | 0 |
| golangci-lint | `/Users/goos/go/bin/golangci-lint run ./internal/hook/...` | `0 issues.` | 0 |

## §C AC Matrix

### AC-ZSP-001 — the three allow-class subtests GREEN

- RED-now: EV-1 (three protected deletions observed `decision="allow"`, want deny).
- Green path: M1 fixes K1/K2/K3; the same command exits 0 with three `--- PASS` subtests.
- Given the zone fixture `probe_zone: paths: ["-zone/"]` (dash-dash case) or `["zone_dir/"]` (other cases) and the `harness-learner` identity,
  when the guard receives each RED command of EV-1,
  then every subtest answers `deny` with the protected-zone category, and `go test -run 'TestProtectedZoneShellParsingBypasses/(dash_dash_operand|conditional_function_declaration|command_env_prefix)'` exits 0.

### AC-ZSP-002 — recursion terminates; stack overflow zero-recurrence

- RED-now: EV-2 (`fatal error: stack overflow` kills the test binary).
- Green path: M1 fix K5 (restore-on-exit plus the global visit budget of audit D2); the subtest terminates and the command is judged. The frozen assertion accepts either verdict and stays untouched — the DENY VERDICT is owned observably by AC-ZSP-004's unbounded recursion cells (audit D3): every unbounded cell in the matrix asserts the fail-closed deny, so an unbounded→allow mutant fails the matrix even though the frozen subtest alone would pass it.
- Given the zone fixture `zone_dir/` and the command `f() { if false; then f; f; fi; }; f; rm zone_dir/a.log`,
  when the guard walks it,
  then the walker terminates within the test timeout, no `fatal error` appears in the output, zero-recurrence holds (the full hook package run of AC-ZSP-006 contains no stack-overflow crash), and the deny verdict for the unbounded case is observable through AC-ZSP-004's matrix cells.

### AC-ZSP-003 — executable-path subtest observed RED, then GREEN

- RED-now: EV-3 (probe observed `decision="allow"` for a real-bash deletion). Plan-audit note (O2): the probe file was deleted after observation, so this criterion is classified regression-guard pending the M1 RED-first observation its green path mandates; the underlying defect was confirmed by code read at `protected_zone_shell.go:426-432`. The M1 subtest must be observed RED before the fix lands.
- Green path: M1 step 2 authors the subtest RED-first, then fix K4 flips it (audit D8a corrected the stale step reference).
- Given a declared `rm` function and the command `rm() { :; }; /bin/rm zone_dir/secret.md`,
  when the guard resolves the call,
  then the path word never resolves to the declared function, the verb judgment runs on the base name, and the protected deletion is denied; a path whose base matches no verb (`/bin/ls …`) stays allow.

### AC-ZSP-004 — parsing-matrix sweep exists and is green

- RED-now: EV-4 (empty swept set).
- Green path: M2 creates the matrix per REQ-ZSP-007; the sweep run exits 0 with a NON-EMPTY cell count visible in output (per-subtest `--- PASS` lines), and an empty cell list fails the runner by construction.
- Given the matrix test of M2,
  when `go test ./internal/hook/ -run TestProtectedZoneShellParsingMatrix -count=1 -v -timeout 10m` runs,
  then it exits 0, the output shows every cell's subtest line (a countable swept set), no `[no tests to run]` token appears, and every unbounded cell's run output carries its deny verdict — the recursion cells (audit D3, the deny ownership AC-ZSP-002 points at), the audit-D5 seed cell, the audit-D6 cd-chain cell, and the audit-D9 absent-manifest cell (a budget-exhausting shape followed by `rm .claude/hooks/a.sh` under an absent manifest, asserting deny) included.

### AC-ZSP-005 — existing behavior passes unchanged

- RED-now / baseline: EV-5 (family ok 4.515s at base).
- Green path: M1–M3 re-run; the family's source assertions are never edited.
- Given the base tree passed `TestProtectedZone$` and the four pre-existing subtests' assertion lines,
  when the fixes land,
  then `go test ./internal/hook/ -run 'TestProtectedZone$' -count=1` exits 0, `git diff` on the four pre-existing subtests shows no assertion change (D4), and the thirteen-mutation-form verdicts are untouched (REQ-ZSP-008).

### AC-ZSP-006 — scoped verification batch green

- RED-now / baseline: EV-5 (vet exit 0, lint 0 issues).
- Green path: M3, under a slot lease for the suite.
- Given the changes are scoped to `internal/hook`,
  when `go test ./internal/hook/ -timeout 30m`, `go vet ./internal/hook/`, and `golangci-lint run ./internal/hook/...` run at close,
  then all three exit 0 and the outputs are carried verbatim into progress.md §E.2.

## §D Edge Cases

- `rm -f -- -zone/secret.md` (options before the separator) — after `--`, the hyphen-leading word is a candidate.
- `cp --target-directory=zone_dir src -- -zone/x` — attached long-option values AND post-`--` operands both collected.
- `env FOO=bar rm zone_dir/x` — assignment consumed, verb reached; bare `env X=1` under-matches as today.
- `env command rm zone_dir/x` — nested wrappers strip within the bounded loop.
- `command ls`, `nohup cat zone_file` (non-mutating verbs under wrappers) — stay allow.
- Mutual recursion `a() { b; }; b() { a; }; a; rm zone_dir/x` — per-name bounds hold, terminates, fail-closed.
- `for i in 1 2; do f; done` with recursive `f` — loop fixed point and recursion bound compose; terminates.
- `function rm { :; }` keyword form then `rm zone_dir/x` — certain shadow stays allow (real bash would not delete); the conditional variant denies via the real-verb interpretation.
- `rm -- '‑zone/x'` quoted forms — literal quoted hyphen-leading operands collected; dynamic words still under-match.

## §E Quality Gates

- TRUST 5 mapping — Tested: AC-001..006; Readable/Unified: comments English matching file density, gofmt/golangci-lint clean; Secured: fail-closed direction (spec §C); Trackable: Conventional commits carrying the card id.
- Gate commands: AC-ZSP-006's three, plus `git status --short` proving scope.

## §F Definition of Done

All six ACs GREEN with evidence in progress.md §E.2/§E.3; scope confined to the two (or three) in-scope files; no `[no tests to run]`, no `fatal error` anywhere in the final batch output; parent SPEC REQ-SIPZ-007 behavior preserved (AC-005).

## §G Traceability

| REQ | AC |
|-----|----|
| REQ-ZSP-001 | AC-ZSP-001, AC-ZSP-004 |
| REQ-ZSP-002 | AC-ZSP-001, AC-ZSP-004 |
| REQ-ZSP-003 | AC-ZSP-001, AC-ZSP-004 |
| REQ-ZSP-004 | AC-ZSP-003, AC-ZSP-004 |
| REQ-ZSP-005 | AC-ZSP-002 |
| REQ-ZSP-006 | AC-ZSP-002, AC-ZSP-004 |
| REQ-ZSP-007 | AC-ZSP-004 |
| REQ-ZSP-008 | AC-ZSP-005 |
| (quality gate — no behavior REQ; preserved baseline EV-5) | AC-ZSP-006 |
