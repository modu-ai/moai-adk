# SPEC-TABSCHEMA-AUTOBRANCH-001 — Sync-phase Independent Audit (card t316)

Auditor: sync-auditor (independent, fresh judgment)
Worktree: `/Users/goos/MoAI/moai-adk-go/.claude/worktrees/t316`
Branch: `WT-tabschema-autobranch` · Audited HEAD: `e36cdcf2b`
Plan-phase baseline: `7ed6edb3e` · Run-phase commit: `914c4edf5`
Integration: `914c4edf5` and `e36cdcf2b` are both ancestors of `origin/develop` (`6310dbf28`) — verified by `git merge-base --is-ancestor`.
Evaluator profile: `default` (`.moai/config/evaluator-profiles/default.md`) — Functionality 40 / Security 25 / Craft 20 / Consistency 15; must-pass = Functionality + Security.

**Verdict: PASS** — weighted harmonic mean **95.3** against the Tier S threshold **75**.
Findings: 0 BLOCKING · 0 MAJOR · 2 MINOR (both optional) · 3 INFO.

---

## 1. Claim

The audited change deletes exactly two question objects from `tab_schema.json` — the ones bound to
the dead paths `git_strategy.personal.auto_branch` (batch 3.3) and `git_strategy.team.auto_branch`
(batch 3.6) — in both the local copy and its template mirror, leaving batch 3.10's canonical
`git_strategy.{mode}.automation.auto_branch` question untouched, both copies byte-identical and
valid JSON, the embedded asset regenerated, and no other question object removed or altered.

Every claim below was **re-measured by this auditor on this tree**, not read from the run-phase
report. Where the run phase recorded a number, this audit reran the command and compared.

---

## 2. Evidence (verbatim command output, this run, this tree)

### 2.1 Scope of the change — measured, not taken on trust

```
$ git diff --stat 7ed6edb3e..e36cdcf2b
 .../moai-workflow-project/schemas/tab_schema.json  |  42 --
 .moai/reports/t316/plan-audit-iter1.md             | 702 +++++++++++++++++++++
 .moai/reports/t316/run-verdict.md                  | 235 +++++++
 .../SPEC-TABSCHEMA-AUTOBRANCH-001/acceptance.md    | 420 ++++++++++++
 .moai/specs/SPEC-TABSCHEMA-AUTOBRANCH-001/plan.md  |  97 +++
 .../SPEC-TABSCHEMA-AUTOBRANCH-001/progress.md      | 160 +++++
 .moai/specs/SPEC-TABSCHEMA-AUTOBRANCH-001/spec.md  | 181 ++++++
 .../moai-workflow-project/schemas/tab_schema.json  |  42 --
 8 files changed, 1795 insertions(+), 84 deletions(-)

$ git status --porcelain
(empty)
```

The only non-artifact files touched are the two `tab_schema.json` copies. No Go production code, no
config, no hook. Confirms REQ-TSA-006's containment half at the file level and the DoD's
"no file outside LOCAL / TMPL / the SPEC directory" bullet.

The diff itself is two pure-deletion hunks per copy, each 21 lines, removing only the two named
question objects (read verbatim from `git diff 7ed6edb3e..e36cdcf2b -- '*/schemas/tab_schema.json'`).

### 2.2 Acceptance battery — re-executed independently

`LOCAL` = `.claude/skills/moai-workflow-project/schemas/tab_schema.json`
`TMPL` = `internal/template/templates/.claude/skills/moai-workflow-project/schemas/tab_schema.json`

```
=== AC-TSA-001 LOCAL ===          === AC-TSA-001 TMPL ===
personal=1                        personal=1
team=1                            team=1
manual=0                          manual=0

=== AC-TSA-002 ===
LOCAL personal=0 team=0
TMPL  personal=0 team=0

=== AC-TSA-003 ===
LOCAL canonical=3
TMPL  canonical=3

=== AC-TSA-004 ===
LOCAL total_auto_branch=3
TMPL  total_auto_branch=3

=== AC-TSA-005 ===
diff_rc=0

=== AC-TSA-006 ===
local_rc=0
template_rc=0

=== AC-TSA-007 ===
TOTAL_QUESTIONS = 46   (LOCAL)
TOTAL_QUESTIONS = 46   (TMPL)

=== AC-TSA-007b ===
REMOVED_OBJECTS = 2 ['git_strategy.personal.auto_branch', 'git_strategy.team.auto_branch']
IDENTICAL_AFTER_REMOVAL = True     (LOCAL)
REMOVED_OBJECTS = 2 ['git_strategy.personal.auto_branch', 'git_strategy.team.auto_branch']
IDENTICAL_AFTER_REMOVAL = True     (TMPL)

=== AC-TSA-008 (scan on TMPL) ===
3:  "schema_updated": "2025-12-22",
604:              "question": "Branch prefix for Personal mode? ... (e.g., feature/SPEC-) ...",
873:              "question": "Branch prefix for Team mode? ... (e.g., feature/SPEC-) ...",
=== content-only compare against the 7ed6edb3e match set ===
neutrality_diff_rc=0
```

AC-TSA-007b is the criterion that decides "nothing else was altered", and it reads `True` on both
copies against the pinned baseline `7ed6edb3e` — not against `HEAD`, which would be vacuous.

Structural cross-check of the three affected batches (post-change), read from the parsed schema:

```
3.3  | n_questions=3 | cond= {'field': 'git_strategy.mode', 'operator': 'equals', 'value': 'personal'}
       git_strategy.personal.workflow / .environment / .github_integration
3.6  | n_questions=3 | cond= {'field': 'git_strategy.mode', 'operator': 'equals', 'value': 'team'}
       git_strategy.team.workflow / .environment / .github_integration
3.10 | n_questions=3 | cond= {'field': 'git_strategy.mode', 'operator': 'not_equals', 'value': 'manual'}
       github.enable_trust_5 / github.auto_delete_branches / git_strategy.{mode}.automation.auto_branch
```

Batch 3.10 — the landed output of SPEC-SYNC-STRATEGY-KEY-001 (card t303) that REQ-TSA-005 exists to
protect — is intact and still carries the canonical field.

### 2.3 AC-TSA-005b — the embedded asset (the criterion most worth falsifying)

`bin/moai` exists on this tree, produced by the run phase's `make build` at `2026-08-27 20:39:37`;
the only later commit (`e36cdcf2b`, 20:46) is a `progress.md` docs backfill that cannot change the
embedded schema. Scan of that artifact:

```
personal=0
team=0
canonical_control=4

$ grep -aoF 'Auto-create branches for Personal mode?' bin/moai | wc -l
0
$ grep -aoF 'Auto-create a branch for each SPEC implementation?' bin/moai | wc -l
1
```

Items 1-3 of AC-TSA-005b hold: dead-path strings absent, control string present. The two extra
question-text probes are this auditor's own addition — they confirm the deleted question's prose is
gone from the binary while the surviving canonical question's prose is present, which is a stronger
signal than the field-path counts alone (a field path could in principle be shared).

The control reading `4` rather than `3` is reproduced and explained: a second embedded file,
`internal/template/templates/.claude/skills/moai/workflows/sync/delivery.md`, carries one occurrence
of the canonical string. `3 + 1 = 4`. The run phase's deviation #1 is accurate.

I deliberately did **not** re-run `make build`: the existing artifact is attributable, and a rebuild
would mutate the tree outside my report.

### 2.4 Consumer search — does the deletion break anything?

```
$ grep -rn 'git_strategy\.personal\.auto_branch\|git_strategy\.team\.auto_branch' --exclude-dir=.git .
```

Every hit is documentary: this SPEC's own `spec.md` / `plan.md` / `acceptance.md` / `progress.md`,
`.moai/reports/t316/*`, and two prior records (`SPEC-SYNC-STRATEGY-KEY-001/progress.md:62`,
`.moai/reports/t303/run-verdict.md:92`) that flagged these exact paths as wrong and out of their
scope. **Zero** hits in `internal/`, `pkg/`, `cmd/`, in any `*.sh`, in any skill body, or in any
template other than the schema itself.

The dead-path claim is verified against the Go type, not inferred. `ModeProfile`
(`internal/config/types.go`) has no `auto_branch` field — its members are `Workflow`,
`Environment`, `GitHubIntegration`, `PushToRemote`, `AutoCheckpoint`, `BranchPrefix`, `MainBranch`,
`DraftPR`, `RequiredReviews`, `BranchProtection`, `MergeMethod`, and the nested `Automation`,
`BranchCreation`, `CommitStyle`, `Hooks`. `AutomationConfig.AutoBranch` carries
`yaml:"auto_branch"`, so `git_strategy.{mode}.automation.auto_branch` is the real path. The
top-level `GitStrategyConfig.AutoBranch` (`types.go:158`) is a third, distinct, explicitly
deprecated flat key and is untouched by this change.

The surviving canonical key **is** live at the config surface, independent of the schema:
`internal/template/templates/.moai/config/sections/git-strategy.yaml.tmpl` emits
`automation: auto_branch:` for the mode profiles, and the sync skill body
`.claude/skills/moai/workflows/sync/delivery.md:31,39` names
`git_strategy.{mode}.automation.auto_branch` as the branch-handling axis.

### 2.5 Mechanical verification per dimension

```
$ go test ./internal/template/...
ok  github.com/modu-ai/moai-adk/internal/template          28.856s
ok  github.com/modu-ai/moai-adk/internal/template/agentemit  0.811s
?   github.com/modu-ai/moai-adk/internal/template/scripts  [no test files]

$ go test ./internal/template/... -run 'TestTemplateNeutralityAudit|TestInternalContentLeak' -count=1
--- PASS: TestTemplateNeutralityAuditC8Preserve (0.05s)
--- PASS: TestTemplateNeutralityAudit (0.00s)
ok  github.com/modu-ai/moai-adk/internal/template  0.502s

$ ./bin/moai spec lint > <scratch>/spec-lint.txt 2>&1; echo "spec_lint_rc=$?"
spec_lint_rc=0
$ grep -c 'SPEC-TABSCHEMA-AUTOBRANCH-001' <scratch>/spec-lint.txt
0
$ tail -1 <scratch>/spec-lint.txt
0 error(s), 64 warning(s)

$ git diff --numstat 7ed6edb3e -- <LOCAL>
0	42
$ git diff --numstat 7ed6edb3e -- <TMPL>
0	42
```

`TestTemplateNeutralityAudit` is the guard the CI workflow `.github/workflows/template-neutrality-check.yaml`
runs on any `internal/template/templates/**` change; it passes on this tree. The
`internal_content_leak_test.go` allowlist row for this file (line 918) pins the date `2025-12-22`
on line 3, which the change does not touch — so the change cannot have invalidated it, and the
package suite confirms it did not.

---

## 3. Baseline-attribution

| Claim | Measured against |
|---|---|
| AC-TSA-001/002/003/004/006/007 counts | The two working-tree files at `e36cdcf2b`, this run |
| AC-TSA-005 identity | `diff -q` of the two files at `e36cdcf2b`, this run |
| AC-TSA-005b binary scan | `bin/moai` built by the run phase at `2026-08-27 20:39:37` from this tree; the only subsequent commit is a docs-only backfill |
| AC-TSA-007b deep equality | `git show 7ed6edb3e:<path>` — the pinned plan-phase baseline, **not** `HEAD` |
| AC-TSA-008 neutrality delta | The `SPEC-\|20[0-9][0-9]-` match set of `git show 7ed6edb3e:<TMPL>`, compared by content |
| Dead-path liveness | `internal/config/types.go` struct tags at `e36cdcf2b`; repo-wide grep excluding `.git` |
| Merge state | `git fetch origin develop` then `git merge-base --is-ancestor`, this run |

No figure in this report is carried over from `run-verdict.md` or `progress.md`. Where the run
phase's number and mine agree, both are stated; where a number came only from the run phase, it is
listed in Gaps.

---

## 4. Gaps — what this audit did NOT observe

1. **No runtime execution of the interview.** `tab_schema.json` has no runtime consumer (see F3), so
   the interview semantics AC-TSA-001 encodes were never executed. The `mode_admits` predicate
   remains a reconstruction with no executable oracle — this audit re-ran the same reconstruction,
   which is corroboration, not independent validation.
2. **`make build` not re-run by this auditor.** AC-TSA-005b was discharged against the run phase's
   existing `bin/moai` rather than a freshly built one, to avoid mutating the tree. A rebuild from
   the current sources is therefore asserted-by-inference (sources are correct ⇒ a build embeds
   them), not observed in this run.
3. **No cross-platform build.** No GOOS matrix build was run here either; the run phase declared the
   same and deferred it to CI. For a JSON asset with no OS-conditional content this is reasonable,
   but it is unobserved.
4. **No full-suite run.** Only `./internal/template/...` was exercised, per the repository's
   standing prohibition on local full-suite runs. Packages outside `internal/template` were not run;
   the change touches no Go code, so the blast radius is bounded, but the full verdict belongs to CI
   on `origin/develop`.
5. **CI status on `origin/develop` after the merge (`6310dbf28`) was not read.** This audit judged
   the tree, not the pipeline.
6. **`manual`-mode `automation` block still confirmed structurally only.** As in plan phase: the
   claim rests on `ModeProfile` being one type for all three modes, not on a rendered manual-mode
   `git-strategy.yaml`.

---

## 5. Residual-risk

- **AC-TSA-007b compares parsed JSON**, so a pure reformat of untouched regions (re-indentation, key
  reordering) would still read `True`. The `0/42` numstat makes this unlikely here — a reformat
  would show additions — but no criterion mechanically rejects it.
- **The binary scan is attributable only while the dead-path strings stay confined to the schema.**
  Measured now: zero occurrences in `internal/`, `pkg/`, `cmd/` outside the two schema copies. A
  future change introducing those strings into compiled code would silently de-attribute AC-TSA-005b.
- **The schema is an orphan (F3).** Nothing mechanically detects a future regression to the dead
  spelling: no parser, no test, no lint reads this file. The correctness this card establishes is
  preserved by convention only.
- **The header counters stay wrong** and the `total_settings` drift widens from `60 vs 48` to
  `60 vs 46` (verified: header reads `total_settings: 60`, `total_batches: 18`; computed 46 / 17).
  Recorded out of scope in `spec.md` §4, so the widened number is attributable to the pre-existing
  drift and not to this deletion — but a later reader who skips §4 will mis-attribute it.
- **Manual-mode operators are now visibly never asked about auto-branching.** The behaviour is
  unchanged (their answer previously went to a dead path), but the silence is no longer masked.
  `spec.md` §4 records this as a separate design question for a separate card.

---

## 6. Dimension scores

| Dimension | Score | Verdict | Evidence |
|-----------|-------|---------|----------|
| Functionality (40%) | 98/100 | PASS | All 10 criteria (AC-TSA-001..008 + 005b + 007b) independently re-measured GREEN (§2.2, §2.3). Structural batch inspection confirms `personal=1 / team=1 / manual=0` and batch 3.10 intact. Repo-wide consumer search finds zero broken readers (§2.4). |
| Security (25%) | 96/100 | PASS | Change is a pure deletion from a JSON interview asset. No credential, no input-handling path, no trust boundary, no network or filesystem surface touched. No Go code changed (`git diff --stat`). No secret introduced (`grep` of the diff: two deletion hunks only). No Critical/High finding. |
| Craft (20%) | 92/100 | PASS | Minimal, purely subtractive diff (`0/42` per copy — no incidental churn). Both copies remain valid JSON after final-element deletion (`json.tool` rc=0 ×2). `internal/template` suite green (28.9s). Run-phase deviations recorded rather than smoothed (§7 honesty check). Deduction: DoD checkboxes left unticked (F1). *Coverage threshold N/A* — the profile's `Coverage >= 85%` gate presupposes production code; none was changed, so it is reported as not-applicable rather than as a PASS. |
| Consistency (15%) | 92/100 | PASS | Template-First honored: template and mirror byte-identical (`diff_rc=0`), embedded asset regenerated and verified. Conventional-commit messages carrying the card id on every commit. Batch 3.10's canonical spelling — another card's landed output — preserved byte-for-byte. Neutrality guard green. Deduction: SPEC frontmatter still `status: in-progress` and `progress.md` §E.4 is `_<pending sync-phase>_` (F2). |

Must-pass firewall: Functionality **PASS**, Security **PASS** — both independently above threshold,
so the firewall does not force a FAIL.

**Weighted harmonic mean** = 0.40/0.25/0.20/0.15 over (98, 96, 92, 92)
= 1 / (0.40/98 + 0.25/96 + 0.20/92 + 0.15/92) = 1 / 0.0104901 = **95.3**

Tier S threshold: **75**. 95.3 ≥ 75.

---

## 7. Honesty check on the run-phase deviations (§E.2)

The dispatch asked specifically whether the three recorded deviations are honest and complete. All
three were re-measured:

| Deviation | Run-phase statement | This audit's measurement | Verdict |
|---|---|---|---|
| #1 control reads `4` not `3` | second embedded file `sync/delivery.md` carries the canonical string | `canonical_control=4`; `grep -rlF` confirms `delivery.md` as the extra carrier | Accurate |
| #2 numstat predicted `2/44`, measured `0/42` | git models the trailing-comma strip inside one contiguous deletion | `git diff --numstat` → `0 42` on both copies | Accurate |
| #3 AC-TSA-008 discharged on line content, prefixes shifted `625→604`, `915→873` | prefixes necessarily shift under a deletion above them | scan returns lines `3`, `604`, `873`; content-only compare `neutrality_diff_rc=0` | Accurate |

No deviation was found that the run phase failed to record. The one further discrepancy this audit
could construct — that `progress.md` §E.3's `run_commit_sha: 914c4edf5` was written by a *later*
commit than the one it names — is disclosed in the file itself (`this line backfilled by
914c4edf5's successor`) and in the commit message of `e36cdcf2b`. That is a disclosed backfill, not
a concealed one.

---

## 8. Findings

- **F1 [MINOR] [optional]** `.moai/specs/SPEC-TABSCHEMA-AUTOBRANCH-001/acceptance.md:412,414,417,419`
  — all four Definition-of-Done checkboxes remain `- [ ]` although every item is discharged with
  recorded evidence (all criteria GREEN, `make build` rc=0, `spec lint` rc=0 with zero matches,
  scope containment clean). Cosmetic traceability drift only; the evidence, not the checkbox, is
  what the DoD is satisfied by. Required fix (if taken): tick the four boxes during sync-phase
  artifact closure.
- **F2 [MINOR] [optional]** `spec.md:5` reads `status: in-progress` and `progress.md` §E.4 reads
  `_<pending sync-phase>_`, while the change is already merged into `origin/develop`. This is the
  sync phase's own remaining stamping obligation rather than a defect in the delivered change; it is
  named here so the close is not declared before the frontmatter transition and the §E.4 signal land.
  Required fix: complete the sync-phase status transition and fill §E.4.
- **F3 [INFO]** `tab_schema.json` has **no runtime consumer** — verified independently, not taken
  from `spec.md` §4. A repo-wide `grep -rln 'tab_schema'` returns only SPEC documents, reports,
  `.moai/manifest.json` (a deployment file listing), and `internal_content_leak_test.go:918` (a
  date-allowlist row). The owning `SKILL.md` does not reference the file at all
  (`grep -n 'schema' .claude/skills/moai-workflow-project/SKILL.md` → two hits, both naming
  `references/configuration.md`, neither naming `schemas/`). Reported as a fact per the dispatch;
  no action recommended beyond the card's scope.
- **F4 [INFO]** The surviving canonical key path **is** genuinely live at the config surface,
  independent of the orphaned schema: `git-strategy.yaml.tmpl` emits `automation.auto_branch` for
  the mode profiles, and `sync/delivery.md:31,39` names `git_strategy.{mode}.automation.auto_branch`
  as the branch-handling axis. So the deletion moved the schema toward a spelling that a real
  consumer uses, even though the schema itself has no reader.
- **F5 [INFO]** Header counter drift persists and widens (`total_settings: 60` vs computed 46;
  `total_batches: 18` vs computed 17). Pre-existing, recorded out of scope in `spec.md` §4,
  re-measured here so the attribution is on record.

No blocking finding. An all-optional findings list does not convert PASS into FAIL.

---

## 9. Verdict

**PASS** — 95.3 / threshold 75 (Tier S). Functionality and Security both clear the must-pass
firewall independently. The change does exactly what it claims, nothing more; every acceptance
criterion reproduces GREEN under independent re-execution; the deletion breaks no consumer because
no consumer of the deleted paths exists anywhere in the tree; and the run phase's recorded
deviations are accurate and complete.
