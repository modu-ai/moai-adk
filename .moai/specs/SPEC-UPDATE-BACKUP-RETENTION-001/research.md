# Research — SPEC-UPDATE-BACKUP-RETENTION-001

Codebase and evidence research for card t1733. Tree under study: `/Users/goos/MoAI/moai-adk-go/.moai/worktrees/t1733` at `e669da81b3f865488dd40f840623bed69839f459` (branch `WT-10-14-class-3`, clean at start). Every file:line below was read in this tree in this run. Nothing in this file is a recommendation about a judgment call; unresolved choices are in `decision-index.md` and `plan.md` § B.

Baseline: main (ec5a4f392) produces the same loss as this tree (reproduction report /private/tmp/t1687-repro-report.md). Table (probe, main `ec5a4f392`, t1687 `fb7e1933b`, pre-revert `8e43f064a^`): P1 user-file copies 1 -> 0 on main and on t1687, 1 on pre-revert; P2 no v1/v2/v3 entry survives on main and t1687 (deleted=4), all survive pre-revert (deleted=0); P3 v1-holding entries 10 -> 0 on main and t1687, 10 pre-revert; namespace-root directories 79 -> 0 on main and t1687, 79 pre-revert; 1700 of 1951 distinct file versions (996 paths) lost on both main `ec5a4f392` and t1687 `fb7e1933b`, 0 lost pre-revert. The defect is therefore pre-existing on main; commit `8e43f064a` removed a protection main never had and introduced no over-deletion. Wherever an earlier draft of this SPEC read "compared with main" or implied that t1687 introduced the loss, that wording is withdrawn. The pre-revert `retention.go` behavior at `8e43f064a^` (namespace runs 79 kept, loss 0) is the reference baseline for the repair (design.md D0). Caveat from the report: its snapshot directory mtimes are copy times, identical across the three sides; absolute numbers on the live project may differ.

## 1. Method

1. Read the t1687 audit reports (`.moai/worktrees/t1687/.moai/reports/t1687/sync-audit-iter2.md` to `iter6.md`, `progress.md`) and the t1799 reports as claims, not as facts.
2. Confirmed each finding against the code in this tree. Where the code named by the audit no longer exists, the finding is recorded against the successor code or marked resolved with the grep that shows it.
3. Measured the present behavior with probe tests injected through `go test -overlay`, so the tree is not modified (the probe sources are in Appendix A). No `moai update` was run in any tree; no live project was read for writing.
4. Read the operator rulings from the decision board (`moai decision read --scope card:t1687 --all`). They are not committed artifacts; they are quoted here as evidence (§ 2).

## 2. Operator rulings (decision board, not committed; evidence only)

| Record | Scope | Content relevant to this SPEC |
|---|---|---|
| a263 | card:t1687 | Past versions: keep only the newest N per file (N a configuration value, default 3); allow cleanup of backups whose only unique content is older versions beyond N; each removal reported. Template-owned paths: `.template-defaults/**` and `pre-clean/` excluded from the guard ("regenerable from the binary"). Non-timestamp entries in `.moai-backups`: report with count and size, never delete. |
| fd8b | card:t1687, supersedes a263 | `pre-clean/` is NOT template-owned: it is the only recovery copy of user files the template does not ship, so retention treats it as unique content; the a263 premise is withdrawn; `templateOwnedPrefixes` and the test change "in the follow-up card". |
| a041 | card:t1687 | Split t1687: F2, F3, F5, F6, F7, M2, M3, M7, M6 land; backup-retention defects (F1, F13, F14, F15 with the pre-clean ruling) and the symlink-root defects (F16, F17) move to an after= follow-up; F18 and F19 are debts there; `moai update` stays forbidden in this repository until the follow-up lands. |
| 242e | card:t1687 | Fix F20 in t1687; codex optional findings F21, F24, F25, F26 move to t1733. |
| t1733 answers 2026-10-16 | card:t1733, relayed by the leader (decided_by=operator, verdict source "operator via brain") | Q1: report only, no deletion of a zero-byte database. Q2: soft cap; the last copy of a protected version is never deleted. Q3: do not reintroduce archive or link removal; report only; authority basis deferred to card 3.3. Principle for 3.2: add no data-deleting behavior. |

The card text of t1733 additionally states the completion judgment (probe-shaped tests P1-P4 and a real-data remeasure with both roots capped and at least one pre-clean user-file copy kept) and the `moai update` prohibition.

## 3. Finding verification

Audit line numbers are those in the cited iteration; "this tree" line numbers were read in this run.

| Finding | Audit claim | This tree | Disposition |
|---|---|---|---|
| F1, F14 | Cap removed 0 of 42 namespace entries; `entryMustStay` compared uniqueness to the window only (`retention.go:219-241`) | `retention.go` does not exist (`ls internal/cli/update/backup/` shows no such file; deleted by `8e43f064a`, `git show --stat`). The pruner is `backup.go:227-252` (run-count only). Measured: it removes everything outside the window (E1, E2, E3, E6). | The audited symptom belongs to reverted code. Present defect is over-deletion. Requirement REQ-UBR-003 to REQ-UBR-006, REQ-UBR-010. |
| F13 | `withinBudgetKeys` counted the window's version (`retention.go:183-213`) | Code gone; no budget exists. | Requirement REQ-UBR-004; probe P2 (E2). |
| F15 | `pre-clean/` listed in `templateOwnedPrefixes` (`retention.go:19-22`) against producer contract `deploy.go:106-115` and consumer `update_template_sync.go:1011-1015` | `templateOwnedPrefixes` gone. Producer contract unchanged (read `deploy/deploy.go:99-115`: "every regular file tmplFS does not carry at the same relative path is copied into the run's pre-clean backup"). Present pruner deletes the pre-clean run once rotated out (E1). | Requirement REQ-UBR-006 (operator ruling fd8b; not reopened). |
| F16 | `symlinkedRoot` checks `.moai`, `.claude`, `.claude/skills` only (`residue.go:246-254`); `removeStaleLink` compares link text (`:168-195`) | `symlinkedRoot` at `residue.go:299-310`, list at `:303`, still three roots. `removeStaleLink` does not exist; `retiredLinkPresent` (`:167-192`) only inspects and the caller reports (`:86-92`). Probe P4a (E4): through a symbolic `.agents/skills` the report names B's link as A's retired link; B's link survives only because nothing removes. | Operator ruling Q3 (2026-10-16): no removal is reintroduced; report only. REQ-UBR-016 (link content untouched) and AC-UBR-04 as a regression guard. Authority basis for cleanup deferred to card 3.3. |
| F17 | `archiveFolder` does not check the parent chain (`residue.go:157-163`) | `archiveFolder` does not exist (grep of non-test code returns nothing). Retired folders are reported (`residue.go:75-94`). Probe P4b (E5): `content-now-outside-project=false`. | Operator ruling Q3: no archive move is reintroduced; REQ-UBR-015 (report-only) and AC-UBR-05 as a regression guard. |
| F18 | `backupBookkeeping` (`backup.go:265-280`) matches bookkeeping by name | Function does not exist (grep). Writers: `backup_metadata.json` at `backup.go:136`, `.complete` at `update_namespace_protect.go:238-246` and `update_preserve_inventory.go:352-357`, run marker `run_marker.go` (`RunMarkerName`). | Requirement REQ-UBR-009. In scope: the rewritten census needs a bookkeeping rule anyway. |
| F19 | Default 3 defined twice; no symmetry test for the state key; no cli test of report rows | `grep -c BackupPastVersionsKeep internal/config/defaults.go` returns 0 (E8); the key is absent from the shipped template (E7). The duplication does not exist because the key does not. | Requirement REQ-UBR-007 (define once), REQ-UBR-020 (tests). In scope. |
| F20 | `CleanupNamespaceBackups` had no destructive-registry row | Function absent (grep). Registry row for `CleanupOldBackups` at `update_destructive_registry.go:105` (`Sites: 1`). `TestDestructiveTargetRegistry_CoversAllSites` was not run here (the `internal/cli` package is heavy). | Resolved on this tree; REQ-UBR-019 for changed sites. |
| F21 | `removeStateResidue` can remove an open zero-byte database (`update_state_residue.go:44-55`) | Present: `update_state_residue.go:32-62`, minimum age `:19`, sidecar check `:71`. Not re-measured; audit probe A3-P3 (iter3 E11) cited. | Operator ruling Q1 (2026-10-16): report only, no deletion. REQ-UBR-018. |
| F22 | `moveFile` alias hides rename sites from the registry scan | Registry row for the single `os.Rename` site `moveOne` at `update_destructive_registry.go:186`; `residue.go:194-199` (`moveOne`, `moveFile`). The pin test was not run here. | Resolved by t1687 follow-ups (iter5/iter6 record F35 fixed); REQ-UBR-019 for changed sites. |
| F23 | `NamespaceBackupRootRefused` had no production caller | Production callers now exist: `update_clean_install.go:557`, `update_template_sync.go:907`, `backup.go:231`. | Resolved. |
| F24 | `renameSkillFile` follows a skill folder swapped for a symlink | `renameSkillFile` at `residue.go:255-276` is path-based (`regularFileAt` check at `:257`, then `moveFile`). t1799 made the rename deferred (`SkipRenames`, `ApplyDeferredRenames` at `:205-244`) but did not bind it to a root. | Requirement REQ-UBR-017. |
| F25 | `freeName` does not reserve the name | `freeName` at `residue.go:281-297` `Lstat`-checks only. | Requirement REQ-UBR-017. |
| F26 | `os.ReadFile` blocks on a FIFO in the unmanaged restore | Fixed: `restore_unmanaged.go:108-110` and `:134-136` refuse non-regular target and side copy; tests `restore_unmanaged_fifo_unix_test.go:40`, `:70`; run in E9 (3 PASS, exit 0). Commit `06fc853bf` (card t1799). | Already fixed; no requirement. |

Items named in the audits but not carried by the operator decision (F8-F12, F27-F31, F34, F36, F37, F38): not respecified. Observed in passing: the Windows skip that F38 asked for is present at `residue_coverage_test.go:18-19`. The others were not checked.

## 4. Code map (read in this tree)

- Pruner and run model: `internal/cli/update/backup/backup.go:210-252` (`CleanupOldBackups`), `:254-261` (run type), `:263-274` (writer name shapes), `:284-289` (roots), `:293-328` (run collection: marker id or mtime for unmarked, ties by path ascending), `:340-362` (directory listing, symlinks not followed), `:368-395` (containment and refusal predicates). Callers: `update_clean_install.go:557-559`, `update_template_sync.go:907-909`, both keepCount 5.
- Run marker: `run_marker.go` (`RunMarkerName = ".moai-run"`, run-id layout), `run_marker_read.go` (regular-file-only read, 4096-byte limit).
- Backup writers: `backup.go:28-150` (`BackupMoaiConfig`: copies `.moai/config/` content to the directory top level, `.template-defaults/` render, `backup_metadata.json`); `update_namespace_protect.go:238-246` and `update_preserve_inventory.go:352-357` (`.complete`); deploy producer `deploy/deploy.go:99-115`.
- Skill residue: `skillresidue/residue.go:19-21` (retired map), `:53-135` (`Clean`), `:167-192` (link inspection), `:194-199` (single `os.Rename`), `:205-244` (deferred renames), `:255-297` (rename and free name), `:299-310` (root refusal). Reporting rows: `update_state_residue.go:174-193` (`skillActionMessage`).
- Registry: `update_destructive_registry.go:44-75` (header and rows), `:105`, `:186`.
- Toolchain: `go.mod` `go 1.26.8`; `go doc os.Root` lists `Rename`, `Remove`, `RemoveAll`, `Lstat`, `Mkdir`, `MkdirAll`, `Link`, `Symlink`, `OpenRoot` (E-note below, run in this tree).

## 5. Evidence ledger

All entries: tree SHA `e669da81b3f865488dd40f840623bed69839f459` (code identical at `3b133bedf`, which adds only this SPEC directory). Every command runs from the tree root exactly as written; no placeholder remains.

**Placing the probe files (self-contained recreation).** The directory `.moai/reports/t1733/probes/` is gitignored local evidence (`.gitignore:235` `.moai/reports/*`), so a fresh clone does not carry it. A reader recreates it in three steps, all from Appendix A: (1) write `backup_probe_test.go` from A.1, `residue_probe_test.go` from A.2 and `backup_probe_p5_test.go` from A.4 into `.moai/reports/t1733/probes/`; (2) write the three overlay files from A.3 beside them; (3) for the real-data entry E6 and the fixed-mtime probe P5 of AC-UBR-06(d), make a modification-time-preserving copy of a project's `.moai-backups` and `.moai/backups` into `/private/tmp/t1733-real/moaikr` (macOS APFS: `cp -Rcp <source> /private/tmp/t1733-real/moaikr`; elsewhere `cp -a`). The overlay maps (relative paths, resolved from the tree root) put the probe source in the target package without creating a file in the tree. Re-execution on this machine in this run printed the same output as each entry below (E1 to E6; E6 14.31 s).

Baseline on main: the probes were run against main `ec5a4f392` extracted with `git archive` by the reproduction report (`/private/tmp/t1687-repro-report.md`), output identical to E1-E3 and E6 (table at the head of this file). On main the `skillresidue` package and `update_state_residue.go` do not exist (E4/E5 have no main cell). The probe log line in E1 labelled `pre-clean-dirs-before` is a misnomer of the probe: it is evaluated after the prune and lists the five newer directories that remain; the old pre-clean directory is absent from it.

### E1 — P1: pre-clean user file after rotation

Command: `go test -count=1 -v -overlay=.moai/reports/t1733/probes/overlay_backup.rel.json -run '^TestT1733ProbeP1$' ./internal/cli/update/backup/`
Exit code: 0
```
=== RUN   TestT1733ProbeP1
    zz_t1733_probe_test.go:80: P1 pre-clean-dirs-before=[20261011_000000 20261012_000000 20261013_000000 20261014_000000 20261015_000000] user-file-copies-before=1 deleted=1 user-file-copies-after=0
--- PASS: TestT1733ProbeP1 (0.01s)
PASS
ok  	github.com/modu-ai/moai-adk/internal/cli/update/backup	0.233s
```

### E2 — P2: v1 v2 v3 v4 v4, keep 1

Command: `go test -count=1 -v -overlay=.moai/reports/t1733/probes/overlay_backup.rel.json -run '^TestT1733ProbeP2$' ./internal/cli/update/backup/`
Exit code: 0
```
=== RUN   TestT1733ProbeP2
    zz_t1733_probe_test.go:111: P2 deleted=4 remaining=[update-2026-10-05T00-00-00Z] v1-entry-survives=false v2-entry-survives=false v3-entry-survives=false
--- PASS: TestT1733ProbeP2 (0.01s)
PASS
ok  	github.com/modu-ai/moai-adk/internal/cli/update/backup	0.407s
```

### E3 — P3: ten copies of v1 and one v2, keep 1

Command: `go test -count=1 -v -overlay=.moai/reports/t1733/probes/overlay_backup.rel.json -run '^TestT1733ProbeP3$' ./internal/cli/update/backup/`
Exit code: 0
```
=== RUN   TestT1733ProbeP3
    zz_t1733_probe_test.go:123: P3 deleted=10 remaining=[update-2026-10-11T00-00-00Z] v1-holding-entries-remaining=0
--- PASS: TestT1733ProbeP3 (0.01s)
PASS
ok  	github.com/modu-ai/moai-adk/internal/cli/update/backup	0.226s
```

### E4 — P4a: symbolic `.agents/skills` shared with project B

Command: `go test -count=1 -v -overlay=.moai/reports/t1733/probes/overlay_residue.rel.json -run '^TestT1733ProbeP4a$' ./internal/cli/update/skillresidue/`
Exit code: 0
```
=== RUN   TestT1733ProbeP4a
    zz_t1733_probe_test.go:48: P4a err=<nil> actions=[{.claude/skills/moai-kanban-foreman reported-retired-skill} {.agents/skills/moai-kanban-foreman reported-retired-link}] b-link-survives=true a-folder-in-place=true
--- PASS: TestT1733ProbeP4a (0.00s)
PASS
ok  	github.com/modu-ai/moai-adk/internal/cli/update/skillresidue	0.090s
```

### E5 — P4b: symbolic `.moai/archive` to an outside directory

Command: `go test -count=1 -v -overlay=.moai/reports/t1733/probes/overlay_residue.rel.json -run '^TestT1733ProbeP4b$' ./internal/cli/update/skillresidue/`
Exit code: 0
```
=== RUN   TestT1733ProbeP4b
    zz_t1733_probe_test.go:73: P4b err=<nil> actions=[{.claude/skills/moai-kanban-foreman reported-retired-skill}] content-now-outside-project=false a-folder-in-place=true
--- PASS: TestT1733ProbeP4b (0.00s)
PASS
ok  	github.com/modu-ai/moai-adk/internal/cli/update/skillresidue	0.088s
```

### E6 — real-data baseline on a clone of the mo.ai.kr backup roots (10-14 snapshot)

Provenance of the input: a copy (`cp -Rcp`, clone-on-write, modification times preserved) of `.moai-backups` and `.moai/backups` from the 2026-10-14 t1687 audit snapshot `…/dfd2478c-11e3-4ea8-a787-c06d8bb4ade9/scratchpad/rc1/moaikr/`, which is itself a clone of the live project made by the t1687 audit. The live project was not touched. The probe prunes the copy, so a re-execution needs a fresh copy; two runs on two fresh copies gave identical counts.
Command: `env T1733_CLONE=/private/tmp/t1733-real/moaikr go test -count=1 -v -overlay=.moai/reports/t1733/probes/overlay_backup.rel.json -run '^TestT1733ProbeRealData$' ./internal/cli/update/backup/`
Exit code: 0
```
=== RUN   TestT1733ProbeRealData
    zz_t1733_probe_test.go:217: REAL cfg-root dirs 20->20 ns-root dirs 79->0 deleted=79 pre-clean-dirs 1->1
    zz_t1733_probe_test.go:218: REAL distinct-file-versions-before=1951 lost=1700 paths-with-lost-versions=996
    zz_t1733_probe_test.go:220: REAL lost-top path=.moai/harness/usage-log.jsonl versions-lost=50
    zz_t1733_probe_test.go:220: REAL lost-top path=.claude/agents/moai/plan-auditor.md versions-lost=17
    zz_t1733_probe_test.go:220: REAL lost-top path=.claude/skills/moai/SKILL.md versions-lost=16
--- PASS: TestT1733ProbeRealData (8.99s)
PASS
ok  	github.com/modu-ai/moai-adk/internal/cli/update/backup	9.156s
```
Reading: the config root keeps its five timestamped runs (the other 15 entries are non-timestamp entries the pruner neither counts nor deletes); the namespace root loses all 79 directories; the `pre-clean` run (`20261008_203921`, 47 files) is inside the window only because the config root holds exactly five timestamped runs. A first run on the first copy printed the same numbers in 15.23 s.

### E7 — budget key absent from the shipped template

Command: `grep -c backup_past_versions_keep internal/template/templates/.moai/config/sections/state.yaml`
Exit code: 1
```
0
```

### E8 — default constant absent from the config defaults

Command: `grep -c BackupPastVersionsKeep internal/config/defaults.go`
Exit code: 1
```
0
```

### E9 — F26 is fixed (non-regular restore targets are refused)

Command: `go test -count=1 -v -run FIFO ./internal/cli/update/backup/`
Exit code: 0
```
=== RUN   TestRestoreUnmanaged_FIFOTargetIsRefusedNotRead
--- PASS: TestRestoreUnmanaged_FIFOTargetIsRefusedNotRead (0.00s)
=== RUN   TestRestoreUnmanaged_FIFOSideCopyIsRefusedNotRead
--- PASS: TestRestoreUnmanaged_FIFOSideCopyIsRefusedNotRead (0.00s)
=== RUN   TestReadRunIDSkipsFIFOMarker
--- PASS: TestReadRunIDSkipsFIFOMarker (0.00s)
PASS
ok  	github.com/modu-ai/moai-adk/internal/cli/update/backup	0.386s
```

### E10 — SPEC ID check

Command: `ID="SPEC-UPDATE-BACKUP-RETENTION-001"; [[ "$ID" =~ ^SPEC(-[A-Z][A-Z0-9]*)+-[0-9]{3}$ ]] && echo PASS || echo FAIL`
Exit code: 0
```
PASS
```
Collision check: `ls .moai/specs | grep -c "SPEC-UPDATE-BACKUP-RETENTION"` returned 0.

## 6. Observations that shaped the design

- Unmarked runs are ordered by directory modification time (`backup.go:318-326`; `listBackupDirs` records `ModTime`). A copy that resets modification times changes the window. The first copy attempt of this run (`cp -Rc` by the t1687 auditor) produced identical modification times for every directory in a listing (all "Oct 14 16:21"); `cp -Rcp` was used here. This is why AC-UBR-06 requires a modification-time-preserving copy.
- The `.moai-backups` timestamped runs and the `.moai/backups` namespace runs are one pool for the window; on the mo.ai.kr snapshot the five config runs are newer than all 79 namespace runs, so the whole namespace root is excess.
- The deploy producer copies into `pre-clean/` only files the template does not carry at the same relative path (`deploy.go:106-115`); on the snapshot the one `pre-clean/` tree holds 47 files, which the pruner would delete on the next rotation.
- `go doc os.Root` in this toolchain lists the methods REQ-UBR-017 needs; the platform behavior of `Root.Link` on Windows was not observed.

## 7. Not verified (Gaps)

1. The destructive-target registry test and the full `internal/cli` suite were not run in this plan phase.
2. F21 was not re-measured; the audit's probe A3-P3 is cited. The claim that the current binary opens project-local `.moai/state/**/*.db` only in migration reads rests on a partial grep (`.db"` literals), not on a complete search.
3. No real mo.ai.kr run exists after the change; the retained-excess count (decision-index Q4, closed: Q2 is ruled) is unmeasured; AC-UBR-06 prints it. The audit's model estimate (30 of 42) was not re-derived.
4. The 10-14 snapshot is four days older than this tree; the live project may have changed.
5. Windows behavior of the project-root handle operations and of hard links was not observed.
6. A user file named `backup_metadata.json` under `.moai/config/` is overwritten by the writer's metadata inside the config backup (observed by reading `backup.go:28-150`, not reproduced); it is outside this SPEC.
7. Whether t1687's F37 (wiring pin) and the other uncarried findings are fixed on this tree was not checked.
8. Probe test sources run as external test packages through an overlay; they were not run with `-race`.

## Appendix A — probe sources (as executed)

### A.1 `backup_probe_test.go` (target: `internal/cli/update/backup/zz_t1733_probe_test.go`)

```go
package backup_test

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"testing"
	"testing/fstest"
	"time"

	"github.com/modu-ai/moai-adk/internal/cli/update/backup"
	"github.com/modu-ai/moai-adk/internal/cli/update/deploy"
)

func mustWrite(t *testing.T, p, body string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(p), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(p, []byte(body), 0o644); err != nil {
		t.Fatal(err)
	}
}

func remaining(t *testing.T, root string) []string {
	t.Helper()
	ents, err := os.ReadDir(root)
	if err != nil {
		return nil
	}
	var out []string
	for _, e := range ents {
		if e.IsDir() {
			out = append(out, e.Name())
		}
	}
	sort.Strings(out)
	return out
}

// P1: the real deploy producer puts a user skill into pre-clean; five newer backups follow.
func TestT1733ProbeP1(t *testing.T) {
	root := t.TempDir()
	mustWrite(t, filepath.Join(root, ".claude/skills/moai-user-notes/SKILL.md"), "user notes\n")
	mustWrite(t, filepath.Join(root, ".claude/skills/moai-foundation-core/SKILL.md"), "tmpl\n")
	tmplFS := fstest.MapFS{".claude/skills/moai-foundation-core/SKILL.md": {Data: []byte("tmpl\n")}}
	if err := deploy.CleanMoaiManagedPaths(root, io.Discard, tmplFS); err != nil {
		t.Fatal(err)
	}
	old := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	for _, name := range remaining(t, filepath.Join(root, ".moai-backups")) {
		_ = os.Chtimes(filepath.Join(root, ".moai-backups", name), old, old)
	}
	count := func() int {
		n := 0
		_ = filepath.WalkDir(root, func(p string, d fs.DirEntry, err error) error {
			if err == nil && d.Type().IsRegular() {
				if b, rerr := os.ReadFile(p); rerr == nil && string(b) == "user notes\n" {
					n++
				}
			}
			return nil
		})
		return n
	}
	before := count()
	for i := 1; i <= 5; i++ {
		d := filepath.Join(root, ".moai-backups", fmt.Sprintf("2026101%d_000000", i))
		mustWrite(t, filepath.Join(d, "x.yaml"), "x\n")
		ts := time.Now().Add(time.Duration(i) * time.Minute)
		_ = os.Chtimes(d, ts, ts)
	}
	deleted := backup.CleanupOldBackups(root, 5)
	t.Logf("P1 pre-clean-dirs-before=%v user-file-copies-before=%d deleted=%d user-file-copies-after=%d", remaining(t, filepath.Join(root, ".moai-backups")), before, deleted, count())
}

func nsEntry(t *testing.T, root string, n int, version string) string {
	t.Helper()
	name := fmt.Sprintf("update-2026-10-%02dT00-00-00Z", n)
	d := filepath.Join(root, ".moai", "backups", name)
	mustWrite(t, filepath.Join(d, ".claude/agents/x.md"), version+"\n")
	ts := time.Date(2026, 10, n, 0, 0, 0, 0, time.UTC)
	if err := os.Chtimes(d, ts, ts); err != nil {
		t.Fatal(err)
	}
	return name
}

// P2: one file backed up as v1, v2, v3, v4, v4; keep 1.
func TestT1733ProbeP2(t *testing.T) {
	root := t.TempDir()
	for i, v := range []string{"v1", "v2", "v3", "v4", "v4"} {
		nsEntry(t, root, i+1, v)
	}
	deleted := backup.CleanupOldBackups(root, 1)
	rem := remaining(t, filepath.Join(root, ".moai", "backups"))
	has := func(prefix string) bool {
		for _, r := range rem {
			if strings.Contains(r, prefix) {
				return true
			}
		}
		return false
	}
	t.Logf("P2 deleted=%d remaining=%v v1-entry-survives=%v v2-entry-survives=%v v3-entry-survives=%v", deleted, rem, has("-01T"), has("-02T"), has("-03T"))
}

// P3: ten backups holding v1, one newer backup holding v2; keep 1.
func TestT1733ProbeP3(t *testing.T) {
	root := t.TempDir()
	for i := 1; i <= 10; i++ {
		nsEntry(t, root, i, "v1")
	}
	nsEntry(t, root, 11, "v2")
	deleted := backup.CleanupOldBackups(root, 1)
	rem := remaining(t, filepath.Join(root, ".moai", "backups"))
	t.Logf("P3 deleted=%d remaining=%v v1-holding-entries-remaining=%d", deleted, rem, len(rem)-1)
}

type fkey struct{ path, sum string }

func collect(root string, skipNames map[string]bool) map[fkey]bool {
	keys := map[fkey]bool{}
	ents, err := os.ReadDir(root)
	if err != nil {
		return keys
	}
	for _, e := range ents {
		if !e.IsDir() {
			continue
		}
		entry := filepath.Join(root, e.Name())
		_ = filepath.WalkDir(entry, func(p string, d fs.DirEntry, err error) error {
			if err != nil || !d.Type().IsRegular() {
				return nil
			}
			rel, _ := filepath.Rel(entry, p)
			if skipNames[rel] || strings.HasPrefix(rel, ".template-defaults/") {
				return nil
			}
			b, rerr := os.ReadFile(p)
			if rerr != nil {
				return nil
			}
			s := sha256.Sum256(b)
			keys[fkey{rel, hex.EncodeToString(s[:])}] = true
			return nil
		})
	}
	return keys
}

// Real-data baseline: runs the current CleanupOldBackups on an APFS clone and reports what it kept.
func TestT1733ProbeRealData(t *testing.T) {
	clone := os.Getenv("T1733_CLONE")
	if clone == "" {
		t.Skip("T1733_CLONE not set")
	}
	skip := map[string]bool{".moai-run": true, ".complete": true, "backup_metadata.json": true}
	cfgRoot := filepath.Join(clone, ".moai-backups")
	nsRoot := filepath.Join(clone, ".moai", "backups")
	count := func(r string) (n int) {
		for _, name := range remaining(t, r) {
			_ = name
			n++
		}
		return
	}
	preClean := func() int {
		n := 0
		for _, name := range remaining(t, cfgRoot) {
			if _, err := os.Stat(filepath.Join(cfgRoot, name, "pre-clean")); err == nil {
				n++
			}
		}
		return n
	}
	beforeCfg, beforeNs, beforePre := count(cfgRoot), count(nsRoot), preClean()
	keysBefore := map[fkey]bool{}
	for k := range collect(cfgRoot, skip) {
		keysBefore[k] = true
	}
	for k := range collect(nsRoot, skip) {
		keysBefore[k] = true
	}
	deleted := backup.CleanupOldBackups(clone, 5)
	keysAfter := map[fkey]bool{}
	for k := range collect(cfgRoot, skip) {
		keysAfter[k] = true
	}
	for k := range collect(nsRoot, skip) {
		keysAfter[k] = true
	}
	lostByPath := map[string]int{}
	lost := 0
	for k := range keysBefore {
		if !keysAfter[k] {
			lost++
			lostByPath[k.path]++
		}
	}
	type pc struct {
		p string
		n int
	}
	var top []pc
	for p, n := range lostByPath {
		top = append(top, pc{p, n})
	}
	sort.Slice(top, func(i, j int) bool { return top[i].n > top[j].n })
	t.Logf("REAL cfg-root dirs %d->%d ns-root dirs %d->%d deleted=%d pre-clean-dirs %d->%d", beforeCfg, count(cfgRoot), beforeNs, count(nsRoot), deleted, beforePre, preClean())
	t.Logf("REAL distinct-file-versions-before=%d lost=%d paths-with-lost-versions=%d", len(keysBefore), lost, len(lostByPath))
	for i := 0; i < len(top) && i < 3; i++ {
		t.Logf("REAL lost-top path=%s versions-lost=%d", top[i].p, top[i].n)
	}
}
```

### A.2 `residue_probe_test.go` (target: `internal/cli/update/skillresidue/zz_t1733_probe_test.go`)

```go
package skillresidue_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/modu-ai/moai-adk/internal/cli/update/skillresidue"
)

func wr(t *testing.T, p, body string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(p), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(p, []byte(body), 0o644); err != nil {
		t.Fatal(err)
	}
}

func lexists(p string) bool {
	_, err := os.Lstat(p)
	return err == nil
}

func shipsAll(string) bool { return true }

// P4a: project A's .agents/skills is a link to project B's .agents/skills, which holds B's live link.
func TestT1733ProbeP4a(t *testing.T) {
	base := t.TempDir()
	a, b := filepath.Join(base, "A"), filepath.Join(base, "B")
	for _, p := range []string{a, b} {
		wr(t, filepath.Join(p, ".claude/skills/moai-kanban-foreman/SKILL.md"), "retired\n")
	}
	if err := os.MkdirAll(filepath.Join(b, ".agents/skills"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink("../../.claude/skills/moai-kanban-foreman", filepath.Join(b, ".agents/skills/moai-kanban-foreman")); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(filepath.Join(a, ".agents"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink(filepath.Join(b, ".agents/skills"), filepath.Join(a, ".agents/skills")); err != nil {
		t.Fatal(err)
	}
	actions, err := skillresidue.Clean(a, skillresidue.Options{Shipped: shipsAll})
	t.Logf("P4a err=%v actions=%v b-link-survives=%v a-folder-in-place=%v", err, actions, lexists(filepath.Join(b, ".agents/skills/moai-kanban-foreman")), lexists(filepath.Join(a, ".claude/skills/moai-kanban-foreman/SKILL.md")))
}

// P4b: project A's .moai/archive is a link to a directory outside the project.
func TestT1733ProbeP4b(t *testing.T) {
	base := t.TempDir()
	a, out := filepath.Join(base, "A"), filepath.Join(base, "outside")
	wr(t, filepath.Join(a, ".claude/skills/moai-kanban-foreman/SKILL.md"), "retired\n")
	if err := os.MkdirAll(filepath.Join(a, ".moai"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(out, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink(out, filepath.Join(a, ".moai/archive")); err != nil {
		t.Fatal(err)
	}
	actions, err := skillresidue.Clean(a, skillresidue.Options{Shipped: shipsAll})
	moved := false
	_ = filepath.WalkDir(out, func(p string, d os.DirEntry, e error) error {
		if e == nil && !d.IsDir() {
			moved = true
		}
		return nil
	})
	t.Logf("P4b err=%v actions=%v content-now-outside-project=%v a-folder-in-place=%v", err, actions, moved, lexists(filepath.Join(a, ".claude/skills/moai-kanban-foreman/SKILL.md")))
}
```

### A.3 overlay files (exact text)

`.moai/reports/t1733/probes/overlay_backup.rel.json`:

```json
{"Replace": {"internal/cli/update/backup/zz_t1733_probe_test.go": ".moai/reports/t1733/probes/backup_probe_test.go"}}
```

`.moai/reports/t1733/probes/overlay_residue.rel.json`:

```json
{"Replace": {"internal/cli/update/skillresidue/zz_t1733_probe_test.go": ".moai/reports/t1733/probes/residue_probe_test.go"}}
```

`.moai/reports/t1733/probes/overlay_backup_p5.rel.json` (the probe P5 of A.4 uses helpers of A.1, so this overlay maps both files):

```json
{"Replace": {"internal/cli/update/backup/zz_t1733_probe_test.go": ".moai/reports/t1733/probes/backup_probe_test.go", "internal/cli/update/backup/zz_t1733_probe_p5_test.go": ".moai/reports/t1733/probes/backup_probe_p5_test.go"}}
```

Note (audit D9): the probes of A.1 and A.2 only `t.Logf`; they assert nothing, so they exit 0 and show the measurement rather than a pass or fail. The RED/GREEN judgment is made on the logged values by the acceptance cells; the run-phase tests (AC-UBR-24) add assertions that fail on this tree. The probe P5 of A.4 is the exception: it asserts the clause it measures, so its RED is exit 1 (audit iteration 2, B2).

### A.4 `backup_probe_p5_test.go` — fixed-mtime real-data prune (AC-UBR-06(d))

Placed at `.moai/reports/t1733/probes/backup_probe_p5_test.go`; run with `-overlay=.moai/reports/t1733/probes/overlay_backup_p5.rel.json`. It stamps every run directory of the clone with the timestamp its directory name encodes (distinct, explicit, set after the copy), prunes at `T1733_KEEP` (default 4), and asserts that at least one file under the config root's `pre-clean/` copy remains. Line numbers in the acceptance cell's output refer to this text as executed.

```go
package backup_test

import (
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/modu-ai/moai-adk/internal/cli/update/backup"
)

// P5: real-data prune on a clone whose run directories all carry a DISTINCT, explicitly fixed
// modification time derived from the directory name, so ordering never falls back to the
// path-ascending tie-break. Reports whether the files under the config root's pre-clean/ copy
// survive a prune at T1733_KEEP (default 4).
func TestT1733ProbeRealDataFixedMtime(t *testing.T) {
	clone := os.Getenv("T1733_CLONE")
	if clone == "" {
		t.Skip("T1733_CLONE not set")
	}
	keep := 4
	if v := os.Getenv("T1733_KEEP"); v != "" {
		n, err := strconv.Atoi(v)
		if err != nil {
			t.Fatal(err)
		}
		keep = n
	}
	cfgRoot := filepath.Join(clone, ".moai-backups")
	nsRoot := filepath.Join(clone, ".moai", "backups")
	cfgName := regexp.MustCompile(`^[0-9]{8}_[0-9]{6}$`)
	nsName := regexp.MustCompile(`^(update|v2-to-v3)-([0-9]{4}-[0-9]{2}-[0-9]{2}T[0-9]{2}-[0-9]{2}-[0-9]{2})Z(-[0-9]+)?$`)

	stamps := map[string]time.Time{}
	for _, name := range remaining(t, cfgRoot) {
		if cfgName.MatchString(name) {
			ts, err := time.Parse("20060102_150405", name)
			if err != nil {
				t.Fatal(err)
			}
			stamps[filepath.Join(cfgRoot, name)] = ts
		}
	}
	for _, name := range remaining(t, nsRoot) {
		m := nsName.FindStringSubmatch(name)
		if m == nil {
			continue
		}
		ts, err := time.Parse("2006-01-02T15-04-05", m[2])
		if err != nil {
			t.Fatal(err)
		}
		if m[3] != "" {
			n, _ := strconv.Atoi(strings.TrimPrefix(m[3], "-"))
			ts = ts.Add(time.Duration(n) * time.Second)
		}
		stamps[filepath.Join(nsRoot, name)] = ts
	}
	seen := map[time.Time]string{}
	for p, ts := range stamps {
		if prev, dup := seen[ts]; dup {
			t.Fatalf("mtime collision %s vs %s at %s", prev, p, ts)
		}
		seen[ts] = p
		if err := os.Chtimes(p, ts, ts); err != nil {
			t.Fatal(err)
		}
	}

	var preCleanRun string
	for _, name := range remaining(t, cfgRoot) {
		if _, err := os.Stat(filepath.Join(cfgRoot, name, "pre-clean")); err == nil {
			preCleanRun = filepath.Join(cfgRoot, name)
		}
	}
	if preCleanRun == "" {
		t.Fatal("no pre-clean run in the clone")
	}
	rank := 1
	for p, ts := range stamps {
		if p != preCleanRun && ts.After(stamps[preCleanRun]) {
			rank++
		}
	}
	oldestConfig := true
	for p, ts := range stamps {
		if strings.HasPrefix(p, cfgRoot+string(filepath.Separator)) && ts.Before(stamps[preCleanRun]) {
			oldestConfig = false
		}
	}
	preCleanFiles := func() int {
		n := 0
		_ = filepath.WalkDir(cfgRoot, func(p string, d fs.DirEntry, err error) error {
			if err == nil && d.Type().IsRegular() && strings.Contains(p, string(filepath.Separator)+"pre-clean"+string(filepath.Separator)) {
				n++
			}
			return nil
		})
		return n
	}
	before := preCleanFiles()
	deleted := backup.CleanupOldBackups(clone, keep)
	_, statErr := os.Stat(preCleanRun)
	t.Logf("P5 keepCount=%d runs=%d distinct-run-mtimes=true pre-clean-run=%s pre-clean-run-rank=%d pre-clean-run-is-oldest-config-run=%v", keep, len(stamps), filepath.Base(preCleanRun), rank, oldestConfig)
	after := preCleanFiles()
	t.Logf("P5 pre-clean-files before=%d after=%d deleted=%d pre-clean-run-exists-after=%v", before, after, deleted, statErr == nil)
	if before == 0 || after < 1 {
		t.Errorf("AC-UBR-06(d): pre-clean files before=%d after=%d, want before>0 and after>=1", before, after)
	}
}
```

## Principle check (operator principle for 3.2: add no data-deleting behavior)

- REQ-UBR-001 to REQ-UBR-014 change the pruner from "delete every run outside the window" to "delete fewer": every deletion the new pruner makes, the present pruner also makes. No deletion is added.
- REQ-UBR-015, REQ-UBR-016: report-only; no move and no removal.
- REQ-UBR-017: the lowercase `skill.md` rename is a move that exists today; it is narrowed (bound to the project root, never overwrites). Not a deletion.
- REQ-UBR-018: removes the zero-byte database deletion; reports instead.
- Residual: the pruner still deletes excess runs that protect no budgeted version. That deletion exists on main and on this tree and is the behavior the operator's budget ruling (a263 C) governs.
