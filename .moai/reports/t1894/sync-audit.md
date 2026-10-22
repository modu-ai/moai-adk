auditor-model: claude-opus-5-5
verdict: FAIL
overall_score: 0.80
must_pass_failed: 1
blocking_count: 1
audited_sha: f7cc3f8ee8c65061382a29c40cf7cc809c1cb7b2
required_backend: codex fail
convergence_overall: fail

# 동기화 단계 감사 — SPEC-CARD-MODE-JUDGMENT-001 (카드 t1894, Tier L, 문서 전용)

감사 대상 트리는 `WT-fail-safe-serial` 브랜치의 HEAD `f7cc3f8ee8c65061382a29c40cf7cc809c1cb7b2`이며 작업 트리는 깨끗했다(`git status --short` 출력 없음). 기준 커밋은 `28c7bb8d36`이다. 호출자의 설명은 검증할 주장 목록으로만 취급했고, 판정은 디스크에서 다시 측정한 결과에만 근거한다. Tier L 통과 기준선은 0.85이며 이번 점수는 그 아래다.

## 1. 판정 요약

판정은 FAIL이다. 근거는 두 가지다.

1. 차단 결함 S1: 새 빈 레인(free lane) 정의가 같은 문서 다섯 줄 아래의 카드 전환 규칙과 정면으로 충돌한다(`factory-dispatch-detail.md` 134행 대 139행). 필수 게이트인 codex 백엔드도 같은 결함을 독립적으로 보고했고(신뢰도 높음), 감사자가 직접 재현했다.
2. 필수 백엔드 codex가 fail을 냈다(`audit_multi` overall_verdict `fail`, 영수증 `rcpt-aeda7d3bfadbf2012f680998`). 필수 게이트가 fail이면 PASS를 낼 수 없다.

수용 기준 원장은 AC-MJ-020(b) 한 칸을 빼고 모두 재현됐다. AC-MJ-020(b)는 감사 SHA에서 `unexpected-paths 1 / CHANGELOG.md`로 빨갛다(S2, 비차단). 범위가 지정된 Go 가드 20개는 이 트리에서 컴파일한 빌드로 모두 통과했다.

## 2. 4차원 평가

| 차원 | 점수(0..1) | 판정 | 근거 |
|---|---|---|---|
| Functionality (40%) | 0.70 | FAIL (must-pass) | 원장 재현은 AC-MJ-020(b) 외 전부 일치. 그러나 핵심 산출물인 빈 레인 정의가 같은 문서의 전환 규칙과 모순(S1, 차단). REQ-MJ-006이 지우라고 한 옛 정의("이전 카드가 `done`에 도달")의 규칙이 139행에 다른 문장으로 살아 있다 |
| Security (25%) | 0.95 | PASS | 문서 전용, Go 소스 변경 0개, 추가 행 비밀정보 탐침 실제 적중 0건(아래 증거 참조) |
| Craft (20%) | 0.82 | PASS | Go 가드 14+6 통과, 빈 스윕 0건. CHANGELOG 과장(S3), AC-MJ-020(b) 범위 결함(S2) |
| Consistency (15%) | 0.78 | PASS(경계) | 템플릿·배포 쌍 바이트 동일, 선언된 포크 1행 유지, `git diff --check` 깨끗. 그러나 M의 폴링 금지 문장과 A §14 일회성 재무장의 충돌(S4, codex 보고, R-9로 공개됨), 실행 정의의 대기 목록 불일치(S5) |

가중 합은 0.4×0.70 + 0.25×0.95 + 0.2×0.82 + 0.15×0.78 = 0.80이다. must-pass 방화벽에 따라 Functionality가 기준에 못 미치므로 다른 점수와 무관하게 전체 판정은 FAIL이다.

## 3. 결함 원장

### S1 [BLOCKING] [blocking] 빈 레인 정의와 카드 전환 경계의 모순 — 신뢰도 높음

- 위치: `internal/template/templates/.claude/rules/moai/workflow/factory-dispatch-detail.md` 134행(신규)과 139행(기존, 미변경). 배포 쌍 `.claude/rules/moai/workflow/factory-dispatch-detail.md`도 동일. 관련 문장: `factory-dispatch.md` 158행("At `done`, evidence read, … the lane moves on by the transition below").
- 내용: 134행은 "a lane that holds only merge-ready, held (operator-wait), or serial-waiting cards is a free lane"이라 정의하고, 133행 Routing 문장은 리더가 카드를 빈 레인에 보낸다고 한다. 그런데 139행은 "A factory lane leaves its card worktree (`ExitWorktree(keep)`) once its card reaches `done`, before the next card is routed to it"이라고 한다. merge-ready 카드를 쥔 레인은 134행에 따르면 다음 카드를 받을 수 있고, 139행에 따르면 그 카드가 `done`이 되기 전에는 받을 수 없다. 두 지시를 동시에 지킬 방법이 없다.
- 왜 차단인가: 이 SPEC의 산출물은 문서 자체이고, 빈 레인 정의는 REQ-MJ-006과 d314 위임 승격(REQ-MJ-007 (b))의 전제다. REQ-MJ-006은 옛 정의 "free means the lane's previous card reached `done`"가 남지 않아야 한다고 요구하는데, 139행은 그 규칙을 다른 말로 그대로 싣고 있다. 어휘 기반 셀은 이를 볼 수 없다고 SPEC 스스로 밝혔고(spec R-8: "the sync audit reads for a paraphrase"), 바로 그 읽기에서 걸렸다. `research.md`·`design.md`·`spec.md`에서 `transition boundary`, `reaches .done`, `ExitWorktree(keep)`를 찾았지만 한 건도 없어, 계획 단계가 이 사이트를 다루지 않았음이 확인된다.
- 재현: 아래 증거 E-S1의 grep 출력 134·135·139행.
- 교차 확인: codex(필수 게이트) P2, "새 빈 레인 정의가 기존 카드 전환 조건과 충돌합니다 — factory-dispatch-detail.md:134", 신뢰도 높음. 감사자 재현 결과와 일치한다.
- 필수 수정: 문장 목록이 닫힌 세계 매니페스트로 묶여 있으므로 run 단계가 직접 고칠 수 없다. manager-spec이 매니페스트를 개정해 D 139행(필요하면 `factory-dispatch.md` 158행도)을 새 정의와 맞춘다. 예를 들어 "카드가 실행을 멈추는 시점(merge-ready·held·serial-waiting 포함)에 `ExitWorktree(keep)`로 그 트리를 보존한 채 떠나고, 착지 슬롯에서 `factory next --card`로 다시 임대해 돌아온다"는 절차를 적는다. 개정 행은 해당 E-셀(E-005d 등)과 E-020d 해시와 함께 갱신하고, 정정 커밋 뒤 이 결함만 델타로 재감사한다.

### S2 [SHOULD-FIX] [optional] AC-MJ-020(b)가 감사 SHA에서 빨강 — 기준 범위 결함, 처분 기록 없음 — 신뢰도 높음

- 위치: `acceptance.md` AC-MJ-020 (b), E-020b. 동기화 커밋 `ea7b93c085`가 `CHANGELOG.md`를 추가했다.
- 내용: E-020b는 원장에 `unexpected-paths 0`, exit 0으로 기록돼 있지만 HEAD `f7cc3f8ee8`에서는 `unexpected-paths 1` / `CHANGELOG.md`, exit 1이 나온다. 이 기준이 허용하는 네 경로(`.moai/specs/SPEC-CARD-MODE-JUDGMENT-001/`, `.moai/reports/`, `.moai/state/`, `.moai/logs/`)에 sync 단계의 필수 산출물인 CHANGELOG가 빠져 있다. run 단계 마지막 커밋까지는 CHANGELOG 변경이 없었다(`git diff --name-only 28c7bb8d36 0638768424 | grep -c CHANGELOG` → `0`). §E.4는 이 빨강을 언급하지 않는다.
- 비차단인 이유: AC-MJ-020이 지키는 요구사항(REQ-MJ-006·-007·-012·-017)은 어느 것도 CHANGELOG 한 줄로 위반되지 않는다. 추가된 경로는 이 SPEC 자신의 [Unreleased] 항목 한 줄뿐이다(아래 E-S2). 결함은 기준의 범위 문장에 있다.
- 필수 수정: 닫기 기록(§E.4 또는 정정 커밋의 증거)에 "AC-MJ-020(b)는 sync 산출물 CHANGELOG.md 때문에 닫는 SHA에서 `unexpected-paths 1`이며, 이는 run 범위 밖"이라는 처분을 남기거나, manager-spec이 (b)의 면제 목록에 sync 단계의 CHANGELOG 항목을 명시한다.

### S3 [SHOULD-FIX] [optional] CHANGELOG의 과장: core 스텁에 Mode judgment 포인터가 있다고 적음 — 신뢰도 높음

- 위치: `CHANGELOG.md` [Unreleased] > Added 첫 항목.
- 내용: "(`factory-dispatch-cards.md` § Mode judgment, with one-line pointers in `factory-dispatch-core.md`, `factory-dispatch.md`, and the gtd workflow)". 실제로 `factory-dispatch-core.md`에는 `Mode judgment`가 0회 나온다(E-S3). K에서 바뀐 것은 재조정 절의 첫 구("A reconciliation is named")와 크론 범위 절뿐이다.
- 필수 수정: CHANGELOG 문장에서 `factory-dispatch-core.md`를 빼거나, core 스텁의 실제 변경(재조정 절 문구, 크론 범위)을 정확히 적는다. 17 REQ / 20 AC 개수, 파일 경로, 실행 커밋 목록, 카탈로그 해시·원장 행(ALB-0276, ALB-0257)은 디스크와 일치함을 확인했다.

### S4 [SHOULD-FIX] [optional] M의 "슬롯·트리거 대기에 폴링 크론 없음"과 A §14 재무장 일회성 재확인의 충돌 — 신뢰도 중간

- 위치: `factory-dispatch-mechanics.md` § Lease renewal 넷째 고정 행(신규)과 `auto-semantics.md` §14(같은 문장을 이번 변경이 편집함), `moai-lane-watchdog/SKILL.md` 121행.
- 내용: M은 "a lane waiting on a slot or a trigger arms and keeps no polling cron for that wait, and the leader sends a notice"라고 한다. A §14는 "While a wait whose `waiting_on` names the leader is open, the lane keeps a one-shot recheck … re-armed by each fire that finds the wait still open"이라고 한다. 슬롯 대기를 리더의 통지를 기다리는 대기로 기록하면(`waiting_on`=leader) 5분마다 재무장되는 일회성 크론이 사실상 폴링이 된다.
- 비차단인 이유: SPEC이 잔여 위험 R-9로 미리 공개했고("this SPEC leaves the one-shot unchanged … the operator may rule on it separately"), 슬롯 대기의 `waiting_on`을 리더가 아닌 슬롯으로 읽으면 두 문장은 겹치지 않는다. 정책 판단이 운영자 몫으로 남아 있다.
- 교차 확인: codex P2 "슬롯 대기의 폴링 금지가 기존 반복 재확인 절차와 충돌합니다 — factory-dispatch-mechanics.md:36". 감사자가 두 문장을 직접 읽어 재현했다.
- 필수 수정: 운영자 판정 후 별도 카드에서 §14와 watchdog §3.1에 "슬롯·트리거 대기는 재무장 대상이 아니며 리더 통지로 재개한다"를 명시하거나, 일회성 재확인이 a672의 금지 밖이라는 판정을 기록한다.

### S5 [MINOR] [optional] 실행(executing) 정의의 제외 목록과 A §14 괄호절의 어긋남 — 신뢰도 중간

- 위치: `factory-dispatch-mechanics.md` § Lease renewal 첫 행("… and is not waiting on a slot, a trigger, or the operator")과 `auto-semantics.md` §14("(an open wait on the leader is not execution)").
- 내용: 정의의 홈(M)은 리더 대기를 제외 목록에 넣지 않았고, A는 리더 대기를 실행이 아니라고 한다. 리더 판정을 기다리는 카드의 임대 갱신 여부와 만료 무해성(M은 merge-ready·held·serial-waiting만 무해하다고 함)이 정의되지 않는다.
- 필수 수정(선택): M의 정의에 "or the leader"를 더하거나 A 괄호절을 M 정의로 가리키게 한다.

### 공개된 기존 잔여(새 결함 아님)

- R-6: `factory-dispatch.md` 40·42행의 [HARD] "outside `--auto` never pick, reorder, or silently promote"와 바로 아래 44행의 위임 승격 포인터가 공존한다. Go 교리 테스트가 옛 문장을 고정하고 있어 의도적으로 남겼다고 SPEC이 밝혔다. 독자 혼란 위험은 그대로다.

## 4. 위임 승격 하위 절의 문언 검토 (d314 대비)

E-006이 고정한 여덟 행을 d314 원문과 하나씩 대조했다.

| d314 제한 | D § Delegated-dispatch promotion | 판단 |
|---|---|---|
| (a) operator-away / standing delegation 동안 | "while the operator has delegated dispatch (operator away, or a standing delegation)" | 일치 |
| (b) live lane이 free(실행 카드 없음) | "only to a free live lane, as defined under Routing above" | 일치(다만 Routing의 정의 자체가 S1로 139행과 모순) |
| (c) parallelizable로 명시 분류된 카드 | "only queued cards explicitly classified `parallelizable` by the mode judgment" | 일치 |
| (d) 큐 우선순위·준비도 우선 | "queue priority and readiness first" | 일치 |
| (e) keep-set·held·[보류 표시 카드는 자동 승격 안 함 | "a keep-set card, a held card, or a [보류-marked card" | 일치(`[보류`는 기존 템플릿 5개 파일에 이미 존재) |
| (f) serial 카드는 슬롯 규칙을 따름 | "outside this case; a serial card follows the slot rules" | 더 좁게 읽힘(확장 아님) |
| (g) 레인 자기 배정 불변 | "lane self-dispatch through `moai factory next`" | 일치 |
| 귀결(제한 아님) | "Consequence, not a further limit: … only while dispatch is delegated, the queued `parallelizable` cards" | 일치 |

확장은 발견하지 못했다. 첫 문장 "Promotion remains an act of the operator. One further case is stated here"는 F 44행 포인터, G 312행 포인터, M 191행 표 행과 같은 방향이고, 어느 포인터도 홈 토큰이나 자체 제한을 싣지 않는다(E-016이 정확히 세 홈만 나열). Mode judgment 하위 절은 0c51 원문의 조건("same files/shared state as an in-flight card or a slot-gated surface", reason에 파일 집합·충돌 확인·serial이면 충돌 카드 이름, 기본값 발급 금지, 재발급, Class C는 pick 때 `--spec`)을 빠짐없이 옮겼고, 재분류 동사는 언급하지 않는다(E-010a 빈 출력).

## 5. 증거 기반 보고 (Claim / Evidence / Baseline-attribution / Gaps / Residual-risk)

### Claim

| # | 주장 | 결과 |
|---|---|---|
| C1 | AC-MJ-001..013, -016..019, -020(a)(c)가 HEAD `f7cc3f8ee8`에서 재현된다 | 확인 |
| C2 | AC-MJ-020(b)가 HEAD에서 재현된다 | 반증 — `unexpected-paths 1` (S2) |
| C3 | 마일스톤 커밋마다 템플릿·배포 쌍이 함께 실린다(pairs 3/5/4/1) | 확인 |
| C4 | `factory-dispatch.md` 쌍의 차이는 기준 커밋에도 있던 선언된 포크 1행뿐이다 | 확인 |
| C5 | core 스텁 예산 `OK 9997` | 확인 |
| C6 | AC-MJ-014 Go 가드 14+6개가 이 트리에서 통과하고 빈 스윕이 없다 | 확인 |
| C7 | 개정된 교리가 이웃 규칙과 모순되지 않는다 | 반증 — S1(차단), S4, S5 |
| C8 | CHANGELOG 항목이 정확하다 | 부분 반증 — S3 |
| C9 | spec.md `status: completed`, 3-phase close 형태, §E.4 `sync_commit_sha`가 실제 동기화 커밋이다 | 확인 |
| C10 | run 단계 차단과 중간 개정이 정직하게 기록됐다 | 확인 |

### Evidence

원장 재실행. `acceptance.md` §2의 모든 `command:`/`commands:` 셀을 수정 없이 `/bin/sh -c`로 실행했다(PATH 맨 앞 `/usr/bin`, cwd는 워크트리 루트). 판정에 쓴 출력 줄(전부 이번 실행, 이 트리):

```text
E-001  exit=0 → 8 / 8
E-002a exit=0 → (빈 출력)
E-002b exit=0 → 74ad197380…:internal/template/templates/.claude/skills/moai/workflows/gtd.md:52:`moai todo add "<description>"`; a bare `/moai todo` runs `moai todo list`.
E-002c exit=0 → 여섯 경로
E-002d exit=0 → …/gtd.md:1 , .claude/skills/moai/workflows/gtd.md:1
E-003  exit=0 → …/factory-dispatch.md:1 ×2
E-004a exit=0 → :1 ×2 ; E-004b exit=1 → :0 ×2 ; E-004c exit=0 → :2 ×2 ; E-004d exit=0 → :1 ×2 ; E-004e exit=0 → :2 ×2
E-005a exit=0 → :1 ×2 ; E-005b exit=1 → :0 ×2 ; E-005c exit=0 → :1 ×2 ; E-005d exit=0 → :3 ×2
E-006  exit=0 → 8 8 1 1 / 8 8 1 1
E-007  exit=0 → 4 4 1 / 4 4 1
E-008a exit=0 → (빈 출력)
E-008b exit=0 → core:1 factory-dispatch.md:2 auto-semantics.md:2 SKILL.md:1 (양쪽 트리)
E-008c exit=0 → core:1 factory-dispatch.md:1 (양쪽) ; E-008d exit=0 → auto-semantics.md:1 ×2 ; E-008e exit=0 → 2 / 2 ; E-008f exit=0 → 1 / 1
E-009  exit=0 → 1 1 0 / 1 1 0
E-010a exit=1 → (빈 출력) ; E-010b exit=0 → :1 ×2 ; E-010c exit=0 → 1
E-011  일곱 cmp 모두 exit=0, 출력 없음
E-011f exit=1 → … differ: char 23761, line 226
E-011g exit=1 → 9579676013 pairs 0 unpaired 7 ; E-011h exit=0 → 1969709ce6 pairs 2 unpaired 0 expected 2
E-012a exit=0 → OK 9997 ; E-012b exit=1 → OVER 26434
E-013a exit=1 → 여덟 파일 모두 :0 ; E-013b exit=0 → 8 ; E-013c exit=0 → auto-semantics.md:2 factory-dispatch-detail.md:2 gtd.md:1
E-016  exit=0 → factory-dispatch-cards.md, factory-dispatch-detail.md, factory-dispatch-mechanics.md
E-017  exit=1 →  1 file changed, 1 insertion(+), 1 deletion(-)
E-018a exit=0 → :1 ×2 ; E-018b exit=1 → :0 ×4 ; E-018c exit=0 → :1 ×4 ; E-018d exit=0 → :1 ×2 ; E-018e exit=0 → :1 ×2
E-019a exit=0 → :1 ×2 ; E-019b exit=1 → 0 ; E-019c exit=0 → :1 ×2
E-020a exit=0 → C/K/F/D/M/A/G/W 모두 "template 0 0 0 0 0 deployed 0 0 0 0 0"
E-020b exit=1 → unexpected-paths 1
                CHANGELOG.md
E-020c exit=1 → unexpected-paths 200 ; E-020d exit=0 → same same ; E-020e exit=1 → DIFFERENT same
E-020f exit=1 → unexpected-paths 1 (.moai/policies/workflow/factory-dispatch-detail.md) ; E-020g exit=1 → unexpected-paths 1 (.moai/workflows/gtd.md)
E-020h exit=1 → C good 0 0 pair-move 44 44
```

E-011f의 문자·행 위치(원장 기록 `char 23127, line 222`)와 E-012b 값(기록 `OVER 25803`)은 run이 행을 더해 이동한 것이며, 두 셀 모두 알려진 실패 입력으로서 여전히 빨갛다.

커밋 그래프 쌍 확인:

```text
$ python3 -I .moai/specs/SPEC-CARD-MODE-JUDGMENT-001/closed-world-check.py pairs .moai/specs/SPEC-CARD-MODE-JUDGMENT-001/closed-world-manifest.txt <sha>:<n>
2fb0be5ed8 pairs 3 unpaired 0 expected 3   exit=0
89eb51de2f pairs 5 unpaired 0 expected 5   exit=0
770b5ae654 pairs 4 unpaired 0 expected 4   exit=0
05bcdc3a8f pairs 1 unpaired 0 expected 1   exit=0
```

선언된 포크(기준 커밋에서도 같은 1행 차이):

```text
$ git diff --shortstat 28c7bb8d36:internal/template/templates/.claude/rules/moai/workflow/factory-dispatch.md 28c7bb8d36:.claude/rules/moai/workflow/factory-dispatch.md
 1 file changed, 1 insertion(+), 1 deletion(-)
```

기준과 HEAD 모두에서 -/+ 줄은 "[HARD] A lane whose card has passed verification does not wait for the leader to integrate it: the lane merges its own branch into the batch's release branch (`release/vX.Y.Z`) itself. …" 한 쌍이고, run의 F 변경 덩어리는 이 줄을 건드리지 않는다.

AC-MJ-014 Go 가드(자원 임대 `moai slot acquire --resource go-heavy --max-duration 15m` → `slot go-heavy acquired …`, 직전 `uptime` 1분 부하 19.28, 32 이하; 실행 후 `slot go-heavy released`):

```text
$ go test ./internal/template -count=1 -v -run '^(TestBindingLedgerIntegrity|…|TestCatalogHashCoversSkillSubfiles)$'   exit=0
--- PASS: TestDeployedAlwaysLoadedCharBudget (0.16s)
--- PASS: TestBindingLedgerIntegrity (8.97s)
--- PASS: TestTemplateNoInternalContentLeak (1.05s)
--- PASS: TestLaneRecheckDoctrine (0.00s)
--- PASS: TestLaneWatchdogRenewsFirst (0.00s)
--- PASS: TestLaneWatchdogRecordsRefusalAndMeasures (0.00s)
--- PASS: TestFactoryDispatchBodyCarriesRenewIntervalLinkage (0.00s)
--- PASS: TestRoleCoreStubGate (0.00s)
--- PASS: TestReviewOwnership_CardReviewDoctrine (0.00s)
--- PASS: TestWorkflowRulePathsPinned (0.00s)
--- PASS: TestDeclaredRuleMirrorForks (0.00s)
--- PASS: TestRuleTemplateMirrorDrift (0.00s)
--- PASS: TestCatalogHashCoversSkillSubfiles (0.02s)
--- PASS: TestCardIDBaselineHasNoStaleEntries (0.02s)
ok  	github.com/modu-ai/moai-adk/internal/template	10.708s
    binding_ledger_test.go:1084: [AC-ALB-021(2)] fragment sweep: binding/normative qualifying lines=560 … violations=0 lines exempt (already in anchor copy)=22

$ go test ./internal/cli -count=1 -v -run '^(TestAutoPickMirrorParity|TestAutoPickDocDoctrine|TestTodoSkillDocumentsClassification|TestAutoRankDoctrineAmendment|TestAutoRankMirrorParity|TestAutoRankMarkerDisclosure)$'   exit=0
--- PASS: TestAutoRankDoctrineAmendment (0.00s)
--- PASS: TestAutoRankMirrorParity (0.00s)
--- PASS: TestAutoRankMarkerDisclosure (0.00s)
--- PASS: TestAutoPickDocDoctrine (0.01s)
--- PASS: TestAutoPickMirrorParity (0.00s)
--- PASS: TestTodoSkillDocumentsClassification (0.00s)
ok  	github.com/modu-ai/moai-adk/internal/cli	1.249s

빈 스윕 확인: grep -c -E 'no tests to run|no test files' → go-template.log:0, go-cli.log:0
```

E-S1 (S1 재현):

```text
$ /usr/bin/grep -rn -E 'free lane|reaches `done`|previous card' internal/template/templates/.claude
…/factory-dispatch-detail.md:134:  - A free lane is a lane on which no card is actually executing; a lane that holds only merge-ready, held (operator-wait), or serial-waiting cards is a free lane.
…/factory-dispatch-detail.md:139:- **The transition boundary is between cards.** A factory lane leaves its card worktree (`ExitWorktree(keep)`) once its card reaches `done`, before the next card is routed to it, exactly as the stub's lane card transition rule requires.
$ /usr/bin/grep -n -i 'transition boundary\|reaches .done\|ExitWorktree(keep)' research.md design.md spec.md   → 출력 없음
```

E-S2:

```text
$ git show --stat ea7b93c085 → progress.md | 23, spec.md | 2, CHANGELOG.md | 1  (CHANGELOG.md 1줄 추가: SPEC-CARD-MODE-JUDGMENT-001 [Unreleased] 항목)
$ git diff --name-only 28c7bb8d36 0638768424 | /usr/bin/grep -c CHANGELOG → 0
$ git diff --name-only 28c7bb8d36 HEAD 중 템플릿·배포·SPEC·reports 밖 경로 → CHANGELOG.md, internal/template/catalog.yaml, internal/template/testdata/binding_ledger.json
```

E-S3:

```text
$ /usr/bin/grep -c 'Mode judgment' …/factory-dispatch-core.md …/factory-dispatch.md …/gtd.md
factory-dispatch-core.md:0  factory-dispatch.md:1  gtd.md:2
$ grep -o -E 'REQ-MJ-[0-9]+' spec.md | sort -u | wc -l → 17 ; grep -o -E 'AC-MJ-[0-9]+' acceptance.md | sort -u | wc -l → 20
```

동기화 산출물:

```text
$ git show --format='%H %s' ea7b93c085 → ea7b93c08553339262f4e99db615b0ab7192d061 docs(SPEC-CARD-MODE-JUDGMENT-001): sync-phase artifacts - 3-phase close (card t1894)
spec.md diff: -status: in-progress / +status: completed
f7cc3f8ee8 diff: -sync_commit_sha: pending-backfill … / +sync_commit_sha: ea7b93c08553339262f4e99db615b0ab7192d061
$ /usr/bin/grep -n -i 'binding run conditions' progress.md → exit 1 (없음)
```

보안·일관성 탐침:

```text
$ git diff --check 28c7bb8d36 HEAD → 출력 없음, exit 0
$ 추가 행 비밀정보 탐침(api key|secret|password|token=|ghp_|sk-…) → 3건, 모두 "ask-and-wait"의 "sk-a" 오탐
$ git diff --name-only 28c7bb8d36 HEAD | grep -c '\.go$' → 0
```

교차 모델 감사(`mcp__moai__audit_multi`, project_root=이 워크트리, target=baseBranch, delta_base=28c7bb8d36). 결과 다이제스트:

| 항목 | 값 |
|---|---|
| overall_verdict | fail |
| gate_unmet | 결과에 필드 없음; residual_risk_note "required-backend FAIL: codex" |
| plan_source | config |
| per_backend: codex | gate required, verdict fail |
| audit_receipt | rcpt-aeda7d3bfadbf2012f680998 |
| build_commit | b1806498fe |

codex 결과는 P2 두 건이며, 감사자가 둘 다 직접 재현했다(S1, S4). 감사 계획(`moai verify audit-plan --project-root <toplevel>`)은 `config_status ok`, model codex, gate codex required / claude off / glm off, `cross_model_active true`였다. 이 판정은 오케스트레이터가 결과 파일로 `moai verify audit-plan --result-file` 확인을 마칠 때까지 최종이 아니다.

### Baseline-attribution

- 트리: HEAD `f7cc3f8ee8c65061382a29c40cf7cc809c1cb7b2`, 작업 트리 깨끗, 기준 `28c7bb8d36`. 모든 측정은 이번 실행에서 이 트리를 대상으로 했다.
- Go 측정: `go version go1.26.8 darwin/arm64`가 이 트리의 소스에서 테스트 바이너리를 컴파일했다. 판정 빌드는 곧 이 트리다.
- 원장 셀: `/usr/bin/grep`(BSD), `awk`, `cmp`, `git`, 그리고 트리 안의 `closed-world-check.py`(`python3 -I`).
- 설치된 `moai`: v3.2.0-rc.38, 커밋 `b1806498fe`. `git merge-base --is-ancestor b1806498fe HEAD` → exit 1(엄밀한 조상이 아님, HEAD 쪽에 29커밋), 따라서 지연(lag) 진술이 필요 없다. 설치 빌드는 `moai slot`, `moai verify audit-plan`, MCP 감사 서버(같은 `b1806498fe`)에만 쓰였고, 교리 문서에 대한 측정값은 하나도 내지 않았다.

### Gaps

- AC-MJ-015(레인 수령 확인)는 리더 몫이라 관찰하지 않았다.
- Go는 이름이 지정된 14+6 테스트만 실행했다. 두 패키지의 나머지와 `go test ./...`는 실행하지 않았고, 저장소 전체 CI 판정은 관찰하지 않았다.
- 원장의 변이(mutant) 목록은 다시 실행하지 않았다. 지금 트리에서 재실행 가능한 알려진 실패 입력(E-010c, E-011f/g, E-012b, E-013b, E-019b, E-020c/e/f/g/h)만 다시 확인했다.
- codex의 델타 범위는 프롬프트로만 전달됐고 기계적으로 묶이지 않았다(residual_risk_note).
- 거부된 도구 호출 1건: 원장 재실행 스크립트를 셸 heredoc으로 쓰는 복합 명령이 워크트리 가드에 거부됐다. Write 도구로 스크립트를 쓴 뒤 단일 `python3 -I` 명령으로 다시 실행했고, 측정을 다른 것으로 대체하지 않았다.
- `go test -v` 원본 로그와 원장 실행 전문은 세션 스크래치에만 있고 여기 옮긴 줄만 인용 가능하다.

### Residual-risk

- 원장은 어휘 기반이라 문장 안의 의미 반전이나 다른 말로 쓴 모순을 보지 못한다(spec R-8). 이번 감사는 변경된 여덟 쌍과 바로 이웃한 규칙을 읽었지만, 다른 문서(예: 에이전트 파일, foreman 스킬)에 "done 뒤 라우팅"을 전제한 문장이 더 있을 수 있다.
- SessionStart 레인 공지(Go, 카드 t1695)는 여전히 범위 없는 크론 문구를 주입한다(R-1). 문서와 공지가 어긋난 채로 남는다.
- "운영자가 디스패치를 위임했다"를 성립시키는 장치는 정의하지 않았다(R-5, Q7 답변). 실무에서 위임을 가정하면 예외가 넓어질 수 있고, 어떤 명령도 이를 보지 못한다.
- S4가 운영자 판정 없이 남으면 슬롯 대기 레인이 5분 일회성 재무장을 폴링처럼 계속할 수 있다.

## 6. 반복 이력

- 이번이 동기화 단계 첫 감사다. 계획 단계 감사(iter0~iter5)는 `.moai/reports/t1894/plan-audit*.md`에 있고, 마지막 계획 감사는 PASS 0.91(`audited_sha` 21974dddf2)이다. 그 뒤 중간 개정(`8b9f610b91`)이 계획 산출물을 바꿨으므로, 이 감사는 그 해시가 아니라 위 증거로 판단했다.
- 재감사 범위 제안: S1 정정 커밋(매니페스트 개정 + D/F 문장 + 해당 E-셀·E-020d) 이후 델타 재감사. S2·S3·S4·S5는 비차단이며 처분 기록 여부만 확인하면 된다.
