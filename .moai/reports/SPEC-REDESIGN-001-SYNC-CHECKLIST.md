# SPEC-REDESIGN-001: 동기화 체크리스트

**작성일**: 2025-11-19
**목표**: 모든 문서 동기화 항목 추적

---

## Phase 1: SPEC 문서 완성 (필수)

### 작업 1.1: spec.md 복구
- [ ] DELIVERABLES.md에서 SPEC 요구사항 추출
- [ ] implementation_progress.md에서 수락 기준 추출
- [ ] EARS 형식으로 spec.md 작성
- [ ] 파일: `.moai/specs/SPEC-REDESIGN-001/spec.md`
- [ ] 예상 라인: 300-400줄
- [ ] 검증: 13개 수락 기준 모두 포함

### 작업 1.2: 상호 참조 정리
- [ ] spec.md → DELIVERABLES.md 링크 추가
- [ ] DELIVERABLES.md → implementation_progress.md 링크 추가
- [ ] implementation_progress.md → tdd_cycle_summary.md 링크 추가
- [ ] 모든 문서에 상호 참조 링크 검증
- [ ] 깨진 링크 수정

---

## Phase 2: 프로젝트 레벨 문서 (필수)

### 작업 2.1: README.md 업데이트

#### 섹션 1: Project Configuration System
- [ ] 3탭 구조 설명 추가
  - Tab 1: Quick Start (2-3분)
  - Tab 2: Documentation (5분)
  - Tab 3: Git Automation (3분)
- [ ] 10개 필수 질문 설명
- [ ] 63% 질문 감소 강조

#### 섹션 2: Smart Defaults & Auto-Detection
- [ ] SmartDefaultsEngine 설명 (16개 기본값)
- [ ] AutoDetectionEngine 설명 (5개 필드)
- [ ] 자동 감지 필드 목록
  - 프로젝트 언어
  - 로케일
  - 언어명
  - 템플릿 버전
  - MoAI 버전

#### 섹션 3: Documentation Generation
- [ ] DocumentationGenerator 기능 설명
- [ ] 3가지 문서 타입 (product/structure/tech)
- [ ] BrainstormQuestionGenerator 설명
- [ ] AgentContextInjector 설명

#### 섹션 4: Module Reference
- [ ] moai_adk.project.schema API
- [ ] moai_adk.project.configuration API
- [ ] moai_adk.project.documentation API
- [ ] 사용 예제 코드 포함

- [ ] 파일: `README.md`
- [ ] 예상 추가 라인: 800-1000줄
- [ ] 검증: 모든 신규 모듈 포함

### 작업 2.2: CHANGELOG.md 업데이트

#### v0.26.0 - Configuration Redesign 섹션
- [ ] Feat: Project Configuration System v3.0.0
- [ ] Feat: Tab-based Schema with 63% question reduction
- [ ] Feat: Smart Defaults Engine (16 defaults)
- [ ] Feat: Auto-Detection System (5 fields)
- [ ] Feat: Documentation Generation System
- [ ] Feat: Configuration Migrator (v2.1.0 → v3.0.0)
- [ ] Fix: 9개 실패 테스트 해결
- [ ] 버그 수정 항목 추가

- [ ] 파일: `CHANGELOG.md`
- [ ] v0.26.0 섹션 추가

### 작업 2.3: README.ko.md 동기화

- [ ] README.md 신규 섹션 한국어 번역
- [ ] 기술 용어 일관성 검증
- [ ] 한국 사용자 가이드 추가
- [ ] 파일: `README.ko.md`
- [ ] 예상 추가 라인: 800-1000줄

---

## Phase 3: API 문서 생성 (필수)

### 작업 3.1: configuration-api.md

**파일**: `.moai/docs/api/configuration-api.md`

#### ConfigurationManager
- [ ] load_from_file() 메서드 문서화
- [ ] save_to_file() 메서드 문서화
- [ ] build_from_responses() 메서드 문서화
- [ ] validate() 메서드 문서화
- [ ] 사용 예제 추가

#### SmartDefaultsEngine
- [ ] apply_defaults() 메서드
- [ ] get_defaults_for_mode() 메서드
- [ ] 16개 기본값 목록
- [ ] 사용 예제

#### AutoDetectionEngine
- [ ] detect_project_language() 메서드
- [ ] detect_locale() 메서드
- [ ] detect_language_name() 메서드
- [ ] detect_template_version() 메서드
- [ ] 지원 언어 목록

#### ConfigurationCoverageValidator
- [ ] validate_coverage() 메서드
- [ ] get_coverage_report() 메서드

#### ConfigurationMigrator
- [ ] migrate_v2_to_v3() 메서드
- [ ] 필드 매핑 테이블

#### TabSchemaValidator
- [ ] 스키마 검증 규칙
- [ ] 에러 메시지 목록

#### 기타 클래스
- [ ] ConditionalBatchRenderer
- [ ] TemplateVariableInterpolator

### 작업 3.2: schema-api.md

**파일**: `.moai/docs/api/schema-api.md`

- [ ] load_tab_schema() 함수 문서화
- [ ] Tab 구조 정의
- [ ] Batch 구조 정의
- [ ] Question 포맷 정의
- [ ] 예제 스키마 JSON
- [ ] 조건부 배치 (show_if) 설명
- [ ] AskUserQuestion API 호환성

### 작업 3.3: documentation-generator-api.md

**파일**: `.moai/docs/api/documentation-generator-api.md`

#### DocumentationGenerator
- [ ] generate_product_md() 메서드
- [ ] generate_structure_md() 메서드
- [ ] generate_tech_md() 메서드
- [ ] save_documents() 메서드
- [ ] load_documents() 메서드

#### BrainstormQuestionGenerator
- [ ] quick_questions() 메서드
- [ ] standard_questions() 메서드
- [ ] deep_questions() 메서드
- [ ] get_questions_by_depth() 메서드

#### AgentContextInjector
- [ ] inject_to_project_manager() 메서드
- [ ] inject_to_tdd_implementer() 메서드
- [ ] inject_to_domain_experts() 메서드

---

## Phase 4: 사용 가이드 (중요)

### 작업 4.1: configuration-v3-setup.md

**파일**: `.moai/docs/guides/configuration-v3-setup.md`

- [ ] Quick Start 가이드 (5분)
- [ ] 3탭 구조 상세 설명
- [ ] Tab 1: Quick Start 설명 (10개 질문)
- [ ] Tab 2: Documentation 설명
- [ ] Tab 3: Git Automation 설명
- [ ] 스마트 기본값 설명
- [ ] 자동 감지 필드 설명
- [ ] 단계별 설정 예제
- [ ] FAQ 섹션
- [ ] 트러블슈팅

### 작업 4.2: tab-schema-usage.md

**파일**: `.moai/docs/guides/tab-schema-usage.md`

- [ ] Schema 구조 개요
- [ ] Tab 정의 방법
- [ ] Batch 정의 방법
- [ ] Question 포맷 설명
- [ ] Option 설정 방법
- [ ] 조건부 배치 (show_if) 사용법
- [ ] 템플릿 변수 보간 설명
- [ ] 예제 스키마 JSON
- [ ] AskUserQuestion API 호환성 규칙
- [ ] 검증 규칙 설명

---

## Phase 5: 마이그레이션 문서 (필수)

### 작업 5.1: config-v2-to-v3.md

**파일**: `.moai/docs/migration/config-v2-to-v3.md`

- [ ] 변경 사항 요약
  - 질문 수: 27 → 10 (63% 감소)
  - 탭 구조: 1탭 → 3탭
  - 스마트 기본값 추가 (16개)
  - 자동 감지 기능 추가 (5개)

- [ ] 자동 마이그레이션 프로세스
  - ConfigurationMigrator 사용법
  - migrate_v2_to_v3() 함수 설명
  - 마이그레이션 결과 검증

- [ ] 필드 매핑 테이블
  - v2.1.0 필드 → v3.0.0 필드 매핑
  - 삭제된 필드 목록
  - 신규 필드 목록

- [ ] 스마트 기본값 적용
  - 자동 적용 기본값 설명
  - 16개 기본값 상세 설명
  - 커스터마이징 방법

- [ ] 수동 마이그레이션 (필요시)
  - 단계별 가이드
  - 검증 체크리스트
  - 복구 방법

---

## Phase 6: 프로젝트 상태 문서 (선택)

### 작업 6.1: SPEC-REDESIGN-001 완료 리포트

**파일**: `.moai/reports/SPEC-REDESIGN-001-COMPLETION.md`

- [ ] 최종 상태 요약
- [ ] 달성한 기준
- [ ] 테스트 결과
- [ ] 코드 커버리지
- [ ] 리뷰 요청 항목
- [ ] 다음 단계
- [ ] 리리스 노트

---

## 검증 및 품질 보장

### TRUST 5 원칙

#### Trackable (추적 가능)
- [ ] 모든 문서에 날짜 추가
- [ ] 버전 정보 포함
- [ ] 상호 참조 링크 검증
- [ ] TAG 추적 기능 테스트

#### Readable (가독성)
- [ ] 명확한 제목 구조
- [ ] 목차 포함
- [ ] 코드 예제 포함
- [ ] 시각적 표 및 다이어그램
- [ ] 명확한 섹션 구분

#### Unified (통일성)
- [ ] 문서 간 용어 일관성
- [ ] Markdown 포맷 통일
- [ ] 링크 포맷 통일
- [ ] 코드 스타일 통일
- [ ] 예제 패턴 일관성

#### Secured (보안)
- [ ] 민감 정보 제외 (API 키, 비밀번호)
- [ ] 보안 가이드라인 포함
- [ ] 인증/인가 정보 문서화

#### Tested (테스트)
- [ ] 모든 코드 예제 실행 검증
- [ ] 링크 유효성 검사 (URL, 파일 경로)
- [ ] 문법 검사
- [ ] 오래된 정보 업데이트

### 최종 검증

- [ ] 모든 파일 생성 확인
- [ ] 파일 크기 검증
- [ ] Markdown 문법 검증
- [ ] 링크 유효성 검사
- [ ] 이미지/표 렌더링 확인
- [ ] 코드 블록 문법 하이라이트
- [ ] 한국어 텍스트 인코딩 확인

---

## 진행 상황 추적

### Timeline

| Phase | 작업 | 상태 | 예상시간 | 실제시간 | 비고 |
|-------|------|------|---------|---------|------|
| 1 | spec.md 복구 | ⏳ | 30분 | | |
| 1 | 상호참조 정리 | ⏳ | 20분 | | |
| 2 | README.md | ⏳ | 1시간 | | |
| 2 | CHANGELOG.md | ⏳ | 30분 | | |
| 2 | README.ko.md | ⏳ | 30분 | | |
| 3 | configuration-api.md | ⏳ | 1시간 | | |
| 3 | schema-api.md | ⏳ | 45분 | | |
| 3 | documentation-generator-api.md | ⏳ | 45분 | | |
| 4 | configuration-v3-setup.md | ⏳ | 1시간 | | |
| 4 | tab-schema-usage.md | ⏳ | 1시간 | | |
| 5 | config-v2-to-v3.md | ⏳ | 45분 | | |
| 6 | COMPLETION.md | ⏳ | 1시간 | | |

---

## 참고 자료

### 기존 SPEC 문서
- `.moai/specs/SPEC-REDESIGN-001/DELIVERABLES.md` (356줄)
- `.moai/specs/SPEC-REDESIGN-001/implementation_progress.md` (299줄)
- `.moai/specs/SPEC-REDESIGN-001/tdd_cycle_summary.md` (393줄)

### 소스 코드
- `src/moai_adk/project/schema.py` (234줄, 100% 커버리지)
- `src/moai_adk/project/configuration.py` (1,001줄, 77.74% 커버리지)
- `src/moai_adk/project/documentation.py` (566줄, 58.10% 커버리지)

### 테스트
- `tests/test_spec_redesign_001_configuration_schema.py` (919줄)
  - 51개 통과 (85%)
  - 9개 실패 (15%)

---

**체크리스트 작성**: 2025-11-19
**상태**: 동기화 작업 준비 완료
**다음 단계**: Phase별 작업 순서대로 진행
