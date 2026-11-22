#!/usr/bin/env python3
"""Inventory of every Codex and OMP mention in the tracked files of one commit, grouped by area.

usage: python3 inventory.py <commit> [--out inventory.tsv]

Reads only the commit's own tree (`git grep` on the commit, so a working copy that already moved on
cannot change the answer). One row per file: area, path, number of mentioning lines, and the line
numbers (`path:line` is every mention). A line counts when it holds `codex` (any case) or the word
`omp` (any case). The summary on stdout is the per-area count the SPEC quotes; `--check N` exits 1
when the number of files differs from N, so the SPEC's count can be verified at the commit.
"""
import argparse
import re
import subprocess
import sys
from collections import defaultdict

# First matching rule wins. Each rule: (area, regex on the path).
AREAS = [
    # Work to restore later: the SPECs and the cards' own text stay, withdrawn, not deleted.
    ("keep-restore-record", r"^\.moai/specs/"),
    ("keep-restore-record", r"^\.moai-cli/reports/"),
    # Generated or per-project state that names Codex: regenerated or edited, not kept as history.
    ("project-state-files", r"^\.moai(-cli)?/(manifests/|project\.yaml)"),
    # Files whose whole purpose is Codex: removed outright.
    ("codex-only-files", r"(^|/)[^/]*codex[^/]*$"),
    ("codex-only-files", r"^\.codex/"),
    ("codex-only-files", r"^tests/fixtures/hooks-pre-p4/codex"),
    # Product code, by what it does.
    ("launcher-and-session", r"^src/(launcher[^/]*|session[^/]*|ops/session[^/]*)\.rs$"),
    ("hooks", r"^src/(hooks|hook_[a-z_]+|update_hooks)\.rs$"),
    ("hooks", r"^src/ops/[a-z_]*hook[a-z_]*\.rs$"),
    ("statusline", r"^src/(statusline/|status_line|progress)"),
    ("lane-and-card-tooling", r"^src/(lane_[a-z_]+|worktree_gc|leader_[a-z_]+|decision[a-z_]*|goal_block|scope|trailers|review|audit|contract|gate|labels_check|card/|ops/(lane_|leader_|card|goal_block|trailers|ledger_))"),
    ("home-layer", r"^src/(host_home|home_[a-z_]+|update[a-z_]*|user_assets|skills_legacy|migrate|shim_check|claude|context/|ops/(home_|agents_home|skills_legacy|update))"),
    ("project-layer", r"^src/(project_[a-z_]+|generation|agents|agent_[a-z_]+|go_state|profile|model_effort[a-z_]*|decision|ui_[a-z_]+|ops/project)"),
    ("cli-surface", r"^src/(main|help)\.rs$"),
    ("other-src", r"^src/"),
    # Everything that is not product code.
    ("tests-and-fixtures", r"^tests/"),
    ("skills-and-agent-text", r"^(skills|agents|templates)/"),
    ("docs", r"^(docs/|README|SUPPORT|SECURITY|CONTRIBUTING|AGENTS)"),
    ("install-ci-tools", r"^(install/|\.github/|scripts/|tools/|schemas/|experiments/|Cargo\.toml)"),
    ("project-state-files", r"^\.moai/"),
]
AREA_ORDER = list(dict.fromkeys(a for a, _ in AREAS))


def area_of(path):
    for name, pattern in AREAS:
        if re.search(pattern, path):
            return name
    return "unassigned"


# Which removal slice edits the file first. Tests travel with the code they exercise, so a test is
# given the slice of the first word of its name that names an area; a file nothing names goes to
# slice 3 (text, docs, scans). `src/main.rs` and `src/help.rs` hold the commands of slices 1 and 2:
# slice 1 owns them first and slice 2 takes them after slice 1 has landed ("1>2").
SLICE_OF_AREA = {
    "keep-restore-record": "keep",
    # `1c` = what the launcher slice left of OMP (`moai omp`, the host word) and the install texts;
    # `1` = the home update writers (`~/.codex`, `~/.agents/skills`, OMP agents).
    "launcher-and-session": "1c",
    "home-layer": "1",
    "hooks": "2",
    "project-layer": "2",
    "statusline": "2",
    "codex-only-files": "2",
    "other-src": "2",
    "lane-and-card-tooling": "4",
    "cli-surface": "1>2",
    "skills-and-agent-text": "3",
    "docs": "3",
    "install-ci-tools": "3",
    "project-state-files": "3",
}
TEST_WORDS = [
    ("1", r"launcher|session_|update|home|host_home|model|shim|user_assets|skills_legacy|migrate|install_"),
    ("4", r"lane|leader|card|decision|scope|trailer|review|audit|contract|gate|ledger"),
    ("2", r"hook|project|generation|agents|go_state|codex|statusline|goal|profile|host|omp"),
]


# Files the leader's decision D6 keeps (Go-era `.codex` handling, `~/.codex/worktrees` guard, protected
# paths, readers of old records) and the tests that exercise them: no removal slice edits them.
KEPT_BY_DECISION = {
    "src/go_state.rs", "src/update_cleanup.rs", "src/worktree_gc.rs", "src/gate.rs",
    "src/decision.rs", "src/user_assets.rs", "src/migrate.rs",
    "tests/go_state_codex_cli.rs", "tests/go_state_classes_cli.rs", "tests/go_state_cli.rs",
    "tests/update_cleanup_cli.rs", "tests/gate_check_cli.rs", "tests/migrate_cli.rs",
    "tests/worktree_gc_install_dir.rs", "tests/decision_append_cli.rs",
    # Landed compatibility layer and its tests: they name the retired hosts on purpose (they read old
    # profiles, receipts and settings), so the final word scan lists them.
    "src/retired_hosts.rs", "src/hosts/config_backup.rs", "tests/retired_hosts_layer.rs",
    "tests/codex_launcher_removed.rs",
}
# Where the impact report's own card assignment (section 2 and 6) differs from what the area rule
# or the test-name words give. `1>2`: slice 1 edits first, slice 2 takes the file after it.
SLICE_OVERRIDES = {
    "src/host_home.rs": "1>2",  # slice 1 drops the users, slice 2 drops the Codex and OMP variants last
    "src/audit.rs": "2", "src/contract.rs": "2", "tests/audit_admission.rs": "2",
    "src/labels_check.rs": "3",
    # The Codex session id (`CODEX_THREAD_ID`) goes with the lane and card tooling, after no lane
    # session runs on Codex (D-6); the report put it with the project layer.
    "src/session_current.rs": "4", "src/ops/session_current.rs": "4",
    "tests/session_current_cli.rs": "4",
    "tests/migration_fixture.rs": "2",
    "tests/project_skill_apply_gate_cli.rs": "2",
    "tests/ci_scope_test.py": "3", "tests/mode_cards_53.rs": "3",
    "tests/host_matrix_test.py": "3", "tests/support_matrix_test.py": "3",
    "tests/host_live_harness_test.py": "3",
    "tests/fixtures/update_tux/preview-blocked.json": "1",
    "tests/fixtures/update_tux/preview-changes.json": "1",
    "tests/fixtures/update_tux/preview-latest.json": "1",
    "tests/fixtures/update_tux/applied.json": "1",
    # The agent renderers and the agent installer write `~/.codex/agents` and `~/.omp/agent/agents`:
    # they go in the same commit as the update's host list (slice 1), not with the project layer.
    "src/agent_defs.rs": "1", "src/agent_install.rs": "1", "tests/agent_defs_package.rs": "1",
    # The project update and its hook step belong to the project layer (slice 2), whatever their
    # test names say.
    "tests/update_hooks_cli.rs": "2", "tests/project_update_hooks_cli.rs": "2",
    "tests/project_update_plan_cli.rs": "2",
    # The OMP launcher and the install texts are one card (slice 1c); the rest of the install folder
    # and the docs stay with slice 3.
    "tests/omp_launcher.rs": "1c", "tests/launcher_agent_entry.rs": "1c",
    "install/install.sh": "1c", "install/install.ps1": "1c", "install/README.md": "1c",
}


def slice_of(area, path):
    if path in KEPT_BY_DECISION:
        return "keep"
    if path in SLICE_OVERRIDES:
        return SLICE_OVERRIDES[path]
    if path == "schemas/protocol/v1/profile.schema.json":
        return "2"  # the profile's host list changes together with the profile code
    if area == "project-state-files":
        # This repository's own profile follows the profile handling (slice 2); the generated
        # manifests and the rest are regenerated at merge (slice 3).
        return "2" if path.endswith("project.yaml") else "3"
    if area == "codex-only-files":
        # Codex-only code and tests go with the project layer; docs, the tracked `.codex/` folder
        # (leader decision 2026-11-21: the shipped-text slice deletes it) and the rest with the
        # text slice.
        return "2" if path.startswith(("src/", "tests/")) else "3"
    if area != "tests-and-fixtures":
        return SLICE_OF_AREA.get(area, "3")
    name = path.rsplit("/", 1)[-1].lower()
    for number, words in TEST_WORDS:
        if re.search(words, name):
            return number
    return "3"


def top():
    return subprocess.run(["/usr/bin/git", "rev-parse", "--show-toplevel"], capture_output=True,
                          text=True, check=True).stdout.strip()


def grep(commit, *flags):
    # From the top of the repository, so paths are repository paths wherever this runs.
    out = subprocess.run(["/usr/bin/git", "grep", "-n", "-I", "-i", *flags, commit, "--", "."],
                         capture_output=True, text=True, check=False, cwd=top())
    if out.returncode not in (0, 1):
        sys.exit(out.stderr)
    rows = set()
    prefix = commit + ":"
    for line in out.stdout.splitlines():
        if not line.startswith(prefix):
            continue
        path, number, _ = line[len(prefix):].split(":", 2)
        rows.add((path, int(number)))
    return rows


NAMED = re.compile(r"codex|(^|[/_.-])omp([/_.-]|$)", re.IGNORECASE)


def named_files(commit):
    """Tracked files of the commit whose path holds `codex`, or `omp` as a word of the path."""
    out = subprocess.run(["/usr/bin/git", "ls-tree", "-r", "--name-only", commit],
                         capture_output=True, text=True, check=True, cwd=top())
    return [p for p in out.stdout.splitlines() if NAMED.search(p)]


def main():
    ap = argparse.ArgumentParser()
    ap.add_argument("commit")
    ap.add_argument("--out", default="inventory.tsv")
    ap.add_argument("--check", type=int, help="exit 1 unless the inventory has this many files")
    args = ap.parse_args()
    commit = subprocess.run(["/usr/bin/git", "rev-parse", args.commit], capture_output=True,
                            text=True, check=True).stdout.strip()
    hits = grep(commit, "-e", "codex") | grep(commit, "-w", "-e", "omp")
    by_file = defaultdict(list)
    for path, number in sorted(hits):
        by_file[path].append(number)
    # A file whose path names Codex or OMP counts even when its text never does (0 lines).
    for path in named_files(commit):
        by_file.setdefault(path, [])
    rows = sorted(((area_of(p), p, len(n), ",".join(map(str, n))) for p, n in by_file.items()),
                  key=lambda r: (AREA_ORDER.index(r[0]) if r[0] in AREA_ORDER else 99, r[1]))
    with open(args.out, "w", encoding="utf-8") as fh:
        fh.write(f"# commit {commit}\n# area\tslice\tpath\tlines\tline numbers\n")
        for area, path, count, numbers in rows:
            # `-` for a file whose path, not its text, names Codex (no line numbers).
            fh.write(f"{area}\t{slice_of(area, path)}\t{path}\t{count}\t{numbers or '-'}\n")
    totals = defaultdict(lambda: [0, 0])
    per_slice = defaultdict(lambda: [0, 0])
    for area, path, count, _ in rows:
        totals[area][0] += 1
        totals[area][1] += count
        owner = slice_of(area, path)
        per_slice[owner][0] += 1
        per_slice[owner][1] += count
    print(f"commit {commit}")
    for area in AREA_ORDER + ["unassigned"]:
        if area in totals:
            print(f"area\t{area}\tfiles={totals[area][0]}\tlines={totals[area][1]}")
    for owner in sorted(per_slice):
        print(f"slice\t{owner}\tfiles={per_slice[owner][0]}\tlines={per_slice[owner][1]}")
    files = len(rows)
    lines = sum(r[2] for r in rows)
    print(f"total\tfiles={files}\tlines={lines}")
    if args.check is not None and files != args.check:
        print(f"expected {args.check} files, found {files}")
        return 1
    return 0


if __name__ == "__main__":
    sys.exit(main())
