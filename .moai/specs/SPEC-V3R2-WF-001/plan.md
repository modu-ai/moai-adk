---
spec_id: SPEC-V3R2-WF-001
title: Implementation Plan — Skill Consolidation Stage 1 (48 → 38)
version: "1.1.0"
status: draft
created: 2026-04-24
updated: 2026-04-25
author: manager-spec (plan.md generation; v1.1.0 revision post plan-audit 2026-04-25)
priority: P1 High
development_mode: DDD
scope_scale: XL
blast_radius: maximum
related_spec: .moai/specs/SPEC-V3R2-WF-001/spec.md
related_audit: .moai/reports/plan-audit/V3R2-triage-audit-2026-04-24.md, .moai/reports/plan-audit/SPEC-V3R2-WF-001-2026-04-25.md
---

# 구현 계획 — SPEC-V3R2-WF-001 Skill Consolidation Stage 1 (48 → 38)

> **스코프 경고**: 이 SPEC은 v3R2 Wave 1 중 가장 큰 폭발 반경(maximum blast radius)을 가진다.
> `.claude/skills/` 및 `internal/template/templates/.claude/skills/` 양쪽에서 **48개 디렉터리 → 38개** (Stage 1), **11개 디렉터리 삭제 + 아카이브**, **4개 merge target 재구성 + 1개 신규**, **다수 agent prompt 문자열 치환**을 수반한다. 단일 PR 머지는 리스크가 지나치게 크므로 본 계획은 7개 Wave(1.1~1.7)로 분할하고, 각 Wave는 독립적으로 revert 가능하도록 설계한다.
>
> **Stage 1 vs Stage 2**: 본 SPEC은 48 → 38 (Stage 1) 만 집행한다. 38 → 24 추가 감축은 `SPEC-V3R3-WF-001` 로 예약되며 Stage 2 에서 monitor/UNCLEAR/REFACTOR 재평가를 수행한다. 결정 근거는 §Decision Log 참조.

---

## 1. 목표 (Objectives)

- **1차 목표 (Stage 1)**: `.claude/skills/` 디렉터리 수를 48개에서 38개로 축소 (48 − 11 RETIRE + 1 NEW = 38) 하면서 §6.2 판정표(spec.md)의 48개 entry 모두에 단일 verdict를 집행한다 (REQ-WF001-001).
- **2차 목표**: Template-First 규칙(CLAUDE.local.md §2)에 따라 `internal/template/templates/.claude/skills/` 와 `.claude/skills/` 을 모든 Wave 끝에서 byte-identical 상태로 유지한다 (REQ-WF001-006).
- **3차 목표**: Retired skill 이름이 하드코딩된 agent prompt 4개 파일(`expert-frontend.md`, `manager-project.md`, `builder-skill.md`, `manager-docs.md`)을 동일 commit 안에서 rewrite한다 (REQ-WF001-014).
- **4차 목표**: SPEC-V3R2-MIG-001이 Phase 8에서 소비할 `.moai/decisions/skill-rename-map.yaml` 산출물을 Wave 1.4 landing commit에 포함한다 (spec.md §9.1.1 공유 계약).
- **5차 목표**: REQ-WF001-009/012/015/016 각각에 대한 CI fixture / dependency invariant 검증을 Wave 1.7 에 포함한다 (plan-audit 2026-04-25 응답).

---

## 1.5 Decision Log (v1.1.0 audit response)

본 문서 v1.1.0 은 `.moai/reports/plan-audit/SPEC-V3R2-WF-001-2026-04-25.md` 의 FAIL 판정에 대한 응답이다. 아래 결정은 명시적으로 기록되며, Stage 2 SPEC 에 핸드오프된다.

### DL-1: 48 → 24 scope resolution — Option B (target = 38, Stage 2 reserved)

- **선택지**: (A) 14개 추가 RETIRE 를 위한 R4 audit 재평가, (B) REQ-001 의 목표를 38 로 정직하게 수정하고 Stage 2 SPEC 예약, (C) 본 SPEC 을 "stage 1 of 2" 로 명시.
- **선택**: B + C hybrid. REQ-001 을 **38** 로 수정하고, 제목 및 HISTORY 에 Stage 1 이 48→38 임을 명시. 38→24 는 `SPEC-V3R3-WF-001` 로 예약 (아직 draft 미작성).
- **근거**: Option A (14 추가 RETIRE) 는 R4 audit 의 monitor/UNCLEAR/REFACTOR 판정을 재개하는 스코프 확장이며 본 SPEC 의 단일 단위 실행 범위를 넘어선다. 본 SPEC 은 §6.2 판정표가 실제로 달성 가능한 것(38)만 집행하고, 추가 감축은 telemetry 근거 확보 후 별도 SPEC 으로 다룬다.
- **영향**:
  - spec.md REQ-WF001-001 `exactly 24` → `exactly 38`
  - acceptance.md AC-WF001-01 `count = 24` → `count = 38`
  - plan.md §8 DoD `정확히 24` → `정확히 38`
  - tasks.md T1.7-1 assertion `24` → `38`
  - 새 SPEC 예약: `SPEC-V3R3-WF-001 — Skill Consolidation Stage 2 (38 → 24)`

### DL-2: §6.1 vs §6.2 single SoT — §6.2 wins

- §6.1 은 논리적 그룹핑으로 demotion. 디렉터리 수 SoT 는 §6.2 단일.
- 관련: **OQ-3 CLOSED** (§Decision Log DL-2).

### DL-3: Dual-verdict rows — single verdict enforced

- `moai-foundation-context` (row 5): R4 column = RETIRE, v3 action = RETIRE. 단일 verdict.
- `moai-workflow-jit-docs` (row 17): R4 column = RETIRE, v3 action = RETIRE. 단일 verdict.
- 관련: **OQ-1 CLOSED**, **OQ-7 CLOSED**.

### DL-4: moai-design-tools split resolution — Pencil → pencil-integration

- Pencil portion → `moai-workflow-pencil-integration` (authoritative; more granular target per audit preference).
- Figma portion → `.moai/archive/skills/v3.0/moai-design-tools/figma/` (archive, no substitute, Stage 2 재평가).
- tasks.md T1.4-8 `Substitute` 필드 는 `moai-workflow-pencil-integration (Pencil) + archive (Figma)` 로 업데이트. 이전의 `moai-design-system (Pencil)` 는 SUPERSEDED.
- 관련: **OQ-4 CLOSED**.

### DL-5: MIG-001 shared contract — dependency edge added

- spec.md §3 Environment 의 "REFERENCES (not depends on)" 문구를 제거하고 §9.1.1 Shared Contract 로 승격.
- plan.md §2.5 스키마는 WF-001 과 MIG-001 이 co-signed 하는 계약. 스키마 변경 시 HUMAN GATE (MIG-001 author review) 필수.
- REQ-WF001-009 의 behavioral contract 는 이제 정당한 의존성 엣지로 지지됨.

### DL-6: CI fixtures for REQ-015 / REQ-016

- AC-WF001-08 (REQ-015): 기존 dry-run 에 더해 `.moai/specs/SPEC-V3R2-WF-001/fixtures/ci-reject/` 의 broken fixture 를 추가하고 T1.7-10 이 이를 검증 (fixture-verifier 가 non-zero exit).
- AC-WF001-18 (REQ-016 신설): `.moai/specs/SPEC-V3R2-WF-001/fixtures/trigger-drop/` 의 broken fixture 를 T1.7-11 이 검증.
- 실제 production CI guard rail 구현은 여전히 follow-up SPEC `SPEC-CI-SKILL-GUARD-001` (proposed) 로 deferred. 본 SPEC 은 fixture 검증으로 REQ-015/016 을 observable 하게 만든다.
- 관련: **OQ-6 PARTIALLY CLOSED** (fixture 까지는 본 SPEC 범위; production CI 구현은 follow-up).

### DL-7: OQ-5 (REFACTOR scope) — deferred

- 6개 REFACTOR skill 의 실제 refactor scope 는 본 SPEC 범위 밖. 본 SPEC 은 `## Refactor Notes` 라벨만 주입하며 실제 refactor 는 6개 독립 sub-SPEC (`SPEC-V3R2-WF-001a`...`f`) 으로 분할 제안.
- Follow-up SPEC ID 블록: `SPEC-V3R2-WF-001a` (workflow-testing), `b` (domain-backend), `c` (domain-frontend), `d` (domain-database scope), `e` (platform-deployment), `f` (platform-auth).
- 관련: **OQ-5 DEFERRED** with explicit follow-up SPEC reservations.

### DL-8: OQ-CONTRACT (new) — MIG-001 schema human gate

- 본 SPEC 이 authoring 하는 `skill-rename-map.yaml` 스키마는 MIG-001 author 의 명시적 승인 후 merge 된다.
- 승인 artifact: GitHub PR review + MIG-001 spec.md 에 "consumes skill-rename-map.yaml schema v1" 문구 포함 확인.
- T1.7-9 가 schema 를 parse 하여 MIG-001 spec.md 에 "schema v1" 문자열 존재 여부를 grep 으로 검증.

---

## 2. 기술 접근 (Technical Approach)

### 2.1 Wave 기반 순차 실행 (Execute-in-Waves)

13개 삭제 + 4개 merge target 재구성 + 14개 REFACTOR + 12개 unchanged KEEP를 단일 commit으로 처리하면 diff가 실질적으로 리뷰 불가능하다. 다음 7개 Wave로 분할하여 각 Wave 끝에서 `make build` + `diff -rq` + 기본 테스트를 수행한다:

| Wave | 명칭 | 목적 | 변경 디렉터리 수 | Revert 단위 |
|---|---|---|---|---|
| 1.1 | **KEEP baseline lock** | 12 KEEP + `moai` + `moai-ref-*` 5 + agency 2 = 20 frozen skills 의 baseline hash 기록 | 0 (read-only) | N/A |
| 1.2 | **MERGE target 확장** | `moai-foundation-thinking`, `moai-workflow-project`, `moai-domain-database`, `moai-design-system`(신설) 에 흡수 skill 내용을 추가 (소스는 아직 유지) | 4 targets | 단일 revert |
| 1.3 | **Trigger keyword union dedup** | 4 merge target의 frontmatter triggers / related-skills 병합 및 중복 제거 | 4 targets (frontmatter only) | 단일 revert |
| 1.4 | **RETIRE archive + skill-rename-map** | 13 absorbed skill을 `.moai/archive/skills/v3.0/<name>/` 에 이동, `RETIRED.md` 작성, `skill-rename-map.yaml` 생성 | −13 dirs, +13 archive entries | 단일 revert |
| 1.5 | **REFACTOR 라벨 주입** | 14개 REFACTOR skill의 SKILL.md에 `## Refactor Notes` 섹션 추가, 2개 UNCLEAR skill에 `## Telemetry Window` 섹션 추가 | 14 + 2 skills | 단일 revert |
| 1.6 | **Agent prompt rewrite** | 4개 agent 파일의 retired skill 이름 치환 (grep-based) | 4 agent files | 단일 revert |
| 1.7 | **CI 검증 + 최종 `make build`** | `diff -rq`, CI guard rails (SKILL_RETIRE_NO_ARCHIVE, SKILL_TRIGGER_DROP) dry-run, skill count assertion | 테스트만 | N/A |

### 2.2 Template + Local 동시쓰기 정책

[HARD] CLAUDE.local.md §2 Template-First Rule 준수. 모든 Wave에서:

1. **먼저** `internal/template/templates/.claude/skills/…` 에 변경을 적용한다.
2. **그 다음** 동일 변경을 `.claude/skills/…` 에 적용한다 (또는 `make build` 후 재배포).
3. Wave 마지막에 `diff -rq .claude/skills internal/template/templates/.claude/skills` 가 empty output인지 확인한다.
4. Wave 1.4 에서 archive 이동 시에는 양쪽 모두에서 동일 경로로 이동한다 (`.moai/archive/skills/v3.0/` 은 리포 루트 하나만 존재).

[WARN] `make build` 는 `internal/template/embedded.go` 를 재생성하므로 Wave 1.2, 1.3, 1.4, 1.5 각 끝에서 실행한다. Wave 1.6 (agent files, not skills) 는 template에도 agent 파일이 있으므로 함께 수정하고 `make build` 재생성.

### 2.3 삭제 순서: Merge target 설정 → Archive → Reference rewrite

REQ-WF001-017 (retired skill 이름 auto-rewrite) 을 깨지 않기 위해 **반드시 다음 순서**를 지킨다:

```
Wave 1.2  merge target 확장 (source 아직 존재, 참조 유효)
  ↓
Wave 1.3  trigger union dedup (source 아직 존재)
  ↓
Wave 1.4  source archive 이동 → skill-rename-map.yaml 생성
  ↓
Wave 1.6  agent prompt rewrite (skill-rename-map 기반)
```

Wave 1.4 전에 agent prompt를 고치면 retired 이름이 코드에 남은 상태로 source skill이 아직 살아 있어 ambiguity가 생기고, Wave 1.4 후에 agent rewrite를 미루면 REQ-WF001-014 위반. 따라서 **merge → archive → rewrite** 순서는 엄격하다.

### 2.4 Rollback 전략

각 Wave는 단일 git commit (또는 topic commit series)로 구성되며 다음 revert 규칙을 따른다:

- **Wave 1.2 revert**: `git revert <wave-1.2-sha>` — merge target에 추가된 content만 제거. 소스 skill은 이미 살아 있으므로 기능 복구.
- **Wave 1.3 revert**: 1.3 단독 revert 가능 (frontmatter만 바뀐 atomic commit).
- **Wave 1.4 revert**: `git revert <wave-1.4-sha>` 후 `make build` 재실행. archive → skills 이동이 자동 복구됨. **[WARN] Wave 1.4 revert 시 Wave 1.5, 1.6 도 함께 revert 해야 한다** (skill-rename-map 의존).
- **Wave 1.5, 1.6 revert**: 각 단독 가능 (subset change).
- **Wave 1.7**: 테스트 wave이므로 revert 불요.

전체 revert 순서: `1.7 → 1.6 → 1.5 → 1.4 → 1.3 → 1.2`. 역순으로 되돌리면 v2.13.2 상태로 정확히 복귀.

### 2.5 `.moai/decisions/skill-rename-map.yaml` 스키마

Wave 1.4에서 생성할 artifact (MIG-001 이 Phase 8에서 소비):

```yaml
# .moai/decisions/skill-rename-map.yaml
version: 1
generated_by: SPEC-V3R2-WF-001
generated_at: 2026-04-24  # 실제 Wave 1.4 commit date
merges:
  - from: moai-foundation-philosopher
    to: moai-foundation-thinking
    verdict: MERGED
  - from: moai-workflow-thinking
    to: moai-foundation-thinking
    verdict: MERGED
  - from: moai-workflow-templates
    to: moai-workflow-project
    verdict: MERGED
  - from: moai-docs-generation
    to: moai-workflow-project
    verdict: MERGED
  - from: moai-workflow-jit-docs
    to: moai-workflow-project
    verdict: MERGED
  - from: moai-design-craft
    to: moai-design-system
    verdict: MERGED
  - from: moai-domain-uiux
    to: moai-design-system
    verdict: MERGED
  - from: moai-design-tools
    to: moai-design-system  # Pencil portion → moai-workflow-pencil-integration also (split)
    verdict: SPLIT
    split_targets:
      - moai-design-system
      - moai-workflow-pencil-integration
  - from: moai-platform-database-cloud
    to: moai-domain-database
    verdict: MERGED
  - from: moai-foundation-context
    to: moai-foundation-core
    verdict: ABSORBED
retires:
  - name: moai-tool-svg
    substitute: null  # niche, no substitute
    archive: .moai/archive/skills/v3.0/moai-tool-svg/
refactors:
  - moai-workflow-testing
  - moai-domain-backend
  - moai-domain-frontend
  - moai-platform-deployment
  - moai-platform-auth
  - # ... (14 total — see §3.5)
unchanged_keep:
  - moai
  - moai-foundation-core
  - moai-foundation-cc
  - moai-foundation-quality
  - moai-workflow-spec
  - moai-workflow-tdd
  - moai-workflow-ddd
  - moai-workflow-worktree
  - moai-workflow-loop
  - moai-workflow-gan-loop
  - moai-workflow-design-import
  - moai-workflow-design-context
  - moai-workflow-research
  - moai-workflow-pencil-integration
  - moai-domain-copywriting       # FROZEN
  - moai-domain-brand-design      # FROZEN
  - moai-domain-db-docs
  - moai-tool-ast-grep
  - moai-library-mermaid
  - moai-library-shadcn
  - moai-library-nextra
  - moai-formats-data              # monitor
  - moai-framework-electron        # UNCLEAR 60-day window
  - moai-platform-chrome-extension # UNCLEAR 60-day window
  - moai-ref-api-patterns
  - moai-ref-git-workflow
  - moai-ref-owasp-checklist
  - moai-ref-react-patterns
  - moai-ref-testing-pyramid
```

---

## 3. Wave 상세 설계

### Wave 1.1 — KEEP Baseline Lock (Read-Only)

**목적**: 변경 없이 baseline hash 기록. 후속 Wave에서 KEEP/FROZEN skill이 실수로 수정되지 않았음을 검증하기 위한 참조 상태 확보.

**수행 작업**:

1. `shasum -a 256 .claude/skills/moai-domain-copywriting/SKILL.md .claude/skills/moai-domain-brand-design/SKILL.md` 기록 (agency FROZEN 검증용).
2. `shasum -a 256 .claude/skills/moai-foundation-{core,cc,quality}/SKILL.md` 기록.
3. 12 KEEP + `moai` + 5 `moai-ref-*` + 2 agency = 20 skill 의 해시를 `.moai/specs/SPEC-V3R2-WF-001/baseline-hashes.txt` 에 보관 (임시 파일, Wave 1.7 끝에서 삭제).

**변경 파일**: 없음 (read-only Wave).

**검증**: baseline-hashes.txt 가 20 라인을 포함.

### Wave 1.2 — MERGE Target Content Absorption

**목적**: 4개 merge target skill 본문에 흡수되는 skill의 Level 2 content를 추가한다. 소스 skill은 아직 유지 → reference resolution 이 어느 시점에도 깨지지 않음.

**변경 파일** (template + local 쌍):

| # | Target skill | 흡수 대상 | 변경 부위 |
|---|---|---|---|
| T1 | `moai-foundation-thinking/SKILL.md` | `moai-foundation-philosopher` (First Principles, Five Whys) + `moai-workflow-thinking` (Sequential Thinking MCP) | SKILL.md 에 `## First Principles (absorbed)` + `## Sequential Thinking MCP (absorbed)` 섹션 삽입. 기존 3-framework 구조 유지. |
| T2 | `moai-workflow-project/SKILL.md` | `moai-workflow-templates` (template optimization) + `moai-docs-generation` (doc generation) + `moai-workflow-jit-docs` (JIT docs) | `## Template Optimization (absorbed from moai-workflow-templates)`, `## Documentation Generation (absorbed from moai-docs-generation)`, `## JIT Docs (absorbed from moai-workflow-jit-docs)` 섹션 추가. |
| T3 | `moai-design-system/SKILL.md` | **신규 디렉터리 생성**: `moai-design-craft` + `moai-domain-uiux` + `moai-design-tools` (Pencil portion) | 신규 skill 생성. 3개 소스 skill 본문을 통합하여 작성. frontmatter 에 `description`, `triggers`, `related-skills` 포함. |
| T4 | `moai-domain-database/SKILL.md` | `moai-platform-database-cloud` (cloud vendor section) | `## Cloud Vendor Guide (absorbed from moai-platform-database-cloud)` 섹션 추가. |
| T5 | `moai-foundation-core/SKILL.md` | `moai-foundation-context` content | `## Token Budget (absorbed from moai-foundation-context)` 섹션에 content 병합. |

**작업 순서**: T1 → T2 → T3 → T4 → T5 순차 (각 target 파일은 단일 소유).

**Bundled resource 이관** (REQ-WF001-010):

- `moai-design-craft/modules/*.md` → `moai-design-system/modules/` 로 복사 (원본은 Wave 1.4에서 archive).
- `moai-domain-uiux/modules/*.md`, `moai-domain-uiux/references/*.md` → `moai-design-system/modules/`, `moai-design-system/references/`.
- `moai-design-tools/reference/*.md` 중 Pencil 관련만 `moai-design-system/references/` 또는 `moai-workflow-pencil-integration/references/` 로 분배. Figma 부분은 Wave 1.4에서 archive.
- `moai-workflow-templates/modules/*.md`, `moai-workflow-templates/references/*.md` → `moai-workflow-project/modules/`, `moai-workflow-project/references/`.
- `moai-docs-generation/modules/*.md`, `moai-docs-generation/references/*.md` → `moai-workflow-project/modules/`, `moai-workflow-project/references/`.
- `moai-foundation-philosopher/modules/*.md`, `moai-foundation-philosopher/references/*.md` → `moai-foundation-thinking/modules/`, `moai-foundation-thinking/references/`.
- `moai-platform-database-cloud/reference/*.md` → `moai-domain-database/references/`.

[HARD] Level 2 token budget 5000 token ceiling (`.claude/rules/moai/development/skill-authoring.md`) 초과 금지. 초과 시 Level 3 (modules/)로 이관.

**Checkpoint**: Wave 1.2 끝에서 `make build && diff -rq .claude/skills internal/template/templates/.claude/skills` 가 empty.

### Wave 1.3 — Trigger Keyword Union + Related-Skills Dedup

**목적**: REQ-WF001-007 집행. 4개 merge target의 frontmatter 에 source skill의 `triggers:`, `related-skills:` union 을 병합하고 중복 제거.

**변경 파일** (frontmatter only):

| Target | 병합할 source frontmatter |
|---|---|
| `moai-foundation-thinking` | `moai-foundation-philosopher`, `moai-workflow-thinking` |
| `moai-workflow-project` | `moai-workflow-templates`, `moai-docs-generation`, `moai-workflow-jit-docs` |
| `moai-design-system` | `moai-design-craft`, `moai-domain-uiux`, `moai-design-tools` (Pencil portion only) |
| `moai-domain-database` | `moai-platform-database-cloud` |
| `moai-foundation-core` | `moai-foundation-context` |

**Dedup 규칙**:
- trigger 문자열 완전 일치 시 제거.
- case-insensitive 중복 제거 (예: "Database" vs "database").
- `related-skills:` 는 retired skill 이름을 그대로 유지(alias로 작동, REQ-WF001-003).

**검증**: Wave 1.3 끝에서 union 결과가 target SKILL.md 의 frontmatter 에만 있고 body 에는 중복 없음. `grep -c "^  - " .claude/skills/<target>/SKILL.md` 로 entry 수 확인.

**Checkpoint**: `make build` + `diff -rq` empty.

### Wave 1.4 — RETIRE Archive + skill-rename-map.yaml

**목적**: REQ-WF001-008 집행. 13개 source skill 을 `.moai/archive/skills/v3.0/<name>/` 로 이동하고 `RETIRED.md` 작성. `skill-rename-map.yaml` 생성하여 MIG-001 에 전달.

**Archive 대상 13개 skill** (§6.2 verdict table 에서 "RETIRE (merged)" + "RETIRE (split)" + "RETIRE"):

| # | 삭제 skill | 대체 skill | Verdict |
|---|---|---|---|
| R1 | `moai-foundation-context` | `moai-foundation-core` | RETIRE (fold) |
| R2 | `moai-foundation-philosopher` | `moai-foundation-thinking` | RETIRE (merged) |
| R3 | `moai-workflow-thinking` | `moai-foundation-thinking` | RETIRE (merged) |
| R4 | `moai-workflow-templates` | `moai-workflow-project` | RETIRE (merged) |
| R5 | `moai-workflow-jit-docs` | `moai-workflow-project` | RETIRE (merged) |
| R6 | `moai-domain-uiux` | `moai-design-system` | RETIRE (merged) |
| R7 | `moai-design-craft` | `moai-design-system` | RETIRE (merged) |
| R8 | `moai-design-tools` | `moai-design-system` (Pencil → `moai-workflow-pencil-integration`) | RETIRE (split) |
| R9 | `moai-docs-generation` | `moai-workflow-project` | RETIRE (merged) |
| R10 | `moai-platform-database-cloud` | `moai-domain-database` | RETIRE (merged) |
| R11 | `moai-tool-svg` | (no substitute; niche) | RETIRE |
| — | — | — | — |

위는 11개. 검증: spec.md §6.2 의 verdict roll-up (v1.1.0 corrected) 는 "RETIRE (archived) = **11 directories**" 로 명시한다. 디렉터리 기준으로 **archive 대상은 총 11** (foundation-context 및 jit-docs 모두 위 R1/R5 항목에 포함).

**[CLOSED OQ-1]**: `moai-workflow-jit-docs` 는 v3 action = RETIRE(merged). DL-3 (v1.1.0): single verdict = RETIRE — jit-docs 는 T1.4-5 에서 archive 완료.

**RETIRED.md 템플릿** (각 archived skill 하위):

```markdown
# RETIRED: <skill-name>

**Date**: 2026-04-24 (Wave 1.4 of SPEC-V3R2-WF-001)
**Verdict**: <MERGED|RETIRE|SPLIT>
**Substitute**: <target-skill-name> (or: none)
**SPEC**: SPEC-V3R2-WF-001
**Migration**: Users should switch trigger keyword "<old-trigger>" to "<new-skill-name>"
**Rationale**: R4 audit §Merge clusters — <brief reason>

## Historical Content Location
Original SKILL.md and bundled resources are preserved under this directory.
If a substitute exists, functional content has been relocated to the substitute skill.

## Reference
- R4 audit: `.moai/design/v3-redesign/research/r4-skill-audit.md`
- Merge map: `.moai/decisions/skill-rename-map.yaml`
```

**Archive 작업 순서**:

1. `mkdir -p .moai/archive/skills/v3.0/`
2. 각 R1~R11 에 대해 `git mv .claude/skills/<name> .moai/archive/skills/v3.0/<name>` (template 쪽도 동일).
3. `RETIRED.md` 를 각 archive 디렉터리에 작성.
4. `.moai/decisions/skill-rename-map.yaml` 작성 (§2.5 스키마).
5. `make build` 실행하여 `internal/template/embedded.go` 업데이트.

**Template 쪽 처리**: `internal/template/templates/.claude/skills/<name>` 도 archive 로 이동. 단, archive 자체는 `.moai/archive/` 리포 루트 단일 경로 → template에는 archive 디렉터리를 복사하지 않는다. 즉 template 에서는 **삭제**하고, local 에서는 **archive 로 이동**한다. (archive 는 dev-only artifact; template 사용자에게는 배포되지 않음.)

**[CLOSED OQ-2]**: archive 는 dev-only artifact. template 에서는 **삭제**, local 에서는 **archive 로 이동**. `diff -rq .claude/skills internal/template/templates/.claude/skills` 는 archive 양쪽 제외이므로 여전히 empty. DL-2 (v1.1.0): archive asymmetry resolved — dev-only path (`.moai/archive/`) 는 template tree 에 포함하지 않음.

**Checkpoint (v1.1.0 — OQ-3 resolved per DL-2)**: Wave 1.4 끝에서:
- `.claude/skills/` 디렉터리 수 = 48 (baseline) − 11 (T1.4-1..11 archive) + 1 (moai-design-system 신설, Wave 1.2) = **38** ✓
- 이 값이 Stage 1 의 최종 목표이며 Wave 1.5, 1.6 에서는 디렉터리 수 변동 없음 (섹션 주입/agent rewrite 만).
- **[OQ-3 CLOSED]** §6.2 가 디렉터리 수의 단일 SoT 이며 §6.1 은 논리적 그룹핑이다 (spec.md §6.1 노트 + §Decision Log DL-2 참조).

### Wave 1.5 — REFACTOR Label Injection + UNCLEAR Telemetry Window

**목적**: REQ-WF001-011 (REFACTOR 라벨) + REQ-WF001-013 (UNCLEAR 60-day window) 집행.

**REFACTOR label 주입 대상 (v1.1.0 최종)** — archive 되지 않고 surviving 하는 REFACTOR verdict skill **6개**:

1. `moai-workflow-testing` (§6.2 line 238)
2. `moai-domain-backend` (§6.2 line 249)
3. `moai-domain-frontend` (§6.2 line 250)
4. `moai-domain-database` (§6.2 line 251) — Wave 1.2 에서 MERGE target 이자 §6.2 verdict 는 REFACTOR → 두 라벨 모두 기록
5. `moai-platform-deployment` (§6.2 line 259)
6. `moai-platform-auth` (§6.2 line 260)

보조: `moai-foundation-cc` 는 directory naming unify 만 수행 (Refactor Notes 섹션 불필요).

Archive 된 REFACTOR 항목 (`moai-tool-svg`, `moai-docs-generation`, `moai-design-tools`, `moai-platform-database-cloud`) 은 Wave 1.4 에서 이미 처리됨 (각 `RETIRED.md` 에 verdict 기록). 이전 v0.1.0 의 [SKIP] 리스트 노이즈는 v1.1.0 에서 제거.

각 skill SKILL.md 하단 (Advanced Implementation 이후, Works Well With 이전)에 다음 섹션 추가:

```markdown
## Refactor Notes

**R4 audit verdict** (2026-04-23): REFACTOR — <audit reason>
**SPEC**: SPEC-V3R2-WF-001 §6.2 line <NN>
**Refactor scope** (deferred to future sub-SPEC):
- <specific refactor target 1>
- <specific refactor target 2>

This skill is retained in v3.0 but its body will be restructured in a follow-up SPEC.
```

**UNCLEAR 2개 skill**:

1. `moai-framework-electron`
2. `moai-platform-chrome-extension`

각 skill SKILL.md 하단에 다음 섹션 추가:

```markdown
## Telemetry Window

**Status**: UNCLEAR (60-day window)
**R4 audit verdict**: KEEP (monitor)
**SPEC**: SPEC-V3R2-WF-001 §6.2 (REQ-WF001-013)
**Window start**: 2026-04-24 (Wave 1.5 commit date)
**Window end**: 2026-06-23 (60 days)
**Re-audit trigger**: SessionStart hook activation count for this skill
**Decision criteria**:
- If activation count ≥ 5 during window → retain permanently
- If activation count = 0 during window → schedule RETIRE in v3.1
- If 0 < count < 5 → retain with "low-use" tag
```

**Checkpoint**: `make build` + `diff -rq` empty. REFACTOR 섹션이 6개 skill 에 추가되었고 Telemetry 섹션이 2개 UNCLEAR skill 에 추가됨.

### Wave 1.6 — Agent Prompt Rewrite

**목적**: REQ-WF001-014 집행. Retired skill 이름이 하드코딩된 agent 파일 4개를 `skill-rename-map.yaml` 기반으로 치환한다.

**대상 파일** (사전 grep 확인):

1. `/Users/goos/MoAI/moai-adk-go/.claude/agents/moai/expert-frontend.md`
2. `/Users/goos/MoAI/moai-adk-go/.claude/agents/moai/manager-project.md`
3. `/Users/goos/MoAI/moai-adk-go/.claude/agents/moai/builder-skill.md`
4. `/Users/goos/MoAI/moai-adk-go/.claude/agents/moai/manager-docs.md`

그리고 template 쌍:

5. `/Users/goos/MoAI/moai-adk-go/internal/template/templates/.claude/agents/moai/expert-frontend.md`
6. `/Users/goos/MoAI/moai-adk-go/internal/template/templates/.claude/agents/moai/manager-project.md`
7. `/Users/goos/MoAI/moai-adk-go/internal/template/templates/.claude/agents/moai/builder-skill.md`
8. `/Users/goos/MoAI/moai-adk-go/internal/template/templates/.claude/agents/moai/manager-docs.md`

**치환 규칙** (grep-driven, Edit tool):

| Old name | New name |
|---|---|
| `moai-foundation-philosopher` | `moai-foundation-thinking` |
| `moai-workflow-thinking` | `moai-foundation-thinking` |
| `moai-workflow-templates` | `moai-workflow-project` |
| `moai-docs-generation` | `moai-workflow-project` |
| `moai-workflow-jit-docs` | `moai-workflow-project` |
| `moai-design-craft` | `moai-design-system` |
| `moai-domain-uiux` | `moai-design-system` |
| `moai-design-tools` (Pencil context) | `moai-design-system` |
| `moai-design-tools` (Figma context) | (remove or archive reference) |
| `moai-platform-database-cloud` | `moai-domain-database` |
| `moai-foundation-context` | `moai-foundation-core` |
| `moai-tool-svg` | (remove reference; no substitute) |

[HARD] `Edit replace_all` 사용 시 각 문자열이 정확히 해당 매핑만 해당해야 함. 특히 `moai-design-tools` 는 문맥에 따라 다른 대체를 요구하므로 **`replace_all` 금지**; 문맥 별 수동 Edit.

[HARD] `moai-domain-brand-design` 과 `moai-domain-copywriting` 참조는 **변경하지 말 것** (FROZEN, REQ-WF001-005).

**Checkpoint**: Wave 1.6 끝에서:
```bash
grep -rl "moai-foundation-philosopher\|moai-workflow-thinking\|moai-design-craft\|moai-design-tools\|moai-domain-uiux\|moai-platform-database-cloud\|moai-workflow-templates\|moai-docs-generation\|moai-workflow-jit-docs\|moai-tool-svg\|moai-foundation-context" /Users/goos/MoAI/moai-adk-go/.claude/agents/ /Users/goos/MoAI/moai-adk-go/internal/template/templates/.claude/agents/
```
→ empty output. `make build` 후 재확인.

### Wave 1.7 — CI Verification + Final make build

**목적**: REQ-WF001-001, 002, 006, 015, 016 을 전수 검증.

**검증 체크리스트**:

1. `[ "$(ls -d .claude/skills/*/ | wc -l)" -eq 38 ]` (REQ-WF001-001; Stage 1 target per DL-1)
2. `[ "$(ls -d internal/template/templates/.claude/skills/*/ | wc -l)" -eq 38 ]`
3. `diff -rq .claude/skills internal/template/templates/.claude/skills` → empty (REQ-WF001-006)
4. `[ -d .moai/archive/skills/v3.0/ ]` 및 각 R1~R11 아카이브 존재 (REQ-WF001-015 dry-run)
5. `[ -f .moai/decisions/skill-rename-map.yaml ]`
6. FROZEN skill byte-compare: `shasum -a 256` 의 2개 agency skill 해시가 Wave 1.1 baseline 과 일치
7. `grep -r "moai-foundation-philosopher\|moai-workflow-thinking" .claude/agents/ | wc -l` == 0
8. `go test ./...` 통과 (기본 테스트 suite)
9. `go vet ./...` 통과
10. `make build` 성공 (embedded.go 재생성)

**추가 guard rail (옵션)**:

- `SKILL_RETIRE_NO_ARCHIVE` CI check: archive 디렉터리 없이 삭제된 skill 탐지 (pre-commit hook 제안).
- `SKILL_TRIGGER_DROP` CI check: merge target 의 trigger union 이 source 의 모든 trigger 를 포함하는지 검증.

본 SPEC 은 이 2개 CI check 의 실제 구현은 SPEC-V3R2-WF-002 또는 별도 sub-SPEC 으로 미룬다. 본 Wave 에서는 **manual dry-run** 만 수행하고 결과를 `.moai/specs/SPEC-V3R2-WF-001/wave-1.7-report.md` 에 기록.

**최종 정리**:

- `.moai/specs/SPEC-V3R2-WF-001/baseline-hashes.txt` (Wave 1.1 산출) 삭제.
- `.moai/specs/SPEC-V3R2-WF-001/wave-1.7-report.md` 생성.

---

## 4. 파일 영향 요약

### 4.1 신규 생성

- `.claude/skills/moai-design-system/` (신규 디렉터리 + SKILL.md + modules/ + references/)
- `internal/template/templates/.claude/skills/moai-design-system/` (동일 내용)
- `.moai/archive/skills/v3.0/<13 names>/RETIRED.md` (13개 파일)
- `.moai/archive/skills/v3.0/<11 names>/SKILL.md` (이관된 원본 — 11개 + 2개 Wave 1.4 해석에 따라)
- `.moai/decisions/skill-rename-map.yaml`

### 4.2 수정

- `.claude/skills/moai-foundation-thinking/SKILL.md` (+ modules/, references/ 추가)
- `.claude/skills/moai-workflow-project/SKILL.md` (+ modules/, references/ 확장)
- `.claude/skills/moai-domain-database/SKILL.md` (+ cloud 섹션)
- `.claude/skills/moai-foundation-core/SKILL.md` (+ context 섹션)
- `.claude/skills/moai-workflow-testing/SKILL.md` (+ Refactor Notes)
- `.claude/skills/moai-domain-backend/SKILL.md` (+ Refactor Notes)
- `.claude/skills/moai-domain-frontend/SKILL.md` (+ Refactor Notes)
- `.claude/skills/moai-platform-deployment/SKILL.md` (+ Refactor Notes)
- `.claude/skills/moai-platform-auth/SKILL.md` (+ Refactor Notes)
- `.claude/skills/moai-framework-electron/SKILL.md` (+ Telemetry Window)
- `.claude/skills/moai-platform-chrome-extension/SKILL.md` (+ Telemetry Window)
- 위 모든 local 파일의 `internal/template/templates/` 쌍
- `.claude/agents/moai/{expert-frontend,manager-project,builder-skill,manager-docs}.md` + template 쌍 (4 pairs)
- `internal/template/embedded.go` (auto-generated by `make build`)

### 4.3 삭제 (archive 로 이동)

- `.claude/skills/moai-foundation-context/`
- `.claude/skills/moai-foundation-philosopher/`
- `.claude/skills/moai-workflow-thinking/`
- `.claude/skills/moai-workflow-templates/`
- `.claude/skills/moai-workflow-jit-docs/`
- `.claude/skills/moai-docs-generation/`
- `.claude/skills/moai-domain-uiux/`
- `.claude/skills/moai-design-craft/`
- `.claude/skills/moai-design-tools/`
- `.claude/skills/moai-platform-database-cloud/`
- `.claude/skills/moai-tool-svg/`
- (동일 template 쌍 11개)

### 4.4 영향받지 않음 (FROZEN + 순수 KEEP)

- `.claude/skills/moai-domain-copywriting/` (FROZEN — agency)
- `.claude/skills/moai-domain-brand-design/` (FROZEN — agency)
- `.claude/skills/moai/` (SPEC-V3R2-WF-002 소관)
- `.claude/skills/moai-ref-*` (5개, 순수 KEEP)

---

## 5. 의존성 / 사전 조건

- Git 작업 트리 깨끗 (uncommitted change 없음)
- `make build` 정상 동작 (Go 1.26+, embedded.go 생성 가능)
- `shasum`, `diff`, `grep` 표준 Unix 도구 가용
- SPEC-V3R2-MIG-001 은 Phase 8 에서 `skill-rename-map.yaml` 을 소비. 본 SPEC 은 MIG-001 보다 먼저 머지되나 MIG-001 이 완성되지 않아도 WF-001 의 수치 집행은 독립적으로 완료 가능.

### 5.1 Blocks / Blocked By (spec.md §9)

- **Blocked by**: SPEC-V3R2-MIG-001 (user 로컬 migrator). 단 본 SPEC 의 리포 내 집행은 MIG-001 없이도 가능. MIG-001 은 user 영향을 관리.
- **Blocks**: SPEC-V3R2-WF-002 (moai root skill workflows/*.md 축소 — WF-001 이 moai skill KEEP 판정을 유지해야 실행 가능).
- **Related**: SPEC-V3R2-WF-005 (language rules vs skills 경계).

---

## 6. 리스크 / 완화 (상세)

| # | 리스크 | 영향 | 완화 |
|---|---|---|---|
| R1 | Wave 1.2 에서 merge target Level 2 토큰 초과 | Progressive Disclosure 위반 | `wc -w` + 토큰 환산 추정치 (4 chars ≈ 1 token) 로 사전 계산; 초과 시 Level 3 (modules/) 로 이관 |
| R2 | Wave 1.4 에서 template/local archive 비대칭 (OQ-2) | diff -rq 실패 | template 에서는 삭제, local 에서는 archive 로 이동; archive 는 dev-only 경로로 명시 |
| R3 | Wave 1.6 의 `replace_all` 오류 | 잘못된 skill 이름 치환 | `moai-design-tools` 문맥별 수동 Edit, 전체 grep 재검증 |
| R4 | Agency FROZEN skill 의도치 않은 수정 | GAN loop 계약 파괴 | Wave 1.1 baseline hash 와 Wave 1.7 재검증 hash 비교 (exact match 요구) |
| R5 | `moai-workflow-jit-docs` 의 KEEP vs RETIRE 모호 **(OQ-1 CLOSED)** | 집계 오류 → 38개 초과/미달 | DL-3 (v1.1.0): single verdict = RETIRE; T1.4-5 에서 archive 완료. 리스크 해소됨. |
| R6 | `skill-rename-map.yaml` 스키마가 MIG-001 과 불일치 | Phase 8 에서 MIG-001 실패 | 본 SPEC §2.5 스키마를 artifact 로 확정하되 MIG-001 spec.md 가 제정되면 재정합 |
| R7 | 13개 삭제 중 하나라도 bundled resource 누락 | 구현 지식 손실 | Wave 1.2 에서 bundled resource 이관을 명시적으로 수행, Wave 1.4 archive 전 재검증 |
| R8 | `make build` 실패 | embedded.go 비정상 | 각 Wave 끝에서 `make build` 를 수행하여 빠른 실패 detection |
| R9 | 6개 REFACTOR skill 의 실제 리팩터 작업이 본 SPEC 범위 밖 | 품질 부채 이월 | 본 SPEC 은 **라벨만** 주입; 실제 리팩터는 별도 sub-SPEC(WF-001a, b 등) 제안 |

---

## 7. OPEN QUESTIONS (v1.1.0 audit response)

| OQ | Status (v1.1.0) | Resolution | Closing deliverable |
|---|---|---|---|
| OQ-1 (jit-docs verdict) | **CLOSED** | DL-3: single verdict = RETIRE (both R4 and v3 action columns) | spec.md §6.2 row 17 edited to RETIRE in R4 column |
| OQ-2 (archive asymmetry) | **CLOSED** | Archive is dev-only; `diff -rq` excludes `.moai/archive/`. Template gets deletion, local gets `git mv`. | tasks.md T1.4-1..11 explicit archive-only-local step |
| OQ-3 (§6.1 vs §6.2) | **CLOSED** | DL-2: §6.2 is single SoT; §6.1 demoted to logical grouping | spec.md §6.1 note prefix + §6.2 corrected roll-up |
| OQ-4 (design-tools split) | **CLOSED** | DL-4: Pencil → `moai-workflow-pencil-integration`; Figma → archive/figma subdir | spec.md §6.2 row 31 updated; tasks.md T1.4-8 Substitute corrected |
| OQ-5 (REFACTOR scope) | **DEFERRED** | DL-7: label-only in this SPEC; 6 sub-SPECs reserved (SPEC-V3R2-WF-001a..f) | Follow-up SPEC IDs registered in DL-7 |
| OQ-6 (CI guard rails) | **PARTIALLY CLOSED** | DL-6: broken-fixture verification in T1.7-10/11 (REQ-015/016). Production CI implementation deferred to `SPEC-CI-SKILL-GUARD-001` | Fixtures at `.moai/specs/SPEC-V3R2-WF-001/fixtures/{ci-reject,trigger-drop}/` + T1.7-10/11 tasks |
| OQ-7 (foundation-context verdict) | **CLOSED** | DL-3: single verdict = RETIRE | spec.md §6.2 row 5 edited to RETIRE in R4 column |
| OQ-CONTRACT (new) | **HUMAN GATE** | DL-8: MIG-001 author must explicitly approve `skill-rename-map.yaml` schema v1 before Wave 1.4 landing commit | PR review trail + MIG-001 spec.md cross-reference check in T1.7-9 |

All M1-blocking OQs (1, 3, 4, 7) are CLOSED. OQ-2 was non-blocking but is also CLOSED. OQ-5 and OQ-6 are documented deferrals, not blockers. OQ-CONTRACT is a HUMAN GATE explicitly gated before Wave 1.4 merge.

---

## 8. Milestones (우선순위 기반, 시간 추정 금지)

| 우선순위 | Milestone | Objective unblock 조건 (v1.1.0) |
|---|---|---|
| **Priority High** | M1: Audit response gates closed | 다음 **모두** 만족: (1) OQ-1 CLOSED (spec.md §6.2 row 17 R4 column = RETIRE), (2) OQ-3 CLOSED (spec.md §6.1 prefix note + §6.2 corrected roll-up present), (3) OQ-4 CLOSED (spec.md §6.2 row 31 Pencil destination = `moai-workflow-pencil-integration` + tasks.md T1.4-8 updated), (4) OQ-7 CLOSED (spec.md §6.2 row 5 R4 column = RETIRE), (5) OQ-CONTRACT HUMAN GATE scheduled (MIG-001 author review item created). OQ-2/5/6 는 비블로킹 deferral 로 명시적 기록 확인. |
| **Priority High** | M2: Wave 1.1 (Baseline hash lock) | M1 완료 + baseline-hashes.txt 생성 |
| **Priority High** | M3: Wave 1.2 (MERGE target 확장 + 신규 design-system 생성) | M2 완료, Level 2 토큰 budget 사전 계산 (4 merge targets 각 `wc -w * 0.75 ≤ 5000`) |
| **Priority High** | M4: Wave 1.3 (Trigger union dedup) | M3 완료 + 4 target frontmatter YAML parse 성공 |
| **Priority High** | M5: Wave 1.4 (Archive + skill-rename-map) | M4 완료 + `.moai/archive/skills/v3.0/` 생성 + OQ-CONTRACT HUMAN GATE 완료 (MIG-001 author approval on PR) |
| **Priority Medium** | M6: Wave 1.5 (REFACTOR + Telemetry 라벨) | M5 완료 |
| **Priority Medium** | M7: Wave 1.6 (Agent prompt rewrite) | M5 완료 (M6 병렬 가능) |
| **Priority High** | M8: Wave 1.7 (CI verification + final make build + fixtures) | M6, M7 완료 + fixtures 디렉터리 2개 생성 (`ci-reject`, `trigger-drop`) |
| **Priority Low** | M9: `baseline-hashes.txt` 삭제, `wave-1.7-report.md` 작성, PR 생성 | M8 통과 |

Milestone 순서: M1 → M2 → M3 → M4 → M5 → {M6, M7 병렬} → M8 → M9.

**M1 unblock objectivity note**: v1.1.0 에서 M1 의 수행 조건을 **파일-레벨 검증 가능한** 체크리스트로 재정의. Audit 검토자가 `grep` 기반으로 5개 closing deliverable 을 확인 가능해야 Wave 1.2 착수.

---

## 9. Definition of Done (DoD)

본 SPEC (Stage 1) 은 다음 기준을 모두 만족할 때 완료된다:

1. `.claude/skills/` 디렉터리 수 정확히 **38** (REQ-WF001-001, Stage 1 target).
2. `internal/template/templates/.claude/skills/` 디렉터리 수 정확히 **38**.
3. `diff -rq .claude/skills internal/template/templates/.claude/skills` 가 empty (REQ-WF001-006).
4. §6.2 판정표 48 entry 전부에 대해 **단일 verdict 라벨** 보장 (REQ-WF001-002) — T1.7-8 verdict-uniqueness validator 통과.
5. `.moai/archive/skills/v3.0/` 에 **11개** archive 디렉터리 존재, 각 디렉터리에 `RETIRED.md` 존재 (REQ-WF001-008, 015).
6. `.moai/decisions/skill-rename-map.yaml` 존재 및 §2.5 스키마 준수 (spec.md §9.1.1 shared contract + OQ-CONTRACT HUMAN GATE 승인 로그).
7. Agency FROZEN skill 2개 (`moai-domain-copywriting`, `moai-domain-brand-design`) 의 SKILL.md 파일 해시가 Wave 1.1 baseline 과 정확히 일치 (REQ-WF001-005).
8. Retired skill 이름이 `.claude/agents/` 및 template 쌍에 0 occurrence (REQ-WF001-014 grep check).
9. `make build` 성공 (exit code 0 명시 검증), `internal/template/embedded.go` 재생성 완료.
10. `go test ./...` 및 `go vet ./...` 통과.
11. Wave 1.7 report (`.moai/specs/SPEC-V3R2-WF-001/wave-1.7-report.md`) 작성 완료.
12. 본 SPEC 의 OPEN QUESTIONS 상태가 §7 표와 동일 (OQ-1/3/4/7 CLOSED; OQ-2 CLOSED; OQ-5 DEFERRED; OQ-6 PARTIALLY CLOSED; OQ-CONTRACT HUMAN GATE 완료).
13. `acceptance.md` 의 **18 AC** 전부 PASS (v1.1.0 에서 AC-16/17/18 추가됨).
14. MIG-001 shared contract 검증 (T1.7-9): `skill-rename-map.yaml` schema v1 과 MIG-001 spec.md 의 `schema v1` 참조 일치.
15. CI fixtures (`ci-reject`, `trigger-drop`) 존재 및 T1.7-10/11 fixture-verifier 가 non-zero exit 로 예상 거부 거동 증명.

---

**계획자 메모**: 본 계획은 단일 PR 이 아닌 **7 Wave 의 commit chain** 으로 구현되어야 하며 각 Wave 는 독립 revert 가능하도록 설계했다. Wave 1.2 시작 전 반드시 M1 (OPEN QUESTIONS 해소) 이 완료되어야 Wave 실행이 안전하다.
