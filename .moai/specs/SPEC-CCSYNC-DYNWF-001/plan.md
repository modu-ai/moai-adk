# Implementation Plan — SPEC-CCSYNC-DYNWF-001

## A. Context

Documentation-only SPEC (Tier S, minimal). Four prose seams to close across four template-distributed documents.
No Go code, no behavior change. The run-phase agent is `manager-docs` (documentation edits) or `manager-develop`
in autofix cycle — the orchestrator routes per the plan-auditor verdict and GATE-2. Because every target file
except local-class files is template-distributed, the dominant risk is template-neutrality and mirror-parity, not
logic correctness.

## B. Known Issues / Pre-existing State

- REQ-2's `/deep-research` bullet and REQ-4's `ultracode` bullet ALREADY EXIST in partial form in
  `dynamic-workflows.md § MoAI Integration Notes` (the `Bundled /deep-research` and `ultracode effort` bullets).
  Run-phase work for REQ-2 (the dynamic-workflows side) and REQ-4 is therefore primarily **augmentation** of
  existing bullets, not net-new sections. The NEW cross-references for REQ-2 land in `CLAUDE.md § 10` and the
  `moai-domain-research` skill body, which currently have NO `/deep-research` mention.
- `CLAUDE.md § 10` is byte-identical between the working copy and its template mirror (verified at plan time).
  The run-phase edit must keep them identical.
- The `moai-domain-research` skill body and its template mirror are byte-identical (9761 bytes each at plan time).

## C. Pre-flight (run-phase entry checks)

Before editing, the run-phase agent re-confirms the four anchor headings still exist (they may drift between plan
and run):

```bash
grep -n "## How a Workflow Runs"            .claude/rules/moai/workflow/dynamic-workflows.md
grep -n "## When to Use a Dynamic Workflow" .claude/rules/moai/workflow/dynamic-workflows.md
grep -n "## MoAI Integration Notes"         .claude/rules/moai/workflow/dynamic-workflows.md
grep -n "## 10. Web Search Protocol"        CLAUDE.md
grep -n "## Works Well With"                .claude/skills/moai-domain-research/SKILL.md
```

If any anchor is missing, return a blocker report rather than guessing a replacement location.

## D. HARD Constraints (run-phase obligations)

### D.1 Template Neutrality [HARD]

Every target file except local (`CLAUDE.local`-class) files is TEMPLATE-DISTRIBUTED. Content added in the file
BODIES MUST be 16-language-neutral per `CLAUDE.local.md § 15` and `§ 25`:

- NO internal SPEC IDs (e.g., no `SPEC-CCSYNC-DYNWF-001` token in the prose body of any template-distributed file)
- NO REQ/AC tokens in the body
- NO audit citations ("Audit N Finding AX")
- NO internal work dates, commit SHAs
- NO macOS-bias absolute paths (`/Users/...`)
- NO `CLAUDE.local.md` references in the body

The added prose describes the Dynamic Workflows MECHANISM in generic terms — this is acceptable content
(mechanism explanation + public Claude Code feature citation), which the neutrality guard permits.

### D.2 Template-First Mirror [HARD]

Every edit to a template-managed file MUST be applied in BOTH locations, then `make build`:

| Working copy | Template mirror |
|--------------|-----------------|
| `.claude/rules/moai/workflow/dynamic-workflows.md` | `internal/template/templates/.claude/rules/moai/workflow/dynamic-workflows.md` |
| `CLAUDE.md` | `internal/template/templates/CLAUDE.md` |
| `.claude/skills/moai-domain-research/SKILL.md` | `internal/template/templates/.claude/skills/moai-domain-research/SKILL.md` |
| `.claude/rules/moai/workflow/goal-directive.md` (only if REQ-4 lands here) | `internal/template/templates/.claude/rules/moai/workflow/goal-directive.md` |

Run-phase sequence per file: edit working copy → edit template mirror with identical content → after all edits,
run `make build` to regenerate `internal/template/embedded.go`. Acceptance includes a mirror-parity check
(`go test ./internal/template/...` mirror-drift test green).

### D.3 Documentation-only [HARD]

Acceptance criteria are grep/content checks on the target files PLUS `go test ./internal/template/...`
(mirror-drift + neutrality tests) staying green. No new Go test for behavior is added.

### D.4 Anchor on headings, not line numbers [HARD]

All run-phase edits and all acceptance checks anchor on section HEADINGS (`## How a Workflow Runs`,
`## When to Use a Dynamic Workflow`, `## MoAI Integration Notes`, `## 10. Web Search Protocol`,
`## Works Well With`), never on line numbers. Line numbers drift between plan and run.

## E. Self-Verification (run-phase exit)

The read-only verification batch (single-turn multi-Bash) at run-phase completion:

```bash
# 1. Mirror-drift + neutrality + leak tests
go test ./internal/template/...

# 2. REQ-1 determinism note present in both copies
grep -n "determinist\|wall-clock\|random" .claude/rules/moai/workflow/dynamic-workflows.md
grep -n "determinist\|wall-clock\|random" internal/template/templates/.claude/rules/moai/workflow/dynamic-workflows.md

# 3. REQ-2 /deep-research wired in CLAUDE.md §10 and research skill
grep -n "deep-research" CLAUDE.md
grep -n "deep-research" .claude/skills/moai-domain-research/SKILL.md

# 4. REQ-3 routing heuristic present
grep -n "dozens-to-hundreds\|sequential subagents" .claude/rules/moai/workflow/dynamic-workflows.md

# 5. REQ-4 ultracode resume pairing present
grep -n "ultracode" .claude/rules/moai/workflow/dynamic-workflows.md

# 6. Neutrality: no internal SPEC ID leaked into any template body
grep -rn "SPEC-CCSYNC-DYNWF" internal/template/templates/ || echo "clean — no leak"
```

## F. Milestones (priority-based, no time estimates)

| Milestone | Scope | REQ | Priority |
|-----------|-------|-----|----------|
| M1 | REQ-1 determinism note in `§ How a Workflow Runs` (working + mirror) | REQ-1 | High |
| M2 | REQ-3 routing heuristic in/under `§ When to Use a Dynamic Workflow` (working + mirror) | REQ-3 | High |
| M3 | REQ-4 `ultracode` resume pairing note in `§ MoAI Integration Notes` (working + mirror) | REQ-4 | Medium |
| M4 | REQ-2 `/deep-research` cross-ref in `CLAUDE.md § 10` (working + mirror) | REQ-2 | High |
| M5 | REQ-2 `/deep-research` cross-ref in `moai-domain-research` skill body (working + mirror) | REQ-2 | High |
| M6 | `make build` + run-phase verification batch (§ E) | all | High |

REQ-4 lands in `dynamic-workflows.md § MoAI Integration Notes` (preferred, augmenting the existing `ultracode`
bullet) rather than `session-handoff.md`, to keep the `ultracode` discussion co-located with the workflow primitive
it belongs to. The SPEC permits either location; this plan selects dynamic-workflows.md.

## G. Anti-Patterns to Avoid

- **AP-1**: Editing only the working copy and forgetting the template mirror → mirror-drift test RED.
- **AP-2**: Leaking the internal SPEC ID / REQ tokens into a template-distributed body → neutrality test RED.
- **AP-3**: Restructuring sections or renaming headings → breaks existing cross-references; additions only.
- **AP-4**: Duplicating the `/deep-research` and `ultracode` bullets in `dynamic-workflows.md` instead of
  augmenting the existing ones → content drift.
- **AP-5**: Anchoring acceptance on line numbers → false failures after drift.
- **AP-6**: Sweeping pre-existing unrelated working-tree changes (docs-site, CHANGELOG, README.ko.md) into the
  plan-phase or run-phase commit → use explicit-path `git add` only.

## H. Cross-References

- `CLAUDE.local.md § 15` (16-language template neutrality), `§ 25` (template internal-content isolation)
- `.moai/docs/template-internal-isolation-doctrine.md` (forbidden/allowed content-class catalogue)
- `internal/template/internal_content_leak_test.go` + `.github/workflows/template-neutrality-check.yaml` (CI guard)
- `.claude/rules/moai/development/spec-frontmatter-schema.md` § Status Transition Ownership Matrix
- Public Claude Code workflows reference: https://code.claude.com/docs/en/workflows
