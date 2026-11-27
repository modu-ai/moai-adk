//! SPEC-20261126-tracked-state-move: a move of the earlier state that would put tracked files
//! where git ignores them names them and waits for the user, and the flaky ledger is written only
//! through real folders and a real file. Every case builds a temporary project with a temporary
//! HOME beside it; the real checkout and the real HOME are never read.

#![cfg(unix)]

mod common;

use common::git::git_ok;
use common::CleanEnv;
use moai_cli::check_flaky::{record, TestName};
use moai_cli::state_root::{
    classify, EARLIER_PROJECT_DIR as OLD, GO_STATE_RECEIPT, LOCAL_DIR, MOAI_DIR as NEW,
    RUST_MARKER, SPECS_DIR,
};
use serde_json::{json, Value};
use std::collections::BTreeMap;
use std::fs;
use std::os::unix::fs::{symlink, PermissionsExt};
use std::path::{Path, PathBuf};
use std::process::Command;
use std::sync::mpsc;
use std::time::Duration;
use tempfile::TempDir;

const BIN: &str = crate::common::MOAI_BIN;
const PERMISSION: &str = "test-permission-ref";
const LEDGER: &str = "check-flaky.jsonl";
const TOOL_PATH: &str = "/usr/bin:/bin";
const STAMP: &str = "20261126T000000Z";
/// REQ-1 and D-5: the preview names at most this many files.
const NAMED: usize = 20;
const OLD_SPEC: &str = "SPEC-20260101-old";

/// A temporary project folder, a temporary HOME beside it and the `PATH` its commands see.
struct Project {
    _tmp: TempDir,
    base: PathBuf,
    root: PathBuf,
    path: String,
}

impl Project {
    /// A project in no git work tree.
    fn plain() -> Project {
        let tmp = TempDir::new().unwrap();
        let base = fs::canonicalize(tmp.path()).unwrap();
        let root = base.join("project");
        fs::create_dir_all(&root).unwrap();
        fs::create_dir_all(base.join("home")).unwrap();
        Project {
            _tmp: tmp,
            base,
            root,
            path: TOOL_PATH.to_string(),
        }
    }

    /// A project that is a git repository.
    fn in_git() -> Project {
        let p = Project::plain();
        git_ok(&p.root, &["init", "-q"]);
        p
    }

    /// `<root>/<rel>` holding `body`.
    fn write(&self, rel: &str, body: &str) {
        let path = self.root.join(rel);
        fs::create_dir_all(path.parent().unwrap()).unwrap();
        fs::write(path, body).unwrap();
    }

    /// The earlier state: one file at the top and one in a folder.
    fn old_state(&self) {
        self.write(&format!("{OLD}/project.yaml"), "schema_version: 2\n");
        self.write(&format!("{OLD}/index/a.json"), "{}\n");
    }

    /// The project's `.gitignore`, one rule per line.
    fn rules(&self, lines: &[String]) {
        self.write(".gitignore", &format!("{}\n", lines.join("\n")));
    }

    /// Everything the rules allow goes into the git index.
    fn track(&self) {
        git_ok(&self.root, &["add", "-A", "--", "."]);
    }

    /// A `git` first on `PATH` that runs `body` for `subcommand` and the real git otherwise.
    fn fake_git(&mut self, subcommand: &str, body: &str) {
        let bin = self.base.join("fake-bin");
        fs::create_dir_all(&bin).unwrap();
        let script = format!(
            "#!/bin/sh\nfor a in \"$@\"; do\n  if [ \"$a\" = {subcommand} ]; then\n{body}\n  fi\ndone\nexec /usr/bin/git \"$@\"\n"
        );
        fs::write(bin.join("git"), script).unwrap();
        fs::set_permissions(bin.join("git"), fs::Permissions::from_mode(0o755)).unwrap();
        self.path = format!("{}:{TOOL_PATH}", bin.display());
    }

    /// A `PATH` that holds no program at all.
    fn without_git(&mut self) {
        let bin = self.base.join("empty-bin");
        fs::create_dir_all(&bin).unwrap();
        self.path = bin.display().to_string();
    }

    fn call(&self, args: &[&str]) -> Value {
        let out = Command::new(BIN)
            .args(args)
            .clean_env()
            .env("HOME", self.base.join("home"))
            .env("PATH", &self.path)
            .env("LC_ALL", "C")
            .output()
            .unwrap();
        serde_json::from_slice(&out.stdout).unwrap_or_else(|_| {
            panic!(
                "{args:?}: {}{}",
                String::from_utf8_lossy(&out.stdout),
                String::from_utf8_lossy(&out.stderr)
            )
        })
    }

    fn arg(&self) -> &str {
        self.root.to_str().unwrap()
    }

    /// The result of the state move preview.
    fn plan(&self) -> Value {
        let v = self.call(&["go-state-plan", self.arg()]);
        assert_eq!(v["ok"], json!(true), "{v}");
        v["result"].clone()
    }

    /// The `go-state` step of the update preview.
    fn step(&self) -> Value {
        let v = self.call(&["project-update-plan", self.arg()]);
        v["result"]["phases"]
            .as_array()
            .unwrap_or_else(|| panic!("{v}"))
            .iter()
            .flat_map(|p| p["steps"].as_array().into_iter().flatten())
            .find(|s| s["step"] == "go-state")
            .cloned()
            .unwrap_or_else(|| panic!("no go-state step: {v}"))
    }

    fn apply(&self, digest: &str) -> Value {
        self.call(&[
            "go-state-apply",
            self.arg(),
            "--plan-digest",
            digest,
            "--permission-reference",
            PERMISSION,
        ])
    }

    fn generation(&self) -> &'static str {
        classify(&self.root).generation
    }

    fn record(&self) -> Result<(), String> {
        record_in(&self.root, &self.base.join("check"))
    }

    fn ledger(&self) -> PathBuf {
        self.root.join(NEW).join(LOCAL_DIR).join(LEDGER)
    }

    /// The preview names exactly these state files (relative to the state folder) of `total`
    /// hidden ones, says whether the marker is hidden, and the update step offers no command.
    fn expect_stopped(&self, rels: &[&str], total: usize, marker: bool) {
        let plan = self.plan();
        let found = &plan["ignored_tracked"];
        assert_eq!(found["state"], json!("ignored"), "{plan}");
        let want: Vec<(String, String)> = rels
            .iter()
            .map(|r| (format!("{OLD}/{r}"), format!("{NEW}/{r}")))
            .collect();
        assert_eq!(named(found), want, "{found}");
        assert_eq!(found["total"], json!(total), "{found}");
        assert_eq!(found["marker"].is_string(), marker, "{found}");

        let step = self.step();
        assert_eq!(step["state"], json!("needs_decision"), "{step}");
        assert_eq!(step["apply"], Value::Null, "{step}");
        let reason = step["reason"].as_str().unwrap();
        for file in found["files"].as_array().unwrap() {
            let rule = file["rule"].as_str().unwrap();
            assert!(rule.contains(".gitignore:"), "{file}");
            assert!(reason.contains(file["file"].as_str().unwrap()), "{reason}");
            assert!(reason.contains(rule), "{reason}");
        }
        if let Some(rule) = found["marker"].as_str() {
            assert!(
                reason.contains(&format!("{NEW}/{RUST_MARKER}")) && reason.contains(rule),
                "{reason}"
            );
        }
    }

    /// The move is offered with a command, applies, and the state arrives.
    fn expect_offered_and_applies(&self, label: &str) {
        let plan = self.plan();
        assert!(plan.get("ignored_tracked").is_none(), "{label}: {plan}");
        let step = self.step();
        assert_eq!(step["state"], json!("needs_decision"), "{label}: {step}");
        assert!(step["apply"]["argv"].is_array(), "{label}: {step}");
        let applied = self.apply(plan["plan_digest"].as_str().unwrap());
        assert_eq!(applied["ok"], json!(true), "{label}: {applied}");
        assert_eq!(
            applied["result"]["state"],
            json!("committed"),
            "{label}: {applied}"
        );
        assert!(
            self.root.join(NEW).join("index/a.json").is_file(),
            "{label}"
        );
        assert!(!self.root.join(OLD).join("index").exists(), "{label}");
    }
}

fn record_in(root: &Path, out: &Path) -> Result<(), String> {
    fs::create_dir_all(out).unwrap();
    let name = TestName {
        binary: "moai-cli::demo_cli".to_string(),
        name: "t_flaky".to_string(),
    };
    record(root, out, "0123456789ab", &[name])
}

/// The `(file, destination)` pairs of the hidden files.
fn named(found: &Value) -> Vec<(String, String)> {
    found["files"]
        .as_array()
        .into_iter()
        .flatten()
        .map(|f| {
            (
                f["file"].as_str().unwrap().to_string(),
                f["to"].as_str().unwrap().to_string(),
            )
        })
        .collect()
}

/// A rule that hides everything inside the state folder.
fn hides_all() -> String {
    format!("/{NEW}/*")
}

/// Every entry below `dir`, folders and links included, with what it holds.
fn tree(dir: &Path) -> BTreeMap<PathBuf, (char, Vec<u8>)> {
    let mut out = BTreeMap::new();
    let mut stack = vec![dir.to_path_buf()];
    while let Some(d) = stack.pop() {
        let Ok(entries) = fs::read_dir(&d) else {
            continue;
        };
        for e in entries {
            let p = e.unwrap().path();
            let m = fs::symlink_metadata(&p).unwrap();
            let held = if m.file_type().is_symlink() {
                let to = fs::read_link(&p).unwrap();
                ('l', to.into_os_string().into_encoded_bytes())
            } else if m.is_dir() {
                stack.push(p.clone());
                ('d', Vec::new())
            } else {
                ('f', fs::read(&p).unwrap())
            };
            out.insert(p, held);
        }
    }
    out
}

/// SPEC-20261126-tracked-state-move AC-1 (REQ-1, REQ-2)
#[test]
fn an_absent_state_folder_names_the_hidden_tracked_files_and_offers_no_command() {
    let p = Project::in_git();
    p.old_state();
    p.rules(&[hides_all()]);
    p.track();
    assert!(!p.root.join(NEW).exists(), "the premise");
    p.expect_stopped(&["index/a.json", "project.yaml"], 2, true);

    // At the limit every file is named; one more is counted and not named.
    for count in [NAMED, NAMED + 1] {
        let p = Project::in_git();
        let rels: Vec<String> = (0..count).map(|n| format!("index/{n:02}.json")).collect();
        for rel in &rels {
            p.write(&format!("{OLD}/{rel}"), "{}\n");
        }
        p.rules(&[format!("/{NEW}/index/")]);
        p.track();
        let listed: Vec<&str> = rels.iter().take(NAMED).map(String::as_str).collect();
        p.expect_stopped(&listed, count, false);
        let step = p.step();
        let reason = step["reason"].as_str().unwrap();
        assert!(
            reason.contains(&format!("{count} tracked file")),
            "{reason}"
        );
        if count > NAMED {
            assert!(!reason.contains(&rels[NAMED]), "{reason}");
        }
    }
}

/// SPEC-20261126-tracked-state-move AC-2 (REQ-1, REQ-2)
#[test]
fn a_local_only_state_folder_names_the_hidden_tracked_files_and_offers_no_command() {
    for finder_file in [false, true] {
        let p = Project::in_git();
        p.old_state();
        p.write(&format!("{NEW}/{LOCAL_DIR}/output.log"), "x\n");
        if finder_file {
            p.write(&format!("{NEW}/.DS_Store"), "x");
        }
        p.rules(&[hides_all(), ".DS_Store".to_string()]);
        p.track();
        assert_eq!(p.generation(), "none", "the premise");
        p.expect_stopped(&["index/a.json", "project.yaml"], 2, true);
    }
}

/// SPEC-20261126-tracked-state-move AC-3 (REQ-7)
#[test]
fn a_flaky_record_and_a_spec_write_leave_the_move_answer_as_it_was() {
    let allow = |rest: &str| format!("!/{NEW}/{rest}");
    let hidden = vec![
        hides_all(),
        allow(&format!("{SPECS_DIR}/")),
        allow(RUST_MARKER),
        format!("/{NEW}/{SPECS_DIR}/{OLD_SPEC}/"),
    ];
    let allowed = vec![
        hides_all(),
        allow("index/"),
        allow("project.yaml"),
        allow(&format!("{SPECS_DIR}/")),
        allow(RUST_MARKER),
    ];
    let hidden_files: Vec<(String, String)> = [
        "index/a.json".to_string(),
        "project.yaml".to_string(),
        format!("{SPECS_DIR}/{OLD_SPEC}/spec.md"),
    ]
    .iter()
    .map(|r| (format!("{OLD}/{r}"), format!("{NEW}/{r}")))
    .collect();
    let answer = |p: &Project| {
        let plan = p.plan();
        let step = p.step();
        (
            p.generation(),
            step["state"].clone(),
            !step["apply"].is_null(),
            named(&plan["ignored_tracked"]),
            plan["ignored_tracked"]["total"].clone(),
            plan["ignored_tracked"]["marker"].clone(),
        )
    };
    let mut problems = Vec::new();
    for (layout, rules, files) in [
        ("hidden", &hidden, hidden_files),
        ("allowed", &allowed, Vec::new()),
    ] {
        for writes in [
            &["record"][..],
            &["spec"],
            &["record", "spec"],
            &["spec", "record"],
        ] {
            let p = Project::in_git();
            p.old_state();
            p.write(&format!("{OLD}/{SPECS_DIR}/{OLD_SPEC}/spec.md"), "# old\n");
            p.rules(rules);
            p.track();
            let before = answer(&p);
            // The premise: the hidden layout names its files and offers no command.
            assert_eq!(before.3, files, "{layout}");
            assert_eq!(before.2, files.is_empty(), "{layout}");
            assert_eq!(before.1, json!("needs_decision"), "{layout}");
            for write in writes {
                if *write == "record" {
                    p.record().unwrap();
                    assert!(p.ledger().is_file(), "the premise");
                } else {
                    p.write(
                        &format!("{NEW}/{SPECS_DIR}/SPEC-20260102-new/spec.md"),
                        "# new\n",
                    );
                }
                let after = answer(&p);
                if after != before {
                    problems.push(format!(
                        "{layout} {writes:?} after {write}: before {before:?}, after {after:?}"
                    ));
                }
            }
        }
    }
    assert!(problems.is_empty(), "{}", problems.join("\n"));
}

/// SPEC-20261126-tracked-state-move AC-4 (REQ-1, REQ-2)
#[test]
fn a_hidden_marker_or_nested_file_stops_the_move() {
    // Only the marker is hidden, while a moving source holds tracked files.
    let p = Project::in_git();
    p.old_state();
    p.rules(&[format!("/{NEW}/{RUST_MARKER}")]);
    p.track();
    p.expect_stopped(&[], 0, true);

    // Only a file of a folder inside a moving entry is hidden.
    let p = Project::in_git();
    p.old_state();
    p.write(&format!("{OLD}/index/deep/b.json"), "{}\n");
    p.rules(&[format!("/{NEW}/index/deep/")]);
    p.track();
    p.expect_stopped(&["index/deep/b.json"], 1, false);

    // More destinations than one git call takes: the last one is still judged.
    let p = Project::in_git();
    for n in 0..70 {
        p.write(&format!("{OLD}/index/{n:02}.json"), "{}\n");
    }
    p.rules(&[format!("/{NEW}/index/69.json")]);
    p.track();
    p.expect_stopped(&["index/69.json"], 1, false);

    // A project in a folder of a larger repository is in a git work tree too.
    let p = Project::plain();
    git_ok(&p.base, &["init", "-q"]);
    p.old_state();
    p.rules(&[hides_all()]);
    p.track();
    p.expect_stopped(&["index/a.json", "project.yaml"], 2, true);
}

/// SPEC-20261126-tracked-state-move AC-5 (REQ-4, REQ-5)
#[test]
fn an_apply_into_an_ignored_place_is_refused_and_writes_nothing() {
    for (local, rule, named_in_refusal) in [
        (false, hides_all(), format!("{OLD}/index/a.json")),
        (true, hides_all(), format!("{OLD}/index/a.json")),
        (
            false,
            format!("/{NEW}/{RUST_MARKER}"),
            format!("{NEW}/{RUST_MARKER}"),
        ),
    ] {
        let p = Project::in_git();
        p.old_state();
        if local {
            p.write(&format!("{NEW}/{LOCAL_DIR}/output.log"), "x\n");
            p.rules(&[format!("/{NEW}/{LOCAL_DIR}/")]);
        }
        p.track();
        // The preview that is confirmed allows every destination.
        let plan = p.plan();
        assert!(plan.get("ignored_tracked").is_none(), "{plan}");
        let digest = plan["plan_digest"].as_str().unwrap().to_string();
        // The rule arrives after the preview, so the apply has to judge again.
        p.rules(&[rule]);
        let before = tree(&p.root);
        let applied = p.apply(&digest);
        assert_eq!(applied["ok"], json!(false), "local {local}: {applied}");
        let message = applied["error"]["message"].as_str().unwrap();
        assert!(
            message.contains("refused") && message.contains(&named_in_refusal),
            "local {local}: {message}"
        );
        assert_eq!(tree(&p.root), before, "local {local}: nothing changes");
    }
}

/// SPEC-20261126-tracked-state-move AC-6 (REQ-3, REQ-5)
#[test]
fn a_check_git_cannot_answer_offers_no_command() {
    /// Prints `junk` for every path after `--`.
    const JUNK: &str = "    seen=0\n    for b in \"$@\"; do\n      if [ $seen = 1 ]; then echo junk; fi\n      if [ \"$b\" = -- ]; then seen=1; fi\n    done\n    exit 0";
    type Break = fn(&mut Project);
    let cases: [(&str, Break); 7] = [
        ("the index cannot be read", |p| {
            fs::write(p.root.join(".git/index"), "not an index").unwrap()
        }),
        ("check-ignore fails", |p| {
            p.fake_git("check-ignore", "    echo 'fatal: boom' >&2\n    exit 128")
        }),
        ("check-ignore answers for no path", |p| {
            p.fake_git("check-ignore", "    exit 1")
        }),
        ("check-ignore answers in another shape", |p| {
            p.fake_git("check-ignore", JUNK)
        }),
        ("rev-parse fails", |p| {
            p.fake_git(
                "rev-parse",
                "    echo 'fatal: detected dubious ownership in repository' >&2\n    exit 128",
            )
        }),
        ("rev-parse answers in another shape", |p| {
            p.fake_git("rev-parse", "    echo maybe\n    exit 0")
        }),
        ("git cannot be run", Project::without_git),
    ];
    for local in [false, true] {
        for (case, break_it) in cases {
            let mut p = Project::in_git();
            p.old_state();
            if local {
                p.write(&format!("{NEW}/{LOCAL_DIR}/output.log"), "x\n");
            }
            p.track();
            break_it(&mut p);
            let plan = p.plan();
            assert_eq!(
                plan["ignored_tracked"]["state"],
                json!("unknown"),
                "{case}, local {local}: {plan}"
            );
            let step = p.step();
            assert_eq!(step["state"], json!("needs_decision"), "{case}: {step}");
            assert_eq!(step["apply"], Value::Null, "{case}: {step}");
            assert_eq!(
                step["reason"], plan["ignored_tracked"]["reason"],
                "{case}: {step}"
            );
            let before = tree(&p.root);
            let applied = p.apply(plan["plan_digest"].as_str().unwrap());
            assert_eq!(applied["ok"], json!(false), "{case}: {applied}");
            assert!(
                applied["error"]["message"]
                    .as_str()
                    .unwrap()
                    .contains("refused"),
                "{case}: {applied}"
            );
            assert_eq!(tree(&p.root), before, "{case}: nothing changes");
        }
    }
}

/// SPEC-20261126-tracked-state-move AC-7 (REQ-6)
#[test]
fn an_allowed_or_untracked_move_is_offered_and_applies() {
    let allowed = || {
        vec![
            hides_all(),
            format!("!/{NEW}/index/"),
            format!("!/{NEW}/project.yaml"),
            format!("!/{NEW}/{RUST_MARKER}"),
        ]
    };
    for local in [false, true] {
        let p = Project::in_git();
        p.old_state();
        if local {
            p.write(&format!("{NEW}/{LOCAL_DIR}/output.log"), "x\n");
        }
        p.rules(&allowed());
        p.track();
        p.expect_offered_and_applies(&format!("allowed destinations, local {local}"));
    }

    let p = Project::in_git();
    p.old_state();
    p.rules(&[hides_all()]);
    p.expect_offered_and_applies("nothing is tracked");

    // The only tracked file sits in an entry that stays, so no moving source holds one and the
    // hidden marker is not asked about.
    let p = Project::in_git();
    p.old_state();
    p.write(&format!("{OLD}/reports/r.md"), "# r\n");
    p.rules(&[hides_all()]);
    git_ok(&p.root, &["add", "--", &format!("{OLD}/reports/r.md")]);
    p.expect_offered_and_applies("tracked only where nothing moves");

    let p = Project::plain();
    p.old_state();
    p.rules(&[hides_all()]);
    p.expect_offered_and_applies("no git work tree");

    let mut p = Project::plain();
    p.old_state();
    p.rules(&[hides_all()]);
    p.without_git();
    p.expect_offered_and_applies("no git work tree and no git program");
}

/// SPEC-20261126-tracked-state-move AC-8 (REQ-8)
#[test]
fn other_kept_names_or_backups_keep_the_move_blocked() {
    let receipt = format!("{NEW}/backups/{STAMP}/{GO_STATE_RECEIPT}");
    let local = (format!("{NEW}/{LOCAL_DIR}/output.log"), "x\n");
    let layouts: [(&str, Vec<(String, &str)>); 4] = [
        (
            "another kept name",
            vec![(format!("{NEW}/evals/e.md"), "# e\n")],
        ),
        (
            "SPECs and backups",
            vec![
                (
                    format!("{NEW}/{SPECS_DIR}/SPEC-20260102-new/spec.md"),
                    "#\n",
                ),
                (format!("{NEW}/backups/note.txt"), "n\n"),
            ],
        ),
        (
            "local and a reverted receipt",
            vec![local.clone(), (receipt, "{\"state\":\"reverted\"}\n")],
        ),
        (
            "local and backups without a receipt",
            vec![local, (format!("{NEW}/backups/note.txt"), "n\n")],
        ),
    ];
    for (layout, files) in layouts {
        let p = Project::in_git();
        p.old_state();
        for (rel, body) in &files {
            p.write(rel, body);
        }
        p.rules(&[hides_all()]);
        p.track();
        let before = tree(&p.root);
        let plan = p.plan();
        assert!(plan.get("ignored_tracked").is_none(), "{layout}: {plan}");
        assert_eq!(plan["state_moves"], json!([]), "{layout}: {plan}");
        let step = p.step();
        assert_eq!(step["state"], json!("blocked"), "{layout}: {step}");
        assert_eq!(step["apply"], Value::Null, "{layout}: {step}");
        assert_eq!(tree(&p.root), before, "{layout}: nothing changes");
    }
}

/// The record is refused with a message naming `path`, and `outside` stays as it was.
fn expect_refused_link(p: &Project, path: &str, outside: &Path) {
    let before = tree(outside);
    let refused = p.record().expect_err("a record through a link");
    assert!(
        refused.contains(&format!("{path}: ")) && refused.contains("symlink"),
        "{refused}"
    );
    assert_eq!(tree(outside), before, "nothing is written through the link");
}

/// A folder outside the project holding one file and a `local` folder.
fn outside_of(p: &Project) -> PathBuf {
    let outside = p.base.join("outside");
    fs::create_dir_all(outside.join(LOCAL_DIR)).unwrap();
    fs::write(outside.join("keep.txt"), "keep\n").unwrap();
    outside
}

/// SPEC-20261126-tracked-state-move AC-9 (REQ-9, REQ-10)
#[test]
fn a_linked_state_folder_refuses_the_flaky_record() {
    let rel = format!("{NEW}/{LOCAL_DIR}");
    // To a folder with a `local` folder, to an empty folder, and to nothing.
    for target in ["with local", "empty", "missing"] {
        let p = Project::plain();
        let outside = outside_of(&p);
        let to = match target {
            "with local" => outside.clone(),
            "empty" => {
                fs::create_dir(outside.join("empty")).unwrap();
                outside.join("empty")
            }
            _ => outside.join("nowhere"),
        };
        symlink(&to, p.root.join(NEW)).unwrap();
        expect_refused_link(&p, &rel, &outside);
    }
}

/// SPEC-20261126-tracked-state-move AC-10 (REQ-10)
#[test]
fn a_linked_local_folder_refuses_the_flaky_record() {
    let rel = format!("{NEW}/{LOCAL_DIR}");
    for target in ["a folder", "missing"] {
        let p = Project::plain();
        let outside = outside_of(&p);
        let to = match target {
            "a folder" => outside.clone(),
            _ => outside.join("nowhere"),
        };
        fs::create_dir(p.root.join(NEW)).unwrap();
        symlink(&to, p.root.join(&rel)).unwrap();
        expect_refused_link(&p, &rel, &outside);
    }
}

/// SPEC-20261126-tracked-state-move AC-11 (REQ-10)
#[test]
fn a_linked_ledger_file_refuses_the_flaky_record() {
    let rel = format!("{NEW}/{LOCAL_DIR}/{LEDGER}");
    for target in ["keep.txt", "nowhere"] {
        let p = Project::plain();
        let outside = outside_of(&p);
        fs::create_dir_all(p.root.join(NEW).join(LOCAL_DIR)).unwrap();
        symlink(outside.join(target), p.ledger()).unwrap();
        expect_refused_link(&p, &rel, &outside);
        assert!(
            fs::symlink_metadata(p.ledger())
                .unwrap()
                .file_type()
                .is_symlink(),
            "the link stays"
        );
    }
}

/// SPEC-20261126-tracked-state-move AC-12 (REQ-9)
#[test]
fn a_real_local_folder_takes_the_flaky_record() {
    let p = Project::plain();
    // The state folder and its local folder are missing at first.
    for rows in 1..=2 {
        p.record().unwrap();
        assert!(fs::symlink_metadata(p.ledger()).unwrap().is_file());
        let text = fs::read_to_string(p.ledger()).unwrap();
        assert_eq!(text.lines().count(), rows, "{text}");
        for line in text.lines() {
            let row: Value = serde_json::from_str(line).unwrap();
            assert_eq!(row["name"], json!("moai-cli::demo_cli t_flaky"), "{row}");
        }
    }
}

/// SPEC-20261126-tracked-state-move AC-12 (REQ-9): only a real file takes a row, and a folder
/// that allows no new file takes none.
#[test]
fn a_ledger_that_is_no_regular_file_refuses_the_flaky_record() {
    let rel = format!("{NEW}/{LOCAL_DIR}/{LEDGER}");

    let p = Project::plain();
    fs::create_dir_all(p.ledger()).unwrap();
    let refused = p.record().expect_err("a folder where the ledger belongs");
    assert!(refused.contains(&rel), "{refused}");

    // A FIFO would hold the writer until someone reads it, so the record runs beside a bound.
    let p = Project::plain();
    fs::create_dir_all(p.root.join(NEW).join(LOCAL_DIR)).unwrap();
    let made = Command::new("mkfifo").arg(p.ledger()).status().unwrap();
    assert!(made.success(), "mkfifo");
    let (root, out) = (p.root.clone(), p.base.join("check"));
    let (tx, rx) = mpsc::channel();
    std::thread::spawn(move || tx.send(record_in(&root, &out)));
    let answer = rx
        .recv_timeout(Duration::from_secs(30))
        .expect("the record opened the FIFO and waits for a reader");
    let refused = answer.expect_err("a FIFO where the ledger belongs");
    assert!(
        refused.contains(&rel) && refused.contains("not a regular file"),
        "{refused}"
    );

    let p = Project::plain();
    let local = p.root.join(NEW).join(LOCAL_DIR);
    fs::create_dir_all(&local).unwrap();
    fs::set_permissions(&local, fs::Permissions::from_mode(0o500)).unwrap();
    let refused = p.record();
    fs::set_permissions(&local, fs::Permissions::from_mode(0o700)).unwrap();
    assert!(refused.expect_err("no permission").contains(&rel));
    assert!(!p.ledger().exists());
}

/// SPEC-20261126-tracked-state-move AC-17 (REQ-11)
#[test]
fn a_flaky_record_beside_backups_alone_keeps_the_verdict() {
    let backups = format!("{NEW}/backups");
    let mut problems = Vec::new();
    for earlier in [false, true] {
        for layout in ["empty", "a file", "an empty stamp folder", "a Finder file"] {
            let p = Project::in_git();
            if earlier {
                p.old_state();
            }
            fs::create_dir_all(p.root.join(&backups)).unwrap();
            match layout {
                "a file" => p.write(&format!("{backups}/note.txt"), "n\n"),
                "an empty stamp folder" => {
                    fs::create_dir(p.root.join(&backups).join(STAMP)).unwrap()
                }
                "a Finder file" => p.write(&format!("{NEW}/.DS_Store"), "x"),
                _ => {}
            }
            p.rules(&[hides_all()]);
            p.track();
            let answer = |p: &Project| {
                let plan = p.plan();
                (
                    p.generation(),
                    p.step()["state"].clone(),
                    plan["layout"].clone(),
                    plan["state_moves"].clone(),
                )
            };
            let before = answer(&p);
            assert_ne!(before.0, "none", "{layout}, earlier {earlier}: the premise");
            p.record().unwrap();
            assert!(p.ledger().is_file(), "the premise");
            let after = answer(&p);
            if after != before {
                problems.push(format!(
                    "{layout}, earlier {earlier}: before {before:?}, after {after:?}"
                ));
            }
        }
    }
    assert!(problems.is_empty(), "{}", problems.join("\n"));
}

/// SPEC-20261126-tracked-state-move AC-18 (REQ-12)
#[test]
fn a_reverted_receipt_beside_a_local_folder_still_reads_none() {
    let receipt = format!("{NEW}/backups/{STAMP}/{GO_STATE_RECEIPT}");
    let reverted = "{\"state\":\"reverted\"}\n";

    let p = Project::plain();
    p.write(&receipt, reverted);
    p.write(&format!("{NEW}/{LOCAL_DIR}/output.log"), "x\n");
    assert_eq!(p.generation(), "none");
    p.record().unwrap();
    assert_eq!(p.generation(), "none");

    // The record itself adds the local folder.
    let p = Project::plain();
    p.write(&receipt, reverted);
    p.record().unwrap();
    assert_eq!(p.generation(), "none");

    // An open receipt still says the move stopped midway.
    let p = Project::plain();
    p.write(&receipt, "{\"state\":\"pending\"}\n");
    p.write(&format!("{NEW}/{LOCAL_DIR}/output.log"), "x\n");
    assert_eq!(p.generation(), "interrupted");
    p.record().unwrap();
    assert_eq!(p.generation(), "interrupted");
}
