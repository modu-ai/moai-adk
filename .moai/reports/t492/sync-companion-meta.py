#!/usr/bin/env python3
"""t492 C2 — apply the three metadata updates to a companion file (live or template mirror).

Literal, non-regex replacements. Exits non-zero if any expected text is absent, so a silent
no-op cannot masquerade as success.
"""
import sys

PAIRS = [
    (
        'description: "Detail companion for verification-claim-integrity.md — per-section '
        'elaboration of the 5-section report format, the cross-reference table, and the two '
        'worked-example incident records"',
        'description: "Detail companion for verification-claim-integrity.md — per-section '
        'elaboration of the 5-section report format, the cross-reference table, the two '
        'worked-example incident records, and the §2.1 moving-ref predicate procedure '
        '(four tests, grounded instances, detection limits)"',
    ),
    (
        "> This file owns what each section contains, the cross-reference table, and the two incident\n"
        "> records the doctrine was written from. Load it when composing an evidence-bearing report for the\n"
        "> first time, or when tracing a clause back to the failure that produced it.",
        "> This file owns what each section contains, the cross-reference table, the two incident\n"
        "> records the doctrine was written from, and the §2.1 moving-ref predicate procedure. Load it when\n"
        "> composing an evidence-bearing report for the first time, when applying the moving-ref predicate,\n"
        "> or when tracing a clause back to the failure that produced it.",
    ),
    (
        "Classification: Lazy companion — rationale, elaboration, cross-references, and incident records\n"
        "only. Every obligation stays in `verification-claim-integrity.md`.",
        "Classification: Lazy companion — rationale, elaboration, cross-references, incident records, and\n"
        "the §2.1 predicate procedure. Every obligation stays in `verification-claim-integrity.md`: the\n"
        "predicate procedure here is HOW a class is reached, never WHETHER it must be applied.",
    ),
]

path = sys.argv[1]
with open(path, encoding="utf-8") as fh:
    body = fh.read()

missing = [old.splitlines()[0][:60] for old, _ in PAIRS if old not in body]
if missing:
    print("ABSENT (no replacement made): " + " | ".join(missing), file=sys.stderr)
    sys.exit(1)

for old, new in PAIRS:
    body = body.replace(old, new, 1)

with open(path, "w", encoding="utf-8") as fh:
    fh.write(body)
print("OK 3/3 replaced: " + path)
