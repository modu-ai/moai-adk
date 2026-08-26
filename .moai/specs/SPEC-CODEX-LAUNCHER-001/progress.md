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

- ~~internal/cli 패키지 실행 시간~~ → **실측됨: 259.848s** (M1 상태, 커버리지 포함) → **M2 상태 재실측: 181.640s** (커버리지 포함, §E.2 M2). 180칸 봉쇄 행렬(SPEC-CODEX-INIT-001, AC-CI-011)의 기준선은 최신값(181.6s)을 쓴다.

---

## M2 — CODEX_HOME 해석 + 리드아웃 조립: **완결** (2026-08-27)

세 번째 위임(t197-m2). 산출: `internal/cli/codex_readiness.go`(신규) + `internal/cli/codex_readiness_test.go`(신규, 시험 12개 함수 / 31 서브테스트). 기존 파일 무변경(`git status` 관측: untracked 2건, 수정 0건). 이 문서의 모든 근거는 이번 세션이 직접 실행해 관측한 것이다.

설계: `codexReadiness` 구조체가 공유 프로브(`codexSetupProbe` seam → `ProbeCodexSetup`) · CODEX_HOME 해석(`resolveCodexHomeDir` 재사용 + `os.Stat` 존재 판정) · 배선 판정(`classifyCodexWiring` — 파일 집합 기반, `codexadapter.ValidateConfig` 화이트리스트 소비) · 에이전트 TOML 수(`.codex/agents/moai/*.toml` glob)를 한 구조체로 조립하고 `rows()`가 plan §C.4 의 고정 6행(`codex`/`home`/`auth`/`wiring`/`agents`/`harness`)을 렌더한다. 배선 상태 토큰 폐집합 `{not wired, partial, wired, invalid}`, 조치 문구 `run moai init --agent codex` 는 불완전 5상태 전부에만 붙는다. 런처는 아무것도 쓰지 않는다(CODEX_HOME 디렉터리 생성 0건 관측).

### E8 — RED (스켈레톤 방식)

구현 전 타입·상수·seam만 실제이고 함수 본체가 영값을 반환하는 스켈레톤을 두고 시험을 먼저 돌렸다:

- 명령: `go test ./internal/cli/ -v -run 'TestResolveCodexHomeDir|TestClassifyCodexWiring|TestCodexReadiness|TestCountCodexAgentTOMLs' -count=1 -timeout 120s`
- 관측 (rc 1): **31 RUN / 신규 동작 6개 함수 FAIL / 2개 PASS** — PASS 2개(`TestResolveCodexHomeDir_EnvAxis` · `_TrailingSeparatorJoinsCleanly`)는 M1이 이미 전달한 `resolveCodexHomeDir` 의 특성화(AC-CL-005 핀)라 원래 초록이며, 이는 특성화 기록이지 RED 가 아니다. RED 대상은 M2 신규 동작 6함수 전부:

```
--- FAIL: TestClassifyCodexWiring_SixStateMatrix (0.01s)
--- FAIL: TestCodexReadiness_NoBannedWordsAllStates (0.00s)
--- FAIL: TestCodexReadiness_SentinelPropagation (0.00s)
--- FAIL: TestCodexReadiness_RowLabelSetAndBinaryAbsent (0.00s)
--- FAIL: TestCodexReadiness_DoesNotCreateCodexHome (0.00s)
--- FAIL: TestCountCodexAgentTOMLs (0.00s)
    codex_readiness_test.go: row 0 is empty — a failed probe must degrade to its token, not skip the row
```

- 정직 기록: 첫 RED 실행에서 `TestResolveCodexHomeDir_GoesThroughHomeSeamNotHOMEEnv` 도 FAIL 했으나 이는 **시험 쪽 결함**이었다(env-set 칸은 seam을 부르지 않는데 호출수 +1을 기대). 시험 기대치를 칸별 seam 히트수로 고친 뒤 구현에 들어갔다.

### E1 — M2 소관 AC 관측

| AC | 상태 | 근거 (전 패키지 실행 안에서 통과) |
|----|--------|---|
| AC-CL-005 (CODEX_HOME 해석·출처) | **PASS (완결)** | `TestResolveCodexHomeDir_EnvAxis`(4칸: 설정 `/tmp/xyz`→env / 미설정·빈·공백뿐→default) · `_TrailingSeparatorJoinsCleanly`(홈 seam `/tmp/h/` → `filepath.Join` 결과 `/tmp/h/.codex`) · `_GoesThroughHomeSeamNotHOMEEnv`(`HOME=""` 상태에서 4칸 동일 + seam 히트수 칸별 0/1/1 검증) |
| AC-CL-004 (리드아웃 실제 값) | PASS — M2 판정 가능 다리 전부; 맨몸/status 형태 동일성·MCP/web 교차는 M3+/블로커 | `TestClassifyCodexWiring_SixStateMatrix`(7 fixture: 디렉터리 부재·빈 디렉터리·한쪽만×2·양쪽 정상·위반·파싱실패 → 이름 붙은 상수와 행 전문 `==`; 두 partial 칸 배타성; 상태 폐집합; 조치 문구 5상태/정상 상태 부재) · `TestCodexReadiness_SentinelPropagation`(세 sentinel `SENTINEL-VER-9x9`·`/sentinel/path/codex`·`sentinel-provider` 각 행 정확 일치 + `ValidateConfig` sentinel 위반이 배선 행에 그대로 반영) |
| AC-CL-006 (정보성) | PASS — M2 다리 | `TestCodexReadiness_NoBannedWordsAllStates`(6상태 × 6행에서 금지어 6종 대소문자 무시 히트 0 + 조치 문구 존재/부재). rc 0·stdout 전용·stderr 0바이트 판정은 M3 커맨드 표면 소관 |
| AC-CL-011 (codex 부재 시 리드아웃) | PASS — M2 다리 | `TestCodexReadiness_RowLabelSetAndBinaryAbsent`(바이너리 부재 × 배선 wired/not wired: 행 라벨 집합 == {codex, home, auth, wiring, agents, harness}, 바이너리 행 `not found`, 배선 행 상수 일치, harness 행 고정) |
| AC-CL-012 (쓰기 없음) | PASS — CODEX_HOME 다리 | `TestCodexReadiness_DoesNotCreateCodexHome`(존재하지 않는 경로 가리킴 → 2회 조립 후에도 `os.Stat` IsNotExist + home 행 `missing`·`run codex login` 보고). 4동사 × 격리 홈 스냅샷은 M3 소관 |
| AC-CL-015 (공유 러너 무회귀) | PASS (재관측) | 아래 회귀축 + 이름 대조 + 커버리지. M2 순수함수 7종 100% |

**전 패키지 실행**: `go test ./internal/cli/ -coverprofile=/tmp/t197_m2_cover_final.out -count=1 -timeout 1200s` → `ok ... 181.640s coverage: 79.0% of statements` (rc 0). M1 상태(259.8s)보다 짧다.

**회귀축 (AC-CL-015)**: `go test ./internal/cli/ -json -run Codex -count=1 -timeout 1200s` → rc 0, 최상위 통과 시험 함수 130개(M1 118 + M2 12). **이름 대조**: `git grep -o -E "^func Test[A-Za-z0-9_]*Codex[A-Za-z0-9_]*" HEAD -- internal/cli` (123) vs 작업 트리 동일 grep (135) → **추가 12건, 삭제·개명 0건**. **skip 5건**: 전부 기존 `TestCodexLive_*` (SPEC-CODEX-PHASE2-001 소관 opt-in 라이브 프로브) — M1 기록과 동일 구성, 이 SPEC 시험의 skip 0, GOOS skip 3칸도 미발동(macOS 실행).

### E3 — 커버리지

`go tool cover -func=/tmp/t197_m2_cover_final.out`:

```
codex_readiness.go:114:  probeCodexReadiness   100.0%
codex_readiness.go:131:  resolveCodexHomeInfo  100.0%
codex_readiness.go:152:  classifyCodexWiring   100.0%
codex_readiness.go:203:  codexPathExists       100.0%
codex_readiness.go:210:  countCodexAgentTOMLs  100.0%
codex_readiness.go:220:  codexWiringRowValue   100.0%
codex_readiness.go:231:  rows                  100.0%
```

- M2 함수 7종 전부 100%. M1 순수함수 5종(`classifyCodexAuthFile`·`readCodexAuthFile`·`combineCodexStreams`·`parseCodexAuthLine`·`classifyCodexAuth`)도 100% 유지 — 회귀 없음. 패키지 전체 79.0%(M1 78.9% → +0.1pp).

### E2 — 크로스 플랫폼 (이번 세션 관측)

- `go build ./...` → rc 0
- `go vet ./internal/cli/` → rc 0
- `GOOS=windows go vet ./internal/cli/` → rc 0
- `GOOS=windows go test -c ./internal/cli/` → rc 0 (AC-CL-014 — vet 보다 강한 링크 판정)
- `gofmt -l` 신규 2파일 → 출력 없음
- 신규 2파일 `go:build` 태그 0건 · `"syscall"` import 0건 (grep 관측)
- 신규 시험 GOOS skip 0건 (기존 fixture 3칸 이름 고정 유지 — 이번 추가 시험은 전 플랫폼 실행형)

### E4 — 서브에이전트 경계

- `grep -rn "AskUserQuestion" internal/cli/codex_readiness.go internal/cli/codex_readiness_test.go` → 0건 (rc 1)

### E5 — lint

- `golangci-lint run` → `0 issues.` (rc 0)

### E7 — 구조적 블로커 1건 (범위 판정 필요)

**AC-CL-007 폐집합 다리가 현행 트리에서 성립하지 않는다.** AC 의 기계 판정: "`internal/` 하위 비시험 `*.go` 중 provider 리터럴(`"chatgpt"` 또는 `"apiKey"`)을 담은 파일의 집합이 `{internal/cli/mcp_codex.go}` 와 같다". 실측(`grep -rln '"chatgpt"\|"apiKey"' internal/ --include="*.go" | grep -v _test`):

```
internal/web/codex_state.go
internal/cli/mcp_codex.go
```

`internal/web/codex_state.go:32-37` 이 표시층 미러 상수 4종(`codexAuthChatGPT`/`codexAuthAPIKey`/`codexAuthProvider`/`codexAuthUnknown`)을 갖고 있다. 이것은 분류 경로가 아니라(`internal/web/schemaform.go:114` `codexAuthProviderLabel` 의 표시 라벨 스위치가 소비) AC 의 기계 판정은 어긋난다. 해소는 (a) web 쪽 라벨 매핑을 CLI 주입으로 바꿔 리터럴을 제거하거나 (b) acceptance 예외를 명시하는 것 — 어느 쪽도 plan §C.1 의 파일 배치(신규 2 + 수정 1) 밖이라 **M2 는 손대지 않았고 리드 판정이 필요하다.** 이 SPEC 의 M2 산출물인 `codex_readiness.go` 자체는 provider 리터럴 0건(공유 프로브의 값을 전파만 한다). AC-CL-007 의 (b) MCP 도구 · (c) web 카드 sentinel 교차 다리도 이 판정과 함께 정리해야 한다(런처 표면 (a) 다리는 이번에 통과).

**처분 (2026-08-27)**: 운영자 결정으로 옵션 (b) acceptance 예외 명시를 채택해 해소 — 폐집합이 `{internal/cli/mcp_codex.go, internal/web/codex_state.go}` 로 확정되고 `codexAuth*` 표시층 미러가 예외 원소로 문서화됐다(커밋 `698de4683`). (b)·(c) sentinel 교차 다리 판정은 별도 잔여 Gap 으로 M3 소관.

### 남은 Gap (정직한 목록)

- **AC-CL-004/006/011/012 의 커맨드 수준 다리** — 맨몸/status 두 형태 실행·rc 0·stdout 전용/stderr 0바이트·10칸·4동사 스냅샷. M3(커맨드 표면) 소관.
- **AC-CL-007 폐집합·(b)(c) 교차** — 위 E7 블로커.
- **부채 D4 · D5 · D6 · D8** — M2 범위 밖, 그대로. (D7 은 M1 에서 acceptance 판이 우선한다고 확정됨 — 이번 구현은 참조 문법 재사용 없음)
- **tty 왕복 · Windows 실행** — 기존 그대로 명시적 Gap (판정 제외 조항).
- **`CODEX_HOME` 환경변수 상수 위치** — 임무 지침이 `internal/config/envkeys.go` 상수를 예시로 들었으나 M1 이 `codexHomeEnvVar` 를 mcp_codex.go 에 국소 상수로 세웠고 M1 검증이 그 상태로 GREEN 이었다. 검증된 M1 코드를 뒤집는 리팩터는 범위 밖으로 판단해 유지했다(인라인 문자열 0건 요건은 충족). 리드가 envkeys.go 이전을 원하면 별도 마이크로 변경으로 처리 가능.

## M3 — 커맨드 표면 + 동사 라우팅 + `--spawn`: **완결** (2026-08-27)

네 번째 위임(t197-m3). 산출: `internal/cli/codex_launcher.go`(신규) · 시험 4파일(`codex_launcher_test.go` · `_readout_` · `_guards_` · `_cross_test.go`, 신규) · fixture 2종(`testdata/codex-app-message.sh` / `.bat` — 플랫폼별 한 다리씩, GOOS skip 0) · `internal/web/codex_card_sentinel_test.go`(신규). 수정 3파일: `spawn.go`(tmux 부재 진단 상수화 + `checkSpawnPrereqs` 추출 — 기존 3호출자가 같은 체크를 공유) · `mcp_codex.go`(`handleCodexSetup`이 `codexSetupProbe` seam 경유로 변경, 1줄 + 주석) · `help.go`(Launchers 섹션에 codex 행 1행). 이 문서의 모든 근거는 이번 세션이 직접 실행해 관측한 것이다.

설계: `codexCmd`는 런처 패밀리 관례(cc.go) 그대로 — launch 그룹 + `DisableFlagParsing` + `SilenceErrors/SilenceUsage`(진단 바이트를 상수 그대로 유지). 동사 라우팅은 패키지 수준 맵 `codexVerbRouting`(폐집합의 원본 — 미지 토큰은 맵 부재로 거부, default 분기 없음). 두 기동 자리는 acceptance 판정 어휘의 "포착 seam" 그대로: 직접 경로 `codexDirectLaunchFn`(*exec.Cmd 를 받아 7필드 포착) · spawn 경로 `codexSpawnLaunchFn`(새 창에서 실행될 대상 (dir, program, args) 계약 — program/argv 는 tmux 가 아니라 codex). tmux 실행 원시(`exec.Command("tmux")`)는 spawn.go 소유 그대로(AC-CL-016 폐집합 축2 — 이 SPEC 파일의 기동 원시는 `exec.Command(req.Program, ...)` 1곳). 종료코드 전파는 `codexPropagateLaunchError`: 자식 `*exec.ExitError` → **의도적** `exitCodeError` 변환(ResolveExitCode 가 raw 를 chain-wide 거부하므로 변환이 전파를 가능하게 하는 유일한 길), 그 외는 `execerr.StatusDetail` 기술.

### E8 — RED (스켈레톤 방식)

타입·상수·seam·등록만 실제이고 라우팅/기동/리드아웃 본체가 영값인 스켈레톤을 두고 시험을 먼저 돌렸다:

- 명령: `go test ./internal/cli/ -v -run 'TestCodex(Command|VerbRouting|Spawn|App|Direct|Readout|Launch|Launcher|Sentinel|SpecFiles)' -count=1 -timeout 300s`
- 관측 (rc 1): **68 RUN / 16 top-level FAIL / 4 PASS**. PASS 4개는 전부 스켈레톤 실체부의 특성화(라우팅 맵 폐집합 · 중립성 스캔 · 빌드태그/syscall 0 · tmux 진단 단일 원천)다.
- RED 중 발견한 **시험 결함 2건** (구현 결함 아님 — 시험을 고친 뒤 구현에 들어갔다): (1) guards 의 `ExecPrimitives` 스캔이 주석 안의 `exec.Command("tmux")` 언급을 오매치 — 주석 라인 필터 추가. (2) web 의 `Installed=false` 칸이 카드의 not-installed 렌더(올바른 동작)를 시임 우회로 오판 — 칸을 `Installed=true` 로 수정.

### E1 — M3 소관 AC 관측

| AC | 상태 | 근거 (전 패키지 실행 안에서 통과) |
|----|--------|---|
| AC-CL-001 (커맨드 등록) | **PASS** | `TestCodexCommand_RegisteredInLaunchGroup` — "Launchers" 섹션 헤딩 정확히 1회 + 블록에 4 런처 동일 블록 + 심볼 비교(`codexCmd.GroupID == ccCmd.GroupID`, `Parent() == rootCmd`) · `TestCodexCommand_HelpExitsZero`(--help/-h rc 0). **해석 기록**: AC 문면의 "LAUNCH COMMANDS 그룹 헤딩"은 이 레포의 실사용자 표면에서 help.go 의 커스텀 `renderRootHelp` "Launchers" 섹션에 대응한다 — cobra 기본 usage 템플릿은 이 레포 루트 help 에 노출되지 않는다(SetHelpFunc 우회). 판정은 사용자 표면 기준으로 옮기고 cobra 그룹 심볼 판정은 그대로 유지했다 |
| AC-CL-002 (동사 라우팅) | **PASS** | `TestCodexVerbRouting_ClosedSets`(맵에서 유도한 launch 집합 == {cli, app} · readout 집합 == {"", status}) · `_LaunchCountsPerVerb`(맨몸/status 기동 0 · cli/app 각 1 — 두 자리 합) · `_UnknownTokenRejected`(bogus·cl·CLI·Cli·--model·-x 6칸: 기동 0 · rc 비영 · stderr == 사용법 상수 `codexUsageDiag`) · `_CwdCrossChecked`(3루트 시나리오 → 포착 Dir == 해석 루트 + seam 1회) · `_AppArgvExact`(argv == [codex, app]) · `_ExitCodePropagation`(0·1·2·126·127 5칸 전파) · `_StdioIdentity`(포착 stdio == 시험 프로세스 os.Stdin/Stdout/Stderr 항등) · `_PassthroughTailExact`(`--` 뒤 `--model o3 "a b" '$x' --flag=v` 토큰 단위 일치) · `_HelpAfterDashDashIsNotHelp`(-- 뒤 --help 는 codex 소유) |
| AC-CL-003 (--spawn 패리티) | **PASS** | `TestCodexSpawn_Parity`(cli/app --spawn → spawn 1회 · 직접 0회 · 같은 꼬리의 spawn 포착 (program, argv) 이 직접 포착과 토큰 동일 — spawn 포착의 stdio 는 nil 명시) · `_TmuxAbsentDiagnosticBytes`(cc 와 codex 의 tmux-부재 error 본문 바이트 동일 + 기동 0) · `_RejectedOnReadoutForms`(맨몸/status --spawn 거부) · `_TmuxDiagnosticSingleSource`(공유 진단 리터럴이 비시험 Go 전체에서 정확히 1회 — `errTmuxSessionRequired` 상수) · `_RealAssemblyThroughStubTmux`(실제 spawn 조립: tmux 에 넘어가는 명령 문자열이 셸 인용된 토큰열 + tmux 실패 래핑). **해석 기록**: AC 문면 "spawnLaunch 를 1회 호출"은 tmux-새-창 경로 seam 을 가리킨다 — (program, argv) 가 tmux 가 아니라 새 창 대상 codex 여야 한다는 단서("기록 대상이 tmux 면 이 등식은 성립할 수 없다")가 기존 `spawnLaunch(out, sub, args)` 의 moai-재발행 조립과 양립하지 않으므로, codex 는 자체 spawn seam(새 창 대상 계약)을 갖되 전제 체크(`checkSpawnPrereqs`)·진단 상수·tmux 원시 호출을 기존 spawnLaunch 와 공유한다. 진단 바이트 동등은 error 본문 직접 비교로 판정했다(cc 의 stderr 는 cobra 가 "Error: " 접두를 붙이므로 전체 stderr 가 아닌 본문 기준) |
| AC-CL-004 (리드아웃 실제 값) | **PASS — 커맨드 다리 완결** | `TestCodexReadout_CommandTenCells`(M2 다리와 동일 상수로 배선 행 `==` + 두 형태 값 동일) · `_SentinelRowsAtCommandSurface`(세 sentinel 행 — `SENTINEL-VER-9x9` · `/sentinel/path/codex` · `sentinel-provider` — 이 커맨드 표면에서 M2 상수와 정확 일치) |
| AC-CL-006 (정보성) | **PASS — 커맨드 다리 완결** | 같은 시험 10칸: rc 0 · 배선 행 상수 일치 · `moai init --agent codex` 포함 · stdout 전용 · stderr 0바이트 · 금지어 6종 히트 0 · `TestCodexReadout_WiredStateOmitsAction`(wired 에서 조치 문구 부재) |
| AC-CL-007 (분류 단일성) | **PASS — (b)(c) 교차 완결** | `TestCodexSentinel_CrossSurfacesCommandAndMCP` — 스텁 하나로 (a) 런처 리드아웃 · (b) `codex_setup` 응답(`auth_provider`·`binary`·`version` 전부 sentinel 정확 일치) 각 1회. (b) 를 위해 `handleCodexSetup` 이 `codexSetupProbe` seam 경유로 변경됐다(기본값이 `ProbeCodexSetup` 이므로 프로덕션 동작 불변). (c) web 카드: `internal/web/codex_card_sentinel_test.go` — 주입 시임에 sentinel 을 주면 렌더에 세 sentinel 전부 원문 등장(`codexAuthProviderLabel` 의 default 분기가 미지 토큰을 분류 없이 통과 — 두 번째 분류기 부재의 구조적 증거). **특성화**: (c) 는 이미 배선돼 있던 소비(SPEC-MCP-CONSOLE-001)를 sentinel 교차로 최초 판정한 것 — RED 가 아니라 이 시점에 판정 가능해진 칸이다. **폐집합 방정식 유지 실측**: `grep -rln '"chatgpt"\|"apiKey"' internal/ --include="*.go" \| grep -v _test` → 정확히 `{internal/cli/mcp_codex.go, internal/web/codex_state.go}` (개정판 그대로) — 이번 신규 파일의 리터럴 기여 0건 |
| AC-CL-010 (판정 불가 조치) | **PASS** | `TestCodexReadout_UnknownAuthFourAxes` — 네 축(auth.json 부재+양스트림 비어 rc 0 · 부재+러너 오류 · 부재+비영 rc+문법 불일치 · auth.json 파싱 실패)을 **실제 사다리 경유**(공유 프로브 seam 스텁 아님 — LookPath/러너 스텁만)로 돌려 네 칸 모두 auth 행 전문이 이름 붙은 상수 `auth ... unknown — run codex login status` 와 정확 일치 + 폐집합 {logged out, logged-out, not logged in, no credentials, signed out} 히트 0. 이 AC 는 커맨드 표면이 있어야만 판정 가능해 M3 소관이었다 — 리드 지침의 열거(004/006/011/012)에 없어 재점검에서 발견, 구현(`rows()` 의 unknown 조치 문구 — 상수 `codexAuthUnknownAction`)과 시험을 이번에 추가했다. sentinel 칸(조치 없음)은 M2 상수 그대로 유지 |
| AC-CL-011 (codex 부재) | **PASS — 커맨드 다리 완결** | `TestCodexReadout_BinaryAbsentFourCells`(바이너리 부재 × 배선 2 × 형태 2 = 4칸: rc 0 · 행 라벨 집합 == 6종 폐집합 · 바이너리 행 `not found` · 배선 행 상수) · `TestCodexLaunch_BinaryAbsentSingleDiagnostic`(cli/app: rc 비영 · 기동 0 · stderr 정확히 1줄 == 설치 상수 `codexInstallHint`) |
| AC-CL-012 (쓰기 없음) | **PASS — 커맨드 다리 완결** | `TestCodexLauncher_NoWriteSnapshot` — 격리 홈(HOME·XDG 3종·TMPDIR·CLAUDE_CONFIG_DIR 전부 격리 아래; 프로젝트 루트·CODEX_HOME·프로필 디렉터리도 그 안)에서 8런(4형태 × CODEX_HOME 존재/부재): 실행 전후 격리 홈 트리 전체(파일 목록·mode·mtime 나노초) 동일 · 부재 CODEX_HOME 은 실행 후에도 `IsNotExist` + home 행이 missing·조치 보고 |
| AC-CL-013 (중립성) | **PASS** | `TestCodexCommand_NeutralityScan` — `codexCmd` 의 string/[]string 필드(리플렉션 전수) + 모든 플래그 usage 에서 금지 패턴 9종(SPEC-/REQ-/카드id/ISO날짜/7자리SHA//Users///home//CLAUDE.local/.moai/reports) 0건 + 비-ASCII 0건. RED 중 상수의 em-dash 2건을 이 스캔이 잡아 하이픈으로 수정했다 — 스캔이 실제로 판정력이 있었다는 증거 |
| AC-CL-014 (게이트·크로스 플랫폼) | **PASS** | 정적: `TestCodexSpecFiles_NoBuildTagsOrSyscall`(이 SPEC 파일 2종에서 OS 빌드태그 0 · `"syscall"` import 0 · 교체 식별자 0 · GOOS 접미 파일 0). 실행: `go build ./...` rc 0 · `go vet ./...` rc 0 · `GOOS=windows go vet ./internal/cli/ ./internal/web/` rc 0 · `GOOS=windows go test -c ./internal/cli/` rc 0(링크 판정). GOOS skip 는 기존 fixture 3칸뿐 — 이번 M3 시험(플랫폼별 fixture 다리 포함)은 전 플랫폼 실행형, macOS 실행에서 GOOS skip 0 관측 |
| AC-CL-015 (공유 러너 무회귀) | **PASS (재관측)** | 회귀축 `go test ./internal/cli/ -json -run Codex -count=1 -timeout 1200s` → rc 0 · 이 SPEC 시험 skip 0 · 기존 skip 5건은 전부 `TestCodexLive_*`(SPEC-CODEX-PHASE2-001 소관 opt-in 라이브 프로브, M1/M2 기록과 동일 구성). **이름 대조**: `git grep` HEAD(135) vs 작업 트리(169) → **삭제·개명 0건 / 추가 34건**. 순수함수 3종 + M1·M2 함수 커버리지 100% 유지(아래 E3) |
| AC-CL-016 (데스크톱 위임) | **PASS** | `TestCodexVerbRouting_AppArgvExact`(argv [codex, app]) · `TestCodexApp_FailureHasNoFollowup`(seam 실패 후 추가 기동 0 · rc 동일) · `_LaunchedProgramsClosedSet`(기동-bearing 칸 전체의 포착 basename 합집합 == {codex} — 축1) · `TestCodexSpecFiles_ExecPrimitivesCodexOnly`(이 SPEC 파일의 기동 원시 호출 0번 인자가 전부 codex 경로 변수 — 축2, `codex_readiness.go` 는 기동 0건) · `_OutputPassthroughDirect`/`_OutputPassthroughSpawn`(fixture 출력 — `install the desktop app from ...` — 이 직접/spawn 양경로에서 바이트 동일) · `_RealChildExitCodePropagates`(실자식 exit 7 → moai rc 7 — env 훅 `CODEX_FIXTURE_EXIT` 로 전 플랫폼 동일 시맨틱) |

### E3 — 커버리지

`go tool cover -func=/tmp/t197_m3_cover4.out` — `codex_launcher.go` 12함수: 11개 100% + `runCodex` 95.7%(미커버는 `-h` 단축 루프의 잔여 라인). M1 순수함수 5종(`classifyCodexAuthFile`·`readCodexAuthFile`·`combineCodexStreams`·`parseCodexAuthLine`·`classifyCodexAuth`)과 M2 함수 7종 100% 유지 — 회귀 없음. 패키지 전체 79.1%(M2 79.0% → +0.1pp). `codex_readiness.go` 도 전 함수 100% 유지(auth unknown 조치 추가 후 포함).

### E2 — 크로스 플랫폼 (이번 세션 관측)

- `go build ./...` → rc 0
- `go vet ./...` → rc 0 · `GOOS=windows go vet ./internal/cli/ ./internal/web/` → rc 0
- `GOOS=windows go test -c ./internal/cli/` → rc 0
- `gofmt -l` 신규/수정 전 파일 → 출력 없음
- `golangci-lint run` → `0 issues.` (초회 8건 — errcheck 7 · unused 1 — 을 수정 후 재측정)

### E4 — 서브에이전트 경계

- 신규 5파일 grep `AskUserQuestion` → 0건. CLI 코드는 프롬프트 없이 동사·플래그·상수 진단으로만 응답한다.

### E5 — lint

- `golangci-lint run` → `0 issues.` (rc 0)

### E6 — 커밋 (push 안 함)

- 이번 세션 커밋 1건 — 아래 §E.3 신호의 `run_commit_sha` 참조. 명시적 pathspec 만 staging. push 는 리드 지시(Do NOT push)에 따라 **하지 않았다**.

### E7 — blocker

없음. verify 스냅샷은 워킹트리 키 `108b77fc7d60428e9e7ce3434764bab9c12f9cd8:d9ec49e84f1d4d56` 로 기록(전체 수트 · 정적 배치 2체크, 전부 rc 0).

### 실행 시간 재실측

- internal/cli 전체(커버리지 포함): **301.465s**(최종판) — M2 181.6s 대비 증가(M3 시험 91케이스 추가). 180칸 봉쇄 행렬(SPEC-CODEX-INIT-001 AC-CI-011) 기준선은 이 최신값으로 갱신한다. 중간 실측 217.6s/221.2s/235.2s — 같은 트리에서 부하 편차가 있으니 상한 여유는 1200s 유지 권장.

### 남은 Gap (정직한 목록)

- **tty 왕복 · Windows 실행** — 판정 제외 조항 그대로 (운영자 수동 측정 · release multi-OS 레그 몫).
- **부채 D4 · D5 · D6 · D8** — M3 범위 밖, 그대로. (D7 은 M1 에서 acceptance 판 우선 확정 — 이번 구현도 참조 문법 재사용 없음)
- **`runCodex` 95.7%** — 미커버는 `-h` 단축 루프 잔여 라인. 순수함수 3종 100% 게이트(AC-CL-015)와 무관.
- **`requireTmuxSpawnEnv` 환경 skip** — tmux/moai 바이너리가 없는 환경에서 spawn 성공-경로 칸 3종(`Spawn_Parity` 본체 · `_RealAssemblyThroughStubTmux` · `LaunchedProgramsClosedSet` · `_OutputPassthroughSpawn`)이 기존 `spawn_test.go` 의 `requireTmuxBinary` 관행(바이너리 부재 시 skip — LookPath 실호출이 판정 대상이므로 스텁하면 그 체크를 덮지 못함)을 따라 skip 될 수 있다. GOOS skip 가 아니라 AC-CL-014의 "GOOS skip 3칸" 계약과 무관하며, 이 머신 관측에서는 skip 0이었다.
- **M4(도움말 문안 + 필요 시 템플릿 문서) 잔여** — `codexCmd` 의 Short/Long 은 동작에 충분한 최소문안으로 들어갔다(중립성 스캔 통과). 사용자 안내문 정비는 M4 소관.


## M4 — 도움말·예시 문안: **완결** (2026-08-27)

다섯 번째 위임(t197-m4, run-phase 마지막 마일스톤). 산출: `codex_launcher.go` Long 재작성 + `Example` 필드 신설 + `codex_launcher_test.go`에 `TestCodexCommand_HelpCopyGuidance`(서브테스트 3) + `go.mod` 정리(pflag indirect→direct). 신규 파일 0건. 이 문서의 모든 근거는 이번 세션이 직접 실행해 관측한 것이다.

### E8 — RED

- 명령: `go test ./internal/cli/ -v -run 'TestCodexCommand_HelpCopyGuidance' -count=1 -timeout 120s`
- 관측 (rc 1) — 서브테스트 2 FAIL / 1 PASS:

```
--- FAIL: TestCodexCommand_HelpCopyGuidance/wiring_action_matches_the_generator
    codex_launcher_test.go:916: Long does not name the wiring action "moai init --agent codex" (derived from codexWiringAction)
--- FAIL: TestCodexCommand_HelpCopyGuidance/examples_are_copy-pasteable
    codex_launcher_test.go:936: Example is empty — M4 requires example copy
--- PASS: TestCodexCommand_HelpCopyGuidance/every_routing_verb_documented
```

- PASS 칸은 기존 Long이 이미 동사 집합을 문서화한 것에 대한 특성화다. 판정 축 3종(전부 파생 비교 — 상수 재기술 없음): (1) help가 안내하는 배선 조치가 리드아웃 상수 `codexWiringAction`에서 파생한 문구와 정합 — t88 생성기의 실제 플래그 표면(`moai init --agent codex`, `init.go` 폐집합 claude/codex/both)과 일치, (2) 폐집합 `codexVerbRouting`의 미빈 동사 토큰 전부가 Long에 문서화(+`--spawn`·`--` 언급), (3) Example 이 비지 않고 주석·빈 행 제외 전 행이 `moai codex` 호출.

### E1 — M4 소관 AC 관측

| AC | 상태 | 근거 |
|----|--------|---|
| AC-CL-013 (중립성) | **PASS — 신규 문안 위에서 재판정** | `TestCodexCommand_NeutralityScan`(reflect 전수 — 신설 `Example` 필드 포함) 통과 + 템플릿 가드 `MOAI_TEMPLATE_LEAK_STRICT=1 go test ./internal/template/...` → `ok 22.836s` 외 3패키지 ok (AC 첫 조항 — 도움말 문안이 확정된 M4 시점에 처음 실행·기록) |
| AC-CL-001 (--help rc 0) | PASS (재판정) | `TestCodexCommand_HelpExitsZero` + 빌드 바이너리 `moai codex --help` rc 0 관측 (커스텀 help 렌더러의 USAGE/EXAMPLES/FLAGS 섹션 전부 렌더) |
| AC-CL-007 (폐집합) | PASS (유지 확인) | `grep -rln '"chatgpt"\|"apiKey"' internal/ --include="*.go" \| grep -v _test` → 정확히 `{internal/cli/mcp_codex.go, internal/web/codex_state.go}` — 개정판 그대로, M4 diff의 리터럴 기여 0건 |
| AC-CL-015 (무회귀) | PASS (재관측) | 아래 회귀축 |

**M4에서 새로 열리는 다리 점검**: AC-CL-013의 템플릿 가드 조항이 유일하게 이 시점에 판정 가능해졌다(도움말 문안 확정 전에는 그 전제가 성립하지 않는다). 나머지 help-표면 판정은 기존 시험이 새 문안 위에서 자동 재판정했다. 그 외 AC는 M4가 새로 판정 가능하게 만들지 않는다 — 16/16 그대로.

### M2 지연 리뷰 — home 행 풀 경로: 결함 아님 (확인 기록)

plan §C.4 예시 블록은 `home ~/.codex (default)`로 적혀 있으나 이는 문서 축약형이다. 실측(빌드 바이너리, 아래 실 바이너리 관측 참조): `home     /Users/goos/.codex (default)` — 풀 경로 렌더. AC-CL-005는 home 행 값 칸을 **해석된 경로 그 자체**와 정확 일치로 판정한다(`/tmp/xyz` → env 칸, `<home>/.codex` → `filepath.Join` 결과와의 비교). `~` 축약은 env 칸에서 성립하지 않고 default 칸의 Join 비교를 깨뜨린다 — REQ-CL-005 가 요구하는 것도 "resolved `CODEX_HOME`" 의 보고다. 같은 예시 블록의 codex 행(`/Users/…/bin/codex`)도 축약 표기임을 뒷받침한다. M4 SPEC 텍스트(도움말·예시 문안)는 리드아웃 행 형태를 요구하지 않는다. **변경하지 않는다.**

### 템플릿 문서 — 불필요 판정 (plan M4 "필요 시")

불필요. 템플릿 트리에서 런처 패밀리를 언급하는 파일(`.worktreeinclude`, `.moai/config/sections/crosssession.yaml` 등)은 전부 **워크트리 진입(`-w`)과 교차세션 설정 번역** 문맥이고, codex 런처는 `-w`가 없고 설정 파일을 쓰지 않는다(REQ-CL-013) — 참여하지 않는 표면에 문서를 추가할 근거가 없다. `moai codex` 를 언급하는 유일한 템플릿 파일(`.claude/rules/moai/core/moai-mcp-tools-catalogue.md`)은 MCP 세션 메시징 소관으로 런처와 무관하다.

### 판정 제외 조항 갈음 관측 — 실 바이너리 auth 왕복

spec 「판정 제외」: "실 바이너리 확인은 `moai codex status` 출력 1회를 progress.md 에 붙이는 것으로 갈음한다". 관측 (2026-08-27, 빌드 바이너리, 이 머신):

```
$ moai codex status; echo $?
codex    codex-cli 0.149.0 (/Users/goos/.local/bin/codex)
home     /Users/goos/.codex (default)
auth     chatgpt
wiring   not wired (.codex/hooks.json and .codex/config.toml absent) — run moai init --agent codex
agents   0 TOML
harness  moai hook --harness codex
0
```

`auth chatgpt` — §A.2 결함(현행 `unknown`)이 실 바이너리 경로에서 해소됐다. `wiring not wired` + 조치 문구는 이 저장소에 `.codex/`가 없는 기본 상태(spec §A.4)와 정확히 일치한다.

### 인계 수정 2건 (M3 잔여, 배차 지시)

1. **go.mod pflag drift — 해소**: M3 의 `codex_launcher_guards_test.go` 가 `github.com/spf13/pflag` 를 직접 import하는데 go.mod 가 `// indirect` 로 표기하고 있었다. `go mod tidy` → pflag 가 direct 블록(go.mod:25)으로 이동하고 `// indirect` 주석이 사라졌다. go.sum 무변경.
2. **`codex_launcher.go:88 unused method "argv"` 진단 — 미재현**: 현재 트리에서 `grep ') argv(' internal/cli/*.go` → 0건(그런 메서드가 존재하지 않음 — 라인 88은 주석), `golangci-lint run` → `0 issues.`(unused 러너 포함). 편집기 세션의 스테일 진단으로 판단 — 제거 대상 dead code 없음.

### E2 — 크로스 플랫폼 (이번 세션 관측)

- `go build ./...` → rc 0
- `go vet ./internal/cli/` → rc 0 · `GOOS=windows go vet ./internal/cli/` → rc 0
- `GOOS=windows go test -c ./internal/cli/` → rc 0 (AC-CL-014 — vet 보다 강한 링크 판정)
- `gofmt -l` 수정 2파일 → 출력 없음
- `MOAI_TEMPLATE_LEAK_STRICT=1 go test ./internal/template/...` → 3패키지 ok

### E3 — 커버리지

- 명령: `go test ./internal/cli/ -coverprofile=/tmp/t197_m4_cover.out -count=1 -timeout 1200s` → `ok ... 195.317s coverage: 79.1% of statements` (rc 0)
- `go tool cover -func` — 순수함수 3종(`classifyCodexAuthFile` 100% · `parseCodexAuthLine` 100% · `combineCodexStreams` 100%) + M1/M2 전 함수 100% 유지 — 회귀 없음. `runCodex` 95.7%(M3 과 동일, `-h` 단축 루프 잔여).

### E4 — 서브에이전트 경계

- 수정 2파일 grep `AskUserQuestion` → 0건 (rc 1)

### E5 — lint

- `golangci-lint run` → `0 issues.` (rc 0)

### 회귀축 (AC-CL-015)

- 명령: `go test ./internal/cli/ -json -run Codex -count=1 -timeout 1200s` → rc 0, top-level 통과 시험 함수 165개.
- **이름 대조**: HEAD(169) vs 작업 트리(170) → **추가 1건(`TestCodexCommand_HelpCopyGuidance`), 삭제·개명 0건**.
- skip 5건 — 전부 기존 `TestCodexLive_*`(SPEC-CODEX-PHASE2-001 소관 opt-in 라이브 프로브, M1/M2/M3 기록과 동일 구성). 이 SPEC 시험 skip 0, GOOS skip 3칸 미발동(macOS 실행).

### 실행 시간 재실측

- internal/cli 전체(커버리지 포함): **195.317s** (M3 301.4s 대비 단축 — 같은 트리 계열 부하 편차 범위). 180칸 봉쇄 행렬(SPEC-CODEX-INIT-001 AC-CI-011) 기준선은 이 최신값으로 갱신한다.

### 남은 Gap (정직한 목록)

- **tty 왕복 · 실 앱 기동 · Windows 실행** — 판정 제외 조항 그대로 (운영자 수동 측정 · release multi-OS 레그 몫).
- **부채 D4 · D5 · D6 · D8** — run-phase 내내 미해소 (acceptance 문면 개선 소관 — sync 로 넘기지 않고 이곳에 기록한다, 리드 지침).
- **`runCodex` 95.7%** — 기존 그대로 (`-h` 단축 루프 잔여).

---

## §E.3 Run-phase Audit-Ready Signal

```yaml
run_complete_at: 2026-08-27
run_commit_sha: pending-backfill-m4
run_status: green
ac_pass_count: 16
ac_fail_count: 0
preserve_list_post_run_count: 3
l44_pre_commit_fetch: skip-worktree-isolated-lane
l44_post_push_fetch: not-pushed
new_warnings_or_lints_introduced: 0
cross_platform_build:
  go_build: pass
  go_vet: pass
  goos_windows_vet: pass
  goos_windows_test_compile: pass
  gofmt: pass
total_run_phase_files: 19
m1_to_mN_commit_strategy: per-milestone
```

**Run-phase 최종 신호 (M1~M4 전부 완결)**. `run_status: green` — 마일스톤 한정 표기(`green-m3`)에서 run-phase 전체 완결형으로 진행. `ac_pass_count: 16` = AC-CL-001~016 전부 판정 가능한 다리에서 PASS — M1(008·009·015) + M2(005) + M3(001·002·003·004·006·007·010·011·012·013·014·016) + M4(013 신규 문안 재판정 + 템플릿 가드 첫 실행, 001/007/015 재판정). 판정 제외 조항 중 실 바이너리 auth 왕복은 M4 가 progress 부착 관측으로 갈음 닫았다(§E.2 M4); tty 왕복 · 실 앱 기동 · PR 단계 Windows 실행은 운영자 수동·CI 레그 몫으로 Gap 유지. 부채 D4·D5·D6·D8 은 acceptance 문면 개선 소관으로 이 문서에 기록된 채 남는다(§E.2 상단 부채 표 + 각 M Gap 목록). `run_commit_sha` 는 M4 코드 커밋 SHA 의 후속 docs 커밋 백필로 채운다(커밋은 자기 SHA 를 알 수 없다 — 이 브랜치에서 이미 쓴 two-commit 패턴).
