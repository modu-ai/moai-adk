//! SPEC-20261126-affected-binary-tests: `check affected` also picks the integration tests that
//! drive the built binary. A marker line joins every source change, a changed command file reaches
//! the tests that name its command, a changed library file reaches the tests of its own command
//! files, and a changed file under `tools/` reaches the tests that hold its path. The selection is
//! read from the built binary (`--list`) in temporary git repositories with a temporary HOME; the
//! real repository is only read.
mod common;
use common::git::{git, git_ok};
use common::home::TempHome;
use serde_json::Value;
use std::collections::BTreeSet;
use std::fs;
use std::path::Path;
use std::process::Stdio;

/// The whole line that joins a test to every source change (REQ-1).
const MARK: &str = "// check-affected: every-source-change";

/// The name of the attribute that loads a helper file. It is assembled in [`loads`], so this file
/// itself carries no such attribute inside a string.
const ATTRIBUTE: &str = "path";

/// An area table that keeps the card set of every fixture path below the audit floor: no fixture
/// test below starts with the word of a changed command or library file.
const AREAS: &str = "[areas]\nrepo = [\"project\"]\ncard = [\"trailers\"]\nupdate = [\"update\"]\n\n[solo]\nfiles = []\n";

/// The three selections: the answer's `selection` word and the flags that ask for it.
const PICKS: [(&str, &[&str]); 3] = [("full", &[]), ("card", &["--card"]), ("lane", &["--lane"])];

/// The test files of this repository that must carry [`MARK`] (REQ-2).
const MARKED: [&str; 5] = [
    "go_project_real_run_cli",
    "go_settings_cli",
    "go_state_preview_cli",
    "project_update_hooks_cli",
    "update_end_to_end_cli",
];

/// A temporary git repository and the temporary HOME its `check affected` runs under.
struct Repo {
    dir: tempfile::TempDir,
    home: TempHome,
}

/// One accepted `--list` answer.
struct Listed {
    selection: String,
    /// The word of what the scan list gave (`none`, `listed:<n>` or `unreadable`).
    scans: String,
    targets: BTreeSet<String>,
}

impl Repo {
    fn init() -> Repo {
        let repo = Repo {
            dir: tempfile::tempdir().unwrap(),
            home: TempHome::new(),
        };
        let p = repo.dir.path();
        git_ok(p, &["init", "-q", "-b", "main"]);
        git_ok(p, &["config", "user.email", "t@example.invalid"]);
        git_ok(p, &["config", "user.name", "t"]);
        git_ok(p, &["config", "core.excludesFile", "/dev/null"]);
        repo.put("tests/areas.toml", AREAS);
        repo
    }

    /// A temporary repository whose one commit holds this checkout's present files, as far as the
    /// selection reads them. Nothing is written in the real repository.
    fn of_this_repository() -> Repo {
        let real = Path::new(env!("CARGO_MANIFEST_DIR"));
        let repo = Repo::init();
        let mut files = vec![
            "Cargo.toml".to_string(),
            "tests/areas.toml".to_string(),
            "tools/repo_scans.json".to_string(),
        ];
        rust_files(real, "src", &mut files);
        rust_files(real, "tests", &mut files);
        for rel in &files {
            let to = repo.dir.path().join(rel);
            fs::create_dir_all(to.parent().unwrap()).unwrap();
            fs::copy(real.join(rel), to).unwrap();
        }
        repo.commit();
        repo
    }

    fn put(&self, rel: &str, text: &str) {
        let path = self.dir.path().join(rel);
        fs::create_dir_all(path.parent().unwrap()).unwrap();
        fs::write(path, text).unwrap();
    }

    /// Commits everything and answers the commit.
    fn commit(&self) -> String {
        let p = self.dir.path();
        git_ok(p, &["add", "-A"]);
        git_ok(p, &["commit", "-q", "-m", "step"]);
        git(p, &["rev-parse", "HEAD"])
    }

    /// One commit on top of `HEAD` that appends a line to each of `paths`; answers its range.
    fn change(&self, paths: &[&str]) -> String {
        let base = git(self.dir.path(), &["rev-parse", "HEAD"]);
        for rel in paths {
            let mut text = fs::read_to_string(self.dir.path().join(rel)).unwrap_or_default();
            text.push_str("\n// the fixture's change\n");
            self.put(rel, &text);
        }
        let head = self.commit();
        format!("{base}..{head}")
    }

    /// `check affected <range> --list`: whether it answered `ok`, and the whole answer.
    fn ask(&self, range: &str, flags: &[&str]) -> (bool, Value) {
        let dir = self.dir.path().to_string_lossy().into_owned();
        let mut args = vec!["check", "affected", range, "--list", "--dir", dir.as_str()];
        args.extend_from_slice(flags);
        let out = self
            .home
            .command(&args)
            .env("PATH", std::env::var_os("PATH").unwrap_or_default())
            .stdin(Stdio::null())
            .output()
            .unwrap();
        let text = String::from_utf8_lossy(&out.stdout).to_string();
        let v: Value = serde_json::from_str(&text).unwrap_or_else(|_| {
            panic!(
                "no JSON answer: {text} / {}",
                String::from_utf8_lossy(&out.stderr)
            )
        });
        (v["ok"] == Value::Bool(true), v)
    }

    /// [`Repo::ask`] for a selection that must be accepted.
    fn listed(&self, range: &str, flags: &[&str]) -> Listed {
        let (ok, v) = self.ask(range, flags);
        assert!(ok, "check affected refused the fixture: {v}");
        let result = &v["result"];
        Listed {
            selection: result["selection"].as_str().unwrap().to_string(),
            scans: result["scans"].as_str().unwrap().to_string(),
            targets: result["rust_targets"]
                .as_array()
                .unwrap()
                .iter()
                .map(|s| s.as_str().unwrap().to_string())
                .collect(),
        }
    }
}

/// A test source that loads `helper` with a [`ATTRIBUTE`] attribute and holds nothing else.
fn loads(helper: &str) -> String {
    format!("#[{ATTRIBUTE} = \"{helper}\"]\nmod helper;\nfn main() {{}}\n")
}

/// The source of a command file that calls `item` of the library crate.
fn calls(item: &str) -> String {
    format!("pub fn run() {{ {}::{item}(); }}\n", "moai_cli")
}

/// Every `.rs` file below the folder `rel` of `root`, as paths relative to `root`. Links are not
/// followed.
fn rust_files(root: &Path, rel: &str, out: &mut Vec<String>) {
    for entry in fs::read_dir(root.join(rel)).unwrap() {
        let entry = entry.unwrap();
        let name = entry.file_name().to_string_lossy().into_owned();
        let child = format!("{rel}/{name}");
        let kind = entry.file_type().unwrap();
        if kind.is_dir() {
            rust_files(root, &child, out);
        } else if kind.is_file() && name.ends_with(".rs") {
            out.push(child);
        }
    }
}

// SPEC-20261126-affected-binary-tests AC-1
#[test]
fn a_source_change_selects_the_marked_binary_tests() {
    let repo = Repo::init();
    repo.put(
        "tests/update_marked_cli.rs",
        &format!("//! A test.\n{MARK}\nfn main() {{}}\n"),
    );
    // The marker as part of a longer line, and a test without it.
    let look_alikes = [
        (
            "update_mark_trailing_cli",
            format!("{MARK} (not the line)\nfn main() {{}}\n"),
        ),
        (
            "update_mark_longer_word_cli",
            format!("{MARK}s\nfn main() {{}}\n"),
        ),
        (
            "update_mark_doc_cli",
            format!("//! {MARK}\nfn main() {{}}\n"),
        ),
        (
            "update_mark_literal_cli",
            format!("const LINE: &str = \"{MARK}\";\nfn main() {{}}\n"),
        ),
        ("update_unmarked_cli", "fn main() {}\n".to_string()),
    ];
    for (name, text) in &look_alikes {
        repo.put(&format!("tests/{name}.rs"), text);
    }
    // Source files of every shape; the last one is on the audit floor.
    let sources = [
        "src/lonely.rs",
        "src/nested/part.rs",
        "src/nested/mod.rs",
        "src/lib.rs",
        "src/main.rs",
        "src/ops/lonely_op.rs",
        "src/gate.rs",
    ];
    // Documents, a file under `src/` that is no `.rs`, and a `.rs` file outside `src/`.
    let others = ["docs/note.md", "src/data.json", "tools/helper.rs"];
    for rel in sources.iter().chain(others.iter()) {
        repo.put(rel, "// a\n");
    }
    repo.commit();

    let mut misses: Vec<String> = Vec::new();
    for path in sources {
        let range = repo.change(&[path]);
        for (pick, flags) in PICKS {
            let got = repo.listed(&range, flags);
            // A card request for a floor path answers the full set.
            let answered = if pick == "card" && path == "src/gate.rs" {
                "full"
            } else {
                pick
            };
            assert_eq!(got.selection, answered, "{path} {pick}");
            if !got.targets.contains("update_marked_cli") {
                misses.push(format!("{path} {pick}: missing update_marked_cli"));
            }
            for (name, _) in &look_alikes {
                assert!(!got.targets.contains(*name), "{path} {pick}: {name}");
            }
        }
    }
    for path in others {
        let range = repo.change(&[path]);
        for (pick, flags) in PICKS {
            let got = repo.listed(&range, flags);
            assert_eq!(got.selection, pick, "{path}");
            assert!(
                !got.targets.contains("update_marked_cli"),
                "{path} {pick}: {:?}",
                got.targets
            );
        }
    }
    assert!(misses.is_empty(), "{misses:#?}");
}

// SPEC-20261126-affected-binary-tests AC-2
#[test]
fn a_command_file_change_selects_the_tests_that_name_the_command() {
    let repo = Repo::init();
    for rel in [
        "src/ops/project_apply.rs",
        "src/ops/mod.rs",
        "src/project_apply.rs",
    ] {
        repo.put(rel, "// a\n");
    }
    repo.put(
        "tests/update_own_literal_cli.rs",
        "fn main() { let _ = [\"project-apply\"]; }\n",
    );
    repo.put("tests/update_helper_cli.rs", &loads("common/runner.rs"));
    repo.put(
        "tests/common/runner.rs",
        "pub const COMMAND: &str = \"project-apply\";\n",
    );
    repo.put(
        "tests/update_applied_cli.rs",
        "fn main() { let _ = [\"project-applied\"]; }\n",
    );
    repo.put(
        "tests/update_other_helper_cli.rs",
        &loads("common/other.rs"),
    );
    repo.put(
        "tests/common/other.rs",
        "pub const COMMAND: &str = \"project-applied\";\n",
    );
    repo.put(
        "tests/update_longer_literal_cli.rs",
        "fn main() { let _ = [\"moai project-apply\"]; }\n",
    );
    repo.put(
        "tests/update_unquoted_cli.rs",
        "// runs project-apply\nfn main() {}\n",
    );
    repo.put(
        "tests/update_mod_cli.rs",
        "fn main() { let _ = [\"mod\"]; }\n",
    );
    repo.commit();

    let mut misses: Vec<String> = Vec::new();
    let range = repo.change(&["src/ops/project_apply.rs"]);
    for (pick, flags) in PICKS {
        let got = repo.listed(&range, flags);
        assert_eq!(got.selection, pick);
        for want in ["update_own_literal_cli", "update_helper_cli"] {
            if !got.targets.contains(want) {
                misses.push(format!("{pick}: missing {want}"));
            }
        }
        for other in [
            "update_applied_cli",
            "update_other_helper_cli",
            "update_longer_literal_cli",
            "update_unquoted_cli",
            "update_mod_cli",
        ] {
            assert!(!got.targets.contains(other), "{pick}: {other}");
        }
    }
    // `mod.rs` is no command file, and a library file of the same name that no command file
    // names reaches no command.
    for path in ["src/ops/mod.rs", "src/project_apply.rs"] {
        let range = repo.change(&[path]);
        for (pick, flags) in PICKS {
            let got = repo.listed(&range, flags);
            assert!(got.targets.is_empty(), "{path} {pick}: {:?}", got.targets);
        }
    }

    // REQ-6: a helper file that a test loads and that is not there stops the selection.
    let repo = Repo::init();
    repo.put("src/ops/project_apply.rs", "// a\n");
    repo.put(
        "tests/update_missing_helper_cli.rs",
        &loads("common/gone_helper.rs"),
    );
    repo.commit();
    let range = repo.change(&["src/ops/project_apply.rs"]);
    for (pick, flags) in PICKS {
        let (ok, answer) = repo.ask(&range, flags);
        if ok {
            misses.push(format!("REQ-6 {pick}: ok=true, no error"));
        } else if !answer["error"].to_string().contains("gone_helper.rs") {
            misses.push(format!(
                "REQ-6 {pick}: the error names no helper file: {}",
                answer["error"]
            ));
        }
    }
    assert!(misses.is_empty(), "{misses:#?}");
}

// SPEC-20261126-affected-binary-tests AC-3
#[test]
fn a_library_change_selects_the_tests_of_its_own_command_files() {
    let repo = Repo::init();
    repo.put("src/trailers.rs", "// a\n");
    // The command files of the library file: its own stem, and the stem without `s` before `_`.
    repo.put("src/ops/trailers.rs", &calls("trailers::format"));
    repo.put("src/ops/trailer_check.rs", &calls("trailers::check"));
    // Command files that call the module and have another name.
    repo.put("src/ops/leader_run.rs", &calls("trailers::format"));
    repo.put("src/ops/trailer.rs", &calls("trailers::format"));
    repo.put("src/ops/trailersmith.rs", &calls("trailers::format"));
    // A fitting name whose file names another module only.
    repo.put("src/ops/trailer_log.rs", &calls("trailers_log::read"));
    for (test, command) in [
        ("update_format_cmd_cli", "trailers"),
        ("update_check_cmd_cli", "trailer-check"),
        ("update_leader_cmd_cli", "leader-run"),
        ("update_singular_cmd_cli", "trailer"),
        ("update_smith_cmd_cli", "trailersmith"),
        ("update_log_cmd_cli", "trailer-log"),
    ] {
        repo.put(
            &format!("tests/{test}.rs"),
            &format!("fn main() {{ let _ = [\"{command}\"]; }}\n"),
        );
    }
    repo.put(
        "tests/update_check_helper_cli.rs",
        &loads("common/checker.rs"),
    );
    repo.put(
        "tests/common/checker.rs",
        "pub const COMMAND: &str = \"trailer-check\";\n",
    );
    repo.commit();

    let mut misses: Vec<String> = Vec::new();
    let range = repo.change(&["src/trailers.rs"]);
    for (pick, flags) in PICKS {
        let got = repo.listed(&range, flags);
        assert_eq!(got.selection, pick);
        for want in [
            "update_format_cmd_cli",
            "update_check_cmd_cli",
            "update_check_helper_cli",
        ] {
            if !got.targets.contains(want) {
                misses.push(format!("{pick}: missing {want}"));
            }
        }
        for other in [
            "update_leader_cmd_cli",
            "update_singular_cmd_cli",
            "update_smith_cmd_cli",
            "update_log_cmd_cli",
        ] {
            assert!(!got.targets.contains(other), "{pick}: {other}");
        }
    }
    assert!(misses.is_empty(), "{misses:#?}");
}

// SPEC-20261126-affected-binary-tests AC-4
#[test]
fn the_missed_binary_tests_of_this_repository_are_selected() {
    let real = Path::new(env!("CARGO_MANIFEST_DIR"));
    let repo = Repo::of_this_repository();

    let mut misses: Vec<String> = Vec::new();
    // REQ-2: each of the five files carries the marker as a whole line.
    for name in MARKED {
        let text = fs::read_to_string(real.join(format!("tests/{name}.rs"))).unwrap();
        if !text.lines().any(|l| l == MARK) {
            misses.push(format!("tests/{name}.rs: no marker line"));
        }
    }
    let changes: [&[&str]; 3] = [
        &["src/launcher_init.rs", "src/ops/project_apply.rs"],
        &["src/state_root.rs"],
        &["src/trailers.rs"],
    ];
    for paths in changes {
        for rel in paths {
            assert!(real.join(rel).is_file(), "{rel} is not in this repository");
        }
        let range = repo.change(paths);
        for (pick, flags) in PICKS {
            let got = repo.listed(&range, flags);
            // The five marked targets hold the three AC-4 names for the first two changes.
            let mut wanted: Vec<&str> = MARKED.to_vec();
            if paths == ["src/trailers.rs"] && pick != "lane" {
                wanted.push("trailer_check_cli");
            }
            for want in wanted {
                if !got.targets.contains(want) {
                    misses.push(format!(
                        "{} {pick} (answered {}): missing {want}",
                        paths.join(" + "),
                        got.selection
                    ));
                }
            }
        }
    }
    assert!(misses.is_empty(), "{misses:#?}");
}

// SPEC-20261126-affected-binary-tests AC-9
#[test]
fn a_tools_change_selects_the_tests_that_name_the_path() {
    let repo = Repo::init();
    // An ordinary file, a nested one and one on the audit floor.
    let tools = [
        ("tools/helper.sh", "update_helper_path_cli"),
        ("tools/nested/table.json", "update_table_path_cli"),
        ("tools/self_check.sh", "update_floor_path_cli"),
    ];
    for (path, test) in tools {
        repo.put(path, "a\n");
        repo.put(
            &format!("tests/{test}.rs"),
            &format!("fn main() {{ let _ = [\"{path}\"]; }}\n"),
        );
        // The same file by its name alone.
        let name = path.rsplit('/').next().unwrap();
        repo.put(
            &format!("tests/{test}_name_only.rs"),
            &format!("fn main() {{ let _ = [\"{name}\"]; }}\n"),
        );
    }
    repo.commit();

    let mut misses: Vec<String> = Vec::new();
    for (path, test) in tools {
        let range = repo.change(&[path]);
        for (pick, flags) in PICKS {
            let got = repo.listed(&range, flags);
            // A card request for a floor path answers the full set.
            let answered = if pick == "card" && path == "tools/self_check.sh" {
                "full"
            } else {
                pick
            };
            assert_eq!(got.selection, answered, "{path} {pick}");
            if !got.targets.contains(test) {
                misses.push(format!("{path} {pick}: missing {test}"));
            }
            for (_, holder) in tools {
                let name_only = format!("{holder}_name_only");
                assert!(
                    !got.targets.contains(&name_only),
                    "{path} {pick}: {name_only}"
                );
            }
        }
    }

    // The repository's present files. The change adds a blank line, so the scan list still reads:
    // an unreadable one selects every test and would prove nothing.
    let real = Path::new(env!("CARGO_MANIFEST_DIR"));
    let registry = "tools/repo_scans.json";
    assert!(real.join("tests/gate_registry_tier.rs").is_file());
    let repo = Repo::of_this_repository();
    let base = git(repo.dir.path(), &["rev-parse", "HEAD"]);
    let text = fs::read_to_string(repo.dir.path().join(registry)).unwrap();
    repo.put(registry, &format!("{text}\n"));
    let range = format!("{base}..{}", repo.commit());
    for (pick, flags) in PICKS {
        let got = repo.listed(&range, flags);
        assert!(got.scans.starts_with("listed:"), "{pick}: {}", got.scans);
        if !got.targets.contains("gate_registry_tier") {
            misses.push(format!(
                "{registry} {pick} (answered {}): missing gate_registry_tier",
                got.selection
            ));
        }
    }
    assert!(misses.is_empty(), "{misses:#?}");
}

/// The command file stems that `src/help.rs` lists no command for.
fn unlisted(stems: &BTreeSet<String>) -> Vec<&str> {
    let listed: BTreeSet<&str> = moai_cli::help::GROUPS
        .iter()
        .flat_map(|g| g.commands.iter().map(|(word, _)| *word))
        .collect();
    stems
        .iter()
        .filter(|stem| !listed.contains(stem.replace('_', "-").as_str()))
        .map(String::as_str)
        .collect()
}

// SPEC-20261126-affected-binary-tests AC-8
#[test]
fn every_command_file_is_named_like_a_listed_command() {
    let mut files = Vec::new();
    rust_files(Path::new(env!("CARGO_MANIFEST_DIR")), "src/ops", &mut files);
    let stems: BTreeSet<String> = files
        .iter()
        .filter_map(|f| f.rsplit('/').next()?.strip_suffix(".rs"))
        .filter(|stem| *stem != "mod")
        .map(String::from)
        .collect();
    assert!(!stems.is_empty(), "no command file was read");
    let wrong = unlisted(&stems);
    assert!(wrong.is_empty(), "no listed command for: {wrong:?}");
    // The check does name a file that no command is listed for.
    let look_alike = BTreeSet::from(["project_applied".to_string()]);
    assert_eq!(unlisted(&look_alike), ["project_applied"]);
}
