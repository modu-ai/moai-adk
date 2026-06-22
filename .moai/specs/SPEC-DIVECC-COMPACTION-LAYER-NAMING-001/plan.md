# Implementation Plan — SPEC-DIVECC-COMPACTION-LAYER-NAMING-001

> Tier S plan. Candidate N5 of Epic Dive-into-CC. DOC-ALIGNMENT only — records the paper's 5 compaction-layer names as cross-references in two existing workflow rule files.

## §A. Context

- **Tier**: S (doc/doctrine-only; 3 files; < 300 LOC prose; no Go, no behavior change).
- **Run-phase target #1 (REQUIRED, template-distributed)**: `.claude/rules/moai/workflow/context-window-management.md` — add a NEW cross-reference naming the five graduated-compaction layers `Budget Reduction → Snip → Microcompact → Context Collapse → Auto-Compact`, framed as Claude Code's layers that **moai-adk consumes (does not implement)**, citing the public paper (arXiv:2604.14228 / VILA-Lab). The file currently has ZERO compaction mention (verified plan-phase).
- **Run-phase target #2 (REQUIRED, template mirror of #1)**: `internal/template/templates/.claude/rules/moai/workflow/context-window-management.md` — apply the **byte-identical** edit. The mirror is currently byte-identical to local (`diff` exit 0, verified plan-phase). AC-CLN-004 requires it STAY identical.
- **Run-phase target #3 (REQUIRED, local-only)**: `.claude/rules/moai/workflow/runtime-recovery-doctrine.md` — ADDITIVE cross-reference adding the VILA-Lab paper as a CONVERGENT second source to book1 ch03, mapping its capitalized layer names onto the existing §1 lowercase sequence (`memory prefetch → snip → microcompact → context-collapse → autocompact`). NO mirror (this file is intentionally NOT template-distributed; it carries internal SPEC IDs).
- **PRESERVE**: in `runtime-recovery-doctrine.md`, the AP-RR-004 book1 named principles (`withheld-recoverable`, `cheapest-first`, `death-spiral`, `narrative consistency`, `hasAttemptedReactiveCompact`, `truncateHeadForPTLRetry`, `MAX_CONSECUTIVE_AUTOCOMPACT_FAILURES`) — all seven confirmed present plan-phase; the edit is ADD-only (REQ-CLN-007). In `context-window-management.md`, all existing sections (threshold table, /clear discipline, detection heuristics) preserved — the cross-reference is additive.

## §B. Known Issues (filtered to Tier S relevance)

- **B6 — spec-lint Out of Scope heading**: spec.md §F uses `### Out of Scope — <topic>` H3 sub-headings — satisfies `OutOfScopeRule`. (Plan-phase concern, handled.)
- **B10 — scope discipline**: touch ONLY the three §A files. Do NOT touch any other rule, the sibling N-candidate surfaces, hook scripts, skill bodies, plugins, or MCP config. Do NOT create a `runtime-recovery-doctrine.md` mirror.
- **Template neutrality (CLAUDE.local.md §15/§25)**: the `context-window-management.md` mirror must carry NO forbidden internal-content class. The paper citation (arXiv:2604.14228, VILA-Lab) is acceptable (public source); but NO internal `SPEC-DIVECC` ID / internal date / commit SHA / internal-only path may appear in that file or its mirror. AC-CLN-005 gates this via `TestTemplateNeutralityAudit` / `TestTemplateNoInternalContentLeak`.
- **Mirror-test allowlist nuance (honest disclosure)**: `context-window-management.md` is NOT in the byte-parity allowlist of `internal/template/rule_template_mirror_test.go`, so `TestRuleTemplateMirrorDrift` does NOT mechanically enforce local↔mirror parity for this file. The mirror IS still scanned by the neutrality tests. Therefore AC-CLN-004 (`diff` exit 0) is the binding parity check for the run-phase — the implementer MUST edit both trees identically by hand; no CI test will catch a parity miss on this specific file.
- **AP-RR-004 verbatim preservation**: the `runtime-recovery-doctrine.md` edit MUST be additive; the seven book1 named terms must survive verbatim (AC-CLN-006). Do NOT rephrase the existing §1 lowercase sequence — ADD the capitalized mapping alongside it.

## §C. Pre-flight (run-phase)

```bash
# 1. Confirm CWM still has ZERO compaction mention before editing (anchor unchanged)
grep -niE 'snip|microcompact|compaction|budget reduction|context collapse|auto-compact|VILA|2604.14228' .claude/rules/moai/workflow/context-window-management.md || echo "ZERO (expected — clean anchor)"
# 2. Confirm CWM local ↔ mirror still byte-identical before editing
diff .claude/rules/moai/workflow/context-window-management.md internal/template/templates/.claude/rules/moai/workflow/context-window-management.md && echo "IDENTICAL (expected)"
# 3. Confirm runtime-recovery-doctrine.md §1 lowercase sequence + AP-RR-004 terms present
grep -nF 'memory prefetch → snip → microcompact → context-collapse → autocompact' .claude/rules/moai/workflow/runtime-recovery-doctrine.md
for t in "withheld-recoverable" "cheapest-first" "death-spiral" "narrative consistency" "hasAttemptedReactiveCompact" "truncateHeadForPTLRetry" "MAX_CONSECUTIVE_AUTOCOMPACT_FAILURES"; do grep -qF "$t" .claude/rules/moai/workflow/runtime-recovery-doctrine.md || echo "MISSING BASELINE: $t"; done
# 4. Confirm runtime-recovery-doctrine.md has NO mirror (must stay that way)
ls internal/template/templates/.claude/rules/moai/workflow/runtime-recovery-doctrine.md 2>/dev/null && echo "UNEXPECTED MIRROR" || echo "NO MIRROR (expected)"
```

## §D. Constraints (DO NOT VIOLATE)

- Doc-only. No Go code change, no hook/skill/plugin/MCP behavior change, no compaction/recovery runtime change (moai-adk consumes CC compaction; cannot alter it).
- The five layer names are the paper's naming — record as provenance citation, framed consume-not-implement; NEVER assert moai-adk implements or measures them (REQ-CLN-003).
- `context-window-management.md` edit MUST be mirrored byte-identically; the mirror MUST pass neutrality (no internal `SPEC-DIVECC` ID / date / SHA / internal path) — REQ-CLN-004/005.
- `runtime-recovery-doctrine.md` edit MUST be additive — the seven AP-RR-004 named terms survive verbatim; NO mirror is created (REQ-CLN-007/008).
- Touch ONLY the three §A files (+ spec.md frontmatter status transition). REQ-CLN-008.
- Conventional Commits; `🗿 MoAI` trailer; NO `--no-verify` / `--amend` / force-push.

## §E. Self-Verification (run-phase deliverable)

Run-phase manager-develop reports the AC-CLN-001..008 PASS/FAIL matrix with verbatim grep/diff output, plus:
- `diff <CWM-local> <CWM-mirror>; echo exit=$?` → `exit=0` (AC-CLN-004).
- `go test ./internal/template/... -run 'TestTemplateNeutralityAudit|TestTemplateNoInternalContentLeak'` PASS (AC-CLN-005).
- The seven AP-RR-004 verbatim-term survival loop (AC-CLN-006) → no output.
- `git show --stat <run-commit>` confirming changed files ⊆ {CWM local, CWM mirror, runtime-recovery-doctrine.md local, spec.md} and that no runtime-recovery-doctrine.md mirror was created (AC-CLN-008).

## §F. Milestones (priority-ordered, no time estimates)

- **M1 — Add the 5-layer cross-reference to `context-window-management.md`**: a short cross-reference block naming `Budget Reduction → Snip → Microcompact → Context Collapse → Auto-Compact`, framed as "Claude Code's graduated-compaction layers (which moai-adk consumes, does not implement)", citing arXiv:2604.14228 / VILA-Lab. Satisfies REQ-CLN-001..003.
- **M2 — Mirror + verify**: apply the byte-identical edit to the template mirror; `diff` exit 0 (AC-CLN-004); run `make build` (sanity); run `TestTemplateNeutralityAudit` + `TestTemplateNoInternalContentLeak` (AC-CLN-005). Satisfies REQ-CLN-004/005.
- **M3 — Add the convergent-second-source cross-reference to `runtime-recovery-doctrine.md`**: ADD a note (near §1 or §5 Cross-References) that the VILA-Lab paper (arXiv:2604.14228) is a CONVERGENT second source to book1 ch03 for the same graduated-compaction concept, mapping the capitalized names onto the existing lowercase sequence. ADDITIVE only — AP-RR-004 terms untouched. Satisfies REQ-CLN-006/007.
- **M4 — Self-verify + commit + push**: run the AC-CLN grep/diff matrix + the AP-RR-004 survival loop; commit (`docs(SPEC-DIVECC-COMPACTION-LAYER-NAMING-001): M1 5-layer compaction naming cross-reference`), push to main (Hybrid Trunk Tier S).

## §G. Anti-Patterns to avoid

- Implying moai-adk implements the five layers (violates REQ-CLN-003 — moai-adk CONSUMES CC compaction).
- Editing CWM local without the byte-identical mirror edit (violates AC-CLN-004; no CI test catches it for this file — the implementer must do it by hand).
- Rephrasing or removing the existing §1 lowercase sequence / any AP-RR-004 named term in `runtime-recovery-doctrine.md` (violates REQ-CLN-007 — additive only).
- Creating a `runtime-recovery-doctrine.md` template mirror (violates REQ-CLN-008 — it is intentionally local-only).
- Leaking an internal `SPEC-DIVECC` ID / internal date / commit SHA into the CWM mirror (violates REQ-CLN-005 — neutrality).

## §H. Cross-References

- spec.md §B (grounding + pinned targets), §C (REQ-CLN-001..008), §D (AC matrix), §G (cross-references).
- `.moai/specs/SPEC-DIVECC-HOOK-FAILURE-MODE-AUDIT-001/ROADMAP.md` §N5 — Epic candidate detail.
- `.claude/rules/moai/workflow/runtime-recovery-doctrine.md` §1 (existing book1 lowercase sequence) + AP-RR-004 (verbatim-term preservation).
- `.moai/docs/template-internal-isolation-doctrine.md` — neutrality gate for the CWM mirror edit (M2).
- `internal/template/rule_template_mirror_test.go` — byte-parity allowlist (CWM is NOT in it; AC-CLN-004 `diff` is the binding parity check).
