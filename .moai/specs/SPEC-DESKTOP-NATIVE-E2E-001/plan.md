# SPEC-DESKTOP-NATIVE-E2E-001 — Implementation Plan

> Tier M. Sibling of `spec.md` (v0.1.0, status: draft). Milestones are ordered by decision-reversibility: content-design decisions most likely to change during review come first; mechanical sync/verification steps come last.

---

## §A Context

### §A.1 Work location & branch

- Project root: `/Users/goos/MoAI/moai-adk-go`, Hybrid Trunk Route A (Tier M → main-direct commits; no PR unless `--pr`).
- Plan-phase artifacts land in `.moai/specs/SPEC-DESKTOP-NATIVE-E2E-001/` (4 files). Plan-phase commit subject pattern: `feat(SPEC-DESKTOP-NATIVE-E2E-001): plan-phase artifacts (M, 4 artifacts)` — `feat(` prefix mandatory (a `docs(` prefix misclassifies the SPEC as in-progress for the StatusGitConsistency lint).

### §A.2 Files in scope (6 write targets, all markdown)

| # | File | Change class |
|---|------|--------------|
| 1 | `.claude/skills/moai/workflows/e2e.md` | Detection Matrix row rework, graceful-exit rework, flags, Tool Matrix + probe-table rows, Execution Summary |
| 2 | `internal/template/templates/.claude/skills/moai/workflows/e2e.md` | byte-identical mirror of #1 |
| 3 | `.claude/agents/moai/e2e-specialist.md` | `### desktop-native` stub → per-OS recipes; artifact-conventions row |
| 4 | `internal/template/templates/.claude/agents/moai/e2e-specialist.md` | byte-identical mirror of #3 |
| 5 | `internal/template/templates/.claude/commands/moai/e2e.md.tmpl` | argument-hint `--platform` + `desktop-native` |
| 6 | `.claude/commands/moai/e2e.md` | same argument-hint delta (render-pattern parity, NOT byte parity) |

### §A.3 PRESERVE list

- Parent SPEC directory `.moai/specs/SPEC-E2E-REVIVAL-001/` — completed, immutable. Do NOT edit.
- All non-e2e detection-matrix rows, the web/mobile/desktop(-hybrid) toolchain recipes, the MCP Escalation Ladder, the Blocker Report Protocol, the Return Contract — untouched except where a desktop-native row is APPENDED.
- `catalog.yaml`, `model_policy.go`, `catalog_tier_audit_test.go`, `catalog_loader_test.go` — zero changes (C-2).
- Runtime-managed paths (`.moai/state/`, `.moai/harness/`, `.moai/cache/`) — untouched.

### §A.4 Measured anchors (2026-07-13; re-measure before run-phase edits — time-variant)

- Skill pair and agent pair each `diff` exit 0 (byte-identical).
- Deferral-wording family baseline: skill 3 matching lines/tree, agent 1/tree (CMD-DNE-011 regex).
- Verbatim removal target sentence (1 per tree, skill only): "There is no opt-in automation path for `desktop-native`."
- Command template body: 6 non-empty LOC. Local command argument-hint: `[--platform web|mobile|desktop]`.
- Content-token anchors (line numbers drift — anchor by token): Detection Matrix `desktop-native` row token "Native toolkit markers WITHOUT Electron/Tauri"; agent stub heading token "### desktop-native (non-Electron/non-Tauri)"; Execution Summary token "(incl. `desktop-native` deferral notice)".

---

## §B Known Issues (filtered to relevant categories)

- **B4 Frontmatter schema**: spec.md uses `created:`/`updated:`/`tags:` canonical names; sibling artifacts carry no lint-validated frontmatter.
- **B6 spec-lint heading convention**: §E carries `### Out of Scope — <topic>` H3 sub-headings with `-` bullets (OutOfScopeRule).
- **B8/B10 working-tree hygiene**: the working tree currently carries many unrelated modified files (config sections, harness proposals). Run-phase commits MUST use specific paths (`git add <file>...`), never `git add -A`.
- **Markdown table-cell pipe hazard** (parent lesson): `--platform web|mobile|desktop|desktop-native` enum text inside table cells requires `\|` escaping in the DELIVERABLE files, but verification commands must NEVER live in table cells — all executable ACs are pinned in acceptance.md § Executable Command Block (unescaped, non-vacuous).
- **Template neutrality (B-class)**: the new recipe prose in the TEMPLATE tree must carry zero SPEC IDs / REQ tokens / internal dates (CLAUDE.local.md §25 CI guard). Write the recipes as generic capability prose.
- **Windowed-grep undercount** (memory lesson): section-anchored greps (awk range) are used only for PRESENCE anchoring within a section, never for repo-wide counting.

---

## §C Pre-flight (run-phase entry checks)

```bash
# P1. Branch + baseline
git branch --show-current && git rev-parse HEAD

# P2. Re-measure byte-parity of both pairs (must be exit 0 before editing)
diff -q .claude/skills/moai/workflows/e2e.md internal/template/templates/.claude/skills/moai/workflows/e2e.md
diff -q .claude/agents/moai/e2e-specialist.md internal/template/templates/.claude/agents/moai/e2e-specialist.md

# P3. Re-measure the deferral-family baseline (expect skill=3, agent=1 per tree; if drifted, re-derive)
grep -cE "deferral notice|deferred to a follow-up|not yet provided|no opt-in automation path|not provided by this agent" \
  .claude/skills/moai/workflows/e2e.md

# P4. CI-guard baseline (must be green BEFORE edits to distinguish NEW failures)
go test ./internal/template/... > /tmp/dne-preflight.log 2>&1; echo "exit=$?"; tail -20 /tmp/dne-preflight.log

# P5. Evidence directory (redirect target for run-phase verification output)
mkdir -p .moai/state/verify/SPEC-DESKTOP-NATIVE-E2E-001
```

---

## §D Constraints (DO NOT VIOLATE)

- C-1..C-6 per spec.md §C (Template-First lockstep; no new agent; thin command; subagent boundary; token minimization; macOS-only live probes).
- Forbidden: `git add -A`, `--no-verify`, editing SPEC-E2E-REVIVAL-001, editing catalog/model-policy/CI-pin files, adding `.claude/agents/**` files.
- Required: Conventional Commits with full SPEC ID scope + `🗿 MoAI` trailer; `make build` after template edits; evidence redirected to `.moai/state/verify/SPEC-DESKTOP-NATIVE-E2E-001/` (not bare `/tmp`).
- Template-tree prose neutrality: recipes written as generic capability documentation (no internal provenance tokens).

---

## §E Self-Verification (run-phase completion deliverables)

Reported per the verification-claim-integrity 5-section format (Claim / Evidence / Baseline-attribution / Gaps / Residual-risk):

- E1: AC binary PASS/FAIL matrix (all 25 gating ACs + the 1 optional non-gating probe, CMD-ID-cited verbatim outputs).
- E2: `go test ./internal/template/...` exit 0 + `make build` exit 0.
- E3: Byte-parity diffs (CMD-DNE-009) exit 0 for both pairs.
- E4: Invariance sweep (CMD-DNE-010/011) — old sentence 0 matches, deferral family 0 matches, WinAppDriver family 0 matches.
- E5: `moai spec lint --strict .moai/specs/SPEC-DESKTOP-NATIVE-E2E-001/spec.md` → 0 errors (StatusGitConsistency WARNING is structurally expected until sync close — never lint.skip it).
- E6: Commit SHAs + push state.
- E7: Blocker report if any user decision is missing (never AskUserQuestion).

---

## §F Milestones (decision-reversibility order)

### M1 — Workflow-skill lane redesign (highest change-likelihood decisions)

The user-facing routing surface: what the Detection Matrix row claims, how the graceful branch splits, and which `--tool` tokens exist are the decisions most likely to be adjusted at review.

- Rework the Detection Matrix `desktop-native` row: positive AppKit/WinUI-Win32/Qt/GTK markers (REQ-DNE-001), notes column routes to the lane; Electron/Tauri rows stay above (REQ-DNE-002).
- Split the No-Target Graceful Exit: remove the deferral paragraph + verbatim sentence; add the desktop-native routing paragraph; keep the genuine-no-target branch (REQ-DNE-003/004).
- Extend Supported Flags: `--platform` enum + 5 `--tool` tokens (REQ-DNE-005). Table-cell pipes escaped (`\|`) in the deliverable.
- Add Tool Matrix per-OS rows with token-cost classes + Toolchain Probe rows (REQ-DNE-006/007).
- Add the host-OS/declarative-recipe rule (REQ-DNE-009) near Phase 0; drop the Execution Summary deferral mention (REQ-DNE-008).
- Edit the LOCAL file first for review, then mirror byte-identically to the template sibling (or edit both in lockstep) — pair parity is checked in M4.

### M2 — e2e-specialist per-OS recipes (toolchain-content decisions)

- Replace the `### desktop-native` stub with three OS subsections per §A.2 of spec.md: macOS (axcli pinned default + appium-mac2/WDIO fallback + TCC prerequisite, REQ-DNE-101..103), Windows declarative (FlaUI.WebDriver pinned + `GET /status` probe; pywinauto fallback, REQ-DNE-104/105), Linux declarative (dogtail 2.x + AT-SPI2 prerequisites + Qt env + Wayland caveat; ydotool/xdotool + screenshot-verify fallback, REQ-DNE-106/107).
- Add the last-resort section: AX-tree snapshot loop + computer-use screenshot-loop cost caveat (REQ-DNE-108) and the token-cost ordering + bounded-tail wiring (REQ-DNE-110).
- Zero WinAppDriver/appium-windows-driver references (REQ-DNE-109).
- Extend Artifact Directory Conventions (`e2e/desktop-native/` scripts row; AX snapshots ride `e2e/.runs/`) (REQ-DNE-111) + missing-toolchain sequence pointer (REQ-DNE-112).
- Mirror byte-identically to the template sibling.

### M3 — Command argument-hint surfaces (mechanical)

- `e2e.md.tmpl`: `[--platform web|mobile|desktop|desktop-native]` in argument-hint; body stays 6 non-empty LOC (REQ-DNE-200).
- Local `.claude/commands/moai/e2e.md`: same argument-hint delta (REQ-DNE-201).

### M4 — Cross-tree sync + build + CI guards (mechanical)

- `diff` both pairs → exit 0 (REQ-DNE-300); `make build` (REQ-DNE-301).
- `go test ./internal/template/...` exit 0; thin-command + frontmatter-consistency + neutrality guards green (REQ-DNE-302..304).

### M5 — Invariance sweep + lint + close-out

- Run the full acceptance.md Executable Command Block; capture evidence under `.moai/state/verify/SPEC-DESKTOP-NATIVE-E2E-001/`.
- Deferral-family + old-sentence + WinAppDriver-family greps all 0 (REQ-DNE-305/109/004).
- `moai spec lint --strict` on spec.md → 0 errors.
- §E self-verification report; commits pushed.

---

## §G Anti-Patterns

- Editing only the local tree (or only the template tree) for the skill/agent pairs — parity break is the #1 historical failure mode for this file family.
- Asserting byte parity for the COMMAND pair — commands are rendered vs `.tmpl` and differ by design; only the argument-hint delta is verified.
- Putting pipe-bearing verification commands in acceptance-table cells — `\|` escaping makes the grep vacuous; commands live only in the fenced Executable Command Block.
- Re-porting the parent's retired-baseline prose verbatim (stale tool facts); the §A.2 matrix in spec.md is the sole normative toolchain source.
- Counting `desktop-native` token occurrences as an AC — token presence is not reachability; ACs anchor content within its owning section instead.
- Writing SPEC-internal provenance (SPEC IDs, dates, REQ tokens) into TEMPLATE-tree recipe prose.

---

## §H Cross-References

- Parent: `.moai/specs/SPEC-E2E-REVIVAL-001/` (completed) — Group A/B doctrine, REQ-E2E-006/007 patterns, §E deferral entry this SPEC discharges.
- `.claude/rules/moai/core/agent-common-protocol.md` § Parallel Execution — file-redirect/bounded-tail contract mirrored by REQ-DNE-110.
- `.claude/rules/moai/development/spec-frontmatter-schema.md` — frontmatter + status-transition ownership.
- `CLAUDE.local.md` §2 (Template-First), §15 (language neutrality), §25 (template internal-content isolation).
- `internal/template/commands_audit_test.go` (`TestCommandsThinPattern`) — the <20 non-empty LOC gate.
- Execution vehicle option for run-phase: single manager-develop delegation (Mode 5), cycle_type=ddd (docs-artifact transformation, behavior-preserving for non-desktop-native lanes).
