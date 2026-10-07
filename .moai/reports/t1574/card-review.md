# card-review — t1574 (SPEC-ZONE-SHELL-PARSING-001)

- Date: 2026-10-08
- Reviewer engine: codex exec (codex-cli 0.160.1, model gpt-6.1-sol, reasoning effort high, sandbox read-only), session id `01a1172a-8f40-72c3-8781-b4846f539c17`, workdir /tmp with the card diff inlined (`git diff b9ef00380..HEAD -- internal/hook/`, 1124 lines)
- Review target: the card's 8 commits vs base `b9ef00380` (HEAD `2ba2ae36f` at review time) — `internal/hook/protected_zone_shell.go` + the two test files
- Worker: dev-t1574 (advisory record only — FIX NOTHING; the lane disposes findings)
- Tree state after the review probes: clean (probe file deleted; `git status --short internal/hook/` empty)

## VERDICT: FAIL

One P1 finding (a soundness hole), empirically reproduced by the lane worker. Provenance: PRE-EXISTING at base — the card neither introduced nor worsened it, and the card's five reproduced classes remain closed; this is an adjacent, still-open instance of the K2 class (see Provenance below).

## Findings

### [P1] internal/hook/protected_zone_shell.go:468 — conditional calls discard the absent-function world (transitive declaration leak)

- Reviewer (verbatim, translated form): For `false && f() { rm() { :; }; }; f; rm zone_dir/x`, walking the conditional `f` installs a certain no-op `rm`, which remains installed during the external fallback and subsequent statements. Real Bash never declared `f`, so the final external `rm` deletes the protected file while the guard allows it.
- Empirical reproduction (lane worker probe, this tree, in-package probe test deleted after observation): `false && f() { rm() { :; }; }; f; rm zone_dir/secret.md` → `decision="allow" reason=""`. Real bash: `f` is undefined (the branch never ran), the `f` call fails with command-not-found, `rm` is never shadowed, and the protected file is deleted for real.
- Mechanism: the K2 dual interpretation walks a conditionally-declared name's bodies as one world and then falls through for THAT name — but the bodies' own FuncDecls install entries into the walker's shared registry as CERTAIN. Subsequent statements (and the absent-world fall-through of other calls) then see `rm` as a certain shadow, though the declaration only exists in the branch world. The dual-interpretation boundary must extend transitively: declarations a CONDITIONAL call's body installs (absent before the call) belong to the may-not-have-run world.
- Provenance (base tree): PRE-EXISTING. At base the branch join registered `f` CERTAIN (no conditional dimension), so the same shape took the identical walker path (certain `f` → body walk → `rm` installed → shadow allow). Probe: the straight-line chain `f() { rm() { :; }; }; f; rm zone_dir/secret.md` — the exact base code path for this shape — answers `decision="allow"`. The card closed the DIRECT K2 instance (`false && rm() { :; }; rm zone_dir/secret.md` → now deny, frozen subtest green); this transitive instance was not among the five reproduced REDs and remains open.
- Suggested disposition direction (lane's call, not applied): when `zoneWalkDeclared` walks the bodies of a CONDITIONAL name, union the post-call registry with the pre-call entry through `mergeZoneWorlds` — the call itself may not have executed in real bash (command-not-found), so everything its body installed keeps the possibility of absence. Every such shape still denies fail-closed today through the OTHER interpretations; this widens soundness, not a verdict flip on covered shapes already asserted.

## Scope notes (what the review attachment could not establish)

The reviewer judged from the inlined diff only and noted three items it could not verify from that attachment. They ARE established by the run-phase evidence and are cross-referenced here for the record:

1. Historical RED results — acceptance.md §B EV-1/EV-2/EV-3/EV-4 (tree b9ef00380) + progress.md §E.2 E8 (K4 RED verbatim, this run).
2. Unchanged-family execution — `TestProtectedZone$` ok (4.999s / 2.738s / 2.564s across the rounds, progress.md §E.2 AC-ZSP-005); frozen assertion lines byte-identical (§E.2 D4 section).
3. vet/lint/suite success — vet exit 0, lint `0 issues.`, suite `ok 406.964s coverage: 87.5%` (§E.2 AC-ZSP-006 + gate round).

Reviewer NOTES also acknowledged: the frozen recursion subtest asserts termination only (per D4 — the DENY verdict is owned by the matrix's unbounded cells, which assert `loop-unbounded`); the matrix's duplicate seed supplies that denial coverage.

## Review-run provenance (why three attempts)

1. Run 1 (worktree, read-only sandbox): the codex agent attempted `moai verify codex-review`; the sandbox refused the receipt temp write (`snapshot tmp create … operation not permitted`) — no code verdict produced.
2. Run 2 (worktree, workspace-write, prompt corrected): the agent entangled with the wired Stop-hook/gate semantics and refused to review ("confirming automation refusal"), maintaining run 1's non-verdict — no code verdict produced.
3. Run 3 (/tmp, read-only, `--skip-git-repo-check`, diff inlined, gate tooling explicitly excluded): produced the verdict above. This run's output is the review of record.

## Receipt

- codex session id: `01a1172a-8f40-72c3-8781-b4846f539c17`
- Raw streams: /tmp/t1574-card-review-raw3.txt (review of record), -raw2/-raw1 (failed attempts), last message /tmp/t1574-card-review-last3.md
- The formal rcpt id follows the lane ceremony (the leader issues it; this record cites the codex session id as the review's mechanical provenance).
