//! Generalization scan of the shipped skill texts that no other change rewrites: the adapter
//! note, the git-safety, commit-trailers and loop references, the output style and the ten
//! language texts. A user project reads them, so they carry no retired host, no Issue, pull
//! request or card number, no internal tool path and no empty cell in a defaults table; the
//! manifest's origin reasons carry no retired host either.
//!
//! The texts are read from the checked-in tree; no process is started.

use std::fs;
use std::path::{Path, PathBuf};

fn package() -> PathBuf {
    Path::new(env!("CARGO_MANIFEST_DIR")).join("skills/moai")
}

/// The files this scan covers, relative to `skills/moai/references`.
fn covered() -> Vec<String> {
    let mut files: Vec<String> = [
        "adapters/orca.md",
        "internal/git-safety.md",
        "internal/commit-trailers.md",
        "internal/loop.md",
        "home/output-style.md",
    ]
    .iter()
    .map(|s| s.to_string())
    .collect();
    let mut languages: Vec<String> = fs::read_dir(package().join("references/languages"))
        .unwrap()
        .map(|e| e.unwrap().file_name().to_string_lossy().into_owned())
        .filter(|n| n.ends_with(".md"))
        .map(|n| format!("languages/{n}"))
        .collect();
    languages.sort();
    assert_eq!(languages.len(), 10, "the language texts: {languages:?}");
    files.extend(languages);
    files
}

fn read(rel: &str) -> String {
    fs::read_to_string(package().join("references").join(rel))
        .unwrap_or_else(|e| panic!("{rel}: {e}"))
}

/// The name of the state folder an earlier generation used. It is built at run time so that this
/// file does not itself spell a state folder name.
fn old_state_folder() -> String {
    format!(".{}-cli", "moai")
}

/// The name of a project's state folder, built at run time for the same reason.
fn state_folder() -> String {
    format!(".{}", "moai")
}

/// The words of a text, split at everything that is not a letter or a digit.
fn words(text: &str) -> Vec<&str> {
    text.split(|c: char| !c.is_ascii_alphanumeric())
        .filter(|w| !w.is_empty())
        .collect()
}

#[test]
fn no_retired_host_or_internal_tool_path_in_the_covered_texts() {
    for rel in covered() {
        let text = read(&rel);
        for banned in ["Kiro", "Cursor", "Codex", "OMP"] {
            assert!(
                !words(&text).contains(&banned),
                "{rel} names the retired host {banned}"
            );
        }
        for banned in [
            "Issue #",
            "tools/",
            old_state_folder().as_str(),
            "Go generation",
        ] {
            assert!(!text.contains(banned), "{rel} contains {banned:?}");
        }
    }
}

#[test]
fn no_issue_pull_request_or_card_number_in_the_covered_texts() {
    for rel in covered() {
        let text = read(&rel);
        let bytes = text.as_bytes();
        for i in 0..bytes.len().saturating_sub(4) {
            if bytes[i] == b'#' && bytes[i + 1..i + 5].iter().all(u8::is_ascii_digit) {
                panic!("{rel} names an Issue or pull request number at byte {i}");
            }
        }
        for i in 0..bytes.len().saturating_sub(2) {
            let starts_word = i == 0 || !bytes[i - 1].is_ascii_alphanumeric();
            if starts_word
                && bytes[i] == b'C'
                && bytes[i + 1] == b'-'
                && bytes[i + 2].is_ascii_digit()
            {
                panic!("{rel} names a card number at byte {i}");
            }
        }
        assert!(!text.contains("SPEC-2026"), "{rel} names a SPEC id");
    }
}

#[test]
fn no_table_of_the_covered_texts_has_an_empty_cell() {
    for rel in covered() {
        let text = read(&rel);
        for (n, line) in text.lines().enumerate() {
            let row = line.trim();
            if !row.starts_with('|') || !row.ends_with('|') {
                continue;
            }
            let inner = &row[1..row.len() - 1];
            let cells: Vec<&str> = inner.split('|').map(str::trim).collect();
            let separator = cells
                .iter()
                .all(|c| !c.is_empty() && c.chars().all(|ch| ch == '-' || ch == ':'));
            if separator {
                continue;
            }
            assert!(
                cells.iter().all(|c| !c.is_empty()),
                "{rel}:{}: a table row has an empty cell (write `n/a: <reason>`): {line}",
                n + 1
            );
        }
    }
}

#[test]
fn the_manifest_origin_reasons_name_no_retired_host_or_route() {
    let manifest: serde_json::Value = serde_json::from_str(
        &fs::read_to_string(package().join("references/source-manifest.json")).unwrap(),
    )
    .unwrap();
    let files = manifest["files"].as_object().expect("`files` is an object");
    for (path, entry) in files {
        let reason = entry["origin"]["reason"].as_str().unwrap_or_default();
        for banned in ["Kiro", "Cursor", "Codex", "OMP"] {
            assert!(
                !words(reason).contains(&banned),
                "{path}: the reason names the retired host {banned}"
            );
        }
        for banned in [
            old_state_folder().as_str(),
            "--mode-agent",
            "--mode-leader",
            "OPTIMIZE",
            "tools/",
        ] {
            assert!(
                !reason.contains(banned),
                "{path}: the reason contains {banned:?}"
            );
        }
    }
}

// ---------------------------------------------------------------------------------------------
// No shipped file names a repository `tools/` script. Every file under `skills/` and `agents/`
// is read, whatever its extension; a user project has no `tools/` folder of this repository, so
// a text that names one of its scripts points at nothing.
// ---------------------------------------------------------------------------------------------

/// Mentions the scan tolerates, as `relative/path: mention`. Empty on purpose: a shipped file
/// names no repository script, so nothing is excused. A new entry needs its own decision.
const ALLOWED: &[&str] = &[];

/// The bytes that may stand inside a path and so may not stand right before `tools/`.
fn is_path_byte(b: u8) -> bool {
    b.is_ascii_alphanumeric() || matches!(b, b'_' | b'.' | b'/' | b'-')
}

/// Every `tools/<path>.<ext>` of a text, with `<ext>` one of `sh`, `py` or `json`. The byte before
/// `tools/` must not be a path byte (the start of the text counts as allowed), so
/// a script under a project's own support folder, `golang.org/x/tools/gopls` and `mytools/x.sh`
/// are not mentions.
fn tools_script_mentions(text: &str) -> Vec<String> {
    const PREFIX: &str = "tools/";
    let bytes = text.as_bytes();
    let mut found = Vec::new();
    let mut from = 0;
    while let Some(offset) = text[from..].find(PREFIX) {
        let start = from + offset;
        from = start + PREFIX.len();
        if start > 0 && is_path_byte(bytes[start - 1]) {
            continue;
        }
        let mut end = from;
        while end < bytes.len() && is_path_byte(bytes[end]) {
            end += 1;
        }
        // A sentence may end right after the name (`tools/a.sh.`).
        let run = text[start..end].trim_end_matches('.');
        let path = &run[PREFIX.len()..];
        let named = [".sh", ".py", ".json"]
            .iter()
            .any(|ext| path.len() > ext.len() && path.ends_with(ext));
        if named {
            found.push(run.to_string());
        }
    }
    found
}

/// Every file under `<root>/skills` and `<root>/agents`, as `/`-separated paths relative to
/// `root`, sorted. Nothing is skipped by extension; a link is never followed and never listed.
fn walked_files(root: &Path) -> Vec<String> {
    fn walk(dir: &Path, rel: &str, out: &mut Vec<String>) {
        let entries = match fs::read_dir(dir) {
            Ok(entries) => entries,
            Err(e) if e.kind() == std::io::ErrorKind::NotFound => return,
            Err(e) => panic!("{}: {e}", dir.display()),
        };
        for entry in entries {
            let entry = entry.unwrap();
            let path = entry.path();
            let name = entry.file_name().to_string_lossy().into_owned();
            let rel = format!("{rel}/{name}");
            let meta = fs::symlink_metadata(&path).unwrap();
            if meta.file_type().is_symlink() {
                continue;
            }
            if meta.is_dir() {
                walk(&path, &rel, out);
            } else if meta.is_file() {
                out.push(rel);
            }
        }
    }
    let mut files = Vec::new();
    for top in ["skills", "agents"] {
        walk(&root.join(top), top, &mut files);
    }
    files.sort();
    files
}

/// `relative/path: mention` for every `tools/` script mention under `root` that `ALLOWED` does
/// not excuse.
fn scan_tree(root: &Path) -> Vec<String> {
    let mut hits = Vec::new();
    for rel in walked_files(root) {
        let bytes = fs::read(root.join(&rel)).unwrap_or_else(|e| panic!("{rel}: {e}"));
        for mention in tools_script_mentions(&String::from_utf8_lossy(&bytes)) {
            let hit = format!("{rel}: {mention}");
            if !ALLOWED.contains(&hit.as_str()) {
                hits.push(hit);
            }
        }
    }
    hits
}

#[test]
fn no_shipped_file_names_a_tools_script() {
    let hits = scan_tree(Path::new(env!("CARGO_MANIFEST_DIR")));
    assert!(
        hits.is_empty(),
        "a shipped file names a repository tools/ script:\n{}",
        hits.join("\n")
    );
}

#[test]
fn the_allow_list_is_empty() {
    assert!(ALLOWED.is_empty());
}

#[test]
fn the_matcher_finds_a_tools_script_in_each_written_form() {
    assert_eq!(
        tools_script_mentions("run `tools/a.sh` now"),
        ["tools/a.sh"]
    );
    assert_eq!(
        tools_script_mentions("tools/smoke/other.sh"),
        ["tools/smoke/other.sh"]
    );
    assert_eq!(tools_script_mentions("see (tools/b.py)"), ["tools/b.py"]);
    assert_eq!(
        tools_script_mentions("tools/some_table.json"),
        ["tools/some_table.json"]
    );
    assert_eq!(
        tools_script_mentions("ends the sentence with tools/a.sh."),
        ["tools/a.sh"]
    );
    assert_eq!(
        tools_script_mentions("tools/a.sh and tools/b.py"),
        ["tools/a.sh", "tools/b.py"]
    );
}

#[test]
fn the_matcher_misses_what_is_not_a_repository_tools_script() {
    let project_script = format!("{}/tools/x.sh", state_folder());
    for text in [
        project_script.as_str(),
        "golang.org/x/tools/gopls@latest",
        "mytools/x.sh",
        "a-tools/x.sh",
        "a_tools/x.py",
        "tools/",
        "the tools/ folder",
        "tools/readme.md",
        "tools/a.sh5",
        "tools/.sh",
        "tools/*.sh",
    ] {
        assert!(
            tools_script_mentions(text).is_empty(),
            "{text:?} matched: {:?}",
            tools_script_mentions(text)
        );
    }
}

#[test]
fn the_scan_finds_every_planted_mention_and_not_the_clean_file() {
    let dir = tempfile::TempDir::new().unwrap();
    let root = dir.path();
    let clean = format!(
        "uses {}/tools/ok.sh and golang.org/x/tools/gopls\n",
        state_folder()
    );
    for (rel, body) in [
        ("skills/moai/x.md", "run `tools/a.sh` first\n"),
        ("agents/y.md", "see (tools/b.py)\n"),
        (
            "skills/moai/references/z.json",
            "{\"gen\": \"tools/c.json\"}\n",
        ),
        ("skills/moai/clean.md", clean.as_str()),
    ] {
        let path = root.join(rel);
        fs::create_dir_all(path.parent().unwrap()).unwrap();
        fs::write(path, body).unwrap();
    }
    let mut hits = scan_tree(root);
    hits.sort();
    assert_eq!(
        hits,
        [
            "agents/y.md: tools/b.py",
            "skills/moai/references/z.json: tools/c.json",
            "skills/moai/x.md: tools/a.sh",
        ]
    );
}

#[cfg(unix)]
#[test]
fn the_walk_does_not_follow_a_link() {
    let dir = tempfile::TempDir::new().unwrap();
    let root = dir.path();
    fs::create_dir_all(root.join("skills")).unwrap();
    fs::create_dir_all(root.join("outside")).unwrap();
    fs::write(root.join("outside/o.md"), "tools/a.sh\n").unwrap();
    std::os::unix::fs::symlink(root.join("outside"), root.join("skills/dir-link")).unwrap();
    std::os::unix::fs::symlink(root.join("outside/o.md"), root.join("skills/file-link.md"))
        .unwrap();
    assert!(walked_files(root).is_empty());
    assert!(scan_tree(root).is_empty());
}

#[test]
fn the_walk_lists_the_agent_and_the_manifest_files_of_the_real_tree() {
    let files = walked_files(Path::new(env!("CARGO_MANIFEST_DIR")));
    for expected in [
        "agents/impl.md",
        "skills/moai/references/source-manifest.json",
    ] {
        assert!(
            files.iter().any(|f| f == expected),
            "the walk does not list {expected}"
        );
    }
}
