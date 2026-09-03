# SPEC-HOOK-WIRING-DRIFT-001 — Implementation Plan

## §A Context

Four milestones, drawn from the three investigation reports in
`.moai/reports/t216/`. The design content those reports already carry — d1 §E6's
five-option table for the update path, d3 §6's six-option table for the MX scan,
d2's per-script disposition table — is **referenced, not duplicated**; the
decisions taken from them are recorded in `spec.md` §C-2 and §A.4 and in §F
below.

**Milestone order is by decision-reversibility, not by dependency.** The two
milestones that create or change a user-facing surface come first, so review
attention lands where a wrong call is expensive to walk back. M1, the mechanical
parity commit, comes last despite being the card's headline — its content is
fully determined by the template, so there is nothing to decide.

| Order | Milestone | Why here |
|---|---|---|
| 1 | **M2** — doctor drift diagnostic | New user-facing surface. Output shape, check name, and the report-vs-repair boundary are all decisions that are costly to change once shipped |
| 2 | **M4** — MX consumer auto-build + hook-side removal | Changes the behaviour of a **live** command (`moai mx query`, invoked by the `/moai review` lean audit) and deletes ~40 lines from the SessionStart path |
| 3 | **M3** — disposition taxonomy in the rule surface | The five-class taxonomy is naming that downstream readers will inherit; it ships to 16-language distribution |
| 4 | **M1** — the settings.json parity commit | Mechanical. Every byte is dictated by the rendered template |

**Two ordering consequences cross this.**

1. **AC-HWD-003's parity test cannot be green until M1 lands.** Land the test
   **red** with M2 (it is the same render-and-compare machinery the diagnostic
   needs), then M1 turns it green. Do not weaken the test to make it pass early.
2. **Every M2 failing input is built from the rendered template, never by copying
   the project** (audit D3). At M2 time the project's `settings.json` is still
   drifting — `grep -c 'chain-event.sh' .claude/settings.json` → `0` — so there is
   no `chain-event.sh` entry to remove and no drift-free copy to take. The three
   fixtures (drift-free, template-only, project-only) and how each is produced are
   specified in acceptance.md §M2's construction note; build them from the render
   and the ordering is a non-issue.

---

## §B Known issues carried in from the investigation

1. **Wiring `chain-event.sh` is inert** (d1 §E3). M1 delivers parity, not
   function. Any comment or commit message claiming otherwise fails AC-HWD-004.
2. **`.claude/settings.json` is hand-maintained, not rendered** (d1 §E2). There
   is no "regenerate" path to invoke; M1 is an edit.
3. **The `moai update` path cannot deliver a hook entry** (d1 §E4). No step in
   this plan may propose `moai update` as a remedy.
4. **`.claude/rules/moai/**` is wiped wholesale by `moai update`**
   (`CleanMoaiManagedPaths`). M3's local mirror is only durable because the
   template twin carries the same content — hence the Template-First order is
   load-bearing, not ceremonial.
5. **The prior audit reports are point-in-time records** and are not rewritten
   (spec.md §G-6). The corrections land in this SPEC and in the rule surface.

---

## §C Pre-flight

```bash
git rev-parse --show-toplevel      # must be .../worktrees/t216
git rev-parse --short HEAD         # 950cb4399 at authoring time
git branch --show-current          # WT-hook-wiring-drift
git status --short                 # re-read immediately before each commit
make build                         # verification uses bin/moai, never the stale global binary
```

The globally-installed `moai` is v3.1.2 and predates this tree. Every acceptance
command that invokes `moai` uses the freshly built binary.

---

## §D Constraints in force

`spec.md` §C 1-6. The two that shape the code most: **report-never-repair**
(§C-2) governs M2, and **template neutrality** (§C-4) governs M3's template edit
— the card numbers `t242`/`t243`/`t244` may appear in this SPEC and never in
`internal/template/templates/**`.

---

## §E Self-verification

Run per milestone, not once at the end:

```bash
go vet ./internal/cli/... ./internal/hook/... ./internal/template/...
go test ./internal/cli/... -timeout 600s
go test ./internal/hook/...
go test ./internal/template/...
golangci-lint run
```

`internal/cli` has a measured ~336 s single-package runtime; a 600 s floor is
mandatory or a passing tree reports FAIL. Do **not** run `go test ./...`
locally — push and read CI for the full-suite verdict.

---

## §F Milestones

### M2 — `moai doctor` hook-wiring drift diagnostic

**Decision taken (from d1 §E6):** option **C** — report, change nothing. Option
A (extend `pruneToShared` to arrays) is rejected in `spec.md` §C-2; option B
(deploy-time snapshot) is deferred to §G-5. The rationale is the hard case: with
no snapshot base, "the template gained this entry" and "the user deliberately
deleted it" are indistinguishable, so any auto-repair silently overrides intent.

**Approach.** Add `checkHookWiringDrift(projectRoot string, verbose bool)
DiagnosticCheck` to `internal/cli/doctor.go`, registered in the check list beside
`checkHooksConfig` (`doctor.go:192`) under the name `Hook Wiring`. It:

1. Renders `.claude/settings.json.tmpl` in memory —
   `template.EmbeddedTemplates()` + `template.NewRenderer(fsys)`, with a
   `TemplateContext` whose `HookOptIn.Enabled` is read via
   `config.LoadSystemHookOptInEnabled(projectRoot)` (the same helper
   `readHookOptInEnabled` uses at `internal/cli/update.go:1204`), so the render
   matches what this project would actually receive.
2. Parses both the render and `<projectRoot>/.claude/settings.json` into hook
   entries keyed on `(event, matcher, script, if, timeout, async)`.
3. Emits a set-diff in **both** directions: template-only entries (missing
   registration) and project-only entries (extra registration), each naming its
   script.
4. Returns `CheckOK` on an empty diff, `CheckWarn` otherwise — never
   `CheckError`, and never a write.

**Extraction note.** The keyed-entry parse is shared with M2's parity test
(AC-HWD-003). Put it in one place — a small unexported helper in
`internal/template` (beside the existing settings tests) or an internal package
both can import — rather than writing the comparison twice. Two copies of a
comparison drift, and a drifted drift-detector is a poor joke.

**[HARD] The template source is an injectable parameter** (audit D2 / mutant
M-2, and the fresh attempt (ii) recorded under AC-HWD-017). The check MUST take
its template source — an `fs.FS`, or the rendered bytes — as a parameter rather
than reaching for the embedded FS internally, and production MUST pass
`template.EmbeddedTemplates()` through that same seam. Two reasons, and the
second is the load-bearing one:

- Without the seam, AC-HWD-017 cannot feed a fixture template and the
  render-vs-hardcode distinction is unobservable by any criterion. A hardcoded
  expected-entry list would then pass every M2 criterion — and it rots against the
  template, re-creating the exact drift class this SPEC exists to close.
- Putting the seam on the **check** rather than only on the helper is what closes
  the one-level-up variant: a correct shared helper plus a check that ignores it
  and consults a hardcoded list. AC-HWD-017 observes the **check's own returned
  message**, so the check must consume what it is handed.

**Failure modes (REQ-HWD-005).** Render error, missing settings file, unparseable
JSON → `CheckWarn` naming the cause; the doctor run's exit status is unchanged.

**Gate-correction obligation.** Per `acceptance.md`, this check is not complete
until it has been run against three constructed failing inputs — an entry
removed, an entry added, and a corrupt file — with each observed output recorded.
Also run it **twice** against the same drifting copy: identical output on the
second run is what proves it did not repair.

Files: `internal/cli/doctor.go` (+ new `_test.go`), possibly one small helper
file in `internal/template`.

---

### M4 — move the MX index build to its consumer, delete the hook-side scan

**Decision taken (from d3 §6):** option **C** with the **B** cleanup it implies.
Option A (run the scan synchronously in SessionStart) is rejected: it would take
cold-session SessionStart from ~49 ms to ~400 ms, paid ~150 times in this repo's
worktree workflow, for a convenience. Option F (raise the time budget) is
rejected as a no-op — the 2 s box is never reached.

**M4a — consumer-side auto-build.** In `internal/cli/mx_query.go`, replace the
`os.Stat` + `SidecarUnavailable` early return at lines 97-103 with: detect
absent / empty / corrupt / stale, build the index in-process (the same
`mx.NewScanner()` + `SetIgnorePatterns(mx.DefaultScanIgnore)` + `ScanDir` +
`mgr.Write` sequence `runMXColdStartScan` performs), then continue to the query.
The build is synchronous and in the caller's own process, so nothing races a
process exit. On a build failure, keep the existing loud error path — do **not**
degrade to an empty result set, which is the mutant AC-HWD-012 is written to
catch.

**M4b — remove the hook-side scan.** From `internal/hook/session_start.go`:
delete `runMXColdStartScan` (~line 1536), `mxIndexNeedsRebuild` (~1499), the
gate call at ~228, the `mxScanNeeded` parameter threaded through
`spawnDeferredAdvisoryScans` (~554-560, taking its arity **5 → 4**), and the
inline test-branch call at ~272. The `mxIndexScanTimeout` /
`mxIndexFreshnessThreshold` constants move to wherever the freshness decision now
lives (M4a) rather than being deleted outright; the 7-day threshold's semantics
are explicitly out of scope.

**The removal must be observed behaviourally, not lexically** (audit D1 / mutant
M-1). Deleting the identifiers is not the same as removing the scan: relocating
them to a sibling file in the same package and renaming them passes any
single-file grep while the scan still fires. AC-HWD-013 therefore binds three
things — package-scoped absence, the `spawnDeferredAdvisoryScans` arity, and a
behavioural assertion — and M4b must satisfy all three.

The five existing MX tests are **inverted or retired in the same change, never
skipped**:

| Test | Disposition |
|---|---|
| `TestSessionStartMXColdStartIntegration` (`session_start_mx_test.go:159`) | **Invert.** It currently asserts the index IS created and passes — measured: `MX index built via cold-start scan … tags=1`, `--- PASS`. After M4b it asserts the index is **not** created. This inverted test is AC-HWD-013 (c) |
| `TestSessionStartMXFreshIndexNoRebuild` (`:184`) | Retire — its subject (no churn on a fresh index) ceases to exist on this path |
| `TestRunMXColdStartScan_*` (`:83`, `:121`, `:135`) | Move with the logic to M4a's consumer-side build, keeping the fail-open and timeout coverage, or retire if M4a's build path is covered by AC-HWD-012's three inputs |
| `TestMXIndexNeedsRebuild_*` (`:26`, `:35`, `:53`, `:71`) | Move with the freshness decision to M4a — absent / fresh / stale / corrupt are exactly AC-HWD-012's cases |

**M4c — correct the false comment.** `session_start.go:251-254` asserts *"The
goroutine continues to completion in the background (durable side effects still
land)"*, which is false for a CLI that exits on return and contradicts the
accurate comment at 1531-1533. Replace it with what actually happens: the
advisory keys for this session are dropped, and the goroutine is torn down at
process exit. This matters beyond this SPEC — the comment as written teaches the
next author that fire-and-forget goroutines in hook processes are safe, which is
the same belief behind the unfixed twin at `internal/hook/file_changed.go:110-118`
(§G-4).

Files: `internal/cli/mx_query.go`, `internal/hook/session_start.go`, their tests.

---

### M3 — record the per-script dispositions

**Target surface:** `internal/template/templates/.claude/rules/moai/development/hook-independence.md`
**first**, then `make build`, then mirror to
`.claude/rules/moai/development/hook-independence.md`. The two files are
byte-identical today; that property is AC-HWD-015.

**Content.** Replace the existing "Other intentionally-dormant surfaces" block —
which lumps everything under "dormant" — with a table carrying one row per
present-but-unwired script and one of five disposition classes:

| Class | Members |
|---|---|
| `reachable-via-template-settings` | `chain-event.sh` (wired in the shipped template; absent from this project until M1) |
| `reachable-via-agent-frontmatter` | `handle-agent-hook.sh` (10 registrations in agent frontmatter `hooks:` blocks — a settings-only view cannot see it) |
| `reachable-via-in-binary-registry` | `handle-session-start-compact.sh` (production firing rides `handle-session-start.sh`; the wrapper is the manual/isolated surface, deliberately unwired to avoid a double fire) |
| `dead-by-decision` | `handle-elicitation.sh`, `handle-elicitation-result.sh`, `handle-notification.sh`, `handle-task-created.sh` (retired obs-only events, guard-test enforced), `handle-worktree-create.sh`, `handle-worktree-remove.sh` (deregistered after a recorded regression: the runtime treats `WorktreeCreate` as an active creator and read the observer handler's `{}` as a path) |
| `open-question` | `handle-session-start-navigator.sh`, `team-ac-verify.sh` |

Plus the two corrections (REQ-HWD-007): 33 hook entries across 20 events, with
the 34th `"type": "command"` occurrence being `statusLine` and not a hook; and
`handle-agent-hook.sh`'s frontmatter registration.

**The `team-ac-verify.sh` wording correction spans four files, not one**
(REQ-HWD-012, audit D6). The current text — "dormant … activates only under
harness `thorough` + team mode prerequisites" — reads as *registered but gated*,
and the wiring says otherwise: it has never appeared in any settings surface, so
flipping `team.enabled: true` activates nothing. Restate it as *not registered —
activation decision pending*, matching the language already used for the worktree
hooks, in **every** surface that carries it:

| File | Occurrence | Note |
|---|---|---|
| `.claude/rules/moai/development/hook-independence.md` + template twin | the dormant-surfaces block | 7 mentions of the script; the primary M3 target |
| `.claude/rules/moai/core/agent-common-protocol.md` + template twin | `:38` — *"`team-ac-verify.sh` (TaskCompleted in team mode, dormant)"* | **always-loaded** — leaving this one wrong means the wrong reading is in every session's context |
| `.claude/rules/moai/core/agent-common-protocol-reference.md` + template twin | `:291` — *"TaskCompleted in team mode (dormant — harness `thorough` + team prerequisites)"* | spec.md §G-3 cites this very line as evidence of the misleading wording |

Template-First applies to all three pairs; AC-HWD-015 asserts mirror identity for
each, and AC-HWD-018 binds the corrected reading across all **six** surfaces
(three files + three twins). Correcting one surface and leaving two — one of them
always-loaded — would leave three surfaces disagreeing, which is worse than the
single original error.

**[HARD] The edit is line-specific.** `hook-independence.md` carries five
occurrences of "dormant" and **only two are wrong**: `:87` and the `:89-93` caveat
block, which carry the registered-but-gated reading. `:96`, `:106`, and `:107` use
the word correctly about *other* surfaces — the worktree lifecycle wrappers and
the `moai hook spec-status` subcommand — and those statements are accurate. A
sweep over the word would corrupt three correct statements to fix two wrong ones.

This is a **factual correction, not a decision**: whether team mode should fire
the hook is §G-3 / card t244 and stays untouched. Say what the wiring is; do not
say what it ought to be.

**Neutrality.** The `open-question` rows name the pending decision and **not**
the card. `t242`/`t243`/`t244` appear in this SPEC's own artifacts (spec.md §G,
frontmatter `tags`, this file, progress.md) and **never in
`internal/template/templates/**`** — matching the corrected §C-4 phrasing
(audit D11 in spec.md, N6 here).

**No deletions** (REQ-HWD-008). All 11 have template twins; removing one is a
distributed act, and two of them are pinned by `retired_wrappers_test.go` and
`m002_settings_cleanup_test.go`, with an m002 migration that strips their entries
from *user* settings on upgrade.

Files: three template rule files and their three mirrors —
`hook-independence.md`, `agent-common-protocol.md`,
`agent-common-protocol-reference.md`.

---

### M1 — the settings.json parity commit

**Purely mechanical**, and the last thing to land because nothing about it is a
decision.

Edit `.claude/settings.json` (local, git-tracked; the template `.tmpl` is already
correct and is not touched):

1. Append the `chain-event.sh` entry to the `SubagentStop` hooks array, matching
   `settings.json.tmpl:198` — same wrapper `command`/`args` shape,
   `"timeout": 5`, `"type": "command"`, **no** `async` key. It goes last, after
   `handle-subagent-stop.sh` and `handle-harness-observe-subagent-stop.sh` (this
   project has hook opt-in **enabled**, so the observe entry is present and the
   template's `{{ if }}` branch is taken).
2. Replace the single unscoped synchronous `status-transition-ownership.sh`
   `PostToolUse` entry with the template's **three**, each with its `if`
   predicate (`Write(**/.moai/specs/**)`, `Edit(…)`, `MultiEdit(…)`),
   `"timeout": 5`, `"async": true`.

Then AC-HWD-003's parity test — landed red with M2 — turns green.

**Note for the commit message and for anyone reading the diff:** the
`chain-event.sh` entry is parity with the shipped template and nothing more. It
records no ledger event, because no chain node exists for a completion edge to
attach to (§G-1 / card t242).

Files: `.claude/settings.json`.

---

## §G Anti-patterns for this SPEC

- **Closing the card by editing `.claude/settings.json` and stopping.** That is
  the card as literally written, and it changes no observable behaviour — with
  no ledger to inspect, nothing would catch it.
- **Proposing `moai update` as the remedy.** It refreshes a script that already
  exists, leaves the missing registration untouched, and destroys any hand-added
  script under `.claude/hooks/moai`.
- **Making the doctor check repair the drift** because it "obviously should".
  See §C-2; it is the same class of defect this card exists to close.
- **Deleting one of the 11 "because it is dead".** All 11 have template twins,
  four are pinned by guard tests, and two carry recorded regression context.
- **Deciding §G-1/2/3 in passing.** Each belongs to another SPEC's owner;
  answering one here steals that SPEC's requirements.
- **Marking a new gate PASS without an observed failure.** t197 spent seven audit
  rounds for want of this.
- **Putting a mitigation in `Mutant:` prose instead of in the Then clause.** Prose
  mitigations are not gates, and they read as gates. Audit iteration 1 defeated
  two criteria whose stated defence lived only in prose — and one of the mutants
  renamed the very symbol that prose relied on.
- **Claiming a criterion is mutant-proof without recording the attempts.** v0.1.0
  asserted "none constructible" for AC-HWD-011 and the auditor produced two
  mutants immediately. v0.2.0 then asserted that AC-009 + AC-010 closed the
  `hook-independence.md` "dormant" case, and they did not (audit N2). An
  overclaimed note is worse than a known-weak criterion, because it stops the next
  reader from strengthening it. Three instances across two iterations — this is
  the failure mode of this document, not an incident.
- **Re-attacking only the criteria a revision rewrote.** v0.2.0 applied the
  fresh-mutant discipline to rewritten criteria only, so AC-HWD-009 went
  unattacked through two iterations — and it was the one the auditor defeated.
  An unattacked criterion is a known gap, not a passing one.
- **Widening an observable without checking it is still satisfiable.** Closing
  AC-007's false-pass hole in v0.2.0 created a false-fail blocker: `moai doctor`'s
  own first-run cache write made the criterion impossible to satisfy by any
  implementation. Every widened assertion gets one run against the real tool
  before it ships.

---

## §H Cross-references

- `spec.md` §A.2 — the five disproved card premises
- `spec.md` §C-2 — the report-never-repair decision and why option A is rejected
- `spec.md` §G — the four deferred items and their evidence
- `acceptance.md` — 16 criteria, each with its measured pre-implementation value,
  mutant analysis, and gate-correction obligation
- `.moai/reports/t216/{d1-chain-event,d2-unwired-scripts,d3-mx-cold-start}.md` —
  the option tables and per-script evidence, referenced rather than duplicated
- `.moai/reports/t216/plan-audit.md` — audit iteration 1: the four mutants
  (M-1 … M-4) and the thirteen findings that produced v0.2.0
- `.moai/reports/t216/plan-audit-iter2.md` — audit iteration 2 (PASS 0.862,
  terminal): the Tier M threshold correction, the three executed mutants
  (N1-M unsatisfiability, N2-M six-vs-four surfaces, N3-M blanket line), and the
  seven findings that produced v0.3.0
- `.moai/reports/t469/plan-summary.md` — the v0.4.0 amendment evidence record
  (card t469): the re-measured divergence, the pre-existence proof, and the
  class-scope measurement

---

## §I — v0.4.0 Completed-SPEC Amendment Plan (2026-09-03, card t469)

> This section is the plan-phase payload for an **in-place amendment of the
> completed SPEC**. The run phase applies the exact texts in §I.4 to `spec.md`
> and `acceptance.md`. No other AC of this SPEC is re-opened, re-verified, or
> re-scored; card t216's landing stands as delivered.

### §I.1 The conflict

REQ-HWD-013 ("byte-identical at closure") and REQ-HWD-014 (no forbidden content
classes in template copies) are **simultaneously unsatisfiable for any file
whose LOCAL copy carries a forbidden-class token**: the template copy is
required to strip the token (REQ-HWD-014, neutrality doctrine `.moai/docs/
template-internal-isolation-doctrine.md` §25.1, CI guard
`.github/workflows/template-neutrality-check.yaml`) and therefore cannot stay
byte-identical (REQ-HWD-013). AC-HWD-015 — "diff reports no difference for
every file M3 touches" — is the criterion where the contradiction becomes
observable, and it is FALSE as written for one of the three M3 files.

This was not introduced by later neutrality work. The same single-line strip
divergence already existed at `a239cf050` (an ancestor of HEAD), so the conflict
coexists with this SPEC's own authoring (v0.1.0, 2026-08-24) rather than
post-dating it. The two requirements were promoted together in v0.2.0 (audit
D12) without the interaction being noticed.

### §I.2 Re-measured evidence (this plan phase, this tree, `@a1d7598ac`)

Every load-bearing number below was re-measured in this plan phase, not carried
from the lane's report. Commands and verbatim outputs are in
`.moai/reports/t469/plan-summary.md`.

| # | Claim | This phase's measurement |
|---|---|---|
| E1 | Two of the three M3 files are byte-identical local↔template | `diff -q` → `hook-independence.md` IDENTICAL, `agent-common-protocol.md` IDENTICAL |
| E2 | `agent-common-protocol-reference.md` differs at exactly one line, wholly the neutrality strip | `diff` → `275c275`; local ends `remain inline there (SPEC-SYNC-PARALLEL-DOCS-001 A9).`, template ends `remain inline there.` |
| E3 | The divergence pre-exists at `a239cf050` | both blobs extracted (`git show a239cf050:<path>` → `/tmp`), diffed → the SAME `275c275` line pair; `git merge-base --is-ancestor a239cf050 HEAD` → exit 0 |
| E4 | The conflict is a class property | 47 of the local/template pairs under the five managed roots (`.claude/rules/moai`, `.claude/agents/moai`, `.claude/skills/moai`, `.claude/commands/moai`, `.claude/output-styles/moai`) differ; 24 of those 47 carry a forbidden-class token (SPEC ID / REQ token / card number / 40-hex SHA) in their diff lines |
| E5 | The amended check passes the real tree and fails its mutants | the §I.4 command → exit 0 on the three real files; mutant (i) editorial text on the token-bearing line → `MISMATCH`, exit 1; mutant (ii) template-side insert is shape-dependent — bare token on the shared line → exit 0 (absorbed by normalization; AC-HWD-016's neutrality scan observed ≥1 SPEC-ID on the mutant copy, the compensating control), prose-wrapped token line → `MISMATCH`, exit 1 (caught directly); mutant (iii) mirror-only edit → exit 1. Iteration-1's "7 SPEC-ID hits" figure was the real LOCAL file's own token count (local 7 / template 0), produced by a wholesale-copy mutant, not an absorbed insertion — corrected in the A4 payload |

### §I.3 Shape decisions

**In-place amendment, not a separate amendment SPEC — as a declared
`completed → in-progress` amendment.** The contradiction lives in THIS SPEC's
requirement pair (REQ-HWD-013 × REQ-HWD-014); routing the fix through a second
SPEC would leave the completed SPEC's requirement layer silently
self-contradictory while the fix landed elsewhere, and would add a full
plan-run-sync ceremony to a change whose run phase is two file edits plus one
command re-run.

*Precedent correction (plan-audit D1):* this plan's iteration 1 cited this
SPEC's own HISTORY rows v0.2.0/v0.3.0 as the in-place precedent — a category
error. Both were **pre-completion** audit iterations; the SPEC has been
`status: completed` since their run/sync closed, and amending a completed SPEC
is a different transition with its own mechanics. The repo's live precedent for
that transition is **SPEC-AGENT-PARALLEL-OPT-001 v0.13.0** (in-place amendment
with the self-referential `amendment_of: SPEC-AGENT-PARALLEL-OPT-001`
frontmatter field).

*Why the declaration is load-bearing:* `internal/spec/audit.go:363-364` returns
no drift for any SPEC whose frontmatter status reads `completed` — an
un-declared v0.4.0 would sit silently `completed` with nothing ever
re-certifying it, and the amendment would be invisible to the lifecycle audit.
Declaring the amendment (`completed → in-progress` + `amendment_of:` +
HISTORY `## Amendments` record per the schema's amendment row,
`.claude/rules/moai/development/spec-frontmatter-schema.md` § Status Enum note
and § Status Transition Ownership Matrix) restores normal drift detection for
the amendment's duration and re-closes the SPEC on this card's sync commit.

**Scope (a): the amendment stays scoped to the three M3 files** ("files this
SPEC edits" — REQ-HWD-013's existing scope is untouched). The 24-pair class
measurement (E4) motivates the fix but does not widen it: a strip-aware mirror
invariant across ALL managed pairs is a general neutrality-doctrine obligation
owned by the doctrine, not by a hook-drift SPEC. It is recorded in
`spec.md`'s Out of Scope section by this amendment (§I.4 A1 payload). Minimal
change, and the general form's risk profile (normalization over 47 pairs
including settings templates) is a different audit than this one.

**Machine-check design: single plain `perl` invocation** — no shell loop, no
process substitution. Both are load-bearing in this repository: worktree-guarded
sessions reject compound commands and `<( )` substitution, so a checker written
in either form is unrunnable by the very sessions most likely to judge it. The
command normalizes forbidden-class tokens (SPEC ID, REQ token, card number,
40-hex SHA — the regex classes AC-HWD-016 scans, plus the adjacent revision
token consumed by the parenthetical form) on BOTH sides, then compares. The
`make build` ordering clause of the original AC is moved to prose: the ordering
was executed at M3 delivery and is not re-observable; binary freshness is
already bound mechanically by AC-HWD-005 and AC-HWD-012, both of which run
`bin/moai`.

### §I.4 Run-phase payloads (exact texts)

**A1 — `spec.md`:** apply the amendment-declaration mechanics, then prepend the
HISTORY row and add the `## Amendments` record.

*Frontmatter — four edits (per the schema's `completed → in-progress
(amendment)` row, `.claude/rules/moai/development/spec-frontmatter-schema.md`):*

- `version: "0.3.0"` → `version: "0.4.0"`
- `updated: 2026-09-03` (no-op — already the v0.3.0 value; see §I.5 #3)
- `status: completed` → `status: in-progress` — the amendment transition.
  Commit subject per the ownership matrix:
  `feat(SPEC-HOOK-WIRING-DRIFT-001): in-place amendment AC-HWD-015 strip-aware mirror identity`
- add `amendment_of: SPEC-HOOK-WIRING-DRIFT-001` (self-referential — the same
  form SPEC-AGENT-PARALLEL-OPT-001 v0.13.0 uses)

Without the transition + declaration, `internal/spec/audit.go:363-364` would
return no drift for the still-`completed` SPEC and nothing would ever
re-certify the amendment (§I.3).

*HISTORY row (prepend to the `## HISTORY` table):*

> `| 0.4.0 | 2026-09-03 | manager-spec | Completed-SPEC amendment (card t469) — AC-HWD-015 strip-aware mirror identity, declared per the completed → in-progress amendment row (amendment_of self-reference). REQ-HWD-013's "byte-identical at closure" and REQ-HWD-014's forbidden-class ban are simultaneously unsatisfiable for any file whose LOCAL copy carries a forbidden-class token: the template copy must strip it (REQ-HWD-014) and therefore cannot stay byte-identical (REQ-HWD-013). Observed on agent-common-protocol-reference.md:275 — local ends "there (SPEC-SYNC-PARALLEL-DOCS-001 A9).", template ends "there." — and the same single-line divergence measured in both blobs at a239cf050 (ancestor of HEAD), so the conflict coexists with this SPEC's authoring rather than post-dating it. Class measurement: 47 of the managed-root local/template pairs differ and 24 of the 47 carry forbidden-class tokens in their diff lines — a class property, not a one-file accident; the amendment stays scoped to the three M3 files and the general strip-aware invariant is recorded out of scope. Fix: REQ-HWD-013 reworded to normalized identity over the four regex-addressable classes (internal date explicitly not normalized); AC-HWD-015 rewritten around ONE machine-checkable command (single plain perl invocation, exit 0/1, worktree-guard-safe) that normalizes forbidden-class tokens on both sides before comparing; four mutants constructed and executed (editorial text on the token-bearing line → exit 1; template-side bare token insert → exit 0 absorbed, closed by AC-HWD-016; template-side prose-wrapped token → exit 1 caught directly; mirror-only edit → exit 1). No retroactive re-judgment: t216's landing stands, no other AC re-opened. |`

*HISTORY `## Amendments` sub-section (insert directly under the `## HISTORY`
heading, before the version table):*

```
### Amendments

| Field | Value |
|---|---|
| prior_completed_version | 0.3.0 |
| prior_completed_sha | 4d57b3dcf |
| prior_completed_record | progress.md §E.4 sync_commit_sha (backfilled at v0.3.0 close) |
| rationale | AC-HWD-015's byte-identical mirror demand conflicts with REQ-HWD-014's neutrality stripping — line-275 divergence on agent-common-protocol-reference.md, pre-existing at a239cf050 |
| scope | spec.md (REQ-HWD-013 reword, §F Out of Scope entry, frontmatter) + acceptance.md (AC-HWD-015 rewrite); no other AC re-opened; t216's landing stands |
| re_close_path | SPEC returns to `completed` riding this card's sync commit (3-phase close convention; the transition is owned by manager-docs on the sync commit) |
```

**A2 — `spec.md` §B:** replace REQ-HWD-013's body with:

```
- **REQ-HWD-013** (Ubiquitous) — Every template-managed file this SPEC edits
  shall be edited at its `internal/template/templates/` source first and
  mirrored afterwards, and the two copies shall be identical at closure after
  normalization of the four regex-addressable forbidden-class tokens REQ-HWD-014
  requires the template copy to omit: the SPEC ID, the REQ token, the internal
  card number, and the commit SHA.

  > Amended v0.4.0 (card t469). "Byte-identical" was unsatisfiable together with
  > REQ-HWD-014 for any file whose local copy carries a forbidden-class token —
  > observed on `agent-common-protocol-reference.md:275`, pre-existing at
  > `a239cf050`. The acceptance-side check is AC-HWD-015. The normalization
  > covers FOUR of REQ-HWD-014's five classes; the internal-date class is
  > explicitly NOT normalized (no agreed regex) — a date-mandated template
  > divergence reports MISMATCH (plan.md §I.5 #2).
```

**A3 — `spec.md` §F Exclusions:** append this H3 after the existing
`### Out of Scope — MX index staleness semantics` block:

```
### Out of Scope — strip-aware mirror invariant beyond this SPEC's files

- A general strip-aware (normalize-then-compare) mirror check across ALL
  template-managed pairs (47 differing pairs measured 2026-09-03, 24 carrying
  forbidden-class tokens). This amendment normalizes only the three files
  REQ-HWD-013 scopes; the fleet-wide invariant belongs to the
  template-neutrality doctrine (`.moai/docs/template-internal-isolation-doctrine.md`
  §25) or a dedicated follow-up SPEC.
```

**A4 — `acceptance.md`:** replace the AC-HWD-015 block (heading through its
`Harness correction:` bullet) with (the payload below is wrapped in a
4-backtick fence because it contains a fenced ```bash block — plan-audit D3;
extract everything between the 4-backtick lines):

````
### AC-HWD-015 — Template-First order was followed (strip-aware mirror identity)

**Given** the commit(s) delivering M3,
**When** each file M3 touches is compared against its
`internal/template/templates/` twin with the forbidden-class tokens normalized
on BOTH sides (the stripping REQ-HWD-014 mandates for the template copy),
**Then** the normalized contents are identical for every file M3 touches —
`hook-independence.md`, `agent-common-protocol.md`, and
`agent-common-protocol-reference.md` — judged by ONE command, run from the
project root, whose exit status separates PASS (0) from FAIL (1):

```bash
perl -e 'local $/; my $rc=0; for my $f (@ARGV){ open my $L,"<",$f or die "open $f: $!"; open my $T,"<","internal/template/templates/$f" or die "open tmpl $f: $!"; my ($a,$b)=(<$L>,<$T>); for ($a,$b){ s/\s*\((?:SPEC-[A-Z][A-Z0-9]*(?:-[A-Z0-9]+)*|REQ-[A-Z][A-Z0-9]*)-[0-9]{3}(?:\s+[A-Z][0-9]{1,2})?\)//g; s/\s*\b(?:SPEC-[A-Z][A-Z0-9]*(?:-[A-Z0-9]+)*|REQ-[A-Z][A-Z0-9]*)-[0-9]{3}\b//g; s/\s*\bt[0-9]{2,3}\b//g; s/\s*\b[0-9a-f]{40}\b//g; } if ($a ne $b){ print "MISMATCH $f\n"; $rc=1; } } exit $rc' \
  .claude/rules/moai/development/hook-independence.md \
  .claude/rules/moai/core/agent-common-protocol.md \
  .claude/rules/moai/core/agent-common-protocol-reference.md
```

The original criterion's `make build` ordering clause was executed at M3
delivery and is not re-observable after the fact; binary freshness stays bound
mechanically by AC-HWD-005 and AC-HWD-012, both of which run `bin/moai` — stale
unless the build ran.

- `Pre-impl observed (v0.4.0):` `@a1d7598ac` — raw `diff -q` local vs template:
  `hook-independence.md` → IDENTICAL; `agent-common-protocol.md` → IDENTICAL;
  `agent-common-protocol-reference.md` → DIFFERS at exactly one line (275),
  wholly the neutrality strip `(SPEC-SYNC-PARALLEL-DOCS-001 A9)`. The unamended
  criterion is FALSE for the third file as written and cannot be made true
  without violating REQ-HWD-014. The command above → exit 0 (PASS) on all three
  real files. Raw commands and verbatim outputs:
  `.moai/reports/t469/plan-summary.md`.
- `Mutant:` **four constructed and executed.** (i) *Non-token editorial text
  inserted on the token-bearing line* — `MISMATCH`, exit 1: normalization
  absorbs only forbidden-class tokens and their parenthetical wrapper, never
  neighboring prose. (ii) *Forbidden token inserted into the template copy* —
  shape-dependent, measured both ways: (ii-bare) a bare ` (SPEC-AUDIT-FAKE-001
  A9)` appended on the otherwise-shared line → exit 0, absorbed by the
  normalized comparison, which is direction-agnostic on token runs — in this
  shape the compensating control is AC-HWD-016, and the neutrality scan
  observed a non-zero SPEC-ID count on the mutant template copy, so the pair of
  criteria closes the absorbed shape; (ii-prose) the same token wrapped in its
  own prose line (`Cross-reference: see (SPEC-AUDIT-FAKE-001 A9) for detail.`)
  → `MISMATCH`, exit 1 — the mirror check catches it directly, no
  compensating control needed. (iii) *Edit only the local mirror* —
  `MISMATCH`, exit 1. Record correction: the "7 SPEC-ID hits" figure in the
  iteration-1 plan record was the real LOCAL file's own token count (local 7 /
  template 0), produced by a wholesale-copy mutant — the template copy made
  byte-identical to the local file — not by an absorbed insertion; it is
  corrected here so the transplanted record carries the actual construction.
  No claim that the space is exhausted.
- `Harness correction:` [HARD] construct at least mutant (i) and observe the
  exit 1 before trusting the command's passing output — a normalized comparison
  that has only ever been seen passing has not been shown to detect anything.
````

**A5 — re-run the §I.4 command and `moai spec lint`** after the edits, and
record both in `progress.md` §E.1 as the amendment's plan→run hand-off signal:

- the §I.4 A4 command → exit 0 (PASS) on the amended tree;
- `go run ./cmd/moai spec lint .moai/specs/SPEC-HOOK-WIRING-DRIFT-001/spec.md`
  → 0 errors. `Pre-impl observed:` plan-audit iteration 1, MP-3 (tree-built
  lint at `f0b926d71`) → **0 error(s), 18 warning(s)** — all pre-existing
  `CoverageIncomplete` from the linter not resolving the spec.md↔acceptance.md
  cross-file REQ↔AC mapping; unchanged by this plan and expected to persist at
  the same count after the amendment.

The run phase commits `spec.md` + `acceptance.md` by explicit pathspec.

### §I.5 Residual risks for the plan-audit

> Updated by the iteration-2 patch (2026-09-03): D1-D5 of plan-audit iteration
> 1 (`.moai/reports/t469/plan-audit.md`, PASS-WITH-DEBT 0.80) applied to §I.3,
> §I.4 A1-A5, and the §I.2 E5 row. The confirming check after this patch is
> textual only — no re-audit is required per the coordinator's disposition.

1. **Normalization over-absorption on token-bearing lines.** The check deletes
   parentheticals whose identifiable content is a forbidden token. A local-side
   parenthetical and a template-side parenthetical could differ in ways the
   normalizer cannot see (mutant (ii-bare)'s absorption, generalized — the
   prose-wrapped shape is caught directly). The compensating control is
   AC-HWD-016 on the template side; the local side is unguarded in the same way
   the unamended AC was. Judged acceptable at this scope (three files, one
   known token-bearing line); the fleet-wide SPEC, if ever opened, should
   tighten this.
2. **Internal-date class not normalized.** AC-HWD-016's forbidden classes
   include "internal date", which has no agreed regex; the normalization covers
   the four regex classes only — now stated in A2's reworded REQ text as well.
   If a future strip removes a date, the mirror check will report MISMATCH for
   a divergence that is neutrality-mandated — the same false-FAIL shape as the
   defect being amended, one class over.
3. **`updated` field collision.** The frontmatter already carries
   `updated: 2026-09-03` from v0.3.0; A1's bump is a no-op for that field. If
   the SPEC lifecycle lint expects `updated` to move per amendment, there is
   nothing to move — verify, don't assume.
