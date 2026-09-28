# M2 attempt — .claude/rules/moai/core/verification-claim-integrity.md | 1. The Invariant — no unobserved-claim (verification, defect, OR premise) ¶6

surface = init
pre_chars = 238
post_chars = 204
chars = 34
post_hash = 93de7321ea747c584768af09d1908ad3e36fd076073a6f2a9d6fb5b7a22b4477

post_hash is the frozen-multiset sha256 of the whole post-attempt surface (every attempted M2 row replaced, every ADMIT M1 row removed), measured with the AC-ALH-004 pipeline on $SCRATCH/post-init.

## before

````
This direction is the more dangerous one, because its failure is silent. A wrong "remove it" claim is contradicted by the next build or test run; a wrong "keep it" claim preserves dead code and is never contradicted by any signal at all.
````

## after

````
This direction is the more dangerous because it fails silently: a wrong "remove it" is contradicted by the next build or test run; a wrong "keep it" preserves dead code and no signal ever contradicts it.
````
