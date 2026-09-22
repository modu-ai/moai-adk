# SPEC Review Report: SPEC-FACTORY-LANE-WORKTREE-HANDOFF-001

Iteration: 2/3  
Verdict: FAIL  
Overall Score: 0.94  
Merge-blocking: YES

Reasoning context ignored per M1 Context Isolation. 이번 iteration은 이전 LIVE gate self-report finding의 수정 delta와 그 회귀면만 감사했다. 구현 코드는 읽거나 판단하지 않았다.

## Must-Pass Results

- [PASS] MP-1 REQ number consistency: `spec.md:L64-L122`에 REQ-FLH-001..015가 중복과 공백 없이 순차 존재한다. 재측정 출력은 `REQ_HEADINGS=15`다.
- [PASS] MP-2 GEARS format compliance — requirement layer: `spec.md:L64-L122`의 15개 REQ는 `While`/`Where`/`When` 또는 ubiquitous `The <subject> SHALL` 형식이다. 이 판정은 REQ layer에만 적용했고 `acceptance.md`의 Given-When-Then AC에는 적용하지 않았다.
- [PASS] MP-3 YAML frontmatter validity: `spec.md:L1-L13`에 canonical 12 fields가 올바른 타입으로 존재한다. `tier`, `depends_on`, `card`는 optional metadata다.
- [N/A] MP-4 language neutrality: `module: internal/factorymsg`인 Go/Codex factory 단일 제품 경계 SPEC이며 universal multi-language template가 아니다.
- [PASS] MP-5 D7 cross-SPEC reconciliation: `SPEC-FACTORY-MIXED-HOOK-001`가 존재하고 status는 `in-progress`다. retired/superseded/archived 참조가 아니므로 D7 BLOCKING finding은 없다.
- [PASS] MP-6 D8 cross-platform discipline: `spec.md`에 literal `syscall`이 없다.
- [PASS] MP-7 clarification gate: `rg -n '\[NEEDS CLARIFICATION' plan.md research.md` 출력이 없었다.
- [PASS] MP-8 RED-now re-execution: AC 16개와 공통 gate-quality selector 1개, 총 17개 exact `rg -n -F 'func Test…(' internal --glob '*_test.go'`를 현재 tree에서 각각 재실행했다. 모두 stdout 0 bytes, exit 1로 RED가 재현됐다. 문서-level tree pin은 `bf39a539d97f49edf3b11517ee7c982239c60df3`다.

## Category Scores

| Dimension | Score | Rubric Band | Evidence |
|---|---:|---|---|
| Clarity | 1.00 | 1.0 | interactive와 headless state/evidence가 `spec.md:L86-L94`, `acceptance.md:L182-L198`에서 명시적으로 분리된다. |
| Completeness | 1.00 | 1.0 | HISTORY/WHY/WHAT/HOW/15 REQ/16 AC/구체적 Out of Scope가 존재하며 Tier L 5개 산출물을 모두 읽었다. |
| Testability | 0.75 | 0.75 | structured evidence와 global runner rejection은 추가됐지만 AC-FLH-013 predicate가 wrong-method mutant를 허용한다(`acceptance.md:L192-L195`). |
| Traceability | 1.00 | 1.0 | 15개 REQ reference와 16개 AC reference의 unique 집합이 각 heading 집합과 정확히 일치한다(`spec.md:L124-L142`, `acceptance.md:L23-L42`). |

Tier L PASS threshold는 0.85다. 점수는 threshold를 넘지만 D2는 stated criterion의 correctness를 직접 깨는 blocking finding이므로 PASS로 상쇄되지 않는다.

## Defects Found

D2. FLH-LIVE-WRONG-METHOD-MUTANT — `acceptance.md:L192-L195`, `spec.md:L118` — AC-FLH-013은 stored history를 전제로 실제 `thread/fork(cwd)`와 fork lineage를 요구하지만 production predicate는 `thread/fork` **또는** `thread/start`를 허용한다. `method="thread/start"`, `forked_from_id=null`인 structured JSON이 exact predicate를 통과했다. 공통 mutant 계획(`acceptance.md:L200-L204`)에도 wrong-method mutant가 없다. Severity: major — Class: blocking — Confidence: High — Reachability: direct, exact predicate execution — Impact: no-history `thread/start` proof가 stored-history fork LIVE gate를 거짓 PASS시켜 history preservation과 lineage 검증 없이 run acceptance를 닫을 수 있다. — Required fix: AC-FLH-013 predicate를 `app_server.method=="thread/fork"` 및 non-empty `forked_from_id`로 고정하고, common production-predicate mutant test에 `thread/start`/null-lineage mutant를 추가한다. `thread/start` fallback은 AC-FLH-004에만 남긴다. — Merge-blocking: YES.

## Regression Check

이전 iteration의 defect:

- D1: LIVE gate가 parent PASS와 자유 형식 stdout sentinel만으로 fixture/mock/direct-registration을 통과시킬 수 있음 — [RESOLVED]: `acceptance.md:L20-L21,L185,L195,L200-L204`가 card-scoped structured JSON, exact typed predicates, 공통 production-predicate mutant selector를 요구한다. AC-FLH-012/013 commands는 정확히 두 parent PASS, global fail/skip 0, `NOT_RUN` 0을 요구한다.
- D1의 global runner 하위면 — [RESOLVED]: 문서 전체 16개 GREEN gate에 global `Action==fail or Action==skip`와 `NOT_RUN` exclusion이 각각 16회 존재한다. AC-FLH-012/013에는 공통 gate-quality parent PASS도 추가됐다.
- New regression D2 — [UNRESOLVED]: headless predicate의 `thread/start` 허용이 AC-FLH-013의 fork-only scope를 다시 넓힌다.

## Verified Non-Findings

- Traceability: `REQ_HEADINGS=15`, `AC_HEADINGS=16`; acceptance의 unique REQ refs는 REQ-FLH-001..015 전부이고 spec의 unique AC refs는 AC-FLH-001..016 전부다. orphan/uncovered/gap이 없다.
- Selector set: acceptance artifact의 unique `TestFactory…` selector는 17개다. 16 named AC selectors와 `TestFactoryLaneHandoffLiveEvidenceGateRejectsMutants`가 중복 없이 존재한다.
- Interactive LIVE predicate: card/SPEC/mode, separate real contexts, process IDs/start identities/binary hashes, actual `/cd`, `SWITCH_PENDING_INTERACTIVE`, next-normal-turn SessionStart, cwd/branch/HEAD equality, zero empty turn/write/loss, receipts/nonces/cleanup, bypass booleans을 exact typed predicate로 검사한다(`acceptance.md:L182-L188`).
- Global runner failures: child/subtest/package `fail`, 모든 `skip`, `NOT_RUN`, missing/malformed evidence는 AC-FLH-012/013을 포함해 fail closed하도록 문서화됐다(`acceptance.md:L15-L21,L185,L195,L204`).
- Common mutant selector: `plan.md:L94-L95`와 `acceptance.md:L200-L204`는 duplicate assertion이 아니라 production predicate 자체를 직접 호출해 missing/fixture/mock/direct-registration/child-fail/child-skip mutants를 거부하도록 요구한다.
- Capability-truth split은 유지된다: interactive는 다음 정상 turn SessionStart 전까지 pending이고, headless controller binding은 SessionStart/empty turn 없이 공식 RPC ID와 provenance readback으로 닫힌다.
- Worktree materializer/base pin/branch trace와 t1074 dependency/compatibility AC는 이번 delta에서 퇴행하지 않았다.

## Evidence

### Claim

이전 stdout-sentinel defect는 구조화 evidence와 공통 mutant gate로 수정됐지만, AC-FLH-013 exact predicate는 wrong-method mutant를 허용하므로 최종 판정은 FAIL이다.

### Evidence

1. 현재 baseline:

```text
command: git rev-parse HEAD && git branch --show-current && git status --short
output:
bf39a539d97f49edf3b11517ee7c982239c60df3
WT-factory-lane-worktree-handoff
?? .moai/specs/SPEC-FACTORY-LANE-WORKTREE-HANDOFF-001/
```

2. 구조/trace/selector 재측정:

```text
command: rg -c '^### REQ-FLH-[0-9]{3} ' spec.md; rg -c '^### AC-FLH-[0-9]{3} ' acceptance.md; rg -o 'TestFactory[A-Za-z0-9_]+' acceptance.md | sort -u | wc -l
output:
15
16
17
```

```text
command: rg -o 'REQ-FLH-[0-9]{3}' acceptance.md | sort -u
output: REQ-FLH-001 through REQ-FLH-015, all 15 unique values

command: rg -o 'AC-FLH-[0-9]{3}' spec.md | sort -u
output: AC-FLH-001 through AC-FLH-016, all 16 unique values
```

3. 17 RED-now selectors:

```text
commands: each ledger command `rg -n -F 'func <selector>(' internal --glob '*_test.go'`
observed output for every selector: <empty> (0 bytes)
observed exit for every selector: 1
selectors checked: 17/17, including TestFactoryLaneHandoffLiveEvidenceGateRejectsMutants
```

4. Global exclusion presence:

```text
command: rg -c 'Action=="fail" or \.Action=="skip"' acceptance.md
output: 16

command: rg -c 'contains\("NOT_RUN"\)' acceptance.md
output: 16
```

5. D2 exact false-pass reproduction. 아래 JSON은 AC-FLH-013의 다른 모든 exact field를 만족하지만 `app_server.method="thread/start"`, `forked_from_id=null`이다. `acceptance.md:L195`의 predicate를 그대로 적용했다.

```text
command: jq -ne --argjson evidence '{"schema_version":1,"card_id":"t1082","spec_id":"SPEC-FACTORY-LANE-WORKTREE-HANDOFF-001","mode":"headless","contexts":{"separate":true,"lead":{"production_cli":true,"real_model":true,"argv":["claude"],"pid":101,"process_start":"lead-start","binary_sha256":"aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"},"lane":{"production_cli":true,"real_model":true,"argv":["codex","app-server"],"pid":202,"process_start":"lane-start","binary_sha256":"bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb"}},"app_server":{"method":"thread/start","request_cwd":"/tmp/card","returned_thread_id":"thr-new","thread_started":true,"forked_from_id":null},"controller_readback":{"cwd":"/tmp/card","branch":"WT-card","head":"cccccccccccccccccccccccccccccccccccccccc"},"target":{"reserved_cwd":"/tmp/card","reserved_branch":"WT-card","pinned_head":"cccccccccccccccccccccccccccccccccccccccc"},"session_start_wait_count":0,"empty_model_turn_count":0,"receipts":{"bound_id":"bound-1","message_id":"msg-1"},"old_endpoint":{"rejected":true,"code":"STALE"},"writes":{"pre_bound":0,"wrong_cwd":0,"primary_branch_switches":0},"messages":{"sent":1,"received":1,"lost":0},"nonces":{"lead_to_lane":"n1","lane_to_lead":"n2"},"cleanup":true,"bypass":{"direct_register_peer":false,"db_seed":false,"mock":false,"fixture":false,"private_control":false}}' '$evidence | <the exact AC-FLH-013 predicate from acceptance.md:L195>'
stdout:
true
exit: 0
```

The executed predicate includes the literal permissive clause:

```jq
((.app_server.method=="thread/fork") or (.app_server.method=="thread/start"))
and (if .app_server.method=="thread/fork"
     then (.app_server.forked_from_id|type=="string" and length>0)
     else (.app_server.forked_from_id==null)
     end)
```

### Baseline-attribution

모든 local 측정은 `/Users/goos/MoAI/moai-adk-go/.claude/worktrees/t1082`, branch `WT-factory-lane-worktree-handoff`, HEAD `bf39a539d97f49edf3b11517ee7c982239c60df3`에서 이 iteration 중 수행했다. SPEC artifacts는 untracked 상태였다. 구현 파일 내용은 읽지 않았다. RED-now는 test 함수명 부재만 `rg`로 측정했다.

### Gaps

- 실제 LIVE run은 plan-phase이므로 실행하지 않았다. 이 감사는 acceptance gate가 요구 행동을 판별할 수 있는지 평가했다.
- t1074 dependency SPEC의 status는 `in-progress`다. 이 사실은 plan defect가 아니지만 strict depends_on pre-flight에서 completed 또는 명시적 override 전까지 run을 별도로 막는다.
- Iteration 1은 session interruption으로 export 파일을 남기지 못했다. 위 regression row는 그 turn에서 parent에게 전달된 binding finding을 기준으로 한다.

### Residual-risk

- D2를 고쳐도 structured JSON의 값이 실제 OS/process/RPC observation에서 생성됐는지는 run-phase test implementation과 evidence readback에서 확인해야 한다. 이 감사는 구현 코드를 보지 않았으므로 그 provenance를 PASS로 주장하지 않는다.
- `old_endpoint.code=="STALE"`는 `REQ-FLH-010`의 `STALE_ENDPOINT`/`STALE_GENERATION`보다 덜 구체적이다. 이번 delta의 binding defect는 fork-only false-pass로 한정했지만, 수정 시 exact error-code vocabulary도 함께 정렬하는 편이 안전하다.

## Recommendation

1. `acceptance.md:L195`의 AC-FLH-013 predicate에서 `thread/start` 분기를 제거하고 `thread/fork` + non-empty `forked_from_id`만 허용한다.
2. `acceptance.md:L202`와 `plan.md:L95`의 common mutant set에 `method=thread/start, forked_from_id=null` wrong-method mutant를 추가해 production predicate가 반드시 거부하도록 한다.
3. AC-FLH-004는 no-history `thread/start(cwd)` fallback 검증을 계속 소유하게 하여 일반 headless capability와 stored-history LIVE proof를 분리한다.
4. 같은 수정에서 stale code를 `STALE_ENDPOINT` 또는 `STALE_GENERATION` exact enum으로 좁히고, generation-bound receipt field를 predicate에 명시할지 검토한다.
