# SPEC-GUARD-LIVENESS-001 — Implementation Plan (card t333)

Baseline tree for every measurement in this plan: `d34a789a4` @ `WT-guard-liveness` (worktree `.claude/worktrees/t333`), equal to `origin/develop` at authoring time.

Milestones are ordered by **decision reversibility** — the manifest schema and the classification vocabulary are the decisions that cost the most to change later, so they lead. The mechanical work (evaluator plumbing, doctrine text) sits at the bottom.

## §A Context

See `spec.md` §A. Five empirical instances ground the requirements; two of them (§A.5 `spec-status-auto-sync` all-`skipped`, §A.6 the 18-vs-11 census gap) were measured during this plan phase and are what force the `fired-with-effect` vocabulary and the `UNDECLARED` classification respectively.

## §B Known issues and constraints carried in

- **B-1 — the self-application constraint.** A watcher that cannot prove its own firing is forbidden. Answered by making the evaluator pull-based rather than scheduled (`spec.md` §D.1). The answer relocates the regress; it does not eliminate it, and §D.3 says so.
- **B-1b — the unprompted-discoverability constraint.** A targeted query can only be issued by someone who already suspects the answer, so a design answering only when queried relocates the problem into whoever is expected to already know (`spec.md` §A.4, §D.2). Independent of B-1: a design can pass one and fail the other, and failing either leaves the same defect one level up.
- **B-2 — `gh run list` is not a complete history.** It reports what the forge retained. Measured: the default 100-run listing spans about three hours on this repository because high-frequency workflows saturate it. Any evaluator built on it inherits the limit; REQ-GDL-007 makes the limit an explicit `UNKNOWN` state rather than a silent "never fired".
- **B-3 — do not conflate "a defect occurred" with "a defect survived to adoption."** The t241 lane mis-signed its own prediction ledger by predicting "0 audit findings" as success, when a rule working *well* at the audit layer drives that number **up**. Every expectation this SPEC writes names its measured quantity (REQ-GDL-003) so the same inversion cannot be written here.
- **B-4 — `moai update` deletes `.moai/config/` wholesale.** `CleanMoaiManagedPaths` removes the root before redeploying templates, so a manifest placed there is deleted on every update. REQ-GDL-001 puts the manifest outside that root.
- **B-5 — the manifest is repository-specific, not template content.** It describes this repository's `.github/workflows/`. It must not be mirrored into `internal/template/templates/`; doing so would ship this repository's CI census to every downstream project.

## §C Pre-flight (measured on `d34a789a4`)

| Check | Command | Result |
|---|---|---|
| Manifest absent | `ls .moai/guards` | `No such file or directory` (rc=1) |
| No `moai guard liveness` verb | `grep -n '"guard"' internal/cli/*.go` | only `internal/cli/constitution.go:49` (`moai constitution guard`) — an unrelated verb |
| Rule carries no continued-firing clause | `grep -nE "last.fired\|continued.firing\|stopped firing\|liveness\|stale guard" .claude/rules/moai/development/verification-completeness.md` | rc=1, no match |
| Workflow census | `ls -1 .github/workflows/` | 18 files |
| Recent-run census | `gh run list --limit 100 --json workflowName ... \| group_by` | 11 distinct workflow names |

## §D Constraints

- Read-only against the forge (REQ-GDL-012). No issue creation, no dispatch, no commit.
- No scheduled workflow is added by this SPEC (REQ-GDL-011).
- No existing text in `verification-completeness.md` is modified (REQ-GDL-015).
- No file under `internal/template/templates/` is touched (B-5).

## §E Self-verification

Every acceptance criterion in `acceptance.md` carries a RED-now cell pinned to `d34a789a4` with its stated reason, and a green-path cell naming the flipping milestone, per `.claude/rules/moai/development/verification-completeness.md` §2. Each criterion was additionally subjected to the mutant probe from that rule's §2; where a single-clause criterion admitted a mutant, a second clause was added and the mutant is recorded in the criterion.

## §F Milestones

### M1 — the manifest schema and the classification vocabulary (least reversible)

The data-model decision, and the one worth the most review attention: every later milestone reads this shape, and changing it after the census is populated means rewriting all 18 entries.

- Define the manifest schema: per-entry `workflow`, `expect_events`, `window`, `measures` (REQ-GDL-002), plus the release-cycle-conditional form (REQ-GDL-005).
- Fix the `measures` vocabulary at exactly three values with the conclusion-set each admits (REQ-GDL-003).
- Fix the classification vocabulary the evaluator emits: `OK`, `STALE`, `UNKNOWN`, `UNDECLARED`.
- Choose the manifest path outside `.moai/config/` (REQ-GDL-001, B-4).
- Populate the census: one entry per workflow file, 18 of 18.

Flips: AC-GDL-001, AC-GDL-002, AC-GDL-003, AC-GDL-004.

### M2 — the evaluator

- Per-workflow query, one query per manifest entry; the repository-global listing is never an evidence source (REQ-GDL-006).
- Empty per-workflow result classifies `UNKNOWN` (REQ-GDL-007); last qualifying run older than the window classifies `STALE` (REQ-GDL-008).
- `fired-with-effect` rejects `skipped` and `cancelled`; `verdict-rendered` additionally requires `success` or `failure` (REQ-GDL-003).
- Workflow file present on disk with no manifest entry classifies `UNDECLARED` (REQ-GDL-004).
- Result carries its measurement timestamp and the four coverage counts (REQ-GDL-009).
- All-clear is refused when the successfully-queried count is zero (REQ-GDL-010).
- Read-only: no forge mutation of any kind (REQ-GDL-012).

Flips: AC-GDL-005, AC-GDL-006, AC-GDL-007, AC-GDL-008, AC-GDL-009.

### M3 — the attended surface

The second-least-reversible decision after M1, because the surfacing contract is what constraint 2 (`spec.md` §D.2) is satisfied or failed by, and a CLI-verb-shaped answer here cannot be repaired later without redoing the surface.

- Invoke the evaluator from an already-attended surface; add no scheduled workflow (REQ-GDL-011).
- Render `STALE` and `UNDECLARED` entries as a non-blocking advisory that arrives **with no operator-supplied guard identifier or query** (REQ-GDL-013). A documented `moai guard liveness` verb is not an acceptable sole answer — it is still a question someone must know to ask.
- Lead with entries whose classification changed since the previous rendered result; carry unchanged `STALE`/`UNDECLARED` entries as a compact standing count (REQ-GDL-016). This requires persisting the previous result to diff against.
- The advisory carries the age of the measurement it reports (REQ-GDL-014).

Flips: AC-GDL-010, AC-GDL-011, AC-GDL-013, AC-GDL-014.

### M4 — doctrine (most mechanical)

- Add the continued-firing clause to `verification-completeness.md` as new text only. Verify by diff that no existing line changed (REQ-GDL-015).

Flips: AC-GDL-012.

## §G Anti-patterns to avoid

- **Adding a scheduled watcher workflow "just as a backstop".** It is subject to the defect it watches for, the forge disables it after repository inactivity, and it starts the regress the design exists to avoid. Rejected in `spec.md` §D, not merely unimplemented.
- **Reporting a clean sweep from a global run listing.** Measured to be incapable of answering the question for a low-frequency guard (§A.4, §A.6).
- **Omitting a release-only guard from the manifest because it is "supposed to be quiet".** Omission is what makes silence unreadable; declare the condition instead (REQ-GDL-005).
- **Counting `skipped` runs as firings.** §A.5 is a live instance of exactly this misread.
- **Shipping a CLI verb as the whole answer to (c).** A verb the operator must know to run reproduces §A.4 exactly: the lead session could run the right query the instant it was handed the workflow's name, and could not know a query was owed. Rejected in `spec.md` §D.2 and excluded by AC-GDL-013's negative clause.
- **Reprinting the full standing list every session.** It is how a new advisory inherits the filter an always-red neighbour has already trained (`spec.md` §A.7). REQ-GDL-016 leads with changes instead.
- **Mirroring the manifest into the template tree.** B-5.

## §H Cross-references

- `.moai/reports/t333/trigger-axis-observation.md` — this SPEC's primary evidence artifact.
- `.claude/rules/moai/development/verification-completeness.md` — the landed rule this SPEC extends additively (`spec.md` §B).
- `SPEC-BINARY-LAG-VISIBILITY-001` (card t326, in flight) — the state axis, out of scope here.
- `.moai/specs/SPEC-INTEGRATION-LOCK-LIVENESS-001/` (card t298) — instance 1's subject.
