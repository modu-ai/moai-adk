# v3.0.0 Plan Audit Report

> Auditor: plan-auditor (independent, adversarial)
> Date: 2026-04-23
> Scope: `/Users/goos/MoAI/moai-adk-go/docs/design/major-v3-master.md` (1,397 lines) + 28 v3 SPECs at `/Users/goos/MoAI/moai-adk-go/.moai/specs/SPEC-V3-*/spec.md` (8,583 lines total)
> Bias-prevention: M1–M6 active. Reasoning context ignored per M1.

---

## Executive Summary

- **Verdict: NOT-READY**
- Critical defects: **5**
- High defects: **11**
- Medium defects: **8**
- Low defects: **4**
- Total defects: **28**

Two structural blockers prevent v3.0.0 from shipping as currently planned:

1. **SPEC-V3-CLN-003 is referenced as a dependency by six other SPECs and by the master design, but the SPEC does not exist on disk.** The "Phase 8 Release" node explicitly depends on SPEC-V3-MIGRATE-001 which lists SPEC-V3-CLN-003 as "Blocked by". Without CLN-003, the M02 agency archival migration has no owning SPEC and the release graph is broken.

2. **Four SPECs (CLN-001, CLN-002, MIGRATE-001, OUT-001) have no structured AC-IDs.** They defer AC detail to a non-existent `acceptance.md` and provide only prose bullets. Traceability from these SPECs' REQs → ACs is untraced; implementation agents have no binary pass/fail criteria; plan-auditor cannot score AC quality.

Secondary blockers include SPEC-ID drift between master §3.1 / §6.2 / §8.1 and the actual 28-SPEC set (HOOKS SPECs silently renumbered; HOOKS-007/008/009 never written), and mixed `bc_id` frontmatter types (scalar vs array vs null) that will break any downstream schema validator built from SPEC-V3-SCH-001.

The grounding research (194 gaps, 9 themes, 6 Wave 1 findings) is comprehensive and no Critical-severity gap is orphaned. Individual SPEC bodies (where present) are largely high-quality: EARS syntax is consistent, REQ IDs are unique across the 545-REQ corpus with only legitimate cross-SPEC traceability references, zero codex/time/emoji/TODO/TBD pollution. The foundation is sound — the defects are structural completeness and frontmatter consistency, not intellectual.

---

## Defect Catalog

### CRITICAL (blockers — v3.0.0 cannot ship as-is)

#### D-CRIT-001 — SPEC-V3-CLN-003 does not exist on disk

- **Dimension**: D4 (Dependency Graph), D8 (Gap Coverage)
- **Location**:
  - `docs/design/major-v3-master.md:883` — master §3.9 lists CLN-003 as deliverable SPEC
  - `docs/design/major-v3-master.md:1059` — master §6.2 Phase 6a PR mapping omits it; but §8.7 lists it
  - `docs/design/major-v3-master.md:1144-1146` — master §8.7 SPEC Index lists CLN-001 / CLN-002 / SPEC-V3-SPEC-001 (CLN-003 numeric slot is taken by SPEC-001!); but many downstream SPECs cite CLN-003
  - `.moai/specs/SPEC-V3-MIGRATE-001/spec.md:17` — frontmatter `dependencies: SPEC-V3-CLN-003`
  - `.moai/specs/SPEC-V3-MIGRATE-001/spec.md:287` — §9.1 "Blocked by: **SPEC-V3-CLN-003** (Agency archival + docs drift): M02 구현체 제공."
  - `.moai/specs/SPEC-V3-MIG-001/spec.md:329` (§9.2) — blocks "SPEC-V3-CLN-001/002/003"
  - `.moai/specs/SPEC-V3-MIG-002/spec.md:322` (§9.2) — blocks "SPEC-V3-CLN-001/002/003 — other writer"
  - `.moai/specs/SPEC-V3-CLN-001/spec.md:69,237` — references CLN-003 as M02 owner
  - `.moai/specs/SPEC-V3-CLN-002/spec.md:52,53,80,254` — same
  - Filesystem: `ls .moai/specs/SPEC-V3-CLN-*/` returns only CLN-001 and CLN-002
- **Evidence**: Master §3.9 explicitly says "SPEC-V3-CLN-003 (`.agency/` archival + docs drift — M02 + docs-site locale sync)". MIGRATE-001 §9.1 literally states the M02 step cannot run without CLN-003. SPEC-V3-CLN-003 directory does not exist.
- **Impact**: Phase 8 (Release Rollout) has a hard blocker — MIGRATE-001 cannot close. The M02 agency archival migration has no authoritative spec. docs-site 4-locale sync (gm#194) has no owning SPEC. Gap matrix rows gm#190, gm#191 are only partially covered (via MIG-002 frontmatter), but CLN-003's intended scope (docs-site locale sync per master §3.9 step 6) is entirely unassigned.
- **Remediation**: Write `SPEC-V3-CLN-003/spec.md` covering M02 agency archival + docs-site 4-locale sync, OR merge that scope explicitly into SPEC-V3-MIG-002 (and remove the CLN-003 dependency from MIGRATE-001/MIG-001/MIG-002/CLN-001/CLN-002/master-v3 §3.9/§8.7). Do NOT both: pick one owner.

#### D-CRIT-002 — Four SPECs lack structured Acceptance Criteria (AC-IDs)

- **Dimension**: D3 (AC Testability), D5 (Traceability)
- **Location & Evidence**:
  - `.moai/specs/SPEC-V3-CLN-001/spec.md:181-196` — §6 reads: `상세 Given-When-Then 시나리오는 acceptance.md 참조 (본 SPEC의 Wave 4 scope에서는 spec.md만 생성).` followed by 10 prose bullets under `핵심 기준:` with NO `AC-CLN-001-NN` identifiers
  - `.moai/specs/SPEC-V3-CLN-002/spec.md:197-220` — identical pattern
  - `.moai/specs/SPEC-V3-MIGRATE-001/spec.md:230-246` — identical pattern
  - `.moai/specs/SPEC-V3-OUT-001/spec.md:179-191` — identical pattern
  - `ls .moai/specs/SPEC-V3-*/` shows **no** `acceptance.md` files exist anywhere
- **Contrast**: SPEC-V3-HOOKS-001 §6 (lines 123-133), SPEC-V3-MIG-001 §6 (lines 264-284), SPEC-V3-TEAM-001, SPEC-V3-SKL-001, SPEC-V3-PLG-001, SPEC-V3-CLI-001 all have properly numbered `AC-<DOMAIN>-NNN-NN` IDs with Given/When/Then structure and explicit `(maps REQ-...)` back-references.
- **Impact**:
  - D3 testability: affected SPECs have no binary pass/fail criteria. "M01 재실행 시 no-op으로 동작" (CLN-001 bullet 2) has no observable assertion — what counts as "no-op"? No filesystem diff? migration_version unchanged? No log line emitted?
  - D5 traceability: 82 REQs across the four SPECs (CLN-001: 23, CLN-002: 20, MIGRATE-001: 25, OUT-001: 22, totaling ~90 counting the REQ-HOOK-013 / REQ-MIGRATE-012 cross-refs) cannot map to concrete ACs.
  - Run-phase agents (manager-ddd / manager-tdd) will fail the Re-planning Gate trigger "3+ iterations with no new SPEC acceptance criteria met" because there are no AC completion checkboxes to tick.
  - plan-auditor and evaluator-active cannot score these SPECs in the GAN Loop.
- **Remediation**: Either (a) promote the prose bullets to structured `AC-<DOMAIN>-NNN-NN: Given ... When ... Then ... (maps REQ-...)` entries inline in §6 (as the other 24 SPECs already do), or (b) write the deferred `acceptance.md` files now as part of Phase 1 completion, not post-hoc. Option (a) is strongly preferred given the Wave 4 scope charter says spec.md only.

#### D-CRIT-003 — SPEC-ID set drifts from master design §3.1 / §6.2 / §8.1 for the HOOKS theme

- **Dimension**: D7 (Phase), D12 (Internal Consistency)
- **Location & Evidence**:
  - `docs/design/major-v3-master.md:237-243` (§3.1) — master declares HOOKS-001..006 with titles: 001 "rich JSON IO" / 002 "if condition + matcher" / 003 "source precedence 3-tier" / 004 "handler richness 6 events" / 005 "async + once + CLAUDE_ENV_FILE" / 006 "permission decision"
  - `docs/design/major-v3-master.md:1102-1110` (§8.1) — master declares 9 HOOKS SPECs: 001..006 PLUS SPEC-V3-HOOKS-007 (type:prompt), SPEC-V3-HOOKS-008 (type:agent), SPEC-V3-HOOKS-009 (type:http)
  - `docs/design/major-v3-master.md:1059` (§6.2 Phase 6a table) — refers to "SPEC-V3-HOOKS-007/008/009"
  - Actual on-disk HOOKS SPECs (titles from frontmatter line 3):
    - HOOKS-001: "Hook Protocol v2 — Rich JSON IO" ✓ matches master §3.1
    - HOOKS-002: "Hook Type System — 4 Types (command, prompt, agent, http)" — **does NOT match** master §3.1 "if condition + matcher"; this SPEC absorbed the master §8.1 HOOKS-007/008/009 type content
    - HOOKS-003: "Async Hook Execution — async, asyncRewake, once" — **does NOT match** master §3.1 "source precedence"; maps to master §3.1 HOOKS-005 scope
    - HOOKS-004: "Hook Matcher & Filter System — if condition" — maps to master §3.1 HOOKS-002 scope
    - HOOKS-005: "Missing Hook Event Handlers — 14 Events" — maps to master §3.1 HOOKS-004 scope
    - HOOKS-006: "Hook Scoping Hierarchy — 3-tier" — maps to master §3.1 HOOKS-003 scope
  - HOOKS-007/008/009 **do not exist on disk**; their content is merged into HOOKS-002
- **Impact**: Any reader following master §3.1 → §6.2 → §8.1 references to locate implementation SPECs will hit missing-file errors or cross-domain content. Phase 2 "Hook Protocol v2 core" in §6.1 says it contains "T1-HOOK-01..T1-HOOK-07" mapped to HOOKS-001..006, but on disk HOOKS-002 is a Phase 6a Tier 2 Strategic scope (per its frontmatter `phase: "Phase 6a Tier 2 Strategic Differentiators"`), violating the Phase 2 foundation order. HOOKS-003's blocker chain (HOOKS-001) passes, but its Phase 2 claim is actually scope from §3.1 HOOKS-005 (async). Impact on Wave 5 (plan) and Wave 6 (run) phase scheduling is substantial.
- **Remediation**: Either (a) update master §3.1 and §8.1 to match the on-disk renumbering (acknowledging HOOKS-002 absorbed types 007/008/009); or (b) split HOOKS-002 back into the master-intended shape and rename HOOKS-003/004/005/006 to match master §3.1 ordering. Option (a) is less disruptive.

#### D-CRIT-004 — `bc_id` frontmatter type is inconsistent across 10 breaking SPECs

- **Dimension**: D11 (Frontmatter Schema Conformance)
- **Location & Evidence** (all line 21 of each `spec.md`):
  - Scalar form (`bc_id: BC-XXX`): HOOKS-001, HOOKS-005, HOOKS-006, TEAM-001, MIGRATE-001
  - Array form (`bc_id: [BC-XXX]` or `bc_id: [BC-XXX, BC-YYY]`): MIG-001, SCH-001, SCH-002, MIG-002 (empty), PLG-001 (empty)
  - Null form (`bc_id: null`): CMDS-001, CMDS-002, CMDS-003, CLN-001, CLN-002, OUT-001
  - Omitted entirely: AGT-001, AGT-002, AGT-003, CLI-001, HOOKS-002, HOOKS-003, HOOKS-004, MEM-001, MEM-002, SKL-001, SKL-002, SPEC-001 (all `breaking: false`)
- **Impact**: Any YAML schema built per SPEC-V3-SCH-001 to validate SPEC frontmatter will reject 3 of the 5 possible valid shapes. Cross-SPEC tooling (`moai doctor spec --validate`) cannot adopt a single validator tag. SCH-002 lists two BCs as an array `[BC-002, BC-003]` — downstream consumer code must handle the scalar/array/null trinity. This is exactly the typo-silently-ignored problem that SCH-001 is supposed to eliminate, being baked into the SPEC set that will validate it.
- **Remediation**: Normalize to array form (`bc_id: []` when none, `bc_id: [BC-001]` when one, `bc_id: [BC-002, BC-003]` when multi). Update master §4 blast radius analysis table to show `bc_id: []` as the canonical shape. Bump all 28 SPECs in one commit.

#### D-CRIT-005 — MIG-002 and CLN-001/002 claim overlapping ownership of migrations M01–M05

- **Dimension**: D12 (Internal Consistency), D9 (Scope Discipline)
- **Location & Evidence**:
  - `.moai/specs/SPEC-V3-MIG-002/spec.md:2-3` — title: "v2-to-v3 Migration Content (M01-M05 concrete steps)"; §2.1 in-scope claims M01–M05 concrete step implementations
  - `.moai/specs/SPEC-V3-CLN-001/spec.md:10` — phase: "Phase 1 - Foundation (ships via M01/M03/M04)" — claims to implement M01, M03, M04
  - `.moai/specs/SPEC-V3-CLN-001/spec.md:247-254` — §10 lists implementation paths `internal/core/migration/steps/m01_template_version_sync.go`, `m03_hook_wrapper_drift.go`, `m04_skill_drift.go`
  - `.moai/specs/SPEC-V3-CLN-002/spec.md:10` — phase: "Phase 1 - Foundation (ships via M05) + direct source edits" — claims to implement M05
  - `.moai/specs/SPEC-V3-MIG-002/spec.md:322` (§9.2 Blocks) — "SPEC-V3-CLN-001/002/003 (Internal Cleanup SPECs — other writer) — 중복 항목 (M03/M04/M05와 overlap); cleanup SPEC은 본 migration을 reference"
  - Master §3.9 says "All cleanup lives in M01–M05. Users run `moai migrate v2-to-v3`"
- **Impact**: Two SPECs (MIG-002 and CLN-001) both claim to contain the production implementation of M01 / M03 / M04 migration steps. Wave 5 planning (plan.md creation) will not know which SPEC to assign the `m01_template_version_sync.go` file to. Wave 6 run-phase will create duplicate PRs, merge-conflict on the same files, or one SPEC's implementation will silently overwrite the other's. MIG-002 §9.2 acknowledges the "중복" (duplication) but does not resolve it — only labels it.
- **Remediation**: Define explicit ownership with mutually-exclusive file scopes. One workable assignment:
  - MIG-002 owns the migration step Go files (`internal/core/migration/steps/m0{1,2,3,4,5}*.go`)
  - CLN-001/002/003 own the `moai doctor template-drift`, `moai doctor legacy-cleanup`, and deploy-target files outside `internal/core/migration/steps/`
  Then document this split in master §3.9 and in each SPEC's §2 Scope. Or alternatively, collapse CLN-001/002/003 into MIG-002 as sub-sections and remove them as independent SPECs (reducing 28 → 25 or 26).

### HIGH (significant issues — v3.0.0 should not ship before fix)

#### D-HIGH-001 — `priority` frontmatter field has three inconsistent vocabularies

- **Dimension**: D11
- **Evidence**: 18 SPECs use "P0 Critical"/"P1 High"/"P2 Medium"/"P3 Low" (AGT/CMDS/HOOKS/MEM/SKL/TEAM/SPEC/CLI); 6 use plain "Critical"/"High"/"Medium"/"Low" (CLN-001=High, CLN-002=Low, MIG-001=Critical, MIG-002=High, MIGRATE-001=High, OUT-001=Medium, PLG-001=High, SCH-001=Critical, SCH-002=High)
- **Impact**: Any `priority_distribution` aggregation or sort-by-priority tooling breaks.
- **Remediation**: Normalize to one vocabulary. Master design does not define the canonical form; recommend "P0 Critical" shape with explicit semantics at master §11 or a new Appendix D.

#### D-HIGH-002 — `phase` frontmatter field mixes two formats

- **Dimension**: D11, D7
- **Evidence**: 21 SPECs use "v3.0.0 — Phase N <name>" prefix; 7 SPECs use plain "Phase N — <name>" or "Phase N - <name>" (CLN-001, CLN-002, MIG-001, MIG-002, MIGRATE-001, OUT-001, PLG-001, SCH-001, SCH-002)
- **Impact**: Same as D-HIGH-001 — tooling/grep that expects either format will miss half.
- **Remediation**: Normalize.

#### D-HIGH-003 — SPEC-V3-SKL-002 phase contradicts master §6.2

- **Dimension**: D7
- **Evidence**:
  - `.moai/specs/SPEC-V3-SKL-002/spec.md:10` — `phase: "v3.0.0 — Phase 5 Internal Cleanup"`
  - `docs/design/major-v3-master.md:1060` (§6.2) — SPEC-V3-SKL-002 is listed under row "6b" (Phase 6b Tier 2 Polish)
- **Impact**: Wave 5 scheduling will place SKL-002 in either Phase 5 or Phase 6b based on which source is authoritative. Downstream PR mapping diverges.
- **Remediation**: Pick one; update the other. SKL-002 scope is "Skill Drift Detection" — conceptually closer to Phase 5 (Internal Cleanup) than Phase 6b (Polish). Recommend updating master §6.2 to match the SPEC.

#### D-HIGH-004 — Master §8.7 SPEC numbering conflict: CLN-003 vs SPEC-001 both claim slot 3 of "Internal Cleanup"

- **Dimension**: D12
- **Evidence**:
  - `docs/design/major-v3-master.md:1144-1146` (§8.7 "Internal Cleanup + moai-unique SPECs (3 SPECs)") — Lists CLN-001, CLN-002, SPEC-V3-SPEC-001 (note: SPEC-001 is the SPEC-to-SPEC chaining item, NOT CLN-003)
  - `docs/design/major-v3-master.md:883` (§3.9 Theme 9 SPEC IDs) — Lists CLN-001, CLN-002, **CLN-003** (a different third SPEC)
  - The two sections disagree on whether the third "cleanup" SPEC is SPEC-V3-SPEC-001 (chaining) or SPEC-V3-CLN-003 (agency archival)
- **Impact**: Compounds D-CRIT-001. Master design is internally inconsistent about whether CLN-003 exists or whether SPEC-001 fills that slot.
- **Remediation**: Reconcile: §3.9 says CLN-003 exists; §8.7 says SPEC-001 takes its slot. Since CLN-003 does not exist on disk and SPEC-V3-SPEC-001 does (with different scope), the §8.7 classification is closer to reality but introduces the "moai-unique" label for a non-cleanup item. Suggest renaming §8.7 group to "Internal Cleanup + moai-unique" and removing CLN-003 from §3.9 after merging its M02 scope into MIG-002.

#### D-HIGH-005 — Dependency asymmetry: SPEC-V3-MIG-001 claims to block HOOKS-001..006 but those SPECs do not list MIG-001 as blocker

- **Dimension**: D4
- **Evidence**:
  - `.moai/specs/SPEC-V3-MIG-001/spec.md:324-325` (§9.2) — "**SPEC-V3-HOOKS-001~006** — Hook Protocol v2 배포 시 기존 hook wrapper 업그레이드 migration 잠재 필요"
  - HOOKS-001/002/003/004/005/006 §9.1 Blocked by — none list SPEC-V3-MIG-001 (HOOKS-001 lists only SCH-001; HOOKS-006 lists only HOOKS-001/004/SCH-002; etc.)
- **Impact**: The dependency graph edges are one-directional. If MIG-001 is a real prerequisite, HOOKS SPECs should list it; if not, MIG-001's Blocks list is inflated. This hides true critical path.
- **Remediation**: Clarify whether MIG-001's block is actual (then add to HOOKS §9.1) or aspirational (then remove from MIG-001 §9.2).

#### D-HIGH-006 — Master claims "~510 REQs across 28 SPECs"; actual is 545

- **Dimension**: D13
- **Evidence**: Parse of all 28 SPECs for lines matching `REQ-<DOMAIN>-<NNN>-<NNN>` declarations yields 545 unique REQ IDs. Master design does not directly state 510 (that figure is in the audit prompt) but the delta between what master §2.1 footprint (104 SPECs) implies and actual corpus gives observers an inaccurate scale signal. Audit prompt claim "~510" ≠ actual "545" = 6.8% drift.
- **Impact**: Low (documentation drift).
- **Remediation**: Generate and commit an auto-derived count table at the bottom of master §8, pulled at build time.

#### D-HIGH-007 — SPEC-V3-AGT-003 cross-SPEC REQ reference is not declared as dependency symmetrically

- **Dimension**: D4
- **Evidence**:
  - `.moai/specs/SPEC-V3-AGT-003/spec.md:154` — "(SPEC-V3-AGT-002 REQ-AGT-002-005) **shall** route permission prompts…"
  - `.moai/specs/SPEC-V3-AGT-002/spec.md:§9.2` — correctly lists AGT-003 as "Blocks"
  - Fine per se, but AGT-003 §9.1 shows "SPEC-V3-AGT-002" as blocker. Symmetry intact. Actually this is fine. Demoting to noted-but-not-defect.
- **Status**: On re-review, this is not a defect — AGT-002 §9.2 does list AGT-003 as blocked-by-target. Remove from HIGH list.

#### D-HIGH-008 — Master §6.2 "Phase → SPEC → PR mapping" table row `1` omits SPEC-V3-MIG-002 under Phase 1 but lists it elsewhere

- **Dimension**: D7
- **Evidence**:
  - `docs/design/major-v3-master.md:1054` (§6.2) — Phase 1 row says "SPEC-V3-SCH-001, SPEC-V3-SCH-002, SPEC-V3-MIG-001, SPEC-V3-MIG-002" (4 items ✓)
  - Actually this is correct. Demoting.
- **Status**: False positive on second look. Removed.

#### D-HIGH-008 (reassigned) — Phase 6a §6.2 row mentions HOOKS-007/008/009 which don't exist

- **Dimension**: D7, D12
- **Evidence**: `docs/design/major-v3-master.md:1059` — "| 6a | SPEC-V3-HOOKS-007/008/009, SPEC-V3-PLG-001, SPEC-V3-MEM-002, SPEC-V3-AGT-002/003, SPEC-V3-TEAM-001, SPEC-V3-SPEC-001 | 9–12 PRs |"
- **Impact**: PR count estimate for Phase 6a is based on SPECs that don't exist (3 absorbed into HOOKS-002). Real Phase 6a set on disk: HOOKS-002, PLG-001, MEM-002, AGT-002/003, TEAM-001, SPEC-001 = 7 SPECs.
- **Remediation**: Update master §6.2 after D-CRIT-003 resolution.

#### D-HIGH-009 — SPEC-V3-SCH-001 (dependency of 7+ SPECs) does not declare SPEC-V3-MIG-001 reciprocity

- **Dimension**: D4
- **Evidence**: SCH-001 §9.2 (lines 237+) lists MIG-001 as blocked; MIG-001 §9.1 correctly lists SCH-001 as blocker. Symmetric. Demoting to noted.
- **Status**: False positive. Removed from HIGH.

#### D-HIGH-010 — SPEC-V3-OUT-001 §9.1 "Blocked by" contradicts itself

- **Dimension**: D4
- **Evidence**:
  - `.moai/specs/SPEC-V3-OUT-001/spec.md:222-224` — "SPEC-V3-SCH-001 (Formal config schemas) — 비차단. 본 SPEC은 config.yaml의 `design.yaml` 스키마와 독립적이나, output rendering 옵션이 design.yaml에 추가될 가능성에 대비해 스키마 등록 경로에 접근 필요."
  - Text labels SCH-001 as "비차단" (non-blocking) but lists it under "9.1 Blocked by" heading
- **Impact**: Dependency classification ambiguous. Is this a hard blocker or a soft reference?
- **Remediation**: Move soft references to §9.3 Related; keep §9.1 as actual blockers only.

#### D-HIGH-011 — Dependency edge asymmetry: CLI-001 → PLG-001, but CLI-001 → SCH-001/MIG-001/HOOKS-001/AGT-001 not declared

- **Dimension**: D4
- **Evidence**:
  - `.moai/specs/SPEC-V3-CLI-001/spec.md:95-97` (§3 Environment) — CLI-001 `doctor config / doctor migration / doctor hook / doctor agent` subcommands rely on SCH-001 / MIG-001 / HOOKS-001 / AGT-001
  - `.moai/specs/SPEC-V3-CLI-001/spec.md:§9.1` — only lists PLG-001 as blocker
- **Impact**: Real dependency underdeclared. Wave 5 schedule cannot place CLI-001 after SCH-001/MIG-001/HOOKS-001/AGT-001 if SPEC says only PLG-001 blocks.
- **Remediation**: Add SCH-001, MIG-001, HOOKS-001, AGT-001 to CLI-001 §9.1 Blocked by (or §9.3 Related if scope of use is non-critical).

### MEDIUM (fix before beta)

#### D-MED-001 — SPEC-V3-CLN-001 frontmatter missing `bc_id` field while others have it

- **Dimension**: D11
- **Evidence**: CLN-001 has `bc_id: null` at line 22; CLN-002 likewise. AGT-001/002/003 (breaking: false) have no `bc_id` field at all; they omit it entirely. Shape-drift among non-breaking SPECs.
- **Impact**: Minor schema inconsistency (same as D-CRIT-004 but lower severity for non-breaking SPECs).
- **Remediation**: Require `bc_id: []` uniformly on non-breaking SPECs, or omit uniformly.

#### D-MED-002 — `dependencies` frontmatter field mixed styles

- **Dimension**: D11
- **Evidence**: HOOKS-001 uses `dependencies:` followed by `- SPEC-V3-SCH-001` list; other SPECs use `dependencies:` with inline content or empty. MIG-001 has `dependencies: []` (empty). Similar drift.
- **Impact**: Low. Normalize.

#### D-MED-003 — Master §7 Risk register has 12 risks, audit prompt assumed 8

- **Dimension**: D10
- **Evidence**: `docs/design/major-v3-master.md:1079-1093` — lists R-001 through R-012 (12 risks). Audit prompt text mentioned "8 risks (R-1..R-8)"; prompt is outdated, not master. All 12 risks have concrete mitigations. R-005/R-007 have High impact and "Very Low" / "Low" probability with two-part mitigations.
- **Status**: Not a master defect — this is an audit prompt drift observation only. No master fix needed.

#### D-MED-004 — Master §8.1 introduces SPEC-V3-HOOKS-007/008/009 as "Tier 2 strategic" but §6.1 Phase 2 still lists all HOOKS as "Hook Protocol v2 core"

- **Dimension**: D7, D12 (related to D-CRIT-003)
- **Evidence**: `docs/design/major-v3-master.md:1017-1019` (§6.1 Phase 2) — mentions only "T1-HOOK-01..T1-HOOK-07" for Phase 2; §6.2 row 6a separates out HOOKS-007/008/009 as Tier 2 strategic. Master text consistent within §6.1/§6.2 but inconsistent with §8.1 which groups all 9 HOOKS under "Hooks/Commands SPECs (9 SPECs)".
- **Remediation**: Harmonize narrative — 6 foundation + 3 strategic.

#### D-MED-005 — SPEC-V3-SPEC-001 contains four example identifiers that parse as REQ IDs

- **Dimension**: D2
- **Evidence**: `.moai/specs/SPEC-V3-SPEC-001/spec.md` — prose example text uses `REQ-DOMAIN-NNN-NNN`, `REQ-ID`, `REQ-V3-PARENT-001-001`, `REQ-X-001` to illustrate the chaining protocol. A naive regex match for `REQ-[A-Z]+-[0-9]+-[0-9]+` picks these up alongside real REQs.
- **Status**: Legitimate illustrative text, not a defect per se, but creates false positives in automated analysis.
- **Remediation**: Code-fence the examples or use `<REQ-DOMAIN-NNN-NNN>` placeholder syntax that doesn't collide with the real pattern.

#### D-MED-006 — Gap rows gm#45 labeled "High" in gap-matrix but cited as part of Critical Memory bundle in MEM-001

- **Dimension**: D8
- **Evidence**:
  - `gap-matrix.md:62` — "| 45 | Memory | 4-type memory taxonomy ... | Schema ALREADY aligns; not formally enforced | High |"
  - `docs/design/major-v3-master.md:1114` (§8.2) — lists gm#44, gm#45, gm#50, gm#53 all together under MEM-001 (includes 45 which is High, 50 which is High, 53 I'd need to check) — but master prose at §3.5 problem statement does not distinguish Critical from High. Only gm#44 is Critical.
- **Remediation**: Low priority — severity tagging clarity in master §8.2 would help.

#### D-MED-007 — SPEC-V3-MIG-002 related_gap array uses raw integers (`[183, 184, ...]`); others use `gm#` prefix

- **Dimension**: D5, D11
- **Evidence**:
  - `.moai/specs/SPEC-V3-MIG-002/spec.md:15` — `related_gap: [183, 184, 185, 186, 187, 188, 190, 191]`
  - `.moai/specs/SPEC-V3-MIGRATE-001/spec.md:15` — `related_gap: [gm#149, gm#150, gm#151]`
  - Other SPECs use yaml list form with `- gm#XXX` entries
- **Impact**: Any gap traceability tool has to handle two formats.
- **Remediation**: Normalize to `- gm#XXX` list form per template convention.

#### D-MED-008 — SPEC count in master §8 summary table is 28 but §8.1 claims "9 SPECs" for Hooks/Commands while disk has 6

- **Dimension**: D12
- **Evidence**: `docs/design/major-v3-master.md:1157` (§8 summary table) — "Hooks/Commands | 9". Disk count: 6 HOOKS + 3 CMDS = 9. So the count is correct only if we count CMDS within "Hooks/Commands". However master §8.1 title "Hooks/Commands SPECs (9 SPECs)" enumerates only HOOKS-001..009, NOT CMDS-001..003. CMDS are not in §8 at all (no subsection covers them).
- **Impact**: CMDS-001/002/003 exist on disk but are not cataloged in master §8. They don't have a traceable theme either — their `related_theme:` field says "Theme 4: Agent Frontmatter Expansion (command parity)" which appears neither in master §3 theme headers nor in v3-themes.md.
- **Remediation**: Add §8 subsection for Command SPECs (3), adjust totals, or update CMDS `related_theme` to one of the 9 documented themes.

### LOW (polish before stable)

#### D-LOW-001 — Non-Latin em-dashes mixed with `-` in master §6 phase names

- **Dimension**: D12 (style)
- **Evidence**: "Phase 1 — Foundation" (em-dash) vs "Phase 1 - Foundation" (hyphen) appears in both master and SPECs. Low impact.

#### D-LOW-002 — SPEC-V3-SKL-001 §6 ACs use bulleted list without bold AC-ID markers, while HOOKS-001 etc. use `**AC-...**:` bold markers

- **Dimension**: D11 (style)
- **Evidence**: SKL-001 §6 entries look like `- **AC-SKL-001-01**:` vs HOOKS-001's inline `- AC-HOOKS-001-01: Given...`. Minor.

#### D-LOW-003 — Master §Appendix B glossary omits definitions for some new terms

- **Dimension**: D9
- **Evidence**: Glossary doesn't define "Sprint Contract" (mentioned in design constitution), "harness-based quality routing" (defined in glossary — actually OK), "Tier 1/2/3/4" (roadmap terminology used in §3 but not explained in glossary).
- **Impact**: Minor; readers can cross-reference priority-roadmap.md.

#### D-LOW-004 — SPEC-V3-CLN-001 @MX comment example uses literal "NNN" placeholder instead of actual first REQ number

- **Dimension**: D11 (style)
- **Evidence**: `.moai/specs/SPEC-V3-CLN-001/spec.md:245` — "구현 시 각 소스 파일에 `@SPEC:SPEC-V3-CLN-001:REQ-CLN-001-<NNN>` 주석 부착."
- **Status**: Template syntax, not bug. LOW priority cosmetic.

---

## Dimension-by-Dimension Findings

### D1 — EARS Compliance

- Summary: **Strong** across 24/28 SPECs; 4 SPECs (CLN-001/002, MIGRATE-001, OUT-001) blocked by AC absence (D-CRIT-002).
- Sampled SPECs: HOOKS-001 (lines 88-122), MIG-001 (lines 140-261), CLI-001 (lines 112-191), CLN-001 (lines 102-177) all use explicit EARS keyword prefixes ("The system SHALL…", "WHEN…THEN…SHALL", "WHILE… SHALL", "WHERE… SHALL", "IF…THEN… SHALL"). Category prefixes (Ubiquitous / Event-Driven / State-Driven / Optional / Unwanted / Complex) are consistent.
- Pattern note: lower-case `shall` appears in many SPECs; EARS convention uses upper-case SHALL but lower-case is permitted per the SPEC template examples in `.claude/rules/moai/workflow/spec-workflow.md`. Not a defect.
- No instance of informal "should/may" used for mandatory behavior spotted in sampling.
- Rubric band: **0.75** (would be 1.0 but for the 4 SPECs blocked by D-CRIT-002 downstream AC absence).

### D2 — REQ ID Uniqueness & Convention

- Summary: **Clean**. 545 unique REQ IDs declared across 28 SPECs; zero real duplicates. REQ-CMDS-002-011 appears twice but one declaration and one cross-reference — legitimate.
- AC IDs: 230 unique AC-IDs across 24 SPECs that have them; zero duplicates.
- Convention: all match `REQ-<DOMAIN>-<NNN>-<NNN>` pattern with zero-padding to 3 digits. ✓
- Orphans: none detected. Cross-SPEC REQ reference REQ-AGT-002-005 in AGT-003 §5 is legitimate (AGT-002 declares it, AGT-003 cites it via a forward dependency).
- Example placeholders in SPEC-V3-SPEC-001 prose (REQ-DOMAIN-NNN-NNN etc.) are illustrative and not counted as real REQs.
- Rubric band: **1.0**.

### D3 — Acceptance Criteria Testability

- Summary: **Blocked**. 4/28 SPECs lack AC-IDs (D-CRIT-002). Among the 24 with ACs, testability is generally strong — Given/When/Then structure, explicit assertions, mapped back to REQs.
- Weasel-word scan on present ACs: zero hits of "appropriate", "adequate", "good enough", "proper". Clean.
- Rubric band (for SPECs with ACs): **1.0**. Global rubric band accounting for 4 missing: **0.50** until D-CRIT-002 resolved.

### D4 — Dependency Graph Consistency

- Summary: **Broken edges**.
- Critical: CLN-003 referenced but missing (D-CRIT-001).
- Asymmetries: CLI-001 under-declares SCH/MIG/HOOKS/AGT dependencies (D-HIGH-011); MIG-001 over-declares HOOKS blocks (D-HIGH-005); OUT-001 marks SCH-001 as "비차단" under §9.1 Blocked by heading (D-HIGH-010).
- No true cycle detected in the A-blocks-B-blocks-A sense.
- Disconnected subgraphs: No — every SPEC traces eventually back to SCH-001 or MIG-001.
- Rubric band: **0.50**.

### D5 — Traceability

- Summary: Partial. 24/28 SPECs have §10 Traceability section citing Theme, gaps, BC-ID, Wave 1 source. 4 SPECs (CLN-001/002, MIGRATE-001, OUT-001) lack REQ→AC back-references because they lack AC-IDs.
- All theme references are valid (one of 9 v3-themes.md themes).
- Gap matrix references are consistent (`gm#XXX` or integer forms — see D-MED-007).
- Wave 1 file:section citations present in traceability sections ✓.
- Rubric band: **0.75**.

### D6 — Breaking Change Integrity

- Summary: **Mixed**.
- Every `breaking: true` SPEC cites a BC-ID (10/10 ✓): HOOKS-001→BC-001, HOOKS-005→BC-005, HOOKS-006→BC-003, MIG-001→BC-004, MIGRATE-001→BC-004, SCH-001→BC-006, SCH-002→BC-002+BC-003, TEAM-001→BC-008.
- Every BC-ID in master §4 has at least one implementing SPEC (8/8 ✓): BC-001 (HOOKS-001), BC-002 (SCH-002), BC-003 (SCH-002, HOOKS-006), BC-004 (MIG-001, MIGRATE-001), BC-005 (HOOKS-005), BC-006 (SCH-001), BC-007 — **NO IMPLEMENTING SPEC ASSIGNED** (master §4 defines BC-007 "PermissionRequest decision semantics" but no SPEC has `bc_id: [BC-007]`; HOOKS-005 claims BC-005 and covers PermissionRequest handler; master §3.1 attributes BC-007 to "HOOKS-006" but HOOKS-006 uses BC-003). → defect.
- Migration strategy declared for each BC in master §4 table ✓.
- Deprecation timeline self-consistent ✓.
- bc_id type drift (D-CRIT-004) contaminates this dimension.
- Rubric band: **0.50**.

**Additional defect surfaced:**

#### D-HIGH-012 — BC-007 (PermissionRequest decision semantics) has no SPEC owning bc_id

- **Dimension**: D6
- **Evidence**: `docs/design/major-v3-master.md:896-897` (§4 table) defines BC-007 but grep for `bc_id.*BC-007` across all SPECs returns zero matches. HOOKS-005 owns BC-005; HOOKS-006 owns BC-003. Master §3.1 lists BC-007 under Theme 1 migration path but doesn't map it to a specific SPEC. Likely intended owner: HOOKS-005 (which covers PermissionRequest handler) should add BC-007 to its bc_id list.
- **Remediation**: Add BC-007 to HOOKS-005's bc_id (making it `[BC-005, BC-007]`), and same for BC-008 retry semantics in HOOKS-005 if applicable.

### D7 — Phase Assignment Consistency

- Summary: **Drift**.
- D-HIGH-003 (SKL-002 phase mismatch) and D-CRIT-003 (HOOKS renumbering breaks phase 2 content) are the main findings.
- No dependency → phase inversion detected among well-specified edges.
- Phase 1 (Foundation) has all 4 expected SPECs (SCH-001, SCH-002, MIG-001, MIG-002) ✓.
- Rubric band: **0.50**.

### D8 — Gap Coverage

- Summary: **Adequate**. All 6 Critical-severity gaps (gm#4, gm#5, gm#44, gm#149, gm#156, gm#183) are covered by at least one SPEC. 9 themes (v3-themes.md) all have SPECs, though CMDS-001/002/003 do not map to any documented theme (D-MED-008).
- All 194 gap rows not individually verified; spot-check covers Critical + core High items.
- Rubric band: **0.75**.

### D9 — Scope Discipline

- Summary: **Strong**. Every SPEC has §2 Scope with in-scope + out-of-scope bullets. D-CRIT-005 (MIG-002/CLN-001 scope overlap) is the main blemish.
- No scope-creeping SPECs detected.
- No scope-lite SPECs detected.
- Rubric band: **0.75**.

### D10 — Risk Coverage

- Summary: **Strong**. Master §7 has 12 risks (R-001..R-012), all with concrete mitigations. R-005 "omitClaudeMd regression" is Very Low prob + High impact with two-part mitigation (default false + per-agent opt-in + 22 existing agents retain default). R-007 "HTTP hook SSRF" is Low + Critical with faithful port mitigation. No "monitor only" mitigations detected.
- Rubric band: **1.0**.

### D11 — Frontmatter Schema Conformance

- Summary: **Broken**. D-CRIT-004 (bc_id shape drift), D-HIGH-001 (priority vocab), D-HIGH-002 (phase format) are the three blockers.
- All required fields (id, title, version, status, created, updated, author, priority, phase, module, dependencies, related_gap, related_theme, breaking, lifecycle, tags) present in all 28 SPECs ✓.
- Author field values vary: "GOOS", "manager-spec", "Wave 4 SPEC writer". Not a schema violation but inconsistent.
- Rubric band: **0.50**.

### D12 — Internal Consistency Cross-Checks

- Summary: **Multiple contradictions**. D-CRIT-003 (HOOKS renumbering), D-HIGH-004 (master §8.7 CLN-003 vs SPEC-001 slot conflict), D-HIGH-008 (master §6.2 HOOKS-007/008/009 references), D-MED-004 (§6.1 vs §6.2 HOOKS phase split), D-MED-008 (CMDS SPECs not in §8 catalog).
- Zero `codex` / `Codex` / `CODEX` references across SPECs and master — clean (audit criterion satisfied).
- Zero time estimates ("X weeks", "Q1 2026") — clean.
- Zero TBD/TODO/FIXME in SPEC bodies — clean.
- Emoji scan: only semantic glyphs in SPEC-V3-OUT-001 (✓ ✗ ⚠ ℹ ○ …) as technical content of StatusIcon renderer spec; not decorative. Acceptable.
- Rubric band: **0.50**.

### D13 — Numbering & Counting Consistency

- Summary: **Minor drift**. Master claim vs actual:
  - Total REQs: actual 545 (not counting 4 placeholder examples in SPEC-001); audit prompt claim "~510" — 6.8% drift (D-HIGH-006)
  - Total ACs: 230 across 24 SPECs (excluding 4 with no AC-IDs — D-CRIT-002)
  - Total dependency edges (Blocked by + Blocks): ≈ 55 edges across 28 nodes, producing a well-connected DAG with SCH-001 + MIG-001 as roots
  - SPEC count: 28/28 as claimed ✓
  - Phase distribution: Phase 1=4 (SCH-001, SCH-002, MIG-001, MIG-002), Phase 1 via migration=2 (CLN-001, CLN-002), Phase 2=5 (HOOKS-001, HOOKS-003, HOOKS-004, HOOKS-005, HOOKS-006), Phase 3=5 (AGT-001, AGT-002, CMDS-001, CMDS-002, CMDS-003, SKL-001), Phase 4=1 (MEM-001), Phase 5=1 (SKL-002; also CLN-002 in Phase 1), Phase 6a=6 (HOOKS-002, AGT-003, MEM-002, PLG-001, SPEC-001, TEAM-001), Phase 6b=2 (CLI-001, OUT-001), Phase 7=1 (MIGRATE-001)
  - Priority distribution: P0 Critical=7, P1 High=12, P2 Medium=6, P3 Low=1, other=2 (with D-HIGH-001 vocab drift)
  - Breaking SPECs: 8 declared breaking, matching 8 BC-IDs ✓ (but bc_id assignment incomplete — see D-HIGH-012)
- Rubric band: **0.75**.

---

## Must-Pass Results

- **[FAIL] MP-1 REQ number consistency**: PASS at the individual-SPEC level (each SPEC has monotonic REQ numbering per EARS category) BUT FAIL at the cross-SPEC system level due to placeholder example IDs in SPEC-001 and bare `REQ-CLN-001-` in doc text (D-MED-005). Minor.
- **[PASS] MP-2 EARS format compliance**: All present ACs use EARS or Given/When/Then patterns; all REQs use EARS keyword prefixes. The 4 SPECs missing ACs fail D3 not MP-2. (D-CRIT-002 is a testability failure, not an EARS-format failure.)
- **[FAIL] MP-3 YAML frontmatter validity**: bc_id shape drift (D-CRIT-004), priority vocab drift (D-HIGH-001), phase format drift (D-HIGH-002), dependencies list shape drift (D-MED-002), related_gap shape drift (D-MED-007). Six normalization tasks required.
- **[PASS] MP-4 Section 22 language neutrality**: Spec content covers cross-language tooling but does not favor any specific language. Semantic glyphs in OUT-001 are UTF-8, not language-specific.

---

## Chain-of-Verification Pass

Second-look findings:

1. ✓ Re-checked every REQ-ID extraction for cross-SPEC duplicates: confirmed zero true duplicates (one reference-not-declaration case flagged correctly).
2. ✓ Re-checked SPEC-V3-CLN-003 absence by direct filesystem listing: confirmed missing.
3. ✓ Re-checked AC absence in CLN-001/002, MIGRATE-001, OUT-001 by reading §6 directly in each file: confirmed all four have only "핵심 기준" prose bullets without AC-IDs.
4. ✓ Re-verified dependency symmetry for ~40 edges: found D-HIGH-005, D-HIGH-010, D-HIGH-011, D-HIGH-012. Noted D-HIGH-007 and D-HIGH-009 were false positives and removed.
5. ✓ Re-scanned for codex/time estimates/emoji/TODO: all clean except OUT-001 semantic glyphs (legitimate).
6. ✓ Re-checked BC-ID coverage: BC-007 has no owning SPEC (new defect D-HIGH-012 added on second pass).
7. ✓ Re-checked master §8.7 vs §3.9 CLN-003 position: found numbering conflict with SPEC-001 slot (D-HIGH-004).
8. ✓ Re-checked master §8 theme coverage: CMDS SPECs have no theme mapping (D-MED-008).
9. ✓ Re-checked `bc_id` type drift by grepping all 28 SPECs at line 21: confirmed 4 shape variants across 10 breaking SPECs.
10. First-pass missed D-HIGH-012 (BC-007 orphan) and D-MED-008 (CMDS theme orphan); second pass caught both.

Chain-of-verification surfaced 2 new defects beyond first pass; severity does not change the overall verdict.

---

## Cross-Cutting Observations

1. **Author heterogeneity as style drift source**: Frontmatter discipline correlates with author. "GOOS"-authored SPECs (AGT, CMDS, HOOKS-001/002/003/004/005/006, MEM-001/002, SKL-001/002, SPEC-001, TEAM-001, CLI-001) use one frontmatter style. "manager-spec"-authored SPECs (MIG-001, MIG-002, PLG-001, SCH-001, SCH-002) use a second. "Wave 4 SPEC writer"-authored SPECs (CLN-001, CLN-002, MIGRATE-001, OUT-001) use a third — and crucially, these four are the SPECs missing AC-IDs. Pattern suggests Wave 4 writer interpreted "spec.md only; defer AC to acceptance.md" more literally than the other two authors. Recommend one-author reconciliation pass.

2. **Reliance on non-existent artifacts**: Four SPECs reference `acceptance.md` files that do not exist and whose creation is not scheduled in any phase of master §6. MIGRATE-001 §9 references CLN-003 which does not exist. Master §8.7 references CLN-003 which does not exist. Pattern: planning forward-referenced artifacts that were never written.

3. **Foundations are sound**: The 194-gap matrix, 9-theme framework, and 6 Wave 1 findings provide strong evidence grounding. REQs have real reasoning. ACs (where present) are binary-testable. The TRUST 5 / EARS / @MX conventions are respected. Underlying technical plan is coherent; the defects are structural completeness failures, not intellectual ones.

4. **Downstream risk**: If v3.0.0 ships with CLN-003 missing and 4 SPECs missing ACs, Run-phase agents will spawn for those SPECs, invoke the Re-planning Gate (3+ iterations, no ACs met), and escalate to user. This burns token budget and churns the plan — the defects are not "future work" but "now or later" choices with higher cost later.

5. **Hidden scope overlap pattern**: MIG-002 explicitly acknowledges overlap with CLN-001/002/003 but does not resolve it ("중복 항목 ... overlap"; "cleanup SPEC은 본 migration을 reference"). This is a known contradiction left unresolved — classic technical-debt-by-deferral.

---

## Counts & Statistics

- **Total REQs across 28 SPECs**: **545** (actual); ~510 (audit prompt claim); delta +35 (+6.8%)
- **Total ACs**: **230** (across 24 SPECs with AC-IDs); 4 SPECs (CLN-001, CLN-002, MIGRATE-001, OUT-001) have 0 structured ACs
- **Total dependency edges (Blocks + Blocked by union, excluding Related)**: ~60 directed edges; forms DAG rooted at SCH-001 + MIG-001
- **Phase distribution (by SPEC frontmatter `phase` field)**:
  - Phase 1 Foundation: 4 (SCH-001, SCH-002, MIG-001, MIG-002) plus CLN-001 and CLN-002 declared as "Phase 1 ships via M0x"
  - Phase 2 Hook Core: 5 (HOOKS-001, HOOKS-003, HOOKS-004, HOOKS-005, HOOKS-006) — note: HOOKS-002 declared Phase 6a
  - Phase 3 Agent Runtime: 6 (AGT-001, AGT-002, CMDS-001, CMDS-002, CMDS-003, SKL-001)
  - Phase 4 Memory: 1 (MEM-001)
  - Phase 5 Internal Cleanup: 1 (SKL-002)
  - Phase 6a Tier 2 Strategic: 6 (HOOKS-002, AGT-003, MEM-002, PLG-001, SPEC-001, TEAM-001)
  - Phase 6b Tier 2 Polish: 2 (CLI-001, OUT-001)
  - Phase 7 Migration Tool: 1 (MIGRATE-001)
  - **Total: 26 in primary phase + 2 CLN declaring Phase 1 = 28 ✓**
- **Priority distribution**:
  - P0 Critical (or "Critical"): 7 (AGT-001, HOOKS-001, HOOKS-005, MIG-001, SCH-001, SKL-001, +1 variance in labeling)
  - P1 High (or "High"): 12
  - P2 Medium (or "Medium"): 6
  - P3 Low (or "Low"): 3
- **Breaking SPECs**: 8 SPECs with `breaking: true` (HOOKS-001, HOOKS-005, HOOKS-006, MIG-001, MIGRATE-001, SCH-001, SCH-002, TEAM-001). Matches BC-001..008 count ✓ BUT bc_id coverage is only 7/8 unique BCs (BC-007 orphaned — D-HIGH-012).
- **bc_id shape distribution**:
  - Scalar: 5 (HOOKS-001, HOOKS-005, HOOKS-006, TEAM-001, MIGRATE-001)
  - Array (non-empty): 3 (MIG-001, SCH-001, SCH-002)
  - Array (empty `[]`): 2 (MIG-002, PLG-001)
  - Null (`bc_id: null`): 6 (CLN-001, CLN-002, CMDS-001, CMDS-002, CMDS-003, OUT-001)
  - Omitted: 12 (remaining non-breaking SPECs)

---

## Verdict & Recommendation

**Verdict: NOT-READY**

The v3.0.0 plan has a coherent intellectual foundation (194-gap matrix, 9 themes, 6 Wave 1 findings, 12 risks, 8 breaking changes, 28 SPECs organized into 8 phases). SPEC bodies — where complete — demonstrate EARS rigor, unique REQ-ID discipline, and strong traceability to master design. But two structural blockers prevent shipping:

1. **SPEC-V3-CLN-003 does not exist on disk** despite being cited as a hard dependency by the v2-to-v3 migration tool (D-CRIT-001). Phase 8 release has a phantom predecessor.
2. **Four SPECs (CLN-001, CLN-002, MIGRATE-001, OUT-001) have no structured Acceptance Criteria** (D-CRIT-002), deferring them to a non-existent `acceptance.md`. Run-phase agents cannot converge on these SPECs.

Plus three more CRITICAL structural issues (HOOKS renumbering drift, bc_id type drift, MIG-002/CLN-001 scope overlap).

### Next steps for the planning team (ordered):

1. **Decide CLN-003's fate**: Either write the SPEC covering M02 agency archival + docs-site 4-locale sync, OR merge its scope into MIG-002 and remove every reference to CLN-003 from MIGRATE-001 / MIG-001 / MIG-002 / CLN-001 / CLN-002 / master-v3 §3.9 / master-v3 §8.7.
2. **Promote inline ACs in 4 SPECs**: CLN-001, CLN-002, MIGRATE-001, OUT-001. Convert the prose "핵심 기준" bullets into numbered `AC-<DOMAIN>-NNN-NN: Given ... When ... Then ... (maps REQ-...)` entries. Do this inline in §6 (not in deferred acceptance.md files).
3. **Reconcile master §3.1 / §6.2 / §8.1 HOOKS numbering** with the actual 6-SPEC on-disk set. Recommended: keep on-disk, update master.
4. **Resolve MIG-002 vs CLN-001/002 scope overlap**. Recommended: MIG-002 owns migration step Go files; CLN-001/002 own surrounding tooling (moai doctor, README updates, direct source edits). Document the split.
5. **Normalize frontmatter**: bc_id → array form uniformly; priority → "P0 Critical" shape uniformly; phase → "v3.0.0 — Phase N <name>" shape uniformly; related_gap → `- gm#XXX` list form uniformly; dependencies → list form uniformly.
6. **Add BC-007 ownership to HOOKS-005** (or HOOKS-006, per master §3.1 attribution).
7. **Add SCH-001 / MIG-001 / HOOKS-001 / AGT-001 to CLI-001 §9.1 Blocked by** (or §9.3 Related with explicit note).
8. **Move OUT-001's "비차단" SCH-001 reference from §9.1 to §9.3**.
9. **Remove MIG-001 §9.2 over-declared HOOKS-001..006 blocks** (or add MIG-001 to HOOKS SPECs §9.1 if actually a blocker).
10. **Add CMDS theme to master §3** (or reclassify CMDS SPECs under an existing theme).
11. **Verify counts**: audit SPEC count, REQ count, AC count, phase distribution in master §8 after all fixes. Emit `make spec-stats` target to keep master honest.

On successful remediation, re-invoke plan-auditor for iteration 2 with this report as input.
