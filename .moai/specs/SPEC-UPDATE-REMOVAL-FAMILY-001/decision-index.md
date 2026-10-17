# SPEC-UPDATE-REMOVAL-FAMILY-001 — Decision index

Stateless on the status axis: the lifecycle lives in spec.md. Each row follows Detect → Explain → Ask. No row carries an embedded recommendation; a `Default:` is a policy application of the published Default rule, not a recommendation. Authority anchors are committed artifacts; an anchor that cannot be verified routes the row to FOUNDER.

### Q1: What does the removal half of AC-RC30-27 (`TestMigrationRemovesSharedCopyWithoutProfileReach`) assert, given that no row text for it ever existed?

Label: FOUNDER
Class: implementation-level
Why unresolved: SPEC A's acceptance.md recorded the removal half only by name and ownership ("the removal half (TestMigrationRemovesSharedCopyWithoutProfileReach) moves with AC-RC30-01 [REF] to card t1668" — acceptance.md:239); it never carried the half as a row, so there is no carved expected-cell text. This SPEC's authored Given-When-Then is derived from the keep half's mirror and the REQ-RC30-003 gate, not from a recorded ruling.
Options (factual; no recommendation):
- **(가)** The fixture enables the removal switch (`enableRemovalForTest`, test-only) on a launch-ledger-unmapped project whose shipped copy meets the removal preconditions, and asserts the copy is removed — with backup-first ordering composed. This is the authored acceptance.md text.
- **(나)** Any reading that keeps the switch off in the fixture — that asserts the copy survives, which duplicates AC-RC30-02's assertion and leaves the moved criterion unjudged.
Default: (가) (rule: the option that preserves current behavior — it exercises the shipped gate in a test-only fixture and changes no shipped default; the alternative leaves the criterion unjudged).
Operator verdict: DEFAULT-APPLIED 2026-10-17T11:54:29Z manager-spec (card t1668, run tn1p4e)

### Q2: What form does bundle G (sync finding F12, cross-card relation t1642) take?

Label: FOUNDER
Class: implementation-level
Why unresolved: The transfer ruling ("bundle G (F12): 이 카드 범위에서 뺀다 → t1668 이관") exists only in the git-ignored card record (t1619 progress.md:5244; sync-audit-30659e8e1a.md F12/E-13), which cannot anchor a DECIDED row. The committed traces — SPEC A acceptance.md:240 ("the lane's own reach test (M2 bundle G) confirms reach and does not replace this criterion") and the ruling body quoted in the card record ("... CLAUDE_CONFIG_DIR 를 그 프로필로 둔 채 이주를 돌려, 공유 사본이 프로젝트 쪽에 남고 표지가 붙는지 확인한다") — describe the procedure but do not pin whether it lands as a committed Go integration test or as a lane-run recorded procedure.
Options (factual; no recommendation):
- **(가)** A lane-run verification procedure whose output is recorded in progress.md §E.2 (sandbox project mapped via the launch ledger, migration run with `CLAUDE_CONFIG_DIR` at the profile, kept + `KeptByProfileReach` observed). No new committed test.
- **(나)** A committed Go integration test running the update/migration path with the profile mapping set (the `runUpdateFixtureErr` + `profileReachSandbox` pattern already in `internal/cli`).
Constraint stated by the record: the worktree-session guard refused inline HOME overrides in a lane session (ruling d-20261010T124923Z-2b09 part 2); a lane-run procedure that swaps HOME may hit the same refusal — the run-phase executor resolves per the guard boundary and records a Gap on refusal.
Default: (가) (rule: the option with the smaller user-visible surface — no new committed artifact; and the wording closest to the ruling's own "레인 시험" and SPEC A acceptance.md:240 "the lane's own reach test").
Alternate: (나)
Operator verdict: DEFAULT-APPLIED 2026-10-17T11:54:29Z manager-spec (card t1668, run tn1p4e)

### Q3: Which judging command does AC-RC30-26 carry, given that its originally pinned test name was never created?

Label: FOUNDER
Class: implementation-level
Why unresolved: SPEC A's AC-RC30-26 row pinned `TestIntegrityProbeCodexOnlyNoClaudeSettings` as the judging test, but the code repair was assigned to card t1578 by leader correction ("the repair stays with card t1578") and landed there (`ac141b2cb`) with its own test names (`TestIntegrityProbe_ClaudeOnlyMemberFollowsHarness`, `TestRunUpdate_CodexOnlyVersionMatch_PrintsNoIntegrityRow`). The pinned-name test does not exist on the tree (this run's grep: no match). No ruling records which judging command this SPEC carries.
Options (factual; no recommendation):
- **(가)** Re-point the criterion at the two existing t1578 tests (the authored acceptance.md command). No new code.
- **(나)** Add a thin alias test carrying the pinned name `TestIntegrityProbeCodexOnlyNoClaudeSettings`, keeping the carved name discoverable.
Default: (가) (rule: the option that preserves current behavior — the repair's own tests already judge the behavior; no new test code).
Alternate: (나)
Operator verdict: DEFAULT-APPLIED 2026-10-17T11:54:29Z manager-spec (card t1668, run tn1p4e)

### Q4: May this SPEC close with AC-RC30-01 unrun (operator-terminal Gap with a resume condition), or does close wait for the profile-visibility run?

Label: FOUNDER
Class: implementation-level
Why unresolved: AC-RC30-01 requires a real Claude Code profile session; a worktree lane session cannot run it (SPEC A Q8 decision d-20261010T223148Z-1e76 moved the runnable form here precisely because of that; ruling d-20261010T124923Z-2b09 part 2 records the worktree guard refusing inline HOME overrides). Whether THIS SPEC's close tolerates the operator-terminal gap is not decided by any committed record — SPEC A's deferral precedent (AC-RC30-28 deferred to t1642 with a resume condition) concerns a different criterion under a different owner.
Options (factual; no recommendation):
- **(가)** Close with Gap G-01 recorded and a resume condition (operator runs the profile-visibility command and records the output in progress.md §E.2), mirroring the SPEC A AC-RC30-28 deferral form.
- **(나)** Hold close until AC-RC30-01 passes on an operator profile session — blocking the card on an action no lane can execute.
Default: (가) (rule: the option that preserves current behavior — it mirrors the recorded SPEC A deferral precedent; the alternative blocks the card on a lane-impossible action while the removal machinery stays gated off either way).
Alternate: (나)
Operator verdict: DEFAULT-APPLIED 2026-10-17T11:54:29Z manager-spec (card t1668, run tn1p4e)
