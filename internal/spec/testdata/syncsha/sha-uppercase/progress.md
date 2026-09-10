# SPEC-SSFA-001 — Progress

## §E.2 Run-phase Evidence

Fixture: an uppercase SHA — 40 characters, inside the length band, and the
shipped alphabet accepts it (`[0-9a-fA-F]`). This fixture is the alphabet axis
(SPEC-SYNCSHA-BAND-BOUNDARY-001 §C residual, card t397): with only lowercase
fixtures present, a mutant narrowing the alphabet to lowercase-only
(`[0-9a-f]`) stayed green, indistinguishable from the shipped rule. The shipped
rule must stay silent on this line; a mutant that narrows the alphabet must
flag it.

## §E.4 Sync-phase Audit-Ready Signal

sync_commit_sha: ABCDEF0123456789ABCDEF0123456789ABCDEF01
