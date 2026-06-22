# Acceptance Criteria — SPEC-DIVECC-HOOK-FAILURE-MODE-AUDIT-001

> GEARS-format acceptance criteria with Given-When-Then scenarios. Every AC is observable (file existence, grep output, classification presence). Per `verification-claim-integrity.md`, each AC names the evidence that proves it. Companion to `spec.md` (the canonical lint target) + `plan.md`. Traceability to REQ-DIVECC-XXX in each AC heading.

## §A. Definition of Done

The SPEC is done when: the shared-failure-mode catalogue exists with cited evidence and per-mode classification; the `hook-independence.md` rule exists (template-first) with the catalogue + authoring checklist; the positive signal is documented; and `git diff` shows zero hook-script changes.

## §B. AC Matrix

### AC-DIVECC-001 — Every shared mode cites observed evidence (covers REQ-DIVECC-001, REQ-DIVECC-002)

- **Given** the run-phase audit has enumerated the hook-layer shared failure modes,
- **When** a reviewer reads each enumerated mode in the catalogue,
- **Then** each mode **shall** cite the `grep`/`Read` command run and its observed output (no unobserved claim per verification-claim-integrity.md §1.1 surface 3).
- **Evidence**: catalogue section in `hook-independence.md` (and/or the SPEC's audit output) shows, per mode, a command + output block. Reviewer can re-run each cited command and reproduce the output.

### AC-DIVECC-002 — Each shared mode is classified with rationale (covers REQ-DIVECC-003)

- **Given** the catalogue enumerates N shared failure modes,
- **When** a reviewer inspects each mode's entry,
- **Then** each **shall** carry exactly one classification ∈ {`acceptable-by-design`, `genuine-risk`} and a stated rationale.
- **Evidence**: `grep -c 'acceptable-by-design\|genuine-risk'` over the catalogue ≥ N; every mode entry has a rationale line.

### AC-DIVECC-003 — `hook-independence.md` exists with catalogue + checklist (covers REQ-DIVECC-007, REQ-DIVECC-008)

- **Given** run-phase M5 completed,
- **When** a reviewer lists `.claude/rules/moai/development/`,
- **Then** `hook-independence.md` **shall** exist and **shall** contain both (a) the shared-failure-mode catalogue and (b) a "does this new hook introduce an independent failure mode?" authoring checklist.
- **Evidence**: file exists; `grep -i 'independent failure mode' hook-independence.md` matches the checklist heading; catalogue table present.

### AC-DIVECC-004 — SPEC modifies zero hook scripts (covers REQ-DIVECC-012)

- **Given** the SPEC's deliverable is an audit + doctrine,
- **When** the run-phase changes are diffed,
- **Then** `git diff --stat .claude/hooks/moai/` **shall** show zero changed files.
- **Evidence**: `git diff --stat .claude/hooks/moai/` → empty.

### AC-DIVECC-005 — Positive signal documented (covers REQ-DIVECC-006)

- **Given** the governance gates do not share shared-failure-mode A,
- **When** a reviewer reads the catalogue's positive-signal section,
- **Then** the catalogue **shall** state that the 3 governance gates do NOT share the moai-binary resolution chain, establishing that the wrapper layer's strongest shared failure does not propagate to the governance layer.
- **Evidence**: positive-signal subsection present in `hook-independence.md`; governance-gate cross-tab row (a) shows NO/NO/NO.

### AC-DIVECC-006 — Governance-gate cross-tab present (covers REQ-DIVECC-004)

- **Given** REQ-DIVECC-004 requires per-gate dependency facts,
- **When** a reviewer reads the catalogue,
- **Then** a cross-tab **shall** record, for each of the 3 governance gates, whether it shares (a) moai-PATH chain, (b) `--skip-hook` bypass, (c) shared skip-log, (d) `jq` dependency, (e) `${CLAUDE_PROJECT_DIR:-$PWD}` fallback, (f) `set -e`, (g) timeout.
- **Evidence**: cross-tab table with 3 gate columns × 7 dependency rows present.

### AC-DIVECC-007 — Wrapper fallback-branch finding recorded (covers REQ-DIVECC-005)

- **Given** the wrappers carry a 3-tier resolution chain (not a bare `command -v moai`),
- **When** a reviewer reads the wrapper-layer section,
- **Then** the audit **shall** record that the wrappers fall back across `command -v moai` → `$HOME/go/bin/moai` → `$HOME/.local/bin/moai` → silent `exit 0`, and that the residual genuine-risk is the all-three-absent silent-simultaneous-degradation case.
- **Evidence**: wrapper-layer subsection names the 3 tiers; mode-A classification cites the all-three-absent trigger (not "not in PATH").

### AC-DIVECC-008 — `--skip-hook` classified acceptable-by-design with justification (covers REQ-DIVECC-009)

- **Given** the `--skip-hook` bypass is documented and audit-logged,
- **When** a reviewer reads mode B's entry,
- **Then** it **shall** be classified `acceptable-by-design` with a justification referencing the `.moai/logs/hook-skip.log` audit trail (Chesterton's Fence — an intentional shared bypass is not a defect).
- **Evidence**: mode B entry classification = acceptable-by-design; rationale cites the skip-log audit trail.

### AC-DIVECC-009 — Genuine-risk modes carry a mitigation recommendation (covers REQ-DIVECC-011)

- **Given** a mode is classified `genuine-risk`,
- **When** a reviewer reads its entry,
- **Then** it **shall** carry a mitigation *recommendation* (not an implementation) OR an explicit "no mitigation recommended, because …".
- **Evidence**: each genuine-risk entry has a "Recommendation:" line; no entry recommends a hook-script edit as part of this SPEC.

### AC-DIVECC-010 — Template-first + neutrality (run-phase) (covers REQ-DIVECC-013, REQ-DIVECC-014)

- **Given** `hook-independence.md` is template-managed,
- **When** the run-phase creates it,
- **Then** the template source `internal/template/templates/.claude/rules/moai/development/hook-independence.md` **shall** exist, `make build` **shall** have run, and the template copy **shall not** leak internal SPEC IDs / dates / commit SHAs.
- **Evidence**: template source file exists; `make build` output clean; `go test ./internal/template/... -run TestTemplateNeutralityAudit` (or the CI guard) passes against the new template file.

### AC-DIVECC-011 — SSOT cross-reference, not duplication (covers REQ-DIVECC-010)

- **Given** the new `hook-independence.md` rule must cross-reference — rather than copy — the canonical hook surfaces (`hooks-system.md`, `agent-common-protocol.md` § Hook Invocation Surface, `runtime-recovery-doctrine.md`),
- **When** a reviewer inspects the rule's cross-reference section and its body,
- **Then** the rule **shall** (a) contain the required cross-reference links to all three canonical surfaces, AND (b) contain no verbatim block longer than 10 consecutive lines copied from any of those canonical surfaces.
- **Evidence** (mechanical):
  - **(a) links present** — `grep -c 'hooks-system.md' hook-independence.md` ≥ 1 AND `grep -c 'agent-common-protocol.md' hook-independence.md` ≥ 1 AND `grep -c 'runtime-recovery-doctrine.md' hook-independence.md` ≥ 1.
  - **(b) no duplication** — for each canonical surface S ∈ {hooks-system.md, agent-common-protocol.md, runtime-recovery-doctrine.md}, the longest run of consecutive lines that appear verbatim in BOTH `hook-independence.md` and S is ≤ 10 lines. Reproducible check (run per surface, expect empty output = PASS):
    ```bash
    # Flag any ≥11-line verbatim block shared between the new rule and a canonical surface.
    # awk over the diff: a shared run of >10 lines (no '<'/'>' divergence) is a duplication finding.
    diff <(grep -vE '^\s*$' hook-independence.md) \
         <(grep -vE '^\s*$' <SURFACE>) \
      | grep -E '^  ' | awk 'c=c+1{} END{exit 0}'  # manual: inspect any common-line run >10
    ```
    A pragmatic equivalent: no single fenced/quoted excerpt of a canonical surface inside `hook-independence.md` exceeds 10 lines; short illustrative quotes (≤10 lines) with attribution are acceptable, full-section copies are not.

## §C. Edge cases

- **EC-1 — hook directory changed since plan-phase**: M1 re-confirmation must measure the current tree; if the wrapper count differs from 31 or a gate was added/removed, the catalogue records the current state (the AC binds the catalogue to observed-at-run-phase evidence, not to the plan-phase numbers).
- **EC-2 — a new governance gate appears**: if a 4th governance gate exists at run-phase, the cross-tab gains a column and its dependencies are recorded (REQ-DIVECC-004 binds "each governance gate", not "exactly 3").
- **EC-3 — `sync-phase-quality-gate.sh` jq exception**: the cross-tab must NOT homogenize the gates — row (d) shows the per-gate jq fact (sync-gate = NO, other two = YES). A layer-wide "all gates depend on jq" claim FAILS AC-DIVECC-006.

## §D. Quality gate

- No unobserved claim: every catalogue evidence block is reproducible (AC-DIVECC-001).
- Scope discipline: zero hook-script edits (AC-DIVECC-004).
- SSOT: `hook-independence.md` cross-references rather than duplicates `hooks-system.md` / `agent-common-protocol.md` / `runtime-recovery-doctrine.md` — bound mechanically by AC-DIVECC-011 (links present + no ≥11-line verbatim block).
- Template-first: local rule has a template source (AC-DIVECC-010).
