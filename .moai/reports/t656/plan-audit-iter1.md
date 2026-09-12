# SPEC Review Report: SPEC-UPDATE-SETTINGS-BASE-SNAPSHOT-001
Iteration: 1/2 (Tier M ceiling per plan-auditor Retry Loop Contract; harness.yaml not re-read — see Gaps)
Verdict: FAIL
Overall Score: 0.73 (harmonic mean of the four dimensions; Tier M PASS threshold 0.80)

Card: t656 · Tree: `/Users/goos/MoAI/moai-adk-go/.claude/worktrees/t656` · Branch `WT-update-value-merge` · HEAD `81c1d58f9` (verified: `git rev-parse --show-toplevel`, `git branch --show-current`, `git rev-parse --short HEAD`).
Reasoning context ignored per M1 Context Isolation. Operator decisions A1/B1/C1 were treated as fixed inputs and audited for faithful encoding only.
Cross-model audit: `audit_model` is not set in `.moai/config/sections/` (grep returned nothing), so this is a Claude-only audit with no MCP backend call.

The verdict is FAIL for two independent reasons:
- MP-7 fails on its own, whatever the score (§Must-Pass).
- Separately, there are six major blocking defects in capture-point correctness and AC falsifiability, and the aggregate score is below the Tier M threshold.

## Must-Pass Results

- [PASS] MP-1 REQ number consistency — REQ-USB-001 … REQ-USB-015 are sequential with no gaps or duplicates and consistent 3-digit padding (spec.md:86-114). AC-USB-001 … AC-USB-015 are likewise sequential (acceptance.md:44-170).
- [PASS] MP-2 GEARS format (requirement layer only) — all 15 `REQ-USB-*` entries in spec.md §C use a GEARS pattern:
  - Ubiquitous: 002, 005, 011, 013, 014, 015.
  - Unwanted (`shall not`): 003.
  - Event-driven (`When`): 001, 009, 010.
  - Capability gate (`Where`): 004.
  - Compound `Where … When`: 006, 007, 008, 012.
  - The Given-When-Then ACs in acceptance.md are the verification layer and were not graded here.
  - Atomicity note (optional, F-15): REQ-USB-013 joins two subjects in one entry.
- [PASS] MP-3 YAML frontmatter — all 12 canonical fields are present with correct types (spec.md:2-13): `version: "0.1.0"` is quoted, `status: draft`, `created`/`updated` are ISO dates, `priority: P1`, `phase: "v3.2.0 target"`, `lifecycle: spec-anchored`, and `tags` is a comma string. No snake_case aliases. `moai spec lint .moai/specs/SPEC-UPDATE-SETTINGS-BASE-SNAPSHOT-001/spec.md` printed `✓ No findings — all SPEC documents are valid` with exit 0. `mcp__moai__spec_audit` (project_root = worktree) returned `drift_findings: []`, `modern_era_clean: 1`.
- [N/A] MP-4 language neutrality — this is a Go-internal CLI change. NFR-USB-004 (spec.md:121) forbids touching `internal/template/templates/`, and the SPEC names no multi-language tooling.
- [PASS] MP-5 D7 cross-SPEC — every referenced SPEC exists, and none is retired, superseded or archived:

  | Referenced SPEC | Status |
  |---|---|
  | TEMPLATE-BASE-SNAPSHOT-001 | completed |
  | HOOK-DELIVERY-001 | completed |
  | REINSTALL-LOOP-001 | completed |
  | MERGE-CONFLICT-BLIND-001 | in-progress |

  No BLOCKING finding. The in-progress sibling's tension is a design decision (D6), not a D7 lifecycle finding.
- [PASS] MP-6 D8 cross-platform — `grep -c syscall` gives 0 in spec.md, plan.md, acceptance.md and progress.md, so this auto-passes.
- [FAIL] MP-7 clarification gate — `grep -n 'NEEDS CLARIFICATION'` found two unresolved markers:
  - plan.md:78 (Decision D5).
  - plan.md:86 (Decision D6).

  research.md does not exist, which is expected at Tier M. This is folded in as F-01 (critical). The orchestrator reports both topics are already escalated. The gate still stands until they are resolved and folded into plan.md and the ACs.

## Category Scores (0.0-1.0, rubric-anchored)

| Dimension | Score | Rubric Band | Evidence |
|-----------|-------|-------------|----------|
| Clarity | 0.75 | 0.75 | Requirements are mostly unambiguous (spec.md:86-114). Material ambiguity remains in three places: "settings.json 스냅숏" does not say whether it is a staging copy or the canonical base (REQ-001 spec.md:86 vs REQ-005 spec.md:94; plan.md:104) — F-06; the B3 abort premise is misstated (plan.md:34) — F-07; the AC-008 warning text is undefined (acceptance.md:112) — F-09. |
| Completeness | 0.90 | 1.0 (structure) | HISTORY spec.md:22, background §A spec.md:26, decisions §B spec.md:70, REQs §C spec.md:84, NFRs §D spec.md:116, limitations §E spec.md:125, and seven `### Out of Scope — <topic>` H3s with bullets (spec.md:135-162) are all present; frontmatter is complete. Deducted for content gaps: the init deploy-skip case (F-05) and the two missing D5(a) cases (F-07). |
| Testability | 0.60 | 0.50 | AC shape and RED reasons are careful (acceptance.md:31-40), but several falsifiability claims fail. The M-09 mutant survives AC-009 (F-04). AC-007's source-position predicates admit placement mutants (F-03). The normal-update REQ-005 path has no falsifying test (F-02). AC-009 miscounts its tests, 8 vs 7 (F-08). AC-008's warning assertion is not discriminating (F-09). |
| Traceability | 0.75 | 0.75 | The acceptance.md:13-29 table covers all 15 REQs, and every AC traces to an existing REQ. Coverage is indirect for three: REQ-005 is tested only on the clean-reinstall path (F-02), REQ-010 only on clean-reinstall (F-11), and REQ-001's init and template_sync write sites only by source position (F-03). |

Harmonic mean: 4 / (1/0.75 + 1/0.90 + 1/0.60 + 1/0.75) ≈ 0.73.

## Defects Found (structured defect-list)

D1. F-01 — plan.md:78, plan.md:86 — two unresolved `[NEEDS CLARIFICATION]` markers (Decisions D5 and D6). acceptance.md:203 also defers a D5-dependent AC. — Severity: critical — Class: blocking (MP-7) — Required fix:
- Resolve D5 and D6 through the orchestrator's operator channel.
- Replace both markers with the decisions taken.
- Add the D5 AC that acceptance.md:203 promises.
- Fold in the F-07 corrections before presenting D5.

D2. F-02 — acceptance.md:87-95 (AC-USB-006), acceptance.md:101 (AC-USB-007 `template_sync`) — REQ-USB-005 has no falsifying acceptance criterion on the normal `moai update` path.
- Where the test runs: AC-006 runs `runCleanReinstall` twice. That path opens only on a v2 fingerprint (`update_clean_install.go:174` returns a no-op otherwise), so production never runs it twice on one project.
- Where the risk is: in the template-sync path, deploy runs inside the Deploy Templates step (`update_template_sync.go:364`) and the merge runs later in Restore Settings (`:544`). This is exactly where an early canonical write turns base into updated.
- Why nothing catches it: an M-06 mutant on the template-sync path places the write between `deployWithMirrorNotice(` and `updatemerge.MergeUserFiles(`. That is the position AC-007's source test requires, so the mutant passes every listed AC.

Severity: major — Class: blocking — Required fix:
- Add a behavioural two-cycle AC on the template-sync ordering. One option: extract the Backup → deploy → capture → merge → promote sequence into a seam testable inside `internal/cli/update/{merge,backup}`, so no slot is needed.
- Add an "M-06 on template_sync" row to §E that turns that AC red.

D3. F-03 — acceptance.md:101, acceptance.md:192-194 — AC-USB-007's source-position predicates for `init` and `template_sync` accept placements that violate REQ-001 and REQ-003.
- `template_sync`: a write inside `if configBackupPath != ""` (`update_template_sync.go:497-532`) sits textually after `:364` and before `:544`, so it passes. Yet plan.md:111 names that very spot as wrong, because it is skipped when `configBackupPath == ""`.
- `init`: a write inside the error branch of `executor.Execute` (`init.go:868-876`) sits textually after `executor.Execute(` and before `ApplyAutonomyTierBundle(`, so it passes, but it never runs on success.

Severity: major — Class: blocking — Required fix: make both subtests behavioural. Failing that, tighten the predicate to "an unconditional statement in the same block as the deploy call's success continuation", and add both placement mutants to §E.

D4. F-04 — acceptance.md:119-122, acceptance.md:196 — the AC-USB-009 claim that M-09 turns the test red is false as written.
- Both projects have "같은 sections 입력" and run `WriteSnapshot` before `HasSnapshot`.
- After `WriteSnapshot`, `sections/` is non-empty in both projects.
- The real `HasSnapshot` (`snapshot.go:114-131`) and the M-09 mutant (root-emptiness check) therefore both return true, and the mutant survives.

Severity: major — Class: blocking — Required fix: add a cell with no sections snapshot but with `.moai/cache/template-snapshot/claude/settings.json` present. Assert `HasSnapshot == false` and that `SaveTemplateBase` takes the embedded-raw fallback (`base_loader.go:27-31`).

D5. F-05 — plan.md:110 (Decision D4, init row), spec.md:86 (REQ-USB-001) — the init capture site cannot tell whether the deploy actually wrote `settings.json`.
- The init deployer is non-force: `init.go:821` `NewDeployerWithRenderer`, and the slim variant is equivalent.
- `deployer.go:239-255` skips any existing file whose manifest provenance is `UserModified` or `UserCreated`. An untracked file is recorded as `UserCreated` and skipped. The skip is recorded in `DeployResult.ProtectedSkips`.
- So when `moai init` runs in a directory that already has its own `.claude/settings.json` (or `init --force` over one), the file on disk after `executor.Execute` is the user's file, not the render.
- Capturing it as the base would make the next update read every user-edited shared leaf as "unchanged" and overwrite it with the template value. That is silent loss of user data: the A2 hazard the design memo rejected (design-options.md:25).
- REQ-001 is worded conditionally ("배포가 렌더된 … 을 쓰면"), but no plan step or AC observes that condition.

Severity: major — Class: blocking — Required fix:
- Record the snapshot only when this flow's deploy actually wrote `settings.json`: consult the skip result, or capture render bytes rather than re-reading disk.
- Add an AC with a pre-existing, untracked user `settings.json` at init, asserting no snapshot is written (or a render-bytes snapshot, whichever D5 decides).
- Add the matching mutant ("capture from disk regardless of skip").

D6. F-06 — plan.md:15, plan.md:104, plan.md:120, spec.md:86 vs spec.md:94 — the two options Decision D3 leaves open are not both compatible with the test contract M1/M3 already pin.
- The contract: AC-001/003/012 call today's `MergeUserFiles(root, backups, out)` with the snapshot planted at the canonical path (plan.md:15, plan.md:120), so the merge reads the base from that path.
- Option 1 (hold the previous bytes in memory in the Backup step while still writing canonically at deploy) then makes the merge read the just-written render. That is M-06, unless the signature or a seam changes, which contradicts plan.md:15.
- The spec never says whether "the settings.json snapshot" recorded right after deploy (REQ-001) is the canonical base or a staging copy. REQ-001 and REQ-005 can only both hold under a staging-then-promote reading.

Severity: major — Class: blocking — Required fix: in spec.md, define the recorded artifact (staging path vs canonical base) and when it is promoted. In plan.md, either fix D3 to staging-then-promote, or state that option 1 needs a base-injection seam and amend plan.md:15.

D7. F-07 — plan.md:34 (B3), plan.md:80 (D5 option a) — the D5 premise and the recommended rule are both inaccurate.
- The abort case is not the same shape as merge-failure preservation. The template-sync path writes the user's `settings.json` to a disk backup before the destructive step (`update_disk_backup.go:47-53`, "in-memory-backups"), and the managed clean removes the live file (`deploy.go:56-60`).
- After an abort between deploy and merge, the live file is therefore the fresh render. Option (b) is then correct (base equals live file). Option (a) keeps render N-1 as the base, so every N-1→N template change reads as a user change.
- The two cases coincide only if the user then restores the file from the backup by hand.
- Option (a)'s promotion condition ("병합 결과를 실제로 썼거나, 사용자 파일이 렌더와 같아") also omits two write-without-merge cases:
  - init, which has no merge at all;
  - a flow where the user had no `settings.json`, so no `FileBackup` exists (`update_template_sync.go:489-493`) and `MergeUserFiles` never sees the file.

  Under (a) as worded, neither case would promote.

Severity: major — Class: blocking — Required fix:
- Restate B3 as three cases: merge-fail-preserve, abort without restore, abort then manual restore.
- Phrase promotion as "the live `settings.json` at the end of the flow incorporates this render", and enumerate init and no-user-file explicitly.
- Give the operator the corrected tradeoff (see Open Decisions).

D8. F-08 — acceptance.md:121 — AC-USB-009 requires "목록의 기존 테스트 8개가 모두 `--- PASS`", but acceptance.md:118 lists 7 existing tests. All 7 exist: `snapshot_survival_test.go:20,76,98` and `snapshot_provenance_test.go:90,138,169,204`. A tester counting 8 cannot pass. — Severity: minor — Class: blocking — Required fix: change 8 to 7.

D9. F-09 — acceptance.md:110-113 — AC-USB-008's "settings 스냅숏 기록 실패 경고가 한 줄" defines no text. The planted `.moai/cache` file also makes the sections snapshot writer emit `Warning: template snapshot write failed: …` (`update_snapshot_hook.go:28-30`) on the same `out` (`update_clean_install.go:494`), so the assertion cannot tell the two warnings apart. — Severity: minor — Class: blocking — Required fix: fix a distinct sentinel or message for the settings-snapshot warning, and count only that line.

D10. F-10 — acceptance.md:81, acceptance.md:90 — AC-005, AC-006 and AC-008 describe a harness that does not exist yet.
- "`stubDeployer` 는 배포 때 렌더 바이트 `R` 을 쓴다", but `stubDeployer.Deploy` writes nothing (`update_clean_install_test.go:47-52`). AC-006 also needs a different render per cycle.
- AC-006's second cycle only runs if cycle 1 leaves a v2 fingerprint in place (`update_clean_install.go:174`). Otherwise it prints "not a v2 project — no-op", and the RED/GREEN reading happens for the wrong reason.

Severity: minor — Class: blocking — Required fix:
- Declare the test-only harness extension (a render-bytes field, per-cycle renders) in plan M4.
- State the fixture condition that keeps cycle 2 on the reinstall path (e.g. `makeScenarioA` keeps `system.yaml` `moai.version: v2.16.1`; `update_clean_install_test.go:93-94`).
- Add an assertion that cycle 2 did not print the no-op line.

D11. F-11 — spec.md:104 vs acceptance.md:107-114 — REQ-USB-010 covers `moai init` and `moai update`, but AC-USB-008 exercises clean-reinstall only. — Severity: minor — Class: optional — Required fix: add init and template_sync cells, or record explicitly that the shared helper (plan.md:114) plus the behavioural wiring tests provide that coverage.

D12. F-12 — plan.md:16 — citation error: `templateManaged` is cited as `merge.go:53-60`. The function is at `internal/cli/update/merge/base.go:53-60`; `merge.go:53-60` is inside `MergeGitignoreFile`. — Severity: minor — Class: blocking (a wrong citation, per audit focus 1) — Required fix: change it to `base.go:53-60`.

D13. F-13 — acceptance.md:134 — the contents of AC-USB-011's fake `.mcp.json` snapshots (`mcp.json`, `.mcp.json`) are unspecified. The "AC-USB-001 값 모양" clause covers only the backup and the render. If the fakes do not hold the old values, M-11 may leave the output bytes unchanged. — Severity: minor — Class: optional — Required fix: set both fakes to AC-001's old render, so any mutant that finds them flips the result.

D14. F-14 — acceptance.md:178 — the third AC-USB-015 command is a pipeline (outside the single-invocation evidence form in verification-completeness.md §2.1). It counts `Setenv("HOME"` on removed and context lines too, and reads the committed range only. — Severity: minor — Class: optional — Required fix: count added lines only, e.g. `git diff "$CARD_BASE"..HEAD -G 'Setenv\("HOME"' --name-only -- '*_test.go'`, which should print nothing, and state that it runs after commit.

D15. F-15 — spec.md:110, spec.md:114 — REQ-USB-013 bundles two subjects (the sections snapshot and `internal/merge`). REQ-USB-015 names a function (`MergeUserFiles`) in the requirement layer. — Severity: minor — Class: optional — Required fix: split REQ-013 into two, and name the behaviour ("the post-deploy per-file merge") instead of the function.

D16. F-16 — `internal/cli/update_clean_install.go:394-397` (not in the SPEC) — the existing comment "settings.json base is unavailable in the embedded FS … MergeUserFiles preserves the user's file wholesale" already contradicts `base.go:53-60`, and will contradict this SPEC's behaviour too. plan.md:134 updates only the `base.go:29-34` comment. — Severity: minor — Class: optional — Required fix: add this comment to M4's comment-update list.

D17. F-17 — spec.md:108 (REQ-USB-012), spec.md:130 — scope assessment.
- REQ-012 is a necessary consequence of A1 plus B1, not scope creep. With a real base, the unchanged engine's `inBase && !inCurrent && inUpdated` branch drops the key (`internal/merge/strategies.go:410-412`). It is disclosed at spec.md:130.
- The consequence is material, and no operator decision records it:
  - A user who removes a template key (a `hooks.<Event>` entry, `statusLine`, `permissions.deny`) never again receives template fixes to that key.
  - Combined with the F-07 merge-fail-preserve case, keys that never reached the user can be dropped permanently.

Severity: major — Class: optional — Required fix: record it as an acknowledged consequence of A1+B1 in spec.md §B (one line), so the operator confirms it together with D5. No REQ change needed.

## Open Decisions (escalated; assessment only, not re-opened)

**D5 — promote the fresh render when the merge preserved the user file wholesale or the flow aborted.**
- The recommended option (a) is the right direction for the merge-fail-preserve case (user JSON corrupt, or an engine error at `merge.go:229-236`). There, promoting render N would misread its new keys as user deletions (REQ-012).
- The plan's reasoning is incomplete on three points (F-07):
  1. For abort without restore, (b) is correct and (a) is harmful, because the live file is the render.
  2. (a) as worded fails to promote in the init and no-user-file cases.
  3. Neither option handles abort then manual restore well, since only a staging-then-promote rule decided by the live file's content can.
- My assessment: a rule of "promote iff the live `settings.json` at end of flow incorporates this render (merge result written, byte-equal to render, or render written with no user file)" dominates both options. It needs D3 to be staging-then-promote (F-06). Present the operator with this corrected three-case tradeoff, not the current binary one.

**D6 — relation to SPEC-UPDATE-MERGE-CONFLICT-BLIND-001 (in-progress).**
- The tension is real on the text:
  - REQ-UMC-010 (sibling spec.md:159) begins as a subsystem-wide invariant: "The merge subsystem shall continue to resolve a shared key in favour of the user's value".
  - Its exclusion at sibling spec.md:208 is unqualified: "Changing which value the merge writes for a shared key (REQ-UMC-010 forbids it)".
- Option (a)'s reading ("a limit on that SPEC's remedies") is plausible from the second clause and from the cited intent (`base.go:109-111`, "a user's edit to that key read as their change", which a snapshot base honours). It still leaves a literal contradiction between two live SPECs.
- The sibling's pinned characterization cells will not turn red: `conflict_blind_repro_test.go` feeds `deriveTemplateBase` straight into `MergeFile` (fixture lines ~69-74; `untouched_shared` predicts `["template-old"]`, line 132-135), so it keeps exercising the derived-base path this SPEC leaves unchanged.
- My assessment: (a)'s landing order is sound. This SPEC changes which conflict branches are reachable for settings.json, so landing it first gives the sibling's M2.1 design (entry gate at sibling plan.md:179) the correct premise. However, resolution should include a **textual amendment** to REQ-UMC-010 and the :208 exclusion, scoped to "a shared key the user edited", made by the sibling's owner. An interpretive reading alone leaves a standing contradiction for the next auditor.

## Audit-focus answers (brief)

1. **Citations.** Verified at `81c1d58f9`:
   - `base.go:13-19, 29-34, 112-131, 128`
   - `merge.go:173-258, 216-217, 217-225, 229-236, 245-247`
   - `strategies.go:360-467, 418-462`
   - `update_template_sync.go:341-371, 364, 487-494, 495-547, 531, 544, 641-645`
   - `update_clean_install.go:398-403, 459, 494, 507, 531`
   - `update.go:379-387`
   - `init.go:867, 889-900, 1030`
   - `autonomy_bundle.go:89-93` (the regeneration branch sits at the `RenderTierPermissions` call, ~:89-92)
   - `snapshot.go:22, 51, 114-131`; `base_loader.go:26-67`
   - `base_test.go:143-159, 269-332`; `update_snapshot_hook_test.go:13-101`
   - `.gitignore:352` = `.moai/cache/`; `templates/.gitignore:241` = `.moai/cache/`
   - `settings.json.tmpl` has 18 `{{` lines

   One error found: plan.md:16 (F-12). One qualifier on acceptance.md:204: `:641-645` is the version skip in `runTemplateSyncWithProgress`; the step-flow function `runTemplateSyncWithReporter` has its own skip at file line ~100, and the claim holds for both.
2. **Capture points.**
   - Normal update: the capture is achievable. The user file is read into memory before Clean, Clean deletes the live file, and deploy writes the render. REQ-005 then depends on D3's mechanism (F-06), and the only test is clean-reinstall (F-02). The v3 `stripRetiredV2DenyEntries` runs before backup, on current rather than render (`update.go:379-387`).
   - Clean-reinstall: correct placement (after `:459`, before `:507`, well before `:531`).
   - Init: flawed whenever the deploy skips an existing `settings.json` (F-05). Otherwise it is correct before `ApplyAutonomyTierBundle`.
3. **AC falsifiability.**
   - RED-now predictions checked against the code and plausible: AC-001 (derived base → all leaves read "only user changed"), AC-003 and AC-013 (conflict line only), AC-012 (key re-added).
   - The JSON strategy sets `HasConflict: len(conflicts) > 0` (`strategies.go:323/352`), so the conflict line is reachable under a snapshot base.
   - Mutant table: M-01, M-02, M-04, M-05, M-06 (clean-reinstall only), M-07c, M-08 and M-11 (conditional, F-13) plausibly flip their ACs. M-09 does not (F-04). M-07a and M-07b are caught, but equivalent placement mutants are not (F-03).
4. **REQ-USB-012.** A necessary consequence and disclosed; confirmation recommended (F-17).
5. **Added exclusions.**
   - `.mcp.json` / `status_line.sh` are justified. They share `MergeUserFiles` and would otherwise widen blast radius. `status_line.sh` is not key-structured anyway (`base.go:79-80` returns false).
   - The conflict-instrumentation exclusion is justified and avoids overlap with the in-progress sibling.
   - The sections-defect exclusion is exactly one bullet (spec.md:138), as the operator required.
6. **Korean register.** The artifacts read as clean native written Korean. No material calques found.

## Regression Check (Iteration 2+ only)

Not applicable (iteration 1).

## Recommendation

FAIL. For manager-spec, in priority order:

1. Close F-05: make the init capture conditional on the deploy actually writing `settings.json`, and add the AC plus mutant.
2. Close F-06 and F-07 together: define staging versus canonical in spec.md, fix D3's mechanism, rewrite B3/D5 as three cases with the live-file promotion rule, and hand D5 back to the orchestrator with the corrected tradeoff.
3. Close F-02 and F-03: add a behavioural template-sync ordering test (via a slot-free seam if possible), make AC-007's init/template_sync subtests behavioural or tighten the predicates, and extend §E with the placement mutants and M-06-on-template_sync.
4. Close F-04, F-08, F-09 and F-10: fix AC-009's cell and count, give AC-008 a unique warning sentinel, declare the stub harness extension and AC-006's cycle-2 precondition.
5. Fix the citation in F-12.
6. After the operator decides D5 and D6, remove the markers at plan.md:78 and :86, add the D5 AC (acceptance.md:203), and record the REQ-UMC-010 amendment path.
7. Optional: F-11, F-13, F-14, F-15, F-16, F-17.

## Gaps (what this audit did not verify)

- No `go test`, `go build` or mutant run, as instructed. Every RED-now prediction and every mutant-flip judgement above comes from reading code, not from observation.
- F-05's premise was read from `deployer.go:239-255` and `init.go:821`. I did not measure how often an existing `.claude/settings.json` reaches init untracked or with `UserModified` provenance, nor whether `DeployResult.ProtectedSkips` reaches the `init.go` call site (initializer.go folds only the mirror notice into `result.Warnings`, `initializer.go:423-430`).
- I did not confirm that the v2 fingerprint survives cycle 1 in AC-006. That reasoning rests on `makeScenarioA` and `detectV2Fingerprint` aggregation (`v2_detection.go:~142-144`) with `stubMigrateRunner` not removing `.agency/`.
- I did not read `restore.go` / `restore_entry.go` in full (a `settings` grep returned nothing), so whether `runUpdateRestore` restores `.claude/settings.json` is unverified. This bears on the "abort then restore" branch of F-07.
- I did not read `newTemplateSyncDeployer`'s force flag. The analysis relies on Clean removing the live file first (`deploy.go:56-60`).
- I did not re-read `.moai/config/sections/harness.yaml` for the Tier M iteration ceiling (the 2 comes from the agent contract).
- I did not verify whether Claude Code discovers a nested `.claude/settings.json` inside the cache (the SPEC's own Gap, spec.md:180).
- I read the sibling SPEC only at spec.md:150-165 and 200-212, plus HISTORY lines 25-27 and plan.md grep hits; its §A.6 body and full M2.1 were not read.
- Cross-model second opinion: not run (`audit_model` unconfigured).

## Residual risk

- Even with every finding fixed, the design keeps a first-cycle window (spec.md:127). It also turns user deletions into permanent opt-outs (F-17).
- Any writer outside the three sites that rewrites project `settings.json` (for example `moai tool-policy build --local-only`, `tool_policy.go:74`, and `ApplyAutonomyTierBundle`) will correctly read as a user change. However, keys those writers remove will read as user deletions and will not be restored.
