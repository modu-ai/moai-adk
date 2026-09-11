# Lane A — ordered; wait for fetch completion before reading origin/main.
git fetch origin main 2>&1
fetch_status=$?
if [ "$fetch_status" -ne 0 ]; then
  printf 'pre-spawn sync blocked: fetch origin/main failed (status=%s)\n' "$fetch_status" >&2
  exit "$fetch_status"
fi
git rev-list --count --left-right origin/main...HEAD

# Lane B — MUST NOT start until Lane A has fully finished; never run it
# concurrently with the fetch.
moai session list --json --filter-spec=<SPEC-ID>
