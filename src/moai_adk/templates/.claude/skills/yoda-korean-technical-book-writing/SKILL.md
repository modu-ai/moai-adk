---
name: yoda-korean-technical-book-writing
version: 1.0.0
category: education
status: active
description: 한국어 기술 서적 작성 베스트 프랙티스. 문체 선택, 목차 구성, 학습 곡선 설계, 코드 예제 작성, 교육 패턴 적용 등 전문적인 기술 도서 집필을 위한 종합 가이드
allowed-tools: Read
created: 2025-11-21
updated: 2025-11-27
tags: Korean, technical-writing, books, pedagogy, style-guide
primary-agents: yoda-book-author
dependencies: yoda-writing-templates, moai-library-mermaid, yoda-system
---

# 한국어 기술 서적 작성 베스트 프랙티스

## 📘 Skill 개요

이 Skill은 고품질 한국어 기술 서적(특히 프로그래밍/IT 분야)을 집필하기 위한 종합 가이드입니다.
2025년 현재 기준 최신 작성 원칙, 교육학적 패턴, 한글 타이포그래피 규칙, 그리고 실전 검증된 구조화 방법론을 제공합니다.

**핵심 가치**:
- ✅ 독자 중심 집필 (Reader-Centric Authoring)
- ✅ 점진적 학습 곡선 설계 (Progressive Learning Curve)
- ✅ 실용적 코드 예제 패턴 (Practical Code Examples)
- ✅ 한국어 문체 최적화 (Korean Style Optimization)
- ✅ 교육학적 콘텐츠 구조화 (Pedagogical Content Structure)

**적용 대상**:
- 첫 번째 기술 서적을 집필하는 개발자
- 기존 블로그/강의를 책으로 확장하려는 저자
- 기술 문서 품질을 향상시키려는 테크니컬 라이터
- 교육 콘텐츠를 체계화하려는 강사

---

## Quick Reference

### 한국어 기술 문체 가이드

**문체 선택**:
- 입문서: 존댓말 (합니다체)
- 레퍼런스: 반말 (한다체)
- 튜토리얼: 존댓말 (해요체)

**핵심 원칙**:
- 명확성: 첫 독해에서 이해 가능
- 간결성: 15-25어절 (최대 35어절)
- 전문 용어: 한글+영문 병기 (첫 등장)

See [Style Guide Module](modules/style-guide.md) for comprehensive writing guidelines.

### 책 구조 베스트 프랙티스

**목차 설계** (MECE 원칙):
- Part: 대주제 구분 (10+ 장)
- Chapter: 4-8개 권장
- Section: 장당 5-9개 권장

**장 구조**:
```markdown
# Chapter N. [제목]
├─ 학습 목표
├─ 도입
├─ 핵심 개념 (3-5개)
├─ 실전 프로젝트
├─ 요약
├─ 연습 문제
└─ 더 읽을거리
```

See [Book Structure Module](modules/book-structure.md) for detailed chapter templates.

### 콘텐츠 타입별 패턴

**개념 설명**: Analogy-First 접근법
**코드 예제**: Progressive Disclosure (최소 → 실용 → 프로덕션)
**연습 문제**: 5단계 난이도 (기억 → 이해 → 적용 → 분석 → 창조)

See [Content Patterns Module](modules/content-patterns.md) for writing patterns.

### 한글 타이포그래피

**맞춤법 및 띄어쓰기**:
- 기본: 모든 단어는 띄어 쓴다
- 조사: 앞말에 붙여 쓴다
- 의존 명사: 띄어 쓴다

**구두점**:
- 마침표: 문장 종결 필수
- 쉼표: 나열/접속/종속
- 괄호: 한글에 붙여, 영문과 띄움

See [Typography Module](modules/typography-rules.md) for complete rules.

### 품질 보증

**코드 검증**:
```bash
pytest tests/code_examples/
tox -e py39,py310,py311,py312
bandit -r examples/
```

**최종 체크리스트**:
- [ ] 모든 코드 실행 가능
- [ ] 용어 일관성 검증
- [ ] 링크 유효성 확인
- [ ] 맞춤법 검사 완료

See [Quality Assurance Module](modules/quality-assurance.md) for comprehensive checklist.

---

## Module References

### Complete Documentation

- [Style Guide](modules/style-guide.md) - 한국어 기술 문체, 문장 구조, 전문 용어 처리
- [Book Structure](modules/book-structure.md) - 목차 설계, Chapter/Section 템플릿, 부록 구성
- [Content Patterns](modules/content-patterns.md) - 개념 설명, 코드 예제, 연습 문제, 다이어그램
- [Typography Rules](modules/typography-rules.md) - 맞춤법, 띄어쓰기, 구두점, Markdown 포매팅
- [Educational Strategies](modules/educational-strategies.md) - 학습 곡선, 개념 강화, 독자 참여, 실전 프로젝트
- [Quality Assurance](modules/quality-assurance.md) - 기술적 정확성, 일관성, 가독성, 출판 전 체크리스트

### Supporting Files

- [IT Terminology](terminology.md) - 270+ IT 용어 한글 표기 사전

---

## Tools & Resources

### 추천 도구

**집필 환경**:
- Markdown 에디터: Obsidian, Typora, VS Code
- 버전 관리: Git + GitHub
- 문서 변환: Pandoc (Markdown → PDF/EPUB/DOCX)

**품질 검증**:
- 맞춤법: 부산대 맞춤법 검사기, Hanspell
- 코드 검증: pytest, tox, pre-commit hooks
- 링크 검사: markdown-link-check

### 참고 자료

**한국어 기술 문서 작성**:
- [Kakao Enterprise 기술 문서 가이드](https://tech.kakaoenterprise.com/)
- [국립국어원 표준국어대사전](https://stdict.korean.go.kr/)
- [GitHub - 기술 문서 작성 시 주의사항](https://gist.github.com/9beach/41b8b51b13e4704653a8)

**프로그래밍 책 집필**:
- [한빛미디어 - 프로그래머의 책쓰기](https://github.com/hanbitmedia/Writing-IT-Books)
- [FreeCodeCamp - How to Write Your First Technical Book](https://www.freecodecamp.org/news/how-to-write-your-first-technical-book/)

**출판 플랫폼 (한국)**:
- 위키북스, 한빛미디어, 제이펍 (전문 출판사)
- 에이콘출판사, 길벗 (IT 전문)

---

## Quick Start Examples

### 예시 1: 책 프로젝트 초기화

```bash
# 프로젝트 구조 생성
mkdir my-python-book
cd my-python-book

# 기본 파일 생성
cat > structure.md << 'EOF'
# Python 마스터하기

## Part 1. 기초
- Chapter 1. 시작하기
- Chapter 2. 변수와 자료형
- Chapter 3. 제어문

## Part 2. 중급
- Chapter 4. 함수
- Chapter 5. 클래스
- Chapter 6. 모듈
EOF

# 장별 파일 생성
for i in {1..6}; do
  touch "chapter_$i.md"
done
```

### 예시 2: 장 템플릿 적용

See [examples.md](examples.md) for:
- Complete chapter template
- Code example patterns
- Exercise design patterns
- Terminology consistency checking

---

## Related Skills

- `yoda-korean-technical-book-writing`: 한국어 기술 서적 작성 (이 스킬)
- `moai-library-mermaid`: Mermaid 다이어그램 활용
- `yoda-writing-templates`: 효과적인 코드 예제 및 템플릿 작성
- `yoda-system`: 교육 콘텐츠 설계 및 템플릿 시스템

---

## Version History

- **1.0.0** (2025-11-21): 초기 버전 공개
  - 한국어 기술 문체 가이드
  - 책 구조 베스트 프랙티스
  - 교육학적 콘텐츠 패턴
  - 타이포그래피 규칙
  - 품질 보증 체크리스트

---

**Research Sources**:
- Kakao Enterprise Technical Writing Guidelines (2025)
- 국립국어원 한글 맞춤법 (2025 개정)
- 한빛미디어 프로그래밍 책 집필 가이드
- TAA Pedagogy of Book Organization
- National Institute of Korean Language Terminology Standards

**Author**: MoAI-ADK Team
**License**: MIT
**Maintainer**: GOOS
