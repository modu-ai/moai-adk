# 한국어 기술 서적 작성 Skill

**Version**: 1.0.0  
**Status**: Production  
**Last Updated**: 2025-11-21

## 📋 개요

이 Skill은 전문적인 한국어 기술 서적(특히 프로그래밍/IT 분야)을 집필하기 위한 종합 가이드입니다. 2025년 현재 기준 최신 작성 원칙, 교육학적 패턴, 한글 타이포그래피 규칙을 제공합니다.

## 🎯 이 Skill로 할 수 있는 것

- ✅ 독자 중심의 명확하고 간결한 한국어 기술 문체 작성
- ✅ MECE 원칙에 따른 체계적인 목차 구조 설계
- ✅ 초급-중급-고급 단계별 학습 곡선 설계
- ✅ 효과적인 코드 예제 및 주석 작성
- ✅ 교육학적 콘텐츠 패턴 (Openers, Closers, Integrated Devices) 적용
- ✅ 한글 맞춤법, 띄어쓰기, 구두점 규칙 준수
- ✅ 품질 보증 체크리스트 기반 검증

## 📦 패키지 구조

```
yoda-korean-technical-book-writing/
├── SKILL.md                    # 메인 Skill 정의 (926줄)
├── examples.md                 # 실전 예시 10가지 (1,068줄)
├── reference.md                # 참고 자료 및 도구 (539줄)
├── templates/
│   ├── chapter-template.md     # 장(Chapter) 구조 템플릿
│   ├── section-template.md     # 섹션(Section) 구조 템플릿
│   └── code-example-template.md # 코드 예제 작성 템플릿
└── README.md                   # 이 파일
```

**총 라인 수**: 2,533줄 (메인 파일) + 템플릿 파일

## 🚀 빠른 시작

### 1. Skill 활성화

Claude Code에서 이 Skill을 자동으로 인식합니다. 다음과 같이 요청하면 됩니다:

```
"Python 초보자를 위한 기술 서적 1장을 작성해줘"
"코드 예제에 효과적인 주석을 다는 방법을 알려줘"
"책의 목차를 MECE 원칙에 따라 검토해줘"
```

### 2. 템플릿 사용

새로운 장을 시작할 때:

```bash
# 장 템플릿 복사
cp .claude/skills/yoda-korean-technical-book-writing/templates/chapter-template.md chapter_01.md

# 편집 시작
vim chapter_01.md
```

## 📚 주요 섹션

### Section 1: 한국어 기술 문체 가이드
- 존댓말 vs 반말 선택 기준
- 문장 구조 최적화 (15-25어절 권장)
- 전문 용어 한글화 vs 영문 병기
- 단락 구성 전략 (주제문-지원-결론)

### Section 2: 책 구조 베스트 프랙티스
- MECE 원칙 기반 목차 설계
- Chapter 구조 템플릿 (학습 목표 → 도입 → 본문 → 요약 → 연습문제)
- Section 구조 패턴 (Why → What → How)
- 부록 및 색인 작성 가이드

### Section 3: 콘텐츠 타입별 작성 패턴
- 개념 설명 패턴 (Analogy-First)
- 코드 예제 패턴 (Progressive Disclosure)
- 연습 문제 설계 (5단계 난이도)
- 다이어그램 통합 전략

### Section 4: 한글 타이포그래피 규칙
- 맞춤법 및 띄어쓰기 (국립국어원 표준)
- 구두점 사용법 (마침표, 쉼표, 괄호)
- 숫자 및 단위 표기
- Markdown 포매팅 규칙

### Section 5: 교육학적 전략
- 학습 곡선 설계 (초급 40% - 중급 35% - 고급 25%)
- 개념 강화 전략 (Spiral Learning, Just-in-Time)
- 독자 참여 유도 (Callout Box, 인터랙티브 요소)
- 실전 프로젝트 통합 (점진적 확장)

### Section 6: 품질 보증 체크리스트
- 기술적 정확성 (코드 테스트 자동화)
- 링크 및 참조 검증
- 일관성 검사 (용어, 문체)
- 가독성 평가 (KLI 지수)

## 🌟 주요 특징

### 1. 연구 기반 콘텐츠
10개 이상의 공식 출처를 기반으로 작성:
- Kakao Enterprise 기술 문서 가이드 (2025)
- 국립국어원 한글 맞춤법 및 표준화 안내서 (2025)
- 한빛미디어 프로그래밍 책 집필 가이드
- TAA (Textbook & Academic Authors Association)
- PLOS Comp Bio - Ten Simple Rules for Technical Books

### 2. 실전 예시 10가지
- 잘 구조화된 전체 장 예시
- Good vs Bad 문체 비교
- 목차 구조 비교 (불균형 vs 균형)
- Callout Box 활용 사례
- 연습 문제 난이도 계층
- 다이어그램 통합 예시
- 실전 프로젝트 점진적 확장
- 요약 섹션 포맷
- 용어 사전(Glossary) 포맷

### 3. 실용적 템플릿
- **chapter-template.md**: 학습 목표, 도입, 본문, 실전 프로젝트, 요약, 연습문제 포함
- **section-template.md**: Why-What-How 구조, 코드 예제, Try It 실습 포함
- **code-example-template.md**: 7가지 코드 작성 패턴, 주석 가이드, 체크리스트

### 4. 도구 및 리소스
- 맞춤법 검사: 부산대 맞춤법 검사기, Hanspell
- 코드 검증: pytest, tox, Black, Bandit
- 문서 변환: Pandoc (Markdown → PDF/EPUB/DOCX)
- 다이어그램: Mermaid, PlantUML, Excalidraw

## 💡 사용 사례

### 사례 1: 첫 번째 기술 서적 집필
```
User: "Python 초보자를 위한 책을 쓰려고 하는데, 목차를 어떻게 구성해야 할까요?"

Claude (with this Skill):
- MECE 원칙 기반 목차 설계 제안
- 초급 40% - 중급 35% - 고급 25% 비율 권장
- 장별 절 개수 균형 (6-8개) 조언
- Part 1 (기초) - Part 2 (핵심) - Part 3 (실전) 구조 제시
```

### 사례 2: 코드 예제 품질 향상
```
User: "코드 예제에 주석을 어떻게 달아야 할까요?"

Claude (with this Skill):
- "왜"를 설명하는 주석 가이드 제공
- Good vs Bad 예시 비교
- 7가지 코드 예제 패턴 제시
- Progressive Enhancement (기본 → 응용 → 고급) 구조 제안
```

### 사례 3: 한국어 문체 개선
```
User: "기술 문서를 쓸 때 존댓말을 써야 할까요, 반말을 써야 할까요?"

Claude (with this Skill):
- 입문서: 존댓말 (합니다체) 권장 - 친근감
- 레퍼런스: 반말 (한다체) 권장 - 간결성
- 튜토리얼: 존댓말 (해요체) 권장 - 대화형
- API 문서: 반말 (명사형) 권장 - 객관성
```

## 📊 품질 메트릭

이 Skill을 사용하면 다음 품질 지표를 달성할 수 있습니다:

| 메트릭 | 목표 | 검증 방법 |
|--------|------|-----------|
| 문장 길이 | 15-25어절 | 자동 분석 도구 |
| 맞춤법 정확도 | 99% 이상 | 부산대 맞춤법 검사기 |
| 코드 실행 성공률 | 100% | pytest 자동 테스트 |
| 링크 유효성 | 100% | markdown-link-check |
| 용어 일관성 | 100% | 커스텀 스크립트 |
| 가독성 (KLI) | ≤ 75 (중급) | KLI 계산기 |

## 🛠️ 관련 도구

### 집필 환경
- **Markdown 에디터**: Obsidian, Typora, VS Code
- **버전 관리**: Git + GitHub
- **문서 변환**: Pandoc

### 품질 검증
- **맞춤법**: 부산대 맞춤법 검사기, Hanspell
- **코드 검증**: pytest, tox, Black, Flake8, Bandit
- **링크 검사**: markdown-link-check

### 협업 및 피드백
- **베타 리더**: Google Docs (댓글 기능)
- **기술 리뷰**: GitHub Pull Request
- **독자 피드백**: Google Forms

## 🔗 관련 Skill

- `yoda-korean-technical-book-writing`: 한국어 기술 서적 작성 (이 스킬)
- `moai-library-mermaid`: Mermaid 다이어그램 활용
- `yoda-writing-templates`: 효과적인 코드 예제 및 템플릿 작성
- `yoda-system`: 교육 콘텐츠 설계 및 템플릿 시스템

## 📖 참고 자료

### 공식 문서
- [국립국어원 표준국어대사전](https://stdict.korean.go.kr/)
- [한글 맞춤법](https://www.korean.go.kr/)
- [Kakao Enterprise 기술 문서 가이드](https://tech.kakaoenterprise.com/)

### 도서
- Docs for Developers 기술 문서 작성 완벽 가이드 (한빛미디어)
- 한빛미디어 - 프로그래머의 책쓰기 (GitHub)

### 온라인 리소스
- [FreeCodeCamp - How to Write Your First Technical Book](https://www.freecodecamp.org/news/how-to-write-your-first-technical-book/)
- [PLOS Comp Bio - Ten Simple Rules](https://journals.plos.org/ploscompbiol/article?id=10.1371/journal.pcbi.1011305)

## 🤝 기여 및 피드백

이 Skill은 지속적으로 개선됩니다. 피드백이나 제안 사항이 있다면:

- GitHub Issues
- 이메일: [maintainer email]
- MoAI-ADK 커뮤니티

## 📝 버전 히스토리

- **1.0.0** (2025-11-21): 초기 버전 공개
  - 6개 주요 섹션 (문체, 구조, 콘텐츠, 타이포그래피, 교육, 품질)
  - 10개 실전 예시
  - 3개 템플릿 (장, 섹션, 코드)
  - 종합 참고 자료 및 도구 가이드

## 📄 라이선스

MIT License

## 👤 작성자

**MoAI-ADK Team**  
Maintainer: GOOS

---

**연구 출처**:
- Kakao Enterprise Technical Writing Guidelines (2025)
- 국립국어원 한글 맞춤법 및 전문용어 표준화 안내서 (2025)
- 한빛미디어 프로그래머의 책쓰기 (GitHub)
- TAA - Pedagogy of Book and Chapter Organization
- PLOS Computational Biology - Ten Simple Rules for Writing a Technical Book
- GitHub - Korean Technical Writing Best Practices (9beach)
- FreeCodeCamp - How to Write Your First Technical Book
- Structuring Textbooks with Pedagogical Elements (Indiana University)

**마지막 업데이트**: 2025-11-21
