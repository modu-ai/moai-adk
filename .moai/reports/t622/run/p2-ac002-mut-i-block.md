# 1. Fetch latest origin/main without merging
git fetch origin main 2>&1

# 2. Count divergence between local HEAD and origin/main
git rev-list --count --left-right origin/main...HEAD

# 3. Query active sessions on this host for the same SPEC scope (L1 of the
#    canonical 4-layer multi-session race mitigation policy).
moai session list --json --filter-spec=<SPEC-ID>
