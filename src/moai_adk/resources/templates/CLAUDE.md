# MoAI-ADK (MoAI Agentic Development Kit) v0.1.17

**Claude Code 최신 표준 기반 Spec-First TDD 완전 자동화 개발 시스템**

## 🗿 시스템 개요

MoAI-ADK는 Claude Code의 2025년 최신 기능(Claude 4.1 Opus 모델, "ask" 권한 defaultMode, MCP 서버 통합, Hook stdin JSON 처리)을 활용하여 4단계 파이프라인(SPECIFY → PLAN → TASKS → IMPLEMENT)을 통한 완전 자동화 개발 환경을 제공합니다.

### 핵심 철학

- **Spec-First**: 명세 없이는 코드 없음
- **TDD-First**: 테스트 없이는 구현 없음
- **Living Document**: 문서와 코드는 항상 동기화
- **Full Traceability**: Core @TAG 시스템으로 완전 추적

## 📚 프로젝트 메모리 (Import Files)

### 개발 가이드라인

@.claude/memory/project_guidelines.md
@.claude/memory/coding_standards.md  
@.claude/memory/team_conventions.md

### 명령어 레퍼런스

@.claude/memory/bash_commands.md
@.claude/memory/git_workflow.md

### Constitution & 거버넌스

@.moai/memory/constitution.md

## 🚀 빠른 시작

### 1. 프로젝트 초기화

```bash
/moai:1-project init
```

대화형 마법사가 프로젝트를 완전 설정합니다.

### 2. 첫 번째 기능 개발 (자동화)

```bash
# 전체 파이프라인 자동 실행
/moai:2-spec user-auth "JWT 기반 사용자 인증 시스템"
/moai:3-plan SPEC-001
/moai:4-tasks PLAN-001
/moai:5-dev T001
```

### 3. 품질 검증

```bash
# 추적성 확인
python .moai/scripts/check-traceability.py --verbose

# 전체 테스트 실행
bash .moai/scripts/run-tests.sh
```

## 🤖 주요 명령어

| 순서  | 명령어            | 담당 에이전트                   | 기능                |
| ----- | ----------------- | ------------------------------- | ------------------- |
| **1** | `/moai:1-project` | steering-architect              | 프로젝트 설정       |
| **2** | `/moai:2-spec`    | spec-manager                    | EARS 형식 명세 작성 |
| **3** | `/moai:3-plan`    | plan-architect                  | Constitution Check  |
| **4** | `/moai:4-tasks`   | task-decomposer                 | TDD 작업 분해       |
| **5** | `/moai:5-dev`     | code-generator + test-automator | 자동 구현           |
| **6** | `/moai:6-sync`    | doc-syncer + tag-indexer        | 문서 동기화         |

## 🤖 에이전트 모델 표준

### 모델명 규칙

MoAI-ADK의 모든 에이전트는 **짧은 모델명만** 사용합니다:

- **opus**: 복잡한 추론과 설계 작업
  - steering-architect, plan-architect, code-generator
- **sonnet**: 균형잡힌 범용 작업
  - spec-manager, task-decomposer, test-automator, claude-code-manager
- **haiku**: 빠른 처리 작업
  - doc-syncer, tag-indexer

**⚠️ 중요**: 전체 모델명(예: `claude-3-5-sonnet-20241022`) 사용 금지
**✅ 올바른 형식**: `model: sonnet`
**❌ 잘못된 형식**: `model: claude-3-5-sonnet-20241022`

### 모델 사용 가이드 (opusplan 권장)

- 계획/설계/리뷰/ADR/Constitution: plan 모드에서 `opusplan` 권장
  - plan 단계에서 `opus`로 추론 후 실행 단계는 자동 `sonnet` 전환
- 구현/리팩터/디버깅/테스트: 기본 `sonnet`
- 문서 동기화/인덱싱/대량 갱신: `haiku`
- 명령/에이전트는 기본적으로 `sonnet`을 상속하며, 속도 중심 태스크만 `haiku`를 명시합니다.

예시:

```bash
# 설계/계획은 opusplan, 구현은 sonnet
claude --model opusplan
> /think plan
> /moai:3-plan SPEC-001
/model sonnet
> /moai:5-dev T001
```

## 🏷️ 핵심 TAG 활용

### 요구사항 추적

```markdown
@REQ:USER-AUTH "사용자 인증 요구사항"
→ @DESIGN:JWT-AUTH "JWT 기반 설계"
→ @TASK:AUTH-API "인증 API 구현"
→ @TEST:AUTH-001 "인증 테스트"
```

### 추적성 체인

- **Primary**: @REQ → @DESIGN → @TASK → @TEST
- **Steering**: @VISION → @STRUCT → @TECH → @ADR
- **Quality**: @PERF → @SEC → @DEBT → @TODO

## 🛡️ 자동화된 품질 보장

### Constitution 5원칙 강제

1. **Simplicity**: 프로젝트 복잡도 ≤ 3개
2. **Architecture**: 모든 기능은 라이브러리로
3. **Testing**: RED-GREEN-REFACTOR 강제
4. **Observability**: 구조화된 로깅 필수
5. **Versioning**: MAJOR.MINOR.BUILD 체계

### Hook 시스템 자동 검증 (v0.1.12 개선)

- **PreToolUse**: Constitution 검증, 보안 검사 (stdin JSON 처리로 개선)
- **PostToolUse**: TAG 동기화, 문서 업데이트
- **SessionStart**: 프로젝트 상태 알림
- **MCP 통합**: memory, filesystem 서버 자동 연결

### Hallucination 저감 3원칙

1. 모르면 “모름”이라고 명시한다.
2. 모든 주장/설계에는 근거를 로컬 파일 경로로 인용한다. 근거가 없으면 해당 문장을 제거한다.
3. 외부 지식은 금지한다(명시적으로 허용된 자료/리서치 범위에서만 사용).

### 파일 생성/변경 가드(운영 규정)

- 단일 응답 내 신규 파일 생성은 최대 5개, 총 생성 용량은 200KB를 권장 상한으로 한다.
- 대량 생성/재생성은 `/moai:6-sync force` 또는 명시적 플래그 사용 시에만 수행한다.
- 민감 경로(`.env`, `.git/`, `keys`, `secrets`) 수정/생성 금지.

## 📂 프로젝트 구조

```
프로젝트/
├── .claude/                    # Claude Code 표준
│   ├── commands/moai/          # MoAI 슬래시 명령어 (연번순)
│   ├── agents/moai/           # MoAI 전문 에이전트 (claude-code-manager 포함)
│   ├── hooks/moai/            # Python Hook Scripts
│   ├── memory/               # Import 메모리 파일들
│   │   ├── project_guidelines.md
│   │   ├── coding_standards.md
│   │   ├── team_conventions.md
│   │   ├── bash_commands.md
│   │   └── git_workflow.md
│   └── settings.json         # 권한 및 Hook 설정
├── .moai/                     # MoAI 문서 시스템
│   ├── steering/             # 프로젝트 방향성
│   ├── specs/               # SPEC 문서
│   ├── memory/              # Constitution & 거버넌스
│   ├── scripts/             # 검증 스크립트
│   └── config.json          # MoAI 설정
└── CLAUDE.md                # 프로젝트 메모리 (이 파일)
```

## 💡 Pro Tips

### 개발 효율성

- **컨텍스트 관리**: 단계별 `/clear` 실행 권장
- **병렬 개발**: [P] 마커 작업은 동시 실행 가능
- **자동 동기화**: 코드 변경시 문서 자동 업데이트

### 품질 관리

- **Hook 검증**: 모든 위험 요소 자동 차단
- **TAG 일관성**: 실시간 추적성 검증
- **TDD 강제**: 테스트 없는 구현 차단

## 🔧 문제 해결

### 일반적인 이슈

1. **Hook 실행 실패**: `chmod +x .claude/hooks/moai/*.py`
2. **TAG 불일치**: `python scripts/repair_tags.py --execute`
3. **커버리지 부족**: `/moai:5-dev` 재실행으로 TDD 사이클 완성

### 품질 게이트 실패

각 단계의 품질 게이트 실패시 이전 단계로 자동 롤백되며, 이슈 해결 후 재실행 가능합니다.

## 📈 성과 지표

### 자동 수집 메트릭

- **개발 생산성**: 명세 완성도 ≥90%, 구현 속도 50% 향상
- **품질 지표**: 테스트 커버리지 ≥80%, 버그 감소 70%
- **추적성**: TAG 정확성 ≥95%, 체인 완성도 ≥90%

### 실전 예제

- 프로젝트 가이드라인: `@.claude/memory/project_guidelines.md`
- Bash 명령어 모음: `@.claude/memory/bash_commands.md`
- Git 워크플로우: `@.claude/memory/git_workflow.md`

---

**마지막 업데이트**: 2025-09-16
**다음 검토 예정**: 2025-10-16
**MoAI-ADK 버전**: v0.1.17

**🗿 "명세가 없으면 코드도 없다. 테스트가 없으면 구현도 없다."**
