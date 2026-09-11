# SPEC Review Report: SPEC-UPDATE-SETTINGS-BASE-SNAPSHOT-001
Iteration: 2/2 (Tier M ceiling, `.moai/config/sections/harness.yaml:75-78` `plan_audit_tier_ceilings: M: 2` — this is the final permitted iteration)
Verdict: FAIL
Overall Score: 0.75 (harmonic mean; Tier M PASS threshold 0.80; iter1 0.73 → iter2 0.75, no regression, so no STOP signal)

Card: t656 · Tree: `/Users/goos/MoAI/moai-adk-go/.claude/worktrees/t656` · Branch `WT-update-value-merge` · HEAD `de59caf62` (parent `81c1d58f9`), working tree clean. Verified with `git rev-parse --show-toplevel`, `git branch --show-current`, `git rev-parse --short HEAD` / `HEAD~1`, `git status --short` (0 lines).

The commit `de59caf62` touches only `.moai/` (per `git show --stat`), so every code citation below was read at `81c1d58f9` and is identical at `de59caf62`. The sibling amendment was assessed against `git show de59caf62 -- .moai/specs/SPEC-UPDATE-MERGE-CONFLICT-BLIND-001`.

Reasoning context was ignored per M1 Context Isolation. The binding decisions (A1, B1, C1, D5, F-05, F-17, D6) were treated as fixed inputs and audited for faithful encoding only.

Cross-model audit: `audit_model` is not set under `.moai/config/sections/` (grep exit 1), so this is a Claude-only audit and no MCP backend was called.

The verdict is FAIL for two independent reasons:

- **MP-2 fails.** REQ-USB-005 carries the D5 conditional rule in prose sentences that have no GEARS modifier and no subject+shall.
- **Four major blocking defects in the D5 falsification surface:**
  - leftover-judgement placement (N-02);
  - AC-016 c4/c5 cannot discriminate (N-03);
  - AC-007 template_sync is blind to wiring order (N-04);
  - N-01 is the MP-2 failure itself.

Every iter1 finding is resolved as instructed; the regression check below has the detail.

## Must-Pass Results

- [PASS] **MP-1 REQ number consistency.** REQ-USB-001 … 016 are sequential with no gaps or duplicates and 3-digit padding (spec.md:142-172). AC-USB-001 … 016 are sequential too (acceptance.md:46-210).
- [FAIL] **MP-2 GEARS format (requirement layer, spec.md §C only).**
  - 15 of 16 REQs conform:
    - Ubiquitous: 002, 003 (`shall not`), 011, 013, 014, 015.
    - Event-driven: 001, 009, 010.
    - Capability gate: 004.
    - Compound `Where … When`: 006, 007, 008, 012.
    - Unwanted with `When`: 016.
  - **REQ-USB-005 (spec.md:150) is mixed formal/informal.** Sentence 1 is a Ubiquitous "The update subsystem shall …". Sentences 2-4 carry three conditional behaviours as untagged prose:
    - "승격은 … 반영하고 있을 때만 일어나며"
    - "정상적으로 끝난 흐름에서는 … 판정한다"
    - "승격 판정 전에 중단된 흐름에서는 … 발견한 시점에 … 판정하며"
  - Those sentences have no `When`/`While` modifier and no subject-shall, although every other conditional REQ in this SPEC bolds its GEARS keyword. This is the "mixed informal/formal within a single requirement" failure. It is new in v0.3.0: iter1's REQ-005 was sentence 1 only.
  - The Given-When-Then ACs in acceptance.md were not graded here; they belong to the verification layer.
- [PASS] **MP-3 YAML frontmatter.** All 12 canonical fields are present with correct types (spec.md:2-13): `version: "0.3.0"` is quoted, `status: draft`, `created`/`updated` are ISO dates, `priority: P1`, `phase: "v3.2.0 target"`, `lifecycle: spec-anchored`, and `tags` is a comma string. There are no rejected aliases.
  - `moai spec lint` on both SPECs printed `0 error(s), 0 warning(s)`, exit 0. The only finding was an INFO `OwnershipTransitionUnmeasured` on the sibling's historical transition.
  - `mcp__moai__spec_audit` (project_root = worktree, filter = this SPEC) returned `drift_findings: []`, `modern_era_clean: 1`.
- [N/A] **MP-4 language neutrality.** This is a Go-internal CLI change. NFR-USB-004 (spec.md:179) forbids touching `internal/template/templates/`, and no multi-language tooling is named.
- [PASS] **MP-5 D7 cross-SPEC.** The extracted references are TEMPLATE-BASE-SNAPSHOT-001 (completed), HOOK-DELIVERY-001 (completed), REINSTALL-LOOP-001 (completed) and MERGE-CONFLICT-BLIND-001 (in-progress). None is retired, superseded or archived, so there is no BLOCKING finding.
- [PASS] **MP-6 D8 cross-platform.** `grep -c syscall` returns 0 in all four artifacts.
- [PASS] **MP-7 clarification gate.** `grep -rn 'NEEDS CLARIFICATION'` over the SPEC directory returned exit 1 (no match). research.md is absent, as expected at Tier M. iter1 F-01 is closed.

## Category Scores (0.0-1.0, rubric-anchored)

| Dimension | Score | Rubric Band | Evidence |
|-----------|-------|-------------|----------|
| Clarity | 0.70 | 0.50-0.75 | Most REQs are single-reading. REQ-005 (spec.md:150) pins the leftover-staging judgement only as "다음 흐름이 새 대기본을 기록하기 전에". Every flow's staging-write site is after the live file has been removed or rewritten, so two reasonable engineers would place it differently, with different outcomes (N-02). AC-016's "다음 흐름 시작 판정으로 확정본 …" has no observation point (N-03). |
| Completeness | 0.90 | 1.0 (structure) | HISTORY spec.md:22-35, §A background, §B decisions (incl. §B.2 two states, §B.4 D5, §B.5 F-17), §C 16 REQs, §D NFRs, §E limitations, and seven `### Out of Scope — <topic>` H3s with bullets (spec.md:194-221) are all present. Deducted for the undisclosed derived-mechanism deviation (N-08). |
| Testability | 0.65 | 0.50-0.75 | The iter1 falsifiability defects are closed (AC-009 cell B, AC-008 sentinel, AC-014, M-07d/e). New gaps: c4/c5 cannot discriminate because R3 == R2 (N-03), AC-007 template_sync cannot see wiring order (N-04), AC-005's observer is infeasible as described (N-05), the AC-007 init cell depends on the environment (N-06), and promotion failure is untested (N-07). |
| Traceability | 0.80 | 0.75-1.0 | acceptance.md:13-30 maps all 16 REQs, and every AC traces to an existing REQ. Coverage is indirect in two places: the REQ-010 promotion-failure half (N-07), and REQ-005's leftover judgement in the real flows, which is covered only through the seam (N-02/N-04). |

Harmonic mean: 4 / (1/0.70 + 1/0.90 + 1/0.65 + 1/0.80) = 4 / 5.328 ≈ 0.75.

## Defects Found (structured defect-list)

D1. **N-01** — spec.md:150 (REQ-USB-005) — MP-2: three conditional behaviours in sentences 2-4 carry no GEARS modifier or subject-shall (see Must-Pass). — Severity: critical — Class: blocking (MP-2) — Required fix: restate REQ-005 as one REQ with three GEARS sentences, which keeps the Tier M ceiling of 16 REQs:
  - "The update subsystem shall change the canonical base only by promoting a staging copy, and shall not promote a flow's staging copy before that flow's settings.json merge step ends."
  - "**When** a flow ends normally, the update subsystem shall promote that flow's staging copy unless the merge wrote the pre-flow user file back wholesale, and shall otherwise discard it."
  - "**When** a flow finds a staging copy left by an earlier flow, the update subsystem shall, before any step of that flow removes or rewrites the live `.claude/settings.json`, promote the leftover if the live file is byte-identical to it and discard it otherwise."

  The third sentence also closes N-02. The Korean body may be kept; each clause still needs its own modifier and `shall`.

D2. **N-02** — spec.md:150; plan.md:110, :164 (D4 bullet) — the leftover-judgement position is pinned only relative to the flow's own staging write. In every flow that write sits after the live file has been replaced:
  - template_sync: Clean deletes it (`update_template_sync.go:329-337`) and Deploy rewrites it (`:364`).
  - clean-reinstall: Step 5 deploy (`update_clean_install.go:459`).
  - init: `executor.Execute` (`init.go:867`).

  Why it matters:
  - A judgement placed at the D4 staging-write site compares the leftover staging with the NEW render, not with the post-abort file.
  - In the ordinary cross-version case (next render ≠ aborted render), operator case "abort, no restore → promote" becomes "discard". That changes a binding D5 outcome.
  - template_sync already captures the pre-flow bytes at `update_template_sync.go:491`, before Clean, so the correct placement is feasible.

  Severity: major — Class: blocking — Required fix:
  - Pin the comparison to "the live file as it stood when the flow began, before any of its steps removed or rewrote it" in REQ-005 (N-01 wording) and in D4. Name the three positions: before Clean (or against the Backup-step bytes) in template_sync, before Step 5 in clean-reinstall, before `executor.Execute` in init.
  - Add a wiring-level cell (see N-04).

D3. **N-03** — acceptance.md:213, :221-222, :234-236 — AC-016 sets the next render `R3 = {"a":2,"K":1}`, byte-identical to `R2`. Worked through:
  - **c4:** with the correct rule, M-D5d or M-D5b, the next flow's merge sees current R2 == updated R3. It takes the no-op branch (`merge.go:207`) and ends with the same bytes, so `a == 2, K == 1` cannot discriminate.
  - **Intermediate assertions have no observation point.** "다음 흐름 시작 판정으로 확정본 R2" (c4) and "확정본 R1 유지" (c5) name no point between the leftover judgement and the next flow's own staging write. After the next flow completes, c5's canonical is R3, not R1, so read as a final-state claim c5 is false.
  - **M-D5f survives c4.** Its "c4 RED" claim (acceptance.md:236, :263) is false: the new staging R3 has the same bytes as R2, so a post-overwrite judgement promotes identical bytes.
  - **The N-02 placement mutant also survives c4,** because the post-deploy live file R3 equals staging R2.

  Severity: major — Class: blocking — Required fix:
  - Make R3 ≠ R2, e.g. `R3 = {"a":3,"K":1,"L":1}`.
  - c4, next-flow Then: `a == 3`, `L == 1`, canonical `R3`. With the correct rule, base R2 gives a template-only change a: 2→3. M-D5d / M-D5b / N-02 placement leave base R1, so a is changed on both sides (1/2/3), conflicts, and stays 2 (RED). M-D5f promotes R3 before the merge, so a stays 2 and L is dropped as a deletion (RED).
  - c5, next-flow Then: `a == 3`, `K == 1`, `L == 1`. Correct base R1 gives template-only changes. M-D5a / M-D5e give base R2, so K is dropped as a deletion (RED).
  - Either name the observation seam for the intermediate canonical assertions or delete them.
  - Add a §E row for the N-02 placement mutant.

D4. **N-04** — acceptance.md:106-117 (AC-007 `template_sync`), plan.md:174 — the template_sync cells assert only that the final canonical equals the render.
  - Suppose the wiring writes the canonical directly in the Deploy Templates step, bypassing the staging and the seam (M-06t at the wiring level). It ends in the same state.
  - The merge at `update_template_sync.go:544` then reads base == updated. Yet the cells carry no prior canonical and no render key missing from the user file, so nothing observes the dropped key.
  - AC-006 exercises only the seam. So plan.md:174's claim that AC-007 "behaviourally confirms" the flows call the seam is overstated: it confirms the end state, not the ordering. This is iter1 F-02's hazard one layer up.

  Severity: major — Class: blocking — Required fix:
  - In both template_sync cells, add a prior canonical `R1`, a user file == `R1`, and a render `R2` that adds key `K` and changes one leaf. Assert `K` and the new leaf value are in the result.
  - Add a §E row "M-06t-w: canonical written in the Deploy Templates step" → RED.

D5. **N-05** — acceptance.md:83 (AC-005 Given) — the "병합 호출 직전" observation is attributed to the deployer test double. The merge runs after `Deploy` returns and after `RestoreMoaiConfig` (`update_clean_install.go:490`, `:507`), so a Deployer double cannot observe it.
  - `MergeUserFiles` is called directly, with no function seam (`grep mergeUserFilesFn` → 0 hits).
  - M4's harness extension (plan.md:205) declares only a render-bytes field and `ResultDeployer`.

  Severity: minor — Class: blocking — Required fix: declare the pre-merge observation seam in M4 (the D8 flow seam's pre-merge hook, or a package-level func var around the merge call) and cite it in AC-005.

D6. **N-06** — acceptance.md:111, :114 (AC-007 `init`) — `ApplyAutonomyTierBundle` rewrites the project settings.json only when all of these hold (`internal/core/project/autonomy_bundle.go:59-94`):
  - the resolved tier is not semi-auto;
  - the sandbox-proof and kill-switch gates do not downgrade it;
  - `tool-policy.yaml` loads.

  The Given names only `tool-policy.yaml`. Under an unpinned tier or gate, the bundle is a no-op and "자율성 번들이 고친 뒤의 settings.json 과 다르다" cannot pass. — Severity: minor — Class: blocking — Required fix: pin `opts.AutonomyTier` and the gate outcome (via a seam), or simulate the bundle as AC-016 c6 does.

D7. **N-07** — spec.md:160 (REQ-USB-010) vs acceptance.md:120-129 (AC-008) — REQ-010 requires a distinct, non-blocking warning for promotion failure too. AC-008 plants only a write failure and defines only `settings-snapshot-write-failed:`. A mutant that returns a flow error on promotion failure survives. — Severity: minor — Class: blocking — Required fix: add a promotion-failure cell (staging present, canonical target unwritable) with its own sentinel and an exactly-once count, plus a matching §E row.

D8. **N-08** — spec.md §E / plan.md:107-110 (derived D5 mechanism) — undisclosed outcome deviation.
  - The case: an abort with no restore, followed by any write to the live file before the next flow (a hand edit, a Claude Code project-settings write, `moai tool-policy build`, the autonomy bundle).
  - The derived next-flow-start judgement then discards the staging, while the operator's rule text ("흐름이 끝났을 때 살아 있는 파일" = pure render) says promote.
  - The deviation is fail-safe. No user data is lost; the aborted render's template changes read as user changes, so later template changes to the same leaves conflict and do not arrive. It is still a departure from a binding decision's literal outcome, and the SPEC records none of it.

  Severity: minor — Class: blocking (faithful-encoding disclosure only; D5 itself is not reopened) — Required fix: add one §E line plus a note under plan D5 "두 시점" recording this case and its fail-safe direction, so the operator sees it. If a different outcome is wanted, escalate; do not change it silently.

D9. **N-09** — acceptance.md:269 (§F version-match skip) — if an abort lands between Deploy and Restore Settings, `.moai/config` has already been cleaned and redeployed (`update_template_sync.go:337`, `:364`).
  - The next plain `moai update` then plausibly returns at `:100-106` before any step, and the leftover judgement waits for a forced or later-version update. This lengthens the N-08 window.
  - The recovery manifest points the user at `moai update --restore`, which does not touch `.claude/settings.json` (see Regression F-07 note).
  - The choice is left to the run phase.

  Severity: minor — Class: optional — Required fix: decide in plan D4 whether the leftover judgement runs before the version-skip return, or record the delay as a limitation.

D10. **N-10** — spec.md:150 / plan.md:109 — the normal-end signal ("흐름 이전의 사용자 파일이 통째로 되돌려 쓰이지 않았는지") is not tied to a mechanism. A byte-compare implementation would read a merge result that happens to be byte-identical to the user file as a wholesale preserve and keep the old base, and no cell covers it. — Severity: minor — Class: optional — Required fix: define the signal as "the merge took a preserve path" (`merge.go:219`, `:232`), not as a byte comparison.

D11. **N-11** — plan.md:145 (D7 option 1) — `DeployResult.ProtectedSkips` is documented as recording published-skill-path skips (`internal/template/skill_mirror.go:69-75`). In non-force mode the deployer actually records every protected skip (`internal/template/deployer.go:244`, `:252`). Option 1 therefore leans on behaviour the field's contract does not state. — Severity: minor — Class: optional — Required fix: prefer option 2 (manifest provenance + hash; a skipped untracked file is tracked `UserCreated`, deployer.go:250-251), or have run-phase update the field comment together with option 1.

D12. **N-12** — sibling SPEC-UPDATE-MERGE-CONFLICT-BLIND-001 — the amendment is otherwise consistent (Regression D6 below), but three passages still read in the old unqualified sense:
  - spec.md:126: "the resolution is designed and stays".
  - spec.md:40 and §D spec.md:206-208: they name `SPEC-UPDATE-TEMPLATE-BASE-SNAPSHOT-001` (sections-scoped) as the deploy-time-snapshot closure, with no pointer to t656 for settings.json.
  - plan.md:210-217: option (b), "a base that can differ from `updated`", sits beside "Whichever is chosen, `REQ-UMC-010` binds: the value written for a shared key does not change". Option (b) would change written values. The tension is pre-existing, and sharper now that t656 is option (b) for settings.json.

  Severity: minor — Class: optional (the M2.1 entry gate, sibling plan.md:193-201, already forces a revisit) — Required fix: qualify spec.md:126 with "for files still merged under the derived base", and add a t656 pointer at spec.md:40, spec.md:206-208 and plan.md option (b).

## Regression Check (Iteration 2)

| iter1 | Disposition | Evidence (current text / code) |
|---|---|---|
| F-01 NEEDS CLARIFICATION | RESOLVED | grep exit 1; plan D5 (plan.md:85-117) and D6 (:119-123) record the decisions; the D5 AC exists (AC-USB-016, acceptance.md:210-236). |
| F-02 normal-update REQ-005 falsifier | RESOLVED as instructed (seam option) | AC-006 `split_deploy_then_restore_order` (acceptance.md:99) and M-06t (:247); seam in plan D8 (:168-174). The wiring-level residual is new finding N-04, not a regression. |
| F-03 AC-007 predicates | RESOLVED | AC-007 is behavioural (acceptance.md:106-118): an empty/filled `configBackupPath` cell kills M-07d, a success-path assertion kills M-07e, and the fallback predicate is tightened (:115). |
| F-04 AC-009 M-09 | RESOLVED | Cell B (acceptance.md:138-146) has no sections and only the canonical. Real `HasSnapshot` returns false (`snapshot.go:114-131`). The M-09 root-emptiness mutant returns true, and `SaveTemplateBase` then walks a missing `sections/` (`base_loader.go:33-64`) → error or bytes ≠ `SaveTemplateDefaults` (`backup.go:152`) → RED. |
| F-05 init skip | RESOLVED | REQ-USB-016 (spec.md:172), §B.1 supplement (:90), AC-014 with 3 cells + M-14 (acceptance.md:180-193). D7 feasibility verified: `init.go:821` is non-force; `deployer.go:239-255` skips and records; `initializer.go:423-430` (internal/core/project) folds only mirror lines, so a new transport is needed, as plan.md:145 states. |
| F-06 staging vs canonical | RESOLVED | spec.md §B.2 (:100-109) and REQ-002 (:144); D3 (plan.md:135-139); `MergeUserFiles` already takes `projectRoot` (`merge.go:173`), so no base-injection seam is needed. |
| F-07 B3/D5 premise | RESOLVED | plan.md:44 B3 is restated; the D5 table lists 8 cases including init and no-user-file (:96-105). A premise note, verified by reading: `moai update --restore` → `RestoreFromBackupDir` → `RestoreMoaiConfig` only (`restore_entry.go:47-79`, `update_restore.go:43-62`). The sections path reads `sections/*.yaml` only (`restore.go:85-114`), and the legacy walk writes under `.moai/config` (`restore.go:197-221`). So `--restore` does not restore `.claude/settings.json`; case 5 is reachable only by a manual restore, and c5 should take the rename plan.md:115 anticipates (not executed; M1 still owns the observation). |
| F-08 test count | RESOLVED | acceptance.md:133-135 lists 7 tests; :145 says "7개". |
| F-09 warning sentinel | RESOLVED | `settings-snapshot-write-failed:` exactly once, and `Warning: template snapshot write failed:` (`update_snapshot_hook.go:29`) 0 times (acceptance.md:125). The promotion half is new finding N-07. |
| F-10 harness | RESOLVED | M4 declares the render-bytes field, `ResultDeployer`, and the v2-fingerprint precondition plus a no-op-line assertion (plan.md:205). The stub still writes nothing today (`update_clean_install_test.go:47-52`), as declared. The AC-005 observer is new finding N-05. |
| F-11 REQ-010 coverage | RESOLVED (optional accepted) | Rationale recorded at acceptance.md:126. |
| F-12 citation | RESOLVED | plan.md:23 now reads `base.go:53-60` (`templateManaged` verified at base.go:52-59). |
| F-13 AC-011 fakes | RESOLVED | acceptance.md:158: both fakes hold AC-001's old render. |
| F-14 AC-015 pipeline | RESOLVED | Single-invocation `git diff … -G … --name-only` run after commit, with a non-empty control (acceptance.md:197-208). |
| F-15 REQ split / function name | RESOLVED | REQ-013 is sections-only (spec.md:166), REQ-014 is the engine (:168), REQ-015 has no function name (:170). |
| F-16 stale comment | RESOLVED | plan.md:48 B7 and :207; the comment is confirmed at `update_clean_install.go:393-396`. |
| F-17 user-deleted keys | RESOLVED | spec.md:138 carries exactly "사용자가 지운 템플릿 키는 이후 템플릿 수정을 받지 않는다 — 운영자 확인 2026-09-11"; §E :188. |

Encode check of the binding decisions:

- **A1** — spec.md:85-89 (write sites :88, post-merge exclusion :89, fallback :86).
- **B1** — :91.
- **C1** — :92 and §F :197 (exactly one sections-defect line).
- **D5** — the five operator cases are at spec.md:122-128. The record that option (a)'s purpose is preserved is at :132. The superseded abort explanation is at :133. The restore M1 item is at :134 and plan.md:112.
- **F-05** — :90.
- **F-17** — as above.
- **D6** — sibling spec.md:162 is verbatim "A remedy for REQ-UMC-008 or REQ-UMC-009 shall not change which value the merge writes for a shared key."; the exclusion is aligned at :213; `status: in-progress` is unchanged; the landing order is at sibling HISTORY :28 and plan M2.1 gate (+10 lines).

All are encoded faithfully.

## Assessment of the derived D5 mechanism (focus 1)

Normally finished flows are judged at the end of the flow. Aborted flows leave the staging copy behind, and the next flow judges it (live == staging).

**The five operator cases keep their outcomes.**

- Merge wrote the result → promote at the end of the flow.
- Merge preserved the user file wholesale → keep at the end of the flow.
- init actually wrote → promote at the end of the flow.
- Abort, live = render → the leftover and live are equal at the next flow → promote. It is the same outcome, deferred. That is harmless because the canonical is consumed only by the next flow's merge.
- Abort then restore → live ≠ leftover → keep.

Deferring the abort case is what makes case 4 (abort) and case 5 (abort then restore) jointly satisfiable. An abort-time decision would see a pure render and promote, then miss the restore. The derivation is sound and does not reopen D5.

**It introduces three effects.**

- **Placement hazard (N-02).** The judgement must precede the flow's first removal or rewrite of the live file. As written, it would compare the leftover with the next render.
- **Intervening edit after an abort (N-08).** Any write to the live file between the abort and the next flow flips "promote" to "discard". The deviation is fail-safe but undisclosed.
- **Version-match skip (N-09).** The judgement is postponed until a non-skipped update, which widens the N-08 window.

**Leftover staging across init is benign in both placements.**

- If init skips the file (REQ-016), a later update sees the user file ≠ leftover and discards.
- If init writes the render, the leftover is resolved first and init's own staging is then promoted at the end, so the final canonical is init's render either way.

**AC-016 mutants (focus 3).**

- **Correctly constructed and killed:**
  - M-D5a kills c2 (next flow) and c5 (next flow). The base R2 carries K, the user file lacks it, and updated == base, so K reads as a user deletion.
  - M-D5b kills c6 and c8. It also kills c3, which the table does not list; that is harmless.
  - M-D5c kills c1 and c6.
  - M-D5e kills c5.
- **Not killed by bytes:**
  - M-D5d and M-D5b-at-c4 survive a byte comparison, because R3 == R2 and the merge takes the no-op branch.
  - M-D5f is not killed at all.

  All three need N-03's R3 ≠ R2 fix.

## Other focus answers

- **GEARS / clarity of REQ-005 (focus 2).** See N-01. The English lead plus Korean body is the SPEC-wide convention and is not the defect. The defect is three conditional clauses with no GEARS modifier, which also left the judgement placement ambiguous (N-02).
- **Tier M ceiling / scope (focus 4).**
  - There are 16 REQs and 16 ACs, exactly at the ceilings (spec-workflow § SPEC Complexity Tier). The N-01 fix keeps 16.
  - The 22 mutant rows are test-side falsifiers mapped 1:1 onto D5 cases and write sites, not product scope. That is not scope creep.
  - Cost note (optional): at least 11 rows (M-05, M-06c, M-07a-e, M-08, M-08s, M-14, plus N-04's new row) need the `internal/cli` slot.
- **Capture-point feasibility (focus 5).**
  - **D7 is feasible.** For option 1, the real deployer implements `DeployWithResult` (`deployer.go:148`) and records skips (`:244`, `:252`); init needs a new transport (initializer.go:423-430). For option 2, a skipped untracked file is tracked `UserCreated` (`deployer.go:250-251`) and a tracked user file keeps `UserModified`/`UserCreated` (`:243`), so "TemplateManaged + hash equal" separates written from skipped.
  - **D8 is feasible.** `internal/cli/update/merge` imports only `internal/cli/update/plan` among update packages, and `backup` imports no `internal/cli/update/*` package (grep matched comments only). A seam in merge that uses backup helpers adds a merge→backup edge with no cycle.
  - **Write sites.**
    - template_sync: after `:364-368`, merge at `:544`, `configBackupPath` block at `:497-532`.
    - clean-reinstall: `:459` → `:507` → `:531`.
    - init: `:867`, error branch `:868-876`, bundle `:890`, sections snapshot `:1030`.

    All match plan D4.
- **Sibling amendment (focus 6).**
  - Consistent:
    - sibling plan.md:73-74 "No remedy changes which value …", which already has remedy scope;
    - acceptance.md:121 and :145 AC-UMC-011/014, which pin M1 control-cell values; iter1 read those cells as feeding `deriveTemplateBase` directly, a path this SPEC leaves unchanged;
    - the traceability row at acceptance.md:194.
  - Residual wording staleness is N-12 (optional).
  - **The D6 wording itself is not blocking.** It is a GEARS Unwanted form with a generalized subject ("A remedy … shall not …"). One optional note: its "shared key" inherits the leaf/container staleness that REQ-UMC-008/009 annotate (sibling spec.md:160), and REQ-010 carries no such annotation.

## Recommendation

FAIL at the final Tier M iteration (ceiling 2). Per the Retry Loop Contract, the orchestrator presents the operator with three options: PASS-with-debt, scope reduction, or an explicit override for a third iteration. The blocking fixes are small and textual, all in spec.md, acceptance.md and plan.md, with no scope change. In priority order:

1. **N-01 + N-02:** rewrite REQ-005 as three GEARS sentences (wording in D1). Add the "before any step removes or rewrites the live file" placement to plan D4, naming the three flow positions.
2. **N-03:** change AC-016's R3 to differ from R2 (e.g. `{"a":3,"K":1,"L":1}`), restate the c4/c5 Then values (D3), name the intermediate observation seam or drop those assertions, and add the placement-mutant row.
3. **N-04:** add a prior canonical, a user file equal to it, and a new-key render to the AC-007 template_sync cells, plus an M-06t-w row.
4. **N-05, N-06, N-07:** declare the AC-005 observation seam in M4; pin the AC-007 init tier and gates or simulate the bundle; add a promotion-failure cell and sentinel for REQ-010.
5. **N-08:** add one disclosure line in §E and plan D5.
6. **Optional:** N-09, N-10, N-11, N-12.

## Gaps (what this audit did not verify)

- No `go test`, `go build` or mutant run. Every RED prediction and mutant-kill judgement above, including the N-03 arithmetic, comes from reading `merge.go`, `strategies.go` semantics as cited, and `base.go`.
- The claim that `moai update --restore` leaves `.claude/settings.json` untouched comes from reading `restore_entry.go`, `update_restore.go` and `restore.go`, not from execution.
- For N-09, `plan.GetProjectConfigVersion` was not read. That the post-abort project version equals the package version is an inference from Clean+Deploy running before Restore Settings.
- `config.SandboxProofKind` and `toolpolicy.RenderTierPermissions` were not read (N-06 rests on `autonomy_bundle.go:59-94`).
- The sibling's `TestSharedKeyControlCells` fixture was not re-read this iteration; the AC-UMC-011/014 consistency judgement carries over the iter1 reading.
- Whether Claude Code reads a nested `.claude/settings.json` inside the cache was not verified (the SPEC's own Gap; the `claude/` subpath avoids a nested `.claude/` directory anyway).
- Cross-model second opinion: not run (`audit_model` unconfigured).

## Residual risk

- The first-cycle window (spec.md:185) and the permanent opt-out on user-deleted keys (spec.md:188) remain by design.
- Even with N-02 fixed, leftover judgement is a byte-equality test on a file other tools write. Any such write after an abort resolves toward the old base (N-08). It is fail-safe, but template changes from the aborted render will not arrive for the leaves they touched.
