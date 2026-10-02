---
id: SPEC-CODEX-REVIEW-ASYNC-001
title: "Claude Stop 훅 codex 리뷰 게이트를 asyncRewake 로 바꾼다 — 턴 종료를 기다리지 않고 실패할 때만 세션을 깨운다"
version: "0.1.0"
status: draft
created: 2026-10-02
updated: 2026-10-02
author: manager-spec
priority: P1
phase: "v3.2.0 target"
module: internal/cli
lifecycle: spec-anchored
tier: M
depends_on: [SPEC-CODEX-REVIEW-OWNERSHIP-001]
related_specs: [SPEC-CODEX-REVIEW-OWNERSHIP-001, SPEC-CODEX-GATE-SCOPE-001, SPEC-MOAI-MCP-SERVER-001, SPEC-DUAL-HARNESS-HOOK-PARITY-001, SPEC-WORKTREE-STATE-ROOT-001]
tags: "codex, review-gate, stop-hook, asyncRewake, lock, stale, wake-loop, t1422, t1425"
---

# SPEC-CODEX-REVIEW-ASYNC-001 — Claude Stop 훅 codex 리뷰 게이트의 asyncRewake 전환

카드: **t1422** (운영자 지시 2026-10-02, 리더 경유). 같은 카드의 형제 SPEC `SPEC-CODEX-REVIEW-OWNERSHIP-001`(소유권 재배치)이 **먼저 착지**하고 이 SPEC 은 그 위에 얹는다(`depends_on`). 형제가 정한 것 — `tree_scope` 키, 자기 리뷰 도구, 카드 t1426 분리 — 은 다시 논의하지 않는다.

## HISTORY

- 2026-10-02 · v0.1.0 · manager-spec · 최초 작성. 측정 원천: 본 트리 `.moai/worktrees/t1422`(브랜치 `WT-codex-review-lane-scope`, HEAD `3ae43ed8e78ffa673ca238227df6ca7202c1ce70`)의 코드 좌표 직접 판독(§A)과 공식 훅 레퍼런스(https://code.claude.com/docs/en/hooks)의 명령 훅 필드 표·종료 코드 2 표 판독. 결정(Jev `jev-1.13.0`, 운영자 위임, 2026-10-02): 중복 실행 방지 = 트리별 락·보유 중이면 건너뜀·죽은 보유자 인수(0.83), 오래된 결과 = 표지를 붙여 전달(0.99, `label_stale`), 구조 = 형제 SPEC 분리(0.93). progress.md「Decision Log」에 옮겼다. plan-audit 1회차 교훈(sibling 보고서 D1-D17)을 처음부터 적용했다: RED 이유 열, 변이 점검, 보수적 개수 규칙, 파리티 편집 마일스톤 배정, 관측하지 않은 외부 동작을 사실로 쓰지 않기.
- 2026-10-02 · v0.1.0 · manager-spec · 오픈 결정 O-A~O-J 반영(제자리 갱신, 버전 불변). 상한(구 REQ-CRA-005 의 일부)은 오케스트레이터 재정 `cap3_notify_once_at_cap`(Jev 판정 `cap3_silent` 0.21 을 대체, 잠정)으로 바꾸고 마지막 알림에 고정 문구 `이후 알림 없음, 상태는 미해결` 을 넣었다. 카드 칸은 카드 워크트리일 때만 디렉터리 이름(REQ-CRA-001), 오래됨은 상태 키 전체(REQ-CRA-006, 리더 수용), 알려진 상태 건너뜀은 pass 에도 적용되는 현행 변경(REQ-CRA-005). 결정 원문은 progress.md「Decision Log」.
- 2026-10-02 · v0.1.0 · manager-spec · plan-audit 1회차(FAIL 0.74, `.moai/reports/t1422/async-plan-audit.md`, 감사 트리 `a0d801409`) 반영(제자리 갱신, 버전 불변). D1: 상한은 findings 지문 없이 **트리의 연속 실패 깨움 횟수**로 센다 — 지문은 LLM 산문 제목에 의존해 회피되고(재서술) 충돌하며(`fail`+findings 0건), 고정 문구 "이후 알림 없음"과도 어긋난다(REQ-CRA-014·015 로 분리). D2: 실패 기록에 세션 id 와 `delivered` 를 두고 stderr 기록 성공 뒤에만 delivered 로 바꾼다(REQ-CRA-005). D3: 종료 코드 2 는 Go 패닉도 내므로 래퍼가 **sentinel 줄**이 있는 종료 코드 2 에서만 깨운다(REQ-CRA-003). D4: 운영 중 stderr 의 첫 줄은 스코프 로그 JSON 행이므로 AC-001 의 "첫 줄" 주장을 sentinel 줄 기준으로 고치고 래퍼가 요약만 추린다. 알려진 상태 키는 다섯 필드로 넓혔고, 락·기록·로그는 작업 트리가 아니라 git 디렉터리에 둔다(REQ-CRA-004·009·013 — `verify.Key` 가 `.gitignore` 에 의존하지 않도록). 보유자는 시간 제한을 가진다. 크기 상한에 숫자를 붙였다.

---

## §0 지배 원칙 [HARD]

> **게이트는 턴을 기다리게 하지 않는다. 그리고 한 번의 리뷰가 낳은 것만 한 번 전달한다 — 중복 실행도, 반복 호출도, 표지 없는 오래된 결과도 없다.**

현행 Claude Stop 게이트는 동기 훅이다. 턴 종료가 리뷰(최대 900s)를 기다리고(§A.1), 같은 트리에서 Stop 이 몰리면 리뷰가 그만큼 겹쳐 돈다(§A.4). `asyncRewake` 는 훅을 백그라운드로 돌리고 종료 코드 2 에서 Claude 를 깨운다고 공식 문서가 적는다(§A.2). 그러나 **깨어나는지·타임아웃이 어떻게 적용되는지**는 문서가 적지 않아(§A.3) 이 SPEC 의 깨움 주장은 가설이고, 레지스트리 변경은 그 관측이 끝난 뒤 마지막에 한다. 이 SPEC 은 리뷰 내용·판정 기준·소유권(형제 SPEC)·Codex 경로를 건드리지 않는다.

---

## §A 배경 (측정과 출처)

### A.0 측정 원천과 핀

좌표는 본 트리 기준 2026-10-02 판독값이다. 핀: HEAD `a0d8014096f894c3a45577482b0eb92892716e04`(`git rev-parse HEAD`; 최초 작성 핀 `3ae43ed8e` 과 `git diff --stat 3ae43ed8e a0d801409 -- internal .claude cmd .moai/config` 가 비어 있어 인용 코드는 동일하다). 설치된 `moai` 는 v3.2.0-rc.24, 이 세션의 MCP 서버 안내문은 rc.23(`d194083fb`)을 알렸다. Claude Code 설치 버전은 코디네이터 제공 2.1.287 이며 이 레인이 직접 확인하지 않았다 — **Gap**. 이 레인은 라이브 Claude Code 세션도 라이브 codex 도 실행하지 않았다.

### A.1 현행 게이트는 동기 훅이다

- 등록: `.claude/settings.json:190-194`(codex 래퍼 항목; 스크립트 경로 `:191`, `"timeout": 900` `:192`, `"type": "command"` `:193`; `async`·`asyncRewake` 키 없음 — `grep -c asyncRewake .claude/settings.json` → 0)와 multi 래퍼 `:196-200`(같은 형태). 등록 명령 형태는 `"command": "bash", "args": ["-c", "[ -f \"$0\" ] && exec bash \"$0\"; …(hook missing 기록) exit 0", "${CLAUDE_PROJECT_DIR}/.claude/hooks/moai/handle-codex-review-gate.sh"]` 이다 — 스크립트가 있으면 `exec` 하므로 스크립트의 종료 상태가 그대로 전달된다. 템플릿 `internal/template/templates/.claude/settings.json.tmpl:191` 도 같은 항목이다(`grep -c asyncRewake` → 0). 같은 파일에서 `async: true` 는 다른 훅에 쓰이고 있다(`grep -c '"async": true' .claude/settings.json` → 11, 키 형태 양성 대조).
- 래퍼 `internal/template/templates/.claude/hooks/moai/handle-codex-review-gate.sh`: 순수 셸 off 스위치 후(`:35,36,60`), 마지막에 `printf … | "$MOAI_BIN" hook codex-review-gate` 를 돌리고 **항상 `exit 0`** 한다(`:84-86` — "the BLOCK decision rides the JSON Decision field on stdout"). 종료 코드를 전파하는 `exit $…` 줄은 없다(`grep -c 'exit \$' …` → 0).
- 핸들러 `internal/cli/codex_review_gate.go:69-119` `HandleCodexReviewGate`: 비활성(:71) → `stop_hook_active`(:74) → 스코프 해상(:81) → 셀프게이트(:83) → codex 조회(:87) → **동기 리뷰**(:103, 상한 `config.DefaultCodexReviewGateTimeout` = 900s, `internal/config/defaults.go:557`) → fail 이면 `HookOutput{Decision: block, Reason}`(:112-117). 실행기 `runCodexReviewGate`(:192-211)가 그 출력을 stdout JSON 으로 내보낸다(`return emitHookOutput(` 3곳 :197,:205,:210 — `grep -c 'return emitHookOutput(' internal/cli/codex_review_gate.go` → 3). 락·상태 파일·재전달 억제 코드는 없다(`grep -rl 'AcquireReviewLock' internal/cli --include='*.go' --exclude='*_test.go'` → 출력 없음; `grep -c exitCodeError internal/cli/codex_review_gate.go` → 0).
- **운영 중 stderr 는 이미 JSON 행으로 시작한다(D4).** 핸들러가 스코프 해상 직후 `reviewGateScopeLogger` → `logReviewScope`(`internal/cli/codex_review_scope.go:340-356`)를 부르고, 그 함수는 `fmt.Fprintln(os.Stderr, …)` 로 프로세스 stderr 에 한 줄을 쓴다(리뷰 전). 인-프로세스 cobra 하네스의 버퍼는 이 줄을 보지 못한다. 실측: 환경 세척 후 `go test ./internal/cli/ -run '^TestCodexReviewGate_TreeScopeRequestShapeUnchanged$' -count=1 -v` 출력에 `{"basis":"no card branch: main","branch":"main","gate":"codex-review-gate","scope":"tree"}` 행이 `=== RUN` 다음에 그대로 나온다.
- 이 트리의 등록 시험: `internal/template/review_gate_registration_test.go`(`reviewGateWrappers` 맵이 두 래퍼의 `timeout` 900 을 고정)과 `internal/hook/review_gate_selfgate_test.go`(비활성 시 바이너리 미호출). 현재 초록: `TestReviewGatesRegisteredInRepoSettings`·`…InTemplateSettings`·`TestReviewGateWrappers_DisabledCostsZeroColdStarts`·`…EnabledReachesBinary` 모두 `--- PASS`(환경 세척 후 실행).

### A.2 공식 문서에서 확인된 사실 (이 레인이 훅 레퍼런스를 직접 읽음)

- 명령 훅 필드 표: `async` — "If `true`, runs in the background without blocking." `asyncRewake` — "If `true`, runs in the background and wakes Claude on exit code 2. The hook's stderr, or stdout if stderr is empty, is shown to Claude as a system reminder so it can react to a long-running background failure." `timeout` — "Seconds before canceling. Claude Code doesn't enforce it on a command hook you run with `async: true`." (기본값 command 600). 오케스트레이터가 검증해 전달한 문장: "Apart from a command hook you run with `async: true`, Claude Code cancels a … hook that reaches its timeout, discarding the hook's output" — 이는 `asyncRewake` 훅에 대한 **문서상 기본 읽기**가 "취소·출력 폐기"임을 뜻한다(이 레인이 받은 페이지 요약에는 없던 문장이며 오케스트레이터 제공이다. `asyncRewake` 에서의 실제 동작은 E-2 로 열어 둔다).
- 종료 코드 2 표: `Stop` — "Prevents Claude from stopping, continues the conversation" (**동기** 등록의 의미). 같은 문서가 종료 코드 2 에서도 "Claude Code still reads any valid JSON output on stdout" 라고 적는다.
- 위 페이지의 "Run hooks in the background" 절은 이 레인이 받은 내용에서 잘려 있었고(`InstructionsLoaded` 결정 제어 절 중간에서 끊김) 훅 가이드 페이지에도 없다. `stop_hook_active`·세션 시작 시점 스냅숏은 두 페이지에서 찾지 못했다(요약 도구 결과 — 문서에 없다는 증명이 아니다). 훅 출력 크기 상한(긴 stderr 의 잘림)도 확인하지 못했다 — **Gap**(plan.md §B.2 의 크기 상한은 보수적 선택이지 문서 근거가 아니다).

### A.3 EVIDENCE-NEEDED — 문서가 적지 않은 것 (사실로 올리지 않는다)

| ID | 미관측 사항 |
|---|---|
| E-1 (a) | **Stop 훅이 턴이 끝난 뒤 끝났을 때 idle 세션에서 새 턴이 실제로 시작되는가** |
| E-2 (b) | **`asyncRewake` 훅에 `timeout` 이 어떻게 적용되는가**(취소? 출력 폐기? 위 `timeout` 행은 `async: true` 만 명시한다) |
| E-3 (c) | **동시성·중복 제거, 훅 실행 중 새 턴이 시작될 때, 훅 실행 중 세션이 종료될 때의 동작**(훅 프로세스가 죽는지 고아가 되는지 — 락 보유자 사망 시나리오에 직결) |
| E-4 (d) | **JSON `decision: block` 출력이 비동기 훅에서 무시되는가**(그러면 옛 신호를 병행할 수 없다) — 같은 프로브에서 **종료 코드 0 훅의 stderr 가 Claude 에 보이는지**도 관측한다 |
| E-5 (e) | **실행 중 세션에서 훅 설정(`.claude/settings.json`) 편집이 언제 반영되는가**(시작 시 스냅숏인가) — 레지스트리 변경이 같은 트리의 진행 중 세션에 미치는 영향 |

이 SPEC 은 위 다섯을 **M1 라이브 프로브**로 관측해 `.moai/reports/t1422/async-probe/` 에 원문을 남기고, 결과가 이후 마일스톤을 게이트한다(§D, plan.md §I). 프로브 하나가 헤드리스로 돌릴 수 없는 대화형 세션을 요구하면 그 사실과 실행 주체(리더/운영자)를 progress.md 에 적는다. 관측 전에는 "깨운다"를 사실로 쓰지 않는다 — 이 문서의 모든 깨움 서술은 가설이다.

### A.4 중복 실행 위험과 같은 부류의 선례

핸들러에 락이 없으므로(A.1) 같은 트리에서 Stop 이 N 번 몰리면 리뷰가 N 번 돌 수 있다 — 코드 판독이며 N 중 실행 자체는 관측하지 않았다. 같은 부류(훅 호출마다 새 프로세스가 같은 일을 겹쳐 수행)의 선례로 카드 t1425(harness 관측 로그 retention 정리 폭주; 카드 본문 — **리더 제공**, 이 레인이 재현하지 않음: load 840, 훅 프로세스 738+282 개 적체)가 있다. 이 SPEC 이 막는 것은 codex 리뷰 게이트 한 곳이고 t1425 의 수리는 그 카드 소관이다.

### A.5 재사용할 락·상태 원시 기능과 상태 키의 실제 구성

- `internal/cli/gate_lock.go`(gate-run 락): 유닉스 `flock(LOCK_EX|LOCK_NB)`(`gate_lock_unix.go:38-61`) — 커널이 프로세스 종료 때 해제하므로 죽은 보유자의 락은 막지 않는다(같은 파일 머리 주석), 윈도우는 원자적 생성+`ClearStaleGateLock`(`gate_lock_windows.go:81,128`), 보유자 신원 `GateLockOwner{PID, CreatedAt}`(`gate_lock.go`), 경합 센티넬 `ErrGateLockHeld`, 생존 판정 `kanban.FactoryProcessAlive`(`internal/kanban/factory_alive_unix.go:22`). 하위 함수 `acquireGateLockImpl(lockPath)` 는 경로 매개변수형이고 윗층(`AcquireGateLock`·`ClearStaleGateLock`)이 고정 이름·고정 경로를 쓴다 — **경로를 매개변수로 만드는 소규모 일반화**가 필요하다.
- 검토했으나 쓰지 않는 것: `internal/lockfile`(Lock/Unlock 만, 보유자 신원 없음), `internal/verify/claim_lock.go`(O_EXCL+5분 TTL — 15분급 리뷰에 맞지 않는 만료), 슬롯 임대(`moai slot`, 운영자 수준의 선언형 임대 — 훅 경로 원시 기능으로 검토하지 않음 — **Gap**).
- 트리 정규화: `internal/auditreceipt/treeroot.go:22` `TreeRootFromCWD`(git toplevel+심볼릭 링크 해소; `CLAUDE_PROJECT_DIR` 를 보지 않는 이유가 파일 주석에 있다).
- **상태 키가 무엇을 보는가 — 락·기록·로그 파일을 어디에 두어야 하는가(D6).** 리뷰 상태 키의 트리 부분은 `verify.Key`(`internal/verify/key.go:40-100`)가 만든다: `git rev-parse HEAD` + `git status --porcelain=v2` 다이제스트 + `git diff HEAD` 해시 + `git ls-files --others --exclude-standard` 가 가리키는 **비무시 비추적 파일의 경로·내용** 해시다. 즉 키에서 빠지는 것은 "git 이 무시하는 파일"이지 `reviewGateRuntimePrefixes`(`codex_review_gate.go:36-43`)가 아니다 — 그 접두 목록은 `hasReviewableChanges`(리뷰할 변경이 있는가)만 정한다. 이 저장소에서는 `.gitignore:398 .moai/state/`·`:397 .moai/logs/` 가 파일을 숨긴다(`git check-ignore -v .moai/state/x .moai/logs/x` 출력). 그러나 그 줄이 없는 저장소(오래된 사용자 프로젝트, 수락 픽스처의 "실제 git 저장소")에서는 락 파일(내용 = 보유자 PID, 실행마다 다름)·재전달 기록·로그가 비추적 내용이 되어 **리뷰 직전·직후의 키가 달라지고**, 알려진 상태 건너뜀이 영원히 발화하지 않으며, `verify.Key` 를 쓰는 Codex 경로의 키까지 바꾼다(REQ-CRA-009 위반). 그러므로 세 파일은 작업 트리가 아니라 **트리의 git 디렉터리**(`git rev-parse --absolute-git-dir` — 연결 워크트리에서는 워크트리별 `…/.git/worktrees/<name>`; 이 트리에서 `/Users/goos/MoAI/moai-adk-go/.git/worktrees/t1422`) 아래 `moai/` 에 둔다. git 디렉터리는 비추적 내용도 diff 도 아니고 워크트리별이다.
- 리뷰 상태 키: `codexReviewReceiptStateForScope`(`internal/cli/codex_review_receipt.go:86`)가 `verify.ReceiptState` 다섯 필드 — `Head`·`TreeDigest`·`ConfigDigest`·`Command`·`ToolVersion`(`reviewReceiptStateFrom`, `codex_review_receipt.go:60-79`; `internal/verify/receipt.go:37-43`) — 를 준다. Codex 경로의 `verify.CheckReceipt` 는 다섯을 모두 비교한다. 이 SPEC 의 알려진 상태 키도 다섯 필드 전부다(codex 바이너리나 리뷰 설정이 바뀌면 같은 트리 상태라도 새 상태다).
- findings 의 실제 모양(D1): `codexFindingsOf`(`internal/cli/mcp_codex.go:1757-1768`)는 `Finding{Severity: sev, Title: msg, Body: msg}` 를 만든다 — `Title` 은 LLM 불릿 한 줄 전체이고 `File`/`Line` 은 산문에 `path:line` 이 있을 때만 채워진다. `flagVerdictFindingsContradiction`(`mcp_codex.go:1795-1799`)은 "verdict `fail` 인데 findings 0건"을 정식으로 다루는 모양이다(내용은 `Summary` 에만 있다). 따라서 findings 에서 만든 지문은 재서술에 약하고(같은 문제도 제목이 달라짐) 0건 fail 에서 충돌한다 — 이 SPEC 의 상한은 지문을 쓰지 않는다(REQ-CRA-014).
- 종료 코드: `exitCodeError{code, msg}`(`internal/cli/constitution.go:305-313`, `ExitCode()`), `cmd/moai/main.go` 의 ExitCoder 처리. `fang.go` 의 `moaiErrorHandler` 는 ExitCoder 오류를 추가로 출력하지 않는다(감사가 관측).
- 와이어링 비교: `internal/template/hook_entries.go:17-30` `HookEntry` 가 `(event, matcher, script, if, timeout, async)` 만 키로 삼는다 — `asyncRewake` 는 키가 아니다(`grep -c AsyncRewake internal/template/hook_entries.go` → 0, 양성 대조 `grep -c Async` → 5). `doctor_hook_wiring.go:79` 의 `DiffHookEntries` 가 이를 쓴다.

### A.6 Codex 경로와 multi 게이트는 이 SPEC 밖이다 (근거)

- Codex Stop 체인 멤버 6(`codex_stop_chain.go:613-669`)은 리뷰를 실행하지 않고 영수증을 읽는다. 영수증 생산자 `produceCodexReviewReceipt`(`codex_review_receipt.go:126-174`)와 `moai verify codex-review` 는 사용자가 명시 실행하는 경로다. 두 경로가 Claude 경로와 공유하는 것은 해상기(`reviewScopeResolver`)·요청 조립기(`reviewRequestParams`)·RPC 드라이버(`runCodexReviewRPC`)·상태 계산(`codexReviewReceiptStateForScope`)이고, 이 SPEC 이 넣는 **락·재전달 상태 파일은 Claude Stop 경로만 읽고 쓴다.**
- multi 게이트(`internal/cli/multi_review_gate.go:75`)는 리뷰를 실행하지 않고 저장된 수렴 결과를 읽는다(헤더 `:17-23`). 대기 시간 문제가 없고 래퍼 등록은 동기·900s 그대로 둔다 — 범위 밖.

### A.7 종료 코드 2 는 판정 신호가 아니다 — 모호성(D3)

`internal/cli/CLAUDE.md` 는 종료 코드를 "`0` 성공, `1` 사용자 오류, `2` 시스템 오류"로 문서화하고(`:13`), Go 런타임은 회수되지 않은 패닉에도 종료 상태 2 를 낸다. 래퍼가 핸들러의 모든 2 를 깨움으로 전달하면 락·상태·지문 코드의 패닉이 고루틴 트레이스와 함께 세션을 깨운다 — 운영자 요구("오류는 조용한 종료 코드 0")와 반대다. 그러므로 깨움은 **판정 실패의 sentinel 줄**(`codex review gate: FAIL`)을 가진 종료 코드 2 에만 한다(REQ-CRA-003).

---

## §B 요구사항 (GEARS)

개수 규칙: `### REQ-` 제목 수 = 15. 묶음: **A. 신호**(REQ-CRA-001~003), **B. 중복 방지와 재전달 억제**(004, 005, 014, 015), **C. 오래된 결과**(006), **D. 등록과 타임아웃**(007~008), **E. 경계**(009~010), **F. 관측 게이트와 안전**(011~013). 번호 014·015 는 plan-audit 1회차에서 구 REQ-CRA-005 를 셋으로 가르며 뒤에 붙였다(기존 번호를 옮기지 않기 위해). 깨움에 관한 서술(REQ-CRA-001 의 "깨운다", 007)은 E-1 관측 전까지 가설이며 REQ-CRA-011 이 그 취급을 고정한다.

### A. 신호 (운영자 요구 1)

### REQ-CRA-001 — 실패는 종료 코드 2 와 stderr 요약이다 (When)

**When** the review of a tree completes with a failing verdict and the gate has not suppressed the delivery (REQ-CRA-005, REQ-CRA-014), the Claude Stop-hook codex review gate shall write a summary to standard error whose first line is exactly `codex review gate: FAIL` and which names the card, the branch, the reviewed HEAD, the verdict summary, and the findings list within the size bounds of the plan, write nothing to standard output, and exit with code 2; the card field shall hold the worktree directory name when the tree is a `WT-` card worktree and shall be blank otherwise.

옛 신호(종료 코드 0 + stdout JSON `decision: block`)는 이 경로에서 더는 내지 않는다. 요약의 첫 줄(sentinel)은 요약 블록의 첫 줄이다 — 운영 중 프로세스 stderr 에는 그 앞에 스코프 로그 JSON 행이 한 줄 있을 수 있고(A.1) 래퍼가 sentinel 줄부터 추려 전달한다(REQ-CRA-003). 코드에는 카드 id 의 출처가 없어 레인 규약(`gitflow-lane-protocol.md` §1: 워크트리 디렉터리가 카드 id 를 유지)에 기댄 워크트리 디렉터리 이름을 카드 칸에 적고, 카드 워크트리가 아닌 트리(예: 리더가 앉은 primary)에서는 칸을 비운다(Q15 확정). **크기 상한(보수적 선택 — 훅 출력 크기 상한은 미확인 Gap):** 나열하는 findings 최대 10건, 각 finding 줄 최대 300자, `summary` 최대 1500자, 요약 전체 최대 8000바이트이며 초과분은 개수(`(+k more)`)나 `[truncated]` 표지로만 적는다. sentinel 줄·stale 표지·마지막 알림 줄은 자르지 않는다. 0건 findings 인 fail(`flagVerdictFindingsContradiction` 모양)도 같은 요약을 낸다(`findings (0)`, 내용은 `summary` 에).

### REQ-CRA-002 — 그 밖의 결과는 조용히 종료한다 (When)

**When** the review passes, is inconclusive, errors, finds no reviewer, finds the gate disabled, finds `stop_hook_active`, finds nothing reviewable, reads an unparseable hook input, or is suppressed by REQ-CRA-004, REQ-CRA-005, or REQ-CRA-014, the gate shall exit with code 0, write nothing to standard output, and write no review summary.

SPEC-MOAI-MCP-SERVER-001 REQ-MCP-012 의 fail-open 은 그대로다(리뷰어 부재·오류·inconclusive 는 허용 방향). 기존 스코프 로그 행(REQ-CGS-010)은 stderr 에 한 줄 남는 현행 동작이며, 종료 코드 0 훅의 stderr 가 Claude 에 보이는지는 E-4 에서 관측한다 — 보이는 것으로 판명되면 그 행을 로그 파일로 옮긴다(plan.md §G).

### REQ-CRA-003 — 래퍼는 판정 실패의 종료 코드 2 에서만 깨운다 (Ubiquitous)

The codex review gate wrapper shall exit 2 with the verdict summary — the handler's standard error from the sentinel line `codex review gate: FAIL` onward and nothing before it — only when the handler exits 2 and its standard error contains that line; shall exit 0 for every other handler exit status, including an exit 2 without the sentinel line, recording a row for such a non-verdict exit 2 in the gate's log; and shall keep its pure-shell off-switch, its exit 0 when the moai binary cannot be resolved, and the exit-status propagation of the registered command form.

래퍼는 지금 `exit 0` 으로 끝난다(§A.1). 종료 코드 2 는 Go 패닉이나 시스템 오류도 내므로(A.7) 상태 코드만으로는 판정 실패를 알 수 없다 — sentinel 줄이 판정 신호의 두 번째 조건이다. 스코프 로그 행 같은 앞선 stderr 줄은 알림에 실리지 않는다. 2 가 아닌 비0 종료(프로세스 오류 등)와 sentinel 없는 2 를 0 으로 접는 것은 "깨움은 판정 실패에만" 규율과 fail-open 의 결합이다. 순수 셸 off 스위치는 비활성일 때 moai 콜드 스타트를 0 으로 유지하는 기존 계약이다. 등록된 명령 형태(`bash -c '[ -f "$0" ] && exec bash "$0"; …'`)는 `exec` 로 종료 상태를 그대로 전달하며 이 SPEC 이 그 형태를 바꾸지 않는다.

### B. 중복 방지와 재전달 억제 (운영자 요구 2)

### REQ-CRA-004 — 트리당 한 번에 하나의 리뷰, 보유 시간은 유한하다 (While + shall not)

**While** a review of a tree is held by a live process, a Stop in that tree shall exit silently without invoking the reviewer; the review lock shall be keyed on the tree's canonical root, shall be released when its holder exits so that a dead holder's lock never blocks a later Stop, and shall be held only across steps that are each time-bounded, so that a live holder cannot keep it longer than the review budget plus the state-computation allowance; and the lock, the redelivery record, and the log shall live in the tree's git directory, never in its working tree.

락은 트리 단위(정규화한 git toplevel, 하위 디렉터리·심볼릭 링크 경로로 들어온 Stop 도 같은 락)이고 세션 단위가 아니다 — 같은 트리의 여러 세션도 서로의 리뷰를 겹치지 않는다. 같은 저장소의 다른 연결 워크트리는 다른 트리다(git 디렉터리가 다르다). "건너뜀"이지 "기다림"이 아니다(기다림은 훅 프로세스를 쌓는다). 기반은 gate-run 락의 flock 기반 원시 기능이며(A.5), 윈도우는 원자적 생성+보유자 PID 생존 확인으로 죽은 보유자를 인수한다. **보유 시간 상한(D7):** 락 안에서 도는 단계는 리뷰어 호출(`DefaultCodexReviewGateTimeout` = 900s)과 상태 계산(트리 루트·git 디렉터리·상태 키·버전 프로브·종료 후 재계산)뿐이고, 상태 계산은 단계 묶음마다 상태 계산 예산(60s)을 갖는다. 시간이 끝나면 `state-unavailable` 행을 남기고 상태 기반 단계를 건너뛰어 진행한다(fail-open). 따라서 보유 상한은 900s + 2×60s 이다. 파일 위치 근거는 A.5 의 마지막 불릿 두 개다 — git 디렉터리 아래라서 `.gitignore` 유무와 무관하게 작업 트리 상태 키를 바꾸지 않는다.

### REQ-CRA-005 — 같은 상태는 다시 리뷰하지도 다시 깨우지도 않는다, 전달 확인 뒤에만 (While + shall not)

**While** the reviewed-state key — the five fields the Codex receipt binds: the HEAD, the working-tree digest, the config digest, the review command, and the tool version — equals the key of the last recorded review of that tree and that record is a pass, or a fail whose summary was written to standard error for the same session within the delivery window of thirty minutes, the gate shall not invoke the reviewer and shall not wake the session; the gate shall mark a failing record as delivered only after its summary was written, and shall treat a failing record that was never so marked, that belongs to another session, or that is older than the window as undelivered, reviewing and delivering again.

깨움 루프(wake loop) 방지의 첫 경계다. 기록은 트리별 상태 파일(`<git 디렉터리>/moai/codex-review-wake.json`)이고 락 안에서 쓴다. inconclusive 는 기록하지 않아 재시도를 허용한다. **첫 절은 현행 동작의 변경이다(Q16).** 오늘은 작업 트리 상태가 그대로여도 변경이 있는 Stop 마다 리뷰가 다시 돈다(핸들러에 재사용 기록이 없다 — §A.1). 이 SPEC 이후 pass 든 전달된 fail 이든 이미 리뷰한 같은 상태는 다음 Stop 에서 다시 리뷰하지 않는다. 상태가 바뀐 Stop 은 오늘처럼 리뷰한다(회귀 칸, AC-005). **전달 확인(D2):** 기록 시점에는 `delivered` 가 거짓이다. 실행기가 stderr 요약을 성공적으로 쓴 직후 같은 기록을 `delivered` 로 바꾼다. 쓰기 전에 훅이 취소·종료된 fail 은 기록에 `delivered` 가 없어 다음 Stop 이 다시 리뷰하고 전달한다. 기록은 세션 id 를 갖는다 — 같은 트리의 다른 세션은 이미 전달된 같은 상태의 fail 도 한 번은 받는다. `delivered` 표지와 기록 갱신 사이의 짧은 구간(실행기가 stderr 를 쓴 뒤 표지를 쓰기 전)에 들어온 Stop 은 미전달로 읽혀 한 번 더 리뷰·전달할 수 있다 — 연속 실패 상한(REQ-CRA-014)이 총량을 묶고 락이 겹침을 막는 한계 안의 잔여 위험이다.

### REQ-CRA-014 — 연속 실패 깨움은 트리당 세 번까지다 (While + shall not)

The gate shall deliver at most three consecutive failure wake-ups per tree, counted from the tree's last passing review or from the creation of its record, irrespective of the findings' wording or number; **while** that count is three, every failing review of that tree shall exit silently with a logged row and no wake-up, and only a passing review shall reset the count to zero.

깨움 루프 방지의 둘째 경계다. 깨어난 세션이 고치고 멈추면 Stop 이 다시 돌고, `stop_hook_active` 는 비동기 훅을 보호한다고 알려진 바가 없다(E-3/E-4 미관측). 그래서 경계는 상태 키와 이 상한이다. 세는 것은 **그 트리의 깨움(전달 시도) 횟수**이지 "동일한 findings"가 아니다 — findings 지문은 LLM 산문 제목에 의존해 재서술로 회피되고, findings 0건 fail 에서 충돌한다(A.5). 상한은 `internal/config` 단일 원천(3)이다. 상한 뒤에도 리뷰는 계속 돌고(pass 를 알아야 재무장할 수 있다) 결과는 조용히 `failure-cap` 행으로만 남는다. 해제 조건은 pass 하나다 — 상태가 바뀌어도, findings 가 달라져도 풀리지 않는다. 마지막 알림의 고정 문구 `이후 알림 없음`은 이 규칙을 그대로 말한다 — 문구가 "알림이 더는 없다"고 하는 한 상한 뒤에 일부 실패만 다시 알리는 규칙은 문구와 모순된다. **이 읽기는 오케스트레이터 문구("동일한 실패는 reviewed state 가 바뀔 때까지 억제")를 좁혀 읽은 것이므로 확인이 필요하다**(plan.md §G, decision-index Q19).

### REQ-CRA-015 — 세 번째 깨움이 마지막 알림이며 미해결을 밝힌다 (When)

**When** a failure wake-up is the third consecutive one for the tree, its summary shall end with a final notice that contains the exact text `이후 알림 없음, 상태는 미해결` and states, in English, that no further notifications will follow, that the state is unresolved, and that a passing review re-arms notifications.

침묵이 "통과"로 오독되는 위험을 막으려고 침묵 대신 마지막 전달 하나로 미해결을 명시한다(Jev `cap3_silent` 0.21 을 오케스트레이터가 `cap3_notify_once_at_cap` 으로 재정, 리더 수용 — 잠정). 세 번째 깨움이 그 마지막 알림 자체이며 네 번째 별도 알림은 없다. 고정 한국어 문구는 시험이 정확히 일치를 보는 문자열이고, 영어 줄은 같은 뜻의 병기다. 마지막 알림 줄은 크기 상한에 잘리지 않는다(REQ-CRA-001).

### C. 오래된 결과 (운영자 요구 3)

### REQ-CRA-006 — 오래된 결과는 표지를 붙여 전달한다 (When)

**When** a review that started at one reviewed-state key completes while the tree's current key differs, the gate shall still deliver a failing result, shall name both the reviewed HEAD and the current HEAD in it, and shall mark it as a review of an earlier state; the gate shall never discard such a result silently and never deliver it without that mark, and the record of that result shall carry the key that was reviewed, not the current one.

판정 근거(Jev 0.99, `label_stale`). 오래됨의 정의는 HEAD 만이 아니라 상태 키 전체다 — HEAD 는 같고 작업 트리 digest 만 달라진 경우(미커밋 편집)도 오래된 결과이며, 이때 두 HEAD 는 같은 값으로 적히고 표지가 붙는다. 반대로 HEAD 만 바뀌고 digest 는 같은 경우(빈 커밋)도 오래된 결과다. 이것은 운영자 문구("HEAD")를 상태 키 전체로 넓힌 것이며 리더가 수용했다(Q11). 통과 결과는 전달할 것이 없다. 기록 키는 리뷰한 상태(시작 상태)다 — 아직 리뷰하지 않은 현재 상태가 "이미 실패로 알려진 상태"로 오인되지 않는다.

### D. 등록과 타임아웃 (운영자 요구 4)

### REQ-CRA-007 — 레지스트리와 와이어링 비교 (Ubiquitous)

The tracked `.claude/settings.json` and its template mirror shall register the codex review gate wrapper with the `asyncRewake` field set to true, shall carry that field on no other hook entry, and shall leave the multi review gate wrapper registration unchanged, and the hook-entry wiring comparison shall include `asyncRewake` in its key and in its ordering so that a missing or extra flag is reported as a divergence deterministically.

`HookEntry` 가 `asyncRewake` 를 키로 삼지 않으면 `moai doctor` 의 와이어링 비교가 이 플래그의 누락을 보지 못한다(A.5). 정렬 비교(`sortHookEntries`)가 새 필드를 무시하면 플래그만 다른 두 항목의 순서가 비결정적이다. 이 요구는 E-1 이 깨움을 확인할 때까지 효력이 없다(REQ-CRA-011).

### REQ-CRA-008 — 타임아웃 정책 (Ubiquitous)

The registered hook timeout of the codex review gate wrapper shall exceed the bound the Go handler enforces on its own execution — the review budget plus the state-computation allowance — by a stated margin, so that the in-process fail-open path fires before any runtime cancellation, and the margin value shall be fixed from the recorded observation of how `timeout` applies to an `asyncRewake` hook.

현행 두 층: 훅 `timeout` 900(동기에서는 Claude Code 가 초과 시 취소·출력 폐기 — 레퍼런스의 일반 문구)과 Go 쪽 `DefaultCodexReviewGateTimeout` 900s(`codex_review_gate.go:96`). 락 도입 뒤 핸들러의 총 상한은 900s + 2×60s = 1020s 이므로 등록 값은 그보다 마진만큼 커야 한다. `asyncRewake` 에서 `timeout` 이 어떻게 적용되는지는 E-2 미관측이라 마진 값은 프로브 뒤에 정한다(plan.md §G Open decision O-B). 이 요구는 마진이 정해질 때까지 효력이 없다(REQ-CRA-011).

### E. 경계

### REQ-CRA-009 — Codex 경로는 변하지 않는다 (Ubiquitous + shall not)

The Codex Stop chain's codex review member, the receipt producer `moai verify codex-review`, and the receipt store shall behave exactly as before this SPEC, shall not read or write the review lock, the redelivery record, or the gate log, and shall compute the same working-tree state key as before whether or not those files exist, the Claude path sharing with them only the scope resolver, the request assembler, the review driver, and the state computation.

락·기록·로그를 git 디렉터리에 두는 이유는 이것이다 — 작업 트리에 두면 `.gitignore` 가 그 경로를 숨기는 저장소에서만 키가 안정하고, 그렇지 않은 저장소에서는 `verify.Key` 를 쓰는 Codex 경로의 키까지 바뀐다(A.5).

### REQ-CRA-010 — 형제 SPEC 의 skip 정책이 락보다 앞선다 (While)

**While** the sibling `tree_scope` policy skips a session, the gate shall create neither the review lock nor the redelivery record nor the gate log and shall not invoke the reviewer.

형제 SPEC(`SPEC-CODEX-REVIEW-OWNERSHIP-001`, draft — 이 SPEC 의 `depends_on`)이 먼저 착지하고 `HandleCodexReviewGate` 시그니처를 바꾸지 않았다. 이 SPEC 도 시그니처를 바꾸지 않는다 — 핸들러가 요약 문구(`Reason`)를 구성하고 실행기가 stderr·종료 코드로 옮긴다(plan.md §B.1). 형제 plan.md §B.2 의 배선 결정과의 경계: 형제는 스코프 클래스가 트리일 때 핸들러 안에서 설정 루트를 다시 구해 `tree_scope` 를 읽고 스코프 해상 직후·셀프게이트 앞에서 skip 하며, 이 SPEC 의 삽입 지점(트리 루트·락·상태 키)은 그 뒤 codex 조회 다음이다 — 두 SPEC 의 삽입 위치가 겹치지 않는다.

### F. 관측 게이트와 안전

### REQ-CRA-011 — 깨움 가설은 프로브가 확인할 때까지 가설이다 (Where + shall not)

**Where** the live probe has not recorded that a finished `asyncRewake` Stop hook starts a new turn in an idle session, neither settings file shall register `asyncRewake`, the wake-dependent wording shall remain a stated hypothesis, and requirements REQ-CRA-007 and REQ-CRA-008 shall be treated as not yet in force; the five probe items shall be recorded under `.moai/reports/t1422/async-probe/` before any milestone that depends on them starts.

E-1 이 "깨우지 않는다"로 나오면 이 SPEC 의 등록 요구(007·008)는 철회되고 핸들러 쪽 개선(001~006, 009, 010, 014, 015)은 동기 등록 아래에서도 유효하다 — 종료 코드 2 는 동기 `Stop` 훅에서 "Claude 가 멈추지 못하게 함"이다(A.2).

### REQ-CRA-012 — 등록은 마지막이고 레인은 스스로를 깨우지 않는다 (Ubiquitous)

The registration change shall be the last run-phase change, and a session whose workflow configuration does not enable the gate shall cost no moai process and shall produce no wake.

이 저장소의 추적 `workflow.yaml` 에는 `review_gate` 키가 없다(`grep -c review_gate .moai/config/sections/workflow.yaml` → 0, 본 트리, 2026-10-02 관측). 순수 셸 off 스위치(REQ-CRA-003)가 래퍼를 즉시 끝내므로 레지스트리가 바뀌어도 설정이 게이트를 켜지 않은 세션은 깨어나지 않는다. 훅이 읽는 `workflow.yaml` 은 래퍼가 `${CLAUDE_PROJECT_DIR:-$PWD}` 기준으로 읽는 파일이며 연결 워크트리 세션에서 `CLAUDE_PROJECT_DIR` 이 primary 를 가리키면 primary 의 로컬 수정본이다 — 그 값은 이동하는 운영자 상태라 이 요구는 값에 의존하지 않는다(plan.md §E). 진행 중 세션에 설정 편집이 언제 반영되는지는 E-5 미관측이다.

### REQ-CRA-013 — 억제는 기록된다 (When)

**When** the gate suppresses a review or a delivery — a held lock, an unchanged state, the failure cap, an unavailable lock or state — or the wrapper folds a non-verdict exit 2 into exit 0, it shall append one row naming the reason to the gate log in the tree's git directory.

"조용히 종료"는 "왜 조용한지 알 수 없음"이 아니어야 한다(verification-completeness §1.3: 멈췄을 때 무엇이 달라 보이는가). 이 로그가 "게이트가 일을 하고 있지 않다"와 "누군가 이미 하고 있다/이미 알렸다"를 가른다. 로그 위치는 `<git 디렉터리>/moai/codex-review-gate.log`(REQ-CRA-004)이며 이전 초안의 `<트리>/.moai/logs/` 가 아니다 — `moai hook codex-review-gate --help` 문구에 경로를 적는다.

---

## §C AC 형태에 대한 구속 [HARD]

- AC 는 **종료 코드·stdout/stderr 내용**, **리뷰어 호출 횟수**(주입 seam `codexSession`/`codexLookPath`/RPC 스텁), **락·상태 파일의 존재와 내용**, **설정 파일의 키·값**을 관측한다. verdict 값 단독은 근거가 못 된다.
- 동시성 AC 는 장벽(barrier)으로 겹침을 **강제**한다 — 보유자가 리뷰어 안에서 다른 호출이 모두 되돌아올 때까지 대기하고, 호출 횟수를 센다. 겹치지 않고 우연히 순차로 돈 초록을 막는다.
- **stderr 는 운영 형태로 관측한다(D4)**: 프로세스 `os.Stderr` 와 `cmd.ErrOrStderr()` 를 모두 잡아 도착 순서대로 합친 스트림을 단정한다. 래퍼 AC 는 **등록된 명령 문자열**(설정 파일의 `command`/`args`)을 그대로 실행한다.
- 라이브 Claude Code·라이브 codex 의존 AC 는 두지 않는다. 라이브 프로브는 AC 가 아니라 **기록되는 관측**이며 skip 은 미관측(Gap)이다.
- RED·GREEN 실행은 `-v` 로 `=== RUN` 을 함께 관측한다(0매칭 초록 방지).

## §D 실행 순서 구속

1. **M1 프로브와 회귀선 먼저.** 프로브 E-1~E-5 와 변경 전 초록 관측을 마친다. 프로브 결과가 M2 이후를 게이트한다: E-3·E-4 → 핸들러·락·래퍼(M2, M3), E-2 → 타임아웃(M4), E-1·E-5 → 등록(M5). E-1 이 부정이면 M5 와 REQ-CRA-007/008 을 철회한다. 리더가 Gap 수용 기록을 남기면 그 항목만 예외다.
2. **등록(M5)은 마지막 run-phase 변경이다.** 프로브 산출물의 sha256 이 tracked `progress.md` 에 기록된 커밋이 등록 커밋보다 먼저여야 한다(커밋 그래프로 순서를 증명 — `verification-claim-integrity.md` §2.3).

---

## §E 범위 밖 (Out of Scope)

### Out of Scope — 소유권·스코프·자기 리뷰 도구 (형제 SPEC)

- `tree_scope` 키, 자기 리뷰 도구 `codex_review`/`glm_review`, 카드 리뷰 단계 교리, 카드 t1426(감사 도구 `baseBranch`)은 `SPEC-CODEX-REVIEW-OWNERSHIP-001` 소관이다. 이 SPEC 은 그 결정을 바꾸지 않고, 형제의 skip 판정이 락보다 앞선다는 순서만 요구한다(REQ-CRA-010).

### Out of Scope — Codex 경로, 영수증, multi 게이트

- Codex Stop 체인·`moai verify codex-review`·검증 영수증 저장소는 변경하지 않는다(REQ-CRA-009). 두 경로가 같은 트리를 동시에 리뷰하는 중복은 이 SPEC 이 막지 않는다(명시 실행 경로는 락을 잡지 않는다).
- multi 리뷰 게이트와 그 래퍼 등록은 바꾸지 않는다(A.6).

### Out of Scope — 리뷰 내용과 Claude Code 자체

- 리뷰 프롬프트·판정 기준·모델 핀·900s 리뷰 예산 값은 바꾸지 않는다. Claude Code 의 `asyncRewake` 구현을 바꾸거나 문서화하는 일, 문서에 없는 동작을 가정한 우회 구현은 하지 않는다.
- 같은 부류의 다른 폭주(t1425 의 harness 관측 로그 retention)는 그 카드 소관이다.

### Out of Scope — 오래된 Claude Code 버전 호환 검증

- `asyncRewake` 를 모르는 Claude Code 가 이 필드를 어떻게 취급하는지(무시? 검증 오류?)는 이 레인이 관측하지 못했다 — Gap. 최소 지원 버전 선언은 하지 않는다.

---

## §F 제약

- **fail-open 불변**(SPEC-MOAI-MCP-SERVER-001 REQ-MCP-012): 리뷰어 부재·오류·inconclusive·락 경합·상태 파일 오류·상태 계산 시간 초과는 모두 종료 코드 0 이다. 락 기계 오류(git 디렉터리 해상·생성 불가 등)는 리뷰를 막지 않고 락 없이 진행하되 REQ-CRA-013 에 기록한다(gate-run 락의 "never blocks, never fails" 선례).
- **단일 해상기·단일 상태 계산**: 새 판별기·새 diff 계산을 만들지 않는다(형제 SPEC §F).
- **단순성**(AGENTS.md §5): gate-run 락 원시 기능을 일반화해 재사용하고, 재전달 상태는 기존 상태 키를 쓴다. 새 의존성 없음. 규모·3배 점검은 plan.md §H.
- **Template-First**: `.claude/settings.json`·래퍼는 `internal/template/templates/` 미러와 같은 변경으로 들어간다. 미러 본문은 SPEC ID·REQ 토큰·카드 번호·날짜를 담지 않는 중립 문구다.
- **크로스 플랫폼**: 락 일반화는 유닉스·윈도우 두 구현을 함께 고친다(`GOOS=windows GOARCH=amd64 go build ./...`).
- **하드코딩 금지**: 연속 실패 상한(3)·전달 창(30분)·상태 계산 예산(60s)·findings/크기 상한·마진은 `internal/config` 단일 원천이다.
- **측정 규율**: 소관 패키지 단위(`./internal/cli/...`, `./internal/hook/...`, `./internal/template/...`, `./internal/config/...`). 전체 스위트 로컬 금지.

## §G 참조

- `internal/cli/codex_review_gate.go:69-119,192-211` · `internal/cli/codex_review_scope.go:340-356` · `internal/cli/gate_lock.go` · `gate_lock_unix.go:38-61` · `gate_lock_windows.go:81,128` · `internal/kanban/factory_alive_unix.go:22` · `internal/auditreceipt/treeroot.go:22`
- `internal/verify/key.go:40-100` · `internal/verify/receipt.go:37-43` · `internal/cli/codex_review_receipt.go:60-100,126-174` · `internal/cli/codex_stop_chain.go:613-669` · `internal/cli/multi_review_gate.go:17-23,75` · `internal/cli/mcp_codex.go:1757-1768,1795-1799`
- `.claude/settings.json:186-200` · `internal/template/templates/.claude/settings.json.tmpl:191` · `…/hooks/moai/handle-codex-review-gate.sh:35-36,60,84-86` · `internal/template/hook_entries.go:17-30,53-56,197` · `internal/cli/doctor_hook_wiring.go:79`
- `internal/template/review_gate_registration_test.go` · `internal/hook/review_gate_selfgate_test.go` · `internal/cli/CLAUDE.md:13`
- 공식 훅 레퍼런스 https://code.claude.com/docs/en/hooks (명령 훅 필드 표, 종료 코드 2 표)
- `SPEC-CODEX-REVIEW-OWNERSHIP-001`(형제, 먼저 착지) · `SPEC-CODEX-GATE-SCOPE-001`(REQ-CGS-010 로그) · `SPEC-MOAI-MCP-SERVER-001`(REQ-MCP-012) · `SPEC-DUAL-HARNESS-HOOK-PARITY-001`(Codex 멤버 6·영수증)
- `.moai/reports/t1422/plan-audit.md`·`async-plan-audit.md` — 형제·이 SPEC plan-audit 보고서(로컬 증거)
