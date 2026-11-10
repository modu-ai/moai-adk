#!/usr/bin/env python3
"""Recorded-AC-command lint (#2352): a `cargo test --test X [filters…]` line in the
recorded verification commands must name an existing test target and its filters must
select at least one test. cargo exits 0 with "running 0 tests" for a filter that
matches nothing, so an AC command that selects nothing has passed as verification
before (survey 2.11).

Usage:
  python3 -I tools/check_ac_commands.py --tree <repo root>
  python3 -I tools/check_ac_commands.py --file <one recorded file>

Scans `docs/**/*.md` and `.moai-cli/reports/**/*.md` under --tree, or the one --file
given. A candidate command is a `cargo test` line inside a fenced code block or a
backtick code span (table rows that only mention commands are handled by this: only
the quoted span is tokenized, trailing prose never becomes a filter). Env prefixes
(`CARGO_HOME=… cargo test …`) are stripped. The target must have `tests/<name>.rs`;
its filters (positional words after `--test <name>` up to the next flag) must select
at least one test, checked against `cargo test --test <name> -- --list` (one listing
per target, cached). Placeholder filters (`<이름>`) are skipped as templates.
Exit 0 = all clear; 1 = violations listed on stdout; 2 = usage.
"""

import re
import subprocess
import sys
from pathlib import Path

CARGO_TEST = re.compile(r"cargo\s+test\b")
TEST_TARGET = re.compile(r"--test\s+([A-Za-z0-9_-]+)")
ENV_PREFIX = re.compile(r"^(?:[A-Z_][A-Z0-9_]*=\S+\s+)+")
CODE_SPAN = re.compile(r"`([^`\n]+)`")
HANGUL = re.compile(r"[가-힣]")


def scan_files(root: Path):
    yield from sorted(root.glob("docs/**/*.md"))
    reports = root / ".moai-cli" / "reports"
    yield from sorted(p for p in reports.glob("*.md"))
    # Frozen RC1 records stay out: the closed-issue mirrors (< 2153, the rc.2 split)
    # and the rc1-issue-design HTML export — survey 2.11 already recorded that their
    # commands select no tests today, and a historical record is not edited to match
    # a renamed test. Live records (rc.2 bodies 2153+, root reports, docs) stay
    # guarded.
    for path in sorted((reports / "issue-bodies").glob("*.md")):
        m = re.match(r"(\d+)-", path.name)
        if m and int(m.group(1)) >= 2153:
            yield path


def candidates(text: str):
    """`cargo test` command strings: fenced-block lines and backtick code spans."""
    in_fence = False
    for line in text.splitlines():
        stripped = line.strip()
        if stripped.startswith("```"):
            in_fence = not in_fence
            continue
        if in_fence:
            yield line
            continue
        for span in CODE_SPAN.findall(line):
            yield span


def tokens_of(candidate: str):
    """Shell-like tokenization good enough for a recorded cargo command."""
    line = candidate.strip()
    while True:
        stripped = ENV_PREFIX.sub("", line)
        if stripped == line:
            break
        line = stripped
    if "`" in line:
        line = line.strip("`")
    return line.split()


def filters_after_target(tokens):
    """Positional filter words after `--test <name>` up to the next flag token."""
    try:
        idx = tokens.index("--test")
    except ValueError:
        return []
    out = []
    for tok in tokens[idx + 2:]:
        if tok.startswith("-"):
            break
        out.append(tok)
    return out


def target_exists(root: Path, name: str) -> bool:
    return (root / "tests" / f"{name}.rs").is_file()


def list_tests(root: Path, name: str, cache: dict):
    if name in cache:
        return cache[name]
    out = subprocess.run(
        ["cargo", "test", "--offline", "--locked", "--test", name, "--", "--list"],
        cwd=root, capture_output=True, text=True,
    )
    tests = set()
    for line in out.stdout.splitlines():
        line = line.strip()
        if ": test" in line:
            tests.add(line.split(":", 1)[0].strip())
    cache[name] = tests
    return tests


def selects_any(filters, tests):
    """cargo's filter rule: a test is selected when ANY positional filter is a
    substring of the full test path; no filters means 'select everything'."""
    if not filters:
        return True
    return any(any(f in t for t in tests) for f in filters)


def violations_in(path: Path, text: str, root: Path, cache: dict):
    if path.suffix == ".html":
        text = re.sub(r"<[^>]+>", " ", text)
    found = []
    for candidate in candidates(text):
        if not CARGO_TEST.search(candidate) or "--test" not in candidate:
            continue
        tokens = tokens_of(candidate)
        m = TEST_TARGET.search(" ".join(tokens))
        if not m:
            continue
        name = m.group(1)
        where = f"{path}"
        if not target_exists(root, name):
            found.append(f"{where}: target tests/{name}.rs does not exist :: {candidate.strip()[:170]}")
            continue
        filters = [f for f in filters_after_target(tokens) if not f.startswith("<")]
        tests = list_tests(root, name, cache)
        if not tests:
            found.append(f"{where}: target {name} lists no tests :: {candidate.strip()[:170]}")
            continue
        if not selects_any(filters, tests):
            found.append(
                f"{where}: filters {filters} select 0 tests in {name} "
                f"({len(tests)} available) :: {candidate.strip()[:170]}")
    return found


def main(argv):
    if len(argv) == 3 and argv[1] == "--tree":
        root, files = Path(argv[2]), list(scan_files(Path(argv[2])))
    elif len(argv) == 3 and argv[1] == "--file":
        root, files = Path.cwd(), [Path(argv[2])]
    else:
        print(__doc__, file=sys.stderr)
        return 2

    cache: dict = {}
    violations = []
    for path in files:
        try:
            text = path.read_text(encoding="utf-8", errors="replace")
        except OSError as exc:
            violations.append(f"{path}: unreadable ({exc})")
            continue
        violations.extend(violations_in(path, text, root, cache))

    print(f"checked {len(files)} recorded files, {len(cache)} --test targets listed")
    for v in violations:
        print("VIOLATION:", v)
    return 1 if violations else 0


if __name__ == "__main__":
    raise SystemExit(main(sys.argv))
