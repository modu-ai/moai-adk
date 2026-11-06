auditor-model: claude-opus-5-5
verdict: PASS
score: 0.94
audited_sha: 16e8071f4fa59f5aa33b194c8426882f85ddbb32

## Evaluation Report
SPEC: 없음 (카드 t2107, `OwnershipTransitionRule` squash 본문 오판 수리)
Overall Verdict: PASS (0.94). 차단 결함 0건, 선택 결함 3건.

감사 대상: 트리 `/Users/goos/.moai/worktrees/projects/moai-adk-go/t2107`, 브랜치 `WT-spec-lint-ownershiptrans`, HEAD `16e8071f4fa59f5aa33b194c8426882f85ddbb32`, merge-base(main) `b25adf4fca05e6dba07abbdf307ccaf2b80b9033`. 감사 시작과 종료 시점에 HEAD가 같고 작업 트리가 깨끗함을 확인했다(`git rev-parse HEAD`, `git status --short` 출력 없음).

### Dimension Scores
| Dimension | Score | Verdict | Evidence |
|-----------|-------|---------|----------|
| Functionality (40%) | 96/100 | PASS | 신규 5개 테스트 `--- PASS`, 대상 묶음 50 PASS / 0 FAIL / 0 SKIP, `ok ... 38.970s`, 패키지 전체 `ok ... coverage: 91.0% of statements`(E2). 변이 2종 모두 사망(E3). 트리에서 빌드한 바이너리로 실제 이력 검사 시 `INFO OwnershipTransitionUnmeasured ... commit 55b20438fd... carries Authored-By-Agent trailers of several actors (manager-spec, manager-develop, manager-docs)`, `0 error(s), 0 warning(s)`(E4) |
| Security (25%) | 100/100 | PASS | 새 입력 경로 없음. 정규식은 기존 `authoredByAgentLine` 그대로이며 `FindAllStringSubmatch`로 commit 본문만 읽는다. 셸 호출·파일 쓰기 추가 없음(E1) |
| Craft (20%) | 86/100 | PASS | `golangci-lint run --new-from-rev=b25adf4fca... ./internal/spec/` → `0 issues.`, `gofmt -l` 출력 없음, `go vet ./internal/spec` exit 0(E5). `Check` 문서 주석 미갱신(F2), 프로덕션에서 쓰이지 않게 된 `parseAuthoredByAgent`(F3)로 감점 |
| Consistency (15%) | 93/100 | PASS | 두 조회 경로(per-SPEC, batch)가 같은 헬퍼 `applyTrailerAttribution`을 쓴다(E1). 문서 로컬·템플릿 사본이 같은 한 줄로 바뀌었고 기존 차이(4줄)는 그대로(E6) |

Must-pass 방화벽(Functionality·Security): 둘 다 PASS.

### Findings (structured defect-list)
- F1 [Low, 확신 높음] [optional] [in-diff] `internal/spec/lint_ownership.go:309-318` — 서로 다른 행위자 수를 셀 때 전환 행렬 밖의 행위자(`manager-git` 등)도 센다. 그래서 squash가 아닌 commit이라도 `manager-develop` + `manager-git` 트레일러를 함께 달면, 예전엔 조용히 통과하던 전환이 이제 Info `OwnershipTransitionUnmeasured`가 된다. overlay 전용 탐침 테스트로 관측했다(E7). Info라 종료 코드·baseline 게이트에는 영향이 없고, 문서 문장("trailers of more than one distinct actor")과도 맞는다. 필요한 수정: 없음(의도된 보수적 동작). 원한다면 `trailerAgentOwnerKind(a) != ownerNone`인 행위자만 세도록 거를 수 있으나, 그 경우 "squash 본문은 귀속 불가"라는 판단 근거가 흐려지므로 권하지 않는다.
- F2 [Low, 확신 높음] [optional] [touched-file] `internal/spec/lint_ownership.go:407-422` — `Check`의 `Behavior:` 목록에 "다중 행위자 본문 → Info `OwnershipTransitionUnmeasured`" 항목이 없다(분기 안 주석에는 있음). 필요한 수정: 목록에 한 줄 추가.
- F3 [Info, 확신 높음] [optional] [touched-file] `internal/spec/lint_ownership.go:281` — `parseAuthoredByAgent`는 이제 테스트(`lint_ownership_test.go:457,478`)에서만 호출된다. 삭제는 범위 밖이며 동작 결함도 아니다. 필요한 수정: 없음(후속 정리 후보로만 기록).

### Recommendations
- t2105가 착지한 뒤 SPEC-FACTORY-COMPLETION-RECOVERY-001의 `lint.skip: [OwnershipTransitionInvalid]` 제거를 manager-spec 재위임(비전환 frontmatter 수정)으로 처리할 것. 이 트리에서 skip 없이 Info만 나오는 것은 E4로 확인했다.

---

## Claim
1. 다중 행위자 squash 본문의 전환은 `OwnershipTransitionInvalid`가 아니라 Info `OwnershipTransitionUnmeasured`로 보고된다.
2. 같은 행위자가 반복된 본문, 트레일러가 하나인 본문, 트레일러가 없는 본문의 판정은 바뀌지 않는다.
3. per-SPEC 조회와 batch 조회가 같은 결과를 낸다(두 경로 모두 헬퍼 사용).
4. 실제 이력에서 commit `55b20438fd`(#1930) 판정이 Info Unmeasured이고 경고가 없다.
5. 문서 SSOT 문장과 템플릿 사본이 동일하게 갱신됐다.
6. 레인이 선언한 "말뭉치 전체 lint 생략" 근거(이 변경은 Warning을 Info로 낮출 수만 있고 Warning을 새로 만들지 않는다)는 타당하다.

## Evidence

### E1 diff 검토
`git diff --stat b25adf4fca05e6dba07abbdf307ccaf2b80b9033..HEAD`:
```
 .../moai/development/spec-frontmatter-schema.md    |   2 +-
 .moai/reports/t2107/progress.md                    |  40 ++++++++
 internal/spec/lint_ownership.go                    |  72 ++++++++++++--
 internal/spec/lint_ownership_batch.go              |  10 +-
 internal/spec/lint_ownership_squash_t2107_test.go  | 110 +++++++++++++++++++++
 .../moai/development/spec-frontmatter-schema.md    |   2 +-
 6 files changed, 222 insertions(+), 14 deletions(-)
```
헬퍼 호출처 검색(`grep -rn "parseAuthoredByAgent\|applyTrailerAttribution\|distinctAuthoredByAgents\|AmbiguousActors" internal cmd`) 결과, 프로덕션 호출은 `lint_ownership_batch.go:238 applyTrailerAttribution(seg.rec, meta.body)`와 `lint_ownership.go:266 applyTrailerAttribution(rec, c.body)` 두 곳뿐이고, `parseAuthoredByAgent`는 `lint_ownership_test.go:457,478`에서만 호출된다. 세 번째 조회 경로는 없다.

판정 순서(`lint_ownership.go:459-540`): 행렬에 없는 전환은 먼저 조용히 빠지고(`expected == ownerNone`), 그다음 `len(rec.AmbiguousActors) > 0` → Info Unmeasured, `AuthoredByAgent == ""` → Info Unmeasured, 그 뒤에야 행위자 비교와 Warning Invalid가 나온다. 따라서 다중 행위자 본문은 어떤 경우에도 Invalid에 닿지 않는다.

### E2 대상 테스트
`GOCACHE=<scratchpad>/gocache go test ./internal/spec -run 'OwnershipSquashBody|DistinctAuthoredByAgents|Ownership|LintG.tBatch|AuthoredBy' -count=1 -v`:
```
--- PASS: TestOwnershipSquashBody_MultiActorIsUnmeasuredNotInvalid (0.93s)
--- PASS: TestOwnershipSquashBody_SameActorRepeatedStillJudged (1.26s)
--- PASS: TestOwnershipSquashBody_SingleTrailerUnchanged (1.24s)
--- PASS: TestOwnershipSquashBody_BatchEqualsOracle (0.71s)
--- PASS: TestDistinctAuthoredByAgents (0.00s)
ok  	github.com/modu-ai/moai-adk/internal/spec	38.970s
```
로그 집계: `--- PASS` 50, `--- FAIL` 0, `--- SKIP` 0.
(정규식의 `LintG.tBatch`는 `LintGitBatch`와 같은 테스트를 고른다. 원문 정규식은 worktree 가드가 "git" 토큰 때문에 실행을 거부해서 이렇게 바꿨다 — Gaps 참조.)

패키지 전체: `GOCACHE=<scratchpad>/gocache go test ./internal/spec -count=1 -cover`:
```
ok  	github.com/modu-ai/moai-adk/internal/spec	503.792s	coverage: 91.0% of statements
```
(패키지 전체 커버리지다. 새 함수만의 커버리지는 따로 재지 않았다. 다만 `distinctAuthoredByAgents`·`applyTrailerAttribution`의 0/1/다수 세 분기와 `Check`의 다중 행위자 분기는 모두 신규 테스트가 지난다.)

### E3 변이 검사 (추적 파일 무수정, `go test -overlay`)
변이 A — `applyTrailerAttribution`을 첫 트레일러 동작으로 되돌림(`if len(actors) > 1 { actors = actors[:1] }` 삽입, 사본은 scratchpad에만 존재):
```
    lint_ownership_squash_t2107_test.go:45: want Info OwnershipTransitionUnmeasured, got OwnershipTransitionInvalid/warning: SPEC SPEC-SQ-001 transition "draft" → "in-progress" expected owner "manager-develop" but commit eb268aebaf6f8e80bba89b6dd9cca2b828890f82 (Authored-By-Agent: manager-spec) maps to "manager-spec"
--- FAIL: TestOwnershipSquashBody_MultiActorIsUnmeasuredNotInvalid (0.63s)
--- PASS: TestOwnershipSquashBody_SameActorRepeatedStillJudged (2.41s)
--- PASS: TestOwnershipSquashBody_SingleTrailerUnchanged (2.27s)
--- PASS: TestOwnershipSquashBody_BatchEqualsOracle (1.78s)
--- PASS: TestDistinctAuthoredByAgents (0.00s)
FAIL	github.com/modu-ai/moai-adk/internal/spec	7.529s
```
변이 A는 수정 전 동작(첫 트레일러 판정)을 그대로 재현하며, 핵심 테스트가 정확히 수정 전의 오판(Invalid, manager-spec)으로 실패한다.

변이 B — batch 경로만 수정 전으로 되돌림(`seg.rec.AuthoredByAgent = parseAuthoredByAgent(meta.body)`):
```
    lint_ownership_squash_t2107_test.go:90: batched "{... AuthoredByAgent:manager-spec AmbiguousActors:[]}" != oracle "{... AuthoredByAgent: AmbiguousActors:[manager-spec manager-develop]}"
--- FAIL: TestOwnershipSquashBody_BatchEqualsOracle (1.16s)
FAIL	github.com/modu-ai/moai-adk/internal/spec	5.728s
```

### E4 트리에서 빌드한 바이너리로 실제 이력 검사
`go build -o <scratchpad>/moai-aud ./cmd/moai` → `build_exit=0`. 빌드 대상은 HEAD `16e8071f4f` 트리이고 작업 트리는 깨끗하므로, 판정한 빌드의 커밋은 트리 HEAD와 같다(`16e8071f4fa59f5aa33b194c8426882f85ddbb32`). `moai-aud version`은 `moai-adk v3.2.0`을 표시한다.
`<scratchpad>/moai-aud spec lint SPEC-FACTORY-COMPLETION-RECOVERY-001` → `lint_exit=0`:
```
SEVERITY  CODE                           FILE  LINE  MESSAGE
INFO      OwnershipTransitionUnmeasured  .../.moai/specs/SPEC-FACTORY-COMPLETION-RECOVERY-001/spec.md  1     SPEC SPEC-FACTORY-COMPLETION-RECOVERY-001 transition "draft" → "in-progress" expected owner "manager-develop" but commit 55b20438fd9d3cc13e671cd8cdbb43b2ebc2d611 ([t1963] fix(factory): completion recovery M2-B — ... (#1930)) carries Authored-By-Agent trailers of several actors (manager-spec, manager-develop, manager-docs) — squash body, ownership transition unmeasured

0 error(s), 0 warning(s)
```
`OwnershipTransitionInvalid` 없음. 이 트리의 해당 SPEC에는 t2105의 `lint.skip`이 없으므로(진행 기록의 Gap 진술과 일치), 이 결과는 skip 없이 얻은 것이다.

### E5 Craft 도구
- `golangci-lint run --new-from-rev=b25adf4fca05e6dba07abbdf307ccaf2b80b9033 ./internal/spec/` → `0 issues.`
- `gofmt -l internal/spec/lint_ownership.go internal/spec/lint_ownership_batch.go internal/spec/lint_ownership_squash_t2107_test.go` → 출력 없음
- `GOCACHE=<scratchpad>/gocache go vet ./internal/spec` → `vet_exit=0`

### E6 문서 사본 동기
- 변경 행(199-202 구간) 비교 `diff <(sed -n '195,205p' local) <(sed -n '195,205p' template)` → `section_diff_exit=0`.
- 전체 파일 로컬↔템플릿 차이 줄 수: main 기준 4줄, HEAD 기준 4줄(남은 차이는 90행 `(internal/spec/audit.go)` 언급 유무로, 이 카드와 무관한 기존 차이).
- main→HEAD 변화 줄 수: 로컬 4줄, 템플릿 4줄(각각 같은 한 줄 교체).

### E7 탐침 (F1 근거, overlay로 추가한 임시 테스트, 추적 파일 아님)
본문 `Authored-By-Agent: manager-develop` + `Authored-By-Agent: manager-git`, 전환 `draft → in-progress`:
```
    zz_audit_probe_test.go:9: probe finding: OwnershipTransitionUnmeasured info
    zz_audit_probe_test.go:11: probe count=1
ok  	github.com/modu-ai/moai-adk/internal/spec	0.993s
```

### E8 말뭉치 생략 근거 판정 (Claim 6)
코드로 확인한 논리:
- 서로 다른 행위자가 정확히 1명이면 그 행위자는 곧 첫 트레일러다(`distinctAuthoredByAgents`가 첫 등장 순서를 유지). 따라서 판정이 수정 전과 같다.
- 0명이면 수정 전과 같은 Info Unmeasured.
- 2명 이상이면 Info Unmeasured. 수정 전에는 첫 트레일러 기준으로 무발견/Invalid(Warning)/무발견(행렬 밖 행위자) 중 하나였다.
→ 새 Warning 집합은 수정 전 Warning 집합의 부분집합이다. 늘 수 있는 것은 Info뿐이다(E7이 그 예).
- baseline 게이트는 Warning만 센다: `internal/spec/lint_baseline.go:100` `if f.Severity != SeverityWarning || f.Advisory { continue }`. CI는 `go run ./cmd/moai spec lint --baseline .moai/spec-lint-baseline.json`(`.github/workflows/spec-lint.yml:156`)로 돌고, baseline 파일에는 `OwnershipTransition` 항목이 0개다(`grep -c` → 0, 등록 규칙은 `"TierArtifactMissing": 4`뿐).
판정: 레인의 근거는 성립한다. 이 변경이 baseline 게이트를 새로 빨갛게 만들 길은 없다.

## Baseline-attribution
- 모든 명령은 이 실행에서, 트리 `/Users/goos/.moai/worktrees/projects/moai-adk-go/t2107`, HEAD `16e8071f4fa59f5aa33b194c8426882f85ddbb32`(시작·종료 시 재확인, 작업 트리 깨끗함)를 대상으로 돌렸다.
- 바이너리 판정 빌드: 같은 트리에서 경로 지정 빌드(`<scratchpad>/moai-aud`), 빌드 커밋 = 트리 HEAD `16e8071f4f`. 설치된 `moai`는 판정에 쓰지 않았다(감사 계획 조회에만 사용).
- 툴체인: `go version go1.27.3 darwin/arm64`.
- 변이·탐침은 `go test -overlay`로만 적용했고, 사본은 scratchpad(`/private/tmp/claude-501/-Users-goos-MoAI-moai-adk-go/98f002f8-4741-40c2-95ef-a06f9bdaac8b/scratchpad/mut/`)에만 있다.

## Gaps
- **교차 모델 감사 미실행.** `moai verify audit-plan --project-root <tree>` 결과는 `config_status: ok`, `cross_model_active: true`, `cross_model_required: false`, 백엔드 `claude off`, `codex advisory`, `glm off`, `enforced_required: []`. 유일하게 켜진 codex 레그는 advisory이고, 감사 요청자가 "codex 도구를 호출하지 말 것"을 지시했으므로 `audit_multi`/`codex_audit`를 부르지 않았다. 필수 게이트가 없으므로 PASS를 막지 않는다. 영수증 없음.
- **원문 정규식 실행 거부.** 요청된 `-run '...|LintGitBatch|...'` 형태와 `cd ... && ... > file` 형태는 worktree 격리 가드가 "git 형태를 검증할 수 없다"며 거부했다. `LintG.tBatch`로 바꿔 같은 테스트 집합을 돌렸다(E2).
- **공유 Go 빌드 캐시 손상.** 기본 `GOCACHE`(`~/Library/Caches/go-build`)로 돌린 시도가 `could not import ... no such file or directory` / `package internal/... is not in std`로 빌드 단계에서 실패했다(다른 세션이 캐시를 동시에 정리하는 것으로 보임). scratchpad의 사설 `GOCACHE`로 다시 돌린 결과를 증거로 썼다. 변이 A·B 실행과 바이너리 빌드는 그 전에 기본 캐시로 성공했다.
- **말뭉치 전체 lint 미실행**(레인 선언 Gap). E8의 코드 근거로 대체했으며 실측은 아니다.
- **t2105 `lint.skip` 제거 미수행**(레인 선언 Gap). 이 트리에 t2105가 흡수되지 않아 skip 자체가 없다. 제거는 t2105 착지 뒤 후속 작업이다.
- 실제 이력에서 다중 행위자 본문 때문에 판정이 Invalid→Info로 바뀐 SPEC이 몇 개인지는 세지 않았다(말뭉치 미실행과 같은 이유).

## Residual-risk
- squash 병합으로 상태가 바뀐 SPEC은 이제 모두 Info로만 보고되므로, squash로 들어온 진짜 소유권 위반은 Warning으로 잡히지 않는다. 진행 기록이 인정한 정직한 상태이며(squash 본문에서는 상태를 바꾼 하위 commit을 되짚을 수 없음), 되돌리려면 squash 이전 이력이 필요하다.
- F1처럼 비-squash commit이 행렬 밖 행위자 트레일러를 함께 달면 조용한 통과가 Info로 바뀐다. 종료 코드·baseline에는 영향이 없지만 lint 출력의 Info 수가 늘 수 있다.
- 말뭉치 실측이 없으므로, Info 증가 규모는 측정되지 않았다.

## Iteration history
- 1회차(본 감사): PASS 0.94.
