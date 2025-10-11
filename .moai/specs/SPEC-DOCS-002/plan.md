# SPEC-DOCS-002 구현 계획

> **SPEC ID**: DOCS-002
> **제목**: MoAI-ADK 핵심 개념 문서
> **버전**: 0.0.1 (draft)
> **작성자**: @Goos

---

## 📋 개요

### 목표
MoAI-ADK의 4가지 핵심 개념(EARS, TRUST, TAG, SPEC-First TDD)을 사용자가 명확히 이해하고 실무에 적용할 수 있도록 체계적인 문서를 제공한다.

### 범위
- EARS 요구사항 작성 가이드
- TRUST 5원칙 상세 설명
- TAG 시스템 사용 가이드
- SPEC-First TDD 워크플로우 문서

### 비범위 (Out of Scope)
- 고급 기능 가이드 (별도 SPEC-DOCS-003)
- 언어별 상세 튜토리얼 (별도 SPEC)
- API 레퍼런스 문서 (별도 SPEC)

---

## 🎯 우선순위별 마일스톤

### 1차 목표: 문서 구조 및 뼈대 작성
**의존성**: SPEC-DOCS-002 승인 완료

#### 작업 항목
1. **디렉토리 구조 생성**
   - `docs/guides/concepts/` 디렉토리 생성
   - 4개 문서 파일 생성 (ears-guide.md, trust-principles.md, tag-system.md, spec-first-tdd.md)

2. **각 문서 뼈대 작성**
   - 목차 구조 정의
   - 섹션별 개요 작성
   - 상호 참조 링크 설정

#### 산출물
- [ ] `docs/guides/concepts/ears-guide.md` (목차)
- [ ] `docs/guides/concepts/trust-principles.md` (목차)
- [ ] `docs/guides/concepts/tag-system.md` (목차)
- [ ] `docs/guides/concepts/spec-first-tdd.md` (목차)

#### 완료 기준
- 4개 문서 파일이 생성되고 기본 구조가 작성되어 있다
- 각 문서 간 상호 참조 링크가 정의되어 있다

---

### 2차 목표: EARS 가이드 작성 완료
**의존성**: 1차 목표 완료

#### 작업 항목
1. **EARS 개념 설명**
   - 정의, 목적, 장점 작성
   - 5가지 구문 유형별 상세 설명

2. **실제 예시 작성**
   - 인증 시스템 (AUTH-001) 예시
   - 파일 업로드 (UPLOAD-001) 예시
   - 결제 시스템 (PAYMENT-001) 예시

3. **안티 패턴 및 개선안**
   - 모호한 요구사항 → 명확한 요구사항 전환 예시
   - 측정 불가능 → 측정 가능 전환 예시

4. **베스트 프랙티스**
   - 측정 가능한 기준 명시 방법
   - 구체적인 액션 동사 사용 가이드

#### 산출물
- [ ] EARS 개념 설명 (500~800자)
- [ ] 5가지 구문별 예시 3개씩 (총 15개)
- [ ] 안티 패턴 → 개선안 5개
- [ ] 실습 가능한 템플릿

#### 완료 기준
- EARS 가이드가 3,000자 이내로 완성되어 있다
- 모든 예시가 실제 동작 가능한 내용이다
- 다른 문서와의 상호 참조가 정확하다

---

### 3차 목표: TRUST 원칙 가이드 작성 완료
**의존성**: 2차 목표 완료

#### 작업 항목
1. **TRUST 개요**
   - 정의 및 목적 작성
   - AI 시대의 코드 품질 보장 필요성 설명

2. **각 원칙별 상세 설명**
   - Test First (T): SPEC → Test → Code 사이클
   - Readable (R): 가독성 기준 및 도구
   - Unified (U): 아키텍처 통합
   - Secured (S): 보안 요구사항
   - Trackable (T): TAG 시스템 연계

3. **언어별 구현 예시**
   - Python: pytest, ruff, mypy, bandit
   - TypeScript: Vitest, Biome, npm audit
   - Go: go test, gofmt, gosec

4. **TRUST 체크리스트**
   - 코드 리뷰 시 확인 항목
   - 자동화 도구 통합 가이드

#### 산출물
- [ ] TRUST 개요 (300~500자)
- [ ] 5원칙별 상세 설명 (각 400~600자)
- [ ] 언어별 구현 예시 (Python, TypeScript, Go 필수)
- [ ] TRUST 체크리스트 템플릿

#### 완료 기준
- TRUST 가이드가 3,000자 이내로 완성되어 있다
- 각 원칙별 언어별 예시가 포함되어 있다
- 실제 코드 리뷰 시나리오가 포함되어 있다

---

### 4차 목표: TAG 시스템 가이드 작성 완료
**의존성**: 3차 목표 완료

#### 작업 항목
1. **TAG 시스템 개요**
   - CODE-FIRST 철학 설명
   - TAG 라이프사이클 다이어그램
   - TAG ID 규칙 설명

2. **TAG 유형별 사용법**
   - @SPEC, @TEST, @CODE, @DOC 각각의 역할
   - TAG 서브 카테고리 (API, UI, DATA, DOMAIN, INFRA)

3. **TAG 검증 방법**
   - rg 명령어 치트 시트
   - 중복 확인 방법
   - 고아 TAG 탐지 방법

4. **실전 예시**
   - JWT 인증 시스템 (AUTH-001) TAG 체인
   - 실제 코드 예시 (Python, TypeScript)

5. **문제 해결**
   - TAG 중복 발생 시 대응
   - 고아 TAG 처리
   - TAG 체인 재구성

#### 산출물
- [ ] TAG 시스템 개요 (400~600자)
- [ ] TAG 체계도 (다이어그램)
- [ ] rg 명령어 치트 시트
- [ ] 실전 예시 3개 (AUTH-001, UPLOAD-001, PAYMENT-001)
- [ ] 문제 해결 시나리오 5개

#### 완료 기준
- TAG 가이드가 3,000자 이내로 완성되어 있다
- TAG 체계도가 명확하게 표현되어 있다
- 실제 코드 예시가 Python, TypeScript 모두 포함되어 있다

---

### 5차 목표: SPEC-First TDD 워크플로우 가이드 작성 완료
**의존성**: 4차 목표 완료

#### 작업 항목
1. **SPEC-First TDD 개요**
   - 정의 및 철학 설명
   - Alfred SuperAgent 역할 설명

2. **3단계 워크플로우 상세 설명**
   - `/alfred:1-spec`: SPEC 작성 단계
   - `/alfred:2-build`: TDD 구현 단계
   - `/alfred:3-sync`: 문서 동기화 단계

3. **모드별 차이점**
   - Personal 모드: 로컬 Git 워크플로우
   - Team 모드: GitHub PR 자동화

4. **실전 예제**
   - TODO App 기능 추가 전체 사이클 실습
   - 단계별 상세 설명
   - 각 커밋별 변경 내역

5. **베스트 프랙티스**
   - SPEC 작성 시 주의사항
   - TDD 사이클 팁
   - 문서 동기화 타이밍

6. **문제 해결**
   - 테스트 실패 시 대응
   - TAG 체인 끊김 해결
   - PR 충돌 해결

#### 산출물
- [ ] SPEC-First TDD 개요 (300~500자)
- [ ] 3단계 워크플로우 다이어그램
- [ ] Personal/Team 모드별 차이점 설명
- [ ] 전체 사이클 실습 예제 (TODO App)
- [ ] 문제 해결 시나리오 3개
- [ ] Quick Start 가이드 링크

#### 완료 기준
- SPEC-First TDD 가이드가 3,000자 이내로 완성되어 있다
- 3단계 워크플로우 다이어그램이 명확하게 표현되어 있다
- 실전 예제가 실습 가능한 상태로 작성되어 있다

---

### 최종 목표: 통합 검증 및 배포
**의존성**: 5차 목표 완료

#### 작업 항목
1. **문서 간 일관성 검증**
   - 상호 참조 링크 확인
   - 용어 통일성 확인
   - 예시 코드 동작 확인

2. **사용자 테스트**
   - 실제 사용자 피드백 수렴
   - 개선 사항 반영

3. **배포 준비**
   - README.md 업데이트 (개념 가이드 링크 추가)
   - docs/README.md 생성 (전체 문서 인덱스)

#### 산출물
- [ ] 문서 간 일관성 검증 보고서
- [ ] 사용자 피드백 반영 내역
- [ ] README.md 업데이트
- [ ] docs/README.md 생성

#### 완료 기준
- 모든 상호 참조 링크가 정상 동작한다
- 사용자가 4개 문서를 통해 핵심 개념을 이해할 수 있다
- README.md에 개념 가이드 링크가 추가되어 있다

---

## 🛠️ 기술적 접근 방법

### 문서 작성 도구
- **형식**: Markdown (.md)
- **린터**: markdownlint
- **버전 관리**: Git

### 다이어그램 작성
- **도구**: Mermaid (Markdown 내 코드 블록)
- **형식**: flowchart, sequence diagram, class diagram
- **예시**:
  ```mermaid
  graph TD
    A[@SPEC:ID] --> B[@TEST:ID]
    B --> C[@CODE:ID]
    C --> D[@DOC:ID]
  ```

### 예시 코드 작성
- **언어**: Python, TypeScript, Go (우선순위)
- **포맷터**: black (Python), Biome (TypeScript), gofmt (Go)
- **실행 가능성**: 모든 예시 코드는 실제 동작 가능해야 함

### 상호 참조 링크
- **형식**: `[링크 텍스트](../path/to/document.md)`
- **검증**: 빌드 시 링크 유효성 자동 확인

---

## 📊 아키텍처 설계 방향

### 문서 구조
```
docs/
├── README.md (전체 문서 인덱스)
├── guides/
│   ├── concepts/ (핵심 개념 - SPEC-DOCS-002)
│   │   ├── ears-guide.md
│   │   ├── trust-principles.md
│   │   ├── tag-system.md
│   │   └── spec-first-tdd.md
│   ├── tutorials/ (실습 가이드 - SPEC-DOCS-001)
│   │   └── todo-app-fullstack.md
│   └── advanced/ (고급 가이드 - 미래 SPEC)
└── api/ (API 레퍼런스 - 미래 SPEC)
```

### 상호 참조 관계
```
CLAUDE.md
  ↓
development-guide.md
  ↓
docs/guides/concepts/
  ├── ears-guide.md ← spec-first-tdd.md
  ├── trust-principles.md ← spec-first-tdd.md
  ├── tag-system.md ← spec-first-tdd.md
  └── spec-first-tdd.md (중심 문서)
```

### 설계 원칙
1. **Single Source of Truth**: 각 개념은 하나의 문서에서만 상세히 설명
2. **Cross-Reference First**: 중복 설명 대신 상호 참조 링크 사용
3. **Example-Driven**: 모든 개념은 실습 가능한 예시와 함께 제공
4. **Progressive Disclosure**: 기본 → 고급으로 단계적 학습 가능

---

## ⚠️ 리스크 및 대응 방안

### 리스크 1: 문서 길이 초과
**위험도**: Medium
**영향**: 사용자가 문서를 읽기 어려워짐

**대응 방안**:
- 각 문서를 3,000자 이내로 제한
- 초과 시 별도 "고급 가이드" 문서로 분리
- 목차 및 앵커 링크 활용으로 빠른 탐색 지원

### 리스크 2: 예시 코드 오류
**위험도**: High
**영향**: 사용자 신뢰도 하락, 학습 방해

**대응 방안**:
- 모든 예시 코드는 실제 실행 후 검증
- TDD 구현 시 예시 코드도 테스트 케이스 작성
- CI/CD 파이프라인에 문서 내 코드 블록 실행 검증 추가

### 리스크 3: 상호 참조 링크 깨짐
**위험도**: Medium
**영향**: 사용자 탐색 경험 저하

**대응 방안**:
- 빌드 시 링크 유효성 자동 검증 스크립트 실행
- 문서 구조 변경 시 링크 일괄 업데이트
- 상대 경로 사용으로 디렉토리 이동 대응

### 리스크 4: 용어 불일치
**위험도**: Low
**영향**: 사용자 혼란

**대응 방안**:
- 용어집(Glossary) 작성 및 참조
- 문서 리뷰 시 용어 통일성 확인
- markdownlint 커스텀 룰로 일관성 검증

---

## 📚 참조 문서

### 프로젝트 문서
- `CLAUDE.md`: 전체 프레임워크 개요
- `development-guide.md`: 개발 가이드
- `product.md`: 제품 정의

### 관련 SPEC
- `SPEC-DOCS-001`: TODO App 풀스택 예제 (선행 작업)
- `SPEC-PROJECT-001`: MoAI-ADK 프로젝트 정의

### 외부 참조
- EARS 방법론: [Easy Approach to Requirements Syntax](https://www.iaria.org/conferences2012/filesICCGI12/Tutorial%20ICCGI%202012%20(EARS).pdf)
- TDD 베스트 프랙티스: Kent Beck, "Test-Driven Development: By Example"
- CODE-FIRST 원칙: 코드가 진실의 유일한 원천(Single Source of Truth)

---

## ✅ 완료 조건 (Definition of Done)

### 기능 완료
- [ ] 4개 문서(ears-guide.md, trust-principles.md, tag-system.md, spec-first-tdd.md) 작성 완료
- [ ] 각 문서가 3,000자 이내
- [ ] 모든 예시 코드가 실행 가능
- [ ] 상호 참조 링크가 정상 동작
- [ ] 다이어그램이 명확히 표현됨

### 품질 기준
- [ ] markdownlint 검증 통과
- [ ] 용어 통일성 확인 완료
- [ ] 사용자 피드백 반영 완료

### 문서화
- [ ] README.md에 개념 가이드 링크 추가
- [ ] docs/README.md 생성 (전체 문서 인덱스)
- [ ] SPEC-DOCS-002 문서 동기화 완료 (v0.1.0)

### 검증
- [ ] TAG 체인 검증 완료 (@SPEC:DOCS-002 → @TEST:DOCS-002 → @CODE:DOCS-002 → @DOC:DOCS-002)
- [ ] 고아 TAG 없음
- [ ] TRUST 원칙 준수 확인

---

**작성일**: 2025-10-11
**작성자**: @Goos
**버전**: 0.0.1 (draft)
