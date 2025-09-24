# MoAI-ADK Agent 시스템

## 🤖 에이전트 개요

MoAI-ADK는 개발 파이프라인의 각 단계를 자동화하기 위해 집중된 역할을 가진 에이전트들을 제공합니다. 모든 에이전트는 `.claude/agents/moai/` 디렉터리에 존재하며, 커맨드가 명시적으로 위임할 때만 동작합니다.

| 에이전트          | 역할                               | 주요 커맨드                      |
| ----------------- | ---------------------------------- | -------------------------------- |
| `project-manager` | 프로젝트 킥오프 인터뷰, 설정 관리  | `/moai:0-project`                |
| `cc-manager`      | Claude Code 권한/훅/환경 점검      | `/moai:0-project` 실행 직후 자동 |
| `spec-builder`    | SPEC 자동 제안 및 문서 생성        | `/moai:1-spec`                   |
| `code-builder`    | TDD 구현(RED→GREEN→REFACTOR)       | `/moai:2-build`                  |
| `doc-syncer`      | 문서/PR 동기화, TAG 인덱스 관리    | `/moai:3-sync`                   |
| `git-manager`     | 체크포인트/브랜치/커밋/동기화 전담 | `/moai:git:*` 계열               |

## 🧭 에이전트별 상세 설명

### project-manager

- `/moai:0-project` 실행 시 신규/레거시 감지, product/structure/tech 인터뷰 진행
- 프로젝트 설정 및 초기화 관리

### cc-manager

- Guard/훅 권한을 점검하고 Claude Code 환경을 최적화
- project-manager 이후 실행되어 설정 충돌을 예방

### spec-builder

- 프로젝트 문서를 요약하여 SPEC 후보를 제안하고 문서를 생성
- 팀 모드에서는 GitHub Issue로 변환할 수 있도록 정보를 제공

### code-builder

- TDD 사이클을 강제하며 단계별 체크포인트/커밋 전략을 수행
- Red-Green-Refactor 패턴으로 체계적인 구현 진행

### doc-syncer

- 코드 ↔ 문서 일관성을 유지하고 16-Core @TAG를 업데이트
- 팀 모드에서 Draft → Ready 전환, 리뷰어 할당 등의 PR 절차를 보조

### git-manager

- 브랜치 생성, 체크포인트, 커밋, 동기화, 롤백 등 모든 Git 작업을 담당
- `/moai:git:*` 명령에서 직접 호출되며 다른 에이전트가 Git 명령을 수행하지 않도록 보장

## 🔄 실행 흐름 요약

```mermaid
flowchart TD
    A[/moai:0-project] -->|project-manager| B[/moai:1-spec]
    B -->|spec-builder| C[/moai:2-build]
    C -->|code-builder| D[/moai:3-sync]
    B -. git tasks .-> G[git-manager]
    C -. git tasks .-> G
    D -. git tasks .-> G
```

## 🧪 에이전트 사용 시 체크리스트

- 각 에이전트는 **단일 책임**만 수행하고 다른 에이전트를 직접 호출하지 않습니다.
- 커맨드에서 순차/병렬 흐름을 명시적으로 관리하여 의존성 충돌을 방지합니다.
- 모든 외부 도구 호출 시 설치 여부를 확인하고 사용자 동의를 구합니다.

## 🔧 커스터마이징 가이드

### 모델 변경

- 에이전트 단위 모델 지정은 각 에이전트 마크다운의 `model` 필드에서 수행합니다.
- project-manager, spec-builder, code-builder, doc-syncer, git-manager 기본값은 `sonnet` 입니다.

### 설정 정책 문서화

```markdown
# .claude/memory/team_conventions.md

- 모든 에이전트는 단일 책임 원칙을 따른다.
- Git 작업은 git-manager 에이전트가 전담한다.
```
