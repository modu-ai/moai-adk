# t660 mutant: delete every decision: ask entry from queryFilterFixture in
# internal/cli/tool_policy_test.go. Saves the original bytes first so
# mutant_revert.py can restore them exactly.
import hashlib
import os
import re
import shutil

ROOT = os.path.abspath(os.path.join(os.path.dirname(__file__), "..", "..", "..", ".."))
TARGET = os.path.join(ROOT, "internal", "cli", "tool_policy_test.go")
BACKUP = os.path.join(os.path.dirname(__file__), "tool_policy_test.go.orig")
OPEN = "const queryFilterFixture = `"


def sha(path):
    return hashlib.sha256(open(path, "rb").read()).hexdigest()


if os.path.exists(BACKUP):
    raise SystemExit("backup already exists; revert first")
src = open(TARGET, encoding="utf-8").read()
start = src.index(OPEN) + len(OPEN)
end = src.index("`", start)
fixture = src[start:end]
entries = re.split(r"(?m)^(?=  - tool: )", fixture)
kept = [e for e in entries if "decision: ask" not in e]
mutated = "".join(kept)
shutil.copyfile(TARGET, BACKUP)
before = sha(TARGET)
with open(TARGET, "w", encoding="utf-8") as f:
    f.write(src[:start] + mutated + src[end:])
print("entries_before", sum(1 for e in entries if e.startswith("  - tool: ")))
print("entries_after ", sum(1 for e in kept if e.startswith("  - tool: ")))
print("ask_lines_before", fixture.count("decision: ask"))
print("ask_lines_after ", mutated.count("decision: ask"))
print("before", before)
print("after ", sha(TARGET))
print("backup", sha(BACKUP))
