---
name: moai:1-spec
description: EARS 명세 작성 + 브랜치/PR 생성
argument-hint: "제목1 제목2 ... | SPEC-ID 수정내용"
allowed-tools: Read, Write, Edit, MultiEdit, Grep, Glob, TodoWrite, Bash
---

# /moai:1-spec — SPEC 자동 제안/생성

## 기능

- ULTRATHINK: `.moai/project/{product,structure,tech}.md`를 분석해 구현 후보를 제안하고 사용자 승인 후 SPEC을 생성합니다.
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

## 🚀 최적화된 워크플로우 실행 순서

당신은 다음 순서로 에이전트들을 호출해야 합니다:

### Phase 1: 병렬 프로젝트 분석 (성능 최적화)

**동시에 수행**:

```
Task 1 (haiku): 프로젝트 구조 스캔
├── 언어/프레임워크 감지
├── 기존 SPEC 목록 수집
└── 우선순위 백로그 초안

Task 2 (sonnet): 심화 문서 분석
├── product.md 요구사항 추출
├── structure.md 아키텍처 분석
└── tech.md 기술적 제약사항
```

**성능 향상**: 기본 스캔과 심화 분석을 병렬 처리하여 대기 시간 최소화

### Phase 2: SPEC 문서 통합 작성

`spec-builder` 에이전트(sonnet)가 병렬 분석 결과를 통합하여:

- 프로젝트 문서 기반 기능 후보 제안
- 사용자 승인 후 SPEC 문서 작성 (MultiEdit 활용)
- 3개 파일 동시 생성 (spec.md, plan.md, acceptance.md)

### Phase 3: Git 작업 처리

`git-manager` 에이전트(haiku)가 최종 처리:

- **브랜치 생성**: 모드별 전략(Personal/Team) 적용
- **GitHub Issue 생성**: Team 모드에서 SPEC Issue 생성
- **초기 커밋**: SPEC 문서 커밋 및 태그 생성

**중요**: 각 에이전트는 독립적으로 실행되며, 에이전트 간 직접 호출은 금지됩니다.

## 작성 팁

- product/structure/tech 문서에 없는 정보는 새로 질문해 보완합니다.
- Acceptance Criteria는 Given/When/Then 3단으로 최소 2개 이상 작성하도록 유도합니다.
- TRUST 원칙 중 Readable(읽기 쉬움) 기준 완화로 인해 모듈 수가 권장치(기본 5)를 초과하는 경우, 근거를 SPEC `context` 섹션에 함께 기록하세요.

## 다음 단계

- `/moai:2-build SPEC-XXX`로 TDD 구현 시작
- 팀 모드: Issue 생성 후 git-manager 에이전트가 자동으로 브랜치 생성
