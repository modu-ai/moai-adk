# t26 — CodeRabbit false-green gate: does PR #1536 close it?

Read-only investigation. Repo `modu-ai/moai-adk`, local checkout `/Users/goos/MoAI/moai-adk-go`, branch `main`, HEAD `3b9b3bf99`, measured 2026-08-15.

---

## Verdict

**PARTIAL.**

- The **auto-merge path is CLOSED** by #1536, and this is production-verified — not inferred from reading the workflow. On PR #1539 the new gate ran end-to-end and withheld the merge on exactly the t26 failure shape.
- The **human / lead manual-merge path is STILL OPEN**, and it is not theoretical: PRs #1538 and #1539 were merged by hand *after* #1536 landed, both in the rate-limited state, minutes after the gate had withheld them. Nothing in #1536 touches the surface a human reads (`gh pr checks`), and that surface still prints `CodeRabbit  pass  0  Review rate limited`.

So the machine no longer merges an unreviewed PR. The human still does, and did, within the hour.

---

## Claim 1 — the false-green signature is a `success` commit status, not a missing one

The card said the check row read pass. Measured precisely: CodeRabbit reports through a **commit status** whose `state` is `success` even when no review ran; the fact that no review ran lives only in the `description` string.

### Evidence

```
$ gh pr checks 1527
...
CodeRabbit	pass	0		Review rate limited
```

```
$ gh pr checks 1526
...
CodeRabbit	pass	0		Review rate limited
```

```
$ gh api repos/modu-ai/moai-adk/commits/<head>/status --jq '.statuses[]|"\(.context) | \(.state) | \(.description)"'

### PR 1527  head=042130f0730dcceea528f507dc2b2d95bf88af25
CodeRabbit | success | Review rate limited | 2026-08-14T13:20:36Z

### PR 1526  head=d49072042fbcbedb4bf2b09bc591b6a09845826a
CodeRabbit | success | Review rate limited | 2026-08-14T12:55:01Z
```

Contrast, a PR that actually got reviewed:

```
#1537 head=a8f6d91
   status=success desc="Review completed"
```

`state` is `success` in both cases. `description` is the only field that separates them: `"Review rate limited"` vs `"Review completed"`.

### The comment bodies, verbatim

PR #1527 — one CodeRabbit comment (`id=5293752711`, created `2026-08-14T13:18:53Z`), opening with:

```
<!-- This is an auto-generated comment: summarize by coderabbit.ai -->
<!-- This is an auto-generated comment: rate limited by coderabbit.ai -->

> [!WARNING]
> ## Review limit reached
>
> `@GoosLab`, you've reached your PR review limit, so we couldn't start this review.
>
> **Next review available in:** **3 minutes**
```

PR #1526 — one CodeRabbit comment (`id=5292866555`, created `11:39:32Z`, **updated `12:54:40Z`**), carrying the same warning:

```
<!-- This is an auto-generated comment: rate limited by coderabbit.ai -->

> [!WARNING]
> ## Review limit reached
>
> `@GoosLab`, you've reached your PR review limit, so we couldn't start this review.
>
> **Next review available in:** **29 minutes**
```

Both merged in that state:

```
1527 MERGED 2026-08-14T13:28:54Z    (status success at 13:20:36 — merged 8 min later)
1526 MERGED 2026-08-14T13:04:21Z    (status success at 12:55:01 — merged 9 min later)
```

---

## Claim 2 — the two PRs differ in *which* guard would have caught them

Grep of the CodeRabbit comment bodies for the verdict markers #1536 parses:

```
######## PR 1527 ########
-- comment count: 1
-- risk markers (final_review_risk_start): 0
-- Merge Risk lines: (none)
-- rate limit marker: 2

######## PR 1526 ########
-- comment count: 1
-- risk markers (final_review_risk_start): 1
-- Merge Risk lines:
122:**Merge Risk:** _🟡 Moderate_ · up to `47cc7`
-- rate limit marker: 2
```

- **#1527** carries no verdict block at all → guard (c) branch `No CodeRabbit verdict comment found - withholding merge` (auto-merge.yml:191-194).
- **#1526** carries a verdict, but it covers `47cc7` while the head is `d490720…` → guard (c) freshness branch `Verdict covers … but head is … - stale, withholding merge` (auto-merge.yml:218-222). The rate-limit warning was written *into the same comment* (note `updated_at`), and CodeRabbit did **not** advance the `up to` prefix while rate-limited — which is what makes the freshness check work here.

Note what does *not* catch either case: **guard (b)**. It reads only `.state` (auto-merge.yml:164-181), which is `success` on both. The description string `"Review rate limited"` is never inspected anywhere in the workflow. All of the protection rests on guard (c) parsing a comment body.

---

## Claim 3 — the auto-merge path is closed, verified in production

PR #1539 (head `175a839`, status `success` / `"Review rate limited"`), auto-merge run `31864150325`:

```
Wait for CodeRabbit and read its Merge Risk verdict
  HEAD_SHA: 175a839ba6d77c686b00944d483a00f44933a8bb
  PR_NUMBER: 1539
  MAX_RISK: low
2026-08-15T04:20:11.109Z CodeRabbit status: success
2026-08-15T04:20:11.446Z ##[notice]No CodeRabbit verdict comment found - withholding merge

=== merge step ===
(empty — step did not run)
```

This is the exact t26 shape (`success` status, rate-limited, no review) reaching the new gate and being refused. The PR body for #1536 listed end-to-end execution as unverified; it is now verified.

---

## Claim 4 — the human path is open and was used

```
#1537 MERGED merged=2026-08-15T04:24:14Z by=GoosLab
#1539 MERGED merged=2026-08-15T04:24:34Z by=GoosLab
#1538 MERGED merged=2026-08-15T04:25:09Z by=GoosLab
```

```
#1538 head=39a74d1  status=success desc="Review rate limited"  → "> ## Review limit reached"
#1539 head=175a839  status=success desc="Review rate limited"  → "> ## Review limit reached"
```

`#1536` merged at `2026-08-15T04:16:06Z`. Eight minutes later, two PRs the new gate had refused were merged by hand, unreviewed. No auto-merge run merged anything after #1536 landed:

```
$ gh run list --workflow=auto-merge.yml --limit 12
04:35:13 skipped   04:25:32 skipped   04:24:42 skipped
04:19:54 success(withheld)  04:18:51 skipped  04:18:24 success ...
```

The workflow is `on: workflow_run` for CI, so `gh pr merge` typed by a human never enters it at all. And the surface the human reads — the `gh pr checks` row — is byte-identical between reviewed and unreviewed.

---

## Baseline-attribution

| What | Command | Observed in this run |
|---|---|---|
| #1536 scope | `gh pr view 1536 --json files,body` | one file, `.github/workflows/auto-merge.yml`, +161/−3, MERGED `2026-08-15T04:16:06Z` |
| Gate source | `Read .github/workflows/auto-merge.yml` at HEAD `3b9b3bf99` | 279 lines; guards at :59-84 (a), :142-181 (b), :183-248 (c); merge `if` at :251-254 |
| Check rows | `gh pr checks 1527` / `1526` | `CodeRabbit  pass  0  Review rate limited` |
| Commit statuses | `gh api repos/modu-ai/moai-adk/commits/<sha>/status` | `success` + description, for 1526/1527/1537/1538/1539 |
| Comment bodies | `gh api repos/modu-ai/moai-adk/issues/<n>/comments --paginate` | 1 CodeRabbit comment each; marker counts as quoted above |
| Gate in production | `gh run view 31864150325 --log` | `No CodeRabbit verdict comment found - withholding merge`, merge step empty |
| Merge actors | `gh pr view <n> --json mergedBy` | `GoosLab` for 1537/1538/1539 |

Repo is `modu-ai/moai-adk` (not `moai-adk-go` — the first `gh api` call 404'd on the wrong name and was corrected).

---

## Gaps — not observed

- No `gh api` call was made against branch-protection settings in this investigation; the required-contexts claim in #1536's body (CodeRabbit absent from required contexts, `strict: false`, `required_approving_review_count: 0`) is quoted from that PR body, **not re-measured here**.
- No search was run for a rule, skill, or doc that instructs the lead how to read a CodeRabbit row. The "nothing changed on the human surface" claim rests on #1536 touching exactly one file (measured) — not on a sweep of `.claude/rules/**` for such guidance. If guidance exists elsewhere and predates #1536, it did not prevent the 04:24 merges.
- The rate-limit strings were sampled from four PRs (#1526, #1527, #1538, #1539). Whether CodeRabbit emits other descriptions for other non-review outcomes (config error, timeout, cancelled) is unmeasured — the ladder of `description` values is not enumerated.
- `guard (b)`'s 15-minute timeout was not exercised; every measured rate-limited PR reached `success` immediately, so the wait loop never waited.

## Residual-risk

1. **The whole auto-merge protection is one comment parse.** Guard (b) accepts `success`; only guard (c)'s body parse distinguishes reviewed from not. If CodeRabbit changes its comment format — drops the `final_review_risk_start` markers, renames `Merge Risk:`, or changes the `` up to `sha` `` suffix — guard (c) fails *closed* (good) but the gate becomes a permanent withhold, which pressures exactly the manual-merge bypass already in use.
2. **The `up to` prefix staying stale while rate-limited is N=1** (#1526). If CodeRabbit ever refreshes that prefix on a re-review it could not complete, the freshness check would pass a stale verdict. Nothing observed says it does; nothing observed rules it out.
3. **The manual bypass is now the normal path.** Three of the last three merges were by hand. A gate the operator routes around on every PR is not a gate.

---

## The closing predicate

The gap is on the human/lead read path, and the cheapest closing predicate is one field earlier than the one #1536 parses:

```bash
gh api "repos/$REPO/commits/$HEAD_SHA/status" \
  --jq '[.statuses[] | select(.context == "CodeRabbit")] | last | .description'
```

**Refuse to treat CodeRabbit as passed unless this equals `Review completed`.** `Review rate limited` — the string measured on #1526, #1527, #1538, #1539 — means no review happened, regardless of `state == "success"`.

Where it needs to be checked:

- **Lead-side, before a card advances out of `review` or `sync`**, per `.claude/rules/moai/workflow/kanban-dispatch.md` § "Completion is read, never trusted". The lead currently has no instruction that `gh pr checks` lies here; it needs one, phrased as: a CodeRabbit row is evidence only when the commit status description reads `Review completed` **and** a `Merge Risk:` line exists whose `` up to `prefix` `` matches the current `headRefOid`. Anything else is a gap, not a pass.
- **Optionally in the workflow too**, as a cheap early guard (b′) before the 15-minute wait — `description != "Review completed"` → withhold immediately, with a clearer notice than "no verdict comment found". This is redundant with guard (c) today but survives a CodeRabbit comment-format change, which guard (c) does not.

Branch protection is *not* the lever: adding `CodeRabbit` to required contexts would not help, because the status state is `success` in precisely the failing case.
