# Plan — SPEC-V3R6-DOCS-CODEMAPS-V3-001

> Implementation plan for the codemaps SSOT generation + docs-truth
> checklist. This is the FIRST SPEC of the 5-SPEC "Sprint 14 Docs-v3"
> cohort (Phase 0 — prerequisite for README / DOCSITE / COVERAGE / i18n
> rewrites).

## §A. Context

moai-adk-go is at v3.0.0-rc2. The docs (README.* + docs-site) describe an
earlier architecture. A 7-agent dynamic-workflow analysis surfaced 34
documentation gaps. The cohort will rewrite the docs across 4 subsequent
SPECs; each of those SPECs MUST rewrite against a shared factual baseline
rather than re-deriving facts independently.

This SPEC produces that baseline:

- `.moai/project/codemaps/overview.md` + `modules.md` (REFRESHED — the
  directory already exists with 5 default skill-emitted files; this SPEC
  regenerates them for v3.0.0-rc2 so they collectively cover the 10
  capability layers)
- `.moai/project/codemaps/docs-truth.md` (NEW — canonical facts checklist
  with ground-truth citations)

`.moai/project/codemaps/` ALREADY EXISTS (confirmed at plan-phase: 5
files present — `overview.md`, `modules.md`, `dependencies.md`,
`entry-points.md`, `data-flow.md`). This SPEC does NOT create a new
directory; it refreshes the existing tree and adds `docs-truth.md`. No
Go source is touched. The skill does NOT emit `architecture.md`;
capability-layer coverage is verified collectively across `overview.md`
+ `modules.md` (REQ-CM-001 re-targeted per §B.1.1 decision (a)).

## §B. Known Issues (pre-flight risks)

| Risk | Mitigation |
|------|------------|
| `.moai/project/codemaps/` content drifts into "Audit N Finding AX" citation style — violating neutrality (REQ-CM-009 / AC-CM-007) | Run a literal-string grep sweep at M3 for `Audit \d+ Finding`, `REQ-[A-Z]+-\d+`, and bare `SPEC-[A-Z0-9-]+-\d{3}` references inside `.moai/project/codemaps/`. Trim any hits before close. |
| `/moai codemaps` Explore fan-out produces insight prose that reads as fact, blurring the deterministic-extraction vs architectural-insight line (C4) | M1 output convention: every insight-level paragraph starts with a leading "Insight:" label OR lives under an "## Architectural Insight" subsection, separated from "## Public Surface" / "## Dependencies" factual sections. |
| `era: V3R6` override forgotten → ClassifyEra H-2 misclassifies the empty progress.md as V3R2-R4 and emits a spurious `EraAutoDetected` INFO finding | `era: V3R6` is already in the spec.md frontmatter. M3 self-verification greps `^era:` to confirm presence before close. |
| docs-truth checklist facts drift from primary sources between M2 authoring and later cohort SPECs | REQ-CM-012 requires every fact to carry a ground-truth citation; later SPECs re-verify against the cited source, not against docs-truth itself. docs-truth is a navigation aid, NOT a new SSOT. |
| `go list -deps -json` output is large and may inflate codemaps file sizes | Per-package maps record direct + reverse dependencies as compact lists, not the raw JSON dump. Raw JSON stays in-memory only. |
| Reviewer assumes the skill emits `architecture.md` (it does NOT) | §B.1.1 decision (a) documents the filename decision; REQ-CM-001 / AC-CM-001 re-target to `overview.md` + `modules.md`. The skill's own Phase 3 output spec (`.claude/skills/moai/workflows/codemaps.md` lines 104-110) is the SSOT for emitted filenames. |
| CLI-surface extraction undercounts because `rootCmd.AddCommand` spans 26 files (not only `root.go`) and regex drops paren'd constructors | AC-CM-005 uses `grep -rnE '\.AddCommand\(' internal/cli/` across the FULL cli tree (regex captures paren'd constructors) AND cross-checks against `moai --help` rendered verb list. docs-truth lists human-facing verb names, not Go identifiers. |
| `/moai` skill-set enumeration omits commands (17 total) | AC-CM-005 uses `ls .claude/commands/moai/*.md \| wc -l` parity (= 17) instead of a hardcoded allowlist. |

## §C. Pre-flight (before M1)

- [ ] Confirm `.moai/project/codemaps/` already exists with the 5 default
      skill-emitted files (`overview.md`, `modules.md`, `dependencies.md`,
      `entry-points.md`, `data-flow.md`). (Already verified at plan-phase:
      PRESENT, 5 files.)
- [ ] Confirm baseline `go build ./...`, `go vet ./...`, `golangci-lint
      run` all pass on the pre-SPEC tree (so the M3 regression check has
      a clean reference baseline).
- [ ] Confirm ground-truth sources are where the brief says they are:
      `internal/spec/status.go`, `internal/spec/lint.go` (FrontmatterSchemaRule
      ~lines 586-602), `internal/config/defaults.go` (DefaultGLM* lines
      40-57), `.claude/agents/moai/` (= 7 files), `.claude/commands/moai/`
      (= 17 files), `internal/cli/*.go` (rootCmd.AddCommand spans 26
      files). (Already verified at plan-phase.)

## §D. Constraints (carried from spec.md §E)

- C1 Tier M — three sequential milestones.
- C2 Lifecycle `spec-anchored`.
- C3 Neutrality — codemaps content MUST be reusable-neutral; enforced via
  AC-CM-007 (the template-neutrality CI guard does NOT cover codemaps/).
- C4 Determinism — deterministic baseline extraction is separated from
  insight augmentation.
- C5 Era override `era: V3R6` to prevent H-2 misclassification.

## §E. Self-Verification (E1..E7 plan-phase deliverable map)

This section maps the plan-phase artifact set to the §E self-verification
slots that manager-develop will populate at run-phase. Plan-phase
populates only §E.1 (the rest are placeholders in `progress.md`).

- **E1 (AC PASS/FAIL matrix)**: will report the 10 AC-CM-* criteria from
  `acceptance.md` (AC-CM-000 docs-truth existence + AC-CM-001..009).
- **E2 (cross-platform build)**: trivial for this SPEC —
  `.moai/project/codemaps/` is Markdown-only, `go build ./...` MUST still
  pass.
- **E3 (coverage)**: N/A — `.moai/project/codemaps/` is documentation
  output, no new Go code paths to cover.
- **E4 (subagent-boundary grep)**: N/A — this SPEC does not touch
  `.claude/agents/` or hook surfaces.
- **E5 (lint)**: `golangci-lint run` baseline preserved (zero NEW
  findings); `moai spec lint` on this SPEC's own spec.md MUST return zero
  findings.
- **E6 (push state)**: single commit on `main` (Hybrid Trunk 1-person OSS
  per `git-local-workflow-doctrine.md`); no PR (Tier M main-직진 per
  §23 Tier-based PR Routing).
- **E7 (neutrality)**: `.moai/project/codemaps/` carries zero internal
  SPEC-ID / REQ / Audit-citation leakage (AC-CM-007).

## §F. Milestones

> Tier M lifecycle. Milestones are priority-ordered; no time estimates
> per `agent-common-protocol.md` § Time Estimation.

### M1 — Refresh codemaps via `/moai codemaps`

**Owner**: manager-develop (run-phase, cycle_type=ddd — the codemaps skill
itself drives the extraction; DDD's PRESERVE posture fits because we
extract existing reality without changing behavior).

**Input**: the v3.0 codebase tree as-is.

**Action**:

1. Invoke `/moai codemaps` (space, NOT `/moai:codemaps`) to scan the
   codebase. Use `--force` so the existing 5 default files are
   regenerated for v3.0.0-rc2 rather than incrementally patched.
2. Confirm `.moai/project/codemaps/overview.md` AND
   `.moai/project/codemaps/modules.md` are refreshed, and that
   COLLECTIVELY across the two files all 10 capability layers (CLI
   surface / Lifecycle / Harness / Hooks / Templates / Quality /
   Multi-LLM / Web / Governance / Foundation) are represented. The skill
   does NOT emit `architecture.md`; judge coverage across its real
   outputs (per §B.1.1 decision (a)).
3. Confirm the named capability-anchor package set (spec, cli, config,
   statusline, hook, template, harness, session, web — per REQ-CM-002
   selection rule) is covered within `modules.md` or per-package files.
   Mark insight-vs-extraction per C4.
4. If the codemaps skill's `codemaps-extract.js` workflow is available,
   invoke it with `effort:low` for the per-package fan-out (read-only
   extraction per the purpose-driven model+effort taxonomy in
   `dynamic-workflows.md`).

**Exit criteria**:

- `.moai/project/codemaps/overview.md` and `modules.md` both refreshed.
- The 10 capability layers are collectively represented across
  `overview.md` + `modules.md`.
- The 9 named capability-anchor packages are covered in `modules.md` (or
  per-package files).
- Every file in `.moai/project/codemaps/` is Markdown (no JSON dumps, no
  binaries).

**AC binding**: AC-CM-001, AC-CM-009.

### M2 — Author `.moai/project/codemaps/docs-truth.md` checklist

**Owner**: manager-develop (run-phase, continuing the M1 context).

**Input**: M1 codemaps output + the ground-truth source files.

**Action**: author `.moai/project/codemaps/docs-truth.md` (NEW file) with
these sections, each populated by reading the cited source and recording
the verbatim fact:

1. **Agent catalog** (8 retained): read `ls .claude/agents/moai/*.md`
   (= 7 MoAI-custom files) and CLAUDE.md §4. Record the 7 MoAI-custom
   files + `Explore` built-in, each with its class label (Manager×4 /
   Evaluator×2 / Builder×1 / Anthropic built-in×1).
2. **Status enum** (8 lowercase values): read
   `internal/spec/status.go` `ValidStatuses`. Record verbatim:
   `draft`, `planned`, `in-progress`, `implemented`, `completed`,
   `superseded`, `archived`, `rejected`.
3. **Frontmatter schema** (12 fields): read
   `internal/spec/lint.go` `FrontmatterSchemaRule` `required` slice
   (~lines 586-602). Record verbatim: `id`, `title`, `version`,
   `status`, `created`, `updated`, `author`, `priority`, `phase`,
   `module`, `lifecycle`, `tags`.
4. **CLI subcommand surface**: read the HUMAN-FACING verb names from
   `moai --help` rendered output, cross-checked against
   `grep -rnE '\.AddCommand\(' internal/cli/` (spans 26 files, regex
   captures paren'd constructors). Enumerate the `/moai` skill set via
   `ls .claude/commands/moai/*.md` (= 17 files). Record BOTH lists with
   the categorization (terminal verbs vs chat skills). List human-facing
   verb names (`init`, `update`, `glm`, `cc`, `cg`, `web`, `session`,
   `spec`, `harness`, `worktree`, …), NOT internal Go identifiers.
5. **GLM→Claude tier mapping**: read `internal/config/defaults.go` lines
   40-57. Record the FULL tier-models table:
   `DefaultGLMBaseURL = "https://api.z.ai/api/anthropic"`,
   `DefaultGLMHigh = "glm-5.2[1m]"`, `DefaultGLMMedium = "glm-4.7"`,
   `DefaultGLMLow = "glm-4.5-air"`, `DefaultGLMSonnet = "glm-4.7"`,
   `DefaultGLMHaiku = "glm-4.5-air"`, `DefaultGLMOpus = "glm-5.2[1m]"`.

Every section MUST carry a `**Source:**` line pointing at the exact file
(and symbol/line where applicable).

**Exit criteria**:

- All 5 sections present and populated.
- Every fact carries a `**Source:**` citation.
- The agent count is exactly 8 (no more, no less).
- The status count is exactly 8.
- The frontmatter count is exactly 12.
- The `/moai` skill-set count is exactly 17 (parity with
  `ls .claude/commands/moai/*.md | wc -l`).

**AC binding**: AC-CM-002, AC-CM-003, AC-CM-004, AC-CM-005, AC-CM-006,
AC-CM-007, AC-CM-010.

### M3 — Verify coverage, neutrality, and regression baseline

**Owner**: manager-develop (run-phase, final milestone).

**Action**:

1. **Coverage check**: confirm `.moai/project/codemaps/overview.md` AND
   `modules.md` COLLECTIVELY represent all 10 capability layers
   (REQ-CM-001). Confirm the 9 named capability-anchor packages are
   covered (AC-CM-009).
2. **Cross-reference check**: for each docs-truth section, open the cited
   source file and re-verify the recorded fact is still accurate (not
   stale by the time M3 runs).
3. **Neutrality check** (AC-CM-007): run
   `grep -rE 'Audit [0-9]+ Finding|REQ-[A-Z]+-[0-9]+|^SPEC-[A-Z0-9-]+-[0-9]{3}|SPEC-[A-Z0-9-]+-[0-9]{3}' .moai/project/codemaps/`
   and confirm the only hits are self-references inside docs-truth.md
   pointing AT this SPEC's own ID (allowed). Trim any other hits.
4. **Regression check** (AC-CM-008): run `go build ./...`,
   `go vet ./...`, `golangci-lint run` and confirm zero NEW failures
   relative to the §C pre-flight baseline. `.moai/project/codemaps/` is
   Markdown-only so the Go toolchain must be unaffected.
5. **Era-override check**: confirm `spec.md` carries `era: V3R6` so
   `moai spec audit` does not emit an `EraAutoDetected` INFO finding on
   this SPEC.
6. **Self-lint**: run `moai spec lint` on this SPEC's spec.md; confirm
   zero findings (FrontmatterInvalid, OwnershipTransitionInvalid, etc.).

**Exit criteria**:

- All 6 sub-checks PASS.
- E1..E7 §E self-verification populated in `progress.md` (run-phase
  evidence — NOT populated at plan-phase).

**AC binding**: AC-CM-001 (re-verified), AC-CM-007, AC-CM-008, AC-CM-010,
and the plan-phase §E.1 audit-ready signal.

### M4 — Run-phase audit-ready signal

**Owner**: manager-develop (run-phase close).

**Action**: emit the §E.1 plan-phase audit-ready signal (this is the only
§E slot populated at plan-phase) plus the §E.2 run-phase evidence start
marker (literal `§E.2` heading) and §E.3..§E.5 placeholder headings in
`progress.md`, so that `ClassifyEra` H-4 path sees the run-evidence
markers it looks for.

**Exit criteria**:

- `progress.md` carries all 5 `§E.1..§E.5` headings (skeleton only at
  plan-phase).
- §E.1 carries the plan-phase audit-ready signal (populated at plan-
  phase by this agent).
- §E.2..§E.5 are placeholder-only headings; content belongs to
  manager-develop (§E.2/§E.3) and manager-docs (§E.4/§E.5) at later
  phases.

**AC binding**: supports era classification (C5).

## §G. Anti-Patterns to Avoid

- **AP-CM-001 — Hand-deriving facts instead of citing sources**: writing
  "the 8 agents are ..." from memory rather than from
  `ls .claude/agents/moai/` + CLAUDE.md §4. Fix: every docs-truth fact
  MUST carry a `**Source:**` citation (REQ-CM-012).
- **AP-CM-002 — Conflating extraction with insight**: presenting an
  architectural judgment ("the harness package is over-coupled") as if
  it were a `go list -deps -json` fact. Fix: C4 labeling convention.
- **AP-CM-003 — Letting codemaps carry SPEC-internal tokens**: writing
  "REQ-CM-004 mandates 8 agents" inside
  `.moai/project/codemaps/overview.md` or `modules.md`.
  Fix: REQ-CM-009 + AC-CM-007 neutrality sweep at M3.
- **AP-CM-004 — Modifying Go source to "fix" docs**: rewriting
  `internal/spec/status.go` because docs-truth records a different
  value. Fix: REQ-CM-010 + AC-CM-008 enforce code-change zero.
- **AP-CM-005 — Forgetting the era override**: omitting `era: V3R6` so
  `moai spec audit` emits an `EraAutoDetected` INFO finding. Fix: C5 +
  M3 sub-step 5.
- **AP-CM-006 — Letting docs-truth become the new SSOT**: later SPECs
  citing docs-truth as ground truth instead of the cited primary source.
  Fix: REQ-CM-012 — docs-truth is a navigation aid, not an SSOT.

## §H. Cross-References

- **spec.md** §D — full REQ-CM-001..REQ-CM-012 enumeration.
- **acceptance.md** §D — AC-CM-001..AC-CM-009 falsifiable criteria.
- **progress.md** — §E.1..§E.5 skeleton (this agent populates §E.1 only
  at plan-phase).
- **`.moai/docs/git-local-workflow-doctrine.md`** §23 — Tier-based PR
  Routing (Tier M = main 직진, no PR required).
- **`.claude/skills/moai-workflow-spec/SKILL.md`** § GEARS Format —
  notation reference for the REQ-CM-* requirements.
