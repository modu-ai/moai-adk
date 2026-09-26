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

### M2 — Contract-sign and contract-decide guard (REQ-AP-003, REQ-AP-004, REQ-AP-005, REQ-AP-009 .. REQ-AP-013; design.md §C)

Implemented in `internal/hook/contract_sign_guard.go` (+
`contract_sign_guard_test.go`, `contract_sign_guard_units_test.go`;
call-site wiring added to `internal/hook/pre_tool.go` PreToolUse after the
push-serializer block; one exclusion row added to
`internal/hook/hmp_source_guard_test.go` `hmpGoLiteralExclusions`). Constants
in `internal/config/envkeys.go` (`EnvFactoryRole`, `FactoryRoleWorker`,
REQ-AP-012); REQ-AP-013 pins in `internal/cli/factory_role_pin_test.go` and
`internal/kanban/factory_label_pin_test.go`. Measured 2026-09-26 by the M2
run lane, branch `WT-push-serialize-sign` (M2 commits follow `7fa7c9bb1`).

Milestone-internal order held: the REQ-AP-012 constants landed before the
role-scoped rule and its tests read them (plan.md §F M2); the value is
`worker` (operator-confirmed), and the retired alias `agent` is rejected by
construction (equality against the constant) and asserted by the AC-AP-018
alias limbs.

- **E8 — RED evidence (verbatim, captured BEFORE any implementation existed):**

  ```
  $ go test ./internal/config/ -run TestFactoryRoleEnvConstant
  internal/config/envkeys_factory_role_test.go:27:5: undefined: EnvFactoryRole
  internal/config/envkeys_factory_role_test.go:28:44: undefined: EnvFactoryRole
  internal/config/envkeys_factory_role_test.go:30:5: undefined: FactoryRoleWorker
  internal/config/envkeys_factory_role_test.go:32:4: undefined: FactoryRoleWorker
  FAIL    github.com/modu-ai/moai-adk/internal/config [build failed]
  $ go test ./internal/cli/ -run TestFactoryRoleTokenPinsGuardConstant
  internal/cli/factory_role_pin_test.go:24:38: undefined: config.FactoryRoleWorker
  ... (4 rows) ...
  FAIL    github.com/modu-ai/moai-adk/internal/cli [build failed]
  $ go test ./internal/kanban/ -run TestFactoryLabelPrefixPinsGuardConstant
  internal/kanban/factory_label_pin_test.go:22:22: undefined: config.FactoryRoleWorker
  ... (4 rows) ...
  FAIL    github.com/modu-ai/moai-adk/internal/kanban [build failed]
  $ go test ./internal/hook/ -run 'TestContractSign|TestContractRole'
  internal/hook/contract_sign_guard_test.go:74:23: undefined: checkContractSign
  internal/hook/contract_sign_guard_test.go:80:73: undefined: contractSignViolationPrefix
  ... (12+ rows) ...
  FAIL    github.com/modu-ai/moai-adk/internal/hook [build failed]
  ```

  The named AC tests were written first; the API and constants they name did
  not exist. (AC-AP-018's pin assertions pass at birth by design — their
  RED-now was the plan-phase absence of the assertion itself; the
  discriminator is the mutation probe below.)

- **E1 — AC matrix** (commands: `go test ./internal/hook/ -run
  'TestContractSign|TestContractRole' -v`; `go test ./internal/config/ -run
  TestFactoryRoleEnvConstant`; `go test ./internal/cli/ -run
  TestFactoryRoleTokenPinsGuardConstant`; `go test ./internal/kanban/ -run
  TestFactoryLabelPrefixPinsGuardConstant` — this run, this tree):

  | AC | Test | Result |
  |----|------|--------|
  | AC-AP-005 | `TestContractSignAgentInvocationDenied` (11 shapes, count limb) | `--- PASS: TestContractSignAgentInvocationDenied (0.00s)` |
  | AC-AP-006 | `TestContractSignPositiveControlsAllowed` (5 controls + armed deny) | `--- PASS: TestContractSignPositiveControlsAllowed (0.00s)` |
  | AC-AP-007 | `TestContractSignUnclassifiedDeniedClosed` (4 shapes + `unclassified` token) | `--- PASS: TestContractSignUnclassifiedDeniedClosed (0.00s)` |
  | AC-AP-008 | `TestContractSignDeniedBeforeVerbExists` (stub precondition FIRST, then sign + decide denies) | `--- PASS: TestContractSignDeniedBeforeVerbExists (0.24s)` |
  | AC-AP-015 | `TestContractRoleScopedDenyUnderWorkerMarker` (6 shapes, eval `unclassified`) | `--- PASS: TestContractRoleScopedDenyUnderWorkerMarker (0.00s)` |
  | AC-AP-016 | `TestContractRoleScopedAllowWithoutWorkerMarker` (2 runs × 6 allows + 2 armed human-path denies) | `--- PASS: TestContractRoleScopedAllowWithoutWorkerMarker (0.00s)` |
  | AC-AP-017 | `TestFactoryRoleEnvConstant` (constants + literal-free hook + env-read closure) | `ok github.com/modu-ai/moai-adk/internal/config` |
  | AC-AP-018 | `TestFactoryRoleTokenPinsGuardConstant` / `TestFactoryLabelPrefixPinsGuardConstant` (each asserts equality with its carrier AND inequality with every retired alias) | `ok internal/cli` · `ok internal/kanban` |
  | AC-AP-009 | `TestContractSignWrapperBypassesDenied` (8 wrappers, resolved-program limb) | `--- PASS: TestContractSignWrapperBypassesDenied (0.00s)` |
  | AC-AP-010 | `TestContractSignUnknownWrapperFailsClosed` (chrt → unclassified deny) | `--- PASS: TestContractSignUnknownWrapperFailsClosed (0.00s)` |

- **AC-AP-010 mutation limb — OBSERVED (run-phase requirement, acceptance.md
  §C):** the unknown-wrapper branch of `classifyContractWords` was removed
  (mutant applied to the working tree), and the flip was observed:

  ```
  --- PASS: TestContractSignWrapperBypassesDenied (0.00s)
      contract_sign_guard_test.go:323: chrt 0 moai contract sign: decision = "" reason = "", want the deny sentinel
  --- FAIL: TestContractSignUnknownWrapperFailsClosed (0.00s)
  ```

  The chrt case flipped denied → allowed while the closed-list wrapper cases
  stayed denied (AC-AP-009 green) — the criterion is not satisfied by the
  closed list alone. Original restored; full suite re-run green after
  restoration.

- **E2 — builds:** `go build ./...` → exit 0 (`BUILD_OK`);
  `GOOS=windows GOARCH=amd64 go build ./...` → exit 0 (`BUILD_OK_WINDOWS`) —
  re-measured after all M2 edits.
- **E3 — coverage:** `internal/hook` FULL suite →
  `ok github.com/modu-ai/moai-adk/internal/hook 261.692s coverage: 85.8% of
  statements` (≥85%). The new guard file per-function
  (`-coverprofile` over the guard + units tests): every function ≥85.2%,
  18-function average 93.8% (`checkContractSign` 100%,
  `classifyContractWords` 85.2%, wrapper skippers 87.5-90%). `internal/config`
  package coverage 82.8% — PRE-EXISTING baseline, not moved by this change:
  the M2 diff to `envkeys.go` adds two constants (zero executable
  statements), and the pin/package tests added no production code. Recorded
  as an attributed Gap against the 85% package gate, not a PASS.
- **E4 — subagent boundary:** `grep -rn 'AskUserQuestion|/mcp__askuser'
  internal/hook/contract_sign_guard.go contract_sign_guard_test.go
  contract_sign_guard_units_test.go internal/config/envkeys.go
  internal/cli/factory_role_pin_test.go internal/kanban/factory_label_pin_test.go`
  → 0 rows. The pre-existing `pre_tool.go` `AskUserQuestion` observer rows
  (HEAD lines 27/32/713) are untouched by the +21-line wiring insertion.
- **E5 — lint:** `golangci-lint run --timeout=2m` → `0 issues.` — identical
  to the pre-change baseline measured in pre-flight. One test-time defect was
  caught and fixed during the milestone: the full `internal/hook` suite
  initially FAILED `TestHMPSourceGuardGoLiterals` (the guard's `bash` is a
  shell program name inside parsed command text, not a `tool_name` branch);
  resolved through the HMP guard's own `hmpGoLiteralExclusions` extension
  point with a stated reason — no production change, no exclusion weakening
  beyond the one designed row.
- **E6 — commits:** two M2 commits follow `7fa7c9bb1` on
  `WT-push-serialize-sign` (constants+pins; guard+tests+wiring), plus the
  §E.2 evidence commit carrying this section. No push (lane mode: push is the
  lead's).
- **E7 — blockers:** none. Notes below.

**Design notes (recorded per design.md §E).**

1. **`script -c CMD` is a `-c` payload, not a skipped option.** design.md
   §C.3 lists `-c CMD` among script's own options, but skipping it outright
   would silently DROP the command string (a false allow — the wrong
   direction for §C). The guard parses CMD as a one-level `-c` payload, the
   fail-closed reading of the same row; `script -q /dev/null moai contract
   sign` (the AC-AP-009 fixture) follows the operand path as designed.
2. **Unknown `--signer` values fail closed.** design.md §C.2 step 6 is
   exhaustive over `human` / `llm` / `llm+jev`; a signer value outside A1's
   closed decider set cannot be shown to be off the agent paths, so it is
   denied with the unclassified mark (design.md §C.6 direction).
3. **The wired guard is ACTIVE in production for the human path.** Unlike M1
   (whose activation waits on the contract resolver), the human-path deny has
   no external precondition: it denies from the moment this commit is
   present. The role-gated half remains nominal until card t1240 stamps
   `MOAI_FACTORY_ROLE` — spec.md §F O5 / §E C7 carry that residual; nothing
   here compensates for it.
4. **`internal/cli` suite timeout (load, not defect).** The first
   `go test ./internal/cli/` run panicked with
   `test timed out after 10m0s` and ZERO individual test failures — the
   package's known contention profile (gitflow lane protocol §8). Re-run
   under a slot lease with `-timeout 22m` → `ok ... 1015.058s`, zero
   failures. The verdict surface remains CI.

### M3 — Documentation of the mode-independent deny (REQ-AP-010; design.md §C.6a, §C.7)

Files created: `.claude/rules/moai/workflow/contract-sign-guard.md`,
`internal/template/templates/.claude/rules/moai/workflow/contract-sign-guard.md`
(byte-identical apart from the front-matter `paths:` value), and the AC test
`internal/template/contract_sign_guard_rule_test.go`. Run on lane
`WT-push-serialize-sign`, TDD cycle_type, HEAD baseline `81f351c6f`.

- **E1 — AC-AP-013 matrix row:**

  | AC | Status | Verification Command | Actual Output |
  |----|--------|---------------------|---------------|
  | AC-AP-013 | PASS | `go test ./internal/template/ -run TestContractSignGuard -count=1 -v` | `--- PASS: TestContractSignGuardRuleDocumentsModeIndependence (0.00s)` / `ok  github.com/modu-ai/moai-adk/internal/template  0.401s` |

  The test reads both files from disk and asserts: (a) the required
  mode-independence statements — "mode-independent", "guided", "escalation
  detector", "nothing changes under", the sentinel
  `CONTRACT_SIGN_AGENT_VIOLATION:`, and the `MOAI_FACTORY_ROLE` marker — are
  present in BOTH trees; (b) parity: byte-identical after stripping the
  front-matter `paths:` line; (c) neutrality: no SPEC id, card id, internal
  date, or commit SHA in either file. All four AC limbs are covered by
  mechanically observed assertions, not reads.

- **E2 — builds:** `make build` on the unmodified baseline tree → exit 0
  (pre-flight); post-mirror `make build` → `BUILD_EXIT=0` with catalog
  regenerated unchanged; `go build ./...` → `GO_BUILD_OK`;
  `GOOS=windows GOARCH=amd64 go build ./...` → `GOOS_WINDOWS_OK`. All
  measured on this tree this run.

- **E3 — coverage: n/a.** The milestone adds zero executable statements —
  two markdown documentation files and one test file. There is no production
  Go surface to cover; stated explicitly rather than omitted.

- **E4 — boundary grep on the new files:**
  `grep -nE 'Audit [0-9]|Finding A[0-9]|spec\.md §|plan\.md §|design\.md §|acceptance\.md §|goos|backups'`
  over both new rule files → no output (clean). Neutrality greps:
  `grep -c 'SPEC-'` both files → `0`; `grep -cE 'REQ-[A-Z]|AC-[A-Z]|t[0-9]{3,4}|[0-9]{4}-[0-9]{2}-[0-9]{2}'`
  both files → `0`; hex-run scan (`\b[0-9a-f]{7,40}\b`) → empty.
  §25.3 five-item checklist (C1 SPEC-id / C2 REQ-AC token / C3 audit
  citation / C4 date + short-sha / C5 memory-archive path) passed manually
  against the staged diff; the automated backstop
  (`TestTemplateNoInternalContentLeak`, `TestRuleProvenance*`) ran green
  inside the package battery below.

- **E5 — lint:** `golangci-lint run --timeout=2m` → `0 issues.` — identical
  to the pre-change baseline measured in pre-flight on the unmodified tree
  at `81f351c6f`.

- **E6 — commits:** one M3 commit follows `81f351c6f` on
  `WT-push-serialize-sign` (`7d766dcd1` — rule file + mirror + AC test),
  plus the §E.2 evidence commit carrying this section. No push (lane mode:
  push is the lead's).

- **E7 — blockers:** none.

- **E8 — verbatim RED output (captured BEFORE the rule files existed):**

  ```
  $ go test ./internal/template/ -run TestContractSignGuard -count=1 -v
  === RUN   TestContractSignGuardRuleDocumentsModeIndependence
      contract_sign_guard_rule_test.go:65: CONTRACT_SIGN_GUARD_RULE_DRIFT: local rule file unreadable /…/.claude/rules/moai/workflow/contract-sign-guard.md: open /…/.claude/rules/moai/workflow/contract-sign-guard.md: no such file or directory
  --- FAIL: TestContractSignGuardRuleDocumentsModeIndependence (0.00s)
  FAIL
  FAIL	github.com/modu-ai/moai-adk/internal/template	0.708s
  ```

  RED is red for the right stated reason: neither the rule file nor its
  mirror exists on the pre-implementation tree — exactly the RED-now state
  acceptance.md records ("neither the rule file nor its mirror exists").

**Design notes.**

1. **The pair is deliberately NOT enrolled in the byte-parity allowlist** of
   `rule_template_mirror_test.go`: that test requires full byte identity,
   while AC-AP-013 requires the two trees to differ in the front-matter
   `paths:` value. The parity invariant for this pair lives in the AC test
   itself (byte-identical apart from the paths line), which is the AC's own
   definition.
2. **The required statements are asserted as neutral prose strings** — no
   SPEC id or REQ token appears in either documented file, satisfying the
   neutrality limb and the §25 CI guard simultaneously. The AC's limb 2
   (state that the companion promise is scoped in words to the escalation
   detector) is carried by the phrases "escalation detector" and "nothing
   changes under", without citing the tracking tokens.
3. **`paths:` scoping mirrors the guard's code surface**: the rule loads
   when a session touches `contract_sign_guard.go`, its tests, or the rule
   itself — a guard's documentation does not earn always-loaded budget.

### M4 — mission-validator projection (REQ-AP-008 / AC-AP-014)

Landing package: `internal/contract` (A1's own package, design.md §E's first
choice — the AC's default location, so no rename to record). Files:
`internal/contract/projection_mission.go`,
`internal/contract/projection_mission_test.go`,
`internal/contract/testdata/mission_surface_baseline.txt`.

**Exported-surface baseline (AC-AP-014 limb 5 — captured FIRST, at M4
milestone start, before any edit; tree `d0d915f3f`):**

```
$ go doc -all ./internal/mission > /tmp/t1245-mission-baseline.txt   # exit 0, 529 lines
$ shasum -a 256 /tmp/t1245-mission-baseline.txt
24d4fdf334e9f0b8fabc5923e4602ccde138a212d0122aa55c4251c8283464c1  /tmp/t1245-mission-baseline.txt
```

The byte-identical copy is committed as the limb-5 golden at
`internal/contract/testdata/mission_surface_baseline.txt` (same sha256,
re-measured after the change: `24d4fdf3…`), and the AC test compares a fresh
`go doc -all ./internal/mission` against it on every run. Post-change
re-diff: `diff baseline after` → empty, "mission surface IDENTICAL".

- **E1 — AC-AP-014 matrix:**

  | AC | Status | Verification Command | Actual Output |
  |----|--------|---------------------|---------------|
  | AC-AP-014 (limbs 1-5, five subtests) | PASS --- 5/5 limbs | `go test ./internal/contract/ -run TestContractProjectsOntoMissionValidator -count=1 -v` | `--- PASS: TestContractProjectsOntoMissionValidator (0.05s)` + all five limb subtests PASS (limb 2's five refusal cases and limb 5's surface baseline included) |
  | AC-AP-014 limb 1 | PASS | same | `limb_1_signed-valid_contract_is_not_refused_incomplete_contract … PASS` — `Approved=true`, `MaxOperations=40>0`, `mission.SealMissionContract` accepts (no `incomplete_contract`) |
  | AC-AP-014 limb 2 | PASS | same | baseline decision returns `ReceiptPrepared`; the five single-field mutants return exactly `mission decision: {mission_not_running, mission_mismatch, policy_mismatch, stale_snapshot, expired_decision}` — the verdicts are mission's own refusal vocabulary, not a local re-implementation |
  | AC-AP-014 limb 3 | PASS | same | `internal/foo/x.go` accepted; `internal/foobar/x.go` refused `mission decision: scope_expansion` — `/**`→prefix neither narrows nor widens |
  | AC-AP-014 limb 4 | PASS | same | `internal/*/x.go` → error naming `ownership.write` and the glob; injected unmapped key `future_field` → named by `unmappedProjectionKeys`; `card` on `DeliberatelyNotProjected`, fixture carries `card: t1245`, projection succeeds |
  | AC-AP-014 limb 5 | PASS | `go doc -all ./internal/mission` vs committed golden (test subtest + independent `diff`) | byte-identical; golden sha256 `24d4fdf334e9f0b8fabc5923e4602ccde138a212d0122aa55c4251c8283464c1` unchanged |

  Limb-4 note (recorded per the AC's "unmapped field" wording): Go structs are
  closed, so a contract field that is neither projected nor enumerated cannot
  be constructed at the struct level; the projection derives its field
  inventory from the live struct via reflection
  (`contractInventoryKeys`) and fails closed on any inventory key absent
  from the projected set and `DeliberatelyNotProjected`. The
  naming-the-field behavior is therefore asserted at the key-list level
  (`unmappedProjectionKeys` with `future_field` appended), which is the only
  constructible form of that input; a schema amendment adding a field makes
  every projection fail closed loudly until the mapping tables are updated.

- **E2 — cross-platform build:**

  ```
  $ go build ./...                          → host build OK (exit 0)
  $ GOOS=windows GOARCH=amd64 go build ./... → windows build OK (exit 0)
  ```

- **E3 — coverage (projection package = `internal/contract`, ≥85%):**

  ```
  $ go test -cover ./internal/contract/ ./internal/mission/ -count=1
  ok  github.com/modu-ai/moai-adk/internal/contract  0.560s  coverage: 96.3% of statements
  ok  github.com/modu-ai/moai-adk/internal/mission   5.552s  coverage: 88.1% of statements
  ```

  Scoped suite: `go test ./internal/contract/ ./internal/mission/ -count=1`
  → both `ok`. Full suite NOT run locally (lane discipline; CI on
  `origin/develop` is the verdict surface).

- **E4 — subagent boundary grep (non-test Go sources only):**

  ```
  $ grep -rn 'AskUserQuestion\|mcp__askuser' internal/contract/ --include='*.go' | grep -v "_test.go" | grep -v '^\s*//'
  (no output — exit 1, 0 matches)
  ```

  (The word `AskUserQuestion` appears only inside the limb-5 golden text
  file — `go doc` output quoting mission's `SupervisorState.AskUserQuestion`
  field — which is a testdata fixture, not code.)

- **E5 — lint:** `golangci-lint run --timeout=2m` → `0 issues.` — identical
  to the pre-change baseline measured in pre-flight on the unmodified tree
  at `d0d915f3f`.

- **E6 — commits:** one M4 commit `21506ceb0` (projection + AC test + limb-5
  golden) on `WT-push-serialize-sign`, child of `d0d915f3f`; plus the §E.2
  evidence commit carrying this section. No push (lane mode: push is the
  lead's).

- **E7 — blockers:** none.

- **E8 — verbatim RED output (captured BEFORE the projection existed):**

  ```
  $ go test ./internal/contract/ -run TestContractProjectsOntoMissionValidator -count=1
  # github.com/modu-ai/moai-adk/internal/contract [github.com/modu-ai/moai-adk/internal/contract.test]
  internal/contract/projection_mission_test.go:65:13: undefined: ProjectToMission
  internal/contract/projection_mission_test.go:116:20: undefined: ProjectToMission
  internal/contract/projection_mission_test.go:198:13: undefined: ProjectToMission
  internal/contract/projection_mission_test.go:213:11: undefined: contractInventoryKeys
  internal/contract/projection_mission_test.go:214:18: undefined: unmappedProjectionKeys
  internal/contract/projection_mission_test.go:218:15: undefined: unmappedProjectionKeys
  internal/contract/projection_mission_test.go:228:23: undefined: DeliberatelyNotProjected
  internal/contract/projection_mission_test.go:229:76: undefined: DeliberatelyNotProjected
  internal/contract/projection_mission_test.go:231:16: undefined: ProjectToMission
  FAIL	github.com/modu-ai/moai-adk/internal/contract [build failed]
  FAIL
  ```

  RED is red for the right stated reason: the projection symbols do not
  exist on the pre-implementation tree — exactly the RED-now state
  acceptance.md records ("no projection exists"). The test was written and
  its RED captured before any implementation line was authored; the
  implementation (`projection_mission.go`) was derived afterward to satisfy
  the test.

**Design notes.**

1. **`Approved` is derived from the seal, not trusted**: the projection
   recomputes `ComputeSeal` and checks the consistency table
   (`signedValid`), so a tampered or unsigned signature fails closed at the
   projection boundary naming `signature`, before mission ever sees it.
2. **`worktree` action is dropped, not failed**: it has no mission
   counterpart and dropping it only narrows allowed actions (fail-closed
   direction); correction 2's "at least one mission-mappable action" guard
   fires only when the mapped list would be empty (error naming `actions`).
3. **Unmapped fields are detected via a reflection-derived inventory**, so a
   future schema amendment fails every projection loudly until the mapping
   tables are updated — the fail-closed semantic REQ-AP-008 asks for, given
   that Go's struct decoder cannot carry an unknown field.
4. **`mission_contract_sha256` is not produced** (A1 `:85` — the contract
   digest remains the sole tamper authority); recorded so a later reader
   does not add it.

## §E.3 Run-phase Audit-Ready Signal

run_status: audit-ready
run_complete_at: 2026-09-26
run_commit_sha: "9c45a6095"

DoD gate (measured by the lane orchestrator, this run, tree `9c45a6095`):

```
$ go test ./internal/hook/... ./internal/config/... ./internal/contract/... ./internal/mission/... -count=1 -v
→ exit 0, 2442 --- PASS, 0 --- FAIL, all 17 packages ok (log: /tmp/t1245-dod-gate.txt, 425s longest package)
$ GOOS=windows GOARCH=amd64 go build ./...   → exit 0
$ golangci-lint run --timeout=2m             → 0 issues. (baseline equal)
$ git status --porcelain                     → empty (clean tree, HEAD 9c45a6095)
```

**Attribution note (verification-claim-integrity §2):** this section is
authored on orchestrator-provided gate evidence — the DoD gate commands and
outputs above were measured by the lane orchestrator, not re-executed by
manager-develop; manager-develop's own M4 measurements are in §E.2 above.

Milestone summary:

| Milestone | Subject | Commits |
|---|---|---|
| M1 | push serializer | `6e831c5a0`, `7fa7c9bb1` |
| M2 | contract-sign guard | `deee2cacf`, `c5e2895a2`, `81f351c6f` |
| M3 | role constants + carrier pins, rule file + mirror | `7d766dcd1`, `d0d915f3f` |
| M4 | mission-validator projection | `21506ceb0`, `9c45a6095` |


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
