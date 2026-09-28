# M2 attempt — .claude/rules/moai/core/moai-constitution.md | URL Verification

surface = live
pre_chars = 497
post_chars = 428
chars = 69
post_hash = d97b33d960c9801d4ec145ca263ed788425b337f43c585594c8d527c1318c6c3

post_hash is the frozen-multiset sha256 of the whole post-attempt surface (every attempted M2 row replaced, every ADMIT M1 row removed), measured with the AC-ALH-004 pipeline on $SCRATCH/post-live.

## before

````
## URL Verification

All URLs must be verified before inclusion in responses.

Rules:
- Use WebFetch to verify URLs from WebSearch results
- Mark unverified information as uncertain
- Include Sources section when WebSearch is used
- Under a GLM backend (`moai glm` / `moai cg` GLM panes), URL verification uses `mcp__web_reader__webReader` and search uses `mcp__web_search_prime__webSearchPrime` instead of the built-in `WebFetch` / `WebSearch` (see `.claude/rules/moai/core/glm-web-tooling.md`)

````

## after

````
## URL Verification

Verify every URL before including it: WebFetch the URLs WebSearch returned, mark unverified information as uncertain, include a Sources section when WebSearch was used. Under a GLM backend (`moai glm` / `moai cg` GLM panes) verification uses `mcp__web_reader__webReader` and search `mcp__web_search_prime__webSearchPrime` instead of `WebFetch` / `WebSearch` (`.claude/rules/moai/core/glm-web-tooling.md`).

````
