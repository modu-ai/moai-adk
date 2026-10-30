# 코드 감사 1회차 처분표 — 재감사 전 (B9)

- 작성일: 2026-10-30
- 감사 대상: `4cdaba5d7e` (`4cdaba5d7e6821aa9751d100cf75a328b4726704`), 판정 FAIL 0.67
- 판정서: `/Users/goos/MoAI/moai-adk-go/.moai/worktrees/agentic-factory-integration/.moai/reports/issue-1948/code-audit-4cdaba5d7e/verdict.md` (이 작성자가 읽음, 읽기만 함)
- 이 표가 말하는 시점(3회 갱신, 2026-10-29T23:53:20Z UTC, `date -u`로 읽음): 아래 각 lane 브랜치 끝을 `git branch --list … --format`으로 다시 읽은 상태다.
  - lane-18 `WT-agentic-factory-core-claim-runner` @ `7196d0b5ea`
  - lane-20 `WT-agentic-monitor-effect-reconcile` @ `89770eaa47`
  - lane-21 `WT-browser-select` @ `988e6e2b80` (1회 작성 때는 `e196a35ff2`. 새 커밋 `988e6e2b80`은 `TestScenarioSuite` 하네스이며 이 표의 어느 발견 항목도 부르지 않는다)
  - lane-17 설정 `WT-agentic-factory-settings-keys` @ `e478302354`
  - lane-17 SPEC `WT-agentic-factory-contract` @ `7632d31597`
  - lane-17 template `WT-agentic-factory-templates` @ `47d8b782b7` (`4cdaba5d7e` 뒤 커밋 0)
  - lane-22 `WT-agentic-factory-integration`(트리 `/Users/goos/MoAI/moai-adk-go/.moai/worktrees/agentic-factory-integration`) @ `f80a6b2a1f`. lane-22의 브랜치가 이것이라는 것은 리더 전달이다. 2회 갱신에서 적은 "끝 `4cdaba5d7e`, 그 뒤 커밋 0"은 더 이상 맞지 않는다(정정). 이 작성자가 다시 읽은 first-parent 사슬은 병합 다섯 개다: `f80a6b2a1f`(lane-17 contract), `8616665dd7`(lane-17 F9 settings `e478302354`), `a415460ab3`(lane-20 `89770eaa47`), `4588c3bdf3`(lane-18, 제목은 `1e28939419`), `5619ec1fb8`(lane-21, 제목은 `3ada6181ed`). `f80a6b2a1f`의 제목은 contract `b9db20533b`를 말하지만, `git branch --contains 7632d31597 --list WT-agentic-factory-integration`은 그 브랜치 이름을 출력했다(contract 끝 `7632d31597`도 통합에 들어 있음).
  - B10 확인: `git log --format='%h %s' 4cdaba5d7e..WT-agentic-factory-integration` 전체 목록에서 B10을 부르는 커밋은 문서 커밋 `269bfe502e`(분류 기록) 하나이고, HostFacts를 부르는 커밋은 lane-18 어댑터 `6b945fd844` 하나다. Facts 배선(runner.guard에 어댑터를 잇는 일)을 말하는 커밋은 없다 — **B10 배선 커밋 없음**. 어댑터 `6b945fd844`는 이제 통합에 들어 있다(`git merge-base --is-ancestor 6b945fd844 WT-agentic-factory-integration`가 출력 없이 끝났고 도구가 오류 종료를 보고하지 않음; `git branch --contains 6b945fd844 --list WT-agentic-factory-integration`가 브랜치 이름을 출력함).
- **재감사 대상**: 재감사가 읽을 트리는 통합 브랜치의 끝이다. 그래서 재감사에 중요한 것은 lane 브랜치에 커밋이 있느냐가 아니라 아래 표의 "integration 반영" 열이다. 통합 끝은 움직인다(다른 세션이 병합한다). 이 표의 "integration 반영" 값은 위에 적은 끝 `f80a6b2a1f`(2026-10-29T23:53:20Z UTC에 읽음) 기준이며, 재감사 때 끝이 다르면 다시 읽어야 한다.
- 갱신: 이 표는 세션이 깨어날 때마다 갱신한다(고정 일정 없음).
- 담당 lane 기준(2회): 리더의 배차 기준을 따른다(리더 전달). F3 = lane-18(claim 잠금)과 lane-20(notify), F5 = lane-18. F17·F18·F19·F23은 판정서의 담당(lane-22)을 그대로 두고 "통합 뒤 처리"로 표시한다. 그 밖의 행은 판정서 "담당" 줄 그대로다.
- 판정 규칙: 커밋 제목이 발견 항목을 부르거나(예: `F1`) 고침을 분명히 말할 때만 그 커밋을 증거로 연결했고, 그때는 `git show --stat`으로 바뀐 파일이 맞는지 확인했다. 같은 영역을 만진 것만으로는 증거가 아니다. `수리됨`은 고침이 커밋에 들어 있다는 뜻이며, **고침이 동작한다는 뜻이 아니다**(아무것도 실행하지 않았다).
- 리더 전달(relayed, 이 작성자가 검증하지 않음): F1은 lane-20+lane-18, F2 대응표는 lane-17이 함, Facts 연결은 lane-22 B10에서 진행 중, HostFacts는 lane-18 `6b945fd844`.
- F2는 성격이 다른 네 부분이 있어 (가)~(라) 네 행으로 나눴다.

"integration 반영" 열: 수리 SHA 칸에 이미 있는 SHA마다 `git branch --contains <sha> --list WT-agentic-factory-integration`을 한 번씩 돌려, 브랜치 이름이 출력되면 yes, 아무것도 출력되지 않으면 no로 적었다(끝 `f80a6b2a1f` 기준). 수리 SHA가 없는 행은 "—"다. 새 증거 규칙은 더하지 않았다.

| 항목 | 심각도 | 발견(판정서에서 줄임, 위치는 판정서 그대로) | 담당 lane(위 기준) | 수리 SHA | integration 반영 | 상태 | 증거 |
|---|---|---|---|---|---|---|---|
| F1 | BLOCKING | effect Executor가 동시 실행에서 같은 키를 두 번 송신 — `effect/intent.go:183-244`, `effect/docstore/docstore.go:46-61` | lane-20 + lane-18 | `ba56671f20`(lane-20). lane-18 몫은 커밋이 필요 없음 | yes | 수리됨 | 제목 "(F1, Issue #1948)"; `git show --stat`: `docstore.go`·`intent.go`·`intent_cas_test.go`·`putif_test.go` 등 6 파일 — 판정서가 짚은 두 파일 포함. 커밋 하나로 충분하다: lane-18 쪽 `claim.PutDoc`의 의미는 이미 맞았다(lane-19 확인, 리더 전달, 이 작성자는 읽지 않음). 동작 확인은 안 함 |
| F2 (가) | BLOCKING | AC 명령이 고르는 시험 이름 47/48이 없음 — `acceptance.md` §C·§E | lane-17 | `17c1401810`, `546f7af570`, `11822055ad` | yes, yes, yes | 수리됨 | 판정서의 수정안 중 "acceptance.md의 명령을 실제 시험 이름으로 다시 쓴다"에 해당. 세 제목이 각각 명령 재작성(F2)·없는 이름의 lane 배정(F2c)·하위 SPEC 명령 제거를 말하고, stat은 두 SPEC의 `acceptance.md`만 바꿈. 기준 자체는 대부분 여전히 미충족(acceptance §G) |
| F2 (나) | BLOCKING | 병합 게이트가 production에서 항상 `AttributionLocked` | lane-17(계약) + lane-18 | `3a3392ddc7`(분류 기록) | yes | 계약상 의도 | REQ-AF-011 잠금(리더 판정, relayed). 제목 "lock kept per REQ-AF-011 is the intended state"; stat은 acceptance·spec·progress. 코드가 잠겨 있다는 관측은 감사 보고(리더 전달)뿐 |
| F2 (다) | BLOCKING | production `runner.guard`가 항상 `identity-unknown`으로 거절(`runner.Facts` 구현 0, `ghidentity.IssueFacts` 서명 불일치) | lane-18(어댑터) + lane-22(배선, B10) | `6b945fd844`(어댑터); lane-22 배선: 없음(커밋 미확인) | 어댑터 yes; 배선 — | 진행 중 | `6b945fd844` 제목 "HostFacts, the production Facts adapter"; stat `runner/hostfacts.go`·`hostfacts_test.go`. 어댑터는 이제 통합 끝 `f80a6b2a1f`에 들어 있다(병합 `4588c3bdf3` 경로로 보임, `--is-ancestor`·`--contains` 둘 다 확인). 배선은 lane-22 B10 진행 중(리더 전달). 통합 브랜치의 `4cdaba5d7e` 뒤 전체 기록에 Facts 배선을 말하는 커밋은 없다 — B10 배선 커밋 없음. 분류 기록 `269bfe502e`(제목 "a defect under repair (B10), not the intended lock") |
| F2 (라) | BLOCKING | IDENTITY REQ-AFI-001~007의 production 구현 없음(`runner.Attributor` 구현 0) | lane-18 | `2e2a714e3c` | yes | 진행 중 | 제목 "the read-only producer attribution lookup of the identity SPEC"; stat `identity/identity.go`·`identity_test.go`·`runner/runner.go` 등 5 파일. 하위 SPEC 기준은 `4cdaba5d7e` 기준 전부 미충족이고 잠금 해제 조건(REQ-AF-011)도 아직이라 진행 중으로 둠 |
| F3 | SHOULD-FIX | 죽은 프로세스의 전역 잠금이 알림·확대를 영원히 멈춤 — `claim/claim.go:282-309`, `claim/docs.go:70-75`, `monitor/notify/notify.go:306-310`, `cli/factory_resident.go:216-222` | lane-18(claim 잠금) + lane-20(notify) — 리더 배차 기준 | 없음 | — | 미착수 | F3을 부르거나 이 고침을 말하는 커밋 없음. `5d383905f5`("documents lock wait covers the resident writers' queue")는 같은 영역이지만 죽은 잠금 재발급·확대를 말하지 않아 증거로 세지 않음 |
| F4 | SHOULD-FIX | 상주 루프의 gh 읽기에 호출별 시간 제한 없음 — `monitor/notify/ghpr/ghpr.go:65-73`, `resident/resident.go:238-245` | lane-20 | `89770eaa47` | yes | 수리됨 | 제목 "(F4, Issue #1948)"; stat `ghpr.go`·`ghpr_test.go`·`resident.go`·`ghexec_test.go`. 동작 확인은 안 함 |
| F5 | SHOULD-FIX | TakeOver·활성 판정이 호출자 불리언에만 기댐 — `claim/claim.go:25-87`, `runner/runner.go:669-719`, `claim/next.go:262-309` | lane-18 — 리더 배차 기준 | 없음 | — | 미착수 | F5를 부르는 커밋 없음. 병합 `0137f1fd88` 제목의 "takeover"는 고침을 말하지 않아 세지 않음 |
| F6 | SHOULD-FIX | runner가 전이표 E01–E27을 거치지 않음 — `runner/runner.go:186-193`, `:739` | lane-18 | 없음 | — | 미착수 | 해당 커밋 없음 |
| F7 | SHOULD-FIX | control ref 권위 미배선, 배선 시 시간 제한 없이 Issue 잠금 보유 — `claim/controlref.go:16-20,47`, `gitref/gitref.go:127-131` | lane-18 | 없음 | — | 미착수 | 해당 커밋 없음 |
| F8 | SHOULD-FIX | effect 의도 `LastError`가 가려지지 않은 채 저장, 가림 엔진 두 벌 — `effect/intent.go:223,229,233,235,237,264`, `ghexec.go:67-69`, `gitexec.go:53-55`, `core/redact.go:13-24`, `monitor/mask.go:12-22` | lane-20 | 없음 | — | 미착수 | lane-20 기록에 F8 커밋 없음. lane-18 `5a97c569b4`("one sanitizer, core.Sanitize")는 stat이 `core/redact.go`·`core/errors.go`·`core/sanitize_test.go`뿐이라 `LastError` 저장과 `monitor/mask.go`를 다루지 않아 이 항목의 증거로 세지 않음 |
| F9 | SHOULD-FIX | 설정 로더가 inbox `max_id_bytes` < 32를 받아 모든 알림 실패 — `config/agentic_factory_settings.go:343-364`, `monitor/notify/inbox.go:24-26,45-50` | lane-17 / lane-20 | `e478302354` | yes | 수리됨 | 제목 "reject notify.inbox.max_id_bytes below 32 in the AF settings reader"; stat `internal/config/agentic_factory_settings.go`·`_test.go`. 동작 확인은 안 함 |
| F10 | SHOULD-FIX | production lint 4건 — `gitref/gitref.go:388,389,391`, `effect/labels/githost.go:115` | lane-18, lane-20 | 없음 | — | 미착수 | 해당 커밋 없음 |
| F11 | SHOULD-FIX | `claim` 커버리지 84.4% < 85% — `internal/agenticfactory/claim` | lane-18 | 없음 | — | 미착수 | 커버리지를 말하는 커밋 없음(재지 않음) |
| F12 | MINOR | ghexec 허용 목록의 틈(`--web=true`, `-R`, 저장소 고정 없음) — `effect/ghexec/ghexec.go:140-157` | lane-20 | 없음 | — | 미착수 | 해당 커밋 없음 |
| F13 | MINOR | `ReissueLock` 확인-삭제 사이 TOCTOU — `claim/lock.go:76-90` | lane-18 | 없음 | — | 미착수 | 해당 커밋 없음 |
| F14 | MINOR | `Release`가 epoch 미확인 — `claim/claim.go:681-696` | lane-18 | 없음 | — | 미착수 | 해당 커밋 없음 |
| F15 | MINOR | 인계 모델 두 벌 — `reconcile/handoff.go:62-102` 대 `runner/runner.go:484-490` | lane-18/20 | 없음 | — | 미착수 | 해당 커밋 없음 |
| F16 | MINOR | 브라우저 선택 키 위치·파서 중복 — `web/browser_select.go:74-95` | lane-21/17 | 없음 | — | 미착수 | lane-21 기록에 F16을 부르는 커밋 없음. `e196a35ff2`("AC-AF-020 cases … under the TestBrowserSelect_ names")는 시험 추가이며 파서 중복 고침을 말하지 않음. lane-17 설정·template 기록에도 없음 |
| F17 | MINOR | 상주 명령 시작 문구가 사실과 다름 — `cli/factory_resident.go:185` | lane-22 — 통합 뒤 처리 | 없음 | — | 미착수 | lane-22 브랜치 `WT-agentic-factory-integration` 끝 `f80a6b2a1f`. `4cdaba5d7e` 뒤 전체 기록(다른 lane의 병합과 그 커밋들)에 이 항목을 부르거나 고침을 말하는 커밋 없음. 읽은 다른 브랜치들에도 없음 |
| F18 | MINOR | 주석 중복 — `cli/codex_review_receipt.go:245-248` | lane-22 — 통합 뒤 처리 | 없음 | — | 미착수 | F17과 같음 |
| F19 | MINOR(가설) | 위치 메모 키·지문의 빈틈 — `homestate/location_memo.go:104-113,170-198` | lane-22 — 통합 뒤 처리 | 없음 | — | 미착수 | F17과 같음 |
| F20 | MINOR(가설) | 자기 Issue 경로가 redirect·저장소 불일치를 보지 않음 — `core/authz.go:156-165` | lane-18 | 없음 | — | 미착수 | 해당 커밋 없음 |
| F21 | MINOR(가설) | `ReadyGate`가 첫 요청만 봄 — `claim/request.go:478-497` | lane-18 | 없음 | — | 미착수 | 해당 커밋 없음 |
| F22 | MINOR | 시험용 `FakeExecutor`가 production 파일에 있음 — `review/provider.go:136-149` | lane-20 | 없음 | — | 미착수 | 해당 커밋 없음 |
| F23 | MINOR | webhook `Serve` 오류가 종료 때까지 안 드러남 — `cli/factory_resident.go:199,232` | lane-22 — 통합 뒤 처리 | 없음 | — | 미착수 | F17과 같음 |
| F24 | MINOR(가설) | `gh pr checks --required`가 필수 검사 없는 저장소에서 매 pass `pr-read-error` — `ghpr.go:69-72` | lane-20 | 없음 | — | 미착수 | `89770eaa47`(F4)는 같은 파일이지만 제목이 시간 제한만 말해 세지 않음 |

셈(27행: F1, F2 네 행, F3~F24): 수리됨 4 · 진행 중 2 · 계약상 의도 1 · 미착수 20. 미착수 20 가운데 4행(F17·F18·F19·F23)은 "통합 뒤 처리" 표시다. 2회 갱신에서 상태가 바뀐 행은 없다(담당 lane과 증거 문구만 바뀜). 3회 갱신에서도 상태가 바뀐 행은 없다(integration 반영 열과 통합 끝 정정만). 수리 SHA가 있는 7행(F1, F2 네 행, F4, F9 — F2 (다)는 어댑터만)의 SHA는 모두 integration 반영 yes다.

- 지시에서 "F2 계약 쪽"으로 받은 일곱 커밋 가운데 `b9db20533b`(B8a, lane 자율 운영)와 `a09a08e1df`(DEC-AF-28, RESTRICTED SHA-256)는 제목이 F2를 부르지도 F2의 고침을 말하지도 않아, 위 규칙에 따라 F2 증거로 세지 않았다. 원장에는 남겼다.
- lane-22 브랜치(`WT-agentic-factory-integration`)의 끝이 `f80a6b2a1f`에서 움직이면 F2 (다)와 F17·F18·F19·F23, 그리고 "integration 반영" 열 전체를 다시 읽는다.

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

2회 갱신(2026-10-30, HEAD `7632d31597`)에서 다시 실행한 명령:

```
git log --format='%h %s' 4cdaba5d7e..WT-agentic-factory-integration
  → 출력 없음
git branch --list WT-agentic-factory-integration WT-agentic-factory-core-claim-runner WT-agentic-monitor-effect-reconcile WT-browser-select WT-agentic-factory-settings-keys WT-agentic-factory-templates WT-agentic-factory-contract --format='%(refname:short) %(objectname:short)'
  → WT-agentic-factory-contract 7632d31597
  (나머지 줄: WT-agentic-factory-core-claim-runner 7196d0b5ea / WT-agentic-factory-integration 4cdaba5d7e / WT-agentic-factory-settings-keys e478302354 / WT-agentic-factory-templates 47d8b782b7 / WT-agentic-monitor-effect-reconcile 89770eaa47 / WT-browser-select 988e6e2b80)
git log --format='%h %s' 4cdaba5d7e..WT-agentic-factory-core-claim-runner
  → 7196d0b5ea [AF-18] fix(agenticfactory/mirror): post a RESTRICTED record when a ledger is refused (design 4.4)
git log --format='%h %s' 4cdaba5d7e..WT-agentic-monitor-effect-reconcile
  → 89770eaa47 fix(agenticfactory/ghpr): bound each gh read with its own time limit (F4, Issue #1948)
git log --format='%h %s' 4cdaba5d7e..WT-browser-select
  → 988e6e2b80 test(e2e): TestScenarioSuite harness under af_e2e - 01 real, 02-07 and the attribution preflight stated as unmet (Issue #1948, B7)
git log --format='%h %s' 4cdaba5d7e..WT-agentic-factory-settings-keys
  → e478302354 fix(config): reject notify.inbox.max_id_bytes below 32 in the AF settings reader (Issue #1948)
git log --format='%h %s' 4cdaba5d7e..WT-agentic-factory-templates
  → 출력 없음
git log --format='%h %s' -3 WT-agentic-factory-contract
  → 7632d31597 docs(spec): SPEC-AGENTIC-FACTORY-001 B9 - audit round 1 disposition table with repair SHAs read from lane branch logs (Issue #1948)
git show -s --format='%h %s' e196a35ff2
  → e196a35ff2 test(browser): AC-AF-020 cases (d), (e) and the resolver half of (f) under the TestBrowserSelect_ names (Issue #1948, B7)
```

3회 갱신(2026-10-29T23:53:20Z UTC, HEAD `7632d31597`)에서 실행한 명령. 2회 갱신 원장의 "`4cdaba5d7e..WT-agentic-factory-integration` → 출력 없음"은 그때 관측한 그대로 남기며, 지금은 맞지 않는다:

```
git rev-parse --short WT-agentic-factory-integration
  → f80a6b2a1f
git log --format='%h %s' --first-parent 4cdaba5d7e..WT-agentic-factory-integration
  → f80a6b2a1f merge(lane-17): SPEC-AGENTIC-FACTORY-001 B8a, acceptance G.2 (WT-agentic-factory-contract b9db20533b) (Issue #1948)
  (나머지 네 줄: 8616665dd7 merge(lane-17) … e478302354, audit F9 / a415460ab3 merge(lane-20) … 89770eaa47 / 4588c3bdf3 merge(lane-18) … 1e28939419 / 5619ec1fb8 merge(lane-21) … 3ada6181ed)
date -u +%Y-%m-%dT%H:%M:%SZ
  → 2026-10-29T23:53:20Z
git log --format='%h %s' 4cdaba5d7e..WT-agentic-factory-integration
  → f80a6b2a1f merge(lane-17): SPEC-AGENTIC-FACTORY-001 B8a, acceptance G.2 (WT-agentic-factory-contract b9db20533b) (Issue #1948)
  (전체 58줄, `git log --format='%h' 4cdaba5d7e..WT-agentic-factory-integration | wc -l` → 58. B10을 부르는 줄: 269bfe502e 하나(문서). HostFacts를 부르는 줄: 6b945fd844 하나(어댑터). Facts 배선을 말하는 줄: 없음)
git merge-base --is-ancestor 6b945fd844 WT-agentic-factory-integration
  → 출력 없음(도구가 오류 종료를 보고하지 않음 = exit 0으로 읽음)
git branch --list WT-agentic-factory-integration WT-agentic-factory-core-claim-runner WT-agentic-monitor-effect-reconcile WT-browser-select WT-agentic-factory-settings-keys WT-agentic-factory-templates WT-agentic-factory-contract --format='%(refname:short) %(objectname:short)'
  → WT-agentic-factory-contract 7632d31597
  (나머지 줄: core-claim-runner 7196d0b5ea / integration f80a6b2a1f / settings-keys e478302354 / templates 47d8b782b7 / monitor-effect-reconcile 89770eaa47 / browser-select 988e6e2b80)
git branch --contains ba56671f20 --list WT-agentic-factory-integration   → + WT-agentic-factory-integration
git branch --contains 17c1401810 --list WT-agentic-factory-integration   → + WT-agentic-factory-integration
git branch --contains 546f7af570 --list WT-agentic-factory-integration   → + WT-agentic-factory-integration
git branch --contains 11822055ad --list WT-agentic-factory-integration   → + WT-agentic-factory-integration
git branch --contains 3a3392ddc7 --list WT-agentic-factory-integration   → + WT-agentic-factory-integration
git branch --contains 6b945fd844 --list WT-agentic-factory-integration   → + WT-agentic-factory-integration
git branch --contains 2e2a714e3c --list WT-agentic-factory-integration   → + WT-agentic-factory-integration
git branch --contains 89770eaa47 --list WT-agentic-factory-integration   → + WT-agentic-factory-integration
git branch --contains e478302354 --list WT-agentic-factory-integration   → + WT-agentic-factory-integration
git branch --contains 7632d31597 --list WT-agentic-factory-integration   → + WT-agentic-factory-integration
```

`echo $?`는 따로 돌리지 않았다: 이 환경에서는 Bash 호출마다 새 프로세스라, 별도 호출의 `$?`는 앞 명령의 종료 코드가 아니다. 대신 `git branch --contains`의 출력으로 판정했다.

lane-17 확인용 SHA(`git show -s --format='%h %s' <sha>`): 3회 추가 `f80a6b2a1f`, `8616665dd7`, `a415460ab3`, `4588c3bdf3`, `5619ec1fb8`; 기존: `ba56671f20`, `89770eaa47`, `e478302354`, `6b945fd844`, `2e2a714e3c`, `17c1401810`, `546f7af570`, `11822055ad`, `3a3392ddc7`, `269bfe502e`.
