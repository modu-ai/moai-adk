---
id: SPEC-UPDATE-DATA-SURVIVAL-001
title: "moai update — user-data survival: on-disk backup before every destructive step, a failure contract with an escape from the marker-gate lockout, and non-vacuous safety guards"
version: "0.5.2"
status: completed
created: 2026-07-31
updated: 2026-08-01
author: manager-spec
priority: P0
phase: "v3.0.2"
module: "internal/cli"
lifecycle: spec-anchored
era: V3R6
tier: M
tags: "cli, update, backup, rollback, data-loss, crash-window, safety-test, mutation-testing, regression"
related_specs: [SPEC-UPDATE-REINSTALL-LOOP-002, SPEC-UPDATE-YAML-PRESERVE-001, SPEC-V3R6-UPDATE-NAMESPACE-PROTECT-001, SPEC-V3R6-UPDATE-ARCHIVE-CONTRACT-001]
depends_on: [SPEC-UPDATE-REINSTALL-LOOP-002]
---

# SPEC-UPDATE-DATA-SURVIVAL-001

## HISTORY

| Version | Date | Change |
|---------|------|--------|
| 0.5.2 | 2026-08-01 | **AC-UDS-019 (a) re-anchored on the merge-base — removes a false NFR-UDS-003 failure caused by branch topology, not by this SPEC.** `manager-develop` returned a blocker: the fixed-SHA form `git diff --name-only 8cc108ddb..HEAD -- internal/template/templates/` returned `8`, reading as a template-neutrality violation. Root cause: foreign commit `9ced435e9` (PR #1266, an agent-body fix) landed 8 template files on `origin/main` after the `8cc108ddb` pin was recorded, and entered this branch via merge `2255165f5`. This SPEC's own commits touch **zero** template files (`git diff --name-only origin/main..HEAD -- internal/template/templates/` → `0`). The pinned command was measuring branch topology. **Re-pinning to `9ced435e9` was rejected** — `origin/main` advances and this branch merges again before the PR closes, so any fixed SHA breaks on the next merge (this is the third stale-absolute-anchor failure this session: coordinates twice, now the pin). (a) is therefore re-anchored on `$(git merge-base origin/main HEAD)`, which is invariant under merges. **Three topologies verified empirically before adoption** (not accepted on reasoning alone): case 1 — after a merge, merge-base *is* `origin/main`, so foreign changes fall at-or-below the diff base and vanish (verified on this tree: `8` → `0`); case 2 — `origin/main` advanced without a merge, merge-base stays behind but the foreign commits are not in HEAD either, so the count is still `0` (verified in a disposable synthetic repo reproducing the topology); case 3 — a template file **this SPEC** modifies sits above the base on either side and is **still caught** (`≥ 1`, verified in both topologies including through a *conflicting* merge where the SPEC's edit survived resolution). The AC's expected value is unchanged (`0`/`0`), and its existing rationale — that (a) is load-bearing because working-tree-only `git status --porcelain` misses the committed case — is preserved verbatim; only the choice of diff base changed. A known limitation is stated rather than hidden: once this SPEC's own PR merges, the merge-base advances and (a) returns `0` for its commits too — correct for a pre-merge AC, which is not a post-merge audit. Also updated: `acceptance.md` §A.4 now draws an explicit **measurement-anchor vs comparison-base** distinction (`8cc108ddb` remains a valid anchor — re-verified `ANCESTOR: yes` on HEAD `184a5bd222` — but is NOT a comparison base, and the stale `= origin/main` parenthetical is removed); the §A.4 time-of-check-to-time-of-use paragraph, which anticipated exactly this event, now records that it **occurred twice** (M1-caused coordinate drift + the foreign template commit) with the general lesson that a check *evaluated later* needs a base computed at evaluation time while a figure that merely *records where it was observed* may keep a literal SHA; §A.3's AC-UDS-019 preservation-guard row is scoped to this SPEC's own commits; and `plan.md` §A drops the same false `= origin/main` parenthetical. Documentation only — no requirement text, no `.go` file, no template file, and no AC expected value changed. |
| 0.5.1 | 2026-08-01 | **M2 Step 0 coordinate re-baseline — documentation-only, no requirement or AC semantics changed.** `plan.md` §F M2 Step 0 requires re-running the §C.0 destructive-call-site source scan before encoding the registry, and requires *reporting* any divergence rather than silently reconciling it. The scan was run on HEAD `2255165f5`: **site total (18), pair count (11), and the (file, function) pair set are all unchanged** — no destructive site was added or removed. Only line coordinates moved, and the cause is **this SPEC's own M1 commit `4ddd35120`** (recovery manifest + restore entry point), NOT a second foreign-SPEC TOCTOU like the E1-caused v0.3.0→v0.4.0 event. Both artifacts now narrate the two drift events as distinguishable so a later reader cannot conflate them. Corrected in `plan.md`: §C.0 row 7 (`:307`→`:315`) and row 9 (`:843`→`:853`); §C.1 rows 1/3 (`update_template_sync.go` `:294`/`:388`→`:300`/`:397`, `:292`/`:381`→`:298`/`:388`), row 12 (`:307`→`:315`), row 13 (`:843`→`:853`); §B.3 marker gate (`update.go:236` → the `checkProjectMarker(cwd)` call at `update.go:251`, predicate extracted by M1 into `update_restore.go:21`/`:27`); §D `userHomeDir()` (`:833`→`:843`) with an in-place note that `:843` is the `userHomeDir()` call and NOT the `os.RemoveAll` site (now `:853`), since the pre-M1 tables recorded the latter at `:843`. Corrected in `acceptance.md`: AC-UDS-005 table rows 7/9 plus a baseline-header re-measurement note and a drift-history paragraph attributing the second shift to `4ddd35120`; AC-UDS-003 (marker-gate baseline); AC-UDS-004 (`RestoreMoaiConfig` callers `:429`/`:397`→`:437`/`:406`); AC-UDS-008 (`:294`/`:388`→`:300`/`:397`); AC-UDS-009 (`:294`/`:388`→`:300`/`:397`, `:348`/`:351`→`:356`/`:359`); AC-UDS-011 (`:832`/`:833`/`:841-843`→`:842`/`:843`/`:851-853`); AC-UDS-013's retracted-baseline note extended with the new value. **No AC verification command or expected value was changed** — AC-UDS-005's guard keys on (file, enclosing function, occurrence count) and never on line numbers. Verified-correct and left untouched: every `dirs.go` coordinate in AC-UDS-007, `mergeBackPreserveInventory` `:400` with failure returns `:416`/`:420`/`:424` (AC-UDS-016 / plan.md §C.2), all `deploy.go` / `backup.go` / `update_archive.go` / `update_cleanup.go` / `update_namespace_protect.go` / `update_residue_cleanup.go` / `glm_tools.go` / `homedir.go` coordinates, and `update_clean_install.go:137`. **Sibling-artifact closure (same version, follow-up round):** the `spec.md` §A stale coordinates first recorded here as a known gap are now CLOSED in this same `0.5.1` entry — leaving them stale would have been the sibling-artifact failure mode (the fix merely relocating to an unedited file). Corrected in `spec.md`: §A Defect 1 (`update_template_sync.go` `:294`/`:388`→`:300`/`:397` — the `mergeableBackups` declaration + append, i.e. the §C.1 row-1 mergeable set, NOT the row-3 `.gitignore` pair; `MergeUserFiles` `:427`→`:436`; `update_clean_install.go` `:348`/`:351`→`:356`/`:359`); §A Defect 3 (`userHomeDir()` `:833`→`:843`, plus `globalHooksDir` `:841-843`→`:851-853`, with the in-place note that `:843` is the `userHomeDir()` call and NOT the `os.RemoveAll` site now at `:853`); §A Defect 5 (marker check `update.go:236` → the `checkProjectMarker(cwd)` call at `update.go:251`, predicate extracted by M1 into `update_restore.go:21`); §D.3 REQ-UDS-013 (`update.go:833`→`:843` with the same conflation note). No requirement text, no AC command, no AC expected value, and no `§A` defect *finding* changed — only the coordinates they cite. All six coordinator-measured values plus the `globalHooksDir` range were independently re-confirmed against the tree before editing. **A within-file sibling sweep then found the same stale tokens in two further `spec.md` locations the 6-row table did not name** — the §A Defect 3 code-block annotations (`// update.go:841`/`:842`/`:843` → `:851`/`:852`/`:853`) and the §H coordinate inventory (`update_template_sync.go` `gitignoreBackup` `:292`/`:381`→`:298`/`:388`, `mergeableBackups` `:294`/`:388`→`:300`/`:397`, `Restore Settings` `:391`→`:400`, `MergeUserFiles` `:427`→`:436`; `update_clean_install.go` removal loop `:305-311`→`:313-318`, mergeable set `:348`/`:351`→`:356`/`:359`, `BackupMoaiConfig` call `:340`→`:348`; `update.go` marker gate `:236`/`:240`→`:251` + `update_restore.go:21`/`:27`, `ensureGlobalSettingsEnv` `:832`→`:842`, `userHomeDir()` `:833`→`:843`, `globalHooksDir` `:841-843`→`:851-853`). The §H header is restated on HEAD `2255165f5` and now records three drift events (iteration-1, E1-external, M1-self-inflicted) rather than two. Two further §A prose hits were caught by the same sweep and corrected: §A Defect 1's `gitignoreBackup` pair (`update_template_sync.go:292`/`:381`→`:298`/`:388`) and §A Defect 5's `"Restore Settings"` case (`update_template_sync.go:391`→`:400`). Verified-correct and left untouched in the same neighbourhood: `CleanMoaiManagedPaths` invocation (`:243-247`), `Deploy Templates` (`:249`), `Clean Managed Paths` step (`:243`), Step 4 marker (`update_clean_install.go:251`). Every other §H entry — `deploy.go`, `update_residue_cleanup.go`, `update_preserve_inventory.go`, `update_namespace_protect.go`, `backup.go`, `update_archive.go`, `update_cleanup.go`, `homedir.go`, `glm_tools.go`, `update_safety_test.go`, `dirs.go` — was re-measured on `2255165f5` and is unchanged, so it was left untouched. `spec.md`, `plan.md`, and `acceptance.md` now agree on every shared coordinate. |
| 0.5.0 | 2026-08-01 | plan-audit iteration 4 (**PASS, 0.892** — harmonic 0.89173 / arithmetic 0.8925, Tier M threshold 0.80, **0 blocking defects**) advisory round, D23-D26. The re-baseline converged: iteration 4 re-derived the registry from Go source with an enclosing-function scan (18 sites / 11 pairs, every row's file + function + line confirmed, closing iteration 3's Gap 7), **falsified** the replacement re-derivation commands by injecting a 19th destructive site into a scratch copy (both moved 18→19 and 11→12, while the retired self-referential form stayed pinned), and re-ran **every** §B baseline against the tree — **0 drift, 19 of 19 reproduce** — disconfirming iteration 3's worry that D17's blast radius was only a lower bound (Gap 8 closed). D1-D7 were re-measured against the moved tree and none regressed (Gap 1 closed). Four minor findings opened and applied in this round: D23 — NFR-UDS-004 was silently absent from the §B.0 coverage map while §A.5 asserted completeness; it is now explicitly declared toolchain-enforced (`gofmt` / `go vet` / `golangci-lint`) rather than AC-covered. D24 — AC-UDS-004's one-line baseline ("no restore entry point exists") understated the existing surface; `backup.RestoreMoaiConfig` (`backup/restore.go:40`, mid-run only, live callers `update_clean_install.go:429` and `update_template_sync.go:397`) does restore from a backup directory, so the baseline is restated as "no entry point with marker-gate bypass + idempotency + foreign-dir refusal", with an M1 obligation to state whether the new entry point extends it or is separate — closing the two-owners hazard for an implementer reading `acceptance.md` alone. D25 — §A.4 named worktree HEAD `89b2e4772` while HEAD had advanced; the anchor is restated on the code baseline `8cc108ddb` with the self-referential-commit hazard documented (the intervening delta is SPEC-artifacts-only, so no coordinate is invalidated). D26 — `plan.md`'s version-header disposition is now recorded as "intentionally none" (`spec.md` frontmatter + HISTORY is the single version record). Informational: row 11's description corrected — `runV3ResidueCleanup` removes `sweep`, the existence-refiltered subset (`update_residue_cleanup.go:95-100`), not the raw `scanDeprecatedPaths` return; the classification conclusion is unaffected. D8 remains deferred and must be resolved in M5 before AC-UDS-015 can be marked PASS. Two run-phase conditions carry forward literally: M2's first act re-runs the §C.0 source scan (E3-E6 share this code baseline, so time-of-check-to-time-of-use recurs), and AC-UDS-005 is marked PASS only after an observed `--- FAIL` against an injected unregistered site. |
| 0.4.0 | 2026-08-01 | plan-audit iteration 3 (FAIL, 0.758) **re-baseline round** — user-approved override of the 3-iteration ceiling. The score regression was external, not a revision failure: E1 (`SPEC-UPDATE-REINSTALL-LOOP-002`) landed its run (`beeb0ebc2`, PR #1261) and sync (`8cc108ddb`, PR #1264) between rounds, so the plan described a tree that no longer existed. D17: the code baseline is re-anchored from the unreachable `a8b42e112`/`d5336214e` to HEAD `89b2e4772` / `8cc108ddb` (verified an ancestor); §A.4's "no Go source differs" premise was measurably false (19 files) and is replaced with the measured re-baseline. The M2 registry grows 17→**18 sites** / 10→**11 pairs** with a new row for `update_residue_cleanup.go:135` (`runV3ResidueCleanup`, classified user-data with E1's REQ-RIL2-019 cross-SPEC backup assignment), and rows 3/4/7/9 take re-measured coordinates. The two "independent re-derivation" commands that grepped `acceptance.md` itself — a §A.3 shape-(a) self-comparison, and the reason the 17→18 drift produced no signal — are replaced with Go-source scans. D18: §H and the §A Defect 3 body re-anchored on `update.go:832`/`:833`/`:841-843` (was `:755`/`:756`/`:764-766`); `mergeBackPreserveInventory` `:330`→`:400` with failure returns `:346`/`:350`/`:354`→`:416`/`:420`/`:424`; `preserveInventoryRoots` `:66-70`→`:68-72`. D19: AC-UDS-001's rollback clause was vacuously satisfiable on an empty removed-set (adding the forbidden rollback could not move it — vacuous by §A.3's own criterion). It is now a four-clause assertion pinned to a literal `plantedMoaiManagedPaths` fixture whose non-emptiness is verified from the test source, never from the checked function's own output; the fixture pin is a `plan.md` M1 obligation. D20: `acceptance.md` version header bumped. D21: AC-UDS-014's coverage transcript restored to its verbatim four lines (`MigrateLegacyMemoryDir` was dropped). D22: the carve-out falsifier corrected `2 → 0` → `2 → 1`, with (b)'s wider binding to the whole Category D banner documented rather than narrowed. D8 remains deferred. |
| 0.3.0 | 2026-08-01 | plan-audit iteration 2 (PASS, 0.84) delta round, D11-D16 + progress backfill. D13: REQ-UDS-020's coverage made real — AC-UDS-001 gains clause 3 (the destroyed paths are still absent on return), replacing a §B.0 map entry that claimed an assertion the AC body did not make. D14: §A.3 gains a preservation-guard carve-out (vacuity is "no change could move it", not "constant"), classifying AC-UDS-007(b) / 010 / 018 / 019 as preservation guards and reconciling the clause with §A.2's `-run` scoping. D15: AC-UDS-007 (a)'s `sed` range re-anchored on content (`.moai/project/brand`) and REQ-UDS-009 gains a placement requirement pinning the SPEC-ID to that window — the audit's suggested backward-widening was rejected because a lookback captures the unrelated `SPEC-V3R6-V2-V3-CLEAN-REINSTALL-001` mention at `dirs.go:302` and would turn the `0` baseline into a false `1`. D11 / D12 / D16 were re-measured and found already applied in v0.2.0 (the report was written against an earlier artifact state); D10 is resolved by fact — E1 reached `status: completed`. D8 / D9 remain deferred. Divergence table in `progress.md` §E.1. |
| 0.2.0 | 2026-07-31 | plan-audit iteration 1 (FAIL, 0.71) delta revision, D1-D7. D2: REQ-UDS-007 rewritten to require independent static-scan enumeration (the prior wording permitted a `count == count` self-comparison). D3: REQ-UDS-013 gains the `userHomeDirFn` seam requirement; NFR-UDS-002's "`userHomeDir` indirection already exists" claim corrected — no seam existed. D4: Defect 2's Category D misattribution retracted; REQ-UDS-009 / AC-UDS-007 restated against the real gap. D1: §H `update.go` coordinates corrected to the post-merge values (`:755`, `:764-766`). D5/D6: AC-UDS-019 / AC-UDS-014 de-vacuified. D7: AC-UDS-020 added for the previously uncovered REQ-UDS-003; every AC now cites its REQ. D8/D9/D10 deferred to the next audit round. |
| 0.1.0 | 2026-07-31 | Initial draft. Four-lens audit of `moai update` / `.moai/config`; Epic SPEC 2 of 6. |

## §A Problem / Motivation

`moai update` destroys and redeploys large parts of a user's project tree. The destruction is
deliberate; the problem is that several destructive steps have **no artifact on disk** to restore
from, and there is **no defined behaviour** when a run dies between destruction and restoration.

Six defects, each observed against this tree.

### Defect 1 — `.claude/settings.json` is destroyed with only an in-memory backup

The normal path's first destructive step is `deploy.CleanMoaiManagedPaths`
(`internal/cli/update/deploy/deploy.go:28`), whose first `cleanTarget` entry is
`.claude/settings.json` (`deploy.go:40-42`, removed in the loop at `deploy.go:105`). Its only
backup is a `[]byte` captured into an `updatemerge.FileBackup` slice (`mergeableBackups`) at
`internal/cli/update_template_sync.go:300`/`:397` and re-applied by `MergeUserFiles` at
`update_template_sync.go:436`. The clean-reinstall path builds the identical in-memory slice at
`internal/cli/update_clean_install.go:356`/`:359`. (Coordinates re-measured on HEAD `2255165f5`;
recorded pre-M1 as `:294`/`:388`, `:427`, and `:348`/`:351`. The `:300`/`:397` pair is the
`mergeableBackups` **declaration** and its **append** — the §C.1 row-1 mergeable set, NOT the
row-3 `.gitignore` pair, which is `gitignoreBackup` at `:298` and `gitignorePath` at `:388`.)

Observed after running the two backup steps followed by the clean step on a fixture tree:

```
live .claude/settings.json              exists=false
backup artifact: .moai-backups/<ts>/sections/user.yaml
backup artifact: .moai-backups/<ts>/backup_metadata.json
   (plus .template-defaults/sections/*.yaml — no settings.json entry)
```

`BackupMoaiConfig` (`internal/cli/update/backup/backup.go:27`) walks `.moai/config` only
(`backup.go:28`), so it cannot cover a `.claude/` path by construction. A grep of the backup
package for any disk write of `settings.json` returns nothing.

Between the `RemoveAll` and the merge-back, the user's `permissions.allow`, `env`, and
`outputStyle` exist **only in the process's heap**. A crash, an ENOSPC, a SIGKILL, or Ctrl+C in
that window loses them permanently with no recovery path.

Two further files share the identical defect shape and the identical window:
`.moai/status_line.sh` (same `mergeableBackups` slice) and `.gitignore` (the `gitignoreBackup`
`[]byte` at `update_template_sync.go:298`/`:388` — re-measured on HEAD `2255165f5`; recorded pre-M1
as `:292`/`:381`).

### Defect 2 — destructive targets that appear in no protection set

Three protection sets exist:

| Set | Covers | Defined at |
|---|---|---|
| PRESERVE inventory | `.moai/specs`, `.moai/project`, `.claude/commands` | `update_preserve_inventory.go:68-72` |
| `BackupMoaiConfig` | `.moai/config` only | `backup/backup.go:27`/`:28` |
| Namespace backup | `.claude/skills`, `.claude/agents`, `.moai/harness` | `update_namespace_protect.go:39-43` |

`MigrateLegacyMemoryDir` (`deploy.go:143`) removes `.moai/memory/` outright when both the legacy
and the new directory exist (`deploy.go:176`). That path is in none of the three sets and has no
backup of any kind.

`.moai/db` is a **`defs.DeprecatedPaths` entry** (`internal/defs/dirs.go:313`, under the
`brand + db directories` grouping) rather than a bespoke removal, so it is deleted by the generic
deprecated-path loop at `update_clean_install.go:307` (and, since E1, by the residue sweep at
`update_residue_cleanup.go:135`). It is likewise in none of the three
protection sets — but its backup is already required by `SPEC-UPDATE-REINSTALL-LOOP-002`
REQ-RIL2-015 (backup before every deprecated-path deletion), so this SPEC depends on that
requirement rather than restating it.

Separately, `.moai/db` sits under the bare group comment `// brand + db directories`
(`dirs.go:305`), which names neither the SPEC that authorises its removal nor its protection status.
Its sibling groupings carry full explanatory banners — Category C (`dirs.go:318-324`) and Category D
(`dirs.go:344-351`) each state what the group covers, which SPEC deprecated it, and why. A reader
auditing db-removal coverage therefore finds the least explanation at the one entry that is
genuinely unprotected.

> **Retracted finding.** An earlier draft of this SPEC asserted that the Category D comment
> (`dirs.go:344-351`) misnames its registered path — describing `.moai/project/db/` while the
> registered entry is `.moai/db` (`dirs.go:313`). Re-reading `dirs.go:305-358` refutes this.
> `.moai/db` belongs to the `// brand + db directories` group, **not** Category D. Category D's
> sole registered entry is `.moai/config/sections/db.yaml`, which the comment names exactly; its
> `.moai/project/db/` mention is a deliberate contrast clause ("Unlike the `.moai/project/db/`
> docs (a preserve root, manual deletion per CHANGELOG), db.yaml is NOT under a preserve root")
> stating what is **not** removed. Comment and code agree. The retracted claim would have required
> deleting an accurate explanatory sentence; REQ-UDS-009 and AC-UDS-007 are restated below against
> the real, much weaker gap.

### Defect 3 — the deletion radius of `~/.claude/hooks/moai` is unpinned

`ensureGlobalSettingsEnv` (`internal/cli/update.go:842`) removes a directory inside the user's real
HOME:

```go
globalHooksDir := filepath.Join(homeDir, defs.ClaudeDir, "hooks", "moai")   // update.go:851
if _, err := os.Stat(globalHooksDir); err == nil {                          // update.go:852
    _ = os.RemoveAll(globalHooksDir)                                        // update.go:853
}
```

A mutation lens widened the target from `~/.claude/hooks/moai` to `~/.claude/hooks` — deleting the
user's non-MoAI global hooks — and ran the suite through an overlay:

```
$ go test -overlay=overlay5.json ./internal/cli/ -run 'GlobalSettings|Update|Hook'
ok  github.com/modu-ai/moai-adk/internal/cli  18.305s
```

Everything passed. Re-verified on HEAD `89b2e4772`: `grep -rn 'globalHooksDir' internal/cli/ --include='*_test.go'`
returns **0** lines — the identifier appears only at `update.go:851-853` (re-measured on HEAD
`2255165f5`; recorded pre-M1 as `:841-843`). This deletion is outside
the project root, unbacked, and unrecoverable, and nothing detects a widening of its radius.

Worse, a guard cannot currently be written at all. `ensureGlobalSettingsEnv` resolves HOME by
calling `userHomeDir()` (`update.go:843` — re-measured on HEAD `2255165f5`; recorded pre-M1 as
`:833`. Note `update.go:843` is the `userHomeDir()` call and is **NOT** the `os.RemoveAll` site,
which is now `update.go:853`; the pre-M1 tables recorded that `os.RemoveAll` site at `:843`, so the
two values can be confused when comparing old and new records), which is a plain function (`homedir.go:14`) with no
injection point. The package does own an injectable seam — `var userHomeDirFn = userHomeDir`
(`glm_tools.go:123`), overridden by tests at `glm_tools_test.go:34-36` — but
`ensureGlobalSettingsEnv` does not use it. `userHomeDir` does read `$HOME` first
(`homedir.go:15`), so `t.Setenv("HOME", …)` would technically redirect it — but NFR-UDS-002 forbids
that, because a process-wide HOME mutation is visible to every parallel test in the package. Under
that constraint the radius guard has no redirection point and is unimplementable until the call is
routed through the seam (REQ-UDS-013). The prohibition is a deliberate isolation policy, not a
technical impossibility, and REQ-UDS-013 is what makes the policy satisfiable.

### Defect 4 — `TestMoaiUpdate_PreservesUserArea` executes zero production code

`internal/cli/update_safety_test.go:60` claims to enforce that `moai update` does not touch user
areas. It calls `simulateMoaiUpdate` (`update_safety_test.go:49-58`, invoked at `:95`), a fake defined in the same
test file whose entire body writes one managed file under `.claude/skills/moai/` and comments
`// Note: we intentionally do NOT touch user areas.` The assertions at `update_safety_test.go:91`/`:99-100`
compare pre/post hashes of three user directories that the fake never visits, so they hold by
construction.

Re-verified on this tree: a grep of that file for any production symbol
(`runUpdate|RunUpdate|CleanMoaiManagedPaths|BackupMoaiConfig|runCleanReinstall|buildPreserveInventory|deploy\.|backup\.`)
returns no matches. The `CLAUDE.local.md` §24 contract — `moai update` never deletes user-authored
harness agents or `hns-*` skills — therefore has no executable defence.

### Defect 5 — no failure contract, and the repair tool locks itself out

Normal path: `CleanMoaiManagedPaths` (invoked at `update_template_sync.go:243-247`) removes
`.moai/config` wholesale (`deploy.go:121`); when the subsequent `Deploy Templates` step
(`update_template_sync.go:249`) fails, the function returns and the `"Restore Settings"` case at
`update_template_sync.go:400` (re-measured on HEAD `2255165f5`; recorded pre-M1 as `:391`) is never
reached. Clean path: a Step 5 failure returns before
Steps 5.5 and 6.

The backups survive on disk, but nothing tells the user which backup belongs to the failed run,
and no command applies one — the only restore entry points are `backup.RestoreMoaiConfig`
(`backup/restore.go:40`, called mid-run only) and `moai migrate restore-skill`
(`migrate_restore_skill.go:90`, a different subsystem).

Worse, the failure is self-sealing. Once `.moai/config/sections/system.yaml` is gone, the project
marker check — the `checkProjectMarker(cwd)` call at `internal/cli/update.go:251`, whose predicate
M1 extracted into `internal/cli/update_restore.go:21` — rejects the next invocation
(coordinate re-measured on HEAD `2255165f5`; recorded pre-M1 as `update.go:236`, which no longer
resolves because M1 both moved the call site and extracted the predicate):

```
not a moai project: .moai/config/sections/system.yaml not found in the current directory
```

The tool that would repair the damage refuses to run on the damaged tree.

### Defect 6 — `mergeBackPreserveInventory` partial-restore branches are uncovered

`mergeBackPreserveInventory` (`internal/cli/update_preserve_inventory.go:400`) restores the PRESERVE
inventory file-by-file. Its three failure returns — the non-`ErrNotExist` stat error
(`:416`), the `MkdirAll` error (`:420`), and the `copyFile` error (`:424`) — each return
immediately from inside the loop. A failure at file *k* leaves files `0..k-1` restored, files
`k..n` absent, and no report of where the boundary fell. The measured statement coverage of the
function is recorded in `plan.md §C`; the three error branches are the uncovered portion.

## §B Goals

- No destructive step runs before the data it destroys exists on disk outside the process.
- Every destructive target is either covered by a named protection set or explicitly and
  deliberately exempt, with the exemption recorded.
- The HOME-scoped deletion cannot widen without a test failing.
- The user-area safety contract is defended by a test that drives real production code.
- A run that dies mid-flight leaves the user a named backup, a stated tree state, and a command
  that restores it — and that command works on a tree whose project marker was destroyed.
- A partial restore reports precisely which files were restored and which were not.

## §C Scope Exclusions

### Out of Scope — v2/v3 detection, PRESERVE∩Deprecated, dry-run

- Semver-normalized v2/v3 version detection, the `preserveInventoryRoots` ∩ `defs.DeprecatedPaths`
  exclusion, the residue-cleanup path, and `--dry-run` reachability. All owned by
  `SPEC-UPDATE-REINSTALL-LOOP-002`.
- **Backup-before-delete for `defs.DeprecatedPaths` entries** (including `.moai/db`) is owned by
  that SPEC's REQ-RIL2-015/016. This SPEC declares `depends_on: [SPEC-UPDATE-REINSTALL-LOOP-002]`
  and treats deprecated-path backup as an inherited precondition, never re-specifying it. The
  dependency is load-bearing in the reverse direction too: E1's PRESERVE-exclusion fix removes the
  incidental protection that `.claude/commands/agency/*.md` enjoys today, which is why E1 binds
  `backupDeprecatedPaths` to its own milestone.

### Out of Scope — YAML merge fidelity

- Comment, key-order, and scalar-style preservation across `.moai/config/sections/*.yaml` merges,
  and the old-only-key drop. Owned by `SPEC-UPDATE-YAML-PRESERVE-001`. This SPEC cares that the
  bytes reach disk before deletion, not how they are later merged.

### Out of Scope — merge tiers, dead keys, CI gates, doc drift

- Merge-tier semantics, dead `.moai/config` keys, CI gate additions, and documentation drift.
  Owned by Epic SPECs 3 through 6.

### Out of Scope — DeprecatedPaths catalogue membership

- Adding, removing, or re-categorising entries in `defs.DeprecatedPaths`. This SPEC amends only the
  `// brand + db directories` **group comment prose** so it names its authorising SPEC and
  protection status (REQ-UDS-009); the slice body is untouched and its count invariant
  (`internal/defs/dirs_test.go`) is preserved. The Category D comment is **not** amended — it is
  accurate, and AC-UDS-007 guards it against deletion.

### Out of Scope — template content

- Any edit under `internal/template/templates/**`. Fixtures live in `_test.go` files only.

### Out of Scope — transactional filesystem rollback

- An automatic, atomic, whole-tree rollback of a failed update. §D.5 records the decision to
  provide an explicitly-invoked restore command instead, with rationale.

## §D Requirements (GEARS)

### D.1 On-disk backup before destruction

- **REQ-UDS-001** — **When** the update reaches its first destructive step on either path, every
  file whose sole backup is currently in-process memory (`.claude/settings.json`,
  `.moai/status_line.sh`, `.gitignore`) shall already exist as a byte-identical copy under a
  backup directory on disk.
- **REQ-UDS-002** — The on-disk backup shall be written under a single run-scoped backup directory
  whose absolute path is reported to the caller, so a later restore can name it.
- **REQ-UDS-003** — **When** writing any of those on-disk backups fails, the update shall abort
  before the first destructive step and shall surface a grep-able sentinel naming the file that
  could not be backed up.
- **REQ-UDS-004** — The in-memory merge-back path shall be retained unchanged; the on-disk copy is
  additive insurance for the crash window, not a replacement for the merge.
- **REQ-UDS-005** — The normal path and the clean-reinstall path shall use the same on-disk backup
  mechanism, so neither path can drift into being the unprotected one.

### D.2 Destructive-target coverage

- **REQ-UDS-006** — Every destructive filesystem operation in the update subsystem shall be
  enumerated in a single in-repo registry naming, per target, the protection set that covers it or
  the recorded reason it is exempt.
- **REQ-UDS-007** — A regression guard shall enumerate destructive call sites **independently of
  the registry**, by statically scanning the update subsystem's Go sources
  (`internal/cli/update/**` and `internal/cli/update*.go`, excluding `_test.go`) for `os.RemoveAll(`
  and `os.Rename(` call sites, and shall fail when the scanned multiset of
  (file, enclosing function, occurrence count) does not equal the registry's recorded multiset.
  The registry shall NOT be the guard's enumeration source: a guard that counts registry rows
  against registry rows is a `count == count` self-comparison that can never fail, and is rejected.
  Each registry row shall additionally carry either a protection-set assignment or a recorded
  exemption reason; a row carrying neither shall fail the guard.
- **REQ-UDS-008** — **When** `MigrateLegacyMemoryDir` removes `.moai/memory/` because both the
  legacy and the new directory exist, it shall back up `.moai/memory/` before removal.
- **REQ-UDS-009** — The `defs.DeprecatedPaths` group comment covering `.moai/db` shall name the
  SPEC that authorises the group's removal and the group's protection status, matching the
  explanatory form the sibling Category C and Category D banners already use. The SPEC-ID shall
  appear **on the `// brand + db directories` line or between that line and the group's first
  entry** — not above the heading. This placement is a requirement, not a stylistic note:
  `dirs.go:302` already carries an unrelated `SPEC-V3R6-V2-V3-CLEAN-REINSTALL-001` mention (a note
  recording that a *different* entry was reversed), so an assertion window wide enough to admit a
  banner placed above the heading would also admit that pre-existing line and pass before this
  requirement is implemented. AC-UDS-007 (a) asserts over exactly the pinned region. The Category D
  comment's existing `.moai/project/db/` contrast clause is accurate and shall be preserved
  verbatim — see the retracted finding in §A Defect 2.
- **REQ-UDS-010** — The registry shall record `defs.DeprecatedPaths` deletion as covered by
  `SPEC-UPDATE-REINSTALL-LOOP-002` REQ-RIL2-015 rather than duplicating that coverage.

### D.3 Deletion-radius pinning

- **REQ-UDS-011** — The HOME-scoped removal in `ensureGlobalSettingsEnv` shall be covered by a test
  that fails when the removal target is widened to any ancestor of `~/.claude/hooks/moai`.
- **REQ-UDS-012** — **While** that test runs, it shall not read from or write to the operator's real
  home directory.
- **REQ-UDS-013** — The removal target shall be derived from a single named symbol so the guard has
  one thing to assert against, rather than re-deriving the path independently. **Additionally**,
  `ensureGlobalSettingsEnv` shall resolve HOME through the package's existing injectable seam
  `userHomeDirFn` (`internal/cli/glm_tools.go:123`) rather than calling the plain function
  `userHomeDir` (`internal/cli/homedir.go:14`) directly as it does today at `update.go:843`
  (re-measured on HEAD `2255165f5`; recorded pre-M1 as `:833`. `update.go:843` is the
  `userHomeDir()` call and is **NOT** the `os.RemoveAll` site, which is now `update.go:853` — the
  pre-M1 tables recorded that `os.RemoveAll` site at `:843`). Without
  this substitution the function has **no** seam a test can redirect, and since NFR-UDS-002 forbids
  `t.Setenv("HOME", …)`, REQ-UDS-011 and REQ-UDS-012 would be unimplementable. The existing
  override precedent is `glm_tools_test.go:34-36`.
- **REQ-UDS-014** — The guard shall be shown to FAIL against a deliberately widened radius before it
  is accepted.

### D.4 Non-vacuous user-area safety

- **REQ-UDS-015** — The user-area preservation guard shall drive a real production entry point of
  the update subsystem, not a fake defined in the test file.
- **REQ-UDS-016** — The replacement guard shall assert that user-owned harness agents and `hns-*`
  skills are byte-identical before and after that entry point runs.
- **REQ-UDS-017** — The replacement guard shall be shown to FAIL when the production code under it
  is mutated to touch a user-owned path.
- **REQ-UDS-018** — The `simulateMoaiUpdate` fake shall be removed, so no later reader mistakes it
  for coverage.

### D.5 Failure contract and lockout escape

- **REQ-UDS-019** — **When** any step of either update path returns an error after the first
  destructive step has run, the update shall write a recovery manifest recording the failed step,
  the run-scoped backup directory, and the restore command, and shall print that manifest.
- **REQ-UDS-020** — The update shall not attempt an automatic rollback of the tree. The decision and
  its rationale are recorded in `plan.md §B`.
- **REQ-UDS-021** — A restore entry point shall accept a backup directory and apply it to the
  project tree.
- **REQ-UDS-022** — **When** the restore entry point runs on a project whose
  `.moai/config/sections/system.yaml` is absent, it shall proceed rather than reject the tree,
  because that absence is the damage it exists to repair.
- **REQ-UDS-023** — The restore entry point shall be idempotent: running it twice against the same
  backup directory shall leave the same tree state as running it once.
- **REQ-UDS-024** — The restore entry point shall refuse a backup directory that is not a
  recognisable update backup, rather than copying arbitrary content into the project.
- **REQ-UDS-025** — The project-marker gate at the update entry point shall remain in force for the
  ordinary update path; only the restore entry point is exempt.

### D.6 Partial-restore reporting

- **REQ-UDS-026** — **When** `mergeBackPreserveInventory` fails part-way through the inventory, its
  error shall name the file at which it stopped and the count of files already restored.
- **REQ-UDS-027** — The three failure returns of `mergeBackPreserveInventory` shall each be covered
  by a test that reaches them deterministically.
- **REQ-UDS-028** — The stat, `MkdirAll`, and `copyFile` failure branches shall be distinguishable
  from one another in the returned error.

## §E Non-Functional Constraints

- **NFR-UDS-001** — Crash-window defects are probabilistic on the success path. Tests shall reach
  the failure window deterministically by invoking the backup step and the destructive step as
  separate calls and asserting on the filesystem between them, or by injecting a failing step —
  never by racing a real crash.
- **NFR-UDS-002** — All tests shall use `t.TempDir()`. A test exercising HOME-scoped behaviour shall
  redirect the home lookup through the injectable `userHomeDirFn` seam (per REQ-UDS-013) rather than
  mutating the process environment, so parallel tests cannot observe each other's HOME. Note that
  `userHomeDir` itself is a plain function and is **not** redirectable; the seam is the package-level
  variable `userHomeDirFn`.
- **NFR-UDS-003** — No file under `internal/template/templates/**` shall be added or modified.
- **NFR-UDS-004** — Go conventions: `snake_case.go` filenames, `fmt.Errorf("…: %w", err)` wrapping,
  English comments and godoc.
- **NFR-UDS-005** — Added backup writes shall not change the success-path outcome of an update: a
  run that succeeds today shall produce the same tree after the change.
- **NFR-UDS-006** — Path handling shall remain correct on linux, darwin, and windows.
- **NFR-UDS-007** — The `defs.DeprecatedPaths` entry count and its guarding test shall be unchanged.

## §F Success Criteria

- After the two backup steps and before the clean step, `.claude/settings.json`,
  `.moai/status_line.sh`, and `.gitignore` each exist under the run-scoped backup directory —
  verified by an executed test that inspects the filesystem between the two calls.
- The destructive-target registry lists every `RemoveAll` / `Rename` site in the update subsystem,
  and its guard — enumerating those sites by static source scan, independently of the registry —
  fails when a new site is added without a registry row.
- A backup write that fails aborts the run before the first destructive step, surfacing a grep-able
  sentinel that names the un-backed-up file, and `CleanMoaiManagedPaths` is never reached.
- Widening the HOME-scoped removal to `~/.claude/hooks` makes the suite fail.
- The replacement user-area guard fails when production code is mutated to touch
  `.claude/agents/harness/`.
- A tree whose `system.yaml` was destroyed can be restored by the restore entry point.
- All three `mergeBackPreserveInventory` failure branches are exercised, and each error names its
  own cause.
- `go build ./...`, `GOOS=windows GOARCH=amd64 go build ./...`, and `go test ./internal/cli/...` pass.

## §G Risk Register

| Risk | Direction | Bound |
|---|---|---|
| On-disk backup writes slow every update or fill the disk | Performance / capacity regression | Backup set is three small files; the existing `CleanupOldBackups` retention applies to the same backup root |
| A restore command becomes a second destructive path | Data loss | REQ-UDS-024 refuses unrecognised directories; REQ-UDS-023 makes it idempotent; it never deletes, only writes |
| Marker-gate exemption widens into a general bypass | Safety regression | REQ-UDS-025 scopes the exemption to the restore entry point alone, asserted by an AC |
| Replacement safety guard is itself vacuous | Undetected regression | REQ-UDS-017 requires a demonstrated mutation failure via `go test -overlay` |
| Deletion-radius guard passes for the wrong reason | False confidence | REQ-UDS-014 requires a demonstrated failure against a widened radius |
| Registry drifts as new destructive sites are added | Silent coverage gap | REQ-UDS-007 makes the guard enumerate sites by static source scan, independently of the registry, so a new unregistered site fails the multiset comparison |
| The drift guard is itself tautological (`count == count`) | False confidence | REQ-UDS-007 names the independent enumeration source explicitly and rejects registry-sourced counting; AC-UDS-005's falsification (§C.4) injects a dummy destructive site and requires an observed `--- FAIL` |
| Backup-failure abort path is never exercised | Data loss on the one path this SPEC exists to protect | REQ-UDS-003 is covered by AC-UDS-020, which forces the backup write to fail and asserts `CleanMoaiManagedPaths` was not called |
| Depending on E1 blocks this SPEC | Schedule coupling | Only the deprecated-path backup is inherited; every other milestone here is independent of E1's merge |

## §H Cross-References

> **All coordinates below re-measured on HEAD `2255165f5` at the M2 Step 0 re-baseline** (previously
> measured on HEAD `89b2e4772` / code baseline `8cc108ddb` in the v0.4.0 round). They have now
> drifted **three** times — once at iteration 1, again when **E1** landed between iterations 2 and 3
> (external), and again when **this SPEC's own M1 commit `4ddd35120`** landed (self-inflicted, NOT a
> third external TOCTOU) — so prefer the named symbol over the line number when navigating; the
> numbers are a convenience, the symbols are the contract. Entries for `update.go`,
> `update_template_sync.go`, and `update_clean_install.go` moved at M1; every other entry below was
> re-measured on `2255165f5` and is unchanged.

- `internal/cli/update/deploy/deploy.go` — `CleanMoaiManagedPaths` (`:28`), the `cleanTarget` list
  (`:38-68`), the removal loop (`:83`, `:105`), the `.moai/config` wipe (`:115`, `:121`),
  `MigrateLegacyMemoryDir` (`:143`, `:169`, `:176`)
- `internal/cli/update_template_sync.go` — `gitignoreBackup` (declared `:298`, populated `:388`),
  `mergeableBackups` (declared `:300`, populated `:397`), `Clean Managed Paths` (`:243`,
  the step name is now the constant `cleanManagedPathsStage`), `Restore Settings` (`:279`, `:400`),
  `MergeUserFiles` (`:436`)
- `internal/cli/update_clean_install.go` — Step 4 deprecated-path removal (`:251`, `:313-318`), the
  in-memory mergeable set (`:356`, `:359`), `BackupMoaiConfig` call (`:348`)
- `internal/cli/update_residue_cleanup.go` — `runV3ResidueCleanup` (`:65`), the
  backup-before-delete step (`:114-131`, `DEPRECATED_BACKUP_FAILED` sentinel at `:116`/`:125`), the
  removal loop (`:135`). Added by E1's run PR #1261 (`beeb0ebc2`) — §C.0 registry row 11
- `internal/cli/update_preserve_inventory.go` — `preserveInventoryRoots` (`:68-72`),
  `mergeBackPreserveInventory` (`:400`, failure returns `:416`/`:420`/`:424`)
- `internal/cli/update_namespace_protect.go` — `userOwnedScanRoots` (`:39-43`),
  `backupUserOwnedNamespace` (`:196`), `verifyNamespaceBackupCoverage` (`:293`)
- `internal/cli/update/backup/backup.go` — `BackupMoaiConfig` (`:27`, `.moai/config` walk at `:28`),
  `CleanupOldBackups` (`:206`)
- `internal/cli/update_archive.go` — `archiveSkill` (`:66`, removal at `:101`),
  `archiveLegacySkills` (`:280`, rename at `:322`)
- `internal/cli/update_cleanup.go` — `removeDeprecatedFile` (`:310`, removal at `:324`)
- `internal/cli/update.go` — project-marker gate: the `checkProjectMarker(cwd)` call (`:251`),
  whose predicate + error text M1 extracted into `internal/cli/update_restore.go` (`:21`, error
  text `:27`); `ensureGlobalSettingsEnv` (`:842`), `userHomeDir()` call (`:843`),
  `globalHooksDir` (`:851-853`)
- `internal/cli/homedir.go` — `userHomeDir` (`:14`), a plain function with no injection seam
- `internal/cli/glm_tools.go` — `userHomeDirFn` (`:123`), the package's existing injectable
  home-lookup variable (test override precedent at `glm_tools_test.go:34-36`)
- `internal/cli/update_safety_test.go` — `simulateMoaiUpdate` (`:49-58`),
  `TestMoaiUpdate_PreservesUserArea` (`:60`), the fake's invocation (`:95`), tautological
  assertions (`:91`, `:99-100`)
- `internal/defs/dirs.go` — `// brand + db directories` group heading (`:305`), `.moai/project/brand`
  (`:307`), `.moai/db` entry (`:313`), the unrelated `SPEC-V3R6-V2-V3-CLEAN-REINSTALL-001` mention
  the AC-UDS-007 (a) window must not reach (`:302`), Category C banner (`:319`), Category D comment
  (`:344-351`, `.moai/project/db/` mentions at `:346` and `:349`), `db.yaml` entry (`:354`)
- `.moai/specs/SPEC-UPDATE-REINSTALL-LOOP-002/spec.md` — REQ-RIL2-015/016 (deprecated-path backup)
- `CLAUDE.local.md` §24 — the user-owned namespace contract Defect 4 leaves undefended
