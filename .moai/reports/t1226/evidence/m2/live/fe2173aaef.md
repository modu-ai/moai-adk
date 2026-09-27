# M2 attempt — .claude/rules/moai/core/verification-claim-integrity.md | 1. The Invariant — no unobserved-claim (verification, defect, OR premise) ¶6

surface = live
pre_chars = 238
post_chars = 204
chars = 34
post_hash = d97b33d960c9801d4ec145ca263ed788425b337f43c585594c8d527c1318c6c3

post_hash is the frozen-multiset sha256 of the whole post-attempt surface (every attempted M2 row replaced, every ADMIT M1 row removed), measured with the AC-ALH-004 pipeline on $SCRATCH/post-live.

## before

````
This direction is the more dangerous one, because its failure is silent. A wrong "remove it" claim is contradicted by the next build or test run; a wrong "keep it" claim preserves dead code and is never contradicted by any signal at all.
````

## after

````
This direction is the more dangerous because it fails silently: a wrong "remove it" is contradicted by the next build or test run; a wrong "keep it" preserves dead code and no signal ever contradicts it.
````
