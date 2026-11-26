//! SPEC-20261126-affected-card-readers: `check affected` also picks the tests that read a changed
//! source module in a way the name and path rules miss: a grouped `use` in a test whose name
//! carries the file's word, and a file word with an underscore held as a string literal. The rules
//! only add targets. Temporary git repositories, and the checked-in test files for one test.
mod common;
use common::git::{git, git_ok};
use moai_cli::check_affected::{choose, read_facts, Pick};
use serde_json::Value;
use std::fs;
use std::path::Path;
use std::process::{Command, Stdio};

const BIN: &str = crate::common::MOAI_BIN;

/// The area table of the fixture: `prop` is in another area than `card` and `ledger`.
const AREAS: &str = "[areas]\ncard = [\"card\", \"ledger\"]\nrepo = [\"prop\"]\nspec = [\"spec\", \"check\"]\nupdate = [\"update\"]\n\n[solo]\nfiles = []\n";

fn put(root: &Path, rel: &str, text: &str) {
    let path = root.join(rel);
    fs::create_dir_all(path.parent().unwrap()).unwrap();
    fs::write(path, text).unwrap();
}

/// A repository with an area table, the given test files in its base commit, and one commit on top
/// that changes the source files `changed`.
struct Fixture {
    dir: tempfile::TempDir,
    range: String,
}

impl Fixture {
    fn new(tests: &[(String, String)], changed: &[&str]) -> Fixture {
        let dir = tempfile::tempdir().unwrap();
        let p = dir.path();
        git_ok(p, &["init", "-q", "-b", "main"]);
        git_ok(p, &["config", "user.email", "t@example.invalid"]);
        git_ok(p, &["config", "user.name", "t"]);
        put(p, "notes.txt", "x\n");
        put(p, "tests/areas.toml", AREAS);
        for (name, text) in tests {
            put(p, &format!("tests/{name}.rs"), text);
        }
        for src in changed {
            put(p, src, "// a\n");
        }
        git_ok(p, &["add", "-A"]);
        git_ok(p, &["commit", "-q", "-m", "base"]);
        let base = git(p, &["rev-parse", "HEAD"]).trim().to_string();
        for src in changed {
            put(p, src, "// b\n");
        }
        git_ok(p, &["add", "-A"]);
        git_ok(p, &["commit", "-q", "-m", "change"]);
        let head = git(p, &["rev-parse", "HEAD"]).trim().to_string();
        Fixture {
            dir,
            range: format!("{base}..{head}"),
        }
    }

    /// The `result` of `check affected <range> --list` with `extra` flags.
    fn list(&self, extra: &[&str]) -> Value {
        let out = Command::new(BIN)
            .args(["check", "affected", &self.range, "--list", "--dir"])
            .arg(self.dir.path())
            .args(extra)
            .env_remove("MOAI_HOME")
            .env_remove("CARGO_TARGET_DIR")
            .stdin(Stdio::null())
            .output()
            .unwrap();
        let text = String::from_utf8_lossy(&out.stdout).to_string();
        let v: Value = serde_json::from_str(&text)
            .unwrap_or_else(|_| panic!("{text} / {}", String::from_utf8_lossy(&out.stderr)));
        assert_eq!(v["ok"], Value::Bool(true), "{v}");
        v["result"].clone()
    }
}

fn targets(answer: &Value) -> Vec<String> {
    answer["rust_targets"]
        .as_array()
        .unwrap()
        .iter()
        .map(|s| s.as_str().unwrap().to_string())
        .collect()
}

/// The reasons the card answer gives for `path`.
fn reasons(answer: &Value, path: &str) -> Vec<String> {
    answer["card_picks"]
        .as_array()
        .unwrap()
        .iter()
        .filter(|p| p["path"] == path)
        .flat_map(|p| p["by"].as_array().unwrap())
        .map(|s| s.as_str().unwrap().to_string())
        .collect()
}

fn file(name: &str, text: &str) -> (String, String) {
    (name.to_string(), text.to_string())
}

/// A test that holds `word` as a string literal and imports no module.
fn holds(name: &str, word: &str) -> (String, String) {
    file(name, &format!("fn main() {{ let _ = [\"{word}\"]; }}\n"))
}

/// Twelve tests outside the `card` area that name the reservation module in one piece, so more
/// than ten targets reach it, and the two tests of the `card` area.
fn common_module_readers() -> Vec<(String, String)> {
    let mut out: Vec<(String, String)> = (0..12)
        .map(|n| {
            file(
                &format!("update_leases_{n}"),
                "use moai_cli::ledger::reservation::lease;\nfn main() { lease(); }\n",
            )
        })
        .collect();
    out.push(file("card_base_cli", "fn main() {}\n"));
    out.push(file("ledger_base", "fn main() {}\n"));
    out
}

#[test]
fn a_changed_module_selects_the_test_that_reads_its_answer_key() {
    // A path of the audit floor: the card run answers the full set, which holds the reader too.
    let fx = Fixture::new(
        &[holds("update_end_to_end_cli", "state_root")],
        &["src/state_root.rs"],
    );
    let card = fx.list(&["--card"]);
    assert_eq!(card["selection"], "full", "{card}");
    assert_eq!(card["card_wide"]["reason"], "audit floor", "{card}");
    for answer in [card, fx.list(&[]), fx.list(&["--lane"])] {
        assert!(
            targets(&answer).contains(&"update_end_to_end_cli".to_string()),
            "{answer}"
        );
    }
    // An ordinary path takes the card set: the reader joins it and the path says why.
    let fx = Fixture::new(
        &[
            holds("update_pipeline_key_cli", "check_pipeline"),
            file("spec_base_cli", "fn main() {}\n"),
        ],
        &["src/check_pipeline.rs"],
    );
    let card = fx.list(&["--card"]);
    assert_eq!(card["selection"], "card", "{card}");
    assert!(card["card_wide"].is_null(), "{card}");
    assert!(
        reasons(&card, "src/check_pipeline.rs").contains(&"module:check_pipeline".to_string()),
        "{card}"
    );
    for answer in [card, fx.list(&[]), fx.list(&["--lane"])] {
        assert!(
            targets(&answer).contains(&"update_pipeline_key_cli".to_string()),
            "{answer}"
        );
    }
}

#[test]
fn a_grouped_import_of_a_common_module_is_selected() {
    let mut tests = common_module_readers();
    tests.push(file(
        "prop_reservation",
        "use moai_cli::ledger::{reservation, workflow};\nfn main() {}\n",
    ));
    tests.push(file(
        "prop_reservation_lines",
        "use moai_cli::ledger::{\n    // the leases, and their ends\n    reservation,\n    workflow,\n};\nfn main() {}\n",
    ));
    tests.push(file(
        "prop_reservation_nested",
        "use moai_cli::{card, ledger::{self, reservation}};\nfn main() {}\n",
    ));
    let fx = Fixture::new(&tests, &["src/ledger/reservation.rs"]);
    let card = fx.list(&["--card"]);
    assert_eq!(card["selection"], "card", "{card}");
    assert!(card["card_wide"].is_null(), "{card}");
    for answer in [card.clone(), fx.list(&[]), fx.list(&["--lane"])] {
        let got = targets(&answer);
        for want in [
            "prop_reservation",
            "prop_reservation_lines",
            "prop_reservation_nested",
        ] {
            assert!(got.contains(&want.to_string()), "{want}: {answer}");
        }
    }
    // Every contiguous reader is outside the card area, so the word names only the new picks.
    assert!(
        reasons(&card, "src/ledger/reservation.rs")
            .contains(&"module:ledger_reservation".to_string()),
        "{card}"
    );
}

#[test]
fn a_look_alike_name_or_literal_adds_nothing() {
    // The grouped rule asks for the file's word in the name and for the module path.
    let mut tests = common_module_readers();
    for (name, text) in [
        ("prop_reservation_name_only", "fn main() {}\n"),
        (
            "prop_other",
            "use moai_cli::ledger::{reservation, workflow};\nfn main() {}\n",
        ),
        (
            "prop_reservation_other_item",
            "use moai_cli::ledger::{reservations, workflow};\nfn main() {}\n",
        ),
        (
            "prop_reservation_other_parent",
            "use other::subledger::{reservation};\nfn main() {}\n",
        ),
        (
            "prop_literals",
            "fn main() { let _ = [\"reservation\", \"ledger_reservation\"]; }\n",
        ),
    ] {
        tests.push(file(name, text));
    }
    let fx = Fixture::new(&tests, &["src/ledger/reservation.rs"]);
    for extra in [&["--card"][..], &[], &["--lane"]] {
        let answer = fx.list(extra);
        let got = targets(&answer);
        assert!(!got.iter().any(|t| t.starts_with("prop_")), "{answer}");
    }
    // A literal counts only as the whole word of a file name with an underscore.
    let fx = Fixture::new(
        &[
            holds("update_plural_cli", "state_roots"),
            holds("update_prefixed_cli", "my_state_root"),
        ],
        &["src/state_root.rs"],
    );
    for extra in [&["--card"][..], &[], &["--lane"]] {
        let answer = fx.list(extra);
        assert!(targets(&answer).is_empty(), "{answer}");
    }
    let fx = Fixture::new(
        &[holds("update_one_word_cli", "lonely")],
        &["src/lonely.rs"],
    );
    for extra in [&["--card"][..], &[], &["--lane"]] {
        let answer = fx.list(extra);
        assert!(targets(&answer).is_empty(), "{answer}");
    }
}

#[test]
fn the_two_missed_tests_of_this_repository_are_selected() {
    let facts = read_facts(Path::new(env!("CARGO_MANIFEST_DIR")), "HEAD..HEAD").unwrap();
    // No scan list and no host: only the rules that read the test files pick.
    let pick = |pick: Pick, path: &str| {
        choose(
            pick,
            &[path.to_string()],
            &facts.tests,
            None,
            facts.areas_text.as_deref(),
            "none",
        )
    };
    let full = pick(Pick::Full, "src/state_root.rs");
    assert!(full.selection.rust.contains("update_end_to_end_cli"));
    let card = pick(Pick::Card, "src/ledger/reservation.rs");
    assert_eq!(card.pick, Pick::Card);
    assert!(card.selection.rust.contains("prop_reservation"));
    let full = pick(Pick::Full, "src/ledger/reservation.rs");
    assert!(full.selection.rust.contains("prop_reservation"));
}
