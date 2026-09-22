# progress.md — SPEC-CODEX-LOCALMD-001

## §E.1 Plan-phase Audit-Ready Signal

- plan_phase: 2026-09-22, manager-spec, tree WT-codex-local-md@1e00e35f8
- artifacts: spec.md + plan.md + acceptance.md + research.md + spec-compact.md + decision-index.md + progress.md (Tier M set, frontmatter `tier: "M"`)
- spec_id_check: PASS (`SPEC-CODEX-LOCALMD-001`, Bash regex, 이번 실행)
- historical_audit: iteration-1 FAIL 0.774 → iteration-2 PASS 1.0. 이 PASS는 구현 전 재검증과 1.0.1 보완 전 baseline에만 적용한다.
- recheck_amendment: parser 5형태, spawn quote 팽창, 양쪽 safe open/fstat, body-slice hash, 입력/문서 경계, 플랫폼 gate, help 역할 구분, 120초 exact-command LIVE predicate를 반영함.
- decision_rows: R1~R4 RESOLVED — 카드 전체 구현·로컬 병합 요청에 따른 구현 판단이며 개별 사용자 승인 기록이 아님.
- plan_complete_at: 2026-09-22 (orchestrator, plan lane t1078)
- re_audit: iteration-3 FAIL 0.9375 (`.moai/reports/t1078/plan-audit-iter3.md`) — Out of Scope 한 문장의 run/sync 경계 모순 1건을 수정함.
- plan_status: PASS 1.0000 (iteration-4 delta 감사, `.moai/reports/t1078/plan-audit-iter4.md`; strict lint `[]`, diff-check exit 0, 동결 6문서 시작·종료 해시 동일). 이 판정은 구현·LIVE 수용 통과를 주장하지 않는다.

## §E.2 Run-phase Evidence

### Claim

2026-09-22 독립 구현 담당이 아래 로컬 구현과 scoped 검증을 완료했다. 실제 인증 LIVE와 Windows 런타임은 수행하지 않았으므로 카드 전체 완료 PASS는 아니다.

- 두 입력을 `CLAUDE.local.md` → `AGENTS.local.md` 순으로 source header와 함께 합성하고 단일 override로 전달한다.
- Unix는 `O_NOFOLLOW|O_NONBLOCK|O_CLOEXEC` open 후 같은 descriptor를 fstat/read한다. Windows는 reparse-point handle을 열어 거부한 뒤 같은 handle을 fstat/read한다. 기타 플랫폼은 존재하는 입력을 fail closed한다.
- operator config 다섯 표기 충돌을 합성값이 있을 때 거부한다. direct 최종 토큰과 spawn 최종 명령의 126,976-byte 상한을 독립 적용한다.
- help와 템플릿 일반 문구를 정합했다. exactly-3 table은 확장하지 않았다.
- bare/cli/app/spawn/worktree/factory lead/agent/agent-2/lane-3를 capture로 확인했고 입력 네 파일의 hash·file identity·비-symlink 상태와 rename seam count 0을 단언했다.

### Baseline-attribution

실행 tree: `/Users/goos/MoAI/moai-adk-go/.claude/worktrees/t1078`, 시작/종료 HEAD `07f3b78d448153360804691bc2b682ce17a83d90`, branch `WT-codex-local-md`. 시작 `git status --short` 출력 없음. 구현 코드·테스트와 이 run evidence만 편집했고 canonical spec/plan/acceptance, 감사 보고서, 다른 카드 변경은 편집하지 않았다. 커밋·브랜치 변경·remote 작업 없음.

요청된 `.agents/skills/moai-workflow-tdd/SKILL.md`는 WT에 없었다. 검색으로 찾은 `.claude/skills/moai-workflow-tdd/SKILL.md` 전체를 읽고 RED→GREEN→REFACTOR를 적용했다. 구조/기술 문서의 CLI/config 경계, 기존 x/sys 의존성, 테스트/플랫폼 규칙도 판독했다.

### Evidence — RED

모든 go 명령의 cwd는 위 WT이고 환경은 단일 invocation에서 `unset CLAUDECODE CODEX_THREAD_ID && GOCACHE=/tmp/t1078-spec-audit-cache`로 지정했다.

M1: `go test ./internal/cli -run '^TestCodexLocalInstructions_(DualFileMatrix|LargeBodySlicesAndFreshRead|InjectedAsDeveloperInstructions|DocumentedInLauncherHelp|ExactlyThreeContractPaths)$' -count=1`.

관측 stdout 발췌(생략된 다른 실패를 통과로 간주하지 않음), exit 1:

```text
--- FAIL: TestCodexLocalInstructions_DocumentedInLauncherHelp (0.00s)
    codex_local_instructions_test.go:146: launcher help does not mention "CLAUDE.local.md"
    codex_local_instructions_test.go:146: launcher help does not mention "shared with Claude"
    codex_local_instructions_test.go:146: launcher help does not mention "non-empty"
    --- FAIL: TestCodexLocalInstructions_DualFileMatrix/body/absent (0.00s)
        codex_local_instructions_test.go:201: want exactly one override pair, got 0 args
--- FAIL: TestCodexLocalInstructions_LargeBodySlicesAndFreshRead (0.00s)
    codex_local_instructions_test.go:231: missing first provenance
FAIL
FAIL	github.com/modu-ai/moai-adk/internal/cli	0.796s
FAIL
```

Exactly-three 테스트는 기존 invariant를 확인하므로 첫 실행부터 PASS였다. 이를 신규 동작의 RED로 주장하지 않는다.

M2: `go test ./internal/cli -run '^TestCodexLocalInstructions_(OperatorCollision|SizeBoundaries|SpawnQuoteExpansion|BothFilesRejectNonRegular)$' -count=1`, exit 1. 다섯 collision 표기 모두 `err=<nil> launches=1`로 실패했다. 길이/quote 관측 출력:

```text
--- FAIL: TestCodexLocalInstructions_SizeBoundaries (0.00s)
    --- FAIL: TestCodexLocalInstructions_SizeBoundaries/spawn=false/delta=1 (0.00s)
        codex_local_instructions_test.go:309: overflow: <nil>, launches=1
    --- FAIL: TestCodexLocalInstructions_SizeBoundaries/spawn=true/delta=1 (0.00s)
        codex_local_instructions_test.go:309: overflow: <nil>, launches=1
--- FAIL: TestCodexLocalInstructions_SpawnQuoteExpansion (0.00s)
    codex_local_instructions_test.go:324: direct token=40069 bytes; final spawn=160125 bytes
    codex_local_instructions_test.go:331: spawn overflow: <nil> launches=2
FAIL
FAIL	github.com/modu-ai/moai-adk/internal/cli	0.828s
FAIL
```

안전 판독 seam 테스트는 구현 전 작성했으며 `go test ./internal/cli -run '^TestCodexLocalInstructions_(OpenDescriptorSurvivesReplacement|DescriptorFailuresFailClosed)$' -count=1`의 최초 결과는 `undefined: codexOpenLocalFileFn`, `undefined: openCodexLocalFile`, `[build failed]`, exit 1이었다. 이는 API 부재의 compile RED이며 실제 취약점 재현이라고 주장하지 않는다. 구현 후 양 파일의 descriptor 교체, open/stat/read 실패, symlink/FIFO/socket/device/permission/directory 셀을 실행했다.

템플릿: `go test ./internal/cli -run '^TestCodexLocalInstructions_TemplateDescribesCommonAndSpecificInputs$' -count=1`, exit 1:

```text
--- FAIL: TestCodexLocalInstructions_TemplateDescribesCommonAndSpecificInputs (0.00s)
    codex_local_instructions_test.go:391: template must describe common and Codex-specific inputs without enumerating local filenames
FAIL
FAIL	github.com/modu-ai/moai-adk/internal/cli	0.804s
FAIL
```

### Evidence — GREEN / REFACTOR

로컬 판독을 공통 descriptor 검증과 작은 플랫폼 open helper로 나눠 구현했다. 기존 라이브러리 x/sys만 사용했고 새 의존성은 없다. lint의 close errcheck와 오류 문자열 소문자 문제를 수정했다. 추가 추상화는 만들지 않았다.

고정 selector `TestCodex(LocalInstructions|InstructionContract|CLI|Spawn|Direct|Child|Worktree|Launcher|Propagate|Verb)`로 변경 전과 변경 후를 측정했다.

```sh
go test ./internal/cli -run 'TestCodex(LocalInstructions|InstructionContract|CLI|Spawn|Direct|Child|Worktree|Launcher|Propagate|Verb)' -count=1 -coverprofile=/tmp/t1078-codex-before.cover
```

```text
ok  	github.com/modu-ai/moai-adk/internal/cli	0.629s	coverage: 7.0% of statements
```

exit 0, 구현 전 baseline. 같은 selector의 구현 후 `/tmp/t1078-codex-after.cover`:

```text
ok  	github.com/modu-ai/moai-adk/internal/cli	2.375s	coverage: 7.7% of statements
```

exit 0. `go tool cover -func=/tmp/t1078-codex-before.cover`의 total은 `6.9%`, after의 total은 `7.8%`였다. go test 출력과 go tool 출력의 수치를 서로 바꿔 인용하지 않는다. 같은 도구끼리 비교하면 모두 증가했다. 이는 scoped selector가 측정한 패키지 수치이지 전체 테스트 패키지 커버리지 목표 충족 주장이 아니다.

`go tool cover -func=/tmp/t1078-codex-after.cover`의 변경 함수 출력:

```text
codexLocalDeveloperInstructionArgs 100.0%
codexHasDeveloperOverride          100.0%
checkCodexInstructionSize          100.0%
defaultCodexSpawnLaunch            100.0%
runCodexLaunch                     96.4%
codexDirectLaunch                  100.0%
codexSpawnLaunch                   100.0%
readCodexLocalInstruction          100.0%
openCodexLocalFile (unix)          100.0%
```

원본 go tool 출력의 함수명과 퍼센트 열을 표기했다. Windows/unsupported 전용 함수는 현재 호스트 coverprofile에 없으며 85%를 측정했다고 주장하지 않는다.

```sh
go test -race ./internal/cli -run 'TestCodex(LocalInstructions|InstructionContract|CLI|Spawn|Direct|Child|Worktree|Launcher|Propagate|Verb)' -count=1
```

```text
ok  	github.com/modu-ai/moai-adk/internal/cli	3.733s
```

exit 0. `go test ./internal/cli -run '^TestCodexLocal' -count=1 -v`도 실제 테스트를 실행해 `PASS`, `ok github.com/modu-ai/moai-adk/internal/cli 2.309s`를 출력했다.

```sh
GOLANGCI_LINT_CACHE=/tmp/t1078-golangci-cache golangci-lint run ./internal/cli/...
```

```text
0 issues.
```

exit 0. `git diff --check` 출력 없음, exit 0.

```sh
GOOS=windows GOARCH=amd64 go build ./...
```

exit 0. 관측 경고(성공과 별도로 보존):

```text
go: writing stat cache: open /Users/goos/go/pkg/mod/cache/download/github.com/modu-ai/moai-adk/@v/v0.0.0-20260910025036-2213871afb7d.info767416723.tmp: operation not permitted
```

`GOOS=windows GOARCH=amd64 go test -c ./internal/cli -o /tmp/t1078-cli-windows.test.exe`: stdout 없음, exit 0. `go list -f '{{range .GoFiles}}{{println .}}{{end}}' ./internal/cli`에서 Windows 대상은 `codex_local_file.go`+`codex_local_file_windows.go`, 호스트 대상은 `codex_local_file.go`+`codex_local_file_unix.go`가 선택됨을 확인했다.

### Gaps

- AC-LMD-012 인증 LIVE는 세 차례 경계를 관측했다. built `/tmp/t1078-moai-live` SHA-256은 `e0116af703c574e8023a639a2c5e4785aa69434b0a5d3eb8cb10c02a0ead596c`다. 1차는 fixture wiring 부재로 exit 130, 동일 binary init 뒤 2차 sandbox 실행은 app-server 권한 오류로 exit 1이었다. sandbox 밖 재실행은 두 fixture 로컬 지침의 외부 Codex 서비스 전송에 대한 명시 승인이 없어 승인 단계에서 거부됐다. 실제 응답/session log 증거가 없으므로 GAP/UNVERIFIED이며 카드 전체 PASS/완료로 판정하지 않는다.
- Windows 실제 런타임과 symlink 권한 환경, unsupported 플랫폼 런타임/coverage는 관측하지 않았다. cross-build는 runtime 증거가 아니다.
- Unix character-device 셀은 권한 없는 mknod 대신 open seam이 실제 `/dev/null` descriptor를 반환하여 production fstat 거부를 확인했다. 실제 파일명에 device node를 생성했다는 주장은 하지 않는다.
- 입력 불변은 byte hash, SameFile, non-symlink 및 rename seam count로 측정했다. OS 전체의 write/link syscall tracing은 수행하지 않았다.
- 첫 baseline의 넓은 `-run TestCodex`는 기존 `TestCodexLive_ReviewStartBaseBranchIsNotRejected`를 포함했고 `handshake failed: codex stdout closed before response to id=1`, exit 1이었다. 기존 테스트가 `.moai/state/verify/t399-live/basebranch-roundtrip.ndjson`을 기록했다. 구현 전 발생한 외부 진단 실패이며 구현 실패로 귀속하지 않는다. 해당 파일은 삭제하지 않았다. 이후 selector를 고정했다.
- 최초 all-funnel 테스트에서 factory registry의 사용자 홈 쓰기가 sandbox에 막혔다. 기존 kanban 테스트와 같은 `MOAI_HOME=t.TempDir()` 격리 후 GREEN을 관측했다. 사용자 홈의 registry를 변경하지 않았다.

### Residual-risk

실제 Codex의 instruction 도달성은 LIVE를 통해 별도 판정해야 한다. Windows 및 기타 플랫폼은 컴파일 이상의 증거가 없으므로 CI/runtime 검증이 남는다. 새로운 안전 reader는 leaf symlink를 따르지 않고 열린 descriptor를 유지하지만, 소스 파일의 동시 in-place 내용 변경에 대한 snapshot 격리는 요구하지도 구현하지도 않았다. 원래 launch error 전달 경로와 init gate의 기존 계약은 유지했다.

## §E.3 Run-phase Audit-Ready Signal

- local_implementation: ready-for-independent-code-review
- scoped_tests: PASS (일반/race), host changed-function coverage: 96.4%–100%, lint: PASS
- windows_cross_build: PASS; windows_runtime: NOT_RUN
- live_acceptance: GAP / UNVERIFIED (fixture init 및 sandbox 실행은 시도했으나 실제 Codex 응답 전 차단; 외부 실행 명시 승인 필요)
- commit: NOT_CREATED; integration: NOT_RUN

## §E.4 Sync-phase Audit-Ready Signal

- independent_audit: 코드 findings 0건; AC-LMD-001~011 PASS; AC-LMD-012 HARD GAP 때문에 overall FAIL/un-PASS
- audit_report: `.moai/reports/t1078/sync-audit.md`
- parent_remeasurement: scoped 일반 `2.261s`, race `3.705s`, 모두 exit 0; `git diff --check` 출력 없음
- integration: NOT_RUN
