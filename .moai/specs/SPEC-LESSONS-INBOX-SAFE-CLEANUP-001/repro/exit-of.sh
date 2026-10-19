#!/usr/bin/env bash
# exit-of.sh <command...> — run a read-only command and print its exit status as its own line.
# The single-invocation cell form forbids `; echo $?`, so the exit code is recorded by this wrapper.
# Output: the command's own stdout/stderr, then a final line `exit=<N>`. The wrapper itself exits 0.
"$@"
echo "exit=$?"
exit 0
