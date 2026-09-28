# M2 attempt — AGENTS.md | 4. How verification is run

surface = live
pre_chars = 1487
post_chars = 1235
chars = 252
post_hash = d97b33d960c9801d4ec145ca263ed788425b337f43c585594c8d527c1318c6c3

post_hash is the frozen-multiset sha256 of the whole post-attempt surface (every attempted M2 row replaced, every ADMIT M1 row removed), measured with the AC-ALH-004 pipeline on $SCRATCH/post-live.

## before

````
## 4. How verification is run

**Scope verification to the change**: run the tests the change can affect, then push and let CI run
the full suite. A full-suite run on a loaded developer machine measures the machine, not the code.

**Never spawn background load.** Where a verification needs contention, the load must be
cleanup-guaranteed — kills registered with the test framework's cleanup hook, or a `timeout`
wrapper bounding the process from outside. A trailing `kill` is not cleanup.

**Scrub the environment in one compound invocation.** Inside a worktree, an environment-scrubbed
verification runs as a single `unset <VARS> && <command>` call; a separate `unset` does not carry
into the next command — each invocation is a fresh process.

**Batch independent read-only verifications rather than serializing them** across turns. Serialize
only for a genuine dependency: one command's output feeding another, writes to the same path, or
shared-state mutation.

**A CodeRabbit row in `gh pr checks` is not evidence that a review ran** — the status reads
`success` and prints `pass` identically whether or not one did. Count the row only when BOTH hold:
(1) `gh api "repos/$repo/commits/$head_sha/status"` reports the `CodeRabbit` context with
`state == "success"` and description `Review completed`; (2) a `Merge Risk:` line exists whose
commit prefix matches the current `headRefOid`. Anything else is a gap, not a pass; `Review rate
limited` means the review never started.

---

````

## after

````
## 4. How verification is run

**Scope verification to the change**: run the tests it can affect, then push and let CI run the
full suite (a full suite on a loaded machine measures the machine).

**Never spawn background load.** Contention load must be cleanup-guaranteed — kills registered with
the test framework's cleanup hook, or a `timeout` wrapper; a trailing `kill` is not cleanup.

**Scrub the environment in one compound invocation**: inside a worktree, `unset <VARS> && <command>`
as a single call — a separate `unset` does not reach the next command (fresh process each call).

**Batch independent read-only verifications** instead of serializing across turns; serialize only
for a genuine dependency (output feeding input, writes to the same path, shared-state mutation).

**A CodeRabbit row in `gh pr checks` is not evidence a review ran** — it reads `success`/`pass`
either way. Count it only when BOTH hold: (1) `gh api "repos/$repo/commits/$head_sha/status"`
reports the `CodeRabbit` context with `state == "success"` and description `Review completed`;
(2) a `Merge Risk:` line exists whose commit prefix matches the current `headRefOid`. Anything else
is a gap; `Review rate limited` means the review never started.

````
