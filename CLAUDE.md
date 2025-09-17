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

## 🔁 작업 루프 & 핵심 원칙

- 모든 커뮤니케이션은 한국어, 작업 루프는 “문제 정의 → Small & Safe 변경 → 리뷰 → 리팩터링”을 반복합니다.
- 변경 전 전체 맥락(정의·참조·호출·테스트·문서)을 전역 검색으로 확인하고, 영향도는 1–3줄로 정리합니다.
- Issue/PR/ADR에는 가정과 최소 두 가지 대안(장단점/위험)을 기록합니다.
- 시크릿·민감정보는 절대 저장소에 남기지 않으며 입력 검증·파라미터화·최소 권한 원칙을 기본으로 합니다.
- 세부 규칙은 @.claude/memory/project_guidelines.md (운영)과 @.claude/memory/shared_checklists.md (PR/테스트/보안)을 참조하세요.

### 코딩 · 테스트 · 보안 요약
- 기본 코딩 기준은 @.claude/memory/coding_standards.md, 언어/프레임워크별 세부 문서는 해당 @imports에서 확인합니다.
- TDD는 Red → Green → Refactor 사이클(@.claude/memory/tdd_guidelines.md)로 수행하고 커버리지는 80% 이상 유지합니다.
- 보안/개인정보는 ISMS-P 규칙(@.claude/memory/security_rules.md)을 준수합니다.

## 📚 프로젝트 메모리 (Import Files)

프로젝트 문서는 아래 카테고리로 구성되어 있으며, 전체 지도는 @.claude/memory/README.md 에서 확인합니다.

### 📍 Steering 문서 (프로젝트 방향성)
- 제품 비전: @.moai/steering/product.md
- 구조 설계: @.moai/steering/structure.md
- 기술 스택: @.moai/steering/tech.md

| 카테고리 | 주요 문서 |
| --- | --- |
| 프로세스/운영 | @.claude/memory/three_phase_process.md, @.claude/memory/project_guidelines.md, @.claude/memory/software_principles.md |
| 개발 표준 | @.claude/memory/coding_standards.md, @.claude/memory/tdd_guidelines.md, @.claude/memory/security_rules.md |
| 협업 & Git | @.claude/memory/team_conventions.md, @.claude/memory/git_workflow.md, @.claude/memory/git_commit_rules.md, @.claude/memory/shared_checklists.md |
| 도구 & 운영 | @.claude/memory/bash_commands.md, @.claude/memory/README.md, `.moai/memory/common.md`, `.moai/memory/<layer>-<tech>.md` |
| 거버넌스 | @.moai/memory/constitution.md |

## 🚀 빠른 시작

### 1. 프로젝트 초기화

```bash
/moai:1-project
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
├── .claude/                    # Claude Code 표준 자산
│   ├── commands/moai/          # MoAI 슬래시 명령어 (연번순)
│   ├── agents/moai/            # 11개 전문 에이전트
│   ├── hooks/moai/             # Python Hook Scripts
│   ├── memory/                 # 공유 메모리(링크만 유지)
│   └── settings.json           # 권한 및 Hook 설정
├── .moai/                      # MoAI 문서 시스템
│   ├── steering/               # 프로젝트 방향성 문서
│   ├── specs/                  # SPEC 문서(동적 생성)
│   ├── memory/                 # 프로젝트 메모리(automatic)
│   │   ├── common.md           # 공통 운영 체크
│   │   ├── backend-*.md        # 백엔드 스택별 메모
│   │   └── frontend-*.md       # 프론트엔드 스택별 메모
│   ├── scripts/                # 검증 스크립트
│   └── config.json             # MoAI 설정
└── CLAUDE.md                   # 프로젝트 메모리 허브 (이 파일)
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
