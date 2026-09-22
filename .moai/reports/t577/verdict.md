# t577 — develop CI golangci-lint errcheck 2건 수리

Card: t577 · Branch: `WT-lint-errcheck` · Base: develop `3ac58b5a1` (origin/develop, CI red head와 동일)
Date: 2026-09-08 · Lane: lane-9 · Dispatch: lead (run 34184232957)

## Claim

develop CI 적색(golangci-lint errcheck 2건)을 수리한다:
(a) `internal/spec/zz_t528_overacceptance_test.go:100` `defer f.Close()` 미검사 — 상속 건 (직전 head 91d25bc61 run 34158833185에서도 실패)
(b) `internal/cli/todo.go:97` `fmt.Fprintf` 미검사 — t536 배치(db45d8209)가 더한 건
수리 후 golangci-lint 0 issues.

## Evidence (본 트리에서 실행, 원문 보존)

| # | 명령 | 관측 출력 | 비고 |
|---|---|---|---|
| 1 | `golangci-lint run` (수리 전) | `2 issues: * errcheck: 2` — 위 2 좌표 정확히 재현 | 배차 재료와 일치 |
| 2 | `golangci-lint run` (수리 후) | `0 issues.` exit=0 | 원문: `golangci-after.txt` |
| 3 | `go test -count=1 ./internal/spec/` | `ok ... internal/spec 94.875s` | zz_t528 테스트 포함 |
| 4 | `go test -count=1 -timeout 1200s ./internal/cli/` | `ok ... internal/cli 584.350s` | 첫 시도 `-timeout` 미지정 → 601.023s FAIL(기본 10분 타임아웃 도달, internal-cli-timeout-floor 교훈의 알려진 패키지 특성 — `-timeout 1200s` 재실행에서 초록) |
| 5 | `git diff --stat` | 2 files changed, 2 insertions(+), 2 deletions(-) | Class A 판정 재료 — 파일당 1줄 |

## Census — 수리 형태 선택 근거 (리드 지시 실측)

- `_, _ = fmt.Fprint*` 관용구: **non-test 1033건** — 압도적 저장소 관용구.
- `.golangci.yml`: `errcheck.check-blank: false` **명시** ("an errcheck default flip cannot drift local vs CI") — blank identifier 대입이 이 저장소의 공식 "의도적 무시" 표기법.
- `todo.go`의 `warnTempOriginQueueRefusal` 주석: stderr 경고는 informational, "exit code is unchanged" 선언 — stderr 쓰기 실패를 보고할 채널은 없고 동작 변경도 아님.
- 결론 (a): todo.go:97 → `_, _ = fmt.Fprintf(...)` (관용구 채택, 에러 처리 추가 아님).
- `defer func() { _ = f.Close() }()` 관용구: **160건** (updater, closer_test 등 전방위). bare `defer f.Close()`는 errcheck이 실제로 잡음(본 건이 증명).
- 결론 (b): zz_t528:100 → `defer func() { _ = f.Close() }()` (테스트 전용 읽기 핸들, 저장소 표준 형태).

## astgrep 충돌 확인 (리드 지시)

- `go-error-ignored-blank` 룰 패턴 `$_, $ERR = $FUNC($$$ARGS)`는 `_, _ = fmt.Fprintf(...)`와 **매칭됨** — 즉 저장소 공식 관용구가 astgrep warning(severity: warning)을 유발하는 상태. `WarnOnlyMode: true`(차단 없음)라 CI 판정과 무관.
- 룰 note가 약속하는 `// nolint:errcheck` 탈출구는 룰 본문에 미구현 — 리드의 별건 관측 확인. 별건 소관 (t577에서 수정하지 않음).
- `defer func() { _ = f.Close() }()`는 이 룰 패턴과 매칭되지 않음.

## Baseline-attribution

모든 측정은 워크트리 `.claude/worktrees/t577`(HEAD `3ac58b5a1`, origin/develop과 동일, "Already up to date" 흡수 확인)에서 본 세션이 직접 실행해 관측한 출력이다. 배차 재료의 CI run 34184232957 리포트와 로컬 재현(표 1행)이 좌표 수준에서 일치해 baseline이 성립한다.

## Gaps

- **CI Test 잡 판정: 미관측.** 종료 조건 2(develop CI의 Test 잡이 skipped가 아니라 실제 실행돼 초록)는 develop push가 리드 일괄 소관이라 레인이 직접 관측할 수 없다. 로컬 축(spec+cli 패키지 초록, lint 0 issues)만 실측됐다.
- bare `fmt.Fprint*` 호출문이 non-test에 323건 존재하는데 이번 lint가 리포트하지 않는 메커니즘은 미확인(가설: lint 메시지 exclusion 정규식의 대소문자/형태 차이 — 관측 아님). 수리 대상 2건 외 건드리지 않음.
- cgo 파일(`measure_cgo.go`)의 bare `defer Close` 6건이 리포트되지 않는 이유 미확인 — 위와 동일하게 별건.

## Residual-risk

- 나머지 bare 호출 300+ 건이 현재 lint에서 침묵하는 만큼, golangci-lint 버전/설정 변화(예: exclusion 기본값 변동)가 나면 대규모 적색이 한 번에 표면화될 수 있다. `.golangci.yml`의 버전 고정 주석이 이 위험을 인지하고 있으나 exclusion 축은 문서화돼 있지 않다.
- `internal/cli` 소요가 584s로 하한(1200s)의 절반 근처 — 부하에 따라 CI와 로컬 편차가 큰 패키지라, 병합 트리의 CI 판정을 로컬 판정으로 대체하지 말 것.
