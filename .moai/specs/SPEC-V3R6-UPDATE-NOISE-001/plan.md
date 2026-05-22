---
title: "SPEC-V3R6-UPDATE-NOISE-001 — Implementation Plan"
version: "0.1.0"
status: draft
created: 2026-05-23
updated: 2026-05-23
author: manager-spec
priority: P2
phase: "v3.6.0"
module: "internal/cli"
lifecycle: spec-anchored
tags: "v3r6, ux, update, noise-suppression, idempotency, state-file, plan"
tier: S
---

# SPEC-V3R6-UPDATE-NOISE-001 — Implementation Plan

본 문서는 `./spec.md` 의 REQ-UN-001~011 을 milestone 단위로 분해한 구현 계획이다. 시간 추정은 포함하지 않으며 priority + ordering 으로 진행 순서를 정의한다 (CLAUDE.md §11 Time Estimation HARD 룰 준수).

---

## 1. Scope + Approach

### 1.1 Affected files

**신규 생성 (없음 — 모든 신규는 코드 외 산출물)**

- `.moai/state/reserved-acknowledged.json` — 런타임 ledger (사용자 머신, gitignored, M1 에서 첫 update 시 자동 생성)
- `.moai/cache/merge-history.json` — 런타임 counter (사용자 머신, gitignored, M2 에서 첫 fallback 시 자동 생성)

**수정 대상 (~2-3 파일)**

- `internal/cli/design_folder.go` — `checkReservedCollision` 시그니처에 `verbose bool` 추가 + ack ledger helper 통합 (M1)
- `internal/cli/update.go` — 3-way fallback 구간 `mergeYAML3Way` 호출부에 merge-history ledger 갱신 + verbose 분기 (M2)
- `cmd/moai/update.go` (Cobra entry) — `--verbose` 플래그 추가 + `checkReservedCollision` · merge-history 호출부 propagation (M3)
- `.gitignore` (로컬) + `internal/template/templates/.gitignore` (template) — `.moai/state/` 와 `.moai/cache/` 패턴 등재 확인 (M0)

**신규 테스트 (~2 파일)**

- `internal/cli/reserved_ack_test.go` — REQ-UN-001~005 + AC-UN-001~003, AC-UN-007 (verbose path), AC-UN-009 (corruption recovery), AC-UN-011 커버
- `internal/cli/merge_history_test.go` — REQ-UN-006~010 + AC-UN-004~006, AC-UN-007 (verbose path), AC-UN-008, AC-UN-012 커버

### 1.2 Approach 원칙

- **Additive only**: 기존 함수 시그니처 보존을 위해 wrapper 패턴 또는 옵션 파라미터 우선. `checkReservedCollision` 은 `verbose bool` 인자 추가가 불가피하므로 모든 caller (스캐폴드 strict 경로 포함) 동시 갱신.
- **Fail-safe parsing**: 두 ledger 파일 모두 JSON parse 실패 시 빈 객체로 재초기화 (REQ-UN-011). silent recovery 가 정책이다.
- **Atomic writes**: ledger 갱신은 temp 파일 → rename 패턴으로 부분 쓰기 방지. Go 표준 `os.WriteFile` 이 일반적으로 충분하나 큰 ledger 의 경우 `os.CreateTemp` + `os.Rename` 명시 권장.
- **Single source of truth for hash**: sha256 계산은 `crypto/sha256` + `io.Copy` (스트리밍) 로 통일. 9KB ~1ms 수준이므로 cache 없이 매 호출 재계산.
- **Test isolation**: 모든 신규 테스트는 `t.TempDir()` 기반 (`CLAUDE.local.md §6` Test Isolation HARD 룰).

### 1.3 Constraints

- [HARD] 코드 외 사용자 산출물 (예: `~/moai/mo.ai.kr/.moai/design/tokens.json`) 은 절대 수정하지 않는다 — REQ-DFF-004 보존 (Non-goal B.2).
- [HARD] 기존 warning 문자열 ("warning: reserved filename: ..." + "Warning: 3-way merge failed for ...") 은 첫 occurrence 와 `--verbose` 출력에서 byte-identical 유지. grep 기반 회귀 테스트 안정성 확보.
- [HARD] state 파일은 git 추적 금지. `.gitignore` 패턴 추가는 M0 의 acceptance gate.

---

## 2. Milestones

각 milestone 은 priority + ordering 으로 정의한다 (no time estimate).

### M0 — State file 인프라 + .gitignore 보호

**Priority**: Critical (다른 모든 M 의 전제)

**Goal**: 두 ledger 파일이 사용자 git history 로 누출되지 않도록 `.gitignore` 패턴 정비.

**Steps**:

1. 로컬 `.gitignore` 에 `.moai/state/` 와 `.moai/cache/` 패턴 존재 여부 확인. 누락 시 추가.
2. `internal/template/templates/.gitignore` (template baseline) 에도 동일 패턴 미러. Template-First Rule (CLAUDE.local.md §2) 준수.
3. `make build` 로 embed 재생성.
4. AC-UN-010 검증: `git check-ignore -v .moai/state/reserved-acknowledged.json` 가 매칭 출력해야 한다.

**Dependencies**: 없음 (entry-point milestone)

**Exit criteria**: AC-UN-010 PASS.

---

### M1 — Reserved ack ledger 구현 (Defect #3)

**Priority**: High

**Goal**: REQ-UN-001~005 + REQ-UN-011 (corruption recovery) 구현.

**Steps**:

1. `internal/cli/design_folder.go` 에 ledger I/O helper 추가:
   - `type AckEntry struct { SHA256 string `json:"sha256"`; AcknowledgedAt string `json:"acknowledged_at"` }`
   - `loadAckLedger(projectRoot string) map[string]AckEntry` — fail-safe (parse 실패 → empty map)
   - `saveAckLedger(projectRoot string, ledger map[string]AckEntry) error` — atomic write (temp + rename)
   - `sha256File(path string) (string, error)` — `crypto/sha256` + `io.Copy`
2. `checkReservedCollision` 시그니처 변경: `func checkReservedCollision(projectRoot string, errOut io.Writer, strict bool, verbose bool) error`. 모든 caller 동시 갱신 (스캐폴드 strict 경로 포함).
3. update path (strict=false) 에서 reserved 충돌 발견 시:
   - `verbose == true` → 무조건 warning emit + ledger 갱신
   - `verbose == false` + 항목 없음 또는 hash mismatch → warning emit + ledger 갱신
   - 그 외 → silent 통과
4. ledger 디렉토리 (`.moai/state/`) 가 없으면 `os.MkdirAll(_, 0755)` 로 생성.

**Test coverage** (`reserved_ack_test.go`):

- `TestReservedAck_FirstOccurrenceEmitsWarning`: 첫 update → warning 출력 + ledger 생성
- `TestReservedAck_SecondCallSilent`: 동일 파일 두 번째 update → warning 없음
- `TestReservedAck_HashDriftReemits`: 파일 내용 변경 후 → warning 재출현 + ledger sha256 갱신
- `TestReservedAck_VerboseBypass`: `verbose=true` → 항목 존재해도 warning 출력
- `TestReservedAck_CorruptedLedgerRecovers`: 손상 JSON → parse 실패 → empty map → warning emit 정상

**Dependencies**: M0 완료 (`.gitignore` 패턴).

**Exit criteria**: AC-UN-001, AC-UN-002, AC-UN-003, AC-UN-009, AC-UN-011 PASS.

---

### M2 — Merge-history ledger 구현 (Defect #4)

**Priority**: High (M1 과 병행 가능)

**Goal**: REQ-UN-006~009 + REQ-UN-011 구현.

**Steps**:

1. `internal/cli/update.go` 에 ledger I/O helper 추가 (M1 의 `loadAckLedger` 패턴 미러):
   - `type MergeHistEntry struct { FallbackCount int `json:"fallback_count"`; LastFailedAt string `json:"last_failed_at"` }`
   - `loadMergeHistory(projectRoot string) map[string]MergeHistEntry`
   - `saveMergeHistory(projectRoot string, hist map[string]MergeHistEntry) error`
2. `update.go:1739-1745` 의 3-way fallback 분기를 다음 helper 호출로 래핑:
   ```go
   recordMergeFallback(projectRoot, relPath, false /*success*/, verbose, os.Stderr)
   ```
3. `recordMergeFallback` 정책:
   - `success == true` → `FallbackCount = 0` (REQ-UN-009)
   - `success == false` + `verbose == true` → 기존 메시지 emit + counter 증가
   - `success == false` + `verbose == false` + `FallbackCount + 1 == 3` → advisory 한 줄 emit + counter 증가
   - `success == false` + `verbose == false` + `FallbackCount + 1 != 3` → silent + counter 증가 (`>= 4` 도 silent — advisory once-per-relPath)
4. ledger 디렉토리 (`.moai/cache/`) 자동 생성.
5. 3-way merge 성공 경로 (`os.WriteFile(targetPath, merged, defs.FilePerm)` 직전) 에도 `recordMergeFallback(_, relPath, true, ...)` 호출 추가하여 counter reset 확보.

**Advisory 메시지 정확 wording** (acceptance grep 대상):

```
hint: 'moai update -c' to resync templates for <relPath>
```

(소문자 hint, single quote 로 둘러싼 `moai update -c`, "to resync templates for" + relPath)

**Test coverage** (`merge_history_test.go`):

- `TestMergeHistory_FirstTwoFallbacksSilent`: 1·2 회 fallback → stderr 빈 출력
- `TestMergeHistory_ThirdFallbackEmitsAdvisory`: 3 회차 → advisory 한 줄
- `TestMergeHistory_FourthFallbackSilent`: 4 회차 → silent (advisory once-per-relPath)
- `TestMergeHistory_SuccessResetsCounter`: 성공 → `FallbackCount = 0`
- `TestMergeHistory_VerboseBypass`: `verbose=true` → 매 회차마다 기존 warning emit
- `TestMergeHistory_CorruptedLedgerRecovers`: 손상 JSON → empty map → 정상 진행

**Dependencies**: M0 완료.

**Exit criteria**: AC-UN-004, AC-UN-005, AC-UN-006, AC-UN-008, AC-UN-012 PASS.

---

### M3 — `--verbose` 플래그 통합

**Priority**: Medium

**Goal**: REQ-UN-005 + REQ-UN-010 의 verbose escape hatch 를 사용자 CLI 표면에 노출.

**Steps**:

1. `cmd/moai/update.go` (Cobra entry) 에 persistent flag 추가:
   ```go
   updateCmd.Flags().BoolVar(&verboseFlag, "verbose", false, "Show all warnings including acknowledged reserved-name and 3-way merge fallback notices (diagnostic mode)")
   ```
2. flag 값을 `checkReservedCollision` 와 `recordMergeFallback` propagate.
3. 기존 다른 `--verbose` 사용처가 있다면 conflict 검토 (현재 없음으로 추정 — `grep -n '"verbose"' cmd/moai/*.go` 로 확인).

**Test coverage**: `reserved_ack_test.go` 와 `merge_history_test.go` 의 verbose bypass 테스트가 충분 (별도 cmd-level 통합 테스트 권장이나 필수 아님).

**Dependencies**: M1 + M2 완료.

**Exit criteria**: AC-UN-007 PASS.

---

### M4 — End-to-end 검증 + cross-platform

**Priority**: Medium

**Goal**: 전체 verification batch 실행 + macOS/Windows 호환성 확인.

**Steps**:

1. `go test ./internal/cli/...` PASS (전체 패키지 회귀).
2. `go test -race ./internal/cli/...` PASS (concurrent map 접근 없음 확인).
3. `golangci-lint run --timeout=2m ./internal/cli/...` 0 NEW issues.
4. Cross-platform: `GOOS=windows go build ./...` + `GOOS=darwin go build ./...` 실행 — 경로 separator (`filepath.Join`) · file mode 0755 호환 확인.
5. C-HRA-008 sentinel scan: `grep -rn 'AskUserQuestion\|mcp__askuser' internal/cli/ | grep -v "_test.go" | grep -v "^[^:]*:[0-9]*:[ \t]*//"` 0 matches (변경 무관 확인).
6. Manual smoke (사용자 실제 환경 옵션): `cd ~/moai/mo.ai.kr && moai update` 2회 연속 실행 → 두 번째 stderr 에 reserved warning 부재 확인.

**Dependencies**: M1 + M2 + M3 완료.

**Exit criteria**: 전체 verification batch PASS, AC-UN-011 + AC-UN-012 PASS.

---

### M5 — chore + status flip + progress.md

**Priority**: Low (sync-phase 이전 wrap-up)

**Goal**: SPEC artifacts 메타데이터 갱신 + 증거 기록.

**Steps**:

1. `progress.md` 작성 (Section A-E 증거):
   - M0~M4 commit hash 나열
   - 각 AC verification 명령 출력 evidence
   - Known limitations (R4 ledger growth) 명시
2. `spec.md` · `plan.md` · `acceptance.md` frontmatter 갱신:
   - `status: draft → implemented`
   - `version: 0.1.0 → 0.2.0`
   - `updated: <implementation merge date>`
3. HISTORY 섹션에 v0.2.0 entry 추가 (3 artifacts 모두).

**Dependencies**: M4 완료.

**Exit criteria**: SPEC artifacts 메타데이터가 implemented 상태로 일관성 있게 갱신.

---

## 3. Technical Approach (deeper detail)

### 3.1 Hash 계산 결정

`crypto/sha256` 표준 라이브러리 사용. 9KB ~ 100KB 파일 단발 sha256 은 < 5ms — cache 도입 불필요. 만약 reserved 파일이 매우 크면 (예: > 10MB) 고민 대상이지만, 본 SPEC 의 reserved 충돌 케이스는 사용자 design system 산출물 (수 KB ~ 수십 KB) 로 한정된다. 따라서 streaming hash (`io.Copy` + `sha256.New()`) 패턴으로 충분.

### 3.2 Atomic write 패턴

```go
func saveAckLedger(projectRoot string, ledger map[string]AckEntry) error {
    dir := filepath.Join(projectRoot, ".moai", "state")
    if err := os.MkdirAll(dir, 0755); err != nil { return err }
    tmp, err := os.CreateTemp(dir, "reserved-acknowledged.*.tmp")
    if err != nil { return err }
    enc := json.NewEncoder(tmp)
    enc.SetIndent("", "  ")
    if err := enc.Encode(ledger); err != nil {
        os.Remove(tmp.Name())
        return err
    }
    tmp.Close()
    return os.Rename(tmp.Name(), filepath.Join(dir, "reserved-acknowledged.json"))
}
```

같은 패턴이 `saveMergeHistory` 에도 적용된다.

### 3.3 Verbose flag propagation

현재 `checkReservedCollision` 는 update CLI 외에도 init CLI (scaffold, strict=true) 에서 호출된다. strict=true 경로는 verbose 영향 없음 (`return error` 가 우선) 이지만 시그니처 통일을 위해 `verbose bool` 인자를 모든 caller 에서 전달. 스캐폴드 시점 (init) 에는 `verbose=false` 하드코딩.

### 3.4 Counter overflow 우려

`FallbackCount int` 은 Go int (64-bit) 이므로 사실상 overflow 불가능 (2^63 - 1 = 9.2 × 10^18 회 fallback 필요). int wrap-around 없음.

### 3.5 Concurrency

`moai update` 는 단일 프로세스 직렬 실행이 보장됨 (사용자가 동일 프로젝트에 두 번 동시 실행하지 않음 — file lock 미구현이 known limitation 이지만 본 SPEC 범위 밖). 따라서 ledger I/O 는 mutex 없이 sequential. race 테스트는 회귀 방지 목적.

---

## 4. Test Strategy

### 4.1 Test pyramid

- **Unit tests** (`internal/cli/reserved_ack_test.go` + `internal/cli/merge_history_test.go`) — REQ-UN-001~011 의 모든 분기 커버
- **Integration tests** — 별도 신규 생성 없음. 기존 `update_test.go` (있다면) 의 시나리오 회귀로 충분.
- **Manual smoke** (M4 step 6) — 사용자 실 환경 (`~/moai/mo.ai.kr`) 에서 두 defect 의 재현 → 억제 확인

### 4.2 Test isolation

모든 신규 테스트는 `t.TempDir()` 사용:

```go
func TestReservedAck_FirstOccurrenceEmitsWarning(t *testing.T) {
    projectRoot := t.TempDir()
    os.MkdirAll(filepath.Join(projectRoot, ".moai", "design"), 0755)
    os.WriteFile(filepath.Join(projectRoot, ".moai", "design", "tokens.json"), []byte("{}"), 0644)
    var buf bytes.Buffer
    err := checkReservedCollision(projectRoot, &buf, false, false)
    require.NoError(t, err)
    assert.Contains(t, buf.String(), "warning: reserved filename: tokens.json")
}
```

`CLAUDE.local.md §6` 의 "All test temp directories MUST be created under `/tmp` and cleaned up automatically" 룰 준수.

### 4.3 Coverage 목표

- `internal/cli/design_folder.go` 의 신규 helper (loadAckLedger · saveAckLedger · sha256File) coverage ≥ 90%
- `internal/cli/update.go` 의 신규 helper (loadMergeHistory · saveMergeHistory · recordMergeFallback) coverage ≥ 90%
- 패키지 전체 coverage 회귀 없음 (기존 baseline 유지)

---

## 5. Rollout

### 5.1 Backward compatibility

- 기존 `~/moai/mo.ai.kr` 환경에서 ledger 파일 부재 시 첫 `moai update` 실행으로 자동 생성 — 사용자 액션 불필요.
- 기존 warning 문자열 byte-identical 유지 — 사용자 스크립트가 stderr 를 grep 한다면 첫 occurrence 시 동일 동작.
- 사용자 ack 가 자동 기록되므로, "이전에 본 warning 이 다시 안 보인다" 는 노이즈 감소 효과를 즉시 체감 (intended).

### 5.2 Documentation 의무

본 SPEC 자체가 user-visible 동작 변경이므로 별도 `.moai/docs/` 또는 docs-site 업데이트는 필요하지 않다 (silent improvement). 다만 `moai update --help` 출력의 `--verbose` 라인이 자동으로 사용자에게 노출되므로 flag description 정확성에 주의.

### 5.3 Sync-phase 책임 분리

본 plan 은 run-phase 까지만 커버한다. PR 머지 후의 다음 산출물은 sync-phase (`/moai sync`) 책임:

- CHANGELOG.md v3.6.0 entry
- README 의 "What's New" 섹션 (선택)
- GitHub Release Notes (release-drafter 자동)

---

## Out of Scope

### Out of Scope — 본 plan 범위 외 작업

- SPEC-V3R6-UPDATE-ARCHIVE-CONTRACT-001 — `--force` archive 전파 + skip-sync 단락 (별도 SPEC)
- SPEC-V3R6-UPDATE-PROGRESS-001 — `\r` overwrite 출력 깨짐 정정 (별도 SPEC)
- Reserved filename 경고 자체의 제거 (REQ-DFF-004 user data preservation 위반)
- 3-way merge fallback 제거 (안전한 fallback chain 보존)
- ack ledger TTL/LRU eviction (장기 사용 시 unbounded growth는 별도 SPEC 후보)

---

End of plan.md
