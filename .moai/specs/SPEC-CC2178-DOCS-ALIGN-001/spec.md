---
id: SPEC-CC2178-DOCS-ALIGN-001
title: "CC 2.1.169→2.1.178 Tier 1 docs-only 기능 — 규칙·docs-site 4-locale 정합화"
version: "0.1.0"
status: implemented
created: 2026-06-16
updated: 2026-06-17
author: GOOS
priority: P2
phase: "v3.0.0"
module: "docs-site,.claude/rules/moai"
lifecycle: spec-anchored
tags: "docs,i18n,cc2178,alignment"
tier: S
---

# SPEC-CC2178-DOCS-ALIGN-001 — CC 2.1.169→2.1.178 Tier 1 docs-only 정합화

## HISTORY

| Version | Date | Author | Description |
|---------|------|--------|-------------|
| 0.1.0 | 2026-06-16 | GOOS | 최초 plan-phase 작성. `.moai/research/cc-update-2.1.163-to-2.1.178.md` P4 + P5(9개 Tier 1 docs-only 항목)를 SPEC-CC2178-MODEL-POLICY-REPAIR-001 sibling로 분리. Go 코드 0, docs-site 4-locale + `.claude/rules/moai/` markdown에 한정. |

## §A — Overview (배경)

CC 2.1.169..2.1.178 창 분석(`.moai/research/cc-update-2.1.163-to-2.1.178.md`)에서 Tier 1(critical) 분류된 11개 항목 중 **9개가 docs-only**로 확인되었다. 그중 2개(`availableModels` / `enforceAvailableModels` allowlist 거버넌스)는 sibling SPEC-CC2178-MODEL-POLICY-REPAIR-001의 비용 레버 영역으로 이미 흡수되었다. 본 SPEC은 **잔여 9개 Tier 1 docs-only 항목**을 다룬다 — `Tool(param:value)` wildcard 권한 규칙, nested `.claude/` closest-wins 선행 규칙, `disableBundledSkills` 토글, `--safe-mode` 플래그, `post-session` 라이프사이클 훅, nested skills 로딩 선행 규칙, auto-mode pre-launch classifier, subagent `disallowedTools` MCP 강제, 그리고 `/cd` 프롬프트 캐시 보존 재개.

이 항목들은 전부 **CC 기능의 존재와 옵션을 문서화**하는 것이지, MoAI가 이를 구현한다고 주장하지 않는다. MoAI는 `post-session` 훅·`disableBundledSkills` 등 일부를 현재 wiring하지 않는다 — 정확성이 완전성보다 우선한다.

결함은 3개 마일스톤으로 그룹화된다:

| 마일스톤 | 도메인 | 항목 수 | 주요 대상 |
|---------|--------|--------|-----------|
| M1 | permissions + skills discovery | 4 | `settings-management.md`, `skill-authoring.md`, `agent-authoring.md`, docs-site `settings-json.md` / `skill-guide.md` / `agent-guide.md` |
| M2 | hooks + agent governance | 3 | `settings-management.md`, `hooks-system.md`, `agent-authoring.md`, docs-site `hooks-reference.md` / `hooks-guide.md` / `agent-guide.md` |
| M3 | session resume (`/cd`) | 1 | `session-handoff.md` Block 0, docs-site `statusline.md` / `moai-sync.md` |

총 8 REQ (1항목 = 1 REQ; 단 `--safe-mode`는 `disableBundledSkills`와 함께 REQ-DA-004로 통합 — CC 2.1.169 governance/debug 토글 2종이 같은 rules 절에 자연스럽게 배치되므로 9항목 → 8 REQ). 모든 REQ는 "관찰 가능한 본문 상태"를 기술하며 구현 세부(함수명/스키마)를 포함하지 않는다.

## §B — File Ownership Map

대상 파일은 두 부류다. 각 경로는 plan-phase에서 Read/Glob/`wc -l`로 존재가 검증되었다.

### B.1 `.claude/rules/moai/` (template-managed: source + mirror 이중 편집)

`.claude/rules/moai/` 하위 파일은 `internal/template/templates/<같은 경로>`에 source가 있고 배포본이 mirror다. 편집 시 source 우선 → `make build` → mirror 동기화 → `internal/template/internal_content_leak_test.go` + `template-neutrality-check.yaml` 통과의 규율을 따른다(CLAUDE.local.md §2/§25).

| 파일 | 줄 수(검증) | 편집 영역 |
|------|-------------|-----------|
| `.claude/rules/moai/core/settings-management.md` | 405 | § Permission Management (`Tool(param:value)` wildcard), § Claude Code Settings (`disableBundledSkills`, `--safe-mode`) |
| `.claude/rules/moai/development/skill-authoring.md` | 348 | § Skill Scope and Priority (nested `.claude/skills` 로딩, closest-wins, `disableBundledSkills`) |
| `.claude/rules/moai/development/agent-authoring.md` | 303 | § Frontmatter Format Rules (`disallowedTools` MCP 강제 선행 규칙), § managed-settings (nested closest-wins) |
| `.claude/rules/moai/core/hooks-system.md` | 384 | § Hook Events (`post-session` 라이프사이클 이벤트 추가) |
| `.claude/rules/moai/workflow/session-handoff.md` | 310 | § Worktree-Anchored Resume Pattern Block 0 (`/cd` 캐시 보존 대안 노트) |
| `.claude/rules/moai/workflow/orchestration-mode-selection.md` | (존재) | §B 결정 트리 (auto-mode pre-launch classifier 노트) |

### B.2 docs-site (4-locale 동시 편집 — ko/en/ja/zh)

`.moai/docs/docs-site-i18n-rules.md` §17.3 [HARD]에 따라, ko 변경 시 동일 PR 내에서 en/ja/zh 모두 반영이 의무다. 각 페이지는 4개 locale에 동일 경로로 존재한다(검증 완료).

| docs-site 페이지 (locale별 동일 경로) | 편집 내용 |
|----------------------------------------|-----------|
| `advanced/settings-json.md` | `Tool(param:value)` wildcard 권한 규칙, `disableBundledSkills`, `--safe-mode` |
| `advanced/hooks-reference.md` + `advanced/hooks-guide.md` | `post-session` 라이프사이클 훅 |
| `advanced/agent-guide.md` | subagent `disallowedTools` MCP 강제, nested closest-wins |
| `advanced/skill-guide.md` | nested `.claude/skills` 로딩, closest-wins, `disableBundledSkills` |
| `advanced/statusline.md` + `workflow-commands/moai-sync.md` | `/cd` 캐시 보존 재개 노트 (M3) |

### B.3 LOCAL-ONLY (template mirror 없음, 직접 편집)

없음. 본 SPEC은 LOCAL-ONLY 파일을 편집하지 않는다.

## §C — GEARS Requirements

> 표기 규약: GEARS notation. `<subject>`은 대상 규칙 파일·docs-site 페이지. 모든 REQ는 "관찰 가능한 본문 상태"를 기술하며 구현 세부(함수명·스키마)를 포함하지 않는다. Template-managed 규칙 파일 본문에는 내부 SPEC ID·REQ 토큰·날짜·commit SHA를 임베드하지 않는다(§25 neutrality). CC 기능 중 MoAI가 wiring하지 않는 항목(`post-session`, `disableBundledSkills`)은 "존재와 옵션"으로 문서화하며 "MoAI가 구현한다"고 주장하지 않는다.

### M1 — permissions + skills discovery (REQ-DA-001 ~ REQ-DA-004)

**REQ-DA-001** [high] — `Tool(param:value)` permission-rule syntax with `*` wildcard(CC 2.1.178)가 `settings-management.md` § Permission Management와 docs-site `advanced/settings-json.md`(4-locale)에 미문서화 상태다. 기존 `Tool(specifier)` 형식은 `settings-json.md` 라인 362에 존재하나 param-scoped wildcard는 부재한다. The `settings-management.md` 규칙 § Permission Management와 docs-site `advanced/settings-json.md`(ko/en/ja/zh 4-locale)은 `Tool(param:value)` wildcard 권한 규칙 문법을 예시와 함께 추가 **shall** 한다. 문서는 MoAI가 현재 param-scoped 규칙을 emit하지 않으나 "옵션으로 존재함"을 명시한다.

**REQ-DA-002** [high] — Nested `.claude/` directories에서 closest-wins on name collision(agent/workflow/output-style, CC 2.1.178) 선행 규칙이 `skill-authoring.md`와 `agent-authoring.md` 및 docs-site `advanced/skill-guide.md`/`advanced/agent-guide.md`(4-locale)에 미문서화 상태다. The `skill-authoring.md` § Skill Scope and Priority와 `agent-authoring.md`(해당 절) 및 docs-site 양 페이지(ko/en/ja/zh)는 nested `.claude/`에서 이름 충돌 시 closest-directory-wins 선행 규칙을 추가 **shall** 한다.

**REQ-DA-003** [med] — Nested `.claude/skills` 로딩 — 해당 디렉터리 작업 시 skills가 로드됨(CC 2.1.178)이 `skill-authoring.md` § Skill Scope and Priority와 docs-site `advanced/skill-guide.md`(4-locale)에 미문서화 상태다. The `skill-authoring.md` 규칙과 docs-site `advanced/skill-guide.md`(ko/en/ja/zh)는 nested `.claude/skills` 디렉터리 로딩 시맨틱을 발견 노트로 추가 **shall** 한다.

**REQ-DA-004** [med] — `disableBundledSkills` 설정 + 환경변수(CC 2.1.169, 번들 skills/workflows 숨김)가 `settings-management.md` § Claude Code Settings와 `skill-authoring.md` 및 docs-site `advanced/settings-json.md`/`advanced/skill-guide.md`(4-locale)에 미문서화 상태다. The `settings-management.md` 규칙과 `skill-authoring.md` 및 docs-site 양 페이지(ko/en/ja/zh)는 `disableBundledSkills` 토글을 추가 **shall** 한다. 문서는 번들 `/deep-research` 등을 비활성화함을 명시하며, MoAI가 이를 emit하지 않음을 옵션 노트로 기록한다.

### M2 — hooks + agent governance (REQ-DA-005 ~ REQ-DA-007)

**REQ-DA-005** [med] — `post-session` 라이프사이클 훅(CC 2.1.169, self-hosted runner, 세션 종료 후 발화)가 `hooks-system.md` § Hook Events와 docs-site `advanced/hooks-reference.md`/`advanced/hooks-guide.md`(4-locale)에 미문서화 상태다. The `hooks-system.md` 규칙 § Hook Events와 docs-site 양 페이지(ko/en/ja/zh)는 `post-session` 라이프사이클 이벤트를 추가 **shall** 한다. 문서는 MoAI가 현재 이 훅을 wiring하지 않음을 "존재와 옵션"으로 명시한다(accuracy over completeness).

**REQ-DA-006** [low] — Subagent `disallowedTools`의 MCP server-level specs 강제(CC 2.1.178, 이전 silent-ignore에서 강제로 변경)가 `agent-authoring.md` § Frontmatter Format Rules와 docs-site `advanced/agent-guide.md`(4-locale)에 미문서화 상태다. The `agent-authoring.md` 규칙과 docs-site `advanced/agent-guide.md`(ko/en/ja/zh)는 `disallowedTools`가 MCP 서버 수준에서 강제됨을 노트로 추가 **shall** 한다. run-phase는 MoAI agent가 이전 silent-ignore 동작에 의존하지 않았음을 grep으로 확인한다.

> **Non-normative prediction note (REQ-DA-006)**: run-phase grep 결과 "의존 없음"이 예상된다 (MoAI agent 정의는 이전 silent-ignore 동작에 의존하는 패턴을 포함하지 않음). 단, 이것은 예측이며 run-phase grep이 SSOT다 — grep이 의존을 발견하면 blocker report로 승격.

**REQ-DA-007** [low] — auto-mode pre-launch classifier(CC 2.1.178, subagent spawn이 classifier에 의해 launch 전 평가)가 `orchestration-mode-selection.md` §B 결정 트리와 docs-site(관련 페이지)에 미문서화 상태다. The `orchestration-mode-selection.md` 규칙은 auto-mode의 pre-launch classifier 게이트를 노트로 추가 **shall** 한다. 이는 `/goal` unattended 루프와의 상보성과 연결된다.

### M3 — session resume / `/cd` (REQ-DA-008)

**REQ-DA-008** [med] — `/cd` 명령(CC 2.1.169, 프롬프트 캐시 보존하며 cwd 이동)이 `session-handoff.md` § Worktree-Anchored Resume Pattern Block 0과 docs-site `advanced/statusline.md`/`workflow-commands/moai-sync.md`(4-locale)에 미문서화 상태다. `/cd`는 새 터미널(Block 0)의 cold-start 대신 캐시 보존 대안이다. The `session-handoff.md` 규칙 § Worktree-Anchored Resume Pattern과 docs-site 양 페이지(ko/en/ja/zh)는 `/cd`를 cache-preserving 재개 대안으로 노트 **shall** 한다. 이 노트는 기존 Block 0 new-terminal 경로를 대체하지 않고 보완한다.

## §D — Constraints (비기능 제약)

- **C1 (4-locale parity MANDATORY)**: `.moai/docs/docs-site-i18n-rules.md` §17.3 [HARD]에 따라, docs-site content 변경 시 ko/en/ja/zh 4개 locale에 동일 PR 내 반영이 의무다. ko-only 머지 금지. `scripts/docs-i18n-check.sh`(pre-commit + CI)가 4-locale 파일 수/경로 일치를 강제한다.
- **C2 (template-first mirror)**: B.1의 모든 template-managed 규칙 파일 편집은 source(`internal/template/templates/...`) 우선 → `make build` → mirror(`.claude/...`) 동기화 → neutrality 통과 순으로 진행한다(CLAUDE.local.md §2/§25).
- **C3 (§25 neutrality)**: template 규칙 본문 편집은 내부 SPEC ID / REQ 토큰 / 내부 날짜 / commit SHA / archive·memory 경로를 임베드하지 않는다.
- **C4 (no Go source)**: 본 SPEC은 어떤 `internal/` / `pkg/` / `cmd/` Go 소스도 수정하지 않는다(`make build`로 인한 `embedded` 재생성은 산출물이며 소스 변경 아님).
- **C5 (accuracy over completeness)**: MoAI가 wiring하지 않는 CC 기능(`post-session`, `disableBundledSkills`, `--safe-mode` 일부)은 "존재와 옵션"으로 문서화하며 "MoAI가 구현한다"고 주장하지 않는다.
- **C6 (no new rules file)**: 본 SPEC은 기존 규칙·문서의 in-place 정합화만 수행한다. 새 `.claude/rules/...` 파일을 만들지 않는다.
- **C7 (line-number drift)**: §B에 인용된 줄 수는 2026-06-16 plan-phase 검증 시점 값이다. run-phase는 grep 앵커(literal 문자열)로 위치를 재확인한 뒤 편집한다.
- **C8 (Mermaid TD-only + 강조/괄호 공백)**: docs-site 편집 시 `.moai/docs/docs-site-i18n-rules.md` §17.2 [HARD] — Mermaid는 TD/TB only, `**강조**(괄호)` 사이 공백 필수, 이모지 유입 금지.

## §E — Exclusions (What NOT to Build)

### Out of Scope — CC2178 docs-only 정합화

[HARD] 본 SPEC 범위에서 명시적으로 제외하는 항목:

- Go 코드 변경 (ZERO) — `internal/` / `pkg/` / `cmd/` / `internal/template/templates/` 하위 Go 소스(`*.go`)는 일절 수정하지 않는다. 본 SPEC은 순수 docs/rules markdown 정합화다. `make build`로 인한 `embedded.go` 재생성은 산출물이며 소스 변경으로 간주하지 않는다.
- `availableModels` / `enforceAvailableModels` allowlist 거버넌스 문서화 — CC 2.1.175/2.1.176 비용 레버. sibling SPEC-CC2178-MODEL-POLICY-REPAIR-001의 P2 영역으로 이미 흡수되었다. 본 SPEC에서 중복 다루지 않는다.
- `[1m]` model-policy constraint 재검증 — CC 2.1.172-2.1.174 1M/suffix 정규화 fix가 constraint의 실패 표면을 좁혔는지 검증. sibling MODEL-POLICY-REPAIR-001의 P3로 위임.
- Fable 5 tier 채택 — CC 2.1.170 신규 모델. deliberate cost-strategy 결정이며 본 SPEC(또는 MODEL-POLICY-REPAIR-001) 범위가 아니다. 별도 전략 SPEC 대기(P6 deferred).
- CC 공식 문서 자체 수정 — `code.claude.com/docs` / `claude.com/docs`는 외부 자산이며 본 SPEC 대상이 아니다. MoAI 측 규칙·docs-site만 정합화한다.
- CHANGELOG / README 갱신 — sync-phase(manager-docs) 책임. plan/run phase에서는 다루지 않는다.
- 새 규칙 파일·새 docs-site 페이지 신규 생성 — 기존 파일 in-place 정정만 수행(C6).
- 번들 26개 SKILL.md 전수 조사 — 본 SPEC은 Tier 1 docs-only 9개 항목에 한정. 기존 skills의 다른 정합 결함은 범위 밖(predecessor SPEC-CC-DOCS-ALIGNMENT-001이 33결함을 이미 처리).
- `--safe-mode` troubleshooting 페이지 신규 생성 — 기존 `settings-management.md` / `settings-json.md`에 노트 추가로 충분. 별도 troubleshooting 섹션 신설은 over-engineering.

## §F — Traceability

| REQ | 마일스톤 | 대상 rules 파일 | 대상 docs-site 페이지(4-locale) | AC |
|-----|---------|----------------|----------------------------------|-----|
| REQ-DA-001 | M1 | settings-management.md | advanced/settings-json.md | AC-DA-001 |
| REQ-DA-002 | M1 | skill-authoring.md + agent-authoring.md | advanced/skill-guide.md + advanced/agent-guide.md | AC-DA-002 |
| REQ-DA-003 | M1 | skill-authoring.md | advanced/skill-guide.md | AC-DA-003 |
| REQ-DA-004 | M1 | settings-management.md + skill-authoring.md | advanced/settings-json.md + advanced/skill-guide.md | AC-DA-004 |
| REQ-DA-005 | M2 | hooks-system.md | advanced/hooks-reference.md + advanced/hooks-guide.md | AC-DA-005 |
| REQ-DA-006 | M2 | agent-authoring.md | advanced/agent-guide.md | AC-DA-006 |
| REQ-DA-007 | M2 | orchestration-mode-selection.md | (rules-only; docs-site auto-mode는 이미 goal-directive.md에 연결) | AC-DA-007 |
| REQ-DA-008 | M3 | session-handoff.md | advanced/statusline.md + workflow-commands/moai-sync.md | AC-DA-008 |

**Cross-SPEC references**:
- Predecessor: `.moai/specs/SPEC-CC-DOCS-ALIGNMENT-001/` (33 docs/rules defects, completed — 본 SPEC은 그 후속 CC 창 정합화)
- Sibling: `.moai/specs/SPEC-CC2178-MODEL-POLICY-REPAIR-001/` (cost/quality model-policy, 본 SPEC과 같은 CC 2.1.163→2.1.178 창에서 분리)
- Source research: `.moai/research/cc-update-2.1.163-to-2.1.178.md` P4 + P5
