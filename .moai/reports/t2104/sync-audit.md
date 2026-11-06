auditor-model: claude-opus-5-5

verdict: FAIL
audited_sha: 8792a156d4e9db5c850bda25c86760cf88f6c58a

# t2104 (GitHub #1938) 동기화 감사, 2회차

- 점수: 0.85 (Functionality 80 FAIL / Security 92 PASS / Craft 84 PASS / Consistency 90 PASS, 가중 32.0 + 23.0 + 16.8 + 13.5 = 85.3)
- 감사 영수증: 없음. `mcp__moai__codex_audit`를 허용된 1회 호출했으나 서버가 실행 전에 거절했다(`stale_server: true`, 아래 E9).
- 판정 근거: (1) 이 트리는 `workflow.audit.gates.codex: required`인데 codex 판정이 없다. 필수 게이트 미충족이라 PASS를 낼 수 없다. (2) 차단 결함 N1과 N2가 남아 있다. 1회차의 F1과 F2는 해소됐다.
- 범위: `git diff 511d8cd964..8792a156d4` 중 운영 코드 `internal/homestate/factory_run_rebind.go`(+10/-1), 시험 `internal/cli/factory_runs_rebind_expired_t2104_test.go`, 문서 `.moai/docs/factory-expired-lease-dead-run.md`, 진행 기록 `.moai/reports/t2104/progress.md`.

## Claim

1. F1 해소: 죽은 실행 두 개가 만료 시각이 다른 임대로 같은 카드를 쥔 경우를 문서 표에 넣었고, 시험 `TestFactoryRunsRebindSecondDeadRunWithDivergingExpiredLease`로 고정했다. 이 시험은 변경 전 코드에서 실패한다(E3 m0).
2. F2 해소: 두 줄 상태의 피해를 측정했다. `status`는 카드를 두 번 보여 주고, 완료 게이트의 행 해석은 살아 있는 행을 고르며, 실행 없는 해석은 `AMBIGUOUS_FACTORY`가 된다. 이번 감사에서 변경 전 코드로 같은 결과를 다시 관측했다(E4).
3. 운영 코드 변경은 유효한 임대를 가진 행을 버리지 않는다. 임대 가드는 카드 반복문 전체를 먼저 돌고, 그 다음에 중복 비교를 한다. 같은 `now`를 쓰고 트랜잭션은 IMMEDIATE다. 유효 임대(UTC, +09:00 오프셋), 만료 시각 없이 보유자만 있는 임대는 모두 거부되고 DB는 그대로다(E5).
4. 해석할 수 없는 만료 시각은 만료로 취급된다. 이것은 `holdsValidLease`에 문서화된 기존 계약이고 쓰기 경로는 항상 RFC3339Nano로 기록한다. 다만 변경 전에는 이 행이 거부됐고 지금은 포인터와 함께 버려진다(E5, N3).
5. `cardPairExcludedColumns`는 바뀌지 않았다. `divergentColumns`의 호출자는 `RebindRun` 하나이고, 임대 두 열은 반환된 슬라이스에서만 빠진다. T1973 회귀 시험 전부와 Rebind 시험 전부가 통과한다(E1, E2).
6. 새 결함 N1: `RebindRun`은 자동 조정 경로(`ResolveActiveRun` → `ReconcileActiveRuns` → `consolidateIntoLive`)도 지난다. 그래서 이 변경으로 실행 없는 동사 하나만 호출해도 운영자 개입 없이 만료 임대 중복 행이 버려지고 죽은 실행이 은퇴한다. 문서와 진행 기록은 이 경로를 말하지 않는다(E4, E7).
7. 새 결함 N2: 임대 가드를 지운 변이에서 새 시험 네 개가 모두 통과하고, 유효 임대 원본 행이 중복 대상 앞에서 조용히 버려진다. 가드 제거는 기존 `TestRed_T1973b_P1`만 잡는다. 그 시험은 대상 행이 없는 모양이다(E3 m1).

## Evidence

### E1 — 지시된 CLI 시험(감사 트리)
`go test ./internal/cli -run TestFactoryRuns -count=1` (cwd = t2104 트리)
```
ok  	github.com/modu-ai/moai-adk/internal/cli	162.886s
exit=0
```

### E2 — 지시된 homestate 시험(감사 트리)
처음에는 `go test ./internal/homestate -run 'Rebind|T1973' -count=1`로 한 번에 돌렸다. 기본 10분 시간 제한에 걸려 실패했다(`panic: test timed out after 10m0s`, 당시 실행 중 `TestRebindRefusesSourceWhoseOwnerIsNotDead` 18s). 시스템 부하 평균은 약 150이었다. 다음 재실행은 디스크가 가득 차서 실패했다(`disk I/O error (4874)`, `no space left on device`). 디스크 여유가 생긴 뒤 둘로 나누고 시간 제한을 늘려 다시 돌렸다. 공유 빌드 캐시가 실행 중에 지워지고 있어서 전용 `GOCACHE`를 썼다.
```
$ go -C <t2104> test ./internal/homestate -run Rebind -count=1 -timeout 60m
ok  	github.com/modu-ai/moai-adk/internal/homestate	133.741s
$ go -C <t2104> test ./internal/homestate -run T1973 -count=1 -timeout 60m
ok  	github.com/modu-ai/moai-adk/internal/homestate	212.526s
```
`go vet ./internal/homestate`는 출력 없이 끝났다. `gofmt -l` 두 변경 파일은 출력이 없었다.

### E3 — 변이 검사(감사 커밋의 `git archive` 사본, 트리는 수정하지 않음)
| 변이 | 바꾼 곳 | 결과 |
|---|---|---|
| m0 변경 전 동작 | `diverged := divergentColumns(c, tgtCard)` (임대 열 제외 제거) | 죽음: `AbandonsExpiredRowDuplicatingLiveRow` FAIL(`diverging on lease_holder, lease_expires_at`), `SecondDeadRunWithDivergingExpiredLease` FAIL(`rebind of run-b: ... diverging on lease_expires_at`), 나머지 2개 PASS |
| m1 임대 가드 제거 | `if holdsValidLease(c, now) {...}` 블록 삭제 | 새 시험 4개 모두 PASS(살아남음). 기존 `TestRed_T1973b_P1_SourceCardHoldingValidLeaseIsRekeyed` 하위 2개는 FAIL(`operation succeeded, want a refusal naming [c1 run-old]`) |
| m2 비교 과잉 완화 | 제외 술어에 `stage` 추가 | 죽음: `TestFactoryRunsRebindStillRefusesNonLeaseDivergence` FAIL, 나머지 3개 PASS |

m1에서 감사자 탐침을 돌리면 유효 임대 원본이 중복 대상 앞에서 버려진다.
```
holder="lane-9" expiry="2026-11-06T08:46:47.023691Z" err=<nil> out=rebound run-src into run-tgt (owner dead) / moved cards: none / abandoned cards: t1 /
rows=run-src:abandoned ; run-tgt:assigned src=retired unchanged=false
holder="lane-9" expiry="" err=<nil> out=... abandoned cards: t1 /
```

### E4 — 두 줄 상태와 자동 조정(F2 재관측 + N1)
탐침 고정물: `run-src` 소유자 죽음, `t1` leased, 임대 `lane-9`, 만료 `2026-09-20T01:00:00Z`. `run-tgt` 살아 있음, `t1` assigned. `card_dispatch` t1 → run-tgt. 순서는 `status` → `RecordedCardRowReadonly` → `factorymsg.ResolveActiveRun(ctx, root, "")` → `runs --retire run-src --rebind-to run-tgt` → 다시 셋.

변경 전 코드(m0):
```
STATUS-before | t1 run=run-src state=leased ... lease=lane-9 expired 2026-09-20T01:00:00Z ...
STATUS-before | t1 run=run-tgt state=assigned ... lease=none ...
GATE-before run=run-tgt state=assigned linked=true err=<nil>
RESOLVE-before err=AMBIGUOUS_FACTORY: run-src (owner dead, ...), run-tgt (owner live, ...)
ROWS-before run-src:leased ; run-tgt:assigned
REBIND err=factory run consolidation refused: card t1: run run-tgt already holds a assigned row diverging on lease_holder, lease_expires_at ...
ROWS-after run-src:leased ; run-tgt:assigned ; src=active
```
감사 커밋(HEAD):
```
STATUS-before | t1 run=run-src state=leased ... lease=lane-9 expired 2026-09-20T01:00:00Z ...
STATUS-before | t1 run=run-tgt state=assigned ... lease=none ...
GATE-before run=run-tgt state=assigned linked=true err=<nil>
RESOLVE-before err=<nil>
ROWS-before run-src:abandoned ; run-tgt:assigned
REBIND err=factory run consolidation refused: source run run-src changed owner or status since classification (status retired)
STATUS-after | t1 run=run-src state=abandoned stage=- version=4 owner=lane-9 lease=lane-9 until 2026-09-20T01:00:00Z ...
GATE-after run=run-tgt state=assigned linked=true err=<nil>
RESOLVE-after run=run-tgt err=<nil>
ROWS-after run-src:abandoned ; run-tgt:assigned ; src=retired
```
HEAD에서는 운영자 동사가 실행되기 전, `ResolveActiveRun`의 조정 단계가 이미 `t1`을 버리고 `run-src`를 은퇴시켰다. 그 뒤의 명시적 `--rebind-to`는 "status retired"로 거절된다.

### E5 — 임대 분류 경계(같은 탐침, 중복 대상 행 있음)
| 원본 임대 | 변경 전(m0) | HEAD | 차이 의도 여부 |
|---|---|---|---|
| 유효, UTC `...Z` | 거부(live lease), DB 그대로 | 거부(live lease), DB 그대로 | 같음 |
| 유효, `+09:00` | 거부, DB 그대로 | 거부, DB 그대로 | 같음 |
| 보유자만, 만료 없음 | 거부(live lease, expires ""), DB 그대로 | 같음 | 같음 |
| 만료, `+09:00` | 거부(lease 열 분기) | 버림 + 은퇴 | 의도(카드 목표) |
| 해석 불가 `2099-01-01 00:00:00` | 거부(lease 열 분기) | 버림 + 은퇴 | 의도로 기록되지 않음(N3) |
| 만료, 실행 없는 동사의 자동 조정 | `AMBIGUOUS_FACTORY`, 쓰기 없음 | 버림 + 은퇴 | 의도로 기록되지 않음(N1) |
| 만료, 두 번째 죽은 실행 | 거부(`lease_expires_at`) | 버림 + 은퇴 | 의도(F1) |
| 만료 + `stage` 분기 | 거부 | 거부 | 같음 |

### E6 — 코드 읽기
- 순서: `factory_run_rebind.go:236-248`이 모든 원본 카드에 `holdsValidLease`와 `refuseThirdRunHolder`를 먼저 적용한다. 버림 분기(`:258-294`)는 그 다음 반복문에 있다. 둘 다 `now := time.Now().UTC()`(`:216`) 하나를 쓴다. 트랜잭션은 IMMEDIATE로 열린다(`leader_approval.go`의 ApprovalGate 주석, "the DSN opens transactions IMMEDIATE").
- `holdsValidLease`(`factory_run_lease_rebound.go:200-208`)는 임대 보유 상태가 아니면 false, 두 임대 열이 비면 false, 그 밖에는 `!LeaseExpired(now)`다. `LeaseExpired`(`card_record.go:147-157`)는 만료 시각이 비면 false(곧 유효), 해석에 실패하면 true(곧 만료)다. 비교는 `now.Before(at)`로 순간끼리 하므로 시간대와 무관하다.
- `abandonWithPointer`(`:531-550`)는 `version=?` 조건으로 상태만 `abandoned`로 바꾼다. 임대 열은 버려진 행에 남는다. 이벤트 페이로드에는 임대 열이 없다(진행 기록의 Residual-risk와 일치).
- 같은 판단의 선례가 있다. `deadRunResidue`(`factory_run_lease_rebound.go:169-194`)는 임대 재획득 경로에서 만료 임대 잔여 행을 비교 없이 버린다.

### E7 — 자동 경로의 호출자
`consolidateIntoLive`(`factory_run_zombie.go:40-48`)가 `RebindRun`을 호출한다. `ReconcileActiveRuns`(`factory_run_retire.go:368-422`)는 살아 있는 소유자가 정확히 하나이고 판정 불가 소유자가 없을 때 그 함수를 부른다. 실행 없는 `ResolveActiveRun`의 운영 호출자는 다음과 같다: `internal/cli/factory_card.go:3956`, `factory.go:522`, `integration_candidate.go:694`, `factory_handoff_recover.go:251`, `todo_auto_lane.go:61,63`.

### E8 — 문서 대조
문서 표의 네 행은 E3와 E5의 관측과 맞다. "With the change above one `--rebind-to` call removes it"는 사실이지만 완전하지 않다(N1). 인계 절의 "state that predates this change ... decide twice, then retire"는 살아 있는 실행이 하나뿐일 때 실행 없는 동사의 자동 조정이 먼저 해소해 버리므로, 그 경우에는 낡은 안내다.

### E9 — 교차 모델 감사
```
mcp__moai__codex_audit(project_root=<t2104>, target=baseBranch, delta_base=511d8cd9643ad175911538d048237d9c827dc2b8, mode=native)
→ error: codex_audit: the serving moai build is stale — stale_server: the serving moai build (/Users/goos/go/bin/moai) differs from the installed build (/Users/goos/go/bin/moai) — reconnect the moai MCP server (restart the host session) so the installed build serves
```
`workflow.yaml`의 게이트는 `codex: required`, `claude: off`, `glm: off`이다.

## Findings (구조화된 결함 목록)

- **N1 [Medium] [blocking] [in-diff, 재현됨] `internal/homestate/factory_run_rebind.go:278-287` + `.moai/docs/factory-expired-lease-dead-run.md`** — 완화가 명시적 운영자 동사뿐 아니라 자동 조정 경로에도 적용된다. 그 경로는 실행 없는 동사 6곳(E7)에서 운영자 결정 없이 실행된다. E4에서 변경 전 `AMBIGUOUS_FACTORY`(쓰기 없음)였던 상태가 HEAD에서는 첫 실행 없는 해석에서 버림 + 은퇴로 바뀐다. 결과 자체는 #1948 §8의 설계 방향과 맞을 수 있다. 그러나 이 변화는 문서, 진행 기록, 커밋 메시지, 시험 어디에도 없다. 판정 함수 카드에 요구되는 허용/거부 비교표에서 표시되지 않은 차이 행이다.
  **Required fix:** 둘 중 하나를 고르고 기록한다. (a) 자동 경로 포함이 의도라면 문서에 "실행 없는 동사(`ResolveActiveRun`)의 조정도 같은 버림을 수행한다"를 쓰고 인계 절을 그에 맞게 고친다. `ReconcileActiveRuns`(또는 `ResolveActiveRun`)로 만료 임대 중복 행이 버려지고 죽은 실행이 은퇴하는 시험 하나를 추가한다. (b) 의도가 아니라면 완화를 명시적 동사로 한정한다(예: `RebindRequest` 필드로 opt-in, `consolidateIntoLive`는 끔). 그리고 자동 경로가 여전히 거부하는 시험을 둔다.
- **N2 [Medium] [blocking] [in-diff, 재현됨] `internal/cli/factory_runs_rebind_expired_t2104_test.go`** — 변경은 두 겹이던 방어(임대 가드 + 임대 열 분기)를 한 겹(가드)으로 줄였다. 남은 가드와 버림 분기의 순서 불변식("The lease guard above already established…")을 묶는 시험이 없다. 가드를 지운 m1에서 새 시험 4개가 모두 통과하고, 유효 임대 원본이 중복 대상 앞에서 버려진다(E3). 기존 T1973b P1 시험은 가드 전체 삭제는 잡지만 대상 행이 없는 모양이다. 그래서 대상 행이 있을 때만 가드를 건너뛰는 회귀, 또는 버림 분기를 가드 앞으로 옮기는 회귀는 잡지 못한다.
  **Required fix:** 유효(미래) 임대 원본 + 중복 `assigned` 대상 → 거부, `lcDump` 그대로인 시험 하나를 추가한다. m1 변이에서 그 시험이 실패하는 것을 진행 기록에 원문으로 남긴다.
- **N3 [Low] [optional] [in-diff] 같은 위치** — 해석할 수 없는 만료 시각을 가진 임대 보유 행이 이제 중복 대상 앞에서 버려진다. 변경 전에는 거부됐다(E5). 기존 계약(`holdsValidLease` 주석, `deadRunResidue` 선례)과 일관되고 쓰기 경로는 항상 RFC3339Nano이므로 손상된 데이터에서만 나타난다. 차단할 일은 아니다. 문서의 "Why the lease alone does not refuse" 절에 한 줄로 밝히면 충분하다.
- **N4 [Info] [optional] [in-diff] 문서 표** — 버림 뒤에도 `status`는 버려진 행을 `state=abandoned ... lease=lane-9 until ...`로 계속 보여 준다(E4). 문서의 "one live row remains"는 정확하다. 다만 운영자가 `status`에서 두 줄을 보고 실패로 오해할 수 있으므로 표에 한 단어("살아 있는 행 하나와 abandoned 행 하나") 정도를 덧붙이면 좋다.
- **G1 [gate] [blocking]** — 필수 codex 게이트가 충족되지 않았다(E9). 코드 결함이 아니다. MCP 서버를 다시 연결한 뒤 codex 감사를 받아야 PASS 계열 판정이 가능하다. 남은 codex 호출 한도는 이 카드 기준 1회다(규칙 d-20261106T064700Z-de2d, 카드당 최대 2회, 이번에 1회 사용).

1회차 결함의 처리: F1 해소(시험 + 문서 행, m0에서 실패 확인). F2 해소(진행 기록의 측정 + E4 재관측). F3 해소(문서가 `resolveFactoryCardRun`과 `ReassignCard`를 함께 명시). F4 해소(진행 기록에 원문 출력이 있음, 다만 그 탐침 파일은 지워져 있어서 감사자는 E4로 따로 재관측했다). F5는 그대로 남아 있다(선택 사항, 새 시험은 `card.rebound` 이벤트 1건만 확인).

## Baseline-attribution

- 트리: 워크트리 t2104, `git rev-parse HEAD` = `8792a156d4e9db5c850bda25c86760cf88f6c58a`, 감사 시작과 종료 시점 `git status --short` 출력 없음. 브랜치 `WT-run-leased-run`.
- E1과 E2는 이 트리에서 직접 돌렸다. E3, E4, E5는 `git archive 8792a156d4` 사본에서 돌렸다. m0, m1, m2는 각각 한 군데만 고쳤다. 탐침 파일은 사본의 `internal/cli/probe_t2104r2_test.go`에만 있고 트리에는 넣지 않았다.
- Go 툴체인은 세션 기본값(go1.26.8 darwin-arm64, 패닉 스택에 찍힌 경로 기준)이다. moai 바이너리로 측정한 값이 없으므로 §2.2 도구 출처 규정은 적용되지 않는다.

## Gaps

- **교차 모델 감사 없음.** codex 호출 1회가 `stale_server`로 실행 전에 거절됐다. 재시도하지 않았다(서버 재연결은 호스트 세션 재시작이 필요하고, 한도도 아낀다). 영수증이 없다. `audit_multi`, `glm_audit`, `claude_audit`은 지시대로 쓰지 않았다. `moai verify audit-plan`은 돌리지 않았다. 게이트 설정은 `workflow.yaml`을 직접 읽어 확인했다.
- 지시된 단일 명령 `go test ./internal/homestate -run 'Rebind|T1973' -count=1`은 통과 결과를 얻지 못했다. 처음에는 10분 제한, 다음에는 디스크 부족 때문이었다. 같은 시험 집합을 둘로 나누고 `-timeout 60m`과 전용 `GOCACHE`로 돌려서 통과를 관측했다(E2).
- 환경 장애: 디스크가 100% 찼다(`df`: 여유 255Mi). 공유 Go 빌드 캐시의 항목이 실행 중에 사라졌다(`could not import strings (open .../go-build/...-d: no such file or directory)`). 변이 실행 3회가 이 때문에 빌드 실패했고, 전용 `GOCACHE`로 다시 돌렸다. 디스크 확보를 위해 내 스크래치 tar 파일(286M)을 지웠다. 스크래치 사본 디렉터리 삭제(`rm -rf`)는 권한 거부로 남아 있다.
- 워크트리 가드 거부: 이 세션은 t2113에 격리되어 있어서 `cd <t2104> && ...` 복합 명령, `$S` 변수가 든 `tar`, `cmd` 경로를 넣은 `tar`, `git -C` 포함 복합 명령이 거부됐다. 모두 단순한 단독 명령(`go -C`, `git -C`, 절대 경로)으로 다시 실행했다. EnterWorktree 도구는 이 에이전트에 없었다.
- 전체 `internal/cli`나 `internal/homestate` 스위트, Windows, CI, 커버리지 측정은 하지 않았다. 변경 범위 밖의 다른 경로 영향은 E7의 호출자 목록까지만 읽었다.
- `factory complete`를 처음부터 끝까지 실행하지는 않았다. 완료 게이트의 행 해석은 `RecordedCardRowReadonly`로만 확인했다(진행 기록과 같은 한계).

## Residual-risk

- N1의 자동 버림은 `refuseLiveWorker`(원본 실행에 등록된 살아 있는 작업자)와 임대 가드가 막는다. 그러나 다른 실행에 등록됐거나 등록이 없는 상태에서 임대만 만료된 채 살아 있는 레인이 있다면, 그 레인의 원본 행은 실행 없는 동사 하나로 버려진다. 진행 필드(`stage`, `worktree_path`, `spec_id`, `evidence_sha` 등)가 비어 있거나 대상과 같을 때만 그렇다. 잃는 것은 만료된 임대 기록뿐이고 그 값은 버려진 행에 그대로 남는다.
- N2가 고쳐지기 전까지는, 가드를 대상 행이 없는 경우에만 적용하도록 바꾸는 미래의 리팩터링을 어떤 시험도 잡지 못한다.
- codex 판정이 없으므로 이 감사는 단일 모델(Claude) 관점이다.

## Iteration history

| 회차 | 감사 커밋 | 판정 | 점수 | 차단 결함 |
|---|---|---|---|---|
| 1 | 511d8cd964 | FAIL | 0.78 | F1(여러 죽은 실행 미문서·미시험), F2(두 줄 상태 피해 미측정) |
| 2 (이 파일) | 8792a156d4 | FAIL | 0.85 | N1(자동 조정 경로 미공개), N2(가드-순서 불변식 미고정), G1(필수 codex 게이트 미충족) |

수리로 생긴 결함 표시: N1과 N2는 모두 511d8cd964 이후 수리 diff 안에 있으므로 repair-born이다. 1회차의 차단 결함은 수리 이전 범위였으므로, 연속 두 번 repair-born 규칙은 아직 발동하지 않는다.
