---
spec_id: SPEC-V3R2-ORC-002
phase: "1B — Implementation Plan"
created_at: 2026-05-10
author: manager-spec
status: audit-ready
base_commit: "ab0fc4dda"
branch: feature/SPEC-V3R2-ORC-002-agent-lint
---

# Plan: Agent Common Protocol CI Lint (`moai agent lint`)

Implementation plan for SPEC-V3R2-ORC-002. Defines milestones M1-M5 with
file:line anchors, mx_plan tags, owner roles, and verification gates. The
implementation methodology is **TDD** per `.moai/config/sections/quality.yaml`
(`development_mode: tdd`) and `.claude/rules/moai/workflow/spec-workflow.md`
(RED → GREEN → REFACTOR). All Go source lives under `internal/cli/` and
tests under `internal/cli/agent_lint_test.go` + `internal/cli/testdata/`.

---

## 1. Goal

Ship a build-time lint command (`moai agent lint`) that scans every
`.claude/agents/moai/*.md` (template tree + local tree) for 8 lint rules
(LR-01 .. LR-08) defined in spec.md §2.1. CI integrates the lint as a
required-status check; non-zero exit blocks PR merge. Concurrently, extract
the Skeptical-Evaluator Mandate block into
`.claude/rules/moai/core/agent-common-protocol.md` §"Skeptical Evaluation
Stance" and remove the duplicates from `manager-quality.md` and
`evaluator-active.md`.

---

## 2. Approach

### 2.1 Methodology — TDD RED → GREEN → REFACTOR

`development_mode: tdd` mandates that every behaviour ships behind a
failing test that turns green when the implementation lands.

- **RED (M3.1)**: author negative-fixture tests + AC-driven assertions in
  `internal/cli/agent_lint_test.go`. Each AC from acceptance.md gets at
  least one test that fails until the relevant code lands.
- **GREEN (M3.2 → M3.5)**: implement the 8 lint rules incrementally; each
  rule's test turns green when its logic is added.
- **REFACTOR (M5)**: collapse duplicated rule-implementation idioms into a
  shared helper; finalise `@MX:ANCHOR`/`@MX:WARN` tags; lint-rule
  registration table becomes data-driven for future extension.

### 2.2 Out of Scope (mirroring spec.md §1.2 + §2.2)

The following items are **explicitly excluded** from this SPEC and belong
to other SPECs:

- Rewriting the 9 LR-01-violating agent bodies — handled by
  SPEC-V3R2-ORC-001 (agents) + SPEC-V3R2-MIG-001 (legacy SPEC bodies).
- Effort-level matrix authoring — SPEC-V3R2-ORC-003.
- `isolation: worktree` mandatory enforcement — SPEC-V3R2-ORC-004.
- Hook handler implementation — SPEC-V3R2-RT-001 / RT-006.
- Modifying the FROZEN section of `agent-common-protocol.md` — only the
  new EVOLVABLE §Skeptical Evaluation Stance subsection is added.
- Adding new subagents or skills — only the `agent` cobra subcommand
  surface is new.
- Runtime check during agent spawn — build-time lint only.
- Lint rules for skill files — SPEC-V3R2-WF-001 territory.
- Lint rules for command files — SPEC-V3R2-WF-002 territory.
- Lint rules for hook wrappers (`.claude/hooks/moai/*.sh`).
- AskUserQuestion occurrences inside fenced code blocks — REQ-015
  exemption.
- Performance optimisation beyond O(N) file pass.

### 2.3 Strategy Overview

The plan is sequenced to minimise rework: M1 captures the live violation
baseline that M3 RED tests assert against; M2 wires the cobra subcommand +
common-protocol amendment so M3 has a registered command to invoke; M3
implements the 8 rules behind RED tests; M4 wires CI; M5 REFACTOR + final
verification gate.

| Milestone | Scope | Owner | Priority | Gate |
|-----------|-------|-------|---------:|------|
| M1 | Baseline capture + delivery contract | manager-spec | P0 | research.md citations re-verified, OQ-1..6 resolved, baseline snapshot file created |
| M2 | Cobra subcommand wiring + agent-common-protocol amendment | expert-backend | P0 | `moai agent lint --help` works (no rules yet); rule file has §Skeptical Evaluation Stance |
| M3 | TDD RED → GREEN: 8 lint rules implemented | expert-backend | P0 | All 14-17 ACs PASS; `go test -run TestAgentLint ./internal/cli/...` zero failures |
| M4 | CI integration + JSON schema freeze | expert-devops | P1 | `.github/workflows/ci.yaml` step active; required-status enforced |
| M5 | REFACTOR + MX tags + completion gate | manager-quality | P0 | `make build` byte-identical, `golangci-lint run` clean, all ACs PASS, PR ready |

---

## 3. Milestone Detail

### 3.1 M1 — Baseline Capture + Delivery Contract

**Owner**: manager-spec (audit), expert-backend (verification)
**Priority**: P0 (blocks everything else)

**Tasks**:

1. M1.1 — Re-verify research.md §1 violation counts:
   - `grep -rn "AskUserQuestion" .claude/agents/moai/` matches the table in
     research.md §1.1.
   - `grep -E "^tools:.*\bAgent\b" .claude/agents/moai/*.md` matches §1.3
     (5 files, including expert-mobile post-R5 finding).
   - `grep -n "Skeptical Evaluation Mandate"
     .claude/agents/moai/*.md` matches §1.4 (2 hits: manager-quality L44,
     evaluator-active L34).
   - Run `wc -l .claude/agents/moai/*.md` to capture file sizes for
     test-fixture sizing.

2. M1.2 — Create baseline snapshot artefact:
   - Save the violation list (from M1.1) to
     `.moai/specs/SPEC-V3R2-ORC-002/baseline-snapshot.txt` as test-fixture
     reference for M3 RED.

3. M1.3 — Resolve open questions OQ-1..6 (research.md §9):
   - OQ-1 (manager-brain carve-out): adopt **Option A** (frontmatter-tools
     self-assertion).
   - OQ-2 (inline-code): NO exemption (strict reading of REQ-015).
   - OQ-3 (LR-08 threshold): warning-only, 50% peer-omission threshold,
     no `--strict` promotion.
   - OQ-4 (LR-04 matcher syntax): simple `\|` split + `LR-04-COMPLEX-MATCHER`
     for regex metacharacters.
   - OQ-5 (LINT_TREE_DRIFT): per-file violation-tuple difference + separate
     `LINT_TREE_FILE_MISMATCH` for presence-difference.
   - OQ-6 (runtime budget): build-time CI metric, no in-binary self-timing.

4. M1.4 — Define delivery contract artefacts:
   - 1 new Go source file: `internal/cli/agent_lint.go` (lint engine).
   - 1 new Go test file: `internal/cli/agent_lint_test.go`.
   - 1 new testdata directory: `internal/cli/testdata/agent_lint/` with 8
     fixture files (per research.md §6).
   - 1 modified rule file: `.claude/rules/moai/core/agent-common-protocol.md`
     (gains §Skeptical Evaluation Stance section).
   - 2 modified agent files: `.claude/agents/moai/manager-quality.md`,
     `.claude/agents/moai/evaluator-active.md` (Skeptical block deleted +
     reference comment added).
   - 1 modified expert-security: drop dead `Agent` from tools list (per
     spec.md §2.1).
   - 1 modified template-tree mirror of each above (Template-First).
   - 1 modified CI workflow: `.github/workflows/ci.yaml` (lint step added).
   - 1 modified `internal/cli/root.go` (subcommand registration).

**Files touched**: NONE for code; produces baseline snapshot artefact only.

**mx_plan tags**: none in M1 (analysis-only milestone).

**Verification gate**:

- All research.md §10 citations re-verified at base `ab0fc4dda`.
- Baseline snapshot saved.
- OQ-1..6 documented as resolved in spec.md HISTORY (added in M5.2 commit).

---

### 3.2 M2 — Cobra Subcommand Wiring + agent-common-protocol Amendment

**Owner**: expert-backend (Go code), manager-spec (rule-file edit)
**Priority**: P0 (blocks M3 — lint engine has nowhere to land without
registration; tests cannot run against an unregistered subcommand)

**Tasks**:

1. M2.1 — Author `internal/cli/agent_lint.go` skeleton:
   - `func newAgentCmd() *cobra.Command` returns `agent` parent + `agent
     lint` child.
   - Stub `RunE` returns `errors.New("not implemented")` so the command
     wires up but M3 RED tests fail until logic lands.
   - Flags: `--path string` (multi-value via repeated flag),
     `--format string` (text|json), `--strict bool`, `--help` (cobra
     default).
   - File:line anchor: new file, ~80 lines initial.

2. M2.2 — Register subcommand in `internal/cli/root.go`:
   - Add `rootCmd.AddCommand(newAgentCmd())` in `init()` after L81 (state
     command).
   - Anchor: `internal/cli/root.go:81-82`.

3. M2.3 — Author `internal/cli/agent_lint_test.go` minimal smoke test:
   - `TestAgentLint_HelpFlag` confirms `--help` returns 0 and prints flag
     names. PASSES at M2.3 even with stub RunE because cobra `--help` is
     short-circuited.
   - `TestAgentLint_StubFails` confirms `moai agent lint` returns the
     `errors.New("not implemented")` (RED — flips to GREEN in M3).

4. M2.4 — Amend `.claude/rules/moai/core/agent-common-protocol.md`:
   - Add new section after `## Output Format` (around L80 in current file):

     ```markdown
     ## Skeptical Evaluation Stance

     The reviewer mode operates as a fresh-judgment auditor:

     - Treat every claim as suspect until evidence is shown
     - Demand reproducible verification, not assertions
     - Consider the null hypothesis: did this change actually fix anything?
     - Score quality as the harmonic mean of dimensions, not the average
     - Reject when must-pass criteria fail, regardless of nice-to-have scores
     - Surface contradictions; never silently override a prior rule
     ```

   - Mark the section as EVOLVABLE in zone-registry.md (extends
     CONST-V3R2-039 family).
   - Anchor: `agent-common-protocol.md:80+` (insertion point).

5. M2.5 — Mirror M2.4 in
   `internal/template/templates/.claude/rules/moai/core/agent-common-protocol.md`
   (Template-First).

6. M2.6 — Remove duplicate Skeptical block from agent files:
   - `.claude/agents/moai/manager-quality.md:44-65` — delete `## Skeptical
     Evaluation Mandate` section + 6 bullets, replace with one-line
     reference: `> See `.claude/rules/moai/core/agent-common-protocol.md`
     §Skeptical Evaluation Stance.`
   - `.claude/agents/moai/evaluator-active.md:34-55` — same surgery.
   - Mirror both in `internal/template/templates/.claude/agents/moai/`.

7. M2.7 — Drop dead `Agent` from `expert-security.md` tools list:
   - Anchor: `.claude/agents/moai/expert-security.md:13` and template
     mirror.
   - This is the only ORC-002-scoped agent body change (per spec.md §2.1
     "Remove the dead `Agent` tool declaration from expert-security…").
   - Note: the other 4 LR-02 violations (3 builders + expert-mobile) stay
     unchanged in this SPEC — ORC-001 + a follow-up handle them. The lint
     reports them as expected.

8. M2.8 — Run `make build` to regenerate `internal/template/embedded.go`.

9. M2.9 — Run `go test ./internal/cli/... -run TestAgentLint` — confirm
   smoke tests in M2.3 pass; full lint tests still RED (expected).

**Files touched** (Template-First; local mirrors via M2.8 build):

- `internal/cli/agent_lint.go` (NEW, skeleton)
- `internal/cli/agent_lint_test.go` (NEW, smoke tests)
- `internal/cli/root.go` (MODIFIED: 1 line addition)
- `.claude/rules/moai/core/agent-common-protocol.md` (MODIFIED: +section)
- `internal/template/templates/.claude/rules/moai/core/agent-common-protocol.md` (MIRROR)
- `.claude/agents/moai/manager-quality.md` (MODIFIED: -section)
- `.claude/agents/moai/evaluator-active.md` (MODIFIED: -section)
- `.claude/agents/moai/expert-security.md` (MODIFIED: -1 token in tools)
- `internal/template/templates/.claude/agents/moai/manager-quality.md` (MIRROR)
- `internal/template/templates/.claude/agents/moai/evaluator-active.md` (MIRROR)
- `internal/template/templates/.claude/agents/moai/expert-security.md` (MIRROR)
- `internal/template/embedded.go` (REGENERATED)
- `.claude/rules/moai/core/zone-registry.md` (MODIFIED: add CONST-V3R2-049
  for §Skeptical Evaluation Stance, EVOLVABLE)

**mx_plan tags**:

- `@MX:ANCHOR cmd-registration` on `internal/cli/root.go` new line (high
  fan_in: every binary invocation routes through here).
- `@MX:NOTE skeptical-extracted` on `agent-common-protocol.md` new section
  header (intent: documents the SPEC-V3R2-ORC-002 extraction).
- `@MX:WARN skeptical-no-duplicates` on the new rule-file section block
  (REASON: REQ-005 + REQ-009 — LR-07 will fire if a copy reappears).

**Verification gate**:

- `moai agent lint --help` exits 0 and prints all 4 flags.
- `grep -c "Skeptical Evaluation Mandate" .claude/agents/moai/*.md` = 0.
- `grep -c "Skeptical Evaluation Stance"
  .claude/rules/moai/core/agent-common-protocol.md` ≥ 1.
- `grep -E "^tools:.*\bAgent\b" .claude/agents/moai/expert-security.md`
  returns 0 lines.
- `make build` exit 0; `diff -r` template↔local clean.
- M2.9 smoke tests pass; full lint tests still RED.

---

### 3.3 M3 — TDD RED → GREEN: 8 Lint Rules

**Owner**: expert-backend
**Priority**: P0 (largest implementation block)

**Tasks**:

1. M3.1 — RED: author all AC-driven test fixtures and expectations:
   - 8 testdata fixture files in `internal/cli/testdata/agent_lint/` per
     research.md §6.
   - Test functions (one per AC, 14-17 total):
     `TestAgentLint_AC01_HelpFlag`,
     `TestAgentLint_AC02_BaselineV2_13_2`,
     `TestAgentLint_AC03_CleanRoster`,
     `TestAgentLint_AC04_JSONOutput`,
     `TestAgentLint_AC05_NewViolationFailsCI`,
     `TestAgentLint_AC06_DeadHook`,
     `TestAgentLint_AC07_DuplicateSkeptical`,
     `TestAgentLint_AC08_MissingEffort_NonStrict`,
     `TestAgentLint_AC09_MissingEffort_Strict`,
     `TestAgentLint_AC10_FencedCodeExempt`,
     `TestAgentLint_AC11_MalformedYAML`,
     `TestAgentLint_AC12_CanonicalSkeptical`,
     `TestAgentLint_OrchestratorAllow`,
     `TestAgentLint_TreeDrift`,
     `TestAgentLint_ComplexMatcher`,
     `TestAgentLint_InlineCodeFlagged`.
   - All tests fail at this point (only RunE stub exists).

2. M3.2 — GREEN part 1: frontmatter parsing + body scanner skeleton:
   - Implement `parseFrontmatter([]byte) (AgentFrontmatter, body string,
     err error)` using `gopkg.in/yaml.v3`.
   - Implement `scanBody(body string)` with fence-state tracking per
     research.md §5.1.
   - Wire `RunE` to walk default paths or `--path` flag; produce empty
     violation list (no rules implemented yet).
   - Tests for AC-01 (`--help`) and AC-11 (malformed YAML → exit 2) turn
     GREEN.

3. M3.3 — GREEN part 2: implement LR-01, LR-02, LR-03 (the simple rules):
   - **LR-01**: body-scan for `AskUserQuestion`, skip fences, exempt
     agents whose frontmatter `tools:` field contains `AskUserQuestion`
     (manager-brain carve-out, OQ-1).
   - **LR-02**: split frontmatter `tools:` CSV by comma+optional whitespace;
     reject any token equal to `Agent`.
   - **LR-03**: check `frontmatter.Effort != ""`; warn if absent, error
     under `--strict`.
   - Tests for AC-02, AC-03, AC-08, AC-09, AC-10 (fenced exempt) turn
     GREEN. AC-05 (new violation fails CI) verifies via fixture.

4. M3.4 — GREEN part 3: implement LR-04, LR-05, LR-06:
   - **LR-04**: parse `hooks:` array from frontmatter; for each hook entry,
     split `matcher:` field by `\|`; if any token has regex
     metacharacters beyond `|`, emit `LR-04-COMPLEX-MATCHER` warning;
     otherwise compare each token to tools-list; mismatched token = LR-04
     error.
   - **LR-05**: detect role profile from frontmatter `name` (or `role:`
     field if present); for write-heavy roles {implementer, tester,
     designer}, warn if `isolation: worktree` missing; error under
     `--strict`.
   - **LR-06**: body-line scan for `--deepthink` substring in description
     field or first 30 lines; warn if found (boilerplate), error under
     `--strict`.
   - Tests for AC-06 (dead hook) turn GREEN.

5. M3.5 — GREEN part 4: implement LR-07, LR-08, JSON output:
   - **LR-07**: scan all rule files first (via known glob
     `.claude/rules/**/*.md`), then agent files; compute Skeptical-block
     fingerprint per research.md §8 (header anchor + 6 bullets +
     SHA-256(sorted lowercased bullets)); first occurrence allowed,
     subsequent = LR-07 error.
   - **LR-08**: group agents by `name:`-prefix family; for each agent,
     compute the union of skill-preloads across the family; if the agent
     omits a skill that ≥50% of its family preloads, emit LR-08 warning.
   - **JSON output (REQ-010)**: emit
     `{version: "1.0", summary: {total, errors, warnings}, violations:
     [{rule, severity, file, line, message}]}` when `--format=json`.
   - **Two-tree drift (REQ-017)**: if `--path` not specified, scan both
     trees; compare per-file violation tuples; emit `LINT_TREE_DRIFT` or
     `LINT_TREE_FILE_MISMATCH` warnings as designed.
   - Tests for AC-04 (JSON), AC-07 (duplicate), AC-12 (canonical), tree
     drift, complex matcher, inline-code-flagged turn GREEN.

6. M3.6 — Run `go test ./internal/cli/... -run TestAgentLint -v` — all
   14-17 tests GREEN.

7. M3.7 — Run `go vet ./internal/cli/... && golangci-lint run
   ./internal/cli/...` — zero warnings.

**Files touched**:

- `internal/cli/agent_lint.go` (grew from skeleton to ~500-700 lines)
- `internal/cli/agent_lint_test.go` (~400-600 lines)
- `internal/cli/testdata/agent_lint/*.md` (8 fixtures, ~30-80 lines each)

**mx_plan tags**:

- `@MX:ANCHOR lint-rule-table` on the 8-rule registration table inside
  `agent_lint.go` (high fan_in: every rule routes through dispatch).
- `@MX:WARN fenced-code-state` on the body scanner state-machine block
  (REASON: REQ-015 false-positive risk; regression tests guard).
- `@MX:NOTE orchestrator-exempt` on the manager-brain carve-out logic
  (intent: OQ-1 resolution).
- `@MX:TODO lr08-threshold-tune` on the 50%-threshold constant (resolved
  in M5 after observation).

**Verification gate**:

- All 14-17 lint tests GREEN.
- `go test -race ./internal/cli/...` zero races.
- `go vet`, `golangci-lint run` clean.
- Coverage on `internal/cli/agent_lint*.go` ≥ 85%.
- Hand-run `moai agent lint` on the live tree produces the violation list
  matching research.md §11 (until ORC-001 lands).

---

### 3.4 M4 — CI Integration + JSON Schema Freeze

**Owner**: expert-devops (CI), manager-spec (review)
**Priority**: P1 (depends on M3 — needs working binary to call)

**Tasks**:

1. M4.1 — Edit `.github/workflows/ci.yaml`:
   - Inside the existing **Lint** job, add a step after the existing Go
     lint step:

     ```yaml
     - name: Run moai agent lint
       run: ./bin/moai agent lint
     ```

   - The `Build` job already emits `./bin/moai`; the Lint job runs
     `Build` as a dependency or rebuilds.
   - Ensure step is required for branch-protection (CLAUDE.local.md §18.7
     `required_status_checks.contexts` already includes `Lint`).

2. M4.2 — Document JSON schema in
   `internal/cli/agent_lint.go` package-level doc comment with `version:
   "1.0"` declaration; lock the field set per REQ-014 (stable through
   v3.0.0 minor versions).

3. M4.3 — Add `pre-commit` hook example to README or
   `.moai/docs/agent-lint.md` (REQ-013 Optional):

   ```yaml
   # .pre-commit-config.yaml
   - id: moai-agent-lint
     name: moai agent lint
     entry: moai agent lint --path .claude/agents/moai/
     language: system
     types: [markdown]
   ```

   (NB: this is **documentation only**; the pre-commit hook itself is
   user-installed.)

4. M4.4 — Manual smoke test:
   - Push a synthetic violation to a feature branch (add
     `AskUserQuestion` to a comment in a non-orchestrator agent), open a
     PR, observe CI failure on Lint job, then revert.

**Files touched**:

- `.github/workflows/ci.yaml` (1 step added)
- `.moai/docs/agent-lint.md` (NEW, optional documentation file)
- README.md or `internal/template/templates/README.md` (optional pre-commit
  reference)

**mx_plan tags**:

- `@MX:NOTE ci-integration` on the new CI step (intent: REQ-011 enforcement).

**Verification gate**:

- CI step runs on every PR + push.
- Synthetic-violation smoke test produces red CI status.
- No false positive on the post-M2 tree.

---

### 3.5 M5 — REFACTOR + MX Tags + Completion Gate

**Owner**: manager-quality (verification), manager-spec (sign-off)
**Priority**: P0 (final gate before PR)

**Tasks**:

1. M5.1 — Apply `@MX` tags planned in M2-M4 to actual code lines:
   - Resolve `@MX:TODO lr08-threshold-tune` (decision: keep 50% — tested
     OK in M3.6).
   - Verify `@MX:ANCHOR`, `@MX:WARN`, `@MX:NOTE` placements.

2. M5.2 — Update `spec.md` HISTORY with v0.1.1 entry:
   - Note OQ-1..6 resolutions (M1.3).
   - No spec-content changes — only HISTORY append.

3. M5.3 — REFACTOR: collapse rule dispatch into table-driven shape:
   - Introduce `lintRuleSpec` struct
     `{ID string, Severity Severity, Run func(...) []Violation}`.
   - Reduce per-rule boilerplate; preserve test coverage.
   - Extract fence-state machine into `bodyScanner` helper struct.

4. M5.4 — Run all AC verifications from `acceptance.md`:
   - AC-ORC-002-01 through AC-ORC-002-12 (manual + scripted).

5. M5.5 — Run `make build && make test` (Go suite must remain green; this
   SPEC adds tests but does not modify other packages).

6. M5.6 — Run `golangci-lint run` (must remain clean across project).

7. M5.7 — Run `go test -race ./...` (no new races introduced).

8. M5.8 — Run `diff -r internal/template/templates/.claude/agents/moai/
   .claude/agents/moai/` and same for `.claude/rules/moai/core/` (must be
   empty).

9. M5.9 — Run `moai agent lint` on the live post-M2 tree — expect remaining
   violations matching the residual list in research.md §11
   (post-M2 state). Document in PR description.

10. M5.10 — Stage commit and push to
    `feature/SPEC-V3R2-ORC-002-agent-lint` branch; open PR with
    plan-auditor request for review.

**Files touched**:

- `internal/cli/agent_lint.go` (REFACTOR pass)
- `internal/cli/agent_lint_test.go` (preserved tests, possibly minor
  signature updates from rule-table refactor)
- `.moai/specs/SPEC-V3R2-ORC-002/spec.md` (HISTORY only)
- `progress.md` (final timestamp + audit-ready marker)

**mx_plan tags** (resolved):

- All M2-M4 `@MX:TODO` tags removed.
- Final `@MX:ANCHOR`, `@MX:WARN`, `@MX:NOTE` set verified by `moai mx`
  scan.

**Verification gate**:

- All 14-17 ACs PASS per acceptance.md.
- `make build` exit 0; `make test` exit 0; `golangci-lint run` exit 0.
- Coverage ≥ 85% on `internal/cli/agent_lint*.go`.
- `diff -r` clean.
- No `@MX:TODO` remains anywhere in the modified file set.

---

## 4. Technical Approach Notes

### 4.1 Template-First Discipline

[HARD] Every file change starts in `internal/template/templates/`. After the
template-tree edit, `make build` regenerates
`internal/template/embedded.go`. CLAUDE.local.md §2 strictly enforced.

### 4.2 Subcommand Tree Shape

```
moai
├── (existing subcommands: cc, cg, glm, init, update, doctor, ...)
├── agent              ← NEW parent (M2.1)
│   └── lint           ← NEW child (M2.1)
└── (other existing children)
```

The `agent` parent is added with `Use: "agent"` and a Short description.
Future SPECs may add `moai agent list`, `moai agent fmt` under the same
parent.

### 4.3 Rule Severity Ladder

- **Always-error** (exit 1 on hit, no `--strict` needed): LR-01, LR-02,
  LR-04, LR-07.
- **Warning-by-default, error under `--strict`**: LR-03, LR-05, LR-06,
  LR-08.
- **Parse error** (exit 2): malformed frontmatter; per-file isolated.
- **IO error** (exit 3): unreadable file; per-file isolated, doesn't stop
  scan.
- **Drift warnings** (always-warning): LR-04-COMPLEX-MATCHER,
  LINT_TREE_DRIFT, LINT_TREE_FILE_MISMATCH.

### 4.4 JSON Schema (REQ-010, REQ-014)

```json
{
  "version": "1.0",
  "scanned_paths": [".claude/agents/moai/", "internal/template/.../"],
  "summary": {
    "total": 18,
    "errors": 14,
    "warnings": 4,
    "files_scanned": 52
  },
  "violations": [
    {
      "rule": "LR-01",
      "severity": "error",
      "file": ".claude/agents/moai/manager-strategy.md",
      "line": 59,
      "message": "literal AskUserQuestion in body",
      "context": ["L57: ...", "L58: ...", "L59: violation line", "..."]
    }
  ]
}
```

`version: "1.0"` is locked through v3.0.0 minor versions. Breaking schema
changes (field rename, type change) bump to `"2.0"`. New optional fields
do not bump.

### 4.5 Manager-Brain Orchestrator Carve-out (OQ-1)

Implementation in `agent_lint.go`:

```go
func (l *linter) isOrchestrator(fm *AgentFrontmatter) bool {
    tools := strings.Split(fm.Tools, ",")
    for _, t := range tools {
        if strings.TrimSpace(t) == "AskUserQuestion" {
            return true
        }
    }
    return false
}

// In LR-01 emit logic:
if l.isOrchestrator(fm) {
    l.recordExemption(file, "tools-asserts-aue")
    return // no LR-01 violations for orchestrator-class agents
}
```

JSON output includes the exemption record so audit traceability is
preserved.

---

## 5. Risks Reflected from spec.md §8

| Risk (from spec.md §8) | M-mapping | Mitigation in plan |
|------------------------|-----------|---------------------|
| LR-01 false positive inside fenced code | M3.5 fence-state machine + AC-10 fixture | REQ-015 explicit; regression test fixture; OQ-2 resolution |
| LR-07 fingerprint false-positive on paraphrases | M3.5 SHA-256 sorted-bullet hash | research.md §8 algorithm; testdata duplicate fixture |
| Contributors bypass lint with nolint markers | (no escape hatch) | Policy: no nolint for LR-01/04/07; warnings have no escape either |
| Template vs local tree drift | M3.5 LINT_TREE_DRIFT detection | REQ-017; per-file tuple comparison + smoke test |
| JSON schema breakage breaks IDE integrations | M4.2 version field freeze | REQ-014; semver for schema; locked at v1.0 through v3.0.0 minor |
| Pre-commit too slow for tight loops | M3.5 single O(N) pass + `--path` scope | <500ms target; `--path` reduces scope |
| agent-common-protocol amendment triggers CON-002 gate | M2.4 EVOLVABLE classification | §Skeptical Evaluation Stance is EVOLVABLE; FROZEN sections untouched |
| Lint reports noisy diffs on v2 → v3 transition | (acknowledged) | LINT_TREE_DRIFT non-blocking; clears after ORC-001 lands |
| OQ-1 carve-out misses a future orchestrator agent | M3.3 frontmatter-tools self-assertion | Data-driven; new orchestrators declare AskUserQuestion in tools |
| LR-08 50% threshold over-flags | M3.5 + M5.1 calibration | Warning-only; observe in CI; tune in follow-up SPEC if noisy |

---

## 6. Test Strategy

### 6.1 Unit Tests (M3 RED)

`internal/cli/agent_lint_test.go` ≥ 14 test functions covering:

- Each AC (AC-ORC-002-01 .. AC-ORC-002-12).
- Each rule (LR-01 .. LR-08) with at least one negative fixture.
- Edge cases: empty frontmatter, BOM-prefixed file, nested fences,
  inline-code, complex matcher, two-tree drift, manager-brain carve-out.

### 6.2 Coverage Target

- Package-level: ≥ 85% line coverage on `internal/cli/agent_lint*.go`.
- Critical rule logic (LR-01 fence-state, LR-07 fingerprint): 100%.

### 6.3 Race Detection

`go test -race ./internal/cli/...` — must pass. The lint engine reads files
sequentially in v0.1.0 (no goroutines), so no races expected. Test still
runs to catch future regressions.

### 6.4 Manual Smoke (M5.9)

Hand-run `moai agent lint` on the live tree post-M2 amendment. Compare to
research.md §11 expected output. Discrepancies indicate bugs in M3
implementation, not changes in the tree state.

### 6.5 CI Integration Smoke (M4.4)

Push a synthetic-violation branch; observe CI red status. Revert and
confirm green resumes. Validates the required-status enforcement.

---

## 7. Rollout Plan

1. **Pre-merge** (this PR): plan artefacts only. plan-auditor reviews
   research/plan/acceptance/tasks/progress, validates traceability, signs
   off.

2. **Plan PR merge**: ships plan-only changes (no Go code, no rule-file
   amendment). Run-phase artefacts created in a separate run-phase PR.

3. **Run-phase PR** (subsequent SPEC iteration):
   - M1 baseline capture (artefact-only commit).
   - M2 cobra wiring + agent-common-protocol amendment +
     manager-quality/evaluator-active surgery + expert-security cleanup.
   - M3 RED tests + GREEN implementation.
   - M4 CI step.
   - M5 REFACTOR + final verification.

4. **Post-merge**: lint becomes mandatory for every subsequent agent-file
   PR. ORC-003 / ORC-004 follow-on SPECs upgrade LR-03 / LR-05 from
   warning to error.

5. **Adoption metric**: ORC-001 cleanup PRs + this lint together drive the
   live LR-01 violation count from 9 → 0; LR-02 from 5 → 0 (after
   ORC-001 expert-security + future expert-mobile + builder-platform
   landings); LR-07 from 2 → 0 (this SPEC's M2.6).

---

## 8. Plan-Audit-Ready Checklist

All 15 criteria PASS:

- [x] Frontmatter v0.1.0 schema (this file)
- [x] HISTORY entry present in spec.md (v0.1.0 row)
- [x] 17 EARS REQs across 5 categories (Ubiquitous 5, Event-Driven 5,
      State-Driven 2, Optional 2, Unwanted 3)
- [x] 12 ACs all map to REQs in acceptance.md (100% coverage)
- [x] BC scope clarity (`breaking: true`, BC-V3R2-004)
- [x] File:line anchors ≥ 30 (research.md cites 35+ unique anchors)
- [x] Exclusions section present in spec.md §1.2 + §2.2 + §2 of this plan
- [x] TDD methodology declared (this §2.1)
- [x] mx_plan section present per milestone (§3.2-§3.5)
- [x] Risk table with mitigations (spec.md §8 + this §5)
- [x] Worktree mode path discipline (worktree at
      `/Users/goos/.moai/worktrees/moai-adk-go/orc-002-plan`,
      Template-First in §4.1)
- [x] No implementation code in plan documents (only pseudocode)
- [x] Acceptance.md G/W/T format with edge cases (see acceptance.md)
- [x] tasks.md owner roles aligned with TDD methodology (manager-tdd /
      expert-backend per phase)
- [x] Cross-SPEC consistency (CON-001 zone model, ORC-001 carry-over,
      ORC-003/004 forward dependencies declared)

---

## 9. Out-of-Plan Concerns (deferred)

These items are documented but NOT implemented in this SPEC:

- ORC-003 effort field matrix population (LR-03 promotion to error).
- ORC-004 worktree MUST upgrade (LR-05 promotion to error).
- ORC-001 cleanup of remaining LR-01/LR-02 violations (handles 9 LR-01
  + 4 LR-02 across builder-{agent,skill,plugin}, manager-strategy,
  manager-spec, expert-frontend, expert-backend, expert-devops,
  expert-mobile, manager-quality). This SPEC fixes only the
  expert-security LR-02 + the LR-07 duplicates.
- MIG-001 v2 → v3 user-SPEC migrator (rewrites legacy SPEC bodies that
  reference retired agents).
- Lint config file (`.moai-agent-lint.yaml`) for project-level overrides
  — future SPEC.
- Goroutine parallelism in lint engine — only if CI exceeds 1s budget.
- IDE integration plugins (VS Code, Cursor) consuming `--format=json` —
  community contribution after schema freeze.

---

End of plan.
