---
spec_id: SPEC-V3R2-WF-001
title: Task Decomposition — Skill Consolidation Stage 1 (48 → 38)
version: "1.1.0"
status: draft
created: 2026-04-24
updated: 2026-04-25
author: manager-spec (tasks.md generation; v1.1.0 revision post plan-audit 2026-04-25)
related_plan: .moai/specs/SPEC-V3R2-WF-001/plan.md
related_spec: .moai/specs/SPEC-V3R2-WF-001/spec.md
---

# 작업 분해 — SPEC-V3R2-WF-001 Skill Consolidation Stage 1 (48 → 38)

> **범례**:
> - **File owner**: 해당 task 가 단독 소유하는 파일 경로. 다른 task 와 중복되면 순차 실행 필수.
> - **Depends on**: 선행 task ID.
> - **Wave**: plan.md 의 Wave 번호 (1.1 ~ 1.7).
> - **Parallel OK**: 동일 Wave 내에서 타 task 와 병렬 실행 가능 여부.
> - **Template + local pair**: 해당 task 가 `internal/template/templates/` + `.claude/skills/` 양쪽 수정을 모두 포함함.

---

## 전체 Task 개요 (v1.1.0)

| Wave | Task 수 | Parallel 가능 | Sequential 필수 |
|---|---|---|---|
| 1.1 Baseline | 1 | — | T1.1-1 |
| 1.2 MERGE target | 5 | T1.2-3 (신규) 만 독립; 나머지 4개 target 은 각각 다른 파일 → parallel | T1.2-1 ~ 5 |
| 1.3 Trigger union | 5 | 4개 target frontmatter 독립 → parallel 가능하되 foundation-core 5번째는 sequential | T1.3-1 ~ 5 |
| 1.4 Archive + map | 12 | Archive 11개는 독립 parallel, map 생성은 archive 완료 후 sequential | T1.4-1 ~ 12 |
| 1.5 REFACTOR + Telemetry | 8 | 각 SKILL.md 독립 → parallel | T1.5-1 ~ 8 |
| 1.6 Agent rewrite | 4 | 4개 agent 파일 독립 → parallel | T1.6-1 ~ 4 |
| 1.7 Verification | **12** (v1.1.0: T1.7-8/9/10/11/12 추가) | 일부 parallel, assertion 은 sequential | T1.7-1 ~ 12 |

**총 core task 수: 47** (v0.1.0 42개 → v1.1.0 47개; 감사 응답으로 5개 verification task 추가)
**Wave checkpoint: 6** (T1.X-END)
**Grand total (checkpoint 포함): 53**

---

## Wave 1.1 — Baseline Lock

### T1.1-1: Baseline hash 기록 (v1.1.0 — moai/workflows/ 해시 추가)

- **File owner**: `.moai/specs/SPEC-V3R2-WF-001/baseline-hashes.txt` (단독, 임시)
- **Depends on**: 없음
- **Parallel OK**: —
- **Inputs**: 20개 FROZEN/KEEP skill 의 SKILL.md 경로 + `.claude/skills/moai/workflows/*.md` 전체 (v1.1.0 추가; REQ-WF001-012 invariant 검증용)
- **Action**:
  1. `shasum -a 256 .claude/skills/moai-domain-copywriting/SKILL.md` 수행, 결과 기록
  2. `shasum -a 256 .claude/skills/moai-domain-brand-design/SKILL.md` 수행
  3. 12 KEEP + `moai` + 5 `moai-ref-*` SKILL.md 해시 기록
  4. **(v1.1.0 신규)** `.claude/skills/moai/workflows/` 하위 모든 `.md` 파일 해시 기록 (REQ-WF001-012 invariant 검증 baseline)
  5. `.moai/specs/SPEC-V3R2-WF-001/baseline-hashes.txt` 에 저장
- **Outputs**: `baseline-hashes.txt`
- **Verification**: 파일이 각 라인 `<hex-hash>  <path>` 형식; 20 skill 해시 + moai/workflows/ 전체 파일 해시 포함
- **Rollback**: 파일 삭제

---

## Wave 1.2 — MERGE Target Content Absorption

[HARD] Template + local pair: 모든 T1.2-N 은 `.claude/skills/…` 및 `internal/template/templates/.claude/skills/…` 양쪽을 수정해야 한다. Task 완료 조건에 `diff -rq` 포함.

### T1.2-1: `moai-foundation-thinking` 흡수 확장

- **File owner**:
  - `.claude/skills/moai-foundation-thinking/SKILL.md`
  - `.claude/skills/moai-foundation-thinking/modules/*.md` (추가되는 파일)
  - `.claude/skills/moai-foundation-thinking/references/*.md` (추가되는 파일)
  - `internal/template/templates/.claude/skills/moai-foundation-thinking/` (동일)
- **Depends on**: T1.1-1
- **Parallel OK**: T1.2-2, T1.2-3, T1.2-4, T1.2-5 와 병렬 가능 (파일 중복 없음)
- **Inputs**:
  - 소스 `moai-foundation-philosopher/SKILL.md` (First Principles, Five Whys)
  - 소스 `moai-workflow-thinking/SKILL.md` (Sequential Thinking MCP)
  - 현 `moai-foundation-thinking/SKILL.md` (3-framework 구조)
- **Action**:
  1. 타겟 SKILL.md 에 `## First Principles (absorbed from moai-foundation-philosopher)` 섹션 추가
  2. `## Sequential Thinking MCP (absorbed from moai-workflow-thinking)` 섹션 추가
  3. philosopher 의 `modules/*.md` 를 `moai-foundation-thinking/modules/` 로 복사
  4. philosopher 의 `references/*.md` 를 `moai-foundation-thinking/references/` 로 복사
  5. Level 2 토큰 계산 (wc -w * 0.75) → 5000 초과 시 Level 3 로 이관
  6. 동일 작업을 `internal/template/templates/.claude/skills/moai-foundation-thinking/` 에 적용
- **Outputs**: 수정된 SKILL.md + 추가 modules/references 파일
- **Verification**:
  - `diff -rq .claude/skills/moai-foundation-thinking internal/template/templates/.claude/skills/moai-foundation-thinking` → empty
  - `wc -w SKILL.md` 기준 Level 2 토큰 ≤ 5000
  - `## First Principles` 및 `## Sequential Thinking MCP` 섹션 존재
- **Rollback**: `git checkout HEAD~1 -- .claude/skills/moai-foundation-thinking internal/template/templates/.claude/skills/moai-foundation-thinking`

### T1.2-2: `moai-workflow-project` 흡수 확장

- **File owner**:
  - `.claude/skills/moai-workflow-project/SKILL.md`
  - `.claude/skills/moai-workflow-project/modules/` 확장
  - `.claude/skills/moai-workflow-project/references/` 확장
  - Template 쌍
- **Depends on**: T1.1-1
- **Parallel OK**: T1.2-1, T1.2-3, T1.2-4, T1.2-5 와 병렬
- **Inputs**:
  - `moai-workflow-templates/SKILL.md`
  - `moai-docs-generation/SKILL.md`
  - `moai-workflow-jit-docs/SKILL.md`
- **Action**:
  1. `## Template Optimization (absorbed)` 섹션 추가
  2. `## Documentation Generation (absorbed)` 섹션 추가
  3. `## JIT Docs (absorbed)` 섹션 추가
  4. 3개 소스의 `modules/`, `references/` 를 target 하위로 복사
  5. Level 2 budget 검증
  6. Template 쌍 동기화
- **Outputs**: 수정 SKILL.md + 추가 modules/references
- **Verification**: `diff -rq` empty, 3개 섹션 존재, Level 2 ≤ 5000 토큰
- **Rollback**: 동일 패턴

### T1.2-3: `moai-design-system` 신규 생성 (신설)

- **File owner**:
  - `.claude/skills/moai-design-system/` (신규 디렉터리)
  - `.claude/skills/moai-design-system/SKILL.md`
  - `.claude/skills/moai-design-system/modules/`
  - `.claude/skills/moai-design-system/references/`
  - Template 쌍
- **Depends on**: T1.1-1
- **Parallel OK**: T1.2-1, T1.2-2, T1.2-4, T1.2-5 와 병렬
- **Inputs**:
  - `moai-design-craft/SKILL.md`
  - `moai-domain-uiux/SKILL.md`
  - `moai-design-tools/SKILL.md` (Pencil portion)
- **Action**:
  1. `mkdir -p .claude/skills/moai-design-system/{modules,references}`
  2. 신규 SKILL.md 작성 — frontmatter: `description`, `triggers` (3 sources union), `related-skills`, `version: 0.1.0`
  3. Body: Quick Reference, Implementation Guide, Advanced 섹션 — 3 source 통합
  4. `moai-design-craft/modules/*` → `moai-design-system/modules/` 복사
  5. `moai-domain-uiux/modules/*`, `references/*` → target 하위 복사
  6. `moai-design-tools/reference/*` 중 Pencil 관련만 target 으로, Figma 는 `moai-design-tools` 원본 유지 (Wave 1.4 archive 시 Figma 섹션은 archive 로 동반 이동)
  7. Template 쌍 생성
- **Outputs**: 신규 `moai-design-system` 디렉터리 (template + local)
- **Verification**:
  - `[ -d .claude/skills/moai-design-system ]`
  - `[ -f .claude/skills/moai-design-system/SKILL.md ]`
  - `diff -rq` empty
  - SKILL.md frontmatter valid YAML
- **Rollback**: `rm -rf` 양쪽 디렉터리

### T1.2-4: `moai-domain-database` 확장 (cloud 흡수)

- **File owner**:
  - `.claude/skills/moai-domain-database/SKILL.md`
  - Template 쌍
- **Depends on**: T1.1-1
- **Parallel OK**: T1.2-1, T1.2-2, T1.2-3, T1.2-5 와 병렬
- **Inputs**: `moai-platform-database-cloud/SKILL.md`
- **Action**:
  1. `## Cloud Vendor Guide (absorbed from moai-platform-database-cloud)` 섹션 추가
  2. `moai-platform-database-cloud/reference/*` → `moai-domain-database/references/` 복사
  3. Level 2 budget 검증
  4. Template 쌍 동기화
- **Outputs**: 수정 SKILL.md
- **Verification**: `diff -rq` empty, cloud 섹션 존재
- **Rollback**: `git checkout HEAD~1 -- <paths>`

### T1.2-5: `moai-foundation-core` 확장 (context 흡수)

- **File owner**:
  - `.claude/skills/moai-foundation-core/SKILL.md`
  - Template 쌍
- **Depends on**: T1.1-1
- **Parallel OK**: T1.2-1 ~ 4 와 병렬
- **Inputs**: `moai-foundation-context/SKILL.md`
- **Action**:
  1. `## Token Budget (absorbed from moai-foundation-context)` 섹션에 context 내용 병합
  2. `moai-foundation-context/modules/*`, `references/*` 중 재사용 가능한 것만 `moai-foundation-core/` 로 복사
  3. Level 2 budget 검증
  4. Template 쌍 동기화
- **Outputs**: 수정 SKILL.md
- **Verification**: `diff -rq` empty, Token Budget 섹션 확장됨
- **Rollback**: 동일

### Checkpoint T1.2-END: Wave 1.2 gate

- **Depends on**: T1.2-1, T1.2-2, T1.2-3, T1.2-4, T1.2-5 모두 완료
- **Action**:
  1. `make build` 실행 — 성공 필수
  2. `diff -rq .claude/skills internal/template/templates/.claude/skills` 실행 — empty 필수
  3. 4개 merge target 의 Level 2 토큰 계산 결과 기록
- **Rollback**: Wave 1.2 전체 revert (`git revert <wave-1.2-merge-sha>`)

---

## Wave 1.3 — Trigger Keyword Union + Related-Skills Dedup

### T1.3-1: `moai-foundation-thinking` frontmatter union

- **File owner**: `moai-foundation-thinking/SKILL.md` (frontmatter 만)
- **Depends on**: T1.2-1
- **Parallel OK**: T1.3-2, T1.3-3, T1.3-4, T1.3-5 와 병렬 (다른 파일)
- **Action**:
  1. 현 `triggers:` 에 `moai-foundation-philosopher/SKILL.md` 의 triggers 병합
  2. `moai-workflow-thinking/SKILL.md` 의 triggers 병합
  3. Case-insensitive dedup
  4. `related-skills:` 에 `moai-foundation-philosopher`, `moai-workflow-thinking` 을 alias 로 추가 (retired 이름 유지)
  5. Template 쌍 동기화
- **Outputs**: frontmatter 수정
- **Verification**:
  - `diff -rq` empty
  - Source 의 모든 trigger 가 target 에 존재 (grep each)
  - Dedup 후 중복 없음

### T1.3-2: `moai-workflow-project` frontmatter union

- **File owner**: `moai-workflow-project/SKILL.md` (frontmatter)
- **Depends on**: T1.2-2
- **Parallel OK**: T1.3-1, 3, 4, 5 와 병렬
- **Action**: `moai-workflow-templates`, `moai-docs-generation`, `moai-workflow-jit-docs` triggers + related-skills 병합 및 dedup
- **Verification**: 동일 패턴

### T1.3-3: `moai-design-system` frontmatter union

- **File owner**: `moai-design-system/SKILL.md` (frontmatter)
- **Depends on**: T1.2-3
- **Parallel OK**: 병렬
- **Action**: `moai-design-craft`, `moai-domain-uiux`, `moai-design-tools` (Pencil portion) triggers 병합
- **Verification**: 동일

### T1.3-4: `moai-domain-database` frontmatter union

- **File owner**: `moai-domain-database/SKILL.md` (frontmatter)
- **Depends on**: T1.2-4
- **Parallel OK**: 병렬
- **Action**: `moai-platform-database-cloud` triggers 병합
- **Verification**: 동일

### T1.3-5: `moai-foundation-core` frontmatter union

- **File owner**: `moai-foundation-core/SKILL.md` (frontmatter)
- **Depends on**: T1.2-5
- **Parallel OK**: 병렬
- **Action**: `moai-foundation-context` triggers 병합 (context 는 minimal trigger 일 가능성)
- **Verification**: 동일

### Checkpoint T1.3-END: Wave 1.3 gate

- **Depends on**: T1.3-1 ~ 5
- **Action**: `make build` + `diff -rq` + frontmatter YAML validation (`python -c "import yaml; yaml.safe_load(...)"` 또는 유사)

---

## Wave 1.4 — RETIRE Archive + skill-rename-map

### T1.4-1 ~ T1.4-11: 각 retired skill archive

각 task 는 다음 패턴을 따른다. 11개 skill 을 병렬로 archive.

| Task | 대상 skill | Target 디렉터리 | Substitute | Verdict |
|---|---|---|---|---|
| T1.4-1 | `moai-foundation-context` | `.moai/archive/skills/v3.0/moai-foundation-context/` | `moai-foundation-core` | RETIRE (fold) |
| T1.4-2 | `moai-foundation-philosopher` | same pattern | `moai-foundation-thinking` | RETIRE (merged) |
| T1.4-3 | `moai-workflow-thinking` | same | `moai-foundation-thinking` | RETIRE (merged) |
| T1.4-4 | `moai-workflow-templates` | same | `moai-workflow-project` | RETIRE (merged) |
| T1.4-5 | `moai-workflow-jit-docs` | same | `moai-workflow-project` | RETIRE (merged) |
| T1.4-6 | `moai-domain-uiux` | same | `moai-design-system` | RETIRE (merged) |
| T1.4-7 | `moai-design-craft` | same | `moai-design-system` | RETIRE (merged) |
| T1.4-8 | `moai-design-tools` | same | **Pencil → `moai-workflow-pencil-integration`** (authoritative, per DL-4 / OQ-4 resolution); **Figma → archive/figma subdir** (no substitute, Stage 2 재평가) | RETIRE (split) |
| T1.4-9 | `moai-docs-generation` | same | `moai-workflow-project` | RETIRE (merged) |
| T1.4-10 | `moai-platform-database-cloud` | same | `moai-domain-database` | RETIRE (merged) |
| T1.4-11 | `moai-tool-svg` | same | (none) | RETIRE |

**공통 Action (per task)**:

1. `mkdir -p .moai/archive/skills/v3.0/<name>/`
2. `git mv .claude/skills/<name>/* .moai/archive/skills/v3.0/<name>/`
3. `rm -rf .claude/skills/<name>` (빈 디렉터리 제거)
4. `rm -rf internal/template/templates/.claude/skills/<name>` — **template 에서는 삭제만** (archive 에 미포함, OQ-2 해소 기준 적용)
5. `RETIRED.md` 작성 (plan.md §3 Wave 1.4 템플릿)
6. `git add` 전체

**File owner**: 각 task 는 해당 skill 디렉터리 전체 + archive 하위 단독 소유. **서로 parallel 가능.**

**Depends on**:
- T1.4-1: T1.3-5 (foundation-core 흡수 완료)
- T1.4-2, T1.4-3: T1.3-1 (thinking union 완료)
- T1.4-4, T1.4-5, T1.4-9: T1.3-2 (project union 완료)
- T1.4-6, T1.4-7, T1.4-8: T1.3-3 (design-system union 완료)
- T1.4-10: T1.3-4 (database union 완료)
- T1.4-11: T1.1-1 (baseline)

### T1.4-12: `skill-rename-map.yaml` 생성

- **File owner**: `.moai/decisions/skill-rename-map.yaml`
- **Depends on**: T1.4-1 ~ T1.4-11 모두 완료
- **Parallel OK**: 없음 (최종 집계 task)
- **Action**:
  1. `mkdir -p .moai/decisions/`
  2. plan.md §2.5 스키마로 YAML 작성
  3. `merges`, `retires`, `refactors`, `unchanged_keep` 섹션 완전 열거
- **Outputs**: `skill-rename-map.yaml`
- **Verification**:
  - `python -c "import yaml; yaml.safe_load(open('.moai/decisions/skill-rename-map.yaml'))"` 성공
  - merges 11+ entries, retires 1 entry, unchanged_keep 정확한 수

### Checkpoint T1.4-END: Wave 1.4 gate

- **Depends on**: T1.4-1 ~ 12
- **Action (v1.1.0 — OQ-1/3/7 closed)**:
  1. `make build` 실행 (exit code 0 확인)
  2. `diff -rq .claude/skills internal/template/templates/.claude/skills` → empty (archive 는 양쪽 제외)
  3. `ls -d .claude/skills/*/ | wc -l` → **38** (산식: 48 baseline − 11 archive [T1.4-1..11] + 1 new [moai-design-system, Wave 1.2] = 38). Stage 1 의 최종 디렉터리 수가 이 Wave 끝에서 확정됨; Wave 1.5/1.6 는 섹션 주입/agent rewrite 만 수행하며 디렉터리 수는 불변.
    - **[OQ-1 CLOSED]** jit-docs 는 RETIRE (T1.4-5 에서 archive).
    - **[OQ-7 CLOSED]** foundation-context 는 RETIRE (T1.4-1 에서 archive).
    - **[OQ-3 CLOSED]** §6.2 가 디렉터리 수 SoT; §6.1 은 논리적 그룹핑.
    - Stage 2 (38→24) 는 본 SPEC 범위 밖 → `SPEC-V3R3-WF-001` 에서 처리.
  4. `[ -f .moai/decisions/skill-rename-map.yaml ]`
  5. 각 archive 디렉터리에 `RETIRED.md` 존재 확인
  6. OQ-CONTRACT HUMAN GATE 확인: MIG-001 author 가 `skill-rename-map.yaml` schema v1 을 PR review 로 approve 하고 `SPEC-V3R2-MIG-001/spec.md` 가 `schema v1` 문자열을 포함

---

## Wave 1.5 — REFACTOR Label + Telemetry Window

### T1.5-1: `moai-workflow-testing` Refactor Notes

- **File owner**: `moai-workflow-testing/SKILL.md` + template 쌍
- **Depends on**: T1.1-1
- **Parallel OK**: T1.5-2 ~ 8 과 병렬
- **Action**: plan.md §3 Wave 1.5 의 `## Refactor Notes` 섹션 추가, `<NN>` 를 `238` 로 치환 (§6.2 line 238 reference)
- **Verification**: 섹션 존재, `diff -rq` empty

### T1.5-2: `moai-domain-backend` Refactor Notes

- **File owner**: `moai-domain-backend/SKILL.md` + template
- **Depends on**: T1.1-1
- **Parallel OK**: 병렬
- **Action**: Refactor Notes 섹션 추가 (§6.2 line 249)

### T1.5-3: `moai-domain-frontend` Refactor Notes

- **File owner**: `moai-domain-frontend/SKILL.md` + template
- **Depends on**: T1.1-1
- **Parallel OK**: 병렬
- **Action**: §6.2 line 250

### T1.5-4: `moai-domain-database` Refactor Notes

- **File owner**: `moai-domain-database/SKILL.md` (Wave 1.2, 1.3 이후)
- **Depends on**: T1.3-4 (database union 완료)
- **Parallel OK**: T1.5-1 ~ 3, 5 ~ 8 과 병렬 (단, T1.2-4 / T1.3-4 완료 후)
- **Action**: §6.2 line 251

### T1.5-5: `moai-platform-deployment` Refactor Notes

- **File owner**: `moai-platform-deployment/SKILL.md` + template
- **Depends on**: T1.1-1
- **Parallel OK**: 병렬
- **Action**: §6.2 line 259

### T1.5-6: `moai-platform-auth` Refactor Notes

- **File owner**: `moai-platform-auth/SKILL.md` + template
- **Depends on**: T1.1-1
- **Parallel OK**: 병렬
- **Action**: §6.2 line 260

### T1.5-7: `moai-framework-electron` Telemetry Window

- **File owner**: `moai-framework-electron/SKILL.md` + template
- **Depends on**: T1.1-1
- **Parallel OK**: 병렬
- **Action**: plan.md §3 Wave 1.5 의 `## Telemetry Window` 섹션 추가 (60-day window)

### T1.5-8: `moai-platform-chrome-extension` Telemetry Window

- **File owner**: `moai-platform-chrome-extension/SKILL.md` + template
- **Depends on**: T1.1-1
- **Parallel OK**: 병렬
- **Action**: 동일 Telemetry Window 섹션 추가

### Checkpoint T1.5-END: Wave 1.5 gate

- **Depends on**: T1.5-1 ~ 8
- **Action**: `make build` + `diff -rq` + 6개 REFACTOR skill 에 `## Refactor Notes` 섹션, 2개 UNCLEAR skill 에 `## Telemetry Window` 섹션 grep 확인

---

## Wave 1.6 — Agent Prompt Rewrite

### T1.6-1: `expert-frontend.md` rewrite

- **File owner**:
  - `.claude/agents/moai/expert-frontend.md`
  - `internal/template/templates/.claude/agents/moai/expert-frontend.md`
- **Depends on**: T1.4-12 (skill-rename-map 존재)
- **Parallel OK**: T1.6-2, 3, 4 와 병렬
- **Action**:
  1. `skill-rename-map.yaml` 로드
  2. Retired skill 이름 grep 후 `Edit` 으로 문맥별 치환
  3. `moai-design-tools` 는 문맥별 (Pencil/Figma) 다른 대체 이름 적용 → **`replace_all` 금지**
  4. Template 쌍 동기화
- **Verification**:
  - `grep -c "moai-foundation-philosopher\|moai-workflow-thinking\|moai-design-craft\|moai-design-tools\|moai-domain-uiux\|moai-platform-database-cloud\|moai-workflow-templates\|moai-docs-generation\|moai-workflow-jit-docs\|moai-tool-svg\|moai-foundation-context" .claude/agents/moai/expert-frontend.md` → **0**
  - `diff -rq .claude/agents/moai/expert-frontend.md internal/template/templates/.claude/agents/moai/expert-frontend.md` → identical (via cmp)

### T1.6-2: `manager-project.md` rewrite

- **File owner**:
  - `.claude/agents/moai/manager-project.md`
  - Template 쌍
- **Depends on**: T1.4-12
- **Parallel OK**: 병렬
- **Action**: 동일 패턴

### T1.6-3: `builder-skill.md` rewrite

- **File owner**:
  - `.claude/agents/moai/builder-skill.md`
  - Template 쌍
- **Depends on**: T1.4-12
- **Parallel OK**: 병렬
- **Action**: 동일

### T1.6-4: `manager-docs.md` rewrite

- **File owner**:
  - `.claude/agents/moai/manager-docs.md`
  - Template 쌍
- **Depends on**: T1.4-12
- **Parallel OK**: 병렬
- **Action**: 동일

### Checkpoint T1.6-END: Wave 1.6 gate

- **Depends on**: T1.6-1 ~ 4
- **Action**:
  1. `make build`
  2. 전수 grep: `grep -rl "<retired-name>" .claude/agents/ internal/template/templates/.claude/agents/` → empty
  3. `diff -rq .claude/agents internal/template/templates/.claude/agents` → 이미 empty 여야 함
  4. FROZEN 검증: `moai-domain-brand-design`, `moai-domain-copywriting` 참조는 유지됨 (grep 으로 카운트 확인)

---

## Wave 1.7 — Verification + Final `make build`

### T1.7-1: Skill count assertion (v1.1.0 — Stage 1 target 38)

- **File owner**: verification only, no write
- **Depends on**: T1.6-END
- **Parallel OK**: T1.7-2 ~ 12 과 병렬
- **Action**:
  1. `ls -d .claude/skills/*/ | wc -l` → 결과를 `wave-1.7-report.md` 에 기록
  2. `ls -d internal/template/templates/.claude/skills/*/ | wc -l` → 기록
  3. 두 값이 **38** 인지 assert (DL-1 / Option B / OQ-3 해소 결과)
- **Verification**: 38 (Stage 1 target). 실패 시 재조정 task 생성; 38→24 추가 감축은 Stage 2 (SPEC-V3R3-WF-001) 에서 처리.

### T1.7-2: Template/local parity check

- **File owner**: verification only
- **Depends on**: T1.6-END
- **Parallel OK**: 병렬
- **Action**: `diff -rq .claude/skills internal/template/templates/.claude/skills` → empty
- **Verification**: empty output 확인

### T1.7-3: FROZEN skill byte-compare

- **File owner**: verification only
- **Depends on**: T1.6-END
- **Parallel OK**: 병렬
- **Action**:
  1. `shasum -a 256` 로 현재 agency skill 2개 해시 계산
  2. Wave 1.1 baseline 과 비교
- **Verification**: exact match 필수 (mismatch 시 Wave 1.2, 1.6 중 어디서 FROZEN 위반이 발생했는지 역추적)

### T1.7-4: Retired skill grep (agent + commands + skills)

- **File owner**: verification only
- **Depends on**: T1.6-END
- **Parallel OK**: 병렬
- **Action**:
  1. `grep -rl "moai-foundation-philosopher\|moai-workflow-thinking\|moai-design-craft\|moai-design-tools\|moai-domain-uiux\|moai-platform-database-cloud\|moai-workflow-templates\|moai-docs-generation\|moai-workflow-jit-docs\|moai-tool-svg\|moai-foundation-context" .claude/ internal/template/templates/.claude/`
  2. 결과가 `.moai/archive/skills/v3.0/` 하위 및 `skill-rename-map.yaml` 하위만 포함하는지 확인
- **Verification**: `.claude/agents/`, `.claude/commands/`, `.claude/rules/`, `.claude/skills/` 에서 0 occurrence

### T1.7-5: Archive 디렉터리 완전성

- **File owner**: verification only
- **Depends on**: T1.4-END
- **Parallel OK**: 병렬
- **Action**:
  1. `ls .moai/archive/skills/v3.0/ | wc -l` → 11 이상 (T1.4 기준)
  2. 각 archive 하위에 `RETIRED.md` 존재 확인
- **Verification**: 11+ 디렉터리, 각 RETIRED.md 존재

### T1.7-6: Go test + vet

- **File owner**: verification only
- **Depends on**: T1.7-1 ~ 5
- **Parallel OK**: 없음 (자원 집약)
- **Action**:
  1. `go vet ./...`
  2. `go test -count=1 ./internal/template/...` (embedded.go 변경 검증)
  3. `go test -race ./...` (전체)
- **Verification**: 모두 pass

### T1.7-7: Wave 1.7 report 작성 + cleanup (v1.1.0 — exit code verification 명시)

- **File owner**:
  - `.moai/specs/SPEC-V3R2-WF-001/wave-1.7-report.md` (신규)
  - `.moai/specs/SPEC-V3R2-WF-001/baseline-hashes.txt` (삭제)
- **Depends on**: T1.7-1 ~ 6
- **Parallel OK**: 없음 (최종 정리)
- **Action**:
  1. `wave-1.7-report.md` 에 모든 검증 결과 기록 (T1.7-1 부터 T1.7-12 까지의 로그 포함)
  2. `baseline-hashes.txt` 삭제 (임시 파일)
  3. 최종 `make build` 실행 — **exit code 0 명시 검증**:
     ```bash
     make build
     mbrc=$?
     [ "$mbrc" -eq 0 ] || { echo "FAIL: make build exit code $mbrc"; exit 1; }
     echo "make build OK (exit 0)" >> wave-1.7-report.md
     ```
  4. `internal/template/embedded.go` 재생성 확인 (`[ -f internal/template/embedded.go ] && [ -s internal/template/embedded.go ]`)
- **Outputs**: `wave-1.7-report.md`
- **Verification**: 파일 존재, baseline-hashes.txt 제거됨, make build exit 0 로그 기록

### T1.7-8: REQ-WF001-002 verdict uniqueness validator (v1.1.0 신규)

- **File owner**: verification only + `.moai/specs/SPEC-V3R2-WF-001/scripts/verify-verdict-uniqueness.sh` (신규, no-op 외부 변경 없음)
- **Depends on**: T1.6-END
- **Parallel OK**: T1.7-1 ~ 6, 9 ~ 12 와 병렬
- **Action**:
  1. spec.md §6.2 판정표를 awk 로 parse 하여 48 행 추출
  2. 각 행에 대해 "v3 action" 컬럼 (column 4) 가 다음 중 **정확히 하나** 의 verdict 레이블을 포함하는지 검증:
     `{KEEP, REFACTOR, MERGE target, MERGE, RETIRE}` (및 변형 KEEP (FROZEN), KEEP (UNCLEAR window), KEEP (monitor), RETIRE (fold), RETIRE (merged), RETIRE (split), KEEP (absorbs Pencil portion of design-tools))
  3. 각 행에 대해 "R4 verdict" 컬럼 (column 3) 도 단일 verdict 검증 (OQ-1/OQ-7 resolution 이후 적용)
- **Verification**: 48 행 모두 단일 verdict → exit 0. 어느 행이라도 복수 verdict 또는 empty → exit 1 + 행 번호 출력

### T1.7-9: REQ-WF001-009 MIG-001 shared contract verifier (v1.1.0 신규)

- **File owner**: verification only
- **Depends on**: T1.4-12 (skill-rename-map.yaml 존재)
- **Parallel OK**: T1.7-1 ~ 8, 10 ~ 12 와 병렬
- **Action**:
  1. `.moai/decisions/skill-rename-map.yaml` YAML parse 성공
  2. top-level `version: 1` 확인
  3. `merges`, `retires`, `refactors`, `unchanged_keep` 4개 section 존재
  4. `.moai/specs/SPEC-V3R2-MIG-001/spec.md` 에서 `schema v1` 또는 `skill-rename-map.yaml` 문자열 grep (HUMAN GATE OQ-CONTRACT 완료 증명)
- **Verification**: acceptance.md AC-16 shell 블록이 zero exit

### T1.7-10: REQ-WF001-015 broken-fixture rejection verifier (v1.1.0 신규)

- **File owner**:
  - 신규 fixture: `.moai/specs/SPEC-V3R2-WF-001/fixtures/ci-reject/archive-without-retired-md/SKILL.md`
  - (RETIRED.md 는 의도적으로 누락)
- **Depends on**: T1.1-1 (fixture 는 Wave 1.1 이후 언제든 생성 가능)
- **Parallel OK**: T1.7-1 ~ 9, 11 ~ 12 와 병렬
- **Action**:
  1. Fixture 디렉터리 생성 (SKILL.md 에는 minimal frontmatter 만)
  2. acceptance.md AC-WF001-08 (b) 섹션의 verifier shell 스크립트 실행
  3. 스크립트가 **non-zero exit** 해야 테스트 pass (정상: broken fixture 가 guard 를 trip)
- **Verification**: `rc != 0` → test pass; `rc == 0` → test fail (guard 가 깨진 것)

### T1.7-11: REQ-WF001-016 trigger-drop fixture verifier (v1.1.0 신규)

- **File owner**:
  - 신규 fixture: `.moai/specs/SPEC-V3R2-WF-001/fixtures/trigger-drop/`
    - `source-skill-a/SKILL.md` (frontmatter triggers: [alpha, beta, gamma])
    - `source-skill-b/SKILL.md` (frontmatter triggers: [delta])
    - `merge-target/SKILL.md` (frontmatter triggers: [alpha, beta]) — intentionally drops gamma + delta
- **Depends on**: T1.1-1
- **Parallel OK**: T1.7-1 ~ 10, 12 와 병렬
- **Action**:
  1. Fixture 디렉터리 생성
  2. acceptance.md AC-WF001-18 의 Python 블록 실행
  3. `rc != 0` 및 `SKILL_TRIGGER_DROP: gamma` / `SKILL_TRIGGER_DROP: delta` 진단 출력 확인
- **Verification**: `rc != 0` → test pass

### T1.7-12: REQ-WF001-012 moai/workflows/ invariant verifier (v1.1.0 신규)

- **File owner**: verification only
- **Depends on**: T1.6-END, T1.1-1 (baseline-hashes.txt 에 moai/workflows/ 해시 포함 필수)
- **Parallel OK**: T1.7-1 ~ 11 과 병렬
- **Action**:
  1. Wave 1.1 T1.1-1 이 기록한 baseline 의 `.claude/skills/moai/workflows/*.md` 해시 목록과
     현재 `.claude/skills/moai/workflows/*.md` 해시를 비교
  2. 완전 일치 필수 (본 SPEC 이 WF-002 담당 디렉터리를 건드리지 않았음을 증명)
- **Verification**: acceptance.md AC-17 shell 블록이 zero exit

---

## Dependency Graph (요약, v1.1.0)

```
T1.1-1 (baseline; includes moai/workflows/ hashes per v1.1.0)
    ├─ T1.2-1 ── T1.3-1 ── T1.4-2, T1.4-3
    ├─ T1.2-2 ── T1.3-2 ── T1.4-4, T1.4-5, T1.4-9
    ├─ T1.2-3 ── T1.3-3 ── T1.4-6, T1.4-7, T1.4-8 (Pencil→pencil-integration per DL-4)
    ├─ T1.2-4 ── T1.3-4 ── T1.4-10 ── T1.5-4
    ├─ T1.2-5 ── T1.3-5 ── T1.4-1
    ├─ T1.4-11 (tool-svg, 독립)
    ├─ T1.5-1, 2, 3, 5, 6, 7, 8 (독립, 병렬)
    ├─ T1.7-10 (ci-reject fixture creation; independent, depends on T1.1-1)
    ├─ T1.7-11 (trigger-drop fixture creation; independent, depends on T1.1-1)
    │
    ├─ [T1.4-1..11] ─ T1.4-12 (skill-rename-map + OQ-CONTRACT HUMAN GATE)
    │                     └─ T1.6-1, 2, 3, 4 (병렬)
    │                             └─ T1.7-1..5, 8, 9, 12 (병렬)
    │                                     └─ T1.7-6 (go test, sequential)
    │                                             └─ T1.7-7 (report + cleanup, make build exit 0 verification)
```

New in v1.1.0:
- T1.7-8: REQ-WF001-002 verdict uniqueness validator
- T1.7-9: REQ-WF001-009 MIG-001 contract verifier
- T1.7-10: REQ-WF001-015 broken-fixture rejection (ci-reject fixture)
- T1.7-11: REQ-WF001-016 trigger-drop fixture verifier
- T1.7-12: REQ-WF001-012 moai/workflows/ invariant verifier

---

## Sequential 필수 (공유 파일)

다음 task 는 공유 파일을 수정하므로 sequential 실행:

1. **`make build`** — Wave 1.2, 1.3, 1.4, 1.5, 1.6 각 끝 + Wave 1.7 최종. 동시 실행 불가 (embedded.go 재생성).
2. **`.moai/decisions/skill-rename-map.yaml`** — T1.4-12 단독. T1.4-1~11 모두 완료 후.
3. **`internal/template/embedded.go`** — `make build` 만 건드림.

---

## 총 태스크 수: **47 core tasks + 6 checkpoints = 53 total (v1.1.0)**

- Wave 1.1: 1 (T1.1-1)
- Wave 1.2: 5 (T1.2-1 ~ 5) + 1 checkpoint
- Wave 1.3: 5 (T1.3-1 ~ 5) + 1 checkpoint
- Wave 1.4: 12 (T1.4-1 ~ 12) + 1 checkpoint
- Wave 1.5: 8 (T1.5-1 ~ 8) + 1 checkpoint
- Wave 1.6: 4 (T1.6-1 ~ 4) + 1 checkpoint
- Wave 1.7: **12** (T1.7-1 ~ 12; v1.1.0 에서 T1.7-8/9/10/11/12 추가)
- **Wave Checkpoints**: 6 (T1.X-END for X in 1.2, 1.3, 1.4, 1.5, 1.6; Wave 1.1 및 1.7 은 Wave 단위 체크포인트 없음)

Core implementation + verification tasks: **47** (v0.1.0 42 → v1.1.0 47; 감사 응답으로 T1.7-8/9/10/11/12 추가).
Grand total (checkpoints 포함): **53**.
이전 v0.1.0 의 "42 / 36" 수치는 사라짐; v1.1.0 의 수치는 "47 / 53" 로 단일화.
