---
id: SPEC-SKILL-BODY-NEUTRALITY-001
title: "Skill-Body Neutrality — implementation plan"
version: "0.1.1"
status: in-progress
created: 2026-06-04
updated: 2026-06-23
author: manager-spec
priority: P1
phase: "v3.0.0"
module: "internal/template/templates/.claude/skills"
lifecycle: spec-anchored
tags: "template-system, neutrality, skills, ci-guard, distribution"
tier: M
---

# Implementation Plan — SPEC-SKILL-BODY-NEUTRALITY-001

## §A Context

Part A purges moai-adk local-dev traces from ~deployed skill bodies (Part A surface is the subset of the 221 skill `.md` files that carry CLASS 1-4 leaks — orchestrator-verified ~40 files). Part B extends the existing neutrality guard so the leaks cannot re-accumulate. The work is sequenced so the **guard test is written first (RED)** to mechanically prove the leaks exist, then Part A purges them (GREEN), then Part B's allowlist + class additions are finalized.

Key architectural facts established at plan-phase (ground-truth, run phase must re-verify):
- The guard `internal/template/internal_content_leak_test.go` already `filepath.WalkDir`s `templatesRoot` and scans `.md`/`.tmpl`/`.yaml`/`.sh`/`.json`. It reaches skill bodies today. The gap is **leak-class pattern coverage**, not walk scope (EXCL-SBN-008).
- The existing C1 SPEC-ID class matches only `SPEC-(V3R6|AGENCY|WORKTREE)-*` — it misses the `SPEC-V3R2..V3R5` families that dominate skill-body leaks (REQ-SBN-014).
- `agentless_audit_test.go` asserts the **sentinel keyword value** is present in the skill body (`strings.Contains(data, "MODE_UNKNOWN")`). It does NOT require the body to name the test file. So the sentinel keyword stays; only the surrounding "CI guards in internal/template/agentless_audit_test.go enforce ..." prose is rewritten (REQ-SBN-002 / REQ-SBN-003). This resolves the self-reference paradox cleanly.
- A `pedagogicalAllowlist` (consulted via `isPedagogicallyAllowed(relForAllowlist, trimmed)`) is the right mechanism to whitelist legitimate placeholders (REQ-SBN-015).
- `//go:embed all:templates` — no `embedded.go`; `make build` picks up edits. Local `.claude/skills/` mirror must be synced (REQ-SBN-011).

## §B Known Issues / Risks

- **R1 — Sentinel-presence regression.** If a Part A edit accidentally drops a sentinel keyword while rewriting prose, `agentless_audit_test.go` fails. Mitigation: REQ-SBN-002 is an explicit invariant; run phase runs `go test ./internal/template/...` after each milestone batch.
- **R2 — Over-purge of legitimate placeholders.** A too-broad Go-impl-path or SPEC-ID regex could match illustrative examples (`internal/auth/login.go` in pr-review example, `SPEC-PAY-001`). Mitigation: pedagogical allowlist (REQ-SBN-015) + the EXCL-SBN-003 keep-list; the guard's allowlist is built BEFORE Part A purge so the RED run distinguishes real leaks from placeholders.
- **R3 — Dual-allow-list drift.** Adding a date/sha class here would overlap ISOLATION-001's strict tier. Mitigation: EXCL-SBN-001 / REQ-SBN-018 forbid date/sha classes in this SPEC.
- **R4 — Local-mirror divergence.** Forgetting to sync `.claude/skills/` leaves the local copy stale. Mitigation: REQ-SBN-011; each Part A milestone edits template + mirror together.
- **R5 — `delivery.md` build-step meaning loss.** Generic-izing the 5-platform GOOS build must keep the "build for multiple platforms in CI" teaching intact while dropping `./cmd/moai/` and the pinned golangci-lint version. Mitigation: M2 generic-izes to a `<your-binary>` / `<your-module>` placeholder pattern, not deletion.

## §C Pre-flight Verification (run-phase entry)

```bash
# 1. Branch clean for our scope (unrelated edits present — specific-path only)
git status --short
# 2. Baseline leak counts (must match spec §B before purge)
grep -rln 'agentless_audit_test' internal/template/templates/.claude/skills/ | wc -l   # expect 6
grep -rlE 'SPEC-V3R[0-9]|CONST-V3R[0-9]' internal/template/templates/.claude/skills/ | wc -l   # expect 37
# 3. Existing guard + sentinel tests green at baseline
go test ./internal/template/... -run 'TestTemplateNoInternalContentLeak|Sentinel|AgentlessControlFlow|LoopAlias' 2>&1 | tail -5
```

## §D Constraints (inherited from spec §E)

- Template-First: edit `internal/template/templates/` SSOT + sync `.claude/skills/` mirror.
- Meaning preservation: generic-ize, never blind-delete.
- Specific-path commit: stage only `.moai/specs/SPEC-SKILL-BODY-NEUTRALITY-001/` for the plan commit; run-phase commits stage only skills + the guard test file.
- `Authored-By-Agent:` trailer per phase owner.

## §E Self-Verification (plan-phase self-check)

- [x] All 18 REQ map to at least one milestone (M1-M6 below) and at least one AC.
- [x] Every leak class in §B has a Part A milestone AND a Part B guard class.
- [x] EXCL list covers the orchestrator-confirmed false positives (placeholders, multi-lang tool lists, illustrative examples, docs-site URL).
- [x] No date/sha class introduced (REQ-SBN-018 / EXCL-SBN-001).
- [x] Sentinel-keyword invariant (REQ-SBN-002) is an explicit milestone gate.

## §F Milestones

Part A is grouped by leak class / file cluster so run-phase can purge in cohesive batches. Part B is TDD (RED guard first, then GREEN). Priority labels per the no-time-estimate rule.

### M1 — Part B RED: extend the neutrality guard to fail on current leaks (Priority High)

- Add leak classes to `internal/template/internal_content_leak_test.go`:
  - `C6-agentless-test-ref` — matches the literal `agentless_audit_test.go` reference (REQ-SBN-012).
  - `C7-internal-go-path` — [HARD per REQ-SBN-013 / D5] matches ONLY real moai-adk package paths via the package-restricted regex `internal/(spec|cli|hook|ciwatch|design)/[a-z0-9_/]*\.go`. MUST NOT use the unrestricted `internal/.*\.go` form (it would match the EXCL-SBN-003 illustrative paths `internal/auth/login.go`, `internal/api/handler.go`, `internal/core/handler.go` and make M6 GREEN unreachable). Keep the package set in sync with AC-SBN-005.
  - Broaden `C1-spec-id-prefix` to also match `SPEC-V3R[0-9]-` and `CONST-V3R[0-9]-` (REQ-SBN-014).
- REQ-token enforcement for skill bodies (REQ-SBN-007): **promote/reuse** the existing opt-in `S3-req-ac-token-any-prefix` regex (`(REQ|AC)-[A-Z]{2,}-[0-9]{3}`, currently strict-tier-only) into the default tier for the skill-body scan — do NOT add a near-identical sibling class (REQ-SBN-018(b) / AC-SBN-018(b) partition guard: at most ONE leakClass matching the REQ-token pattern across `leakClasses` + `strictLeakClasses`). Partition away from the strict-tier date/sha classes (REQ-SBN-018(a)).
- Build the `pedagogicalAllowlist` entries for the EXCL-SBN-003 keep-list (REQ-SBN-015), INCLUDING explicit belt-and-suspenders entries for the 3 illustrative Go paths (`internal/auth/login.go`, `internal/api/handler.go`, `internal/core/handler.go`) per REQ-SBN-013, so they are excluded from the RED run and protected even if the C7 regex restriction regresses.
- **Gate (RED):** the extended test FAILS, reporting the CLASS 1-4 leaks enumerated in spec §B and NOT reporting any allow-listed placeholder. Capture the failing finding list as run-phase evidence.

### M2 — Part A CLASS 1 + CLASS 2 build/Go-path purge: workflow skills (Priority High)

- CLASS 1 (6 files): rewrite the `agentless_audit_test.go` prose to keep the sentinel keyword value, drop the test-file name + REQ token (REQ-SBN-001/002/003). Pattern: "CI guards in `internal/template/agentless_audit_test.go` enforce the literal `MODE_UNKNOWN` sentinel (REQ-WF003-010)" → "`MODE_UNKNOWN` is a stable error key emitted for an unrecognized `--mode` value; keep it verbatim."
- CLASS 2 release build (`sync/delivery.md`): generic-ize the 5-platform `GOOS=... ./cmd/moai/` block to a `<your-binary>` / `<your-module>` multi-platform pattern; drop the pinned `golangci-lint@v2.1.6` version (REQ-SBN-004).
- CLASS 2 Go-impl paths in workflow skills (`quality-gates-context.md`, `plan/spec-assembly.md`, `harness.md`, `team/plan.md`): generic-ize each real path to its mechanism description (REQ-SBN-005).
- Sync local `.claude/skills/` mirror (REQ-SBN-011).
- **Gate:** `go test ./internal/template/... -run Sentinel` still green (sentinels retained).

### M3 — Part A CLASS 2 Go-path purge: non-workflow skills (Priority Medium)

- `moai-workflow-worktree/SKILL.md`, `moai-workflow-spec/SKILL.md` + `references/reference.md`, `moai-workflow-design/SKILL.md`, `moai-workflow-ci-loop/SKILL.md`: generic-ize remaining real Go-impl paths (REQ-SBN-005).
- Sync mirror (REQ-SBN-011).

### M4 — Part A CLASS 3 purge: internal SPEC IDs + REQ tokens (Priority High)

- Across the 37 files with V3R-family SPEC IDs + the REQ-token files (worst-affected: `design.md`, `brain.md`, `harness.md`, `plan/spec-assembly.md`): replace each real internal SPEC ID and REQ token with its plain-language policy/mechanism description (REQ-SBN-006/007). Preserve placeholders per EXCL-SBN-003.
- Drop the docs-site URL's "4-locale" maintainer annotation at the §B.5 enumerated sites (REQ-SBN-019), keeping the `adk.mo.ai.kr` URL (EXCL-SBN-005). The 4 sites: `moai-foundation-core/modules/spec-ears-format.md:11`, `moai-workflow-spec/SKILL.md:65`, `moai-workflow-spec/SKILL.md:146`, `moai-workflow-spec/references/reference.md:27`. Handle both surface forms (`4-locale: en / ko / ja / zh` and `4-locale (en / ko / ja / zh)`). (The `moai/SKILL.md:76,240` 4-locale annotations are removed by M5 wholesale with the dev-only `release-update` entry, not here.)
- Sync mirror (REQ-SBN-011).

### M5 — Part A CLASS 4 purge: dev-only self-reference + maintainer doctrine (Priority Medium)

- `moai/SKILL.md` (76, 238, 244-245): remove the `release-update` dev-only entry + "NOT distributed to user projects (97-release-update.md)" self-contradiction from the user-facing command surface (REQ-SBN-008).
- `commands-reference.md` (21, 264, 272, 329) + `INDEX.md` (143): remove the `/moai:99-release` dev-only-reserved command rows (REQ-SBN-008).
- `moai/references/reference.md` (244): remove the `Note: /moai:99-release is a separate local-only command ...` self-reference line — this is the **6th** `99-release` baseline hit (§B.4 / D1); it MUST be purged for AC-SBN-008's `grep -rn '99-release'` → 0 (REQ-SBN-008).
- `moai-meta-harness/SKILL.md` (168): replace the maintainer doctrine note (internal date + catch-up-SPEC ref) with a generic namespace-separation statement (REQ-SBN-009).
- Sync mirror (REQ-SBN-011).

### M6 — Part B GREEN + finalize (Priority High)

- Re-run the extended neutrality guard: it now reports ZERO skill-body leak-class findings (REQ-SBN-016).
- Add a focused recurrence-regression assertion documenting the RED→GREEN transition (REQ-SBN-017) so a future re-leak fails.
- Wire the extended test into `.github/workflows/template-neutrality-check.yaml` trigger coverage if not already covered by the path glob (verify the workflow's path filter includes `internal/template/templates/.claude/skills/**` and the test invocation runs the leak test).
- Full verification batch: `go test ./internal/template/...`, `go build ./...`, `go vet ./internal/template/...`.

## §G Anti-Patterns (do NOT do in run phase)

- Blind-deleting a leak line (strips user-facing meaning) instead of generic-izing.
- Dropping a mode sentinel keyword while rewriting prose (breaks `agentless_audit_test.go`).
- Adding a date/sha leak class (overlaps ISOLATION-001 strict tier → dual-allow-list drift).
- Editing the template SSOT but forgetting the `.claude/skills/` mirror.
- Over-broad Go-path regex that swallows illustrative example paths — build the allowlist first.
- Sweeping the unrelated working-tree edits into a commit.

## §H Cross-References

- spec.md §B — orchestrator-verified leak inventory (file:line).
- acceptance.md — grep-verifiable AC per requirement.
- `internal/template/internal_content_leak_test.go` — Part B wiring point (leakClasses, strictLeakClasses, pedagogicalAllowlist, WalkDir).
- `internal/template/agentless_audit_test.go` — sentinel-presence tests (the REQ-SBN-002 invariant guard).
- `.github/workflows/template-neutrality-check.yaml` — CI trigger.
- `SPEC-V3R6-TEMPLATE-NEUTRALITY-AUDIT-001` — rules-file neutrality (complementary; rules not skills).
- `SPEC-V3R6-TEMPLATE-INTERNAL-ISOLATION-001` — owns date/sha strict-tier classes (partition boundary, EXCL-SBN-001).
- CLAUDE.local.md §2 (Template-First), §15 (language neutrality), §21 (dev-only isolation), §25 (template internal-content isolation).
