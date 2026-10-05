# M2 attempt — .claude/rules/moai/core/verification-claim-integrity.md | 2.2 Tool-provenance attribution — which build judged the tree ¶2

surface = live
pre_chars = 437
post_chars = 312
chars = 125
post_hash = d97b33d960c9801d4ec145ca263ed788425b337f43c585594c8d527c1318c6c3

post_hash is the frozen-multiset sha256 of the whole post-attempt surface (every attempted M2 row replaced, every ADMIT M1 row removed), measured with the AC-ALH-004 pipeline on $SCRATCH/post-live.

## before

````
The failure is silent by construction, and its silence is **symmetric**. A build behind the tree simply does not run the rules that landed after it — clean pass, exit zero, empty error stream; a build matching the tree produces *the same three signals*. Nothing in either result says which case occurred, so a green result is not evidence the checks passed, only that whatever checks the invoked build happens to carry reported nothing.
````

## after

````
The failure is silent and **symmetric**: a build behind the tree skips the rules that landed after it — clean pass, exit zero, empty error stream — exactly *the same three signals* a matching build gives. So a green result shows only that the invoked build's checks reported nothing, not that the checks passed.
````
