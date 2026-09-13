# Lane A
{
git fetch origin main 2>&1
fetch_status=$?
if [ "$fetch_status" -ne 0 ]; then
  exit "$fetch_status"
fi
} &
git rev-list --count --left-right origin/main...HEAD

moai session list --json --filter-spec=<SPEC-ID>
