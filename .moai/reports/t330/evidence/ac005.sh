#!/bin/bash
# AC-TDG-005 downgrade half: a binary predating the archive must still OPEN a
# database that holds archived rows, rather than aborting on a version mismatch.
# Runs entirely in an isolated repository (acceptance.md §A.1).
set -u
NEW="$1"   # the implemented binary
OLD="$2"   # the binary built from 812ee01fc

REPO="$(mktemp -d)"
cd "$REPO" || exit 9
git init -q .
git config user.email t@example.com
git config user.name t
git commit --allow-empty -q -m init
export CLAUDE_PROJECT_DIR="$REPO"

echo "== isolated repo: $REPO"
"$NEW" todo add "alpha work" || exit 1
"$NEW" todo add "beta work" || exit 1

out=$("$NEW" todo done t1 2>&1); rc=$?
echo "new done t1 -> rc=$rc out=$out"
[ "$rc" -eq 0 ] || exit 1

echo "== archive rows in the database"
sqlite3 "$REPO/.moai/state/todo/backlog.db" "select id, position from archived_items;" 2>/dev/null \
  || echo "(sqlite3 CLI unavailable; the Go-side assertion covers this)"
echo "== stamped schema_version"
sqlite3 "$REPO/.moai/state/todo/backlog.db" "select value from meta where key='schema_version';" 2>/dev/null \
  || echo "(sqlite3 CLI unavailable)"

echo "== OLD binary opens the same database"
out=$("$OLD" todo list 2>&1); rc=$?
echo "old list -> rc=$rc"
echo "$out"
[ "$rc" -eq 0 ] || { echo "FAIL: the pre-archive binary refused the database"; exit 1; }
case "$out" in
  *t2*) : ;;
  *) echo "FAIL: the pre-archive binary does not see the live card t2"; exit 1 ;;
esac
case "$out" in
  *t1*) echo "FAIL: the pre-archive binary shows the archived card as live"; exit 1 ;;
  *) : ;;
esac

echo "== OLD binary can still mutate the queue"
out=$("$OLD" todo add "gamma work" 2>&1); rc=$?
echo "old add -> rc=$rc out=$out"
[ "$rc" -eq 0 ] || { echo "FAIL: the pre-archive binary cannot write"; exit 1; }

echo "PASS: AC-TDG-005 downgrade half"
