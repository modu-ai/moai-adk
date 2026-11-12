#!/usr/bin/env python3
"""Classify a change set as docs-only or full scope (#2527, design comment 6822993459).

Reads changed paths, one per line, from stdin. Emits JSON on stdout:

    {"scope": "docs-only" | "full", "reasons": [...]}

Always exits 0. Any internal error fails safe to scope "full" with a reason,
so a broken classifier can never skip a heavy CI step (#2527).
"""

import json
import sys

# Rule table per design comment 6822993459 on #2527 (fail-safe default: full).
# Tests read these exact files, so renaming or editing them is not docs-only.
EXACT_DENY = frozenset(
    {
        # Collected mechanically: every docs/ path that a test, tool or src file
        # opens or compares (#2538 B2). Source: grep for the quoted path across
        # tests/, tools/ and src/; DocumentationDriftTest below re-runs that scan
        # so a newly referenced doc fails the suite until it is added here.
        "docs/decisions-2026-11-08.md",
        "docs/host-support-matrix.md",
        "docs/zcode-support.md",
        "docs/codex-adapter-contract.md",
        "docs/design-review/appendix/interfaces.md",
        "docs/design-review/appendix/state-vocabulary-crosswalk.md",
        "docs/design-review/autonomy-modes-spec.md",
        "docs/release-attestation.md",
        "docs/2383-statusline-switch-design.md",
        "docs/rust-product-path.md",
        "docs/llms.txt",
        "docs/user-entry-draft.md",
        "docs/product-naming-inventory.md",
        "docs/pr-contract.md",
        "docs/support-matrix.md",
        "docs/release-checklist.md",
        "docs/release-review.md",
        "docs/design-review/applied.md",
        "docs/design-review/pages.tsv",
    }
)
# Directory prefixes that are never docs-only even under docs/.
# docs/design-review/fixes/: every fix file is hashed by tests/space_applied_record_test.py (#2207).
PREFIX_DENY = ("docs/e2e-receipts/", "docs/design-review/fixes/")
# Allowed suffixes under docs/ (anything else under docs/ is full scope).
DOC_SUFFIXES = (".md", ".txt", ".png", ".svg")
# Report output is safe regardless of suffix.
REPORTS_PREFIX = ".moai-cli/reports/"


def classify(paths):
    """Return (scope, reasons) for an iterable of changed path strings."""
    # The workflow pipes `git diff --name-only` output; an empty diff arrives as
    # one empty line, which must fall back to full exactly like empty stdin (#2527).
    paths = [path for path in paths if path]
    reasons = []
    for path in paths:
        if not path:
            continue
        if path.startswith(REPORTS_PREFIX):
            continue
        if not path.startswith("docs/"):
            reasons.append(f"outside docs/: {path}")
            continue
        if path in EXACT_DENY:
            reasons.append(f"exact-denied (read by tests): {path}")
            continue
        if path.startswith(PREFIX_DENY):
            reasons.append(f"prefix-denied: {path}")
            continue
        if not path.endswith(DOC_SUFFIXES):
            reasons.append(f"docs path with non-doc suffix: {path}")
            continue
    if not reasons and paths:
        return "docs-only", []
    if not paths:
        reasons.append("no changed paths (empty diff input)")
    return "full", reasons


def main():
    try:
        paths = [line.rstrip("\n") for line in sys.stdin]
        scope, reasons = classify(paths)
    except Exception as exc:  # fail safe: any error means full scope
        scope, reasons = "full", [f"classifier internal error: {exc}"]
    json.dump({"scope": scope, "reasons": reasons}, sys.stdout)
    sys.stdout.write("\n")
    return 0


if __name__ == "__main__":
    sys.exit(main())
