//! SPEC-20261124-registry-line-add-tier: a change that only grows the scan registry
//! (`tools/repo_scans.json`: a rule file, a test name or a `--test` option added, a scan added)
//! does not set the audit floor to tier 2 by its path; a change that edits or removes anything
//! the registry already said stays tier 2. The library function is judged on fixtures; the
//! `audit_floor` of `moai bundle check` is judged in a temporary repository. Every command runs
//! against a temporary HOME; no checked-in file is read.
#![cfg(unix)]

mod common;
#[path = "common/ready_receipt.rs"]
mod ready_receipt;

use common::home::TempHome;
use moai_cli::gate::{registry_only_grew, tier_of, tier_of_changes, ChangedFile};
use moai_cli::ledger;
use serde_json::{json, Value};
use std::fs;
use std::io::Write;
use std::path::PathBuf;
use std::process::Stdio;

const REG: &str = "tools/repo_scans.json";

fn scan(id: &str, tests: &[&str], checks: &[&str], rules: &[&str]) -> Value {
    json!({
        "id": id,
        "tests": tests,
        "checks": checks,
        "rules": rules,
        "scope": [{"prefix": "src/", "suffixes": [".rs"]}],
        "rejects": "a change the scan guards",
    })
}

fn registry(scans: Vec<Value>) -> String {
    serde_json::to_string_pretty(&json!({"about": "fixture", "scans": scans})).unwrap()
}

const CHECK: &str = "cargo test --offline --locked --test alpha --test gamma";

fn base_scans() -> Vec<Value> {
    vec![
        scan(
            "reads-src",
            &["tests/alpha.rs::t"],
            &[CHECK, "cargo test --offline --locked --lib one"],
            &["tests/alpha.rs", "tests/gamma.rs"],
        ),
        scan(
            "second",
            &["src/x.rs::t"],
            &["cargo test --lib x"],
            &["src/x.rs"],
        ),
    ]
}

fn base() -> String {
    registry(base_scans())
}

fn file(path: &str, old: &str, new: &str) -> ChangedFile {
    ChangedFile {
        path: path.to_string(),
        old: Some(old.to_string()),
        new: Some(new.to_string()),
    }
}

fn floor(files: &[ChangedFile]) -> u8 {
    tier_of_changes(Some(0), files)
}

/// The floor of the one registry change from `base()` to the scans `edit` returns.
fn floor_after(edit: impl Fn(&mut Vec<Value>)) -> u8 {
    let mut scans = base_scans();
    edit(&mut scans);
    floor(&[file(REG, &base(), &registry(scans))])
}

#[test]
fn a_registration_keeps_the_floor_at_tier_1() {
    // The real shape of a new test file's registration: a rule file and a `--test` option.
    assert_eq!(
        floor_after(|s| {
            s[0]["rules"] = json!(["tests/alpha.rs", "tests/beta.rs", "tests/gamma.rs"]);
            s[0]["checks"][0] =
                json!("cargo test --offline --locked --test alpha --test beta --test gamma");
        }),
        1
    );
    // A test name, a second `--test` and a scan added.
    assert_eq!(
        floor_after(|s| {
            s[0]["tests"] = json!(["tests/alpha.rs::t", "tests/beta.rs::t"]);
            s[0]["checks"][0] = json!(format!("{CHECK} --test delta --test epsilon"));
            s.push(scan(
                "third",
                &["tests/z.rs"],
                &["cargo test --test z"],
                &["tests/z.rs"],
            ));
        }),
        1
    );
    // Nothing changed at all is no change of the floor either.
    assert_eq!(floor_after(|_| {}), 1);
}

#[test]
fn the_names_alone_still_give_tier_2() {
    assert_eq!(tier_of(Some(0), &[REG.to_string()]), 2);
    assert_eq!(floor(&[ChangedFile::named(REG)]), 2);
    // A side that is missing is never growth.
    let only_new = ChangedFile {
        path: REG.to_string(),
        old: None,
        new: Some(base()),
    };
    let only_old = ChangedFile {
        path: REG.to_string(),
        old: Some(base()),
        new: None,
    };
    assert_eq!(floor(&[only_new]), 2);
    assert_eq!(floor(&[only_old]), 2);
}

#[test]
fn an_edited_or_removed_line_is_tier_2() {
    // A text field of a scan edited.
    assert_eq!(floor_after(|s| s[0]["rejects"] = json!("another text")), 2);
    assert_eq!(
        floor_after(|s| s[1]["scope"] = json!([{"prefix": "docs/"}])),
        2
    );
    assert_eq!(floor_after(|s| s[1]["id"] = json!("renamed")), 2);
    // A rule file or a test name removed, or replaced by another.
    assert_eq!(
        floor_after(|s| s[0]["rules"] = json!(["tests/alpha.rs"])),
        2
    );
    assert_eq!(
        floor_after(|s| s[0]["rules"] = json!(["tests/alpha.rs", "tests/other.rs"])),
        2
    );
    assert_eq!(floor_after(|s| s[0]["tests"] = json!([])), 2);
    // A whole scan removed.
    assert_eq!(
        floor_after(|s| {
            s.pop();
        }),
        2
    );
}

#[test]
fn a_check_line_changed_in_any_other_way_is_tier_2() {
    // A `--test` option removed.
    assert_eq!(
        floor_after(|s| s[0]["checks"][0] = json!("cargo test --offline --locked --test alpha")),
        2
    );
    // A `--test` option replaced by another one.
    assert_eq!(
        floor_after(|s| {
            s[0]["checks"][0] = json!("cargo test --offline --locked --test alpha --test other")
        }),
        2
    );
    // Any other word added or changed: a flag, a filter, another program, the profile.
    for line in [
        format!("{CHECK} --release"),
        format!("{CHECK} -- --ignored"),
        format!("{CHECK} --lib zero"),
        "cargo test --offline --test alpha --test gamma".to_string(),
        format!("{CHECK} ; true"),
        "true".to_string(),
    ] {
        assert_eq!(
            floor_after(|s| s[0]["checks"][0] = json!(line)),
            2,
            "{line}"
        );
    }
    // A `--test` with no name after it is not a registration.
    assert_eq!(
        floor_after(|s| s[0]["checks"][0] = json!(format!("{CHECK} --test"))),
        2
    );
    // A check line added or removed.
    assert_eq!(
        floor_after(|s| s[0]["checks"] = json!([
            CHECK,
            "cargo test --offline --locked --lib one",
            "cargo test --lib two"
        ])),
        2
    );
    assert_eq!(floor_after(|s| s[0]["checks"] = json!([CHECK])), 2);
}

#[test]
fn a_line_that_only_adds_text_but_overrides_an_old_value_is_tier_2() {
    // One added line, nothing removed: a repeated key after the old one wins when the text is read.
    let old = base();
    let marker = "\"rules\": [\n        \"tests/alpha.rs\",\n        \"tests/gamma.rs\"\n      ],";
    assert!(old.contains(marker), "{old}");
    let new = old.replacen(marker, &format!("{marker}\n      \"rules\": [],"), 1);
    assert_eq!(new.lines().count(), old.lines().count() + 1);
    assert!(old.lines().all(|l| new.lines().any(|n| n == l)));
    assert_eq!(floor(&[file(REG, &old, &new)]), 2);
    // The same line at the top level.
    let new = old.replacen(
        "\"about\": \"fixture\",",
        "\"about\": \"fixture\",\n  \"about\": \"\",",
        1,
    );
    assert_eq!(floor(&[file(REG, &old, &new)]), 2);
}

#[test]
fn a_registry_of_another_shape_is_tier_2() {
    let old = base();
    let scans = base_scans();
    // A key added at the top level, a key of a scan added, a scan with a repeated id.
    let extra =
        serde_json::to_string_pretty(&json!({"about": "fixture", "scans": scans, "more": 1}))
            .unwrap();
    assert_eq!(floor(&[file(REG, &old, &extra)]), 2);
    assert_eq!(floor_after(|s| s[0]["note"] = json!("x")), 2);
    assert_eq!(floor_after(|s| s.push(scan("second", &[], &[], &[]))), 2);
    // About text changed, text that is not JSON, JSON that is not a registry.
    let about =
        serde_json::to_string_pretty(&json!({"about": "other", "scans": base_scans()})).unwrap();
    assert_eq!(floor(&[file(REG, &old, &about)]), 2);
    assert_eq!(floor(&[file(REG, &old, "{ not json")]), 2);
    assert_eq!(floor(&[file(REG, "{ not json", &old)]), 2);
    assert_eq!(floor(&[file(REG, &old, "[]")]), 2);
    assert_eq!(
        floor(&[file(REG, &old, "{\"about\":\"fixture\",\"scans\":\"x\"}")]),
        2
    );
    assert!(!registry_only_grew(REG, &old, "{}"));
}

#[test]
fn only_the_registry_path_is_judged_by_its_text() {
    let old = base();
    let mut scans = base_scans();
    scans[0]["rules"] = json!(["tests/alpha.rs", "tests/beta.rs", "tests/gamma.rs"]);
    let new = registry(scans);
    assert!(registry_only_grew(REG, &old, &new));
    // The same texts under any other name keep the path rule: the other tier 2 prefixes, a
    // longer name that starts with the registry's, the protected-check list next to it.
    for path in [
        "tools/repo_scans.json.orig",
        "tools/protected_checks.json",
        "tools/self_check.sh",
        "src/gate.rs",
        ".github/workflows/ci.yml",
        "tools/repo_scans.jsonl",
    ] {
        assert!(!registry_only_grew(path, &old, &new), "{path}");
        assert_eq!(floor(&[file(path, &old, &new)]), 2, "{path}");
    }
}

#[test]
fn another_tier_2_path_in_the_change_keeps_the_floor_at_tier_2() {
    let old = base();
    let mut scans = base_scans();
    scans[0]["rules"] = json!(["tests/alpha.rs", "tests/beta.rs", "tests/gamma.rs"]);
    let new = registry(scans);
    let grown = file(REG, &old, &new);
    assert_eq!(floor(&[grown.clone()]), 1);
    for other in [
        "src/gate.rs",
        "src/hook_entry.rs",
        ".claude/settings.json",
        "CLAUDE.md",
        "tools/check_guard.py",
    ] {
        assert_eq!(
            floor(&[grown.clone(), ChangedFile::named(other)]),
            2,
            "{other}"
        );
        assert_eq!(
            floor(&[ChangedFile::named(other), grown.clone()]),
            2,
            "{other}"
        );
    }
    // Paths that are not tier 2 leave a registration at its own tier 1, tests and code alike.
    assert_eq!(
        floor(&[grown.clone(), ChangedFile::named("tests/beta.rs")]),
        1
    );
    assert_eq!(
        floor(&[grown.clone(), ChangedFile::named("src/other.rs")]),
        1
    );
    // A registration never lowers a change below tier 1 and the body never lowers it.
    assert_eq!(tier_of_changes(None, &[grown.clone()]), 2);
    assert_eq!(tier_of_changes(Some(2), &[grown.clone()]), 2);
    assert_eq!(tier_of_changes(Some(7), &[grown]), 2);
}

#[test]
fn the_two_other_registries_are_tier_0_for_any_change() {
    // `tests/areas.toml` and the state folder allow list sit under `tests/`: tier 0 whether a
    // line is added or edited, so the registry rule has nothing to do for them.
    for path in [
        "tests/areas.toml",
        "tests/fixtures/state_folder_literals/allow.tsv",
    ] {
        assert_eq!(floor(&[ChangedFile::named(path)]), 0, "{path}");
        assert_eq!(floor(&[file(path, "a\n", "a\nb\n")]), 0, "{path}");
        assert_eq!(floor(&[file(path, "a\n", "c\n")]), 0, "{path}");
    }
}

// ---- `moai bundle check` in a temporary repository -------------------------------------------

struct World {
    home: TempHome,
    _root: tempfile::TempDir,
    project: String,
    work: PathBuf,
}

impl World {
    fn new() -> World {
        let home = TempHome::new();
        let project = home.project("demo");
        ledger::register(home.path(), "demo").unwrap();
        let root = tempfile::Builder::new()
            .prefix("moai-test-registry-tier-")
            .tempdir()
            .unwrap();
        let work = root.path().join("work");
        fs::create_dir_all(&work).unwrap();
        let w = World {
            home,
            _root: root,
            project: project.to_str().unwrap().to_string(),
            work,
        };
        w.git(&["init", "-q", "-b", "trunk"]);
        w.put("README.md", "tip\n");
        w.put(REG, &base());
        w.git(&["add", "-A"]);
        w.git(&["commit", "-q", "-m", "chore: the tip"]);
        w
    }

    fn put(&self, path: &str, text: &str) {
        let f = self.work.join(path);
        fs::create_dir_all(f.parent().unwrap()).unwrap();
        fs::write(f, text).unwrap();
    }

    fn git(&self, args: &[&str]) -> String {
        let mut all = vec![
            "-c",
            "user.name=Test",
            "-c",
            "user.email=test@example.invalid",
            "-c",
            "commit.gpgsign=false",
        ];
        all.extend_from_slice(args);
        let out = self
            .home
            .command_for("git", &all)
            .current_dir(&self.work)
            .stdin(Stdio::null())
            .output()
            .expect("run git");
        assert!(
            out.status.success(),
            "git {args:?}: {}",
            String::from_utf8_lossy(&out.stderr)
        );
        String::from_utf8_lossy(&out.stdout).trim().to_string()
    }

    /// A branch of one commit on the tip with these files; returns the commit.
    fn head(&self, branch: &str, files: &[(&str, &str)]) -> String {
        self.git(&["checkout", "-q", "-B", branch, "trunk"]);
        for (p, c) in files {
            self.put(p, c);
        }
        self.git(&["add", "-A"]);
        self.git(&["commit", "-q", "-m", &format!("feat: {branch}")]);
        self.git(&["rev-parse", "HEAD"])
    }

    fn run(&self, args: &[&str], stdin: Option<&str>) -> (Value, i32) {
        let mut cmd = self.home.command(args);
        cmd.stdin(if stdin.is_some() {
            Stdio::piped()
        } else {
            Stdio::null()
        })
        .stdout(Stdio::piped())
        .stderr(Stdio::piped());
        let mut child = cmd.spawn().expect("run moai");
        if let Some(t) = stdin {
            child.stdin.take().unwrap().write_all(t.as_bytes()).unwrap();
        }
        let out = child.wait_with_output().unwrap();
        let line = String::from_utf8_lossy(&out.stdout)
            .lines()
            .next()
            .unwrap_or("")
            .to_string();
        let v: Value = serde_json::from_str(&line)
            .unwrap_or_else(|e| panic!("{args:?}: {e}: {}", String::from_utf8_lossy(&out.stderr)));
        (v, out.status.code().unwrap_or(-1))
    }

    fn card(&self, n: u32) -> String {
        let scope = format!("area{n}/");
        let (v, _) = self.run(
            &[
                "card",
                "add",
                "--title",
                "work",
                "--scope",
                &scope,
                "--ac",
                "done",
                "--project",
                &self.project,
            ],
            None,
        );
        assert_eq!(v["ok"], true, "{v}");
        v["result"]["name"].as_str().unwrap().to_string()
    }

    /// The `audit_floor` `bundle check` answers for a head of this repository.
    fn audit_floor(&self, n: u32, head: &str) -> Value {
        let card = self.card(n);
        ready_receipt::add_note_unchecked(
            self.home.path(),
            "demo",
            &card,
            &format!("LANE-READY-INTEGRATION {card} head={head} spec=SPEC-20261124-demo"),
        );
        let (v, _) = self.run(
            &[
                "bundle",
                "check",
                &card,
                "--project",
                &self.project,
                "--dir",
                self.work.to_str().unwrap(),
                "--tip",
                "trunk",
            ],
            None,
        );
        assert_eq!(v["ok"], true, "{v}");
        v["result"]["heads"][0]["audit_floor"].clone()
    }
}

#[test]
fn bundle_check_reads_a_registration_as_tier_1_and_an_edit_as_tier_2() {
    let w = World::new();
    let mut grown = base_scans();
    grown[0]["rules"] = json!(["tests/alpha.rs", "tests/beta.rs", "tests/gamma.rs"]);
    grown[0]["checks"][0] =
        json!("cargo test --offline --locked --test alpha --test beta --test gamma");
    let mut edited = base_scans();
    edited[0]["rejects"] = json!("another text");
    let mut removed = base_scans();
    removed[0]["rules"] = json!(["tests/alpha.rs"]);

    let registered = w.head(
        "spec/registered",
        &[(REG, &registry(grown.clone())), ("tests/beta.rs", "// t\n")],
    );
    let with_code = w.head(
        "spec/with-gate",
        &[(REG, &registry(grown)), ("src/gate.rs", "// g\n")],
    );
    let edit = w.head("spec/edited", &[(REG, &registry(edited))]);
    let remove = w.head("spec/removed", &[(REG, &registry(removed))]);
    let docs = w.head("spec/docs", &[("notes/x.md", "x\n")]);

    assert_eq!(w.audit_floor(1, &registered), 1, "a registration is tier 1");
    assert_eq!(
        w.audit_floor(2, &with_code),
        2,
        "another tier 2 path keeps tier 2"
    );
    assert_eq!(w.audit_floor(3, &edit), 2, "an edited line is tier 2");
    assert_eq!(
        w.audit_floor(4, &remove),
        2,
        "a removed rule file is tier 2"
    );
    assert_eq!(w.audit_floor(5, &docs), 0, "a document is still tier 0");
}

#[test]
fn bundle_check_judges_a_head_behind_the_tip_by_its_own_registry_change() {
    let w = World::new();
    let mut grown = base_scans();
    grown[0]["rules"] = json!(["tests/alpha.rs", "tests/beta.rs", "tests/gamma.rs"]);
    let head = w.head("spec/registered", &[(REG, &registry(grown))]);
    // The tip moves on with its own, different edit of the registry; the head is judged against
    // its merge base, so its own change is still the only one that counts.
    let mut moved = base_scans();
    moved[1]["rejects"] = json!("the tip changed this");
    w.git(&["checkout", "-q", "trunk"]);
    w.put(REG, &registry(moved));
    w.git(&["add", "-A"]);
    w.git(&["commit", "-q", "-m", "feat: the tip moves"]);
    assert_eq!(w.audit_floor(1, &head), 1);
}
