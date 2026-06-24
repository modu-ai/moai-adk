# Implementation Plan — SPEC-V3R6-RULES-SSOT-DEDUP-001

## §A. Context

Tier L SSOT de-duplication + structural consolidation across `.claude/rules/moai/`, with the
matching template mirror edits under `internal/template/templates/.claude/rules/moai/` per the
per-file mirror semantics (spec.md §B3). 9 de-dup target groups + cross-cutting mirror/neutrality
constraints. Each milestone is one de-dup target group; zone-registry (highest risk, CLI-gated) is
its own milestone with a research gate.

## §B. Known Issues / Risks (carried from research)

| ID | Risk | Mitigation |
|----|------|------------|
| RK-1 | zone-registry `clause:` reduction breaks `moai constitution validate` (B4) | REQ-SSD-009 scoped SHOULD/partial; verbatim-shorter-excerpt only; full reduction → follow-up SPEC blocker note |
| RK-2 | `agent-common-protocol.md` is §25-sanitized — naive `diff` mirror AC fails | ACs use `TestTemplateNoInternalContentLeak` semantics, not byte-diff (B3) |
| RK-3 | `agent-teams-pattern.md` deletion races CATALOG-SCRUB-001 frontmatter edit | §F dependency; orchestrator sequences; CATALOG drops the superseded milestone |
| RK-4 | Deleting `agent-teams-pattern.md` leaves dangling inbound references | AC greps all referrers repo-wide before/after; rewrite to cookbook 6th pattern |
| RK-5 | Pointer-only reduction accidentally drops a locally-unique delta | Each de-dup AC asserts the *delta* survives, not just that the copy is gone |
| RK-6 | `lifecycle-sync-gate.md` has no mirror — accidental mirror-edit attempt fails build | M5 explicitly skips mirror edit (B3) |
| RK-7 | Contradiction reconciliation (timeout, alwaysLoad, max_teammates) changes meaning | design.md §6 fixes the single canonical value per contradiction before edit |

## §C. Pre-flight (run-phase entry checks)

1. `git fetch origin main` + divergence check (multi-session race per agent-common-protocol).
2. Confirm sibling cohort sequencing with orchestrator (§F: SSOT-DEDUP before/independent of CATALOG-SCRUB).
3. Re-confirm B4 (zone-registry CLI dependency) still holds: re-grep `internal/constitution/validator.go`
   for the `strings.Contains(normalizedSource, normalizedClause)` check before touching any `clause:`.
4. Baseline the Go-test trio green: `go test ./internal/template/... -run 'TestRuleTemplateMirrorDrift|TestTemplateNoInternalContentLeak|TestTemplateNeutralityAudit'`.

## §D. Constraints

- Template-First: every mirrored `.claude/rules/**` edit lands its mirror in the SAME commit (REQ-SSD-010).
- No `embedded.go` edit (it does not exist — B1).
- No Go source edit (spec.md §G). The zone-registry full reduction does NOT proceed if it needs `validator.go`.
- Time-estimate-free milestones; priority-ordered.

## §E. Self-Verification Deliverables (plan-phase)

See spec.md §E (E-PLAN-1..6). progress.md carries the §E.1..§E.5 skeleton; this agent populates
§E.1 only at plan-phase.

## §F. Milestones

Milestones are grouped one-per-de-dup-target. **Priority labels** (High/Medium/Low), no time
estimates. Each milestone edits the deployed file AND its mirror (per B3) unless noted.

### M1 — AskUserQuestion 4-way de-dup (REQ-SSD-001) — Priority High
- SSOT: `core/askuser-protocol.md` (unchanged owner).
- Edit `core/agent-common-protocol.md` §User Interaction Boundary → pointer + subagent-prohibition delta only.
- Edit `core/moai-constitution.md` §MoAI Orchestrator → pointer + orchestrator-obligation delta only.
- Mirror: agent-common-protocol = §25-sanitized (leak-test); moai-constitution = byte-parity (verify allowlist membership at run time).
- Forward-item: note CLAUDE.md §8 as out-of-scope 4th copy (do NOT edit).

### M2 — Hooks config/timeout de-dup + contradiction reconcile (REQ-SSD-002) — Priority High
- SSOT: `core/hooks-system.md` (hook JSON config + path-quoting + timeout table).
- Reconcile FIRST (design.md §6): synchronous default 5s; PostToolUse = 10s + `async: true` exception (rationale baked in); single `alwaysLoad` version value.
- [D7] The timeout reconciliation MUST edit **BOTH** surfaces in hooks-system.md: (1) the L323 timeout TABLE (move PostToolUse out of the 5s synchronous group → the 10s+async exception — the table currently contradicts the JSON example), AND (2) confirm the L244 JSON example (already 10s+async). Editing only the JSON example is the D7 defect.
- [D6] alwaysLoad: state the INTRODUCTION version once (canonical earlier `v2.1.119`); label any `v2.1.121` mention as the separate `updatedToolOutput` extension, NOT a 2nd introduction.
- Edit `core/settings-management.md` §Hooks Configuration → pointer + StatusLine-no-env-var delta only; apply the reconciled timeout + alwaysLoad value.
- Mirror: hooks-system = **byte-parity allowlist** (the ONLY in-scope enrolled file — byte-identical edit, `TestRuleTemplateMirrorDrift` genuinely catches drift); settings-management = byte-identical-by-discipline (NOT enrolled — explicit `diff -q`, NOT the vacuous mirror test).

### M3 — Context-window threshold 5-way de-dup (REQ-SSD-003) — Priority High
- SSOT: `workflow/context-window-management.md`.
- Reduce copies to pointers in: `core/settings-management.md`, `core/zone-registry.md` (CONST-V3R5-022 entry), `development/skill-ab-testing.md`.
- Leave `workflow/session-handoff.md` untouched (already defers).
- NOTE: the zone-registry CONST-V3R5-022 edit touches a `clause:`/entry — coordinate with M_Z (do the threshold pointer as a body edit, not a `clause:` blanking, to avoid the B4 validate break).
- Mirror: settings-management + skill-ab-testing = byte-identical-by-discipline (explicit `diff -q`); zone-registry = §25-divergent (leak-test + scoped `diff`, NOT `diff -q` — pre-existing 13-line delta).

### M4 — 7-item verification batch: confirm-defer + sentinel (REQ-SSD-004) — Priority Low (RE-TARGETED per D4)
- [D4] PREMISE CORRECTION: verification-batch-pattern.md has **NO verbatim 7-command block to remove** — it ALREADY correctly defers to agent-common-protocol §Parallel Execution. The only `coverprofile=cover.out` match is the L29 anti-pattern race-note prose (RETAINED). T4 is largely a **verified no-op**, re-targeted from "removal" to "confirm-defer + add re-sync sentinel note."
- SSOT: `core/agent-common-protocol.md` §Parallel Execution (verbatim 7-command block stays, untouched).
- Edit `workflow/verification-batch-pattern.md` → ADD a "re-sync if the 7-item list changes" sentinel note pointing to the SSOT (the only substantive edit). Do NOT attempt to delete a 7-cmd block (there is none); do NOT assert `grep -c coverprofile=cover.out == 0` (false-fails on the retained L29 race-note).
- [D3] The "intact" AC for agent-common-protocol uses a **per-keyword loop** (each of the 7 distinct commands present ≥1), NOT a vacuous `grep -c "<7-alternation>"` line-count (which returns 17, never 7).
- Mirror: both files are §25-sanitized (leak-test semantics + scoped `diff`, B3).

### M5 — Status Transition Ownership Matrix de-dup (REQ-SSD-005) — Priority Medium
- SSOT: `development/spec-frontmatter-schema.md`.
- Edit `workflow/lifecycle-sync-gate.md` → reduce full matrix table to a pointer; keep close-subject-full-ID one-liner.
- Mirror: **NONE for lifecycle-sync-gate.md** (deployed-only — B3/RK-6). spec-frontmatter-schema = byte-identical-by-discipline (NOT enrolled — explicit `diff -q` IF edited); it is the SSOT and is NOT edited unless it lacks the close-subject line.

### M6 — Team files 4 → 2 + 1 pointer consolidation + file DELETION (REQ-SSD-008) — Priority High
- KEEP `workflow/team-protocol.md` + `workflow/team-pattern-cookbook.md`.
- Fold `workflow/agent-teams-pattern.md` 5+1+1 composition into `team-pattern-cookbook.md` as a 6th pattern (deployed + mirror).
- **DELETE `workflow/agent-teams-pattern.md`** in BOTH trees (deployed + `internal/template/templates/...`).
- [D8] Grep repo-wide for inbound referrers to `agent-teams-pattern.md` with `internal/` BROADLY (NOT just `.claude/ internal/template/templates/` — that scope MISSED 3 comment referrers). Known referrers to rewrite/remove: `internal/config/workflow_role_profiles_test.go:21,45`, `internal/config/defaults_test.go:527` (comments → repoint to team-pattern-cookbook.md), `internal/template/rule_template_mirror_test.go:66` (§25-leak-list comment → drop the deleted entry, do NOT change test logic), `internal/template/templates/.moai/config/sections/workflow.yaml:28` (comment → repoint). ALL comment-only (no Go logic depends on the file).
- Reduce `workflow/worktree-integration.md` §Team Protocol (L380) → cross-reference to `team-protocol.md` (distinctive-line delta on `Mailbox`/`Role Matrix`).
- Reconcile `team-protocol.md` `max_teammates default 10` vs 3-5 ceiling (design.md §6.3): retain the L114 reject-ceiling line AND add the 3-5 recommended-starting-size framing so the contradiction is gone.
- §F dependency: surface to orchestrator (CATALOG-SCRUB-001 frontmatter milestone superseded).
- Mirror: team-protocol, team-pattern-cookbook, worktree-integration = byte-identical-by-discipline (NOT enrolled — explicit `diff -q` each); agent-teams-pattern = §25-sanitized + DELETED (AC verifies absence in both trees).

### M7 — Skill 3-file SSOT boundary (REQ-SSD-006) — Priority Medium
- SSOT split: skill-authoring (schema) / skill-writing-craft (prose) / skill-ab-testing (A/B).
- Edit `development/skill-writing-craft.md` → reduce duplicated frontmatter-schema table (L197-244) to a pointer to skill-authoring.md.
- Edit `development/skill-ab-testing.md` → reduce duplicated frontmatter/progressive-disclosure checklist (L160-161) to a pointer.
- Mirror: all three (skill-authoring, skill-writing-craft, skill-ab-testing) = byte-identical-by-discipline (NOT enrolled — explicit `diff -q` each EDITED file).

### M8 — Agent 3-file SSOT boundary (REQ-SSD-007) — Priority Low
- SSOT: agent-patterns (per-domain tool-whitelist table, L225-230) / orchestrator-templates / agent-authoring.
- agent-authoring.md L302 already points to agent-patterns.md — verify and, if any residual per-domain whitelist/matrix duplication remains in agent-authoring or orchestrator-templates, reduce to a pointer.
- Lighter than feared (research showed the boundary is mostly clean). Scope to the actual residual duplication only.
- Mirror: agent-patterns / agent-authoring / orchestrator-templates = byte-identical-by-discipline (NONE in the §25-leak list, NONE allowlist-enrolled — explicit `diff -q` per EDITED file, NOT the vacuous mirror test). If T8 is a verified no-op, no mirror action.

### M_Z — zone-registry bloat reduction (REQ-SSD-009) — Priority Medium, CLI-GATED — own milestone
- **GATE**: re-confirm B4 (validator.go substring check). If confirmed, do NOT blank `clause:` to labels.
- (a) Shorten each `clause:` to a still-verbatim shorter excerpt of the source HARD clause (keeps `validate` green); measure bloat reduction.
- (b) Narrow `paths:` trigger from `.claude/**,.moai/specs/**,.claude/rules/**` to the minimum preserving intended load (design.md §7 proposes the narrowed value; verify no rule that must load on `.moai/specs/**` edits relies on this).
- If (a) cannot meaningfully reduce bloat without label-only entries (which need a `validator.go` change), record a blocker note recommending follow-up SPEC `SPEC-V3R6-CONST-VALIDATE-LABEL-001` (make validator.go label-aware) and ship only the `paths:` narrowing + safe partial.
- Mirror: zone-registry.md = §25-divergent (NOT byte-parity — pre-existing 13-line CONST-V3R6-001 delta): leak-test PASS + scoped `diff` confirming only the de-dup edit changed; run `moai constitution validate` after the edit as run-phase evidence that no `SentinelDrift` was introduced.

## §G. Anti-Patterns (do NOT)

- AP-1: Editing `embedded.go` (does not exist — B1). Edit `internal/template/templates/` directly.
- AP-2: Using `diff -q` (byte-identity) to assert mirror parity on §25-sanitized files (agent-common-protocol, verification-batch-pattern) OR zone-registry (§25-divergent) — these are NOT byte-identical; use leak-test + scoped `diff` semantics. Conversely, using `TestRuleTemplateMirrorDrift` to assert parity on a NON-enrolled byte-identical-by-discipline file (D1/D2 defect — vacuously green; only hooks-system.md is enrolled) — use explicit `diff -q` for those.
- AP-3: Blanking zone-registry `clause:` to a non-verbatim 3-5 word label (breaks `moai constitution validate` — B4).
- AP-4: Deleting `agent-teams-pattern.md` without first grepping `internal/` BROADLY + rewriting every inbound referrer (D8 — the narrow `.claude/ internal/template/templates/` scope misses 3 comment referrers in `internal/config/`).
- AP-5: Reducing a copy to a pointer while silently dropping the non-owner's locally-unique delta.
- AP-6: Touching CLAUDE.md §8 (out of scope — forward-item only).
- AP-7: Changing the *meaning* of a clause during de-dup (except the two explicit reconciliations, M2/M6).
- AP-8 (D5): Using a pointer-presence grep (`grep -c "<owner>.md" >= 1`) as the "(absent)" proof — pointers are ALREADY present pre-edit, so this proves nothing. Use a distinctive-line before(N)/after(<N) delta.
- AP-9 (D3): Using a `grep -c "<7-alternation>"` line-count to assert the 7 commands are intact — returns 17, never 7. Use a per-keyword loop.
- AP-10 (D4): Asserting `grep -c "coverprofile=cover.out" == 0` on verification-batch-pattern.md — false-fails on the retained L29 race-note prose (there is no 7-cmd block to remove).
- AP-11 (D7): Editing only the hooks-system.md JSON example for the timeout reconciliation, leaving the L323 timeout TABLE contradictory.

## §H. Cross-References

- spec.md §B (verified repo facts), §F (CATALOG-SCRUB dependency).
- research.md (zone-registry CLI consumption analysis).
- design.md §6 (contradiction reconciliations), §7 (zone-registry reduction design), SSOT-owner decision table.
- acceptance.md (per-target grep ACs + deletion 3-clause AC).
- CLAUDE.local.md §2 (Template-First), §25 (template internal-content isolation).
- `internal/template/rule_template_mirror_test.go` (byte-parity allowlist), `internal/template/internal_content_leak_test.go` (§25 leak test).
