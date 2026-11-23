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
at least one test, checked against the test names read from `tests/<name>.rs` and the
module files it declares, as cargo prints them (`module::name`; one reading per target,
cached). No cargo is started (SPEC-20261122-ac-lint-no-cargo): a `cargo --list` here built
any target not built yet in the middle of the test. Names a `macro_rules!` body makes
cannot be read; a filter that only such a name would satisfy is reported as such. Placeholder filters (`<이름>`) are skipped as templates.
Exit 0 = all clear; 1 = violations listed on stdout; 2 = usage.
"""

import re
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


def resolve_target(root: Path, name: str):
    """(file to read, folder its `mod x;` lines resolve in) of the test target `name`, or None.

    A target is a top-level `tests/<name>.rs`, an area crate `tests/<name>/main.rs`, or the old
    name of a file moved into an area (`tests/<area>/<name>.rs`, a module of that crate): closed
    SPECs keep `--test <old name>` and it still resolves (SPEC-20261123-test-area-crates REQ-6).
    Found by where the files are, so no table is read; tests/test_area_layout.rs keeps the table
    and the files in step. Anything else is None (REQ-12).
    """
    tests_dir = root / "tests"
    top = tests_dir / f"{name}.rs"
    if top.is_file():
        return top, tests_dir
    crate = tests_dir / name / "main.rs"
    if crate.is_file():
        return crate, tests_dir / name
    if name != "main" and tests_dir.is_dir():
        for area in sorted(p for p in tests_dir.iterdir() if p.is_dir()):
            module = area / f"{name}.rs"
            if (area / "main.rs").is_file() and module.is_file():
                return module, area / name
    return None


def target_exists(root: Path, name: str) -> bool:
    return resolve_target(root, name) is not None


MOD_OPEN = re.compile(r"(?:pub(?:\([^)]*\))?\s+)?mod\s+([A-Za-z_][A-Za-z0-9_]*)\s*([{;])")
FN_NAME = re.compile(r"(?:pub(?:\([^)]*\))?\s+)?(?:const\s+)?(?:async\s+)?(?:unsafe\s+)?fn\s+([A-Za-z_][A-Za-z0-9_]*)")
PATH_ATTR = re.compile(r'path\s*=\s*"([^"]+)"')


def code_only(src: str) -> str:
    """The source with comments, string, raw-string and char literals blanked to spaces
    (string contents kept only inside attributes, where `#[path = "…"]` needs them).
    Lengths are not kept; only the token order matters to the reader."""
    out = []
    i, n = 0, len(src)
    attr_depth = 0  # > 0 while inside `#[ … ]`, so `#[path = "x"]` keeps its string
    while i < n:
        c = src[i]
        if src.startswith("//", i):
            j = src.find("\n", i)
            i = n if j < 0 else j
            continue
        if src.startswith("/*", i):
            depth, i = 1, i + 2
            while i < n and depth:
                if src.startswith("/*", i):
                    depth, i = depth + 1, i + 2
                elif src.startswith("*/", i):
                    depth, i = depth - 1, i + 2
                else:
                    i += 1
            out.append(" ")
            continue
        m = re.match(r'b?r(#*)"', src[i:])
        if m and (i == 0 or not (src[i - 1].isalnum() or src[i - 1] == "_")):
            close = '"' + m.group(1)
            j = src.find(close, i + len(m.group(0)))
            j = n if j < 0 else j + len(close)
            out.append(src[i:j] if attr_depth else '""')
            i = j
            continue
        if c == '"':
            j = i + 1
            while j < n and src[j] != '"':
                j += 2 if src[j] == "\\" else 1
            out.append(src[i:j + 1] if attr_depth else '""')
            i = j + 1
            continue
        if c == "'":
            m = re.match(r"'(?:\\(?:x[0-9a-fA-F]{2}|u\{[0-9a-fA-F]{1,6}\}|.)|[^\\'\n])'", src[i:])
            if m:
                out.append("' '")
                i += len(m.group(0))
                continue
        if src.startswith("#[", i) or src.startswith("#![", i):
            attr_depth += 1
        elif c == "]" and attr_depth:
            attr_depth -= 1
        out.append(c)
        i += 1
    return "".join(out)


def module_file(base: Path, name: str):
    """The file of `mod name;` declared in a file whose child modules live under `base`."""
    for candidate in (base / f"{name}.rs", base / name / "mod.rs"):
        if candidate.is_file():
            return candidate
    return None


def read_tests(file: Path, base: Path, prefix: str, out: set, seen: set) -> bool:
    """Adds `module::name` for every test function in `file` (and the module files it
    declares) to `out`. Returns whether a `macro_rules!` body there holds a test attribute."""
    if file in seen or not file.is_file():
        return False
    seen.add(file)
    code = code_only(file.read_text(encoding="utf-8", errors="replace"))
    macro_tests = False
    for m in re.finditer(r"macro_rules!\s*[A-Za-z_][A-Za-z0-9_]*\s*\{", code):
        depth, j = 1, m.end()
        while j < len(code) and depth:
            depth += {"{": 1, "}": -1}.get(code[j], 0)
            j += 1
        if re.search(r"#\[\s*(?:[A-Za-z_][A-Za-z0-9_]*\s*::\s*)*test\s*[\](]", code[m.end():j]):
            macro_tests = True
    stack = []  # (module name, brace depth at its opening)
    depth = 0
    pending_test = False
    pending_path = None
    i, n = 0, len(code)
    while i < n:
        c = code[i]
        if c == "#" and code.startswith("[", i + 1):
            depth_a, j = 1, i + 2
            while j < n and depth_a:
                depth_a += {"[": 1, "]": -1}.get(code[j], 0)
                j += 1
            body = code[i + 2:j - 1].strip()
            head = re.match(r"([A-Za-z_][A-Za-z0-9_]*(?:\s*::\s*[A-Za-z_][A-Za-z0-9_]*)*)", body)
            rest = body[len(head.group(1)):].strip() if head else ""
            if head and head.group(1).split("::")[-1].strip() == "test" and (not rest or rest.startswith("(")):
                pending_test = True
            pm = PATH_ATTR.match(body)
            if pm:
                pending_path = pm.group(1)
            i = j
            continue
        if c == "{":
            depth += 1
            i += 1
            continue
        if c == "}":
            depth -= 1
            if stack and stack[-1][1] == depth:
                stack.pop()
            i += 1
            continue
        if c.isalpha() or c == "_":
            m = MOD_OPEN.match(code, i)
            if m and (i == 0 or not (code[i - 1].isalnum() or code[i - 1] == "_")):
                name, kind = m.group(1), m.group(2)
                if kind == "{":
                    stack.append((name, depth))
                    depth += 1
                else:
                    inner = [s[0] for s in stack]
                    sub_base = base.joinpath(*inner) if inner else base
                    if pending_path:
                        # rustc: relative to the declaring file's folder outside inline modules
                        target = (file.parent if not inner else sub_base) / pending_path
                    else:
                        target = module_file(sub_base, name)
                    if target is not None and target.is_file():
                        child_base = target.parent if target.name == "mod.rs" or pending_path else target.parent / target.stem
                        mod_prefix = "::".join(p for p in (prefix, *inner, name) if p)
                        macro_tests |= read_tests(target, child_base, mod_prefix, out, seen)
                pending_test, pending_path = False, None
                i = m.end()
                continue
            m = FN_NAME.match(code, i)
            if m and (i == 0 or not (code[i - 1].isalnum() or code[i - 1] == "_")):
                if pending_test:
                    out.add("::".join(p for p in (prefix, *(s[0] for s in stack), m.group(1)) if p))
                pending_test, pending_path = False, None
                i = m.end()
                continue
            j = i
            while j < n and (code[j].isalnum() or code[j] == "_"):
                j += 1
            pending_test, pending_path = False, None
            i = j
            continue
        if c == ";":
            pending_test, pending_path = False, None
        i += 1
    return macro_tests


def list_tests(root: Path, name: str, cache: dict):
    """(test names, whether a macro_rules! body makes tests) of `tests/<name>.rs`, read from
    the sources: no cargo is started (SPEC-20261122-ac-lint-no-cargo). A moved file's names keep
    their old form (no module prefix), so an old recorded filter selects what it always did."""
    if name not in cache:
        tests: set = set()
        source, base = resolve_target(root, name)
        macro = read_tests(source, base, "", tests, set())
        cache[name] = (tests, macro)
    return cache[name]


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
        tests, macro = list_tests(root, name, cache)
        if macro and not selects_any(filters, tests):
            found.append(
                f"{where}: filters {filters} select 0 written tests in {name}; its other tests "
                f"are made by a macro_rules! body, whose names cannot be read without cargo: "
                f"name a written test :: {candidate.strip()[:170]}")
            continue
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
