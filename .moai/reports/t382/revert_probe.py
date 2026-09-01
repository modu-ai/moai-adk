"""t382 necessity probe — remove ONLY the fixtures[2] phase override from audit_test.go.

Answers the question the criterion amendment turns on: is the fourth changed file
NECESSARY (an existing test asserts the pre-fix behaviour without it) or merely
CONVENIENT? Run the package test after this and observe.

Restore from the backup afterwards; this script does not restore.
"""
import sys

path = "internal/spec/audit_test.go"
target = '\tfixtures[2].specMD = strings.Replace(fixtures[2].specMD, `phase: "v3.0.0"`, `phase: "legacy"`, 1)\n'

with open(path, encoding="utf-8") as fh:
    text = fh.read()

if target not in text:
    print("FAIL: override line not found — nothing reverted")
    sys.exit(2)

with open(path, "w", encoding="utf-8") as fh:
    fh.write(text.replace(target, "", 1))
print("reverted: fixtures[2] phase override removed (1 line)")
