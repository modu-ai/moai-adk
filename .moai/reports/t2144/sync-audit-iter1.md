auditor-model: claude-opus-5-5

# t2144 sync 감사 판정 — 통합 이름 태그 가림

verdict: FAIL
audited_sha: c89eac461c89e9bbb017a7d1bb44fdccd146196e
score: 0.79
audit_receipt: none

- 카드: t2144 (Class B, SPEC 없음), 브랜치 WT-merge-readiness-resolves, 감사 범위 `git diff main..HEAD` (main f9d8866951 이 조상)
- 감사일: 2026-11-07, 감사자: sync-auditor (claude-opus-5-5)
- 반복: 1회차 (이전 감사 없음)

## 차원 점수

| 차원 | 점수 | 판정 | 근거 |
|---|---|---|---|
| Functionality (40%) | 68/100 | FAIL | 맨이름 `main` 경로(merge ready)는 고쳐졌으나, 실제 운영 호출자 `factory_card_pr.go:273` 가 넘기는 `origin/<target>` 은 여전히 같은 이름의 태그에 가려진다 — 프로브 P3 재현 (F1) |
| Security (25%) | 82/100 | PASS | 옵션 주입(`--output=...`, `--git-dir`, `-h`, `--all`)·빈 이름 모두 fail-closed, 부작용 파일 없음 (P1/P2 PASS). 무결성 우회(F1)는 Functionality 에서 계산 |
| Craft (20%) | 78/100 | PASS | RED 5건→GREEN, 패키지 전체 통과. 변이 시험은 디스크 고갈로 실행 못 함 (Gap) |
| Consistency (15%) | 90/100 | PASS | `branchTipSHA` 와 같은 "SHA 로 해석 후 소비" 모양, 메시지 형식 유지, gofmt/vet 깨끗 |

must-pass 방화벽: Functionality 가 차단 결함 F1 로 FAIL → 전체 FAIL.

## 결함 목록

- **F1 [High] [blocking] [in-diff, 재현 증거 있음]** `internal/factorylane/merge.go:346` — `integrationRevSHA` 의 2단계 대체 경로가 `rev-parse <name>^{commit}` 로 태그 우선 해석을 그대로 쓴다. `moai factory complete` 의 PR 전달 경로(`internal/cli/factory_card_pr.go:252,278`)는 통합 이름을 `"origin/"+target` 으로 넘기므로, `origin/main` 이라는 이름의 로컬 태그 하나로 conflict-free·tree-identity·흡수 검사가 모두 태그 커밋을 잰다. 카드가 고치려던 결함("통합 브랜치와 같은 이름의 로컬 태그가 브랜치를 가린다")이 두 생산 호출자 중 하나에서 그대로 남아 있다. 진척 기록은 이를 잔여 위험으로 공개했지만, `main` 태그를 막을 근거와 `origin/main` 태그를 막을 근거는 같다.
  - Required fix: 이름이 `refs/` 로 시작하지 않으면 `show-ref --verify --hash` 로 정확한 경로 `refs/heads/<name>` → `refs/remotes/<name>` 순서로만 해석하고, `refs/` 이름은 같은 방식으로 정확한 경로만 확인한다. 태그 우선 `rev-parse` 대체 경로는 없애고 둘 다 실패하면 거절한다. 회귀 시험 1건 추가: `refs/remotes/origin/main` 이 최신이고 태그 `origin/main` 이 옛 커밋일 때 `Develop: "origin/main"` 의 tree-identity 가 거절돼야 한다(아래 P3 그대로). (대안: 호출자가 `refs/remotes/origin/<target>` 을 넘기도록 바꾸는 것도 되지만, 도우미에서 막는 쪽이 모든 호출자를 덮는다.)
- **F2 [Low] [optional] [in-diff]** 같은 곳 — 로컬 브랜치 `origin/main` 이 있으면 1단계가 원격 추적 ref 보다 그것을 먼저 고른다 (P4 재현). git 규칙 순서(heads 가 remotes 보다 앞)와 같아 회귀는 아니다. F1 수리 때 `refs/heads/<name>` 와 `refs/remotes/<name>` 가 둘 다 있으면 모호성으로 거절하는 방안을 검토할 것.
- **F3 [Low] [optional] [in-diff]** `main~0`, `main^` 처럼 접미사가 붙은 이름은 1단계를 건너뛰고 태그 우선으로 해석된다 (P5: 옛 커밋 `28273c97…` 로 해석). 운영자가 넘기는 `--develop` 값이나 설정값에서만 나오므로 도달성은 낮다. F1 수리(태그 우선 대체 경로 제거)로 함께 닫힌다.
- **F4 [Low] [optional] [outside]** `Develop=HEAD` 면 카드 HEAD 를 자기 자신과 합쳐 tree-identity 가 무조건 통과한다 (P6). 이 카드 이전부터 있던 동작이며 값은 운영자 플래그나 설정에서만 온다. 기록만 한다.
- **F5 [Low] [optional] [outside]** `internal/cli/factory_card_pr.go:258` `gitio.IsAncestor(wt, ref, "HEAD")` 도 맨이름 `origin/<target>` 을 쓴다. 안내 문구(behind 표시)에만 쓰이고 판정에는 들어가지 않는다. 기록만 한다.

## factory_merge.go 230/232 (`Branch: developRef`, `worktreeForBranch(developRef)`) 판단

리비전 해석이 없어 태그 가림에 노출되지 않는다. 별도 카드는 필요 없다.
- `worktreeForBranch` → `worktreeForBranchFromList` (`internal/cli/integration.go:229-241`): `git worktree list --porcelain` 출력의 `branch refs/heads/<name>` 줄을 문자열로 정확히 비교한다. 태그가 끼어들 자리가 없다.
- 잠금 기록의 `Branch` 를 읽는 쪽 `internal/cli/integration_merge.go:55-68` → `internal/factory/integration_merge_step.go:299,454,458,771` 은 모두 `refs/heads/<branch>` 정규형이나 `symbolic-ref` 비교를 쓴다. 정규형 `rev-parse refs/heads/X` 는 브랜치가 없을 때만 `refs/tags/refs/heads/X` 태그에 닿는 알려진 F4 류(t1928)이며 이 카드 범위 밖이다.

## 실행한 명령과 관측 출력

### Claim / Evidence

1. `go test ./internal/factorylane/ -count=1` → `ok  	github.com/modu-ai/moai-adk/internal/factorylane	208.445s`, exit 0
2. `go test ./internal/cli/ -run MergeReady -count=1` → 1차: `--- FAIL: TestFactoryMergeReady_HeldWindowRefusedWithHolderNamed (0.00s)` / `testing.go:1581: TempDir: mkdir /var/folders/.../TestFactoryMergeReady_HeldWindowRefusedWithHolderNamed1187764445: no space left on device` (같은 출력에 `no space` 12회), exit 1. 2차: `clang: error: linker command failed with exit code 1` / `FAIL	github.com/modu-ai/moai-adk/internal/cli [build failed]`, exit 1. 두 번 모두 디스크 고갈(`df`: 3.6Ti 중 가용 428Mi → 289Mi → 171Mi, 100%) 때문이며 코드 판정이 아니다.
3. `go test ./internal/cli/ -run SpecLess -count=1`, `go test ./internal/cli/ -run FactoryMerge -count=1` → 디스크 고갈로 실행하지 못함 (Gaps).
4. `go vet ./internal/factorylane/` → 출력 없음, `vet exit=0`
5. `gofmt -l internal/factorylane` → 출력 없음, `gofmt exit=0`
6. `git --version` → `git version 2.54.0 (Apple Git-157)`
7. 적대 프로브 (커밋하지 않은 오버레이 `probe_audit_t2144_test.go`, `go test ./internal/factorylane/ -overlay <probe.json> -run TestAuditProbeT2144 -count=1 -v`, exit 1 — 실패가 곧 재현):

```
--- PASS: TestAuditProbeT2144/P1_option_injection_name_refused_without_side_effect (1.36s)
--- PASS: TestAuditProbeT2144/P2_empty_and_blank_name_refused (1.18s)
--- FAIL: TestAuditProbeT2144/P3_tag_named_origin_main_shadows_remote_tracking_ref (1.85s)
--- FAIL: TestAuditProbeT2144/P4_local_branch_named_origin_main_shadows_remote_tracking_ref (2.50s)
--- FAIL: TestAuditProbeT2144/P5_suffixed_name_falls_back_to_tag_first (2.38s)
--- FAIL: TestAuditProbeT2144/P6_HEAD_as_integration_name (2.12s)
probe_audit_t2144_test.go:58: tree-identity passed against the tag origin/main (old commit), not refs/remotes/origin/main: merge result tree ada05530bab81a7667ff79bd69e466bd5d0fe32a == card branch tree ada05530bab81a7667ff79bd69e466bd5d0fe32a (HEAD^{{tree}} == HEAD^2^{{tree}} holds)
probe_audit_t2144_test.go:70: tree-identity passed against local branch origin/main (old commit), not refs/remotes/origin/main: ...
probe_audit_t2144_test.go:80: main~0 -> ok=true sha=28273c97fa785c79c4577df26a069b05c5dca6eb why= (base=28273c97fa785c79c4577df26a069b05c5dca6eb)
probe_audit_t2144_test.go:91: HEAD: tree-identity passed=true: merge result tree ada05530... == card branch tree ada05530...
probe_audit_t2144_test.go:20: conflict-free refused: merge-tree --output=.../pwned + refs/heads/WT-card not attempted: the integration ref "--output=..." ...
probe_audit_t2144_test.go:45: absorb refused "": the integration branch is unknown (fail-closed)
```

P1 은 `--output=<path>` 이름으로 파일이 만들어지지 않았고 `--git-dir`/`--show-toplevel`/`-h`/`--all` 모두 해석되지 않았다. `show-ref` 인자는 `refs/heads/` 접두로 옵션이 될 수 없고, `rev-parse` 는 `--verify` 아래 `^{commit}` 이 붙어 옵션으로 소비되지 않는다.

8. 코드 읽기: 통합 이름이 git 리비전으로 쓰이는 생산 위치를 `git grep` 으로 전수 확인 — `EvaluateMergeTriple` 호출자는 `internal/cli/factory_merge.go:125` (`Develop: developRef`, 설정값 또는 `--develop`)와 `internal/cli/factory_card_pr.go:273` (`Develop: ref`, `ref = "origin/" + target`) 두 곳. factorylane 안의 리비전 사용(merge-tree 2곳, `merge-base --is-ancestor`, `cardPatch`)은 모두 해석된 SHA 를 쓴다. `checkSyncAuditVerdictFile` 은 이름을 메시지와 `absorbOnlyChange` 전달에만 쓴다.

### Baseline-attribution

모든 측정은 이 실행에서 트리 `c89eac461c89e9bbb017a7d1bb44fdccd146196e` (작업 트리 clean) 에 대해, 시스템 `go` 와 시스템 git 2.54.0 으로 했다. 도구는 `go test` 가 이 트리를 직접 빌드한 것이다.

### Gaps

- **변이 시험(c) 미실행.** 변이 6종(merge-tree 호출 2곳 맨이름 복귀, refs/heads 우선 제거, 미해석 오류 삼키기, 태그로 해석, absorbOnlyChange 맨이름 복귀)의 오버레이 원본을 만드는 스크립트 쓰기가 `ENOSPC: no space left on device` 로 두 번 거절됐고, 이후 Bash 자체가 "this computer's disk is full (ENOSPC)" 로 시작조차 되지 않았다. 변이가 잡히는지는 관측하지 못했다. 다만 레인의 RED 기록(`red-output.txt`)이 맨이름 시절 코드에서 conflict-free·tree-identity·흡수 위조·거절 문구 5건이 실패함을 보여 m1/m2/m6/m4 에 해당하는 원래 동작은 잡힌다는 간접 증거다 — 관측이 아니라 추론이다.
- `go test ./internal/cli/ -run MergeReady|SpecLess|FactoryMerge` 세 가지 모두 디스크 고갈로 관측 실패 (위 Evidence 2·3). 레인 기록 `v2/v3/v4-*.txt` 의 `ok` 는 재관측하지 않았다.
- **Codex 감사 미실시.** `mcp__moai__codex_audit` (native, baseBranch, project_root 지정) 1회 호출이 리뷰 시작 전 `stale_server: the serving moai build (/Users/goos/go/bin/moai) differs from the installed build ... reconnect the moai MCP server` 로 거절됐다. 영수증 없음. 결정 d-20261106T163347Z-7adc 의 단계당 Codex 2회 중 1회로 계산할지는 리더가 판단할 것.
- CI 미검증: 원격 CI 결과를 읽지 않았다.
- `internal/integration` 패키지와 `factory_card_pr.go` 의 나머지 git 호출은 태그 가림 관점으로 전수 확인하지 않았다 (위 F5 만 확인).

### Residual-risk

- F1 수리 뒤에도 정규형 `refs/heads/X` 를 `rev-parse` 로 읽는 다른 패키지 위치는 브랜치가 없을 때 `refs/tags/refs/heads/X` 태그에 닿을 수 있다(t1928 F4 류). 이 카드 범위 밖이다.
- 디스크 고갈 상태에서 측정했으므로, 통과한 측정(패키지 테스트, 프로브)도 같은 조건에서 반복되리라는 보장은 없다. 실패한 측정은 환경 원인으로만 해석했다.

## 다음 단계

F1 만 수리하면 된다(도우미의 대체 경로를 정확한 경로 조회로 바꾸고 `origin/main` 태그 회귀 시험 1건 추가). 재감사는 F1 델타와, 이번에 못 한 변이 시험·`internal/cli` 세 패턴 재실행으로 범위를 좁힌다. 디스크 공간 확보가 먼저다.
