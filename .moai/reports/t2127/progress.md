# Card t2127 progress record (class B, no SPEC, cycle_type=tdd)

Defect (Codex P2 review, PR #1969, internal/hook/branch_guard.go:621 @04d8c27a7e): `cdPrimaryTarget`
runs `anyCdTargetPattern` over `substituteQuotedArguments(command)` only, so a `cd /abs` inside a shell
comment (e.g. `git commit -m x # cd /absolute/primary` from a linked worktree) is returned as the
command's target and the protected-commit judgment (caller ~line 781) denies a legitimate worktree commit.

## Cause-establishing evidence — RED

Tree SHA: 04d8c27a7e (HEAD == local main; test file untracked at measurement time).
Command: `go test ./internal/hook/ -run 'TestCdPrimaryTargetIgnoresCdInShellComment' -count=1 -v`
Exit code: 1

Verbatim result (paths shortened here only by the temp dir; full text in red.txt, untracked scratch):

```
--- FAIL: TestCdPrimaryTargetIgnoresCdInShellComment (4.46s)
    --- FAIL: .../trailing_comment            cdPrimaryTarget("git commit -m x # cd <primary>") = ("<primary>", true), want ok=false
    --- FAIL: .../comment_after_quoted_message cdPrimaryTarget("git commit -m \"it's\" # cd <primary>") = ("<primary>", true), want ok=false
    --- FAIL: .../whole-line_comment_on_a_later_line cdPrimaryTarget("git commit -m x\n# cd <primary>") = ("<primary>", true), want ok=false
    --- PASS: cd_before_commit, cd_after_commit, real_cd_then_comment, hash_inside_quotes_then_cd,
              cd_inside_quotes_is_data, mid-word_hash_then_cd, cd_to_worktree
FAIL	github.com/modu-ai/moai-adk/internal/hook	5.385s
exit=1
```

Cause: the cd scan reads the quote-collapsed text without eliding comments; `normalizeCommandForScan`
elides them via `substituteShellComments`, `cdPrimaryTarget` did not.

## Fix

`cdPrimaryTarget` now scans `substituteShellComments(substituteQuotedArguments(command))`, the order
`normalizeCommandForScan` uses; function comment updated. Only that function changed.

## GREEN evidence (tree: 04d8c27a7e + this card's working changes)

- `go test ./internal/hook/ -run TestCdPrimaryTargetIgnoresCdInShellComment -count=1 -v`: 10 subtests PASS
  (3 comment-only cases now ok=false; 7 controls unchanged), `ok ... 5.859s`.
- `go test ./internal/hook/ -run 'BranchGuard|Branch|CdPrimary|ProtectedCommit|Comment' -count=1 -v -timeout 20m`:
  438 RUN, 438 PASS, 0 FAIL, 0 SKIP; `ok ... 114.672s`.
- `go test ./internal/hook/ -count=1 -v -timeout 20m` (full package): exit 0, `ok ... 1026.145s`;
  4242 RUN, 4236 PASS (`^\s*--- PASS:`), 0 FAIL (`^\s*--- FAIL:`), 7 SKIP (pre-existing).
  A naive `grep -c -- '--- FAIL'` returns 6: those are fixture log lines
  (`error="Exit code 1\n--- FAIL: TestX"` in TestDetectorNeverAltersToolCall), not results.
- `go vet ./internal/hook/`: no output, exit 0. `gofmt -l internal/hook`: no output.
- `GOOS=windows GOARCH=amd64 go build ./...`: exit=0.

Gate (`/moai gate` via verify) on this tree:
- Snapshot reuse: `moai verify check --key-current` exit 1, "no snapshot recorded for the current working-tree
  key" (key 04d8c27a7ed77afc09217f8623519806d48df3c4:3a6e9af39f2147e1); nothing reused, checks run directly above.
- Lint: `golangci-lint run --timeout 5m ./internal/hook/...` was REFUSED twice, exit 3, "parallel golangci-lint
  is running" (another session holds the lint lock; not killed, per the no-pattern-kill rule). golangci-lint is
  therefore NOT observed on this tree; the substitute is `go vet` + `gofmt -l` (both clean). CI is the lint owner.
- Test: the gate's full-suite test is replaced by the scoped + full `internal/hook` runs above (full `go test ./...`
  is barred locally by the lane protocol); recorded here, not hidden.

Gaps: golangci-lint not observed (refused, see Gate above); repository-wide test verdict is owned by CI on the integration branch and is PENDING; only
internal/hook was swept. Raw run logs (red.txt, green-*.txt) are untracked local scratch.
