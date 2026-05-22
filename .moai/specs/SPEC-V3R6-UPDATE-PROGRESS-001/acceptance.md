---
id: SPEC-V3R6-UPDATE-PROGRESS-001
title: "SPEC-V3R6-UPDATE-PROGRESS-001 — Acceptance Criteria"
version: "0.1.0"
status: draft
created: 2026-05-23
updated: 2026-05-23
author: manager-spec
priority: P3
phase: "v3.6.0"
module: "internal/cli, internal/tui"
lifecycle: spec-anchored
tags: "v3r6, ux, tui, ansi-escape, progress-line, acceptance"
tier: S
---

# Acceptance Criteria — SPEC-V3R6-UPDATE-PROGRESS-001

본 문서는 spec.md REQ-UPR-001~007에 대응하는 9개 binary acceptance criteria를 정의한다. 각 AC는 검증 명령(`verify`)과 기대 결과(`expected`)를 함께 제공하며, PASS / FAIL로 판정한다.

## Baseline (사전 상태 측정)

마이그레이션 전 baseline:

```bash
# 호출 사이트 수 (PRE-migration)
grep -rn '\\r  %s' /Users/goos/MoAI/moai-adk-go/internal/cli/ | grep -v _test.go | wc -l
# 측정값 (2026-05-23): 22 (internal/cli/update.go에서만)
# init.go + update_cleanup.go 포함 전체 PRE 카운트는 M0 시작 시 재측정
```

마이그레이션 후 (POST-migration):
- `\r  %s` 페어 count = 0
- `tui.ProgressLine\b` 호출 count = PRE 페어 수 (= PRE pairs)

## Quality Gate Criteria

본 SPEC 완료 조건:

- ✅ 모든 AC-UPR-001~009 PASS (단 AC-UPR-009 optional path은 M0 단계에서 Update method 미구현 시 N/A 허용)
- ✅ `go test ./internal/tui/` PASS
- ✅ `go test ./internal/cli/` PASS (회귀 부재)
- ✅ `golangci-lint run ./internal/tui/ ./internal/cli/` zero NEW issues (baseline drift만 허용)
- ✅ Manual smoke test (AC-UPR-008) PASS — 실제 TTY에서 garbled trailing 부재
- ✅ Cross-platform: darwin 필수 PASS, linux CI PASS

---

## AC-UPR-001 — `tui.ProgressLine` API 존재 (REQ-UPR-001)

**Description**: `internal/tui` 패키지가 `ProgressLine(out io.Writer, message string, theme *Theme) *ProgressLineHandle` 함수를 export하며, 반환된 handle은 `Done`/`Fail` 메서드를 노출한다.

**Verification**:

```bash
# API 시그니처 존재 확인 (function declaration)
grep -n '^func ProgressLine(' /Users/goos/MoAI/moai-adk-go/internal/tui/*.go

# Handle method 존재 확인
grep -nE '^func \(h \*ProgressLineHandle\) (Done|Fail)\(' /Users/goos/MoAI/moai-adk-go/internal/tui/*.go
```

**Expected**:
- `ProgressLine` function declaration in `internal/tui/` 최소 1건
- `Done` method declaration 1건
- `Fail` method declaration 1건

**PASS condition**: 위 3 grep 모두 ≥ 1건 매치.

---

## AC-UPR-002 — TTY path에서 `\r\033[2K` prefix 적용 (REQ-UPR-002)

**Description**: TTY 출력 환경에서 `Done`/`Fail` 호출 시 결과 메시지가 `\r\033[2K` ANSI escape sequence로 prefix된다.

**Verification**:

```bash
# Golden test 실행
cd /Users/goos/MoAI/moai-adk-go
go test -run TestProgressLine_TTY ./internal/tui/ -v

# 테스트 구현 내부 검증:
# - out := tmpfile.NewFile() or os.Pipe()로 TTY-like fd 생성
# - tui.ProgressLine(out, "msg", nil).Done("done")
# - captured output에 "\r\033[2K" substring 존재 확인
```

**Expected**: `TestProgressLine_TTY` PASS, golden output에 `\x1b[2K` (= `\033[2K`) 또는 `\r\x1b[2K` 패턴 포함.

**PASS condition**: `go test -run TestProgressLine_TTY` exit 0.

---

## AC-UPR-003 — Non-TTY path에서 `\n` 분리 출력 (REQ-UPR-003)

**Description**: `bytes.Buffer` 등 non-TTY writer에 `Done`/`Fail` 호출 시 progress + result가 별도 라인(`\n` 분리)으로 출력되며 ANSI escape / `\r` 부재.

**Verification**:

```bash
go test -run TestProgressLineNonTTY ./internal/tui/ -v

# 테스트 구현 내부 검증:
# - buf := &bytes.Buffer{}
# - tui.ProgressLine(buf, "Working...", nil).Done("Complete")
# - assert: strings.Count(buf.String(), "\n") == 2  // progress\n + result\n
# - assert: !strings.Contains(buf.String(), "\r")
# - assert: !strings.Contains(buf.String(), "\x1b[")  // no CSI
```

**Expected**: `TestProgressLineNonTTY` PASS.

**PASS condition**: `go test -run TestProgressLineNonTTY` exit 0.

---

## AC-UPR-004 — `'\r  %s'` 패턴 zero 마이그레이션 (REQ-UPR-004)

**Description**: `internal/cli/` 패키지 전체 (테스트 파일 제외)에서 `\r  %s` literal 패턴이 0건이어야 한다.

**Verification**:

```bash
grep -rn '\\r  %s' /Users/goos/MoAI/moai-adk-go/internal/cli/ | grep -v _test.go | wc -l
```

**Expected**: `0`

**PASS condition**: 출력값 정확히 `0`.

**Note**: 일부 정상 사용 (예: 에러 메시지 포맷팅의 일부로 `\r`이 의도된 경우)이 있다면 AC를 `\r  %s` literal 정확 매치로 한정하여 false-positive 회피. 현재 baseline 22건 모두 progress 덮어쓰기 용도임을 확인했다.

---

## AC-UPR-005 — ProgressLine 호출 사이트 수 일치 (REQ-UPR-004)

**Description**: `tui.ProgressLine(` 호출 사이트 수가 PRE-migration의 `\r  %s` 페어 수와 일치해야 한다 (각 페어가 1개 ProgressLine 호출로 통합).

**Verification**:

```bash
# POST count
grep -rn 'tui\.ProgressLine(' /Users/goos/MoAI/moai-adk-go/internal/cli/ | grep -v _test.go | wc -l

# 페어 카운트는 baseline (M0 시작 시 측정)와 비교
# baseline_pairs == post_count (단, 일부 페어가 단독 progress만 있고 result 없는 경우는 예외 — M1 site inventory 확정 시 처리)
```

**Expected**: `tui.ProgressLine(` 호출 수 ≥ 20 (페어가 progress↔(success|error) 짝이므로 사이트 수는 페어 수와 동일하거나 약간 적을 수 있음 — 예: 동일 progress 메시지에서 분기되는 경우)

**PASS condition**: ProgressLine 호출 수 ≥ 18 (보수적 하한; M1 inventory 확정 후 정확한 수치로 갱신).

---

## AC-UPR-006 — 가시 메시지 byte-identical (REQ-UPR-005)

**Description**: TTY 출력 환경에서 결과 메시지의 가시 텍스트(ANSI escape 제거 후)가 마이그레이션 이전과 byte-identical이어야 한다.

**Verification**:

```bash
# Pre-migration baseline 캡쳐 (M0 시작 시 1회 수행)
moai update --yes 2>&1 | tee /tmp/moai-update-PRE.log
# strip ANSI:
sed 's/\x1b\[[0-9;]*[A-Za-z]//g' /tmp/moai-update-PRE.log | sed 's/\r//g' > /tmp/moai-update-PRE-stripped.log

# Post-migration capture (M3 시점)
moai update --yes 2>&1 | tee /tmp/moai-update-POST.log
sed 's/\x1b\[[0-9;]*[A-Za-z]//g' /tmp/moai-update-POST.log | sed 's/\r//g' > /tmp/moai-update-POST-stripped.log

# Diff (가시 텍스트만 비교)
diff /tmp/moai-update-PRE-stripped.log /tmp/moai-update-POST-stripped.log
```

**Expected**: diff output empty (또는 baseline drift만 — 메시지 wording 변경 없음).

**PASS condition**: 가시 메시지 string 변경 없음. 단, PRE 로그에 garbled trailing 문자가 포함된 라인은 POST에서 정상 출력되므로 라인별로 정확한 비교가 어려울 수 있다. 그 경우는 메시지 토큰 단위(예: "backed up", "Removed", "validated" 등 결과 키워드) 존재 여부로 비교:

```bash
for kw in "backed up" "validated" "deployed" "Removed" "restored" "Migrated"; do
  pre_count=$(grep -c "$kw" /tmp/moai-update-PRE-stripped.log)
  post_count=$(grep -c "$kw" /tmp/moai-update-POST-stripped.log)
  if [ "$pre_count" != "$post_count" ]; then
    echo "MISMATCH: $kw PRE=$pre_count POST=$post_count"
  fi
done
```

**PASS condition**: 모든 결과 키워드 count 일치.

---

## AC-UPR-007 — golden tests 통과 (REQ-UPR-006)

**Description**: `internal/tui` 패키지의 `TestProgressLine*` 테스트가 모두 통과한다.

**Verification**:

```bash
cd /Users/goos/MoAI/moai-adk-go
go test ./internal/tui/ -run TestProgressLine -v
```

**Expected**: 모든 `TestProgressLine` prefix 테스트 PASS. 최소 2개 (`TestProgressLine_TTY`, `TestProgressLineNonTTY`) 필수, `TestProgressLine_Update`는 REQ-UPR-007 구현 시 PASS, 미구현 시 SKIP 허용.

**PASS condition**: `go test` exit 0.

---

## AC-UPR-008 — Smoke test (실제 `moai update`) — garbled tail 부재 (REQ-UPR-002 통합)

**Description**: 실제 TTY 환경에서 `moai update --yes` 실행 시 출력에 garbled trailing 문자가 없어야 한다.

**Verification** (수동, M3 단계):

```bash
# 깨끗한 테스트 프로젝트에서 moai update 실행
cd /tmp
moai init smoke-test-progress
cd smoke-test-progress
# 약간의 변경 (선택)
moai update --yes 2>&1 | tee /tmp/moai-update-smoke.log

# Sentinel grep: "✓"로 시작하는 결과 라인에서 단어 끝에 "..." 또는 단일 문자 + "..."이 남는 패턴
# 예: "✓ .moai/config backed upg..." (g 잔여)
#     "✓ Removed .claude/settings.jsonn..." (n 잔여)
# Pattern: 단어 끝 영문자 + "..." 또는 단어 + 영문자 + "."
grep -E '✓.*[a-z]\.{3,}|✗.*[a-z]\.{3,}' /tmp/moai-update-smoke.log | grep -v "Backing\|Restoring\|Validating\|Deploying" | wc -l
```

**Expected**: `0` (결과 라인에서 garbled trailing 잔여 없음). progress 메시지 자체에 포함된 "..." (예: "Backing up...") 은 progress 라인에서만 나타나므로 grep 제외 필터로 분리.

**PASS condition**: 출력값 `0`.

**Fallback Verification** (non-TTY):

```bash
moai update --yes 2>&1 | cat | grep -E '✓.*[a-z]\.{3,}' | wc -l
# expected: 0 (non-TTY fallback도 garbled 부재)
```

---

## AC-UPR-009 — `Update` method 동작 (REQ-UPR-007 optional)

**Description**: `ProgressLineHandle.Update(message)` 호출 시 라인이 새 메시지로 재렌더링되며, TTY/non-TTY 분기에서 각각의 line-clear 의미가 보존된다.

**Verification**:

```bash
go test -run TestProgressLine_Update ./internal/tui/ -v
```

**Expected**:
- TTY: Update 호출 후 `"\r\033[2K"` prefix + 새 메시지 + (개행 없음, 다음 Done/Fail 대기)
- non-TTY: 새 메시지 + `"\n"` (별도 라인)

**PASS condition**: REQ-UPR-007 구현 시 `TestProgressLine_Update` PASS. M0 단계에서 Update 미구현 결정 시 `N/A` 처리 가능 (그 경우 spec.md REQ-UPR-007을 deferred로 명시).

---

## Edge Cases (참고)

### EC-1 — 동일 ProgressLineHandle에 Done 후 Fail 호출

설계 결정 (M0):
- Option A: idempotent (두 번째 호출 무시)
- Option B: panic (programmer error)
- 권장: Option B — 명확한 의도 표현, 테스트로 강제. `TestProgressLine_DoneAfterFail_Panics` 테스트 추가.

### EC-2 — `out`가 nil

설계 결정: `ProgressLine(nil, ...)` 시 panic (Go convention; io.Writer nil은 caller 책임). 별도 가드 불요.

### EC-3 — theme nil

기존 stateless precedent (`Spinner`, `Progress`) 와 동일하게 `LightTheme()` fallback. 검증: `TestProgressLine_NilTheme` (선택, golden test로 충분).

### EC-4 — 메시지에 `\n` 포함

설계 결정: progress message에 `\n` 포함은 호출자 오류로 간주. 본 SPEC은 single-line progress만 지원. M0 시 검증 또는 godoc 명시.

### EC-5 — 매우 긴 메시지 (터미널 폭 초과)

설계 결정: 줄바꿈 동작은 터미널이 처리. `\033[2K`는 현재 라인만 clear하므로 줄바꿈된 이전 라인은 그대로 남을 수 있음. 본 SPEC scope 외 (NG: multi-line progress bar).

---

## Definition of Done

본 SPEC은 다음 조건을 모두 만족할 때 `status: implemented` (또는 `completed`)로 전환된다:

- [ ] M0 완료: `tui.ProgressLine` API 신설, golden tests 작성
- [ ] M1 완료: `internal/cli/update.go` 22 사이트 마이그레이션
- [ ] M2 완료: `internal/cli/init.go` + `internal/cli/update_cleanup.go` 잔여 사이트 마이그레이션
- [ ] M3 완료: smoke test PASS, status/version 갱신
- [ ] AC-UPR-001 PASS
- [ ] AC-UPR-002 PASS
- [ ] AC-UPR-003 PASS
- [ ] AC-UPR-004 PASS (grep zero)
- [ ] AC-UPR-005 PASS (call site count match)
- [ ] AC-UPR-006 PASS (visible message byte-identical)
- [ ] AC-UPR-007 PASS (golden tests)
- [ ] AC-UPR-008 PASS (smoke test garbled zero)
- [ ] AC-UPR-009 PASS (Update method) — 또는 REQ-UPR-007 deferred 처리
- [ ] `go vet ./internal/tui/ ./internal/cli/` clean
- [ ] `golangci-lint run ./internal/tui/ ./internal/cli/` zero NEW issues
- [ ] `progress.md` 작성 완료 (M0-M3 evidence)
- [ ] `spec.md` `status: draft → implemented`, `version: 0.1.0 → 0.2.0`

---

## Out of Scope

### Out of Scope — 본 acceptance 범위 외

- SPEC-V3R6-UPDATE-ARCHIVE-CONTRACT-001 — `--force` archive 전파 + skip-sync 단락 (별도 SPEC)
- SPEC-V3R6-UPDATE-NOISE-001 — reserved filename + 3-way merge 노이즈 억제 (별도 SPEC)
- Animated spinners / bubbles spinner 통합 (out of scope per REQ-UPR-007 optional path)
- Color theme overhaul (cliPrimary/cliSuccess 기존 사용)
- 레거시 Windows cmd.exe 호환성 검증 (lipgloss가 자체 처리)

---

## HISTORY

- **2026-05-23 v0.1.0 draft** — 초기 작성. 9 binary AC + 5 edge case + DoD checklist. AC-UPR-004/005 grep sentinel + AC-UPR-006 visible-text diff + AC-UPR-008 manual smoke 조합으로 회귀 차단.
