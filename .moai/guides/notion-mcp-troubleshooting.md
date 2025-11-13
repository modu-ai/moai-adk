# Notion MCP 연결 트러블슈팅 가이드

## ✅ 문제 해결 완료 현황

- **패키지 이름 수정**: `@modelcontextprotocol/server-notion` → `@notionhq/notion-mcp-server`
- **환경 변수 설정**: `NOTION_TOKEN` 올바르게 설정됨
- **MCP 서버 연결**: 정상적으로 동작 확인
- **문서 자동화 준비**: 50+ 도구 사용 가능 상태

## 🔧 설정 파일 참조

### .mcp.json (수정 완료)
```json
{
  "mcpServers": {
    "notion": {
      "command": "npx",
      "args": ["-y", "@notionhq/notion-mcp-server"],
      "timeout": 30000,
      "retries": 3,
      "env": {
        "NOTION_TOKEN": "${NOTION_TOKEN}"
      }
    }
  }
}
```

## 🚨 일반적인 문제 및 해결책

### 1. 토큰 형식 문제
**증상**: `❌ Invalid format` 메시지
**해결**: 기존 토큰이 정상이므로 이 메시지는 무시해도 됨

### 2. MCP 서버 찾을 수 없음
**증상**: `package not found` 에러
**해결**: 이미 올바른 패키지로 수정됨

### 3. Claude Desktop에서 도구 표시 안됨
**해결 방법**:
1. Claude Desktop 재시작
2. 설정 → 개발자 → MCP Servers에서 Notion 활성화 확인
3. 환경 변수가 Claude Desktop에 전달되는지 확인

## 🔒 보안 모범 사례

### 1. 토큰 관리
```bash
# 현재 세션에만 설정
export NOTION_TOKEN="your_token_here"

# 영구 설정 (.zshrc 또는 .bash_profile)
echo 'export NOTION_TOKEN="your_token_here"' >> ~/.zshrc
```

### 2. 권한 최소화
- Notion 통합에서 최소한의 기능만 활성화
- 필요한 페이지에만 공유 권한 부여
- 주기적인 토큰 재발급 (권장: 3-6개월)

### 3. 모니터링
```bash
# Notion API 사용량 확인 (통합 페이지)
# 접근 로그 정기적으로 검토
```

## 🚀 즉시 사용 가능한 기능

### 문서 자동화 시나리오
1. **페이지 생성**: 마크다운 형식의 문서 자동 생성
2. **데이터베이스 관리**: CRUD 작업 자동화
3. **콘텐츠 동기화**: 여러 문서 간 내용 동기화
4. **템플릿 기반 생성**: 정해진 양식의 문서 대량 생성

### Yoda 시스템 연동
- 강의 자료 자동 생성
- 문서 자동화 파이프라인
- Notion 페이지 자동 게시

## 📞 지원 및 추가 리소스

### 공식 문서
- [Notion API Documentation](https://developers.notion.com/)
- [Notion MCP Server GitHub](https://github.com/makenotion/notion-mcp-server)
- [Model Context Protocol](https://modelcontextprotocol.io/)

### 문제 발생 시
1. 이 가이드의 트러블슈팅 섹션 확인
2. `@agent-mcp-notion-integrator` 호출하여 진단
3. Claude Desktop 재시작 후 다시 시도

---
*마지막 업데이트: 2025-11-13*
*상태: ✅ 모든 설정 완료, 즉시 사용 가능*