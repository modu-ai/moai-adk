# M5 — CLI 배선 (`moai feedback scrub` + 큐 동사)

SPEC-FEEDBACK-AUTO-SUBMIT-001 · cycle_type=tdd · 워크트리 `.claude/worktrees/t170`, 브랜치 `WT-auto-feedback`, base `3210da7d3`, M5 착수 시 HEAD `55dc0ec0a`.

## §1 Claim (주장)

1. `moai feedback scrub` 은 제목(`--title`)과 본문(stdin)을 **둘 다** 스크러버에 통과시키고, 결과를 `verdict`/`title`/`body`/`findings`/`reason` 5필드 단일 JSON 객체로 stdout 에 낸다. 각 finding 은 `where` 를 담는다. 종료 코드 0. — **AC-F-003**
2. 정책을 로드할 수 없으면 종료 코드가 0이 아니고 stdout 에 JSON 이 나오지 않는다(fail-closed). — **AC-F-004**
3. 판정 축과 종료 코드 축은 분리돼 있다: 정책 차단은 exit 0 + `"verdict":"blocked"`.
4. M4 가 넘긴 위험 1(빈 `Options.ProjectRoot` = 산출물 없음)은 `--root` + `ResolveProjectRoot` 폴백 배선으로 닫혔고, 마스킹 로그와 큐 파일이 **실제로 생성되는 것**을 빌드된 바이너리로 관측했다.
5. M4 가 넘긴 위험 2(D4 초안/큐 혼동)는 CLI 표면에서도 유지된다: 큐 동사는 `queue.json` 외에 아무것도 읽지 않는다.

## §2 Evidence (증거)

### AC-F-003

```
$ go test ./internal/cli/ -run 'TestFeedbackScrubContract' -v
=== RUN   TestFeedbackScrubContract
--- PASS: TestFeedbackScrubContract (0.00s)
PASS
ok  	github.com/modu-ai/moai-adk/internal/cli	0.647s
```

### AC-F-004

```
$ go test ./internal/cli/ -run 'TestFeedbackScrubToolFailureExitsNonZero' -v
=== RUN   TestFeedbackScrubToolFailureExitsNonZero
--- PASS: TestFeedbackScrubToolFailureExitsNonZero (0.00s)
PASS
ok  	github.com/modu-ai/moai-adk/internal/cli	0.643s
```

### RED (구현 전)

```
$ go test ./internal/cli/ -run 'TestFeedbackScrubContract|TestFeedbackScrubToolFailureExitsNonZero' -v
# github.com/modu-ai/moai-adk/internal/cli [github.com/modu-ai/moai-adk/internal/cli.test]
internal/cli/feedback_test.go:49:9: undefined: newFeedbackCmd
internal/cli/feedback_test.go:129:52: undefined: feedbackSecurityFileName
internal/cli/feedback_test.go:197:14: undefined: resolveFeedbackRoot
FAIL	github.com/modu-ai/moai-adk/internal/cli [build failed]
FAIL
```

### 바이너리 스모크 (`go build -o /tmp/claude-501/moai-m5 ./cmd/moai`, exit 0)

scratch 프로젝트 `/tmp/claude-501/m5-smoke`(`.moai/config/sections/security.yaml` 에 `env_scrub_extra: [SMOKE_SECRET_NAME]`).

(a) stdout JSON — `title` 필드 포함, 제목에서도 실제로 마스킹됨:

```
$ SMOKE_SECRET_NAME=smoke-value-0123456789 moai-m5 feedback scrub \
    --root /tmp/claude-501/m5-smoke --title 'crash while handling smoke-value-0123456789' < m5-body.txt
{
  "verdict": "ok",
  "title": "crash while handling s...6789",
  "body": "body mentions s...6789 twice: s...6789\n",
  "findings": [
    { "kind": "env", "where": "title", "count": 1 },
    { "kind": "env", "where": "body",  "count": 2 }
  ],
  "reason": ""
}
scrub_exit=0
```

(b) 마스킹 로그가 실제로 생성됨:

```
$ ls -l /tmp/claude-501/m5-smoke/.moai/logs/feedback-mask.log
-rw-------@ 1 goos  wheel  97 Aug 23 18:55 …/feedback-mask.log
$ cat …/feedback-mask.log
2026-08-23T18:55:11+09:00 | total=3 | kind=env where=title count=1 | kind=env where=body count=2
```

(c) 큐 파일이 실제로 생성됨 + list/resolve:

```
$ moai-m5 feedback queue enqueue --root /tmp/claude-501/m5-smoke < m5-result.json
{ "id": "f1", "title": "crash while handling s...6789", "body": "…", "queued_at": "2026-08-23T09:55:16Z", "attempts": 0 }
enqueue_exit=0
$ ls -l /tmp/claude-501/m5-smoke/.moai/state/feedback/queue.json
-rw-------@ 1 goos  wheel  253 Aug 23 18:55 …/queue.json
$ moai-m5 feedback queue list --root /tmp/claude-501/m5-smoke      → items 1건, list_exit=0
$ moai-m5 feedback queue resolve f1 --root /tmp/claude-501/m5-smoke → { "id": "f1", "removed": true }, resolve_exit=0
```

(d) acceptance.md 의 jq 스모크 라인 그대로:

```
$ echo 'hello' | moai-m5 feedback scrub --root /tmp/claude-501/m5-smoke --title 'a title' \
    | jq -e '.verdict and (.title|type=="string") and (.findings|type=="array")'
true
jq_smoke_exit=0
```

(e) 바이너리에서의 fail-closed (security.yaml 파손):

```
$ moai-m5 feedback scrub --root /tmp/claude-501/m5-smoke --title 't' < m5-body.txt > out 2> err
exit=1
$ wc -c < out ; wc -c < err
0
272
```

### 회귀 + 정적 검사

```
$ go test -timeout 900s ./internal/cli/... ./internal/feedback/...
ok  	github.com/modu-ai/moai-adk/internal/cli	343.442s
ok  	github.com/modu-ai/moai-adk/internal/feedback	5.636s
…(하위 패키지 전부 ok)
--- FAIL: TestConstitutionCrossReference (0.00s)
    agent_lint_test.go:1249: moai-constitution.md should cross-reference agent-authoring.md for effort matrix
FAIL	github.com/modu-ai/moai-adk/internal/cli/agentlint	1.194s
regression_exit=1

$ go test -race ./internal/feedback/          → ok … 1.967s (race_exit=0)
$ go vet ./internal/cli/... ./internal/feedback/...              → vet_exit=0
$ GOOS=windows go vet ./internal/cli/... ./internal/feedback/... → winvet_exit=0
$ golangci-lint run --timeout=3m ./internal/cli/...              → 0 issues.
$ gofmt -l internal/cli/feedback.go internal/cli/feedback_test.go internal/cli/root.go → (무출력)
```

**FAIL 1건의 귀속**: `TestConstitutionCrossReference` 는 `.claude/rules/moai/core/moai-constitution.md` 를 읽어 `agent-authoring.md` 인용을 요구한다. 그 파일은 `grep -c agent-authoring` → `0` 이고, M5 diff 는 Go 파일 3개뿐이다(`git status --porcelain` → ` M internal/cli/root.go` + 신규 `internal/cli/feedback.go`, `internal/cli/feedback_test.go`). 규칙 파일 부채이며 M5 가 만든 것이 아니다.

## §3 Baseline-attribution (baseline 귀속)

- 트리: 워크트리 `.claude/worktrees/t170`, 브랜치 `WT-auto-feedback`, M5 착수 HEAD `55dc0ec0a`, base `3210da7d3`.
- 부재 baseline (M5 가 신규임을 고정):
  - `git ls-tree -r 3210da7d3 --name-only -- internal/cli/feedback.go internal/cli/feedback_test.go | wc -l` → `0`
  - `git show 3210da7d3:internal/cli/root.go | grep -c newFeedbackCmd` → `0`
  - `git show 55dc0ec0a:internal/cli/root.go | grep -c newFeedbackCmd` → `0`
- RED baseline: 위 §2 의 build-failed 출력(구현 심볼 3개 부재). 실행 시점은 `feedback_test.go` 만 존재하고 `feedback.go` 는 없던 트리.
- 회귀 baseline: 위 명령들은 전부 이번 회차, 이 트리에서 실행했다. 인용한 수치는 다른 마일스톤·다른 시점의 측정을 옮겨온 것이 아니다.
- 산출물 형태 baseline: 마스킹 값 형태 `s...6789` 는 `internal/github` 의 기존 마스커 출력이며 M5 가 새로 만든 형태가 아니다.

## §4 Gaps (미검증)

1. **`--root` 미지정 폴백의 산출물 경로는 스모크로 확인하지 않았다.** 폴백 자체는 단위 테스트(`resolveFeedbackRoot("", nested)` → 프로젝트 루트)와 앰비언트 실행으로 관측했고, 산출물(로그·큐) 생성은 **명시 `--root`** 로만 관측했다. Bash 도구의 cwd 가 호출마다 워크트리 루트로 복귀해 scratch 안에서 cwd 기반 실행을 유지할 수 없었다.
2. **프로덕션 소비자 부재.** `moai feedback` 을 실제로 호출하는 것은 아직 없다 — 스킬 본문(M6)이 그 소비자다. 따라서 "스킬이 제목을 넘긴다"는 M6 의 AC-F-019 소관이며 M5 는 넘길 **수단**만 관측했다.
3. **`golangci-lint` 는 `./internal/cli/...` 만 돌렸다.** 리포 전체 lint(M9)는 아직.
4. **`make build` / 템플릿 중립성 / docs-site** 는 M5 범위 밖(M8).
5. **전체 스위트(`go test ./...`)는 로컬에서 돌리지 않았다** — CLAUDE.local.md §4 규율. 전 패키지 판정은 CI 몫.
6. **다중 프로세스 큐 경합**은 여전히 미관측(M4 잔여 위험 그대로). 스모크는 순차 1프로세스다.
7. **Windows 실동작 미관측.** `GOOS=windows go vet` 은 컴파일만 증명하며, Windows 에서의 경로·권한 동작은 CI 매트릭스 몫.

## §5 Residual-risk (잔여 위험)

1. **fail-closed 정책 로드가 기존 사용자에게 새 실패 지점이 된다.** `security.yaml` 이 깨져 있던 프로젝트는 지금까지 탐지기의 fail-open 덕분에 아무 증상이 없었으나, `moai feedback scrub` 은 거기서 실패한다. 의도된 방향(공개 채널로 나가는 텍스트의 마스킹이 조용히 약해지는 것보다 낫다)이지만, 사용자에게는 "피드백만 안 된다"로 보인다. 에러 메시지가 파일 경로와 파싱 위치를 그대로 싣는 것이 완화책이다.
2. **`queue enqueue` 는 stdin JSON 을 신뢰한다.** `Result` 로 파싱되고 `verdict == ok` 이면 적재한다 — 손으로 만든 JSON 을 넣으면 스크럽을 거치지 않은 텍스트가 큐에 들어갈 수 있다. 큐는 그 자체로 전송하지 않으므로(전송은 스킬 본문의 `gh issue create`) 즉시 유출은 아니지만, 재전송 시점에 **큐 내용을 다시 스크럽하지 않는다면** 유출 경로가 된다. 파이프라인이 멱등이므로 M6 는 재전송 직전에 다시 스크럽할 수 있고, 그렇게 하는 것이 안전하다.
3. **잠금 아티팩트 잔존.** `queue.lock` 은 프로세스가 비정상 종료하면 남고 자동 해제되지 않는다(M4 판단 그대로). CLI 동사가 생기면서 사용자가 직접 중단(Ctrl-C)할 수 있는 표면이 넓어졌다 — stale-lock 정리 경로는 이 SPEC 범위 밖이다.
4. **`internal/cli` 스위트는 343초다.** M6 이후 이 패키지를 다시 돌릴 때 타임아웃 하한 600초를 지켜야 한다(300초로 걸면 통과하는 트리에서 FAIL 한다).
5. **스크러버 도입은 규약 강제이지 샌드박스가 아니다.** `moai feedback` 을 호출하지 않고 직접 이슈를 여는 경로는 그대로 열려 있다. M5 는 "마스킹이 이제 강제된다"를 만들지 않았다 — 통과시킬 **도구**를 만들었을 뿐이다(plan.md AP-12).
