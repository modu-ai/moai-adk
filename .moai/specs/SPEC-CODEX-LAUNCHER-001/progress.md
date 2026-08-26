# SPEC-CODEX-LAUNCHER-001 — progress

## §E.1 Plan-phase Audit-Ready Signal

plan_status: audit-ready
plan_complete_at: 2026-08-25

plan-phase 는 8차 반복으로 닫혔다. 궤적과 각 라운드의 판정서는 `.moai/reports/t197/` (`verdict-iter3` ~ `verdict-iter8`, `verdict-init-1` ~ `verdict-init-4`), 규칙 적용 기록은 같은 디렉터리 `rules-applied.md`.

마지막 확정 감사(`verdict-iter8.md`)는 FAIL 0.82 였고, 그 시점의 잔여 지적 중 **모순 3건** 은 착수 전에 정리했으며(운영자 판정) 나머지는 run-phase 부채로 인계됐다. 부채 목록은 아래 §E.2 에 기록한다.

mutant 표: 라운드 시작 시점 16/16 MUTANT-WRITABLE → 현재 13/16 MUTANT-FREE.

## §F Phase 4 Mode Selection

입력 파라미터:

| 항목 | 값 |
|---|---|
| tier | M |
| scope (파일 수) | 신규 2 + 수정 1 = 3 (`codex_launcher.go` · `codex_readiness.go` · `mcp_codex.go`) + 시험 파일 |
| domain count | 1 (Go — `internal/cli`) |
| file language mix | 100% Go |
| concurrency benefit | LOW — 코딩 중심 작업 |

모드 평가:

| 모드 | 선택 | 사유 |
|---|---|---|
| `direct` | 아니오 | 사소하지 않다 — 신규 커맨드 표면 + 분류기 재설계 |
| `serial` | **선택** | 코딩 중심 + 단일 도메인. Anthropic 의 coding-task 병렬성 유보에 따라 순차가 기본 |
| `fanout` | 아니오 | 다도메인이 아니고 연구 중심도 아니다 |
| `sweep` | 아니오 | 기계적 대량 변환이 아니다 — 새 코드 작성이다 |

Decision: serial

정당화: 단일 Go 패키지에 새 코드를 쓰는 작업이라 병렬화 가능한 독립 단위가 사실상 없다. M1 은 독립적 가치가 있고(웹 콘솔·MCP 도구의 auth 오표시가 그 자체로 해소된다) M2→M3 는 순차 의존, M4 는 마지막이므로 마일스톤 단위 순차 위임이 그대로 작업 구조와 일치한다. Implementation Kickoff Approval 은 통과했다 (운영자 판정, 리드 전달).

## §E.2 Run-phase Evidence

(마일스톤별로 아래에 append 한다.)

### 인계된 run-phase 부채

확정 감사가 남긴 지적 중 모순이 아닌 것들. 구현 중 닫히는 것과 못 닫는 것을 갈라 기록한다 — 못 닫으면 sync 로 넘기지 않고 그 시점에 보고한다 (리드 지침).

| id | AC | 내용 |
|---|---|---|
| D2 | AC-CL-008 · AC-CL-010 | REQ-CL-008 이 요구한 "거부된 `auth.json` → 명령 프로브 하강" 을 어느 AC 도 단언하지 않는다. 통합 3칸이 (유효 파일 / 파일 부재) 뿐이라 `readCodexAuthFile` 이 `ok=false` 를 돌려줄 때 러너를 부르지 않는 구현이 전 AC 를 통과한다 |
| D4 | AC-CL-012 | 무쓰기 관측 범위가 여전히 열거(격리 홈 트리 전체). `os.TempDir()` 을 거치지 않는 하드코딩 절대 경로 쓰기가 관측되지 않는다 |
| D5 | AC-CL-009 | 산문의 케이스 수(11)가 표의 데이터 행(13)과 어긋난다 — 단언은 표에 연동돼 있어 약화는 없다 |
| D6 | REQ-CL-004 vs AC-CL-011 | REQ 는 5행을 열거하고 AC 는 6라벨 폐집합을 요구한다 |
| D7 | plan §C.2 vs AC-CL-009 | 두 문서의 참조 문법이 다르다 (acceptance 판이 구속력을 가진다) |
| D8 | plan §C.3 | `CODEX_HOME` 부재 조치(`codex login`)를 어느 AC 도 판정하지 않는다 |

### 실측해야 할 것

- `internal/cli` 패키지 실행 시간 — 단독 실측 336초 기준선. 이 SPEC 이 시험을 얹으므로 M 단위로 재실측하고, `-timeout 1200s` 로도 모자라면 그 자리에서 상향해 기록한다.
- tty 왕복 — CI 에서 관측 불가. `os.Stdin`/`os.Stdout`/`os.Stderr` 값 항등만 단언하고 왕복은 Gap 으로 남는다. 실제로 깨지는 것이 관측되면 빌드 태그 문제가 운영자에게 되돌아간다 (spec.md 「판정 제외」).

### 중단된 위임 1건 (ledger 닫기)

M1 의 첫 `manager-develop` 위임이 사전 점검 단계에서 **중단** 됐다 — 반환값 없음, 사유는 계정 사용량 한도(작업 내용과 무관). 관측: 중단 직후 `git status --short` 에 그 위임이 만든 변경 0건, `internal/cli/` 에 신규 파일 0건, HEAD 무변동(`92987e653`). 즉 **아무것도 쓰지 않았고 재작업 대상도 없다.**

이 항목을 남기는 이유: 중단은 blocker report 와 다르다. blocker 는 돌아온 것이고 중단은 돌아오지 않은 것이라, 기록이 없으면 다음 읽는 사람이 "M1 이 한 번 돌았는데 결과가 없다" 로 읽는다.

---

## M1 — auth 2단 사다리: **완결** (2026-08-26)

리드의 마감 지시로 GREEN 도중 위임을 정지시켰다. 억지로 완성하지 않고 상태를 그대로 적는다. **아래 근거는 전부 내가(오케스트레이터) 직접 실행해 관측한 것** 이다 — 정지된 위임의 자기보고가 아니다.

핀: `fd92ecf58` + 아래 미커밋 변경. 트리: `internal/cli/mcp_codex.go` 수정 1건, 신규 시험 1건, testdata fixture 3건.

### 어디까지 됐나

설계(plan §C.2)의 5개 seam 이 전부 들어갔다 — `classifyCodexAuthFile`(순수 파일 판정) · `readCodexAuthFile`(경로를 아는 유일한 층) · `codexLoginStatusRunner`(stdout/stderr/exitCode **분리** 반환) · `combineCodexStreams` · `parseCodexAuthLine`(전체 행 문법). `classifyCodexAuth` 는 얇은 조립부로 재작성됐다. 신규 시험 13개 함수 / 74 서브테스트.

### 관측한 것

| 항목 | 명령 | 관측 |
|---|---|---|
| 빌드 | `go build ./...` | rc 0 |
| vet | `go vet ./internal/cli/` | rc 0 |
| 크로스 플랫폼 vet | `GOOS=windows go vet ./internal/cli/` | rc 0 |
| 포맷 | `gofmt -l <두 파일>` | 출력 없음 |
| OS 빌드 태그 0건 (AC-CL-014) | `grep -c 'go:build' codex_auth_ladder_test.go` | `0` |
| `syscall` import 0건 (AC-CL-014) | `grep -c '"syscall"' <두 파일>` | 각 `0` |
| 신규 시험 | `go test ./internal/cli/ -v -run '<13개 이름 정규식>' -timeout 540s` | **74 RUN / 73 PASS / 1 FAIL**, 8.857s |

### 미해결 결함 1건 — 기존 시험 `TestClassifyCodexAuth_Branches` 가 깨진다

```
mcp_codex_test.go:475: classifyCodexAuth("Auth mode: API key (sk-...)") = "chatgpt", want "apiKey"
mcp_codex_test.go:475: classifyCodexAuth("Configured custom provider 'foo'") = "chatgpt", want "provider"
mcp_codex_test.go:475: classifyCodexAuth("") = "chatgpt", want "unknown"
mcp_codex_test.go:475: classifyCodexAuth("something unrecognized") = "chatgpt", want "unknown"
mcp_codex_test.go:482: classifyCodexAuth on runner error = "chatgpt", want "unknown"
```

원인은 **두 개** 이고 둘 다 실재한다:

1. **그 시험은 이 SPEC 이 의도적으로 폐기한 계약을 인코딩한다.** `"Auth mode: API key (sk-...)"` → `apiKey` 는 부분 일치 시절의 기대값이고, 전체 행 문법은 이것을 `unknown` 으로 내리는 것이 **맞다**(REQ-CL-009 · AC-CL-009). 시험을 새 계약으로 갱신해야 한다 — AC-CL-015 가 함수 **삭제·개명** 0건을 요구하므로 제자리 수정이다.
2. **더 나쁜 쪽 — 그 시험은 `CODEX_HOME` 을 격리하지 않는다.** 옛 seam(`codexRunner`)만 스텁하므로 새 1단이 개발자의 **실제 `~/.codex/auth.json`**(이 머신은 `auth_mode=chatgpt`)을 읽고 단락한다. 그래서 빈 입력에도 `chatgpt` 가 나온다 — 다섯 줄이 전부 `chatgpt` 인 것이 그 증거다. 계약을 갱신해도 격리 없이는 **머신 상태에 의존하는 시험** 으로 남는다. 다음 세션은 둘 다 고쳐야 한다.

### 이 관측 자체가 이 라운드 규칙의 사례다

처음엔 좁은 이름 정규식(`-run 'TestCodex(Auth|LoginStatus|...)'`)으로 돌려 **rc 0** 을 받았다. 그 정규식이 깨지는 시험을 고르지 않았을 뿐이다. 13개 함수 이름을 명시하고 `-v` 로 서브테스트 수를 세고 나서야 FAIL 이 드러났다. **이름 정규식의 rc 0 은 "통과" 가 아니라 "그 정규식이 고른 것들이 통과" 다** — plan §F 를 패키지 전체 실행으로 바꾼 이유가 같은 자리에서 한 번 더 확인됐다.

### 하지 않은 것 (Gaps)

- **RED 선행 출력 미보존.** 정지된 위임이 "RED-2 captured" 라고 했으나 progress.md 에 옮기기 전에 죽었다. 즉 **이 문서에 RED 근거는 없다.** AC-CL-008 의 통합 케이스가 수정 전 트리에서 실패해야 한다는 요구는 다음 세션에서 재관측해야 한다 (`git stash` 없이 `git show fd92ecf58:internal/cli/mcp_codex.go` 로 이전 판을 되살려 확인하는 편이 안전하다 — stash 는 저장소 전역이라 다른 레인 것을 삼킨다).
- **커버리지 미측정** — 순수 함수 3종 100% (AC-CL-015) 미확인.
- **lint 미실행** — `golangci-lint run` 안 돌렸다.
- **패키지 전체 실행 시간 미측정** — 신규 시험만 8.857초로 쟀을 뿐, `internal/cli` 전체(단독 336초 기준선)에 얹은 뒤의 값은 **미측정** 이다. 따라서 **180칸 봉쇄 행렬(INIT AC-CI-011)의 실행 시간도 미측정** — 그 SPEC 은 착수조차 하지 않았다.
- **부채 D2** — 거부된 `auth.json` → 명령 프로브 하강. 구현은 들어갔고 `TestClassifyCodexAuth_RejectedAuthFileFallsBackToProbe`(3 서브테스트: 빈 토큰 객체 / 미지 모드 / 파싱 실패)가 통과하는 것을 관측했다. **닫힌 것으로 본다** — 다만 그 시험이 AC 문면에 편입되지는 않았으므로 AC 자체의 구멍은 남는다.
- **부채 D4 · D5 · D6 · D7 · D8** — 손대지 않았다. M1 범위 밖이다.

---

## M1 완결 — 2026-08-26 (두 번째 위임, lane-13 이어받기)

리드의 이어받기 판정(verdict.md, decision=RESUME)에 따라 정지됐던 M1을 완결했다. 첫 위임이 남긴 미커밋 diff 3파일(+62/−16: `codex_auth_ladder_test.go` +23 · `mcp_codex.go` +12/−4 · `mcp_codex_test.go` +43/−16)을 전수 리뷰한 결과 **계약 갱신의 정확한 시작부**로 판정 — 레거시 시험의 기대값을 전체 행 문법 계약(AC-CL-009 표와 정확 일치)으로 바꾸고 `t.Setenv("CODEX_HOME", t.TempDir())` 격리를 얹었으며, 폐기한 hunk는 0건, 그대로 이어받았다. `resolveCodexHomeDir`가 `codexHomeEnvVar` 상수로 env를 읽으므로 격리가 stage 1에 실제로 먹힌다. 이 문서의 모든 근거는 이번 세션이 직접 실행해 관측한 것이다.

### E8 — RED 재관측 (뮤턴트 방식, AC-CL-008 기준선 근거)

WIP가 미보존한 RED 를 재관측했다. 수정 전 트리(`fd92ecf58`)의 판정 규칙 — `strings.Contains` 3-way 부분 일치 — 를 `parseCodexAuthLine`에 임시로 심는 뮤턴트를 만들고 갱신된 시험을 돌렸다:

- 뮤턴트: `low := strings.ToLower(string(combined)); switch { case strings.Contains(low, "chatgpt"): ... case strings.Contains(low, "api key"), strings.Contains(low, "apikey"): ... }`
- 명령: `go test ./internal/cli/ -v -run 'TestClassifyCodexAuth_Branches|TestCodexSetup_GoProbeNoNodeBridge|TestCodexAuthLadder_PinnedSkipNamesAreTheFixtureCallers' -count=1 -timeout 300s`
- 관측 (rc 1):

```
--- FAIL: TestClassifyCodexAuth_Branches (0.00s)
    mcp_codex_test.go:493: classifyCodexAuth("Logged in to ChatGPT") = "chatgpt", want "unknown"
    mcp_codex_test.go:493: classifyCodexAuth("Auth mode: API key (sk-...)") = "apiKey", want "unknown"
FAIL
FAIL	github.com/modu-ai/moai-adk/internal/cli	1.146s
```

- 뮤턴트 원복 후 동일 명령: `--- PASS: TestClassifyCodexAuthLadder_PinnedSkipNamesAreTheFixtureCallers` / `--- PASS: TestCodexSetup_GoProbeNoNodeBridge` / `--- PASS: TestClassifyCodexAuth_Branches`, `ok ... 1.153s` (rc 0). 원복 확인: `grep -c MUTANT internal/cli/mcp_codex.go` → 0, `git diff --stat` → 상속 diff 와 동일한 3파일 +62/−16.

뮤턴트가 갱신된 시험의 정확히 두 칸(부분 일치가 오류 문구를 인증 성공으로 읽는 칸)을 잡았다 — 시험이 옛 계약 인코딩이 아니라 새 계약을 판정한다는 증거.

### E1 — M1 소관 AC 관측 (전 패키지 실행 + 회귀축)

- 명령: `go test ./internal/cli/ -coverprofile=/tmp/t197_cover.out -count=1 -timeout 1100s` (백그라운드)
- 관측: `ok  github.com/modu-ai/moai-adk/internal/cli	259.848s	coverage: 78.9% of statements` (rc 0) — 74 서브테스트 전수 초록. 단독 기준선 336s 대비 259.8s.

| AC | 상태 | 근거 |
|---|---|---|
| AC-CL-008 (1단 표 + 통합 3칸) | PASS | `TestClassifyCodexAuthFile_Table` · `TestClassifyCodexAuth_LadderIntegration` · `TestReadCodexAuthFile_NoSentinelLeak` · `TestCodexAuthFileTypes_NoSecretRetention` — 전 패키지 실행 안에서 통과. 통합 2번째 칸(stderr-only)의 RED 는 위 뮤턴트 관측으로 뒷받침 |
| AC-CL-009 (파서 표 + 정규화축 + 속성) | PASS | `TestParseCodexAuthLine_Table` · `_NormalizationAxis` · `_PropertyEquivalence` 통과 |
| AC-CL-015 (공유 러너 무회귀) | PASS (관측 상세 아래) | 회귀축 실행 + 이름 대조 + 순수함수 커버리지 |

**회귀축 (AC-CL-015)**: `go test ./internal/cli/ -json -run Codex -count=1 -timeout 1100s` → rc 0, 최상위 통과 시험 함수 118개(그 안에 `TestClassifyCodexAuth_Branches` 포함 — 삭제·개명 0건). 이름 목록 대조는 실행이 아니라 소스에서 했다: `git grep -o -E "^func Test[A-Za-z0-9_]*Codex[A-Za-z0-9_]*" fd92ecf58 -- internal/cli` 와 현재 트리의 동일 grep 을 diff → **추가 14건(WIP 13 + pin 1), 삭제·개명 0건** (AC: 새 이름 추가 허용).

**skip 5건의 정체 (숨기지 않고 기록)**: 회귀축 `-json` 에서 `"Action":"skip"` 5건 — 전부 `TestCodexLive_*` (`ThreadReuseAndTurnInterrupt` · `SandboxPolicyStickiness` · `OmittedSandboxPolicyBaseline` · `ExplicitReadOnlyApprovalStall` · `ReviewStartEmitsTurnStarted`). 이들은 `codex_live_protocol_probe_test.go` 의 opt-in 게이트(`probeLiveEnv != 1 — live protocol probe is opt-in (it spends real codex quota)`)로 skip 된 **pre-existing 라이브 프로브**다. 해당 파일은 SPEC-CODEX-PHASE2-001(커밋 `ac37c4aea`) 소유이고 이번 diff 는 비접촉. 이 SPEC 이 추가한 시험의 skip 은 0이고 macOS 실행에서 GOOS skip 3칸도 발동하지 않았다. AC-CL-015 문면("skip 0, fixture 3칸 제외")의 의도 — 이 SPEC 의 수정이 깨지는 시험에 skip 을 붙여 숨기는 것 방지 — 에 비추어 이 5건은 이 SPEC 소관이 아니므로 판정에 반영하지 않되, sync 감사관이 다르게 읽을 수 있어 여기에 전문으로 남긴다.

### E2 — 크로스 플랫폼

- `go build ./...` → rc 0 (출력 없음)
- `go vet ./internal/cli/` → rc 0 (출력 없음)
- `GOOS=windows go vet ./internal/cli/` → rc 0 (출력 없음)
- `gofmt -l internal/cli/mcp_codex.go internal/cli/codex_auth_ladder_test.go internal/cli/mcp_codex_test.go` → 출력 없음 (3파일 전부 포맷 준수)

### E3 — 커버리지

- 명령: `go tool cover -func=/tmp/t197_cover.out | grep -E "combineCodexStreams|parseCodexAuthLine|classifyCodexAuthFile|readCodexAuthFile|classifyCodexAuth"`

```
mcp_codex.go:1442: classifyCodexAuthFile  100.0%
mcp_codex.go:1462: readCodexAuthFile      100.0%
mcp_codex.go:1507: combineCodexStreams    100.0%
mcp_codex.go:1536: parseCodexAuthLine     100.0%
mcp_codex.go:1564: classifyCodexAuth      100.0%
```

- AC-CL-015 대상 순수 함수 3종(`classifyCodexAuthFile` · `parseCodexAuthLine` · `combineCodexStreams`) 각 100% — **게이트 통과**. 패키지 전체는 78.9%, `mcp_codex.go` 파일 함수 평균은 84.2%(60개 함수) — 아래 Gap 참조.

### E4 — 서브에이전트 경계 (C-HRA-008)

- 명령: `grep -rn "AskUserQuestion" internal/cli/ --include="*.go" | grep -v "_test.go" | grep -v "// " | wc -l` → 18
- 18건 전수 확인 결과 전부 **문서 문자열·agentlint 의 LR-01 규칙 구현 코드·원래 있던 주석**("never invokes AskUserQuestion" 계열)이며 실제 도구 호출 0건. 이번 diff 3파일이 기여한 히트: `mcp_codex.go:1189` 주석 1건(pre-existing, diff 밖 영역) — 이번 변경이 새로 넣은 것 0건.

### E5 — lint

- `golangci-lint run` → `0 issues.` (rc 0)

### E6 — 커밋 (push 안 함)

- 이번 세션 커밋 1건: 아래 E.3 신호의 `run_commit_sha` 참조. 커밋은 명시적 pathspec 4건(코드 3파일 + progress.md)으로만 staging. push 는 리드 지시(미션: Do NOT push)에 따라 **하지 않았다**.

### E7 — blocker

없음. 상속 diff 리뷰 중 실제 자격증명 형태의 값 유입 0건 확인(심어진 것은 anti-leak 프로브 `SENTINEL-TOKEN-9x9` 뿐 — 유지).

### 남은 Gap (정직한 목록)

- **패키지 전체 커버리지 78.9% / mcp_codex.go 함수 평균 84.2%** — 프로젝트 일반 하한(85%, package-level) 미달 칸이 이 파일에 남는다. 다만 그 미달 함수들은 이 SPEC 이전부터 있던 분들(SPEC 소관 아님)이고 이 SPEC 의 게이트(AC-CL-015: 순수 함수 3종 각 100%)는 통과. 100% 미만 함수를 끌어올리는 것은 범위 밖(drive-by 금지).
- **tty 왕복** — 기존 그대로 명시적 Gap (spec.md 「판정 제외」). 운영자 수동 측정 항목.
- **Windows 실행** — `GOOS=windows go vet` · (WIP 관측의) `go test -c` 컴파일 판정만 있고 실행은 release multi-OS 레그 몫 (AC-CL-014 관측 범위 서술대로 Gap).
- **부채 D4 · D5 · D6 · D7 · D8** — M1 범위 밖, 그대로.

### 실측해야 할 것 — 갱신

- ~~internal/cli 패키지 실행 시간~~ → **실측됨: 259.848s** (커버리지 프로파일 포함 실행, 이번 세션). 180칸 봉쇄 행렬(SPEC-CODEX-INIT-001, AC-CI-011)의 기준선으로 이 값을 쓸 수 있다.

## §E.3 Run-phase Audit-Ready Signal

```yaml
run_complete_at: 2026-08-26
run_commit_sha: pending-backfill-m1-complete
run_status: green-m1
ac_pass_count: 3
ac_fail_count: 0
preserve_list_post_run_count: 3
l44_pre_commit_fetch: skip-worktree-isolated-lane
l44_post_push_fetch: not-pushed
new_warnings_or_lints_introduced: 0
cross_platform_build:
  go_build: pass
  go_vet: pass
  goos_windows_vet: pass
  gofmt: pass
total_run_phase_files: 4
m1_to_mN_commit_strategy: per-milestone
```

M1 완결 시점의 신호. M2(CODEX_HOME 해석 + 리드아웃 조립) · M3(커맨드 표면) · M4(도움말 문안) 잔여 — `run_status: green-m1` 은 M1 한정 GREEN 이지 run-phase 전체 완료가 아니다. §E.3 최종본(전 마일스톤 완료)은 그때 갱신한다.
