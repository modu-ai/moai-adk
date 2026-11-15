//! #3005 AC-7 and AC-8: Orca is optional for every workflow. These tests read the checked-in
//! tree: the `moai` tool itself never runs `orca`, the word `orca` stays in the two mode
//! workflows and the Orca adapter, every Orca command in a mode sits next to its "without
//! Orca" rule, and the adapter table has a row for every route with one of three words.
//! (Registered in `tools/repo_scans.json` as `orca-stays-optional`.)
use std::{
    collections::BTreeSet,
    fs,
    path::{Path, PathBuf},
};

fn root() -> PathBuf {
    Path::new(env!("CARGO_MANIFEST_DIR")).to_path_buf()
}

fn rel(path: &Path) -> String {
    path.strip_prefix(root())
        .unwrap()
        .to_string_lossy()
        .replace('\\', "/")
}

fn files_under(dir: &Path, exts: &[&str], out: &mut Vec<PathBuf>) {
    let Ok(entries) = fs::read_dir(dir) else {
        return;
    };
    let mut paths: Vec<PathBuf> = entries.flatten().map(|e| e.path()).collect();
    paths.sort();
    for path in paths {
        if path.is_dir() {
            files_under(&path, exts, out);
        } else if path
            .extension()
            .and_then(|e| e.to_str())
            .is_some_and(|e| exts.contains(&e))
        {
            out.push(path);
        }
    }
}

fn read(relative: &str) -> String {
    fs::read_to_string(root().join(relative)).unwrap_or_else(|e| panic!("{relative}: {e}"))
}

/// The body of a `## <title>` section, up to the next `## ` heading.
fn section<'a>(text: &'a str, title: &str) -> &'a str {
    let head = format!("\n## {title}\n");
    let start = text
        .find(&head)
        .unwrap_or_else(|| panic!("a `## {title}` section"))
        + head.len();
    let body = &text[start..];
    &body[..body.find("\n## ").unwrap_or(body.len())]
}

fn mentions_orca(text: &str) -> bool {
    text.to_lowercase().contains("orca")
}

/// The files of `src/` that may say "orca": Orca's own hook entries that a Go-state
/// migration reads (`go_hooks.rs`), a status line command in a settings fixture
/// (`claude.rs`), and the leader round's description of the screen it is given
/// (`leader_round.rs`, `ops/leader_round.rs`). None of them starts the program.
const SRC_ORCA_FILES: [&str; 4] = [
    "src/claude.rs",
    "src/go_hooks.rs",
    "src/leader_round.rs",
    "src/ops/leader_round.rs",
];

/// The only places in the skill package and the agent files that name Orca. The research
/// agent lists the adapter among the documents it reads (checked below to be that line only).
const PACKAGE_ORCA_FILES: [&str; 4] = [
    "agents/research.md",
    "skills/moai/references/adapters/orca.md",
    "skills/moai/workflows/mode-agent.md",
    "skills/moai/workflows/mode-leader.md",
];

/// Words that make an Orca command line safe to read on a machine without Orca.
const NO_ORCA_PHRASES: [&str; 3] = ["without Orca", "no Orca", "neither Orca"];

fn src_files() -> Vec<PathBuf> {
    let mut files = Vec::new();
    files_under(&root().join("src"), &["rs"], &mut files);
    assert!(files.len() > 50, "src walk found {} files", files.len());
    files
}

#[test]
fn nothing_in_src_runs_orca() {
    let mut spawns = Vec::new();
    for file in src_files() {
        let text = fs::read_to_string(&file).unwrap();
        for (at, _) in text.match_indices("Command::new(") {
            let call: String = text[at..].chars().take(120).collect();
            let call = call.split(')').next().unwrap_or("");
            if mentions_orca(call) {
                spawns.push(format!("{}: {call})", rel(&file)));
            }
        }
    }
    assert!(
        spawns.is_empty(),
        "a command of the tool starts orca (Orca stays optional, #3005 AC-8): {spawns:#?}"
    );
}

#[test]
fn the_word_orca_stays_in_the_known_src_files() {
    let found: BTreeSet<String> = src_files()
        .into_iter()
        .filter(|f| mentions_orca(&fs::read_to_string(f).unwrap()))
        .map(|f| rel(&f))
        .collect();
    let expected: BTreeSet<String> = SRC_ORCA_FILES.iter().map(|s| s.to_string()).collect();
    assert_eq!(
        found, expected,
        "a src file began or stopped naming orca: if it runs the program, Orca is no longer \
         optional; if it only names it, add it to SRC_ORCA_FILES"
    );
}

#[test]
fn the_word_orca_stays_in_the_two_modes_and_the_adapter() {
    let mut files = Vec::new();
    files_under(
        &root().join("skills/moai"),
        &["md", "json", "yaml", "yml", "toml", "txt", "sh", "py"],
        &mut files,
    );
    files_under(&root().join("agents"), &["md", "yaml", "toml"], &mut files);
    assert!(files.len() > 20, "package walk found {} files", files.len());
    let found: BTreeSet<String> = files
        .into_iter()
        .map(|f| rel(&f))
        // the manifest lists package paths, so it names the adapter's file name
        .filter(|f| f != "skills/moai/references/source-manifest.json")
        .filter(|f| mentions_orca(&fs::read_to_string(root().join(f)).unwrap()))
        .collect();
    let expected: BTreeSet<String> = PACKAGE_ORCA_FILES.iter().map(|s| s.to_string()).collect();
    assert_eq!(
        found, expected,
        "Orca is named outside the two modes and its adapter: a workflow that needs Orca \
         would stop on a machine without it (#3005 AC-8)"
    );
    let research = read("agents/research.md");
    let naming: Vec<&str> = research.lines().filter(|l| mentions_orca(l)).collect();
    assert!(
        naming.len() == 1 && naming[0].starts_with("reads: "),
        "agents/research.md may name Orca only in its `reads:` line: {naming:?}"
    );
}

#[test]
fn every_orca_command_in_a_mode_sits_next_to_its_rule_for_no_orca() {
    for (file, entry_rule) in [
        (
            "skills/moai/workflows/mode-agent.md",
            "in the autonomous mode (no Orca, or no channel between sessions) go straight to the self-claim round",
        ),
        (
            "skills/moai/workflows/mode-leader.md",
            "Refuse to enter only when there is neither Orca nor a timer",
        ),
    ] {
        let text = read(file);
        let orca_section = section(&text, "Session control with Orca");
        let intro = &orca_section[..orca_section.find("\n1. ").expect("a numbered list")];
        for phrase in [
            "installed is not chosen",
            "Otherwise report the session-control step `not_run`",
        ] {
            assert!(intro.contains(phrase), "{file}: the Orca section lacks `{phrase}`");
        }
        assert!(text.contains(entry_rule), "{file}: lacks `{entry_rule}`");
        let mut bare = Vec::new();
        for line in text.lines() {
            if orca_section.contains(line) || !line.contains("`orca ") {
                continue;
            }
            if !NO_ORCA_PHRASES.iter().any(|p| line.contains(p)) {
                bare.push(line.chars().take(90).collect::<String>());
            }
        }
        assert!(
            bare.is_empty(),
            "{file}: a line outside the Orca section runs an Orca command and says nothing \
             about the case without Orca (#3005 AC-8): {bare:#?}"
        );
    }
}

#[test]
fn the_adapter_table_has_a_row_for_every_route_with_a_status_word() {
    let routes: serde_json::Value =
        serde_json::from_str(&read("skills/moai/references/routes.json")).unwrap();
    let route_names: BTreeSet<String> = routes["routes"]
        .as_object()
        .expect("routes object")
        .keys()
        .cloned()
        .collect();
    assert_eq!(
        route_names.len(),
        12,
        "the plan counts 12 workflows: {route_names:?}"
    );

    let adapter = read("skills/moai/references/adapters/orca.md");
    let missing = section(&adapter, "When Orca is missing");
    let mut rows = BTreeSet::new();
    for line in missing.lines().filter(|l| l.starts_with("| ")) {
        let cells: Vec<&str> = line.trim_matches('|').split(" | ").map(str::trim).collect();
        let name = cells[0];
        if name == "Workflow" || name.starts_with("---") {
            continue;
        }
        assert_eq!(cells.len(), 4, "row `{name}` needs 4 cells: {line}");
        for (column, cell) in [("With Orca", cells[1]), ("Without Orca", cells[2])] {
            assert!(
                ["works", "replaced", "unsupported"]
                    .iter()
                    .any(|word| cell.starts_with(word)),
                "row `{name}`, column {column} must start with works, replaced or unsupported: {cell}"
            );
        }
        // The audit of #3018 (B1): an Orca profile moves the settings folder to
        // `CLAUDE_CONFIG_DIR`, which `moai update` does not look at (Issue #2936), so SETUP and
        // UPDATE are not "works" with Orca.
        if name == "SETUP" || name == "UPDATE" {
            assert!(
                cells[1].starts_with("unsupported") && cells[1].contains("#2936"),
                "row `{name}`, With Orca must start with unsupported and name Issue #2936: {}",
                cells[1]
            );
        }
        assert!(rows.insert(name.to_string()), "row `{name}` appears twice");
    }
    assert_eq!(
        rows, route_names,
        "the table and routes.json list different workflows"
    );
    // The list itself, not the table: a cell elsewhere must not satisfy these.
    let list = missing
        .lines()
        .find(|l| l.starts_with("Unsupported today"))
        .expect("the section has an `Unsupported today` paragraph");
    for cell in [
        "the OMP spelling",
        "an OMP lane's wake",
        "a Codex or OMP leader's periodic run",
        "the settings folder of an Orca profile",
        "Issue #2936",
    ] {
        assert!(list.contains(cell), "the unsupported list lost `{cell}`");
    }
}
