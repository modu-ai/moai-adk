---
name: moai:1-spec
description: EARS 명세 작성 + 브랜치/PR 생성
argument-hint: ["제목1" "제목2" ...] | [SPEC-ID "수정내용"]
allowed-tools: Read, Write, Edit, MultiEdit, Grep, Glob, TodoWrite, Bash
---

# /moai:1-spec — SPEC 자동 제안/생성

## 기능

- `.moai/project/{product,structure,tech}.md`를 분석해 구현 후보를 제안하고 사용자 승인 후 SPEC을 생성합니다.
- **Personal 모드**: `.moai/specs/SPEC-XXX/` 디렉터리와 템플릿 문서를 만듭니다.
- **Team 모드**: GitHub Issue(또는 Discussion)를 생성하고 브랜치 템플릿과 연결합니다.

## 에이전트 협업 구조

- **1단계**: `spec-builder` 에이전트가 프로젝트 문서 분석 및 SPEC 문서 작성을 전담합니다.
- **2단계**: `git-manager` 에이전트가 브랜치 생성, GitHub Issue/PR 생성을 전담합니다.
- **단일 책임 원칙**: spec-builder는 SPEC 작성만, git-manager는 Git/GitHub 작업만 수행합니다.
- **순차 실행**: spec-builder → git-manager 순서로 실행하여 명확한 의존성을 유지합니다.
- **에이전트 간 호출 금지**: 각 에이전트는 다른 에이전트를 직접 호출하지 않고, 커맨드 레벨에서만 순차 실행합니다.

## 사용법

```bash
/moai:1-spec                      # 프로젝트 문서 기반 자동 제안 (권장)
/moai:1-spec "JWT 인증 시스템"       # 단일 SPEC 수동 생성
/moai:1-spec SPEC-001 "보안 보강"   # 기존 SPEC 보완
```

입력하지 않으면 Q&A 결과를 기반으로 우선순위 3~5건을 제안하며, 승인한 항목만 실제 SPEC으로 확정됩니다.

## 모드별 처리 요약

| 모드     | 산출물                                                               | 추가 작업                                       |
| -------- | -------------------------------------------------------------------- | ----------------------------------------------- |
| Personal | `.moai/specs/SPEC-XXX/spec.md`, `plan.md`, `acceptance.md` 등 템플릿 | git-manager 에이전트가 자동으로 체크포인트 생성 |
| Team     | GitHub Issue(`[SPEC-XXX] 제목`), Draft PR(옵션)                      | `gh` CLI 로그인 유지, 라벨/담당자 지정 안내     |

## 입력 옵션

- **자동 제안**: `/moai:1-spec` → 프로젝트 문서 핵심 bullet을 기반으로 후보 리스트 작성
- **수동 생성**: 제목을 인수로 전달 → 1건만 생성, Acceptance 템플릿은 회신 후 보완
- **보완 모드**: `SPEC-ID "메모"` 형식으로 전달 → 기존 SPEC 문서/Issue를 업데이트

## 브레인스토밍(선택)

- `.moai/config.json.brainstorming.enabled` 가 `true` 이고 `providers` 배열이 비어 있지 않은 경우 다음 단계를 추가합니다.
  1. `codex` 가 포함되어 있으면 `codex-bridge` 에이전트를 호출하여 `codex exec --model gpt-5-codex "..."` 형태의 headless 분석을 실행하고 대안 아이디어를 수집합니다. (예: `Task: use codex-bridge to run "codex exec --model gpt-5-codex \"Summarize design risks\""`)
  2. `gemini` 가 포함되어 있으면 `gemini-bridge` 에이전트를 호출하여 `gemini -m gemini-2.5-pro -p "..." --output-format json` 명령을 실행하고 구조화된 제안을 수집합니다. (예: `Task: use gemini-bridge to run "gemini -m gemini-2.5-pro -p 'List alternative solution paths' --output-format json"`)
  3. `claude` 는 기본 분석 경로로 유지하며, 외부 응답과 비교해 Self-Consistency/ToT/Meta-Prompting 절차로 최적안을 선정합니다.
- 외부 브레인스토밍을 사용하지 않는 경우(기본값)에는 Claude Code만 활용합니다.

## 워크플로우 실행 순서

당신은 다음 순서로 에이전트들을 **순차 호출**해야 합니다:

### 1단계: SPEC 문서 작성

먼저 `spec-builder` 에이전트를 호출하여 프로젝트 문서 분석 및 SPEC 작성을 완료합니다.

### 2단계: Git 작업 처리

`spec-builder` 완료 후, `git-manager` 에이전트를 호출하여 다음 작업을 수행합니다:

- **브랜치 생성**: 모드별 전략(Personal/Team) 적용
- **GitHub Issue 생성**: Team 모드에서 SPEC Issue 생성
- **초기 커밋**: SPEC 문서 커밋 및 태그 생성

**중요**: 각 에이전트는 독립적으로 실행되며, 에이전트 간 직접 호출은 금지됩니다.

## 작성 팁

- product/structure/tech 문서에 없는 정보는 새로 질문해 보완합니다.
- Acceptance Criteria는 Given/When/Then 3단으로 최소 2개 이상 작성하도록 유도합니다.
- Constitution 제1조 완화로 인해 모듈 수가 권장치(기본 3)를 초과하는 경우, 근거를 SPEC `context` 섹션에 함께 기록하세요.

## 다음 단계

- `/moai:2-build SPEC-XXX`로 TDD 구현 시작
- 팀 모드: Issue 생성 후 git-manager 에이전트가 자동으로 브랜치 생성
