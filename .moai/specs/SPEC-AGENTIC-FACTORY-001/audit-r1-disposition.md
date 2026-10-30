# 코드 감사 1회차 처분표 — 재감사 전 (B9)

- 작성일: 2026-10-30
- 감사 대상: `4cdaba5d7e` (`4cdaba5d7e6821aa9751d100cf75a328b4726704`), 판정 FAIL 0.67
- 판정서: `/Users/goos/MoAI/moai-adk-go/.moai/worktrees/agentic-factory-integration/.moai/reports/issue-1948/code-audit-4cdaba5d7e/verdict.md` (이 작성자가 읽음, 읽기만 함)
- 이 표가 말하는 시점(4회 갱신, 2026-10-30T00:46:48Z UTC, `date -u`로 읽음; 이 작성자의 HEAD `a544f54033`): 아래 각 lane 브랜치 끝을 `git branch --list … --format`으로 다시 읽은 상태다. 3회 갱신의 끝은 괄호 안에 남겼다.
  - lane-18 `WT-agentic-factory-core-claim-runner` @ `6b61a2cc07` (3회: `7196d0b5ea`)
  - lane-20 `WT-agentic-monitor-effect-reconcile` @ `cb55457588` (3회: `89770eaa47`). 지시에 적힌 `aa194b23c7`과 다르다 — 이 작성자가 읽은 끝은 lane-18 끝을 병합한 `cb55457588`("Merge lane-18 tip (sanitizer, PositionUnknown codes) into lane-20 tree")이고, 그 아래 `8a3920354f`(B6, 발견 항목을 부르지 않음), `aa194b23c7`(F1b)가 있다
  - lane-21 `WT-browser-select` @ `47d5abcb99` (3회: `988e6e2b80`)
  - lane-17 설정 `WT-agentic-factory-settings-keys` @ `e478302354` (바뀌지 않음)
  - lane-17 SPEC `WT-agentic-factory-contract` @ `a544f54033`
  - lane-17 template `WT-agentic-factory-templates` @ `47d8b782b7` (바뀌지 않음, `4cdaba5d7e` 뒤 커밋 0)
  - lane-22 `WT-agentic-factory-integration`(트리 `/Users/goos/MoAI/moai-adk-go/.moai/worktrees/agentic-factory-integration`) @ `70e28a5a78` (3회: `f80a6b2a1f`). lane-22의 브랜치가 이것이라는 것은 리더 전달이다. `4cdaba5d7e` 뒤 커밋 70개(`git rev-list --count`). first-parent 사슬은 12개다: `70e28a5a78` merge(lane-18 `e58e362926`, "audit F5 core side"), `548999a5c5`(직접 커밋, F19), `81a9214cf3`(직접 커밋), `3345a3a69e` merge(lane-17 `3ffd6a1e6f`), `6e97514ffe` merge(lane-17 `cc4621daab`), `17041f0922` merge(lane-21 `47d5abcb99`), `11feb35dd0` merge(lane-18 `9e34d6bb87`, "F3/F13"), 그리고 3회에 적은 `f80a6b2a1f`·`8616665dd7`·`a415460ab3`·`4588c3bdf3`·`5619ec1fb8`.
  - B10 확인(4회): `git log --format='%h %s' 4cdaba5d7e..WT-agentic-factory-integration --grep=B10`은 문서 커밋 두 개(`cc4621daab`, `269bfe502e`)만 출력했고, `--grep=Facts -i`는 문서 커밋들과 lane-18 어댑터 `6b945fd844`, 그리고 Facts 배선과 무관한 `cb190575b5`·`8b983ff2d5`·`84f09dab47`를 출력했다. Facts를 runner.guard에 잇는다고 말하는 커밋은 없다 — **B10 wiring commit: none found**.
- **재감사 대상**: 재감사가 읽을 트리는 통합 브랜치의 끝이다. 그래서 재감사에 중요한 것은 lane 브랜치에 커밋이 있느냐가 아니라 아래 표의 "integration 반영" 열이다. 통합 끝은 움직인다(다른 세션이 병합한다). 이 표의 "integration 반영" 값은 위에 적은 끝 `70e28a5a78`(2026-10-30T00:46:48Z UTC에 읽음) 기준이며, 재감사 때 끝이 다르면 다시 읽어야 한다.
- 갱신: 이 표는 세션이 깨어날 때마다 갱신한다(고정 일정 없음).
- 담당 lane 기준(2회): 리더의 배차 기준을 따른다(리더 전달). F3 = lane-18(claim 잠금)과 lane-20(notify), F5 = lane-18. F17·F18·F19·F23은 판정서의 담당(lane-22)을 그대로 두고 "통합 뒤 처리"로 표시한다. 그 밖의 행은 판정서 "담당" 줄 그대로다. 4회: F17·F18·F19·F23은 통합 브랜치에 커밋(`81a9214cf3`, `548999a5c5`)이 들어와 "통합 뒤 처리" 표시를 지우고 담당 lane-22만 남겼다.
- 판정 규칙: 커밋 제목이 발견 항목을 부르거나(예: `F1`) 고침을 분명히 말할 때만 그 커밋을 증거로 연결했고, 그때는 `git show --stat`으로 바뀐 파일이 맞는지 확인했다. 같은 영역을 만진 것만으로는 증거가 아니다. `수리됨`은 고침이 커밋에 들어 있다는 뜻이며, **고침이 동작한다는 뜻이 아니다**(아무것도 실행하지 않았다).
- 리더 전달(relayed, 이 작성자가 검증하지 않음): F1은 lane-20+lane-18, F2 대응표는 lane-17이 함, Facts 연결은 lane-22 B10에서 진행 중, HostFacts는 lane-18 `6b945fd844`.
- 리더 전달(4회, relayed, not verified by this author): `81a9214cf3`이 F17·F18·F23을 고친다; lane-19의 F5 core 검토에 P1 "운영 호출자 0"(OD-28 B로 다음 단계로 옮겨짐), P2 "SafeStop 검증 없음", P3 "fail-open"이 열려 있다; B6의 opaque-token 지적은 `1e28939419`로 답했다; 남은 정제 부채는 DEC-AF-29와 lane-18 S1/S2(S1/S2 내용은 이 작성자가 읽지 않음). lane-20 F1b·B6이 미커밋이라는 전달은 측정과 다르다 — F1b는 lane-20 브랜치에 `aa194b23c7`로, B6은 `8a3920354f`로 커밋되어 있다(통합에는 둘 다 없음).
- F2는 성격이 다른 네 부분이 있어 (가)~(라) 네 행으로 나눴다.

"integration 반영" 열: 수리 SHA 칸에 있는 SHA마다 `git branch --contains <sha> --list WT-agentic-factory-integration`을 한 번씩 돌려, 브랜치 이름이 출력되면 yes, 아무것도 출력되지 않으면 no로 적었다(4회, 끝 `70e28a5a78` 기준). 수리 SHA가 없는 행은 "—"다. 새 증거 규칙은 더하지 않았다.

| 항목 | 심각도 | 발견(판정서에서 줄임, 위치는 판정서 그대로) | 담당 lane(위 기준) | 수리 SHA | integration 반영 | 상태 | 증거 |
|---|---|---|---|---|---|---|---|
| F1 | BLOCKING | effect Executor가 동시 실행에서 같은 키를 두 번 송신 — `effect/intent.go:183-244`, `effect/docstore/docstore.go:46-61` | lane-20 + lane-18 | `ba56671f20`(lane-20). lane-18 몫은 커밋이 필요 없음 | yes | 수리됨 | 제목 "(F1, Issue #1948)"; `git show --stat`: `docstore.go`·`intent.go`·`intent_cas_test.go`·`putif_test.go` 등 6 파일 — 판정서가 짚은 두 파일 포함. 커밋 하나로 충분하다: lane-18 쪽 `claim.PutDoc`의 의미는 이미 맞았다(lane-19 확인, 리더 전달, 이 작성자는 읽지 않음). 동작 확인은 안 함. 메모(4회): 후속 F1b `aa194b23c7`("no unconditional intent writes, and no Failed while a send is in flight (F1b, Issue #1948)", stat `effect/intent.go`·`ctxledger.go`·`docstore.go`·`effect.go`·`labels.go` + 시험 2)는 lane-20 브랜치에만 있다 — integration 반영 no. 새 발견 행이 아니라 메모로 둠 |
| F2 (가) | BLOCKING | AC 명령이 고르는 시험 이름 47/48이 없음 — `acceptance.md` §C·§E | lane-17 | `17c1401810`, `546f7af570`, `11822055ad` | yes, yes, yes | 수리됨 | 판정서의 수정안 중 "acceptance.md의 명령을 실제 시험 이름으로 다시 쓴다"에 해당. 세 제목이 각각 명령 재작성(F2)·없는 이름의 lane 배정(F2c)·하위 SPEC 명령 제거를 말하고, stat은 두 SPEC의 `acceptance.md`만 바꿈. 기준 자체는 대부분 여전히 미충족(acceptance §G) |
| F2 (나) | BLOCKING | 병합 게이트가 production에서 항상 `AttributionLocked` | lane-17(계약) + lane-18 | `3a3392ddc7`(분류 기록) | yes | 계약상 의도 | REQ-AF-011 잠금(리더 판정, relayed). 제목 "lock kept per REQ-AF-011 is the intended state"; stat은 acceptance·spec·progress. 코드가 잠겨 있다는 관측은 감사 보고(리더 전달)뿐 |
| F2 (다) | BLOCKING | production `runner.guard`가 항상 `identity-unknown`으로 거절(`runner.Facts` 구현 0, `ghidentity.IssueFacts` 서명 불일치) | lane-18(어댑터) + lane-22(배선, B10) | `6b945fd844`(어댑터); lane-22 배선: 없음(커밋 미확인) | 어댑터 yes; 배선 — | 진행 중 | `6b945fd844` 제목 "HostFacts, the production Facts adapter"; stat `runner/hostfacts.go`·`hostfacts_test.go`. 어댑터는 통합 끝 `70e28a5a78`에 들어 있다(4회 `--contains` 확인). 배선은 lane-22 B10 진행 중(리더 전달). 4회 `--grep=B10`·`--grep=Facts -i`로 찾았을 때 Facts를 runner.guard에 잇는 커밋은 없다 — B10 wiring commit: none found. 분류 기록 `269bfe502e`(제목 "a defect under repair (B10), not the intended lock") |
| F2 (라) | BLOCKING | IDENTITY REQ-AFI-001~007의 production 구현 없음(`runner.Attributor` 구현 0) | lane-18 | `2e2a714e3c` | yes | 진행 중 | 제목 "the read-only producer attribution lookup of the identity SPEC"; stat `identity/identity.go`·`identity_test.go`·`runner/runner.go` 등 5 파일. 하위 SPEC 기준은 `4cdaba5d7e` 기준 전부 미충족이고 잠금 해제 조건(REQ-AF-011)도 아직이라 진행 중으로 둠 |
| F3 | SHOULD-FIX | 죽은 프로세스의 전역 잠금이 알림·확대를 영원히 멈춤 — `claim/claim.go:282-309`, `claim/docs.go:70-75`, `monitor/notify/notify.go:306-310`, `cli/factory_resident.go:216-222` | lane-18(claim 잠금) + lane-20(notify) — 리더 배차 기준 | `9e34d6bb87`(claim 잠금 쪽) | yes | 진행 중 | claim 잠금 쪽 수리: 제목 "store-wide locks can be inspected and reissued, and consecutive Busy is reported (F3, F13)", stat `claim/claim.go`·`claim/lock.go`·`claim/globallock_test.go`. 알림 쪽(`monitor/notify/notify.go:306-310`)과 `cli/factory_resident.go:216-222`는 F3을 부르는 커밋이 없다 — `81a9214cf3`은 `factory_resident.go`를 바꾸지만 제목이 F3도 이 고침도 말하지 않아 세지 않음. 그래서 진행 중. `5d383905f5`는 그대로 세지 않음 |
| F4 | SHOULD-FIX | 상주 루프의 gh 읽기에 호출별 시간 제한 없음 — `monitor/notify/ghpr/ghpr.go:65-73`, `resident/resident.go:238-245` | lane-20 | `89770eaa47` | yes | 수리됨 | 제목 "(F4, Issue #1948)"; stat `ghpr.go`·`ghpr_test.go`·`resident.go`·`ghexec_test.go`. 동작 확인은 안 함 |
| F5 | SHOULD-FIX | TakeOver·활성 판정이 호출자 불리언에만 기댐 — `claim/claim.go:25-87`, `runner/runner.go:669-719`, `claim/next.go:262-309` | lane-18 — 리더 배차 기준 | `e58e362926`; 후속 `6b61a2cc07`(A3b P2·P3) | yes; `6b61a2cc07` no | 진행 중 | `e58e362926` 제목 "TakeOver reads the owner's recorded process instead of the caller's word (F5)", stat `claim/presence.go`(+시험)·`claim/claim.go`·`runner/runner.go`·`runner/takeover_presence_test.go`·`mirror/mirror_test.go`; 통합 병합 `70e28a5a78`("audit F5 core side"). `6b61a2cc07` 제목 "a safe stop must be a record of this owner, and a damaged presence blocks a takeover (A3b P2, P3)"는 lane-18 브랜치에만 있다. lane-19 검토의 P1 "운영 호출자 0"(OD-28 B로 다음 단계), P2 "SafeStop 검증 없음", P3 "fail-open"은 열려 있다(relayed, not verified by this author). 그래서 수리됨으로 적지 않음 |
| F6 | SHOULD-FIX | runner가 전이표 E01–E27을 거치지 않음 — `runner/runner.go:186-193`, `:739` | lane-18 | 없음 | — | 미착수 | 해당 커밋 없음 |
| F7 | SHOULD-FIX | control ref 권위 미배선, 배선 시 시간 제한 없이 Issue 잠금 보유 — `claim/controlref.go:16-20,47`, `gitref/gitref.go:127-131` | lane-18 | 없음 | — | 미착수 | 해당 커밋 없음 |
| F8 | SHOULD-FIX | effect 의도 `LastError`가 가려지지 않은 채 저장, 가림 엔진 두 벌 — `effect/intent.go:223,229,233,235,237,264`, `ghexec.go:67-69`, `gitexec.go:53-55`, `core/redact.go:13-24`, `monitor/mask.go:12-22` | lane-20 | 없음 | — | 미착수 | lane-20 기록에 F8 커밋 없음. lane-18 `5a97c569b4`("one sanitizer, core.Sanitize")는 stat이 `core/redact.go`·`core/errors.go`·`core/sanitize_test.go`뿐이라 `LastError` 저장과 `monitor/mask.go`를 다루지 않아 이 항목의 증거로 세지 않음 |
| F9 | SHOULD-FIX | 설정 로더가 inbox `max_id_bytes` < 32를 받아 모든 알림 실패 — `config/agentic_factory_settings.go:343-364`, `monitor/notify/inbox.go:24-26,45-50` | lane-17 / lane-20 | `e478302354` | yes | 수리됨 | 제목 "reject notify.inbox.max_id_bytes below 32 in the AF settings reader"; stat `internal/config/agentic_factory_settings.go`·`_test.go`. 동작 확인은 안 함 |
| F10 | SHOULD-FIX | production lint 4건 — `gitref/gitref.go:388,389,391`, `effect/labels/githost.go:115` | lane-18, lane-20 | 없음 | — | 미착수 | 해당 커밋 없음 |
| F11 | SHOULD-FIX | `claim` 커버리지 84.4% < 85% — `internal/agenticfactory/claim` | lane-18 | 없음 | — | 미착수 | 커버리지를 말하는 커밋 없음(재지 않음) |
| F12 | MINOR | ghexec 허용 목록의 틈(`--web=true`, `-R`, 저장소 고정 없음) — `effect/ghexec/ghexec.go:140-157` | lane-20 | 없음 | — | 미착수 | 해당 커밋 없음 |
| F13 | MINOR | `ReissueLock` 확인-삭제 사이 TOCTOU — `claim/lock.go:76-90` | lane-18 | `9e34d6bb87` | yes | 수리됨 | 제목이 "(F3, F13)"을 부름; stat `claim/lock.go`·`claim/claim.go`·`claim/globallock_test.go` — 판정서가 짚은 `lock.go` 포함. 동작 확인은 안 함 |
| F14 | MINOR | `Release`가 epoch 미확인 — `claim/claim.go:681-696` | lane-18 | 없음 | — | 미착수 | 해당 커밋 없음 |
| F15 | MINOR | 인계 모델 두 벌 — `reconcile/handoff.go:62-102` 대 `runner/runner.go:484-490` | lane-18/20 | 없음 | — | 미착수 | 해당 커밋 없음 |
| F16 | MINOR | 브라우저 선택 키 위치·파서 중복 — `web/browser_select.go:74-95` | lane-21/17 | 없음 | — | 미착수 | lane-21 기록에 F16을 부르는 커밋 없음. `e196a35ff2`("AC-AF-020 cases … under the TestBrowserSelect_ names")는 시험 추가이며 파서 중복 고침을 말하지 않음. lane-17 설정·template 기록에도 없음 |
| F17 | MINOR | 상주 명령 시작 문구가 사실과 다름 — `cli/factory_resident.go:185` | lane-22 | `81a9214cf3` | yes | 수리됨 | 제목 "resident reports failed actions, states local writes, surfaces Serve errors; drop duplicate comment; …"은 F 번호를 부르지 않는다. "states local writes"를 이 발견(시작 문구가 사실과 다름)에 대응시킨 것은 이 작성자가 발견 문장에 비추어 읽은 것이고, `81a9214cf3`이 F17을 고친다는 것은 리더 전달(relayed, not verified by this author)이다. stat `cli/factory_resident.go`(+시험) 포함. 동작 확인은 안 함 |
| F18 | MINOR | 주석 중복 — `cli/codex_review_receipt.go:245-248` | lane-22 | `81a9214cf3` | yes | 수리됨 | 제목의 "drop duplicate comment"가 발견 문장과 맞고, stat에 `cli/codex_review_receipt.go`(2줄 삭제)가 있다. F18 대응은 리더 전달(relayed, not verified by this author) |
| F19 | MINOR(가설) | 위치 메모 키·지문의 빈틈 — `homestate/location_memo.go:104-113,170-198` | lane-22 | `548999a5c5` | yes | 수리됨 | 제목이 "audit F19"를 부름; stat `homestate/location_memo.go`·`location_memo_test.go`. 통합 브랜치의 직접 커밋. 동작 확인은 안 함 |
| F20 | MINOR(가설) | 자기 Issue 경로가 redirect·저장소 불일치를 보지 않음 — `core/authz.go:156-165` | lane-18 | 없음 | — | 미착수 | 해당 커밋 없음 |
| F21 | MINOR(가설) | `ReadyGate`가 첫 요청만 봄 — `claim/request.go:478-497` | lane-18 | 없음 | — | 미착수 | 해당 커밋 없음 |
| F22 | MINOR | 시험용 `FakeExecutor`가 production 파일에 있음 — `review/provider.go:136-149` | lane-20 | 없음 | — | 미착수 | 해당 커밋 없음 |
| F23 | MINOR | webhook `Serve` 오류가 종료 때까지 안 드러남 — `cli/factory_resident.go:199,232` | lane-22 | `81a9214cf3` | yes | 수리됨 | 제목의 "surfaces Serve errors"가 발견 문장과 맞고, stat에 `cli/factory_resident.go`가 있다. F23 대응은 리더 전달(relayed, not verified by this author). 동작 확인은 안 함 |
| F24 | MINOR(가설) | `gh pr checks --required`가 필수 검사 없는 저장소에서 매 pass `pr-read-error` — `ghpr.go:69-72` | lane-20 | 없음 | — | 미착수 | `89770eaa47`(F4)는 같은 파일이지만 제목이 시간 제한만 말해 세지 않음 |

지난 셈(1~3회, 기록으로 남김 — 지금 셈은 아래 4회): 셈(27행: F1, F2 네 행, F3~F24): 수리됨 4 · 진행 중 2 · 계약상 의도 1 · 미착수 20. 미착수 20 가운데 4행(F17·F18·F19·F23)은 "통합 뒤 처리" 표시다. 2회 갱신에서 상태가 바뀐 행은 없다(담당 lane과 증거 문구만 바뀜). 3회 갱신에서도 상태가 바뀐 행은 없다(integration 반영 열과 통합 끝 정정만). 수리 SHA가 있는 7행(F1, F2 네 행, F4, F9 — F2 (다)는 어댑터만)의 SHA는 모두 integration 반영 yes다.

셈(4회, 27행, 이 작성자가 표에서 다시 셈): **수리됨 9**(F1, F2 (가), F4, F9, F13, F17, F18, F19, F23) · **진행 중 4**(F2 (다), F2 (라), F3, F5) · **계약상 의도 1**(F2 (나)) · **미착수 13**(F6, F7, F8, F10, F11, F12, F14, F15, F16, F20, F21, F22, F24). 4회에서 바뀐 행: F3 미착수→진행 중, F5 미착수→진행 중, F13·F17·F18·F19·F23 미착수→수리됨. F17·F18·F19·F23의 "통합 뒤 처리" 표시는 통합 브랜치에 커밋이 들어와 지웠다. 미착수 13행은 `git log --format='%h %s' 4cdaba5d7e..<lane 브랜치>`(lane-18·lane-20·lane-21)와 통합 기록, 그리고 통합의 `-E --grep='F[0-9]+'` 결과에서 그 항목을 부르거나 고침을 말하는 커밋이 없어 그대로 두었다. 수리 SHA가 있는 행 중 integration 반영 no는 F5의 후속 `6b61a2cc07`과 F1 메모의 `aa194b23c7` 둘뿐이다. `수리됨`은 고침이 커밋에 있다는 뜻이며 동작한다는 뜻이 아니다.

- 지시에서 "F2 계약 쪽"으로 받은 일곱 커밋 가운데 `b9db20533b`(B8a, lane 자율 운영)와 `a09a08e1df`(DEC-AF-28, RESTRICTED SHA-256)는 제목이 F2를 부르지도 F2의 고침을 말하지도 않아, 위 규칙에 따라 F2 증거로 세지 않았다. 원장에는 남겼다.
- lane-22 브랜치(`WT-agentic-factory-integration`)의 끝이 `70e28a5a78`에서 움직이면 F2 (다)·F3·F5와 "integration 반영" 열 전체를 다시 읽는다.

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

4회 갱신(2026-10-30T00:46:48Z UTC, HEAD `a544f54033`)에서 실행한 명령(모두 따로 실행한 단순 읽기 전용 명령):

```
date -u +%Y-%m-%dT%H:%M:%SZ → 2026-10-30T00:46:48Z
git rev-parse --short HEAD → a544f54033
git branch --list <일곱 브랜치> --format='%(refname:short) %(objectname:short)' → WT-agentic-factory-contract a544f54033
  (나머지 줄: core-claim-runner 6b61a2cc07 / integration 70e28a5a78 / settings-keys e478302354 / templates 47d8b782b7 / monitor-effect-reconcile cb55457588 / browser-select 47d5abcb99)
git log --format='%h %s' --first-parent 4cdaba5d7e..WT-agentic-factory-integration → 70e28a5a78 merge(lane-18): TakeOver reads the owner process record (e58e362926, audit F5 core side) (Issue #1948)   (12줄)
git log --format='%h %s' 4cdaba5d7e..WT-agentic-factory-core-claim-runner → 6b61a2cc07 [AF-18] fix(agenticfactory/claim,runner): a safe stop must be a record of this owner, and a damaged presence blocks a takeover (A3b P2, P3)
git log --format='%h %s' 4cdaba5d7e..WT-agentic-monitor-effect-reconcile → cb55457588 Merge lane-18 tip (sanitizer, PositionUnknown codes) into lane-20 tree (Issue #1948)
git log --format='%h %s' 4cdaba5d7e..WT-browser-select → 47d5abcb99 test(web): AC-AF-020 (f) read-back - a saved selection reads back as the same value (Issue #1948, B7)
git log --format='%h %s' 4cdaba5d7e..WT-agentic-factory-integration --grep=B10 → cc4621daab docs(spec): SPEC-AGENTIC-FACTORY-001 B9 refresh - integration head f80a6b2a1f, integration-reach column, lane-21 browser test fact (Issue #1948)   (2줄, 둘 다 문서)
git log --format='%h %s' 4cdaba5d7e..WT-agentic-factory-integration --grep=Facts -i → 3345a3a69e merge(lane-17): SPEC lane-21 test facts, OD-27 candidate (WT-agentic-factory-contract 3ffd6a1e6f) (Issue #1948)   (11줄, Facts 배선 커밋 없음)
git log --format='%h %s' -E --grep='F[0-9]+' 4cdaba5d7e..WT-agentic-factory-integration → 70e28a5a78 merge(lane-18): TakeOver reads the owner process record (e58e362926, audit F5 core side) (Issue #1948)   (16줄; 부르는 F 번호: F1, F2(문서), F3, F4, F5, F9, F13, F19)
git rev-list --count 4cdaba5d7e..WT-agentic-factory-integration → 70
git show --stat --format='%h %s' 81a9214cf3 → 81a9214cf3 fix(cli): resident reports failed actions, states local writes, surfaces Serve errors; drop duplicate comment; factorymsg test reaches the slow-git path through an alias (Issue #1948)
git show --stat --format='%h %s' 548999a5c5 → 548999a5c5 fix(homestate): location memo fingerprints a linked worktree's git directory and keys on GIT_DISCOVERY_ACROSS_FILESYSTEM (Issue #1948, audit F19)
git show --stat --format='%h %s' 9e34d6bb87 → 9e34d6bb87 [AF-18] fix(agenticfactory/claim): store-wide locks can be inspected and reissued, and consecutive Busy is reported (F3, F13)
git show --stat --format='%h %s' e58e362926 → e58e362926 [AF-18] feat(agenticfactory/claim,runner): TakeOver reads the owner's recorded process instead of the caller's word (F5)
git show --stat --format='%h %s' aa194b23c7 → aa194b23c7 fix(agenticfactory/effect): no unconditional intent writes, and no Failed while a send is in flight (F1b, Issue #1948)
git show --stat --format='%h %s' 6b61a2cc07 → 6b61a2cc07 [AF-18] fix(agenticfactory/claim,runner): a safe stop must be a record of this owner, and a damaged presence blocks a takeover (A3b P2, P3)
git branch --contains <sha> --list WT-agentic-factory-integration → "+ WT-agentic-factory-integration" for:
  ba56671f20 17c1401810 546f7af570 11822055ad 3a3392ddc7 6b945fd844 2e2a714e3c 9e34d6bb87 89770eaa47 e58e362926 e478302354 548999a5c5 81a9214cf3 1e28939419
git branch --contains <sha> --list WT-agentic-factory-integration → 출력 없음 for: aa194b23c7, 6b61a2cc07
```

lane-17 확인용 SHA(`git show -s --format='%h %s' <sha>`): 4회 추가 `70e28a5a78`, `548999a5c5`, `81a9214cf3`, `9e34d6bb87`, `e58e362926`, `6b61a2cc07`, `aa194b23c7`, `cb55457588`, `8a3920354f`; 3회 추가 `f80a6b2a1f`, `8616665dd7`, `a415460ab3`, `4588c3bdf3`, `5619ec1fb8`; 기존: `ba56671f20`, `89770eaa47`, `e478302354`, `6b945fd844`, `2e2a714e3c`, `17c1401810`, `546f7af570`, `11822055ad`, `3a3392ddc7`, `269bfe502e`.
