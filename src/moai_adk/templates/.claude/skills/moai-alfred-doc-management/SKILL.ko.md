---
name: moai-alfred-doc-management-ko
version: 1.0.0
created: 2025-11-05
updated: 2025-11-05
status: active
description: 내부 문서 배치 규칙, 금지된 패턴 및 서브-에이전트 출력 가이드라인
keywords: ['documentation', 'file-locations', 'conventions', 'policies', 'management']
allowed-tools:
  - Read
  - Write
  - Bash
---

# 문서 관리 규칙

**중요**: Alfred 및 모든 Sub-agent는 이 문서 배치 규칙을 따라야 합니다.

## 허용된 문서 위치

| 문서 유형 | 위치 | 예시 |
| ----------------------- | --------------------- | ------------------------------------ |
| **내부 가이드** | `.moai/docs/` | 구현 가이드, 전략 문서 |
| **탐색 보고서** | `.moai/docs/` | 분석, 조사 결과 |
| **SPEC 문서** | `.moai/specs/SPEC-*/` | spec.md, plan.md, acceptance.md |
| **동기화 보고서** | `.moai/reports/` | 동기화 분석, TAG 검증 |
| **기술 분석** | `.moai/analysis/` | 아키텍처 연구, 최적화 |
| **메모리 파일** | `.moai/memory/` | 세션 상태만 (런타임 데이터) |
| **지식 베이스** | `.claude/skills/moai-alfred-*` | Alfred 워크플로우 가이드 (온디맨드) |

## 금지됨: 루트 디렉토리

**사용자가 명시적으로 요청하지 않으면 프로젝트 루트에 문서를 적극적으로 생성하지 마세요:**

- ❌ `IMPLEMENTATION_GUIDE.md`
- ❌ `EXPLORATION_REPORT.md`
- ❌ `*_ANALYSIS.md`
- ❌ `*_GUIDE.md`
- ❌ `*_REPORT.md`

**예외** (루트에 허용되는 파일만):

- ✅ `README.md` - 공식 사용자 문서
- ✅ `CHANGELOG.md` - 버전 이력
- ✅ `CONTRIBUTING.md` - 기여 가이드라인
- ✅ `LICENSE` - 라이선스 파일

## 문서 생성 결정 트리

```
.md 파일을 생성해야 하나요?
    ↓
사용자 대면 공식 문서입니까?
    ├─ 예 → 루트 (README.md, CHANGELOG.md만)
    └─ 아니요 → Alfred/워크플로우 내부용입니까?
             ├─ 예 → 유형 확인:
             │    ├─ 관련 → .moai/specs/SPEC-*/
             │    ├─ 동기화 보고서 → .moai/reports/
             │    ├─ 분석 → .moai/analysis/
             │    └─ 가이드/전략 → .moai/docs/
             └─ 아니요 → 생성하기 전에 사용자에게 명시적으로 요청
```

## 문서 명명 규칙

**`.moai/docs/`의 내부 문서**:

- `implementation-{SPEC-ID}.md` - 구현 가이드
- `exploration-{topic}.md` - 탐색/분석 보고서
- `strategy-{topic}.md` - 전략 계획 문서
- `guide-{topic}.md` - Alfred 사용 방법 가이드

## 서브-에이전트 출력 가이드라인

| 서브-에이전트 | 기본 출력 위치 | 문서 유형 |
| ---------------------- | ----------------------- | ------------------------ |
| implementation-planner | `.moai/docs/` | implementation-{SPEC}.md |
| Explore | `.moai/docs/` | exploration-{topic}.md |
| Plan | `.moai/docs/` | strategy-{topic}.md |
| doc-syncer | `.moai/reports/` | sync-report-{type}.md |
| tag-agent | `.moai/reports/` | tag-validation-{date}.md |

## 디렉토리 구조

**예상되는 MoAI 디렉토리 레이아웃**:

```
.moai/
├── config.json              # 프로젝트 설정
├── docs/                    # 내부 문서
│   ├── implementation-*.md
│   ├── exploration-*.md
│   ├── strategy-*.md
│   └── guide-*.md
├── specs/                   # SPEC 문서
│   ├── SPEC-ID-001/
│   │   ├── spec.md
│   │   ├── plan.md
│   │   └── acceptance.md
│   └── SPEC-ID-002/
├── reports/                 # 생성된 보고서
│   ├── sync-report-*.md
│   └── tag-validation-*.md
├── analysis/                # 기술 분석
│   └── *-analysis.md
└── memory/                  # 세션 상태 (런타임만)
    └── session-state.json
```

## 강제 규칙

### ❌ 방지해야할 위반 사항

1. **루트 레벨 생성 문서 없음**
   - 프로젝트 루트에 `IMPLEMENTATION_REPORT.md`, `ANALYSIS.md` 생성
   - **수정**: `.moai/docs/`, `.moai/analysis/`, 또는 `.moai/reports/`에 배치

2. **임의 파일 위치 없음**
   - `.moai/specs/SPEC-*/` 외부에 SPEC 문서 생성
   - **수정**: `.moai/specs/SPEC-ID-XXX/spec.md` 구조 사용

3. **동기화/분석 보고서를 docs/에 두지 않음**
   - `.moai/docs/`에 동기화 보고서 배치 (잘못됨)
   - **수정**: `.moai/reports/sync-report-*.md` 사용

### ✅ 올바른 패턴

**사용자에게 물어볼 때**:
- ".moai/docs/에 상세한 구현 가이드를 생성할까요?"
- ".moai/reports/에 동기화 보고서를 생성할 수 있습니다. 진행할까요?"

**자동 생성할 때**:
- SPEC 문서: `/alfred:1-plan` 중 → `.moai/specs/SPEC-*/`
- 동기화 보고서: `/alfred:3-sync` 중 → `.moai/reports/`
- 구현 가이드: `/alfred:2-run` 중 → `.moai/docs/` (접근법 문서화 시)

## 문서 수명 주기

| 단계 | 문서 유형 | 위치 | 생성자 |
|-------|---------------|----------|---------|
| SPEC | spec.md, plan.md, acceptance.md | `.moai/specs/SPEC-*` | spec-builder |
| BUILD | 구현 가이드 (선택사항) | `.moai/docs/` | tdd-implementer (필요시) |
| SYNC | sync-report-*.md | `.moai/reports/` | doc-syncer |
| ARCHIVE | README.md, CHANGELOG.md | 루트 (공개) | 프로젝트 소유자 관리 |