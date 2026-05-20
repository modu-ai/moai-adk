---
title: 보안 노트
description: "MoAI-ADK v2.20.0-rc1 보안 강화 변경 사항 — CWE-732/214/345 매핑, 사용자 자체 점검 절차"
weight: 72
draft: false
tags: ["security", "cwe", "audit"]
---

# 보안 노트 (Security Notes)

본 페이지는 MoAI-ADK v2.20.0-rc1 시점에 도입된 **사용자 가시 보안 변경 사항**을 정리합니다. 각 항목은 CWE 매핑, 변경된 동작, 자체 점검 명령을 포함합니다.

## Why — 왜 이 페이지가 존재하는가

`SPEC-V3R5-SECURITY-CRIT-001` (PR #1032, merge commit `03a2552a2`) 은 v2.14.0 → v2.20.0-rc1 사이의 코드 리뷰에서 발견된 **P0 release blocker 보안 결함 3건**을 정정했습니다. 본 페이지는 그 정정 사실과 사용자가 자신의 환경에서 새 보호가 작동하는지 확인할 수 있는 절차를 4-locale 공식 안내로 명문화합니다.

세 결함은 모두 GLM 통합 + 자동 업데이트 경로와 관련됩니다:

- **CWE-732 / CWE-552** — `.claude/settings.local.json` 파일 mode `0o600` 강제 (소유자 전용 read/write)
- **CWE-214** — `moai cg` 의 tmux 환경 변수 주입이 argv 대신 source-file 경유 (GLM token argv 비가시화)
- **CWE-345** — `moai update` 의 checksum 검증 mandatory (다운로드 실패 시 update 거부)

각 항목은 회귀 테스트로 잠겨 있어 미래 회귀가 차단됩니다.

## CWE-732 — settings.local.json 권한 강화 (Permission Hardening) {#cwe-732}

### 변경 사항

`.claude/settings.local.json` 파일이 생성·갱신될 때 파일 권한이 **`0o600`** (소유자만 read/write) 로 강제됩니다. 이전에는 `0o644` (소유자 read/write + group/world read) 로 생성되어 다중 사용자 워크스테이션에서 다른 로컬 사용자가 `ANTHROPIC_AUTH_TOKEN` 등 민감 자격증명을 읽을 수 있었습니다.

### 위협 모델

- **공격자**: 같은 호스트의 저권한 로컬 사용자
- **공격 표면**: `.claude/settings.local.json` 의 group/world read 권한
- **누설 정보**: GLM API token (`ANTHROPIC_AUTH_TOKEN`), OAuth refresh token, 기타 `settings.Env` 값
- **CWE 매핑**: CWE-732 (Incorrect Permission Assignment for Critical Resource), CWE-552 (Files or Directories Accessible to External Parties)

### 구현 위치

- `internal/hook/settings_io.go` — `secureSettingsMode os.FileMode = 0o600` 상수 + `writeSettingsSecure` 헬퍼
- `internal/hook/session_start.go` — `ensureGLMCredentials`, `ensureClaudeEnvFile` 등 모든 `settings.local.json` writer
- `internal/hook/session_end.go` — GLM keys write-back 경로

### 자체 점검

기존 `settings.local.json` 권한 확인:

```bash
# Linux
stat -c '%a' .claude/settings.local.json
# 기대값: 600

# macOS
stat -f '%A' .claude/settings.local.json
# 기대값: 600
```

권한이 `644` 또는 그 외 더 느슨한 값으로 표시되면, MoAI-ADK 가 다음 세션 시작 시 자동으로 `0o600` 으로 정정합니다. 즉시 정정하려면:

```bash
chmod 0600 .claude/settings.local.json
```

### 영향 (Trade-off)

`group-readable` 을 기대하는 워크플로우 (동일 프로젝트 디렉터리를 별도 OS 사용자가 read 하는 매우 드문 시나리오) 는 깨질 수 있습니다. 이 트레이드오프는 의도된 것이며 보안 회복이 분명한 우선입니다.

## CWE-214 — tmux IPC token argv 노출 차단 {#cwe-214}

### 변경 사항

`moai cg` (CG 모드) 가 GLM token (`ANTHROPIC_AUTH_TOKEN`) 을 tmux 세션 환경 변수에 주입할 때, **argv 채널** (`tmux set-environment <KEY> <VALUE>`) 대신 **source-file 채널** (`tmux source-file <tmp>`) 을 사용합니다. token 은 더 이상 `ps auxe`, `/proc/<pid>/cmdline`, auditd 로그, sysmon 추적, 크래시 덤프에 평문으로 노출되지 않습니다.

### 구현 흐름

1. `~/.moai/run/` 아래 임시 파일을 `mkstemp` 로 생성 (mode `0o600` 자동 + 명시적 `chmod 0o600`)
2. `set-environment -t <session> <KEY> <VALUE>` 한 줄을 임시 파일에 기록
3. `tmux source-file <tmp>` 로 tmux 가 그 파일을 읽어 환경에 주입
4. 주입 직후 임시 파일을 `os.Remove` 로 unlink

argv 에는 임시 파일 경로만 노출되며 token 자체는 노출되지 않습니다.

### 위협 모델

- **공격자**: 같은 호스트의 로컬 사용자 + 시스템 로그 수집 (`ps`, `/proc`, auditd, sysmon)
- **공격 표면**: tmux env injection 의 argv 채널
- **누설 정보**: GLM API token 의 순간적 가시화
- **CWE 매핑**: CWE-214 (Invocation of Process Using Visible Sensitive Information)

### 구현 위치

- `internal/tmux/session.go` — `InjectSensitiveEnv` 메서드, `sensitiveTempDir = ".moai/run"`, `mkstemp` + `chmod 0o600` + `tmux source-file` + `os.Remove`
- `internal/tmux/errors.go` — `ErrTmuxSensitiveInjectFailed` sentinel
- `internal/hook/glm_tmux.go` — `ensureTmuxGLMEnv` 에서 `ANTHROPIC_AUTH_TOKEN` 만 sensitive 경로로 분기 (나머지 URL, model 이름 등 non-sensitive 값은 기존 argv 경로 유지)

### Non-sensitive 값은 argv 유지

`CLAUDE_CONFIG_DIR` (디렉터리 경로), `ANTHROPIC_BASE_URL` (URL), `ANTHROPIC_DEFAULT_*_MODEL` (모델 이름) 등 token 이 아닌 값은 argv 경로를 유지합니다. 이는 명시적 의도이며 토큰 누설 위험과 무관합니다.

### 실패 시 동작

source-file 주입이 실패하면 (디스크 가득 참, tmux source-file 실패 등) **argv fallback 으로 누설하지 않고** `ErrTmuxSensitiveInjectFailed` sentinel error 를 반환하여 주입 자체를 abort 합니다.

### 자체 점검

CG 모드 실행 중 token 이 argv 에 노출되는지 확인:

```bash
# moai cg 실행 후 새 tmux 세션 내에서
ps auxe | grep -i 'tmux set-environment.*ANTHROPIC_AUTH_TOKEN'
# 기대값: 0 matches (token 이 argv 에 없음)
```

임시 파일이 정상적으로 unlink 되는지 확인:

```bash
ls -la ~/.moai/run/ 2>/dev/null
# 기대값: 빈 디렉터리 또는 stale 파일 없음
```

세션 종료 후 `~/.moai/run/` 에 잔존 파일이 있다면 수동으로 제거 가능합니다 (보안 위협은 아님 — 이미 unlink 시도된 파일).

### 사용자 책임

`~/.moai/.env.glm` source 파일은 사용자 환경에서 `0o600` 권한을 유지해야 합니다. 이는 `moai glm` 명령이 자동으로 설정합니다:

```bash
stat -c '%a' ~/.moai/.env.glm    # Linux: 600
stat -f '%A' ~/.moai/.env.glm    # macOS: 600
```

자세한 내용: [CG 모드](/ko/multi-llm/cg-mode/)

## CWE-345 — Update 흐름 mandatory checksum 검증 {#cwe-345}

### 변경 사항

`moai update` 의 자동 업데이트 흐름이 **checksum 검증을 우회할 수 없습니다**. release 의 `checksums.txt` 다운로드가 실패하거나 파싱이 실패하면 sentinel error `ErrChecksumUnavailable` 를 반환하고 update 흐름을 **abort** 합니다 — binary 다운로드를 시도하지 않습니다.

### Retry 정책

`checksums.txt` 다운로드는 **3회 retry** 를 지수 백오프로 시도합니다:

| 시도 | 대기 시간 |
|------|-----------|
| 1차 (즉시) | 0s |
| 2차 retry | 2s 대기 |
| 3차 retry | 4s 대기 |
| 추가 retry 없음 | 합계 ~6s 대기 후 실패 |

(내부 구현: base delay 2s × 2^(attempt-1) 지수 백오프)

모든 retry 가 실패하면 `ErrChecksumUnavailable` sentinel 로 종료합니다. **`--skip-checksum` 같은 우회 옵션은 존재하지 않습니다**.

### Defense-in-depth

`version.Checksum` 필드가 empty string 인 상태로 `downloadAndVerify` 에 도달하면 binary 다운로드를 진행하지 않고 `ErrChecksumUnavailable` 를 반환합니다. 이중 보호 (checker 단계 + updater 단계) 로 묵음 우회를 차단합니다.

### 위협 모델

- **공격자**: 네트워크 MITM (전체 차단은 못하지만 `checksums.txt` URL 만 선택 차단·throttle 가능)
- **공격 표면**: checksums.txt 없이도 binary 가 설치되던 silent fallback
- **누설 결과**: 서명되지 않은 백도어 바이너리의 무경고 설치
- **CWE 매핑**: CWE-345 (Insufficient Verification of Data Authenticity)

### 구현 위치

- `internal/update/checker.go` — `downloadChecksumWithRetry(checksumsURL, archiveName, maxAttempts, baseDelay)` (`defaultChecksumMaxAttempts=3`, `defaultChecksumBaseDelay=2*time.Second`), `ErrChecksumUnavailable` sentinel
- `internal/update/updater.go` — `downloadAndVerify` empty-checksum guard
- domain whitelist (`https://github.com/modu-ai/moai-adk/...`) 는 기존 그대로 유지 (SSRF 표면 변화 없음)

### 자체 점검

```bash
# release 정보 + checksums.txt 존재 확인
moai update --check-only

# 정상 흐름 (성공 시)
moai update
# 출력 예: Downloaded checksums.txt (verified)

# checksums.txt 다운로드 실패 시 (의도적 차단 예: VPN 단절 후 실행)
moai update
# 출력 예: error: checksum unavailable: persistent retry failure after 3 attempts
```

`ErrChecksumUnavailable` 메시지가 표시되면:

1. 네트워크 연결 확인 (`curl -I https://github.com/modu-ai/moai-adk/releases/latest`)
2. Proxy / firewall 이 GitHub release asset 도메인을 허용하는지 확인
3. 일시적 GitHub CDN 장애 가능성 — 잠시 후 재시도
4. **`--skip-checksum` 같은 우회 옵션은 제공되지 않습니다** — 이는 의도된 정책

영구 차단 시 수동 binary 설치를 권장합니다:

```bash
# 수동 설치 (사용자가 직접 무결성 검증)
curl -fsSL https://raw.githubusercontent.com/modu-ai/moai-adk/main/install.sh | bash
```

자세한 내용: [업데이트](/ko/getting-started/update/)

## 자체 점검 체크리스트 (Self-Audit Checklist)

```bash
# 1. CWE-732 — settings.local.json 권한
stat -c '%a' .claude/settings.local.json 2>/dev/null \
  || stat -f '%A' .claude/settings.local.json 2>/dev/null
# 기대값: 600

# 2. CWE-214 — CG 모드 실행 중 token argv 노출 (cg 모드 활성 상태에서)
ps auxe 2>/dev/null | grep -i 'tmux set-environment.*ANTHROPIC_AUTH_TOKEN'
# 기대값: 0 matches

# 3. CWE-214 — tmux sensitive temp 디렉터리 정합성
ls -la ~/.moai/run/ 2>/dev/null
# 기대값: 빈 디렉터리 또는 stale 파일 없음

# 4. CWE-345 — Update flow checksum 동작
moai update --check-only
# 기대값: release + checksums.txt 정상 확인

# 5. GLM source 파일 권한 (사용자 책임)
stat -c '%a' ~/.moai/.env.glm 2>/dev/null \
  || stat -f '%A' ~/.moai/.env.glm 2>/dev/null
# 기대값: 600 (해당 파일이 존재하는 경우)
```

위 5 항목이 모두 기대값을 충족하면 v2.20.0-rc1 보안 강화가 정상 작동하고 있습니다.

## References

### CHANGELOG

[CHANGELOG `[Unreleased]` v2.20.0-rc1 Security 섹션](https://github.com/modu-ai/moai-adk/blob/main/CHANGELOG.md)

### SPEC

- `SPEC-V3R5-SECURITY-CRIT-001` — upstream source of truth, status `implemented` v0.2.0
- PR #1032 merge commit `03a2552a2`

### Commits

- `b48bd86cb` — M1 settings.local.json 0o600 hardening (CWE-732/552)
- `10776c4b8` — M2 tmux sensitive env source-file injection (CWE-214)
- `ee1335282` — M3 mandatory checksum verification with retry (CWE-345)
- `b4e7115cb` — M4 cross-cutting verification + frontmatter

### CWE / OWASP

- [CWE-732](https://cwe.mitre.org/data/definitions/732.html) — Incorrect Permission Assignment for Critical Resource
- [CWE-552](https://cwe.mitre.org/data/definitions/552.html) — Files or Directories Accessible to External Parties
- [CWE-214](https://cwe.mitre.org/data/definitions/214.html) — Invocation of Process Using Visible Sensitive Information
- [CWE-345](https://cwe.mitre.org/data/definitions/345.html) — Insufficient Verification of Data Authenticity

### 관련 페이지

- [settings.json 가이드](/ko/advanced/settings-json/) — `settings.local.json` 권한 섹션
- [업데이트](/ko/getting-started/update/) — checksum 검증 섹션
- [CG 모드](/ko/multi-llm/cg-mode/) — tmux 환경 변수 주입 보안 모델
