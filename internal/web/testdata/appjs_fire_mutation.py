#!/usr/bin/env python3
"""app.js historical-defect re-introducer (SPEC-APPJS-FIRE-GUARD-001, card t1060).

Mechanizes the plan-phase reverse prototype: moves the
`htmx:afterSettle -> stampRefreshed` registration from its own top-level IIFE
into the FIRST IIFE, right after the initConsole binding — the historical
defect shape that kills every top-level effect after the insertion point
(ReferenceError: stampRefreshed is not defined, at load).

PRECONDITION ASSERTION (REQ-AFG-008): this tool's silence would be the red
phase's poison — a mutation that never happens makes the expected exit-1
probe run come out 0 and the guard reads "0 violations". So every failure to
synthesize the mutation is a NON-ZERO exit with a named reason:

  - registration line not found (app.js evolved past the anchor) -> exit 1
  - registration line found more than once (ambiguous)           -> exit 1
  - initConsole anchor not found / ambiguous                     -> exit 1
  - target already looks mutated (registration already adjacent
    to the anchor)                                               -> exit 1

Matching is by CONTENT PATTERN, never by line-number constant (line numbers
drift as app.js evolves; the plan forbids pinning coordinates). A pristine
byte copy of the target is written BEFORE any mutation; restoring is the
caller's `cp` + `cmp` + `git status` duty (REQ-AFG-009).

Usage:
  appjs_fire_mutation.py [--target internal/web/assets/app.js] [--pristine PATH]
"""

import argparse
import shutil
import sys

REGISTRATION = 'document.addEventListener("htmx:afterSettle", stampRefreshed);'
ANCHOR = 'document.addEventListener("htmx:afterSettle", initConsole);'


def fail(message):
    print("MUTATION FAILED: " + message)
    return 1


def main():
    ap = argparse.ArgumentParser(description="re-introduce the historical IIFE-crossing registration defect")
    ap.add_argument("--target", default="internal/web/assets/app.js", help="app.js path (default: internal/web/assets/app.js)")
    ap.add_argument("--pristine", default=None, help="where to snapshot the pristine copy (default: <target>.pristine)")
    args = ap.parse_args()
    pristine = args.pristine or (args.target + ".pristine")

    try:
        with open(args.target, encoding="utf-8") as f:
            src = f.read()
    except OSError as exc:
        return fail("cannot read target %s: %s" % (args.target, exc))

    lines = src.splitlines(keepends=True)
    matches = [i for i, line in enumerate(lines) if REGISTRATION in line]
    if len(matches) == 0:
        return fail(
            "registration pattern not found in %s: %r — app.js evolved past this mutation anchor; "
            "update REGISTRATION here before trusting a green red-phase" % (args.target, REGISTRATION)
        )
    if len(matches) > 1:
        return fail("registration pattern ambiguous: %d matches in %s" % (len(matches), args.target))
    anchors = [i for i, line in enumerate(lines) if ANCHOR in line]
    if len(anchors) == 0:
        return fail("initConsole anchor not found: %r" % ANCHOR)
    if len(anchors) > 1:
        return fail("initConsole anchor ambiguous: %d matches" % len(anchors))
    if abs(matches[0] - anchors[0]) <= 1:
        return fail(
            "target already looks mutated (registration adjacent to the initConsole anchor) — "
            "restore the pristine copy before mutating"
        )

    # Snapshot BEFORE touching anything (REQ-AFG-009's restore depends on it).
    shutil.copyfile(args.target, pristine)

    removed_1based = matches[0] + 1
    reg_line = lines.pop(matches[0])
    if matches[0] < anchors[0]:
        anchor_index = anchors[0] - 1  # shift after the removal
    else:
        anchor_index = anchors[0]
    lines.insert(anchor_index + 1, reg_line)
    inserted_1based = anchor_index + 2  # 1-based line number of the inserted line

    with open(args.target, "w", encoding="utf-8") as f:
        f.write("".join(lines))

    print("MUTATED removed@%d inserted@%d pristine=%s" % (removed_1based, inserted_1based, pristine))
    return 0


if __name__ == "__main__":
    sys.exit(main())
