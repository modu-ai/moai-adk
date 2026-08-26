# research.md — SPEC-CODEX-SESSION-MSG-001

> plan-phase 조사 산출물. 모든 사실 주장은 아래 출처 중 하나에 귀속된다: (a) 2026-08-23 감사 보고서, (b) 본 워크트리 코드 실측(file:line), (c) 2026-08-23 A2A 공식 규격 페치(§3 — 페치한 URL 명시). 사전 지식 단독 판정은 없다.

## §1. 측정 전제 (감사 보고서 인용)

`.moai/reports/t187/codex-support-audit.md` (release/v3.1.3 HEAD 4505df411, 2026-08-23):

| 전제 | 근거 절 |
|------|---------|
| Codex CLI 0.147.0 네이티브 세션 메시징 부재 | §0, §3 |
| `codex mcp add moai -- moai mcp-server` 수동 등록 → 21도구 인식 → `spec_progress` 실호출 성공, CODEX_HOME 격리 (iter-2 D4 정정: 최초 `session_list`로 오기 — 감사 원문 :94의 실호출 도구는 `spec_progress`) | §2.6 (t91 §5) |
| 비대화형 `codex exec` + 무승인정책 MCP 호출 → `"user cancelled MCP tool call"` 실패 | §2.6, §3 gap 6 |
| Codex 세션 moai 훅 미발화(hooks.json 생성기 부재) | §3 gap 1 |
| 테스트 위생: CODEX_HOME 격리, 0.147.0 버전 고정 | §4 |
| t88/M4(배선 생성기)는 대기 상태, 수동 경로가 오늘의 유일한 제품급 MCP 경로 | §0-1, §3 gap 2 |

모든 Codex 동작 문은 codex-cli 0.147.0 기준이다(감사 §4 버전 고정 원칙 준수).

## §2. 코드 표면 조사 (본 워크트리 WT-codex-session-msg @ 76b2c4ece 실측)

| 표면 | 실측 | 시사점 |
|------|------|--------|
| 도구 등록 패턴 | `internal/cli/mcp_server.go:113-135` — `add(name, tool, handler)` 헬퍼가 per-tool enablement(`readMCPToolEnablement`, mcp.yaml 부재 시 전부 enabled)를 게이트한 뒤 `s.AddTool` | 신규 4도구는 동일 패턴으로 등록하면 enablement·콘솔 표면이 공짜로 따라온다 |
| 단일 카탈로그 | `internal/mcp/catalog.go:38-60` — 21도구 선언 + 가드 테스트 `TestMoaiMCPServer_RegistrationMatchesCatalog` 동등성 강제 + 콘솔 스키마가 `MoaiMCPToolNames()`에서 파생 | 카탈로그에 추가하는 것만으로 3면(등록·가드·콘솔)이 정합된다 |
| 세션 레지스트리 | `internal/session/registry.go:84-95` — Entry 스키마 주석 "Schema is frozen per REQ-COORD-002 and REQ-COORD-024" | codex 엔트리를 `active-sessions.json`에 직접 쓰는 설계는 동결 위반이다 → 별도 스토어 결정의 근거 |
| 세션 id 사이드채널 | `internal/session/registry.go:52` — `.moai/state/current-session-id.txt`, SessionStart 훅이 기록 | Claude 세션 자기 식별 파생 후보. 단 codex가 스폰한 mcp-server도 같은 파일을 읽으면 오소유된다 → 명시적 kind+name 등록으로 결정(design.md §4.2) |
| 록·원자쓰기 | `registry_lock_unix.go`, `LockTimeout = 2s`(registry.go:71), `internal/atomicfile` 패키지 | 사서함 동시성 제어에 재사용 — 신규 록 프리미티브 발명 불필요 |
| codex_task 경계 | `internal/cli/codex_task.go`·`mcp_codex.go`·`codex_jobs.go`(잡 레코드 `.moai/state/codex-jobs/`, SPEC-CODEX-PHASE2-001 REQ-CX2-003) | 위임(서브프로세스 잡)과 메시징(살아 있는 세션)의 구분 선례 — §4 비교표와 spec.md §F.1의 근거 |
| C-HRA-008 가드 선례 | `internal/cli/mcp_boundary_test.go` — `TestMCP_NoAskUserQuestion`·`TestMCP_NoInlineGetenv`(정적 grep 가드) | 신규 소스 `mcp_session_msg.go`·`internal/sessionmsg/`에 동일 패턴 |
| 임계값 원천 | `internal/config/defaults.go:293-327` — `DefaultCodexReviewGateTimeout` 등 `var` 선언 패턴(테스트 대체 가능) | 세션 메시징 TTL류도 같은 형태로 |
| 부재 확인 | `grep -rn -e "session_msg" -e "session-msg" internal/ .moai/specs/` → 구현 0건 / SPEC 충돌 0건 (2026-08-23) | 신규 네임스페이스 무충돌 |

## §3. A2A 프로토콜 실측 기록 (운영자 질의 '구글 A2A' — 페치 후 판정)

### §3.1 페치 로그 (2026-08-23)

| URL | 결과 |
|-----|------|
| `https://a2a-protocol.github.io/A2A/latest/` | 404 (카드가 제시한 추정 URL — 부재 확인) |
| `https://github.com/a2a-protocol/A2A` | 404 (조직명 추정 오류 — 실측으로 정정: `a2aproject`) |
| `https://a2aprotocol.ai/` | **성공** — 개요·생태계·전송 문장·spec 링크 확보 |
| `https://github.com/a2aproject/A2A/tree/main/specification` | **성공** — specification 디렉터리 = `a2a.proto` + `json/` |
| `https://raw.githubusercontent.com/a2aproject/A2A/main/specification/a2a.proto` | **성공** — 데이터 모델 원문 (아래 §3.2는 여기서 인용) |
| `https://a2a-protocol.org/latest/` | **성공** — v1.0 공지, Linux Foundation 이관, 주제 목록 |
| `https://a2a-protocol.org/latest/specification/protocol-definition/` | 404 — 세부 프로토콜 정의 페이지는 페치 못함 (§6 Gaps) |

### §3.2 데이터 모델 (a2a.proto 원문에서 축어 인용, package `lf.a2a.v1`)

**AgentCard**: `name`, `description`, `supported_interfaces`, `provider`, `version`, `documentation_url`, `capabilities`, `security_schemes`, `security_requirements`, `default_input_modes`, `default_output_modes`, `skills`, `signatures`, `icon_url`

**Message**: `message_id`(필수), `context_id`(선택), `task_id`(선택), `role`(필수, enum `ROLE_UNSPECIFIED|ROLE_USER|ROLE_AGENT`), `parts`(필수, 반복), `metadata`, `extensions`, `reference_task_ids`

**Part** (oneof `content`): `text`(string) | `raw`(bytes) | `url`(string) | `data`(Value) + 공통 `metadata`, `filename`, `media_type`

**Task**: `id`(필수), `context_id`, `status`(필수, TaskStatus = `state`+`message`+`timestamp`), `artifacts`, `history`(반복 Message), `metadata`

**TaskState**: `TASK_STATE_UNSPECIFIED|SUBMITTED|WORKING|COMPLETED|FAILED|CANCELED|INPUT_REQUIRED|REJECTED|AUTH_REQUIRED` (종결: COMPLETED/FAILED/CANCELED/REJECTED)

**Artifact**: `artifact_id`(필수), `name`, `description`, `parts`(필수), `metadata`, `extensions`

**전송·발견**: a2aprotocol.ai — "Built on existing standards including HTTP, SSE, and JSON-RPC"; 발견은 well-known AgentCard URL(에이전트 목록에 `/.well-known/agent-card.json` 경로 관측); 과제 종결 상태 "completed/failed/canceled". 거버넌스: 원래 Google 개발 → Linux Foundation 기부, Apache 2.0, v1.0. "We recommend MCP for tools and A2A for agents."

### §3.3 JSON 표현의 명명 규약에 대한 고지

원문 페치는 proto(snake_case)다. A2A HTTP JSON 와이어 형상은 proto3 JSON 매핑(camelCase: `messageId`)을 따르는 것으로 알려져 있으나, 본 조사는 `specification/json/` 스키마 디렉터리를 페치하지 않았다(§6). 따라서 camelCase 채택은 "proto3 JSON 명명 규약 준수"의 추론으로 설계 결정으로 기록하며(design.md §3.1), A2A 공식 JSON 스키마와의 필드단위 동일성 주장은 하지 않는다.

## §4. 3축 설계 비교 (카드 지정 — 전부 비용과 함께 기록)

비교 기준: 단일 머신, Codex 폴 기반 수신(네이티브 런타임 부재 P1), MCP 경로 실증(P2), t88 미착지, MCP 서버는 세션 소유 서브프로세스(P5).

| 축 | 내용 | 비용 (측정/구조 근거) | 판정 |
|----|------|----------------------|------|
| (i) A2A HTTP를 전송으로 채택 | 각 세션(또는 공용 데몬)이 A2A HTTP 서버를 띄우고 well-known AgentCard로 발견 | **서버 수명**: 누가 HTTP 엔드포인트를 소유하는가 — Claude 세션도 codex 세션도 데몬을 띄우지 않는다; moai mcp-server는 세션 소유 stdio 서브프로세스(P5, t184 스큐 선례)라 공용 데몬이 되려면 새로운 수명 관리 대상이 필요. **포트 할당**: 다중 세션이 한 머신에서 포트 충돌·방화벽 프롬프트. **발견**: HTTP 서버가 떠 있어야 well-known URL이 답한다 — 파일 발견은 이미 레지스트리로 실증. **보안면**: 로컬 리스너라는 새로운 공격면. 얻는 것은 단일 머신에서 0 (네트워크 이득 없음) | **기각** |
| (ii) A2A 정합 의미론만 | 엔벨로프가 A2A 데이터 모델(AgentCard·Message·Part·Task 수명주기를 참조 형상)에 필드명 수준으로 정렬, 전송은 moai MCP 브로커 + 파일 스토어 하이브리드 | **비용**: 필드 명명 규율(`messageId` 등 A2A 이름 추적 — 정합 테스트로 고정 가능, AC-CSM-001); 정합 범위를 문서로 명시해야 함(어디까지 A2A이고 어디부터 브로커 소유인가 — design.md §3.1). **이득**: 실증된 MCP 경로(P2) 재사용, 네트워크 0, 운영자 질의('구글 A2A')에 대한 실질 응답, 크로스 머신 A2A HTTP 이식 시 번역 계층 불필요(스키마 이식 경로 가시) | **채택 (권장)** |
| (iii) 자체 스키마 | A2A 비정렬 — 엔벨로프 필드명을 자체적으로 정한다 | **비용**: 이식 경로 상실 — 나중에 A2A HTTP를 얹으려면 어차피 번역 계층 필요; 명명이 자의적으로 표류. **이득**: 필드 이름 짓는 자유뿐 — (ii)가 이미 명명을 A2A에 위임하므로 한계 이득 | **기각** |

(ii)의 잔여 의무: 엔벨로프가 A2A 코어 필드를 운반할 것(REQ-CSM-002), 에이전트 기록이 AgentCard 참조 형상일 것(REQ-CSM-003), 이식 경로가 문서에 보일 것(design.md §7).

## §5. 주소 체계 결정 근거 (카드 지정 질문: 레지스트리 재사용 vs 별도)

Codex 세션은 moai 훅을 발화하지 않는다(P4) → SessionStart 자동 등록 경로가 없다. 선택지는 (a) `active-sessions.json`에 직접 기록, (b) 브로커 소유 별도 스토어 + MCP 자가등록.

- (a) 기각: Entry 스키마 동결(REQ-COORD-024, §2 실측) 위반; 훅이 파일을 재작성하는 경로와 충돌 위험; kind 개념이 없는 레지스트리 표면 오염.
- (b) 채택: `.moai/state/session-msg/agents/<agentId>.json`에 A2A AgentCard 참조 형상 기록. `session_msg_register`는 kind+name으로 멱등 등록(안정적 agentId 반환)하므로 Codex가 자기 id를 저장할 부담도 없다. Claude 세션은 동일한 대칭 절차를 쓴다(사이드채널 파생은 오소유 위험 — §2 — 으로 v1 범위 밖).

## §6. Gaps (본 조사가 검증하지 못한 것)

- A2A 공식 JSON 스키마(`specification/json/`)와 프로토콜 정의 세부 페이지(404)를 페치하지 못했다 — §3.3 고지대로 camelCase는 proto3 JSON 규약 준수 추론이다.
- codex-cli 0.147.0 **이후** 버전의 MCP 서버 수명 주기 변동(세션당 재스폰 여부 등)은 미확인 — e2e의 재시작 절차(0단계)가 버전 무관하게 성립하도록 절차를 짰다.
- Claude Code MCP 클라이언트의 도구 목록 리프레시 시점(`/mcp` 재접속 vs 세션 재시작)은 공식 문장을 페치하지 않고 선례(t184)에 기대고 있다 — e2e 0단계는 세션 재시작으로 보수적으로 통일.
- `codex mcp add` 등록의 지속 범위(프로젝트 vs 글로벌 config.toml)는 감사 보고서가 재현 절차를 다루지 않아 e2e에서 격리 CODEX_HOME으로 무효화한다(위생 §1).
