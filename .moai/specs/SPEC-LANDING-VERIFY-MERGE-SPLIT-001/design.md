# SPEC-LANDING-VERIFY-MERGE-SPLIT-001 — design (plan phase)

Design intent only. Identifiers named here are proposals for the run phase; the observable contracts are in `spec.md` §3 and `acceptance.md`. Code evidence for the current behaviour is in `research.md`.

## 1. The shape of the fix

Today one object carries three jobs: the re-measure record says "a command passed", is keyed to a tree, and is valid only while its base equals the tip (`integration_merge_step.go:321-333`). The split separates them:

| Job | New home |
|---|---|
| "Everything that had to be verified was verified, on this tree" | manifest + attempt records + the completeness verifier (REQ-LVM-001..007) |
| "A relevant verification does not need repeating after main moved" | the reuse record, decided from scope content (REQ-LVM-016..020), opened last |
| "Which waiting cards can land right now" | the derived ready view (REQ-LVM-014/015) |
| "How long landing waits, and why" | the landing-event log and metrics (REQ-LVM-009) |
| "How many heavy runs may overlap, and when a dead owner's lease clears" | slot capacity and auto-release (REQ-LVM-010..013) |

The merge step stays the only writer of the integration branch (REQ-LVM-021); everything new either reads, or writes machine-local state under `.moai/state/`.

## 2. Records (field sets)

Stored beside the existing store, machine-local like `.moai/state/remeasure/`.

- **Manifest** (`…/remeasure/manifests/<card>.json`): card id; derived-from base and card HEAD; `calls[]` each with `call_id`, `command`, `packages[]` (resolved by `go list` at derivation, so "zero missing packages" has a denominator), `reason` (`changed` | `importer-of:<pkg>` | `non-go-reader:<path>`), `narrowing` (absent, or `35e9` and permitted only for `internal/cli`), `allow_no_tests` per package; `env_fingerprint` (GOOS, GOARCH, CGO_ENABLED, GOFLAGS, `go version`, tags, declared env names with value digests); `no_go_changes` flag.
- **Attempt** (`…/remeasure/attempts/<card>/<attempt_id>.json`): the fields of REQ-LVM-002; `termination` ∈ `completed` | `aborted:signal` | `aborted:timeout` | `aborted:partial-json`; `supersedes` (attempt id of the aborted attempt a restart replaces — informational, never erasing).
- **Reuse record** (extends the tree-keyed store; the file is keyed by the NEW tree): `form: reuse`, `source_tree`, `source_base`, `new_base`, `moved_commits[]`, `scope_digest_source`, `scope_digest_new`, `decision`, `reasons[]`, plus the copied command, exit code and package terminal events of the source so the existing verifier fields stay populated.

The existing `RemeasureRecord` (`integration_remeasure.go:45-62`) is not changed in meaning: a pre-existing record remains valid under gate-off behaviour (REQ-LVM-007). Whether attempt data is embedded in that struct or lives beside it is a run-phase choice, constrained only by AC-LVM-002.

## 3. Package completeness is observed, not declared

`go test -json` already prints one terminal event per package (`Action` pass/fail/skip, no `Test`). The gate takes the expected package set from `go list` at manifest time and compares it with the terminal events actually seen. Three consequences:

- A partial stream (killed run, `tee` truncation, a build failure before any package event) leaves a package without a terminal event → aborted, never pass. This is the missing denominator in today's record.
- A script-wrapped call (`sh gate.sh`) emits no recognized stream to the verb (the verb classifies only a `go test -json` command, `integration_remeasure.go:194-209`); with the gate on it cannot satisfy the manifest. The known stdout-swallow defect is out of scope, so a lane must run each manifest call as its own `moai integration remeasure -- go test … -json …` invocation. That is the intended cost of part 6.
- `skip` with `[no test files]` is accepted only where the manifest recorded `allow_no_tests`; otherwise it is an empty sweep (REQ-LVM-004), consistent with the "swept set empty asserts nothing" rule.

Abort versus restart is a property of the attempt, not of the call: a signal or timeout is detectable from the child's wait status (`ExitError` signaled flag and the verb's own deadline), a cut stream from the final line not being a complete JSON object or an expected package lacking its terminal event. A restart is simply a new execution and a new attempt ID; the verdict counts the most recent completed attempt per call on the candidate tree.

## 4. Threat and failure analysis of the overlap decision

The decision answers: "if main moved, can the earlier measurement still describe the merged tree?" Chosen mechanism: **content equality of the verification scope**, not path-matching of moved commits. Compute the scope's object identity on the source tree and on the new candidate tree (git tree/blob ids of every scope directory and file); the decision is `no-overlap` only if the two digests are equal and every undecidable class is absent. The moved-commit list is evidence for the reader, not the predicate. This is stricter than "moved files miss the closure" in the cases that matter (a revert-then-reapply, a rename, a move that restores the same content) and cannot be fooled by commit shape.

What the closure can miss, and how each is made impossible or fail-closed:

| Hazard | Why the closure alone misses it | Handling |
|---|---|---|
| Build tags / OS-specific files | `go list` honours only the current GOOS/GOARCH/tags; tag-excluded files are absent from `GoFiles` | scope is **directory-level and recursive**, so excluded files (`IgnoredGoFiles`) are in scope; fingerprint (GOOS, GOARCH, CGO, tags) must equal the source's or the decision is undecidable → overlap |
| `go:embed` | embedded files are not Go files | they live inside the package directory subtree (embed cannot use `..`), so directory scope covers them; `EmbedFiles` additionally listed for the digest |
| `testdata` | ignored by `go list` | directory scope covers it |
| Generated files | produced by `go generate`, may be untracked or ignored | the digest is over the committed tree; a measured package whose directory holds ignored files at record time is undecidable (the attempt's cleanliness check, `RunRemeasure`, already requires a clean worktree) |
| Environment-dependent tests | depend on env vars, locale, TMPDIR, PATH | manifest carries a declared env-name list with value digests; any difference → fingerprint mismatch → overlap; an undeclared env read cannot be detected (residual risk R2) |
| Tests that read non-Go files outside their package (`.claude/`, `.moai/`, templates, docs) | not imports, so invisible to `go list` | **default strict**: any moved non-Go file outside the closure directories is overlap, unless it matches an operator-declared never-read glob (decision-index Q8). Measured effect of strict: 1 of 37 adjacent landings keep a record (`research.md` §6) |
| Tests that read live git state (refs, history, `origin/main`, branch lists) | the tip itself is the input; no file changed | a measured package whose tests invoke git against the working repository rather than a temporary one is undecidable → overlap. The classification is by an explicit per-package declaration; a package with no declaration is treated as live-git-dependent (fail-closed). Residual risk R3: a wrong "hermetic" declaration |
| Module graph | `go.mod`/`go.sum`/`go.work` changes alter every closure | in scope; any difference → overlap |
| Dependencies outside the module | not edited by landings, but `replace` directives can point into the repo | `replace` target inside the repo is added to the scope; outside the repo → undecidable |
| cgo / `go:generate` inputs | C sources are non-Go files in the package directory | directory scope covers them |
| Source tree vs new tree closure | a moved commit can add an import that pulls a new package into the closure | closure computed on both trees; scope is the union |
| Flaky/time-dependent tests | outcome varies without any file change | outside the decision; unchanged from today (a green record is a point observation). Residual risk R4 |
| Reuse chains | each hop non-overlapping, cumulative overlapping | no chaining; moved set always from the root measured record (REQ-LVM-020) |
| A tampered or stale record | hand-written JSON | consumers recompute from the repository (REQ-LVM-019); forgery of the build-identity field stays the existing accepted residual of `integration_remeasure.go`'s trust model |

Residual risks carried to the operator (decision-index): R2 undeclared environment reads, R3 wrong hermetic declaration, R4 flakiness. Batch CI on the remote integration branch is the existing backstop for all three; this SPEC does not remove it.

## 5. Alternatives considered

- **Trace the measured test processes' file reads** to obtain an observed read set. Rejected for this SPEC: macOS has no unprivileged file-access tracer usable in CI-like runs, and a tracer-built set is still a sample of one run. Left as the principled upgrade for strict-mode eligibility.
- **Declare the read set per package in a manifest** (chosen as the narrowing path, default-deny). Accepts R3.
- **Keep base==tip and speed up the re-measure** instead. Not in the card's approved scope; the capacity and instrumentation parts address wall-clock directly.
- **Persist a `ready` marker on the card.** Rejected (REQ-LVM-015): a marker goes stale when main moves, and a stale-marker consumer is exactly the "continued firing" failure (`verification-completeness` §1.3). A derived view cannot go stale; instrumentation records first-observed-ready separately in the event log.

## 6. Ready view

For each card in `merge-ready`: (1) card branch contains the integration tip, or the reuse decision holds against it; (2) manifest gate-complete on the candidate tree; (3) a valid tree-keyed record exists (measured or reuse); (4) the audit evidence names the card's current HEAD — for SPEC-less cards the existing `audited_sha` bind (`factorylane/merge_specless.go:35, 108-138`); for SPEC cards the sync-audit verdict file must name the HEAD (today only `sync_status: complete` is read, `factorylane/merge.go` `checkSyncAudit` — a measured gap this view closes for readiness without changing `complete`'s existing gate); (5) `git merge-tree --write-tree` dry run is clean (`merge.go:289`). The first failing predicate is reported.

Note on HEAD confirmation: after a re-absorb the card HEAD changes, so predicate (4) fails until the audit is repeated. That cost is deliberate and unchanged; the view makes it visible instead of discovered at `complete`.

## 7. Slot capacity and auto-release

- Record shape: the per-resource record gains a `holders[]` list guarded by the existing per-resource mutation lock (`withSlotLeaseMutation`, `slot_lease.go:446`); capacity 1 keeps the single-record behaviour byte-compatible (REQ-LVM-011). Admission of the second or later holder reads load average and available memory through injectable readers; a reader failure refuses (REQ-LVM-012). Thresholds have no shipped numeric default: capacity above 1 with thresholds unset behaves as capacity 1 (fail-closed), which lets the operator choose numbers (decision-index Q9) without this SPEC guessing them.
- Wall-clock-bound exclusion: holders declare `--packages` at acquire; the config list of wall-clock-bound packages is operator-held evidence (decision-index Q11); empty by default.
- Auto-release: the three read paths (`status`, `acquire`, the guard) already compute `Stale()`; the change makes them remove the record and append an `auto-release` row. Liveness stays bound to the owning session's pid (`cli/slot.go:131-133`); binding to the heavy command's process is the open question Q10.

## 8. Instrumentation

Append-only JSONL under `.moai/state/landing/` (machine-local, like the slot audit log); row kinds `ready`, `remeasure`, `reuse-considered`, `landed` with card, tree, SHA, UTC time. The metrics verb derives landing intervals (consecutive `landed`), ready-wait (`landed` minus first `ready` for the card), execution and reuse counts, and the **shadow eligibility** count: at every landing where a re-measure was executed because the tip had moved, whether the reuse decision would have read no-overlap. That shadow counter is the evidence for decision-index Q8 and is why instrumentation precedes reuse.

## 9. Split recommendation (the leader invited it)

The six parts are authored as one Tier L SPEC at 22/22 requirements and criteria, 3 below the ceiling on each axis. Plan audits in this repository repeatedly hit the ceiling on oversized SPECs, and the parts are not equally coupled:

| Proposed SPEC | Parts | Why separable | Tier |
|---|---|---|---|
| A — completeness gate | 6 (REQ-LVM-001..007) + ordering AC | must land first, stands alone, closes the evidence gap for every later landing | M |
| B — heavy-slot capacity and dead-owner release | 3, 4 (REQ-LVM-010..013) | touches only `slot_lease.go`, `cli/slot.go`, config; shares no code with landing | S/M |
| C — ready view, metrics, gated reuse | 2, 5, 1 (REQ-LVM-009, 014..020, 008) | depends on A; reuse depends on ready and metrics | L |

Recommendation: split, with `depends_on: [A]` on C and the REQ-LVM-008 ordering rule becoming a hard dependency between SPECs (stronger than a commit-graph AC). The single SPEC is nonetheless complete and auditable as written; if the plan audit holds at the ceiling, split along the table rather than trimming criteria. The leader decides (decision-index Q1).

## 10. Premise checks recorded

- Card item (4): "the record remains until its bound" is only partly true. A dead owner is already reclaimable without `--force`, and the integration window already clears dead holders; what this SPEC specifies (record removal) is the measured remainder (`research.md` §4).
- Card item (1): the stated benefit depends on eligibility, measured at 1/37 (strict) to 10/37 (lenient) adjacent landings (`research.md` §6).
- Part (6): the weakness of today's record is observed in a real landed tree (`research.md` §3) and in the repository's own stored records.
