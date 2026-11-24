//! The test switches `MOAI_TEST_*` are read by debug builds and test builds only, and the release
//! and dist profiles keep debug assertions off (SPEC-20261125-test-switch-release-scan).
//!
//! Two scans read text at run time, so a mutant of the scanned files needs no rebuild:
//!
//! - the source scan: every code line of `src/**/*.rs` that spells the prefix `MOAI_TEST` sits
//!   under a `cfg` attribute that only a debug build or a test build compiles, and so does every
//!   line that uses a constant holding such a spelling. The whole prefix is read, not a list of
//!   names, so a new switch is covered the day it is written; the scan has no exemption list.
//! - the profile scan: `Cargo.toml` and `.cargo/config.toml` never turn debug assertions on for
//!   the `release` profile, the `dist` profile or any profile that inherits one of them (directly
//!   or through other profiles), and `dist` inherits `release`. It reads quoted and blank-padded
//!   table names, dotted keys, inline tables, the profile environment variables, and `rustflags`
//!   on one line or in an array over several lines.
//!
//! What the source scan does not follow (a change that needs one of these fails closed or has to
//! extend the scan): a name built from parts that never spells the prefix whole (`concat!`); a
//! name assembled in lower case and raised afterwards (`"moai_test_one".to_uppercase()`); a read
//! that lists the environment (`std::env::vars()` filtered by a `MOAI_` prefix); a file that a
//! `#[cfg(test)] mod x;` line of another file gates as a whole; an alias of a constant made with
//! `use ... as`; code that rustfmt would not leave as it is, because the scan reads the
//! indentation of the statement, the block and the items around a line; a spelling in a trailing
//! comment counts as code.
//!
//! What neither scan reads: the workflows under `.github/workflows` and the release scripts under
//! `tools`, which can pass `RUSTFLAGS`, `CARGO_PROFILE_*` variables or `--config` arguments to
//! cargo, and the environment of the person who builds.

use std::collections::BTreeSet;
use std::fs;
use std::path::{Path, PathBuf};

const MANIFEST: &str = env!("CARGO_MANIFEST_DIR");
const PREFIX: &str = "MOAI_TEST";

/// The attribute lines that only a debug build or a test build compiles. A negated or any other
/// `cfg` is not on the list: a new form is added here on purpose, never by accident.
const DEBUG_ONLY: [&str; 4] = [
    "#[cfg(debug_assertions)]",
    "#[cfg(test)]",
    "#[cfg(any(test, debug_assertions))]",
    "#[cfg(any(debug_assertions, test))]",
];

fn is_code(line: &str) -> bool {
    let t = line.trim();
    !t.is_empty() && !t.starts_with("//")
}

fn indent(line: &str) -> usize {
    line.len() - line.trim_start().len()
}

/// The line without leading space and without a visibility (`pub`, `pub(crate)`).
fn item_text(line: &str) -> &str {
    let t = line.trim_start();
    let Some(rest) = t.strip_prefix("pub") else {
        return t;
    };
    if let Some(inner) = rest.strip_prefix('(') {
        return inner
            .split_once(')')
            .map_or(t, |(_, after)| after.trim_start());
    }
    if rest.starts_with(' ') {
        rest.trim_start()
    } else {
        t
    }
}

/// The name of the constant or static item that the line starts (`pub const NAME: &str = ...`).
/// A `const fn` and the unnamed `const _` are no such item.
fn constant_name(line: &str) -> Option<String> {
    let t = item_text(line);
    let t = t
        .strip_prefix("const ")
        .or_else(|| t.strip_prefix("static "))?;
    let t = t.strip_prefix("mut ").unwrap_or(t);
    let name: String = t
        .chars()
        .take_while(|c| c.is_ascii_alphanumeric() || *c == '_')
        .collect();
    let typed = t[name.len()..].trim_start().starts_with(':');
    if name.is_empty() || name == "_" || !typed {
        return None;
    }
    Some(name)
}

/// Does the line hold `word` as a whole identifier?
fn has_word(line: &str, word: &str) -> bool {
    let part = |c: Option<char>| c.is_some_and(|c| c.is_ascii_alphanumeric() || c == '_');
    line.match_indices(word).any(|(at, _)| {
        let before = line[..at].chars().next_back();
        let after = line[at + word.len()..].chars().next();
        !part(before) && !part(after)
    })
}

/// Every full switch name (`MOAI_TEST_` and at least one more character) on the line.
fn collect_names(line: &str, out: &mut BTreeSet<String>) {
    let start = format!("{PREFIX}_");
    for (at, _) in line.match_indices(&start) {
        let rest = &line[at..];
        let len = rest
            .chars()
            .take_while(|c| c.is_ascii_uppercase() || c.is_ascii_digit() || *c == '_')
            .count();
        if len > start.len() {
            out.insert(rest[..len].to_string());
        }
    }
}

/// A debug-only attribute stacked directly above line `i` (blank lines, comments and other
/// attributes between them do not break the stack), or written in front of the item on the line.
fn debug_only_attribute(lines: &[&str], i: usize) -> bool {
    let own = lines[i].trim();
    if DEBUG_ONLY
        .iter()
        .any(|cfg| own.starts_with(&format!("{cfg} ")))
    {
        return true;
    }
    for line in lines[..i].iter().rev() {
        let t = line.trim();
        if t.is_empty() || t.starts_with("//") {
            continue;
        }
        if !t.starts_with("#[") {
            return false;
        }
        if DEBUG_ONLY.contains(&t) {
            return true;
        }
    }
    false
}

/// The line itself, then the nearest line above with a smaller indent, and so on up to column 0:
/// the statement, the blocks and the items that hold line `i`. The closing line of a multi-line
/// header (`) -> R {`) stands for its opener, where the attribute sits.
fn chain(lines: &[&str], i: usize) -> Vec<usize> {
    let mut out = vec![i];
    let mut at = i;
    while indent(lines[at]) > 0 {
        let depth = indent(lines[at]);
        let Some(mut j) = (0..at)
            .rev()
            .find(|&j| is_code(lines[j]) && indent(lines[j]) < depth)
        else {
            break;
        };
        if lines[j].trim_start().starts_with([')', ']']) {
            let closing = indent(lines[j]);
            let opener = (0..j).rev().find(|&k| {
                is_code(lines[k])
                    && indent(lines[k]) == closing
                    && !lines[k].trim_start().starts_with([')', ']', '}', '#'])
            });
            if let Some(k) = opener {
                j = k;
            }
        }
        out.push(j);
        at = j;
    }
    out
}

/// Is line `i` inside code that only a debug build or a test build compiles?
fn debug_only(lines: &[&str], i: usize) -> bool {
    chain(lines, i)
        .iter()
        .any(|&at| debug_only_attribute(lines, at))
}

#[derive(Debug, Default)]
struct SourceScan {
    findings: Vec<String>,
    names: BTreeSet<String>,
}

/// Scans `(path, text)` pairs of Rust sources. A line that spells the prefix and is no constant
/// definition has to be debug-only; a constant that holds a spelling is collected, and every
/// other line that uses it (an import is no use) has to be debug-only too.
fn scan_sources(files: &[(String, String)]) -> SourceScan {
    let mut out = SourceScan::default();
    let mut constants = BTreeSet::new();
    for (path, text) in files {
        let lines: Vec<&str> = text.lines().collect();
        for (i, line) in lines.iter().enumerate() {
            if !is_code(line) || !line.contains(PREFIX) {
                continue;
            }
            collect_names(line, &mut out.names);
            if let Some(name) = chain(&lines, i)
                .iter()
                .find_map(|&at| constant_name(lines[at]))
            {
                constants.insert(name);
            } else if !debug_only(&lines, i) {
                out.findings.push(format!(
                    "{path}:{}: spells {PREFIX} outside debug-only code: {}",
                    i + 1,
                    line.trim()
                ));
            }
        }
    }
    for name in &constants {
        for (path, text) in files {
            let lines: Vec<&str> = text.lines().collect();
            for (i, line) in lines.iter().enumerate() {
                let defines = constant_name(line).as_deref() == Some(name.as_str());
                if !is_code(line)
                    || !has_word(line, name)
                    || defines
                    || item_text(line).starts_with("use ")
                {
                    continue;
                }
                if !debug_only(&lines, i) {
                    out.findings.push(format!(
                        "{path}:{}: uses the constant {name}, which holds a {PREFIX} spelling, \
                         outside debug-only code: {}",
                        i + 1,
                        line.trim()
                    ));
                }
            }
        }
    }
    out
}

fn rust_files(dir: &Path, out: &mut Vec<PathBuf>) {
    let mut entries: Vec<_> = fs::read_dir(dir)
        .unwrap_or_else(|e| panic!("cannot read {}: {e}", dir.display()))
        .map(|e| e.unwrap())
        .collect();
    entries.sort_by_key(|e| e.path());
    for entry in entries {
        let kind = entry.file_type().unwrap();
        let path = entry.path();
        if kind.is_dir() {
            rust_files(&path, out);
        } else if kind.is_file() && path.extension().is_some_and(|e| e == "rs") {
            out.push(path);
        }
    }
}

fn source_files() -> Vec<(String, String)> {
    let root = Path::new(MANIFEST);
    let mut paths = Vec::new();
    rust_files(&root.join("src"), &mut paths);
    paths
        .into_iter()
        .map(|p| {
            let rel = p
                .strip_prefix(root)
                .unwrap()
                .to_string_lossy()
                .replace('\\', "/");
            let text = fs::read_to_string(&p).unwrap_or_else(|e| panic!("cannot read {rel}: {e}"));
            (rel, text)
        })
        .collect()
}

#[test]
fn every_test_switch_read_in_src_is_debug_or_test_only() {
    let files = source_files();
    let scan = scan_sources(&files);
    assert!(
        files.len() >= 50 && scan.names.len() >= 10,
        "the scan looked at {} files and found {} switch names ({:?}); a wrong root or a changed \
         layout would let it pass on nothing",
        files.len(),
        scan.names.len(),
        scan.names
    );
    assert!(
        scan.findings.is_empty(),
        "a release build must not read a test-only switch; put the read under \
         #[cfg(debug_assertions)] or #[cfg(test)]:\n{}",
        scan.findings.join("\n")
    );
}

fn scan_one(text: &str) -> SourceScan {
    scan_sources(&[("fixture.rs".to_string(), text.to_string())])
}

#[test]
fn the_gate_reader_tells_a_gated_read_from_an_ungated_one() {
    // (what the fixture shows, its text, the findings it must give)
    let cases: [(&str, &str, usize); 22] = [
        (
            "a bare read",
            r#"fn a() -> Option<String> {
    std::env::var("MOAI_TEST_ONE").ok()
}
"#,
            1,
        ),
        (
            "a read under a statement attribute",
            r#"fn a() {
    #[cfg(debug_assertions)]
    if let Some(n) = std::env::var("MOAI_TEST_ONE")
        .ok()
    {
        drop(n);
    }
}
"#,
            0,
        ),
        (
            "a read in a function under an attribute",
            r#"#[cfg(debug_assertions)]
fn a() -> Option<String> {
    std::env::var("MOAI_TEST_ONE").ok()
}
"#,
            0,
        ),
        (
            "an attribute stack with a comment and another attribute",
            r#"#[cfg(debug_assertions)]
// the seam
#[allow(dead_code)]
fn a() -> Option<String> {
    std::env::var("MOAI_TEST_ONE").ok()
}
"#,
            0,
        ),
        (
            "a function with a multi-line header",
            r#"#[cfg(debug_assertions)]
fn a(
    x: u8,
) -> Option<String> {
    std::env::var("MOAI_TEST_ONE").ok()
}
"#,
            0,
        ),
        (
            "a spelling in a test module",
            r#"#[cfg(test)]
mod tests {
    #[test]
    fn a() {
        std::env::set_var("MOAI_TEST_ONE", "1");
    }
}
"#,
            0,
        ),
        (
            "an attribute written in front of the statement",
            r#"fn a() {
    #[cfg(debug_assertions)] let v = std::env::var("MOAI_TEST_ONE");
}
"#,
            0,
        ),
        (
            "a debug or test attribute",
            r#"#[cfg(any(test, debug_assertions))]
fn a() -> Option<String> {
    std::env::var("MOAI_TEST_ONE").ok()
}
"#,
            0,
        ),
        (
            "a negated attribute",
            r#"#[cfg(not(debug_assertions))]
fn a() -> Option<String> {
    std::env::var("MOAI_TEST_ONE").ok()
}
"#,
            1,
        ),
        (
            "an attribute that is neither debug nor test",
            r#"#[cfg(unix)]
fn a() -> Option<String> {
    std::env::var("MOAI_TEST_ONE").ok()
}
"#,
            1,
        ),
        (
            "a run-time cfg check, which compiles the read into a release build",
            r#"fn a() -> Option<String> {
    if cfg!(debug_assertions) {
        return std::env::var("MOAI_TEST_ONE").ok();
    }
    None
}
"#,
            1,
        ),
        (
            "an attribute on the statement above, not on this one",
            r#"fn a() {
    #[cfg(debug_assertions)]
    let x = 1;
    let v = std::env::var("MOAI_TEST_ONE");
}
"#,
            1,
        ),
        (
            "a mention in a comment",
            r#"// MOAI_TEST_ONE is read in debug builds only
/// docs: MOAI_TEST_ONE
fn a() {}
"#,
            0,
        ),
        (
            "a name that is built from the prefix",
            r#"fn a(x: &str) -> String {
    format!("MOAI_TEST_{x}")
}
"#,
            1,
        ),
        (
            "a name that is joined from parts",
            r#"fn a() -> String {
    ["MOAI_TEST", "ONE"].join("_")
}
"#,
            1,
        ),
        (
            "a constant that is only defined",
            r#"pub const VAR: &str = "MOAI_TEST_ONE";
"#,
            0,
        ),
        (
            "a constant that is read without a cfg",
            r#"pub const VAR: &str = "MOAI_TEST_ONE";
fn a() -> Option<String> {
    std::env::var(VAR).ok()
}
"#,
            1,
        ),
        (
            "a constant that is read under a cfg",
            r#"pub const VAR: &str = "MOAI_TEST_ONE";
#[cfg(debug_assertions)]
fn a() -> Option<String> {
    std::env::var(VAR).ok()
}
"#,
            0,
        ),
        (
            "a constant list of names that is read without a cfg",
            r#"const NAMES: [&str; 2] = [
    "MOAI_TEST_ONE",
    "MOAI_TEST_TWO",
];
fn a() {
    for n in NAMES {
        let _ = std::env::var(n);
    }
}
"#,
            1,
        ),
        (
            "a constant that is only imported, and a longer identifier",
            r#"pub const VAR: &str = "MOAI_TEST_ONE";
use crate::VAR;
fn a() {
    let VAR_TWO = 1;
}
"#,
            0,
        ),
        (
            "a const fn is no constant definition",
            r#"const fn a() -> &'static str {
    "MOAI_TEST_ONE"
}
"#,
            1,
        ),
        (
            "a public crate constant that is read without a cfg",
            r#"pub(crate) const VAR: &str = "MOAI_TEST_ONE";
fn a() -> bool {
    std::env::var_os(VAR).is_some()
}
"#,
            1,
        ),
    ];
    for (what, text, expected) in cases {
        let scan = scan_one(text);
        assert_eq!(
            scan.findings.len(),
            expected,
            "{what}: {:?}\n{text}",
            scan.findings
        );
    }
    let two = scan_one(
        "fn a() {\n    let _ = (\"MOAI_TEST_ONE\", \"MOAI_TEST_TWO_X\", \"MOAI_TEST_\");\n}\n",
    );
    let expected: BTreeSet<String> = ["MOAI_TEST_ONE", "MOAI_TEST_TWO_X"]
        .iter()
        .map(|s| s.to_string())
        .collect();
    assert_eq!(
        two.names, expected,
        "a full name counts, a bare prefix does not"
    );
}

/// A header or a key path without quotes and blanks: `[ profile . "release" ]` reads as
/// `profile.release`.
fn plain(text: &str) -> String {
    text.chars()
        .filter(|c| !c.is_whitespace() && *c != '"' && *c != '\'')
        .collect()
}

/// One meaningful line of a manifest or config file: the table it sits in (as `plain` text), the
/// line without its comment, and whether it belongs to a `rustflags` setting (that line, or a line
/// inside the array that the setting opens over several lines).
struct TomlLine {
    number: usize,
    section: String,
    text: String,
    flags: bool,
}

fn toml_lines(text: &str) -> Vec<TomlLine> {
    let mut out = Vec::new();
    let mut section = String::new();
    // the `[` that a `rustflags = [` has opened and not closed yet
    let mut depth: i32 = 0;
    for (n, raw) in text.lines().enumerate() {
        let line = raw.split('#').next().unwrap_or("").trim();
        if line.is_empty() {
            continue;
        }
        if depth == 0 && line.starts_with('[') && line.ends_with(']') {
            section = plain(line.trim_matches(['[', ']']));
            continue;
        }
        let flags = depth > 0 || line.to_ascii_lowercase().contains("rustflags");
        if flags {
            let opens = line.matches('[').count() as i32 - line.matches(']').count() as i32;
            depth = (depth + opens).max(0);
        }
        out.push(TomlLine {
            number: n + 1,
            section: section.clone(),
            text: line.to_string(),
            flags,
        });
    }
    out
}

/// The profile that a line belongs to: its table `[profile.NAME...]`, a dotted key `NAME.key` or an
/// inline table `NAME = {..}` under `[profile]`, or a dotted key `profile.NAME.key` at the root.
fn profile_of(section: &str, text: &str) -> Option<String> {
    let key = plain(text.split('=').next().unwrap_or(""));
    let path = if section.is_empty() {
        key
    } else if section == "profile" {
        format!("profile.{key}")
    } else {
        section.to_string()
    };
    let mut parts = path.split('.');
    if parts.next() != Some("profile") {
        return None;
    }
    parts.next().map(str::to_string)
}

/// The profile named by an `inherits = "NAME"` anywhere in the line (an inline table too).
fn inherited(text: &str) -> Option<String> {
    let flat = plain(text);
    let at = flat.find("inherits=")?;
    let name: String = flat[at + "inherits=".len()..]
        .chars()
        .take_while(|c| *c != ',' && *c != '}')
        .collect();
    (!name.is_empty()).then_some(name)
}

/// What the profile scan found in `(file, text)` pairs of manifest and config files. A release
/// profile is `release`, `dist` and every profile that inherits one of them, directly or through
/// other profiles, whatever the order of the tables and whichever file names the `inherits`.
fn profile_findings(files: &[(&str, &str)]) -> Vec<String> {
    let parsed: Vec<(&str, Vec<TomlLine>)> =
        files.iter().map(|(f, t)| (*f, toml_lines(t))).collect();
    let mut edges = Vec::new();
    let mut dist_inherits = Vec::new();
    for (_, lines) in &parsed {
        for l in lines {
            if let (Some(profile), Some(parent)) =
                (profile_of(&l.section, &l.text), inherited(&l.text))
            {
                if profile == "dist" {
                    dist_inherits.push(parent.clone());
                }
                edges.push((profile, parent));
            }
        }
    }
    let mut family = BTreeSet::from(["release".to_string(), "dist".to_string()]);
    loop {
        let known = family.len();
        for (profile, parent) in &edges {
            if family.contains(parent) {
                family.insert(profile.clone());
            }
        }
        if family.len() == known {
            break;
        }
    }
    let mut out = Vec::new();
    for (file, lines) in &parsed {
        for l in lines {
            let upper = l.text.to_ascii_uppercase();
            let in_family = profile_of(&l.section, &l.text).is_some_and(|p| family.contains(&p));
            let in_release = in_family || l.text.contains("release") || l.text.contains("dist");
            let env_name = family.iter().any(|p| {
                let name = p.to_ascii_uppercase().replace('-', "_");
                upper.contains(&format!("CARGO_PROFILE_{name}_DEBUG_ASSERTIONS"))
            });
            if env_name {
                out.push(format!(
                    "{file}:{}: sets the debug assertions of a release profile through the environment",
                    l.number
                ));
            } else if (in_release || l.flags) && debug_assertions_on(&l.text) {
                out.push(format!(
                    "{file}:{}: turns debug assertions on for a release build: {}",
                    l.number, l.text
                ));
            }
        }
    }
    if dist_inherits.is_empty() || dist_inherits.iter().any(|v| v != "release") {
        out.push(format!(
            "the dist profile has to inherit release, it inherits {dist_inherits:?}"
        ));
    }
    out
}

/// Does the line name debug assertions with a value other than off (a bare rustc flag is on)?
fn debug_assertions_on(line: &str) -> bool {
    let lower = line.to_ascii_lowercase();
    ["debug-assertions", "debug_assertions"].iter().any(|key| {
        lower.match_indices(key).any(|(at, _)| {
            let rest = lower[at + key.len()..].trim_start_matches(['=', ' ', '"']);
            let value: String = rest
                .chars()
                .take_while(|c| c.is_ascii_alphanumeric())
                .collect();
            !matches!(value.as_str(), "false" | "off" | "no" | "n" | "0")
        })
    })
}

#[test]
fn the_release_and_dist_profiles_keep_debug_assertions_off() {
    let root = Path::new(MANIFEST);
    let manifest = fs::read_to_string(root.join("Cargo.toml")).unwrap();
    let config = fs::read_to_string(root.join(".cargo/config.toml")).unwrap_or_default();
    let findings = profile_findings(&[("Cargo.toml", &manifest), (".cargo/config.toml", &config)]);
    assert!(
        findings.is_empty(),
        "a release build must keep debug assertions off, or every test switch is read in it:\n{}",
        findings.join("\n")
    );
}

#[test]
fn the_profile_reader_tells_an_on_switch_from_an_off_one() {
    let dist = "\n[profile.dist]\ninherits = \"release\"\n";
    // (what the fixture shows, its manifest text, the findings it must give)
    let cases: [(&str, String, usize); 32] = [
        (
            "dev may keep them on",
            format!("[profile.dev]\ndebug-assertions = true\n\n[profile.dev.package.\"*\"]\ndebug-assertions = true\n\n[profile.release]\nopt-level = 3\n{dist}"),
            0,
        ),
        (
            "release on",
            format!("[profile.release]\ndebug-assertions = true\n{dist}"),
            1,
        ),
        (
            "release off",
            format!("[profile.release]\ndebug-assertions = false\n{dist}"),
            0,
        ),
        (
            "dist on",
            "[profile.release]\nopt-level = 3\n\n[profile.dist]\ninherits = \"release\"\ndebug-assertions = true\n".to_string(),
            1,
        ),
        (
            "a release package table on",
            format!("[profile.release.package.\"*\"]\ndebug-assertions = true\n{dist}"),
            1,
        ),
        (
            "a release build override on",
            format!("[profile.release.build-override]\ndebug-assertions = true\n{dist}"),
            1,
        ),
        (
            "a dotted key under the profile table",
            format!("[profile]\nrelease.debug-assertions = true\n{dist}"),
            1,
        ),
        (
            "an inline release table",
            format!("[profile]\nrelease = {{ debug-assertions = true }}\n{dist}"),
            1,
        ),
        (
            "a comment that names it",
            format!("[profile.release]\n# debug-assertions = true\nopt-level = 3\n{dist}"),
            0,
        ),
        (
            "dist inherits dev",
            "[profile.dist]\ninherits = \"dev\"\n".to_string(),
            1,
        ),
        (
            "dist inherits nothing",
            "[profile.dist]\nlto = \"thin\"\n".to_string(),
            1,
        ),
        (
            "a rustflags line that turns them on",
            format!("[build]\nrustflags = [\"-C\", \"debug-assertions=on\"]\n{dist}"),
            1,
        ),
        (
            "a rustflags line that turns them off",
            format!("[build]\nrustflags = [\"-C\", \"debug-assertions=off\"]\n{dist}"),
            0,
        ),
        (
            "a bare rustc flag",
            format!("[build]\nrustflags = [\"-Cdebug-assertions\"]\n{dist}"),
            1,
        ),
        (
            "the profile variable of the environment",
            format!("[env]\nCARGO_PROFILE_RELEASE_DEBUG_ASSERTIONS = \"true\"\n{dist}"),
            1,
        ),
        (
            "a rustflags array over several lines that turns them on",
            format!("[build]\nrustflags = [\n    \"-C\",\n    \"debug-assertions=on\",\n]\n{dist}"),
            1,
        ),
        (
            "a rustflags array over several lines that turns them off",
            format!("[build]\nrustflags = [\n    \"-C\",\n    \"debug-assertions=off\",\n]\n{dist}"),
            0,
        ),
        (
            "a bare rustc flag in a rustflags array of a target table",
            format!("[target.x86_64-unknown-linux-gnu]\nrustflags = [\n    \"-Cdebug-assertions\",\n]\n{dist}"),
            1,
        ),
        (
            "a table after a rustflags array is read again",
            format!("[build]\nrustflags = [\n    \"-C\",\n    \"debug-assertions=off\",\n]\n\n[profile.release]\ndebug-assertions = true\n{dist}"),
            1,
        ),
        (
            "a quoted release table on",
            format!("[profile.\"release\"]\ndebug-assertions = true\n{dist}"),
            1,
        ),
        (
            "a quoted dist table on",
            "[profile.release]\nopt-level = 3\n\n[profile.\"dist\"]\ninherits = \"release\"\ndebug-assertions = true\n".to_string(),
            1,
        ),
        (
            "a table name with blanks",
            format!("[ profile . release ]\ndebug-assertions = true\n{dist}"),
            1,
        ),
        (
            "another profile that inherits release, on",
            format!("[profile.production]\ninherits = \"release\"\ndebug-assertions = true\n{dist}"),
            1,
        ),
        (
            "another profile that inherits release, off",
            format!("[profile.production]\ninherits = \"release\"\ndebug-assertions = false\n{dist}"),
            0,
        ),
        (
            "inherits written after the setting",
            format!("[profile.production]\ndebug-assertions = true\ninherits = \"release\"\n{dist}"),
            1,
        ),
        (
            "a profile that inherits release through another profile",
            format!("[profile.fast]\ninherits = \"release\"\n\n[profile.production]\ninherits = \"fast\"\ndebug-assertions = true\n{dist}"),
            1,
        ),
        (
            "a profile that inherits dev may keep them on",
            format!("[profile.debuggable]\ninherits = \"dev\"\ndebug-assertions = true\n{dist}"),
            0,
        ),
        (
            "dotted keys of another profile under the profile table",
            format!("[profile]\nproduction.inherits = \"release\"\nproduction.debug-assertions = true\n{dist}"),
            1,
        ),
        (
            "an inline table of another profile",
            format!("[profile]\nproduction = {{ inherits = \"release\", debug-assertions = true }}\n{dist}"),
            1,
        ),
        (
            "dotted keys of another profile at the root",
            format!("profile.production.inherits = \"release\"\nprofile.production.debug-assertions = true\n{dist}"),
            1,
        ),
        (
            "the profile variable of another release profile",
            format!("[profile.production]\ninherits = \"release\"\n\n[env]\nCARGO_PROFILE_PRODUCTION_DEBUG_ASSERTIONS = \"true\"\n{dist}"),
            1,
        ),
        (
            "the profile variable of a profile that is no release profile",
            format!("[profile.debuggable]\ninherits = \"dev\"\n\n[env]\nCARGO_PROFILE_DEBUGGABLE_DEBUG_ASSERTIONS = \"true\"\n{dist}"),
            0,
        ),
    ];
    for (what, text, expected) in cases {
        let findings = profile_findings(&[("Cargo.toml", &text)]);
        assert_eq!(findings.len(), expected, "{what}: {findings:?}\n{text}");
    }
}
