#!/usr/bin/env python3
"""M4 mutant runner (card t607, SPEC-RESOURCE-SLOT-LEASE-001).

For each mutant: copy the guard file to a scratch backup, apply ONE textual
substitution (asserted to match exactly once), run the named test, restore the
file by copying the backup back, and prove the restore with a byte comparison.
Strictly sequential: one `go test` at a time. The committed-version check
(`git diff --quiet -- <file>`) is run separately after this script.

Usage: python3 m4_mutants.py <scratch-dir>
"""
import filecmp
import os
import shutil
import subprocess
import sys

HERE = os.path.dirname(os.path.abspath(__file__))
ROOT = os.path.abspath(os.path.join(HERE, "..", "..", "..", ".."))
GUARD = "internal/hook/slot_lease_guard.go"
OUT = os.path.join(HERE, "mutants")

MUTANTS = [
    ("n-enabled", "settings check deleted",
     "\tif !cfg.Enabled || input == nil || len(input.ToolInput) == 0 {",
     "\tif input == nil || len(input.ToolInput) == 0 {",
     "^TestSlotLeaseGuard_DenyMatrix$"),
    ("n-match", "pattern check deleted",
     "\t\t\tif re.MatchString(scrubbed) {",
     "\t\t\tif true || re.MatchString(scrubbed) {",
     "^TestSlotLeaseGuard_DenyMatrix$"),
    ("n-quoted", "quote scrub deleted",
     "\tscrubbed := substituteQuotedArguments(command)",
     "\tscrubbed := command",
     "^TestSlotLeaseGuard_DenyMatrix$"),
    ("n-self", "session comparison deleted",
     "\t\tcase lease.SessionID == input.SessionID:\n\t\t\tentry.Event = \"allow-self\"\n",
     "",
     "^TestSlotLeaseGuard_DenyMatrix$"),
    ("n-alive", "liveness check deleted",
     "\t\tcase lease.Stale():\n\t\t\tentry.Event = \"allow-stale\"\n",
     "",
     "^TestSlotLeaseGuard_DenyMatrix$"),
    # Disabled rather than deleted: deleting the case leaves `now` unused and
    # the package no longer compiles, which measures nothing (first run).
    ("n-bound", "expiry check disabled",
     "\t\tcase lease.Expired(now):",
     "\t\tcase false && lease.Expired(now):",
     "^TestSlotLeaseGuard_DenyMatrix$"),
    ("n-held", "holder-present check deleted",
     "\t\tcase !lease.Held():\n\t\t\tentry.Event = \"allow-unheld\"\n",
     "",
     "^TestSlotLeaseGuard_DenyMatrix$"),
    ("multi", "only the first attributed resource is judged",
     "\tfor _, name := range matched {",
     "\tfor _, name := range matched[:1] {",
     "^TestSlotLeaseGuard_DenyMatrix$"),
    ("ac016-no-normalize", "root normalization removed (records read from the hook root)",
     "\troot, err := kanban.ResolveSlotLeaseRoot(hookRoot)",
     "\troot, err := hookRoot, error(nil)",
     "^TestSlotLeaseGuard_NormalizesWorktreeRootToPrimary$"),
    ("ac016-fallthrough", "normalization failure falls through to the unnormalized root",
     "\tif err != nil {\n\t\tadvise(\"cannot normalize %s to the shared root (%v)\", hookRoot, err)\n"
     "\t\tauditSlotGuard(hookRoot, kanban.SlotLeaseAuditEntry{Event: \"fail-open\", Reason: \"root-unresolved: \" + err.Error(), SessionID: input.SessionID})\n"
     "\t\treturn \"\", \"\"\n\t}\n",
     "\tif err != nil {\n\t\troot = hookRoot\n\t}\n",
     "^TestSlotLeaseGuard_NormalizesWorktreeRootToPrimary$"),
]


def main():
    scratch = sys.argv[1]
    only = set(sys.argv[2:])
    os.makedirs(OUT, exist_ok=True)
    target = os.path.join(ROOT, GUARD)
    backup = os.path.join(scratch, "slot_lease_guard.go.bak")
    shutil.copyfile(target, backup)
    original = open(backup).read()
    summary = []
    for name, desc, old, new, test in MUTANTS:
        if only and name not in only:
            continue
        if original.count(old) != 1:
            print(f"{name}: anchor found {original.count(old)} times; aborting", file=sys.stderr)
            return 2
        with open(target, "w") as f:
            f.write(original.replace(old, new, 1))
        proc = subprocess.run(
            ["go", "test", "./internal/hook/", "-run", test, "-count=1", "-v"],
            cwd=ROOT, capture_output=True, text=True)
        with open(os.path.join(OUT, f"{name}.txt"), "w") as f:
            f.write(proc.stdout + proc.stderr)
        shutil.copyfile(backup, target)
        same = filecmp.cmp(backup, target, shallow=False)
        fails = [ln.strip() for ln in proc.stdout.splitlines() if ln.strip().startswith("--- FAIL")]
        summary.append(f"{name}\t{desc}\texit={proc.returncode}\trestored_cmp_equal={same}\tfails={fails}")
        if not same:
            print("RESTORE FAILED", file=sys.stderr)
            return 3
    with open(os.path.join(OUT, "summary.tsv"), "a") as f:
        f.write("\n".join(summary) + "\n")
    print("\n".join(summary))
    return 0


if __name__ == "__main__":
    sys.exit(main())
