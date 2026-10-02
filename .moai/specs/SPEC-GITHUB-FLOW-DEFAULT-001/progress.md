# SPEC-GITHUB-FLOW-DEFAULT-001 — 진행 기록

## §E.1 Plan-phase Audit-Ready Signal

plan_complete_at: 2026-10-02T15:10:29Z
plan_status: audit-ready
plan_revision: 3
plan_revision_1_complete_at: 2026-10-02T12:02:24Z
plan_revision_2_complete_at: 2026-10-02T12:45:25Z
plan_audit_verdict: PASS 0.91 (iteration 3/3, audited_sha 95811b2d4c9cd40f53e841f5faa54a901cd7df09, report .moai/reports/t1453/plan-audit-iter3.md; iteration 1 FAIL 0.74, iteration 2 FAIL 0.84)

decision record: decided_by=lane-5/orchestrator evidence_refs=.moai/reports/t1453/plan-audit-iter3.md+AUDIT-VERDICT:PASS@95811b2d4 ladder_path=gate:plan→run Kickoff (auto-semantics §9.1 autonomous entry — verdict PASS, plan_status audit-ready, plan-artifact hashes unchanged since the verdict, no open blocker; keep-set steps are outside this card's run scope: the base flip, the main protection change, develop deletion and `moai constitution amend` stay operator/leader-held per SPEC runbook and §G.1)

개정 3(spec 0.1.2, 마지막 감사 회차용)은 plan-audit 2회차(0.84, FAIL, 필수 기준 아홉 개 모두 PASS)의 결함 D1~D6(N-04·N-01·N-03·N-05·N-07·N-02)을 반영했고 선택 D7~D12 도 모두 반영했다. 신호를 적는 조건 — `moai spec lint SPEC-GITHUB-FLOW-DEFAULT-001`(이 트리 HEAD 에서 빌드한 `moai`) 종료 코드 0·`0 error(s), 0 warning(s)`, SPEC 디렉터리 안 표지 문자열 0, REQ 22개·AC 23개 — 은 이 개정 끝에 관측했다. REQ·AC 개수는 늘지 않았다(상한 25·25). 3회차 plan-audit 는 델타 감사(D1~D12 + 회귀 확인 + 순서 동사)로 진행된다.

개정 2(spec 0.1.1)는 plan-audit 1회차(0.74, FAIL, MP-7·MP-9 실패)의 필수 결함 F-01~F-15·F-21 과 권고 F-16~F-19·F-22·F-23 을 반영했고 F-20 은 M3(f) 삭제와 워크플로 이름 정정으로 처분했다. 신호를 적는 조건 — `moai spec lint SPEC-GITHUB-FLOW-DEFAULT-001` 종료 코드 0, SPEC 디렉터리 안 표지 문자열 0 — 은 이 개정 끝에 관측했다. 2회차 plan-audit 는 델타 감사(F 항목 + 회귀 확인)로 진행된다.

## §E.2 Run-phase Evidence

### M1 — PRE-CUTOVER-SAFE base-branch re-point (card t1453)

commits: 9fa6f9bc8 (draft → in-progress), 0b06a57bb (characterization + table RED tests, no production code), 8ebf75fe9 (re-point + fixture precondition seeds). Evidence measured in this run on this tree; test runs executed in the worktree with the lane env scrubbed in one compound call. Raw outputs are machine-local scratch (`/private/tmp/claude-501/-Users-goos-MoAI-moai-adk-go/b3cb8287-30f2-4d77-ac03-cf5f7e4d4065/scratchpad/m1/`), not citation targets; the deciding lines are copied here.

Sites re-pointed to `config.LoadGitFlowIntegrationConfig(root).IntegrationTarget`: E-01 `worktree/done.go` landing base (`landingBase`), E-02 `worktree/sweep.go` `--base` default (`sweepDefaultBase`, run-time, seam `sweepConfigRoot`), E-28 `factory_card.go` `factoryResolveIntegrationBranch` and `factory_merge.go` merge-ready target, E-29 `codex_review_scope.go` `cardMergeBase` and `goal.go` approve `MergeTarget`. A row with no target is a refusal (done `ORIGIN_LANDING_UNCONFIRMED`, sweep asks for `--base`, card base → tree scope, goal approve refuses); the E-28 readers keep their existing caller fallback / refusal. Decision recorded by the orchestrator: option A (no `develop` fallback).

Left alone: `session_worktree_automerge.go:176,183`, `integration.go:353` (design D-1; guard `TestSessionExitAutoMergeInertUnderGitHubFlow`). Read and not re-pointed: `session_worktree.go:803` ref (→ M2(c)), `scripts/jev/triage.py:27` (env-overridable, no YAML reader; recommendation only), kanban landed-ref chain (already config → origin/HEAD → origin/main; not a literal develop). Follow-up finding, not touched: `internal/contract/projection_mission.go:82 missionMergeTarget = "develop"`.

| Item | Command | Observed | Exit | HEAD |
|---|---|---|---|---|
| RED (EXPECTED_RED) | `go test ./internal/cli/ ./internal/cli/worktree/ -run '^(TestDeliveryBaseResolvesFromIntegrationTarget\|TestBaseDefaultsFollowIntegrationTarget)$' -count=1 -v` on 0b06a57bb tree | e.g. `delivery_base_target_test.go:305: MergeTarget = "develop", want "main"`; `base_defaults_test.go:148: sweep must fetch origin/main exactly once, observed [origin/develop]` | 1 | 0b06a57bb |
| E-01 | `git grep -n "landingBaseBranch = " -- internal/cli/worktree/done.go` | (empty) | 1 | 8ebf75fe9 |
| E-02 | `git grep -n "\"origin/develop\"" -- internal/cli/worktree/sweep.go` | (empty) | 1 | 8ebf75fe9 |
| E-28 | `git grep -n "\.DevelopBranch" -- internal/cli/factory_card.go internal/cli/factory_merge.go` | (empty) | 1 | 8ebf75fe9 |
| E-29 | `git grep -n -o -E 'cardBaseBranch +[=] +"develop"\|MergeTarget: "develop"' -- internal/cli/codex_review_scope.go internal/cli/goal.go` | (empty) | 1 | 8ebf75fe9 |
| AC-GFD-001/003 tests, cli | chunk selector incl. `TestDeliveryBase\|TestBaseDefaults\|TestSessionExitAutoMerge` | `--- PASS: TestDeliveryBaseGitFlowUnchanged`, `…ResolvesFromIntegrationTarget`, `TestBaseDefaultsGitFlowUnchanged`, `…FollowIntegrationTarget`, `TestSessionExitAutoMergeInertUnderGitHubFlow`; chunk 3: 307 PASS, 1 FAIL (`TestSD_AC018_ParentCheckoutUntouched`, fixed by `add -f`, re-run `--- PASS`) | 1 → 0 | working tree = 8ebf75fe9 content |
| AC-GFD-003 tests, worktree | `go test ./internal/cli/worktree/ -count=1 -v` (whole package) | `--- PASS: TestBaseDefaultsFollowIntegrationTarget` (21 sub-tests), `…GitFlowUnchanged`; `ok … 166.475s` | 0 | working tree = 8ebf75fe9 content |
| 14 earlier failures | chunk 1 `-run 'TestCodexReview\|TestTreeScope\|TestProduceCodexReviewReceipt\|TestCard'`; chunk 2 `-run 'TestFR_AC\|TestGoal\|TestAutoMission\|TestGTD\|TestAuthoritativeDispatch\|TestNewAutoMission'` | chunk 1 `ok … 180.461s` 33 PASS (all 12 codex card-scope names PASS); chunk 2 `ok … 279.664s` 66 PASS (both `TestFR_AC025_*`, every goal approve test PASS) | 0 | working tree = 8ebf75fe9 content |
| Other cli families | chunk 3 (`TestFactory\|TestSD_\|TestMerge\|TestIntegration\|…`), chunk 4/5 (`TestTodo…`, `TestRun…`) | chunk 4: 551 PASS, 0 FAIL, then `panic: test timed out after 15m0s` (selector too broad). Chunk 5 (71 tests unfinished in chunk 4): 69 PASS, 2 FAIL | 1 | working tree |
| Windows build | `GOOS=windows GOARCH=amd64 go build ./...` | (no output) | 0 | 8ebf75fe9 |
| Lint | `golangci-lint run --timeout=5m ./internal/cli/worktree/ ./internal/cli/` (v2.1.6 = CI pin) | `0 issues.` | 0 | 8ebf75fe9 |
| gofmt | `gofmt -l <13 changed files>` | (empty) | 0 | 8ebf75fe9 |

Baseline (pre-change tree 0028671ed): `./internal/cli/worktree` `ok 101.037s` (256 PASS); `./internal/template` `ok 125.393s`; `./internal/guardstate` FAIL, pre-existing: `census_test.go:80: disk\manifest: .github/workflows/workflow-parse-guard.yaml exists on disk with no manifest entry`, `census_test.go:89: declared 19 entries against 20 workflow files` (`TestCensus_SetDifferenceEmptyBothDirections`).

help text: `moai worktree sweep --help` changed. `--base` flag was `Remote integration base the landing check compares against (default diverges from clean --stale's origin/main)` with default `origin/develop`; it is now `Remote integration base the landing check compares against (default: origin/<configured integration target>)` with an empty flag default resolved at run time. Long text now says `The default base is origin/<the configured integration target> — origin/develop under git-flow, origin/main under github-flow; with no target configured the sweep stops and asks for --base.` `moai worktree done --help` step 2 now names `origin/<integration target> (origin/develop under git-flow; …)`.

Gaps (unobserved, not claimed): (1) whole `internal/cli` package result unknown — the full run timed out at 20m and the broad chunk 4 timed out at 15m; families not covered by chunks 1-5 were not run. (2) Chunk 5 failures `TestTodoConcurrentAdd_8Processes` and `TestTodoAddPick_ConcurrentProcesses`: `todo_test.go:527 / :1033: concurrent add … failed: exit status 4`; exit 4 is the helper-process branch `CLAUDE_PROJECT_DIR == ""` (todo_test.go:572), reached before any code M1 touches, so attributed to the environment/harness rather than M1 — NOT re-run on the baseline tree, so this attribution is inferred, not measured. (3) After-change `./internal/guardstate` and `./internal/template` re-measure not run (no further test runs were permitted); baselines above. (4) `mission_surface_baseline.txt` not re-run through `internal/contract` tests; unchanged by construction — no file under `internal/contract` or `internal/mission` is in the diff. (5) No before/after binary pair for the help text: the "before" strings are quoted from the 0b06a57bb source diff, the "after" from a binary built at 8ebf75fe9. (6) `GOOS=windows go vet ./internal/cli/worktree/...` baseline red (`sweep_test.go parseLsofCWDs`) not re-measured. (7) Under this repo's own git-flow config no real `moai worktree sweep`/`done` run was compared; invariance rests on the characterization tests.

Residual risk: legacy fixtures outside the five seeded constructors that read `cardMergeBase`, `goal approve` or the sweep/done base without a config would now see a refusal; none surfaced in the families run.

§E.3 is NOT written for M1: §E.3 is the run-phase signal for the whole run phase (M1-M6), written only when the run phase is complete.

#### M1 independent re-verification by the orchestrator (this run, HEAD 2f5334456, lane env scrubbed in one compound call; closes Gaps 3, 4 and the guardstate/template part of Gap 3 above)

| Item | Command | Observed | Exit |
|---|---|---|---|
| Ledger E-01/E-02/E-28/E-29 | the four `git grep` commands of the table above, re-run | all four empty | 1 each |
| AC-GFD-001/003 named tests | `go test ./internal/cli ./internal/cli/worktree -run 'TestDeliveryBase\|TestSessionExitAutoMergeInert\|TestBaseDefaults' -count=1 -v` | cli: `--- PASS` TestDeliveryBaseGitFlowUnchanged, TestDeliveryBaseResolvesFromIntegrationTarget, TestBaseDefaultsGitFlowUnchanged, TestBaseDefaultsFollowIntegrationTarget, TestSessionExitAutoMergeInertUnderGitHubFlow, `ok 38.106s`; worktree: both TestBaseDefaults* PASS, `ok 50.624s` | 0 |
| guardstate / template after the change | `go test ./internal/guardstate ./internal/template -count=1` | template `ok 121.984s`; guardstate `FAIL` with exactly the baseline pair `census_test.go:80 … workflow-parse-guard.yaml exists on disk with no manifest entry` and `census_test.go:89 declared 19 entries against 20 workflow files` (no new failure) | 1 (baseline-identical) |
| contract / mission | `go test ./internal/contract ./internal/mission -count=1` | `ok 0.515s`, `ok 6.319s` — the mission surface baseline is unchanged | 0 |
| gofmt | `gofmt -l` on the 9 production and new-test files | (empty) | 0 |

Gap 2 closed by reading, not by a baseline run: `TestTodoConcurrentAdd_8Processes` and `TestTodoAddPick_ConcurrentProcesses` fail with `exit status 4` here too (re-run alone, exit 1). Cause: develop commit `dd44df3cf` (card t1423, "TestMain project-dir scrub") added `os.Unsetenv(config.EnvClaudeProjectDir)` to `internal/cli/main_test.go` `TestMain`; `TestMain` also runs inside the re-executed helper child, so it clears the `CLAUDE_PROJECT_DIR` that the parent put in `cmd.Env`, and `TestTodoHelperProcess` (`todo_test.go:572`) then exits 4. `git diff 0028671ed HEAD -- internal/cli/todo_test.go` shows M1's only change there is the one-line fixture seed (`seedGitFlowPrecondition`), so the failure is a pre-existing develop defect, not M1. Attribution is by reading the code path; the same tests were not run on a pre-M1 tree (still unmeasured) — recommend to the leader as a red-repair input.

## §E.3 Run-phase Audit-Ready Signal

_<pending run-phase>_

## §E.4 Sync-phase Audit-Ready Signal

_<pending sync-phase>_

## §G 리더에게 하는 권고 (이 카드는 카드를 발행하지 않는다 — 카드 발행은 리더의 일)

각 항목은 권고이고 카드 id 는 **리더가 카드를 발행한 뒤 이 자리에 적는다**. 이 카드는 id 를 만들지 않는다.

| # | 권고 | 근거 | 리더가 발행한 카드 id |
|---|---|---|---|
| 1 | 병합 큐 후속 측정: 호스티드 러너 청구 분, CodeRabbit 한도 이력, 녹색 PR 뒤 적색 main 의 빈도를 재현 가능한 방식(창 고정)으로 측정하고 채택 임계를 정해 병합 큐를 다시 판정 | design D-3·D-20, `research.md` §9 | (발행 전) |
| 2 | 와이어 식별자 개명(`push_develop`·`push-develop`·`local-merge-develop`·`--develop-worktree`)과 `contract/testdata/mission_surface_baseline.txt` 의 호환 이주 | design D-12·D-21 | (발행 전) |
| 3 | 계약 모드 에스컬레이션 분류기(`internal/escalation` 의 `pushWhy` 류)가 github-flow 의 카드 브랜치 push 를 허용하도록 하는 변경. 새 행동 어휘와 git-flow 비영향 증명이 필요하며, 그 전까지 계약 모드 레인은 PR 전달의 카드 브랜치 push 에서 에스컬레이션을 만난다 | design D-23 (1회차 감사 F-12) | (발행 전) |
| 4 | develop 삭제 뒤 정리 카드: 워크플로 push 트리거의 `develop` 제거와 잔존 점검(`git grep -n -w develop -- .github/workflows` 종료 코드 1) | design D-13 단계 4, AC-GFD-023 | (발행 전) |
| 5 | develop 적색 CI 수리 카드들(절체의 선행 조건, D-17). 측정된 적색 집합은 `.moai/reports/t1453/m0-develop-ci-red.md` — 수리 여부와 분할은 리더의 판단 | design D-17 | (발행 전) |
| 6 | t1452 카드 본문을 (c) 동일 테스트 명령 재실행 억제와 PR 전 병합 준비 점검으로 좁히는 편집(리더 결정 D-11, 본문 편집은 리더 몫) | design D-11 | 해당 없음(기존 카드) |
| 7 | t810 의 닫는 처분: t1453 이 t810 을 흡수했으므로(D-19) 이 카드가 닫힐 때 운영자에게 처분을 올린다. t810 의 카드·워크트리·`SPEC-LATE-BRANCH-REDESIGN-001` 은 이 카드가 건드리지 않았다 | design D-19 | 해당 없음(기존 카드) |

## §J plan-audit 3회차 이월 항목 (blocking 없음 — run-phase 테스트 설계 입력)

출처: `.moai/reports/t1453/plan-audit-iter3.md` (선택 항목 D13~D19). 처분은 해당 테스트·런북을 쓰는 시점에 한다.

| # | 항목 | 반영 시점 |
|---|---|---|
| D13 | AC-GFD-008 에 rc 태그 대상 검사 1~3 픽스처 부재 (REQ-008 이 1~4·7 을 유지한다고 함) — 픽스처 3개 추가 | AC-008 테스트 작성 시 (M3) |
| D14 | W1 이 문구 grep 이라 인라인 검사를 말만 바꿔 남기고 스크립트 호출만 추가하는 변이가 통과 — W3 이 `run` 본문에 CHANGELOG·`system.yaml` 참조가 없음을 단언하거나 본문 전체 재현 | AC-008 테스트 작성 시 (M3) |
| D15 | 스윕 가드의 `to develop <한정사>` 동사 제외가 브랜치어 규칙보다 먼저 돌아 `Lanes merge to develop and push.` 를 통과시킴 — 제외는 브랜치어 토큰이 없는 적중에만 적용, 적색 픽스처 F-red-15 추가 | 스윕 가드 작성 시 (M4) |
| D16 | 런북 0·9b 단계가 `gh run list --limit 3` 에 기댐(D-28 이 비결정적이라 적시한 바로 그 호출) — 팁 SHA 로 읽고 `gh run view` 로 확인 | 런북 작성 시 (M6) |
| D17 | 허용 목록 상한 40 은 측정 없는 예측 — M4 시작 시점에 실제 항목이 필요한 줄 수를 센다(감사가 git-flow 선택지 파일 쌍에서 `develop` 포함 64줄 관측) | M4 시작 |
| D18 | 표지 채널이 무제한(가짜 RETIRED 제목·H1 표지·살아 있는 줄의 날짜 표지가 살아 있는 문장을 침묵시킴) — 면제 줄 수에 래칫 | 스윕 가드 작성 시 (M4) |
| D19 | REQ-008 은 `-rc.N` 만 명명하나 `prerelease: auto` 는 일반적임, plan.md 의 "(d)" 가 삭제된 M2(d) 와 이름이 겹침 — 문구 정리 | 다음 SPEC 개정 기회 |

### §J.1 리더 사전 공지 — run 단계의 순서·재측정 입력 (plan 산출물은 고치지 않는다: 감사 해시 보존)

리더 공지(10-02 밤, cross-session). 이 항목은 plan 산출물의 문장을 바꾸지 않고 run 단계가 따라야 할 순서 제약과 재측정 의무만 적는다.

- **t1399 개명(lane-3, 마지막 마일스톤 M10, 병합 창 직전)** — `.claude/rules/moai/workflow/kanban-dispatch.md`·`-detail.md`·`-mechanics.md` 를 `factory-dispatch*.md` 로(로컬+템플릿 미러), foreman 스킬 `moai-kanban-foreman` 을 `moai-factory-foreman` 으로, `AGENTS.md`·`AGENTS.local.md`·`CLAUDE.md`·README 4종·docs-site 의 옛 진입 서술을, Go 패키지 `internal/kanban` 을 `internal/factory` 로(import 183 파일) 바꾼다. 착지 시점은 리더가 전 레인에 공지한다.
- **순서(M4 문서 마일스톤)**: t1448 착지 뒤, 그리고 t1399 가 그 전에 착지하면 그 뒤. M4 시작 시 개명이 착지했는지 `git grep -l "kanban-dispatch" -- .claude/rules AGENTS.md CLAUDE.md` 로 관측해 경로 기준을 정한다.
- **재측정 의무(M4 시작·M1 시작)**: plan 산출물의 증거 장부가 인용하는 `internal/kanban/*` 좌표와 `kanban-dispatch*.md` 경로는 개명 뒤 stale 이다. 개명 착지 뒤에는 해당 장부 행을 개명 후 경로로 재측정하고, 스윕 가드(AC-GFD-011/021)의 표면 집합과 허용 목록 파일 경로도 개명 후 경로로 도출한다. 개명 전에 쓴 가드가 옛 경로를 보게 두지 않는다.
- **develop 원격 push**: 리더 공지에 따르면 origin/develop 이 `7109e0900` 까지 push 됐다(t1451·t1442·t1423 포함). 이 SPEC 의 발산 수치(`7614 1`, 트리 `4bf547bca`·`284e09c44` 기준)는 그 시점 측정이고 지금은 stale 일 수 있다 — run 단계 진입 시 develop 을 흡수하고, 발산 수치를 인용하는 모든 판정은 재측정한 SHA 로 다시 적는다. t1451 이 `auto-semantics.md`·`kanban-dispatch.md`·워치독 스킬을 고쳤으므로 M4 의 해당 파일 편집은 흡수 뒤 본문 기준이다.
- 위 사실들은 리더 공지에서 온 것이며 이 레인이 직접 관측한 값이 아니다(Gap).

## §G.1 운영자가 직접 수행하는 단계 (런북, 이 카드의 레인은 실행하지 않는다)

- `moai constitution amend` 두 번(`CONST-V3R5-027`, `CONST-V3R5-028`) — 5단째 인간 승인이 대화형 Y/N 이다(D-26).
- develop→main PR 병합(운영자·리더), develop 보호·삭제.
- main 필수 체크 목록에서 `Release PR Multi-OS Gate` 제거(D-22).
- 리더 몫: develop push 세 번(런북 0·2·4단계), 레인 정지와 재기동.
