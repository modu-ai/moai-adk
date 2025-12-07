# 사전 자료 조사 프로세스

**목적**: 공식 출처 기반 신뢰할 수 있는 자료 수집

---

## 1. 필수 조사 단계 (각 절 작성 전)

**모든 절(Section) 작성 전에 다음 순서로 자료를 수집한다:**

### Step 1: 공식 문서 우선 (Context7 MCP)

```python
mcp__context7__resolve-library-id("Claude Code")
mcp__context7__get-library-docs(context7CompatibleLibraryID, topic)
```

**조사 항목**:
- 공식 API 문서
- 공식 가이드 및 튜토리얼
- 릴리스 노트 및 변경 사항
- 베스트 프랙티스

### Step 2: 공식 출처만 허용 (WebSearch)

**⚠️ 중요: 공식 자료만 인용한다**

```python
WebSearch(query, allowed_domains=[
  "anthropic.com",
  "docs.anthropic.com",
  "docs.claude.com",
  "claude.ai",
  "github.com/anthropics"
])
```

**허용 도메인 (공식 출처만)**:
- ✅ Anthropic 공식 사이트: anthropic.com
- ✅ Claude Code 공식 문서: docs.anthropic.com, docs.claude.com
- ✅ Claude 공식 사이트: claude.ai
- ✅ Anthropic 공식 GitHub: github.com/anthropics
- ✅ Anthropic 공식 블로그: anthropic.com/news, anthropic.com/research

**추가 허용 (예외적으로 사용 가능)**:
- ⚠️ Stack Overflow 공식 설문조사: stackoverflow.blog (공식 통계만)
- ⚠️ 주요 기술 언론사: techcrunch.com, theverge.com (Anthropic 관련 공식 발표만)

**절대 금지 도메인**:
- ❌ 개인 블로그 (Medium, Dev.to 등 개인 작성 콘텐츠)
- ❌ 개발자 커뮤니티 (Reddit, Hacker News 등)
- ❌ AI 생성 콘텐츠 사이트
- ❌ 비공식 튜토리얼 사이트
- ❌ 광고성 사이트
- ❌ 검증되지 않은 제3자 리뷰

**원칙**:
1. **공식 문서 우선**: Anthropic이 직접 작성하고 검증한 자료만 사용
2. **공식 발표 확인**: 기술 언론사의 경우 Anthropic의 공식 발표문을 인용한 것인지 확인
3. **의심스러우면 제외**: 출처가 불명확하거나 공식 확인이 안 되면 사용하지 않음

### Step 3: 최신성 검증

- 발행일 확인 (최근 6개월 이내 우선)
- 버전 정보 확인 (Claude Code v1.0+ 기준)
- 최신 업데이트 반영 여부

### Step 4: 공식 출처 검증 (필수)

- **공식 문서 우선**: Anthropic 공식 문서에서 1차 확인
- **공식 GitHub 확인**: github.com/anthropics에서 코드 예제 및 릴리스 노트 확인
- **상충 시 공식 문서 우선**: 제3자 자료와 공식 문서가 다르면 항상 공식 문서 채택
- **불확실한 정보 제외**: 공식 출처에서 확인되지 않으면 작성하지 않음
- **개인 블로그/커뮤니티 인용 금지**: 아무리 좋은 내용이어도 공식 출처가 아니면 사용 불가

---

## 2. 자료 조사 문서화

각 절 작성 전 다음 형식으로 조사 결과 기록:

```markdown
## Section X-Y 자료 조사 결과

### 공식 문서 (Context7 MCP)
1. [제목](URL) - 핵심 내용 요약
   - 인용 가능 정보: ...
   - 버전: Claude Code v1.0.58
   - 발행일: 2025-11-20
   - 출처 유형: ✅ Anthropic 공식 문서

### 공식 GitHub (WebSearch or Context7)
2. [제목](URL) - 핵심 내용 요약
   - 저자/출처: github.com/anthropics/claude-code
   - 발행일: 2025-11-15
   - 출처 유형: ✅ Anthropic 공식 GitHub

### 공식 블로그/뉴스 (WebSearch)
3. [제목](URL) - 핵심 내용 요약
   - 저자/출처: Anthropic 공식 블로그 (anthropic.com/news)
   - 발행일: 2025-11-10
   - 출처 유형: ✅ Anthropic 공식 발표

### 공식 통계 (예외적 허용)
- 통계1: "84% 개발자가 AI 도구 사용" - 출처: Stack Overflow Developer Survey 2025 (공식 통계)
- 통계2: "51% 매일 사용" - 출처: Stack Overflow Developer Survey 2025 (공식 통계)

⚠️ **중요**: 개인 블로그, 커뮤니티, 비공식 튜토리얼 사이트의 내용은 절대 인용하지 않음
```

---

**버전**: 1.0.0
**최종 업데이트**: 2025-11-27
