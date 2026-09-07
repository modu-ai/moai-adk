#!/bin/sh
# AC-UMH-013 mutant probe, run against an ISOLATED /tmp corpus.
#
# acceptance.md AC-UMH-014 specifies this mutant as "create
# .agents/skills/mutant-probe in THIS repository, then run the criterion's
# snapshot diff". That instruction collides with REQ-UMH-009 / C-3, which
# forbids this card from writing this repository's .agents/ state (an
# observation subject for sibling cards t498 and t510). C-3 is treated as
# controlling, so the SAME recipe is exercised against a scratch corpus whose
# shape matches: the recipe is corpus-independent shell, and what the mutant
# tests is the recipe's ability to report a difference.
#
# The substitution is recorded as a deviation in progress.md §E.2 rather than
# presented as the literal criterion.
set -e
D=/tmp/t520-ac013-mutant
rm -rf "$D"
mkdir -p "$D/.agents/skills/real-entry" "$D/out"
cd "$D"
# A symlink entry too, so the readlink half of the recipe is exercised.
ln -s ../../real-target .agents/skills/linked-entry

snap() {
	P="out/agents-$1.txt"
	: >"$P"
	test -e .agents
	echo "agents_exists_rc=$?" >>"$P"
	find .agents -print >"out/agents-$1.raw"
	find_rc=$?
	echo "find_rc=$find_rc" >>"$P"
	LC_ALL=C sort "out/agents-$1.raw" >>"$P"
	find .agents -type l -print >"out/agents-$1.links"
	link_rc=$?
	echo "link_rc=$link_rc" >>"$P"
	while read -r l; do printf '%s -> %s\n' "$l" "$(readlink "$l")"; done \
		<"out/agents-$1.links" | LC_ALL=C sort >>"$P"
}

snap before
echo "--- before snapshot ---"
cat out/agents-before.txt

# THE MUTANT: the forbidden thing happens.
mkdir -p .agents/skills/mutant-probe

snap after
echo "--- after snapshot ---"
cat out/agents-after.txt

echo "--- diff: MUST be non-empty (guard flips red) ---"
if diff out/agents-before.txt out/agents-after.txt; then
	echo "MUTANT NOT CAUGHT: diff reported no difference"
	exit 1
fi
echo "diff reported a difference -> AC-UMH-013 mutant CAUGHT"

echo "--- CONTROL: unmutated re-snapshot must be identical ---"
rm -rf .agents/skills/mutant-probe
snap control
if diff out/agents-before.txt out/agents-control.txt; then
	echo "control OK: the recipe is stable when nothing changes"
else
	echo "CONTROL FAILED: the recipe reports a difference with no change"
	exit 1
fi
