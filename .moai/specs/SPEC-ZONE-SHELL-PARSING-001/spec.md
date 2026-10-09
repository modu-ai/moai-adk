---
id: SPEC-ZONE-SHELL-PARSING-001
title: "Protected-zone shell guard — close five shell-parsing bypasses in the mutation-verb walk (-- operands, conditional declarations, wrapper prefixes, executable-path function absorption, recursion state reset)"
version: "0.5.0"
status: completed
created: 2026-10-07
updated: 2026-10-08
author: manager-spec
priority: P1
phase: "v3.2.0 target"
module: "internal/hook"
lifecycle: spec-anchored
tier: M
depends_on: [SPEC-SELF-IMPROVE-PROTECTED-ZONE-001]
tags: "protected-zone, shell-parsing, pretooluse, guard, fail-closed, t1574"
---

# SPEC-ZONE-SHELL-PARSING-001

## HISTORY

| Version | Date | Author | Change |
|---------|------|--------|--------|
| 0.1.0 | 2026-10-07 | manager-spec | Initial plan-phase draft (card t1574, security P1×2+P3, Class C). Five verified shell-parsing defect classes in the protected-zone mutation walk, all reproduced RED at tree b9ef003808da2ac1dfe42e5374d5bde4302f46b3 (four by the card's regression test, the fifth by a plan-phase probe and confirmed by the merge-gate ledger). Parsing-matrix sweep mandated by the card. Evidence ledger: `acceptance.md` §B. |
| 0.2.0 | 2026-10-07 | manager-spec | Plan-audit iteration 1 repair (FAIL 0.79; blocking D1-D4, optional O1-O4 folded). D1: wrapper-stripped heads no longer consult the funcs table — `command`/`command -p`/`builtin` bypass function lookup and `env`/`nohup` exec their argument in real bash, so `rm() { :; }; command rm zone/x` denies instead of walking the no-op body; `builtin`'s over-match disposition stated; wrapper × declared-function deny cells added to the matrix (REQ-ZSP-003/007, D3). D2: a global visit budget joins restore-on-exit as the work bound — a depth-only bound leaves Θ(W^(B-1)) walks (REQ-ZSP-006). D3/D4 (acceptance.md): the unbounded deny is owned observably by AC-ZSP-004's cells; its command carries `-v`. No REQ added or removed; count stays 8. |
| 0.3.0 | 2026-10-07 | manager-spec | Gate round-2 finding folded in (lane decision: same worker, same allow-consequence, `--` shares REQ-ZSP-001's separator semantics, and the absolute-cd fix reuses the parent SPEC's REQ-SIPZ-006 lexical strip — no frozen-family test asserts either reset). New REQ-ZSP-009 (cd tracking soundness: the `--` terminator yields one directory operand; an in-project absolute cd destination is tracked under its root-relative form), two cd cells in the matrix (REQ-ZSP-007), AC-ZSP-007 + EV-6 probe evidence (both shapes observed allow at b9ef00380). REQ count 8 → 9. |
| 0.4.0 | 2026-10-07 | manager-spec | Leader scope decision: the card closes at its ORIGINAL defect scope with a single delta re-verdict. Separated: the cd-tracking class (REQ-ZSP-009, AC-ZSP-007, EV-6, plan K6/M2 cd cells and their allow controls) moves to follow-up card t1584 under the parent SPEC — REQ count 9 → 8 restored, Out of Scope carries the handoff reference, probe evidence preserved at `.moai/reports/t1574/`. Folded in: D5 — the unbounded denial now precedes the mutating-only short-circuit (a budget exhausted mid-walk can abort before the mutating statement is reached; REQ-ZSP-006 + plan M1 + a seed-command matrix cell); D6 — the possible-directory-set budget joins the walk bounds (cd chains square the set; REQ-ZSP-006 + plan M1 + a cd-chain matrix cell); D8a — the stale M1-step reference in acceptance.md corrected. |
| 0.5.0 | 2026-10-07 | manager-spec | Leader conditional-final ruling — one narrow D9 repair, hunk-scoped re-verdict to follow. D9(a): REQ-ZSP-006 now states the unbounded fail-closed denial applies REGARDLESS of manifest state — an aborted walk is an incomplete walk and may not answer allow; the `ZoneStateAbsent` exception survives only for a COMPLETED walk. D9(c): the absent-manifest deny cell added in the three places the D5/D6 cells live — REQ-ZSP-007's matrix minimum, plan M2's cell catalogue, AC-ZSP-004's enumeration (a budget-exhausting shape followed by `rm .claude/hooks/a.sh`, the frozen-instruction baseline path, under an absent manifest, asserting deny). plan M1 step 6 (K7) manifest-state sentence aligned. O7: M1 heading annotated (steps 1-5 interpretation fixes, steps 6-7 audit repairs). O8: spec §C one-line note extending the fail-closed direction to REQ-ZSP-006's budget semantics. REQ/AC id sets unchanged (8/6); K7's mechanism (order swap + the two budgets) unchanged. |

## §A Problem

The protected-zone shell guard (`internal/hook/protected_zone_shell.go`) decides whether a Bash call carrying a self-improvement identity (`harness-learner`) may run. The walk parses the command with mvdan/sh and judges mutating verbs, candidates, and control flow as a set of possible worlds. Five parsing shapes defeat that judgment today. Every observation below was made at tree `b9ef003808da2ac1dfe42e5374d5bde4302f46b3` (branch `WT-shell-parsing-guard`), 2026-10-07; the raw outputs live in `acceptance.md` §B and `.moai/reports/t1574/red-reproduction.md`.

1. **`--` separator operands are dropped (:624).** `zonePathCandidates` skips every hyphen-leading word, so after a bare `--` separator — which in POSIX makes every following word an operand — `rm -- -zone/secret.md` names zero candidates and a protected file whose name legitimately starts with a hyphen is deleted. Observed allow (`acceptance.md` EV-1).
2. **A conditional declaration absorbs the real verb (:432).** `false && rm() { :; }; rm zone_dir/secret.md`: the declaration sits in a never-executed branch, but the branch-join union leaves the name holding ONLY the no-op body — the joined world loses the possibility that the name is undefined. The following real `rm` walks the no-op body and the protected deletion is allowed. Observed allow (EV-1).
3. **Static wrapper prefixes are not stripped (:598).** `command rm zone_dir/a.log`, `env rm …`, `command -p rm …`: the raw head word is matched against the mutation verbs, matches nothing, and the call under-matches to allow. The Claude Code permission layer strips the command/builtin wrappers, so the two layers now interpret the same command differently. Observed allow (EV-1).
4. **An executable path is absorbed as a function call (:430).** `rm() { :; }; /bin/rm zone_dir/secret.md`: the path fold (`path.Base`) runs BEFORE the function lookup, so `/bin/rm` resolves to the in-command `rm` function. Real bash dispatches functions by bare name only — the external binary runs and the file is deleted for real, while the guard walks the no-op body. Observed allow by a plan-phase probe (EV-3), independently listed by the merge-gate ledger round 1.
5. **Recursion accounting resets at frame exit (:461/:468).** An inner call's frame exit runs `delete(w.calling, name)`, erasing the OUTER frame's in-flight marker; the next sibling re-entry resets the bounded count to 1 and the bound never holds. `f() { if false; then f; f; fi; }; f; rm zone_dir/a.log` — a command real bash terminates instantly — crashes the Go walker with `fatal error: stack overflow`. A crashed hook process is the fail-open death path. Observed crash, twice independently (EV-2); this was the gate's previously undemonstrated claim, demonstrated by this card (card condition "실증 시 편입" → 편입 확정).
6. **(Separated — cd tracking.)** A sixth class reproduced in this tree (`cd -- zone_dir …` and an in-project absolute `cd` degenerating the possible-directory set to the root, probe-observed allow) was SEPARATED from this SPEC by the leader's scope decision (2026-10-07): the card closes at its original defect scope, and the cd class moves to follow-up card t1584 under the parent SPEC. Probe evidence is preserved at `.moai/reports/t1574/` for t1584 to inherit; see Out of Scope.

The card mandates a parsing-matrix sweep over the combinations these classes expose, and requires the existing thirteen-mutation-form table and the whole existing `TestProtectedZone` family to stay green unchanged.

## §B Requirements (GEARS)

- **REQ-ZSP-001** (`--` operands) — When a mutating command's argument list carries the bare `--` separator word, the guard shall treat every word after the separator as an operand and shall collect each fully-literal operand — a hyphen-leading name included — as a zone candidate; the separator word itself shall not become a candidate, and dynamic (non-literal) words keep the existing under-match.
- **REQ-ZSP-002** (conditional declarations keep absence) — When a function declaration is reachable only inside a conditional world (a branch of an `if`/`&&`/`||` chain, a loop body, or a `case` arm) and a later call in the same command resolves that name, the guard shall judge the call under BOTH interpretations — every possibly-declared body AND the built-in/external verb resolution the shell would use when the declaration did not run — and shall deny the call when either interpretation pairs a mutating form with a zone-covered target. A declaration seen unconditionally (straight-line) keeps today's shadowing behavior.
- **REQ-ZSP-003** (static wrapper prefixes) — When a call's head word is one of the static wrapper forms (`command`, `command -p`, `env`, `nohup`, `builtin`), the guard shall strip the wrapper prefix and judge the underlying command as an executable invocation — the stripped head shall never resolve to a function declared in the same command, because these wrappers bypass shell functions in real bash (`command`/`builtin` skip function lookup; `env`/`nohup` are external binaries that exec their argument); functions apply to bare, unwrapped heads only. The enumerated set shall match what the Claude Code permission layer strips (command/builtin) plus the card matrix (env, nohup), and wrapper forms outside the enumerated set remain documented under-matches.
- **REQ-ZSP-004** (executable paths are not function calls) — When a command word contains a path separator, the guard shall never resolve it to a function declared in the same command — a shell dispatches functions by bare name only — and shall resolve the path's base name through the verb judgment instead.
- **REQ-ZSP-005** (recursion accounting survives frame exits) — While a function declared in the command is being walked, the guard shall preserve the in-flight marker and the bounded call count across each frame's exit, so that a sibling or outer re-entry of the same name continues the count instead of restarting it.
- **REQ-ZSP-006** (bounded termination) — When any walk bound is exceeded — the recursion bound, a loop fixed point, the loop literal cap, the walker's GLOBAL visit budget (a cap on total walker entries across the whole command), or the possible-directory-set budget (a cap on how large a chain of `cd` statements may grow the cwds set) — the guard shall set the unbounded flag and deny fail-closed, and a bound exceeded mid-walk shall abort the walk immediately rather than let remaining branches run; the unbounded denial shall be evaluated BEFORE the mutating-only short-circuit, because a budget exhausted mid-walk can abort the walk before the mutating statement is ever reached and the early return would answer allow while the shell deletes; and the unbounded fail-closed denial applies REGARDLESS of manifest state — an aborted walk is an incomplete walk and an incomplete walk may not answer allow, the `ZoneStateAbsent` exception surviving only for a COMPLETED walk (its original meaning: the bounds of a walk that did finish and collect its candidates); restore-on-exit accounting bounds recursion DEPTH while the global budgets bound TOTAL WORK, so a finite command with wide branching or a long cd chain stays inside the hook's execution budget; the walker shall terminate on every parseable command, and a command that a real shell terminates shall never crash the hook process.
- **REQ-ZSP-007** (parsing-matrix sweep) — The guard's shell parsing shall carry a regression matrix covering at minimum: both function-declaration forms (`function f { … }` and `f() { … }`) under each fixed class, the enumerated wrapper prefixes (`command`/`env`/`nohup`, with and without `-p` where applicable), wrapper × declared-function cells (a declared function shadowing a mutation verb, invoked through each enumerated wrapper — every such cell asserting deny, never the function body), recursion shapes (sibling calls, mutual recursion between two names, loop-nested recursion, and a branch-width stress cell whose total work a depth-only bound would not survive), `--` operand positions (first argument, after options, across `rm`/`cp`/`mv`), the seed recursion command itself asserting the fail-closed deny even when the budget aborts the walk before the mutating statement is reached, a cd-chain cell whose directory set exceeds the set budget, an absent-manifest cell (a budget-exhausting shape followed by `rm .claude/hooks/a.sh`, the frozen-instruction baseline path, under a fixture root with no zone manifest — asserting deny, so an ordering or manifest-exception mutant fails the cell by construction), executable-path forms, and quoting combinations crossing each class — every zone-covered cell asserting deny and every unbounded recursion cell asserting the fail-closed deny verdict, not termination alone; the matrix runner shall make an empty cell list fail rather than pass.
- **REQ-ZSP-008** (preserved behavior) — The guard shall keep the existing thirteen-mutation-form table (`rm`, `unlink`, `mv`, `cp`, `tee`, `truncate`, `sed -i`, `>` redirection, `>>` redirection, `git rm`, `git checkout`, `git restore`, `git apply`) and the existing `TestProtectedZone` test family passing unchanged; the repairs of REQ-ZSP-001..006 are additive interpretation fixes that do not alter the verdict of any already-covered form.

## §C Constraints and Decisions

- **Scope.** `internal/hook/protected_zone_shell.go` and `internal/hook/protected_zone_shell_parsing_test.go` (plus an optional new sweep test file if the matrix outgrows the parsing file). Nothing else: no template, config, settings matcher, or Makefile change.
- **Fail-closed direction.** Within the newly interpreted classes (REQ-ZSP-001..005) resolution must keep a zone-covered deletion denied; when a new interpretation cannot decide, deny. The PRE-EXISTING accepted under-match classes — parse failure and dynamic words (parent SPEC §C.6) — remain unchanged; this SPEC does not widen or narrow them.
- The fail-closed direction extends to REQ-ZSP-006's budget semantics (audit D9/O8): a walk that cannot complete within its bounds answers deny, not allow — an incomplete walk is not evidence of safety.
- **D1 — conditional-declaration semantics = dual interpretation, not blanket deny.** A blanket deny on any conditionally-declared name would deny benign scripts whose helper never touches the zone; dual interpretation denies only when one of the two interpretations actually reaches the zone, and fail-closed holds because a single reaching interpretation suffices to deny. The joined world must preserve the POSSIBILITY OF ABSENCE for any name declared only inside a branch.
- **D2 — wrapper strip set = {`command`, `command -p`, `env`, `nohup`, `builtin`}.** Parity with the Claude Code permission layer (command/builtin) is the floor; env and nohup come from the card matrix and RED seeds. `exec`, `nice`, `timeout`, `xargs`, `sudo` and every other unenumerated prefix stay accepted under-matches (documented).
- **D3 — wrapper-stripped heads never consult the funcs table.** The RAW (unwrapped) head keeps today's funcs lookup — functions shadow bare invocations in real bash. But `command`, `command -p` and `builtin` bypass function lookup, and `env`/`nohup` are external binaries that exec their argument, so once a wrapper is present the stripped head is resolved ONLY through the verb/executable judgment: `rm() { :; }; command rm zone_dir/x` must deny, never walk the no-op body (plan-audit D1). `env` assignment words (`NAME=VALUE`) are consumed during the strip; a stripped head that resolves to no verb (e.g. bare `env X=1`) under-matches exactly as today. `builtin` over-matches by design — bash rejects `builtin rm` (rm is not a builtin) and deletes nothing, so judging it as the verb is the safe direction; this is the documented disposition.
- **D4 — the frozen recursion assertion stays frozen.** The four RED subtests' assertions are the regression gate; run-phase flips them GREEN without editing them. Tightening the recursion subtest from "terminates" to "denies fail-closed" is an AC re-tightening that goes through the D-NEW-1 orchestrator re-delegation path, never a silent test edit.
- **Methodology.** `constitution.development_mode: tdd` — the RED subtests exist; run-phase lands the :430 subtest RED-first, then flips all five GREEN.

## §D Acceptance Criteria

The canonical AC enumeration, evidence ledger, and Given-When-Then scenarios live in `acceptance.md` (Tier M). Summary:

| AC | Claim | RED-now | Green path |
|----|-------|---------|------------|
| AC-ZSP-001 | The three allow-class RED subtests (dash-dash, conditional declaration, wrapper prefix) are GREEN | EV-1 | M1 |
| AC-ZSP-002 | The recursion subtest terminates with zero recurrence of the stack overflow | EV-2 | M1 |
| AC-ZSP-003 | The :430 executable-path subtest is observed RED, then GREEN | EV-3 | M1 |
| AC-ZSP-004 | The parsing-matrix sweep exists, sweeps a non-empty cell list with `-v` per-cell output, is GREEN, and carries every unbounded cell's deny verdict | EV-4 | M2 |
| AC-ZSP-005 | The existing `TestProtectedZone` family passes unchanged alongside the new tests | EV-5 | M1–M3 |
| AC-ZSP-006 | Scoped verification batch green: hook package suite, `go vet`, `golangci-lint` on `internal/hook` | EV-5 | M3 |

## Out of Scope

### Out of Scope — quote-class internals

- Backslash and double-quote escape resolution inside `zoneWordText` (sites :113/:121) is owned by card t1570; this SPEC exercises quoting only as it exists and changes none of it.

### Out of Scope — cd tracking soundness (separated to card t1584)

- The cd-tracking class — `cd -- zone_dir && rm a.log` and an in-project absolute `cd` degenerating the possible-directory set, formerly drafted here as REQ-ZSP-009/AC-ZSP-007 with probe evidence EV-6 — is owned by follow-up card t1584 (issued) under SPEC-SELF-IMPROVE-PROTECTED-ZONE-001; its allow-control cells (an outside-root destination, a hyphen-leading directory name) moved with it.
- The probe evidence is preserved at `.moai/reports/t1574/` (card evidence path) for t1584 to inherit.

### Out of Scope — file-path layer and other systems

- `protected_zone_path.go:183` POSIX backslash handling (`zoneSlash`) is an OPEN QUESTION (leader judgment pending); t1556 explicitly reserved it as a follow-up. It is not a requirement here and is re-surfaced in the plan report.
- `integration_remeasure.go:225` belongs to a different system; relayed to the owning lane, not carried here.

### Out of Scope — wrapper verbs beyond the static set

- `exec`, `nice`, `timeout`, `xargs`, `sudo` and every other unenumerated wrapper prefix stay documented under-matches (D2); adding them is a future hardening decision, not this card.

### Out of Scope — surfaces outside the two files

- No template, `.moai/config`, settings.json matcher, or Makefile change; the PreToolUse matcher stays `Write|Edit|Bash` (parent SPEC §A.5).
