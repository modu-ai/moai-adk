# SPEC-GRAPH-STAMP-ANCESTRY-001 sync-audit (card t688)

- 감사 대상 트리: `/Users/goos/MoAI/moai-adk-go/.claude/worktrees/t688`
- 브랜치 / HEAD: `WT-graph-stamp-freshness` / `2f984c4b4ba26d54d3031bd74bf4be5d62ea5544`
- 변경 범위 기준점: `8275c82a5..HEAD` (커밋 8개)
- 감사 시각 기준 워킹트리: `git status --porcelain` 0행 (clean)
- 감사자: sync-auditor (독립 재실행. progress.md 주장은 근거로 채택하지 않고 전부 직접 측정)

## 종합 판정

**PASS** — 가중 조화평균 **0.914** (임계 0.80). must-pass 두 축(Functionality 0.96, Security 0.92) 모두 독립 통과.

blocking 결함 0건. optional 3건(아래 §Findings).

## 차원 점수

| 차원 | 점수 | 판정 | 근거(직접 실행) |
|---|---|---|---|
| Functionality (40%) | 0.96 | PASS | `go test -count=1 ./internal/graph/...` → `ok ... 45.796s`; CLI 선택자 20 테스트 전부 `--- PASS`; RED 원장 4본 전부 GREEN(exit 0) |
| Security (25%) | 0.92 | PASS | 신규 의존성 0(`go.mod`/`go.sum` 미변경); git 인자는 shell 미경유 exec 전달; ancestry 실패는 fail-closed |
| Craft (20%) | 0.90 | PASS | mutant guard 테스트 존재; workflow 가드를 텍스트 스캔이 아니라 **실행**해 검증; `golangci-lint` 카드 파일 지적 0건 |
| Consistency (15%) | 0.88 | PASS | verdict enum·threshold·provenance schema 무변경; `.moai/state/**` 0건; 타 SPEC 0건 |

조화평균 = 4 / (1/0.96 + 1/0.92 + 1/0.90 + 1/0.88) = **0.9141**

---

## Claim

1. 두 패키지가 빌드·vet를 통과한다.
2. `internal/graph` 전체 스위트와 acceptance가 지목한 `internal/cli` 선택자가 모두 통과하며, 0건 실행 선택자가 없다.
3. RED 원장 4항목(`RED-GSA-001/002/006/008`)이 현재 HEAD에서 전부 GREEN이다.
4. M4 종결의 두 독립 관측 — stamp ancestry exit 0, codemaps `value<40` `fresh` — 이 같은 트리에서 성립한다.
5. workflow push 분기가 `TARGET="HEAD"`를 쓰고 ancestry 이전 조기 성공이 없으며, 감사에서 지적된 release 조건 도달성 위험이 테스트로 덮여 있다.
6. 불변식 프로브 7종(의존성/verdict enum/provenance 필드/threshold 40/warning-only/자동 restamp/`.moai/state`)이 모두 무변경이다.
7. sync 산출물(CHANGELOG, 상태 전이, spec lint, spec audit drift, lint)이 계약대로다.

## Evidence (명령 + 원문 출력)

### E1 — build / vet

```
$ go build ./internal/graph/... ./internal/cli/...
BUILD_EXIT=0
$ go vet ./internal/graph/... ./internal/cli/...
VET_EXIT=0
```

### E2 — graph 패키지 전체

```
$ go test -count=1 ./internal/graph/...
ok  	github.com/modu-ai/moai-adk/internal/graph	45.796s
ok  	github.com/modu-ai/moai-adk/internal/graph/symbol	1.090s
```

### E3 — CLI 선택자 (acceptance §C / progress:336)

```
$ go test -count=1 -v -run '^(TestGraphCheckCmd_.*|TestGraphStampCmd_.*)$' ./internal/cli
--- PASS: TestGraphCheckCmd_StaleStderrNamesDrivingPaths (1.01s)
--- PASS: TestGraphCheckCmd_JSONCarriesAttribution (0.93s)
--- PASS: TestGraphCheckCmd_CitationsRowStaleExitsOne (1.43s)
--- PASS: TestGraphCheckCmd_AllFreshExitZero (0.93s)
--- PASS: TestGraphCheckCmd_StaleExitsOne (0.95s)
--- PASS: TestGraphCheckCmd_NotComparableExitsTwo (0.37s)
--- PASS: TestGraphCheckCmd_MalformedGateYamlExitsTwo (0.66s)
--- PASS: TestGraphCheckCmd_AbsentExitsOne (0.74s)
--- PASS: TestGraphStampCmd_ExplicitCommitRecordsResolvedSHA (0.85s)
--- PASS: TestGraphStampCmd_ExplicitCommitShortRevResolvesFull (0.61s)
--- PASS: TestGraphStampCmd_UnresolvableCommitRejected (0.46s)
--- PASS: TestGraphStampCmd_CommitWithDirtyTreeRejectedPreWrite (0.44s)
--- PASS: TestGraphStampCmd_FlaglessCharacterization (0.61s)
--- PASS: TestGraphStampCmd_LongTextCarriesRecipe (0.00s)
--- PASS: TestGraphStampCmd_ValidAndReplacement (0.49s)
--- PASS: TestGraphStampCmd_FSErrorsCarryNoAbsolutePath (1.39s)
--- PASS: TestGraphCheckCmd_ExistingNonAncestorStampExitsTwoWithRecovery (1.13s)
--- PASS: TestGraphCheckCmd_UnresolvableStampOmitsRegenerationAdvice (0.63s)
PASS
ok  	github.com/modu-ai/moai-adk/internal/cli	14.682s
```

`=== RUN` 20건 / `--- PASS` 20건 / `--- FAIL` 0건 / `--- SKIP` 0건. **`[no tests to run]` 0건** — 선택자 공허 통과 아님(acceptance §A 마지막 규칙, §F 10항 충족).

M3 회귀 묶음 별도 재실행:

```
$ go test -count=1 -v -run '^(TestCheckCodemaps_BareRestampStaysStale|TestCheckCodemaps_CommittedRegenerationIsFresh|TestCheckCodemaps_UncommittedRegenerationIsFresh|TestCheckCodemaps_BodyAbsentIsAbsentWithoutError|TestCheckCodemaps_DirtyPathCarriesNoAnchor|TestCheckFreshness_AllFresh|TestCheckFreshness_AgedLayerFails|TestCheckFreshness_ThresholdBoundaryOneBelowFresh|TestCheckFreshness_NotComparableIsSystemError)$' ./internal/graph
--- PASS 9건
ok  	github.com/modu-ai/moai-adk/internal/graph	15.073s
```

### E4 — RED 원장 4본 재실행 (RED→GREEN)

```
$ bash .moai/reports/t688/red-ancestry.sh
ok  	github.com/modu-ai/moai-adk/internal/graph	1.432s   → exit 0
$ bash .moai/reports/t688/red-cli-unreachable.sh
ok  	github.com/modu-ai/moai-adk/internal/cli	2.138s     → exit 0
$ bash .moai/reports/t688/red-history-topologies.sh
ok  	github.com/modu-ai/moai-adk/internal/graph	4.416s   → exit 0
$ python3 .moai/reports/t688/red-push-guard.py
push_branch_has_target_head=true
push_branch_exits_before_ancestry=false                    → exit 0
$ grep -c "no tests to run" <3개 로그>
0 / 0 / 0
```

RED 시점 원문(acceptance §D)은 각각 `--- FAIL` + exit 1이었고, 동일 명령이 현재 HEAD에서 exit 0으로 뒤집혔다. AC-GSA-001/002/006/008 GREEN 조건 충족.

### E5 — M4 두 독립 관측 (AC-GSA-009)

```
$ python3 -c "...provenance.json['commit_sha']"
7097e6e214195c45e65cab5fa565b19ca4514c4e
$ git merge-base --is-ancestor 7097e6e21... HEAD
ancestor_exit=0

$ go build -o /tmp/moai-t688 ./cmd/moai && /tmp/moai-t688 graph check --root .
codemaps  metric=described-source-diff value=2 threshold=40 verdict=fresh
mx-index  metric=inventory-content-diff value=0 threshold=1 verdict=fresh
edges     metric=source-fingerprint-mismatch value=2 threshold=0 verdict=stale  (source set(s) moved: reports, specs)
citations metric=positive-cited-path-absence value=0 threshold=0 verdict=fresh
graphcheck_exit=1
```

codemaps 층: `value=2 < 40`, `verdict=fresh`. ancestry exit 0. **같은 트리(HEAD `2f984c4b4`)에 귀속된 두 관측**이며 서로 독립이다.

bare restamp 아님을 별도 확인:

```
$ git diff --stat 8275c82a5..HEAD -- .moai/project/codemaps/
 data-flow.md    | 41 ++++++
 dependencies.md | 37 ++++---
 entry-points.md | 24 ++++
 modules.md      |  3 +-
 overview.md     | 21 ++++-
 provenance.json |  6 +--
 6 files changed, 113 insertions(+), 19 deletions(-)
```

본문 5문서에 113행 삽입 — provenance만 바꾼 bare restamp가 아니다. provenance diff는 `tree_root`/`commit_sha`/`generated_at` 3필드뿐이며 `schema_version: 1` 유지, 필드 추가·삭제 0.

### E6 — workflow 가드

`.github/workflows/graph-freshness.yml` push 분기 원문:

```yaml
          if [ -z "${GITHUB_BASE_REF:-}" ]; then
            TARGET="HEAD"
            echo "guard: push event (no base ref) — judging reachability from HEAD"
          elif [ "${GITHUB_HEAD_REF#release/}" != "$GITHUB_HEAD_REF" ]; then
            TARGET="HEAD"
          else
            TARGET="origin/${GITHUB_BASE_REF}"
          fi
          if ! git -C "$REPO" merge-base --is-ancestor "$SHA" "$TARGET" 2>/dev/null; then
```

`TARGET="HEAD"` 존재 ✅, ancestry 앞의 `exit 0` 부재 ✅.

감사자가 지적한 위험(“`set -euo pipefail` 아래 release 조건이 처음으로 도달 가능해진다”)은 **텍스트 스캔이 아니라 실행 테스트**로 덮여 있다 — `internal/graph/t688_push_guard_test.go`가 workflow YAML에서 스텝 `run:` 본문을 이름으로 추출해 합성 git 이력에 대고 실제로 돌린다:

```
$ go test -count=1 -v -run 'TestGraphFreshnessReachabilityGuard' ./internal/graph
=== RUN   .../push_rejects_an_object-present_non-ancestor_stamp
=== RUN   .../push_accepts_an_ancestor_stamp
=== RUN   .../ordinary_PR_judges_base_ancestry,_not_HEAD
=== RUN   .../release_PR_judges_the_merge-preview_HEAD
=== RUN   .../anchorless_provenance_skips_with_a_reason
=== RUN   .../missing_object_fails_on_every_event_shape
--- PASS: TestGraphFreshnessReachabilityGuard_TargetSelection (7.56s)
ok  	github.com/modu-ai/moai-adk/internal/graph	7.925s
```

6개 서브테스트가 push·ordinary PR·release PR 세 target 경로를 모두 **실행**하며(§G 체크리스트 6번째 항목), `requireBinaries`가 bash/jq/git 부재 시 skip이 아니라 `t.Fatalf`로 죽는다 — 공허 통과 구조적 차단.

### E7 — 불변식 프로브 7종

```
$ git diff --name-only 8275c82a5..HEAD | grep -E 'go\.(mod|sum)'   → NONE
$ git diff --name-only 8275c82a5..HEAD | grep '\.moai/state'        → NONE
$ git diff --name-only 8275c82a5..HEAD | grep '^\.moai/specs/' | grep -v SPEC-GRAPH-STAMP-ANCESTRY-001 → NONE
$ git diff --name-only 8275c82a5..HEAD | grep -vE '^\.moai/(reports|specs|project)/'
.github/workflows/graph-freshness.yml
CHANGELOG.md
internal/cli/graph_check.go
internal/cli/t688_graph_check_ancestry_test.go
internal/graph/check.go
internal/graph/t688_ancestry_test.go
internal/graph/t688_push_guard_test.go
```

비문서 변경 파일은 정확히 7개(workflow 1 + CHANGELOG 1 + 구현 2 + 테스트 3). verdict enum은 `check.go:19-21`의 `fresh/stale/absent` 3개 그대로이며 신규 값 없음. threshold는 `th.CodemapsChangedFiles`를 그대로 읽고 리터럴 40 신규 도입 없음. `warning-only` / 자동 restamp 도입 문자열 0건(유일한 매칭은 CLI 복구 안내문의 “A bare re-stamp … is NOT a fix”로, 자동화가 아니라 그 반대의 금지 문구다).

body-absent C1 경로 보존: `TestGraphCheckCmd_NotComparableExitsTwo` exit 2 PASS, `TestGraphCheckCmd_AbsentExitsOne` exit 1 PASS, `TestCheckCodemaps_BodyAbsentIsAbsentWithoutError` PASS.

### E8 — sync 산출물

```
$ /tmp/moai-t688 spec lint SPEC-GRAPH-STAMP-ANCESTRY-001 --strict
✓ No findings — all SPEC documents are valid            → exit 0

$ /tmp/moai-t688 spec audit --base-dir .
Total SPECs: 851 / Drift findings: 589
(SPEC-GRAPH-STAMP-ANCESTRY-001 관련 finding: 0건 — grep 결과 없음)

$ golangci-lint run ./internal/graph/... ./internal/cli/...
37 issues: errcheck 36, staticcheck 1
(카드 변경 파일 graph_check.go / check.go / t688_* 에 대한 지적: 0건)
```

`golangci-lint` 37건은 전부 `migrate_cg.go`, `gateway_factory.go` 등 카드 무관 파일의 선재 지적이며, run 단계 로그 `.moai/reports/t688/run/m5-lint.log` 말미와 건수·항목이 일치한다(선재 상태 정직 보고).

CHANGELOG 항목 존재 확인: `Changed` 섹션 최상단에 SPEC 링크·요약·잔여 위험을 담은 1행 추가.

상태 전이 커밋:

```
5ea0bac5f docs(...): sync-phase — implemented (t688)     [progress.md, spec.md, CHANGELOG.md]
2f984c4b4 chore(...): Mx-phase audit-ready signal + 3-phase close  [progress.md, spec.md]
```

## Baseline-attribution

모든 수치는 **이번 실행에서, 이 트리(`/Users/goos/MoAI/moai-adk-go/.claude/worktrees/t688`, HEAD `2f984c4b4`)에 대고** 측정했다. progress.md·acceptance.md에 기록된 run 단계 수치는 **비교 대상**으로만 인용했고 근거로 승계하지 않았다. RED 원장의 이전 관측(`tree_sha: 7097e6e21…`, `b4626c042…`)은 서로 다른 트리의 과거 측정으로 명시 인용했다. 감사용 바이너리는 설치본이 아니라 이 HEAD에서 새로 빌드했다(`go build -o /tmp/moai-t688 ./cmd/moai`) — 설치본 지연으로 인한 오귀속 차단.

## Gaps (관측하지 않은 것)

- **G1 — `go test ./...` 전체 스위트 미실행.** 배차 제약이자 `CLAUDE.local.md` §4.1 규율. `internal/cli` 루트 패키지도 재실행하지 않았다(run 단계가 45m 타임아웃으로 이미 실행). 전 패키지 판정은 origin/develop CI 몫이다.
- **G2 — 커버리지 미측정.** `-cover`를 돌리지 않아 Craft 축의 커버리지 성분은 정성 판단(테스트 설계 품질)에만 근거한다. 85% 임계 대조 없음.
- **G3 — 크로스 플랫폼 빌드 미실행.** darwin/arm64만 측정. windows/amd64·linux 판정은 CI 매트릭스 몫.
- **G4 — workflow는 합성 이력 대고만 실행.** 실제 GitHub push / ordinary PR / release PR 이벤트에서의 동작은 관측하지 않았다(구현 측도 CHANGELOG에 같은 잔여 위험으로 기록해 둠 — 은폐 아님).
- **G5 — `moai spec drift`는 이 SPEC 행을 내지 못한다.** 해당 서브커맨드에 `--base-dir`가 없어 primary 체크아웃을 읽고, 이 SPEC은 미병합 브랜치에만 있어 표에 부재한다. drift 판정은 `spec audit --base-dir .`의 finding 0건으로 대체했다.
- **G6 — mutant probe 직접 실행 안 함.** acceptance §C가 열거한 mutant를 내가 주입해 실패를 확인하지는 않았다. 대신 테스트 소스를 읽어 판별력을 확인했다(`TestGraphCheckCmd_UnresolvableStampOmitsRegenerationAdvice`가 명시적 mutant guard, push guard 테스트의 각 행이 “틀린 target이면 반대 판정”이 되도록 fixture를 구성).

## Residual-risk (관측했음에도 여전히 틀릴 수 있는 것)

- **R1 — ancestry fail-closed의 이유 병합.** `verifyStampAncestorOfHEAD`가 `merge-base --is-ancestor`의 비영 종료를 전부 `ErrStampUnreachable`로 접는다. 진짜 비조상과 읽을 수 없는 이력(얕은 클론 등)이 같은 문구로 보고된다. 안전 방향(fail-closed)이고 구현 주석·CHANGELOG가 이를 명시하지만, 얕은 fetch 환경에서 사용자가 잘못된 복구(본문 재생성)를 시도할 여지는 남는다.
- **R2 — `commit_sha` 형식 미검증.** provenance의 `commit_sha`가 hex라는 보장 없이 git 인자로 전달된다. shell 미경유 exec이고 `^{commit}` 접미사가 붙어 선행 하이픈 값도 git 옵션 파싱 실패 → exit 2로 fail-closed하므로 착취 가능하지 않다. 신뢰 경계 안(리포 로컬 산출물)이라 blocking으로 보지 않는다.
- **R3 — `edges` 층 stale로 `graph check` 전체는 exit 1.** 이 카드 자신의 reports/specs 추가가 원인이다. 선재 상태이며(RED 원장 baseline에서도 edges stale), CI workflow는 검사 전에 mx-index·edges를 bootstrap하므로 CI 신호에는 영향이 없다. 다만 로컬에서 `moai graph check`를 돌리는 사람에게는 계속 빨간불이다.
- **R4 — 감사 자체가 단일 관측자.** 교차 모델 백엔드(codex/glm)를 호출하지 않았다. 판정은 이 세션의 직접 실행 근거에만 기반한다.

---

## Findings

blocking 0건.

- **F1** [Low] [optional] `CHANGELOG.md:12` — 항목이 “`spec.md` frontmatter `in-progress → implemented`”라고 적었으나 현재 `spec.md:5`는 `status: completed`다. 후속 Mx 커밋(`2f984c4b4`)이 한 단계 더 옮긴 결과라 CHANGELOG 문구가 최종 상태와 어긋난다. 필요한 수정: 해당 문구를 `in-progress → completed`로 정정하거나, Mx 종결이 별도 커밋임을 한 절로 덧붙인다.
- **F2** [Low] [optional] `CHANGELOG.md:12` — “`sync_commit_sha`는 `pending-backfill-sync` 자리표시자로 쓴다”고 적었으나 `spec.md` frontmatter에 `sync_commit_sha` 키 자체가 없다(grep 0건). 백필 대상이 실재하지 않으므로 문구가 존재하지 않는 후속 작업을 약속한다. 필요한 수정: 문장을 삭제하거나 실제로 키를 추가한다.
- **F3** [Low] [optional] `internal/graph/check.go:262` (`verifyStampAncestorOfHEAD`) — R1의 코드 대응. `merge-base --is-ancestor`의 종료코드 1(진짜 비조상)과 128(이력 판독 불가)을 구분해 reason 문구를 갈라 주면 얕은 클론에서의 오유도를 없앨 수 있다. 현재는 주석이 병합을 의도적 선택으로 명시하고 있어 결함이 아니라 개선 후보다.

세 건 모두 정확성·SPEC 명시 요구사항에 걸리지 않으므로 verdict를 뒤집지 않는다.

## §G Definition of Done 대조

| 항목 | 판정 | 근거 |
|---|---|---|
| `moai spec lint --strict` exit 0 | ✅ | E8 |
| Tier M artifact set + progress skeleton | ✅ | spec/plan/acceptance/progress 4종 존재 |
| AC-001/002/006/008 RED → GREEN 동일 명령 exit 0 | ✅ | E4 |
| AC-004/005/007/010 회귀 가드 ≥1 테스트 실제 실행 통과 | ✅ | E3 (20건), E6 (6건) |
| merge/squash/rebase-like 3 fixture 전부 실행 | ✅ | E4 red-history-topologies 3 서브테스트 |
| push / ordinary PR / release PR 3 target 전부 실행 | ✅ | E6 |
| genuine 재생성 + reachable stamping 별도 증거 | ✅ | E5 (본문 113행 + ancestry exit 0) |
| `<40` 과 ancestry exit 0 이 같은 tree SHA 귀속 | ✅ | E5 (HEAD 2f984c4b4) |
| bare/automatic restamp·schema/verdict/threshold/predicate 무변경 | ✅ | E5, E7 |
| 영향 package test/vet/lint 통과, 전체 suite는 CI | ✅ / G1 | E1·E2·E3·E8; 전체 suite 미실행 |

---

Auditor: sync-auditor (독립 재실행 기반)
