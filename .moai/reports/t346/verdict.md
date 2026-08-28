# t346 run-phase verdict — SPEC-CI-DOCTOR-BIN-001

- card: t346
- branch: WT-ci-doctor-bin
- base: origin/develop `d566ecc75` (merged into this branch, merge commit `5345b76ed`)
- 측정 트리: `.claude/worktrees/t346`

## Claim

1. bin/moai 부재 트리에서 `TestRunDoctor_*` / `TestDoctorCmd_*` 9건이 통과한다 (AC-CDB-004 green cell).
2. bin 부재 스킵은 status `ok`이고 extractor를 호출하지 않으며 부재 경로와 처방을 이름으로 넣는다 (AC-CDB-001·002).
3. 바이너리가 있을 때의 fail 경로는 수정 없이 보존된다 (AC-CDB-003).
4. develop CI 적색은 **5개 계열**이고 본 카드가 닫는 것은 그중 1개다 — 나머지 4계열은 착지 후에도 남는다.

## Evidence

### E8 — RED (구현 이전 코드, bin 부재; 실측 2026-08-28)

`git checkout HEAD -- <2 files>` 로 원본 복원 후 실행, 이후 작업본 원복:

```
$ go test ./internal/cli/ -run 'TestRunDoctor_WithExport$|TestAgentEmitEmbed_MissingBinaryFails$' -count=1 -v
=== RUN   TestRunDoctor_WithExport
    coverage_improvement_test.go:715: runDoctor error: doctor: 1 check(s) failed
--- FAIL: TestRunDoctor_WithExport (4.68s)
=== RUN   TestAgentEmitEmbed_MissingBinaryFails
--- PASS: TestAgentEmitEmbed_MissingBinaryFails (0.00s)
FAIL	github.com/modu-ai/moai-adk/internal/cli	5.517s
```

구 테스트가 통과한다는 것이 결함을 **단언하고 있었다**는 증거다 (plan §B-기대역전).

### E1 — AC 매트릭스

| AC | 판정 | 명령 + 관측 |
|---|---|---|
| AC-CDB-001 | PASS | `go test ./internal/cli/ -run 'TestAgentEmitEmbed' -count=1 -v` → 11개 서브테스트 전부 `--- PASS`, 그중 `TestAgentEmitEmbed_MissingBinarySkips` (status ok + extractor 미호출 플래그 단언) |
| AC-CDB-002 | PASS | 라이브: `MOAI_EMBED_CHECK_BIN=/nonexistent/moai ./bin/moai doctor --check 'Agent Emit Embed'` → `ok  Agent Emit Embed  skipped: no readable binary to judge at /nonexistent/moai — 11 committed artifacts unjudged; build one with `make build` or aim the check elsewhere with MOAI_EMBED_CHECK_BIN=<path>` / `Pass 1 Warn 0 Fail 0` |
| AC-CDB-003 | PASS | 비회귀: `TestAgentEmitEmbed_(ExtractionErrorFails\|DriftFailsAndNamesPath\|PartialExtractionFails)` 수정 없이 PASS. 라이브 판정: `./bin/moai doctor --check 'Agent Emit Embed' -v` → `ok  11/11 embedded agent-emit artifacts match the committed set (moai)` — 스킵이 아니라 실제 11건 비교가 일어남 |
| AC-CDB-004 | PASS | RED-now = 위 E8 전문. GREEN = 아래 9/9 |

### AC-CDB-004 green cell — 선택자 매치 수까지 계수

```
$ go test ./internal/cli/ -count=1 -v -run 'TestRunDoctor_WithFix$|TestRunDoctor_WithExport$|TestRunDoctor_VerboseAndDetail$|TestRunDoctor_Verbose$|TestRunDoctor_ExportMode$|TestRunDoctor_AllFlags$|TestDoctorCmd_VerboseExecution$|TestDoctorCmd_ExportFlag$|TestDoctorCmd_Execution$'
--- PASS: TestRunDoctor_WithExport (1.92s)      --- PASS: TestRunDoctor_WithFix (1.83s)
--- PASS: TestRunDoctor_Verbose (1.89s)          --- PASS: TestRunDoctor_AllFlags (2.13s)
--- PASS: TestRunDoctor_VerboseAndDetail (1.97s) --- PASS: TestRunDoctor_ExportMode (2.37s)
--- PASS: TestDoctorCmd_Execution (2.48s)        --- PASS: TestDoctorCmd_ExportFlag (2.10s)
--- PASS: TestDoctorCmd_VerboseExecution (2.05s)
ok  	github.com/modu-ai/moai-adk/internal/cli	19.447s
```

`grep -cE '^--- PASS'` = **9** — 선택자가 0개를 고르고 초록을 낸 형태가 아님. 9는 CI 적색 목록의 doctor 계열 9종과 일치.

bin-present 대조 (`bin/moai` 복원 후 같은 9종 + embed 11종): `grep -cE '^--- PASS'` = **20**, `ok internal/cli 21.949s`.

### E2/E3/E5 — 게이트

- `go vet ./internal/cli/` → rc=0
- `GOOS=windows GOARCH=amd64 go build ./...` → rc=0
- `golangci-lint run --timeout=5m ./internal/cli/...` → `0 issues.` rc=0
- `go test ./internal/cli/... -count=1 -cover` → rc=0, 17개 패키지 전부 `ok`. `internal/cli` coverage **79.9%**

### E4 — 서브에이전트 경계

해당 없음 (에이전트 표면 미변경).

## Baseline-attribution

- develop CI 적색 원판: run `33128899299` (head `d566ecc75`, 2026-08-28T00:11Z). 실패 잡: `Test (ubuntu-latest)` + `Race Test`. `gh run view 33128899299 --log-failed` 전문 판독.
- RED/GREEN 실측은 전부 이 워크트리 `5345b76ed` (origin/develop `d566ecc75` 병합 후)에서 이번 실행으로 측정.

## develop CI 적색 전수 — 5계열 (한 계열만 지목하지 않기)

`--- FAIL:` 전수 계수 (2잡 합산 발생 수):

| # | 계열 | 실패 테스트 | 원인 관측 | 소관 |
|---|---|---|---|---|
| 1 | doctor 판정 | `TestRunDoctor_{WithFix,WithExport,VerboseAndDetail,Verbose,ExportMode,AllFlags}` + `TestDoctorCmd_{Execution,ExportFlag,VerboseExecution}` = **9종** | `doctor: 1 check(s) failed` | **본 카드 t346 — 닫음** |
| 2 | codex init | `TestCodexInitAcceptDelegation`(22) · `TestCodexInitGateInjectedState`(10) · `TestCodexInitGateStateMatrix`(6) | `--spawn needs the moai binary in PATH` | **t349 — 별도 카드, 범위 밖** |
| 3 | graph TempDir 정리 | `TestGitDiffNameCount_Predicate` (양 잡) | `TempDir RemoveAll cleanup: unlinkat /tmp/.../001/.git/objects: directory not empty` — 어서션 실패가 아니라 정리 경합 | **t322** — 도입 커밋 `5d95a2e8d feat(SPEC-GRAPH-FRESHNESS-CADENCE-001): M1 … (t322)` 이 `internal/graph/check_predicate_test.go` 를 만듦. lane-3 이 보유 중이라고 회신 |
| 4 | hook 데이터 레이스 | `TestSessionStart_BlockingComparerDoesNotStallSessionStart` (Race 잡) | `race detected during execution of test` — 스택이 `session_start_binary_lag_test.go:146` (`t.Cleanup` 에서 전역 `binlag.Comparer` 복원) 과 같은 전역을 읽는 비동기 deferred-scan 고루틴을 가리킴 | **t326** — 도입 커밋 `c70c6aed9 feat(t326): make the binary-lag verdict reach an observer unprompted` 이 `internal/hook/session_start_binary_lag_test.go` 를 만듦 |
| 5 | kanban 락 경합 | `TestConcurrencyStress` (Race 잡) | `2/48 adds failed under contention; ... kanban board lock held` — 12 writer × 4 adds 동시 진입에서 락 획득 실패 | **t306** — 도입 커밋 `83a1d492a feat(SPEC-TODO-SQLITE-001): M2 WIP — … concurrency and store tests (t306)`. 카드 t346 본문이 언급했으나 본 카드 소관 아님 (plan §B-동반실패) |


계열 3·4·5 의 귀속은 `git log -- <테스트 파일>` 로 도입 커밋을 직접 읽어 세웠다 (각 파일의 커밋 이력은 1건 = 도입 커밋). 계열 1 의 귀속은 이름·증상이 아니라 **기제로 확인**했다 — 구현을 되돌리면 붉어지고(E8) 되돌리기를 원복하면 초록이 된다.

## Gaps

- **본 수리만으로 develop CI 는 초록이 되지 않는다.** 위 계열 2·3·4·5가 남는다 — 다른 레인의 착지 판정(CI 판독) 봉쇄는 t346 착지로 해소되지 않는다.
- 계열 3·4·5 는 도입 커밋까지만 확인했다 — **각 실패의 근본 원인을 고치는 방법은 조사하지 않았다** (범위 밖).
- `internal/cli` 커버리지 79.9% 는 패키지 임계 85% 미만이나, 본 변경 이전 baseline 을 이번 실행에서 측정하지 않았다 — 개선/악화 방향을 단언하지 않는다. 변경 2파일은 신규 분기 없이 기존 분기의 verdict 만 바꿨고 그 분기는 테스트가 덮는다.
- 전체 스위트(`go test ./...`)는 로컬에서 돌리지 않았다 (CLAUDE.local.md §4 규율) — 전 패키지 판정은 CI 몫.

## Residual-risk

- `make embed-check` 의 바이너리 부재 게이트가 이제 exit 0 이다 (`BIN=/nonexistent/moai make embed-check` → rc=0, 스킵 메시지). 이는 REQ-CDB-001 이 선언한 supersession 의 검증 계층 파생물이며 의도된 것이나, 구 SPEC `SPEC-AGENT-EMIT-LINEAGE-001/acceptance.md` AC-AEL-003 「바이너리 부재 — 게이트」 문면(exit ≠ 0)은 문서 층에서 여전히 옛 거동을 적고 있다. 기계적으로 이를 단언하던 유일한 라이브 테스트는 본 수리가 뒤집은 `TestAgentEmitEmbed_MissingBinaryFails` 하나였고 (`grep -rln 'embed-check\|MissingBinary' internal/ --include='*_test.go'` 의 나머지 2건은 무관한 테스트로 확인), 지금은 남은 기계 단언이 없다. 구 SPEC 문서 갱신은 본 카드 범위 밖 — 후속 카드 후보.
- 스킵이 `ok` 이므로, 개발자가 `make build` 를 잊은 채 doctor 를 돌리면 임베드 드리프트가 판정되지 않은 채 초록으로 보인다. 메시지가 그 사실을 이름으로 말하는 것(REQ-CDB-002)이 유일한 방어선이다.
