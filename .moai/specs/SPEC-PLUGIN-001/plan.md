# SPEC-PLUGIN-001 구현 계획

> **플러그인 구조 설계 및 마이그레이션 전략**

---

## 🎯 구현 목표

1. Claude Code 플러그인 표준 구조 정의
2. `plugin.json` 매니페스트 설계
3. `hooks.json` 설정 변환
4. 기존 템플릿 마이그레이션 계획

---

## 📂 디렉토리 구조 설계

### 최종 구조

```
MoAI-ADK/
├── .claude-plugin/              # 플러그인 루트 (새로 추가)
│   └── plugin.json              # 플러그인 매니페스트
│
├── commands/                    # Alfred 커맨드 (기존 위치)
│   └── alfred/
│       ├── 1-spec.md
│       ├── 2-build.md
│       ├── 3-sync.md
│       └── ...
│
├── agents/                      # 전문 에이전트 (기존 위치)
│   └── alfred/
│       ├── spec-builder.md
│       ├── code-builder.md
│       └── ...
│
├── hooks/                       # 후크 설정 (새로 추가)
│   └── hooks.json               # PostToolUse, PreToolUse 등
│
├── templates/                   # 프로젝트 템플릿 (기존 유지)
│   └── .moai/
│       ├── project/
│       └── memory/
│
├── scripts/                     # 자동화 스크립트 (새로 추가)
│   ├── format-code.sh           # 코드 포맷팅
│   └── validate-spec.sh         # SPEC 검증
│
├── moai-adk-ts/                 # TypeScript 구현 (기존 유지)
│   ├── src/
│   └── tests/
│
└── .mcp.json                    # MCP 서버 설정 (선택)
```

### 기존 구조와의 차이점

| 항목 | 기존 | 신규 | 변경 내용 |
|------|------|------|-----------|
| 플러그인 루트 | `.claude/` | `.claude-plugin/` | 표준 디렉토리명 |
| 매니페스트 | `settings.json` | `plugin.json` | 표준 파일명 |
| 후크 설정 | `settings.json` 내부 | `hooks/hooks.json` | 분리된 파일 |
| 환경변수 | 없음 | `${CLAUDE_PLUGIN_ROOT}` | 경로 참조 표준화 |

---

## 🔧 plugin.json 매니페스트 설계

### 핵심 필드

```json
{
  "name": "moai-adk",
  "version": "0.3.0",
  "description": "MoAI Agentic Development Kit - SPEC-First TDD Framework with Alfred SuperAgent",
  "author": "MoAI (modu.ai)",
  "homepage": "https://github.com/modu-ai/moai-adk",
  "license": "MIT",

  "commands": "./commands/alfred",
  "agents": "./agents/alfred",
  "hooks": "./hooks/hooks.json",

  "mcpServers": {
    "moai-adk-server": {
      "command": "node",
      "args": ["${CLAUDE_PLUGIN_ROOT}/moai-adk-ts/dist/mcp-server.js"]
    }
  }
}
```

### 필드 설명

| 필드 | 타입 | 필수 | 설명 |
|------|------|------|------|
| `name` | string | ✅ | 플러그인 고유 ID (소문자, 하이픈) |
| `version` | string | ✅ | Semantic Version (0.3.0) |
| `description` | string | ⚠️ | 플러그인 설명 |
| `author` | string | ⚠️ | 작성자 |
| `homepage` | string | ⚠️ | 프로젝트 홈페이지 |
| `license` | string | ⚠️ | 라이선스 (MIT, Apache-2.0 등) |
| `commands` | string | ⚠️ | 커맨드 디렉토리 상대 경로 |
| `agents` | string | ⚠️ | 에이전트 디렉토리 상대 경로 |
| `hooks` | string | ⚠️ | hooks.json 파일 경로 |
| `mcpServers` | object | ⚠️ | MCP 서버 정의 (선택) |

---

## 🪝 hooks.json 설계

### 기존 settings.json 구조 (참고)

```json
{
  "hooks": {
    "PostToolUse": [
      {
        "matcher": "Write|Edit",
        "hooks": [
          {
            "type": "command",
            "command": "npm run format"
          }
        ]
      }
    ]
  }
}
```

### 신규 hooks/hooks.json 구조

```json
{
  "hooks": {
    "PostToolUse": [
      {
        "matcher": "Write|Edit",
        "hooks": [
          {
            "type": "command",
            "command": "${CLAUDE_PLUGIN_ROOT}/scripts/format-code.sh"
          }
        ]
      }
    ],
    "PreToolUse": [
      {
        "matcher": "Write",
        "hooks": [
          {
            "type": "command",
            "command": "${CLAUDE_PLUGIN_ROOT}/scripts/validate-spec.sh"
          }
        ]
      }
    ]
  }
}
```

### 후크 타입 설명

| 후크 타입 | 트리거 시점 | 사용 사례 |
|----------|-------------|----------|
| `PreToolUse` | 도구 실행 전 | 파일 검증, 백업, 사전 체크 |
| `PostToolUse` | 도구 실행 후 | 포맷팅, 린트, TAG 업데이트 |

---

## 🔄 마이그레이션 전략

### 1차 목표: 구조 변경

- ✅ `.claude-plugin/` 디렉토리 생성
- ✅ `plugin.json` 작성
- ✅ `hooks/hooks.json` 분리
- ✅ `scripts/` 디렉토리 추가

### 2차 목표: 호환성 유지

- ✅ 기존 `commands/`, `agents/` 디렉토리 위치 유지
- ✅ `templates/` 디렉토리 구조 보존
- ✅ `moai-adk-ts/` 구현 코드 변경 없음

### 3차 목표: 자동화 스크립트

- ✅ `scripts/format-code.sh` - Biome 기반 포맷팅
- ✅ `scripts/validate-spec.sh` - SPEC 메타데이터 검증
- ✅ `scripts/sync-tags.sh` - TAG 체인 검증 자동화

---

## 🛠️ 기술적 접근 방법

### 환경변수 활용

**문제점**:
- 절대 경로 하드코딩 시 플러그인 이식성 저하
- 사용자별 경로 차이로 인한 호환성 문제

**해결책**:
- `${CLAUDE_PLUGIN_ROOT}` 환경변수 사용
- 모든 내부 파일 참조에 이 변수 활용

**예시**:
```json
{
  "command": "${CLAUDE_PLUGIN_ROOT}/scripts/format-code.sh"
}
```

### 상대 경로 규칙

**규칙**:
- 외부 파일 참조: `./`로 시작
- 내부 파일 참조: `${CLAUDE_PLUGIN_ROOT}` 사용

**예시**:
```json
{
  "commands": "./commands/alfred",
  "agents": "./agents/alfred"
}
```

---

## 🎯 우선순위별 마일스톤

### 1차 마일스톤: 기본 구조 확립

- [ ] `.claude-plugin/plugin.json` 작성
- [ ] `hooks/hooks.json` 작성
- [ ] `scripts/` 디렉토리 생성 및 기본 스크립트 추가
- [ ] 디렉토리 구조 문서화

### 2차 마일스톤: 기능 검증

- [ ] 플러그인 설치 테스트
- [ ] 커맨드 로딩 검증
- [ ] 에이전트 활성화 검증
- [ ] 후크 실행 검증

### 3차 마일스톤: 자동화 강화

- [ ] PostToolUse 후크로 자동 포맷팅
- [ ] PreToolUse 후크로 SPEC 검증
- [ ] TAG 체인 자동 검증 스크립트

### 최종 마일스톤: 문서화 및 배포

- [ ] 플러그인 구조 가이드 작성
- [ ] 마이그레이션 가이드 작성
- [ ] 배포 체크리스트 작성

---

## ⚠️ 리스크 및 대응 방안

### 리스크 1: 기존 사용자 호환성

**리스크**:
- 기존 `.claude/settings.json` 사용자가 업데이트 시 설정 유실

**대응**:
- 마이그레이션 스크립트 제공
- 자동 감지 및 경고 메시지

### 리스크 2: 환경변수 미지원

**리스크**:
- 일부 환경에서 `${CLAUDE_PLUGIN_ROOT}` 미지원 가능성

**대응**:
- Fallback 경로 제공
- 설치 시 환경변수 검증

### 리스크 3: 성능 저하

**리스크**:
- 후크 실행으로 인한 도구 사용 지연

**대응**:
- 비동기 후크 실행
- 타임아웃 설정 (5초)

---

## 📝 다음 단계

1. **SPEC 승인 후**: `/alfred:2-build SPEC-PLUGIN-001` 실행
2. **TDD 구현**:
   - RED: 플러그인 구조 검증 테스트
   - GREEN: `plugin.json`, `hooks.json` 생성
   - REFACTOR: 스크립트 최적화
3. **문서 동기화**: `/alfred:3-sync` 실행

---

**작성자**: @Goos
**작성일**: 2025-10-10
**참조**: @SPEC:PLUGIN-001
