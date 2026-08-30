# M2 — Snapshot integrity, scope derivation, opening digest (card t332)

## §1 Snapshot provenance

| Field | Value |
|---|---|
| Deciding snapshot | `.moai/reports/t332/queue-snapshot-run.tsv` |
| Capture command | `moai todo` (bare, no filter, no `head`/`tail`/`grep`) |
| Capture time (UTC) | **2026-08-30T11:16:45Z** |
| Captured from | primary checkout `/Users/goos/MoAI/moai-adk-go` (the queue is primary-checkout-only) |
| Tree HEAD at capture | `6165f9f5e` (worktree t332, `WT-backlog-hygiene`, post-`develop` absorption) |
| Rows | 101 |
| Superseded snapshot | `.moai/reports/t332/queue-snapshot.tsv` (plan-phase, captured 2026-08-29 20:36, 100 rows) — retained for the delta comparison below, **not** the source of any card entry |

**Why a second snapshot.** The plan-phase snapshot is 24 hours old and the lead moved seven cards
between plan and run. Reading the sweep from it would have classified seven live lanes' cards as
dead cards — the exact failure this SPEC exists to catch. REQ-BH-003's event branch fires; the
in-scope set below is re-derived from the run-phase snapshot alone.

## §2 State counts (REQ-BH-001, REQ-BH-002)

`cut -f2 queue-snapshot-run.tsv | sort | uniq -c`:

```
  62 queued
  17 picked
  18 dropped
   4 (relation rows: "↳ contains …" / "↳ absorbs …")
```

62 + 17 + 18 = **97 cards**, + 4 relation rows = 101 snapshot rows. The partition is exact.

Cross-check against the live store, an independent count:
`sqlite3 backlog.db "select count(*) from items;"` → **97**. The two agree.

## §3 Delta against spec.md §B.1 (REQ-BH-003 — the event branch, fired)

| | plan-phase (§B.1) | run-phase | Δ |
|---|---|---|---|
| queued | 68 | **62** | −6 |
| picked | 10 | **17** | +7 |
| dropped | 18 | 18 | 0 |
| cards | 96 | **97** | +1 |
| `t332` own state | `queued` | **`picked`** | moved |

**The delta reconciles exactly.** Seven cards moved `queued → picked`:

```
t294 t299 t318 t332 t336 t343 t362
```

and one card was added `queued`: **`t363`** (absent from the plan-phase snapshot and from the
frozen `backlog.json`; present in the live store). 68 − 7 + 1 = 62. ✓

This matches the lead's own correction of the dispatch (measured `queued 61 / picked 17` at its
send time); `t363` was admitted between that message and this capture, which accounts for the
remaining 1.

**Consequence for scope.** `t332` is now `picked`, so the queued set no longer contains it and the
`− 1` self-exclusion term of AC-BH-003 evaluates to zero:

> in-scope = queued − {t332} = 62 − 0 = **62**

AC-BH-003's arithmetic is stated as "(snapshot queued count − 1)" on the plan-phase assumption that
`t332` sits in the queued set. It does not any more. The criterion's **intent** — one entry per
queued card other than this one — is unchanged and is what the report satisfies: **62 entries**.

## §4 In-scope id list (62) — the authority for M3's partition

```
t90  t125 t154 t191 t196 t201 t204 t216 t223 t224 t231 t233 t236 t237 t239 t240
t242 t243 t244 t247 t248 t252 t253 t254 t255 t260 t262 t263 t264 t280 t281 t284
t286 t287 t288 t295 t296 t297 t300 t302 t304 t305 t313 t315 t319 t320 t323 t324
t325 t327 t329 t337 t339 t344 t345 t347 t348 t353 t359 t360 t361 t363
```

## §5 Exclusion lists, loaded before the first card read (plan.md §C.4)

**picked (17) — live owning lanes, not read, not proposed:**

```
t278 t294 t299 t318 t332 t333 t336 t338 t341 t343 t346 t350 t354 t356 t357 t358 t362
```

**dropped (18) — already decided, never re-litigated, no un-drop proposed:**

```
t6 t7 t10 t18 t87 t177 t226 t245 t251 t256 t258 t275 t277 t285 t307 t309 t312 t328
```

Note the plan-phase picked list (10 ids) is a **subset** of the run-phase list (17). AC-BH-002's
intersection test is run against the run-phase list, which is the stricter of the two.

## §6 Opening card-row digest (REQ-BH-007, opening half)

**Step 1 — resolve the primary checkout** (one plain call):

```
$ git rev-parse --path-format=absolute --git-common-dir
/Users/goos/MoAI/moai-adk-go/.git
```

Parent → `/Users/goos/MoAI/moai-adk-go`. Store at
`/Users/goos/MoAI/moai-adk-go/.moai/state/todo/backlog.db`.

**Step 2 — digest, resolved path passed as a literal:**

```
$ sqlite3 -json /Users/goos/MoAI/moai-adk-go/.moai/state/todo/backlog.db \
    "select id,state,text from items order by id;" | shasum -a 256
86a3fb05de900816670a1a4b6e6f3f3e9c6af8e6b4cc6eb3997d0e268bffdc20  -
```

**OPENING DIGEST = `86a3fb05de900816670a1a4b6e6f3f3e9c6af8e6b4cc6eb3997d0e268bffdc20`**
(captured 2026-08-30T11:19Z, Step 1 output recorded above.)

The observable is the SQLite `items` table, not `backlog.json`. Grounds, the migration evidence, and
the four measured controls are in `00-tooling-baseline.md` § Finding B1'. In short: `backlog.json`
stopped being written at the t306 SQLite migration, so a digest over it cannot red — it returned the
plan-phase value `56e1387e…b346b` unchanged after the queue had already moved seven cards and gained
one. acceptance.md AC-BH-006 is repaired to name the live store (v0.3.0).

M5 re-captures with the identical two-step form and records its own Step 1 output.

## §7 Evidence sections for this milestone

**Claim.** The run-phase snapshot is untruncated, its counts partition the card set exactly, the
delta against plan-phase reconciles to the card, and the opening digest is taken against the store
the queue actually writes.

**Evidence.** The commands and verbatim outputs of §2, §3, §6 above, plus the independent
`sqlite3 count(*)` = 97 cross-check.

**Baseline-attribution.** Every figure measured in this run, in worktree
`.claude/worktrees/t332` at HEAD `6165f9f5e`, against store
`/Users/goos/MoAI/moai-adk-go/.moai/state/todo/backlog.db` and the refs pinned in
`00-tooling-baseline.md`. The plan-phase figures are cited only as the left column of the delta
table and decide nothing.

**Gaps.**
- The 4 relation rows were counted by subtraction (101 − 97) and by `uniq -c` grouping them
  separately; their individual content is read at M4, not here.
- Whether `t363` is genuinely new or a re-add of a previously-dropped id was not checked here; it is
  a card entry in M3 and settled there.
- No check was made that the plan-phase snapshot itself was untruncated — it is superseded and
  decides nothing.

**Residual risk.** Another session can admit or move a card after 11:16:45Z. Such a card is outside
this snapshot by construction and would show as a digest change at M5; that change would be
correctly attributable to the other session rather than to this sweep, and the M5 comparison records
the distinction rather than asserting the sweep is innocent.

---

## §8 Closing card-row digest (REQ-BH-007, closing half — captured at M5)

**Step 1 — resolve the primary checkout** (re-run, not carried over from §6):

```
$ git rev-parse --path-format=absolute --git-common-dir
/Users/goos/MoAI/moai-adk-go/.git
```

Identical to the §6 capture, so both digests were taken against the same tree.

**Step 2 — digest:**

```
$ sqlite3 -json /Users/goos/MoAI/moai-adk-go/.moai/state/todo/backlog.db \
    "select id,state,text from items order by id;" | shasum -a 256
053e99917e1f4bd60ed088c3e1308664c0f38221f4a6cfed48dbba490de57b33  -
```

| | Digest | Captured |
|---|---|---|
| opening (M2) | `86a3fb05de900816670a1a4b6e6f3f3e9c6af8e6b4cc6eb3997d0e268bffdc20` | 2026-08-30T11:19Z |
| closing (M5) | `053e99917e1f4bd60ed088c3e1308664c0f38221f4a6cfed48dbba490de57b33` | 2026-08-30T11:43Z |

**The digests differ.** Per AC-BH-006's attribution branch (v0.3.0) this is the expected state on a
shared live queue, and the criterion is decided by naming the movers — not by the hash.

### §8.1 Attribution — exactly one card differs, and it is not this sweep's

Id-level diff between the opening snapshot and the closing store:

```
$ comm -23 snapshot-open-ids.txt store-close-ids.txt      # present at open, absent at close
t346
$ comm -13 snapshot-open-ids.txt store-close-ids.txt      # absent at open, present at close
(empty)
```

| Card | M2 value | M5 value | Attributed to |
|---|---|---|---|
| `t346` | `picked` | **absent** (closed) | the lead, closing a completed card |

Attribution evidence — `t346` is `SPEC-CI-DOCTOR-BIN-001`, whose sync-phase close is already on the
pinned integration ref:

```
$ git log ee50984abe4f11ac337382b48a26328f091e200a --perl-regexp --grep='\bt346\b' --oneline
282daef19 chore(SPEC-CI-DOCTOR-BIN-001): backfill sync_commit_sha (t346)
```

`t346` was on this sweep's **picked exclusion list** (§5) from before the first card was read: it was
never read, never re-verified, and appears in no disposition row. Store counts move
`97 → 96` cards and `picked 17 → 16` accordingly; `queued` is unchanged at **62** and `dropped`
unchanged at **18**.

### §8.2 The decisive check — no in-scope card moved

The digest says *something* changed; this says *what did not*. Projecting both captures to the 62
in-scope rows and diffing:

```
$ diff open-inscope.tsv close-inscope.tsv
(no output, exit 0)
IN-SCOPE ROWS BYTE-IDENTICAL: all 62 cards unchanged in id, state, and text
```

No in-scope card was dropped, edited, closed, reordered, unpicked, or picked. The one differing card
is out of scope and attributed to another actor with a cited commit.

**AC-BH-006 verdict: PASS via the attribution branch.** No differing card is one this sweep could
have mutated, and no difference is left unattributed — the "some other session probably did it"
shrug the criterion forbids is not being relied on here; `282daef19` is the citation.

### §8.3 `findings` count unchanged — the negative control, observed live

```
$ sqlite3 backlog.db "select count(*) from findings;"   →  2   (M2: 2, M5: 2)
```

The sweep recorded zero relations (M4 §3.1 — the one attempt was refused as a duplicate), so the
`findings` table is unmoved. The negative control measured at baseline (a `findings` insert leaves
the `items` projection unchanged) was therefore not exercised by this run's own ordering; it remains
established by the scratch-copy probe in `00-tooling-baseline.md` § Finding B1'. Recorded as a gap
rather than claimed as a live observation.
