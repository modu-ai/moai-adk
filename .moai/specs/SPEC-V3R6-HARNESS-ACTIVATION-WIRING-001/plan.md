# Implementation Plan — SPEC-V3R6-HARNESS-ACTIVATION-WIRING-001

## A. Context

The harness auto-trigger chain is broken at two systemic points (B1 marker absent, B2 main.md absent),
both caused by **orphaned-but-working** Go functions that no flow ever calls. The plan connects existing
mechanisms; it does not invent new ones. All `.claude/**` edits follow the Template-First ordering
(CLAUDE.local.md §2): edit `internal/template/templates/.claude/...` → `make build` → sync.

## B. Tier Classification — Tier M (justified)

| Signal | Assessment |
|--------|------------|
| Components touched | skill body (`moai-meta-harness/SKILL.md`) + workflow body (`project/meta-harness.md` Phase 7) + Go smoke gate (`doctor_harness.go`) + Go tests | 
| Files (edit) | ~6-8 (2 template skill/workflow + 2 mirror sync + `doctor_harness.go` + `doctor_harness_test.go` + possibly a wiring call site) |
| Methodology | TDD per `quality.yaml development_mode` for the Go smoke gate (RED smoke-gate test → GREEN extension) |
| Risk | Medium — touches `doctor_harness.go` (shared diagnosis) + skill bodies (template-first cascade); no algorithm rewrite |
| Cross-cutting | Yes — orchestration skill + Go diagnosis + generated-artifact template |

This is more than a single-doc or single-function change (would be Tier S) but well short of a multi-module
architectural rework (would be Tier L). The dominant work is **wiring existing pieces + a TDD smoke-gate
extension** across the skill/Go boundary, which is the canonical Tier M envelope. **Tier M selected.**

A `design.md` accompanies this plan to record the wiring-mechanism decision (where the `InjectMarker` /
`ScaffoldHarnessDir` call path is anchored: orchestrator Phase 7 skill instruction vs a new CLI surface).

## C. Scope-Item → Milestone Map

The 4 scope items from spec.md §B map onto the milestones below:

| Scope item | Milestone(s) |
|------------|--------------|
| 1. Marker installation ownership (B1, REQ-HAW-001..004) | M2 (decision), M3 (Phase 7 wiring) |
| 2. main.md router generation (B2, REQ-HAW-005..007) | M3 (wiring), M4 (router structure) |
| 3. Generated-artifact self-activation (B4, REQ-HAW-008..009) | M4 |
| 4. Phase-6 post-generation smoke gate (REQ-HAW-010..014) | M5 (TDD) |

## D. Constraints (HARD)

- **D-1 Template-First** (REQ-HAW-015): every `.claude/**` edit originates in
  `internal/template/templates/.claude/...`, then `make build`, then sync to the working `.claude/`. The
  mirror must stay byte-consistent (mirror-drift tests).
- **D-2 Prefix stability** (REQ-HAW-016): generation stays on `my-harness-*`. No `harness-*` rename. The
  protected-prefix Go code (`doctor_harness.go checkLayer1Triggers` matches `my-harness-`) is the
  authority for which prefix is active.
- **D-3 Subagent boundary** (REQ-HAW-003): any Go/CLI surface reached here must not call
  `AskUserQuestion` — the smoke-gate extension is read-only diagnosis (no prompts).
- **D-4 No core-logic rewrite** (EX-6): `InjectMarker` / `ScaffoldHarnessDir` algorithms are wired, not
  rewritten. `main.md` body structure (REQ-HAW-006) MAY be adjusted inside `ScaffoldHarnessDir`'s
  `mainMD()` builder, since that is body content, not the scaffolding algorithm.
- **D-5 No touch** (EX-7, EX-8): `.claude/rules/moai/core/askuser-protocol.md` and any other SPEC's body
  are untouched.

## E. Self-Verification (orchestrator read-only batch on completion)

```bash
# 1. Smoke-gate + doctor harness tests
go test ./internal/cli/ -run TestRunHarnessCheck
go test ./internal/harness/...
# 2. Marker installer caller now exists (was 0 — must be >0 after wiring, OR documented as skill-orchestrated)
grep -rn "InjectMarker\|ScaffoldHarnessDir" --include="*.go" internal/ | grep -v "_test.go"
# 3. Phase 7 body present in workflow (was absent)
grep -n "Phase 7\|5-Layer Activation\|InjectMarker\|main.md" .claude/skills/moai/workflows/project/meta-harness.md
# 4. Template-first mirror consistency
go test ./internal/template/... -run TestRuleTemplateMirror
# 5. Full suite (cascade catch)
go test ./...
# 6. Lint baseline
golangci-lint run --timeout=2m
```

## F. Milestones

### M1 — RED baseline + scope lock (no code change)
- Capture the failing ground truth as assertions: `InjectMarker`/`ScaffoldHarnessDir` have 0 non-test
  callers; template CLAUDE.md has 0 markers; `project/meta-harness.md` has no Phase 7 body.
- Confirm `doctor harness` L3 (marker) + L5 (`main.md`) checks already exist (they do) — the smoke gate
  reuses them.
- Output: `progress.md` §E.2 RED-evidence block.

### M2 — Wiring-mechanism decision (design.md)
- Decide WHERE the post-generation flow calls `InjectMarker` + `ScaffoldHarnessDir`:
  - Option A: orchestrator-driven — `project/meta-harness.md` Phase 7 instructs the orchestrator to invoke
    the marker install + scaffold via a `moai harness install` CLI surface (Go-side call path, testable).
  - Option B: skill-body-only orchestration instruction (no new Go call site; relies on the meta-harness
    skill to drive `InjectMarker` via Bash/CLI).
- Record the chosen mechanism + rationale in `design.md`. The decision determines whether M3 adds a Go
  call site (Option A) or a skill-body Phase 7 (Option B). Default lean: Option A (a thin
  `moai harness install` CLI command makes the wiring testable and removes the dead-code recurrence risk
  the diagnosis flags).

### M3 — Phase 7 wiring (template-first)
- Template-first: in `internal/template/templates/.claude/skills/moai/workflows/project/meta-harness.md`,
  add the **Phase 7 (5-Layer Activation)** body that (per M2 decision) installs the CLAUDE.md marker block
  (`InjectMarker`, REQ-HAW-001..002), ensures `.moai/harness/main.md` exists (`ScaffoldHarnessDir`,
  REQ-HAW-005), and runs the smoke gate (REQ-HAW-010..014).
- If M2 = Option A: add the `moai harness install` CLI command (or extend an existing harness subcommand)
  that calls `InjectMarker` + `ScaffoldHarnessDir`, with `--spec-id`/`--domain` flags, no AskUserQuestion
  (REQ-HAW-003), structured error on CLAUDE.md write failure (REQ-HAW-004).
- `make build`, sync the mirror.

### M4 — main.md router structure + generated-agent self-activation (template-first)
- Adjust `mainMD()` in `internal/harness/layer5.go` so the generated `main.md` is a task-shape → specialist
  ROUTER manifest: domain summary + routing table (task-shape → `.claude/agents/harness/*` specialist) +
  Linked Files section (REQ-HAW-006). Keep the scaffolding algorithm intact (D-4).
- Template-first: update the generated-agent emission contract in
  `internal/template/templates/.claude/skills/moai-meta-harness/SKILL.md` (and the project workflow) so
  generated `.claude/agents/harness/*.md` agents declare `skills:` frontmatter preload of their companion
  `my-harness-*` skill (REQ-HAW-008) and a non-empty trigger-shaped description (REQ-HAW-009).
- Fix the `moai-harness-<domain>-patterns` prefix reference in `project/meta-harness.md` §6.4 Expected
  Outputs (D-2 constraint) — this is a within-scope correction of an emission contract, NOT the EX-1 rename.
  **Disambiguation (EX-1 boundary — the correction target is `my-harness-*`, NOT `harness-*`)**: the §6.4
  reference is corrected to the code-side **`my-harness-*` prefix that the Go code + generator actually use
  today** (the prefix `doctor_harness.go checkLayer1Triggers` matches and that
  `SPEC-V3R6-UPDATE-NAMESPACE-PROTECT-001` protects). It is NOT renamed to the bare `harness-*` prefix;
  `meta-harness SKILL.md:168` documents the doctrine(`harness-*`)-vs-code(`my-harness-*`) drift and warns
  that `harness-*` generation has no protection yet, so this SPEC deliberately stays on the code-side
  `my-harness-*` per EX-1. Advancing to `harness-*` is the future `SPEC-V3R6-HARNESS-NAMESPACE-V2-001`
  (planned), explicitly OUT OF SCOPE here.
- `make build`, sync.

### M5 — Phase-6 smoke gate (TDD, Go)
- RED: add `doctor_harness_test.go` cases asserting the smoke gate fails when (a) `main.md` absent
  (REQ-HAW-010), (b) CLAUDE.md markers absent/unpaired (REQ-HAW-011), (c) a generated agent has empty
  `description` (REQ-HAW-012), (d) a generated agent's `skills:` preload references a non-existent skill
  dir (REQ-HAW-013), (e) a generated agent OMITS the `skills:` key entirely (REQ-HAW-013b / AC-HAW-015 —
  runtime enforcement of REQ-HAW-008's emission contract, distinct from the dangling case (d)).
- GREEN: extend `runHarnessCheck` (or add a `checkLayer6SmokeGate` / agent-frontmatter check) in
  `doctor_harness.go`, preserving L1-L5 semantics (REQ-HAW-014). The L3 (marker) + L5 (`main.md`) checks
  already cover REQ-HAW-010/011 partially; add the agent-description + missing-`skills:` + dangling-skill
  checks.
- REFACTOR: dedupe with the existing layer-check helpers.
- Full `go test ./...` to catch cascades.

### M6 — Retrofit note + verification
- Add a short retrofit note (in the meta-harness skill or `.moai/harness/README.md` scaffold body) telling
  users of existing incomplete harnesses (MINK et al.) to re-run the harness generation flow to install
  markers + `main.md` (EX-3: documentation only, no external-project changes).
- Run the M-E self-verification batch. Confirm `InjectMarker`/`ScaffoldHarnessDir` now have a live call
  path (or a documented orchestrated path), Phase 7 body present, smoke-gate tests green, mirror
  consistent.

## G. Anti-Patterns to Avoid

- **AP-1 — Rewriting the installers**: do NOT reimplement `InjectMarker` / `ScaffoldHarnessDir`. They work;
  they are merely unwired. Wiring is the fix.
- **AP-2 — Prefix drift**: do NOT "fix" the `moai-harness-<domain>-patterns` reference by renaming to
  `harness-*`. The protected prefix is `my-harness-*` (EX-1 / D-2). The §6.4 correction targets
  `my-harness-*`.
- **AP-3 — Dead-code recurrence**: do NOT add the wiring without a test/smoke-gate that would catch its
  removal. The diagnosis's root cause is precisely an installer that was "completed" but never invoked and
  never verified.
- **AP-4 — Skipping template-first**: do NOT edit the working `.claude/` copy directly; originate in
  `internal/template/templates/.claude/...` then `make build`.
- **AP-5 — AskUserQuestion in the smoke gate**: the gate is read-only diagnosis; no prompts (D-3).

## H. Cross-References

- spec.md §C (REQ-HAW-001..016), §E (Exclusions EX-1..EX-8)
- acceptance.md (AC-HAW-001..014)
- design.md (M2 wiring-mechanism decision)
- `.moai/docs/harness-delivery-strategy.md` §6.5 (Option A vs the namespace dependency clarification)
- `internal/cli/CLAUDE.md` (subagent boundary, exit-code discipline, absolute-path rule for the CLI surface)
- `internal/harness/layer3.go`, `internal/harness/layer5.go`, `internal/cli/doctor_harness.go`
