#!/usr/bin/env bash
# Run a plan-auditor Group 4 traceability verb against fixture directories (card t524).
# usage: run-fixtures.sh <verb-file> <out-dir> <work-dir> <fixture-dir>...
# Each fixture dir may hold spec.md and acceptance.md; they are copied into a
# fresh work subdir as new-spec.md / acceptance.md (the verb placeholders), the
# verb runs there, and its combined output plus exit code land in <out-dir>.
verbfile=$1
out=$2
work=$3
shift 3
verb=$(sed -e 's/<new-spec\.md>/new-spec.md/g' -e 's/<acceptance\.md>/acceptance.md/g' "$verbfile")
mkdir -p "$out"
for fx in "$@"; do
  name=$(basename "$fx")
  dir="$work/$name"
  mkdir -p "$dir"
  rm -f "$dir/new-spec.md" "$dir/acceptance.md"
  if [ -f "$fx/spec.md" ]; then cp "$fx/spec.md" "$dir/new-spec.md"; fi
  if [ -f "$fx/acceptance.md" ]; then cp "$fx/acceptance.md" "$dir/acceptance.md"; fi
  (cd "$dir" && bash -c "$verb") > "$out/$name.txt" 2>&1
  echo "exit=$?" > "$out/$name.exit"
done
