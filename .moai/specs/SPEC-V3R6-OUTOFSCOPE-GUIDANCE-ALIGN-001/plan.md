# Implementation Plan — SPEC-V3R6-OUTOFSCOPE-GUIDANCE-ALIGN-001

> Tier S chore. Prose-only guidance alignment. Go change = 0.
> Acceptance criteria live inline in `spec.md` §D (Tier S LEAN 2-artifact).

## §A. Context

Three SPEC-authoring guidance surfaces instruct authors to write an exclusions heading
shape (`## Exclusions (What NOT to Build)`, "Exclusions", "Non-goals") that the enforced
`OutOfScopeRule` lint (`internal/spec/lint.go`) rejects, because the rule requires a literal
"out of scope" substring + an `###` "Out of Scope" heading + ≥1 `-` bullet. The fix aligns
the three guidance surfaces TO the rule. The rule and `internal/spec/CLAUDE.md` are the
unchanged SSOT. See spec.md §A for full drift evidence (347/353 specs already patched
per-SPEC at the symptom).

## §B. Known Issues / Risks

- **Mirror parity risk**: each `.claude/...` edit MUST be mirrored to its
  `internal/template/templates/.claude/...` counterpart or the Template-First contract
  (CLAUDE.local.md §2) is violated and CI may flag drift. Mitigation: edit source + mirror
  as a paired unit per file (milestones M1–M3 below each cover the pair).
- **Wording-only risk**: the change is prose; there is no behavior to test beyond grep
  presence + the optional CI guard. Mitigation: AC-OSG-001..004 are grep-verifiable;
  AC-OSG-007 is `moai spec lint` self-verification.
- **Line-number drift**: the cited line numbers (manager-spec:84, plan-auditor:~259,
  SKILL.md:~263/~324) are plan-phase snapshots. Run-phase MUST locate the text by content
  (grep) rather than trusting the line number, since prior edits may have shifted lines.
- **Template-neutrality**: the template mirror edits add the generic phrase "Out of Scope"
  and a generic `### Out of Scope — <topic>` example. These are mechanism descriptions /
  generic prose (kept content classes per CLAUDE.local.md §25 / coding-standards.md § MUST);
  do NOT introduce internal SPEC IDs, REQ tokens, dates, or commit SHAs into the mirrors.

## §C. Pre-flight (run-phase entry checks)

1. `git rev-parse --abbrev-ref HEAD` — confirm working branch.
2. Confirm the 3 source files + 3 mirror files exist (verified at plan-phase; all 6 present).
3. `grep -n "Exclusions (What NOT to Build)" .claude/agents/moai/manager-spec.md` — locate
   the exact current text before editing (do not trust line 84 verbatim).
4. `moai spec lint .moai/specs/SPEC-V3R6-OUTOFSCOPE-GUIDANCE-ALIGN-001/spec.md` — 0 errors
   (this SPEC's own self-pass; already confirmed at plan-phase).

## §D. Constraints

- Go change = 0 for the core fix (M1–M4). The optional CI guard (M5) is the only Go that
  may be touched, and only if elected.
- Do NOT edit `internal/spec/lint.go` or `internal/spec/CLAUDE.md` — they are the SSOT.
- Do NOT touch any `spec.md` other than this SPEC's own.
- Template-First: every `.claude/...` edit is mirrored to
  `internal/template/templates/.claude/...`, then `make build` regenerates `embedded.go`.
- Status transitions: `draft → in-progress` on M1 first commit (manager-develop owns);
  `in-progress → implemented` at sync (manager-docs owns). manager-spec owns only the
  `(none) → draft` already set in this artifact set.

## §E. Self-Verification (run-phase deliverables expected from manager-develop)

- E1: AC matrix AC-OSG-001..007 PASS/FAIL with grep evidence per AC.
- E2: cross-file mirror parity diff (source vs template mirror) for each of the 3 files.
- E3: `make build` exit 0 + `embedded.go` regeneration evidence.
- E4: `moai spec lint` re-run on this SPEC (still 0 errors post-edit, since this SPEC's
  spec.md is untouched by the run-phase).
- E5: optional CI guard test result (if M5 elected).

## §F. Milestones (priority-ordered, no time estimates)

### F.1 M1 — Align manager-spec (source + mirror) [REQ-OSG-001, REQ-OSG-005]

- Edit `.claude/agents/moai/manager-spec.md`: replace the `## Exclusions (What NOT to Build)`
  H2-only instruction with guidance directing authors to include at least one
  `### Out of Scope — <topic>` H3 sub-heading carrying `-` bullet item(s), preserving the
  "What NOT to Build" intent and the "at least one entry" requirement.
- Mirror the identical change to
  `internal/template/templates/.claude/agents/moai/manager-spec.md`.
- **Run-phase warning (D1, carried from spec.md AC-OSG-001)**: the implementer MUST NOT leave
  the literal bullet-mandate string
  `- [HARD] Every spec.md MUST include \`## Exclusions (What NOT to Build)\`` as a contrast
  example in the edited guidance, or AC-OSG-001's mandate-scoped negative grep
  (`grep -n '^- \[HARD\] Every spec.md MUST include .## Exclusions'`) false-fails. If a contrast
  example is desired, phrase it WITHOUT the verbatim bullet-mandate prefix.
- **D2 note (positive grep target)**: the AC-OSG-001 positive grep is `grep -c '### Out of Scope'`
  (H3-prefix anchored), NOT a case-insensitive `out of scope` count — the pre-existing
  `OUT OF SCOPE:` delegation-scope line at manager-spec.md L77 would otherwise false-PASS a bare
  token count. Do NOT remove or alter the L77 delegation-scope line (out of scope for this SPEC).
- Verifies AC-OSG-001 (source) + contributes to AC-OSG-005 (mirror).

### F.2 M2 — Align plan-auditor (source + mirror) [REQ-OSG-002, REQ-OSG-005]

- Edit `.claude/agents/moai/plan-auditor.md`: align SC-6 (~line 259) and the related
  references (~52 / ~91 / ~147) to describe the exclusions check as an "Out of Scope" H3
  sub-heading with ≥1 bullet, matching `OutOfScopeRule`.
- Mirror to `internal/template/templates/.claude/agents/moai/plan-auditor.md`.
- Verifies AC-OSG-002 (source) + contributes to AC-OSG-005 (mirror).

### F.3 M3 — Align moai-workflow-spec SKILL.md (source + mirror) [REQ-OSG-003, REQ-OSG-005]

- Edit `.claude/skills/moai-workflow-spec/SKILL.md`: align the `### Exclusion Rules` block
  (~263) and the verification-checklist item "Non-goals section present" (~324) to reference
  the `### Out of Scope — <topic>` convention.
- Mirror to `internal/template/templates/.claude/skills/moai-workflow-spec/SKILL.md`.
- Verifies AC-OSG-003 (source) + contributes to AC-OSG-005 (mirror).

### F.4 M4 — make build + intent review [REQ-OSG-004, REQ-OSG-005]

- Run `make build` to regenerate `internal/template/embedded.go` from the 3 mirror edits.
- Review the full diff to confirm intent preservation (AC-OSG-004): each surface still
  requires ≥1 exclusion entry and still conveys "what NOT to build".
- Verifies AC-OSG-004 + completes AC-OSG-005.

### F.5 M5 — Optional CI re-drift guard (SHOULD) [REQ-OSG-006]

- If elected: add a Go test (under `internal/template/`, alongside the existing neutrality /
  audit tests) asserting the manager-spec guidance string contains the `### Out of Scope`
  form, so future edits cannot silently regress to a bare `## Exclusions` H2.
- Verifies AC-OSG-006. If deferred, record AC-OSG-006 as PASS-WITH-DEBT with a follow-up
  note; it does NOT block AC-OSG-001..005 or SPEC close.

## §G. Anti-Patterns (do NOT do)

- Editing `internal/spec/lint.go` to "loosen" the rule instead of fixing the guidance —
  the rule is the correct SSOT (out of scope per spec.md §B.1).
- Broad-sweeping every doc mentioning "Exclusions" — only the 3 named surfaces are in scope.
- Editing the 6 currently non-compliant downstream `spec.md` files as part of this SPEC —
  retroactive remediation is out of scope.
- Editing source without the mirror (or vice versa) — Template-First requires the pair.
- Trusting the plan-phase line numbers verbatim — locate text by grep before editing.

## §H. Cross-References

- `internal/spec/lint.go` `OutOfScopeRule` — the enforced SSOT (unchanged).
- `internal/spec/CLAUDE.md` § Heading convention — canonical `### Out of Scope — <topic>` form.
- `.claude/rules/moai/development/spec-frontmatter-schema.md` § Status Transition Ownership Matrix.
- `CLAUDE.local.md` §2 Template-First Rule / §25 Template Internal-Content Isolation.
- spec.md §A (drift evidence), §C (REQ-OSG-001..007), §D (AC-OSG-001..007).
