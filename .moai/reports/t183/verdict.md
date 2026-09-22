# t183 — `filepath.Abs` does not resolve symlinks (t171 sync-audit F-6)

Card class B, Tier S. Branch `WT-symlink-bound`, based on `release/v3.1.3`.

---

## Claim

1. `validateProjectRoot` (`internal/cli/mcp_project_root.go`) now canonicalizes an
   accepted `project_root` with `filepath.EvalSymlinks`, so the root a tool acts on is
   the real directory rather than whichever spelling reached the handler.
2. A canonicalization failure is a rejection, on the same terms as the other
   unusable-path rejections — never a silent fallback to the uncanonicalized path.
3. A boundary regression test pins the behaviour: a symlinked project root resolves to
   its canonical target.
4. The constraint is stated where a caller reads it — the tool's `project_root` input
   description and `moai-mcp-tools.md` § *The `project_root` input* (plus its template
   mirror).

What this does **not** claim: that a containment boundary exists today. It does not.
F-6's point is that the omission is harmless only while nothing compares this result
against another path, and that the comparison the follow-up would add (root must sit
under the repository's git common dir) would be walked straight through by a symlink.
This card removes that trap ahead of the constraint; it does not add the constraint.

---

## Evidence

### The defect, observed (RED, before the fix)

```
$ go test ./internal/cli/ -run 'TestValidateProjectRoot_' -count=1
--- FAIL: TestValidateProjectRoot_ResolvesSymlinkedRoot (0.00s)
    mcp_project_root_test.go:289: root="/var/folders/kt/.../002/linked-root",
      want the canonical target "/private/var/folders/kt/.../001"
      (symlink was "/var/folders/kt/.../002/linked-root")
FAIL	github.com/modu-ai/moai-adk/internal/cli	0.815s
```

The returned root is the symlink path verbatim. This is the F-6 observation reproduced
as a failing test, not inferred from reading the code.

### After the fix (GREEN)

```
$ go test ./internal/cli/ -run 'TestValidateProjectRoot_|TestResolveToolProjectRoot_|TestSpec' -count=1
ok  	github.com/modu-ai/moai-adk/internal/cli	1.835s
```

### The change

`internal/cli/mcp_project_root.go` — `validateProjectRoot`: after the existing
`Abs` / `Stat` / `.moai` checks' exists-and-is-a-directory half, the path is
canonicalized and the `.moai` probe plus the return value both use the canonical form.

```go
canonical, err := filepath.EvalSymlinks(abs)
if err != nil {
    return "", fmt.Errorf("project_root %q cannot be canonicalized: %w", raw, err)
}
moaiDir := filepath.Join(canonical, ".moai")
...
return canonical, nil
```

Error messages keep naming `raw` — the path the caller actually typed — rather than the
resolved one it never saw.

### A regression the targeted run missed, and the full-package run caught

The first full `internal/cli` run failed — on two tests my `-run` filter had excluded:

```
--- FAIL: TestCodexAudit_ProjectRootLandsInTheParamsMap (0.08s)
    mcp_project_root_codex_test.go:209: codex params cwd =
      "/private/var/folders/.../001", want the named tree "/var/folders/.../001"
--- FAIL: TestAuditMulti_ProjectRootLandsInTheCodexParamsMap (0.00s)
    mcp_project_root_codex_test.go:269: fan-out codex params cwd =
      "/private/var/folders/.../001", want the named tree "/var/folders/.../001"
FAIL	github.com/modu-ai/moai-adk/internal/cli	360.292s
```

Both compare a handler's answer against the raw `t.TempDir()` value, which on macOS is
itself a symlinked path (`/var` → `/private/var`). Canonicalization made the two sides
two spellings of one directory, so they failed on the spelling, not on the tree.

Repaired at the single shared source: `newProbeProject` now returns the canonical path,
with the reason in a comment. No assertion was weakened — each still demands that the
handler's cwd BE the probe tree; the probe tree is now named the way the handler names
it. The alternative (relaxing each comparison to "same directory") would have removed
the property the tests exist to hold.

Recorded because the miss is mine: a `-run`-filtered run is a scoping choice, and this
one excluded the tests most likely to observe the change. The full-package run is what
turned that into a caught regression rather than a shipped one.

### Verification batch (this worktree, this tree)

| Command | Observed |
|---|---|
| `gofmt -l internal/cli/mcp_project_root.go internal/cli/mcp_project_root_test.go` | no output |
| `go build ./...` | exit 0 |
| `go vet ./internal/cli/` | exit 0 |
| `GOOS=windows GOARCH=amd64 go vet ./internal/cli/` | exit 0 |
| `GOOS=linux GOARCH=amd64 go vet ./internal/cli/` | exit 0 |
| `golangci-lint run ./internal/cli/...` | `0 issues.` |
| `go test ./internal/cli/ -count=1 -timeout 900s` | `ok … 368.636s`, exit 0, `grep -c '^--- FAIL'` → `0` |
| `MOAI_TEMPLATE_LEAK_STRICT=1 go test ./internal/template/... -count=1` | `ok … 21.114s` (+ agentemit `ok`) |
| `go test ./internal/config/ -run 'TestAlwaysLoadedTokenBudget$'` | `67016 tokens (budget 76000, headroom 8984)` — PASS |
| `git diff --stat` | 4 files, +89 −6 |

---

## Baseline-attribution

Every figure above was measured by me, in `.claude/worktrees/t183`, on branch
`WT-symlink-bound`, against the tree that carries the final diff.

**One rebase happened after the full suite ran, and is accounted for.** The branch was
first built by merging `origin/release/v3.1.3` into a `origin/main`-based worktree
(HEAD `db09e6885`); merging that back into release would have carried seven unrelated
main-side files (3 SPEC docs, 4 docs-site pages) into the release branch, so the merge
was undone (`git reset --merge`) and the card commit was rebased onto
`origin/release/v3.1.3` (HEAD `3c2480d1e`). The Go tree is provably unchanged across
that move — `git diff --stat db09e6885 3c2480d1e -- '*.go' go.mod go.sum` reports
exactly the two files this card edits and nothing else — so the 368s full-suite result
attributes to `3c2480d1e`'s Go tree. Re-run on `3c2480d1e` directly: `go build ./...`
exit 0, `go vet ./internal/cli/` exit 0, and the project_root test set
`ok … 1.938s`. The RED transcript is from the same tree with only the fix hunk absent —
the test file was written first, run, and observed failing before
`mcp_project_root.go` was touched.

The always-loaded token figure moved 66926 → 67016 (+90) for the six added doc lines
and their mirror; the budget is 76000, so the headroom claim is measured, not assumed.

Nothing was carried over from the t171 run. The t171 sync-audit is cited as the source
of the finding, not as evidence for any measurement here.

---

## Gaps — what was NOT observed

- **No containment boundary was built or tested.** The git-common-dir constraint F-6
  anticipates remains unwritten; this card only removes the symlink hole ahead of it.
  A test proving "a symlink cannot escape the boundary" cannot exist until the boundary
  does.
- **`TestValidateProjectRoot_RejectsBrokenSymlink` was green before the fix.** `os.Stat`
  already rejected a dangling link. It is a characterization test pinning the
  Stat-before-EvalSymlinks ordering (so a future reorder does not silently change the
  error text), NOT a demonstration of a repaired defect. Stated here so its green is
  not read as evidence it caught something.
- **Symlink behaviour on Windows was not executed**, only compiled (`GOOS=windows go vet`).
  Per the standing lesson, vet proves compilation and nothing about behaviour. Windows
  `EvalSymlinks` additionally normalizes case, which no test observes. CI's Windows job
  is the measurement.
- **The absent-parameter path is deliberately NOT canonicalized.** `resolveToolProjectRoot`
  still returns `resolveProjectDir()` unchanged when the parameter is absent, because
  REQ-2 of SPEC-MCP-WORKTREE-ROOT-001 requires the absent case to resolve exactly as
  before, and `TestResolveToolProjectRoot_AbsentAndEmptyMatchTheDefault` enforces it.
  So the canonicalization guarantee covers caller-supplied roots only.
- **No live MCP-server probe was run.** Per t171 F-7, a running `moai mcp-server` keeps
  its old build until restarted; nothing here was verified through a live tool call.
- **Downstream consumers were not re-probed individually.** `mcp_codex.go`,
  `mcp_glm.go`, `mcp_server.go`, and `mcp_audit_multi.go` were read (a canonical
  absolute path is valid everywhere the previous value was), but their behaviour with a
  symlinked root was not separately exercised beyond the package suite.

---

## Residual-risk

- **Behaviour change for a caller that passed a symlinked root and compared the answer.**
  The returned string differs now (`/private/var/...` where it used to say `/var/...`
  on macOS). This is not hypothetical — it is exactly what broke the two codex tests
  above. In-tree consumers are now reconciled; an EXTERNAL consumer that string-compares
  the value it passed against the value a tool reports would see a different, still-valid
  path. Callers should compare directories, not spellings.
- **`EvalSymlinks` adds a filesystem walk per accepted root.** Negligible against the
  `Stat` calls already present, unmeasured.
- **A path on a filesystem where `EvalSymlinks` fails** (an exotic mount, a permission
  boundary mid-path) is now rejected where it previously succeeded. This is the
  deliberate rejection-over-fallback choice, but it is a widening of the reject set.
