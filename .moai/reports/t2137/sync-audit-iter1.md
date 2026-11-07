auditor-model: claude-opus-5-5

---
verdict: FAIL
audited_sha: 76b6dad2666ca62989beec395b3b6ca3f6e72a9b
---

# Sync audit — card t2137 (Class A, main red: home-guard) — iteration 1

- Card diff: `04d8c27a7ed77afc09217f8623519806d48df3c4..76b6dad2666ca62989beec395b3b6ca3f6e72a9b` (fix 629bf3ea8b, progress record 76b6dad266)
- Worktree: `/Users/goos/.moai/worktrees/projects/moai-adk-go/t2137` (branch `WT-internal-testguard-testh-2`)
- Mode: Claude-only, as instructed by the dispatch (Codex not approved for this card). Flat weighted scoring, built-in default profile.
- Overall score (substance): 0.93 — Functionality 95 / Security 95 / Craft 90 / Consistency 92
- **Overall verdict: FAIL — process gate only.** The code fix has no defect. The verdict is FAIL because this tree's audit plan marks the codex backend `required` (`enforced_required: ["codex"]`), the dispatch forbids calling codex, and the receipt guard refuses a PASS that cites no codex receipt. Per the dispatch ("if any receipt/verdict guard insists on a codex receipt and blocks you, record verdict FAIL with that reason"), the verdict is recorded as FAIL. No code change is required; resolving the gate (operator approves a codex review, or the leader/operator accepts the card without one) is the only open item.

## Dimension scores

| Dimension | Score | Verdict | Evidence |
|---|---|---|---|
| Functionality (40%) | 95/100 | PASS | `go test -count=1 -v -run TestHomeGuard_RepositoryHasNoNewHomeAccess ./internal/testguard/` → `scanned 3927 test files: 160 violations, seed 184 entries, 0 new, 24 stale seed entries ...` / `--- PASS` / `exit=0`; both touched tests PASS (below) |
| Security (25%) | 95/100 | PASS | Diff is test-only (2 test files + 1 report). Secret probe `git diff ... \| grep -niE 'password\|secret\|token\|api[_-]?key'` → one hit, line 44 `// digits only: a number, not a mixed token` (context line, word "token", not a secret). The allow marker exempts exactly one literal (scan.go:96 checks `markerLines[line] \|\| markerLines[line-1]`), so the guard is not widened |
| Craft (20%) | 90/100 | PASS | `golangci-lint run --new-from-rev=04d8c27a7e ./internal/agenticfactory/core/ ./internal/cli/` → `0 issues.` exit 0; `gofmt -l` on both files → no output, exit 0; `go test -count=1 ./internal/testguard/` → `ok ... 6.889s` |
| Consistency (15%) | 92/100 | PASS | Marker format matches `allowComment` regex `//\s*home-guard:allow\s+\S` (scan.go:30) with a real reason; commits carry `[t2137]` and card id; seed untouched (`git diff --quiet 04d8c27a7e HEAD -- internal/testguard/` exit 0) |

Must-pass firewall (Functionality, Security): both PASS on substance.

## Findings (structured defect-list)

- **F1** [High, process] [blocking] [outside the diff — tree configuration] `.moai/config/sections/workflow.yaml:38` (`codex: required`) — the audit plan (`moai verify audit-plan`) reports `cross_model_required: true`, `enforced_required: ["codex"]`. The dispatch forbids every cross-model tool, so the required codex gate is unmet and no `audit_receipt` exists. Required fix: none in code. The leader takes this to the operator: either approve a codex review for this card (then re-run the audit and cite the receipt), or accept the card on the Claude-only verdict. Confidence: high.
- **F2** [Low] [optional] [in-diff] `internal/agenticfactory/core/sanitize_test.go:95-96` — the marker sits on its own line inside a composite-literal list, which separates line 96's trailing comment from the gofmt alignment column of its neighbours (lines 97-109). gofmt accepts it (`gofmt -l` empty). Purely cosmetic; no action needed. Confidence: high.
- **F3** [Info] [optional] [in-diff] `internal/cli/factory_fixture_home_test.go:36` — the reworded message ("the real user home's .moai") carries the same meaning; it is a `t.Errorf` string only, so test behaviour is unchanged (test PASS). The file's comments at lines 14 and 17 still contain `~/.moai`; they are comments, which the AST scanner ignores by design (scan.go:68-71), so they are not violations. No action. Confidence: high.

No repair-born findings (first audit). Not a verdict-function card: the guard code (`internal/testguard/`) is unchanged.

## Check 1 — the fix

Claim: the allow marker is in an accepted placement and format with a real reason; the tilde use is deliberate; the reword keeps behaviour; only the two tests and the progress record changed; seed untouched.

Evidence:

```
$ git diff --stat 04d8c27a7e HEAD
 .moai/reports/t2137/progress.md               | 30 +++++++++++++++++++++++++++
 internal/agenticfactory/core/sanitize_test.go |  1 +
 internal/cli/factory_fixture_home_test.go     |  2 +-
 3 files changed, 32 insertions(+), 1 deletion(-)
```

```
+		// home-guard:allow deliberate input: the sanitizer must leave a tilde path alone
 		"~/work/notes.md", // a tilde path names no user
```

```
-		t.Errorf("MoaiHome() = %q resolves outside the test temp dir %q: factory tests would write project folders into the real ~/.moai", home, tmp)
+		t.Errorf("MoaiHome() = %q resolves outside the test temp dir %q: factory tests would write project folders into the real user home's .moai", home, tmp)
```

scan.go (read at audit HEAD):

```
30	var allowComment = regexp.MustCompile(`//\s*home-guard:allow\s+\S`)
84				if allowComment.MatchString(c.Text) {
85					markerLines[fset.Position(c.Slash).Line] = true
96			if seen[key] || markerLines[line] || markerLines[line-1] {
```

The marker is a real comment token on line 95, directly above the literal on line 96 — the "line directly above" placement scan.go accepts. Its reason text (`deliberate input: the sanitizer must leave a tilde path alone`) is non-empty. The use is deliberate: the literal is one entry of the `TestSanitizeLeavesOrdinaryTextAlone` table of ordinary inputs the sanitizer must leave untouched, and its existing trailing comment says so (`// a tilde path names no user`); rewording it would remove the case under test.

```
$ git diff --quiet 04d8c27a7e HEAD -- internal/testguard/; echo "testguard_diff_exit=$?"
testguard_diff_exit=0
$ git log --oneline 04d8c27a7e..HEAD
76b6dad266 [t2137] docs: progress record for the home-guard fix (card t2137)
629bf3ea8b [t2137] test: clear two new home-guard violations — allow the deliberate tilde input, reword the ~/.moai message (card t2137)
```

## Check 2 — tests and lint (run by this auditor, one at a time, `GOCACHE=.../scratchpad/gocache-lane28`)

```
$ go test -count=1 -v -run TestHomeGuard_RepositoryHasNoNewHomeAccess ./internal/testguard/; echo "exit=$?"
=== RUN   TestHomeGuard_RepositoryHasNoNewHomeAccess
    repo_guard_test.go:96: scanned 3927 test files: 160 violations, seed 184 entries, 0 new, 24 stale seed entries (stale entries do not fail; regenerate with TESTGUARD_WRITE_SEED=1)
--- PASS: TestHomeGuard_RepositoryHasNoNewHomeAccess (11.06s)
PASS
ok  	github.com/modu-ai/moai-adk/internal/testguard	11.166s
exit=0
```

```
$ go test -count=1 ./internal/testguard/; echo "exit=$?"
ok  	github.com/modu-ai/moai-adk/internal/testguard	6.889s
exit=0
```

```
$ go test -count=1 -v -run TestSanitizeLeavesOrdinaryTextAlone ./internal/agenticfactory/core/
=== RUN   TestSanitizeLeavesOrdinaryTextAlone
--- PASS: TestSanitizeLeavesOrdinaryTextAlone (0.00s)
PASS
ok  	github.com/modu-ai/moai-adk/internal/agenticfactory/core	0.103s
exit=0
```

```
$ go test -count=1 -v -run TestFcFixtureKeepsMoaiHomeInsideTheTestSandbox ./internal/cli/
=== RUN   TestFcFixtureKeepsMoaiHomeInsideTheTestSandbox
--- PASS: TestFcFixtureKeepsMoaiHomeInsideTheTestSandbox (1.00s)
PASS
ok  	github.com/modu-ai/moai-adk/internal/cli	3.118s
exit=0
```

```
$ golangci-lint version
golangci-lint has version 2.15.0 built with go1.27.2 from 25c5369 on 2026-10-16T10:51:54Z
$ golangci-lint run --new-from-rev=04d8c27a7e ./internal/agenticfactory/core/ ./internal/cli/
0 issues.
exit=0
```

```
$ gofmt -l internal/agenticfactory/core/sanitize_test.go internal/cli/factory_fixture_home_test.go; echo "gofmt_exit=$?"
gofmt_exit=0
```

## Audit plan (read before scoring)

```
$ moai verify audit-plan --project-root /Users/goos/.moai/worktrees/projects/moai-adk-go/t2137
  "config_status": "ok",
  "model": "codex",
  "backends": [ claude gate "off", codex gate "required", glm gate "off" ]  (all source config.gates, explicit)
  "cross_model_active": true,
  "cross_model_required": true,
  "enforced_required": [ "codex" ]
exit=0
```

`audit_multi` / `codex_audit` were NOT called, per the dispatch. No `audit_receipt` was issued.

## Five-section summary

**Claim.** The fix clears both NEW home-guard violations with no behaviour change: the deliberate tilde input carries a valid `home-guard:allow` marker with a reason, the `~/.moai` message is reworded, the guard and its seed are untouched, and the touched packages lint clean on the changed lines.

**Evidence.** The verbatim outputs above: guard test `0 new` exit 0; full testguard package ok; both touched tests PASS; `golangci-lint --new-from-rev` `0 issues.`; `gofmt -l` empty; diff stat of 3 files; testguard diff empty.

**Baseline-attribution.** All commands ran in this session against the worktree at HEAD `76b6dad2666ca62989beec395b3b6ca3f6e72a9b`, base `04d8c27a7ed77afc09217f8623519806d48df3c4`. Go tests were compiled from this tree by `go test`, so the judging build is the tree itself. `moai verify audit-plan` used the installed binary `/Users/goos/go/bin/moai` (its commit was not compared with HEAD — see Gaps); the moai MCP server reports build commit `04d8c27a7e`, which is the card base.

**Gaps.**
- The required codex gate was not exercised (dispatch prohibition) — F1, blocking.
- The pre-fix failure (2 new violations on 04d8c27a7e) was not re-run by this auditor; it rests on the dispatch and the lane's progress record.
- `golangci-lint` was scoped to new lines (`--new-from-rev`); pre-existing findings in the two packages were not measured.
- Only the two named tests were run in `internal/agenticfactory/core` and `internal/cli`, not the full packages. Coverage was not measured (test-only diff, no production code changed).
- The installed `moai` binary's commit was not compared with HEAD (§2.2); the audit-plan reading is config-driven and matches `workflow.yaml:38` read directly.

**Residual-risk.** A merge of newer `main` into this branch may bring other new home-guard violations; the guard must be re-run after integration. The 24 stale seed entries remain by design (not failing).

## Iteration history

- iter1 (this file): substance clean (0.93); FAIL on the unmet required codex gate only.
