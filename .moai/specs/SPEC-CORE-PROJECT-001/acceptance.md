# SPEC-CORE-PROJECT-001 수락 기준

## Given-When-Then 테스트 시나리오

### 시나리오 1: Python 프로젝트 감지
**Given**: *.py 파일이 존재함
**When**: `LanguageDetector().detect()` 호출
**Then**:
- [ ] "python" 반환

### 시나리오 2: TypeScript 프로젝트 감지
**Given**: package.json과 tsconfig.json이 존재함
**When**: `LanguageDetector().detect()` 호출
**Then**:
- [ ] "typescript" 반환

### 시나리오 3: 프로젝트 초기화
**Given**: 빈 디렉토리
**When**: `ProjectInitializer().initialize()` 호출
**Then**:
- [ ] .moai/ 디렉토리 생성
- [ ] config.json 파일 생성
- [ ] product.md, structure.md, tech.md 생성

### 시나리오 4: 시스템 체크
**Given**: git과 python3가 설치됨
**When**: `SystemChecker().check_all()` 호출
**Then**:
- [ ] {"git": True, "python": True, ...} 반환

### 시나리오 5: 멀티 언어 감지
**Given**: *.py와 *.ts 파일이 모두 존재함
**When**: `LanguageDetector().detect_multiple()` 호출
**Then**:
- [ ] ["python", "typescript"] 반환

---

## 품질 게이트 기준

### 1. 언어 감지 완성도
- [ ] 20개 언어 모두 감지 가능
- [ ] 멀티 언어 감지 정상 작동
- [ ] 감지 실패 시 None 또는 "generic" 반환

### 2. 초기화 완성도
- [ ] .moai/ 디렉토리 구조 올바름
- [ ] config.json 유효한 JSON
- [ ] 언어별 템플릿 적용됨

### 3. 시스템 체크
- [ ] 필수 도구(git, python) 확인
- [ ] 선택 도구(gh, docker) 확인
- [ ] 도구 없을 때 False 반환

---

## 검증 방법 및 도구

### 자동화 테스트
```python
# tests/core/project/test_detector.py
from moai_adk.core.project.detector import LanguageDetector

def test_detect_python(tmp_path):
    (tmp_path / "main.py").touch()
    detector = LanguageDetector()
    assert detector.detect(str(tmp_path)) == "python"

def test_detect_typescript(tmp_path):
    (tmp_path / "tsconfig.json").touch()
    detector = LanguageDetector()
    assert detector.detect(str(tmp_path)) == "typescript"
```

### 수동 검증
1. **언어 감지**: 각 언어 샘플 프로젝트에서 테스트
2. **초기화**: 빈 디렉토리에서 `moai init .` 실행
3. **시스템 체크**: `moai doctor` 실행

---

## 완료 조건 (Definition of Done)

### 필수 조건
- [ ] LanguageDetector 구현
- [ ] ProjectInitializer 구현
- [ ] SystemChecker 구현
- [ ] 20개 언어 지원
- [ ] 모든 테스트 통과

### 선택 조건
- [ ] 언어 우선순위 설정
- [ ] 커스텀 템플릿 지원

### 문서화
- [ ] 지원 언어 목록
- [ ] 초기화 가이드
