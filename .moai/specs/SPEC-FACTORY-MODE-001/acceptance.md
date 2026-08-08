# SPEC-FACTORY-MODE-001 — Acceptance Criteria

Version: 0.4.0 | Tier: L | **36 acceptance-criterion leaves** — declared over the Tier L ceiling of 25, per `spec.md` §D Budget exception

> **Count declaration.** A "leaf" is one independently-asserted binary criterion: `AC-FM-020a` and `AC-FM-020b` are two leaves, not one. Counting parent IDs alone yields 26, which is the number v0.2.0 reported (as "25") and is not what a reviewer must actually evaluate. The honest leaf count is 36. It is declared, not adjusted — see `spec.md` §D for why the excess is verification depth on one security-critical chain rather than scope creep, and why the SPEC was not split to fit.

## §A Conventions

Every criterion is binary-testable and written Given-When-Then. A criterion whose judgement command is a grep names the exact pattern; a criterion whose judgement is a Go test names the assertion.

### A.1 Command-authoring rules

- A judgement command placed inside a markdown table cell must not rely on an unescaped `|`; `grep -E` treats `\|` as a literal pipe while BRE treats it as alternation. Where a pattern needs alternation, the command lives in a fenced block, not a table cell.
- A negative (absence) criterion must be paired with a positive control that demonstrates the pattern would match if the condition were violated. An absence grep with no positive control is a vacuous GREEN.
- Coverage and test-result criteria cite the command and its observed output; a remembered figure is not a baseline.
- **A criterion asserting "the documented behavior is X" is not binary-testable by a machine.** Every doctrine criterion below therefore names the file, the literal pattern, and the expected match count. Where a criterion has multiple binary sub-criteria within one logical concern, they carry lowercase sub-IDs (`AC-FM-020a`, `AC-FM-020b`, …); a sub-ID PASSes independently and the parent AC PASSes only when every sub-ID PASSes. **Each sub-ID is one leaf for counting purposes** — the declared total at the head of this file is the leaf count, not the parent-ID count, because the leaf count is what a reviewer must actually evaluate and what the Definition of Done requires evidence for.
- **Every grep pattern below is run from the repository root** and is written for `grep -c` (count) or `grep -n` (line-numbered) unless the criterion states otherwise. Patterns containing alternation appear only inside fenced blocks.

### A.2 Doctrine file set (the enumerated scope for absence criteria)

The complete set of workflow files this SPEC edits. Absence criteria are bounded to this set — an unbounded "anywhere in the chain" assertion is unfalsifiable.

```
.claude/skills/moai/workflows/moai.md
.claude/skills/moai/workflows/run.md
.claude/skills/moai/workflows/review.md
.claude/skills/moai/workflows/factory.md                      (new file)
.claude/skills/moai/workflows/sync/quality-gates-quality.md
.claude/rules/moai/workflow/goal-directive.md                 (rule file — added v0.3.0 per REQ-FM-028)
```

Shell variables used by several criteria below. `MIRRORS` names the template counterpart of each entry, which is the scope AC-FM-025 is bounded to.

**These MUST be arrays, and every expansion MUST be quoted.** The project's shell is `zsh`, which does not word-split an unquoted scalar the way `bash` does: a `DOCS="a b c"` scalar plus `for f in $DOCS` runs the loop **once** with `$f` bound to the whole six-path string, printing a single `0`, and the derived `MIRRORS` scalar becomes one non-existent path so `grep -n … $MIRRORS` errors to stderr and produces **no stdout** — which is AC-FM-025's literal PASS condition. The mirror-neutrality check would pass vacuously on the default shell. The array form below runs correctly under both `zsh` and `bash`.

```bash
DOCS=(
  .claude/skills/moai/workflows/moai.md
  .claude/skills/moai/workflows/run.md
  .claude/skills/moai/workflows/review.md
  .claude/skills/moai/workflows/factory.md
  .claude/skills/moai/workflows/sync/quality-gates-quality.md
  .claude/rules/moai/workflow/goal-directive.md
)

MIRRORS=()
for f in "${DOCS[@]}"; do MIRRORS+=("internal/template/templates/$f"); done
```

**Existence guard (mandatory before any `$MIRRORS` grep).** A missing mirror file must be a FAIL, not an empty result — an absent path is exactly the condition an absence grep cannot distinguish from a clean one:

```bash
for m in "${MIRRORS[@]}"; do
  [ -f "$m" ] || { printf 'FAIL missing mirror: %s\n' "$m"; }
done
```

Post-change every entry must exist; `factory.md` and its mirror are created by M1, so pre-change the guard legitimately reports the two `factory.md` paths absent and every other entry present. A post-change run reporting any `FAIL missing mirror` line is an AC-FM-025 FAIL regardless of what the neutrality grep returns.

## §B Acceptance criteria

### B.1 Entry surface

**AC-FM-001** — Given `moai cc --factory`, When `runCC` executes with the `unifiedLaunchFunc` seam installed, Then the seam receives an argument slice containing neither `--factory` nor `-f`, and the factory-enabled flag observed by the test is `true`.

**AC-FM-002** — Given `moai cc -f SPEC-PLACEHOLDER`, When `parseFactoryFlag` runs, Then it returns `spec == "SPEC-PLACEHOLDER"`, `enabled == true`, and a rest slice with both tokens removed. (`SPEC-PLACEHOLDER` is deliberately not SPEC-ID-shaped so a cross-SPEC scan does not report a missing directory.)

**AC-FM-003** — Given `moai cc -- --factory`, When `parseFactoryFlag` runs, Then `enabled == false` and the rest slice retains `--` and `--factory` verbatim, proving the `--` pass-through boundary is respected exactly as `stripSpawnFlag` respects it. Traces to REQ-FM-002 (the `--` discipline folded in at v0.2.0).

**AC-FM-004** — Given `moai cg --factory`, When `runCG` executes, Then it returns a non-nil error whose message contains the literal `FACTORY_MODE_UNSUPPORTED_BACKEND`, and the `unifiedLaunchFunc` seam is never invoked.

**AC-FM-005** — Given `moai glm --factory`, When `runGLM` executes with the seam installed, Then the seam is invoked with mode `glm` and a factory-enabled session, confirming parity with `moai cc`.

**AC-FM-006** — Given the current tree, When `grep -rn '"-f"' internal/cli/cc.go internal/cli/glm.go internal/cli/cg.go` is run before the change, Then it produces no match — establishing the baseline that `-f` was previously unbound on these three commands. Positive control: after the change, the same grep against `internal/cli/factory.go` does match.

### B.2 Chain contract and verify gate

**AC-FM-007** — Given the edited `moai.md`, When the run→sync chaining policy section is read, Then both of the following hold:

```bash
grep -c 'factory` contract .*extends.*`full-pipeline' .claude/skills/moai/workflows/moai.md   # pre-change 0, post-change >= 1
grep -c 'gate-sync-1' .claude/skills/moai/workflows/moai.md                                    # pre-change 1 (measured), post-change >= 2
```

Judgement: the first count rises from 0 to at least 1 (the extension relationship is stated, not merely implied). The second count rises from its **measured pre-change baseline of 1** to at least 2 — the pre-existing `gate-sync-1` mention is not the factory clause, so a bare `>= 1` assertion would have passed before any edit was made. A delta of at least 1 proves the factory clause itself names the inherited gate. Record the pre-change count in the §C pre-flight before M1 edits the file; if the measured baseline differs from 1 at implementation time, the post-change threshold is `baseline + 1`, not the literal 2.

**AC-FM-008** — Given the edited `run.md`, When the verify gate section is read, Then the exact invocation string appears verbatim and is positioned at run-phase exit:

```bash
grep -c -- '/moai review --security --deep --repo' .claude/skills/moai/workflows/run.md   # >= 1
grep -c 'exit gate of run-phase' .claude/skills/moai/workflows/run.md                     # >= 1
```

Both counts at least 1. Positive control: the pre-change `run.md` returns 0 for both.

**AC-FM-009** — Given the edited `run.md`, When the CRITICAL/HIGH branch is read, Then the re-entry rule and the non-proceed rule are both stated as literals:

```bash
grep -c 're-enter run-phase scoped to the changed surface' .claude/skills/moai/workflows/run.md   # >= 1
grep -c 'shall not proceed to sync' .claude/skills/moai/workflows/run.md                          # >= 1
```

Positive control: a variant of the section that only says "re-enter run" without the non-proceed clause fails the second grep — the two clauses are independently required because the audit found the first without the second is compatible with proceeding anyway.

**AC-FM-010** — Given the edited `run.md`, When the re-entry ceiling is read, Then:

```bash
grep -c 'at most two verify re-entries' .claude/skills/moai/workflows/run.md    # >= 1
grep -c 'Baseline-attribution' .claude/skills/moai/workflows/run.md            # >= 1
```

The second pattern is the discriminating token of the 5-section verdict — a halt clause that omits the verdict format fails it.

**AC-FM-011** — Given the edited `run.md`, When the MEDIUM/LOW branch is read, Then:

```bash
grep -c 'inherited sync-phase evidence' .claude/skills/moai/workflows/run.md        # >= 1
grep -c 'readable result' .claude/skills/moai/workflows/run.md                       # >= 1
```

The second pattern proves the branch is guarded on readability, so a no-result verify cannot fall into it (the D3 defect).

**AC-FM-012** — Given the enumerated doctrine file set of §A.2, When the per-file gate-token counts are compared against their recorded pre-change baselines, Then each file's post-change count equals `baseline + N` with `N` the exact number of gate tokens the factory clauses introduce in that file, enumerated in advance.

This criterion is the sole criterion for REQ-FM-012, the most safety-relevant requirement here, so it must be able to detect the thing it exists to detect. The v0.2.0 form could not: it emitted a `sort -u` list of context-free fragments and asked a human to classify them, and a fourth gate would have appeared as one more `HUMAN GATE` token among many pre-existing ones — indistinguishable by inspection. (`-n` in the output also made `sort -u` dedupe nothing, so the list was not even shortened.) A count delta detects an unplanned gate because an unplanned gate is an unplanned token.

**The token set includes `AskUserQuestion` (v0.4.0).** The v0.3.0 set — `HUMAN GATE`, `gate-sync-1`, `gate-sync-2`, `Implementation Kickoff Approval` — was gate-shaped only by accident: a fifth gate introduced as plain prose ("the orchestrator shall ask the operator whether to continue"), or as an `AskUserQuestion` round carrying no gate label, changes no count and passes. REQ-FM-012's "No fifth gate shall be introduced" was therefore only partly testable by its sole criterion. Every human gate in this SPEC is an orchestrator-issued `AskUserQuestion` round (REQ-FM-023, `design.md` §6), so the token is the mechanical signature a label-free gate cannot avoid; adding it closes the widest remaining hole. A gate expressed in prose with no `AskUserQuestion` round behind it remains outside mechanical reach — that residual is stated, not claimed closed.

**Pre-flight step (§C item 5), run before M1 edits any file:**

```bash
for f in "${DOCS[@]}"; do
  if [ ! -f "$f" ]; then printf '%s ABSENT\n' "$f"; continue; fi
  printf '%s ' "$f"
  grep -o 'HUMAN GATE\|gate-sync-1\|gate-sync-2\|Implementation Kickoff Approval\|AskUserQuestion' "$f" | wc -l
done
```

Baselines **re-measured at v0.4.0 against the widened pattern** (command above, run in this worktree; re-measure at implementation time — these are the reference, not a substitute for the pre-flight run). `N` is the gate tokens the factory clauses introduce; `A` is the `AskUserQuestion` tokens they introduce; expected = baseline + `N` + `A`:

| File | Baseline (5-token) | Planned `N` | Planned `A` | Expected post-change |
|---|---:|---:|---:|---:|
| `moai.md` | 23 | 2 (`gate-sync-1` + `gate-sync-2`, named inside the factory clause) | 0 | 25 |
| `run.md` | 33 | 1 (`Implementation Kickoff Approval`, in the verify-gate ordering sentence) | 0 | 34 |
| `review.md` | 3 | 0 (the REQ-FM-019 correction adds no gate token) | 0 | 3 |
| `factory.md` | 0 (file absent) | 4 (`Implementation Kickoff Approval` ×1 in the arming rule, `gate-sync-1` ×1 and `gate-sync-2` ×1 in the inherited-gate list, `HUMAN GATE` ×1 labelling the verify decision) | 1 (the verify decision stated as an orchestrator-issued `AskUserQuestion` round) | 5 |
| `sync/quality-gates-quality.md` | 2 | 0 (the dedup gate is a suppression condition, not a human gate) | 0 | 2 |
| `goal-directive.md` | 18 | 0 (the REQ-FM-028 amendment adds no gate token) | 0 | 18 |

For reference, the same files under the v0.3.0 four-token pattern measure 12 / 23 / 0 / absent / 0 / 12 — re-measured at v0.4.0 and unchanged, so the widened figures differ only by the `AskUserQuestion` occurrences. Where the authored text at implementation time carries a different number of `AskUserQuestion` mentions than the planned `A`, re-derive `A` in the pre-flight and record it before editing; an unrecorded `A` is the same evidence-destroying mistake as an unmeasured baseline.

Judgement, all three required: (a) every file's post-change count equals its recorded baseline plus its planned `N` + `A`; (b) every token contributing to a non-zero `N` or `A` is one of the four gates REQ-FM-012 enumerates — Implementation Kickoff Approval, the verify CRITICAL/HIGH decision, `gate-sync-1`, `gate-sync-2` — verified by reading the `grep -n` context of the new lines only, a bounded set of at most 8 lines rather than ~47 fragments; (c) no file's count exceeds `baseline + N + A`.

**Escape valve on (c).** A per-file excess is a FAIL **unless** the `grep -n` context of every excess line shows it referencing an already-enumerated gate — for example a second, benign mention of `Implementation Kickoff Approval` in `run.md`'s verify-gate ordering prose, or an `AskUserQuestion` token in a sentence describing one of the four gates already counted. The v0.3.0 form hard-FAILed on any excess "whether or not the excess token names a gate a human would recognize as new", which meant a single harmless cross-reference failed the sole criterion for the most safety-relevant requirement here. The valve is narrow by construction: it admits only excess lines that name an **already-enumerated** gate, so an excess naming any new gate — or any excess whose context is ambiguous — remains a FAIL. The context read is bounded to the excess lines themselves.

Positive control: the pre-flight baselines are non-zero for five of the six files, proving the pattern fires and the delta is measured against a real, non-empty starting point.

**AC-FM-013** — Given the edited `run.md` and `factory.md`, When the DEGRADED handling is read, Then:

```bash
grep -c 'DEGRADED' .claude/skills/moai/workflows/run.md                     # >= 1
grep -c 'verify_rung' .claude/skills/moai/workflows/factory.md              # >= 1
```

The `verify_rung` token proves the rung is recorded in the state record rather than only mentioned in prose — the field the AC-FM-020c exclusion reads.

### B.3 Dedup contract (R1)

**AC-FM-014** — Given a fixture directory in `t.TempDir()` containing a valid `revision.json` whose `scanned_commit` equals the supplied head SHA, whose `scope` is `repo`, a `findings.jsonl` whose every line parses as JSON, and a clean tree, When `Matches` is called, Then it returns `true`. This is the sole happy path.

**AC-FM-015** — Given a fixture directory containing no `revision.json`, When `LoadRevision` and `Matches` are called, Then `Matches` returns `false`. A skip is impossible on absence.

**AC-FM-016** — Given a `revision.json` containing malformed JSON, When `LoadRevision` is called, Then it returns an error and `Matches` returns `false`.

**AC-FM-017** — Given a valid `revision.json` whose `scanned_commit` differs from the supplied head SHA, When `Matches` is called, Then it returns `false`.

**AC-FM-018** — Given a valid `revision.json` whose `scope` is `branch` (not `repo`), When `Matches` is called with an otherwise-matching head SHA, Then it returns `false`.

**AC-FM-019a** — Given a valid `revision.json` with `working_tree_included: false` and a matching `scanned_commit`, When `Matches` is called with `treeDirty == true`, Then it returns `false`; and when called with `treeDirty == false`, Then it returns `true`. Both halves must be asserted in the same test to prove the dirty-tree clause is the discriminating factor.

**AC-FM-019b** — Given a fixture directory whose `revision.json` is valid and matching but whose `findings.jsonl` is absent, When `Matches` is called, Then it returns `false`; and given the same fixture with a zero-line `findings.jsonl` present, Then it returns `true`. Both halves in one test — this is the D3 completeness precondition, and the zero-line half proves a genuinely clean scan is not mistaken for an aborted one.

**AC-FM-020** — The sync Phase 8 dedup gate doctrine. All four sub-criteria must PASS.

- **AC-FM-020a (disclosure)** — Given the edited `sync/quality-gates-quality.md`, When the Phase 8 dedup section is read, Then `grep -c 'scanned_commit' .claude/skills/moai/workflows/sync/quality-gates-quality.md` is at least 1 and the same file contains the literal `inherited from the factory verify stage`, so a skip is distinguishable from a clean scan in the sync report. Positive control: the pre-change file returns 0 for both.

- **AC-FM-020b (manifest audit exempt)** — Given the edited file, When the skip scope is read, Then it names Step 0.55.1 as the sole skip target and states the manifest audit is unaffected:

```bash
Q=.claude/skills/moai/workflows/sync/quality-gates-quality.md
grep -c 'Step 0.55.1' "$Q"                    # >= 1
grep -c 'dependency manifest audit' "$Q"      # >= 1  (subject named)
grep -c 'run unconditionally\|runs unconditionally' "$Q"   # >= 1  (predicate stated, either tense)
```

  The v0.2.0 form used the single pattern `dependency manifest audit .*runs unconditionally`, which requires a literal space after `audit` and the present-tense `runs`. REQ-FM-014's own prose writes `**dependency manifest audit**,` — the closing bold marker occupies the position the pattern expects a space in — and `It shall run unconditionally`. An implementer mirroring the requirement's phrasing verbatim would have failed the grep. The two patterns are counted separately so neither the bold markup nor the tense can break the judgement; the subject and the predicate are each independently asserted.

  Positive control (falsification): a variant of the edit that says only "sync Phase 8 shall be skipped" without naming Step 0.55.1 fails the first grep. The pre-existing always-on sentence (`regardless of whether manifest files changed in this SPEC`) must still be present and unmodified — verified by `grep -c 'regardless of whether manifest files changed' "$Q"` returning its pre-change count.

- **AC-FM-020c (rung allow-list — suppression requires a recorded PRIMARY or FALLBACK)** — Given the edited file, When the suppression condition is read, Then it is stated as an allow-list over `verify_rung`, not as a `!= DEGRADED` deny-list:

```bash
Q=.claude/skills/moai/workflows/sync/quality-gates-quality.md
grep -c 'PRIMARY' "$Q"    # >= 1
grep -c 'FALLBACK' "$Q"   # >= 1
grep -c 'DEGRADED' "$Q"   # >= 1
```

  and the matching prose states that suppression requires `verify_rung` to be **recorded and equal to** `PRIMARY` or `FALLBACK`, with every other value — including absent and empty — yielding no suppression. A file that names only `DEGRADED` as an exclusion is a FAIL even though the third grep passes: a deny-list phrasing is the defect this sub-criterion exists to catch.

  Go table test over the composed suppression decision, all three cases asserted in one test so the discriminating field is provable:

  | `verify_rung` on the record | predicate result | expected suppression |
  |---|---|---|
  | `"PRIMARY"` | TRUE | `true` (positive control) |
  | `"DEGRADED"` | TRUE | `false` |
  | `""` (field never written) | TRUE | `false` |

  The third row is the v0.2.0 fail-open case: `""` is not `DEGRADED`, so the old deny-list permitted suppression on a record whose rung the orchestrator never wrote — reachable because the state record is best-effort and `deepscan_dir` / `verify_rung` / `verify_reentries` are written independently (REQ-FM-024). Asserting rows 2 and 3 without row 1 would be vacuous (a function returning `false` unconditionally would pass); asserting row 1 without row 3 is the v0.2.0 gap.

- **AC-FM-020d (predicate is consumed, not dead code)** — Given the edited file, When the dedup gate procedure is read, Then it invokes the predicate with both derived inputs named:

```bash
grep -c 'git rev-parse HEAD' .claude/skills/moai/workflows/sync/quality-gates-quality.md     # >= 1
grep -c 'git status --porcelain' .claude/skills/moai/workflows/sync/quality-gates-quality.md # >= 1
grep -c 'revision-match predicate' .claude/skills/moai/workflows/sync/quality-gates-quality.md # >= 1
```

  All three at least 1. This is the D4 fix: without it, every other AC in §B.3 passes against a pure function nothing calls. Positive control: the pre-change file returns 0 for all three.

### B.4 Scope contradiction (R2)

**AC-FM-021** — Given the edited `review.md`, When the `--lean` scope subsection is read, Then it no longer asserts that `--repo` is honored only in `--lean` mode, and instead states that `--repo` is honored in both `--lean` and `--deep`. Positive control — the pre-change file matches the **verbatim** live text (note: no backticks around `--lean`), and the post-change file does not:

```bash
grep -c 'honored only in --lean mode' .claude/skills/moai/workflows/review.md
# pre-change: 1     post-change: 0
grep -c 'honored in both --lean and --deep' .claude/skills/moai/workflows/review.md
# pre-change: 0     post-change: >= 1
```

### B.5 Autonomy engine

**AC-FM-022** — The block-cap inject and the no-new-runtime constraint. All three sub-criteria must PASS.

- **AC-FM-022a (unconditional factory inject)** — Given a Go test with `t.Setenv(config.EnvMoaiFactory, "1")` (non-parallel test, per the OTEL/env isolation rule), When `injectStopHookBlockCapForGoal(ctx, base, tmpProjectRoot, "")` is called against a `t.TempDir()` project root with **no armed goal present**, Then the returned slice contains `CLAUDE_CODE_STOP_HOOK_BLOCK_CAP=200`. Positive control: the identical call with `MOAI_FACTORY` unset returns `base` unchanged, proving the factory branch — not the pre-existing goal branch — supplies the entry. This asserts against the concrete seam the existing `launcher_blockcap_infinite_test.go` already uses.

- **AC-FM-022b (no new runtime)** — Given the completed change set, When the repository is scanned for new autonomy runtime surfaces, Then no new Stop hook, evaluator, or hook script was added:

```bash
git diff --name-only --diff-filter=A origin/main...HEAD -- .claude/hooks/ internal/hook/
# expected: no output
grep -c 'stop-goal' .claude/skills/moai/workflows/factory.md   # >= 1
```

  Positive control: the second grep proves `factory.md` names the **existing** evaluator, so the empty first result means "reuses the existing runtime", not "never mentions a runtime".

- **AC-FM-022c (arm only after Kickoff, with the settled bounds)** — Given the edited `factory.md`, When the arming rule is read, Then all of the following appear:

```bash
grep -c 'Implementation Kickoff Approval' .claude/skills/moai/workflows/factory.md   # >= 1
grep -c -- '--max-turns 0' .claude/skills/moai/workflows/factory.md                  # >= 1
grep -c -- '--max-duration 14400' .claude/skills/moai/workflows/factory.md           # >= 1
```

  and the arming sentence orders them as arm-after-approval. Negative control: `grep -c 'stop after' .claude/skills/moai/workflows/factory.md` returns 0 — a prose turn clause is not parsed and must not be authored.

**AC-FM-023** — Factory signal propagation. All four sub-criteria must PASS.

- **AC-FM-023a (state record)** — Given `moai cc --factory SPEC-PLACEHOLDER`, When the launcher runs against a `t.TempDir()` project root, Then a record exists at `.moai/state/factory/<session>.json` carrying `session_id`, `spec_id`, `backend`, `entered_at`, and `verify_rung`; and when the state directory is made unwritable, Then the launch still succeeds (fail-open).

- **AC-FM-023b (env constants declared)** — Given the completed change set, When `internal/config/envkeys.go` is read, Then it declares both constants:

```bash
grep -c 'EnvMoaiFactory = "MOAI_FACTORY"' internal/config/envkeys.go       # == 1
grep -c 'EnvMoaiFactorySpec = "MOAI_FACTORY_SPEC"' internal/config/envkeys.go # == 1
```

  Positive control: the pre-change file returns 0 for both.

- **AC-FM-023c (env reaches the child environment)** — Given a Go test that sets `MOAI_FACTORY=1` and `MOAI_FACTORY_SPEC=SPEC-PLACEHOLDER` in the process environment, When the launch environment is built through the same `os.Environ()`-derived path used at `internal/cli/launcher.go:783` / `:786`, Then the resulting slice contains both `MOAI_FACTORY=1` and `MOAI_FACTORY_SPEC=SPEC-PLACEHOLDER`. This is the load-bearing link asserted by REQ-FM-023: if the variables do not survive into the child environment, the cap raise of AC-FM-022a is unreachable in production even though its unit test passes.

- **AC-FM-023d (the process-environment mutation is restored)** — Given a non-parallel Go test that records `before, beforeSet := os.LookupEnv(config.EnvMoaiFactory)` (and the same pair for `config.EnvMoaiFactorySpec`), When `runCC` is invoked with `--factory SPEC-PLACEHOLDER` against a `t.TempDir()` project root with the `unifiedLaunchFunc` seam installed, Then after `runCC` returns, `os.LookupEnv` for each variable yields exactly the recorded `(before, beforeSet)` pair — an initially-unset variable is unset again, not set to `""`. Both the success path and the error path (seam returns a non-nil error) must be asserted, because a `defer`-based restore that only runs on success is the same leak with a narrower trigger.

  This is the criterion that makes AC-FM-022a's stated negative control meaningful. Without it, an unrestored `os.Setenv` inside `runCC` leaves `MOAI_FACTORY=1` set for the remainder of the `internal/cli` test binary, so "the identical call with `MOAI_FACTORY` unset returns `base` unchanged" passes or fails by test-execution order — the ordering-dependent flake class the project's test-isolation rules exist to prevent. Positive control (falsification): remove the `defer` restore from the implementation and this criterion FAILs while every other AC in §B.5 still passes, demonstrating it is the only criterion sensitive to the leak.

**AC-FM-024** — The verify-gate fail-closed path. Both sub-criteria must PASS.

- **AC-FM-024a (no-result HALTs)** — Given the edited `run.md` and `factory.md`, When the no-result branch is read, Then it halts rather than proceeding:

```bash
grep -c 'no readable result' .claude/skills/moai/workflows/run.md   # >= 1
grep -c 'HALT' .claude/skills/moai/workflows/run.md                 # >= 1
```

  and the matching prose states the attempt does not count against the two-re-entry ceiling. Negative control (falsification of the D3 defect): `grep -c 'no confirmed findings' .claude/skills/moai/workflows/run.md` returns matches ONLY on lines that also carry the word `readable` — an unguarded "no confirmed findings → proceed to sync" line is a FAIL.

- **AC-FM-024b (3-case severity partition × orthogonal rung attribute, with stated precedence)** — Given the edited `run.md`, When the verify-outcome structure is read, Then it presents a **three-case severity partition** and a **separate rung attribute**, and states the precedence between them.

  The v0.2.0 form greped for the literal word `disjoint` and eyeballed four labels — a check a *false* disjointness claim satisfies perfectly, which is exactly what v0.2.0 asserted. The word is no longer evidence and is no longer the judgement. Three independent assertions replace it:

```bash
R=.claude/skills/moai/workflows/run.md
# (1) the severity axis names exactly three cases, and readability separates the third
grep -c 'no readable result' "$R"       # >= 1
grep -c 'readable result' "$R"          # >= 2  (the S2 guard and the S3 case)
# (2) the rung axis is named as an attribute, not as a fourth peer outcome
grep -c 'orthogonal' "$R"               # >= 1
# (3) the precedence sentence exists and binds both directions
grep -c 'governs routing' "$R"          # >= 1
grep -c 'governs suppression' "$R"      # >= 1
```

  Judgement, all required: (a) the three severity cases S1 / S2 / S3 are enumerated and each is stated to be mutually exclusive with the other two on the readability-plus-severity axis; (b) the rung (`PRIMARY` / `FALLBACK` / `DEGRADED`) is described as an attribute *of* an S1 or S2 result, not as a peer outcome alongside them, and S3 is stated to carry no rung; (c) the precedence sentence states both halves — the severity case governs routing and the rung never changes it, the rung governs suppression and the severity case never relaxes it. **A file that lists four peer outcomes and calls them disjoint is a FAIL**, regardless of the `orthogonal` grep, because that is the v0.2.0 defect verbatim.

  Falsification control: apply the three greps to the v0.2.0 draft text — patterns (2) and (3) return 0 while pattern (1) returns matches, so the criterion distinguishes the corrected structure from the defective one rather than passing on both.

### B.6 Distribution and quality

**AC-FM-025** — Given the completed change set, When `go test ./...`, `golangci-lint run`, `go test -cover ./internal/cli/... ./internal/factory/...`, `make build`, and the template-neutrality guard are run, Then all exit 0, coverage is at least 90% for `internal/cli` and 85% for `internal/factory`, every `.claude/` file touched has a mirrored counterpart under `internal/template/templates/.claude/`, and **no mirrored file this SPEC touches** contains a SPEC identifier, requirement token, acceptance token, internal date, commit SHA, or internal Go package path. The package-path clause is judged by:

```bash
for m in "${MIRRORS[@]}"; do
  [ -f "$m" ] || { printf 'FAIL missing mirror: %s\n' "$m"; }
done
grep -n 'internal/factory\|internal/cli' "${MIRRORS[@]}"
# expected: no `FAIL missing mirror` line, and no grep output
# (measured pre-change baseline over the five mirrors that exist: 0 matches, grep exit 1;
#  the sixth, factory.md, is created by M1)
```

The existence guard is not decoration. Without it the grep's "no output" PASS condition is satisfied by a missing file just as well as by a clean one — and under the pre-v0.4.0 scalar `MIRRORS` every path was missing, so the criterion passed vacuously on every run.

**Scope bound (v0.3.0).** The v0.2.0 form ran this grep across the entire `internal/template/templates/.claude/` tree, where **8 pre-existing matches** live in files this SPEC never touches — `plan-auditor.md:210`, `askuser-protocol-reference.md:149`, `agent-hooks.md:69`, `worktree-integration.md:307,369,380`, and `worktree-state-guard.md:2,15`. The criterion therefore failed on every run regardless of this SPEC's cleanliness, and the obvious "fix" — scrubbing five unrelated shipped rule files — is a scope-discipline violation with its own mirror-parity blast radius. REQ-FM-025 was already correctly scoped to the files this SPEC mirrors; only the criterion over-reached. Those 8 references are recorded as out of scope in `spec.md` §C.

Positive control: the same pattern applied to the corresponding **live** files (`grep -n 'internal/factory\|internal/cli' $DOCS`) DOES match after M4 and M5 land — the plan and the live doctrine legitimately name `internal/factory` — proving the pattern is capable of firing and the empty `$MIRRORS` result is a real absence rather than a mis-typed pattern.

**AC-FM-026** — Given the edited `.claude/rules/moai/workflow/goal-directive.md` and its template mirror, When the block-cap trigger sentence is read, Then it names both trigger conditions rather than only the launch-time armed-goal condition:

```bash
G=.claude/rules/moai/workflow/goal-directive.md
M=internal/template/templates/$G
# the launch-time-only framing no longer stands alone
grep -c 'at launch time' "$G"     # pre-change 1 (measured), post-change >= 1 — the clause survives, qualified
grep -c 'Factory Mode' "$G"       # pre-change 0 (measured), post-change >= 1 — the second trigger is named
grep -c 'Factory Mode' "$M"       # post-change >= 1 — the mirror carries the identical amendment
cmp "$G" "$M" && echo BYTE-IDENTICAL   # the two files are byte-identical pre-change (verified); they must remain so
```

Judgement, all required: (a) the amended sentence states that the launchers inject the raised cap under **either** an armed `--max-turns 0` goal at launch time **or** Factory Mode, so a reader who consults this file to explain a session that unexpectedly ran 200 blocks finds the real answer; (b) the mirror carries the identical amendment and `cmp` still reports the two files byte-identical; (c) the amendment introduces no SPEC identifier, requirement token, acceptance token, internal date, commit SHA, or `internal/...` package path — verified by `grep -n 'SPEC-FACTORY\|REQ-FM-\|AC-FM-\|internal/factory\|internal/cli' "$M"` returning no output. Clause (c) binds only the text this SPEC adds; the pre-existing `SPEC-INFINITE-GOAL-001` heading already present in both trees is untouched and out of scope.

Positive control: `grep -c 'Factory Mode' "$G"` returns 0 pre-change (measured), so the post-change match is a real addition rather than a pattern that was already satisfied. Falsification: amending only the live file leaves `cmp` reporting a difference and the third grep returning 0 — the criterion fails on a half-applied mirror, which is the failure mode a byte-identical pair invites.

## §C Traceability

Every requirement maps to at least one criterion. A requirement with no criterion is untestable; a criterion with no requirement is scope creep.

| Requirement | Criteria |
|---|---|
| REQ-FM-001 | AC-FM-001 |
| REQ-FM-002 | AC-FM-002, AC-FM-003 |
| REQ-FM-004 | AC-FM-004 |
| REQ-FM-005 | AC-FM-002, AC-FM-023a |
| REQ-FM-006 | AC-FM-006 |
| REQ-FM-007 | AC-FM-007 |
| REQ-FM-008 | AC-FM-008 |
| REQ-FM-009 | AC-FM-009 |
| REQ-FM-010 | AC-FM-010 |
| REQ-FM-011 | AC-FM-011, AC-FM-024b |
| REQ-FM-012 | AC-FM-012 |
| REQ-FM-013 | AC-FM-013, AC-FM-024b |
| REQ-FM-014 | AC-FM-020a, AC-FM-020b, AC-FM-020c |
| REQ-FM-015 | AC-FM-014, AC-FM-017, AC-FM-018, AC-FM-019a, AC-FM-019b |
| REQ-FM-016 | AC-FM-015, AC-FM-016, AC-FM-019b |
| REQ-FM-017 | AC-FM-020d |
| REQ-FM-018 | AC-FM-020a |
| REQ-FM-019 | AC-FM-021, AC-FM-008 |
| REQ-FM-021 | AC-FM-022b |
| REQ-FM-022 | AC-FM-022c |
| REQ-FM-023 | AC-FM-022a, AC-FM-023c, AC-FM-023d |
| REQ-FM-024 | AC-FM-023a, AC-FM-023b, AC-FM-013 |
| REQ-FM-025 | AC-FM-025 |
| REQ-FM-026 | AC-FM-024a, AC-FM-024b, AC-FM-019b |
| REQ-FM-027 | AC-FM-022c |
| REQ-FM-028 | AC-FM-026 |

Withdrawn requirements REQ-FM-003 and REQ-FM-020 carry no rows; their criteria (AC-FM-003, AC-FM-021) trace to the requirements that absorbed them (REQ-FM-002, REQ-FM-019).

## §D Definition of Done

- All **36 acceptance-criterion leaves** PASS with cited command output — every lowercase sub-ID counts as its own leaf and none may be reported by its parent ID alone.
- Zero unresolved clarification markers remain in `plan.md` and `research.md`, verified by the runtime-assembled sweep in `plan.md` §E. The token is assembled from fragments rather than spelled out, because a criterion that writes its own search pattern in prose matches itself and reports a self-trip as a finding.
- `moai spec lint` reports zero errors for this SPEC directory.
- No gate beyond the four REQ-FM-012 enumerates exists in the doctrine file set of §A.2, verified by the AC-FM-012 per-file baseline-delta rather than by inspection.
- No criterion in §B is phrased as "the documented behavior is …" without a named file and pattern.
