# M2 attempt — AGENTS.md | 5. Core behaviors ¶6

surface = init
pre_chars = 387
post_chars = 317
chars = 70
post_hash = 93de7321ea747c584768af09d1908ad3e36fd076073a6f2a9d6fb5b7a22b4477

post_hash is the frozen-multiset sha256 of the whole post-attempt surface (every attempted M2 row replaced, every ADMIT M1 row removed), measured with the AC-ALH-004 pipeline on $SCRATCH/post-init.

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
