# Lane A — ordered; wait for fetch completion before reading origin/main.
git fetch origin main 2>&1
git rev-list --count --left-right origin/main...HEAD

# Lane B — can be started while Lane A is fetching, then joined before the
# divergence/session decision is surfaced (L1 of the canonical 4-layer policy).
moai session list --json --filter-spec=<SPEC-ID>
