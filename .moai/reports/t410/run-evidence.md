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

