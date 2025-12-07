# 한국어 기술 서적 작성 참고 자료

이 문서는 한국어 기술 서적을 집필할 때 참고할 수 있는 공식 자료, 도서, 도구, 플랫폼 등을 정리합니다.

---

## 📚 한국어 기술 문서 작성 가이드

### 공식 자료

1. **국립국어원 표준국어대사전**
   - URL: https://stdict.korean.go.kr/
   - 목적: 한글 맞춤법, 표준어 규정, 띄어쓰기 확인
   - 활용: 전문 용어 한글 표기 검증

2. **국립국어원 한국어교수학습샘터**
   - URL: https://kcenter.korean.go.kr/
   - 목적: 한국어 교육 자료 및 교수법
   - 활용: 교육 콘텐츠 설계 시 참고

3. **2025 전문용어 표준화 안내서**
   - 발행: 국립국어원
   - 목적: 정부 및 공공기관 전문 용어 표준화
   - 활용: 기술 용어의 공식 한글 표기 확인

4. **한글 맞춤법 (제5장-제6장)**
   - URL: https://www.korean.go.kr/
   - 목적: 띄어쓰기 규정 (11개 조항)
   - 핵심 원칙: "모든 단어는 띄어 쓴다"

### IT 용어 표준화 자료

1. **IT 글쓰기와 번역 노트 - 4.1 IT 용어 일반**
   - URL: https://wikidocs.net/67110
   - 목적: 한영 IT 용어 표준 표기 가이드
   - 범위: 270개 이상의 일반 IT 용어
   - 특징:
     - 올바른/잘못된 표기 명시 (예: 아키텍처 ○, 아키텍쳐 ×)
     - 맥락별 번역 제공 (예: function → 함수/기능)
     - 분야별 분류 (Programming, Database, Security 등)
   - 활용: 기술 서적 집필 시 일관된 용어 사용
   - **로컬 파일**: `terminology.md` (270+ 용어 전체 수록)

2. **주요 오류 표기 (절대 금지)**
   - 어플리케이션 (❌) → 애플리케이션 (✅)
   - 아키텍쳐 (❌) → 아키텍처 (✅)
   - 캡쳐 (❌) → 캡처 (✅)
   - 챠트 (❌) → 차트 (✅)
   - 컨텐츠 (❌) → 콘텐츠 (✅)
   - 디렉토리 (❌) → 디렉터리 (✅)
   - 프레임웍 (❌) → 프레임워크 (✅)
   - 메세지 (❌) → 메시지 (✅)
   - 포탈 (❌) → 포털 (✅)
   - 스크린샷 (❌) → 스크린숏 (✅)
   - 쓰레드 (❌) → 스레드 (✅)
   - 스케쥴 (❌) → 스케줄 (✅)

3. **사용 권장 사항**
   - ✅ 집필 시작 전에 `terminology.md` 훑어보기
   - ✅ 불확실한 용어는 즉시 검색 (Ctrl/Cmd + F)
   - ✅ 한 책 내에서 동일 용어의 일관된 번역 유지
   - ✅ 첫 등장 시 영문+한글 병기 (예: 컴포넌트(component))

### 기업 기술 문서 가이드

1. **Kakao Enterprise 기술 문서 작성 가이드**
   - URL: https://tech.kakaoenterprise.com/
   - 주요 시리즈:
     - [TW] 기술문서 쉽게 쓰기 지침
     - [TW] 기술 문서 작성 5단계
     - [TW] 목차의 중요성
   - 핵심 원칙:
     - **명확성**: 첫 독해에서 이해 가능
     - **간결성**: 단순하고 쉬운 단어 사용
     - **독자 중심**: 자신의 생각 정리가 아닌 독자 이해 우선

2. **LINE Engineering Blog - 코드 가독성**
   - URL: https://engineering.linecorp.com/ko/blog/code-readability-vol2/
   - 주제: 명명(naming)과 주석 작성법
   - 활용: 코드 예제 작성 시 참고

### 커뮤니티 가이드

1. **GitHub - 기술 문서 작성 시 주의사항**
   - URL: https://gist.github.com/9beach/41b8b51b13e4704653a8
   - 저자: 9beach
   - 내용:
     - 한영 혼용 규칙
     - Markdown 포매팅
     - 구두점 사용법
     - 띄어쓰기 세부 규칙
   - 특징: 실용적이고 구체적인 예시 제공

2. **DevOcean (SK) - 개발자의 글쓰기**
   - URL: https://devocean.sk.com/blog/techBoardDetail.do?ID=165343
   - 주제: Technical Writing 기초
   - 활용: 기술 블로그를 책으로 확장 시 참고

---

## 📖 추천 도서

### 한국어 기술 글쓰기

1. **Docs for Developers 기술 문서 작성 완벽 가이드**
   - 저자: 자레드 바티(Jared Bhatti) 외
   - 출판사: 한빛미디어
   - ISBN: 9791169210485
   - 주제: 기술 문서 작성 전반
   - 목차:
     - 독자 이해하기
     - 계획하고 작성하기
     - 편집과 수정
     - 문서 조직화
     - 예제와 다이어그램
   - 추천 대상: 테크니컬 라이터, API 문서 작성자

2. **개발자의 글쓰기**
   - 저자: 김철수
   - 목적: 개발자를 위한 실용 글쓰기
   - 활용: 기술 블로그, README, 문서 작성

### 프로그래밍 책 집필

1. **한빛미디어 - 프로그래머의 책쓰기 (GitHub 저장소)**
   - URL: https://github.com/hanbitmedia/Writing-IT-Books
   - 저자: 한빛미디어 편집팀
   - 내용:
     - 책 집필 5단계 (분석-설계-구현-테스트-패턴)
     - 목차 설계 방법
     - 대상 독자 분석
     - 실전 집필 팁
   - 특징: 소프트웨어 개발에 비유한 집필 프레임워크
   - 라이선스: MIT

2. **프로그래밍 패턴: 프로그램을 작성하는 33가지 방법**
   - 출판사: 위키북스
   - 특징: 동일 주제를 33가지 스타일로 작성
   - 활용: 다양한 코드 예제 작성 패턴 학습

3. **게임 프로그래밍 패턴**
   - 출판사: 한빛미디어
   - 구조: 8개 절 (의도, 동기, 패턴, 언제 쓸 것인가, 주의사항, 예제, 디자인 결정, 관련자료)
   - 활용: 패턴 기반 콘텐츠 구조화 참고

### 교육학적 설계

1. **Teaching Tech Together (한국어판)**
   - 저자: Greg Wilson
   - 주제: 프로그래밍 교육 방법론
   - 내용: 학습 곡선 설계, 개념 맵, 평가 전략

---

## 🌐 온라인 리소스

### 기술 글쓰기 가이드

1. **Google Developer Documentation Style Guide**
   - URL: https://developers.google.com/style
   - 언어: 영문 (한국어 참고 가능)
   - 내용: API 문서, 튜토리얼 작성 가이드
   - 활용: 구조와 톤 참고

2. **Microsoft Writing Style Guide**
   - URL: https://learn.microsoft.com/en-us/style-guide/
   - 언어: 영문
   - 활용: 기술 문서 스타일 통일

3. **Write the Docs**
   - URL: https://www.writethedocs.org/
   - 커뮤니티: 글로벌 테크니컬 라이팅 커뮤니티
   - 활용: 베스트 프랙티스, 사례 연구

### 교육 콘텐츠 설계

1. **TAA (Textbook & Academic Authors Association)**
   - URL: https://blog.taaonline.net/
   - 주요 글:
     - "Pedagogy of Book and Chapter Organization"
     - 교육학적 요소 설계
   - 활용: 책 구조 설계 시 참고

2. **Structuring Textbooks with Pedagogical Elements**
   - URL: https://iu.pressbooks.pub/eastcmtf/chapter/structuring-a-textbook-with-pedagogical-elements/
   - 내용:
     - Openers (학습 목표, 도입)
     - Closers (요약, 연습문제)
     - Integrated Devices (사례 연구, 용어)
   - 활용: 장 구조 템플릿 설계

3. **FreeCodeCamp - How to Write Your First Technical Book**
   - URL: https://www.freecodecamp.org/news/how-to-write-your-first-technical-book/
   - 언어: 영문
   - 내용: 기술 서적 집필 전체 프로세스
   - 활용: 초보 저자를 위한 로드맵

4. **PLOS Comp Bio - Ten Simple Rules for Writing a Technical Book**
   - URL: https://journals.plos.org/ploscompbiol/article?id=10.1371/journal.pcbi.1011305
   - 발행: 2023
   - 내용: 기술 서적 집필 10대 원칙
   - 특징: 학술적이면서도 실용적

### 디자인 패턴

1. **Refactoring.Guru - 디자인 패턴 (한국어)**
   - URL: https://refactoring.guru/ko/design-patterns
   - 언어: 한국어 완전 번역
   - 내용:
     - 23가지 디자인 패턴
     - Java, Python, C++, TypeScript 등 코드 예시
     - 시각적 다이어그램
   - 활용: 패턴 기반 코드 예제 작성

2. **Design Patterns (Gang of Four) 번역서**
   - 출판사: 프로텍미디어
   - 활용: 전통적 디자인 패턴 설명 구조 참고

---

## 🛠️ 집필 도구

### Markdown 에디터

1. **Obsidian**
   - URL: https://obsidian.md/
   - 특징:
     - 로컬 파일 기반
     - 링크 그래프 시각화
     - 플러그인 생태계
   - 활용: 장간 연결 관리, 메모 정리

2. **Typora**
   - URL: https://typora.io/
   - 특징:
     - WYSIWYG Markdown 편집
     - 깔끔한 인터페이스
     - 다양한 테마
   - 활용: 직관적 집필 환경

3. **VS Code**
   - 확장:
     - Markdown All in One
     - Markdown Preview Enhanced
     - Korean Language Pack
   - 활용: 코드와 문서 통합 작성

### 버전 관리

1. **Git + GitHub**
   - 전략: Gitbook flow
   - 브랜치:
     - `main`: 출판 버전
     - `chapters/*`: 장별 작업
   - 활용: 변경 이력 추적, 협업

2. **GitBook**
   - URL: https://www.gitbook.com/
   - 특징: Git 기반 웹 출판
   - 활용: 온라인 프리뷰, 베타 리더 공유

### 문서 변환

1. **Pandoc**
   - URL: https://pandoc.org/
   - 기능: Markdown → PDF, EPUB, DOCX, HTML
   - 명령어:
     ```bash
     # PDF 생성
     pandoc book.md -o book.pdf --pdf-engine=xelatex -V mainfont="Noto Sans KR"
     
     # EPUB 생성
     pandoc book.md -o book.epub --toc
     
     # DOCX 생성
     pandoc book.md -o book.docx
     ```

2. **LaTeX (한글 지원)**
   - 패키지: kotex
   - 활용: 고품질 PDF 생성
   - 예시:
     ```latex
     \documentclass[a4paper]{oblivoir}
     \usepackage{kotex}
     ```

### 다이어그램 도구

1. **Mermaid**
   - URL: https://mermaid.js.org/
   - 유형: 텍스트 기반
   - 활용: Markdown 내 다이어그램 삽입
   - 예시:
     ```mermaid
     graph LR
         A[시작] --> B{조건}
         B -->|Yes| C[처리]
         B -->|No| D[종료]
     ```

2. **PlantUML**
   - URL: https://plantuml.com/ko/
   - 유형: 텍스트 기반 UML
   - 활용: 클래스 다이어그램, 시퀀스 다이어그램

3. **Excalidraw**
   - URL: https://excalidraw.com/
   - 유형: 손그림 스타일
   - 활용: 개념 설명용 스케치

4. **draw.io (diagrams.net)**
   - URL: https://app.diagrams.net/
   - 유형: GUI 기반
   - 활용: 복잡한 아키텍처 다이어그램

---

## ✅ 품질 검증 도구

### 맞춤법 검사

1. **부산대학교 한국어 맞춤법/문법 검사기**
   - URL: http://speller.cs.pusan.ac.kr/
   - 기능: 맞춤법, 띄어쓰기, 표준어
   - API: 제공 (대용량 문서 자동 검사)

2. **네이버 맞춤법 검사기**
   - URL: https://search.naver.com/search.naver?query=맞춤법+검사
   - 기능: 실시간 교정
   - 한계: 500자 제한

3. **Hanspell (Python 라이브러리)**
   - 설치: `pip install hanspell`
   - 사용:
     ```python
     from hanspell import spell_checker
     
     result = spell_checker.check("안녕 하세요")
     print(result.checked)  # "안녕하세요"
     ```

4. **VS Code 확장: Korean Spell Checker**
   - 기능: 실시간 맞춤법 하이라이트
   - 활용: 집필 중 즉시 교정

### 가독성 분석

1. **KLI (Korean Legibility Index) 계산기**
   - 기준: 문장 길이, 어휘 난이도
   - 목표치:
     - 초급: KLI ≤ 60
     - 중급: KLI ≤ 75
     - 고급: KLI ≤ 90

2. **텍스트 분석 도구**
   - 평균 문장 길이
   - 어휘 다양성 (TTR: Type-Token Ratio)
   - 한자어/외래어 비율

### 코드 검증

1. **pytest**
   - 목적: 코드 예제 자동 테스트
   - 구조:
     ```
     tests/
     ├── chapter_03/
     │   ├── test_list_examples.py
     │   ├── test_tuple_examples.py
     ```

2. **tox**
   - 목적: 다중 Python 버전 테스트
   - 설정:
     ```ini
     [tox]
     envlist = py39,py310,py311,py312
     ```

3. **Black + Flake8**
   - Black: 코드 포매팅
   - Flake8: 스타일 검사
   - 활용: 일관된 코드 스타일 유지

4. **Bandit**
   - 목적: 보안 취약점 스캔
   - 사용: `bandit -r examples/`

### 링크 검증

1. **markdown-link-check (npm)**
   - 설치: `npm install -g markdown-link-check`
   - 사용: `markdown-link-check **/*.md`
   - 기능: 깨진 링크 자동 탐지

2. **Python 스크립트 (커스텀)**
   ```python
   import requests
   import re
   from pathlib import Path
   
   def check_links(md_file):
       with open(md_file, 'r') as f:
           content = f.read()
       
       urls = re.findall(r'https?://[^\s\)]+', content)
       
       for url in urls:
           try:
               response = requests.head(url, timeout=5)
               if response.status_code >= 400:
                   print(f"❌ {url} - {response.status_code}")
           except Exception as e:
               print(f"❌ {url} - {e}")
   ```

---

## 📱 출판 플랫폼 (한국)

### 전통 출판사 (IT 전문)

1. **한빛미디어**
   - URL: https://www.hanbit.co.kr/
   - 특징: IT 서적 최대 출판사
   - 절차: 기획안 제출 → 검토 → 계약
   - 지원: 전문 편집자, 디자이너, 마케팅

2. **위키북스**
   - URL: https://wikibook.co.kr/
   - 특징: 개발자 친화적
   - 오픈소스: GitHub 연동 출판

3. **제이펍 (J-PUB)**
   - URL: https://jpub.tistory.com/
   - 특징: 번역서 + 국내 저자
   - 강점: 고품질 편집

4. **길벗**
   - URL: https://www.gilbut.co.kr/
   - 시리즈: "Do it!" 시리즈
   - 대상: 초중급 입문서

5. **에이콘출판사**
   - URL: https://www.acornpub.co.kr/
   - 특징: 전문서 중심
   - 강점: 심도 있는 기술서

### 전자책 자가출판

1. **리디북스**
   - URL: https://ridibooks.com/
   - 플랫폼: 한국 최대 전자책 플랫폼
   - 정산: 판매가의 60-70%
   - 형식: EPUB, PDF

2. **크몽 (전자책 출판)**
   - URL: https://kmong.com/
   - 형식: PDF, EPUB
   - 특징: 간단한 출판 프로세스

3. **텀블벅**
   - URL: https://tumblbug.com/
   - 형식: 크라우드펀딩 + 자가출판
   - 활용: 베타 버전 펀딩 후 정식 출간

### 온라인 플랫폼

1. **GitBook**
   - URL: https://www.gitbook.com/
   - 특징: Git 연동, 무료 공개 가능
   - 활용: 웹 버전 먼저 공개 후 종이책 출간

2. **Notion Press**
   - 특징: Notion 기반 출판
   - 플러그인: Super, Notion2Book

---

## 📊 시장 조사 자료

### 한국 IT 서적 시장

1. **교보문고 IT 분야 베스트셀러**
   - URL: https://www.kyobobook.co.kr/
   - 카테고리: IT모바일 > 프로그래밍언어
   - 활용: 트렌드 파악

2. **예스24 IT 신간**
   - URL: https://www.yes24.com/
   - 활용: 경쟁 도서 분석

3. **알라딘 IT/프로그래밍**
   - URL: https://www.aladin.co.kr/
   - 활용: 독자 리뷰 분석

### 트렌드 키워드 (2025)

- Python, JavaScript, TypeScript
- 클라우드 (AWS, GCP, Azure)
- 컨테이너 (Docker, Kubernetes)
- AI/ML (LLM, RAG, Fine-tuning)
- 웹 프레임워크 (React, Next.js, FastAPI)
- 데이터 엔지니어링

---

## 🎓 교육 프로그램

### 온라인 강의

1. **인프런 - 테크니컬 라이팅 가이드**
   - URL: https://www.inflearn.com/course/기술-문서-작성-가이드
   - 강사: 김지선 (테크니컬 라이터)
   - 내용: 기술 문서 작성 기초부터 실전까지

2. **패스트캠퍼스 - 개발자 글쓰기**
   - 내용: README, API 문서, 기술 블로그 작성

### 워크숍/컨퍼런스

1. **Write the Docs Korea**
   - 형식: 연례 컨퍼런스
   - 주제: 테크니컬 라이팅 베스트 프랙티스

---

## 📞 커뮤니티 및 지원

### 온라인 커뮤니티

1. **페이스북 그룹**
   - "한국 테크니컬 라이터"
   - "IT 책 저자 모임"

2. **Slack/Discord**
   - 개발자 커뮤니티 (#writing 채널)

3. **오픈 카톡방**
   - "기술 블로그 작가 모임"

### 베타 리더 모집

1. **당근마켓 (지역 커뮤니티)**
2. **오픈채팅 (카카오톡)**
3. **Reddit - r/Korea**

---

## 🔖 관련 Skill

- `yoda-korean-technical-book-writing`: 한국어 기술 서적 작성 (이 스킬)
- `moai-library-mermaid`: Mermaid 다이어그램 활용
- `yoda-writing-templates`: 효과적인 코드 예제 및 템플릿 작성
- `yoda-system`: 교육 콘텐츠 설계 및 템플릿 시스템

---

**업데이트 주기**: 분기별 (매 3개월)
**다음 업데이트**: 2026-02-21
**피드백**: GitHub Issues 또는 이메일

---

**연구 출처**:
- Kakao Enterprise 기술 문서 가이드 (2025)
- 국립국어원 표준화 안내서 (2025)
- 한빛미디어 프로그래머의 책쓰기 (GitHub)
- TAA Pedagogy of Book Organization
- PLOS Comp Bio - Ten Simple Rules
- GitHub - Korean Technical Writing Guidelines

---

## 📚 한국어 온라인 기술 서적 (Best Practice 예시)

### 1. 점프 투 파이썬 (Jump to Python)

- **URL**: https://wikidocs.net/book/1
- **저자**: 박응용
- **플랫폼**: WikiDocs
- **최초 발행**: 2005년
- **최근 업데이트**: 2025년 현재까지 지속 업데이트
- **특징**: 한국에서 가장 성공적인 무료 온라인 프로그래밍 교재

**핵심 강점**:
1. **대화형 문체**: "우리가", "당신은", "~해 보자" 등으로 독자와 대화
2. **실습 중심**: "말로 설명하는 것보다 직접 실행해 보면서" 철학
3. **점진적 학습**: 간단한 예제부터 시작하여 단계적으로 복잡도 증가
4. **인터프리터 스타일**: `>>>` 프롬프트로 실제 실행 과정 재현
5. **비유 활용**: "기초 공사 없이 빌딩 세우기" 같은 일상적 은유

**교육학적 패턴**:
- **오류 먼저 보여주기**: 초보자가 겪을 실수를 미리 시연하고 해결
- **개념 반복 강조**: 중요한 원칙을 굵게 표시하고 여러 번 반복
- **박스형 보충**: 본문 흐름을 방해하지 않는 추가 설명 제공
- **목표 중심 제목**: "~하기", "~마스터하기" 형식

**적용 예시**:
```markdown
## 리스트와 튜플의 차이

당신은 리스트와 튜플 중 무엇을 사용해야 할지 고민한 적이 있는가?

>>> my_list = [1, 2, 3]
>>> my_list[0] = 99
>>> print(my_list)
[99, 2, 3]

리스트는 수정이 가능하다. 이번에는 튜플을 보자.

>>> my_tuple = (1, 2, 3)
>>> my_tuple[0] = 99
TypeError: 'tuple' object does not support item assignment

**튜플은 한번 생성하면 수정할 수 없다.** 이것이 가장 중요한 차이점이다.
```

**대상 독자**: 프로그래밍 입문자, 비전공자, 독학자
**추천 활용**: 입문서 집필 시 문체 및 예제 구성 참고

---

### 2. 생활코딩 (opentutorials.org)

- **URL**: https://opentutorials.org/
- **저자**: 이고잉(egoing)
- **형식**: 동영상 + 텍스트
- **특징**: 프로그래밍 입문자를 위한 무료 강의

**강점**:
- 쉬운 언어로 복잡한 개념 설명
- 실생활 예시 중심
- 커뮤니티 활성화

**적용**: 비유와 설명 방식 참고

---

### 3. WikiDocs 플랫폼 전체

- **URL**: https://wikidocs.net/
- **형식**: Markdown 기반 위키 문서
- **특징**: 한국어 기술 문서 공유 플랫폼

**인기 도서**:
- 점프 투 파이썬
- 점프 투 장고 (Django)
- 점프 투 플라스크 (Flask)
- 왕초보를 위한 Python

**장점**:
- 버전 관리 지원
- 독자 댓글/피드백 기능
- 무료 배포 가능
- Markdown 포맷

**적용**: 베타 버전 온라인 배포 플랫폼으로 활용

