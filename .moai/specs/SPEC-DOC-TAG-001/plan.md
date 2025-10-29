# Implementation Plan: @DOC 태그 자동 생성 인프라

**SPEC ID**: DOC-TAG-001
**Version**: 0.0.1
**Status**: draft
**Created**: 2025-10-29
**Author**: @Goos

---

## 개요 (Overview)

### Phase 1의 목표
Phase 1은 MoAI-ADK의 @DOC 태그 자동 생성 시스템의 **핵심 인프라**를 구축하는 단계입니다. 42개의 문서 파일에 수동으로 태그를 추가하는 대신, 자동화된 라이브러리를 통해 일관되고 추적 가능한 태그를 생성합니다.

**핵심 가치**:
- ✅ **자동화**: 수동 작업 42회 → 자동 생성 1회 실행
- ✅ **일관성**: 모든 TAG가 `@DOC:DOMAIN-NNN` 형식 준수
- ✅ **추적성**: SPEC-DOC 체인 자동 연결
- ✅ **안전성**: 85%+ 테스트 커버리지로 품질 보장

### 최종 성과물 (Deliverables)

| # | 파일 경로 | 라인 수 (예상) | 설명 |
|---|-----------|---------------|------|
| 1 | `src/moai_adk/core/tags/generator.py` | ~150 | TAG ID 생성 로직 |
| 2 | `src/moai_adk/core/tags/inserter.py` | ~200 | 마크다운 헤더 삽입 |
| 3 | `src/moai_adk/core/tags/mapper.py` | ~180 | SPEC-DOC 매핑 |
| 4 | `src/moai_adk/core/tags/parser.py` | ~100 | SPEC ID 추출 |
| 5 | `src/moai_adk/core/tags/__init__.py` | ~30 | 모듈 초기화 |
| 6 | `tests/unit/core/tags/test_generator.py` | ~250 | generator 테스트 |
| 7 | `tests/unit/core/tags/test_inserter.py` | ~300 | inserter 테스트 |
| 8 | `tests/unit/core/tags/test_mapper.py` | ~200 | mapper 테스트 |

**총계**: 8개 파일, 약 1,410 라인

---

## 상세 구현 계획

### 모듈 1: generator.py - TAG ID 생성 로직

**책임**:
- 도메인 기반으로 고유한 @DOC TAG ID 생성
- 중복 ID 자동 탐지 및 회피
- 번호 자동 증가 로직

**핵심 함수**:

```python
def generate_doc_tag(domain: str, existing_tags: list[str]) -> str:
    """
    Generate unique @DOC TAG ID for given domain.

    Args:
        domain: Domain name (e.g., "API", "CLI")
        existing_tags: List of existing TAG IDs to check duplicates

    Returns:
        Generated TAG ID (e.g., "@DOC:API-001")

    Example:
        >>> generate_doc_tag("API", ["@DOC:API-001", "@DOC:CLI-001"])
        "@DOC:API-002"
    """
```

**구현 상세**:
1. 입력 도메인 검증 (대문자 영문자만 허용)
2. `existing_tags`에서 동일 도메인 TAG 필터링
3. 정규식으로 번호 추출 (`r"@DOC:([A-Z-]+)-(\d{3})"`)
4. 최대 번호 + 1 계산
5. `@DOC:{domain}-{number:03d}` 형식으로 반환

**테스트 시나리오** (5개):
1. 빈 리스트 → `@DOC:API-001` 생성
2. 기존 API-001 존재 → `@DOC:API-002` 생성
3. 여러 도메인 혼재 → 올바른 도메인만 카운트
4. 번호 999 도달 → 오류 발생 (제약사항)
5. 잘못된 도메인 입력 → ValueError 발생

**예상 라인 수**: ~150 라인 (docstring 포함)

---

### 모듈 2: inserter.py - 마크다운 헤더 삽입

**책임**:
- 마크다운 파일의 첫 번째 헤더 탐지
- @DOC TAG 및 Chain 정보 삽입
- 파일 무결성 보장 (원본 내용 손실 방지)

**핵심 함수**:

```python
def insert_tag_to_markdown(
    file_path: str,
    tag_id: str,
    chain: list[str] | None = None
) -> bool:
    """
    Insert @DOC TAG to markdown file header.

    Args:
        file_path: Path to markdown file
        tag_id: Generated TAG ID (e.g., "@DOC:API-001")
        chain: Optional chain of related TAGs (e.g., ["@SPEC:AUTH-001"])

    Returns:
        True if insertion successful, False otherwise

    Example:
        >>> insert_tag_to_markdown("docs/api/auth.md", "@DOC:API-001", ["@SPEC:AUTH-001"])
        True
        # File header changes:
        # Before: # API Authentication
        # After:  # @DOC:API-001 | Chain: @SPEC:AUTH-001 -> @DOC:API-001
    """
```

**구현 상세**:
1. 파일 존재 여부 확인 (`Path(file_path).exists()`)
2. 전체 내용 읽기 (UTF-8 인코딩)
3. 정규식으로 첫 번째 `# Header` 탐지
4. Chain 정보 포맷팅 (있으면 `| Chain: ...`, 없으면 생략)
5. 헤더 라인 교체
6. 파일 쓰기 (백업 없이 직접 덮어쓰기, Git에서 관리)
7. 성공/실패 반환

**테스트 시나리오** (4개):
1. 정상 파일 → 헤더 변경 성공
2. Chain 없음 → `# @DOC:API-001` 형식
3. Chain 있음 → `# @DOC:API-001 | Chain: @SPEC:AUTH-001 -> @DOC:API-001`
4. 헤더 없는 파일 → 파일 맨 앞에 헤더 추가

**예상 라인 수**: ~200 라인 (에러 처리 포함)

---

### 모듈 3: mapper.py - SPEC-DOC 매핑

**책임**:
- 문서 파일과 연관된 SPEC ID 자동 추출
- 파일 경로 및 내용 분석
- 신뢰도 점수 계산

**핵심 함수**:

```python
@dataclass
class SpecDocMapping:
    """SPEC-DOC mapping result."""
    spec_ids: list[str]
    confidence: float  # 0.0-1.0
    reasoning: str

def map_spec_to_doc(file_path: str) -> SpecDocMapping:
    """
    Map SPEC IDs to documentation file.

    Args:
        file_path: Path to documentation file

    Returns:
        SpecDocMapping with detected SPEC IDs and confidence

    Example:
        >>> result = map_spec_to_doc("docs/api/auth.md")
        >>> result.spec_ids
        ["@SPEC:AUTH-001"]
        >>> result.confidence
        1.0
    """
```

**구현 상세**:
1. 파일 내용 읽기
2. `parse_spec_id_from_content()` 호출하여 SPEC ID 추출
3. 파일 경로 분석 (예: `docs/api/` → API 도메인)
4. 신뢰도 계산:
   - 명시적 `@SPEC:XXX` 참조 있음 → 1.0
   - 파일 경로만으로 추정 → 0.5-0.8
   - 매핑 불가 → 0.0
5. SpecDocMapping 객체 반환

**테스트 시나리오** (3개):
1. 명시적 SPEC 참조 → confidence 1.0
2. 경로 기반 추정 → confidence 0.5-0.8
3. 매핑 불가 → 빈 리스트, confidence 0.0

**예상 라인 수**: ~180 라인 (dataclass 포함)

---

### 모듈 4: parser.py - SPEC ID 추출

**책임**:
- 텍스트에서 `@SPEC:DOMAIN-NNN` 패턴 추출
- 정규식 기반 파싱
- 중복 제거

**핵심 함수**:

```python
def parse_spec_id_from_content(content: str) -> list[str]:
    """
    Parse SPEC IDs from text content.

    Args:
        content: Text content to parse

    Returns:
        List of unique SPEC IDs found

    Example:
        >>> parse_spec_id_from_content("See @SPEC:AUTH-001 and @SPEC:API-001")
        ["@SPEC:AUTH-001", "@SPEC:API-001"]
    """
```

**구현 상세**:
1. 정규식 패턴: `r"@SPEC:[A-Z-]+-\d{3}"`
2. `re.findall()` 사용
3. 중복 제거 (set 변환 후 list 반환)
4. 빈 리스트 허용 (매칭 없을 경우)

**테스트 시나리오** (2개):
1. 여러 SPEC ID 있음 → 모두 추출
2. SPEC ID 없음 → 빈 리스트

**예상 라인 수**: ~100 라인 (간단한 유틸리티)

---

### 모듈 5: tags.py 수정 - 검증 강화

**책임**:
- 기존 `tags.py`에 `suggest_tag_for_file()` 함수 추가
- TagSuggestion 데이터 클래스 정의
- pre-commit hook 통합 준비

**추가 함수**:

```python
@dataclass
class TagSuggestion:
    """TAG suggestion for a file."""
    domain: str
    next_id: str
    confidence: float
    reasoning: str

def suggest_tag_for_file(file_path: str) -> TagSuggestion:
    """
    Suggest TAG for given file.

    Args:
        file_path: Path to file needing TAG

    Returns:
        TagSuggestion with recommended TAG and reasoning

    Example:
        >>> suggestion = suggest_tag_for_file("docs/api/new-endpoint.md")
        >>> suggestion.domain
        "API"
        >>> suggestion.next_id
        "@DOC:API-003"
    """
```

**구현 상세**:
1. 파일 경로에서 도메인 추출 (`docs/api/` → `API`)
2. `map_spec_to_doc()`로 연관 SPEC 확인
3. ripgrep으로 기존 TAG 검색
4. `generate_doc_tag()`로 다음 ID 생성
5. TagSuggestion 반환

**예상 수정 라인**: ~80 라인 (기존 파일에 추가)

---

### 모듈 6: 테스트 작성

**test_generator.py** (~250 라인):
- 5개 시나리오 (위 "모듈 1" 참조)
- fixtures: `existing_tags`, `temp_dir`
- pytest parametrize 사용

**test_inserter.py** (~300 라인):
- 4개 시나리오 (위 "모듈 2" 참조)
- fixtures: `temp_markdown_file`
- 파일 읽기/쓰기 검증

**test_mapper.py** (~200 라인):
- 3개 시나리오 (위 "모듈 3" 참조)
- fixtures: `sample_docs`
- 신뢰도 점수 검증

**테스트 전략**:
- 모든 테스트는 임시 디렉토리 사용 (`pytest tmpdir`)
- 실제 문서 파일 건드리지 않음
- 격리된 환경에서 실행

---

## 폴더 구조

```
src/moai_adk/core/tags/
├── __init__.py              # 모듈 초기화 (~30줄)
├── generator.py             # TAG ID 생성 (~150줄)
├── inserter.py              # 마크다운 삽입 (~200줄)
├── mapper.py                # SPEC-DOC 매핑 (~180줄)
└── parser.py                # SPEC ID 추출 (~100줄)

tests/unit/core/tags/
├── __init__.py              # 테스트 초기화
├── test_generator.py        # generator 테스트 (~250줄)
├── test_inserter.py         # inserter 테스트 (~300줄)
└── test_mapper.py           # mapper 테스트 (~200줄)

.moai/specs/SPEC-DOC-TAG-001/
├── spec.md                  # 본 SPEC 문서
├── plan.md                  # 본 계획 문서
└── acceptance.md            # 수락 기준
```

**총 파일 수**: 11개 (코드 5 + 테스트 4 + SPEC 3)

---

## 기술 스택

### 코어 라이브러리
- **Python**: 3.13+
- **표준 라이브러리**:
  - `re`: 정규식 처리
  - `pathlib`: 파일 경로 관리
  - `dataclasses`: 데이터 클래스 정의
  - `typing`: Type hints

### 테스트 도구
- **pytest**: 8.4.2 (테스트 프레임워크)
- **pytest-cov**: 커버리지 측정 (85%+ 목표)
- **pytest-mock**: Mock 객체 지원

### 코드 품질 도구
- **ruff**: 0.13.1 (linter + formatter)
- **mypy**: 1.8.0 (정적 타입 검사)
- **black**: 코드 포맷팅 (ruff로 대체 가능)

### 외부 도구
- **ripgrep (rg)**: TAG ID 중복 검사용
- **Git**: 버전 관리

---

## 일정 및 마일스톤

### Week 1: 코어 모듈 + 기본 테스트

**Milestone 1: 코어 라이브러리 구현**
- [ ] `generator.py` 작성 및 테스트
- [ ] `inserter.py` 작성 및 테스트
- [ ] `mapper.py` 작성 및 테스트
- [ ] `parser.py` 작성 및 테스트
- [ ] `tags.py` 수정 (suggest_tag_for_file 추가)

**확인 기준**:
- ✅ 모든 테스트 통과 (`pytest tests/unit/core/tags/`)
- ✅ 커버리지 85% 이상 (`pytest --cov`)
- ✅ ruff 검사 통과 (`ruff check src/moai_adk/core/tags/`)
- ✅ mypy 검사 통과 (`mypy src/moai_adk/core/tags/`)

**Milestone 2: 수동 검증**
- [ ] `generate_doc_tag("TEST", [])` 실행 → `@DOC:TEST-001` 확인
- [ ] 임시 마크다운 파일에 TAG 삽입 테스트
- [ ] 기존 2개 수동 태그 문서 변경 없음 확인

**확인 기준**:
- ✅ 수동 테스트 3개 모두 성공
- ✅ 역호환성 유지 (기존 문서 무손상)

---

## 리스크 및 대응책

### 리스크 1: 파일 손상 위험
- **설명**: 마크다운 삽입 중 원본 파일 손상 가능성
- **영향도**: High
- **대응책**:
  - Git으로 모든 변경사항 추적
  - 테스트 단계에서 임시 파일만 사용
  - Phase 2 적용 전 백업 생성

### 리스크 2: 중복 ID 생성
- **설명**: ripgrep 검색 실패 시 중복 ID 생성 가능
- **영향도**: Medium
- **대응책**:
  - ripgrep 설치 여부 사전 확인
  - fallback으로 `grep` 또는 Python `glob` 사용
  - 테스트에서 중복 시나리오 커버

### 리스크 3: 테스트 커버리지 미달
- **설명**: 85% 목표 달성 실패
- **영향도**: Low
- **대응책**:
  - TDD 원칙 준수 (테스트 우선 작성)
  - 엣지 케이스 시나리오 추가
  - pytest-cov로 실시간 모니터링

### 리스크 4: SPEC 매핑 정확도 낮음
- **설명**: 자동 매핑이 잘못된 SPEC 연결
- **영향도**: Medium
- **대응책**:
  - 신뢰도 점수로 불확실성 표시
  - Phase 2에서 수동 검토 단계 추가
  - Phase 3에서 사용자 피드백 반영

---

## 성공 지표

### 정량적 지표
- ✅ **테스트 커버리지**: ≥ 85%
- ✅ **테스트 통과율**: 100% (0 failures)
- ✅ **코드 품질**: ruff 0 errors, mypy 0 errors
- ✅ **파일 생성**: 8개 모두 완료

### 정성적 지표
- ✅ **코드 가독성**: 모든 함수에 docstring 및 type hints
- ✅ **유지보수성**: 모듈 간 낮은 결합도 (독립 실행 가능)
- ✅ **확장성**: Phase 2-4 통합 준비 완료

---

## 다음 단계 (Phase 2 준비)

Phase 1 완료 후:
1. **검토**: spec-builder가 SPEC 문서 검토
2. **승인**: 사용자 승인 후 구현 시작 (`/alfred:2-run`)
3. **구현**: tdd-implementer가 TDD 사이클 실행
4. **통합**: git-manager가 branch/PR 생성
5. **Phase 2**: 42개 문서 대량 적용 준비

---

## 참고 자료

- **EARS 방법론**: `.claude/skills/moai-foundation-ears.md`
- **TAG 규칙**: `CLAUDE-RULES.md` (Section: @TAG Lifecycle)
- **pytest 가이드**: https://docs.pytest.org/
- **ruff 설정**: `pyproject.toml`
