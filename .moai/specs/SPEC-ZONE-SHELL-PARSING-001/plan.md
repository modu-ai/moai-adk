---
id: SPEC-ZONE-SHELL-PARSING-001
title: "Implementation plan — five shell-parsing bypass fixes, parsing-matrix sweep, scoped verification batch"
version: "0.5.0"
created: 2026-10-07
updated: 2026-10-07
author: manager-spec
tier: M
---

# SPEC-ZONE-SHELL-PARSING-001 — plan

## §A Context

Card t1574 (security P1×2+P3, Class C). The protected-zone shell guard walk (`internal/hook/protected_zone_shell.go`, mvdan/sh-based possible-worlds walker) admits four classes of protected deletion and crashes on a fifth. All five are RED at base `b9ef003808da2ac1dfe42e5374d5bde4302f46b3` — four via `TestProtectedZoneShellParsingBypasses` (assertions frozen), the fifth via a plan-phase probe and the merge-gate ledger round 1. Development mode: TDD; the RED-first obligation is already half discharged (4 subtests), M1 discharges the rest.

## §B Known Issues

| # | Site | Class | RED evidence |
|---|------|-------|--------------|
| K1 | `protected_zone_shell.go:611-630` (`zonePathCandidates`, the `-`-prefix skip at :624) | `--` operands dropped | EV-1 subtest `dash_dash_operand` |
| K2 | `protected_zone_shell.go:432` (`zoneCall` funcs lookup) + join unions (:845, :883, :894, :914, :963) | conditional declaration absorbs the real verb | EV-1 subtest `conditional_function_declaration` |
| K3 | `protected_zone_shell.go:598` (verb check on the raw head) | static wrapper prefixes not stripped | EV-1 subtest `command_env_prefix` |
| K4 | `protected_zone_shell.go:426-431` (path fold at :430 BEFORE the funcs lookup) | executable path absorbed as function | EV-3 probe + gate ledger round 1 |
| K5 | `protected_zone_shell.go:468` (`delete(w.calling, name)` on frame exit) | recursion state reset → unbounded walk → stack overflow | EV-2 subtest `recursion_state_reset` |
| K7 | `protected_zone_shell.go:997` vs :1031 (verdict order) | the unbounded denial is evaluated AFTER the `!w.mutating` early return — a budget exhausted mid-walk aborts before the mutating statement is reached and the walk answers allow (audit D5) | review finding; regression cell = the seed command in the M2 matrix |
| K8 | the cd branch (:471-489) | every cd unions the pre-cd set into the next (`next = append(next, w.cwds...)`), squaring the directory set toward 2^n states (audit D6) | review finding; regression cell = the cd-chain cell in the M2 matrix |

## §C Pre-flight

1. `git rev-parse --short HEAD` → `b9ef00380`; branch `WT-shell-parsing-guard`; only the untracked RED test file is in the tree.
2. Baselines observed at this tree (acceptance.md §B): family `TestProtectedZone` ok 4.515s; `go vet ./internal/hook/` exit 0; `golangci-lint run ./internal/hook/...` 0 issues; sweep selector `[no tests to run]`.
3. Slot lease before the minutes-long hook package suite (M3): `moai slot acquire --resource <shared-type> --max-duration <bound>` → run → `moai slot release` (lane-local verification rules).

## §D Constraints

- Files: `protected_zone_shell.go` + `protected_zone_shell_parsing_test.go` (+ optional `protected_zone_shell_matrix_test.go`). Nothing else.
- The four RED subtests' assertions are frozen (D4 of spec §C); :430's subtest lands RED-first.
- Fail-closed direction inside the new classes; pre-existing under-match classes unchanged.
- Code comments English, matching the file's density (the file documents every fix with a round-style comment).

## §E Self-Verification

Run-phase closes with verbatim outputs for: E1 the AC matrix (acceptance.md §C) each row GREEN; E2 the hook package suite `go test ./internal/hook/ -timeout 30m` under a slot lease; E3 `go vet ./internal/hook/`; E4 `golangci-lint run ./internal/hook/...`; E5 `git status --short` showing only the in-scope files; E6 the frozen-subtest assertion text unchanged (`git diff` on the four pre-existing subtests shows no assertion edit); E7 progress.md §E.2 evidence written before the completion report.

## §F Milestones

Ordered by decision-reversibility: the world-semantics fix (highest design uncertainty, most likely to be revised in review) leads; mechanical fixes follow.

### M1 (High) — five interpretation fixes (steps 1–5, RED-first) plus two audit repairs (steps 6–7: K7 verdict order, K8 cwds budget)

1. **K2 first (design-heavy).** Give the function registry a per-name "conditional" dimension: a straight-line `FuncDecl` (:918-925) sets it false (replaces, today's semantics); every branch-join union (:845 `&&`/`||`, :883 for-loop, :894 while, :914 case, :963 `walkIfChain`) marks a name PRESENT only in the post-world (absent in the pre-branch entry) as conditional. `zoneCall` (:432): a declared-and-certain name walks its bodies and returns (unchanged); a declared-and-conditional name walks the bodies AND falls through to the real-verb resolution — the joined world keeps the possibility of absence. `cloneZoneFuncs`/`mergeZoneFuncs`/`zoneFuncsEqual` must carry the flag so the loop fixed point (:713) still converges on it.
2. **K4 (RED-first, per audit O1 the gate lives in this step).** FIRST extend `TestProtectedZoneShellParsingBypasses` with the `executable_path_function_absorption` subtest (`rm() { :; }; /bin/rm zone_dir/secret.md` → wantZoneDeny) and observe it RED (EV-3 predicts `decision="allow"`); THEN fix: look the funcs table up with the ORIGINAL command word and fold `path.Base` only for verb resolution (move the fold below the funcs branch, or keep an `origName`). Negative control: a path whose base matches no verb (`/bin/ls`) stays allow.
3. **K3 (audit D1).** In `zoneCall`, after the funcs check on the RAW head, strip the static wrapper set {`command` (consuming `-p`), `env` (consuming `NAME=VALUE` words), `nohup`, `builtin`} with a bounded strip loop. The stripped head is resolved as an EXECUTABLE ONLY — it MUST NOT re-enter the funcs branch, because bash `command`/`command -p`/`builtin` bypass function lookup and `env`/`nohup` exec their argument: `rm() { :; }; command rm zone_dir/x` must deny, never walk the no-op body. `builtin` over-matches by design (bash rejects `builtin rm` and deletes nothing; judging it as the verb is the safe direction — spec D3). Slice `cmd.Args` so all downstream arg processing (cd/sed/git/verbs/candidates) sees the stripped form. Empty stripped head → under-match as today. Non-enumerated wrappers stay accepted under-matches.
4. **K1.** In `zonePathCandidates`, track a `seenDashDash` flag; detect the exact word `--` before the `--`-long-option branch and the `-`-prefix skip; after the flag, every fully-literal word — hyphen-leading included — becomes a candidate; `--` itself never does.
5. **K5 (audit D2 adds the work bound).** In the `zoneCall` funcs branch, capture `prevCalling`/`prevCalls` before mutating and RESTORE them at frame exit instead of `delete(w.calling, name)`; the bound-exceeded early return already touches nothing. Restore-on-exit fixes the state-reset defect but bounds DEPTH only — a width-W body over bound B yields Θ(W^(B-1)) body walks (W=10, B=8 ≈ 10^7; codex measured a 3s timeout on a width-10 seed), so the walker ALSO carries a GLOBAL visit budget: one counter of walker entries across the whole command, checked at every `zoneWalkStmt` entry; on exceed it sets `unbounded` and aborts the walk immediately (return without descending); `checkProtectedZoneShell` then denies fail-closed (:1031). Suggested constant in the `zoneLoop*Cap` family's order — thousands, not millions (e.g. `zoneWalkerEntryBudget = 10000`, tunable at run-phase), sitting well inside the 5-10s hook budget.
6. **K7 (audit D5 — verdict order).** In `checkProtectedZoneShell`, evaluate the unbounded denial BEFORE the `!w.mutating` short-circuit (:997): the seed command's `f` explosion can exhaust the global visit budget during the first `f` call and abort the walk before the `rm` statement is ever walked — `mutating` stays false and the early return would answer allow while real bash deletes. The denial applies REGARDLESS of manifest state (audit D9a): an aborted walk may not answer allow whether the manifest is OK, invalid, or ABSENT; the `ZoneStateAbsent` degrade keeps its COMPLETED-walk meaning only — a walk that finished and collected its candidates still degrades to the baseline floor + audit row when no manifest exists.
7. **K8 (audit D6 — cwds state budget).** The cd branch unions the pre-cd set into the next set on EVERY cd (`next = append(next, w.cwds...)`), so a chain of n cds squares the possible-directory set toward 2^n states; add a possible-directory-set bound to the same budget family as the D2 counters (suggested `zoneCwdsBound`, same order as the other caps) — on exceed, set `unbounded` and abort the walk (deny fail-closed), never silent truncation.
8. TDD invariants: the K4 subtest is authored and observed RED BEFORE its fix lands (step 2), and the four existing subtests' assertion lines stay byte-identical (D4 of spec §C).

### M2 (High) — parsing-matrix sweep

Extend the parsing test file (or add `protected_zone_shell_matrix_test.go` if it outgrows it). Matrix dimensions per REQ-ZSP-007: declaration forms (kw vs `f()`) crossing the K2 class AND the certain-shadow negative control (`rm() { :; }; rm zone_dir/x` stays allow — real bash does not delete); wrappers (`command`/`env`/`nohup` × bare/`-p` where applicable, × {rm, cp, mv} × zone/non-zone targets, plus `command ls` and bare-`env X=1` controls) AND the audit-D1 cross — a declared function shadowing a mutation verb invoked through EACH enumerated wrapper (`rm() { :; }; command rm zone_dir/x` plus its `command -p`/`env`/`nohup` variants): every such cell asserts DENY, never the function body; recursion shapes (sibling, mutual a↔b, loop-nested, plus the audit-D2 branch-width stress cell whose total work a depth-only bound would not survive) — every unbounded cell asserts the fail-closed deny verdict (loop-unbounded category / handler deny), not termination alone; `--` positions (first arg, after options, across rm/cp/mv; negative control `rm -- outside.txt`); the audit-D5 cell — the seed command `f() { if false; then f; f; fi; }; f; rm zone_dir/a.log` itself asserting the fail-closed deny even when the budget aborts the walk before the rm statement is reached; the audit-D6 cell — a cd-chain cell whose directory set exceeds the set budget (16 consecutive `cd` statements followed by a zone-covered rm) asserting the unbounded deny; the audit-D9 cell — a budget-exhausting form followed by `rm .claude/hooks/a.sh` (the frozen-instruction baseline path) under an ABSENT manifest (a fixture root with no zone manifest), asserting deny so an ordering or manifest-exception mutant fails by construction; executable-path forms (`/bin/rm` with and without a declared shadow, `./rm`); quoting combinations crossing each class (single, double, escaped — exercising `zoneWordText` as it exists). Guard: the sweep command runs with `-v` (AC-ZSP-004's observable is the per-cell output), the runner `t.Fatal`s on an empty cell list, and every subtest prints its own `--- PASS` line so a zero-match selector cannot pass for a sweep.

### M3 (Medium) — scoped verification batch

Slot lease → `unset`-scrubbed single invocation `go test ./internal/hook/ -timeout 30m` → `go vet ./internal/hook/` → `golangci-lint run ./internal/hook/...` → flip acceptance.md cells → progress.md §E.2/§E.3. Full-suite judgment stays CI's; the local batch is scoped to `internal/hook` per lane-local rules.

## §G Anti-Patterns

- Do NOT weaken an existing denial to make a new test pass; every fix is additive interpretation.
- Do NOT touch `zoneWordText` quote handling (:86-128) — t1570 owns it.
- Do NOT remove the certain-shadow path (a straight-line `rm() { :; }` must still shadow — real bash would not delete).
- Do NOT edit the four frozen subtests' assertions; the recursion verdict-tightening goes through D-NEW-1 re-delegation.
- Do NOT run the full hook package suite without the slot lease; do NOT run `go test ./...` locally.

## §H Cross-References

- Parent: SPEC-SELF-IMPROVE-PROTECTED-ZONE-001 (REQ-SIPZ-007 shell mutation; §C.6 accepted under-matches) — this SPEC hardens its shell half, no requirement of it is amended.
- Evidence: `.moai/reports/t1574/red-reproduction.md` (gate ledger round 1: 5 findings, all in this scope); `acceptance.md` §B EV-1..EV-5.
- Open question routed to the leader: `protected_zone_path.go:183` backslash (`zoneSlash`, t1556 §F reservation) — OUT of this SPEC pending judgment (spec § Out of Scope).
- Separated by the leader's scope decision (2026-10-07): the cd-tracking class (former REQ-ZSP-009/AC-ZSP-007/EV-6, plan K6 and its M2 cd cells) moves to follow-up card **t1584** under SPEC-SELF-IMPROVE-PROTECTED-ZONE-001; probe evidence preserved at `.moai/reports/t1574/`.
- Boundary: quote-class internals → card t1570; `integration_remeasure.go:225` → relayed to the owning lane.
