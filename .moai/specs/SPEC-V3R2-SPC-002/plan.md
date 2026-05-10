# SPEC-V3R2-SPC-002 Implementation Plan

> Implementation plan for **@MX TAG v2 with hook JSON integration and sidecar index**.
> Companion to `spec.md` v0.1.0, `research.md` v0.1.0, `acceptance.md` v0.1.0, `tasks.md` v0.1.0.
> Authored on branch `plan/SPEC-V3R2-SPC-002` (Step 1 plan-in-main; base `origin/main` HEAD `fcb486c87`).
> Run phase will execute on a fresh worktree `feat/SPEC-V3R2-SPC-002` per `.claude/rules/moai/workflow/spec-workflow.md` § SPEC Phase Discipline Step 2.

## HISTORY

| Version | Date | Author | Description |
|---------|------|--------|-------------|
| 0.1.0   | 2026-05-10 | manager-spec (Plan workflow Phase 1B) | Initial implementation plan. Scope: PostToolUse hook integration (G-01/G-02), `/moai mx` flag extensions (G-03), `MOAI_MX_HOOK_SILENT` env handling (G-04), mx.yaml `ignore:` wire-up (G-05), MissingReasonForWarn + DuplicateAnchorID emission tests (G-06/G-07), anchor-audit reporter (G-08). Existing `internal/mx/` package (Wave 3 PR #741 + SPC-004 PR #746) provides 80% of the code surface. |

---

## 1. Plan Overview

### 1.1 Goal restatement

`spec.md` §1 의 핵심 목표를 milestone 분해:

> Layer two additive capabilities on top of the FROZEN inline @MX TAG protocol (mx-tag-protocol.md, CONST-V3R2-003): (1) machine-readable JSON sidecar at `.moai/state/mx-index.json` with `schema_version: 2` and atomic temp+rename writes, (2) PostToolUse hook integration emitting `additionalContext` (human-readable summary) + `hookSpecificOutput.mxTags` (structured Tag array). Inline source-code convention stays FROZEN; sidecar is a view, never the source of truth.

### 1.2 Current State Audit (research.md §2 cross-reference)

`internal/mx/` 패키지는 이미 main 에 머지되어 있다 (research.md [E-02] commit `3f0933550`, [E-03] commit `68795dbe3`). 본 SPEC 의 구현 진척도:

| 영역 | 상태 | Evidence |
|------|------|----------|
| `Tag` struct + 5-kind enum | ✅ 구현 | research [E-04], [E-05] |
| `Sidecar` + `schema_version: 2` | ✅ 구현 | research [E-06] |
| Atomic temp+rename write | ✅ 구현 | research [E-07] |
| 16-언어 comment-prefix lookup | ✅ 구현 | research [E-08], [E-26] |
| AnchorID 중복 추적 (Scanner 내부) | ✅ 구현 | research [E-09] |
| `IsStale()` 7-일 TTL | ✅ 구현 | research [E-10] |
| `Manager.UpdateFile` API | ✅ 구현 | research [E-12] |
| **PostToolUse handler 의 mxTags emission** (G-01) | ❌ 격차 | research [E-13] |
| **`HookSpecificOutput.MxTags` 필드** (G-02) | ❌ 격차 | research [E-14] |
| **`/moai mx` 의 `--full` / `--index-only` / `--json` / `--anchor-audit` flags** (G-03) | ❌ 격차 | research [E-15], [E-22] |
| **`MOAI_MX_HOOK_SILENT` env** (G-04) | ❌ 격차 | (G-01 의존) |
| **mx.yaml `ignore:` wire-up** (G-05) | ❌ 격차 | research [E-16] |
| **MissingReasonForWarn 명시 fixture** (G-06) | ⚠ 부분 (Scanner 로직 존재, fixture 부재) | spec §5.1 REQ-006 + §5.5 REQ-040 |
| **DuplicateAnchorID write-refuse 명시 검증** (G-07) | ⚠ 부분 | spec §5.3 REQ-021 |
| **anchor-audit fan_in < 3 reporter** (G-08) | ❌ 격차 | research [E-15] |

Net actionable scope (Run-phase): **8 격차 해소** + 정합성 검증 (template parity, race detector, coverage ≥ 85%).

### 1.2.1 Acknowledged Discrepancies

본 plan 이 spec.md 와 의도적으로 다르게 처리하는 부분 (research evidence 기반):

- **Spec §1 은 본 SPEC 이 "introduces" 하는 것처럼 서술하나, 실제로는 80% 가 이미 main 에 존재** — Wave 3 PR #741 (`3f0933550`, 2026-04-XX) 가 spec.md 의 §2.1 "In Scope" 첫 4개 bullet 을 모두 구현. SPC-004 PR #746 가 resolver / fan_in / danger_category / spec_association 을 추가. 본 plan 의 milestone 은 따라서 "build" 가 아니라 **"integrate + close gaps"** 로 reframe 됨. Sync-phase HISTORY 에서 spec.md §1 "introduces" 표현을 "completes" 로 reconcile 권장.
- **spec §5.1 REQ-006 vs §5.5 REQ-040 의 의도적 분리** — 두 REQ 모두 같은 코드 경로 (Scanner 의 WARN 검출 후 3-라인 lookahead) 로 충족된다. tasks.md 는 단일 task 로 묶음.
- **`HookSpecificOutput.MxTags` 필드 vs `additionalContext` 임베드** — research §7 OQ-1 결정에 따라 `MxTags` 필드 신규 추가. `additionalContext` 는 human-readable summary 만 담음. spec §5.2 REQ-011 verbatim 과 일치.
- **Schema version 변경 금지** — SPC-004 가 이미 sidecar schema 를 소비 중 (PR #746 머지). 본 SPEC 의 모든 변경은 schema_version: 2 backward-compatible 이어야 한다. 새 필드는 omitempty 로 추가, 기존 필드 의미 변경 금지.
- **`internal/hook/file_changed.go:supportedExtensions` 와 `internal/mx/comment_prefixes.go` 의 drift 통합** — 본 plan 은 이를 advisory 격차 (`G-extra`) 로 분류하며 sync 통합은 별도 SPEC (혹은 WF-005 16-language enum SPEC) 로 위임. tasks.md 에서는 drift detection 테스트만 추가 (T-SPC002-15).

### 1.3 Implementation Methodology

`.moai/config/sections/quality.yaml`: `development_mode: tdd`. Run phase MUST follow RED → GREEN → REFACTOR per `.claude/rules/moai/workflow/spec-workflow.md` § Run Phase.

- **RED**: 8개 격차 영역마다 새 RED test fixture 추가:
  - `TestPostToolUseHandler_EmitsMxTags` (G-01) — synthetic Edit on a Go file with new `@MX:NOTE`; expects HookOutput with `MxTags` populated.
  - `TestHookSpecificOutput_MxTagsField` (G-02) — JSON marshalling test that `mxTags` appears in JSON when populated.
  - `TestMxCmd_FullFlag_RebuildsSidecar` (G-03/full) — end-to-end CLI test.
  - `TestMxCmd_IndexOnlyFlag` (G-03/index-only) — silent mode CLI test.
  - `TestMxCmd_JsonFlag_DumpsSidecar` (G-03/json) — stdout JSON dump test.
  - `TestMxCmd_AnchorAuditFlag_ReportsLowFanIn` (G-08) — audit reporter test.
  - `TestPostToolUseHandler_HookSilentEnv` (G-04) — env var disables additionalContext.
  - `TestScanner_RespectsMxYamlIgnore` (G-05) — ignore patterns wire-up.
  - `TestScanner_MissingReasonForWarn_3LineLookahead` (G-06) — explicit fixture.
  - `TestScanner_DuplicateAnchorID_RefuseWrite` (G-07) — explicit fixture.
  - `TestPostToolUseHandler_HookSpecificOutputMismatch` (REQ-041) — wrong hookEventName rejection.
  - `TestSidecar_AtomicWrite_NoPartialReads` (REQ-004 강화) — concurrent read during write fixture.
  - `TestSidecar_StaleArchiveSweep_8Days` (REQ-020) — explicit 8-day fixture.
  - `TestComment_AllSixteenLanguages` (REQ-005, AC-15) — single fixture per language.

- **GREEN**: 각 RED test 를 GREEN 으로 전환하기 위한 production code 추가/수정:
  - `internal/hook/post_tool_mx.go` 신규: PostToolUse 핸들러의 MX scan + sidecar update + mxTags emission.
  - `internal/hook/types.go` 수정: `HookSpecificOutput` 에 `MxTags []mx.Tag` 필드 추가 (omitempty).
  - `internal/cli/mx.go` 확장: 4개 flag (`--full`, `--index-only`, `--json`, `--anchor-audit`) 추가 + RunE 분기.
  - `internal/mx/scanner.go` 보강: mx.yaml ignore 패턴 wire-up (`SetIgnorePatterns` 호출 site 추가).
  - `internal/mx/sidecar.go` 보강: archive sweep 로직 호출 (`--full` path 에서 trigger).
  - `internal/mx/anchor_audit.go` 신규: low-value anchor 리포터.

- **REFACTOR**: 공유 logic 추출:
  - `internal/mx/config.go` 신규: mx.yaml load + 기본값 (ignore patterns, hook.max_additional_context_bytes).
  - PostToolUse handler 와 FileChanged handler 의 공통 scan-then-update 경로를 `internal/mx/integration.go` 의 helper 로 통합.
  - `@MX:NOTE` 태그로 결정 사유 명시.

### 1.4 Deliverables

| Deliverable | Path | REQ Coverage |
|-------------|------|--------------|
| PostToolUse MX handler | `internal/hook/post_tool_mx.go` (신규 ~150 LOC) | REQ-010, REQ-011, REQ-012, REQ-031, REQ-041 |
| `HookSpecificOutput.MxTags` field | `internal/hook/types.go` (+5 LOC) | REQ-011 |
| `/moai mx` flag extensions | `internal/cli/mx.go` (현 30 LOC → ~150 LOC) | REQ-013, REQ-030, REQ-032, REQ-042 |
| mx.yaml config loader | `internal/mx/config.go` (신규 ~80 LOC) | REQ-030, REQ-031 |
| Anchor audit reporter | `internal/mx/anchor_audit.go` (신규 ~80 LOC) | REQ-042 |
| Scan-update integration helper | `internal/mx/integration.go` (신규 ~60 LOC) | REQ-010, REQ-012 |
| Archive sweep on `--full` | `internal/mx/sidecar.go` 확장 (+30 LOC) | REQ-014, REQ-020 |
| 14 신규 RED tests | `internal/hook/post_tool_mx_test.go`, `internal/cli/mx_test.go`, `internal/mx/scanner_test.go`, `internal/mx/sidecar_test.go`, `internal/mx/anchor_audit_test.go` (+~600 LOC) | T-SPC002-01..14 |
| Template parity for hook + CLI changes | `internal/template/templates/.claude/hooks/moai/handle-post-tool.sh` 검증 (likely no change) + `internal/template/embedded.go` regenerate via `make build` | TRUST 5 Trackable |
| CHANGELOG entry | `CHANGELOG.md` Unreleased | TRUST 5 Trackable |
| MX tags per §6 | 6 files (per §6 below) | mx_plan |

Embedded-template parity는 **applicable** (`make build` 후 `internal/template/embedded.go` 재생성). 그러나 본 SPEC 의 변경은 모두 `internal/` Go 코드이고 template 자산은 거의 변경되지 않음 (handle-post-tool.sh wrapper 는 `moai hook post-tool` 호출로 binary 위임, 실제 로직은 Go 측). 검증: `diff -r .claude/ internal/template/templates/.claude/` 변경 0 line 기대.

### 1.5 Traceability Matrix (REQ → AC → Task)

Per plan-auditor PASS criterion #2 (every REQ maps to at least one AC and at least one Task). Built **after** tasks.md was finalized; each row references actual T-SPC002-NN IDs.

| REQ ID | Category | Mapped AC(s) | Mapped Task(s) |
|--------|----------|--------------|----------------|
| REQ-SPC-002-001 | Ubiquitous (inline syntax FROZEN preserved) | (no AC; invariant tested via existing SPC-004 tests + T-SPC002-15 16-lang fixture) | T-SPC002-15 16-language fixture (verifies all 5 kinds across 16 langs) |
| REQ-SPC-002-002 | Ubiquitous (Tag struct definition) | AC-01 | T-SPC002-15 (existing tag.go covers; new fixture exercises) |
| REQ-SPC-002-003 | Ubiquitous (sidecar at `.moai/state/mx-index.json`) | AC-02 | T-SPC002-04 (mx --full path), T-SPC002-12 (atomic write fixture) |
| REQ-SPC-002-004 | Ubiquitous (atomic temp+rename) | AC-03 | T-SPC002-12 (concurrent read fixture) |
| REQ-SPC-002-005 | Ubiquitous (16-language scanner support) | AC-15 | T-SPC002-15 (per-lang fixture) |
| REQ-SPC-002-006 | Ubiquitous (`@MX:WARN` requires sibling `@MX:REASON`) | AC-05 | T-SPC002-09 (MissingReasonForWarn fixture) |
| REQ-SPC-002-007 | Ubiquitous (AnchorID uniqueness) | AC-06 | T-SPC002-10 (DuplicateAnchorID fixture) |
| REQ-SPC-002-008 | Ubiquitous (`schema_version: 2`) | AC-02 | T-SPC002-04 (CLI full path), T-SPC002-12 |
| REQ-SPC-002-010 | Event-driven (PostToolUse re-scan + delta detect) | AC-04 | T-SPC002-01 (PostToolUse handler) |
| REQ-SPC-002-011 | Event-driven (HookResponse with mxTags + additionalContext) | AC-04 | T-SPC002-01, T-SPC002-02 (MxTags field), T-SPC002-03 (handler emission) |
| REQ-SPC-002-012 | Event-driven (sidecar atomic update on PostToolUse) | AC-04 | T-SPC002-01, T-SPC002-12 |
| REQ-SPC-002-013 | Event-driven (`/moai mx --full` rescan) | AC-02 | T-SPC002-04 (--full flag) |
| REQ-SPC-002-014 | Event-driven (LastSeenAt preservation when not found within 7d) | AC-07 | T-SPC002-13 (stale not yet archived fixture) |
| REQ-SPC-002-020 | State-driven (8-day stale → archive) | AC-08 | T-SPC002-14 (archive sweep fixture) |
| REQ-SPC-002-021 | State-driven (DuplicateAnchorID refuse write) | AC-06 | T-SPC002-10 |
| REQ-SPC-002-022 | State-driven (corrupt sidecar repair suggestion) | AC-09 | T-SPC002-11 (corrupt sidecar fixture) |
| REQ-SPC-002-030 | Optional (mx.yaml ignore patterns) | AC-10 | T-SPC002-08 (Scanner ignore wire-up) |
| REQ-SPC-002-031 | Optional (`MOAI_MX_HOOK_SILENT` env) | AC-11 | T-SPC002-07 (silent env fixture) |
| REQ-SPC-002-032 | Optional (`/moai mx --json` stdout dump) | AC-12 | T-SPC002-05 (--json flag) |
| REQ-SPC-002-040 | Complex (3-line lookahead for REASON) | AC-05 | T-SPC002-09 |
| REQ-SPC-002-041 | Complex (HookSpecificOutputMismatch on wrong hookEventName) | AC-13 | T-SPC002-03 (handler validates hookEventName) |
| REQ-SPC-002-042 | Complex (anchor-audit low fan_in candidate) | AC-14 | T-SPC002-06 (--anchor-audit flag) |

Coverage: **22 unique REQs (001..008, 010..014, 020..022, 030..032, 040..042) → 15 ACs (AC-01..AC-15) → 16 tasks (T-SPC002-01..16)**. AC-01 maps to existing scanner code; verification via T-SPC002-15 fixture.

→ All REQ IDs from spec §5 are mapped. AC-01..AC-15 from spec §6 are mapped.

---

## 2. Milestone Breakdown (M1-M6)

각 milestone 은 **priority-ordered** (no time estimates per `.claude/rules/moai/core/agent-common-protocol.md` § Time Estimation HARD rule).

### M1: PostToolUse handler 신설 + HookSpecificOutput 확장 (G-01/G-02) — Priority P0

Owner role: `expert-backend` (Go hook system + JSON protocol).

Tasks:
- T-SPC002-01: 신규 `internal/hook/post_tool_mx.go` 생성. 핸들러 인터페이스: PostToolUse event 의 `tool_name in {Write, Edit}` AND `tool_input.file_path` 가 `comment_prefixes.go` 의 supported ext 인 경우, `mx.NewScanner().ScanFile(filePath)` 호출 → `Manager.UpdateFile(filePath, newTags)` 호출 → 결과로 `HookOutput` 구성.
- T-SPC002-02: `internal/hook/types.go` 수정. `HookSpecificOutput` 구조체에 `MxTags []mx.Tag` field 추가 (`json:"mxTags,omitempty"`). 기존 PostToolUse 출력과 backward-compatible (omitempty).
- T-SPC002-03: PostToolUse handler 의 `additionalContext` 구성. Format: `"@MX:WARN at file.go:42: goroutine leak (no Done() signal)"` (한 줄당 한 태그). spec §1 의 예시 verbatim.
- T-SPC002-15: 16-언어 fixture — `internal/mx/scanner_test.go` 에 `TestScanner_AllSixteenLanguages` 추가. fixture: 16개 임시 파일 (각 언어별 1개 `@MX:NOTE`). 기대: 16개 Tag 모두 추출.

RED 검증: `TestPostToolUseHandler_EmitsMxTags` + `TestHookSpecificOutput_MxTagsField` FAIL on `go test ./internal/hook/ ./internal/mx/` (handler 미구현 + field 부재).

GREEN 검증: 두 test PASS. existing PostToolUse tests (post_tool_astgrep_test.go 등) still PASS — backward compatibility.

### M2: `/moai mx` flag 확장 (G-03) — Priority P0

Owner role: `expert-backend` (CLI subcommand).

Tasks:
- T-SPC002-04: `internal/cli/mx.go` 확장. parent command 의 `RunE` 를 4-flag dispatcher 로 전환:
  - `--full`: 전체 rescan + sidecar rebuild + archive sweep + console summary.
  - `--index-only`: full path 와 동일 + (no console output) — CI 친화.
  - `--json`: 현 sidecar 를 stdout 으로 dump.
  - `--anchor-audit`: fan_in < 3 anchor 리포트 (M5 에서 wire).
  - 충돌 검증: 두 flag 동시 지정 시 error.
- T-SPC002-05: `--json` 구현. `internal/mx/sidecar.go` 의 `Manager.Load()` 호출 → `json.MarshalIndent` → `os.Stdout.Write`.
- T-SPC002-06: `--anchor-audit` 구현 placeholder — M5 에서 본격 wire. 임시 출력 "audit not yet wired" 로 RED test 통과.

RED 검증: 4개 RunE 분기 test 모두 FAIL.

GREEN 검증: 4개 모두 PASS. `moai mx query` (SPC-004) subcommand still works (regression).

### M3: PostToolUse silent mode + mx.yaml ignore wire-up (G-04/G-05) — Priority P1

Owner role: `expert-backend` + `expert-devops` (env var convention).

Tasks:
- T-SPC002-07: PostToolUse handler 가 `os.Getenv("MOAI_MX_HOOK_SILENT") == "1"` 검사. true 면 sidecar 갱신은 수행하되 `additionalContext` 는 비움 (`""`). `MxTags` 필드는 그대로 emit (구조화 데이터는 silent 와 무관 — CI 가 읽을 수 있도록).
- T-SPC002-08: `internal/mx/config.go` 신규. mx.yaml 의 `ignore:` 키 (없으면 기본값) load. `SetIgnorePatterns` 를 PostToolUse handler + CLI `--full` path 모두에서 호출.

RED 검증: `TestPostToolUseHandler_HookSilentEnv` + `TestScanner_RespectsMxYamlIgnore` FAIL.

GREEN 검증: 두 test PASS. CI 환경 (`MOAI_MX_HOOK_SILENT=1` 가 settings.json 에 미리 주입된) 에서 model-turn 토큰 절약 실증.

### M4: Scanner 검증 fixture (G-06/G-07/REQ-040/REQ-041/REQ-022) — Priority P1

Owner role: `expert-backend`.

Tasks:
- T-SPC002-09: `TestScanner_MissingReasonForWarn_3LineLookahead` 추가. fixture: Go 파일 with `// @MX:WARN bug here` 후 다음 3-라인 안에 `@MX:REASON` 없음. 기대: Scanner.warnings 에 1건 (file:line + 메시지 "MissingReasonForWarn").
- T-SPC002-10: `TestScanner_DuplicateAnchorID_RefuseWrite` 추가. fixture: 두 파일에 `// @MX:ANCHOR auth-handler-v1` 동시 존재. 기대: ScanDirectory 가 `DuplicateAnchorID` 에러 반환, Manager.Write 거부.
- T-SPC002-11: `TestSidecar_CorruptJSON_RepairSuggestion` 추가. fixture: `.moai/state/mx-index.json` 에 깨진 JSON 작성 후 `Manager.Load()` 호출. 기대: 빈 sidecar 반환 + stderr 에 repair suggestion (`"WARNING: sidecar corrupt"` 포함).
- T-SPC002-16: `TestPostToolUseHandler_HookSpecificOutputMismatch` 추가. fixture: handler 가 PreToolUse event 수신 시 mxTags emission 거부. 기대: error `HookSpecificOutputMismatch`.

RED 검증: 4 tests FAIL.

GREEN 검증: 4 PASS.

### M5: Archive sweep + Anchor audit (G-08/REQ-014/REQ-020) — Priority P1

Owner role: `expert-backend`.

Tasks:
- T-SPC002-12: `TestSidecar_AtomicWrite_NoPartialReads` 추가. fixture: writer goroutine 이 sidecar 에 큰 페이로드 write 하는 동안 reader goroutine 들이 반복 read. 기대: 모든 read 가 valid JSON (parse 성공) 또는 file-not-exist (writer 가 temp 단계).
- T-SPC002-13: `TestSidecar_StaleNotYetArchived_7Days` 추가. fixture: tag with `LastSeenAt: now - 6days`, scan 결과 not found. 기대: tag 가 sidecar 에 보존됨.
- T-SPC002-14: `TestSidecar_StaleArchived_8Days` 추가. fixture: tag with `LastSeenAt: now - 8days`, scan 결과 not found. 기대: tag 가 sidecar 에서 제거 + `mx-archive.json` 에 추가.
- T-SPC002-06 (continued): `--anchor-audit` 본격 wire. `internal/mx/anchor_audit.go` 신규. logic: `Manager.GetAllTags()` → MXAnchor only filter → `fanin.Count(anchorID)` (SPC-004 의 fanin.go 재사용) → fan_in < 3 만 수집 → markdown 표 stdout 출력.

RED 검증: 3 sidecar test + anchor-audit test FAIL.

GREEN 검증: 4 PASS.

### M6: Verification gates + audit consolidation — Priority P0

Owner role: `manager-cycle` + `manager-quality`.

Tasks:
- T-SPC002-17: 전체 테스트 스위트 — `go test -race -count=1 ./...` PASS, 0 회귀.
- T-SPC002-18: `golangci-lint run` clean.
- T-SPC002-19: `make build` exits 0; `internal/template/embedded.go` 재생성.
- T-SPC002-20: `CHANGELOG.md` Unreleased 업데이트:
  - "feat(mx/SPEC-V3R2-SPC-002): PostToolUse hook MX tag integration with structured mxTags emission"
  - "feat(mx/SPEC-V3R2-SPC-002): /moai mx --full / --index-only / --json / --anchor-audit flags"
  - "feat(mx/SPEC-V3R2-SPC-002): mx.yaml ignore patterns wire-up + MOAI_MX_HOOK_SILENT env handling"
  - "feat(mx/SPEC-V3R2-SPC-002): 7-day stale tag archive sweep on full scan"
- T-SPC002-21: @MX 태그 적용 per §6.
- T-SPC002-22: 수동 검증 — 임시 Go file 에 `@MX:NOTE` 추가 후 `claude` 세션에서 Edit 호출, settings.json 의 PostToolUse hook 가 실제로 mxTags 를 emit 하는지 stdout 캡처.

Verification gate: All AC-SPC-002-01 through AC-SPC-002-15 verified per acceptance.md.

---

## 3. File-Level Modification Map

### 3.1 Files modified (existing)

| File | Lines added/changed | Purpose |
|------|---------------------|---------|
| `internal/hook/types.go` | +5 LOC | `HookSpecificOutput.MxTags` field 추가 |
| `internal/cli/mx.go` | +120 LOC | 4-flag dispatcher + RunE 분기 |
| `internal/mx/scanner.go` | +20 LOC | `MissingReasonForWarn` warning emission 보강 (이미 일부 존재) |
| `internal/mx/sidecar.go` | +30 LOC | Archive sweep on full-scan path |
| `CHANGELOG.md` | +4 lines | Unreleased entry |

### 3.2 Files created (new)

| File | LOC (approx) | Purpose |
|------|--------------|---------|
| `internal/hook/post_tool_mx.go` | ~150 | PostToolUse MX 핸들러 |
| `internal/hook/post_tool_mx_test.go` | ~250 | 4 RED tests for handler (G-01/G-02/G-04/REQ-041) |
| `internal/mx/config.go` | ~80 | mx.yaml load + 기본값 |
| `internal/mx/integration.go` | ~60 | Scan-update helper (PostToolUse + FileChanged 공유) |
| `internal/mx/anchor_audit.go` | ~80 | Low fan_in anchor reporter |
| `internal/mx/anchor_audit_test.go` | ~80 | T-SPC002-06 fixture |
| `internal/cli/mx_test.go` | ~150 | 4 flag tests (T-SPC002-04..06) |
| `internal/mx/scanner_test_g06_g07.go` (or extend `scanner_test.go`) | ~120 | T-SPC002-09, T-SPC002-10, T-SPC002-15 |
| `internal/mx/sidecar_test_g08.go` (or extend `sidecar_test.go`) | ~150 | T-SPC002-11, T-SPC002-12, T-SPC002-13, T-SPC002-14 |

### 3.3 Files removed

None.

### 3.4 Files NOT modified (out-of-scope)

- `.claude/rules/moai/workflow/mx-tag-protocol.md` — FROZEN per CONST-V3R2-003.
- `internal/mx/tag.go` — already complete (Wave 3).
- `internal/mx/comment_prefixes.go` — 16-언어 매핑 이미 완전.
- `internal/mx/resolver_query.go` / `fanin.go` / `danger_category.go` / `spec_association.go` — SPC-004 territory.
- `internal/hook/file_changed.go` — 별도 path; FileChanged event 는 본 SPEC scope 아님 (PostToolUse 에 집중). 단 helper 함수 (`internal/mx/integration.go`) 를 통한 공유는 acceptable.
- `.claude/hooks/moai/handle-post-tool.sh` — 단순 wrapper; binary 위임 구조 변경 없음.

---

## 4. Technical Approach

### 4.1 PostToolUse handler 의 dispatch 로직

```go
// internal/hook/post_tool_mx.go (신규)
func (h *postToolMxHandler) Handle(ctx context.Context, input *HookInput) (*HookOutput, error) {
    // 1. Filter: only Write/Edit tools matter
    if input.ToolName != "Write" && input.ToolName != "Edit" {
        return &HookOutput{}, nil
    }

    // 2. Filter: only supported file extensions
    filePath := extractFilePath(input.ToolInput)
    ext := strings.ToLower(filepath.Ext(filePath))
    if mx.GetCommentPrefix(ext) == "" {
        return &HookOutput{}, nil
    }

    // 3. Scan file for current tags
    scanner := mx.NewScanner()
    cfg, _ := mx.LoadConfig(input.Cwd) // mx.yaml ignore patterns
    scanner.SetIgnorePatterns(cfg.IgnorePatterns)
    tags, err := scanner.ScanFile(filePath)
    if err != nil {
        return &HookOutput{}, nil // graceful degradation
    }

    // 4. Update sidecar
    stateDir := filepath.Join(input.Cwd, ".moai/state")
    mgr := mx.NewManager(stateDir)
    if _, err := mgr.UpdateFile(filePath, tags); err != nil {
        slog.Warn("sidecar update failed", "err", err)
        return &HookOutput{}, nil
    }

    // 5. Build response — silent mode handling
    silent := os.Getenv("MOAI_MX_HOOK_SILENT") == "1"
    var additionalContext string
    if !silent {
        additionalContext = formatTagsForContext(tags)
    }

    return &HookOutput{
        HookSpecificOutput: &HookSpecificOutput{
            HookEventName:     "PostToolUse",
            AdditionalContext: additionalContext,
            MxTags:            tags, // structured field — always populated
        },
    }, nil
}
```

### 4.2 `HookSpecificOutput.MxTags` 필드 위치

```go
// internal/hook/types.go (수정)
type HookSpecificOutput struct {
    HookEventName            string          `json:"hookEventName,omitempty"`
    PermissionDecision       string          `json:"permissionDecision,omitempty"`
    PermissionDecisionReason string          `json:"permissionDecisionReason,omitempty"`
    AdditionalContext        string          `json:"additionalContext,omitempty"`
    SessionTitle             string          `json:"sessionTitle,omitempty"`
    UpdatedInput             json.RawMessage `json:"updatedInput,omitempty"`
    UpdatedMCPToolOutput     string          `json:"updatedMCPToolOutput,omitempty"`
    UpdatedToolOutput        string          `json:"updatedToolOutput,omitempty"`
    MxTags                   []mx.Tag        `json:"mxTags,omitempty"` // SPEC-V3R2-SPC-002 신규
}
```

Cyclic import 회피: `internal/hook` 가 `internal/mx` 를 import (`internal/mx` 가 `internal/hook` 을 import 하지 않음). 현재 `internal/hook/file_changed.go` 가 이미 `internal/mx` import 중이므로 cycle 없음.

### 4.3 `/moai mx` flag dispatcher

```go
// internal/cli/mx.go (확장)
func newMxCmd() *cobra.Command {
    var (
        fullFlag       bool
        indexOnlyFlag  bool
        jsonFlag       bool
        anchorAuditFlag bool
    )
    cmd := &cobra.Command{
        Use:   "mx",
        Short: "@MX TAG 관리 도구",
        RunE: func(cmd *cobra.Command, args []string) error {
            // 충돌 검증
            count := boolToInt(fullFlag) + boolToInt(indexOnlyFlag) + boolToInt(jsonFlag) + boolToInt(anchorAuditFlag)
            if count > 1 {
                return fmt.Errorf("flags --full, --index-only, --json, --anchor-audit are mutually exclusive")
            }
            switch {
            case fullFlag:
                return runMxFull(cmd, false /* console output */)
            case indexOnlyFlag:
                return runMxFull(cmd, true /* silent */)
            case jsonFlag:
                return runMxJsonDump(cmd)
            case anchorAuditFlag:
                return runMxAnchorAudit(cmd)
            default:
                return cmd.Help()
            }
        },
    }
    cmd.Flags().BoolVar(&fullFlag, "full", false, "전체 rescan + sidecar rebuild + archive sweep")
    cmd.Flags().BoolVar(&indexOnlyFlag, "index-only", false, "전체 rescan (silent — CI 친화)")
    cmd.Flags().BoolVar(&jsonFlag, "json", false, "현 sidecar 를 stdout 으로 dump")
    cmd.Flags().BoolVar(&anchorAuditFlag, "anchor-audit", false, "fan_in < 3 anchor 리포트")
    cmd.AddCommand(newMxQueryCmd()) // SPC-004 — 보존
    return cmd
}
```

### 4.4 mx.yaml config loader

```go
// internal/mx/config.go (신규)
type Config struct {
    IgnorePatterns                []string
    HookMaxAdditionalContextBytes int
}

func LoadConfig(projectRoot string) (*Config, error) {
    path := filepath.Join(projectRoot, ".moai/config/sections/mx.yaml")
    data, err := os.ReadFile(path)
    if err != nil {
        if os.IsNotExist(err) {
            return defaultConfig(), nil
        }
        return nil, err
    }
    var raw map[string]interface{}
    if err := yaml.Unmarshal(data, &raw); err != nil {
        return defaultConfig(), nil // graceful
    }
    // Extract ignore: patterns + hook.max_additional_context_bytes
    return parseConfig(raw), nil
}

func defaultConfig() *Config {
    return &Config{
        IgnorePatterns: []string{
            "vendor/**", "node_modules/**", "dist/**", ".git/**",
            "**/*_generated.go", "**/mock_*.go",
        },
        HookMaxAdditionalContextBytes: 4096,
    }
}
```

### 4.5 Anchor audit reporter

```go
// internal/mx/anchor_audit.go (신규)
type AnchorAuditEntry struct {
    AnchorID string
    File     string
    Line     int
    FanIn    int
}

func RunAnchorAudit(mgr *Manager, threshold int) ([]AnchorAuditEntry, error) {
    sidecar, err := mgr.Load()
    if err != nil {
        return nil, err
    }
    var lowValue []AnchorAuditEntry
    for _, tag := range sidecar.Tags {
        if tag.Kind != MXAnchor {
            continue
        }
        fanIn := CountFanIn(tag.AnchorID, sidecar.Tags) // SPC-004 fanin.go reuse
        if fanIn < threshold {
            lowValue = append(lowValue, AnchorAuditEntry{
                AnchorID: tag.AnchorID, File: tag.File, Line: tag.Line, FanIn: fanIn,
            })
        }
    }
    return lowValue, nil
}
```

### 4.6 Performance budget

Spec §7 의 budget:

| Budget | 충족 전략 |
|--------|-----------|
| Full-scan ≤ 2s for 10K files | Scanner 의 prefix-table approach (no AST). 측정: T-SPC002-12 의 후속으로 benchmark test 추가 (optional) |
| Incremental update ≤ 100ms per file | UpdateFile 은 단일 파일 in-memory map 갱신 + 단일 sidecar write. 5KB sidecar write < 1ms; 100ms 여유 충분 |
| Sidecar size ≤ 5MB for 10K tags | 평균 ~500 bytes/tag (JSON-encoded). 10K tags = 5MB upper bound; 실제 평균 < 300 bytes 이므로 여유 |
| Atomic write no partial reads | T-SPC002-12 fixture |

### 4.7 Backward compatibility

- `HookSpecificOutput.MxTags` 는 omitempty — 기존 PostToolUse 출력 (astgrep, duration 등) 에 영향 없음.
- `/moai mx` parent command 의 default RunE (no flag) 는 Help() 출력으로 유지 — 기존 SPC-004 의 `mx query` subcommand 에 영향 없음.
- Schema version 2 유지 — SPC-004 의 mx_query 가 sidecar 를 read 하는 path 에 break 없음.
- mx.yaml 기존 키 (`comment_syntax`, `discovery`, `auto_tag`) 유지; `ignore:` 는 신규 키.

### 4.8 Cross-platform behavior

- `os.Rename` atomic 보장: POSIX 표준. Windows 도 같은 보장 (NTFS).
- `MOAI_MX_HOOK_SILENT` env: cross-platform.
- `internal/mx/comment_prefixes.go` 의 확장자 매핑 은 case-insensitive (이미 `strings.ToLower(ext)` 처리).

---

## 5. Quality Gates

Per `.moai/config/sections/quality.yaml`:

| Gate | Requirement | Verification command |
|------|-------------|-----------------------|
| Coverage | ≥ 85% per modified file | `go test -cover ./internal/hook/ ./internal/mx/ ./internal/cli/` |
| Race | `go test -race ./...` clean | `go test -race -count=1 ./...` |
| Lint | golangci-lint clean | `golangci-lint run` |
| Build | embedded.go regenerated | `make build` |
| Template parity | `diff -r` byte-identical | `diff -r .claude/ internal/template/templates/.claude/` |
| Sidecar contract | schema_version: 2 보존 | T-SPC002-12 fixture |
| MX | @MX tags applied per §6 | manual (post-implementation review) |

---

## 6. @MX Tag Plan (mx_plan)

Apply per `.claude/rules/moai/workflow/mx-tag-protocol.md`. Language: ko (per `code_comments: ko`). Note: 본 plan 자체는 plan markdown 이므로 @MX 태그를 추가하지 않음 (이는 source code 대상). 아래 표는 run-phase 에서 Go 코드에 추가될 태그.

| File | Tag | Reason |
|------|-----|--------|
| `internal/hook/post_tool_mx.go:Handle` | `@MX:ANCHOR @MX:REASON: SPEC-V3R2-SPC-002 의 PostToolUse MX 통합 진입점; sidecar 갱신 + mxTags emission 의 단일 책임 — fan_in ≥ 3 (PostToolUse hook dispatcher + FileChanged shared path + 향후 SessionEnd MX validation)` | invariant contract + high fan_in |
| `internal/hook/types.go:HookSpecificOutput.MxTags` | `@MX:NOTE: SPEC-V3R2-SPC-002 신규 필드; omitempty 로 기존 출력 backward-compatible; consumers (evaluator-active, ralph engine) 가 직렬화된 Tag 배열을 직접 소비` | new field rationale |
| `internal/mx/config.go:LoadConfig` | `@MX:NOTE: mx.yaml 부재 시 graceful degradation (defaultConfig); CI 환경 isolation 보장 — config 미존재가 lint break 가 되지 않음` | non-obvious behavior |
| `internal/mx/anchor_audit.go:RunAnchorAudit` | `@MX:NOTE: SPC-004 fanin.go 의 CountFanIn 재사용; threshold 3 은 spec §5.5 REQ-042 에서 유래 — hardcoded 가 의도적 (config 의 evolution 표면 분리)` | cross-SPEC dependency note |
| `internal/cli/mx.go:newMxCmd` | `@MX:WARN @MX:REASON: 4-flag mutual exclusion; 사용자가 두 flag 동시 지정 시 silent ambiguity 위험 — 명시적 error 반환으로 회피. 향후 flag 추가 시 카운트 로직 갱신 의무` | mutability hazard |
| `internal/mx/sidecar.go:archiveSweep` (신규 helper) | `@MX:WARN @MX:REASON: archive 후 원본 sidecar 에서 entry 제거하므로 partial failure 위험 — archive write 성공 후에만 sidecar rewrite (two-phase). archive write 실패 시 sidecar 무변경 (idempotent retry 가능)` | failure-recovery semantic |

---

## 7. Risk Mitigation Plan (spec §8 risks → run-phase tasks)

| spec §8 risk | Mitigation in run-phase |
|--------------|--------------------------|
| Row 1 — Sidecar drift from source truth | T-SPC002-04 (`/moai mx --full`) 가 언제든 rebuild; CI guard 잠재 추가 (out-of-scope) |
| Row 2 — Cross-language parser 엣지 케이스 | T-SPC002-15 16-언어 fixture; v3.1 에서 escape-aware refinement (out-of-scope) |
| Row 3 — PostToolUse JSON 토큰 폭증 | mx.yaml `hook.max_additional_context_bytes` (T-SPC002-08); MOAI_MX_HOOK_SILENT (T-SPC002-07) |
| Row 4 — Stale 태그 7-일 TTL 위반 | T-SPC002-13 + T-SPC002-14 fixtures; archive 는 `--full` path 에서 트리거 (research §7 OQ-3) |
| Row 5 — Duplicate AnchorID block | T-SPC002-10 fixture; Scanner 가 명시 에러 + Manager refuse-write |
| Row 6 — Binary size growth | prefix-table approach (no per-language parsers); risk 사실상 해소 |

추가 mitigation:
- **OOQ — `internal/hook/file_changed.go` 와 PostToolUse 의 sidecar race**: sidecar Manager 가 이미 `sync.RWMutex` 보유 (research [E-21]). 추가 작업 불필요.
- **OOQ — handler 가 PostToolUseFailure 에서 동작?**: spec 미명시. plan 결정: PostToolUse 만 처리. PostToolUseFailure 는 별도 핸들러가 처리. `internal/hook/post_tool_mx.go` 의 `EventType()` 는 `EventPostToolUse` 만 반환.

---

## 8. Dependencies (status as of `fcb486c87`)

### 8.1 Blocking (consumed)

- ✅ **SPEC-V3R2-CON-001** (FROZEN/EVOLVABLE codification) — assumed merged; CONST-V3R2-003 가 mx-tag-protocol.md 를 FROZEN 으로 등록.
- ✅ **SPEC-V3R2-RT-001** (JSON hook protocol) — `internal/hook/types.go` 가 RT-001 의 dual-protocol 구현 보유 (research [E-14]).

### 8.2 Blocked by (none active)

All blockers are merged or adjacent. RT-001 의 protocol 표면 변경이 본 SPEC 의 기간 내에 발생하면 plan 재조정 필요.

### 8.3 Blocks (downstream consumers)

- **SPEC-V3R2-SPC-004** (resolver query API) — **이미 머지** (PR #746). 본 SPEC 이 sidecar contract 를 변경하지 않음 (schema_version: 2 유지) 으로 SPC-004 의 read path 에 영향 없음.
- **SPEC-V3R2-HRN-003** (evaluator MX 점수) — downstream; 본 SPEC 의 sidecar 가 publish 되면 score 계산에 사용 가능.
- **SPEC-V3R2-WF-005** (16-언어 enum SoT) — comment_prefixes.go drift detection 은 advisory (out-of-scope).

---

## 9. Verification Plan

### 9.1 Pre-merge verification (run-phase end)

- [ ] All 22 tasks (T-SPC002-01..22) complete per tasks.md (M3 의 13 sub-task 별도 카운트)
- [ ] All 15 ACs (AC-SPC-002-01..15) verified per acceptance.md
- [ ] `go test -race -count=1 ./...` PASS (no regressions)
- [ ] `golangci-lint run` clean
- [ ] `make build` regenerates `internal/template/embedded.go` correctly
- [ ] Coverage `go test -cover ./internal/{hook,mx,cli}/` ≥ 85% per package
- [ ] `diff -r .claude/ internal/template/templates/.claude/` byte-identical (no template change expected)
- [ ] CHANGELOG entry written in Unreleased section
- [ ] @MX tags applied per §6 (6 tags)
- [ ] Manual end-to-end smoke test: edit a Go file via `claude` session, observe `mxTags` in PostToolUse hook stdout

### 9.2 Plan-auditor target

- [ ] All 22 unique REQs mapped in §1.5 traceability matrix
- [ ] All 15 ACs mapped to ≥1 task
- [ ] No orphan tasks (every task supports ≥1 REQ)
- [ ] research.md evidence anchors cited (≥30 per plan-audit mandate)
- [ ] §1.2.1 explicitly addresses spec/plan discrepancies (Wave 3 prior implementation, REQ-006/REQ-040 unification, MxTags vs additionalContext)
- [ ] FROZEN clause CONST-V3R2-003 explicitly preserved (inline syntax untouched)
- [ ] Schema_version 변경 금지 명시
- [ ] Worktree-base alignment per Step 2 (run-phase) called out (§10 below)
- [ ] §6 mx_plan covers ≥3 of {ANCHOR, WARN, NOTE} types (covered: 1 ANCHOR + 2 WARN + 3 NOTE = 모두 3)
- [ ] No time estimates (P0/P1 priority labels only)
- [ ] Parallel SPEC isolation: only `internal/hook/`, `internal/cli/`, `internal/mx/`, `CHANGELOG.md`. .claude/ 트리 미변경.

### 9.3 Plan-auditor risk areas (front-loaded mitigations)

- **Risk: Wave 3 가 이미 80% 구현 → REQ-001..008 이 redundant 처럼 보임** → §1.2.1 acknowledged; sync-phase HISTORY 에서 spec.md "introduces" → "completes" reconcile. plan 의 task 는 격차 8개 (G-01..G-08) 에 집중.
- **Risk: SPC-004 가 이미 main 에 있는데 본 SPEC 의 schema 변경이 break 가능** → schema_version: 2 보존 서약 + 새 필드는 omitempty + T-SPC002-12 atomic write fixture.
- **Risk: PostToolUse hook 토큰 폭증 (model context bloat)** → MOAI_MX_HOOK_SILENT env (T-SPC002-07) + mx.yaml `hook.max_additional_context_bytes` (T-SPC002-08).
- **Risk: PostToolUse handler 가 모든 Write/Edit 에 동작 → CI 환경에서 noise** → MOAI_MX_HOOK_SILENT=1 을 settings.local.json 의 CI variant 에 미리 주입; spec §5.4 REQ-031 verbatim.
- **Risk: PostToolUse handler 가 cyclic deadlock (Manager.UpdateFile mutex 와 PostToolUse 동시 발화)** → sidecar Manager 가 RWMutex 보유; sidecar 는 single-writer model (PostToolUse handler + `--full` CLI 가 동일 mutex 공유).
- **Risk: `fcb486c87` baseline drift if main advances during plan PR review** → run-phase 가 명시적으로 `origin/main` 에서 rebase (Step 2 `moai worktree new --base origin/main`).
- **Risk: `internal/hook/post_tool_mx.go` 와 기존 PostToolUse handler (`post_tool_astgrep`, `post_tool_duration`) 충돌** → Go hook 시스템은 EventType-based dispatch + multiple handlers chain; handler 등록 order 는 settings.json hook config 가 결정. 본 SPEC 의 handler 는 별도 등록 entry (e.g. settings.json `PostToolUse[1]`) — 기존 handler 영향 없음.

### 9.4 Rollback plan (if SPC-004 discovers schema gap)

만약 run-phase 후 SPC-004 가 sidecar schema 의 새로 발견된 gap (예: anchor_id 의 namespace 충돌) 을 보고하면:
1. `MxTags` 필드 emission 만 비활성화 (`HookSpecificOutput.MxTags = nil` 강제) — RT-001 protocol 호환성 유지.
2. Sidecar write 는 계속 (consumer 영향 없음).
3. Hotfix SPEC 으로 schema_version: 3 도입 + migrate.

---

## 10. Run-Phase Entry Conditions

After plan PR squash-merged into main:

1. `git checkout main && git pull` (host checkout).
2. `moai worktree new SPEC-V3R2-SPC-002 --base origin/main` per Step 2 spec-workflow.md.
3. `cd ~/.moai/worktrees/moai-adk/SPEC-V3R2-SPC-002`.
4. `git rev-parse --show-toplevel` should output the worktree path (Block 0 verification per session-handoff.md).
5. `git rev-parse HEAD` should match plan-merge commit SHA on main.
6. Verify `internal/mx/` 패키지 무결성: `go test ./internal/mx/...` — 기존 SPC-004 tests still PASS.
7. `/moai run SPEC-V3R2-SPC-002` invokes Phase 0.5 plan-audit gate, then proceeds to M1.

---

Version: 0.1.0
Status: Plan artifact for SPEC-V3R2-SPC-002
Run-phase methodology: TDD (per `.moai/config/sections/quality.yaml` `development_mode: tdd`)
Estimated artifacts: 5 new Go source files + 5 new test files + 4 modified Go files + CHANGELOG = ~1,250 LOC delta (production ~480 + test ~770)
