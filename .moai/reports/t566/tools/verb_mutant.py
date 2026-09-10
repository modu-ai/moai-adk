# t566 verb mutant: disable the fold branch of the judgment verb in the codemaps
# workflow template copy (the copy the Go test extracts), or restore it.
#
# Usage: python3 verb_mutant.py apply|revert
import hashlib
import os
import shutil
import sys

ROOT = os.path.abspath(os.path.join(os.path.dirname(__file__), "..", "..", "..", ".."))
TARGET = os.path.join(ROOT, "internal", "template", "templates", ".claude", "skills", "moai", "workflows", "codemaps.md")
BACKUP = os.path.join(os.path.dirname(__file__), "codemaps.md.verb-orig")
LIVE = 'if (kind[i] == "fold" && hits[i] > 0) print "FOLD-PROSE: "'
DEAD = 'if (kind[i] == "fold" && hits[i] < 0) print "FOLD-PROSE: "'


def sha(path):
    return hashlib.sha256(open(path, "rb").read()).hexdigest()


if sys.argv[1] == "apply":
    if os.path.exists(BACKUP):
        raise SystemExit("backup already exists; revert first")
    text = open(TARGET, encoding="utf-8").read()
    if text.count(LIVE) != 1:
        raise SystemExit("fold branch count != 1")
    shutil.copyfile(TARGET, BACKUP)
    with open(TARGET, "w", encoding="utf-8") as f:
        f.write(text.replace(LIVE, DEAD))
    print("original", sha(BACKUP))
    print("mutated ", sha(TARGET))
else:
    if not os.path.exists(BACKUP):
        raise SystemExit("no backup to restore from")
    mutated = sha(TARGET)
    shutil.copyfile(BACKUP, TARGET)
    original = sha(BACKUP)
    os.remove(BACKUP)
    print("mutated ", mutated)
    print("original", original)
    print("restored", sha(TARGET))
    print("RESTORED_MATCHES_ORIGINAL", sha(TARGET) == original)
