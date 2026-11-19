"""AC-4 of SPEC-20261120-statusline-docs-items-2-3: every name the statusline section of
docs/rust-product-path.md gives for a test file, a ledger fixture, a golden case or a test
function exists, and each golden case it names has its own ledger named. Run from the
repository root: python3 -I .moai/specs/SPEC-20261120-statusline-docs-items-2-3/check.py
"""
import json
import pathlib
import re

doc = pathlib.Path("docs/rust-product-path.md").read_text()
section = doc[doc.index("2. Repository:"):doc.index("Directory, PR review state")]
bad = []

paths = set(re.findall(r"tests/[\w./-]+\.(?:rs|py|sqlite)", section))
bad += [p for p in paths if not pathlib.Path(p).exists()]

ledger_dir = pathlib.Path("tests/fixtures/statusline-claude/ledger")
ledgers = set(re.findall(r"cards-[\w-]+\.sqlite", section))
bad += [l for l in ledgers if not (ledger_dir / l).exists()]

manifest = json.load(open("tests/fixtures/statusline-claude/cases.json"))
cases = {c["name"]: c for c in (manifest["cases"] if "cases" in manifest else manifest)}
named = set(re.findall(r"`([ab]\d\d-[\w-]+)`", section))
bad += [g for g in named if g not in cases]
for g in named & set(cases):
    for step in cases[g].get("setup", []):
        if step.get("kind") == "copy":
            ledger = pathlib.PurePath(step["from"]).name
            if ledger not in section:
                bad.append(f"{g}: its ledger {ledger} is not named")

source = "".join(
    p.read_text() for root in ("src", "tests") for p in pathlib.Path(root).rglob("*.rs")
)
bad += [
    f
    for f in set(re.findall(r"`([a-z]+(?:_[a-z0-9]+){3,})`", section))
    if f"fn {f}" not in source
]

assert paths and ledgers and named and not bad, bad
print("ok:", len(paths), "files,", len(ledgers), "ledgers,", len(named), "golden cases")
