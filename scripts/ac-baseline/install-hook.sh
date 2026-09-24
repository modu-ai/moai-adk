#!/bin/sh
# install-hook.sh — arm the commit-time AC-snapshot guard (card t1150,
# SPEC-ACSNAPSHOT-COMMIT-GUARD-001). Local-only dev tool: no template mirror,
# never distributed.
#
# Writes exactly two keys into the repository git config:
#   hook.ac-baseline-guard.event   = pre-commit
#   hook.ac-baseline-guard.command = <runs scripts/ac-baseline/check-staged.sh>
# Idempotent (--replace-all leaves one value per key). Never touches
# .git/hooks/*, the managed hook's provenance file, or core.hooksPath: the
# config-defined hook runs alongside the managed hookdir file and still runs
# when core.hooksPath points at /dev/null.
#
# Owner: the lead, once, after the card branch is merged into local develop.
# The repository config is shared by every linked worktree, so one run arms
# every lane. A tree whose branch predates the checker prints a NOT CHECKED
# line on every commit and is never blocked, until it absorbs develop.

name=ac-baseline-guard
need_major=2
need_minor=54

ver=$(git --version 2>/dev/null) || {
	echo "install-hook: git not found on PATH; nothing written" >&2
	exit 1
}
num=$(printf '%s\n' "$ver" | awk '{ print $3 }')
major=${num%%.*}
rest=${num#*.}
minor=${rest%%.*}
case "$major:$minor" in
*[!0-9:]* | :* | *:)
	echo "install-hook: cannot parse git version from '$ver'; nothing written" >&2
	exit 1
	;;
esac
if [ "$major" -lt "$need_major" ] || { [ "$major" -eq "$need_major" ] && [ "$minor" -lt "$need_minor" ]; }; then
	echo "install-hook: found git $num, but config-defined hooks need git >= $need_major.$need_minor; nothing written" >&2
	exit 1
fi

# The hook command runs from the top of whichever worktree is committing.
# Only the checker's exit 1 (a measured mismatch) is forwarded as a rejection;
# any other non-zero exit (a syntax error, a missing tool, a signal) is a tool
# fault and fails open, so one broken checker copy cannot stall every commit.
# The command carries no interpolated path: it is a fixed string.
cmd='if [ -f scripts/ac-baseline/check-staged.sh ]; then sh scripts/ac-baseline/check-staged.sh; rc=$?; if [ "$rc" -eq 0 ] || [ "$rc" -eq 1 ]; then exit "$rc"; fi; echo "ac-baseline-guard: NOT CHECKED (checker exited $rc)" >&2; exit 0; else echo "ac-baseline-guard: NOT CHECKED (scripts/ac-baseline/check-staged.sh absent in this tree; it is armed once the tree absorbs develop)" >&2; fi'

git config --local --replace-all "hook.$name.event" pre-commit || exit 1
git config --local --replace-all "hook.$name.command" "$cmd" || exit 1

echo "install-hook: armed hook.$name (event=pre-commit) in the repository git config" >&2
