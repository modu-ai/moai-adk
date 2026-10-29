# 코드 감사 1회차 처분표 — 재감사 전 (B9)

- 작성일: 2026-10-30
- 감사 대상: `4cdaba5d7e` (`4cdaba5d7e6821aa9751d100cf75a328b4726704`), 판정 FAIL 0.67
- 판정서: `/Users/goos/MoAI/moai-adk-go/.moai/worktrees/agentic-factory-integration/.moai/reports/issue-1948/code-audit-4cdaba5d7e/verdict.md` (이 작성자가 읽음, 읽기만 함)
- 이 표가 말하는 시점: 아래 각 lane 브랜치 기록의 끝을 읽은 상태다.
  - lane-18 `WT-agentic-factory-core-claim-runner` @ `7196d0b5ea`
  - lane-20 `WT-agentic-monitor-effect-reconcile` @ `89770eaa47`
  - lane-21 `WT-browser-select` @ `e196a35ff2`
  - lane-17 설정 `WT-agentic-factory-settings-keys` @ `e478302354`
  - lane-17 SPEC `WT-agentic-factory-contract` @ `a09a08e1df`
  - lane-17 template `WT-agentic-factory-templates` @ `47d8b782b7` (`4cdaba5d7e` 뒤 커밋 0)
  - 통합 `WT-agentic-factory-integration` @ `4cdaba5d7e` (`4cdaba5d7e` 뒤 커밋 0)
  - lane-22: **lane-22 branch unknown**. `git branch --list`에서 lane-22로 식별할 이름을 찾지 못했고, `--grep='AF-22'`·`--grep='B10'`로도 lane-22 커밋을 찾지 못했다. 추측하지 않았다.
- 갱신: 이 표는 세션이 깨어날 때마다 갱신한다(고정 일정 없음).
- 판정 규칙: 커밋 제목이 발견 항목을 부르거나(예: `F1`) 고침을 분명히 말할 때만 그 커밋을 증거로 연결했고, 그때는 `git show --stat`으로 바뀐 파일이 맞는지 확인했다. 같은 영역을 만진 것만으로는 증거가 아니다. `수리됨`은 고침이 커밋에 들어 있다는 뜻이며, **고침이 동작한다는 뜻이 아니다**(아무것도 실행하지 않았다).
- 리더 전달(relayed, 이 작성자가 검증하지 않음): F1은 lane-20+lane-18, F2 대응표는 lane-17이 함, Facts 연결은 lane-22 B10에서 진행 중, HostFacts는 lane-18 `6b945fd844`.
- F2는 성격이 다른 네 부분이 있어 (가)~(라) 네 행으로 나눴다.

| 항목 | 심각도 | 발견(판정서에서 줄임, 위치는 판정서 그대로) | 담당 lane(판정서 "담당") | 수리 SHA | 상태 | 증거 |
|---|---|---|---|---|---|---|
| F1 | BLOCKING | effect Executor가 동시 실행에서 같은 키를 두 번 송신 — `effect/intent.go:183-244`, `effect/docstore/docstore.go:46-61` | lane-20 + lane-18 | `ba56671f20` | 수리됨 | 제목 "(F1, Issue #1948)"; `git show --stat`: `docstore.go`·`intent.go`·`intent_cas_test.go`·`putif_test.go` 등 6 파일 — 판정서가 짚은 두 파일 포함. 동작 확인은 안 함 |
| F2 (가) | BLOCKING | AC 명령이 고르는 시험 이름 47/48이 없음 — `acceptance.md` §C·§E | lane-17 | `17c1401810`, `546f7af570`, `11822055ad` | 수리됨 | 판정서의 수정안 중 "acceptance.md의 명령을 실제 시험 이름으로 다시 쓴다"에 해당. 세 제목이 각각 명령 재작성(F2)·없는 이름의 lane 배정(F2c)·하위 SPEC 명령 제거를 말하고, stat은 두 SPEC의 `acceptance.md`만 바꿈. 기준 자체는 대부분 여전히 미충족(acceptance §G) |
| F2 (나) | BLOCKING | 병합 게이트가 production에서 항상 `AttributionLocked` | lane-17(계약) + lane-18 | `3a3392ddc7`(분류 기록) | 계약상 의도 | REQ-AF-011 잠금(리더 판정, relayed). 제목 "lock kept per REQ-AF-011 is the intended state"; stat은 acceptance·spec·progress. 코드가 잠겨 있다는 관측은 감사 보고(리더 전달)뿐 |
| F2 (다) | BLOCKING | production `runner.guard`가 항상 `identity-unknown`으로 거절(`runner.Facts` 구현 0, `ghidentity.IssueFacts` 서명 불일치) | lane-18(어댑터) + lane-22(배선, B10) | `6b945fd844`(어댑터); lane-22 배선: 없음(커밋 미확인) | 진행 중 | `6b945fd844` 제목 "HostFacts, the production Facts adapter"; stat `runner/hostfacts.go`·`hostfacts_test.go`. 배선은 lane-22 B10 진행 중(리더 전달), lane-22 branch unknown. 분류 기록 `269bfe502e`(제목 "a defect under repair (B10), not the intended lock") |
| F2 (라) | BLOCKING | IDENTITY REQ-AFI-001~007의 production 구현 없음(`runner.Attributor` 구현 0) | lane-18 | `2e2a714e3c` | 진행 중 | 제목 "the read-only producer attribution lookup of the identity SPEC"; stat `identity/identity.go`·`identity_test.go`·`runner/runner.go` 등 5 파일. 하위 SPEC 기준은 `4cdaba5d7e` 기준 전부 미충족이고 잠금 해제 조건(REQ-AF-011)도 아직이라 진행 중으로 둠 |
| F3 | SHOULD-FIX | 죽은 프로세스의 전역 잠금이 알림·확대를 영원히 멈춤 — `claim/claim.go:282-309`, `claim/docs.go:70-75`, `monitor/notify/notify.go:306-310`, `cli/factory_resident.go:216-222` | lane-20 + lane-18 + lane-22 | 없음 | 미착수 | F3을 부르거나 이 고침을 말하는 커밋 없음. `5d383905f5`("documents lock wait covers the resident writers' queue")는 같은 영역이지만 죽은 잠금 재발급·확대를 말하지 않아 증거로 세지 않음. lane-22 branch unknown |
| F4 | SHOULD-FIX | 상주 루프의 gh 읽기에 호출별 시간 제한 없음 — `monitor/notify/ghpr/ghpr.go:65-73`, `resident/resident.go:238-245` | lane-20 | `89770eaa47` | 수리됨 | 제목 "(F4, Issue #1948)"; stat `ghpr.go`·`ghpr_test.go`·`resident.go`·`ghexec_test.go`. 동작 확인은 안 함 |
| F5 | SHOULD-FIX | TakeOver·활성 판정이 호출자 불리언에만 기댐 — `claim/claim.go:25-87`, `runner/runner.go:669-719`, `claim/next.go:262-309` | lane-18 + lane-22 | 없음 | 미착수 | F5를 부르는 커밋 없음. 병합 `0137f1fd88` 제목의 "takeover"는 고침을 말하지 않아 세지 않음. lane-22 branch unknown |
| F6 | SHOULD-FIX | runner가 전이표 E01–E27을 거치지 않음 — `runner/runner.go:186-193`, `:739` | lane-18 | 없음 | 미착수 | 해당 커밋 없음 |
| F7 | SHOULD-FIX | control ref 권위 미배선, 배선 시 시간 제한 없이 Issue 잠금 보유 — `claim/controlref.go:16-20,47`, `gitref/gitref.go:127-131` | lane-18 | 없음 | 미착수 | 해당 커밋 없음 |
| F8 | SHOULD-FIX | effect 의도 `LastError`가 가려지지 않은 채 저장, 가림 엔진 두 벌 — `effect/intent.go:223,229,233,235,237,264`, `ghexec.go:67-69`, `gitexec.go:53-55`, `core/redact.go:13-24`, `monitor/mask.go:12-22` | lane-20 | 없음 | 미착수 | lane-20 기록에 F8 커밋 없음. lane-18 `5a97c569b4`("one sanitizer, core.Sanitize")는 stat이 `core/redact.go`·`core/errors.go`·`core/sanitize_test.go`뿐이라 `LastError` 저장과 `monitor/mask.go`를 다루지 않아 이 항목의 증거로 세지 않음 |
| F9 | SHOULD-FIX | 설정 로더가 inbox `max_id_bytes` < 32를 받아 모든 알림 실패 — `config/agentic_factory_settings.go:343-364`, `monitor/notify/inbox.go:24-26,45-50` | lane-17 / lane-20 | `e478302354` | 수리됨 | 제목 "reject notify.inbox.max_id_bytes below 32 in the AF settings reader"; stat `internal/config/agentic_factory_settings.go`·`_test.go`. 동작 확인은 안 함 |
| F10 | SHOULD-FIX | production lint 4건 — `gitref/gitref.go:388,389,391`, `effect/labels/githost.go:115` | lane-18, lane-20 | 없음 | 미착수 | 해당 커밋 없음 |
| F11 | SHOULD-FIX | `claim` 커버리지 84.4% < 85% — `internal/agenticfactory/claim` | lane-18 | 없음 | 미착수 | 커버리지를 말하는 커밋 없음(재지 않음) |
| F12 | MINOR | ghexec 허용 목록의 틈(`--web=true`, `-R`, 저장소 고정 없음) — `effect/ghexec/ghexec.go:140-157` | lane-20 | 없음 | 미착수 | 해당 커밋 없음 |
| F13 | MINOR | `ReissueLock` 확인-삭제 사이 TOCTOU — `claim/lock.go:76-90` | lane-18 | 없음 | 미착수 | 해당 커밋 없음 |
| F14 | MINOR | `Release`가 epoch 미확인 — `claim/claim.go:681-696` | lane-18 | 없음 | 미착수 | 해당 커밋 없음 |
| F15 | MINOR | 인계 모델 두 벌 — `reconcile/handoff.go:62-102` 대 `runner/runner.go:484-490` | lane-18/20 | 없음 | 미착수 | 해당 커밋 없음 |
| F16 | MINOR | 브라우저 선택 키 위치·파서 중복 — `web/browser_select.go:74-95` | lane-21/17 | 없음 | 미착수 | lane-21 기록에 F16을 부르는 커밋 없음. `e196a35ff2`("AC-AF-020 cases … under the TestBrowserSelect_ names")는 시험 추가이며 파서 중복 고침을 말하지 않음. lane-17 설정·template 기록에도 없음 |
| F17 | MINOR | 상주 명령 시작 문구가 사실과 다름 — `cli/factory_resident.go:185` | lane-22 | 없음(lane-22 branch unknown) | 미착수 | lane-22 브랜치를 찾지 못해 그 기록은 읽지 못함. 읽은 브랜치들에는 해당 커밋 없음 |
| F18 | MINOR | 주석 중복 — `cli/codex_review_receipt.go:245-248` | lane-22 | 없음(lane-22 branch unknown) | 미착수 | F17과 같음 |
| F19 | MINOR(가설) | 위치 메모 키·지문의 빈틈 — `homestate/location_memo.go:104-113,170-198` | lane-22 | 없음(lane-22 branch unknown) | 미착수 | F17과 같음 |
| F20 | MINOR(가설) | 자기 Issue 경로가 redirect·저장소 불일치를 보지 않음 — `core/authz.go:156-165` | lane-18 | 없음 | 미착수 | 해당 커밋 없음 |
| F21 | MINOR(가설) | `ReadyGate`가 첫 요청만 봄 — `claim/request.go:478-497` | lane-18 | 없음 | 미착수 | 해당 커밋 없음 |
| F22 | MINOR | 시험용 `FakeExecutor`가 production 파일에 있음 — `review/provider.go:136-149` | lane-20 | 없음 | 미착수 | 해당 커밋 없음 |
| F23 | MINOR | webhook `Serve` 오류가 종료 때까지 안 드러남 — `cli/factory_resident.go:199,232` | lane-22 | 없음(lane-22 branch unknown) | 미착수 | F17과 같음 |
| F24 | MINOR(가설) | `gh pr checks --required`가 필수 검사 없는 저장소에서 매 pass `pr-read-error` — `ghpr.go:69-72` | lane-20 | 없음 | 미착수 | `89770eaa47`(F4)는 같은 파일이지만 제목이 시간 제한만 말해 세지 않음 |

셈(27행: F1, F2 네 행, F3~F24): 수리됨 4 · 진행 중 2 · 계약상 의도 1 · 미착수 20.

- 지시에서 "F2 계약 쪽"으로 받은 일곱 커밋 가운데 `b9db20533b`(B8a, lane 자율 운영)와 `a09a08e1df`(DEC-AF-28, RESTRICTED SHA-256)는 제목이 F2를 부르지도 F2의 고침을 말하지도 않아, 위 규칙에 따라 F2 증거로 세지 않았다. 원장에는 남겼다.
- lane-22 브랜치를 알게 되면 F3·F5·F17·F18·F19·F23과 F2 (다)를 다시 읽는다.

## 원장 (실행한 git 명령과 첫 출력 줄)

모두 읽기 전용이다. 체크아웃·트리 전환·쓰기 없음.

```
git branch --list --format='%(refname:short) %(objectname:short)'
  → GoosLab/main-fork a845881c00
  (긴 출력 중 관련 줄: WT-agentic-factory-contract a09a08e1df / WT-agentic-factory-core-claim-runner 7196d0b5ea / WT-agentic-factory-integration 4cdaba5d7e / WT-agentic-factory-review cc5709b3a9 / WT-agentic-factory-settings-keys e478302354 / WT-agentic-factory-templates 47d8b782b7 / WT-agentic-monitor-effect-reconcile 89770eaa47 / WT-browser-select e196a35ff2)
git log --format='%h %s' 4cdaba5d7e..WT-agentic-factory-core-claim-runner
  → 7196d0b5ea [AF-18] fix(agenticfactory/mirror): post a RESTRICTED record when a ledger is refused (design 4.4)
git log --format='%h %s' 4cdaba5d7e..WT-agentic-monitor-effect-reconcile
  → 89770eaa47 fix(agenticfactory/ghpr): bound each gh read with its own time limit (F4, Issue #1948)
git log --format='%h %s' 4cdaba5d7e..WT-browser-select
  → e196a35ff2 test(browser): AC-AF-020 cases (d), (e) and the resolver half of (f) under the TestBrowserSelect_ names (Issue #1948, B7)
git log --format='%h %s' 4cdaba5d7e..WT-agentic-factory-settings-keys
  → e478302354 fix(config): reject notify.inbox.max_id_bytes below 32 in the AF settings reader (Issue #1948)
git log --format='%h %s' 4cdaba5d7e..WT-agentic-factory-contract
  → a09a08e1df docs(spec): SPEC-AGENTIC-FACTORY-001 DEC-AF-28 - RESTRICTED SHA-256 is over the masked ledger form, never the raw value (Issue #1948)
git log --format='%h %s' 4cdaba5d7e..WT-agentic-factory-templates
  → 출력 없음
git log --format='%h %s' 4cdaba5d7e..WT-agentic-factory-review
  → cc5709b3a9 [AF-19] merge lane-18 3dd74138d2 (S46 stage/sub_stage/park/gate/refusal) (Issue #1948)   (lane-19 커밋이라 lane-22로 보지 않음)
git log --format='%h %s' 4cdaba5d7e..WT-agentic-factory-integration
  → 출력 없음
git log --format='%h %s' -1 WT-agentic-factory-integration
  → 4cdaba5d7e merge(lane-20): docstore Busy handling in notify, webhook 503, concurrent-save test (49192a6971) (Issue #1948)
git log --branches --format='%h %s' --grep='AF-22' -n 10
  → 출력 없음
git log --branches --format='%h %s' --grep='B10' -n 10
  → 269bfe502e docs(spec): SPEC-AGENTIC-FACTORY-001 acceptance - runner.guard identity-unknown is a defect under repair (B10), not the intended lock (Issue #1948)   (나머지 9줄은 다른 SPEC 커밋)
git log --branches --format='%h %s' --grep='Issue #1948' --since=2026-10-28 -n 80
  → a09a08e1df docs(spec): SPEC-AGENTIC-FACTORY-001 DEC-AF-28 - RESTRICTED SHA-256 is over the masked ledger form, never the raw value (Issue #1948)
git show --stat --format='%h %s' ba56671f20
  → ba56671f20 fix(agenticfactory/effect): make the right to send a compare-and-swap (F1, Issue #1948)
git show --stat --format='%h %s' 89770eaa47
  → 89770eaa47 fix(agenticfactory/ghpr): bound each gh read with its own time limit (F4, Issue #1948)
git show --stat --format='%h %s' e478302354
  → e478302354 fix(config): reject notify.inbox.max_id_bytes below 32 in the AF settings reader (Issue #1948)
git show --stat --format='%h %s' 6b945fd844
  → 6b945fd844 [AF-18] feat(runner): HostFacts, the production Facts adapter, and StoreOwner (Issue #1948)
git show --stat --format='%h %s' 2e2a714e3c
  → 2e2a714e3c [AF-18] feat(identity,runner): the read-only producer attribution lookup of the identity SPEC (Issue #1948)
git show --stat --format='%h %s' 5a97c569b4
  → 5a97c569b4 [AF-18] feat(core): one sanitizer, core.Sanitize, with categories and a refusing helper (Issue #1948)
git show --stat --format='%h %s' 17c1401810 546f7af570 11822055ad 3a3392ddc7 b9db20533b 269bfe502e a09a08e1df
  → 17c1401810 docs(spec): SPEC-AGENTIC-FACTORY-001 0.6.7 F2 - acceptance commands select only tests that exist at 4cdaba5d7e; uncovered cases marked unmet (Issue #1948)
  (같은 출력의 각 커밋 첫 줄:
   546f7af570 docs(spec): SPEC-AGENTIC-FACTORY-001 0.6.7 F2c - acceptance G.2 missing test names assigned to owning lanes (Issue #1948)
   11822055ad docs(spec): SPEC-AGENTIC-FACTORY-IDENTITY-001 acceptance - commands removed where no test or implementation exists at 4cdaba5d7e (Issue #1948)
   3a3392ddc7 docs(spec): SPEC-AGENTIC-FACTORY-001 acceptance - IDENTITY-related ACs split into behavior state and test state; lock kept per REQ-AF-011 is the intended state (Issue #1948)
   b9db20533b docs(spec): SPEC-AGENTIC-FACTORY-001 0.6.7 B8a - lane autonomy folded into REQ-AF-010/016, DEC-AF-27, proposed AC cases (Issue #1948)
   269bfe502e docs(spec): SPEC-AGENTIC-FACTORY-001 acceptance - runner.guard identity-unknown is a defect under repair (B10), not the intended lock (Issue #1948)
   a09a08e1df docs(spec): SPEC-AGENTIC-FACTORY-001 DEC-AF-28 - RESTRICTED SHA-256 is over the masked ledger form, never the raw value (Issue #1948))
git rev-parse --short HEAD
  → a09a08e1df
```

lane-17 확인용 SHA(`git show -s --format='%h %s' <sha>`): `ba56671f20`, `89770eaa47`, `e478302354`, `6b945fd844`, `2e2a714e3c`, `17c1401810`, `546f7af570`, `11822055ad`, `3a3392ddc7`, `269bfe502e`.
