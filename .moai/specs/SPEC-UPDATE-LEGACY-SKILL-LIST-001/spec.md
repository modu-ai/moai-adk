---
id: SPEC-UPDATE-LEGACY-SKILL-LIST-001
title: "moai update — legacySkillIDs holds three live template skills, so the v2.16 archive drift-check can never converge (list correction + cross-check guard + wrong-archive removal + non-aborting archive loop)"
version: "0.3.0"
status: draft
created: 2026-07-31
updated: 2026-07-31
author: manager-spec
priority: P1
phase: "v3.0.2"
module: cli
lifecycle: spec-anchored
era: V3R6
tier: M
tags: "cli, update, archive, legacy-skills, drift, guard-test, embedded-manifest, git-hygiene"
issue_number: null
related_specs: [SPEC-UPDATE-REINSTALL-LOOP-002, SPEC-UPDATE-DATA-SURVIVAL-001, SPEC-CONFIG-TIER-PERSIST-001, SPEC-CONFIG-KEY-HONESTY-001, SPEC-UPDATE-CI-GUARD-001, SPEC-UPDATE-DOC-DRIFT-001, SPEC-V3R6-UPDATE-ARCHIVE-CONTRACT-001]
---

# SPEC-UPDATE-LEGACY-SKILL-LIST-001

## HISTORY

| Version | Date | Change |
|---------|------|--------|
| 0.1.0 | 2026-07-31 | Initial draft. Four-lens audit of `moai update` / `.moai/config`; member of that Epic, joined after the initially-numbered set. |
| 0.2.0 | 2026-07-31 | Plan-audit iteration-1 revision (M1-M4 + S1-S10). Epic ordinal dropped (the existing "of 6" numbering is inconsistent with the actual 7-member roster — plan.md §A D1). `REQ-LSL-008` restated as a guard-subject event-detected requirement; `REQ-LSL-012` rewritten as a single-subject prohibition; `REQ-LSL-018` added so `AC-LSL-006` has an owning requirement. Frontmatter adopted the Epic's canonical style (quoted `version`, `P1`, release-target `phase`). |
| 0.3.0 | 2026-07-31 | Plan-audit iteration-2 revision (MF-1..MF-3 + SF-a..SF-h). Anchoring swept across all 16 ACs with a standing rule added at acceptance.md §A 3a; plan.md §A D1's mechanism argument retracted in favour of a plain scope statement, with both prior retractions recorded; `AC-LSL-015(b)` given a measured numeric EXPECT; `REQ-LSL-018` split into 018/019 so `AC-LSL-006(a)` and `(c)` each trace to one requirement. |

## §A Problem / Motivation

`moai update` archives a fixed list of "legacy" skills into `.moai/archive/skills/v2.16/<id>/` before the clean-reinstall step wipes and redeploys `.claude/skills/`. Three entries on that list are not legacy at all — they are live, currently-shipped template skills. Because a live skill is re-created by every update, the archive drift-check compares a fresh copy against a frozen 2026-05 snapshot and reports `ARCHIVE_DRIFT` on every single run, forever.

### Defect 1 — three live template skills are on the legacy list

`legacySkillIDs` (`internal/cli/update_archive.go`, the `var legacySkillIDs = []string{` block) is documented as "the 16 skill IDs removed in BC-V3R3-007". Three of the 16 exist in the shipped template tree today:

```
$ for s in moai-domain-backend moai-domain-frontend moai-domain-database; do \
    printf "%-26s tmpl=%s\n" "$s" \
      "$([ -f internal/template/templates/.claude/skills/$s/SKILL.md ] && echo YES || echo no)"; done
moai-domain-backend        tmpl=YES
moai-domain-frontend       tmpl=YES
moai-domain-database       tmpl=YES
```

All three are registered in the deployment catalog (`internal/template/catalog.yaml` lines 157/159, 162/164, 220/222 — a `name:` and a `path: templates/.claude/skills/<id>/` for each), and all three are heavily referenced by the rest of the harness. Counting files that mention each ID across `.claude/agents`, `.claude/rules`, `.claude/skills`, `.moai/config` and their template mirrors, excluding each skill's own directory:

| Skill | Referencing files | Total occurrences |
|-------|------------------:|------------------:|
| `moai-domain-backend` | 56 | 68 |
| `moai-domain-frontend` | 26 | 34 |
| `moai-domain-database` | 30 | 30 |

> Correction to the delegation brief: the brief cited 46 / 17 / 25. The measured file counts are 56 / 26 / 30 (occurrence counts 68 / 34 / 30). The conclusion is unchanged and strengthened — these are among the most-referenced skills in the repository.

The remaining 13 entries are genuinely gone from both trees:

```
$ for s in moai-domain-db-docs moai-domain-mobile moai-framework-electron \
           moai-library-shadcn moai-library-mermaid moai-library-nextra \
           moai-tool-ast-grep moai-platform-auth moai-platform-deployment \
           moai-platform-chrome-extension moai-workflow-research \
           moai-workflow-pencil-integration moai-formats-data; do \
    [ -d internal/template/templates/.claude/skills/$s ] && echo "$s STILL PRESENT"; done
# (no output — all 13 absent)
```

The skill inventory is otherwise in perfect sync: 31 template skill directories, 31 catalog `path: templates/.claude/skills/` entries, and 31 local directories (30 `moai-*` plus the bare `moai` unified skill; the 7 local `hns-*` directories are user-owned and correctly absent from the template tree). There are no orphan or missing skills. **The defect is the list, not the inventory.**

### Defect 2 — the drift is structurally permanent, not a stale-state accident

Four verified facts compose into a loop that cannot terminate:

1. The clean-reinstall target list (`internal/cli/update/deploy/deploy.go`, the `targets := []cleanTarget{` block) includes `.claude/skills/moai*` as a glob (`isGlob: true`). Every update deletes all `moai*` skills and redeploys them from the embedded template.
2. That target list does **not** include `.moai/archive/`. Under `.moai/` only `config/` is removed (`os.RemoveAll(configDir)` further down the same function). The archive survives every update by design — it is user-data backup.
3. `archiveLegacySkills` runs **after** redeployment, in the "Post-sync steps" block of `internal/cli/update.go`.
4. `checkArchiveDrift` therefore compares a freshly-redeployed `SKILL.md` against a frozen archive snapshot. `archiveSkill` short-circuits only when the source is *absent*; these three sources are re-created on every run, so the short-circuit never fires.

Observed in this repository (live vs archived `SKILL.md`, md5):

| Skill | live | archived |
|-------|------|----------|
| `moai-domain-backend` | `e9edb94ad7fbdca57cff8a96e54fd11f` | `7469bb4e09bae816ea7268e940368095` |
| `moai-domain-database` | `534149429f1db713b1df268dbde9dc60` | `ea4d38eb507dd73a47e7da2ff80a63bc` |
| `moai-domain-frontend` | `080bc37188d9e5ea7a5af7838b22013e` | `4f58f1b22314a95f36321ad1ad7f8bbc` |

Reproduced on two projects: a user project and this development repository.

### Defect 3 — how the list went wrong, and why nothing corrected it

| Commit | Date | Effect |
|--------|------|--------|
| `74bae50f4` | 2026-04-27 | "feat(template): 16 정적 skills 제거 (BC-V3R3-007)" — deleted all 16 from both `.claude/skills/` and the template tree. |
| `ec0e9e257` | 2026-04-27 | "feat(update,migrate): archive 마이그레이터 + restore-skill 서브커맨드 (M4)" — authored `legacySkillIDs` with the 16 names. |
| `697a6e2c7` | 2026-04-28 (PR #709) | "feat(v3R2): Wave 1 Bundle - Skill Consolidation Stage 1 …" — re-added `moai-domain-backend` / `-frontend` / `-database` `SKILL.md` to both trees as consolidation targets. |

Only `SKILL.md` came back. The original legacy bodies had more: at `74bae50f4^`, `moai-domain-backend` carried `SKILL.md`, `references/examples.md`, and `references/reference.md`.

`git log -S"moai-domain-backend" -- internal/cli/update_archive.go` returns exactly one commit (`ec0e9e257`). The list has never been edited since the day it was written — it was never updated for the next-day revival.

### Defect 4 — a sibling test was corrected for the revival; the list was not

`internal/template/skills_removal_test.go` (`TestRemovedSkillsNotPresent`) asserts the removed directories are absent from the template tree. Its `removed` slice carries only **9** entries and an explicit in-source note:

```go
// NOTE: Some skills from the original 16-skill removal list still exist in the template tree.
// This test only verifies removal of skills that have actually been deleted.
…
// NOT removed (still exist): moai-domain-backend, moai-domain-frontend, moai-domain-database,
// moai-framework-electron, moai-platform-auth, moai-platform-deployment, moai-platform-chrome-extension
```

So the revival *was* noticed — in the template package — and the test was narrowed to accommodate it. The parallel correction to `legacySkillIDs` in the CLI package never happened. (The comment's second line is itself now stale: `moai-framework-electron`, `moai-platform-auth`, `moai-platform-deployment`, and `moai-platform-chrome-extension` were subsequently removed and are absent from both trees today. That staleness is recorded as an observation, not as scope — see §C.)

### Defect 5 — the existing tests are structurally blind

Six test files reference `legacySkillIDs`: `update_idempotency_test.go`, `update_archive_flow_test.go`, `migrate_restore_skill_test.go`, `update_archive_force_test.go`, `update_skip_sync_test.go`, `update_archive_test.go`. Every one is **self-referential**: it reads `legacySkillIDs` and seeds synthetic skill directories from it inside a `t.TempDir()`, then asserts against the same list (`update_archive_flow_test.go:26` seeds from the list; `:44` asserts `archived == len(legacySkillIDs)`). Whatever the list contains, they pass. No test cross-checks the list against the real embedded template tree.

A ready-made helper for the missing check already exists: `template.EmbeddedMoaiSkillNames()` (`internal/template/skills_manifest.go`) returns the `moai-*` skill directory names under the embedded templates' `.claude/skills/`, and its doc contract states that callers **must treat an empty derived set as "manifest unavailable" and degrade gracefully rather than mis-classify**. A temporary probe run inside `internal/cli` measured the intersection directly:

```
legacySkillIDs=16 embedded=30 overlap=3 [moai-domain-backend moai-domain-frontend moai-domain-database]
```

### Defect 6 — three archive copies of live skills are committed to git

`git ls-files .moai/archive/skills/v2.16` lists 7 files, one `SKILL.md` per directory, committed by `9373e558f` (2026-05-11). Three of them archive live template skills; `.gitignore` carries no `archive` rule. Each contains only a `SKILL.md` — a snapshot of the *post-revival consolidated* skill, not the deleted legacy content. The genuinely deleted legacy bodies (`references/`, `modules/`) exist only in git history at `74bae50f4^`. So these three archived files preserve nothing that needs preserving and actively feed the drift loop.

### Defect 7 — one failing entry aborts the whole archive pass

`archiveLegacySkills` returns on the first per-skill error (`return archived, fmt.Errorf("archive %s: %w", id, err)` inside the loop, plus two drift-backup error returns above it). A single failing entry skips every remaining entry and suppresses the `total: N skills archived` summary line — so the operator sees neither the full failure set nor the count of what did succeed.

### Relationship to SPEC-V3R6-UPDATE-ARCHIVE-CONTRACT-001

That SPEC met the identical `ARCHIVE_DRIFT` symptom on another project (its `spec.md:35` records the diff evidence) but diagnosed it as "`--force` is not wired" plus "`--skip-sync` still triggers the archive check", and explicitly placed the list itself out of scope (`spec.md:76` — "`legacySkillIDs` 목록의 변경은 다루지 않는다"; also `plan.md:44`, `acceptance.md:268`).

Its `--force` path moves the existing archive to `v2.16-drift-<UTC-stamp>/` and re-archives from the live source. For these three entries that writes a **current** skill into a directory tree labelled "v2.16 legacy", and does so afresh on every forced run. `--force` is therefore a symptom workaround for this defect, not a fix; this SPEC removes the cause the earlier SPEC deferred. Nothing in `SPEC-V3R6-UPDATE-ARCHIVE-CONTRACT-001` is reverted.

## §B Goals

1. `legacySkillIDs` names only skills that are genuinely absent from the shipped template set.
2. A mechanical guard makes the class of defect non-recurring: no future edit can put a live template skill back on the legacy list without a red test.
3. The three wrongly-committed archive copies leave the repository; the four genuine ones stay.
4. A failing archive entry no longer hides the fate of the remaining entries.
5. No behavioural change for the 13 correctly-listed skills, and no change to `archiveVersion` or the archive directory scheme.

## §C Scope Exclusions

### Out of Scope — `moai migrate restore-skill` guard

- `restoreSkill` (`internal/cli/migrate_restore_skill.go`) copies archive → `.claude/skills/`, and with `--force` calls `os.RemoveAll(targetDir)` before copying. For these three IDs it would overwrite a current skill with the stale archive. This is a real hazard and is a named follow-up, not this SPEC's work: it needs its own guard design and its own acceptance criteria.

### Out of Scope — cleanup of wrong archives in downstream user projects

- Correcting the list stops new wrong archives from being written, but archives already created in user projects are user files under `.moai/archive/`. Removing them requires backup, rollback, and consent design of its own. Deferred.

### Out of Scope — retirement of `moai-meta-harness`

- Its `SKILL.md` self-declares DEPRECATED and redirects to the v4 Builder, while rules and tests still reference it. Unrelated to the archive list; a separate follow-up.

### Out of Scope — the other 13 list entries, `archiveVersion`, and the archive scheme

- The 13 retained IDs are not re-examined, re-ordered, or re-worded. `archiveVersion` stays `v2.16`. The `.moai/archive/skills/<version>/<id>/` layout is unchanged.

### Out of Scope — the stale comment in `skills_removal_test.go`

- Its "NOT removed (still exist)" note lists four IDs that have since been removed, and its `removed` slice checks only 9 of the 13 genuinely-removed skills. Both are real coverage/documentation gaps in a different package; recorded as an observation in §A Defect 4 and left for a follow-up.

### Out of Scope — template content

- No file under `internal/template/templates/` is created, edited, or deleted by this SPEC. The three skills stay exactly as shipped.

## §D Requirements (GEARS)

### D.1 Legacy-list correctness

- **REQ-LSL-001** — The `legacySkillIDs` list shall contain only skill IDs that are absent from the embedded template `moai-*` skill set.
- **REQ-LSL-002** — The corrected list shall contain exactly 13 entries, retaining the 13 genuinely-removed IDs verbatim and in their current relative order.
- **REQ-LSL-003** — The list shall not contain `moai-domain-backend`, `moai-domain-frontend`, or `moai-domain-database`.
- **REQ-LSL-004** — The list's doc comment shall state the corrected count and record that three BC-V3R3-007 entries were revived by `697a6e2c7` and are therefore not legacy.

### D.2 Cross-check guard

- **REQ-LSL-005** — When the Go test suite runs, the guard test shall compute the intersection of `legacySkillIDs` with `template.EmbeddedMoaiSkillNames()` and shall fail when that intersection is non-empty.
- **REQ-LSL-006** — On failure the guard shall name every intersecting ID in its message, so the operator does not have to re-derive the set.
- **REQ-LSL-007** — When `template.EmbeddedMoaiSkillNames()` returns an error, or returns an empty set, the guard shall skip rather than pass or fail — per that helper's documented "manifest unavailable" contract.
- **REQ-LSL-008** — **When** the guard is evaluated against the pre-correction 16-entry list, the guard shall fail and shall name the offending IDs — so its ability to catch this defect is observed rather than assumed.
- **REQ-LSL-009** — The guard shall derive the live set from the embedded manifest at test time; it shall not hard-code a second copy of the skill inventory.

### D.3 Wrong-archive removal

- **REQ-LSL-010** — The repository shall not track archive copies of skills that exist in the embedded template set.
- **REQ-LSL-011** — The four genuine archive entries (`moai-framework-electron`, `moai-platform-auth`, `moai-platform-chrome-extension`, `moai-platform-deployment`) shall remain tracked and byte-unchanged.
- **REQ-LSL-012** — The SPEC shall not add a `.gitignore` rule for the archive tree. (The rationale and the follow-up disposition are recorded in plan.md §A D5.)

### D.4 Non-aborting archive loop

- **REQ-LSL-013** — When a per-skill step inside `archiveLegacySkills` fails, the function shall record the failure and continue with the remaining entries instead of returning immediately.
- **REQ-LSL-014** — The function shall emit its `total: N skills archived` summary line even when one or more entries failed.
- **REQ-LSL-015** — The function shall return an aggregate error naming every failed skill ID, and shall return `nil` when no entry failed.
- **REQ-LSL-016** — The reported archived count shall count successful archives only.
- **REQ-LSL-017** — The literal output keywords `archive: ` and `total: ` shall be preserved verbatim, per the `@MX:NOTE` contract above `archiveLegacySkills`.

### D.5 Regression containment

- **REQ-LSL-018** — The six pre-existing test files that read `legacySkillIDs` shall continue to pass with their assertion behaviour unmodified.
- **REQ-LSL-019** — The `…_All16…` test-name disposition (rename or keep) shall be applied consistently across both files that carry it; a rename shall preserve the `TestArchiveSkill_` / `TestRestoreSkill_` prefixes.

## §E Non-Functional Constraints

- **NFR-LSL-001** — All new and modified tests use `t.TempDir()` for filesystem fixtures; the development project's `.claude/` and `.moai/archive/` are never touched by a test.
- **NFR-LSL-002** — No file under `internal/template/templates/` is created, edited, or deleted.
- **NFR-LSL-003** — Behaviour for the 13 retained IDs is unchanged: same archive destination, same idempotency short-circuit, same `--force` drift-backup path.
- **NFR-LSL-004** — Files this SPEC touches are `gofmt`-clean and `go vet`-clean.
- **NFR-LSL-005** — The guard test adds no dependency from `internal/cli` to any package it does not already import (`internal/template` is already imported by `internal/cli/doctor_skills.go`).

## §F Success Criteria

1. `legacySkillIDs` has 13 entries and none of them resolve to an embedded template skill.
2. The new guard fails on the pre-correction list and passes on the corrected one, and skips when the manifest is unavailable.
3. All six pre-existing `legacySkillIDs` test files still pass unmodified in behaviour.
4. `git ls-files .moai/archive/skills/v2.16` lists 4 files, all under the four genuine directories.
5. An injected per-entry failure no longer suppresses the remaining entries or the `total:` line.
6. `go build ./...`, `go vet ./...`, and `go test ./internal/cli/... ./internal/template/...` all pass.

## §G Risk Register

| # | Risk | Likelihood | Impact | Mitigation |
|---|------|-----------|--------|------------|
| 1 | Shrinking the list shifts what positionally-indexed tests exercise (`legacySkillIDs[0..3]`, `[:5]`, `[:8]`) | High (certain) | Low | All indices stay in range at 13 entries; the tests seed synthetic dirs from whatever ID they draw, so they remain valid. Verified in plan.md §C. |
| 2 | Test names `…_All16Skills` / `…_All16RoundTrip` become misnomers | High | Very low | Rename is optional and behaviour-neutral; covered by an AC so the choice is recorded either way. |
| 3 | The guard fails in an environment where the embedded FS cannot be read | Low | Medium | REQ-LSL-007 requires a skip, not a failure, on manifest-unavailable. Covered by its own AC. |
| 4 | Removing archive files from git deletes something a user still needs | Low | Medium | The three files hold post-revival consolidated `SKILL.md` snapshots, not the deleted legacy bodies; the real legacy content remains in git history at `74bae50f4^`. The four genuine entries are protected by REQ-LSL-011 and its AC. |
| 5 | Error accumulation changes the caller's expectations in `runUpdate` | Medium | Low | The call site already treats a non-nil error as a warning line and continues; the aggregate error is strictly more informative. Verified in plan.md §C. |
| 6 | A future edit re-adds a live skill to the list | Medium | High | This is exactly what the REQ-LSL-005 guard exists to catch, and REQ-LSL-008 makes its catching power observed rather than assumed. |

## §H Cross-References

- `internal/cli/update_archive.go` — `legacySkillIDs`, `archiveSkill`, `checkArchiveDrift`, `archiveLegacySkills`
- `internal/cli/update.go` — "Post-sync steps" block, the `archiveLegacySkills(cwd, out, getBoolFlag(cmd, "force"))` call
- `internal/cli/update/deploy/deploy.go` — clean-reinstall `targets` list, the `.claude/skills/moai*` glob, the `.moai/config` removal
- `internal/template/skills_manifest.go` — `EmbeddedMoaiSkillNames()` and its manifest-unavailable contract
- `internal/template/skills_removal_test.go` — `TestRemovedSkillsNotPresent`, the narrowed 9-entry `removed` slice
- `internal/template/catalog.yaml` — lines 157/159, 162/164, 220/222
- `internal/cli/migrate_restore_skill.go` — `restoreSkill` (out-of-scope follow-up)
- `SPEC-V3R6-UPDATE-ARCHIVE-CONTRACT-001` — the prior SPEC that met this symptom and deferred the list
