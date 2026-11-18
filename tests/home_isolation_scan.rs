//! Issue #3068 P2 (requirement 9, AC8): two scans of the checked-in tree that keep the local
//! ledger away from the real HOME and from the network, and the check of the `TempHome` helper
//! they rely on.
//!
//! Scan A. A test in the ledger area (`tests/ledger*`) runs the `moai` binary only through
//! `TempHome` (`TempHome::command`, `TempHome::command_for`): never `MOAI_BIN`, `MOAI_CLI_BIN`,
//! `CARGO_BIN_EXE_*` or `common::cli` directly. Two older files carry the name for another
//! reason (Issue #2384, the forge-only work ledger) and use a plain temporary HOME; they are
//! named in `LEGACY_LEDGER_NAMED`, a list that must not grow. Beyond the ledger area, a test file
//! that runs the binary and never mentions HOME (or `TempHome`) in code can only be one of those
//! in the baseline file; that file may shrink and a new name in it fails the scan.
//!
//! Scan B. Nothing under `src/ledger/` spawns a process (`Command::new(`), opens a socket
//! (`std::net`), names a program (`"gh"`, `"git"`, `"curl"`, `"ssh"`), holds an `http(s)://`
//! endpoint or reaches into `crate::forge`. Comments are not code and are ignored. The same
//! holds for `src/decide/` (cards C-37 and C-54), except the two provider endpoint constants
//! and, in the one file that sends (`src/decide/transport.rs`), the four socket names.
//!
//! Both scans are text matches: they show a name is absent, not that a path is never reached.
//! The entry-count tool `tools/home_entry_count.py` measures the real `~/.moai` before and
//! after a verification run, which is the dynamic half of AC8 and is not run by any test.

mod common;

use common::home::TempHome;
use std::collections::{BTreeMap, BTreeSet};
use std::fs;
use std::path::{Path, PathBuf};

const ROOT: &str = env!("CARGO_MANIFEST_DIR");

/// Test files that run the binary and mention no HOME in code. Names only; computed by the scan
/// and checked against this file, which may shrink and never grows.
const BASELINE: &str = "tests/fixtures/home_isolation/binary_without_home.txt";

/// Ledger-named test files that predate the ledger and use their own temporary HOME (#2384).
/// Adding a name here is a change to this scan, visible in review; the count is pinned below.
const LEGACY_LEDGER_NAMED: [&str; 2] = [
    "tests/ledger_forge_only_cli.rs",
    "tests/ledger_forge_unconnected_cli.rs",
];

/// A name searched for in code. `word_start` / `word_end` require that the character before or
/// after is not part of an identifier.
struct Tok {
    text: &'static str,
    word_start: bool,
    word_end: bool,
}

const fn word(text: &'static str) -> Tok {
    Tok {
        text,
        word_start: true,
        word_end: true,
    }
}

const fn start(text: &'static str) -> Tok {
    Tok {
        text,
        word_start: true,
        word_end: false,
    }
}

const fn raw(text: &'static str) -> Tok {
    Tok {
        text,
        word_start: false,
        word_end: false,
    }
}

/// How a test reaches the binary without `TempHome`.
const BINARY_TOKENS: [Tok; 4] = [
    word("MOAI_BIN"),
    word("MOAI_CLI_BIN"),
    start("CARGO_BIN_EXE_"),
    word("common::cli"),
];

/// What the ledger library must not contain.
const EXTERNAL_TOKENS: [Tok; 15] = [
    start("Command::new("),
    word("process::Command"),
    word("std::net"),
    start("net::"),
    word("TcpStream"),
    word("TcpListener"),
    word("UdpSocket"),
    raw("http://"),
    raw("https://"),
    raw("\"gh\""),
    raw("\"git\""),
    raw("\"curl\""),
    raw("\"ssh\""),
    word("crate::forge"),
    word("moai_cli::forge"),
];

fn is_ident(c: char) -> bool {
    c.is_alphanumeric() || c == '_'
}

/// `text` with comments blanked: every character of a `//` or `/* */` comment becomes a space
/// (newlines stay, so line numbers hold). String, raw string and character literals are copied
/// as they are, so a `//` inside `"https://x"` does not start a comment and a URL in a string is
/// still code.
fn blank_comments(text: &str) -> String {
    let b: Vec<char> = text.chars().collect();
    let mut out = String::with_capacity(text.len());
    let mut i = 0;
    while i < b.len() {
        let c = b[i];
        let next = b.get(i + 1).copied();
        if c == '/' && next == Some('/') {
            while i < b.len() && b[i] != '\n' {
                out.push(' ');
                i += 1;
            }
        } else if c == '/' && next == Some('*') {
            let mut depth = 0;
            while i < b.len() {
                if b[i] == '/' && b.get(i + 1) == Some(&'*') {
                    depth += 1;
                    out.push_str("  ");
                    i += 2;
                } else if b[i] == '*' && b.get(i + 1) == Some(&'/') {
                    depth -= 1;
                    out.push_str("  ");
                    i += 2;
                    if depth == 0 {
                        break;
                    }
                } else {
                    out.push(if b[i] == '\n' { '\n' } else { ' ' });
                    i += 1;
                }
            }
        } else if c == 'r' && (i == 0 || !is_ident(b[i - 1])) && raw_string_end(&b, i).is_some() {
            let end = raw_string_end(&b, i).unwrap();
            out.extend(&b[i..end]);
            i = end;
        } else if c == '"' {
            out.push(c);
            i += 1;
            while i < b.len() {
                out.push(b[i]);
                if b[i] == '\\' && i + 1 < b.len() {
                    out.push(b[i + 1]);
                    i += 2;
                    continue;
                }
                i += 1;
                if b[i - 1] == '"' {
                    break;
                }
            }
        } else if c == '\'' && next == Some('\\') {
            // an escaped character literal: copy up to the closing quote
            out.push(c);
            i += 1;
            while i < b.len() {
                out.push(b[i]);
                i += 1;
                if b[i - 1] == '\'' && b[i - 2] != '\\' {
                    break;
                }
            }
        } else if c == '\'' && b.get(i + 2) == Some(&'\'') {
            out.extend(&b[i..i + 3]);
            i += 3;
        } else {
            out.push(c);
            i += 1;
        }
    }
    out
}

/// Where the raw string starting at `b[at] == 'r'` ends (exclusive), or `None` when it is not
/// one (`r"..."`, `r#"..."#`, ...).
fn raw_string_end(b: &[char], at: usize) -> Option<usize> {
    let mut j = at + 1;
    let mut hashes = 0;
    while b.get(j) == Some(&'#') {
        hashes += 1;
        j += 1;
    }
    if b.get(j) != Some(&'"') {
        return None;
    }
    j += 1;
    while j < b.len() {
        if b[j] == '"' && (0..hashes).all(|k| b.get(j + 1 + k) == Some(&'#')) {
            return Some(j + 1 + hashes);
        }
        j += 1;
    }
    Some(b.len())
}

/// The characters of `code` with every run of whitespace reduced to one space, and that space
/// dropped next to `:`, `(`, `)`, `,`, `;` and `.`, so a call rustfmt wrapped onto the next
/// line (`Command\n    ::new(`) still reads as `Command::new(`. Each character keeps its line.
fn normalize(code: &str) -> (Vec<char>, Vec<usize>) {
    let (mut chars, mut lines) = (Vec::new(), Vec::new());
    let (mut line, mut gap): (usize, Option<usize>) = (1, None);
    for c in code.chars() {
        if c.is_whitespace() {
            if gap.is_none() {
                gap = Some(line);
            }
            if c == '\n' {
                line += 1;
            }
            continue;
        }
        if let Some(gap_line) = gap.take() {
            let tight_before = chars.last().is_none_or(|p| ":(.".contains(*p));
            if !tight_before && !":(),;.".contains(c) {
                chars.push(' ');
                lines.push(gap_line);
            }
        }
        chars.push(c);
        lines.push(line);
    }
    (chars, lines)
}

/// Every place one of `toks` is written in the code of `text`, as (line, name), sorted.
fn find(text: &str, toks: &[Tok]) -> Vec<(usize, &'static str)> {
    let (chars, lines) = normalize(&blank_comments(text));
    let mut hits = Vec::new();
    for t in toks {
        let pat: Vec<char> = t.text.chars().collect();
        let mut i = 0;
        while i + pat.len() <= chars.len() {
            if chars[i..i + pat.len()] != pat[..] {
                i += 1;
                continue;
            }
            let before_ok = !t.word_start || i == 0 || !is_ident(chars[i - 1]);
            let after_ok = !t.word_end || chars.get(i + pat.len()).is_none_or(|c| !is_ident(*c));
            if before_ok && after_ok {
                hits.push((lines[i], t.text));
            }
            i += pat.len();
        }
    }
    hits.sort();
    hits
}

/// The `.rs` files under `dir`, folders named in `skip` left out, sorted.
fn rust_files(dir: &Path, skip: &[&str]) -> Vec<PathBuf> {
    let mut out = Vec::new();
    let mut stack = vec![dir.to_path_buf()];
    while let Some(d) = stack.pop() {
        for entry in fs::read_dir(&d).unwrap() {
            let path = entry.unwrap().path();
            if path.is_dir() {
                let name = path.file_name().and_then(|n| n.to_str()).unwrap_or("");
                if !skip.contains(&name) {
                    stack.push(path);
                }
            } else if path.extension().is_some_and(|e| e == "rs") {
                out.push(path);
            }
        }
    }
    out.sort();
    out
}

fn rel(root: &Path, path: &Path) -> String {
    path.strip_prefix(root)
        .unwrap()
        .to_string_lossy()
        .replace('\\', "/")
}

/// Scan B over `dir`: the number of files read and one `path:line name` per hit.
fn external_call_hits(root: &Path, dir: &Path) -> (usize, Vec<String>) {
    let files = rust_files(dir, &[]);
    let mut found = Vec::new();
    for path in &files {
        let text = fs::read_to_string(path).unwrap();
        for (line, name) in find(&text, &EXTERNAL_TOKENS) {
            found.push(format!("{}:{line} {name}", rel(root, path)));
        }
    }
    (files.len(), found)
}

/// The test files of the ledger area under `<root>/tests`: a `.rs` file or a folder whose name
/// starts with `ledger`, folders searched through.
fn ledger_area_files(root: &Path) -> Vec<PathBuf> {
    let tests = root.join("tests");
    let mut out = Vec::new();
    for entry in fs::read_dir(&tests).unwrap() {
        let path = entry.unwrap().path();
        let name = path.file_name().and_then(|n| n.to_str()).unwrap_or("");
        if !name.starts_with("ledger") {
            continue;
        }
        if path.is_dir() {
            out.extend(rust_files(&path, &[]));
        } else if path.extension().is_some_and(|e| e == "rs") {
            out.push(path);
        }
    }
    out.sort();
    out
}

/// Scan A, ledger area: the number of files read and one `path:line name` per direct way to the
/// binary. The legacy pair is skipped.
fn ledger_area_problems(root: &Path) -> (usize, Vec<String>) {
    let files = ledger_area_files(root);
    let mut found = Vec::new();
    for path in &files {
        let name = rel(root, path);
        if LEGACY_LEDGER_NAMED.contains(&name.as_str()) {
            continue;
        }
        let text = fs::read_to_string(path).unwrap();
        for (line, token) in find(&text, &BINARY_TOKENS) {
            found.push(format!(
                "{name}:{line} runs the binary through {token}; use TempHome::command"
            ));
        }
    }
    (files.len(), found)
}

/// True when `text` runs the binary (`BINARY_TOKENS`) and no code in it names HOME.
fn runs_binary_without_home(text: &str) -> bool {
    let code = blank_comments(text);
    !find(text, &BINARY_TOKENS).is_empty() && !code.contains("HOME") && !code.contains("TempHome")
}

/// Every test file under `<root>/tests` (fixtures left out) that runs the binary and names no
/// HOME in code.
fn binary_without_home(root: &Path) -> BTreeSet<String> {
    rust_files(&root.join("tests"), &["fixtures"])
        .iter()
        .filter(|p| runs_binary_without_home(&fs::read_to_string(p).unwrap()))
        .map(|p| rel(root, p))
        .collect()
}

fn read_baseline(root: &Path) -> BTreeSet<String> {
    fs::read_to_string(root.join(BASELINE))
        .unwrap_or_default()
        .lines()
        .map(str::trim)
        .filter(|l| !l.is_empty() && !l.starts_with('#'))
        .map(str::to_string)
        .collect()
}

/// (names found in the tree but not in the baseline, names in the baseline no longer found).
fn baseline_problems(
    found: &BTreeSet<String>,
    baseline: &BTreeSet<String>,
) -> (Vec<String>, Vec<String>) {
    (
        found.difference(baseline).cloned().collect(),
        baseline.difference(found).cloned().collect(),
    )
}

#[test]
fn ledger_tests_run_the_binary_only_through_temp_home() {
    let (files, problems) = ledger_area_problems(Path::new(ROOT));
    // The scan is live: the ledger tests of P1 and the legacy pair are all in the area.
    assert!(files >= 5, "the ledger area has {files} files; moved?");
    assert!(problems.is_empty(), "{}", problems.join("\n"));
}

#[test]
fn the_ledger_area_still_uses_temp_home() {
    // A scan that finds no direct binary use is empty also when the tests stopped using the
    // helper altogether: at least the CLI test of P1 must go through it.
    let text = fs::read_to_string(Path::new(ROOT).join("tests/ledger_cli.rs")).unwrap();
    assert!(text.contains("home.command("), "tests/ledger_cli.rs");
}

#[test]
fn the_legacy_ledger_named_list_does_not_grow() {
    assert_eq!(LEGACY_LEDGER_NAMED.len(), 2);
    for name in LEGACY_LEDGER_NAMED {
        let text = fs::read_to_string(Path::new(ROOT).join(name)).unwrap();
        // they still carry their reason: a plain temporary HOME of their own
        assert!(text.contains("\"HOME\""), "{name}");
    }
}

/// Scan A, baseline: what is wrong with the tests under `<root>/tests` against the baseline list
/// (empty when all is well).
fn baseline_failures(root: &Path) -> Vec<String> {
    let baseline = read_baseline(root);
    if baseline.is_empty() {
        return vec![format!(
            "{BASELINE} is empty or missing; it lists the tests that predate the rule"
        )];
    }
    let (new, stale) = baseline_problems(&binary_without_home(root), &baseline);
    let mut out = Vec::new();
    for name in new {
        out.push(format!(
            "{name} runs the binary without naming HOME; run it through TempHome (or set HOME to \
             a temporary folder) instead of adding it to {BASELINE}"
        ));
    }
    for name in stale {
        out.push(format!(
            "{BASELINE} names {name}, which no longer runs the binary without a HOME; delete the \
             line (the list only shrinks)"
        ));
    }
    out
}

#[test]
fn no_new_test_runs_the_binary_without_a_home() {
    let failures = baseline_failures(Path::new(ROOT));
    assert!(failures.is_empty(), "{}", failures.join("\n"));
}

#[test]
fn the_baseline_scan_fails_on_a_new_file_and_on_a_stale_name() {
    let dir = tempfile::tempdir().unwrap();
    let tests = dir.path().join("tests");
    fs::create_dir_all(tests.join("fixtures/home_isolation")).unwrap();
    let baseline = dir.path().join(BASELINE);
    let runs = "const B: &str = common::MOAI_BIN;\n";
    fs::write(tests.join("old.rs"), runs).unwrap();
    // a fixture sample under tests/fixtures is not a test
    fs::write(tests.join("fixtures/sample.rs"), runs).unwrap();
    fs::write(
        tests.join("fine.rs"),
        format!("{runs}fn h() {{ e(\"HOME\"); }}"),
    )
    .unwrap();

    // missing or empty baseline: reported once, not as a file per line
    assert_eq!(baseline_failures(dir.path()).len(), 1);
    fs::write(&baseline, "# only a comment\n").unwrap();
    assert_eq!(baseline_failures(dir.path()).len(), 1);
    // N: the baseline equals what the scan finds
    fs::write(&baseline, "tests/old.rs\n").unwrap();
    assert_eq!(baseline_failures(dir.path()), Vec::<String>::new());
    // N+1: a second file runs the binary without a HOME and is not on the list
    fs::write(tests.join("added.rs"), runs).unwrap();
    let failures = baseline_failures(dir.path());
    assert_eq!(failures.len(), 1, "{failures:?}");
    assert!(
        failures[0].starts_with("tests/added.rs runs the binary"),
        "{failures:?}"
    );
    // a name on the list that stopped qualifying is stale
    fs::write(&baseline, "tests/added.rs\ntests/old.rs\ntests/fine.rs\n").unwrap();
    let failures = baseline_failures(dir.path());
    assert_eq!(failures.len(), 1, "{failures:?}");
    assert!(failures[0].contains("names tests/fine.rs"), "{failures:?}");
}

#[test]
fn the_ledger_library_has_no_external_call_names() {
    let root = Path::new(ROOT);
    let (files, found) = external_call_hits(root, &root.join("src/ledger"));
    assert!(files >= 6, "src/ledger has {files} files; moved?");
    assert!(found.is_empty(), "{}", found.join("\n"));
}

/// The two provider endpoints of `moai decide` (cards C-37 and C-54): the only URLs the code
/// under `src/decide/` may hold, each as a whole string literal.
const DECIDE_ENDPOINTS: [&str; 2] = [
    "\"https://api.typesafe.ai/v1/systemone\"",
    "\"https://api.openai.com/v1/decisions\"",
];

/// The one file under `src/decide/` that sends (slice S4 of card C-37, SPEC-20261118-decide-round-trip)
/// and the four names it may hold; it may hold no other name of the scan, and no other file
/// may hold these.
const DECIDE_SENDER: &str = "src/decide/transport.rs";
const DECIDE_SENDER_NAMES: [&str; 4] = ["std::net", "net::", "TcpStream", "TcpListener"];

/// Scan B over the decide library at `dir`, with two exceptions: a URL hit on a line whose
/// every URL is one of [`DECIDE_ENDPOINTS`], and the four socket names of
/// [`DECIDE_SENDER_NAMES`] in [`DECIDE_SENDER`]. The line is read as written, so a URL in a
/// comment beside an endpoint is a hit too. Returns the number of files read and one
/// `path:line name` per hit.
fn decide_external_hits(root: &Path, dir: &Path) -> (usize, Vec<String>) {
    let files = rust_files(dir, &[]);
    let mut found = Vec::new();
    for path in &files {
        let text = fs::read_to_string(path).unwrap();
        let lines: Vec<&str> = text.lines().collect();
        let is_sender = rel(root, path) == DECIDE_SENDER;
        for (line, name) in find(&text, &EXTERNAL_TOKENS) {
            if is_sender && DECIDE_SENDER_NAMES.contains(&name) {
                continue;
            }
            if name.ends_with("://") {
                let mut rest = lines[line - 1].to_string();
                for endpoint in DECIDE_ENDPOINTS {
                    rest = rest.replace(endpoint, "");
                }
                if !rest.contains("http://") && !rest.contains("https://") {
                    continue;
                }
            }
            found.push(format!("{}:{line} {name}", rel(root, path)));
        }
    }
    found.dedup();
    (files.len(), found)
}

/// Second audit of card C-37 S2, N11: the decide library builds bytes and reads bytes. It
/// opens no socket, starts no process and names no URL but the two endpoint constants.
/// The one place that sends (slice S4) has its own exception: [`DECIDE_SENDER`].
#[test]
fn the_decide_library_has_no_external_call_names() {
    let root = Path::new(ROOT);
    let (files, found) = decide_external_hits(root, &root.join("src/decide"));
    assert!(files >= 5, "src/decide has {files} files; moved?");
    assert!(found.is_empty(), "{}", found.join("\n"));
}

#[test]
fn the_decide_scan_allows_the_two_endpoints_and_nothing_else() {
    let dir = tempfile::tempdir().unwrap();
    let decide = dir.path().join("src/decide");
    fs::create_dir_all(&decide).unwrap();
    let allowed = [
        "pub const A: &str = \"https://api.typesafe.ai/v1/systemone\";",
        "pub const B: &str = \"https://api.openai.com/v1/decisions\";",
        "assert_eq!((a, b), (\"https://api.typesafe.ai/v1/systemone\", \"https://api.openai.com/v1/decisions\"));",
        "// see https://example.com/docs",
    ];
    let refused = [
        (
            "let u = \"https://api.openai.com/v1/decisions/x\";",
            "https://",
        ),
        (
            "let u = \"https://api.openai.com/v1/decision\";",
            "https://",
        ),
        ("let u = \"http://api.openai.com/v1/decisions\";", "http://"),
        (
            "let u = \"https://api.typesafe.ai.example.com/v1/systemone\";",
            "https://",
        ),
        (
            "let u = format!(\"https://{host}/v1/decisions\");",
            "https://",
        ),
        (
            "let p = (\"https://api.openai.com/v1/decisions\", \"https://example.com\");",
            "https://",
        ),
        ("let c = std::process::Command::new(x);", "Command::new("),
        ("let s = std::net::TcpStream::connect(a);", "TcpStream"),
        ("let p = \"curl\";", "\"curl\""),
    ];
    fs::write(decide.join("ok.rs"), allowed.join("\n")).unwrap();
    assert_eq!(decide_external_hits(dir.path(), &decide), (1, vec![]));
    for (at, (sample, name)) in refused.iter().enumerate() {
        let text = format!("{}\n{sample}\n", allowed[0]);
        fs::write(decide.join("bad.rs"), text).unwrap();
        let (files, found) = decide_external_hits(dir.path(), &decide);
        assert_eq!(files, 2);
        let want = format!("src/decide/bad.rs:2 {name}");
        assert!(found.contains(&want), "sample {at}: {found:?}");
        assert!(
            found
                .iter()
                .all(|hit| hit.starts_with("src/decide/bad.rs:2 ")),
            "sample {at}: {found:?}"
        );
    }
}

/// Slice S4 of card C-37 (SPEC-20261118-decide-round-trip): the one file that sends may hold
/// the four socket names; nothing else is allowed there or anywhere else under `src/decide/`.
#[test]
fn the_decide_scan_allows_the_four_socket_names_in_the_sending_file_only() {
    let dir = tempfile::tempdir().unwrap();
    let decide = dir.path().join("src/decide");
    fs::create_dir_all(&decide).unwrap();
    let sockets = [
        "use std::net::{SocketAddr, TcpStream};",
        "let l = TcpListener::bind(a);",
        "let s = TcpStream::connect(a);",
        "let a: net::SocketAddr = x;",
    ]
    .join("\n");
    // Allowed: those four names, in the sending file, and nowhere else.
    fs::write(decide.join("transport.rs"), &sockets).unwrap();
    assert_eq!(decide_external_hits(dir.path(), &decide), (1, vec![]));
    fs::write(decide.join("other.rs"), &sockets).unwrap();
    let (files, found) = decide_external_hits(dir.path(), &decide);
    assert_eq!(files, 2);
    assert!(
        !found.is_empty(),
        "the same text in another file must be a hit"
    );
    assert!(
        found
            .iter()
            .all(|hit| hit.starts_with("src/decide/other.rs:")),
        "{found:?}"
    );
    fs::remove_file(decide.join("other.rs")).unwrap();
    // Still a hit in the sending file: any other name of the scan, and a name of the same
    // family that is not one of the four.
    for (sample, name) in [
        ("let c = std::process::Command::new(x);", "Command::new("),
        ("let u = \"https://example.com\";", "https://"),
        ("let u = \"http://example.com\";", "http://"),
        ("let p = \"curl\";", "\"curl\""),
        ("let p = \"git\";", "\"git\""),
        ("let s = UdpSocket::bind(a);", "UdpSocket"),
        ("use crate::forge;", "crate::forge"),
    ] {
        fs::write(
            decide.join("transport.rs"),
            format!("{sockets}\n{sample}\n"),
        )
        .unwrap();
        let (_, found) = decide_external_hits(dir.path(), &decide);
        let want = format!("src/decide/transport.rs:5 {name}");
        assert!(found.contains(&want), "{sample}: {found:?}");
        assert!(
            found
                .iter()
                .all(|hit| hit.starts_with("src/decide/transport.rs:5 ")),
            "{sample}: {found:?}"
        );
    }
}

/// The exception above is only as good as the sending file's guard: the real file checks that
/// the address is a loopback address before it opens a connection, and it keeps the key in a
/// type with no value in its `Debug`.
#[test]
fn the_sending_file_checks_the_address_before_it_connects() {
    let text = fs::read_to_string(Path::new(ROOT).join("src/decide/transport.rs")).unwrap();
    let code = blank_comments(&text);
    let guard = code
        .find("is_loopback()")
        .expect("the loopback guard is gone");
    let connect = code
        .find("connect_timeout(")
        .expect("no connection is made?");
    assert!(
        guard < connect,
        "the loopback check must come before the connection"
    );
    assert!(
        code.contains("f.write_str(\"ApiKey(..)\")"),
        "the key type must print nothing of its value"
    );
}

#[test]
fn a_temp_home_child_sees_only_the_temp_home() {
    let home = TempHome::new();
    let out = home
        .command_for(
            if cfg!(target_os = "macos") {
                "/usr/bin/env"
            } else {
                "env"
            },
            &[],
        )
        .output()
        .expect("env");
    let text = String::from_utf8_lossy(&out.stdout);
    let vars: BTreeMap<&str, &str> = text.lines().filter_map(|l| l.split_once('=')).collect();
    let path = home.path().to_str().unwrap();
    assert_eq!(vars.get("HOME"), Some(&path), "{text}");
    assert_eq!(vars.get("USERPROFILE"), Some(&path), "{text}");
    // nothing else of the caller's environment is passed on (the developer folder excepted)
    let extra: Vec<&&str> = vars
        .keys()
        .filter(|k| !["HOME", "USERPROFILE", "DEVELOPER_DIR"].contains(*k))
        .collect();
    assert!(extra.is_empty(), "{extra:?}");
}

// --- the scanners themselves: a sample they must flag and a sample they must pass ---

#[test]
fn scan_b_flags_every_name_and_ignores_comments() {
    // each sample holds one name the ledger library must not contain; the hit names it
    let positive = [
        ("let c = Command::new(prog);", "Command::new("),
        (
            "let c = std::process::Command::from(x);",
            "process::Command",
        ),
        ("use std::net::TcpStream;", "std::net"),
        ("use std::{io, net::Ipv4Addr};", "net::"),
        ("let s = TcpStream::connect(a);", "TcpStream"),
        ("let l = TcpListener::bind(a);", "TcpListener"),
        ("let u = UdpSocket::bind(a);", "UdpSocket"),
        ("let url = \"http://x\";", "http://"),
        ("let url = \"https://x\";", "https://"),
        ("let p = \"gh\";", "\"gh\""),
        ("let p = \"git\";", "\"git\""),
        ("let p = \"curl\";", "\"curl\""),
        ("let p = \"ssh\";", "\"ssh\""),
        ("use crate::forge::transport;", "crate::forge"),
        ("use moai_cli::forge;", "moai_cli::forge"),
    ];
    for (sample, name) in positive {
        let hits = find(sample, &EXTERNAL_TOKENS);
        assert!(hits.iter().any(|(_, n)| *n == name), "{sample}: {hits:?}");
    }
    // a call rustfmt wrapped is the same call, on the line it starts
    let wrapped = "let c = std::process\n    ::Command\n    ::new(x);\n";
    let hits = find(wrapped, &EXTERNAL_TOKENS);
    assert!(hits.iter().any(|(_, n)| *n == "Command::new("), "{hits:?}");
    // a hit reports its own line
    let two = "fn a() {}\n\nfn b() { let u = \"https://x\"; }\n";
    assert_eq!(find(two, &EXTERNAL_TOKENS), vec![(3, "https://")]);

    let negative = [
        "// Command::new(prog) and https://example.com and \"gh\"",
        "/// calls `std::net::TcpStream` or `git`",
        "/* block: Command::new(x)\n and http://x */ let a = 1;",
        "let a = OsCommand::new(x);",
        "use std::os::unix::process::CommandExt;",
        "let a = \"github\"; let b = \"digit\"; let c = \"see the git docs\";",
        "let a = \"ssh key\";",
        "let a = TcpStreamLike;",
        "let n = ipnet::x();",
        "let a = crate::forge_like::x;",
        "let r = r#\"plain\"#;",
    ];
    for sample in negative {
        assert_eq!(find(sample, &EXTERNAL_TOKENS), vec![], "{sample}");
    }
    // the lexer: a comment marker inside a string or a raw string is not a comment, and a
    // quote inside a character literal does not open a string
    let tricky = [
        "let a = \"//\"; let c = Command::new(x);",
        "let a = r#\"// not a comment\"#; let c = Command::new(x);",
        "let q = '\"'; let c = Command::new(x);",
        "fn f<'a>(x: &'a str) { let c = Command::new(x); }",
        "let a = \"\\\"\"; let c = Command::new(x);",
        "let n = '\\''; let c = Command::new(x);",
    ];
    for sample in tricky {
        assert_eq!(find(sample, &EXTERNAL_TOKENS).len(), 1, "{sample}");
    }
    assert_eq!(
        find("let a = r#\"https://x\"#;", &EXTERNAL_TOKENS).len(),
        1,
        "a URL in a raw string is code"
    );
}

#[test]
fn scan_b_walks_src_ledger_through_subfolders() {
    let dir = tempfile::tempdir().unwrap();
    let ledger = dir.path().join("src/ledger");
    fs::create_dir_all(ledger.join("deep/er")).unwrap();
    fs::write(ledger.join("a.rs"), "// Command::new(x)\nfn a() {}\n").unwrap();
    fs::write(
        ledger.join("deep/er/b.rs"),
        "fn b() {\n    let u = \"https://x\";\n}\n",
    )
    .unwrap();
    fs::write(ledger.join("deep/notes.txt"), "Command::new(x)").unwrap();
    let (files, found) = external_call_hits(dir.path(), &ledger);
    assert_eq!(files, 2);
    assert_eq!(found, vec!["src/ledger/deep/er/b.rs:2 https://"]);
    // an empty tree has no hits and reads no file
    let empty = dir.path().join("src/other");
    fs::create_dir_all(&empty).unwrap();
    assert_eq!(external_call_hits(dir.path(), &empty), (0, vec![]));
}

fn violating_tree() -> tempfile::TempDir {
    let dir = tempfile::tempdir().unwrap();
    let tests = dir.path().join("tests");
    fs::create_dir_all(tests.join("ledger_more")).unwrap();
    fs::create_dir_all(tests.join("other")).unwrap();
    // runs the binary through TempHome only
    fs::write(
        tests.join("ledger_good.rs"),
        "use common::home::TempHome;\nfn t() { let h = TempHome::new(); h.command(&[\"x\"]); }\n",
    )
    .unwrap();
    // the HOME of the real user would be inherited: the binary path used directly
    fs::write(
        tests.join("ledger_bad.rs"),
        "use std::process::Command;\n\nfn t() {\n    Command::new(common::MOAI_BIN).output();\n}\n",
    )
    .unwrap();
    // a folder of the ledger area is searched through
    fs::write(
        tests.join("ledger_more/deep.rs"),
        "const B: &str = env!(\"CARGO_BIN_EXE_moai\");\n",
    )
    .unwrap();
    // outside the ledger area the same text is not this scan's business
    fs::write(
        tests.join("other/free.rs"),
        "fn t() { Command::new(common::MOAI_BIN); }\n",
    )
    .unwrap();
    fs::write(tests.join("ledger_notes.txt"), "MOAI_BIN").unwrap();
    dir
}

#[test]
fn scan_a_names_the_file_and_line_of_a_direct_binary_use() {
    let dir = violating_tree();
    let (files, problems) = ledger_area_problems(dir.path());
    assert_eq!(files, 3);
    assert_eq!(
        problems,
        vec![
            "tests/ledger_bad.rs:4 runs the binary through MOAI_BIN; use TempHome::command",
            "tests/ledger_more/deep.rs:1 runs the binary through CARGO_BIN_EXE_; use TempHome::command",
        ]
    );
}

#[test]
fn scan_a_binary_tokens_are_found_and_comments_are_not() {
    let hit = |s: &str| find(s, &BINARY_TOKENS).len();
    assert_eq!(hit("let b = common::MOAI_BIN;"), 1);
    assert_eq!(hit("let b = MOAI_BIN;"), 1);
    assert_eq!(hit("let b = crate::common::MOAI_BIN;"), 1);
    assert_eq!(hit("let b = MOAI_CLI_BIN;"), 1);
    assert_eq!(hit("let b = env!(\"CARGO_BIN_EXE_moai-cli\");"), 1);
    assert_eq!(hit("use common::cli::cli;"), 1);
    assert_eq!(hit("use common::\n    cli::cli;"), 1);
    // not the binary
    assert_eq!(hit("let b = MOAI_BIN_NAME;"), 0);
    assert_eq!(hit("// MOAI_BIN CARGO_BIN_EXE_ common::cli"), 0);
    assert_eq!(hit("use common::client::x;"), 0);
    assert_eq!(hit("let b = home.command(&[\"x\"]);"), 0);
}

#[test]
fn the_baseline_names_only_tests_that_run_the_binary_without_a_home() {
    let sample_run = "const B: &str = common::MOAI_BIN;\nfn t() { Command::new(B); }\n";
    // N: no HOME in code (a comment naming it does not count)
    assert!(runs_binary_without_home(sample_run));
    assert!(runs_binary_without_home(&format!(
        "// HOME is real\n{sample_run}"
    )));
    // N+1: one HOME, one TempHome in code
    assert!(!runs_binary_without_home(&format!(
        "{sample_run}fn h() {{ e(\"HOME\"); }}"
    )));
    assert!(!runs_binary_without_home(&format!(
        "{sample_run}use x::TempHome;"
    )));
    // a file that never runs the binary is outside the baseline
    assert!(!runs_binary_without_home("fn t() { run(); }"));
}

#[test]
fn the_baseline_only_shrinks() {
    let set =
        |names: &[&str]| -> BTreeSet<String> { names.iter().map(|s| s.to_string()).collect() };
    // equal: nothing to report
    let (new, stale) = baseline_problems(&set(&["a", "b"]), &set(&["a", "b"]));
    assert!(new.is_empty() && stale.is_empty());
    // a new file outside the baseline fails (N+1)
    let (new, stale) = baseline_problems(&set(&["a", "b", "c"]), &set(&["a", "b"]));
    assert_eq!((new, stale), (vec!["c".to_string()], vec![]));
    // a baseline name that no longer applies is reported too: the file must shrink with it
    let (new, stale) = baseline_problems(&set(&["a"]), &set(&["a", "b"]));
    assert_eq!((new, stale), (vec![], vec!["b".to_string()]));
    // an empty baseline reports every file
    let (new, _) = baseline_problems(&set(&["a"]), &set(&[]));
    assert_eq!(new, vec!["a".to_string()]);
}

#[test]
fn the_baseline_file_is_read_without_comments_or_blank_lines() {
    let dir = tempfile::tempdir().unwrap();
    let file = dir.path().join(BASELINE);
    fs::create_dir_all(file.parent().unwrap()).unwrap();
    fs::write(
        &file,
        "# header\n\ntests/a.rs\n  tests/b.rs  \n#tests/c.rs\n",
    )
    .unwrap();
    assert_eq!(
        read_baseline(dir.path()),
        ["tests/a.rs", "tests/b.rs"]
            .iter()
            .map(|s| s.to_string())
            .collect()
    );
    // a missing file reads as empty, which the real-tree scan then rejects
    assert!(read_baseline(&dir.path().join("nowhere")).is_empty());
}
