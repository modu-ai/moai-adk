# MoAI-ADK (MoAI Agentic Development Kit)

**Claude Code 최신 표준 기반 Spec-First TDD 완전 자동화 개발 시스템**

## 🗿 시스템 개요

MoAI-ADK는 4단계 파이프라인(SPECIFY → PLAN → TASKS → IMPLEMENT)을 통한 완전 자동화 개발 환경을 제공합니다.

### 메모리 계층 & 임포트 규칙(요약)
- 계층: 조직 정책 → 프로젝트 메모리(이 파일 및 @.claude/memory/*) → 사용자 메모리(`~/.claude/CLAUDE.md`)
- 임포트: `@path` 문법(상대/절대/`~/`), 최대 5-hop, 코드 블록/스팬 내 제외
- 개인 선호는 저장소 외부(사용자 메모리)로 관리하고, 이 저장소엔 팀 공유 지침만 포함

자세한 안내: @.claude/memory/README.md

### 핵심 철학

- **Spec-First**: 명세 없이는 코드 없음
- **TDD-First**: 테스트 없이는 구현 없음
- **Living Document**: 문서와 코드는 항상 동기화
- **Full Traceability**: Core @TAG 시스템으로 완전 추적

## 🔁 작업 루프 & 운영 원칙(중요)

- 대화/문서/커밋 언어: 한국어 고정
- 작업 루프: 문제 정의 → 작고 안전한 변경 → 변경 리뷰 → 리팩터링(반복)
- 변경 전 전체 맥락 파악: 관련 파일은 처음부터 끝까지 읽고, 정의·참조·호출·테스트·문서·설정을 전역 검색으로 확인
- 작게 나누기: 변경/커밋/PR의 범위를 최소화하고, 영향도는 1–3줄로 요약
- 가정 기록: Issue/PR/ADR에 가정·제약·의사결정 근거를 남김
- 대안 비교: 최소 2가지 대안을 장단점/위험 1줄씩 비교 후 가장 단순한 해법 선택
- 비밀/보안: 시크릿을 코드/로그/문서에 남기지 않음. 모든 입력 검증·정규화·인코딩, 파라미터화된 접근, 최소 권한 원칙

### 코딩 규칙(요약)
- 파일 ≤ 300 LOC, 함수 ≤ 50 LOC, 매개변수 ≤ 5, 순환 복잡도 ≤ 10(초과 시 분리)
- 입력 → 처리 → 반환 구조, 가드절 우선, 부수효과(I/O·네트워크·전역)는 경계층으로 격리
- 예외는 구체 타입만 처리, 구조화 로깅(민감정보 금지)과 요청/상관관계 ID 전파
- 시간대/DST 고려(저장 UTC, 표시 로캘), 상수는 심볼화(하드코딩 금지)

### 테스트/보안/클린 코드
- 새 코드엔 새 테스트, 버그 수정엔 회귀 테스트(먼저 실패하도록 작성)
- 테스트는 결정적·독립적, 외부 시스템은 가짜/계약 테스트로 대체, E2E 최소 성공/실패 각 1개
- 동시성/락/재시도 위험(중복/데드락 등) 선제 평가 및 테스트
- 안티패턴 금지: 전체 문맥 무시 수정, 비밀 노출, 경고 무시, 근거 없는 최적화/추상화, 광범위 예외

## 📚 프로젝트 메모리 (Import Files)

프로젝트 문서는 아래 카테고리로 구성되어 있으며, 전체 지도는 @.claude/memory/README.md 에서 확인합니다.

### 프로세스 & 핵심 원칙
- @.claude/memory/three_phase_process.md — 응답 구조(탐색→계획→구현) 표준
- @.claude/memory/project_guidelines.md — Small & Safe 루프, 에이전트 운영 원칙
- @.claude/memory/software_principles.md — Refactoring/Clean Code/TDD/API 패턴 요약

### 개발 표준
- @.claude/memory/coding_standards.md — 언어/프레임워크별 코딩 규칙 링크
- @.claude/memory/tdd_guidelines.md — Red→Green→Refactor 사이클 지침
- @.claude/memory/security_rules.md — ISMS-P 기반 보안/개인정보 규칙

### 협업 & Git
- @.claude/memory/team_conventions.md — 회의/PR/문서화 규약
- @.claude/memory/git_workflow.md — 브랜치 전략·리베이스·pre-commit 흐름
- @.claude/memory/git_commit_rules.md — Conventional Commit 규칙
- @.claude/memory/shared_checklists.md — PR/테스트/보안 공통 체크리스트

### 도구 & 운영
- @.claude/memory/bash_commands.md — 쉘 안전 수칙과 권장 도구
- @.claude/memory/README.md — 메모리 계층/임포트/템플릿 가이드
- @.moai/memory/common.md — 프로젝트 공통 운영 메모(자동 생성)
- @.moai/memory/<layer>-<tech>.md — 선택한 기술 스택별 메모(예: backend-python.md)

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
/moai:2-spec "JWT 기반 사용자 인증 시스템"
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

### 모델 사용 가이드 (opusplan 권장)

- 계획/설계/리뷰/ADR/Constitution:`opusplan` 모델을 선택 후 plan 단계에서 `opus`로 추론 > 실행 단계는 자동 `sonnet` 전환
- 구현/리팩터/디버깅/테스트: 기본 `sonnet`
- 문서 동기화/인덱싱/대량 갱신: `haiku`
- 명령/에이전트는 기본적으로 `sonnet`을 상속하며, 속도 중심 태스크만 `haiku`를 명시합니다.

예시:
```bash
# 설계/계획은 opusplan, 구현은 sonnet
claude --model opusplan
# ⏸ plan mode on (shift+tab to cycle) 선택 이후 계획 수립
# 계획 수립 이후 자동 ⏵⏵ accept edits on (shift+tab to cycle) 전환 
> /moai:3-plan SPEC-001
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

### Hook 시스템 자동 검증

- **PreToolUse**: Constitution 검증, 보안 검사
- **PostToolUse**: TAG 동기화, 문서 업데이트
- **SessionStart**: 프로젝트 상태 알림

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

> 버전과 최신 상태는 `moai status` 또는 `.moai/version.json`에서 확인하세요.

**🗿 "명세가 없으면 코드도 없다. 테스트가 없으면 구현도 없다."**
