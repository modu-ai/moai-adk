//! SPEC-20261126-install-docs-match-code: the install documents (`README.md`, `README.ko.md`,
//! `install/README.md`) say what the code and the installer say. The documents and
//! `install/install.sh` are read from the tree and the built binary's help is read by running it
//! under a temporary HOME; nothing is written. Each check also proves on an in-memory text that its
//! extractor catches a wrong line, so a check cannot pass because it found nothing to read.

mod common;
use common::CleanEnv;

use std::collections::BTreeSet;
use std::path::PathBuf;
use std::process::Command;

const DOCS: [&str; 3] = ["README.md", "README.ko.md", "install/README.md"];

fn root() -> PathBuf {
    PathBuf::from(env!("CARGO_MANIFEST_DIR"))
}

fn read(rel: &str) -> String {
    std::fs::read_to_string(root().join(rel))
        .unwrap_or_else(|e| panic!("{rel} cannot be read: {e}"))
}

/// The lines outside fenced code blocks, with their 1-based numbers.
fn prose_lines(text: &str) -> Vec<(usize, &str)> {
    let mut in_fence = false;
    let mut out = Vec::new();
    for (i, line) in text.lines().enumerate() {
        if line.trim_start().starts_with("```") {
            in_fence = !in_fence;
        } else if !in_fence {
            out.push((i + 1, line));
        }
    }
    out
}

/// The closed single-backtick spans of one line.
fn code_spans(line: &str) -> Vec<&str> {
    let parts: Vec<&str> = line.split('`').collect();
    parts
        .iter()
        .enumerate()
        .filter(|(i, p)| i % 2 == 1 && i + 1 < parts.len() && !p.is_empty())
        .map(|(_, p)| *p)
        .collect()
}

/// `(line, word)` of every span of the form `moai <word>` or `moai-cli <word>` whose word is
/// lower case (letters, digits and `-`). A span such as `moai --version` names no subcommand.
fn command_words(text: &str) -> Vec<(usize, String)> {
    let mut out = Vec::new();
    for (n, line) in prose_lines(text) {
        for span in code_spans(line) {
            let span = span.trim();
            let Some(rest) = span
                .strip_prefix("moai-cli ")
                .or_else(|| span.strip_prefix("moai "))
            else {
                continue;
            };
            let Some(word) = rest.split_whitespace().next() else {
                continue;
            };
            let lower = word.chars().next().is_some_and(|c| c.is_ascii_lowercase())
                && word
                    .chars()
                    .all(|c| c.is_ascii_lowercase() || c.is_ascii_digit() || c == '-');
            if lower {
                out.push((n, word.to_string()));
            }
        }
    }
    out
}

/// An option as written (`-l`, `--name`): its dashes and letters, without `=value` or a comma.
fn clean_option(token: &str) -> Option<String> {
    let t: String = token
        .chars()
        .take_while(|c| *c == '-' || c.is_ascii_alphanumeric())
        .collect();
    (t.starts_with('-') && t.chars().any(|c| c != '-')).then_some(t)
}

/// `(line, option)` of every option written after `moai cc` or `moai glm` in a span.
fn launcher_options(text: &str) -> Vec<(usize, String)> {
    let mut out = Vec::new();
    for (n, line) in prose_lines(text) {
        for span in code_spans(line) {
            let mut words = span.split_whitespace();
            if words.next() != Some("moai") || !matches!(words.next(), Some("cc" | "glm")) {
                continue;
            }
            out.extend(
                words
                    .filter(|w| w.starts_with('-'))
                    .filter_map(clean_option)
                    .map(|o| (n, o)),
            );
        }
    }
    out
}

/// `(line, path)` of every relative markdown link `](path)`: not a web address, a mail address
/// or an anchor, and without a `#fragment` or `?query`.
fn relative_links(text: &str) -> Vec<(usize, String)> {
    let mut out = Vec::new();
    for (n, line) in prose_lines(text) {
        for (i, _) in line.match_indices("](") {
            let rest = &line[i + 2..];
            let Some(end) = rest.find(')') else { continue };
            let target = rest[..end].trim().trim_start_matches('<');
            let target = target.split_whitespace().next().unwrap_or("");
            let target = target.trim_end_matches('>');
            if target.is_empty()
                || target.starts_with('#')
                || target.starts_with("mailto:")
                || target.contains("://")
            {
                continue;
            }
            let path = target.split(['#', '?']).next().unwrap_or("");
            if !path.is_empty() {
                out.push((n, path.to_string()));
            }
        }
    }
    out
}

/// The subcommands a help text lists: rows indented by exactly two spaces that start with a
/// lower-case word (option rows start with `-`, wrapped text is indented deeper).
fn listed_subcommands(help: &str) -> BTreeSet<String> {
    help.lines()
        .filter_map(|line| line.strip_prefix("  "))
        .filter(|rest| rest.chars().next().is_some_and(|c| c.is_ascii_lowercase()))
        .filter_map(|rest| rest.split_whitespace().next())
        .map(str::to_string)
        .collect()
}

/// The options a help text declares: the leading `-x, --long <VALUE>` words of each option row.
fn declared_options(help: &str) -> BTreeSet<String> {
    let mut out = BTreeSet::new();
    for line in help.lines().map(str::trim_start) {
        if !line.starts_with('-') {
            continue;
        }
        for token in line.split_whitespace() {
            if !token.starts_with('-') {
                break;
            }
            out.extend(clean_option(token));
        }
    }
    out
}

/// Runs the built binary under a temporary HOME. An error names the command it ran.
fn run_binary(args: &[&str]) -> Result<String, String> {
    let shown = format!("moai {}", args.join(" "));
    let home = tempfile::TempDir::new()
        .map_err(|e| format!("`{shown}` was not run: no temporary HOME: {e}"))?;
    let out = Command::new(common::MOAI_BIN)
        .args(args)
        .clean_env()
        .env("HOME", home.path())
        .current_dir(home.path())
        .output()
        .map_err(|e| format!("`{shown}` could not be run ({}): {e}", common::MOAI_BIN))?;
    let text = String::from_utf8_lossy(&out.stdout).into_owned();
    if !out.status.success() {
        return Err(format!(
            "`{shown}` exited with {:?}; stderr: {}",
            out.status.code(),
            String::from_utf8_lossy(&out.stderr).trim()
        ));
    }
    if text.trim().is_empty() {
        return Err(format!("`{shown}` printed nothing"));
    }
    Ok(text)
}

fn unknown_commands(doc: &str, text: &str, listed: &BTreeSet<String>) -> Vec<String> {
    command_words(text)
        .into_iter()
        .filter(|(_, word)| !listed.contains(word))
        .map(|(n, word)| {
            format!("{doc}:{n}: `{word}` is not a subcommand listed by `moai help --all`")
        })
        .collect()
}

fn unknown_options(doc: &str, text: &str, declared: &BTreeSet<String>) -> Vec<String> {
    launcher_options(text)
        .into_iter()
        .filter(|(_, opt)| !declared.contains(opt))
        .map(|(n, opt)| format!("{doc}:{n}: option `{opt}` is not listed by `moai cc --help`"))
        .collect()
}

fn broken_links(doc: &str, text: &str) -> Vec<String> {
    let base = root()
        .join(doc)
        .parent()
        .map(PathBuf::from)
        .unwrap_or_else(root);
    let tree = std::fs::canonicalize(root()).expect("the repository root exists");
    relative_links(text)
        .into_iter()
        .filter_map(|(n, path)| match std::fs::canonicalize(base.join(&path)) {
            Err(_) => Some(format!(
                "{doc}:{n}: link `{path}` is no file or folder of the tree"
            )),
            Ok(real) if !real.starts_with(&tree) => {
                Some(format!("{doc}:{n}: link `{path}` leaves the tree"))
            }
            Ok(_) => None,
        })
        .collect()
}

/// A text with one wrong line of each kind, to prove the extractors catch it.
const FAKE: &str = "see `moai nonesuch` and `moai-cli card list` and `moai --version`\n\
```\n`moai fenced`\n```\n\
[a](missing/file.md) [b](https://x.y/z) [c](#top) [d](Cargo.toml#frag)\n\
and `moai cc --nonesuch -l` plus `moai glm --name <n>`\n";

/// Run first by every check below: an extractor that misses a wrong line fails the check itself.
fn the_extractors_catch_a_wrong_line() {
    let commands: Vec<_> = command_words(FAKE);
    assert_eq!(
        commands,
        [
            (1, "nonesuch".to_string()),
            (1, "card".to_string()),
            (6, "cc".to_string()),
            (6, "glm".to_string())
        ]
    );
    let listed: BTreeSet<String> = ["card", "cc", "glm"].map(String::from).into();
    let bad = unknown_commands("fake.md", FAKE, &listed);
    assert_eq!(bad.len(), 1, "{bad:?}");
    assert!(
        bad[0].starts_with("fake.md:1:") && bad[0].contains("`nonesuch`"),
        "{bad:?}"
    );

    let options = launcher_options(FAKE);
    assert_eq!(
        options,
        [
            (6, "--nonesuch".to_string()),
            (6, "-l".to_string()),
            (6, "--name".to_string())
        ]
    );
    let help =
        "Options:\n  -l, --leader\n          text\n\n      --name <NAME>\n          text -x\n";
    let declared = declared_options(help);
    assert_eq!(
        declared.iter().map(String::as_str).collect::<Vec<_>>(),
        ["--leader", "--name", "-l"]
    );
    let bad = unknown_options("fake.md", FAKE, &declared);
    assert_eq!(bad.len(), 1, "{bad:?}");
    assert!(
        bad[0].starts_with("fake.md:6:") && bad[0].contains("`--nonesuch`"),
        "{bad:?}"
    );

    assert_eq!(
        relative_links(FAKE),
        [
            (5, "missing/file.md".to_string()),
            (5, "Cargo.toml".to_string())
        ]
    );
    let bad = broken_links("README.md", FAKE);
    assert_eq!(bad.len(), 1, "{bad:?}");
    assert!(
        bad[0].starts_with("README.md:5:") && bad[0].contains("`missing/file.md`"),
        "{bad:?}"
    );

    let help_all = "Usage: moai <command>\n       moai cc|glm    start\n\nStart\n  cc     Start\n  glm    Start\n\nOptions:\n  -l, --leader  text\n                wrapped\n";
    assert_eq!(
        listed_subcommands(help_all).into_iter().collect::<Vec<_>>(),
        ["cc", "glm"]
    );
}

// REQ-5, REQ-6, AC-7.
#[test]
fn every_command_the_documents_name_is_a_subcommand_of_the_binary() {
    the_extractors_catch_a_wrong_line();
    // The binary's own list: if it cannot be read or is empty the test fails, never skips.
    let help = run_binary(&["help", "--all"]).unwrap_or_else(|e| panic!("{e}"));
    let listed = listed_subcommands(&help);
    assert!(
        !listed.is_empty(),
        "`moai help --all` listed no subcommand (rows indented by two spaces)"
    );
    let mut problems = Vec::new();
    let mut found = 0;
    for doc in DOCS {
        let text = read(doc);
        found += command_words(&text).len();
        problems.extend(unknown_commands(doc, &text, &listed));
    }
    assert!(
        found > 0,
        "no `moai <word>` or `moai-cli <word>` code span was found in {DOCS:?}, so nothing was checked"
    );
    assert!(problems.is_empty(), "\n{}", problems.join("\n"));
}

// REQ-7, AC-8.
#[test]
fn every_launcher_option_the_documents_name_is_in_the_launcher_help() {
    the_extractors_catch_a_wrong_line();
    let help = run_binary(&["cc", "--help"]).unwrap_or_else(|e| panic!("{e}"));
    let declared = declared_options(&help);
    assert!(
        declared.contains("--name"),
        "`moai cc --help` declared no `--name` option (read: {declared:?})"
    );
    let mut problems = Vec::new();
    let mut found = 0;
    for doc in DOCS {
        let text = read(doc);
        found += launcher_options(&text).len();
        problems.extend(unknown_options(doc, &text, &declared));
    }
    assert!(
        found > 0,
        "no option after `moai cc` or `moai glm` was found in {DOCS:?}, so nothing was checked"
    );
    assert!(problems.is_empty(), "\n{}", problems.join("\n"));
}

// REQ-8, AC-9.
#[test]
fn every_relative_link_of_the_documents_points_at_a_file_of_the_tree() {
    the_extractors_catch_a_wrong_line();
    let mut problems = Vec::new();
    let mut found = 0;
    for doc in DOCS {
        let text = read(doc);
        found += relative_links(&text).len();
        problems.extend(broken_links(doc, &text));
    }
    assert!(
        found > 0,
        "no relative link was found in {DOCS:?}, so nothing was checked"
    );
    assert!(problems.is_empty(), "\n{}", problems.join("\n"));
}

/// What differs between the two READMEs: the headings and the words of the language.
struct Readme {
    file: &'static str,
    install_heading: &'static str,
    first_run_heading: &'static str,
    state_row_label: &'static str,
    /// The word that says the kept state folder belongs to the earlier RC.
    earlier: &'static str,
    /// Wording that says the installer is not published, which the tree contradicts.
    banned: &'static [&'static str],
}

const READMES: [Readme; 2] = [
    Readme {
        file: "README.md",
        install_heading: "## Install",
        first_run_heading: "## First run",
        state_row_label: "State folders",
        earlier: "earlier",
        banned: &["not_run", "not published"],
    },
    Readme {
        file: "README.ko.md",
        install_heading: "## 설치",
        first_run_heading: "## 첫 실행",
        state_row_label: "상태 폴더",
        earlier: "이전",
        banned: &["not_run", "게시되지"],
    },
];

/// The body of a `## ` section: its first line number and its lines up to the next `## `.
fn section(text: &str, heading: &str) -> Option<(usize, Vec<(usize, String)>)> {
    let lines: Vec<&str> = text.lines().collect();
    let start = lines.iter().position(|l| l.trim_end() == heading)?;
    let body = lines[start + 1..]
        .iter()
        .enumerate()
        .take_while(|(_, l)| !l.starts_with("## "))
        .map(|(i, l)| (start + 2 + i, l.to_string()))
        .collect();
    Some((start + 1, body))
}

/// Whether `text` names `folder` as a whole name: the new folder name followed by `-cli` is the
/// earlier folder, not this one.
fn names_folder(text: &str, folder: &str) -> bool {
    text.match_indices(folder).any(|(i, _)| {
        let next = text[i + folder.len()..].chars().next();
        !next.is_some_and(|c| c == '-' || c == '_' || c.is_alphanumeric())
    })
}

/// `Ok` when each word comes after the one before it, else the first one that does not.
fn in_order(text: &str, words: &[&str]) -> Result<(), String> {
    let mut at = 0;
    for word in words {
        match text[at..].find(word) {
            Some(i) => at += i + word.len(),
            None => {
                return Err(format!(
                    "`{word}` is missing, or comes before an earlier step"
                ))
            }
        }
    }
    Ok(())
}

// REQ-1, AC-1.
#[test]
fn the_state_folder_sentences_match_the_state_root_rules() {
    let new_dir = moai_cli::state_root::MOAI_DIR;
    let kept: Vec<&str> = {
        let mut v = vec![moai_cli::PROJECT_STATE_DIR, moai_cli::HOME_STATE_DIR];
        v.dedup();
        v
    };
    let mut problems = Vec::new();
    for readme in &READMES {
        let file = readme.file;
        let text = read(file);
        // The table row: the RC1 column names the new folder and the kept one.
        let row = text.lines().enumerate().find(|(_, l)| {
            l.starts_with('|')
                && l.split('|')
                    .nth(1)
                    .is_some_and(|c| c.trim() == readme.state_row_label)
        });
        match row {
            None => problems.push(format!(
                "{file}: no table row labelled `{}` was found",
                readme.state_row_label
            )),
            Some((i, line)) => {
                let cell = line.split('|').nth(2).unwrap_or("");
                for folder in std::iter::once(&new_dir).chain(kept.iter()) {
                    if !names_folder(cell, folder) {
                        problems.push(format!(
                            "{file}:{}: the RC1 cell of the `{}` row is missing `{folder}`",
                            i + 1,
                            readme.state_row_label
                        ));
                    }
                }
            }
        }
        // Every sentence group (not a table row) that names the kept folder also names the new
        // folder and says the kept one belongs to the earlier RC.
        let mut group: Vec<(usize, &str)> = Vec::new();
        let mut groups: Vec<Vec<(usize, &str)>> = Vec::new();
        for (n, line) in prose_lines(&text) {
            if line.trim().is_empty() {
                groups.push(std::mem::take(&mut group));
            } else if !line.starts_with('|') {
                group.push((n, line));
            }
        }
        groups.push(group);
        for group in groups.iter().filter(|g| !g.is_empty()) {
            let body: String = group.iter().map(|(_, l)| *l).collect::<Vec<_>>().join("\n");
            if !kept.iter().any(|k| names_folder(&body, k)) {
                continue;
            }
            let at = group[0].0;
            if !names_folder(&body, new_dir) {
                problems.push(format!(
                    "{file}:{at}: a sentence names `{}` but not `{new_dir}`",
                    kept[0]
                ));
            }
            if !body.contains(readme.earlier) {
                problems.push(format!(
                    "{file}:{at}: a sentence names `{}` without the word `{}` (the earlier RC's folder)",
                    kept[0], readme.earlier
                ));
            }
        }
    }
    assert!(problems.is_empty(), "\n{}", problems.join("\n"));
}

// REQ-2, REQ-3, AC-2, AC-3.
#[test]
fn the_installer_section_names_the_scripts_and_the_entry_the_installer_prints() {
    let steps = ["moai update", "moai cc", "/moai project"];
    let mut problems = Vec::new();
    for readme in &READMES {
        let file = readme.file;
        let text = read(file);
        match section(&text, readme.install_heading) {
            None => problems.push(format!("{file}: no `{}` section", readme.install_heading)),
            Some((start, body)) => {
                let joined: String = body
                    .iter()
                    .map(|(_, l)| l.as_str())
                    .collect::<Vec<_>>()
                    .join("\n");
                for word in [
                    "install/install.sh",
                    "install/install.ps1",
                    "https://adk.mo.ai.kr/install.sh",
                    "https://adk.mo.ai.kr/install.ps1",
                    "](install/README.md)",
                ] {
                    if !joined.contains(word) {
                        problems.push(format!(
                            "{file}:{start}: the `{}` section is missing `{word}`",
                            readme.install_heading
                        ));
                    }
                }
                for banned in readme.banned {
                    for (n, line) in &body {
                        if line.to_lowercase().contains(&banned.to_lowercase()) {
                            problems
                                .push(format!("{file}:{n}: the installer section says `{banned}`"));
                        }
                    }
                }
            }
        }
        match section(&text, readme.first_run_heading) {
            None => problems.push(format!("{file}: no `{}` section", readme.first_run_heading)),
            Some((start, body)) => {
                let joined: String = body
                    .iter()
                    .map(|(_, l)| l.as_str())
                    .collect::<Vec<_>>()
                    .join("\n");
                if let Err(e) = in_order(&joined, &steps) {
                    problems.push(format!(
                        "{file}:{start}: the first run does not give {steps:?} in that order: {e}"
                    ));
                }
            }
        }
    }
    // The same three steps in the installer's own closing output.
    let script = read("install/install.sh");
    match script.find("Next steps").map(|s| &script[s..]) {
        None => problems.push("install/install.sh: no `Next steps` output".to_string()),
        Some(tail) => {
            let tail = tail.split("Documentation:").next().unwrap_or(tail);
            if let Err(e) = in_order(tail, &steps) {
                problems.push(format!(
                    "install/install.sh: `Next steps` does not give {steps:?}: {e}"
                ));
            }
        }
    }
    assert!(problems.is_empty(), "\n{}", problems.join("\n"));
}

/// The text of the shell function `name`: from its `name() {` line to its closing `}` at column 0
/// (a function written on one line is that line). `None` when the function is not there or never
/// closes.
fn function_text(script: &str, name: &str) -> Option<String> {
    let head = format!("{name}() {{");
    let lines: Vec<&str> = script.lines().collect();
    let start = lines.iter().position(|l| l.starts_with(&head))?;
    if lines[start].trim_end().ends_with('}') {
        return Some(lines[start].to_string());
    }
    let end = lines[start + 1..].iter().position(|l| *l == "}")? + start + 1;
    Some(lines[start..=end].join("\n"))
}

/// Whether `text` calls or names `word` as a whole word (not part of a longer name).
fn mentions_word(text: &str, word: &str) -> bool {
    let name_char = |c: char| c.is_ascii_alphanumeric() || c == '_' || c == '-';
    text.match_indices(word).any(|(i, _)| {
        !text[..i].chars().next_back().is_some_and(name_char)
            && !text[i + word.len()..].chars().next().is_some_and(name_char)
    })
}

/// `function` and every function of `script` it calls, directly or through another one, joined.
fn function_with_helpers(script: &str, function: &str) -> Option<String> {
    let names: Vec<&str> = script
        .lines()
        .filter_map(|l| l.split_once("() {").map(|(n, _)| n))
        .filter(|n| !n.is_empty() && n.chars().all(|c| c.is_ascii_lowercase() || c == '_'))
        .collect();
    let mut included = vec![function.to_string()];
    let mut texts = vec![function_text(script, function)?];
    loop {
        let before = included.len();
        for name in &names {
            if included.iter().any(|i| i == name) {
                continue;
            }
            if texts.iter().any(|t| mentions_word(t, name)) {
                texts.push(function_text(script, name)?);
                included.push(name.to_string());
            }
        }
        if included.len() == before {
            return Some(texts.join("\n"));
        }
    }
}

/// Proves on an in-memory script that the function extractor finds a function, takes a helper it
/// calls, and answers `None` for a missing or unclosed one.
fn the_function_extractor_finds_what_it_should() {
    let script = "one() { echo 1; }\ntwo() {\n    one\n    echo 2\n}\nthree() {\n    echo 3\n";
    assert_eq!(
        function_text(script, "one").as_deref(),
        Some("one() { echo 1; }")
    );
    assert_eq!(
        function_text(script, "two").as_deref(),
        Some("two() {\n    one\n    echo 2\n}")
    );
    assert_eq!(
        function_text(script, "three"),
        None,
        "an unclosed function is not found"
    );
    assert_eq!(function_text(script, "four"), None);
    assert_eq!(
        function_with_helpers(script, "two").as_deref(),
        Some("two() {\n    one\n    echo 2\n}\none() { echo 1; }")
    );
}

// REQ-4, AC-4.
#[test]
fn the_question_source_sentence_matches_the_installer() {
    the_function_extractor_finds_what_it_should();
    let mut problems = Vec::new();

    // The sentence of install/README.md that says where an answer comes from.
    let doc = "install/README.md";
    let text = read(doc);
    for (n, line) in prose_lines(&text) {
        if line.contains("never shows the menu") {
            problems.push(format!(
                "{doc}:{n}: the text still says `never shows the menu`"
            ));
        }
    }
    match prose_lines(&text)
        .into_iter()
        .find(|(_, l)| l.contains("/dev/tty"))
    {
        None => problems.push(format!(
            "{doc}: no sentence names `/dev/tty`, the terminal a piped installer asks on"
        )),
        Some((n, line)) => {
            let at = line.find("/dev/tty").unwrap_or(0);
            let from = line[..at].rfind(". ").map_or(0, |i| i + 2);
            let to = line[at..].find(". ").map_or(line.len(), |i| at + i);
            let sentence = &line[from..to];
            if let Err(e) = in_order(sentence, &["standard input", "/dev/tty", "nobody"]) {
                problems.push(format!(
                    "{doc}:{n}: the question-source sentence does not say `standard input`, `/dev/tty` and `nobody` in that order: {e}"
                ));
            }
        }
    }

    // The three branches of `answer_source` in the script.
    let script_path = "install/install.sh";
    let script = read(script_path);
    match function_text(&script, "answer_source") {
        None => problems.push(format!(
            "{script_path}: no `answer_source()` function was found"
        )),
        Some(body) => {
            for word in ["-t 0", "'stdin'", "/dev/tty", "'tty'", "ANSWER_FROM=\"\""] {
                if !body.contains(word) {
                    problems.push(format!(
                        "{script_path}: `answer_source` is missing `{word}`"
                    ));
                }
            }
        }
    }
    assert!(problems.is_empty(), "\n{}", problems.join("\n"));
}

/// What `detect_hosts` of the script leaves in `HOSTS` with `home` as HOME, a PATH of the empty
/// folder `bin` (no `claude` command) and `CLAUDE_CONFIG_DIR` as given. The shell is
/// `/bin/sh`; the script text is written outside `home`.
fn detected_hosts(
    function: &str,
    home: &std::path::Path,
    bin: &std::path::Path,
    config_dir: Option<&std::path::Path>,
) -> String {
    let work = tempfile::TempDir::new().unwrap();
    let file = work.path().join("detect.sh");
    std::fs::write(
        &file,
        format!("{function}\ndetect_hosts\nprintf 'HOSTS=[%s]\\n' \"$HOSTS\"\n"),
    )
    .unwrap();
    let mut cmd = Command::new("/bin/sh");
    cmd.arg(&file)
        .clean_env()
        .env("HOME", home)
        .env("PATH", bin)
        .current_dir(home);
    if let Some(dir) = config_dir {
        cmd.env("CLAUDE_CONFIG_DIR", dir);
    }
    let out = cmd.output().expect("/bin/sh could not be run");
    let stdout = String::from_utf8_lossy(&out.stdout).into_owned();
    assert!(
        out.status.success(),
        "`/bin/sh detect.sh` exited with {:?}; stdout: {stdout}; stderr: {}",
        out.status.code(),
        String::from_utf8_lossy(&out.stderr)
    );
    stdout
        .lines()
        .find_map(|l| l.strip_prefix("HOSTS=[").and_then(|r| r.strip_suffix(']')))
        .unwrap_or_else(|| panic!("`detect_hosts` printed no HOSTS line: {stdout}"))
        .trim()
        .to_string()
}

// REQ-10, REQ-11, AC-6.
#[test]
fn the_installer_host_detection_follows_the_claude_config_dir() {
    the_function_extractor_finds_what_it_should();
    let script_path = "install/install.sh";
    let script = read(script_path);
    let function = function_with_helpers(&script, "detect_hosts").unwrap_or_else(|| {
        panic!("{script_path}: no `detect_hosts()` function was found, so nothing was run")
    });

    let tmp = tempfile::TempDir::new().unwrap();
    let home = std::fs::canonicalize(tmp.path()).unwrap().join("home");
    let bin = std::fs::canonicalize(tmp.path()).unwrap().join("empty-bin");
    std::fs::create_dir_all(&home).unwrap();
    std::fs::create_dir_all(&bin).unwrap();
    let moved = home.join("moved-claude");

    // (a) The variable names a folder that exists; there is no `~/.claude` and no command.
    let mut problems = Vec::new();
    let mut case = |label: &str, want: &str, got: String| {
        if got != want {
            problems.push(format!(
                "{script_path}: detect_hosts case {label} found {got:?}, expected {want:?}"
            ));
        }
    };
    std::fs::create_dir_all(&moved).unwrap();
    case(
        "(a) CLAUDE_CONFIG_DIR names an existing folder, no `.claude`",
        "claude",
        detected_hosts(&function, &home, &bin, Some(&moved)),
    );
    // (b) The variable unset and only `~/.claude`.
    std::fs::remove_dir_all(&moved).unwrap();
    std::fs::create_dir_all(home.join(".claude")).unwrap();
    case(
        "(b) variable unset, only `~/.claude`",
        "claude",
        detected_hosts(&function, &home, &bin, None),
    );
    // (c) Neither exists: no host.
    std::fs::remove_dir_all(home.join(".claude")).unwrap();
    case(
        "(c) no variable, no `.claude`, no command",
        "",
        detected_hosts(&function, &home, &bin, None),
    );

    // The warning for no host names both places it looked.
    let warning = script
        .lines()
        .enumerate()
        .find(|(_, l)| l.contains("No Claude Code found"))
        .unwrap_or_else(|| panic!("{script_path}: no `No Claude Code found` warning line"));
    for word in ["claude command", "CLAUDE_CONFIG_DIR"] {
        if !warning.1.contains(word) {
            problems.push(format!(
                "{script_path}:{}: the no-host warning is missing `{word}`",
                warning.0 + 1
            ));
        }
    }
    assert!(problems.is_empty(), "\n{}", problems.join("\n"));
}
