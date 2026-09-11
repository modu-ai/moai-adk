# sync-audit 델타 재감사 — card t525, SPEC-SPECLINT-GATE-SIGNAL-001

- 감사자: sync-auditor (델타 범위 재감사)
- 대상 트리: `.claude/worktrees/t525`, 브랜치 `WT-speclint-red`, HEAD `f580100fa44cb12ce33ebd5964ed21cfa87f4da0`
- 범위: 수리 델타 `cb1b9a16d..f580100fa` 만. 이전 감사(`.moai/reports/t525/sync-audit.md`, HEAD `ed4242845`)에서 PASS 로 판정한 AC 는 델타가 바꾸지 않는 한 다시 판정하지 않는다.
- 감사 창 동안 이 트리의 작성자는 감사자 하나였다. 시작·종료 시 `git status --porcelain | wc -l` = `0`, HEAD 불변(`f580100fa…`). 이 파일이 감사자가 트리에 남긴 유일한 쓰기다(커밋·스테이징 없음).

## 판정 요약

| 항목 | 판정 |
|---|---|
| F1 (blocking) | **CLOSED** |
| 델타 판정 | **PASS** |
| 새 차단 결함 | 없음 (새 발견 3건, 모두 optional) |
| 카드 전체 sync-audit | **PASS 로 전환** — 이전 FAIL 의 유일한 사유가 F1 이었다. AC-SLGS-011 의 CI 로그 절반은 여전히 develop 푸시 뒤에만 관측 가능한 리드 소관이다(결함 아님). |

델타 4커밋:

```
$ git log --oneline cb1b9a16d..HEAD
f580100fa docs(SPEC-SPECLINT-GATE-SIGNAL-001): correct deferred AC-SLGS-011 half and note visible flag rejections (card t525)
44f937400 test(t525): record F1 slot RED/GREEN run (card t525)
25f689832 fix(SPEC-SPECLINT-GATE-SIGNAL-001): print --json --sarif rejection to stderr (card t525)
85cf76b2d fix(SPEC-SPECLINT-GATE-SIGNAL-001): print baseline flag rejections to stderr (card t525)
```

코드 변경은 `internal/cli/spec_lint.go`(+23/−16), `internal/cli/spec_lint_test.go`(+45/−9), `CHANGELOG.md`(1줄) 세 파일뿐이고, 나머지는 증거 파일이다.

---

## 1. Claim

1. **F1 은 닫혔다.** 이전 감사가 지목한 거절 경로 전부와 기준선 로드·쓰기 실패가, 트리에서 새로 빌드한 바이너리를 교차 프로세스로 부를 때 stderr 에 원인 한 줄을 남긴다. 거절 경로는 모두 exit 3, stdout 0바이트, stderr 1줄이다. 공백 사유 거절 뒤 기준선 파일의 sha256 은 그대로다. 정상 `--baseline` 실행은 exit 0, `baseline: OK`, stderr 0바이트다. 같은 메시지가 두 번 찍히는 곳은 없다.
2. **테스트가 이제 사용자가 보는 것을 단언한다.** `TestSpecLintBaseline_FlagContractRejections` 는 exit 3, 빈 stdout, 원인을 담은 stderr 정확히 한 줄을 요구한다. 레인의 RED 증거는 신뢰할 만하다. 헬퍼 변경은 이 파일의 다른 테스트를 약화하지 않고, 양성 단언 몇 개를 오히려 강화한다.
3. **CHANGELOG F5 는 정정됐고**, 새 문장은 코드와 일치한다(열거가 완전하지 않다는 사소한 점은 N2).
4. **델타가 들여온 회귀는 관측되지 않았다.** exit 코드 불변, fang/cobra 경유 중복 출력 없음, windows 교차 빌드 exit 0, `go vet`·`golangci-lint`·`gofmt` 청결.

## 2. Evidence

### 2.1 판정 빌드 (이 감사가 직접 만든 바이너리)

```
$ go version && go build -o <scratch>/t525d/moai ./cmd/moai; echo build_exit=$?
go version go1.26.8 darwin/arm64
build_exit=0

$ <scratch>/t525d/moai version; echo exit=$?
 v3.1.3   none   built unknown
exit=0
```

`<scratch>` = `/private/tmp/claude-501/-Users-goos-MoAI-moai-adk-go/7e33c4bd-a91e-4a67-bf0c-15ad255da456/scratchpad`. ldflags 없이 빌드해서 commit 필드가 `none` 이다. 빌드 시점 트리는 HEAD `f580100fa` 에서 추적 수정 0(`git status --porcelain | wc -l` = `0`)이었고, 바이너리는 경로로만 호출했다(설치된 `moai` 미사용). 판정 빌드 좌표: 트리 `f580100fa` = 빌드 소스(VCI §2.2).

### 2.2 교차 프로세스 탐침 — stdout/stderr 분리

각 탐침은 워크트리 루트에서 `<bin> spec lint <args> >pN.out 2>pN.err; echo exit=$?` 로 실행했다. 공백 사유 탐침(P3)은 추적 파일을 건드리지 않도록 기준선 사본(`copy-baseline.json`)에 대고 돌렸고, 빈 사유(P3b)는 추적 파일에 대고 돌렸다.

| # | 인자 | exit |
|---|---|---|
| P1 | `--baseline .moai/spec-lint-baseline.json --strict` | 3 |
| P2 | `--baseline <scratch>/t525d/absent.json` (없는 파일) | 3 |
| P3 | `--baseline <scratch>/t525d/copy-baseline.json --update-baseline --reason "   "` | 3 |
| P3b | `--baseline .moai/spec-lint-baseline.json --update-baseline --reason ""` | 3 |
| P4 | `--baseline .moai/spec-lint-baseline.json --reason x` | 3 |
| P5 | `--update-baseline --reason x` | 3 |
| P6 | `--baseline .moai/spec-lint-baseline.json --json` | 3 |
| P7 | `--baseline .moai/spec-lint-baseline.json --sarif` | 3 |
| P8 | `--json --sarif` | 3 |
| P9 | `--baseline <scratch>/t525d/bad-baseline.json` (내용 `{not json`) | 3 |
| P10 | `--baseline <scratch>/t525d/afile/b.json --update-baseline --reason probe` (`afile` 은 일반 파일) | 2 |
| P11 | `--baseline .moai/spec-lint-baseline.json` (정상 대조군) | 0 |

스트림 크기와 stderr 원문 (verbatim):

```
== p1 stdout_bytes=0 stderr_bytes=133 stderr_lines=1
cannot use --baseline and --strict together: both gate on non-advisory warnings, and --baseline supersedes --strict for that purpose
== p2 stdout_bytes=0 stderr_bytes=199 stderr_lines=1
--baseline file not found: /private/tmp/claude-501/-Users-goos-MoAI-moai-adk-go/7e33c4bd-a91e-4a67-bf0c-15ad255da456/scratchpad/t525d/absent.json (create it with --update-baseline --reason "<text>")
== p3 stdout_bytes=0 stderr_bytes=109 stderr_lines=1
--update-baseline requires a non-empty --reason "<text>"; an unexplained rebaseline is baseline manipulation
== p3b stdout_bytes=0 stderr_bytes=109 stderr_lines=1
--update-baseline requires a non-empty --reason "<text>"; an unexplained rebaseline is baseline manipulation
== p4 stdout_bytes=0 stderr_bytes=51 stderr_lines=1
--reason is only meaningful with --update-baseline
== p5 stdout_bytes=0 stderr_bytes=70 stderr_lines=1
--update-baseline requires --baseline <path> naming the file to write
== p6 stdout_bytes=0 stderr_bytes=81 stderr_lines=1
cannot use --baseline with --json (the JSON payload carries no baseline verdict)
== p7 stdout_bytes=0 stderr_bytes=83 stderr_lines=1
cannot use --baseline with --sarif (the SARIF payload carries no baseline verdict)
== p8 stdout_bytes=0 stderr_bytes=39 stderr_lines=1
cannot use --json and --sarif together
== p9 stdout_bytes=1142289 stderr_bytes=219 stderr_lines=1
spec lint: parse baseline "/private/tmp/claude-501/-Users-goos-MoAI-moai-adk-go/7e33c4bd-a91e-4a67-bf0c-15ad255da456/scratchpad/t525d/bad-baseline.json": invalid character 'n' looking for beginning of object key string
== p10 stdout_bytes=1142289 stderr_bytes=291 stderr_lines=1
spec lint: write baseline "/private/tmp/claude-501/-Users-goos-MoAI-moai-adk-go/7e33c4bd-a91e-4a67-bf0c-15ad255da456/scratchpad/t525d/afile/b.json": open /private/tmp/claude-501/-Users-goos-MoAI-moai-adk-go/7e33c4bd-a91e-4a67-bf0c-15ad255da456/scratchpad/t525d/afile/b.json: not a directory
== p11 stdout_bytes=1142480 stderr_bytes=0 stderr_lines=0

e9124a70ec0cc3359917a78e39364f8f025a971a673294102342b7b2115455f6  copy-baseline.json
e9124a70ec0cc3359917a78e39364f8f025a971a673294102342b7b2115455f6  /Users/goos/MoAI/moai-adk-go/.claude/worktrees/t525/.moai/spec-lint-baseline.json
```

P3·P3b 전후 sha256: 탐침 전 `shasum -a 256 .moai/spec-lint-baseline.json` = `e9124a70ec0cc3359917a78e39364f8f025a971a673294102342b7b2115455f6`, 탐침 후 사본과 추적 파일 모두 같은 값(위 두 줄). 거절된 재기준은 파일을 건드리지 않았다.

P9·P10 은 lint 가 먼저 돌아 표를 stdout 에 쓴 뒤 기준선 단계에서 실패한다. 그래서 stdout 이 비어 있지 않은 게 정상이고, 요점은 원인이 stderr 에 한 번만, 그리고 stdout 에는 섞이지 않고 찍혔는지다.

정상 대조군(P11)과 중복 여부:

```
$ tail -5 p11.out
0 error(s), 3133 warning(s)

baseline: OK — .moai/spec-lint-baseline.json
  inventory: 3133 warning(s) total (advisory included), 0 non-advisory tracked across 0 recorded rule(s)
  recorded at: 4ac93f755 (2026-09-11)
---
$ /usr/bin/grep -c "baseline: OK" p11.out            → 1
$ /usr/bin/grep -c "parse baseline" p9.out           → 0
$ /usr/bin/grep -c "write baseline" p10.out          → 0
$ /usr/bin/grep -c "spec lint:" p9.err p10.err       → p9.err:1  p10.err:1
--- control
$ /usr/bin/grep -c "warning(s)" p11.out              → 2
```

0 이라는 수가 공허하지 않음을 보이려고 같은 도구로 양성 대조(`warning(s)` → 2, `spec lint:` → 1)를 함께 쟀다. P11 stdout 1142480바이트는 개발자 교차 프로세스 증거 `c4`(stdout 1142480바이트, `.moai/reports/t525/f1/commands.txt`)와 같다.

### 2.3 exit-3 경로의 전수 — 남은 무음 경로가 없다

```
$ /usr/bin/grep -c 'exitCodeError{code: 3' internal/cli/spec_lint.go   → 1
$ /usr/bin/grep -n 'exitCodeError{code: 3' internal/cli/spec_lint.go
415:	return &exitCodeError{code: 3, msg: msg}
$ /usr/bin/grep -c 'argumentError(' internal/cli/spec_lint.go          → 13
```

유일하게 남은 exit-3 리터럴은 `argumentError` 본문(`spec_lint.go:412-416`, stderr 에 먼저 쓰고 반환)이다. exit-3 을 내는 모든 호출(76, 168, 171, 174, 182, 185, 190, 194, 243, 384, 390, 402행)이 이 함수를 거친다. exit-2 쓰기 실패(`:216-220`)는 `Fprintln(cmd.ErrOrStderr(), msg)` 뒤 반환한다. 그러니 `:346-351` 에서 고친 주석("Every exit-3 path in this command therefore writes its own diagnostic through argumentError")은 코드와 맞는다.

중복 출력이 없는 기제: `argumentError` 는 `*exitCodeError` 를 반환하고, `moaiErrorHandler`(`internal/cli/fang.go:140-145`)는 `ResolveExitCode` 가 참이면 아무것도 렌더하지 않는다. 그래서 명령이 직접 쓴 한 줄만 남는다. P1–P10 의 `stderr_lines=1` 이 이를 교차 프로세스로 확인한다.

### 2.4 테스트 쪽 — 레인 슬롯 증거 판독 (이 감사는 `internal/cli` 테스트를 돌리지 않았다)

슬롯이 반납돼 있어 `internal/cli` 테스트는 한 건도 돌리지 않았다. 아래는 레인이 남긴 `.moai/reports/t525/f1-slot/` 를 읽고 대조한 결과다.

- **RED 뮤턴트의 정체**: `mutant-sha.txt` 에 `git hash-object` = `f088f2b913efa4150abc071d8b88130067ac6d98` 가 두 번 적혀 있고, 이 값은 델타 diff 의 `index f088f2b91..d7a283afb` 에서 수리 전 blob 과 같다. 수리 전 파일 전체로 되돌린 상태에서 새 테스트를 돌렸다는 뜻이다.
- **RED 결과** (`red.txt`, `red.exit` = `exit=1`): 서브테스트 8개 모두 같은 줄에서 실패했다.
  ```
      spec_lint_test.go:942: the rejection must be written to stderr; stderr is empty (the user sees nothing)
  ...
  --- FAIL: TestSpecLintBaseline_FlagContractRejections (0.25s)
  FAIL	github.com/modu-ai/moai-adk/internal/cli	1.572s
  ```
  수리 전 코드도 exit 3 은 맞게 냈으므로 앞의 `code != 3` 가드는 통과했고, 실패는 수리 대상인 "stderr 가 비어 있다"는 성질 자체에서 났다. 실패 이유가 결함과 정확히 겹친다. 신뢰할 만한 RED 다.
- **GREEN** (`green.txt`, `green.exit` = `exit=0`): 선택자 9/9 PASS. 파일 계열(`-run 'TestSpecLint'`, `green-file.txt`) 계수:
  ```
  $ /usr/bin/grep -c "^=== RUN" green-file.txt   → 33
  $ /usr/bin/grep -c -- "--- PASS" green-file.txt → 33
  $ /usr/bin/grep -c -- "--- FAIL" green-file.txt → 0
  $ /usr/bin/grep -c -- "--- SKIP" green-file.txt → 0
  ok  	github.com/modu-ai/moai-adk/internal/cli	18.348s
  ```
- **슬롯 실행 트리와 현재 HEAD 의 동일성**: 슬롯은 HEAD `25f689832` 에서 돌았다. 그 뒤 코드는 바뀌지 않았다.
  ```
  $ git diff --stat 25f689832 HEAD -- internal/
  (출력 없음)
  $ shasum -a 256 internal/cli/spec_lint.go
  6f2e27f5533a46c8e10df57b2bd8f4d3b7dc58d32db2f67a8770aca0f9ef9d85  internal/cli/spec_lint.go
  $ git hash-object internal/cli/spec_lint.go
  d7a283afb3b8284a68cebc759e9183ec31583a2a
  ```
  sha256 은 `summary.md` 가 기록한 복원 후 값(`6f2e27f5…`)과 같고, blob 은 델타 diff 의 수리 후 index(`d7a283afb`)와 같다.

**헬퍼 변경이 다른 테스트를 약화하는가 — 아니다.** `slRunLint` 는 이제 `*exitCodeError` 일 때 `err.Error()` 를 덧붙이지 않는다. 이 파일에서 `slRunLint` 를 쓰는 테스트를 모두 대조했다(`spec_lint_test.go:624-997`).

- 코드 3 을 기대하는 `TestSpecLintUpdateBaseline_RejectsMissingOrEmptyReason`(`:748-772`)은 exit 코드와 파일 sha 만 단언하고 메시지 텍스트는 보지 않는다. 영향 없음.
- 코드 1 경로의 양성 단언(`EXCEEDED`, `ERROR-GATED`, `recorded 0`, `SpecsDirMissingSpecFile`)은 예전에는 반환 오류의 문구(`"spec lint: baseline exceeded"` 등)만으로도 통과할 수 있었다. 이제는 명령이 실제로 찍은 줄(`spec_lint.go:235,250,256,264,266`)에서만 통과한다. **강화**다.
- 음성 단언(`ErrorFailsRegardlessOfBaseline`·`ErrorOutranksASimultaneousIncrease` 의 "EXCEEDED 가 없어야 한다")은 오류 문구가 `"spec lint: error-severity findings detected"` 라 예전에도 EXCEEDED 를 담지 않았다. 문구를 떼어낸다고 느슨해지지 않는다.
- `TestSpecLintUpdateBaseline_ErrorStillExitsOne` 의 `"unknown flag"` 가드는 cobra 의 비-ExitCoder 오류를 겨눈다. 헬퍼는 이 경우 여전히 `err.Error()` 를 stderr 에 덧붙이므로(`:617`, fang 기본 렌더러를 흉내) 가드가 살아 있다.

패키지의 다른 파일도 확인했다. spec lint 의 exit 코드를 보는 `exitcode_contract_test.go:60-74,191-204` 는 `assertExitCode(…, 3)` 만 단언하고, 스트림을 버퍼로 받아 내용을 보지 않는다. 델타는 exit 코드를 바꾸지 않았으므로 영향이 없다(판독만, 실행하지 않음).

### 2.5 CHANGELOG

```
$ git diff cb1b9a16d..HEAD -- CHANGELOG.md   (요지)
- … 12 acceptance criteria; one of them (the wiring-completeness half of `AC-SLGS-011`, …
+ … Every rejected baseline flag combination — `--baseline` with `--strict`, a missing baseline file, a blank `--reason`, `--json` or `--sarif` with `--baseline` — now prints a one-line reason to stderr before exiting 3, rather than exiting 3 silently. 12 acceptance criteria; one of them (the CI-log half of `AC-SLGS-011`, …
```

- F5: "the CI-log half of `AC-SLGS-011`" — 이전 감사가 요구한 문구와 글자 그대로 같다. **정정됨.**
- 새 문장: 열거된 다섯 경우는 P1(`--strict`), P2(없는 파일), P3/P3b(공백·빈 사유), P6(`--json`), P7(`--sarif`) 로 각각 exit 3 + stderr 한 줄이 관측됐다. "rather than exiting 3 silently" 는 이전 감사의 0바이트 실측과 수리 커밋 본문에 부합한다. **정확하다.** 열거가 전부는 아니다(N2).

### 2.6 회귀 점검

```
$ GOOS=windows GOARCH=amd64 go build ./... ; echo win_exit=$?
win_exit=0

$ gofmt -l internal/cli/spec_lint.go internal/cli/spec_lint_test.go; echo gofmt_exit=$?
gofmt_exit=0
$ go vet ./internal/cli/ ; echo vet_exit=$?
vet_exit=0

$ golangci-lint run ./internal/cli/ 2>&1 | tail -15
0 issues.
lint_exit=0
```

`gofmt -l` 은 아무것도 출력하지 않았다(포맷 차이 없음). exit 코드: 모든 거절은 전후 모두 3, 쓰기 실패는 2, 정상은 0. 메시지 텍스트도 수리 전 `msg` 필드와 한 글자도 다르지 않다(diff 에서 문자열 리터럴이 그대로 옮겨졌다).

## 3. Baseline-attribution

- 모든 측정은 이 감사 창 안에서, 워크트리 `.claude/worktrees/t525` HEAD `f580100fa44cb12ce33ebd5964ed21cfa87f4da0`, 추적 수정 0 인 트리에 대해 수행했다. 창 시작과 끝에서 `git rev-parse HEAD` 가 같았다.
- 교차 프로세스 탐침은 위 트리에서 `go1.26.8` 로 이 감사가 직접 빌드한 바이너리(판정 빌드 = 트리)로만 쟀다.
- 테스트 결과(RED 8/8 FAIL, GREEN 9/9·33/33)는 **레인의 측정**이다. 이 감사의 측정이 아니다. 슬롯 실행 트리(`25f689832`)와 현재 HEAD 의 `internal/` 차이가 없다는 것(`git diff --stat` 무출력)과 blob/sha256 일치는 이 감사가 쟀다.
- 이전 감사의 수치(커버리지 81.4%, AC 11개 PASS)는 재측정하지 않은 인용이다.

## 4. Gaps (관측하지 않은 것)

- `internal/cli` 테스트를 한 건도 돌리지 않았다(슬롯 반납). 테스트 쪽 판정은 전적으로 레인 슬롯 증거를 판독한 결과다.
- `internal/cli` 패키지의 다른 테스트 파일 전체(`exitcode_contract_test.go` 등)는 판독만 했고 실행하지 않았다. 레인 슬롯도 `TestSpecLint` 계열(33)만 돌렸다.
- 기준선 로드 실패(exit 3)·쓰기 실패(exit 2) 경로에는 테스트 단언이 없다. 이 감사의 교차 프로세스 탐침 P9/P10 이 유일한 관측이다(N1).
- windows·linux 실제 실행과 CI. windows 는 교차 컴파일만 확인했다.
- 커버리지 재측정 없음. Craft 차원의 85% 판정은 이전 감사와 같이 UNVERIFIED 로 남는다.
- 교차 모델 감사(`audit_multi`/codex/glm)는 부르지 않았다. `uncommittedChanges` 가 비어 있고, `baseBranch` 는 서버가 원격 기본 헤드를 기준으로 잡아 카드 델타 `cb1b9a16d..f580100fa` 를 겨눌 수 없다. 틀린 트리에 대한 판정을 만들지 않으려고 생략했다.
- AC-SLGS-011 의 CI 로그 절반(develop 푸시 뒤에만 존재).

## 5. Residual-risk

- 로드·쓰기 실패 경로는 지금 옳게 동작하지만(P9/P10), 이를 지키는 테스트가 없다. 누가 `runBaselineGate` 를 고치면서 stderr 쓰기를 빠뜨려도 스위트는 초록일 수 있다(N1).
- 레인 슬롯은 부하 22.61 에서 돌았다. 결과가 결정적인 성질(스트림 내용)이라 부하 탓에 판정이 뒤집힐 가능성은 낮지만, 이 감사가 독립적으로 재현한 것은 아니다.
- 새 헬퍼가 비-ExitCoder 오류에 대해 fang 기본 렌더러를 "흉내"만 낸다. 실제 fang 은 스타일을 입힌 박스로 렌더하므로, 텍스트 포함 여부만 보는 단언에는 문제없지만 줄 수를 세는 단언을 비-ExitCoder 경로에 쓰면 실제 바이너리와 어긋날 수 있다. 현재 그런 단언은 없다.
- 병합 트리 미관측. 병합 창에서 develop 을 흡수한 뒤에는 병합 트리에서 다시 재야 한다(§4.1 규율 5).

---

## 6. 차원 점수 (default 프로필, flat) — 델타 반영

| Dimension | Score | Verdict | Evidence |
|---|---|---|---|
| Functionality (40%) | 90/100 | PASS | F1 경계 사례("명확한 인자 오류") 충족: P1–P8 모두 `exit=3`, `stdout_bytes=0`, `stderr_lines=1`, 원인 이름 포함. P3 전후 sha256 `e9124a70…` 불변. P11 `exit=0`, `baseline: OK`, `stderr_bytes=0`. AC-SLGS-011 로그 절반은 여전히 푸시 뒤에만 관측 가능(리드 소관). |
| Security (25%) | 95/100 | PASS | 델타는 출력 경로만 바꾼다. 새 입력 표면이나 exec, 의존성 변경이 없다. stderr 에 찍히는 것은 사용자가 넘긴 경로와 파싱 오류뿐이다(P2/P9/P10 원문). |
| Craft (20%) | 80/100 | UNVERIFIED | `golangci-lint … 0 issues.`, `vet_exit=0`, `gofmt_exit=0`, `win_exit=0`. 레인 RED 8/8 → GREEN 9/9·33/33. 로드·쓰기 실패 경로 테스트 부재(N1). 커버리지 미재측정 → UNVERIFIED 유지. |
| Consistency (15%) | 90/100 | PASS | 이전 F1 이 지적한 "파일 안 `argumentError` 관례 이탈"이 해소됐다. exit-3 리터럴 1건 = `argumentError` 본문(`:415`), 호출 13건. 헬퍼는 기존 `runSpecLint` 문서 규칙과 같은 규칙을 따른다. |

must-pass 방화벽: Functionality PASS, Security PASS. 가중 조화평균 ≈ 89.5.

## 7. Findings (structured defect-list)

### F1 — 상태: **CLOSED**

이전 감사의 Required fix 세 항목을 대조했다.

| Required fix | 판정 | 근거 |
|---|---|---|
| (1) `validateBaselineFlags` 가 stderr 를 받아 `argumentError` 로 반환, 로드(exit 3)·쓰기(exit 2) 실패도 stderr 에 기록 | 충족 | diff `spec_lint.go:162-194,213-220,243`; §2.3 전수; P1–P10 |
| (2) 헬퍼가 `err.Error()` 를 덧붙이지 않게 하고, 수정 전 테스트가 붉어지는 것을 먼저 관측 | 충족 | diff `spec_lint_test.go:579-618`; `f1-slot/red.txt` 8/8 FAIL "stderr is empty" |
| (3) 빌드 바이너리로 `--baseline <없는 경로>` 교차 프로세스 확인 | 충족(범위 초과) | 개발자 `.moai/reports/t525/f1/c2-*` + 이 감사 P2 (및 P1–P11 전수) |

### 새 발견

- **N1** [Low] [optional] `internal/cli/spec_lint_test.go:898-953` — 기준선 **로드** 실패(`spec_lint.go:243`, exit 3)와 **쓰기** 실패(`:216-220`, exit 2)에는 테스트 단언이 없다. 두 경로 모두 수리 커밋이 고친 경로인데, 지금은 이 감사의 P9/P10 교차 프로세스 탐침만이 관측이다(레인 `summary.md` "미검증" 절도 같은 사실을 적었다). 확신도 high. — Required fix(선택): `FlagContractRejections` 와 같은 모양으로 두 경우(내용이 `{not json` 인 기준선 파일 → exit 3 / 일반 파일 아래 경로로 `--update-baseline` → exit 2)를 추가해 stderr 한 줄과 원인 단어를 단언한다. 이 경로들은 lint 를 먼저 돌리므로 stdout 비어 있음 단언은 빼야 한다.
- **N2** [Info] [optional] `CHANGELOG.md:12` — "Every rejected baseline flag combination — …" 뒤의 대시 목록이 전수처럼 읽히지만, 8가지 거절 모양 가운데 다섯만 든다. `--reason` 을 `--update-baseline` 없이 쓴 경우, `--update-baseline` 을 `--baseline` 없이 쓴 경우, `--json --sarif`(기준선 플래그는 아님), 깨진 기준선 파일(exit 3)이 빠졌다. 열거된 내용은 모두 사실이다(§2.5). 확신도 high. — Required fix(선택): "— for example …" 로 바꾸거나 빠진 경우를 덧붙인다.
- **N3** [Info] [optional] 커밋 `44f937400` — 트레일러 `Authored-By-Agent: manager-develop` 가 붙어 있는데, 리드 설명으로는 슬롯 실행을 레인 세션이 직접 했다. 귀속 필드는 값이 아니라 주장이므로, 실행 주체와 어긋나면 증거의 출처 추적이 흐려진다. 증거 내용(RED/GREEN)의 신뢰도에는 영향이 없다. 확신도 medium(실행 주체는 리드 진술에 기댐). — Required fix(선택): 이력은 다시 쓰지 말고, progress 기록에 "슬롯 실행: 레인 세션, 트레일러는 오기"를 한 줄 남긴다.

### 이전 감사의 optional 발견 (델타가 다루지 않음 — 상태 불변)

F2(워크플로 출력명 `strict`), F3(트리거 경로 밖 게이트 코드), F4(progress 548행 SHA·재기준 횟수), F6(AC-SLGS-004 출처 서술), F7(오류 상태 재기준 기록), F8(`internal/spec` 단위 테스트·커버리지 81.4%)은 모두 optional 이며 델타 범위 밖이라 다시 판정하지 않았다. 차단 사유가 아니다.

## 8. 종합

- **델타 판정: PASS.** F1 은 사용자 표면에서 교차 프로세스로 닫혔고, 테스트는 사용자가 보는 스트림을 단언하며, 그 RED 는 결함 자체를 이유로 붉었다. 델타가 들여온 회귀는 관측되지 않았다.
- **카드 전체 sync-audit: PASS 로 전환.** 이전 FAIL 의 유일한 차단 사유였던 F1 이 닫혔다. 남은 발견은 모두 optional(F2–F4, F6–F8, N1–N3)이다. AC-SLGS-011 의 CI 로그 절반은 develop 푸시 뒤 리드가 확인할 몫이다(결함 아님).
- 병합 창에서 develop 을 흡수하면 병합 트리에서 다시 재야 한다. 이 판정은 `f580100fa` 트리에 대한 것이다.
