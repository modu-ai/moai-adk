# M2 attempt — .claude/rules/moai/core/moai-constitution.md | URL Verification

surface = init
pre_chars = 475
post_chars = 406
chars = 69
post_hash = 93de7321ea747c584768af09d1908ad3e36fd076073a6f2a9d6fb5b7a22b4477

post_hash is the frozen-multiset sha256 of the whole post-attempt surface (every attempted M2 row replaced, every ADMIT M1 row removed), measured with the AC-ALH-004 pipeline on $SCRATCH/post-init.

## before

````
## URL Verification

All URLs must be verified before inclusion in responses.

Rules:
- Use WebFetch to verify URLs from WebSearch results
- Mark unverified information as uncertain
- Include Sources section when WebSearch is used
- Under a GLM backend (`moai glm`), URL verification uses `mcp__web_reader__webReader` and search uses `mcp__web_search_prime__webSearchPrime` instead of the built-in `WebFetch` / `WebSearch` (see `.claude/rules/moai/core/glm-web-tooling.md`)

````

## after

````
## URL Verification

Verify every URL before including it: WebFetch the URLs WebSearch returned, mark unverified information as uncertain, include a Sources section when WebSearch was used. Under a GLM backend (`moai glm`) verification uses `mcp__web_reader__webReader` and search `mcp__web_search_prime__webSearchPrime` instead of `WebFetch` / `WebSearch` (`.claude/rules/moai/core/glm-web-tooling.md`).

````
