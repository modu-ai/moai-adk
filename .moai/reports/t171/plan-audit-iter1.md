# SPEC Review Report: SPEC-MCP-WORKTREE-ROOT-001

Iteration: 1/1 (Tier S — `plan_audit_tier_ceilings` S=1; this is the only plan-audit iteration)
Verdict: **FAIL**
Overall Score: 0.81 (aggregate ≥ 0.80, but 2 blocking findings force FAIL per M6 routing)

Auditor: plan-auditor (independent, card t171). Reasoning context (dossier narrative) ignored
as verdict evidence per M1 Context Isolation; every load-bearing code assertion was re-measured
against the worktree source at the pinned commit.

## Fixed-State Confirmation

```
$ git -C .claude/worktrees/t171 log --oneline -3
576f9b893 feat(SPEC-MCP-WORKTREE-ROOT-001): plan-phase artifacts (Tier S, 2 artifacts)
edcbf593c docs(t171): confirm the cause before planning — one premise falsified, one re-attributed
1519f2660 feat(gate): add a typecheck axis so a broken build cannot pass (#1592)
$ git status --porcelain   → (empty)
```

HEAD `576f9b893`, parent `edcbf593c`, tree clean — matches the pinned state. Tier S artifact
set confirmed: the SPEC directory contains exactly `spec.md` + `plan.md`.

## Code Re-Verification (context isolation — all claims re-measured, none inherited)

| SPEC claim | Re-measured at | Result |
|---|---|---|
| `resolveProjectDir()` reads `CLAUDE_PROJECT_DIR` before `os.Getwd()` (spec.md:26-34) | `internal/cli/session.go:242-250` | **VERIFIED** — quoted block matches source verbatim |
| `spec_progress` handler resolves root via the seam (spec.md:78) | `internal/cli/mcp_server.go:526` → `spec.ListDocs(resolveProjectDir())` | **VERIFIED** |
| `spec_audit` handler (spec.md:76, §4 `:583`) | `mcp_server.go:583` → `AuditOptions{BaseDir: resolveProjectDir()}` | **VERIFIED** |
| `spec_drift` handler (spec.md:77, §4 `:597`) | `mcp_server.go:597` → `spec.Audit(...BaseDir: resolveProjectDir())` | **VERIFIED** |
| codex audit `cwd` (spec.md:79, §4 `mcp_codex.go:1170`) | `mcp_codex.go:1170` → `"cwd": resolveProjectDir()` in `handleCodexAudit` | **VERIFIED** |
| `spec.Audit` defaults `baseDir` to `"."`, cwd-based, no project-root resolution (spec.md:55-58, §4 `audit.go:157`) | `internal/spec/audit.go:157-161` | **VERIFIED** — premise-1 refutation is code-true |
| Out-of-scope: `goal_arm`/`goal_status` (`:466`/`:483`) | `mcp_server.go:466` (`goal.LoadGoal`), `:483` (`handleGoalArm` root) | **VERIFIED** |
| Out-of-scope: `verify_snapshot`/`verify_trend` (`:539`/`:569`) | `mcp_server.go:539`, `:569` | **VERIFIED** |
| Out-of-scope: convergence state dir (`mcp_convergence.go:561`) | `mcp_convergence.go:560-566` (`defaultConvergenceStateDir`) | **VERIFIED** |
| Non-MCP: `goal.go:159/:188`, `memory.go:165/:287`, `launcher_blockcap_infinite.go:139` | grep confirms all six sites | **VERIFIED** |
| Deferral rationale consistency (per-caller verdict undetermined; survey not repair) | All four undecided sites are real `resolveProjectDir()` consumers whose correct scope is genuinely undetermined in code; convergence dir plausibly wants a project-shared primary (state is written to `<root>/.moai/state`) | **CONSISTENT** |
| `audit_multi` accepts `project_root` (presupposed by REQ-6, spec.md:178-180) | `mcp_server.go:367-375` — registered inputs are exactly `claude_verdict`, `target`, `focus`, `gates`, `session_id`; fan-out via `backendCall(ctx, name, target, focus)` → codex `cwd` fixed at `resolveProjectDir()` | **REFUTED** → finding D1 |

Cross-backend audit tools were NOT invoked in this audit (inadmissible here by dispatch
instruction — the tools' project-root resolution is the defect under audit).

## Must-Pass Results

| # | Criterion | Result | Evidence |
|---|---|---|---|
| MP-1 | REQ number consistency | **PASS** | spec.md:116-186 — `### REQ-1` … `### REQ-6`, sequential, no gaps, no duplicates |
| MP-2 | EARS/GEARS compliance (requirement layer only) | **PASS** | All six REQs carry SHALL with a subject or fronted condition (REQ-1 ubiquitous, REQ-2/3 condition-fronted state/event form, REQ-4/5 ubiquitous, REQ-6 event-driven). No informal "should"/"may" in normative text. ACs are Given-When-Then in the verification layer (§3, Tier S inline) — correct format, not penalized. Judgment made entirely against the `REQ-XXX` layer in spec.md §3 |
| MP-3 | YAML frontmatter validity | **PASS** | spec.md:1-16 — all 12 canonical fields present with correct types: `id` (string, schema-valid), `title` (quoted), `version: "0.1.0"` (quoted semver), `status: draft` (enum), `created`/`updated: 2026-08-22` (ISO), `author: lane-7`, `priority: P1`, `phase: "v3.1.3"` (release target — not a prohibited stage name), `module: "internal/cli"`, `lifecycle: spec-anchored`, `tags` (comma string). Optional extras `era: V3R6`, `tier: S`. Zero snake_case aliases |
| MP-4 | Language neutrality | **N/A** | Single-language Go project (`module: internal/cli`); no multi-language tooling surface |
| MP-5 | D7 cross-SPEC reconciliation | **PASS** | `grep -Eo 'SPEC-([A-Z][A-Z0-9]+-)+[0-9]+'` over spec.md + plan.md → only the SPEC's own ID. No external SPEC references → no reconciliation obligation, no BLOCKING finding |
| MP-6 | D8 cross-platform discipline | **PASS** | `grep -c syscall` → 0 in both artifacts (auto-pass) |
| MP-7 | Clarification gate | **PASS/N/A** | `grep '\[NEEDS CLARIFICATION' plan.md` → 0 matches; `research.md` does not exist (Tier S 2-artifact set) → N/A for research.md per MP-7's own rule |

All seven must-pass criteria hold. The FAIL verdict comes from blocking (non-must-pass)
findings D1 and D2 below, which affect the SPEC's internal consistency and its own stated
verification criteria — these are fix-before-verdict defects under M6, not score-offsettable.

## Category Scores

| Dimension | Score | Band | Evidence |
|---|---|---|---|
| Clarity | 0.75 | one requirement needs interpretation | REQ-6 (spec.md:178-186) presupposes a `project_root` carrier on `audit_multi` that REQ-1's four-tool enumeration (spec.md:116-120) does not create and the code does not have (mcp_server.go:367-375) |
| Completeness | 0.75 | one non-critical section substantively incomplete | §2.1 (spec.md:91-104): "The four undecided consumers" names five items; the MCP-path consumer enumeration omits `mcp_glm.go:97-98` (`projectDirResolver`) and `mcp_server.go:105` (registration enablement read), and the non-MCP list omits `plan.go:67`, `todo.go:68`, `verify.go:72`, `session.go:224/:357`, `migrate_profiles.go:378` (all confirmed by grep) |
| Testability | 0.75 | ACs measurable, one with interpretation; one direction unasserted | AC-1's positive direction names only `spec_audit` (spec.md:122-127); plan.md M2 (plan.md:98-101) exits the codex milestone on AC-2+AC-3 only — no criterion anywhere asserts the parameter-present path for the codex `cwd`, contradicting plan.md §D "Both directions asserted on every new test" (plan.md:54-55). AC-6's "states which tree each backend read" (spec.md:182-184) presupposes an observable the current ConvergenceResult does not expose |
| Traceability | 1.0 | full bijection | REQ-1→AC-1 … REQ-6→AC-6, each AC inline under its REQ heading; no orphan ACs, no uncovered REQs |

Aggregate: (0.75 + 0.75 + 0.75 + 1.0) / 4 = **0.81** (≥ 0.80 threshold; FAIL is verdict-driving via blocking findings, not score).

## Defects Found

**D1** — spec.md:178-180 (REQ-6) vs spec.md:116-120 (REQ-1) vs `internal/cli/mcp_server.go:367-375` — REQ-6 requires `audit_multi` to be "re-run from a worktree with `project_root` naming that worktree", but `audit_multi` is not among the four tools REQ-1 gives the parameter, and its registered inputs (`claude_verdict`/`target`/`focus`/`gates`/`session_id`) have no `project_root`; its codex fan-out pins `cwd` at `resolveProjectDir()` regardless. As written, AC-6 is not executable without a change REQ-1 does not authorize — internal inconsistency between the scope enumeration and the terminal requirement. — Severity: major — Class: **blocking** — Fix: one sentence, either direction — (a) add `audit_multi` to REQ-1 as a forwarding carrier ("accepts `project_root` and passes it to the codex/glm backends' tree resolution"), or (b) rewrite REQ-6/AC-6 to re-run the `codex_audit` tool directly (which does receive `project_root`) instead of `audit_multi`.

**D2** — plan.md:98-101 (M2) vs plan.md:54-55 (§D) and spec.md:122-127 (AC-1) — No acceptance criterion asserts the parameter-present direction for the codex `cwd` path: AC-1 names only `spec_audit`, plan M1 generalizes to the three SPEC tools, and M2's exit is "AC-2 and AC-3 met there" — the absent/invalid directions only. The codex `cwd` is the surface the card was motivated by (lane-9) and the surface AC-6's post-repair check depends on, yet nothing in the SPEC proves the redirect works there; this contradicts plan §D's own "Both directions asserted on every new test". — Severity: major — Class: **blocking** — Fix: extend AC-1 (or M2's exit) to assert the codex positive direction at the unit level — e.g. "Given a codex audit call carrying `project_root` naming a worktree, the constructed review params carry that path as `cwd`" (assertable on the `params` map at `mcp_codex.go:1167-1171` without a live backend).

**D3** — spec.md:97 — "The four undecided consumers" then names five items (`goal_arm`, `goal_status`, `verify_snapshot`, `verify_trend`, convergence state directory). Arithmetic slip in the scope partition. — Severity: minor — Class: optional — Fix: "The undecided consumers — four MCP tools and the convergence state directory —".

**D4** — spec.md:91-104 — §2/§2.1's consumer enumerations read as exhaustive partitions but are not: the MCP-path set also contains `mcp_glm.go:97-98` (`projectDirResolver`, GLM llm.yaml location) and `mcp_server.go:105` (tool-enablement read), and the non-MCP list omits six further confirmed CLI-side call sites. REQ-4's grep-based "every `resolveProjectDir()` call site" mandate (spec.md:151-156) does cover them all, so no repair escapes this SPEC — but the prose under-states the survey surface. — Severity: minor — Class: optional — Fix: mark both lists as illustrative ("including …; the full set is re-derived by grep at M3") or extend them.

**D5** — plan.md:17-23 (§A surfaces) — The touch surface is ~8-10 paths once template mirrors and the two auditor agent bodies are counted, exceeding the Tier S "< 5 files" guidance axis (LOC and REQ/AC-budget axes fit comfortably; core code is 2 Go files + tests). Tier S remains defensible — the overrun is mechanical mirror bookkeeping, and the tier anti-pattern (under-tiering 1000+ LOC) does not apply — but the lead should confirm the call consciously. — Severity: minor — Class: optional — Fix: no text change required; lead confirms the tier call, or trims M4's doc surfaces if the file axis is judged load-bearing.

## Verified Non-Findings (defect-hunt items checked and clean)

- **REQ-3 rejection observability**: AC-3 (spec.md:144-149) names the observable (tool error naming the offending path, nothing audited) across three invalid inputs (non-existent / file / no `.moai`) — binary-testable, and the error-not-fallback rationale is stated in the requirement. Clean.
- **AC-4 "what to measure" column enforcement**: AC-4's explicit fail sentence (spec.md:158-162 — "A row reading 'unknown' with no statement of what would settle it fails this criterion") does fail a deferred row lacking the settling-evidence statement. Enforceable as written. Clean.
- **AC-6 recurrence semantics**: "A recurrence does not fail this criterion … What fails it is not looking" (spec.md:184-186) — pass/fail binds to the recording act, unambiguous. (The separate observability wrinkle is scored under Testability and resolved by D1's fix.) Clean as worded.
- **Scope-overrun veto marker**: spec.md:81-89 — "**Note on `spec_progress` — added beyond the dispatched scope, flagged for veto.** … drop it if that call is wrong" — marker present, mirrored in plan.md §D2 (plan.md:68-75). The lead can veto item-by-item. Clean, and good discipline.
- **Exclusions vs requirements**: deferring the undecided consumers (repair) vs REQ-4 surveying them (record-only) — no contradiction; the survey/not-repair split is coherent against the code.
- **§5 constraints vs REQs**: "no change to `resolveProjectDir()`'s body" is consistent with REQ-1's handler-level override. Clean.

## Regression Check

Iteration 1 — no prior defects to regress.

## Recommendation

FAIL on two blocking findings; both are one-sentence fixes to spec.md/plan.md, and both
were verified against source rather than inferred. Because Tier S caps plan-audit at one
iteration (`plan_audit_tier_ceilings` S=1), the route is: the author applies D1 (choose
carrier (a) or (b)) and D2 (extend the codex positive-direction assertion), bumps
`version`/`updated`, and the delta is re-checked against this report's defect list — a
fresh from-scratch re-audit is not required to clear two enumerated line-fixes. D3-D5 are
at the lead's discretion and do not block.

The SPEC's diagnostic core is strong and fully code-true: every line citation in §1/§2/§4
re-verified exactly (session.go:242, mcp_server.go:526/583/597, mcp_codex.go:1170,
audit.go:157, and all nine out-of-scope sites), the premise-1 refutation (CLI already
correct via `baseDir = "."`) is confirmed, and the `spec_progress` scope excess is honestly
flagged for veto. The two defects are in the requirement chain's last mile, not in the
cause analysis.
