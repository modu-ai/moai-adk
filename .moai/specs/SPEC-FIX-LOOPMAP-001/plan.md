# SPEC-FIX-LOOPMAP-001 — Implementation Plan

## §A Context

### §A.1 Problem summary

`/moai fix` Phase 4 (Verification) is prose-only ("Confirm ... Detect ...") with no evidence binding, no baseline-comparable regression guard, and no residue persistence on unresolved exit. Neither fix.md nor loop.md documents where the workflow sits in the turn-based/goal-based/time-based/proactive loop taxonomy. This is the L5 audit finding SPEC-LOOP-VERDICT-CONTRACT-001 deliberately deferred (spec.md §Out of Scope).

### §A.2 Evidence baselines (measured 2026-07-09 by this agent via Bash/Read, vci §2 attribution)

```
$ grep -n "^## Phase 4" .claude/skills/moai/workflows/fix.md
189:## Phase 4: Verification
195:## Phase 4.5: MX Tag Update
234:## Phase 4.6: Dead Code Cleanup (Optional)

$ sed -n '189,193p' .claude/skills/moai/workflows/fix.md
## Phase 4: Verification

- Re-run affected diagnostics on modified files
- Confirm fixes resolved the targeted issues
- Detect any regressions introduced by fixes

$ grep -n "^## " .claude/skills/moai/workflows/loop.md
36:## Invocation Routes
48:## Relationship to the Pipeline-Level Agentic Completion Loop
52:## Supported Flags
62:## Per-Iteration Cycle
157:## Completion Conditions
171:## MX Tag Integration
196:## Snapshot Management
213:## Language-Specific Commands
243:## Cancellation
247:## Safe Development Protocol
254:## Execution Summary

$ diff .claude/skills/moai/workflows/fix.md internal/template/templates/.claude/skills/moai/workflows/fix.md
(no output — byte-identical)
$ diff .claude/skills/moai/workflows/loop.md internal/template/templates/.claude/skills/moai/workflows/loop.md
(no output — byte-identical)

$ wc -l .claude/skills/moai/workflows/fix.md .claude/skills/moai/workflows/loop.md
     320 fix.md
     267 loop.md
```

SPEC-LOOP-VERDICT-CONTRACT-001 plan.md §D D2 (verdict schema, cited verbatim):
> "Schema minimum: `spec_or_scope`, `exit_kind` (ceiling|manual-residue), `iterations_used`, `ceiling_applied` + its source (flag|ralph|loop_prevention), `conditions` final state, `remaining_issues[]` ({severity, description, file, suggested_action}), `vci_report_ref`, `created_at`."

SPEC-LOOP-VERDICT-CONTRACT-001 plan.md §G (sibling scope boundary, cited verbatim):
> "Touching fix.md Phase 4 (L5 is explicitly out of scope — record, don't remediate)."

### §A.3 Approach — three milestones, doc-primary

1. **M1 — Phase 4 evidence + regression guard** — rewrite fix.md Phase 4 into a claim/evidence contract with a full-rescan-vs-baseline regression guard.
2. **M2 — residue persistence + escalation recommendation** — new Phase 4.7 subsection: persist unresolved residue to the loop-verdict schema (`exit_kind: "one-shot-residue"`) and recommend (never auto-invoke) `/moai loop`.
3. **M3 — Loop Taxonomy Position + template sync** — add the ≤15-line section to fix.md immediately; add the same section to loop.md ONLY after SPEC-LOOP-VERDICT-CONTRACT-001's loop.md rewrite has landed (landing-order gate, §D D1); `make build` template sync for both files.

### §A.4 Tier evidence (S)

- Files touched: 2 (`fix.md`, `loop.md`) + their 2 template mirrors = 4 files total, all markdown, no Go code.
- LOC delta estimate: fix.md +40-60 lines (Phase 4 rewrite + Phase 4.7 + Loop Taxonomy Position); loop.md +12-15 lines (Loop Taxonomy Position only, additive).
- No new Go types, no new config keys, no new CLI subcommands.
- Tier S table (`spec-workflow.md` § SPEC Complexity Tier): < 300 LOC, < 5 files → **S**.

### §A.5 PRESERVE / EXTEND map

| File | Action | Scope |
|------|--------|-------|
| `.claude/skills/moai/workflows/fix.md` | EXTEND | Phase 4 rewrite (evidence + regression guard), new Phase 4.7 (residue persistence + escalation), new Loop Taxonomy Position section |
| `.claude/skills/moai/workflows/loop.md` | EXTEND (additive only, gated) | New Loop Taxonomy Position section ONLY — no edits to Steps 1/4/9 or § Completion Conditions (SPEC-LOOP-VERDICT-CONTRACT-001 scope) |
| `internal/template/templates/.claude/skills/moai/workflows/{fix,loop}.md` | EXTEND (mirror) | Identical edits, template-first or synced same-commit |
| `.claude/skills/moai/workflows/moai.md` | PRESERVE | untouched |
| `.moai/config/sections/{ralph,workflow}.yaml` | PRESERVE | untouched |
| `.claude/skills/moai/workflows/{gate,review}.md` | PRESERVE | untouched |
| Phase 3 static Level→agent dispatch table (fix.md lines 176-179) | PRESERVE | untouched — Agentless contract |

---

## §B Known Issues (filtered, Tier S — doc-primary)

- **B4 (Frontmatter Canonical Schema)** — `created:`/`updated:`/`tags:` used; no snake_case aliases; `depends_on: [SPEC-LOOP-VERDICT-CONTRACT-001]` optional field included.
- **B6 (spec-lint Heading Convention)** — Out of Scope uses `### Out of Scope — <topic>` H3 sub-headings with `-` bullets (4 topics).
- **B10 (Untouched Paths PRESERVE)** — `moai.md`, `ralph.yaml`, `workflow.yaml`, `gate.md`, `review.md` MUST NOT be touched; loop.md touch is additive-only and landing-order-gated.
- **Agentless-contract regression risk (custom, fix.md-specific)** — any new Phase 4/4.7 text MUST NOT match `forbiddenControlFlowPatterns` in `internal/template/agentless_audit_test.go` (no "Use the X subagent to decide/determine/choose/select/orchestrate/route/dispatch" phrasing). Run `TestAgentlessUtilityNoLLMControlFlow` before considering M1/M2 complete (AC-FLM-010).

---

## §C Pre-flight checklist

```bash
# 1. Branch + baseline
git branch --show-current
git rev-parse HEAD

# 2. live == template mirror (must be empty diff before AND after edits, modulo the intended change)
diff .claude/skills/moai/workflows/fix.md internal/template/templates/.claude/skills/moai/workflows/fix.md
diff .claude/skills/moai/workflows/loop.md internal/template/templates/.claude/skills/moai/workflows/loop.md

# 3. Landing-order gate for M3's loop.md touch — verify SPEC-LOOP-VERDICT-CONTRACT-001's loop.md rewrite
#    has landed before adding the Loop Taxonomy Position section to loop.md:
grep -n "sync_commit_sha" .moai/specs/SPEC-LOOP-VERDICT-CONTRACT-001/progress.md
git log --oneline --all -- .claude/skills/moai/workflows/loop.md | head -5

# 4. Agentless-contract sanity baseline (pre-edit; expect PASS)
go test -run TestAgentlessUtilityNoLLMControlFlow ./internal/template/...

# 5. This SPEC's own artifacts lint clean
moai spec lint .moai/specs/SPEC-FIX-LOOPMAP-001/spec.md .moai/specs/SPEC-FIX-LOOPMAP-001/plan.md
```

If pre-flight check 3 shows SPEC-LOOP-VERDICT-CONTRACT-001's loop.md rewrite has NOT landed, M1 and M2 (fix.md-only) MAY still proceed (no dependency — §D D1); only M3's loop.md half is blocked until the gate clears.

---

## §D Constraints + open decisions

Constraints: see spec.md § Constraints (Agentless preservation, 3-phase contract preservation, loop.md additive-only + landing-order gate, no Go code, scope discipline).

Open decisions (run-phase discretion unless marked):

1. **D1 — loop.md landing-order gate mechanics** (Constraint 3): the run-phase implementer of THIS SPEC checks SPEC-LOOP-VERDICT-CONTRACT-001's `progress.md §E.4` for a populated `sync_commit_sha`, OR `git log` for its loop.md-touching commits, before adding the Loop Taxonomy Position section to loop.md. If the gate has not cleared, M3 splits: fix.md's Loop Taxonomy Position section lands now (no dependency), loop.md's counterpart is deferred and recorded as an explicit residual milestone in this SPEC's own progress.md.
2. **D2 — `exit_kind` enum extension surface** (REQ-FLM-003): this SPEC extends the `ceiling|manual-residue` enum with `one-shot-residue` by documenting it inline in fix.md's own Phase 4.7 text (citing the base schema by SPEC-ID + plan.md §D reference). This SPEC does NOT edit SPEC-LOOP-VERDICT-CONTRACT-001's artifacts (frozen, sibling-owned scope). If a future shared doctrine location for the verdict schema emerges, back-porting the enum value there is a follow-up concern, not blocking here.
3. **D3 — Phase 4.7 placement** (REQ-FLM-003/004): inserted between the existing `Phase 4.6: Dead Code Cleanup (Optional)` and `## Task Tracking` sections, matching the existing Phase 4.x sub-numbering convention (4, 4.5, 4.6, **4.7**) rather than folding residue-persistence into Phase 4 itself — keeps Phase 4 focused on verification per REQ-FLM-001/002.

---

## §E Self-Verification (run-phase deliverables)

Per `manager-develop-prompt-template.md` §E, vci 5-section format each:
- **E1**: AC matrix (spec.md §3, 10 rows) with verbatim grep outputs.
- **E2**: `make build` green (template edits embedded); `go build ./...` unaffected.
- **E3**: n/a for coverage (doc-primary); `go test -run TestAgentlessUtilityNoLLMControlFlow ./internal/template/...` green (AC-FLM-010) is the relevant Go-side check, not a coverage percentage.
- **E4**: no `AskUserQuestion` additions to skill bodies — `grep -rn 'AskUserQuestion' .claude/skills/moai/workflows/{fix,loop}.md` returns 0 new instances.
- **E5**: template-neutrality — `grep -rn "SPEC-FIX-LOOPMAP" internal/template/templates/` returns 0 matches (AC-FLM-009).
- **E6**: commit SHAs + push state (Route A main-direct, per Tier S / M default).
- **E7**: blocker report if live/template pre-existing divergence found on `fix.md`/`loop.md` (§C check 2), or if the M3 landing-order gate (§D D1) has not cleared at run-phase entry.

---

## §F Milestones (priority-ordered; no time estimates)

| Milestone | Scope | REQs | Exit criterion |
|-----------|-------|------|-----------------|
| M1 — Phase 4 evidence + regression guard | fix.md Phase 4 rewrite: claim/evidence table, full re-scan + baseline diff, revert-or-report-failed regression handling | REQ-FLM-001, REQ-FLM-002 | AC-FLM-001..004 PASS |
| M2 — residue persistence + escalation recommendation | fix.md new Phase 4.7: loop-verdict schema write (`exit_kind: "one-shot-residue"`), `/moai loop` recommendation, Repeatability clause preserved verbatim | REQ-FLM-003, REQ-FLM-004 | AC-FLM-005, AC-FLM-006 PASS |
| M3 — Loop Taxonomy Position + template sync | fix.md section (immediate); loop.md section (landing-order gated per §D D1); template-first sync for both | REQ-FLM-005, REQ-FLM-006 | AC-FLM-007..010 PASS |

Ordering rationale: M1 establishes the evidence contract M2's escalation report references ("regression-guard failure" as a residue trigger); M3 is independent documentation work sequenced last so the landing-order gate has maximum time to clear.

---

## §G Anti-Patterns (do NOT)

- Touching `moai.md`, `ralph.yaml`, `workflow.yaml`, `gate.md`, or `review.md` — out of scope discipline (spec.md Constraint 5).
- Adding the Loop Taxonomy Position section to loop.md BEFORE SPEC-LOOP-VERDICT-CONTRACT-001's loop.md rewrite lands — risks a diff/merge collision on a file mid-rewrite by a sibling SPEC (Constraint 3).
- Modifying the Phase 3 static Level→agent dispatch table (fix.md lines 176-179) — breaks the Agentless contract regression guard `TestAgentlessUtilityNoLLMControlFlow`.
- Introducing "Use the X subagent to decide/determine/..." or "delegate to ... orchestrator/router/dispatcher" phrasing anywhere in fix.md — trips `forbiddenControlFlowPatterns`.
- Spawning an independent-verifier `Agent()` for fix's Phase 4 re-scan — explicitly out of scope (spec.md § Out of Scope — independent-agent verification spawn); the one-shot pipeline's independence is mechanical (re-run + diff), not agent-based.
- Editing SPEC-LOOP-VERDICT-CONTRACT-001's own artifacts to backfill the `exit_kind` enum extension — that sibling's SPEC directory is frozen/closed scope for this SPEC (D2).
- Writing this SPEC's ID into template-tree files (neutrality guard, AC-FLM-009).

---

## §H Cross-References

- spec.md (SSOT — REQ/AC matrix), progress.md (§E skeleton).
- SPEC-LOOP-VERDICT-CONTRACT-001 spec.md (REQ-LVC-005) + plan.md (§D D2 schema, §G scope boundary).
- `verification-claim-integrity.md` §1.1 + §3 (evidence-bearing format source).
- `internal/template/agentless_audit_test.go` (`TestAgentlessUtilityNoLLMControlFlow`, `forbiddenControlFlowPatterns`).
- `.claude/rules/moai/workflow/spec-workflow.md#subcommand-classification` (Agentless pipeline contract).
