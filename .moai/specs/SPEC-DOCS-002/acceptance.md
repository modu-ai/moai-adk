# SPEC-DOCS-002 수락 기준 (Acceptance Criteria)

> **SPEC ID**: DOCS-002
> **제목**: MoAI-ADK 핵심 개념 문서
> **버전**: 0.0.1 (draft)
> **작성자**: @Goos

---

## 📋 개요

본 문서는 SPEC-DOCS-002의 상세한 수락 기준을 정의합니다. 모든 기준은 Given-When-Then 형식으로 작성되며, 검증 가능하고 측정 가능해야 합니다.

---

## 1. EARS 가이드 (ears-guide.md)

### AC-001: EARS 개념 설명 완료
**Given**: 사용자가 EARS 방법론을 처음 접한다
**When**: ears-guide.md를 읽는다
**Then**:
- [ ] EARS의 정의, 목적, 장점이 명확히 설명되어 있다
- [ ] 5가지 구문 유형(Ubiquitous, Event-driven, State-driven, Optional, Constraints)이 각각 설명되어 있다
- [ ] 각 구문의 형식이 명확히 제시되어 있다 (예: "시스템은 [기능]을 제공해야 한다")

**검증 방법**:
```bash
# EARS 5가지 구문이 모두 언급되었는지 확인
rg -i "(ubiquitous|event-driven|state-driven|optional|constraints)" docs/guides/concepts/ears-guide.md
```

---

### AC-002: EARS 구문별 실제 예시 제공
**Given**: 사용자가 EARS 구문을 이해했다
**When**: 실제 예시 섹션을 읽는다
**Then**:
- [ ] 각 구문별로 3개 이상의 실제 예시가 제공된다 (총 15개 이상)
- [ ] 인증 시스템(AUTH-001) 예시가 포함된다
- [ ] 파일 업로드(UPLOAD-001) 예시가 포함된다
- [ ] 결제 시스템(PAYMENT-001) 예시가 포함된다
- [ ] 각 예시는 실제 프로젝트에 바로 적용 가능한 수준이다

**검증 방법**:
```bash
# AUTH-001, UPLOAD-001, PAYMENT-001 예시가 모두 포함되었는지 확인
rg "(AUTH-001|UPLOAD-001|PAYMENT-001)" docs/guides/concepts/ears-guide.md
```

---

### AC-003: 안티 패턴 → 개선안 전환 예시
**Given**: 사용자가 모호한 요구사항을 작성했다
**When**: 안티 패턴 섹션을 읽는다
**Then**:
- [ ] 5개 이상의 안티 패턴 예시가 제공된다
- [ ] 각 안티 패턴에 대응하는 개선안이 명확히 제시된다
- [ ] 개선안은 EARS 구문을 따른다
- [ ] 왜 안티 패턴인지 설명이 포함된다

**예시 형식**:
```
❌ 안티 패턴: "사용자 친화적이어야 한다"
   문제점: 측정 불가능, 모호함
✅ 개선안: "WHEN 사용자가 입력 오류를 발생시키면, 시스템은 3초 이내에 명확한 오류 메시지를 표시해야 한다"
```

**검증 방법**:
```bash
# 안티 패턴 → 개선안 패턴이 5개 이상 있는지 확인
rg "❌.*안티.*패턴|✅.*개선.*안" docs/guides/concepts/ears-guide.md | wc -l
```

---

### AC-004: 실습 가능한 템플릿 제공
**Given**: 사용자가 새로운 SPEC을 작성하려고 한다
**When**: EARS 가이드 끝부분의 템플릿 섹션을 본다
**Then**:
- [ ] 복사-붙여넣기 가능한 EARS 템플릿이 제공된다
- [ ] 템플릿에 각 구문별 예시 주석이 포함된다
- [ ] 템플릿을 사용한 간단한 실습 예제가 포함된다

**검증 방법**:
```bash
# 템플릿 섹션이 존재하는지 확인
rg -i "template|템플릿" docs/guides/concepts/ears-guide.md
```

---

### AC-005: 문서 길이 제약
**Given**: EARS 가이드가 작성되었다
**When**: 문서 전체 길이를 확인한다
**Then**:
- [ ] 문서 전체가 3,000자 이내이다 (공백 제외)
- [ ] 초과 시 별도 "EARS 고급 가이드" 문서로 분리한다

**검증 방법**:
```bash
# 문서 길이 확인 (공백 제외)
wc -m docs/guides/concepts/ears-guide.md
```

---

## 2. TRUST 원칙 가이드 (trust-principles.md)

### AC-006: TRUST 개요 및 각 원칙 설명
**Given**: 사용자가 TRUST 원칙을 처음 접한다
**When**: trust-principles.md를 읽는다
**Then**:
- [ ] TRUST의 정의(Test First, Readable, Unified, Secured, Trackable)가 명확히 설명된다
- [ ] AI 시대의 코드 품질 보장 필요성이 설명된다
- [ ] 각 원칙별로 400~600자 분량의 상세 설명이 제공된다
- [ ] 각 원칙이 왜 중요한지, 어떻게 적용하는지 명확히 설명된다

**검증 방법**:
```bash
# TRUST 각 원칙이 모두 언급되었는지 확인
rg "(Test First|Readable|Unified|Secured|Trackable)" docs/guides/concepts/trust-principles.md
```

---

### AC-007: 언어별 구현 예시 제공
**Given**: 사용자가 Python, TypeScript, Go 프로젝트를 진행한다
**When**: 언어별 구현 섹션을 읽는다
**Then**:
- [ ] Python 구현 예시가 제공된다 (pytest, ruff, mypy, bandit)
- [ ] TypeScript 구현 예시가 제공된다 (Vitest, Biome, npm audit)
- [ ] Go 구현 예시가 제공된다 (go test, gofmt, gosec)
- [ ] 각 언어별로 실제 실행 가능한 예시 코드가 포함된다
- [ ] 도구 설치 방법이 간략히 안내된다

**검증 방법**:
```bash
# 언어별 도구가 모두 언급되었는지 확인
rg "(pytest|Vitest|go test)" docs/guides/concepts/trust-principles.md
```

---

### AC-008: TRUST 체크리스트 템플릿 제공
**Given**: 사용자가 코드 리뷰를 진행한다
**When**: TRUST 체크리스트 섹션을 본다
**Then**:
- [ ] 코드 리뷰 시 확인할 TRUST 항목이 체크리스트로 제공된다
- [ ] 각 항목은 측정 가능하고 명확하다
- [ ] 체크리스트는 복사-붙여넣기 가능한 형식이다

**예시 형식**:
```markdown
## TRUST 체크리스트
- [ ] **T** - 테스트 커버리지 ≥85%
- [ ] **R** - 함수 ≤50 LOC, 복잡도 ≤10
- [ ] **U** - 아키텍처 일관성 유지
- [ ] **S** - 보안 취약점 0건
- [ ] **T** - TAG 체인 무결성 확인
```

**검증 방법**:
```bash
# 체크리스트 섹션이 존재하는지 확인
rg "체크리스트|checklist" docs/guides/concepts/trust-principles.md -i
```

---

### AC-009: 실제 코드 리뷰 시나리오 포함
**Given**: 사용자가 TRUST 원칙을 코드 리뷰에 적용하려고 한다
**When**: 코드 리뷰 시나리오 섹션을 읽는다
**Then**:
- [ ] 1개 이상의 실제 코드 리뷰 시나리오가 제공된다
- [ ] 시나리오는 "Before(문제 코드) → After(개선 코드)" 형식이다
- [ ] 각 시나리오에 TRUST 원칙 적용 과정이 설명된다

**검증 방법**:
```bash
# 코드 리뷰 시나리오가 포함되었는지 확인
rg "(시나리오|scenario|Before|After)" docs/guides/concepts/trust-principles.md -i
```

---

## 3. TAG 시스템 가이드 (tag-system.md)

### AC-010: TAG 시스템 개요 및 TAG 체계도
**Given**: 사용자가 TAG 시스템을 처음 접한다
**When**: tag-system.md를 읽는다
**Then**:
- [ ] CODE-FIRST 철학이 명확히 설명된다
- [ ] TAG 라이프사이클(SPEC → TEST → CODE → DOC)이 설명된다
- [ ] TAG ID 규칙(`<DOMAIN>-<3자리>`)이 명확히 설명된다
- [ ] TAG 체계도가 다이어그램(Mermaid)으로 제공된다

**예시 다이어그램**:
```mermaid
graph TD
  A[@SPEC:AUTH-001] --> B[@TEST:AUTH-001]
  B --> C[@CODE:AUTH-001]
  C --> D[@DOC:AUTH-001]
```

**검증 방법**:
```bash
# Mermaid 다이어그램이 포함되었는지 확인
rg "mermaid" docs/guides/concepts/tag-system.md -i
```

---

### AC-011: TAG 유형별 사용법 및 서브 카테고리 설명
**Given**: 사용자가 TAG를 코드에 적용하려고 한다
**When**: TAG 유형별 사용법 섹션을 읽는다
**Then**:
- [ ] @SPEC, @TEST, @CODE, @DOC 각각의 역할이 설명된다
- [ ] @CODE 서브 카테고리(API, UI, DATA, DOMAIN, INFRA)가 설명된다
- [ ] 각 TAG 유형별 실제 코드 예시가 제공된다

**검증 방법**:
```bash
# TAG 유형이 모두 언급되었는지 확인
rg "(@SPEC|@TEST|@CODE|@DOC)" docs/guides/concepts/tag-system.md
```

---

### AC-012: rg 명령어 치트 시트 제공
**Given**: 사용자가 TAG 검증을 수행하려고 한다
**When**: rg 명령어 치트 시트 섹션을 본다
**Then**:
- [ ] TAG 중복 확인 명령어가 제공된다 (`rg "@SPEC:AUTH" -n`)
- [ ] TAG 체인 검증 명령어가 제공된다 (`rg '@(SPEC|TEST|CODE|DOC):' -n`)
- [ ] 고아 TAG 탐지 명령어가 제공된다
- [ ] 각 명령어에 간단한 설명이 포함된다

**검증 방법**:
```bash
# rg 명령어 예시가 포함되었는지 확인
rg "rg \"@" docs/guides/concepts/tag-system.md
```

---

### AC-013: 실전 예시 (AUTH-001, UPLOAD-001, PAYMENT-001)
**Given**: 사용자가 TAG 체계를 실제 프로젝트에 적용하려고 한다
**When**: 실전 예시 섹션을 읽는다
**Then**:
- [ ] JWT 인증 시스템(AUTH-001) TAG 체인 예시가 제공된다
- [ ] 파일 업로드(UPLOAD-001) TAG 체인 예시가 제공된다
- [ ] 결제 시스템(PAYMENT-001) TAG 체인 예시가 제공된다
- [ ] 각 예시는 Python 또는 TypeScript로 작성된 실제 코드를 포함한다

**검증 방법**:
```bash
# 3개 예시가 모두 포함되었는지 확인
rg "(AUTH-001|UPLOAD-001|PAYMENT-001)" docs/guides/concepts/tag-system.md
```

---

### AC-014: 문제 해결 시나리오 5개 이상
**Given**: 사용자가 TAG 관련 문제에 직면한다
**When**: 문제 해결 섹션을 읽는다
**Then**:
- [ ] TAG 중복 발생 시 대응 방법이 설명된다
- [ ] 고아 TAG 발생 시 처리 방법이 설명된다
- [ ] TAG 체인 재구성 방법이 설명된다
- [ ] 5개 이상의 문제 해결 시나리오가 제공된다
- [ ] 각 시나리오는 "문제 → 진단 → 해결" 형식이다

**검증 방법**:
```bash
# 문제 해결 시나리오 개수 확인
rg "(문제|진단|해결|Problem|Diagnosis|Solution)" docs/guides/concepts/tag-system.md -i | wc -l
```

---

## 4. SPEC-First TDD 워크플로우 가이드 (spec-first-tdd.md)

### AC-015: SPEC-First TDD 개요 및 3단계 워크플로우 설명
**Given**: 사용자가 MoAI-ADK를 처음 사용한다
**When**: spec-first-tdd.md를 읽는다
**Then**:
- [ ] SPEC-First TDD의 정의가 명확히 설명된다
- [ ] "명세 없으면 코드 없다. 테스트 없으면 구현 없다" 철학이 설명된다
- [ ] Alfred SuperAgent의 역할이 설명된다
- [ ] 3단계 워크플로우(/alfred:1-spec, /alfred:2-build, /alfred:3-sync)가 각각 설명된다

**검증 방법**:
```bash
# 3단계 워크플로우가 모두 언급되었는지 확인
rg "(/alfred:1-spec|/alfred:2-build|/alfred:3-sync)" docs/guides/concepts/spec-first-tdd.md
```

---

### AC-016: 3단계 워크플로우 다이어그램 포함
**Given**: 사용자가 전체 워크플로우를 시각적으로 이해하려고 한다
**When**: 다이어그램 섹션을 본다
**Then**:
- [ ] 3단계 워크플로우를 표현한 Mermaid 다이어그램이 포함된다
- [ ] 각 단계의 입력과 출력이 명확히 표시된다
- [ ] 반복 사이클(1-spec → 2-build → 3-sync → 1-spec)이 표현된다

**예시 다이어그램**:
```mermaid
graph TD
  A[/alfred:1-spec] --> B[SPEC 작성 + 브랜치 생성]
  B --> C[/alfred:2-build]
  C --> D[RED → GREEN → REFACTOR]
  D --> E[/alfred:3-sync]
  E --> F[TAG 검증 + Living Document]
  F --> A
```

**검증 방법**:
```bash
# Mermaid 다이어그램이 포함되었는지 확인
rg "mermaid" docs/guides/concepts/spec-first-tdd.md -i
```

---

### AC-017: Personal/Team 모드별 차이점 명확히 설명
**Given**: 사용자가 Personal 또는 Team 모드를 선택하려고 한다
**When**: 모드별 차이점 섹션을 읽는다
**Then**:
- [ ] Personal 모드의 특징(로컬 Git 워크플로우)이 설명된다
- [ ] Team 모드의 특징(GitHub PR 자동화)이 설명된다
- [ ] 두 모드의 차이점이 표로 정리된다

**예시 표**:
| 항목 | Personal 모드 | Team 모드 |
|------|--------------|-----------|
| 브랜치 전략 | 로컬 브랜치 | feature/SPEC-XXX |
| PR 생성 | 수동 | 자동 (Draft PR) |
| 동기화 | 로컬 머지 | PR 머지 |

**검증 방법**:
```bash
# Personal/Team 모드가 모두 언급되었는지 확인
rg "(Personal.*모드|Team.*모드)" docs/guides/concepts/spec-first-tdd.md
```

---

### AC-018: 전체 사이클 실습 가능한 예제 제공 (TODO App)
**Given**: 사용자가 3단계 워크플로우를 직접 실습하려고 한다
**When**: 실전 예제 섹션을 읽는다
**Then**:
- [ ] TODO App 기능 추가 전체 사이클 예제가 제공된다
- [ ] 각 단계별 상세 설명이 포함된다
- [ ] 각 커밋별 변경 내역이 명시된다
- [ ] 사용자가 따라 할 수 있는 명령어가 제공된다

**검증 방법**:
```bash
# TODO App 예제가 포함되었는지 확인
rg "TODO.*App" docs/guides/concepts/spec-first-tdd.md -i
```

---

### AC-019: 문제 해결 시나리오 3개 이상
**Given**: 사용자가 워크플로우 진행 중 문제에 직면한다
**When**: 문제 해결 섹션을 읽는다
**Then**:
- [ ] 테스트 실패 시 대응 방법이 설명된다
- [ ] TAG 체인 끊김 해결 방법이 설명된다
- [ ] PR 충돌 해결 방법이 설명된다
- [ ] 각 시나리오는 "문제 → 진단 → 해결" 형식이다

**검증 방법**:
```bash
# 문제 해결 시나리오 개수 확인
rg "(테스트.*실패|TAG.*체인|PR.*충돌)" docs/guides/concepts/spec-first-tdd.md -i | wc -l
```

---

### AC-020: Quick Start 가이드 링크 포함
**Given**: 사용자가 빠르게 시작하려고 한다
**When**: Quick Start 섹션을 본다
**Then**:
- [ ] Quick Start 가이드로의 링크가 제공된다
- [ ] 링크는 실제 동작하는 유효한 링크이다

**검증 방법**:
```bash
# Quick Start 링크가 포함되었는지 확인
rg "Quick.*Start|빠른.*시작" docs/guides/concepts/spec-first-tdd.md -i
```

---

## 5. 통합 검증

### AC-021: 문서 간 상호 참조 링크 정상 동작
**Given**: 4개 문서가 모두 작성되었다
**When**: 상호 참조 링크를 클릭한다
**Then**:
- [ ] 모든 상호 참조 링크가 정상 동작한다
- [ ] 깨진 링크가 없다
- [ ] 상대 경로가 올바르게 설정되어 있다

**검증 방법**:
```bash
# 링크 유효성 검증 스크립트 실행
# (추후 CI/CD 파이프라인에 통합 예정)
```

---

### AC-022: 용어 통일성 확인
**Given**: 4개 문서가 모두 작성되었다
**When**: 용어 사용을 검토한다
**Then**:
- [ ] 같은 개념은 같은 용어로 일관되게 표현된다
- [ ] 기술 용어는 영어 원문을 병기한다 (예: "테스트 주도 개발(Test-Driven Development)")
- [ ] 약어는 최초 언급 시 풀어서 설명한다 (예: "TDD(Test-Driven Development)")

**검증 방법**:
```bash
# 주요 용어가 일관되게 사용되었는지 확인
rg "(TDD|EARS|TRUST|TAG)" docs/guides/concepts/*.md
```

---

### AC-023: 예시 코드 실행 가능성 확인
**Given**: 4개 문서에 예시 코드가 포함되었다
**When**: 예시 코드를 실행한다
**Then**:
- [ ] 모든 예시 코드가 실제 실행 가능하다
- [ ] 예시 코드에 명확한 주석이 포함된다
- [ ] 예시 코드는 포맷터로 정리되어 있다

**검증 방법**:
```bash
# 예시 코드 블록 추출 및 실행 테스트
# (추후 CI/CD 파이프라인에 통합 예정)
```

---

### AC-024: README.md 업데이트 완료
**Given**: 4개 문서가 모두 작성되었다
**When**: 프로젝트 루트의 README.md를 확인한다
**Then**:
- [ ] README.md에 "핵심 개념 가이드" 섹션이 추가되어 있다
- [ ] 4개 문서로의 링크가 포함되어 있다
- [ ] 각 문서의 간단한 설명이 포함되어 있다

**예시 형식**:
```markdown
## 핵심 개념 가이드
- [EARS 요구사항 작성 가이드](docs/guides/concepts/ears-guide.md) - 명확하고 구조화된 요구사항 작성 방법
- [TRUST 5원칙 가이드](docs/guides/concepts/trust-principles.md) - AI 시대의 코드 품질 보장 원칙
- [TAG 시스템 가이드](docs/guides/concepts/tag-system.md) - 코드 추적성 확보 방법
- [SPEC-First TDD 워크플로우](docs/guides/concepts/spec-first-tdd.md) - 3단계 개발 워크플로우
```

**검증 방법**:
```bash
# README.md에 링크가 포함되었는지 확인
rg "docs/guides/concepts/(ears-guide|trust-principles|tag-system|spec-first-tdd).md" README.md
```

---

### AC-025: docs/README.md 생성 완료
**Given**: 4개 문서가 모두 작성되었다
**When**: docs/README.md를 확인한다
**Then**:
- [ ] docs/README.md가 생성되어 있다
- [ ] 전체 문서 인덱스가 제공된다
- [ ] 문서 카테고리별로 정리되어 있다 (개념, 튜토리얼, 고급)

**검증 방법**:
```bash
# docs/README.md 존재 확인
test -f docs/README.md && echo "docs/README.md 존재" || echo "docs/README.md 없음"
```

---

### AC-026: TAG 체인 검증 완료
**Given**: SPEC-DOCS-002가 구현되었다
**When**: TAG 체인 검증을 수행한다
**Then**:
- [ ] @SPEC:DOCS-002가 존재한다
- [ ] @TEST:DOCS-002가 존재한다 (문서 검증 테스트)
- [ ] @CODE:DOCS-002가 존재한다 (문서 작성)
- [ ] @DOC:DOCS-002가 존재한다 (최종 문서)
- [ ] 고아 TAG가 없다

**검증 방법**:
```bash
# TAG 체인 검증
rg "@(SPEC|TEST|CODE|DOC):DOCS-002" -n .moai/specs/ tests/ src/ docs/
```

---

### AC-027: TRUST 원칙 준수 확인
**Given**: SPEC-DOCS-002가 구현되었다
**When**: TRUST 체크리스트를 확인한다
**Then**:
- [ ] **T** - 문서 검증 테스트가 작성되었다
- [ ] **R** - 각 문서가 3,000자 이내로 가독성이 확보되었다
- [ ] **U** - 문서 구조가 일관되게 유지되었다
- [ ] **S** - 보안 관련 내용이 적절히 설명되었다 (TRUST의 S)
- [ ] **T** - TAG 체인이 완전하다

**검증 방법**:
```bash
# 문서 길이 확인
wc -m docs/guides/concepts/*.md
```

---

### AC-028: 사용자 피드백 반영 완료
**Given**: 초기 버전 문서가 작성되었다
**When**: 사용자 피드백을 수렴한다
**Then**:
- [ ] 최소 2명 이상의 사용자 피드백이 수집되었다
- [ ] 피드백 내용이 반영되었다
- [ ] 반영 내역이 HISTORY 섹션에 기록되었다

**검증 방법**:
```bash
# HISTORY 섹션 확인
rg "## HISTORY" .moai/specs/SPEC-DOCS-002/spec.md -A 10
```

---

## ✅ 전체 완료 조건 (Definition of Done)

### 기능 완료
- [ ] AC-001 ~ AC-020: 4개 문서 작성 완료 (각 문서별 수락 기준 충족)
- [ ] AC-021 ~ AC-025: 통합 검증 완료 (상호 참조, 용어 통일, README 업데이트)
- [ ] AC-026 ~ AC-028: 품질 검증 완료 (TAG 체인, TRUST 원칙, 피드백 반영)

### 품질 기준
- [ ] markdownlint 검증 통과
- [ ] 링크 유효성 검증 통과
- [ ] 예시 코드 실행 가능성 확인 완료

### 문서화
- [ ] SPEC-DOCS-002 spec.md v0.0.1 작성 완료
- [ ] SPEC-DOCS-002 plan.md v0.0.1 작성 완료
- [ ] SPEC-DOCS-002 acceptance.md v0.0.1 작성 완료
- [ ] README.md 및 docs/README.md 업데이트 완료

### 검증
- [ ] TAG 체인 검증 완료 (@SPEC:DOCS-002 → @TEST:DOCS-002 → @CODE:DOCS-002 → @DOC:DOCS-002)
- [ ] 고아 TAG 없음
- [ ] TRUST 원칙 준수 확인

### 배포 준비
- [ ] /alfred:3-sync 실행 완료
- [ ] SPEC-DOCS-002 v0.1.0으로 버전 업데이트 (TDD 구현 완료 시)
- [ ] PR Ready 전환 (Team 모드)

---

**작성일**: 2025-10-11
**작성자**: @Goos
**버전**: 0.0.1 (draft)
