# SPEC Review Report: SPEC-AGENTS-MD-CANON-001

Iteration: 6/6 (ceiling)
Document: frontmatter `version: "0.3.6"` — tree `58f0bdd43`
Verdict: **PASS**
Overall Score: **0.925** (Tier L threshold 0.85)
Score trajectory: iter4 0.87 → iter5 0.90 → iter6 0.925 — no regression, no STOP signal.

Reasoning context ignored per M1 Context Isolation. The dispatcher's own verification was
re-executed rather than repeated on trust; where my figure differs from a cited one it is said so.

Freeze held: `git status --short .moai/specs` empty at open and at close; `git log
f10c8691f..58f0bdd43` is exactly two commits (the iter5 verdict record + the F1/F2 revision), and
`git diff f10c8691f 58f0bdd43` touches only `version:`/HISTORY, `REQ-AMC-010`, `AC-AMC-015`,
`AC-AMC-007`, plan.md M1 Arm B, and progress.md. No unannounced change.

---

## Must-Pass Results

- **[PASS] MP-1 REQ number consistency** — `REQ-AMC-001`…`REQ-AMC-018`, 18 entries, each appearing
  exactly once, zero-padded consistently (`grep -oE '^\*\*REQ-AMC-[0-9]+' spec.md | sort | uniq -c`
  → 18 rows, all count 1). ACs `AC-AMC-001`…`AC-AMC-024`, 24 unique, no gap. Both inside the Tier L
  ceiling of 25/25 (`spec-workflow.md` § SPEC Complexity Tier).
- **[PASS] MP-2 GEARS format compliance (requirement layer)** — judged against the 18 `REQ-AMC-*`
  entries in `spec.md`, not against any `AC-AMC-*`. Each carries an explicit modality label and the
  matching pattern. The only entry changed this iteration is `REQ-AMC-010` (spec.md:216-220),
  Unwanted: *"The redistribution **shall not** change Claude Code rule-loading semantics, hook
  wiring, or any existing test's expected behavior — **except** that an assertion whose expected
  value is the *cardinality* of the always-loaded surface … is exempt."* A scoped exception on a
  `shall not` is a narrowing of the prohibition's scope, not a departure from the pattern.
- **[PASS] MP-3 YAML frontmatter validity** — all 12 canonical fields present with correct types
  (spec.md:1-15): `id`, `title`, `version: "0.3.6"` (quoted), `status: draft`, `created`/`updated`
  `2026-08-22` (ISO), `author`, `priority: P1`, `phase`, `module`, `lifecycle: spec-anchored`,
  `tags` (comma-separated string), plus optional `tier: L`. No rejected snake_case alias present.
  `moai spec lint .moai/specs/SPEC-AGENTS-MD-CANON-001/spec.md` → `✓ No findings`; repo-wide
  `moai spec lint` exit 0.
- **[N/A] MP-4 language neutrality** — this SPEC is scoped to one repository's own Go guard and its
  template mirror; it is not a multi-language tooling SPEC. Template-side neutrality is separately
  bound by `REQ-AMC-016` / `AC-AMC-022`.
- **[PASS] MP-5 D7 cross-SPEC reconciliation** — the only external SPEC reference in `spec.md` is
  `SPEC-ALWAYS-LOADED-DIET-001`; `grep '^status:'` on it → `completed`, which is not in
  {retired, superseded, archived}. No BLOCKING finding.
- **[PASS] MP-6 D8 cross-platform discipline** — `grep -c syscall spec.md` → `0`. Auto-pass.
- **[PASS] MP-7 clarification gate** — `grep -rn 'NEEDS CLARIFICATION' .moai/specs/SPEC-AGENTS-MD-CANON-001/`
  → 0 matches across all six artifacts.

---

## Category Scores

| Dimension | Score | Band | Evidence |
|---|---|---|---|
| Clarity | 0.90 | 0.75–1.0 | The F1 carve-out states its scope three ways and enumerates its two assertions in a before/after table (`acceptance.md:117-138`). Residual ambiguity: "`@`-imported **contract document**" (`REQ-AMC-008`, spec.md:205-206) is never defined — see O1. |
| Completeness | 0.90 | 0.75–1.0 | All required sections present; five `### Out of Scope — <topic>` H3 blocks with specific bullets (spec.md:513, 522, 527, 536, 541); Tier L 5-artifact set complete. Gap: the enumeration extension's *upper* ordering bound is unstated — see O3. |
| Testability | 0.90 | 0.75–1.0 | Most ACs cite an executable command with a decidable output (`AC-AMC-010`'s `git ls-files` re-executed here, exit 0, empty output — correct pre-run state; `AC-AMC-018`/`019` read a logged figure and a band). `AC-AMC-007` is conditional-Given and can pass vacuously — see O2. |
| Traceability | 1.00 | 1.0 | Every REQ maps to ≥1 AC and every AC to an existing REQ, checked pair by pair: 001/002→008,011 · 003→004 · 004→009 · 005→010 · 006→012 · 007,009→016(b) · 008→017 · 010→015 · 011→013,014 · 012→014 · 013→018,019,020 · 014→018 · 015→021 · 016→022 · 017→023 · 018→024. No orphan, no uncovered REQ. |

Aggregate = (0.90 + 0.90 + 0.90 + 1.00) / 4 = **0.925**.

---

## Judgment 1 — is the carve-out narrow enough to be safe and wide enough to be usable?

**Yes, on both counts. This is the finding that closes F1.**

I read `internal/config/token_budget_guard_test.go` in full (286 lines) and
`token_budget_guard.go` in full (210 lines), then swept the repository for every other consumer of
the enumeration. Adding a fourth fixed slot changes the value of exactly two assertions:

| Location | Assertion | Today | After |
|---|---|---|---|
| `token_budget_guard_test.go:129-131` | `wantTotal := wantRuleCount + 3` | `+ 3` | `+ 4` |
| `token_budget_guard_test.go:197-199` | `if len(surface) != 4` (temp tree) | `4` | `5` |

Both are named by the carve-out. Everything else in the file is unaffected, and I checked each
rather than assuming:

- `TestAlwaysLoadedTokenBudget` (:66) compares a token *sum*, not a cardinality; the fourth slot is
  hermetic (`measureAlwaysLoaded` skips an unreadable path), so a missing `AGENTS.md` contributes 0.
- `TestAlwaysLoadedTokenBudget_OverBudgetFails` (:101-107) derives its fixture size from
  `AlwaysLoadedTokenBudget*4 + 4096`, so it tracks the constant automatically and is indifferent to
  slot count.
- `TestMeasureAlwaysLoaded_WithMemory`'s lower bound (:203) and upper bound (:207) are token sums
  over a temp tree with no `AGENTS.md` — unchanged at 0 tokens added.
- The `paths:`-exclusion assertion (:136-141), the `MEMORY.md` head-cap assertions (:203-209,
  :246-255), `TestHasPathsRestriction`, `TestEstimateTokens`, `TestFindRepoRoot` — none reads the
  cardinality. The carve-out explicitly excludes the first three by name.
- Repo-wide sweep for other consumers: `grep -rn 'alwaysLoadedSurface\|measureAlwaysLoaded\|AlwaysLoadedTokenBudget\|MeasureAlwaysLoadedSection' --include='*.go' .`
  returns, outside this pair of files, only `internal/harness/curator/budget_test.go` (calls
  `MeasureAlwaysLoadedSection`, a per-section byte reader that never touches the enumeration) and
  `.moai/reports/t114/measure_tool.go` (a report-local replica, not part of the build). Neither
  asserts a surface count.

So the exemption's set is exactly the affected set: not one assertion short, not one too many. It
does not return the goalpost-moving hazard, because the two escapes an actor would actually reach
for — relaxing `AlwaysLoadedTokenBudget`, or dropping the `paths:` exclusion — are both outside it
and named as outside it.

**Premise re-verified independently** (not carried from iteration 5):
`go test ./internal/config/ -run 'TestAlwaysLoadedSurfaceEnumeration|TestAlwaysLoadedTokenBudget|TestMeasureAlwaysLoaded_WithMemory' -count=1`
→ `ok github.com/modu-ai/moai-adk/internal/config 1.308s`. Both counts hold at `+3` / `4` today, so
the break is attributable to this SPEC's change.

## Judgment 2 — does F2's qualifier bind?

**Partly. It binds the plan layer firmly and the verification layer only conditionally.**

`plan.md:104-112` carries it as a bolded imperative — *"Baseline the projection on the
integration-branch figure recorded at pre-flight (§C), not on this worktree's"* — with the
consequence stated in measured units (4,070 tokens, 37 % of the required cut). The artifact it
points at exists: `plan.md` §C checkbox 2 requires the integration-branch surface to be measured
with branch and ahead-count recorded, so the correct number is on hand before M1 runs. That is a
real binding, not decoration.

`AC-AMC-007` folds the qualifier into its Given clause (`acceptance.md:46-50`), which is the right
place for it. The residual weakness is structural rather than textual: the criterion fires only
*when an arm trips*. An actor who baselines on the worktree and consequently does **not** trip Arm B
leaves the Given unmet, and the criterion passes vacuously — which is the exact case the qualifier
exists to catch. Detection then falls to `AC-AMC-018` at M5, i.e. after M2-M4's least-reversible
work, the late discovery Arm B was added to prevent. Recorded as O2; not blocking, because the plan
instruction is unambiguous, the correct figure is a pre-flight deliverable, and the failure is
caught before close rather than shipped.

## Judgment 3 — is the SPEC implementable end to end?

I walked M1 → M6 as a run-phase actor and looked specifically for a second F1-shaped pair (two
requirements neither of which can be satisfied without failing the other). **I found one near-miss,
not a second contradiction.** It has an exit, so it does not block — but it is the single most
useful thing this report can hand run phase, and it is stated as O3 below.

The near-miss: `AC-AMC-015` ("full existing test suite green") sits under **§D — M3**, not M4. By
M3, `AGENTS.md` exists (M2 authored it) and is `@`-imported. If M5's enumeration extension has
already landed at that point, the always-loaded surface includes `AGENTS.md` while the M4 diet has
not yet happened. Measured on this tree:

```
no-paths: rules   202,621 B  (14 files — grep -rLE '^paths:' --include='*.md' .claude/rules/moai)
CLAUDE.md          20,523 B
moai.md            61,706 B
MEMORY.md               0 B  (absent at repo root)
                  ---------
total             284,850 B / 4 = 71,212 tokens   ← matches spec.md §A.1 exactly
```

Headroom to the 76,000 constant is 4,788 tokens = **19,152 B**. `AGENTS.md`'s ceiling is 24,576 B,
so any contract document larger than 19,152 B trips `TestAlwaysLoadedTokenBudget` at M3. On the
integration branch (75,282 tokens) the headroom is 718 tokens = **2,872 B**, so it trips
essentially for certain.

Why it is not a contradiction: a viable ordering exists. `REQ-AMC-008` fixes only a *lower* bound —
the extension must precede any measurement quoted as a ratchet basis (M4's running totals, M5's
achieved figure). It does not require the extension before M3. Land it after M3's suite check and
before M4's first quoted measurement, and every criterion is satisfiable: M3 measures the
unextended surface (71,212 ≤ 76,000, counts unchanged, no carve-out even needed), and the counts
move under the carve-out when the extension lands. The trap is that `plan.md` M1's sequencing note
*invites* the early landing ("M1 may run in parallel with M5's extension") without naming the upper
bound.

Everything else in the walk is consistent. Two orderings I checked and found already handled: the
pre-flight baseline is taken before the extension but is unaffected by it (`AGENTS.md` absent →
0 tokens, and `plan.md` M5 says so explicitly); and lowering `AlwaysLoadedTokenBudget` at M5 touches
no assertion, because `TestEstimateTokens` only asserts it is positive and the over-budget fixture
derives its size from the constant.

## Judgment 4 — regression over iterations 1-5

Spot re-executed, not re-read:

| Origin | Fix | Status | Evidence |
|---|---|---|---|
| iter1 D2 | `AGENTS.md` singleton via tracked files | RESOLVED | `git ls-files --full-name ':(top)*AGENTS.md' ':(exclude,top)internal/template/templates/*'` runs, exit 0, empty — the correct pre-run answer |
| iter1 D3 | integration branch given a discriminator | RESOLVED | `REQ-AMC-014` (spec.md:310-316) requires `git rev-parse --abbrev-ref HEAD` + `git rev-list --count main..HEAD` |
| iter1 D7 | nested-`CLAUDE.md` asymmetry stated | RESOLVED | spec.md §D.6:412-423, two-property table |
| iter1 D9 | `REQ-AMC-006` recast onto this SPEC | RESOLVED | spec.md:190-193 + §D.6:425-428 revival conditions |
| iter2 N1 | headroom ratio pinned | RESOLVED | `REQ-AMC-013`: "15 %, within ±2 percentage points … at or below 75,000" |
| iter2 N2 | design.md §4 cross-ref corrected | RESOLVED | design.md:173 → `AC-AMC-013`, which is indeed the duplicate-injection criterion |
| iter3 D1 | `AC-AMC-019` reads the band | RESOLVED | acceptance.md:186-189, band + ±1,000 agreement, both halves |
| iter3 D2 | enumeration must include `AGENTS.md` | RESOLVED | `REQ-AMC-008` + `AC-AMC-017` + `AC-AMC-018` ordering assertion |
| iter3 D4 | version/HISTORY provenance | RESOLVED | `version: "0.3.6"` = last HISTORY row; provenance rule spec.md:22-28 |
| iter4 E1 | §C.4 cuts recomputed | RESOLVED | arithmetic re-derived here: 71,212 + 6,144 − 66,371 = **10,985**; 75,282 + 6,144 − 66,371 = **15,055**; 24,576 ÷ 4 = 6,144 |
| iter4 E2 | Arm B added | RESOLVED | plan.md:104-112 + `AC-AMC-007` |
| iter5 F1 | cardinality carve-out | RESOLVED | judgment 1 above |
| iter5 F2 | Arm B baselined on the integration figure | RESOLVED (with O2) | `grep -c 'integration-branch figure recorded at pre-flight'` → 1 in plan.md, 1 in acceptance.md; M1 block quote headlines 15,055 and reconciles 10,985 (plan.md:91-95) |

No prior finding reopened. The 0.3.6 delta is additive to all of them.

---

## Defects Found

No blocking defects.

- **O1** — `spec.md:205-206` (`REQ-AMC-008`) — "the root `AGENTS.md` **and every `@`-imported
  contract document**" never defines *contract document*. `design.md` §4 shows exactly one
  (`@AGENTS.md`), with `@.moai/config/sections/*.yaml` marked "unchanged", and every figure in the
  SPEC is computed with K = 1 — but a literal reader could count the two YAML imports, grow the
  surface by three entries instead of one, and find the carve-out's `+ 4` / `5` wrong. — Severity:
  minor — Class: **optional** — Required fix (if ever routed): one clause in `REQ-AMC-008` reading
  "contract document = the root `AGENTS.md`; the pre-existing `.moai/config/sections/*.yaml`
  imports are not contract documents and stay out of the enumeration."
- **O2** — `acceptance.md:46-50` (`AC-AMC-007`) — the Arm B baseline binds inside a conditional
  Given, so a worktree-baselined projection that fails to trip passes the criterion vacuously.
  Detection falls to `AC-AMC-018` at M5. — Severity: major — Class: **optional** — Required fix:
  add an unconditional half-criterion: "the projection's cited baseline equals the integration-branch
  figure recorded at `plan.md` §C."
- **O3** — `plan.md:63-70` (M1 sequencing note) + `acceptance.md:114-115` (`AC-AMC-015`, an **M3**
  criterion) — the enumeration extension has a stated lower ordering bound and no upper one. Landing
  it before M3 makes the suite red at M3 for any `AGENTS.md` over 19,152 B on this tree (2,872 B on
  the integration branch), because the diet does not land until M4. — Severity: major — Class:
  **optional** (an exit exists: land the extension after M3's suite check and before M4's first
  quoted measurement) — Required fix: one sentence in `plan.md` M5 naming that window, and naming
  raising `AlwaysLoadedTokenBudget` as the forbidden remedy.
- **O4** — `internal/config/token_budget_guard_test.go:113, 131, 196, 198` — the doc comment, the
  two `t.Errorf` format strings and the inline comment all say "3 fixed" / "want 4" and go stale
  with the extension. They are message text and comments, not expected-behavior assertions, so
  `AC-AMC-015`'s no-edit rule never bound them and the carve-out's "nothing else" must not be read
  as forbidding their update. — Severity: minor — Class: **optional** — Required fix: none in the
  SPEC; noted so a run-phase actor does not leave a message contradicting its own assertion.
- **O5** — `.moai/specs/SPEC-TOKEN-EFFICIENCY-001/acceptance.md:31,40` — that closed SPEC's
  `AC-TEF-002` records "+ 3 fixed surfaces" as landed fact. The fourth slot supersedes the wording.
  — Severity: minor — Class: **optional** — Required fix: none. A closed SPEC is a landing-time
  record; rewriting it to match later behavior would make the record false.

The six long-standing optional findings (§5.2 ordering, the §D.5 numbering gap, `probe-fixture.sh`
printing rather than asserting, `AC-AMC-021`'s "make build has been run", `AC-AMC-016`'s prose/GWT
mix, the four "(Event-detected)" labels) were re-examined against the 0.3.6 text. Nothing changed
that would make any of them blocking, and none is raised.

---

## Residual risk run phase inherits

Iteration 5's list is carried with item 5 discharged (F2 was folded in, so the documentation-only
residual no longer applies) and three items added. Nothing on iteration 5's list was overstated.

1. **Arm B is expected to fire at the integration baseline.** The required cut there is 15,055
   tokens. A blocker return from M1 is the mechanism working; do not read it as a stall, and do not
   proceed on the assumption M4 will find the difference.
2. **The `[HARD]` proxy is a bracket, not a point** (§A.4). Every downstream byte figure inherits
   error in both directions; the 8,192 B reserve absorbs it, and trading against that reserve is a
   decision to state rather than slack to spend.
3. **The codex probe is single-platform** — macOS, `codex-cli` 0.147.0 (§D.9). A smaller upstream
   default on another OS or version is caught only by re-probing.
4. **The 66,371 bound is conservative by ~885 tokens** ((75,000 + 1,000) ÷ 1.13 = 67,256.6).
   Deliberate; not slack.
5. **The suite goes red between M2 and M4 if the enumeration extension lands early** (O3). Land the
   extension after M3's `AC-AMC-015` suite check and before M4's first quoted measurement. If it is
   red anyway, the remedy is the diet or the ordering — **never** raising
   `AlwaysLoadedTokenBudget`, which contradicts `REQ-AMC-013` outright.
6. **A worktree-baselined Arm B projection passes `AC-AMC-007` silently** (O2). The pre-flight
   integration figure is the baseline; record it before M1 starts, and quote it in the projection so
   the baseline is visible rather than assumed.
7. **The integration-branch figure 75,282 is carried, not re-measured here.** It comes from
   `spec.md` §A.1 / `REQ-AMC-014`; `release/vX.Y.Z` was not checked out in this audit. Every Arm B
   and ratchet number derived from it inherits that provenance until pre-flight re-measures it.
8. **`AGENTS.md` sizing has a hard interaction with the pre-diet constant.** With the constant at
   76,000, the pre-diet headroom is 19,152 B on this worktree and 2,872 B on the integration branch.
   M2 should know that before choosing how much of the 24,576 B ceiling to spend.

---

## Verification appendix

All commands run with `pwd` = `/Users/goos/MoAI/moai-adk-go/.claude/worktrees/t82`, confirmed before
the first measurement; no `cd` into a nested `internal/*/` tree or into
`internal/template/templates/`.

| Claim | Command | Observed |
|---|---|---|
| Judged tree | `git rev-parse --short HEAD` | `58f0bdd43` |
| Document version | `sed -n '4p' spec.md` | `version: "0.3.6"` |
| Freeze held | `git status --short .moai/specs` | no output (open and close) |
| Delta bounded | `git log f10c8691f..58f0bdd43` / `git diff --stat` | 2 commits; spec 8, plan 19, acceptance 35, progress 11 lines |
| REQ count | `grep -oE '^\*\*REQ-AMC-[0-9]+' spec.md \| sort \| uniq -c` | 18 rows, each count 1 |
| AC count | `grep -oE '^\*\*AC-AMC-[0-9]+' acceptance.md \| sort -u \| wc -l` | 24 |
| Surface, rules | `grep -rLE '^paths:' --include='*.md' .claude/rules/moai \| xargs wc -c` | 14 files, 202,621 B |
| Surface, fixed | `wc -c CLAUDE.md .claude/output-styles/moai/moai.md`; `ls MEMORY.md` | 20,523; 61,706; MEMORY.md absent (0) |
| Surface total | 284,850 ÷ 4 | 71,212 tokens — matches §A.1 |
| Arithmetic | 24,576÷4; 71,212+6,144−66,371; 75,282+6,144−66,371 | 6,144; 10,985; 15,055 |
| F1 premise | `cat -n internal/config/token_budget_guard_test.go` | :129 `wantRuleCount + 3`; :197 `len(surface) != 4` |
| F1 tests green now | `go test ./internal/config/ -run 'TestAlwaysLoadedSurfaceEnumeration\|TestAlwaysLoadedTokenBudget\|TestMeasureAlwaysLoaded_WithMemory' -count=1` | `ok … 1.308s` |
| No other consumer | `grep -rn 'alwaysLoadedSurface\|measureAlwaysLoaded\|AlwaysLoadedTokenBudget\|MeasureAlwaysLoadedSection' --include='*.go' .` | only the guard pair + `MeasureAlwaysLoadedSection` callers + a report-local replica |
| F2 present | `grep -c 'integration-branch figure recorded at pre-flight'` | plan.md 1, acceptance.md 1 |
| `AC-AMC-010` executes | the criterion's `git ls-files` command verbatim | exit 0, empty (correct pre-run) |
| Lint | `moai spec lint <spec.md>`; `moai spec lint` | `✓ No findings`; exit 0 |
| D7 | `grep -m1 '^status:' .moai/specs/SPEC-ALWAYS-LOADED-DIET-001/spec.md` | `completed` |
| D8 | `grep -c syscall spec.md` | 0 |
| MP-7 | `grep -rn 'NEEDS CLARIFICATION' <spec dir>` | 0 |
| Out of Scope | `grep -n 'Out of Scope' spec.md` | 5 H3 blocks (:513, :522, :527, :536, :541), each with specific bullets |
| Tier threshold | `spec-workflow.md` § SPEC Complexity Tier | L → 0.85; REQ/AC ceiling 25/25 |

**Gaps (not observed).** The full Go test suite was not run (excluded by dispatch); only the three
affected tests in `internal/config` were executed. The integration branch `release/vX.Y.Z` was not
checked out, so 75,282 is carried from `spec.md` §A.1 rather than re-measured — every judgment that
uses it inherits that provenance. `probe-fixture.sh` was not executed. `design.md` and `research.md`
were read for the Tier L input contract, the import architecture, and cross-reference regression,
but not audited clause by clause. `|AGENTS.md|` is modelled at its 24,576 B ceiling throughout; the
actual document does not exist yet, so every "at ceiling" figure is a worst case, not a measurement.

---

## Recommendation

**PASS.** Every must-pass criterion holds on cited evidence, the aggregate is 0.925 against a Tier L
threshold of 0.85, and the score has risen for two consecutive iterations. F1 — the finding that
survived four audits — is closed correctly: the carve-out's set is exactly the affected set,
verified by reading the whole test file and sweeping the repository rather than by trusting the
table. F2 is folded in and binds the plan layer.

The five findings above are all optional. Routing any of them would cost another edit pass on a
document that has cleared the threshold twice, and none of them makes the SPEC unimplementable —
which is the bar this iteration was asked to apply. O3 is the one worth carrying forward: it is not
a defect in the SPEC's logic, it is a sequencing fact a run-phase actor should be told before it
discovers it as a red test, and it is stated in the residual-risk list precisely so it can be
carried without an edit.
