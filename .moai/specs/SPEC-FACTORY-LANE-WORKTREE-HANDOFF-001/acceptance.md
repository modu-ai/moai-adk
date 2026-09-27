---
id: SPEC-FACTORY-LANE-WORKTREE-HANDOFF-001
document: acceptance
created: 2026-09-22
updated: 2026-09-24
author: manager-spec
card: t1082
module: "internal/factorymsg"
---

# Acceptance — SPEC-FACTORY-LANE-WORKTREE-HANDOFF-001

## Acceptance policy

- AC-FLH-001..011과 AC-FLH-014..020은 모두 MUST-PASS다. AC-FLH-012·013은 0.5.11에서 이 SPEC 범위 밖(§ Exclusions — 카드 t1145로 이관)으로 옮겼으므로 이 SPEC의 MUST-PASS 대상이 아니다.
- Unit/fixture evidence는 해당 named contract만 증명한다.
- 모든 GREEN command는 정확히 한 parent test의 `Action=pass`, 전체 child/subtest/package의 `Action=fail` 및 `Action=skip` 0건, 전체 log의 `NOT_RUN` 0건을 요구한다.
- `go test -run`의 empty match, package setup failure, missing log, mock-only LIVE, production 경로를 우회한 fixture 준비(직접 peer 등록, 수동 DB seed)는 vacuous PASS가 아니라 FAIL이다. 피시험 대상 자체인 production store entry point — AC-FLH-018/019가 hook·launcher 경로와 같은 함수를 racer의 production `Open` 핸들에서 부르는 것 — 는 fixture 준비가 아니므로 이 금지 대상이 아니다. 그 호출에는 test 전용 등록 경로나 변형을 두지 않는다.
- 각 log는 실행 직전 `git rev-parse HEAD`와 built binary identity를 verdict에 귀속해야 한다.

## Summary and traceability

| AC | Requirements | Named test | Observable outcome |
|---|---|---|---|
| AC-FLH-001 | REQ-FLH-001, REQ-FLH-003 | `TestFactoryLaneHandoffAdmissionFailClosed` | active turn, permission wait, interrupt, dirty source, untrusted cwd, path/branch collision, stale/multiple reservation이 모두 side effect 0의 NACK다. |
| AC-FLH-002 | REQ-FLH-004, REQ-FLH-005 | `TestFactoryLaneHandoffDevelopPinAndTraceability` | local develop pin에서 launcher L1 WT가 생기고 path/branch/card/SPEC traceability가 일치한다. |
| AC-FLH-003 | REQ-FLH-002, REQ-FLH-006, REQ-FLH-010 | `TestFactoryLaneHandoffInteractiveStateMachine` | idle interactive 경로가 사용자 `/cd` 뒤 다음 정상 turn의 SessionStart/cwd/branch evidence로만 BOUND되며 empty turn은 0이다. BOUND 뒤 tombstone된 이전 session의 SessionStart는 기존 endpoint-replaced 안내문을 내고 행을 바꾸지 않는다. 현재 endpoint가 launch-pending 행일 때도 같은 session의 SessionStart 안내문은 같은 상태의 UserPromptSubmit 안내문(빈 session 렌더링)과 바이트 동일하다. |
| AC-FLH-004 | REQ-FLH-002, REQ-FLH-007 | `TestFactoryLaneHandoffHeadlessAppServerStateMachine` | idle headless 경로가 official fork/start 반환 ID와 controller provenance로 직접 BOUND하고 active turn을 거부한다. |
| AC-FLH-005 | REQ-FLH-008 | `TestFactoryLaneHandoffAtomicModeEvidenceRebind` | interactive SessionStart와 headless RPC-result rebind가 각각 all-or-nothing이다. |
| AC-FLH-006 | REQ-FLH-009 | `TestFactoryLaneHandoffDispatchAfterBound` | body는 BOUND 뒤 current generation에만 release된다. |
| AC-FLH-007 | REQ-FLH-001, REQ-FLH-010 | `TestFactoryLaneHandoffStaleEndpointRejected` | old/stale endpoint send/read/ACK가 current endpoint hint와 함께 거부되고, tombstone된 session UUID는 turn 등록 경로와 launcher resume 경로(`BindLaunchPending`) 모두에서 generation과 무관하게 `STALE_ENDPOINT`로 거부되며 launch-pending 행은 launcher 등록이 commit한 그대로 남는다. 이미 BOUND인 행에서도 tombstone된 UUID의 `BindLaunchPending`은 `STALE_ENDPOINT`로 거부되고 아무것도 쓰지 않는다. |
| AC-FLH-008 | REQ-FLH-009 | `TestFactoryLaneHandoffDuplicateAndSameLaneRedispatch` | idempotency 기준과 무관하게, 같은 recipient generation 안의 같은 key duplicate는 한 번만 실행되고, BOUND 뒤 재전송은 새 key를 쓰며, 이전 generation은 stale NACK다. |
| AC-FLH-009 | REQ-FLH-011 | `TestFactoryLaneHandoffCrashRecovery` | create/rebind/receipt crash points가 deterministic resume/finalize/NACK로 복구된다. |
| AC-FLH-010 | REQ-FLH-011 | `TestFactoryLaneHandoffAbandonedWorktreeRecovery` | dirty/unmerged/unknown-owner WT가 ABANDONED로 보존되고 자동 삭제되지 않는다. |
| AC-FLH-011 | REQ-FLH-008, REQ-FLH-009, REQ-FLH-013 | `TestFactoryLaneHandoffNoPreBoundWrites` | BOUND 전 code/commit/task ACK 0, wrong cwd write 0, primary branch switch 0, message loss 0이다. |
| AC-FLH-014 | REQ-FLH-006, REQ-FLH-007, REQ-FLH-012 | `TestFactoryLaneHandoffNoPrivateControl` | slash automation/tmux/private socket/model-cd/Desktop emulation 호출이 0이다. |
| AC-FLH-015 | REQ-FLH-015 | `TestFactoryLaneHandoffT1074Compatibility` | 기존 broker/roster/receipt/catalog가 유지되고 새 broker/daemon/store가 없다. |
| AC-FLH-016 | REQ-FLH-004 | `TestFactoryLaneHandoffCreationBaseDriftRejected` | t1082에서 실제 관측된 main→develop creation-base drift mutant가 BASE_DRIFT로 fail closed한다. |
| AC-FLH-017 | REQ-FLH-016 | `TestFactoryLaneHandoffLaunchPendingSourceNack` | launch-pending lane에 대한 handoff 요청이 `ENDPOINT_LAUNCH_PENDING` NACK이며 reservation/tombstone/worktree/app-server/endpoint 변화가 0이다. |
| AC-FLH-018 | REQ-FLH-015, REQ-FLH-017 | `TestFactoryLaneHandoffRebindVsLaunchBindRace` | 같은 slot에서 launcher provisional registration/bind와 handoff rebind를 동시에 구동해도 bound owner 1개, 패자의 이름 있는 거부, 단조 generation, orphan launch-pending 0이다. launcher 등록이 먼저 commit하면 그 commit에서 handoff가 `NACK`/`STALE_GENERATION`이 되고, SessionStart 선행 순서에서도 UserPromptSubmit이 provisional 행을 결합한다. |
| AC-FLH-019 | REQ-FLH-016, REQ-FLH-017, REQ-FLH-018 | `TestFactoryLaneHandoffRebindVsUserPromptRegisterRace` | 비종결 handoff 동안 같은 owner·새 session UUID의 UserPromptSubmit registration이 별도 핸들에서 경합해도 tombstone·receipt 없는 endpoint 이동 0, 단조 generation, 패자의 이름 있는 거부다. 미commit reservation을 쥔 상대가 있을 때 registration·launcher 등록이 handoff 상태를 자기 transaction 안에서 읽음을 판별하고, 반대로 미commit launcher 등록을 쥔 상대가 있을 때 reservation이 source 행을 자기 transaction 안에서 읽음(순서 (vii))을 판별한다. |
| AC-FLH-020 | REQ-FLH-011 | `TestFactoryLaneHandoffOperatorAbandon` | 막힌 비종결 handoff를 operator 명령 `moai factory handoff abandon-lane --slot <slot>`이 source owner가 current가 아닐 때만 한 transaction에서 `ABANDONED`/`OPERATOR_ABANDONED`로 종결하고, live owner면 `SOURCE_OWNER_LIVE`로 무쓰기 거부하며, worktree를 보존하고 BOUND·tombstone·receipt·release를 쓰지 않으며, 종결 뒤 UserPromptSubmit 등록이 t1074 의미로 복귀한다. |

## RED-now ledger

다음 exact selectors를 subject tree `bf39a539d97f49edf3b11517ee7c982239c60df3`에서 실행했다. 각 command는 stdout 0 bytes, exit 1이었다. 따라서 현재는 16개 criterion 모두 RED다. Test 존재 자체는 GREEN이 아니며, 각 절의 exact Go/JQ gate까지 통과해야 한다.

AC-FLH-012·013과 두 AC의 공통 gate-quality 행은 0.5.11에서 § Exclusions로 옮겼다. 아래 표에는 이 SPEC 범위에 남은 criterion만 있다.

| AC | Exact RED-now command | Observed stdout | Exit |
|---|---|---|---:|
| AC-FLH-001 | `rg -n -F 'func TestFactoryLaneHandoffAdmissionFailClosed(' internal --glob '*_test.go'` | `<empty>` | 1 |
| AC-FLH-002 | `rg -n -F 'func TestFactoryLaneHandoffDevelopPinAndTraceability(' internal --glob '*_test.go'` | `<empty>` | 1 |
| AC-FLH-003 | `rg -n -F 'func TestFactoryLaneHandoffInteractiveStateMachine(' internal --glob '*_test.go'` | `<empty>` | 1 |
| AC-FLH-004 | `rg -n -F 'func TestFactoryLaneHandoffHeadlessAppServerStateMachine(' internal --glob '*_test.go'` | `<empty>` | 1 |
| AC-FLH-005 | `rg -n -F 'func TestFactoryLaneHandoffAtomicModeEvidenceRebind(' internal --glob '*_test.go'` | `<empty>` | 1 |
| AC-FLH-006 | `rg -n -F 'func TestFactoryLaneHandoffDispatchAfterBound(' internal --glob '*_test.go'` | `<empty>` | 1 |
| AC-FLH-007 | `rg -n -F 'func TestFactoryLaneHandoffStaleEndpointRejected(' internal --glob '*_test.go'` | `<empty>` | 1 |
| AC-FLH-008 | `rg -n -F 'func TestFactoryLaneHandoffDuplicateAndSameLaneRedispatch(' internal --glob '*_test.go'` | `<empty>` | 1 |
| AC-FLH-009 | `rg -n -F 'func TestFactoryLaneHandoffCrashRecovery(' internal --glob '*_test.go'` | `<empty>` | 1 |
| AC-FLH-010 | `rg -n -F 'func TestFactoryLaneHandoffAbandonedWorktreeRecovery(' internal --glob '*_test.go'` | `<empty>` | 1 |
| AC-FLH-011 | `rg -n -F 'func TestFactoryLaneHandoffNoPreBoundWrites(' internal --glob '*_test.go'` | `<empty>` | 1 |
| AC-FLH-014 | `rg -n -F 'func TestFactoryLaneHandoffNoPrivateControl(' internal --glob '*_test.go'` | `<empty>` | 1 |
| AC-FLH-015 | `rg -n -F 'func TestFactoryLaneHandoffT1074Compatibility(' internal --glob '*_test.go'` | `<empty>` | 1 |
| AC-FLH-016 | `rg -n -F 'func TestFactoryLaneHandoffCreationBaseDriftRejected(' internal --glob '*_test.go'` | `<empty>` | 1 |
| AC-FLH-017 | `rg -n -F 'func TestFactoryLaneHandoffLaunchPendingSourceNack(' internal --glob '*_test.go'` | `<empty>` | 1 |
| AC-FLH-018 | `rg -n -F 'func TestFactoryLaneHandoffRebindVsLaunchBindRace(' internal --glob '*_test.go'` | `<empty>` | 1 |
| AC-FLH-019 | `rg -n -F 'func TestFactoryLaneHandoffRebindVsUserPromptRegisterRace(' internal --glob '*_test.go'` | `<empty>` | 1 |
| AC-FLH-020 | `rg -n -F 'func TestFactoryLaneHandoffOperatorAbandon(' internal --glob '*_test.go'` | `<empty>` | 1 |

AC-FLH-017/018 두 행은 위 subject tree가 아니라 t1074 흡수 후 HEAD `1487f97a0`에서 2026-09-23에 측정했다. AC-FLH-019 행은 HEAD `d28ed9aa4`에서 2026-09-23에 측정했다. 세 행 모두 RED 이유는 named test 부재다. 0.5.0 개정으로 기준이 바뀐 AC-FLH-018/019 두 행은 HEAD `2e3ec3d4b`에서 2026-09-23에 같은 command로 재측정했고, 둘 다 stdout 0 bytes, exit 1이었다. 새 named test는 추가되지 않았다.

AC-FLH-020 행은 0.5.5 개정에서 HEAD `730139bd6`에서 2026-09-23에 측정했다. stdout 0 bytes, exit 1이며 RED 이유는 named test 부재다.

0.5.1 개정으로 AC-FLH-019에 추가된 순서 (vii)(reservation의 source 행 판독 위치 판별)는 run phase의 첫 RED다. run phase는 named test `TestFactoryLaneHandoffRebindVsUserPromptRegisterRace`를 순서 (vii)부터 작성해, reservation 구현이 존재하기 전 그 순서가 실패함을 먼저 관측한다. 위 표의 AC-FLH-019 행은 named test 부재로 여전히 RED이며, 이 개정은 새 named test를 추가하지 않는다.

0.5.6 개정으로 AC-FLH-007에 추가된 tombstone 재등록 다리(UserPromptSubmit 경로의 `STALE_ENDPOINT` 거부)는 named test `TestFactoryLaneHandoffStaleEndpointRejected`가 아직 단언하지 않는다. 이 다리의 RED 이유는 named test 부재가 아니라, named test 본문에 turn registration 호출이 없다는 것이다. HEAD `e2c2d33b1f21c3bf018c8f586d735e640202013a`에서 2026-09-23에 측정했다.

| 측정 | Command | Observed stdout | Exit |
|---|---|---|---:|
| AC-FLH-007 tombstone 재등록 다리 (RED) | `awk '/^func TestFactoryLaneHandoffStaleEndpointRejected\(/{f=1} f&&/RegisterPeer/{print FILENAME":"FNR": "$0; n++} f&&/^}/{exit} END{exit n?0:1}' internal/factorymsg/handoff_bind_test.go` | `<empty>` | 1 |
| 앵커 대조 (같은 범위 진입 확인) | 위 command에서 `RegisterPeer`를 `requireStale`로 바꾼 것 | `handoff_bind_test.go:545`, `:565` 두 줄 | 0 |
| 양성 대조 (같은 프로브가 발화하는가) | 위 command에서 함수명을 `TestFactoryLaneHandoffRebindVsUserPromptRegisterRace`로, 파일을 `internal/factorymsg/lane_handoff_race_test.go`로 바꾼 것 | `lane_handoff_race_test.go:316`, `:372`, `:404`, `:439`, `:476`, `:552` 여섯 줄 | 0 |

앵커 대조는 프로브가 named test 본문에 실제로 들어갔음을, 양성 대조는 같은 프로브가 `RegisterPeer` 호출을 찾을 수 있음을 보인다. 따라서 첫 행의 빈 출력은 측정 실패가 아니라 부재다. 이 다리는 run phase에서 named test에 추가되며, 명령·jq 게이트는 바뀌지 않는다.

0.5.7 개정으로 AC-FLH-007에 추가된 launcher resume 다리(`RegisterLaunchPending` 뒤 tombstone된 UUID의 `BindLaunchPending`이 `STALE_ENDPOINT`로 거부됨)도 named test가 아직 단언하지 않는다. RED 이유는 named test 본문에 `BindLaunchPending` 호출이 없다는 것이다. HEAD `113daeb8331d1590cc6dfbcf2216c4de89651609`에서 2026-09-23에 측정했다.

| 측정 | Command | Observed stdout | Exit |
|---|---|---|---:|
| AC-FLH-007 launcher resume 다리 (RED) | `awk '/^func TestFactoryLaneHandoffStaleEndpointRejected\(/{f=1} f&&/BindLaunchPending/{print FILENAME":"FNR": "$0; n++} f&&/^}/{exit} END{exit n?0:1}' internal/factorymsg/handoff_bind_test.go` | `<empty>` | 1 |
| 앵커 대조 (같은 범위 진입 확인) | 위 command에서 `BindLaunchPending`을 `requireStale`로 바꾼 것 | `handoff_bind_test.go:545`, `:565` 두 줄 | 0 |
| 양성 대조 (같은 프로브가 발화하는가) | 위 command에서 함수명을 같은 파일의 `seedBoundPeer`로 바꾼 것 | `handoff_bind_test.go:46` 한 줄 | 0 |

첫 행의 빈 출력은 앵커 대조가 발화하므로 측정 실패가 아니라 부재다. 행동 수준에서는 델타 plan-audit이 임시 탐침 `TestZZAuditProbeTombstoneLauncherRevival`(실행 후 삭제)로, 같은 HEAD에서 tombstone된 `src-uuid`가 `RegisterLaunchPending`(gen 4)과 `BindLaunchPending`(gen 5)을 거쳐 현재 endpoint로 결합되고 send도 성공함을 관측했다(`.moai/reports/t1082/plan-audit-delta-v056.md` 항목 5). 그 관측은 감사자의 측정이며 이 개정에서 다시 실행하지 않았다. 이 다리는 M4에서 named test에 추가되며, 명령·jq 게이트는 바뀌지 않는다.

0.5.8 개정으로 두 다리가 더해졌다. AC-FLH-007의 BOUND 행 다리(이미 BOUND인 행에서 tombstone된 UUID의 `BindLaunchPending`이 `STALE_ENDPOINT`)와 AC-FLH-003의 hook 안내문 다리(BOUND 뒤 이전 session의 SessionStart가 endpoint-replaced 안내문을 냄)다. 두 다리 모두 named test는 있지만 본문이 아직 단언하지 않는다. 아래 측정은 모두 HEAD `03390136aef7e6f2a7d30a37769c686189c4d355`에서 2026-09-23에 이 개정 작성자가 실행했다.

| 측정 | Command | Observed stdout | Exit |
|---|---|---|---:|
| AC-FLH-007 BOUND 행 다리 (RED) | `awk '/^func TestFactoryLaneHandoffStaleEndpointRejected\(/{f=1} f&&/BindLaunchPending/{print FILENAME":"FNR": "$0; n++} f&&/^}/{exit} END{exit n?0:1}' internal/factorymsg/handoff_bind_test.go` | `<empty>` | 1 |
| 앵커 대조 (같은 범위 진입 확인) | 위 command에서 `BindLaunchPending`을 `requireStale`로 바꾼 것 | `handoff_bind_test.go:545`, `:565` 두 줄 | 0 |
| 양성 대조 (같은 프로브가 발화하는가) | 위 command에서 함수명을 같은 파일의 `seedBoundPeer`로 바꾼 것 | `handoff_bind_test.go:46` 한 줄 | 0 |
| AC-FLH-003 hook 안내문 다리 (RED) | `awk '/^func TestFactoryLaneHandoffInteractiveStateMachine\(/{f=1} f&&/factory endpoint replaced/{print FILENAME":"FNR": "$0; n++} f&&/^}/{exit} END{exit n?0:1}' internal/hook/factory_handoff_bind_test.go` | `<empty>` | 1 |
| 앵커 대조 (같은 범위 진입 확인) | 위 command에서 `factory endpoint replaced`를 `sessionStart\("src-uuid"`로 바꾼 것 | `factory_handoff_bind_test.go:198` 한 줄(BOUND 전 호출) | 0 |
| 양성 대조 (같은 프로브가 발화하는가) | 위 command에서 함수명을 `TestFactoryUserPromptRegistrationSurfacesHandoffRefusals`로, 파일을 `internal/hook/factory_handoff_race_test.go`로 바꾼 것 | `factory_handoff_race_test.go:757` 한 줄 | 0 |

BOUND 행 다리의 RED 명령은 0.5.7 launcher 다리 행과 같은 명령이다. 두 다리의 RED 이유가 같기 때문이다(named test 본문에 `BindLaunchPending` 호출이 없다). 이 HEAD에서 다시 실행했고 결과도 같았다. hook 다리의 앵커가 잡은 `:198`은 BOUND 전의 SessionStart이며, 그 줄은 안내문이 빈 값이기를 요구한다. 양성 대조가 발화하므로 첫 행의 빈 출력은 측정 실패가 아니라 부재다.

행동 수준 관측: 이 개정 작성자가 같은 HEAD에서 임시 탐침 두 개(`internal/factorymsg/zz_t1082_v058_probe_test.go`, `internal/hook/zz_t1082_v058_probe_test.go`)를 실행한 뒤 삭제했다. 삭제 뒤 `git status --short`는 비어 있었다. 출력 발췌:

```
$ go test ./internal/factorymsg -run '^TestZZT1082V058Probe$' -count=1 -v
TOMB gen=2
BOUND-ROW BIND src-uuid ok=false err=<nil> got.session=post-cd-uuid got.gen=3
BOUND-ROW countsEqual=true rowEqual=true
PENDING gen=4 ; LAUNCHER-BIND in.gen=1 ok=true err=<nil> bound.gen=5
SEND from src-uuid gen=5 err=<nil>
SEND to src-uuid gen=5 err=<nil>
SEND from src-uuid gen=2(tomb) err=stale or unregistered peer: STALE_ENDPOINT; lane lane-1 is current at src-uuid generation 5
ok  github.com/modu-ai/moai-adk/internal/factorymsg 0.369s
$ go test ./internal/hook -run '^TestZZT1082V058HookProbe$' -count=1 -v
BOUND SessionStart src-uuid notice="" rowEqual=true
BOUND UserPromptSubmit src-uuid notice="factory endpoint replaced STALE_ENDPOINT: slot=lane-1 is current at post-cd-uuid generation 3; this session receives no factory messages" rowEqual=true
ok  github.com/modu-ai/moai-adk/internal/hook 1.307s
```

해석: BOUND 행에서 tombstone된 UUID의 bind는 오류 없이 조용히 넘어가고(`ok=false err=<nil>`), hook의 SessionStart 안내문은 빈 값이다. 두 다리 모두 HEAD에서 RED다. 입력 `Generation`을 1로 둔 launcher bind는 결합에 성공하고(`bound.gen=5`), 그 identity(gen 5)로 보낸 송신과 그 identity로 받는 송신도 성공한다. 따라서 N2에서 고정한 송신 다리 두 개는 HEAD에서 FAIL한다. tombstone generation(gen 2)의 송신자는 HEAD에서도 거부되므로 부활을 판별하지 못한다. 같은 상태에서 UserPromptSubmit은 이미 안내문을 내므로, hook 다리의 기대 문자열은 기존 렌더링을 그대로 쓴다. 명령·jq 게이트는 두 AC 모두 바뀌지 않는다. AC-FLH-003의 명령은 이미 `./internal/hook`을 돌린다.

0.5.9 개정은 두 가지다. AC-FLH-007에서는 BOUND 행 다리를 launcher 다리 앞으로 옮겨 서술 순서를 상태 진행 순서(BOUND → launcher 등록 → launch-pending owner 비-current)와 맞췄다. 의미·명령·jq 게이트·named test는 그대로이므로 새 측정이 없다. AC-FLH-003에서는 hook 안내문 다리에, 현재 endpoint가 launch-pending 행일 때(redirect session이 빈 값)의 SessionStart 안내문 단언을 더했다. 이 단언도 named test에 아직 없다. RED 이유는 named test 본문에 `RegisterLaunchPending` 호출이 없다는 것이다. 아래 측정은 모두 HEAD `ec33efa6a627d95c1aea9f43aed588eae939c83b`에서 2026-09-23에 이 개정 작성자가 실행했다.

| 측정 | Command | Observed stdout | Exit |
|---|---|---|---:|
| AC-FLH-003 launch-pending 안내문 다리 (RED) | `awk '/^func TestFactoryLaneHandoffInteractiveStateMachine\(/{f=1} f&&/RegisterLaunchPending/{print FILENAME":"FNR": "$0; n++} f&&/^}/{exit} END{exit n?0:1}' internal/hook/factory_handoff_bind_test.go` | `<empty>` | 1 |
| 앵커 대조 (같은 범위 진입 확인) | 위 command에서 `RegisterLaunchPending`을 `sessionStart\("src-uuid"`로 바꾼 것 | `factory_handoff_bind_test.go:198` 한 줄(BOUND 전 호출) | 0 |
| 양성 대조 (같은 프로브가 발화하는가) | 위 command에서 함수명을 `TestFactoryLaneHandoffRebindVsLaunchBindRace`로, 파일을 `internal/hook/factory_handoff_race_test.go`로 바꾼 것 | `factory_handoff_race_test.go:421`, `:483`, `:523`, `:562`, `:600` 다섯 줄 | 0 |

양성 대조가 발화하므로 첫 행의 빈 출력은 측정 실패가 아니라 부재다.

기대 문자열은 같은 상태의 UserPromptSubmit 안내문과의 바이트 동일로 정했다. 두 경로가 모두 `factoryHandoffRegistrationNotice`(`internal/hook/factory_handoff_bind.go:92`)를 거친다는 것이 design §2.1(N4)의 설계이고, UserPromptSubmit 쪽은 HEAD에서 이미 빈 session 렌더링을 낸다. 문자열을 리터럴로만 고정하면 generation 값과 공백 개수를 테스트가 따로 맞춰야 하지만, 같은 상태의 다른 경로 출력과 비교하면 공유 렌더러와 store redirect의 비공개 토큰 가림이 함께 고정된다. 행동 수준 관측: 이 개정 작성자가 같은 HEAD에서 임시 탐침 `internal/hook/zz_t1082_v059_probe_test.go`를 실행한 뒤 삭제했다. 삭제 뒤 `git status --short`는 비어 있었다. 출력 발췌:

```
$ go test ./internal/hook -run '^TestZZT1082V059HookProbe$' -count=1 -v
helper state=dead
TOMB gen=2
PENDING gen=4
LP UserPromptSubmit src-uuid notice="factory endpoint replaced STALE_ENDPOINT: slot=lane-1 is current at  generation 4; this session receives no factory messages"
LP SessionStart src-uuid notice="factory messaging bound: run=run-v059-probe slot=lane-1 generation=5; messages arrive at turn boundaries, not idle wake"
LANE after session=src-uuid gen=5
ok  github.com/modu-ai/moai-adk/internal/hook 2.109s
```

해석: 탐침은 BOUND owner를 살아 있는 helper 프로세스로 결합한 뒤 그 프로세스를 종료·회수해 current가 아니게 만들고, hook이 스스로 해석하는 owner identity로 `RegisterLaunchPending`을 실행했다. 이 상태에서 UserPromptSubmit은 이미 빈 session 안내문을 내지만, SessionStart는 tombstone된 `src-uuid`를 launch-pending generation + 1(gen 5)로 결합하고 `factory messaging bound`를 낸다. 따라서 이 다리는 HEAD에서 FAIL한다. 명령·jq 게이트는 바뀌지 않는다.

## Exact acceptance gates

### AC-FLH-001 — Admission fails closed

**Given** active-turn, permission-wait, interrupted, dirty-source, symlink-escape cwd, occupied branch/path, stale reservation, and same-lane in-flight fixtures, **when** reserve is attempted, **then** each returns its exact NACK and creates no worktree, app-server request, endpoint mutation, or dispatch release. **And** given a `NACK`ed handoff for the same card whose target worktree is clean, on the reserved `WT-*` branch, and at the fresh develop pin, with the lane's cwd already that target, a fresh reserve is admitted and reaches `WT_READY` by reconstruction with a materializer call count of 0, while the same setup with the target HEAD moved off the fresh pin returns `BASE_DRIFT`.

```bash
unset ANTHROPIC_BASE_URL ANTHROPIC_API_KEY ANTHROPIC_AUTH_TOKEN Z_AI_API_KEY MOAI_KANBAN_BACKEND MOAI_FACTORY_WORKERS && mkdir -p .moai/reports/t1082 && MOAI_HOME=/tmp/t1082-ac01-home GOCACHE=/tmp/t1082-ac01-cache go test -json ./internal/factorymsg ./internal/cli -run '^TestFactoryLaneHandoffAdmissionFailClosed$' -count=1 -timeout=90s > .moai/reports/t1082/ac01.jsonl && jq -se '([.[]|select(.Action=="pass" and .Test=="TestFactoryLaneHandoffAdmissionFailClosed")]|length)==1 and ([.[]|select(.Action=="fail" or .Action=="skip")]|length)==0 and ([.[]|select((.Output//"")|contains("NOT_RUN"))]|length)==0' .moai/reports/t1082/ac01.jsonl
```

Expected final output: `true`; otherwise FAIL.

### AC-FLH-002 — Develop pin, L1 creation, and traceability

**Given** a local develop ref, card/SPEC/slug, and clean canonical primary fixture, **when** handoff creates the target, **then** it calls the existing MoAI materializer once, target HEAD equals the pre-create pin, directory is `<primary>/.claude/worktrees/<card-id>`, branch is `WT-<slug>`, and dispatch/commit/report trace fields are exact without primary branch movement.

```bash
unset ANTHROPIC_BASE_URL ANTHROPIC_API_KEY ANTHROPIC_AUTH_TOKEN Z_AI_API_KEY MOAI_KANBAN_BACKEND MOAI_FACTORY_WORKERS && mkdir -p .moai/reports/t1082 && MOAI_HOME=/tmp/t1082-ac02-home GOCACHE=/tmp/t1082-ac02-cache go test -json ./internal/cli -run '^TestFactoryLaneHandoffDevelopPinAndTraceability$' -count=1 -timeout=90s > .moai/reports/t1082/ac02.jsonl && jq -se '([.[]|select(.Action=="pass" and .Test=="TestFactoryLaneHandoffDevelopPinAndTraceability")]|length)==1 and ([.[]|select(.Action=="fail" or .Action=="skip")]|length)==0 and ([.[]|select((.Output//"")|contains("NOT_RUN"))]|length)==0' .moai/reports/t1082/ac02.jsonl
```

Expected final output: `true`; otherwise FAIL.

### AC-FLH-003 — Interactive user-driven relocation

**Given** an idle interactive Codex lane in `WT_READY`, **when** the handoff emits `/cd <target>` guidance, the user actually executes it, and the user's next normal turn produces matching SessionStart, **then** only verified target cwd/branch/HEAD/new UUID advances to BOUND; before that it remains `SWITCH_PENDING_INTERACTIVE`, no automatic slash execution or empty model turn occurs, and mismatched evidence NACKs. **And** after that BOUND, **when** the SessionStart of the tombstoned source session (`src-uuid`, resumed at the primary checkout) runs through the production hook entry (`registerFactorySessionStartPeer` → `registerFactoryHookPeer`, `internal/hook/factory_messages.go:53`), **then** its additional-context notice is the existing endpoint-replaced rendering of `factoryHandoffRegistrationNotice` (`internal/hook/factory_handoff_bind.go:92`) — byte-equal to the notice that the UserPromptSubmit registration of the same session (`registerFactoryUserPromptPeer`) returns in the same state, i.e. `factory endpoint replaced STALE_ENDPOINT: slot=<lane slot> is current at post-cd-uuid generation <handoff-bound generation>; this session receives no factory messages` — and contains neither `factory messaging degraded` nor `factory messaging bound`; the lane endpoint row (session UUID, generation, PID, process-start), the tombstone rows, and the BOUND receipt rows are identical before and after that SessionStart. A hook that surfaced the store refusal as `factory messaging degraded: <err>` fails this leg, and so does a store that reads the tombstone only after finding a launch-pending row, because the bind on the BOUND row then returns without an error and the hook emits the empty string. **And** given that BOUND state after the handoff-bound owner is made not current under the t1074 live-owner rule and `RegisterLaunchPending` has committed a launch-pending row for the lane's slot carrying the PID and process-start the production hook resolves for itself (`session.ResolveOwnerPID` and `homestate.ProbeProcessIdentity` inside `registerFactoryHookPeer`), so that the lane's current endpoint is that launch-pending row, **when** the UserPromptSubmit registration of the tombstoned `src-uuid` (`registerFactoryUserPromptPeer`) runs first and the SessionStart of the same session (`registerFactorySessionStartPeer`) runs next, **then** the SessionStart notice is byte-equal to the UserPromptSubmit notice returned in that state — the endpoint-replaced rendering of `factoryHandoffRegistrationNotice` with the redirect session redacted to the empty string, i.e. `factory endpoint replaced STALE_ENDPOINT: slot=<lane slot> is current at  generation <launch-pending generation>; this session receives no factory messages` (two spaces between `at` and `generation`) — and contains neither `factory messaging bound` nor `factory messaging degraded`; the lane endpoint row (private token, generation, PID, process-start), the tombstone rows, and the BOUND receipt rows are identical before the UserPromptSubmit call and after the SessionStart call. A SessionStart that binds the tombstoned session at the launch-pending generation + 1 and returns `factory messaging bound` fails this leg, and so does one that renders its own text instead of the shared notice, or whose redirect carries the private launch-pending token instead of the redacted empty session, because the UserPromptSubmit notice it is compared with renders the empty session.

```bash
unset ANTHROPIC_BASE_URL ANTHROPIC_API_KEY ANTHROPIC_AUTH_TOKEN Z_AI_API_KEY MOAI_KANBAN_BACKEND MOAI_FACTORY_WORKERS && mkdir -p .moai/reports/t1082 && MOAI_HOME=/tmp/t1082-ac03-home GOCACHE=/tmp/t1082-ac03-cache go test -json ./internal/factorymsg ./internal/hook ./internal/cli -run '^TestFactoryLaneHandoffInteractiveStateMachine$' -count=1 -timeout=90s > .moai/reports/t1082/ac03.jsonl && jq -se '([.[]|select(.Action=="pass" and .Test=="TestFactoryLaneHandoffInteractiveStateMachine")]|length)==1 and ([.[]|select(.Action=="fail" or .Action=="skip")]|length)==0 and ([.[]|select((.Output//"")|contains("NOT_RUN"))]|length)==0' .moai/reports/t1082/ac03.jsonl
```

Expected final output: `true`; otherwise FAIL.

### AC-FLH-004 — Headless official app-server relocation

**Given** idle stored-thread, idle no-history, active-turn, permission-wait, and interrupt fixtures, **when** headless relocation runs, **then** stored history uses `thread/fork(cwd)` and observes new ID/`forkedFromId`/`thread/started`, no-history uses `thread/start(cwd)`, controller readback verifies target cwd/branch/HEAD before direct BOUND, active/wait/interrupt NACK, and no SessionStart wait, empty `turn/start`, or `turn/steer` request exists.

```bash
unset ANTHROPIC_BASE_URL ANTHROPIC_API_KEY ANTHROPIC_AUTH_TOKEN Z_AI_API_KEY MOAI_KANBAN_BACKEND MOAI_FACTORY_WORKERS && mkdir -p .moai/reports/t1082 && MOAI_HOME=/tmp/t1082-ac04-home GOCACHE=/tmp/t1082-ac04-cache go test -json ./internal/cli -run '^TestFactoryLaneHandoffHeadlessAppServerStateMachine$' -count=1 -timeout=90s > .moai/reports/t1082/ac04.jsonl && jq -se '([.[]|select(.Action=="pass" and .Test=="TestFactoryLaneHandoffHeadlessAppServerStateMachine")]|length)==1 and ([.[]|select(.Action=="fail" or .Action=="skip")]|length)==0 and ([.[]|select((.Output//"")|contains("NOT_RUN"))]|length)==0' .moai/reports/t1082/ac04.jsonl
```

Expected final output: `true`; otherwise FAIL.

### AC-FLH-005 — Atomic mode-evidence rebind

**Given** matching and mismatching interactive next-turn SessionStart and headless official RPC-result/controller-readback evidence plus injected failures at every write boundary, **when** mode-specific rebind runs, **then** peer replacement, old tombstone, BOUND state, receipt, and release marker are either all visible or all absent.

```bash
unset ANTHROPIC_BASE_URL ANTHROPIC_API_KEY ANTHROPIC_AUTH_TOKEN Z_AI_API_KEY MOAI_KANBAN_BACKEND MOAI_FACTORY_WORKERS && mkdir -p .moai/reports/t1082 && MOAI_HOME=/tmp/t1082-ac05-home GOCACHE=/tmp/t1082-ac05-cache go test -json ./internal/factorymsg ./internal/hook ./internal/cli -run '^TestFactoryLaneHandoffAtomicModeEvidenceRebind$' -count=1 -timeout=90s > .moai/reports/t1082/ac05.jsonl && jq -se '([.[]|select(.Action=="pass" and .Test=="TestFactoryLaneHandoffAtomicModeEvidenceRebind")]|length)==1 and ([.[]|select(.Action=="fail" or .Action=="skip")]|length)==0 and ([.[]|select((.Output//"")|contains("NOT_RUN"))]|length)==0' .moai/reports/t1082/ac05.jsonl
```

Expected final output: `true`; otherwise FAIL.

### AC-FLH-006 — Dispatch body only after BOUND

**Given** one dispatch sent before relocation and one during SWITCH_PENDING, **when** old/new endpoints list and read them around rebind, **then** metadata remains durable, body claim/read is denied before BOUND, and both bodies release exactly once to the new generation after BOUND.

```bash
unset ANTHROPIC_BASE_URL ANTHROPIC_API_KEY ANTHROPIC_AUTH_TOKEN Z_AI_API_KEY MOAI_KANBAN_BACKEND MOAI_FACTORY_WORKERS && mkdir -p .moai/reports/t1082 && MOAI_HOME=/tmp/t1082-ac06-home GOCACHE=/tmp/t1082-ac06-cache go test -json ./internal/factorymsg -run '^TestFactoryLaneHandoffDispatchAfterBound$' -count=1 -timeout=90s > .moai/reports/t1082/ac06.jsonl && jq -se '([.[]|select(.Action=="pass" and .Test=="TestFactoryLaneHandoffDispatchAfterBound")]|length)==1 and ([.[]|select(.Action=="fail" or .Action=="skip")]|length)==0 and ([.[]|select((.Output//"")|contains("NOT_RUN"))]|length)==0' .moai/reports/t1082/ac06.jsonl
```

Expected final output: `true`; otherwise FAIL.

### AC-FLH-007 — Old endpoint tombstone and stale rejection

**Given** old UUID/generation, stale reservation, stale claim token, restarted process, and current endpoint, **when** send/read/receipt/ACK execute, **then** every stale operation NACKs without mutation, returns current non-secret endpoint/generation metadata, and current operations succeed. **And** given the source endpoint that a BOUND rebind tombstoned, **when** a turn registration carrying that old (tombstoned) session UUID is submitted through the t1074 UserPromptSubmit registration path (`RegisterPeer` with a session UUID that is not launch-pending) — once as a later resume of the old session with its recorded PID and process-start, once as a restart of it with a new PID and process-start, and once more after the broker handle is reopened — **then** each is refused with `STALE_ENDPOINT`, the refusal's `StaleEndpoint` redirect names the lane's current endpoint session UUID and generation, and the lane endpoint row (session UUID, generation, PID, process-start), the tombstone rows, and the BOUND receipt rows are identical before and after each attempt. **And** given the BOUND state before any launcher registration, with the handoff-bound endpoint current, **when** `BindLaunchPending` is called with the tombstoned session UUID, the handoff-bound owner's PID and process-start, and `Generation: 1`, **then** it returns `STALE_ENDPOINT` whose `StaleEndpoint` redirect names the handoff-bound endpoint's session UUID and generation, and the lane endpoint row (session UUID, generation, PID, process-start), the tombstone rows, and the BOUND receipt rows are identical before and after the call: a bound row is refused, not silently passed over. **And** given the same tombstoned source session UUID after the handoff-bound owner is made not current under the t1074 live-owner rule, **when** the launcher resume path runs — `RegisterLaunchPending` for the lane's slot with a new PID and process-start, then the SessionStart bind `BindLaunchPending` carrying the tombstoned session UUID with that launcher-observed PID and process-start and `Generation: 1` — the constant the production SessionStart hook puts in its bind input (`internal/hook/factory_messages.go:86`, inside `registerFactoryHookPeer` :53) — after the test has asserted that 1 differs from the tombstoned generation — **then** the launcher registration commits a launch-pending row at a generation strictly higher than the handoff-bound generation; the bind is refused with `STALE_ENDPOINT`, and its `StaleEndpoint` redirect names the lane's current endpoint, which is that launch-pending row (session UUID redacted to empty, its generation); after the refusal the lane endpoint row (private token, generation, PID, process-start) equals the row the launcher registration committed, and the tombstone rows and BOUND receipt rows are identical to their state before `RegisterLaunchPending`; a following turn registration (`RegisterPeer`) with the tombstoned UUID and the same PID and process-start is refused with `STALE_ENDPOINT` and leaves that row unchanged; and a send from, and a send to, the identity the refused bind would have produced — the tombstoned session UUID at the launch-pending generation + 1, with the launcher PID and process-start — both fail with an error in the `ErrStalePeer` class. **And** once that launch-pending owner is made not current, a fresh `RegisterLaunchPending` followed by `BindLaunchPending` with a session UUID that has no tombstone binds and becomes the current endpoint at a strictly higher generation, so the refusal is keyed to the tombstone and does not disable the launcher path.

The pinned values are what make these legs discriminate. `BindLaunchPending` does not use its input `Generation` (it writes the row's generation + 1), so the pin matters only against a wrong implementation. One that reused the (session UUID, generation) pair match of `staleOrUnregistered` (`internal/factorymsg/handoff_bind.go:421`) against the input generation refuses when that generation equals the tombstoned one and lets the bind through otherwise — the delta plan-audit probe observed the pair lookup refuse at generation 2 (the tombstoned generation) and let it through at 1, 4, and 5 — so with the input pinned to 1, which the test has asserted is not the tombstoned generation, that implementation lets the bind through and the launcher leg fails it; a pair match against the new generation (launch-pending + 1) lets it through as well. The two send legs discriminate revival for the same reason: at HEAD `03390136a`, once the tombstoned UUID has been bound at launch-pending generation + 1, a send from and a send to that identity both succeed, so these legs fail at HEAD, whereas a sender at the tombstoned generation is refused even at HEAD and would detect nothing. An implementation that reads the tombstone only after it has found a launch-pending row fails the BOUND-row leg, and one that refuses every launcher bind fails the positive leg.

```bash
unset ANTHROPIC_BASE_URL ANTHROPIC_API_KEY ANTHROPIC_AUTH_TOKEN Z_AI_API_KEY MOAI_KANBAN_BACKEND MOAI_FACTORY_WORKERS && mkdir -p .moai/reports/t1082 && MOAI_HOME=/tmp/t1082-ac07-home GOCACHE=/tmp/t1082-ac07-cache go test -json ./internal/factorymsg -run '^TestFactoryLaneHandoffStaleEndpointRejected$' -count=1 -timeout=90s > .moai/reports/t1082/ac07.jsonl && jq -se '([.[]|select(.Action=="pass" and .Test=="TestFactoryLaneHandoffStaleEndpointRejected")]|length)==1 and ([.[]|select(.Action=="fail" or .Action=="skip")]|length)==0 and ([.[]|select((.Output//"")|contains("NOT_RUN"))]|length)==0' .moai/reports/t1082/ac07.jsonl
```

Expected final output: `true`; otherwise FAIL.

### AC-FLH-008 — Duplicate dispatch and same-lane redispatch

**Given** a dispatch sent with idempotency key K1, an identical repeat of that K1 request within the same recipient generation, a resend after BOUND to the rebound endpoint using a new key K2, a request after BOUND that reuses K1 with a different recipient session/generation, and a retry addressed to the old generation, **when** dispatch repeats, **then** each check holds at the layer it names:

- Sender layer (`Store.Send`): the identical same-generation K1 repeat, sent by the same sender with no sender restart in between, returns the existing envelope (same message id, one `messages` row for K1); the K1 request with a different recipient is not merged into the existing K1 envelope — it does not return the K1 message id and adds no delivery of the K1 envelope — and whether it is rejected or stored as its own envelope depends on the idempotency basis, so the named test asserts neither; the K2 resend, whose key the sender has not used before, is accepted as its own envelope addressed to the current generation.
- Recipient layer (receipt disposition): the K1 body executes once, and exactly one receipt row is accepted by the broker for the K1 envelope; the same K1 envelope being claimed again within the same recipient generation, because its lease expired before a receipt, is acknowledged with `DispositionDuplicate` and does not execute the body again.
- Stale-generation layer: the retry addressed to the old generation returns stale NACK.

Same-key duplicate handling is asserted only within one recipient generation. Same-key deduplication across BOUND is not asserted. t1082 does not change the idempotency basis itself; the basis is owned by t1100 (SPEC-DUAL-HARNESS-RECOVERY-001), and every check above holds whichever basis is in force. The test fixtures must not depend on the key basis: no assertion may rely on `sender_session` being part of the uniqueness key, or on a sender restart opening a fresh key scope. Using a new key for a resend after BOUND remains a t1082 requirement on the sender regardless of basis.

```bash
unset ANTHROPIC_BASE_URL ANTHROPIC_API_KEY ANTHROPIC_AUTH_TOKEN Z_AI_API_KEY MOAI_KANBAN_BACKEND MOAI_FACTORY_WORKERS && mkdir -p .moai/reports/t1082 && MOAI_HOME=/tmp/t1082-ac08-home GOCACHE=/tmp/t1082-ac08-cache go test -json ./internal/factorymsg -run '^TestFactoryLaneHandoffDuplicateAndSameLaneRedispatch$' -count=1 -timeout=90s > .moai/reports/t1082/ac08.jsonl && jq -se '([.[]|select(.Action=="pass" and .Test=="TestFactoryLaneHandoffDuplicateAndSameLaneRedispatch")]|length)==1 and ([.[]|select(.Action=="fail" or .Action=="skip")]|length)==0 and ([.[]|select((.Output//"")|contains("NOT_RUN"))]|length)==0' .moai/reports/t1082/ac08.jsonl
```

Expected final output: `true`; otherwise FAIL.

### AC-FLH-009 — Crash and restart recovery

**Given** failpoints after reserve, create, rename, SWITCH_PENDING, each rebind write, commit, and before receipt delivery, **when** the process restarts repeatedly, **then** each case deterministically resumes, finalizes idempotently, or NACKs; no case has two current endpoints, duplicate release, or lost receipt.

```bash
unset ANTHROPIC_BASE_URL ANTHROPIC_API_KEY ANTHROPIC_AUTH_TOKEN Z_AI_API_KEY MOAI_KANBAN_BACKEND MOAI_FACTORY_WORKERS && mkdir -p .moai/reports/t1082 && MOAI_HOME=/tmp/t1082-ac09-home GOCACHE=/tmp/t1082-ac09-cache go test -json ./internal/factorymsg ./internal/cli -run '^TestFactoryLaneHandoffCrashRecovery$' -count=1 -timeout=120s > .moai/reports/t1082/ac09.jsonl && jq -se '([.[]|select(.Action=="pass" and .Test=="TestFactoryLaneHandoffCrashRecovery")]|length)==1 and ([.[]|select(.Action=="fail" or .Action=="skip")]|length)==0 and ([.[]|select((.Output//"")|contains("NOT_RUN"))]|length)==0' .moai/reports/t1082/ac09.jsonl
```

Expected final output: `true`; otherwise FAIL.

### AC-FLH-010 — Abandoned worktree preservation

**Given** dirty, unmerged, base-drifted, branch-collided, and unknown-owner target worktrees, **when** recovery cannot prove a safe resume, **then** it records ABANDONED with reason, preserves files/branch/commits byte-for-byte, does not remove the worktree, and does not bind or dispatch.

```bash
unset ANTHROPIC_BASE_URL ANTHROPIC_API_KEY ANTHROPIC_AUTH_TOKEN Z_AI_API_KEY MOAI_KANBAN_BACKEND MOAI_FACTORY_WORKERS && mkdir -p .moai/reports/t1082 && MOAI_HOME=/tmp/t1082-ac10-home GOCACHE=/tmp/t1082-ac10-cache go test -json ./internal/cli ./internal/factorymsg -run '^TestFactoryLaneHandoffAbandonedWorktreeRecovery$' -count=1 -timeout=90s > .moai/reports/t1082/ac10.jsonl && jq -se '([.[]|select(.Action=="pass" and .Test=="TestFactoryLaneHandoffAbandonedWorktreeRecovery")]|length)==1 and ([.[]|select(.Action=="fail" or .Action=="skip")]|length)==0 and ([.[]|select((.Output//"")|contains("NOT_RUN"))]|length)==0' .moai/reports/t1082/ac10.jsonl
```

Expected final output: `true`; otherwise FAIL.

### AC-FLH-011 — Zero writes before BOUND and no message loss

**Given** instrumented filesystem/Git/broker adapters across all states, **when** handoff proceeds through BOUND and one dispatch, **then** counters are exactly: pre-BOUND code writes 0, wrong-cwd writes 0, primary branch switches 0, pre-BOUND commits 0, task ACKs 0, and messages sent/received/lost `N/N/0`.

```bash
unset ANTHROPIC_BASE_URL ANTHROPIC_API_KEY ANTHROPIC_AUTH_TOKEN Z_AI_API_KEY MOAI_KANBAN_BACKEND MOAI_FACTORY_WORKERS && mkdir -p .moai/reports/t1082 && MOAI_HOME=/tmp/t1082-ac11-home GOCACHE=/tmp/t1082-ac11-cache go test -json ./internal/factorymsg ./internal/cli ./internal/hook -run '^TestFactoryLaneHandoffNoPreBoundWrites$' -count=1 -timeout=90s > .moai/reports/t1082/ac11.jsonl && jq -se '([.[]|select(.Action=="pass" and .Test=="TestFactoryLaneHandoffNoPreBoundWrites")]|length)==1 and ([.[]|select(.Action=="fail" or .Action=="skip")]|length)==0 and ([.[]|select((.Output//"")|contains("NOT_RUN"))]|length)==0' .moai/reports/t1082/ac11.jsonl
```

Expected final output: `true`; otherwise FAIL.

### AC-FLH-014 — No private or overstated control path

**Given** spies for hook, MCP, tmux, sockets, model messages, Desktop Handoff, app-server, and user guidance, **when** interactive and headless handoffs run, **then** interactive emits guidance only, headless uses official app-server only, and slash execution/tmux/private socket/model-cd/Desktop-Handoff emulation counts are all 0.

```bash
unset ANTHROPIC_BASE_URL ANTHROPIC_API_KEY ANTHROPIC_AUTH_TOKEN Z_AI_API_KEY MOAI_KANBAN_BACKEND MOAI_FACTORY_WORKERS && mkdir -p .moai/reports/t1082 && MOAI_HOME=/tmp/t1082-ac14-home GOCACHE=/tmp/t1082-ac14-cache go test -json ./internal/cli ./internal/hook -run '^TestFactoryLaneHandoffNoPrivateControl$' -count=1 -timeout=90s > .moai/reports/t1082/ac14.jsonl && jq -se '([.[]|select(.Action=="pass" and .Test=="TestFactoryLaneHandoffNoPrivateControl")]|length)==1 and ([.[]|select(.Action=="fail" or .Action=="skip")]|length)==0 and ([.[]|select((.Output//"")|contains("NOT_RUN"))]|length)==0' .moai/reports/t1082/ac14.jsonl
```

Expected final output: `true`; otherwise FAIL.

### AC-FLH-015 — t1074 and catalog compatibility

**Given** existing t1074 broker/roster/receipt/SessionStart/MCP catalog fixtures plus filesystem/process inventory, **when** handoff additions run, **then** legacy behavior remains passing, the MCP catalog total equals the count declared by `.claude/rules/moai/core/moai-mcp-tools.md` (pinned by `wantCatalogSize` in `internal/mcp/catalog_test.go`), the write/read split equals the per-tool `WriteCapable` flags in `internal/mcp/catalog.go` (pinned by `TestMoaiMCPTools_WriteCapableSet`), and the registered server matches both — this SPEC adds no MCP tool, and a count change is authorized only by a card that updates those sources — and exactly zero new broker DB roots, daemons, private sockets, or parallel message stores exist.

```bash
unset ANTHROPIC_BASE_URL ANTHROPIC_API_KEY ANTHROPIC_AUTH_TOKEN Z_AI_API_KEY MOAI_KANBAN_BACKEND MOAI_FACTORY_WORKERS && mkdir -p .moai/reports/t1082 && MOAI_HOME=/tmp/t1082-ac15-home GOCACHE=/tmp/t1082-ac15-cache go test -json ./internal/factorymsg ./internal/hook ./internal/cli ./internal/mcp -run '^TestFactoryLaneHandoffT1074Compatibility$' -count=1 -timeout=120s > .moai/reports/t1082/ac15.jsonl && jq -se '([.[]|select(.Action=="pass" and .Test=="TestFactoryLaneHandoffT1074Compatibility")]|length)==1 and ([.[]|select(.Action=="fail" or .Action=="skip")]|length)==0 and ([.[]|select((.Output//"")|contains("NOT_RUN"))]|length)==0' .moai/reports/t1082/ac15.jsonl
```

Expected final output: `true`; otherwise FAIL.

### AC-FLH-016 — Creation-base drift fail-closed

**Given** a mutant reproducing t1082 creation from `main@2213871af` while reserved local develop pin is `3f3ffbb57`, plus develop-moving-during-create and matching-base controls, **when** post-create provenance validation runs, **then** both mismatches return `BASE_DRIFT`, emit no SWITCH_PENDING/BOUND/dispatch, preserve the target for inspection, and only the exact-match control advances.

```bash
unset ANTHROPIC_BASE_URL ANTHROPIC_API_KEY ANTHROPIC_AUTH_TOKEN Z_AI_API_KEY MOAI_KANBAN_BACKEND MOAI_FACTORY_WORKERS && mkdir -p .moai/reports/t1082 && MOAI_HOME=/tmp/t1082-ac16-home GOCACHE=/tmp/t1082-ac16-cache go test -json ./internal/cli -run '^TestFactoryLaneHandoffCreationBaseDriftRejected$' -count=1 -timeout=90s > .moai/reports/t1082/ac16.jsonl && jq -se '([.[]|select(.Action=="pass" and .Test=="TestFactoryLaneHandoffCreationBaseDriftRejected")]|length)==1 and ([.[]|select(.Action=="fail" or .Action=="skip")]|length)==0 and ([.[]|select((.Output//"")|contains("NOT_RUN"))]|length)==0' .moai/reports/t1082/ac16.jsonl
```

Expected final output: `true`; otherwise FAIL.

### AC-FLH-017 — Launch-pending source endpoint NACK

**Given** a lane slot whose endpoint was registered through the production launcher provisional path (`RegisterLaunchPending`, not a manual DB seed) so that lane resolution returns `ErrEndpointLaunchPending`, **when** a handoff reserve is attempted for that lane in both interactive and headless mode, **then** each returns NACK with reason `ENDPOINT_LAUNCH_PENDING`; handoff/reservation rows, tombstone rows, and dispatch release markers each count 0; no `<primary>/.claude/worktrees/<card-id>` path exists; the app-server request spy and the rollback spy each count 0; and the provisional endpoint row (slot, session key, generation, PID, process-start, updated_at) is byte-identical before and after. **And** as a control, after the production `BindLaunchPending` binds that slot, a fresh reserve for the same lane is admitted.

```bash
unset ANTHROPIC_BASE_URL ANTHROPIC_API_KEY ANTHROPIC_AUTH_TOKEN Z_AI_API_KEY MOAI_KANBAN_BACKEND MOAI_FACTORY_WORKERS && mkdir -p .moai/reports/t1082 && MOAI_HOME=/tmp/t1082-ac17-home GOCACHE=/tmp/t1082-ac17-cache go test -json ./internal/factorymsg ./internal/cli -run '^TestFactoryLaneHandoffLaunchPendingSourceNack$' -count=1 -timeout=90s > .moai/reports/t1082/ac17.jsonl && jq -se '([.[]|select(.Action=="pass" and .Test=="TestFactoryLaneHandoffLaunchPendingSourceNack")]|length)==1 and ([.[]|select(.Action=="fail" or .Action=="skip")]|length)==0 and ([.[]|select((.Output//"")|contains("NOT_RUN"))]|length)==0' .moai/reports/t1082/ac17.jsonl
```

Expected final output: `true`; otherwise FAIL.

### AC-FLH-018 — Launcher bind versus handoff rebind race on one slot

**Given** one broker path with a lane bound at generation `g`, a handoff reservation in `SWITCH_PENDING` pinned to that source endpoint, the source owner arranged as not current so that a relaunch may re-register the slot, and the handoff's new endpoint owner arranged with a PID/process-start identity different from both the source and A's relaunched identity and arranged as current (the only fixture under which the launcher-loses branch reaches the t1074 live-owner rejection; with the same identity as the source, or a non-current new owner, the launcher overwrites the row as launch-pending instead), and racers A and B each holding their own `Store` handle obtained by a separate production `Open` call on that same broker path (distinct `*Store` and distinct underlying `*sql.DB`), **when** goroutine A runs the production launcher path (`RegisterLaunchPending` then `BindLaunchPending` for a relaunched process) and goroutine B runs the handoff atomic rebind on the same slot — first under the forced interleavings below via a test barrier, then under 200 unforced concurrent iterations on fresh brokers — **then** after every run: exactly one endpoint row exists for the slot and it is not launch-pending; the generation is greater than `g` and never decreased across observed commits; if B committed first, the current owner is the handoff endpoint, the handoff is `BOUND` with exactly one tombstone and one BOUND receipt, and A's registration returned the t1074 live-owner rejection without changing the row or the handoff; if A's registration committed first, the handoff read immediately after that commit and before B starts is already `NACK` with reason `STALE_GENERATION`, B then returns `STALE_GENERATION` and writes nothing, tombstone/receipt/release counts are 0, and the current owner is the launcher-bound session.

Forced interleavings. Owner currency is produced by the process-start fingerprint, never by injecting `Store.ownerCurrent`: that field is unexported, and the test for orders (iv) and (v) must live in package `hook` (where `registerFactorySessionStartPeer`/`registerFactoryUserPromptPeer` are unexported and from which `ownerCurrent` cannot be set). A fake process-start makes an owner not current; a live process's real PID and process-start fingerprint make it current. The source row is seeded through the production launcher path as a registered peer — `RegisterLaunchPending` (which writes through `RegisterPeer`, `internal/factorymsg/store.go:315`) then `BindLaunchPending` — carrying a fake process-start, never by a raw DB insert or a direct `RegisterPeer` seed. The relaunched owner identity A registers equals the identity the production hook path resolves in the test process. The handoff's new endpoint owner is a live process distinct from the test process, so that it is current and differs from A's identity:

- (i) A-register→B→A-bind, (ii) B→A, (iii) A→B, with outcomes as above.
- (iv) SessionStart before A-register: the relaunched process's SessionStart peer registration (production `BindLaunchPending`, as reached from `registerFactorySessionStartPeer`) runs first against the still-bound source row and emits no bind and no row change; A-register commits and moves the handoff to `NACK`/`STALE_GENERATION` in the same commit; then the relaunched process's UserPromptSubmit registration (production `RegisterPeer`, as reached from `registerFactoryUserPromptPeer`) binds the provisional row; then B returns `STALE_GENERATION`. Expected end state: the row is the UserPromptSubmit session, not launch-pending, at generation `g+2`; tombstone, BOUND receipt, and dispatch release counts are each 0; a dispatch body claim is refused.
- (v) Alias correction after A-register: A-register commits (handoff `NACK` in the same commit); SessionStart binds an alias session UUID to the provisional row at `g+2`; the real session's UserPromptSubmit registration with the same PID/process-start replaces it at `g+3`. Expected: the row is the real session at `g+3`, the handoff stays `NACK`, tombstone/receipt/release counts are 0 — t1074 alias correction still works because the handoff is already final.

Each forced interleaving must be observed exactly as forced, or the test FAILs rather than passing vacuously. **And** every unforced run ends with no non-final handoff whose reserved source tuple differs from the row, and with no launch-pending row. The test runs under `-race`.

**Separate-handle requirement.** Sharing one `Store` handle between A and B is a FAIL condition of the test itself: each handle is opened with `SetMaxOpenConns(1)`, so a shared handle serializes the racers in the Go connection pool and never exercises the SQLite `BEGIN IMMEDIATE` boundary this criterion exists to test. The test proves two handles were used by (1) asserting pointer inequality of the two `*Store` values and of their underlying `*sql.DB` values before any racer runs; (2) asserting that while one racer holds its write transaction at a forced barrier, the waiting racer's own handle reports `db.Stats().InUse == 1` and `db.Stats().WaitCount == 0` and its write attempt has not returned — i.e., it holds its own connection and waits on the SQLite lock, not in the holder's pool; and (3) first running its handle-distinctness guard on a deliberately shared pair and observing the guard fail before the real racers run, so the guard is known to go red. Only after (1)-(3) pass does the test emit the log line `RACER_HANDLES_DISTINCT=2`, which the gate below requires.

```bash
unset ANTHROPIC_BASE_URL ANTHROPIC_API_KEY ANTHROPIC_AUTH_TOKEN Z_AI_API_KEY MOAI_KANBAN_BACKEND MOAI_FACTORY_WORKERS && mkdir -p .moai/reports/t1082 && MOAI_HOME=/tmp/t1082-ac18-home GOCACHE=/tmp/t1082-ac18-cache go test -json -race ./internal/factorymsg ./internal/hook -run '^TestFactoryLaneHandoffRebindVsLaunchBindRace$' -count=1 -timeout=180s > .moai/reports/t1082/ac18.jsonl && jq -se '([.[]|select(.Action=="pass" and .Test=="TestFactoryLaneHandoffRebindVsLaunchBindRace")]|length)==1 and ([.[]|select(.Action=="fail" or .Action=="skip")]|length)==0 and ([.[]|select((.Output//"")|contains("NOT_RUN") or contains("DATA RACE"))]|length)==0 and ([.[]|select((.Output//"")|contains("RACER_HANDLES_DISTINCT=2"))]|length)>=1' .moai/reports/t1082/ac18.jsonl
```

Expected final output: `true`; otherwise FAIL.

### AC-FLH-019 — Handoff rebind versus UserPromptSubmit registration on one slot

**Given** one broker path; a lane slot bound at generation `g` through the production launcher path (`RegisterLaunchPending` then `BindLaunchPending`, not a manual DB seed) to source session `src-uuid` owned by a live PID `p` with process-start `s` (orders (v) and (vii) use the not-current source variant defined below instead); a handoff for that lane reserved and in `SWITCH_PENDING_INTERACTIVE` pinned to that source tuple; racer H and racer R each holding their own `Store` handle obtained by a separate production `Open` call on that same broker path (distinct `*Store` and distinct underlying `*sql.DB`); racer H calling the production store entry point that the UserPromptSubmit hook path calls (`RegisterPeer`, `internal/factorymsg/store.go:315`, reached from `registerFactoryUserPromptPeer` as of t1074) with the same PID `p`, the same process-start `s`, and a new session UUID `post-cd-uuid` — the shape of the turn after an interactive `/cd`; and racer R running the handoff atomic rebind with matching interactive evidence for `post-cd-uuid`. **When** the two racers run under forced interleavings via a test barrier — (i) H's write transaction starts first and is held at the barrier until R's rebind is blocked on its own handle, then released; (ii) R's rebind transaction starts first and is held until H's registration is blocked on its own handle, then released, followed by one further H call carrying the tombstoned `src-uuid` with the same `p` and `s`; (iii) R's rebind fails validation (branch mismatch) first, then H runs; (iv)–(vii) the same-transaction read orders defined below — and then under 200 unforced concurrent H-versus-R iterations on fresh brokers, **then**:

- In (i), H returns `ENDPOINT_HANDOFF_PENDING`, and the endpoint row (session, generation, PID, process-start, updated_at) observed between H's return and R's commit is byte-identical to the row before H started. R then commits: the endpoint is `post-cd-uuid` at generation `g+1`, the handoff is `BOUND`, exactly one tombstone names `src-uuid`/`g`, and exactly one BOUND receipt exists.
- In (ii), R commits first with the same end state as (i). H's `post-cd-uuid` call then leaves session, generation, PID, and process-start unchanged, with the tombstone count still 1 and the receipt count still 1. The `src-uuid` call returns `STALE_ENDPOINT` and changes nothing.
- In (iii), the handoff is `NACK`. Tombstone, BOUND receipt, and dispatch release counts are each 0. The endpoint stays `src-uuid` at `g` until H runs. H then follows t1074 semantics (endpoint `post-cd-uuid` at `g+1`), and the dispatch body remains denied because no handoff is `BOUND`.
- In every forced and unforced run, while the handoff is non-final, no committed transaction other than R's rebind, or a launcher registration that finalizes the handoff in the same commit, changes the endpoint. Every endpoint change by R's rebind is accompanied by exactly one tombstone and one BOUND receipt. The generation never decreases across observed commits. Each unforced run ends in exactly one of the (i) or (ii) end states.

**Same-transaction read orders (iv)–(vii).** Orders (iv)–(vi) discriminate where the registration paths read the handoff state; order (vii) discriminates where the reservation reads the source row. Every order starts from a bound source row with no handoff for the lane. Orders (iv) and (vi) use the Given's source row owned by live `p`/`s`. Orders (v) and (vii) use the **not-current source variant**: the source row is seeded through the same production launcher path (`RegisterLaunchPending` then `BindLaunchPending`, as a registered peer) with a fake process-start `s'`, so the source owner is not current from seeding onward and a relaunch registration is admitted by the t1074 live-owner rule; H is not used in these two orders. In (iv)–(vi), V's reserved source tuple is that row. Adversary V is the production reservation path (t1082 new code, which is where the barrier seam lives). V opens `BEGIN IMMEDIATE` on its own `Store` handle, inserts a `RESERVED` handoff row for the slot without committing, and stops at a barrier.

- (iv) Subject H, as defined above, starts on a separate handle and is observed blocked on the SQLite lock: its handle reports `InUse == 1` and `WaitCount == 0`, and its call has not returned. V then commits within the handle's `busy_timeout` (2500 ms for the production `Open`), and H returns. PASS: H returns `ENDPOINT_HANDOFF_PENDING`, and the endpoint row is byte-identical to the reserved source tuple, including `updated_at`. An implementation that reads the handoff state before its own `BEGIN` fails here, because its snapshot misses the uncommitted `RESERVED` row.
- (v) The same structure with subject A, the production launcher registration (`RegisterLaunchPending`, `internal/factorymsg/store.go:391`, the store entry point `registerFactoryLaunchPending` in `internal/cli/factory_launch_pending.go:34` calls), on the not-current source variant. PASS: the row is launch-pending AND the handoff is `NACK`/`STALE_GENERATION`, both produced by A's single commit. FAIL: the row is launch-pending while the handoff is still `RESERVED`.
- (vi) Control only: H runs to completion before V begins. H follows t1074 semantics and V then reserves against the resulting row. Both a read-inside-transaction and a read-before-`BEGIN` implementation pass this order; it proves the barrier fixture itself, not the read position.
- (vii) Reversed roles: A is the adversary and V is the subject, on the not-current source variant. A runs the production launcher provisional registration (`RegisterLaunchPending` → `RegisterPeer`, the path `registerFactoryLaunchPending` reaches) on its own `Store` handle; inside its write transaction it writes the slot row as launch-pending and stops at a barrier without committing. A's barrier seam lives only in t1082-added code: the handoff-state read-and-finalize step that REQ-FLH-017 adds inside `RegisterPeer`'s write transaction (`internal/factorymsg/store.go:315`), placed after the slot-row write and before commit; t1074-landed lines carry no seam. The subject V, the production reservation path, then starts on a separate handle for the same lane and is observed blocked on the SQLite lock: its handle reports `InUse == 1` and `WaitCount == 0`, and its call has not returned. A then commits within the handle's `busy_timeout`, and V returns. PASS: V returns NACK with reason `ENDPOINT_LAUNCH_PENDING` — the reservation-rejection reason REQ-FLH-016 prescribes for a launch-pending source endpoint, which is how REQ-FLH-017's "order the t1074 launcher provisional-endpoint bind before handoff admission" is enforced (REQ-FLH-017 itself defines no reservation-rejection reason; its `STALE_GENERATION` applies to a handoff that already exists) — no handoff/reservation row exists for the lane, and the slot row (session, generation, PID, process-start, updated_at) is byte-identical to the row A committed. FAIL: a reservation row exists whose reserved source tuple differs from the current slot row. A read-before-`BEGIN` reservation produces exactly this: it reads the pre-A bound row, passes the REQ-FLH-016 check against it, waits for A's commit, and then commits `RESERVED` pinned to the stale tuple.

For (iv) the test records `H_blocked_observed=true`, for (v) `A_blocked_observed=true`, and for both timestamps showing V's commit earlier than the subject's return. For (vii) the test records `V_blocked_observed=true` and timestamps showing A's commit earlier than V's return. A missing observation or a reversed timestamp order is a FAIL.

**Subject and seam placement.** The subject H is the production `RegisterPeer` store entry point as shipped. It is the exact function the UserPromptSubmit hook path reaches (`registerFactoryUserPromptPeer`, `internal/hook/factory_messages.go:49`, through `registerFactoryHookPeer` at `:53`), invoked on the racer's own production-`Open`ed handle, and no test-only registration path or variant exists. It is the entry under test, not fixture preparation, so the policy against direct peer registration does not apply to it. Barrier seams exist only in t1082-added code: V's reservation, R's rebind, the handoff-state check that REQ-FLH-018 adds inside the registration transaction, which forced order (i) uses, and the handoff-state read-and-finalize step that REQ-FLH-017 adds inside the same `RegisterPeer` transaction, which forced order (vii) uses. Code landed by t1074 carries no seam.

Each forced interleaving must be observed exactly as forced — the recorded start order of the two write transactions and the commit order must equal the forced order. An order that was not actually observed is a FAIL, never a vacuous PASS. The separate-handle requirement and its three-part proof from AC-FLH-018 apply verbatim to every racer pair here (H/R, V/H, V/A, A/V), including the `InUse`/`WaitCount` proof, the shared-pair guard probe, and the `RACER_HANDLES_DISTINCT=2` log line. The test runs under `-race`.

```bash
unset ANTHROPIC_BASE_URL ANTHROPIC_API_KEY ANTHROPIC_AUTH_TOKEN Z_AI_API_KEY MOAI_KANBAN_BACKEND MOAI_FACTORY_WORKERS && mkdir -p .moai/reports/t1082 && MOAI_HOME=/tmp/t1082-ac19-home GOCACHE=/tmp/t1082-ac19-cache go test -json -race ./internal/factorymsg ./internal/hook -run '^TestFactoryLaneHandoffRebindVsUserPromptRegisterRace$' -count=1 -timeout=180s > .moai/reports/t1082/ac19.jsonl && jq -se '([.[]|select(.Action=="pass" and .Test=="TestFactoryLaneHandoffRebindVsUserPromptRegisterRace")]|length)==1 and ([.[]|select(.Action=="fail" or .Action=="skip")]|length)==0 and ([.[]|select((.Output//"")|contains("NOT_RUN") or contains("DATA RACE"))]|length)==0 and ([.[]|select((.Output//"")|contains("RACER_HANDLES_DISTINCT=2"))]|length)>=1' .moai/reports/t1082/ac19.jsonl
```

Expected final output: `true`; otherwise FAIL.

### AC-FLH-020 — Operator termination of a stuck handoff

**Given** a lane with a non-final handoff in each of `RESERVED`, `WT_READY`, `SWITCH_PENDING_INTERACTIVE`, and `SWITCH_PENDING_HEADLESS`, a dirty target worktree with an unmerged commit, a UserPromptSubmit registration for that slot that is refused with `ENDPOINT_HANDOFF_PENDING` before termination, and three source-owner cases — owner current by the t1074 PID and process-start rule, owner liveness that cannot be established, and owner not current — **when** the operator runs `moai factory handoff abandon-lane --slot <slot>`, **then**:

- Owner current, and owner liveness not established: the command is refused with `SOURCE_OWNER_LIVE`, and the handoff row, endpoint row, generation, tombstones, receipts, and dispatch release markers are byte-for-byte unchanged.
- Owner not current: the handoff is `ABANDONED` with reason `OPERATOR_ABANDONED`, written in one broker write transaction; the target worktree's files, branch, and commits are preserved and the worktree is not removed; the count of `BOUND` states, tombstones, BOUND receipts, and dispatch releases written is 0; and a UserPromptSubmit registration for the slot issued after the commit follows t1074 semantics and is not refused with `ENDPOINT_HANDOFF_PENDING`, while dispatch bodies stay withheld because no `BOUND` exists.
- A second run on the same slot, and a run on a slot with no non-final handoff, report `HANDOFF_NOT_PENDING` and write nothing.

The command is reached through the production `moai` command tree, not by calling the store function directly. The only permitted liveness seam is named: the abandon store function in `internal/factorymsg` takes the source-owner probe as an argument of type `func(int) (string, homestate.ProcessIdentityState)` (precedent `internal/cli/factory_handoff_recover.go:30`, which passes `homestate.ProbeProcessIdentity` to `RecoverLegacyResume`); the `internal/cli` command wires that argument through one package-level variable whose production value is `homestate.ProbeProcessIdentity`; and the test swaps only that variable to return `ProcessIdentityLive` with the recorded process-start (owner current), `ProcessIdentityIndeterminate` (liveness not established), and `ProcessIdentityDead` (owner not current). No other probe, store field, or broker row is stubbed. The indeterminate leg fails an implementation that reuses the `ownerCurrent` bool seam (`internal/factorymsg/store.go:116`), because that seam folds `ProcessIdentityIndeterminate` into "not current" and would abandon the handoff instead of refusing with `SOURCE_OWNER_LIVE`.

```bash
unset ANTHROPIC_BASE_URL ANTHROPIC_API_KEY ANTHROPIC_AUTH_TOKEN Z_AI_API_KEY MOAI_KANBAN_BACKEND MOAI_FACTORY_WORKERS && mkdir -p .moai/reports/t1082 && MOAI_HOME=/tmp/t1082-ac20-home GOCACHE=/tmp/t1082-ac20-cache go test -json ./internal/cli ./internal/factorymsg -run '^TestFactoryLaneHandoffOperatorAbandon$' -count=1 -timeout=120s > .moai/reports/t1082/ac20.jsonl && jq -se '([.[]|select(.Action=="pass" and .Test=="TestFactoryLaneHandoffOperatorAbandon")]|length)==1 and ([.[]|select(.Action=="fail" or .Action=="skip")]|length)==0 and ([.[]|select((.Output//"")|contains("NOT_RUN"))]|length)==0' .moai/reports/t1082/ac20.jsonl
```

Expected final output: `true`; otherwise FAIL.

## Exclusions — 카드 t1145로 이관 (0.5.11)

### Out of Scope — LIVE cross-harness proof (AC-FLH-012·013)

- 리드 결정 (a)(sync-audit `.moai/reports/t1082/sync-audit-deep.md` F1)에 따라 REQ-FLH-014와 AC-FLH-012·013을 이 SPEC 범위에서 빼고 카드 t1145에 넘겼다. 이 SPEC은 두 AC가 통과했다고 주장하지 않는다.
- 아래 원문(요약표 행, Acceptance policy의 LIVE 조항, RED-now 원장 행, AC 본문의 Given/When/Then·명령·jq predicate·기대 결과)은 t1145가 바꾸지 않고 이어받도록 0.5.10 문장 그대로 옮겼다. 제목만 한 단계 내렸다(`###` → `####`).
- 두 AC의 명령에는 0.5.11의 환경 정리 목록 보강(`MOAI_KANBAN_BACKEND`·`MOAI_FACTORY_WORKERS`)을 원문 보존을 위해 적용하지 않았다. t1145가 이어받을 때 이 두 변수를 더해야 한다.
- `internal/cli/factory_lane_handoff_live_gate_test.go`는 이 문서에서 샵 세 개 뒤에 해당 AC ID가 오는 문자열이 처음 나오는 곳을 찾고, 그 뒤의 첫 bash 블록을 읽는다. 아래 `####` 제목은 그 문자열을 포함하므로 테스트는 옮긴 뒤에도 같은 predicate를 읽는다. 그 문자열을 이 제목들보다 앞에 쓰면 테스트가 엉뚱한 블록을 읽으니 쓰지 않는다.

#### 원래 요약표 행

| AC | Requirements | Named test | Observable outcome |
|---|---|---|---|
| AC-FLH-012 | REQ-FLH-013, REQ-FLH-014 | `TestFactoryLiveCodexCodexWorktreeHandoff` | real Codex↔Codex가 실제 interactive `/cd`와 다음 정상 turn SessionStart, empty-turn 0을 증명한다. **`NOT_RUN → t1145`** (progress.md § M5 lane record) |
| AC-FLH-013 | REQ-FLH-013, REQ-FLH-014 | `TestFactoryLiveClaudeCodexWorktreeHandoff` | real Claude lead↔Codex가 실제 headless `thread/fork(cwd)`와 반환 ID direct BOUND를 증명한다. **`NOT_RUN → t1145`** (progress.md § M5 lane record) |

#### 원래 Acceptance policy의 LIVE 조항

- Unit/fixture evidence는 해당 named contract만 증명한다. AC-FLH-012/013은 실제 별도 CLI/model contexts가 아니면 PASS가 아니다.
- LIVE parent test는 자유 형식 stdout 문자열로 자기 증명하지 않는다. 각 test는 card-scoped structured evidence JSON을 원자적으로 기록하고, 아래 `jq -e` predicate가 필드의 타입·값·상호 일치를 직접 검증해야 한다.
- `TestFactoryLaneHandoffLiveEvidenceGateRejectsMutants`는 필수 필드 누락, fixture/mock/direct-registration 표식, child failure, child skip, 그리고 stored-history 증거를 `thread/start`로 바꾼 `wrong_method_thread_start`를 각각 주입한 gate mutant가 모두 거부됨을 증명한다. 이 selector는 AC 수를 늘리지 않고 AC-FLH-012/013의 공통 gate-quality 조건이다.

#### 원래 RED-now 원장 행

| AC | Exact RED-now command | Observed stdout | Exit |
|---|---|---|---:|
| AC-FLH-012 | `rg -n -F 'func TestFactoryLiveCodexCodexWorktreeHandoff(' internal --glob '*_test.go'` | `<empty>` | 1 |
| AC-FLH-013 | `rg -n -F 'func TestFactoryLiveClaudeCodexWorktreeHandoff(' internal --glob '*_test.go'` | `<empty>` | 1 |
| AC-FLH-012/013 gate quality | `rg -n -F 'func TestFactoryLaneHandoffLiveEvidenceGateRejectsMutants(' internal --glob '*_test.go'` | `<empty>` | 1 |

#### AC-FLH-012 — LIVE Codex lead ↔ Codex lane

**상태: `NOT_RUN → t1145`.** 이 SPEC 안에서는 실행되지 않았고 FAIL로 남는다. 운영자의 실제 `/cd`가 필요하고, handoff를 시작하는 production 경로가 없기 때문이다(progress.md § M5 lane record). 아래 기대 결과·명령·jq predicate는 바꾸지 않았다.

**Given** two real separately launched Codex model contexts in one factory run, **when** the lane receives a card and the operator actually executes `/cd` before sending the next normal user turn, **then** evidence shows `SWITCH_PENDING_INTERACTIVE` before that turn, its SessionStart-driven new thread/session generation, target statusline/cwd/`WT-*` branch, empty model turn count 0, BOUND receipt, explicit message receipt, old endpoint rejection, pre-BOUND writes 0, primary branch switch 0, and message loss 0.

```bash
unset ANTHROPIC_BASE_URL ANTHROPIC_API_KEY ANTHROPIC_AUTH_TOKEN Z_AI_API_KEY && mkdir -p .moai/reports/t1082 && MOAI_FACTORY_LIVE=1 MOAI_FACTORY_LIVE_CASE=codex-codex-interactive-cd-handoff MOAI_FACTORY_EVIDENCE=.moai/reports/t1082/ac12-evidence.json MOAI_HOME=/tmp/t1082-ac12-home GOCACHE=/tmp/t1082-ac12-cache go test -json ./internal/cli -run '^(TestFactoryLiveCodexCodexWorktreeHandoff|TestFactoryLaneHandoffLiveEvidenceGateRejectsMutants)$' -count=1 -timeout=240s > .moai/reports/t1082/ac12.jsonl && jq -se '([.[]|select(.Action=="pass" and .Test=="TestFactoryLiveCodexCodexWorktreeHandoff")]|length)==1 and ([.[]|select(.Action=="pass" and .Test=="TestFactoryLaneHandoffLiveEvidenceGateRejectsMutants")]|length)==1 and ([.[]|select(.Action=="fail" or .Action=="skip")]|length)==0 and ([.[]|select((.Output//"")|contains("NOT_RUN"))]|length)==0' .moai/reports/t1082/ac12.jsonl && jq -e '(.schema_version==1) and (.card_id=="t1082") and (.spec_id=="SPEC-FACTORY-LANE-WORKTREE-HANDOFF-001") and (.mode=="interactive") and (.contexts.separate==true) and (.contexts.lead.production_cli==true) and (.contexts.lead.real_model==true) and (.contexts.lane.production_cli==true) and (.contexts.lane.real_model==true) and (.contexts.lead.argv|type=="array" and length>0 and all(.[]; type=="string" and length>0)) and (.contexts.lane.argv|type=="array" and length>0 and all(.[]; type=="string" and length>0)) and (.contexts.lead.pid|type=="number" and .>1) and (.contexts.lane.pid|type=="number" and .>1) and (.contexts.lead.pid != .contexts.lane.pid) and (.contexts.lead.process_start|type=="string" and length>0) and (.contexts.lane.process_start|type=="string" and length>0) and (.contexts.lead.binary_sha256|type=="string" and test("^[0-9a-f]{64}$")) and (.contexts.lane.binary_sha256|type=="string" and test("^[0-9a-f]{64}$")) and (.interactive.actual_cd==true) and (.interactive.state_before_session_start=="SWITCH_PENDING_INTERACTIVE") and (.interactive.next_normal_turn_session_start==true) and (.target.reserved_cwd|type=="string" and startswith("/")) and (.target.observed_cwd==.target.reserved_cwd) and (.target.reserved_branch|type=="string" and startswith("WT-") and length>3) and (.target.observed_branch==.target.reserved_branch) and (.target.pinned_head|type=="string" and test("^[0-9a-f]{40}$")) and (.target.observed_head==.target.pinned_head) and (.empty_model_turn_count==0) and (.receipts.bound_id|type=="string" and length>0) and (.receipts.message_id|type=="string" and length>0) and (.old_endpoint.rejected==true) and (.old_endpoint.code=="STALE") and (.writes.pre_bound==0) and (.writes.wrong_cwd==0) and (.writes.primary_branch_switches==0) and (.messages.sent|type=="number" and .>0) and (.messages.received==.messages.sent) and (.messages.lost==0) and (.nonces.lead_to_lane|type=="string" and length>0) and (.nonces.lane_to_lead|type=="string" and length>0) and (.nonces.lead_to_lane != .nonces.lane_to_lead) and (.cleanup==true) and (.bypass.direct_register_peer==false) and (.bypass.db_seed==false) and (.bypass.mock==false) and (.bypass.fixture==false) and (.bypass.private_control==false)' .moai/reports/t1082/ac12-evidence.json
```

Expected outputs: 두 `jq` 모두 `true`. `.moai/reports/t1082/ac12-evidence.json`이 없거나 malformed이거나 위 exact predicate 중 하나라도 거짓이면 FAIL이다. stdout 문자열은 대체 증거가 아니다.

#### AC-FLH-013 — LIVE Claude lead ↔ Codex lane

**상태: `NOT_RUN → t1145`.** 이 SPEC 안에서는 실행되지 않았고 FAIL로 남는다. lead가 `thread/fork(cwd)`를 일으킬 production 경로가 없고, 격리 home에서는 두 harness 모두 인증되지 않았기 때문이다(progress.md § M5 lane record). 아래 기대 결과·명령·jq predicate는 바꾸지 않았다.

**Given** a real Claude lead and real separately launched headless Codex lane with stored history in one factory run, **when** the lead triggers actual `thread/fork(cwd)` and dispatches after verified direct BOUND, **then** evidence shows the official returned thread ID/lineage, controller cwd/branch/HEAD readback, SessionStart wait 0, empty model turn count 0, generation/receipt/stale-reject/zero-write/zero-loss, and bidirectional unique nonces.

```bash
unset ANTHROPIC_BASE_URL ANTHROPIC_API_KEY ANTHROPIC_AUTH_TOKEN Z_AI_API_KEY && mkdir -p .moai/reports/t1082 && MOAI_FACTORY_LIVE=1 MOAI_FACTORY_LIVE_CASE=claude-codex-headless-fork-cwd-handoff MOAI_FACTORY_EVIDENCE=.moai/reports/t1082/ac13-evidence.json MOAI_HOME=/tmp/t1082-ac13-home GOCACHE=/tmp/t1082-ac13-cache go test -json ./internal/cli -run '^(TestFactoryLiveClaudeCodexWorktreeHandoff|TestFactoryLaneHandoffLiveEvidenceGateRejectsMutants)$' -count=1 -timeout=240s > .moai/reports/t1082/ac13.jsonl && jq -se '([.[]|select(.Action=="pass" and .Test=="TestFactoryLiveClaudeCodexWorktreeHandoff")]|length)==1 and ([.[]|select(.Action=="pass" and .Test=="TestFactoryLaneHandoffLiveEvidenceGateRejectsMutants")]|length)==1 and ([.[]|select(.Action=="fail" or .Action=="skip")]|length)==0 and ([.[]|select((.Output//"")|contains("NOT_RUN"))]|length)==0' .moai/reports/t1082/ac13.jsonl && jq -e '(.schema_version==1) and (.card_id=="t1082") and (.spec_id=="SPEC-FACTORY-LANE-WORKTREE-HANDOFF-001") and (.mode=="headless") and (.contexts.separate==true) and (.contexts.lead.production_cli==true) and (.contexts.lead.real_model==true) and (.contexts.lane.production_cli==true) and (.contexts.lane.real_model==true) and (.contexts.lead.argv|type=="array" and length>0 and all(.[]; type=="string" and length>0)) and (.contexts.lane.argv|type=="array" and length>0 and all(.[]; type=="string" and length>0)) and (.contexts.lead.pid|type=="number" and .>1) and (.contexts.lane.pid|type=="number" and .>1) and (.contexts.lead.pid != .contexts.lane.pid) and (.contexts.lead.process_start|type=="string" and length>0) and (.contexts.lane.process_start|type=="string" and length>0) and (.contexts.lead.binary_sha256|type=="string" and test("^[0-9a-f]{64}$")) and (.contexts.lane.binary_sha256|type=="string" and test("^[0-9a-f]{64}$")) and (.app_server.method=="thread/fork") and (.app_server.request_cwd==.target.reserved_cwd) and (.app_server.returned_thread_id|type=="string" and length>0) and (.app_server.forked_from_id|type=="string" and length>0) and (.app_server.thread_started==true) and (.controller_readback.cwd==.target.reserved_cwd) and (.controller_readback.branch==.target.reserved_branch) and (.controller_readback.head==.target.pinned_head) and (.target.reserved_cwd|type=="string" and startswith("/")) and (.target.reserved_branch|type=="string" and startswith("WT-") and length>3) and (.target.pinned_head|type=="string" and test("^[0-9a-f]{40}$")) and (.session_start_wait_count==0) and (.empty_model_turn_count==0) and (.receipts.bound_id|type=="string" and length>0) and (.receipts.message_id|type=="string" and length>0) and (.old_endpoint.rejected==true) and (.old_endpoint.code=="STALE") and (.writes.pre_bound==0) and (.writes.wrong_cwd==0) and (.writes.primary_branch_switches==0) and (.messages.sent|type=="number" and .>0) and (.messages.received==.messages.sent) and (.messages.lost==0) and (.nonces.lead_to_lane|type=="string" and length>0) and (.nonces.lane_to_lead|type=="string" and length>0) and (.nonces.lead_to_lane != .nonces.lane_to_lead) and (.cleanup==true) and (.bypass.direct_register_peer==false) and (.bypass.db_seed==false) and (.bypass.mock==false) and (.bypass.fixture==false) and (.bypass.private_control==false)' .moai/reports/t1082/ac13-evidence.json
```

Expected outputs: 두 `jq` 모두 `true`. `.moai/reports/t1082/ac13-evidence.json`이 없거나 malformed이거나 위 exact predicate 중 하나라도 거짓이면 FAIL이다. stdout 문자열은 대체 증거가 아니다.

#### Common LIVE evidence-gate mutant rejection (AC-FLH-012/013)

**Given** one valid interactive evidence document, one valid stored-history headless `thread/fork` evidence document, and mutants with a required field removed, `fixture=true`, `mock=true`, `direct_register_peer=true`, a child `Action=fail`, a child `Action=skip`, or `wrong_method_thread_start` that changes the stored-history method to `thread/start` and clears `forked_from_id`, **when** the same production predicates used by AC-FLH-012/013 evaluate them, **then** both valid controls pass and every mutant fails closed. The test shall compare the actual predicate implementation, not a duplicate permissive assertion. No-history `thread/start` remains valid only under AC-FLH-004 and is not a valid AC-FLH-013 LIVE control.

This selector is mandatory inside both AC-FLH-012 and AC-FLH-013 commands. Its own parent PASS cannot override any global child/subtest/package `fail` or `skip`, any `NOT_RUN`, or missing/malformed card-scoped evidence.

## Required LIVE evidence shape

`.moai/reports/t1082/verdict.md`는 구현 이후에만 작성하며 다음을 포함해야 한다.

- Claim, exact command와 verbatim-output log path, measured HEAD/binary attribution, Gaps, Residual-risk.
- Per-AC PASS/FAIL. LIVE의 `NOT_RUN`, skip, auth/capacity blocker는 FAIL/GAP으로 남긴다.
- `ac12-evidence.json`과 `ac13-evidence.json`은 위 predicate가 검사하는 versioned schema를 그대로 사용한다. 자유 형식 log key/value는 판정 입력이 아니다.
- Run/lane/card/SPEC, reservation nonce digest, old/new endpoint IDs, generations, fork lineage, mode-specific binding evidence, BOUND receipt ID, dispatch/receipt IDs.
- Interactive 실제 `/cd`, next-normal-turn SessionStart, empty-turn count와 headless 실제 `thread/fork(cwd)`, returned thread ID, controller provenance readback, SessionStart-wait count.
- Target statusline screenshot/text, `pwd`, `git branch --show-current`, `git rev-parse HEAD`, primary branch before/after.
- Pre-BOUND code/commit/task-ACK counts, wrong-cwd write count, sent/received/lost message counts.
- Process cleanup evidence. Payload body와 credential은 기록하지 않는다.
