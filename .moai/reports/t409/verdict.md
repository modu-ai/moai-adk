# t409 Verdict — always-loaded budget diet (shared-flag-slot constraint)

Card: t409 (Tier S · Class B, no SPEC per dispatch). Worktree `.claude/worktrees/t409`,
branch `WT-budget-diet`, based on develop head `e8ae9798a`.

## Claim

The fifth availability constraint (`25b341dcb`, card t400) is compressed to its stub
essentials with the diagnostic one-liner relocated to the paths-scoped detail companion;
`TestAlwaysLoadedTokenBudget` goes green with headroom +81 and no content deleted — every
fact remains reachable through the `cross-session-messaging-detail.md` § The shared flag
slot pointer.

## Evidence

RED (pre-edit, this run):

```
$ go test ./internal/config/ -run 'TestAlwaysLoadedTokenBudget$' -count=1 -v
    token_budget_guard_test.go:69: always-loaded surface = 76292 tokens (budget 76210, headroom -82, 17 entries)
--- FAIL: TestAlwaysLoadedTokenBudget (0.00s)
```
exit 1, measured on tree `e8ae9798a` (develop head, absorbed clean).

GREEN (post-edit, this run):

```
$ make build            # exit 0
$ go test ./internal/config/ -run 'TestAlwaysLoadedTokenBudget$' -count=1 -v
    token_budget_guard_test.go:69: always-loaded surface = 76129 tokens (budget 76210, headroom 81, 17 entries)
    --- PASS: TestAlwaysLoadedTokenBudget (0.00s)
```
exit 0. Recovery: −163 tokens (652 bytes off the always-loaded stub; the one-liner's
~300 bytes moved into the paths-scoped detail, which the surface does not count).

Scope checks this run: `go test ./internal/template/ -count=1` → ok 24.1s (mirror
parity, catalog, neutrality, commands audit); `go vet ./internal/config/
./internal/template/` → exit 0. Changed files: the rule pair (local + template mirror)
and the detail pair (local + template mirror) — four files, no generated artifacts.

## Baseline-attribution

Both numbers above were produced in this session, in this worktree, by the commands
quoted; RED pinned to tree `e8ae9798a`, GREEN measured on the post-edit working tree
before commit. The guard is `len(bytes)/4` over CLAUDE.md + AGENTS.md(+mirror) + no-paths
rules + moai.md per `internal/config/token_budget_guard.go`.

## Gaps

- The full suite was not run locally (scoped to touched packages; CI is the full-suite
  judge).
- No CI verdict yet — nothing pushed at the time of writing; the develop quiet-head
  observation is post-integration.
- t196 had NOT landed on origin/develop when this was measured (head still `e8ae9798a`).

## Residual-risk

- **t196** is expected to add +136 tokens to both AGENTS.md copies ("양 사본"). Headroom
  +81 absorbs neither +136 nor +272 — if t196 lands after this card, the guard trips
  again on the then-head, visibly, in t196's own verification or CI. This card's
  obligation (green on the tree it measured) is met; the next diet belongs to whoever
  lands next.
- The stub bullet still carries the slot identity and the first-party/inheritance facts;
  the api.anthropic.com host mechanism and the measurement record live only in the
  detail companion — reachable, but one pointer hop away from the always-loaded surface.
