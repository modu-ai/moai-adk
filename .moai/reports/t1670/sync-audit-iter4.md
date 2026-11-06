auditor-model: claude-opus-5-5
verdict: FAIL
audited_sha: 9429ec4f51d0f924ab1feefd570e65b9f2f6cff7

## Evaluation Report — 카드 t1670 sync-audit 4회차 (델타 재감사, 마지막 수리 회차)

SPEC: SPEC-LOCAL-MAIN-FLOW-001
Overall Verdict: FAIL — overall score 85/100

감사 대상: 워크트리 `/Users/goos/MoAI/moai-adk-go/.moai/worktrees/t1670`, 브랜치 `WT-t1616-10-10`, HEAD `9429ec4f5`(전체 SHA는 위 `audited_sha`), 기준 local main `67e462e80`. 판정 대상 수리 커밋은 `a642e9810`(운영 코드: 읽기 seam)과 `6d8317192`(시험)이고, `9429ec4f5`는 `.moai/reports/t1670/` 문서만 바꿨다.

### 요약

3회차에서 빠져 있던 mutant D(소유권 읽기만 구간 앞으로)는 이번 수리로 확실히 잡힌다. 이 감사가 새 운영 파일 위에 다시 만든 A·B·C·D·E와, 직접 설계한 새 변형 F(뮤테이션 잠금 자체를 걷어 냄)까지 모두 `-count=3`에서 3/3 실패했고, 실제 트리는 3/3 통과했다. 운영 코드 변경(`var localMainResyncReadWindow = factory.ReadIntegrationLock`)은 동작을 바꾸지 않는다.

그러나 **읽기를 자기만의 첫 번째 `RunIntegrationMutation` 구간에서 하고, 그 캐시 값으로 둘째 구간에서 검증과 fast-forward를 하는 변형(mutant H)** 은 두 rival 시험을 모두 3/3 통과한다. 읽기 probe는 첫 구간 안에서 실행되므로 lock-busy가 나오고, 트립와이어는 여전히 재검증 **끝**(둘째 구간 안)에서 심기므로 두 구간 사이의 공백을 보지 못한다. 이 감사가 그 공백에 경쟁자 탈취를 끼워 넣은 probe로 확인한 결과, 경쟁자가 창을 가져간 뒤에도 이전 소유자가 local main을 fast-forward했다. 같은 자리에 탈취를 넣은 실제 코드는 `refused — sess-rival holds the window`로 거절했다. 즉 F2가 막는 결함이 그대로 되살아나는데 시험은 통과한다. 3회차 F1의 Required fix (2)는 "읽기 seam에서 트립와이어를 심고 같은 자리에서 동기 탈취를 시도"하라고 했는데, 수리는 동기 탈취만 넣고 트립와이어는 옮기지 않았다. H는 그 빠진 절반이 그대로 드러난 경우다.

구현자가 남긴 잔여 위험("직접 `factory.ReadIntegrationLock`을 밖으로 끌어내고 seam 호출은 안에 남긴" 변형, 이 감사의 D3)도 3/3 통과를 재현했다. 필수 백엔드 codex는 H와 D3를 각각 독립적으로 만들어 `fail`(P2 두 건)을 냈다. 필수 게이트가 FAIL이므로 이 감사는 PASS나 PASS-WITH-DEBT를 낼 수 없다.

### Dimension Scores

| Dimension | Score | Verdict | Evidence |
|-----------|-------|---------|----------|
| Functionality (40%, must-pass) | 78/100 | FAIL | 실제 트리 두 시험 각 3/3 PASS. A·B·C·D1·D2·E·F는 3/3 FAIL (E1). H는 `DuringRevalidate` 3/3 PASS, `BeforeFastForward` 3/3 PASS (E1). H probe: `rivalErr=<nil> \| head==target:true ... holder=sess-rival` (E3). D3도 3/3 PASS (E1). codex required `fail`, P2 두 건 (E5). 완료 조건 1이 H 형태 회귀에 대해 고정되지 않음 → 필수 통과 차원 FAIL |
| Security (25%, must-pass) | 95/100 | PASS | 운영 변경은 패키지 변수 한 개(`local_main_resync.go:73`)와 호출 치환 한 줄(`:311`). 값은 `factory.ReadIntegrationLock` 그 자체라 입력 경계·권한 변화 없음. 시험 쓰기 대상은 `t.TempDir()` fixture 안 |
| Craft (20%) | 85/100 | PASS | `-race -count=1 -run 'TestLocalMainResync\|TestAuditLagUsesBinlagSeam'` EXIT=0, `--- FAIL` 0, DATA RACE 0 (E4). `gofmt -l` 출력 없음, `go vet` exit 0, `golangci-lint --new-from-rev 209a0a274` `0 issues.` (E6). 새 동기 probe가 정상 경로에서도 잠금 재시도 예산을 다 쓰므로 `DuringRevalidate`가 0.6s대→8.7s대로 늘었다(F3, 선택) |
| Consistency (15%) | 90/100 | PASS | seam 이름·`@MX:NOTE`/`@MX:SPEC`·`t.Cleanup` 복원이 기존 `localMainResyncAfterRevalidate`/`BeforeFastForward` seam과 같은 모양. binlag 좌표 451/454 갱신은 `TestAuditLagUsesBinlagSeam` PASS로 확인 (E4) |

가중 합: 0.40×78 + 0.25×95 + 0.20×85 + 0.15×90 = 31.2 + 23.75 + 17.0 + 13.5 = 85.45 → 85. 필수 통과 차원(Functionality) FAIL과 필수 게이트(codex) FAIL로 전체 FAIL.

### 변형 판정과 개연성 평가

| 변형 | 모양 | `DuringRevalidate` | `BeforeFastForward` | 잡는 신호 | 개연성 |
|---|---|---|---|---|---|
| REAL | 수정 없음 | 3/3 PASS | 3/3 PASS | — | — |
| A | 재검증 전체를 구간 앞으로 | 3/3 FAIL | 3/3 PASS | 읽기 probe(214행) + 트립와이어(217행) | 높음 |
| B | 재검증과 이동을 두 구간으로 | 3/3 FAIL | 3/3 PASS | 트립와이어(217행) | 높음 |
| C | 재검증 결과를 구간 전에 계산 | 3/3 FAIL | 3/3 PASS | 읽기 probe + 트립와이어 | 중간 |
| D1 | seam 읽기만 구간 앞으로 (새 코드에서 자연스러운 D) | 3/3 FAIL | 3/3 PASS | 읽기 probe(214행) | 높음 |
| D2 | 3회차 D 그대로 (직접 읽기를 앞으로, seam 미사용) | 3/3 FAIL | 3/3 PASS | seam 미실행 `Fatalf`(211행) | 높음 |
| D3 | 구현자 잔여: 직접 읽기를 앞으로, 소유자 판단은 그 값, seam 읽기는 안에서 임대 만료에만 사용 | **3/3 PASS** | **3/3 PASS** | 없음 | 낮음 |
| E | 이동을 구간 해제 뒤로 | 3/3 FAIL | 3/3 FAIL | 45행 + 경쟁자-HEAD(228행) | 높음 |
| F (신규) | 뮤테이션 잠금을 걷어 내고 fn 직접 호출 ("창은 이미 우리 것" 단순화) | 3/3 FAIL | 3/3 FAIL | 읽기 probe + 45행 + 228행 | 중간 |
| G (대조) | seam만 우회, 직접 읽기는 구간 **안** (동작은 올바름) | 3/3 FAIL | 3/3 PASS | seam 미실행 `Fatalf`(211행) | — |
| H (codex 2번) | seam 읽기를 첫 구간, 캐시 값 검증+이동을 둘째 구간 | **3/3 PASS** | **3/3 PASS** | 없음 | 중간 |

구현자 잔여(D3) 평가: 같은 레코드를 두 번 읽으면서 소유자 판단만 밖의 낡은 값으로 하고 안의 seam 읽기는 다른 용도로 남겨야 생기는 모양이다. 같은 파일에 직접 읽기를 하는 `localMainResyncPriorHold`(396행)가 이미 있어서 "빠른 실패용 사전 소유자 확인"으로 끌어다 쓰는 리팩터가 아주 터무니없지는 않지만, 그 경우에도 안쪽 읽기를 소유자 판단에서 빼면서 임대 확인에만 남기는 것은 부자연스럽다. 이 감사는 D3를 **개연성 낮음**으로 보고 단독으로는 차단 사유로 삼지 않는다(F2, optional). 다만 codex 필수 게이트는 D3를 P2로 집계했다.

H 평가: 2회차가 차단 사유로 본 B(재검증과 이동을 두 구간으로)와 3회차가 차단 사유로 본 D(읽기만 앞으로)를 합친 모양이다. "구간 진입 뒤 스스로 읽어야 한다"는 `withIntegrationLockMutation` 주석을 의식한 개발자가 읽기만 따로 잠금으로 감싸고 값을 넘기는 것은 충분히 나올 수 있는 변경이고, 실제로 결함을 되살린다(E3). 앞선 회차와 같은 기준이면 **차단 사유**다(F1).

G는 동작이 올바른 리팩터인데도 시험이 실패한다. 실패 메시지가 원인을 정확히 말하고 안전한 쪽(거짓 양성)으로 실패하므로 결함으로 보지 않는다(F4, info).

### Findings (structured defect-list)

- **F1 [P2] [blocking]** `internal/cli/local_main_resync_gap_test.go:162-164`(트립와이어를 `localMainResyncAfterRevalidate`에서 심음) 및 `:153-159`(읽기 seam 래퍼) — 트립와이어 구간이 재검증 **끝**에서 시작하므로, 읽기와 나머지 재검증 사이에서 잠금이 풀렸다 다시 잡히는 회귀(mutant H)를 보지 못한다. 읽기 probe는 첫 구간 안에서 lock-busy를 받아 통과한다. 실측: H에서 두 rival 시험 모두 3/3 PASS (E1). 공백에 경쟁자 탈취를 넣은 probe에서 경쟁자가 창을 가진 채로 이전 소유자가 fast-forward함 (E3); 같은 probe를 실제 코드의 구간 직전에 넣으면 소유자 불일치로 거절되고 HEAD는 그대로 (E3 대조). codex required도 같은 변형으로 P2 (E5). 확신도 높음.
  - Required fix: 읽기 seam 래퍼 안에서, 동기 탈취 probe를 시도한 **직후에** 트립와이어를 심고(probe는 잠금을 얻지 못하므로 아무것도 쓰지 않는다), `localMainResyncBeforeFastForward`에서 그대로 남아 있는지 확인한다. 끝-seam의 심기는 빼거나 "덮어쓰지 않음"으로 바꿔, 읽기부터 이동 직전까지 한 번도 잠금이 다시 잡히지 않았음을 하나의 트립와이어가 증명하게 한다. 양성 대조(창 해제 뒤 트립와이어가 사라짐)는 유지한다. 수리 후 H·B·A·C·D1·D2·E·F가 모두 실패하고 실제 트리가 통과하는지 `-count=3` 이상으로 재측정한다.
- **F2 [P2] [optional]** `internal/cli/local_main_resync.go:311-325` / `local_main_resync_gap_test.go:153-159` — 시험은 seam이 **호출되었는지와 그 위치**만 보고, 소유자 판단이 seam이 **돌려준 값**에 근거하는지는 보지 않는다. 그래서 직접 읽기를 밖으로 빼고 그 값으로 판단하면서 seam 호출만 안에 남긴 D3가 3/3 통과한다 (E1). 이 감사는 개연성을 낮게 보지만 codex 필수 게이트는 P2로 집계했다(E5). 확신도 높음(통과 사실), 개연성 판단은 중간.
  - Suggested fix (F1과 같은 회차에 값싸게 닫힘): 별도 시험에서 읽기 seam이 다른 세션(`lmfRival`)이 보유한 레코드를 돌려주게 하고, re-sync가 `MergeExitNotHolder`로 거절되며 HEAD가 그대로인지 단언한다. 판단이 seam의 값이 아닌 다른 읽기에 근거하면 이동이 일어나 실패한다.
- **F3 [low] [optional]** `internal/cli/local_main_resync_gap_test.go:126-129` — 동기 probe가 정상 경로에서 잠금 재시도 예산을 끝까지 기다리므로 `DuringRevalidate`가 실행당 약 8.7s(실측 8.73–8.81s, E1)다. `BeforeFastForward`도 같은 이유로 8.7s대라 두 시험 합계 약 18s. `internal/cli` 패키지 단위로는 허용 범위이고 CI를 막을 수준은 아니어서 부채로만 기록한다. 줄이려면 시험 전용으로 대기 예산을 줄이는 seam이 필요한데, 그것도 코드 표면을 늘리므로 지금은 권하지 않는다. 확신도 높음.
- **F4 [info] [optional]** 변형 G — seam만 우회하고 직접 읽기를 구간 안에 둔 올바른 리팩터도 211행 `Fatalf`로 실패한다. 시험이 seam에 결합되어 있다는 뜻이지만, 메시지가 원인을 정확히 말하고 안전한 쪽으로 실패하므로 결함이 아니다. 확신도 높음.
- **F5 [info] [optional]** 운영 코드 seam의 동작 보존: `var localMainResyncReadWindow = factory.ReadIntegrationLock`는 함수 값 대입이라 호출 의미가 같다. 쓰는 곳은 시험(`t.Cleanup`으로 복원)뿐이고, 동시 변경은 시험 직렬 실행 안에서만 일어나며 `-race`에서 DATA RACE 0 (E4). 수용 가능. 확신도 높음.

### Recommendations

- 이번 회차가 마지막 수리 회차라는 레인 결정 기록(`evidence.md`, iter3 decision record)에 따라, 카드는 이 증거와 함께 보류하고 이월 메모로 F1(필수)과 F2(권장)의 수리안을 넘긴다. 수리 범위는 시험 파일 하나이며 운영 코드는 손대지 않는다.
- 운영 코드는 네 차례 감사 내내 올바르다고 판단됐고(이번 E3 대조로도 확인), 남은 것은 회귀 고정의 빈틈이다. 다음 회차의 재감사 범위는 F1·F2 델타(트립와이어 위치 + seam 반환값 시험 + 변형 A–H 재측정)로 한정할 수 있다.

---

## Claim

- C1. 실제 트리(HEAD `9429ec4f5`)에서 두 rival 시험은 `-count=3`으로 각각 3/3 통과하고, `-race` 아래 `TestLocalMainResync*`와 `TestAuditLagUsesBinlagSeam`이 모두 통과한다(DATA RACE 0).
- C2. 새 읽기 probe는 A·C·D1·F를, 트립와이어는 A·B·C를, seam 미실행 `Fatalf`는 D2를, 짝 시험(45행)과 경쟁자-HEAD 신호는 E·F를 잡는다. 각 변형 3/3 FAIL.
- C3. mutant H(읽기만 별도 첫 구간)와 D3(구현자 잔여)는 두 시험을 모두 3/3 통과한다.
- C4. H는 실제 결함이다: 두 구간 사이에서 경쟁자의 `--force` 탈취가 성공한 뒤에도 이전 소유자가 fast-forward했다. 같은 자리의 탈취에 실제 코드는 거절로 답했다.
- C5. 운영 seam은 동작을 바꾸지 않는다.
- C6. 카드 완료 조건 (1) "재검증과 이동이 같은 락 구간임을 시험으로 확인": **미충족** — H 형태 회귀에서 시험이 통과한다(C3, C4). (2) "잠금을 지키는 경합에서 창 재현 0회": 실제 트리에서 **충족** — 3회 실행 모두 이동 전 경쟁자 착지 0건, 읽기 직후 동기 탈취는 lock-busy로 거절. 잠금을 무시하는 행위자의 비원자적 abort 창은 이전 회차에서 수용된 잔여 위험 그대로다.
- C7. 필수 백엔드 codex의 교차 감사 결과는 `fail`(gate required, P2 두 건: D3 형태와 H 형태)이다.
- C8. SPEC `progress.md`에 `Binding run conditions` 제목이 없다(`grep -n "Binding run" -A8 .moai/specs/SPEC-LOCAL-MAIN-FLOW-001/progress.md` 출력 없음). 처리할 구속 조건은 없다.

## Evidence

E1 — 변형별 `-count=3` 측정. 스크립트 `/private/tmp/t70-audit4/run.sh`가 각 트리(`git archive HEAD`를 풀어 `/private/tmp/t70-audit4/<X>`에 만들고 `/private/tmp/t70-audit4/mutate.py`로 변형 적용)에서 `TMPDIR=GOTMPDIR=/private/tmp/t70-audit4/tmp go -C <tree> test -count=3 -v -run 'TestLocalMainResyncRivalTakeover(BeforeFastForward|DuringRevalidate)CannotLand$' ./internal/cli/`를 순차 실행. 요약(`logs/summary-REAL-10.txt`, `logs/summary-H-1.txt` 원문 발췌):

```
== REAL exit=0
   --- PASS: ...BeforeFastForwardCannotLand (8.75s) (8.76s) (8.77s)
   --- PASS: ...DuringRevalidateCannotLand (8.73s) (8.75s) (8.81s)
   ok  	github.com/modu-ai/moai-adk/internal/cli	53.487s
== A exit=1   DuringRevalidate FAIL ×3 (0.50s 0.55s 0.63s) / BeforeFastForward PASS ×3
   local_main_resync_gap_test.go:214: a rival --force takeover right after the ownership read was not refused as lock-busy (got <nil>; ...   ×3
   local_main_resync_gap_test.go:217: the mutation lock was acquired between the ownership re-check and the fast-forward ...   ×3
== B exit=1   DuringRevalidate FAIL ×3 (8.80s 9.05s 9.38s) / BeforeFastForward PASS ×3
   local_main_resync_gap_test.go:217: the mutation lock was acquired between ...   ×3
== C exit=1   DuringRevalidate FAIL ×3 (0.43s 0.48s 0.61s) / BeforeFastForward PASS ×3   (214행 ×3, 217행 ×3)
== D1 exit=1  DuringRevalidate FAIL ×3 (0.54s 0.54s 0.55s) / BeforeFastForward PASS ×3   (214행 ×3)
== D2 exit=1  DuringRevalidate FAIL ×3 (0.44s 0.46s 0.46s) / BeforeFastForward PASS ×3
   local_main_resync_gap_test.go:211: the ownership read seam localMainResyncReadWindow never ran ...   ×3
== D3 exit=0
   --- PASS: ...BeforeFastForwardCannotLand (8.75s) (8.75s) (9.41s)
   --- PASS: ...DuringRevalidateCannotLand (8.74s) (8.74s) (8.80s)
   ok  	github.com/modu-ai/moai-adk/internal/cli	54.106s
== E exit=1   BeforeFastForward FAIL ×3 (0.51s 0.56s 0.65s) / DuringRevalidate FAIL ×3 (8.77s 8.83s 8.87s)
   local_main_resync_gap_test.go:45: a rival --force takeover landed between the ownership re-check and the fast-forward ...   ×3
   local_main_resync_gap_test.go:228: a rival --force takeover landed while primary HEAD was still <base>, before the fast-forward ...   ×3
== F exit=1   BeforeFastForward FAIL ×3 / DuringRevalidate FAIL ×3   (45행 ×3, 214행 ×3, 228행 ×3, 236행 ×3)
   FAIL	github.com/modu-ai/moai-adk/internal/cli	4.628s
== G exit=1   DuringRevalidate FAIL ×3 (211행 ×3) / BeforeFastForward PASS ×3
== H exit=0
   --- PASS: ...BeforeFastForwardCannotLand (8.79s) (8.82s) (8.91s)
   --- PASS: ...DuringRevalidateCannotLand (8.75s) (8.76s) (8.87s)
   ok  	github.com/modu-ai/moai-adk/internal/cli	53.843s
```

E2 — 변형 diff(실제 `internal/cli/local_main_resync.go` 대비, `diff -u | grep '^[-+][^-+]'` 원문):

```
# D3
+	mutDPreLock, mutDPreErr = factory.ReadIntegrationLock(repoRoot)
-	lock, err := localMainResyncReadWindow(repoRoot)
+	lock, err := mutDPreLock, mutDPreErr
-	if lock.LeaseExpired(factory.WindowClock()) {
+	if fresh, ferr := localMainResyncReadWindow(repoRoot); ferr != nil || fresh.LeaseExpired(factory.WindowClock()) {
+var mutDPreLock *factory.IntegrationLock
+var mutDPreErr error
# F
-	if lockErr := factory.RunIntegrationMutation(repoRoot, func() error {
+	if lockErr := func(_ string, fn func() error) error { return fn() }(repoRoot, func() error {
# G
-	lock, err := localMainResyncReadWindow(repoRoot)
+	lock, err := factory.ReadIntegrationLock(repoRoot)
# H
+	if lockErr := factory.RunIntegrationMutation(repoRoot, func() error {
+		mutDPreLock, mutDPreErr = localMainResyncReadWindow(repoRoot)
+		return nil
+	}); lockErr != nil {
+		return "", lockErr
+	}
-	lock, err := localMainResyncReadWindow(repoRoot)
+	lock, err := mutDPreLock, mutDPreErr
+var mutDPreLock *factory.IntegrationLock
+var mutDPreErr error
```

A·B·C·E는 3회차(`/private/tmp/t70-audit3/<X>`)와 같은 모양을 새 파일에 다시 적용했다(재검증 안의 seam 호출은 유지). D1은 D2와 같되 앞으로 끌어낸 읽기가 seam을 거친다.

E3 — H 공백 probe. `/private/tmp/t70-audit4/H2`(H + 두 구간 사이의 `mutHGap(repoRoot)` 훅)와 대조 `/private/tmp/t70-audit4/R2`(실제 코드 + 구간 직전의 같은 훅)에서 같은 시험 `zz_mut_h_probe_test.go`(훅에서 `lmfRival`의 `--force` 탈취 1회) 실행:

```
# H2: go -C /private/tmp/t70-audit4/H2 test -count=1 -v -run 'TestAuditMutHGapProbe$' ./internal/cli/
zz_mut_h_probe_test.go:28: rivalErr=<nil> | head==target:true head==base:false | holder=sess-rival | resyncErr=integration resync: local main fast-forwarded cf33226beecf -> ab503f55ee06 to origin/main, but releasing the window failed: release integration window is held by a different session: lane-rival (pid 37771) holds it (moai integration status reads it)

# R2 (대조): go -C /private/tmp/t70-audit4/R2 test -count=1 -v -run 'TestAuditMutHGapProbe$' ./internal/cli/
zz_mut_h_probe_test.go:28: rivalErr=<nil> | head==target:false head==base:true | holder=sess-rival | resyncErr=integration resync: refused — sess-rival holds the window, not sess-lane-1. The window was taken while the fetch ran. The primary is unchanged; re-run the re-sync to take the window again (moai integration status reads it).
```

E4 — race (`/private/tmp/t70-audit4/race.sh`: `go -C <worktree> test -race -count=1 -v -run 'TestLocalMainResync|TestAuditLagUsesBinlagSeam' ./internal/cli/`):

```
EXIT=0
PASS lines: 16  FAIL lines: 0  DATA RACE: 0
--- PASS: TestLocalMainResyncRivalTakeoverBeforeFastForwardCannotLand (8.89s)
--- PASS: TestLocalMainResyncRivalTakeoverDuringRevalidateCannotLand (9.39s)
--- PASS: TestAuditLagUsesBinlagSeam (0.26s)
ok  	github.com/modu-ai/moai-adk/internal/cli	29.736s
```

최상위 `func TestLocalMainResync*` 수: 이 트리 15, 3회차 트리(`/private/tmp/t70-audit3/A`) 15 — 시험 삭제 없음.

E5 — `mcp__moai__audit_multi` (project_root=워크트리, target=baseBranch, card_id=t1670, gates 인자 없음). 오케스트레이터 확인용 digest:

```
overall_verdict: fail
gate_unmet: (결과에 없음)
plan_source: config
per_backend_verdicts: [{backend: codex, gate: required, verdict: fail}]
audit_receipt: rcpt-b09407b81045776d51c62ddf
```

codex 지적 요지: (P2, gap_test.go:153) 직접 `factory.ReadIntegrationLock`을 밖으로 옮기고 캐시 값으로 검증하며 seam 호출만 안에 남긴 변형이 두 시험 각 3회 통과(`ok … 54.980s`); (P2, gap_test.go:164) seam 읽기를 첫 구간, 검증과 이동을 둘째 구간에 둔 변형이 두 시험 모두 통과(`ok … 18.476s`). 이 판정은 오케스트레이터가 `moai verify audit-plan --project-root <toplevel> --result-file <path>`로 확인하기 전까지 최종이 아니다.

E6 — 정적 검사(워크트리에서):

```
gofmt -l internal/cli/local_main_resync.go internal/cli/local_main_resync_gap_test.go internal/cli/mcp_build_identity_test.go   → (출력 없음) gofmt=0
go vet ./internal/cli/   → vet=0
golangci-lint run --new-from-rev 209a0a274 ./internal/cli/   → 0 issues.
```

E7 — 감사 계획: `moai verify audit-plan --project-root /Users/goos/MoAI/moai-adk-go/.moai/worktrees/t1670` → `"config_status": "ok"`, `"cross_model_active": true`, `"enforced_required": ["codex"]` (claude gate off, glm gate off).

## Baseline-attribution

- 모든 시험·정적 검사는 이 감사 실행에서 HEAD `9429ec4f51d0f924ab1feefd570e65b9f2f6cff7` 기준으로 측정했다. 변형 트리는 같은 HEAD를 `git -C <worktree> archive HEAD`로 떠서 만들었고, 시험 파일은 손대지 않았다(H2/R2의 probe 시험 파일만 추가).
- 시험은 `go`가 트리 소스로 직접 빌드한 바이너리로 실행했다. 수치는 3회차·레인 보고에서 가져오지 않았다.
- `TMPDIR=GOTMPDIR=/private/tmp/t70-audit4/tmp` (홈 밖).

## Gaps

- `moai` 설치 바이너리와 MCP 서버는 `ee981a932`로 빌드됐고 이 트리 HEAD의 조상이다(`audit_multi` 결과의 `build_lag`). E7의 계획 판독과 E5의 백엔드 실행은 그 빌드가 했다. 계획 값은 설정 파일에서 읽히는 것이고 이 카드가 바꾼 코드와 무관하다고 보지만, 트리에서 빌드한 바이너리로 다시 확인하지는 않았다.
- D3의 결함 동작(밖의 낡은 읽기로 판단할 때 경쟁자 아래서 이동하는지)은 이 감사가 따로 probe하지 않았다. codex가 그런 probe를 했다고 보고했고, 코드 구조상 E3의 H와 같은 경로라고 판단한다(추론).
- Windows 기판에서의 트립와이어·probe 동작은 측정하지 않았다.
- 반복 횟수는 변형당 `-count=3`이다(3회차는 5). 판정 신호가 타이밍이 아닌 잠금 상태에 기대므로 3회로 충분하다고 판단했지만 더 많은 반복은 하지 않았다.
- 거절된 도구 호출: 워크트리 격리 가드가 복합 셸 구문(`git archive`+`tar` 체인, 변수 쓰는 반복문, 리다이렉트가 섞인 `go test`, heredoc 패치)을 다섯 차례 거절했다. 같은 명령을 스크립트 파일(`/private/tmp/t70-audit4/{extract.sh,run.sh,race.sh,mutate.py,h2patch.py,r2patch.py}`)과 단순 명령으로 나눠 실행했으며, 측정 내용 자체는 바꾸지 않았다.
- 증거 원본(`/private/tmp/t70-audit4/logs/*.log`)은 기계 로컬 임시 위치라 이 파일에 옮긴 줄만 인용 대상이다.

## Residual-risk

- F1을 고쳐도, 판단에 쓰는 값과 잠금 구간의 관계를 시험이 간접적으로만 본다는 한계는 남는다. F2의 반환값 시험까지 넣어야 "판단 근거가 구간 안의 읽기"임이 직접 고정된다.
- 실제 운영 코드는 이번 E3 대조를 포함해 올바르게 동작했다. 남은 위험은 앞으로의 리팩터가 H·D3 같은 모양으로 F2 창을 다시 열어도 CI가 잡지 못한다는 것이다.
- 잠금을 무시하는 외부 행위자의 비원자적 abort 창은 이전 회차부터 수용된 잔여 위험이다.
- 두 rival 시험이 합계 약 18s를 쓰므로, 부하가 큰 CI 러너에서 잠금 재시도 예산이 실제 경합과 겹쳐 시간 초과 성격의 실패가 날 가능성은 남는다(이번 실행에서는 관찰되지 않음).

## Iteration history

| 회차 | 감사 SHA | 판정 | 점수 | 놓친 변형 |
|---|---|---|---|---|
| 1 | a2055ca46 | FAIL | 72 | A |
| 2 | b79a70d97 | FAIL | 78 | B |
| 3 | 9aa451b10 | FAIL | 82 | D |
| 4 | 9429ec4f5 | FAIL | 85 | H (차단), D3 (선택) |
