# Sync-phase audit — SPEC-MCP-WORKTREE-ROOT-001

> Auditor: sync-auditor (independent). Tree: `.claude/worktrees/t171`, branch
> `WT-audit-worktree-blind`, HEAD **`34de07740`** (verified, not re-resolved).
> Profile: `.moai/config/evaluator-profiles/default.md` (harness
> `default_profile: "default"`; the SPEC frontmatter declares no
> `evaluator_profile`). Mode: flat weighted-percentage.
> Must-pass firewall: Functionality + Security.

## Overall verdict: **PASS** (weighted harmonic mean 87.5)

Both must-pass dimensions clear independently. No blocking finding. Seven
optional findings, all with named owners.

---

## Method note — what I did not take on trust

Three claims in this card are the kind that pass review by being read rather
than tested, so each was **mutation-tested**: the production line was removed,
the test re-run, and the file restored byte-identically. `git status --porcelain`
is empty at the end of this audit; the tree is as I found it.

The MCP version skew the dispatch flagged is real and I worked around it rather
than through it: every `mcp__moai__*` tool in this session is served by a process
started before M1, which declares no `project_root`, so its results describe the
PRIMARY checkout. **No MCP tool result is cited anywhere in this report.** All
evidence below is Bash/Read/Grep inside this worktree. That skew is itself scored
under Consistency.

---

## Dimension scores

| Dimension | Score | Verdict | Evidence (verbatim, abridged to the load-bearing line) |
|---|---|---|---|
| Functionality (40%) | 95/100 | **PASS** (must-pass) | `go test ./internal/cli/ -run 'ProjectRoot\|CodexAudit_\|AuditMulti_\|SpecAudit_\|SpecTools_\|CodexTools_\|RunMultiAudit' -count=1` → `ok github.com/modu-ai/moai-adk/internal/cli 0.848s`; three mutation checks each turned a green test red (§ Functionality) |
| Security (25%) | 82/100 | **PASS** (must-pass) | `grep -nE 'exec\.Command\|os\.Setenv\|Symlink\|EvalSymlinks\|\.\./\|os\.RemoveAll' internal/cli/mcp_project_root.go` → `none`; `golangci-lint run ./internal/cli/...` → `0 issues.` No Critical/High |
| Craft (20%) | 80/100 | PASS (scoped to the change) / FAIL on the profile's package-level coverage line | `go test ./internal/cli/ -count=1 -cover -timeout 900s` → `ok … 378.307s coverage: 78.5% of statements`; new-code per-function 92.3–100% (§ Craft) |
| Consistency (15%) | 90/100 | PASS | `diff` of all four changed docs against their template mirrors → three `rc=0`, one pre-existing unrelated divergence; `MOAI_TEMPLATE_LEAK_STRICT=1 go test ./internal/template/...` → `ok … 20.577s` |

Weighted harmonic mean = `1 / (0.40/95 + 0.25/82 + 0.20/80 + 0.15/90)` = **87.52**.

Firewall: Functionality 95 and Security 82 each clear their own threshold, so
neither forces a FAIL. Craft is not a must-pass dimension in this profile
(only "Security FAIL = Overall FAIL" is stated as overriding), and the coverage
shortfall it carries is package-scoped and pre-existing — see § Craft.

---

## Functionality — AC by AC

### AC-1a — the SPEC redirect, both directions

**Claim**: a `spec_audit` rooted at a worktree sees a SPEC that exists only
there; without the parameter it does not.

**Evidence** — unit, `internal/cli/mcp_project_root_test.go:83` and `:97`:

```
$ go test ./internal/cli/ -run 'SpecAudit_' -count=1
ok  	github.com/modu-ai/moai-adk/internal/cli	0.848s
```

**Evidence** — live, the M5 probe's own JSON, read directly rather than through
the report's summary of it (`.moai/cache/m5-multi-probe.json`):

```
spec_audit_without_project_root  total_specs = 627
spec_audit_with_project_root     total_specs = 628
spec_audit_bad_project_root      isError = true,
  "spec_audit: project_root \"…/t171/no-such-tree\" does not exist"
schema_declares_project_root = {audit_multi: true, codex_audit: true,
  spec_audit: true, spec_drift: true, spec_progress: true}
```

**Baseline-attribution**: both runs are the same rebuilt server process, same
environment, in the JSON committed at `.moai/cache/` (untracked but present on
disk at this HEAD). `modern_era_clean` moves 349 → 350 in the same pair, so the
`+1` is a modern-era SPEC, which the worktree-only one is.

**Gap** — see F-5: the report's Stage-1 table has a column *"Contains the
worktree-only SPEC"* reading `yes`. The string `SPEC-MCP-WORKTREE-ROOT-001` does
**not** appear anywhere in the with-parameter result — a clean SPEC produces no
drift finding, so it is never named. What the JSON supports is the count delta,
which the report's prose states correctly two lines later. The column overstates
what was observed.

**Verdict: MET.**

### AC-1b — the codex `cwd`, on **both** paths, asserted on the params map

This is the claim the dispatch asked me to attack directly, so it was
mutation-tested rather than read.

**Single-backend path.** `TestCodexAudit_ProjectRootLandsInTheParamsMap` swaps
`codexLookPath` and `codexReviewRPC` (`mcp_project_root_codex_test.go:70-77`) and
asserts `params["cwd"]`. It reads the map, not a double's argument.

**Fan-out path — does the assertion actually enter `performCodexAudit`?** Yes,
and structurally it cannot do otherwise: the test drives `handleAuditMulti` with
the **real** `backendCall`, and `grep -rn 'codexReviewRPC' internal/` shows the
only fan-out route to that seam is `mcp_convergence.go:422`, inside
`performCodexAudit`. Proven mechanically:

```
$ perl -0pi -e 's/\tif root := strings\.TrimSpace\(projectRoot\); root != "" \{\n\t\tparams\["cwd"\] = root\n\t\}\n//' internal/cli/mcp_convergence.go
$ go test ./internal/cli/ -run 'TestAuditMulti_ProjectRootLandsInTheCodexParamsMap' -count=1
--- FAIL: TestAuditMulti_ProjectRootLandsInTheCodexParamsMap (0.00s)
    mcp_project_root_codex_test.go:266: fan-out codex params carry no cwd key at all; params=map[model: prompt:… target:uncommittedChanges]
FAIL
```

Restored: `git diff --stat internal/cli/mcp_convergence.go` → empty.

**Live corroboration, from the raw probe rather than the report**: codex's
findings in `audit_multi_with_project_root` cite
`/Users/goos/MoAI/moai-adk-go/.claude/worktrees/t171/…` **6 times** and paths
outside the worktree **0 times** — I counted them in the JSON, and the count
matches the report's claim exactly.

**Verdict: MET.** The gap iter-2 identified (a `backendCall` double satisfying
the criterion without entering the function that builds the params) is closed,
and the closure is verified by mutation rather than by inspection.

### AC-2 — absent parameter, checked on each path separately

The two paths use **different resolvers** on purpose, and the dispatch was right
that a wrong one would be invisible. Checked separately:

- `codex_audit` → `resolveToolProjectRoot`, empty ⇒ `resolveProjectDir()`
  (`mcp_project_root.go:85`). Asserted by
  `TestCodexAudit_WithoutProjectRootKeepsTodaysCwd` against
  `resolveProjectDir()` itself, so it fails if the fallback is dropped.
- `audit_multi` → `resolveOptionalToolProjectRoot`, empty ⇒ `""`
  (`mcp_project_root.go:103`), and `performCodexAudit` writes the key only when
  the root is non-empty. Mutation-tested:

```
# resolveOptionalToolProjectRoot's empty branch changed to return resolveProjectDir()
$ go test ./internal/cli/ -run 'TestAuditMulti_WithoutProjectRootAddsNoCwd' -count=1
--- FAIL: TestAuditMulti_WithoutProjectRootAddsNoCwd (0.00s)
    mcp_project_root_codex_test.go:287: fan-out codex params gained a cwd key with no project_root supplied; today's behaviour was to pass none
FAIL
```

Restored; `git status --porcelain` empty.

The live probe's no-parameter call returning **627** — the primary's count — is
the same guarantee observed end to end.

**Verdict: MET on both paths, and the divergence between them is deliberate,
documented at `mcp_project_root.go:90-99`, and separately asserted.**

### AC-3 — rejection, not fallback

`TestResolveToolProjectRoot_RejectsAnUnusableRoot` covers all three shapes
(missing / file / no `.moai`), asserts the error names the path, and asserts the
returned root is empty. `TestSpecAudit_RejectedProjectRootDoesNotAuditTheDefault`
and `TestCodexAudit_RejectsAnUnusableProjectRoot` re-assert at the handler
boundary, the latter also checking `cap.count() == 0` — codex is never reached.
The live probe row above shows the same, from a real server.

**Verdict: MET.**

### AC-4 — the verdict table covers every grep-found call site

I re-derived the set rather than trusting the table's own total.

```
$ grep -rn 'resolveProjectDir()' internal/ | grep -v _test.go | wc -l
      25
```

25 matches decompose into: 6 doc-comment mentions (`mcp_project_root.go` ×5,
`mcp_codex.go:1169`), 1 definition (`session.go:242`), and **18 live call
sites**. All 18 map to a row:

| Call site (HEAD) | Row |
|---|---|
| `mcp_server.go:105` | 1 |
| `mcp_server.go:474`, `:491` (goal) | 2, 3 |
| `mcp_server.go:551`, `:581` (verify) | 4, 5 |
| `mcp_convergence.go:590` | 6 |
| `mcp_project_root.go:85` | 7 |
| `mcp_codex.go` (doc comment for the repair) | 8 |
| `goal.go:159`, `:188` | 9, 10 |
| `launcher_blockcap_infinite.go:139` | 11 |
| `session.go:224`, `:357` | 12, 13 |
| `todo.go:68` | 14 |
| `verify.go:72` | 15 |
| `plan.go:67` | 16 |
| `memory.go:165`, `:287` | 17, 18 |
| `migrate_profiles.go:378` | 19 |

The table's line numbers are `2c0efade0`'s, as it states. `git diff --stat
2c0efade0 34de07740` shows only `mcp_project_root.go` and `mcp_server.go` changed
since, which accounts for the +3 drift on the `mcp_server.go` rows and +23 on
row 7. No call site is missing and no row is orphaned.

Every deferred row carries a settling statement. I read all eleven: rows
2/3/9/10 delegate to § The goal rows, which states the measurement explicitly
(arm from a worktree, read `.moai/state/goal/` in both trees, record which the
evaluator reads); rows 4/5/15 name a single shared measurement; row 6 names the
keying decision; row 11 states it is not independently decidable and is bound to
the goal rows; row 12 names the hook's own resolution as the thing to read; row
16 names the missing item as a decision and states the known repair shape. **No
row reads "unknown" without saying what would settle it.**

**Verdict: MET.**

### AC-5 — the goal rows record a suspected shared cause

Present, and marked unverified in three places: the heading text ("**may share
this seam**"), an explicit "**This connection is SUSPECTED, not established.
Nothing in this card measured it.**", and a falsification clause ("If they agree,
… this note should be struck rather than carried forward"). The row-10 caveat —
that the guard compares against a project-scoped registry, so flipping the goal
rows to tree scope would break it silently — is a genuine addition beyond what
REQ-5 asked for.

**Verdict: MET.**

### AC-6 — post-repair check, including the recurrences

I read `.moai/cache/m5-multi-probe.json` and checked the report against it
rather than against its own summary.

**The self-contradicting verdict.** Confirmed verbatim in the raw JSON:

```
per_backend_verdicts[codex].verdict  = "pass"
per_backend_verdicts[codex].summary  = "Verdict: fail — merge blocked.\n\nFindings:\n\n1. High severity…"
overall_verdict                      = "pass"
fail_open_backends                   = []
```

The report's reasoning that this is upstream of `project_root` **holds, and I
verified the load-bearing half of it independently**: codex cited the worktree 6
times and anything else 0 times, so the review was of the right tree; the
mis-synthesis therefore cannot be caused by reading the wrong one. This is the
most serious thing the run surfaced and the report says so.

**The hallucinated API.** Confirmed verbatim: GLM returns `verdict: "fail"`
citing "failing to resolve the project root path before validating against the
repository whitelist" and "relative path traversal (e.g., `../other_repo`)".
Neither exists — there is no whitelist in this code, and `validateProjectRoot`
calls `filepath.Abs` on line 112, before every check. The report's attribution
(GLM receives no tree at all, so a wrong tree cannot be its cause) is sound and
matches `defaultBackendCaller`'s explicit drop at `mcp_convergence.go:381-384`.

**The report's honesty about its own limits is real, not decorative.** Its "What
this check does NOT establish" section correctly disclaims the four things it
cannot show, including that it never exercised the session's own MCP server and
never measured where the convergence state file landed (the probe passed no
`session_id`, so persistence was a no-op — I confirmed no `session_id` key in the
`audit_multi` request).

**Verdict: MET.** AC-6 asks whether the check looked. It looked, found two
recurrences, attributed both away from this defect on evidence, and found a
third defect **in the card's own work** which was then fixed in the same
milestone.

### Scope discipline (plan §D, §G)

```
$ git diff 1f9deed0c~1 34de07740 -- internal/cli/session.go internal/spec/
(empty)
```

`resolveProjectDir()`'s body is untouched; `internal/spec/audit.go`'s
`baseDir := opts.BaseDir` default is untouched. Every §G anti-pattern was
checked individually and none was hit — including the two the plan-audit
iterations added (`backendCall`-double-only, and absent/invalid-only assertions
on the codex path).

---

## Security

**Verified mechanically**: `golangci-lint run ./internal/cli/...` → `0 issues.`;
`go vet ./internal/cli/` → exit 0; the new file contains no `exec.Command`, no
`os.Setenv`, no `os.RemoveAll`, no symlink handling; every error message uses
`%q`, so a path with shell metacharacters cannot break the rendering.

**My own view of the exposure, formed before reading codex's.** The parameter
widens what an MCP caller reaches from "one tree the server resolved" to "any
directory on this filesystem containing a `.moai` subdirectory". Two distinct
consequences, and only the second matters much:

1. **Local read widening** — `spec_audit`/`spec_drift`/`spec_progress` will
   enumerate SPEC frontmatter from an unrelated project and return it to the
   model. Marginal, because the calling agent already holds `Read` and `Bash`.
2. **Egress widening** — `codex_audit` and the `audit_multi` fan-out launch the
   codex CLI with `cwd` set to that directory, and codex sends what it reads to
   a remote service. This is the one that is genuinely new in kind: it converts
   a directory *name* into a data-egress instruction, and the name can originate
   from text the model read (a SPEC body, a report, a dispatch). That is a
   confused-deputy shape — OWASP LLM-06 excessive agency crossed with LLM-02
   sensitive-information disclosure.

**Is REQ-3's validation the right boundary?** For what REQ-3 actually claims —
catch a mistyped path and refuse rather than fall back — it is exactly right,
and the error-not-fallback choice is the strongest single decision in this card.
The `.moai` check is a **sanity check, not a containment boundary**, and
critically, nothing in the code or the shipped documentation claims otherwise.
`validateProjectRoot`'s doc comment says "turns a caller-supplied path into an
absolute project root, or rejects it" — accurate. So there is no
security-invariant claim here that the code fails to keep, which is the failure
shape that would make this a High.

I therefore **do not ratify codex's framing** of finding 3 as a defect: the
constraint it proposes (require a shared git common directory) is a new
requirement, not a missing one. I do reach the same place on disposition — it
belongs in the follow-up — and I add one thing codex did not: `filepath.Abs`
does not resolve symlinks, so **if** the follow-up adds a git-common-dir
constraint, it must canonicalize with `filepath.EvalSymlinks` first or the
constraint is bypassable by a symlink. Recording that now is what stops the
follow-up from shipping a boundary that looks like one and is not.

**No Critical or High finding.** Security PASSES the must-pass gate.

---

## Craft

**Change-scoped coverage** (`go tool cover -func` over the SPEC's own tests —
a floor, since the filtered run excludes other tests that also reach these
functions):

```
mcp_project_root.go:82   resolveToolProjectRoot          100.0%
mcp_project_root.go:100  resolveOptionalToolProjectRoot  100.0%
mcp_project_root.go:111  validateProjectRoot              92.3%
mcp_convergence.go:406   performCodexAudit               100.0%
mcp_codex.go:1161        handleCodexAudit                 95.8%
mcp_audit_multi.go:45    handleAuditMulti                 93.3%
mcp_server.go:593        handleSpecAudit                  87.5%
mcp_server.go:533        handleSpecProgress                0.0%   ← F-2
mcp_server.go:612        handleSpecDrift                   0.0%   ← F-2
```

**Package-wide coverage**:

```
$ go test ./internal/cli/ -count=1 -cover -timeout 900s
ok  	github.com/modu-ai/moai-adk/internal/cli	378.307s	coverage: 78.5% of statements
```

78.5% is below the profile's 85% line. **It is not attributable to this change**,
and I state the inference rather than assert the conclusion: the card adds ~130
statements of production code covered at 92-100% plus ~530 lines of tests, and a
delta better covered than the package average raises the average, it cannot
lower it. I did **not** measure the pre-card baseline (that needs a second
378-second run at a commit I cannot check out from this worktree), so the
*magnitude* of the improvement is a Gap. Craft is not a must-pass dimension, so
this pre-existing package shortfall does not reach the verdict — but it is
recorded rather than argued away.

**Race**: `go test ./internal/cli/ -race -run '<the SPEC's tests>'` → `ok …
2.319s`. The new `codexReviewRPC` package var is swapped with `t.Cleanup` and no
test in the codex/convergence files calls `t.Parallel()`; the pattern matches the
pre-existing `codexLookPath` var idiom.

**Format**: one committed file is not gofmt-clean — see F-1.

**What is genuinely well made here**, and worth saying because it is the reason
the score is not lower: the comments carry the *reasoning* rather than restating
the code. `mcp_project_root.go:90-99` explains why two resolvers exist at all,
in terms of what breaks if they are merged. That is the comment that stops the
next editor from "simplifying" the two into one and silently breaking REQ-2 on
the fan-out path. Three separate mutation checks confirmed the tests around it
observe rather than decorate.

---

## Consistency

**Template-First mirrors** — the added content is byte-identical in all four:

```
$ diff .claude/rules/moai/core/moai-mcp-tools.md internal/template/templates/.claude/rules/moai/core/moai-mcp-tools.md ; echo rc=$?
rc=0
$ diff .claude/agents/moai/sync-auditor.md internal/template/…/sync-auditor.md ; echo rc=$?
rc=0
$ diff .claude/skills/moai-ref-cross-model-audit/SKILL.md internal/template/…/SKILL.md ; echo rc=$?
rc=0
```

`plan-auditor.md` diverges by 3 lines — a **pre-existing** divergence
(a SPEC-ID-bearing paragraph the template strips per the neutrality catalogue),
not introduced here. The M4 addition itself is byte-identical in both copies,
which I verified from the diff directly.

**Neutrality guard**: `MOAI_TEMPLATE_LEAK_STRICT=1 go test ./internal/template/...`
→ `ok … 20.577s`. The new agent-body text carries no SPEC ID, no date, no commit
SHA, no macOS path.

**Catalog**: the three hashes for the changed template files are updated in
`internal/template/catalog.yaml`. `moai-mcp-tools.md` needs none —
`grep -n 'moai-mcp-tools\|rules/moai/core' internal/template/catalog.yaml`
returns nothing, so rules are not catalog-tracked.

**The independence invariant** (the dispatch asked me to judge the reword on
substance, not wording). My reading:

- The original comment claimed the seam takes "(ctx, backend, target, focus) and
  NOTHING ELSE", with "a future edit that tried to thread claude_verdict in
  would not compile". That was always **stronger than what the type enforces** —
  adding a parameter is a compile-legal edit, and all the parameters were already
  untyped `string`, so a determined caller could have smuggled analysis through
  `focus` without any compiler complaint. The comment described a discipline, not
  a type-level guarantee.
- The reword ("carries **NO verdict of any kind**") is what the signature
  actually expresses today, and it is *more* honest than what it replaced, not
  less. It also names why it was reworded, which is the part that stops the next
  reader from restoring the false version.
- **Did widening weaken independence in substance?** No. `projectRoot` names a
  directory and carries no analysis; it reaches `performCodexAudit` and becomes a
  `cwd`, and reaches GLM not at all. Nothing in the widening creates a channel
  from `claude_verdict` to a backend. `runMultiAudit` still consumes
  `claudeVerdict` only in `converge`.
- The compile-time guard is preserved in the right place and honestly labelled:
  `var _ backendCallFn = func(_ context.Context, _, _, _, _ string) ReviewOutput`
  at the end of the test file breaks if a sixth `ReviewOutput` parameter appears.
  It is a change-detector rather than a semantic guard — it would also break on a
  benign signature change — but that is the correct strength for the claim, and
  the SKILL.md prose keeps the operative rule (do not paste Claude's reasoning
  into `focus`) which is the part no type can enforce.

**Judgment: the reword preserves the invariant and improves its accuracy.** This
is the one place where a careless edit would have left a comment guarding
something the code no longer did, and the card handled it deliberately — the
plan named it as a hazard at iter-3 and the change discharged it.

**MCP version skew** (scored here, as the dispatch directed). Every
`mcp__moai__*` tool this session can reach is served by a pre-M1 process and acts
on the primary checkout. This is not a defect in the card — it is the card's own
thesis, observed in the wild: a long-lived subprocess does not reload when the
binary beneath it changes. It is correctly disclaimed in the M5 report's "does
NOT establish" section. The operational consequence is real and belongs in the
follow-up: **the repair reaches nobody until the binary is installed and every
running server is restarted**, and nothing announces which state a given session
is in.

---

## Findings

Severity uses the profile's scale. Confidence is my own.

### Blocking (fix before merge)

**None.**

I am stating that as a finding rather than as an absence. I looked for one
specifically at the two places the dispatch pointed me — the fan-out assertion
and the two-resolver AC-2 claim — and both survived mutation testing. The card's
own M5 check found the one real defect in this work (the shared description) and
fixed it in the same milestone with a regression test that I confirmed observes.

### Optional (follow-up; the orchestrator decides, and none of these gate the merge)

| # | Severity / Confidence | Location | Finding | Required fix |
|---|---|---|---|---|
| **F-1** | Low / certain | `internal/cli/mcp_project_root_test.go:133-135` | Committed file is not gofmt-clean (map literal key alignment). `golangci-lint` does not flag it and no CI job runs `gofmt`, so nothing catches it — it will surface as a spurious diff in the next editor save. | `gofmt -w internal/cli/mcp_project_root_test.go`. One-line whitespace change; zero behavioural risk. |
| **F-2** | Low / certain | `internal/cli/mcp_server.go:533`, `:612` | `spec_progress` and `spec_drift` get the redirect **schema-asserted only**. Their parameter-present behaviour is never exercised — not by a unit test (0.0% in the SPEC's own coverage run) and not by the M5 live probe, which called only `spec_audit`. AC-1a names `spec_audit`, so this is not an AC miss; it is 2 of the 5 in-scope tools resting on read-the-diff assurance. Risk is genuinely low (both share the 100%-covered `resolveToolProjectRoot`), which is why this is optional. | Two tests reusing the existing `newProbeProject` helper, mirroring `TestSpecAudit_ProjectRootRedirectsTheCatalogue`. ~10 lines each. |
| **F-3** | Medium / high | `internal/cli/mcp_convergence.go:381-384` | `audit_multi` accepts a `project_root` that one of its two backends structurally cannot use, and **the result does not say so**. A caller reading GLM's verdict has no way to learn it is about no tree. The M5 run is the demonstration: GLM returned a confident `fail` about code it never saw, indistinguishable in the output shape from a grounded one. | Surface it in the result — a per-backend `tree_read` (or a `root_ignored` note on GLM's verdict). The card's report already raises this; it needs an owner. |
| **F-4** | Medium / high | `internal/cli/mcp_convergence.go:590`, `persistConvergenceResult` | Convergence state persists to `resolveProjectDir()`, ignoring `cfg.ProjectRoot`, so a worktree audit writes its verdict into the primary checkout's `.moai/state/audit-multi/`. Already row 6 of the verdict table (deferred) and already codex's finding 1. **I agree with codex that it is a live defect rather than an open question, and with the table that it is out of this card's scope** — the two are not in conflict: what is undecided is the *keying*, not whether the current behaviour is wrong. Unmeasured either way: the M5 probe passed no `session_id`, so persistence was a no-op. | Follow-up card. First step is the measurement the table names, not the repair. |
| **F-5** | Low / certain | `.moai/reports/t171/post-repair-check.md`, Stage 1 table | The column *"Contains the worktree-only SPEC"* reads `yes`, but the SPEC ID appears nowhere in the with-parameter result — a clean SPEC produces no drift finding. The supporting observation is the count delta (627→628, `modern_era_clean` 349→350), which the prose states correctly. The table column claims an observation that was not made. | Retitle the column to what was measured (`total_specs` / `modern_era_clean` delta), or state the SPEC's absence-by-cleanliness in a footnote. |
| **F-6** | Low / medium | `internal/cli/mcp_project_root.go:112` | `filepath.Abs` does not resolve symlinks. Harmless **today**, because there is no containment boundary to bypass. It becomes load-bearing the moment the follow-up adds codex's suggested git-common-dir constraint — a symlink would walk straight through it. | Record it on the follow-up card that would add the constraint: canonicalize with `filepath.EvalSymlinks` **before** comparing git dirs, or the boundary is decorative. |
| **F-7** | Low / medium | operational, not code | The repair reaches no running session until the binary is installed and each long-lived `moai mcp-server` is restarted, and nothing tells a session which side of that line it is on. This audit hit it directly — every `mcp__moai__*` tool available to me still answers about the primary checkout. | Follow-up: either surface the server's build identity in a tool result, or add a note to `moai-mcp-tools.md` § `project_root` telling an agent how to tell whether its server has the parameter (call `ListTools` and look for it — which the M5 probe already demonstrates works). |

### Observation (no action requested, recorded so it is not rediscovered)

`.moai/specs/SPEC-MCP-WORKTREE-ROOT-001/.moai/state/{config-cache,context-usage}.json`
exists on disk (mtime 13:02, during the run phase). Something resolved a project
dir to the SPEC directory and wrote state there. It is gitignored
(`.gitignore:212 **/.moai/state/`) so nothing is committed and the tree is clean.
The writers are not `resolveProjectDir()` call sites in `internal/cli`, so this is
outside AC-4's grep-derived mandate and is **not** a miss in the verdict table —
but it is another consumer resolving a root wrongly, which is exactly the family
the follow-up card is about, and worth one line there.

---

## Baseline-attribution

Every command below was run by me, in this worktree, at HEAD `34de07740`, with
the tree confirmed clean before and after.

| Command | Observed |
|---|---|
| `git rev-parse --short HEAD` | `34de07740` (matches the pin) |
| `go test ./internal/cli/ -run '<SPEC tests>' -count=1` | `ok … 0.848s` |
| `go test ./internal/cli/ -race -run '<SPEC tests>' -count=1` | `ok … 2.319s` |
| `go test ./internal/cli/ -count=1 -cover -timeout 900s` | `ok … 378.307s coverage: 78.5% of statements` |
| `go vet ./internal/cli/` | exit 0 |
| `golangci-lint run ./internal/cli/...` | `0 issues.` |
| `gofmt -l <7 changed Go files>` | `internal/cli/mcp_project_root_test.go` |
| `MOAI_TEMPLATE_LEAK_STRICT=1 go test ./internal/template/...` | `ok … 20.577s` |
| `grep -rn 'resolveProjectDir()' internal/ \| grep -v _test.go \| wc -l` | `25` (→ 18 live call sites) |
| 3 × mutation check + restore | each turned a green test red; `git status --porcelain` empty afterwards |

**Reused as attributed** (the dispatch offered these; I re-ran the first with
`-cover` so it is now my own measurement, and I re-ran the third):
`GOOS=windows GOARCH=amd64 go vet ./internal/cli/` → exit 0 — **not re-run by
me**, taken as attributed from the dispatch.

---

## Gaps — what I did NOT observe

- **The pre-card coverage baseline.** I measured 78.5% at this HEAD but not at
  `1f9deed0c~1`, so the *direction* of the change is reasoned (better-covered
  delta raises an average) rather than measured. Settling it needs one more
  378-second run at a commit this worktree cannot check out.
- **`GOOS=windows` vet.** Attributed to the dispatch, not re-run. Note the
  standing caveat that cross-compile vet proves compilation, not behaviour — and
  that it does not compile test files at all.
- **The M5 probe was one run, and I did not repeat it.** Both backends are
  non-deterministic. I verified the report against the JSON it produced; I did
  not verify that a second run would produce the same shape.
- **`spec_progress` / `spec_drift` parameter-present behaviour** — never
  exercised by anything (F-2). I read the handler wiring in the diff; I did not
  observe it running.
- **Where the convergence state file actually lands** (F-4). The probe passed no
  `session_id`, so persistence was a no-op — I confirmed the absence of the key
  in the request rather than inferring it from the report.
- **The full `./...` suite.** Not run, per the dispatch's load constraint.
  Targeted package runs only.
- **No `mcp__moai__*` tool was used for any claim in this report**, because the
  session's server predates the change and would have answered about the primary
  checkout.

## Residual risk

- **The most serious defect in the neighbourhood is not this card's, and it is
  still live**: codex's adversarial path synthesizes `verdict: "pass"` from text
  reading `Verdict: fail — merge blocked.`, and the convergence engine then
  reports `overall_verdict: "pass"`. A **required** gate silently inverts. This
  card found it, attributed it correctly, and routed it to a follow-up. Until
  that card lands, an `audit_multi` `pass` is not trustworthy on the codex path —
  including for whoever reads this SPEC's own audit trail.
- **GLM's confident `fail` about nothing** (F-3) is the same hazard from the
  other direction: a fail-shaped output with no grounding is worse than
  `inconclusive`, because it looks like a finding.
- The repair is correct but **inert until deployment** (F-7): every currently
  running MCP server still audits the primary checkout, and the failure stays
  silent by construction — which is precisely the property this SPEC exists to
  remove.
- Two of five tools rest on read-the-diff assurance (F-2). Low risk, non-zero.

---

*Verdict authority rests with this agent. The findings list above is the
structured defect-list: zero blocking, seven optional. An all-optional list does
not convert a PASS into a FAIL, and I am not manufacturing a blocker to look
rigorous — the two claims most likely to be flattering were mutation-tested and
both held.*
