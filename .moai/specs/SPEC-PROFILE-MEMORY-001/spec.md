---
id: SPEC-PROFILE-MEMORY-001
title: "프로필 기억 3대 결함 — 프로젝트별 기억·디렉터리 검증·최초 전환 고지"
version: "0.3.0"
status: completed
created: 2026-08-02
updated: 2026-08-02
author: GOOS
priority: P1
phase: "v3.1.0"
module: "internal/profile"
lifecycle: spec-anchored
amendment_of: SPEC-PROFILE-MEMORY-001
tags: "profile, launcher, ledger, cli, claude-config-dir"
tier: L
---

# SPEC-PROFILE-MEMORY-001 — 프로필 기억 3대 결함

## HISTORY

| 날짜 | 버전 | 변경 |
|------|------|------|
| 2026-08-02 | 0.1.0 | 최초 작성 (plan-phase 산출물, Tier M) |
| 2026-08-02 | 0.2.0 | plan-audit iter-1 FAIL(0.68) 대응 — Tier M→L 승격(REQ 23 > Tier M 상한 16), D1-D13 12건 반영, REQ-PM-024 추가(웹 읽기 측 project-scoped), 알려진 한계 L-003 추가, design.md·research.md 신규. 재감사 임계 **0.85** |
| 2026-08-02 | 0.2.2 | plan-audit iter-3 PASS(0.88) 후 잔여 debt 5건 선제 해소 — NEW-1(stderr `io.Writer` 시임 신설) / NEW-2(design §D.3 게이트 변수 정정) / NEW-3(넓힌 게이트 안전성 근거) / NEW-4(AC-PM-018 케이스 C 양 경로 분리) / NEW-5(`named` 술어 정의). REQ 24 / AC 21 유지 |
| 2026-08-02 | 0.2.1 | plan-audit iter-2 FAIL(0.82) 델타 대응 — N1(고지 게이트를 해석 후 `profileName` 으로, bare 런치 공백 제거) / N2(AC-PM-020 실패 유도 레시피 고정) / N3(REQ-PM-005 매트릭스 복원) / N4(RED-first 명시) / N5(테스트 함수명 명시) / N6(REQ-PM-024 env 우선순위 조건) / N7(프로필 4개 실측 정정). REQ 24 유지 |

## Amendments

- **prior completed version:** `0.2.2`
- **prior_completed_sha:** `7a4341750` (origin/main PR #1289 squash-merge; 로컬 `progress.md §E.4 sync_commit_sha` 인 `53756d4f1` 은 squash 이전 로컬 히스토리에만 존재하는 사전-squash 동등 커밋이다)
- **rationale:** falsification 왕복 부채 하위 기준 두 건(`AC-PM-014(3)` 및 `AC-PM-021(3)`) 폐쇄 — 원 종료 시점에 실행되지 않은 두 PASS-WITH-DEBT sandbox-guard AC 의 (3) 하위 기준(guard disabled → FAIL 관측 → restored → PASS 관측)을 실행한다.
- **scope:** 검증 전용 amendment; production-code 변경 없음 (guard 는 동일 run 안에서 임시 비활성화 후 복원된다; 최종 트리는 guard 파일에 대해 `7a4341750` 와 byte-identical하다).

---

## §A 배경

`moai cc | glm | cg` 의 `-p/--profile` 플래그는 `~/.moai/claude-profiles/<name>` 을 `CLAUDE_CONFIG_DIR` 로 지정해 Claude Code 설정을 격리한다. 마지막으로 쓴 프로필은 `~/.moai/claude-profiles/launch.yaml` 의 `last_profile` 키 하나에 기록된다.

HEAD `c907db541` 코드 판독으로 확인된 결함 3건:

1. **프로젝트별 기억 부재** — `GetBaseDir()` 는 홈 기준 단일 경로만 돌려주고, 원장에는 전역 `last_profile` 키 하나뿐이다. 프로젝트 B에서 프로필을 띄우면 프로젝트 A의 기억이 덮인다.
2. **존재하지 않는 디렉터리 이름을 원장이 수용** — `RecordLastUsedProfile` 은 이름의 *형태*(`isValidProfileName`)만 검사하고 디렉터리 존재는 확인하지 않는다. `ResolveLaunchProfile` 은 읽기 측에서만 `os.Stat` 로 걸러내므로, 유령 이름이 사용자에게 아무 신호 없이 원장에 영구 잔류한다.
3. **새 프로필 전환 시 로그인이 조용히 풀림** — 새 프로필 디렉터리에는 `.claude.json`(oauthAccount / hasCompletedOnboarding / userID 보유)이 없다. `EnsureDir` 는 `os.MkdirAll` + `os.Setenv` 만 하므로 아무것도 시드하지 않고 아무 경고도 내지 않는다. 사용자는 예고 없이 로그인/온보딩 화면을 만난다.

### 결함 3의 기전 정정

이전 기록은 원인을 "`.credentials.json` 이 `~/.claude` 에만 있어서"로 적었으나 macOS에서는 불완전하다. 실측:

- macOS 인증 토큰은 **Keychain** 의 `Claude Code-credentials` 서비스에 있다. 서비스 스코프 전역이라 `CLAUDE_CONFIG_DIR` 변경으로 사라지지 않는다.
- 설정 디렉터리별 상태는 **`.claude.json`** 이다. `CLAUDE_CONFIG_DIR` 미설정 시 `~/.claude.json`, 설정 시 `<profileDir>/.claude.json`.
- `internal/statusline/usage.go` 는 읽기 우선순위를 macOS Keychain → `~/.claude/.credentials.json` → `~/.claude/credentials.json` 로 문서화한다. Linux/WSL2에서는 파일 경로가 실제 운반체이므로 `.credentials.json` 부재가 직접 문제가 된다.

즉 정확한 기전은 **새 프로필 디렉터리에 `.claude.json` 이 없어 계정 상태가 승계되지 않는 것**이며, 플랫폼별 인증 운반체가 다르다는 사실이 자동 시드를 범위 밖으로 두는 근거다.

---

## §B 요구사항 (GEARS)

### B.1 프로젝트별 프로필 기억 (결함 1)

- **REQ-PM-001**: The launch ledger shall carry a `projects:` mapping whose keys are absolute project paths and whose values are named profile names, alongside the retained top-level `last_profile` key.
- **REQ-PM-002**: **When** a launch records a named profile and a project root is known, the profile subsystem shall write both the project-scoped entry under `projects:` and the top-level `last_profile` key.
- **REQ-PM-003**: **When** a bare launch resolves a profile, the launcher shall consult the project-scoped entry for the current project first, then the top-level `last_profile`, then yield `""` (default semantics).
- **REQ-PM-004**: **Where** the ledger carries no `projects:` key (a pre-upgrade ledger), the profile subsystem shall resolve from `last_profile` alone, producing behavior identical to the pre-upgrade resolver.
- **REQ-PM-005**: The ledger writer shall remain a read-modify-write that preserves unknown and legacy keys (`bypass:`, `model:`) and shall remain atomic (temp file plus `os.Rename`).
- **REQ-PM-006**: **Where** `MOAI_NO_PROFILE_FALLBACK` is set to `1`, the profile subsystem shall disable both the project-scoped lookup and the global `last_profile` fallback.
- **REQ-PM-007**: **When** the caller supplies no project root, the profile subsystem shall fall through to the top-level `last_profile` and shall not return an error.
- **REQ-PM-008**: **When** a project-scoped entry names a profile whose directory is absent, the resolver shall skip that entry and continue to the global fallback.
- **REQ-PM-009**: The profile subsystem shall normalize project-path keys by symlink resolution with a lexical clean fallback, so a single project root maps to a single ledger key.
- **REQ-PM-010**: The web console write seam `recordLastProfile func(name string) error` shall keep its signature; the project root shall be supplied by the closure wired in `newApp`.
- **REQ-PM-024**: **Where** `CLAUDE_CONFIG_DIR` is unset, the `moai web` command shall resolve its displayed profile name project-scoped, so the console's read side and its write side agree on the same project.

> **REQ-PM-024 발화 조건 주석 (감사 iter-2 N6)**: `GetCurrentName()` 및 그 project-aware 변종은 `CLAUDE_CONFIG_DIR` 을 **먼저** 읽고, 설정돼 있으면 원장을 아예 조회하지 않는다. 이 env 우선순위는 이 SPEC 이전부터 존재하는 올바른 동작이며 이 SPEC은 바꾸지 않는다. 결과적으로 `moai cc -p X` 로 띄운 Claude Code 세션 **안에서** `moai web` 을 실행하면 항상 env 경로를 타므로 프로젝트 스코프 조회가 발화하지 않는다. REQ-PM-024는 `CLAUDE_CONFIG_DIR` 미설정 경로(별도 터미널에서 `moai web` 직접 실행)에서 발화한다. 판정: AC-PM-019 (Given 이 미설정을 고정한다).

> **REQ-PM-007 도달성 주석 (감사 D13)**: `unifiedLaunchDefault` 은 `findProjectRootFn()` 실패 시 즉시 에러를 반환하며 이 SPEC은 그 동작을 바꾸지 않는다. 따라서 런치 경로에서는 "project root 를 얻지 못한 채 해석이 진행되는" 상태가 발생하지 않는다. REQ-PM-007 이 규정하는 대상은 **무변경 래퍼 계약**(`projectRoot == ""` 로 호출된 경우)이며, 그 호출자는 `moai profile current` / `moai update` / `moai init` 이다. 판정은 함수 직접 호출(AC-PM-005)로 한다. 관련 결정: `plan.md` §C RESOLVED-3.

### B.2 쓰기 측 디렉터리 검증 (결함 2)

- **REQ-PM-011**: **When** the ledger writer is asked to record a profile name whose directory does not exist under the profile base, the writer shall refuse with an error and shall not mutate the ledger.
- **REQ-PM-012**: The launcher shall record the last-used profile only after the profile directory has been created and before the process is replaced by the Claude Code binary.
- **REQ-PM-013**: **When** a user launches a not-yet-existing profile with `-p <new-profile>` for the first time, the launcher shall still record that profile in the ledger.
- **REQ-PM-014**: **When** the ledger write fails, the launcher shall emit a warning on stderr and shall continue the launch.
- **REQ-PM-015**: The launcher shall not record a profile name that differs from the profile directory it created for the same launch.

### B.3 새 프로필 전환 고지 (결함 3)

- **REQ-PM-016**: **When** a launch targets a named profile whose directory carries no Claude configuration state, the launcher shall emit a stderr notice stating that a login or onboarding screen will appear, exactly once per launch.
- **REQ-PM-017**: **While** the target profile directory already carries Claude configuration state, the launcher shall emit no such notice.
- **REQ-PM-018**: The profile subsystem shall decide the absent-configuration condition solely by the presence of the Claude configuration state file in the profile directory, and shall not consult any platform-specific credential store.
- **REQ-PM-019**: The launcher shall not copy, move, or synthesize any credential or account file between profile directories.
- **REQ-PM-020**: The notice text shall not name a command that this SPEC does not deliver.

### B.4 횡단 제약

- **REQ-PM-021**: Every test added by this SPEC shall sandbox the profile base directory through `profile.BaseDirOverride` and `t.TempDir()`, and shall neither read nor write the real home directory.
- **REQ-PM-022**: The `internal/profile` and `internal/web` packages shall each carry a package-level `TestMain` that sandboxes `profile.BaseDirOverride` before `m.Run()`, mirroring the `internal/cli` guard.
- **REQ-PM-023**: New environment variable names shall live in `internal/config/envkeys.go`; new ledger YAML keys and file names shall be named constants beside `lastProfileKey`.

---

## §C 범위 제외 (out of scope)

이 SPEC이 만들지 **않는** 것.

### Out of Scope — 인증 자동 시드

- 프로필 디렉터리 간 `.claude.json` / `.credentials.json` 복사·이동·합성을 하지 않는다.
- macOS Keychain 항목을 읽거나 쓰지 않는다.
- `--seed-auth` 류의 신규 플래그·서브커맨드를 만들지 않는다. 결함 3의 처리는 **경고만**이다.

### Out of Scope — 원장 정리(GC)

- 삭제·이동된 프로젝트의 `projects:` 항목을 자동 제거하지 않는다. 이동/이름변경은 항목을 고아로 남기며, 이는 아래 L-002에 기록된 수용된 한계다.
- `moai profile` 에 원장 편집·정리용 신규 서브커맨드(`prune` 등)를 추가하지 않는다.
- 고아 항목에 대한 런치 시점 고지를 하지 않는다.

### Out of Scope — 원장 동시 쓰기 잠금

- 원장에 파일 잠금(flock 등)을 도입하지 않는다. 크로스 플랫폼 잠금은 제약 C5와 얽히며 이 SPEC의 3대 결함과 독립된 문제다. 결과는 아래 L-003에 수용된 한계로 기록한다.

### Out of Scope — 코드 범위

- `internal/template/templates/` 를 수정하지 않는다. 템플릿 중립성 규칙은 이 SPEC에 적용되지 않는다.
- `internal/profile/preferences.go`, `sync.go` 의 선호값 읽기·쓰기 경로를 바꾸지 않는다.
- 프로필 목록·삭제(`List`, `Delete`) 의 표시 의미를 바꾸지 않는다.
- 지나가는 리팩터링을 하지 않는다. 접촉 대상은 `internal/profile/`, `internal/cli/launcher.go`, `internal/cli/web.go` 1줄, `internal/web/app.go` 1지점, 그리고 신규·수정 테스트 파일뿐이다.

### 알려진 한계 — 사용자 수용 (2026-08-02)

아래 세 결과는 결함이 아니라 **명시적으로 수용된 설계 귀결**이다. 근거는 `plan.md` §C RESOLVED-1 / RESOLVED-2 / RESOLVED-4.

- **L-001 — `moai profile current` 표시와 실제 런치 프로필의 불일치.** 프로젝트 스코프 기억은 `moai cc | glm | cg` 런치 경로와 `moai web`(REQ-PM-024) 에만 적용된다. `moai profile current`, `moai update`, `moai init` 은 기존 동작(전역 `last_profile` 폴백)을 유지하므로, 프로젝트 스코프 항목이 전역값과 다른 프로젝트에서는 `moai profile current` 가 **바로 그 프로젝트에서 bare `moai cc` 가 실제로 띄울 프로필과 다른 이름을 표시할 수 있다**. 기존 사용자가 보던 출력을 바꾸지 않는 쪽을 택한 결과다. `moai web` 은 이 한계에 포함되지 않는다.
- **L-002 — `projects:` 맵의 무제한 증가.** 정리 메커니즘이 없으므로 원장의 `projects:` 항목은 시간이 지나며 단조 증가하고, 이동·삭제된 프로젝트의 고아 항목이 영구 잔류한다. 해석 시 조용히 건너뛰므로 동작상 무해하나 파일은 계속 커진다. 심링크 해소 실패로 한 프로젝트가 두 키를 차지하는 경우(`plan.md` §D1)도 동일 분류다.
- **L-003 — 동시 런치 시 프로젝트 항목 유실.** 원장 쓰기는 read-modify-write 이고 각 쓰기는 원자적(temp + `os.Rename`)이지만 **프로세스 간 RMW 는 원자적이지 않다**(파일 잠금 없음). 종전에는 유실 가능 필드가 스칼라 `last_profile` 하나뿐이라 "마지막 런치가 이긴다"가 곧 의도된 의미였으나, `projects:` 맵이 생기면 서로 다른 프로젝트에서 동시에 실행된 두 `moai cc -p …` 가 **한 프로젝트의 항목을 통째로 잃을 수 있다**. 사용자에게는 "그 프로젝트만 기억을 잊음"으로 관측된다. 이 리포는 복수 세션 동시 운용이 일상이므로 가설이 아니라 실현 가능한 시나리오다. 잠금 도입은 위 「Out of Scope — 원장 동시 쓰기 잠금」에 따라 하지 않는다.

---

## §D 영향 표면

| 경로 | 성격 |
|------|------|
| `internal/profile/profile.go` | 원장 스키마·해석기·기록기 (주 변경) |
| `internal/cli/launcher.go` | 해석·기록 순서 재배열 + 신규 프로필 고지 + 기록 시임 |
| `internal/cli/web.go` | `ProfileName` 을 project-scoped 로 (1줄, REQ-PM-024) |
| `internal/web/app.go` | `newApp` 의 `recordLastProfile` 배선 1지점 |
| `internal/profile/main_test.go` | 신규 — 패키지 샌드박스 `TestMain` |
| `internal/profile/zz_sandbox_guard_test.go` | 신규 — 오염 이후 실행되는 후행 가드 |
| `internal/web/main_test.go` | 신규 — 패키지 샌드박스 `TestMain` |
