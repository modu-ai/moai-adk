# SPEC Review Report: SPEC-CODEX-E2E-GUARD-001

Iteration: 1/2 (Tier M ceiling per `harness.plan_audit_tier_ceilings` — S=1, M=2, L=3)
Verdict: FAIL
Overall Score: 0.94 (Tier M PASS threshold 0.80 — exceeded; FAIL is driven solely by blocking defect D1)

Auditor: plan-auditor (independent). Tree: `ace1c5440`, branch `WT-codex-e2e-guard`,
worktree `.claude/worktrees/t500`. Every verification below was executed in this run,
against this tree. M1 Context Isolation honored: only the SPEC artifacts + the tree were
used; no author reasoning consumed.

## Must-Pass Results

- [PASS] MP-1 REQ number consistency: REQ-CEG-001..008 sequential, no gaps, no duplicates
  (spec.md:131-163). AC-CEG-001..007 sequential (acceptance.md:7-75). Counts within Tier M
  ceilings: 8 REQ ≤ 16, 7 AC ≤ 16.
- [PASS] MP-2 EARS/GEARS format compliance — judged against the REQUIREMENT layer only
  (spec.md §C REQ-XXX entries; the Given-When-Then entries in acceptance.md §D are the
  verification layer and are graded under Group 4, never here): REQ-CEG-001/002 Event-driven
  (When…shall), REQ-CEG-003/005/006/008 Ubiquitous (…shall), REQ-CEG-004 State-driven
  (While…shall), REQ-CEG-007 compound Ubiquitous+Where (see D2, optional). No informal
  language, no GWT-as-REQ, no mixed formal/informal within a single requirement.
- [PASS] MP-3 YAML frontmatter validity: all 12 canonical fields present with correct types
  (spec.md:1-16) — `id`/`title`/`version` "0.1.0"/`status: draft`/`created: 2026-09-07`/
  `updated: 2026-09-07`/`author`/`priority: P2`/`phase: "v3.1.4 target"` (release label, not
  a prohibited stage name)/`module`/`lifecycle: spec-anchored`/`tags` CSV string; plus
  optional `tier: M`. No rejected snake_case aliases. `moai spec lint` returned 0 findings,
  exit 0 (see Gaps for the binary-lag attribution caveat on this evidence).
- [PASS] MP-4 Section 22 language neutrality: N/A — single-language (Go test files in this
  repo) SPEC; no multi-language tooling surface is covered. Auto-pass per MP-4.
- [PASS] MP-5 D7 cross-SPEC reconciliation: verification verb executed — all five referenced
  SPEC IDs exist and read `status: completed` (SPEC-CODEX-E2E-MEASURE-001,
  SPEC-CODEX-WIRING-001, SPEC-CODEX-LAUNCHER-001, SPEC-CODEX-INIT-001,
  SPEC-INIT-HARNESS-PROMPT-001). No retired/superseded/archived reference, therefore no
  reconciliation obligation and no BLOCKING finding.
- [PASS] MP-6 D8 cross-platform discipline: `syscall` appears only as a negative assertion
  and as the name of the build-tag/syscall guard being extended (spec.md:69, :90, :148-149,
  :189); the literal `//go:build` is present in the same context (spec.md:90, 2 matches).
  The SPEC extends AC-CL-014 (the existing syscall/build-tag guard) rather than introducing
  any syscall usage. Zero build tags / zero syscall imports / zero process-replacement
  identifiers / zero GOOS-suffixed files re-verified on all 10 uncovered files this run.
  No BLOCKING finding.
- [PASS] MP-7 clarification gate: `grep -rn '\[NEEDS CLARIFICATION' plan.md research.md` →
  zero matches in plan.md; research.md does not exist (Tier M — not a required artifact, so
  the check ran on the one artifact that exists). No clarification gate open.

## Category Scores (0.0-1.0, rubric-anchored)

| Dimension | Score | Rubric Band | Evidence |
|-----------|-------|-------------|----------|
| Clarity | 1.0 | 1.0 band | §B Definitions pins every load-bearing term (Real init path, Doctor judgment, Hermetic, Mutant proof, 12-file set); each REQ names its exact seams; descope rule explicit (plan.md M2.2, acceptance.md §D.1) |
| Completeness | 0.75 | 0.75 band | All sections + 12/12 frontmatter fields + five `### Out of Scope — <topic>` H3s with bullets (spec.md:240-263) + HISTORY carrying measurement commands (spec.md:265-269); one sparse layer — the AC adoption cells (D1) |
| Testability | 1.0 | 1.0 band | Every AC is Given-When-Then, binary-testable, single-invocation evidence commands, zero weasel words (acceptance.md §D); edge cases enumerated (§D.2) |
| Traceability | 1.0 | 1.0 band | spec.md §E table: 8/8 REQs covered (REQ-003 and REQ-008 share AC-004 legitimately); 7/7 ACs reference existing REQs; no orphans |

## Fact Verification (this run, tree ace1c5440)

Every load-bearing claim in the SPEC was re-measured and verified TRUE:

1. `doctor_codex_test.go:26-34` wires via direct `codexwiring.Wire` (:30) — axis-1 gap real.
2. `statusline_test.go:47` `TestStatusLineDefaultSubsetOfAllowlist` exists, plus `:31/:64/:79`
   — the axis-2 premise refutation (§F.1) is correct and honestly recorded. Baseline
   `go test ./internal/codexwiring/ -run TestStatusLine -count=1` → `ok … 0.636s` (SPEC
   recorded 0.617s; timing variance only).
3. `codexSpecFiles = 2` (guards_test.go:30); exactly 12 non-test codex files in internal/cli
   (measured); AC-CL-013 `:82` / AC-CL-014 `:119` / AC-CL-016 `:150` line citations exact;
   comment-skip at `:156-160` exact; reconciliation comment `:27-29` exact.
4. Exec call-site table EXACT: launcher `:457` req.Program; mcp_codex `:350/:433/:1889`
   binaryPath; review_gate `:129` `"git"`; other 9 files zero sites. Raw mention counts
   (launcher 2, mcp 4) are call + comment-mention; SPEC §A counts real call sites only and
   §D.2.3 documents both comment lines (`:136`, `:339`) — internally consistent.
5. "41" figure: `find internal -name '*codex*_test.go' | wc -l` → 41; exactly 2 `func Test`-less
   files, the build-tag fixtures at 62/18 lines — §F.2 verified in full.
6. `configtoml.go` allowlist 29 tokens (:32-40), default 5 (:45-47), mutant target `:46`
   carries `"git-branch"` — M1's target line is correct.
7. M2 positive-path design verified executable against `checkCodexWiring` semantics
   (doctor_codex.go:86-207): `stubMoaiLookup(t, true)` is name-agnostic (answers both `moai`
   and `codex` lookups), wired-project branch passes hooks/sidecar/config checks on
   init-written artifacts, and `codexStaleSkillFinding` (:376-388) is fail-open on an empty
   temp home. The negative companion's `stubCodexLookup(t, true, false)` hits exactly the
   `!wired && !codexInstalled` silent-OK branch (:96-103). Expected-GREEN premise holds.
8. Guard baseline: `go test ./internal/cli/ -run 'TestCodexSpecFiles|…' -count=1 -timeout
   600s` → `ok … 0.881s` (SPEC recorded 0.929s).
9. `moai spec lint .moai/specs/SPEC-CODEX-E2E-GUARD-001/spec.md` → 0 findings, exit 0.

## Defects Found

D1. RED-now adoption cells missing on all release-blocking ACs; extension-AC evidence is
   non-discriminating — acceptance.md:7-75 (+ plan.md §C/§F) — Six of seven ACs claim
   must-pass severity, and none carries a RED-now adoption cell (single-invocation command +
   verbatim stdout + exit code + tree SHA measured on the pre-implementation tree), per the
   two-cell adoption discipline (`.claude/rules/moai/development/verification-completeness.md`
   §2/§2.1: "A criterion with one cell is unadopted"). Green-path cells exist; the RED cells
   do not. Worse, for AC-CEG-005/006/007 the evidence command
   (`go test -run TestCodexSpecFiles_…`) produces a byte-identical `ok` line BEFORE the work
   (2-file guard form) and AFTER it (12-file form) — the green cannot certify that the
   extension happened; this is the vacuous-green / empty-swept-set shape §1.1 exists to stop
   ("a pass whose swept set is empty asserts nothing"). AC-CEG-003's shape is judged
   COHERENT: the mutant procedure is deterministic, documented, and re-executable at any
   time (it is not a one-time historical event), so it does not hit the §2 undecidable
   disposition, and its must-pass classification gates a genuine discovered-defect
   possibility (a vacuous guard). — Severity: major — Class: blocking — Required fix:
   (a) add to acceptance.md §D one RED-now cell per release-blocking AC, pinned to
   `ace1c5440`, e.g. AC-CEG-001: run
   `go test ./internal/cli/ -run TestRunInit_ThenDoctorCodexWiringHealthy -count=1` today
   and record the `[no tests to run]` observation + exit 0 (proves the gap is real, so the
   future green is not vacuous); AC-CEG-005/006/007: record the current 2-file scope
   observation (e.g. verbatim output of
   `grep -n "codexSpecFiles = " internal/cli/codex_launcher_guards_test.go`) AND add a
   positive swept-count assertion to each extended guard (`len(fileSet) == 12`) so the green
   is self-describing; AC-CEG-003: record explicitly that the criterion is self-RED (the
   mutant observation IS its red-now cell); AC-CEG-004: record that §F is already present at
   plan phase as its starting observation.

D2. REQ-CEG-007 is a compound requirement — spec.md:155-160 — Three shall-clauses in one
   entry ("shall extend…; Where…shall narrow or allowlist…; shall reconcile…"). Each clause
   is individually pattern-conformant and the GEARS compound tolerance covers chained
   modifiers, but the cleaner form is one pattern per REQ or the modifier-chained form.
   — Severity: minor — Class: optional — Required fix: split into REQ-CEG-007a/b or rewrite
   as `[Where …] The AC-CL-013 neutrality judgment shall …` with the reconcile clause as a
   second REQ.

D3. `related_specs` is not a schema field — spec.md:15 — The canonical optional set is
   `issue_number` / `depends_on` / `lint.skip` / `bc_id` / `amendment_of` / `tier`
   (spec-frontmatter-schema.md § Optional Fields). The decoder drops unknown fields, so no
   lint finding fires, but the field is undeclared and its cross-reference semantics are
   ungoverned (unlike `depends_on`, which the run-gate pre-flight reads).
   — Severity: minor — Class: optional — Required fix: either rename to `depends_on` if
   fulfillment semantics are intended (note: that would make the run-gate require
   `completed` status on all four — currently true) or document `related_specs` as an
   accepted extra in the SPEC authoring convention.

D4. Grammar nit — acceptance.md:22 — "an `CheckOK` informational outcome" → "a `CheckOK`…".
   — Severity: trivial — Class: optional — Required fix: one-word edit.

## Gaps (explicitly not observed)

- The spec-lint evidence was produced by the INSTALLED `moai` binary, which carries a build
  lag (built from `e79c010b8`, an ancestor of this tree's HEAD `ace1c5440`; `bin/moai` does
  not exist in this worktree and rebuilding would write to the audited tree). The 0-findings
  result is therefore attributed to the installed binary's lint rules, not to this tree's
  build. The relevant rule families (frontmatter schema, Out of Scope, EARS modality) are
  stable across that range, but the attribution is stated rather than silently assumed.
- `golangci-lint` was not run by the auditor (plan-phase scope; it is a run-phase gate in
  §E/plan.md M5).
- The mutant RED was not executed by the auditor — executing it would mutate
  `internal/codexwiring/configtoml.go` inside the audited worktree. M1's RED/GREEN evidence
  remains a run-phase obligation (E5), correctly placed there.
- Cross-model second opinion: `mcp__moai__audit_multi` was invoked with `project_root` pinned
  to this worktree and the synthesized anchor verdict; it returned `participant_count: 0`
  with no backend verdicts (fail-open). Recorded as INCONCLUSIVE — never as a pass. Verdict
  authority stays with this in-session audit.

## Regression Check

Iteration 1 — no prior defects to regress.

## Recommendation

FAIL on a single blocking defect; the aggregate (0.94) exceeds the Tier M threshold and
every measured fact in the SPEC verified true, so the confirming iteration-2 re-audit should
be scoped to the D1 delta only:

1. Add the RED-now adoption cells per D1's fix (acceptance.md §D edits + the
   `len(fileSet) == 12` swept-count assertion in the M3/M4 guard designs — plan.md M3.3/M4.2
   should carry the count assertion so the extended guard's green is self-describing).
2. Optionally apply D2/D3/D4 (wording-level; no behavioral impact).
3. On D1's repair, re-audit the delta; expected outcome PASS.
