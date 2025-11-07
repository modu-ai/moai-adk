# Translation Contributing Guide

MoAI-ADK 문서 번역에 기여해 주셔서 감사합니다! 이 가이드는 번역 작업을 시작하는 방법을 안내합니다.

## 📊 Current Translation Status

번역 현황을 확인하려면 [Translation Status Dashboard](../translation-status.md)를 참조하세요.

## 🌍 Supported Languages

현재 지원하는 언어:

- **English (en)** - 영어
- **Japanese (ja)** - 일본어
- **Chinese (zh)** - 중국어 (간체)

## 🚀 Quick Start

### 1. 번역할 파일 선택

[Translation Status Dashboard](../translation-status.md)에서 누락된 파일 목록을 확인하세요.

각 언어별로 번역이 필요한 파일이 표시됩니다.

### 2. 파일 구조 이해

```
docs/src/
├── index.md                    # 한국어 (기본)
├── getting-started/
│   ├── installation.md
│   └── quick-start.md
├── en/                         # 영어 번역
│   ├── getting-started/
│   │   ├── installation.md
│   │   └── quick-start.md
├── ja/                         # 일본어 번역
│   └── ...
└── zh/                         # 중국어 번역
    └── ...
```

**핵심 원칙**:
- 한국어 원본: `docs/src/` 루트 및 하위 디렉토리
- 번역본: `docs/src/{언어코드}/` 하위에 동일한 디렉토리 구조 유지

### 3. 번역 작업 시작

#### 방법 A: GitHub Web UI 사용

1. GitHub에서 번역할 파일 찾기
2. "Edit" 버튼 클릭
3. 번역 내용 작성
4. "Propose changes" 클릭
5. Pull Request 생성

#### 방법 B: 로컬 환경 사용

```bash
# 1. Repository fork 및 clone
git clone https://github.com/YOUR_USERNAME/moai-adk.git
cd moai-adk

# 2. 번역 브랜치 생성
git checkout -b translate-ja-getting-started

# 3. 번역 파일 생성
# 예: docs/src/getting-started/installation.md를 일본어로 번역
mkdir -p docs/src/ja/getting-started
cp docs/src/getting-started/installation.md docs/src/ja/getting-started/installation.md

# 4. 파일 번역 (에디터로 열어서 작업)

# 5. 변경사항 확인
python docs/scripts/check_translation_status.py

# 6. Commit 및 Push
git add docs/src/ja/
git commit -m "docs: Add Japanese translation for installation guide"
git push origin translate-ja-getting-started

# 7. GitHub에서 Pull Request 생성
```

## 📝 Translation Guidelines

### 용어 통일

주요 기술 용어는 가급적 원어를 유지하되, 필요시 번역 후 괄호에 원어를 병기합니다.

| 한국어 | English | Japanese | Chinese |
|--------|---------|----------|---------|
| SPEC | SPEC | SPEC | SPEC |
| TAG | TAG | TAG | TAG |
| Alfred | Alfred | Alfred | Alfred |
| 테스트 주도 개발 | Test-Driven Development (TDD) | テスト駆動開発 (TDD) | 测试驱动开发 (TDD) |
| 요구사항 | Requirements | 要件 | 需求 |
| 구현 | Implementation | 実装 | 实现 |

### 문체

- **공손하고 전문적인 톤** 유지
- **2인칭 사용**: "당신"(한국어), "you"(영어), "あなた"(일본어), "您"(중국어)
- **명확하고 간결한 표현** 사용

### 코드 블록

코드 예제는 번역하지 않고 원본 유지:

```python
# Keep code as-is (do not translate comments in code blocks)
def hello_world():
    print("Hello, World!")
```

### 링크 및 참조

- **내부 링크**: 번역된 페이지가 있으면 해당 언어 경로로 변경
  ```markdown
  <!-- Korean -->
  [설치 가이드](getting-started/installation.md)

  <!-- English -->
  [Installation Guide](../en/getting-started/installation.md)
  ```

- **외부 링크**: 가능하면 해당 언어 버전 링크로 변경

## ✅ Quality Checklist

번역 완료 후 다음 사항을 확인하세요:

- [ ] **파일 구조**: 한국어 원본과 동일한 디렉토리 구조 유지
- [ ] **파일 이름**: 원본과 동일한 파일명 사용
- [ ] **마크다운 문법**: 제목, 링크, 코드 블록 등 문법 오류 없음
- [ ] **용어 통일**: 주요 용어가 일관되게 번역됨
- [ ] **코드 유지**: 코드 예제는 원본 그대로 유지
- [ ] **링크 검증**: 내부/외부 링크가 올바르게 작동
- [ ] **로컬 빌드 테스트**: `mkdocs serve`로 렌더링 확인

## 🔍 Testing Your Translation

로컬에서 번역 결과를 확인하려면:

```bash
# 1. Documentation dependencies 설치
cd docs
pip install -r requirements.txt

# 2. MkDocs 개발 서버 실행
mkdocs serve

# 3. 브라우저에서 확인
# http://localhost:8000
```

## 🤝 Review Process

1. **Pull Request 생성**: 번역 완료 후 PR 제출
2. **자동 검증**: CI/CD가 문법 및 링크 검증 자동 수행
3. **리뷰**: 메인테이너 또는 언어별 리뷰어가 검토
4. **수정 요청**: 필요시 피드백 반영
5. **병합**: 승인 후 main 브랜치에 병합

## 📧 Contact

질문이나 도움이 필요하시면:

- **GitHub Issues**: [moai-adk/issues](https://github.com/modu-ai/moai-adk/issues)
- **GitHub Discussions**: [moai-adk/discussions](https://github.com/modu-ai/moai-adk/discussions)

## 🎖️ Contributors

번역에 기여해 주신 분들:

- 기여자 목록은 [Contributors](https://github.com/modu-ai/moai-adk/graphs/contributors)에서 확인할 수 있습니다.

---

**감사합니다!** 여러분의 기여로 MoAI-ADK가 더 많은 사용자에게 다가갈 수 있습니다. 🌏
