---
id: DOC-TAG-001
version: 0.0.1
status: draft
created: 2025-10-29
updated: 2025-10-29
author: @Goos
priority: high
category: Infrastructure / Quality
labels: [documentation, tags, automation]
depends_on: []
blocks: []
related_specs: []
scope: "Phase 1 of 4-phase @DOC TAG automatic generation system"
---

# @SPEC:DOC-TAG-001: @DOC 태그 자동 생성 인프라

## HISTORY

### v0.0.1 (2025-10-29)
- **INITIAL**: @DOC 태그 자동 생성 인프라 기초 구축
- **AUTHOR**: @Goos
- **SCOPE**: TAG 생성 라이브러리 (generator, inserter, mapper, parser), 검증 강화
- **CONTEXT**: MoAI-ADK에서 42개 문서에 @DOC 태그 자동 추가 필요, 단계적 4주 계획의 첫 단계

---

## Environment (환경)

### 프로젝트 컨텍스트
- **프로젝트**: MoAI-ADK v0.8.1
- **문서 현황**: 42개 문서 중 2개만 수동으로 @DOC 태그 보유
- **목표**: 전체 문서에 자동 @DOC 태그 생성 및 삽입
- **기존 시스템**: `src/moai_adk/core/tags.py`에 기본 TAG 검증 로직 존재

### 기술 환경
- **언어**: Python 3.13+
- **테스트 프레임워크**: pytest 8.4.2
- **코드 품질 도구**: ruff 0.13.1, mypy 1.8.0
- **파일 검색**: ripgrep (rg) 명령어 사용
- **문서 형식**: Markdown (.md)

### 실행 환경
- **작업 디렉토리**: `/Users/goos/MoAI/MoAI-ADK`
- **모듈 경로**: `src/moai_adk/core/tags/`
- **테스트 경로**: `tests/unit/core/tags/`
- **문서 경로**: `docs/`, `.moai/docs/`, `.claude/`, 등

---

## Assumptions (가정)

### 시스템 가정
1. **파일 시스템 접근**: 모든 마크다운 파일에 대한 읽기/쓰기 권한 보유
2. **ripgrep 사용 가능**: `rg` 명령어로 전체 코드베이스 검색 가능
3. **Git 환경**: 변경사항 추적 및 버전 관리 활성화
4. **기존 TAG 형식**: 수동 태그가 `@DOC:DOMAIN-NNN` 형식 준수

### 비즈니스 가정
1. **도메인 매핑**: 파일 경로에서 도메인 추출 가능 (예: `docs/api/` → `API`)
2. **SPEC 참조**: 대부분의 문서가 하나 이상의 SPEC과 연결 가능
3. **중복 방지 필요**: 동일 도메인 내 중복 ID 자동 탐지 및 회피
4. **단계적 적용**: Phase 1에서는 라이브러리만 구축, Phase 2에서 대량 적용

### 기술 가정
1. **마크다운 구조**: 모든 문서가 `# Header` 형식의 첫 번째 헤더 보유
2. **파일 인코딩**: 모든 마크다운 파일이 UTF-8 인코딩
3. **테스트 커버리지**: 85% 이상 달성 가능
4. **역호환성**: 기존 2개 수동 태그 문서에 영향 없음

---

## Requirements (요구사항)

### 1. Ubiquitous Requirements (보편 요구사항)

**UR-1: TAG ID 생성 규칙**
- **WHERE** 시스템이 새로운 @DOC 태그를 생성하는 모든 위치에서
- **SYSTEM SHALL** 다음 형식을 준수해야 한다: `@DOC:DOMAIN-NNN`
  - `DOMAIN`: 대문자 영문 (예: API, CLI, GUIDE)
  - `NNN`: 001부터 시작하는 3자리 숫자
- **RATIONALE**: 일관된 TAG 형식으로 검색 및 추적성 보장

**UR-2: 중복 ID 방지**
- **WHERE** TAG ID를 생성하는 모든 시점에
- **SYSTEM SHALL** ripgrep 검색으로 기존 ID 전체 스캔 후 고유 ID 할당
- **RATIONALE**: 충돌 방지 및 TAG 무결성 유지

**UR-3: 파일 무결성 보장**
- **WHERE** 마크다운 파일에 TAG를 삽입하는 모든 경우
- **SYSTEM SHALL** 원본 파일 내용 손실 없이 헤더만 수정
- **RATIONALE**: 데이터 손실 방지 및 안전한 자동화

**UR-4: 테스트 커버리지**
- **WHERE** 모든 신규 모듈 (generator, inserter, mapper, parser)
- **SYSTEM SHALL** 85% 이상의 테스트 커버리지 유지
- **RATIONALE**: 코드 품질 및 신뢰성 보장

**UR-5: 타입 안전성**
- **WHERE** 모든 함수 및 메서드
- **SYSTEM SHALL** type hints 및 mypy 검사 통과
- **RATIONALE**: 런타임 오류 사전 방지

---

### 2. Event-driven Requirements (이벤트 기반 요구사항)

**ER-1: TAG 생성 트리거**
- **WHEN** 사용자가 `generate_doc_tag(domain, existing_tags)` 호출
- **SYSTEM SHALL** 다음을 수행한다:
  1. `existing_tags` 리스트에서 동일 도메인의 최대 번호 추출
  2. 최대 번호 + 1로 새 ID 생성
  3. `@DOC:DOMAIN-NNN` 형식 문자열 반환
- **EXAMPLE**: `generate_doc_tag("API", ["@DOC:API-001"])` → `"@DOC:API-002"`

**ER-2: 마크다운 삽입 트리거**
- **WHEN** 사용자가 `insert_tag_to_markdown(file_path, tag_id, chain)` 호출
- **SYSTEM SHALL** 다음을 수행한다:
  1. 파일의 첫 번째 `#` 헤더 라인 탐지
  2. 헤더를 `# @DOC:XXX | Chain: @SPEC:YYY -> @DOC:XXX` 형식으로 변경
  3. 파일 저장 후 성공/실패 상태 반환
- **EXAMPLE**: `# API Authentication` → `# @DOC:API-001 | Chain: @SPEC:AUTH-001 -> @DOC:API-001`

**ER-3: SPEC 매핑 트리거**
- **WHEN** 사용자가 `map_spec_to_doc(file_path)` 호출
- **SYSTEM SHALL** 다음을 수행한다:
  1. 파일 경로 및 내용 분석
  2. 관련 SPEC ID 추출 (파일 내 `@SPEC:XXX` 참조 검색)
  3. SPEC ID 리스트 또는 빈 리스트 반환
- **EXAMPLE**: `docs/api/auth.md` 내 `@SPEC:AUTH-001` 참조 → `["@SPEC:AUTH-001"]`

**ER-4: SPEC ID 파싱 트리거**
- **WHEN** 사용자가 `parse_spec_id_from_content(content)` 호출
- **SYSTEM SHALL** 정규식으로 `@SPEC:DOMAIN-NNN` 패턴 추출
- **EXAMPLE**: `"See @SPEC:AUTH-001 for details"` → `["@SPEC:AUTH-001"]`

**ER-5: 검증 hook 통합**
- **WHEN** `tags.py`의 `suggest_tag_for_file(file_path)` 호출
- **SYSTEM SHALL** TagSuggestion 객체 반환:
  - `domain`: 추천 도메인
  - `next_id`: 다음 가용 ID
  - `confidence`: 신뢰도 점수 (0.0-1.0)
  - `reasoning`: 추천 근거

---

### 3. State-driven Requirements (상태 기반 요구사항)

**SR-1: 파일 존재 상태 검증**
- **WHILE** 마크다운 파일이 존재하는 동안
- **SYSTEM SHALL** 파일 읽기 전 존재 여부 확인
- **IF NOT** 파일이 존재하지 않으면 FileNotFoundError 발생

**SR-2: TAG 중복 상태 검증**
- **WHILE** 새 TAG ID를 생성하는 동안
- **SYSTEM SHALL** ripgrep으로 전체 스캔 수행
- **IF** 생성된 ID가 이미 존재하면 번호 증가 후 재시도

**SR-3: 마크다운 헤더 상태 검증**
- **WHILE** 파일에 TAG를 삽입하려는 동안
- **SYSTEM SHALL** 첫 번째 `#` 헤더 존재 여부 확인
- **IF NOT** 헤더가 없으면 파일 맨 앞에 새 헤더 생성

**SR-4: 테스트 통과 상태 유지**
- **WHILE** 개발이 진행되는 동안
- **SYSTEM SHALL** 모든 pytest 테스트가 통과 상태 유지
- **IF** 테스트 실패 시 코드 수정 또는 테스트 수정

---

### 4. Optional Requirements (선택적 요구사항)

**OR-1: 신뢰도 점수 계산**
- **IF** SPEC 매핑이 파일 경로만으로 결정 가능
- **SYSTEM MAY** 신뢰도를 1.0으로 설정
- **ELSE** 파일 내용 분석 필요 시 0.5-0.8 범위 설정

**OR-2: 배치 처리 지원**
- **IF** 미래 Phase 2에서 대량 파일 처리 요구
- **SYSTEM MAY** `generate_tags_for_directory(path)` 함수 제공
- **NOTE**: Phase 1에서는 단일 파일 처리만 구현

**OR-3: 수동 SPEC 매핑 지원**
- **IF** 자동 매핑이 실패하거나 신뢰도가 낮음
- **SYSTEM MAY** 사용자 입력으로 SPEC ID 수동 지정 허용
- **NOTE**: CLI 인터페이스는 Phase 3에서 구현

**OR-4: TAG 체인 시각화**
- **IF** 사용자가 TAG 체인 관계 파악 필요
- **SYSTEM MAY** graphviz 또는 mermaid 다이어그램 생성
- **NOTE**: Phase 4 기능으로 미래 계획

---

### 5. Unwanted Behaviors (원하지 않는 동작)

**IF** 입력 파일이 마크다운 파일이 아니면 **THE SYSTEM SHALL NOT** 처리를 진행해야 함

**IF** 태그 대상이 기존의 수동 생성 문서면 **THE SYSTEM SHALL NOT** 변경 또는 덮어쓰기를 수행해야 함

**IF** TAG ID 생성 시간이 2초를 초과하면 **THE SYSTEM SHALL NOT** 해당 제안을 사용자에게 표시해야 함

**IF** 도메인명에 소문자나 특수문자(하이픈, 언더스코어 제외)가 포함되면 **THE SYSTEM SHALL NOT** TAG를 생성해야 함

**IF** 도메인의 시퀀스 번호가 999를 초과하면 **THE SYSTEM SHALL NOT** 새로운 TAG를 할당해야 함

**IF** 테스트 환경이 격리되지 않았으면(임시 파일 미사용) **THE SYSTEM SHALL NOT** 실제 문서 파일에 접근해야 함

**IF** 의존성으로 요청된 외부 라이브러리가 표준 라이브러리로 대체 가능하면 **THE SYSTEM SHALL NOT** 외부 라이브러리 추가를 허용해야 함

---

## Traceability (추적성)

### TAG 체인
- **@SPEC:DOC-TAG-001** (본 문서) → Phase 1 구현 → 8개 파일 생성
- **@SPEC:DOC-TAG-001** → **@TEST:DOC-TAG-001** (테스트 파일)
- **@SPEC:DOC-TAG-001** → **@CODE:DOC-TAG-001** (구현 파일)
- **@SPEC:DOC-TAG-001** → **@DOC:DOC-TAG-001** (사용자 가이드, Phase 2)

### 연관 SPEC
- 없음 (Phase 1은 독립 인프라 구축)

### 연관 문서
- `.moai/specs/SPEC-DOC-TAG-001/plan.md` - 구현 계획
- `.moai/specs/SPEC-DOC-TAG-001/acceptance.md` - 수락 기준
- `.moai/memory/spec-metadata.md` - SPEC 메타데이터 표준

### 영향받는 시스템
- `src/moai_adk/core/tags.py` - 기존 TAG 검증 로직 확장
- `tests/unit/core/tags/` - 새 테스트 파일 추가
- `.moai/docs/` - 42개 문서에 향후 TAG 삽입 (Phase 2)

---

## Implementation Notes (구현 참고사항)

### Phase 2-4 통합 계획
- **Phase 2**: 대량 적용 - 42개 문서에 자동 TAG 삽입 (Week 2)
- **Phase 3**: 검증 통합 - pre-commit hook 및 CI/CD 연동 (Week 3)
- **Phase 4**: 문서화 - 사용자 가이드 및 체인 시각화 (Week 4)

### 기술 부채 방지
- 모든 함수에 docstring 및 type hints 필수
- 테스트 우선 작성 (TDD 원칙 준수)
- ruff, mypy 검사 자동화 (CI/CD 통합)

### 확장성 고려
- 도메인 자동 추출 로직 분리 (미래 확장 용이)
- TAG 형식 변경 시 파서만 수정 가능하도록 설계
- 플러그인 아키텍처 고려 (미래 커스텀 TAG 타입 지원)

---

## References (참고 자료)

- MoAI-ADK Development Guide: `docs/development-guide.md`
- EARS Methodology: `.claude/skills/moai-foundation-ears.md`
- TAG Rules: `CLAUDE-RULES.md` (Section: @TAG Lifecycle)
- Python Style Guide: PEP 8, ruff configuration
