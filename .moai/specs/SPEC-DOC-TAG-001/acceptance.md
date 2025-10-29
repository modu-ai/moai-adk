# Acceptance Criteria: @DOC 태그 자동 생성 인프라

**SPEC ID**: DOC-TAG-001
**Version**: 0.0.1
**Status**: draft
**Created**: 2025-10-29
**Author**: @Goos

---

## Test Scenarios (Given-When-Then Format)

### 시나리오 1: @DOC ID 생성

**목표**: 새로운 문서 파일에 대해 고유한 @DOC TAG ID를 자동 생성한다.

**GIVEN**:
- `docs/api/new-endpoint.md` 파일이 @DOC 태그 없이 존재
- 시스템이 누락된 태그를 스캔
- 파일이 DOC 타입으로 식별됨 (`.md` 확장자)

**WHEN**:
- `generate_doc_tag("API", existing_tags)` 함수 호출
- `existing_tags`에 `["@DOC:API-001", "@DOC:CLI-001"]` 포함

**THEN**:
- 시스템이 `@DOC:API-002` 생성 (API 도메인의 다음 번호)
- 반환값이 `str` 타입
- 형식이 `@DOC:DOMAIN-NNN` 패턴 준수

**AND**:
- 높은 신뢰도와 함께 TagSuggestion 반환 (confidence ≥ 0.8)
- 중복 ID 없음 (ripgrep 검색 결과 빈 리스트)

**검증 방법**:
```python
# Unit test
def test_generate_doc_tag_incremental():
    existing = ["@DOC:API-001", "@DOC:CLI-001"]
    result = generate_doc_tag("API", existing)
    assert result == "@DOC:API-002"
    assert isinstance(result, str)
    assert re.match(r"@DOC:[A-Z-]+-\d{3}", result)
```

---

### 시나리오 2: 마크다운 헤더에 TAG 삽입

**목표**: 생성된 @DOC TAG를 마크다운 파일의 헤더에 안전하게 삽입한다.

**GIVEN**:
- 생성된 `@DOC:API-001` 태그
- 파일 경로 `docs/api/authentication.md`
- 파일 헤더: `# API Authentication`
- 연관 SPEC: `@SPEC:AUTH-001`

**WHEN**:
- `insert_tag_to_markdown("docs/api/authentication.md", "@DOC:API-001", ["@SPEC:AUTH-001"])` 호출
- 시스템이 파일 읽기/쓰기 수행

**THEN**:
- 파일 헤더가 다음으로 변경:
  ```markdown
  # @DOC:API-001 | Chain: @SPEC:AUTH-001 -> @DOC:API-001
  ```
- 원본 파일 내용 (헤더 제외) 손실 없음
- 파일 인코딩 UTF-8 유지

**AND**:
- 함수 반환값 `True` (성공)
- 파일이 올바른 줄바꿈 포함 (마크다운 표준 준수)
- Git diff로 헤더만 변경 확인 가능

**검증 방법**:
```python
# Unit test
def test_insert_tag_to_markdown_with_chain(tmp_path):
    # Arrange
    test_file = tmp_path / "test.md"
    test_file.write_text("# API Authentication\n\nContent here.")

    # Act
    result = insert_tag_to_markdown(
        str(test_file),
        "@DOC:API-001",
        ["@SPEC:AUTH-001"]
    )

    # Assert
    assert result is True
    content = test_file.read_text()
    assert content.startswith(
        "# @DOC:API-001 | Chain: @SPEC:AUTH-001 -> @DOC:API-001"
    )
    assert "Content here." in content
```

---

### 시나리오 3: 중복 TAG ID 방지

**목표**: 동일 도메인 내에서 중복 TAG ID가 생성되지 않도록 보장한다.

**GIVEN**:
- `docs/api/authentication.md`에 기존 `@DOC:AUTH-001` 존재
- ripgrep 검색 결과: `["@DOC:AUTH-001"]`

**WHEN**:
- 시스템이 AUTH 도메인에 대한 새 ID 생성 시도
- `generate_doc_tag("AUTH", ["@DOC:AUTH-001"])` 호출

**THEN**:
- 시스템이 `@DOC:AUTH-002` 반환 (AUTH-001이 **아님**)
- 중복 검사 로직 실행 확인 (rg 명령어 호출 로그)

**AND**:
- 전체 코드베이스 스캔으로 충돌 방지
- 여러 번 호출 시 일관된 결과 (멱등성)

**검증 방법**:
```python
# Unit test
def test_generate_doc_tag_no_duplicate():
    existing = ["@DOC:AUTH-001"]
    result = generate_doc_tag("AUTH", existing)
    assert result == "@DOC:AUTH-002"

    # Verify no duplicate
    assert result not in existing
```

**통합 테스트**:
```bash
# Manual test
$ rg "@DOC:AUTH-001" .
docs/api/authentication.md:1:# @DOC:AUTH-001 | Chain: ...

$ python -c "from moai_adk.core.tags import generate_doc_tag; print(generate_doc_tag('AUTH', ['@DOC:AUTH-001']))"
@DOC:AUTH-002
```

---

### 시나리오 4: SPEC 매핑 누락 처리

**목표**: SPEC 매핑을 자동으로 결정할 수 없을 때 독립형 TAG를 생성한다.

**GIVEN**:
- 새 문서 파일 `docs/misc/troubleshooting.md`
- 파일 내용에 `@SPEC:XXX` 참조 없음
- 파일 경로로 도메인 추정 불가 (misc 디렉토리)

**WHEN**:
- `map_spec_to_doc("docs/misc/troubleshooting.md")` 호출
- 시스템이 SPEC 매핑 실패 감지

**THEN**:
- 시스템이 독립형 @DOC 태그 생성 (`@DOC:MISC-001`)
- SpecDocMapping 반환:
  - `spec_ids`: `[]` (빈 리스트)
  - `confidence`: `0.0` (매핑 불가)
  - `reasoning`: "No SPEC reference found in content or path"

**AND**:
- 선택적 체인 참조 없이 헤더 생성:
  ```markdown
  # @DOC:MISC-001
  ```
- 사용자가 수동 검토 필요 알림 수신 (Phase 3 기능)

**검증 방법**:
```python
# Unit test
def test_map_spec_to_doc_no_mapping(tmp_path):
    # Arrange
    test_file = tmp_path / "troubleshooting.md"
    test_file.write_text("# Troubleshooting\n\nNo SPEC reference.")

    # Act
    result = map_spec_to_doc(str(test_file))

    # Assert
    assert result.spec_ids == []
    assert result.confidence == 0.0
    assert "No SPEC reference" in result.reasoning
```

---

### 시나리오 5: 검증 hook 통합

**목표**: 업데이트된 `tags.py`의 `suggest_tag_for_file()` 함수가 pre-commit hook과 통합된다.

**GIVEN**:
- 새로운 `suggest_tag_for_file()` 함수가 있는 업데이트된 `tags.py`
- pre-commit hook 설정 파일 (`.pre-commit-config.yaml`)
- 변경된 파일: `docs/api/new-feature.md`

**WHEN**:
- pre-commit hook이 변경된 파일 스캔
- `suggest_tag_for_file("docs/api/new-feature.md")` 자동 호출

**THEN**:
- hook이 오류가 **아닌** 태그 제안 제공 (non-blocking)
- TagSuggestion 객체 반환:
  - `domain`: `"API"`
  - `next_id`: `"@DOC:API-003"`
  - `confidence`: `0.8` (경로 기반 추정)
  - `reasoning`: "Domain extracted from file path: docs/api/"

**AND**:
- 제안이 터미널에 출력 (사용자 확인 가능)
- Git commit 차단하지 않음 (Phase 1에서는 제안만 제공)
- Phase 3에서 자동 적용 옵션 추가 예정

**검증 방법**:
```python
# Unit test
def test_suggest_tag_for_file():
    suggestion = suggest_tag_for_file("docs/api/new-feature.md")

    assert suggestion.domain == "API"
    assert suggestion.next_id.startswith("@DOC:API-")
    assert 0.0 <= suggestion.confidence <= 1.0
    assert len(suggestion.reasoning) > 0
```

**통합 테스트**:
```bash
# Manual test
$ git add docs/api/new-feature.md
$ git commit -m "test"
# Expected output:
# [TAG Suggestion] docs/api/new-feature.md -> @DOC:API-003 (confidence: 0.8)
# Reason: Domain extracted from file path: docs/api/
```

---

## 성공 기준 (Acceptance Criteria Definition)

### 1. 코드 품질

**테스트 커버리지**:
- ✅ 전체 커버리지 ≥ 85% (`pytest --cov=src/moai_adk/core/tags`)
- ✅ 각 모듈별 커버리지:
  - `generator.py`: ≥ 90%
  - `inserter.py`: ≥ 85%
  - `mapper.py`: ≥ 80%
  - `parser.py`: ≥ 95% (간단한 유틸리티)

**테스트 통과**:
- ✅ 모든 테스트 통과: `pytest tests/unit/core/tags/` (0 failures, 0 errors)
- ✅ 총 테스트 수 ≥ 20개 (5 + 4 + 3 + 2 + α)

**Code Style**:
- ✅ ruff 검사 통과: `ruff check src/moai_adk/core/tags/` (0 errors)
- ✅ ruff format 검사: `ruff format --check src/moai_adk/core/tags/` (0 changes)

**Type Checking**:
- ✅ mypy 검사 통과: `mypy src/moai_adk/core/tags/` (0 errors, 0 warnings)
- ✅ 모든 함수에 type hints 포함

**검증 명령어**:
```bash
# Run all quality checks
pytest tests/unit/core/tags/ --cov=src/moai_adk/core/tags --cov-report=term-missing
ruff check src/moai_adk/core/tags/
mypy src/moai_adk/core/tags/
```

---

### 2. 기능 검증

**generator.py 테스트** (5개 시나리오):
- ✅ 빈 리스트 → `@DOC:API-001` 생성
- ✅ 기존 API-001 존재 → `@DOC:API-002` 생성
- ✅ 여러 도메인 혼재 → 올바른 도메인만 카운트
- ✅ 번호 999 도달 → ValueError 발생
- ✅ 잘못된 도메인 입력 → ValueError 발생

**inserter.py 테스트** (4개 시나리오):
- ✅ 정상 파일 → 헤더 변경 성공
- ✅ Chain 없음 → `# @DOC:API-001` 형식
- ✅ Chain 있음 → `# @DOC:API-001 | Chain: ...` 형식
- ✅ 헤더 없는 파일 → 파일 맨 앞에 헤더 추가

**mapper.py 테스트** (3개 시나리오):
- ✅ 명시적 SPEC 참조 → confidence 1.0, 정확한 SPEC ID 반환
- ✅ 경로 기반 추정 → confidence 0.5-0.8, 추정 근거 제공
- ✅ 매핑 불가 → 빈 리스트, confidence 0.0, 이유 설명

**parser.py 테스트** (2개 시나리오):
- ✅ 여러 SPEC ID 있음 → 모두 추출 (중복 제거)
- ✅ SPEC ID 없음 → 빈 리스트 반환

**tags.py 검증** (1개 기능):
- ✅ `suggest_tag_for_file()` 함수 추가 완료
- ✅ TagSuggestion 반환 기능 확인

**검증 명령어**:
```bash
# Run specific test files
pytest tests/unit/core/tags/test_generator.py -v
pytest tests/unit/core/tags/test_inserter.py -v
pytest tests/unit/core/tags/test_mapper.py -v
```

---

### 3. 통합 검증

**수동 테스트 1: ID 생성**
```bash
$ python -c "
from moai_adk.core.tags import generate_doc_tag
result = generate_doc_tag('TEST', [])
print(f'Result: {result}')
assert result == '@DOC:TEST-001', f'Expected @DOC:TEST-001, got {result}'
print('✅ PASS: generate_doc_tag')
"
```
- ✅ 출력: `@DOC:TEST-001`
- ✅ 성공 메시지 확인

**수동 테스트 2: 마크다운 삽입**
```bash
$ cat > /tmp/test.md << 'EOF'
# Test Document

Content here.
EOF

$ python -c "
from moai_adk.core.tags import insert_tag_to_markdown
result = insert_tag_to_markdown('/tmp/test.md', '@DOC:TEST-001', ['@SPEC:TEST-001'])
print(f'Result: {result}')
assert result is True, 'Insertion failed'
print('✅ PASS: insert_tag_to_markdown')
"

$ cat /tmp/test.md
# Expected output:
# @DOC:TEST-001 | Chain: @SPEC:TEST-001 -> @DOC:TEST-001
#
# Content here.
```
- ✅ 헤더 변경 확인
- ✅ 원본 내용 보존 확인

**수동 테스트 3: 역호환성**
```bash
# Verify existing manual tags unchanged
$ rg "@DOC:" .moai/docs/ --count
# Expected: 2 (only manually tagged docs)

$ git diff .moai/docs/
# Expected: (empty - no changes)
```
- ✅ 기존 2개 문서 변경 없음
- ✅ Git diff 결과 비어있음

---

### 4. 문서화

**Docstring 완성도**:
- ✅ 모든 public 함수에 docstring 포함
- ✅ Docstring 형식: Google Style 또는 NumPy Style
- ✅ Args, Returns, Example 섹션 포함

**Type Hints 완성도**:
- ✅ 모든 함수 시그니처에 type hints
- ✅ 반환 타입 명시 (`-> str`, `-> bool`, 등)
- ✅ Optional 및 Union 타입 올바르게 사용

**사용 예제**:
- ✅ README 또는 개발 가이드에 사용 예제 추가
- ✅ 최소 3개 예제:
  1. TAG ID 생성
  2. 마크다운 삽입
  3. SPEC 매핑

**검증 방법**:
```bash
# Check docstring presence
python -c "
from moai_adk.core.tags import generate_doc_tag, insert_tag_to_markdown
assert generate_doc_tag.__doc__ is not None
assert insert_tag_to_markdown.__doc__ is not None
print('✅ All functions have docstrings')
"
```

---

## Phase 1 완료 체크리스트

### 파일 생성 (8개)
- [ ] `src/moai_adk/core/tags/__init__.py` (30줄)
- [ ] `src/moai_adk/core/tags/generator.py` (150줄)
- [ ] `src/moai_adk/core/tags/inserter.py` (200줄)
- [ ] `src/moai_adk/core/tags/mapper.py` (180줄)
- [ ] `src/moai_adk/core/tags/parser.py` (100줄)
- [ ] `tests/unit/core/tags/test_generator.py` (250줄)
- [ ] `tests/unit/core/tags/test_inserter.py` (300줄)
- [ ] `tests/unit/core/tags/test_mapper.py` (200줄)

### 품질 게이트
- [ ] pytest 커버리지 ≥ 85% 달성
- [ ] 모든 테스트 통과 (0 failures)
- [ ] ruff 검사 통과 (0 errors)
- [ ] mypy 검사 통과 (0 errors)

### 기능 검증
- [ ] 수동 검증: ID 생성 테스트 성공
- [ ] 수동 검증: 마크다운 삽입 테스트 성공
- [ ] 역호환성 확인 (기존 2개 문서 무손상)

### Git 작업 (git-manager 담당)
- [ ] Feature branch 생성 (`feature/DOC-TAG-001-auto-generation`)
- [ ] Commit 생성 (TDD RED → GREEN → REFACTOR)
- [ ] Draft PR 생성
- [ ] PR 설명 작성 (SPEC 링크 포함)

### 문서화
- [ ] README 또는 개발 가이드에 사용 예제 추가
- [ ] 모든 함수 docstring 완성
- [ ] 모든 함수 type hints 완성

---

## Phase 2 준비 체크리스트

Phase 1 완료 후, Phase 2 (대량 적용)로 진행하기 전:

- [ ] Phase 1 SPEC 검토 완료 (spec-builder)
- [ ] 사용자 승인 획득 (`/alfred:1-plan` 완료)
- [ ] 구현 완료 및 테스트 통과 (`/alfred:2-run` 완료)
- [ ] PR 병합 완료 (main 브랜치)
- [ ] 42개 문서 리스트 작성 (Phase 2 입력)
- [ ] Phase 2 SPEC 작성 시작 (`/alfred:1-plan "Phase 2: Apply tags to 42 docs"`)

---

## Definition of Done

**Phase 1은 다음 조건을 모두 만족할 때 완료된다**:

1. ✅ **모든 파일 생성 완료** (8개 파일)
2. ✅ **모든 테스트 통과** (0 failures, 0 errors)
3. ✅ **품질 게이트 통과** (커버리지 ≥ 85%, ruff 0 errors, mypy 0 errors)
4. ✅ **수동 검증 완료** (3개 시나리오 모두 성공)
5. ✅ **문서화 완료** (docstring, type hints, 사용 예제)
6. ✅ **Git 작업 완료** (branch, commit, PR)
7. ✅ **역호환성 확인** (기존 2개 문서 무손상)
8. ✅ **사용자 승인** (SPEC 검토 및 구현 승인)

**최종 확인 명령어**:
```bash
# All-in-one verification
pytest tests/unit/core/tags/ --cov=src/moai_adk/core/tags --cov-report=term-missing && \
ruff check src/moai_adk/core/tags/ && \
mypy src/moai_adk/core/tags/ && \
echo "✅ Phase 1 완료!"
```

---

## 참고 자료

- **EARS 방법론**: `.claude/skills/moai-foundation-ears.md`
- **TAG 규칙**: `CLAUDE-RULES.md` (Section: @TAG Lifecycle)
- **pytest 가이드**: https://docs.pytest.org/
- **Given-When-Then 형식**: https://martinfowler.com/bliki/GivenWhenThen.html
