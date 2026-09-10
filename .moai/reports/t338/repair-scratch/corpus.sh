#!/usr/bin/env bash
# Corpus sweep: run the adjacency counter over the depth-1 glob and report halts.
cd "$(git rev-parse --show-toplevel)" || exit 1
SCRIPT="$(pwd)/.moai/reports/t338/repair-scratch/counter.py"
files=0
halting=0
for f in .moai/specs/*/acceptance.md; do
    files=$((files + 1))
    out=$(python3 "$SCRIPT" "$f" sameline)
    case "$out" in
        AMBIGUOUS*)
            halting=$((halting + 1))
            echo "HALT $f :: $out"
            ;;
    esac
done
echo "files=$files halting=$halting"
