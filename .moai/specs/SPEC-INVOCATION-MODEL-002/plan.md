# SPEC-INVOCATION-MODEL-002 — Implementation Plan

## Tier Classification

**Tier S (Simple).** Edit surface is doc/policy-only across a small file set:

| Run-phase target | Change kind |
|------------------|-------------|
| `.claude/skills/moai/workflows/review.md` (local) | Phase 2 compose note + conditional caveat + fallback (~1 paragraph) |
| `internal/template/templates/.claude/skills/moai/workflows/review.md` (template mirror) | identical neutral edit |
| `.moai/specs/SPEC-INVOCATION-MODEL-001/spec.md` (closed) | 1 errata-pointer line appended |

Estimated < 60 LOC of prose change, 3 files. Well within Tier S (< 300 LOC, < 5 files). No Go code, no test files, no hook/lint. The design reasoning (capability verdicts) is non-trivial, but the resulting edit is minimal — the clean↔/simplify scope-out removes an entire mapping's worth of edits.

> Artifact-set note: Tier S's minimal set is 2 files (spec.md + plan.md, AC inline). This plan-phase produces the full 4-file set (spec.md + plan.md + acceptance.md + progress.md) per the explicit plan-phase request — a superset of the Tier S minimum, which is permitted.

## §A. Context — Per-Mapping Axis-A Capability Evaluation

The Axis-A candidacy of each mapping was NOT assumed valid; each was checked against the actual skill body read during plan-phase.

### A.1 clean ↔ /simplify — VERDICT: scoped OUT (capability mismatch)

Read `.claude/skills/moai/workflows/clean.md`:
- Phase 1 Static Analysis Scan → dead-code detection (`go vet`, `staticcheck`, `deadcode`, `vulture`, `ts-prune`, `cargo clippy`) across imports/variables/functions/types/files/dependencies.
- Phase 2 Usage Graph Analysis → grep all references, classify Confirmed Dead / Test-Only / Likely Dead / False Positive; `@MX` cross-check.
- Phase 4 Safe Removal → DELETE dead code in reverse dependency order; Phase 5 test-verify; rollback false positives.

Native `/simplify` = "Review the changed code for reuse, simplification, efficiency, and altitude cleanups, then apply the fixes. Quality only — it does not hunt for bugs."

**Delta**: `/moai clean` operates whole-tree (or `--file`-targeted) on ALL code and DELETES what a usage graph proves unreferenced; `/simplify` operates on the CHANGED diff and APPLIES quality refactors without any dead-code usage-graph model. Dead-code removal ≠ changed-code quality refactoring. `/moai clean` does not reinvent `/simplify` — they are complementary. Additionally, `/simplify`'s capability (reuse/simplification/efficiency/altitude) maps CLOSER to the existing `/moai review --lean` over-engineering audit (`delete:`/`stdlib:`/`native:`/`yagni:`/`shrink:` tags) than to `/moai clean`. Forcing a swap would degrade safety (replace a dead-code pipeline that has usage-graph + rollback + MX-ANCHOR protection with a quality-refactor pass). Neither a swap NOR a narrower composition holds — injecting `/simplify` into `/moai clean` would blur clean's dead-code-only scope into quality-refactoring, violating scope discipline.

Action: **record the finding in spec.md §E; make NO edit to clean.md.**

### A.2 review ↔ /code-review — VERDICT: COMPOSE (not swap)

Read `.claude/skills/moai/workflows/review.md`. Current Phase 2: `[HARD] Delegate review to the sync-auditor subagent with all perspectives` → Perspective 1 Security (OWASP + dependency-vuln scan + full-git-history secrets scan + data-isolation) / Perspective 2 Performance / Perspective 3 Quality / Perspective 4 UX. Phase 3 = `@MX` tag-compliance. Phase 4 = report consolidation. Phase 4.5 = design review (`--design`/`--critique`). Plus `--lean` / `--team` / `--security`.

Native `/code-review` = "Review the current diff for correctness bugs and reuse/simplification/efficiency cleanups." (PROGRAMMATIC — `[Skill]` marker.)

**Overlap**: `/code-review` covers correctness bugs + reuse/simplification/efficiency, overlapping `/moai review` Perspective 2 (Performance) + Perspective 3 (Quality). It does NOT cover: Perspective 1 Security depth (full-history secrets, dependency-vuln, data-isolation), Phase 3 `@MX` compliance (MoAI-specific), Phase 4.5 design review, or the sync-auditor 4-dimension skeptical synthesis + verdict.

**Compose design (insertion point = Phase 2, minimal edit)**:
- Add a note in Phase 2: **Where** native `/code-review` is auto-invocable (verified per the conditional-PROGRAMMATIC caveat in `native-invocation-model.md`), the orchestrator MAY invoke it via `Skill()` to cover the correctness-bug + reuse/simplification/efficiency component, feeding its findings into the sync-auditor synthesis. The Security (P1), `@MX` (Phase 3), UX (P4), and design (Phase 4.5) composition is PRESERVED — native `/code-review` augments, never replaces.
- Fallback: **Where** native `/code-review` is NOT auto-invocable (a bundled skill with `disable-model-invocation: true`, a session with `disableBundledSkills`, or a denied `Skill` tool), Phase 2 runs entirely via sync-auditor exactly as today. The current path IS the fallback — no new mechanism required.

Action: **edit review.md (local + template mirror); `make build`.** Keep the edit to ~1 paragraph (anti-over-engineering — the current sync-auditor Phase 2 is the fallback, so no scaffolding).

#### Exact compose-note text (run-phase inserts this VERBATIM at the end of Phase 2, immediately before `## Phase 3`)

This canonical block is the contract between run-phase and the discriminating ACs (AC-IM2-006 / AC-IM2-007). Run-phase MUST insert it verbatim (identical in local + template mirror; neutral — no SPEC ID / REQ / date). The distinctive literals it introduces (`### Native /code-review compose (Axis A)`, `Skill("code-review")`, `disable-model-invocation`, `not auto-invocable`) are ALL absent from the baseline (verified: each returns 0 on the unedited tree), so the ACs fail before the edit and pass only after it.

```markdown
### Native /code-review compose (Axis A)

Where native `/code-review` is auto-invocable, the orchestrator MAY invoke it via `Skill("code-review")` as one Phase 2 finding source, covering the correctness-bug + reuse/simplification/efficiency portion; its findings feed the sync-auditor synthesis. The Security review (Perspective 1), `@MX` tag-compliance (Phase 3), UX review (Perspective 4), and design review (Phase 4.5) composition is preserved — native `/code-review` augments, never replaces, the MoAI-specific perspectives.

Conditional-PROGRAMMATIC caveat: before relying on `Skill("code-review")`, verify auto-invocability at runtime — a bundled skill with `disable-model-invocation: true`, a session with `disableBundledSkills`, or a denied `Skill` tool all remove auto-invocability.

Compose fallback: where native `/code-review` is not auto-invocable, Phase 2 runs entirely via the sync-auditor as today. See `native-invocation-model.md` Axis A.
```

Discriminating-literal ↔ AC map (baseline count = 0 for each novel literal, verified against the unedited tree):
- `Skill("code-review")` (0 on baseline) + Phase-2-awk-scoped `code-review` (0 on baseline — whole-file `code-review` matches only frontmatter L13) + heading `### Native /code-review compose (Axis A)` (0 on baseline) → AC-IM2-006.
- `disable-model-invocation` / `disableBundledSkills` (0 on baseline) [caveat DISCRIMINATOR] + `not auto-invocable` (0 on baseline) [fallback DISCRIMINATOR] + `sync-auditor` co-located in the Phase 2 awk region (pre-exists at L64/L68, so it is a co-location confirmation of the fallback agent, NOT the discriminator) → AC-IM2-007.

## §B. Immutability / Errata Handling (Deliverable 1)

The closed `SPEC-INVOCATION-MODEL-001/spec.md` is 142 lines ending at §F Cross-References (no Version-History tail). Run-phase appends exactly one line at the file end:

```
> Errata (SPEC-INVOCATION-MODEL-002): /security-review and /review are PROGRAMMATIC (built-in exposed via the Skill tool); the §A HUMAN-ONLY classification for these two is superseded — see native-invocation-model.md and this follow-up SPEC.
```

Constraints:
- This is the ONLY permitted edit to the closed body (documented exception, NOT a content correction).
- §A body lines ~31-41 stay byte-unchanged.
- The closed SPEC is a local `.moai/specs/` artifact (non-template) — referencing this SPEC's ID in the errata is allowed (neutrality applies only to template-distributed assets).
- The authoritative corrected record lives in THIS SPEC's §A + §C (not in the closed body).

**Actor + ownership (status-transition hook safety):** the errata append is performed **orchestrator-direct** (`Authored-By-Agent: orchestrator-direct`), NOT by manager-develop/manager-docs. It appends prose ONLY and carries **NO `status:` frontmatter change** (the closed SPEC stays `status: completed`). Because no status transition occurs, the PostToolUse `status-transition-ownership.sh` hook has no transition to validate and passes (exit 0). This avoids a spurious `OwnershipTransitionInvalid` finding: an errata append is a doc annotation on a terminal-state SPEC, not a lifecycle transition.

## §C. Template-First Mirror Plan

| Local path | Template mirror path | Mirror required? |
|------------|----------------------|------------------|
| `.claude/skills/moai/workflows/review.md` | `internal/template/templates/.claude/skills/moai/workflows/review.md` | YES — identical neutral edit |
| `.moai/specs/SPEC-INVOCATION-MODEL-001/spec.md` | (none — `.moai/specs/` is not template-distributed) | NO |
| `.claude/skills/moai/workflows/clean.md` | `internal/template/templates/...clean.md` | N/A — no edit (scoped out) |
| `.claude/rules/moai/workflow/native-invocation-model.md` | `internal/template/templates/...native-invocation-model.md` | N/A — no edit (cross-ref scoped out for neutrality/parity) |

Template neutrality for the `review.md` mirror edit (CLAUDE.local.md §15/§25):
- Allowed in the edit: native command name `/code-review` (Claude Code system identifier), rule citation `native-invocation-model.md` (permanent rule), the generic caveat terms `disable-model-invocation` / `disableBundledSkills` (official Claude Code settings).
- Forbidden in the edit: this SPEC's ID, any REQ token, ISO dates, commit SHAs. The template `review.md` MUST NOT contain `SPEC-INVOCATION`.

After the mirror edit: run `make build` to recompile the embedded templates (`//go:embed all:templates` in `internal/template/embed.go`; there is NO generated `embedded.go`).

## §D. Constraints (DO NOT VIOLATE)

- PRESERVE: `clean.md` (local + template) — untouched. `native-invocation-model.md` (local + template) — untouched. All Perspective 1/2/3/4 + Phase 3 MX + Phase 4.5 design headings in `review.md` — preserved.
- Closed `spec.md`: append exactly ONE errata line; no other edit.
- Anti-over-engineering: no hook, no lint rule, no Go runtime, no new review flag. The review compose is a ~1-paragraph Phase 2 note whose fallback is the existing sync-auditor path.
- Template parity: local `review.md` == template `review.md` after the edit.
- Neutrality: no SPEC ID / REQ token / date / SHA in the template `review.md`.
- Conventional Commits; `Authored-By-Agent:` trailer per ownership matrix.

## §E. Milestones

Priority-ordered (no time estimates per agent-common-protocol § Time Estimation).

- **M1 — review↔/code-review compose (Priority High).** Edit local `review.md` Phase 2: add the compose note (native `/code-review` via `Skill()` as one correctness/quality component), the conditional-PROGRAMMATIC caveat, and the sync-auditor fallback. Mirror the identical neutral edit to the template `review.md`. Run `make build`. Verify local == template; verify Perspective 1/2/3/4 + Phase 3 + Phase 4.5 headings preserved; verify no SPEC-ID leak in template.
- **M2 — divergence errata (Priority High).** Append the single errata-pointer line to the bottom of the closed `SPEC-INVOCATION-MODEL-001/spec.md`, **orchestrator-direct** (`Authored-By-Agent: orchestrator-direct`), appending prose ONLY with **NO `status:` change** (closed SPEC stays `status: completed`) so the PostToolUse `status-transition-ownership.sh` hook sees no transition and passes. Verify `git diff` shows exactly one added line at file end and §A body unchanged.

Ordering rationale: M1 and M2 are independent (different files) and could run in either order; M1 first because the compose is the larger deliverable and its verification (`make build`) is the gating check. M2 touches a closed immutable artifact and is a single-line, low-risk append.

## §F. Anti-Patterns to Avoid

- Forcing the clean↔/simplify swap despite the capability mismatch (would degrade dead-code-removal safety). → scoped out, recorded.
- Editing the closed `spec.md` §A prose in place instead of appending an errata pointer. → immutability violation.
- Leaking this SPEC's ID into the template `review.md` (or into the template `native-invocation-model.md`). → neutrality violation (§25).
- Adding a runtime mechanism (hook/lint) to enforce the classification. → the doctrine is codification-only.
- Over-engineering the review compose with new scaffolding when the existing sync-auditor Phase 2 already IS the fallback path.

## §G. Cross-References

- `.claude/rules/moai/workflow/native-invocation-model.md` — doctrine SSOT (Axis A/B, conditional-PROGRAMMATIC caveat).
- `.claude/rules/moai/development/spec-frontmatter-schema.md` — 12-field canonical frontmatter + Status Transition Ownership Matrix.
- `CLAUDE.local.md` §2/§15/§25 — Template-First + neutrality + internal-content isolation.
- `acceptance.md` — measurable AC enumeration bound to REQ-IM2-001..011.
