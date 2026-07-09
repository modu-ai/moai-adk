---
id: SPEC-MEMORY-DIET-001
title: "Safe always-loaded context diet — implementation plan"
version: "0.1.0"
status: in-progress
created: 2026-07-10
updated: 2026-07-10
author: GOOS행님
priority: P2
phase: "v14.4.0 target"
module: ".claude/rules/moai/workflow + ~/.claude/projects/*/memory"
lifecycle: spec-anchored
tags: "context-diet, always-loaded, path-scope, memory-index, template-first, safe-diet"
tier: M
---

# plan.md — SPEC-MEMORY-DIET-001

## §A Context

### A.1 Baseline measurements (orchestrator Discovery, 2026-07-10)

| File group | File | Lines | Bytes | Tokens (via /context) |
|------------|------|-------|-------|-----------------------|
| REQ-1 | `.claude/rules/moai/workflow/cadence-bridge.md` | 88 | 9,855 | 2.3k |
| REQ-2 | `.claude/rules/moai/workflow/session-handoff.md` | 478 | 56,598 | 13.3k |
| REQ-3 | `~/.claude/projects/-Users-goos-MoAI-moai-adk-go/memory/MEMORY.md` | 90 | 16,758 | 5.8k |
| **Total** | | **656** | **83,211** | **21.4k** |

**Always-loaded surface total (orchestrator /context):** 62.1k tokens (6.2%) across 12 files.

### A.2 Expected savings

| REQ | Target | Expected saving |
|-----|--------|-----------------|
| REQ-1 | cadence-bridge path-scoped | ~2.3k tokens (file leaves always-loaded) |
| REQ-2 | session-handoff example extraction | ~2k tokens (2 Example sections ~51 lines + ja/zh Localization rows ~8 lines relocated) |
| REQ-3 | MEMORY.md archive pruning | ~1.5k tokens (~20 stable ✅ entries moved to archive) |
| **Combined** | | **~5.8k tokens** (target ≥ ~5.5k per REQ-MD-016) |

### A.3 Template-mirror status (verified 2026-07-10)

| File | Local bytes | Template bytes | Status |
|------|-------------|----------------|--------|
| cadence-bridge.md | 9,855 | 9,855 | byte-identical |
| session-handoff.md | 56,598 | 56,598 | byte-identical |
| MEMORY.md | (auto-memory) | (no mirror) | N/A — no template obligation |

### A.4 PRESERVE list (untouched paths — scope discipline B10)

The following are explicitly NOT touched by this SPEC:
- `.claude/rules/moai/core/agent-common-protocol.md` (KEEP per RULE-DIET-002)
- `.claude/rules/moai/core/askuser-protocol.md` (KEEP per RULE-DIET-002)
- `.claude/rules/moai/core/verification-claim-integrity.md` (KEEP per RULE-DIET-002)
- `.claude/rules/moai/workflow/context-window-management.md` (KEEP per RULE-DIET-002)
- `.claude/rules/moai/workflow/goal-directive.md` (already scoped per RULE-DIET-002)
- `CLAUDE.md`, `CLAUDE.local.md` (owned by completed diet SPECs)
- `.claude/output-styles/moai/moai.md` (render surface — REQ-2 verifies parity but does NOT edit it)
- All Go source files (`internal/`, `pkg/`, `cmd/`)
- All `.claude/hooks/moai/*.sh` scripts

---

## §B Known Issues (auto-injected per manager-develop-prompt-template §B)

### B1. Cross-platform build tags
- Not applicable — this SPEC edits markdown doctrine files only. No Go source changes.

### B2. Cross-SPEC policy conflict
- REQ-1 (cadence-bridge path-scope) is the same mechanism family as RULE-DIET-002. No conflict — RULE-DIET-002 is `completed` and explicitly did NOT cover cadence-bridge.md (which post-dates it).
- REQ-2 (session-handoff extraction) builds on RULES-COMPRESS-001 (implemented). No conflict — RULES-COMPRESS-001 compressed prose; REQ-2 extracts illustrative content. The two are complementary.

### B3. C-HRA-008 / Subagent boundary
- Not applicable — no `internal/harness/` or `internal/hook/` code changes.

### B4. Frontmatter canonical schema
- This SPEC's own frontmatter uses the 12 canonical fields. The RULE files edited by REQ-1/REQ-2 use a different schema (`paths:` + `description:`, per coding-standards.md § Paths Frontmatter) — this is correct and intentional.

### B5. CI 3-tier
- spec-lint: the SPEC must pass `moai spec lint`.
- golangci-lint: no Go changes → baseline preserved.
- Test: no Go test changes → baseline preserved.
- Template neutrality CI: REQ-1/REQ-2 template mirrors must not introduce SPEC IDs / REQ tokens / dates / SHAs.

### B6. spec-lint heading convention
- The Out of Scope section uses `### Out of Scope — <topic>` H3 sub-headings (5 of them) — satisfies `OutOfScopeRule`.

### B7-B12
- Not applicable (no observer.go, no working-tree runtime files, no hook changes, no git-strategy changes). B9 (Git Commit + Push): per Hybrid Trunk Tier M, manager-develop commits + pushes to main directly. B11 (AskUserQuestion): subagent returns blocker reports, never prompts.

---

## §C Pre-flight checks (run-phase, before any edit)

```bash
# 1. Branch + baseline
git branch --show-current
git rev-parse HEAD

# 2. Build baseline (no Go changes expected, but verify)
go build ./...

# 3. Lint baseline
golangci-lint run --timeout=2m 2>&1 | tail -5

# 4. Template-neutrality baseline
go test ./internal/template/... 2>&1 | tail -10

# 5. SPEC lint baseline
moai spec lint .moai/specs/SPEC-MEMORY-DIET-001/spec.md

# 6. Token baseline (via /context — orchestrator-side, not Bash)
# Record: "Memory files: Xk tokens (Y%)"

# 7. cadence-bridge frontmatter baseline (confirm no existing frontmatter)
head -3 .claude/rules/moai/workflow/cadence-bridge.md
head -3 internal/template/templates/.claude/rules/moai/workflow/cadence-bridge.md

# 8. session-handoff section inventory (confirm Example + Localization sections exist)
grep -n '^### Example\|^### Localization Table' .claude/rules/moai/workflow/session-handoff.md

# 9. MEMORY.md marker distribution baseline
grep -c '✅' ~/.claude/projects/-Users-goos-MoAI-moai-adk-go/memory/MEMORY.md
grep -c '🟢\|🟡\|🆕\|⏸️\|⚠️\|🔍' ~/.claude/projects/-Users-goos-MoAI-moai-adk-go/memory/MEMORY.md

# 10. Archive file existence
ls -la ~/.claude/projects/-Users-goos-MoAI-moai-adk-go/memory/MEMORY-archive-2026-06-02.md
```

---

## §D Constraints (DO NOT VIOLATE)

1. **PRESERVE list (§A.4)**: no file outside the 3 REQ targets + their template mirrors + the sibling references file (REQ-2) + the archive file (REQ-3) shall be modified.
2. **No `--no-verify`**: pre-commit hook warn-only is normal; never bypass.
3. **No `--amend` / force-push** to main (Hybrid Trunk Tier M = direct push, no rewrite).
4. **Conventional Commits**: `feat(SPEC-MEMORY-DIET-001): M{N} <subject>` + `🗿 MoAI` trailer.
5. **Template-First**: edit template tree first, then `make build`, then verify live-tree parity.
6. **Verbatim preservation (REQ-2)**: `✂` (U+2702), `─` (U+2500), 6-block skeleton headings, Pre-emit self-check labels — byte-identical survival.
7. **Neutrality**: no SPEC IDs / REQ tokens / internal dates / commit SHAs / audit citations in template mirrors.
8. **MEMORY.md no-deletion**: archive (move), never delete. Active-marker entries untouched.

---

## §E Self-Verification deliverables (manager-develop §E matrix)

Run-phase completion report MUST include:

- **E1**: AC binary PASS/FAIL matrix (all AC-MD-001 .. AC-MD-0NN)
- **E2**: Cross-platform build (`go build ./...` + `GOOS=windows GOARCH=amd64 go build ./...`)
- **E3**: Coverage (N/A — no Go source changes; report "no Go source modified")
- **E4**: Subagent boundary grep (N/A — no harness/hook code)
- **E5**: Lint status (`golangci-lint run` + `moai spec lint`)
- **E6**: Branch HEAD + push state (commit SHAs + push result)
- **E7**: Blocker report (if any)

Plus SPEC-specific verification:
- **E-MD-token**: `/context` before/after token delta (orchestrator-side, reported in progress.md §E.2)
- **E-MD-grep**: cut-line marker + 6-block skeleton + Pre-emit label grep survival (REQ-2)
- **E-MD-neutrality**: template-neutrality CI pass (REQ-1/REQ-2)
- **E-MD-memory**: MEMORY.md active-marker count unchanged + archive line-count delta (REQ-3)

---

## §F Milestones

### M1 — REQ-1: cadence-bridge.md path-scope (local + template)

**Files edited:**
- `internal/template/templates/.claude/rules/moai/workflow/cadence-bridge.md` (template-first)
- `.claude/rules/moai/workflow/cadence-bridge.md` (local mirror)

**Actions:**
1. Prepend YAML frontmatter (`description:` + `paths:`) to both files
2. Rewrite the "Loading scope" prose from "Intentionally always-loaded" → path-match declaration with rationale + trigger-condition list
3. Verify trigger-condition list includes `/loop` + `/moai` composition surface (REQ-MD-002)
4. `make build` to re-embed template
5. Verify byte-identical parity (frontmatter block identical between trees)

**Verification:** AC-MD-001 .. AC-MD-005

**Commit:** `feat(SPEC-MEMORY-DIET-001): M1 cadence-bridge path-scope + loading-scope rewrite`

### M2 — REQ-2: session-handoff.md illustrative content extraction (local + template)

**Files edited:**
- `internal/template/templates/.claude/rules/moai/workflow/session-handoff.md` (template-first)
- `.claude/rules/moai/workflow/session-handoff.md` (local mirror)
- `internal/template/templates/.claude/rules/moai/workflow/session-handoff-examples.md` (NEW sibling, template-first)
- `.claude/rules/moai/workflow/session-handoff-examples.md` (NEW sibling, local mirror)

**Actions:**
1. Create sibling references file `session-handoff-examples.md` (with path-scoped frontmatter pointing to session-handoff.md per REQ-MD-011)
2. Move the 2 Example sections (`### Example (Illustrative; substitute...)` + `### Example with Block 0 (Illustrative)`) to the sibling file
3. Replace each extracted section in session-handoff.md with a one-line pointer: `> Examples: see \`session-handoff-examples.md\` (path-scoped reference).`
4. Condense the Localization Table: keep en/ko columns inline, move ja/zh columns to the sibling file, add one-line pointer
5. Verify CORE DOCTRINE byte-identical survival: 6-block skeleton, Cut-line Marker Spec, Field-by-Field Spec, Pre-emit self-check labels, Auto-Memory Integration, Post-Paste /goal Follow-up Block, Diet Constraints, Worktree-Anchored Resume Pattern
6. `make build`
7. Verify parity (session-handoff.md content identical between trees; sibling file identical between trees)

**Verification:** AC-MD-006 .. AC-MD-011

**Commit:** `feat(SPEC-MEMORY-DIET-001): M2 session-handoff example extraction + localization condense`

### M3 — REQ-3: MEMORY.md archive pruning (auto-memory, no template)

**Files edited:**
- `~/.claude/projects/-Users-goos-MoAI-moai-adk-go/memory/MEMORY.md` (prune stable ✅ entries)
- `~/.claude/projects/-Users-goos-MoAI-moai-adk-go/memory/MEMORY-archive-2026-06-02.md` (append archived entries)

**Actions:**
1. Enumerate all `✅` entries in MEMORY.md
2. Classify each: stable (closed SPEC, no open follow-up, ≥ N days old) vs load-bearing (references pending next step / cross-Epic tracking / deferred debt)
3. Move stable entries to `MEMORY-archive-2026-06-02.md` (append, do not delete from archive)
4. Remove moved entries from MEMORY.md index
5. Verify active-marker entries (`🟢🟡🆕⏸️⚠️🔍`) count is unchanged
6. Verify no load-bearing `✅` entry was removed (grep for "다음=" / "handoff" / "deferred" / "debt" in the remaining index)

**Verification:** AC-MD-012 .. AC-MD-015

**Commit:** `feat(SPEC-MEMORY-DIET-001): M3 MEMORY.md archive pruning (N stable entries → archive)`

### M4 — Combined verification + neutrality gate

**Actions:**
1. Run `/context` (orchestrator-side) to measure combined token delta — verify ≥ ~5.5k reduction
2. Run `go test ./...` — verify all pass
3. Run `go vet ./...` — verify clean
4. Run template-neutrality CI guard — verify pass (no SPEC IDs / REQ tokens / dates / SHAs introduced)
5. Run grep survival checks: `✂` markers, 6-block skeleton headings, Pre-emit self-check labels
6. Verify cadence-bridge trigger surface includes `/loop` + `/moai` composition

**Verification:** AC-MD-016 .. AC-MD-017 + all grep ACs

**Commit:** (folded into M3 commit OR separate `test(SPEC-MEMORY-DIET-001): M4 combined verification` if deferred evidence)

---

## §G Anti-Patterns

- **AP-MD-001**: Scoping `session-handoff.md` out of always-loaded (REQ-2 extracts content but the doctrine STAYS always-loaded — it is KEEP-class per RULE-DIET-002).
- **AP-MD-002**: Deleting MEMORY.md entries instead of archiving (REQ-3 MUST move to `MEMORY-archive-2026-06-02.md`, never delete — archive preserves the audit trail per moai-memory.md § Memory Hygiene).
- **AP-MD-003**: Creating the sibling references file WITHOUT a `paths:` frontmatter (it would become a new always-loaded surface, defeating the diet — REQ-MD-011 binds).
- **AP-MD-004**: Editing only the local tree and forgetting the template mirror (CLAUDE.local.md §2 [HARD] Template-First Rule — must edit BOTH, template-first).
- **AP-MD-005**: Rewriting the "Loading scope" prose without a trigger-condition list (REQ-MD-003 requires the trigger surface to be documented so the gate does not silently drop the doctrine).
- **AP-MD-006**: Removing `ja`/`zh` Localization columns without a pointer to the full table in the sibling file (REQ-MD-007 requires the pointer).
- **AP-MD-007**: Pruning a `✅` entry that references a pending next step (REQ-MD-014 forbids — cross-Epic tracking is load-bearing despite the ✅ marker).

---

## §H Cross-References

- spec.md: `.moai/specs/SPEC-MEMORY-DIET-001/spec.md` (requirements + Prior-Art Review)
- acceptance.md: `.moai/specs/SPEC-MEMORY-DIET-001/acceptance.md` (AC matrix + Given-When-Then)
- RULE-DIET-002: `.moai/specs/SPEC-RULE-DIET-002/` (self-referential `paths:` mechanism precedent)
- RULES-PATH-SCOPE-001: `.moai/specs/SPEC-V3R6-RULES-PATH-SCOPE-001/` (original `paths:` frontmatter scoping)
- RULES-COMPRESS-001: `.moai/specs/SPEC-V3R6-RULES-COMPRESS-001/` (session-handoff.md prose compression baseline)
- CLAUDE.local.md §2 [HARD] Template-First Rule
- CLAUDE.local.md §25 Template Internal-Content Isolation (neutrality guard)
- `.claude/rules/moai/development/coding-standards.md` § Paths Frontmatter (CSV string syntax)
- `.claude/rules/moai/workflow/moai-memory.md` § Memory Hygiene (archive discipline)
