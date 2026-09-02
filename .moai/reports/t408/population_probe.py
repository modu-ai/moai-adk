"""t408 probe: which progress.md files are invisible to the era parser?

era.go classifies by string-matching the headings §E.2 / §E.3 / §E.4 / §E.5.
A progress.md that letters its phase signals under §F.N instead carries none of
them, so ClassifyEra falls to H-2 ("progress.md without §E.* markers") and a
modern SPEC reads as V3R2-R4.

This probe measures that population directly, so the guard's pinned list is a
measurement rather than a recollection.
"""

import glob
import io
import os
import re

MARKERS = ("§E.2", "§E.3", "§E.4", "§E.5")
HEADING = re.compile(r"^#{2,6}\s*(§[A-Z](?:\.[0-9.]+)?)[^\n]*", re.M)
PHASE_F = re.compile(r"^#{2,6}\s*§F\.[0-9][^\n]*?(Plan-phase|Run-phase|Sync-phase|Mx-phase)", re.M)


def has_marker(text):
    for m in MARKERS:
        for prefix in ("## ", "### ", "#### ", "##### "):
            if prefix + m in text:
                return True
    return False


def main():
    invisible, f_phase = [], []
    for p in sorted(glob.glob(".moai/specs/*/progress.md")):
        spec = os.path.basename(os.path.dirname(p))
        text = io.open(p, encoding="utf-8").read()
        vis = has_marker(text)
        fp = bool(PHASE_F.search(text))
        if not vis:
            invisible.append((spec, fp))
        if fp:
            f_phase.append((spec, vis))

    print("progress.md files scanned: %d" % len(glob.glob(".moai/specs/*/progress.md")))
    print()
    print("era-invisible (no §E.{2,3,4,5} heading): %d" % len(invisible))
    for s, fp in invisible:
        print("   %-46s letters phase signals under §F.N: %s" % (s, fp))
    print()
    print("letters a phase signal under §F.N: %d" % len(f_phase))
    for s, vis in f_phase:
        print("   %-46s also carries an §E.* marker: %s" % (s, vis))


if __name__ == "__main__":
    main()
