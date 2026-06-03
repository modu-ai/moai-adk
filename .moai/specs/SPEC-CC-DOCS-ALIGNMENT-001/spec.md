---
id: SPEC-CC-DOCS-ALIGNMENT-001
title: "Claude Code 공식 문서 대비 규칙·문서 정합화 (33 documentation/rules defects)"
version: "0.2.0"
status: implemented
created: 2026-06-03
updated: 2026-06-03
author: manager-spec
priority: P2
phase: "v3.0.0"
module: ".claude/rules/moai, internal/template/templates/.claude/rules/moai, CLAUDE.md, CLAUDE.local.md"
lifecycle: spec-anchored
tags: "docs-alignment, claude-code, rules, template-mirror, neutrality"
tier: M
---

# SPEC-CC-DOCS-ALIGNMENT-001 — Claude Code 공식 문서 대비 규칙·문서 정합화

## HISTORY

| Version | Date | Author | Description |
|---------|------|--------|-------------|
| 0.1.0 | 2026-06-03 | manager-spec | 최초 plan-phase 작성 — 5개 공식 Claude Code 문서(workflows / skills / hooks-guide / goal / sub-agents) 대비 감사에서 도출된 33개 documentation/rules 결함을 5개 마일스톤(M1~M5)으로 그룹화. Go-code 결함 1종(hook EventType + CoverageTable)은 sibling SPEC로 분리(본 SPEC 범위 밖). |

## §A — Overview (배경)

5개 공식 Claude Code 문서를 MoAI-ADK의 규칙(`.claude/rules/...`) 및 코드와 대조한 감사에서 **34개 정합 결함**이 파일 단위로 독립 검증되었다. 본 SPEC은 그중 **33개 documentation/rules 결함**을 다룬다. 유일한 Go-code 결함(hook EventType 상수 + CoverageTable)은 OUT OF SCOPE이며 별도 sibling SPEC(SPEC-HOOK-EVENT-REGISTRY-001)에서 처리한다 — 본 SPEC에는 어떤 Go 소스 변경도 포함하지 않는다.

결함은 도메인별로 5개 그룹(마일스톤)으로 정리된다:

| 마일스톤 | 도메인 | 결함 수 | 주요 대상 파일 |
|---------|--------|--------|----------------|
| M1 | workflows | 6 | `dynamic-workflows.md`, `orchestration-mode-selection.md` |
| M2 | skills | 8 | `skill-writing-craft.md`, `skill-authoring.md`, `coding-standards.md`, `CLAUDE.md`, skills reference mirror |
| M3 | hooks (doc only) | 5 | `hooks-system.md`, `agent-common-protocol.md`, `CLAUDE.local.md`, `internal/hook/CLAUDE.md` |
| M4 | goal | 6 | `goal-directive.md` |
| M5 | sub-agents | 8 | `agent-authoring.md`, `agent-hooks.md`, `archived-agent-rejection.md`, `model-policy.md` |

총 33 결함 → 33 REQ (1 REQ = 1 finding, 일부 응집 클러스터는 같은 REQ 내 다중 AC로 표현).

## §B — File Ownership Map (검증 완료)

대상 파일은 두 부류로 나뉜다. 각 경로는 plan-phase에서 Read/Glob으로 존재·라인 영역이 검증되었다.

### B.1 Template-managed (source + mirror 이중 편집 필수)

다음 파일은 `internal/template/templates/<같은 경로>`에 source가 있고 배포본 `.claude/...`이 mirror다. 편집 시 source 우선 → `make build` → mirror 동기화 → neutrality 통과의 4단 규율을 따른다.

| 파일 | source↔mirror 현재 상태 | neutrality 비고 |
|------|--------------------------|------------------|
| `.claude/rules/moai/workflow/dynamic-workflows.md` | byte-IDENTICAL | 중립 본문 |
| `.claude/rules/moai/workflow/orchestration-mode-selection.md` | **DIFFER** (mirror에 내부 REQ/Finding/CONST 토큰, source는 중립화) | §25 neutrality split — byte-parity 단언 금지 |
| `.claude/rules/moai/development/skill-writing-craft.md` | byte-IDENTICAL | 중립 본문 |
| `.claude/rules/moai/development/skill-authoring.md` | byte-IDENTICAL | 중립 본문 |
| `.claude/rules/moai/development/coding-standards.md` | byte-IDENTICAL | 중립 본문 |
| `.claude/rules/moai/core/hooks-system.md` | byte-IDENTICAL | 중립 본문 |
| `.claude/rules/moai/core/agent-common-protocol.md` | (편집 영역 중립) | 중립 본문 |
| `.claude/rules/moai/workflow/goal-directive.md` | byte-IDENTICAL | 중립 본문 |
| `.claude/rules/moai/development/agent-authoring.md` | **DIFFER** (mirror↔source 5-byte; neutrality split) | byte-parity 단언 금지 |
| `.claude/rules/moai/core/agent-hooks.md` | byte-IDENTICAL | 중립 본문 |
| `.claude/rules/moai/workflow/archived-agent-rejection.md` | (편집 영역 검증) | 중립 본문 |
| `.claude/rules/moai/development/model-policy.md` | byte-IDENTICAL | 중립 본문 |
| `CLAUDE.md` | **DIFFER** (mirror에 §19.1 등 dev-local 토큰, source 중립) | byte-parity 단언 금지 |
| `.claude/skills/moai-foundation-cc/reference/claude-code-skills-official.md` | byte-IDENTICAL | 중립 본문 |

### B.2 LOCAL-ONLY (직접 편집, mirror 없음)

다음 파일은 template 트리에 존재하지 않는다(Glob 확인). mirror 동기화·neutrality 검사 대상이 아니다.

| 파일 | template mirror 존재? |
|------|------------------------|
| `CLAUDE.local.md` | 없음 (LOCAL-ONLY 정책 — CLAUDE.local.md §2) |
| `internal/hook/CLAUDE.md` | 없음 (`internal/template/templates/internal/hook/CLAUDE.md` 부재 확인) |

## §C — GEARS Requirements

> 표기 규약: GEARS notation 사용. `<subject>`은 대상 규칙 파일/문서. 모든 REQ는 "관찰 가능한 본문 상태"를 기술하며 구현 세부(함수명/스키마)는 포함하지 않는다. Template-managed 파일에 대한 REQ는 본문 내용에 내부 SPEC ID/REQ 토큰/날짜를 임베드하지 않는다(§25 neutrality) — 즉 어떤 규칙 본문에도 "per SPEC-CC-DOCS-ALIGNMENT-001"을 쓰지 않는다.

### M1 — workflows (REQ-CDA-001 ~ REQ-CDA-006)

**REQ-CDA-001** [high] — `dynamic-workflows.md` § Disabling Workflows의 "the `workflow` keyword no longer triggers a run" 문장은 트리거 키워드가 v2.1.160에서 `ultracode`로 개명된 사실과 어긋난다. The `dynamic-workflows.md` 규칙은 § Disabling Workflows 본문에서 트리거 키워드를 `ultracode`로 갱신하고, `workflow`가 v2.1.160 이전 키워드였음(자연어 요청은 양쪽 모두에서 동작)을 명시 **shall** 한다.

**REQ-CDA-002** [high] — `orchestration-mode-selection.md` Mode 6 행은 트리거를 `workflow`로 라벨링하나 `dynamic-workflows.md`는 `ultracode`를 `/effort` 레벨로만 문서화한다. The `dynamic-workflows.md` 규칙(및 필요 시 `orchestration-mode-selection.md`)은 `ultracode`(또는 "use a workflow")가 **per-prompt 트리거**이며, 세션 전역 `/effort ultracode`와 구별됨을 문서화 **shall** 한다.

**REQ-CDA-003** [med] — `/workflows` TUI(list/watch/pause/resume/save; 키 바인딩 p/x/s/r)가 미문서화 상태다. The `dynamic-workflows.md` 규칙은 "Manage runs" 하위 절을 추가하여 `/workflows` TUI 동작과 키 바인딩을 기술 **shall** 한다.

**REQ-CDA-004** [med] — Saved-workflow `args` global input이 미문서화 상태다. The `dynamic-workflows.md` 규칙은 § Saved workflows 아래에 `args` global input 항목 한 줄을 추가 **shall** 한다.

**REQ-CDA-005** [low] — Plan/provider 가용성(유료 플랜 + API/Bedrock/Vertex/Foundry; Pro는 `/config` 경유)이 부재하다. The `dynamic-workflows.md` 규칙은 plan/provider 가용성 1줄 노트를 추가 **shall** 한다.

**REQ-CDA-006** [low] — per-run approval prompt + permission-mode 매트릭스(Default/accept-edits는 매 실행 prompt; Auto는 first-launch; Bypass/`-p`/SDK는 never)가 § How a Workflow Runs 영역에 부재하다. The `dynamic-workflows.md` 규칙은 GATE-2 framing과 연결되는 간단한 per-run approval 노트를 추가 **shall** 한다.

### M2 — skills (REQ-CDA-007 ~ REQ-CDA-014)

**REQ-CDA-007** [high] — `skill-writing-craft.md`는 frontmatter 필드 `type: skill`을 REQUIRED로 선언(라인 203/219/278/291/304)하나 공식 skills 스펙에는 `type` 필드가 없고 실제 26개 SKILL.md 중 0개가 사용한다. The `skill-writing-craft.md` 규칙은 schema 테이블 및 4개 example 블록 전부에서 `type` 필드를 제거 **shall** 한다.

**REQ-CDA-008** [high] — description 길이 한계가 상호 모순된다: `skill-writing-craft.md`(라인 218/320)는 "≤80 chars/single sentence", `skill-authoring.md`(라인 17/258)는 "max 1024" 및 "under 250 for menu". 공식 cap은 `description`+`when_to_use` 합산 **1,536 chars**다. The skills 규칙(`skill-writing-craft.md` + `skill-authoring.md`)은 1,536 hard cap으로 통일하고 80/250 heuristic을 제거하거나 완화 **shall** 한다.

**REQ-CDA-009** [med] — `skillOverrides` settings 키(4 상태: on/name-only/user-invocable-only/off)가 부재하다. The `skill-authoring.md` 규칙 § Skill Invocation Control은 `skillOverrides` 하위 절을 추가(plugin skills에는 영향 없음 명시) **shall** 한다.

**REQ-CDA-010** [med] — `name` 필드는 공식상 optional(디렉터리명으로 기본값)이나 MoAI는 required로 표기(`skill-authoring.md` 라인 16, `skill-writing-craft.md` 라인 217). The skills 규칙은 `name`을 optional로 재분류하고 `moai-{category}-{name}`을 권고로 유지 **shall** 한다.

**REQ-CDA-011** [med] — nested/monorepo auto-discovery + `--add-dir` skill-loading 예외가 `skill-authoring.md` § Skill Scope and Priority에 미문서화 상태다. The `skill-authoring.md` 규칙은 discovery 노트(parent-walk, nested, `--add-dir` vs `permissions.additionalDirectories`)를 추가 **shall** 한다.

**REQ-CDA-012** [low] — `/run` `/verify` `/run-skill-generator` 번들 skills(v2.1.145+)가 reference mirror `claude-code-skills-official.md`에 부재하다. The `claude-code-skills-official.md` 참조 미러는 해당 번들 skills를 포함하도록 갱신 **shall** 한다.

**REQ-CDA-013** [low] — "custom commands가 skills로 병합"(`.claude/commands/X.md` ≡ `.claude/skills/X/SKILL.md`, skill 우선)이 `coding-standards.md` § Thin Command Pattern에 반영되지 않았다. The `coding-standards.md` 규칙은 § Thin Command Pattern에 1줄 노트를 추가 **shall** 한다.

**REQ-CDA-014** [low] — `CLAUDE.md`(라인 427-434) + `skill-authoring.md`(라인 131-139) §13 Level 2 "loaded when triggers match"는 description-listing과 body-invocation을 혼동한다. The `CLAUDE.md` 및 `skill-authoring.md`의 Level 2 문구는 "description은 항상 listing" vs "body는 invocation 시 로드되어 유지"로 구분되도록 정제 **shall** 한다.

### M3 — hooks DOC (REQ-CDA-015 ~ REQ-CDA-019, NO Go source)

**REQ-CDA-015** [med] — `if` 필드 버전: `hooks-system.md` 라인 191은 `v2.1.84+`라 하나 공식은 **v2.1.85+**이며, 라인 170은 v2.1.85라 하여 파일 내부가 자기모순이다. The `hooks-system.md` 규칙은 `if` 필드 버전을 v2.1.85+로 정정하고 파일 내부를 일관화 **shall** 한다.

**REQ-CDA-016** [low] — Stop-hook block-cap(연속 8회 block 후 override; `CLAUDE_CODE_STOP_HOOK_BLOCK_CAP`)이 미문서화 상태다. The `hooks-system.md` 규칙 Stop 절은 block-cap 1줄을 추가 **shall** 한다.

**REQ-CDA-017** [low] — hook "default 5s"가 `CLAUDE.local.md` 라인 386 + `internal/hook/CLAUDE.md` 라인 7/16에서 플랫폼 기본값으로 제시되나 플랫폼 기본은 10분이고 5s는 MoAI 정책이다. The 두 LOCAL-ONLY 파일은 "MoAI policy default 5s (platform default 10 min)"으로 재기술 **shall** 한다.

**REQ-CDA-018** [low] — `exec form`(`"args": []`, shell-bypass) troubleshooting이 `hooks-system.md`에 미문서화 상태다. The `hooks-system.md` 규칙은 exec form 노트(+ `if [[ $- == *i* ]]` interactive-shell guard)를 추가 **shall** 한다.

**REQ-CDA-019** [low] — `agent-common-protocol.md`의 Stop 관련 서술은 Stop == sync-commit completion으로 가정하나, 공식 caveat은 Stop이 매 turn-end마다 발화하며(task 완료 시점만이 아님) user interrupt 시에는 발화하지 않는다. The `agent-common-protocol.md` 규칙은 Stop self-gate caveat을 추가 **shall** 한다.

### M4 — goal (REQ-CDA-020 ~ REQ-CDA-025)

**REQ-CDA-020** [med] — `goal-directive.md`(라인 17-21) 비교 테이블 2행이 native `/loop`(트리거=time interval)을 `/moai loop`(diagnostic-driven)으로 치환하여, 독자가 공식 `/loop`을 오독한다. The `goal-directive.md` 규칙은 `/moai loop` 비교를 유지하면서 native `/loop`(time-interval)을 별도 행/노트로 추가 **shall** 한다.

**REQ-CDA-021** [med] — auto mode(per-tool)가 `/goal`(per-turn)과 상보적임이 미언급 상태다. The `goal-directive.md` 규칙은 auto mode와 `/goal`이 unattended `ac_converge` loop로 짝을 이룸을 MoAI Integration 노트로 추가 **shall** 한다.

**REQ-CDA-022** [low] — non-interactive surfaces가 desktop app + Remote Control도 포함하나 `goal-directive.md` 라인 44는 `-p`만 나열한다. The `goal-directive.md` 규칙은 non-interactive surfaces에 desktop app + Remote Control을 추가 **shall** 한다.

**REQ-CDA-023** [low] — `◎ /goal active` 인디케이터 + per-turn evaluator reason + bare-`/goal` turns/tokens 표시가 미문서화 상태다. The `goal-directive.md` 규칙은 해당 표시 1문장을 추가 **shall** 한다.

**REQ-CDA-024** [low] — disable scope가 부정확하다(라인 11): `disableAllHooks`=any level, `allowManagedHooksOnly`=managed only, command가 이유 설명. The `goal-directive.md` 규칙은 per-flag scope를 추가 **shall** 한다.

**REQ-CDA-025** [low] — evaluator token cost(small fast model, negligible, 세션 provider에서 실행 — GLM 포함)가 미기재 상태다. The `goal-directive.md` 규칙은 evaluator cost 1줄 노트를 추가 **shall** 한다.

### M5 — sub-agents (REQ-CDA-026 ~ REQ-CDA-033)

**REQ-CDA-026** [high] — `agent-authoring.md` 라인 54는 `maxTurns`를 deprecated로 표기하고 존재하지 않는 `maxContextSize` 필드를 가리킨다. 공식은 `maxTurns`를 current로 나열하며 `maxContextSize`는 부재한다. The `agent-authoring.md` 규칙은 deprecation 노트 + `maxContextSize`를 제거하고 `maxTurns`를 current optional 필드로 복원 **shall** 한다.

**REQ-CDA-027** [med] — `SubagentStart` hook event가 미문서화 상태다(`agent-authoring.md` 라인 75는 Pre/PostToolUse/SubagentStop만; `agent-hooks.md`도 부재). The `agent-authoring.md` frontmatter hook-event 리스트 + `agent-hooks.md` 테이블은 `SubagentStart`(agent-type name으로 매칭)를 추가 **shall** 한다.

**REQ-CDA-028** [med] — fork subagents(`/fork`, `CLAUDE_CODE_FORK_SUBAGENT`, v2.1.117+)가 전혀 미문서화 상태다(agent-authoring.md / agent-patterns.md 양쪽 grep 0). The sub-agents 규칙(agent-authoring.md 또는 agent-patterns.md)은 fork 노트(experimental, opt-in env, parent context 상속, nest 불가)를 추가 **shall** 한다.

**REQ-CDA-029** [med] — `claude-code-guide`는 CURRENT 공식 built-in helper agent이나 `archived-agent-rejection.md`(라인 84/131) + `CLAUDE.md`(라인 127)는 archived로 취급하여 `ARCHIVED_AGENT_REJECTED`를 강제한다. The sub-agents 규칙은 MoAI-archived `claude-code-guide`(과거 MoAI custom 파일)와 동명의 공식 built-in을 구분하여, built-in은 별개의 유효 agent이며 rejection을 트리거하지 않음을 명시 **shall** 한다.

**REQ-CDA-030** [low] — `model` full-ID form(`claude-opus-4-8`)은 공식상 허용되나 MoAI는 prohibited(의도적)다. The `model-policy.md` 규칙은 제한을 유지하되 full-ID form이 "official-but-intentionally-disallowed"임을 명시([1m] entitlement 근거 인용)하여 향후 "moai is outdated" 오독을 방지 **shall** 한다.

**REQ-CDA-031** [low] — `agent-authoring.md`(라인 105/185) "Permission Modes"는 공식 enum(default/acceptEdits/auto/dontAsk/bypassPermissions/plan)에 없는 `delegate`를 포함한다. The `agent-authoring.md` 규칙은 `delegate`를 canonical Frontmatter Format Rules enum(라인 185)에서 빼내어 "MoAI experimental extension" 노트로 명확히 표기 **shall** 한다.

**REQ-CDA-032** [low] — `name` 필드 의미(filename 일치 불필요; hooks는 이를 `agent_type`으로 수신)가 `agent-authoring.md` 라인 48에 미포착 상태다. The `agent-authoring.md` 규칙은 두 공식 명확화(filename 불일치 허용 + hook `agent_type` 수신)를 추가 **shall** 한다.

**REQ-CDA-033** [low] — managed-settings scope(org-wide, priority 1, `.claude/agents/` override)가 `agent-authoring.md` 라인 11-25 영역에 부재하다. The `agent-authoring.md` 규칙은 precedence 1줄 노트를 추가 **shall** 한다.

## §D — Constraints (비기능 제약)

- **C1 (template-first mirror)**: B.1의 모든 template-managed 파일 편집은 source(`internal/template/templates/...`) 우선 → `make build` → mirror(`.claude/...`) 동기화 → `internal/template/internal_content_leak_test.go` + `template-neutrality-check.yaml` 통과 순으로 진행한다(CLAUDE.local.md §2/§25).
- **C2 (§25 neutrality)**: template 규칙 본문 편집은 내부 SPEC ID / REQ 토큰 / 내부 날짜 / commit SHA / archive·memory 경로를 임베드하지 않는다.
- **C3 (neutrality split 보존)**: orchestration-mode-selection.md / agent-authoring.md / CLAUDE.md는 현재 source↔mirror가 의도적으로 DIFFER한다(mirror에만 내부 토큰). 이 split을 보존하며, source는 중립 표현·mirror는 dev-local 표현을 각각 갱신한다. byte-parity를 강제하지 않는다.
- **C4 (no Go source)**: 본 SPEC은 어떤 `internal/` / `pkg/` / `cmd/` Go 소스도 수정하지 않는다(`make build`로 인한 `embedded` 재생성은 산출물이며 소스 변경 아님).
- **C5 (line-number drift)**: §C에 인용된 라인 번호는 2026-06-03 plan-phase 검증 시점 값이다. run-phase는 grep 앵커(literal 문자열)로 위치를 재확인한 뒤 편집한다.

## §E — Exclusions (What NOT to Build)

[HARD] 본 SPEC 범위에서 명시적으로 제외하는 항목:

1. **Go hook EventType 상수 + CoverageTable 결함** — 감사 34건 중 유일한 Go-code 결함. sibling SPEC SPEC-HOOK-EVENT-REGISTRY-001에서 처리. 본 SPEC은 `internal/hook/types.go` 등 어떤 Go 소스도 건드리지 않는다.
2. **공식 문서 자체 수정** — Anthropic Claude Code 공식 문서(code.claude.com/docs)는 외부 자산이며 본 SPEC 대상이 아니다. 본 SPEC은 MoAI 측 규칙·문서만 정합화한다.
3. **새 규칙 파일 신규 생성** — 본 SPEC은 기존 규칙·문서의 in-place 정정만 수행한다. 새 `.claude/rules/...` 파일을 만들지 않는다(fork 노트 REQ-CDA-028은 기존 agent-authoring.md 또는 agent-patterns.md에 추가).
4. **CHANGELOG / README / docs-site 갱신** — sync-phase(manager-docs) 책임. plan/run phase에서는 다루지 않는다.
5. **byte-parity 강제 정합화** — C3에 따라 의도적 neutrality split이 있는 3개 파일(orchestration-mode-selection.md / agent-authoring.md / CLAUDE.md)을 byte-identical로 만드는 작업은 제외한다(neutrality 위반 유발).
6. **frontmatter status 전이 자동화 / lint 룰 확장** — 본 SPEC은 본문 내용 정합화이며 lint 엔진 변경을 포함하지 않는다.
7. **deprecated EARS `IF/THEN` 일괄 마이그레이션** — 기존 규칙 본문의 다른 EARS 잔재 정리는 본 SPEC 범위 밖.

## §F — Traceability

| REQ | 마일스톤 | 대상 파일 | template-managed? | AC |
|-----|---------|-----------|-------------------|-----|
| REQ-CDA-001 | M1 | dynamic-workflows.md | Y (src+mirror) | AC-CDA-001 |
| REQ-CDA-002 | M1 | dynamic-workflows.md (+orch-mode) | Y | AC-CDA-002 |
| REQ-CDA-003 | M1 | dynamic-workflows.md | Y | AC-CDA-003 |
| REQ-CDA-004 | M1 | dynamic-workflows.md | Y | AC-CDA-004 |
| REQ-CDA-005 | M1 | dynamic-workflows.md | Y | AC-CDA-005 |
| REQ-CDA-006 | M1 | dynamic-workflows.md | Y | AC-CDA-006 |
| REQ-CDA-007 | M2 | skill-writing-craft.md | Y | AC-CDA-007 |
| REQ-CDA-008 | M2 | skill-writing-craft.md + skill-authoring.md | Y | AC-CDA-008 |
| REQ-CDA-009 | M2 | skill-authoring.md | Y | AC-CDA-009 |
| REQ-CDA-010 | M2 | skill-authoring.md + skill-writing-craft.md | Y | AC-CDA-010 |
| REQ-CDA-011 | M2 | skill-authoring.md | Y | AC-CDA-011 |
| REQ-CDA-012 | M2 | claude-code-skills-official.md | Y | AC-CDA-012 |
| REQ-CDA-013 | M2 | coding-standards.md | Y | AC-CDA-013 |
| REQ-CDA-014 | M2 | CLAUDE.md + skill-authoring.md | Y (CLAUDE.md split) | AC-CDA-014 |
| REQ-CDA-015 | M3 | hooks-system.md | Y | AC-CDA-015 |
| REQ-CDA-016 | M3 | hooks-system.md | Y | AC-CDA-016 |
| REQ-CDA-017 | M3 | CLAUDE.local.md + internal/hook/CLAUDE.md | N (LOCAL-ONLY) | AC-CDA-017 |
| REQ-CDA-018 | M3 | hooks-system.md | Y | AC-CDA-018 |
| REQ-CDA-019 | M3 | agent-common-protocol.md | Y | AC-CDA-019 |
| REQ-CDA-020 | M4 | goal-directive.md | Y | AC-CDA-020 |
| REQ-CDA-021 | M4 | goal-directive.md | Y | AC-CDA-021 |
| REQ-CDA-022 | M4 | goal-directive.md | Y | AC-CDA-022 |
| REQ-CDA-023 | M4 | goal-directive.md | Y | AC-CDA-023 |
| REQ-CDA-024 | M4 | goal-directive.md | Y | AC-CDA-024 |
| REQ-CDA-025 | M4 | goal-directive.md | Y | AC-CDA-025 |
| REQ-CDA-026 | M5 | agent-authoring.md | Y (split) | AC-CDA-026 |
| REQ-CDA-027 | M5 | agent-authoring.md + agent-hooks.md | Y | AC-CDA-027 |
| REQ-CDA-028 | M5 | agent-authoring.md or agent-patterns.md | Y | AC-CDA-028 |
| REQ-CDA-029 | M5 | archived-agent-rejection.md + CLAUDE.md | Y (CLAUDE.md split) | AC-CDA-029 |
| REQ-CDA-030 | M5 | model-policy.md | Y | AC-CDA-030 |
| REQ-CDA-031 | M5 | agent-authoring.md | Y (split) | AC-CDA-031 |
| REQ-CDA-032 | M5 | agent-authoring.md | Y (split) | AC-CDA-032 |
| REQ-CDA-033 | M5 | agent-authoring.md | Y (split) | AC-CDA-033 |
