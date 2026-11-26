# Moves every plain `cargo test --offline --locked --test a --test b ...` command of a scan's
# `checks` into the sorted, de-duplicated `test_targets` array (one name per line once printed).
# Other commands stay in `checks`; a scan with no plain command is left as it is. Idempotent.
# Run: jq -f convert.jq tools/repo_scans.json   (jq's 2-space output equals the table's form).
def plain: test("^cargo test --offline --locked( --test [a-z0-9_]+)+$");
.scans |= map(
  if has("checks") and ([.checks[] | select(plain)] | length) > 0 then
    . as $s
    | ([$s.checks[] | select(plain) | scan(" --test ([a-z0-9_]+)") | .[0]] + ($s.test_targets // []) | unique) as $names
    | ([$s.checks[] | select(plain | not)]) as $keep
    | reduce ($s | keys_unsorted[]) as $k ({};
        if $k == "checks" then
          . + {test_targets: $names} + (if ($keep | length) > 0 then {checks: $keep} else {} end)
        elif $k == "test_targets" then .
        else . + {($k): $s[$k]} end)
  else . end
)
