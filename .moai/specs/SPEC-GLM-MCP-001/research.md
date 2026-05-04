# SPEC-GLM-MCP-001 Research Artifact (Phase 0.5)

## HISTORY

| Version | Date | Author | Description |
|---------|------|--------|-------------|
| 0.1.0 | 2026-05-01 | manager-spec | 초기 작성 — Z.AI MCP 서버 통합 deep research, 3개 통합 옵션 분석, integration point 매핑, 위험 카탈로그 |

## 연구 목표

`moai glm` 모드에서 사용자가 Claude Code 내부에서 Z.AI 의 공식 MCP 서버 3종(Vision, Web Search, Web Reader) 에 접근할 수 있도록 통합하는 방안을 연구한다. 본 연구는 코드 수정을 수반하지 않는 plan-stage 산출물이며, manager-cycle 위임 시점의 의사결정 근거를 제공한다.

연구 범위:
- Z.AI 공식 MCP 패키지 (`@z_ai/mcp-server`) 의 능력, 가격, 환경 요구사항 식별
- Claude Code 의 공식 MCP 등록 메커니즘 검토
- moai-adk-go 의 기존 GLM 통합 지점 (`internal/cli/glm.go`, settings 템플릿) 매핑
- 3가지 통합 옵션의 장단점 비교 + 권장안 도출
- 핵심 위험 카탈로그 (R1 ~ R5)

## 1. Z.AI 공식 MCP 서버 사실 카탈로그

WebSearch (2026-04-26) 로 확인된 사실. 본 SPEC 작성 시점에서는 추가 WebFetch 호출 없이 이 카탈로그를 인용한다.

### 1.1 패키지 정보

- npm 패키지: `@z_ai/mcp-server`
- 최소 버전: `>= v0.1.2` (GLM-4.6V Vision 능력 포함)
- 실행 방식: `npx -y @z_ai/mcp-server@latest`
- Node.js 전제: `>= v22.0.0`

### 1.2 노출 도구 (3종)

| MCP 서버 | 능력 | 사용 사례 |
|----------|------|-----------|
| Vision | GLM-4.6V 이미지 분석, OCR, UI 분석, 다이어그램 이해 | 스크린샷 분석, 문서 OCR, 디자인 검토 |
| Web Search | 최신 기술 문서, API 변경, OSS 업데이트 검색 | 최근 라이브러리 변경 확인, 보안 권고 검색 |
| Web Reader | 웹페이지 본문 + 구조화된 메타데이터 추출 | 블로그/문서/공지 본문 가져오기 |

### 1.3 가격 / 플랜 등급

- **Lite ($3/월)**: GLM 코어만 (Vision, Web Search 미포함)
- **Pro ($9/월) 이상**: Vision + Web Search 활성화

### 1.4 리전 모드

| 환경변수 | 값 | 엔드포인트 |
|---------|-----|------------|
| `Z_AI_MODE` | `ZAI` | 국제 (https://api.z.ai) |
| `Z_AI_MODE` | `ZHIPU` | Mainland China (https://open.bigmodel.cn/api/anthropic) |

### 1.5 공식 등록 명령

Claude Code 네이티브 CLI 방식:

```bash
claude mcp add -s user zai-mcp-server \
  --env Z_AI_API_KEY=<key> Z_AI_MODE=ZAI \
  -- npx -y "@z_ai/mcp-server@latest"
```

수동 `~/.claude.json` 의 `mcpServers` 엔트리 추가도 가능. `command: "npx"`, `args: ["-y", "@z_ai/mcp-server"]`, `env: {Z_AI_API_KEY: ..., Z_AI_MODE: ...}` 구조.

### 1.6 출처 검증

- 공식 문서: `https://docs.z.ai/devpack/mcp/vision-mcp-server` (필요 시 WebFetch 1회 한도)
- 본 연구는 사전 로드된 사실(WebSearch 2026-04-26) 인용. 추가 외부 호출 없이 SPEC 작성 가능.

## 2. moai-adk-go 기존 통합 지점

본 SPEC 의 변경은 통합 지점을 *참조*하지만 *수정하지 않는다* (사용자 명시 제약). 통합 지점 식별은 manager-cycle 위임 시 정확한 라인 범위를 활용하기 위함이다.

### 2.1 GLM 환경 주입 코드

| 파일 | 라인 (대략) | 역할 |
|-----|------------|------|
| `internal/cli/glm.go` | 167 (`setGLMEnv`) | 프로세스 env 주입 (ANTHROPIC_BASE_URL, AUTH_TOKEN, DISABLE_BETAS, DISABLE_PROMPT_CACHING) |
| `internal/cli/glm.go` | 372 / 548 / 827 / 872 | settings.json env writer (cc / glm / cg 모드 전환별) |
| `internal/cli/launcher.go` | 241 | `moai cc` 전환 시 DISABLE_PROMPT_CACHING 제거 |
| `internal/hook/session_start.go` | 245 | SessionStart 자동 감지 |
| `internal/template/templates/.claude/settings.json.tmpl` | 전역 | settings 템플릿 (env 블록 포함) |

### 2.2 GLM 토큰 저장소

- `~/.moai/.env.glm` 의 `GLM_AUTH_TOKEN` 키
- `loadGLMKey()` 함수가 파일을 읽어 `setGLMEnv()` 와 `injectTmuxSessionEnv()` 가 사용
- `moai github auth glm <token>` 패턴이 이미 존재 — 본 SPEC 의 옵션 2(opt-in subcommand) 와 일관된 UX 모델

### 2.3 SPEC-GLM-001 과의 관계

- SPEC-GLM-001 (`completed`) 은 호환성 자동화: `CLAUDE_CODE_DISABLE_EXPERIMENTAL_BETAS=1` + `DISABLE_PROMPT_CACHING=1` 자동 주입/제거
- 본 SPEC 은 SPEC-GLM-001 위에 *MCP 서버 등록* 기능을 추가하며, 캐싱/베타 헤더 정책을 변경하지 않는다
- 이를 통해 Task B(in-flight cache_control 검토) 와 충돌하지 않음을 보장

## 3. 통합 옵션 비교

### 3.1 옵션 1 — Auto-inject

`moai glm setup` 또는 `moai glm` 자체가 `~/.claude.json` 또는 settings.json 에 `mcpServers.zai-mcp-server` 를 자동 추가한다.

**장점:**
- 사용자 마찰 0 — `moai glm` 한 번만 실행
- 신규 사용자 온보딩 시간 단축

**단점:**
- 침습적 — 사용자의 기존 `~/.claude.json` MCP 설정과 충돌 가능
- 사용자가 Vision/WebSearch 를 원하지 않는 경우에도 강제 주입
- 옵트아웃 메커니즘 별도 필요
- 명시적 동의 부재 — settings.json 의 신뢰 모델을 위반할 수 있음

### 3.2 옵션 2 — Opt-in Subcommand (권장)

`moai glm tools enable <vision|websearch|webreader|all>` 신규 subcommand 가 사용자 요청 시 동일 설정을 작성. 대응 `moai glm tools disable <...>` 로 제거.

**장점:**
- 명시적 동의 — 사용자가 직접 활성화 요청
- idempotent — 반복 호출 안전
- reversible — disable 명령으로 깔끔하게 제거 가능
- 기존 패턴 (`moai github auth glm <token>`) 과 일관됨
- Lite/Pro 플랜 검증 + 친절한 에러 메시지를 한 곳에서 수행 가능
- 사용자의 기존 MCP 설정과 충돌 시 명확한 시각적 경고 가능

**단점:**
- 추가 단계 1회 필요 (한 번만 실행하면 영구)
- subcommand parser/route 확장 코드 필요

### 3.3 옵션 3 — Documentation Guide

README + CLAUDE.md 에 가이드 섹션을 추가하고 코드 변경은 하지 않음. 사용자가 직접 `claude mcp add ...` 실행.

**장점:**
- 코드 표면 0 — 충돌 위험 없음
- 가장 보수적

**단점:**
- 사용자 마찰 큼 — 문서 읽고 명령어 복사/붙여넣기
- GLM 토큰 재입력 부담 (`~/.moai/.env.glm` 활용 불가)
- 통합 추적 불가 (몇 명이 실제로 사용하는지 측정 불가)
- moai-adk-go 의 다른 통합 (`moai github auth glm`) 과 일관되지 않음

### 3.4 권장: 옵션 2

근거:
1. 기존 moai-adk-go 패턴 (`moai github auth glm`, `moai cc/glm/cg`) 과 동일한 명시적 사용자 의도 모델
2. 안전성 — 사용자가 자신의 MCP 설정을 통제
3. 가역성 — 활성화 시 disable 도 함께 제공하면 사용자 신뢰도 ↑
4. 토큰 재사용 — 기존 `~/.moai/.env.glm` `GLM_AUTH_TOKEN` 을 자동 활용하여 중복 입력 회피
5. 플랜 검증 — Pro 플랜 미보유 시 Vision/WebSearch 활성화 거부 + 명확한 안내

## 4. 위험 카탈로그

| ID | 위험 | 영향도 | 완화 전략 |
|----|------|--------|-----------|
| R1 | 사용자 `~/.claude.json` 에 이미 `zai-mcp-server` 엔트리 존재 — 덮어쓰기 vs 보존 | High | `tools enable` 시 기존 엔트리 감지 → AskUserQuestion(orchestrator) 또는 명시적 `--force` 플래그 요구. 옵션 2 의 명시적 UX 가 이 결정을 자연스럽게 해결 |
| R2 | 사용자가 GLM Coding Plan Lite ($3) 보유 — Vision MCP 호출 시 401/403 | Medium | enable 명령에서 사전 플랜 검증(가능한 경우 API ping) 또는 enable 시 "Pro 플랜 필요" 명시. 런타임 에러는 Z.AI 가 반환하므로 본 SPEC 은 *명확한 에러 메시지 가이드*만 제공 |
| R3 | Node.js < 22 — `npx @z_ai/mcp-server` 실패 | High | enable 명령 진입 시 `node --version` 체크. 부재/구버전 시 graceful fail + 설치 가이드 출력 |
| R4 | `~/.claude.json` (user scope) vs `.mcp.json` (project scope) 결정 | Medium | Z.AI 공식 가이드 따라 `-s user` (사용자 전역) 권장. 본 SPEC 의 enable 명령은 user scope 기본, `--scope project` 플래그로 옵트인 가능 |
| R5 | `moai cc` ↔ `moai glm` 모드 전환 시 MCP 설정 영향 | Medium | MCP 설정은 모드 독립적 — 한 번 enable 하면 cc 모드에서도 도구가 노출됨. 단, Z.AI API key 가 필요하므로 cc 모드에서는 호출 시 401 반환 가능. 본 SPEC 은 모드별 자동 enable/disable 을 *하지 않는다* (사용자가 명시적으로 disable 해야 함) — 이것이 가장 단순하고 예측 가능한 동작 |

## 5. 16개 언어 중립성

본 통합은 moai-adk-go 의 *내부 CLI 기능* 이며 `internal/template/templates/` 의 사용자 프로젝트 템플릿을 변경하지 않는다. 따라서 16개 언어 중립성 정책 (CLAUDE.local.md §15) 의 직접적 검증 대상이 아니다.

다만 enable 명령의 도움말/에러 메시지가 *한국어 사용자만* 또는 *영어 사용자만* 우대하지 않도록 i18n 컨벤션 (오류 영어, 사용자 메시지 conversation_language) 을 따른다.

## 6. 캐싱 / cache_control 정책 영향

SPEC-GLM-001 의 `DISABLE_PROMPT_CACHING=1` 정책은 본 SPEC 변경의 영향을 받지 않는다.

근거:
- MCP 서버는 *Claude Code 의 도구 호출 경로* 를 통해 작동
- prompt caching 은 *Anthropic API 의 인풋 토큰 캐싱* 을 제어
- MCP 서버 등록(`mcpServers` JSON 추가) 은 환경변수 또는 헤더에 영향을 주지 않음
- 따라서 Task B (in-flight cache_control 검토) 와 직교적

## 7. Open Questions (Plan-Stage Resolution Required)

| OQ | 질문 | 제안 답변 |
|----|------|----------|
| OQ-1 | Vision/WebSearch/WebReader 를 별도 enable 가능 vs `all` 만 지원? | 별도 + `all` 둘 다 지원. 사용자가 Vision 만 원하는 경우(Pro 플랜 가입은 했으나 WebSearch 데이터 사용 우려) 를 배려 |
| OQ-2 | enable 결과를 `~/.claude.json` (user scope) vs `.mcp.json` (project scope) 어디에 기록? | 기본 user scope (`~/.claude.json`), `--scope project` 옵션 제공 |
| OQ-3 | 기존 zai-mcp-server 엔트리 감지 시 동작? | (a) 사용자 토큰 일치 시 idempotent skip, (b) 토큰 다른 경우 에러 + `--force` 안내 |
| OQ-4 | Lite 플랜 사용자가 vision/websearch enable 시 동작? | 경고 출력 + enable 진행 (사전 검증 API 가 없거나 비싸면). 런타임 401 시 Z.AI 응답 그대로 노출 |
| OQ-5 | Mainland China (`Z_AI_MODE=ZHIPU`) 를 v0.1 에서 지원? | **제외 (v0.2 로 연기)**. v0.1 은 `Z_AI_MODE=ZAI` 국제 엔드포인트만 지원 |
| OQ-6 | enable 시 sample test (Vision OCR ping 등) 자동 실행? | 옵트인 (`--test`) — 기본 비활성. test 가 npm 다운로드를 트리거하므로 첫 실행 ~10s 지연 발생 가능 |

OQ-5 는 spec.md 의 Exclusions (EX-1) 으로 명시 분리. 그 외 OQ 는 plan.md 의 결정 표로 추적.

## 8. References

- Z.AI 공식 MCP 문서: https://docs.z.ai/devpack/mcp/vision-mcp-server (필요 시 WebFetch 1회 한도)
- Claude Code MCP 가이드: `claude mcp add --help`
- `.moai/specs/SPEC-GLM-001/spec.md` — 베이스 GLM 호환성
- `.moai/specs/SPEC-CC2122-MCP-001/spec.md` — `.mcp.json.tmpl` 변경 패턴 (구조 참조)
- `.moai/specs/SPEC-LSPMCP-001/spec.md` — moai-lsp MCP 통합 패턴 (구조 참조)
- `internal/cli/glm.go` — 환경 주입 진입점 (직접 수정하지 않음)
- `internal/cli/github_auth.go` — `moai github auth glm <token>` 패턴 참조
- CLAUDE.local.md §13 (GLM Integration Testing), §15 (Template Language Neutrality)

## 9. 결론

옵션 2 (opt-in `moai glm tools enable/disable`) 가 이 프로젝트의 기존 패턴, 안전성, 가역성, 사용자 신뢰 모델 모두에서 우월하다. 본 결정은 spec.md 의 REQ 분배에 반영되며, plan.md 의 마일스톤 M1 ~ M4 가 해당 옵션을 구현 가이드로 사용한다.
