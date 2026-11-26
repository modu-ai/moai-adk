//! SPEC-20261126-affected-binary-tests: `check affected` also picks the integration tests that
//! drive the built binary. A marker line joins every source change, a changed command file reaches
//! the tests that name its command, a changed library file reaches the tests of its own command
//! files, a changed file under `tools/` reaches the tests that hold its path, and a test file may
//! declare a path whose change selects it. The selection is read from the built binary (`--list`)
//! in temporary git repositories with a temporary HOME; the real repository is only read.
mod common;
use common::git::{git, git_ok};
use common::home::TempHome;
use serde_json::Value;
use std::collections::BTreeSet;
use std::fs;
use std::path::{Component, Path};
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

/// The start of the whole line by which a test file names a path it depends on (REQ-10); the path
/// follows it. It is only ever joined to a path inside a string, so no line of this file declares
/// one.
const DECLARATION: &str = "// check-affected: when-changed ";

/// The test files of this repository that must declare [`DECLARED_PATH`] (REQ-11).
const DECLARING: [&str; 4] = [
    "tests/card/card_submit_cli.rs",
    "tests/card/card_wip_limit.rs",
    "tests/card/orchestrate_cli.rs",
    "tests/leader_brief_cli.rs",
];

/// The source file those tests drive from the commands above it.
const DECLARED_PATH: &str = "src/card/work.rs";

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
        let path = std::env::var_os("PATH").unwrap_or_default();
        self.ask_on(range, flags, &path.to_string_lossy())
    }

    /// [`Repo::ask`] under the `PATH` given.
    fn ask_on(&self, range: &str, flags: &[&str], path: &str) -> (bool, Value) {
        let dir = self.dir.path().to_string_lossy().into_owned();
        let mut args = vec!["check", "affected", range, "--list", "--dir", dir.as_str()];
        args.extend_from_slice(flags);
        let out = self
            .home
            .command(&args)
            .env("PATH", path)
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

    /// [`Repo::ask`] with nextest as the planned runner, for a selection that must be accepted.
    fn planned(&self, range: &str, flags: &[&str]) -> Value {
        let (_tools, path) = nextest_path();
        let (ok, v) = self.ask_on(range, flags, &path);
        assert!(ok, "check affected refused the fixture: {v}");
        v
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

/// A folder of a fake `cargo` that exits 0, so the planned runner is nextest and the `test` step
/// carries a filter, in front of the system's `git`.
fn nextest_path() -> (tempfile::TempDir, String) {
    let dir = tempfile::tempdir().unwrap();
    let cargo = dir.path().join("cargo");
    fs::write(&cargo, "#!/bin/sh\nexit 0\n").unwrap();
    #[cfg(unix)]
    {
        use std::os::unix::fs::PermissionsExt;
        fs::set_permissions(&cargo, fs::Permissions::from_mode(0o755)).unwrap();
    }
    let real = std::env::var_os("PATH").unwrap_or_default();
    let path = format!("{}:{}", dir.path().display(), real.to_string_lossy());
    (dir, path)
}

/// The `-E` filter of the planned `test` step; `None` when the step runs everything it builds.
fn test_filter(answer: &Value) -> Option<String> {
    let steps = answer["result"]["steps"].as_array().unwrap();
    let step = steps
        .iter()
        .find(|s| s["label"] == "test")
        .expect("a test step");
    let args = step["command"].as_array().unwrap();
    assert!(args.iter().any(|a| a == "nextest"), "{args:?}");
    let at = args.iter().position(|a| a == "-E")?;
    args.get(at + 1).and_then(Value::as_str).map(String::from)
}

/// Whether the filter leaves the library tests whole.
fn whole_library(filter: Option<&str>) -> bool {
    filter.is_none_or(|f| f.split(" | ").any(|term| term == "kind(lib)"))
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

    // REQ-6: so does a command file that the library rule reads and that is no text.
    let repo = Repo::init();
    repo.put("src/trailers.rs", "// a\n");
    repo.put("src/ops/trailer_check.rs", "// a\n");
    let unreadable = repo.dir.path().join("src/ops/trailer_check.rs");
    fs::write(unreadable, [0xff, 0xfe, b'\n']).unwrap();
    repo.commit();
    let range = repo.change(&["src/trailers.rs"]);
    for (pick, flags) in PICKS {
        let (ok, answer) = repo.ask(&range, flags);
        if ok {
            misses.push(format!(
                "REQ-6 {pick}: ok=true for an unreadable command file"
            ));
        } else if !answer["error"].to_string().contains("trailer_check.rs") {
            misses.push(format!(
                "REQ-6 {pick}: the error names no command file: {}",
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

// SPEC-20261126-affected-binary-tests AC-10
#[test]
fn a_source_change_in_a_scan_scope_selects_its_library_tests() {
    let repo = Repo::init();
    // Four scans with a library check, none of whose rule files the changes below touch: a named
    // test beside an integration target, two names after `--`, one scoped to the documents, and
    // one that names no test.
    let scans = r#"{"scans":[
      {"id":"table","checks":["cargo test --offline --locked --test update_table_cli",
        "cargo test --offline --locked --lib every_name_is_in_the_table"],
        "rules":["src/table.rs"],"scope":[{"prefix":"src/","suffixes":[".rs"]}]},
      {"id":"walk","checks":["cargo test --offline --locked --lib -- first_wide_read second_wide_read"],
        "rules":["src/walker.rs"],"scope":[{"prefix":"src/","suffixes":[".rs"]}]},
      {"id":"documents","checks":["cargo test --offline --locked --lib documents_only_read"],
        "rules":["src/pages.rs"],"scope":[{"prefix":"docs/","suffixes":[".md"]}]},
      {"id":"all","checks":["cargo test --offline --locked --lib"],
        "rules":["src/every.rs"],"scope":[{"prefix":"src/deep/","suffixes":[".rs"]}]}
    ]}"#;
    repo.put("tools/repo_scans.json", scans);
    repo.put("tests/update_table_cli.rs", "fn main() {}\n");
    repo.put("src/widget.rs", "// a\n");
    repo.put("src/deep/part.rs", "// a\n");
    repo.commit();

    let mut misses: Vec<String> = Vec::new();
    let range = repo.change(&["src/widget.rs"]);
    let card = repo.planned(&range, &["--card"]);
    assert_eq!(card["result"]["selection"], "card", "{card}");
    let filter = test_filter(&card).expect("the card set narrows the library tests");
    assert!(!whole_library(Some(&filter)), "{filter}");
    for name in [
        "every_name_is_in_the_table",
        "first_wide_read",
        "second_wide_read",
    ] {
        if !filter.contains(name) {
            misses.push(format!("card src/widget.rs: missing {name}"));
        }
    }
    assert!(!filter.contains("documents_only_read"), "{filter}");
    let targets = card["result"]["rust_targets"].to_string();
    assert!(!targets.contains("update_table_cli"), "{targets}");
    // The full and the lane set keep the whole library.
    for (pick, flags) in [PICKS[0], PICKS[2]] {
        let got = repo.planned(&range, flags);
        let filter = test_filter(&got);
        assert!(whole_library(filter.as_deref()), "{pick}: {filter:?}");
    }
    // A library check that names no test asks for the whole library.
    let range = repo.change(&["src/deep/part.rs"]);
    let filter = test_filter(&repo.planned(&range, &["--card"]));
    if !whole_library(filter.as_deref()) {
        misses.push(format!(
            "card src/deep/part.rs: the library is narrowed: {filter:?}"
        ));
    }

    // The repository's present files: the home name scan reads every source file.
    let repo = Repo::of_this_repository();
    let range = repo.change(&["src/update.rs"]);
    let card = repo.planned(&range, &["--card"]);
    assert_eq!(card["result"]["selection"], "card", "{card}");
    let filter = test_filter(&card);
    let name = "every_rust_written_home_name_is_in_the_table";
    if !filter.as_deref().is_none_or(|f| f.contains(name)) && !whole_library(filter.as_deref()) {
        misses.push(format!("card src/update.rs: missing {name}"));
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

// SPEC-20261126-affected-binary-tests AC-11
#[test]
fn a_declared_path_change_selects_the_test_that_names_it() {
    let repo = Repo::init();
    let shared = format!("{DECLARATION}data/shared.txt");
    // A target of its own with two declarations.
    repo.put(
        "tests/update_declared_cli.rs",
        &format!("//! A test.\n{shared}\n{DECLARATION}src/part.rs\nfn main() {{}}\n"),
    );
    // A module of an area test binary beside a module that declares nothing; the binary's
    // `main.rs` and a helper file are no test files.
    repo.put(
        "tests/card/main.rs",
        &format!("{DECLARATION}data/main.txt\nmod card_declared_cli;\nmod card_quiet_cli;\n"),
    );
    repo.put(
        "tests/card/card_declared_cli.rs",
        &format!("{DECLARATION}data/module.txt\n"),
    );
    repo.put("tests/card/card_quiet_cli.rs", "// a\n");
    repo.put(
        "tests/common/declares.rs",
        &format!("{DECLARATION}data/helper.txt\n"),
    );
    repo.put(
        "tests/update_declare_helper_cli.rs",
        &loads("common/declares.rs"),
    );
    // Lines that are not the declaration of a changed path.
    let look_alikes = [
        (
            "update_declare_trailing_cli",
            format!("{shared} (not the line)"),
        ),
        ("update_declare_longer_cli", format!("{shared}.bak")),
        (
            "update_declare_prefix_cli",
            format!("{DECLARATION}data\n{DECLARATION}src/part"),
        ),
        ("update_declare_doc_cli", format!("//! {shared}")),
        (
            "update_declare_inline_cli",
            format!("fn quiet() {{}} {shared}"),
        ),
        (
            "update_declare_literal_cli",
            format!("const LINE: &str = \"{shared}\";"),
        ),
    ];
    for (name, lines) in &look_alikes {
        repo.put(
            &format!("tests/{name}.rs"),
            &format!("{lines}\nfn main() {{}}\n"),
        );
    }
    let changes: [(&str, &[&str]); 6] = [
        ("data/shared.txt", &["update_declared_cli"]),
        ("src/part.rs", &["update_declared_cli"]),
        ("data/module.txt", &["card_declared_cli"]),
        ("data/unnamed.txt", &[]),
        ("data/main.txt", &[]),
        ("data/helper.txt", &[]),
    ];
    for (path, _) in changes {
        repo.put(path, "// a\n");
    }
    repo.commit();

    let mut misses: Vec<String> = Vec::new();
    for (path, wanted) in changes {
        let range = repo.change(&[path]);
        for (pick, flags) in PICKS {
            let got = repo.listed(&range, flags);
            assert_eq!(got.selection, pick, "{path}");
            for target in &got.targets {
                assert!(wanted.contains(&target.as_str()), "{path} {pick}: {target}");
            }
            for want in wanted {
                if !got.targets.contains(*want) {
                    misses.push(format!("{path} {pick}: missing {want}"));
                }
            }
            // A module of an area binary is selected as that module alone.
            if got.targets.contains("card_declared_cli") {
                let filter = test_filter(&repo.planned(&range, flags)).unwrap_or_default();
                let narrowed = "(binary(=card) & (test(/^card_declared_cli::/)))";
                assert!(filter.contains(narrowed), "{path} {pick}: {filter}");
                assert!(!filter.contains("card_quiet_cli"), "{pick}: {filter}");
                assert!(
                    !filter.split(" | ").any(|term| term == "binary(=card)"),
                    "{pick}: {filter}"
                );
            }
        }
    }

    // REQ-11 and the repository's present files.
    let real = Path::new(env!("CARGO_MANIFEST_DIR"));
    let line = format!("{DECLARATION}{DECLARED_PATH}");
    for rel in DECLARING {
        let text = fs::read_to_string(real.join(rel)).unwrap();
        if !text.lines().any(|l| l == line) {
            misses.push(format!("{rel}: no declaration line"));
        }
    }
    assert!(real.join(DECLARED_PATH).is_file());
    let repo = Repo::of_this_repository();
    let range = repo.change(&[DECLARED_PATH]);
    for (pick, flags) in PICKS {
        let got = repo.listed(&range, flags);
        for rel in DECLARING {
            let name = rel.rsplit('/').next().unwrap().trim_end_matches(".rs");
            if !got.targets.contains(name) {
                misses.push(format!(
                    "{DECLARED_PATH} {pick} (answered {}): missing {name}",
                    got.selection
                ));
            }
        }
    }
    assert!(misses.is_empty(), "{misses:#?}");
}

/// How many declarations the Rust files under `tests/` of `root` hold, and those that name no
/// file inside `root`, as `<file>: <path>`.
fn declarations(root: &Path) -> (usize, Vec<String>) {
    let mut files = Vec::new();
    rust_files(root, "tests", &mut files);
    let mut read = 0;
    let mut wrong = Vec::new();
    for file in &files {
        let text = fs::read_to_string(root.join(file)).unwrap();
        for path in text.lines().filter_map(|l| l.strip_prefix(DECLARATION)) {
            read += 1;
            let inside = Path::new(path)
                .components()
                .all(|c| matches!(c, Component::Normal(_)));
            if !inside || !root.join(path).is_file() {
                wrong.push(format!("{file}: {path}"));
            }
        }
    }
    wrong.sort();
    (read, wrong)
}

// SPEC-20261126-affected-binary-tests AC-12
#[test]
fn every_declared_path_names_a_file() {
    // The check does name a declaration whose path is no file of the repository.
    let dir = tempfile::tempdir().unwrap();
    let root = dir.path().join("repo");
    for folder in ["src", "tests"] {
        fs::create_dir_all(root.join(folder)).unwrap();
    }
    fs::write(root.join("src/there.rs"), "// a\n").unwrap();
    fs::write(dir.path().join("outside.rs"), "// a\n").unwrap();
    let paths = [
        "src/there.rs",
        "src/gone.rs",
        "src",
        "src/there.rs (note)",
        "../outside.rs",
    ];
    let lines: String = paths
        .iter()
        .map(|p| format!("{DECLARATION}{p}\n"))
        .collect();
    fs::write(root.join("tests/some_cli.rs"), lines).unwrap();
    let named = ["../outside.rs", "src", "src/gone.rs", "src/there.rs (note)"]
        .map(|p| format!("tests/some_cli.rs: {p}"));
    assert_eq!(declarations(&root), (paths.len(), named.to_vec()));

    let (read, wrong) = declarations(Path::new(env!("CARGO_MANIFEST_DIR")));
    assert!(wrong.is_empty(), "no such file: {wrong:#?}");
    assert!(
        read >= DECLARING.len(),
        "this repository holds {read} declarations"
    );
}
