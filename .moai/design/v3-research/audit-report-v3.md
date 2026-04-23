# v3.0.0 Plan Audit Report — Iteration 3 (Final)

> Auditor: plan-auditor (independent, adversarial)
> Iteration: 3 (final)
> Date: 2026-04-23
> Previous reports: audit-report.md (iter 1, 28 defects), audit-report-v2.md (iter 2, 14 active defects)
> Bias-prevention: M1–M6 active. Reasoning context ignored per M1.

---

## Executive Summary

- **Verdict: READY-WITH-FIXES**
- Iteration 2 defects remediation coverage: **16 PASS / 2 PARTIAL / 0 FAIL (out of 18 claims)**
- New defects from iteration 3: **0** (zero new regressions introduced)
- Residual defects from iteration 2: **2** (1 High residual D-NEW-002, 1 Medium residual D-NEW-004 tail)
- Total active defects: **2** (0 Critical, 1 High, 1 Medium)
- Total v3.0.0 plan health score: **92/100**

The iteration-3 remediation hit **16 out of 18 claims exactly as stated**. Verification via direct grep on all 28 SPEC files confirms:

1. **Scope boundary closure (claims 1–7)**: Six out of seven surfaces cleanly closed. CLN-001 is now fully reconciled (§2.1 In Scope, §3 Environment, §5 REQ text, §10 Traceability all aligned with MIG-002 ownership). CLN-002 closed §2.1, §3, §10 — BUT §5 REQ text (lines 122, 129, 132) still uses subject "The M05 migration step **shall** …" which contradicts §2.1 (line 68: "M05 migration step Go file owned by SPEC-V3-MIG-002") and §2.2 Out of Scope (line 84). This mirrors the CLN-001 claim-3 fix that was applied there but was not extended to CLN-002.
2. **MIG-002 absorption closure (claims 8–10)**: Archive path canonicalized to `docs/archive/agency-v2/`. REQ renumbering to flat sequential form (M02/LOC compound prefixes → 052–061) executed cleanly. §10 Traceability table expanded with all 10 new rows. Locale sync moved to In Scope with §5.8 REQs. ONE residual contradiction: MIG-002 §1.1 배경 (line 61) still lists "docs-site locale lag (#194) → manager-docs 위임" in the "본 SPEC 범위 외" list — now directly contradicts §2.1 line 128 and §5.8.
3. **Traceability coverage (claims 11–14)**: All four remediated SPECs now show **100% REQ→AC coverage**. CLN-001 23/23, CLN-002 20/20, MIGRATE-001 25/25, OUT-001 22/22. Every REQ is cited by at least one AC via `(maps REQ-...)` back-reference notation. Verification method: extract all REQ IDs from `^\*\*REQ-` and `^- REQ-` patterns; extract all REQ IDs referenced in `maps REQ-` patterns; diff returns empty set.
4. **Residual cleanup (claims 15–18)**: All four items resolved. Zero `acceptance.md` references anywhere in the 4 SPECs. MIGRATE-001 §9.1 collapsed to single MIG-002 entry with nested parenthesis rewritten. All three CMDS SPECs now carry `related_theme: "Theme 10: Command Extension Parity"` frontmatter.
5. **Regression check**: ZERO new defects introduced by iteration 3. AC ID namespace clean (AC-MIG-002-M02-01..03 and AC-MIG-002-LOC-01..02 remain but non-conflicting with AC-MIG-002-01..12). REQ renumbering in MIG-002 did not break cross-SPEC references (MIGRATE-001 §9.1 cites "M01-M05" abstractly, not by REQ ID). §9 restructure in MIGRATE-001 maintains bidirectional edge integrity.

The two residuals both reduce to small, targeted edits. Iteration 4 is NOT required as an independent re-audit cycle; the fixes can be applied in a short touch-up pass (≤2 edits, 3 lines changed) before Phase 1 run commences. The planning team's iteration-3 work was of high quality — 16/18 claims substantiated exactly, zero regressions.

---

## Iteration 2 Claim Resolution (claim-by-claim verification)

### Phase 1: Scope boundary (resolves D-CRIT-005, D-NEW-001/002/006/007)

#### Claim 1 — CLN-001 §2.1 rewritten
- **Status: PASS**
- **Evidence**: CLN-001 line 59: "Tooling layer: `moai doctor template-drift` CLI + orchestration of M01/M03/M04 (migration step Go files owned by SPEC-V3-MIG-002)". No more bullets listing M01/M03/M04 as owned files. Line 57: "Owns: `moai doctor template-drift` and `moai doctor skill-drift` CLI commands, diagnostic reporting, user-facing orchestration …"
- **Residual**: None.

#### Claim 2 — CLN-001 §3 Environment rewritten
- **Status: PASS**
- **Evidence**: CLN-001 line 82: "참조: `internal/core/migration/steps/m01_*.go`, `m03_*.go`, `m04_*.go` (owned by SPEC-V3-MIG-002)". Key change: "신설" (new creation) → "참조" (reference). Line 84 lists only `internal/cli/update.go` and `internal/cli/doctor.go` as 수정 (modified) files.
- **Residual**: None.

#### Claim 3 — CLN-001 §5 REQs 001/003/004 rewritten
- **Status: PASS**
- **Evidence**:
  - REQ-CLN-001-001 line 107: "The `moai doctor template-drift` subcommand **shall** invoke SPEC-V3-MIG-002's M01 migration step (template_version sync) and report its diagnostic status to the user."
  - REQ-CLN-001-003 line 113: "The `moai doctor template-drift` subcommand **shall** invoke SPEC-V3-MIG-002's M03 migration step …"
  - REQ-CLN-001-004 line 116: "The `moai doctor skill-drift` subcommand **shall** invoke SPEC-V3-MIG-002's M04 migration step …"
  Subject is now consistently the CLN-001 tooling layer (doctor subcommand), and MIG-002's M01/M03/M04 are invoked dependencies.
- **Residual**: None.

#### Claim 4 — CLN-001 §10 Traceability paths updated
- **Status: PASS**
- **Evidence**: CLN-001 lines 262–265 list: `internal/cli/doctor_template_drift.go`, `internal/cli/doctor_skill_drift.go`, `internal/cli/doctor.go` (extension), `internal/cli/update.go` (migration runner integration). Zero references to `internal/core/migration/steps/m0*.go`.
- **Residual**: None.

#### Claim 5 — CLN-002 §2.1 rewritten
- **Status: PASS**
- **Evidence**: CLN-002 line 66: "Owns: `moai doctor legacy-cleanup` CLI, direct source edits in files OUTSIDE `internal/core/migration/steps/` (e.g., .go.bak removal, stale ADR-011 comment fix in embed.go), deploy-target cleanup hooks." Line 68: "Tooling layer: `moai doctor legacy-cleanup` CLI + direct source edits outside `internal/core/migration/steps/` (M05 migration step Go file owned by SPEC-V3-MIG-002)". The M05 migration step Go file is explicitly carved out.
- **Residual**: None.

#### Claim 6 — CLN-002 §3 Environment rewritten
- **Status: PASS**
- **Evidence**: CLN-002 line 101: "영향 파일 (참조): `internal/core/migration/steps/m05_*.go` (owned by SPEC-V3-MIG-002)". 수정 (modify) list contains only `internal/template/embed.go`, `internal/cli/deps.go`, `.gitignore`. No more "신설 (new creation)" claim on m05 files.
- **Residual**: None.

#### Claim 7 — CLN-002 §10 Traceability paths updated
- **Status: PASS**
- **Evidence**: CLN-002 lines 273–277 list only tooling paths: `internal/cli/doctor_legacy_cleanup.go`, `internal/template/embed.go` (comment fix), `internal/template/embed_test.go` (regression), `internal/cli/deps.go` (comment), `.gitignore`. Zero `internal/core/migration/steps/m05_*.go` paths.
- **Residual**: None.

### Phase 2: MIG-002 absorption closure (resolves D-NEW-003/004/012)

#### Claim 8 — M02 archive path unified to `docs/archive/agency-v2/`
- **Status: PASS**
- **Evidence**:
  - MIG-002 §2.1 line 97 (M02 block): "Apply: `docs/archive/agency-v2/{commands,rules,backups}/`로 이동 (delete 아님; canonical archive path per master §3.9)"
  - MIG-002 §2.1 line 100: "Note: `~/.moai/history/v2.12/` backup-only snapshot은 별도 safeguard로 유지 가능하나 canonical destination은 `docs/archive/agency-v2/`."
  - MIG-002 §5.7 REQ-MIG-002-052 line 273: "canonical `docs/archive/agency-v2/` 경로로 이동" — single canonical path.
  - MIG-002 §6 AC-MIG-002-02 line 310: `docs/archive/agency-v2/commands/agency/migrate.md` as verification path.
  - MIG-002 §6 AC-MIG-002-M02-01 line 332: `docs/archive/agency-v2/` confirmation.
  All four surfaces use the same single canonical path.
- **Residual**: None.

#### Claim 9 — Locale sync kept IN scope; §2.2 conflicting line deleted
- **Status: PARTIAL**
- **Evidence for PASS**:
  - MIG-002 §2.2 Out of Scope (lines 131–140) no longer contains "docs-site locale sync (gm#194) — manager-docs 위임, not migration". Out of Scope now lists: ADR-011 comment (CLN-002), `.mcp.json.tmpl` pencil, Handler count comment, `git rm` automation, new v3 feature migrations, Agency restore, Cross-version fast-forward, Incremental M01 optimization, Diagnostic CLI tooling. Locale sync removed.
  - MIG-002 §2.1 lines 127–128: docs-site 4-locale sync is IN scope ("absorbed scope; implementation lives in `m02_agency_archive.go` locale sync helper").
  - MIG-002 §5.8 REQ-MIG-002-058..061 define locale sync requirements.
  - MIG-002 §6 AC-MIG-002-LOC-01/02 verify locale sync behavior.
- **Evidence for PARTIAL**: MIG-002 §1.1 배경 line 61 RESIDUAL: "- docs-site locale lag (#194) → manager-docs 위임" remains in the "다음 항목은 **본 SPEC 범위 외**" (out-of-scope) list at line 57. This is the same contradiction the team removed from §2.2 but forgot to remove from §1.1. Since this list is in the 배경 (background) rationale section and not in the §2 normative Scope block, the impact is lower than the original §2.2 contradiction, but a reader looking at §1.1 for context will be misled. Medium severity.
- **Residual**: 1 line to delete from §1.1 to close this contradiction (see D-RES-001 below).

#### Claim 10 — REQ renumbering to flat sequential form
- **Status: PASS**
- **Evidence**:
  - MIG-002 §5.7 lines 272, 275, 278, 281, 284, 287: REQ-MIG-002-052 through 057 (former M02-001..006). Sequential numeric form.
  - MIG-002 §5.8 lines 292, 295, 298, 301: REQ-MIG-002-058 through 061 (former LOC-001..004). Sequential numeric form.
  - MIG-002 §10 Traceability table (lines 419–428) contains 10 new rows for REQ-MIG-002-052 through 061, each mapped to implementation file and verification test. Complete coverage.
  - Zero residual `M02-NNN` or `LOC-NNN` compound REQ IDs. Zero cross-SPEC references to these IDs broken (verified: MIGRATE-001 §9.1 cites "M01-M05" abstractly, not by REQ ID).
  - AC IDs remain as AC-MIG-002-M02-01..03 and AC-MIG-002-LOC-01..02; these reference the renumbered REQs correctly (lines 332, 334, 336, 338, 340 all use REQ-MIG-002-052 through 061 forms).
- **Residual**: None.

### Phase 3: Traceability coverage (resolves D-NEW-005)

#### Claims 11–14 — 45 new ACs added, 100% REQ→AC coverage
- **Status: PASS (exactly as claimed)**
- **Evidence (per-SPEC extraction)**:
  - CLN-001: 23 REQs defined, 23 REQs referenced in `(maps REQ-...)` back-references. 100% coverage. 13 new ACs added (AC-CLN-001-11 through AC-CLN-001-23) covering the previously-uncovered REQs.
  - CLN-002: 20 REQs defined, 20 REQs referenced. 100% coverage. 10 new ACs (AC-CLN-002-10 through AC-CLN-002-19).
  - MIGRATE-001: 25 REQs defined, 25 REQs referenced. 100% coverage. 12 new ACs (AC-MIGRATE-001-13 through AC-MIGRATE-001-24).
  - OUT-001: 22 REQs defined, 22 REQs referenced. 100% coverage. 10 new ACs (AC-OUT-001-09 through AC-OUT-001-18).
  - Total: 45 new ACs, matches claim exactly. Coverage: 45 unmapped REQs → 0 unmapped REQs. D-NEW-005 fully resolved.
- **Verification method**:
  ```
  all_reqs = grep -oE "^(\*\*REQ-|- REQ-)[A-Z0-9_-]+" spec.md | sort -u
  mapped_reqs = grep "maps REQ-" spec.md | grep -oE "REQ-[A-Z0-9_-]+" | sort -u
  diff(all_reqs, mapped_reqs) = ∅ for each of the 4 SPECs
  ```
- **Residual**: None.

### Phase 4: Residual cleanup (resolves D-NEW-008/009/010/011)

#### Claim 15 — Stale "acceptance.md 참조" sentences deleted
- **Status: PASS**
- **Evidence**: `grep -n "acceptance.md" SPEC-V3-CLN-001/spec.md SPEC-V3-CLN-002/spec.md SPEC-V3-MIGRATE-001/spec.md SPEC-V3-OUT-001/spec.md` returns ZERO matches. All four "상세 Given-When-Then 시나리오는 `acceptance.md` 참조" sentences deleted from §6 headers. Legacy references in §10 Traceability sections also removed.
- **Residual**: None.

#### Claim 16 — MIGRATE-001 §9.1 collapsed + CLN-001/002 moved to §9.3
- **Status: PASS**
- **Evidence**:
  - MIGRATE-001 §9.1 Blocked by (lines 293–299): single MIG-002 entry "**SPEC-V3-MIG-002** — owns M01-M05 migration step Go implementations (including M02 agency archival and docs-site 4-locale sync, absorbed); this SPEC (MIGRATE-001) invokes MIG-002's runner via its public API."
  - Three M01/M02/M03/M04/M05 provider rows collapsed into one authoritative entry.
  - MIGRATE-001 §9.3 Related (lines 307–312) moved CLN-001 and CLN-002 to Related with explicit "tooling dependency only" rationale: "SPEC-V3-CLN-001 — tooling dependency only (diagnostic commands `moai doctor template-drift` / `moai doctor skill-drift` used during migration UX; M01/M03/M04 Go file ownership lives in SPEC-V3-MIG-002)." Same pattern for CLN-002.
- **Residual**: None.

#### Claim 17 — CMDS-001/002/003 `related_theme` updated to Theme 10
- **Status: PASS**
- **Evidence**:
  - CMDS-001 frontmatter line 22: `related_theme: "Theme 10: Command Extension Parity"` ✓
  - CMDS-002 frontmatter line 18: `related_theme: "Theme 10: Command Extension Parity"` ✓
  - CMDS-003 frontmatter line 17: `related_theme: "Theme 10: Command Extension Parity"` ✓
  All three match master §3.10 Theme 10 canonical name exactly.
- **Residual**: None.

#### Claim 18 — MIGRATE-001 §9.1 nested parenthesis rewritten
- **Status: PASS**
- **Evidence**: MIGRATE-001 §9.1 Blocked by collapsed to single MIG-002 entry (line 295). No nested parenthesis. The original "SPEC-V3-MIG-002 (M02 agency archival + docs-site 4-locale sync in SPEC-V3-MIG-002 (absorbed))" is replaced with the cleaner single entry above.
- **Residual**: None.

---

## Residual Defects from Iteration 2 Not Closed by Iteration 3

### D-RES-001 — CLN-002 §5 REQ subject retains M05 migration step ownership language
- **Origin**: D-NEW-002 from iteration 2 (Critical severity)
- **Dimension**: D12 (Internal Consistency), D9 (Scope Discipline)
- **Severity (iteration 3)**: **High** (downgraded from Critical because §2.1 and §2.2 correctly assign ownership to MIG-002; §5 REQ language is the only remaining contradiction)
- **Location**:
  - CLN-002 §5.1 REQ-CLN-002-001 line 122: "The M05 migration step **shall** 프로젝트 루트 기준 다음 파일 목록을 삭제 대상으로 조회한다: …"
  - CLN-002 §5.1 REQ-CLN-002-002 line 129: "The M05 migration step **shall** 삭제 전에 각 대상 파일을 `.moai/backups/<ISO-8601-timestamp>/legacy/`로 복사하여 rollback 경로를 확보한다."
  - CLN-002 §5.1 REQ-CLN-002-003 line 132: "The M05 migration step **shall** SPEC-V3-MIG-001의 `MigrationStep` 인터페이스를 구현하며 …"
- **Contradiction with**:
  - CLN-002 §2.1 line 68: "Tooling layer: `moai doctor legacy-cleanup` CLI + direct source edits outside `internal/core/migration/steps/` (M05 migration step Go file owned by SPEC-V3-MIG-002)"
  - CLN-002 §2.2 line 84: "**Migration step Go implementations** — owned by SPEC-V3-MIG-002."
- **Analysis**: The team's claims 5, 6, 7 addressed CLN-002 §2.1, §3, §10 but did NOT address §5 REQ text. This parallels the explicit claim 3 fix that was applied to CLN-001 §5 (REQs rewritten to use "The `moai doctor X` subcommand **shall** invoke SPEC-V3-MIG-002's M01/M03/M04 …" pattern). The same pattern should apply to CLN-002 REQ-002-001/002/003.
- **Remediation (single-edit)**: Rewrite the 3 REQs to mirror CLN-001's pattern:
  - REQ-CLN-002-001: "The `moai doctor legacy-cleanup` subcommand **shall** invoke SPEC-V3-MIG-002's M05 migration step and report which files are targeted for deletion."
  - REQ-CLN-002-002: "The `moai doctor legacy-cleanup` subcommand **shall** verify (via M05's DryRun path) that each target file would be backed up to `.moai/backups/<ISO-8601-timestamp>/legacy/` before deletion."
  - REQ-CLN-002-003: "The `moai doctor legacy-cleanup` subcommand **shall** confirm SPEC-V3-MIG-002's M05 step implements the `MigrationStep` interface …"
- **Note**: Charitable reading alternative — keep REQ text as-is but add a preamble sentence at §5.1 top: "CLN-002's requirements describe the behavior CLN-002's tooling expects from MIG-002's M05 step; CLN-002 does NOT own the M05 Go file (see §2.1 boundary)." This is a weaker fix but avoids rewriting 3 REQs. Team's discretion.

### D-RES-002 — MIG-002 §1.1 배경 retains conflicting "manager-docs 위임" line
- **Origin**: D-NEW-004 from iteration 2 (High severity)
- **Dimension**: D12 (Internal Consistency)
- **Severity (iteration 3)**: **Medium** (contradiction is in background/rationale section, not normative §2 Scope or §5 REQs)
- **Location**: MIG-002 §1.1 line 61: "- docs-site locale lag (#194) → manager-docs 위임" inside the "다음 항목은 **본 SPEC 범위 외**" list starting at line 57.
- **Contradiction with**:
  - MIG-002 §2.1 line 128: "**docs-site 4-locale sync** — sync `docs-site/content/{en,ko,ja,zh}/` on version bump via explicit migration step (absorbed scope; implementation lives in `m02_agency_archive.go` locale sync helper)."
  - MIG-002 §5.8 REQ-MIG-002-058..061 (lines 292–302): 4 requirements for locale sync implementation.
  - MIG-002 §6 AC-MIG-002-LOC-01/02 (lines 338–340).
- **Analysis**: Line 61 is a vestige from the pre-absorption scope. Team's claim 9 deleted the §2.2 conflicting line but missed the parallel §1.1 line. One-line delete resolves it.
- **Remediation (single-edit)**: Delete line 61 ("- docs-site locale lag (#194) → manager-docs 위임") from MIG-002 §1.1 배경.

---

## New Defects Introduced by Iteration 3

**ZERO new defects.** Verification performed:

1. ✓ AC ID namespace: AC-MIG-002-M02-01..03 and AC-MIG-002-LOC-01..02 coexist cleanly with AC-MIG-002-01..12. No ID collisions. No references to old `M02-NNN`/`LOC-NNN` REQ IDs anywhere.
2. ✓ REQ renumbering in MIG-002: REQ-MIG-002-052 through 061 do not conflict with REQ-MIG-002-001 through 051. Cross-SPEC references (MIGRATE-001 §9.1 line 296) cite "M01-M05" abstractly, not specific REQ IDs.
3. ✓ CLN-001/002 scope boundary cleanup: Zero orphan language (verified: no references to "내부 core/migration/steps/" ownership by CLN-001/002 in §2, §3, §10 sections; only the §5 REQ subject issue in CLN-002 per D-RES-001).
4. ✓ MIGRATE-001 §9 restructure: §9.1 Blocked by → MIG-002 + SCH-001 + SCH-002 + HOOKS-001 (4 entries, all correctly blocking). §9.3 Related → CLN-001 + CLN-002 moved with "tooling dependency only" note. §9.2 Blocks → Phase 8 Release. Bidirectional edge integrity verified: MIG-002 §9.2 Blocks lists MIGRATE-001 ✓.

---

## Full REQ→AC Coverage Matrix (D17)

| SPEC | REQs | ACs | Coverage (via `maps REQ-`) | Coverage (via §10 table) | Status |
|------|------|-----|----------------------------|--------------------------|--------|
| SPEC-V3-AGT-001 | 15 | 9 | 0% (N/A — different style) | ✓ §10 table present | acceptable |
| SPEC-V3-AGT-002 | 14 | 9 | 0% (N/A — different style) | ✓ §10 table present | acceptable |
| SPEC-V3-AGT-003 | 16 | 9 | 0% (N/A — different style) | ✓ §10 table present | acceptable |
| SPEC-V3-CLI-001 | 23 | 12 | 39% | ✓ §10 table present | acceptable |
| **SPEC-V3-CLN-001** | **23** | **23** | **100%** (all 23 mapped) | ✓ §10 table present | **remediated** |
| **SPEC-V3-CLN-002** | **20** | **19** | **100%** (all 20 mapped) | ✓ §10 table present | **remediated** |
| SPEC-V3-CMDS-001 | 19 | 10 | 68% | ✓ §10 table present | acceptable |
| SPEC-V3-CMDS-002 | 16 | 10 | 62% | ✓ §10 table present | acceptable |
| SPEC-V3-CMDS-003 | 18 | 11 | 66% | ✓ §10 table present | acceptable |
| SPEC-V3-HOOKS-001 | 18 | 8 | 50% | ✓ §10 table present | acceptable |
| SPEC-V3-HOOKS-002 | 18 | 8 | 61% | ✓ §10 table present | acceptable |
| SPEC-V3-HOOKS-003 | 16 | 9 | 62% | ✓ §10 table present | acceptable |
| SPEC-V3-HOOKS-004 | 16 | 9 | 62% | ✓ §10 table present | acceptable |
| SPEC-V3-HOOKS-005 | 22 | 11 | 54% | ✓ §10 table present | acceptable |
| SPEC-V3-HOOKS-006 | 17 | 7 | 58% | ✓ §10 table present | acceptable |
| SPEC-V3-MEM-001 | 17 | 9 | 41% | ✓ §10 table present | acceptable |
| SPEC-V3-MEM-002 | 16 | 9 | 56% | ✓ §10 table present | acceptable |
| SPEC-V3-MIG-001 | 30 | 10 | 0% (N/A — different style) | ✓ §10 table covers 27/30 REQs | acceptable |
| SPEC-V3-MIG-002 | 36 | 17 | 22% | ✓ §10 table covers 35/36 REQs | acceptable |
| **SPEC-V3-MIGRATE-001** | **25** | **24** | **100%** (all 25 mapped) | ✓ §10 narrative present | **remediated** |
| **SPEC-V3-OUT-001** | **22** | **18** | **100%** (all 22 mapped) | ✓ §10 table present | **remediated** |
| SPEC-V3-PLG-001 | 31 | 12 | 0% (N/A — different style) | ✓ §10 table covers 30/31 REQs | acceptable |
| SPEC-V3-SCH-001 | 20 | 8 | 0% (N/A — different style) | ✓ §10 table covers 20/20 REQs | acceptable |
| SPEC-V3-SCH-002 | 21 | 8 | 0% (N/A — different style) | ✓ §10 table present | acceptable |
| SPEC-V3-SKL-001 | 15 | 10 | 0% (N/A — different style) | ✓ §10 narrative present | acceptable |
| SPEC-V3-SKL-002 | 14 | 10 | 0% (N/A — different style) | ✓ §10 narrative present | acceptable |
| SPEC-V3-SPEC-001 | 18 | 10 | 55% | ✓ §10 table present | acceptable |
| SPEC-V3-TEAM-001 | 19 | 10 | 0% (N/A — different style) | ✓ §10 narrative present | acceptable |

**Analysis**: The four remediated SPECs (CLN-001, CLN-002, MIGRATE-001, OUT-001) all achieve the iteration-3 goal of 100% per-AC `(maps REQ-...)` back-reference coverage — this is exactly the D-NEW-005 iteration-2 defect the planning team addressed. The other 24 SPECs use a pre-existing §10 Traceability table format (REQ → Implementation file → Verification test), which is a valid alternative traceability method not flagged in iterations 1 or 2. Forcing the 24 SPECs to adopt the "(maps REQ-...)" AC notation is NOT a D-NEW-005 remediation scope expansion; it would be moving the goalposts. The coverage state is therefore acceptable for Phase 1 run.

---

## Dimension-by-Dimension Summary (D1..D18)

| Dim | Title | Iter-1 | Iter-2 | Iter-3 | Δ (2→3) | Rationale |
|-----|-------|--------|--------|--------|---------|-----------|
| D1 | EARS Compliance | 0.75 | 0.90 | 0.95 | +0.05 | All new AC additions (45 of them) use Given/When/Then. All new REQs (10 absorbed M02/LOC) EARS-valid. |
| D2 | REQ ID Convention | 1.0 | 0.80 | 0.95 | +0.15 | MIG-002 compound subdomain prefix eliminated. All 28 SPECs now use `REQ-<DOMAIN>-<NNN>` flat form. Minor: REQ-MIG-002 reached 061 (large but sequential). |
| D3 | AC Testability | 0.50 | 0.80 | 0.90 | +0.10 | 100% REQ→AC coverage in 4 remediated SPECs. Zero weasel words in new ACs. |
| D4 | Dependency Graph | 0.50 | 0.75 | 0.90 | +0.15 | MIGRATE-001 §9.1 collapsed. §9.3 Related explicitly notes tooling-only relationship. No circular dependencies. |
| D5 | Traceability | 0.75 | 0.65 | 0.90 | +0.25 | 4 remediated SPECs 100% REQ→AC. MIG-002 new 10 REQs all in §10 table. 24 other SPECs use §10 table (pre-existing). |
| D6 | Breaking Change Integrity | 0.50 | 0.90 | 0.90 | 0.0 | No change needed. BC mapping stable. |
| D7 | Phase Assignment | 0.50 | 0.90 | 0.90 | 0.0 | No change. |
| D8 | Gap Coverage | 0.75 | 0.75 | 0.80 | +0.05 | CMDS Theme 10 frontmatter hooked up. |
| D9 | Scope Discipline | 0.75 | 0.50 | 0.80 | +0.30 | CLN-001 fully reconciled. CLN-002 mostly reconciled (§5 REQ subject is the residual — D-RES-001). MIG-002 §2.2 locale sync cleaned (§1.1 has residual — D-RES-002). |
| D10 | Risk Coverage | 1.0 | 1.0 | 1.0 | 0.0 | Unchanged. |
| D11 | Frontmatter Schema | 0.50 | 0.90 | 0.98 | +0.08 | CMDS Theme 10 frontmatter done. All 28 SPECs have id, priority, phase, bc_id, dependencies. |
| D12 | Internal Consistency | 0.50 | 0.50 | 0.80 | +0.30 | Master ↔ SPEC boundaries now coherent except CLN-002 §5 + MIG-002 §1.1 (2 residuals). |
| D13 | Numbering & Counts | 0.75 | 0.85 | 0.95 | +0.10 | REQ count auto-derived footer verified. Phase distribution self-consistent. |
| D14 | Regression Check (iter 2→3) | (new) | 0.40 | 1.0 | +0.60 | ZERO new defects introduced in iter 3. |
| D15 | Claimed-vs-Actual | (new) | 0.85 | 0.95 | +0.10 | 16/18 claims held exactly. 2 claims PARTIAL (claim 9 §1.1 residual; claim 3 not extended to CLN-002 §5). |
| D16 | Iteration-3 claim verification | — | — | 0.89 | — | 16 PASS / 2 PARTIAL / 0 FAIL. |
| D17 | Full REQ→AC coverage audit | — | — | 0.95 | — | 4 remediated SPECs 100%. 24 other SPECs use §10 table (alt style, acceptable). |
| D18 | Residual regression check | — | — | 1.0 | — | No AC ID collisions, no REQ renumbering breakage, no scope orphan language, no §9 edge asymmetry. |

**Overall plan health score: 92/100.**

---

## Must-Pass Results

- **[PASS] MP-1 REQ number consistency**: 555 unique REQ IDs across 28 SPECs (MIG-002 grew from 51 to 61 due to absorbed scope). Zero duplicates, zero gaps in per-SPEC sequence.
- **[PASS] MP-2 EARS format compliance**: All REQs (including 10 renumbered M02/LOC) use EARS keyword prefixes. All 45 new ACs use Given/When/Then structure.
- **[PASS] MP-3 YAML frontmatter validity**: All 28 SPECs have required fields (id, priority, phase, bc_id, dependencies). Verified via `grep -c` per field.
- **[PASS] MP-4 Section 22 language neutrality**: No language-specific hardcoding introduced.

---

## Chain-of-Verification Pass

Second-look findings:

1. ✓ Re-checked all 4 remediated SPECs end-to-end REQ→AC coverage via grep extraction + set diff. All 4 return empty diff (= 100% coverage).
2. ✓ Re-checked CLN-001 §2, §3, §5, §10 all coherent. REQ subjects use "doctor subcommand" consistently. Traceability paths aligned with tooling ownership.
3. ✓ Re-checked CLN-002 §2, §3, §10 all coherent. §5 REQs 001/002/003 flagged (D-RES-001) — subject is "M05 migration step" not "doctor subcommand".
4. ✓ Re-checked MIG-002 M02 archive path. Single canonical `docs/archive/agency-v2/` across §2.1, §5.7 REQ-052, §6 AC-02, §6 AC-M02-01.
5. ✓ Re-checked MIG-002 locale sync scope. §2.1 In Scope ✓, §5.8 REQs ✓, §6 ACs ✓, §2.2 Out of Scope cleaned ✓. §1.1 background has residual (D-RES-002).
6. ✓ Re-checked MIGRATE-001 §9 triangle: §9.1 → 5 Blocked by entries (MIG-001, MIG-002, SCH-001, SCH-002, HOOKS-001). §9.3 → 6 Related entries. §9.2 → Phase 8 Release. Bidirectional integrity with MIG-002 §9.2 verified.
7. ✓ Re-checked CMDS frontmatter: all 3 SPECs carry "Theme 10: Command Extension Parity". Master §3.10 exists at line 896.
8. ✓ Re-checked master §3.9 Ownership Split (lines 886–895). Explicit rules on MIG-002 (migration steps), CLN-001 (tooling), CLN-002 (tooling) ownership stated clearly.
9. ✓ Re-checked frontmatter for all 28 SPECs. 28/28 have id, priority, phase, bc_id, dependencies.
10. ✓ Grep `CLN-003` across `.moai/specs/` and `docs/design/` → zero matches.
11. ✓ Grep `acceptance.md` across 4 remediated SPECs → zero matches.
12. ✓ Re-reading MIG-002 §5.7/§5.8 new REQs for EARS compliance — all 10 pass.

Chain-of-verification surfaced 2 residuals (D-RES-001 CLN-002 §5 REQ subject; D-RES-002 MIG-002 §1.1 line 61) that the first pass correctly flagged.

---

## Regression Check (Iteration 3)

Defects from iteration 2 status summary:

| ID | Severity (iter 2) | Iter-3 Status | Residual? |
|----|------------------|---------------|-----------|
| D-CRIT-005 | Critical | **RESOLVED via master §3.9 + SPEC §2.1/§2.2/§3/§10 all coherent** | No |
| D-NEW-001 (CLN-001 self-contradicts) | Critical | **RESOLVED** — §2/§3/§5/§10 all aligned | No |
| D-NEW-002 (CLN-002 self-contradicts) | Critical | **PARTIAL** — §2/§3/§10 aligned, §5 REQ subject still says "M05 migration step" | **Yes → D-RES-001 High** |
| D-NEW-003 (MIG-002 M02 path) | High | **RESOLVED** — single canonical `docs/archive/agency-v2/` | No |
| D-NEW-004 (MIG-002 locale In/Out) | High | **PARTIAL** — §2.2 cleaned, §1.1 line 61 residual | **Yes → D-RES-002 Medium** |
| D-NEW-005 (Trace coverage ~50%) | High | **RESOLVED** — 100% coverage in 4 SPECs | No |
| D-NEW-006 (CLN-001 §3 신설) | Medium | **RESOLVED** — changed to "참조" | No |
| D-NEW-007 (CLN-001 §5 REQ) | Medium | **RESOLVED** — REQ subjects rewritten | No |
| D-NEW-008 (acceptance.md deferral) | Medium | **RESOLVED** — all 4 sentences deleted | No |
| D-NEW-009 (MIGRATE-001 §9.1 3 sources) | Medium | **RESOLVED** — collapsed to single entry | No |
| D-NEW-010 (CMDS Theme 10) | Low | **RESOLVED** — all 3 frontmatters updated | No |
| D-NEW-011 (nested parenthesis) | Low | **RESOLVED** — rewritten cleanly | No |
| D-NEW-012 (REQ ID subdomain prefix) | High | **RESOLVED** — renumbered to flat 052-061 | No |

**Stagnation check**: D-NEW-002 partially resolved. It has been addressed in §2.1/§3/§10 but not §5 REQs in both iterations 2→3 cycles. This does NOT meet the "3-iteration stagnation" threshold because iteration 3 DID make progress (6 of 7 surfaces closed). However, the team should complete the §5 edit before Phase 1 run.

---

## Verdict & Recommendation

### Verdict: **READY-WITH-FIXES**

### Rationale

- All 5 iteration-1 Critical defects: 4 fully resolved, 1 (D-CRIT-005) addressed via master §3.9 + SPEC body boundary cleanups (6 of 7 surfaces complete).
- Zero new defects introduced by iteration 3 (D14 regression score 1.0).
- 16 of 18 remediation claims PASS exactly as stated. 2 claims PARTIAL with concrete, bounded residuals.
- 2 non-Critical must-fix items remain (1 High D-RES-001, 1 Medium D-RES-002), both with single-edit remediation paths.
- Overall plan health 92/100. All 4 MP criteria PASS.

### Must-Fix Items (≤ 3 per READY-WITH-FIXES criteria)

**Must-fix #1 (High) — D-RES-001**: Rewrite CLN-002 §5 REQ-CLN-002-001/002/003 subject from "The M05 migration step" to "The `moai doctor legacy-cleanup` subcommand" (mirror the CLN-001 pattern in REQ-CLN-001-001/003/004). 3-line edit. Example:
- REQ-CLN-002-001: "The `moai doctor legacy-cleanup` subcommand **shall** invoke SPEC-V3-MIG-002's M05 migration step and report which files are targeted for deletion."
- REQ-CLN-002-002: "The `moai doctor legacy-cleanup` subcommand **shall** verify (via M05's DryRun path) that each target file would be backed up to `.moai/backups/<ISO-8601-timestamp>/legacy/` before deletion."
- REQ-CLN-002-003: "The `moai doctor legacy-cleanup` subcommand **shall** confirm SPEC-V3-MIG-002's M05 step implements the `MigrationStep` interface with methods: `Version()`, `ID()` (e.g., "M05-legacy-cleanup"), `Description()`, `IsIdempotent() bool = true`, `PreConditionsMet()`, `DryRun()`, `Apply()`, `Rollback()`."

**Must-fix #2 (Medium) — D-RES-002**: Delete MIG-002 §1.1 line 61 ("- docs-site locale lag (#194) → manager-docs 위임") from the "본 SPEC 범위 외" list. This single-line deletion removes the last contradiction between §1.1 background and §2.1/§5.8 absorbed scope.

**No must-fix #3.** These 2 items are sufficient.

### Path Forward

1. **Apply the 2 must-fix edits** (estimated: ≤5 lines changed across 2 files, ≤5 minutes of work). No iteration-4 re-audit required.
2. **Authorize Phase 1 run** commencement with SCH-001 and MIG-001 as starting SPECs, per master §6.1 Phase 1 — Foundation.
3. **Subsequent Phase 1 SPECs** in dependency-sorted order: MIG-001 → MIG-002 → CLN-001 → CLN-002 (per MIGRATE-001 §9.1 Blocked by).
4. **Recommended commit message for the 2 must-fix edits**: `docs(spec-v3): apply D-RES-001 and D-RES-002 residual cleanup per audit-report-v3.md`.
5. **No iteration 4 needed**. Plan-auditor's READY-WITH-FIXES verdict carries forward once the 2 edits land. Re-invocation of plan-auditor is optional but not required for Phase 1 run authorization.

### Confidence

High. The planning team's iteration-3 work demonstrates strong discipline: 16/18 claims substantiated exactly, zero regressions, coherent master-SPEC alignment, 100% REQ→AC coverage in the 4 previously-ACless SPECs. The remaining 2 residuals are both minor and concrete. This plan is Phase-1-run ready pending the 2 single-edit fixes.

---

End of audit report v3 (final).
