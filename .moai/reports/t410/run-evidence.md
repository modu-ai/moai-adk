# SPEC-DRIFT-CLOSE-BODY-001 — run-phase 원장 (카드 t410)

각 마일스톤 경계에서 `verification-claim-integrity.md` §3의 5절(Claim / Evidence / Baseline-attribution / Gaps / Residual-risk)로 기록한다. Evidence에는 **명령과 축자 출력과 exit code**를 적는다 — 요약은 증거가 아니다.

## 좌표 (run-phase 착수 시점)

```
$ git rev-parse --show-toplevel
/Users/goos/MoAI/moai-adk-go/.claude/worktrees/t410
$ git branch --show-current
WT-drift-false-positive
$ git rev-parse --short HEAD
c323bb491
$ git status --short
?? .moai/reports/t410/drift-before.txt
```

착수 HEAD `c323bb491` · 브랜치 `WT-drift-false-positive` · 트리 `.claude/worktrees/t410`.

## §C 사전 점검 — 오케스트레이터 실행분 인용 (재실행하지 않음)

| 명령 | 결과 |
|---|---|
| `go build -o /tmp/moai-t410 ./cmd/moai` | rc 0 |
| `/tmp/moai-t410 spec drift --no-cache > .moai/reports/t410/drift-before.txt` | rc 0, 637 SPEC 행 |
| `grep SPEC-V3R6-SESSION-HANDOFF-AUTO-001 .moai/reports/t410/drift-before.txt` | `completed / in-progress / DRIFT` |

plan.md §C의 정지 조건(확정 대상 행이 이미 DRIFT가 아니면 blocker 반환)은 **불성립** — 결함이 이 트리에 재현되므로 착수한다.

**귀속 주의** — 위 세 줄은 오케스트레이터가 이 트리에서 실행한 측정을 **인용**한 것이지 이 에이전트가 재측정한 것이 아니다. AC-DCB-004/005의 "수리 전" 표는 M3에서 이 트리 좌표와 함께 다시 확인한다.

---

## M1 — 술어 정의와 RED

**Claim** — 픽스처가 결함을 실제로 재현하고, 수리 이전에 붉다.

**Evidence**

RED-1 (컴파일 실패, 구현이 아직 없음). HEAD `c323bb491`:

```
$ go test ./internal/spec/ -run TestDriftCloseBody -count=1 ; echo rc=$?
# github.com/modu-ai/moai-adk/internal/spec [github.com/modu-ai/moai-adk/internal/spec.test]
internal/spec/drift_close_body_test.go:228:10: undefined: bodyDeclaresClose
internal/spec/drift_close_body_test.go:251:6: undefined: bodyDeclaresClose
internal/spec/drift_close_body_test.go:356:5: undefined: inMemBodyDeclaredClose
internal/spec/drift_close_body_test.go:390:7: undefined: bodyDeclaresClose
internal/spec/drift_close_body_test.go:395:5: undefined: bodyDeclaresClose
FAIL	github.com/modu-ai/moai-adk/internal/spec [build failed]
FAIL
rc=1
```

RED-2 (단언 층). 컴파일 실패는 "심볼이 없다"만 말하고 **픽스처가 결함을 재현하는지는 말하지 않는다**. 그래서 두 함수를 `return false` 스텁으로 두고 다시 쟀다 — 이때 통과하는 음성 케이스는 스텁이라서 통과한 것이고, 실패하는 양성 케이스가 결함의 재현이다. HEAD `88b5ff686`:

```
$ go test ./internal/spec/ -run TestDriftCloseBody -count=1 ; echo rc=$?
--- FAIL: TestDriftCloseBody_BodyDeclaredCloseRecognized (0.00s)
    drift_close_body_test.go:88: GitImpliedStatus = "in-progress", want "completed" — the body line `- SPEC-FIX-ALPHA-001: Mx verdict ...` declares the close
    drift_close_body_test.go:91: Drifted = true, want false — a body-declared close must clear the false positive
--- FAIL: TestDriftCloseBody_PredicateTable (0.00s)
    (라인 1·2·3·4·12 = false, want true — 선언 5줄 전부 붉음)
--- FAIL: TestDriftCloseBody_FullBodyScan (0.00s)
--- FAIL: TestDriftCloseBody_ListMarkersStripped (0.00s)
FAIL	github.com/modu-ai/moai-adk/internal/spec	0.427s
rc=1
```

RED-2에서 **통과한** 것이 공허 방지의 핵심이다 — `TestDriftCloseBody_PrimaryWalkYieldsInProgress`(1차 워크가 실제로 `in-progress`를 낸다) · `_GammaPrimaryWalkPrecondition` · `_DeltaPrimaryWalkPrecondition`(둘 다 비-terminal 비-`completed`) · `_FallbackOnlyInputGate` · `_MentionIsNotClose` · `_ExistingFallbackStillFirst`. 즉 픽스처는 fallback 게이트에 **도달하며**(D7이 지적한 전제), 1차 워크는 결함을 재현한다. 전자가 거짓이면 어떤 뮤턴트도 죽지 않고, 후자가 거짓이면 GREEN이 아무것도 주장하지 않는다.

**Baseline-attribution** — 트리 `.claude/worktrees/t410`, 브랜치 `WT-drift-false-positive`. RED-1은 HEAD `c323bb491`(커밋 전 워킹트리), RED-2는 HEAD `88b5ff686`(M1 커밋 후, 스텁은 미커밋). `go vet` 및 gofmt는 M2에서 측정.

**Gaps** — M1에서는 코퍼스를 재지 않았다. `drift-before.txt`는 오케스트레이터 측정분이고 이 에이전트가 재측정한 것은 M3다. 12줄 픽스처의 **출처 정확성**은 각 커밋 본문을 `git show -s --format=%b <sha>`로 직접 읽어 대조했으나, 그 대조 자체의 축자 출력은 이 원장에 싣지 않았다(줄이 길어 인용이 본문을 압도한다) — 대신 각 케이스가 `source` 필드로 커밋을 명명하므로 재현 가능하다.

**Residual-risk** — 12줄은 **측정된 표본이지 전수가 아니다**. spec.md §6.1이 적었듯 LOOSE-only 18건 중 4건만 본문 대조됐다. 이 목록 밖의 반례는 M3 전수 대조에서만 드러난다.

---

## M2 — fallback 배선과 뮤테이션

**Claim** — 본문 선언 close가 세 번째 FALLBACK-ONLY 축으로 배선됐고, 술어는 12줄 전부를 맞히며, 명명된 뮤턴트 넷이 **전부 죽는다**.

**Evidence**

배선: `internal/spec/drift_index.go`에 `inMemBodyDeclaredClose` + `bodyDeclaresClose` + `hasDeclarationDenyKey` 추가, `internal/spec/drift.go` ① 블록에 `else if`로 연결(기존 `inMemCombinedScopeClose`가 **먼저** — REQ-DCB-006).

GREEN:

```
$ gofmt -l internal/spec/          # 출력 없음
$ go vet ./internal/spec/ ; echo rc=$?
rc=0
$ go test ./internal/spec/ -run TestDriftCloseBody -count=1 -v ; echo rc=$?
--- PASS: TestDriftCloseBody_PrimaryWalkYieldsInProgress (0.00s)
--- PASS: TestDriftCloseBody_BodyDeclaredCloseRecognized (2.21s)
--- PASS: TestDriftCloseBody_GammaPrimaryWalkPrecondition (0.00s)
--- PASS: TestDriftCloseBody_MentionIsNotClose (0.00s)
--- PASS: TestDriftCloseBody_PredicateTable (0.00s)
--- PASS: TestDriftCloseBody_FullBodyScan (0.00s)
--- PASS: TestDriftCloseBody_FallbackOnlyInputGate (0.01s)
    --- PASS: .../(a)_in-progress (0.00s)
    --- PASS: .../(b)_superseded (0.00s)
    --- PASS: .../(c)_draft (0.01s)
--- PASS: TestDriftCloseBody_DeltaPrimaryWalkPrecondition (0.00s)
--- PASS: TestDriftCloseBody_OutputIsCompletedOrNothing (0.01s)
--- PASS: TestDriftCloseBody_SubjectNamingSpecIDIsPrimaryWalkTerritory (0.00s)
--- PASS: TestDriftCloseBody_ExistingFallbackStillFirst (0.00s)
--- PASS: TestDriftCloseBody_ListMarkersStripped (0.00s)
ok  	github.com/modu-ai/moai-adk/internal/spec	2.676s
rc=0
```

**뮤테이션 4종** — 각각 소스를 일시 편집해 실행하고 되돌렸다(`cp /tmp/t410_drift_index.orig.go`로 복원, 매번 rc 확인).

| # | 뮤턴트 | 죽는가 | 죽인 입력 |
|---|---|---|---|
| 1 | 술어 전체를 `strings.Contains(body, specID)`로 교체 | **죽음** rc 1 | 5·6·7번 줄 + `_MentionIsNotClose` + `_OutputIsCompletedOrNothing` |
| 2 | 모양 A의 `:` 요구 제거 (`HasPrefix(stripped, specID)`) | **죽음** rc 1 | 8번(줄바꿈 함정) · 10번(`:` 없음) |
| 3 | 첫 모양-B 줄에서 반환(전수 훑기 제거) | **죽음** rc 1 | `_FullBodyScan`(11·12 쌍) **단독** — 1~10번 표는 통과 |
| 4 | 모양 B의 conventional-commit 구조 전제 제거 | **죽음** rc 1 | 7번 줄 |

축자 출력(발췌):

```
# 뮤턴트 1
drift_close_body_test.go:230: line 5 (a83934d55, SPEC-AUTONOMY-TIERS-001): bodyDeclaresClose = true, want false
--- FAIL: TestDriftCloseBody_MentionIsNotClose (1.24s)
--- FAIL: TestDriftCloseBody_OutputIsCompletedOrNothing (0.01s)

# 뮤턴트 2
drift_close_body_test.go:230: line 8 (51d18d3fe, SPEC-INTERNAL-SECURITY-001): bodyDeclaresClose = true, want false — 72-column wrap artifact
drift_close_body_test.go:230: line 10 (2f449e189, SPEC-GLM-KEY-INPUT-001): bodyDeclaresClose = true, want false — no ':' after the ID

# 뮤턴트 3
--- FAIL: TestDriftCloseBody_FullBodyScan (0.00s)
    bodyDeclaresClose = false, want true — the scan must sweep the whole body

# 뮤턴트 4
drift_close_body_test.go:230: line 7 (80dea9684, SPEC-INTERNAL-TEST-002): bodyDeclaresClose = true, want false
```

뮤턴트 3이 **`_FullBodyScan` 하나에서만** 죽는 것은 판정서 D2의 판독(10줄 픽스처로는 이 뮤턴트가 안 죽는다)을 실행으로 확인한 것이다 — D2 잔여 위험 (2)("실행이 아니라 판독으로 판정했다")가 이로써 닫힌다.

**뮤턴트 4는 계획에 없던 발견이다.** spec.md §5.3은 모양 B를 "표지를 걷어낸 줄이 그 자체로 conventional-commit subject → 기존 체인에 그대로 먹인다"로 정의한다. 체인에 **그대로** 먹이면 7번 줄이 `completed`를 낸다 — `ClassifyPRTitle`이 close-infix를 prefix 루프보다 먼저 보는데(`transitions.go`), 7번은 산문이면서 `3-phase close` 리터럴을 담고 있기 때문이다. 그 줄은 `SPEC-INTERNAL-TEST-002`를 **후속** SPEC으로 지목한다 — close의 정반대다. 따라서 "그 자체로 conventional-commit subject"를 **기계적 전제**(`^[a-z]+(\(...\))?: `)로 세워야 하고, 세우지 않으면 §6.1이 오탐으로 못박은 행 하나가 그대로 해제된다. 구현은 전제를 세웠고 뮤턴트 4가 그 전제의 필요성을 실측한다.

**후보 창 게이트 결정 (§5.4)** — subject 쪽 close 신호를 **쓰지 않는다**. 게이트는 (a) subject가 specID를 담지 않을 것 하나뿐이며, 판별 무게는 전부 본문 줄 술어가 진다(§5.4 조건 1·2). 조건 3은 subject close 신호를 쓸 때만 발동하므로 해당 없음. `closeInfixMatch` 재사용은 §5.4가 이미 측정으로 기각했고, 그 상수 집합은 `shouldSkipCommitTitle`·combined-scope gate·`ClassifyPRTitle` 셋이 공유하므로 넓히는 것은 REQ-DCB-006 위반이다(판정서 부록 B와 동일 판단).

**skip 필터 결정 (plan.md §B-3)** — **우회하지 않는다**. 모양 B는 `shouldSkipCommitTitle`을 통과해야 하며, 근거는 코드 주석에 남겼다: subject였다면 AC-LSCSK-003 / REQ-DCA-002가 거부했을 텍스트가 본문 줄이라는 이유로 결론에 도달하면 그 가드를 뒷문으로 여는 것이다.

**패키지 회귀** — `go test ./internal/spec/... -count=1 -v` rc 1, PASS 1207건, FAIL **1건**:

```
catalog_hash_test.go:192: CATALOG_HASH_DRIFT: entry "sync-auditor" | stored=f1b4487f… | computed=545d03d9…
  (source=internal/template/templates/.claude/agents/moai/sync-auditor.md) — run gen-catalog-hashes.go --all to refresh
--- FAIL: TestCatalogHashParity (0.18s)
```

**상속 적색이며 이 카드 소관이 아니다.** 귀속:

```
$ git diff --stat c323bb491..HEAD          # 내 변경 4파일 — 템플릿 트리 0
$ git log -1 --format='%h %ad %s' --date=short -- internal/template/templates/.claude/agents/moai/sync-auditor.md
4244c4a06 2026-09-02 docs(t386/t387): sync-auditor export-mandate clause — unblocked after lane-9 t302 settled
```

AC-DCB-006이 명명한 가드가 **실제로 실행됐음**을 `-v` 테스트 이름으로 확인(셀렉터 0매치는 초록이 아니다):

```
--- PASS: TestGetGitImpliedStatus_ChoreSkip (3.62s)                    # drift_chore_skip_test.go (AC-LSCSK-003)
--- PASS: TestGetGitImpliedStatus_SPECIDWordBoundary (0.00s)           # drift_specid_grep_test.go (LSGF-001)
--- PASS: TestCombinedScopeCloseMatches (0.00s)                        # drift_combined_scope_test.go
--- PASS: TestDetectDrift_CombinedScopeFallback (1.45s)
--- PASS: TestDetectDrift_CombinedScopeCollisionGuard (1.09s)
--- PASS: TestDetectDrift_CombinedScopeNoCloseInfixNoFallback (1.41s)
--- PASS: TestDetectDrift_Characterization_FiveCategories (1.44s)       # drift_characterization_test.go
--- PASS: TestDetectDrift_Characterization_ChoreSkipAndWordBoundary (1.86s)
--- PASS: TestDetectDrift_ConstantGitLogInvocations (0.27s)             # drift_seam_test.go
--- PASS: TestDetectDrift_PreFilterPerformsNoGitWork (0.13s)
--- PASS: TestDetectDrift_CacheHitPerformsZeroGitLogWork (0.01s)
```

**의존 패키지 재실행** — 재실행 범위는 파일 델타 패키지(`internal/spec`) ∪ 그것을 import 하는 패키지로 잡았다:

```
$ grep -rln '"github.com/modu-ai/moai-adk/internal/spec"' --include='*.go' . | ...
internal/cli  internal/epic  internal/harness/router  internal/hook  internal/web

$ go test ./internal/epic/... ./internal/harness/router/... ./internal/web/... -count=1 ; echo rc=$?
ok  	github.com/modu-ai/moai-adk/internal/epic	0.778s
ok  	github.com/modu-ai/moai-adk/internal/harness/router	0.805s
ok  	github.com/modu-ai/moai-adk/internal/web	11.527s
rc=0

$ go test ./internal/hook/... -count=1 ; echo rc=$?
ok  	github.com/modu-ai/moai-adk/internal/hook/perf	98.664s
ok  	github.com/modu-ai/moai-adk/internal/hook/quality	40.942s
ok  	github.com/modu-ai/moai-adk/internal/hook/security	18.875s
ok  	github.com/modu-ai/moai-adk/internal/hook/testutil	4.790s
ok  	github.com/modu-ai/moai-adk/internal/hook/trace	4.183s
rc=0
```

```
$ go test ./internal/cli/... -count=1 -timeout 900s ; echo rc=$?
(17개 패키지 전부 ok, FAIL 0)
ok  	github.com/modu-ai/moai-adk/internal/cli/update/report	5.515s
ok  	github.com/modu-ai/moai-adk/internal/cli/wizard	4.581s
ok  	github.com/modu-ai/moai-adk/internal/cli/worktree	20.568s
rc=0
```

**Baseline-attribution** — 전부 이 실행에서, 이 트리(`.claude/worktrees/t410`), 브랜치 `WT-drift-false-positive`, M1 커밋 `88b5ff686` 위의 미커밋 워킹트리에 대해 측정.

**Gaps** — (a) `go test ./...` 전체 스위트는 **의도적으로 돌리지 않았다**(CLAUDE.local.md §6: 병렬 레인이 함께 돌려 load 413에 이른 사고). 전 패키지 판정은 CI 몫이다. (b) 크로스플랫폼 빌드(`GOOS=windows`)를 재지 않았다 — 이 변경은 순수 Go 문자열/정규식이고 syscall·빌드태그를 건드리지 않는다. (c) `golangci-lint`를 돌리지 않았다.

**Residual-risk** — (1) 술어는 **후보이지 확정이 아니다**(spec.md §5.3 말미). 12줄 밖의 반례는 M3 전수 대조에서만 드러나며, 나오면 술어를 좁히거나 그 행을 미해제로 남긴다 — **넓혀서 삼키지 않는다**. (2) 모양 B가 기존 체인을 재사용하므로, 훗날 `ClassifyPRTitle`이나 `closeInfixMatch`가 넓어지면 이 축의 판정 표면도 함께 넓어진다. 그 결합은 의도된 것(새 텍스트 술어를 만들지 않는 대가)이지만, 저 두 함수를 고치는 카드는 이 축도 함께 재측정해야 한다. (3) 상속 적색 `TestCatalogHashParity` 아래에 이 변경이 만든 다른 적색이 가려져 있을 가능성은 낮다 — FAIL 카운트가 정확히 1이고 그 1건이 템플릿 해시라서 — 그러나 `internal/spec` 밖 전 패키지 판정은 (a)에 따라 CI에 남겨 뒀다.

**절차 관측 (기록 의무)** — M1 스테이징 중 `git add`가 한 번 `Unable to create ... index.lock: File exists`로 실패했다. 직후 `ls`로 확인했을 때 lock 파일은 이미 없었고(과도적 경합), 재실행이 정상 통과했다. 이 트리에 대한 외부 작성자를 관측한 것은 아니며 `git status`는 내 파일만 보였다.

---

## M3 — 코퍼스 전수 판정

**Claim** — 이 트리·이 브랜치에서 수리 전후 표를 각각 측정해 대조한 결과, DRIFT → 해제 **18행**, 비-DRIFT → DRIFT **0행**. 18행 전부를 해제시킨 커밋과 본문 줄을 행별로 확인했고, 사람이 close 선언 여부를 판정했다. 전수 대조에서 **12줄 픽스처 밖의 반례 1건**을 발견해 술어를 **좁혔다**.

**Evidence**

측정 좌표. 판정 브랜치는 `main`(`cachedMainBranch()`), 잰 트리는 이 워크트리다.

```
$ git rev-parse --show-toplevel
/Users/goos/MoAI/moai-adk-go/.claude/worktrees/t410
$ git branch --show-current
WT-drift-false-positive
$ git rev-parse HEAD
c7126f5269ebb88bb2673a7e875bda75f8ceba74     (M2 커밋; M3 코드 편집은 그 위의 워킹트리)
$ git rev-parse main
7ad9f8534dc48719854c67e2b9a06db97b594eaf
$ git rev-list --count main..develop
1443                                          (spec.md §6의 참조값 1,362를 이 시점에 재측정)
```

**수리 전 표를 이 에이전트가 재측정했다.** 오케스트레이터가 남긴 `drift-before.txt`는 다른 HEAD(`c323bb491`)에서 잰 것이고, 그것을 그대로 "이전"으로 쓰면 코드 델타와 코퍼스 델타가 섞인다. 그래서 두 소스 파일을 M1 커밋(`88b5ff686` — 수리 없음) 판본으로 되돌려 별도 바이너리를 만들고 **같은 코퍼스**에서 다시 쟀다:

```
$ git show 88b5ff686:internal/spec/drift.go       > /tmp/t410_drift.pre.go
$ git show 88b5ff686:internal/spec/drift_index.go > /tmp/t410_drift_index.pre.go
$ (워킹트리에 복사) ; go build -o /tmp/moai-t410-pre ./cmd/moai ; echo rc=$?
rc=0
$ /tmp/moai-t410-pre spec drift --no-cache > .moai/reports/t410/drift-before-remeasured.txt ; echo rc=$?
rc=0
$ (수리본 복원) ; go build -o /tmp/moai-t410 ./cmd/moai
$ /tmp/moai-t410 spec drift --no-cache > .moai/reports/t410/drift-after.txt ; echo rc=$?
rc=0
$ grep -c '^SPEC-' drift-before-remeasured.txt ; grep -c '^SPEC-' drift-after.txt
637
637
```

대조:

```
$ diff <(sort drift-before-remeasured.txt) <(sort drift-after.txt) ; echo rc=$?
(18개 행 변경 + 요약 1줄)
640c640
< Summary: 196/636 SPECs have status drift
---
> Summary: 178/636 SPECs have status drift
rc=1

$ grep '^> ' /tmp/t410_diff2.txt | grep -c 'DRIFT'
0                                  ← 비-DRIFT → DRIFT 회귀 0건 (AC-DCB-005 ①)
```

AC-DCB-004 — 확정 대상 행은 **존재하면서 DRIFT가 아니다**(행이 사라진 것이 아니다):

```
before: SPEC-V3R6-SESSION-HANDOFF-AUTO-001 completed  in-progress  DRIFT
after : SPEC-V3R6-SESSION-HANDOFF-AUTO-001 completed  completed    aligned
```

실측 반례가 해제되지 않았음:

```
$ grep 'SPEC-AUTONOMY-TIERS-001' drift-after.txt
SPEC-AUTONOMY-TIERS-001  completed  implemented  DRIFT      ← 여전히 DRIFT (plan.md §F M3-4)
```

### 행별 판정표 (AC-DCB-005 ②③)

18행 전부. `모양` A = `<full-ID>:` 줄 선두, B = squash 하위 subject. `판정`은 사람이 본문 줄을 읽고 내린 것이다.

| # | SPEC-ID | 전(git 함의) | 해제 커밋 | 커밋 subject | 모양 | 본문 줄 | 판정 |
|---|---|---|---|---|---|---|---|
| 1 | SPEC-ASTGREP-EDIT-001 | in-progress | `cd21df594` | `docs: close out 4 A-tier SPECs with sync-phase (… ASTGREP-EDIT-001) (#1215)` | A | `SPEC-ASTGREP-EDIT-001: status already completed; re-anchored the orphan` | **close 선언** — 4건을 닫는 커밋의 per-SPEC 기록. 이 행만 "이미 completed였고 고아 아티팩트를 재연결했다"고 적는다 |
| 2 | SPEC-CLI-TUI-MODERNIZE-001 | implemented | `cd21df594` | 위와 같음 | A | `SPEC-CLI-TUI-MODERNIZE-001: in-progress -> completed. sync_commit_sha` | **close 선언** — 상태 전이 명시 |
| 3 | SPEC-CLIFIX-LINTER-STALE-001 | implemented | `2f449e189` | `docs(specs): batch sync-phase close — 5 B-grade SPECs (3-phase close) (#1240)` | B | `docs(SPEC-CLIFIX-LINTER-STALE-001): sync-phase artifacts — 3-phase close` | **close 선언** — 규약 close subject 그대로 |
| 4 | SPEC-CONTEXT-ENGINE-RIGHTSIZE-001 | implemented | `2f449e189` | 위와 같음 | B | `docs(SPEC-CONTEXT-ENGINE-RIGHTSIZE-001): sync-phase artifacts — 3-phase close` | **close 선언** |
| 5 | SPEC-EPIC-STATUS-001 | in-progress | `6da952899` | `feat(factory+epic): Factory Mode multi-session bootstrap + epic status producer (#1448)` | B | `docs(SPEC-EPIC-STATUS-001): sync-phase artifacts + 3-phase close` | **close 선언** — subject에 close 신호가 없고 본문 줄이 규약 close다 |
| 6 | SPEC-FACTORY-BOOTSTRAP-001 | in-progress | `6da952899` | 위와 같음 | B | `docs(SPEC-FACTORY-BOOTSTRAP-001): sync-phase artifacts + 3-phase close` | **close 선언** |
| 7 | SPEC-GLM-KEY-INPUT-001 | implemented | `2f449e189` | 위와 같음 | B | `docs(SPEC-GLM-KEY-INPUT-001): sync-phase artifacts — 3-phase close` | **close 선언** (픽스처 3번과 동일 줄) |
| 8 | SPEC-I18N-GOVERNANCE-001 | implemented | `2f449e189` | 위와 같음 | B | `docs(SPEC-I18N-GOVERNANCE-001): sync-phase artifacts — 3-phase close` | **close 선언** |
| 9 | SPEC-INVOCATION-MODEL-002 | implemented | `cd21df594` | 위와 같음 | A | `SPEC-INVOCATION-MODEL-002: in-progress -> completed. sync_commit_sha` | **close 선언** |
| 10 | SPEC-KANBAN-RENAME-001 | implemented | `cd80f0644` | `chore(spec): close KANBAN-RENAME-001 + AGENT-MODEL-ENFORCE-001, supersede CONFIG-TIER-PERSIST-001 (#1516)` | A | `SPEC-KANBAN-RENAME-001: in-progress -> completed. Rename landed at 7f61332ef (PR #1513); …` | **close 선언** — 아래 별도 논의 |
| 11 | SPEC-TDD-ANTICHEAT-001 | implemented | `2f449e189` | 위와 같음 | B | `docs(SPEC-TDD-ANTICHEAT-001): sync-phase artifacts — 3-phase close` | **close 선언** |
| 12 | SPEC-V3R6-CI-FLAKY-STABILIZE-003 | in-progress | `7beda68a5` | `Close out 2 SPECs with 3-phase lifecycle completion (doc-only) (#1210)` | B | `chore(SPEC-V3R6-CI-FLAKY-STABILIZE-003): sync-phase artifacts — 3-phase close (e24067904)` | **close 선언** |
| 13 | SPEC-V3R6-CODERABBIT-ADOPTION-001 | in-progress | `7beda68a5` | 위와 같음 | B | `chore(SPEC-V3R6-CODERABBIT-ADOPTION-001): sync-phase artifacts — 3-phase close (4be491a0b)` | **close 선언** (픽스처 4번) |
| 14 | SPEC-V3R6-MAIN-RED-REMEDIATION-001 | in-progress | `e979a4d13` | `chore(SPEC group C): Mx-phase close (status implemented→completed, 2026-06-02)` | A | `SPEC-V3R6-MAIN-RED-REMEDIATION-001: Mx verdict SKIP-JUSTIFIED (0 production .go)` | **close 선언** — 이 카드의 확정 대상 커밋 |
| 15 | SPEC-V3R6-PROMPT-CACHE-001 | implemented | `e979a4d13` | 위와 같음 | A | `SPEC-V3R6-PROMPT-CACHE-001: Mx verdict EVALUATE-PASS, 10 @MX:ANCHOR tags` | **close 선언** (픽스처 2번) |
| 16 | **SPEC-V3R6-SESSION-HANDOFF-AUTO-001** | in-progress | `e979a4d13` | 위와 같음 | A | `SPEC-V3R6-SESSION-HANDOFF-AUTO-001: Mx verdict EVALUATE-PASS, 1 @MX:TODO deferred` | **close 선언** — **AC-DCB-004의 확정 행** (픽스처 1번) |
| 17 | SPEC-WORKFLOW-CACHE-OPT-001 | implemented | `cd21df594` | 위와 같음 | A | `SPEC-WORKFLOW-CACHE-OPT-001: in-progress -> completed. sync_commit_sha` | **close 선언** |
| 18 | SPEC-WORKTREE-ENTRY-STRATEGY-001 | implemented | `7beda68a5` | 위와 같음 | B | `docs(SPEC-WORKTREE-ENTRY-STRATEGY-001): sync-phase artifacts — 3-phase close (CHANGELOG + completed transition)` | **close 선언** (픽스처 12번 — 같은 본문 안 비-close 모양-B 줄들을 지나 도달) |

해제 커밋은 6개다: `cd21df594`(4행) · `2f449e189`(5행) · `7beda68a5`(3행) · `e979a4d13`(3행) · `6da952899`(2행) · `cd80f0644`(1행). AC-DCB-005의 공허 방지 조항(해제 0건이면 실패)은 18건으로 충족.

**10번 행에 대한 별도 논의 (판단이 갈릴 수 있는 유일한 행).** `cd80f0644`의 subject는 `chore(spec):` — `shouldSkipCommitTitle`이 metadata-sweep으로 **건너뛰는** 접두사다(AC-LSCSK-003). 이 축은 그 필터를 커밋 subject가 아니라 모양-B **본문 줄**에 적용하므로 이 커밋이 후보로 들어온다. 본문을 읽어 판정했다:

```
$ git show -s --format=%B cd80f0644 | head -10
chore(spec): close KANBAN-RENAME-001 + AGENT-MODEL-ENFORCE-001, supersede CONFIG-TIER-PERSIST-001 (#1516)
…
Lifecycle hygiene only — no source, template, or behaviour change.

SPEC-KANBAN-RENAME-001: in-progress -> completed. Rename landed at
7f61332ef (PR #1513); verified in-tree (…). 28/28 AC carried forward …
```

subject가 "close KANBAN-RENAME-001"이라고 명시하고 본문이 상태 전이를 적는다 — **진짜 close 커밋이며 metadata sweep이 아니다.** 판정: close 선언. 다만 `chore(spec):` scope가 skip 접두사와 겹친다는 사실은 남겨 둔다 — 훗날 후보 subject에 skip 필터를 적용하자는 제안이 오면 이 행이 그 대가다.

### 12줄 밖의 반례 1건 — 술어를 **좁혔다**

전수 대조 1차 실행에서 해제 행이 **19건**이었고, 그중 하나가 close가 아니었다.

```
$ git show -s --format=%B ac3e38a0b | sed -n '1,20p'
feat(SPEC-MCP-DEFAULT-ON-001): moai MCP server as first-class default (plan+amendment+run) (#1455)
…
* feat(SPEC-MOAI-MCP-SERVER-001, SPEC-TREND-MCP-001): in-place amendments for default-on gate inversion

Both completed SPECs amended per SPEC-MCP-DEFAULT-ON-001 REQ-A-6/REQ-A-7:
- SPEC-MOAI-MCP-SERVER-001: REQ-MCP-002 opt-in->default-on, REQ-MCP-015 opt-out flag, AC-MCP-002/006 amended (0.1.0 -> 0.2.0)
…
status: completed preserved (owner direction); amendment_of omitted because the
canonical transition pairs it with status in-progress, which the owner held at
completed.
```

이 줄은 **amendment 기록**이다 — REQ/AC가 개정되고 버전이 `0.1.0 -> 0.2.0`으로 올랐다는 서술이며, 커밋 본문이 스스로 "status: completed **preserved**"라고 적는다. 전이가 아니다. 모양 A(`<full-ID>:` + 임의 텍스트)에 걸린 것뿐이다.

`SPEC-V3R6-PROMPT-CACHE-001: Mx verdict EVALUATE-PASS…`(close)와 이 줄을 **줄 텍스트만으로** 가르는 구조적 성질은 없다. 키워드로 가르는 것은 §5.3이 판별자로 쓰지 말라고 못박은 방법이다. 그래서 §5.3이 정한 대로 **넓히지 않고 좁혔다** — 모양 A에 한해 **담고 있는 커밋의 subject가 close 신호를 갖도록** 요구한다(`subjectCloseSignal`, 대소문자 무시 `close` 부분문자열).

§5.4 조건 3(subject 쪽 close 신호를 쓰면 실측 close 커밋 전부를 통과시켜야 한다) 실측:

| subject | `close` 부분문자열 | `closeInfixMatch` |
|---|---|---|
| `Close out 2 SPECs with 3-phase lifecycle completion (doc-only) (#1210)` (7beda68a5) | 통과 | **탈락** |
| `chore(SPEC group C): Mx-phase close (…)` (e979a4d13) | 통과 | **탈락** |
| `docs(specs): batch sync-phase close — 5 B-grade SPECs (3-phase close) (#1240)` (2f449e189) | 통과 | 통과 |
| `docs: close out 4 A-tier SPECs with sync-phase (…) (#1215)` (cd21df594) | 통과 | 탈락 |
| `chore(spec): close KANBAN-RENAME-001 + … (#1516)` (cd80f0644) | 통과 | 탈락 |
| `docs(SPEC-INTERNAL-TEST-001): sync-phase artifacts + 3-phase close` | 통과 | 통과 |
| `feat(SPEC-HIERARCHICAL-TEAM-001): … (Tier M, 3-phase close) (#1394)` | 통과 | 통과 |

전부 통과한다 — `closeInfixMatch`가 떨어뜨리는 확정 대상 포함. 그것이 §5.4가 `closeInfixMatch` 재사용을 기각한 이유와 같은 근거이며, 그래서 게이트는 그 상수 집합이 아니라 넓은 `close` 부분문자열이다. 이 술어는 `transitions.go`의 close 규약 매처를 **건드리지 않으므로** REQ-DCB-006 위반이 아니다.

게이트는 **모양 A에만** 건다. 모양 B는 줄 자체가 검증된 체인에서 `completed`를 내야 하고, `6da952899`가 그 구분을 실측으로 못박는다 — 그 커밋 subject에는 close 신호가 없는데 본문 줄 둘은 진짜 close다(표 5·6행). 모양 B까지 게이트를 걸면 그 둘을 잃는다.

좁힘의 실측 효과 (좁힘은 해제 집합을 **줄일 뿐** 늘리지 못하므로 회귀는 원리상 불가):

```
좁히기 전: 해제 19행 (Summary 196 → 177)
좁힌 후  : 해제 18행 (Summary 196 → 178)
잃은 행  : SPEC-MOAI-MCP-SERVER-001 하나 — 위 amendment 오탐
$ grep 'SPEC-MOAI-MCP-SERVER-001' drift-after.txt
SPEC-MOAI-MCP-SERVER-001  completed  in-progress  DRIFT      ← 미해제로 남음
```

테스트 추가: `TestDriftCloseBody_AmendmentRecordIsNotClose`(반례 + 역방향 — 게이트가 모양 A를 **비활성화**하지 않고 **좁히는지**), `TestDriftCloseBody_SubjectCloseSignalCoversMeasuredCloses`(§5.4 조건 3 + `closeInfixMatch`가 실제로 둘을 떨어뜨린다는 기각 근거의 재측정).

### 뮤테이션 재실행 (좁힘 이후, 5종)

좁힘이 코드를 바꿨으므로 뮤턴트를 다시 돌렸다. 전부 죽는다.

| # | 뮤턴트 | rc | 죽인 입력 |
|---|---|---|---|
| 1 | 술어 전체 → `strings.Contains(body, specID)` | 1 | 5·6·7번 줄 · `_MentionIsNotClose` · `_OutputIsCompletedOrNothing` · `_AmendmentRecordIsNotClose` |
| 2 | 모양 A의 `:` 요구 제거 | 1 | 8번 · 10번 |
| 3 | 첫 모양-B 줄에서 반환 | 1 | `_FullBodyScan`(11·12 쌍) 단독 |
| 4 | 모양 B의 conventional-commit 전제 제거 | 1 | 7번 줄 |
| 5 | **신규** `subjectCloseSignal` → 항상 `true`(게이트 무력화) | 1 | `_AmendmentRecordIsNotClose` |

**Baseline-attribution** — 코퍼스 두 판(before-remeasured / after) 모두 이 트리(`.claude/worktrees/t410`), 브랜치 `WT-drift-false-positive`, HEAD `c7126f526` 위의 워킹트리에서, 판정 브랜치 `main` = `7ad9f8534`에 대해, **이 트리에서 빌드한** 바이너리(`/tmp/moai-t410-pre`, `/tmp/moai-t410`)로 `--no-cache`로 측정했다. 설치본은 쓰지 않았다(도구 출처 귀속, VCI §2.2). 행별 판정의 커밋·subject·본문 줄은 `git log main --no-merges --format=…` 한 번의 덤프(6.4MB)를 오프라인 분석해 얻었고, 각 커밋 본문은 `git show -s --format=%B <sha>`로 직접 확인했다.

**Gaps** — (a) 해제되지 **않은** 178행은 전수 조사하지 않았다. 이 카드의 대상은 "close가 main에 있는데도 못 보는" 부분집합이고 나머지 대부분은 spec.md §6이 적은 브랜치 격차 소음이지만, 그 178행 안에 이 술어가 놓치는 또 다른 body-declared close가 있는지는 **관측하지 않았다**. (b) `SPEC-MOAI-MCP-SERVER-001`이 실제로 main에서 닫혔는지(다른 커밋에) 확인하지 않았다 — 이 카드는 그 행을 미해제로 남겼을 뿐 그 SPEC의 frontmatter가 옳은지 판정하지 않는다. (c) 픽스처 밖 반례 탐색은 **해제된 19행에 대해서만** 수행했다. (d) `git rev-list --count main..develop` = 1443은 이 시점 참조값이며 이후 갱신된다.

**Residual-risk** — (1) 모양 A의 `subjectCloseSignal` 게이트는 `close`라는 **단어에** 의존한다. close 커밋이 그 단어 없이 작성되면 그 SPEC의 모양-A 선언은 안 보인다 — 의도된 보수적 방향(놓침 > 오탐)이지만 규약이 바뀌면 재측정해야 한다. (2) 10번 행(`chore(spec):` scope의 진짜 close)은 후보 subject에 skip 필터를 적용하는 순간 사라진다. 그 제안이 오면 이 행을 대가로 명시해야 한다. (3) (a)의 178행 미조사로 인해 **하한은 여전히 미확정**이다 — 18은 이 술어가 이 코퍼스에서 낸 값이지 "본문 선언 close의 총수"가 아니다.

---

## M4 — t382 HISTORY 정정

**Claim** — `SPEC-ERA-H3-NARROWING-001`의 원인 서술 정정이 HISTORY 1행 + `version:`/`updated:` 두 필드에 국한됐고, `status:`는 건드리지 않았다. 그 판정을 세 번 실행해(초록 → 붉음 → 되돌린 초록) 확인했다.

**Evidence**

편집 내용: HISTORY에 `0.5.1` 행 추가(원인 서술 ①이 t410 재현으로 반증됐다는 사실 + 근거 경로 + "판정·상태·AC는 유효" 명시), `version: "0.5.0"` → `"0.5.1"`, `updated: 2026-09-01` → `2026-09-03`.

세 실행. 술어는 spec.md v0.3.0이 D1 상환으로 교체한 `^[-+]status:` 형태다(중간 `^`가 없어 BSD/GNU 양쪽에서 같은 판정을 낸다).

```
# RUN 1 — 초록 (HISTORY 편집 후, 워킹트리 diff 비어 있지 않음)
$ git diff -U0 -- .moai/specs/SPEC-ERA-H3-NARROWING-001/spec.md | grep -cE '^[-+]status:'
0
$ git diff -U0 -- .moai/specs/SPEC-ERA-H3-NARROWING-001/spec.md | grep -cE '^[-+]'
7                                   ← 대조 대상이 실재한다 (공허 방지 ②: 빈 diff의 0을 인정하지 않는다)

# RUN 2 — 붉음 (status를 일시적으로 implemented로 바꿈)
$ git diff -U0 -- … | grep -cE '^[-+]status:'
2
$ git diff -U0 -- … | grep -E '^[-+]status:'
-status: completed
+status: implemented

# RUN 3 — 되돌린 초록
$ git diff -U0 -- … | grep -cE '^[-+]status:'
0
$ git diff -U0 -- … | grep -cE '^[-+]'
7
```

AC-DCB-007 세 항목:

1. `^[-+]status:` 카운트 **0** ✓ — `status: completed` 그대로.
2. 변경이 HISTORY 새 행 1개 + `version:`/`updated:`에 국한 ✓ — diff 헤더를 뺀 변경 줄 전체:

```
-version: "0.5.0"
+version: "0.5.1"
-updated: 2026-09-01
+updated: 2026-09-03
+| 0.5.1 | 2026-09-03 | manager-develop | **원인 서술 정정 …** |
```

3. 새 HISTORY 행이 근거 경로를 인용 ✓:

```
$ grep -c 'r1-walker-trace' .moai/specs/SPEC-ERA-H3-NARROWING-001/spec.md
1
```

**Baseline-attribution** — 이 트리(`.claude/worktrees/t410`), 브랜치 `WT-drift-false-positive`, HEAD `c7126f526` 위의 워킹트리. RUN 1~3은 전부 **커밋 이전** 워킹트리 diff에 대한 측정이므로 기준 SHA는 HEAD다. 판정 시점에 편집이 이미 커밋됐을 경우를 위한 대체 경로(공허 방지 ②의 `git diff -U0 <편집 직전 SHA>`)는 **필요하지 않았다** — 세 실행 모두 워킹트리가 더러운 상태에서 이뤄졌고 대조 카운트 7이 그 사실을 증명한다.

**Gaps** — (a) `moai spec lint`를 이 파일에 대해 다시 돌리지 않았다. 변경이 frontmatter 두 필드와 표 한 행이고 스키마 필드 집합·값 형식을 바꾸지 않으므로 새 finding이 날 경로가 없다고 판단했으나, **판단이지 측정이 아니다**. (b) `SPEC-ERA-H3-NARROWING-001`의 AC를 재실행하지 않았다 — spec.md §4가 그 SPEC의 재판정을 명시적으로 범위 밖에 두므로 의도적이다.

**Residual-risk** — HISTORY 행에 적은 반증 서술은 `.moai/reports/t410/r1-walker-trace.log`의 축자 추적에 근거한다. 그 추적은 조사 단계(plan 이전)에 만들어졌고 이 run-phase가 재생성하지 않았다 — 다만 §M3의 행별 판정표가 같은 기제를 독립적으로 재확인한다(모양 A/B가 실제로 어떤 커밋에서 오는지).

---

## 이월 부채 처리 — 판정서 D4 · D6

plan-audit 판정서(PASS-WITH-DEBT 0.80)가 run-phase 원장 처리로 이월한 두 건이다. 둘 다 **문서 층 결함**이고, 이 카드의 구현은 어느 쪽도 어기지 않는다.

### D4 — REQ-DCB-002의 FALLBACK-ONLY 조건이 한 갈래에서 도달 불가능하다

판정서 요지: REQ-DCB-002는 "1차 워크가 `completed`도 terminal도 내지 못한 동안" 조회한다고 적는데, `inMemImpliedStatus`가 **오류**를 내면 `DetectDrift`가 ① 게이트에 닿기 전에 `continue` 한다. 오류 반환도 그 서술을 문면상 만족하므로 요구가 계획된 배치보다 넓다.

**현재 코드에서 재확인** (`internal/spec/drift.go`, HEAD `c7126f526` 기준):

```go
gitStatus, err := inMemImpliedStatus(commits, a.specID)
if err != nil {
    continue                      // ← ① 블록에 도달하지 않는다
}
```

**판정: 문서 층 결함이며 구현은 정확하다. 이 카드에서 수리하지 않는다.**

- **효과는 무해하다.** 창이 통째로 분류 불가능한 SPEC은 record 자체가 안 나가므로 표에서 빠진다 — DRIFT 오탐을 만들지 않는다. 판정서가 severity를 minor로 매긴 이유다.
- **수리하려면 REQ-DCB-002 문면을 "상태를 **반환했고**"로 좁히고 오류 경로를 §4에 한 줄 넣어야 한다.** 그것은 `spec.md` **본문** 편집이고, manager-develop에게 금지된 표면이다(`spec-frontmatter-schema.md` § Forbidden ownership crossings — `status:`/`updated:` 외 본문 수정 금지). 규정된 경로는 blocker 보고 → manager-spec 재위임이다.
- **blocker로 올리지 않는 이유**: 결함이 구현을 막지 않는다. 구현은 이미 좁은 쪽(1차 워크가 상태를 반환한 경우에만 발화)으로 되어 있고, AC-DCB-003이 그 좁은 동작을 측정한다. 즉 **AC는 문면이 아니라 구현이 실제로 하는 일을 재고 있으며, 둘이 어긋나는 방향은 문면이 더 넓은 쪽**이다 — 안전한 방향이다.
- **남기는 것**: REQ-DCB-002의 문면이 오류 경로까지 주장한다는 사실은 이 원장에 기록된 상태로 남는다. sync-phase나 후속 카드가 `spec.md` 본문을 열 때 함께 좁히면 된다. 좁히지 않아도 구현은 바뀌지 않는다.

### D6 — Tier S 근거가 예측뿐이고, "범위를 줄인다" 규칙에 정의된 동작이 없다

판정서 요지: plan.md §A의 예상 4파일/~210 LOC가 계획 자신이 요구하는 산출물(원장·progress.md·SPEC 자신)을 빠뜨렸고, "넘으면 Tier를 올리지 말고 범위를 줄인다"는 초과가 범위 **안** 작업에서 나온 경우 수행 가능한 동작이 없다.

**실측** (plan.md §A가 요구한 M1 착수 시 재측정):

```
$ git diff --stat c323bb491..HEAD (M2까지) + 워킹트리
```

| 모집단 | 파일 | LOC |
|---|---|---|
| **소스 파일** (Go) | `internal/spec/drift.go` · `internal/spec/drift_index.go` · `internal/spec/drift_close_body_test.go` — **3** | 약 620 (테스트 ~470 포함) |
| SPEC 아티팩트 | `SPEC-DRIFT-CLOSE-BODY-001/spec.md`(frontmatter 1줄) · `progress.md` · `SPEC-ERA-H3-NARROWING-001/spec.md`(HISTORY 1행 + 2필드) — 3 | 소량 |
| 원장·측정 산출물 | `.moai/reports/t410/{run-evidence.md, drift-before.txt, drift-before-remeasured.txt, drift-after.txt}` — 4 | 대부분 기계 출력 |

**판정: 소스 파일 3개로 plan.md §A의 4파일 예측 **안**이며, Tier S 상한(`< 5 files`)을 넘지 않는다. Tier 재판정 blocker는 발동하지 않는다.**

- **모집단을 명시한다**(판정서의 Required fix 취지): 파일 수 판정의 모집단은 **소스 파일**이다. 원장·`progress.md`·SPEC 아티팩트는 어떤 AC가 요구하는 산출물이라 "줄일 수 있는 범위"가 아니고, 그것들을 세면 Tier 판정이 **자기 문서화의 양**에 좌우된다 — 측정하려던 것(구현의 크기)과 다른 것을 재게 된다.
- **LOC는 예측(~210)을 넘었다** — 테스트 파일이 ~470줄로 예측의 배 이상이다. 늘어난 원인은 전부 AC가 요구한 것이다: 12줄 실측 픽스처 표, 뮤테이션이 죽을 입력, 세 개의 공허 방지 선행 단언, 그리고 M3에서 발견한 반례 테스트 2개. **초과분이 범위 밖 작업에서 온 것이 아니므로 "범위를 줄인다"의 대상이 없다.**
- 그래서 판정서가 제안한 대체 동작("원장에 기록하고 Tier 재판정을 리드에게 blocker로 올린다")을 이 카드에서는 **기록까지만** 수행한다 — 파일 수가 상한 안이고 초과가 LOC 축에서만, 그것도 테스트 코드에서 발생했기 때문이다. Tier를 올릴 근거로 삼기에 약하다.
- **남기는 것**: plan.md §A의 파일 수 모집단 정의와 초과 시 동작은 여전히 문서에 없다. plan.md 본문 역시 manager-develop에게 금지된 표면이므로 이 원장이 그 기록이다.

---

## 카드 전제 재검토 (리드 지시)

리드 라운드 보고에 "카드 t410 전제의 절반이 거짓"이라는 기록이 있다. run-phase가 그 판단을 다시 밟았는지 명시한다.

**밟지 않았다 — 그 정정은 이미 plan 층에서 흡수됐고, run-phase의 측정이 그것을 재확인했다.** t382가 적은 두 원인 중 ①이 거짓이라는 것이 그 "절반"이며, spec.md §1.3이 그 반증을 이미 싣고 있다. run-phase가 새로 발견한 거짓 전제는 없다. 다만 **spec.md 자신의 서술 하나가 실측으로 불완전함이 드러났다**:

- §5.3의 모양 B 정의 "표지를 걷어낸 줄이 그 자체로 conventional-commit subject — 기존 필터 체인에 **그대로 먹이고**"는, 문자 그대로 구현하면 픽스처 7번(산문 + close-infix)을 `completed`로 낸다. "그 자체로 conventional-commit subject"가 기계적 전제로 서야 함을 M2 뮤턴트 4가 실측했다. **모순이 아니라 미명세**이며, 구현이 그 전제를 세워 §5.3의 기대(7번 = 언급)를 만족시킨다.
- §5.3 말미의 "술어는 후보이지 확정이 아니다"는 서술은 **실제로 발동했다** — M3에서 12줄 밖 반례 1건이 나왔고, 규정대로 좁혔다.

둘 다 spec.md 본문 수정이 필요한 사안이 아니다(전자는 구현이 이미 만족, 후자는 §5.3이 예견한 절차의 정상 작동).

---

## 절차 사고 기록 — 워크트리 index.lock 반복 경합 (리드 보고 대상)

이 트리의 `index.lock`이 run-phase 동안 **세 번** 커밋/스테이징을 막았다. 기록 의무에 따라 관측한 것만 적는다.

| 시점 | 명령 | 결과 |
|---|---|---|
| M1 스테이징 | `git add …` | `Unable to create … index.lock: File exists`. 직후 `ls` → **없음**(과도적) |
| M3 스테이징 (17:44~18:01) | `git add …` | 같은 실패. lock이 **17분간 존속** |
| M3 커밋 (18:02) | `git commit -F …` | 같은 실패. 직후 `ls` → **없음** |

17분 존속 건에 대한 staleness 근거 3종을 측정했다:

```
$ ls -l /Users/goos/MoAI/moai-adk-go/.git/worktrees/t410/index.lock
-rw-r--r--  1 goos  staff  0 Sep  3 17:44 …           ← 0바이트, mtime 고정 (18:01 시점 17분 경과)
$ ps -eo pid,etime,command | grep '[g]it '
(이 저장소를 대상으로 하는 git 프로세스 0건 — 목록은 전부 다른 저장소: mo.ai.kr, moai-cowork)
$ lsof /Users/goos/MoAI/moai-adk-go/.git/worktrees/t410/index.lock
(출력 없음 — 이 파일을 열고 있는 프로세스 없음)
```

세 근거가 일치해 **stale lock**으로 판정하고 제거했다(`rm -f`). 제거 대상은 **이 워크트리 전용** index lock이며 공유 브랜치 상태가 아니다. 제거 직후 `git status`가 정상 응답했고 스테이징이 통과했다.

그런데 **그 직후 커밋이 다시 같은 오류로 실패**했고, 재확인 시점(18:02:45)에는 lock이 또 없었다. 즉 실제 상황은 "죽은 lock 하나"가 아니라 **짧게 lock을 잡았다 놓는 행위자가 주기적으로 존재한다**는 것이다 — 17분 존속 건은 그 행위자가 남긴 잔재였을 가능성이 높다. 행위자는 특정하지 못했다(다른 Claude 세션의 statusline, SessionStart 훅, 또는 외부 도구).

재시도 판단: 실패한 명령이 **커밋**이고, 재시도 전에 `git rev-parse --short HEAD` = `c7126f526`(M2 그대로) · `git log -1 --format=%s` = M2 subject로 **효과가 착지하지 않았음을 확인**한 뒤 한 번 재시도해 통과했다. 부작용 있는 명령의 애매한 실패는 상태를 먼저 관측하고 효과 부재가 확인될 때만 재시도한다는 규율(`agent-common-protocol.md` § Error Recovery)에 따른 것이다.

**리드에게 보고할 사항**: 이 워크트리에 대해 정체 불명의 행위자가 간헐적으로 git index lock을 잡는다. 이 카드의 작업은 영향을 받지 않았으나(모든 커밋 착지 확인), 병합 창에서 같은 경합이 나면 병합 커밋 직전 `HEAD` 재판독 규율이 더 중요해진다.

---

## 리드 독립 재측정과의 7행 격차 — 요약줄이 아니라 grep이 과대매칭이다

리드가 `grep -c '^SPEC-.*DRIFT'`로 203 → 185를 얻었고, 원장의 요약줄은 196 → 178이다. 양쪽 다 정확히 **7행씩** 차이 나고 델타(18)는 같다. 원인을 측정했다.

```
$ grep -c '^SPEC-.*DRIFT' drift-before-remeasured.txt ; grep -c '^SPEC-.*DRIFT' drift-after.txt
203
185
$ grep -cE '^SPEC-.*[[:space:]]DRIFT[[:space:]]*$' drift-before-remeasured.txt ; (같은 것을 after에)
196
178
$ grep -c '^SPEC-[A-Z0-9-]*DRIFT' drift-after.txt
9
```

**`^SPEC-.*DRIFT`의 `.*DRIFT`가 Drift? 열이 아니라 SPEC-ID 안의 `DRIFT` 문자열을 문다.** ID에 `DRIFT`를 담은 행이 9건이고, 그중 2건(`SPEC-UPDATE-DOC-DRIFT-001` · `SPEC-V3R6-TEMPLATE-MIRROR-DRIFT-001`)은 실제로 DRIFT라 양쪽이 함께 센다. 나머지 **7건은 `aligned`인데 ID 때문에 매치**된다:

```
SPEC-DRIFT-001                        implemented  era-exempt  aligned
SPEC-V3R4-STATUS-DRIFT-FOLLOWUP-001   completed    era-exempt  aligned
SPEC-V3R4-STATUS-DRIFT-FOLLOWUP-002   completed    era-exempt  aligned
SPEC-V3R6-CI-BASELINE-DRIFT-001       implemented  era-exempt  aligned
SPEC-V3R6-DOCS-USER-DRIFT-001         implemented  era-exempt  aligned
SPEC-V3R6-DRIFT-CONVENTION-ALIGN-001  completed    completed   aligned
SPEC-V3R6-DRIFT-LEGACY-CONVENTION-001 completed    completed   aligned
```

196 + 7 = 203, 178 + 7 = 185. **요약줄이 제외하는 행 부류는 없다** — 요약줄의 `report.Count`는 `Drifted == true`인 record 수이고(`internal/cli/spec_drift.go:132`), 열 위치를 고정한 `grep -cE '^SPEC-.*[[:space:]]DRIFT[[:space:]]*$'`가 그 값을 정확히 재현한다. 두 수 중 요약줄이 옳고 `^SPEC-.*DRIFT`가 7만큼 과대매칭한다.

이것이 이 카드가 다룬 결함과 **같은 형태**라는 점은 기록해 둘 만하다: 토큰(`DRIFT`)이 두 부류의 줄에 공통으로 나타나므로 키워드만으로는 가를 수 없고, 열 위치라는 구조를 요구해야 갈린다 — spec.md §5.3이 `(completed)`를 판별자로 쓰지 말라고 적은 것과 같은 이유다.

**동시에 발견한 이 원장 자신의 오차 1건 (정정).** §C 사전 점검의 `grep -c '^SPEC-'` = 637은 **표 머리글 `SPEC-ID …` 한 줄을 포함한 값**이다. 실제 데이터 행은 636이며 요약줄의 분모 `/636`과 일치한다. 이 값은 "0이면 측정이 잘못된 것"이라는 온전성 검사로만 쓰였고 어떤 판정에도 들어가지 않았으나, 수를 적은 이상 정확해야 하므로 정정한다: **표의 SPEC 행 수는 636이다.**

