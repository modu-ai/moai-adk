auditor-model: claude-opus-5-5

# t2087 sync 감사 — 2회차(수리 델타) 판정

verdict: FAIL
audited_sha: a6663e9417718fcd3481c91aa573aefec4617d4d

- 카드: t2087 (Class B, SPEC 없음)
- 점수: 84/100 — FAIL (차단 결함 1건, 수리에서 생김)
- 감사 범위: 수리 델타 `140c1ca406..a6663e9417` (커밋 `23e4ae36de` 1회차 판정 기록, `fb2c9dbb37` F1/F2/F3 수리, `a6663e9417` 픽스처 유지보수 차단). 1회차 지적이 수리된 코드에서도 여전한지 함께 확인했다.
- Codex 영수증: `rcpt-4bf94c4b99b843a5d0c8b2e4` (카드당 2회 중 2번째, 결정 `d-20261106T064700Z-de2d`)

## 요약

F1, F2, F3 수리와 플레이크 수리는 모두 확인됐다. F2와 F3는 세 바이너리(기준 `9b632c1eb3`, 수리 전 `140c1ca406`, 수리 후 `a6663e9417`)로 재현했다. 기준과 수리 후는 결과가 같았고 수리 전만 달랐다. 기존 검사는 모두 통과했다.

그런데 이번 수리가 새 시험 `TestLintGitBatch_QuotedPathHeaderClosesSegment`를 넣으면서 Windows 결함을 하나 새로 만들었다. 이 시험은 디렉터리 이름에 `"`를 쓰는데, Windows는 파일 이름에 `"`를 허용하지 않는다. Windows 제외 처리도 없어서 `release-pr-multi-os.yml`의 windows `remainder` 레그가 실패한다. Codex가 P1으로 지적했고 나도 정적으로 확인했다. 1회차 F1과 같은 부류(Windows 릴리스 레그 차단)라서 차단으로 판정한다.

## 결함 목록

| id | 심각도 | 분류 | 위치 | 내용 | 필요한 수정 |
|---|---|---|---|---|---|
| N1 | High (Codex P1) | blocking, in-diff, 수리에서 생김(`fb2c9dbb37`), 정적 검증 | `internal/spec/lint_gitbatch_test.go:498-528` (`TestLintGitBatch_QuotedPathHeaderClosesSegment`) | `r.spec(\`SPEC-Z"Q-001\`, ...)`가 `os.MkdirAll`로 `"`가 든 디렉터리를 만든다. Windows는 파일 이름에 `"`를 허용하지 않으므로 `specAs`의 `r.t.Fatalf("mkdir: %v", err)`로 시험이 실패한다. 파일 안의 GOOS 제외는 348행(`gbShim`)과 457행(구버전 git 대체 경로 시험)뿐이고 이 시험에는 없다. `release-pr-multi-os.yml` 행렬은 `os: [ubuntu-latest, macos-latest, windows-latest]` × `shard: [..., remainder]`이고, `remainder`는 cli 샤드를 뺀 모든 패키지를 돌리므로 `internal/spec`도 포함된다. | 둘 중 하나를 고른다. (a) 권장: 파일을 만들지 않고 합성한 `git log -p` 입력(`diff --git "a/.moai/specs/SPEC-Z\"Q-001/spec.md" ...` 헤더 포함)을 `parseOwnershipBatch(io.Reader)`에 직접 넣는 시험으로 바꾼다. 파서 수정 자체는 OS와 무관하므로 Windows에서도 F3 회귀를 지킨다. (b) 다른 시험처럼 `runtime.GOOS == "windows"`이면 이유를 적어 `t.Skip`한다. 이 경우 Windows에서는 F3 시험이 돌지 않는다. 수리한 뒤에는 `GOOS=windows go vet`만으로는 부족하다. 이 결함은 컴파일이 아니라 실행 시점 결함이기 때문이다. |
| O1 | Low | optional, outside(같은 패키지의 다른 시험 파일), 추론 | `internal/spec/{drift_era_terminal,ac_baseline_commit_guard,lint_status_unreachable,closer,archive_git,lint_transition,drift_chore_skip}_test.go` | 이 픽스처들도 `git init` 저장소를 만들지만 `gc.auto=0`/`maintenance.auto=false`를 설정하지 않는다. 기본 설정에서는 커밋 3개짜리 저장소도 분리된 `maintenance run` 자식을 3개 띄운다(실측, 아래). 작업이 없는 실행이 `.git/objects`에 쓰는지(잠금 파일 등)는 측정하지 않았다. 같은 정리 플레이크가 다른 시험에서 날 수 있는 경로다. | 차단 아님. 같은 정리 실패가 다른 시험에서 보이면 공통 헬퍼로 같은 설정을 넣는다. 지금 넣을 필요는 없다. |
| F4 | Low | debt(1회차부터 이월, 이번 판정 대상 아님) | `batchedOwnershipTransition` root별 pathspec | `_archive` 경로 이름 변경에서 Info 출력이 다르다. | 부채로 유지 |
| F5 | Info | debt(잔여 위험) | 이력 단순화 | 집합 단위 단순화와 `--follow`의 차이 | 부채로 유지 |
| F6 | Low | debt | 배치 실패 시 per-SPEC 경로로 조용히 대체 | 대체가 일어나도 알리지 않는다(결과는 정확하지만 느리다) | 부채로 유지 |

반복 수리 중단 규칙: 이번이 수리에서 생긴 차단 결함의 첫 번째 사례다(1회차는 첫 감사라 해당 없음). 규칙의 한도(연속 2회)에는 아직 닿지 않았다. 다음 감사에서도 수리에서 생긴 차단 결함이 나오면 패치를 더 요청하지 말고 되돌리기·재설계·범위 축소·분할 중 하나를 고르고 그 선택을 기록해야 한다.

## 차원별 판단

| 차원 | 점수 | 판정 | 근거 |
|---|---|---|---|
| Functionality (40%) | 88 | PASS | F2, F3 재현에서 기준과 수리 후가 일치하고 수리 전만 달랐다(아래 증거 1, 2). 구버전 git 대체 경로 시험이 통과했다. 베이스라인 위반 출력은 기준과 바이트 단위로 같다. |
| Security (25%) | 95 | PASS | 델타는 `git log` 플래그 하나 추가(`--no-relative`), 헤더 파싱 조건, 시험 코드뿐이다. 새 입력 경로나 비밀값은 없다. `exec.Command`에 셸을 거치지 않는 인자 배열을 넘기는 방식은 그대로다. |
| Craft (20%) | 60 | FAIL | N1: 새 시험이 Windows 릴리스 레그를 깨뜨린다. 그 밖에는 build, vet, gofmt, 집중 시험, count=3 반복이 모두 통과했다. |
| Consistency (15%) | 88 | PASS | 플레이크 수리와 Windows 제외 문구는 파일의 기존 관례를 따른다. N1은 같은 파일의 GOOS 제외 관례를 빠뜨린 것이다. |

가중합: 0.4×88 + 0.25×95 + 0.2×60 + 0.15×88 = 84.0. 최종 판정은 차단 결함 N1 때문에 FAIL이다.

## 증거 (Claim / Evidence / Baseline-attribution / Gaps / Residual-risk)

### 0. 트리 확인 (모든 작업 전, Codex 호출 직전에 다시 확인)
- 주장: 맞는 트리에서 감사했다.
- 증거: `git rev-parse --show-toplevel` → `/Users/goos/.moai/worktrees/projects/moai-adk-go/t2087`; `git branch --show-current` → `WT-spec-lint-baseline-2`; `git rev-parse HEAD` → `a6663e9417718fcd3481c91aa573aefec4617d4d`; `git status --short` → 빈 출력. Codex 호출 직전에도 HEAD가 같고 상태가 비어 있었다.
- 델타 범위: `git diff --stat 140c1ca406 a6663e9417` → `.moai/reports/t2087/sync-audit-verdict.md | 125 +`, `internal/spec/lint_gitbatch_test.go | 146 +`, `internal/spec/lint_ownership_batch.go | 24 +-`. 판정 파일은 감사 기록 커밋 `23e4ae36de`에서 왔고 코드는 두 파일뿐이다. `git diff --stat 9b632c1eb3 a6663e9417 -- .github .moai/spec-lint-baseline.json` → 빈 출력(exit 0): 카드 전체로 봐도 워크플로와 베이스라인을 건드리지 않았다.

### 1. F2 — `diff.relative=true` + 하위 디렉터리 실행
- 주장: 수리 후 바이너리는 기준과 같은 OwnershipTransitionInvalid를 내고, 수리 전 바이너리는 그 경고를 잃는다.
- 바이너리(도구 출처 §2.2: 모두 경로로 직접 호출): `moai-base`는 `git archive 9b632c1eb3` 추출본에서 빌드(exit 0), `moai-unrep`는 `git archive 140c1ca406` 추출본에서 빌드(exit 0), `moai-cand`는 이 트리(HEAD `a6663e9417`)에서 `go build -o ... ./cmd/moai`(exit 0).
- 시나리오(스크래치 `repro.sh`): 격리한 사용자 설정(`GIT_CONFIG_GLOBAL`)에 `diff.relative = true`를 넣었다. `SPEC-REL-001`을 만들고(커밋 트레일러 manager-spec), `draft → in-progress` 커밋은 트레일러를 `manager-docs`로 했다. `.moai/specs`에서 `spec lint SPEC-REL-001/spec.md`를 실행했다. 실효 설정: `F2 effective diff.relative: true`.
- 증거(그대로 인용):
  - 기준: `[f2-base] exit=0` / `WARNING   OwnershipTransitionInvalid  SPEC-REL-001/spec.md  1     SPEC SPEC-REL-001 transition "draft" → "in-progress" expected owner "manager-develop" but commit 71a7d91706da4c68230db27c1ca407b16409b103 (Authored-By-Agent: manager-docs) maps to "manager-docs"`
  - 수리 후: `[f2-cand] exit=0` / 같은 경고, 같은 커밋 `71a7d917…`
  - 수리 전(대조군): `[f2-unrep] exit=0` / `[f2-unrep] (no Ownership line)`
  - 수리 후, 저장소 루트에서 실행(대조): 같은 경고
- 구버전 git 대체 경로: `TestLintGitBatch_OldGitRejectingNoRelativeFallsBack`(`--no-relative`를 받으면 exit 129로 끝나는 셸 심)이 count=3 실행에서 3번 모두 PASS했다. 이 시험은 `ownershipBatches[".moai/specs"].err != nil`도 단언하므로 배치 실패 경로를 실제로 탄다는 것까지 고정된다. 코드상으로도 `batchedOwnershipTransition`은 `b.err != nil`이면 `(nil, false)`를 돌려주어 per-SPEC 경로로 넘긴다(94-96행).

### 2. F3 — 인용 경로 헤더
- 주장: 인용되지 않은 SPEC의 판정이 기준과 같다.
- 시나리오: `SPEC-A-001`(본문만 수정)과 `SPEC-Z"Q-001`(`draft → in-progress`, manager-docs)을 같은 커밋에 넣었다. `F3 quoted header present: 1`(`git log -p -1`에 `diff --git "a/` 헤더 1개).
- 증거:
  - `spec lint .moai/specs/SPEC-A-001/spec.md`: 기준 `[f3-base] (no Ownership line)`, 수리 후 `[f3-cand] (no Ownership line)`, 수리 전 `WARNING   OwnershipTransitionInvalid  .moai/specs/SPEC-A-001/spec.md  1     SPEC SPEC-A-001 transition "draft" → "in-progress" ... commit bbf044812f5b... (Authored-By-Agent: manager-docs)` (잘못된 귀속)
  - 전체 말뭉치(인자 없음): 기준과 수리 후 모두 `SPEC-Z"Q-001`에 대한 Invalid 1건만 냈다(같은 커밋 `bbf04481…`). 인용 경로 SPEC 자신은 `batchSafePath`가 `"`를 거부해 per-SPEC 경로로 가므로 기준과 같다.
- 코드 확인: 이제 `diff --git `로 시작하는 줄은 모두 `flush()`하고, `a/... b/...` 형태만 세그먼트를 연다(`inSeg = true`가 ` b/` 검사 안으로 옮겨졌다). `batchSafePath`가 ` b/`가 든 경로를 거부하므로 정상 경로의 동작은 바뀌지 않는다.

### 3. F1 — Windows PATH 심 시험
- 주장: 상수 호출 수 시험은 Windows에서 이유를 밝히고 건너뛰며, 패키지는 Windows용으로 컴파일된다.
- 증거: `gbShim` 348-349행 `if runtime.GOOS == "windows" { t.Skip("PATH shim is an extensionless sh script, not resolvable by exec.LookPath on Windows") }`, 같은 문구가 457-458행에도 있다. `GOOS=windows go vet ./internal/spec/` → `windows_vet_exit=0`.
- Windows 실행은 관찰하지 않았다. Skip이 적절한지에 대해: `.cmd` 심의 인자 인용을 Windows 없이 증명할 수 없다는 구현자 판단은 타당하다. Unix 두 레그가 상수 호출 수를 계속 증명하므로 받아들일 만한 비용이다. F1은 해결로 본다. 다만 N1에서 보듯 같은 수리 묶음의 다른 새 시험에는 같은 조치가 빠졌다.

### 4. 플레이크 수리
- 증거:
  - `go test -count=2 -run TestLintGitBatch_FixtureCoversConventions ./internal/spec/` → `flake_count2_exit=0`, `ok  	github.com/modu-ai/moai-adk/internal/spec	16.372s`
  - `go test -count=3 -v -run TestLintGitBatch ./internal/spec/` → `gitbatch_count3_exit=0`, `ok ... 63.637s`. 9개 시험(하위 시험 포함) 각각 `--- PASS` 3회, FAIL·SKIP 0건(macOS 기준).
- 원인 실측(스크래치 `maint.sh`, 격리 설정, `git version 2.54.0 (Apple Git-157)`, `GIT_TRACE2_EVENT`):
  - 기본 설정, 100커밋: `maintenance child starts: 100`, `repack argv events: 8`, `packs: 2`
  - `gc.auto=0` + `maintenance.auto=false`, 100커밋: `maintenance child starts: 0`, `repack argv events: 0`, `packs: 0`
  - 기본 설정, 3커밋: `maintenance child starts: 3`, `repack argv events: 0`, `packs: 0`
- 판단: 구현자가 적은 기제(커밋마다 분리된 `maintenance run --auto`, 그중 일부가 repack해 `.git/objects`에 씀)를 독립적으로 재현했다. 이 기제가 원래 실패(`TempDir RemoveAll cleanup: ... .git/objects: directory not empty`)의 원인일 신뢰도는 중상이다. 기제는 측정됐고 실패 메시지와도 맞지만, 원래 실패는 나도 구현자도 수리 전 코드에서 다시 재현하지 못했다. 두 번째 원인 후보도 남아 있다. (i) 작업이 없는 분리 실행이 `.git/objects`에 잠금 파일을 만들 가능성(측정 안 함), (ii) 진행 기록에 적힌 `never yields ... completed` 메시지의 원인(구현자 기록상 미규명). `newGBRepo` 밖에서 저장소를 만드는 같은 패키지의 다른 시험은 이 설정을 갖지 않는다(O1). git-batch 시험 파일 안에서 저장소를 만드는 곳은 `newGBRepo` 하나다(`grep '"init"'` 결과 36행뿐).

### 5. 기존 검사
- `go build ./...` → `go_build_exit=0`
- `go vet ./internal/spec/...` → `go_vet_exit=0`
- `gofmt -l internal/spec` → 출력 없음, `gofmt_exit=0`
- `go test -count=1 -run 'TestLintGitBatch|Ownership|ImpliedStatus|StatusGitConsistency|ConstantGitLog|Unreachable' ./internal/spec/` → `focused_exit=0`, `ok ... 40.621s`
- 기준 귀속: 모두 이 트리(HEAD `a6663e9417`)에서 이번 실행에 측정했다.

### 6. 베이스라인 위반 (이 카드 결함 아님, t2105 소유)
- 증거: `moai-base spec lint --baseline .moai/spec-lint-baseline.json SPEC-PROGRESS-RECORD-IO-001 SPEC-FACTORY-COMPLETION-RECOVERY-001 SPEC-LSPMCP-001` → `base_exit=1`, 같은 명령의 후보 `cand_exit=1`, `cmp` → `cmp_exit=0`(바이트 동일). 출력: `baseline: EXCEEDED — .moai/spec-lint-baseline.json` / `OwnershipTransitionInvalid: recorded 0 -> current 1 (+1)` / `inventory: 3 warning(s) total (advisory included), 1 non-advisory tracked across 1 recorded rule(s)` / `recorded at: d5638ba81 (2026-09-24)`.
- 판정: 위반은 1회차와 같은 모습으로 남아 있고 기존 결함이다. t2105가 소유한다. 베이스라인 JSON과 워크플로는 바뀌지 않았다(0절 diffstat).

## Codex 감사 (재감사 1회, 수리 델타 한정)
- 승인: 결정 `d-20261106T064700Z-de2d`(운영자, `d-20261106T061529Z-5d7e`를 대체, 내용 동일). 카드당 최대 2회, 재감사는 수리 델타만. 1회차 `rcpt-265f061edb178a80d58392da`에 이어 2번째 호출이다. 재시도하지 않았다.
- 넘긴 인자: `project_root=/Users/goos/.moai/worktrees/projects/moai-adk-go/t2087`, `target=baseBranch`, `delta_base=140c1ca4067df3d67619376cd75f04b7d065cab9`, `mode=native`(native만 delta를 정확히 따른다). `model`·`focus`는 넘기지 않았다.
- 결과(그대로 인용): `"verdict":"fail"`, `"audit_receipt":"rcpt-4bf94c4b99b843a5d0c8b2e4"`, `"review_base":"delta_base (merge base 140c1ca4067df3d67619376cd75f04b7d065cab9)"`, `"resolved_model":"gpt-6.1-sol"`, `"build_commit":"646955327e"`. 지적: `[P1] 인용 경로 회귀시험이 Windows에서도 실행 가능하도록 고쳐라 — internal/spec/lint_gitbatch_test.go:502 / Windows에서는 디렉터리 이름에 \`"\`를 쓸 수 없어 \`r.spec(quoted, ...)\`가 \`os.MkdirAll\`에서 실패하네. 이 시험에는 Windows 제외 처리가 없으므로 \`release-pr-multi-os.yml\`의 Windows remainder 시험과 릴리스 게이트를 막게 되네.`
- 내 검증: 정적으로 확인했다(N1). 시험 본문에 GOOS 제외가 없고, 워크플로 행렬에 windows × remainder 레그가 있다. Windows 실행으로는 확인하지 않았다.
- 해결되지 않은 Codex 차단 지적: 1건(N1).

## 반복 이력
- 1회차(`140c1ca406`): FAIL 77. F1(Windows PATH 심, 차단), F2(`diff.relative`, 차단), F3(인용 헤더, 수리 권고), F4·F5·F6 부채. 베이스라인 위반은 기존 결함으로 판정해 t2105에 넘겼다.
- 수리: `fb2c9dbb37`(F1 Skip, F2 `--no-relative`와 대체 경로 시험, F3 헤더 flush와 시험), `a6663e9417`(픽스처 유지보수 차단과 고정 시험).
- 2회차(`a6663e9417`): F1·F2·F3·플레이크 수리 확인. 수리에서 생긴 N1(Windows에서 쓸 수 없는 이름) 때문에 FAIL 84.

## Gaps (관찰하지 않은 것)
- 도구 거부: 처음 보낸 복합 명령 2개(`git archive | tar` 파이프와 셸 변수를 쓴 `go build`)를 worktree 격리 가드가 거부했다. 같은 작업을 리터럴 경로의 단순 명령(`git archive -o <tar>`, `tar -x`, `go build -o <path>`)으로 다시 실행했고 결과에 대체나 추론은 없다.
- Windows 실행: 관찰하지 않았다. F1의 Skip과 N1의 실패 모두 정적 판단이다. N1의 근거는 Windows 파일 이름 규칙(`"` 금지)과 코드, 워크플로 정의다.
- 원래 정리 플레이크는 수리 전 코드에서 재현하지 않았다. 수리 후 count=2와 count=3 통과만 관찰했다.
- 작업이 없는 분리 maintenance 실행이 `.git/objects`에 쓰는지는 측정하지 않았다.
- 실행하지 않은 검사: `./internal/spec/...` 전체(지시에 따라 제외), golangci-lint, 커버리지, 전체 말뭉치 `spec lint --baseline` 실행 시간.
- `moai verify audit-plan`과 `moai decision read`는 설치된 바이너리(MCP 서버 기준 rc.42, `646955327e`)로 실행했다. 이 트리에서 빌드한 것이 아니다(§2.2). 둘 다 판정 근거가 아니라 절차 확인용이다.

## Residual-risk (관찰했는데도 남는 위험)
- 플레이크의 두 번째 원인이 있다면 CI에서 다시 보일 수 있다(4절).
- F2 대체 경로는 `--no-relative`를 모르는 git(2.28 미만으로 추정, 측정 안 함)에서 느리지만 정확한 per-SPEC 경로로 간다. 이때 상수 호출 수 이점은 사라진다.
- 스크래치 증거(`/private/tmp/claude-501/t2087-audit2/` 아래 `repro.log`, `control.log`, `maint-*.log`, `gb3.log`, `breach-*.txt`)는 이 기기에만 있다. 판정에 쓴 줄은 위에 그대로 옮겼고, 나머지 원본은 기기 밖으로 나가면 잃는다.

## 운영 메모 (미검증)
- `inferred`: N1을 (a) 합성 입력 방식으로 고치면 F3 회귀 시험이 Windows에서도 돈다. Skip(b)보다 범위가 넓다. 근거 규칙: `parseOwnershipBatch`는 `io.Reader`를 받는 순수 함수다(207행).
- `assumption`: Codex 2회를 모두 썼으므로 N1 수리 뒤 확인 감사에 Codex 영수증이 다시 필요하면 운영자 승인(3회째)이 필요하다. 결정 본문의 "3회째는 운영자 승인" 문구에 기댄 해석이다.
