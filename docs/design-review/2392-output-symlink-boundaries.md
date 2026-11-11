# Issue #2392: incomplete output and symlink boundary preparation

Status: RED reproduced; production repair not implemented. No PR is authorized.
Base: `523052622385751f7ccc473cddff709d99834bb1`.
Claim: `Moai-Claim: 2392#6782198349` (agent-19, epoch 4).
Sources actually read: Issue #2392 body/comments, comparison 6793261122,
`src/verify.rs`, `src/scope.rs`, and the committed decisions log.
Design comment: 6807658907. Reproduction and decision request: 6807717170.

## Observed RED

Run explicitly: `cargo test --offline --locked --test boundary_output_scope -- --ignored --nocapture`.
Both tests select and execute real product entry points. Compilation failures,
zero-selected runs, and fixture setup failures are not the claimed RED.

- `incomplete_output_must_not_be_passed`: the command exits 0 after a ready
  handshake with a detached synthetic pipe holder. The holder cannot close stdout
  or stderr until `run_check` returns and the controller releases it. Observed:
  `output_complete=false`, both digests null, `result=passed`. The final assertion
  rejects Passed without choosing between Failed, NotRun, or a new Unknown value.
  Socket timeouts are safety bounds; elapsed time and sleeps do not decide RED.
  The holder is released before assertions about the returned check.
- `scope_symlink_escape_must_be_blocked`: a committed synthetic Git repository and
  seeded claim journal exercise native `scope-check --task`; its `check_and_record`
  result is `clear` for
  `safe/**` despite `safe/link` targeting a sibling directory outside the synthetic
  project. A conditional simulated authorized write changes the external sentinel
  from literal `ORIGINAL` to `ESCAPED`. Its preservation assertion fails. Only
  disposable fixture files are accessed; no actual HOME or user asset is used.

The tests remain explicitly ignored while the production defects remain open,
following `tests/boundary_repro.rs`. Default-suite success does not mean these
boundaries passed. The explicit RED command must remain part of the report.
Unix runtime is exercised; Windows runtime is not_run.

## F5c decision gate

No user decision for F5c was found in the read ledger. The earlier author's
Unknown expectation is a proposal, not approval. Production `verify.rs` is unchanged.
Options reported to the leader before implementation:

1. NotRun plus an explicit evidence-incomplete reason; preserve the process exit.
   The name needs clarification because the command actually ran.
2. Failed plus EnvironmentFailure; never count collection failure as assertion RED.
3. A new Unknown/incomplete verdict with run/audit acceptance and compatibility work.

Raw output is currently not retained. Digest-only provenance must be stated
honestly; adding raw-log persistence requires a retention/redaction decision.
No log store or schema change is silently included in this preparation.

## F6 repair design

Recommended first repair: conservatively refuse a scope whose reachable existing
path contains a symlink. Do not claim that lexical overlap is a physical-path guard,
and do not follow a link merely to decide whether its destination is acceptable.
This deliberately blocks even an internal symlink; canonical alias support can be
separate work if required. The repair remains not_run in this preparation.

The root-aware entry point is `check_and_record(project, ...)`; `overlap(a, b)` is
an owned lexical interface and cannot determine physical aliases without a root.
Keep that interface compatible. Resolve the authoritative worktree root from the
real project/journal identity before examining scope paths: do not trust arbitrary
`ClaimRecord.worktree.path` supplied by a caller or scan an unrelated HOME path.

Proposed algorithm and limits:

1. Validate root identity and relative scope syntax. Absolute paths, a prefix that
   leaves the root, malformed/unsupported patterns, and unobservable metadata must
   produce a blocked/unknown result rather than clear. Inspect existing components
   before discarding `..`: `link/../x` must not erase the symlink observation.
2. Inspect with `symlink_metadata` before descending. Never traverse the target of
   any link. Dangling links, cycles, internal aliases, and links to a sibling or
   external directory all block under the conservative rule.
3. Apply the existing case-conservative matching rules. A plain directory covers
   descendants; `*` and `**` require checking reachable entries, not just the
   literal prefix before the first wildcard. Missing new tails are allowed only
   after existing ancestors are observed safe. Metadata/read-directory errors and
   any inspection budget exhaustion must fail closed, with a measured reason.
4. Check relevant active claims as well as the current scope. A lexical no-overlap
   result must not hide a physical alias in another worktree. Only registered
   worktrees of this clone may be inspected; unknown identity stays uncertain.
5. Emit a distinct symlink/inspection conflict with root-relative path and reason,
   preserve the existing record-owner and journal fencing behavior, and do not let
   an integration-order string override a filesystem escape. Do not write either
   claimant's files or resolve an unknown state to PASS.

Required RED-to-GREEN matrix before claiming repair:

| Boundary | Required result | Current execution |
| --- | --- | --- |
| Directory prefix plus symlink escape | blocked, sentinel unchanged | RED reproduced |
| Literal link, internal alias, dangling link, cycle | blocked without target traversal | not_run |
| Wildcard-matched link and nested `**` | blocked | not_run |
| `link/../x`, absolute, above-root scope | blocked | not_run |
| Case variant and shared-interface alias | conservative conflict | not_run |
| Safe literal/new tail and unrelated link | valid scope remains usable | not_run |
| Other active worktree physical alias | conflict/uncertain, never clear | not_run |
| Metadata error or inspection budget | blocked/unknown | not_run |
| Journal update failure and integration-order bypass | no false clear/write | not_run |
| Guard deletion/bypass mutants | runtime assertion FAILED, selected >0 | not_run |

This is an observation at check time, not a filesystem sandbox or an atomic
check-and-write guarantee. A later symlink swap is a separate race; do not claim
this design eliminates it. Production changes, GREEN, mutants, and exact-head
hosted CI remain unverified until implementation and actual execution.

## Preparation verification (base 5230526223)

Production source changes: zero. These results validate preparation only, not repair.

| Command | Exit | Observed result |
| --- | --- | --- |
| `cargo test --offline --locked --test boundary_output_scope -- --ignored --nocapture` | 101 | Two selected, two runtime assertion failures: incomplete evidence accepted as Passed; native scope-check allowed the sentinel escape |
| `cargo test --offline --locked` | 0 | 109 target results; 1,216 passed, zero failed, five ignored (three existing plus the two new unresolved RED tests) |
| `cargo clippy --offline --locked --all-targets -- -D warnings` | 0 | Completed with no warnings |
| `cargo fmt --check` | 0 | Formatting accepted |
| `python3 -m unittest discover -s tests -p '*_test.py'` | 0 | 303 tests, OK, one skipped |
| `git diff --cached --check` | 0 | No whitespace errors |

All native checks used synthetic HOME and a dedicated target. Initially, a runner
TMPDIR inside the checkout caused two unrelated library assertions to fail
(245 passed, two failed). Moving the synthetic TMPDIR outside any checkout made
all 247 library tests pass, then the complete suite passed as above. An intermediate
controller setup also exposed macOS inheritance of nonblocking socket mode; the
fixture now explicitly makes the accepted control stream blocking. Only final
failures at the intended product assertions count as the reported RED.

F5c classification/raw-log policy remains awaiting the leader's decision request
6807717170. F6 repair implementation, GREEN, repair mutants, Windows runtime, and
hosted CI for this preparation are not_run. The existing Passed misclassification
and symlink escape remain unfixed. No PR is opened by this preparation.
