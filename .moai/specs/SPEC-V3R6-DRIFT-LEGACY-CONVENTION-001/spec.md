---
id: SPEC-V3R6-DRIFT-LEGACY-CONVENTION-001
title: "moai spec drift — comprehensive false-positive elimination + era/grandfather alignment + close-subject doctrine"
version: "0.1.1"
status: in-progress
created: 2026-06-03
updated: 2026-06-03
author: manager-spec
priority: P2
phase: "v3.0.0"
module: "internal/spec"
lifecycle: spec-anchored
era: V3R6
tier: M
tags: "drift, lifecycle, commit-convention, era, grandfather, false-positive, internal-spec"
---

# moai spec drift — comprehensive false-positive elimination + era/grandfather alignment + close-subject doctrine

## HISTORY

| Version | Date | Author | Change |
|---------|------|--------|--------|
| 0.1.0 | 2026-06-03 | manager-spec | Initial plan-phase authoring (Tier M). Successor to SPEC-V3R6-DRIFT-CONVENTION-ALIGN-001 (close-infix + narrow backfill-skip; reduced drift 67→50). Comprehensive detector alignment: fix the 4 residual false-positive mechanisms + align `moai spec drift` with `moai spec audit`'s era/grandfather policy. Genuine-incomplete backfill (mechanism ⑤) is explicitly OUT OF SCOPE (separate operational follow-up). Root cause pre-verified by orchestrator; per-mechanism git-log spot-verified during plan-phase. |
| 0.1.1 | 2026-06-03 | manager-spec | plan-audit PASS-WITH-DEBT 0.835 → patch D1/D2/D3 (target re-audit ≥0.90). **D1 (design-level)**: M3 redesigned from a dead in-window gate to a SECONDARY scope-prefix grep fallback — empirically confirmed the combined-scope close `cf7d78a9c` is NOT in `git log --grep=SPEC-CCSYNC-CLAUDEMD-001` (the per-SPEC window) but IS in `git log --grep=SPEC-CCSYNC`. REQ-DLC-001/002 rewritten; 3-gate LSGF-001 safety (FALLBACK-ONLY + closeInfixMatch + distinguishing-segment cross-check); AC-DLC-001 rebound (deterministic fixture resolving BOTH siblings = binding; live-repo CCSYNC proof); new AC-DLC-012 (non-sibling partial-prefix collision guard); new AP-7 (dead in-window gate). **D2**: AC-DLC-011 grep oracle tightened to require the prohibition co-located (≤400 chars) with the combined/abbreviated-scope subject. **D3**: REQ-DLC-005/006 made normative — era-exempt/terminal SPECs emit `Drifted: false` + `continue` (record PRESERVED in `Records[]`, not dropped), matching `drift.go:73-84`. 11 REQ × 12 AC; bidirectional traceability restated. |

---

## A. Background and Problem Statement

### A.1 Symptom

`moai spec drift` reports **51 drifted records** (`moai spec drift --json` → `.Records[] | select(.Drifted==true)`; the live `--count` reads 53 at the time of authoring, the small delta being parallel-session SPEC churn). The predecessor SPEC-V3R6-DRIFT-CONVENTION-ALIGN-001 carved out the canonically-closed false-positives (close-infix `4-phase close` / `Mx-phase audit-ready` recognition + narrow backfill-skip) and reduced the count from 67 to 50. The residual ~51 records are **not** a single class.

The canonical lifecycle audit tool `moai spec audit` reports the great majority of these same SPECs as **CLEAN (0 MUST-FIX)** — grandfather-protecting pre-V3R6 eras and treating terminal states as authoritative. The two tools disagree because `moai spec drift` (`internal/spec/drift.go`) performs a **raw frontmatter-vs-git compare with NO era awareness, NO terminal-state authority, and a stale legacy transition rule**, whereas `moai spec audit` (`internal/spec/audit.go`) consumes `ClassifyEra` and exempts grandfathered eras. The drift output is wrong for these records; the SPEC state on disk (frontmatter + audit) is correct.

### A.2 Verified decomposition into 5 mechanisms

Each mechanism below was spot-verified by per-SPEC `git log main --grep=<specID>` during plan-phase (HEAD at authoring). The run-phase MUST re-classify the full 51-record set freshly (`moai spec drift --json` is the SSOT; the bucketing below is the plan-phase hypothesis, not a frozen contract — re-verify per-SPEC before asserting any AC).

| # | Mechanism | Root cause (verified) | Verdict | Representative evidence |
|---|-----------|----------------------|---------|------------------------|
| ① | combined-scope close | Close subject abbreviates scope, e.g. `docs(SPEC-CCSYNC): sync-phase artifacts (CLAUDEMD + TOOLCAT status→implemented...)`. `commitMatchesSPECID` runs `ExtractSPECIDs` on the `--oneline` SUBJECT; the regex `SPEC-[A-Z0-9-]+-[0-9]+` finds no full-ID token (`SPEC-CCSYNC` has no trailing `-NNN`), so close + sync commits are rejected; the walker falls through to `feat(SPEC-CCSYNC-CLAUDEMD-001): M1-M5…` → infers `implemented`; frontmatter = `completed`. | FALSE-POSITIVE (spec.md=truth) | `SPEC-CCSYNC-CLAUDEMD-001` / `SPEC-CCSYNC-TOOLCAT-001` [completed↔implemented]; close commit `cf7d78a9c`, sync commit `da2fbcedf` |
| ② | stale `sync`/`docs(sync)`→completed rule | `transitions.go` rules `{"sync"→completed}` and `{"docs(sync)"→completed}` are legacy. In the modern 4-phase model, sync-phase = `implemented` and ONLY close-infix → `completed`. So `sync(SPEC-X): … implemented` git-infers `completed` while frontmatter (correctly) says `implemented`. | FALSE-POSITIVE (frontmatter `implemented` is correct) | `SPEC-V3R5-STATUSLINE-STDINFIELDS-001` (`438e5b214 sync(...): lifecycle complete — v0.3.0 implemented`), `SPEC-V3R6-PLAN-AUDITOR-GEARS-ALIGN-001`, `SPEC-V3R6-SKILL-GEARS-ALIGN-001` — all [implemented↔completed] |
| ③ | terminal-state uninferrable | `superseded`/`archived`/`rejected` cannot be inferred from any git commit convention; the archive commit `chore(specs): … archived` is skip-meta. The walker infers from the earlier feat/plan commit → `in-progress` or `implemented`. | FALSE-POSITIVE (terminal frontmatter is authoritative) | `SPEC-V3R3-HARNESS-001` [superseded↔in-progress], `SPEC-V3R6-AGENT-FOLDER-SPLIT-001` [superseded↔in-progress], `SPEC-V3R6-AGENT-MODEL-ROUTING-001` [archived↔implemented] |
| ④ | grandfather/legacy era | V3R2/V3R3 pre-4-phase SPECs were bulk-closed via sweep commits (`sync(specs): V3 bulk status drift …`) carrying no individual full-IDs. `audit.go`/`era.go` ALREADY grandfather-protect V2.x/V3R2-R4/V3R5 (`Era.IsModern()` true only for V3R6 — see `.claude/rules/moai/workflow/lifecycle-sync-gate.md`), but `drift.go` performs a RAW compare with NO era awareness — policy inconsistency. | FALSE-POSITIVE (era-exempt) | V3R2 ×6, V3R3 ×6, V3R4 ×1, DESIGN ×1 (e.g. `SPEC-V3R2-SPC-002` [completed↔planned]) |
| ⑤ | genuine incomplete close | Modern V3R6 where the close was genuinely not performed or the frontmatter is genuinely wrong. The ONLY mechanism needing real remediation. | GENUINE | Residue after ①②③④ removed — the exact set is run-phase classification work, deferred to a separate operational follow-up SPEC |

### A.3 The policy inconsistency (the heart of the defect)

`moai spec audit` and `moai spec drift` are sibling lifecycle tools that SHOULD report consistent state. The audit engine applies three policies the drift detector lacks:

1. **Era awareness** (mechanism ④): `auditSpec` calls `LoadEraSignalsFromDir` + `ClassifyEra`; grandfather-protected eras (`Era.EraFinal() == true`) emit no MUST-FIX. `drift.go` `DetectDrift` has no `ClassifyEra` call at all.
2. **Terminal-state authority** (mechanism ③): `checkV3R6Drift` returns nil (no drift) when `specStatus == "superseded" || "archived" || "rejected"`. `drift.go` blindly compares frontmatter against a git-inferred status that can never resolve to a terminal state.
3. **Modern close convention** (mechanisms ① ②): the audit engine never relies on the stale `sync`/`docs(sync)`→completed transition rules; the drift walker does, via `ClassifyPRTitle` → `transitionRules`.

This SPEC brings `moai spec drift` into agreement with `moai spec audit`'s era/grandfather/terminal policy AND fixes the combined-scope + stale-rule classifier defects, so the drift count reflects only GENUINE modern drift (mechanism ⑤).

### A.4 The combined-scope dual remediation (binding user decision)

Mechanism ① is addressed by **BOTH** detector accommodation AND a convention doctrine change (binding user decision, locked):

- **Detector accommodation** (covers historical commits): the walker recognizes combined-scope close subjects (e.g. `docs(SPEC-CCSYNC): …`, `chore(SPEC-CCSYNC): … 4-phase close (CLAUDEMD + TOOLCAT …)`) and maps them to the sibling SPECs they actually close. This is necessary because the historical commits already exist and cannot be rewritten.
- **Convention doctrine change** (prevents recurrence): a doctrine amendment mandates individual full-IDs in close-commit subjects going forward, so future closes do not regenerate this false-positive.

### A.5 Source files (run-phase targets — NOT modified in plan-phase)

- `internal/spec/drift.go` — `DetectDrift`, `getGitImpliedStatus` (N=50 walker window), `commitMatchesSPECID`, `shouldSkipCommitTitle`. NO era awareness — the ④ gap. NO terminal-state authority — the ③ gap. Subject-only `commitMatchesSPECID` — the ① gap.
- `internal/spec/transitions.go` — `transitionRules` slice (the stale `sync`/`docs(sync)`→completed rules = ②), `ClassifyPRTitle`, `closeInfixMatch`, `isSyncPhaseDocs`, `ExtractSPECIDs`/`TransitionSPECIDPattern` (the ① subject-only extraction).
- `internal/spec/era.go` — `ClassifyEra(EraSignals)`, `LoadEraSignalsFromDir`, `Era.IsModern()`, `Era.EraFinal()`, `modernEraThreshold`. drift.go SHOULD reuse this for ④ (READ-ONLY consumption — era.go itself is NOT modified).
- `internal/spec/audit.go` — `Audit()` already applies grandfather + terminal authority; the reference pattern for ④/③ alignment (READ-ONLY — audit.go behavior is NOT changed).
- `.claude/rules/moai/workflow/lifecycle-sync-gate.md` — SSOT for era heuristic + grandfather clause + Status Transition Ownership Matrix close-subject pattern. Doctrine amendment target (⑤doctrine REQ).
- `.claude/rules/moai/development/spec-frontmatter-schema.md` — Status Transition Ownership Matrix close-subject canonical pattern. Doctrine amendment target (⑤doctrine REQ).
- `internal/spec/CLAUDE.md` — module conventions (lint/detector rules observation-only, table-driven tests, sibling PRESERVE).

---

## B. Out of Scope (What NOT to Build)

### Out of Scope — Genuine incomplete-close backfill (mechanism ⑤)

`completed`-status modern V3R6 SPECs that have **NO** close commit at all (mechanism ⑤) are **REAL drift**, not false-positives. The frontmatter claims `completed`/`implemented` but the lifecycle step was never actually performed. This SPEC MUST NOT silently clear them by changing the classifier. The correct remediation is operational (actually run `moai spec close` so a genuine close commit exists), deferred to a separate follow-up SPEC. Any change here that makes the drift walker report a terminal/completed status for a SPEC lacking the corresponding close commit would mask a real lifecycle gap and is an explicit anti-goal (plan.md §G). Run-phase residual classification of ⑤ is informational only (it identifies the deferred set); it does not remediate it.

### Out of Scope — Changing `moai spec audit` behavior

`moai spec audit` already reports these SPECs correctly. The audit engine, its era classification (`era.go`), its grandfather clause, and its terminal-state authority are the REFERENCE, NOT the target. `era.go` and `audit.go` are consumed READ-ONLY; their behavior is not modified. The fix aligns `drift.go` TO the audit policy, not the reverse.

### Out of Scope — Word-boundary SPEC-ID filter regression

The LSGF-001 word-boundary filter (`commitMatchesSPECID` / `ExtractSPECIDs` substring-collision protection, e.g. `HARNESS-001` vs `HARNESS-NAMESPACE-001`) MUST be preserved. The combined-scope sibling-mapping fix (①) extends recognition to abbreviated-scope subjects WITHOUT weakening the word-boundary guard — a regression guard AC (AC-DLC-011) enforces this explicitly.

### Out of Scope — Metadata-sweep skip regression

The existing `chore(spec):` / `chore(specs):` metadata-sweep skip (SPEC-V3R4-LINT-STATUS-CHORE-SKIP-001 AC-LSCSK-003) and the narrow SPEC-ID-scoped backfill-skip (SPEC-V3R6-DRIFT-CONVENTION-ALIGN-001 AC-DCA-003/008) MUST remain green. This SPEC extends the classifier; it does not relax existing skip behavior.

### Out of Scope — New commit-convention introduction (beyond the close-subject full-ID mandate)

The ⑤doctrine amendment mandates individual full-IDs in close-commit subjects (a tightening of an EXISTING convention — the Status Transition Ownership Matrix already uses `SPEC-{ID}` placeholders that imply a single full ID). It does NOT introduce, rename, or deprecate any commit-subject *prefix* or *infix* pattern. The close-infix (`4-phase close` / `Mx-phase audit-ready`) and sync (`sync-phase artifacts`) conventions are unchanged.

### Out of Scope — Reducing all 51 drift records to zero

Only mechanisms ①②③④ (the false-positives) are in scope. Mechanism ⑤ (genuine drift) MUST remain reported. A strict-decrease to the genuine-⑤ residual (not zero) is the directional target — see AC-DLC-009.

### Out of Scope — Go implementation in plan-phase

This is a plan-phase authoring SPEC. No Go code is written here. The fix design lives in plan.md; the `internal/spec` edits and tests are the run-phase deliverable.

---

## C. Requirements (GEARS)

### C.1 Functional Requirements (false-positive elimination)

- **REQ-DLC-001 — combined-scope sibling mapping via secondary prefix-grep fallback** (Event-driven): **When** the per-SPEC primary walk (`git log --grep=<full-specID>`) yields no `completed`/terminal signal AND the SPEC was closed by a combined-scope commit whose subject names only the scope-prefix (a `SPEC-{PREFIX}` token without a trailing `-NNN`, e.g. `SPEC-CCSYNC`), the drift detector shall run a **secondary scope-prefix grep** (`git log --grep=<scope-prefix>`) to locate the combined-scope close commit and resolve the target sibling SPEC to `completed`. The scope-prefix is derived by stripping the trailing distinguishing-segment(s)+number from the full SPEC-ID (`SPEC-CCSYNC-CLAUDEMD-001` → `SPEC-CCSYNC`). The fallback fires ONLY when the candidate combined-scope commit (a) has subject prefix `chore(SPEC-{PREFIX})` / `docs(SPEC-{PREFIX})` where `{PREFIX}` carries NO trailing `-NNN`, AND (b) satisfies `closeInfixMatch(subject) == true`, AND (c) the subject (or a parsed combined-scope group) references this SPEC's distinguishing segment (e.g. `CLAUDEMD`) OR explicitly closes all siblings under that prefix.

> Design rationale (plan-audit D1): the predecessor M3 gate operated on in-window commits only — but `getGitImpliedStatus` fetches commits via `git log --grep=<full-specID>`, so the combined-scope close commit (e.g. `cf7d78a9c chore(SPEC-CCSYNC): … 4-phase close (CLAUDEMD + TOOLCAT …)`) is **never in the per-SPEC window** (it names only `SPEC-CCSYNC`, not `SPEC-CCSYNC-CLAUDEMD-001`). The secondary scope-prefix grep is the only mechanism that can reach it. Verified empirically: `git log --grep=SPEC-CCSYNC-CLAUDEMD-001` returns 4 commits (sync `da2fbcedf` → `implemented`, feat `a97206dc7` → `implemented`, 2 plan), NONE being the close; `git log --grep=SPEC-CCSYNC` DOES contain `cf7d78a9c` (the `4-phase close`). The close subject `chore(SPEC-CCSYNC): Mx-phase 4-phase close (CLAUDEMD + TOOLCAT status→completed, §E.5)` names both distinguishing segments, satisfying gate (c).

- **REQ-DLC-002 — combined-scope word-boundary preservation** (Unwanted behavior): The secondary prefix-grep fallback shall not weaken the LSGF-001 substring-collision protection. The scope-prefix grep is the one real collision-risk site, so it shall be gated tightly: (i) it fires ONLY as a FALLBACK after the exact-token primary walk yields no `completed`/terminal signal; (ii) the candidate combined-scope commit MUST satisfy `closeInfixMatch == true`; (iii) the distinguishing-segment cross-check (gate (c) of REQ-DLC-001) MUST pass, so a `SPEC-CCSYNC-CLAUDEMD-001` does NOT absorb an unrelated `SPEC-CCSYNC-OTHER-002` it does not actually close. The exact-token `commitMatchesSPECID` primary path is unchanged (additive only); a `SPEC-{PREFIX}` token shall map ONLY to siblings that begin with `SPEC-{PREFIX}-` (hyphen-delimited prefix boundary), never to a SPEC that merely contains the prefix substring (e.g. `SPEC-CCSYNC` does NOT map to `SPEC-CCSYNCEXTRA-001`).

- **REQ-DLC-003 — stale sync→completed rule correction** (Ubiquitous): The classifier shall treat sync-phase commits (`sync(SPEC-{ID}): …`, `sync(SPEC-{ID}): … implemented`, `docs(SPEC-{ID}): … sync-phase …`) as resolving to `implemented`, NOT `completed`. The legacy `{"sync"→completed}` and `{"docs(sync)"→completed}` transition rules shall be corrected to the 4-phase model where sync-phase = `implemented` and ONLY a close-infix → `completed`.

- **REQ-DLC-004 — close-infix remains sole completed signal** (Unwanted behavior): After the stale-rule correction, the classifier shall not resolve any SPEC to `completed` from a sync/feat/docs commit — `completed` shall be inferred ONLY via the canonical close-infix (`4-phase close` / `Mx-phase audit-ready`), preserving SPEC-V3R6-DRIFT-CONVENTION-ALIGN-001 anti-pattern AP-2.

- **REQ-DLC-005 — terminal-state authority** (State-driven): **While** a SPEC's frontmatter status is a terminal state (`superseded`, `archived`, or `rejected`), the drift detector shall treat the frontmatter as authoritative and shall not report drift, because no git commit convention can positively infer a terminal state (matching `audit.go` `checkV3R6Drift` terminal-state early-return). **Record-emission contract (D3 normative)**: the detector shall append a `DriftRecord` with `Drifted: false` and `continue` to the next SPEC — it MUST NOT drop the record entirely. This preserves the existing `DetectDrift` append pattern (`drift.go:73-84`, one record per SPEC) so that `Records[]` consumers (e.g., the run-phase M5 residual classifier) still see the SPEC with `Drifted == false`.

- **REQ-DLC-006 — era/grandfather alignment** (Where — capability gate): **Where** a SPEC is classified by `ClassifyEra` as a grandfather-protected era (`Era.EraFinal() == true`: V2.x / V3R2-R4 / V3R5), the drift detector shall exempt it from drift reporting, reusing the existing `LoadEraSignalsFromDir` + `ClassifyEra` from `era.go` so that `moai spec drift` and `moai spec audit` apply the SAME grandfather policy. **Record-emission contract (D3 normative)**: the detector shall append a `DriftRecord` with `Drifted: false` and `continue` — it MUST NOT drop the record. The grandfather-exempt SPEC remains in `Records[]` with `Drifted == false` (same append pattern as REQ-DLC-005), so `Records[]` consumers observe the full SPEC corpus and only the `Drifted` flag distinguishes exempt from drifting.

### C.2 Regression-Preservation Requirements

- **REQ-DLC-007 — metadata-sweep + backfill skip preserved** (Unwanted behavior): The classifier shall not break the existing skip behavior — `chore(spec):` / `chore(specs):` metadata-sweep commits (AC-LSCSK-003) and the narrow SPEC-ID-scoped backfill-skip without close-infix (AC-DCA-003/008) shall continue to be skipped unchanged.

- **REQ-DLC-008 — word-boundary + audit unchanged** (Ubiquitous): The classifier shall preserve the LSGF-001 word-boundary SPEC-ID filter behavior, and the era-consumption change shall not alter `moai spec audit` output (audit.go / era.go are consumed READ-ONLY).

### C.3 Verification Requirements

- **REQ-DLC-009 — drift count strictly decreases to genuine residual** (Event-driven): **When** the fix is implemented and `moai spec drift --count` is run against this repository, the reported count shall strictly decrease from the verified baseline (51 drifted records / `--count` 53 at authoring) toward the genuine-⑤ residual. The binding success signal is the per-mechanism named-exemplar transition (AC-DLC-001..004), not a raw count band; the expected residual is stated as a range in AC-DLC-009, to be finalized when run-phase classifies ⑤.

- **REQ-DLC-010 — audit stays clean** (Ubiquitous): The `moai spec audit` result shall remain unchanged (no new MUST-FIX finding introduced, no existing finding removed, by the drift detector change).

- **REQ-DLC-011 — close-subject doctrine amendment** (Ubiquitous): The doctrine SSOT files (`.claude/rules/moai/workflow/lifecycle-sync-gate.md` Status Transition Ownership Matrix + `.claude/rules/moai/development/spec-frontmatter-schema.md` Status Transition Ownership Matrix) shall mandate individual full-IDs in close-commit subjects, prohibiting combined/abbreviated scope (e.g. `chore(SPEC-CCSYNC): …`) in favor of one full SPEC-ID per close commit (e.g. `chore(SPEC-CCSYNC-CLAUDEMD-001): … 4-phase close`), so combined-scope false-positives do not recur. The template mirror obligation applies (see plan.md §H Mirror Obligation).

---

## 3. Acceptance Criteria (measurable, REQ↔AC bidirectional traceability)

> All AC are run-phase verifiable. This plan-phase SPEC defines them as the implementation oracle. Where an AC names a real-repo exemplar, the run-phase MUST re-verify the exemplar's git log freshly before asserting (per `internal/spec/CLAUDE.md` sibling-PRESERVE discipline); if an exemplar was re-touched such that its newest exact-token commit changed, substitute a re-verified equivalent of the same mechanism.

### AC-DLC-001 — combined-scope sibling SPECs transition DRIFT → aligned — verifies REQ-DLC-001, REQ-DLC-002

> Rebound per plan-audit D1: the binding signal is the **deterministic synthetic fixture** (immune to repo churn) PLUS a **live-repo integration proof** for the motivating CCSYNC case. The synthetic fixture asserts the secondary prefix-grep fallback resolves BOTH siblings of one combined-scope close.

- **Live-repo integration proof (the case the SPEC exists for)** — **Given** the repository at the post-fix state, **When** `moai spec drift --json` is executed, **Then** BOTH `SPEC-CCSYNC-CLAUDEMD-001` AND `SPEC-CCSYNC-TOOLCAT-001` MUST be absent from the DRIFT set (`Drifted == false`). The detector's secondary scope-prefix grep (`git log --grep=SPEC-CCSYNC`) locates the combined-scope close commit `cf7d78a9c` (`chore(SPEC-CCSYNC): Mx-phase 4-phase close (CLAUDEMD + TOOLCAT status→completed, §E.5)`), passes the distinguishing-segment cross-check for each sibling, and resolves both to `completed`.
- **Deterministic fixture-backed unit AC (BINDING)** — a synthetic git-log fixture (newest-first) `chore(SPEC-ABC): Mx-phase audit-ready signal + 4-phase close (FOO + BAR)`, then `feat(SPEC-ABC-FOO-001): M1`, then `feat(SPEC-ABC-BAR-001): M1`:
  - With target `specID = SPEC-ABC-FOO-001`: the per-`SPEC-ABC-FOO-001` primary walk yields no `completed` (only the feat → `implemented`); the secondary prefix-grep (`SPEC-ABC`) finds the close-infix combined-scope commit referencing `FOO` → resolves `completed`.
  - With target `specID = SPEC-ABC-BAR-001`: same fallback resolves `completed` (the combined close references `BAR`).
  - Both siblings of ONE combined-scope close resolve to `completed` — this is the binding assertion.

### AC-DLC-012 — non-sibling partial-prefix SPEC is NOT falsely mapped (LSGF-001 collision guard) — verifies REQ-DLC-002

> New per plan-audit D1: the secondary prefix-grep is the one collision-risk site; this AC pins that an unrelated SPEC sharing a partial prefix is NOT absorbed.

- **Given** a synthetic git-log fixture (newest-first) `chore(SPEC-ABC): Mx-phase audit-ready signal + 4-phase close (FOO + BAR)`, then `feat(SPEC-ABC-OTHER-002): M1` (a SPEC under the same `SPEC-ABC` prefix but NOT named in the combined close), and separately a `SPEC-ABCD-001` (a SPEC whose prefix merely CONTAINS `ABC` as a substring, different hyphen boundary)
- **When** the walker resolves git-implied status for `specID = SPEC-ABC-OTHER-002` and for `specID = SPEC-ABCD-001`
- **Then** NEITHER is resolved to `completed` by the combined-scope fallback: `SPEC-ABC-OTHER-002` fails the distinguishing-segment cross-check (the close names only `FOO`/`BAR`, not `OTHER`) so it remains at its primary-walk status; `SPEC-ABCD-001` fails the hyphen-delimited prefix boundary (`SPEC-ABC-` does not prefix-match `SPEC-ABCD-001`) so the `SPEC-ABC` scope-prefix grep does not even apply to it. No false sibling mapping occurs.

### AC-DLC-002 — stale sync→completed rule corrected to implemented — verifies REQ-DLC-003, REQ-DLC-004

- **Given** the post-fix classifier
- **When** the commit subject `sync(SPEC-EXAMPLE-001): lifecycle complete — v0.3.0 implemented` is classified
- **Then** the resolved status is `implemented` (NOT `completed`); AND the real-repo exemplar `SPEC-V3R5-STATUSLINE-STDINFIELDS-001` (frontmatter `implemented`) is absent from the DRIFT set after the fix.
- **Anti-regression sub-assertion**: a `feat(SPEC-EXAMPLE-001): M1` commit still resolves `implemented`, and NO commit subject without a close-infix resolves `completed`.

### AC-DLC-003 — terminal-state frontmatter is authoritative — verifies REQ-DLC-005

- **Given** SPECs whose frontmatter is a terminal state
- **When** `moai spec drift --json` is executed
- **Then** the terminal-state exemplars `SPEC-V3R3-HARNESS-001` (superseded), `SPEC-V3R6-AGENT-FOLDER-SPLIT-001` (superseded), `SPEC-V3R6-AGENT-MODEL-ROUTING-001` (archived) MUST be absent from the DRIFT set (no drift reported for terminal-state frontmatter).
- **Fixture-backed unit AC**: a SPEC fixture with frontmatter `status: superseded` (and `status: archived`, `status: rejected`) and any git-inferred non-terminal status → `Drifted == false`.

### AC-DLC-004 — grandfathered eras are exempt — verifies REQ-DLC-006

- **Given** the era classification `ClassifyEra` and the post-fix drift detector
- **When** `moai spec drift --json` is executed
- **Then** grandfather-protected-era exemplars (`SPEC-V3R2-SPC-002` [V3R2/V3R3-R4 era], `SPEC-V3R3-PROJECT-HARNESS-001`) MUST be absent from the DRIFT set; the drift detector calls `LoadEraSignalsFromDir` + `ClassifyEra` and skips any SPEC with `Era.EraFinal() == true`.
- **Fixture-backed unit AC**: a SPEC fixture with no `progress.md` (H-1 → V2.x, `EraFinal == true`) and a deliberate frontmatter↔git mismatch → `Drifted == false`; AND a V3R6 fixture (H-4 §E.2+§E.5+both SHAs) with a genuine mismatch → `Drifted == true` (modern era still detected).
- **Era subtlety guard** (plan.md §G AP-3): the exemption MUST rely on the H-1..H-4 progress.md signals as `ClassifyEra` computes them — NOT on a standalone `created` date comparison. (`SPEC-V3R2-SPC-002` has `created: 2026-04-23`, which is after `modernEraThreshold` 2026-04-01; H-1/H-2/H-3 progress.md signals must classify it as a grandfathered era BEFORE the H-5 date tie-breaker fires. The fixture-backed unit AC above pins this.)

### AC-DLC-005 — metadata-sweep + backfill skip regression green — verifies REQ-DLC-007

- **Given** the existing tests `drift_chore_skip_test.go` (LSCSK-001 / AC-LSCSK-003) and the DRIFT-CONVENTION-ALIGN backfill-skip tests
- **When** `go test ./internal/spec/...` is run after the fix
- **Then** all chore-skip + backfill-skip regression tests pass with no behavioral change.

### AC-DLC-006 — word-boundary filter unchanged — verifies REQ-DLC-008 (LSGF-001 part)

- **Given** the LSGF-001 word-boundary tests `drift_specid_grep_test.go` (`TestGetGitImpliedStatus_SPECIDWordBoundary` 5 sub-cases + `TestGetGitImpliedStatus_HARNESS001Resolution`)
- **When** `go test ./internal/spec/...` is run after the fix
- **Then** the word-boundary tests pass unchanged (substring-collision protection preserved — `HARNESS-001` does NOT match `HARNESS-NAMESPACE-001`).

### AC-DLC-007 — audit output unchanged — verifies REQ-DLC-008 (audit part), REQ-DLC-010

- **Given** `moai spec audit --json` before and after the fix
- **When** both are compared
- **Then** the audit output is identical (same Grandfathered count, same ModernEraClean count, same DriftFindings set) — the era-consumption change reuses `era.go` READ-ONLY and does not alter `audit.go` behavior.

### AC-DLC-008 — no false-positive on grandfathered sibling SPECs — verifies REQ-DLC-006, REQ-DLC-002 (module PRESERVE convention)

- **Given** a sibling-mapping + era-exemption fixture using a grandfathered-era SPEC whose status was set before this detector shipped (per `internal/spec/CLAUDE.md` sibling-PRESERVE: use a V3R2/V3R3 grandfathered fixture OR ARR-001/COORD-001)
- **When** the post-fix drift detector runs against the fixture
- **Then** the grandfathered sibling SPEC is NOT newly flagged as drift by the combined-scope mapping or any other new code path (no regression introduced against pre-existing closed/grandfathered SPECs).

### AC-DLC-009 — drift count strictly decreases to genuine-⑤ residual — verifies REQ-DLC-009

- **Given** the verified baseline `moai spec drift --json` → 51 drifted records (`--count` 53 at authoring)
- **When** `moai spec drift --count` is executed at the post-fix state
- **Then** the count is **strictly less than the baseline** and approaches the genuine-⑤ residual. Directional expectation (NOT a binding band): after removing ① (~2), ② (~3), ③ (~3+), ④ (~14: V3R2 ×6 + V3R3 ×6 + V3R4 ×1 + DESIGN ×1), the residual is the genuine-⑤ modern-drift set (run-phase finalizes the exact value). Only strict decrease is asserted; the per-mechanism named-exemplar ACs (001–004) are the binding success signal.

### AC-DLC-010 — full package test suite green — verifies all REQ (gate)

- **Given** the post-fix `internal/spec` package
- **When** `go test ./internal/spec/...` is run
- **Then** the suite is green with no skipped or failing tests attributable to this change, and coverage for the modified files does not regress below the package baseline (85%).

### AC-DLC-011 — close-subject doctrine amendment present + mirrored — verifies REQ-DLC-011

- **Given** the doctrine SSOT files after the fix
- **When** the tightened oracle is run (D2 — requires the PROHIBITION co-located with the combined/abbreviated-scope subject, not a bare substring that passes on incidental prose): a per-file check that the amendment paragraph contains BOTH a combined/abbreviated-scope reference AND a prohibition verb in proximity, e.g.

  ```bash
  # PASS requires: (combined|abbreviated scope) AND (MUST use individual full-ID | prohibited | disallowed) in the same amendment block.
  for f in .claude/rules/moai/workflow/lifecycle-sync-gate.md \
           .claude/rules/moai/development/spec-frontmatter-schema.md; do
    grep -Pzo '(?s)(combined|abbreviated)[^\n]*scope.{0,400}?(MUST use individual full-ID|prohibited|disallowed)' "$f" \
      || grep -Pzo '(?s)(MUST use individual full-ID|prohibited|disallowed).{0,400}?(combined|abbreviated)[^\n]*scope' "$f" \
      || { echo "FAIL: $f missing co-located prohibition"; exit 1; }
  done
  ```
- **Then** both files contain the amendment where the combined/abbreviated-scope phrase and the prohibition (`MUST use individual full-ID` / `prohibited` / `disallowed`) are co-located within the same amendment block (≤400 chars apart, either order) — an incidental mention of "full SPEC-ID" elsewhere does NOT satisfy the oracle; AND if `.claude/rules/` is template-mirrored, the corresponding `internal/template/templates/.claude/rules/...` mirror is byte-consistent (template-mirror obligation — see plan.md §H; run-phase verifies whether these two specific files are mirrored and, if so, updates the mirror with the SPEC-ID cross-reference generalized per §25 forbidden-class C2).

### REQ ↔ AC Traceability Matrix

| REQ | Covered by AC |
|-----|---------------|
| REQ-DLC-001 | AC-DLC-001 |
| REQ-DLC-002 | AC-DLC-001, AC-DLC-008, AC-DLC-012 |
| REQ-DLC-003 | AC-DLC-002 |
| REQ-DLC-004 | AC-DLC-002 |
| REQ-DLC-005 | AC-DLC-003 |
| REQ-DLC-006 | AC-DLC-004, AC-DLC-008 |
| REQ-DLC-007 | AC-DLC-005 |
| REQ-DLC-008 | AC-DLC-006, AC-DLC-007 |
| REQ-DLC-009 | AC-DLC-009 |
| REQ-DLC-010 | AC-DLC-007 |
| REQ-DLC-011 | AC-DLC-011 |

Every REQ maps to ≥1 AC; every AC traces back to ≥1 REQ. (11 REQ × 12 AC: AC-DLC-001..011 + the new AC-DLC-012 non-sibling-prefix collision guard. Mechanisms ①②③④ each carry ≥1 fixture-backed AC; mechanism ① (combined-scope) additionally carries AC-DLC-012 LSGF-001 collision guard; ⑤doctrine carries the doctrine-content AC-DLC-011.)
