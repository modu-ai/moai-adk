# Claude Code 개발 환경 설정 가이드

> **For English Users**: This guide is in Korean. Project documentation in CLAUDE.md uses your configured conversation language.

이 가이드는 Claude Code v4.0+에서 MoAI-ADK 프로젝트를 최적으로 사용하기 위한 선택사항 설정을 안내합니다.

---

## 🔧 선택사항: .env 파일 접근 설정

기본적으로 `.claude/settings.json`은 보안을 우선하여 `.env` 파일 접근을 제한합니다. 하지만 로컬 개발 환경에서 `.env` 파일이 필요한 경우, 다음과 같이 설정할 수 있습니다.

### 문제 상황

다음과 같은 경우 `.env` 접근이 필요합니다:

- 🔑 API 키 관리 (외부 서비스 통합)
- 🗄️ 데이터베이스 연결 문자열
- 🔐 인증 토큰 설정
- ⚙️ 개발 환경 변수 관리
- 🧪 테스트 환경 설정

### 해결 방법

프로젝트 루트에 `.claude/settings.local.json` 파일을 생성하여 로컬 개발 환경에서만 `.env` 접근을 허용하세요.

#### Step 1: 파일 생성

```bash
mkdir -p .claude
touch .claude/settings.local.json
```

#### Step 2: 설정 추가

`.claude/settings.local.json`에 다음 내용을 추가하세요:

```json
{
  "permissions": {
    "allow": [
      "Read(./.env)",
      "Read(./.env.*)",
      "Write(./.env)",
      "Edit(./.env)",
      "Read(./.env.local)",
      "Write(./.env.local)"
    ]
  }
}
```

#### Step 3: Claude Code 재시작

Claude Code를 완전히 재시작하여 새 설정을 적용하세요.

### 설정 완료 확인

설정이 올바르게 적용되었는지 확인하세요:

```bash
# .env 파일 읽기 테스트
cat .env  # 또는 Read tool 사용

# .env.local 파일 편집 테스트 (로컬 개발 전용)
echo "TEST_KEY=value" > .env.local
```

---

## ⚠️ 보안 주의사항

### 로컬 개발에서만 사용하세요

**절대**: `.claude/settings.local.json`를 버전 관리(Git)에 커밋하지 마세요.

```bash
# .gitignore에 다음 추가
echo ".claude/settings.local.json" >> .gitignore
```

### 민감한 정보 보호

- 🔐 프로덕션 환경 변수는 **절대** `.env`에 저장하지 마세요
- 🔑 개인 API 키나 토큰은 로컬에서만 유지하세요
- 📝 실수로 커밋한 경우 **즉시** 토큰을 재발급하세요

### Git 실수 시 조치

실수로 `.env`나 민감 정보를 커밋한 경우:

```bash
# 1. 히스토리에서 제거 (강제)
git filter-branch --tree-filter 'rm -f .env' HEAD && git push --force

# 2. 토큰 즉시 재발급
# → AWS, GitHub, API 서비스 대시보드에서 재발급

# 3. 감사 로그 확인
# → GitHub, AWS CloudTrail 등에서 비정상 접근 확인
```

---

## 📚 추가 설정 옵션

### 1. 다른 민감 파일도 제한하기

필요에 따라 다른 파일도 제한할 수 있습니다:

```json
{
  "permissions": {
    "deny": [
      ".aws/credentials",
      ".ssh/*",
      ".env.production",
      "config/secrets.json",
      "private-keys/*"
    ]
  }
}
```

### 2. 특정 디렉토리만 허용하기

```json
{
  "permissions": {
    "allow": [
      "Read(./src/**)",
      "Edit(./tests/**)",
      "Read(./.env)"
    ],
    "deny": [
      "Write(./src/core/**)",
      "Edit(./deployment/**)"
    ]
  }
}
```

### 3. MCP 서버 설정

외부 서비스와 통합할 때:

```json
{
  "mcpServers": {
    "github": {
      "command": "npx",
      "args": ["-y", "@anthropic-ai/mcp-server-github"]
    }
  }
}
```

---

## 🔄 다양한 환경별 설정

### 개발 환경 (.env.local)

```bash
# .env.local (로컬 개발 전용, .gitignore에 추가)
DEBUG=true
LOG_LEVEL=debug
DATABASE_URL=postgresql://localhost/dev
API_KEY=local-test-key-123
```

### 테스트 환경 (.env.test)

```bash
# .env.test (테스트 전용)
DEBUG=false
LOG_LEVEL=error
DATABASE_URL=postgresql://localhost/test
API_KEY=test-key-456
```

### 프로덕션 환경

```bash
# CI/CD 또는 배포 플랫폼에서 설정
# (절대 .env 파일에 저장하지 마세요)
# → GitHub Secrets, AWS Systems Manager, Vercel Secrets 등
```

---

## 🚀 Best Practices

### ✅ 권장 사항

- ✅ **로컬 개발에서만** `.claude/settings.local.json` 사용
- ✅ 프로덕션 환경은 **플랫폼 제공 시크릿 관리** 사용
- ✅ `.env.example` 파일로 필수 환경 변수 문서화
- ✅ 자동으로 `.env` 및 설정 파일을 `.gitignore`에 추가
- ✅ 정기적으로 토큰 및 키 **로테이션**

### ❌ 피해야 할 사항

- ❌ `.env` 파일을 버전 관리에 커밋하기
- ❌ 프로덕션 키를 로컬 `.env`에 저장하기
- ❌ `.claude/settings.local.json`을 Git에 추적하기
- ❌ 민감 정보를 로그나 에러 메시지에 출력하기
- ❌ 개발용 키를 프로덕션에서 사용하기

---

## 📖 추가 리소스

- **CLAUDE.md**: 전체 MoAI-ADK 설정 및 워크플로우 가이드
- **Security & Best Practices**: CLAUDE.md의 보안 섹션 참고
- **Claude Code 문서**: https://code.claude.com/docs

---

## 도움말

설정 중 문제가 발생한 경우:

1. **로그 확인**: `.moai/logs/` 디렉토리 확인
2. **설정 검증**: `.claude/settings.local.json` JSON 형식 검증
3. **Claude Code 재시작**: 완전히 종료 후 재시작
4. **GitHub Issues**: https://github.com/anthropics/claude-code/issues

---

**Last Updated**: 2025-11-18
**Version**: 0.25.11
**Language**: Korean (한국어)
