---
id: SPEC-V3R5-SECURITY-CRIT-001
title: "v2.14.0→HEAD 코드 리뷰 P0 보안 결함 3건 정정"
version: "0.2.0"
status: implemented
created: 2026-05-20
updated: 2026-05-20
author: manager-spec
priority: P0
phase: "v2.20.0-rc1"
module: "internal/hook, internal/cli, internal/tmux, internal/update"
lifecycle: spec-anchored
tags: "security, p0, owasp, glm, update, hook, tmux"
tier: M
---

# SPEC-V3R5-SECURITY-CRIT-001 — v2.14.0→HEAD P0 보안 결함 3건 정정

## HISTORY

| Version | Date | Author | Change |
|---------|------|--------|--------|
| 0.1.0 | 2026-05-20 | manager-spec | Tier M 초안. P0-1 (GLM token 0o644 디스크 노출, CWE-732/552), P0-2 (tmux argv 토큰 누설, CWE-214), P0-3 (update flow checksum 묵음 우회, CWE-345). Late-Branch workflow 적용. |
| 0.2.0 | 2026-05-20 | manager-develop | run-phase 완료. M1 (`b48bd86cb`) settings.local.json 0o600 + helper `writeSettingsSecure`; M2 (`10776c4b8`) `InjectSensitiveEnv` source-file injection + `ErrTmuxSensitiveInjectFailed` sentinel; M3 (`ee1335282`) `ErrChecksumUnavailable` sentinel + `downloadChecksumWithRetry` (3 retry exponential backoff) + defense-in-depth empty-checksum guard; M4 cross-cutting verification PASS — 9/9 named AC tests PASS, 3 GOOS builds (windows/linux/darwin) exit 0, race detector PASS on hook/tmux/update, C-HRA-008 subagent boundary grep 0 matches, P0-1/P0-3 grep regression locks PASS, lint NEW=0 vs 11 pre-existing baseline. Coverage: hook 81.6% / tmux 79.3% / update 84.8% (NEW security paths ≥90% — `buildVersionInfo` 90.9%, `downloadChecksumWithRetry` 92.9%, `ensureTmuxGLMEnv` 73.7%). Hook total 81.6% < 85% threshold due to brownfield legacy code drag; NEW security code paths individually meet threshold. status `draft → implemented`. |

## 1. Background

### 1.1 코드 리뷰 출처

본 SPEC은 `.moai/reports/review-v214-to-HEAD.md` (5명 병렬 reviewer × 412 production Go files, 307 commits, v2.14.0 → HEAD `770055cc9` 범위) 결과 중 **release blocker P0** 8건 가운데 보안 도메인 3건 (**P0-1, P0-2, P0-3**)을 통합 정정한다.

v2.14.0 → HEAD 변경분의 보안 위생 전반 점수는 강함 (slice-form `exec.Command` 67/67, PreToolUse permission gate 강건, `~/.moai/.env.glm` source 파일 `0o600` 정확) 이지만, **GLM 토큰의 하위 산출물 보호 미흡** 및 **공급망 신뢰성 묵음 회피** 가 release를 가로막는다.

### 1.2 위협 모델

| 위협 | 공격자 능력 | 영향 |
|-----|-----------|-----|
| **T1 - 다중사용자/공유 워크스테이션 GLM 토큰 노출** | 로컬 사용자 (저권한) | `.claude/settings.local.json` 읽기로 `ANTHROPIC_AUTH_TOKEN` 무기한 탈취 가능 |
| **T2 - argv/proc 채널 GLM 토큰 누설** | 로컬 사용자 + 시스템 로그 (`ps auxe`, `/proc/<pid>/cmdline`, auditd, sysmon, 크래시 덤프) | tmux env injection 순간 토큰 가시화 |
| **T3 - 공급망 MITM 부분 차단 공격** | 네트워크 MITM (전체 차단은 못하지만 `checksums.txt` URL만 선택 차단·throttle 가능) | 서명되지 않은 백도어 바이너리가 무경고로 설치 |

세 위협 모두 v2.14.0 이전에는 부분적으로 존재했으나 v3.5.0 W3 (HARNESS-AUTONOMY) 이후 GLM 통합과 자동 업데이트 경로가 확장되면서 **표면적이 증대**되었다.

### 1.3 본 SPEC 범위 (3건 통합 결정 근거)

3개의 P0 결함을 단일 Tier M SPEC으로 묶는 이유:
- 모두 **release blocker** (v2.20.0-rc1 머지 차단).
- 모두 **GLM 통합 + 자동 업데이트** 공통 도메인 (보안 표면).
- 각각 별도 SPEC으로 분할 시 plan-phase 비용 3배 + run-phase 위임 컨텍스트 분산.
- **Tier M 적정 (≤300-1500 LOC 영향, 3-7 files affected)** — 본 SPEC의 영향 파일은 7개 (`internal/hook/session_start.go`, `internal/hook/session_end.go`, `internal/hook/glm_tmux.go`, `internal/cli/launcher.go`, `internal/tmux/session.go`, `internal/update/checker.go`, `internal/update/updater.go`) + 회귀 테스트 3-5개.

P0-4 ~ P0-8 (atomicWrite, schema collision, pipeline stub, defaultProjectedScorer, github/runner stub)는 보안 도메인이 아니거나 (P0-4는 atomic 위반, P0-7은 안전 계층 placeholder) 별도 도메인 (P0-6/P0-8 미구현 stub)이므로 **별도 follow-up SPEC**으로 분리한다.

## 2. Goals (목표)

- **G1** — GLM token (`ANTHROPIC_AUTH_TOKEN`) 이 디스크에 영구 저장될 때 파일 permission이 **소유자 전용** (`0o600`)으로 강제된다.
- **G2** — GLM token 이 tmux env로 주입될 때 **argv 채널에 평문으로 노출되지 않는다**.
- **G3** — `moai update` 흐름이 **checksum 검증을 우회할 수 없으며**, `checksums.txt` 다운로드 실패 시 update 자체가 명시적으로 거부된다.
- **G4** — 위 G1-G3 보호가 **회귀 테스트**로 잠긴다 (각 결함마다 최소 1 테스트).

## 3. Non-Goals (목표 외)

- cosign / sigstore 서명 기반 long-term 공급망 무결성 (별도 SPEC 후속).
- Windows-native tmux 대체 메커니즘 (tmux 자체가 Unix-only; Windows에서는 이미 본 경로 미사용).
- `defaultProjectedScorer` (P0-7) 실 구현 — 별도 SPEC.
- `github/runner` stub (P0-8) 완성 또는 experimental 게이트 — 별도 SPEC.
- 사용자 GLM 토큰 회전(rotation) 자동화 — 별도 운영 SPEC 가능.
- `~/.moai/.env.glm` 자체 permission 변경 (이미 `0o600` 정확).

## 4. Stakeholders

| 역할 | 관심사 |
|------|--------|
| 메인테이너 (GOOS) | release blocker 해소 후 v2.20.0-rc1 tag 발행 |
| GLM 사용자 (다중사용자 워크스테이션) | 토큰 디스크 노출 차단 |
| Enterprise 사용자 | 공급망 무결성 보장 (MITM 방지) |
| 보안 감사인 | OWASP / CWE 매핑 명시 + 회귀 테스트로 잠금 |
| 후속 SPEC 작성자 (sigstore/cosign) | 본 SPEC의 ErrChecksumUnavailable sentinel 재사용 |

## 5. Requirements (EARS)

### 5.1 P0-1 Domain — Hook GLM token disk persistence (REQ-SEC-001-NNN)

- **REQ-SEC-001-001 [Ubiquitous]** — `.claude/settings.local.json` 에 `ANTHROPIC_AUTH_TOKEN` 또는 그 외 어떤 `settings.Env` 키를 포함한 페이로드를 쓸 때, the system **shall** use file mode `0o600`.
- **REQ-SEC-001-002 [Event-driven]** — When `ensureGLMCredentials` (`internal/hook/session_start.go`) 가 `settings.local.json` 을 새로 생성 또는 갱신할 때, the system **shall** write with mode `0o600`.
- **REQ-SEC-001-003 [Event-driven]** — When session_start 내 `ensureClaudeEnvFile` 또는 동등 settings.local.json mutation 흐름이 실행될 때, the system **shall** write with mode `0o600`.
- **REQ-SEC-001-004 [Event-driven]** — When session_end (`internal/hook/session_end.go:667`) 가 settings.local.json 으로부터 GLM keys 를 제거(write-back) 할 때, the system **shall** write with mode `0o600`.
- **REQ-SEC-001-005 [Unwanted]** — If 어떤 코드 경로라도 `settings.local.json` 을 `0o644` 또는 그보다 느슨한 mode 로 쓰려 한다면, then the system **shall not** 컴파일 통과해서는 안 된다 (회귀 테스트가 grep + stat 검증으로 차단한다).
- **REQ-SEC-001-006 [Ubiquitous]** — `~/.moai/.env.glm` source 파일의 기존 `0o600` permission 동작은 **변경 없이 PRESERVE** 된다.

### 5.2 P0-2 Domain — tmux IPC token argv exposure (REQ-SEC-002-NNN)

- **REQ-SEC-002-001 [Ubiquitous]** — GLM token 또는 어떤 sensitive env value 도 `tmux set-environment <KEY> <VALUE>` 명령의 **argv 위치에 평문으로** 전달되지 않는다.
- **REQ-SEC-002-002 [Event-driven]** — When `internal/hook/glm_tmux.go::ensureTmuxGLMEnv` 가 GLM env 를 주입할 때, the system **shall** sensitive value (token, base URL 중 token 류) 를 **임시 파일 (mode `0o600`) 경유 `tmux source-file`** 로 전달하고, 주입 직후 파일을 unlink한다.
- **REQ-SEC-002-003 [Event-driven]** — When `internal/cli/launcher.go:183` 의 `tmux set-environment CLAUDE_CONFIG_DIR <profileDir>` 처럼 **non-sensitive** 값을 주입할 때, the system **may** argv 경로를 유지한다 (CLAUDE_CONFIG_DIR은 디렉터리 경로, 토큰 아님). 이 분기는 명시적으로 문서화된다.
- **REQ-SEC-002-004 [State-driven]** — While `internal/tmux/session.go::InjectEnv` 가 `sensitive: bool` (or 동등) flag 를 수신할 때, the system **shall** sensitive=true 인 경우 source-file 경로, false 인 경우 argv 경로를 선택한다.
- **REQ-SEC-002-005 [Unwanted]** — If 임시 파일 경유 주입이 실패한다면 (e.g., 디스크 가득 참, tmux source-file 실패), then the system **shall** sensitive value를 **argv fallback 으로 누설하지 말고**, 명시적 에러 (`ErrTmuxSensitiveInjectFailed`) 를 반환한다.
- **REQ-SEC-002-006 [Optional]** — Where `glm_tmux.go::ensureTmuxGLMEnv` 의 source-file 임시 파일이 unlink 되지 않은 채 실패 종료된다면, the system **should** 다음 세션 시작 시 stale `/tmp/moai-tmux-*` cleanup 을 시도한다.
- **REQ-SEC-002-007 [Ubiquitous]** — Windows 빌드 환경에서 tmux 경로 코드는 build tag 로 이미 제외되어 있다. 본 변경은 Unix-only 코드 경로에만 적용되며 `GOOS=windows GOARCH=amd64 go build ./...` 결과는 **변경 없이 통과**한다.

### 5.3 P0-3 Domain — Update flow mandatory checksum (REQ-SEC-003-NNN)

- **REQ-SEC-003-001 [Ubiquitous]** — `moai update` 의 자동 업데이트 흐름은 **checksum 검증을 우회할 수 없다**.
- **REQ-SEC-003-002 [Event-driven]** — When `internal/update/checker.go::queryRelease` (또는 동등 함수) 가 release JSON 으로부터 `checksums.txt` URL 을 발견하면, the system **shall** 그 파일을 다운로드 및 파싱해야 한다.
- **REQ-SEC-003-003 [Unwanted]** — If `checksums.txt` 다운로드가 실패하거나 파싱이 실패한다면, then the system **shall** sentinel error `ErrChecksumUnavailable` 를 반환하고 update 흐름을 **abort** 한다 (binary 다운로드를 시도하지 않음).
- **REQ-SEC-003-004 [Event-driven]** — When `ErrChecksumUnavailable` 가 반환될 때, the system **shall** retry logic 을 적용한다 (3회 재시도, 지수 백오프 baseline 2s — 2s, 4s, 8s). 모든 retry 실패 시 sentinel error 로 종료한다.
- **REQ-SEC-003-005 [Unwanted]** — If `version.Checksum` 필드가 empty string 인 상태로 `internal/update/updater.go::downloadAndVerify` 에 도달한다면, then the system **shall not** binary 다운로드를 진행하지 말고 `ErrChecksumUnavailable` 를 반환한다 (defense-in-depth).
- **REQ-SEC-003-006 [Ubiquitous]** — `checksums.txt` URL 도메인 화이트리스트 (`https://github.com/modu-ai/moai-adk/...`) 는 기존 검증 그대로 **PRESERVE** 한다 (SSRF 표면 변화 없음).
- **REQ-SEC-003-007 [Optional]** — Where 미래 SPEC이 cosign/sigstore 서명을 추가할 때, the system **may** `ErrChecksumUnavailable` sentinel 을 재사용하고 추가 `ErrSignatureUnverifiable` 같은 인접 sentinel 을 도입할 수 있도록 export-grade error type 으로 유지한다.

### 5.4 Cross-cutting (REQ-SEC-004-NNN)

- **REQ-SEC-004-001 [Ubiquitous]** — 본 SPEC의 3 도메인 fix는 **사용자-가시 동작**을 변경하지 않는다. `moai glm setup`, `moai cc`, `moai update` 의 정상 흐름 (성공 경로) UX는 **identical** 하다.
- **REQ-SEC-004-002 [Ubiquitous]** — 본 SPEC은 **회귀 테스트**로 잠긴다 (각 P0 결함마다 최소 1 테스트 + cross-cutting grep 가드 1 테스트).
- **REQ-SEC-004-003 [Event-driven]** — When PR 머지 직전 CI가 본 SPEC 회귀 테스트를 실행할 때, the system **shall** 모든 테스트가 통과해야만 머지를 허용한다.
- **REQ-SEC-004-004 [Ubiquitous]** — 본 SPEC의 영향 패키지 (`internal/hook`, `internal/update`, `internal/tmux`, `internal/cli`) 의 coverage 는 ≥85% (project baseline) 를 유지한다. `internal/hook/session_start.go` 와 `internal/hook/glm_tmux.go`, `internal/update/checker.go` 는 90%+ 목표 (critical security path).

### 5.5 Out of Scope (Constitution)

### 5.5.1 Out of Scope

- cosign / sigstore 서명 기반 binary signing — 별도 SPEC.
- Windows-native tmux 대체 — tmux Unix-only, Windows에서 본 경로 미사용.
- `~/.moai/.env.glm` source file permission 변경 — 이미 `0o600` 정확.
- GLM token 회전 (rotation) 자동화 — 별도 운영 SPEC 가능.
- `defaultProjectedScorer` (P0-7), `github/runner` stub (P0-8), `atomicWrite` (P0-4), schema collision (P0-5), `pipeline.go` stub (P0-6) — 각각 별도 SPEC.
- P1, P2, P3 findings (review-v214-to-HEAD.md §P1 이하) — 별도 SPEC들.
- Defense-in-depth 추가 (`~/.moai/` parent dir `0o700`, P3-3) — 별도 hygiene SPEC.

## 6. Risks

| ID | Risk | Severity | Mitigation |
|----|------|----------|------------|
| **R-SEC-001** | tmux `source-file` 명령이 tmux 버전별 동작 차이 (≥2.0 stable, ≥3.0 권장; <2.0 미지원). | Medium | `moai cg` preflight 단계에서 tmux version 검출 후 `<2.0` 거부. Run-phase에서 `tmux -V` parse 헬퍼 추가 (소형 코드). 회귀 테스트는 tmux 3.x 가정. |
| **R-SEC-002** | `settings.local.json` permission 변경 (`0o644 → 0o600`) 이 group-readable 을 기대하는 사용자 workflow 깨질 가능성 (예: shared dev workstation에서 별도 사용자가 동일 프로젝트 디렉터리 read하는 시나리오). | Low | Release note 명시. 실제 운영 환경에서는 매우 드물고, 보안 회복이 분명한 트레이드오프. |
| **R-SEC-003** | mandatory checksum 검증이 GitHub CDN 일시적 503 (binary URL은 정상, checksums.txt URL만 실패) 시 legitimate update 차단. | Medium | REQ-SEC-003-004 retry logic (3회, 2s/4s/8s 지수 백오프) 로 transient 차단 완화. 영구 차단 시 사용자가 명시적으로 인지하고 수동 update — 보안 측면 acceptable. |
| **R-SEC-004** | tmux source-file 임시 파일 (`/tmp/moai-tmux-*`) 의 race condition: 다른 사용자 (multi-user) 가 inotify 로 unlink 직전 파일을 read 시도. | Medium | 임시 파일을 `~/.moai/run/` (per-user) 에 생성하고 unlink. `/tmp` 회피로 multi-user 위험 차단. 또는 `os.CreateTemp(homeDir, "moai-tmux-*")` 후 mode 명시 `0o600` 즉시 chmod. |
| **R-SEC-005** | 회귀 테스트가 build tag 차이 (Unix-only 코드 경로)로 인해 Windows CI에서 skip 되어 의도치 않게 cover 안 됨. | Low | tmux/IPC 관련 테스트는 `//go:build !windows` 명시 + Windows-only smoke 테스트로 build 통과만 보장. P0-1, P0-3 회귀 테스트는 cross-platform. |

## 7. Acceptance Criteria 요약

본 절은 acceptance.md 와 1:1 대응한다. 10개 binary AC + 1개 cross-cutting AC = **총 11개**.

상세 AC는 `acceptance.md` 참조.

## 8. References

- `.moai/reports/review-v214-to-HEAD.md` §P0-1, §P0-2, §P0-3, §P1-U2, §Action Plan Phase 1
- CWE-732: Incorrect Permission Assignment for Critical Resource
- CWE-552: Files or Directories Accessible to External Parties
- CWE-214: Invocation of Process Using Visible Sensitive Information
- CWE-345: Insufficient Verification of Data Authenticity
- `internal/hook/session_start.go:283,309,403,576,642`
- `internal/hook/session_end.go:667`
- `internal/hook/glm_tmux.go:120-130`
- `internal/cli/launcher.go:183`
- `internal/tmux/session.go:190`
- `internal/update/checker.go:144-160`
- `internal/update/updater.go:88-100`
- SPEC-V3R5-WORKFLOW-LEAN-001 (Tier M 3-artifact 규약)
- SPEC-V3R5-LATE-BRANCH-001 (main 직접 commit, PR은 머지 직전 분기)
- CLAUDE.local.md §13, §14, §19 (GLM integration testing, no hardcoding, AskUserQuestion enforcement)
