# t160 — statusline forge counts: totals, render, budget guard

Class B, run → sync in-lane.

- branch: `WT-forge-counts`
- base: `origin/release/v3.1.1` @ `f5598871f` (checked before editing — `EnterWorktree` opened on `4100d8767`, moved with `git merge --ff-only`)
- commit: `eb95ff228`
- pushed: **no** (per the dispatch)

## 1. Totals instead of enumeration (GitHub)

`forgeGitHub` now calls `gh repo view --json issues,pullRequests
--jq '.issues.totalCount, .pullRequests.totalCount'` — one call answering
**both** counts. `issueArgs`/`prArgs` remain as the shape a forge without a
totals endpoint uses; GitHub no longer reaches them.

Measured on `cli/cli`, 2026-08-20:

| | call | GraphQL cost | reported |
|---|---|---|---|
| totals | `gh repo view --json issues,pullRequests` | **1** | 1012 open issues, 76 open PRs |
| enumeration | `gh issue list --state open --limit 1000` | **10** | **1000** — the cap, not the count |

So the enumeration was both dearer and, past `--limit`, wrong. **This settles
the item the card flagged as unverified**: pagination does cost extra (one
point per page of a hundred), and the truncation risk is not theoretical — it
fires at 1000, and `cli/cli` is already past it.

## 2. GitLab — asymmetry, deliberate and written down

`glab` exposes no count-only listing: `issue list` / `mr list` return items,
and the total lives in the REST `X-Total` pagination header that `glab` does
not surface through its own JSON output. Reaching it would mean calling the
API around `glab`, i.e. a second authentication path for a status bar to get
right. GitLab therefore keeps the page-bounded enumeration, and the asymmetry
is documented on `forgeGitLab` in `forge.go`.

**Unverified, and marked as such in the comment**: `glab` is not installed on
this machine (measured absent from PATH), so that paragraph rests on glab's
documented output rather than on a run.

## 3. Render restored, in a shape that cannot repeat the misread

`renderForgeCounts` appends the counts to the repo part:

```
📡 modu-ai/moai-adk, 0/0 🐛12 🔀3 | 🅱️ WT-forge-counts +5 | 💾 +0 M5 ?0
```

The 2026-08-18 removal was right about its cause — two slash-joined pairs side
by side is what let "59/0" read as issues over PRs — so the restored form
separates them **by shape, not by position**: the a/b pair keeps its slash,
each count carries its own glyph and no slash joins them. A test pins this
mechanically (exactly one `\d+/\d+` on the line). A zero count is omitted,
matching the dirty count in the same segment; `SegmentGitHub` gates the whole
suffix.

## 4. Budget guard

Before fetching, `forgeBudgetAllows` asks the forge what budget remains and
skips the refresh below `forgeRateFloor` (100), keeping the stale count. The
question is free — `gh api rate_limit` consumes neither bucket (measured:
three consecutive reads, zero delta in core and graphql alike).

Fail-open in both directions that matter: a forge with no rate query, and a
query that errors or answers unparseably, both proceed. An unanswerable
question about the budget is not evidence the budget is gone, and treating it
as such would silence the segment on every forge that cannot be asked.

## Evidence

| Claim | Command | Observed |
|---|---|---|
| totals == enumeration below the cap (repo 1) | `gh repo view --json issues,pullRequests` vs `gh issue list`/`pr list` on `modu-ai/moai-adk` | `0/1` both ways (closed issues exist, so the OPEN filter is real) |
| totals == enumeration below the cap (repo 2) | same on `cli/cli`, PR side | totals `76`, enumeration `76` |
| enumeration truncates above the cap | `gh issue list --repo cli/cli --limit 1000` vs totals | `1000` vs `1012` |
| cost per call | rate-limit delta around each call | totals `1` point; enumeration `10` points |
| rate query is free | three consecutive `gh api rate_limit` reads | core and graphql unchanged |
| end-to-end refresh costs 1 point | `bin/moai statusline --refresh-github` with rate-limit reads either side, twice | `4870→4869`, `4868→4867` |
| refresh writes the right counts | the cache after that run | `{"open_issues":0,"open_prs":1}` — matches the enumeration |
| render is visually distinct | `bin/moai statusline` with a seeded cache | `📡 modu-ai/moai-adk, 0/0 🐛12 🔀3 \| 🅱️ …` |
| zero omitted | same, real cache (`0` issues, `1` PR) | `📡 modu-ai/moai-adk, 0/0 🔀1 \| 🅱️ …` |
| package green | `go test ./internal/statusline/ -count=1` | `ok … 13.563s` |
| CLI entry point unaffected | `go test ./internal/cli/ -run 'Statusline\|StatusLine' -count=1` | `ok … 1.506s` |
| static analysis | `go vet` for host, `GOOS=windows`, `GOOS=linux` | clean on all three |

Measurement note worth recording: two earlier end-to-end deltas read 34 and 28
points. They were **not** this code — other sessions on this machine were
using `gh` concurrently. Control intervals of similar length with no forge
call showed zero drift, and two later back-to-back measurements both showed
exactly 1. The attributable figure is 1; the 34/28 readings are contention and
are reported here rather than dropped.

## Tests

| Test | Pins |
|---|---|
| `TestRefreshGitHubCounts_FetchesValuesViaGitHubStub` (rewritten) | GitHub makes the totals call and does **not** enumerate |
| `TestRefreshGitHubCounts_SkipsFetchNearTheRateFloor` | below the floor, no counts call, stale values kept |
| `TestRefreshGitHubCounts_UnreadableBudgetStillFetches` | an unparseable budget answer does not block the fetch |
| `TestRefreshGitHubCounts_ShortTotalsAnswerKeepsStaleValues` | a one-number answer is a failed fetch, not a zero |
| `TestRender_ForgeCountsRenderDistinctlyFromAheadBehind` (replaces the removal test) | glyph-tagged counts, exactly one number/number pair on the line |
| `TestRender_ForgeCountsOmitZeros` | zero omitted; unavailable cache renders nothing |
| `TestRefreshGitHubCounts_FetchesValuesViaGitLabStub` (unchanged) | GitLab still enumerates `issue`/`mr` |

`TestRender_GitHubCountsPrefixRemoved` was **deleted**: it pinned the absence
this card was dispatched to reverse. Its sibling
`TestRenderSessionLine_GitHubSegmentDisablable` survives — the gate still
hides the counts — with its comment corrected.

## Gaps (not observed)

- GitLab's totals claim is documentation-based; `glab` is not installed here.
- Windows/linux: `go vet` only, no run. No new syscall or path surface.
- Behaviour when `gh` is authenticated but the repo is inaccessible (private,
  revoked token) is unchanged and untested here — it takes the existing
  failed-fetch path and keeps the stale cache.
- Full repository suite not run locally, per the standing rule.

## Residual risk

- `gh repo view`'s `issues`/`pullRequests` totalCount being OPEN-scoped is a
  measured property (a repo with closed issues reported 0), not a documented
  guarantee. If gh ever changed that scoping the counts would silently
  include closed items; the GitLab path is unaffected.
- The two-line parse depends on the jq filter's output order (issues first).
  It is fixed in `countArgs` beside the parser, and a short answer is refused,
  but a reordered filter would silently swap the two numbers.
- `forgeRateFloor` at 100 is a judgement, not a measurement: a refresh costs
  one point against a 5000/hour ceiling, so the floor exists to leave the
  operator's own tooling room, not to protect the statusline.
