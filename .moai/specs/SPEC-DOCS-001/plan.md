---
id: DOCS-001
version: 1.0.0
status: draft
created: 2025-01-06
updated: 2025-01-06
author: @Goos
priority: high
category: documentation
phase: planning
related_specs:
  - SPEC-INSTALL-001
  - SPEC-INIT-001
  - SPEC-CONFIG-001
traceability:
  spec: "@SPEC:DOCS-001"
  test: "@TEST:DOCS-001"
  code: "@CODE:DOCS-001"
---

# `@PLAN:DOCS-001: Document-master 에이전트 온라인 문서 구현 계획`

## 요약 (Summary)

SPEC-DOCS-001 v2.0.0의 요구사항을 만족시키기 위한 Document-master 에이전트 기반 온라인 문서 생성 시스템 구현 계획입니다. Nextra 프레임워크와 Context7 통합을 통해 자동화된 문서 사이트를 구축하여 169개 Python 파일과 55개 Skills, 19개 에이전트를 종합적으로 문서화합니다.

## 목표 (Objectives)

1. **문서 모듈화**: 거대한 README.ko.md를 관리 가능한 크기로 분할
2. **실제 코드 기반**: 모든 예제를 실제 `src/moai_adk/` 코드와 연동
3. **시각화 강화**: Mermaid 다이어그램으로 이해도 향상
4. **다국어 지원**: 한국어 → 영어 → 일본어 → 중국어 순차 확장

## 실행 단계 (Implementation Phases)

### Phase 1: README.ko.md 분할 (우선순위: 높음)

#### 1.1 분할 준비
- [ ] **분할 섹션 식별**
  - 빠른 시작 (라인 66-197)
  - 핵심 개념 (라인 1423-1606)
  - 워크플로우 (라인 99-131, 881-1104)
  - 예제 및 튜토리얼 (라인 1610-2296)
  - 자주 묻는 질문 (라인 3008-3019)
  - 문제 해결 (라인 2551-2874)

- [ ] **분할 스크립트 개발** (`scripts/split_readme.py`)
  ```python
  # README.ko.md를 섹션별로 분리하는 스크립트
  # 각 섹션을 별도 .md 파일로 저장
  # 링크 업데이트 및 내용 검증
  ```

- [ ] **Nextra 설정 업데이트**
  ```javascript
  // theme.config.cjs에 새 페이지 경로 추가
  // 사이드바 메뉴 구조 정의
  // 검색 인덱스 최적화
  ```

#### 1.2 문서 파일 생성
- [ ] `docs/getting-started/`
  - `installation.md` - 설치 가이드
  - `quick-start.md` - 3분 초고속 시작
  - `first-project.md` - 첫 프로젝트 생성
  - `verification.md` - 설치 확인

- [ ] `docs/concepts/`
  - `spec-first.md` - SPEC-First 개념
  - `tdd.md` - TDD 개념
  - `tag-system.md` - @TAG 시스템
  - `trust-principles.md` - TRUST 5원칙
  - `alfred-superagent.md` - Alfred 슈퍼에이전트
  - `workflow.md` - 4단계 워크플로우

- [ ] `docs/guides/`
  - `0-project.md` - 프로젝트 초기화 가이드
  - `1-plan.md` - SPEC 작성 가이드
  - `2-run.md` - TDD 구현 가이드
  - `3-sync.md` - 동기화 가이드
  - `best-practices.md` - 모범 사례

#### 1.3 원본 업데이트
- [ ] README.ko.md를 간결한 소개로 재구성
- [ ] 분할된 문서로의 링크 추가
- [ ] 배지 및 빠른 링크 유지

### Phase 2: 실제 코드 기반 예제 (우선순위: 높음)

#### 2.1 코드 예제 분석
- [ ] `src/moai_adk/` 구조 분석
  ```
  src/moai_adk/
  ├── cli/commands/      # CLI 명령어 구현
  ├── core/              # 핵심 기능
  ├── utils/             # 유틸리티 함수
  └── __init__.py        # 패키지 초기화
  ```

- [ ] 문서화할 코드 후보 선정
  - `init.py`: 프로젝트 초기화
  - `update.py`: 업데이트 기능
  - `template_engine.py`: 템플릿 처리
  - `logger.py`: 로깅 설정

#### 2.2 예제 문서 작성
- [ ] 실제 코드 기반 예제 생성
  ```python
  # 실제 코드에서 발췌
  from moai_adk.cli.commands.init import init_command

  # 사용 예제
  init_command("hello-world")
  ```

- [ ] 각 예제에 실행 방법 추가
  ```bash
  # 실행 명령어
  python -m moai_adk init hello-world

  # 검증 방법
  ls -la hello-world/.moai/
  ```

- [ ] `@CODE:` 태그로 실제 파일 연결
  ```markdown
  <!-- 코드 예제 -->
  `@CODE:DOCS-001:INIT-EXAMPLE | SPEC: SPEC-DOCS-001`
  ```

### Phase 3: Mermaid 다이어그램 (우선순위: 중간)

#### 3.1 다이어그램 설계
- [ ] 4단계 워크플로우 다이어그램
  ```mermaid
  graph TD
      Start([사용자 요청]) --> Project[0.Project Init]
      Project --> Plan[1.Plan & SPEC]
      Plan --> Run[2.Run & TDD]
      Run --> Sync[3.Sync & Docs]
      Sync --> Plan
      Sync -.-> End([릴리스])
  ```

- [ ] 에이전트 아키텍처 다이어그램
  ```mermaid
  graph BT
      Alfred[Alfred SuperAgent]
      Alfred --> SpecBuilder[spec-builder]
      Alfred --> CodeBuilder[code-builder]
      Alfred --> TestEngineer[test-engineer]
      Alfred --> DocSyncer[doc-syncer]
      Alfred --> GitManager[git-manager]
  ```

- [ ] TAG 체인 시스템 다이어그램
  ```mermaid
  graph LR
      SPEC[@SPEC:ID] --> TEST[@TEST:ID]
      TEST --> CODE[@CODE:ID]
      CODE --> DOC[@DOC:ID]
  ```

#### 3.2 다이어그램 구현
- [ ] Nextra Mermaid 플러그인 설정
- [ ] 각 다이어그램을 관련 문서에 삽입
- [ ] 대체 텍스트 및 설명 추가

### Phase 4: 표 형식 정리 (우선순위: 중간)

#### 4.1 명령어 요약표
| 명령 | 기능 | 산출물 | 시간 |
|------|------|--------|------|
| `/alfred:0-project` | 프로젝트 초기화 | 설정 파일, 문서 | 30초 |
| `/alfred:1-plan` | SPEC 작성 | `.moai/specs/` | 2-3분 |
| `/alfred:2-run` | TDD 구현 | 테스트, 코드 | 5-10분 |
| `/alfred:3-sync` | 동기화 | 문서 업데이트 | 1-2분 |

#### 4.2 에이전트 목록표
| 에이전트 | 역할 | 모델 | 전문 분야 |
|----------|------|------|----------|
| spec-builder | 명세 작성 | Sonnet | EARS, 요구사항 |
| code-builder | TDD 구현 | Sonnet | Python, 테스트 |
| test-engineer | 테스트 전략 | Sonnet | pytest, 커버리지 |
| doc-syncer | 문서 동기화 | Haiku | Markdown, Nextra |

#### 4.3 버전 히스토리표
| 버전 | 주요 기능 | 날짜 |
|------|----------|------|
| v0.18.0 | 언어 스킬 v3.0.0 | 2025-11-06 |
| v0.17.0 | 세션 분석 CLI | 2025-11-06 |
| v0.16.x | Alfred 명령어 완성 | 2025-11-03 |

### Phase 5: 다국어 지원 (우선순위: 낮음)

#### 5.1 구조 준비
- [ ] 다국어 디렉토리 구조
  ```
  docs/
  ├── ko/    # 한국어 (기본)
  ├── en/    # 영어
  ├── ja/    # 일본어
  └── zh/    # 중국어
  ```

- [ ] Nextra 국제화(i18n) 설정
- [ ] 번역 템플릿 생성

#### 5.2 번역 실행
- [ ] 1단계: 한국어 문서 완성 검증
- [ ] 2단계: 영어 번역
  - 전문 번역가 또는 AI 번역
  - 기술 용어 일관성 검증
  - 코드 예제 영문화

- [ ] 3단계: 일본어 번역
- [ ] 4단계: 중국어 번역

## 실행 계획 (Timeline)

| 주차 | 작업 | 산출물 |
|------|------|--------|
| 1주차 | Phase 1: README 분할 | 분할된 문서 파일 |
| 2주차 | Phase 2: 코드 예제 | 실제 코드 기반 예제 |
| 3주차 | Phase 3: 다이어그램 | Mermaid 다이어그램 |
| 4주차 | Phase 4: 표 정리 | 구조화된 표 |
| 5-6주차 | Phase 5: 다국어 | 영어 번역 완료 |
| 7-8주차 | 검증 및 수정 | 최종 문서 |

## 위험 요소 및 대응 (Risks & Mitigations)

### 위험 1: 내용 누락
- **위험**: README 분할 시 중요 내용 누락
- **대응**: 자동화된 검증 스크립트 개발
- **조치**: 원본과 분할 후 내용 비교 검증

### 위험 2: 코드 예제 부정확
- **위험**: 실제 코드와 문서 불일치
- **대응**: 자동화된 예제 실행 테스트
- **조치**: CI/CD 파이프라인에 테스트 추가

### 위험 3: 번역 품질
- **위험**: 기술 용어 번역 부정확
- **대응**: 전문 번역가 검수
- **조치**: 용어 사전 및 가이드라인 제공

## 필요한 리소스 (Required Resources)

### 인적 리소스
- 기술 작가 1명 (문서 작성)
- 개발자 1명 (코드 예제 검증)
- 번역가 1명 (다국어 지원)

### 기술 리소스
- Python 3.13+ 환경
- Nextra 문서 사이트
- Mermaid.js 지원
- 자동화된 검증 스크립트

### 외부 서비스
- GitHub (코드 호스팅)
- GitHub Actions (CI/CD)
- Crowdin/Lokalise (번역 관리, 선택사항)

## 성공 측정 (Success Metrics)

### 정량적 지표
- 문서 분할률: 3295줄 → 0줄 (README)
- 새 문서 수: 20+개 파일
- 코드 예제 수: 15+개 실제 예제
- 다국어 진행률: 한(100%) → 영(80%)

### 정성적 지표
- 문서 명확도: 사용자 피드백 4.5/5.0
- 예제 유용성: 실제 실행 성공률 95%
- 검색 가능성: 핵심 정보 3클릭 내 접근

## 다음 단계 (Next Steps)

1. **즉시 실행**:
   - README.ko.md 분할 스크립트 개발
   - Phase 1 시작 (빠른 시작 가이드 분할)

2. **1주 내**:
   - Phase 1 완료 및 검증
   - Phase 2 준비 (코드 예제 분석)

3. **장기 계획**:
   - 전체 Phase 완료 후 사용자 피드백 수집
   - 지속적인 개선 및 업데이트

---

## 연락 정보 (Contact)

- **담당자**: @Goos
- **리뷰어**: Alfred SuperAgent
- **관련 SPEC**: SPEC-INSTALL-001, SPEC-INIT-001