#!/usr/bin/env bash
# Gate a native Windows go-test JSON stream; compilation is not runtime evidence.
set -euo pipefail
base="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
stream="${1:?test stream required}"
manifest="${2:-$base/windows-gateway-required.json}"
if [[ ! -s "$stream" ]]; then echo 'FAIL Windows gateway test stream missing/empty' >&2;exit 1;fi
jq -n -e --slurpfile wanted "$manifest" '
  ($wanted[0]) as $req |
  if ($req|type)!="array" or ($req|length)==0 then error("empty required test set") else . end |
  (reduce inputs as $e ({tests:{},packages:{}};
    if ($e.Action=="run" or $e.Action=="pass" or $e.Action=="skip" or $e.Action=="fail") then
      if $e.Test != null and any($req[]; .Package==$e.Package and .Test==$e.Test) then
        ($e.Package+"/"+$e.Test) as $key | .tests[$key] += [$e.Action]
      elif $e.Test == null and any($req[]; .Package==$e.Package) then
        .packages[$e.Package] += [$e.Action]
      else . end
    else . end
  )) as $seen |
  [$req[] | . as $r | ($r.Package+"/"+$r.Test) as $key |
    select($seen.tests[$key] != ["run","pass"] or
      ($seen.packages[$r.Package] | map(select(.=="pass" or .=="skip" or .=="fail"))) != ["pass"]) |
    $key] as $bad |
  if ($bad|length)>0 then error("required Windows tests not run/pass with passing package: "+($bad|join(", ")))
  else "PASS Windows gateway native evidence: \($req|length) required tests" end
' "$stream"
