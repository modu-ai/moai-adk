# SPEC-GLM-EFFORT-REBALANCE-001 — Implementation Plan

> Ordered by decision-reversibility. The decisions most likely to change sit at the top (M1-M2); the mechanical edits that simply follow from them sit at the bottom (M4-M6). Review attention belongs at the top.

## §A Context

Two source-of-truth edits (six matrix cells + one return value) plus the coupled surfaces that mechanically depend on them. The full rationale, the delivery-wire distinction, and the accepted global-scope trade-off are in `spec.md` §1.

## §B Known issues discovered during planning

These were verified against the tree before this plan was written, and each one changes what the run phase has to touch.

### B-0 — the local config shadows the Go matrix, and can make the whole change inert

This is the finding that reorders the plan. `ResolveAgentModelEffort` consults `llm.profiles` from config **before** the Go default (`internal/template/profile_matrix.go`, precedence steps 2 and 3), so on any machine whose `.moai/config/sections/llm.yaml` already carries a populated `profiles:` block — which is every machine that has run `moai init` — editing `defaultProfileMatrix` alone changes nothing. `ResolveHarnessAgentModelEffort` shadows the harness rows the same way.

Verified on the primary checkout: `llm.yaml` is 5121 bytes, carries `manager-spec: {model: opus, effort: max}` under `profiles.high`, and `moai model profile --json` reports `profile: medium` with `manager-spec effort: max`.

No existing code path refreshes those blocks:

| Writer | What it writes | Touches `profiles:` / `harness_agents:`? |
|---|---|---|
| `ApplyProfile` | the `profile:` line | no |
| `ApplyPerformanceTier` | the `performance_tier:` line | no |
| `stripRetiredLLMKeys` | removes `plan_type` + `claude_models` | no |
| `saveLLMSection` (`internal/cli/glm.go`) | marshals the whole loaded `LLMConfig` | re-serializes whatever was loaded — preserves stale cells |
| `saveSection` (`internal/config/manager.go`), reached from `internal/web/agentfm.go` | whole-file struct re-marshal | same: preserves stale cells, and additionally **drops every comment** in the file |

The last row explains the local file's shape: it is a struct re-marshal, so the local `llm.yaml` is **block-style with no comments**, whereas the shipped template is hand-written flow-style (`{ effort: high }`) with `# <- manager-docs row` annotations. Any criterion that greps the two files must not assume one shape (D6 / AC-GER-013).

The mechanism REQ-GER-012 selects is an **in-place refresh of the affected cells**, preserving every other key. Milestone M0 carries it, and AC-GER-006 is written to fail if the change is not observable on this checkout with its existing config in place.

### B-1 — the local `llm.yaml` is gitignored, not absent

The delegation brief named a local ↔ template mirror pair for `llm.yaml`. The earlier planning pass reported the local file as non-existent; that reading came from a fresh worktree and the stated reason was wrong. The file **exists** on the primary checkout and is gitignored:

```
git ls-files .moai/config/sections/llm.yaml       → (empty)
git check-ignore -v .moai/config/sections/llm.yaml
  → .gitignore:188:.moai/config/sections/llm.yaml
```

So it is untracked **per-project runtime state**, which is why a fresh worktree has no copy while a real install does. The conclusion the earlier pass drew still stands — there is no *tracked* mirror pair, and only the template copy is under version control — but the reason matters: a run-phase agent told "does not exist" would find a 5121-byte file and not know which to trust. Practical consequences:

- Only `internal/template/templates/.moai/config/sections/llm.yaml` is committed; the local copy is edited but never committed.
- An acceptance criterion phrased as "local and template mirrors are byte-consistent" is still unsatisfiable — the local file legitimately carries runtime keys the template does not.
- Editing the local file IS required despite it being untracked, because of B-0.

### B-2 — `harness_agents` cells silently split from their matrix row

`internal/template/templates/.moai/config/sections/llm.yaml` carries a `harness_agents` block whose cells are annotated with the matrix row they mirror:

```
synthesize:  { effort: high }    # <- manager-docs row
research:    { effort: max }     # <- plan-auditor row
```

The **cells** exist in both the `high` and `medium` columns; the **annotations do not** — the `medium` column carries the same `synthesize` / `research` cells bare, with no `# <- ... row` comment. Any requirement or criterion gated on the comment would silently exempt half the work, which is why REQ-GER-005 binds the `harnessClassRow` mapping rather than the presence of a comment.

`ResolveHarnessAgentModelEffort` reads the config cell **first** and only falls through to `harnessClassRow` -> the matrix row when the cell is absent. The file's own comment names the hazard: a stale cell means a project carrying the shipped file gets the old effort while a project without it gets the new one — a split that produces no error anywhere.

This was not in the delegation brief and is the highest-consequence omission found. It is REQ-GER-005.

### B-3 — four test files outside the named set assert the changed values

- `internal/template/profile_matrix_test.go` `TestResolveHarnessAgentModelEffort` (the `want` map) asserts `HarnessClassSynthesize -> high` (manager-docs row) and `HarnessClassResearch -> max` (plan-auditor row) at profile `high`. Two assertions break.
- `internal/web/g3_profile_matrix_test.go` hardcodes `manager-spec` cells: the high-column effort select is asserted to be `max`, and the medium-column default is asserted to be `opus/max` in the override-clearing case. Both break.
- `internal/cli/glm_reasoning_overlay_test.go` asserts `glmReasoningEnvVars()[ANTHROPIC_REASONING_EFFORT] == "max"`. Breaks under M2.
- `internal/cli/glm_test.go` has the table case `{"empty → session default (reasoning max)", "", true, template.GLMReasoningEffortMax}` for `glmReasoningEnvVarsForEffort`. Breaks under M2 — this is the main-session empty-effort fallback of B-8.

Neither `internal/web/...` nor the parent `internal/cli/` package was in the delegation brief's test list. The parent package matters specifically: `./internal/cli/agentlint/...` does **not** include `./internal/cli/`, and `go vet` compiles tests without running them, so a package set that omits `./internal/cli/...` cannot see either failure.

### B-4 — `internal/template/model_policy_test.go` does not need editing

It asserts only `EffortLevelMax == "max"` (constant identity). No cell assertion. Harmless to run, but it is not a required edit.

### B-5 — `agent-authoring.md` calibration table is a fourth prose surface

`.claude/rules/moai/development/agent-authoring.md` § Effort-Level Calibration Matrix carries a "Default effort (medium column)" table with rows for all three agents, and a template mirror. It is **already stale** against the current medium column, and stays stale for two of three rows unless updated. Not enrolled in byte-parity CI, so nothing catches it mechanically — which is exactly why it needs an explicit step.

### B-6 — nothing guards the template mirror against the Go matrix

No test asserts that `internal/template/templates/.moai/config/sections/llm.yaml` agrees with `template.DefaultProfileMatrix()`. The two are aligned by hand and by the file's own comment. That absence is why B-2 was reachable at all, and why M3 has to be done deliberately rather than trusted to CI. Adding the guard is out of scope (`spec.md` §4) but recorded here so the run phase knows the mirror edit has no safety net.

### B-8 — `SessionGLMReasoningState()` has two consumer paths, not one

The earlier planning pass called `glmReasoningEnvVars()` the sole non-test consumer. That is false. `SessionGLMReasoningStateForEffort` falls back to `SessionGLMReasoningState()` on its empty-effort branch (`glm_effort_overlay.go`), and that function is consumed by `glmReasoningEnvVarsForEffort` (`internal/cli/glm.go`), called from `internal/cli/launcher.go` on the **main-session launch path**.

So change (2) moves two things: the sub-agent / `settings.local.json` parity wire, and the main session's fallback when no effort preference is set. M2's scope statement and the `glm_test.go` empty-effort case (B-3) both follow from this.

### B-9 — a stale `~/go/bin/moai` makes the `moai`-invoking criteria lie

`make build` writes `bin/moai`; it does not touch `~/go/bin/moai`, which is what a bare `moai` on `PATH` resolves to. Verified on the primary checkout: the installed binary was built at commit `c55c61aa5` against a HEAD of `88275dac0`, and that stale binary's `moai agent lint` emitted **12 findings naming exactly the three agents AC-GER-004 greps for**, while a fresh build of identical source emitted zero.

A criterion that shells out to bare `moai` therefore measures whatever was last installed, not the change. §D now mandates `make build && make install`, and every `moai`-invoking criterion carries a freshness precondition comparing `moai version`'s commit against `git rev-parse --short HEAD`.

### B-10 — docs-site drift is pre-existing

`docs-site/content/*/advanced/profile-matrix.md` already disagrees with the committed Go matrix. Out of scope per `spec.md` §4; recorded here so the run phase does not treat it as newly introduced.

---

## §C Milestones

### M0 — Local config refresh (the decision that makes the rest observable)

This sits first because it is the decision most likely to be revisited: it selects *how* an install picks up the new cells, and getting it wrong makes M1 inert rather than wrong-looking.

In `.moai/config/sections/llm.yaml` on this checkout — gitignored, never committed, so it appears in no diff:

- `profiles.high` and `profiles.medium`: the six cells of M1.
- `harness_agents.high` and `harness_agents.medium`: `synthesize` → `low`, `research` → `high`.
- **Leave every other key untouched** — `profile`, `performance_tier`, `team_mode`, `glm.*`, `agent_overrides`. An in-place cell edit, not a regeneration from the template.

Do this edit only after M1 and M3 have settled the target values, so the three surfaces are written from one decision rather than three. Ordering it first in the plan reflects review priority, not execution order — the run phase may sequence M1 → M3 → M0 and then verify.

Recovery note for a drifted install: because the Go default is consulted only when the config cell is absent, deleting the `profiles:` block makes a file track the Go matrix permanently. That is a documented recovery, not this milestone's mechanism — it also discards any deliberate per-project customization.

### M1 — Matrix cells (the change itself)

`internal/template/profile_matrix.go`, `defaultProfileMatrix`:

| Column | Agent | From | To |
|---|---|---|---|
| `PerformanceTierHigh` | `manager-spec` | `EffortLevelMax` | `EffortLevelHigh` |
| `PerformanceTierHigh` | `plan-auditor` | `EffortLevelMax` | `EffortLevelHigh` |
| `PerformanceTierHigh` | `manager-docs` | `EffortLevelHigh` | `EffortLevelLow` |
| `PerformanceTierMedium` | `manager-spec` | `EffortLevelMax` | `EffortLevelHigh` |
| `PerformanceTierMedium` | `plan-auditor` | `EffortLevelMax` | `EffortLevelHigh` |
| `PerformanceTierMedium` | `manager-docs` | `EffortLevelHigh` | `EffortLevelLow` |

Six cells. Nothing else in the map moves.

The `defaultProfileMatrix` doc comment carries a phase-weighted-override paragraph naming which phase takes which level, and a GLM-observability paragraph claiming "the run row is unchanged ... and so is review. Plan and sync do move." Both need re-statement: plan now steps down rather than up, and the GLM sentence should not be read as a claim about delivered behavior (see `spec.md` §1.3).

Monotonicity check before proceeding — `TestDefaultProfileMatrix_Monotone` ranks `model*10 + effort`:

- `manager-spec` / `plan-auditor`: high `opus/high` (12) >= medium `opus/high` (12) >= low `opus/low` (10). Holds.
- `manager-docs`: high `opus/low` (10) >= medium `opus/low` (10) >= low `sonnet/low` (0). Holds.

### M2 — GLM session-global reasoning state (second-highest reversibility)

`internal/template/glm_effort_overlay.go`:

- `SessionGLMReasoningState()` returns `glmReasoningHigh` instead of `glmReasoningMax`.
- Its doc comment currently derives the max return from `manager-develop` being "the representative code-producing active spawn". That derivation no longer holds and is replaced with the operator's cost rationale. Keep the existing delivery-granularity note and the UNVERIFIED-shim note — both remain true.
- `SessionGLMReasoningStateForEffort`'s doc comment describes the empty-effort fallback as "the coding-max session default". That phrase becomes wrong; restate it as the session default without the coding-max derivation. The function body does **not** change.

**Two delivery surfaces move, not one** (B-8):

| Surface | Path | Effect of M2 |
|---|---|---|
| Sub-agent / `settings.local.json` parity wire | `glmReasoningEnvVars()` → `SessionGLMReasoningState()` | `max` → `high` |
| Main-session launch, no effort preference set | `launcher.go` → `glmReasoningEnvVarsForEffort("")` → `SessionGLMReasoningStateForEffort("")` → `SessionGLMReasoningState()` | `max` → `high` |

A non-empty main-session effort is unaffected — it collapses that effort directly and never reaches the fallback.

Note that `glmCodingMaxOverrideAgents` (the `{manager-develop}` set) and `ResolveGLMReasoning` are **not** touched. The per-agent override still exists; it simply is not what either wire carries.

### M3 — Config mirror: profiles + harness_agents (B-2)

`internal/template/templates/.moai/config/sections/llm.yaml`:

- `profiles.high` and `profiles.medium`: the same six cells as M1.
- `harness_agents.high` and `harness_agents.medium`: `synthesize` -> `low` (tracks manager-docs), `research` -> `high` (tracks plan-auditor).
- The column-intent comment block (the paragraph naming which phase takes `max`) needs the same re-statement as M1's doc comment.

Do not touch `harness_agents.*.implement` (manager-develop row) or `verify-judge` (sync-auditor row) — both rows are unchanged.

### M4 — Agent frontmatter, both trees (mechanical; LR-12 forces it)

Six files, `effort:` only:

| File | From | To |
|---|---|---|
| `.claude/agents/moai/manager-spec.md` | `max` | `high` |
| `.claude/agents/moai/plan-auditor.md` | `max` | `high` |
| `.claude/agents/moai/manager-docs.md` | `high` | `low` |
| `internal/template/templates/.claude/agents/moai/manager-spec.md` | `max` | `high` |
| `internal/template/templates/.claude/agents/moai/plan-auditor.md` | `max` | `high` |
| `internal/template/templates/.claude/agents/moai/manager-docs.md` | `high` | `low` |

The template tree pins frontmatter to the **medium** column by policy, which is why all six land on the medium-column value.

### M5 — Tests (mechanical; each is a one-value update)

| File | What changes |
|---|---|
| `internal/template/profile_matrix_test.go` | `TestResolveHarnessAgentModelEffort` `want`: `synthesize` -> `EffortLevelLow`, `research` -> `EffortLevelHigh`. Any explicit-cell assertion elsewhere in the file. |
| `internal/template/glm_effort_overlay_test.go` | `TestSessionGLMReasoningState`: expect `GLMStateReasoningHigh` + `GLMReasoningEffortHigh`. The `TestSessionGLMReasoningStateForEffort` empty-effort fallback case follows the same value. |
| `internal/web/g3_profile_matrix_test.go` | The `manager-spec` high-column effort assertion (`max` -> `high`) and the medium-column override-clearing case (`opus/max` -> `opus/high`). |
| `internal/cli/glm_reasoning_overlay_test.go` | The `glmReasoningEnvVars()` assertion: `"max"` -> `"high"`, plus the "coding-max session default" wording in the message and the file's header comment. |
| `internal/cli/glm_test.go` | The `glmReasoningEnvVarsForEffort` table case `{"empty → session default (reasoning max)", "", true, template.GLMReasoningEffortMax}` -> `GLMReasoningEffortHigh`, with the case name updated. The other five cases feed effort explicitly and are unchanged. |

`CollapseClaudeEffortToGLM`'s own table-driven cases pass unchanged — they feed effort explicitly and do not read the matrix. `internal/template/model_policy_test.go` needs no edit (B-4).

### M6 — Prose, both trees (mechanical; last because it follows from M1)

| File | Lines / section |
|---|---|
| `.claude/rules/moai/development/model-policy.md` | The tier table's `high` and `medium` effort-baseline cells; the "six matrix cells use `max`" sentence; the phase-weighting paragraph. The `max is reserved for manager-develop / super-advisor, high column only` sentence becomes *true* after the change — verify rather than rewrite it. |
| `internal/template/templates/.claude/rules/moai/development/model-policy.md` | Byte-identical edits (byte-parity CI). |
| `.claude/rules/moai/development/agent-authoring.md` | § Effort-Level Calibration Matrix rows for the three agents (B-5). |
| `internal/template/templates/.claude/rules/moai/development/agent-authoring.md` | Same rows. |

---

## §D Build step — non-optional

[HARD] **The run phase must finish with `make build && make install`.** Both halves are load-bearing and neither substitutes for the other.

- `make build` re-embeds the templates. They reach the binary through `//go:embed all:templates` (`internal/template/embed.go`); there is no generated `embedded.go`, but the binary carries a compiled-in copy of `templates/`, so a template edit is invisible until a rebuild. Note that `EmbeddedTemplates()` returns `fs.Sub(embeddedRaw, "templates")` — the prefix is stripped, so AC-GER-010's test reads `.moai/config/sections/llm.yaml`, not `templates/.moai/...`.
- `make install` puts that binary where the criteria actually look. `make build` writes `bin/moai`; every `moai`-invoking criterion shells out to bare `moai`, which resolves through `PATH` to `~/go/bin/moai`. Rebuilding without installing leaves the criteria measuring whatever was last installed — verified on this checkout as a binary 1 commit stale whose `moai agent lint` emitted 12 findings against the three agents AC-GER-004 greps for, where a fresh build of identical source emitted zero (B-9).

`make install` runs `go install $(LDFLAGS) ./cmd/moai`, which carries the version ldflags. Do **not** substitute a bare `go install ./cmd/moai` (it drops `LDFLAGS`, so `moai version` reports `Commit=none` and the freshness precondition below becomes uncheckable), and do not substitute a bare `cp bin/moai ~/go/bin/moai` over an existing binary — CLAUDE.local.md §11 records that doing so can yield exit 137 (SIGKILL) even at an identical SHA. If copying rather than installing, `rm -f ~/go/bin/moai` first so the inode is replaced.

**Freshness precondition.** Every `moai`-invoking criterion (AC-GER-004, AC-GER-006, AC-GER-013) asserts first that the installed binary is the current tree:

```bash
test "$(moai version 2>&1 | grep -oE '[0-9a-f]{7,}' | head -1)" = "$(git rev-parse --short HEAD)"
```

Order: edit sources and templates -> `make build && make install` -> confirm freshness -> run the verification batch. Running the batch first produces a result that does not describe the change, in either direction.

## §E Self-verification

Run as a single-turn parallel batch (independent, read-only):

```
go test ./internal/template/... ./internal/cli/... ./internal/web/... -count=1
go vet ./...
GOOS=windows go vet ./...
moai model profile --json  # against this checkout's existing llm.yaml — not a pristine config
moai agent lint            # LR-12
```

`./internal/cli/...` is the parent package, not `./internal/cli/agentlint/...` — the narrower path excludes `glm_test.go` and `glm_reasoning_overlay_test.go`, which is where M2's breakage lives (B-3).

Cite exit codes and verbatim output per `verification-claim-integrity.md` §2. `GOOS=windows go build` is **not** a substitute for `GOOS=windows go vet` — the former never compiles `_test.go`.

**Record a pre-change vet baseline.** This checkout is shared by several live sessions, so an absolute `go vet` exit-0 assertion is hostage to unrelated churn: during this SPEC's own plan audit, `go vet ./...` exited 1 on `internal/cli/preference/home_isolation_test.go` (`undefined: userHomeDir`) — a failure introduced and then removed by a parallel session mid-audit, unrelated to this change. Capture both vet runs into `progress.md` §E.2 **before** the first edit, so AC-GER-009a/009b can be judged as "no NEW finding" rather than "exit 0".

## §F Risks

| Risk | Why it matters | Mitigation |
|---|---|---|
| Local `llm.yaml` left carrying the old `profiles:` block | Config precedence beats the Go default, so **every reported cell and every surface derived from the resolver** — `moai model profile`, the web console's preview, harness-specialist generation — keeps returning the old values, and nothing reports that the Go edit never took effect. (The delivered *agent* effort does still change once M4 lands, because frontmatter is the load-bearing channel on the `Agent` tool path — so the failure is partial and inconsistent, which is worse than uniform inertness.) | M0 is a required milestone; AC-GER-006 runs against this checkout with its existing config in place, so a config-only miss fails the criterion instead of passing vacuously |
| Verification run against a stale `~/go/bin/moai` | `make build` alone does not update the binary the criteria invoke; a stale binary emits findings for source it never saw (B-9) | §D mandates `make build && make install`; AC-GER-004 / 006 / 013 each assert `moai version`'s commit equals `git rev-parse --short HEAD` |
| `harness_agents` cells left stale | Silent split: file-carrying projects and bare projects resolve different efforts, with no error on either path | M3 + M0 are required milestones, and AC-GER-005 / AC-GER-013 check the cells against the rows mechanically |
| Other installs keep the old cells | Their gitignored `llm.yaml` still shadows the Go matrix; no migrator ships with this change | Named as a known limitation in `spec.md` §4 rather than assumed away; the absent-cell fallback is the documented recovery |
| Verification run against a stale binary | `moai model profile --json` reads the compiled-in template; a pre-`make build` run reports the old cells and looks like a failure, or reports new cells for the wrong reason | §D ordering; AC-GER-010 asserts source/embedded agreement |
| `GOOS=windows go build` mistaken for test-layer verification | A duplicate symbol or a broken test helper compiles clean under `build` and fails under `vet` | AC-GER-009b names `vet` explicitly and requires the exit code |
| Prose left describing the old assignment | Four prose surfaces across two trees; only `model-policy.md` is byte-parity-enforced, so `agent-authoring.md` drift is caught by nothing | M6 enumerates all four; AC-GER-008 greps for the stale phrasing |
| Scope creep into the docs-site table | The 4-locale table is visibly wrong and invites a drive-by fix that triples the diff | Declared out of scope in `spec.md` §4 with the reason recorded |

## §G Anti-patterns

- Editing `defaultProfileMatrix` and stopping there. Config precedence beats the Go default, so the change would be inert on every machine that has a populated `llm.yaml` — and nothing would report it (B-0).
- Regenerating the local `llm.yaml` from the template instead of refreshing the affected cells. That discards `profile`, `team_mode`, and the GLM settings.
- Treating `moai update` as the migration path for other installs without having run it and observed the result on a populated config (`spec.md` §4).
- Re-deriving the matrix back onto the benchmark cost anchor. The `high`/`medium` columns are a deliberate operator policy, not a derivation; both the Go doc comment and `llm.yaml` say so explicitly.
- Touching `sync-auditor` because "it sounds like sync". The code's phase mapping assigns it to review.
- Editing only the working-tree `.claude/` copy. Template-First: the template source is edited, then mirrored.
- Claiming the change reduces GLM cost. Neither the shim consumption nor the spend delta is verified here (`spec.md` §1.3).

## §H Cross-references

- `spec.md` §1.3 — what moves the GLM wire and what does not.
- `spec.md` §4 — the exclusions this plan honours.
- `acceptance.md` — the binary-testable criteria.
- CLAUDE.local.md §2 (Template-First), §6 (cross-platform verification), §15 (template language neutrality).
