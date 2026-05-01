---
id: SPEC-GLM-MCP-001
artifact: acceptance
version: "0.1.0"
created_at: 2026-05-01
parent_spec: spec.md
---

# SPEC-GLM-MCP-001 Acceptance Criteria

22개 Given/When/Then 시나리오. 모든 10개 REQ (REQ-GMC-001 ~ REQ-GMC-010) 를 커버하며, edge cases 를 포함한다. 각 시나리오는 자동화된 단위/통합 테스트로 검증 가능하다.

## REQ Coverage Map

| GWT # | 검증 대상 REQ | 카테고리 |
|-------|---------------|---------|
| GWT-01 | REQ-GMC-001, REQ-GMC-002 | Happy path — 단일 도구 enable |
| GWT-02 | REQ-GMC-001, REQ-GMC-007 | Happy path — `all` enable + status |
| GWT-03 | REQ-GMC-007 | Happy path — disable single |
| GWT-04 | REQ-GMC-007 | Happy path — disable all + status |
| GWT-05 | REQ-GMC-002 | Opt-in 원칙 — `moai init` 부영향 |
| GWT-06 | REQ-GMC-002 | Opt-in 원칙 — `moai update` 부영향 |
| GWT-07 | REQ-GMC-003 | 자격증명 재사용 — 토큰 매핑 검증 |
| GWT-08 | REQ-GMC-003 | 자격증명 재사용 — `~/.moai/.env.glm` 부재 시 실패 |
| GWT-09 | REQ-GMC-004 | Edge case — Node.js 부재 |
| GWT-10 | REQ-GMC-004 | Edge case — npx 부재 (node 만 존재) |
| GWT-11 | REQ-GMC-005 | 충돌 처리 — 기존 `glm-vision` 키 존재 (대화형 보존 선택) |
| GWT-12 | REQ-GMC-005 | 충돌 처리 — 백업 옵션 선택 |
| GWT-13 | REQ-GMC-005 | 충돌 처리 — 비표준 사용자 키 (`zai-mcp-server`) 보존 |
| GWT-14 | REQ-GMC-006 | Pro 플랜 미가입 — vision enable 거부 |
| GWT-15 | REQ-GMC-006 | Pro 플랜 미가입 — websearch/webreader 영향 없음 |
| GWT-16 | REQ-GMC-007 | Status 출력 정확성 |
| GWT-17 | REQ-GMC-008 | Scope project 옵션 — `.mcp.json` 작성 |
| GWT-18 | REQ-GMC-008 | Scope 기본값 — user scope 작성 |
| GWT-19 | REQ-GMC-009 | 모드 전환 무영향 — `moai cc` 후 엔트리 보존 |
| GWT-20 | REQ-GMC-009 | 모드 전환 무영향 — `moai cg` 후 엔트리 보존 |
| GWT-21 | REQ-GMC-010 | 비대화형 — 충돌 시 즉시 실패 |
| GWT-22 | REQ-GMC-010 | 비대화형 — 원자적 쓰기 (M3 결과 명세는 D2 에서 확정) |

---

## GWT Scenarios

### GWT-01: 단일 도구 enable (Happy Path)

**Given** 사용자가 `moai glm` 모드로 전환되어 있고, `~/.moai/.env.glm` 에 유효한 `GLM_AUTH_TOKEN=sk-glm-xxx` 가 존재하며, `~/.claude.json` 의 `mcpServers` 에는 본 SPEC 의 정규 키(`glm-vision`/`glm-websearch`/`glm-webreader`) 가 모두 부재한 상태이다. 시스템 PATH 에 `node` (>=18) 와 `npx` 가 존재한다.

**When** 사용자가 터미널에서 `moai glm tools enable websearch` 를 실행한다.

**Then** 명령은 종료 코드 0 으로 5초 이내에 종료되며, `~/.claude.json` 의 `mcpServers["glm-websearch"]` 엔트리가 다음을 포함한다: (a) `command: "npx"`, (b) `args` 배열에 `-y` 와 `@z_ai/mcp-server` 와 `websearch` 가 포함됨, (c) 환경변수 `Z_AI_API_KEY` 가 `GLM_AUTH_TOKEN` 의 값과 동일하게 설정됨. 다른 두 도구(`glm-vision`, `glm-webreader`) 는 추가되지 않는다. 표준 출력에 활성화 성공 메시지가 표시된다.

---

### GWT-02: `all` enable + status (Happy Path)

**Given** GWT-01 의 사전 조건이 동일하게 성립하며, 추가로 사용자의 GLM 구독 티어가 Pro 이상이거나 검출 불가 상태이다 (Pro fallback).

**When** 사용자가 `moai glm tools enable all` 을 실행한 후 `moai glm tools status` 를 실행한다.

**Then** `enable all` 명령은 정규 키 3종(`glm-vision`, `glm-websearch`, `glm-webreader`) 모두를 `~/.claude.json` 의 `mcpServers` 에 추가한다. 후속 `status` 명령은 정확히 다음 형식으로 3개 행을 출력한다 — `glm-vision: enabled (user scope)`, `glm-websearch: enabled (user scope)`, `glm-webreader: enabled (user scope)`. 사용자 정의 비표준 키는 출력에 포함되지 않는다.

---

### GWT-03: 단일 도구 disable (Happy Path)

**Given** GWT-02 의 결과 상태가 성립한다 (3종 모두 enabled).

**When** 사용자가 `moai glm tools disable vision` 을 실행한다.

**Then** `~/.claude.json` 의 `mcpServers["glm-vision"]` 엔트리가 제거된다. `glm-websearch` 와 `glm-webreader` 엔트리는 변경 없이 보존된다. 사용자 정의 비표준 mcpServers 엔트리(예: 사용자가 직접 추가한 `context7`) 도 보존된다.

---

### GWT-04: `all` disable + status (Happy Path)

**Given** GWT-02 결과 상태 (3종 모두 enabled).

**When** 사용자가 `moai glm tools disable all` 을 실행한 후 `moai glm tools status` 를 실행한다.

**Then** `~/.claude.json` 의 `mcpServers` 에서 정규 키 3종이 모두 제거된다. 후속 `status` 명령은 다음을 출력한다 — `glm-vision: disabled`, `glm-websearch: disabled`, `glm-webreader: disabled`. 종료 코드 0.

---

### GWT-05: `moai init` 후 자동 등록 없음 (Opt-in 원칙)

**Given** 빈 디렉토리에서 사용자가 `moai init test-project` 를 실행하려는 상태이다.

**When** `moai init test-project` 명령이 완료된다.

**Then** 생성된 프로젝트의 `.mcp.json` (만약 템플릿이 생성한다면) 과 사용자의 `~/.claude.json` 어디에도 본 SPEC 의 정규 키(`glm-vision`/`glm-websearch`/`glm-webreader`) 가 자동으로 추가되지 않는다. 기존 템플릿이 제공하는 다른 mcpServers 엔트리(`context7`, `sequential-thinking`, `moai-lsp`) 는 영향을 받지 않는다.

---

### GWT-06: `moai update` 후 자동 등록 없음 (Opt-in 원칙)

**Given** 기존에 `moai glm tools enable websearch` 로 `glm-websearch` 만 활성화한 사용자의 프로젝트.

**When** 사용자가 `moai update` 를 실행한다.

**Then** `~/.claude.json` 의 `glm-websearch` 엔트리는 보존되며, `glm-vision` 또는 `glm-webreader` 가 자동으로 추가되지 않는다. `moai update` 의 기존 머지 정책 외에 본 SPEC 이 추가하는 동작은 없다.

---

### GWT-07: 자격증명 토큰 재사용 검증

**Given** `~/.moai/.env.glm` 의 내용이 `GLM_AUTH_TOKEN=sk-glm-test-12345` 이다.

**When** 사용자가 `moai glm tools enable webreader` 를 실행한다.

**Then** `~/.claude.json` 의 `mcpServers["glm-webreader"].env` 객체에서 `Z_AI_API_KEY` 의 값이 `sk-glm-test-12345` 와 정확히 일치한다 (D1 에서 확정될 추가 호환 변수도 동일 값으로 매핑됨). 명령은 사용자에게 신규 토큰 입력 프롬프트를 제시하지 않는다.

---

### GWT-08: GLM_AUTH_TOKEN 부재 시 실패

**Given** `~/.moai/.env.glm` 파일이 존재하지 않거나 `GLM_AUTH_TOKEN` 키가 비어 있다. 사용자는 비대화형 환경(예: `--yes` 플래그) 에서 명령을 실행한다.

**When** 사용자가 `moai glm tools enable vision --yes` 를 실행한다.

**Then** 명령은 종료 코드 비-0 으로 즉시 실패한다. 표준 에러 출력에 다음을 모두 포함하는 메시지가 표시된다: (a) `GLM_AUTH_TOKEN` 이 `~/.moai/.env.glm` 에서 발견되지 않았다는 사실, (b) `moai glm` 모드 활성화로 토큰을 먼저 설정하라는 안내. `~/.claude.json` 은 변경되지 않는다 (REQ-GMC-010 의 원자성).

---

### GWT-09: Node.js 부재 graceful 실패

**Given** 시스템 PATH 에서 `node` 와 `npx` 가 모두 발견되지 않는다 (예: PATH 가 빈 임시 환경).

**When** 사용자가 `moai glm tools enable all` 을 실행한다.

**Then** 명령은 종료 코드 비-0 으로 실패한다. 표준 에러 출력 메시지는 다음 3요소를 모두 포함한다: (a) Node.js 가 PATH 에서 발견되지 않음, (b) Node.js 18+ 가 권장됨, (c) 설치 가이드 URL 또는 `nvm install --lts` 등 즉시 실행 가능한 권장 명령. 어떠한 settings 파일도 변경되지 않는다.

---

### GWT-10: npx 부재 (node 만 존재)

**Given** 시스템 PATH 에 `node` 는 존재하지만 `npx` 가 부재한다 (드문 경우, 손상된 Node 설치).

**When** 사용자가 `moai glm tools enable websearch` 를 실행한다.

**Then** 명령은 종료 코드 비-0 으로 실패하며, GWT-09 와 동일한 형식의 에러 메시지를 출력한다 (단, 메시지는 "npx not found, but node is present" 와 같이 구체적 진단을 포함). settings 는 변경되지 않는다.

---

### GWT-11: 기존 `glm-vision` 충돌 (대화형 — 보존 선택)

**Given** `~/.claude.json` 의 `mcpServers["glm-vision"]` 에 사용자가 임의로 추가한 엔트리 (예: 다른 패키지를 가리키는 command/args) 가 이미 존재한다. 사용자는 대화형 터미널에서 명령을 실행한다.

**When** 사용자가 `moai glm tools enable vision` 을 실행하고, 충돌 처리 프롬프트에서 옵션 (i) "기존 보존, 등록 건너뜀" 을 선택한다.

**Then** 기존 `glm-vision` 엔트리는 변경 없이 보존되고, 본 SPEC 이 의도한 `@z_ai/mcp-server` 엔트리는 추가되지 않는다. 명령은 종료 코드 0 또는 비-0(D2 에서 확정) 으로 종료하며, 사용자에게 명확히 "skipped due to existing entry" 메시지를 출력한다.

---

### GWT-12: 기존 충돌 (대화형 — 백업 후 덮어쓰기)

**Given** GWT-11 과 동일한 충돌 상태.

**When** 사용자가 `moai glm tools enable vision` 을 실행하고 옵션 (ii) "백업 후 덮어쓰기" 를 선택한다.

**Then** 기존 `glm-vision` 엔트리가 인접 키 `glm-vision.bak.<timestamp>` (timestamp 는 ISO-8601 또는 unix epoch 형식; D2 에서 확정) 에 보존되고, `glm-vision` 키에는 본 SPEC 의 새 엔트리가 기록된다. 명령은 종료 코드 0 으로 성공한다.

---

### GWT-13: 비표준 사용자 키 보존

**Given** `~/.claude.json` 의 `mcpServers` 에 사용자가 추가한 비표준 키 `zai-mcp-server` 와 `my-custom-server` 가 존재한다. 본 SPEC 의 정규 키는 모두 부재한다.

**When** 사용자가 `moai glm tools enable all` 을 실행한다.

**Then** 정규 키 3종이 새로 추가되고, 비표준 키 `zai-mcp-server` 와 `my-custom-server` 는 변경 없이 보존된다. `moai glm tools status` 출력은 정규 키 3종만 포함하고 비표준 키는 무시한다 (EX-4 의 원칙).

---

### GWT-14: Pro 플랜 미가입 시 vision 거부

**Given** 사용자의 GLM 구독 티어가 Lite 로 검출되거나 (티어 검출 mechanism 의 정확한 호출은 D2 에서 확정), 또는 검출이 명시적으로 "Lite" 응답을 반환한다.

**When** 사용자가 `moai glm tools enable vision` 을 실행한다.

**Then** 명령은 종료 코드 비-0 으로 실패한다. 에러 메시지는 다음을 모두 포함한다: (a) Vision 도구가 Pro 플랜 이상을 요구함, (b) 검출된 현재 티어 ("Lite"), (c) 업그레이드 안내 URL. `~/.claude.json` 의 `glm-vision` 엔트리는 추가되지 않는다.

---

### GWT-15: Pro 플랜 미가입 — websearch/webreader 영향 없음

**Given** GWT-14 와 동일한 Lite 티어 상태.

**When** 사용자가 `moai glm tools enable websearch webreader` 또는 `moai glm tools enable all` 을 실행한다.

**Then** `enable websearch webreader` 의 경우 두 도구가 정상 등록되며 종료 코드 0 으로 성공한다. `enable all` 의 경우 `glm-websearch` 와 `glm-webreader` 는 등록되지만 `glm-vision` 등록은 GWT-14 의 Pro 플랜 검사로 차단된다 — 부분 성공 시 종료 코드는 비-0 이며 어떤 도구가 성공/실패했는지 명확히 출력한다.

---

### GWT-16: Status 명령 출력 정확성

**Given** `~/.claude.json` 의 `mcpServers` 에 `glm-vision` 만 존재하고, 추가로 비표준 `my-custom` 엔트리가 존재한다.

**When** 사용자가 `moai glm tools status` 를 실행한다.

**Then** 출력은 정확히 3개 행을 포함하며 — `glm-vision: enabled (user scope)`, `glm-websearch: disabled`, `glm-webreader: disabled`. 비표준 `my-custom` 엔트리는 출력에 포함되지 않는다. 종료 코드 0.

---

### GWT-17: `--scope project` 옵션

**Given** 빈 프로젝트 루트에 `.mcp.json` 이 부재한다. `~/.claude.json` 도 본 SPEC 정규 키 부재.

**When** 사용자가 `moai glm tools enable websearch --scope project` 를 실행한다.

**Then** 프로젝트 루트에 `.mcp.json` 파일이 생성되거나 갱신되며 `mcpServers["glm-websearch"]` 엔트리를 포함한다. `~/.claude.json` 은 변경되지 않는다. `moai glm tools status` 호출 시 `glm-websearch: enabled (project scope)` 가 출력된다.

---

### GWT-18: 기본 scope = user

**Given** `--scope` 옵션 없이 명령 실행.

**When** 사용자가 `moai glm tools enable websearch` (옵션 부재) 를 실행한다.

**Then** `~/.claude.json` 에 엔트리가 추가되며 프로젝트의 `.mcp.json` 은 변경되지 않는다 (또는 부재한 채로 유지된다). 출력에 `(user scope)` 가 표시된다.

---

### GWT-19: `moai cc` 모드 전환 후 보존

**Given** 사용자가 `moai glm tools enable all` 로 3종 모두 활성화한 상태.

**When** 사용자가 `moai cc` 명령으로 Anthropic API 모드로 전환한다 (D3 에서 정의되는 mock injection mechanism 으로 외부 API 호출은 모킹).

**Then** `~/.claude.json` 의 `mcpServers` 의 정규 키 3종은 변경 없이 보존된다. settings.json 의 `env`, `model`, `apiKeyHelper` 등 모드별 키만 갱신된다. 후속 `moai glm tools status` 출력은 GWT-04 직전 상태와 동일하다 (3종 모두 enabled).

---

### GWT-20: `moai cg` 모드 전환 후 보존

**Given** GWT-19 와 동일한 사전 상태.

**When** 사용자가 `moai cg` 명령으로 CG 모드(tmux 기반 GLM teammate) 로 전환한다.

**Then** `~/.claude.json` 의 정규 키 3종 mcpServers 엔트리는 변경 없이 보존된다. tmux 세션 환경변수와 settings.local.json 의 모드별 키만 갱신된다.

---

### GWT-21: 비대화형 환경에서 충돌 시 즉시 실패

**Given** GWT-11 과 동일한 충돌 상태(기존 `glm-vision` 엔트리 존재). 환경은 비대화형(`stdin` 이 TTY 가 아님 또는 `--yes` 플래그 부재).

**When** 사용자가 `moai glm tools enable vision` 을 실행한다 (`--yes` 또는 `--force` 미지정).

**Then** 명령은 대화형 프롬프트를 띄우지 않고 즉시 종료 코드 비-0 으로 실패한다. 에러 메시지는 (a) 충돌 발생 사실, (b) 비대화형 환경 감지, (c) 권장 해결 방법 (`--scope project` 또는 `--force` 또는 수동 편집) 을 명시한다. `~/.claude.json` 은 변경되지 않는다 — partial write 없음.

---

### GWT-22: 원자적 쓰기 (M3 결과 명세는 D2 에서 확정)

**Given** `~/.claude.json` 이 정상 상태이며, 임시로 인위적인 디스크 쓰기 실패(예: 디렉토리 권한 박탈, mock `os.WriteFile` 실패) 를 발생시킬 수 있다.

**When** 사용자가 `moai glm tools enable all` 을 실행하고, settings 쓰기 단계에서 시뮬레이션된 디스크 실패가 발생한다.

**Then** 명령은 종료 코드 비-0 으로 실패한다. `~/.claude.json` 은 명령 실행 이전 상태와 byte-for-byte 동일하다 (atomic rename 또는 임시 파일 + 검증된 swap mechanism 사용 — 정확한 mechanism 은 D2 에서 확정). 어떠한 정규 키도 부분 적용되지 않는다 (REQ-GMC-010).

---

## Verification Strategy

- **단위 테스트**: GWT-01 ~ GWT-22 각각 1:1 매핑되는 Go 테스트 함수 (`TestGlmTools_GWT01_EnableSingle` 등 검색 가능한 prefix 사용).
- **통합 테스트**: 실제 `~/.claude.json` 대신 `t.TempDir()` 의 격리된 settings 파일을 사용 (CLAUDE.local.md §6 의 test isolation 규칙 준수).
- **모킹**: Z.AI API 호출(Pro 플랜 검출), Node.js 부재 검사(`exec.LookPath` 모킹) 는 인터페이스 주입을 통해 테스트한다 (정확한 mechanism 은 D3 refinement).
- **회귀 방지**: GWT-05, GWT-06, GWT-13, GWT-19, GWT-20 은 본 SPEC 의 변경이 기존 동작을 깨지 않는지 검증하는 회귀 테스트 역할을 한다.

## Coverage Summary

- 10/10 REQs covered.
- EARS 분포: Ubiquitous (REQ-001/002) 6 GWTs, Event-Driven (REQ-003/004/005/007) 11 GWTs, State-Driven (REQ-006) 2 GWTs, Optional (REQ-008) 2 GWTs, Unwanted (REQ-009/010) 4 GWTs.
- Edge cases: Node.js 부재(GWT-09/10), 자격증명 부재(GWT-08), 충돌(GWT-11/12/21), Pro 미가입(GWT-14/15), 비대화형(GWT-21/22), 모드 전환(GWT-19/20), scope 분기(GWT-17/18).
