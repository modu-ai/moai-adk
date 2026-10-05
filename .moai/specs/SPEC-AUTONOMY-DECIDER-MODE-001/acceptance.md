# acceptance.md — SPEC-AUTONOMY-DECIDER-MODE-001 (v0.1.0)

Every AC is Given-When-Then with a command and expected output. Commands are single-line simple
invocations that pass the worktree session guard — no `git` inside `$( )`, `<( )`, or heredocs;
a git check needing a base ref takes it as an environment variable (`MOAI_DM_BASE` — this card's
fork point `422aa1e20`; refresh to the actual merge-base if develop is absorbed first).

**An empty selection is not a pass.** grep-family ACs print their match count alongside; a 0
count is a FAIL, not a silent pass.

**The two AC layers.** This is a measurement-design card, so ACs split into **[PLAN]** — decidable
NOW at plan phase (design-artifact presence and content) — and **[RUN]** — decidable only against
run-phase records. [RUN] ACs read `--- PENDING-RUN` until the run record exists; that is not a
FAIL.

## §A. Acceptance criteria

### AC-DM-001 — Population inventory, pinning rule, and 판별 불가 exit [PLAN] — maps REQ-DM-001, REQ-DM-002

- **Given** spec.md
- **When** reading for the three candidate axes (C-A kickoff rounds with the sweep-exhausted
  fact, C-B other surfaces census-gated, C-C synthetic control), the pre-fixed pinning rule
  (lowest qualifying baseline ≤ 90%; natural preferred at a tie), and the 판별 불가 verdict as a
  valid termination with its evidence requirements
- **Then** all three elements exist

```bash
grep -c "synthetic control" .moai/specs/SPEC-AUTONOMY-DECIDER-MODE-001/spec.md; grep -c "pinning rule" .moai/specs/SPEC-AUTONOMY-DECIDER-MODE-001/spec.md; grep -c "판별 불가" .moai/specs/SPEC-AUTONOMY-DECIDER-MODE-001/spec.md
```

Expected: all three counts ≥ 1.

### AC-DM-002 — Power arithmetic committed before measurement [PLAN] — maps REQ-DM-003

- **Given** spec.md and research.md
- **When** reading for the structural-limit statement (baseline > 90% kills baseline+10%p), the
  balanced-50% restoration (band = 60% becomes meaningful), the two-arm N envelope (≈120–240;
  target 200; minimum batch 120 with its MDE ≈ 11–12%p), the pre-measurement commitment
  requirement, AND the research.md §3 worked rows at the stated parameters (π_d = 0.15 →
  n_d ≈ 18 / N ≈ 118, all rows one parameter set)
- **Then** all six elements exist

```bash
grep -c "90%" .moai/specs/SPEC-AUTONOMY-DECIDER-MODE-001/spec.md; grep -c "60%" .moai/specs/SPEC-AUTONOMY-DECIDER-MODE-001/spec.md; grep -c "120–240" .moai/specs/SPEC-AUTONOMY-DECIDER-MODE-001/spec.md; grep -c "N = 120" .moai/specs/SPEC-AUTONOMY-DECIDER-MODE-001/spec.md; grep -c "≈ 18" .moai/specs/SPEC-AUTONOMY-DECIDER-MODE-001/research.md; grep -c "120–240" .moai/specs/SPEC-AUTONOMY-DECIDER-MODE-001/research.md
```

Expected: all six counts ≥ 1.

### AC-DM-003 — Predicate pre-fixed with confidence fields [PLAN] — maps REQ-DM-006

- **Given** spec.md
- **When** reading for the exact discordant-pair binomial (two-sided α = 0.05), the
  never-report-alone baselines, the authoritative `probabilities` field declaration (t1244
  sync-audit D3 named as the reason), and the post-commit change prohibition
- **Then** all four elements exist

```bash
grep -c "discordant-pair exact binomial" .moai/specs/SPEC-AUTONOMY-DECIDER-MODE-001/spec.md; grep -c "probabilities" .moai/specs/SPEC-AUTONOMY-DECIDER-MODE-001/spec.md; grep -c "never-report-alone" .moai/specs/SPEC-AUTONOMY-DECIDER-MODE-001/spec.md
```

Expected: all three counts ≥ 1.

### AC-DM-004 — Composite rule and mode mapping pre-fixed [PLAN] — maps REQ-DM-007, REQ-DM-008

- **Given** spec.md
- **When** reading for the composite rule (agree → label; disagreement → `hold`), the explicit
  statement that `jev_min_confidence` gating is NOT applied inside the composite, the δ = 5%p
  mapping thresholds with better/worse/equal definitions, and the TypeSafe choice-with-criteria
  instrument shape (the ask.sh 422 lesson)
- **Then** all four elements exist

```bash
grep -c "disagreement → \`hold\`" .moai/specs/SPEC-AUTONOMY-DECIDER-MODE-001/spec.md; grep -c "5%p" .moai/specs/SPEC-AUTONOMY-DECIDER-MODE-001/spec.md; grep -c "criteria map" .moai/specs/SPEC-AUTONOMY-DECIDER-MODE-001/spec.md
```

Expected: all three counts ≥ 1.

### AC-DM-005 — Both arms, one protocol [PLAN] — maps REQ-DM-005

- **Given** spec.md
- **When** reading for same population/predicate/scrub/caps/pre-registration/identical payloads
  and the equal payload_id sets requirement with the asymmetric-drop prohibition
- **Then** both elements exist

```bash
grep -c "identical payloads" .moai/specs/SPEC-AUTONOMY-DECIDER-MODE-001/spec.md; grep -c "payload_id" .moai/specs/SPEC-AUTONOMY-DECIDER-MODE-001/spec.md
```

Expected: both counts ≥ 1.

### AC-DM-006 — LLM arm pool named with timing dependency [PLAN] — maps REQ-DM-009

- **Given** spec.md
- **When** reading for the primary pool (GLM / z.ai / `glm_task`) with model AND quota pool, the
  pre-declared Anthropic (Opus) supplement gated on the lead's quota window (externally gated
  timing recorded), and Codex as the named fallback
- **Then** all three elements exist

```bash
grep -c "glm_task" .moai/specs/SPEC-AUTONOMY-DECIDER-MODE-001/spec.md; grep -c "EXTERNALLY GATED" .moai/specs/SPEC-AUTONOMY-DECIDER-MODE-001/spec.md; grep -c "codex_task" .moai/specs/SPEC-AUTONOMY-DECIDER-MODE-001/spec.md
```

Expected: all three counts ≥ 1.

### AC-DM-007 — Synthetic construction spec [PLAN] — maps REQ-DM-004

- **Given** spec.md
- **When** reading for the ≥ 6 pre-fixed defect classes (mandatory-section omission, scope
  violation, violated precondition, fabricated-evidence claim, AC-count mismatch, internal
  contradiction), the 50:50 balance (baseline 50%), per-class ground-truth labels in the 4-value
  space, and the pilot rule (n = 20, difficulty-diagnostic only, revision via NEW criteria
  commit before the main batch)
- **Then** all four elements exist

```bash
grep -c "mandatory-section omission" .moai/specs/SPEC-AUTONOMY-DECIDER-MODE-001/spec.md; grep -c "50:50" .moai/specs/SPEC-AUTONOMY-DECIDER-MODE-001/spec.md; grep -c "NEW criteria commit" .moai/specs/SPEC-AUTONOMY-DECIDER-MODE-001/spec.md
```

Expected: all three counts ≥ 1.

### AC-DM-008 — Scrub gate and caps specs [PLAN] — maps REQ-DM-012, REQ-DM-013

- **Given** spec.md
- **When** reading for the four scrub classes (key-shaped strings incl. AKIA, settings.local
  family, absolute paths, PII lower-bound), the positive control before FIRST transmission, the
  fail-closed block (no strip-and-send), the three caps with `declared_at`, the call_cap formula
  2 × (N + controls), and the no-extension rule
- **Then** all six elements exist

```bash
grep -c "AKIA" .moai/specs/SPEC-AUTONOMY-DECIDER-MODE-001/spec.md; grep -c "FIRST transmission" .moai/specs/SPEC-AUTONOMY-DECIDER-MODE-001/spec.md; grep -c "strip-and-send" .moai/specs/SPEC-AUTONOMY-DECIDER-MODE-001/spec.md; grep -c "declared_at" .moai/specs/SPEC-AUTONOMY-DECIDER-MODE-001/spec.md; grep -c "2 × (pinned N + controls)" .moai/specs/SPEC-AUTONOMY-DECIDER-MODE-001/spec.md
```

Expected: all five counts ≥ 1.

### AC-DM-009 — Korean original and t943 discipline [PLAN] — maps REQ-DM-010, REQ-DM-011

- **Given** spec.md
- **When** reading for the no-translation clause, the t943 non-rerun clause with the
  constant-baseline-row citation rule, and the three-domain disjointness statement
  (t943 premise-triage / t1244 judge agreement / t1261 decider quality)
- **Then** all three elements exist

```bash
grep -c "not translate" .moai/specs/SPEC-AUTONOMY-DECIDER-MODE-001/spec.md; grep -c "re-run" .moai/specs/SPEC-AUTONOMY-DECIDER-MODE-001/spec.md; grep -c "premise-triage" .moai/specs/SPEC-AUTONOMY-DECIDER-MODE-001/spec.md
```

Expected: all three counts ≥ 1.

### AC-DM-010 — Autonomous kickoff record and evidence layout [PLAN] — maps REQ-DM-014, REQ-DM-015

- **Given** spec.md
- **When** reading for the autonomy record (card t1266 policy, bundled conditions a/b/c), the
  `.moai/reports/t1261/` layout with the tracked criteria_commit requirement on this branch, and
  the 5-section verdict format
- **Then** all three elements exist

```bash
grep -c "t1266" .moai/specs/SPEC-AUTONOMY-DECIDER-MODE-001/spec.md; grep -c "criteria_commit" .moai/specs/SPEC-AUTONOMY-DECIDER-MODE-001/spec.md; grep -c "Residual-risk" .moai/specs/SPEC-AUTONOMY-DECIDER-MODE-001/spec.md
```

Expected: all three counts ≥ 1.

### AC-DM-011 — t1244 attribution integrity and scope guard [PLAN] — maps REQ-DM-011, REQ-DM-015 (context)

- **Given** spec.md and this card's commit set
- **When** reading for the t1244 numbers attributed as t1244's measurements (105 rounds /
  approve 104 / 99.05% / 91.43% / bands 0) — attributed, not re-measured — and checking this
  card's commits touch no product code
- **Then** the numbers appear with attribution and the changed-file set is empty outside
  `.moai/specs/`

```bash
grep -c "99.05%" .moai/specs/SPEC-AUTONOMY-DECIDER-MODE-001/spec.md
MOAI_DM_BASE=422aa1e20 git diff --name-only "$MOAI_DM_BASE" HEAD -- internal/ internal/template/ cmd/ pkg/ | wc -l
```

Expected: first count ≥ 1; second count `0`.

### AC-DM-012 — M0 census artifacts and criteria_commit precedence [RUN] — maps REQ-DM-001, REQ-DM-002, REQ-DM-003, REQ-DM-014

- **Given** the run record's census section (per-candidate label distributions, constant
  baselines, exclusion lists) and its `criteria_commit:` field
- **When** verifying every candidate axis has a measured (or construction-derived) baseline and
  arithmetic verdict, that exactly one population is pinned (or the 판별 불가 evidence set is
  complete), and that the criteria commit exists on this branch with a commit time earlier than
  the first judge call timestamp
- **Then** all hold — a missing census row, a missing pin, or a post-call criteria commit is a
  [RUN] FAIL

```bash
git cat-file -e "<criteria_commit>"^{commit}; echo $?
grep -c "criteria_commit" .moai/reports/t1261/run-record.md
```

`<criteria_commit>` is substituted from the run record. Expected: `0` and ≥ 1; the time
comparison reads `git log -1 --format=%ct "<criteria_commit>"` against the first `runs/` row
timestamp. RESOLVED [RUN] **PASS** — run-record.md §1: `criteria_commit: 568d11754` (restart-1
registration), `git cat-file -e` exit 0, commit epoch 17:14:56 < declared_at 17:16:50 < first
call 17:17:04; census and pin evidence at progress.md §E.2 (M0).

### AC-DM-013 — Caps declared before the first call, execution within them [RUN] — maps REQ-DM-013, REQ-DM-014

- **Given** the run record's `turn_cap:` / `call_cap:` / `wall_clock_cap:` fields with their
  `declared_at` timestamps, and the executed call counts
- **When** comparing each `declared_at` against the first judge call timestamp, and the executed
  counts against the caps
- **Then** all declarations precede the first call and all execution stayed within caps — a cap
  reached means the verdict was written with what was measured, and any extension is a FAIL

```bash
grep -c "declared_at" .moai/reports/t1261/run-record.md
```

Expected: `declared_at` ≥ 3 (one per cap). RESOLVED [RUN] **PASS** — measured `declared_at` 7;
run-record.md §2 (three caps declared_at 2026-09-27T17:16:50+09:00, 14s before the restart-1
first call) + §5 call accounting (298/298 executed, retries 0, no cap reached, unmeasured 0).

### AC-DM-014 — Scrub execution: positive control precedes first transmission, zero hit-sent [RUN] — maps REQ-DM-012

- **Given** the run record's scrub section (positive-control result, per-payload
  `payload_id` + `scan: clean|hit` rows)
- **When** verifying the positive control (dummy secrets observed firing across the planted
  classes) is recorded before the first transmission row, every payload has a scan row, and no
  `hit` row is followed by a transmission of that payload
- **Then** all three hold — any hit-and-sent payload voids every judge number in the verdict

```bash
grep -c "positive-control" .moai/reports/t1261/run-record.md; grep -c "scan: hit" .moai/reports/t1261/run-record.md
```

Expected: positive-control ≥ 1 (pre-first-transmission); every `scan: hit` row carries a
`blocked` marker and hit-without-block rows number 0. RESOLVED [RUN] **PASS** — run-record.md
§6: positive-control FIRED 8/8 at 2026-09-26T20:51:27+09:00 (precedes both registrations' first
calls); restart-1 298 call rows all `scan_exit=0`; `scan: hit` data rows 0, hit-sent 0.

### AC-DM-015 — Identical payload sets across arms [RUN] — maps REQ-DM-005

- **Given** both arms' run records (or the arm manifest) with their judged `payload_id` sets
- **When** comparing the llm 단독 arm's set against the llm+jev arm's set
- **Then** the sets are equal — any item judged by one arm only is recorded unmeasured on the
  other, and unexplained asymmetry is a [RUN] FAIL

```bash
grep -c "payload_id set" .moai/reports/t1261/run-record.md
```

Expected: ≥ 1 recorded set comparison with an equality statement; the sets themselves are the
evidence. RESOLVED [RUN] **PASS** — run-record.md §7: payload_id set glm ≡ jev — C 9/9, P 20/20,
M 120/120, unmeasured 0, asymmetry 0 (recompute: `.moai/reports/t1261/analysis/output.txt`).

### AC-DM-016 — Verdict applies the pre-fixed predicate in the 5-section format [RUN] — maps REQ-DM-002, REQ-DM-006, REQ-DM-007, REQ-DM-015

- **Given** `.moai/reports/t1261/verdict.md`
- **When** reading for: the exact discordant-pair test applied as committed (or its 판별 불가 /
  measurement-impossible valid exit with per-candidate evidence), constant baselines alongside
  every accuracy figure, the REQ-DM-007 mapping applied verbatim with the pool named, and the
  5-section format (Claim / Evidence / Baseline-attribution / Gaps / Residual-risk) — and,
  whenever the mode recommendation is issued from the committed primary pool alone (no supplement
  batch ran), a Gaps/Residual-risk row naming the pool-fidelity limitation
- **Then** all hold — a verdict citing a metric not present in the criteria commit, reporting
  an accuracy without its constant baseline alongside, or making a primary-pool-only
  recommendation without the pool-fidelity Gaps/Residual-risk row, is a FAIL

```bash
grep -c "Baseline-attribution" .moai/reports/t1261/verdict.md 2>/dev/null || echo "verdict-absent"
```

Expected: the five section headers present (Claim / Evidence / Baseline-attribution / Gaps /
Residual-risk). RESOLVED [RUN] **PASS** — `.moai/reports/t1261/verdict.md` carries all five
headers; the committed predicate applied with constant baselines alongside every accuracy
figure; the pool-fidelity Gaps row present (Gaps #1). Per-item outcomes:
progress.md §E.2 self-verification table.

## §B. Checks deferred to run-phase judgment

The EXECUTION of REQ-DM-001 (census actually run), REQ-DM-002 (pin actually applied or 판별 불가
documented), REQ-DM-003 (arithmetic actually committed), REQ-DM-004 (construction actually
instantiated to spec), REQ-DM-005 (arms actually paired — AC-DM-015), REQ-DM-006 (predicate
actually applied — AC-DM-016), REQ-DM-007 (mapping actually applied), REQ-DM-008 (Jev stage
actually run on the committed instrument shape), REQ-DM-009 (pools actually as committed),
REQ-DM-012 (scrub actually executed — AC-DM-014), REQ-DM-013 (caps actually declared and
respected — AC-DM-013), REQ-DM-014 (bundled conditions actually met — AC-DM-012/013), and
REQ-DM-015 (layout actually produced) is judged against the run record and the verdict at
run-phase. The plan-phase artifacts warrant that the discipline is fixed in documentation before
measurement — nothing more, and exactly that.
