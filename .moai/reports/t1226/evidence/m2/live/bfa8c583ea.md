# M2 attempt — AGENTS.md | 5. Core behaviors ¶6

surface = live
pre_chars = 387
post_chars = 317
chars = 70
post_hash = d97b33d960c9801d4ec145ca263ed788425b337f43c585594c8d527c1318c6c3

post_hash is the frozen-multiset sha256 of the whole post-attempt surface (every attempted M2 row replaced, every ADMIT M1 row removed), measured with the AC-ALH-004 pipeline on $SCRATCH/post-live.

## before

````
**6. Verify, don't assume.** Every task requires evidence of completion; "seems right" is never
sufficient. Tests passing means showing the test output; a build succeeding, the build output; a
file created, reading it back; behavior correct, the runtime evidence. For ad-hoc work without a
spec, define the goal as a testable assertion first — "done when X produces Y" — then verify it.
````

## after

````
**6. Verify, don't assume.** Every task needs evidence of completion; "seems right" never suffices:
test output for tests, build output for builds, a read-back for a created file, runtime evidence
for behavior. Without a spec, first define the goal as a testable assertion ("done when X produces
Y"), then verify it.
````
