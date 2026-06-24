# plan.md — SPEC-V3R6-DOCS-V3-README-001

> Implementation plan for README v3 rewrite. Tier M (standard).
> 본 문서는 spec.md의 파생 산출물이며, 모든 사실 단언은 spec.md §B.2의 1차 소스 재검증 결과에 기반한다.

---

## §A. Context

본 SPEC은 Sprint 14 Docs-v3 코호트의 2/5 SPEC. 직전 CODEMAPS-V3-001 (origin/main 4a6f4b4d3 종료)이 작성한 `.moai/project/codemaps/docs-truth.md`를 canonical facts checklist로 삼아, repo-root의 `README.md` (1370L) + `README.ko.md` (1418L) 두 파일의 stale claim을 정정한다.

**Hard scope**: README 2개 파일만. docs-site 4-locale / Go 코드 / CLAUDE.md는 전부 EXCLUDE.

**사전 조건 (전수 1차 소스 재검증 완료 at commit 4a6f4b4d3)**:
- §1: `ls .claude/agents/moai/*.md` → 7 MoAI-custom files ✓
- §2: `internal/spec/status.go` `ValidStatuses` → 8 lowercase values ✓
- §3: `internal/spec/lint.go` required slice (lines 590-601) → 12 entries ✓
- §4.2: `ls -1 .claude/commands/moai/*.md | wc -l` → 17 ✓
- §5: `internal/config/defaults.go` lines 42-57 → `DefaultGLMHigh = "glm-5.2[1m]"` ✓

모든 1차 소스가 docs-truth 기준선과 일치 → blocker 없음.

---

## §B. Known Issues / Debt

- **B1**: README.md L292 "Manager: 8"이 `spec, ddd, tdd, docs, quality, project, strategy, git` 나열 — 이 중 `manager-strategy`/`manager-quality`/`manager-project`는 archived, `ddd`/`tdd`는 cycle_type이지 agent가 아님. 전체 rewrite 필요.
- **B2**: README.md L293 "Expert: 8"이 `backend, frontend, security, devops, performance, debug, testing, refactoring` 나열 — 6종 `expert-*` 전부 archived. 카테고리 자체 제거.
- **B3**: README.md L302 "47 Skills" 헤더 — `/moai` command set (17)과 skill catalog (47)가 혼동. 본 SPEC은 `/moai` 17-command 명시 + "47 Skills" 헤더 제거만 수행. 정확한 skill 총수 산출은 별도 SPEC.
- **B4**: README.ko.md L40 "52개 스킬" vs L110 "47개 스킬" — ko 내부 모순. en/ko 동기화 시一并 정정.
- **B5**: README.md L682 `GLM-5.1` → `glm-5.2[1m]` 정정 시 pricing 표의 input/output 단가도 z.ai 현재 pricing과 정합 필요 (본 SPEC은 모델명 정정만 수행, 단가는 z.ai pricing page 참조 권장이지만 본 SPEC 범위 밖).

---

## §C. Pre-flight Checks

run-phase 진입 전 manager-develop이 확인할 항목:

- [ ] `git status` clean (본 SPEC 작업 전)
- [ ] `git log --oneline -1` → 4a6f4b4d3 (CODEMAPS-V3-001 close) 또는 그 이후 커밋
- [ ] `.moai/project/codemaps/docs-truth.md` 존재 (122L)
- [ ] 5개 1차 소스 파일 존재 및 위 재검증 결과와 일치:
  - `internal/spec/status.go` (ValidStatuses 8 values)
  - `internal/spec/lint.go` (required slice 12 entries)
  - `internal/config/defaults.go` (DefaultGLM* block)
  - `.claude/agents/moai/*.md` (7 files)
  - `.claude/commands/moai/*.md` (17 files)
- [ ] `README.md` (1370L) + `README.ko.md` (1418L) 존재

---

## §D. Constraints

- **D1 (scope)**: README.md + README.ko.md 2개 파일만 수정. docs-site / `.go` / CLAUDE.md / `internal/template/templates/`는 금지.
- **D2 (doc-only, no template scope)**: 본 SPEC은 documentation-only이며 Go change LOC = 0. Repo-root `README.md` / `README.ko.md`는 project-owned asset이지 template-distributed asset이 아님 (`find internal/template/templates -iname "README*"` → 0 results). 따라서 template-neutrality CI guard scope는 본 SPEC과 무관.
- **D3 (en/ko sync)**: 양쪽 파일의 사실값 (count, 이름, 모델명)은 정확히 일치. 번역 품질은 ko가 en 의미 반영.
- **D4 (preservation)**: statusline v3 + preset retire 섹션은 이미 반영됨 — 본 SPEC 작업 중 회귀 금지.
- **D5 (anti-overengineering)**: 추상화/설정 시스템/미래 확장 hook 추가 금지. 사실 정정(reconciliation)만.
- **D6 (no invented facts)**: 모든 신규 claim은 docs-truth §1-§5의 1차 소스로 소급 가능해야 함.

---

## §E. Self-Verification (run-phase E1-E7)

manager-develop의 run-phase §E self-verification 매트릭스 (본 SPEC은 documentation-only이므로 E2 build / E3 coverage는 N/A):

- **E1 (AC PASS/FAIL)**: acceptance.md의 8 AC 전수 기계적 검증 (grep 기반)
- **E2 (cross-platform build)**: N/A — Go 코드 변경 없음
- **E3 (coverage)**: N/A — Go 코드 변경 없음
- **E4 (subagent-boundary grep)**: N/A — 본 SPEC은 orchestrator-direct 또는 manager-develop 단독 수행
- **E5 (lint)**: `moai spec lint SPEC-V3R6-DOCS-V3-README-001` → 0 findings
- **E6 (push state)**: commit + push 성공
- **E7 (neutrality)**: `internal/template/internal_content_leak_test.go` PASS (template 변경 없으므로 no-op)

---

## §F. Milestones (Tier M, M1..M6)

> 파일-단위로 잘린 마일스톤. 각 M은 단일 PR 커밋 단위.

### §F.1 M1 — Agent catalog 섹션 + tier-mapping table rewrite (REQ-README-001, en 우선)

**Scope**: `README.md` L288-300 "Agent Categories" 섹션 + **L335-360 "Agent Model Assignment by Tier" tier-mapping table** (D2 정정 — 둘 다 본 마일스톤에서 처리).

**작업**:
1. "27 agents" → "8 retained agents" (7 MoAI-custom + 1 `Explore`)
2. Manager 카테고리: 4 agents (manager-spec, manager-develop, manager-docs, manager-git)
3. Evaluator 카테고리: 2 agents (plan-auditor, sync-auditor)
4. Builder 카테고리: 1 agent (builder-harness)
5. Anthropic built-in: 1 (`Explore`)
6. Expert / Design System 행 제거 (archived 또는 skill 카테고리)
7. archived agent migration 참조 추가 (`.claude/rules/moai/workflow/archived-agent-rejection.md`)
8. **L335-360 tier-mapping table (D2)**: 10 archived active 행 (`manager-strategy`, `manager-project`, `manager-quality`, `expert-backend`, `expert-frontend`, `expert-security`, `expert-debug`, `expert-refactoring`, `expert-devops`, `expert-performance`) 전부 제거. Retained 8종에 대한 tier-mapping 행만 유지. Archived 이름은 최대 1줄 "see `archived-agent-rejection.md` migration table" 참조로 대체 (active tier row 유지 금지).

**Commit**: `docs(SPEC-V3R6-DOCS-V3-README-001): M1 agent catalog section + tier-mapping table rewrite (en)`

### §F.2 M2 — GLM tier-model 정정 (REQ-README-003, en)

**Scope**: `README.md` L670-686 GLM 섹션.

**작업**:
1. L676 "GLM-5.1, GLM-4.7, GLM-4.5-Air" → "glm-5.2[1m], glm-4.7, glm-4.5-air"
2. L682 Opus-tier `GLM-5.1` → `glm-5.2[1m]`
3. `[1m]` suffix 설명 부가 (1M context mode 활성화, Claude Code가 파싱 후 strip)
4. L683 Sonnet `glm-4.7` 유지, L684 Haiku `glm-4.5-air` case 표준화

**Commit**: `docs(SPEC-V3R6-DOCS-V3-README-001): M2 GLM tier-model 정정 glm-5.2[1m] (en)`

### §F.3 M3 — `/moai` command set 17 정정 + "47 Skills" 헤더 제거 (REQ-README-002, en)

**Scope**: `README.md` L302-321 "47 Skills (Progressive Disclosure)" 섹션.

**작업**:
1. 헤더 "### 47 Skills (Progressive Disclosure)" → "### `/moai` Slash Commands (17)" 또는 동등
2. 17-command 명시: brain/clean/codemaps/coverage/design/e2e/feedback/fix/gate/harness/loop/mx/plan/project/review/run/sync
3. **13-row skill-category count 표 disposition (D4 정정)**: 기존 L306-320의 13-row skill-category count 표 (Foundation 6 / Workflow 12 / Domain 4 / Format 1 / Platform 4 / Library 3 / Reference 5 / Tool 2 / Design 2 / Framework 1 / Legacy 5 / Docs 1 / Language Rules 16)는 **stale 상태로 방치하지 않고 간결한 17-command `/moai` 라스팅으로 치환**한다 (plan-auditor decision iii). Wholesale 삭제만 하고 대체물을 두지 않는 것은 금지 — 사용자가 `/moai` command set을 한눈에 볼 수 있는 대체 라스팅이 항상 뒤따라야 한다. skill catalog 정확한 총수 산출은 별도 skill-audit SPEC scope (본 SPEC 범위 밖).
4. Progressive Disclosure 3-level 시스템 설명은 보존 (skill loading 메커니즘은 여전히 유효)

**Commit**: `docs(SPEC-V3R6-DOCS-V3-README-001): M3 /moai 17-command set 정정 + 47 Skills 헤더 제거 (en)`

### §F.4 M4 — README.ko.md 동기화 (REQ-README-001/002/003, ko)

**Scope**: `README.ko.md` 대응 섹션 전체.

**작업**:
1. L334 "에이전트 카테고리" 표 → M1의 en 구조와 동일 (8 retained)
2. **5개 stale agent-count surface 전수 정정 (D1)**:
   - L40 "24개 전문 AI 에이전트" → "8 retained 에이전트"
   - L110 "26개 전문 AI 에이전트 + 47개 스킬" → "8 retained 에이전트 + 17 `/moai` 명령"
   - L308 "24개 전문 에이전트에게 작업을 위임" (`AI` 누락 variant) → "8 retained 에이전트에게 작업을 위임"
   - L372 "24개 에이전트에 최적의 AI 모델을 할당" (또 다른 suffix variant) → "8 retained 에이전트에 최적의 AI 모델을 할당"
   - L334 카테고리 표 자체 (위 1번 작업)
3. **L380-410 "티어별 에이전트 모델 할당" tier-mapping table (D2)**: en L335-360과 동일 정정 — archived 11종 (`expert-testing` 포함) active 행 제거, retained 8종 행만 유지, `archived-agent-rejection.md` 1줄 참조
4. **L348 "47개 스킬" 헤더 + L350-368 13-row count 표 (D3)**: en M3과 동일 — `### /moai` Slash Commands (17)` 헤더로 교체, 13-row count 표는 17-command `/moai` 라스팅으로 치환
5. L722/L728 GLM 섹션 → M2의 en 값과 동일 (`glm-5.2[1m]` Opus)
6. ko 번역 품질: en의 의미를 정확히 반영

**Commit**: `docs(SPEC-V3R6-DOCS-V3-README-001): M4 README.ko.md 동기화 (agent + GLM + command)`

### §F.5 M5 — statusline 보존 검증 + scope boundary 확인 (REQ-README-005/006)

**Scope**: en/ko 양쪽 statusline 섹션 보존 상태 확인 + `.go` 파일 미수정 확인.

**작업**:
1. README.md L1282 `preset` retire 문구 보존 확인 (grep)
2. README.ko.md L1351 동일 보존 확인
3. statusline v3 multi-line layout 설명 보존 확인
4. `git diff --stat`에 `.go` 파일 부재 확인
5. `git diff --stat`에 `docs-site/`, `CLAUDE.md`, `internal/template/templates/` 부재 확인

**Commit**: `docs(SPEC-V3R6-DOCS-V3-README-001): M5 statusline 보존 검증 + scope boundary 확인`

### §F.6 M6 — en/ko cross-check + 최종 AC 기계적 검증 (REQ-README-004)

**Scope**: 양쪽 파일 사실값 일치 diff + acceptance.md 8 AC 전수 grep 검증.

**작업**:
1. en/ko agent count 일치 (둘 다 8 retained)
2. en/ko GLM Opus model 일치 (둘 다 `glm-5.2[1m]`)
3. en/ko `/moai` command count 일치 (둘 다 17)
4. en/ko "47 Skills" / "52개 스킬" / "26개" 부재 (grep)
5. acceptance.md AC-1 ~ AC-8 전수 grep 검증
6. `moai spec lint SPEC-V3R6-DOCS-V3-README-001` → 0 findings

**Commit**: `docs(SPEC-V3R6-DOCS-V3-README-001): M6 en/ko cross-check + 최종 AC 기계적 검증`

---

## §G. Anti-Patterns (금지 패턴)

- **AP-1**: skill catalog 정확한 총수 산출하여 README에 추가 — 본 SPEC 범위 밖 (별도 skill-audit SPEC).
- **AP-2**: docs-site 4-locale content 동기화 — DOCSITE-001 scope.
- **AP-3**: Go 코드 변경 ( lint rule / status enum / GLM default 등) — 본 SPEC은 doc-only. 코드가 사실이면 코드를 따름 (코드 수정 금지).
- **AP-4": archived agent 12종 이름을 README에 명시적 나열 — `archived-agent-rejection.md` 참조로 처리.
- **AP-5**: statusline preset retire 섹션을 실수로 rollback — M5 preservation check로 방지.
- **AP-6**: en만 수정하고 ko를 방치 — M4가 ko 동기화 전담.
- **AP-7": "future-proofing" 목적의 추상화/설정 시스템 추가 — 사실 정정만.

---

## §H. Cross-References

- `spec.md` — 본 SPEC의 canonical requirements (REQ-README-001 ~ 006)
- `acceptance.md` — 8 AC의 Given-When-Then 시나리오
- `.moai/project/codemaps/docs-truth.md` — canonical facts checklist
- `.moai/specs/SPEC-V3R6-DOCS-CODEMAPS-V3-001/` — 선행 SPEC
- `.claude/rules/moai/workflow/archived-agent-rejection.md` — archived agent migration table
- `.claude/rules/moai/development/spec-frontmatter-schema.md` — frontmatter schema SSOT

---

## §I. Risks

| Risk | Probability | Impact | Mitigation |
|------|-------------|--------|------------|
| en/ko 사실값 미세 불일치 | Med | Low | M6 cross-check grep 검증 |
| statusline 섹션 실수 rollback | Low | Med | M5 preservation check |
| docs-site 또는 `.go` 실수 수정 | Low | High | M5 scope boundary `git diff --stat` 확인 |
| GLM pricing 표 단가까지 정정 시도 | Med | Low | 본 SPEC은 모델명만, 단가는 z.ai pricing page 참조 (scope 밖 명시) |
| skill catalog 47 숫자를 지우면서 관련 유효 설명까지 삭제 | Med | Med | M3는 "47 Skills" 헤더 + count 표만, Progressive Disclosure 메커니즘 설명은 보존 |

---

## HISTORY

- 2026-06-17: plan-phase artifacts authored. Tier M. 6 milestones (M1 agent catalog en / M2 GLM en / M3 command set en / M4 ko 동기화 / M5 statusline 보존 + scope / M6 en/ko cross-check + AC 검증).
- 2026-06-17 (iter-2, v0.2.0): plan-auditor PASS-WITH-DEBT 0.83 → D1/D2/D3/D4 + neutrality over-specification 정정. M1에 en L335-360 tier-mapping table scope 추가 (D2), M3에 13-row count 표 disposition 명시 (D4), M4에 ko L308/L372 5 stale surface + L380-410 tier table + L348/L350-368 "47개 스킬" 헤더 정정 추가 (D1/D2/D3). §D2 neutrality over-specification 경량화 (repo-root README는 project-owned not template-distributed).
