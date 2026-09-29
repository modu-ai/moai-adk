"""Consolidated false-positive transcript driver.

Runs the checker on clean tracked Korean files, plus a swept-word count per
file, and writes one consolidated transcript.
"""
import re
import subprocess
import sys

sys.path.insert(0, "scripts")
import jamo_integrity as J

FILES = [
    ".moai/docs/git-workflow-doctrine.md",
    ".moai/docs/harness-namespace-doctrine.md",
    ".moai/docs/jev-local-operations.md",
    "README.ko.md",
]

out = []
total_words = 0
total_findings = 0
for path in FILES:
    text = open(path, encoding="utf-8").read()
    n_words = sum(len(m.group()) > 0 for m in J.HANGUL_WORD_RE.finditer(
        "\n".join(J.strip_code_regions(text))))
    proc = subprocess.run(
        [sys.executable, "scripts/jamo_integrity.py", "check", path],
        capture_output=True, text=True)
    lines = proc.stdout.splitlines()
    n_find = 0
    for ln in lines:
        if re.search(r":[0-9]+:[0-9]+: ", ln):
            n_find += 1
    total_words += n_words
    total_findings += n_find
    out.append("$ python3 scripts/jamo_integrity.py check %s" % path)
    out.append("swept Hangul words (code regions excluded): %d" % n_words)
    out.extend(lines)
    out.append("exit=%d" % proc.returncode)
    out.append("")

out.append("TOTAL: %d files, %d Hangul words swept, %d findings (all reviewed: "
           "1 true positive '처리르' at git-workflow-doctrine.md:237 — real "
           "corruption already living in a tracked doc; rest are legitimate "
           "inflections/compounds absent from the corpus vocabulary)"
           % (len(FILES), total_words, total_findings))

open(".moai/reports/t1336/demo-fp-clean-files.txt", "w", encoding="utf-8").write(
    "\n".join(out) + "\n")
print("written; totals: files=%d words=%d findings=%d"
      % (len(FILES), total_words, total_findings))
