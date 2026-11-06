auditor-model: claude-opus-5-5

# t2144 sync 재감사 판정 (수리 1회차 델타) — 통합 이름 정확 경로 해석

verdict: FAIL
audited_sha: b73989497adbfeb6c8bfdcc6929ea2212349b088
score: 0.90
audit_receipt: none

- 카드: t2144 (Class B, SPEC 없음), 브랜치 WT-merge-readiness-resolves, 작업 트리 clean
- 감사 범위: 수리 델타 `git diff c89eac461c89e9bbb017a7d1bb44fdccd146196e..HEAD -- internal/ .moai/reports/t2144/progress.md` (6a1fd2780d RED, e403a6e996 GREEN, b73989497a 진척 기록). 카드 전체는 `git diff main..HEAD` (main f9d8866951 이 조상)
- 감사일: 2026-11-07, 감사자: sync-auditor (claude-opus-5-5)
- 판정 성격: **코드 결함으로 인한 FAIL 이 아니다.** 차단 결함 F1 은 닫혔고 새 차단 결함은 없다. 감사 계획이 Codex 를 `required` 게이트로 지정하는데 Codex 감사가 리뷰 시작 전에 `stale_server` 로 거절돼 게이트가 충족되지 않았다. 필수 게이트 미충족은 PASS 를 막는다(감사자 규칙: 응답하지 않은 필수 백엔드는 이름을 걸고 fail-closed).

## 반복 이력

| 회차 | 감사 SHA | 판정 | 점수 | 차단 사유 |
|---|---|---|---|---|
| 1 | c89eac461c | FAIL | 0.79 | F1 — 태그 우선 대체 경로가 `origin/main` 태그로 원격 추적 ref 를 가림 (`sync-audit-iter1.md`) |
| 2 | b73989497a | FAIL | 0.90 | 코드 차단 결함 없음. 필수 Codex 게이트 미충족(`stale_server`, 영수증 없음) |

## 차원 점수 (코드 기준)

| 차원 | 점수 | 판정 | 근거 |
|---|---|---|---|
| Functionality (40%) | 92/100 | PASS | P3 재현이 이제 원격 추적 ref 를 잰다(`P3 resolved c11de79c… (remote-tracking), tag at 0b8f43e7…; tree-identity passed=false`). 실제 호출자 전부가 refs/heads 또는 refs/remotes 로 해석됨(아래 호출자 표). cli 4패턴 모두 `ok` |
| Security (25%) | 88/100 | PASS | P1 옵션 주입 5종 거절·부작용 파일 없음, P2 빈 이름 거절. 태그 경로 자체가 없어져 무결성 우회 면 축소. N1(`refs/...` 이름의 커밋 검증 부재)은 fail-closed 로 끝나 Low |
| Craft (20%) | 92/100 | PASS | 변이 7종 모두 커밋된 시험이 잡음(7/7). factorylane 패키지 전체 통과, vet/gofmt 깨끗 |
| Consistency (15%) | 90/100 | PASS | `branchTipSHA` 와 같은 `show-ref --verify --hash` 모양, 거절 문구가 후보 형식을 모두 이름으로 댐. 주석이 F2 동작을 의도로 명시 |

가중 점수 0.90. must-pass(Functionality, Security) 둘 다 PASS. 전체 판정은 필수 감사 게이트 미충족으로 FAIL.

## 이전 결함 처리 상태

| ID | 상태 | 증거 |
|---|---|---|
| F1 [High, blocking] 태그 우선 대체 경로 | **닫힘** | 대체 경로 삭제(merge.go 344-353). P3 프로브 PASS. 변이 m1(대체 경로 복원)을 시험 6건이 잡음 |
| F2 [Low] 로컬 브랜치 `origin/main` 이 원격 추적 ref 를 이김 | 의도된 동작으로 고정 | P4: 해석값 `a3a59e3d…` = `git rev-parse origin/main` 값 = `refs/heads/origin/main` (git 도 `warning: refname 'origin/main' is ambiguous.` 를 내고 같은 값). 시험 `f2_local_branch_named_origin_main_wins_like_git` 이 고정, 변이 m3 가 그 시험에 걸림 |
| F3 [Low] 접미사 이름의 태그 우선 해석 | **닫힘** | P5: `main~0`, `main^`, `main@{0}`, `main^{commit}`, `main:base.txt` 모두 거절. 변이 m6 을 `c_suffixed_names_fail_closed` 가 잡음 |
| F4 [Low, outside] `Develop=HEAD` 자기 비교 통과 | 부수적으로 닫힘 | P6: `HEAD` 거절, `HEAD tree-identity passed=false` |
| F5 [Low, outside] `factory_card_pr.go:258` 안내 문구의 맨이름 | 변화 없음(범위 밖) | 판정에 들어가지 않음. 기록만 |

## 새 결함

- **N1 [Low] [optional] [in-diff]** `internal/factorylane/merge.go:344-351` — `refs/` 로 시작하는 이름은 `show-ref --verify --hash` 로 존재만 확인하고 커밋인지는 확인하지 않는다. 주석과 거절 문구는 "resolved to a commit" 이라고 말한다. P7: `refs/tags/v1`(주석 태그) 해석값이 태그 객체 `e82e2c2a…` 이지 커밋 `17095212…` 이 아니다(merge-tree 가 벗겨 써서 conflict-free/tree-identity 결과는 정상). P7b: 트리를 가리키는 `refs/tags/v2` 는 `ok=true` 로 해석된 뒤 merge-tree 실패가 `reports unresolved conflicts` 로 표시된다 — 결과는 fail-closed 지만 사유가 틀린다. 이전 대체 경로는 `^{commit}` 으로 벗겼다. 도달 경로는 운영자의 `--develop refs/tags/...` 뿐이며 실제 호출자는 넘기지 않는다.
  - Required fix (선택): 해석된 SHA 에 `rev-parse --verify --quiet <sha>^{commit}` 을 한 번 더 적용해 커밋으로 벗기고 실패하면 거절하거나(SHA 입력이라 가림 문제 없음), 주석을 "ref 가 가리키는 객체" 로 고칠 것.
- **N2 [Info] [optional] [in-diff]** 동작 변화: 맨 커밋 SHA(전체·축약), `HEAD`, 접미사 이름, 그리고 git DWIM 의 `refs/remotes/<name>/HEAD` 형식(예: 이름 `origin`)이 더는 해석되지 않는다. 실제 호출자는 이런 값을 넘기지 않는다(아래 표). 운영자가 `--develop <sha>` 를 쓰던 경우만 거절로 바뀌며, 거절 문구가 받는 형식을 알려 준다. 회귀로 보지 않는다.
- **N3 [Info] [outside]** release-flow 의 통합 대상은 `ReleaseBranchPrefix`(`internal/config/loader_workflow_disposition.go:140`) 라 접두사이지 브랜치가 아니다. 수리 전에도 해석되지 않았으므로 이 카드의 회귀가 아니다. 기록만 한다.

## 실제 호출자별 해석 경로 (요청 (b))

`git grep -n -e integrationRevSHA -e EvaluateMergeTriple -- 'internal/*.go' ':!*_test.go'` 결과, 생산 호출자는 두 곳이고 해석기는 factorylane 안의 세 소비처(`checkConflictFree` 366, `checkTreeIdentityPreMerge` 393, `absorbOnlyChange` merge_specless.go:220)에서만 쓰인다. MCP 경로는 `EvaluateMergeTriple` 을 직접 부르지 않는다.

| 호출자 | 넘기는 이름 | 일반 저장소 | 로컬 `main` 없고 `refs/remotes/origin/main` 만 | 연결 worktree | shallow clone |
|---|---|---|---|---|---|
| `factory_merge.go:130` (`moai factory merge ready`), 설정값 `IntegrationTarget` — github-flow/af-flow 고정 `main` | `main` | refs/heads/main | 거절. 수리 전 대체 경로도 실패(P8: `rev-parse main^{commit} err=exit status 1`) → 회귀 아님 | refs/heads/main (P9 ok) | — |
| 같은 경로, git-flow `develop_branch` / gitlab-flow `environment` | `develop` 등 | refs/heads/<name> (cli 시험이 `--develop develop` 로 통과) | 같은 이유로 거절, 회귀 아님 | ok | — |
| 같은 경로, `--develop origin/main` | `origin/main` | refs/remotes/origin/main | refs/remotes/origin/main (P8 ok) | ok (P9) | ok (P9) |
| 같은 경로, `--develop release/vX` | `release/vX` | refs/heads/release/vX | — | ok | — |
| `factory_card_pr.go:273-278` (`moai factory complete` PR 경로), `refs/remotes/origin/<target>` 확인 뒤 | `origin/<target>` | refs/remotes/origin/<target> (태그 `origin/main` 무시 — P3) | 같음 | ok (P9) | ok (P9, `--depth 1` 클론에서 `origin/main` 해석) |
| 같은 경로, 원격 추적 ref 가 없을 때 | `<target>` | refs/heads/<target>; fetch 실패면 어차피 `unverified` 거절 | 거절(수리 전에도 해석 불가) | ok | — |
| `factory_card_pr.go:248` | `"origin/"+target` | 기록용 라벨만, 해석 없음 | — | — | — |

`HEAD` 나 SHA 를 넘기는 실제 호출자는 없다: PR 경로의 `target` 은 `cfg.IntegrationTarget`(197행)이며 `gh pr create --base` 에도 쓰이는 브랜치 이름이다. `--develop` 는 운영자 입력이다.

## 실행한 명령과 관측 출력

### Claim / Evidence

1. `df -h /private/tmp` (시작) → `/dev/disk3s5 3.6Ti 3.6Ti 27Gi 100% …`; (종료) → `… 24Gi 100% …`
2. `go test ./internal/factorylane/ -count=1` → `ok  	github.com/modu-ai/moai-adk/internal/factorylane	167.613s`, exit 0
3. `go test ./internal/cli/ -run MergeReady -count=1` → `ok  	github.com/modu-ai/moai-adk/internal/cli	104.672s`, exit 0
4. `go test ./internal/cli/ -run SpecLess -count=1` → `ok  	github.com/modu-ai/moai-adk/internal/cli	43.727s`, exit 0
5. `go test ./internal/cli/ -run FactoryMerge -count=1` → `ok  	github.com/modu-ai/moai-adk/internal/cli	28.081s`, exit 0
6. `go test ./internal/cli/ -run MergeReadinessBeforePR -count=1` → `ok  	github.com/modu-ai/moai-adk/internal/cli	67.547s`, exit 0. 패턴이 빈 집합이 아님을 `-list` 로 확인: `TestMergeReadinessBeforePRSpecLess`, `TestMergeReadinessBeforePR`
7. `go vet ./internal/factorylane/` → 출력 없음, `vet_exit=0`
8. `gofmt -l internal/factorylane` → 출력 없음, `gofmt_exit=0`
9. 적대 프로브 — 커밋하지 않은 오버레이 시험 `probe_audit2_t2144_test.go` (스크래치 경로), `go test ./internal/factorylane/ -overlay <probe.json> -run TestAudit2ProbeT2144 -count=1 -v`, exit 1 (P4 의 실패는 프로브 자체의 비교 오류 — git 의 `CombinedOutput` 에 ambiguous 경고가 섞여 문자열이 달랐을 뿐, 값은 같음):

```
--- PASS: TestAudit2ProbeT2144/P1_option_injection_refused_no_side_effect (1.19s)
--- PASS: TestAudit2ProbeT2144/P2_empty_blank_refused (0.63s)
--- PASS: TestAudit2ProbeT2144/P3_tag_origin_main_vs_remote_tracking (1.59s)
--- FAIL: TestAudit2ProbeT2144/P4_local_branch_origin_main_beats_remote (1.19s)
--- PASS: TestAudit2ProbeT2144/P5_suffixed_names_refuse (1.54s)
--- PASS: TestAudit2ProbeT2144/P6_HEAD_and_bare_sha_refuse (0.96s)
--- PASS: TestAudit2ProbeT2144/P7_annotated_fully_qualified_tag (1.31s)
--- PASS: TestAudit2ProbeT2144/P8_no_local_main_only_remote_tracking (0.88s)
--- PASS: TestAudit2ProbeT2144/P9_linked_worktree_and_shallow_clone (1.66s)
P1 detail: merge-tree --output=/var/folders/.../pwned + refs/heads/WT-card not attempted: the integration ref "--output=..." cannot be resolved to a commit as refs/heads/--output=... or refs/remotes/--output=... (fail-closed)
P3 resolved c11de79c3f752860dfe321c8887844da4f76a2a0 (remote-tracking), tag at 0b8f43e73956ab1973525952018fcb9d8bb7db24; tree-identity passed=false
P4 resolver=a3a59e3da937223747d0a0a204f5e9a5c1d3b23b git-rev-parse=warning: refname 'origin/main' is ambiguous.
   a3a59e3da937223747d0a0a204f5e9a5c1d3b23b localbranch=a3a59e3da937223747d0a0a204f5e9a5c1d3b23b
P6 "HEAD" refused: the integration ref "HEAD" cannot be resolved to a commit as refs/heads/HEAD or refs/remotes/HEAD (fail-closed)
P6 "285517ec1f…" refused (full and 10-char SHA)
P6 HEAD tree-identity passed=false
P7 resolver=e82e2c2a10f7a8e7fd2d39eadcdc96ba671b6f7a ok=true why="" commit=170952124122185b4589070bf2d548ad91d8d88b tagobj=e82e2c2a10f7a8e7fd2d39eadcdc96ba671b6f7a
P7 conflict-free=true tree-identity=true
P7b tree-tag resolver=4b36dfd79db36d8c59d1fb032de66b57f0457b65 ok=true conflict-free=false detail=merge-tree refs/tags/v2 + refs/heads/WT-card@946749ad… reports unresolved conflicts:
P8 bare main ok=false why=the integration ref "main" cannot be resolved to a commit as refs/heads/main or refs/remotes/main (fail-closed)
P8 old fallback rev-parse main^{commit} err=exit status 1
P8 origin/HEAD symref ok=true sha=f5fd5fab… want=f5fd5fab…
P9 linked-wt "main" ok=true / "origin/main" ok=true / "WT-card" ok=true
P9 shallow origin/main ok=true sha=f68d3524… want=f68d3524…
```

10. 변이 시험 — `merge.go` 복사본을 `-overlay` 로 바꿔 끼우고 `go test ./internal/factorylane/ -overlay <ov_mN.json> -run T2144 -count=1` (커밋된 두 시험 파일 `TestT2144TagShadow`, `TestT2144TagShadowRepair`). 모두 exit 1 = 잡힘:

| 변이 | 내용 | 잡은 시험 (발췌) |
|---|---|---|
| m1 | refs/heads 뒤 태그 우선 `rev-parse <name>^{commit}` 대체 경로 복원(수리 전 코드) | `P3_..._tree_identity`, `P3_..._conflict_free`, `P3_..._does_not_forge_the_absorb_check`, `a_tag_named_like_remote_ref...`, `b_only_a_tag_exists...`, `c_suffixed_names_fail_closed` (6건) |
| m2 | refs/remotes 단계 삭제 | `a_tag_named_like_remote_ref...`, `b_only_a_tag...`, `d_exact_forms_resolve`, `qualified_name_still_works_origin_main` (4건) |
| m3 | refs/heads 단계 삭제 | `b_only_a_tag...`, `d_exact_forms_resolve`, `f2_local_branch_named_origin_main_wins_like_git`, `verdict_valid_only_against_the_branch_is_accepted`, `no_tag_behaviour_unchanged`, `nonexistent_name_is_refused_naming_both_candidates` (6건) |
| m4 | 맨이름에 `refs/tags/<name>` 후보를 맨 앞에 추가 | P3 3건, `a_…`, `b_…`, `d_…`, `TestT2144TagShadow` 4건 (10건) |
| m5 | 미해석 오류를 삼키고 이름 그대로 ok 반환 | `b_only_a_tag...`, `c_suffixed_names_fail_closed`, `nonexistent_name_is_refused_naming_both_candidates` (3건) |
| m6 | `~^@:` 가 든 이름은 `rev-parse` 로 해석 | `c_suffixed_names_fail_closed` (1건) |
| m7 | `refs/` 이름 그대로 확인 분기 삭제 | `d_exact_forms_resolve`, `qualified_name_still_works_refs_heads_main` (2건) |

변이 7/7 사망. 브리프의 "refs/tags/... 허용" 변이는 현재 코드가 이미 `refs/tags/...` 정규형을 받으므로 항등이 된다 — 대신 의미 있는 형태인 m4(맨이름의 태그 후보 추가)로 바꿨다.

11. `moai verify audit-plan --project-root /Users/goos/.moai/worktrees/projects/moai-adk-go/t2144` → exit 0, `"config_status": "ok"`, `"model": "codex"`, codex `"gate": "required"`, claude/glm `"gate": "off"`, `"cross_model_required": true`, `"enforced_required": ["codex"]`
12. `mcp__moai__codex_audit` (mode native, target baseBranch, delta_base `c89eac461c89e9bbb017a7d1bb44fdccd146196e`, project_root 위 경로) 1회 → `{"error":"codex_audit: the serving moai build is stale — stale_server: the serving moai build (/Users/goos/go/bin/moai) differs from the installed build (/Users/goos/go/bin/moai) — reconnect the moai MCP server (restart the host session) so the installed build serves","stale_server":true}` (서버 `v3.2.1-rc.2`, commit `04d8c27a7e`). 재시도하지 않음. 영수증 없음.
13. `git status --short` (종료) → 출력 없음 — 추적 파일 무변경.

### Baseline-attribution

모든 `go test`/`go vet`/`gofmt` 측정은 이 실행에서 트리 `b73989497adbfeb6c8bfdcc6929ea2212349b088`(작업 트리 clean)를 `go test` 가 직접 빌드해 잰 것이다. 프로브와 변이는 `-overlay` 로만 바꿔 끼웠고 트리 파일은 건드리지 않았다. `moai verify audit-plan` 은 설치된 바이너리로 실행했다 — 이 카드의 코드 측정이 아니라 설정 읽기라 도구 출처 귀속(§2.2) 대상이 아니다. git 은 시스템 git(1회차 기록 기준 2.54.0, 이번에 재확인하지 않음).

### Gaps

- **필수 Codex 게이트 미충족 (PASS 차단).** 위 12번 `stale_server`. 결정 d-20261106T163347Z-7adc 의 단계당 Codex 2회 중 이번 호출이 몇 번째로 계산되는지(두 번 모두 리뷰 시작 전 거절)는 리더 판단이다. MCP 서버 재연결 뒤 같은 인자로 1회 실행해 영수증을 받으면 이 Gap 이 닫힌다.
- **거절된 명령 2건 (worktree 격리 가드).** (a) 변수와 `printf` 를 섞은 프로브 실행 명령이 "construct too complex to verify" 로 거절 → 오버레이 JSON 을 Write 로 만들고 평이한 명령으로 다시 실행. (b) `moai verify audit-plan … | head; …; git status` 묶음이 거절 → 세 명령으로 나눠 다시 실행. 둘 다 같은 측정을 그대로 다시 했을 뿐 다른 방법으로 대체한 것이 아니다.
- CI 미검증: 원격 CI 결과를 읽지 않았다(푸시 상태 확인 안 함).
- Windows 미검증: 모든 측정은 darwin 에서만 했다. `show-ref` 경로 처리 차이는 관측하지 않았다.
- cli 4패턴의 개별 시험 이름은 MergeReadinessBeforePR 만 `-list` 로 확인했다. 나머지 셋은 실행 시간(104s/43s/28s)으로 비어 있지 않다고 추론했을 뿐 목록은 보지 않았다.
- `internal/cli` 전체 스위트와 `internal/integration` 은 돌리지 않았다(범위 밖, 부하 규율).

### Residual-risk

- N1: `refs/...` 정규형으로 비커밋 객체를 가리키면 거절은 되지만 사유가 "conflict" 로 잘못 표시될 수 있다.
- 로컬 브랜치 `origin/main`(F2)이 생기면 PR 경로가 그 브랜치를 잰다. git 과 같은 규칙이고 시험으로 고정됐지만, 1회차가 권한 "둘 다 있으면 모호성으로 거절" 안은 채택되지 않았다.
- 정규형 `refs/heads/X` 를 `rev-parse` 로 읽는 다른 패키지(t1928 F4 류)는 이 카드 범위 밖이다.
- 스크래치 증거(프로브·변이 원본, 출력 파일)는 머신 로컬이며 이 파일에 옮긴 줄만 증거로 남는다.

## 다음 단계

코드 수리는 필요 없다(N1 은 선택). MCP 서버를 재연결한 뒤 `mcp__moai__codex_audit` 를 같은 `delta_base` 로 1회 실행하고, 영수증과 Codex 결과에 차단 결함이 없으면 이 판정을 PASS 로 다시 쓰면 된다.
