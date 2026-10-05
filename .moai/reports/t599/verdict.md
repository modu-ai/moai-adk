# Card t599 Verdict — Hooks Audit H04: bulk deletion inside a directory named `{}`

- **Card**: t599 (Class B, no SPEC) — hooks audit 2026-09-11 finding H04, P2, operator-selected
- **Branch**: `WT-braces-dir-delete` (worktree `.claude/worktrees/t599`)
- **Base**: `16f3b8a81` (local develop at dispatch time)
- **Defect surface**: `cleanupBogusRootDir`, `internal/hook/session_end.go` (was lines 900-935 at base)
- **Audit provenance**: `reports/hooks-audit-20260911-01a08e35/hooks-audit.md` H04; baseline audit tree `2213871af` (pre-reproduction re-check was mandatory and done — defect confirmed live at base)

## Claim

`cleanupBogusRootDir` no longer deletes a project-root `{}` directory wholesale. It now treats `{}` as bug residue only when the MoAI evidence signature (`.claude/agent-memory` subdirectory, the package's `agentMemorySegment` convention) is present, removes only that marked residue subtree, preserves every other content with a warning, and removes the now-empty `{}`/`{}/.claude` shells best-effort only.

## Evidence

1. **Audit-side reproduction (pre-existing, cited)** — `reports/hooks-audit-20260911-01a08e35/probes-go-confirmed.log`, `TestAuditBogusDirectoryData`: `unattributed_user_file_deleted=true` after writing `{}/user-data.txt` and invoking the cleanup.
2. **RED on the pre-fix tree (this run)** — `go test ./internal/hook/ -run 'TestCleanupBogusRootDir' -count=1 -v` at HEAD `16f3b8a81` with only the new tests present: **3 FAIL** (`PreservesUnmarkedUserDir` — "unmarked {} directory should have been preserved" + "user file inside unmarked {} directory should have been preserved" — plus `MixedContentRemovesResidueOnly`, `EmptyDirPreserved`), 6 PASS. RED is for the stated reason: whole-directory deletion of unattributed content. (Command output observed by manager-develop on the pre-fix tree, same worktree; the failing assertion text quoted verbatim above.)
3. **GREEN after the fix (re-measured by the lane lead, this run)**:

```
go test ./internal/hook/ -run 'TestCleanupBogusRootDir' -count=1 -v
--- PASS: TestCleanupBogusRootDir_IgnoresFile (0.00s)
--- PASS: TestCleanupBogusRootDir_NoDirectory (0.00s)
--- PASS: TestCleanupBogusRootDir_NoClaudeDir (0.00s)
--- PASS: TestCleanupBogusRootDir_EmptyDirPreserved (0.00s)
--- PASS: TestCleanupBogusRootDir_IgnoresSymlink (0.00s)
--- PASS: TestCleanupBogusRootDir_PreservesUnmarkedUserDir (0.00s)
--- PASS: TestCleanupBogusRootDir_RemovesMarkedResidue (0.00s)
--- PASS: TestCleanupBogusRootDir_RemovesDirectory (0.00s)
--- PASS: TestCleanupBogusRootDir_MixedContentRemovesResidueOnly (0.00s)
PASS
ok  	github.com/modu-ai/moai-adk/internal/hook	0.483s
```

Selector matched **9 tests** (RUN lines above — non-empty sweep; 4 new + 5 pre-existing including `NoClaudeDir` which lives in `session_end_extra_coverage_test.go`).

4. **Affected package** — `go test ./internal/hook/ -count=1` → `ok ... 172.809s` (manager-develop, this tree); `go vet ./internal/hook/` → exit 0, no output (re-observed by lane lead).
5. **Diff scope** — `git diff --stat`: `internal/hook/session_end.go` (+~38/−9, `cleanupBogusRootDir` body + doc comment only), `internal/hook/session_end_test.go` (+113, 4 new tests, no existing tests modified). Call site at line 107 untouched. Shared-file note (H02/H03): no other regions touched, so sibling cards' surfaces are preserved at this base.

## Completion-condition mapping (card text)

| Card condition | Result |
|---|---|
| Same-named user directory preserved (no marker) | ✅ T1 `PreservesUnmarkedUserDir` — `{}/user-data.txt`-only dir survives; audit reproduction shape now green |
| Only explicitly-marked residue cleaned | ✅ T2 `RemovesMarkedResidue` + T3 `MixedContentRemovesResidueOnly` — residue subtree removed, co-located user file + notes/ preserved |
| Normal control group | ✅ T2 (residue → removed, empty shells cleaned) + pre-existing `RemovesDirectory`/`NoDirectory`/`IgnoresFile`/`NoClaudeDir`/`IgnoresSymlink` unchanged-green |
| Failure reproduction group | ✅ T1 (RED observed pre-fix at `16f3b8a81`, GREEN post-fix) |
| Command + output + target commit as evidence | ✅ commands and verbatim outputs in this verdict; target commit recorded below |
| Reproduction isolation (dispatch HARD rule) | ✅ all 4 new tests use `t.TempDir()`; no real directory was touched by any reproduction |

## Baseline-attribution

All measurements in this run against worktree `/Users/goos/MoAI/moai-adk-go/.claude/worktrees/t599`, branch `WT-braces-dir-delete`; RED at HEAD `16f3b8a81` (pre-fix), GREEN on the same tree with the uncommitted fix applied, then committed (see commit SHA in the git log of this branch — verdict file committed alongside the fix).

## Gaps

- Full-suite (`go test ./...`) and CI verdict NOT run locally — per lane-local verification discipline this is CI's verdict on the batched develop push; the lane did not push (dispatch HARD rule).
- `-race` variant of the selector not run — no concurrency introduced; function is single-threaded file ops.
- The marker-based design cannot distinguish a user directory that itself contains exactly `{}/.claude/agent-memory` from true bug residue — accepted scope boundary of the card's evidence-signature gate; such content inside the marked subtree is removed (T3 covers preservation only OUTSIDE the marked subtree).

## Residual-risk

- A mixed `{}` directory produces two "preserved content" warnings (`.claude` shell + `{}` shell) — cosmetic, evidence-preserving by design.
- `os.Stat` follows symlinks for the evidence gate; `RemoveAll` on the residue path removes the symlink entry itself rather than following it, so a symlinked `.claude/agent-memory` under `{}` deletes only the link — no escape, but the warn-vs-delete boundary for that exotic case is untested.
- Stale-binary hazard: users on installed binaries older than this fix keep the old wholesale-removal behavior until they update (§1.3 continued-firing shape; deployment, not code).

🗿 MoAI
