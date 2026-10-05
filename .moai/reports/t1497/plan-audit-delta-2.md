auditor-model: glm-5.3-flash[1m]

# SPEC Review Report: SPEC-HARNESS-DETACHED-PRUNE-001 (card t1497) — DELTA AUDIT ROUND 2 (TREE-FINAL)
Verdict: PASS
Overall Score: 0.94 (delta-1 base preserved; the consolidated round ADDS a directly-reproduced defect repair with an anti-vacuous regression design)
Plan Artifact Hash: 655948dcb43f3f531d132096a2bbdd9a3d41e4e489df92cad8957975627b673e
Auditor Version: plan-auditor/v1 (GLM lane)

verdict: PASS
audited_sha: 73264bddf8034c4d309c411393dfe9f710492679
overall_score: 0.94
must_pass_failed: 0
blocking_count: 0
plan_artifact_hash: 655948dcb43f3f531d132096a2bbdd9a3d41e4e489df92cad8957975627b673e
scope: delta
fix_scope: none (PASS)
defect_class: none (no blocking findings)
reread_hunks: spec.md#REQ-DP-009, spec.md#History-v0.2.0-v0.2.1, acceptance.md#LEDGER-DP-I, acceptance.md#LEDGER-DP-D-extension, acceptance.md#LEDGER-DP-GREEN-B-effect-half, acceptance.md#AC-DP-008-row, plan.md#file-6, plan.md#file-7-classifier-window, plan.md#file-8-sweep, plan.md#M2-exit, plan.md#M3-exit

Header facts: auditor model glm-5.3-flash[1m] (GLM lane); tree SHA `73264bddf8034c4d309c411393dfe9f710492679` (HEAD re-read this run, branch `WT-harness-prune-detached`); delta-2 scope = the two post-delta-1 commits: `949fbfc32` (consolidated: REQ-DP-009 + AC-DP-008 + LEDGER-DP-I + fake-sweep obligation + §E.1 divergence record; spec v0.2.0) and `73264bddf` (round-4 fold: child tests moved to cli + aggregation-input filter placement + mixed-vintage regression + history-row restore; spec v0.2.1); date 2026-10-05. Base attribution: delta-1 PASS at `8a430d101` (plan-audit-delta-1.md, hash `e0cd9f75…c1e7`), which rides iter2 PASS at `2d29b0509` — everything outside the amended hunks rides those bases unchanged. Plan-artifact hash pinned at the final tree via the `internal/runtime/audit_cache.go` `ComputeHash` algorithm replicated (Tier M set, 3 files).

## Delta Scope and Adjudication

### 949fbfc32 — consolidated gate round (spec v0.2.0)

- **REQ-DP-009 + AC-DP-008 (finding 1)**: the classifier must apply the retention window itself. The causal finding — with the synchronous prune gone, `classifyHarnessPatterns` aggregates retention-expired events into the learning curator — is **DIRECTLY REPRODUCED BY THIS AUDITOR on the final tree**: a fixture of 5 `agent_invocation` events dated 2026-08-06 (60 days old; `defaultRetentionDays` = 30 per `integration_test.go:129` and the observer calls) fed through the tree-built binary's `hook harness-classify` (CLAUDE_PROJECT_DIR-rooted fixture) returned `harness-classify: 1 patterns → 1 promotions written`, exit 0 — the spurious promotion, end to end. The gate's overlay measurement (5×60d → 1 promotion + 1 proposal) is corroborated at its core.
- **LEDGER-DP-I both commands re-run this round**: `go test -list 'TestClassificationAppliesRetentionWindow|TestStopClassificationFiltersExpiredEvents|TestClassifyExcludesExpiredEvents' ./internal/cli` → `ok ... 1.116s`, zero names (the AC's test is genuinely new); `grep -c "Timestamp\|retentionDays\|cutoff" internal/harness/learner.go` → `0`, exit 1 (no retention token anywhere on the aggregation path — the zero is meaningful because the file aggregates timestamped events, exactly the ledger's stated why).
- **Mechanism claims verified structurally**: `Pattern` (`internal/harness/types.go:304`) carries no per-event timestamp in its fields; `AggregatePatterns(logPath string)` (learner.go:46) takes no cutoff — the post-aggregate alternative was indeed unimplementable, grounding round 4's narrowing.
- **Finding 2 — sweep obligation**: file 8's named-subset enumeration replaced by `grep -rn "runHarnessObserve" internal/cli --include="*_test.go"` enumeration + `retentionSpawnImpl` override on every handler-driving test + whole-package re-verification `go test -timeout 30m ./internal/cli` — the one package-wide run the sweep justifies, matching the house `-timeout 30m` discipline. Sound: a named subset cannot guarantee completeness; a grep-enumerated sweep with a package-wide claim-closing run can.
- **Finding 3 — §E.1 divergence (NOTE ONLY)**: acknowledged per the record's request — this auditor treats `plan_status: audit-ready` (schema prose) and `audit_ready: true` (factory consumer) as the same signal pending a schema-consumer reconciliation; the lane writes the recognized flag only after this PASS. No action required from this audit.

### 73264bddf — round-4 fold (spec v0.2.1)

- **Child tests moved harness → cli** (`TestDetachedChildPrunes`, `TestDetachedChildDoubleSpawnCollapses`): the import-cycle ground is real (`internal/cli` imports `internal/harness`; a harness-side test importing `cli` cycles), and the placement is semantically right — both tests drive the CLI verb's run function. Consistency verified: M2 harness `-list` drops to exactly 3; M3 cli `-list` rises to exactly 5; the `-race` suite splits per package; LEDGER-DP-GREEN-B's effect half re-scoped to `./internal/cli` (count-first form preserved, no alternation — form-sound); **LEDGER-DP-D's second command re-run this round**: `go test -list 'TestDetachedChildPrunes|TestDetachedChildDoubleSpawnCollapses' ./internal/cli` → `ok ... 0.885s`, zero names ✓.
- **REQ-DP-009 narrowed to the AGGREGATION INPUT**: grounded in the re-measured `Pattern` structure (no timestamps — verified above) and the gate's measurement that 4-expired+1-recent aggregates identically to 5-recent. The regression design is the **exact anti-vacuous mutant probe** this discipline demands: mixed-vintage (4 expired + 1 recent) with a threshold-crossing assertion (the recent-only pattern classifies `observation`, never `rule`) — an over-aggressive filter that dropped everything would FAIL it, and an all-expired fixture (the naive design) would have passed vacuously. Exemplary.
- **GEARS**: REQ-DP-009 is a When-pattern ("When the classifier aggregates usage-log events for tier promotion") with the narrowed mechanism; v0.2.0→v0.2.1 quoted semver; the v0.2.0 history row preserved verbatim per the Lessons convention.

## Delta Verification (re-run this round, final tree)

- LEDGER-DP-I: both commands re-run — reproduce (above).
- LEDGER-DP-D second command: re-run — reproduces (above).
- Direct classification RED: reproduced end-to-end (above — the strongest single evidence in this delta).
- Trace verb: `COLLECTED: 9 REQ definitions (acceptance input: read)`, zero UNCOVERED, zero ORPHAN (REQ-DP-009 collected and covered by AC-DP-008; Tier M budget 9 REQ / 8 AC ≤ 16/16, plan header updated).
- CN-4 verb (mandatory full re-run): `COLLECTED: 4 milestones in plan order (M1 M2 M3 M4), 0 exit bindings, 13 ordering candidates`; zero CONFLICT lines; the 13 candidates are the previously-adjudicated benign classes plus the new AC-DP-008/GREEN-B rows (count-first descriptors, "never waiting on the detached child" — a classification-time semantics statement, not a milestone ordering). MP-9 re-decided PASS.
- spec lint --strict: `✓ No findings — all SPEC documents are valid`, exit 0, source-built at the final tree.

## Carry-Over (named per procedure)

- From delta-1 (PASS at `8a430d101`): MP-2 adjudications for REQ-DP-001..008 (unchanged by both delta-2 commits except REQ-DP-009's addition); MP-4 N/A; MP-5 D7 (no new SPEC references); MP-6 D8 (plan §E untouched — syscall placement unchanged); MP-7 (no clarification markers in either diff — full diffs read); MP-8's LEDGER-DP-A..C, E..H, GREEN-A, FORM (byte-identical — the delta-2 diffs touch only LEDGER-DP-D [extended], LEDGER-DP-I [new], GREEN-B [effect-half re-scope]); the registration-half probe (acceptance unchanged since — its red reproduced at delta-1); the anchors (`hook_harness_multi_event_test.go:46`, 9 invocations); the iter2 GOOS baseline.
- From iter2 (at `2d29b0509`): the LEDGER-DP-A..H RED reproductions with the identical-source-state bridge (learner.go and all probed Go sources are byte-identical through every amendment — the four amendment commits touch only the three artifact .md files, verified per `git show --stat`).

## Fail Receipts — Adjudicated Inputs (listed per procedure)

- `rcpt-b1711dc2551a0500a04de518` (13:45, reviewed fail): round-1 P2s — ADOPTED, repaired in `eccd62531` (delta-1 scope).
- `rcpt-25a06b9e9ee7c784137d0615` (14:02, reviewed fail): round-2 P2 — ADOPTED, repaired in `8a430d101` (delta-1 scope); its findings' correctness was re-confirmed by round 4's deeper analysis.
- `rcpt-b58c078c10b08c840293b603` (14:10, gate-unmet fail): this auditor's stdio mint — codex stream closed; no verdict delivered; diagnosed (contention), not an adjudicated finding.
- `rcpt-d59eddb51a4f46ecf1a5c71c` (14:16, gate-unmet fail): second stdio mint, same stream-close — diagnosed, not an adjudicated finding.
- **`rcpt-e10d81162cc1ac86b41e811c` (14:54:31Z, PASS — CITED)**: tool `codex_audit`, tree_root this worktree, minted by the leader's rc.27 server on the TREE-FINAL signal; its review snapshot postdates both delta-2 commits, so no coverage residual applies to the citation.

## Defects Found (delta)

None blocking. Carried advisories (optional class, unchanged by this delta): DA-1 — stale "function field" wording in REQ-DP-007 (spec.md) and AC-DP-006's cell (acceptance.md; now describes a two-layer seam: gate parameter + wrapper var); DA-2 — the registration-half parent-help mutant (test-half covered); DA-3 — the `LEDGER-ACR-J` label unlocatable (substance independent). One new observation, optional: the sweep's completeness claim rests on the `runHarnessObserve` grep — a handler-driving test named differently would evade the enumeration; the whole-package run + REQ-DP-007's no-real-child assertion at review bounds it (acceptable as written).

## Claim / Evidence / Baseline-attribution / Gaps / Residual-risk

- **Claim**: the consolidated round converts a real, directly-reproduced defect (retention-expired events feeding the learning curator) into a precisely-placed, anti-vacuously-tested requirement; the round-4 refinements are technically grounded; the SPEC at `73264bddf` is audit-ready.
- **Evidence**: all commands and outputs cited inline (the end-to-end classify probe; LEDGER-DP-I both commands; LEDGER-DP-D second command; trace 9/0/0; CN-4 0 conflicts; lint 0 findings; hash pin).
- **Baseline-attribution**: everything measured this run, this tree (`73264bddf`), the probe binary built from this tree (`/tmp/t1497-audit/moai-bin`, module build); carried items carry their named prior attributions.
- **Gaps**: the new GREEN commands (TestStopClassificationFiltersExpiredEvents, the moved child tests) are form-verified only — no implementation exists to execute them (run-phase M3 executes them, with the no-tests-token-as-FAIL clauses governing); Windows vet/lint/coverage remain M4 obligations; the leader's pass receipt reviews the final tree but its finding text is in the leader's session (the store carries the verdict token).
- **Residual-risk**: DA-2/DA-3 as above; the sweep enumeration's name-shape dependence (bounded by the package-wide run); the §E.1 dual-spelling reconciliation is deferred to its own concern per the recorded note.
