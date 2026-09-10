#!/usr/bin/env python3
"""Mutant runner for SPEC-SPEC-LINT-BLIND-AXES-001 M-A2 (card t518), axis 2 branch B.

Each mutant reverts or corrupts one part of the repair and records which tests go
RED. Mutants that are NOT caught are recorded too — they draw the boundary of the
guard (acceptance.md §A rule 4).

The whole internal/spec package is run, not only the new tests: a mutant caught
ONLY by a pre-existing test is a different fact from one caught by this
milestone's guards, and the distinction is invisible under a -run filter.

Read-only with respect to the corpus. It edits one source file in place and
restores it from a backup after every run.
"""
import os
import re
import shutil
import subprocess
import sys

# worktree root = three levels up from .moai/reports/t518/<this file>
ROOT = os.path.abspath(os.path.join(os.path.dirname(os.path.abspath(__file__)), "..", "..", ".."))
LINT = os.path.join(ROOT, "internal/spec/lint.go")

MUTANTS = [
    ("M1 unjudged emission removed (the announcement is deleted)", LINT,
     "\t\tcase modalityUnjudged:\n",
     "\t\tcase modalityUnjudged:\n\t\t\tif true {\n\t\t\t\tbreak\n\t\t\t}\n"),

    ("M2 SHALL contact reverted to leading-space (strings.Contains(upper, \" SHALL\"))", LINT,
     "\thasShall := shallWordPattern.MatchString(upper)",
     "\thasShall := strings.Contains(upper, \" SHALL\")"),

    ("M3 unjudged verdict routed into conforming (the third state is erased)", LINT,
     "\tif hasShall {\n\t\treturn modalityJudgedConforming\n\t}\n\treturn modalityUnjudged\n}",
     "\tif hasShall {\n\t\treturn modalityJudgedConforming\n\t}\n\treturn modalityJudgedConforming\n}"),

    ("M4 unjudged finding loses its Advisory marking", LINT,
     "\t\t\t\tAdvisory: true, // reports, never gates — see the block above",
     "\t\t\t\tAdvisory: false, // reports, never gates — see the block above"),

    ("M5 unjudged finding promoted to error severity", LINT,
     "\t\t\t\tSeverity: SeverityWarning,\n\t\t\t\tAdvisory: true, // reports, never gates",
     "\t\t\t\tSeverity: SeverityError,\n\t\t\t\tAdvisory: true, // reports, never gates"),

    ("M6 [boundary] SHALL word boundary widened to a bare substring (SHALLOW counts)", LINT,
     "regexp.MustCompile(`\\bSHALL\\b`)",
     "regexp.MustCompile(`SHALL`)"),

    ("M7 [boundary] the SHALL-token route to judgeability removed (prefix only)", LINT,
     "\tif hasShall {\n\t\treturn modalityJudgedConforming\n\t}\n\treturn modalityUnjudged\n}",
     "\treturn modalityUnjudged\n}"),

    ("M8 [boundary] one English prefix dropped from the list (\"THE \")", LINT,
     'var modalityPrefixes = []string{"WHEN ", "WHILE ", "WHERE ", "IF ", "THE "}',
     'var modalityPrefixes = []string{"WHEN ", "WHILE ", "WHERE ", "IF "}'),

    ("M9 [boundary] prefix match made case-sensitive against the RAW text", LINT,
     "\t\tif strings.HasPrefix(upper, prefix) {",
     "\t\tif strings.HasPrefix(text, prefix) {"),
]


def run_tests():
    p = subprocess.run(["go", "test", "./internal/spec/", "-count=1", "-v"],
                       cwd=ROOT, capture_output=True, text=True)
    failed = sorted(set(re.findall(r"--- FAIL: (\S+)", p.stdout)))
    build_failed = "build failed" in p.stdout or "build failed" in p.stderr
    return p.returncode, failed, build_failed


print("=== baseline (unmutated) ===")
rc, failed, bf = run_tests()
print("exit=%d failed=%s build_failed=%s\n" % (rc, failed, bf))
if rc != 0:
    sys.exit("baseline is not green; aborting")

results = []
for name, path, old, new in MUTANTS:
    bak = path + ".t518bak"
    shutil.copyfile(path, bak)
    src = open(path, encoding="utf-8").read()
    if old not in src:
        os.remove(bak)
        results.append((name, "APPLY-FAILED", []))
        print("!! %s -- anchor not found\n" % name)
        continue
    open(path, "w", encoding="utf-8").write(src.replace(old, new, 1))
    rc, failed, bf = run_tests()
    shutil.move(bak, path)
    verdict = "CAUGHT" if rc != 0 else "NOT CAUGHT"
    if bf:
        verdict += " (build failed)"
    results.append((name, verdict, failed))
    print("%-78s %s\n   red: %s\n" % (name, verdict, failed or "-"))

print("=== restore check ===")
rc, failed, bf = run_tests()
print("exit=%d failed=%s\n" % (rc, failed))

print("=== SUMMARY ===")
for name, verdict, failed in results:
    print("%-78s %s" % (name, verdict))
    for f in failed:
        print("      RED: %s" % f)
