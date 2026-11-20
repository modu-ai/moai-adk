#!/usr/bin/env python3
"""Classify a change set as docs-only or full scope (#2527, design comment 6822993459),
and decide whether a PR needs the macOS target (C-76, SPEC-20261119-macos-on-pr).

Reads changed paths, one per line, from stdin. Emits JSON on stdout:

    {"scope": "docs-only" | "full", "reasons": [...],
     "macos_needed": true | false, "macos_reasons": [...]}

Always exits 0. Any internal error fails safe to scope "full" and
"macos_needed": true with a reason, so a broken classifier can never skip a
heavy CI step (#2527) or narrow the matrix to linux-only (C-76).
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
        "docs/design-review/appendix/ledger_reservation.md",
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
        "docs/release-rehearsal-2026-11-12.md",
        "docs/lane-loop-contract.md",
        "docs/protocol-v1.md",
        "docs/message-test-isolation.md",
        "docs/e2e-matrix.md",
    }
)
# Directory prefixes that are never docs-only even under docs/.
# docs/design-review/fixes/: every fix file is hashed by tests/space_applied_record_test.py (#2207).
# docs/design-review/appendix/operations/: one fragment per one-file operation, read by the
# operations_doc test in src/main.rs and tests/hot_file_ops_merge.rs (#2617).
PREFIX_DENY = (
    "docs/e2e-receipts/",
    "docs/design-review/fixes/",
    "docs/design-review/appendix/operations/",
)
# Allowed suffixes under docs/ (anything else under docs/ is full scope).
DOC_SUFFIXES = (".md", ".txt", ".png", ".svg")
# Report output is safe regardless of suffix.
REPORTS_PREFIX = ".moai-cli/reports/"

# C-76 (SPEC-20261119-macos-on-pr): platform-sensitive paths. A PR touching one
# of these runs the macos-arm64 target of the `check` job before merge; every
# other PR keeps the linux-x86_64-only matrix of #2554 (U10). The six categories
# were measured from the 47 macOS-only post-merge failures
# (.moai/cards-pending/pr-history-analysis-20261117.md, read 2026-11-17); the
# rules were verified against the merged file lists of all 69 failed ci push
# runs on v2026.11.0 (read 2026-11-19) — tests/ci_scope_test.py embeds those
# lists and names the 40 runs that stay linux-only.
PLATFORM_EXACT = {
    "src/launcher_terminal.rs": "pty/terminal",
    "tests/help_layout.rs": "pty/terminal",
    "src/heavy_slot.rs": "process/signal",
    "tools/heavy_check_slot.py": "process/signal",
    "tests/heavy_slot_cli.rs": "process/signal",
    "tests/heavy_check_slot_test.py": "process/signal",
    "src/hook_engine.rs": "process/signal",
    "src/worktree_gc.rs": "fs links/permissions",
    "tests/exe_link_cli.rs": "fs links/permissions",
    "src/go_state.rs": "fs links/permissions",
    "src/state_root.rs": "fs links/permissions",
    "src/update.rs": "fs links/permissions",
    "src/self_update.rs": "fs links/permissions",
    "src/versioning.rs": "fs links/permissions",
    "src/headver.rs": "fs links/permissions",
    "src/agent_install.rs": "fs links/permissions",
    "tests/home_isolation_scan.rs": "fs links/permissions",
    "tools/home_entry_count.py": "fs links/permissions",
    "src/launcher_init.rs": "launcher",
    "build.rs": "build/toolchain",
    "Cargo.toml": "build/toolchain",
    "Cargo.lock": "build/toolchain",
    "rust-toolchain.toml": "build/toolchain",
    "HEADVER": "build/toolchain",
    "dist-workspace.toml": "build/toolchain",
    "tools/local_check.sh": "build/toolchain",
}
PLATFORM_PREFIX = (
    ("src/statusline/", "statusline"),
    ("tests/statusline", "statusline"),
    ("tests/fixtures/statusline", "statusline"),
    ("src/launcher_", "launcher"),
    ("tests/launcher_", "launcher"),
    ("tests/worktree_gc", "fs links/permissions"),
    ("src/home_", "fs links/permissions"),
    ("tests/update_", "fs links/permissions"),
    ("src/ops/agents_home_", "fs links/permissions"),
    ("tests/agents_home", "fs links/permissions"),
    ("tests/home_isolation", "fs links/permissions"),
    ("tests/home_entry_count", "fs links/permissions"),
    ("tools/home_entry_count", "fs links/permissions"),
    (".cargo/", "build/toolchain"),
    (".github/workflows/", "build/toolchain"),
    ("tools/ci-", "build/toolchain"),
    ("tools/ci_", "build/toolchain"),
    ("tests/ci_", "build/toolchain"),
    ("install/", "build/toolchain"),
)
PLATFORM_SUFFIX = (
    ("_tty_test.py", "pty/terminal"),
)


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


def macos_needed(paths):
    """Return (needed, reasons) for an iterable of changed path strings (C-76).

    Fail safe: no usable paths, or a path that cannot be judged (control
    characters), answers needed=True — the wider matrix, never silently
    linux-only.
    """
    usable = [path for path in paths if path]
    if not usable:
        return True, ["no changed paths (empty diff input)"]
    reasons = []
    for path in usable:
        if any(ord(ch) < 32 or ord(ch) == 127 for ch in path):
            return True, ["a changed path is not readable text; failing safe"]
        category = PLATFORM_EXACT.get(path)
        if category is not None:
            reasons.append(f"platform-sensitive [{category}]: {path}")
            continue
        for prefix, cat in PLATFORM_PREFIX:
            if path.startswith(prefix):
                reasons.append(f"platform-sensitive [{cat}]: {path}")
                break
        else:
            for suffix, cat in PLATFORM_SUFFIX:
                if path.startswith("tests/") and path.endswith(suffix):
                    reasons.append(f"platform-sensitive [{cat}]: {path}")
                    break
    return bool(reasons), reasons


def main():
    try:
        paths = [line.rstrip("\n") for line in sys.stdin]
        scope, reasons = classify(paths)
    except Exception as exc:  # fail safe: any error means full scope
        scope, reasons = "full", [f"classifier internal error: {exc}"]
    try:
        needed, macos_reasons = macos_needed(paths)
    except Exception as exc:  # fail safe: any error keeps the wider matrix
        needed, macos_reasons = True, [f"platform classifier internal error: {exc}"]
    json.dump(
        {
            "scope": scope,
            "reasons": reasons,
            "macos_needed": needed,
            "macos_reasons": macos_reasons,
        },
        sys.stdout,
    )
    sys.stdout.write("\n")
    return 0


if __name__ == "__main__":
    sys.exit(main())
