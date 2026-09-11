# t566: extract the fold/omission judgment verb from the codemaps workflow
# template copy (first bash block under the heading) and run it with bash from
# the repository root. Writes verb/<label>.txt with the output and exit code.
#
# Usage: python3 run_verb.py <label> [--doc <workflow path>]
import os
import subprocess
import sys

ROOT = os.path.abspath(os.path.join(os.path.dirname(__file__), "..", "..", "..", ".."))
DOC = os.path.join(ROOT, "internal", "template", "templates", ".claude", "skills", "moai", "workflows", "codemaps.md")
HEADING = "### Fold and Omission Judgment Check"
FENCE = "`" * 3


def extract(doc):
    in_section = in_block = False
    body = []
    for line in open(doc, encoding="utf-8").read().split("\n"):
        if line.startswith(HEADING):
            in_section = True
        elif in_section and not in_block and line == FENCE + "bash":
            in_block = True
        elif in_block and line == FENCE:
            return "\n".join(body)
        elif in_block:
            body.append(line)
    raise SystemExit("no bash block under heading in " + doc)


label = sys.argv[1]
doc = DOC
if "--doc" in sys.argv:
    doc = os.path.abspath(sys.argv[sys.argv.index("--doc") + 1])
verb = extract(doc)
proc = subprocess.run(["bash", "-c", verb], cwd=ROOT, capture_output=True, text=True)
out_dir = os.path.join(os.path.dirname(__file__), "..", "verb")
os.makedirs(out_dir, exist_ok=True)
text = "doc: %s\n--- stdout ---\n%s--- stderr ---\n%sEXIT=%d\n" % (
    os.path.relpath(doc, ROOT), proc.stdout, proc.stderr, proc.returncode)
with open(os.path.join(out_dir, label + ".txt"), "w", encoding="utf-8") as f:
    f.write(text)
print(text, end="")
