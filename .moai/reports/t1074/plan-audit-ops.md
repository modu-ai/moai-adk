# SPEC Review Report: SPEC-FACTORY-MIXED-HOOK-001

Iteration: 1/3
Verdict: **FAIL**
Overall Score: **0.81**
Plan Artifact Hash: `72745ae25edf860cc3846317eda28e4df149e1dd27aeb6aff2eb4975ef9a54d1`
Auditor Version: `plan-auditor/v1`

Reasoning context ignored per M1 Context Isolation. 이 감사는 `spec.md`, `plan.md`, `acceptance.md`, `operational-lane-status-addendum.md`와 Tier L 필수 입력인 `design.md`, `research.md`만 판단 근거로 사용했다.

## Claim

기존 12개 REQ와 15개 AC의 구조는 보존되었고, 신규 OPS 8개 REQ와 6개 AC의 정적 추적성도 완결되어 있다. 그러나 신규 live 계약은 사용자가 요구한 정확한 production launcher 경로를 강제하지 않으며, release-blocking OPS AC의 RED-now 원장과 변경 영향 회귀 게이트가 없다. 따라서 현재 계획은 구현 착수/완료 판정에 사용할 수 없다.

## Must-Pass Results

- [PASS] MP-1 REQ number consistency: `spec.md:53-99`의 `REQ-FMH-001..012`와 `operational-lane-status-addendum.md:24-75`의 `REQ-FMH-OPS-001..008`은 각 namespace에서 연속이고 중복이 없다.
- [PASS] MP-2 GEARS format compliance: requirement layer만 판정했다. 기존 REQ는 `SHALL` 기반 ubiquitous/event/state/capability 문장이고(`spec.md:55-99`), OPS REQ는 한국어의 동등한 `호출되면 ... 반환해야 한다`, `... 해야 한다`, `... 해서는 안 된다` 규범형이다(`operational-lane-status-addendum.md:26-75`). AC의 Given-When-Then은 이 판정에 포함하지 않았다.
- [PASS] MP-3 YAML frontmatter validity: strict SPEC lint가 error/warning 없이 informational `OwnershipTransitionUnmeasured` 1건만 반환했다. canonical 12개 필드는 `spec.md:2-14`에 존재한다.
- [N/A] MP-4 Section 22 language neutrality: Go 기반 MoAI runtime/CLI 기능이며 범용 다중언어 템플릿 도구 요구가 아니다.
- [PASS] MP-5 D7 cross-SPEC reconciliation: `spec.md`에서 수집된 SPEC ID는 자기 자신 `SPEC-FACTORY-MIXED-HOOK-001`뿐이며 status는 `in-progress`다. retired/superseded/archived 참조가 없다.
- [PASS] MP-6 D8 cross-platform discipline: `rg -n 'syscall' spec.md` 출력이 없었다. D8 syscall/build-tag 조건이 발동하지 않는다.
- [PASS] MP-7 clarification gate: `plan.md`와 `research.md`에서 `[NEEDS CLARIFICATION` 검색 결과가 없었다.
- [FAIL] MP-8 RED-now cell re-execution: `acceptance.md:53-73`은 구 tree `758314007...`의 기존 15개 AC만 기록한다. 현재 tree에서는 15개 selector가 모두 존재해 기존 `<empty>`, exit 1 RED가 재현되지 않았고, 신규 OPS 6개는 RED-now 행 자체가 없다. 신규 예정 test selector 5개는 현재 모두 exit 1이지만 그 stdout/exit/current tree SHA가 criterion별 RED-now cell로 문서화되지 않았다.
- [PASS] MP-9 cross-artifact ordering consistency: canonical 계획은 `M1 → M2 → M3 → M4 → M5`, M5가 addendum의 내부 `M1 → M2 → M3 → M4`를 순서대로 실행하도록 명시한다(`plan.md:55-60`, `operational-lane-status-addendum.md:96-122`). 검색된 before/after 후보는 prompt 전 등록과 assertion 후 종료 순서를 요구할 뿐 milestone 배치와 충돌하지 않는다.

## Category Scores

| Dimension | Score | Rubric Band | Evidence |
|---|---:|---|---|
| Clarity | 0.75 | 대부분 명확하나 핵심 live launcher 경로가 한 가지로 고정되지 않음 | `operational-lane-status-addendum.md:114-122,126-150` |
| Completeness | 0.75 | 필수 Tier L 문서와 OPS 범위는 있으나 RED-now 및 기존 15 AC 회귀 계획 누락 | `acceptance.md:53-73`, `plan.md:55-70` |
| Testability | 0.75 | OPS 6개 모두 Given-When-Then이나 exact production launcher/process-chain을 gate가 판별하지 못함 | `operational-lane-status-addendum.md:126-168` |
| Traceability | 1.00 | 기계 판독 결과 `COLLECTED: 20 REQ definitions, 20 mapped IDs`; uncovered/orphan 없음 | `acceptance.md:15-41` |

단순 평균: `(0.75 + 0.75 + 0.75 + 1.00) / 4 = 0.8125`, 보고 점수 `0.81`. Tier L PASS 기준 `0.85` 미달이며 MP-8 firewall도 실패했다.

## Defects Found

D1. **OPS-LIVE-PRODUCTION-PATH** — `operational-lane-status-addendum.md:114-122,126-150,162-168` — M4와 AC-FMH-OPS-001/006은 “실제 세션”, 설치 binary, MCP 재시작 및 sentinel 출력만 요구한다. 정확히 한 개의 `moai codex -f` lead와 두 개의 `moai codex -f agent` worker를 production launcher로 시작했다는 argv/process-tree 증거, launcher가 만든 실제 Codex owner PID와 SessionStart 등록 `(pid, process_start)`의 일치, lead의 MCP가 같은 endpoint를 해석했다는 증거가 없다. 따라서 direct `codex exec`, 외부 `MOAI_SESSION_PID` 주입, 별도 owner process 같은 우회가 gate의 sentinel을 만족할 수 있다. — Severity: **critical** — Class: **blocking** — Required fix: live 계획과 AC에 exact 세 launcher argv, direct-Codex 금지, `MOAI_SESSION_PID`/수동 `RegisterPeer`/pre-seed 금지, 실제 Codex parent PID·fingerprint와 SessionStart peer row의 동일성, lead 세션 MCP에서 status 호출, 세 process tree와 종료 정리를 필수 증거로 추가한다.

D2. **OPS-RED-NOW-MISSING** — `acceptance.md:13,39-41,53-73`; `operational-lane-status-addendum.md:152-168` — 모든 AC가 MUST-PASS지만 OPS 6개에 current-tree RED-now cell이 없다. 기존 15개 원장은 구 tree에 고정되어 현재 tree에서 RED가 재현되지 않는다. addendum가 실행 게이트를 GREEN 명령으로만 제시하고 “현재 구현 전에는 실행 대상이 아니다”라고 둔 것은 MP-8의 plan-phase RED 재실행 계약을 충족하지 못한다. — Severity: **critical** — Class: **blocking** — Required fix: 현재 subject tree SHA에 대해 AC-FMH-OPS-001..006 각각의 read-only RED-now command, verbatim stdout, exit code를 기록하고 재실행 가능한 selector-to-AC 매핑을 추가한다. 기존 15개가 M5의 historical regression set이라면 release-blocking delta와 regression 분류를 명시적으로 분리한다.

D3. **LEGACY-15-REGRESSION-GATE-ABSENT** — `plan.md:55-70`; `acceptance.md:13-41` — M5는 `factorymsg`, CLI, MCP 등록, SessionStart/hook을 변경하면서 기존 AC-FMH-001..015를 “historically attributed”로만 둔다. addendum §6의 새 selector 다섯 개만 실행하므로 기존 12 REQ/15 AC의 현재-baseline 회귀를 판별하지 못한다. 과거 PASS는 새 변경 기준선의 PASS가 아니다. — Severity: **major** — Class: **blocking** — Required fix: M5 exit gate에 기존 15 AC의 exact non-empty regression matrix를 추가하고, live criterion이 환경상 실행되지 않으면 PASS로 상속하지 말고 FAIL/GAP으로 남긴다. 최소한 변경 영향 패키지의 기존 AC selector 전부와 live production-chain 축을 현재 binary/commit에 재귀속한다.

## Evidence

### E1 — 대상 고정과 hash 안정성

Command:

```text
git rev-parse HEAD
git branch --show-current
shasum -a 256 spec.md plan.md acceptance.md operational-lane-status-addendum.md | shasum -a 256
```

Observed output at audit start and immediately before report export:

```text
99d77dd25d3f2b1f6bae68d5fbccb5f8f4cf3bca
WT-factory-mixed-hook
72745ae25edf860cc3846317eda28e4df149e1dd27aeb6aff2eb4975ef9a54d1  -
```

### E2 — REQ/AC 수와 추적성

Command: heading enumeration plus an awk mapping of requirement headings to `acceptance.md` AC table references.

Observed output:

```text
COLLECTED: 20 REQ definitions, 20 mapped IDs
legacy_headings=15
ops_headings=6
```

No `UNCOVERED`, `ORPHAN`, or `DUPLICATE` line was emitted.

### E3 — RED-now 재실행

Command: each exact legacy selector from `acceptance.md:59-73` and each planned OPS selector from addendum §6 was searched with `rg -n -F 'func <TestName>(' internal`.

Observed output summary preserving every selector result:

```text
legacy selectors (15/15): rc=0, each existing test definition printed
TestFactoryLaneRoster rc=1
TestFactoryMsgStatusReadOnlyRoster rc=1
TestFactorySessionStartRegistersBeforePrompt rc=1
TestFactoryLeadNoticeUsesOperationalStatus rc=1
TestFactoryLiveOperationalLaneRoster rc=1
```

The existing ledger's claimed legacy `<empty>`/exit 1 does not reproduce on this tree. The new OPS absence does reproduce mechanically, but no OPS criterion carries the required RED-now cell.

### E4 — strict lint and clarification/cross-platform checks

Command:

```text
GOCACHE=/tmp/t1074-ops-plan-audit-cache go run ./cmd/moai spec lint SPEC-FACTORY-MIXED-HOOK-001 --strict --json
rg -n '\[NEEDS CLARIFICATION' plan.md research.md
rg -n 'syscall' spec.md
git diff --check -- spec.md plan.md acceptance.md operational-lane-status-addendum.md
```

Observed output:

```json
[{"line":1,"severity":"info","code":"OwnershipTransitionUnmeasured","message":"SPEC SPEC-FACTORY-MIXED-HOOK-001 transition \"draft\" → \"in-progress\" expected owner \"manager-develop\" but commit cb099897a4ee802b5a2478faf1c0159f5b2c3099 (feat(t1074): M1 bind canonical factory runs and peers) has no Authored-By-Agent trailer — ownership transition unmeasured"}]
```

Lint exit `0`; clarification and `syscall` searches emitted no output; targeted `git diff --check` emitted no output and exit `0`.

## Baseline-attribution

- Worktree: `/Users/goos/MoAI/moai-adk-go/.claude/worktrees/t1074`
- Branch/HEAD: `WT-factory-mixed-hook@99d77dd25d3f2b1f6bae68d5fbccb5f8f4cf3bca`
- Audited subject: canonical `spec.md`, `plan.md`, `acceptance.md`, normative `operational-lane-status-addendum.md`; Tier L context `design.md`, `research.md` was also read.
- Subject hash: `72745ae25edf860cc3846317eda28e4df149e1dd27aeb6aff2eb4975ef9a54d1`.
- The pre-existing `verdict.md` and implementation files were read-only and were not modified by this audit.

## Gaps

- Plan audit only: no broad package test or live Codex/Claude session was run.
- The OPS implementation does not exist, so no claim is made that the production launcher chain currently passes or fails at runtime.
- Windows/Linux process-tree behavior and an actual external user project were not executed; the required separate-project proof remains a run-phase obligation.

## Residual-risk

- Even after roster visibility is implemented, `task_state` remains truthfully `unknown` until an explicit activity observation source exists. This audit does not reinterpret `live` process state as task `idle`/`busy`.
- The exact installed Codex hook/MCP parent chain can drift with Codex versions; the repaired live gate must record the tested Codex version and process tree.

## Iteration history

- Iteration 1: initial OPS addendum audit. No prior `plan-audit-ops.md` existed.

## Recommendation

1. Close D1 by pinning the live test to exact production launcher argv and real parent/SessionStart/MCP identity evidence.
2. Close D2 with current-tree criterion-level RED-now cells for all six OPS ACs.
3. Close D3 with a non-empty current-baseline regression gate for the existing 12 REQ/15 AC set.
