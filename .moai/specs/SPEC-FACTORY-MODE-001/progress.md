# SPEC-FACTORY-MODE-001 — Progress

## §E.1 Plan-phase Audit-Ready Signal

- Tier: L (5 artifacts + progress.md)
- Artifacts authored: spec.md, plan.md, acceptance.md, design.md, research.md
- Requirements: **26** · Acceptance-criterion leaves: **36** — both declared over the Tier L ceiling of 25, per `spec.md` §D Budget exception. The v0.2.0 declaration of "25 criteria" was wrong: the real leaf count was 34. It is corrected upward, not restored downward.
- Open blockers: none. No unresolved clarification markers remain (plan.md §H is now Resolved decisions).
- plan_status: audit-ready

### Revision history

| Iteration | Verdict | Disposition |
|---|---|---|
| iter-1 | FAIL (0.68, Tier L threshold 0.85) | 4 BLOCKING (D1-D4), 3 MAJOR (D5-D7), 5 MINOR (D8-D12) |
| iter-2 | FAIL (0.80, Tier L threshold 0.85) | 10 of 12 iter-1 defects genuinely CLOSED; 6 new BLOCKING (B1-B6) + 5 optional |
| iter-3 | FAIL (0.84, Tier L threshold 0.85) | v0.3.0 — B1-B6 closed; 3 of 5 optional items fixed; AC count declared honestly at 36 leaves with the over-budget exception recorded in `spec.md` §D. 6 blocking stale-text items raised (S1-S6): survivors the v0.3.0 corrections did not sweep |
| iter-4 | **PASS (0.94, Tier L threshold 0.85)** | v0.4.0 — S1-S6 swept and independently verified by the auditor (both shell loops executed; all six gate-token baselines re-measured against the widened and prior token sets). Confirming audit found no requirement-contradicting survivor; D1 (this record) and D3 (`design.md` §1 gate numbering) fixed as a trivial post-audit edit pass |

iter-3 changes by defect:

- **B1** (verify partition not disjoint; REQ-FM-011 vs REQ-FM-013 conflict on a readable DEGRADED result with no confirmed findings, and REQ-FM-013's gate unreachable for exactly that result) → restated as a 3-case **severity partition** (S1/S2/S3) crossed with an **orthogonal rung attribute**, with explicit precedence (severity governs routing, rung governs suppression). Resolution **(b)** chosen from the auditor's two options: a DEGRADED result auto-proceeds with Phase 8 suppression forced OFF, rather than widening the human gate — rationale recorded in REQ-FM-013, and option (a) kept as the named rejected alternative. The `disjoint`-of-four claim is withdrawn from `spec.md` §A, `design.md` §1, and (swept in v0.4.0) `plan.md` §B R4 Resolution, which the v0.3.0 pass missed. AC-FM-024b rewritten: it no longer greps for the word `disjoint` (a false disjointness claim satisfied that check) and instead asserts the two-axis structure plus both halves of the precedence sentence, with a falsification control against the v0.2.0 text.
- **B2** (rung exclusion failed open on an absent field — reachable, because `deepscan_dir` / `verify_rung` / `verify_reentries` are written independently on a best-effort record) → REQ-FM-014 inverted from a `!= DEGRADED` deny-list to an allow-list requiring `verify_rung` **recorded and equal to** `PRIMARY` or `FALLBACK`. AC-FM-020c gains a third table row: empty `verify_rung` + otherwise-TRUE predicate composes to `false`. `design.md` §7 records why absence is reachable. **v0.4.0 sweep**: the v0.3.0 pass left the deny-list formula standing in `design.md` §4's composed-decision pseudocode (`verify_rung != "DEGRADED"`); it is now `verify_rung ∈ {"PRIMARY", "FALLBACK"}`. The sites carrying the allow-list are therefore `spec.md` REQ-FM-014, `design.md` §4 and §7, `plan.md` §G AP-13, and `acceptance.md` AC-FM-020c.
- **B3** (AC-FM-025's mirror grep unsatisfiable at baseline — 8 pre-existing matches in 5 shipped files this SPEC never touches) → the grep is bounded to `$MIRRORS`, the mirrored counterparts of the §A.2 set, with the measured baseline of 0 recorded. The 8 pre-existing references are enumerated as out of scope in `spec.md` §C; scrubbing five unrelated rule files was rejected as a scope-discipline violation.
- **B4** (`os.Setenv` leaked process-globally and broke AC-FM-022a's own negative control by test ordering) → REQ-FM-023 gains a scoping clause requiring capture-and-`defer`-restore of prior value *and* prior presence, on both the success and error paths; new AC-FM-023d asserts `os.LookupEnv` returns the recorded pair after `runCC` returns. `design.md` §6 shows the capture/set/defer shape and states why the launch-env-slice alternative does not work with an `os.Getenv` reader.
- **B5** (AC-FM-012 could not decide its own invariant — ~35 context-free gate-token fragments, `sort -u` deduping nothing under `-n`, and a fourth gate indistinguishable from the pre-existing ones) → converted to a per-file **baseline-delta** check with measured baselines (`moai.md` 12, `run.md` 23, `review.md` 0, `factory.md` 0/absent, `quality-gates-quality.md` 0, `goal-directive.md` 12) and a planned per-file `N` enumerated in advance. The pre-flight capture is added as `plan.md` §C item 5, because running it after M1 destroys the evidence permanently.
- **B6** (REQ-FM-023 falsifies `goal-directive.md`'s launch-time-only sentence; the file appeared only as a plan §I cross-reference, in no REQ and no milestone) → new REQ-FM-028 + AC-FM-026; the file and its byte-identical template mirror are added to acceptance §A.2 and to M1's edit list, and to `plan.md` §I as an edit target rather than a reference.

Optional items: **fixed** — the REQ-FM-012 gate count (four fire, one added by this SPEC, no fifth), AC-FM-007's second grep (measured baseline 1 → delta assertion), AC-FM-020b's over-strict single pattern (split into separately-counted subject and predicate patterns, tolerating the bold marker and either tense). **Left as-is** — REQ-FM-026/027 numeric ordering (renumbering would churn every traceability row for a presentational gain), and the `Where`-vs-`When` notation on `spec.md` REQ-FM-001 / 005 / 023 (every entry still matches a valid GEARS pattern; REQ-FM-014 was reopened for B2 and took the corrected `When` form in passing).

iter-2 changes by defect:

- **D1** (rigor-reduced scan suppresses the independent one) → REQ-FM-013 records `verify_rung`; REQ-FM-014 excludes `DEGRADED` from suppression; AC-FM-020c asserts the exclusion with a negative Go test. `effort_tier` deliberately not added to the predicate — an effort level is not a rung.
- **D2** (dependency audit disabled) → REQ-FM-014 scoped to Phase 8 Step 0.55.1 by name; manifest audit stated to run unconditionally; AC-FM-020b asserts the pre-existing always-on sentence survives unmodified.
- **D3** (verify gate fails open) → REQ-FM-026 (no readable result → HALT, off the re-entry ceiling); REQ-FM-011 guarded on *readable*; REQ-FM-015 gains the `findings.jsonl` completeness conjunct; AC-FM-024a/b + AC-FM-019b.
- **D4** (unfalsifiable predicate) → REQ-FM-015 defines `headSHA` / `treeDirty` derivation; REQ-FM-017 adds the consumption clause; AC-FM-020d greps the sync doctrine for the predicate call and both derived inputs.
- **D5** (inject unreachable) → REQ-FM-023 names the environment route, the single call site `launcher.go:790`, and the `os.Environ()` seam at `:783`/`:786`; AC-FM-022a restated against `injectStopHookBlockCapForGoal` directly; AC-FM-023c asserts the variables reach the child environment.
- **D6** (markers re-opening answered questions) → plan.md §H replaced with Resolved decisions; markers 2 and 3 closed by citing REQ-FM-005 and REQ-FM-018 + §C; marker 1 closed by the user decision.
- **D7** (mirror-timing contradiction) → §D same-milestone policy retained and justified (CI fires on push); M6 rewritten as verification-only; mirror obligation extended into M1-M4.
- **D8** (internal package path in mirrored prose) → REQ-FM-025 forbidden list extended with internal Go package paths; AC-FM-025 grep extended with a positive control.
- **D9** (unflagged milestone dependencies) → every milestone carries a `Depends on:` line; M3 and M4 swapped so the state-record schema lands before the gate that reads `verify_rung`.
- **D10** (session-wide cap blast radius) → acknowledged in REQ-FM-023, design §6, and `factory.md`; per-goal scoping listed out of scope with its reason.
- **D11** (SPEC-ID-shaped placeholders) → the two example identifiers in acceptance.md and design.md are replaced with forms that do not match the mechanical SPEC-ID regex, so a cross-SPEC scan no longer reports them as missing directories.
- **D12** (misquoted `review.md:221`) → corrected to the verbatim live text (no backticks around `--lean`); recorded in research.md R-13.
- **Testability** → 8 prose ACs converted to grep-anchored form; AC-FM-012 bounded to the enumerated file set of §A.2; ACs added for REQ-FM-021, REQ-FM-022, REQ-FM-002 constants, REQ-FM-024 child-env export; §C traceability table maps every REQ to its ACs.
- **Budget** → REQ-FM-003 folded into REQ-FM-002, REQ-FM-020 into REQ-FM-019, holding 25 REQ / 25 AC at the Tier L ceiling while admitting REQ-FM-026 and REQ-FM-027.

## §E.2 Run-phase Evidence

### Pre-flight baselines (captured before any M1 edit)

Captured in the run-phase entry turn, in this worktree, at HEAD `7171880a9`, per `plan.md` §C items 2-5. These are the measured values the delta-based acceptance criteria compare against; a post-M1 run cannot reconstruct them.

AC-FM-012 per-file gate-token counts (widened five-token pattern `HUMAN GATE|gate-sync-1|gate-sync-2|Implementation Kickoff Approval|AskUserQuestion`):

| File | Baseline |
|---|---|
| `.claude/skills/moai/workflows/moai.md` | 23 |
| `.claude/skills/moai/workflows/run.md` | 33 |
| `.claude/skills/moai/workflows/review.md` | 3 |
| `.claude/skills/moai/workflows/factory.md` | ABSENT |
| `.claude/skills/moai/workflows/sync/quality-gates-quality.md` | 2 |
| `.claude/rules/moai/workflow/goal-directive.md` | 18 |

Matches the authoring-time record in `plan.md` §C exactly — the tree has not moved.

Other pre-change captures:

- `grep -c 'gate-sync-1' .claude/skills/moai/workflows/moai.md` → `1` (AC-FM-007 delta baseline)
- `grep -c 'Factory Mode' .claude/rules/moai/workflow/goal-directive.md` → `0` (AC-FM-026 baseline)
- Mirror presence: five of six exist; `internal/template/templates/.claude/skills/moai/workflows/factory.md` reported MIRROR-ABSENT (expected — M1 creates it)
- AC-FM-025 bounded mirror grep (`internal/factory|internal/cli` over the five existing mirrors) → `0` matches
- `go test ./internal/cli/... ./internal/config/...` → all packages `ok` (baseline green)
- `moai spec lint .moai/specs/SPEC-FACTORY-MODE-001/spec.md` → `✓ No findings`
- `git rev-list --count --left-right origin/main...HEAD` → `0	0`

### M1 — Pipeline contract and the verify exit gate (doctrine)

All evidence below was captured in this run, in this worktree, on branch `feat/factory-mode` at HEAD `7171880a9` (pre-commit). Commands are quoted verbatim; counts are the literal command output.

#### Post-change gate-token counts vs the pre-flight baselines (AC-FM-012)

Command, per file: `grep -o 'HUMAN GATE\|gate-sync-1\|gate-sync-2\|Implementation Kickoff Approval\|AskUserQuestion' <file> | wc -l`

| File | Baseline | Planned `N` | Planned `A` | Expected | Observed | Status |
|---|---:|---:|---:|---:|---:|---|
| `moai.md` | 23 | 2 | 0 | 25 | **25** | PASS |
| `run.md` | 33 | 1 | 0 | 34 | **34** | PASS |
| `review.md` | 3 | 0 | 0 | 3 | **3** | PASS |
| `factory.md` | 0 (absent) | 4 | 1 | 5 | **5** | PASS |
| `sync/quality-gates-quality.md` | 2 | 0 | 0 | 2 | **2** | PASS (untouched by M1) |
| `goal-directive.md` | 18 | 0 | 0 | 18 | **18** | PASS |

Judgement (b) — token attribution for the non-zero deltas, read from the new lines only:

- `moai.md` +2 — `gate-sync-1` ×1 and `gate-sync-2` ×1, both inside the new `factory` contract bullet naming the inherited sync gates. Both are already-enumerated gates.
- `run.md` +1 — `Implementation Kickoff Approval` ×1, in the verify-gate ordering sentence. Already-enumerated gate.
- `factory.md` +5 — `Implementation Kickoff Approval` ×1 (gate row 1), `gate-sync-1` ×1 (row 3), `gate-sync-2` ×1 (row 4), `HUMAN GATE` ×1 (row 2, the verify decision), `AskUserQuestion` ×1 (row 2, the same verify decision expressed as an orchestrator-issued round). Verified by `grep -o … | sort | uniq -c` → `1 AskUserQuestion / 1 gate-sync-1 / 1 gate-sync-2 / 1 HUMAN GATE / 1 Implementation Kickoff Approval`.

Judgement (c) — no file exceeds `baseline + N + A`; every observed count equals the expected value exactly, so the escape valve was not needed.

#### Per-AC matrix (M1-scoped)

| AC | Status | Command | Observed output |
|---|---|---|---|
| AC-FM-007 (a) | PASS | ``grep -c 'factory` contract .*extends.*`full-pipeline' .claude/skills/moai/workflows/moai.md`` | `1` (pre-change 0) |
| AC-FM-007 (b) | PASS | `grep -c 'gate-sync-1' .claude/skills/moai/workflows/moai.md` | `2` (baseline 1 → delta +1) |
| AC-FM-008 (a) | PASS | `grep -c -- '/moai review --security --deep --repo' .claude/skills/moai/workflows/run.md` | `1` |
| AC-FM-008 (b) | PASS | `grep -c 'exit gate of run-phase' .claude/skills/moai/workflows/run.md` | `1` |
| AC-FM-009 (a) | PASS | `grep -c 're-enter run-phase scoped to the changed surface' .claude/skills/moai/workflows/run.md` | `1` |
| AC-FM-009 (b) | PASS | `grep -c 'shall not proceed to sync' .claude/skills/moai/workflows/run.md` | `1` |
| AC-FM-010 (a) | PASS | `grep -c 'at most two verify re-entries' .claude/skills/moai/workflows/run.md` | `1` |
| AC-FM-010 (b) | PASS | `grep -c 'Baseline-attribution' .claude/skills/moai/workflows/run.md` | `1` (pre-change 0) |
| AC-FM-011 (a) | PASS | `grep -c 'inherited sync-phase evidence' .claude/skills/moai/workflows/run.md` | `1` |
| AC-FM-011 (b) | PASS | `grep -c 'readable result' .claude/skills/moai/workflows/run.md` | `5` |
| AC-FM-012 | PASS | per-file loop above | all six equal `baseline + N + A` |
| AC-FM-013 (a) | PASS | `grep -c 'DEGRADED' .claude/skills/moai/workflows/run.md` | `3` |
| AC-FM-013 (b) | PASS | `grep -c 'verify_rung' .claude/skills/moai/workflows/factory.md` | `2` |
| AC-FM-021 (a) | PASS | `grep -c 'honored only in --lean mode' .claude/skills/moai/workflows/review.md` | `0` (pre-change 1) |
| AC-FM-021 (b) | PASS | `grep -c 'honored in both --lean and --deep' .claude/skills/moai/workflows/review.md` | `1` (pre-change 0) |
| AC-FM-024a (a) | PASS | `grep -c 'no readable result' .claude/skills/moai/workflows/run.md` | `2` |
| AC-FM-024a (b) | PASS | `grep -c 'HALT' .claude/skills/moai/workflows/run.md` | `2` |
| AC-FM-024a (neg control) | PASS | `grep -n 'no confirmed findings' .claude/skills/moai/workflows/run.md` | 3 lines, at 193 / 198 / 207 — **every one also carries `readable`**; no unguarded "no confirmed findings → proceed" line exists |
| AC-FM-024b (1) | PASS | `grep -c 'no readable result'` / `grep -c 'readable result'` | `2` / `5` (≥1 and ≥2) |
| AC-FM-024b (2) | PASS | `grep -c 'orthogonal' .claude/skills/moai/workflows/run.md` | `1` |
| AC-FM-024b (3) | PASS | `grep -c 'governs routing'` / `grep -c 'governs suppression'` | `1` / `1` |
| AC-FM-025 (mirror existence, M1 scope) | PASS | `ls -1 <6 mirrors>` | all six listed, including the newly created `factory.md` mirror |
| AC-FM-025 (bounded neutrality grep) | PASS | `grep -n 'internal/factory\|internal/cli' <6 mirrors>` | no output, exit 1 (baseline 0 preserved) |
| AC-FM-026 (a) | PASS | `grep -c 'at launch time' .claude/rules/moai/workflow/goal-directive.md` | `1` (clause survives, now qualified as trigger (1) of two) |
| AC-FM-026 (a) | PASS | `grep -c 'Factory Mode' .claude/rules/moai/workflow/goal-directive.md` | `1` (pre-change 0) |
| AC-FM-026 (b) | PASS | `grep -c 'Factory Mode' internal/template/templates/.claude/rules/moai/workflow/goal-directive.md` | `1` |
| AC-FM-026 (b) | PASS | `cmp <live> <mirror>` | exit 0 — BYTE-IDENTICAL preserved |
| AC-FM-026 (c) | PASS | `grep -n 'SPEC-FACTORY\|REQ-FM-\|AC-FM-\|internal/factory\|internal/cli' <mirror>` | no output, exit 1 |
| AC-FM-022c (neg control, partial) | PASS | `grep -c 'stop after' .claude/skills/moai/workflows/factory.md` | `0` — no prose turn clause authored |

AC-FM-024b judgement (a)/(b)/(c) by reading: `run.md` § Verify Exit Gate enumerates S1 / S2 / S3 as a three-row severity table titled "mutually exclusive and jointly exhaustive"; the rung is introduced in a separate subsection as "an **attribute of** an S1 or S2 result … not a fourth case standing beside S1/S2/S3", with "S3 produced no result and therefore carries no rung at all"; the precedence paragraph states both halves. The word `disjoint` does not appear, and no four-peer-outcome list exists.

#### Build, lint, boundary, and suite evidence

| Item | Command | Observed |
|---|---|---|
| E2 host build | `go build ./...` | exit 0 |
| E2 cross build | `GOOS=windows GOARCH=amd64 go build ./...` | exit 0 |
| E4 subagent boundary | `grep -rn 'AskUserQuestion' internal/template/templates/.claude/hooks/moai/ \| grep -v '^[^:]*:[0-9]*:[ \t]*#'` | no output, exit 1 |
| E5 lint | `golangci-lint run --timeout=2m` | `0 issues.` — no NEW findings, baseline was also clean |
| Template guards | `go test ./internal/template/...` | `ok github.com/modu-ai/moai-adk/internal/template 38.480s` (mirror-parity + neutrality guards pass) |
| Pre-flight packages | `go test ./internal/cli/... ./internal/config/...` | all packages `ok` |
| SPEC lint | `moai spec lint .moai/specs/SPEC-FACTORY-MODE-001/spec.md` | `✓ No findings — all SPEC documents are valid` |
| Template rebuild | `make build` | `catalog.yaml updated successfully (12403 bytes)` + `go build` exit 0; `git diff --stat internal/template/catalog.yaml` produced no diff (the M1 files are workflow sub-files, not catalog hash subjects), so no catalog change to commit |

#### Gaps (explicitly NOT observed in M1)

- `go test ./...` (full suite) was NOT run for M1 — only `./internal/template/...`, `./internal/cli/...`, and `./internal/config/...`. M1 is doctrine-only and touches no Go source, so no other package's behavior can change; the full suite is M6's obligation.
- Coverage (`go test -cover`) was NOT measured — M1 adds no Go code, so the `internal/cli` 90% / `internal/factory` 85% targets are neither advanced nor regressed. `internal/factory` does not exist yet (M3).
- AC-FM-022b's `grep -c 'stop-goal' factory.md` and `git diff --diff-filter=A … .claude/hooks/ internal/hook/` were not evaluated as a criterion here — REQ-FM-021 is M2's scope. `factory.md` does name `stop-goal` as the existing evaluator, so the criterion is not obstructed.
- AC-FM-022c's `--max-turns 0` / `--max-duration 14400` greps are NOT satisfied yet — the arming rule is authored in M2. Only the negative control (`stop after` → 0) was checked, and it holds.
- The `$MIRRORS` neutrality grep was run over the enumerated six mirrors only, per the AC-FM-025 scope bound. The 8 pre-existing `internal/...` references in unrelated shipped files were not touched and not scanned.

#### Residual risk

- The `factory.md` gate-token budget is exactly at its planned ceiling of 5. M2 writes the goal-preset arming rule into the same file; if it re-mentions `Implementation Kickoff Approval` rather than referencing the existing gate row, the AC-FM-012 count moves to 6 and the escape valve must be invoked (the excess would name an already-enumerated gate, so it is admissible — but it is better avoided).
- `goal-directive.md` and its mirror are byte-identical now. Any later edit to either side must be applied to both in the same commit or AC-FM-026 (b) fails.
- The `review.md` correction widens the `--repo` statement inside the `--lean` section and forward-references the `--deep` section. If a later edit reorders those sections, the cross-reference wording ("see the --deep Mode section below") goes stale, though no acceptance criterion depends on it.

### M2 — The `factory_chain` goal preset

All evidence below was captured in this run, in this worktree, on branch `feat/factory-mode` at HEAD `c0279a8fb` (pre-commit; M1 is the parent commit). Commands are quoted verbatim; counts are the literal command output. M2 touches exactly two files — `.claude/skills/moai/workflows/factory.md` and its template mirror — plus this progress record.

#### Post-change gate-token counts vs the M1 post-change values (AC-FM-012)

Command, per file: `grep -o 'HUMAN GATE\|gate-sync-1\|gate-sync-2\|Implementation Kickoff Approval\|AskUserQuestion' <file> | wc -l`

| File | Pre-flight baseline | Planned `N` + `A` | Expected | Observed after M2 | Status |
|---|---:|---:|---:|---:|---|
| `moai.md` | 23 | 2 + 0 | 25 | **25** | PASS (untouched by M2) |
| `run.md` | 33 | 1 + 0 | 34 | **34** | PASS (untouched by M2) |
| `review.md` | 3 | 0 + 0 | 3 | **3** | PASS (untouched by M2) |
| `factory.md` | 0 (absent) | 4 + 1 | 5 | **5** | PASS |
| `sync/quality-gates-quality.md` | 2 | 0 + 0 | 2 | **2** | PASS (untouched by M2) |
| `goal-directive.md` | 18 | 0 + 0 | 18 | **18** | PASS (untouched by M2) |

`factory.md` breakdown, command `grep -o 'HUMAN GATE\|gate-sync-1\|gate-sync-2\|Implementation Kickoff Approval\|AskUserQuestion' .claude/skills/moai/workflows/factory.md | sort | uniq -c`, verbatim output:

```
   1 AskUserQuestion
   1 gate-sync-1
   1 gate-sync-2
   1 HUMAN GATE
   1 Implementation Kickoff Approval
```

Judgement (a) — every file equals its expected value. Judgement (b) — M2 introduces **no new gate token at all**, so there is no new line to attribute; the five tokens are the same five M1 recorded. Judgement (c) — no file exceeds `baseline + N + A`; the escape valve was not needed. The M1 residual-risk note (that an M2 re-mention of the approval gate would push the count to 6) is **closed**: the new arming rule refers to the existing gate row by position ("the plan-to-run approval of gate 1 in § Human gates") rather than repeating the token.

#### Per-AC matrix (M2-scoped)

`§C` traceability maps REQ-FM-021 → AC-FM-022b, REQ-FM-022 → AC-FM-022c, and REQ-FM-027 → AC-FM-022c. Every leaf of both is evaluated below.

| AC | Status | Command | Observed output |
|---|---|---|---|
| AC-FM-022b (1) | PASS | `git diff --name-only --diff-filter=A origin/main...HEAD -- .claude/hooks/ internal/hook/` | no output, exit 0 — no Stop hook, evaluator, or hook script added |
| AC-FM-022b (2, positive control) | PASS | `grep -c 'stop-goal' .claude/skills/moai/workflows/factory.md` | `2` — the file names the **existing** evaluator, so the empty first result reads as "reuses the existing runtime", not "never mentions a runtime" |
| AC-FM-022c (a) | PASS | `grep -c 'Implementation Kickoff Approval' .claude/skills/moai/workflows/factory.md` | `1` |
| AC-FM-022c (b) | PASS | `grep -c -- '--max-turns 0' .claude/skills/moai/workflows/factory.md` | `1` |
| AC-FM-022c (c) | PASS | `grep -c -- '--max-duration 14400' .claude/skills/moai/workflows/factory.md` | `1` |
| AC-FM-022c (neg control) | PASS | `grep -c 'stop after' .claude/skills/moai/workflows/factory.md` | `0` — no prose turn clause authored |
| AC-FM-012 | PASS | per-file loop above | all six equal `baseline + N + A`; M2 adds zero tokens |
| AC-FM-013 (b) | PASS | `grep -c 'verify_rung' .claude/skills/moai/workflows/factory.md` | `2` (unchanged by M2) |
| AC-FM-025 (mirror existence) | PASS | `ls -1 <6 mirrors>` | all six listed |
| AC-FM-025 (bounded neutrality grep) | PASS | `grep -n 'internal/factory\|internal/cli' <6 mirrors>` | no output, exit 1 (baseline 0 preserved) |
| AC-FM-026 (b) | PASS | `cmp .claude/rules/moai/workflow/goal-directive.md internal/template/templates/.claude/rules/moai/workflow/goal-directive.md` | exit 0 — BYTE-IDENTICAL preserved |

AC-FM-022c judgement by reading: § The `factory_chain` goal preset → Arming rules states, as its first bullet, "Arm only after the plan-to-run approval of gate 1 in § Human gates is cleared", then "Arm alongside the work, never in place of it", then the flag bound and the accepted four-hour token risk. The ordering is arm-after-approval, and the arm-only consequence (idle turns spinning to a bound) is stated as the reason for the second bullet rather than asserted bare.

AC-FM-022b judgement by reading: the preset section opens by naming the `stop-goal` Stop-hook evaluator as the thing that evaluates the condition and states that the preset "introduces no new runtime, no new hook, and no new evaluator". The condition itself is a `text` block of model conditions only — no shell command whose exit code decides — so nothing in it requires a new mechanical evaluator either.

#### Mirror, build, lint, and guard evidence

| Item | Command | Observed |
|---|---|---|
| Mirror parity | `cmp .claude/skills/moai/workflows/factory.md internal/template/templates/.claude/skills/moai/workflows/factory.md` | exit 0 — BYTE-IDENTICAL (M1 mirrored verbatim; the authored text carries no internal token, so a verbatim mirror is also a neutral one) |
| Mirror neutrality (this file) | `grep -n 'SPEC-FACTORY\|REQ-FM-\|AC-FM-\|internal/factory\|internal/cli' <factory.md mirror>` | no output, exit 1 |
| Mirror neutrality (date / SHA) | `grep -nE '20[0-9]{2}-[0-9]{2}-[0-9]{2}\|\b[0-9a-f]{7,40}\b' <factory.md mirror>` | no output, exit 1 |
| E2 host build | `go build ./...` | exit 0 |
| E2 cross build | `GOOS=windows GOARCH=amd64 go build ./...` | exit 0 |
| E5 lint | `golangci-lint run --timeout=2m` | `0 issues.` — no NEW findings; the M1 baseline was also clean |
| Template guards | `go test ./internal/template/...` | `ok github.com/modu-ai/moai-adk/internal/template 43.207s` (mirror-parity + neutrality guards pass) |
| SPEC lint | `moai spec lint .moai/specs/SPEC-FACTORY-MODE-001/spec.md` | `✓ No findings — all SPEC documents are valid` |
| Template rebuild | `make build` | `catalog.yaml updated successfully (12403 bytes)` + `go build` exit 0; `git diff --stat internal/template/catalog.yaml` produced no diff — workflow sub-files are not catalog hash subjects, so there is no catalog change to commit (same result as M1) |
| Working-tree hygiene | `git status --porcelain` | exactly two modified paths, both `factory.md` (live + mirror); no runtime-managed or unrelated file touched |

#### Gaps (explicitly NOT observed in M2)

- `go test ./...` (full suite) was NOT run. M2 is doctrine-only and touches no Go source, so no package's behavior can change; only `./internal/template/...` was run, because that is the package whose guards observe the mirrored markdown. The full suite is M6's obligation.
- Coverage (`go test -cover`) was NOT measured. M2 adds no Go code, so the `internal/cli` 90% / `internal/factory` 85% targets are neither advanced nor regressed, and `internal/factory` does not exist yet (M3).
- E4's subagent-boundary grep over `internal/template/templates/.claude/hooks/moai/` was NOT re-run — M2 touches no hook script and no file under that directory, so the M1 result stands unchanged rather than being re-asserted here.
- `./internal/cli/...` and `./internal/config/...` were NOT re-run for M2 (unchanged since the M1 evidence, which recorded them `ok`); no Go package was touched.
- The AC-FM-012 escape valve was not exercised, so its judgement path remains unverified by observation — it is untested because it was not needed, not because it was checked and found sound.
- Whether the authored condition actually converges under a live `stop-goal` evaluation was NOT observed. The condition is doctrine text; no factory session has been run, and running one requires the launcher entry that M5 delivers.

#### Residual risk

- The `factory.md` gate-token budget is still exactly at its planned ceiling of 5, and M4 cross-references `factory.md` from the sync dedup doctrine. A future edit that quotes any of the four gate names into this file moves the count to 6 and forces the escape valve; the valve would admit it (the token would name an already-enumerated gate), but the cleaner move remains a positional reference.
- The condition's convergence rests on the orchestrator actually surfacing each named line — the audit verdict, the per-criterion PASS lines, the verify case and rung, the close record. A phase that completes its work without surfacing its line leaves the goal unmet and the chain running to a bound. This is inherent to model conditions and is the same exposure `ac_converge` carries; it is stated rather than mitigated.
- The four-hour wall clock is an accepted token risk, not a safety bound. It is the only automatic stop between the stagnation guard and chain completion, so a chain that neither converges nor stagnates will spend the full budget.

## §F Phase 4 Mode Selection

Input parameters: tier **L**; scope ~14 files (6 doctrine + 6 mirrors + 3 Go source + tests); domain count 3 (workflow doctrine / rules, Go source under `internal/`, template mirrors); file language mix markdown-heavy with a Go minority; concurrency benefit **LOW** (coding-heavy, and the milestones carry declared `Depends on:` edges — M2 depends on M1, M4 on M1+M3, M5 on M3).

| Mode | Selected | Rationale |
|---|---|---|
| 1 `trivial` | no | Multi-file semantic change across doctrine and Go source. |
| 2 `background` | no | Write-heavy; not read-only. |
| 3 `agent-team` | no | RETIRED tombstone; never selected. |
| 4 `parallel` | no | Coding-heavy with inter-milestone dependencies — the coding-task parallelism caveat routes this to sequential. |
| 5 `sub-agent` | **yes** | Sequential per-milestone delegation preserves the declared dependency order and the per-milestone reporting seam the user selected (semi-autonomous progression). |
| 6 `workflow` | no | Scope is far below the ~30-file mechanical threshold and the transform is not a single uniform rule. |

Decision: `sub-agent`

Justification: the six milestones carry explicit `Depends on:` edges that forbid a parallel fan-out — M2 writes into the `factory.md` M1 creates, and M4's dedup gate reads the `verify_rung` field M3 defines. The work is also coding-heavy rather than research-heavy, which is the case Anthropic's parallelism caveat routes to sequential sub-agent delegation. Implementation Kickoff Approval was obtained in the plan-phase session, together with the semi-autonomous progression choice; sequential Mode 5 is what makes that per-milestone reporting seam possible, which a `manager-lead` fan-out would collapse.

## §E.3 Run-phase Audit-Ready Signal

_<pending run-phase>_

## §E.4 Sync-phase Audit-Ready Signal

_<pending sync-phase>_
