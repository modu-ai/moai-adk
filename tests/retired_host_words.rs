//! SPEC-20261122-remove-codex-shipped-text: what MoAI ships and shows names Claude Code and GLM
//! only. The scan reads named lists of files (no directory walk): the GitHub template manifest's
//! templates, the repository's own issue forms and label files, and the named paths of each landed
//! slice. A retired host word is `codex` anywhere, `omp` or `oh my pi` as a whole word, the Codex
//! entry spelling `$moai` followed by a space or a dash, and the OMP spelling `/skill:moai`. Each
//! later slice adds its own case when its files change.
use moai_cli::labels_check::parse_axes;
use std::fs;
use std::path::{Path, PathBuf};

fn root() -> PathBuf {
    PathBuf::from(env!("CARGO_MANIFEST_DIR"))
}

fn read(rel: &str) -> String {
    fs::read_to_string(root().join(rel)).unwrap_or_else(|e| panic!("{rel}: {e}"))
}

/// True when `word` stands alone: no ASCII letter or digit right before or after it.
fn whole_word(lower: &str, word: &str) -> bool {
    let bytes = lower.as_bytes();
    let mut from = 0;
    while let Some(at) = lower[from..].find(word) {
        let start = from + at;
        let end = start + word.len();
        let before = start == 0 || !bytes[start - 1].is_ascii_alphanumeric();
        let after = end == bytes.len() || !bytes[end].is_ascii_alphanumeric();
        if before && after {
            return true;
        }
        from = start + 1;
    }
    false
}

/// True when the line holds a retired host word.
fn names_a_retired_host(line: &str) -> bool {
    let lower = line.to_ascii_lowercase();
    if lower.contains("codex") || whole_word(&lower, "omp") || lower.contains("oh my pi") {
        return true;
    }
    if line.contains("/skill:moai") {
        return true;
    }
    // The Codex entry spelling: `$moai project`, `$moai --mode-agent`. A shell variable such as
    // `$MOAI_LANGUAGE` (another case) or `"$moai"` (a quote follows) is not one.
    let mut from = 0;
    while let Some(at) = line[from..].find("$moai") {
        let end = from + at + "$moai".len();
        if matches!(line.as_bytes().get(end), Some(b' ') | Some(b'-')) {
            return true;
        }
        from = end;
    }
    false
}

/// Every `file:line: text` of `text` that names a retired host, except the lines `allowed` lists
/// for that file.
fn findings(rel: &str, text: &str, allowed: &[&str]) -> Vec<String> {
    text.lines()
        .enumerate()
        .filter(|(_, line)| names_a_retired_host(line) && !allowed.contains(&line.trim()))
        .map(|(i, line)| format!("{rel}:{}: {}", i + 1, line.trim()))
        .collect()
}

fn scan(paths: &[String], allow: &[(&str, &str)]) -> Vec<String> {
    let mut out = Vec::new();
    for rel in paths {
        let allowed: Vec<&str> = allow
            .iter()
            .filter(|(p, _)| p == rel)
            .map(|(_, line)| *line)
            .collect();
        out.extend(findings(rel, &read(rel), &allowed));
    }
    out
}

/// The GitHub templates the product ships, from the template manifest, and the repository forms
/// they derive from.
fn github_forms() -> Vec<String> {
    let manifest: serde_json::Value =
        serde_json::from_str(&read("templates/github/manifest.json")).unwrap();
    let mut paths = Vec::new();
    for entry in manifest["files"].as_array().unwrap() {
        for key in ["template_path", "derived_from"] {
            if let Some(p) = entry[key].as_str() {
                if root().join(p).is_file() && !paths.iter().any(|q: &String| q == p) {
                    paths.push(p.to_string());
                }
            }
        }
    }
    assert!(
        paths.iter().any(|p| p.ends_with("moai-bugfix.yml")),
        "the manifest names the issue forms: {paths:?}"
    );
    paths
}

/// AC-9: a planted word is reported with its file and line; look-alikes are not.
#[test]
fn a_planted_word_is_reported_with_file_and_line() {
    let text =
        "first line\n    - codex\nrun `$moai project` here\nOh My Pi\nOMP\n/skill:moai update\n";
    assert_eq!(
        findings("x.md", text, &[]),
        [
            "x.md:2: - codex",
            "x.md:3: run `$moai project` here",
            "x.md:4: Oh My Pi",
            "x.md:5: OMP",
            "x.md:6: /skill:moai update",
        ]
    );
    for clean in [
        "compose the prompt",
        "export $MOAI_LANGUAGE",
        "case \"$moai\" in",
        "a promptly omputer",
        "moai cc | moai glm",
    ] {
        assert!(findings("y.md", clean, &[]).is_empty(), "{clean}");
    }
    assert!(findings("z.md", "  - codex  ", &["- codex"]).is_empty());
}

/// AC-4: the issue forms and their product templates offer no Codex choice.
#[test]
fn github_forms_have_no_retired_host_words() {
    let paths = github_forms();
    let found = scan(&paths, &[]);
    assert!(found.is_empty(), "{found:#?}");
}

/// AC-5: `host:codex` is kept as a retired name under `not_used` in this repository's label list,
/// and the product label template offers no `host:codex` label.
#[test]
fn host_codex_is_retired_in_the_label_lists() {
    let axes = parse_axes(&read(".github/moai/labels.yml")).unwrap();
    assert!(axes.not_used.contains("host:codex"), "not_used");
    assert!(!axes.known.contains("host:codex"), "known");
    assert!(axes.known.contains("host:claude"), "host:claude");
    let template = read("templates/github/labels.yml");
    assert!(!template.contains("host:codex"), "template");
    assert!(
        template.contains("- name: host:claude"),
        "template host:claude"
    );
    let readme = read(".github/moai/README.md");
    assert!(findings(".github/moai/README.md", &readme, &[]).is_empty());
}

/// AC-8: the repository tracks no file under its own `.codex` folder.
#[test]
fn the_repository_tracks_no_codex_folder() {
    for rel in [
        ".codex/agents/prj-local-check.toml",
        ".codex/agents/prj-reviewer.toml",
    ] {
        assert!(!Path::new(&root().join(rel)).exists(), "{rel}");
    }
}
