# SPEC-V3R4-STATUS-DRIFT-FOLLOWUP-001 — Compact

> Auto-extracted from spec.md (REQ + AC + Exclusions only). Source: spec.md v0.1.0.

## ID & Status

- id: SPEC-V3R4-STATUS-DRIFT-FOLLOWUP-001
- version: 0.1.0
- status: draft
- priority: P1
- labels: v3r4, lint, spec-frontmatter, status-drift, plan-in-main

## Goal (1-line)

LSKC-001 cleanup으로 노출된 64 `StatusGitConsistency` WARN을 frontmatter status 일괄 동기화 + terminal state detector exemption 으로 0건으로 정리. 4-단계 cleanup chain 종결.

---

## Requirements (EARS)

### Ubiquitous

- **REQ-SDF-001**: The system shall preserve all `superseded` and `archived` SPEC frontmatter `status` values without modification (terminal lifecycle state intent preservation).
- **REQ-SDF-002**: The system shall preserve every affected SPEC's body content byte-for-byte, except for one new HISTORY table row.
- **REQ-SDF-003**: The system shall bump each affected SPEC's `version` field by a patch increment and update `updated_at` to the synchronization commit date.

### Event-Driven

- **REQ-SDF-004**: When a SPEC matches Pattern A (`completed → implemented`), the system shall downgrade frontmatter status via bulk script.
- **REQ-SDF-005**: When a SPEC matches Pattern B (`completed → in-progress`), the system shall conduct per-SPEC verification then downgrade or document why retained.
- **REQ-SDF-006**: When a SPEC matches Pattern C (`implemented → in-progress`), same verification flow as REQ-SDF-005.
- **REQ-SDF-007**: When `moai spec lint --strict` runs after synchronization, the system shall report 0 `StatusGitConsistency` warnings.

### State-Driven

- **REQ-SDF-008**: While SPEC frontmatter `status` is `superseded` or `archived`, the detector shall apply terminal-state exemption (no finding).
- **REQ-SDF-009**: While Wave 5 (Pattern H) runs during sync-phase, the system shall recursively re-run the bulk script against the 4 cleanup-chain SPECs.

### Unwanted

- **REQ-SDF-010**: The system shall not introduce any new `lint.skip` entry into any SPEC frontmatter.
- **REQ-SDF-011**: The system shall not modify the walker filter scope in `internal/spec/drift.go::shouldSkipCommitTitle`.
- **REQ-SDF-012**: The system shall not deprecate, disable, or weaken the `StatusGitConsistencyRule` itself.
- **REQ-SDF-013**: The system shall not modify any SPEC body for the 64 affected SPECs (HISTORY 1 row exception).
- **REQ-SDF-014**: The system shall not modify any SPEC outside the 64-affected list.

### Optional

- **REQ-SDF-015**: Where the run-phase agent externalizes bulk logic as `.moai/scripts/status-drift-cleanup.go`, the system shall record idempotency property.
- **REQ-SDF-016**: Where additional cleanup waves are needed, the system shall extend affected-list files without renumbering.

---

## Acceptance Criteria

- **AC-SDF-001**: `moai spec lint --strict 2>&1 | grep -c "StatusGitConsistency"` outputs exactly `0`.
- **AC-SDF-002**: Pattern A 50 SPECs frontmatter `status: implemented`, version patch bump, `updated_at` updated, HISTORY row added.
- **AC-SDF-003**: Pattern B+C 10 SPECs verification results recorded in `run-verification.md` with decision: downgrade or keep+sync-pending.
- **AC-SDF-004**: terminal-state exemption applied in `internal/spec/lint.go::StatusGitConsistencyRule::Check` + 4 unit tests (D/E/F/G cases) PASS.
- **AC-SDF-005**: Pattern H 4 cleanup-chain SPECs (including this SPEC) reconciled during sync-phase.
- **AC-SDF-006**: No new `lint.skip` entry introduced (grep count baseline preserved).
- **AC-SDF-007**: Only 63 SPECs + 1 detector code + 1 test file + 1 bulk script + this SPEC's own artifacts modified (out-of-scope files untouched).
- **AC-SDF-008**: Bulk script second run is no-op (idempotency).

---

## Exclusions (What NOT to Build)

[HARD] 본 SPEC 은 다음 항목을 명시적으로 만들지 않는다:

1. **walker filter expansion** — `chore:`, `ci:`, `docs:`, `style:`, `refactor:`, `perf:`, `test:` 어떤 prefix 도 `shouldSkipCommitTitle` 에 추가하지 않음. (옵션 b 거부)
2. **StatusGitConsistencyRule 비활성화 / deprecation** — rule 자체는 영구 active. (옵션 c 거부)
3. **새 `lint.skip` entry** — 본 SPEC 자체 frontmatter 포함 어떤 SPEC에도 `lint.skip` 추가 0건. (LSKC-001 정신 영구 보존)
4. **detector 의미 광범위 변경** — `getGitImpliedStatus` 동작, `ClassifyPRTitle` prefix 매핑, lifecycle status enum 어느 것도 변경 없음. exemption 만 narrow scope.
5. **64 SPEC body content 변경** — REQ/AC/HISTORY 본문 수정 0줄 (HISTORY 1줄 추가만 허용).
6. **64 외 SPEC 수정** — 다른 SPEC frontmatter / body 0건 변경.
7. **CI workflow 변경** — `.github/workflows/spec-lint.yml` 미수정.
8. **새 lint rule** — `StatusGitConsistencyRule` 외 rule 추가 / 변경 0건.
9. **새 CLI flag** — `moai spec lint` 명령에 새 옵션 추가 0건.
10. **docs-site 4-locale 동기화** — sync-phase 또는 별도 SPEC scope.
11. **AC-SDF-004 외 새 테스트** — terminal state exemption 4 cases 외 추가 테스트 0건 (절제).
12. **frontmatter format / ordering 변경** — baseline key order, 들여쓰기 스타일, quote style 보존 (yaml.Node API 사용 강제).
13. **CHANGELOG entry** — run-phase 미작성, sync-phase 책임.

---

## REQ ↔ AC Matrix

| REQ | Mapped ACs |
|-----|-----------|
| REQ-SDF-001 | AC-SDF-004, AC-SDF-007 |
| REQ-SDF-002 | AC-SDF-002, 003, 005, 007 |
| REQ-SDF-003 | AC-SDF-002 |
| REQ-SDF-004 | AC-SDF-002 |
| REQ-SDF-005 | AC-SDF-003 |
| REQ-SDF-006 | AC-SDF-003 |
| REQ-SDF-007 | AC-SDF-001 (primary) |
| REQ-SDF-008 | AC-SDF-004 |
| REQ-SDF-009 | AC-SDF-005 |
| REQ-SDF-010 | AC-SDF-006 |
| REQ-SDF-011 | AC-SDF-007 |
| REQ-SDF-012 | AC-SDF-007 |
| REQ-SDF-013 | AC-SDF-002, 003, 005 |
| REQ-SDF-014 | AC-SDF-007 |
| REQ-SDF-015 | AC-SDF-008 |
| REQ-SDF-016 | (Optional — no AC) |
