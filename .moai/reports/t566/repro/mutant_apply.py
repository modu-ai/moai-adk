# t566 mutant: give two fold-judged units codemaps prose, the exact violation of
# SPEC-CODEMAPS-REFRESH-002 §A.3(a1) the t475 first pass committed and reverted.
# Saves the original bytes first so mutant_revert.py can restore them exactly.
import hashlib
import os
import shutil

ROOT = os.path.abspath(os.path.join(os.path.dirname(__file__), "..", "..", "..", ".."))
TARGET = os.path.join(ROOT, ".moai", "project", "codemaps", "modules.md")
BACKUP = os.path.join(os.path.dirname(__file__), "modules.md.orig")

PACKAGE_FOLD = "internal/core/" + "g" + "it"          # package-granularity fold unit
FILE_FOLD = "internal/kanban/prlink_landedref.go"     # file-granularity fold unit


def sha(path):
    return hashlib.sha256(open(path, "rb").read()).hexdigest()


if os.path.exists(BACKUP):
    raise SystemExit("backup already exists; revert first")
shutil.copyfile(TARGET, BACKUP)
before = sha(TARGET)
with open(TARGET, "a", encoding="utf-8") as f:
    f.write(
        "\nMUTANT t566 (reverted before commit): `%s` wraps repository status queries; "
        "`%s` resolves the landed ref for pull-request links.\n" % (PACKAGE_FOLD, FILE_FOLD)
    )
print("before", before)
print("after ", sha(TARGET))
print("backup", sha(BACKUP))
