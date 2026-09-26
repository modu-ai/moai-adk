# SPEC-AUTONOMY-PRECONDITION-001 — Progress

Card t1245, track A2b. Worktree `.claude/worktrees/t1245`, branch `WT-push-serialize-sign`,
base `develop` at `553e224f3`.

## §E.1 Plan-phase Audit-Ready Signal

- Artifacts: `spec.md`, `plan.md`, `acceptance.md`, `design.md`, `progress.md`.
- **Tier M**, derived against the actual Tier predicate. The predicate is the S/M/L table in
  `.claude/rules/moai/workflow/spec-workflow.md` § SPEC Complexity Tier — the table at `:136-140`
  (LOC band, files affected, artifact set, PASS threshold), Tier M being its `:139` row — plus the
  REQ/AC budget table at `:144-147`. Against it:
  - **files affected** — estimated 8-12 (two guards + tests in `internal/hook`, the env constants in
    `internal/config`, the projection + test in `internal/contract` or a sibling, the new rule file and
    its template mirror). Tier M's band is 5-15.
  - **artifact set** — 5 files against Tier M's specified 3. `design.md` is justified by two genuine
    design decisions (design.md §A) and `progress.md` is emitted at every Tier, so it is not counted in
    the Tier total.
  - **REQ/AC budget** — 11 live requirements and 15 criteria, against Tier M's independent ceilings of
    16 and 16. Within budget on both axes; the criteria axis has one slot of headroom, which is worth
    knowing before another criterion is added.
  - **LOC band** — **not measured.** Implementation does not exist, so the 300-1000 LOC column has no
    value to compare against; Tier M rests on the files, artifact-set and budget axes only.
  - **PASS threshold: 0.80**, read from the Tier M row `spec-workflow.md:139`.
  - v0.1.1 cited "the Tier L threshold of ≥3 milestones **and** ≥10 files" as its rationale. That is
    the `manager-lead` **routing** predicate from `CLAUDE.md` §4, not the Tier predicate — and
    `plan.md` carries four milestones, so the cited predicate's own first conjunct was satisfied. The
    citation is corrected here; the Tier itself is unchanged.
- **Moving-ref census, re-run before this revision and again after it, both at `eee5f635e`.** The SPEC
  directory was swept with the pattern on the next line:
  <!-- moving-ref-ok: the two branch names are this census's own search terms and the subject of its finding, not coordinates anything is read from -->
  `grep -rnE 'WT-escalation-detector|WT-contract-schema' .moai/specs/SPEC-AUTONOMY-PRECONDITION-001/`.
  Before: 9 hits, 5 of them evidence coordinates citing a relative ref that had slid past the split.
  After: **0 evidence coordinates**, and every remaining hit is
  subject-class with a non-empty `<!-- moving-ref-ok: … -->` marker on the flagged line or the line
  immediately above it, per
  `.claude/rules/moai/core/verification-claim-integrity-detail.md` § The exemption marker:

  | Location | Class | Why the ref stays |
  |---|---|---|
  | `spec.md:32`, `design.md:8` | SUBJECT / S1 | the ref IS the recorded miscitation; a SHA would erase which citation was wrong |
  | `progress.md:32` | SUBJECT / S1 | this census's own search pattern |
  | `spec.md:71` | SUBJECT / S1 | narrates the citation rule v0.1.1 applied |
  | `spec.md:657` | SUBJECT / S2 → R4 | asserts what the sibling's branch *currently* maps; written command-first with a dated reference value |
  | `spec.md:723` | SUBJECT / S1 | the sentence states the citation rule, so substituting a SHA would make it violate itself |
  | `spec.md:727` | SUBJECT / S1 | an address saying which branch the sibling card lives on |

  The iter-2 predicate is therefore asserted on **both** conditions, not one: (a) zero moving refs
  among evidence coordinates, and (b) every remaining moving ref carries a non-empty exemption marker.
- **The lead-supplied "current mapping" was already stale when it arrived, which is the census's own
  best evidence.** The dispatch gave `AC-AE-022`→`REQ-AE-019,020` / `AC-AE-023`→`REQ-AE-022` /
  `AC-AE-024`→`REQ-AE-023`. Measured at the sibling's current HEAD `78a95ee03` (v0.4.2):
  `AC-AE-022`→`REQ-AE-018,019` / `AC-AE-023`→`REQ-AE-019,020` / `AC-AE-024`→`REQ-AE-022` /
  `AC-AE-025`→`REQ-AE-023` — one row further along again, because two commits landed on that branch
  after the dispatch was written (`0c647d01d`, `78a95ee03`). §H.1 therefore records the command
  before the value and dates the value.
- **v0.1.2 revisions** (iter-1 audit delta + three lead rulings): citations re-pinned to `d8926ff9a`
  (transferred text) and `25283ebf8` (A1 v0.5.2, superseding the v0.5.1 `65e0a9167` the dispatch named
  — v0.5.2 is the revision that corrected the mission-projection and sign-deny attributions to A2b,
  the statements this SPEC cites, and the v0.5.1→v0.5.2 diff shifts later lines non-uniformly — +1 at
  the § Out of Scope heading, +3 at `:329` / `:410` / `:462`, and `:231` added with no v0.5.1
  counterpart, 520 → 524 lines — so no constant offset rescues a stale pin);
  the outright deny split into three rules (REQ-AP-003 boundary-keyed human path; REQ-AP-011 role-keyed
  non-interactive path and `decide`; a no-marker session allowed to `decide`); REQ-AP-012 defining the
  marker constants so the role gate is evaluable without card t1240; REQ-AP-010 narrowed to a rule file
  this SPEC creates (closing the audit's D1); REQ-AP-008 specified against A1's drafted mapping with
  three measured corrections; §H rebuilt from `d8926ff9a`. Requirement count 9 → 11 live, criteria
  12 → 15.
- **AC count measured with the canonical counter, before and after.** Command:
  `AC_FILE=.moai/specs/SPEC-AUTONOMY-PRECONDITION-001/acceptance.md` piped through the counter body in
  `.claude/agents/moai/manager-docs.md` (`# MOAI-AC-COUNTER-BEGIN`..`END`). Before: `live=12
  excluded=3 ambiguous=0`, exit 0. After: `live=15 excluded=3 ambiguous=0`, exit 0. One intermediate
  run returned `AMBIGUOUS AC-AP-012` exit 3 — a newly written sentence mentioned the retired id without
  its marker — and was resolved by marking that occurrence, which is the halt behaving as designed.
  The `[RETIRED]` / `[REF]` tokens remain **inside** the code span; a closing backtick before the token
  breaks adjacency and the counter stops registering it.
- This SPEC is absent from the AC-snapshot corpus at `.moai/reports/t338/ac-count-baseline.txt`
  (`grep -n 'AUTONOMY-PRECONDITION'` → no output), so the commit-time guard's `ABSENT` path applies:
  `scripts/ac-baseline/check-staged.sh:250-254` reports it as "unrecorded (report only)" and never
  fails. No snapshot regeneration is required by this revision.
- SPEC ID regex check executed as Bash against the `internal/spec/lint.go` pattern: `PASS`.
- Requirements in GEARS notation; every live requirement traces to ≥1 criterion and every criterion
  maps one requirement, with a per-criterion **ownership column** naming the surface each Then reads
  and who creates it (acceptance.md §G). The ownership column replaces v0.1.1's blanket
  "no criterion depends on unowned work" sentence, which the audit showed was false for `AC-AP-013`.
- Every live criterion carries a named `Discriminator:` line stating what makes it RED when the guard
  is absent or fails open. `AC-AP-002` and `AC-AP-006` — the two of the audit's three absence-only
  criteria that are still live (`AC-AP-012` is retired) — gained armed positive controls inside the
  same test.
- Exclusions section present with five `### Out of Scope — …` H3 sub-headings, each carrying `-`
  bullets (`grep -c '^### Out of Scope — ' spec.md` → 5), satisfying the `OutOfScopeRule` convention.
- ID-reuse and ID-retirement records written in A1's own format (spec.md §H.1-H.2); no retired id
  re-used, and the three new ids continue the numbering rather than filling the 011-012 gap.
- Premises measured against `553e224f3` and re-measured at `eee5f635e`; the premises that did not hold
  are recorded as absent rather than assumed (spec.md §C.2, §C.3, §C.6, §C.7, §C.10).
- Open items: O3 (audit sink) remains open; O1, O2, O4 and **O5** are resolved. O5 closed by operator
  ruling — the `MOAI_FACTORY_ROLE` value is **`worker`**, the canonical CLI role spelling, not the
  retired alias `agent`; card t1240 must stamp `worker`. The open item does not block a criterion's
  evaluability.
- Dependency picture after v0.1.2: **no milestone depends on another track.** M2 and M3 depend on
  nothing outside this SPEC; M1 and M4 consume A1's `show --json` as a **fixture**, which is the
  correct evaluation basis per spec.md §E C4, not a wait.

### Gaps — what this plan phase did NOT measure

Stated explicitly so a later reader does not read the section above as exhaustive.

- **LOC.** The Tier predicate's first column is unmeasured; there is no implementation to measure.
  If run-phase finds the files-affected count above 15, the Tier is re-judged.
- **`AC-AP-010`'s mutant.** Not built, not run. First observed at M2.
- **`AC-AP-014`'s "not a re-implementation" and exported-surface baseline.** Neither measured; the
  baseline is captured at M4's start.
- **Whether every harness populates the hook process's environment identically to the session's.**
  Not measured. It bounds what REQ-AP-011's role read can be relied on to see (spec.md §C.7).
- **`sudo`'s effect on the guard.** Reasoned from where the guard runs (PreToolUse, before the command
  executes), not measured end-to-end.
- **The sibling SPECs' `status:` fields.** `SPEC-AUTONOMY-CONTRACT-001` and
  `SPEC-AUTONOMY-ESCALATION-001` are not present in this worktree; they were read through committed
  refs only, so whether either is retired or superseded is **unverified** rather than verified absent.
- **Go build, tests and lint.** Not run — plan phase adds no code.

### F8 disposition — 2026-09-26 (plan-audit iter-2)

F8's premise expired positively: the acceptance.md §H DoD gate command names `./internal/contract/...`,
which did not exist at plan time but exists in this tree since the A1 absorb (merge `e0a471d1e`).
The existing hedge in acceptance.md ("or the package the projection lands in") stands; no text
change is required, and package confirmation is deferred to run milestone M4.

## §E.2 Run-phase Evidence

### M1 — Push serializer (REQ-AP-001, REQ-AP-002, REQ-AP-007; design.md §B)

Implemented in `internal/hook/push_serializer.go` (+ `push_serializer_test.go`,
`push_serializer_units_test.go`; call-site wiring in `internal/hook/pre_tool.go`
PreToolUse and `internal/hook/post_tool.go` PostToolUse). Commit
`6e831c5a0` (this tree, branch `WT-push-serialize-sign`, parent `a98139343`).
Measured 2026-09-26 by the M1 run lane.

- **E8 — RED evidence (verbatim, captured BEFORE any implementation existed):**

  ```
  $ go test ./internal/hook/ -run 'TestPushSerial|TestPushLease'
  # github.com/modu-ai/moai-adk/internal/hook [github.com/modu-ai/moai-adk/internal/hook.test]
  internal/hook/push_serializer_test.go:62:35: undefined: PushShowJSON
  internal/hook/push_serializer_test.go:65:15: undefined: DecodePushShowJSON
  internal/hook/push_serializer_test.go:100:17: undefined: checkPushSerializer
  internal/hook/push_serializer_test.go:109:16: undefined: checkPushSerializer
  internal/hook/push_serializer_test.go:143:17: undefined: DecodePushShowJSON
  internal/hook/push_serializer_test.go:149:19: undefined: checkPushSerializer
  internal/hook/push_serializer_test.go:153:18: undefined: checkPushSerializer
  internal/hook/push_serializer_test.go:189:20: undefined: checkPushSerializer
  internal/hook/push_serializer_test.go:198:30: too many arguments in call to mustJSON
  FAIL	github.com/modu-ai/moai-adk/internal/hook [build failed]
  ```

  The named AC tests were written first; the API they name did not exist.

- **E1 — AC matrix** (command: `go test ./internal/hook/ -run
  'TestPushSerial|TestPushLease|...' -v`, this run, this tree):

  | AC | Test | Result |
  |----|------|--------|
  | AC-AP-001 | `TestPushSerializationDeniesSecondPush` | `--- PASS: TestPushSerializationDeniesSecondPush (0.43s)` |
  | AC-AP-002 | `TestPushSerializationInactiveWhenNotArmed` (5 conditions incl. armed (e)) | `--- PASS: TestPushSerializationInactiveWhenNotArmed (2.13s)` |
  | AC-AP-003 | `TestPushLeaseReleasedOnFailedPush` | `--- PASS: TestPushLeaseReleasedOnFailedPush (0.48s)` |
  | AC-AP-004 | `TestPushLeaseReclaimAndFailOpen` (3 limbs) | `--- PASS: TestPushLeaseReclaimAndFailOpen (1.25s)` |

  Full run: `ok github.com/modu-ai/moai-adk/internal/hook 8.369s` — 13 test
  functions, 4 AC tests + 9 unit tables, swept count non-empty (`--- PASS:` lines above).

- **E2 — builds:** `go build ./...` → exit 0; `GOOS=windows GOARCH=amd64 go
  build ./...` → exit 0 (both re-measured after wiring; `WINDOWS_OK` / `NATIVE_OK`).
- **E3 — coverage:** `go test -coverprofile` over the AC + unit runs →
  push_serializer.go per-function: 9 functions at 100%,
  `releasePushLeaseOnFailure` 94.7% (only the defensive release-error
  advisory limb uncovered). New-file aggregate is above the 85% gate. Whole
  `internal/hook` package coverage under the FULL suite: measured separately
  (see the run log line in the completion report).
- **E4 — subagent boundary:** `grep -rn 'AskUserQuestion'
  internal/hook/push_serializer*.go` → 0 rows. pre_tool.go/post_tool.go
  insertions (+35 lines) add none; pre-existing rows in those files are the
  AskUserQuestion observer, untouched.
- **E5 — lint:** `golangci-lint run --timeout=2m` after the change →
  `0 issues.` — identical to the pre-change baseline (`0 issues.` measured in
  pre-flight). No new issue.
- **E6 — commits:** `6e831c5a0` (M1 code + tests + spec.md draft→in-progress).
  No push (lane mode: push is the lead's).
- **E7 — blockers:** none. One design-note below.

**Design note — production activation seam (design.md §E record).** The
activation triple is read from the `moai contract show --json` document via
`DecodePushShowJSON`; which contract governs a session is the contract
resolver of SPEC-AUTONOMY-ESCALATION-001 REQ-AE-002 (card t1235), not yet in
this tree. The PreToolUse/PostToolUse call sites therefore resolve activation
through the `pushShowJSONLoader` seam, whose nil default keeps the serializer
inactive — the same inert posture as the opt-in guards beside it: on the
inactive path no record is read and no audit line is written. All four ACs are
fixture-based and fully exercisable without the resolver (acceptance.md §G:
"the fixture stands in for A1's command"). The AC-AP-002 conditions (b)/(c)
(action present with the field false, field absent) are likewise
fixture-expressible because the real derivation folds the field into the
action's presence (`internal/contract/derived.go` `fillDerived`).

## §E.3 Run-phase Audit-Ready Signal

_<pending run-phase>_

## §F Phase 4 Mode Selection

Recorded by the lane orchestrator (worker-62) before the first run-phase spawn, 2026-09-26.

Input parameters: tier M · scope ~12 files across internal/cli, internal/hook (+tests), a new
`.claude/rules` file + template mirror, internal/mission projection, docs · domains 4 (Go cli/hook
guard, rules+template, mission projection, docs) · language mix Go + markdown + shell · concurrency
benefit LOW (coding-heavy; milestone dependencies M1→M2→M3, M4 reads A1's landed surface) ·
agent-team prereqs: not requested.

| Mode | Selected | Rationale |
|------|----------|-----------|
| direct | no | multi-file Go implementation, far beyond typo/single-line |
| serial | **yes** | coding-heavy work per Anthropic's coding-task parallelism caveat; single-writer-per-tree bars parallel write spawns here; M2's guard tests build on M1's lease record, M3 documents M2's deny |
| fanout | no | write collision risk in one tree (one writer per working tree); not research |
| sweep | no | new-code semantic work, not a ≥30-file single uniform mechanical transform |

Decision: serial — one manager-develop spawn per milestone (M1→M4), orchestrator verifies evidence
between spawns.

Justification: the milestones are sequentially dependent and touch shared packages
(internal/hook guard tests assert against M1's serializer lease record), so parallel write spawns
would collide on one tree. Serial also matches the single-spawn guidance for coding tasks. Kickoff
cleared via operator policy 2026-09-26 (recorded in .moai/reports/t1245/verdict.md §12); run-phase
autonomy armed as `ac_converge` with the SPEC-scoped DoD gate (not `go test ./...`, per repo-local
load discipline).

## §E.4 Sync-phase Audit-Ready Signal

_<pending sync-phase>_
