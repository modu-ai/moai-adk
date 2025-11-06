---
id: DOCS-001
version: 1.0.0
status: draft
created: 2025-01-06
updated: 2025-01-06
author: @Goos
priority: high
category: documentation
related_specs:
  - SPEC-INSTALL-001
  - SPEC-INIT-001
  - SPEC-CONFIG-001
traceability:
  spec: "@SPEC:DOCS-001"
  test: "@TEST:DOCS-001"
  code: "@CODE:DOCS-001"
---

# `@ACCEPTANCE:DOCS-001: MoAI-ADK 문서 개선 수락 기준`

## Given-When-Then 테스트 시나리오

이 문서는 SPEC-DOCS-001의 수락 기준을 Given-When-Then 형식으로 정의합니다.

---

## 시나리오 1: README.ko.md 분할 완료

### Given (전제 조건)
- README.ko.md 파일이 3295줄로 존재함
- 분할 스크립트(`scripts/split_readme.py`)가 준비됨
- Nextra 설정(`theme.config.cjs`)이 준비됨

### When (실행 조건)
- 개발자가 분할 스크립트를 실행함
```bash
python scripts/split_readme.py
```

### Then (예상 결과)
- README.ko.md가 간결한 소개로 재구성됨 (100줄 이내)
- 20+개의 새로운 문서 파일이 생성됨
- 모든 섹션이 논리적으로 분리됨
- 내용 누락률 ≤ 1%

### 검증 방법
```bash
# 원본과 분할 후 내용 비교
python scripts/validate_split.py --original README.ko.md --target docs/

# 기대 출력:
# Content preservation: 99.2%
# Sections split: 23
# New files created: 21
# Content preserved: 3270/3295 lines
```

---

## 시나리오 2: 실제 코드 기반 예제 검증

### Given
- 모든 예제가 `src/moai_adk/` 실제 코드를 참조함
- `@CODE:` 태그로 실제 파일과 연동됨

### When
- 사용자가 문서의 코드 예제를 실행함

### Then
- 모든 예제가 실제로 실행 가능함
- `@CODE:` 태그가 유효한 파일을 참조함
- 예제 코드와 실제 구현이 일치함

### 검증 방법
```bash
# 모든 @CODE: 태그 유효성 검증
rg '@CODE:DOCS-001' docs/ -A1 -B1 | xargs -I {} sh -c 'file=$(echo {} | grep -o "src/[^:]*"); [ -f "$file" ] && echo "✅ $file exists" || echo "❌ $file not found"'

# 예제 실행 테스트
python scripts/test_examples.py
```

---

## 시나리오 3: Mermaid 다이어그램 렌더링

### Given
- Nextra Mermaid 플러그인이 설정됨
- 다이어그램이 문서에 포함됨

### When
- 사용자가 다이어그램이 포함된 페이지를 방문함

### Then
- 모든 Mermaid 다이어그램이 정상적으로 렌더링됨
- 노드 수 ≤ 20개 (가독성 유지)
- 대체 텍스트가 제공됨

### 검증 방법
```bash
# 다이어그램 문법 검증
python scripts/validate_mermaid.py docs/

# 브라우저에서 시각적 확인
npm run docs:dev
# http://localhost:5173 접속
```

---

## 시나리오 4: 표 형식 구조화

### Given
- 문서에 표 형식이 포함됨
- 마크다운 테이블 문법이 올바르게 사용됨

### When
- 사용자가 표가 포함된 페이지를 조회함

### Then
- 모든 표가 올바르게 렌더링됨
- 헤더 행이 명확하게 구분됨
- 정렬이 올바르게 적용됨

### 검증 방법
```markdown
| 명령 | 기능 | 산출물 |
|------|------|--------|
| /alfred:0-project | 프로젝트 초기화 | 설정 파일 |
```

---

## 시나리오 5: 다국어 구조 준비

### Given
- 한국어 문서가 완성됨
- 다국어 디렉토리 구조가 준비됨

### When
- Nextra i18n 설정이 활성화됨

### Then
- 언어 전환 기능이 제공됨
- 각 언어별로 올바른 경로 구조를 가짐
- 번역 관리가 용이함

### 검증 방법
```bash
# 다국어 구조 확인
ls docs/ko docs/en docs/ja docs/zh

# Nextra 설정 확인
grep -n "i18n" docs/nextra.config.js
```

---

## 시나리오 6: 문서 탐색 용이성

### Given
- 문서가 주제별로 분할됨
- 검색 인덱스가 생성됨

### When
- 사용자가 특정 정보를 검색함

### Then
- 핵심 정보를 3클릭 내에 접근 가능
- 검색 결과가 관련성 순으로 정렬됨
- 빠른 시작 가이드를 5분 내에 완료 가능

### 검증 방법
```bash
# 검색 기능 테스트
npm run docs:dev
# Cmd/Ctrl+K로 검색창 열기
# 핵심 키워드 검색: "설치", "SPEC", "@TAG"
```

---

## 품질 게이트 기준

### 문서 분할 품질
- ✅ README.ko.md: 3295줄 → ≤100줄
- ✅ 새 문서: 20+개 파일
- ✅ 내용 보존률: ≥99%
- ✅ 섹션별 논리적 분리 완료

### 코드 예제 품질
- ✅ 실제 실행 가능 예제: 100%
- ✅ `@CODE:` 태그 연결: 100%
- ✅ 예제-실제 코드 일치: 100%
- ✅ 실행 검증 테스트 통과

### 시각화 품질
- ✅ Mermaid 다이어그램 렌더링: 100%
- ✅ 노드 수 제한 준수 (≤20개)
- ✅ 대체 텍스트 제공: 100%
- ✅ 색상 구분 명확성

### 다국어 품질
- ✅ 한국어 문서: 100% 완료
- ✅ 영어 번역: ≥80% 진행
- ✅ 일본어 번역: ≥60% 진행
- ✅ 중국어 번역: ≥40% 진행

### 성능 품질
- ✅ 페이지 로딩 시간: <2초
- ✅ 검색 응답 시간: <500ms
- ✅ 문서 빌드 시간: <30초
- ✅ 핫 리로드 시간: <3초

---

## Definition of Done (완료 조건)

### Phase 1: README 분할 완료
- ✅ README.ko.md를 5개 주제로 분할
- ✅ 분할 스크립트 실행 완료
- ✅ 내용 검증 통과 (99% 보존)
- ✅ Nextra 경로 설정 완료

### Phase 2: 코드 예제 완료
- ✅ 15+개 실제 코드 예제 작성
- ✅ 모든 예제 실행 가능 검증
- ✅ `@CODE:` 태그로 파일 연동
- ✅ 예제 실행 가이드 제공

### Phase 3: 시각화 완료
- ✅ 워크플로우 다이어그램 3개 작성
- ✅ 아키텍처 다이어그램 작성
- ✅ TAG 체인 다이어그램 작성
- ✅ 모든 다이어그램 렌더링 검증

### Phase 4: 표 형식 완료
- ✅ 명령어 요약표 작성
- ✅ 에이전트 목록표 작성
- ✅ 버전 히스토리표 작성
- ✅ 모든 표 마크다운 형식 검증

### Phase 5: 다국어 준비 완료
- ✅ 다국어 디렉토리 구조 준비
- ✅ Nextra i18n 설정
- ✅ 영어 번역 80% 완료
- ✅ 번역 관리 가이드 작성

---

## 수락 테스트 체크리스트

### 기능 테스트
- [ ] README.ko.md 분할 완료
- [ ] 새 문서 20+개 생성 확인
- [ ] 내용 보존률 99% 검증
- [ ] 코드 예제 실행 가능 확인
- [ ] Mermaid 다이어그램 렌더링
- [ ] 표 형식 올바른 표시
- [ ] 검색 기능 정상 동작
- [ ] 다국어 전환 기능

### 품질 테스트
- [ ] 문서 빌드 성공
- [ ] 링크 유효성 검증
- [ ] 코드 예제 실행 테스트
- [ ] 다국어 렌더링 테스트
- [ ] 성능 기준 준수
- [ ] 접근성 가이드라인 준수

### 사용자 경험 테스트
- [ ] 빠른 시작 5분 완료 가능
- [ ] 핵심 정보 3클릭 내 접근
- [ ] 문서 명확도 4.5/5.0
- [ ] 예제 유용성 4.7/5.0

---

## 검증 스크립트 예시

### 분할 검증 스크립트 (`scripts/validate_split.py`)
```python
import os
import hashlib

def validate_content_preservation():
    """README 분할 후 내용 보존률 검증"""
    with open('README.ko.md.bak', 'r') as f:
        original_hash = hashlib.md5(f.read().encode()).hexdigest()

    # 분할된 문서 내용 취합
    combined_content = ""
    for root, dirs, files in os.walk('docs/'):
        for file in files:
            if file.endswith('.md'):
                with open(os.path.join(root, file), 'r') as f:
                    combined_content += f.read()

    combined_hash = hashlib.md5(combined_content.encode()).hexdigest()
    preservation_rate = (1 - abs(int(original_hash, 16) - int(combined_hash, 16)) / int(original_hash, 16)) * 100

    print(f"Content preservation: {preservation_rate:.1f}%")
    return preservation_rate >= 99.0

if __name__ == "__main__":
    validate_content_preservation()
```

### 코드 예제 검증 스크립트 (`scripts/test_examples.py`)
```python
import subprocess
import re

def extract_code_from_doc(doc_path):
    """문서에서 코드 예제 추출"""
    with open(doc_path, 'r') as f:
        content = f.read()

    # 코드 블록 추출
    code_blocks = re.findall(r'```python\n(.*?)\n```', content, re.DOTALL)
    return code_blocks

def test_example(code):
    """코드 예제 실행 테스트"""
    try:
        result = subprocess.run(['python', '-c', code],
                              capture_output=True, text=True, timeout=10)
        return result.returncode == 0
    except subprocess.TimeoutExpired:
        return False
    except Exception:
        return False

def validate_all_examples():
    """모든 코드 예제 검증"""
    total_examples = 0
    passing_examples = 0

    for root, dirs, files in os.walk('docs/'):
        for file in files:
            if file.endswith('.md'):
                doc_path = os.path.join(root, file)
                examples = extract_code_from_doc(doc_path)

                for example in examples:
                    total_examples += 1
                    if test_example(example):
                        passing_examples += 1

    success_rate = (passing_examples / total_examples) * 100 if total_examples > 0 else 0
    print(f"Example test pass rate: {success_rate:.1f}% ({passing_examples}/{total_examples})")
    return success_rate >= 100.0
```

---

**작성일**: 2025-01-06
**버전**: 1.0.0
**관련 SPEC**: @SPEC:DOCS-001