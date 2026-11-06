---
card: t2137
reviewed_sha: 7a5ce7e9f745e582b6634fb73e5199c1ae149f0a
reviewer: lane-28 self-review (Claude)
codex: one operator-approved codex_audit (d-20261106T164223Z-2fc5), receipt cited separately in codex-receipt.md
---

# t2137 card review

## Diff reviewed (04d8c27a7e..7a5ce7e9f7; code commit 629bf3ea8b)

- `internal/agenticfactory/core/sanitize_test.go:95`: adds `// home-guard:allow deliberate input: the sanitizer must leave a tilde path alone` on the line above the `"~/work/notes.md"` table entry. The tilde path is the case under test, so marking it fits better than rewording.
- `internal/cli/factory_fixture_home_test.go:36`: the `t.Errorf` text now says "the real user home's .moai" instead of `~/.moai`. Test behaviour is unchanged.
- Nothing under `internal/testguard/` changes, including the seed file and its 24 stale entries.

## Lane observations

- Before the fix, on base 04d8c27a7e (lane run): the guard exited 1 and reported 2 NEW home accesses, at the two lines above.
- After the fix (lane run): `scanned 3927 test files: 160 violations, seed 184 entries, 0 new, 24 stale seed entries`, then `ok … internal/testguard`.
- The auditor's lint on changed lines reported `0 issues.`. The full `internal/testguard` package passes (auditor run).

## Debt (non-blocking)

- F2: the marker line breaks the comment alignment of the table entry below it. This is cosmetic.
- F3: the comments at `factory_fixture_home_test.go:14` and `:17` still mention `~/.moai`. The scanner ignores comments by design.
