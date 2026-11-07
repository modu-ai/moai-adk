auditor-model: claude-opus-5-5

verdict: PASS
score: 0.91
audited_sha: d325cf75d7b871bdf58c21745ce12f43830944e2
audit_receipt: rcpt-121841c4ceba8e412225bddd

# t2104 (GitHub #1938) 동기화 감사, 4회차(게이트 마감 재감사)

- 점수: 0.91 (Functionality 92 PASS / Security 92 PASS / Craft 88 PASS / Consistency 90 PASS, 가중 36.8 + 23.0 + 17.6 + 13.5 = 90.9). 3회차 점수를 그대로 쓴다. 이번 회차에 코드는 바뀌지 않았고(HEAD 동일), codex 검토도 결함을 내지 않았다.
- 감사 영수증: `rcpt-121841c4ceba8e412225bddd`(E7). 운영자가 rc.2 설치 뒤 `/mcp`를 다시 연결했고, 이번 `mcp__moai__codex_audit` 1회 호출은 서버가 검토까지 실행해 `verdict: pass`, 결함 0건을 돌려줬다.
- 판정 근거: 3회차의 유일한 FAIL 사유였던 G1(필수 codex 게이트 미충족)이 해소됐다. 코드 차단 결함은 3회차 측정(E1~E5) 그대로 없다. 남은 O1·O2·F5는 모두 optional이라 판정을 바꾸지 않는다.
- 3회차 기록: 아래 E1~E6과 Findings는 3회차 측정 원문이다. 4회차에서 새로 관측한 것은 E7뿐이다.
- 범위: `git diff 511d8cd964..d325cf75d7`. 운영 코드 `internal/homestate/factory_run_rebind.go`, `internal/cli/factory_handoff_recover.go`(1줄). 시험 `internal/cli/factory_runs_rebind_expired_t2104_test.go`. 문서 `.moai/docs/factory-expired-lease-dead-run.md`. 진행 기록 `.moai/reports/t2104/progress.md`. 이번 회차의 실질 수리분은 `8792a156d4..d325cf75d7`이다.

## Claim

1. N1 해소: 임대 열 무시는 이제 `RebindRequest.AbandonExpiredLeaseDuplicate`가 참일 때만 일어난다. 운영 코드에서 이 필드를 참으로 두는 곳은 `runRebindRetire`(`runs --retire <dead> --rebind-to <live>`) 하나뿐이다. `consolidateIntoLive`, 합류 선택기(`mergeChosenRun`), 리더 인수(`factory_run_adopt.go`)는 필드를 쓰지 않으므로 거짓이다(E2).
2. 필드가 거짓이면 비교는 변경 전(511d8cd964)과 같다. 기준 코드는 `divergentColumns(c, tgtCard)`를 걸러내지 않고 그대로 썼고, 지금 거짓 분기도 같은 슬라이스를 그대로 쓴다(E2).
3. 자동 경로 시험 `TestFactoryAutomaticReconcileKeepsRefusingExpiredLeaseDuplicate`는 실효가 있다. 완화를 늘 켜는 변이(m1)에서 이 시험만 실패한다(E4).
4. 명시적 동사의 opt-in도 시험이 묶고 있다. 동사가 거짓을 넘기는 변이(m2)에서 기존 버림 시험 두 개가 실패한다(E4).
5. N2 해소: `TestFactoryRunsRebindValidLeaseSourceWithDuplicateTargetRefuses`가 유효 임대 원본 + 중복 `assigned` 대상 → 거부(`live lease`) + `lcDump` 불변을 확인한다. 레인이 가드를 지운 overlay 변이에서 이 시험이 실패한 원문을 진행 기록에 남겼다. 이번 감사에서는 그 변이를 다시 돌리지 않았다(Gaps).
6. 문서는 N1(b) 선택, 자동 경로 표 행, 유효 임대 행, N3(해석 불가 만료 시각) 한 줄, N4(`status`에 abandoned 행이 남음) 한 줄을 모두 반영했다(E5).
7. 지시된 시험 두 묶음은 이 트리에서 통과했다. `gofmt -l`과 `go vet`도 깨끗하다(E1, E3).

## Evidence

### E1 — 지시된 시험(감사 트리, HEAD d325cf75d7)
```
$ go test ./internal/cli -run TestFactoryRuns -count=1 -timeout 60m
ok  	github.com/modu-ai/moai-adk/internal/cli	25.990s
exit=0

$ go test ./internal/homestate -run 'Rebind|T1973' -count=1 -timeout 30m
ok  	github.com/modu-ai/moai-adk/internal/homestate	182.883s
exit=0
```
이번에는 지시된 단일 명령이 한 번에 통과했다(2회차의 시간 초과와 디스크 부족은 재발하지 않았다).

### E2 — 호출자와 기준 비교
`grep -rn --include='*.go' -e 'RebindRun(' -e 'RebindRequest{' -e 'AbandonExpiredLeaseDuplicate' internal cmd pkg` 중 운영 코드 줄:
```
internal/homestate/factory_run_zombie.go:41:	if _, err := f.RebindRun(ctx, RebindRequest{Source: src, Target: target, RequireStampBasis: true}, opts); err != nil {
internal/cli/factory_run_select.go:421:	res, err = db.RebindRun(context.Background(), homestate.RebindRequest{Source: *src, Target: *tgt}, opts)
internal/cli/factory_run_adopt.go:191:		res, err := db.RebindRun(ctx, homestate.RebindRequest{Source: src, Target: target, RequireStampBasis: true}, opts)
internal/cli/factory_handoff_recover.go:214:	res, err := db.RebindRun(cmd.Context(), homestate.RebindRequest{Source: *src, Target: *tgt, AbandonExpiredLeaseDuplicate: true}, opts)
```
나머지 일치 줄은 모두 `_test.go`이며 어느 것도 이 필드를 설정하지 않는다. `runRebindRetire`의 유일한 호출자는 `factory_handoff_recover.go:150`의 `if rebindTo != ""` 분기다.

기준 커밋 비교: `git show 511d8cd964:internal/homestate/factory_run_rebind.go`의 해당 줄
```
278:			if diverged := divergentColumns(c, tgtCard); len(diverged) > 0 {
```
HEAD의 거짓 분기는 `diverged := divergentColumns(c, tgtCard)` 뒤에 `if req.AbandonExpiredLeaseDuplicate {...}`만 붙인다. 따라서 거짓일 때 거부 조건은 기준과 같다.

### E3 — 형식과 정적 검사
```
$ gofmt -l internal/homestate/factory_run_rebind.go internal/cli/factory_handoff_recover.go internal/cli/factory_runs_rebind_expired_t2104_test.go
gofmt_exit=0   (출력 없음)
$ go vet ./internal/homestate ./internal/cli
vet_exit=0     (출력 없음)
```

### E4 — 변이 검사(`go test -overlay`, 트리는 수정하지 않음)
각 변이는 한 줄이다(`diff`로 확인).
```
m1  factory_run_rebind.go:293   if req.AbandonExpiredLeaseDuplicate {  ->  if true || req.AbandonExpiredLeaseDuplicate {
m2  factory_handoff_recover.go:214   AbandonExpiredLeaseDuplicate: true}  ->  AbandonExpiredLeaseDuplicate: false}
```
명령: `go test -overlay <mN.json> ./internal/cli -run 'TestFactoryAutomaticReconcile|TestFactoryRunsRebind' -count=1 -v -timeout 30m`

m1(완화 기본 켜짐) — 죽음:
```
    factory_runs_rebind_expired_t2104_test.go:34: automatic reconcile changed state:
    factory_runs_rebind_expired_t2104_test.go:37: source row state = "abandoned", want leased (not silently abandoned)
    factory_runs_rebind_expired_t2104_test.go:40: source status = "retired", want active (not silently retired)
--- FAIL: TestFactoryAutomaticReconcileKeepsRefusingExpiredLeaseDuplicate (7.20s)
--- PASS: TestFactoryRunsRebindValidLeaseSourceWithDuplicateTargetRefuses (3.86s)
--- PASS: TestFactoryRunsRebindMovesExpiredLeasedRow (2.88s)
--- PASS: TestFactoryRunsRebindAbandonsExpiredRowDuplicatingLiveRow (2.93s)
--- PASS: TestFactoryRunsRebindSecondDeadRunWithDivergingExpiredLease (2.99s)
--- PASS: TestFactoryRunsRebindStillRefusesNonLeaseDivergence (3.55s)
FAIL	github.com/modu-ai/moai-adk/internal/cli	28.462s
exit=1
```
m2(명시적 동사 opt-in 끔) — 죽음:
```
--- PASS: TestFactoryAutomaticReconcileKeepsRefusingExpiredLeaseDuplicate (5.32s)
--- PASS: TestFactoryRunsRebindValidLeaseSourceWithDuplicateTargetRefuses (3.36s)
--- PASS: TestFactoryRunsRebindMovesExpiredLeasedRow (2.83s)
    factory_runs_rebind_expired_t2104_test.go:117: an expired lease alone must not block the consolidation: factory run consolidation refused: card t1: run run-tgt already holds a assigned row diverging on lease_holder, lease_expires_at (target evidence "", source ""); the live row of run run-src stays and the conflict needs a fresh decision
--- FAIL: TestFactoryRunsRebindAbandonsExpiredRowDuplicatingLiveRow (2.63s)
    factory_runs_rebind_expired_t2104_test.go:151: rebind of run-b: factory run consolidation refused: card t1: run run-tgt already holds a leased row diverging on lease_expires_at (target evidence "", source ""); the live row of run run-b stays and the conflict needs a fresh decision
--- FAIL: TestFactoryRunsRebindSecondDeadRunWithDivergingExpiredLease (3.44s)
--- PASS: TestFactoryRunsRebindStillRefusesNonLeaseDivergence (2.20s)
FAIL	github.com/modu-ai/moai-adk/internal/cli	22.697s
exit=1
```

### E5 — 문서와 진행 기록 대조
- 문서 신설 문단: "Only the explicit verb relaxes the lease comparison ... The automatic paths (`ResolveActiveRun` -> `ReconcileActiveRuns` -> `consolidateIntoLive`, the join chooser, leader adoption) leave it false and behave exactly as before this card". E2와 일치한다.
- 표 행 추가: 자동 경로(변화 없음), 유효 임대(거부). E1·E4와 일치한다.
- N3 문장("An expiry that does not parse counts as expired ... only with corrupted data"), N4 문장("`status` still lists the abandoned source row") 반영.
- 인계 절: "the automatic path does not (it stays ambiguous until the operator runs that verb)". 2회차 E8에서 지적한 낡은 안내가 고쳐졌다.
- 진행 기록 "Repair round 3": N1 결정 기록((b) 선택, (a) 기각 이유), RED(8792a156d4 코드에서 자동 경로 시험 실패 원문), GREEN, N2 가드 제거 변이 실패 원문이 있다.

### E6 — 교차 모델 감사(1회)
```
mcp__moai__codex_audit(project_root=/Users/goos/.moai/worktrees/projects/moai-adk-go/t2104, target=baseBranch, delta_base=511d8cd9643ad175911538d048237d9c827dc2b8, mode=native)
-> {"_server":{"commit":"b00c35d103","version":"v3.2.1-rc.1"},"error":"codex_audit: the serving moai build is stale — stale_server: the serving moai build (/Users/goos/go/bin/moai) differs from the installed build (/Users/goos/go/bin/moai) — reconnect the moai MCP server (restart the host session) so the installed build serves","stale_server":true}
```
게이트 설정(`.moai/config/sections/workflow.yaml:36-39`):
```
        gates:
            claude: off
            codex: required
            glm: off
```
운영자가 서버를 다시 연결했다고 전달받았지만, 이 감사 세션이 보는 서버(pid 23018, commit b00c35d103)는 여전히 설치된 빌드와 다르다고 스스로 보고한다.

### E7 — 교차 모델 감사(4회차, 1회, 게이트 마감)
서버 자기 보고(MCP 지시문): `moai self-hosted MCP server, build v3.2.1-rc.2 (commit: 04d8c27a7e, built: 2026-11-06T12:10:41Z) (pid 11855)`.
감사 전 확인:
```
$ git rev-parse HEAD
d325cf75d7b871bdf58c21745ce12f43830944e2
$ git rev-parse 511d8cd964 && git merge-base --is-ancestor 511d8cd964 HEAD && echo ancestor-ok
511d8cd9643ad175911538d048237d9c827dc2b8
ancestor-ok
```
호출과 결과(원문):
```
mcp__moai__codex_audit(project_root=/Users/goos/.moai/worktrees/projects/moai-adk-go/t2104, target=baseBranch, mode=native, delta_base=511d8cd9643ad175911538d048237d9c827dc2b8)
-> {"verdict":"pass","summary":"Verdict: pass\n\n구스 행님, 지정한 기준 커밋 대비 변경에서 조치할 코드 결함은 찾지 못했어. 정적 검토만 했고, 지침대로 테스트·빌드·재현 명령은 실행하지 않았어.","findings":[],"next_steps":[],
    "provenance":{"backend":"codex","transport":"codex app-server","source":"moai mcp-server","resolved_model":"gpt-6.1-sol","requested_effort":"high","tool_surface":"codex_audit","session_persisted":false,"error_code":""},
    "audit_receipt":"rcpt-121841c4ceba8e412225bddd","build_commit":"04d8c27a7e","review_base":"delta_base (merge base 511d8cd9643ad175911538d048237d9c827dc2b8)"}
```
결함 0건이므로 재현할 항목이 없다. 시험 스위트는 다시 돌리지 않았다(지시: codex 결함이 있을 때만).

## Findings (구조화된 결함 목록)

- **G1 [gate] [blocking] — 4회차에 해소(E7, 영수증 `rcpt-121841c4ceba8e412225bddd`).** 3회차 기록: 필수 codex 게이트가 충족되지 않았다(E6). 코드 결함이 아니다.
  **Required fix:** 이 감사를 띄우는 호스트 세션 자체를 재시작해 MCP 서버를 현재 설치된 빌드로 다시 띄운 뒤(`moai doctor`로 서빙 빌드와 설치 빌드 일치 확인), codex 감사를 한 번 받는다. 범위는 `delta_base=511d8cd964` 또는 `d325cf75d7` 기준이면 된다. 코드 수리는 필요 없다. 카드당 codex 호출 상한(규칙 d-20261106T064700Z-de2d, 최대 2회)은 이번 호출로 소진됐다. 다음 호출에는 운영자 승인이 필요하다.
- **O1 [Low] [optional] [touched-file] `.moai/docs/factory-expired-lease-dead-run.md` 신설 문단** — 문서가 합류 선택기(`mergeChosenRun`)를 "automatic paths"로 묶는다. 이 선택기는 운영자가 화면에서 고르는 대화형 병합이라 엄밀히는 자동 경로가 아니다. 동작(완화 없음)은 결정 기록과 같고 보수적인 방향이다. 다만 운영자가 선택기에서 병합을 골라도 만료 임대 중복 행 때문에 거부될 수 있다는 점은 문서에 드러나지 않는다. 신뢰도 중간. 고치려면 "the join chooser (an interactive merge, which also keeps refusing; use the explicit verb)" 정도로 한 구절 바꾸면 된다.
- **O2 [Low] [optional] [in-diff]** — 선택기와 리더 인수가 필드를 거짓으로 두는 것을 묶는 전용 시험이 없다. 지금은 구조체 리터럴에 필드가 없다는 사실로만 보장된다. 누군가 두 호출자 중 하나에서만 참으로 바꾸면 m1 같은 전역 변이와 달리 어떤 시험도 잡지 않는다. 레인도 Gaps에 적어 두었다. 결정 기록상 범위 밖이므로 차단 사유가 아니다.
- **F5 [Info] [optional]** — 1회차 F5(`card_dispatch` 재지정, `run.retired` 이벤트) 그대로. 새 시험은 `card.rebound` 1건만 확인한다.

2회차 결함 처리: N1 해소(E2, E4 m1). N2 해소(시험 존재와 내용은 이번에 읽고 실행해서 확인했고, 가드 제거 변이에서의 실패는 레인 기록 원문으로 확인). N3 해소(문서 한 줄). N4 해소(문서 한 줄). G1 미해소.

수리 정지 규칙: 이번 회차에는 수리 diff(`8792a156d4..d325cf75d7`) 안의 차단 결함이 없다. 연속 두 번 repair-born 조건은 발동하지 않는다.

허용/거부 비교(판정 함수 카드, 2회차 E5 표에 대한 이번 회차 변화):

| 경우 | 511d8cd964(기준) | d325cf75d7(HEAD) | 의도 여부 |
|---|---|---|---|
| 만료 임대 중복, 명시적 동사 | 거부 | 버림 + 은퇴 | 의도(카드 목표) |
| 만료 임대 중복, 실행 없는 동사의 자동 조정 | `AMBIGUOUS_FACTORY`, 쓰기 없음 | 같음(E4 m1 대조) | 같음 |
| 만료 임대 중복, 합류 선택기/리더 인수 | 거부 | 거부(필드 없음, E2) | 같음 |
| 유효 임대 원본 + 중복 대상 | 거부 | 거부(시험 고정) | 같음 |
| 만료 임대 + `stage` 분기, 명시적 동사 | 거부 | 거부 | 같음 |
| 두 번째 죽은 실행, 만료 시각 다름, 명시적 동사 | 거부 | 버림 + 은퇴 | 의도(F1) |
| 해석 불가 만료 시각, 명시적 동사 | 거부 | 버림 + 은퇴 | 의도(문서화, N3) |

## Baseline-attribution

- 트리: 워크트리 `/Users/goos/.moai/worktrees/projects/moai-adk-go/t2104`, 브랜치 `WT-run-leased-run`. `git rev-parse HEAD` = `d325cf75d7b871bdf58c21745ce12f43830944e2`(감사 시작 시 관측). 감사 시작 시 `git status --short` 출력 없음.
- E1, E3, E4는 이 트리에서 직접 돌렸다. E4는 `-overlay`로 스크래치 사본 파일을 끼워 넣었고 트리 파일은 바꾸지 않았다.
- Go 툴체인은 세션 기본값이다. moai 바이너리로 측정한 값은 없으므로 §2.2 도구 출처 규정은 적용되지 않는다. E6의 서버 빌드 정보는 서버가 스스로 보고한 값이다.

## Gaps

- (4회차) codex 검토는 정적 검토다. codex 스스로 시험·빌드·재현을 돌리지 않았다고 밝혔다(E7). 동적 증거는 3회차 E1·E4뿐이다.
- (4회차) `moai verify audit-plan`은 이번에도 돌리지 않았다. 게이트는 3회차에 `workflow.yaml`을 직접 읽어 확인한 값(`codex: required`)이고, 영수증 발급이 그 게이트가 required로 읽혔다는 서버 쪽 관측이다.
- (4회차) 결과의 `provenance.transport`가 `codex app-server`다. 이는 `codex_audit` MCP 도구의 내부 전송이며, 감사자가 app-server 경로를 직접 호출한 것은 아니다. `audit_multi`, `glm_audit`, `claude_audit`, `codex_review`, `glm_review`는 쓰지 않았다.
- (4회차) 종료 훅이 영수증을 "receipt unknown to the store"로 거부했다. 거부 기록 `/Users/goos/.moai/worktrees/projects/moai-adk-go/t1974/.moai/state/audit-receipts/rejections/sync-auditor--t2104.json`의 `tree_root`는 t1974(이 감사자를 띄운 세션의 트리)다. 영수증 원본은 `/Users/goos/.moai/worktrees/projects/moai-adk-go/t2104/.moai/state/audit-receipts/receipts/rcpt-121841c4ceba8e412225bddd.json`에 있고, 그 기록은 `tree_root` t2104, `head_commit` d325cf75d7…, `codex_verdict` pass, `created_at` 2026-11-06T12:34:49Z이다. 훅이 감사 대상 트리가 아니라 띄운 세션의 트리 저장소를 읽어서 생긴 불일치다. t1974를 대상으로 codex를 다시 부르면 다른 트리를 감사하게 되므로 부르지 않았다.
- (3회차 기록) 교차 모델 감사 없음(G1). codex 1회 호출이 `stale_server`로 실행 전에 거절됐다. 4회차 E7로 대체됐다.
- N2 가드 제거 변이는 이번 감사에서 다시 돌리지 않았다. 레인의 진행 기록 원문(`consolidation accepted a source row with a valid lease ... abandoned cards: t1`)을 근거로 삼았다. 이번에 내가 직접 관측한 것은 그 시험이 통과하고 내용이 2회차 Required fix와 맞는다는 점이다.
- `internal/cli`와 `internal/homestate` 전체 스위트, Windows, CI, 커버리지는 측정하지 않았다.
- 선택기와 리더 인수 경로는 코드 읽기(E2)로만 확인했고, 그 경로를 직접 구동하지는 않았다(O2).
- 워크트리 가드 거부 2건: 변수 `T`를 쓴 `sed` 복합 명령, `git show ... | grep`과 `gofmt`를 묶은 복합 명령. 둘 다 단순한 단독 명령으로 나눠 같은 결과를 얻었다.
- 1회차 E4 성격의 두 줄 상태 탐침은 이번에 다시 돌리지 않았다. 이번 수리는 그 동작을 바꾸지 않았고, 자동 경로 불변은 m1 대조 시험으로 확인했다.

## Residual-risk

- codex 판정은 결함 0건 PASS지만 정적 검토이고 근거 문장이 짧다. 두 모델이 같은 맹점을 공유할 가능성은 남는다.
- 명시적 동사 아래에서, 다른 실행에 붙어 살아 있는 레인이 임대만 만료된 상태라면 그 원본 행은 운영자 동사 한 번으로 버려진다. 지워지는 것은 없고 임대 열은 버려진 행에 남는다(레인 기록과 같음).
- O2: 선택기나 리더 인수 호출자 하나에만 필드를 켜는 미래 변경은 현재 시험이 잡지 못한다.

## Iteration history

| 회차 | 감사 커밋 | 판정 | 점수 | 차단 결함 |
|---|---|---|---|---|
| 1 | 511d8cd964 | FAIL | 0.78 | F1(여러 죽은 실행 미문서·미시험), F2(두 줄 상태 피해 미측정) |
| 2 | 8792a156d4 | FAIL | 0.85 | N1(자동 조정 경로 미공개), N2(가드-순서 불변식 미고정), G1(필수 codex 게이트 미충족) |
| 3 | d325cf75d7 | FAIL | 0.91 | G1만(필수 codex 게이트 미충족, 서버 stale 재발). 코드 차단 결함 없음 |
| 4 (이 파일) | d325cf75d7 | PASS | 0.91 | 없음. G1 해소: codex_audit(rc.2 서버 04d8c27a7e, delta_base 511d8cd964) pass, 결함 0건, 영수증 `rcpt-121841c4ceba8e412225bddd`. 잔여 optional O1·O2·F5 |
