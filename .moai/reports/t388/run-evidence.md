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

M1 commit: `<backfilled>`

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
