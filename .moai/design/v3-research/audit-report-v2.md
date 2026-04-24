# v3.0.0 Plan Audit Report — Iteration 2

> Auditor: plan-auditor (independent, adversarial)
> Iteration: 2
> Date: 2026-04-23
> Previous report: audit-report.md (28 defects, NOT-READY)
> Bias-prevention: M1–M6 active. Reasoning context ignored per M1.

---

## Executive Summary

- **Verdict: NOT-READY**
- Iteration 1 defects remediation coverage: **24 PASS / 2 PARTIAL / 2 observation-only (out of 28)**
- Residual defects from iteration 1: **2** (D-CRIT-005 partial, D-MED-008 partial)
- New defects introduced by remediation: **12**
  - Critical: 2
  - High: 4
  - Medium: 4
  - Low: 2
- Total active defects: **14** (2 partial iteration-1 + 12 new regressions)

The remediation resolved most surface defects cleanly. The master document (major-v3-master.md) and frontmatter normalization passes are **largely successful**: all 28 SPECs use uniform `bc_id: [...]` array form, uniform `P0..P3` priority vocabulary, uniform `"v3.0.0 — Phase N <name>"` phase format, and all CLN-003 references are purged. Master §3.1, §6.1, §6.2, §8.1 HOOKS numbering and §8.7 SPEC-001 slot assignment are now coherent. Four previously-ACless SPECs (CLN-001, CLN-002, MIGRATE-001, OUT-001) now carry 39 inline Given/When/Then AC-IDs total (matches claim exactly).

However, the D-CRIT-005 (MIG-002 ↔ CLN-001/002 scope overlap) remediation was **only half-completed**. The master design §3.9 added a clear Ownership Split clause stating that migration step Go files (`internal/core/migration/steps/m0*.go`) live in MIG-002 and CLN-001/002 own only diagnostic tooling — but CLN-001 and CLN-002 SPEC bodies still claim ownership of those same files in four separate places each (§2.1 In Scope, §3 Environment, §5 REQs, §10 Traceability paths). The scope overlap is now explicitly documented at the master level and explicitly contradicted at the SPEC level.

Additionally, the MIG-002 absorption of M02 agency archival + docs-site 4-locale sync introduced internal inconsistency within MIG-002 itself: M02's archive destination path is specified four different ways across §2.1, §5.7 (new REQ), §6 (pre-existing AC), and §6 (new AC); the locale sync is listed as both In Scope and Out of Scope in the same SPEC. The new REQs use a compound `REQ-MIG-002-M02-NNN` / `REQ-MIG-002-LOC-NNN` subdomain prefix pattern that breaks the single-word domain convention used by all 28 other SPECs. Traceability for those 10 new REQs is absent from MIG-002 §10 table.

AC addition in the four formerly-ACless SPECs hit the claimed count (10/9/12/8 = 39) but covers only ~50% of the REQs in those SPECs (e.g., CLN-001 has 23 REQs, only 10 unique REQs are mapped by its 10 ACs; 13 REQs are uncovered). This is a traceability regression from "0% AC coverage" (blocker) to "50% AC coverage" (still below quality bar for Phase 1 run).

**Iteration 3 is required.** The residual work is bounded and concrete — see Recommendation section.

---

## Iteration 1 Defect Resolution Verification

### CRITICAL defects

#### D-CRIT-001 — SPEC-V3-CLN-003 purge
- **Status: PASS**
- **Evidence**: `grep -rn "CLN-003" .moai/specs/ docs/design/ → 0 matches` across all SPECs and master. Directory listing confirms only 28 SPEC-V3-* directories. Master §3.9 §8.7 both explicitly note "The retired third CLN slot … has been absorbed into SPEC-V3-MIG-002."
- **Residual**: None.

#### D-CRIT-002 — Four SPECs lack structured AC-IDs
- **Status: PASS (count) / PARTIAL (traceability coverage)**
- **Evidence**:
  - CLN-001: 10 AC-IDs (lines 191-200), matches claim
  - CLN-002: 9 AC-IDs (lines 209-217), matches claim
  - MIGRATE-001: 12 AC-IDs (lines 236-247), matches claim
  - OUT-001: 8 AC-IDs (lines 196-203), matches claim
  - Total: 39, matches claim
  - Each AC uses Given/When/Then + `(maps REQ-...)` structure ✓
- **Residual (NEW DEFECT D-NEW-005)**: REQ→AC coverage is only ~50%. CLN-001 23 REQs / 10 mapped. CLN-002 20 / 10. MIGRATE-001 25 / 13. OUT-001 22 / 12. 45 REQs total uncovered across the four SPECs. This is a traceability regression promoted from the original "blocked by AC absence" state.

#### D-CRIT-003 — HOOKS numbering drift
- **Status: PASS**
- **Evidence**:
  - Master §3.1 (lines 237-243) lists HOOKS-001..006 matching on-disk
  - Master §6.1 Phase 2 (line 1039) lists 5 HOOKS (001, 003, 004, 005, 006); HOOKS-002 correctly deferred to Phase 6a
  - Master §6.2 Phase 6a (line 1059) references "SPEC-V3-HOOKS-002 (absorbs type:prompt/agent/http)"
  - Master §8.1 (line 1102) "Hooks/Commands SPECs (9 SPECs: 6 HOOKS + 3 CMDS)" and explicitly notes absorption
  - HOOKS-007/008/009 references: grep returns 0 matches in master
- **Residual**: None.

#### D-CRIT-004 — bc_id frontmatter type drift
- **Status: PASS**
- **Evidence**: All 28 SPECs use array form. Breaking SPECs populate it (HOOKS-001 `[BC-001]`, HOOKS-005 `[BC-005, BC-007]`, HOOKS-006 `[BC-003]`, MIG-001 `[BC-004]`, MIGRATE-001 `[BC-004]`, SCH-001 `[BC-006]`, SCH-002 `[BC-002, BC-003]`, TEAM-001 `[BC-008]`). Non-breaking SPECs use `bc_id: []` (MIG-002, PLG-001, all 6 CMDS/CLN/OUT/SKL non-breaking, all 3 AGT, etc.).
- **Residual**: None. Zero scalar/null/omitted forms remain.

#### D-CRIT-005 — MIG-002 vs CLN-001/002 scope overlap
- **Status: PARTIAL**
- **Evidence for PASS**:
  - Master §3.9 Ownership Split subsection added (lines 888-896) with explicit rules: "SPEC-V3-MIG-002 owns ALL migration step Go implementations in `internal/core/migration/steps/m0*.go`. SPEC-V3-CLN-001 owns surrounding tooling. … CLN-001 MUST NOT implement migration step Go files directly — those live in MIG-002."
  - MIG-002 §2.2 Out of Scope line 140: "Diagnostic CLI tooling … owned by SPEC-V3-CLN-001/002."
  - CLN-001 §2.2 Out of Scope line 75: "Migration step Go implementations … owned by SPEC-V3-MIG-002. CLN-001 calls into MIG-002's migrate functions but does not implement them."
  - CLN-002 §2.2 Out of Scope line 88: "Migration step Go implementations — owned by SPEC-V3-MIG-002."
- **Evidence for PARTIAL**: boundary markers added but contradictory claims retained in-place. Four surfaces still claim ownership by CLN-001/002:
  1. CLN-001 §2.1 In Scope lines 59-61: "MigrationStep 구현체 M01 (`m01_template_version_sync.go`): …", "M03 (`m03_hook_wrapper_drift.go`): …", "M04 (`m04_skill_drift.go`): …"
  2. CLN-001 §3 Environment line 84: "신설: `internal/core/migration/steps/m01_template_version_sync.go`, `m03_hook_wrapper_drift.go`, `m04_skill_drift.go`"
  3. CLN-001 §5 REQ-CLN-001-001/003/004 language: "The M01 migration step **shall** …", "The M03 migration step **shall** …", "The M04 migration step **shall** …"
  4. CLN-001 §10 Traceability lines 253-256: lists `internal/core/migration/steps/m01_template_version_sync.go`, `m03_hook_wrapper_drift.go`, `m04_skill_drift.go` as implementation paths
  Same for CLN-002: §2.1 line 68 "M05 (Migration Step — MigrationStep 구현체 `m05_legacy_cleanup.go`)"; §3 line 106 "신설: `internal/core/migration/steps/m05_legacy_cleanup.go`"; §10 line 271.
- **Residual**: CLN-001 and CLN-002 must have their In-Scope, Environment, REQ text, and Traceability paths edited to match their Out-of-Scope boundary. See D-NEW-001 and D-NEW-002.

### HIGH defects

#### D-HIGH-001 — priority vocabulary
- **Status: PASS** — all 28 SPECs use P0/P1/P2/P3 form. Distribution: P0 Critical×5, P1 High×13, P2 Medium×6, P3 Low×3, remaining 1. (No "Critical"/"High" standalone anywhere.)

#### D-HIGH-002 — phase format
- **Status: PASS** — all 28 SPECs prefix with `"v3.0.0 —"` followed by Phase N and name.

#### D-HIGH-003 — SKL-002 phase
- **Status: PASS** — SKL-002 phase is "v3.0.0 — Phase 5 Internal Cleanup". Master §6.2 Phase 5 row (line 1079) lists "SPEC-V3-CLN-001, SPEC-V3-CLN-002, SPEC-V3-SKL-002" explicitly.

#### D-HIGH-004 — §8.7 slot
- **Status: PASS** — §8.7 renamed "Internal Cleanup + moai-unique SPECs (3 SPECs)" with CLN-001, CLN-002 under "Internal Cleanup" and SPEC-001 under "moai-unique". Explicit note about CLN-003 retirement.

#### D-HIGH-005 — MIG-001 §9.2 inflated
- **Status: PASS** — HOOKS-001~006 moved from §9.2 Blocks to §9.3 Related (line 341) with explicit note "(확정 시 각 HOOKS SPEC §9.1 Blocked by로 이동)".

#### D-HIGH-006 — REQ count
- **Status: PASS** — actual count is 545 (unique REQ IDs across 28 SPECs); master §1 line 13 and §8 table line 1198 both state 545. Footer note line 1202: "(REQ counts auto-derived via `grep -c '^- \*\*REQ-' .moai/specs/SPEC-V3-*/spec.md`; pre-Stage 2 baseline 545.)".

#### D-HIGH-007 — false positive (AGT-003 symmetry)
- **Status: N/A** — iteration 1 reclassified as false positive.

#### D-HIGH-008 — Phase 6a HOOKS-007/008/009
- **Status: PASS** — resolved via CRIT-003. §6.2 now correctly references HOOKS-002 only.

#### D-HIGH-009 — false positive (SCH-001 ↔ MIG-001)
- **Status: N/A** — iteration 1 reclassified as false positive.

#### D-HIGH-010 — OUT-001 §9.1 "비차단"
- **Status: PASS** — SCH-001 moved to §9.3 Related (line 244). §9.1 Blocked by now empty with explanatory note.

#### D-HIGH-011 — CLI-001 under-declared deps
- **Status: PASS** — §9.1 Blocked by (lines 245-247): SPEC-V3-PLG-001, SPEC-V3-SCH-001, SPEC-V3-MIG-001. §9.3 Related (lines 256-257): SPEC-V3-HOOKS-001, SPEC-V3-AGT-001 (both with "Related not Blocked because CLI can ship with subcommand disabled" rationale).

#### D-HIGH-012 — BC-007 orphan
- **Status: PASS** — HOOKS-005 frontmatter line 27: `bc_id: [BC-005, BC-007]`. Confirmed.

### MEDIUM defects

#### D-MED-001 — bc_id non-breaking missing
- **Status: PASS** — subsumed by D-CRIT-004; all non-breaking SPECs use `bc_id: []`.

#### D-MED-002 — dependencies shape
- **Status: PASS** — all 28 SPECs use YAML list form (or `dependencies: []` for SCH-001 which has none).

#### D-MED-003 — 12 risks vs 8
- **Status: N/A** — observation only, no fix required.

#### D-MED-004 — §6.1 vs §6.2 HOOKS split
- **Status: PASS** — subsumed by D-CRIT-003. §6.1 Phase 2 now says "Phase 2 lands 5 HOOKS SPECs" (line 1039).

#### D-MED-005 — SPEC-001 placeholder syntax
- **Status: PASS** — SPEC-V3-SPEC-001 uses `REQ-{DOMAIN}-{NNN}-{NNN}` curly-brace form (line 231) instead of literal regex-matching `REQ-DOMAIN-NNN-NNN`.

#### D-MED-006 — gap severity labeling
- **Status: UNVERIFIED** — low priority, deferred. No blocking impact.

#### D-MED-007 — MIG-002 related_gap integers
- **Status: PASS** — MIG-002 related_gap now uses `- gm#183` list form (lines 16-22 of frontmatter).

#### D-MED-008 — CMDS theme orphan
- **Status: PARTIAL**
- **Evidence for PASS**: Master §3.10 Theme 10 section added (line 896) with Scope, Breaking, SPECs, Rationale subsections.
- **Evidence for PARTIAL**: SPEC-V3-CMDS-001 frontmatter line 22 still says `related_theme: "Theme 4: Agent Frontmatter Expansion (command parity)"`. CMDS-002 says `"Command argument substitution parity"`. CMDS-003 says `"Command inline shell execution parity"`. None reference "Theme 10". See D-NEW-010.

### LOW defects

#### D-LOW-001 — em-dash
- **Status: PASS** — all Phase headers in master §6.1 lines 1034-1067 use em-dash.

#### D-LOW-002 — SKL-001 AC style
- **Status: NOT ACTIONABLE** — iteration 1 noted SKL-001 already uses `- **AC-SKL-001-01**:` bold form; no change required.

#### D-LOW-003 — glossary
- **Status: PASS** — Appendix B now includes "Tier 1 / Tier 2 / Tier 3 / Tier 4" (line 1340), "Harness level" (line 1341), "Sprint Contract" (line 1342) with full definitions.

#### D-LOW-004 — NNN placeholder
- **Status: PASS** — CLN-001 line 250 now uses `REQ-CLN-001-{REQ_NUM}` with clarifying note.

---

## New Defects (regression / introduced)

### CRITICAL

#### D-NEW-001 — CLN-001 self-contradicts on migration step ownership
- **Dimension**: D12 (Internal Consistency), D9 (Scope Discipline)
- **Severity: CRITICAL** (direct contradiction inside a single SPEC)
- **Location & Evidence**:
  - CLN-001 §2.1 In Scope (lines 59-61): claims ownership of `m01_template_version_sync.go`, `m03_hook_wrapper_drift.go`, `m04_skill_drift.go` — "MigrationStep 구현체 M01/M03/M04"
  - CLN-001 §2.2 Out of Scope (line 75): "Migration step Go implementations in `internal/core/migration/steps/` — owned by SPEC-V3-MIG-002. CLN-001 calls into MIG-002's migrate functions but does not implement them."
  - CLN-001 §3 Environment (line 84): "신설: `internal/core/migration/steps/m01_template_version_sync.go`, `m03_hook_wrapper_drift.go`, `m04_skill_drift.go`"
  - CLN-001 §5 REQ-CLN-001-001 (line 108): "The M01 migration step **shall** `project.yaml.template_version` 값을 읽고…"; REQ-CLN-001-003 (line 114): "The M03 migration step **shall** …"; REQ-CLN-001-004 (line 117): "The M04 migration step **shall** …"
  - CLN-001 §10 Traceability (lines 253-256): lists migration step Go files as CLN-001's implementation paths
- **Impact**: D-CRIT-005 scope overlap is declared resolved at master level but concretely unresolved at SPEC body level. Wave 5 planning will still face ambiguity: if CLN-001 requires the M01/M03/M04 migration step go files per its own §2.1 and §5 REQs, MIG-002 also requires them per its §2.1 line 77-84, they both cannot own the same files in exclusive fashion.
- **Remediation**: Edit CLN-001 §2.1 to remove ownership of m01/m03/m04 Go files (relocate to "consumes" language). Edit §3 Environment to list migration step files as "dependency" not "신설". Rewrite REQ-CLN-001-001/003/004 to focus on tooling (`moai doctor template-drift` behavior) rather than migration step internals. Update §10 Traceability to point to tooling files (`internal/cli/doctor_template_drift.go`) instead of migration step files.

#### D-NEW-002 — CLN-002 self-contradicts on M05 migration step ownership
- **Dimension**: D12, D9
- **Severity: CRITICAL**
- **Location & Evidence**:
  - CLN-002 §2.1 In Scope (line 68): "**M05 (Migration Step — MigrationStep 구현체 `m05_legacy_cleanup.go`):**" followed by implementation behavior specifications
  - CLN-002 §2.2 Out of Scope (line 88): "Migration step Go implementations — owned by SPEC-V3-MIG-002."
  - CLN-002 §3 Environment (line 106): "영향 파일 (신설): `internal/core/migration/steps/m05_legacy_cleanup.go`, `internal/core/migration/steps/m05_test.go`"
  - CLN-002 §10 Traceability (lines 271-272): lists `internal/core/migration/steps/m05_legacy_cleanup.go` as implementation path
- **Impact**: Same as D-NEW-001 for the M05 case.
- **Remediation**: Same pattern — remove M05 Go file ownership claims; keep ownership only of tooling (`moai doctor legacy-cleanup` CLI) and direct source edits (embed.go comment fix, deps.go comment add, .gitignore updates).

### HIGH

#### D-NEW-003 — MIG-002 M02 archive destination path specified four ways
- **Dimension**: D1 (EARS Compliance), D12 (Internal Consistency)
- **Severity: HIGH**
- **Location & Evidence**:
  - MIG-002 §2.1 line 97: "Apply: `~/.moai/history/v2.12/{commands,rules,backups}/`로 이동 (delete 아님)"
  - MIG-002 §5.7 REQ-MIG-002-M02-001 (line 273): "`.agency/` 하위의 모든 디렉터리를 `docs/archive/agency-v2/` 대상 경로와 `~/.moai/history/v2.12/` 양쪽으로 이동"
  - MIG-002 §6 AC-MIG-002-02 (line 310): verifies `~/.moai/history/v2.12/commands/agency/migrate.md` — single path
  - MIG-002 §6 AC-MIG-002-M02-01 (line 332): verifies `docs/archive/agency-v2/` — different single path
- **Impact**: The same M02 step has four different specifications of archive destination — `~/.moai/history/`, `docs/archive/agency-v2/`, both, or varies by pre-existing vs absorbed AC. Run-phase implementer will choose one arbitrarily and fail the other ACs.
- **Remediation**: Pick ONE canonical archive path (recommend `docs/archive/agency-v2/` per master §3.9 absorbed scope statement). Update §2.1, §5.7, §6, and all REQs/ACs to use that single path consistently.

#### D-NEW-004 — MIG-002 locale sync is both In Scope and Out of Scope
- **Dimension**: D9, D12
- **Severity: HIGH**
- **Location & Evidence**:
  - MIG-002 §2.1 line 127 IN SCOPE: "**docs-site 4-locale sync** — sync `docs-site/content/{en,ko,ja,zh}/` on version bump via explicit migration step (absorbed scope; implementation lives in `m02_agency_archive.go` locale sync helper)."
  - MIG-002 §2.2 line 134 OUT OF SCOPE: "docs-site locale sync (gm#194) — manager-docs 위임, not migration"
  - MIG-002 §5.8 REQ-MIG-002-LOC-001..004 (lines 292-301): locale sync specified as requirements (In Scope)
  - MIG-002 §6 AC-MIG-002-LOC-01/02 (lines 338-340): locale sync as acceptance tests (In Scope)
- **Impact**: Single SPEC lists locale sync as both IN and OUT of scope. Wave 5 cannot decide whether manager-docs or manager-ddd owns locale sync implementation. CI cannot assign the locale sync file to the correct owning SPEC.
- **Remediation**: Pick one. The absorbed-scope argument supports In Scope; in that case delete line 134 from §2.2. The "manager-docs 위임" argument supports Out of Scope; in that case delete §5.8 REQ-MIG-002-LOC-* and §6 AC-MIG-002-LOC-* and create a new SPEC-V3-DOCS-001 for the docs team.

#### D-NEW-005 — Traceability coverage regression in remediated SPECs (~50% REQ→AC)
- **Dimension**: D5 (Traceability)
- **Severity: HIGH**
- **Location & Evidence**:
  - CLN-001: 23 REQs, 10 ACs, 10 REQs mapped → 13 REQs uncovered (001, 002, 003, 004, 005, 006, 007, 012, 015, 016 mapped; 008, 009, 010, 011, 013, 014, 017, 018, 019, 020, 021, 022, 023 uncovered)
  - CLN-002: 20 REQs, 9 ACs, 10 REQs mapped → 10 REQs uncovered
  - MIGRATE-001: 25 REQs, 12 ACs, 13 REQs mapped → 12 REQs uncovered
  - OUT-001: 22 REQs, 8 ACs, 12 REQs mapped → 10 REQs uncovered
  - Total: 45 REQs across the four SPECs have no corresponding AC
- **Impact**: D-CRIT-002 iteration 1 was "0% AC coverage → blocker"; iteration 2 is "~50% AC coverage → Run phase will fail Re-planning Gate trigger for those uncovered REQs". plan-auditor dimension D5 scoring for these SPECs remains below 0.75.
- **Remediation**: Add ACs for the remaining REQs OR prune REQs that are not binary-testable / not required. Target 1:1 or 1:N REQ→AC minimum coverage.

#### D-NEW-012 — MIG-002 REQ IDs use non-standard subdomain prefix + incomplete traceability
- **Dimension**: D2 (REQ Convention), D5 (Traceability), D11 (Frontmatter Schema Conformance)
- **Severity: HIGH**
- **Location & Evidence**:
  - MIG-002 §5.7 lines 272-287: REQ-MIG-002-M02-001 through M02-006 (6 REQs)
  - MIG-002 §5.8 lines 292-301: REQ-MIG-002-LOC-001 through LOC-004 (4 REQs)
  - Compare to other 27 SPECs: all use `REQ-<DOMAIN>-<NNN>-<NNN>` — single domain token, numeric only after
  - MIG-002 §10 Traceability table (lines 398-418): rows cover REQ-MIG-002-001 through 051 only. Zero rows for REQ-MIG-002-M02-* or REQ-MIG-002-LOC-*.
- **Impact**:
  - Any regex built against `REQ-[A-Z]+-\d+-\d+` misses 10 REQs in MIG-002
  - SPEC-V3-SCH-001 schema validation will reject these 10 REQs if it enforces the standard pattern
  - Master §8 auto-derived REQ count footer `grep -c '^- \*\*REQ-'` returns 545 but this 545 includes the 10 non-standard REQs — if tooling parses them out, actual usable REQ count drops to 535
  - MIG-002's own §10 Traceability section is incomplete for the 10 new REQs, meaning they have no designated implementation file or verification test
- **Remediation**: Convert to flat sequential numbering. Renumber REQ-MIG-002-M02-001 → REQ-MIG-002-052, M02-002 → 053, …, LOC-001 → 058, …, LOC-004 → 061. Update §10 Traceability table to include all 10. Update AC-MIG-002-M02-01/02/03 and AC-MIG-002-LOC-01/02 `(maps REQ-...)` back-references accordingly.

### MEDIUM

#### D-NEW-006 — CLN-001 §3 Environment contradicts scope boundary
- **Dimension**: D12
- **Severity: MEDIUM**
- **Location & Evidence**: CLN-001 §3 Environment line 84: "신설: `internal/core/migration/steps/m01_template_version_sync.go`, `m03_hook_wrapper_drift.go`, `m04_skill_drift.go`"
- **Impact**: "신설" (new creation) of these files is explicitly reserved for MIG-002 per master §3.9 and per CLN-001 §2.2 Out of Scope line 75. Environment list contradicts.
- **Remediation**: Change "신설" to "referenced (owned by SPEC-V3-MIG-002)" or remove these file paths from CLN-001 §3 entirely.

#### D-NEW-007 — CLN-001 REQ text claims M01/M03/M04 ownership
- **Dimension**: D12, D2
- **Severity: MEDIUM** (folds into D-NEW-001 critical boundary but flagged separately as REQ-level surface)
- **Location & Evidence**: CLN-001 §5 REQ-CLN-001-001/003/004 language starts with "The M01 migration step **shall**…", "The M03 migration step **shall**…", "The M04 migration step **shall**…" — the subject of the EARS requirement is the migration step (owned by MIG-002 per master), not the CLN-001 tooling layer.
- **Impact**: Compounds D-NEW-001. The EARS subject is wrong: a CLN-001 REQ whose subject is "M01 migration step" asserts ownership of M01 migration step behavior.
- **Remediation**: Rewrite REQ text to use CLN-001's subject (`moai doctor template-drift` / cleanup orchestration / tooling layer) while referencing M01/M03/M04 as dependencies.

#### D-NEW-008 — Legacy "acceptance.md deferral" text retained after AC addition
- **Dimension**: D12 (contradiction), style
- **Severity: MEDIUM**
- **Location & Evidence**: Four SPECs kept the pre-remediation deferral sentence at the top of §6 Acceptance Criteria:
  - CLN-001 line 189: "상세 Given-When-Then 시나리오는 `acceptance.md` 참조 (본 SPEC의 Wave 4 scope에서는 spec.md만 생성)."
  - CLN-002 line 207: same phrase
  - MIGRATE-001 line 234: same phrase
  - OUT-001 line 194: same phrase
  After this sentence, the remediated ACs are now listed inline. `acceptance.md` still does not exist in any SPEC-V3-* directory.
- **Impact**: Reader confusion. The sentence implies ACs are deferred, but they are in fact inline below it. Tool parsing that detects "defers to acceptance.md" marker will flag these 4 SPECs as incomplete.
- **Remediation**: Delete the sentence from §6 of all four SPECs. Also remove the reference to `acceptance.md` from §10 Traceability sections (lines 249, 267, 309 / CLN-001 / CLN-002 / MIGRATE-001 / OUT-001:252).

#### D-NEW-009 — MIGRATE-001 §9.1 lists three sources for M01-M05
- **Dimension**: D4 (Dependency Graph), D12
- **Severity: MEDIUM**
- **Location & Evidence**: MIGRATE-001 §9.1 Blocked by (lines 287-289):
  ```
  - **SPEC-V3-CLN-001** (Template drift resolution): M01/M03/M04 구현체 제공.
  - **SPEC-V3-CLN-002** (Legacy code removal): M05 구현체 제공.
  - **SPEC-V3-MIG-002** (M02 agency archival + docs-site 4-locale sync in SPEC-V3-MIG-002 (absorbed)): M02 agency archival 및 locale sync 구현체 제공.
  ```
  Plus line 286: "SPEC-V3-MIG-002 (Initial migration set M01-M05): M01-M05 구현체가 존재해야 본 SPEC이 step 3(migration runner)을 실행할 수 있음."
- **Impact**: MIG-002 is listed as providing M01-M05 in one entry AND M02+locale-sync in another entry. CLN-001 is listed as providing M01/M03/M04 in another. This contradicts master §3.9 Ownership Split which says MIG-002 is the sole owner of M01-M05 Go files.
- **Remediation**: Collapse entries. Single authoritative Blocked-by entry for MIG-002 covering all M01-M05 implementation. CLN-001/002 entries describe tooling dependency, not Go file provision.

### LOW

#### D-NEW-010 — CMDS SPECs `related_theme` frontmatter not updated for Theme 10
- **Dimension**: D11 (Frontmatter Schema)
- **Severity: LOW**
- **Location & Evidence**:
  - Master §3.10 Theme 10 added (line 896 onwards) defining the CMDS theme
  - CMDS-001 line 22: `related_theme: "Theme 4: Agent Frontmatter Expansion (command parity)"` — should be Theme 10
  - CMDS-002 line 18: `related_theme: "Command argument substitution parity"` — free-form
  - CMDS-003 line 17: `related_theme: "Command inline shell execution parity"` — free-form
- **Impact**: Frontmatter doesn't cite the new master §3.10 theme; tooling that looks up `related_theme` → master §3.x fails to find Theme 10.
- **Remediation**: Set all 3 CMDS SPECs `related_theme: "Theme 10: Command Extension Parity"`.

#### D-NEW-011 — MIGRATE-001 §9.1 nested parenthesis
- **Dimension**: Style
- **Severity: LOW**
- **Location & Evidence**: MIGRATE-001 line 289: "**SPEC-V3-MIG-002** (M02 agency archival + docs-site 4-locale sync in SPEC-V3-MIG-002 (absorbed)): …". SPEC-V3-MIG-002 appears inside its own bracket expansion.
- **Remediation**: "**SPEC-V3-MIG-002** (M02 agency archival + docs-site 4-locale sync, absorbed): …"

---

## Dimension-by-Dimension Results (D1..D15)

Comparison to iteration 1 rubric bands:

| Dim | Title | Iter-1 Band | Iter-2 Band | Δ | Rationale |
|-----|-------|-------------|-------------|---|-----------|
| D1 | EARS Compliance | 0.75 | 0.90 | +0.15 | 4 formerly-ACless SPECs now carry EARS-structured ACs. All 10 new MIG-002 M02/LOC REQs are EARS-valid. |
| D2 | REQ ID Convention | 1.0 | 0.80 | −0.20 | MIG-002 M02-/LOC- compound subdomain prefix breaks the `REQ-<DOMAIN>-<NNN>-<NNN>` single-word convention used by all 27 others (D-NEW-012). |
| D3 | AC Testability | 0.50 | 0.80 | +0.30 | ACs present in all 28 SPECs; binary-testable; zero weasel words. Would be 1.0 but for D-NEW-005 (~50% REQ→AC mapping in 4 SPECs). |
| D4 | Dependency Graph | 0.50 | 0.75 | +0.25 | Asymmetries resolved (CLI-001 deps added, OUT-001 비차단 moved, MIG-001 §9.2 inflated moved). D-NEW-009 minor residual. No true cycle. |
| D5 | Traceability | 0.75 | 0.65 | −0.10 | Previous blocker (AC absence) resolved. New blockers: ~50% REQ→AC in 4 SPECs (D-NEW-005), 10 MIG-002 M02/LOC REQs missing from §10 Traceability (D-NEW-012), plus MIG-002 M02 path ambiguity (D-NEW-003). |
| D6 | Breaking Change Integrity | 0.50 | 0.90 | +0.40 | BC-007 now mapped to HOOKS-005 bc_id. bc_id array form uniform. All 8 BCs owned. |
| D7 | Phase Assignment | 0.50 | 0.90 | +0.40 | Phase format uniform. SKL-002 Phase 5 placement reconciled. HOOKS 6 SPECs phase-mapped correctly. |
| D8 | Gap Coverage | 0.75 | 0.75 | 0.00 | No change. Critical gaps all covered. CMDS theme frontmatter lag (D-NEW-010) is minor. |
| D9 | Scope Discipline | 0.75 | 0.50 | −0.25 | D-NEW-001, D-NEW-002 CLN-001/002 self-contradicts on migration step ownership. D-NEW-004 MIG-002 locale-sync In+Out. |
| D10 | Risk Coverage | 1.0 | 1.0 | 0.00 | Unchanged. |
| D11 | Frontmatter Schema | 0.50 | 0.90 | +0.40 | bc_id, priority, phase, dependencies, related_gap all normalized. D-NEW-010 CMDS theme and D-NEW-012 REQ-ID pattern left. |
| D12 | Internal Consistency | 0.50 | 0.50 | 0.00 | HOOKS numbering + §8.7 slot + CMDS theme hookup resolved (positive). BUT D-NEW-001/002/003/004/006/007/008/009 added (negative). Net unchanged. |
| D13 | Numbering & Counts | 0.75 | 0.85 | +0.10 | 545 REQ count verified. Auto-derived footer added. Phase distribution self-consistent. Minor: D-NEW-012 subdomain prefix. |
| D14 | Regression Check | (new) | 0.40 | — | 12 new defects introduced. 2 Critical, 4 High, 4 Medium, 2 Low. Remediation introduced bugs faster than resolving them in CRIT-005 / CRIT-002 scope. |
| D15 | Claimed-vs-Actual | (new) | 0.85 | — | 24/28 iteration-1 claims held exactly as stated. 2 partial (CRIT-005, MED-008). 2 observation-only. Zero outright FAIL on claims themselves, but remediation introduced side-effect defects. |

---

## Must-Pass Results

- **[PASS] MP-1 REQ number consistency**: 545 unique REQ IDs, zero duplicates. Individual SPECs are monotonic.
  - **Caveat**: MIG-002 uses compound subdomain prefix (M02-/LOC-) for 10 REQs. If MP-1 requires single-domain pattern, this is a FAIL; but the iteration-1 MP-1 criterion was "REQ numbers sequential per SPEC with no gaps/duplicates" — not specifically shape. Leaving as PASS with recorded concern.
- **[PASS] MP-2 EARS format compliance**: All REQs (including the 10 new M02/LOC in MIG-002) use EARS keyword prefixes. All 39 new ACs use Given/When/Then structure.
- **[PASS] MP-3 YAML frontmatter validity**: all fields present with correct types across all 28 SPECs.
- **[PASS] MP-4 Section 22 language neutrality**: no language-specific hardcoding introduced.

---

## Chain-of-Verification Pass

Second-look findings:

1. ✓ Re-checked every AC-ID in CLN-001, CLN-002, MIGRATE-001, OUT-001 — confirmed 10/9/12/8 = 39 matches claim exactly.
2. ✓ Re-checked REQ→AC mapping coverage via grep for `maps REQ-...` occurrences — confirmed ~50% coverage regression (D-NEW-005).
3. ✓ Re-checked bc_id types across all 28 SPECs line-by-line — confirmed 100% array form.
4. ✓ Re-verified CLN-003 purge via cross-directory grep — confirmed zero residual references.
5. ✓ Re-checked CLN-001 scope boundary by reading §2 through §10 carefully — found four separate surfaces still contradicting the new §2.2 boundary (D-NEW-001).
6. ✓ Same for CLN-002 (D-NEW-002).
7. ✓ Re-checked MIG-002 scope coherence end-to-end — found M02 path inconsistency (D-NEW-003) and locale-sync In/Out contradiction (D-NEW-004).
8. ✓ Re-checked REQ pattern across all 28 SPECs — found MIG-002 M02/LOC subdomain pattern breaks convention (D-NEW-012).
9. ✓ Re-checked master §8.7 and §3.9 reconciliation — both now consistent (D-HIGH-004 resolved).
10. ✓ Re-checked CLI-001 and OUT-001 dependency sections — both correctly restructured.
11. First-pass missed D-NEW-008 (legacy acceptance.md deferral text); caught on second reading of §6 headers.
12. First-pass thought D-MED-008 was fully resolved because master §3.10 exists; second pass found CMDS SPECs frontmatter unchanged (D-NEW-010).

Chain-of-verification surfaced 2 additional defects (D-NEW-008, D-NEW-010) beyond the first pass.

---

## Regression Check (Iteration 2)

Defects from iteration 1 status summary:

| ID | Severity | Iter-2 Status |
|----|----------|---------------|
| D-CRIT-001 | Critical | **RESOLVED** |
| D-CRIT-002 | Critical | **RESOLVED (count) / PARTIAL (traceability)** — see D-NEW-005 |
| D-CRIT-003 | Critical | **RESOLVED** |
| D-CRIT-004 | Critical | **RESOLVED** |
| D-CRIT-005 | Critical | **UNRESOLVED** — master-level fix added but SPEC-body contradictions remain (D-NEW-001, D-NEW-002) |
| D-HIGH-001..012 | High (9 real) | **RESOLVED** (all real HIGHs; HIGH-007/009 were false positives) |
| D-MED-001..008 | Medium | **RESOLVED** (except D-MED-008 PARTIAL — D-NEW-010; D-MED-006 not verified) |
| D-LOW-001..004 | Low | **RESOLVED** (except D-LOW-002 not actionable) |

**Stagnation check**: D-CRIT-005 is the only iteration-1 Critical defect to survive iteration 2 in partial form. The fix is close — master design is correct, SPEC bodies are inconsistent. Not a stagnation pattern; it's a completeness pattern. Predict iteration 3 can close it in one editing pass.

---

## Verdict & Recommendation

**Verdict: NOT-READY**

### Rationale
- 5 iteration-1 Critical defects: 4 fully resolved, 1 partial (D-CRIT-005)
- 12 new defects introduced by remediation (2 Critical, 4 High, 4 Medium, 2 Low)
- The READY criterion requires ALL Critical resolved AND new defects ≤ 2 Medium + 0 Critical/High. We are 2 Critical and 4 High away.
- The residual work is bounded. One more iteration should suffice.

### Recommended iteration 3 scope (ordered by priority)

**Phase 1: Scope boundary closure (resolves D-CRIT-005, D-NEW-001, D-NEW-002, D-NEW-006, D-NEW-007)**

1. **CLN-001 §2.1 In Scope**: delete lines 59-61 (MigrationStep 구현체 M01/M03/M04 bullets). Replace with a single line: "Tooling layer: `moai doctor template-drift` CLI + orchestration of M01/M03/M04 (migration step Go files owned by SPEC-V3-MIG-002)."
2. **CLN-001 §3 Environment**: change line 84 "신설" to "참조 (owned by SPEC-V3-MIG-002)".
3. **CLN-001 §5 REQs 001/003/004**: rewrite subject from "M01/M03/M04 migration step" to CLN-001's tooling subject. Example: "The `moai doctor template-drift` subcommand **shall** invoke SPEC-V3-MIG-002's M01 migration step and report its status."
4. **CLN-001 §10 Traceability paths (lines 253-256)**: replace migration step file paths with `internal/cli/doctor_template_drift.go`, `internal/cli/doctor_skill_drift.go`.
5. **CLN-002 §2.1**: delete line 68 "M05 (Migration Step — ...)" block. Replace with single line about `moai doctor legacy-cleanup` CLI + direct source edits.
6. **CLN-002 §3 Environment**: delete `internal/core/migration/steps/m05_*.go` from 신설 list. Keep only `embed.go`, `deps.go`, `.gitignore` direct edits.
7. **CLN-002 §10 Traceability**: remove migration step file paths (lines 271-272).

**Phase 2: MIG-002 absorption closure (resolves D-NEW-003, D-NEW-004, D-NEW-012)**

8. **MIG-002 M02 archive path**: pick `docs/archive/agency-v2/` (matches master §3.9 absorption language). Update §2.1 line 97, §5.7 REQ-MIG-002-M02-001 wording, §6 AC-MIG-002-02 path, and §6 AC-MIG-002-M02-01 path to use this one path consistently. Remove `~/.moai/history/v2.12/` references (or document it as backup-only snapshot).
9. **MIG-002 locale sync**: decide In or Out. If In (recommended, matches master §3.9 absorption): delete §2.2 line 134 "docs-site locale sync — manager-docs 위임, not migration". If Out: delete §5.8 REQ-MIG-002-LOC-001..004 and §6 AC-MIG-002-LOC-01/02.
10. **MIG-002 REQ IDs**: convert REQ-MIG-002-M02-001..006 to REQ-MIG-002-052..057. Convert REQ-MIG-002-LOC-001..004 to REQ-MIG-002-058..061. Update §6 AC `(maps REQ-...)` back-references accordingly. Add 10 rows to §10 Traceability table covering these new REQs.

**Phase 3: Traceability coverage (resolves D-NEW-005)**

11. **CLN-001 §6 ACs**: add ACs for 13 uncovered REQs (008, 009, 010, 011, 013, 014, 017, 018, 019, 020, 021, 022, 023). Target: 23/23 REQ→AC coverage.
12. **CLN-002 §6 ACs**: same for 10 uncovered REQs.
13. **MIGRATE-001 §6 ACs**: same for 12 uncovered REQs.
14. **OUT-001 §6 ACs**: same for 10 uncovered REQs.

**Phase 4: Residual cleanup (resolves D-NEW-008, D-NEW-009, D-NEW-010, D-NEW-011)**

15. Delete the "상세 Given-When-Then 시나리오는 `acceptance.md` 참조" sentence from CLN-001 line 189, CLN-002 line 207, MIGRATE-001 line 234, OUT-001 line 194. Delete same reference from §10 Traceability lines.
16. MIGRATE-001 §9.1: collapse the three M01-M05 provider entries into single "**SPEC-V3-MIG-002** — owns M01-M05 migration step Go implementations; this SPEC invokes MIG-002's runner." Remove CLN-001 and CLN-002 from Blocked-by entirely (or move to Related with tooling-only dependency note).
17. CMDS-001/002/003 frontmatter: update `related_theme` to `"Theme 10: Command Extension Parity"`.
18. MIGRATE-001 §9.1 line 289: rewrite awkward nested parenthesis.

**Phase 5: Low-priority (optional)**

19. D-MED-006 (gap severity labeling in master §8.2): verify against gap-matrix.md; update if actually drifted.

### Estimated iteration 3 churn
- CLN-001: ~6-8 edits (lines in §2, §3, §5, §6, §10)
- CLN-002: ~4-6 edits (lines in §2, §3, §6, §10)
- MIG-002: ~10-15 edits (M02 path, locale sync decision, 10 REQ renumbers, §10 table expansion)
- MIGRATE-001: ~4 edits (§6 acceptance.md sentence, §9.1 collapse, line 289 wording)
- OUT-001: ~2 edits (§6 sentence + §10 reference)
- CMDS-001/002/003: 3 frontmatter edits

Total estimated: ~30-40 line-level edits across 7 files. Plan-auditor can re-audit iteration 3 in a single pass.

### Path forward

Proceed with iteration 3 remediation per the 18-item list above. On completion, re-invoke plan-auditor with this report as prior-iteration input. If iteration 3 does not fully close D-CRIT-005, D-NEW-001, D-NEW-002 (the three most important), escalate to user intervention.

---

End of audit report v2.
