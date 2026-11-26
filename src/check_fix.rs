//! `check fix` (SPEC-20261125-check-fast-and-fix): writes the table lines and hashes a range needs
//! when no human has to choose them, and prints the line of every other violation as it is. The
//! writes are three: the sha256 values of the skill package manifest, the first word of a new test
//! file in the area the caller names, and a test target of an `always` scan (a name in its
//! `test_targets` list, or, for a scan without the list, a `--test` option on a check line that
//! has no name filter; with no such line the line to add is printed). Every write goes through a
//! temporary file and a rename and is refused when the
//! file or a folder above it is a link. After writing, the same in-process judges that
//! `check affected` uses are asked again, so the answer is `PASS` only when nothing is left to print.

use crate::check_affected::{read_facts, scan_violations, table_violations, Facts};
use serde_json::{json, Map, Value};
use sha2::{Digest, Sha256};
use std::path::Path;
use std::process::Command;

type Result<T> = std::result::Result<T, String>;

/// The skill package folder and the manifest inside it.
const PACKAGE: &str = "skills/moai";
const MANIFEST_IN_PACKAGE: &str = "references/source-manifest.json";
const AREA_TABLE: &str = "tests/areas.toml";

/// What a scan finding says when an `always` scan lists a test file as a rule file and none of its
/// checks runs it.
const UNRUN_RULE_FILE: &str = " is a rule file but no check runs --test ";

/// What `run` needs besides the range.
#[derive(Clone, Debug, Default)]
pub struct Options {
    /// `word=area` pairs (`--area`): the area that gets the first word of a new test file.
    pub areas: Vec<(String, String)>,
    pub integration_ref: Option<String>,
}

fn manifest_path() -> String {
    format!("{PACKAGE}/{MANIFEST_IN_PACKAGE}")
}

fn sha256_hex(bytes: &[u8]) -> String {
    Sha256::digest(bytes)
        .iter()
        .map(|b| format!("{b:02x}"))
        .collect()
}

/// The manifest text with every `sha256` recomputed from the file on disk. A file that is gone
/// keeps its old hash and is listed in the second part when `lenient`; otherwise it is an error.
fn render_with(root: &Path, text: &str, lenient: bool) -> Result<(String, Vec<String>)> {
    let data: Value =
        serde_json::from_str(text).map_err(|e| format!("{}: not JSON ({e})", manifest_path()))?;
    let package = root.join(PACKAGE);
    let mut out = Map::new();
    let mut gone = Vec::new();
    for (key, value) in data
        .as_object()
        .ok_or_else(|| format!("{}: not an object", manifest_path()))?
    {
        if key != "files" {
            out.insert(key.clone(), value.clone());
            continue;
        }
        let mut files: Vec<(&String, &Value)> = value
            .as_object()
            .ok_or_else(|| format!("{}: `files` is not an object", manifest_path()))?
            .iter()
            .collect();
        files.sort_by(|a, b| a.0.cmp(b.0));
        let mut sorted = Map::new();
        for (rel, entry) in files {
            let mut entry = entry
                .as_object()
                .ok_or_else(|| format!("{rel}: an entry is not an object"))?
                .clone();
            match std::fs::read(package.join(rel)) {
                Ok(bytes) => {
                    entry.insert("sha256".into(), Value::String(sha256_hex(&bytes)));
                }
                Err(e) if lenient && e.kind() == std::io::ErrorKind::NotFound => {
                    gone.push(rel.clone());
                }
                Err(e) => return Err(format!("{rel}: {e}")),
            }
            sorted.insert(rel.clone(), Value::Object(entry));
        }
        out.insert(key.clone(), Value::Object(sorted));
    }
    let text = serde_json::to_string_pretty(&Value::Object(out))
        .map_err(|e| format!("{}: {e}", manifest_path()))?
        + "\n";
    Ok((text, gone))
}

/// The manifest text the tool writes: every `sha256` recomputed from the file on disk under
/// `<root>/skills/moai/`, `files` in sorted key order, every other key in place, two-space indent,
/// non-ASCII kept, one newline. An unreadable file is the error `<rel>: <error>`.
pub fn render_manifest(root: &Path) -> Result<String> {
    let path = root.join(manifest_path());
    let text = std::fs::read_to_string(&path).map_err(|e| format!("{}: {e}", manifest_path()))?;
    render_with(root, &text, false).map(|(rendered, _)| rendered)
}

/// Write `bytes` over the existing regular file `rel` of `root`: into a temporary file in the same
/// folder, then a rename. Every component below `root`, the file included, is looked at without
/// following links; a link or a file that is not regular is refused. `Err` is the line to print.
fn write_safely(root: &Path, rel: &str, bytes: &[u8]) -> std::result::Result<(), String> {
    use std::io::Write;
    let parts: Vec<&str> = rel.split('/').collect();
    let mut at = root.to_path_buf();
    let mut shown = String::new();
    let mut permissions = None;
    for (i, part) in parts.iter().enumerate() {
        at.push(part);
        if !shown.is_empty() {
            shown.push('/');
        }
        shown.push_str(part);
        let meta = std::fs::symlink_metadata(&at)
            .map_err(|e| format!("{rel}: nothing written, `{shown}` cannot be inspected ({e})"))?;
        if meta.file_type().is_symlink() {
            return Err(format!(
                "{rel}: nothing written, `{shown}` is a link; replace the link with the real file and run check fix again"
            ));
        }
        let last = i + 1 == parts.len();
        if (last && !meta.is_file()) || (!last && !meta.is_dir()) {
            return Err(format!(
                "{rel}: nothing written, `{shown}` is not a regular {}",
                if last { "file" } else { "folder" }
            ));
        }
        if last {
            permissions = Some(meta.permissions());
        }
    }
    let name = parts.last().copied().unwrap_or(rel);
    let temp = at.with_file_name(format!(".{name}.fix-{}", std::process::id()));
    let done = (|| -> std::io::Result<()> {
        let mut file = std::fs::OpenOptions::new()
            .write(true)
            .create_new(true)
            .open(&temp)?;
        file.write_all(bytes)?;
        if let Some(p) = permissions {
            file.set_permissions(p)?;
        }
        file.sync_all()?;
        drop(file);
        std::fs::rename(&temp, &at)
    })();
    done.map_err(|e| {
        let _ = std::fs::remove_file(&temp);
        format!("{rel}: nothing written, the replacement failed ({e})")
    })
}

/// One git call in `dir`, with the inherited git environment scrubbed.
fn git(dir: &Path, args: &[&str]) -> Result<String> {
    let mut cmd = Command::new("git");
    crate::identity::scrub_git_env(&mut cmd);
    let o = cmd
        .args(args)
        .current_dir(dir)
        .output()
        .map_err(|e| format!("cannot run git: {e}"))?;
    if !o.status.success() {
        return Err(format!(
            "git {} failed: {}",
            args.first().copied().unwrap_or(""),
            String::from_utf8_lossy(&o.stderr).trim()
        ));
    }
    Ok(String::from_utf8_lossy(&o.stdout).into_owned())
}

fn short(sha: &str) -> &str {
    sha.get(..12).unwrap_or(sha)
}

/// The writes and the printed lines of one run.
#[derive(Default)]
struct Outcome {
    wrote: Vec<(String, String)>,
    print: Vec<String>,
    /// Judge lines that a line in `print` says more exactly; the rest leaves them out.
    replaced: Vec<String>,
}

impl Outcome {
    fn write(&mut self, root: &Path, rel: &str, text: &str, what: String) {
        match write_safely(root, rel, text.as_bytes()) {
            Ok(()) => self.wrote.push((rel.to_string(), what)),
            Err(line) => self.print.push(line),
        }
    }
}

/// The sha256 of every manifest entry, refreshed. An entry is neither added nor dropped.
fn fix_manifest(root: &Path, out: &mut Outcome) {
    let rel = manifest_path();
    let text = match std::fs::read_to_string(root.join(&rel)) {
        Ok(t) => t,
        Err(e) if e.kind() == std::io::ErrorKind::NotFound => return,
        Err(e) => {
            out.print
                .push(format!("{rel}: cannot be read ({e}); fix or restore it"));
            return;
        }
    };
    match render_with(root, &text, true) {
        Err(e) => out.print.push(format!(
            "{rel}: the hashes cannot be refreshed ({e}); fix or restore the manifest"
        )),
        Ok((rendered, _)) if rendered != text => out.write(
            root,
            &rel,
            &rendered,
            "refreshed the sha256 of every entry".to_string(),
        ),
        Ok(_) => {}
    }
}

/// The manifest lines nobody can write for a human: a file of the package that has no entry (its
/// `origin` is a choice) and an entry of a file that is gone.
fn manifest_lines(facts: &Facts) -> Vec<String> {
    let rel = manifest_path();
    let Ok(text) = std::fs::read_to_string(facts.root.join(&rel)) else {
        return Vec::new();
    };
    let Ok(doc) = serde_json::from_str::<Value>(&text) else {
        return Vec::new();
    };
    let Some(files) = doc.get("files").and_then(Value::as_object) else {
        return Vec::new();
    };
    let package = facts.root.join(PACKAGE);
    let mut out = Vec::new();
    for path in &facts.present {
        let Some(key) = path.strip_prefix(&format!("{PACKAGE}/")) else {
            continue;
        };
        if key != MANIFEST_IN_PACKAGE && !files.contains_key(key) {
            out.push(format!(
                "{path}: has no entry in {rel}; add `\"{key}\"` with an `origin` that names where the file came from (check fix writes its sha256)"
            ));
        }
    }
    for key in files.keys() {
        let gone = matches!(
            std::fs::symlink_metadata(package.join(key)),
            Err(e) if e.kind() == std::io::ErrorKind::NotFound
        );
        if gone {
            out.push(format!(
                "{rel}: the entry `{key}` names a file that is gone; remove the entry"
            ));
        }
    }
    out
}

/// The `[areas]` arrays of the area table by lines (one array per line), as
/// `tests/test_area_layout.rs` keeps them: the area name and its words, in file order.
fn area_arrays(text: &str) -> Vec<(String, Vec<String>)> {
    let mut out = Vec::new();
    let mut section = "";
    for line in text.lines().map(str::trim) {
        if line.starts_with('[') && !line.contains('=') {
            section = line.trim_matches(|c| c == '[' || c == ']').trim();
            continue;
        }
        if section != "areas" || line.starts_with('#') {
            continue;
        }
        if let Some((key, rest)) = line.split_once('=') {
            let words = rest
                .split('"')
                .enumerate()
                .filter(|(i, _)| i % 2 == 1)
                .map(|(_, w)| w.to_string())
                .collect();
            out.push((key.trim().to_string(), words));
        }
    }
    out
}

/// `text` with `word` appended at the end of the one-line array of `area` in `[areas]`; `None`
/// when the table has no such array line.
fn append_word(text: &str, area: &str, word: &str) -> Option<String> {
    let mut out = String::with_capacity(text.len() + word.len() + 4);
    let mut section = "";
    let mut done = false;
    for line in text.split_inclusive('\n') {
        let trimmed = line.trim();
        if trimmed.starts_with('[') && !trimmed.contains('=') {
            section = trimmed.trim_matches(|c| c == '[' || c == ']').trim();
        }
        let is_area = !done
            && section == "areas"
            && !trimmed.starts_with('#')
            && trimmed
                .split_once('=')
                .is_some_and(|(k, _)| k.trim() == area);
        match line.rfind(']').filter(|_| is_area) {
            Some(close) => {
                let before = line[..close].trim_end();
                let sep = if before.ends_with('[') {
                    ""
                } else if before.ends_with(',') {
                    " "
                } else {
                    ", "
                };
                out.push_str(before);
                out.push_str(sep);
                out.push_str(&format!("\"{word}\""));
                out.push_str(&line[close..]);
                done = true;
            }
            None => out.push_str(line),
        }
    }
    done.then_some(out)
}

fn is_plain_word(word: &str) -> bool {
    !word.is_empty() && word.chars().all(|c| c.is_ascii_alphanumeric() || c == '-')
}

/// The first word of each added top-level test file that no area lists gets the area that
/// `--area` names for it. A word without a pair, or with a pair that names no area, is left to the
/// judge's own line, which lists the area names.
fn fix_areas(facts: &Facts, opts: &Options, out: &mut Outcome) {
    let Some(text) = facts.areas_text.as_deref() else {
        return;
    };
    let table = area_arrays(text);
    let mut current = text.to_string();
    let mut what = Vec::new();
    for path in &facts.added {
        let Some(stem) = path
            .strip_prefix("tests/")
            .and_then(|r| r.strip_suffix(".rs"))
            .filter(|s| !s.contains('/'))
        else {
            continue;
        };
        let word = stem.split('_').next().unwrap_or(stem);
        let listed = table
            .iter()
            .any(|(_, words)| words.iter().any(|w| w == word));
        let Some((_, area)) = opts.areas.iter().find(|(w, _)| w == word) else {
            continue;
        };
        if listed || !is_plain_word(word) || what.iter().any(|(w, _): &(String, String)| w == word)
        {
            continue;
        }
        match append_word(&current, area, word) {
            Some(next) => {
                current = next;
                what.push((word.to_string(), area.clone()));
            }
            None => out.print.push(format!(
                "--area {word}={area}: {AREA_TABLE} has no area of that name (areas: {})",
                table
                    .iter()
                    .map(|(n, _)| n.as_str())
                    .collect::<Vec<_>>()
                    .join(", ")
            )),
        }
    }
    if !what.is_empty() {
        let said = what
            .iter()
            .map(|(w, a)| format!("added `{w}` to the area `{a}`"))
            .collect::<Vec<_>>()
            .join("; ");
        out.write(&facts.root, AREA_TABLE, &current, said);
    }
}

/// The entry of the scan `id` in a parsed scan registry.
fn scan_of<'a>(registry: &'a Value, id: &str) -> Option<&'a Value> {
    registry
        .get("scans")?
        .as_array()?
        .iter()
        .find(|s| s.get("id").and_then(Value::as_str) == Some(id))
}

/// The scan `id` keeps its cargo targets in a `test_targets` field.
fn lists_targets(text: &str, id: &str) -> bool {
    serde_json::from_str::<Value>(text)
        .ok()
        .and_then(|v| {
            scan_of(&v, id)?
                .get("test_targets")
                .map(|list| !list.is_null())
        })
        .unwrap_or(false)
}

/// `text` of the scan registry with `stem` added to the `test_targets` array of the scan `id` at
/// its place in the ascending list, edited in place: one name is inserted with the white space
/// its neighbour has, and nothing else is re-written. `None` when the stem is no target name or
/// is listed already, the array holds anything but names, or its text cannot be found.
fn add_target_name(text: &str, id: &str, stem: &str) -> Option<String> {
    let value: Value = serde_json::from_str(text).ok()?;
    let names: Vec<&str> = scan_of(&value, id)?
        .get("test_targets")?
        .as_array()?
        .iter()
        .map(Value::as_str)
        .collect::<Option<_>>()?;
    if !crate::repo_scans::is_target_name(stem) || names.contains(&stem) {
        return None;
    }
    let at = names
        .iter()
        .position(|name| *name > stem)
        .unwrap_or(names.len());
    // The one document the edit may give: the old one with the name at its place.
    let mut grown = names.clone();
    grown.insert(at, stem);
    let mut want = value.clone();
    *want
        .get_mut("scans")?
        .as_array_mut()?
        .iter_mut()
        .find(|s| s.get("id").and_then(Value::as_str) == Some(id))?
        .get_mut("test_targets")? = json!(grown);
    // The name the new one goes before, or the last name when it goes at the end.
    let neighbour = format!("\"{}\"", names.get(at).or_else(|| names.last())?);
    let added = format!("\"{stem}\"");
    // The same array text can stand in several scans: keep the edit that gives `want`.
    text.match_indices("\"test_targets\"")
        .find_map(|(key, found)| {
            let open = key + found.len() + text[key + found.len()..].find('[')?;
            let close = open + text[open..].find(']')?;
            let start = open + text[open..close].find(&neighbour)?;
            let lead = &text[text[..start].trim_end().len()..start];
            let candidate = if at < names.len() {
                format!("{}{added},{lead}{}", &text[..start], &text[start..])
            } else {
                let end = start + neighbour.len();
                format!("{},{lead}{added}{}", &text[..end], &text[end..])
            };
            let doc: Value = serde_json::from_str(&candidate).ok()?;
            (doc == want).then_some(candidate)
        })
}

/// A cargo check line that runs whole targets: it names at least one `--test`, and every other
/// word is a `--bin <name>` pair or one of the flags below, so no name filter stands on it. A
/// `--test` added to such a line runs all of the new target. A word this reader does not know
/// counts as a filter.
fn runs_whole_targets(line: &str) -> bool {
    let Some(rest) = line.strip_prefix("cargo test ") else {
        return false;
    };
    let mut words = rest.split_whitespace();
    let mut tests = 0;
    while let Some(word) = words.next() {
        match word {
            "--test" | "--bin" => {
                if words.next().is_none() {
                    return false;
                }
                tests += usize::from(word == "--test");
            }
            "--offline" | "--locked" | "--lib" | "--bins" | "--no-fail-fast" | "--quiet" => {}
            _ => return false,
        }
    }
    tests > 0
}

/// The first check line of the scan `id` that runs whole targets.
fn whole_target_line(text: &str, id: &str) -> Option<String> {
    let value: Value = serde_json::from_str(text).ok()?;
    scan_of(&value, id)?
        .get("checks")?
        .as_array()?
        .iter()
        .filter_map(Value::as_str)
        .find(|c| runs_whole_targets(c))
        .map(String::from)
}

/// `text` of the scan registry with `stem` added to the targets of the scan `id`. A scan with a
/// `test_targets` field gets the name there and keeps its check lines as they are: a name put
/// into a filtered line would cut the module filters behind it out of the selection of
/// `check affected`. A scan without the field gets ` --test <stem>` on a check line without a
/// name filter ([`runs_whole_targets`]), edited in place: the JSON text of the line is replaced
/// and nothing else is re-written. A filtered line is never changed. `None` when the scan has no
/// such line or the line's text cannot be found as it is.
fn add_test_option(text: &str, id: &str, stem: &str) -> Option<String> {
    if lists_targets(text, id) {
        return add_target_name(text, id, stem);
    }
    let line = whole_target_line(text, id)?;
    let grown = format!("{line} --test {stem}");
    let (old_literal, new_literal) = (
        serde_json::to_string(&line).ok()?,
        serde_json::to_string(&grown).ok()?,
    );
    // The same line can stand in several scans: keep the replacement that lands in this one.
    text.match_indices(&old_literal).find_map(|(at, _)| {
        let candidate = format!(
            "{}{new_literal}{}",
            &text[..at],
            &text[at + old_literal.len()..]
        );
        let doc: Value = serde_json::from_str(&candidate).ok()?;
        let lands = doc
            .get("scans")?
            .as_array()?
            .iter()
            .find(|s| s.get("id").and_then(Value::as_str) == Some(id))?
            .get("checks")?
            .as_array()?
            .iter()
            .any(|c| c.as_str() == Some(grown.as_str()));
        lands.then_some(candidate)
    })
}

/// A registry edit is kept only when `gate::registry_only_grew` reads the whole change as growth;
/// a finding whose edit is not kept stays, as the judge's line, in the printed rest. A scan with
/// no `test_targets` field and no line that runs whole targets is not written: its judge line is
/// replaced by the check line a person adds to it.
fn fix_registry(facts: &Facts, out: &mut Outcome) {
    let rel = crate::gate::SCAN_REGISTRY;
    let findings: Vec<String> = scan_violations(
        &facts.added,
        &facts.tests,
        facts.scan_list.as_deref(),
        facts.tracked.as_deref(),
    )
    .into_iter()
    .filter(|l| l.contains(UNRUN_RULE_FILE))
    .collect();
    if findings.is_empty() {
        return;
    }
    let Ok(old) = std::fs::read_to_string(facts.root.join(rel)) else {
        return;
    };
    let mut current = old.clone();
    let mut what = Vec::new();
    for line in &findings {
        let Some((id, rest)) = line.split_once(": ") else {
            continue;
        };
        let Some(stem) = rest
            .split_once(UNRUN_RULE_FILE)
            .and_then(|(_, tail)| tail.split(';').next())
        else {
            continue;
        };
        if !lists_targets(&current, id) && whole_target_line(&current, id).is_none() {
            let finding = line.split(';').next().unwrap_or(line);
            out.print.push(format!(
                "{finding}; check fix wrote nothing, because no check line of this scan runs whole `--test` targets without a name filter: add the check line `cargo test --offline --locked --test {stem}` to the scan `{id}` in {rel} by hand"
            ));
            out.replaced.push(line.clone());
            continue;
        }
        if let Some(candidate) = add_test_option(&current, id, stem) {
            if crate::gate::registry_only_grew(rel, &old, &candidate) {
                what.push(if lists_targets(&candidate, id) {
                    format!("added `{stem}` to `test_targets` of the scan `{id}`")
                } else {
                    format!("added `--test {stem}` to a check of the scan `{id}`")
                });
                current = candidate;
            }
        }
    }
    if !what.is_empty() {
        out.write(&facts.root, rel, &current, what.join("; "));
    }
}

/// The scan registry's text is a list of scans whose checks can be read.
fn registry_readable(text: &str) -> bool {
    let Ok(value) = serde_json::from_str::<Value>(text) else {
        return false;
    };
    value
        .get("scans")
        .and_then(Value::as_array)
        .is_some_and(|scans| {
            scans
                .iter()
                .all(|s| crate::repo_scans::checks_of(s).is_ok())
        })
}

/// What is still wrong in the working tree: the table and scan judges of `check affected` on fresh
/// facts, the tables that cannot be read, and the manifest lines only a human can settle.
fn remaining(facts: &Facts) -> Vec<String> {
    let mut lines = table_violations(
        &facts.added,
        &facts.tests,
        facts.areas_text.as_deref(),
        facts.allow_text.as_deref(),
        &facts.moved,
    );
    if let Some(list) = facts.scan_list.as_deref() {
        if !registry_readable(list) {
            lines.push(format!(
                "{}: the table exists but cannot be read (it is not JSON with a `scans` list whose entries each have readable checks); fix or restore it, it is not treated as absent",
                crate::gate::SCAN_REGISTRY
            ));
        }
    }
    lines.extend(scan_violations(
        &facts.added,
        &facts.tests,
        facts.scan_list.as_deref(),
        facts.tracked.as_deref(),
    ));
    for (table, error) in &facts.unreadable {
        lines.push(format!(
            "{table}: the table exists but cannot be read ({error}); fix or restore it, it is not treated as absent"
        ));
    }
    lines.extend(manifest_lines(facts));
    lines
}

/// `check fix <base>..<head>` (REQ-9 to REQ-13): refuse unless the checkout is the range's head and
/// no tracked file is modified, write what is mechanical, then ask the judges again. The answer
/// lists what was written and every line still to print; `result` is `PASS` only when nothing is
/// left to print. The lane commits the writes and runs the command again on the new head.
pub fn run(dir: &Path, range: &str, opts: &Options) -> Result<Value> {
    let facts = read_facts(dir, range)?;
    let root = facts.root.as_path();
    let at = git(root, &["rev-parse", "HEAD"])?.trim().to_string();
    if at != facts.head {
        return Err(format!(
            "check fix refused: the checkout is at {} but the range ends at {}; check out the head first",
            short(&at),
            short(&facts.head)
        ));
    }
    let dirty = git(root, &["status", "--porcelain", "--untracked-files=no"])?;
    if !dirty.trim().is_empty() {
        return Err(
            "check fix refused: a tracked file is modified; commit or restore it first".to_string(),
        );
    }
    let mut out = Outcome::default();
    fix_manifest(root, &mut out);
    fix_areas(&facts, opts, &mut out);
    fix_registry(&facts, &mut out);
    let after = read_facts(root, range)?;
    let replaced = std::mem::take(&mut out.replaced);
    let rest = remaining(&after).into_iter();
    out.print.extend(rest.filter(|l| !replaced.contains(l)));
    let mut seen = std::collections::BTreeSet::new();
    out.print.retain(|l| seen.insert(l.clone()));
    let pass = out.print.is_empty();
    let nothing_to_do = pass && out.wrote.is_empty();
    let result = if pass { "PASS" } else { "FAIL" };
    Ok(json!({
        "action": "fix",
        "result": result,
        "summary": format!(
            "check-fix base={} head={} wrote={} print={} result={result}",
            short(&facts.base),
            short(&facts.head),
            out.wrote.len(),
            out.print.len()
        ),
        "wrote": out
            .wrote
            .iter()
            .map(|(path, what)| json!({"path": path, "what": what}))
            .collect::<Vec<_>>(),
        "print": out.print,
        "nothing_to_do": nothing_to_do,
        "writes": true,
    }))
}

#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    fn a_word_is_appended_at_the_end_of_the_named_array_only() {
        let text = "# areas\n[areas]\ncard = [\"card\"]\nrepo = [\"repo\", \"atlas\"]\nnone = []\n\n[solo]\nfiles = []\n";
        let one = append_word(text, "repo", "zz").unwrap();
        assert_eq!(
            one,
            text.replace("\"atlas\"]", "\"atlas\", \"zz\"]"),
            "the word follows the last word of the array"
        );
        let empty = append_word(text, "none", "zz").unwrap();
        assert_eq!(empty, text.replace("none = []", "none = [\"zz\"]"));
        assert_eq!(append_word(text, "files", "zz"), None, "[solo] is no area");
        assert_eq!(append_word(text, "gone", "zz"), None);
    }

    #[test]
    fn only_a_line_without_a_name_filter_gains_the_option() {
        let registry = |checks: &[&str]| {
            let doc = json!({"scans": [{"id": "a", "checks": checks}]});
            serde_json::to_string_pretty(&doc).unwrap() + "\n"
        };
        // A name filter before or after `--`, a module filter, a word this reader does not know
        // and a line that names no `--test` stay as they are.
        for line in [
            "cargo test --offline --locked --test one name",
            "cargo test --offline --locked --test one -- --exact name",
            "cargo test --offline --locked --test card one:: -- two::",
            "cargo test --offline --locked --lib -- name",
            "cargo test --offline --locked --test one --features x",
            "cargo test --offline --locked",
        ] {
            let grown = add_test_option(&registry(&[line]), "a", "two");
            assert_eq!(grown, None, "{line}");
        }
        // The line without a filter gains the name, wherever it stands.
        let filtered = "cargo test --offline --locked --test one name";
        let plain = "cargo test --offline --locked --lib --test three";
        let old = registry(&[filtered, plain]);
        let grown = add_test_option(&old, "a", "two").unwrap();
        assert_eq!(grown, old.replace(plain, &format!("{plain} --test two")));
        assert!(crate::gate::registry_only_grew(
            crate::gate::SCAN_REGISTRY,
            &old,
            &grown
        ));
        assert_eq!(add_test_option(&old, "b", "two"), None);
    }

    /// The scan table of this repository: no scan without `test_targets` gets a name into a line
    /// with a filter, and these four `always` scans, whose one line is filtered, get no write.
    #[test]
    fn no_scan_of_this_repository_gains_a_target_in_a_filtered_line() {
        const TABLE: &str = include_str!("../tools/repo_scans.json");
        for id in [
            "clean-env-only",
            "test-seams-read-only-in-debug-builds",
            "include-str-card-generated-scope",
            "include-str-release-suite",
        ] {
            if !lists_targets(TABLE, id) {
                assert_eq!(add_test_option(TABLE, id, "zz_new_cli"), None, "{id}");
            }
        }
        let value: Value = serde_json::from_str(TABLE).unwrap();
        for scan in value["scans"].as_array().unwrap() {
            let id = scan["id"].as_str().unwrap();
            if lists_targets(TABLE, id) {
                continue;
            }
            let Some(grown) = add_test_option(TABLE, id, "zz_new_cli") else {
                continue;
            };
            let grown: Value = serde_json::from_str(&grown).unwrap();
            let before = scan["checks"].as_array().unwrap();
            let after = scan_of(&grown, id).unwrap()["checks"].as_array().unwrap();
            for (old, _) in before.iter().zip(after).filter(|(o, n)| o != n) {
                let line = old.as_str().unwrap();
                let mut words = line.split_whitespace().skip(2);
                while let Some(word) = words.next() {
                    if word == "--test" || word == "--bin" {
                        words.next();
                    } else {
                        assert!(word.starts_with("--") && word != "--", "{id}: {line}");
                    }
                }
            }
        }
    }

    #[test]
    fn a_scan_with_test_targets_gets_the_name_at_its_place_and_keeps_its_check_lines() {
        // Two scans with the same target list: only the named one changes.
        let entry = |id: &str| {
            json!({
                "id": id,
                "test_targets": ["bb", "dd"],
                "checks": ["cargo test --offline --locked --test card one:: -- two::"],
            })
        };
        let doc = json!({"scans": [entry("a"), entry("b")]});
        let registry = serde_json::to_string_pretty(&doc).unwrap() + "\n";
        let array = |names: &[&str]| {
            let lines: Vec<String> = names.iter().map(|n| format!("        \"{n}\"")).collect();
            format!("\"test_targets\": [\n{}\n      ]", lines.join(",\n"))
        };
        let old_array = array(&["bb", "dd"]);
        for (stem, names) in [
            ("aa", ["aa", "bb", "dd"]),
            ("cc", ["bb", "cc", "dd"]),
            ("ee", ["bb", "dd", "ee"]),
        ] {
            let grown = add_test_option(&registry, "b", stem).unwrap();
            let at = registry.rfind(&old_array).unwrap();
            let want = format!(
                "{}{}{}",
                &registry[..at],
                array(&names),
                &registry[at + old_array.len()..]
            );
            assert_eq!(grown, want, "{stem}: one new line in the second scan only");
            assert!(crate::gate::registry_only_grew(
                crate::gate::SCAN_REGISTRY,
                &registry,
                &grown
            ));
        }
        // A one-line array stays on its line.
        let compact = r#"{"scans":[{"id":"a","test_targets":["bb"]}]}"#;
        assert_eq!(
            add_test_option(compact, "a", "cc").as_deref(),
            Some(r#"{"scans":[{"id":"a","test_targets":["bb","cc"]}]}"#)
        );
        // Listed already, no target name, no such scan: nothing, and never a check line.
        assert_eq!(add_test_option(&registry, "a", "bb"), None);
        assert_eq!(add_test_option(&registry, "a", "Bad"), None);
        assert_eq!(add_test_option(&registry, "z", "cc"), None);
    }
}
