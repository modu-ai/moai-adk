#!/usr/bin/env python3
"""Resolution table for SPEC-EVIDENCE-CITATION-CANON-001.

Run from the worktree root:

    git grep -l '\\.moai/state/verify' -- '*.md' > /tmp/citers.txt
    python3 .moai/reports/t375/measure_resolution.py /tmp/citers.txt \\
        > .moai/reports/t375/cited-path-resolution.txt

Emits one row per distinct concrete cited path, recording whether that path
resolves on disk in the primary checkout and in this worktree. It applies the
same `concrete` predicate as measure_citations.py -- a tail beginning with an
ASCII alphanumeric character -- but the two files DEFINE that predicate
separately rather than sharing one definition, so nothing mechanically stops
them drifting apart. Edit both, and re-run both when either changes.

The worktree column is the load-bearing one: a fresh worktree is what any
reader who is not on this machine sees, since `.moai/state/` is gitignored and
therefore travels with no clone.

Population exclusion: this SPEC's own directory is excluded. Without it the
figures move the moment the plan-close commit lands, because `git grep` reads
the tracked set and the SPEC's own artifacts cite the path they are about
(measured: 184 -> 187, 124 -> 127, 231 -> 234). Excluding the SPEC's own
directory is what makes the tracked-set choice hold past its own close.
"""
import io
import os
import re
import sys

SPEC_OWN_DIR = '.moai/specs/SPEC-EVIDENCE-CITATION-CANON-001/'
PATH_RE = re.compile(r'\.moai/state/verify/([^\s`)\]\'",;]*)')
CONCRETE_RE = re.compile(r'^[A-Za-z0-9]')
PRIMARY = '/Users/goos/MoAI/moai-adk-go'


def main(list_path):
    files = [line.strip() for line in open(list_path) if line.strip()]
    files = [f for f in files if not f.startswith(SPEC_OWN_DIR)]
    rows = []
    seen = []

    for path in files:
        text = io.open(path, encoding='utf-8', errors='replace').read()
        for match in PATH_RE.finditer(text):
            tail = match.group(1)
            if not CONCRETE_RE.match(tail):
                continue
            cited = '.moai/state/verify/' + tail.rstrip('/')
            if cited in seen:
                continue
            seen.append(cited)
            rows.append((
                cited,
                path,
                os.path.exists(os.path.join(PRIMARY, cited)),
                os.path.exists(cited),
            ))

    out = sys.stdout
    out.write('cited_path\tfirst_citing_doc\tresolves_in_primary\tresolves_in_worktree\n')
    for cited, path, in_primary, in_worktree in rows:
        out.write('%s\t%s\t%s\t%s\n' % (cited, path, in_primary, in_worktree))
    sys.stderr.write('rows %d | primary %d | worktree %d\n' % (
        len(rows),
        sum(1 for r in rows if r[2]),
        sum(1 for r in rows if r[3]),
    ))


if __name__ == '__main__':
    main(sys.argv[1])
