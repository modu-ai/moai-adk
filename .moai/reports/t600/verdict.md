# t600 — internal/cli 선재 적색 4건: 원인 기록 (run 진행 중)

- card: t600 (Class B, Tier M)
- worktree: `.claude/worktrees/t600`, branch `WT-cli-prered-four`
- base: 로컬 develop = origin/develop = `81c1d58f9`
- 상태: run 완료 — 원인 확립, 운영자 결정 A 로 수리, 병합 트리 `0a16d2bc1` 에서 대조군·vet·lint 통과. 수리 커밋 `5b7927b15`, develop `ee99507fb` 흡수 `0a16d2bc1`. 미푸시. develop 병합은 리드 창 지명 대기.

## 1. 재현 (HEAD `81c1d58f9`, 수정 전 트리)

리드 슬롯 승인 1회. 실행 직전 `git status --short --untracked-files=no` 출력 없음, `git rev-parse --short HEAD` = `81c1d58f9`.

```
go test ./internal/cli/ -count=1 -run '^(TestHomeStateChangedSurfaceCoverageConsumesFreshProfile|TestChangedProductionFilesDerivesCurrentHeadDiffAndPlatformDisposition|TestHomeStateChangedSurfaceCoverageRunsBoundedFocusedSuite|TestAuditLagUsesBinlagSeam)$' -v
exit=1   (전체 출력: run1-head-81c1d58f9.txt)

--- FAIL: TestHomeStateChangedSurfaceCoverageConsumesFreshProfile (4.91s)
--- FAIL: TestHomeStateChangedSurfaceCoverageRunsBoundedFocusedSuite (187.10s)
--- FAIL: TestChangedProductionFilesDerivesCurrentHeadDiffAndPlatformDisposition (3.03s)
--- FAIL: TestAuditLagUsesBinlagSeam (1.04s)
    home_state_coverage_test.go:420: audited production file changed after coverage tip: internal/cli/launcher.go
    home_state_coverage_test.go:468: coverage={Covered:0 Total:0 Percent:0} err=coverage cli tests: exit status 1
            home_state_coverage_test.go:420: audited production file changed after coverage tip: internal/hook/session_end.go
            home_state_coverage_test.go:520: audited production file changed after coverage tip: internal/cli/launcher.go
    home_state_coverage_test.go:520: audited production file changed after coverage tip: internal/hook/session_end.go
    mcp_build_identity_test.go:643: sweep: baseline hit home_state_coverage.go:243 MISSING — ...
    mcp_build_identity_test.go:643: sweep: baseline hit home_state_coverage.go:251 MISSING — ...
    mcp_build_identity_test.go:648: sweep: NEW ancestry hit home_state_coverage.go:253 — ... (REQ-ABI-006 violation)
    mcp_build_identity_test.go:648: sweep: NEW ancestry hit home_state_coverage.go:245 — ... (REQ-ABI-006 violation)
```

선택자 공허 통과 방지 계수: 네 테스트 각각 top-level `--- FAIL` 1줄, `--- PASS` 0줄.

## 2. 원인 A — TestAuditLagUsesBinlagSeam: 줄 좌표 표류

이 테스트는 `internal/cli` 비테스트 소스에서 `merge-base`/`is-ancestor` 가 나오는 **줄 번호** 집합을 고정 목록과 정확히 비교한다.

| 커밋 | `home_state_coverage.go` 의 `is-ancestor` 줄 | 기대 목록 |
|---|---|---|
| `d8e9896e5` (PR #1699 병합) | 243, 251 | home_state 항목 없음 → 적색 |
| `1cfc6f544` (13:03) | 243, 251 | 243/251 추가 → 초록 |
| `64a68b118` (13:20) | **245, 253** | 243/251 그대로 → 적색 시작 |

```
$ git show 1cfc6f544:internal/cli/home_state_coverage.go | grep -n "is-ancestor"
243: ... "merge-base", "--is-ancestor", previous, commit ...
251: ... "merge-base", "--is-ancestor", result.Tip, "HEAD" ...
$ git show 64a68b118:internal/cli/home_state_coverage.go | grep -n "is-ancestor"
245: ...
253: ...
```

`64a68b118` 은 상수 `homeStateCoverageCertificationSubject` 한 줄과 `subjects` 목록 항목 한 줄을 호출부 위에 추가했다(`git show 64a68b118 -- internal/cli/home_state_coverage.go`). 호출 두 곳의 내용은 그대로이고 새 비교 구현은 없다. 테스트가 말하는 "REQ-ABI-006 위반"은 좌표 기반 검사가 이동을 추가+제거로 읽은 것이며, 실제 위반이 아니다.

## 3. 원인 B — 커버리지 3건: 감사 대상 blob 을 HEAD 에 동결

`resolveHomeStateCoverageChangeSet` (`internal/cli/home_state_coverage.go:196-362`) 은 t592 증거 marker 커밋(최대 4개, subject 로 식별)이 바꾼 production 파일마다 marker 시점 blob 을 기억하고, `HEAD` 의 blob 과 하나라도 다르면 실패한다(`:295-299`). 테스트는 저장소 자신(`../..`)을 입력으로 쓰므로, t592 체인 이후 develop 에 이 파일들을 건드리는 커밋이 들어오는 순간 구조적으로 적색이 된다.

### 언제부터 — 커밋별 blob 비교 모델

`:259-299` 를 git 기록으로 옮긴 모델(입력: 각 marker 의 `git diff --name-status -M <m>^1 <m> -- '*.go'`, 각 리비전의 `git ls-tree -r <rev> -- internal`; 보존: `.moai/state/verify/t600/model/`). 대조군: 4-marker 체인을 인증 marker `743c06225` 자신에 대면 불일치 0.

| 리비전 (develop first-parent) | marker 수 | 감사 파일 | 불일치 |
|---|---|---|---|
| `d8e9896e5` PR #1699 병합 | 3 | 28 | 0 |
| `1cfc6f544` | 3 | 28 | 10 (일시 적색) |
| `8418f3021` 인증 marker `743c06225` 병합 | 4 | 29 | 0 |
| `fe154282f` | 4 | 29 | 0 |
| `d060e0d13` t593 병합 (`launcher.go`) | 4 | 29 | **1 ← 현재 적색 시작** |
| `bdbc09788` t556 (`mcp_server.go`) | 4 | 29 | 2 |
| `81c1d58f9` 현재 develop | 4 | 29 | 4: `internal/cli/launcher.go`, `internal/cli/mcp_server.go`, `internal/hook/session_end.go`, `internal/kanban/todo_root.go` |

`git diff --stat 64a68b118 743c06225` 는 출력이 없다 — 인증 marker 는 `64a68b118` 과 같은 트리를 새 커밋으로 박아 `1cfc6f544` 의 일시 적색을 되돌린 것이다. 즉 이 설계는 감사 대상 29개 중 하나가 바뀔 때마다 재인증 커밋을 요구한다.

실측과의 대조: 모델이 예측한 HEAD 불일치 4개 중 `launcher.go` 와 `session_end.go` 가 한 번의 실행 안에서 서로 다른 지목으로 나타났다(§1). 지목 파일이 map 순회 순서로 흔들리는 성질은 t606 의 대상이다.

### 판정 전 두 층

- `TestChangedProductionFiles…` / `…ConsumesFreshProfile`: `resolveHomeStateCoverageChangeSet` 가 곧바로 불일치를 반환.
- `…RunsBoundedFocusedSuite`: `measureChangedSurfaceCoverageResultWith` (`:85-110`) 가 먼저 `runChangedSurfaceCoverageSuite` 로 자식 `go test` 5회를 돌리는데, 자식 패턴이 `TestHomeState.*`·`TestChangedProduction.*` 를 포함해 위 두 테스트가 자식 안에서도 실패 → `coverage cli tests: exit status 1`. 자식이 통과했더라도 `:100` 의 resolve 에서 같은 불일치로 실패한다.

## 4. production 경로 — 코드 판독

같은 검사가 `moai migrate home-state --apply --verified-live` 의 사전 검증에 걸려 있다.

```
internal/cli/migrate_home_state.go:571-578   if r.verifiedLive { verifier = validateLivePreApply ... authorization, err = verifier(ctx, root) }
internal/cli/migrate_home_state.go:156-157   validateLivePreApply → validateLivePreApplyWith(..., measureChangedSurfaceCoverage)
internal/cli/migrate_home_state.go:160-~211  HEAD/빌드 일치 → named tests → race → vet → strict spec lint → windows compile → coverage(ctx, root)
internal/cli/home_state_coverage.go:36-37     measureChangedSurfaceCoverage → measureChangedSurfaceCoverageResultWith(..., runChangedSurfaceCoverageSuite)
internal/cli/home_state_coverage.go:96-100    run(...) 뒤 resolveHomeStateCoverageChangeSet(root)
internal/cli/home_state_coverage.go:295-299   "audited production file changed after coverage tip"
```

`SPEC-HOME-STATE-ROLLOUT-001/progress.md` 기준 live apply 는 미실행(`AC-HSR-022 | PENDING-LIVE`, `run_status: post-commit-remediation-re-audit-pending-live-rollout-pending`).

## 5. 수리 (운영자 결정 A)

운영자 결정(리드 전달): production 게이트 의미는 그대로 두고, 라이브 저장소 이력을 읽는 테스트만 고친다. `resolveHomeStateCoverageChangeSet` 과 `migrate_home_state.go` 경로는 수정하지 않았다. 재인증은 이 카드의 범위가 아니다.

| 파일 | 변경 |
|---|---|
| `internal/cli/mcp_build_identity_test.go` | 원인 A — 이 파일이 빨간 넷 중 하나인 `TestAuditLagUsesBinlagSeam` 의 소재지다. 기대 좌표 `home_state_coverage.go:243/251` 과 주석 속 좌표를 245/253 으로 갱신(`git diff --numstat` = `5 5`, 좌표 두 줄 + 주석 세 줄). 커버리지 감사 대상과의 관계: 이 파일은 인증 marker `743c06225` 의 name-status 에 `M internal/cli/mcp_build_identity_test.go` 로 나오지만, `isProductionCoverageFile`(`home_state_coverage.go:400-403`)이 `_test.go` 를 제외하므로 blob 동결 대상(`expectedBlobs`)에 들어가지 않는다 — 수정해도 게이트를 건드리지 않는다 |
| `internal/cli/home_state_coverage_test.go` | `…ConsumesFreshProfile`, `…ChangedProductionFilesDerives…` 를 fixture 저장소 기반으로 전환. `…RunsBoundedFocusedSuite` 는 opt-in 뒤로 이동 |
| `internal/config/envkeys.go` | `EnvTestHomeStateLiveCoverage = "MOAI_TEST_HOME_STATE_LIVE_COVERAGE"` 추가(테스트 전용 블록) |

### 테스트별 내용과 기존 fixture 테스트와의 겹침

- **`TestHomeStateChangedSurfaceCoverageConsumesFreshProfile`** — `committedCoverageRepo` 위에서 측정 경로(`measureChangedSurfaceCoverageWith` → `changedProductionLineRanges` → `parseChangedLineCoverage`)가 100% 를 내는지 확인하고, 이어서 감사 파일을 인증 뒤 커밋으로 바꾸면 같은 측정 경로가 `audited production file changed after coverage tip: internal/x/a.go` 로 거부하는지 단정한다. 나머지 오류 경로 단정(러너 오류, 없는 저장소, 임시 저장소 불가)은 그대로다.
  - 겹침: 정상 경로의 change-set 해석은 `TestCommittedCoverageChangeSetWorksInCleanRepositoryAndAfterUnrelatedCommit` 과, 거부 경로는 `TestCommittedCoverageChangeSetRejectsStaleOrAmbiguousEvidence/covered_path_changed_after_audited_tip` 과 같은 fixture·같은 변경을 쓴다. 이 테스트만의 몫은 live pre-apply 가 실제로 부르는 **측정 계층**을 통과한다는 점과, 거부 메시지·지목 파일까지 단정한다는 점이다(기존 서브테스트는 `err != nil` 만 본다).
- **`TestChangedProductionFilesDerivesCurrentHeadDiffAndPlatformDisposition`** — 인증 marker 커밋 하나가 `internal/y/b.go`, 플랫폼 외 파일(`only_windows.go`, windows 에서는 `only_unix.go`), `internal/x/a_test.go`, `cmd/tool/main.go` 를 추가할 때 native 목록과 disposition 목록을 **정확히** 비교한다.
  - 겹침: cross-compile 분류는 `TestChangedProductionLineRangesTracksModifiedRenamedAndUntracked` 가 미추적 파일 경로로, `_test.go` 제외는 `TestCommittedCoverageChangeSetRejectsInvalidGitAndMalformedEvidence` 가 `productionFilesFromNameStatus` 단위로 이미 본다. 이 테스트만의 몫은 커밋된 marker 경로, 정렬된 목록 전체의 동등 비교, `internal/` 밖 파일 제외다.
- **`TestHomeStateChangedSurfaceCoverageRunsBoundedFocusedSuite`** — 제거하지 않고 `MOAI_TEST_HOME_STATE_LIVE_COVERAGE=1` 일 때만 돌도록 옮겼다. 이 테스트는 실제 저장소에서 자식 `go test` 5회를 돌려 진짜 커버리지를 재는 유일한 경로라서 fixture 로 대체할 수 없고, 재인증 시점에 사람이 켜서 쓸 가치가 남는다. 켜면 현재 develop 에서는 원인 B 때문에 실패하는 것이 정상이다.

## 6. 검증

### 6.1 뮤턴트 — blob 동결 비교 무력화 (리드 슬롯 승인)

`internal/cli/home_state_coverage.go:297` 을 `if false && (err != nil || headBlob != expected) {` 로 잠시 바꾼 뒤 실행했다.

```
go test ./internal/cli/ -count=1 -run '^(TestCommittedCoverageChangeSetRejectsStaleOrAmbiguousEvidence|TestHomeStateChangedSurfaceCoverageConsumesFreshProfile)$' -v
exit=1   (run2-mutant-blobfreeze-off.txt)

TestCommittedCoverageChangeSetRejectsStaleOrAmbiguousEvidence FAIL=1 PASS=0
TestHomeStateChangedSurfaceCoverageConsumesFreshProfile FAIL=1 PASS=0
    home_state_coverage_test.go:217: stale audited production blob accepted
    --- FAIL: TestCommittedCoverageChangeSetRejectsStaleOrAmbiguousEvidence/covered_path_changed_after_audited_tip (1.05s)
    home_state_coverage_test.go:458: post-tip audited change accepted: <nil>
```

원복은 `git restore -- internal/cli/home_state_coverage.go` 하나만 썼다. 원복 직후 `sed -n 297p` = `if err != nil || headBlob != expected {`. 원복 뒤 `git diff --stat`(after-mutant-restore-diffstat.txt):

```
 internal/cli/home_state_coverage_test.go | 66 +++++++++++++++++++++++++++-----
 internal/cli/mcp_build_identity_test.go  | 10 ++---
 internal/config/envkeys.go               |  7 ++++
 3 files changed, 68 insertions(+), 15 deletions(-)
```

### 6.2 수리한 넷 (리드 슬롯 승인)

```
go test ./internal/cli/ -count=1 -run '^(TestHomeStateChangedSurfaceCoverageConsumesFreshProfile|TestChangedProductionFilesDerivesCurrentHeadDiffAndPlatformDisposition|TestHomeStateChangedSurfaceCoverageRunsBoundedFocusedSuite|TestAuditLagUsesBinlagSeam)$' -v
exit=0   (run3-repaired-four.txt)

TestHomeStateChangedSurfaceCoverageConsumesFreshProfile PASS=1 FAIL=0 SKIP=0
TestChangedProductionFilesDerivesCurrentHeadDiffAndPlatformDisposition PASS=1 FAIL=0 SKIP=0
TestHomeStateChangedSurfaceCoverageRunsBoundedFocusedSuite PASS=0 FAIL=0 SKIP=1
TestAuditLagUsesBinlagSeam PASS=1 FAIL=0 SKIP=0
    home_state_coverage_test.go:485: set MOAI_TEST_HOME_STATE_LIVE_COVERAGE=1 to measure changed-surface coverage against the live repository
ok  	github.com/modu-ai/moai-adk/internal/cli	5.590s
```

### 6.3 커밋과 develop 흡수

- 수리 커밋 `5b7927b15` (경로 명시 3파일: `internal/cli/home_state_coverage_test.go`, `internal/cli/mcp_build_identity_test.go`, `internal/config/envkeys.go`).
- 로컬 develop `ee99507fb` 흡수 → 병합 커밋 `0a16d2bc1` (부모 `5b7927b15`, `ee99507fb`). 충돌 없음. `git diff --stat 81c1d58f9 ee99507fb -- <카드 파일 4개 + migrate_home_state.go>` 출력 없음. 병합 뒤에도 `is-ancestor` 는 245/253.

### 6.4 대조군 — 병합 트리 `0a16d2bc1` (리드 슬롯 승인)

흡수 전 수리 트리에서 한 번 돌렸던 대조군(run4)은 `internal/cli` 루트가 `panic: test timed out after 10m0s` 로 끊겼다. 그 시점 `--- FAIL` 0줄, `--- PASS` 2297줄, 실행 중이던 테스트는 `TestHomeStateChangedSurfaceCoverageConsumesFreshProfile (0s)` — 테스트 실패가 아니라 패키지 전체 실행 시간이 10분을 넘은 것이다. 하위 16패키지는 그때도 모두 ok. 그래서 병합 트리에서는 `-timeout 1200s` 로 다시 쟀다.

```
go test ./internal/cli/... -count=1 -timeout 1200s -v   (요약 run5-cli-all-merged.summary.txt, 전체 로그 .moai/state/verify/t600/run5-cli-all-merged.txt)
cli-exit=0
ok  	github.com/modu-ai/moai-adk/internal/cli	975.040s
ok 줄 17개 = internal/cli 루트 1 + 하위 16 (agentlint, harness, pr, preference, printer, specid,
             taskledger, uikit, update, update/backup, update/deploy, update/merge, update/plan,
             update/report, wizard, worktree)
top-level: --- FAIL 0 / --- PASS 4475 / --- SKIP 27

TestHomeStateChangedSurfaceCoverageConsumesFreshProfile PASS=1 FAIL=0 SKIP=0
TestChangedProductionFilesDerivesCurrentHeadDiffAndPlatformDisposition PASS=1 FAIL=0 SKIP=0
TestHomeStateChangedSurfaceCoverageRunsBoundedFocusedSuite PASS=0 FAIL=0 SKIP=1
TestAuditLagUsesBinlagSeam PASS=1 FAIL=0 SKIP=0

go test ./internal/config/ -count=1   (run5-config.txt)
config-exit=0
ok  	github.com/modu-ai/moai-adk/internal/config	2.592s
```

리드가 알린 `ee99507fb` CI 의 새 실패 7개(미러 드리프트 계열)는 이 범위(`internal/cli/...`, `internal/config`)에 나타나지 않았다. FAIL 0 이므로 "develop 기원, 카드 무관"으로 따로 분리할 항목이 없다.

### 6.5 vet / lint — 병합 트리

```
go vet ./internal/cli/ ./internal/config/          vet-exit=0   (run6-vet.txt, 출력 없음)
golangci-lint run ./internal/cli/ ./internal/config/   lint-exit=0   0 issues.   (run6-lint.txt)
```

## 7. 동작상 부수 효과

live apply 의 자식 스위트 패턴(`TestHomeState.*`, `TestChangedProduction.*`)에 위 두 테스트가 들어 있다. 이제 둘은 fixture 기반이라 자식 실행에서 먼저 실패하지 않고, 거부는 `home_state_coverage.go:100` 의 resolve 한 곳에서 일어난다. 게이트의 판정 자체는 바뀌지 않는다.

## Gaps

- live apply 경로가 막힌다는 판단은 **코드 판독**이다. `moai migrate home-state --apply --verified-live` 는 실제 홈 상태를 바꾸므로 실행하지 않았다. 게다가 그 경로에서는 coverage 앞의 단계(빌드 커밋 일치, named tests 등)가 먼저 실패할 수도 있어, coverage 가 첫 거부 사유인지는 관측하지 않았다.
- 커밋별 적색 시점은 git 기록 모델이며, `d060e0d13` 등 과거 리비전에서 테스트를 실제로 돌리지는 않았다. HEAD 한 점만 실측과 대조했다.
- `1cfc6f544` 시점에서 binlag 테스트가 실제로 초록이었는지는 기대 목록과 소스 줄의 대조로만 판단했다(나머지 좌표 4개는 HEAD 에서만 확인).
- `internal/cli/...` 전체와 `internal/config` 대조군, vet·lint 는 아직 돌리지 않았다(§6.3).
- 뮤턴트는 blob 비교 무력화 한 가지만 확인했다. 새 disposition 테스트의 분류 규칙 변이는 따로 돌리지 않았다.
- opt-in 테스트를 켠(`MOAI_TEST_HOME_STATE_LIVE_COVERAGE=1`) 상태는 실행하지 않았다. 켜면 현재 develop 에서 실패하는 것이 설계상 기대값이지만 관측하지 않았다.
- darwin 에서만 실행했다. windows 분기(`only_unix.go`)는 CI 매트릭스가 판정한다.
- 병합 트리 대조군 시작 직전 `ps` 에서 go test/vet/build·golangci 패턴에 2건이 걸렸다(run5-uptime-before.txt 둘째 줄). 어느 프로세스인지 기록하지 않아 귀속하지 못했다. 결과는 FAIL 0 이지만 실행 시간(루트 975s)은 이 부하의 영향을 받았을 수 있다.
- 전체 `-v` 로그(run4 676K, run5 1.1M)는 커밋하지 않고 `.moai/state/verify/t600/` 에 보존했다(git 추적 밖). 커밋본은 `ok`/`FAIL`/`--- FAIL`/`--- SKIP`/panic 줄과 카드 테스트 PASS 줄만 뽑은 `*.summary.txt` 다.

## Residual-risk

- binlag 검사는 줄 좌표 기반이다. 이번 수리는 `home_state_coverage.go` 를 바꾸지 않아 245/253 이 유지되지만, 앞으로 그 파일 245행 위를 고치는 변경은 다시 이 테스트를 빨갛게 만든다(설계상의 성질, 이 카드 범위 밖).
- opt-in 뒤로 옮긴 테스트는 기본 실행에서 돌지 않으므로, 재인증 시점에 누군가 켜지 않으면 실제 저장소 커버리지 측정은 관측되지 않은 채 남는다. live apply 자체는 여전히 같은 측정을 실행한다.
