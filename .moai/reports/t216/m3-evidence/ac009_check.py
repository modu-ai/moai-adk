"""AC-HWD-009 machine check.

For each of the 11 present-but-unwired wrappers, assert the rule file carries
exactly ONE row that names that wrapper and no other of the 11, and that the row
carries exactly ONE of the five disposition classes, and that the class is the
expected one. Run against both sides of the template pair.
"""

expected = {
    "chain-event.sh": "reachable-via-template-settings",
    "handle-agent-hook.sh": "reachable-via-agent-frontmatter",
    "handle-session-start-compact.sh": "reachable-via-in-binary-registry",
    "handle-elicitation.sh": "dead-by-decision",
    "handle-elicitation-result.sh": "dead-by-decision",
    "handle-notification.sh": "dead-by-decision",
    "handle-task-created.sh": "dead-by-decision",
    "handle-worktree-create.sh": "dead-by-decision",
    "handle-worktree-remove.sh": "dead-by-decision",
    "handle-session-start-navigator.sh": "open-question",
    "team-ac-verify.sh": "open-question",
}

classes = [
    "reachable-via-template-settings",
    "reachable-via-agent-frontmatter",
    "reachable-via-in-binary-registry",
    "dead-by-decision",
    "open-question",
]

# Longest first so a name that contains a shorter one is consumed as itself.
names = sorted(expected, key=len, reverse=True)

paths = [
    ".claude/rules/moai/development/hook-independence.md",
    "internal/template/templates/.claude/rules/moai/development/hook-independence.md",
]

failures = 0
for path in paths:
    print("==", path)
    lines = open(path).read().split("\n")
    for script, cls in expected.items():
        rows = [
            (i + 1, text)
            for i, text in enumerate(lines)
            if text.startswith("| `" + script + "`")
        ]
        if len(rows) != 1:
            print("  FAIL %s: %d own-row(s)" % (script, len(rows)))
            failures += 1
            continue
        ln, row = rows[0]
        named = set()
        rest = row
        for n in names:
            if n in rest:
                named.add(n)
                rest = rest.replace(n, "")
        found = [c for c in classes if c in row]
        if named != {script}:
            print("  FAIL %s line %d: also names %s" % (script, ln, sorted(named - {script})))
            failures += 1
        elif len(found) != 1:
            print("  FAIL %s line %d: %d classes %s" % (script, ln, len(found), found))
            failures += 1
        elif found[0] != cls:
            print("  FAIL %s line %d: class %s != expected %s" % (script, ln, found[0], cls))
            failures += 1
        else:
            print("  ok   %-36s line %4d  %s" % (script, ln, cls))

print("TOTAL FAILURES:", failures)
