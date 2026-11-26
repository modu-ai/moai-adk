//! The skill-file references of the checked-in tree resolve, and the quality guides match the
//! languages the detector reports. The files are read from the checked-in tree; the only process
//! started is `git ls-files`, through the shared fixture runner.

mod common;

use common::git::git;
use moai_cli::project_detect::{shipped_code_languages, shipped_manifest_languages};
use moai_cli::spec_doc::SPECS_DIR;
use std::collections::{BTreeMap, BTreeSet};
use std::path::PathBuf;

const PR_TEMPLATE: &str = ".github/pull_request_template.md";
/// Folders under `docs/` whose files are dated records (design reviews, surveys).
const HISTORY_DOCS: [&str; 2] = ["design-review", "surveys"];
/// Shipped files name a workflow relative to the skill (`workflows/run.md`), not by its
/// repository path, so the repository references alone miss them.
const SHIPPED: [&str; 2] = ["skills/moai/", "agents/"];
const QUALITY: &str = "skills/moai/references/internal/quality.md";
const LANGUAGES: &str = "skills/moai/references/languages";
const GUIDE_LIMIT: usize = 4096;
/// Detected languages that have no guide yet. A name leaves this list when its guide is written;
/// the test below fails on a name that has a row or that the detector no longer reports.
const NO_GUIDE_YET: [&str; 1] = ["lua"];
/// Files that tests write into a temporary project, by the test that names them (each path
/// without its `skills/` head, so that this list is no reference itself). Only these pairs may
/// miss in the checked-in tree, and a pair that no longer occurs fails the scan.
const FIXTURE_PATHS: [(&str, &str); 8] = [
    ("tests/check_affected_modes_cli.rs", "moai/workflows/x.md"),
    ("tests/check_fix_cli.rs", "moai/references/a.md"),
    ("tests/check_fix_cli.rs", "moai/references/new.md"),
    ("tests/e2e_temp_home_cli.rs", "mine/SKILL.md"),
    ("tests/repo_scans_table.rs", "moai/commented.md"),
    ("tests/skill_generalization_scan.rs", "moai/clean.md"),
    (
        "tests/skill_generalization_scan.rs",
        "moai/references/z.json",
    ),
    ("tests/skill_generalization_scan.rs", "moai/x.md"),
];

fn root() -> PathBuf {
    PathBuf::from(env!("CARGO_MANIFEST_DIR"))
}

fn read(rel: &str) -> String {
    std::fs::read_to_string(root().join(rel)).unwrap_or_else(|e| panic!("{rel}: {e}"))
}

/// Every tracked file that reads as UTF-8, with its path from the repository root.
fn tracked_texts() -> Vec<(String, String)> {
    git(&root(), &["ls-files", "-z"])
        .split('\0')
        .filter(|name| !name.is_empty())
        .filter_map(|name| {
            let bytes = std::fs::read(root().join(name)).unwrap_or_else(|e| panic!("{name}: {e}"));
            String::from_utf8(bytes)
                .ok()
                .map(|text| (name.to_string(), text))
        })
        .collect()
}

fn name_byte(c: u8) -> bool {
    c.is_ascii_alphanumeric() || c == b'_' || c == b'.' || c == b'-'
}

/// The `skills/<package>/…/<file>.<extension>` references of a line, left to right and without
/// overlap, each with the byte it starts at.
fn references(line: &str) -> Vec<(usize, &str)> {
    let mut out = Vec::new();
    let mut from = 0;
    while let Some(at) = line[from..].find("skills/") {
        let start = from + at;
        match reference_end(line.as_bytes(), start + "skills/".len()) {
            Some(end) => {
                out.push((start, &line[start..end]));
                from = end;
            }
            None => from = start + 1,
        }
    }
    out
}

/// Where the longest reference whose package folder starts at `at` ends: the last path segment
/// behind a folder that reads as a file name.
fn reference_end(bytes: &[u8], at: usize) -> Option<usize> {
    let mut start = at;
    let mut folders = 0;
    let mut end = None;
    loop {
        let len = bytes[start..].iter().take_while(|c| name_byte(**c)).count();
        if len == 0 {
            return end;
        }
        if folders > 0 {
            if let Some(file) = file_end(&bytes[start..start + len]) {
                end = Some(start + file);
            }
        }
        if bytes.get(start + len) != Some(&b'/') {
            return end;
        }
        folders += 1;
        start += len + 1;
    }
}

/// Where `<name>.<extension>` ends inside one path segment: behind the extension of the last dot
/// that has a name before it and an extension behind it.
fn file_end(segment: &[u8]) -> Option<usize> {
    (1..segment.len()).rev().find_map(|dot| {
        if segment[dot] != b'.' {
            return None;
        }
        let extension = segment[dot + 1..]
            .iter()
            .take_while(|c| **c != b'.')
            .count();
        (extension > 0).then_some(dot + 1 + extension)
    })
}

#[derive(Debug, PartialEq, Eq)]
enum Kind {
    /// A mirror of an earlier ledger or inventory that describes another tree.
    LedgerSnapshot,
    /// A pinned measurement with dated notes, not an instruction to load a package file.
    DatedInventory,
    /// A closed SPEC or a dated document records a past state.
    History,
    /// A path of a host folder or of a temporary fixture, not of this repository.
    RuntimeOrExample,
    Current,
}

fn kind(name: &str, line: &str, start: usize) -> Kind {
    // The earlier state folder, spelled at run time so that no source line holds its name.
    if name.starts_with(&format!(".{}-{}/", "moai", "cli")) {
        return Kind::LedgerSnapshot;
    }
    if name == "docs/product-naming-inventory.md" {
        return Kind::DatedInventory;
    }
    let dated_doc = name.strip_prefix("docs/").is_some_and(|rest| {
        rest.split_once('/')
            .is_some_and(|(folder, _)| HISTORY_DOCS.contains(&folder))
            || (rest.starts_with("decisions-") && rest.ends_with(".md") && !rest.contains('/'))
    });
    let spec_record = name
        .strip_prefix(SPECS_DIR)
        .is_some_and(|rest| rest.starts_with('/'));
    if spec_record || dated_doc {
        return Kind::History;
    }
    let before = &line[..start];
    if before.ends_with('/') && !before.contains("github.com/modu-ai/moai-adk/blob/") {
        return Kind::RuntimeOrExample;
    }
    Kind::Current
}

/// The workflow file names (`workflows/<name>.md`) a line gives relative to the skill: the word
/// starts there, it is not the tail of a longer path.
fn workflow_names(line: &str) -> Vec<&str> {
    line.match_indices("workflows/")
        .filter_map(|(at, word)| {
            if at > 0 {
                let before = line.as_bytes()[at - 1];
                if name_byte(before) || before == b'/' {
                    return None;
                }
            }
            let rest = &line[at + word.len()..];
            let len = rest.bytes().take_while(|c| name_byte(*c)).count();
            let end = rest[..len].rfind(".md").filter(|dot| *dot > 0)? + ".md".len();
            Some(&rest[..end])
        })
        .collect()
}

#[test]
fn a_reference_is_read_to_its_last_file_name_and_sorted_by_where_it_stands() {
    // The paths are put together at run time: the scan below reads this file too.
    let skills = "skills";
    let line = format!(
        "see `{skills}/moai/workflows/run.md`, {skills}/moai/a.b.json. and {skills}/moai only"
    );
    let found: Vec<&str> = references(&line).into_iter().map(|(_, r)| r).collect();
    assert_eq!(
        found,
        [
            format!("{skills}/moai/workflows/run.md"),
            format!("{skills}/moai/a.b.json"),
        ]
    );
    let cut = format!("x-{skills}/p/f.md/folder {skills}//f.md {skills}/p/.hidden");
    assert_eq!(references(&cut), [(2, format!("{skills}/p/f.md").as_str())]);

    let sorted = |name: &str, line: &str| {
        let (start, _) = references(line)[0];
        kind(name, line, start)
    };
    let plain = format!("{skills}/p/f.md");
    assert_eq!(sorted("notes.md", &plain), Kind::Current);
    assert_eq!(sorted("docs/decisions-2026.md", &plain), Kind::History);
    assert_eq!(sorted("docs/surveys/a.md", &plain), Kind::History);
    assert_eq!(sorted("docs/surveys.md", &plain), Kind::Current);
    assert_eq!(
        sorted(&format!("{SPECS_DIR}/S/spec.md"), &plain),
        Kind::History
    );
    assert_eq!(
        sorted("docs/product-naming-inventory.md", &plain),
        Kind::DatedInventory
    );
    assert_eq!(
        sorted(&format!(".{}-{}/x.md", "moai", "cli"), &plain),
        Kind::LedgerSnapshot
    );
    let host = format!("~/.claude/{plain}");
    assert_eq!(sorted("notes.md", &host), Kind::RuntimeOrExample);
    let blob = format!("https://github.com/modu-ai/moai-adk/blob/main/{plain}");
    assert_eq!(sorted("notes.md", &blob), Kind::Current);

    assert_eq!(
        workflow_names("read workflows/run.md and `workflows/a-b.md`."),
        ["run.md", "a-b.md"]
    );
    assert!(workflow_names("x/workflows/run.md .workflows/run.md workflows/.md").is_empty());
}

#[test]
fn the_pr_template_names_a_workflow_that_exists() {
    let text = read(PR_TEMPLATE);
    let named: Vec<&str> = text
        .lines()
        .flat_map(references)
        .map(|(_, reference)| reference)
        .filter(|reference| reference.contains("/workflows/"))
        .collect();
    assert!(
        !named.is_empty(),
        "the PR template must identify its check rules"
    );
    for reference in named {
        assert!(root().join(reference).is_file(), "{reference}");
    }
}

#[test]
fn every_current_skill_file_reference_resolves() {
    let mut seen = 0;
    let mut broken = Vec::new();
    let mut fixtures = BTreeSet::new();
    for (name, text) in tracked_texts() {
        for (number, line) in text.lines().enumerate() {
            for (start, reference) in references(line) {
                if kind(&name, line, start) != Kind::Current {
                    continue;
                }
                seen += 1;
                if root().join(reference).is_file() {
                    continue;
                }
                let fixture = FIXTURE_PATHS.iter().find(|(file, path)| {
                    *file == name && reference.strip_prefix("skills/") == Some(*path)
                });
                match fixture {
                    Some(pair) => {
                        fixtures.insert(*pair);
                    }
                    None => broken.push(format!("{name}:{}: {reference}", number + 1)),
                }
            }
        }
    }
    assert!(
        seen > 0,
        "the tracked reference inventory must not be empty"
    );
    assert!(
        broken.is_empty(),
        "current repository skill references are broken:\n{}",
        broken.join("\n")
    );
    assert_eq!(
        fixtures,
        BTreeSet::from(FIXTURE_PATHS),
        "a pair of FIXTURE_PATHS no longer occurs or its file exists now: take it off the list"
    );
}

#[test]
fn shipped_files_name_only_existing_workflows() {
    let mut seen = 0;
    let mut broken = Vec::new();
    for (name, text) in tracked_texts() {
        if !SHIPPED.iter().any(|prefix| name.starts_with(prefix)) {
            continue;
        }
        for (number, line) in text.lines().enumerate() {
            for workflow in workflow_names(line) {
                seen += 1;
                if !root()
                    .join("skills/moai/workflows")
                    .join(workflow)
                    .is_file()
                {
                    broken.push(format!("{name}:{}: workflows/{workflow}", number + 1));
                }
            }
        }
    }
    assert!(
        seen > 0,
        "the shipped workflow-name inventory must not be empty"
    );
    assert!(
        broken.is_empty(),
        "shipped files name a workflow that does not exist:\n{}",
        broken.join("\n")
    );
}

/// The `| <language> | languages/<guide>.md |` rows of internal/quality.
fn guide_rows() -> BTreeMap<String, String> {
    read(QUALITY)
        .lines()
        .filter_map(|line| {
            let (language, guide) = line
                .strip_prefix("| ")?
                .strip_suffix(" |")?
                .split_once(" | ")?;
            let file = guide.strip_prefix("languages/")?;
            let stem = file.strip_suffix(".md")?;
            let plain = !language.is_empty()
                && language.bytes().all(|c| c.is_ascii_lowercase())
                && !stem.is_empty()
                && stem.bytes().all(|c| c.is_ascii_lowercase() || c == b'-');
            plain.then(|| (language.to_string(), file.to_string()))
        })
        .collect()
}

/// Every guide file of the languages folder.
fn guide_files() -> BTreeSet<String> {
    std::fs::read_dir(root().join(LANGUAGES))
        .unwrap()
        .map(|entry| entry.unwrap().file_name().to_string_lossy().into_owned())
        .filter(|name| name.ends_with(".md"))
        .collect()
}

#[test]
fn every_detected_language_has_a_guide_row() {
    let mut detected: BTreeSet<String> = shipped_manifest_languages().into_iter().collect();
    detected.extend(shipped_code_languages());
    assert!(detected.contains("rust"), "{detected:?}");
    let rows = guide_rows();
    for gap in NO_GUIDE_YET {
        assert!(
            detected.remove(gap) && !rows.contains_key(gap),
            "`{gap}` is no gap any more: take it off NO_GUIDE_YET"
        );
    }
    let named: BTreeSet<String> = rows.into_keys().collect();
    assert_eq!(named, detected, "{QUALITY} names one guide per language");
}

#[test]
fn every_named_guide_exists_and_every_guide_is_named() {
    let named: BTreeSet<String> = guide_rows().into_values().collect();
    assert_eq!(named, guide_files());
}

#[test]
fn guides_are_short_and_point_back_to_quality() {
    let files = guide_files();
    assert!(!files.is_empty());
    for file in files {
        let text = read(&format!("{LANGUAGES}/{file}"));
        assert!(text.len() <= GUIDE_LIMIT, "{file} has {} bytes", text.len());
        assert!(text.starts_with("# Language: "), "{file}");
        assert!(
            text.contains("When to read: `internal/quality` sent you here"),
            "{file}"
        );
    }
}
