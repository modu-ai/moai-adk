auditor-model: claude-opus-5-5

verdict: PASS
audited_sha: e40ce3868c352b4fcefd0df867ae752d4c912cc7

# t2153 sync 감사 판정 — 린트 델타 (Class B)

- 점수: 0.97 (Functionality 100 / Security 100 / Craft 90 / Consistency 100, 기본 가중치 40/25/20/15)
- 감사 범위: `git diff main..HEAD` (main = `b25adf4fca05e6dba07abbdf307ccaf2b80b9033`) — 코드 2파일
  `internal/agenticfactory/claim/durable.go`(1줄), `internal/cli/worktree/migrate.go`(+4/−18)와 증거 파일 `.moai/reports/t2153/progress.md`
- 차단 결함: 0건. 선택 결함 1건(F1, outside), 환경 참고 1건(F2)
- 감사 1회차(반복 이력 없음). 감사 중 트리는 깨끗했고 HEAD는 처음과 끝 모두 `e40ce3868c`

## 평가 차원

| 차원 | 점수 | 판정 | 근거 |
|---|---|---|---|
| Functionality (40%) | 100 | PASS | 기존 표 테스트 PASS, 원식과 새 식 36조합 전수 비교 일치, 오답 변이 검출, `migrateOneTreeWith` Go 참조 0건, 빌드·vet 통과 |
| Security (25%) | 100 | PASS | 추가 줄에 exec/환경변수/비밀/HTTP/SQL 표면 0건. 부울 재작성과 함수 삭제뿐이라 공격 표면 변화 없음 |
| Craft (20%) | 90 | PASS | CI 고정 v2.1.6과 설치본 v2.15.0 모두 `0 issues.`. 커버리지 미측정(Gap). F1: 표 테스트에 (비 Windows, 비 권한 오류) 행 없음 |
| Consistency (15%) | 100 | PASS | gofmt/린트 통과, 삭제 함수의 설명은 파일 머리말(19~21행)과 `reverseAfterFailure`/`finishTree` 주석에 남아 있음 |

## Claim

1. De Morgan 재작성 `!(goos == "windows" && errors.Is(err, fs.ErrPermission))` → `goos != "windows" || !errors.Is(err, fs.ErrPermission)`은 모든 (goos, err) 조합에서 동작이 같다. `err == nil`은 83~85행에서 먼저 `false`를 반환하며 이 분기는 바뀌지 않았다.
2. 기존 dir-sync 표 테스트는 잘못된 형태 `goos != "windows" && !errors.Is(...)`를 잡아낸다.
3. 재작성 전 형태는 CI 고정 린터에서 QF1001을 내고, 재작성 후 세 패키지는 0건이다.
4. `migrateOneTreeWith`는 저장소의 어떤 `.go` 파일에서도 참조되지 않으며 `prepareTree`/`finishTree`/`forwardGitMove`는 계속 쓰인다.
5. `internal/cli/...`와 `cmd/...`가 빌드되고, 두 패키지(테스트 파일 포함) vet가 통과한다.

## Evidence

모든 명령은 `/Users/goos/.moai/worktrees/projects/moai-adk-go/t2153` 또는 그 `git archive` 사본에서 이번 실행 중에 돌렸다.

- 트리 확인: `git rev-parse --show-toplevel HEAD main` →
  `/Users/goos/.moai/worktrees/projects/moai-adk-go/t2153` / `e40ce3868c352b4fcefd0df867ae752d4c912cc7` / `b25adf4fca05e6dba07abbdf307ccaf2b80b9033`; `git status --short` → 출력 없음
- `git diff --stat main..HEAD` →
  ```
  .moai/reports/t2153/progress.md          | 43 ++++++++++++++++++++++++++++++++
  internal/agenticfactory/claim/durable.go |  2 +-
  internal/cli/worktree/migrate.go         | 22 +++-------------
  3 files changed, 48 insertions(+), 19 deletions(-)
  ```
- Claim 1 — 기존 표 테스트: `go test -count=1 -run DirSync -v ./internal/agenticfactory/claim/` →
  ```
  --- PASS: TestDirSyncFailureIsFatal (0.00s)
  ok  	github.com/modu-ai/moai-adk/internal/agenticfactory/claim	0.201s
  ```
- Claim 1 — 전수 진리표: 스크래치 사본에서 `dirSyncFailureIsFatal`을 main 원식으로 되돌리고, HEAD 식을 `newForm`으로 둔 임시 테스트로 goos 6종(`windows`, `linux`, `darwin`, `freebsd`, `""`, `Windows`) × err 6종(nil, `fs.ErrPermission`, `*fs.PathError{Err: ErrPermission}`, `%w` 래핑 ErrPermission, 일반 오류, `fs.ErrNotExist`)을 비교했다. `go test -count=1 -v -run AuditTruthTable ./internal/agenticfactory/claim/` →
  ```
  zz_audit_truth_test.go:37: compared 36 (goos, err) combinations
  --- PASS: TestAuditTruthTable (0.00s)
  ```
  논리적으로도 ¬(W ∧ P) ≡ ¬W ∨ ¬P이며 `errors.Is` 호출은 순수 함수라 단락 평가 순서 차이가 부수효과를 만들지 않는다.
- Claim 2 — 변이: 스크래치 사본에서 `return goos != "windows" && !errors.Is(err, fs.ErrPermission)`로 바꾸고 `go test -count=1 -run DirSync ./internal/agenticfactory/claim/` →
  ```
  --- FAIL: TestDirSyncFailureIsFatal (0.00s)
      durable_dirsync_t2119_test.go:32: windows other failures are still fatal: dirSyncFailureIsFatal("windows", input/output error) = false, want true
      durable_dirsync_t2119_test.go:32: linux access denied is fatal: dirSyncFailureIsFatal("linux", sync C:\work\docs\checkpoint: permission denied) = false, want true
      durable_dirsync_t2119_test.go:32: darwin access denied is fatal: dirSyncFailureIsFatal("darwin", sync C:\work\docs\checkpoint: permission denied) = false, want true
  FAIL	github.com/modu-ai/moai-adk/internal/agenticfactory/claim	0.388s
  ```
  (exit 1)
- Claim 3 — 기준선(RED): 스크래치 사본의 원식에 `GOTOOLCHAIN=go1.26.8 GOCACHE=<scratchpad>/gocache-t2153-126 ~/go/bin/golangci-lint.v2.1.6.bak run --allow-parallel-runners --config <worktree>/.golangci.yml ./internal/agenticfactory/claim` →
  ```
  .../internal/agenticfactory/claim/durable.go:86:9: QF1001: could apply De Morgan's law (staticcheck)
  	return !(goos == "windows" && errors.Is(err, fs.ErrPermission))
  1 issues:
  * staticcheck: 1
  ```
  (exit 1)
- Claim 3 — 감사 트리, CI 고정 린터: `GOTOOLCHAIN=go1.26.8 GOCACHE=.../gocache-t2153-126 ~/go/bin/golangci-lint.v2.1.6.bak run --allow-parallel-runners ./internal/agenticfactory/claim ./internal/cli/worktree ./internal/hook` → `0 issues.` (exit 0)
- Claim 3 — 감사 트리, 설치본: `golangci-lint run --allow-parallel-runners ./internal/cli/worktree` → `0 issues.` (exit 0)
- 린터 버전: `golangci-lint.v2.1.6.bak --version` → `golangci-lint has version v2.1.6 built with go1.26.8 ...`; `golangci-lint --version` → `golangci-lint has version 2.15.0 built with go1.27.2 from 25c5369 on 2026-10-16T10:51:54Z`; `grep -n 'golangci-lint@' .github/workflows/ci.yml` → `865: run: go install github.com/golangci/golangci-lint/v2/cmd/golangci-lint@v2.1.6`
- Claim 4 — 호출처: `grep -rn --include='*.go' 'migrateOneTreeWith' <worktree>` → 출력 없음, exit 1. Go 외 참조는 `.moai/reports/t2153/progress.md`, `.moai/reports/t2094/sync-audit-iter2.md`, `.moai/reports/t2070/sync-audit.md`, `.moai/reports/t2070/progress.md`(과거 보고서, 무해).
  남은 헬퍼 사용처: `migrate.go:1006 prep := prepareTree(...)`, `migrate.go:1021 a.moveErr = forwardGitMove(...)`, `migrate.go:1069 out = finishTree(...)`
- Claim 5 — `go build ./internal/cli/... ./cmd/...` → 출력 없음, exit 0; `go vet ./internal/cli/worktree/ ./internal/agenticfactory/claim/` → 출력 없음, exit 0 (vet은 `_test.go`도 컴파일하므로 테스트 쪽 호출처 부재도 확인됨)
- 보안 탐침: `git diff main..HEAD -- internal/ | grep -E '^\+' | grep -ciE 'exec\.|os\.Getenv|password|token|secret|http\.|sql'` → `0`
- 감사 계획: `moai verify audit-plan --project-root <worktree>` → `config_status: ok`, `cross_model_active: true`, `cross_model_required: false`, `enforced_required: []`, codex gate `advisory`, claude/glm `off`

## Findings (structured defect-list)

- F1 [Low] [optional] [outside] `internal/agenticfactory/claim/durable_dirsync_t2119_test.go:24-28` — 표 테스트에 (비 Windows, 비 권한 오류) 행이 없다. 이 칸만 다른 변이, 예컨대 XOR 형태 `(goos == "windows") != errors.Is(err, fs.ErrPermission)`는 기존 5행을 모두 통과한다(손 계산, 실행 안 함). 이번 재작성은 전수 비교로 같음이 확인되어 판정에는 영향이 없다. Required fix(선택): `{"linux other failures are fatal", "linux", other, true}` 행 추가. 신뢰도 높음.
- F2 [Info] [optional] [outside] 로컬 환경 — CI 고정 v2.1.6 바이너리가 go1.26.8로 빌드되어 로컬 go1.27.3에서는 `GOTOOLCHAIN=go1.26.8` 없이 패닉한다(progress.md 기록과 일치; 이번 실행은 그 변수로만 돌림). 코드 결함 아님. Required fix: 없음(필요하면 로컬 고정 린터를 현 툴체인으로 재빌드).

## Baseline-attribution

- 감사 트리: `e40ce3868c352b4fcefd0df867ae752d4c912cc7`(깨끗한 상태), 비교 기준 main `b25adf4fca05e6dba07abbdf307ccaf2b80b9033`
- 변이·진리표·기준선 린트는 `git archive e40ce3868c... go.mod go.sum internal pkg`를 `<scratchpad>/t2153-mutant`에 풀어 만든 사본에서만 실행했다. 감사 트리는 수정하지 않았다.
- 린터 판정 빌드: CI 고정 `golangci-lint v2.1.6`(go1.26.8 빌드, `GOTOOLCHAIN=go1.26.8`)과 설치본 `v2.15.0`(커밋 `25c5369`). 두 린터 모두 외부 릴리스이며 감사 대상 트리에서 빌드되는 도구가 아니므로 §2.2 도구-출처 지연 문제는 해당 없음.

## Gaps

- 패키지 전체 테스트(`go test ./internal/cli/worktree/`, `./internal/agenticfactory/claim/` 전체)와 커버리지는 이번 실행에서 돌리지 않았다. 기계 부하(감사 시작 시 load 92)를 고려해 대상 테스트·빌드·vet·린트만 실행했다. progress.md의 패키지 재실행 결과는 FF 이전 트리 기준이며 이번 감사가 재관측한 것이 아니다.
- `go build`는 `./internal/cli/...`와 `./cmd/...`만 대상이다. `internal/cli/worktree`를 가져오는 다른 패키지가 그 밖에 있다면 빌드되지 않았다(삭제 함수가 비공개라 다른 패키지에서 참조할 수 없으므로 실제 위험은 없음).
- Windows에서 실행하지 않았다. 진리표는 `goos` 문자열을 인자로 넣어 플랫폼 독립적으로 검증했다.
- 도구 거부 1건: 스크래치 사본에서 `go test -count=1 -v -run 'AuditTruthTable|DirSync' ...` 실행이 worktree 격리 가드에 거부됐다(사유: "this command runs go with the text AuditTruthTable|DirSync inside a construct too complex to verify ... Refusing to run it"). 대신 `-run AuditTruthTable`만으로 다시 실행했고, DirSync는 감사 트리와 변이 사본에서 따로 돌렸다.
- Codex 교차 감사: 이 파일을 쓴 뒤 선택 사항으로 판단한다(아래 부록 참조).

## Residual-risk

- F1의 테스트 공백으로, 앞으로 이 함수를 다시 고칠 때 (비 Windows, 비 권한 오류) 칸만 바뀌는 회귀는 기존 테스트가 잡지 못한다.
- 이번 감사는 린트 델타(1줄 재작성, 비공개 함수 삭제)만 다룬다. main에서 넘어온 나머지 트리 내용(리더 배치 `4fb7c8fdf1`)은 범위 밖이다.

## 부록 — Codex 교차 감사 (advisory)

- 호출 1회: `mcp__moai__codex_audit(mode=native, target=baseBranch, delta_base=b25adf4fca05e6dba07abbdf307ccaf2b80b9033, project_root=/Users/goos/.moai/worktrees/projects/moai-adk-go/t2153)`
- 결과: 거부(리뷰 미실행), 원문:
  `codex_audit: the serving moai build is stale — stale_server: the serving moai build (/Users/goos/go/bin/moai) differs from the installed build (/Users/goos/go/bin/moai) — reconnect the moai MCP server (restart the host session) so the installed build serves` (`_server.commit` `04d8c27a7e`, `v3.2.1-rc.2`)
- 지시에 따라 재시도하지 않았다. receipt 없음. codex 게이트가 advisory(`enforced_required: []`)라 이 미실행은 PASS를 막지 않지만, 교차 모델 의견은 없는 상태로 남는다(Gap).
