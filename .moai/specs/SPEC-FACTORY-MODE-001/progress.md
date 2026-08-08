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

### M3 — Signal propagation and the state record

All evidence below was captured in this run, in this worktree, on branch `feat/factory-mode` at HEAD `598cf280e` (pre-commit; M2 is the parent commit). Commands are quoted verbatim; outputs are the literal command output. M3 is the first Go milestone and ran `cycle_type=tdd`. It touches exactly three paths — `internal/config/envkeys.go`, the new `internal/factory/` package, and this progress record. No `.claude/` file is touched, so M3 carries no mirror obligation and `make build` was deliberately not run (there is no template change for it to embed).

#### RED evidence (E8 — captured before any implementation existed)

Test-first was executed in two observable stages, both before `record.go` carried an implementation.

Stage 1 — `record_test.go` written first, with no `record.go` at all. Command `go test ./internal/factory/...`, verbatim output:

```
# github.com/modu-ai/moai-adk/internal/factory [github.com/modu-ai/moai-adk/internal/factory.test]
internal/factory/record_test.go:13:20: undefined: Record
internal/factory/record_test.go:14:10: undefined: RungPrimary
internal/factory/record_test.go:15:10: undefined: Record
internal/factory/record_test.go:18:20: undefined: BackendClaude
internal/factory/record_test.go:30:12: undefined: Write
internal/factory/record_test.go:34:14: undefined: Read
internal/factory/record_test.go:58:41: undefined: RungPrimary
internal/factory/record_test.go:60:24: undefined: RungPrimary
internal/factory/record_test.go:61:57: undefined: RungPrimary
internal/factory/record_test.go:69:12: undefined: RecordPath
internal/factory/record_test.go:69:12: too many errors
FAIL	github.com/modu-ai/moai-adk/internal/factory [build failed]
FAIL
```

Stage 2 — a compiling stub (declarations returning zero values) replaced the build failure with assertion-level RED, so the failure demonstrates the tests exercise behavior rather than merely referencing absent symbols. Command `go test ./internal/factory/...`, verbatim head of output:

```
--- FAIL: TestWriteThenReadRoundTripsEveryField (0.00s)
panic: runtime error: invalid memory address or nil pointer dereference [recovered, repanicked]
[signal SIGSEGV: segmentation violation code=0x2 addr=0x0 pc=0x100bcfe84]
...
github.com/modu-ai/moai-adk/internal/factory.TestWriteThenReadRoundTripsEveryField(0x7dba8d540248)
	.../internal/factory/record_test.go:39 +0xa4
...
FAIL	github.com/modu-ai/moai-adk/internal/factory	0.390s
FAIL
```

Line 39 is the `Read` result dereference: the stub returned `nil, nil`, so the round-trip assertion had nothing to read back. No implementation code was written before its failing test, so the delete-and-re-derive obligation did not arise.

#### Per-AC matrix (M3-scoped)

`§C` traceability maps REQ-FM-024 → AC-FM-023a, AC-FM-023b, AC-FM-013. Of those, AC-FM-023b is wholly M3's; AC-FM-023a's record layer is M3's while its launcher half belongs to M5; AC-FM-013 was already satisfied by M1/M2 doctrine and is re-confirmed unchanged.

| AC | Status | Command | Observed output |
|---|---|---|---|
| AC-FM-023b (1) | PASS | `grep -c 'EnvMoaiFactory = "MOAI_FACTORY"' internal/config/envkeys.go` | `1` (pre-change baseline `0`, measured before the edit — the positive control holds) |
| AC-FM-023b (2) | PASS | `grep -c 'EnvMoaiFactorySpec = "MOAI_FACTORY_SPEC"' internal/config/envkeys.go` | `1` (pre-change baseline `0`) |
| AC-FM-023a (record schema, M3 half) | PASS | `go test -run TestWrittenJSONCarriesTheDocumentedKeys ./internal/factory/...` | `ok` — the written JSON carries `session_id`, `spec_id`, `backend`, `entered_at`, `deepscan_dir`, `verify_rung`, `verify_reentries` |
| AC-FM-023a (round-trip, M3 half) | PASS | `go test -run TestWriteThenReadRoundTripsEveryField ./internal/factory/...` | `ok` — every field survives write→read |
| AC-FM-023a (fail-open, M3 half) | PASS | `go test -run TestWriteBestEffortFailsOpenOnUnwritableStateDirectory ./internal/factory/...` | `ok` — `Write` reports the failure, `WriteBestEffort` returns nothing a caller could gate on |
| AC-FM-023a (launcher half) | DEFERRED to M5 | — | the `moai cc --factory` path that writes the record is M5's deliverable; M3 delivers only the record layer it calls |
| AC-FM-013 (b) | PASS (unchanged) | `grep -c 'verify_rung' .claude/skills/moai/workflows/factory.md` | `2` — untouched by M3 |
| Rung round-trip (plan.md §F M3) | PASS | `go test -run TestVerifyRungRoundTripsEveryRung ./internal/factory/...` | all three subtests PASS: `PRIMARY`, `FALLBACK`, `DEGRADED` |
| Absent-vs-empty rung (design.md §7 / AP-13) | PASS | `go test -run TestAbsentRungIsDistinguishableFromEmptyRung ./internal/factory/...` | both subtests PASS — absent decodes to `nil`, recorded-empty decodes to a non-nil pointer to `""` |

Full verbose suite, command `go test ./internal/factory/... -v`, tail:

```
--- PASS: TestWriteThenReadRoundTripsEveryField (0.00s)
--- PASS: TestRecordPathIsSessionKeyedUnderStateFactory (0.00s)
--- PASS: TestWrittenJSONCarriesTheDocumentedKeys (0.00s)
--- PASS: TestVerifyRungRoundTripsEveryRung (0.00s)
    --- PASS: TestVerifyRungRoundTripsEveryRung/PRIMARY (0.00s)
    --- PASS: TestVerifyRungRoundTripsEveryRung/FALLBACK (0.00s)
    --- PASS: TestVerifyRungRoundTripsEveryRung/DEGRADED (0.00s)
--- PASS: TestAbsentRungIsDistinguishableFromEmptyRung (0.00s)
    --- PASS: TestAbsentRungIsDistinguishableFromEmptyRung/absent (0.00s)
    --- PASS: TestAbsentRungIsDistinguishableFromEmptyRung/recorded_empty (0.00s)
--- PASS: TestReadReturnsErrorWhenRecordIsAbsent (0.00s)
--- PASS: TestReadReturnsErrorOnMalformedJSON (0.00s)
--- PASS: TestWriteRejectsUnusableSessionIDs (0.00s)
    --- PASS: TestWriteRejectsUnusableSessionIDs/empty (0.00s)
    --- PASS: TestWriteRejectsUnusableSessionIDs/path_separator (0.00s)
    --- PASS: TestWriteRejectsUnusableSessionIDs/parent_escape (0.00s)
--- PASS: TestWriteBestEffortFailsOpenOnUnwritableStateDirectory (0.00s)
--- PASS: TestNewRecordStampsAnRFC3339EnteredAt (0.00s)
--- PASS: TestRecordWithoutSpecIDRoundTrips (0.00s)
PASS
ok  	github.com/modu-ai/moai-adk/internal/factory	0.345s
```

#### Build, coverage, lint, boundary, and suite evidence

| Item | Command | Observed |
|---|---|---|
| E2 host build | `go build ./...` | exit 0 |
| E2 cross build | `GOOS=windows GOARCH=amd64 go build ./...` | exit 0 — the record layer uses only `os` / `filepath` / `encoding/json` / `time`, so no build tag is required; verified rather than assumed |
| E3 coverage | `go test -cover ./internal/factory/...` | `ok github.com/modu-ai/moai-adk/internal/factory 0.392s coverage: 88.2% of statements` — above the 85% target for this package |
| E4 subagent boundary | `grep -rn 'AskUserQuestion\|mcp__askuser' internal/factory/ \| grep -v '_test.go' \| grep -v '// '` | no output, exit 1 |
| E5 lint | `golangci-lint run --timeout=2m` | `0 issues.` — no NEW findings; the M2 baseline was also clean |
| Formatting | `gofmt -l internal/factory internal/config` | neither `internal/factory/*.go` nor `internal/config/envkeys.go` is listed; the six files it does list are pre-existing and untouched by M3 |
| Regression — config | `go test ./internal/config/...` | `ok internal/config 1.861s`, `ok internal/config/toolpolicy` |
| Regression — cli | `go test ./internal/cli/...` | all 19 packages `ok` (`internal/cli` 222.300s) |
| Regression — template guards | `go test ./internal/template/...` | `ok internal/template 36.497s` (mirror-parity + neutrality guards pass, unchanged by M3) |
| SPEC lint | `moai spec lint .moai/specs/SPEC-FACTORY-MODE-001/spec.md` | `✓ No findings — all SPEC documents are valid` |
| Template rebuild | (not run) | M3 touches no file under `.claude/` or `internal/template/templates/`, so there is no template change to embed; running `make build` would have been a no-op asserted as evidence |
| Working-tree hygiene | `git status --porcelain` | exactly ` M internal/config/envkeys.go` and `?? internal/factory/` — no runtime-managed, `.moai/state/`, or unrelated file touched |

#### Design note — why `verify_rung` is a pointer

`design.md` §7 establishes that `deepscan_dir`, `verify_rung`, and `verify_reentries` are written independently on a best-effort record, so a record carrying a scan directory but no rung is reachable through an orchestrator omission, a mid-chain `/clear`, or a partial write. A plain `string` field collapses that state into the same `""` a deliberately-blank rung would produce. The field is therefore `*Rung` with `omitempty`: absent marshals away entirely and decodes to `nil`, while a recorded empty rung decodes to a non-nil pointer to `""`. Both read as "not a recorded `PRIMARY` or `FALLBACK`", so the M4 allow-list reaches the same answer either way — but it reaches it on observed evidence rather than on a coincidence of representation, which is the distinction AP-13 exists to preserve.

#### Gaps (explicitly NOT observed in M3)

- `go test ./...` (the full suite) was NOT run. The three suites most likely to observe M3's change were run instead (`./internal/factory/...`, `./internal/config/...`, `./internal/cli/...`) plus `./internal/template/...` for the mirror guards. The full suite is M6's obligation.
- `internal/cli` coverage was NOT measured against its 90% target. M3 adds no `internal/cli` code; that target is M5's to meet.
- AC-FM-023a's launcher half was NOT exercised — no test invokes `moai cc --factory`, because the flag does not exist until M5. What M3 evidences is that the record layer M5 will call round-trips correctly and fails open; that the launcher actually calls it is unverified and out of M3's scope.
- The fail-open test SKIPS on Windows and when running as root (POSIX directory permissions do not deny the write in either case), so the fail-open path is observed on POSIX-as-non-root only. It was observed here.
- `WriteBestEffort` is not yet called from any production path, so its fail-open guarantee is verified by unit test but not by a live launch. M5 supplies the caller.
- No `.moai/state/factory/` directory was created in this repository — every test writes under `t.TempDir()`. This was confirmed by the clean `git status --porcelain` above rather than assumed.

#### Residual risk

- `Write` and `WriteBestEffort` are a deliberate pair: the error-returning form exists for tests and tooling, the no-return form for the launch path. Nothing mechanically stops M5 from calling `Write` and gating the launch on its error, which would reintroduce exactly the blocking behavior the fail-open requirement forbids. The signature makes the safe call available; it does not make the unsafe one unreachable.
- The `verify_rung` allow-list itself is NOT implemented in M3 — only the schema that makes it expressible. M4 owns the composed suppression decision and must treat a `nil` pointer and a pointer to `""` alike as "no suppression". A `!= DEGRADED` deny-list written over this pointer would still be wrong, and this package cannot prevent it.
- `EnteredAt` is stored as a preformatted RFC3339 string rather than a `time.Time`. That matches the documented schema and avoids monotonic-clock round-trip surprises, but it also means a malformed timestamp written by a future caller would round-trip silently — nothing validates the format on read.
- The session-id validation rejects path separators and directory references, which covers the traversal shape reachable from a launcher-supplied identifier. It does not constrain length or character set beyond that, so an exotic-but-separator-free identifier would still produce a filename this package accepts.

### M4 — Dedup contract and the `revision.json` reader

Baseline attribution for every row below: **this run, this tree, HEAD `0c8b38ef8`**, branch `feat/factory-mode`. Pre-flight baselines were re-measured before the first M4 edit and matched the recorded §C figures.

#### Pre-flight baselines (re-measured before any M4 edit)

| Measurement | Command | Observed |
|---|---|---|
| Gate-token count, `sync/quality-gates-quality.md` | `grep -o 'HUMAN GATE\|gate-sync-1\|gate-sync-2\|Implementation Kickoff Approval\|AskUserQuestion' <file> \| wc -l` | `2` |
| `Step 0.55.1` | `grep -on '...' <file>` | 1 occurrence (line 112) |
| `regardless of whether manifest files changed` | same | 1 occurrence (line 121) |
| `scanned_commit` / `inherited from the factory verify stage` / `dependency manifest audit` / `run unconditionally` / `runs unconditionally` / `PRIMARY` / `FALLBACK` / `DEGRADED` / `git rev-parse HEAD` / `git status --porcelain` / `revision-match predicate` | same | **0 occurrences each** — the AC-FM-020a/c/d positive controls |
| Bounded mirror neutrality grep (6 mirrors) | `grep -rn 'internal/factory\|internal/cli' <6 mirror paths>` | no output (0 matches) |
| `goal-directive.md` mirror parity | `cmp <op> <mirror>` | exit 0 |
| Coverage baseline | `go test -cover ./internal/factory/...` | `coverage: 88.2% of statements` |

#### E8 — RED evidence (verbatim, captured BEFORE GREEN)

Two RED runs were captured. The first is the compile-level RED at the moment `revision_test.go` existed and `revision.go` did not:

```
$ go test ./internal/factory/...
# github.com/modu-ai/moai-adk/internal/factory [github.com/modu-ai/moai-adk/internal/factory.test]
internal/factory/revision_test.go:59:12: undefined: RevisionMatch
internal/factory/revision_test.go:67:12: undefined: RevisionMatch
internal/factory/revision_test.go:76:15: undefined: LoadRevision
internal/factory/revision_test.go:79:12: undefined: RevisionMatch
internal/factory/revision_test.go:102:15: undefined: LoadRevision
internal/factory/revision_test.go:105:12: undefined: RevisionMatch
internal/factory/revision_test.go:117:15: undefined: LoadRevision
internal/factory/revision_test.go:120:12: undefined: RevisionMatch
internal/factory/revision_test.go:132:12: undefined: RevisionMatch
internal/factory/revision_test.go:144:12: undefined: RevisionMatch
internal/factory/revision_test.go:144:12: too many errors
FAIL	github.com/modu-ai/moai-adk/internal/factory [build failed]
FAIL
```

A build failure is a weak RED — it proves the symbols are absent, not that the assertions discriminate. So signature stubs returning zero values were written and the suite re-run, producing the behavioral RED:

```
$ go test -run 'TestRevisionMatch|TestLoadRevision|TestMatches|TestSuppressStep0551' -v ./internal/factory/...
=== RUN   TestRevisionMatchHappyPath
    revision_test.go:60: RevisionMatch on a complete matching fixture = false, want true
--- FAIL: TestRevisionMatchHappyPath (0.00s)
=== RUN   TestRevisionMatchAbsentDirectoryIsFalse
--- PASS: TestRevisionMatchAbsentDirectoryIsFalse (0.00s)
=== RUN   TestRevisionMatchAbsentRevisionJSONIsFalse
--- PASS: TestRevisionMatchAbsentRevisionJSONIsFalse (0.00s)
=== RUN   TestLoadRevisionUnreadableFileIsError
--- PASS: TestLoadRevisionUnreadableFileIsError (0.00s)
=== RUN   TestLoadRevisionMalformedJSONIsError
--- PASS: TestLoadRevisionMalformedJSONIsError (0.00s)
=== RUN   TestRevisionMatchCommitMismatchIsFalse
--- PASS: TestRevisionMatchCommitMismatchIsFalse (0.00s)
=== RUN   TestRevisionMatchNonRepoScopeIsFalse
--- PASS: TestRevisionMatchNonRepoScopeIsFalse (0.00s)
=== RUN   TestRevisionMatchDirtyTreeWithWorkingTreeExcluded
    revision_test.go:162: RevisionMatch with a clean tree and working_tree_included=false = false, want true
--- FAIL: TestRevisionMatchDirtyTreeWithWorkingTreeExcluded (0.00s)
=== RUN   TestRevisionMatchFindingsCompleteness
    revision_test.go:179: RevisionMatch with a zero-line findings.jsonl = false, want true
--- FAIL: TestRevisionMatchFindingsCompleteness (0.00s)
=== RUN   TestRevisionMatchUnparseableFindingsLineIsFalse
--- PASS: TestRevisionMatchUnparseableFindingsLineIsFalse (0.00s)
=== RUN   TestMatchesNilRevisionIsFalse
--- PASS: TestMatchesNilRevisionIsFalse (0.00s)
=== RUN   TestSuppressStep0551RungAllowList
=== RUN   TestSuppressStep0551RungAllowList/recorded_PRIMARY_suppresses
    revision_test.go:229: SuppressStep0551 = false, want true
=== RUN   TestSuppressStep0551RungAllowList/recorded_FALLBACK_suppresses
    revision_test.go:229: SuppressStep0551 = false, want true
--- FAIL: TestSuppressStep0551RungAllowList (0.00s)
    --- FAIL: TestSuppressStep0551RungAllowList/recorded_PRIMARY_suppresses (0.00s)
    --- FAIL: TestSuppressStep0551RungAllowList/recorded_FALLBACK_suppresses (0.00s)
    --- PASS: TestSuppressStep0551RungAllowList/recorded_DEGRADED_does_not_suppress (0.00s)
    --- PASS: TestSuppressStep0551RungAllowList/never-recorded_rung_does_not_suppress (0.00s)
    --- PASS: TestSuppressStep0551RungAllowList/recorded-empty_rung_does_not_suppress (0.00s)
    --- PASS: TestSuppressStep0551RungAllowList/unrecognized_rung_does_not_suppress (0.00s)
    --- PASS: TestSuppressStep0551RungAllowList/failing_predicate_does_not_suppress_even_on_PRIMARY (0.00s)
FAIL
FAIL	github.com/modu-ai/moai-adk/internal/factory	4.210s
```

The stub RED is worth recording for what it demonstrates about the criteria themselves: an unconditionally-false predicate **passes every negative case** — absent directory, commit mismatch, non-repo scope, unparseable findings line, `DEGRADED`, `nil`, recorded-empty. Only the four positive controls fail. This is the mechanical demonstration of why AC-FM-019a, AC-FM-019b, and the AC-FM-020c row-1 positive control are stated as both-halves-in-one-test: without them the whole §B.3 suite is satisfiable by a function that never suppresses anything, which is safe but useless, and indistinguishable from a correct implementation by the test suite alone.

#### E1 — per-AC matrix

| AC | Status | Command | Observed output |
|---|---|---|---|
| AC-FM-014 (happy path) | PASS | `go test -run TestRevisionMatchHappyPath ./internal/factory/...` | `ok github.com/modu-ai/moai-adk/internal/factory` (within the 90.9% suite run below) |
| AC-FM-015 (absent `revision.json`) | PASS | `TestRevisionMatchAbsentRevisionJSONIsFalse` | asserts `LoadRevision` errors AND `RevisionMatch` false |
| AC-FM-016 (malformed JSON) | PASS | `TestLoadRevisionMalformedJSONIsError` | asserts error + false |
| AC-FM-017 (commit mismatch) | PASS | `TestRevisionMatchCommitMismatchIsFalse` | false |
| AC-FM-018 (`scope != repo`) | PASS | `TestRevisionMatchNonRepoScopeIsFalse` | false |
| AC-FM-019a (dirty tree, both halves) | PASS | `TestRevisionMatchDirtyTreeWithWorkingTreeExcluded` | dirty→false AND clean→true in one test |
| AC-FM-019b (`findings.jsonl`, both halves) | PASS | `TestRevisionMatchFindingsCompleteness` | absent→false AND zero-line→true in one test |
| AC-FM-020a (disclosure) | PASS | `grep -c 'scanned_commit' <Q>` = 2 (baseline 0); `grep -c 'inherited from the factory verify stage' <Q>` = 1 (baseline 0) | both positive controls fired |
| AC-FM-020b (manifest audit exempt) | PASS | `grep -c 'Step 0.55.1'` = 7 (baseline 1); `grep -c 'dependency manifest audit'` = 1 (baseline 0); `grep -c 'run unconditionally\|runs unconditionally'` = 1 (baseline 0); `grep -c 'regardless of whether manifest files changed'` = 1 (**unchanged from baseline 1** — the pre-existing always-on sentence survives verbatim at line 140) | all four |
| AC-FM-020c (rung allow-list) | PASS | `grep -c 'PRIMARY'` = 1, `'FALLBACK'` = 1, `'DEGRADED'` = 2 (all baseline 0); prose states suppression requires a **recorded** `PRIMARY` or `FALLBACK` and explicitly rejects the "not `DEGRADED`" deny-list phrasing. Go table `TestSuppressStep0551RungAllowList` asserts all three AC rows (`PRIMARY`→true positive control, `DEGRADED`→false, `""`→false) plus `nil`, an unrecognized rung, and a failing predicate | PASS |
| AC-FM-020d (predicate consumed) | PASS | `grep -c 'git rev-parse HEAD'` = 1, `'git status --porcelain'` = 1, `'revision-match predicate'` = 1 (all baseline 0) | all three ≥ 1 |
| AC-FM-012 (gate-token delta) | PASS | `grep -o '<5-token pattern>' <Q> \| wc -l` → `2`; baseline 2 + planned `N`=0 + planned `A`=0 = **2**. Exact match, no excess, escape valve not needed | `2` |

#### E2 — cross-platform build

```
$ go build ./...                           → EXIT_0_native
$ GOOS=windows GOARCH=amd64 go build ./... → EXIT_0_windows
```

#### E3 — coverage

```
$ go test -cover ./internal/factory/...
ok  	github.com/modu-ai/moai-adk/internal/factory	0.361s	coverage: 90.9% of statements
```

Target 85% met; the 88.2% M3 baseline is not regressed (+2.7pp).

#### E4 — subagent boundary

```
$ grep -rn 'AskUserQuestion\|mcp__askuser' internal/factory/
GREP_EXIT=1
```

Exit 1 with no output — no matches, including in `_test.go`.

#### E5 — lint

```
$ golangci-lint run --timeout=2m
0 issues.
```

No NEW findings; the clean baseline is preserved.

#### Additional verification

| Check | Command | Observed |
|---|---|---|
| Mirror-parity + neutrality guards | `go test ./internal/template/...` | `ok github.com/modu-ai/moai-adk/internal/template 47.336s` |
| Bounded mirror neutrality grep | `grep -rn 'internal/factory\|internal/cli' <6 mirrors>` | `MIRROR_NEUTRALITY_GREP_EXIT=1` — still 0 matches |
| Edited-mirror token leak | `grep -n 'SPEC-\|REQ-FM\|AC-FM\|internal/factory' <mirror Q>` | `TOKEN_LEAK_EXIT=1` — no matches |
| `goal-directive.md` mirror parity | `cmp <op> <mirror>` | `CMP_EXIT=0` |
| Template rebuild | `make build` | succeeded; `catalog.yaml updated successfully (12403 bytes)` then `go build -o bin/moai`. **`catalog.yaml` content did not change** — it is absent from `git status --porcelain`, because the regenerated hashes are identical (the edited workflow file is not a catalog-hashed skill entry) |
| SPEC lint | `moai spec lint .moai/specs/SPEC-FACTORY-MODE-001/spec.md` | `✓ No findings — all SPEC documents are valid` |
| Working-tree hygiene | `git status --porcelain` | exactly the two edited `quality-gates-quality.md` files plus the two new `internal/factory/` files — no `.moai/state/`, `.moai/harness/`, `.moai/cache/`, or `.moai/logs/` path touched |

#### Interpretation note — `Matches` signature vs the AC wording

`plan.md` §F fixes the pure predicate's signature as `Matches(rev *Revision, headSHA string, treeDirty bool) bool`, which takes an already-decoded revision and therefore cannot observe `findings.jsonl` or a missing directory. `acceptance.md` AC-FM-014 / AC-FM-015 / AC-FM-019b, meanwhile, describe fixture **directories** and say "when `Matches` is called". The two are not jointly satisfiable by one function.

Both were honored rather than one being chosen: `Matches` ships with the plan-mandated signature exactly, and `RevisionMatch(deepScanDir, headSHA, treeDirty)` is the directory-level composition that performs the existence check, the `findings.jsonl` completeness check, `LoadRevision`, and then `Matches`. The directory-level ACs assert against `RevisionMatch`; `Matches` additionally carries its own nil-revision test. The semantics every AC specifies are preserved verbatim — only the entry-point name differs from the AC's prose, and the doctrine's "revision-match predicate" wording (AC-FM-020d) names the composition rather than either function, so no grep is affected. This is recorded as an interpretation, not a silent choice.

#### Gaps (explicitly NOT observed in M4)

- `go test ./...` (the full suite) was NOT run. `./internal/factory/...` and `./internal/template/...` were run; `./internal/cli/...` and `./internal/config/...` were NOT re-run, because M4 touches neither. The full suite is M6's obligation.
- `internal/cli` coverage was NOT measured — M4 adds no `internal/cli` code; that 90% target is M5's.
- **Nothing in production calls `RevisionMatch` or `SuppressStep0551`.** This is AP-11 in its exact shape: the Go suite is green whether or not sync Phase 8 ever consults the predicate. The only evidence that the gate is wired is AC-FM-020d's doctrine grep, which verifies the *doctrine* names the call and both derived inputs — it does not verify any runtime invocation, because the consumer is an orchestrator-read workflow file, not Go code. No Go-level test of the consumption exists or can exist under this design.
- No `revision.json` **writer** was implemented or tested. Per `plan.md` §H the producer stays with the `--deep` workflow agents; nothing in this SPEC writes the artifact, so the reader has never been exercised against a real producer's output — only against `t.TempDir()` fixtures this milestone authored.
- The unreadable-file test SKIPS on Windows and when running as root. It was observed here on POSIX-as-non-root.
- The `treeDirty` derivation itself (the `git status --porcelain` exclusion set) is NOT implemented in Go — it is stated in doctrine as an orchestrator obligation and passed into `RevisionMatch` as a bare boolean. Nothing verifies the orchestrator derives it correctly.
- AC-FM-025's full template-neutrality judgement was not re-run beyond the bounded grep and the token-leak grep above; the CI guard `go test ./internal/template/...` passing is the mechanical evidence offered.

#### Residual risk

- The mirror was first updated by a blind `cp` of the operational file, which clobbered a line the mirror carried in neutrality-stripped form (the Phase 9 P1/P2 sentence, which omits an internal SPEC ID the operational file names). The `cp` was reverted via `git restore` and replaced with a surgical `Edit`, and the token-leak grep above confirms the mirror is clean. The residual risk is general rather than specific: **the two files are not byte-identical by design**, so any future full-file copy in either direction silently reintroduces internal tokens or drops operational content, and only the `internal/template` CI guard stands between that and a merge.
- `SuppressStep0551` takes `revisionMatched bool` rather than calling `RevisionMatch` itself. That keeps it pure and independently testable, but it also means a caller can pass `true` from any source — the allow-list constrains the rung, not the provenance of the match.
- `findingsComplete` uses a 4 MiB scanner buffer ceiling. A `findings.jsonl` line longer than that yields a scanner error, which the function reports as incomplete — FALSE, so the safe direction — but it would be reported as "aborted scan" rather than "line too long", and nothing distinguishes the two for an operator reading the sync report.
- Blank lines in `findings.jsonl` are tolerated as formatting rather than treated as records. If a producer ever emits a semantically meaningful empty line, this reader silently ignores it.
- The doctrine section is prose an orchestrator reads, not a mechanism that executes. A future edit that softens "the gate defaults to RUN" into a permissive phrasing would pass every grep in AC-FM-020 while inverting the fail-safe direction — the greps assert token presence, not semantic direction.

### M5 — Launcher entry

Baseline attribution for every row below: **this run, this tree, HEAD `003f9555e`**, branch `feat/factory-mode`. `cycle_type=tdd` — the verbatim pre-GREEN failing output is recorded under "RED evidence" and persisted at `.moai/state/verify/factory-m5/red.txt`.

#### Pre-flight baselines (re-measured before the first M5 edit)

| Check | Command | Observed |
|---|---|---|
| build | `go build ./...` | `BUILD_OK` (exit 0) |
| lint | `golangci-lint run --timeout=3m` | `0 issues.` |
| working tree | `git status --porcelain` | empty |
| AC-FM-006 pre-change | `grep -rn '"-f"' internal/cli/cc.go internal/cli/glm.go internal/cli/cg.go` | no match (exit 1) |
| factory.go absent | `ls internal/cli/factory.go` | `No such file or directory` |
| inject pre-change | `grep -n 'os.Getenv(config.EnvMoaiFactory)' internal/cli/launcher_blockcap_infinite.go` | no match (exit 1) |

#### RED evidence (E8 — verbatim, captured BEFORE any implementation)

Command: `go test ./internal/cli/ -run 'TestParseFactoryFlag|TestCC_Factory|TestCG_Factory|TestCG_WithoutFactory|TestGLM_FactoryFlagParity|TestACFM022a|TestACFM023c' -count=1`

The RED was taken against a stub `internal/cli/factory.go` whose three functions returned zero values, with no launcher wiring and no inject change — so the failures are behavioural assertions, not build errors:

```
--- FAIL: TestParseFactoryFlag_LongFormWithoutSpec (0.00s)
    cc_test.go:292: --factory must enable Factory Mode
    cc_test.go:298: --factory must be stripped from the forwarded args, got [--factory -b]
--- FAIL: TestParseFactoryFlag_ShortFormWithSpec (0.00s)
    cc_test.go:308: AC-FM-002: -f must enable Factory Mode
    cc_test.go:311: AC-FM-002: expected spec "SPEC-PLACEHOLDER", got ""
    cc_test.go:314: AC-FM-002: both tokens must be stripped, got [-f SPEC-PLACEHOLDER --print]
--- FAIL: TestParseFactoryFlag_SpecIsNotStolenFromAFlag (0.00s)
    cc_test.go:339: --factory must enable Factory Mode
    cc_test.go:345: the following flag must survive, got [--factory --print]
--- FAIL: TestCC_FactoryFlagStrippedBeforeLaunch (0.00s)
    cc_test.go:382: AC-FM-001: factory token must not reach the launcher, got [--factory SPEC-PLACEHOLDER]
    cc_test.go:386: AC-FM-001: MOAI_FACTORY must be set at launch, got ""
    cc_test.go:389: AC-FM-001: MOAI_FACTORY_SPEC must carry the identifier at launch, got ""
--- FAIL: TestCC_FactoryWritesStateRecord (0.00s)
    cc_test.go:430: AC-FM-023a: factory record not readable: read factory record: open .../.moai/state/factory/factory-record-session.json: no such file or directory
--- FAIL: TestCG_FactoryFlagRejected (0.00s)
    --- FAIL: TestCG_FactoryFlagRejected/--factory (0.00s)
        cg_test.go:209: AC-FM-004: runCG(--factory) must return an error
    --- FAIL: TestCG_FactoryFlagRejected/-f (0.00s)
        cg_test.go:209: AC-FM-004: runCG(-f) must return an error
--- FAIL: TestGLM_FactoryFlagParity (0.00s)
    glm_test.go:677: AC-FM-005: factory token must not reach the launcher, got [--factory SPEC-PLACEHOLDER]
    glm_test.go:681: AC-FM-005: MOAI_FACTORY must be set at launch, got ""
    glm_test.go:684: AC-FM-005: MOAI_FACTORY_SPEC must carry the identifier at launch, got ""
--- FAIL: TestACFM022a_FactoryRaisesBlockCapUnconditionally (0.00s)
    launcher_blockcap_infinite_test.go:137: AC-FM-022a: expected "CLAUDE_CODE_STOP_HOOK_BLOCK_CAP=200" in the launch env, got [PATH=/usr/bin HOME=/tmp]
--- FAIL: TestACFM022a_FactoryCapReplacesPreexistingEntry (0.00s)
    launcher_blockcap_infinite_test.go:155: AC-FM-022a: stale cap survived: "CLAUDE_CODE_STOP_HOOK_BLOCK_CAP=8"
FAIL
FAIL	github.com/modu-ai/moai-adk/internal/cli	1.101s
```

Three of the fifteen M5 tests passed at RED — `TestParseFactoryFlag_PassThroughBoundary`, `TestCG_WithoutFactoryFlagStillLaunches`, and `TestACFM023c_FactoryEnvReachesChildEnvironment`. This is recorded rather than hidden: each is a **negative control or a pre-existing-property assertion**, so passing before the implementation is the correct outcome for them. `TestCC_FactoryEnvMutationIsRestored` also passed at RED for the same structural reason (with no mutation there is nothing to leak); its falsification control below is what gives it teeth.

#### E1 — AC PASS/FAIL matrix

Command for every test row: `go test ./internal/cli/ -run '<name>' -count=1 -v`; the consolidated run is `go test ./internal/cli/ -run 'TestParseFactoryFlag|TestCC_Factory|TestEnterFactoryMode|TestCG_|TestGLM_Factory|TestACFM|TestAC003_' -count=1`.

| AC | Status | Verification | Actual output |
|---|---|---|---|
| AC-FM-001 (token stripped, signal published) | PASS | `TestCC_FactoryFlagStrippedBeforeLaunch` | `--- PASS: TestCC_FactoryFlagStrippedBeforeLaunch (0.00s)` |
| AC-FM-002 (`-f SPEC-PLACEHOLDER` parse) | PASS | `TestParseFactoryFlag_ShortFormWithSpec` | `--- PASS: TestParseFactoryFlag_ShortFormWithSpec (0.00s)` |
| AC-FM-003 (`--` pass-through boundary) | PASS | `TestParseFactoryFlag_PassThroughBoundary` | `--- PASS: TestParseFactoryFlag_PassThroughBoundary (0.00s)` |
| AC-FM-004 (`cg` sentinel rejection, seam never invoked) | PASS | `TestCG_FactoryFlagRejected` (`--factory` + `-f` sub-tests) | `--- PASS: TestCG_FactoryFlagRejected (0.00s)` |
| AC-FM-005 (`glm` parity, mode `glm`) | PASS | `TestGLM_FactoryFlagParity` | `--- PASS: TestGLM_FactoryFlagParity (0.00s)` |
| AC-FM-006 (pre-change `-f` unbound; post-change positive control) | PASS | pre: `grep -rn '"-f"' internal/cli/{cc,glm,cg}.go` → no match; post: `grep -c '"-f"' internal/cli/factory.go` | pre `exit=1`; post `1` |
| AC-FM-022a (unconditional factory inject + negative control) | PASS | `TestACFM022a_FactoryRaisesBlockCapUnconditionally`, `TestACFM022a_FactoryCapReplacesPreexistingEntry` | `--- PASS` on both |
| AC-FM-023a (state record; unwritable dir fail-open) | PASS-WITH-INTERPRETATION | `TestCC_FactoryWritesStateRecord` | `--- PASS: TestCC_FactoryWritesStateRecord (0.01s)` — see the interpretation note below |
| AC-FM-023b (env constants declared) | PASS (M3 deliverable, re-confirmed) | `grep -c 'EnvMoaiFactory = "MOAI_FACTORY"' internal/config/envkeys.go` / `...FactorySpec...` | `1` / `1` |
| AC-FM-023c (env reaches the child environment) | PASS | `TestACFM023c_FactoryEnvReachesChildEnvironment` | `--- PASS: TestACFM023c_FactoryEnvReachesChildEnvironment (0.00s)` |
| AC-FM-023d (mutation restored, success AND error path) | PASS | `TestCC_FactoryEnvMutationIsRestored` (+ `TestEnterFactoryMode_RestoresPriorValue`) | `--- PASS` on both, including the `success path` and `error path` sub-tests |
| Non-regression: SPEC-INFINITE-GOAL-001 AC-003 | PASS | `TestAC003_LauncherInjectsRaisedBlockCapForInfiniteGoal` | `--- PASS: TestAC003_LauncherInjectsRaisedBlockCapForInfiniteGoal (0.00s)` |

**AC-FM-023a interpretation (recorded, not silently resolved).** The criterion lists `verify_rung` among the fields the launch-time record "carries". The M3 record contract (`design.md` §7, and the `*Rung` + `omitempty` decision recorded in this file's M3 block) deliberately leaves that field **unrecorded at launch** so a later reader can distinguish "never written" from "written blank" — with `omitempty`, an unrecorded rung marshals away entirely. The test therefore asserts `session_id` / `spec_id` / `backend` / `entered_at` are populated and that `VerifyRung` is `nil`, i.e. the field exists in the schema and is legitimately unrecorded at this point in the chain. Reading the criterion as requiring a `verify_rung` JSON key at launch would contradict M3's own resolution of AP-13. This is flagged for the M6 verification pass rather than decided unilaterally.

#### E2 — Cross-platform build

```
$ go build ./...                           → host-build-exit=0
$ GOOS=windows GOARCH=amd64 go build ./... → windows-build-exit=0
```

#### E3 — Coverage

```
$ go test -cover ./internal/cli/... ./internal/factory/...
ok  github.com/modu-ai/moai-adk/internal/cli       270.855s  coverage: 76.6% of statements
ok  github.com/modu-ai/moai-adk/internal/factory   (cached)  coverage: 90.9% of statements
```

`internal/factory` is at **90.9%**, above its 85% target and unchanged by M5 (which adds no code there). `internal/cli` reports **76.6%**, below the 90% target stated in `plan.md` §D — reported honestly rather than framed as a pass. The package is large and pre-existing, and the M5 change cannot have lowered it: every function M5 adds is fully covered.

```
$ go tool cover -func=<M5 subset profile> | grep 'cli/factory.go'
internal/cli/factory.go:50:   parseFactoryFlag       100.0%
internal/cli/factory.go:89:   enterFactoryMode       100.0%
internal/cli/factory.go:106:  captureEnvState        100.0%
internal/cli/factory.go:125:  recordFactorySession   100.0%
internal/cli/factory.go:136:  rejectFactoryOnCG      100.0%
internal/cli/launcher_blockcap_infinite.go:39:  injectStopHookBlockCapForGoal  100.0%
internal/cli/launcher_blockcap_infinite.go:65:  setStopHookBlockCap            100.0%
```

#### E4 — Subagent boundary

```
$ grep -rn 'AskUserQuestion\|mcp__askuser' internal/cli/factory.go internal/factory/ | grep -v '_test.go' | grep -v '// '
grep-exit=1   (no matches)
```

#### E5 — Lint

```
$ golangci-lint run --timeout=3m
0 issues.
```

Identical to the pre-flight baseline — zero NEW findings.

#### Falsification control (AC-FM-023d)

The `defer enterFactoryMode(specID)()` in `runCC` was temporarily replaced with `_ = enterFactoryMode(specID)` (restore discarded), the M5 suite re-run, and the edit reverted immediately (`grep -n FALSIFICATION internal/cli/cc.go` → exit 1, no residue). Observed:

```
--- FAIL: TestCC_FactoryEnvMutationIsRestored/success_path
    AC-FM-023d: MOAI_FACTORY presence = true after runCC, want false
--- FAIL: TestAC003_LauncherInjectsRaisedBlockCapForInfiniteGoal
    AC-003 backward compat: block cap injected with no armed goal: "CLAUDE_CODE_STOP_HOOK_BLOCK_CAP=200"
    AC-003: block cap injected for a FINITE (MaxTurns>0) goal: "CLAUDE_CODE_STOP_HOOK_BLOCK_CAP=200"
--- FAIL: TestACFM022a_FactoryRaisesBlockCapUnconditionally
    AC-FM-022a negative control: with MOAI_FACTORY unset and no armed goal the env must be unchanged, got [... CLAUDE_CODE_STOP_HOOK_BLOCK_CAP=200]
```

The leak is **worse than the acceptance criterion predicted**: `acceptance.md` states AC-FM-023d would fail "while every other AC in §B.5 still passes". In practice the unrestored `MOAI_FACTORY` also broke the pre-existing SPEC-INFINITE-GOAL-001 AC-003 test and AC-FM-022a's own negative control — the cross-test contamination AP-14 names, observed directly rather than argued.

#### Other verifications

- `go test ./internal/cli/ ./internal/factory/... ./internal/config/...` → all `ok` (`internal/cli` 254.755s, `internal/factory` 2.249s, `internal/config` 4.620s, `internal/config/toolpolicy` 3.152s).
- `go test -race` on the M5 subset → `ok github.com/modu-ai/moai-adk/internal/cli 4.831s`. The env-mutation tests are non-parallel by construction (`t.Setenv`).
- `cmp .claude/rules/moai/workflow/goal-directive.md internal/template/templates/.claude/rules/moai/workflow/goal-directive.md` → exit 0 (mirror parity intact, untouched by M5).
- `git diff --name-only -- .claude/ internal/template/templates/ | wc -l` → `0`. M5 edits no `.claude/` doctrine, so **no mirror obligation and no `make build` applies** — `make build` regenerates the embedded template FS, and no template input changed.
- `moai spec lint .moai/specs/SPEC-FACTORY-MODE-001/spec.md` → `✓ No findings — all SPEC documents are valid`.

#### Design notes

- **Parser placement.** `parseFactoryFlag` runs after `parseProfileFlag` and before `resolveWorktreeL2Path` / `normalizeWorktreeFlag`, per `design.md` §5: after `--spawn` stripping so a spawned session carries the token through, and before worktree handling so a factory token can never be consumed as a `-w` value.
- **Optional SPEC identifier (REQ-FM-005).** A following token is consumed as the identifier only when it is neither `--` nor flag-shaped, so `moai cc --factory --print` yields `spec == ""` with `--print` intact. Absence is a valid parse, not an error.
- **`setStopHookBlockCap` extraction.** The replace-in-place loop was extracted from the goal branch so both branches share one setter rather than duplicating fifteen lines. The goal branch's condition and semantics are unchanged, and `TestAC003_LauncherInjectsRaisedBlockCapForInfiniteGoal` (a pre-existing test) passes unmodified — the "preserved verbatim" obligation is honored at the behavioural level, with the shared setter noted here as a deliberate REFACTOR-step change to the file's text.
- **`cg` rejection placement.** `rejectFactoryOnCG` runs before `stripSpawnFlag`, so `moai cg --factory --spawn` fails in the operator's own terminal rather than inside a spawned tmux window they are not watching.

#### Gaps (explicitly NOT observed in M5)

- **`internal/cli`'s pre-change package coverage was NOT measured**, so the 76.6% figure above is reported as an absolute, not as a delta. Measuring the baseline would have required stashing or a second worktree; `git stash` is forbidden by the branch-state guard, and a second 270-second run was judged not to earn its cost. The claim that coverage did not decrease rests on the per-function evidence (every added function ≥ 100% except none) rather than on a measured before/after pair — it is an inference, not an observation.
- `go test ./...` (the full repository suite) was NOT run. `./internal/cli/...`, `./internal/factory/...`, and `./internal/config/...` were. The full suite is M6's obligation.
- **Nothing consumes `MOAI_FACTORY_SPEC` in Go.** M5 publishes it into the child environment (AC-FM-023c) but no code in this repository reads it; its only consumer is the orchestrator inside the launched session. A rename would break the chain with no Go test failing.
- **The record's three orchestrator-written fields are never written by any code in this SPEC.** `DeepScanDir`, `VerifyRung`, and `VerifyReentries` are filled in by the orchestrator as the chain progresses; M5 writes only the four launch-time fields.
- **No end-to-end launch was exercised.** Every test drives the `unifiedLaunchFunc` seam; the real `syscall.Exec` path — and therefore the actual delivery of the raised cap to a live Claude Code process — was not run, per the project's prohibition on running the real launcher flow in this development checkout.
- The unwritable-state-directory fail-open case relies on POSIX directory permissions and was observed on POSIX as a non-root user. It is not skipped on Windows or as root, where `os.Chmod(dir, 0o500)` does not block writes — the fail-open control assertion would then report a spurious failure. This is a latent portability gap in the test, not in the implementation.

#### Residual risk

- **The raised block cap is session-wide and applied before Implementation Kickoff Approval.** A user who declines the gate still carries a 200-block ceiling for the rest of the session. `design.md` §6 accepts this and requires `factory.md` to document it (an M1 deliverable); M5 implements the inject, not the disclosure.
- **The factory branch keys on `os.Getenv(config.EnvMoaiFactory) != ""`**, so any value — including `0` or `false` — enables the raise. Only the launcher sets the variable today, but an operator who exports `MOAI_FACTORY=0` expecting it to disable the mode would get the opposite.
- **The `defer` never runs on the production success path.** `syscall.Exec` replaces the process, so the restore executes only on the error path and inside the test binary. That is precisely where it is needed, but it means the production success path is protected by process replacement rather than by the restore — and on Windows, where the launcher spawns a child and exits instead, the restore's coverage of the parent is equally moot.
- **`recordFactorySession` silently no-ops when the session id cannot be resolved.** The side-channel file is written by the SessionStart hook of the *outgoing* session, so a launch from a shell with no prior MoAI session produces no record. The chain then runs with no stored state — safe by design (the dedup predicate reads absence as "run the check"), but an operator debugging a missing record has no signal that anything was skipped.
- **`parseFactoryFlag` accepts the token more than once.** `moai cc --factory A --factory B` yields `spec == "B"` with no warning; last-wins is an accident of the loop rather than a stated contract.

### M5.1 — Regression repair: relocate the verify exit gate out of the `run.md` entry router

Scope-limited fix milestone. No new feature; no Go source touched. Repairs a regression this SPEC introduced in M1.

#### Claim

M1 inlined the `## Verify Exit Gate (factory contract)` block directly into the `run.md` entry router, pushing it from exactly 200 LOC to 234 and breaking `TestEntryRouterLOCCeiling`. The block has been relocated verbatim into `.claude/skills/moai/workflows/run/mode-orchestration.md`; `run.md` is back to exactly 200 LOC; the guard passes; all fourteen doctrine literals survived the move; both mirrors were updated surgically.

#### Evidence

(a) command / (b) verbatim output / (c) attribution — all `(this run, this tree, HEAD b56b64353 + working tree)`.

Guard, before:

```
$ go test -count=1 ./internal/skills/
--- FAIL: TestEntryRouterLOCCeiling (0.00s)
    workflow_split_test.go:154: ENTRY_ROUTER_LOC_VIOLATION: run.md has 234 LOC (ceiling: 200)
FAIL
FAIL	github.com/modu-ai/moai-adk/internal/skills	0.389s
FAIL
```

Guard, after:

```
$ go test -count=1 ./internal/skills/
ok  	github.com/modu-ai/moai-adk/internal/skills	0.227s
```

LOC:

```
$ wc -l < .claude/skills/moai/workflows/run.md
     200
$ wc -l < .claude/skills/moai/workflows/run/mode-orchestration.md
     156
$ wc -l internal/template/templates/.claude/skills/moai/workflows/run.md internal/template/templates/.claude/skills/moai/workflows/run/mode-orchestration.md
     200 internal/template/templates/.claude/skills/moai/workflows/run.md
     154 internal/template/templates/.claude/skills/moai/workflows/run/mode-orchestration.md
```

`run.md` is at the ceiling exactly, not under it — see Residual risk.

Literal preservation. Occurrence counts at the NEW location, and the same alternation against `run.md` (empty — a clean move, not a copy):

```
$ grep -oE '<14-literal alternation>' .claude/skills/moai/workflows/run/mode-orchestration.md | sort | uniq -c
   1 /moai review --security --deep --repo
   1 at most two verify re-entries
   1 Baseline-attribution
   5 DEGRADED
   2 exit gate of run-phase
   1 governs routing
   1 governs suppression
   2 HALT
   1 inherited sync-phase evidence
   2 no readable result
   1 orthogonal
   1 re-enter run-phase scoped to the changed surface
   4 readable result
   1 shall not proceed to sync

$ grep -oE '<same alternation>' .claude/skills/moai/workflows/run.md | sort | uniq -c
(no output)
```

Line counts in AC-grep semantics (`grep -c` counts lines, not occurrences), plus the AC-FM-024a D3 negative control:

```
readable result (lines): 5
no readable result (lines): 2
HALT (lines): 2
DEGRADED (lines): 3
no confirmed findings (lines): 3
D3 negative control — 'no confirmed findings' lines lacking 'readable': 0
```

Every AC-FM-008 / 009 / 010 / 011 / 013 / 024a / 024b threshold is met at the new path.

Gate-token count under the AC-FM-012 pattern:

```
run.md:                 33
mode-orchestration.md:   2
pre-SPEC baseline run.md (7171880a9): 33
```

`run.md` is back to its pre-M1 baseline of 33 exactly. The new file's 2 = the one `Implementation Kickoff Approval` mention carried in by the moved ordering sentence, plus one pre-existing `AskUserQuestion` in its Error Flow scenario.

Build, lint, mirror guards:

```
$ go build ./...                                 → exit 0
$ GOOS=windows GOARCH=amd64 go build ./...       → exit 0
$ golangci-lint run --timeout=3m                 → 0 issues.
$ go test -count=1 ./internal/template/...
ok  	github.com/modu-ai/moai-adk/internal/template	20.742s
?   	github.com/modu-ai/moai-adk/internal/template/scripts	[no test files]
$ cmp .claude/rules/moai/workflow/goal-directive.md internal/template/templates/.claude/rules/moai/workflow/goal-directive.md
(identical)
$ moai spec lint .moai/specs/SPEC-FACTORY-MODE-001/spec.md
✓ No findings — all SPEC documents are valid
$ make build                                     → exit 0; catalog.yaml regenerated byte-identical (absent from git status)
```

Bounded mirror neutrality grep over the two changed mirror files returns exactly one match — `updated: "2026-02-23"` in the `run.md` frontmatter, byte-identical at baseline `7171880a9` and therefore pre-existing, not introduced here. No SPEC identifier, requirement or acceptance token, commit SHA, or internal Go package path appears.

#### Baseline-attribution

- Pre-SPEC `run.md` LOC: `git show 7171880a9:.claude/skills/moai/workflows/run.md | wc -l` → `200`.
- M1 delta: `git diff --numstat 7171880a9..HEAD -- .claude/skills/moai/workflows/run.md` → `34	0`.
- The arithmetic is exact and binding: 234 − 34 = 200, and the baseline was already *at* the ceiling. `run.md` could therefore absorb no pointer line at all — a one-line "see also" would have re-broken the guard at 201. That is why the block was moved whole with no residual stub, and why discoverability was carried by an in-place edit rather than an added line.

#### Choice of new home, and why

`run/mode-orchestration.md`, not a new `run/verify-exit-gate.md`. Three reasons, in order of weight:

1. **Discoverability was the deciding constraint.** `run.md`'s Phase Routing Table maps each phase group to one of the four existing sub-skill files. A fifth file would need a fifth row, and no row can be added at 200/200. The only honest alternative — naming the new file inside a different row's description — would point readers at the wrong file. Appending to an existing file let the *correct* row's description cell be extended in place, net-zero: `… completion criteria, verify exit gate (factory contract), test scenarios`. Verified net-zero by `wc -l` (200 before and after the cell edit), not assumed.
2. **Semantic fit.** That file already owns Completion Criteria — the conditions under which run-phase ends. An exit gate of run-phase belongs beside them, and the block was placed immediately after that section.
3. **Headroom and mirror cost.** 122 → 156 LOC against the 600 sub-skill ceiling (`TestSubSkillLOCCeiling`), and no new mirror file, so `TestTemplateMirrorParity`'s path-set equality is untouched rather than merely re-satisfied.

The sub-skill's frontmatter `description` was also extended in place to name the gate, so the routing table and the file's own metadata agree.

#### Gaps — explicitly not verified

- **The acceptance criteria still point at `run.md`.** AC-FM-008 / 009 / 010 / 011 / 013 / 024a / 024b all grep `.claude/skills/moai/workflows/run.md` literally, and every one of those greps now returns 0. Re-pointing them is manager-spec's job and was deliberately not done here (constraint 6). The literal table above is the evidence that re-pointing is a path substitution and nothing more. **Until that re-point lands, those seven criteria read as FAIL.**
- **Only `./internal/skills/` and `./internal/template/...` were run, not the full `go test ./...`.** The change is markdown-only and no Go source was touched, but a cross-cutting guard living outside those two packages would not have been observed.
- **No renderer or live-session check.** The moved block's markdown was not rendered, and no session was started to confirm the router actually reaches the relocated section at runtime. The routing-table cell edit is prose, and prose discoverability is not mechanically testable.
- **Heading levels were preserved verbatim**, so the block enters a `#`-level file at `##`. Verbatim preservation was ranked above structural tidiness per the milestone constraint; no test asserts heading level.

#### Residual risk

- **`run.md` sits at exactly 200/200 with zero headroom.** This is the same condition that let M1 break the guard in the first place. The next edit that adds a single line to `run.md` — including a well-intentioned pointer to the relocated gate — re-breaks `TestEntryRouterLOCCeiling`. The structural fix (trimming `run.md` to create headroom) is out of this milestone's scope and remains open.
- **The doctrine is now one hop further from its entry point.** A reader of `run.md` reaches the verify exit gate only via the routing-table description cell. That is weaker discoverability than the inline form M1 chose, and it is the price the LOC ceiling extracts.
- **Two copies of the block now exist** (live tree and template mirror), as for every mirrored file. They were edited surgically and separately rather than copied, so a future edit to one that skips the other will drift silently until `TestTemplateMirrorParity` — which compares path sets, not contents — fails to catch it.

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
