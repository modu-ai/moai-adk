---
id: SPEC-V3R6-UPDATE-NOISE-001
title: "SPEC-V3R6-UPDATE-NOISE-001 — Implementation Progress"
version: "0.2.0"
status: implemented
created: 2026-05-22
updated: 2026-05-22
author: manager-develop
priority: P2
phase: "v3.6.0"
module: "internal/cli"
lifecycle: spec-anchored
tags: "v3r6, ux, update, noise-suppression, idempotency, state-file, progress"
tier: S
---

# SPEC-V3R6-UPDATE-NOISE-001 — Implementation Progress

본 문서는 `/moai run SPEC-V3R6-UPDATE-NOISE-001` (Tier S DDD cycle_type=ddd) 결과물의 evidence trail 이다. M0~M5 milestone 별 변경 + 12 AC binary 검증 + 운영 lesson candidates 를 담는다.

---

## 1. Milestone Evidence

### M0 — State file 인프라 + `.gitignore` 보호 (Critical)

**Status**: PASS

- 로컬 `.gitignore` line 236 (`.moai/cache/`) 과 line 238 (`.moai/state/`) 가 이미 존재 — 추가 변경 불필요.
- `internal/template/templates/.gitignore` 에 새 섹션 추가 (line 186-192): "Update Noise-Suppression Ledgers (per-machine, never tracked)" 헤더 + `.moai/state/` + `.moai/cache/` 패턴.
- Template-First Rule (CLAUDE.local.md §2) 준수.
- AC-UN-010 검증 (E1 표 참조).

### M1 — Reserved ack ledger 구현 (Defect #3)

**Status**: PASS

**신규 파일**: `internal/cli/update_noise.go` (188 LOC)

추가된 helper:
- `loadReservedAckLedger(projectRoot) map[string]reservedAckEntry` — JSON parse 실패 시 빈 map 반환 (REQ-UN-011)
- `saveReservedAckLedger(projectRoot, ledger) error` — `os.MkdirAll` + atomic `atomicWriteJSON` (REQ-UN-001)
- `sha256FileHex(path) (string, error)` — `crypto/sha256` streaming hash (plan.md §3.1)
- `shouldEmitReservedWarning(ledger, name, currentHash, verbose) bool` — REQ-UN-002 + REQ-UN-005 결정 매트릭스
- `recordReservedAck(ledger, name, currentHash)` — REQ-UN-003 stamp helper
- `atomicWriteJSON(targetPath, value)` — temp file + rename (plan.md §3.2)

**기존 파일 변경**: `internal/cli/design_folder.go`
- `checkReservedCollision` body 에 ack-ledger 분기 통합 (signature 보존 — `verbose` 는 package-level `updateVerboseMode` 통해 전달)
- `reservedExact` 루프 + `reservedGlobs` walk 양쪽에 `shouldEmitReservedWarning` + `recordReservedAck` 호출 삽입
- `strict=true` 경로는 ack-ledger 미참조 (scaffold 의 hard-error 보존)
- 함수 끝에서 `ledgerDirty == true` 일 때만 `saveReservedAckLedger` 호출 (불필요한 write 방지)

### M2 — Merge-history ledger 구현 (Defect #4)

**Status**: PASS

**`update_noise.go` 에 추가된 helper**:
- `loadMergeHistoryLedger(projectRoot) map[string]mergeHistoryEntry`
- `saveMergeHistoryLedger(projectRoot, hist) error`
- `recordMergeFallback(projectRoot, relPath, success, verbose, errOut)` — REQ-UN-007/008/009/010 결정 매트릭스

**기존 파일 변경**: `internal/cli/update.go`
- `restoreMoaiConfig` 의 3-way merge 호출부 (around line 1769-1780) 수정
- 성공 경로: `recordMergeFallback(projectRoot, relPath, true, updateVerboseMode, os.Stderr)` 호출하여 counter reset (REQ-UN-009)
- 실패 경로: 기존 unconditional `fmt.Fprintf` 제거 → `recordMergeFallback(projectRoot, relPath, false, updateVerboseMode, os.Stderr)` 로 대체 (REQ-UN-007/008/010 위임)
- 2-way fallback 메시지 (`Warning: merge failed for X, restoring backup`) 는 기존 그대로 보존 — 본 SPEC 범위 밖.

**Advisory wording** (acceptance grep 대상 — byte-identical):
```
hint: 'moai update -c' to resync templates for <relPath>
```

### M3 — `--verbose` 플래그 통합

**Status**: PASS

- Pre-flight 확인: `grep -n '"verbose"\|--verbose' cmd/moai/*.go internal/cli/*.go` → 충돌 없음 (doctor 의 `-v, --verbose` 와 다른 명령). 결과:
  ```
  internal/cli/doctor.go:62: doctorCmd.Flags().BoolP("verbose", "v", false, ...)
  ```
- `internal/cli/update.go` `init()` 에 `updateCmd.Flags().Bool("verbose", false, ...)` 추가 (line 91).
- `runUpdate` 진입부에 `updateVerboseMode = getBoolFlag(cmd, "verbose")` + `defer reset` 추가.
- `cmd/moai/update.go` 는 존재하지 않음 — update 명령은 `internal/cli/update.go` 의 `var updateCmd` 로 등록된다.

### M4 — End-to-end 검증 + cross-platform

**Status**: PASS

- `go build ./...` exit 0
- `GOOS=windows GOARCH=amd64 go build ./...` exit 0
- `GOOS=linux GOARCH=amd64 go build ./...` exit 0
- `go test -run TestReservedAck ./internal/cli/ -v` PASS (5/5 sub-tests)
- `go test -run TestMergeHistory ./internal/cli/ -v` PASS (6/6 sub-tests)
- `go test -count=1 ./internal/cli/` baseline 6 FAILs identical (TestDoctor_*, TestStatus_* version drift — 본 SPEC 무관)
- `go test -cover ./internal/cli/` → 70.7% (baseline 70.5% — +0.2% delta)
- C-HRA-008 grep → 0 NEW matches in production code
- golangci-lint NEW issues = 0 (baseline 17 issues retained identical: errcheck 6 + ineffassign 1 + unused 10)

### M5 — chore + status flip + progress.md

**Status**: PASS

- `spec.md` frontmatter: status `draft → implemented`, version `0.1.0 → 0.2.0`, HISTORY v0.2.0 entry 추가
- `plan.md` frontmatter: status `draft → implemented`, version `0.1.0 → 0.2.0`
- `acceptance.md` frontmatter: status `draft → implemented`, version `0.1.0 → 0.2.0`, AC-UN-005 reproducer recipe 추가 (SHOULD-FIX 반영)
- spec.md §E.1 에 **R6 (Low) — Concurrent `moai update` ledger race** 추가 (SHOULD-FIX 반영, plan.md §3.5 atomic write 패턴 cross-reference)
- 본 `progress.md` 신규 작성

---

## 2. Acceptance Criteria Binary Matrix

| AC | Status | Verification Command | Result Summary |
|----|--------|---------------------|----------------|
| AC-UN-001 | PASS | `TestReservedAck_FirstOccurrenceEmitsWarning` | 첫 update → ledger 자동 생성, schema = {path: {sha256:64hex, acknowledged_at:RFC3339}} 검증 |
| AC-UN-002 | PASS | `TestReservedAck_SecondCallSilent` | 1차 emit + 2차 silent, byte-identical 출력 비교 |
| AC-UN-003 | PASS | `TestReservedAck_HashDriftReemits` | 파일 변경 후 warning 재출현 + ledger sha256 byte-diff 검증 |
| AC-UN-004 | PASS | `TestMergeHistory_FirstTwoFallbacksSilent` | recordMergeFallback 후 ledger 자동 생성, schema = {relPath: {fallback_count:int, last_failed_at:string}} |
| AC-UN-005 | PASS | `TestMergeHistory_FirstTwoFallbacksSilent` | 1·2회 fallback → `buf.Len() == 0` |
| AC-UN-006 | PASS | `TestMergeHistory_ThirdFallbackEmitsAdvisory` + `TestMergeHistory_FourthFallbackSilent` | 3회차 advisory exactly-once, wording `hint: 'moai update -c' to resync templates for quality.yaml`, 4회차 silent |
| AC-UN-007 | PASS | `TestReservedAck_VerboseBypass` + `TestMergeHistory_VerboseBypass` | verbose=true → reserved warning 우회 emit + merge legacy warning 4/4 emit, advisory 0회 |
| AC-UN-008 | PASS | `TestMergeHistory_SuccessResetsCounter` | success=true 후 FallbackCount = 0 + 후속 1·2회 silent (fresh count) |
| AC-UN-009 | PASS | `TestReservedAck_CorruptedLedgerRecovers` + `TestMergeHistory_CorruptedLedgerRecovers` | 손상 JSON → no panic, no error, ledger 재작성 검증 |
| AC-UN-010 | PASS | `git check-ignore -v .moai/state/reserved-acknowledged.json` → 매칭 + `grep '^.moai/(state|cache)/' internal/template/templates/.gitignore` → 2 lines | 로컬 .gitignore line 236/238 매칭, template .gitignore line 190/191 매칭 |
| AC-UN-011 | PASS | `go test -run TestReservedAck ./internal/cli/ -v -count=1` | 5/5 sub-tests PASS |
| AC-UN-012 | PASS | `go test -run TestMergeHistory ./internal/cli/ -v -count=1` | 6/6 sub-tests PASS |

**총 12/12 binary PASS.**

---

## 3. Known Limitations / Follow-up Candidates

- **R4 — Ack ledger 무한 성장**: TTL/LRU eviction 미구현 (Non-goal B.2). 실측 1KB 초과 사례 발생 시 follow-up SPEC 후보.
- **R6 — Concurrent update race**: file lock 미구현. atomic rename 으로 partial write 방지, last-rename-wins 시나리오는 fail-safe (R2 패턴) 로 동작. 사용자 책임 영역.
- **AC-UN-005 실 환경 검증**: 본 progress.md 는 단위 테스트로 binary PASS 판정. 실 환경 smoke (`/tmp/test-project-noise` 기반) 는 acceptance.md AC-UN-005 setup recipe 로 사용자 임의 재현 가능 (필수 아님).
- **`cmd/moai/update.go` 부재**: spec.md plan.md 가 언급한 Cobra entry 파일은 실제로는 `internal/cli/update.go` 의 `var updateCmd` 로 통합되어 있음. spawn prompt 의 "cmd/moai/update.go" 언급은 historical artifact — 실제 변경은 `internal/cli/update.go` 한 곳에 집중.

---

## 4. Operational Lesson Candidates

본 run-phase 에서 발견된 lesson candidates (memory entry 후보):

1. **단위 테스트 vs 실 환경 reproducer 비율** — Tier S 12 AC 중 8개를 단위 테스트가 직접 커버, 4개 (AC-UN-001/004/010 + AC-UN-009 부분) 는 단위 + grep 혼합. 실 환경 reproducer (AC-UN-005) 는 setup recipe 만 제공 — Tier S minimal form 의 효율적 패턴.
2. **Package-level flag 전달 패턴** — `updateVerboseMode` 를 var 로 두고 `runUpdate` 진입부에서 set + `defer reset` 으로 thread-through 함으로써 `restoreMoaiConfig` (8 test callers) signature 변경 회피. 단일 프로세스 직렬 실행 보장 (R6 known limitation) 하에 mutex 불필요.
3. **Linter intervention midstream** — `design_folder.go` 첫 Edit 직후 환경의 자동 lint 이 변경을 revert. 재-Edit 으로 복구. Tier S 단일 commit 단위에서 발생 가능한 패턴 — Edit 직후 Read 로 상태 재검증 의무.
4. **`cmd/moai/X.go` vs `internal/cli/X.go` 위치 판단** — `internal/cli/X.go` 안에 `var XCmd = &cobra.Command{}` + `init() { rootCmd.AddCommand(XCmd) }` 패턴이 표준. `cmd/moai/` 는 main entry 만 보유. SPEC plan 에서 `cmd/moai/update.go` 를 언급하더라도 실제 변경 지점은 `internal/cli/` 인 경우가 잦다.

---

End of progress.md
