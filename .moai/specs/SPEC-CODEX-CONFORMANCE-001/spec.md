---
id: SPEC-CODEX-CONFORMANCE-001
title: "Codex CLI 0.161.0 adapter conformance re-measure"
version: "0.1.0"
status: in-progress
created: 2026-10-09
updated: 2026-10-09
author: manager-spec
priority: P1
phase: "v3.2.0"
module: "internal/cli"
lifecycle: spec-anchored
tags: "codex, conformance, fixtures, upstream, adapter"
tier: M
---

# SPEC-CODEX-CONFORMANCE-001 — Codex CLI 0.161.0 어댑터 정합성 재측정

## HISTORY

- 0.1.0 (2026-10-09): 팩토리 카드 t1607(CARD-4) plan-phase 최초 작성. 상위 근거: `.moai/reports/release-update-20261008/upstream-update-20261008.md` 축 2 (Codex 0.160.1 → 0.161.0 안정 승격 관측), `summary.json` card_proposals.CARD-4, 큐레이팅 릴리스 본문 `codex-0.161.0-release-body.md`. 이 SPEC은 기존 어댑터의 정합성 재측정이지 신기능 구축이 아니다.

## §A. 측정 전제 (Verified baseline)

### §A.1 삼방 버전 드리프트 (leader 지시 판정 대상)

| 축 | 값 | 근거 |
|---|---|---|
| 리포 fixture 핀 | `codex-0.160.0` (`internal/cli/testdata/codex-0.160.0/`) | 본 트리 `ls` 실측 |
| 설치 바이너리 | codex-cli 0.160.1 | 카드 배차문 기재 (2026-10-08T13:11Z 실측 — 본 SPEC 작업환경에서 재측정 아님, §A.5 갭) |
| 업스트림 안정 | 0.161.0 (rust-v0.161.0, 2026-10-07T15:58:45Z 승격, 0.162.0-alpha는 여전히 prerelease) | 상위 리포 npm/gh 3-way 관측 |

**갱신 판정 (plan §A에 상세)**: fixture 목표 버전은 0.161.0이다. 0.160.0 세트는 소비자 이전과 함께 교체한다(유지 아님) — 근거는 fixture 디렉터의 커밋된 README와 `.moai/docs/factory-managed-session.md` 의 재생성 규율("최소 지원 codex 버전이 올라가면 사본을 다시 만든다")이다. 라이브 바이너리 측정은 자동 AC에서 제외하고, fixture가 판정할 수 없는 항목만 수동 검증으로 기록한다.

### §A.2 현재 핀의 소비자 — 본 트리 전수 재측정 (카드 지시 이행)

리포트(t1538 트리)의 소비자 목록을 본 트리(81786284e)에서 다시 잤다. 결과는 리포트와 일치하며, 이 목록이 본 SPEC의 소비자 정의다:

- `internal/cli/managed_hardening_test.go:523` — `hardenCompileSchema`가 `testdata/codex-0.160.0/<file>`을 읽고, `:533`에서 스키마 리소스 URL `https://moai.invalid/codex-0.160.0/<file>`을 만든다. 소비 스키마 목록은 `hardenResultSchemas` 지도 7종(+`JSONRPCError.json`)이다.
- `internal/cli/managed_codex_tui_test.go:82` — `tuiResumeHelpFixture = "testdata/codex-0.160.0/resume-help.txt"` (vendored `codex resume --help` 텍스트).
- 문서 참조(코드 소비자 아님): `.moai/docs/factory-managed-session.md:51,120`, `CHANGELOG.md:80`. 본 SPEC은 코드 소비자만 이전시키고, 문서는 M4에서 버전 표기를 갱신한다.

### §A.3 어댑터 표면 실측 (변경 판정의 재료)

- **auth 2단 사다리** (`internal/cli/mcp_codex.go:2196` 이후, SPEC-CODEX-LAUNCHER-001 M1): Stage 1이 `<CODEX_HOME>/auth.json`을 순수 판정하고, `classifyCodexAuth`(:2440 근처)의 문서와 구현은 **파일 부재를 포함한 모든 기각을 stage 2 하강으로 처리**한다("a rejected file — … or absent — is a DESCENT to stage 2, never an outcome"). Stage 2는 `codex login status` 전 스트림을 정통 라인 문법 `^[ \t]*logged in using (chatgpt|api key)[ \t]*\r?$`로 분류하고, 문법 불일치는 `codexAuthUnknown`(갭, 판정 아님)이다. 즉 **auth.json 부재는 이미 "미인증"으로 읽히지 않는다** — keyring 저장 전환(#49361)에 대한 어댑터 위험은 stage 2 문법이 0.161.0 출력을 여전히 분류하는가로 환원된다. 시험 파일: `internal/cli/codex_auth_ladder_test.go`, `internal/cli/mcp_codex_test.go`.
- **resume --help 소비**: `probeManagedCodexRemote`(`internal/cli/managed_codex_tui.go:100-119`)가 `codex resume --help`를 실행해 `managedCodexRemoteSupport`로 `--remote` + `--remote-auth-token-env` 동시 보유를 판정하고, 미지원이면 문서화된 폴백으로 떨어진다. 관리 TUI의 재개 인자는 `resume --remote <url> --remote-auth-token-env <env>`(:338)다.
- **서버 요청 응답 정책**: 관리 소유자는 **소유(owner) 턴**의 서버-발생 요청을 기각 정책표로 답하고, 답 본문 7종을 vendored 스키마에 대해 `TestManagedServerRequestPolicyMatchesCodexSchema`로 검증한다(별도 정확 단언은 `TestManagedCodexServerRequestPolicy`). 오퍼레이터가 붙은 턴의 요청은 소유자가 답지 않고 오퍼레이터 TUI에 남긴다 — `leavesForOperator`(`internal/cli/managed_codex_tui.go:654`, 분기 `internal/cli/managed_codex_factory.go:435`, 로그 `Factory server request left for the operator` `managed_codex_tui.go:670`). 이 오퍼레이터 경로는 §D.4 REQ-CONF-008의 NO-OP 판정에서 알려진 노출면으로 다룬다.
- **훅 이벤트 적응 계열**: `codex_event_adaptation_test.go`는 PreCompact/PostCompact/PermissionRequest/Interrupt의 훅-이벤트 매핑 골든 시험이다 — **thread/rollout 이벤트 소비면이 아니다**(카드 표현 정정, §A.4).

### §A.4 카드 앵커 정정 (드리프트된 인용 3건)

| 카드/리포트 인용 | 본 트리 실측 | 파급 |
|---|---|---|
| `internal/cli/closed_sets.go:96` | 파일은 `internal/config/closed_sets.go`이고 라인 96은 일치 (`DefaultCodexAuditModel = "gpt-6.1-sol"`) | 경로 정정 — 핀 자체는 유효 |
| `internal/config/defaults.go:1647` | 해당 라인은 constitution 기본값. Codex 핀 `{gpt-6.1-sol, high}`는 `:1545` 근처(SPEC-MODEL-MATRIX-UPDATE-001 주석과 함께) + `internal/config/audit_models.go:95` + `internal/cli/mcp_codex.go:62-65` 단일 소싱 | 라인 정정 — 핀 자체는 유효 |
| "codex_event_adaptation 계열이 thread/rollout 이벤트 소비" | 위 계열은 훅-이벤트 매핑 시험. rollout을 소비하는 프로덕션 표면은 `internal/cli/codex_role_fingerprint.go`가 유일한 grep 적중 | 재검증 표면을 resume-help 프로브·관리 프레임 처리로 특정 |

### §A.5 갭 (본 계획이 측정하지 않은 것)

- 0.161.0 관측 상태(갱신 — 레인 실측 2026-10-08T16:14-16:15Z, 본 트리): `npx -y @openai/codex@0.161.0 --version` → `codex-cli 0.161.0`, exit 0; `npx -y @openai/codex@0.161.0 app-server --help`에 `generate-json-schema` 하위명령 확인, exit 0 — 서브명령 존재 갭은 해소됐고 M1이 실제 생성 **출력**을 관측한다. 잔여 미관측: 0.161.0 `resume --help` 옵션 집합(재생성과 함께 관측), keyring 저장 시 `codex login status` 출력 형태, thread-resume 재생 프레임의 실제 추가 필드 — fixture 생성 환경에서 확정.
- 0.160.1 설치 바이너리는 본 작업환경에서 재측정하지 않았다(카드 배차문 인용).
- `codex_role_fingerprint.go`의 rollout 소비가 0.161.0에서 형태 변화를 겪는지는 미측정 — M3 재확인 항목으로 이월(범위 밖 발견 시 비목표로 기록).

## §B. 문제 진술

리포의 Codex 어댑터 정합성 스위트는 0.160.0 fixture에 고정돼 있고, 업스트림 안정은 0.161.0으로 승격했다. 최소 지원 버전이 움직였으므로 커밋된 재생성 규율에 따라 vendored 스키마·resume-help fixture를 0.161.0으로 다시 만들고, 어댑터가 0.161.0이 만질 수 있는 표면(auth 저장소 전환, resume 재개 출력·재생 페이로드, 승인 확장 의미, 기본 모델 카탈로그)을 재검증해야 한다. 재검증 없이는 스키마 가드가 낡은 프로토콜 형상을 검사하는 것을 넘어, 실제 세션이 새 형상을 만났을 때 어댑터가 오분류·오락 없이 버티는지 아무것이 보장하지 못한다.

## §C. 범위

### 포함

1. **fixture 갱신**: `internal/cli/testdata/codex-0.160.0/` → `internal/cli/testdata/codex-0.161.0/` 재생성(app-server 응답 스키마 8종 + `resume-help.txt` + README) 및 소비자 2곳 이전.
2. **auth.json/keyring**: 부재 시 stage 2 하강 계약 유지 검증 + stage 2 문법의 0.161.0 출력 정합 확인(가능하면 sanitized 출력 샘플 fixture, 불가하면 수동 검증 기록).
3. **resume 표면**: vendored `resume --help` 0.161.0 재검증(`managedCodexRemoteSupport`) + 재개 재생 페이로드 내성(#49599 재생 이력이 확장 필드를 실어 와도 디코드 실패·오락 턴 실패 없음).
4. **filesystem escalation #49353**: 기록된 NO-OP — 기각 정책 불변 검증으로 충분하다는 판정과 그 근거를 SPEC에 명시.
5. **C-1 모델 핀**: `{gpt-6.1-sol, high}` 핀(`internal/config/closed_sets.go:96`, `internal/config/defaults.go` Codex 핀)의 정합 확인 — 변경 없음, 기존 시험 재실행.

### Out of Scope — 업스트림 신기능 채택

- `/mcp login <name>` TUI, 음성 대화 장치, Daybreak/Cyber 라우팅, Bedrock 멀티에이전트 V2/Ultra/GovCloud, enterprise MCP auth, npm dist-tag 관리 — moai 어댑터 표면이 없어(c-9~14 판정) 채택 대상이 아니다.

### Out of Scope — keyring 저장소 클라이언트 구현

- moai가 keyring 리더를 새로 두지 않는다. keyring 저장 인증 상태의 권위 소스는 codex 바이너리 자체(`codex login status`, stage 2)이며 어댑터는 그 문법 정합만 책임진다. keyring 자체 접근은 상위 기능 구축으로 본 SPEC 범위 밖.

### Out of Scope — 라이브 바이너리 CI 게이트

- 설치 바이너리에 의존하는 검사를 자동 AC나 CI 요구사항으로 만들지 않는다. 0.161.0 실행이 필요한 확인은 fixture 생성 환경에서 1회 수행하거나 수동 검증으로 기록한다(§A.1 갱신 판정).

### Out of Scope — 최소 지원 버전 정책의 정의

- 본 SPEC은 현재 안정(0.161.0)으로의 재측정만 한다. "언제 최소 지원 버전을 올리는가"의 정책은 별도 후속으로 이월한다.

## §D. 요구사항 (GEARS)

### D.1 fixture 갱신

- **REQ-CONF-001** — The conformance fixture set shall carry vendored artifacts generated from a codex-cli 0.161.0 binary under `internal/cli/testdata/codex-0.161.0/`: the eight response schemas `hardenResultSchemas` consumes (ApplyPatchApprovalResponse, ExecCommandApprovalResponse, CommandExecutionRequestApprovalResponse, FileChangeRequestApprovalResponse, PermissionsRequestApprovalResponse, McpServerElicitationRequestResponse, DynamicToolCallResponse, JSONRPCError), a `resume-help.txt` regenerated from 0.161.0, and a README naming codex-cli 0.161.0. 세부 규율: 스키마 파일은 기계 생성 산출물 그대로다(손편집 금지).
- **REQ-CONF-002** — **When** the 0.161.0 fixture set lands, the fixture consumers shall reference the new directory — `hardenCompileSchema` (path and resource-URL literals, `internal/cli/managed_hardening_test.go`) and `tuiResumeHelpFixture` (`internal/cli/managed_codex_tui_test.go`) — and the `codex-0.160.0` literal occurrence count under `internal/cli/` shall be zero.
- **REQ-CONF-003** — `TestManagedServerRequestPolicyMatchesCodexSchema` shall pass with the owner's seven result-body answers validated against the vendored 0.161.0 schemas.

### D.2 resume 표면

- **REQ-CONF-004** — The vendored 0.161.0 `codex resume --help` text shall be classified by `managedCodexRemoteSupport` consistently with the option set the text actually offers. **When** the 0.161.0 text offers fewer than both `--remote` and `--remote-auth-token-env`, the managed TUI probe shall report remote support absent and the managed attach shall keep its documented fallback path (부착 실패가 아니다).
- **REQ-CONF-005** — **When** a resumed thread replays the 0.161.0 authoritative replay history (#49599), including frames carrying fields the adapter does not model, the managed owner's frame handling shall consume the replay without a decode failure or a marked turn failure and shall ignore the fields it does not model.

### D.3 auth 상태 (§A.3 계약 유지)

- **REQ-CONF-006** — **While** the two-stage auth ladder is in service, an absent or unreadable `<CODEX_HOME>/auth.json` shall descend to stage 2 (`codex login status`) and shall never classify as a definitive "not authenticated" provider verdict; a stage-2 result without a whole-line grammar match shall remain `codexAuthUnknown` — 읽을 수 없는 프루브는 갭이지 판정이 아니다.
- **REQ-CONF-007** — The stage-2 whole-line grammar (`codexAuthStatusLine`) shall classify a codex-cli 0.161.0 `codex login status` output line by the same rule as the 0.160-era line. **Where** capturing a sanitized 0.161.0 output sample is impossible in the fixture-generation environment, the run phase shall record a manual-verification observation that carries its own evidence — the capturing command, the verbatim output, and the binary version (verification-claim-integrity §3 형식) — in progress.md §E.2, with the unauthenticated 0.161.0 output shape captured in the fixture-generation environment as the minimum observation and only the keyring-logged-in variant eligible as the recorded gap (증거 없는 갭 기록은 manual verification가 아니라 미측정이다).

### D.4 불변 표면 (NO-OP 검증)

- **REQ-CONF-008** — **While** a server-request arrives on an owner turn, the managed owner's decline policy table shall remain unchanged across the fixture refresh; the `{gpt-6.1-sol, high}` audit pin (`internal/config/closed_sets.go:96` constant, the `internal/config/defaults.go` Codex pin) shall remain unchanged; and the existing test families (`TestManagedCodexServerRequestPolicy`, the `internal/config` package tests) shall verify both on the refreshed fixtures. 알려진 노출면(기록 전용 — 새 승인 처리 금지): 오퍼레이터가 붙은 턴의 요청은 `leavesForOperator`(`internal/cli/managed_codex_tui.go:654`, 분기 `managed_codex_factory.go:435`)에 따라 오퍼레이터 TUI에 남는다. 승인된 파일시스템 escalation(#49353)의 의미 확장은 그 **오퍼레이터 승인** 경로가 풀 수 있는 권한 폭을 넓힐 수 있으나, 이는 codex 자체 승인 프롬프트의 영역이지 어댑터 답 정책의 변화가 아니므로 어댑터 변경을 수반하지 않는다 — M4가 이 노출면을 관측 기록으로 남긴다.

## §E. 비기능 요구

- **무네트워크·무바이너리 시험**: 갱신 후 정합성 시험은 vendored fixture만으로 실행된다(codex 바이너리·네트워크 불요 — 기존 스키마 가드 규율 계승).
- **무자격증 유출**: 포획한 `codex login status` 샘플은 토큰·계정 식별자를 제거한 정화본만 vendoring한다(REQ-CL-008 정화 규율과 상위 리포 #49384 계측 취지 준수).
- **크로스플랫폼**: 손대는 시험 파일은 `GOOS=windows GOARCH=amd64 go build ./...`를 유지한다.
- **시험 런타임**: 관리 TUI 계열의 기존 워치독 상한(tuiWatchdog 등)을 넓히지 않는다.

## §F. 성공 판정

acceptance.md §D AC 행렬과 증거 장부(RED-now 셀)가 판정면이다. 요약: fixture 교체 3건(AC-CONF-001~003)은 차단 기준, auth·resume 내성 2건(AC-CONF-004~006)은 계약 유지·신규 검증(AC-CONF-006의 차단 채용은 M3 E8 시점), 불변 표면 2건(AC-CONF-007~008)은 회귀 가드다. 전체 CLI 스위트는 카드 워크트리에서 구조적으로 적색이므로 레인-로컬 판정은 명명된 시험 계열로 하고, **전체 스위트 등판면은 이 카드의 main 통합 PR과 그 병합 헤드에서 실행된 CI run**이다 — develop push에는 워크플로 트리거가 없다(`.github/workflows/ci.yml`의 트리거는 `on.push.branches: [main]` / `on.pull_request.branches: [main]`뿐, 실측 2026-10-09).
