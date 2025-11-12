# Notion MCP 설정 가이드

**최종 업데이트**: 2025년 1월 (Notion API v2025-09-03)

## 📋 설정 요약

MoAI-ADK 프로젝트에 Notion MCP (Model Context Protocol)가 다음과 같이 설정되었습니다:

### 환경 설정

**파일**: `.env`
```bash
NOTION_TOKEN=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
```

### MCP 설정

**파일**: `.claude/mcp.json`
```json
{
  "notion": {
    "command": "npx",
    "args": ["-y", "@notionhq/client"],
    "enabled": true,
    "description": "Notion - Database and page creation API (API v2025-09-03)",
    "env": {
      "NOTION_TOKEN": "${NOTION_TOKEN}"
    }
  }
}
```

### 보안 설정

**파일**: `.gitignore` (라인 83)
```gitignore
.env*  # .env 파일은 Git에서 무시됨
```

## 🔐 토큰 정보

### 토큰 형식
- **형식**: `ntn_` 접두사 (2024년 9월 25일 이후 신규 생성)
- **생성**: Notion 개발자 포털 → My Integrations
- **유형**: Internal Integration Token

### 토큰 특징
- ✅ 최신 형식의 보안 강화된 토큰
- ✅ 모든 Notion API와 호환
- ✅ MCP 클라이언트 완전 호환
- ✅ Bearer 토큰 인증 방식 사용

## 🔄 API 호출 방식

### Authorization 헤더
```http
Authorization: Bearer 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
```

### 필수 헤더
```json
{
  "Authorization": "Bearer {NOTION_TOKEN}",
  "Notion-Version": "2025-09-03",
  "Content-Type": "application/json"
}
```

## 📚 주요 변경사항 (2025-09-03)

### Multi-Source Database 지원
- 데이터베이스가 여러 데이터 소스(테이블)를 포함 가능
- 새로운 엔드포인트: `/v1/data_sources/{id}`
- 기존 호환성: 유지됨

### API 엔드포인트
| 기존 | 신규 | 설명 |
|------|------|------|
| `/v1/databases/{id}` | `/v1/data_sources/{id}` | 데이터 소스 조회 |
| `/v1/databases/{id}/query` | `/v1/data_sources/{id}/query` | 데이터 쿼리 |

## ⚠️ 토큰 유효성 검증

### 테스트 명령
```bash
# .env 파일에서 토큰 로드
source .env

# Notion API 테스트
curl -X GET https://api.notion.com/v1/users/me \
  -H "Authorization: Bearer $NOTION_TOKEN" \
  -H "Notion-Version: 2025-09-03"
```

### 예상 응답

**성공 (200)**:
```json
{
  "object": "user",
  "id": "xxx-xxx-xxx",
  "type": "person",
  "person": { ... }
}
```

**실패 (401)**:
```json
{
  "object": "error",
  "status": 401,
  "code": "unauthorized",
  "message": "API token is invalid."
}
```

## 🛠️ 문제 해결

### 401 Unauthorized 에러

**원인 1: 토큰 자체가 유효하지 않음**
- Notion 개발자 포털에서 토큰 재확인
- 토큰 재생성 필요 여부 확인
- 토큰에 공백이 포함되었는지 확인

**원인 2: Integration 권한 부족**
- Notion 워크스페이스에서 Integration 활성화 확인
- 페이지/데이터베이스에 Integration 연결 확인
- "Add connections" 메뉴에서 통합 추가

**원인 3: API 버전 호환성**
- Notion-Version 헤더 확인
- Node.js SDK v5+ 사용 시 2025-09-03 이상 권장

### 환경 변수 미로드

```bash
# .env 파일에서 변수 로드
source .env

# 확인
echo $NOTION_TOKEN
```

### MCP 서버 연결 실패

```bash
# MCP 서버 재시작
# Claude Code 재시작 필요

# .claude/mcp.json JSON 문법 확인
jq '.' .claude/mcp.json
```

## 📖 사용 예제

### Claude Code에서 Notion 페이지 조회

```
Claude: "Notion에서 최근 생성된 페이지 5개를 보여주세요"

→ MCP가 NOTION_TOKEN을 사용하여 API 호출
→ Bearer 토큰 자동 처리
→ 페이지 목록 반환
```

### Node.js SDK 사용

```javascript
const { Client } = require("@notionhq/client")

const notion = new Client({
  auth: process.env.NOTION_TOKEN  // ntn_로 시작하는 토큰
})

// 페이지 조회
const page = await notion.pages.retrieve({
  page_id: "page-id-here"
})
```

## 🔒 보안 모범 사례

### ❌ 하지 말 것
```bash
# Git에 토큰 커밋
git add .env  # 금지!

# 로그에 토큰 출력
console.log("Token:", NOTION_TOKEN)  # 금지!

# 토큰 형식 검증 (regex 사용)
if (token.startsWith('ntn_')) { ... }  # 금지!
```

### ✅ 올바른 방법
```bash
# .env 파일은 .gitignore에 포함
# 자동으로 무시됨

# 환경 변수로 전달
export NOTION_TOKEN="..."

# 토큰을 opaque string으로 취급
// 그냥 그대로 사용하면 됨
```

## 📋 체크리스트

프로젝트 설정 확인:

- [x] `.env` 파일 생성
- [x] `NOTION_TOKEN` 설정
- [x] `.claude/mcp.json` 구성
- [x] Notion MCP 활성화
- [x] `.gitignore`에 `.env*` 포함
- [ ] Notion API 토큰 유효성 확인 (수동)
- [ ] Notion Integration 권한 확인 (수동)
- [ ] 페이지/데이터베이스 연결 확인 (수동)

## 📞 다음 단계

1. **Notion Integration 권한 확인**
   - https://www.notion.so/my-integrations 방문
   - 생성한 Integration 권한 확인
   - 필요한 페이지/데이터베이스에 연결

2. **토큰 유효성 검증**
   - 위의 테스트 명령 실행
   - 성공 응답 확인

3. **MCP 연결 테스트**
   - Claude Code에서 Notion 쿼리 시도
   - 데이터 반환 확인

## 📚 참고 자료

- **Notion API 공식 문서**: https://developers.notion.com/docs
- **업그레이드 가이드**: https://developers.notion.com/docs/upgrade-guide-2025-09-03
- **MCP 프로토콜**: https://spec.modelprop.com/

---

**작성일**: 2025년 1월
**API 버전**: 2025-09-03
**토큰 형식**: ntn_ (최신)
