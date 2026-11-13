#!/usr/bin/env python3
"""PR self-audit section checker (Issue #2279, scope 3).

Every lane PR must carry a `## Self-Audit` section with four measured lines
(read-only checker; it judges text, it never posts anything):

  mutation: what guard was mutated and that the mutation FAILED a check
  clone:    the check set run from a fresh clone (symlink paths like /tmp cannot
            leak pre-seeded state in)
  host:     external hosts exercised through a fake executable (stub), not the
            network
  receipt:  how a false-pass receipt was ruled out (digest/reference checked)

A PR body without the section, missing any of the four lines, or a mutation
line that does not report FAILED exits 1.

Optional CI citation lines (Issue #2385; rules in skills/moai/workflows/
verify.md, section "Check-run rules"). Since decision D4 (Issue #1975 comment
6840449133, Issue #2612) the author writes no such line: the auditor reads the
head's CI run and cites it in the verdict. During the transition a body with
the line and a body without it both pass; a line that is present is still
judged by the rules below:

  ci: ci run=<run id> head=<40-hex head SHA>
  ci: not_run <reason>

The workflow was named `rc-ci` before Issue #2563, so open PRs may still carry
`ci: rc-ci run=…`. Only the keyed `run=` and `head=` tokens are judged, never
the workflow word, so both spellings pass under exactly the same token rules.

Each `ci:` line in the section is checked. A citation must carry the keyed
field `run=` with one or more digits and the keyed field `head=` with exactly
40 lowercase hex characters, each as a whole whitespace-separated token (39 or
41 hex, or a suffix such as `-dirty`, fail); otherwise it must start with
`not_run` followed by a reason containing a letter or digit. A body with no `ci:` line is accepted
(decision D4: the citation belongs to the verdict). This checker reads
text only: it cannot tell whether the run exists or whether its head equals
the PR head; the auditor reads the run back with `gh api`. The independent audit stays in
place — this tool only pulls the audit evidence forward, before the PR is
opened.

Re-audits (Issue #2388) require explicit `--reaudit` (or check(body, reaudit=True));
neither template fields nor a claim epoch enables this mode. In the exact
`## 독립 review` section, up to the next Markdown heading, require exactly one
visible line for each field (whitespace and Markdown list markers are accepted).
HTML comments, fenced examples and indented code are excluded; multiple visible
review headings are ambiguous and fail:

  - 수리 확인 범위: F1 path/boundary repaired and touched scope
  - 새 지적 분류: 없음

The classification can instead be `차단: <concrete reason>` or
`후속 Issue #<positive ASCII integer>: <concrete reason>`. Angle-bracket text is
documentation notation, not an accepted placeholder. Empty, duplicate,
punctuation-only and unfinished TODO/TBD/not_run values fail. Completed prose
may reference path identifiers, inline code and comparisons. Contract violations, data
damage and security findings are blocking; other new findings belong in a
follow-up Issue. This is a text-completeness check only: it cannot judge repair
truth, classification policy compliance, or whether the follow-up Issue exists.
Initial audits keep the existing checks and output; the AUDIT verdict format
and CI requirements are unchanged.

Zero-selected evidence (Issue #2403, item 5). In the section and in every
markdown table row of the body:

  - a line that reports `0 passed` (a run that selected no test) must also say
    `not_run` on the same line; otherwise it is blocked as zero-selected;
  - a `cargo test` command that passes a test-name filter must select exactly
    one target (`--lib`, `--bin X`, `--test X`, ...), because cargo applies the
    filter to every target of the call; a filtered call with two or more
    targets, or with none (cargo then runs several by default), is blocked as
    filter-multi-target. Only commands written in a code span (`...`) are
    read; prose that mentions cargo test is not a command.
    Quoted option values remain single POSIX shell words; shell redirections
    and their destinations are not filters. Cargo global options before `test`
    are recognized. Tables may omit their edge pipes when a header and Markdown
    delimiter establish table scope. Named target glob patterns (`*`, `?`, `[]`)
    may select several targets, including in `--test=X` form, so they cannot
    establish single-target filtered evidence. No Cargo command or GitHub
    request is executed by this checker.

Usage:
  python3 -I tools/check_pr_self_audit.py --file BODY.md
  python3 -I tools/check_pr_self_audit.py < BODY.md
  python3 -I tools/check_pr_self_audit.py --reaudit < BODY.md
  python3 -I tools/check_pr_self_audit.py --reaudit --file BODY.md
  python3 -I tools/check_pr_self_audit.py --file BODY.md --reaudit

Exit codes: 0 = section complete, 1 = section/items missing or unmeasured,
2 = usage.
"""

import re
import shlex
import sys
from pathlib import Path

SECTION = "## self-audit"
TAGS = ("mutation", "clone", "host", "receipt")
CI_TAG = "ci:"
# Keyed fields are whole whitespace-separated tokens: whitespace (or the line
# edge) on both sides, so `rerun=7`, `.run=7`, a 41-hex string or a suffix such
# as `<40hex>-dirty` never passes as `run=` or `head=`.
CI_RUN = re.compile(r"(?<!\S)run=\d+(?!\S)")
CI_HEAD = re.compile(r"(?<!\S)head=[0-9a-f]{40}(?!\S)")
# `not_run` followed by a reason that has at least one letter or digit.
CI_NOT_RUN = re.compile(r"^not_run\b(?=.*\w)")


def fail(missing):
    print("self-audit: blocked missing=" + ",".join(missing))
    raise SystemExit(1)


def strip_marker(line: str) -> str:
    """Drop leading whitespace and markdown list markers before tag matching."""
    return line.strip().lstrip("-*").strip().lower()


# `0 passed` as a whole count: not preceded by another digit, so `10 passed`
# and `20 passed` are not zero.
ZERO_PASSED = re.compile(r"(?<![0-9])0 passed\b")
# cargo test options that take a value as the next token (the value is not a
# test-name filter).
CARGO_VALUE_OPTS = {
    "--test", "--bin", "--example", "--bench", "-p", "--package", "-F",
    "--features", "--manifest-path", "--target", "--target-dir", "-j",
    "--jobs", "--profile", "--color", "--message-format", "--config", "-Z",
    "--exclude",
}
# Target selectors and how many targets each selects.
CARGO_TARGETS_ONE = {"--lib", "--doc", "--test", "--bin", "--example", "--bench"}
CARGO_TARGETS_MANY = {"--tests", "--bins", "--examples", "--benches", "--all-targets"}
# Test-harness options after `--` that take a value as the next token.
HARNESS_VALUE_OPTS = {"--skip", "--test-threads", "--format", "--color", "-Z", "--logfile"}
CARGO_GLOBAL_FLAGS = {"--offline", "--locked", "--frozen", "-q", "--quiet", "-v", "--verbose"}
CARGO_GLOBAL_VALUES = {"--color", "--config", "-Z", "-C"}
# Preserve whether punctuation was shell syntax or a quoted literal word.
SHELL_TOKEN = re.compile(
    r"""(?P<redirect>&>>|&>|\d*(?:<<<|<<|<>|>>|>\||>&|<&|>|<))"""
    r"""|(?P<separator>&&|\|\||[;&|()])"""
    r"""|(?P<word>(?:[^\s;&|()<>'"\\]|\\.|'[^']*'|"(?:\\.|[^"\\])*")+)"""
)


def cargo_test_commands(line: str):
    """The token lists of the `cargo test` commands written in code spans on one
    line: each runs from `test` to the end of its span or a shell separator.
    Prose such as "cargo test 671 passed" outside a code span is not a command."""
    out = []
    for span in re.findall(r"`([^`]+)`", line):
        if not re.search(r"(?<![\w'])cargo(?=\s|$)", span):
            continue
        words = []
        position = 0
        malformed = False
        for match in SHELL_TOKEN.finditer(span):
            if span[position:match.start()].strip():
                malformed = True
                break
            position = match.end()
            raw = match.group()
            if match.lastgroup == "word":
                try:
                    words.append((shlex.split(raw, posix=True)[0], "word"))
                except (ValueError, IndexError):
                    malformed = True
                    break
            else:
                words.append((raw, match.lastgroup))
        if malformed or span[position:].strip():
            continue
        i = 0
        while i < len(words):
            if words[i] != ("cargo", "word"):
                i += 1
                continue
            test = i + 1
            while test < len(words):
                token, kind = words[test]
                if kind == "redirect":
                    test += 1
                    if test < len(words) and words[test][1] == "word":
                        test += 1
                    continue
                if kind != "word":
                    break
                name = token.split("=", 1)[0]
                if token.startswith("+") or token in CARGO_GLOBAL_FLAGS:
                    test += 1
                elif name in CARGO_GLOBAL_VALUES:
                    test += 1 if "=" in token else 2
                elif (re.fullmatch(r"-[qv]+", token)
                      or (token.startswith(("-Z", "-C")) and len(token) > 2)):
                    test += 1
                else:
                    break
            if test >= len(words) or words[test] != ("test", "word"):
                i += 1
                continue
            end = test + 1
            tokens = []
            while end < len(words):
                token, kind = words[end]
                if kind == "separator":
                    break
                if kind == "redirect":
                    # Both file destinations and fd duplication operands are
                    # shell plumbing, not positional Cargo/harness arguments.
                    end += 1
                    if end < len(words) and words[end][1] == "word":
                        end += 1
                    continue
                tokens.append(token)
                end += 1
            out.append(tokens)
            i = end
    return out


def filtered_target_count(tokens):
    """(has a name filter, number of targets selected; None = cargo's default
    set, which is more than one)."""
    has_filter = False
    targets = 0
    many = False
    after_dashdash = False
    skip_next = None
    for tok in tokens:
        if skip_next:
            if skip_next in CARGO_TARGETS_ONE and any(c in tok for c in "*?["):
                many = True
            skip_next = None
            continue
        if after_dashdash:
            if tok in HARNESS_VALUE_OPTS:
                skip_next = tok
            elif not tok.startswith("-"):
                has_filter = True
            continue
        if tok == "--":
            after_dashdash = True
            continue
        name = tok.split("=", 1)[0]
        if name in CARGO_TARGETS_ONE:
            targets += 1
            if "=" in tok and any(c in tok.split("=", 1)[1] for c in "*?["):
                many = True
        elif name in CARGO_TARGETS_MANY:
            many = True
        if name in CARGO_VALUE_OPTS and "=" not in tok:
            skip_next = name
        elif not tok.startswith("-"):
            has_filter = True
    if many:
        return has_filter, 2
    return has_filter, (targets if targets else None)


def check_zero_selected(section, lines) -> None:
    """Block zero-selected test evidence in the section and in table rows."""
    scope = list(section) + [ln for ln in lines if ln.lstrip().startswith("|")]
    # Optional-edge-pipe tables require a header and a delimiter, rather than
    # treating arbitrary prose containing a pipe as test evidence.
    def cells(line):
        return line.strip().strip("|").split("|")

    for index in range(1, len(lines)):
        delimiter = cells(lines[index])
        if (len(delimiter) < 2
                or not all(re.fullmatch(r"\s*:?-{3,}:?\s*", cell) for cell in delimiter)
                or "|" not in lines[index - 1]
                or len(cells(lines[index - 1])) != len(delimiter)):
            continue
        scope.append(lines[index - 1])
        for row in lines[index + 1:]:
            if "|" not in row:
                break
            scope.append(row)
    for line in scope:
        if ZERO_PASSED.search(line) and "not_run" not in line:
            fail(["zero-selected"])
        for tokens in cargo_test_commands(line):
            has_filter, count = filtered_target_count(tokens)
            if has_filter and (count is None or count > 1):
                fail(["filter-multi-target"])


def check(body: str, *, reaudit: bool = False) -> None:
    lines = body.splitlines()
    starts = [i for i, ln in enumerate(lines) if ln.strip().lower() == SECTION]
    if not starts:
        fail([SECTION])
    start = starts[0]
    # The section runs to the next markdown heading of any level.
    section = []
    for line in lines[start + 1:]:
        if line.startswith("#"):
            break
        section.append(line)
    missing = []
    for tag in TAGS:
        hit = [ln for ln in section if strip_marker(ln).startswith(tag + ":")]
        if not hit or not hit[0].split(":", 1)[1].strip():
            missing.append(tag)
    if missing:
        fail(missing)
    check_zero_selected(section, lines)
    # The mutation line must report the mutation actually failing a check
    # (measured, not asserted): the word FAILED must appear on it.
    mut = [ln for ln in section if strip_marker(ln).startswith("mutation:")][0]
    if "failed" not in mut.lower():
        fail(["mutation:FAILED"])
    check_ci_lines(section)
    if reaudit:
        check_reaudit(lines)
    print("self-audit: pass (mutation/clone/host/receipt measured)")


def check_ci_lines(section) -> None:
    """Every `ci:` line is either a keyed citation (run id + 40-hex head) or
    `not_run` with a reason."""
    for line in section:
        if not strip_marker(line).startswith(CI_TAG):
            continue
        value = line.split(":", 1)[1].strip()
        if CI_NOT_RUN.match(value):
            continue
        missing = []
        if not CI_RUN.search(value):
            missing.append("ci:run")
        if not CI_HEAD.search(value):
            missing.append("ci:head")
        if missing:
            fail(missing)


def meaningful(value: str) -> bool:
    """Reject unfilled evidence without attempting to judge its truth."""
    # Inline code and path identifiers describe evidence, not unfinished work.
    prose = re.sub(r"(`+)(?!`)(.+?)(?<!`)\1(?!`)", "", value)
    return (
        any(char.isalnum() for char in value)
        and not re.search(
            r"(?<![\w./\\-])(?:TODO|TBD|not_run)(?![\w./\\-])",
            prose, re.IGNORECASE,
        )
        and not re.search(r"<[^\s<>]+(?:[ \t]+[^\s<>]+)*>", value)
        and not re.fullmatch(r"\s*[`*_]*\s*(?:TODO|TBD|not_run)[\s`*_.-]*",
                             value, re.IGNORECASE)
    )


def visible_reaudit_lines(lines):
    """Exclude Markdown examples and HTML comments from re-audit evidence."""
    in_comment = False
    fence = None
    for line in lines:
        if fence:
            if re.fullmatch(r" {0,3}" + re.escape(fence[0]) +
                            r"{" + str(fence[1]) + r",}[ \t]*", line):
                fence = None
            continue
        # Indented examples cannot open comments or fenced blocks. Retain the
        # existing tab-prefixed list-marker convention for evidence fields.
        if not in_comment and (
                line.startswith("    ") or line.startswith("\t")) and not re.match(
                    r"^\t[-*+]\s", line):
            continue
        visible = []
        position = 0
        while position < len(line):
            if in_comment:
                end = line.find("-->", position)
                if end < 0:
                    break
                in_comment = False
                position = end + 3
            else:
                token = re.search(r"\\[`\\]|`+|<!--", line[position:])
                if token is None:
                    visible.append(line[position:])
                    break
                start = position + token.start()
                end = position + token.end()
                visible.append(line[position:start])
                if token[0].startswith("\\"):
                    visible.append(token[0])
                    position = end
                    continue
                if token[0] == "<!--":
                    in_comment = True
                    position = end
                    continue
                # Only an equal-length backtick run closes an inline code
                # span. An unmatched run is literal text, not comment masking.
                closer = re.search(
                    r"(?<!`)" + re.escape(token[0]) + r"(?!`)", line[end:],
                )
                position = end if closer is None else end + closer.end()
                visible.append(line[start:position])
        line = "".join(visible)
        opener = re.match(r"^ {0,3}(`{3,}|~{3,})(.*)$", line)
        if opener and (opener[1][0] != "`" or "`" not in opener[2]):
            fence = (opener[1][0], len(opener[1]))
            continue
        yield line


def check_reaudit(lines) -> None:
    """Require one repair scope and one explicit new-finding classification."""
    lines = list(visible_reaudit_lines(lines))
    starts = [i for i, line in enumerate(lines) if line.strip() == "## 독립 review"]
    if len(starts) != 1:
        fail(["## 독립 review"])
    fields = {"수리 확인 범위": [], "새 지적 분류": []}
    for line in lines[starts[0] + 1:]:
        if re.match(r"^\s*#{1,6}(?:\s|$)", line):
            break
        match = re.fullmatch(
            r"\s*(?:(?:[-*+]|[0-9]+[.)])\s+)?"
            r"(수리 확인 범위|새 지적 분류)\s*:\s*(.*)", line,
        )
        if match:
            fields[match[1]].append(match[2].strip())
    missing = []
    scope = fields["수리 확인 범위"]
    if len(scope) != 1 or not meaningful(scope[0]):
        missing.append("reaudit:scope")
    findings = fields["새 지적 분류"]
    valid_findings = False
    if len(findings) == 1 and meaningful(findings[0]):
        if findings[0] == "없음":
            valid_findings = True
        else:
            match = re.fullmatch(
                r"(?:차단|후속\s+Issue\s+#[1-9][0-9]*)\s*:\s*(.*)",
                findings[0],
            )
            valid_findings = bool(match and meaningful(match[1]))
    if not valid_findings:
        missing.append("reaudit:findings")
    if missing:
        fail(missing)


def usage() -> None:
    print(__doc__)
    raise SystemExit(2)


def main() -> None:
    args = [a for a in sys.argv[1:]]
    reaudit = False
    if args and args[0] == "--reaudit":
        reaudit = True
        args.pop(0)
    elif args and args[-1] == "--reaudit":
        reaudit = True
        args.pop()
    if args and args[0] == "--file" and len(args) == 2 and not args[1].startswith("--"):
        body = Path(args[1]).read_text(encoding="utf-8")
    elif args:
        usage()
    elif sys.stdin.isatty():
        usage()
    else:
        body = sys.stdin.read()
    check(body, reaudit=reaudit)


if __name__ == "__main__":
    main()
