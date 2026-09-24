#!/bin/sh
# check-armed.sh — report when the commit-time AC-snapshot guard is NOT armed
# (card t1161, follow-up to t1150 F9). Local-only dev tool: no template mirror,
# never distributed, not part of any product surface.
#
# Why this exists: every way the guard can stop working is silent. Deleting
# either git config key, running a git older than the config-defined-hooks
# floor, or losing the checker file all leave commits passing exactly as they
# would with a guard that ran and found nothing. Silence therefore carries no
# information, and the only fix is a surface that speaks when the guard cannot.
#
# Read-only: reads two git config keys, the git version, and one file's
# existence. Writes nothing, changes no config, blocks nothing. Always exits 0
# so a SessionStart wiring can never stall or fail a session.
#
# Output contract: SILENT when armed. One line per fault to stderr otherwise,
# each naming what is missing and how to repair it. Callers that want a machine
# answer read the exit-0 stdout marker instead:
#   ARMED            — all checks passed
#   NOT-ARMED <n>    — n faults found (the stderr lines say which)

name=ac-baseline-guard
need_major=2
need_minor=54
checker=scripts/ac-baseline/check-staged.sh
faults=0

warn() {
	printf 'ac-baseline-guard: %s\n' "$1" >&2
	faults=$((faults + 1))
}

# Not a git repository (or git absent): the guard is not a concept here, so this
# is not a fault — report nothing and leave. A warning here would fire in every
# unrelated directory a SessionStart hook happens to start in.
git rev-parse --git-dir >/dev/null 2>&1 || {
	echo "ARMED"
	exit 0
}

# 1. The two config keys. --get (not --get-all) is deliberate: install-hook.sh
#    uses --replace-all, so a second value means someone edited the config by
#    hand and the extra value is itself worth reporting.
if [ -z "$(git config --get "hook.$name.event" 2>/dev/null)" ]; then
	warn "NOT ARMED: git config hook.$name.event is unset — run scripts/ac-baseline/install-hook.sh"
fi
if [ -z "$(git config --get "hook.$name.command" 2>/dev/null)" ]; then
	warn "NOT ARMED: git config hook.$name.command is unset — run scripts/ac-baseline/install-hook.sh"
fi

# 2. The git version floor. Config-defined hooks are what this guard rides on;
#    below the floor the keys are present and simply never consulted, which is
#    the most deceptive of the three faults.
ver=$(git --version 2>/dev/null)
num=$(printf '%s\n' "$ver" | awk '{ print $3 }')
major=${num%%.*}
rest=${num#*.}
minor=${rest%%.*}
case "$major:$minor" in
*[!0-9:]* | :* | *:)
	warn "cannot parse git version from '$ver' — cannot confirm config-defined hooks are supported (need >= $need_major.$need_minor)"
	;;
*)
	if [ "$major" -lt "$need_major" ] || { [ "$major" -eq "$need_major" ] && [ "$minor" -lt "$need_minor" ]; }; then
		warn "NOT ARMED: git $num does not support config-defined hooks (need >= $need_major.$need_minor) — the config keys are present but never consulted"
	fi
	;;
esac

# 3. The checker file, resolved against the repository top level rather than the
#    caller's cwd: a SessionStart hook does not promise to start at the top.
top=$(git rev-parse --show-toplevel 2>/dev/null)
if [ -n "$top" ] && [ ! -f "$top/$checker" ]; then
	warn "$checker is absent from this tree — every commit here prints NOT CHECKED and is never blocked, until the tree absorbs develop"
fi

if [ "$faults" -eq 0 ]; then
	echo "ARMED"
else
	echo "NOT-ARMED $faults"
fi
exit 0
