# SPEC-SEC-HARDEN-005 — Acceptance Criteria

> AC matrix SSOT. 모든 grep AC는 `$` end-anchor로 정확 테스트명을 지목하여 vacuous-pass(substring collision / nonexistent-test / `[no tests to run]`)를 방지한다. 각 verification command는 명시된 테스트를 실제로 실행한다.

## Given-When-Then 시나리오

### 시나리오 1 — §F.1 `${IFS}` word-split 우회 봉쇄 (재현)
- **Given** `Bash(go test:*)` permission allow-rule이 활성화되어 있고
- **When** `go test ${IFS}curl${IFS}evil` (또는 `$IFS` 변형) 입력이 매처에 전달되면
- **Then** 매처는 비매칭(deny path fall-through)을 반환한다 — 픽스 전에는 ALLOW(true)였다(RED 입증).

### 시나리오 2 — §F.1 정상 입력 ALLOW 보존 (회귀)
- **Given** `Bash(go test:*)` (또는 `Bash(echo:*)`) 룰이 활성화되어 있고
- **When** `go test -race ./...` / `go test $HOME/x` / `go test ${HOME}` / quoted `echo "${IFS}"` 같은 word-split 의도 없는 입력이 전달되면
- **Then** 매처는 매칭(ALLOW)을 유지한다 — 픽스 전 거동과 동일.

### 시나리오 3 — §F.1 malformed shell fail-closed
- **Given** `:*` prefix 룰 remainder가 셸 파서로 파싱 불가한 malformed 입력이고
- **When** 매처가 이를 평가하면
- **Then** 매처는 fail-closed로 비매칭(DENY)을 반환한다 — allow 아님.

### 시나리오 4 — §F.2 update env-trust 거부 (재현)
- **Given** `MOAI_UPDATE_URL`이 non-https 또는 allowlist 외 host로 설정되어 있고
- **When** `EnsureUpdate`가 호출되면
- **Then** 구조화 에러로 fail-closed 거부되고 disallowed source를 가리키는 update checker가 구성되지 않는다 — 픽스 전에는 구성되었다(RED 입증).

### 시나리오 5 — §F.2 default 경로 no-regression
- **Given** update 관련 env가 미설정이고
- **When** `EnsureUpdate`가 호출되면
- **Then** canonical `api.github.com` 기반 update checker가 정상 구성된다(회귀 없음).

---

## AC Matrix

### §F.1 — `${IFS}` shell-aware word-split (REQ-SEC5-001..006)

| AC | 유형 | 요구 | Verification Command (non-vacuous) | 기대 |
|----|------|------|------------------------------------|------|
| **AC-SEC5-001** | 재현 | REQ-SEC5-001/002 | `go test -run 'TestMatches_IFSWordSplit_Reproduction$' -v ./internal/permission/` | `--- PASS` (픽스 후 `${IFS}curl${IFS}evil` → false; 픽스 전 동일 테스트 FAIL) |
| **AC-SEC5-002** | 재현 | REQ-SEC5-002 | `go test -run 'TestMatches_IFSVariants$' -v ./internal/permission/` | `--- PASS` (`${IFS}`/`$IFS`/`${IFS}` 다중삽입 변형 전부 false) |
| **AC-SEC5-003** | 회귀 | REQ-SEC5-005 | `go test -run 'TestMatches_SeparatorVariants$' -v ./internal/permission/` | `--- PASS` (기존 separator DENY 스위트 green 유지) |
| **AC-SEC5-004** | 회귀 | REQ-SEC5-005 | `go test -run 'TestMatches_PrefixChainBypass_Reproduction$' -v ./internal/permission/` | `--- PASS` (SEC-HARDEN-001 M1 chain bypass DENY 유지) |
| **AC-SEC5-005** | 회귀 | REQ-SEC5-006 | `go test -run 'TestMatches_IFSLegitNotRejected$' -v ./internal/permission/` | `--- PASS` (`$HOME`/`${HOME}`/`TestX$`/quoted `${IFS}` 등 9-sample legit set ALLOW 유지). **경계 케이스 `TestX$`(trailing-`$`)는 REQ-SEC5-004(parse-fail→DENY)와 REQ-SEC5-006(legit→ALLOW)의 충돌점 — 반드시 ALLOW여야 한다. 테스트 케이스에 `go test TestX$`를 명시 포함하여 over-deny 회귀를 포착한다(파서가 trailing-`$`를 parse-error로 보면 fail-closed가 DENY로 만들어 REQ-006 위반 → design D.1.4 trailing-`$` special-case ALLOW 결정으로 봉쇄).** |
| **AC-SEC5-006** | fail-closed | REQ-SEC5-004 | `go test -run 'TestMatches_MalformedShellFailClosed$' -v ./internal/permission/` | `--- PASS` (파싱 불가 입력 → false/DENY, allow 아님) |
| **AC-SEC5-007** | 의존성 | REQ-SEC5-003 | `grep -E '^\smvdan\.cc/sh/v3 ' go.mod` | 1 match (직접 require 존재) — 그리고 `grep -c 'blacklist' internal/permission/stack.go` → 0 ($-blacklist 부재) |

### §F.2 — update env-trust allowlist (REQ-SEC5-007..011)

| AC | 유형 | 요구 | Verification Command (non-vacuous) | 기대 |
|----|------|------|------------------------------------|------|
| **AC-SEC5-008** | 재현 | REQ-SEC5-007/008 | `go test -run 'TestEnsureUpdate_RejectsNonHTTPSUpdateURL$' -v ./internal/cli/` | `--- PASS` (non-https `MOAI_UPDATE_URL` → fail-closed 에러; 픽스 전 FAIL) |
| **AC-SEC5-009** | 재현 | REQ-SEC5-008 | `go test -run 'TestEnsureUpdate_RejectsDisallowedHost$' -v ./internal/cli/` | `--- PASS` (allowlist 외 host https → fail-closed) |
| **AC-SEC5-010** | 재현 | REQ-SEC5-009 | `go test -run 'TestEnsureUpdate_RejectsURLShapedReleasesDir$' -v ./internal/cli/` | `--- PASS` (URL-shaped `MOAI_RELEASES_DIR` → fail-closed) |
| **AC-SEC5-011** | 회귀 | REQ-SEC5-010 | `go test -run 'TestEnsureUpdate_DefaultPathNoRegression$' -v ./internal/cli/` | `--- PASS` (env 미설정 → `api.github.com` checker 정상 구성) |
| **AC-SEC5-012** | 범위 | REQ-SEC5-011 + NFR-SEC5-004 | 두 명령을 **각각 독립 실행**(`&&` 단축평가 금지): (1) `grep -nE '"https"' internal/cli/deps.go internal/config/*.go` (2) `grep -c 'EnvUpdateSource\|EnvUpdateURL\|EnvReleasesDir' internal/cli/deps.go` | (1) scheme const `"https"` 존재(≥1 match) **그리고** (2) 3종 env만 참조(≥1, 확장 env 없음) — 둘 다 독립 통과 |

### 전역 검증 (NFR)

| AC | 유형 | 요구 | Verification Command | 기대 |
|----|------|------|----------------------|------|
| **AC-SEC5-013** | NFR | NFR-SEC5-001/002/003 + C-HRA-008 | (4-command batch, 아래) | 전부 PASS |

AC-SEC5-013 batch:
```bash
# build (cross-platform)
go build ./... ; echo "linux=$?"
GOOS=windows GOARCH=amd64 go build ./... ; echo "win=$?"
# full test (no regression)
go test ./internal/permission/... ./internal/cli/...
# subagent boundary (C-HRA-008) — canonical filter (grep -rn 출력은 file:line:CONTENT 이므로
# 주석 제외는 file:line: prefix를 건너뛴 뒤 // 를 매칭해야 함; scope는 본 SPEC 수정 파일 2종으로 한정)
grep -rn 'AskUserQuestion\|mcp__askuser' internal/permission/stack.go internal/cli/deps.go | grep -v '_test.go' | grep -v "^[^:]*:[0-9]*:[ \t]*//"
# lint (NEW vs baseline)
golangci-lint run --timeout=2m ./internal/permission/... ./internal/cli/...
```
기대: linux=0, win=0, full test ok(all packages), C-HRA-008 grep 0 매치, lint NEW 0.

### OPTIONAL (비요구 — 게이트 아님)

| 항목 | 유형 | 요구 | 비고 |
|------|------|------|------|
| **OPT-SEC5-001** | OPTIONAL | OPT-SEC5-001 | `restoreTargetContained`/`parentChainContained`/`runMXScan` godoc에 TOCTOU note. 코드 동작 변경 없음. **AC가 아님 — 구현 게이트하지 않음.** 추가 시 검증: `grep -c 'TOCTOU\|check-vs-use' internal/cli/update.go internal/hook/file_changed.go` ≥ 1. |

---

## AC 집계

- **총 AC: 13** (게이트하는 acceptance criteria)
  - **재현(reproduction) AC: 5** — AC-SEC5-001, -002 (§F.1 ${IFS}), AC-SEC5-008, -009, -010 (§F.2 env-trust). 각각 픽스 전 RED → 픽스 후 GREEN 입증.
  - **회귀(regression) AC: 4** — AC-SEC5-003, -004, -005 (§F.1 separator/chain/legit 보존), AC-SEC5-011 (§F.2 default no-regression).
  - **fail-closed AC: 1** — AC-SEC5-006 (§F.1 malformed → DENY).
  - **의존성/범위 AC: 2** — AC-SEC5-007 (mvdan.cc/sh 직접 require + no-$-blacklist), AC-SEC5-012 (scheme const + env 3종 한정).
  - **전역 NFR AC: 1** — AC-SEC5-013 (build/test/C-HRA-008/lint batch).
- **OPTIONAL(비게이트): 1** — OPT-SEC5-001 (TOCTOU godoc, 코드 동작 변경 없음).

## Definition of Done

- [ ] 13 AC 전부 PASS (재현 5 + 회귀 4 + fail-closed 1 + 의존성/범위 2 + 전역 NFR 1)
- [ ] `mvdan.cc/sh/v3` 직접 go.mod require + `go mod tidy` clean
- [ ] `$`-blacklist 부재(REQ-SEC5-003)
- [ ] env-trust 검증 update source 3종 env 한정(REQ-SEC5-011)
- [ ] cross-platform build (linux + windows) exit 0
- [ ] C-HRA-008 grep 0 매치
- [ ] coverage no-regression vs baseline
- [ ] lint NEW issues 0
- [ ] PRESERVE 거동 불변(`hasUnquotedShellSeparator` separator DENY, TOCTOU 코드 동작)
- [ ] Conventional Commits + `🗿 MoAI` trailer, main 직진 push
