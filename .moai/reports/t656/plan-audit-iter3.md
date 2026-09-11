# SPEC Review Report: SPEC-UPDATE-SETTINGS-BASE-SNAPSHOT-001

운영자 한정 연장 승인(2026-09-11) — scoped iteration 3 beyond the Tier M ceiling of 2 (`harness.plan_audit_tier_ceilings: M: 2`). Scope: (1) verify the iteration-2 blocking findings N-01..N-08 against the current text, and record the dispositions of N-09..N-11 and the N-12 deferral; (2) find new defects introduced by `git diff 04a8ab731 04eec3108` only; (3) verify the N-02 placement against the code at HEAD. Unchanged text and settled iteration-1/iteration-2 items are not re-audited.

Iteration: 3/3 (operator-approved one-time extension; no further iteration)
Verdict: PASS
Overall Score: 0.88 (harmonic mean; Tier M PASS threshold 0.80; iter1 0.73 → iter2 0.75 → iter3 0.88, no regression)

Card: t656 · Tree `/Users/goos/MoAI/moai-adk-go/.claude/worktrees/t656` · Branch `WT-update-value-merge` · HEAD `04eec3108`, working tree clean (`git status --short` empty). `git diff --stat 81c1d58f9 HEAD -- internal/` is empty, so every code citation below was read at HEAD and is identical to `81c1d58f9`. The audited diff touches exactly four files: spec.md, plan.md, acceptance.md, and progress.md of this SPEC.

Reasoning context was ignored per M1 Context Isolation. The binding decisions were treated as fixed inputs and checked for encoding only. `audit_model` is not configured, so this is a Claude-only audit.

## Must-Pass Results

- [PASS] **MP-1 REQ number consistency.** The diff adds or removes no REQ. REQ-USB-001..016 are unchanged in number; `progress.md` counts requirements 16 and ACs 16.
- [PASS] **MP-2 GEARS (requirement layer, spec.md §C).** REQ-USB-005 (spec.md:166) is now three sentences, each carrying a modifier and a `shall`:
  - Ubiquitous: "The update subsystem shall 확정본을 대기본의 승격으로만 바꾸며 … 승격하지 않는다".
  - Event-driven: "**When** 한 흐름이 정상적으로 끝나면, the update subsystem shall …".
  - Event-driven: "**When** 한 흐름이 이전 흐름이 남긴 대기본을 발견하면, the update subsystem shall …".

  No other REQ changed. The ACs belong to the verification layer and were not graded here.
- [PASS] **MP-3 frontmatter.** spec.md:2-13 carries all 12 canonical fields; `version: "0.4.0"` is quoted. `moai spec lint SPEC-UPDATE-SETTINGS-BASE-SNAPSHOT-001` printed `0 error(s), 0 warning(s)`; its only finding is an INFO `OwnershipTransitionUnmeasured`. `mcp__moai__spec_audit` (project_root = this worktree, filter = this SPEC) returned `"drift_findings":[]`, `"modern_era_clean":1`.
- [N/A] **MP-4 language neutrality.** This is a Go-internal CLI change with no multi-language tooling.
- [PASS] **MP-5 D7.** The SPEC IDs added by the diff are only `SPEC-UPDATE-MERGE-CONFLICT-BLIND-001` (in-progress) and the SPEC's own ID. Neither is retired, superseded or archived.
- [PASS] **MP-6 D8.** `grep -c syscall` returns 0 in all four artifacts.
- [PASS] **MP-7 clarification gate.** `grep -rn 'NEEDS CLARIFICATION'` over the SPEC directory exits 1. research.md is absent, as expected at Tier M.

## Category Scores

| Dimension | Score | Rubric Band | Evidence |
|---|---|---|---|
| Clarity | 0.85 | 0.75-1.0 | REQ-005 placement now has one reading (spec.md:166). The position is named in code terms at spec.md:140-143 and plan.md:100-111, and in the D4 table at plan.md:169-177. Deducted for N3-03 (preserve-path signal channel unspecified) and N3-04 (hook milestone stated twice). |
| Completeness | 0.95 | 1.0 | All sections are present. The N-08 disclosure is at spec.md:205 and plan.md:126-130. §A.4 is extended with the first-rewrite fact (spec.md:80-84). |
| Testability | 0.82 | 0.75-1.0 | c4 and c5 now discriminate, because R3 ≠ R2 (acceptance.md:226, :241-242). The wiring cells now observe order (acceptance.md:112-117). A promotion-failure cell exists (:135). Residual precision issues are N3-01, N3-02, N3-05 and N3-06, all optional. |
| Traceability | 0.90 | 0.75-1.0 | acceptance.md:20 maps REQ-005 to AC-007, whose `update_leftover_*` cells cover the placement. Every AC traces to an existing REQ. The leftover-judgement failure path is covered by argument only (N3-06). |

Harmonic mean: 4 / (1/0.85 + 1/0.95 + 1/0.82 + 1/0.90) = 4 / 4.560 ≈ 0.88.

## N-01..N-12 Verification (read against current text, not the author's disposition)

| id | iter2 class | Status | Evidence |
|---|---|---|---|
| N-01 | blocking (MP-2) | RESOLVED | spec.md:166 has three GEARS sentences (see MP-2). |
| N-02 | blocking | RESOLVED, verified against code | See § N-02 code verification. The placement text is at spec.md:140-143, plan.md:100-111 and plan D4 ① (plan.md:171, :174). The superseded placement is recorded at spec.md:37, :149 and plan.md:48 (B8). |
| N-03 | blocking | RESOLVED | R3 = `{"a":3,"K":1,"L":1}` ≠ R2 (acceptance.md:226). Observation points are defined (acceptance.md:10, :227), and the "다음 흐름 시작 판정으로 확정본 …" intermediate assertions are gone. The placement mutant row is M-D5g (:253, :284). The arithmetic is re-derived below. |
| N-04 | blocking | RESOLVED | The template_sync cells have canonical R1, user file R1, and a render R2 that adds K and changes a (acceptance.md:113-114). M-06t-w is at :266. Check: base R2 = updated, current R1 → a reads as user-changed and stays 1; K reads as a user deletion and is dropped → RED. The pre-implementation derived base (`pruneToShared` at base.go:112-131 gives `{"a":2}`) also keeps a == 1, so the RED-now label is consistent. |
| N-05 | blocking | RESOLVED | The pre-merge observation hook is declared as a package-level function variable in D8 (plan.md:183) and introduced in M4 (plan.md:217). It is cited by AC-005 (acceptance.md:81, :84-85). Residual wording issue: N3-04. |
| N-06 | blocking | RESOLVED | The `init` cell replaces the bundle through the D8 variable (plan.md:184) with a deterministic a → 9 rewrite, independent of tier and gates (acceptance.md:117). M-07a is killed: the staging would carry a:9, so canonical ≠ R2. |
| N-07 | blocking | RESOLVED | The `promote_failure` cell (acceptance.md:135) asserts: sentinel `settings-snapshot-promote-failed:` exactly once, `settings-snapshot-write-failed:` 0 times, return value nil. M-08p is at :274. A file renamed onto an existing directory fails on every OS, so the planted failure is real. |
| N-08 | blocking | RESOLVED | The disclosure is at spec.md:205 (§E), plan.md:126-130 and acceptance.md:297, with the fail-safe direction stated and "do not change silently" recorded. |
| N-09 | optional | RESOLVED | The judgement precedes both version-match skips (code verified). A `update_leftover_version_skip` cell was added (acceptance.md:116). |
| N-10 | optional | RESOLVED | The signal is the preserve path at merge.go:197-204, :217-225 and :229-237; the :207 skip is not a preserve path (spec.md:139, plan.md:94-98). M-D5i is killed by c3. Citations verified at HEAD. |
| N-11 | optional | RESOLVED | D7 prefers option 2, manifest provenance plus hash (plan.md:155-160). |
| N-12 | optional | DEFERRED (accepted) | Binding D6 freezes the sibling's REQ-UMC-010 wording, and this diff does not touch the sibling (`git diff --name-only 04a8ab731 04eec3108` lists only this SPEC's four files). The sibling's M2.1 entry gate still forces a revisit. |

## N-02 Code Verification (HEAD = 81c1d58f9 code)

**Nothing rewrites `.claude/settings.json` before the placement.** I traced `runUpdate` (update.go:139-384):

- **Before the early returns:**
  - `--version` → binary-only (:146).
  - Profile prompt (:172-192): `runProfileSetup`. Grepping `internal/cli/profile*.go` for `settings.json` returns 0 files; it writes profile yaml and settings.local.json only.
  - `--config` → `runInitWizard` returns (:196).
  - `--shell-env` and `--check` return (:204, :209).
  - `--restore` → `runUpdateRestore` returns (:269-275).
  - `checkProjectMarker`: read-only (:282).
  - `acquireUpdateLock`: writes `.moai/.update.lock` (:295).
  - `runBinaryUpdateStep` / `reexecNewBinary` (:303-325, :757-782): replaces the process (Unix) or spawns a child and exits (Windows), so the judgement runs in the new process.
- **Between the `--binary` return (:329-332) and the `--dry-run` return (:335-365) and :384:** only `os.Getwd`.
- **At :384:** `stripRetiredV2DenyEntries`, the first possible rewrite. It rewrites only when an exact retired entry is present (update_deny_migration.go:84-86). The claim at spec.md:84 holds.

**The placement precedes every destructive or skip point.**

- The v2 → clean-reinstall branch (:405, :420; its deploy at update_clean_install.go:459).
- `runTemplateSyncWithProgress` (:489), including:
  - its version skip at update_template_sync.go:640-644;
  - the second skip inside `runTemplateSyncWithReporter` at :98-107;
  - Clean (:329-337), Deploy (:364), the Backup-step read (:491) and the merge (:544).
- The only other `runCleanReinstall` caller is the dry-run planner (update.go:629), which takes `DryRun`.

**init.** Before init.go:867 there are the session-worktree chdir, the wizard, and profile preference reads. Grepping `internal/cli/wizard/` for `settings.json` returns 0 files, so none writes the project settings.json. The first write is `executor.Execute` at :867; the bundle follows at :890.

**The five operator D5 outcomes, as the new ACs fix them:**

- **c1 (merge writes → promote).** Result `{"a":2,"K":1,"u":1}`, canonical R2.
- **c2 (preserve → keep).** Canonical R1. Next flow: base R1, current `{"a":1}`, updated R3 → a = 3, K = 1, L = 1.
- **c4 (abort, no restore → promote; R3 ≠ R2).** The judgement sees live R2 == leftover R2 and promotes. The next merge has base R2, current R2, updated R3 → a = 3 (template-only), K = 1, L added, and the end canonical is R3. The same holds wired into the real flow: `update_leftover_abort` (acceptance.md:115); `update_leftover_version_skip` (:116) promotes R2 even though sync is skipped.
- **c5 (abort, then the user file reverts → keep).** Live `{"a":1}` ≠ R2, so the leftover is discarded. Base R1 → a = 3, K = 1, L = 1.
- **c6 (init writes → promote).** Canonical R2 while the live file has a = 9.

**Mutant kills.**

| Mutant | Killing cell and outcome | Result |
|---|---|---|
| M-D5a | c2 and c5, next flow: base R2 → a conflict leaves 1, K is dropped as a deletion | killed |
| M-D5b | c4 next flow (base R1 → a 1/2/3 conflicts, stays 2), c6, c8 | killed |
| M-D5c | c1, c6 (bytes ≠ render) | killed |
| M-D5d | c4 next flow (a == 2) | killed |
| M-D5e | c5 next flow | killed |
| M-D5f | c4 next flow: R3 is promoted before the merge, so base = updated → a stays 2, L is dropped | killed |
| M-D5g | c4 next flow: live R3 ≠ leftover R2 → discard → a == 2 | killed |
| M-D5i | c3: the :207 skip is read as preserve → discard | killed |
| M-D5g-w | killed, but the per-cell attribution is imprecise (N3-01) | killed |
| M-06t-w | template_sync cells | killed |
| M-08p | promote_failure | killed |

There is no M-D5h row. That is a naming gap only; nothing is missing.

**Binding-decision encode check.**

| Decision | Where encoded |
|---|---|
| A1, F-05 | spec.md:92, :97; REQ-USB-016 at spec.md:188 |
| B1 | spec.md:98 |
| C1 | spec.md:99 |
| D5, five cases | spec.md:127-135, plan.md:113-124 |
| F-17 | spec.md:154 (exact sentence), :204 |
| D6 | sibling files are untouched |
| Two-point judgement recorded as "운영자 규칙의 구현 방식, 리드 수용" | spec.md:149, plan.md:92, progress.md `D5_implementation` |
| Lane constraints (t.TempDir, userHomeDirFn, no `t.Setenv("HOME")`, no real update) | acceptance.md:6, plan.md:78 |

All are encoded faithfully.

## Defects Found (new in the 04a8ab731 → 04eec3108 diff)

D1. **N3-01** — acceptance.md:124, :285 — The M-D5g-w row lumps two variants.
  - **Backup-step variant.** At update_template_sync.go:491 the live file is still R2, because the :384 strip is a no-op on a pure render. So `update_leftover_abort` yields a == 3, not the `a == 2` the row states. This variant is killed only by `update_leftover_version_skip`.
  - **After-deploy variant.** Killed by both cells.

  The mutant is killed overall; the attribution misleads M5 observers. — Severity: minor — Class: optional — Fix: split the row into M-D5g-wb (backup step → version_skip) and M-D5g-wd (after deploy → abort `a == 2` and version_skip).

D2. **N3-02** — acceptance.md:115-116, spec.md:166, plan.md:102 — The :384 boundary itself is not behaviourally discriminated.
  - A mutant that places the judgement after `stripRetiredV2DenyEntries` but before the version skip (for example at the top of `runTemplateSyncWithProgress`) passes both `update_leftover_*` cells, because the strip is a no-op on the leftover R2.
  - Under today's `retiredV2DenyEntries` (update_deny_migration.go:20-33, v2-era entries that no post-retirement render carries), this is an equivalent mutant: no reachable outcome differs. It becomes observable only if the retired list grows.

  — Severity: minor — Class: optional (hardening) — Fix: in `update_leftover_version_skip`, give the leftover R2 and the live file one retired entry, e.g. `"Write(./secrets/**)"` under `permissions.deny`. The correct placement promotes R2; the after-strip mutant compares against the stripped file, discards, and leaves canonical R1 → RED. Add a row "M-D5g-s".

D3. **N3-03** — plan.md:151 (D3) vs plan.md:207 (M3) — D3 keeps the `MergeUserFiles` signature unchanged, while M3 has `MergeUserFiles` tell the flow seam whether it took a preserve path. No channel is named, so two engineers could pick a global flag or a wrapper returning per-file outcomes. — Severity: minor — Class: optional — Fix: name the channel (e.g. a sibling function that returns a per-path outcome, with `MergeUserFiles` kept as a wrapper), or amend D3's wording.

D4. **N3-04** — plan.md:208 vs :217, acceptance.md:81 — The pre-merge observation hook is introduced in both M3 and M4; AC-005 says M4. — Severity: minor — Class: optional — Fix: state it in one milestone only. M3 fits, because AC-016 and AC-006 run on the seam.

D5. **N3-05** — acceptance.md:232, :235 — The next-flow cells of c2 and c5 are labelled "guard" but now assert `a == 3`. §C (acceptance.md:43) defines guard as "passes before implementation". On the pre-implementation merge, the derived base is `{"a":3}` (base.go:98, :112-131), so a stays 1: the cell is not green before implementation. The label is true only against an empty-promotion stub with base selection already in place. — Severity: minor — Class: optional — Fix: relabel these as green against the stub and falsified by M-D5a / M-D5e, or add that sense to §C.

D6. **N3-06** — acceptance.md:135-137 — The leftover-judgement call sites in `runUpdate` (before :384) and `runInit` (before :867) are new with this diff. A mutant that turns a leftover-promotion failure at those call sites into a returned error survives: `promote_failure` exercises only the seam's end-of-flow promotion. It is covered only by the same-primitive argument already accepted for init and template_sync. — Severity: minor — Class: optional — Fix: none required, or add a `helper`-style cell that plants a directory at the canonical path with a leftover present and asserts `runUpdate` returns nil and prints `settings-snapshot-promote-failed:` exactly once.

No blocking or must-pass defects were found.

## Regression Check

All eight iteration-2 blocking findings (N-01..N-08) are RESOLVED with text evidence, and N-02 is additionally verified against code (tables above). N-09..N-11 are resolved; N-12 is deferred under binding D6. No defect has persisted unchanged across all three iterations.

## Recommendation

PASS at 0.88, above the Tier M threshold of 0.80, with every must-pass criterion passing (MP-4 N/A). The six new findings are minor and optional. N3-01 and N3-02 are one-row table edits worth folding into the run-phase M5 plan. None requires another plan iteration. The mandatory Implementation Kickoff Approval gate is unaffected by this verdict.

## Gaps (what this audit did not verify)

- No `go test`, `go build`, `make` or mutant run. Every RED and kill judgement comes from reading merge.go:173-258, base.go:69-131, update_deny_migration.go:20-100, and applying the shared engine's 3-way rules as the SPEC describes them. `internal/merge/strategies.go` was not re-read this iteration.
- `runInitWizard` in reconfigure mode (`moai update -c`) was not traced into the packages it calls. It returns before the judgement, so if it writes the project settings.json it is an N-08-class intervening write that ends in discard.
- `runProfileSetup` was checked by grep only: no `settings.json` token in `internal/cli/profile*.go` or `internal/profile/`.
- `restore.go` was not read. Whether `RestoreMoaiConfig` writes `.claude/settings.json` remains the run-phase M1 item.
- Whether `runUpdate` can be driven in-process for the `update_leftover_*` cells (cwd via chdir, the `newTemplateSyncDeployer` seam, `--yes`, no network in the binary step) is a run-phase harness question and was not proven here.

## Residual risk

- Byte-equality leftover judgement remains exposed to any write between an abort and the next flow (N-08, disclosed, fail-safe).
- If `retiredV2DenyEntries` ever grows to cover entries a then-current render carried, the :384 placement becomes load-bearing. Without N3-02's cell, a misplaced judgement would then go unnoticed.
