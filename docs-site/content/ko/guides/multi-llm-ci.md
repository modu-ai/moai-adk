---
title: "Multi-LLM CI 가이드"
description: "GitHub Actions에서 여러 AI 모델로 코드 리뷰 자동화"
date: 2026-04-27
draft: false
weight: 10
---

# Multi-LLM CI 가이드

MoAI-ADK의 Multi-LLM CI 기능을 사용하여 GitHub Actions에서 다양한 LLM으로 코드 리뷰를 설정하는 방법을 안내합니다.

## 개요

### Multi-LLM CI란?

MoAI-ADK의 Multi-LLM CI 기능은 GitHub Actions에서 여러 AI 모델로 동시에 코드 리뷰를 수행하는 통합 CI/CD 파이프라인을 제공합니다.

### 지원하는 LLM

| LLM | 제공자 | 트리거 방식 | 특징 |
|-----|--------|-------------|------|
| **Claude** | Anthropic | `/claude` 코멘트 | Issue/PR 리뷰, OAuth 인증 |
| **Codex** | OpenAI | PR open 자동 | ⚠️ 비공개 레포 전용 |
| **Gemini** | Google | PR open 자동 | API Key 인증 |
| **GLM** | Zhipu AI | PR open 자동 | 토큰 인증 |

## 시작하기

### 사전 요구사항

- macOS (arm64) - v1.0 기준
- Go 1.23+
- GitHub repository
- 각 LLM 계정 및 API 토큰

### 초기 설정

```bash
moai github init
```

이 명령이 수행하는 작업:
- `.github/workflows/` 디렉토리 생성
- workflow 템플릿 배포
- composite actions 배포
- GitHub Secrets 설정 가이드

### LLM 인증 설정

```bash
# Claude (OAuth)
moai github auth claude

# Codex (비공개 레포)
moai github auth codex

# Gemini
moai github auth gemini

# GLM
moai github auth glm
```

### GitHub Secrets 설정

각 LLM별 필요한 Secrets:
- `CLAUDE_CODE_OAUTH_TOKEN` - Claude OAuth 토큰
- `CODEX_AUTH_JSON` - Codex 인증 JSON (base64 인코딩)
- `GEMINI_API_KEY` - Gemini API Key
- `GLM_API_KEY` - GLM API Token

### 첫 번째 PR 테스트

PR을 생성하면 자동으로 LLM Panel 코멘트가 추가됩니다:

```markdown
## LLM Code Review Status

| LLM | Status |
|-----|--------|
| Claude | Pending (add `/claude` comment) |
| Codex | ✓ Ready |
| Gemini | ⚠️ Token missing |
| GLM | ✓ Ready |

Trigger individual reviews:
- Add `/claude` comment to trigger Claude
- Add `/codex` comment to trigger Codex
- Add `/gemini` comment to trigger Gemini
- Add `/glm` comment to trigger GLM
```

## LLM 인증 설정

### Claude 설정

#### OAuth 토큰 발급

1. [Claude Code](https://claude.ai/download) 설치
2. 로그인 후 OAuth 토큰 발급
3. `.claude/settings.local.json`에 자동 저장

#### moai github auth claude

```bash
moai github auth claude
```

**대화형 설정 프로세스:**
```
Claude OAuth 토큰을 찾을 수 없습니다.
Claude Code를 설치하고 로그인해 주시겠습니까? (y/n): y

[확인됨] OAuth 토큰이 settings.local.json에 저장되었습니다.
GitHub Secret: CLAUDE_CODE_OAUTH_TOKEN에 다음 값을 설정하세요:
<token-value>
```

### Codex 설정 (비공개 레포 전용)

#### 인증 JSON 생성

```json
{
  "token": "sk-...",
  "base_url": "https://api.openai.com/v1"
}
```

#### moai github auth codex

```bash
moai github auth codex
```

**대화형 설정:**
```
OpenAI auth.json 파일 경로: ~/.codex/auth.json
파일을 읽어 GitHub Secret을 생성합니다...
⚠️ Codex는 비공개 레포에서만 사용 가능합니다 (REQ-SEC-001)

생성된 Secret:
CODEX_AUTH_JSON=eyJ0...
```

### Gemini 설정

```bash
moai github auth gemini
```

API Key 입력 후 자동으로 GitHub Secret 설정 가이드 제공.

### GLM 설정

```bash
moai github auth glm
```

GLM 토큰 경로 (`~/.moai/.env.glm`)에서 자동 읽기.

## Workflow 템플릿 이해

### llm-panel.yml

**트리거:** PR opened

**역할:** 각 LLM의 상태를 시각적으로 표시하는 패널 코멘트 자동 생성

**비고:** `/claude`, `/codex`, `/gemini`, `/glm` 코멘트로 개별 리뷰 트리거

### claude.yml / claude-code-review.yml

- **claude.yml**: Issue 트리거 (초안 리뷰)
- **claude-code-review.yml**: PR 트리거 (변경사항 리뷰)

**특징:** `/claude` 코멘트로만 트리거

### codex-review.yml

**보안 제약:**
- `private` 레포에서만 동작 (REQ-SEC-001)
- `visibility` 체크로 공개 레포 차단

**workflow:**
```yaml
private-guard:
  runs-on: ubuntu-latest
  steps:
    - name: Check Repository Visibility
      run: |
        if [[ "${{ github.repository_visibility }}" == "public" ]]; then
          echo "::error::Codex review is restricted to private repositories"
          exit 1
        fi
```

### gemini-review.yml

- 자동 언어 감지 (detect-language action)
- PR synchronized 시 자동 트리거

### glm-review.yml

- GLM 전용 환경 설정 (setup-glm-env action)
- 환경 변수 자동 주입

### Composite Actions

#### detect-language

**입력:** repository 루트 경로
**출력:** language 환경 변수 (`detected_language`)

**지원 언어:** Go, Python, TypeScript, JavaScript, Rust, Java, Kotlin, C#, Ruby, PHP, Elixir, C++, Scala, R, Flutter, Swift (16개)

#### setup-glm-env

GLM 팀 모드에서 필요한 환경 변수 설정:
- `ANTHROPIC_AUTH_TOKEN` (GLM endpoint)
- `ANTHROPIC_BASE_URL` (https://glm.modu-ai.kr)

## 고급 설정

### github-actions.yaml 커스터마이징

#### 기본 구조

```yaml
# .moai/config/sections/github-actions.yaml
llm_review:
  enabled: true
  runners:
    claude: true
    codex: true
    gemini: true
    glm: true
  triggers:
    on_pr_open: true
    on_comment:
      claude: "/claude"
      codex: "/codex"
      gemini: "/gemini"
      glm: "/glm"
```

#### 언어별 LLM 할당

```yaml
language_rules:
  go:
    - gemini
    - claude
  python:
    - claude
    - glm
  typescript:
    - codex
    - claude
```

### Runner 버전 관리

#### 자동 업데이트 확인

```bash
moai github status
```

**출력 예시:**
```
✓ GitHub Actions Runner
  Version: 2.700.1 (10 days old)
  Status: OK

⚠️ Update available: 2.701.0
Run: moai doctor --fix
```

#### Doctor 통합

```bash
moai doctor
```

runner 버전 체크가 시스템 진단에 통합됩니다.

## 트러블슈팅

### PR 코멘트 트리거가 동작하지 않을 때

#### Checklist

1. ✅ GitHub Actions workflow가 활성화되어 있는가?
   - Repository → Actions → workflows 확인

2. ✅ GitHub Secrets가 설정되어 있는가?
   - Settings → Secrets and variables → Actions

3. ✅ Workflow permissions이 올바른가?
   - `contents: read`, `pull-requests: write` 필요

### LLM별 에러 대응

#### Claude

**Error:** `CLAUDE_CODE_OAUTH_TOKEN expired`
**해결:** `moai github auth claude` 재실행

#### Codex

**Error:** `repository visibility check failed`
**원인:** 공개 레포에서 Codex 사용 시도
**해결:** 레포를 비공개로 전환

#### Gemini

**Error:** `GEMINI_API_KEY quota exceeded`
**해결:** Google Cloud Console에서 quota 증설

#### GLM

**Error:** `GLM_API_KEY authentication failed`
**해결:** `~/.moai/.env.glm` 토큰 확인

## 다음 단계

- [CLI 레퍼런스 참조](/docs/commands/)
- [Workflow 설정 참조](/docs/configuration/)
- [보안 정책 확인](/docs/security/)
