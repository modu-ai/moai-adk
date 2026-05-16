# SPEC-V3R4-SPECLINT-DEBT-002 — Implementation Plan

## HISTORY

| Version | Date       | Author       | Description |
|---------|------------|--------------|-------------|
| 0.2.0   | 2026-05-16 | manager-spec | Iteration 2 revision. D5: T-Wave3-002 fixture finding count 6→3 정정 (3 missing canonical fields, 3 unknown snake_case aliases는 lint.go struct decoder가 silently drop). T-Wave3-003 expect 조건도 "≥3" → "exactly 3" 강화. D2: Wave 3 T-Wave3-001을 Case C (team/plan.md에 Pre-Write Frontmatter Checklist 섹션 신설)로 재구성 — 파일 존재하나 checklist 부재 plan-phase 실측 완료. OQ1 제거. D7: line 인용 정확도 (lint.go:518-533, plan.md:449-458/461-465). D1 영향: Wave 1 사전측정 task 갱신 (75 → 197 SPECs, snake_case dual-field 패턴 인식). |
| 0.1.0   | 2026-05-16 | manager-spec | 초기 draft. plan workflow Pre-Write Frontmatter Checklist를 lint.go 12-field canonical에 정렬하는 3-Wave 계획 작성. plan-in-main 원칙 (PR #822 doctrine) 준수, BODP signals all-¬ → base `origin/main`. 본 SPEC은 dual-schema drift 영구 해소 SPEC으로, SDF-001 hotfix 사건 (commit `b2b7f32c7`)의 follow-up. |

---

## 1. Overview

본 plan은 SPEC-V3R4-SPECLINT-DEBT-002의 5개 EARS requirements를 3-Wave로 분해하여 실행한다. 각 Wave는 독립 검증 가능하며, Wave 간 의존성은 plan.md → SSOT 문서 → team variant + fixture 순서.

### 1.1 Execution Mode

- **Plan-in-main**: BODP signals A=¬ B=¬ C=¬ → main @ origin/main (CLAUDE.local.md §18.12 BODP)
- **Worktree**: 사용 안 함 (Phase Discipline Step 1 — plan-in-main + 단일 source file 위주 수정)
- **Branch (run-phase)**: `feat/SPEC-V3R4-SPECLINT-DEBT-002`
- **Lifecycle**: spec-anchored (plan + implementation 함께 evolve)

### 1.2 Wave Breakdown

| Wave | Goal | Files | EARS coverage | Priority |
|------|------|-------|---------------|----------|
| Wave 1 | plan.md Pre-Write Checklist 12-field 정렬 | `.claude/skills/moai/workflows/plan.md` | REQ-001, REQ-002 | High |
| Wave 2 | SSOT 문서 신설 + cross-reference | `.claude/rules/moai/development/spec-frontmatter-schema.md`, plan.md 주석, lint.go 주석 | REQ-005 | High |
| Wave 3 | team/plan.md에 Pre-Write Checklist 섹션 신설 (Case C) + regression fixture | `.claude/skills/moai/team/plan.md`, `internal/spec/testdata/frontmatter-schema/` | REQ-003, REQ-004 | Medium |

**Corpus context (plan-phase 실측 baseline)**: 197 SPECs (main HEAD `b14290946`), 51건 `created_at:` dual-field, 53건 `labels:` dual-field, lint regression 0건. 본 SPEC은 미래 SPEC 작성 가이드만 정렬 — 기존 197 SPECs 마이그레이션은 opportunistic (별도 SPEC 후보).

---

## 2. Wave 1 — plan.md Pre-Write Checklist 12-field 정렬

### 2.1 Objectives

- `.claude/skills/moai/workflows/plan.md` (lines 447-473) "Pre-Write Frontmatter Checklist" body를 lint.go canonical 12-field로 완전 정렬.
- "Rejected legacy aliases" 섹션 방향 반전 — snake_case (`created_at:`, `updated_at:`, `labels:`)를 reject로 명시.
- manager-spec self-audit 절차 12-field에 맞춰 갱신.

### 2.2 Tasks

#### T-Wave1-001: Required fields 9 → 12 갱신

- **What**: plan.md lines 449-458 의 "Required 9 fields" → "Required 12 fields"로 변경. canonical 순서:
  1. `id: SPEC-{DOMAIN}-{NUM}`
  2. `title: <SPEC title in H1 equivalent>`
  3. `version: "X.Y.Z"` (quoted semver)
  4. `status: draft` (enum: draft|planned|in-progress|implemented|completed|superseded|archived|rejected)
  5. `created: YYYY-MM-DD` (ISO date, NOT `created_at`)
  6. `updated: YYYY-MM-DD` (ISO date, NOT `updated_at`)
  7. `author: <name>`
  8. `priority: High|Medium|Low|Critical` (alt: P0|P1|P2|P3)
  9. `phase: <release-phase string>`
  10. `module: <affected module paths>`
  11. `lifecycle: spec-first | spec-anchored | spec-as-source`
  12. `tags: "csv, of, lowercase, tags"` (CSV string, NOT YAML array)
- **Where**: `.claude/skills/moai/workflows/plan.md:449-458`
- **Verification**: 갱신 후 12개 fields가 명시되어 있고, lint.go FrontmatterSchemaRule.Check() `required[]` slice (`internal/spec/lint.go:518-533`) 12개 fields와 1:1 매칭됨을 grep으로 확인.

#### T-Wave1-002: Optional fields 섹션 신설

- **What**: 12 required 다음에 "Optional fields (include when applicable)" 섹션 신설:
  - `issue_number: <int> | null` — GitHub Issue 번호 (있을 때만 명시)
  - `dependencies: []` — blocking SPEC IDs
  - `related_specs: []` — non-blocking references
  - `breaking: true|false` — backwards compatibility 영향
  - `bc_id: []` — breaking change IDs (when breaking=true)
  - `related_theme: <string>` — release theme grouping
  - `target_release: vX.Y.Z` — release target
  - `superseded_by: SPEC-NEW-001` (when status=superseded)
- **Where**: plan.md, T-Wave1-001 갱신 직후
- **Verification**: optional fields 목록이 SPEC-V3R4-SPECLINT-DEBT-001 spec.md 사용 패턴과 일치.

#### T-Wave1-003: "Rejected legacy aliases" 섹션 방향 반전

- **What**: plan.md lines 461-465 의 reject 대상 변경:
  - **이전 (drift 방향)**: `created:`, `updated:`, `spec_id:`, `title:` in frontmatter
  - **수정 (canonical 방향)**: `created_at:`, `updated_at:`, `labels:`, `spec_id:` (rejected aliases)
  - `title:` in frontmatter: KEEP as rejected? → **NO, REMOVE from rejection list**. lint.go FrontmatterSchemaRule.Check()가 `title`을 required로 검증함 (`internal/spec/lint.go:521` `{"title", fm.Title}`). H1 heading만 두면 lint fail.
- **Where**: `.claude/skills/moai/workflows/plan.md:461-465`
- **Verification**: 수정 후 plan.md를 따라 작성한 frontmatter가 lint.go에서 0 finding 통과해야 함.

#### T-Wave1-004: Pre-write gate behavior 갱신

- **What**: plan.md lines 467-471 의 4-step gate behavior에서 "9-field checklist" → "12-field checklist" 텍스트 갱신. 추가로:
  - Step 2 self-audit에서 12 fields 모두 verify + optional fields 적절성 확인.
  - Step 4 plan-auditor 재검증에서 동일 schema 적용.
- **Where**: `.claude/skills/moai/workflows/plan.md:467-471`
- **Verification**: gate behavior 텍스트가 12-field schema와 정합.

#### T-Wave1-005: Wave 1 검증 (verification gate)

- **What**: 다음 명령으로 self-verification:
  - `grep -E '^- \[ \] `(id|title|version|status|created|updated|author|priority|phase|module|lifecycle|tags)`:' .claude/skills/moai/workflows/plan.md` → 12 lines 출력 확인
  - `grep -E '`(created_at|updated_at|labels)`' .claude/skills/moai/workflows/plan.md` → reject 컨텍스트에서만 등장 확인
  - Test fixture (Wave 3) 적용 전 dry-run: 본 SPEC-V3R4-SPECLINT-DEBT-002 spec.md 자체 lint.go 통과 (이미 12-field로 작성됨)
- **Where**: bash command 실행
- **Verification**: 3개 grep 모두 expected output.

### 2.3 Risks

- **R-W1-001 (low)**: plan.md 9-field 본문에 다른 cross-reference (CLAUDE.md, schema doc 등)가 의존. → Mitigation: Wave 1 작업 전 `grep -rn "9 required fields" .claude/ CLAUDE.md` 확인 후 동반 업데이트.
- **R-W1-002 (medium)**: "Rejected legacy aliases" 반전이 manager-spec 동작에 cascading 영향 (실제 동작 vs 문서 lag). → Mitigation: Wave 3 fixture로 실제 동작 검증.
- **R-W1-003 (low)**: 51+53 dual-field SPECs (197 corpus 기준)가 본 task 적용 후 cleanup 압박. → Mitigation: 본 SPEC scope 외임을 spec.md §1.2 #1 + F-OUT-001에 명시 — cleanup은 별도 SPEC 후보 (opportunistic, 향후 redundant 메타데이터 제거 follow-up).

---

## 3. Wave 2 — SSOT 문서 신설 + Cross-reference

### 3.1 Objectives

- 단일 진실 공급원 (SSOT) 문서 `.claude/rules/moai/development/spec-frontmatter-schema.md` 신설.
- plan.md, lint.go 주석, plan-auditor 가 모두 SSOT 문서를 참조하도록 cross-reference 추가.

### 3.2 Tasks

#### T-Wave2-001: SSOT 문서 작성

- **What**: `.claude/rules/moai/development/spec-frontmatter-schema.md` 신설. 내용 구성:
  - § Identity: lint.go FrontmatterSchemaRule이 enforce, plan workflow가 guide, plan-auditor가 second-line-of-defense.
  - § Required 12 Fields (canonical order, table format)
  - § Optional Fields (when applicable, table format)
  - § Rejected Legacy Aliases (avoid these — fail-closed)
  - § Examples: 정확한 PASS fixture + 의도적 FAIL fixture (각 1개)
  - § Enforcement Layers: plan.md (Phase 2 pre-write), lint.go (FrontmatterSchemaRule), plan-auditor (independent re-verify), CI (spec-lint job)
  - § Cross-references: plan.md, lint.go:502-569, plan-auditor doc, CLAUDE.local.md §14 frontmatter section
- **Where**: 신규 파일 `.claude/rules/moai/development/spec-frontmatter-schema.md`
- **Verification**: 파일 존재 + 12-field table 포함 + cross-reference 링크 4개+ 포함.

#### T-Wave2-002: plan.md 본문에 SSOT cross-reference 추가

- **What**: `.claude/skills/moai/workflows/plan.md` § Pre-Write Frontmatter Checklist 상단에 다음 한 줄 추가:
  - `> **Canonical source**: see `.claude/rules/moai/development/spec-frontmatter-schema.md` for the SSOT 12-field schema. Changes to schema MUST update both this checklist and the SSOT document.`
- **Where**: `.claude/skills/moai/workflows/plan.md:447` 직전 (checklist 헤더 직후)
- **Verification**: grep으로 cross-reference 한 줄 존재 확인.

#### T-Wave2-003: lint.go 주석에 SSOT cross-reference 추가

- **What**: `internal/spec/lint.go:502-503` `FrontmatterSchemaRule` Godoc 주석에 SSOT 참조 추가:
  - 기존: `// FrontmatterSchemaRule checks SPEC frontmatter schema\n// Implements REQ-SPC-003-006`
  - 갱신: `// FrontmatterSchemaRule checks SPEC frontmatter schema\n// Implements REQ-SPC-003-006\n// Canonical schema: .claude/rules/moai/development/spec-frontmatter-schema.md`
- **Where**: `internal/spec/lint.go:502-503`
- **Verification**: `grep "spec-frontmatter-schema.md" internal/spec/lint.go` 1+ matches.

#### T-Wave2-004: coding-standards.md (또는 유사 cross-reference 허브)에 link

- **What**: `.claude/rules/moai/development/coding-standards.md` 의 적절한 section (예: "Single source of truth" 또는 새 section "SPEC schema source")에 SSOT 문서 link 1줄 추가. 또는 `.claude/rules/moai/workflow/spec-workflow.md` § Plan Phase에 link 추가.
- **Where**: 후보 1: `.claude/rules/moai/development/coding-standards.md`. 후보 2: `.claude/rules/moai/workflow/spec-workflow.md`. run-phase에서 적절성 판단 후 결정.
- **Verification**: 신설 SSOT 문서가 코드베이스 grep으로 발견 가능 (최소 3곳: plan.md, lint.go, 1+ hub).

### 3.3 Risks

- **R-W2-001 (low)**: SSOT 문서 신설이 .claude/rules/moai/development/ 디렉토리의 다른 문서와 중복. → Mitigation: 사전 `ls .claude/rules/moai/development/` 검토 + 기존 frontmatter 관련 문서 부재 확인.
- **R-W2-002 (low)**: lint.go 주석 추가가 Go test 실패 트리거 가능 (e.g., golint comment ratio). → Mitigation: Wave 2 후 `go test ./internal/spec/...` 실행.

---

## 4. Wave 3 — team/plan.md Pre-Write Checklist 신설 (Case C) + Regression fixture

### 4.1 Objectives

- `.claude/skills/moai/team/plan.md`에 **Pre-Write Frontmatter Checklist 섹션 신설** (Case C: 신규 추가, 기존 sync 아님). plan-phase 실측: 파일 존재 (8065 bytes, 2026-05-14) + 12-field/9-field checklist 부재 (`grep -i "frontmatter\|12 required\|9 required" team/plan.md` → 0 matches).
- `internal/spec/testdata/frontmatter-schema/` regression fixture 추가 (PASS + FAIL).

### 4.2 Tasks

#### T-Wave3-001: team/plan.md에 Pre-Write Frontmatter Checklist 섹션 신설 (Case C)

- **What**: `.claude/skills/moai/team/plan.md`의 manager-spec delegation 직후 (line 155 "Delegate SPEC creation to manager-spec subagent" 다음 + line 160 "SPEC output at:" 직전)에 solo plan.md의 Pre-Write Frontmatter Checklist 섹션을 mirror하여 신설. Solo plan.md (lines 446-471)의 12-field checklist + rejected aliases + pre-write gate behavior 구조를 verbatim 복사 후 team-mode 적응 (예: "manager-spec coordinator (team-mode)" 명시).
- **Insertion point**: `.claude/skills/moai/team/plan.md:155` (delegation step) 직후, "## Phase 4: User Approval" 헤더 직전.
- **Expected LOC**: ~75 LOC 삽입 (header + 12 fields + optional fields + rejected aliases + gate behavior + 2 cross-reference lines).
- **Verification**:
  - `grep -A 30 'Pre-Write Frontmatter Checklist' .claude/skills/moai/team/plan.md` → 12-field 목록 출력.
  - `diff <(grep -A 25 'Required 12 fields' .claude/skills/moai/workflows/plan.md) <(grep -A 25 'Required 12 fields' .claude/skills/moai/team/plan.md)` → 차이 없음 (whitespace 제외) 또는 team-mode 적응 단어만 다름.
- **Rationale (Case C 채택)**: team 모드도 SPEC 작성을 manager-spec subagent에 위임 (team/plan.md:157)하므로 solo와 team에서 동일한 12-field schema 가이드를 제공해야 일관성 확보.

#### T-Wave3-002: Regression fixture 디렉토리 생성

- **What**: `internal/spec/testdata/frontmatter-schema/` 디렉토리 생성. 다음 2개 fixture 추가:
  - `valid-12-field/spec.md`: lint.go canonical 12-field 모두 포함. 기대 결과: **0 FrontmatterInvalid findings**.
  - `invalid-snake-case-only/spec.md`: snake_case alias만 보유 (`created_at:`, `updated_at:`, `labels:`) + canonical 누락 (`created:`, `updated:`, `tags:`). 다른 9 required fields (`id`, `title`, `version`, `status`, `author`, `priority`, `phase`, `module`, `lifecycle`)는 모두 정상. 기대 결과: **정확히 3 FrontmatterInvalid findings** (각각 `created`, `updated`, `tags` 누락). lint.go 동작 근거: `internal/spec/lint.go:534-545` — unknown YAML keys (`created_at`/`updated_at`/`labels`)는 struct decoder가 silently drop, 선언된 `required[]` slice의 12 fields 중 빈 값만 finding 생성.
- **Where**: `internal/spec/testdata/frontmatter-schema/`
- **Verification**: fixture 2개 파일 존재 + (FAIL fixture) lint.go 직접 실행 시 정확히 3개 finding ("Frontmatter required field missing: created", "...: updated", "...: tags") 출력.

#### T-Wave3-003: Regression test 추가

- **What**: `internal/spec/lint_test.go` 또는 `internal/spec/lint_frontmatter_schema_test.go` 신설:
  - `TestFrontmatterSchemaRule_Valid12Field`: valid fixture 로드 → FrontmatterSchemaRule.Check() → expect **exactly 0 findings**.
  - `TestFrontmatterSchemaRule_SnakeCaseRejected`: invalid fixture 로드 → expect **exactly 3** "Frontmatter required field missing" findings — 정확히 `created`, `updated`, `tags` 3개 field. assertion은 `len(findings) == 3` + finding messages가 set으로 정확히 {created, updated, tags}와 일치.
- **Where**: `internal/spec/lint_test.go` 또는 신규 파일
- **Verification**: `go test ./internal/spec/ -run TestFrontmatterSchemaRule -v` 두 테스트 모두 PASS. count 정확성 (3, not ≥3)으로 AC-002와 정합.

#### T-Wave3-004: plan-auditor schema check fixture (optional, deferred)

- **What**: plan-auditor가 plan.md 본문을 parse하여 12-field schema 강제 여부 검증. plan-auditor 구현이 별도 SPEC scope에 있을 수 있어 run-phase에서 적절성 판단:
  - plan-auditor가 본문 schema 강제 → 본 SPEC에서 testfixture로 검증.
  - plan-auditor가 frontmatter only 검증 → Wave 3 fixture로 충분, 추가 작업 없음.
- **Where**: plan-auditor codebase (run-phase에서 위치 확인)
- **Verification**: plan-auditor 동작 분석 후 결정.

### 4.3 Risks

- **R-W3-001 (low)**: team/plan.md 신설 위치 (line 155 vs 다른 위치)가 team variant 본문 흐름과 어색. → Mitigation: plan-phase에서 위치 결정 완료 (delegation 직후 + Phase 4 헤더 직전이 자연스러움 확인).
- **R-W3-002 (low)**: Regression test가 lint.go internal API (FrontmatterSchemaRule struct) 의존. → Mitigation: 기존 lint_test.go 패턴 재사용 (FrontmatterSchemaRule이 이미 export됨).
- **R-W3-003 (low)**: invalid-snake-case-only fixture 작성 시 다른 12 required field 누락하면 finding 수가 3 초과. → Mitigation: T-Wave3-002 fixture 작성 시 9 non-snake fields (id/title/version/status/author/priority/phase/module/lifecycle) 모두 정상값 보유 — 정확히 3 finding 생성 보장.

---

## 5. Technical Approach

### 5.1 Lint.go canonical 보호 전략

본 SPEC은 lint.go FrontmatterSchemaRule을 **canonical source**로 고정한다. 이는 다음 근거:

- **현실 부합**: 197 SPECs 모두 canonical (`created:`/`updated:`/`tags:`) 보유 → lint regression 0건. 51건의 `created_at:` + 53건의 `labels:` snake_case alias는 canonical과 dual-field로 공존 (redundant duplication, 능동적 drift 아님). plan.md 정렬로 신규 SPEC의 dual-field 양산 차단.
- **언어 일관성**: lint.go는 Go 코드. Go/YAML 컨벤션에 `created`/`updated`/`tags` (no underscore suffix) 자연스러움. snake_case `_at` suffix는 SQL/Rails 관례.
- **마이그레이션 비용**: lint.go 변경 시 197 SPECs 양방향 migration + `tags`→`labels` array 변환 + lint.go semantic 변경 위험 — high cost. plan.md 변경 시 단일 markdown 본문 수정 — low cost.

### 5.2 SSOT 문서 설계 원칙

- **단일 source**: schema 변경 시 단 한 곳만 수정 (`.claude/rules/moai/development/spec-frontmatter-schema.md`). plan.md / lint.go 주석 / plan-auditor가 cross-reference로 참조.
- **Discoverability**: 코드베이스 grep으로 3+ 위치에서 SSOT 파일명 발견 가능.
- **Living document**: 향후 field 추가/제거 SPEC이 생길 때 SSOT 문서를 첫 번째 수정 대상으로 명시.

### 5.3 Regression fixture 설계 원칙

- **PASS fixture**: 본 SPEC spec.md와 동일 패턴 (12-field 모두 + optional metadata 일부).
- **FAIL fixture**: SDF-001 hotfix 직전 상태 재현 (snake_case 사용 → lint fail). hotfix가 자동 재발 방지됨을 입증.
- **Brittleness 최소화**: fixture는 schema field 존재만 검증. 값 의미 검증은 별도 rule (DependencyExistsRule 등).

---

## 6. Risks (Aggregate)

| Risk | Wave | Probability | Impact | Mitigation |
|------|------|-------------|--------|------------|
| plan.md 9-field 본문 다른 cross-reference 존재 | W1 | Low | Medium | 사전 `grep -rn "9 required fields" .claude/ CLAUDE.md` |
| "Rejected legacy aliases" 반전 cascading 영향 | W1 | Medium | Medium | Wave 3 fixture로 실측 검증 |
| 51+53 dual-field SPECs cleanup 압박 (scope creep) | W1 | Medium | Low | spec.md §1.2 #1 + F-OUT-001 명시 — opportunistic, 별도 SPEC 후보 |
| SSOT 신설이 기존 문서와 중복 | W2 | Low | Low | 사전 `ls .claude/rules/moai/development/` 검토 |
| lint.go 주석 추가가 Go test 실패 트리거 | W2 | Low | Low | Wave 2 후 `go test ./internal/spec/...` |
| team/plan.md Pre-Write Checklist 신설 위치 어색 | W3 | Low | Low | plan-phase 위치 결정 완료 (delegation 직후, Phase 4 직전) |
| plan-auditor 검증 범위 미확정 | W3 | Medium | Low | run-phase에서 plan-auditor 동작 분석 후 결정 |
| FAIL fixture 작성 오류 → finding count 3 초과 | W3 | Low | Medium | 9 non-snake fields 모두 정상값 보유 보장 (T-Wave3-002) |

---

## 7. Open Questions

다음 OQ는 run-phase에서 해소된다. (OQ1 — team/plan.md 존재 여부 — plan-phase 검증 완료, 제거)

- **OQ2**: plan-auditor가 plan.md 본문 schema를 검증하는지, frontmatter only 검증하는지 (Wave 3 T-Wave3-004 결정).
- **OQ3**: SSOT 문서를 hub 1곳에 cross-reference 추가할 위치 — `.claude/rules/moai/development/coding-standards.md` vs `.claude/rules/moai/workflow/spec-workflow.md` (Wave 2 T-Wave2-004 결정).
- **OQ4**: lint.go 주석 추가 시 기존 godoc 스타일 일관성 확인 (Wave 2 T-Wave2-003 실행 시).

---

## 8. Out of Scope (Plan-level)

본 plan에서 **명시적으로 제외**되는 항목 (spec.md §5와 동일):

- 197 기존 SPECs frontmatter dual-field cleanup (51+53 redundant snake_case aliases — opportunistic 별도 SPEC)
- lint.go FrontmatterSchemaRule 의미적 변경 (field 추가/제거)
- title/phase/module/lifecycle 값 검증 강화 (별도 SPEC 후보)
- issue_number → lint.go required 추가
- snake_case 표준화
- plan workflow body의 Pre-Write Frontmatter Checklist 외 다른 section 변경
- DependencyExistsRule / DependencyCycleRule 등 다른 lint rule 변경
- SPEC frontmatter YAML parser 변경 (internal/spec/parse.go)

---

## 9. Estimation

Priority-based decomposition (no time estimates per CLAUDE.md HARD rule "Never use time predictions in plans or reports"):

| Wave | Priority | Files modified | LOC delta | Test impact |
|------|----------|----------------|-----------|-------------|
| Wave 1 | High | 1 file (plan.md) | ~30 lines (textual revision) | 0 (markdown only) |
| Wave 2 | High | 3-4 files (SSOT new + plan.md + lint.go + hub) | ~100 lines (new SSOT) + ~3 lines (cross-refs) | 0 (comment-level lint.go change) |
| Wave 3 | Medium | 4 files (team/plan.md ~75 LOC 신설 + 2 fixtures + 1 test) | ~200 lines (75 + fixtures + tests) | +2 unit tests |

---

## 10. Definition of Done (Implementation Phase)

1. Wave 1: plan.md Pre-Write Frontmatter Checklist 12-field로 갱신, "Rejected legacy aliases" 반전 (REQ-001, REQ-002).
2. Wave 2: SSOT 문서 신설 + 3+ 위치 cross-reference (REQ-005).
3. Wave 3: team/plan.md에 Pre-Write Frontmatter Checklist 섹션 신설 (Case C, ~75 LOC) + regression fixture 2건 + test 2건 (REQ-003, REQ-004).
4. `moai spec lint --strict` 본 SPEC 디렉토리 0 ERROR / 0 WARNING.
5. `go test ./internal/spec/ -run TestFrontmatterSchemaRule` PASS.
6. AC-SDBT-002-001 ~ 005 모두 verified (acceptance.md 참조).
7. plan-auditor (run-phase Phase 0.5) PASS.
8. PR merge → CHANGELOG `[Unreleased]` 추가 (sync-phase).
