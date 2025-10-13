# SPEC-CORE-TEMPLATE-001 수락 기준

## Given-When-Then 테스트 시나리오

### 시나리오 1: 템플릿 렌더링
**Given**: config.json.j2 템플릿이 존재함
**When**: `render("config.json.j2", {"version": "0.3.0"})` 호출
**Then**:
- [ ] "0.3.0"이 포함된 JSON 문자열 반환
- [ ] 유효한 JSON 형식

### 시나리오 2: 파일 렌더링
**Given**: 템플릿과 컨텍스트가 준비됨
**When**: `render_to_file("config.json.j2", ".moai/config.json", context)` 호출
**Then**:
- [ ] .moai/config.json 파일 생성됨
- [ ] 파일 내용이 올바르게 렌더링됨

### 시나리오 3: config.json 읽기
**Given**: .moai/config.json 파일이 존재함
**When**: `ConfigManager().load()` 호출
**Then**:
- [ ] 딕셔너리 반환
- [ ] 필수 키 존재 (mode, locale, git, spec)

### 시나리오 4: config.json 업데이트
**Given**: 기존 config.json이 있음
**When**: `update({"locale": "en"})` 호출
**Then**:
- [ ] locale만 "en"으로 변경
- [ ] 다른 필드는 유지

### 시나리오 5: 언어 템플릿 조회
**Given**: LANGUAGE_TEMPLATES 정의됨
**When**: `get_language_template("python")` 호출
**Then**:
- [ ] ".moai/project/tech/python.md.j2" 반환

---

## 품질 게이트 기준

### 1. 템플릿 완성도
- [ ] 20개 언어 템플릿 전환 완료
- [ ] 템플릿 변수 모두 치환됨
- [ ] 렌더링 오류 없음

### 2. config.json 관리
- [ ] 읽기/쓰기/업데이트 성공
- [ ] deep_merge 정상 작동
- [ ] 인코딩 문제 없음

### 3. 에러 처리
- [ ] 템플릿 파일 없을 때 처리
- [ ] 잘못된 JSON 처리
- [ ] 변수 누락 시 처리

---

## 검증 방법 및 도구

### 자동화 테스트
```python
# tests/core/template/test_processor.py
from moai_adk.core.template import TemplateProcessor

def test_render():
    processor = TemplateProcessor("tests/fixtures/templates")
    result = processor.render("test.j2", {"name": "MoAI"})
    assert "MoAI" in result

def test_config_load():
    config = ConfigManager(".moai/config.json").load()
    assert "mode" in config
    assert config["locale"] in ["ko", "en"]
```

### 수동 검증
1. **템플릿 렌더링**: 각 언어 템플릿 수동 확인
2. **config.json**: 생성된 파일 JSON Lint 확인
3. **변수 치환**: 모든 {{ variable }} 치환됨

---

## 완료 조건 (Definition of Done)

### 필수 조건
- [ ] TemplateProcessor 구현
- [ ] ConfigManager 구현
- [ ] 20개 언어 템플릿 전환
- [ ] 모든 테스트 통과

### 선택 조건
- [ ] 템플릿 캐싱 최적화
- [ ] 템플릿 검증 도구

### 문서화
- [ ] 템플릿 변수 목록
- [ ] 예제 코드 추가
