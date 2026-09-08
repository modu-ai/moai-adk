# SPEC-CODEX-COVER-RESIDUAL-001 — sync-audit (card t519, lens --deep)

- 감사자: sync-auditor (독립 재판정 — 실행 단계 보고를 근거로 삼지 않고 전수 재측정)
- 측정 트리: `/Users/goos/MoAI/moai-adk-go/.claude/worktrees/t519` — 브랜치 `WT-codex-cover-residual`, HEAD `d1b66d06e` (측정 전후 불변 확인)
- Diff base: `bf779ecf2` (plan 진입 트리) → HEAD `d1b66d06e`
- 감사 모드: Claude 단독 (`audit_model` 미설정 → 교차모델 백엔드 호출 없음)
- 감사일: 2026-09-07

## Evaluation Report

SPEC: SPEC-CODEX-COVER-RESIDUAL-001
Overall Verdict: **PASS** (harmonic mean 96.8/100 — must-pass 차원 Functionality·Security 각각 독립 통과)

### Dimension Scores

| Dimension | Score | Verdict | Evidence (이번 실행, 이 트리에서 직접 재측정) |
|-----------|-------|---------|----------|
| Functionality (40%) | 98/100 | PASS | 6개 신규 테스트 전부 env-scrub 재실행 후 `--- PASS` 직접 관측 — `--- PASS: TestRunCodexReviewGate_InvalidStdinFailsOpen (0.00s)` / `_EmptyStdinFailsOpen` / `_HappyPathAllow` / `_HandlerErrorFailsOpen` / `_BlockVerdictPropagates` / `--- PASS: TestCodexSessionHandlePid (0.00s)`, `ok github.com/modu-ai/moai-adk/internal/cli 1.172s`. `^--- PASS: Test` 행 수 = 6(`$` 앵커 미사용), `no tests to run` 토큰 0회 — 공허 스윕 아님. AC-CCR-001..011 전부 PASS 근거 보유, AC-CCR-012는 DEFERRED로만 표기(PASS 주장 없음) |
| Security (25%) | 97/100 | PASS | 유니언 게이트 3중 재측정: `git diff --name-only bf779ecf2..d1b66d06e \| grep '\.go$' \| grep -v '_test\.go$'` → 무출력 rc=1, `git status --short` → 무출력, `git diff bf779ecf2..d1b66d06e -- internal/cli/codex_review_gate.go internal/cli/mcp_codex.go` → 무출력. 프로덕션 보안면 변경 0. fail-open 계약(malformed stdin / handler error → ALLOW) 4개 테스트로 보존 확인. 시크릿·주입면 그렙(`password\|secret\|token\|api[_-]?key\|exec\.Command\|os\.Setenv\|t\.Setenv`) 신규 테스트 코드에서 0매치(rc=1). 격리: `writeWorkflowYAML`이 `t.TempDir()` 사용, 실제 codex 바이너리 미실행(3개 시임 전부 스텁) |
| Craft (20%) | 94/100 | PASS | `go vet ./internal/cli/` rc=0 무출력 · `golangci-lint run --timeout=3m ./internal/cli/` → `0 issues.` · `gofmt -l internal/cli/` 무출력 — 전부 본 실행 재측정, pre-flight baseline(`0 issues.`)과 동일하므로 NEW 지적 0. **감사자 독자 변이체 M7이 RED**(아래 §Independent mutant). 추가로 `-race` 스코프 실행 GREEN(`ok ... 2.325s`) — 실행 단계가 Gap으로 남긴 항목을 이 감사가 스코프 한정으로 닫음. 감점: 커밋 본문 오타(F1, 이미 자체 기록), 패키지 커버리지 81.0%는 기관 목표 85% 미만(계승분, 범위 밖·정직 공시) |
| Consistency (15%) | 97/100 | PASS | 형제 파일 `multi_review_gate_wiring_test.go`의 배치·헬퍼 구조를 그대로 따름. 헬퍼 이름 충돌 0 — `newCodexGateCmd`(신규) vs `newGateCmd`(형제) 의도적 분리, `grep 'func newCodexGateCmd\|func newGateCmd'` 각 1행. `t.Parallel()` 호출 0회(주석 1행만 매치) — 패키지 레벨 시임 변조와 정합. `@MX:SPEC: SPEC-CODEX-COVER-RESIDUAL-001` 태그 존재. 기존 헬퍼(`withChangeDetector` / `withCodexLookPath` / `withCodexSession` / `codexSessionScript` / `writeWorkflowYAML` / `assertAllowJSON`) 재사용, 신규 중복 정의 0 |

Harmonic mean = 1 / (0.40/98 + 0.25/97 + 0.20/94 + 0.15/97) = **96.78** → 96.8

---

## 5-Section Evidence

### Claim

1. `bf779ecf2..d1b66d06e` diff는 Go 파일 중 테스트 파일 2개만 건드리며, 프로덕션 소스 2개의 blob은 base와 동일하다.
2. 신규 테스트 6개는 이 트리에서 실제로 실행되어 전부 통과하고, 셀렉터는 공허하지 않다.
3. 테스트는 물성이 있다 — 적용 변이체 목록에 없는 독자 변이체가 RED를 낸다.
4. 픽스처는 공허하지 않다(load-bearing 시임 스왑, LookPath 가드 배치, 병렬 금지, 이름 충돌 없음).
5. 커버리지 인용치(`runCodexReviewGate 92.3%`, `pid 100.0%`)는 최종 트리의 Go 내용에 귀속 가능하다.
6. S1 문서화 스킵(`out == nil`)은 코드 판독으로 도달 불가가 증명된다.
7. AC 표는 정직하다 — DEFERRED를 PASS로 주장하지 않고, 32행 그렙과 두 건의 정정을 감추지 않는다.
8. 범위 규율이 지켜졌다 — 인접 저커버 함수 2개는 손대지 않고 잔여 위험으로 공시됐다.
9. 최종 트리는 vet/lint/gofmt 청정이며 NEW 지적이 0이다.

### Evidence

**1) 유니언 게이트 (공격 항목 1)**

```
$ git diff --name-only bf779ecf2..d1b66d06e | grep '\.go$' | grep -v '_test\.go$'
exit=1                       ← 무출력

$ git status --short
[END status]                 ← 무출력

$ git diff bf779ecf2..d1b66d06e -- internal/cli/codex_review_gate.go internal/cli/mcp_codex.go
[END prod diff]              ← 무출력
```

diff의 `.go` 피연산자는 `internal/cli/codex_review_gate_wiring_test.go`(신규) + `internal/cli/mcp_codex_test.go`(추가) 2개 — 비어있지 않은 피연산자 위에서 필터가 0을 낸 것이므로 공허한 0이 아니다.

**2) 직접 관측 — env-scrub 스코프 실행 (공격 항목 2)**

```
$ unset MOAI_KANBAN MOAI_KANBAN_ID MOAI_KANBAN_LABEL MOAI_KANBAN_LEAD_ADDR MOAI_KANBAN_SETTINGS_INJECTED && go test -count=1 -run 'TestRunCodexReviewGate|TestCodexSessionHandlePid' -v ./internal/cli/
=== RUN   TestRunCodexReviewGate_InvalidStdinFailsOpen
--- PASS: TestRunCodexReviewGate_InvalidStdinFailsOpen (0.00s)
=== RUN   TestRunCodexReviewGate_EmptyStdinFailsOpen
--- PASS: TestRunCodexReviewGate_EmptyStdinFailsOpen (0.00s)
=== RUN   TestRunCodexReviewGate_HappyPathAllow
--- PASS: TestRunCodexReviewGate_HappyPathAllow (0.00s)
=== RUN   TestRunCodexReviewGate_HandlerErrorFailsOpen
--- PASS: TestRunCodexReviewGate_HandlerErrorFailsOpen (0.00s)
=== RUN   TestRunCodexReviewGate_BlockVerdictPropagates
--- PASS: TestRunCodexReviewGate_BlockVerdictPropagates (0.00s)
=== RUN   TestCodexSessionHandlePid
--- PASS: TestCodexSessionHandlePid (0.00s)
PASS
ok  	github.com/modu-ai/moai-adk/internal/cli	1.172s
```

`^--- PASS: Test` 행 6개. 실행 단계·레인의 실행을 인용하지 않고 이 감사가 직접 실행한 값이다.

추가(요구 밖, 실행 단계 Gap 축소): `-race` 동일 셀렉터 → `ok github.com/modu-ai/moai-adk/internal/cli 2.325s`.

**3) 독자 변이체 M7 — 적용 목록 밖 (공격 항목 3, 하중 부담 검사)**

적용 전 무결성: `shasum -a 256 internal/cli/codex_review_gate.go` → `9356669bdeb39f433897b7fc82c7e4cf036ca8ab24197301558b8c7bcedc3500` — mutants.md가 기록한 mutation 이전 해시와 **독립 일치**.

변이 내용(ledger의 M4와 다름: M4는 출력 전체를 빈 ALLOW로 치환, M7은 `decision`은 남기고 `reason`만 탈락):

```
codex_review_gate.go:201
-	return emitHookOutput(cmd.OutOrStdout(), out)
+	return emitHookOutput(cmd.OutOrStdout(), &hook.HookOutput{Decision: out.Decision})
```

변이 후 해시 `a0a79454fc7cb551f8ace0cb3cd82b75e2f8b93dbcf5c594d15b37ddb4643f0d`.

**RED (verbatim)**:

```
=== RUN   TestRunCodexReviewGate_BlockVerdictPropagates
    codex_review_gate_wiring_test.go:180: BLOCK must carry a non-empty reason, got "{\"decision\":\"block\"}\n"
--- FAIL: TestRunCodexReviewGate_BlockVerdictPropagates (0.00s)
FAIL
FAIL	github.com/modu-ai/moai-adk/internal/cli	1.064s
```

판정: **AC-CCR-005의 non-empty reason 단언은 decision 단언과 독립적으로 하중을 진다.** `decision` 단언(:177)은 침묵하고 `reason` 단언(:180)만 발화했다 — 두 단언이 한 덩어리로 붙어 우연히 초록인 구조가 아니라는 뜻이다.

완전 복원 증명:

```
$ shasum -a 256 internal/cli/codex_review_gate.go
9356669bdeb39f433897b7fc82c7e4cf036ca8ab24197301558b8c7bcedc3500   ← 원본 해시 복귀
$ git status --short
[END status]                 ← 무출력
$ git diff --stat
[END diffstat]               ← 무출력
```

변이체는 트리에 남아 있지 않다.

**4) 픽스처 공허성 감사 (공격 항목 4)**

| 요건 | 관측 |
|---|---|
| (a) AC-CCR-003이 `withChangeDetector(t, true)` + `codexLookPath` t.Fatal 가드 동반 | `codex_review_gate_wiring_test.go:107-110` 둘 다 존재. ledger M2-cf가 detector 제거 시 M2가 공허해짐을 실측으로 증명 |
| (b) AC-CCR-004가 3개 시임을 한 `t.Cleanup`으로 스왑 | `:140-144` — `codexRunner` / `codexLookPath` / `codexSession` 동시 스왑 후 단일 `t.Cleanup` 복원 |
| (c) `withCodexSession`과 LookPath t.Fatal 가드 병용 없음 (A6 권고) | `withCodexSession` 사용 테스트는 `_BlockVerdictPropagates` 1건이며 LookPath 가드 없음. `withCodexLookPath` t.Fatal 사용 테스트는 `_HappyPathAllow` 1건이며 세션 헬퍼 미사용 — 교차 0. `mcp_codex_test.go:108-119`에서 `withCodexSession`이 `codexLookPath`를 덮어쓰는 것 확인 |
| (d) `t.Parallel()` 없음 | `grep -n 't\.Parallel'` → 19행 주석 1건만. 호출 0 |
| (e) 헬퍼 이름 충돌 없음 | `func newCodexGateCmd` 1건(신규 파일), `func newGateCmd` 1건(형제 파일) — 패키지 스코프 충돌 없음 |

**5) pid 테스트 (공격 항목 5)**

`mcp_codex_test.go:676-686` 3개 arm 확인:

```go
if got := (*codexSessionHandle)(nil).pid(); got != 0 {          // 타입드 nil 리시버
if got := (&codexSessionHandle{}).pid(); got != 0 {             // 빈 구조체(conn nil)
if got := (&codexSessionHandle{conn: &fakeCodexConn{}}).pid(); got != fakeCodexConnPID {
```

`fakeCodexConnPID = 424242`(`codex_jobs_test.go:31`), `(*fakeCodexConn).pid()`가 이를 반환(`:35`). ledger M5a는 panic 트레이스(`mcp_codex.go:702` ← `mcp_codex_test.go:677`), M5b는 `pid on a nil handle = -1, want 0` / `pid on a handle with no conn = -1, want 0` 2행 FAIL로 기록 — 둘 다 요구된 형태. ledger 인용 행번호 `mcp_codex.go:702-704`를 최종 트리에서 재확인: `if h == nil || h.conn == nil {` / `return 0` / `}` 일치.

**6) 커버리지 귀속 (공격 항목 6)**

```
$ git diff --name-only a26feb41e..d1b66d06e | grep '\.go$'
grep_exit=1                  ← 무출력

$ git diff --name-only a26feb41e..d1b66d06e
.moai/reports/t519/coverage-after-run.log
.moai/reports/t519/coverage-after-targets.txt
.moai/reports/t519/mutants.md
.moai/reports/t519/sweep-six-tests.log
.moai/specs/SPEC-CODEX-COVER-RESIDUAL-001/progress.md
```

**명시 판정: `a26feb41e`에서 측정된 커버리지 수치는 최종 트리 `d1b66d06e`의 Go 내용에 귀속 가능하다.** 두 커밋 사이 델타는 증거 파일 4개 + progress.md뿐이며 `.go` 파일 0개다. 따라서 450초 전체 패키지 재측정은 불필요하고, 실행하지 않았다(디스패치 조건 충족).

인용치 대조(`coverage-after-targets.txt` vs `coverage-baseline-targets.txt`):

| 함수 | baseline (`bf779ecf2`) | after (`a26feb41e`) | 임계 | 판정 |
|---|---|---|---|---|
| `runCodexReviewGate` | 0.0% | 92.3% | ≥90.0% | PASS |
| `(codexSessionHandle).pid` | 66.7% | 100.0% | 100.0% | PASS |
| `reviewableFromPorcelain` | 83.3% | 83.3% | — | 불변(범위 밖) |
| `readHookInput` | 85.7% | 85.7% | — | 불변(범위 밖) |

**7) S1 문서화 스킵 — 감사자 독자 코드 판독 (공격 항목 7)**

`codex_review_gate.go:66-110`을 직접 읽어 `HandleCodexReviewGate`의 모든 return을 열거했다. `allow := &hook.HookOutput{}`(:67)은 non-nil이고, return은 정확히 7개:

| 행 | 반환값 | first value nil? |
|---|---|---|
| 69 | `allow, nil` | 아니오 |
| 72 | `allow, nil` | 아니오 |
| 75 | `allow, nil` | 아니오 |
| 80 | `allow, nil` | 아니오 |
| 101 | `allow, rpcErr` | 아니오 (err non-nil이나 out도 non-nil) |
| 104 | `&hook.HookOutput{Decision:…, Reason:…}, nil` | 아니오 |
| 109 | `allow, nil` | 아니오 |

`(nil, nil)`을 반환하는 경로가 없으므로 `runCodexReviewGate:198`의 `if out == nil` arm은 실제 핸들러를 통해 도달 불가다. 또한 :101 경로는 err non-nil이므로 호출부에서 `gateErr != nil` 분기가 먼저 흡수한다 — 이중으로 막힌다.

**progress.md §E.2.d의 증명은 완전하다** — 7개 return을 전부 열거했고, 코드 판독을 1차 근거로, 블록 커버리지를 보조 신호로 올바르게 위계화했다(도구 침묵을 근거로 쓰지 않음). 보조 신호도 재확인: 13문 함수에서 count=0 블록은 `198.16,200.3` 1개뿐이고 그 이웃 `198.2,198.16`·`201.2,201.47`은 count=1 — 12/13 = 92.3%와 정합.

**8) AC 표 정직성 (공격 항목 8)**

- AC-CCR-001..011: 각 행이 명령 + verbatim 관측을 인용하고, 채택 근거 변이체를 이름으로 지목. 확인 완료.
- AC-CCR-012: `**DEFERRED to sync**` — PASS로 주장하지 않음. §E.3 `ac_deferred_count: 1`, acceptance.md:36에도 `undecidable disposition (post-close only)`로 사전 기록. **주장 부풀림 없음.**
- E4 서브에이전트 경계 그렙: §E.2.h가 **32행 rc=0을 있는 그대로 보고**하고 "clean empty"로 부드럽게 만들지 않았다. 귀속 논거(비테스트 `.go` diff 0행 → 32행 전부 계승)를 함께 제시 — 이 감사의 유니언 게이트 재측정과 일치.
- `total_run_phase_files` 5→8 정정: `§ Corrections recorded, not silently fixed`에 기록. 재측정 `git diff --name-only 21a52d507..d1b66d06e | wc -l` → **8**, 항목화(테스트 2 + SPEC 2 + 증거 4)와 일치. 정정이 옳다.
- 커밋 본문 오타: `git log -1 --format=%B 22915e790` 12행에 `HandleCadexReviewGate` 실재 확인. 감추지 않고 기록했으며 `--amend` 금지 제약을 근거로 제시. **정직성 요건 충족**(오타 자체는 F1 advisory).

**9) 범위 규율 (공격 항목 9)**

`reviewableFromPorcelain` 83.3%, `readHookInput` 85.7% — baseline과 after 동일(위 표). 두 함수 모두 `§ Gaps`의 "adjacent under-covered functions were not measured for improvement … declined as scope creep (plan §A.6 D6)"에 잔여로 명시. 전달물로 주장한 흔적 없음.

**10) 최종 트리 품질 게이트 (공격 항목 10)**

```
$ go vet ./internal/cli/
vet_exit=0                   ← 무출력

$ golangci-lint run --timeout=3m ./internal/cli/ | tail -3
0 issues.

$ gofmt -l internal/cli/
[END gofmt]                  ← 무출력
```

분류: **NEW 0건**. pre-flight baseline이 `0 issues.`였으므로 계승 지적 뒤에 신규가 숨을 여지 자체가 없다.

### Baseline-attribution

이 보고의 모든 수치는 다음에 귀속된다.

| 축 | 값 |
|---|---|
| 트리 | `/Users/goos/MoAI/moai-adk-go/.claude/worktrees/t519` |
| HEAD | `d1b66d06e` (`git rev-parse HEAD`, 감사 개시 시점) |
| 브랜치 | `WT-codex-cover-residual` (`git branch --show-current`) |
| diff base | `bf779ecf2` |
| 커버리지 인용 출처 | `a26feb41e` 측정치 — 위 (6)에서 `.go` 델타 0으로 최종 트리 귀속 성립 |
| 프로덕션 무결성 | `codex_review_gate.go` SHA256 `9356669b…` (감사 시작·변이 후 복원 모두에서 관측) |
| 테스트 실행 환경 | env-scrub 단일 복합 호출, `-count=1` (캐시 무효화) |

carry-over 수치 0건 — 실행 단계 보고의 어떤 숫자도 재측정 없이 인용하지 않았다.

### Gaps — 이 감사가 관측하지 않은 것

- **전체 패키지 커버리지 재측정 미실시.** `.go` 델타 0 증명으로 귀속이 성립해 디스패치가 명시적으로 면제한 항목이다. 따라서 `81.0%` 패키지 수치는 실행 단계의 측정을 귀속 근거와 함께 수용한 값이며, 이 감사가 독립 재측정한 값이 아니다.
- **`go test ./...` 전체 스위트 미실행.** 로컬 전체 스위트 금지 규율(2026-08-15 load 413 사고). 전 패키지 판정은 CI 몫.
- **`-race`는 스코프 셀렉터에만 적용.** 패키지 전체 race 재측정은 하지 않았다.
- **ledger의 M1/M1b/M2/M3a/M3b/M4/M5a/M5b 8건을 재현하지 않았다.** 행번호 정확성(186/188/191/195/196/201, 702-704)과 사전 해시 일치는 독립 확인했으나, RED 출력 자체는 재현하지 않고 독자 변이체 M7 1건으로 물성을 독립 검증했다.
- **CodeRabbit / CI 판정 미확인.** 이 카드는 아직 push되지 않았고(`pushed: false`), sync close도 일어나지 않았다. CI 판정은 develop push 이후의 사안이다.
- **AC-CCR-012 미검증.** sync close 커밋이 아직 없다 — 구조적으로 이 시점에 판정 불가.

### Residual-risk — 관측했음에도 여전히 틀릴 수 있는 것

- **S1 스킵 증명은 오늘의 핸들러에 대한 코드 판독이다.** 장래 `HandleCodexReviewGate`가 `(nil, nil)`을 반환할 수 있게 바뀌면 증명이 조용히 무효가 되며 이를 잡는 기계 가드는 없다. 실행 단계도 같은 위험을 공시했다 — 동의한다.
- **`t.Parallel()` 금지는 파일 헤더 주석으로만 강제된다.** 장래 작성자가 어느 한 테스트에 추가하면 순서 의존 플레이크가 생기고, 로컬 초록으로는 드러나지 않는다.
- **`fakeCodexConn.pid()`는 상수를 반환한다.** AC-CCR-006의 세 번째 arm은 인터페이스 디스패치 도달을 증명할 뿐, 실제 서브프로세스 pid 판독의 정확성을 증명하지 않는다(그것은 `TestRealCodexConnPid`의 소관).
- **92.3%는 함수 단위 수치다.** `internal/cli` 패키지의 81.0%나 여전히 0.0%인 다른 함수들에 대해 아무것도 말하지 않는다.
- **sync close 이후의 착지 순서.** t488 불변(닫기가 마지막 쓰기)은 아직 판정 대상이 아니다 — close 커밋 후 `git diff --name-only <close>..HEAD | grep '\.go$'`가 무출력이어야 한다는 조건은 sync 단계에서 확인되어야 한다.

---

## Findings

| id | severity | blocking? | file:line | 결함 | 필요한 수리 |
|---|---|---|---|---|---|
| F1 | Low | **advisory** | 커밋 `22915e790` 본문 12행 | 커밋 메시지 본문 오타 `HandleCadexReviewGate` (→ `HandleCodexReviewGate`) | 수리 불필요. `--amend`가 plan §D 제약 7로 금지돼 있고, progress.md `§ Corrections recorded` 절에 이미 정직하게 기록됐다. 코드·문서 본문에는 오타가 없다(§E.2.d는 올바른 철자) |
| F2 | Low | advisory | `codex_review_gate.go:198` | S1 스킵의 도달 불가 증명을 지키는 기계 가드가 없다 | 후속 카드 후보. 지금 도입하면 도달 불가 분기를 위한 방어 코드가 되어 Enforce Simplicity와 충돌 — 잔여 위험 공시가 현 시점의 올바른 처리 |
| F3 | Low | advisory | `internal/cli` 패키지 | 패키지 커버리지 81.0%가 기관 목표 85% 미만 | 이 SPEC의 범위 밖(계승분, baseline 80.9%). progress.md에 "recorded and NOT gated"로 정직 공시됨. 별도 카드 소관 |
| F4 | Info | advisory | `codex_review_gate_wiring_test.go` 전반 | `-race` 검증이 실행 단계 Gap으로 남아 있었다 | 이 감사가 스코프 한정으로 닫았다(`ok ... 2.325s`). 추가 조치 불필요 |

**Blocking findings: 0건.** 4건 모두 advisory이며, 전부 실행 단계가 이미 공시했거나(F1·F2·F3) 이 감사가 그 자리에서 닫았다(F4). Finding-consumption discipline에 따라 advisory 목록만으로 PASS가 FAIL로 뒤집히지 않으며, 이들 중 어느 것도 자동 수리 라우팅 대상이 아니다.

---

## Deep-lens 공격 항목 총괄 (10/10)

| # | 항목 | 판정 |
|---|---|---|
| 1 | 유니언 게이트 (3중) | 통과 — 3개 명령 전부 무출력, 피연산자 비공허 |
| 2 | 직접 관측 6 PASS | 통과 — 감사자 직접 실행, 6행 |
| 3 | 독자 변이체 | **RED 발화** — 적용 목록 밖 M7, 완전 복원 증명 |
| 4 | 픽스처 공허성 (a~e) | 통과 — 5개 하위 요건 전부 충족 |
| 5 | pid 3-arm | 통과 — 타입드 nil / 빈 구조체 / fake conn 확인, ledger 행번호 재확인 |
| 6 | 커버리지 귀속 | 통과 — `.go` 델타 0, 명시 판정 기재, 450초 재측정 면제 성립 |
| 7 | S1 스킵 코드 판독 | 통과 — 감사자 독자 판독으로 7개 return 전수 확인, 증명 완전 |
| 8 | AC 표 정직성 | 통과 — DEFERRED 유지, 32행 무가공 보고, 정정 2건 기록 |
| 9 | 범위 규율 | 통과 — 인접 2함수 불변, 잔여 위험으로 공시 |
| 10 | vet/lint/gofmt | 통과 — 전부 청정, NEW 0 |

---

VERDICT: PASS score=96.8 blocking=0 advisory=4
