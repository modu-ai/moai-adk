# t660 mutant revert: restores internal/cli/tool_policy_test.go from the saved
# original and proves the bytes match, then removes the backup.
import hashlib
import os
import shutil

ROOT = os.path.abspath(os.path.join(os.path.dirname(__file__), "..", "..", "..", ".."))
TARGET = os.path.join(ROOT, "internal", "cli", "tool_policy_test.go")
BACKUP = os.path.join(os.path.dirname(__file__), "tool_policy_test.go.orig")


def sha(path):
    return hashlib.sha256(open(path, "rb").read()).hexdigest()


if not os.path.exists(BACKUP):
    raise SystemExit("no backup to restore from")
mutated = sha(TARGET)
original = sha(BACKUP)
shutil.copyfile(BACKUP, TARGET)
restored = sha(TARGET)
os.remove(BACKUP)
print("mutated ", mutated)
print("original", original)
print("restored", restored)
print("RESTORED_MATCHES_ORIGINAL", restored == original)
