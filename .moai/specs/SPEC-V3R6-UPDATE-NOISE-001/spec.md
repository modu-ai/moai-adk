---
id: SPEC-V3R6-UPDATE-NOISE-001
title: "moai update 반복 경고 노이즈 억제 (reserved filename ack + 3-strike merge fallback)"
version: "0.2.0"
status: implemented
created: 2026-05-23
updated: 2026-05-22
author: manager-spec
priority: P2
phase: "v3.6.0"
module: "internal/cli"
lifecycle: spec-anchored
tags: "v3r6, ux, update, noise-suppression, idempotency, state-file"
tier: S
---

# SPEC-V3R6-UPDATE-NOISE-001 — moai update 반복 경고 노이즈 억제

## HISTORY

- 2026-05-23 — v0.1.0 draft 작성. 2026-05-23 audit 세션에서 발견된 UX defect #3 (reserved filename warning) + defect #4 (3-way merge fallback warning) 양자 통합 처리. Wave 0 후속 V3R6 신규 SPEC, Tier S 분류.
- 2026-05-22 — v0.2.0 implemented. M0~M5 완료. internal/cli/update_noise.go (helpers) + internal/cli/design_folder.go (checkReservedCollision ack 분기 통합) + internal/cli/update.go (`--verbose` flag + recordMergeFallback 호출) + internal/template/templates/.gitignore (state/cache pattern 미러) + reserved_ack_test.go (5 sub-tests) + merge_history_test.go (6 sub-tests) 추가. 12/12 AC PASS. plan-auditor SHOULD-FIX 3건 반영 (verbose 충돌 사전 점검 / AC-UN-005 recipe / R6 동시성 risk).

---

## A. Context + Motivation

### A.1 배경

`moai update` 명령은 사용자 프로젝트의 `.claude/`, `.moai/`, `.agency/` 템플릿을 최신 버전과 동기화하는 핵심 유지보수 진입점이다. 그러나 2026-05-23 audit 세션에서 두 가지 UX defect 가 동시에 노출되었다 — 모두 **반복적 경고 노이즈**가 본질이다. 사용자 데이터 보존 (REQ-DFF-004) 원칙은 유지되어야 하지만, 이미 사용자가 인지·수용한 상태에 대해 매 업데이트마다 동일한 stderr 경고가 재방출되어 silent idempotent re-run 이라는 운영자 기대를 침해한다.

### A.2 Audit 증거

**Defect #3 — Reserved filename warning repeats (Medium UX)**

`internal/cli/design_folder.go:90-92` 와 `:113-115` 는 `checkReservedCollision(strict=false)` 경로 (update path) 에서 reserved name 충돌을 발견할 때마다 무조건 다음 warning 을 emit 한다:

```
warning: reserved filename: <name> (preserved; rename to use canonical templates)
```

실제 재현 환경:

- `~/moai/mo.ai.kr/.moai/design/tokens.json` (9,258 bytes, 2026-04-27 사용자 생성)
- `~/moai/mo.ai.kr/.moai/design/components.json` (9,477 bytes, 2026-04-27 사용자 생성)

두 파일 모두 사용자 design system 산출물 (tokens · components manifest) 로 legitimate user content 이다. 매 `moai update` 호출마다 2개의 동일한 warning 이 emit 되며, 사용자는 다음 두 옵션 중에서만 선택 가능:

1. 파일 이름 변경 또는 삭제 — 사용자 작업 손실
2. warning 을 영구히 무시 — silent re-run 기대 침해

REQ-DFF-004 (user data preservation) 는 파일 변경을 금지하므로 옵션 1 은 권장 불가. 본 SPEC 은 옵션 2 를 silent ack 모델로 정식화한다.

**Defect #4 — 3-way merge fallback warning repeats (Low UX)**

`internal/cli/update.go:1742-1743` 는 3-way merge 실패 시 다음 warning 을 무조건 emit 한다:

```
Warning: 3-way merge failed for <relPath>, falling back to 2-way
```

Fallback chain (3-way → 2-way → backup restore) 은 well-tested 이며 데이터 손실 위험이 없다. 그러나 `.moai/config/sections/quality.yaml` 처럼 사용자가 깊이 수정한 YAML 의 경우 3-way 가 매번 재현되며, 사용자는 actionable signal 이 없는 동일 warning 을 반복 수신한다.

### A.3 노이즈 누적 비용

두 defect 모두 **단발 1회는 정보 가치, N+1회는 노이즈** 라는 공통 패턴을 보인다. `moai update` 가 정기 유지보수 흐름이라는 점을 고려할 때:

- N=10회 update → reserved warning 20회 + merge fallback warning 1회 × 영향 파일 수 = ~30+ 경고
- 사용자는 점차 stderr 출력을 무시하기 시작 → 실제 새로운 경고 (예: 새 reserved 파일 추가, merge 회귀) 가 묻힘
- "잠잠한 update" 가 운영 신호이지만, 노이즈가 그 신호를 가린다

### A.4 References

- REQ-DFF-004 — 사용자 데이터 보존 원칙 (SPEC-V3R3-DESIGN-FOLDER-FIX-001)
- SPEC-V3R3-DESIGN-FOLDER-FIX-001 — reserved name policy 도입
- `.claude/rules/moai/workflow/spec-workflow.md` — Tier S 정의 (3 artifacts, S domain)
- `internal/cli/design_folder.go:64-93` — `checkReservedCollision` 진입점
- `internal/cli/update.go:1730-1756` — 3-way merge fallback 구간

---

## B. Goals + Non-goals

### B.1 Goals

- G1: `moai update` 두 번째 실행부터는 reserved filename warning 이 출력되지 않는다 (사용자 파일이 첫 실행 이후 변경되지 않은 한).
- G2: `moai update` 1·2 회차 3-way merge fallback 은 silent 하다. 3회 연속 같은 파일이 fallback 되면 한 줄 advisory 가 emit 된다.
- G3: 사용자는 `--verbose` 플래그로 두 가지 억제를 동시에 해제하여 진단 출력을 얻을 수 있다.
- G4: 첫 occurrence 의 정보 가치는 유지된다 (silent ack 는 자동 기록되지만 user-visible warning 은 정상 emit).
- G5: 사용자 파일 hash 가 변경되면 reserved warning 이 재출현하여 새 상태에 대한 ack 를 유도한다.

### B.2 Non-goals

- 사용자 파일 변경 (REQ-DFF-004 보존). reserved 파일 자동 rename · 자동 archive · 자동 삭제는 본 SPEC 범위 밖.
- 첫 occurrence warning 제거. 첫 ack 는 반드시 user-visible warning 과 동반된다.
- Merge 알고리즘 개선 (3-way → 2-way → backup) — fallback chain 자체는 well-tested 이며 본 SPEC 범위 밖.
- ack 파일 TTL · LRU eviction. 무한 성장 대응은 R4 로 known limitation 으로 문서화하되 본 SPEC 에서 구현하지 않는다.
- 다국어 메시지 변경. 기존 warning 문자열은 보존 (테스트 grep 안정성 위해).

### B.3 Out of Scope (h3)

다음 인접 SPEC 으로 분리되었거나 별도 처리 예정:

- **SPEC-V3R6-UPDATE-ARCHIVE-CONTRACT-001** — `--force` 시맨틱 명확화 + skip-sync archive 정책. reserved 파일 자동 archive 도 여기 흡수.
- **SPEC-V3R6-UPDATE-PROGRESS-001** — `moai update` 진행 상황 spinner · progress bar 출력 UX.
- **Reserved filename warning 의 전면 제거** — REQ-DFF-004 user data preservation 의무가 지속되므로 warning 자체는 첫 occurrence 에서 반드시 emit 되어야 한다.

---

## C. Functional Requirements (GEARS notation)

### C.1 Reserved filename acknowledgment 그룹 (Defect #3)

- **REQ-UN-001 (Ubiquitous)**: 시스템은 `.moai/state/reserved-acknowledged.json` 을 ack ledger 로 사용해야 한다. 스키마는 `{path: string → {sha256: string, acknowledged_at: ISO-8601 string}}` 매핑이다. 파일이 없으면 `checkReservedCollision(strict=false)` 첫 호출 시 빈 JSON 객체 `{}` 로 생성한다.

- **REQ-UN-002 (Event-driven)**: `checkReservedCollision(strict=false)` 가 reserved name 충돌을 발견했을 때, 시스템은 ack ledger 에서 해당 경로 항목을 조회해야 한다. 항목이 없거나 저장된 `sha256` 이 현재 파일 hash 와 다르면 기존 warning 문자열을 stderr 로 emit 한다. 그 외 경우 silent 처리한다.

- **REQ-UN-003 (Event-driven)**: 시스템이 warning 을 emit 한 직후, ack ledger 의 해당 경로 항목을 `{sha256: <current hash>, acknowledged_at: <now>}` 로 갱신 또는 생성해야 한다. 사용자 prompt 는 발생하지 않는다 — 자동 기록이다.

- **REQ-UN-004 (Event-driven)**: 사용자가 reserved 파일 내용을 변경하여 hash 가 ack ledger 의 `sha256` 과 달라졌을 때, REQ-UN-002 에 따라 warning 이 재출현해야 하며 REQ-UN-003 에 따라 ack hash 가 자동 갱신된다.

- **REQ-UN-005 (Where-modality)**: `--verbose` 플래그가 활성화된 경우 REQ-UN-002 의 ack 조회 분기는 우회되어야 하며, 모든 reserved 충돌에 대해 warning 이 emit 된다 (진단 목적). Ack ledger 갱신 (REQ-UN-003) 은 그대로 수행된다.

### C.2 3-way merge fallback advisory 그룹 (Defect #4)

- **REQ-UN-006 (Ubiquitous)**: 시스템은 `.moai/cache/merge-history.json` 을 fallback counter 로 사용해야 한다. 스키마는 `{relPath: string → {fallback_count: int, last_failed_at: ISO-8601 string}}` 매핑이다. 파일이 없으면 첫 3-way 실패 시 빈 JSON 객체 `{}` 로 생성한다.

- **REQ-UN-007 (Event-driven)**: `mergeYAML3Way` 가 실패하여 2-way fallback 으로 전환할 때, 시스템은 merge-history ledger 의 `fallback_count` 를 1 증가시켜야 한다. `fallback_count` 가 3 미만이면 stderr 로 어떤 출력도 emit 하지 않는다 (silent fallback).

- **REQ-UN-008 (Event-driven)**: `fallback_count >= 3` 이 되는 순간, 시스템은 단 한 줄 advisory 를 stderr 로 emit 해야 한다: `hint: 'moai update -c' to resync templates for <relPath>` (정확한 wording 은 acceptance 의 grep 검증 대상). Advisory 는 같은 `relPath` 에 대해 한 번만 emit 되며 (counter 가 추가 증가해도 재출력 없음), 사용자 명시 액션 (예: `--clean`) 으로 counter 가 reset 될 때까지 silent 상태를 유지한다.

- **REQ-UN-009 (Event-driven)**: `mergeYAML3Way` 가 성공하거나 사용자가 `moai update -c` 등으로 templates 를 재동기화한 경우, 시스템은 해당 `relPath` 의 `fallback_count` 를 0 으로 reset 해야 한다.

- **REQ-UN-010 (Where-modality)**: `--verbose` 플래그가 활성화된 경우 REQ-UN-007 의 silent 분기는 우회되어야 하며, 모든 fallback 발생에 대해 기존 `Warning: 3-way merge failed for <relPath>, falling back to 2-way` 메시지가 emit 된다. Counter 갱신 (REQ-UN-007) 과 advisory threshold (REQ-UN-008) 는 그대로 수행된다.

### C.3 공통 무결성 그룹

- **REQ-UN-011 (Unwanted, fallback safety)**: 시스템은 ack ledger 또는 merge-history 파일이 깨진 JSON 일 때 fail 하지 않아야 한다. JSON parse 실패 시 빈 객체 `{}` 로 재초기화하고 정상 진행한다 (의도된 corruption recovery, R2 mitigation).

---

## D. Acceptance Criteria (binary ACs)

상세 검증 명령은 `./acceptance.md` 를 참조한다. 각 AC 는 정확한 grep · cat · go test 명령으로 binary PASS/FAIL 가능하다.

| AC ID | 한 줄 요약 |
|-------|----------|
| AC-UN-001 | `.moai/state/reserved-acknowledged.json` 스키마 (path → {sha256, acknowledged_at}) 가 첫 update 후 생성된다 |
| AC-UN-002 | 같은 reserved 파일에 대해 `moai update` 2회 연속 실행 시 두 번째 실행에서 reserved warning 이 stderr 에 없다 |
| AC-UN-003 | reserved 파일 내용을 변경한 후 `moai update` 실행 시 warning 이 재출현하고 ack ledger 의 sha256 이 갱신된다 |
| AC-UN-004 | `.moai/cache/merge-history.json` 스키마 (relPath → {fallback_count, last_failed_at}) 가 첫 3-way fallback 후 생성된다 |
| AC-UN-005 | 같은 파일이 3-way fallback 1·2회 발생할 때 stderr 에 fallback 관련 출력이 없다 |
| AC-UN-006 | 3회 연속 fallback 발생 시 정확히 한 줄의 advisory (`hint: 'moai update -c' ...`) 가 stderr 에 emit 된다 |
| AC-UN-007 | `moai update --verbose` 실행 시 REQ-UN-002·007 의 억제 분기가 우회되어 reserved warning 과 3-way warning 이 동시에 출력된다 |
| AC-UN-008 | `mergeYAML3Way` 성공 시 해당 relPath 의 fallback_count 가 0 으로 reset 된다 |
| AC-UN-009 | ack ledger JSON 이 손상되었을 때 update 실행이 success 종료하며 ledger 가 빈 객체로 재초기화된다 |
| AC-UN-010 | `.moai/state/reserved-acknowledged.json` 과 `.moai/cache/merge-history.json` 이 `.gitignore` 에 등재되어 있다 (또는 `.moai/state/`, `.moai/cache/` 디렉토리 패턴으로 포함) |
| AC-UN-011 | `go test -run TestReservedAck ./internal/cli/` 가 PASS — ack 기록·hash drift 재출현·verbose bypass 시나리오 커버 |
| AC-UN-012 | `go test -run TestMergeHistory ./internal/cli/` 가 PASS — counter 증가·threshold advisory·reset·verbose bypass 시나리오 커버 |

총 12 ACs · 모두 binary (PASS / FAIL).

---

## E. Risks + Plan

### E.1 Risks

- **R1 (Medium) — Silent merge fallback 이 실제 회귀를 가린다**: 3-way 실패가 매번 silent 면 새로운 merge 회귀가 도입되어도 사용자가 모를 수 있다. **Mitigation**: 3-strike threshold (REQ-UN-008) 가 누적 실패를 advisory 로 노출. `--verbose` (REQ-UN-010) 가 진단 escape hatch 로 작동. CI 의 spec-lint 또는 quality gate 가 별도로 merge regression 을 감지하는 것이 본 SPEC 범위 밖이지만, advisory 메시지는 사용자에게 `moai update -c` 액션을 안내한다.

- **R2 (Low) — Ack ledger 손상이 새 충돌을 가린다**: `.moai/state/reserved-acknowledged.json` 이 손상되어 모든 항목이 사라지면, 사용자는 새 reserved 충돌을 첫 occurrence 로 인식하지 못할 수 있다. **Mitigation**: JSON parse 실패 시 빈 객체로 재초기화 (REQ-UN-011) → 모든 기존 충돌이 unknown 상태가 되어 다음 update 에서 모두 재경고 emit (REQ-UN-002 분기 자연 작동). 즉 손상은 "지나친 경고" 방향으로 fail-safe.

- **R3 (Medium) — state 파일이 `.gitignore` 누락 시 커밋 오염**: `.moai/state/reserved-acknowledged.json` 과 `.moai/cache/merge-history.json` 은 per-machine state 이므로 git 추적되면 안 된다. **Mitigation**: AC-UN-010 으로 `.gitignore` 등재 검증. plan 단계에서 `internal/template/templates/.gitignore` 와 로컬 `.gitignore` 양쪽 patch 의무.

- **R4 (Low) — Ack ledger 무한 성장**: 사용자 프로젝트가 reserved 충돌 파일을 매우 많이 생성하면 ledger 가 비정상적으로 커질 수 있다. **Mitigation**: 본 SPEC 범위에서 TTL · LRU eviction 을 구현하지 않는다 (Non-goal B.2). known limitation 으로 문서화하며, 실제 ledger 크기 모니터링 결과 1KB 초과 사례가 나오면 follow-up SPEC 으로 처리한다.

- **R5 (Low) — Hash 계산 비용**: 매 `checkReservedCollision` 호출마다 모든 reserved 충돌 파일의 sha256 을 계산하면 large file (예: design/tokens.json 9KB) 의 경우 미미하지만 누적 cost 가 있다. **Mitigation**: 9KB 단발 sha256 은 < 1ms — 본 SPEC 범위에서 cache 도입하지 않는다. 실측 후 결정.

- **R6 (Low) — Concurrent `moai update` ledger race**: 사용자가 동일 프로젝트에 두 개의 `moai update` 를 동시에 실행하면 두 프로세스가 같은 ledger 파일을 동시에 갱신할 수 있다. **Mitigation**: plan.md §3.5 의 atomic `os.Rename` write 패턴 (temp file → rename on same filesystem) 으로 부분 쓰기는 방지된다. 마지막 rename 이 승리하는 시나리오 (lost update) 는 가능하지만, 다음 update 에서 누락된 ack 가 재경고로 노출되어 fail-safe 방향으로 동작한다 (R2 와 동일한 추론). 본 SPEC 범위에서는 file lock 을 도입하지 않으며, 동시 실행은 사용자 책임 영역으로 known limitation 으로 문서화한다.

### E.2 Plan 요약

상세 milestone 분해는 `./plan.md` 를 참조한다. 핵심 milestone:

- **M0**: `.moai/state/reserved-acknowledged.json` · `.moai/cache/merge-history.json` 스키마 정의 + `.gitignore` patch (template + local)
- **M1**: `internal/cli/design_folder.go` ack ledger read/write helper + `checkReservedCollision` 분기 (REQ-UN-001~005)
- **M2**: `internal/cli/update.go` merge-history ledger helper + 3-way fallback 분기 (REQ-UN-006~010)
- **M3**: `--verbose` 플래그 추가 + REQ-UN-005·010 우회 분기 통합
- **M4**: `internal/cli/reserved_ack_test.go` + `internal/cli/merge_history_test.go` 추가
- **M5**: progress.md + AC 검증 + chore (status: draft → implemented, version 0.1.0 → 0.2.0)

Tier S 분류 근거: 3 artifacts (spec · plan · acceptance), 영향 코드 ~2-3 files (`design_folder.go`, `update.go`, optional `cmd/moai/update.go` for `--verbose`), 신규 state file 2개, test file 2개. 총 5-6 file 변경, 약 250-350 LOC 추가.

---

## F. Solution Sketch (informative)

### F.1 Reserved ack ledger 의사코드

```
func loadAckLedger(projectRoot string) map[string]AckEntry {
    path := filepath.Join(projectRoot, ".moai", "state", "reserved-acknowledged.json")
    data, err := os.ReadFile(path)
    if err != nil { return map[string]AckEntry{} }
    var ledger map[string]AckEntry
    if json.Unmarshal(data, &ledger) != nil {
        return map[string]AckEntry{}  // REQ-UN-011 fallback
    }
    return ledger
}

func checkReservedCollision(projectRoot string, errOut io.Writer, strict bool, verbose bool) error {
    ledger := loadAckLedger(projectRoot)
    for _, name := range reservedExact {
        target := filepath.Join(projectRoot, designDir, name)
        if _, err := os.Stat(target); err != nil { continue }
        if strict { /* original hard error path */ }
        currentHash := sha256File(target)
        entry, acknowledged := ledger[name]
        if verbose || !acknowledged || entry.SHA256 != currentHash {
            fmt.Fprintf(errOut, "warning: reserved filename: %s (preserved; rename to use canonical templates)\n", name)
            ledger[name] = AckEntry{SHA256: currentHash, AcknowledgedAt: time.Now().UTC().Format(time.RFC3339)}
        }
    }
    saveAckLedger(projectRoot, ledger)
    return nil
}
```

### F.2 Merge-history 의사코드

```
func updateMergeHistory(projectRoot, relPath string, success, verbose bool, errOut io.Writer) {
    hist := loadMergeHistory(projectRoot)
    entry := hist[relPath]
    if success {
        entry.FallbackCount = 0  // REQ-UN-009
    } else {
        entry.FallbackCount++
        entry.LastFailedAt = time.Now().UTC().Format(time.RFC3339)
        if verbose {
            fmt.Fprintf(errOut, "Warning: 3-way merge failed for %s, falling back to 2-way\n", relPath)
        } else if entry.FallbackCount == 3 {
            fmt.Fprintf(errOut, "hint: 'moai update -c' to resync templates for %s\n", relPath)
        }
        // FallbackCount >= 4 : silent (already advised once)
    }
    hist[relPath] = entry
    saveMergeHistory(projectRoot, hist)
}
```

이 의사코드는 informative 이며 실제 구현 세부는 plan.md M1·M2 milestone 에서 확정한다.

---

End of spec.md
