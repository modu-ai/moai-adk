# SPEC-FACTORY-LEADER-ADOPT-CORE-001 — Implementation Plan

Methodology: TDD (`.moai/config/sections/quality.yaml` `constitution.development_mode: tdd`). Milestones are ordered by decision-reversibility: the new finder type first, then the change to the existing `--factory-run` flag's write path (the most challengeable decision), then the new refusal flow, then the mechanical text surfaces. Execution order is M1 → M2 → M3 → M4; no milestone depends on a later one. Priority labels only; no time estimates.

## A. Context and decisions most likely to change (read these first)

The shrunk core of the held SPEC-FACTORY-LEADER-ADOPT-001 (card t1849): a promptless refusal with guidance on a fresh-mint leader start over an orphan run, the deliberate `--factory-run` adopt pinned to the adoption write, and the distributed guidance text. The interactive chooser, the stream-based interactivity detection and the `--spawn` pre-refusal are a follow-up card — this plan builds no seam for them (the refusal reads no stream).

| # | Decision | Where it lives | Why it may change |
|---|---|---|---|
| A1 | Non-interactive refusal, no run created (DP-1, REQ-FLAC-002/005). | gate in `internal/cli/factory_leader_orphan.go` (M3) | Ruled by the leader under delegation, carried by the shrink; the board record cannot be digest-pinned, so the operator confirms it (decision-index Q1). Flipping it back changes one branch and AC-FLAC-005. |
| A2 | Every `--factory-run` leader start takes over through the adoption write — including runs whose owner is live or indeterminate (DP-2, REQ-FLAC-003). | `adoptExplicitLeaderRun` (M2) | One write path and the address-republish fix everywhere on the explicit path, versus preserving today's unconditional restamp for live-owner takeovers. FOUNDER (decision-index Q2); a flip confines the edit to the gate helper, AC-FLAC-003 and one M2 item. |
| A3 | The refusal is terminal-independent until the chooser card lands (DP-3). | gate condition (M3) | A terminal start over an orphan is also refused. FOUNDER (decision-index Q3). |
| A4 | The adoption basis string is `leader-start-adoption`, distinct from the resume adoption's `leader-resume-adoption` (`internal/cli/factory_run_adopt.go:26`), so the `run.resumed` event tells the two entries apart. | adoption call (M2) | Cosmetic to behavior, visible in the audit trail only. |

## B. Known issues this plan must not reproduce

The predecessor failed three plan-audit rounds; the iter-3 defect classes bind this plan:

- **D14 (unpinned over-refusal)**: the exemption table of AC-FLAC-006 covers every exemption against the refusal surface, and names the over-refusal mutant. There is no `--spawn` surface in this SPEC, so the main gate is the only refusal surface.
- **D15 (one-directional seam)**: no seam-injected default exists in this SPEC (no interactivity detection at all); acceptance.md states this and exercises the one input seam (`classifyNamedRun`) in both directions (AC-FLAC-003 live-owner proceed, AC-FLAC-004(c) abort).
- **D16 (plan/AC contradiction on the README edit)**: M4 REPLACEs the dead-leader sentence in the four READMEs so the banned locale substrings disappear; it never qualifies the sentence in place. The four substrings are present at `README.md:70` / `README.ko.md:70` / `README.ja.md:70` / `README.zh.md:70` at HEAD (verified by grep, ledger E5 note).
- **D17 (stale counts)**: every artifact of this SPEC states 10 REQs and 10 ACs; the SPEC directory carries exactly `^- **REQ-FLAC` ×10 and `### AC-FLAC` ×10.

Pre-existing defect this plan fixes as a side effect of A2: the explicit `--factory-run` leader start leaves `EnvFactoryLeadAddr` pointing at the minted id's socket (`internal/cli/factory.go:799` stamps it; `enterSelectedFactoryRun` replaces only `EnvFactoryRunID`, `:502-503`).

## C. Pre-flight for the run phase

Run before M1, from this tree: `git rev-parse --short HEAD`, `git branch --show-current`, `go build ./...`, and the GREEN baselines of ledger E3 and E4 (acceptance.md) — they must be green before any edit. Heavy suites follow the lane protocol: run the named tests only, never `go test ./...`.

## D. Constraints

- The adopt path never calls `RecordRun`, `ReconcileActiveRuns`, or any retirement (REQ-FLAC-007); the refusal path writes nothing (REQ-FLAC-005).
- The resume adoption's outputs are untouched: `adoptResumedLeaderRun` keeps its three warn lines and its mint-on-no-candidate behavior (`internal/cli/factory_run_adopt.go:64-112`; ledger E3 pins it).
- The gate reuses `strictResumeCarrierSeen` and `leaderRunIDWasMinted` as is — never the broader `scanResumeArgs` vocabulary.
- Template-First for every distributed file: edit `internal/template/templates/...` first, make the local twin byte-identical (`cmp`), then `make build`.
- New skill prose carries no backticked span starting `moai cc -`, `moai glm -` or `moai codex -` that the entry table does not replay (`internal/cli/factory_skill_assertions_test.go`; `TestFactorySkillAssertionsMatchBehavior` is an M4 exit criterion).
- The four locales change in the same commit (README set and docs-site set); Korean is canonical for docs-site content, the anchors are English backticked spans in every locale.
- No time estimates; priorities only. Comments in English (`code_comments: en`).

## E. Self-Verification (what the run phase must re-observe)

- Each milestone's RED step records the behavioral RED verbatim in `progress.md` §E.2 before writing the implementation.
- Ledger E1/E2 (RED-now) flip to GREEN under the swept-count rule; ledger E3/E4 stay GREEN after every milestone.
- E5 flips: the two anchor spans appear exactly once in all ten guidance files; the four banned README substrings are gone.
- `moai spec lint --strict` reports no findings on the final tree, naming the judging build.
- The twin skill files are byte-identical and `make build` leaves no tracked-file change.

## F. Milestones

### M1 — Orphan finder (Priority High; new type interface)

RED first: `TestFindOrphanRunsRequiresDeadStampAndCards`, `TestFindOrphanRunsExcludesNonQualifyingRuns`, `TestFindOrphanRunsCountsNonTerminalCardsOnly` (AC-FLAC-001; new file `internal/homestate/factory_run_orphan_test.go`).

Then in `internal/homestate/factory_run_adopt.go`, beside `FindResumableRun`: a read-only `FindOrphanRuns(ctx, opts ReconcileOptions) ([]OrphanRun, error)` and an `OrphanRun` value that embeds `RunOwner` and carries the non-terminal card count. Approach: `classifyRuns(ctx, opts, "active")` (`internal/homestate/factory_run_retire.go:172-277`), keep owners with `Classification == OwnerDead && Basis == BasisStamp` (the same admit rule as `FindResumableRun` at `factory_run_adopt.go:92`), then count `cards` rows outside `sqlTerminalStates` per kept run and keep runs with a count of at least one — the count is what excludes workers-only runs, where `runHasLiveRecords` would not. The card count is a read on the pool handle after the row cursor is drained (`classifyRuns` drains before classifying); no transaction is open, so the single-connection pool cannot deadlock. SELECT only; the finder writes nothing. The excluded boot-proof fixture of AC-FLAC-001 (run H) is staged through the `ReconcileOptions` seams `BootTime` and `LegacyPeerRecordAbsent` (`factory_run_retire.go:96-101`): the test injects a boot time later than the run's seeded timestamps, so the no-identity row classifies dead on `BasisBoot` (`:266-274`) and the exclusion is realizable without touching the host clock.

Exit: AC-FLAC-001 passes; ledger E4 and the `TestFindResumableRun*` names stay green.

### M2 — Deliberate adopt through the adoption write (Priority High; change to an existing flag)

RED first: `TestExplicitFactoryRunAdoptTakesOverAtomically`, `TestExplicitFactoryRunAdoptRefusals` (AC-FLAC-003, AC-FLAC-004; new file `internal/cli/factory_leader_orphan_test.go`, driving the entry through the `driveFactoryEntry`/`unifiedLaunchFunc` pattern of `internal/cli/codex_factory_retire_test.go:460-474` or `installFactoryLaunchSeam` of `internal/cli/factory_test.go:756-778`, which already captures `EnvFactoryLeadAddr`).

1. In `internal/homestate/factory_run_adopt.go`, a small classifier entry `FindRunOwner(ctx, opts, runID) (RunOwner, error)`: `classifyRuns` filtered to the named run, whatever its classification — the deliberate adopt classifies the run it is pointed at, it does not require an orphan.
2. In `internal/cli/factory_run_adopt.go`, `adoptExplicitLeaderRun(root, runID, backend string) (string, error)`: `FindRunOwner`, then `AdoptRun(ctx, owner, os.Getpid(), homestate.CurrentProcessFingerprint(), backend, "leader-start-adoption")` (A4; the backend argument per item 3), then republish `EnvFactoryRunID` and `EnvFactoryLeadAddr = factory.FactoryLeaderSocketPath(runID)` exactly as `adoptResumedLeaderRun` does at `:109-110`. The codex-led and live-legacy-peer refusals are NOT re-implemented here: `enterSelectedFactoryRun` already applies both as errors before this point (`factory.go:489`, `:496-501`), and REQ-FLAC-004's abort is that error returning unchanged — nothing minted, nothing modified (the first write is still `recordLeaderRunStart`, which this precedes).
3. **The runs-row backend write (audit D1).** `AdoptRun`'s UPDATE sets only `lead_pid, lead_process_start, updated_at` (`internal/homestate/factory_run_adopt.go:128-129`), and `factory.RecordFactoryRunStart` writes the todo runtime store, not the `runs` row (`internal/factory/factory_runtime.go:28-36`); `runs.lead_backend` is otherwise written only by `RecordRun` (`internal/homestate/runtime.go:44-46`) and the resume-discovery INSERT (`internal/homestate/factory_run_resume.go:58`). M2 therefore extends `AdoptRun`'s UPDATE to `SET lead_pid=?, lead_process_start=?, lead_backend=?, updated_at=? WHERE ...` — the adopting leader's backend lands on the row in the SAME transaction as the stamp, so the takeover cannot leave the prior leader's backend behind. The signature gains the backend parameter and both call sites pass their launcher's: the explicit path passes `factory.BackendClaude` / `factory.BackendGLM`; the resume path (`adoptResumedLeaderRun`, `internal/cli/factory_run_adopt.go:106`) passes the resuming launcher's backend — a same-direction truth (the row already names this leader as its owner), and the E3 resume tests assert no `runs.lead_backend` value (grep-verified against `internal/cli/factory_run_adopt_test.go`), so their outputs are unchanged. The five direct `AdoptRun` test callers (`internal/homestate/factory_run_adopt_test.go:148,178,182,204`, `internal/homestate/factory_run_retire_guard_test.go:225`) gain the argument mechanically; no assertion is weakened.
4. Call sites: `internal/cli/cc.go` between `leaderEntryRunAdoption` (`:214`) and `recordLeaderRunStart` (`:218`) — when `adopted == ""` and `entry.FactoryRun != ""`, `adopted, adoptErr = adoptExplicitLeaderRun(launchProjectRoot(), entry.FactoryRun, factory.BackendClaude)`; mirror in `internal/cli/glm.go` (`:274`/`:281`, passing `factory.BackendGLM`). `recordLeaderRunStart` then records only the todo runtime facts under the adopted id (`factory_run_adopt.go:173-179` → `factory.RecordFactoryRunStart`); the runs row's backend comes from item 3's extended `AdoptRun` write, not from this record.
5. The owner-change race (AC-FLAC-004(c)) is testable through a `classifyNamedRun` variable seam inside `adoptExplicitLeaderRun` that the test overrides to restamp the row between classification and write; production initializes it to the `FindRunOwner` call. This is an input seam, not a default — both directions are driven (AC-FLAC-003 proceeds, AC-FLAC-004(c) aborts).

Watch the pinning tests: `TestFactoryJoinAndLeadAcceptNonCodexRun` (lead subtest, `codex_factory_retire_test.go:524-536`) and `TestFactoryLeadRefusesCodexLedRun` (`:540-558`) — both in ledger E3. Their asserted outcomes (success/active/backend for the lead subtest; refusal text, launches 0, row unchanged for the codex one) must still hold through the adoption write; adapt the fixture only if its seeded classification makes the compare-and-set miss, without weakening any assertion. The lead subtest's backend assertion reads the runs row through `runRow` — with item 3's extension the row's backend is written by `AdoptRun` itself, which is what keeps that assertion green.

Exit: AC-FLAC-003 and AC-FLAC-004 pass; ledger E3 green.

### M3 — Refusal gate and guidance block (Priority High; new user-facing flow)

RED first: `TestLeaderStartOrphanRefusalPrecedesRunWrite`, `TestLeaderStartOrphanGuidanceRendersFacts`, `TestLeaderStartOrphanGating`, `TestLeaderStartOrphanTouchesNoOtherRun`, `TestLeaderStartOrphanGLMMirrorsCC`, `TestLeaderStartOrphanTimesRenderInKST` (AC-FLAC-002, 005, 006, 007, 008, 009; same new test file). The no-stdin guarantee of AC-FLAC-002 is instrumented with a counting reader: the test drives the entry with the cobra command's input set to a reader that records every `Read` invocation and asserts zero calls — a gate that prompts (or reads input in any form through the command) trips the counter; an empty reader would leave no consumption trace and prove nothing (audit D4).

1. New file `internal/cli/factory_leader_orphan.go`: the refusal error sentinel; the guidance renderer (one block: one header line, then one line per orphan run with run id, `dead`, lead pid, start time in KST, non-terminal card count, then one line naming `<entry command> --factory-run <run-id>` per the block's own list, then one line stating no run was created and no session was started); the KST renderer (`time.FixedZone("KST", 9*3600)`, layout `2006-01-02 15:04:05` plus ` KST`; a value that does not parse as RFC 3339 returns unchanged — no non-test KST formatter exists in the tree, grep-verified); the two documentation anchor constants of AC-FLAC-010. The gate takes the launcher's entry command (`moai cc -f` from `cc.go`, `moai glm -f` from `glm.go`) so the guidance names the launcher's own command, never the hard-coded `moai cc -f` of the resume warn line (`factory_run_adopt.go:79`).
2. Gate condition, evaluated after the resume-adoption decision and before `recordLeaderRunStart`: `!strictResumeCarrierSeen(args) && entry.FactoryRun == "" && leaderRunIDWasMinted(leaderLabel, priorRunID)`, then `FindOrphanRuns`; with at least one orphan, render the block to `cmd.ErrOrStderr()` and return the refusal error — `cc.go`/`glm.go` return it without the "re-adopt factory run" wrap, giving a non-zero exit. With zero orphans, return as today. Everything before the record changes only environment variables restored by the existing defers (`cc.go:202-207`), so the refusal writes no run and starts no session. No stdin is read and no terminal is probed anywhere (DP-3).
3. Both call sites: `internal/cli/cc.go` (between `:214` and `:218`) and `internal/cli/glm.go` (between `:274` and `:281`), one shared function, the entry command the only per-launcher argument (REQ-FLAC-008).

Exit: AC-FLAC-002 and AC-FLAC-005..009 pass; ledger E3 and E4 green.

### M4 — Guidance text (Priority Medium; mechanical)

RED first: `TestLeaderAdoptGuidanceDocsMatchAnchors` (AC-FLAC-010; the anchors are built from the same constants M3 defines).

Template-First: edit the template file, make the local twin byte-identical (`cmp`), then `make build`. Two anchor lines, English backticked spans in every locale, each exactly once per file: `` `leader-start/orphan: refused; no run created; adopt with --factory-run <run-id>` `` and `` `lane-join/dead-owner-with-cards-or-workers: kept` ``.

| Surface | Files | Change |
|---|---|---|
| Factory skill | `internal/template/templates/.claude/skills/moai/workflows/factory.md` and `.claude/skills/moai/workflows/factory.md` (twins, 100 lines each today) | One new H2 between `## Entry forms` (`:32`) and `## Session record` (`:68`): the leader start over an orphan run is refused with guidance and creates no run; `--factory-run <run-id>` is the deliberate adopt; a lane join does not retire a dead-owner run with in-progress cards or workers. Both anchor lines carried. No backticked `moai cc -`/`moai glm -`/`moai codex -` span outside the entry table. `metadata.version` bumped. |
| docs-site | `docs-site/content/{en,ko,ja,zh}/advanced/factory-mode.md` (9 level-2 headings each today) | One new H2 per locale, same position, same content meaning, both anchors inside it. The existing dead-owner paragraph inside `## When several runs are alive` (en `:67`) is not rewritten — the retire-guard text there landed with SPEC-FACTORY-RUN-RETIRE-001 and stays. Heading-count equality across the four locales is asserted, not a fixed number. |
| README | `README.md`, `README.ko.md`, `README.ja.md`, `README.zh.md` (the dead-leader paragraph, line 70 in each) | REPLACE the sentence that says a dead leader's run is retired automatically when the next lane joins — the banned substrings (`is retired automatically when the next lane joins`, `자동으로 회수되므로`, `次にレーンが合流した時点で自動的に retire されるので`, `会在下一条泳道加入时自动退役`) must DISAPPEAR, never be qualified in place (known-issue D16) — with the qualified statement (retired only when it holds no in-progress cards or workers), and add the leader-start sentence: a start without a resume option over such a run is refused with guidance and creates no run; adopt deliberately with `--factory-run <run-id>`. Both anchor lines carried. |

Not touched: the `factory-dispatch*.md` and `cross-session-messaging*.md` rule files (spec.md §F).

Run-phase verification for this milestone also includes the `hns-oss-docs-verify` recipe (warning-free Hugo build, 4-locale parity, README heading parity). Exit: AC-FLAC-010 passes; `TestFactorySkillAssertionsMatchBehavior` green.

## G. Anti-Patterns for this change

- Calling `RecordRun`, `ReconcileActiveRuns`, `RetireRunIfDead` or any retirement from any path this SPEC adds (REQ-FLAC-007).
- Falling through to a mint after the refusal (REQ-FLAC-005) or after a refused / owner-changed adopt (REQ-FLAC-004).
- Guidance text naming the other launcher's command, or copying the hard-coded `moai cc -f` of `factory_run_adopt.go:79` (REQ-FLAC-008).
- Widening the resume test to `scanResumeArgs`'s broader vocabulary (REQ-FLAC-006; `-c` is not a resume carrier).
- Re-deriving "dead" from a pid probe in the launcher instead of the finder's classification (REQ-FLAC-001).
- Reading stdin, probing a terminal, or rendering any prompt (spec.md Terms; the chooser is a follow-up card).
- Qualifying the README dead-leader sentence in place instead of replacing it (known-issue D16).

## H. Cross-References

- `spec.md` §C — the root-cause reading with this tree's line numbers; `acceptance.md` — the evidence ledger (E1-E5) and the ten ACs.
- Siblings: SPEC-FACTORY-RUN-RETIRE-001 (completed in this tree — the retire guard whose exits already name the adopt command), SPEC-FACTORY-RUN-SELECT-CORE-001 (lane-join chooser lineage), SPEC-FACTORY-LEADER-ADOPT-001 (the held predecessor; its iter-3 audit's D14-D17 classes bind §B).
- `decision-index.md` — Q1/Q2/Q3 (FOUNDER rows the operator settles before Kickoff).
