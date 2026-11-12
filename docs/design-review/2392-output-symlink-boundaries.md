# Issue #2392: incomplete output and symlink boundary preparation

Status: F6 repaired (see "F6 repair as implemented"); F5c still an ignored RED awaiting
the classification decision (request 6807717170, options and recommendation 6809803334).
Preparation base: `523052622385751f7ccc473cddff709d99834bb1` (agent-19).
Claims: `Moai-Claim: 2392#6782198349` (agent-19, epoch 4, preparation), handed off by
6809689773 to `Moai-Claim: 2392#6809707646` (agent-7, epoch 5, F6 repair).
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

## F6 repair as implemented (agent-7)

`scope::check_and_record` now measures the project top level with
`worktree::identity(project)` (never the record's own `worktree.path`) and, before the
journal lock, runs `physical_findings(root, write_scope)`:

- The raw spelling of each entry is walked from the root; `..` is not folded. An absolute
  entry blocks as `absolute`, any `..` component as `parent_component`.
- Windows spellings are refused as text on every host (PR #2484 audit N1, Codex review
  4545024327): a leading `\`, a UNC path or a drive prefix (`C:`, `c:\`, `C:x`) blocks as
  `absolute`, and any other `\` blocks as `backslash`. `overlap` is unchanged and still
  splits on `/` only; the physical check is what refuses these entries.
- An entry with no part left after trimming and dropping `.` (`.`, `./`, `./.`, blank) covers
  the whole repository in `overlap`, so it is inspected exactly as a root `**`: the same
  subtree walk under the same budget (PR #2484 audit B1, Codex review 4545024323).
- Each level is listed with `read_dir` and checked with `symlink_metadata`; names match
  case-insensitively with the existing `*` rule. A link anywhere the entry reaches (an
  intermediate part, the matched entry, or below a covered directory or `**`) blocks as
  `symlink`. Links are never followed, so internal aliases, dangling links and cycles block
  alike.
- A missing part matches no entry and stays a new path. Any listing or metadata error blocks
  as `unreadable`; more than 100,000 entries read blocks as `inspection_budget`.
- A finding sets `decision=blocked`, `record_update_request.reason=scope_symlink` and the
  journal effect `scope_symlink` with the findings. An integration order does not lift it.
  `physical_check` reports the inspected root and the findings. `check`/`check_with` and
  `overlap` stay lexical and unchanged.

Measured before choosing the conservative rule: a built worktree held no symlink outside
`.git` (`find . -path ./.git -prune -o -type l -print` printed nothing), the repository
tracks no symlink, and the local claim journal does not exist, so no current scope is
blocked by it.

| Boundary | Result | Test |
| --- | --- | --- |
| Directory prefix plus symlink escape | blocked, sentinel unchanged | `boundary_output_scope::scope_symlink_escape_must_be_blocked` (no longer ignored) |
| Literal link, link as an intermediate part, `*`, `l*k`, case variant | blocked | `scope_symlink_cli::links_reached_by_any_scope_form_block` |
| Internal alias, dangling link, self-loop, nested below `**` | blocked, not followed | `scope_symlink_cli::internal_dangling_and_nested_links_block_without_traversal` |
| `link/../x`, `../outside/**`, absolute | blocked | `scope_symlink_cli::absolute_and_parent_entries_block`, unit `inspection_refuses_absolute_parent_and_budget` |
| Plain scope, `*.rs` not matching a link, link outside the scope, new tail | clear | `scope_symlink_cli::plain_scopes_and_links_outside_the_scope_stay_clear` |
| Integration order, journal effect, repeat adds no event | blocked; effect `scope_symlink` | `scope_symlink_cli::integration_order_does_not_lift_a_link_block_and_the_journal_records_it` |
| Inspection budget, unreadable directory | blocked | unit `inspection_refuses_absolute_parent_and_budget`, `inspection_fails_closed_on_an_unreadable_directory` |
| Whole-repository entry `.`, `./`, `./.`, blank, ` . ` | walked as root `**`: link blocks, budget 4 clear and 3 `inspection_budget` (same as `**`) | unit `inspection_walks_the_root_for_an_entry_with_no_parts`, `scope_symlink_cli::links_reached_by_any_scope_form_block` (`.`, `./`) |
| Whole-repository entry without any link | clear (negative control) | `scope_symlink_cli::a_whole_repository_entry_without_links_stays_clear` |
| Windows spellings `C:/x`, `c:\x`, `C:x`, `\x`, `\\server\share`, `..\x`, `safe\link` | blocked as `absolute` or `backslash` on every host; `safe/c:`, `cc:/x` stay plain | unit `inspection_refuses_windows_spellings` |
| Built repository walked by `**` or a whole-repository entry | may block as `inspection_budget` once `target/` passes 100,000 entries: fails closed, an operational false block, not measured | not_run |
| Other active worktree physical alias | not covered: only the scope-check worktree is inspected | not_run |
| Link created after the check (swap race) | out of scope: check-time observation only | not_run |
| Windows runtime | the new tests are `cfg(unix)` | not_run |

Guard mutants (one unique line each, three targets run separately, source restored and
`git diff --quiet HEAD` checked after each): 11 of 12 FAILED a runtime assertion with no
compile error. The survivor, "a missing directory blocks", sat on a `NotFound` branch that
cannot be reached because only observed directories are listed; that branch was removed and
every listing error now fails closed, whose mutant then FAILED.

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
