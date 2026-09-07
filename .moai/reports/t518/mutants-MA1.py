#!/usr/bin/env python3
"""Mutant runner for SPEC-SPEC-LINT-BLIND-AXES-001 M-A1 (card t518).

Each mutant reverts or corrupts one part of the repair and records which tests
go RED. Mutants that are NOT caught are recorded too — they draw the boundary of
the guard (acceptance.md §A rule 4).

Read-only with respect to the corpus. It edits three source files in place and
restores each from a backup after every run.
"""
import os
import re
import shutil
import subprocess
import sys

# worktree root = four levels up from .moai/reports/t518/<this file>
ROOT = os.path.abspath(os.path.join(os.path.dirname(os.path.abspath(__file__)), "..", "..", ".."))
TABLE = os.path.join(ROOT, "internal/spec/lint_req_table.go")
WIDEN = os.path.join(ROOT, "internal/spec/lint_req_widen.go")
LINT = os.path.join(ROOT, "internal/spec/lint.go")

MUTANTS = [
    ("M1 table collection reverted (parseREQsTable returns nil)", TABLE,
     "\tvar reqs []REQEntry\n\tfor i, line := range strings.Split(body, \"\\n\") {",
     "\tvar reqs []REQEntry\n\tif true {\n\t\treturn nil\n\t}\n\tfor i, line := range strings.Split(body, \"\\n\") {"),

    ("M2 discriminator C-d removed (every ID-leading row admitted)", TABLE,
     "func isTableDefinitionRow(line string) bool {\n\tfor _, marker := range reqDefinitionLexiconL1 {",
     "func isTableDefinitionRow(line string) bool {\n\tif true {\n\t\treturn true\n\t}\n\tfor _, marker := range reqDefinitionLexiconL1 {"),

    ("M3 advisory marking removed from table entries (Widened: false)", TABLE,
     "\t\t\tLine:    i + 1,\n\t\t\tWidened: true,",
     "\t\t\tLine:    i + 1,\n\t\t\tWidened: false,"),

    ("M4 Source wired into reqFindingSeverity (second severity axis)", LINT,
     "func reqFindingSeverity(req REQEntry, base Severity) (Severity, bool) {\n\tif req.Widened {",
     "func reqFindingSeverity(req REQEntry, base Severity) (Severity, bool) {\n\tif req.Source == REQSourceTable {\n\t\treturn base, false\n\t}\n\tif req.Widened {"),

    ("M5 list branch corrupted by one character (bullet class '[-*]' -> '[*]')", WIDEN,
     "`^\\s*[-*]\\s+\\**\\s*(REQ-",
     "`^\\s*[*]\\s+\\**\\s*(REQ-"),

    ("M6 [boundary] L1 matching widened to case-insensitive", TABLE,
     "\t\tif strings.Contains(line, marker) {",
     "\t\tif strings.Contains(strings.ToUpper(line), strings.ToUpper(marker)) {"),

    ("M7 [boundary] row pattern loses bold-marker tolerance around the ID", TABLE,
     "`^\\s*\\|\\s*\\**\\s*(REQ-[A-Z0-9]+(?:-[A-Z0-9]+)*-\\d+)\\s*\\**\\s*\\|`",
     "`^\\s*\\|\\s*(REQ-[A-Z0-9]+(?:-[A-Z0-9]+)*-\\d+)\\s*\\|`"),

    ("M8 [boundary] table body taken from the FIRST cell after the ID", TABLE,
     "\tfor i := len(cells) - 1; i >= 1; i-- {",
     "\tfor i := 1; i < len(cells); i++ {"),
]


def run_tests():
    p = subprocess.run(["go", "test", "./internal/spec/", "-run", "TestTableCollection", "-v"],
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
    print("%-72s %s\n   red: %s\n" % (name, verdict, failed or "-"))

print("=== restore check ===")
rc, failed, bf = run_tests()
print("exit=%d failed=%s\n" % (rc, failed))

print("=== SUMMARY ===")
for name, verdict, failed in results:
    print("%-72s %s" % (name, verdict))
    for f in failed:
        print("      RED: %s" % f)
