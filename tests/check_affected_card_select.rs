//! SPEC-20261124-card-select-batch-land: `check affected --card` picks the tests a card's own
//! change reaches (its changed or added tests, the tests that name a changed module, the touched
//! scans) and falls back to the full set for a path whose effect is not clear. Temporary git
//! repositories and fake `cargo` and `rustfmt` scripts only.
mod common;
use common::git::{git, git_ok};
use serde_json::Value;
use std::fs;
use std::path::Path;
use std::process::{Command, Stdio};

const BIN: &str = crate::common::MOAI_BIN;

const SCANS: &str = r#"{"about":"fixture","scans":[
  {"id":"ledger_rows","checks":["cargo test --offline --locked --test scan_ledger_rows"],
   "scope":[{"prefix":"src/ledger/","suffixes":[".rs"]}],"rules":["src/ledger/sweep.rs"]},
  {"id":"docs_only","checks":["cargo test --offline --locked --test scan_docs_only"],
   "scope":[{"prefix":"docs/"}],"rules":["docs/note.md"]},
  {"id":"reads_src","always":true,"checks":["cargo test --offline --locked --test scan_reads_src"],
   "scope":[{"prefix":"src/","suffixes":[".rs"]}],"rules":["tests/scan_reads_src.rs"]},
  {"id":"reads_tests","checks":["cargo test --offline --locked --test scan_reads_tests"],
   "scope":[{"prefix":"tests/"},{"prefix":"src/"}],"rules":["tests/scan_reads_tests.rs"]}]}"#;

fn put(root: &Path, rel: &str, text: &str) {
    let path = root.join(rel);
    fs::create_dir_all(path.parent().unwrap()).unwrap();
    fs::write(path, text).unwrap();
}

/// A repository whose `tests/` holds twelve `card_other_*` targets (the word `card` is common), the
/// tests that name modules, two scans, and one commit on top of the base that changes `changed`.
struct Fixture {
    dir: tempfile::TempDir,
    base: String,
    head: String,
}

impl Fixture {
    fn new(changed: &[(&str, &str)]) -> Fixture {
        let dir = tempfile::tempdir().unwrap();
        let p = dir.path();
        git_ok(p, &["init", "-q", "-b", "main"]);
        git_ok(p, &["config", "user.email", "t@example.invalid"]);
        git_ok(p, &["config", "user.name", "t"]);
        put(p, "README.md", "x\n");
        put(p, "src/ledger/sweep.rs", "// a\n");
        put(p, "src/ledger/schema.rs", "// a\n");
        put(p, "src/ledger/migrate.rs", "// a\n");
        put(p, "src/card/sweep.rs", "// a\n");
        put(p, "src/card/work.rs", "// a\n");
        put(p, "src/ops/card.rs", "// a\n");
        put(p, "src/hook_guards.rs", "// a\n");
        put(p, "Cargo.toml", "[package]\nname = \"x\"\n");
        put(p, "Cargo.lock", "# lock\n");
        put(p, "docs/note.md", "note\n");
        for n in 0..12 {
            put(p, &format!("tests/card_other_{n}.rs"), "fn main() {}\n");
        }
        put(
            p,
            "tests/ledger_sweep.rs",
            "use moai_cli::ledger::sweep::run;\nfn main() { run(); }\n",
        );
        put(
            p,
            "tests/pins_card_work.rs",
            "use moai_cli::card::work;\nfn main() { work(); }\n",
        );
        put(p, "tests/scan_ledger_rows.rs", "fn main() {}\n");
        put(p, "tests/scan_docs_only.rs", "fn main() {}\n");
        put(p, "tests/scan_reads_src.rs", "fn main() {}\n");
        put(p, "tests/scan_reads_tests.rs", "fn main() {}\n");
        put(
            p,
            "tests/names_the_note.rs",
            "// reads docs/note.md\nfn main() {}\n",
        );
        put(p, "tools/repo_scans.json", SCANS);
        git_ok(p, &["add", "-A"]);
        git_ok(p, &["commit", "-q", "-m", "base"]);
        let base = git(p, &["rev-parse", "HEAD"]).trim().to_string();
        for (rel, text) in changed {
            put(p, rel, text);
        }
        git_ok(p, &["add", "-A"]);
        git_ok(p, &["commit", "-q", "-m", "change"]);
        let head = git(p, &["rev-parse", "HEAD"]).trim().to_string();
        Fixture { dir, base, head }
    }

    fn root(&self) -> &Path {
        self.dir.path()
    }

    fn range(&self) -> String {
        format!("{}..{}", self.base, self.head)
    }
}

fn affected(fx: &Fixture, extra: &[&str], path: Option<&str>) -> (bool, Value) {
    let mut cmd = Command::new(BIN);
    cmd.args(["check", "affected", &fx.range(), "--dir"])
        .arg(fx.root())
        .args(extra)
        .env_remove("MOAI_HOME")
        .env_remove("CARGO_TARGET_DIR")
        .stdin(Stdio::null());
    if let Some(p) = path {
        cmd.env("PATH", p);
    }
    let out = cmd.output().unwrap();
    let text = String::from_utf8_lossy(&out.stdout).to_string();
    let v: Value = serde_json::from_str(&text)
        .unwrap_or_else(|_| panic!("{text} / {}", String::from_utf8_lossy(&out.stderr)));
    let ok = v["ok"] == Value::Bool(true);
    (
        ok,
        if ok {
            v["result"].clone()
        } else {
            v["error"].clone()
        },
    )
}

fn names(v: &Value) -> Vec<String> {
    v.as_array()
        .unwrap()
        .iter()
        .map(|s| s.as_str().unwrap().to_string())
        .collect()
}

/// A range of modules that are no hub: ledger and card modules and an added test, shaped like
/// C-464's change without the shared paths it also touched (`C464_HUBS`).
const NON_HUB_RANGE: [(&str, &str); 4] = [
    ("src/ledger/sweep.rs", "// b\n"),
    ("src/card/sweep.rs", "// b\n"),
    ("src/card/work.rs", "// b\n"),
    ("tests/card_sweep_cli.rs", "fn main() {}\n"),
];

/// The shared paths C-464 changed besides its modules. Many tests rest on them, so a change here
/// runs the full set and says why (decision note 8636).
const C464_HUBS: [&str; 4] = [
    "src/ledger/mod.rs",
    "src/ledger/reservation.rs",
    "src/ops/card.rs",
    "src/profile.rs",
];

#[test]
fn card_set_of_non_hub_modules_is_small_and_names_no_ignored_word() {
    let fx = Fixture::new(&NON_HUB_RANGE);
    let (ok, card) = affected(&fx, &["--list", "--card"], None);
    assert!(ok, "{card}");
    assert_eq!(card["selection"], "card", "{card}");
    let rust = names(&card["rust_targets"]);
    assert!(rust.len() <= 20, "{} targets: {rust:?}", rust.len());
    // The added test, the tests that name a changed module, and the scan the range touches.
    for want in [
        "card_sweep_cli",
        "ledger_sweep",
        "pins_card_work",
        "scan_ledger_rows",
    ] {
        assert!(rust.contains(&want.to_string()), "{want} in {rust:?}");
    }
    // The area rule of the full set would take every `card_*` target; the card set does not.
    for n in 0..12 {
        let t = format!("card_other_{n}");
        assert!(!rust.contains(&t), "{t} in {rust:?}");
    }
    assert!(!rust.contains(&"scan_docs_only".to_string()), "{rust:?}");
    // A scan that reads the whole tree is the landing's and the batch's, not the card's.
    for broad in ["scan_reads_src", "scan_reads_tests"] {
        assert!(!rust.contains(&broad.to_string()), "{broad} in {rust:?}");
    }
    // No path of the range names a word that more than 10 targets share.
    assert!(names(&card["card_ignored_words"]).is_empty(), "{card}");
    assert!(card["card_wide"].is_null(), "{card}");
    // The full set of the same range is the larger one (the bare command is not changed).
    let (ok, full) = affected(&fx, &["--list"], None);
    assert!(ok, "{full}");
    assert_eq!(full["selection"], "full");
    assert!(
        names(&full["rust_targets"]).len() > rust.len(),
        "{full} / {card}"
    );
}

#[test]
fn card_set_holds_a_changed_test_a_named_module_and_a_touched_scan_and_nothing_else() {
    let fx = Fixture::new(&[
        ("tests/card_other_3.rs", "fn main() { /* edited */ }\n"),
        ("docs/note.md", "changed note\n"),
        ("src/card/work.rs", "// b\n"),
    ]);
    let (ok, r) = affected(&fx, &["--list", "--card"], None);
    assert!(ok, "{r}");
    let rust: std::collections::BTreeSet<String> = names(&r["rust_targets"]).into_iter().collect();
    // The changed test, the module-naming test, the test that names the changed doc, the scan
    // the doc touches.
    let want: std::collections::BTreeSet<String> = [
        "card_other_3",
        "pins_card_work",
        "names_the_note",
        "scan_docs_only",
    ]
    .iter()
    .map(|s| s.to_string())
    .collect();
    assert_eq!(rust, want, "{r}");
}

#[test]
fn wide_classes_fall_back_to_the_full_set_and_say_why() {
    for (path, reason) in [
        ("Cargo.lock", "dependencies"),
        ("Cargo.toml", "dependencies"),
        ("src/ledger/schema.rs", "ledger storage"),
        ("src/ledger/migrate.rs", "ledger storage"),
        ("src/hook_guards.rs", "permission"),
        (".moai/project.yaml", "common settings"),
        (".cargo/config.toml", "build settings"),
        // The audit's table of paths that picked no test and fell back to nothing (B1): the audit
        // floor, the git hooks and the landing code, a word more than 10 targets share, and a
        // source file no test can name.
        ("src/gate.rs", "audit floor"),
        ("install/git-hooks/pre-push", "audit floor"),
        ("src/git_hooks/mod.rs", "audit floor"),
        ("tools/local_check.sh", "audit floor"),
        ("REVIEW.md", "audit floor"),
        ("src/card/land/check.rs", "audit floor"),
        ("src/ops/card.rs", "common name word"),
        ("src/ledger/mod.rs", "no test reached"),
        ("src/lonely.rs", "no test reached"),
    ] {
        let fx = Fixture::new(&[(path, "# changed\n"), ("src/card/work.rs", "// b\n")]);
        let (ok, r) = affected(&fx, &["--list", "--card"], None);
        assert!(ok, "{path}: {r}");
        assert_eq!(r["selection"], "full", "{path}: {r}");
        assert_eq!(r["card_wide"]["reason"], reason, "{path}: {r}");
        assert_eq!(
            r["card_wide"]["paths"],
            serde_json::json!([path]),
            "{path}: {r}"
        );
        // The set is the one the bare command picks for the same range.
        let (ok, full) = affected(&fx, &["--list"], None);
        assert!(ok, "{full}");
        assert_eq!(r["rust_targets"], full["rust_targets"], "{path}");
    }
}

#[test]
fn wide_a_common_word_runs_the_full_set_and_still_names_the_word() {
    let fx = Fixture::new(&[("src/ops/card.rs", "// b\n")]);
    let (ok, r) = affected(&fx, &["--list", "--card"], None);
    assert!(ok, "{r}");
    assert_eq!(r["selection"], "full", "{r}");
    assert_eq!(r["card_wide"]["reason"], "common name word", "{r}");
    assert_eq!(names(&r["card_ignored_words"]), ["card"], "{r}");
}

#[test]
fn wide_the_hub_paths_of_c464_run_the_full_set_and_name_each_path() {
    // The range of C-464 as it was: its modules, its added test and its four hub paths.
    let mut changed: Vec<(&str, &str)> = NON_HUB_RANGE.to_vec();
    changed.extend(C464_HUBS.iter().map(|p| (*p, "// b\n")));
    let fx = Fixture::new(&changed);
    let (ok, r) = affected(&fx, &["--list", "--card"], None);
    assert!(ok, "{r}");
    assert_eq!(r["selection"], "full", "{r}");
    let paths = names(&r["card_wide"]["paths"]);
    for hub in C464_HUBS {
        assert!(paths.contains(&hub.to_string()), "{hub} in {paths:?}: {r}");
    }
    // Every hub path carries a reason, and the modules beside them do not add a path of theirs.
    assert!(r["card_wide"]["reason"].is_string(), "{r}");
    for module in [
        "src/card/sweep.rs",
        "src/card/work.rs",
        "src/ledger/sweep.rs",
    ] {
        assert!(!paths.contains(&module.to_string()), "{module}: {r}");
    }
}

#[test]
fn wide_classes_do_not_catch_the_rest_of_the_ledger() {
    let fx = Fixture::new(&[("src/ledger/sweep.rs", "// b\n")]);
    let (ok, r) = affected(&fx, &["--list", "--card"], None);
    assert!(ok, "{r}");
    assert_eq!(r["selection"], "card", "{r}");
    assert!(r["card_wide"].is_null(), "{r}");
}

#[test]
fn card_and_lane_cannot_be_asked_together() {
    let fx = Fixture::new(&[("src/card/work.rs", "// b\n")]);
    let (ok, r) = affected(&fx, &["--list", "--card", "--lane"], None);
    assert!(!ok, "{r}");
}

/// A folder of fake `cargo` and `rustfmt` that exit 0, in front of the system's `git`.
fn fake_path() -> (tempfile::TempDir, String) {
    let dir = tempfile::tempdir().unwrap();
    for n in ["cargo", "rustfmt"] {
        let path = dir.path().join(n);
        fs::write(&path, "#!/bin/sh\nexit 0\n").unwrap();
        #[cfg(unix)]
        {
            use std::os::unix::fs::PermissionsExt;
            fs::set_permissions(&path, fs::Permissions::from_mode(0o755)).unwrap();
        }
    }
    let git_dir = Command::new("sh")
        .args(["-c", "dirname \"$(command -v git)\""])
        .output()
        .unwrap();
    let path = format!(
        "{}:{}:/usr/bin:/bin",
        dir.path().display(),
        String::from_utf8_lossy(&git_dir.stdout).trim()
    );
    (dir, path)
}

fn slot_dir() -> tempfile::TempDir {
    let d = tempfile::tempdir().unwrap();
    fs::write(
        d.path().join("limits.json"),
        r#"{"max":4,"min_free_gb":0,"max_swap_pct":100,"min_mem_free_pct":0}"#,
    )
    .unwrap();
    d
}

#[test]
fn receipt_a_card_run_writes_selection_card_and_a_wide_fallback_writes_none() {
    let (_tools, path) = fake_path();
    let slots = slot_dir();
    let slot = slots.path().to_string_lossy().to_string();
    // A narrowed run says so after `scans=`.
    let fx = Fixture::new(&[("src/card/work.rs", "// b\n")]);
    let (ok, r) = affected(&fx, &["--card", "--slot-dir", slot.as_str()], Some(&path));
    assert!(ok, "{r}");
    let line = r["summary"].as_str().unwrap();
    assert!(line.ends_with(" result=PASS"), "{line}");
    let words: Vec<&str> = line.split(' ').collect();
    let scans = words.iter().position(|w| w.starts_with("scans=")).unwrap();
    assert_eq!(words.get(scans + 1), Some(&"selection=card"), "{line}");
    let out = Path::new(r["out"].as_str().unwrap());
    assert_eq!(
        fs::read_to_string(out.join("summary.txt")).unwrap(),
        format!("{line}\n")
    );
    // A fallback to the full set is a bare run for the reader of the receipt.
    let fx = Fixture::new(&[("Cargo.lock", "# changed\n")]);
    let (ok, r) = affected(&fx, &["--card", "--slot-dir", slot.as_str()], Some(&path));
    assert!(ok, "{r}");
    let line = r["summary"].as_str().unwrap();
    assert!(!line.contains("selection="), "{line}");
    assert_eq!(r["selection"], "full", "{r}");
}
