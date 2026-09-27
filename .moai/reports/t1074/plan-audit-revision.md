# SPEC Review Report: SPEC-FACTORY-MIXED-HOOK-001

Revision audit: 1/3
Verdict: **PASS**
Overall Score: **1.00**
Merge-blocking findings: **none**
Plan Artifact Hash: `947acb8b643c82e7ca346ce9e9ec8edafccc79b025809bfa6d0efee0fa308054`

Reasoning context ignored per M1 Context Isolation. 감사 범위는 `spec.md`, `plan.md`, `acceptance.md`, `design.md`, `research.md`, `operational-lane-status-addendum.md`, `progress.md`의 계획 계약뿐이며 구현 코드는 판정하지 않았다.

## Claim

수정된 계획은 `launch-pending`을 실제 launcher child PID/process-start 기반의 provisional 상태로만 취급하고, 첫 정상·빈 값이 아닌 사용자 turn에서 관측한 SessionStart 증거가 있을 때만 실제 session UUID로 `bound` 재결합한다. provisional 상태는 hook-bound delivery/receipt에 사용할 수 없고, 새 unit/LIVE gate와 기존 15 AC 회귀 gate는 child/subtest/package 범위의 `skip`·`fail` 및 해당 gate에서 요구하는 `NOT_RUN`을 전역 거부한다.

## Must-Pass Results

- [PASS] MP-1 REQ consistency: `REQ-FMH-001..012`와 `REQ-FMH-OPS-001..008`은 각 namespace에서 연속이고 중복이 없다. 기계 판독 결과 `req_defs=20 unique=20`, `legacy_seq=001..012`, `ops_seq=001..008`.
- [PASS] MP-2 GEARS requirement layer: `spec.md:54-100`의 요구와 `operational-lane-status-addendum.md:26-78`의 한국어 동등 규범형은 ubiquitous/event/state/capability 패턴을 사용한다. Given-When-Then AC는 verification layer로만 판정했다.
- [PASS] MP-3 YAML: `spec.md:2-14`에 canonical 12개 필드가 있고 strict lint는 exit 0이었다. `OwnershipTransitionUnmeasured` info 1건만 관측됐다.
- [N/A] MP-4 language neutrality: 단일 Go runtime/CLI 기능이며 범용 다중언어 template tooling이 아니다.
- [PASS] MP-5 D7: 본문 SPEC 참조는 자기 자신뿐이며 retired/superseded/archived 교차 참조가 없다.
- [PASS] MP-6 D8: `spec.md`에 `syscall`이 없다.
- [PASS] MP-7 clarification gate: `plan.md`와 `research.md`에 `[NEEDS CLARIFICATION` marker가 없다.
- [PASS] MP-8 revision RED-now: capability-truth 변경분의 두 release-blocking selector는 HEAD `8c5d9be993ad45d5c91ae8a95af904b27c935599`에서 각각 stdout `<empty>`, exit 1로 재현됐다(`acceptance.md:55-56`). 기존 6개 selector는 같은 HEAD에서 PRESENT임을 별도 행으로 사실대로 기록했고 현재 RED로 오표기하지 않는다(`acceptance.md:45-56`).

## Category Scores

| Dimension | Score | Evidence |
|---|---:|---|
| Clarity | 1.00 | `spec.md:56,92`; `design.md:31,43-47` |
| Completeness | 1.00 | `acceptance.md:43-126`; addendum `:160-178` |
| Testability | 1.00 | `acceptance.md:64-74,103,122`; addendum `:165,173,175` |
| Traceability | 1.00 | `acceptance.md:15-37`; uncovered/orphan 모두 없음 |

## Defects Found

No defects found.

## Delta Regression Check

- D1 RED ledger drift — **RESOLVED**: 현재 HEAD 실측이 기존 6개 PRESENT + 신규 capability-truth 2개 CURRENT RED로 정확히 분리됐다.
- D2 child/package skip false-pass — **RESOLVED**: acceptance/addendum의 JQ gate 20개가 모두 전역 `Action=skip`과 `Action=fail`을 거부한다. exact mutant는 generic `[false,false,true]`; OPS unit, OPS LIVE, 기존 10-AC regression의 child/package skip mutant는 각각 `false`였다.

## Verified Non-findings

- 요구↔AC 추적성: `req_defs=20 unique=20`, `ac_rows=42 unique=21`, `uncovered=` 빈 값, `orphans=` 빈 값. 42행은 요약표와 상세 정의의 중복 표기이며 AC ID는 21개다.
- Capability truth: `launch-pending`은 session UUID 미관측 상태이고 hook-bound delivery가 거부된다(`spec.md:56,92`; `acceptance.md:19`; addendum `:72-78,112-117`).
- Pending non-routability: 실제 SessionStart UUID/generation/process-start tuple 전에는 consume/acknowledge가 허용되지 않는다(`acceptance.md:19`; `design.md:31`).
- 공식 Codex 경계와 모순 없음: 공식 Hooks 문서는 non-managed/project hook의 trust/hash 검토와 idle 상태의 background completion이 새 turn을 시작하지 않음을 명시하고, App Server 문서는 `thread/start`를 thread 생성, `turn/start`를 user input과 generation 시작으로 구분한다. 계획은 pre-turn SessionStart를 추정하지 않고 launcher evidence를 provisional로 분리한다. 출처: [Codex Hooks](https://developers.openai.com/codex/hooks), [Codex App Server](https://developers.openai.com/codex/app-server/).

## Evidence

### Selector ledger re-execution

```text
TestFactoryLiveOperationalRosterBeforePrompt: exit=0; internal/cli/factory_operational_live_test.go:32
TestFactoryLaneRosterStateTruth: exit=0; internal/factorymsg/roster_test.go:34
TestFactoryLaneRosterProjectIsolation: exit=0; internal/factorymsg/roster_test.go:108
TestFactoryMsgStatusReadOnlyRoster: exit=0; internal/cli/mcp_factory_msg_test.go:121
TestFactoryLeadNoticeUsesOperationalStatus: exit=0; internal/cli/mcp_factory_msg_test.go:197
TestFactoryLiveOperationalLauncherChain: exit=0; internal/cli/factory_operational_live_test.go:37
TestFactoryLauncherRegistersLaunchPendingPeers: exit=1; stdout=<empty>
TestFactorySessionStartRebindsLaunchPendingPeer: exit=1; stdout=<empty>
```

### JQ gate and mutant checks

```text
jq_gates=20
global_skip_fail=PASS

generic mutant output:
[
  false,
  false,
  true
]
OPS unit child/package skip mutant: false
OPS LIVE child/package skip mutant: false
existing 10-AC regression child/package skip mutant: false
```

### Static checks

```text
git rev-parse HEAD
8c5d9be993ad45d5c91ae8a95af904b27c935599

git branch --show-current
WT-factory-mixed-hook

GOCACHE=/tmp/t1074-plan-audit-revision-final-cache go run ./cmd/moai spec lint SPEC-FACTORY-MIXED-HOOK-001 --strict --json
[{"severity":"info","code":"OwnershipTransitionUnmeasured",...}]
exit 0

git diff --check -- <seven audited plan artifacts>
stdout=<empty>
exit 0
```

## Baseline-attribution

- Worktree: `/Users/goos/MoAI/moai-adk-go/.claude/worktrees/t1074`
- Branch/HEAD: `WT-factory-mixed-hook@8c5d9be993ad45d5c91ae8a95af904b27c935599`
- Seven-artifact hash: `947acb8b643c82e7ca346ce9e9ec8edafccc79b025809bfa6d0efee0fa308054`
- 감사가 작성한 파일은 이 보고서 하나뿐이다.

## Gaps

- 구현 코드의 동작·품질은 검사하지 않았다.
- provider-backed Codex/Claude LIVE gate와 benchmark를 실행하지 않았다. 이 PASS는 계획 계약 판정이며 runtime PASS가 아니다.
- 공식 문서와 현재 계획의 의미 정합성만 확인했다. Codex 버전별 hook timing, MCP 재연결, process cleanup은 addendum §6 LIVE gate가 실제로 측정해야 한다.

## Residual-risk

- `launch-pending → bound`와 lead-owned restarted MCP 동작은 구현 후 exact LIVE gate가 통과하기 전까지 미검증이다.
- provider 또는 hook trust 상태 때문에 LIVE row가 실행되지 않으면 `skip`/`NOT_RUN`이 전역 거부되므로 완료 판정은 닫히지 않는다.

## Recommendation

계획은 구현/검증 단계로 진행 가능하다. 완료 판정은 addendum §6의 6-name unit gate, 두 production LIVE gate, 그리고 `acceptance.md:98-126`의 기존 15 AC current-baseline regression gate가 모두 실제 출력으로 통과한 뒤에만 내린다.
