# Post-repair check — does the repair reach the symptom it was built for?

> SPEC-MCP-WORKTREE-ROOT-001 M5 — REQ-6 (AC-6).
> Run in the worktree `.claude/worktrees/t171` at `df4e3d573`.

## What had to be worked around first

The MCP server this session talks to was started **before** M1 landed. Its
declared schema carries no `project_root` on any tool, so calling it would have
measured the old binary and proved nothing about the repair. An MCP server is a
long-lived subprocess; it does not reload when the binary underneath it changes.

So the check drives the **rebuilt** `bin/moai mcp-server` directly over stdio,
launched with:

- `CLAUDE_PROJECT_DIR` = `/Users/goos/MoAI/moai-adk-go` (the primary checkout) —
  the exact condition that produces the defect;
- working directory = the primary checkout, so neither resolution branch can
  reach the worktree by accident;
- `project_root` = the worktree, passed as the tool input.

That is the real deployment shape, not a favourable one: both of the server's own
resolution paths point away from the tree under test, and only the parameter
points at it.

Probe driver: `.moai/cache/m5probe.py` (local-only, not committed — `.moai/cache/`
is excluded from the repository).

## Stage 1 — which tree does the server read?

**The two trees differ by exactly one SPEC.** Measured, not assumed:

```
$ /bin/ls -1 <primary>/.moai/specs | wc -l      → 629
$ /bin/ls -1 <worktree>/.moai/specs | wc -l     → 630
$ comm -13 primary-specs.txt wt-specs.txt       → SPEC-MCP-WORKTREE-ROOT-001
$ comm -23 primary-specs.txt wt-specs.txt       → (empty)
```

One SPEC exists only in the worktree, none only in the primary. That makes the
catalogue count a clean discriminator: a difference of exactly one, in exactly
one direction, is attributable to that SPEC and to nothing else.

**Schema, on the rebuilt server:**

| Tool | declares `project_root` |
|---|---|
| `spec_progress` | yes |
| `spec_audit` | yes |
| `spec_drift` | yes |
| `codex_audit` | yes |
| `audit_multi` | yes |

**Behaviour**, same server, same environment, three calls:

| Call | `total_specs` | `modern_era_clean` |
|---|---|---|
| `spec_audit` with no `project_root` | **627** | 349 |
| `spec_audit` with `project_root: <worktree>` | **628** | 350 |
| `spec_audit` with `project_root: <worktree>/no-such-tree` | — | rejected: `spec_audit: project_root "…/no-such-tree" does not exist`, `isError: true` |

The counts are the measurement; the SPEC's own ID appears in **neither** result,
and that is expected rather than a gap. `spec_audit` reports drift findings, and
a clean SPEC produces none — so the worktree-only SPEC is present in the
catalogue and silent in the output, which is exactly why the count is the
discriminator and a name search is not. An earlier draft of this table claimed a
"contains the SPEC" column; it asserted an observation that was never made, and
the sync audit was right to flag it (F-5).

The `+1` is the SPEC that exists only in the worktree, and the direction is the
one the repair predicts. The no-parameter call reads the primary checkout —
which is not a regression, it is REQ-2 holding: an unaware caller sees exactly
what it saw before.

The third row is the part that is easy to skip and matters most. A mistyped path
is refused by name. Had it fallen back to the default, the caller would have been
returned to auditing the primary checkout and told it succeeded — this SPEC's
defect, arriving through the mechanism built to fix it.

(627 audited vs 629 directories: two entries in each tree are not auditable SPEC
directories. The count is consistent across both trees, so it does not affect the
comparison.)

## Stage 2 — `audit_multi` against live backends

One call, real backends, no doubles: `target: baseBranch`, gates
`claude=required, codex=required, glm=advisory`, `project_root` = the worktree.
codex-cli 0.147.0 on PATH; the GLM key present at `~/.moai/.env.glm`. No backend
failed open (`fail_open_backends: []`), so all three verdicts are real.

### Which tree did each backend read?

| Backend | Tree it read | How that is known |
|---|---|---|
| claude | — | the anchor verdict was supplied by the probe, not produced by a review |
| **codex** | **the worktree** | its findings cite `…/.claude/worktrees/t171/internal/cli/…` **6 times and the primary checkout 0 times**, at line numbers that match this branch's edits |
| glm | **no tree at all** | GLM is an HTTP call to z.ai with no working directory to bind a root to; it never receives one, by construction |

codex's citations are the load-bearing evidence, and they are not a self-report:
it names files at line offsets that exist only on this branch. The parameter
carried the root through the fan-out and into the params map, and codex acted on
it.

**GLM's row is a finding, not a footnote.** The tool now accepts a root that one
of its two backends structurally cannot use. Nothing in the result says so, so a
caller reading a GLM verdict has no way to know which tree it is about — because
it is about no tree.

### What codex found — including two defects in this card's own work

codex reviewed the branch it was pointed at and returned three findings. Two
land on this card:

1. **`audit_multi` persists its convergence state to the wrong tree.**
   `defaultConvergenceStateDir` is a package-level value computed from
   `resolveProjectDir()` at load time, and `persistConvergenceResult` never
   consults `cfg.ProjectRoot`. A worktree audit writes its verdict into the
   primary checkout's `.moai/state/audit-multi/`.
   **Status: already surveyed, still deferred.** This is row 6 of
   `resolve-project-dir-verdicts.md`, recorded there as undecided. codex argues
   it is a live defect rather than an open question, which is a stronger claim
   than the table makes — and the table's own "what would settle it" note
   (decide whether the cache is keyed by session alone or by session+tree) is
   exactly the decision codex's recommendation presumes. Left for the follow-up,
   now with a second opinion attached.

2. **The shared parameter description promised the wrong fallback.** One
   description string was attached to all five tools, and it stated that an
   absent parameter resolves through `resolveProjectDir()`. That is true of
   four of them and false of `audit_multi`, whose absent case supplies no root
   at all. A caller reading `audit_multi`'s schema was told the wrong thing.
   **Status: FIXED in this milestone** — `projectRootDesc` and
   `projectRootPassthroughDesc` now differ in exactly the way the two resolvers
   differ, and `TestProjectRootDescription_MatchesEachToolsResolver` pins each
   tool to the right one. Checked against a pre-fix tree: pointing `audit_multi`
   back at the fallback description makes the test fail, so it observes
   something.

3. **A named root is not constrained to this repository.** Any directory
   containing `.moai` is accepted, so a caller can point the server at an
   unrelated project. This is REQ-3 as written, not a deviation from it — but
   the parameter does widen what an MCP caller can reach, and that widening was
   never stated as a consequence. Recorded for the follow-up; the suggested
   constraint (require a shared git common directory) is a scope change, not a
   bug fix.

Finding 2 is the one that matters for judging this check: **the post-repair run
found a real defect in the repair, in code the unit tests all passed on.** The
tests asserted behaviour; the description is a promise about behaviour, and
nothing was comparing the two until a reader that had not written the code read
the schema.

## Lane-9 symptoms — both recur, and neither is explained by this defect

REQ-6 asks whether the two lane-9 symptoms reappear. Both do. Per AC-6 that does
not fail this criterion; what would fail it is not looking.

**The self-contradicting verdict RECURS.** codex's `per_backend_verdicts` entry
carries `verdict: "pass"` while the summary it wraps opens:

> `Verdict: fail — merge blocked.`

The convergence engine then read that `pass` and produced `overall_verdict:
"pass"` — a merge-blocking review recorded as a passing one. The mis-read is in
codex verdict synthesis (the adversarial path's free-text verdict is not being
extracted), and it is entirely upstream of `project_root`: the review was of the
right tree and its content is sound. **This defect is independent of the one
this card repaired, and it silently inverts a required-gate verdict.** It is the
most serious thing this check surfaced.

**The hallucinated API RECURS, on GLM.** GLM returned `verdict: "fail"` citing "a
critical vulnerability … failing to resolve the project root path before
validating against the repository whitelist" and "relative path traversal
(e.g. `../other_repo`)". There is no repository whitelist in this code, and
`validateProjectRoot` calls `filepath.Abs` before every check — the resolution
it says is missing is the first thing the function does.

The attribution is clean, and it is the useful part: **GLM never receives a tree
at all**, so its hallucination cannot be caused by reading the wrong one. That
rules this card's defect out as the cause. The likely cause is visible in the
same run — GLM was given a focus string and no diff, so it had nothing to review
and produced plausible-sounding prose about code it never saw. A backend with no
input returning a confident `fail` is worse than one returning `inconclusive`.

Both recurrences point the same way: **the lane-9 symptoms had a second cause,
and the worktree root was not it.** The repair is still correct — codex demonstrably
read the right tree — it is just not the fix for those two symptoms.

## What this check does NOT establish

- **It does not exercise the MCP server this session is connected to.** That
  server predates the repair and is unchanged; the check drove a separate
  process built from this branch. Users get the repair when the binary is
  installed and their server restarts, which this check did not do and cannot
  verify.
- **It does not establish that codex read ONLY the worktree.** It establishes
  that everything codex cited is in the worktree. A read outside it that
  produced no finding would leave no trace here.
- **It does not measure the convergence state file's landing tree.** Finding 1
  is codex's reading of the code, and this run did not check where the file
  actually went — the probe passed no `session_id`, so persistence was a no-op.
  Verifying it needs a run with a session id and a look in both trees.
- **It is one run.** Both backends are non-deterministic; a second run may find
  different things. Nothing here is a claim about what codex or GLM will say
  next time — only about what they said, and which tree they said it about.

## Follow-ups this check produced

| # | Finding | Where it goes |
|---|---|---|
| 1 | codex adversarial verdict is synthesized as `pass` when its text says `fail` — a required gate silently inverted | new card; highest severity found here |
| 2 | GLM returns a confident `fail` reviewing nothing, and takes no tree — its verdict is unattributable to any code | new card |
| 3 | `audit_multi` convergence state persists to `resolveProjectDir()`, ignoring `cfg.ProjectRoot` | already row 6 of the verdict table; codex's second opinion attached |
| 4 | any `.moai`-bearing directory is an acceptable root — cross-project reach was never stated as a consequence | verdict table follow-up |
