# Progress — SPEC-GLM-CLEANUP-SSOT-001

> Plan-phase skeleton. §E.2–§E.3 are populated by manager-develop (run-phase) and §E.4 by
> manager-docs (sync-phase); this agent emits only §E.1.

## §E.1 Plan-phase Audit-Ready Signal

- **Plan-phase artifacts emitted**: spec.md, plan.md, acceptance.md, progress.md (this file).
- **Tier**: M (standard set). **Era**: V3R6. **Class**: C (design change).
- **Status**: `draft` — set at creation; this agent owns the `(none) → draft` transition only.
- **Card**: t888. **Origin**: card t802 R3+R4. **Base**: `0cca34439`, worktree
  `.claude/worktrees/t888`, branch `WT-glm-cleanup-ssot`.
- **Pre-write self-checks executed**:
  - SPEC ID regex `^SPEC(-[A-Z][A-Z0-9]*)+-[0-9]{3}$` against `SPEC-GLM-CLEANUP-SSOT-001` → `PASS`
    (run as Bash in this tree).
  - Dedup: no `.moai/specs/SPEC-GLM-CLEANUP-SSOT-001/` existed before creation.
  - Frontmatter: all 12 canonical fields present in spec.md, plus optional `era`, `tier`,
    `related_specs`. `phase` carries a release target, not a lifecycle token.
  - Sibling artifacts carry no `status:` field (artifact statelessness).
- **Evidence basis**: all measurements in spec.md §A were taken by the lane in this tree at
  `0cca34439`; the three function locations, the `residue_probe` files, and `config.GLMEnvVarSet`
  were independently confirmed present by this agent before authoring.
- **Superseded premise recorded**: the card's "`CLAUDE_CODE_MAX_CONTEXT_TOKENS` is cleared nowhere"
  is closed in this tree by commit `03d1904a7` (card t802) — recorded in spec.md §A.2 as
  superseded, not as an open defect.
- **Lead decisions encoded, not re-opened**: option (가) conservative indicator widening (REQ-5);
  the `ANTHROPIC_BASE_URL` early return is design, not a defect (REQ-6); the producer is untouched
  and the marker-key approach is a separate card (REQ-11, spec.md §B); the mutation demonstration
  is an acceptance criterion rather than advice (REQ-9 / AC-008); the deliberately-red probe carries
  its rationale in prose (REQ-10 / AC-006).
- **Accepted non-closure declared**: `TestProbeStrandedResidueSurvivesIndicatorLoss` remains FAIL
  after this SPEC. Recorded in spec.md §C REQ-10, acceptance.md AC-006 and §D.3.
- **Open clarifications**: none. No `[NEEDS CLARIFICATION]` marker is outstanding.

### Plan-audit iteration 2 — rework applied (v0.2.0, 2026-09-19)

Iteration 1 (`.moai/reports/t888/plan-audit.md`) returned **FAIL 0.69** against the Tier M
threshold 0.80. All seven must-pass criteria passed; the FAIL was located entirely in the
**verification layer**. The problem statement (§A set relations A=14 / B=13 / C=9, C ⊂ A,
B = A − FABLE; four probes RED) was reproduced by the auditor with zero mismatches and is
unchanged in v0.2.0.

Eight blocking findings, all closed in this amendment:

| Finding | Closure |
|---|---|
| D1 — AC-002's grep cannot see its own defect | AC-002 rewritten: runtime equivalence test per consumer is primary; textual scan demoted to secondary with a hits→few transition control (plan §E D7) |
| D2 — M5 collides with the tmux-parity guard | One declaration, two derived views (REQ-1/REQ-8); cleanup functions take the cleanup view, the tmux guard keeps the live view unmodified; pinned by AC-013 (plan §E D3) |
| D3 — `MOAI_STATUSLINE_CONTEXT_SIZE` reopens the regression REQ-5 forbids | Indicator revised to option (가') = exactly `{ANTHROPIC_BASE_URL, MOAI_BACKUP_AUTH_TOKEN}`; the key excluded with both code citations; AC-007 + AC-014 |
| D4 — three `git diff` instruments pass vacuously after commit | All range-scoped to `"$CARD_BASE"..HEAD` with a non-empty-range control; AC-001's existence check moved from diff to file grep (acceptance §D.0) |
| D5 — AC-007's new test lacked `scrubGatewayEnv(t)` | Required in plan M2 and checked by AC-007 control 1 |
| D6 — the literal-ban clause rested on one dead instrument | Split across AC-002 (runtime) and AC-016 (structural), mapped individually in §D.1 |
| D7 — REQ-7 ordering stated only in M4 | REQ-7 rebound to all three functions with the mechanism of violation named; restated in plan M2/M4/M5; AC-010 runs both sides |
| D8 — AC-004 asserted what its command did not measure | Second command added: `TestStripGLMCredsClearsEveryLiveKey`, which asserts `TeammateMode` |

Minor findings (operator discretion): **D9** closed — §A.2's quoted command now carries the scrub
and `-timeout 30m`. **D10** closed — §D.1 now reads "one or more". **D11** closed — AC-006 uses
content search instead of the fixed `sed -n '1,60p'` window. **D12** closed — REQ-9 re-subjected to
the run-phase evidence. **D13** closed — REQ-8 now requires both helper doc comments to be updated,
checked by AC-013.

Additionally, the plan-audit's Gap 3 is closed as a *claim*: plan §C.3 no longer asserts the default
suite is green. It is now an unverified pre-flight step the run phase must execute and record, with
the consequence of skipping it stated (AC-012 becomes unresolvable for a red package).

**Not re-opened.** The four lead decisions (option 가/가', negative control fixed, deliberate red
probe, producer untouched) remain settled. The indicator revision from three keys to two is itself
a lead decision of 2026-09-19 and is encoded, not debated.

**New in v0.2.0**: REQ-12 (repair the `CLAUDE_CODE_TEAMMATE_DISPLAY` stand-in fixture — a GLM-owned
key was standing in for "a user's own key"), plan milestone M3, and AC-015. Milestones renumbered
M1-M9; REQ-1..REQ-12 and AC-001..AC-016 are both within the Tier M ceiling of 16.

## §E.2 Run-phase Evidence

### Run-phase entry record (2026-09-19)

- **Kickoff**: Implementation Kickoff Approval granted by the operator via the lead, 2026-09-19.
- **Position**: worktree `.claude/worktrees/t888`, branch `WT-glm-cleanup-ssot`, HEAD `0cca34439`
  (re-read at entry: `git rev-parse --short HEAD` → `0cca34439`, `git branch --show-current` →
  `WT-glm-cleanup-ssot`, `git rev-parse --show-toplevel` → the worktree path above).

### [HARD] Audit state at entry — read this before treating the SPEC as audited

**No third full plan-audit was run, and no 0.80-threshold verdict exists for v0.3.0.**

The Tier M plan-audit iteration ceiling is 2 (`harness.yaml:77`) and both iterations were spent:

| Iteration | Version audited | Verdict | Report |
|---|---|---|---|
| 1 | v0.1.0 | **FAIL 0.69** (threshold 0.80) | `.moai/reports/t888/plan-audit.md` |
| 2 | v0.2.0 | **FAIL 0.76** (threshold 0.80), no score regression | `.moai/reports/t888/plan-audit-2.md` |
| — | **v0.3.0 (current)** | **no verdict exists** | — |

v0.3.0 closed the five blocking findings iteration 2 raised (N1-N5). Those five were verified
**individually by the lane**, at the lead's direction, in place of a third audit. That is a
targeted check of five named fixes — it is **not** a comprehensive verdict on the document, and it
produced no score. A later reader must not cite it as "the SPEC passed audit".

What the five verifications do establish, and by what evidence:

| Finding | How verified | Evidence |
|---|---|---|
| N1 — AC-002's fixture forced a token-destroying implementation | **Executed**, both directions, against today's `removeGLMEnv` | `.moai/state/verify/t888/n1-ac002-shape.txt` — as-written FAIL (`ANTHROPIC_AUTH_TOKEN="user-oauth-token"` survives), as-amended PASS, `EXIT:1` |
| N5 — §C's positive control was necessarily 0 at dispatch | **Executed** the redefined control | `CARD_BASE=$(git merge-base develop HEAD) && git diff --name-only "$CARD_BASE~1"..HEAD \| wc -l` → `1` |
| N2 — REQ-2/3/4 × REQ-7 destroyed the restored token | Clause read in file | spec.md:300 carve-out; bound at :189/:196/:206; restated in plan M2/M4/M5 |
| N3 — `CARD_BASE` pinning contradicted a [HARD] repo rule | Clause read in file | 10 inline re-resolutions in acceptance.md; `gitflow-lane-protocol.md` §8 cited |
| N4 — AC-015's control rejected M3's seeding instruction | Measured table compared to the real fixture | plan.md:333-335 vs `internal/hook/session_end_test.go` cases at :331/:349/:367 |

**Attribution correction carried forward.** `moai spec lint SPEC-GLM-CLEANUP-SSOT-001` →
`✓ No findings`, `EXIT:0`, with a discriminating control (a non-existent ID on the same binary
exits 3 and names the path it tried). That result covers **`spec.md` only** — `plan.md` and
`acceptance.md` are not lint subjects, and most of the N1/N3/N4/N5 edits landed in those two files.
The lint says the document is well-formed; it says nothing about whether N1-N5 are correctly closed.

### Gaps declared at entry (not observed)

- No third full audit; no threshold verdict for v0.3.0 (above).
- `moai spec lint`'s subject set is `spec.md` alone.
- Plan-audit optional finding N9 (progress.md §E.1 stale-option line) left unaddressed —
  operator discretion.
- The default-suite baseline (§C step 3) is being measured at entry; until it lands, nothing here
  claims any package is green.

### §C Pre-flight

1. **Position** — confirmed above.
2. **Probe RED baseline** — all four probes FAIL at this HEAD, measured by the lane before any
   edit: `.moai/state/verify/t888/baseline-probe-hook.txt` and `baseline-probe-cli.txt`, both
   `EXIT:1`. The 13 existing hook cleanup guards, including the negative control
   `TestCleanupGLMSettingsLocalLeavesNonGLMFileAlone`, all PASS:
   `baseline-guards-hook.txt`, `EXIT:0`.
3. **Default-suite baseline** — split into three per-package measurements after the first attempt
   was aborted. Status: `config` done, `hook` deferred, `cli` deferred to the integration window.

   **First attempt, ABORTED — recorded as not-a-baseline.** A single run over all three packages
   was started under resource slot `default-suite-baseline` (EXIT:0 on acquire) and stopped before
   `internal/cli` completed. Machine load had gone `8.93 → 164.15 → 198.39` (1-min), driven by two
   other lanes running full `internal/cli` suites concurrently; the lead identified one of them
   (`t951`) as an out-of-scope full run and had it stopped. The packages that DID complete before
   the abort are present in `.moai/state/verify/t888/baseline-default-suite.txt` but are explicitly
   **not cited as a baseline** — a green measured under that contention describes the machine, not
   the code. The slot was released rather than held idle.

   **Gate doctrine, revised by the lead mid-flight — the first form was unusable.** The initial
   rule ("start only while the 1-min load average is < 40") is unreachable by construction in this
   repository: one `internal/cli` package run parks the machine in the 90s on its own, because that
   package's tests spawn subprocesses. The lane's wait on that condition was stopped. The revised
   discriminant is the count of **concurrent full-package runs** — targeted `-run` executions carry
   no gate at all. Separately, the lead established that the load figure alone does not say what is
   saturated: `top -l 2` at 22:26 read `Load Avg: 154.58` with `48.28% idle`, i.e. I/O congestion
   rather than CPU saturation. Load figures are therefore **recorded for attribution, never used as
   a gate**, and are reported with idle% beside them.

   **`internal/config` — measured.** `go test -count=1 ./internal/config/...` → all three packages
   `ok`, `EXIT:0`. Evidence `.moai/state/verify/t888/baseline-config.txt`, which carries the
   attribution header (command, exit code, uptime triple, and the concurrent-full-run list: one,
   the integration-window holder's). One deviation is recorded in that file rather than hidden: the
   `uptime` read immediately before the run was `32.43` but the `sysctl` sample at run start read
   `44.24`, straddling the then-current gate. The result is retained because contention produces
   false RED, never false GREEN — a pass under load is still a pass; a red would have been
   discarded and re-measured.

   **`internal/hook` — deferred by lane request, lead-approved.** Running it now would make the
   count of concurrent full-package runs two, and the other one is the integration-window holder's
   verdict-bearing run. `internal/hook/perf` (102s), `quality` (69s) and `security` (35s) are
   timing-sensitive and spawn subprocesses, so this lane's run could push that lane's suite red.
   The asymmetry is the reason: a contaminated measurement of one's own is recoverable and visible,
   whereas contaminating another lane's verdict is silent — that lane reads the red as a statement
   about its code without ever learning this run existed. Deferred until the window holder's full
   run ends (bounded: its parent carries `timeout 1500`).

   **`internal/cli` — deferred to the integration window.** [HARD, lead] For the duration of this
   batch the right to run the full `internal/cli` suite belongs to the lane holding the integration
   window. This baseline is measured inside that window when it is granted.

### Commit-ordering deviation — the plan-phase commit was deferred, and is recorded rather than hidden

The canonical ordering commits the plan-phase artifacts as a `(none) → draft` commit during the
plan phase, and the first run-phase commit then carries `draft → in-progress`. This card did not
follow that ordering, and the reason is worth stating so the ordering is not read as an accident:
the artifacts went through two plan-audit iterations (FAIL 0.69, FAIL 0.76) and three amendments
(v0.1.0 → v0.2.0 → v0.3.0), and the lane deliberately held **zero commits** for the whole of that
period, so the branch would carry no version of the SPEC that an audit had already rejected.

The consequence is that milestone M1 (`10f00a125`) landed while the artifacts were still untracked
— the M1 agent surfaced this rather than improvising, and was right to. The artifacts are committed
immediately after, carrying `status: in-progress`, which is the true state: the run phase has
started. They are deliberately NOT committed as `draft` first — the branch never held a draft-state
version of them, and manufacturing one would be a fabricated history rather than a recovered one.

### M1 — canonical settings-axis declaration (`10f00a125`)

- **Files**: `internal/config/envkeys.go` (+115), `internal/config/settings_axis_test.go` (new, +274).
  Two constants added for keys previously spelled only as string literals
  (`EnvMoaiBackupAuthToken`, `EnvAPITimeoutMs`).
- **Shape**: one ordered declaration `settingsAxis` whose entries carry a `legacy` flag; both views
  derive from it by filtering, so the live view is not a second literal. Each accessor returns a
  fresh slice, mechanising the "callers MUST NOT mutate" contract that `GLMEnvVarSet` states in
  prose only.
- **Cardinalities** (independently re-verified by the lane): cleanup **14**, live **9**, legacy tail
  **5**; 9 + 5 = 14.

**Lane verification, independent of the agent's report:**

| Check | Command | Observed |
|---|---|---|
| config suite | `unset … && go test -count=1 ./internal/config/...` | 3/3 `ok`, `EXIT:0` |
| AC-001 doc statements | `grep -ci` × 3 on `envkeys.go` | `1 / 1 / 1` |
| AC-001 control (absent→present) | same three patterns at `HEAD~1` | `0 / 0 / 0`, base file 607 lines (instrument alive) |
| `GLMEnvVarSet` untouched | `git diff HEAD~1 HEAD -- internal/config/envkeys.go \| grep -cE '^[+-].*func GLMEnvVarSet'` | `0`, with 117 `+/-` lines as the positive control |

**AC-001 instrument corrected during verification.** The criterion's three doc-statement checks were
written case-sensitively (`grep -c 'settings axis'`) and returned `0/0/0` against a doc comment that
states all three in uppercase for emphasis. The content requirement was met; the instrument was not
able to see it. The greps are now `grep -ci`. This is an instrument fix, not a relaxation — the
requirement is that the three statements be present, never that they be lowercase.

**Two directions of RED the M1 agent supplied**, and the distinction between them matters:
compile-failure RED (the accessors undefined) proves the symbols were absent but cannot show the
membership assertions bite; so a second RED was produced by mutating one declaration entry's
`legacy` flag, which turned `TestSettingsAxisViewsShareOneDeclaration` and
`TestSettingsAxisLiveViewMembership` red with cardinality diagnostics. The mutation was reverted and
green re-measured. This is NOT the AC-008 mutation demonstration (that one diverges a *consumer* and
belongs to M7); it is the narrower claim that M1's own test is not vacuous.

**Open decision the agent correctly declined to take on itself.** Its instruction scoped it to
`internal/config`, and the status transition plus artifact commit sit outside that scope. It stopped
and asked rather than widening its own mandate. The orchestrator resolved it as recorded above.

_<milestone evidence follows>_

## §E.3 Run-phase Audit-Ready Signal

_<pending run-phase>_

## §E.4 Sync-phase Audit-Ready Signal

_<pending sync-phase>_
