auditor-model: claude-opus-5-5

# t2087 sync 감사 판정 — spec lint git 하위 프로세스 수 상수화

verdict: FAIL
audited_sha: 140c1ca4067df3d67619376cd75f04b7d065cab9
required_backend_fail: codex

- 카드: t2087 (Class B, SPEC 없음)
- 감사 대상: 커밋 `140c1ca406` (기준 `9b632c1eb3` 위 단일 커밋, `internal/spec/` 6개 파일)
- 감사 트리: `/Users/goos/.moai/worktrees/projects/moai-adk-go/t2087`, 브랜치 `WT-spec-lint-baseline-2`
- 종합 점수: 77/100 — 필수 통과 차원(Functionality)이 차단 결함 F2로 실패하여 전체 FAIL
- Codex receipt: `rcpt-265f061edb178a80d58392da` (단 1회 호출, 재시도 없음)
- 반복 이력: 이 카드의 sync 감사 1회차. 이전 감사 없음, repair-born 해당 없음.

## 결론 요약

말뭉치(corpus) 경로의 핵심 주장은 확인됐다. 실제 SPEC 17개에서 기준 바이너리와 후보 바이너리의 출력이 바이트 단위로 같았고, git 하위 프로세스 수는 SPEC 수와 무관하게 상수였다(후보 3개→6회, 17개→7회 / 기준 17개→26회). 그러나 Codex가 낸 두 결함을 직접 검증한 결과 둘 다 성립한다. F2(`diff.relative=true` 설정 + 하위 디렉터리 실행 시 OwnershipTransitionInvalid 경고가 조용히 사라짐)는 재현했고, F1(PATH 셰 계수 시험이 Windows에서 실패해 release-pr-multi-os Windows 레그를 막음)은 Go 표준 라이브러리 소스와 워크플로 정의로 정적 확인했다. 감사 중 추가로 두 건의 판정 불일치(F3 인용 경로 헤더, F4 `_archive` 이름 변경)를 재현했다. F1·F2는 차단 결함이다.

## 차원별 점수

| 차원 | 점수 | 판정 | 근거 |
|---|---|---|---|
| Functionality (40%) | 70 | FAIL | 말뭉치 17개 SPEC 출력 `cmp exit=0`; 그러나 F2 재현(기준 `Ownership` 1건 / 후보 0건), F3·F4 재현 |
| Security (25%) | 95 | PASS | 새 git 호출은 셸 없이 `exec.Command` 인자 배열로 전달, 경로는 `batchSafePath`로 pathspec magic(`:`)·인용 대상 문자 거부, 버퍼 1 MiB 상한. 비밀값·외부 입력 없음 |
| Craft (20%) | 65 | FAIL | 동치·상수 시험은 의미 있음(아래 증거). 그러나 F1으로 Windows 릴리스 게이트에서 시험 실패. 커버리지 미측정(Gaps) |
| Consistency (15%) | 85 | PASS | `gofmt -l` 빈 출력, `go vet` exit 0, 기존 sentinel·에러 문구 유지, 주석 언어는 파일별 기존 관행 따름 |

가중 합: 0.4×70 + 0.25×95 + 0.2×65 + 0.15×85 = 77.5 → 77. 필수 통과 차원(Functionality·Security) 중 Functionality 실패로 전체 FAIL.

## 결함 목록

| ID | 심각도 | 분류 | 위치 | 내용 | 필수 수정 |
|---|---|---|---|---|---|
| F1 | High (Codex P1) | blocking, in-diff | `internal/spec/lint_gitbatch_test.go:321` 부근 `gbShim` / `TestLintGitBatch_ConstantSubprocessCount` | 확장자 없는 `#!/bin/sh` 파일 `git`을 PATH 앞에 둔다. Windows의 `exec.LookPath`는 확장자 없는 이름을 PATHEXT 확장자를 붙여서만 찾으므로(`findExecutable`: `hasExt(file)`가 거짓이면 `file + e`만 확인) 셰가 무시되고 실제 `git.exe`가 실행된다. 로그 파일이 생기지 않아 `gbCountCalls`가 0을 돌려주고 `t.Fatal("shim saw no git invocations ...")`로 실패한다. `release-pr-multi-os.yml`의 `windows-latest` × `remainder` 레그(internal/cli·rosterguard 외 전 패키지)가 이 시험을 실행하며, 이 워크플로는 모든 OS 레그가 릴리스 게이트를 막는다고 적고 있다. | 셰 시험을 Windows에서 건너뛰거나(`runtime.GOOS == "windows"` → `t.Skip`, `ac_baseline_commit_guard_test.go:123`의 기존 관행) Windows용 `.bat`/`.cmd` 셰를 함께 둔다. |
| F2 | Medium (Codex P2) | blocking, in-diff, 재현 증거 있음 | `internal/spec/lint_ownership_batch.go` `readOwnershipBatch`의 `git log` 인자 | 사용자 git 설정 `diff.relative=true`이고 lint를 하위 디렉터리에서 실행하면 패치 경로가 현재 디렉터리 기준으로 바뀐다. `byPath` 키(저장소 기준 경로)와 맞지 않아 `resolve()`가 nil을 돌려주고, `ok=true`라 per-SPEC 경로로 넘어가지도 않는다. 결과적으로 경고가 조용히 사라진다. 하위 디렉터리 실행은 지원되는 사용법이다(`TestSpecLint_IDArg_FromSubdirectory`). CI(저장소 루트 실행)는 영향 없음. | 배치 명령에 `--no-relative`를 추가하고, `diff.relative=true` + 하위 디렉터리 조건의 동치 시험을 추가한다. (`--no-relative`를 모르는 오래된 git은 배치가 실패해 per-SPEC로 돌아가므로 정답 유지.) |
| F3 | Low | optional, in-diff, 재현 증거 있음 | `lint_ownership_batch.go` `parseOwnershipBatch`의 `diff --git a/` 분기 | git이 경로를 인용하는 경우(`"`·`\`·제어 문자·탭; `core.quotepath=false`여도 이 문자들은 인용됨) 헤더가 `diff --git "a/...`로 시작해 `CutPrefix("diff --git a/")`에 걸리지 않는다. 이전 세그먼트가 닫히지 않아 인용 경로 파일의 `status:` 줄이 같은 커밋의 앞선 SPEC 세그먼트에 붙는다. 이 저장소 이력에는 인용 경로가 0건이라(`grep -c '^"'` → 0) 현재 말뭉치는 영향 없고, SPEC ID 스키마가 이런 문자를 금지하므로 도구 경로로는 생기지 않는다. | `diff --git `으로 시작하는 모든 줄에서 `flush()` 후 `inSeg=false`로 두고, 인용되지 않은 형식만 세그먼트로 연다(한 줄 수준 수정). |
| F4 | Low | optional(부채), in-diff, 재현 증거 있음 | `batchedOwnershipTransition` — root별 pathspec `:(top,glob)<root>/*/spec.md` | `.moai/specs/_archive/SPEC-X/spec.md`처럼 root 밖(`.moai/specs/SPEC-X/spec.md`)에서 이름이 바뀌어 들어온 파일은 배치에서 "추가"로 보이고, per-SPEC `--follow`는 이름 변경을 따라 더 오래된 이력으로 간다. `_archive` 6개 중 3개에서 Info 출력이 달라졌다. 말뭉치 탐색(`discoverSPECs`는 `SPEC-*` 접두 디렉터리만)과 CI 단계에는 닿지 않고, `_archive` 경로·디렉터리를 명시해 lint할 때만 생긴다. Info 등급이라 종료 코드는 같다(둘 다 exit 1). 코드 주석이 이 근사를 이미 적고 있으나 "measured on this repository"라는 문구와 달리 이 경우는 측정되지 않았다. | 수정하거나, 부채로 기록하고 주석의 측정 문구를 실제 측정 범위로 고친다. 수정 예: 이름 변경 원천이 root 밖일 수 있는 root(예: `_archive`)는 per-SPEC 경로로 보내거나 pathspec을 상위 root까지 넓힌다. |
| F5 | Info | optional, residual | 이력 단순화(history simplification) | `--follow`는 git 내부에서 가지치기를 끄므로(실측: 가지치기된 커밋 `f1208eba45`·`a5454a5051`·`ea335d1331`이 oracle 로그에 나타남) 배치(집합 단위 단순화)보다 커밋을 더 본다. 현재 저장소에서 집합 단순화로 빠지는 커밋은 8개, 영향 SPEC 3개이며 셋 모두 더 새로운 squash 커밋(`26c5a7d544`, `da791eb0ab`)이 첫 전이라 결과가 같다. 앞으로 squash 커밋에 status 변경이 없는 경우가 생기면 달라질 수 있다. | 없음(잔여 위험으로 기록). 원하면 `--full-history`로 배치를 돌려 가지치기 차이를 없앨 수 있다(병합 커밋은 `-p`에서 diff가 없으므로 세그먼트 수는 그대로). |
| F6 | Info | optional | `batchedOwnershipTransition` 실패 분기 | 배치가 실패하면(파싱 오류, 1 MiB 초과 줄, git 오류) SPEC마다 per-SPEC로 돌아가 다시 O(n)이 되는데, 아무 기록도 남기지 않는다. 정답은 유지되지만 CI 시간 초과가 조용히 재발할 수 있다. | 선택: 실패 시 한 번 진단(로그 또는 Info finding)을 남긴다. |

## 증거 (Claim / Evidence / Baseline-attribution)

모든 명령은 이 트리(`HEAD 140c1ca4067df3d67619376cd75f04b7d065cab9`)에서 실행했다. 판정 바이너리: 후보 `moai-cand`는 이 트리에서 `go build -o .../moai-cand ./cmd/moai`로, 기준 `moai-base`는 `git archive -o base.tar 9b632c1eb3` → tar 해제 → `go build -C <해제 디렉터리> -o .../moai-base ./cmd/moai`로 빌드(설치 바이너리 미사용, §2.2 충족). 스크래치: `/private/tmp/claude-501/t2087-audit/`.

### 0. 트리 확인 (hard gate)
- `git rev-parse --show-toplevel` → `/Users/goos/.moai/worktrees/projects/moai-adk-go/t2087`
- `git branch --show-current` → `WT-spec-lint-baseline-2`
- `git rev-parse HEAD` → `140c1ca4067df3d67619376cd75f04b7d065cab9`
- `git status --short` → 빈 출력. Codex 호출 직전 재확인도 빈 출력, HEAD 동일.

### 1. 상시 점검 (항목 4)
- `timeout 1200 go build ./...` → `go build exit=0`
- `timeout 900 go vet ./internal/spec/...` → `go vet exit=0`
- `gofmt -l internal/spec` → `gofmt exit=0 lines=0`
- `timeout 1500 go test -count=1 -run 'TestLintGitBatch|Ownership|ImpliedStatus|StatusGitConsistency|ConstantGitLog|Unreachable' ./internal/spec/` → `ok  	github.com/modu-ai/moai-adk/internal/spec	46.625s` (Bash 도구가 비정상 종료를 보고하지 않음 = exit 0; 파이프 없이 실행)
- 전체 `./internal/spec/...` 묶음은 지시대로 실행하지 않음(Gaps).

### 2. 판정 동치 — 실제 SPEC 17개, 기준 vs 후보 (항목 1)
- 대상: 항목 3의 3개(SPEC-PROGRESS-RECORD-IO-001, SPEC-FACTORY-COMPLETION-RECOVERY-001, SPEC-LSPMCP-001), 이름 변경 대상 2개(SPEC-V3R4-CC2X-MCP-001, SPEC-V3R4-CC2X-MSGR-001 — `git log -M --diff-filter=R -- .moai/specs`로 찾은 `R069` 2건), 이력 단순화 위험 3개(SPEC-GLM-FLASH-DEFAULT-001, SPEC-V3R6-GRAPH-FRESHNESS-001, SPEC-V3R6-GRAPH-FRESHNESS-002), 결정적 표본 9개(`ls .moai/specs | awk 'NR%140==11'`).
- `moai-base spec lint <17 paths>` → `base exit=0`; `moai-cand spec lint <17 paths>` → `cand exit=0`
- `cmp spot-base.out spot-cand.out` → `cmp exit=0` (각 54줄, stderr 0줄)
- 두 출력에 공통으로 나타난 Ownership 줄: Invalid 3건(위 3개 SPEC), Unmeasured 4건(CC2X-MCP, CC2X-MSGR, EVALLIB-001, UTIL-004).
- 저장소 수준 교차 확인: `git log --no-merges -- ':(top,glob).moai/specs/*/spec.md'` 3535개 vs `--full-history` 3543개, 차이 8개(`only-in-full-history=8 only-in-simplified=0`). 8개가 건드린 spec.md는 위 단순화 위험 3개뿐. 세 파일의 `git log -50 --follow -p` 첫 `+status:` 커밋은 `26c5a7d544`(GLM-FLASH)·`da791eb0ab`(GRAPH-FRESHNESS-001/002)이고 둘 다 단순화 목록에 있다(`simp.txt` 2119·2124행) → 첫 전이 동일.
- 구현자 주장 "227개 표본 차이 0"은 내 표본과 모순되지 않는다. 다만 `_archive`(F4)와 설정 의존 경우(F2)는 그 표본 밖이었다.

### 3. 재현된 불일치
- F4 `_archive`: `moai-base`/`moai-cand spec lint .moai/specs/_archive/<6개>/spec.md` → 둘 다 exit 1, `diff` exit 1. 기준만: `SPEC-INFRA-001`·`SPEC-ORCH-001` Unmeasured(`implemented → completed`, 커밋 `cca40d6a5e`); `SPEC-DESIGN-CONST-AMEND-001` 커밋이 기준 `4271fd8a8c` / 후보 `697a6e2c71`로 다름. 원인: `R100 .moai/specs/SPEC-X/spec.md → .moai/specs/_archive/SPEC-X/spec.md` 6건.
- F3 인용 경로: 스크래치 저장소 `qrepo`에 `SPEC-A-001`(본문만 수정)과 `SPEC-Z"Q-001`(status `draft → in-progress`, 트레일러 manager-docs)을 같은 커밋 `7c58907`에 둠. 헤더 `diff --git "a/.moai/specs/SPEC-Z\"Q-001/spec.md" ...` 확인. `spec lint .moai/specs/SPEC-A-001/spec.md`: 기준 `1 error(s), 0 warning(s)`, 후보 `1 error(s), 1 warning(s)` + `WARNING OwnershipTransitionInvalid ... SPEC SPEC-A-001 transition "draft" → "in-progress" ... commit 7c58907a57b7...` (기준에는 없음).
- F2 `diff.relative`: `qrepo`에 `SPEC-B-001`(`1cdc2dd` 생성 manager-spec, `4cc2efa` `draft → in-progress` manager-docs) 추가, `git -C qrepo config diff.relative true`, `cd qrepo/.moai/specs`에서 `spec lint SPEC-B-001/spec.md` → 기준 Ownership 1건(`WARNING OwnershipTransitionInvalid ... SPEC-B-001 ... commit 4cc2efa67f38...`), 후보 0건. 대조: `diff.relative` 해제 후 같은 위치에서 후보 → Ownership 1건. 원인이 설정임을 확인.
- F1 정적 확인: Go 1.26.8 `src/os/exec/lp_windows.go` `findExecutable` 37–57행 — 확장자 없으면 `file + e`만 검사. `.github/workflows/release-pr-multi-os.yml` 125–152행 `os: [ubuntu-latest, macos-latest, windows-latest]` × `shard: [..., remainder]`, `pkg: remainder` = internal/cli·rosterguard를 제외한 모든 패키지, 316–318행에서 `go test -json -race ...` 실행. Windows에서 직접 실행하지는 않았다(Gaps).

### 4. 하위 프로세스 수 (항목 2)
- 셰 `/private/tmp/claude-501/t2087-audit/shim/git`(호출 인자 기록 후 `/usr/bin/git` exec), `env PATH=<shim>:/usr/bin:... timeout 600 <binary> spec lint ...`
- 후보, SPEC 3개 → 6회: `rev-parse`(경로 확인) / `rev-parse --git-dir` / `rev-parse --show-toplevel --show-prefix` / `-c core.quotepath=false log -M ... -p -- :(top,glob)...` / `rev-parse --verify main` / `log main --no-merges --format=%s%x1f%B%x1e`
- 후보, SPEC 17개 → 7회: 위 6회 + `rev-parse --is-shallow-repository`(실행당 1회 캐시, 오류 형태 ②/③ 판단용)
- 기준, SPEC 17개 → 26회: `log -50 --follow` 17회, `log main --oneline --no-merges` 5회, 그 외 4회 → SPEC 수에 비례
- 결론: 후보의 git 호출 수는 SPEC 수와 무관한 상수(최대 7). `TestLintGitBatch_ConstantSubprocessCount`도 Unix에서 통과.
- 견고성 검토(코드 읽기): 배치는 `StdoutPipe`로 스트리밍, 줄 버퍼 1 MiB 상한, 파싱 오류 시 `Kill`+`Wait`, 정상 시 `Wait` — 프로세스·goroutine 누수 없음. stderr는 nil(널 장치)이라 파이프 교착 없음. 메모리는 세그먼트당 레코드만 보존(본문은 트레일러만 파싱). `cachedCommitIndex`는 실패도 캐시해 깨진 저장소에서 1회만 spawn. `getGitImpliedStatusPerSpec`과 `gitLogAllFullMessage` 모두 `%w: %s`로 `errGitQueryFailed`("git log failed")를 감싸고, `inMemImpliedStatus`의 두 소진 경로도 같은 문구와 sentinel(`errNoGitHistory`, `errNoClassifiableCommit`)을 쓴다 → `gitObservationUnreachable`/`takeUnreachableEmission` 판단 입력이 같다. git 부재 시 소유권 경로는 `cachedGitDirAvailable()` 거짓 → per-SPEC로 가서 기존 `git unreachable` 오류를 그대로 낸다. `Unreachable` 패턴 시험 통과.

### 5. 베이스라인 게이트 위반 (항목 3)
- `git diff --stat 9b632c1eb3 9cf8a1fc14 -- .moai/specs .moai/spec-lint-baseline.json internal/spec` → 빈 출력(`diffstat exit=0`). 전체 차이는 `.moai/config/sections/git-strategy.yaml`(7줄)과 `.moai/reports/t2086/sync-audit.md` 2개뿐. 즉 SPEC 말뭉치·베이스라인 파일·linter 코드가 main `9cf8a1fc14`와 기준 `9b632c1eb3`에서 같다. `git merge-base --is-ancestor 9b632c1eb3 9cf8a1fc14` → 참.
- `git diff --stat 9b632c1eb3 140c1ca406 -- .github .moai/spec-lint-baseline.json` → 빈 출력: 이 카드는 워크플로·베이스라인을 바꾸지 않았다.
- 기준 바이너리(`9b632c1eb3`에서 빌드): `moai-base spec lint --baseline .moai/spec-lint-baseline.json <3개 SPEC>` → exit 1, 출력:
  `baseline: EXCEEDED — .moai/spec-lint-baseline.json` / `OwnershipTransitionInvalid: recorded 0 -> current 1 (+1)` / `inventory: 3 warning(s) total (advisory included), 1 non-advisory tracked across 1 recorded rule(s)` / `recorded at: d5638ba81 (2026-09-24)`
- 같은 명령의 후보 출력과 `cmp exit=0`(바이트 동일).
- **판정: 이 위반은 diff와 무관하게 존재하는 기존 결함이다 — pre-existing, owned by t2105.** 이 카드의 결함이 아니다. 기준 linter가 같은 말뭉치에서 같은 3건과 같은 `+1`을 낸다. 지금까지 보이지 않았던 이유는 기준 linter의 전체 실행이 시간 안에 끝나지 않았기 때문이다.
- 비교하지 못한 것: 기준 바이너리의 전체 말뭉치 실행(끝나지 않음, 구현자 측정 600 s exit 124)이므로, 다른 SPEC에서 기준이 추가 위반을 낼지는 관찰하지 않았다. `9cf8a1fc14` 체크아웃에서 직접 돌리지 않았고 대신 diffstat으로 동일성을 확인했다(차이인 `git-strategy.yaml`은 브랜치 설정이며, OwnershipTransition은 기준 브랜치가 아닌 HEAD 이력의 `--follow`를 읽는다).

## Codex 감사 (운영자 승인 d-20261106T055816Z-449d, 카드당 1회)

- 감사 계획: `moai verify audit-plan --project-root /Users/goos/.moai/worktrees/projects/moai-adk-go/t2087` → `config_status: ok`, `codex` gate `required`, `enforced_required: ["codex"]`, claude·glm `off`.
- 호출: `mcp__moai__codex_audit` 1회, `project_root`=위 경로, `target: baseBranch`, `mode: native`, `delta_base: 9b632c1eb3`. 다른 audit 도구 호출 없음, 재시도 없음. 호출 전 다른 작업 모두 완료, 트리 변경 없음.
- 결과 결정 줄(원문): `"verdict":"fail"`, `"audit_receipt":"rcpt-265f061edb178a80d58392da"`, `"review_base":"delta_base (merge base 9b632c1eb3544cef5f275e502922a375e48d20aa)"`, `"resolved_model":"gpt-6.1-sol"`, `"build_commit":"646955327e"`
- 발견 1(원문): `[P1] Windows에서도 실행 가능한 Git 계수기를 써라 — internal/spec/lint_gitbatch_test.go:321` → F1, 직접 정적 검증함(실행은 못 함).
- 발견 2(원문): `[P2] 배치 조회에서 상대 경로 diff 설정을 해제하라 — internal/spec/lint_ownership_batch.go:180` → F2, 직접 재현함.
- Codex는 "정적 검토 결과이며 시험·빌드·재현 명령은 실행하지 않았"다고 밝혔다. 두 건의 실재 여부는 위 §3에서 내가 따로 확인했다.

## Gaps (관찰하지 않은 것)

- **도구 거부(워크트리 격리 가드)**: 다음 명령이 실행되지 않고 거부됐다 — (a) `cd ... && go test ... ; tail` 복합 명령, (b) `git archive | tar` 파이프와 빌드를 묶은 명령, (c) 변수 `$S`를 쓴 `git log`/`sort`/`comm` 명령, (d) `$(go env GOROOT)`를 쓴 `sed`. 대응: (a)–(c)는 같은 명령을 리터럴 경로의 단순 명령으로 나눠 다시 실행했다(측정 내용 동일, 대체 없음). (d)는 같은 파일을 `go env GOROOT` 출력 경로로 Read 도구로 읽었다(같은 소스, 대체 없음).
- F1을 Windows에서 실행하지 않았다(정적 근거만).
- 전체 `./internal/spec/...` 시험 묶음, 커버리지 측정(`go test -cover`), `golangci-lint`를 실행하지 않았다.
- 기준 바이너리의 전체 말뭉치 실행(끝나지 않음). 따라서 전체 말뭉치 출력의 기준/후보 바이트 동일성은 관찰하지 않았다. 함수 수준 동치는 내 표본 17개 + 저장소 수준 단순화 분석 + 구현자 기록(검증하지 않은 주장으로 취급)에 기댄다.
- 얕은 클론(shallow clone)에서의 새 경로는 기존 `Unreachable` 시험 통과로만 확인했다.
- Linux/CI 러너, CI 소요 시간은 측정하지 않았다. 부하(load 55–62) 아래 측정이라 시간 값은 기록하지 않았다.
- `grep.patternType` 같은 사용자 git 설정의 영향은 이 저장소 설정(미설정)에서만 보았다.

## Residual-risk (관찰했음에도 남는 위험)

- F5: squash 커밋에 status 변경이 없는 이력이 생기면 배치와 `--follow`가 다른 커밋을 첫 전이로 고를 수 있다.
- 사용자 diff/log 설정은 배치가 `--no-color --no-ext-diff --src-prefix --dst-prefix -c core.quotepath=false`로 고정하지만 `diff.relative`처럼 고정하지 않은 다른 설정(예: `log.showSignature`, `diff.noprefix` 외)이 더 있을 수 있다.
- F6: 배치 실패가 조용히 O(n) 경로로 돌아가므로 시간 초과가 다시 나타나도 원인이 바로 보이지 않는다.
- 기준 전체 실행이 끝나지 않으므로 `+1` 외에 다른 기존 위반이 있는지는 후보 실행 결과에만 기댄다.

## 다음 조치 (차단 결함 수리 범위)

1. F1: 셰 계수 시험을 Windows에서 건너뛰거나 Windows용 셰를 추가한다.
2. F2: 배치 `git log`에 `--no-relative` 추가 + `diff.relative=true`·하위 디렉터리 동치 시험 추가.
3. 같은 수리에서 F3(한 줄 수준)도 함께 고치기를 권한다. F4는 수정 또는 부채 기록 + 주석의 측정 문구 정정. F5·F6은 선택.
4. 재감사 범위: F1·F2(및 함께 고친 경우 F3·F4)의 수정 diff로 한정.

## Operational Notes (unverified)

- `inferred`: 이 카드의 목적(CI 단계 완료)은 후보에서 달성된다 — 구현자 기록상 전체 실행 51.56 s, exit 1(베이스라인 위반은 t2105 소유). 측정 명령: `timeout -k 10 600 <후보> spec lint --baseline .moai/spec-lint-baseline.json`. 나는 전체 실행 시간을 직접 재지 않았다.
