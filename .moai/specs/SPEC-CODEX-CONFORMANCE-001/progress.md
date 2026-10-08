# SPEC-CODEX-CONFORMANCE-001 — progress.md

status: in-progress

## §E.1 Plan-phase Audit-Ready Signal

plan_status: audit-ready
plan_complete_at: 2026-10-09
plan_audit: iter 1 FAIL 0.875 (MP-8 fail, blocking 6, required backend codex fail — receipt rcpt-f829a5596d422b6433546552, verdict `.moai/reports/t1607/plan-audit.md`); iter 2 repair applied — fixes 1-7(D1-D8) + lane inputs (D6 CI-trigger correction, 0.161.0 하위명령 실측 해소); iter 2 재감사는 감사자 429 사망으로 미완료; plan amendment round (2026-10-08T17:12:42Z) — 레인 실측(2026-10-08T16:56Z+) 접기: 제너레이터 출력 형상(39파일·v1/v2 분할·소비 8종 전부 존재, 이전 감사의 "이름 누락"은 ls 별칭 인공물로 반증) + vendoring 구성 결정(소비 부분집합, decision-index Q7 기본 적용); iter 2 FAIL 0.94 (MP-8 잔여 AC-CONF-005 + N2/N3/N4, receipt rcpt-2f21e144af2e5bef92bb82f6) — round-3 repairs applied (N4 005 채용 지연+기준선 원리 정정+Q4 앵커 갱신, N2 양세계 family-green 완결 경로, N3 죽은 출처 참조 제거, O2 스탬프 UTC 정정, O3 스테이징 디렉터)
artifacts: spec.md, plan.md, acceptance.md, decision-index.md, progress.md (Tier M set + decision gate on)
red_now_ledger: acceptance.md §B (LEDGER-1/2/3 — tree 81786284e; AC-CONF-006의 RED 채용은 M3 E8 시점)
open_items: decision-index Q4 (EVIDENCE-NEEDED — 0.161.0 login-status 출력; 미인증 형태 포획은 AC-CONF-005의 최소 필수 관측으로 강화, keyring 변형만 기록 갭 허용)

## §E.2 Run-phase Evidence

### M1 — fixture 재생성 + 소비자 이전 (2026-10-09, tree dce9596be 이후 본 카드 브랜치)

**생성 환경 관측 (npx -y @openai/codex@0.161.0, lane 지정 핀)**:

- `npx -y @openai/codex@0.161.0 --version` → stdout `codex-cli 0.161.0`, exit 0 (본 실행 재관측; 레인 선행 실측과 일치).
- `npx -y @openai/codex@0.161.0 app-server generate-json-schema --out /tmp/t1607-m1/schema-staging` → exit 0. 출력: 최상위 39파일 + `v1/`(2파일) + `v2/`(274파일) = 총 315파일, 4.3MB. 소비 8종 Response 스키마는 최상위에 전부 존재(레인 선행 관측과 일치; plan §A "실파일 39개"는 최상위 기준).
- **소비 8종 바이트 동일성**: `cmp`로 0.160.0 vendored 세트와 전수 대조 → 8종 전부 IDENTICAL (ApplyPatchApprovalResponse/CommandExecutionRequestApprovalResponse/DynamicToolCallResponse/ExecCommandApprovalResponse/FileChangeRequestApprovalResponse/JSONRPCError/McpServerElicitationRequestResponse/PermissionsRequestApprovalResponse). 0.160.0 → 0.161.0 응답 스키마 불변 — M1 전환이 동작 보존임을 기계가 증명.
- **REQ-CONF-004 판정 재료 — resume --help 옵션 집합 관측**: `npx -y @openai/codex@0.161.0 resume --help` → exit 0, 115행. `--remote <ADDR>`(37행)와 `--remote-auth-token-env <ENV_VAR>`(42행) **둘 다 존재** → **옵션 존재 세계** 확정. `real_help_supported` 기대값 갱신 불요(acceptance.md §C AC-CONF-003의 첫 Then이 완결 경로). 신규 캡처 본문은 커밋된 0.160.0 resume-help.txt와 `diff` exit 0 = 바이트 동일.
- vendoring: 소비 8종을 스테이징에서 이름 지정 기계 복사(내용 편집 0) + resume-help.txt 캡처본 + README 재생성(0.161.0 명기, 생성 명령, 부분집합 구성 근거) → `internal/cli/testdata/codex-0.161.0/` 8+2 구성. `codex-0.160.0/` 디렉터 삭제. decision-index Q7(소비 부분집합)·Q1(교체) 기본 적용 이행.

**RED 원문 (소비자 이전 직후·vendoring 전, tree dce9596be)**:

```text
$ go test ./internal/cli -run '^(TestManagedServerRequestPolicyMatchesCodexSchema|TestManagedCodexRemoteSupportProbe|TestManagedCodexServerRequestPolicy)$' -count=1
--- FAIL: TestManagedCodexRemoteSupportProbe (0.00s)
    --- FAIL: TestManagedCodexRemoteSupportProbe/real_help_supported (0.00s)
        managed_codex_tui_test.go:1057: open testdata/codex-0.161.0/resume-help.txt: no such file or directory
--- FAIL: TestManagedServerRequestPolicyMatchesCodexSchema (0.00s)
    managed_hardening_test.go:626: read vendored schema JSONRPCError.json: open testdata/codex-0.161.0/JSONRPCError.json: no such file or directory
FAIL
FAIL	github.com/modu-ai/moai-adk/internal/cli	1.835s
```

(acceptance.md §B가 예고한 두 실패 입력 그대로 — resume-help `os.ReadFile` 실패 + `read vendored schema` 실패. `TestManagedCodexServerRequestPolicy`는 fixture 디렉터를 직접 읽지 않아 적색 대상이 아니며, 그 0.161.0 정합은 AC-CONF-002/003의 가드가 대리한다.)

**GREEN 원문 (vendoring 후)**:

```text
$ go test ./internal/cli -run '^(TestManagedServerRequestPolicyMatchesCodexSchema|TestManagedCodexRemoteSupportProbe|TestManagedCodexServerRequestPolicy)$' -count=1
ok  	github.com/modu-ai/moai-adk/internal/cli	2.100s
```

**M1 폐쇄 관측 (두 세계 공통 요구 — family green)**:

```text
$ go test ./internal/cli -run '^TestManagedCodexTUIPreconditionsAndFallback$' -count=1
ok  	github.com/modu-ai/moai-adk/internal/cli	13.329s
$ grep -rn "codex-0.160.0" internal/cli | wc -l
       0
```

(프로브는 실제 새 help 텍스트를 입력으로 실행됐다 — fake의 `tuiFakeHelpFileEnv`가 `tuiResumeHelpFixture` 절대경로를 읽고, 그 fixture가 이제 0.161.0 캡처이다. 합성 텍스트만의 녹색 아님.)

**빌드**: `go build ./...` exit 0 / `GOOS=windows GOARCH=amd64 go build ./...` exit 0.

**AC-CONF-005 예비 관측 (M2 소관이지만 동일 생성 환경에서 포획 — REQ-CONF-007 최소 필수 관측)**:

- 명령: `CODEX_HOME=<빈 디렉터> npx -y @openai/codex@0.161.0 login status` (미인증 상태 보장 — 격리 CODEX_HOME)
- verbatim 출력: **stderr에 `Not logged in` 한 줄**, stdout 빈 송출, **exit 1**
- 바이너리 버전: codex-cli 0.161.0
- 판정 재료: 이 출력은 stage 2 전체라인 문법 `logged in using (chatgpt|api key)`와 불일치 → `parseCodexAuthLine`은 `codexAuthUnknown`(갭)이어야 하며, "미인증" 판정을 내선 안 된다 — REQ-CONF-006 계약과 정합. M2가 이 포획본을 vendored 샘플로 삼아 분류 시험으로 봉인한다.

### M2 — auth/keyring 정합 (2026-10-09, tree 0d42ec4f3 이후)

**AC-CONF-004 부재-하강 명시 시험 확인 (기존 시험 재확인 — 추가 불요)**:

- `TestClassifyCodexAuth_LadderIntegration`(`internal/cli/codex_auth_ladder_test.go:523`)의 서브시험 "no auth.json + stderr-only probe"(:539)과 "no auth.json + stdout-only probe"(:553)이 **부재 auth.json → stage 2 하강**을 명시 단언한다(`stub.calls == 1` — 프로브가 정확히 1회 호출). `TestClassifyCodexAuth_RejectedAuthFileFallsBackToProbe`(:568)가 존재-기각 3형(빈 토큰·미지 모드·파스 실패)의 하강을, `TestClassifyCodexAuth_UnreadableProbeIsAGap`(:595)가 4축 갭(양 스트림 빈송출·러너 오류·비영exit 무문법·파스 실패+무음 프로브)을 각각 담는다. REQ-CONF-006의 계약이 이미 시험으로 봉인돼 있어 신규 RED/GREEN 쌍 불요 — plan §F M2의 "없으면 추가" 조건 불발.

**AC-CONF-005 봉인 — vendored 샘플 + 분류 시험 (TDD RED/GREEN)**:

- 신규 시험 `TestClassifyCodexAuth_Codex0161CapturedOutputIsAGap`(`codex_auth_ladder_test.go` 말미): vendored 포획본(`testdata/codex-0.161.0-auth/login-status-not-logged-in.txt`, `cp` 바이트 동일 복사 + README에 캡처 출처 기록)을 바이트 단정한 뒤 ① 순수 파서 `combineCodexStreams(nil, raw)` → `parseCodexAuthLine(combined, 1)` ② 전체 사다리(부재 auth.json + 포획 스트림을 그대로 돌려주는 러너 스텁) 두 경로 모두 `codexAuthUnknown`을 단언.
- **RED 원문 (vendoring 전)**:

```text
$ go test ./internal/cli -run '^TestClassifyCodexAuth_Codex0161CapturedOutputIsAGap$' -count=1
--- FAIL: TestClassifyCodexAuth_Codex0161CapturedOutputIsAGap (0.00s)
    codex_auth_ladder_test.go:644: read vendored 0.161.0 login-status sample: open testdata/codex-0.161.0-auth/login-status-not-logged-in.txt: no such file or directory
FAIL
FAIL	github.com/modu-ai/moai-adk/internal/cli	0.949s
```

- **GREEN 원문 (vendoring 후)**:

```text
$ go test ./internal/cli -run '^TestClassifyCodexAuth_Codex0161CapturedOutputIsAGap$' -count=1 -v
--- PASS: TestClassifyCodexAuth_Codex0161CapturedOutputIsAGap (0.00s)
    --- PASS: TestClassifyCodexAuth_Codex0161CapturedOutputIsAGap/pure_parser_on_combined_capture (0.00s)
    --- PASS: TestClassifyCodexAuth_Codex0161CapturedOutputIsAGap/full_ladder_on_captured_streams (0.00s)
ok  	github.com/modu-ai/moai-adk/internal/cli	0.880s
```

- 사다리 계열 전체 재실행: `go test ./internal/cli -run '^(TestClassifyCodexAuth|TestParseCodexAuthLine|TestCodexLoginStatusRunner|TestCombineCodexStreams|TestReadCodexAuthFile|TestCodexTokenSet|TestCodexAuthFileTypes|TestCodexAuthLadder)' -count=1` → `ok ... 1.780s`. `go test ./internal/cli -run 'TestMcpCodexAuth|TestCodexAuth' -count=1` → `ok ... 0.902s`.

**기록 갭 (REQ-CONF-007가 허용하는 유일한 것)**: keyring 로그인 변형("Logged in using …" 실측 형태의 keyring 저장 시 출력)은 포획 불가 — 생성 환경에 자격증명이 없어 로그인 자체가 불가능. 미인증 형태 포획(최소 필수 관측)은 성립했으므로 AC-CONF-005는 자동 PASS; keyring 변형은 수동 검증 대기 항목으로 남는다(증거 없는 갭 기록이 아니라 포획 불가능성이 기록된 갭).

**결론**: stage 2 전체라인 문법은 0.161.0 출력(미인증 형태)에 대해 0.160 세대와 동일 규칙으로 분류한다 — 문법 일치 라인이 없으면 `codexAuthUnknown`(갭), 어느 경우에도 "미인증" 판정 아님. `internal/cli/mcp_codex.go` 어댑터 본문 변경 0건(PRESERVE 준수).

### M3 — resume 표면 재검증 (2026-10-09, tree 49e9c65ee 이후)

**AC-CONF-006 재생 프레임 내성 — 신규 시험 `TestManagedCodexTUIConsumesResumedThreadReplayFrames`(`managed_codex_tui_test.go` AC-CONF-006 절)**:

- **프레임 근거 (REQ-CONF-005의 근거 절 이행)**: 재생 프레임 필드는 **0.161.0 바이너리가 자체 생성한 프로토콜 스키마에서 유도**했다 — M1 생성 산출의 `v2/ItemStartedNotification.json`·`v2/ItemCompletedNotification.json`·`v2/TurnCompletedNotification.json`·`v2/ThreadTokenUsageUpdatedNotification.json`·`v2/TurnDiffUpdatedNotification.json`(npx @openai/codex@0.161.0 generate-json-schema 출력; rust-v0.161.0 태그=커밋 979011409de0a60b52f179721948e65531d26144의 `codex-rs/app-server-protocol/src/protocol/v2/item.rs`·`turn.rs`·`thread.rs` 정의와 교차 확인). fake 서버 자체 생성 응답은 근거로 사용하지 않았다(자기-생성 순환 배제 — mutant-probe 채용 규칙). 재생 버스트 5프레임: agentMessage ThreadItem(phase/memoryCitation/delivery/questions 포함)의 item/started, commandExecution ThreadItem(commandActions/durationMs/pluginId 등 전 필드)의 item/completed, 0.161.0 Turn 형상(id/items/status 필수)을 실은 **외래 턴**의 turn/completed, 어댑터가 모르는 메서드 thread/tokenUsage/updated, turn/diff/updated.
- **양성 대조 (AC 요구)**: 재생 버스트 소비 직후 ① 서버 요청이 여전히 답변됨(read loop 생존 증명) ② `claims.release`로 소유자 자신의 턴(fake-turn-2)이 재생 프레임 **뒤에서** 완료됨 ③ 세션 로그에 `Factory turn failed` 부재 ④ 드라이버가 bye/exit로 오류 없이 종료.
- **실패 관측 가능성 — canary 서브시험**: 동일 버스트 + `tuiFakeStatusesEnv=completed,failed`로 소유자 턴을 스크립트 실패시키면 `Factory turn failed`가 **관측된다** — 위 부재 단언이 살아 있는 신호에 묶여 있음을 증명(verification-completeness §1.1의 관측된 실패 축).
- **mutant probe (채용 전 관찰된 실패 — 본 AC의 E8 RED 원문)**: read loop의 미모델 프레임 폐기 지점(`managed_codex_factory.go` `read()`의 `case event.Method != "" || !event.hasID(): continue`)을 일시 변이(`continue` → `return` — 디코드 실패형 접속 종료 시뮬레이션)한 뒤 시험 실행 → **적색 관측**:

```text
$ go test ./internal/cli -run '^TestManagedCodexTUIConsumesResumedThreadReplayFrames$' -count=1
--- FAIL: TestManagedCodexTUIConsumesResumedThreadReplayFrames (22.00s)
    --- FAIL: TestManagedCodexTUIConsumesResumedThreadReplayFrames/replay_frames_keep_the_owner_turn_alive (11.48s)
        managed_codex_tui_test.go:1484: a server request after the replay burst was not answered: the reader died on the replay frames
        managed_codex_tui_test.go:1490: the owner's own turn after the replay burst never completed:
            [...fake app-server 로그 발췌...]
FAIL
```

  변이는 즉시 원복됐고 원복 확인: `git diff --stat internal/cli/managed_codex_factory.go` 빈 출력 + 재실행 GREEN:

```text
$ go test ./internal/cli -run '^TestManagedCodexTUIConsumesResumedThreadReplayFrames$' -count=1
ok  	github.com/modu-ai/moai-adk/internal/cli	2.589s
```

  (정상 트리에서의 이 시험은 착시적 green이 아니다 — canary가 실패 관측을, mutant probe가 결함 클래스 검출을 각각 실측했다. 내성 그 자체는 구조적 lenient 디코드가 이미 보유한 계약이라 구현 변경 0건.)

**codex_role_fingerprint.go rollout 소비 재확인 (plan §F M3)**:

- 측정: rust-v0.161.0(커밋 979011409de0a60b52f179721948e65531d26144)의 `codex-rs/history/src/lib.rs` RolloutItem 열거와 `codex-rs/history/src/rollout_payload.rs` RolloutItemWire 직렬화(`#[serde(tag = "type", rename_all = "snake_case")]` — `session_meta`/`response_item`/`turn_context` 등 태그 + `payload` 래퍼)를 fetch해 어댑터의 `codexRolloutLine`(`{type, ordinal, payload}`) 소비 형상과 대조.
- 판정: **라인 형상 불변** — type 판별자(snake_case)·payload·ordinal 필드 모두 0.161.0에서 유지. `agent_role` 필드도 세션 소스 표면에 존재 유지(상위 자체 테스트 픽스처에서 `agent_role: None` 필드 관측). `codex_role_fingerprint.go` 소비는 failure-soft(파스 실패 → 라벨 없음, 오류 아님 — REQ-RLP-015)라 잔여 리스크가 낮다. **어댑터 변경 0건 — 비목표/후속 없음.**
- 갭: `session_meta` payload 내 `subagent.thread_spawn.agent_role` 중첩 경로의 바이트 수준 재확인은 상위 프로토콜 크레이트 추가 fetch 없이는 미수행 — 라인 형상과 agent_role 존재로 뒷받침되는 결정이며, failure-soft 소비가 형상 변화를 "오류"가 아닌 "라벨 없음"으로 강하므로 본 카드 범위의 재검증 결론에는 영향 없다.




## §E.3 Run-phase Audit-Ready Signal

_<pending run-phase>_

## §E.4 Sync-phase Audit-Ready Signal

_<pending sync-phase>_

## §F Phase 4 Mode Selection

- Input parameters: tier M; scope ~6 files (fixture dir + 2 test consumers + 1 ops doc + SPEC artifacts); domains 1 (Go CLI adapter + its tests); language mix Go + machine-generated fixtures; concurrency benefit LOW (coding-heavy, M1 fixture regeneration gates M2/M3); Agent Teams prereqs n/a (explicit-request-only).
- Mode evaluation: direct — not selected (semantic multi-milestone work); serial — selected; fanout — not selected (not research-heavy, single domain); sweep — not selected (semantic work, not mechanical-uniform bulk); agent-team — not selected (no operator request).
- Decision: serial
- Justification: single-domain coding-heavy conformance work with a strict milestone dependency chain — sequential manager-develop delegation with per-milestone commits is the safe default per the coding-task parallelism caveat; no other mode's selection criteria are unambiguously met.
- Kickoff record (autonomous form): decision record: decided_by=lane-21 evidence_refs=.moai/reports/t1607/plan-audit.md (iter-4 PASS 1.0, must_pass 0, blocking 0; codex required backend pass, governing receipt rcpt-f16b809f87c02e929297a2e9 fresh on the unchanged final state; plan_artifact_hash 93f2e64f8bbb38ba93a0ec664ddfa00796321566a9300da3f7223b115f34b795 via the canonical runtime.ComputeHash recipe — name:len:NUL+bytes+NUL per planArtifactNames order — unchanged since verdict; the earlier 21aba6a7 figure was the concatenation recipe and is superseded per gate review) ladder_path=plan-to-run Kickoff autonomous form — verdict PASS + 1.0 ≥ 0.80 Tier M threshold + artifact hash unchanged + no open blocker — run-phase entry approved.
