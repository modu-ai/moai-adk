---
id: SPEC-SKILL-BODY-NEUTRALITY-001
title: "Skill-Body Neutrality — acceptance criteria"
version: "0.1.0"
status: draft
created: 2026-06-04
updated: 2026-06-04
author: manager-spec
priority: P1
phase: "v3.0.0"
module: "internal/template/templates/.claude/skills"
lifecycle: spec-anchored
tags: "template-system, neutrality, skills, ci-guard, distribution"
tier: M
---

# Acceptance Criteria — SPEC-SKILL-BODY-NEUTRALITY-001

All AC are grep-verifiable or test-runnable. `$SK` denotes `internal/template/templates/.claude/skills`. Every "0 matches" assertion is scoped to `$SK` and excludes the EXCL-SBN-003 allow-list. Run from repo root.

## §A Part A — Purge (Given-When-Then)

### AC-SBN-001 — CLASS 1 CI-test-file reference removed (REQ-SBN-001)

- **Given** 6 skill files name `internal/template/agentless_audit_test.go`,
- **When** Part A (M2) completes,
- **Then** `grep -rln 'agentless_audit_test' $SK/` returns **0 files**.

### AC-SBN-002 — Mode sentinels retained (REQ-SBN-002) [MUST-PASS]

- **Given** the existing sentinel-presence tests require the keyword values in skill bodies,
- **When** Part A rewrites the surrounding prose,
- **Then** all of the following still match:
  - `grep -rl 'MODE_UNKNOWN' $SK/moai/workflows/` (run.md + design.md present)
  - `grep -rl 'MODE_PIPELINE_ONLY_UTILITY' $SK/moai/workflows/` (≥4 files present)
  - `grep -rl 'MODE_FLAG_IGNORED_FOR_UTILITY' $SK/moai/workflows/` (present)
- **And** `go test ./internal/template/... -run 'Sentinel|AgentlessControlFlow|LoopAlias'` passes (0 failures).

### AC-SBN-003 — Sentinel prose generic-ized (REQ-SBN-003)

- **Given** sentinel prose previously named the internal test + a REQ token,
- **When** Part A (M2) completes,
- **Then** in the 6 CLASS-1 files, no line both contains a `MODE_*` sentinel AND a `REQ-WF` token:
  - `grep -rnE 'MODE_(UNKNOWN|PIPELINE_ONLY_UTILITY|FLAG_IGNORED_FOR_UTILITY|TEAM_UNAVAILABLE)' $SK/ | grep -E 'REQ-WF[0-9]'` returns **0 matches**.

### AC-SBN-004 — Release build generic-ized (REQ-SBN-004)

- **Given** `sync/delivery.md` hardcodes moai-adk's 5-platform `./cmd/moai/` build + pinned golangci-lint,
- **When** Part A (M2) completes,
- **Then**:
  - `grep -nE 'GOOS=(linux|darwin|windows).*cmd/moai/' $SK/moai/workflows/sync/delivery.md` returns **0 matches**.
  - `grep -n 'golangci-lint.*@v2\.1\.6' $SK/moai/workflows/sync/delivery.md` returns **0 matches**.
- **And** the file still teaches a multi-platform build pattern (manual: `<your-binary>`/`<your-module>` or equivalent placeholder present).

### AC-SBN-005 — Real internal Go-impl paths removed (REQ-SBN-005)

- **Given** skill bodies reference real moai-adk `internal/<pkg>/<file>.go` paths,
- **When** Part A (M2+M3) completes,
- **Then** none of the real internal paths remain in `$SK` (excluding the EXCL-SBN-003 illustrative examples):
  - `grep -rnoE 'internal/(spec|cli|hook|ciwatch|design)/[a-z0-9_/]*\.go' $SK/` returns **0 matches**.
  - (Allow-listed illustrative example paths `internal/auth/login.go`, `internal/api/handler.go` in `pr-review-multi-agent.md` and `internal/core/handler.go` in `mx.md` are NOT matched by the above regex and remain present.)

### AC-SBN-006 — Internal V3R-family SPEC IDs removed (REQ-SBN-006)

- **Given** 37 files carry `SPEC-V3R[0-9]-*` / `CONST-V3R[0-9]-*` plus named real IDs,
- **When** Part A (M4) completes,
- **Then**:
  - `grep -rnoE 'SPEC-V3R[0-9]-[A-Z0-9-]+|CONST-V3R[0-9]-[0-9]+' $SK/` returns **0 matches**.
  - `grep -rnE 'SPEC-WF-AUDIT-GATE-001|SPEC-MX-001' $SK/` returns **0 matches**.
- **And** format-example placeholders are preserved (AC-SBN-010).

### AC-SBN-007 — Internal REQ/AC tokens removed (REQ-SBN-007)

- **Given** REQ-token families (`REQ-WF003`, `REQ-BRAIN`, `REQ-SKILL`, `REQ-ROUTE`, `REQ-PH`, `REQ-LB`, `REQ-DPL`, `REQ-BRIEF`, etc.) appear in skill prose,
- **When** Part A (M4) completes,
- **Then** `grep -rnoE '\bREQ-[A-Z][A-Z0-9]*-[0-9]+\b|\bREQ-WF[0-9]+-[0-9]+\b' $SK/` returns **0 matches** (excluding any entry registered in the pedagogical allowlist, of which there are none for REQ tokens).

### AC-SBN-008 — Dev-only command self-references removed (REQ-SBN-008)

- **Given** `moai/SKILL.md` + `commands-reference.md` + `INDEX.md` reference `97-release-update` / `99-release`,
- **When** Part A (M5) completes,
- **Then**:
  - `grep -rnE '97-release-update|/moai:99-release|99-release' $SK/` returns **0 matches**.
  - `grep -rn 'NOT distributed to user projects' $SK/` returns **0 matches**.

### AC-SBN-009 — Maintainer doctrine removed (REQ-SBN-009)

- **Given** `moai-meta-harness/SKILL.md:168` carries the maintainer doctrine with an internal date + "catch-up SPEC",
- **When** Part A (M5) completes,
- **Then**:
  - `grep -nE 'Doctrine-code drift|catch-up SPEC|maintainer doctrine' $SK/moai-meta-harness/SKILL.md` returns **0 matches**.
  - `grep -nE '\b2026-05-26\b' $SK/moai-meta-harness/SKILL.md` returns **0 matches**.
- **And** a generic namespace-separation statement remains (manual: the `harness-*` vs `moai-*` ownership policy is still described).

### AC-SBN-010 — Format-example placeholders preserved (REQ-SBN-010) [keep-list guard]

- **Given** the EXCL-SBN-003 keep-list,
- **When** Part A completes,
- **Then** all of the following STILL match (purge did not over-reach):
  - `grep -rl 'SPEC-AUTH-001' $SK/` (≥1 file)
  - `grep -rl 'SPEC-001' $SK/` (≥1 file)
  - `grep -rl 'internal/auth/login.go' $SK/moai-workflow-testing/references/pr-review-multi-agent.md` (present)
  - `grep -rl 'internal/core/handler.go' $SK/moai/workflows/mx.md` (present)

### AC-SBN-011 — Local mirror synced (REQ-SBN-011)

- **Given** Part A edits template SSOT files,
- **When** each Part A milestone completes,
- **Then** for every edited file path, `diff <(cat internal/template/templates/.claude/skills/<path>) <(cat .claude/skills/<path>)` shows **no difference** (template and local mirror identical).

## §B Part B — Guard extension (Given-When-Then)

### AC-SBN-012 — Guard fails on CI-test-file leak (REQ-SBN-012, RED evidence)

- **Given** the extended guard before Part A purge,
- **When** `go test ./internal/template/... -run TestTemplateNoInternalContentLeak` runs at the M1 RED checkpoint,
- **Then** the test FAILS and its output names the `agentless_audit_test` class with ≥6 file findings.

### AC-SBN-013 — Guard has a Go-impl-path class (REQ-SBN-013)

- **Given** the extended guard,
- **When** inspecting `internal/template/internal_content_leak_test.go`,
- **Then** `grep -nE 'internal/.*\\.go|internal-go-path|C7' internal/template/internal_content_leak_test.go` shows a leak class whose regex matches `internal/<pkg>/<file>.go`.

### AC-SBN-014 — SPEC-ID class broadened to V3R0-9 (REQ-SBN-014)

- **Given** the existing C1 matched only `SPEC-(V3R6|AGENCY|WORKTREE)-`,
- **When** Part B (M1) completes,
- **Then** `grep -nE 'V3R\[0-9\]|V3R\[2-6\]|V3R0-9' internal/template/internal_content_leak_test.go` shows the broadened SPEC-ID pattern matching the V3R2-V3R5 families.

### AC-SBN-015 — Pedagogical allowlist covers placeholders (REQ-SBN-015)

- **Given** the EXCL-SBN-003 keep-list,
- **When** the extended guard runs after Part A,
- **Then** the allow-listed placeholders produce **0 findings** — i.e., the final guard run (AC-SBN-016) is GREEN despite `SPEC-AUTH-001` / illustrative paths still being present (proving the allowlist, not deletion, suppresses them).

### AC-SBN-016 — Guard GREEN after purge (REQ-SBN-016) [MUST-PASS]

- **Given** Part A complete + Part B classes/allowlist in place,
- **When** `go test ./internal/template/... -run TestTemplateNoInternalContentLeak` runs at M6,
- **Then** the test PASSES (0 findings for CLASS 1-4 skill-body leak classes).

### AC-SBN-017 — Recurrence regression backstop (REQ-SBN-017)

- **Given** the GREEN guard,
- **When** a CLASS 1-4 leak is reintroduced into any skill body (simulated in a focused sub-test or documented manual check),
- **Then** the guard FAILS — demonstrated by the RED→GREEN evidence captured at M1 and M6.

### AC-SBN-018 — No date/sha class overlap (REQ-SBN-018 / EXCL-SBN-001) [partition guard]

- **Given** ISOLATION-001 owns the date + short-sha strict-tier classes,
- **When** inspecting this SPEC's guard additions,
- **Then** the NEW classes added by this SPEC do NOT include a generic date regex (`202[6-9]-MM-DD`) or a short-sha regex (`[0-9a-f]{7,8}`) — verified by diffing the added `leakClass` entries against the existing `strictLeakClasses` (no new date/sha class).

## §C Quality Gate / Definition of Done

- [ ] All 18 AC pass (grep assertions return the asserted counts; tests pass).
- [ ] `go test ./internal/template/...` passes (extended leak test GREEN + sentinel tests GREEN).
- [ ] `go build ./...` succeeds; `go vet ./internal/template/...` clean.
- [ ] `.github/workflows/template-neutrality-check.yaml` path filter covers `internal/template/templates/.claude/skills/**` and runs the leak test.
- [ ] Local `.claude/skills/` mirror is byte-identical to `internal/template/templates/.claude/skills/` for every edited file.
- [ ] RED→GREEN evidence captured: M1 failing finding list + M6 passing run recorded in run-phase progress.
- [ ] No unrelated working-tree edits committed (specific-path discipline).

## §D Edge Cases

- **EC-1** — A V3R-family SPEC ID embedded inside a longer token that is actually a placeholder (none found at baseline; if one appears, the pedagogical allowlist takes precedence over the leak class).
- **EC-2** — A skill body that legitimately documents the `go install github.com/modu-ai/moai-adk/cmd/moai@latest` user-facing install command (EXCL-SBN-004) — the Go-impl-path regex (`internal/...`) does not match an install URL, so no special-casing needed; verify it remains present.
- **EC-3** — The docs-site URL `adk.mo.ai.kr` (EXCL-SBN-005) — keep the URL; only the "4-locale" annotation is dropped. A grep for `adk.mo.ai.kr` after M4 should still return ≥1 match.
- **EC-4** — Mixed-language modified-files list in `mx.md` (`internal/core/handler.go`, `src/api/server.ts`, `lib/utils/helper.py`) — illustrative; the Go-impl-path regex is scoped to real moai-adk packages (`spec|cli|hook|ciwatch|design`) and does not match `internal/core/`, so the example survives without an allowlist entry, OR an explicit allowlist entry is added for belt-and-suspenders.
