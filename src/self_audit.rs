//! PR body Self-Audit checker (Issue #2708; the rules of #2279 scope 3, #2385, #2388 AC-3b
//! and #2403 item 5), moved from `tools/check_pr_self_audit.py` so a user project needs only
//! `moai-cli`. It judges text and never posts anything.
//!
//! A lane PR body carries a `## Self-Audit` section with four measured lines — `mutation:`
//! (and the mutation reports FAILED), `clone:`, `host:`, `receipt:` — and the round's
//! `plan:` (the lane's PLAN comment, by an id of at least seven digits) and `red:` (the test
//! that failed first, or why there is none) lines (#2880). An optional `ci:` line is
//! either `ci: … run=<digits> head=<40 lowercase hex>` (keyed whole tokens) or `ci: not_run
//! <reason>`. In the section and in every Markdown table row, `0 passed` needs `not_run` on the
//! same line, and a filtered `cargo test` in a code span must select exactly one target. With
//! `reaudit`, the one `## 독립 review` section needs one visible `수리 확인 범위:` and one `새
//! 지적 분류:` line (`없음`, `차단: <reason>` or `후속 Issue #<n>: <reason>`), outside HTML
//! comments, fences and indented code.
//!
//! The outcome is the Python tool's, line for line: `self-audit: pass (mutation/clone/host/
//! receipt measured)` or `self-audit: blocked missing=<a,b>`. No regular-expression crate is
//! used; each pattern of the Python tool is a hand-written matcher with the same meaning.

const SECTION: &str = "## self-audit";
/// The measured lines, then the lane round's PLAN comment and RED evidence (#2880 W1·W2).
const TAGS: [&str; 6] = ["mutation", "clone", "host", "receipt", "plan", "red"];
/// The shortest number taken as a comment id on a `plan:` line: GitHub comment ids run to ten
/// digits, Issue and pull request numbers here to four.
const PLAN_ID_DIGITS: usize = 7;
pub const PASS: &str = "self-audit: pass (mutation/clone/host/receipt measured)";

/// `Ok(())` on pass, `Err(missing)` with the names the Python tool prints.
pub fn check(body: &str, reaudit: bool) -> Result<(), Vec<String>> {
    let lines: Vec<&str> = python_splitlines(body);
    let Some(start) = lines
        .iter()
        .position(|l| py_strip(l).to_lowercase() == SECTION)
    else {
        return Err(vec![SECTION.into()]);
    };
    let section: Vec<&str> = lines[start + 1..]
        .iter()
        .take_while(|l| !l.starts_with('#'))
        .copied()
        .collect();
    let mut missing = Vec::new();
    for tag in TAGS {
        let prefix = format!("{tag}:");
        let hit = section
            .iter()
            .find(|l| strip_marker(l).starts_with(&prefix));
        let filled = hit.is_some_and(|l| {
            l.split_once(':').is_some_and(|(_, v)| {
                // `plan:` names the PLAN comment by its id: a run of at least seven digits.
                !py_strip(v).is_empty()
                    && (tag != "plan"
                        || v.split(|c: char| !c.is_ascii_digit())
                            .any(|run| run.len() >= PLAN_ID_DIGITS))
            })
        });
        if !filled {
            missing.push(tag.to_string());
        }
    }
    if !missing.is_empty() {
        return Err(missing);
    }
    check_zero_selected(&section, &lines)?;
    let mutation = section
        .iter()
        .find(|l| strip_marker(l).starts_with("mutation:"))
        .expect("checked above");
    if !mutation.to_lowercase().contains("failed") {
        return Err(vec!["mutation:FAILED".into()]);
    }
    check_ci_lines(&section)?;
    if reaudit {
        check_reaudit(&lines)?;
    }
    Ok(())
}

/// The line the tool prints for a result.
pub fn render(result: &Result<(), Vec<String>>) -> String {
    match result {
        Ok(()) => PASS.to_string(),
        Err(missing) => format!("self-audit: blocked missing={}", missing.join(",")),
    }
}

// ---- Python string semantics ----

/// `str.splitlines()`: splits on every Unicode line boundary Python knows, drops them.
fn python_splitlines(s: &str) -> Vec<&str> {
    let mut out = Vec::new();
    let mut start = 0;
    let bytes: Vec<(usize, char)> = s.char_indices().collect();
    let mut i = 0;
    while i < bytes.len() {
        let (pos, c) = bytes[i];
        let boundary = matches!(
            c,
            '\n' | '\r'
                | '\u{0b}'
                | '\u{0c}'
                | '\u{1c}'
                | '\u{1d}'
                | '\u{1e}'
                | '\u{85}'
                | '\u{2028}'
                | '\u{2029}'
        );
        if boundary {
            out.push(&s[start..pos]);
            let mut next = pos + c.len_utf8();
            if c == '\r' && bytes.get(i + 1).is_some_and(|(_, n)| *n == '\n') {
                next += 1;
                i += 1;
            }
            start = next;
        }
        i += 1;
    }
    if start < s.len() {
        out.push(&s[start..]);
    }
    out
}

/// Python's `str.isspace()` for one character.
fn py_space(c: char) -> bool {
    c.is_whitespace() || matches!(c, '\u{1c}'..='\u{1f}')
}

fn py_strip(s: &str) -> &str {
    s.trim_matches(py_space)
}

/// Python `\w` (Unicode): alphanumeric or underscore.
fn word(c: char) -> bool {
    c.is_alphanumeric() || c == '_'
}

/// `line.strip().lstrip("-*").strip().lower()`.
fn strip_marker(line: &str) -> String {
    py_strip(py_strip(line).trim_start_matches(['-', '*'])).to_lowercase()
}

// ---- ci: lines ----

fn ci_tokens_ok(value: &str) -> Vec<String> {
    let tokens: Vec<&str> = value.split(py_space).filter(|t| !t.is_empty()).collect();
    let run = tokens.iter().any(|t| {
        t.strip_prefix("run=")
            .is_some_and(|d| !d.is_empty() && d.bytes().all(|b| b.is_ascii_digit()))
    });
    let head = tokens.iter().any(|t| {
        t.strip_prefix("head=").is_some_and(|h| {
            h.len() == 40
                && h.bytes()
                    .all(|b| b.is_ascii_digit() || (b'a'..=b'f').contains(&b))
        })
    });
    let mut missing = Vec::new();
    if !run {
        missing.push("ci:run".into());
    }
    if !head {
        missing.push("ci:head".into());
    }
    missing
}

/// `^not_run\b(?=.*\w)`.
fn ci_not_run(value: &str) -> bool {
    let Some(rest) = value.strip_prefix("not_run") else {
        return false;
    };
    let mut chars = rest.chars();
    let boundary = chars.clone().next().is_none_or(|c| !word(c));
    boundary && chars.any(word)
}

fn check_ci_lines(section: &[&str]) -> Result<(), Vec<String>> {
    for line in section {
        if !strip_marker(line).starts_with("ci:") {
            continue;
        }
        let value = py_strip(line.split_once(':').map_or("", |(_, v)| v));
        if ci_not_run(value) {
            continue;
        }
        let missing = ci_tokens_ok(value);
        if !missing.is_empty() {
            return Err(missing);
        }
    }
    Ok(())
}

// ---- zero-selected evidence ----

/// `(?<![0-9])0 passed\b`.
fn zero_passed(line: &str) -> bool {
    let chars: Vec<char> = line.chars().collect();
    let pat: Vec<char> = "0 passed".chars().collect();
    (0..chars.len()).any(|i| {
        chars[i..].starts_with(&pat)
            && (i == 0 || !chars[i - 1].is_ascii_digit())
            && chars.get(i + pat.len()).is_none_or(|c| !word(*c))
    })
}

const CARGO_VALUE_OPTS: [&str; 19] = [
    "--test",
    "--bin",
    "--example",
    "--bench",
    "-p",
    "--package",
    "-F",
    "--features",
    "--manifest-path",
    "--target",
    "--target-dir",
    "-j",
    "--jobs",
    "--profile",
    "--color",
    "--message-format",
    "--config",
    "-Z",
    "--exclude",
];
const CARGO_TARGETS_ONE: [&str; 6] = ["--lib", "--doc", "--test", "--bin", "--example", "--bench"];
const CARGO_TARGETS_MANY: [&str; 5] = [
    "--tests",
    "--bins",
    "--examples",
    "--benches",
    "--all-targets",
];
const HARNESS_VALUE_OPTS: [&str; 6] = [
    "--skip",
    "--test-threads",
    "--format",
    "--color",
    "-Z",
    "--logfile",
];
const CARGO_GLOBAL_FLAGS: [&str; 7] = [
    "--offline",
    "--locked",
    "--frozen",
    "-q",
    "--quiet",
    "-v",
    "--verbose",
];
const CARGO_GLOBAL_VALUES: [&str; 4] = ["--color", "--config", "-Z", "-C"];

#[derive(Clone, Copy, PartialEq, Eq)]
enum Kind {
    Word,
    Redirect,
    Separator,
}

/// `` `([^`]+)` `` matches of a line, non-overlapping, left to right.
fn code_spans(line: &str) -> Vec<&str> {
    let mut out = Vec::new();
    let mut rest = line;
    while let Some(open) = rest.find('`') {
        let after = &rest[open + 1..];
        match after.find('`') {
            Some(0) => rest = after,
            Some(close) => {
                out.push(&after[..close]);
                rest = &after[close + 1..];
            }
            None => break,
        }
    }
    out
}

/// `(?<![\w'])cargo(?=\s|$)` somewhere in the span.
fn mentions_cargo(span: &str) -> bool {
    let chars: Vec<char> = span.chars().collect();
    let pat: Vec<char> = "cargo".chars().collect();
    (0..chars.len()).any(|i| {
        chars[i..].starts_with(&pat)
            && (i == 0 || !(word(chars[i - 1]) || chars[i - 1] == '\''))
            && chars.get(i + 5).is_none_or(|c| py_space(*c))
    })
}

/// One `SHELL_TOKEN` match at `chars[i..]`: (kind, length), or None when nothing matches.
fn shell_token(chars: &[char], i: usize) -> Option<(Kind, usize)> {
    // redirect: \d*(?:<<<|<<|<>|>>|>\||>&|<&|>|<), and &>> / &> first.
    let rest = &chars[i..];
    let starts = |p: &str| {
        let p: Vec<char> = p.chars().collect();
        rest.starts_with(&p)
    };
    if starts("&>>") {
        return Some((Kind::Redirect, 3));
    }
    if starts("&>") {
        return Some((Kind::Redirect, 2));
    }
    let digits = rest.iter().take_while(|c| c.is_ascii_digit()).count();
    // `\d*` backtracks: any digit prefix length works as long as an operator follows the
    // digits actually taken; operators never start with a digit, so only the full run counts.
    let ops = ["<<<", "<<", "<>", ">>", ">|", ">&", "<&", ">", "<"];
    for op in ops {
        let op: Vec<char> = op.chars().collect();
        if rest[digits..].starts_with(&op) {
            return Some((Kind::Redirect, digits + op.len()));
        }
    }
    // separator: &&|\|\||[;&|()]
    if starts("&&") || starts("||") {
        return Some((Kind::Separator, 2));
    }
    if matches!(rest.first(), Some(';' | '&' | '|' | '(' | ')')) {
        return Some((Kind::Separator, 1));
    }
    // word: (?:[^\s;&|()<>'"\\]|\\.|'[^']*'|"(?:\\.|[^"\\])*")+
    let mut j = 0;
    while let Some(&c) = rest.get(j) {
        if c == '\\' {
            match rest.get(j + 1) {
                Some(n) if *n != '\n' => j += 2,
                _ => break,
            }
        } else if c == '\'' {
            match rest[j + 1..].iter().position(|x| *x == '\'') {
                Some(k) => j += k + 2,
                None => break,
            }
        } else if c == '"' {
            let mut k = j + 1;
            let mut closed = false;
            while let Some(&d) = rest.get(k) {
                if d == '\\' {
                    if rest.get(k + 1).is_some_and(|n| *n != '\n') {
                        k += 2;
                        continue;
                    }
                    break;
                }
                if d == '"' {
                    closed = true;
                    break;
                }
                k += 1;
            }
            if closed {
                j = k + 1;
            } else {
                break;
            }
        } else if py_space(c) || matches!(c, ';' | '&' | '|' | '(' | ')' | '<' | '>') {
            break;
        } else {
            j += 1;
        }
    }
    (j > 0).then_some((Kind::Word, j))
}

/// `shlex.split(raw, posix=True)[0]` for one word token: quotes removed, escapes resolved.
fn shlex_word(raw: &[char]) -> String {
    let mut out = String::new();
    let mut i = 0;
    while i < raw.len() {
        let c = raw[i];
        match c {
            '\\' => {
                out.push(raw[i + 1]);
                i += 2;
            }
            '\'' => {
                let end = i + 1 + raw[i + 1..].iter().position(|x| *x == '\'').unwrap();
                out.extend(&raw[i + 1..end]);
                i = end + 1;
            }
            '"' => {
                let mut k = i + 1;
                while raw[k] != '"' {
                    if raw[k] == '\\' && matches!(raw[k + 1], '\\' | '"') {
                        out.push(raw[k + 1]);
                        k += 2;
                    } else {
                        out.push(raw[k]);
                        k += 1;
                    }
                }
                i = k + 1;
            }
            _ => {
                out.push(c);
                i += 1;
            }
        }
    }
    out
}

/// The token lists of the `cargo test` commands written in code spans on one line.
fn cargo_test_commands(line: &str) -> Vec<Vec<String>> {
    let mut out = Vec::new();
    for span in code_spans(line) {
        if !mentions_cargo(span) {
            continue;
        }
        let chars: Vec<char> = span.chars().collect();
        let mut words: Vec<(String, Kind)> = Vec::new();
        let mut i = 0;
        let mut malformed = false;
        while i < chars.len() {
            if py_space(chars[i]) {
                i += 1;
                continue;
            }
            match shell_token(&chars, i) {
                Some((kind, len)) => {
                    let raw = &chars[i..i + len];
                    let text = if kind == Kind::Word {
                        shlex_word(raw)
                    } else {
                        raw.iter().collect()
                    };
                    words.push((text, kind));
                    i += len;
                }
                None => {
                    malformed = true;
                    break;
                }
            }
        }
        if malformed {
            continue;
        }
        let mut i = 0;
        while i < words.len() {
            if words[i] != ("cargo".to_string(), Kind::Word) {
                i += 1;
                continue;
            }
            let mut test = i + 1;
            while test < words.len() {
                let (token, kind) = (&words[test].0, words[test].1);
                if kind == Kind::Redirect {
                    test += 1;
                    if test < words.len() && words[test].1 == Kind::Word {
                        test += 1;
                    }
                    continue;
                }
                if kind != Kind::Word {
                    break;
                }
                let name = token.split('=').next().unwrap_or("");
                if token.starts_with('+') || CARGO_GLOBAL_FLAGS.contains(&token.as_str()) {
                    test += 1;
                } else if CARGO_GLOBAL_VALUES.contains(&name) {
                    test += if token.contains('=') { 1 } else { 2 };
                } else if (token.len() >= 2
                    && token.starts_with('-')
                    && token[1..].bytes().all(|b| b == b'q' || b == b'v'))
                    || ((token.starts_with("-Z") || token.starts_with("-C")) && token.len() > 2)
                {
                    test += 1;
                } else {
                    break;
                }
            }
            if test >= words.len() || words[test] != ("test".to_string(), Kind::Word) {
                i += 1;
                continue;
            }
            let mut end = test + 1;
            let mut tokens = Vec::new();
            while end < words.len() {
                let (token, kind) = (&words[end].0, words[end].1);
                if kind == Kind::Separator {
                    break;
                }
                if kind == Kind::Redirect {
                    end += 1;
                    if end < words.len() && words[end].1 == Kind::Word {
                        end += 1;
                    }
                    continue;
                }
                tokens.push(token.clone());
                end += 1;
            }
            out.push(tokens);
            i = end;
        }
    }
    out
}

fn glob(s: &str) -> bool {
    s.contains(['*', '?', '['])
}

/// (has a name filter, targets selected; None = cargo's default set, more than one).
fn filtered_target_count(tokens: &[String]) -> (bool, Option<usize>) {
    let (mut has_filter, mut targets, mut many, mut after_dashdash) = (false, 0, false, false);
    let mut skip_next: Option<String> = None;
    for tok in tokens {
        if let Some(skipped) = skip_next.take() {
            if CARGO_TARGETS_ONE.contains(&skipped.as_str()) && glob(tok) {
                many = true;
            }
            continue;
        }
        if after_dashdash {
            if HARNESS_VALUE_OPTS.contains(&tok.as_str()) {
                skip_next = Some(tok.clone());
            } else if !tok.starts_with('-') {
                has_filter = true;
            }
            continue;
        }
        if tok == "--" {
            after_dashdash = true;
            continue;
        }
        let name = tok.split('=').next().unwrap_or("");
        if CARGO_TARGETS_ONE.contains(&name) {
            targets += 1;
            if let Some((_, value)) = tok.split_once('=') {
                if glob(value) {
                    many = true;
                }
            }
        } else if CARGO_TARGETS_MANY.contains(&name) {
            many = true;
        }
        if CARGO_VALUE_OPTS.contains(&name) && !tok.contains('=') {
            skip_next = Some(name.to_string());
        } else if !tok.starts_with('-') {
            has_filter = true;
        }
    }
    if many {
        return (has_filter, Some(2));
    }
    (has_filter, (targets > 0).then_some(targets))
}

/// `line.strip().strip("|").split("|")`.
fn cells(line: &str) -> Vec<&str> {
    py_strip(line).trim_matches('|').split('|').collect()
}

/// `\s*:?-{3,}:?\s*` as a full match.
fn delimiter_cell(cell: &str) -> bool {
    let c = py_strip(cell);
    let c = c.strip_prefix(':').unwrap_or(c);
    let c = c.strip_suffix(':').unwrap_or(c);
    c.len() >= 3 && c.bytes().all(|b| b == b'-')
}

fn check_zero_selected(section: &[&str], lines: &[&str]) -> Result<(), Vec<String>> {
    let mut scope: Vec<&str> = section.to_vec();
    scope.extend(
        lines
            .iter()
            .filter(|l| l.trim_start_matches(py_space).starts_with('|')),
    );
    for index in 1..lines.len() {
        let delimiter = cells(lines[index]);
        if delimiter.len() < 2
            || !delimiter.iter().all(|c| delimiter_cell(c))
            || !lines[index - 1].contains('|')
            || cells(lines[index - 1]).len() != delimiter.len()
        {
            continue;
        }
        scope.push(lines[index - 1]);
        for row in &lines[index + 1..] {
            if !row.contains('|') {
                break;
            }
            scope.push(row);
        }
    }
    for line in scope {
        if zero_passed(line) && !line.contains("not_run") {
            return Err(vec!["zero-selected".into()]);
        }
        for tokens in cargo_test_commands(line) {
            let (has_filter, count) = filtered_target_count(&tokens);
            if has_filter && count.is_none_or(|n| n > 1) {
                return Err(vec!["filter-multi-target".into()]);
            }
        }
    }
    Ok(())
}

// ---- re-audit ----

/// Removes `` (`+)(?!`)(.+?)(?<!`)\1(?!`) `` matches (inline code), left to right.
fn strip_inline_code(value: &str) -> String {
    let chars: Vec<char> = value.chars().collect();
    let mut out = String::new();
    let mut i = 0;
    'scan: while i < chars.len() {
        if chars[i] == '`' {
            let n = chars[i..].iter().take_while(|c| **c == '`').count();
            let open_end = i + n;
            if open_end < chars.len() {
                // The lazy body: the first closing run of exactly n that is preceded by a
                // non-backtick and not followed by a backtick.
                let mut j = open_end + 1;
                while j + n <= chars.len() {
                    if chars[j - 1] != '`'
                        && chars[j..j + n].iter().all(|c| *c == '`')
                        && chars.get(j + n).is_none_or(|c| *c != '`')
                    {
                        i = j + n;
                        continue 'scan;
                    }
                    j += 1;
                }
            }
        }
        out.push(chars[i]);
        i += 1;
    }
    out
}

/// `(?<![\w./\\-])(?:TODO|TBD|not_run)(?![\w./\\-])`, case-insensitive.
fn unfinished_marker(prose: &str) -> bool {
    let chars: Vec<char> = prose.chars().collect();
    let edge = |c: char| word(c) || matches!(c, '.' | '/' | '\\' | '-');
    for marker in ["todo", "tbd", "not_run"] {
        let m: Vec<char> = marker.chars().collect();
        for i in 0..chars.len() {
            let window = chars.get(i..i + m.len());
            let hit = window.is_some_and(|w| {
                w.iter()
                    .zip(&m)
                    .all(|(a, b)| a.to_lowercase().eq(b.to_lowercase()))
            });
            if hit
                && (i == 0 || !edge(chars[i - 1]))
                && chars.get(i + m.len()).is_none_or(|c| !edge(*c))
            {
                return true;
            }
        }
    }
    false
}

/// `<[^\s<>]+(?:[ \t]+[^\s<>]+)*>` somewhere in the value.
fn angle_placeholder(value: &str) -> bool {
    let chars: Vec<char> = value.chars().collect();
    for i in 0..chars.len() {
        if chars[i] != '<' {
            continue;
        }
        let mut j = i + 1;
        while j < chars.len() {
            let c = chars[j];
            if c == '<' || c == '>' || (py_space(c) && c != ' ' && c != '\t') {
                break;
            }
            j += 1;
        }
        if j < chars.len() && chars[j] == '>' && j > i + 1 {
            let body = &chars[i + 1..j];
            let blank = |c: &char| *c == ' ' || *c == '\t';
            if !blank(&body[0]) && !blank(&body[body.len() - 1]) {
                return true;
            }
        }
    }
    false
}

/// `\s*[`*_]*\s*(?:TODO|TBD|not_run)[\s`*_.-]*` as a full match, case-insensitive.
fn only_unfinished(value: &str) -> bool {
    let v = value.trim_start_matches(py_space);
    let v = v.trim_start_matches(['`', '*', '_']);
    let v = v.trim_start_matches(py_space);
    for marker in ["todo", "tbd", "not_run"] {
        let n = marker.chars().count();
        let head: String = v.chars().take(n).collect();
        if head.to_lowercase() == marker {
            let rest: String = v.chars().skip(n).collect();
            if rest
                .chars()
                .all(|c| py_space(c) || matches!(c, '`' | '*' | '_' | '.' | '-'))
            {
                return true;
            }
        }
    }
    false
}

fn meaningful(value: &str) -> bool {
    let prose = strip_inline_code(value);
    value.chars().any(char::is_alphanumeric)
        && !unfinished_marker(&prose)
        && !angle_placeholder(value)
        && !only_unfinished(value)
}

/// Lines with HTML comments, fenced blocks and indented code removed.
fn visible_reaudit_lines(lines: &[&str]) -> Vec<String> {
    let mut out = Vec::new();
    let mut in_comment = false;
    let mut fence: Option<(char, usize)> = None;
    for line in lines {
        if let Some((ch, n)) = fence {
            // ` {0,3}` + fence char{n,} + `[ \t]*`, full match.
            let spaces = line.chars().take_while(|c| *c == ' ').count();
            let rest: String = line.chars().skip(spaces).collect();
            let run = rest.chars().take_while(|c| *c == ch).count();
            let tail: String = rest.chars().skip(run).collect();
            if spaces <= 3 && run >= n && tail.chars().all(|c| c == ' ' || c == '\t') {
                fence = None;
            }
            continue;
        }
        let tab_marker = {
            let mut c = line.chars();
            c.next() == Some('\t')
                && matches!(c.next(), Some('-' | '*' | '+'))
                && c.next().is_some_and(py_space)
        };
        if !in_comment && (line.starts_with("    ") || line.starts_with('\t')) && !tab_marker {
            continue;
        }
        let chars: Vec<char> = line.chars().collect();
        let mut visible = String::new();
        let mut position = 0;
        while position < chars.len() {
            if in_comment {
                match find_seq(&chars, position, &['-', '-', '>']) {
                    Some(end) => {
                        in_comment = false;
                        position = end + 3;
                    }
                    None => break,
                }
                continue;
            }
            // \\[`\\]|`+|<!--
            let mut k = position;
            let mut token: Option<(usize, usize, &str)> = None;
            while k < chars.len() {
                if chars[k] == '\\' && matches!(chars.get(k + 1), Some('`' | '\\')) {
                    token = Some((k, k + 2, "esc"));
                    break;
                }
                if chars[k] == '`' {
                    let n = chars[k..].iter().take_while(|c| **c == '`').count();
                    token = Some((k, k + n, "tick"));
                    break;
                }
                if chars[k..].starts_with(&['<', '!', '-', '-']) {
                    token = Some((k, k + 4, "comment"));
                    break;
                }
                k += 1;
            }
            let Some((start, end, kind)) = token else {
                visible.extend(&chars[position..]);
                break;
            };
            visible.extend(&chars[position..start]);
            match kind {
                "esc" => {
                    visible.extend(&chars[start..end]);
                    position = end;
                }
                "comment" => {
                    in_comment = true;
                    position = end;
                }
                _ => {
                    let n = end - start;
                    let rest = &chars[end..];
                    let mut closer = None;
                    for p in 0..rest.len() {
                        if rest[p..].len() >= n
                            && rest[p..p + n].iter().all(|c| *c == '`')
                            && (p == 0 || rest[p - 1] != '`')
                            && rest.get(p + n).is_none_or(|c| *c != '`')
                        {
                            closer = Some(p + n);
                            break;
                        }
                    }
                    position = match closer {
                        None => end,
                        Some(c) => end + c,
                    };
                    visible.extend(&chars[start..position]);
                }
            }
        }
        // ^ {0,3}(`{3,}|~{3,})(.*)$
        let spaces = visible.chars().take_while(|c| *c == ' ').count();
        let after: String = visible.chars().skip(spaces).collect();
        let first = after.chars().next();
        if spaces <= 3 && matches!(first, Some('`' | '~')) {
            let ch = first.unwrap();
            let run = after.chars().take_while(|c| *c == ch).count();
            if run >= 3 {
                let info: String = after.chars().skip(run).collect();
                if ch != '`' || !info.contains('`') {
                    fence = Some((ch, run));
                    continue;
                }
            }
        }
        out.push(visible);
    }
    out
}

fn find_seq(chars: &[char], from: usize, seq: &[char]) -> Option<usize> {
    (from..chars.len()).find(|&i| chars[i..].starts_with(seq))
}

/// `^\s*#{1,6}(?:\s|$)`.
fn heading(line: &str) -> bool {
    let l = line.trim_start_matches(py_space);
    let n = l.chars().take_while(|c| *c == '#').count();
    (1..=6).contains(&n) && l.chars().nth(n).is_none_or(py_space)
}

/// `\s*(?:(?:[-*+]|[0-9]+[.)])\s+)?(수리 확인 범위|새 지적 분류)\s*:\s*(.*)`, full match.
fn review_field(line: &str) -> Option<(&'static str, String)> {
    let l = line.trim_start_matches(py_space);
    let try_field = |s: &str| -> Option<(&'static str, String)> {
        for name in ["수리 확인 범위", "새 지적 분류"] {
            if let Some(rest) = s.strip_prefix(name) {
                let rest = rest.trim_start_matches(py_space);
                if let Some(value) = rest.strip_prefix(':') {
                    return Some((name, value.trim_start_matches(py_space).to_string()));
                }
            }
        }
        None
    };
    // With a list marker (then at least one whitespace), or without one.
    let marker_len = if l.starts_with(['-', '*', '+']) {
        Some(1)
    } else {
        let digits = l.bytes().take_while(u8::is_ascii_digit).count();
        (digits > 0 && matches!(l.as_bytes().get(digits), Some(b'.' | b')'))).then_some(digits + 1)
    };
    if let Some(m) = marker_len {
        let after = &l[m..];
        if after.starts_with(py_space) {
            if let Some(hit) = try_field(after.trim_start_matches(py_space)) {
                return Some(hit);
            }
        }
    }
    try_field(l)
}

fn check_reaudit(lines: &[&str]) -> Result<(), Vec<String>> {
    let lines = visible_reaudit_lines(lines);
    let starts: Vec<usize> = lines
        .iter()
        .enumerate()
        .filter(|(_, l)| py_strip(l) == "## 독립 review")
        .map(|(i, _)| i)
        .collect();
    if starts.len() != 1 {
        return Err(vec!["## 독립 review".into()]);
    }
    let (mut scope, mut findings) = (Vec::new(), Vec::new());
    for line in &lines[starts[0] + 1..] {
        if heading(line) {
            break;
        }
        if let Some((name, value)) = review_field(line) {
            let value = py_strip(&value).to_string();
            if name == "수리 확인 범위" {
                scope.push(value);
            } else {
                findings.push(value);
            }
        }
    }
    let mut missing = Vec::new();
    if scope.len() != 1 || !meaningful(&scope[0]) {
        missing.push("reaudit:scope".to_string());
    }
    let valid = findings.len() == 1 && meaningful(&findings[0]) && {
        let f = &findings[0];
        if f == "없음" {
            true
        } else {
            classified_reason(f).is_some_and(meaningful)
        }
    };
    if !valid {
        missing.push("reaudit:findings".into());
    }
    if missing.is_empty() {
        Ok(())
    } else {
        Err(missing)
    }
}

/// `(?:차단|후속\s+Issue\s+#[1-9][0-9]*)\s*:\s*(.*)` as a full match: the reason.
fn classified_reason(f: &str) -> Option<&str> {
    let rest = if let Some(r) = f.strip_prefix("차단") {
        r
    } else {
        let r = f.strip_prefix("후속")?;
        let r2 = r.trim_start_matches(py_space);
        if r2.len() == r.len() {
            return None;
        }
        let r = r2.strip_prefix("Issue")?;
        let r2 = r.trim_start_matches(py_space);
        if r2.len() == r.len() {
            return None;
        }
        let r = r2.strip_prefix('#')?;
        let first = r.chars().next()?;
        if !('1'..='9').contains(&first) {
            return None;
        }
        let digits = r.bytes().take_while(u8::is_ascii_digit).count();
        &r[digits..]
    };
    let rest = rest.trim_start_matches(py_space);
    let value = rest.strip_prefix(':')?;
    Some(value.trim_start_matches(py_space))
}

#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    fn splitlines_follows_python() {
        assert_eq!(python_splitlines("a\r\nb\rc\n"), vec!["a", "b", "c"]);
        assert_eq!(python_splitlines("a\u{2028}b"), vec!["a", "b"]);
        assert_eq!(python_splitlines(""), Vec::<&str>::new());
    }

    #[test]
    fn inline_code_strip_matches_the_python_pattern() {
        assert_eq!(strip_inline_code("a `TODO` b"), "a  b");
        assert_eq!(strip_inline_code("``b`"), "`");
        assert_eq!(strip_inline_code("x ``a`b`` y"), "x  y");
    }

    #[test]
    fn shell_words_unquote_like_shlex() {
        let w: Vec<char> = r#"a"b\"c"'d'\ e"#.chars().collect();
        assert_eq!(shlex_word(&w), "ab\"cd e");
    }
}
