# SPEC-GLM-EFFORT-REBALANCE-001 — Acceptance Criteria

Every criterion below is a Given-When-Then scenario with a binary outcome. Each names the command that decides it. Commands are run from the repository root.

## §B.0 Freshness gate — applies before AC-GER-004, AC-GER-006, AC-GER-013

[HARD] `make build` alone is not sufficient. It writes `bin/moai`; the criteria shell out to bare `moai`, which resolves through `PATH` to `~/go/bin/moai`. Run **both**:

```bash
make build && make install
test "$(moai version 2>&1 | grep -oE '[0-9a-f]{7,}' | head -1)" = "$(git rev-parse --short HEAD)" \
  && echo FRESH || echo "STALE — criteria below will not describe this change"
```

`make install` is `go install $(LDFLAGS) ./cmd/moai`. Do not substitute a bare `go install ./cmd/moai` (drops `LDFLAGS`, so `moai version` reports no usable commit and this gate becomes uncheckable), and do not `cp` over an existing `~/go/bin/moai` without `rm -f` first (CLAUDE.local.md §11: can yield exit 137 even at an identical SHA).

Evidence this gate is load-bearing: a binary one commit stale was observed emitting 12 `moai agent lint` findings naming exactly the three agents AC-GER-004 greps for, while a fresh build of identical source emitted zero.

---

## §A Acceptance matrix

| AC | Covers | Decided by |
|---|---|---|
| AC-GER-001 | REQ-GER-001, REQ-GER-002 | `go test ./internal/template/...` + source grep |
| AC-GER-002 | REQ-GER-003, REQ-GER-009 | `TestDefaultProfileMatrix_Shape` + `TestDefaultProfileMatrix_Monotone` |
| AC-GER-003 | REQ-GER-004 | `TestSessionGLMReasoningState` |
| AC-GER-004 | REQ-GER-006 | `moai agent lint` (LR-12) |
| AC-GER-005 | REQ-GER-005 | `harness_agents` vs matrix-row comparison (template copy) |
| AC-GER-006 | REQ-GER-007, REQ-GER-012, REQ-GER-013 | `moai model profile --json` on this checkout, config present |
| AC-GER-013 | REQ-GER-005, REQ-GER-012 | local `harness_agents` cells + resolution test |
| AC-GER-014 | REQ-GER-012 | non-matrix keys unchanged after the in-place edit |
| AC-GER-007 | test suite | `go test ./internal/template/... ./internal/cli/... ./internal/web/...` |
| AC-GER-008 | REQ-GER-008 | two greps across four prose surfaces + a per-line required-state table |
| AC-GER-009a | build integrity (host) | `go vet ./...` diffed against a recorded baseline |
| AC-GER-009b | build integrity (Windows test layer) | `GOOS=windows go vet ./...` diffed against a recorded baseline |
| AC-GER-010 | REQ-GER-010 | new test reading `EmbeddedTemplates()` vs `DefaultProfileMatrix()` |
| AC-GER-011 | mirror parity | `TestRuleTemplateMirrorDrift` + frontmatter pair diff |
| AC-GER-012 | template neutrality | neutrality guard |
| AC-GER-015 | REQ-GER-004 delivery | both env-injection helpers yield `high` |
| AC-GER-016 | REQ-GER-014 | shipped template `profiles` + `harness_agents` cells |
| §E DoD item 3 | REQ-GER-011 | completion report asserts no GLM cost reduction and no change-(1) GLM behavioral claim — a report-content check, not a command |

---

## §B Scenarios

### AC-GER-001 — the six cells carry the new values

**Given** `internal/template/profile_matrix.go` has been edited,
**When** the `PerformanceTierHigh` and `PerformanceTierMedium` blocks of `defaultProfileMatrix` are read,
**Then** `manager-spec` and `plan-auditor` carry `EffortLevelHigh` in both blocks, and `manager-docs` carries `EffortLevelLow` in both blocks.

```bash
# Expect exactly 0 lines: no EffortLevelMax on manager-spec/plan-auditor anywhere in the file,
# and no EffortLevelHigh on manager-docs outside the (unchanged) low column.
grep -nE '"(manager-spec|plan-auditor)":\s*\{Model: "opus", Effort: EffortLevelMax\}' \
  internal/template/profile_matrix.go
grep -nE '"manager-docs":\s*\{Model: "opus", Effort: EffortLevelHigh\}' \
  internal/template/profile_matrix.go
```

**PASS** when both greps produce zero matches (grep exit 1).

### AC-GER-002 — structural invariants survive

**Given** the six cells have changed,
**When** the matrix shape and monotonicity tests run,
**Then** the cell count is still 33, every model value is in `{opus, sonnet}`, every effort value is in `{low, medium, high, max}`, no `inherit` appears inside the matrix, and every row satisfies `high >= medium >= low`.

```bash
go test ./internal/template/... -run 'TestDefaultProfileMatrix_(Shape|Monotone)' -count=1
```

**PASS** when exit code is 0.

### AC-GER-015 — both GLM env-injection paths now yield `high`

**Given** `SessionGLMReasoningState()` returns `reasoning-high`,
**When** the two env-injection helpers are exercised,
**Then** both emit `ANTHROPIC_REASONING_EFFORT=high`.

```bash
go test ./internal/cli/... -run 'TestGLMReasoningEnvVars|TestGLMReasoningEnvVarsForEffort' -count=1 -v
```

**PASS** when exit code is 0 and both of these hold in the updated tests:

- `glmReasoningEnvVars()[ANTHROPIC_REASONING_EFFORT] == "high"` (sub-agent / `settings.local.json` parity wire).
- `glmReasoningEnvVarsForEffort("")[ANTHROPIC_REASONING_EFFORT] == "high"` (main-session launch path, no effort preference set). The five non-empty-effort cases in that table are unchanged.

> This is the criterion that actually observes the delivered wire value. AC-GER-003 covers the deriver in isolation; this one covers what `internal/cli/launcher.go` and the settings writer put on the env. Without it, nothing asserts that the change reaches the only surface it can reach.

### AC-GER-003 — GLM session state is reasoning-high

**Given** `SessionGLMReasoningState()` has been changed,
**When** it is called,
**Then** it returns `Name == GLMStateReasoningHigh`, `ThinkingEnabled == true`, and `ReasoningEffort == GLMReasoningEffortHigh`.

```bash
go test ./internal/template/... -run 'TestSessionGLMReasoningState' -count=1
```

**PASS** when exit code is 0 (the test body is updated to the new expectation as part of M5).

### AC-GER-004 — frontmatter agrees with the medium column (LR-12)

**Given** both the Go matrix and the six agent files have been edited,
**When** the agent linter runs,
**Then** it reports no LR-12 finding for `manager-spec`, `plan-auditor`, or `manager-docs`.

```bash
# Precondition — binary freshness (see §B.0 Freshness gate)
test "$(moai version 2>&1 | grep -oE '[0-9a-f]{7,}' | head -1)" = "$(git rev-parse --short HEAD)" \
  || echo "PRECONDITION FAILED: ~/go/bin/moai is stale — run make build \&\& make install"
moai agent lint 2>&1 | grep -E 'LR-12.*(manager-spec|plan-auditor|manager-docs)'
```

**PASS** when the precondition line prints nothing, the grep produces zero matches (exit 1), **and** `moai agent lint` itself exits 0.

> The freshness precondition is not ceremony here. A binary one commit stale was observed emitting 12 `moai agent lint` findings naming exactly these three agents, where a fresh build of identical source emitted zero (`plan.md` §B-9). Without the precondition this criterion fails against correct work, or passes against a binary that never saw it.

Cross-check on the raw values — all six files:

```bash
grep -H '^effort:' \
  .claude/agents/moai/manager-spec.md \
  .claude/agents/moai/plan-auditor.md \
  .claude/agents/moai/manager-docs.md \
  internal/template/templates/.claude/agents/moai/manager-spec.md \
  internal/template/templates/.claude/agents/moai/plan-auditor.md \
  internal/template/templates/.claude/agents/moai/manager-docs.md
```

**PASS** when `manager-spec` and `plan-auditor` read `effort: high` and `manager-docs` reads `effort: low`, in both trees.

### AC-GER-005 — harness_agents cells track their named rows

**Given** the matrix rows for `manager-docs` and `plan-auditor` have changed,
**When** the `harness_agents` block of the shipped `llm.yaml` is read,
**Then** `synthesize` carries `low` (the `manager-docs` row) and `research` carries `high` (the `plan-auditor` row), in both the `high` and `medium` columns.

```bash
grep -nE '^\s+(synthesize|research):' \
  internal/template/templates/.moai/config/sections/llm.yaml
```

**PASS** when every `synthesize` line under `harness_agents.high` and `harness_agents.medium` reads `{ effort: low }` and every `research` line under those two columns reads `{ effort: high }`. The `low` column's cells are unchanged.

Behavioural cross-check — the fall-through path resolves to the new row efforts:

```bash
go test ./internal/template/... -run 'TestResolveHarnessAgentModelEffort' -count=1
```

**PASS** when exit code is 0.

> Scope of what this test proves: its main loop passes `config.LLMConfig{Profile: "high"}` with **no** `HarnessAgents` map, so it exercises the **fall-through** path (config cell absent → `harnessClassRow` → matrix row) only. Its one config-present case is an explicit override assertion, not an agreement check. So this test confirms the matrix rows are right; it does **not** confirm the shipped or local config cells agree with them — that is what the greps above are for.

> Why this AC exists: `ResolveHarnessAgentModelEffort` reads the config cell first and falls through to the matrix row only when the cell is absent. A stale cell produces no error on either path — it silently gives shipped-file projects a different effort from bare projects.

### AC-GER-016 — the shipped template config carries the new cells

**Given** `internal/template/templates/.moai/config/sections/llm.yaml` has been edited,
**When** its `profiles.high` and `profiles.medium` blocks are read,
**Then** `manager-spec` and `plan-auditor` read `{ model: opus, effort: high }` and `manager-docs` reads `{ model: opus, effort: low }` in both.

```bash
sed -n '/^    high:/,/^    low:/p' internal/template/templates/.moai/config/sections/llm.yaml \
  | grep -E '(manager-spec|plan-auditor|manager-docs):'
```

**PASS** when the six lines read `manager-spec: { model: opus, effort: high }`, `plan-auditor: { model: opus, effort: high }`, and `manager-docs: { model: opus, effort: low }` under each of the `high` and `medium` columns. The `low` column's three lines are unchanged (`opus/low`, `opus/low`, `sonnet/low`).

> This exists because the template's `profiles:` cells were previously unowned. REQ-GER-001/002/003 bind the Go matrix, REQ-GER-012 binds the per-project runtime file, and REQ-GER-005 binds only `harness_agents` — leaving template lines for the three agents edited by the plan with no requirement behind them. Since no test compares the template against `DefaultProfileMatrix()`, omitting them would ship every new install a config that shadows the Go matrix straight back to the pre-change values. REQ-GER-014 closes that, and AC-GER-010's new embed test is its mechanical backstop.

### AC-GER-006 — the resolver reports the new cells ON THIS CHECKOUT, with its existing config in place

**Given** the binary has been rebuilt with `make build`, **and** `.moai/config/sections/llm.yaml` is present and populated on this checkout (it is gitignored, so it is not restored by any checkout operation — do NOT move, rename, or delete it to make this criterion pass),
**When** `moai model profile --json` runs under the active profile,
**Then** the payload reports `{"model":"opus","effort":"high"}` for `manager-spec` and `plan-auditor`, and `{"model":"opus","effort":"low"}` for `manager-docs`.

```bash
# Precondition 1 — binary freshness (see §B.0 Freshness gate)
test "$(moai version 2>&1 | grep -oE '[0-9a-f]{7,}' | head -1)" = "$(git rev-parse --short HEAD)" \
  || echo "PRECONDITION FAILED: ~/go/bin/moai is stale — run make build \&\& make install"
# Precondition 2 — the config that shadows the matrix must be present
test -s .moai/config/sections/llm.yaml || echo "PRECONDITION FAILED: config absent — this criterion is vacuous without it"
moai model profile --json \
  | jq -r '.agents[] | select(.agent=="manager-spec" or .agent=="plan-auditor" or .agent=="manager-docs")
           | "\(.agent) \(.model)/\(.effort)"'
```

**PASS** when the precondition line prints nothing and the output is exactly:

```
manager-spec opus/high
plan-auditor opus/high
manager-docs opus/low
```

> This is the criterion that fails on the config-shadow hazard. `llm.profiles` is consulted **before** the Go default, so a change that edits only `defaultProfileMatrix` leaves this output reporting the old `max` / `high` values. Running it against a pristine or absent config would pass vacuously and prove nothing — the populated config is the point of the test.

### AC-GER-013 — the harness classes resolve to the new efforts on this checkout

**Given** both the matrix rows and the `harness_agents` cells have been refreshed in the local config and the template,
**When** the harness resolution is exercised,
**Then** `synthesize` resolves to `low` (the `manager-docs` row) and `research` to `high` (the `plan-auditor` row).

```bash
# Precondition — binary freshness (see §B.0 Freshness gate)
test "$(moai version 2>&1 | grep -oE '[0-9a-f]{7,}' | head -1)" = "$(git rev-parse --short HEAD)" \
  || echo "PRECONDITION FAILED: ~/go/bin/moai is stale — run make build \&\& make install"
# Value-bearing: the local file is BLOCK style, so the value is on a following line
grep -A2 -nE '^\s+(synthesize|research):' .moai/config/sections/llm.yaml
# The template file is FLOW style, value on the same line
grep -nE '^\s+(synthesize|research):' internal/template/templates/.moai/config/sections/llm.yaml
go test ./internal/template/... -run 'TestResolveHarnessAgentModelEffort' -count=1
```

**PASS** when all three hold:

1. In the **local** file, under `harness_agents.high` and `harness_agents.medium`, the `synthesize:` key is followed by an `effort: low` line and the `research:` key by an `effort: high` line (a `model: ""` line may sit between the key and its effort — that is expected). **The `low` column is unchanged and still reads `research: … effort: low` / `synthesize: … effort: low`; do not read it as a failure.**
2. In the **template** file, the corresponding flow-style cells read `synthesize: { effort: low }` and `research: { effort: high }` in both columns.
3. The test exits 0.

> **Column order in the local file is alphabetical, not semantic.** The marshaller emits map keys sorted, so the columns appear `high:` → **`low:`** → `medium:`, and `grep -A2` prints no column headers. The three `research:` hits therefore arrive in the order high / **low** / medium — the middle one is the untouched `low` column reading `effort: low`, which looks like a failure and is not. Read the hits by line number against the column headers, or run `grep -n -E '^[[:space:]]*(high|medium|low):'` first to locate the boundaries. AC-GER-005 already carries this caveat for the template file; it binds here too.
>
> **The two files have different shapes and one command cannot read both.** The local `llm.yaml` is produced by a whole-struct re-marshal (`saveSection` / `saveLLMSection`, `plan.md` §B-0), so it is block-style and comment-free: `research:` / `model: ""` / `effort: max` on three separate lines. The shipped template is hand-written flow-style with `# <- plan-auditor row` annotations. A bare `grep '^\s+(synthesize|research):'` returns only the key lines and carries **no value information at all** in the local file — it cannot decide this criterion. `-A2` is what makes the block-style value visible.
>
> `ResolveHarnessAgentModelEffort` reads the config cell first, so a local config left stale shadows the matrix row for generated harness specialists by exactly the mechanism in AC-GER-006.

### AC-GER-014 — the config refresh preserved runtime state

**Given** the local `llm.yaml` cells were refreshed in place,
**When** the file's non-matrix keys are read,
**Then** `profile`, `performance_tier`, `team_mode`, the `glm.*` block, and any `agent_overrides` carry the same values they had before the edit.

```bash
# Whole-file diff against the pre-edit copy is the strongest form; the greps below
# are the fallback when no copy was taken.
diff .moai/specs/SPEC-GLM-EFFORT-REBALANCE-001/baseline-llm.yaml .moai/config/sections/llm.yaml
# Fallback: the glm block ENTIRE (not just base_url), plus the scalar keys
grep -n -A20 -E '^[[:space:]]*glm:' .moai/config/sections/llm.yaml
grep -nE '^[[:space:]]*(profile|performance_tier|team_mode|glm_env_var|mode):' .moai/config/sections/llm.yaml
grep -n -A6 'agent_overrides' .moai/config/sections/llm.yaml
```

**PASS** when the `diff` shows **only** the six `profiles` cells and the four `harness_agents` cells as changed lines, and nothing else. Using the fallback, PASS when every listed value matches the pre-edit reading recorded in `progress.md` §E.2.

> **Two corrections to earlier forms of this criterion, both of which made it decide nothing:**
>
> 1. **`base_url:` alone was too narrow.** The `glm:` block also carries `models.*`, `context_windows`, and the store-only `effort.{high,medium,low,fable}` tier fields (`internal/config/types.go`); a regeneration that preserved `base_url` while dropping the rest passed a `base_url`-only check. The block is now extracted whole.
> 2. **A two-space-anchored `sed` range returned zero lines.** The local file is 4-space indented and `glm:` sits well down the file, so `sed -n '/^  glm:/,/^  [a-z_]*:/p'` never matched — and empty output is indistinguishable from "verified unchanged", making it a false PASS of exactly the shape correction 1 fixed. The replacement is `grep -A20` with an indent-agnostic `^[[:space:]]*` anchor: it cannot return empty when `glm:` exists, and unlike a `sed` range it does not depend on which sibling key the marshaller happens to emit next. Confirm the output is non-empty before reading it as a pass.
>
> `effort_level` was also dropped from the scalar alternation — it is not a key in `LLMConfig`; the real keys are `mode`, `team_mode`, `glm_env_var`, `performance_tier`, and `profile`.
>
> A regeneration from the template rather than an in-place cell edit is exactly what this criterion exists to catch.

### AC-GER-007 — the affected test packages pass

**Given** all edits are in place,
**When** the affected packages are tested,
**Then** every test passes.

```bash
go test ./internal/template/... ./internal/cli/... ./internal/web/... -count=1
```

**PASS** when exit code is 0.

> The package set is load-bearing. `./internal/cli/...` is the **parent** package — `./internal/cli/agentlint/...` does not include it, and the two files that break under M2 live there: `glm_reasoning_overlay_test.go` (asserts `glmReasoningEnvVars()` yields `"max"`) and `glm_test.go` (the `{"empty → session default (reasoning max)", "", true, GLMReasoningEffortMax}` table case). `go vet` compiles tests without running them, so no vet-based criterion can see either. `./internal/web/...` is included because `g3_profile_matrix_test.go` hardcodes the `manager-spec` high- and medium-column cells (`plan.md` §B-3).

### AC-GER-008 — no prose still describes the old assignment

**Given** the four prose surfaces have been updated,
**When** they are scanned for the pre-change phrasing,
**Then** no surface still claims `manager-spec` / `plan-auditor` take `max`, or that `manager-docs` takes `high`.

Two greps, because the two rule files phrase the assignment differently and one pattern set cannot reach both.

```bash
# (a) model-policy.md, both trees — the tier table (122, 123), the max-cell census (196),
#     and the phase-weighting paragraph (198). The 198 pattern is separate because its
#     comma ("plan (`manager-spec`, `plan-auditor`) takes `max`") defeats the "+" form.
grep -rn 'Six matrix cells use\|max` on manager-spec\|max` on the plan rows\|manager-spec + plan-auditor\|takes `max`' \
  .claude/rules/moai/development/model-policy.md \
  internal/template/templates/.claude/rules/moai/development/model-policy.md

# (b) agent-authoring.md, both trees — the Effort-Level Calibration Matrix rows are a
#     markdown table, not prose, so they match none of the (a) patterns.
grep -rnE '^\| `(manager-spec|plan-auditor|manager-docs)` \| (medium|high|low) \|' \
  .claude/rules/moai/development/agent-authoring.md \
  internal/template/templates/.claude/rules/moai/development/agent-authoring.md
```

**PASS** when every surviving match describes the post-change assignment. A match is not automatically a failure — the sentence or row is read, and the criterion is whether the statement is now true. Any sentence still asserting the pre-change assignment is a FAIL.

Specifically, after the change these must read:

| Surface | Required post-change state |
|---|---|
| `model-policy.md:122` (high tier row) | `max` named for `manager-develop` + `super-advisor` only |
| `model-policy.md:123` (medium tier row) | no `max`; `high` on manager-spec / plan-auditor; `low` on manager-docs |
| `model-policy.md:196` | the max-cell census is **two**, not six |
| `model-policy.md:198` | the phase-weighting paragraph no longer says plan "takes `max`" |
| `model-policy.md:131` | already asserts max is reserved for manager-develop / super-advisor, high column only — becomes **true**; verify, do not rewrite |
| `agent-authoring.md:344` | `` | `manager-spec` | high | `` |
| `agent-authoring.md:349` | `` | `plan-auditor` | high | `` |
| `agent-authoring.md:347` | `` | `manager-docs` | low | `` — **already correct**, must stay |

> The earlier pattern set was vacuous on half its surfaces: run verbatim it matched only `model-policy.md` lines 122, 123, and 196, returned **zero** matches in `agent-authoring.md` in either tree, and missed `model-policy.md:198` entirely. `agent-authoring.md` is the one surface with no byte-parity CI behind it, and this grep was its stated mitigation — a grep that cannot match it mitigated nothing.

### AC-GER-009a — host static analysis introduces no new finding

**Given** a pre-change `go vet ./...` baseline was captured into `progress.md` §E.2 before the first edit,
**When** `go vet` runs over the whole module after the edits,
**Then** it reports no finding that is absent from that baseline.

```bash
go vet ./... > /tmp/vet-after.txt 2>&1; echo "exit=$?"
diff <(sort .moai/specs/SPEC-GLM-EFFORT-REBALANCE-001/baseline-vet-host.txt) <(sort /tmp/vet-after.txt)
```

**PASS** when the `diff` shows no line present in the after-run and absent from the baseline. An exit-0 after-run trivially passes; a non-zero after-run passes only when every finding also appears in the baseline.

### AC-GER-009b — Windows test layer introduces no new finding

**Given** a pre-change `GOOS=windows go vet ./...` baseline was captured into `progress.md` §E.2,
**When** `go vet` runs with `GOOS=windows` after the edits,
**Then** it reports no finding absent from that baseline.

```bash
GOOS=windows go vet ./... > /tmp/vet-after-win.txt 2>&1; echo "exit=$?"
diff <(sort .moai/specs/SPEC-GLM-EFFORT-REBALANCE-001/baseline-vet-windows.txt) <(sort /tmp/vet-after-win.txt)
```

**PASS** on the same rule as AC-GER-009a.

> `GOOS=windows go build ./...` does **not** satisfy this criterion — it never compiles `_test.go` files, so a broken test layer exits 0 under `build` and 1 under `vet`.
>
> **Why baseline-relative rather than absolute exit 0.** This checkout is shared by several live sessions. During this SPEC's plan audit, `go vet ./...` exited 1 on `internal/cli/preference/home_isolation_test.go` (`undefined: userHomeDir`) — a fault a parallel session introduced and then removed mid-audit, unrelated to this change. A re-run after the audit exited 0 with no output. An absolute assertion would have failed this criterion for something the change never touched, and could equally pass by luck. The baseline diff attributes findings to the change instead of to the tree's weather.

### AC-GER-010 — the embedded template FS carries the new cells

**Given** the template source has been edited,
**When** the `template` package's embedded filesystem is read and compared against `DefaultProfileMatrix()`,
**Then** the embedded `llm.yaml` reports the post-change cells for the three agents.

Add a test to `internal/template` that reads the embed directly — this is the only way to observe it:

```go
func TestEmbeddedLLMYAMLMatchesMatrix(t *testing.T) {
    fsys, err := EmbeddedTemplates()
    // EmbeddedTemplates() returns fs.Sub(embeddedRaw, "templates"), so the
    // "templates/" prefix is ALREADY STRIPPED — the key below carries no prefix.
    // ... read ".moai/config/sections/llm.yaml" from fsys,
    //     unmarshal profiles, compare high+medium cells for
    //     manager-spec / plan-auditor / manager-docs against
    //     DefaultProfileMatrix()
}
```

```bash
go test ./internal/template/... -run 'TestEmbeddedLLMYAMLMatchesMatrix' -count=1 -v \
  | tee /tmp/embed-test.txt
grep -c '^--- PASS: TestEmbeddedLLMYAMLMatchesMatrix' /tmp/embed-test.txt
```

**PASS** requires BOTH:

1. `go test` exits 0, **and**
2. the `grep -c` reports `1` — the test actually ran.

> **Why condition 2 exists.** `go test -run` with a selector matching nothing exits **0** and prints `[no tests to run]`. An exit-code-only criterion therefore passes on a tree where `TestEmbeddedLLMYAMLMatchesMatrix` was never written — which is precisely the vacuity this criterion was added to eliminate. The `--- PASS:` line is emitted only by a test that executed, so it fails on absence.
>
> Verified at authoring time: `grep -rn "TestEmbeddedLLMYAMLMatchesMatrix" internal/` returns nothing repo-wide. The test does not exist yet; writing it is part of the run phase.

> The previous formulation of this criterion could not decide it. `runModelProfile` (`internal/cli/model.go`) loads the on-disk project config through `config.NewLoader().Load(...)` and never touches the embedded FS, so `moai model profile --json` returns the same output whether or not the binary was rebuilt — the criterion passed identically in both states and observed nothing. Reading `EmbeddedTemplates()` is what makes the embed observable.
>
> The test also closes the gap named in `spec.md` §4 for the one file this SPEC touches: nothing else compares the shipped `llm.yaml` against `DefaultProfileMatrix()`. It is scoped to the six `profiles` cells this change moves, not a general parity guard.

### AC-GER-011 — mirror parity holds

**Given** both trees have been edited,
**When** the mirror-parity test runs and the frontmatter pairs are diffed,
**Then** `model-policy.md` is byte-identical across trees and each agent file's `effort:` matches its mirror.

```bash
go test ./internal/template/... -run 'TestRuleTemplateMirrorDrift' -count=1
diff <(grep '^effort:' .claude/agents/moai/manager-docs.md) \
     <(grep '^effort:' internal/template/templates/.claude/agents/moai/manager-docs.md)
```

**PASS** when the test exits 0 and each `diff` produces no output. Repeat the `diff` for `manager-spec` and `plan-auditor`.

> Scope note: `.moai/config/sections/llm.yaml` exists on a real checkout but is **gitignored per-project runtime state** (`plan.md` §B-1), so it is not a tracked mirror and legitimately carries runtime keys the template does not. A local↔template byte diff of that file is therefore not part of this criterion — the local copy's correctness is covered by AC-GER-006, AC-GER-013, and AC-GER-014 instead.

### AC-GER-012 — template tree stays neutral

**Given** files under `internal/template/templates/` were edited,
**When** the neutrality guard runs,
**Then** no SPEC ID, REQ token, internal date, or commit SHA has leaked into the template tree.

```bash
grep -rnE 'SPEC-[A-Z0-9-]+-[0-9]{3}|REQ-GER-[0-9]{3}|AC-GER-[0-9]{3}' \
  internal/template/templates/.moai/config/sections/llm.yaml \
  internal/template/templates/.claude/agents/moai/manager-spec.md \
  internal/template/templates/.claude/agents/moai/plan-auditor.md \
  internal/template/templates/.claude/agents/moai/manager-docs.md \
  internal/template/templates/.claude/rules/moai/development/model-policy.md \
  internal/template/templates/.claude/rules/moai/development/agent-authoring.md
go test ./internal/template/... -run 'TestTemplateNoInternalContentLeak' -count=1
```

**PASS** when the grep produces zero matches for tokens **introduced by this change** and the leak test exits 0.

---

## §C Edge cases

| Case | Expected behaviour |
|---|---|
| Another machine carries an `llm.yaml` from before this change | `ResolveAgentModelEffort` reads the stale config cell and returns the old effort — the change is inert there until that config is refreshed. This is the documented precedence (config before Go default), and it is a **known limitation of this SPEC**, not a regression: no migrator ships here (`spec.md` §4). Recovery: refresh the cells, or delete the `profiles:` block so the file tracks the Go matrix. |
| A checkout has no `llm.yaml` at all (fresh worktree) | Resolution falls through to the Go default, so the new cells apply immediately. AC-GER-006 would pass — which is exactly why it carries a precondition asserting the config IS present; passing on an absent config proves nothing. |
| A project sets `llm.agent_overrides["manager-docs"]` | The override wins over both the config cell and the Go default. Unchanged by this SPEC. |
| Profile is `low` | All three agents already resolve to `low`; the change is a no-op for that column. |
| Backend is GLM, agent is `manager-develop` | The coding-max override still forces `reasoning-max` at the per-agent layer. Unchanged. What the wire carries is the session-global value, now `reasoning-high`. |
| Unknown harness purpose class | Falls back to the `implement` class (`manager-develop` row), which this SPEC does not touch. |

---

## §D Verification gates

0. **Baseline gate** — the two `go vet` baselines and the pre-edit `llm.yaml` copy are recorded in `progress.md` §E.2 **before** the first edit. AC-GER-009a/009b/014 cannot be judged without them.
1. **Build gate** — `make build && make install` completes and the §B.0 freshness check reports FRESH before any `moai`-invoking criterion is evaluated.
2. **Unit gate** — AC-GER-002, AC-GER-003, AC-GER-005, AC-GER-007, AC-GER-010, AC-GER-011, AC-GER-012, AC-GER-015, AC-GER-016 all exit 0.
3. **Static gate** — AC-GER-009a and AC-GER-009b introduce no finding absent from their baselines.
4. **Integration gate** — AC-GER-004, AC-GER-006, and AC-GER-013 report the new cells from the freshly installed binary, with the local config present.
5. **Config-integrity gate** — AC-GER-014 confirms the in-place edit preserved every non-matrix key.
6. **Prose gate** — AC-GER-008 leaves no surface asserting the pre-change assignment, on both greps.

## §E Definition of Done

- All seventeen criteria in §A evaluated, each with the command run and its verbatim output cited.
- Any criterion not evaluated is reported as a Gap, not silently omitted.
- The completion report does **not** claim a GLM cost reduction and does **not** claim that change (1) alters delivered GLM behavior (REQ-GER-011).

## §F Not mechanically checkable — stated as residual risk

These are named here rather than dressed up as criteria, because no command in this repository can decide them.

1. **Does z.ai honour `ANTHROPIC_REASONING_EFFORT`?** `internal/cli/glm.go` marks the shim's consumption UNVERIFIED. If z.ai requires `reasoning_effort` in the request body instead, the env var is inert and change (2) delivers nothing. Settling this needs a live z.ai session, which is out of scope (`spec.md` §4).
2. **Does GLM spend actually fall?** Observable only in billing over time, not in CI. No criterion asserts it.
3. **Is `high` sufficient for plan-phase quality, and `low` for sync-phase?** A judgment about output quality. The reversal path is `llm.agent_overrides`, per-agent, with no code change.
4. **AC-GER-008 needs a human read.** The grep locates the sentences; whether each rewritten sentence is now *true* is a reading, not an exit code.
5. **Other machines are not covered.** Every install carrying a pre-change `profiles:` block in its gitignored `llm.yaml` keeps resolving the old cells, and no criterion here can reach them. AC-GER-006 verifies one checkout — this one. Whether `moai update`'s template-managed config sync would refresh a populated `llm.yaml`, overwrite it whole, or prompt was NOT executed and observed during planning, so it is not claimed as the migration path (`spec.md` §4).
6. **The template mirror has no mechanical guard.** Nothing asserts the shipped `llm.yaml` agrees with `DefaultProfileMatrix()`, so AC-GER-005's grep is the only thing standing between a correct edit and a silent re-drift on the next matrix change.
