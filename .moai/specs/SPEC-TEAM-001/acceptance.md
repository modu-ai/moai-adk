---
spec_id: SPEC-TEAM-001
version: 1.0.0
status: backfilled
created_at: 2026-04-24
author: manager-spec (backfill)
backfill_reason: |
  Post-implementation SDD artifact. 정적 team-*.md 에이전트 5개가 삭제되고
  `workflow.yaml`에 role_profiles 섹션이 추가되었으며 `team/run.md`가
  `Agent(subagent_type: "general-purpose")` 기반 동적 생성으로 재작성됨.
  관련 결정 기록: ~/.claude/projects/-Users-goos-MoAI-moai-adk-go/
  memory/decision_dynamic_team_generation.md.
  spec.md의 REQ-1/2/3/4/5를 실제 파일 상태와 대조하여 AC 역도출.
  plan-auditor 2026-04-24 감사 시 acceptance.md 부재 확인.
---

# Acceptance Criteria — SPEC-TEAM-001

정적 `team-*` 에이전트 정의 파일을 제거하고 `Agent(subagent_type: "general-purpose")` + `workflow.yaml` role_profiles 기반 동적 팀 생성으로 전환하는 구현의 관찰 가능한 인수 기준.

## Traceability

| REQ ID | AC ID | Test / Evidence Reference |
|--------|-------|---------------------------|
| REQ-1 (정적 team-*.md 파일 제거) | AC-001, AC-002 | `ls .claude/agents/moai/ \| grep team` (결과 없음), 템플릿 동일 |
| REQ-2 (role_profiles 추가) | AC-003 ~ AC-005 | `.moai/config/sections/workflow.yaml:26-64` |
| REQ-3 (team/run.md 동적 생성) | AC-006, AC-007 | `.claude/skills/moai/team/run.md:5,32,43,114,145,154,179,201,307` |
| REQ-4 (문서 업데이트) | AC-008 | `CLAUDE.md` §4, §15 (Dynamic Team Generation) |
| REQ-5 (템플릿 싱크) | AC-009 | 로컬 + 템플릿 양쪽 동기화 |

## AC-001: 로컬 프로젝트에 team-*.md 정적 에이전트 파일이 존재하지 않는다

**Given** `.claude/agents/moai/` 디렉터리에서,
**When** `team-coder.md`, `team-reader.md`, `team-tester.md`, `team-designer.md`, `team-validator.md` 5개 파일의 존재 여부를 확인하면,
**Then** 5개 파일 모두 존재하지 않아야 한다(`ls .claude/agents/moai/ | grep team`의 결과가 빈 출력이어야 함).

**Verification**: `ls /Users/goos/MoAI/moai-adk-go/.claude/agents/moai/ | grep team` → empty. 현 디렉터리에는 builder-*, evaluator-*, expert-*, manager-*, plan-auditor.md, researcher.md만 존재.

## AC-002: 템플릿에도 team-*.md 정적 에이전트 파일이 존재하지 않는다

**Given** `internal/template/templates/.claude/agents/moai/` 디렉터리에서,
**When** 동일한 5개 team-*.md 파일을 찾으면,
**Then** 모든 파일이 부재해야 한다. 이는 `moai init`/`moai update`로 신규 프로젝트 생성 시에도 정적 team 에이전트가 배포되지 않음을 보장한다.

**Verification**: `ls internal/template/templates/.claude/agents/moai/ | grep team` → empty (grep 실행 결과 양쪽 모두 매칭 없음).

## AC-003: workflow.yaml의 team 섹션이 role_profiles 키를 포함한다

**Given** `.moai/config/sections/workflow.yaml` (그리고 템플릿 복사본)에서,
**When** `workflow.team.role_profiles` 경로를 조회하면,
**Then** 7개 프로파일 키가 정의되어야 한다: `researcher`, `analyst`, `architect`, `implementer`, `tester`, `designer`, `reviewer`.

**Verification**: `.moai/config/sections/workflow.yaml:26-64` — role_profiles 섹션에 정확히 7개 키 존재.

## AC-004: 각 role_profile이 mode / model / isolation / description을 정의한다

**Given** `workflow.team.role_profiles.<role>`의 각 항목에서,
**When** 스키마를 검사하면,
**Then** 모든 프로파일이 다음 필드를 가져야 한다:
- `mode`: `plan` 또는 `acceptEdits`
- `model`: `haiku` / `sonnet` / `opus` 중 하나
- `isolation`: `none` 또는 `worktree`
- `description`: 1줄 설명 문자열

**Verification**: `.moai/config/sections/workflow.yaml:28-64` — 7개 프로파일 모두 4필드 보유. 예: `implementer: {mode: acceptEdits, model: sonnet, isolation: worktree, description: "Code implementation (backend, frontend, full-stack)"}`.

## AC-005: 읽기 전용 role은 mode=plan + isolation=none을 갖는다

**Given** `role_profiles.researcher`, `role_profiles.analyst`, `role_profiles.architect`, `role_profiles.reviewer` 4개 프로파일에서,
**When** mode / isolation 값을 확인하면,
**Then** 모든 4개 프로파일이 `mode: plan`, `isolation: none`으로 설정되어 CLAUDE.md Section 14 Worktree Isolation Rules의 HARD 규칙("읽기 전용 teammates MUST NOT use isolation: worktree")을 만족해야 한다.

**Verification**: `.moai/config/sections/workflow.yaml` — researcher(line 28-32), analyst(33-37), architect(38-42), reviewer(58-62) 모두 mode: plan + isolation: none.

## AC-006: team/run.md가 subagent_type: "general-purpose"로 teammate를 생성한다

**Given** `.claude/skills/moai/team/run.md`(로컬 + 템플릿)에서,
**When** teammate 생성 예시를 조회하면,
**Then** 모든 Agent() 호출이 `subagent_type: "general-purpose"`를 사용해야 하며, `subagent_type: "team-coder"` 등의 정적 에이전트 참조가 전혀 존재하지 않아야 한다.

**Verification**: `.claude/skills/moai/team/run.md:43` (`| subagent_type | Always "general-purpose" | Full tool access |`), line 154/179/201/307에서 `Agent(subagent_type: "general-purpose", ...)` 사용. grep `subagent_type: "team-` 결과 0 matches.

## AC-007: team/run.md는 role profile 기반으로 model/mode/isolation을 런타임 주입한다

**Given** team/run.md의 spawn 예시에서,
**When** Agent() 파라미터 패턴을 검사하면,
**Then** `model`, `mode`, `isolation` 값이 정적 agent 정의 파일이 아닌 Agent() 호출 인자로 직접 전달되어야 한다. 예: `Agent(subagent_type: "general-purpose", team_name: "...", name: "backend-dev", model: "sonnet", mode: "acceptEdits", isolation: "worktree", prompt: "...")`.

**Verification**: `.claude/skills/moai/team/run.md:307` — `Agent(subagent_type: "general-purpose", team_name: "moai-run-SPEC-XXX", name: "backend-dev", model: "sonnet", mode: "acceptEdits", isolation: "worktree", prompt: "Backend role. File ownership: server-side code. ...")`.

## AC-008: CLAUDE.md가 동적 팀 생성 아키텍처를 문서화한다

**Given** `CLAUDE.md` 및 템플릿 `CLAUDE.md`에서,
**When** "Agent Catalog" 및 "Agent Teams (Experimental)" 섹션을 조회하면,
**Then** (a) Section 4에서 "Dynamic Team Generation (Experimental)" 하위 섹션이 존재하며 `Agent(subagent_type: "general-purpose")`와 `workflow.yaml role profiles`를 명시해야 하고, (b) Section 15에서도 동일 동적 생성 기술이 기재되며 "No static team agent definition files are used" 문장이 포함되어야 한다.

**Verification**: `CLAUDE.md` Section 4 "Dynamic Team Generation (Experimental)" 블록, Section 15 "Dynamic Team Generation" 블록. 템플릿 동일.

## AC-009: 로컬과 템플릿의 workflow.yaml이 동일한 role_profiles를 가진다

**Given** `.moai/config/sections/workflow.yaml`와 `internal/template/templates/.moai/config/sections/workflow.yaml`에서,
**When** `team.role_profiles` 하위 키 목록 및 필드 구성을 비교하면,
**Then** 두 파일이 동일한 7개 프로파일(researcher, analyst, architect, implementer, tester, designer, reviewer)과 동일한 스키마(mode/model/isolation/description)를 가져야 한다.

**Verification**: 양쪽 파일에서 `grep "role_profiles" workflow.yaml` 모두 hit. 로컬 vs 템플릿 동기화는 `make build` 이후 보장.

## Edge Cases

- **EC-01**: `CLAUDE_CODE_EXPERIMENTAL_AGENT_TEAMS` env 미설정 → 동적 팀 생성이 trigger되지 않고 sub-agent 모드로 graceful fallback (CLAUDE.md Section 15 참조).
- **EC-02**: `workflow.team.enabled: false` → 팀 모드 비활성화, 동일 sub-agent 폴백.
- **EC-03**: 삭제된 `team-*.md`에 대한 잔존 레거시 참조가 내부 Go 코드에 없음 — spec.md §Out of Scope에 "no Go code references deleted agent files" 명시 및 검증됨.
- **EC-04**: 스킬/훅 제거로 인한 기능 공백 보완 — Global TeammateIdle/TaskCompleted 훅 + prompt 임베디드 지시문으로 대체(team/run.md Quality verification 섹션).

## Partial / Deferred

- **팀 훅 스킬 프리로드 부재(리스크 R 항목)**: 각 teammate가 자체적으로 `Skill()` 호출하거나 프롬프트에서 필요한 스킬 ID를 참조받도록 전환. 검증은 test suite에 포함되지 않아 런타임 관찰 필요.
- **general-purpose agent가 과도한 도구 세트를 가짐(R3)**: Accepted trade-off로 문서화. 추가 제한이 필요하면 후속 SPEC에서 다룸.

## Definition of Done

- [x] `.claude/agents/moai/team-*.md` 5개 파일 부재 (로컬) (AC-001)
- [x] `internal/template/templates/.claude/agents/moai/team-*.md` 5개 파일 부재 (템플릿) (AC-002)
- [x] `workflow.yaml`에 7개 role_profiles 정의 (AC-003, AC-004, AC-005)
- [x] `team/run.md`가 `subagent_type: "general-purpose"`로 spawn (AC-006, AC-007)
- [x] CLAUDE.md Section 4, 15 동적 팀 생성 문서화 (AC-008)
- [x] 로컬-템플릿 동기화 (AC-009)
- [x] `go test ./...` 통과 (static team file 참조 없음 검증)
- [x] decision_dynamic_team_generation.md 기록 생성
