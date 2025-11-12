# Notion MCP 설정 가이드

Master Yoda 시스템에서 Notion 자동 발행을 사용하려면 Notion API를 설정해야 합니다.

## 1단계: Notion 통합 생성

### 1.1 Notion 계정 설정

1. [Notion 개발자 페이지](https://www.notion.so/my-integrations)에 접속
2. "새 통합(New Integration)" 클릭
3. 이름: `MoAI-ADK Yoda` 입력
4. "제출" 클릭

### 1.2 API 키 복사

1. 생성된 통합에서 "API 키 표시(Show API key)" 클릭
2. 키를 복사 (나중에 필요함)

```
Internal Integration Token:
secret_XXXXXXXXXXXXXXXXXXXXXXXXXXXXX
```

## 2단계: 환경변수 설정

### 2.1 .env 파일 생성 (로컬 전용, git에서 제외됨)

```bash
# .env (절대 git에 커밋하지 마세요!)
NOTION_API_KEY=secret_XXXXXXXXXXXXXXXXXXXXXXXXXXXXX
NOTION_DATABASE_ID=your-database-id
```

**⚠️ 보안 경고**:
- 이 파일을 git에 커밋하지 마세요
- `.gitignore`에 `.env*`가 있는지 확인하세요
- API 키는 공개하지 마세요

### 2.2 또는 환경변수 직접 설정

```bash
# macOS/Linux
export NOTION_API_KEY="secret_XXXXXXXXXXXXXXXXXXXXXXXXXXXXX"
export NOTION_DATABASE_ID="your-database-id"

# Windows (PowerShell)
$env:NOTION_API_KEY="secret_XXXXXXXXXXXXXXXXXXXXXXXXXXXXX"
$env:NOTION_DATABASE_ID="your-database-id"
```

## 3단계: Notion 데이터베이스 설정

### 3.1 데이터베이스 생성

1. Notion 워크스페이스에서 새 페이지 생성
2. 우측 상단 "+" → "데이터베이스" → "테이블"
3. 이름: `MoAI-ADK Lectures` 입력

### 3.2 데이터베이스 속성 설정

다음 속성을 추가하세요:

| 속성명 | 타입 | 설명 |
|--------|------|------|
| **Title** | 제목 (기본) | 강의 제목 |
| **Topic** | 텍스트 | 강의 주제 |
| **Instructor** | 텍스트 | 강사명 |
| **Difficulty** | 선택 (easy/medium/hard) | 난이도 |
| **Date** | 날짜 | 강의 날짜 |
| **Format** | 선택 (education/presentation/workshop) | 강의 형식 |
| **MarkdownURL** | URL | 마크다운 파일 링크 |
| **PDFPath** | URL | PDF 파일 경로 |
| **Status** | 선택 (Draft/Published) | 발행 상태 |

### 3.3 데이터베이스 ID 복사

1. 데이터베이스 페이지 URL 복사
   ```
   https://notion.so/[DATABASE_ID]?v=[VIEW_ID]
   ```
2. `[DATABASE_ID]` 부분이 당신의 데이터베이스 ID입니다
3. 하이픈 제거: `DATABASE_ID = "abc123def456"` (하이픈 없음)

## 4단계: 통합에 데이터베이스 접근 권한 부여

### 4.1 데이터베이스에서 권한 설정

1. 데이터베이스 페이지 우측 상단 "공유(Share)"
2. "초대하기(Invite)" 클릭
3. 생성한 통합 `MoAI-ADK Yoda` 검색 후 추가
4. 권한: "편집자(Editor)" 선택
5. "공유(Share)" 클릭

## 5단계: .claude/mcp.json 업데이트

`.claude/mcp.json`의 Notion 설정을 업데이트하세요:

```json
{
  "mcpServers": {
    "notion": {
      "command": "mcp-server-notion",
      "enabled": true,
      "config": {
        "authToken": "secret_XXXXXXXXXXXXXXXXXXXXXXXXXXXXX",
        "database": {
          "name": "MoAI-ADK Lectures",
          "id": "your-actual-database-id-without-hyphens"
        },
        "pageProperties": {
          "title": "Title",
          "topic": "Topic",
          "instructor": "Instructor",
          "difficulty": "Difficulty",
          "date": "Date",
          "format": "Format",
          "status": "Status"
        }
      }
    }
  }
}
```

## 6단계: 테스트

### 6.1 Notion 연결 테스트

```bash
# Notion MCP가 정상 작동하는지 확인
/yoda:generate --topic "Notion Test" --output "notion" --format "education"
```

### 6.2 예상 결과

1. `.moai/yoda/output/notion-test.md` 생성
2. Notion 데이터베이스에 새 페이지 자동 생성
3. `.moai/yoda/output/notion-test-notion-link.txt` 에 공개 링크 저장

## 트러블슈팅

### 문제 1: "API 키가 유효하지 않음" 오류

**해결책**:
- `secret_` 로 시작하는지 확인
- 전체 키가 복사되었는지 확인
- 환경변수가 올바르게 설정되었는지 확인

```bash
# 확인 명령어
echo $NOTION_API_KEY
```

---

### 문제 2: "데이터베이스에 접근할 수 없음" 오류

**해결책**:
- 데이터베이스 ID가 올바른지 확인 (하이픈 없음)
- 통합이 데이터베이스 접근 권한을 가지고 있는지 확인
- Notion에서 "공유" → 통합 추가 확인

---

### 문제 3: "페이지 생성 실패" 오류

**해결책**:
- 통합 권한이 "편집자(Editor)"인지 확인
- 데이터베이스 속성이 올바르게 설정되었는지 확인
- Notion API 상태 페이지 확인: https://status.notion.so/

---

### 문제 4: 환경변수가 인식되지 않음

**해결책**:
- `.env` 파일을 사용할 경우, `python-dotenv` 설치 확인
- 환경변수를 설정한 후 터미널을 다시 실행
- 환경변수 확인:
  ```bash
  printenv | grep NOTION
  ```

---

## 보안 주의사항

⚠️ **절대 하지 마세요**:
- ❌ API 키를 git에 커밋하기
- ❌ API 키를 로그에 출력하기
- ❌ API 키를 GitHub에 공개하기
- ❌ 개발 API 키와 프로덕션 API 키 혼동하기

✅ **안전한 관행**:
- ✅ 환경변수로 관리하기
- ✅ `.env` 파일을 `.gitignore`에 추가하기
- ✅ 주기적으로 API 키 재설정하기
- ✅ 불필요한 권한 제거하기

---

## Notion 공개 페이지 설정 (옵션)

생성된 페이지를 공개하려면:

1. Notion 페이지 우측 상단 "공유(Share)"
2. "링크 복사(Copy link)" 또는 "공개(Share to web)"
3. "공개 링크 사용(Use public link)" 활성화
4. 링크가 자동으로 `.moai/yoda/output/{topic}-notion-link.txt`에 저장됨

---

## 다음 단계

Notion 설정이 완료되면:

```bash
# 전체 형식으로 강의 생성
/yoda:generate \
  --topic "FastAPI 기초" \
  --instructor "김철학" \
  --format "education" \
  --output "pdf,pptx,docx,notion"
```

모든 형식이 동시에 생성되고 Notion에 자동 발행됩니다! 🚀

---

*Notion MCP 설정 완료 후 Master Yoda의 완전한 기능을 사용할 수 있습니다.*
