# t388 run-phase evidence — SPEC-VERSION-STAMP-GUARD-001

Worktree `.claude/worktrees/t388`, branch `WT-version-sync-list`.

## Preflight (plan.md §C)

| Check | Command | Output | Verdict |
|---|---|---|---|
| C.1 HEAD vs measurement tree | `git diff --stat 9328a5242 HEAD -- <8 pinned paths>` | (empty) | measured paths identical to `9328a5242`; §B figures hold |
| C.2 name collision | `test -e internal/cli/version_sync_list_test.go` | exit 1 | absent |
| C.3 list formatting | `awk 'NR>=60 && NR<=95' .moai/docs/version-management.md` | `- README.md (Version line)` form, no backticks on stamp paths; lines 73-74 carry backticks inside the parenthetical | parser strips the parenthetical and any backticks |

Preflight HEAD: `6854a9306`. Working tree clean (`git status --short` empty).

---

## M1 — the check lands before any documentation edit

### Expectation, written BEFORE measurement

The anchor constant `**Version Stamps:**` points at a subheading **M2 creates**. At M1 the
document does not carry it, so the section scan yields nothing.

- Count assertion **fires**. Expected output literal: `version-stamp entries: parsed=0 expected=7`.
  Cause: **the anchor subheading is absent**, so the parse is empty.
- Existence assertion is **silent** — it iterates an empty entry list and sees no path at all.
  This red says nothing about the existence assertion and is NOT its evidence.
- The ghost (`internal/template/templates/.moai/config/config.yaml`, doc line 78) sits under
  `**Configuration Files:**`, which the anchor-scoped scan never reads. **This red is not
  "the check caught the ghost".**

This is AC-VSG-005's RED evidence (plan.md E3-a).

### Observation — E3-a

Tree: HEAD `6854a9306` + the untracked `internal/cli/version_sync_list_test.go`. The M1 commit SHA
is backfilled below once the commit exists; the run is the pre-commit working tree, whose test file
is byte-identical to what that commit carries.

M1 commit: `f270d2df5`

```
$ go test ./internal/cli/... -run TestVersionSyncList -v      # exit 1
=== RUN   TestVersionSyncListNamesOnlyExistingPaths
    version_sync_list_test.go:83: version-stamp entries: parsed=0 expected=7 (anchor "**Version Stamps:**" in .moai/docs/version-management.md)
--- FAIL: TestVersionSyncListNamesOnlyExistingPaths (0.00s)
FAIL	github.com/modu-ai/moai-adk/internal/cli	4.379s
```

Full output: `.moai/reports/t388/m1-red.log`.

**Expected vs observed: identical.** The pinned literal `version-stamp entries: parsed=0 expected=7`
appears verbatim.

**Cause: the anchor subheading is absent.** Not the ghost. Two things establish this:

1. The existence assertion produced no output at all —
   `grep -c "does not exist" .moai/reports/t388/m1-red.log` → `0`, exit 1. It named no path,
   so this run asserts nothing about it (`verification-completeness.md` §1.1).
2. The ghost sits at doc line 78, under `**Configuration Files:**`. The scan starts only after
   `**Version Stamps:**`, which this tree's document does not contain, so the scan never reached
   that section. A parse of 0 entries cannot have "caught" an entry.

Supporting checks on the same tree:

```
$ gofmt -l internal/cli/version_sync_list_test.go   # (empty) exit 0
$ go vet ./internal/cli/...                         # exit 0, no output
```

---

## M2.1 — subheadings created, omissions filled, ghost deliberately left in place

### Expectation, written BEFORE measurement

Replacing the documentation/configuration axis with the stamp/artifact axis moves both entries
under `**Configuration Files:**` into the stamp list, and one of them is the ghost. The stamp list
therefore holds **8** entries, not 7.

- Count assertion fires: `version-stamp entries: parsed=8 expected=7`.
- Existence assertion fires and names `internal/template/templates/.moai/config/config.yaml`.
- Both appear in **one** run, because reporting is non-fatal.

**Two causes.** This is a supporting observation (plan.md E3-b) and is the single evidence of
neither AC — AC-VSG-004 needs a red whose only cause is the existence assertion, which M2.3
produces.

### Observation — E3-b

Tree: M1 commit `f270d2df5` + the M2.1 documentation edit (committed as the M2.1 commit below).

```
$ go test ./internal/cli/ -run TestVersionSyncList -v      # exit 1
=== RUN   TestVersionSyncListNamesOnlyExistingPaths
    version_sync_list_test.go:83: version-stamp entries: parsed=8 expected=7 (anchor "**Version Stamps:**" in .moai/docs/version-management.md)
    version_sync_list_test.go:91: version-sync list names a path that does not exist: internal/template/templates/.moai/config/config.yaml
--- FAIL: TestVersionSyncListNamesOnlyExistingPaths (0.00s)
FAIL	github.com/modu-ai/moai-adk/internal/cli	0.865s
```

Full output: `.moai/reports/t388/m2-1-red.log`.

**Expected vs observed: identical**, both literals verbatim. This is the only moment the check
meets the real ghost, which is why the step is committed rather than folded into M2.2.

M2.1 commit: `d595faa9d`

---

## M2.2 — the ghost is removed

### Expectation, written BEFORE measurement

Seven entries remain and all exist, so both assertions go silent and the test passes.

### Observation — E4

Tree: M2.1 commit `d595faa9d` + the one-line deletion (committed as the M2.2 commit below).

```
$ go test ./internal/cli/ -run TestVersionSyncList -v      # exit 0
=== RUN   TestVersionSyncListNamesOnlyExistingPaths
--- PASS: TestVersionSyncListNamesOnlyExistingPaths (0.00s)
ok  	github.com/modu-ai/moai-adk/internal/cli	0.908s
```

Full output: `.moai/reports/t388/m2-2-green.log`.

A pass prints nothing, so the parsed count is corroborated two ways rather than assumed:

```
$ sed -n '/^\*\*Version Stamps:\*\*$/,/^\*\*Release Artifacts:\*\*$/p' \
    .moai/docs/version-management.md | grep -c '^- '
7
```

and by M2.3 below, where the count assertion stays silent while the existence assertion fires —
which it could not do if the parse had drifted off 7.

M2.2 commit: `d7af3d22d`

---

## M2.3 — the existence assertion's single-cause RED

### Expectation, written BEFORE measurement

One stamp line is **substituted**, not added: `docs-site/hugo.toml` →
`docs-site/nonexistent-stamp.toml`. The entry count stays 7, so the count assertion stays silent
and the only cause of the red is the existence assertion.

- Existence assertion fires with the pinned literal:
  `version-sync list names a path that does not exist: docs-site/nonexistent-stamp.toml`
- **No `parsed=` line appears.** If one did, the edit was an addition rather than a substitution
  and the observation would not be this AC's evidence.
- Reverting the substitution returns the check to green.

This is AC-VSG-004's RED → GREEN evidence (plan.md E3-c).

### Observation — E3-c

Tree: M2.2 commit `d7af3d22d` + the one-token substitution in the working tree. The substituted
state is deliberately **not committed**; the revert is proven below by `git status`, not by eye.

Precondition — the planted path is genuinely absent:

```
$ test -e docs-site/nonexistent-stamp.toml      # exit 1
```

RED:

```
$ go test ./internal/cli/ -run TestVersionSyncList -v      # exit 1
=== RUN   TestVersionSyncListNamesOnlyExistingPaths
    version_sync_list_test.go:91: version-sync list names a path that does not exist: docs-site/nonexistent-stamp.toml
--- FAIL: TestVersionSyncListNamesOnlyExistingPaths (0.00s)
FAIL	github.com/modu-ai/moai-adk/internal/cli	0.892s
```

Single cause confirmed mechanically:

```
$ grep -c "parsed=" .moai/reports/t388/m2-3-red.log
0                                               # rc=1 — no count line, as pinned
```

GREEN after revert:

```
$ go test ./internal/cli/ -run TestVersionSyncList -v      # exit 0
--- PASS: TestVersionSyncListNamesOnlyExistingPaths (0.00s)
$ git status --short                                        # (empty) — tree byte-identical to d7af3d22d
```

**Expected vs observed: identical** on all three points (literal, absent count line, return to
green). Logs: `m2-3-red.log`, `m2-3-green.log`.

The two mutants AC-VSG-004 names are dead: an always-empty implementation cannot produce the RED,
an always-failing one cannot produce the GREEN, and one that fails without naming the path is
killed by the literal grep above.

---

## Which red belongs to which assertion

| Observation | Tree | Assertion that fired | Cause | Serves |
|---|---|---|---|---|
| E3-a | `6854a9306` + test | count only (`parsed=0 expected=7`) | anchor subheading absent — the scan read nothing | AC-VSG-005 RED |
| E3-b | `d595faa9d` | both (`parsed=8 expected=7` + ghost named) | two: the extra entry, and the ghost | supporting only, neither AC's single evidence |
| E4 | `d7af3d22d` | neither (pass) | list correct at 7, all present | AC-VSG-005 GREEN |
| E3-c | `d7af3d22d` + substitution | existence only | one: the planted path | AC-VSG-004 RED → GREEN |

No red is used as evidence for an assertion other than the one that produced it.

---

## AC matrix

| AC | Judgment command | Output | Verdict |
|---|---|---|---|
| AC-VSG-001 | `grep -c '<ghost>' .moai/docs/version-management.md` | `0`, rc=1 | PASS — ghost gone |
| AC-VSG-001 | stamp bullets, parenthetical stripped, sorted | the 7 literal paths, both set differences empty | PASS |
| AC-VSG-002 | stamp section grepped for `CHANGELOG.md\|release-notes` | `0`, rc=1 | PASS |
| AC-VSG-002 | `comm -12` over the two sections' path sets | (empty) | PASS — disjoint |
| AC-VSG-002 | release-artifact heading states a bump does not touch it | prose above the two lists | PASS |
| AC-VSG-003 (1) | `grep -nE 'reads from git tags at build time\|via .git describe.'` | rc=1 | PASS — 0 derived-value assertions |
| AC-VSG-003 (2) | `grep -n 'Makefile:20\|goreleaser.yml:22\|fallback'` | lines 8, 12, 13, 80 | PASS — fallback stated, both injection points cited |
| AC-VSG-003 (3) | `grep -niE 'constant\|상수'` | rc=1 | PASS |
| AC-VSG-004 | E3-c above | RED names the path, GREEN on revert | PASS |
| AC-VSG-005 | E3-a (`parsed=0`) → E4 (pass at 7) | pinned literal, then silence | PASS |
| AC-VSG-006 (a) | `grep -c` for `partial` / `does not detect` over the section | `1` / `1` | PASS |
| AC-VSG-006 (2) | `grep -c t392` over the section | `1` | PASS |
| AC-VSG-006 (b) | `grep -cF` over the five rejection literals | `0`, rc=1 | PASS — none present |

## Scope and closing checks

| Check | Command | Output |
|---|---|---|
| Files changed by the card | `git diff --name-only 6854a9306 HEAD` | 1 Go file, 1 doc, 1 SPEC frontmatter, 6 evidence files |
| New packages | `git diff --name-only --diff-filter=A ... -- internal/` | `internal/cli/version_sync_list_test.go` only — existing package, no new directory |
| New CI job / template mirror | `git diff --name-only ... -- .github/ internal/template/ .claude/` | (empty) — Template-First does not apply, verified rather than assumed |
| Package suite | `go test ./internal/cli/...` | exit 0, all 17 packages ok (`final-test.log`) |
| Vet | `go vet ./internal/cli/...` | exit 0 (`final-vet.log`) |
| Format | `gofmt -l internal/cli/version_sync_list_test.go` | (empty), exit 0 |

Not run locally: `go test ./...`. The full-suite verdict is CI's on the pushed head
(`CLAUDE.local.md` §4).

## Residual risk, updated from run-phase observation

- **R-4 / R-5 held up under measurement.** The M1 run is exactly the shape R-4 warns about — a
  parse that matched nothing — and the count assertion turned it into a failure rather than a
  silent pass. What stays open is R-5's human half: if the list legitimately grows to eight
  entries, `expectedVersionStampEntries` has to move with it, and nothing enforces that pairing.
- **The half this card does not close is unchanged.** An omitted stamp site is still invisible;
  card t392 owns it.
