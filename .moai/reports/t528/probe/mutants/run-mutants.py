#!/usr/bin/env python3
"""t528 M4 — mutant probe driver.

One mutant at a time: inject, VERIFY THE INJECTION LANDED, measure, restore,
verify the restore. Run from the worktree root.

The injection check is the load-bearing part. A mutant whose edit silently
failed to apply produces a green test run that reads exactly like "the guard
did not catch it" — a non-detection that is really a non-injection. This card
has already produced five confident figures from mis-shaped measurement, so the
driver refuses to report a result it did not actually cause: an edit whose old
text is absent is reported as INJECTION-FAILED, never as NOT DETECTED.

The tree is deliberately damaged while this runs, so the pristine copy is taken
once up front and restored after every mutant including on an early exit.
"""

import os
import shutil
import subprocess
import sys
import tempfile

PARSER = "internal/spec/parser.go"
OUT = ".moai/reports/t528/probe/mutants"
CMD = ["go", "test", "./internal/spec/", "-run", "TestT528", "-count=1", "-timeout", "600s"]

# (number, description, [(old, new), ...]) — expectations live in expectations.md
# and were declared before any of this ran.
MUTANTS = [
    (1, "drop the numeric-tail requirement from the id", [
        ("(AC-(?:[A-Za-z0-9]+-)*[0-9]+", "(AC-(?:[A-Za-z0-9]+-)*[A-Za-z0-9]+"),
    ]),
    # Mutants 2, 3 and 5 anchor on strings that ALSO occur in the doc comment
    # above the pattern, so each old-text below carries enough surrounding regex
    # to be unique to the pattern line. The driver's count!=1 check caught this
    # on the first run and reported INJECTION-FAILED rather than a false
    # "not detected".
    (2, "drop em dash U+2014 from the separator set", [
        ("[:—–]\\s*`)", "[:–]\\s*`)"),
    ]),
    (3, "drop the parenthesised-qualifier allowance", [
        (r"\*{0,2}\s*(?:\([^()]*\)\s*)?", r"\*{0,2}\s*"),
    ]),
    (4, "drop the closing-bold allowance (both positions)", [
        (r")\*{0,2}\s*(?:\(", r")\s*(?:\("),
        (r")?\*{0,2}\s*[:—–]", r")?\s*[:—–]"),
    ]),
    (5, "drop the sub-id suffix (.a / .a.i) recognition", [
        (r"(?:\.[a-z](?:\.[a-z]+)?)?)", ")"),
    ]),
    # "return i * 0" rather than "return 0": the loop variable must stay used or
    # the package does not compile, and a build failure is a non-zero exit that
    # would masquerade as a detection.
    (6, "findACSectionStart returns 0 AND the ## break in extractACLines is removed", [
        ("\t\t\treturn i + 1", "\t\t\treturn i * 0"),
        ('\t\tif strings.HasPrefix(trimmed, "##") {', "\t\tif false {"),
    ]),
    (7, "remove ONLY the ## break in extractACLines", [
        ('\t\tif strings.HasPrefix(trimmed, "##") {', "\t\tif false {"),
    ]),
    (8, "relax the AC- prefix to an arbitrary uppercase token", [
        ("^(AC-", "^([A-Z]+-"),
    ]),
]


def sh(args):
    return subprocess.run(args, capture_output=True, text=True)


def main():
    os.makedirs(OUT, exist_ok=True)
    pristine = tempfile.NamedTemporaryFile(delete=False).name
    shutil.copy(PARSER, pristine)
    results = []
    try:
        for num, desc, edits in MUTANTS:
            src = open(pristine, encoding="utf-8").read()
            failed = None
            for old, new in edits:
                if src.count(old) != 1:
                    failed = "old text occurs %d times, expected exactly 1: %r" % (src.count(old), old)
                    break
                src = src.replace(old, new, 1)
            log = os.path.join(OUT, "mutant-%d.txt" % num)

            if failed:
                with open(log, "w", encoding="utf-8") as fh:
                    fh.write("# mutant %d — %s\n\n## INJECTION FAILED\n%s\n" % (num, desc, failed))
                results.append((num, "INJECTION-FAILED", desc))
                print("mutant %d: INJECTION-FAILED — %s" % (num, failed))
                continue

            open(PARSER, "w", encoding="utf-8").write(src)
            diff = sh(["git", "diff", "--", PARSER]).stdout
            if not diff.strip():
                shutil.copy(pristine, PARSER)
                results.append((num, "INJECTION-FAILED", desc))
                print("mutant %d: INJECTION-FAILED — write produced no diff" % num)
                continue

            proc = sh(CMD)
            body = proc.stdout + proc.stderr
            fails = body.count("--- FAIL")
            detected = proc.returncode != 0
            with open(log, "w", encoding="utf-8") as fh:
                fh.write("# mutant %d — %s\n" % (num, desc))
                fh.write("# expectation pre-declared in expectations.md\n\n")
                fh.write("## injected diff\n%s\n" % diff)
                fh.write("## judging command\n%s\n\n" % " ".join(CMD))
                fh.write("## result\n%s\nEXIT=%d\n" % (body, proc.returncode))
                fh.write("DETECTED=%s  --- FAIL lines=%d\n" % (detected, fails))

            results.append((num, "DETECTED" if detected else "NOT-DETECTED", desc))
            print("mutant %d: %-13s (exit=%d, --- FAIL lines=%d)"
                  % (num, "DETECTED" if detected else "NOT DETECTED", proc.returncode, fails))

            shutil.copy(pristine, PARSER)
            if sh(["git", "diff", "--quiet", "--", PARSER]).returncode != 0:
                print("  !! RESTORE FAILED for mutant %d" % num, file=sys.stderr)
                return 9
    finally:
        shutil.copy(pristine, PARSER)
        os.unlink(pristine)

    print("\nrestore check — git diff --stat on the parser must be empty:")
    print(repr(sh(["git", "diff", "--stat", "--", PARSER]).stdout))
    with open(os.path.join(OUT, "summary.txt"), "w", encoding="utf-8") as fh:
        for num, verdict, desc in results:
            fh.write("%d\t%s\t%s\n" % (num, verdict, desc))
    return 0


if __name__ == "__main__":
    sys.exit(main())
