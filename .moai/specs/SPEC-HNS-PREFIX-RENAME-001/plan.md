---
id: SPEC-HNS-PREFIX-RENAME-001
version: "0.1.1"
updated: 2026-07-13
document: plan
---

# Implementation Plan — SPEC-HNS-PREFIX-RENAME-001

## §A Context

Rename the Builder-generated user-owned artifact prefix `harness-<name>` → `hns-<name>` across template contract, Go recognition logic, CI guards, and this repo's local artifacts, with tri-generation backward compatibility (`hns-*` canonical, `harness-*` + `my-harness-*` legacy). Directory names are unchanged. Full requirement set: spec.md §B (REQ-HPR-001..026).

### A.1 Measured scope inventory (plan-time baseline, 2026-07-13)

**Template contract docs carrying the `harness-<name>` placeholder contract (4 pairs, template + live byte-parity mirrors — parity verified for harness-builder.md at plan time):**

| # | Template path | Live mirror |
|---|---|---|
| 1 | `internal/template/templates/.claude/skills/moai/SKILL.md` | `.claude/skills/moai/SKILL.md` |
| 2 | `internal/template/templates/.claude/skills/moai/workflows/harness-builder.md` | `.claude/skills/moai/workflows/harness-builder.md` |
| 3 | `internal/template/templates/.claude/skills/moai/workflows/harness-build-entry.md` | `.claude/skills/moai/workflows/harness-build-entry.md` |
| 4 | `internal/template/templates/.claude/skills/moai-meta-harness/SKILL.md` | `.claude/skills/moai-meta-harness/SKILL.md` |

**Remaining template files matching `harness-` (25 more; classify per REQ-HPR-004):** likely rename-target (artifact-name examples / namespace policy prose): `workflows/harness.md`, `moai-meta-harness/references/{agent-cross-references,seven-phase-workflow}.md`, `workflows/project/meta-harness.md`, `rules/moai/development/skill-authoring.md` (Skills Namespace Policy), `moai-harness-learner/SKILL.md` (policy refs only — skill name itself is non-target), `agents/moai/builder-harness.md`. Likely non-target (hook names, config keys, state paths): `hooks/moai/handle-harness-observe*.sh.tmpl` (4), `settings.json.tmpl`, `.moai/config/sections/{harness.yaml,interview.yaml,system.yaml.tmpl,workflow.yaml}`, `rules/moai/core/{agent-common-protocol,hooks-system}.md`, `rules/moai/development/coding-standards.md`, `rules/moai/workflow/runtime-recovery-doctrine.md`, `workflows/project/{codebase-analysis,doc-generation,mode-detection}.md`, `workflows/project.md`, `workflows/run/context-loading.md`, `skills/moai/SKILL.md` non-contract lines. Final disposition recorded in progress.md §E.2 at run time.

**Go production files (content-token anchors; line numbers drift):**

| File | Anchor | Change |
|---|---|---|
| `internal/cli/update.go` | two classifier blocks carrying `REQ-HNS-001` / `REQ-UNP-001` comment tokens (~L1299–1410) | add `.claude/skills/hns-` + `.claude/workflows/hns-` recognitions |
| `internal/cli/update_preserve_inventory.go` | header inventory comment | add `hns-*` rows to the preserved-inventory doc/logic |
| `internal/cli/doctor_skills.go` | `recognize both harness-*` comment | classify `hns-` prefix as user customization (INFO) |
| `internal/cli/doctor_harness.go` | `checkLayer1Triggers` dir scan + `skills:` ref resolution (`HasPrefix(e.Name(), "harness-")`; NAME-prefix resolution `strings.HasPrefix(ref, "harness-")`, plan-time ~L276) | add `hns-` to both recognition sites |
| `internal/harness/frozen_guard.go` | protected-prefix list (`.claude/skills/harness-`) | add `.claude/skills/hns-` |
| `internal/harness/prefix_conflict.go` | `case strings.HasPrefix(name, "harness-")` + `TrimPrefix` chain | add `hns-` recognition + trim |
| `internal/cli/harness/v4lifecycle.go` | `art := "harness-" + n` artifact-name matcher | dual-pattern: match `hns-<name>*` OR `harness-<name>*` |
| `internal/cli/harness/doctor.go` | `runnerSpecialistRE = regexp.MustCompile(\`harness-[a-z0-9-]+-specialist\`)` (specialist refs); Runner resolved from manifest `runner_workflow` via prefix-agnostic path join | dual-pattern specialist regex only (`(harness|hns)-[a-z0-9-]+-specialist`); Runner resolution needs NO change (REQ-HPR-012) |
| `internal/cli/harness/install.go`, `internal/cli/init.go`, `internal/cli/harness_route.go`, `internal/cli/hook.go` | verify-only | expected no change (CLAUDE.md markers / hook names / route are non-target) — confirm at run time |

**CI guard / test files:** `internal/template/split_namespace_test.go` (sentinel `SPLIT_HARNESS_NAMESPACE_LEAK`; dev-only name set `harness-{release-update,github,release}`), `internal/template/namespace_protection_audit_test.go`, `internal/template/embedded_namespace_test.go` (verify-only), plus CLI test extensions: `update_safety_test.go`, `update_namespace_harness_v2_test.go`, `update_namespace_harness_v4_test.go`, `update_preserve_my_harness_test.go`, `update_preserve_inventory_test.go`, `inventory_test.go`, `update_security_m2_test.go`, `internal/cli/harness/{v4lifecycle_test.go,doctor_test.go}`, `doctor_skills_test.go`, `doctor_harness_test.go`.

**Local artifacts (git mv) + cross-refs:** 3 agents (`.claude/agents/harness/harness-{github,release,release-update}-specialist.md`), 3 skills (`.claude/skills/harness-moaiadk-{best-practices,dev-reference,patterns}/`), 1 Runner (`.claude/workflows/harness-release-update-run.js`); cross-refs in `.claude/commands/harness/{github,release-update,release}.md`, `.claude/commands/harness/release-update/manifest.json` (`agent_file`, `runner_workflow`), Runner JS internal strings, the four unprefixed specialists (`cli-template`, `hook-ci`, `quality`, `workflow` — they matched the cross-ref grep; verify their `skills:` frontmatter), `.claude/rules/moai/development/skill-authoring.md` (live + template mirror).

**Doctrine docs:** `.moai/docs/harness-namespace-doctrine.md`, `.moai/docs/dev-only-commands-isolation.md`.

## §B Known Issues / Open Questions

1. **Prefix case (RESOLVED — final user decision, do not re-open)**: the new artifact prefix is **lowercase `hns-`**. The Claude Code skill/agent naming convention is lowercase-kebab (every existing skill/agent `name:` in this repo is lowercase), so the lowercase form eliminates the uppercase runtime-acceptance risk entirely — NO pre-M3 probe, NO fallback branch, and no mid-run clarification gate exist in this plan. Lowercase `hns-` also satisfies the doctor NAME-prefix resolution without any third pattern: `doctor_harness.go` resolves `skills:` frontmatter references by name prefix (content anchor `strings.HasPrefix(ref, "harness-")`, plan-time ~L276, ground-truth read 2026-07-13), and adding `hns-` there is the same additive lowercase-prefix extension as every other recognition site. The SPEC ID (SPEC-HNS-PREFIX-RENAME-001) and REQ/AC ID tokens keep their uppercase forms — IDs are not artifact names.

2. **Substring-collision analysis (case-sensitivity survey, live-grep grounded 2026-07-13)**:
   - **Uppercase-collision finding (auditor iteration 1, confirmed by survey)**: legacy `REQ-HNS-*` / `AC-HNS-*` comment tokens from SPEC-V3R6-HARNESS-NAMESPACE-V2-001 exist in **5 of the 8** §A.1 production Go files — per-file counts: `update.go` 10, `doctor_harness.go` 2, `prefix_conflict.go` 2, `doctor_skills.go` 1, `frozen_guard.go` 1 (16 occurrences total); `update_preserve_inventory.go` / `v4lifecycle.go` / `doctor.go` 0. The originally-proposed uppercase prefix would have substring-collided with those tokens in every sweep. Under **case-sensitive matching, lowercase `hns-` does not match `REQ-HNS-` / `AC-HNS-`** — the collision disappears entirely.
   - **Case-sensitive sweep discipline [HARD]**: every run-phase grep in this SPEC MUST be case-sensitive (`grep` WITHOUT `-i`) and word/boundary-anchored where needed (e.g. `(^|[^a-z-])hns-`). A case-insensitive grep would resurrect the `REQ-HNS-*`/`AC-HNS-*` token collision.
   - **Live zero-collision baseline**: `grep -rn 'hns-' internal/ .moai/docs/ Makefile .claude/skills/ .claude/agents/ .claude/commands/ .claude/workflows/ .claude/rules/ .claude/hooks/` (excluding this SPEC's own directory) = **0 matches** (executed 2026-07-13). Lowercase `hns-` has zero case-sensitive collisions in the production/artifact tree today.
   - **Legacy-token boundary hazard (unchanged from iteration 1)**: `grep 'harness-'` also matches `my-harness-` and `moai-harness-`. All acceptance greps on legacy tokens use boundary-anchored patterns (e.g. `(^|[^a-z-])harness-`) or explicit exact-token forms, and all Go matching uses `strings.HasPrefix` on normalized path segments (REQ-HPR-008). Note `"my-harness-"` and `"moai-harness-"` do NOT have prefix `"harness-"`, so the Go-side switch ordering is safe by construction.

3. **Stale-anchor risk**: all line numbers cited in §A.1 are plan-time measurements; run-phase must re-verify each content-token anchor before editing (project lesson: re-measure anchors at run time).

## §C Pre-flight Checks (run-phase entry)

1. `git status` clean enough to isolate SPEC commits; SPEC artifacts committed before any parallel work (project lesson: plan-phase artifacts committed immediately, `feat(SPEC-HNS-PREFIX-RENAME-001): ...` subject).
2. Re-run the §A.1 scope greps; diff against this baseline; record drift in progress.md §E.2.
3. Verify byte-parity for all 4 template↔live contract-doc pairs BEFORE editing (plan-time check confirmed parity for harness-builder.md only; verify the other 3).

## §D Constraints

- **M2-before-M3 ordering invariant** [HARD]: Go `hns-` recognition must be merged before any local artifact is renamed (deletion hazard under an intervening `moai update`; see spec.md §C).
- **Additive-only recognition**: no legacy-prefix assertion may be weakened; existing tests keep passing without assertion changes (REQ-HPR-017).
- **Exact `HasPrefix`, never `Contains`; byte-exact case-sensitive** (REQ-HPR-008).
- **Case-sensitive sweeps only** (§B.2): no `grep -i` anywhere in run-phase verification.
- **Template neutrality** (REQ-HPR-024): no SPEC IDs / REQ tokens / internal dates enter `internal/template/templates/**`; run `TestInternalContentLeak` + neutrality tests after M1.
- **Non-target verbatim list** (spec.md §D): the sweep must never touch those tokens.
- **Worktree snapshots + historical SPEC/report artifacts untouched** (REQ-HPR-022).
- **CLAUDE.local.md untouched by agents** (REQ-HPR-025) — flag list delivered instead.
- **Template-First Rule**: every template edit precedes `make build`; live mirrors synchronized in the same milestone.
- No time estimates; priority/order language only.

## §E Self-Verification (run-phase, per milestone)

- E1: per-AC PASS/FAIL matrix vs acceptance.md §D with verbatim command outputs.
- E2: `make build` + `go test ./...` (full suite after every milestone; not only touched packages).
- E3: coverage non-regression for touched packages (`go test -cover ./internal/cli/... ./internal/harness/... ./internal/template/...`).
- E4: subagent-boundary grep (standard batch item) unaffected — no `AskUserQuestion` in hooks/harness code.
- E5: `golangci-lint run` clean on touched files.
- E6: scoped stale-ref grep (AC-HPR-011 formula) after M3.
- E7: `moai harness doctor` + `moai doctor` exit codes after M3.

## §F Milestones

Sequence fixed by user decision. Highest-reversibility decision surfaced first inside each milestone (naming-contract wording before mechanical sweeps).

### M1 — Template Builder contract docs (+ live byte-parity mirrors)

Priority: High. Scope: the 4 contract-doc pairs (§A.1 table) + classification sweep over the remaining 25 template `harness-` files.

1. Edit the GENERATE artifact-naming contract in `harness-builder.md`: all `harness-<name>` placeholder tokens → `hns-<name>` (Runner path, specialist path, skill namespace, verify-skill name, manifest `runner_workflow`, thin-command Runner reference). Non-placeholder tokens (`harness-build-entry.md` filename refs, `harness-spec.yaml`, `.moai/harness/`) stay verbatim.
2. Same pass over `harness-build-entry.md`, `moai/SKILL.md`, `moai-meta-harness/SKILL.md` (emission-contract lines only).
3. Classification sweep (REQ-HPR-004): per-file disposition for the remaining 25 template files; edit rename-targets (`workflows/harness.md`, meta-harness references, `skill-authoring.md` namespace policy → dual-pattern statement, `builder-harness.md` agent, `project/meta-harness.md`); record dispositions.
4. Mirror every template edit to its live counterpart; verify byte-parity per pair (`diff -q`).
5. Neutrality gate: `go test ./internal/template/... -run 'Neutrality|InternalContentLeak'`.

Exit: AC-HPR-001, AC-HPR-002, AC-HPR-012 (template subset), AC-HPR-014 (neutrality subset).

### M2 — Go dual-pattern recognition + tests (MUST precede M3)

Priority: High. TDD: write failing `hns-` recognition tests first.

1. `internal/cli/update.go`: add `.claude/skills/hns-` + `.claude/workflows/hns-` branches to both classifier blocks; update block comments to the tri-generation matrix.
2. `internal/cli/update_preserve_inventory.go`: add `hns-*` inventory rows.
3. `internal/harness/frozen_guard.go` + `prefix_conflict.go`: add `hns-` recognition/trim.
4. `internal/cli/harness/v4lifecycle.go`: dual-pattern artifact matcher (`hns-<name>*` OR `harness-<name>*`); mixed-generation harness resolves as one entry (REQ-HPR-011).
5. `internal/cli/harness/doctor.go`: `runnerSpecialistRE` → dual-pattern `(harness|hns)-[a-z0-9-]+-specialist`; Runner resolution UNCHANGED — it is manifest `runner_workflow`-driven (prefix-agnostic path join), so no filename lookup change exists (REQ-HPR-012).
6. `internal/cli/doctor_skills.go` + `doctor_harness.go`: add `hns-` to prefix recognition sites (dir scan + `skills:` NAME-prefix reference resolution).
7. CI guards: `split_namespace_test.go` name set gains `hns-{release-update,github,release}`; `namespace_protection_audit_test.go` gains `hns-` leak rejection.
8. Test extensions per §A.1 list; E2E preservation sandbox: plant `hns-*` + `harness-*` + `my-harness-*` artifacts in a `t.TempDir()` project, run the update flow, assert all survive byte-identical (pattern precedent: `update_preserve_my_harness_test.go`).
9. Verify-only files (`install.go`, `init.go`, `harness_route.go`, `hook.go`): confirm no change needed; record in progress.md.

Exit: AC-HPR-003..007, AC-HPR-009, AC-HPR-010; full `go test ./...` green.

### M3 — Local artifact rename + cross-refs (gated on M2 merged)

Priority: Medium.

1. `git mv` 3 agents → `hns-{github,release,release-update}-specialist.md`; update frontmatter `name:`.
2. `git mv` 3 skills dirs → `hns-moaiadk-{best-practices,dev-reference,patterns}`; update `SKILL.md` `name:`.
3. `git mv` Runner → `hns-release-update-run.js`; update internal self-refs (header comment, specialist-name strings).
4. Cross-ref sweep: 3 command bodies, `manifest.json` (`agent_file`, `runner_workflow`), 4 unprefixed specialists' `skills:` refs, `skill-authoring.md` concrete examples (live + template mirror re-parity).
5. Verification: `moai harness doctor` exit 0; scoped stale-ref grep = 0 (AC-HPR-011); `moai doctor` skill layers report `hns-*` as user customization.

Exit: AC-HPR-008, AC-HPR-011, AC-HPR-016 (partial).

### M4 — Doctrine docs + build + full verification

Priority: Medium.

1. `.moai/docs/harness-namespace-doctrine.md`: prefix table → `hns-*` canonical + `harness-*`/`my-harness-*` legacy recognition matrix; update the `moai update` behavior contract rows.
2. `.moai/docs/dev-only-commands-isolation.md`: artifact tables + verification checklists → `hns-` names; note the dual-name CI guard.
3. Compose the CLAUDE.local.md §21/§24 pointer-update flag list for the user (REQ-HPR-025; NOT edited).
4. `make build` (re-embed templates); full verification batch: `go test ./...`, `golangci-lint run`, `moai harness doctor`, neutrality tests, scoped stale-ref grep, non-target baseline-delta greps.

Exit: AC-HPR-013..016; Definition of Done (acceptance.md §F).

## §G Anti-Patterns (prohibited)

- Blind repo-wide `sed`/rename across `harness-` matches — every file gets a classified disposition first (substring collisions: `my-harness-`, `moai-harness-`, `harness-spec.yaml`, hook names).
- Case-insensitive greps (`grep -i`) in any run-phase sweep — they resurrect the `REQ-HNS-*`/`AC-HNS-*` legacy-token collision (§B.2).
- Token-presence-only ACs (`grep -c` on prose) as sole evidence — matching-logic reachability requires go test runs + doctor exit codes (project lesson: AC verifies reachability, not token presence).
- Renaming the `.claude/agents/harness/` / `.claude/commands/harness/` directories or the `/harness:` namespace.
- Editing `.claude/worktrees/**`, historical SPECs/reports, or CLAUDE.local.md.
- Weakening or deleting legacy-prefix test assertions to make `hns-` tests pass.
- M3 before M2 (unprotected-prefix deletion hazard).
- Adding SPEC-ID/REQ tokens to template files (neutrality violation).

## §H Cross-References

- spec.md §B (REQ-HPR-001..026), §D Out of Scope, §E traceability.
- acceptance.md §D AC matrix, §F Definition of Done.
- Precedents: SPEC-V3R6-HARNESS-NAMESPACE-V2-001 (dual-recognition mechanism, REQ-HNS-004/005), SPEC-V3R6-HARNESS-V4-001 (v4 artifact set + lifecycle verbs), SPEC-V3R6-DEV-HARNESS-SPLIT-001 (3 dev-only specialists + split-namespace guard), SPEC-V3R6-HARNESS-RENAME-001 (my-harness → harness generation-1 rename).
- Doctrine: `.moai/docs/harness-namespace-doctrine.md`, `.moai/docs/dev-only-commands-isolation.md`, CLAUDE.local.md §21/§24/§25 (pointers only).
