#!/usr/bin/env python3
"""Canonical measurement for SPEC-EVIDENCE-CITATION-CANON-001 corpus figures.

Run from the worktree root:

    git grep -l '\\.moai/state/verify' -- '*.md' > /tmp/citers.txt
    python3 .moai/reports/t375/measure_citations.py /tmp/citers.txt

Population: the git-TRACKED `*.md` set at the current HEAD. `git grep` is used
rather than `grep -r` on purpose -- `grep -r` also walks untracked files, which
includes this SPEC's own artifacts, so its count moves as the card is written.

"Concrete" citation: a `.moai/state/verify/<tail>` occurrence whose <tail>
begins with an ASCII alphanumeric character. Every placeholder form observed in
this corpus opens with a non-alphanumeric character -- `<session>`, `{session}`,
`$MOAI_SESSION_ID` -- so the predicate is stated as a positive character-class
test rather than as a list of placeholder spellings to exclude. That choice is
load-bearing: an exclusion list silently miscounts any placeholder spelling it
forgot, and three earlier hand-run variants of this measurement disagreed
(129 / 131 / 132) for exactly that reason.

Population exclusion: this card's own output directories are excluded. Without it the
figures move the moment the plan-close commit lands, because `git grep` reads
the tracked set and the SPEC's own artifacts cite the path they are about
(measured: 184 -> 187, 124 -> 127, 231 -> 234). Excluding the SPEC's own
directory is what makes the tracked-set choice hold past its own close.
"""
import io
import re
import sys

# Both of this card's own output directories. Excluding only the SPEC
# directory was not enough: the plan-close commit tracked the audit reports
# too, and those cite the path they are about, so the figures moved anyway
# (184 -> 185, 124 -> 125, 231 -> 232 -- measured after that commit).
OWN_DIRS = (
    '.moai/specs/SPEC-EVIDENCE-CITATION-CANON-001/',
    '.moai/reports/t375/',
)
PATH_RE = re.compile(r'\.moai/state/verify/([^\s`)\]\'",;]*)')
CONCRETE_RE = re.compile(r'^[A-Za-z0-9]')


def main(list_path):
    files = [line.strip() for line in open(list_path) if line.strip()]
    files = [f for f in files if not f.startswith(OWN_DIRS)]
    occurrences = 0
    concrete_occurrences = 0
    concrete_files = []
    concrete_paths = []

    for path in files:
        text = io.open(path, encoding='utf-8', errors='replace').read()
        for match in PATH_RE.finditer(text):
            occurrences += 1
            tail = match.group(1)
            if not CONCRETE_RE.match(tail):
                continue
            concrete_occurrences += 1
            if path not in concrete_files:
                concrete_files.append(path)
            cited = '.moai/state/verify/' + tail.rstrip('/')
            if cited not in concrete_paths:
                concrete_paths.append(cited)

    print('files scanned (tracked *.md matching the string):', len(files))
    print('path occurrences:', occurrences)
    print('concrete occurrences:', concrete_occurrences)
    print('files carrying >=1 concrete citation:', len(concrete_files))
    print('distinct concrete cited paths:', len(concrete_paths))


if __name__ == '__main__':
    main(sys.argv[1])
