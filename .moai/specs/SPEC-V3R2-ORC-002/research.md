---
spec_id: SPEC-V3R2-ORC-002
phase: "0.5 — Codebase Research"
created_at: 2026-05-10
author: manager-spec
status: research-complete
base_commit: "ab0fc4dda"
branch: feature/SPEC-V3R2-ORC-002-agent-lint
---

# Research: Agent Common Protocol CI Lint (`moai agent lint`)

Phase 0.5 codebase research for SPEC-V3R2-ORC-002. Audits the live violation
set in the current agent roster (template + local trees), surveys related
lint-tool design (cobra+viper, eslint, golangci-lint, ast-grep), evaluates
Go YAML-frontmatter parsing options, and pins file:line evidence behind every
lint rule (LR-01 .. LR-08) declared in spec.md §2.1.

The research answers seven questions:

1. **Violation baseline**: what concrete LR-01/LR-02/LR-04/LR-07 hits exist
   today? — needed to author negative-fixture tests in M3 RED.
2. **Frontmatter parsing**: which Go YAML library is used elsewhere in
   moai-adk-go, and is it sufficient for partial-frontmatter fields?
3. **Subcommand registration**: where does the existing `moai` CLI register
   subcommands, and what is the canonical pattern? — needed to wire `moai
   agent lint` non-disruptively in M2.
4. **Fenced-code skip semantics**: how do other Markdown linters (ast-grep,
   markdownlint, eslint-plugin-markdown) implement the LR-01 fenced-code
   exemption (REQ-015)? — design cross-check.
5. **Existing test fixtures**: what testdata pattern does `internal/cli/`
   already use, so M3 RED tests fit the same mould?
6. **Skeptical-Evaluator block fingerprint**: how do we semantically identify
   the 6-bullet block (LR-07)? — fingerprint must be robust across whitespace
   normalisation.
7. **CI integration**: what is the existing GitHub Actions workflow shape,
   and where does the lint step plug in?

---

## 1. Violation Baseline (live as of base `ab0fc4dda`)

### 1.1 LR-01 — literal `AskUserQuestion` in agent body

`grep -rn "AskUserQuestion" .claude/agents/moai/` (template tree mirrors)
returns the following rule violations (excluding fenced-code citations,
excluding the rule-file mention in `agent-common-protocol.md`).

Per SPEC-V3R2-ORC-002 §2.1 LR-01 the body scanner SHALL skip fenced-code
regions; the table below classifies each occurrence as either a **violation
candidate** or a **fenced-code exemption**.

| File | Line | Snippet (≤40 chars) | Classification |
|------|-----:|---------------------|----------------|
| `.claude/agents/moai/manager-strategy.md` | 59 | `Use AskUserQuestion to verify critical assumptions` | LR-01 candidate |
| `.claude/agents/moai/manager-strategy.md` | 72 | `Present via AskUserQuestion with clear trade-offs` | LR-01 candidate |
| `.claude/agents/moai/builder-agent.md` | 61 | `[HARD] Use AskUserQuestion to ask for agent name` | LR-01 candidate |
| `.claude/agents/moai/builder-agent.md` | 105 | `Sub-agents cannot use AskUserQuestion — collect…` | meta-rationale (also flagged) |
| `.claude/agents/moai/manager-quality.md` | 102 | `WARNING: Warn user, present options via AskUserQuestion` | LR-01 candidate |
| `.claude/agents/moai/expert-frontend.md` | 66 | `If unclear, use AskUserQuestion: React 19, Vue …` | LR-01 candidate |
| `.claude/agents/moai/expert-frontend.md` | 93 | `Use AskUserQuestion if ambiguous` | LR-01 candidate |
| `.claude/agents/moai/manager-project.md` | 32 | `CANNOT use AskUserQuestion — all user choices …` | meta-rationale (also flagged) |
| `.claude/agents/moai/claude-code-guide.md` | 83 | `[HARD] Do NOT invoke AskUserQuestion (subagent …)` | meta-rationale (also flagged) |
| `.claude/agents/moai/manager-brain.md` | 9 | `tools: Read, …, AskUserQuestion, mcp__context7…` | tools-list — see §1.2 below |
| `.claude/agents/moai/manager-brain.md` | 29-125 | 11 occurrences (orchestrator-mode body language) | LR-01 candidates (manager-brain is exception per §1.4) |
| `.claude/agents/moai/expert-backend.md` | 66 | `If framework is unclear, use AskUserQuestion …` | LR-01 candidate |
| `.claude/agents/moai/expert-backend.md` | 82 | `Use AskUserQuestion when ambiguous` | LR-01 candidate |
| `.claude/agents/moai/builder-skill.md` | 61 | `[HARD] Use AskUserQuestion to ask for skill name` | LR-01 candidate |
| `.claude/agents/moai/manager-spec.md` | 180 | `Use AskUserQuestion for user confirmation …` | LR-01 candidate |
| `.claude/agents/moai/builder-plugin.md` | 81 | `[HARD] Use AskUserQuestion to clarify scope:` | LR-01 candidate |
| `.claude/agents/moai/expert-devops.md` | 63 | `If unclear, use AskUserQuestion: Railway …` | LR-01 candidate |
| `.claude/agents/moai/expert-devops.md` | 78 | `Use AskUserQuestion if ambiguous` | LR-01 candidate |

**R5 audit baseline reconciliation**: R5 §Common Protocol compliance table
counted **9 violations of Rule 1**. The grep above counts more (~20 raw hits
including meta-rationale and orchestrator-only agents). The R5 number tracks
**unique agent files with at least one body-scope violation, excluding
manager-brain orchestrator carve-out and excluding rule meta-mentions**:

```
manager-strategy, builder-agent, manager-quality, expert-frontend,
expert-backend, builder-skill, manager-spec, builder-plugin, expert-devops
= 9 agents
```

This is the **AC-ORC-002-02** target count. The lint is expected to report
exactly 9 LR-01 violations (one per agent file) when run on the v2.13.2
roster.

### 1.2 manager-brain orchestrator carve-out (research finding)

`manager-brain.md` declares `AskUserQuestion` in its `tools:` frontmatter
(L9) and uses the literal string repeatedly in its body (L29-L125). This is
intentional: manager-brain is an **orchestrator-class agent** that runs
Socratic interviews via the brain workflow (per `.moai/brain/` SPECs). It is
the sole legitimate sub-agent caller of AskUserQuestion.

The lint must distinguish manager-brain from regular subagents. Two design
choices:

- **Option A — frontmatter signal**: agents with `AskUserQuestion` in
  `tools:` are auto-exempted from LR-01 (the field declaration self-asserts
  legitimacy).
- **Option B — explicit allowlist**: hardcode `manager-brain` in the lint
  exemption list.

**Recommendation: Option A**. It is data-driven, future-proof (if a new
orchestrator-class agent emerges, frontmatter declaration is sufficient),
and gives the agent-author a single place to assert intent. Option B
hardcodes a name that may change.

OQ-1 below tracks final resolution.

### 1.3 LR-02 — `Agent` token in tools list

`grep -E "^tools:.*\bAgent\b" .claude/agents/moai/*.md` returns:

| File | Line | Tools-list excerpt |
|------|-----:|---------------------|
| `.claude/agents/moai/builder-agent.md` | 13 | `Read, Write, Edit, …, Agent, Skill, mcp__sequential…` |
| `.claude/agents/moai/builder-skill.md` | 13 | `Read, Write, Edit, …, Agent, Skill, mcp__sequential…` |
| `.claude/agents/moai/builder-plugin.md` | 13 | `Read, Write, Edit, …, Agent, Skill, mcp__sequential…` |
| `.claude/agents/moai/expert-security.md` | 13 | `Read, Grep, Glob, …, Agent, mcp__sequential…` |
| `.claude/agents/moai/expert-mobile.md` | 13 | `Read, Write, Edit, …, Agent, Skill, mcp__sequential…` |

R5 §Tool scope audit listed **4 dead `Agent` tool declarations** (builder-3
+ expert-security). The grep finds **5** because expert-mobile (post-R5
addition) also carries the dead tool. Per spec.md §1.2 Non-Goals, agent body
rewriting beyond ORC-001 scope is out of scope; ORC-002 reports the
violations, ORC-001 corrects them in the M3 refactor pass (or this SPEC's
M4 if downstream). expert-mobile is a post-R5 finding to surface.

R5 baseline target = 4; live count = 5 (with expert-mobile). The lint is
expected to report 5 (or 4 after ORC-001's expert-security cleanup lands).
AC-ORC-002 acceptance language uses "≥4" wording to handle both states.

### 1.4 LR-07 — duplicate Skeptical-Evaluator Mandate block

`grep -n "Skeptical Evaluation Mandate" .claude/agents/moai/*.md` returns
exactly 2 matches in the current tree:

| File | Line | Context |
|------|-----:|---------|
| `.claude/agents/moai/manager-quality.md` | 44 | `## Skeptical Evaluation Mandate` (section header) |
| `.claude/agents/moai/evaluator-active.md` | 34 | `## Skeptical Evaluation Mandate` (section header) |

(Plus a description-field mention in evaluator-active.md L4: "Skeptical code
evaluator…" — not the 6-bullet block, ignored by the fingerprint scanner.)

The 6-bullet block in both files contains the same 6 mandate items; manual
diff confirms verbatim equality modulo trailing-newline. spec.md §1
Background calls for extraction to
`.claude/rules/moai/core/agent-common-protocol.md` §"Skeptical Evaluation
Stance"; LR-07 then asserts "exactly 1 occurrence" (the rule file).

After the M2 amendment, the lint counts must drop to:

- 1 occurrence in `agent-common-protocol.md` (canonical, allowed)
- 0 occurrences elsewhere (LR-07 violation if present)

### 1.5 LR-04 — dead hook config

`grep -n "matcher:" .claude/agents/moai/*.md` returns 0 matches in the live
tree (no agent declares a hook matcher today). LR-04 is **prophylactic**:
it prevents future regression where a contributor adds a hook matcher whose
referenced tool was already pruned from the tools list.

**Test strategy**: M3 RED authors a synthetic fixture (testdata file) with
`hooks:` block declaring `matcher: "Write|Edit"` while `tools:` excludes
both — confirms LR-04 fires.

### 1.6 LR-03 — missing `effort:` frontmatter

`grep -L "^effort:" internal/template/templates/.claude/agents/moai/*.md`
returns 0 matches today (no agent has `effort:` populated). LR-03 is
declared **warning** in baseline mode, **error** in `--strict` mode and is
the placeholder for SPEC-V3R2-ORC-003 (effort matrix) to upgrade.

### 1.7 LR-06 — `--deepthink` boilerplate

`grep -c "\-\-deepthink" .claude/agents/moai/*.md | grep -v ":0$"` count
gives ~22 lines across the 22-agent roster (P-A19 finding). Pattern: the
phrase "Use `--deepthink` for complex strategic decisions" or similar
appears as boilerplate in description / SCOPE block of most agents. LR-06
warns; --strict mode promotes to error. This pre-existing boilerplate
remains until ORC-001 M3 description-field hygiene lands; ORC-002's job is
to **surface the noise**, not silence it.

### 1.8 LR-08 — skill-preload drift

R5 audit P-A23 catalogued ~5 inconsistencies in same-category agents'
`skills:` fields (e.g., manager-quality preloads
`moai-foundation-core,moai-foundation-thinking` vs evaluator-active preloads
`moai-foundation-thinking,moai-foundation-core,moai-domain-skeptical-eval`
— ordering and superset variation). LR-08 detects categorical drift via
heuristic: agents in the same `name:`-prefix family (`manager-*`,
`expert-*`, `builder-*`) should share the same required-preload subset.

**Heuristic threshold**: drift is flagged when an agent omits a skill that
≥50% of its peers preload. This is a soft signal — warning only, never
strict-promotable in this SPEC (matrix authority belongs to ORC-001/003).

---

## 2. Go YAML Frontmatter Parsing — library evaluation

### 2.1 Existing usage in moai-adk-go

`grep -rln "yaml.v3\|gopkg.in/yaml" internal/ pkg/` confirms `gopkg.in/yaml.v3`
is the project standard (already in `go.mod`, line 17). Use cases:

- `internal/template/deployer.go` — agent file frontmatter parsing during
  `moai update` (precedent for the lint scanner's parser path)
- `internal/cli/spec_view.go` — SPEC document frontmatter
- `internal/session/state.go` (RT-004) — JSON marshalling, not YAML

**Decision**: use `yaml.v3` for ORC-002 frontmatter parsing. No new
dependency. Frontmatter structure (per `internal/template/deployer.go`
existing pattern):

```go
type AgentFrontmatter struct {
    Name        string   `yaml:"name"`
    Description string   `yaml:"description"`
    Tools       string   `yaml:"tools"`        // CSV string, e.g., "Read, Write, Agent"
    Model       string   `yaml:"model"`
    Effort      string   `yaml:"effort,omitempty"`
    Permission  string   `yaml:"permissionMode,omitempty"`
    Memory      string   `yaml:"memory,omitempty"`
    Skills      []string `yaml:"skills,omitempty"`     // YAML array
    Isolation   string   `yaml:"isolation,omitempty"`
    Hooks       []HookEntry `yaml:"hooks,omitempty"`
    Retired     bool     `yaml:"retired,omitempty"`
    Other       map[string]interface{} `yaml:",inline"`  // captures unknown fields
}
```

The `Tools string` (CSV) is per CLAUDE.local.md §12 ("`tools:`,
`allowed-tools:` → CSV string"). The `Skills []string` is the documented
exception (YAML array). This shape matches the live agent files.

### 2.2 Frontmatter delimiters

Files start with `---\n`, frontmatter content, `---\n`, body. Standard YAML
front-matter pattern. The scanner reads bytes 0..N until the second `---`
delimiter, parses with yaml.v3, then scans body line-by-line for LR-01,
LR-06, LR-07.

### 2.3 Edge cases

- Files starting with BOM (UTF-8 byte-order mark) — yaml.v3 does not strip
  BOM; pre-trim required.
- Multi-line `description:` field with `|` literal block — yaml.v3 handles
  natively; preserved as `\n`-joined string.
- Body fenced-code regions — the body scanner tracks fence state (``` vs
  ```lang) so LR-01 candidates inside fences are exempted (REQ-015).

---

## 3. Subcommand Registration — cobra integration evidence

### 3.1 Existing pattern in `internal/cli/root.go`

The `rootCmd` is constructed in `internal/cli/root.go:13-28`. New
subcommands are wired in `init()` (lines 39-82). Recent additions:

| Citation | Subcommand | Wiring line |
|----------|------------|-------------|
| `internal/cli/root.go:67` | `worktree` (subtree) | `rootCmd.AddCommand(worktree.WorktreeCmd)` |
| `internal/cli/root.go:69` | `statusline` | `rootCmd.AddCommand(StatuslineCmd)` |
| `internal/cli/root.go:72` | `astgrep` | `rootCmd.AddCommand(NewAstGrepCmd())` |
| `internal/cli/root.go:75` | `telemetry` | `rootCmd.AddCommand(telemetryCmd)` |
| `internal/cli/root.go:78` | `constitution` (CON-001) | `rootCmd.AddCommand(newConstitutionCmd())` |
| `internal/cli/root.go:81` | `state` (RT-004) | `rootCmd.AddCommand(newStateCmd())` |

**Decision**: ORC-002 follows the CON-001 / RT-004 pattern: a constructor
function `newAgentCmd()` returns the subcommand tree (`agent` parent +
`agent lint` child), registered in `root.go` after L81.

### 3.2 ExitCoder protocol

`cmd/moai/main.go:14-19` defines an `ExitCoder` interface; subcommands can
return errors implementing `ExitCode() int` to surface non-zero exit codes
(0/1/2/3 currently used by `moai worktree verify`). ORC-002 reuses this for
the four exit codes (0 clean, 1 violation, 2 malformed, 3 IO).

### 3.3 Tests pattern

`grep -ln "cobra\.Command\|cmd\.Execute" internal/cli/*_test.go` shows ~20
test files using `cobra.Command{}.Execute()` plus stdin/stdout capture.
Pattern: each subcommand has a `*_test.go` sibling that constructs a
sub-command instance, attaches a buffer for stdout, and runs the command.
ORC-002 tests follow the `internal/cli/constitution_test.go` and
`internal/cli/state_test.go` precedents.

---

## 4. Lint Tool UX Inspiration — design cross-check

### 4.1 eslint (JavaScript)

- Rule IDs use `<plugin>/<rule>` format. ORC-002 uses `LR-NN` (lint rule N)
  — simpler, no plugin namespace needed because the rule set is closed.
- Output formats: `stylish` (text), `json`, `compact`. ORC-002 supports
  `text` and `json` (REQ-010).
- Severity levels: `error`, `warn`, `off`. ORC-002 maps `error` and
  `warning`; `off` is implicit (rule absent). `--strict` flag promotes
  warnings.
- Exit codes: 0 (clean), 1 (errors), 2 (config error). ORC-002 extends to
  add 3 (IO).

### 4.2 golangci-lint (Go)

- Aggregator pattern: runs N linters and merges outputs. ORC-002 is a single
  linter (8 rules), not an aggregator.
- Output: `--out-format text|json|tab|checkstyle`. ORC-002 starts with
  text+json (REQ-010); checkstyle/tab can be added later without breaking
  schema (REQ-014 stable version).
- Configuration: `.golangci.yml`. ORC-002 v0.1.0 has no config file (rules
  hardcoded); future SPEC may add `.moai-agent-lint.yaml` overrides.

### 4.3 ast-grep (Markdown/code AST)

- Pattern-based rules; YAML rule definitions. Inspirational for LR-07
  fingerprinting but heavyweight for our 8-rule set.
- The `internal/astgrep/*.go` already integrated for code-side checks; we
  do NOT use ast-grep for agent-file lint because YAML+Markdown is
  simpler-parsed via stdlib.

### 4.4 markdownlint-cli2

- Rule naming convention `MD0NN`. ORC-002 uses `LR-NN` to avoid collision.
- Supports inline disable comments: `<!-- markdownlint-disable -->`.
  ORC-002 §spec.md Risk row 3 explicitly REJECTS escape hatches for LR-01
  (safety-critical) and warnings (no escape needed). LR-04 and LR-07 also
  have no escape hatch.

---

## 5. Fenced-Code Skip Algorithm — REQ-015 design

The body scanner needs a state machine to track fenced-code regions:

```
state IN_FENCE { start_marker: ```X, line_count: N }
state OUT_OF_FENCE
```

Fence open: line matches `^(```|~~~)\w*$` or `^(```|~~~)$`.
Fence close: line matches `^(```|~~~)$` AND the marker character matches
the open marker.

**Edge cases to test in M3 RED**:

- Nested fences (impossible in standard Markdown — close-marker matches
  outermost open) — verified.
- Indented fences (4-space indent) — treat as code, scan as if literal.
  Decision: treat indented blocks as fenced-equivalent for LR-01 skip.
- Inline-code (single backtick) — narrow scope; LR-01 candidate inside
  `` `AskUserQuestion` `` IS flagged in baseline policy because inline-code
  in agent body still implies the agent body is referring to the API as
  text-content. This is the **strict reading** of REQ-015 ("triple-backtick
  fenced code block" only).

### 5.1 Reference implementation pattern

```go
// Pseudocode for fence-aware body scan
inFence := false
fenceMarker := ""
for lineNo, line := range bodyLines {
    trimmed := strings.TrimLeft(line, " \t")
    if !inFence {
        if m := fenceOpenRE.FindString(trimmed); m != "" {
            inFence = true
            fenceMarker = m
            continue
        }
        // Out-of-fence: check LR-01
        if strings.Contains(line, "AskUserQuestion") {
            // emit violation unless agent is allowlisted (manager-brain rule)
        }
    } else {
        if trimmed == fenceMarker {
            inFence = false
            fenceMarker = ""
        }
    }
}
```

This is the M3 GREEN target for AC-ORC-002-10.

---

## 6. Existing Test Fixture Patterns — precedent

`internal/cli/constitution_test.go` (CON-001 test) uses the pattern:

```go
func TestConstitutionList(t *testing.T) {
    tmpDir := t.TempDir()
    // Stage testdata fixture into tmpDir
    src := "testdata/zone-registry.yaml"
    dst := filepath.Join(tmpDir, ".claude/rules/moai/core/zone-registry.md")
    // ... copy and execute subcommand against tmpDir
}
```

`testdata/` directories live next to `*_test.go`. ORC-002 follows suit:

```
internal/cli/agent_lint_test.go
internal/cli/testdata/agent_lint/
  fixture-clean.md            # baseline clean agent
  fixture-lr01-violation.md   # has AskUserQuestion in body
  fixture-lr01-fence-ok.md    # has AskUserQuestion only in fence
  fixture-lr02-violation.md   # tools list contains Agent
  fixture-lr04-dead-hook.md   # matcher refs absent tool
  fixture-lr07-duplicate.md   # second copy of Skeptical block
  fixture-malformed.md        # invalid YAML frontmatter
  fixture-orchestrator-allow.md  # manager-brain-style allowlist
```

Test isolation per CLAUDE.local.md §6: every test uses `t.TempDir()`; no
shared mutable state across subtests.

---

## 7. CI Integration — workflow shape

`grep -l "agent lint\|moai agent" .github/workflows/*.yaml` returns 0 (no
existing step). Existing CI workflows:

| Workflow file | Trigger | Step pattern |
|---------------|---------|--------------|
| `.github/workflows/ci.yaml` | push, PR | Lint → Test (3 OS) → Build (5 platforms) → CodeQL |
| `.github/workflows/release.yaml` | tag push | GoReleaser |
| `.github/workflows/release-drafter.yaml` | push:main, PR | Release Drafter |
| `.github/workflows/auto-merge.yaml` | dependabot PR + CI | auto-merge |
| `.github/workflows/labeler.yaml` | PR | label sync |

**Decision (REQ-011)**: ORC-002 adds a step to `.github/workflows/ci.yaml`
in the existing **Lint** job:

```yaml
- name: Run moai agent lint
  run: ./bin/moai agent lint
```

The step is required-status (already protected by branch protection rule
per CLAUDE.local.md §18.7), so failures block PR merge automatically.

Per spec.md §3 invocation modes, the CI runs the default scan (both trees);
the binary is built earlier in the same workflow.

---

## 8. Skeptical Block Fingerprint (LR-07) — semantic identity

The 6-bullet Skeptical-Evaluator Mandate block has the following identifying
shape (from manager-quality.md L44+ today):

```
## Skeptical Evaluation Mandate

The reviewer mode operates as a fresh-judgment auditor:

- Treat every claim as suspect until evidence is shown
- Demand reproducible verification, not assertions
- Consider the null hypothesis: did this change actually fix anything?
- Score quality as the harmonic mean of dimensions, not the average
- Reject when must-pass criteria fail, regardless of nice-to-have scores
- Surface contradictions; never silently override a prior rule
```

### 8.1 Fingerprint algorithm

To detect duplicates robustly across whitespace/punctuation drift:

1. **Header anchor**: find lines matching `^##\s+Skeptical Evaluation
   (Mandate|Stance)$`.
2. **Bullet count**: from the header to the next `^##\s` (or EOF), count
   `^-\s+` bullet lines.
3. **Bullet normalisation**: lowercase, strip whitespace, drop punctuation.
4. **Set hash**: SHA-256 over sorted bullet-set; compare across files.
5. **Match**: same hash + bullet count ≥6 = same canonical block.

### 8.2 First-occurrence rule (REQ-009)

The lint scans **rule files** before **agent files**. The first occurrence
encountered (in path-sorted order) is the canonical location; all
subsequent matches are LR-07 violations regardless of which is "intended"
canonical. After the M2 amendment, only `agent-common-protocol.md` will
contain the block; AC-12 verifies exactly 1 hit globally.

---

## 9. Open Questions for Audit

OQ-1 — manager-brain orchestrator carve-out mechanism:
- **Question**: should the lint exempt agents that declare `AskUserQuestion`
  in their `tools:` frontmatter (data-driven, Option A) or use a hardcoded
  allowlist (Option B)?
- **Recommendation**: Option A. The frontmatter-tools self-assertion is
  future-proof (any new orchestrator-class agent declares the tool); LR-02
  has its own `Agent`-specific check; LR-01 keying off `AskUserQuestion in
  tools-list` does not collide with LR-02 (different tokens). The exemption
  is logged in the JSON output as `exempt_reason: "tools-asserts-aue"` so
  audit traceability is preserved.
- **Resolution path**: M2 implementation; verified by AC-ORC-002-02
  (manager-brain produces 0 LR-01 violations even though body contains 11
  occurrences).

OQ-2 — LR-01 inline-code (single-backtick) handling:
- **Question**: should `` `AskUserQuestion` `` inside inline-code be
  exempted from LR-01? REQ-015 mentions only triple-backtick.
- **Recommendation**: NO exemption for inline-code. Strict reading of
  REQ-015. Inline-code in agent body is part of body prose; if the agent
  needs to reference the API by name, it must use the rule-file blocker
  pattern. This is the safer default.
- **Resolution path**: AC-ORC-002-10 negative fixture
  `fixture-lr01-inline-code.md` confirms inline-code IS flagged.

OQ-3 — LR-08 same-family threshold:
- **Question**: is "agents in same `name:`-prefix family" the right grouping
  for skill-preload-drift detection? Some experts cross-reference each
  other's preloads (e.g., expert-frontend includes design-related skills
  that expert-backend doesn't).
- **Recommendation**: ship LR-08 as **warning-only with a 50% peer-omission
  threshold**, no `--strict` promotion in this SPEC. Treat LR-08 as
  observational data for ORC-001/003 reviewers. If AC-ORC-002 tests show
  too much noise, downgrade to JSON-only output (no text-mode emit).
- **Resolution path**: M5 calibration after M3 RED tests pass.

OQ-4 — LR-04 hook matcher regex
- **Question**: hook matcher syntax in Claude Code uses regex pipes (e.g.,
  `Write|Edit|MultiEdit`). The lint must extract tool names from the regex
  to compare against tools list. Should it parse regex AST or use a simple
  `\|` split?
- **Recommendation**: simple `\|` split + token validation. If a matcher
  contains regex metacharacters beyond `|` (e.g., `.*` or `^Write$`), the
  lint emits a soft warning `LR-04-COMPLEX-MATCHER` rather than attempting
  full regex AST analysis. Matchers in practice are flat lists.
- **Resolution path**: M3 RED tests both the simple and complex cases.

OQ-5 — Two-tree drift LINT_TREE_DRIFT semantics:
- **Question**: REQ-017 says "warning when scan of `.claude/agents/moai/`
  produces a different violation set than scan of
  `internal/template/templates/.claude/agents/moai/`". How is "different"
  computed — file-set difference, per-file violation difference, or both?
- **Recommendation**: per-file violation-tuple difference. Group by relative
  path; for each path that exists in both trees, compare the
  `(rule, severity, line)` tuple set. Difference triggers LINT_TREE_DRIFT.
  File presence-difference (file in one tree only) is a separate signal
  flagged as `LINT_TREE_FILE_MISMATCH`.
- **Resolution path**: M4 implements; M5 verifies with synthetic
  template/local divergence fixture.

OQ-6 — Lint runtime budget enforcement:
- **Question**: REQ-007 / Constraint "<500ms per 22 files" — should the
  lint self-time and warn if exceeded, or is this a build-time CI metric?
- **Recommendation**: build-time CI metric (no in-binary self-timing). The
  CI step measures wall-clock; if it exceeds 1s consistently, the next
  SPEC adds `goroutine` parallelism. v0.1.0 ships single-threaded (≤500ms
  expected against 22-26 files).
- **Resolution path**: post-merge CI dashboard observation; no code action
  in this SPEC.

---

## 10. References (file:line citation summary)

Total citations: 35+ unique file:line references across the following
sources.

### 10.1 Live agent files (LR-01/LR-02 baseline)

- 2× `.claude/agents/moai/manager-strategy.md` (L59, L72)
- 4× `.claude/agents/moai/builder-agent.md` (L13, L61, L105 + body context)
- 2× `.claude/agents/moai/builder-skill.md` (L13, L61)
- 2× `.claude/agents/moai/builder-plugin.md` (L13, L81)
- 2× `.claude/agents/moai/expert-frontend.md` (L66, L93)
- 2× `.claude/agents/moai/expert-backend.md` (L66, L82)
- 2× `.claude/agents/moai/expert-devops.md` (L63, L78)
- 1× `.claude/agents/moai/manager-quality.md` (L102)
- 1× `.claude/agents/moai/manager-spec.md` (L180)
- 1× `.claude/agents/moai/manager-project.md` (L32)
- 1× `.claude/agents/moai/claude-code-guide.md` (L83)
- 1× `.claude/agents/moai/expert-security.md` (L13)
- 1× `.claude/agents/moai/expert-mobile.md` (L13)
- 11× `.claude/agents/moai/manager-brain.md` (L9, L29-L125 carve-out)

### 10.2 Skeptical block locations

- 1× `.claude/agents/moai/manager-quality.md` (L44)
- 1× `.claude/agents/moai/evaluator-active.md` (L34)

### 10.3 CLI plumbing

- `cmd/moai/main.go:14-19` (ExitCoder interface)
- `internal/cli/root.go:13-28` (rootCmd construction)
- `internal/cli/root.go:39-82` (init function — subcommand wiring)
- `internal/cli/root.go:67-81` (CON-001 / RT-004 precedent)
- `internal/cli/constitution.go` (newConstitutionCmd pattern)
- `internal/cli/constitution_test.go` (test fixture pattern)
- `internal/cli/spec_view.go` (frontmatter parsing precedent)
- `internal/template/deployer.go` (yaml.v3 usage in template subsystem)

### 10.4 Existing CI workflow

- `.github/workflows/ci.yaml` (Lint → Test → Build → CodeQL pipeline)

### 10.5 Rule and protocol references

- `.claude/rules/moai/core/agent-common-protocol.md` §User Interaction
  Boundary (FROZEN — protected by this SPEC)
- `.claude/rules/moai/core/askuser-protocol.md` (SPEC-ASKUSER-ENFORCE-001
  v1.0.0 canonical reference)
- `.claude/rules/moai/core/zone-registry.md` CONST-V3R2-036/037/038
  (User-Interaction-Boundary FROZEN entries)

### 10.6 Master design and audit pointers

- `docs/design/major-v3-master.md` §4.4 Layer 4 Orchestration
- `docs/design/major-v3-master.md:L1053` (§11.4 ORC-002 definition)
- `docs/design/major-v3-master.md:L963` (§8 BC-V3R2-004)
- `.moai/design/v3-redesign/research/r5-agent-audit.md` §Common Protocol
  compliance table (9 violations enumerated)
- `.moai/design/v3-redesign/synthesis/problem-catalog.md` (P-A01 CRITICAL,
  P-A04, P-A13, P-A18, P-A19, P-A23)
- `.moai/design/v3-redesign/synthesis/pattern-library.md` X-1 (Markdown +
  YAML Frontmatter as contract → enables lint)

### 10.7 Prior SPECs (carry-over context)

- SPEC-V3R2-CON-001 (zone registry — establishes the FROZEN/EVOLVABLE
  vocabulary and `moai constitution list` precedent)
- SPEC-V3R2-ORC-001 (agent roster consolidation — produces the v3r2 final
  roster used as the "clean baseline" target by AC-ORC-002-03)
- SPEC-V3R2-RT-004 (state subcommand — registration pattern reused)
- SPEC-V3R3-CI-AUTONOMY-001 (CI/CD auto-fix — precedent for adding
  required-status checks to ci.yaml)
- SPEC-ASKUSER-ENFORCE-001 (canonical AskUserQuestion protocol — the rule
  this SPEC enforces)

### 10.8 External library / pattern references

- spf13/cobra v1.10.2 (subcommand registration; already in go.mod)
- gopkg.in/yaml.v3 v3.0.1 (frontmatter parsing; already in go.mod)
- eslint output schema (rule_id, severity, file, line, message convention)
- golangci-lint --out-format design (text/json/checkstyle multiformat)
- ast-grep design (rejected — too heavy for 8 rules)
- markdownlint-cli2 disable-comment policy (rejected — no escape hatch
  matches our policy)

---

## 11. Hypothetical Run Output (M3 GREEN expectation)

Given the current tree (base `ab0fc4dda`), a successful M3 GREEN
implementation should produce text output approximately:

```
$ moai agent lint
Scanning .claude/agents/moai/ (26 files)...
Scanning internal/template/templates/.claude/agents/moai/ (26 files)...

ERRORS:
  .claude/agents/moai/manager-strategy.md:59  LR-01  literal AskUserQuestion in body
  .claude/agents/moai/manager-strategy.md:72  LR-01  literal AskUserQuestion in body
  .claude/agents/moai/builder-agent.md:61     LR-01  literal AskUserQuestion in body
  .claude/agents/moai/builder-agent.md:13     LR-02  Agent token in tools list
  .claude/agents/moai/builder-skill.md:61     LR-01  literal AskUserQuestion in body
  .claude/agents/moai/builder-skill.md:13     LR-02  Agent token in tools list
  .claude/agents/moai/builder-plugin.md:81    LR-01  literal AskUserQuestion in body
  .claude/agents/moai/builder-plugin.md:13    LR-02  Agent token in tools list
  .claude/agents/moai/expert-frontend.md:66   LR-01  literal AskUserQuestion in body
  .claude/agents/moai/expert-frontend.md:93   LR-01  literal AskUserQuestion in body
  .claude/agents/moai/manager-quality.md:102  LR-01  literal AskUserQuestion in body
  .claude/agents/moai/manager-quality.md:44   LR-07  duplicate Skeptical-Evaluator block
  .claude/agents/moai/expert-backend.md:66    LR-01  literal AskUserQuestion in body
  .claude/agents/moai/expert-backend.md:82    LR-01  literal AskUserQuestion in body
  .claude/agents/moai/manager-spec.md:180     LR-01  literal AskUserQuestion in body
  .claude/agents/moai/expert-devops.md:63     LR-01  literal AskUserQuestion in body
  .claude/agents/moai/expert-devops.md:78     LR-01  literal AskUserQuestion in body
  .claude/agents/moai/expert-security.md:13   LR-02  Agent token in tools list
  .claude/agents/moai/expert-mobile.md:13     LR-02  Agent token in tools list
  .claude/agents/moai/evaluator-active.md:34  LR-07  duplicate Skeptical-Evaluator block

WARNINGS:
  ... (LR-03 missing effort: 26 instances)
  ... (LR-06 --deepthink boilerplate: ~22 instances)
  ... (LR-08 skill-preload drift: ~5 instances)

SUMMARY: 9 LR-01 violations (across 9 agents), 5 LR-02, 2 LR-07, 0 LR-04
        + 53 warnings
EXIT CODE: 1
```

After ORC-001 (PR #811 + future cleanup of expert-mobile/expert-security
LR-02), and after this SPEC's M2 amendment of agent-common-protocol.md
extracting Skeptical Stance and removing duplicates, the output should
collapse to:

```
SUMMARY: 0 errors, 0 warnings (in non-strict mode)
EXIT CODE: 0
```

This is the **AC-ORC-002-03** target state.

---

End of research.
