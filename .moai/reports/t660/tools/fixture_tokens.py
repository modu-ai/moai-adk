# t660 fixture token check: for each entry in queryFilterFixture, the filter
# tokens may appear only in the field that the filter actually matches.
# Control: a synthetic entry whose audit carries "ask" must be flagged.
import os
import re

ROOT = os.path.abspath(os.path.join(os.path.dirname(__file__), "..", "..", "..", ".."))
TARGET = os.path.join(ROOT, "internal", "cli", "tool_policy_test.go")
OPEN = "const queryFilterFixture = `"
TOKENS = {"irreversible": "risk_tier", "allow": "decision", "deny": "decision",
          "ask": "decision", "Bash": "tool"}


def parse(block):
    return dict(re.findall(r'(?m)^\s*-?\s*(\w+): "?([^"\n]*)"?$', block))


def leaks(fields):
    out = []
    for tok, home in TOKENS.items():
        for name, val in fields.items():
            if name != home and tok in val:
                out.append((tok, name, val))
    return out


ctl = parse('  - tool: "Read"\n    decision: allow\n    audit: "task list"\n')
print("control_leaks", len(leaks(ctl)))
src = open(TARGET, encoding="utf-8").read()
start = src.index(OPEN) + len(OPEN)
fixture = src[start:src.index("`", start)]
blocks = [b for b in re.split(r"(?m)^(?=  - tool: )", fixture) if b.startswith("  - tool: ")]
total = 0
for b in blocks:
    f = parse(b)
    found = leaks(f)
    total += len(found)
    print(f.get("tool"), f.get("risk_tier"), f.get("decision"), "leaks", found)
print("entries", len(blocks), "total_leaks", total)
