# SPEC Review Report: SPEC-SETTINGS-ORIGIN-001 (plan-audit, card t487)

Iteration: 1/1 (Tier S ceiling — `plan_audit_tier_ceilings` S=1)
Auditor: plan-auditor (independent, adversarial, fresh-judgment)
Date: 2026-09-05
Verdict: **PASS** (aggregate 0.80 ≥ Tier S threshold 0.75; no must-pass failure) — **conditioned on pre-run application of fixes D1–D3** (routed to manager-spec; with the Tier S ceiling of 1 there is no iteration-2 re-audit, so D1–D3 are verified mechanically by the lead: does AC-008 exist, is the AC-001 parenthetical removed/extended, is the §B grep pattern replaced).

Reasoning context ignored per M1 Context Isolation (author self-reported premise risks were re-derived from the artifacts and the tree, not accepted from the prompt).

---

## Claim

SPEC-SETTINGS-ORIGIN-001 (Tier S investigation SPEC, 4 artifacts: spec.md + plan.md + additive acceptance.md + progress.md) is fit for run-phase entry: the investigation it prescribes can answer the card's three questions (Q1 writer inventory, Q2 conditional external inflow, Q3 one recurrence-prevention recommendation) plus the SWEEP, its acceptance criteria are observable rather than assertable, and it respects the t480 baseline and the t485 C4 dead-axis ruling. Two major verification-layer gaps (REQ-008 uncovered; AC-001's cross-check scope narrowed below "repo-wide") and one vacuous command prescription (plan §B) must be fixed before run-phase kickoff; none is a must-pass failure.

## Evidence

Every row below was measured in this audit, this run, against the t487 worktree at `WT-settings-origin` / `25a3212a9` (read-back verified below).

### E1 — Preserved-artifact baseline (the SPEC's load-bearing premise)

| # | Command | Observed output | Matches SPEC? |
|---|---------|-----------------|---------------|
| E1a | `md5 -q .../t480/.../settings.json.worktree-dirty` | `b669972dc738d1bf925281dcc90f152e` | spec.md:30 — YES |
| E1b | `wc -c` same file | `24667` bytes | spec.md:32 — YES |
| E1c | `md5 -q .../settings.json.develop` | `568a3d3a32a360731d1d02f686f27d24` | spec.md:33 — YES |

### E2 — Dirty-copy fingerprint (§1 "dirty copy" column, independently re-measured)

| # | Command | Observed output | Matches SPEC? |
|---|---------|-----------------|---------------|
| E2a | `jq -r 'keys_unsorted\|join(", ")'` | `$schema, respectGitignore, cleanupPeriodDays, skillListingBudgetFraction, env, attribution, permissions, hooks, statusLine, outputStyle, showThinkingSummaries` (11 keys) | spec.md:43 — YES, exact order |
| E2b | `jq -c '.permissions.ask'` | `["Bash(sudo:*)"]` | spec.md:45 — YES (1 entry) |
| E2c | `grep -n 'Write\|Edit\|MultiEdit'` | `314: "matcher": "Write\|Edit\|MultiEdit",` | spec.md:46 — YES (line number too) |
| E2d | `grep -c 'status-transition-ownership'` | `1` | spec.md:47 — YES (exactly 1; lane-8's correction of t480's "부재" wording confirmed) |

### E3 — Template column of §1 (the relayed lane-8 baseline — every value independently reproduced)

`grep -nE '"(statusLine\|skillListing…\|plansDirectory\|hooks)"[[:space:]]*:'` on `internal/template/templates/.claude/settings.json.tmpl`:

```
3:   "hooks": {                    404: "skillListingBudgetFraction": 0.02,
395: "statusLine": {               405: "showThinkingSummaries": false,
406: "cleanupPeriodDays": 30,      407: "model": "sonnet",
408: "outputStyle": "MoAI-Easy",   409: "env": {
420: "permissions": {              600: "attribution": {
605: "respectGitignore": true,     606: "includeGitInstructions": true,
607: "plansDirectory": ".moai/plans"
```

All 13 relayed line numbers and the 14-key order in spec.md:43 reproduce **exactly**; `model`/`includeGitInstructions`/`plansDirectory` confirmed absent from the dirty copy (present in template). The relayed value is therefore accurate, not merely attributed.

### E4 — Structural must-pass checks

| Check | Command | Observed | Result |
|-------|---------|----------|--------|
| MP-1 REQ sequence | `grep -rE 'REQ-[0-9]{3}' spec.md` | REQ-001..REQ-008 at spec.md:60–67, no gaps, no duplicates | PASS |
| MP-2 GEARS (requirement layer) | read spec.md §2 | 7× ubiquitous ("The investigation shall…"), 1× event-driven (REQ-004 "**When** … shall…"), REQ-007 uses canonical negative "shall not be reported as zero". ACs are Given-When-Then in acceptance.md = correct verification-layer format (not penalized) | PASS |
| MP-3 frontmatter | read spec.md:1–16 | all 12 canonical fields present, correct types, no snake_case aliases; `priority: P2`, `phase: "v3.2.0 target"`, `lifecycle: spec-lite`, `tags` string | PASS (notes: `related_specs: []` is an undocumented extra field, decoder-ignored; `module:` value not path-like — neither fails the 12-field check) |
| MP-4 language neutrality | — | single-repo investigation SPEC, no multi-language tooling content | N/A |
| MP-5 D7 cross-SPEC | `grep -rEo 'SPEC-([A-Z][A-Z0-9]+-)+[0-9]+' <SPEC dir>` | only self-references (`SPEC-SETTINGS-ORIGIN-001`) in all 4 artifacts; no external SPEC-ID | PASS (no finding) |
| MP-6 D8 syscall | `grep -rc 'syscall'` each artifact | `0` in all four | PASS (auto — substring absent) |
| MP-7 clarification gate | `grep -rn 'NEEDS CLARIFICATION' <SPEC dir>` | none; research.md absent (Tier S) | PASS |
| Weasel words | `grep -nE 'appropriate\|adequate\|reasonable\|proper' acceptance.md` | none | — |

### E5 — Plan §B factual claims

| # | Command | Observed | Matches plan? |
|---|---------|----------|---------------|
| E5a | `jq . settings.json.tmpl` | `jq: parse error: Expected separator between values at line 130, column 8` | plan.md:21 "parse error at line 130" — YES, exact |
| E5b | `grep -c '^[a-zA-Z]' settings.json.tmpl` and `grep -c '^"'` | `0` and `0` (keys are indented + quoted; line 3 = `  "hooks": {`) | **plan.md:21's prescribed pattern `grep -n '^[a-zA-Z]'` is VACUOUS on this file** → defect D3 |

### E6 — Tree facts

| # | Command | Observed | Matches? |
|---|---------|----------|----------|
| E6a | `git -C <t487> branch --show-current; rev-parse --short HEAD` | `WT-settings-origin`; `25a3212a9` | plan §C:3 / progress.md — YES |
| E6b | `git worktree list \| wc -l` | `79` | plan §B "70+" — YES |
| E6c | `git status --porcelain -- .moai/specs/SPEC-SETTINGS-ORIGIN-001/` | `?? .moai/specs/SPEC-SETTINGS-ORIGIN-001/` (artifacts untracked; plan-phase commit pending — expected pre-audit state) | observation |
| E6d | `ls -d scripts .claude/workflows .github/workflows` (in t487 tree) | all three exist (16 / 9 / 20 entries) | **omitted from REQ-001/AC-001 enumeration** → defect D2 |
| E6e | `ls .moai/reports/t485/verdict.md` in primary vs worktree | primary: exists (8,722 B); worktree t487: `No such file or directory` | spec §6 reference unresolvable from run-phase worktree → defect D4 |

## Baseline-attribution

All measurements above were taken in this audit run, in the t487 worktree (`WT-settings-origin`, HEAD `25a3212a9`), against: the t480 preserved copies at `.claude/worktrees/t480/.moai/reports/t480/preserved-copies/` (md5-verified before use, E1); the template at `internal/template/templates/.claude/settings.json.tmpl` (worktree copy); and the SPEC artifacts under `.moai/specs/SPEC-SETTINGS-ORIGIN-001/`. Nothing in this report is carried over from t480's, lane-8's, or the author's measurements — every load-bearing figure was re-derived (and in the case of the §1 fingerprint, fully reproduced).

## Gaps (explicitly NOT observed)

- **Cross-model convergence (audit_multi) not run.** `grep -rn 'audit_model' .moai/config/` (worktree and primary) returns no key — the project config does not request a cross-backend second opinion. Additionally the audit subject is untracked markdown (E6c); both backends collect a git diff (`uncommittedChanges`/`baseBranch`), which does not represent untracked files, so the call is structurally uninformative for this subject. Backends are fail-open; the verdict rests entirely on the in-session verified evidence above.
- Run-phase executability of the milestones (M1–M5) was assessed by reading, not by execution — no sweep or inventory grep was run against the 79 worktrees (that is run-phase work; running it here would exceed the audit's read-only shape and duplicate M1).
- Whether any writer of `.claude/settings.json` actually exists under `scripts/`, `.claude/workflows/`, or `.github/workflows/` was NOT determined — only that the directories exist and fall outside AC-001's enumerated cross-check scope (D2 is about the check's swept set, not a claim that a writer is there).
- progress.md §E.2–§E.4 are placeholders (expected — run/sync phases not started).

## Residual-risk

- If D2 is not fixed and run-phase executes the cross-check as parenthesized in AC-001, a writer outside the enumerated roots would be invisible and Q1's "exhaustive" answer would be unfalsifiably incomplete.
- If D3 is not fixed, the §B textual extraction runs vacuous (0 lines) and a run-phase actor must notice the empty output to avoid a false "template has no keys" reading; REQ-007/AC-007 make that misreading a detectable violation, but the trap should not exist.
- The sweep (M1) uses plain `git status`, which may contend on `index.lock` in trees other lanes are actively using; the locked-tree Gap edge case makes the failure safe (recorded, not silently clean) but noisy (D7).
- The dirty-copy md5 could drift between this audit and run-phase pre-flight; plan §C:1 + acceptance.md edge case 2 already gate on exactly that (halt and report) — residual accepted.
- `sync expected N/A` is a lead-side close decision outside this audit's scope; nothing here verifies it.

---

## Must-Pass Results

- [PASS] MP-1 REQ number consistency — REQ-001..008 sequential, spec.md:60–67 (E4).
- [PASS] MP-2 EARS/GEARS compliance — requirement layer (spec.md §2) only: 7 ubiquitous + 1 event-driven (REQ-004); ACs correctly GWT in the verification layer (E4).
- [PASS] MP-3 YAML frontmatter validity — 12/12 canonical fields, correct types, no rejected aliases, spec.md:1–16 (E4).
- [N/A] MP-4 language neutrality — single-repo investigation SPEC; no multi-language tooling content.
- [PASS] MP-5 D7 cross-SPEC reconciliation — only self-references found; no BLOCKING finding (E4).
- [PASS] MP-6 D8 cross-platform discipline — `syscall` absent from all four artifacts (count 0 each); auto-pass (E4).
- [PASS] MP-7 clarification gate — no `[NEEDS CLARIFICATION]` markers; research.md absent (Tier S) → that half N/A (E4).

## Category Scores (rubric-anchored)

| Dimension | Score | Band | Evidence |
|-----------|-------|------|----------|
| Clarity | 0.75 | minor ambiguity in 1–2 items | AC-001 "repo-wide" vs parenthetical narrowing (contradicted by plan §A "starting map, NOT exhaustive" and plan §E "repo-wide grep" — the AC is the outlier); AC-003's N-frame ("N trees plus the primary" vs REQ-003 "primary included" — `git worktree list` already contains the primary) |
| Completeness | 1.00 | all sections + frontmatter | HISTORY (spec.md:20), Context §1, Scope §5 with four `### Out of Scope — <topic>` H3s each with specific bullets (spec.md:87–102); 8 REQ / 7 AC within Tier S ceilings (8/8) |
| Testability | 0.75 | one AC measurable with minor interpretation | AC-002's could-produce judgment is interpretive, mitigated by the marker-tied-reasoning form requirement; all other ACs binary; no weasel words (E4) |
| Traceability | 0.75 | one REQ uncovered | REQ-001..007 → AC-001..007 one-to-one; **REQ-008 (spec.md:67) has no AC and no plan §E item** |

Aggregate (harmonic mean, skeptical stance): 4 / (1/0.75 + 1/1.00 + 1/0.75 + 1/0.75) = **0.80** ≥ 0.75 (Tier S threshold) → **PASS**.

## Defects Found

| ID | Artifact:line | Description | Severity | Class | Required fix |
|----|---------------|-------------|----------|-------|--------------|
| D1 | acceptance.md §D (vs spec.md:67) | REQ-008 (baseline reuse: build on t480 without re-running; treat lane-8 fingerprint as hypotheses) has NO corresponding AC — AC-001..007 cover REQ-001..007 only; plan §E self-check items don't cover it either. Breaks the two-layer coverage invariant (every REQ ≥ 1 AC). | major | **blocking** | Add AC-008, e.g.: "Given the completed investigation, When the verdict is reviewed, Then t480's six eliminations appear only as cited baseline (no re-execution commands in the verdict's Evidence) and the §1 lane-8 fingerprint observations appear as hypotheses in the H1–H4 tree, not as conclusions." |
| D2 | acceptance.md:7 (AC-001) + spec.md:60 (REQ-001) | The parenthetical directory enumeration (`internal/, pkg/, cmd/, .claude/hooks/, template sources, Makefile`) narrows the "repo-wide" search and omits writer-capable roots verified present in this tree: `scripts/` (16 entries), `.claude/workflows/` (9), `.github/workflows/` (20) (E6d). A cross-check executed as parenthesized cannot surface a writer outside the enumeration — Q1's exhaustiveness claim is unfalsifiable over part of the repository. plan §A ("starting map, NOT exhaustive") and plan §E ("repo-wide grep") show the intent is repo-wide. | major | **blocking** | Delete the parenthetical narrowing in AC-001 (keep "repo-wide") or extend it with `scripts/`, `.claude/workflows/`, `.github/`, "any top-level shell/JS/Python", and state the cross-check grep is bounded only by the repo root. Mirror the same scope wording in REQ-001 if edited. |
| D3 | plan.md:21 (§B) | The prescribed textual-extraction command `grep -n '^[a-zA-Z]'` is vacuous on the actual template file — verified 0 matches (keys are indented and quoted; line 3 is `  "hooks": {`) (E5b). Executed as written it produces empty output, the exact "empty output read as zero" hazard REQ-007/AC-007 forbid; the M3 key-order comparison would be silently vacuous. | major | **blocking** | Replace with a working extraction, e.g. `grep -nE '^[[:space:]]*"[a-zA-Z$]+"' <tmpl>` (verified: yields all 13 top-level keys with the exact line numbers already cited in spec §1), or render first (`moai init` into a temp dir / renderer harness). |
| D4 | spec.md:107 (§6) | The t485 verdict reference `.moai/reports/t485/verdict.md` has no tree annotation; the file exists only in the primary checkout and NOT under the t487 worktree (E6e), so a run-phase executor resolving the relative path in-worktree finds nothing. The t480 reference is annotated ("worktree t480 copy is canonical"); the t485 one is not. | minor | optional | Annotate: "`.moai/reports/t485/verdict.md` (primary-checkout copy)". |
| D5 | spec.md:39 (§1 header) | Table header attributes the whole table to "measured … on the preserved copy", but the template column was measured on the .tmpl by grep (lane-8) — attribution wording imprecise for that column. Values themselves independently verified accurate (E3); risk is a reader mis-taking the template column as author-measured on the artifact. REQ-008 + M3 already downgrade it to hypothesis + re-verification. | minor | optional | Annotate the template column's provenance ("lane-8 grep line numbers on the .tmpl; re-verified in M3"). |
| D6 | spec.md:62 / acceptance.md:9 (AC-003) | Counting-frame inconsistency: REQ-003 puts the primary inside the enumeration ("every worktree in `git worktree list` — the primary checkout included"); AC-003 reads as primary additional to N ("N trees plus the primary checkout"). `git worktree list` includes the primary as entry 1, so the two frames differ by one — the count-equality test's N is ambiguous. | minor | optional | Pick one frame, e.g. "N = `git worktree list` line count (the primary is entry 1); trees checked must equal N." |
| D7 | spec.md:62 (REQ-003) | Plain `git -C <wt> status --porcelain` may take an index write lock in trees other lanes are actively using (documented repo hazard: statusline status calls take index write locks; transient index.lock under lane load). Serial execution + the locked-tree Gap edge case bound the risk to noise, not false-clean. | minor | optional | Prescribe `git --no-optional-locks -C <wt> status --porcelain -- .claude/settings.json` (or `git -C <wt> diff --name-only HEAD -- .claude/settings.json`). |
| D8 | plan.md §F M4 (vs spec.md:54 H2) | H2's old-template-render sub-branch has no milestone explicitly owning its measurement; M4's "old binaries on disk" plausibly covers it (old binaries embed old templates) but the mapping is implicit. | minor | optional | One clause in M4: "old binaries on disk — including their embedded templates (covers H2's render variant)." |

Observations (not defects): `related_specs: []` is an undocumented extra frontmatter field (decoder-ignored); `module:` value not path-like (present/non-empty, passes the 12-field check); SPEC artifacts currently untracked (`??`), plan-phase commit pending — expected pre-audit state; Tier S plan-audit ceiling = 1 iteration, so D1–D3 fixes are lead-verified mechanical text checks, not a re-audit.

## Premise-risk adjudication (requested by the dispatch)

1. **Relayed .tmpl key-order baseline (spec §1)** — handling ACCEPTABLE. The value is attributed (lane-8, 2026-09-05), explicitly downgraded to hypothesis by REQ-008, and re-verification is named (M3). Decisive upgrade: this audit independently reproduced **every** value in both columns (E2, E3), so the baseline is now verified, not relayed. Residuals are D3 (the vacuous extraction command meant to re-verify it) and D5 (attribution wording).
2. **H1 "accumulated/edited file" hypothesis** — adequately framed: stated as hypothesis with named discriminators (H4 distribution via M1 sweep; Q2 CC-write-signature research with "verify from CC docs/changelog, not assumption", plan.md:51). Not presented as conclusion anywhere.
3. **Tier S + additive acceptance.md** — coherent. Tier S minimum is 2 artifacts; the SPEC adds a third because the dispatch enumerates it (documented at spec.md:71 and plan.md:4). REQ/AC budgets sit within Tier S ceilings (8 REQ / 7 AC ≤ 8/8); tier-relevant consequences (threshold 0.75, audit ceiling 1) are stated in plan.md:4. No defect.

## Investigation answerability (the audit's core question)

- **Q1** answerable via M2+M3 + AC-001/002 — **provided D2 is fixed** (otherwise the exhaustiveness claim is unfalsifiable over `scripts/`, `.claude/workflows/`, `.github/`).
- **Q2** answerable via M4 + AC-004 with the conditional honored in both directions (the "Q1 DID find a producer" branch is explicitly covered).
- **Q3** answerable via M5 + AC-005: exactly one recommendation, evidence-premised, t485-C4-compliant (process-level if proposing observation).
- **SWEEP** answerable via M1 + AC-003: N-vs-checked equality, per-dirty-tree path+md5+3 markers, locked/mid-merge trees recorded as Gaps (no silent skip), secrets never pasted (acceptance.md edge cases). 79 trees confirmed present (E6b).
- **t485 C4 respect**: spec.md §1 "Position relative to t485 ruling C4" states the one-respect difference (artifact persisted → shape forensics valid; naming the writer from state remains dead) once and consistently; REQ-005/AC-005 bind it. No contradiction found.
- ACs are observable, not assertable: AC-007 binds every quantitative/negative claim to command + observed output, and the verdict artifact (AC-006) requires verbatim output per Evidence entry.

## Regression Check

Iteration 1 — N/A (no prior iterations).

## Recommendation

PASS at 0.80. Before run-phase kickoff, manager-spec applies D1–D3 (three small text edits: add AC-008; un-narrow AC-001; replace the §B grep pattern); D4–D8 are at the lead's discretion. Because the Tier S audit ceiling is 1, the D1–D3 fixes are verified by the lead as mechanical text checks (existence of AC-008, absence/extension of the AC-001 parenthetical, replacement of the `^[a-zA-Z]` pattern) rather than by a re-audit. Implementation Kickoff Approval remains mandatory and is not affected by this PASS.
