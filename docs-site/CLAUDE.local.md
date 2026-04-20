# 프로젝트 로컬 설정

이 파일은 개인 프로젝트 설정을 위한 것입니다. MoAI-ADK 업데이트 시 덮어쓰기되지 않습니다.

## 프로젝트 URL 정보

**중요:** 모든 문서와 코드에서 올바른 URL을 사용하세요.

| 리소스 | URL |
|--------|-----|
| **문서 홈페이지** | `https://adk.mo.ai.kr` |
| GitHub 저장소 | `https://github.com/modu-ai/moai-adk` |
| Discord 커뮤니티 | `https://discord.gg/moai-adk` |
| NPM 패키지 | `https://www.npmjs.com/package/moai-adk` |

### 잘못된 URL (사용 금지)
- ❌ `docs.moai-ai.dev` (구 주소, 사용하지 마세요)
- ❌ `adk.moai.com`
- ❌ `adk.moai.kr` (오타: `adk.mo.ai.kr` 가 정확한 주소입니다)

### URL 업데이트 체크리스트

URL을 변경할 때 다음 파일들을 확인하세요:
- `theme.config.tsx` - og:url, og:image, alternate 링크
- `public/robots.txt` - Sitemap URL
- `public/llms.txt` - 문서 링크
- 메타 태그의 모든 URL 참조

## 문서 작성 가이드라인

### MDX 렌더링 오류 방지

**마크다운 강조(**)와 특수문자 사이에 공백 필수**

MDX에서 강조 표시(`**text**`)와 특수문자(괄호, 따옴표 등)를 직접 붙이면 렌더링 오류가 발생합니다.

| ❌ 잘못된 표기 | ✅ 올바른 표기 |
|---------------|---------------|
| `**바이브코딩(Vibe Coding)**` | `**바이브코딩** (Vibe Coding)` |
| `**SPEC(Specification)**` | `**SPEC** (Specification)` |
| `**DDD(Domain-Driven)**` | `**DDD** (Domain-Driven)` |

**규칙:** 강조 표시와 괄호 사이에 반드시 공백을 넣으세요.

### Mermaid 다이어그램 방향

모든 Mermaid 다이어그램은 **세로 방향(Top-Down)**을 사용하세요:

- 사용: `flowchart TD` 또는 `graph TB`
- 지양: `flowchart LR` 또는 `graph LR`

**예시:**

```mermaid
\`\`\`mermaid
flowchart TD
    A[시작] --> B[단계 1]
    B --> C[단계 2]
    C --> D[완료]
\`\`\`
```

**이유:** 문서에서 세로 방향이 가독성이 더 좋으며 모바일에서도 잘 표시됩니다.

### 지원 언어

문서에서 다음 언어만 사용합니다:

- 🇰🇷 한국어 (Korean)
- 🇺🇸 영어 (English)
- 🇯🇵 일본어 (Japanese)
- 🇨🇳 중국어 (Chinese)

### 터미널 명령어 표기

터미널에서 입력하는 명령어는 백틱 코드 블록으로 표기합니다:

```bash
moai init my-project
```
