# Lane A — ordered; wait for fetch completion before reading origin/main.
git fetch origin main 2>&1 | tee /tmp/prespawn-fetch.log
fetch_status=$?
if [ "$fetch_status" -ne 0 ]; then
  printf 'pre-spawn sync blocked: fetch origin/main failed (status=%s)\n' "$fetch_status" >&2
  exit "$fetch_status"
fi
git rev-list --count --left-right origin/main...HEAD

# Lane B — can be started while Lane A is fetching, then joined before the
# divergence/session decision is surfaced (L1 of the canonical 4-layer policy).
moai session list --json --filter-spec=<SPEC-ID>
