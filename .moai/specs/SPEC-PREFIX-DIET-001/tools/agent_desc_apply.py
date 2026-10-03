#!/usr/bin/env python3
"""Replace the `description:` block of selected agents (SPEC-PREFIX-DIET-001, M5).

Usage (worktree root): python3 .moai/specs/SPEC-PREFIX-DIET-001/tools/agent_desc_apply.py <new.json> [--dry-run]
<new.json> maps agent name -> new description block text (the lines after `description:`, newline-terminated).
The template file is edited first; the local mirror receives the identical block (only the description block
is touched in either file). Prints before/after UTF-16 sizes.
"""
import json
import re
import sys

TEMPLATE = "internal/template/templates/.claude/agents/moai/%s.md"
LOCAL = ".claude/agents/moai/%s.md"
BLOCK = re.compile(r"^description:(.*?)(?=^\S)", re.S | re.M)


def u16(s):
    return len(s.encode("utf-16-le")) // 2


def replace(path, new_block, dry):
    text = open(path, encoding="utf-8").read()
    head, sep, rest = text.partition("---\n")  # leading "" + frontmatter start
    assert head == "" and sep, path
    fm, sep2, body = rest.partition("\n---\n")
    assert sep2, path
    fm += "\n"
    m = BLOCK.search(fm)
    assert m, path
    old = m.group(1)
    fm_new = fm[: m.start(1)] + new_block + fm[m.end(1) :]
    out = "---\n" + fm_new.rstrip("\n") + "\n---\n" + body
    if not dry:
        open(path, "w", encoding="utf-8").write(out)
    return u16(old), u16(new_block)


if __name__ == "__main__":
    spec = json.load(open(sys.argv[1], encoding="utf-8"))
    dry = "--dry-run" in sys.argv
    for name, block in spec.items():
        for path in (TEMPLATE % name, LOCAL % name):
            before, after = replace(path, block, dry)
            print("%-16s %-62s %5d -> %5d" % (name, path, before, after))
