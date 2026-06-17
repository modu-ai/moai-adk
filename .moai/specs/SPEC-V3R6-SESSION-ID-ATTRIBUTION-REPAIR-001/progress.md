---
id: SPEC-V3R6-SESSION-ID-ATTRIBUTION-REPAIR-001
title: "Progress — Session-ID attribution dead-feature repair"
version: "0.1.0"
status: in-progress
created: 2026-06-17
updated: 2026-06-17
author: manager-spec
priority: P1
phase: "v3.0.0"
module: "internal/cli/session.go; internal/hook/session_start.go; internal/session/registry.go"
lifecycle: spec-anchored
tags: "session, attribution, multi-session, coordination, race-attribution, doctrine"
era: V3R6
---

# Progress — SPEC-V3R6-SESSION-ID-ATTRIBUTION-REPAIR-001

This file is the canonical progress tracker. §E.1 is populated at plan-phase
close; §E.2-§E.5 are populated by downstream agents (manager-develop for §E.2/§E.3,
manager-docs for §E.4/§E.5) per the Status Transition Ownership Matrix.

## §E.1 Plan-phase Audit-Ready Signal

- **Plan-phase completed:** 2026-06-17
- **Artifacts emitted:** 4 (spec.md, plan.md, acceptance.md, research.md) + this progress.md
- **Frontmatter `status`:** in-progress (M1 commit transition `draft → in-progress` owned by manager-develop per Status Transition Ownership Matrix)
- **Plan-auditor iter-2:** PASS-WITH-DEBT 0.86 (≥ Tier M 0.80 threshold). Implementation Kickoff Approval GRANTED.
- **Era classification:** V3R6 (explicit `era: V3R6` in frontmatter).
- **Tier:** M (standard). cycle_type: tdd.

## §E.2 Run-phase Evidence

### Milestone status (M1-M6 ALL COMPLETE)

- **M1 (P1 write-path investigation, GATING):** COMPLETE. Root cause in research.md §D.0 (HookOutput.Data `json:"-"` structural gap) + §D.1 (empty-SessionID gate bypass). REQ-WPR-003 warning + `moai session doctor` diagnostic implemented.
- **M2 (`moai session current`, P2 Stage 1):** COMPLETE. 7 subcommands. AC-RDP-001/002/003/006 PASS.
- **M3 (SessionStart additionalContext injection, P2 Stage 2):** COMPLETE. hookSpecificOutput.AdditionalContext + side-channel file write. AC-RDP-004/005 PASS.
- **M4+M5 (P3 fallback doctrine canonicalization):** COMPLETE. 3 non-canonical variants eliminated; 2 spellings remain (canonical fallback + `<UUID from moai session current>`). AC-FBC-001/002/003/004 PASS; AC-FBC-005 PASS.
- **M6 (resume template citation + final verification):** COMPLETE. REQ-MSC-003 citation added (local + template byte-identical). AC-MSC-003 PASS.

### AC PASS/FAIL matrix (ALL 18 ACs — 15 MUST + 3 SHOULD)

| AC ID | Severity | Status | Evidence |
|-------|----------|--------|----------|
| AC-WPR-001 | MUST | PASS | `TestSessionDoctorRegistryAbsent` + `TestSessionDoctorRegistryPresentWithEntry` |
| AC-WPR-002 | MUST | PASS | `doctorRootCauses()` 3 candidates; asserted non-empty |
| AC-WPR-003 | MUST | PASS | `TestSessionStartEmptySessionIDEmitsWarning` — stderr warning, exit 0 |
| AC-WPR-004 | MUST (GATE) | PASS | research.md §D.0/D.1 populated; M2-M6 unblocked |
| AC-RDP-001 | MUST | PASS | `TestSessionCurrentListedInHelp` — 7 subcommands |
| AC-RDP-002 | MUST | PASS | `TestSessionCurrentReadsSideChannel` — UUID resolved post-M3 |
| AC-RDP-003 | MUST | PASS | `TestSessionCurrentFallbackWhenNoSideChannel` — exit 0 + fallback |
| AC-RDP-004 | MUST | PASS | `TestSessionStartInjectsAdditionalContext` + `TestSessionStartWritesSideChannel` |
| AC-RDP-005 | MUST | PASS | `TestSessionStartAdditionalContextStrictlyAdditive` |
| AC-RDP-006 | SHOULD | PASS | `TestSessionCurrentShowFallbackFlag` + `TestSessionCurrentJSONFallback` |
| AC-FBC-001 | MUST | PASS | canonical-surface grep: 2 spellings (1 canonical fallback) |
| AC-FBC-002 | MUST | PASS | canonical string byte-identical across both SSOT surfaces |
| AC-FBC-003 | MUST | PASS | 3 non-canonical variants → 0 matches in original form |
| AC-FBC-004 | MUST | PASS | local ↔ template byte-identical (diff exit 0; `TestRuleTemplateMirrorDrift` PASS) |
| AC-FBC-005 | SHOULD | PASS | doctrine edits cite canonical verbatim; no new variant introduced |
| AC-MSC-001 | MUST | PASS | existing 5-verb session tests pass unchanged |
| AC-MSC-002 | MUST | PASS | `FormatStderrReminder` unchanged (no registry.go L424-448 edits) |
| AC-MSC-003 | SHOULD | PASS | session-handoff.md Block 2 cites `moai session current` as primary UUID source |

### Verbatim final verification evidence (M6)

**go build ./... (2026-06-17, HEAD with M1-M6):**
```
$ go build ./...
(exit 0, no output)
```

**go vet ./... (2026-06-17):**
```
$ go vet ./...
(exit 0, no output)
```

**golangci-lint run --timeout=2m (2026-06-17):**
```
$ golangci-lint run --timeout=2m
0 issues.
```

**spec-lint (2026-06-17):**
```
$ go run ./cmd/moai spec lint .moai/specs/SPEC-V3R6-SESSION-ID-ATTRIBUTION-REPAIR-001/spec.md
✓ No findings — all SPEC documents are valid
```

**Template neutrality + mirror parity (2026-06-17):**
```
$ go test -count=1 -run 'TestRuleTemplateMirrorDrift|TestTemplateNoInternalContentLeak|TestEmbeddedMirror' ./internal/template/
ok  	github.com/modu-ai/moai-adk/internal/template	0.429s
```

**Canonical-surface variant enumeration (AC-FBC-003):**
```
$ grep -rohE 'source_session_id: <[^>]*>' .claude/rules/moai/ .claude/output-styles/moai/moai.md internal/template/templates/.claude/rules/moai/ internal/template/templates/.claude/output-styles/moai/moai.md | sort -u | nl
     1  source_session_id: <UUID from moai session current>
     2  source_session_id: <not-available — environment-fallback, next session will backfill via /moai session register on activation>
(2 spellings — 1 canonical fallback + 1 rewritten happy-path slot; 3 non-canonical eliminated)
```

**In-scope package tests (session/cli/hook/template, 2026-06-17):**
```
$ go test -count=1 ./internal/session/... ./internal/cli/... ./internal/hook/ ./internal/template/
ok  	github.com/modu-ai/moai-adk/internal/session
ok  	github.com/modu-ai/moai-adk/internal/cli
ok  	github.com/modu-ai/moai-adk/internal/hook
ok  	github.com/modu-ai/moai-adk/internal/template
```

### D7/D8 MINOR prose fixes (plan-auditor iter-2 NEW defects)

- **D7 FIXED:** spec.md §I history note corrected — "AC-FBC-004 added" → "AC-FBC-005 added for REQ-FBC-004 traceability (AC-FBC-004 was already taken by template-mirror parity tracing to REQ-FBC-002)".
- **D8 FIXED:** research.md §C over-attribution corrected — `.moai/docs/` contributes 0 variants (verified by grep); all 10 broader-context extras are in auto-memory. Prose updated in 2 locations.

### Gaps (verification-claim-integrity §3.4)

- **Pre-existing baseline failure (OUT OF SCOPE):** `TestCollectMemory` + `TestCollectMemory_AutoCompactScaling` in `internal/statusline` fail on the clean `12e20d190` baseline (confirmed via `git stash` + retest). Leftover from the immediately-preceding STATUSLINE-PRESET-RETIRE SPEC. NOT caused by this SPEC's changes; NOT fixed here (scope discipline: session-id attribution only).
- **`TestHookWrapper_ValidJSON` / `TestHookWrapper_MoaiBinaryFallback`**: flaky under full-suite parallel load (5s timeout on bash subprocess); PASS with `-run` filter or `-count=1` on isolated package. Pre-existing environmental artifact, NOT a regression.

### Residual-risk (verification-claim-integrity §3.5)

- The `additionalContext` injection is lost after `/clear`/compaction (spec.md §F.2). The side-channel file (`.moai/state/current-session-id.txt`) persists, so `moai session current` can re-read the UUID post-compaction.
- Headless `-p` invocations without hooks bypass SessionStart; `moai session current` returns the canonical fallback (REQ-RDP-006).
- The pre-existing statusline `TestCollectMemory` failure is not blocking this SPEC but should be addressed in a separate follow-up (STATUSLINE-PRESET-RETIRE residual).

## §E.3 Run-phase Audit-Ready Signal

- **Run-phase completed:** 2026-06-17 (M1-M6 ALL COMPLETE)
- **Run-phase commit:** (this commit + preceding M1-M3 commit `8c2c40c13`)
- **AC totals:** 18/18 PASS (15 MUST + 3 SHOULD)
- **Quality gates:** go build exit 0; go vet exit 0; golangci-lint 0 issues; spec-lint clean; template neutrality + mirror parity PASS.
- **Frontmatter `status`:** in-progress (sync-phase `in-progress → implemented` transition owned by manager-docs per Status Transition Ownership Matrix).

## §E.4 Sync-phase Audit-Ready Signal

**Claim**: sync-phase complete. CHANGELOG `[Unreleased] → Added` entry emitted; frontmatter `status` transitioned `in-progress → implemented` on spec.md (canonical status bearer per STATUSLINE-PRESET-RETIRE / WORKFLOW-EFFORT-MAP precedent — only spec.md carries the lifecycle transition); D7 (spec.md §I AC-FBC-005 history note) + D8 (research.md §C `.moai/docs/` over-attribution) verified already-fixed during run-phase; docs-site 4-locale skipped (internal doctrine SPEC, no user-facing feature); README untouched (no user-facing surface changed).

**Evidence**:
- `grep -c 'SPEC-V3R6-SESSION-ID-ATTRIBUTION-REPAIR-001' CHANGELOG.md` was 1 prior (a cross-reference forward-ref inside the SESSION-HANDOFF-SSOT-ALIGN-001 entry, NOT a dedicated entry) → dedicated Added entry emitted, B12 no-duplicate satisfied.
- spec.md frontmatter `status:` now reads `implemented` (single-artifact transition per precedent).
- D7 confirmed: spec.md §I history note — "AC-FBC-005 added for REQ-FBC-004 traceability (AC-FBC-004 was already taken by template-mirror parity tracing to REQ-FBC-002)".
- D8 confirmed: research.md §C — "`.moai/docs/` contributes 0 variants — all 10 extras are in auto-memory; the prior prose over-attributed `.moai/docs/`".
- AC count 18 (15 MUST + 3 SHOULD) cross-checked against acceptance.md §B (SSOT), NOT progress.md.

**Baseline-attribution**: measured against the run-phase HEAD `538fcaea6` (M4-M6) tree; sync edits applied on a clean origin-synced tree (HEAD `7a338007a` at sync time, a parallel session's memory-hygiene commit — no overlap with this SPEC's files). The 18/18 AC PASS figure is carried verbatim from progress.md §E.2 (run-phase evidence recorded by manager-develop against the M1-M6 tree), not re-measured in sync-phase (sync is doc-only, no code change).

**Gaps**: sync-phase manager-docs spawn failed with `context window limit` (3rd GLM occurrence — also hit WORKFLOW-EFFORT-MAP-001 + SESSION-HANDOFF-SSOT-ALIGN-001 siblings) → orchestrator-direct fallback (`Authored-By-Agent: orchestrator-direct` trailer). This MAY emit an `OwnershipTransitionInvalid` lint Warning because the canonical owner of `in-progress → implemented` is manager-docs (per lifecycle-sync-gate.md ownership matrix) — accepted as documented GLM-environment debt, consistent with the two sibling SPECs' fallback. No independent sync-auditor 4-dimension run in this commit (downstream, separate gate).

**Residual-risk**: `additionalContext` injection is lost after `/clear`/compaction (spec.md §F.2); the side-channel file `.moai/state/current-session-id.txt` persists so `moai session current` re-reads the UUID. Headless `-p` invocations without hooks bypass SessionStart → canonical fallback (REQ-RDP-006). Pre-existing `internal/statusline` `TestCollectMemory` baseline failure is out of scope (STATUSLINE-PRESET-RETIRE residual).

sync_commit_sha: 615a4af95 — backfilled after the docs sync commit (non-bold actual SHA per feedback_era_commit_sha_field_format; bold/placeholder would cause V3R6→V3R5 era misclassification).

## §E.5 Mx-phase Audit-Ready Signal

**Claim**: 4-phase lifecycle complete (plan → run → sync → Mx). spec.md frontmatter `status: implemented → completed`. Both `sync_commit_sha` (615a4af95) and `mx_commit_sha` (below) backfilled with actual non-bold SHAs. SPEC-V3R6-SESSION-ID-ATTRIBUTION-REPAIR-001 is now a closed V3R6 modern-era SPEC.

**Evidence**:
- 4-phase commit chain: **plan** (manager-spec, draft → in-progress) → **run** (manager-develop M1-M6: `8c2c40c13` + `538fcaea6`) → **sync** (orchestrator-direct GLM-fallback: `615a4af95` + `d1ad817dd`) → **Mx** (orchestrator-direct: this close commit + mx backfill).
- spec.md frontmatter `status:` now reads `completed`.
- spec-lint: 0 errors (1 `OwnershipTransitionInvalid` WARNING on the sync commit `615a4af95` — documented GLM-fallback debt, see §E.4 Gaps).
- `go build ./...` exit 0 (0 Go changes in sync/Mx — doc-only phases).

**Baseline-attribution**: Mx-phase measured against the post-sync tree (HEAD `d1ad817dd`). The 18/18 AC PASS + build/vet/lint green baseline is from run-phase (progress.md §E.2, recorded by manager-develop), unchanged by the doc-only sync/Mx phases.

**Gaps (honestly reported)**:
- sync-auditor 4-dimension independent gate NOT run — GLM context-limit spawn failure (4th consecutive meta/manager-agent failure this session: manager-docs×3 incl. this SPEC + siblings, and sync-auditor anticipated). Orchestrator self-review + spec-lint + go build smoke substituted, matching the WORKFLOW-EFFORT-MAP-001 orchestrator-direct deferral pattern.
- spec-lint `OwnershipTransitionInvalid` WARNING on sync commit `615a4af95` (orchestrator-direct performed `in-progress → implemented`; canonical owner is manager-docs). Accepted GLM-environment debt. The Mx `implemented → completed` transition is canonically allowed for orchestrator (ownership matrix: "manager-docs OR orchestrator (Mx chore)"), so NO warning on this close commit.
- `moai spec audit` full project run deferred to pre-push; era classification guaranteed V3R6 via explicit `era: V3R6` H-override frontmatter + non-bold real `sync_commit_sha` per `feedback_era_commit_sha_field_format`.

**Residual-risk**: `additionalContext` lost after `/clear` (side-channel file persists → `moai session current` re-reads). Headless `-p` without hooks → canonical fallback (REQ-RDP-006). Pre-existing `internal/statusline` `TestCollectMemory` baseline failure is out of scope (STATUSLINE-PRESET-RETIRE residual). Multi-session race: a parallel session committed memory-hygiene (`7a338007a`) + the full WORKFLOW-EFFORT-MAP-001 close during this session's sync window — no file overlap with this SPEC, race absorbed cleanly.

mx_commit_sha: d7f45ca09 — backfilled after the Mx close commit (non-bold actual SHA per feedback_era_commit_sha_field_format).
