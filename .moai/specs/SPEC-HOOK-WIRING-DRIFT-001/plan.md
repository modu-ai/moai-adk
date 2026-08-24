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

**One ordering dependency crosses this:** AC-HWD-003's parity test cannot be
green until M1 lands. Land the test **red** with M2 (it is the same
render-and-compare machinery the diagnostic needs), then M1 turns it green. Do
not weaken the test to make it pass early.

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
`spawnDeferredAdvisoryScans` (~560-573), and the inline test-branch call at
~272. Retire — do not skip — any test asserting the scan's presence. The
`mxIndexScanTimeout` / `mxIndexFreshnessThreshold` constants move to wherever the
freshness decision now lives (M4a) rather than being deleted outright; the 7-day
threshold's semantics are explicitly out of scope.

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

Also correct the existing `team-ac-verify.sh` wording. The current text —
"dormant … activates only under harness `thorough` + team mode prerequisites" —
reads as *registered but gated*, and the wiring says otherwise: it has never
appeared in any settings surface, so flipping `team.enabled: true` activates
nothing. Restate it as *not registered — activation decision pending*, matching
the language already used for the worktree hooks.

**Neutrality.** The `open-question` rows name the pending decision and **not**
the card. `t242`/`t243`/`t244` appear only in `spec.md` §G.

**No deletions** (REQ-HWD-008). All 11 have template twins; removing one is a
distributed act, and two of them are pinned by `retired_wrappers_test.go` and
`m002_settings_cleanup_test.go`, with an m002 migration that strips their entries
from *user* settings on upgrade.

Files: the template rule file, its mirror.

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

---

## §H Cross-references

- `spec.md` §A.2 — the five disproved card premises
- `spec.md` §C-2 — the report-never-repair decision and why option A is rejected
- `spec.md` §G — the four deferred items and their evidence
- `acceptance.md` — 16 criteria, each with its measured pre-implementation value,
  mutant analysis, and gate-correction obligation
- `.moai/reports/t216/{d1-chain-event,d2-unwired-scripts,d3-mx-cold-start}.md` —
  the option tables and per-script evidence, referenced rather than duplicated
