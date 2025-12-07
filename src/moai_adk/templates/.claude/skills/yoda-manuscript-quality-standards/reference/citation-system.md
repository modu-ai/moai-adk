# 인용 시스템

**목적**: 출처 투명성 및 신뢰도 확보

---

## 1. 인용 번호 규칙

### 본문 작성 시

- 출처가 있는 정보는 문장/문단 끝에 `(숫자)` 형식으로 인용 번호 추가
- 인용 번호는 해당 절(Section) 내에서 순차적으로 부여
- 동일 출처를 여러 번 인용하면 같은 번호 사용

### 예시

```markdown
Claude Code는 Anthropic이 2024년 중반에 출시한 터미널 기반 AI 코딩 어시스턴트다 (1).
이 도구는 200K 토큰의 컨텍스트 윈도우를 지원하여 대규모 프로젝트를 효과적으로 처리할 수 있다 (1).
2025년 11월 기준, 전 세계 개발자의 78%가 AI 코드 도구를 하나 이상 사용하고 있다 (2).
```

---

## 2. 코드 예제 인용

공식 문서나 GitHub 저장소의 코드를 사용할 경우:

```python
# 출처: Anthropic Claude Code 공식 문서 (3)
def hello_world():
    print("Hello, World!")
```

---

## 3. 인용문 섹션 (절 하단)

**모든 절(Section) 하단에 "인용문" 섹션 추가**:

**⚠️ 공식 출처만 인용 가능**

```markdown
## 인용문

1. Anthropic. "Claude Code Overview". *Claude Code Official Documentation*. 2025-11-20. [https://docs.anthropic.com/claude-code/overview](https://docs.anthropic.com/claude-code/overview)

2. Anthropic. "Claude Code Release Notes". *Anthropic Official Blog*. 2025-11-15. [https://www.anthropic.com/news/claude-code-release](https://www.anthropic.com/news/claude-code-release)

3. Anthropic. "Claude Code Examples". *GitHub - anthropics/claude-code*. 2025-11-18. [https://github.com/anthropics/claude-code/examples](https://github.com/anthropics/claude-code/examples)

4. Stack Overflow. "Developer Survey 2025". *Stack Overflow Official Blog*. 2025-05-15. [https://stackoverflow.blog/2025/developer-survey-results](https://stackoverflow.blog/2025/developer-survey-results)
```

**형식**:
```
번호. 저자/기관. "제목". *출처*. 발행일. [URL](URL)
```

**허용되는 인용문 예시**:
- ✅ Anthropic 공식 문서: docs.anthropic.com, docs.claude.com
- ✅ Anthropic 공식 블로그: anthropic.com/news, anthropic.com/research
- ✅ Anthropic 공식 GitHub: github.com/anthropics
- ✅ Stack Overflow 공식 통계: stackoverflow.blog (Developer Survey만)
- ✅ 주요 기술 언론사: techcrunch.com, theverge.com (Anthropic 공식 발표만)

**절대 사용 금지 예시**:
- ❌ Medium 개인 블로그: medium.com/@개인계정
- ❌ Dev.to 개인 글: dev.to/개인계정
- ❌ Reddit 커뮤니티: reddit.com
- ❌ 비공식 튜토리얼: claudelog.com, opentools.ai, 기타 제3자 사이트
- ❌ AI 생성 콘텐츠 사이트

---

## 4. 인용이 필요한 정보

### 반드시 인용해야 하는 정보

- ✅ 통계 및 수치 ("78% 개발자가...")
- ✅ 공식 정의 및 용어
- ✅ 기술 사양 ("200K 토큰", "v1.0.58")
- ✅ 출시일 및 버전 정보
- ✅ 공식 명령어 및 코드
- ✅ 벤치마크 결과
- ✅ 다른 도구와의 비교

### 인용 불필요한 정보

- ❌ 일반 상식 ("코딩은 프로그래밍 언어로...")
- ❌ 저자의 해석 및 의견
- ❌ 비유 및 예시 ("마치 자율주행 자동차처럼...")

---

**버전**: 1.0.0
**최종 업데이트**: 2025-11-27
