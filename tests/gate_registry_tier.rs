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
use moai_cli::audit_gate::{self, Input, RegistryTexts};
use moai_cli::card::land::check::floor_tier;
use moai_cli::forge::transport::{Exchange, HttpRequest, HttpResponse, RecordedTransport};
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

/// The floor of the one registry change that turns the first check line of the first scan from
/// `old` into `new`.
fn floor_of_line(old: &str, new: &str) -> u8 {
    let mut before = base_scans();
    before[0]["checks"][0] = json!(old);
    let mut after = base_scans();
    after[0]["checks"][0] = json!(new);
    floor(&[file(REG, &registry(before), &registry(after))])
}

/// Audit B1 (P1 to P6): a `--test` pair counts as growth only when its value is a target name
/// (`repo_scans` reads `[a-z0-9_]+`), so an option, a filter or a shell fragment in the value
/// slot turns a red check green without being a registration.
#[test]
fn a_test_option_whose_value_is_no_target_name_is_tier_2() {
    for value in [
        "--list",
        "--exact",
        "--skip",
        "--ignored",
        "x||true",
        "x;true",
        "x&&true",
        "$(true)",
        "X",
        "a-b",
        "a.b",
        "a/b",
        "--",
    ] {
        let line = format!("{CHECK} --test {value}");
        assert_eq!(floor_of_line(CHECK, &line), 2, "{line}");
    }
    // The same names between the old words are no registration either.
    assert_eq!(
        floor_of_line(
            CHECK,
            "cargo test --offline --locked --test --list --test alpha --test gamma"
        ),
        2
    );
    // The control: a plain target name is.
    assert_eq!(floor_of_line(CHECK, &format!("{CHECK} --test beta_2")), 1);
}

/// Audit B1: after the first `--` the words are libtest's, so a `--test` there is no cargo
/// option, whatever its value is.
#[test]
fn a_test_option_after_the_libtest_separator_is_tier_2() {
    const OLD: &str = "cargo test --offline --locked --test alpha -- --nocapture";
    for tail in [
        "--test beta",
        "--test --list",
        "--test --exact",
        "--test --skip",
        "--test --ignored",
        "--test x||true",
        "--test x;true",
    ] {
        let line = format!("{OLD} {tail}");
        assert_eq!(floor_of_line(OLD, &line), 2, "{line}");
    }
    // Inserted between the old words after the separator.
    assert_eq!(
        floor_of_line(
            OLD,
            "cargo test --offline --locked --test alpha -- --test beta --nocapture"
        ),
        2
    );
    // The control: the same pair before the separator is a registration.
    assert_eq!(
        floor_of_line(
            OLD,
            "cargo test --offline --locked --test alpha --test beta -- --nocapture"
        ),
        1
    );
}

/// The growth rule reads a name as `repo_scans` reads a `test_targets` entry: a name that one
/// accepts and the other refuses would make a registration the table cannot run, or the other
/// way round.
#[test]
fn a_test_option_value_is_read_as_repo_scans_reads_a_target_name() {
    // One word each: a name with a blank is two words and never one name.
    for name in [
        "alpha", "a_1", "0", "_", "A", "a-b", "a.b", "x;true", "--list", "ünï",
    ] {
        let entry = json!({"id": "x", "test_targets": [name]});
        let accepted = moai_cli::repo_scans::checks_of(&entry).is_ok();
        let grew = floor_of_line(CHECK, &format!("{CHECK} --test {name}")) == 1;
        assert_eq!(grew, accepted, "{name:?}");
        // The one function both read (REQ-11).
        assert_eq!(
            moai_cli::repo_scans::is_target_name(name),
            accepted,
            "{name:?}"
        );
    }
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

// ---- `test_targets` (SPEC-20261126-gate-test-targets-growth) ---------------------------------

/// A scan in the per-line form: its `test_targets` as given, one `checks` line that is no plain
/// target list.
fn tt_scan(id: &str, targets: Value) -> Value {
    json!({
        "id": id,
        "tests": ["tests/alpha.rs::t"],
        "test_targets": targets,
        "checks": ["cargo test --offline --locked --lib one"],
        "rules": ["tests/alpha.rs", "tests/gamma.rs"],
        "scope": [{"prefix": "src/", "suffixes": [".rs"]}],
        "rejects": "a change the scan guards",
    })
}

fn tt_base_scans() -> Vec<Value> {
    vec![
        tt_scan("reads-src", json!(["alpha", "gamma"])),
        tt_scan("second", json!(["beta", "delta"])),
    ]
}

fn tt_base() -> String {
    registry(tt_base_scans())
}

/// The floor of the one registry change from `tt_base()` to the scans `edit` returns.
fn tt_floor_after(edit: impl Fn(&mut Vec<Value>)) -> u8 {
    let mut scans = tt_base_scans();
    edit(&mut scans);
    floor(&[file(REG, &tt_base(), &registry(scans))])
}

/// The floor of the change that turns the first scan's `test_targets` from `old` into `new`.
fn tt_floor_between(old: Value, new: Value) -> u8 {
    let mut before = tt_base_scans();
    before[0]["test_targets"] = old;
    let mut after = tt_base_scans();
    after[0]["test_targets"] = new;
    floor(&[file(REG, &registry(before), &registry(after))])
}

/// REQ-1: names inserted at their sorted places keep the low floor, wherever they go.
#[test]
fn a_target_name_added_to_test_targets_keeps_the_floor_at_tier_1() {
    for (what, new) in [
        ("in the middle", json!(["alpha", "beta", "gamma"])),
        ("at the start", json!(["aaa", "alpha", "gamma"])),
        ("at the end", json!(["alpha", "gamma", "zeta"])),
        (
            "several at once",
            json!(["a1", "alpha", "beta", "gamma", "zeta_2"]),
        ),
        (
            "digits and an underscore",
            json!(["0", "alpha", "g_1", "gamma"]),
        ),
    ] {
        assert_eq!(
            tt_floor_between(json!(["alpha", "gamma"]), new),
            1,
            "{what}"
        );
    }
    // Both scans grow, next to a grown rule list.
    assert_eq!(
        tt_floor_after(|s| {
            s[0]["test_targets"] = json!(["alpha", "beta", "gamma"]);
            s[1]["test_targets"] = json!(["beta", "delta", "epsilon"]);
            s[0]["rules"] = json!(["tests/alpha.rs", "tests/beta.rs", "tests/gamma.rs"]);
        }),
        1
    );
    // A scan added with its own `test_targets`.
    assert_eq!(
        tt_floor_after(|s| s.push(tt_scan("third", json!(["zeta"])))),
        1
    );
    // The library answer is the same one.
    let mut grown = tt_base_scans();
    grown[0]["test_targets"] = json!(["alpha", "beta", "gamma"]);
    assert!(registry_only_grew(REG, &tt_base(), &registry(grown)));
}

/// REQ-2: every other change of the list is no growth.
#[test]
fn a_test_targets_change_that_is_no_growth_is_tier_2() {
    let old = json!(["alpha", "gamma"]);
    for (what, new) in [
        ("a name removed", json!(["alpha"])),
        ("every name removed", json!([])),
        ("a name replaced", json!(["alpha", "delta"])),
        (
            "a name replaced and one added",
            json!(["alpha", "beta", "delta"]),
        ),
        (
            "an unsorted insert at the end",
            json!(["alpha", "gamma", "beta"]),
        ),
        (
            "an unsorted insert at the start",
            json!(["beta", "alpha", "gamma"]),
        ),
        ("the old order reversed", json!(["gamma", "alpha"])),
        ("a name twice", json!(["alpha", "alpha", "gamma"])),
        ("the last name twice", json!(["alpha", "gamma", "gamma"])),
        ("an uppercase name", json!(["Alpha", "alpha", "gamma"])),
        ("a name with a dash", json!(["a-b", "alpha", "gamma"])),
        ("a name with a dot", json!(["a.b", "alpha", "gamma"])),
        ("a name with a slash", json!(["a/b", "alpha", "gamma"])),
        ("a shell fragment", json!(["alpha", "gamma", "x;true"])),
        ("an option", json!(["--list", "alpha", "gamma"])),
        ("an empty name", json!(["", "alpha", "gamma"])),
        ("a number", json!(["alpha", "gamma", 7])),
        ("a null in the list", json!(["alpha", "gamma", null])),
        ("null instead of a list", json!(null)),
        ("a string instead of a list", json!("alpha")),
        ("an object instead of a list", json!({"alpha": 1})),
    ] {
        assert_eq!(
            tt_floor_between(old.clone(), new.clone()),
            2,
            "{what}: {new}"
        );
    }
    // An empty list at the base is no registration to grow from.
    assert_eq!(
        tt_floor_between(json!([]), json!(["alpha"])),
        2,
        "from empty"
    );
    // A list that was no list of names stays no list of names when a good one is added.
    assert_eq!(
        tt_floor_between(json!(["Bad", "alpha"]), json!(["Bad", "alpha", "beta"])),
        2,
        "a bad name at both sides"
    );
    assert_eq!(
        tt_floor_between(json!("alpha"), json!(["alpha"])),
        2,
        "from a string"
    );
}

/// REQ-1 and REQ-2 (audit repair): growth is names added to a list that was itself sorted and
/// free of repeats. A base that is not, and a head that only reorders or de-duplicates it, are no
/// growth, whether or not the head is the sorted, de-duplicated form (the three cases the audit
/// probed, and the unsorted or repeated base with a name really added).
#[test]
fn a_test_targets_base_that_is_not_sorted_and_unique_is_no_growth_even_beside_a_sorted_head() {
    for (what, old, new) in [
        (
            "a reorder, no name added",
            json!(["gamma", "alpha"]),
            json!(["alpha", "gamma"]),
        ),
        (
            "a repeat removed, no name added",
            json!(["a", "a"]),
            json!(["a"]),
        ),
        (
            "an unsorted base and a name added",
            json!(["gamma", "alpha"]),
            json!(["alpha", "beta", "gamma"]),
        ),
        (
            "a base with a repeat and a name added",
            json!(["a", "a"]),
            json!(["a", "b"]),
        ),
        (
            "a base with a last name twice and a name added",
            json!(["alpha", "gamma", "gamma"]),
            json!(["alpha", "beta", "delta", "gamma"]),
        ),
        (
            "the same length, a name replaced",
            json!(["alpha", "gamma"]),
            json!(["alpha", "delta"]),
        ),
    ] {
        assert_eq!(tt_floor_between(old, new.clone()), 2, "{what}: {new}");
    }
    // The controls: a sorted base with one name added, and a sorted base with two.
    assert_eq!(
        tt_floor_between(json!(["alpha", "gamma"]), json!(["alpha", "beta", "gamma"])),
        1
    );
    assert_eq!(tt_floor_between(json!(["a"]), json!(["a", "b", "c"])), 1);
}

/// REQ-3 and REQ-6: an unchanged list is no difference, whatever it holds; a key that appears or
/// goes is a change of the table's shape; the other fields of the scan are judged as before.
#[test]
fn an_unchanged_test_targets_is_no_difference_and_a_changed_key_set_is_tier_2() {
    let grown_rules = |s: &mut Vec<Value>| {
        s[0]["rules"] = json!(["tests/alpha.rs", "tests/beta.rs", "tests/gamma.rs"]);
    };
    // The same list of any content beside a registration of another kind.
    for list in [
        json!(["Bad", "alpha"]),
        json!([]),
        json!(null),
        json!("alpha"),
        json!(["gamma", "alpha"]),
    ] {
        let mut before = tt_base_scans();
        before[0]["test_targets"] = list.clone();
        let mut after = before.clone();
        grown_rules(&mut after);
        assert_eq!(
            floor(&[file(REG, &registry(before), &registry(after))]),
            1,
            "unchanged {list}"
        );
    }
    // A `test_targets` key added to a scan, and one taken away: the conversion's own shape.
    let mut converted = base_scans();
    converted[0]["test_targets"] = json!(["alpha"]);
    assert_eq!(
        floor(&[file(REG, &base(), &registry(converted.clone()))]),
        2,
        "key added"
    );
    assert_eq!(
        floor(&[file(REG, &registry(converted), &base())]),
        2,
        "key removed"
    );
    // The scan's other fields are judged as before, beside a grown list.
    let grow = |s: &mut Vec<Value>| s[0]["test_targets"] = json!(["alpha", "beta", "gamma"]);
    assert_eq!(
        tt_floor_after(|s| {
            grow(s);
            s[0]["rejects"] = json!("another text");
        }),
        2,
        "a text edited"
    );
    assert_eq!(
        tt_floor_after(|s| {
            grow(s);
            s[0]["checks"][0] = json!("cargo test --offline --locked --lib other");
        }),
        2,
        "a check line edited"
    );
    assert_eq!(
        tt_floor_after(|s| {
            grow(s);
            s[0]["rules"] = json!(["tests/alpha.rs"]);
        }),
        2,
        "a rule file removed"
    );
    assert_eq!(
        tt_floor_after(|s| {
            grow(s);
            s[0]["checks"][0] = json!("cargo test --offline --locked --lib one --test omega");
        }),
        1,
        "a check line that gains a test option, the old form"
    );
    // The two forms in one table: the list shrinks while a check line gains a test option.
    assert_eq!(
        tt_floor_after(|s| {
            s[0]["test_targets"] = json!(["alpha"]);
            s[0]["checks"][0] = json!("cargo test --offline --locked --lib one --test omega");
        }),
        2,
        "a list that shrinks beside an old-form registration"
    );
}

/// The growth rule reads a name as `repo_scans` reads a `test_targets` entry: a name that one
/// accepts and the other refuses would make a registration the table cannot run, or the other way
/// round (REQ-4).
#[test]
fn a_test_targets_name_is_read_as_repo_scans_reads_a_target_name() {
    for name in [
        "alpha",
        "a_1",
        "0",
        "_",
        "A",
        "a-b",
        "a.b",
        "x;true",
        "--list",
        "ünï",
        "",
        "alpha beta",
    ] {
        let entry = json!({"id": "x", "test_targets": [name]});
        let accepted = moai_cli::repo_scans::checks_of(&entry).is_ok();
        // The old list is one valid name; the new list is sorted, so only the name decides.
        let mut new = vec!["m_old", name];
        new.sort_unstable();
        let grew = tt_floor_between(json!(["m_old"]), json!(new)) == 1;
        assert_eq!(grew, accepted, "{name:?}");
    }
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
    assert_eq!(floor(std::slice::from_ref(&grown)), 1);
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
    assert_eq!(tier_of_changes(None, std::slice::from_ref(&grown)), 2);
    assert_eq!(tier_of_changes(Some(2), std::slice::from_ref(&grown)), 2);
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
        World::with_registry(&base())
    }

    /// A world whose tip holds this text as the registry.
    fn with_registry(text: &str) -> World {
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
        w.put(REG, text);
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

    /// The tier the landing check judges `head` by when no verdict states more than 0.
    fn landing_floor(&self, head: &str) -> u64 {
        u64::from(floor_tier(&self.work, "trunk", head, 0).expect("landing floor"))
    }

    /// The text of `path` at `rev`, or `None` when git cannot show it.
    fn show(&self, rev: &str, path: &str) -> Option<String> {
        let out = self
            .home
            .command_for("git", &["show", &format!("{rev}:{path}")])
            .current_dir(&self.work)
            .stdin(Stdio::null())
            .output()
            .expect("run git");
        out.status
            .success()
            .then(|| String::from_utf8_lossy(&out.stdout).into_owned())
    }

    /// The tier `audit-gate` decides `head` by, from what GitHub would hand it: the changed
    /// files with both names of a rename, and the registry texts at the merge base and the head.
    fn audit_gate_tier(&self, head: &str) -> u64 {
        let listed = self.git(&[
            "diff",
            "--name-only",
            "--no-renames",
            &format!("trunk...{head}"),
        ]);
        let files: Vec<&str> = listed.lines().collect();
        let base = self.git(&["merge-base", "trunk", head]);
        let registry = files.contains(&REG).then(|| RegistryTexts {
            old: self.show(&base, REG),
            new: self.show(head, REG),
        });
        let mut input = gate_input(&files, None);
        input.registry = registry;
        gate_tier(&input)
    }

    /// A branch of one commit on the tip that renames the registry to `to`.
    fn renamed_head(&self, branch: &str, to: &str) -> String {
        self.git(&["checkout", "-q", "-B", branch, "trunk"]);
        self.git(&["mv", REG, to]);
        self.git(&["commit", "-q", "-m", &format!("feat: {branch}")]);
        self.git(&["rev-parse", "HEAD"])
    }

    /// A branch of one commit on the tip in which the registry path is a symbolic link (mode
    /// 120000) whose blob holds `target`, made with git's plumbing so the text may be any length;
    /// returns the commit. The working tree is left as it was.
    fn linked_head(&self, branch: &str, target: &str) -> String {
        self.git(&["checkout", "-q", "-B", branch, "trunk"]);
        self.put("link-payload.tmp", target);
        let blob = self.git(&["hash-object", "-w", "link-payload.tmp"]);
        fs::remove_file(self.work.join("link-payload.tmp")).unwrap();
        self.git(&[
            "update-index",
            "--cacheinfo",
            &format!("120000,{blob},{REG}"),
        ]);
        self.git(&["commit", "-q", "-m", &format!("feat: {branch}")]);
        self.git(&["rev-parse", "HEAD"])
    }

    /// Moves the tip on with its own edit of the registry, after the heads were made.
    fn move_the_tip(&self) {
        let mut moved = base_scans();
        moved[1]["rejects"] = json!("the tip changed this");
        self.git(&["checkout", "-q", "trunk"]);
        self.put(REG, &registry(moved));
        self.git(&["add", "-A"]);
        self.git(&["commit", "-q", "-m", "feat: the tip moves"]);
    }
}

/// The heads of the scenes every judgment is asked about, each with its floor.
fn scenes(w: &World) -> Vec<(&'static str, String, u64)> {
    let mut grown = base_scans();
    grown[0]["rules"] = json!(["tests/alpha.rs", "tests/beta.rs", "tests/gamma.rs"]);
    grown[0]["checks"][0] =
        json!("cargo test --offline --locked --test alpha --test beta --test gamma");
    let mut edited = base_scans();
    edited[0]["rejects"] = json!("another text");
    let mut removed = base_scans();
    removed[0]["rules"] = json!(["tests/alpha.rs"]);
    vec![
        (
            "a registration",
            w.head(
                "spec/registered",
                &[(REG, &registry(grown.clone())), ("tests/beta.rs", "// t\n")],
            ),
            1,
        ),
        (
            "a registration next to src/gate.rs",
            w.head(
                "spec/with-gate",
                &[(REG, &registry(grown)), ("src/gate.rs", "// g\n")],
            ),
            2,
        ),
        (
            "an edited line",
            w.head("spec/edited", &[(REG, &registry(edited))]),
            2,
        ),
        (
            "a removed rule file",
            w.head("spec/removed", &[(REG, &registry(removed))]),
            2,
        ),
        (
            "a document",
            w.head("spec/docs", &[("notes/x.md", "x\n")]),
            0,
        ),
    ]
}

#[test]
fn the_landing_floor_reads_a_registration_as_tier_1_and_an_edit_as_tier_2() {
    let w = World::new();
    let scenes = scenes(&w);
    // Every head is behind the tip once it has its own edit of the registry: each is judged by
    // its own change from the merge base.
    w.move_the_tip();
    for (what, head, floor) in &scenes {
        assert_eq!(w.landing_floor(head), *floor, "{what}");
    }
}

#[test]
fn the_three_judgments_give_one_tier_for_the_same_head() {
    let w = World::new();
    let scenes = scenes(&w);
    w.move_the_tip();
    for (n, (what, head, floor)) in scenes.iter().enumerate() {
        let bundle = w.audit_floor(n as u32 + 1, head).as_u64().unwrap();
        assert_eq!(bundle, *floor, "bundle check: {what}");
        assert_eq!(w.landing_floor(head), bundle, "landing: {what}");
        assert_eq!(w.audit_gate_tier(head), bundle, "audit-gate: {what}");
    }
}

/// REQ-1 and REQ-2 through the three judgments, on a tip whose registry is in the per-line form:
/// a name added to `test_targets` is tier 1 for `bundle check`, the landing floor and `audit-gate`
/// alike, and the changes that are no growth are tier 2 for all three.
#[test]
fn the_three_judgments_give_one_tier_for_a_test_targets_registration() {
    let w = World::with_registry(&tt_base());
    let edit = |f: &dyn Fn(&mut Vec<Value>)| {
        let mut scans = tt_base_scans();
        f(&mut scans);
        registry(scans)
    };
    let scenes = [
        (
            "a name added",
            w.head(
                "spec/registered",
                &[
                    (
                        REG,
                        &edit(&|s| s[0]["test_targets"] = json!(["alpha", "beta", "gamma"])),
                    ),
                    ("tests/beta.rs", "// t\n"),
                ],
            ),
            1,
        ),
        (
            "a name added next to src/gate.rs",
            w.head(
                "spec/with-gate",
                &[
                    (
                        REG,
                        &edit(&|s| s[0]["test_targets"] = json!(["alpha", "beta", "gamma"])),
                    ),
                    ("src/gate.rs", "// g\n"),
                ],
            ),
            2,
        ),
        (
            "a name removed",
            w.head(
                "spec/removed",
                &[(REG, &edit(&|s| s[0]["test_targets"] = json!(["alpha"])))],
            ),
            2,
        ),
        (
            "an unsorted name",
            w.head(
                "spec/unsorted",
                &[(
                    REG,
                    &edit(&|s| s[0]["test_targets"] = json!(["alpha", "gamma", "beta"])),
                )],
            ),
            2,
        ),
        (
            "a list made empty",
            w.head(
                "spec/emptied",
                &[(REG, &edit(&|s| s[0]["test_targets"] = json!([])))],
            ),
            2,
        ),
    ];
    for (n, (what, head, floor)) in scenes.iter().enumerate() {
        let bundle = w.audit_floor(n as u32 + 1, head).as_u64().unwrap();
        assert_eq!(bundle, *floor, "bundle check: {what}");
        assert_eq!(w.landing_floor(head), bundle, "landing: {what}");
        assert_eq!(w.audit_gate_tier(head), bundle, "audit-gate: {what}");
    }
}

#[test]
fn a_renamed_registry_is_tier_2_in_bundle_check_and_the_landing_floor() {
    let w = World::new();
    // The same text under another name: the old name is removed, so nothing was registered.
    let head = w.renamed_head("spec/renamed", "tools/other.json");
    assert_eq!(w.audit_floor(1, &head), 2, "bundle check");
    assert_eq!(w.landing_floor(&head), 2, "landing");
    assert_eq!(w.audit_gate_tier(&head), 2, "audit-gate");
}

/// The registry replaced by a symbolic link is tier 2 for the two judgments that read git, whatever
/// the link's blob holds: a path, or the very text of a registration (the blob of a link has no
/// length limit, and `git show` prints it as a file's text).
#[test]
fn a_registry_that_is_a_symbolic_link_is_tier_2_in_bundle_check_and_the_landing_floor() {
    let mut grown = base_scans();
    grown[0]["rules"] = json!(["tests/alpha.rs", "tests/beta.rs", "tests/gamma.rs"]);
    grown[0]["checks"][0] =
        json!("cargo test --offline --locked --test alpha --test beta --test gamma");
    for (what, target) in [
        ("a link to another file", "tools/other.json".to_string()),
        ("a link whose blob is a registration", registry(grown)),
    ] {
        let w = World::new();
        let head = w.linked_head("spec/linked", &target);
        assert_eq!(w.audit_floor(1, &head), 2, "bundle check: {what}");
        assert_eq!(w.landing_floor(&head), 2, "landing: {what}");
    }
}

// ---- `audit-gate`: the input and the reads of GitHub --------------------------------------

const H: &str = "1111111111111111111111111111111111111111";
const B: &str = "3333333333333333333333333333333333333333";
const M: &str = "2222222222222222222222222222222222222222";
const API: &str = "https://api.github.com/repos/o/r";

/// An input of PR 9 that declares tier 0, changes `files` and carries one author.
fn gate_input(files: &[&str], registry: Option<(Option<&str>, Option<&str>)>) -> Input {
    Input {
        schema_version: 1,
        pr: 9,
        head: H.into(),
        body: "audit-tier: 0\nbody".into(),
        files: Some(files.iter().map(|s| s.to_string()).collect()),
        commit_messages: Some(vec!["feat: x\n\nMoai-Session: agent-1".into()]),
        comments: vec![],
        registry: registry.map(|(old, new)| RegistryTexts {
            old: old.map(String::from),
            new: new.map(String::from),
        }),
        release_base: None,
    }
}

fn gate_tier(input: &Input) -> u64 {
    audit_gate::decide(input)["tier"].as_u64().expect("a tier")
}

#[test]
fn audit_gate_reads_a_registration_as_tier_1_and_an_edit_as_tier_2() {
    let old = base();
    let mut grown = base_scans();
    grown[0]["rules"] = json!(["tests/alpha.rs", "tests/beta.rs", "tests/gamma.rs"]);
    let grown = registry(grown);
    let mut edited = base_scans();
    edited[0]["rejects"] = json!("another text");
    let edited = registry(edited);
    let mut removed = base_scans();
    removed[0]["rules"] = json!(["tests/alpha.rs"]);
    let removed = registry(removed);
    let tier = |files: &[&str], old: Option<&str>, new: Option<&str>| {
        gate_tier(&gate_input(files, Some((old, new))))
    };
    assert_eq!(tier(&[REG], Some(&old), Some(&grown)), 1, "a registration");
    assert_eq!(
        tier(&[REG, "src/gate.rs"], Some(&old), Some(&grown)),
        2,
        "next to another tier 2 path"
    );
    assert_eq!(tier(&[REG], Some(&old), Some(&edited)), 2, "an edit");
    assert_eq!(tier(&[REG], Some(&old), Some(&removed)), 2, "a removal");
    assert_eq!(tier(&[REG], None, Some(&grown)), 2, "no text at the base");
    assert_eq!(tier(&[REG], Some(&old), None), 2, "no text at the head");
    assert_eq!(
        tier(&[REG, "tools/other.json"], Some(&old), None),
        2,
        "a rename lists both names and has no text at the head"
    );
    assert_eq!(tier(&["notes/x.md"], None, None), 0, "a document");
}

#[test]
fn audit_gate_without_registry_texts_keeps_tier_2() {
    assert_eq!(gate_tier(&gate_input(&[REG], None)), 2);
    // An input file of the older shape (no `registry` field) still reads, and keeps tier 2 for
    // the registry whatever tier its body declares.
    let older = json!({"schema_version": 1, "pr": 9, "head": H, "body": "audit-tier: 1",
                       "files": [REG], "commit_messages": [], "comments": []});
    let input: Input = serde_json::from_value(older).expect("an older input reads");
    assert_eq!(input.registry, None);
    assert_eq!(gate_tier(&input), 2);
}

fn ex(url: String, status: u16, body: Value) -> Exchange {
    Exchange {
        request: HttpRequest::new("GET", url),
        response: HttpResponse::new(status).with_body(body),
    }
}

/// Base64 as the GitHub contents API writes it: lines of 60 characters.
fn b64(text: &str) -> String {
    const A: &[u8; 64] = b"ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789+/";
    let mut out = String::new();
    for chunk in text.as_bytes().chunks(3) {
        let n = chunk
            .iter()
            .enumerate()
            .fold(0u32, |n, (i, b)| n | (u32::from(*b) << (16 - 8 * i)));
        for i in 0..4 {
            if i <= chunk.len() {
                out.push(A[((n >> (18 - 6 * i)) & 63) as usize] as char);
            } else {
                out.push('=');
            }
        }
    }
    let lines: Vec<&str> = out
        .as_bytes()
        .chunks(60)
        .map(|c| std::str::from_utf8(c).unwrap())
        .collect();
    lines.join("\n") + "\n"
}

/// A 40-character object id made of one letter.
fn sha(c: char) -> String {
    c.to_string().repeat(40)
}

/// One entry of a Git Trees answer.
fn entry(path: &str, mode: &str, kind: &str, id: &str) -> Value {
    json!({"path": path, "mode": mode, "type": kind, "sha": id})
}

/// The Git Trees answer for `id` (a commit or a tree), not truncated.
fn tree(id: &str, entries: Vec<Value>) -> Exchange {
    ex(
        format!("{API}/git/trees/{id}"),
        200,
        json!({"sha": id, "truncated": false, "tree": entries}),
    )
}

/// A Git Blobs answer for `id` with this body.
fn blob(id: &str, body: Value) -> Exchange {
    ex(format!("{API}/git/blobs/{id}"), 200, body)
}

fn blob_text(id: &str, text: &str) -> Exchange {
    blob(
        id,
        json!({"sha": id, "encoding": "base64", "content": b64(text)}),
    )
}

/// The three reads of one commit when `tools/repo_scans.json` is a regular file: the root tree
/// of `rev`, the `tools` tree and the blob named by its entry.
fn side(rev: &str, tools: &str, file: &str, text: &str) -> Vec<Exchange> {
    vec![
        tree(rev, vec![entry("tools", "040000", "tree", tools)]),
        tree(
            tools,
            vec![entry("repo_scans.json", "100644", "blob", file)],
        ),
        blob_text(file, text),
    ]
}

/// The four reads `collect` always makes for PR 9, with `files` as the changed files.
fn pr_exchanges(files: Value) -> Vec<Exchange> {
    vec![
        ex(
            format!("{API}/pulls/9"),
            200,
            json!({"head": {"sha": H}, "base": {"sha": B}, "body": "audit-tier: 0", "commits": 1}),
        ),
        ex(format!("{API}/pulls/9/files?per_page=100"), 200, files),
        ex(
            format!("{API}/pulls/9/commits?per_page=100"),
            200,
            json!([{"commit": {"message": "x\n\nMoai-Session: agent-1"}}]),
        ),
        ex(
            format!("{API}/issues/9/comments?per_page=100"),
            200,
            json!([]),
        ),
    ]
}

fn compare(status: u16, body: Value) -> Exchange {
    ex(format!("{API}/compare/{B}...{H}?per_page=1"), status, body)
}

#[test]
fn collect_reads_the_registry_at_the_merge_base_and_the_head() {
    let mut grown = base_scans();
    grown[0]["rules"] = json!(["tests/alpha.rs", "tests/beta.rs", "tests/gamma.rs"]);
    let (old, new) = (base(), registry(grown));
    let mut all = pr_exchanges(json!([{"filename": REG}]));
    all.push(compare(200, json!({"merge_base_commit": {"sha": M}})));
    all.extend(side(M, &sha('a'), &sha('c'), &old));
    all.extend(side(H, &sha('b'), &sha('d'), &new));
    let t = RecordedTransport::new(all);
    let input = audit_gate::collect("o", "r", 9, None, &t).unwrap();
    assert_eq!(t.remaining(), 0);
    assert_eq!(
        input.registry,
        Some(RegistryTexts {
            old: Some(old),
            new: Some(new)
        })
    );
    assert_eq!(gate_tier(&input), 1, "a registration is tier 1");
}

#[test]
fn collect_reads_no_registry_text_for_other_files() {
    let t = RecordedTransport::new(pr_exchanges(json!([{"filename": "src/a.rs"}])));
    let input = audit_gate::collect("o", "r", 9, None, &t).unwrap();
    assert_eq!(t.remaining(), 0, "only the four reads of a pull request");
    assert_eq!(t.sent.borrow().len(), 4);
    assert_eq!(input.registry, None);
}

#[test]
fn collect_leaves_the_texts_out_when_a_side_cannot_be_read() {
    // The head's text is a real registration: only a text that was read as such may count it.
    let mut grown = base_scans();
    grown[0]["rules"] = json!(["tests/alpha.rs", "tests/beta.rs", "tests/gamma.rs"]);
    let (old, new) = (base(), registry(grown));
    let not_found = json!({"message": "Not Found"});
    let (a, b, c, d) = (sha('a'), sha('b'), sha('c'), sha('d'));
    let merge = compare(200, json!({"merge_base_commit": {"sha": M}}));
    // The reads of a head side that holds the real registration, and of a base side that holds
    // the real old text: a case breaks one side and keeps the other good, so only a text read
    // as a regular file may count.
    let good_new = || side(H, &b, &d, &new);
    let good_old = || side(M, &a, &c, &old);
    // A tools tree that holds `repo_scans.json` as `mode`/`kind`, for the head side: the reads
    // stop before any blob.
    let head_entry = |mode: &str, kind: &str| {
        vec![
            tree(H, vec![entry("tools", "040000", "tree", &b)]),
            tree(&b, vec![entry("repo_scans.json", mode, kind, &d)]),
        ]
    };
    let mut truncated_root = tree(H, vec![entry("tools", "040000", "tree", &b)]);
    truncated_root.response = HttpResponse::new(200).with_body(
        json!({"sha": H, "truncated": true, "tree": [entry("tools", "040000", "tree", &b)]}),
    );
    let mut no_truncated_key = tree(H, vec![entry("tools", "040000", "tree", &b)]);
    no_truncated_key.response = HttpResponse::new(200)
        .with_body(json!({"sha": H, "tree": [entry("tools", "040000", "tree", &b)]}));
    let mut bad_blob_id = side(H, &b, &d, &new);
    bad_blob_id[1] = tree(&b, vec![entry("repo_scans.json", "100644", "blob", "main")]);
    // Each case: what the base side reads, then what the head side reads. After the four reads
    // of the pull request and the merge base, every read it lists is made and no other.
    let cases: Vec<(&str, Vec<Exchange>, Vec<Exchange>)> = vec![
        (
            "a registry that is a symbolic link at the head",
            good_old(),
            head_entry("120000", "blob"),
        ),
        (
            "a registry that is a symbolic link at the base",
            vec![
                tree(M, vec![entry("tools", "040000", "tree", &a)]),
                tree(&a, vec![entry("repo_scans.json", "120000", "blob", &c)]),
            ],
            good_new(),
        ),
        (
            "a registry that is a directory",
            good_old(),
            head_entry("040000", "tree"),
        ),
        (
            "a registry that is a submodule",
            good_old(),
            head_entry("160000", "commit"),
        ),
        (
            "a regular file entry whose type is not a blob",
            good_old(),
            head_entry("100644", "tree"),
        ),
        (
            "a mode that is no regular file",
            good_old(),
            head_entry("100664", "blob"),
        ),
        (
            "a tools folder that is a symbolic link",
            good_old(),
            vec![tree(H, vec![entry("tools", "120000", "blob", &b)])],
        ),
        (
            "a tools folder that is a submodule",
            good_old(),
            vec![tree(H, vec![entry("tools", "160000", "commit", &b)])],
        ),
        (
            "no tools folder",
            good_old(),
            vec![tree(H, vec![entry("src", "040000", "tree", &b)])],
        ),
        (
            "no registry in the tools folder",
            good_old(),
            vec![
                tree(H, vec![entry("tools", "040000", "tree", &b)]),
                tree(&b, vec![]),
            ],
        ),
        (
            "a root tree that is truncated",
            good_old(),
            vec![truncated_root],
        ),
        (
            "a root tree that does not say whether it is truncated",
            good_old(),
            vec![no_truncated_key],
        ),
        (
            "an entry whose id is no object id",
            good_old(),
            bad_blob_id[..2].to_vec(),
        ),
        (
            "no tree at the base",
            vec![ex(format!("{API}/git/trees/{M}"), 404, not_found.clone())],
            good_new(),
        ),
        (
            "a blob answer for another object",
            good_old(),
            vec![
                good_new()[0].clone(),
                good_new()[1].clone(),
                blob(
                    &d,
                    json!({"sha": sha('e'), "encoding": "base64", "content": b64(&new)}),
                ),
            ],
        ),
        (
            "a blob answer that is a list",
            good_old(),
            vec![
                good_new()[0].clone(),
                good_new()[1].clone(),
                blob(&d, json!([{"type": "file"}])),
            ],
        ),
        (
            "base64 that does not decode",
            good_old(),
            vec![
                good_new()[0].clone(),
                good_new()[1].clone(),
                blob(
                    &d,
                    json!({"sha": d, "encoding": "base64", "content": "!!!!"}),
                ),
            ],
        ),
        (
            "a valid text whose encoding is not base64",
            good_old(),
            vec![
                good_new()[0].clone(),
                good_new()[1].clone(),
                blob(
                    &d,
                    json!({"sha": d, "encoding": "none", "content": b64(&new)}),
                ),
            ],
        ),
        (
            "base64 with a character outside the alphabet after a valid text",
            good_old(),
            vec![
                good_new()[0].clone(),
                good_new()[1].clone(),
                blob(
                    &d,
                    json!({"sha": d, "encoding": "base64", "content": format!("{}!", b64(&new))}),
                ),
            ],
        ),
    ];
    // The two sides before the merge base answer: no merge base, or one that is no commit id.
    let no_merge_base: Vec<(&str, Vec<Exchange>)> = vec![
        ("no merge base answer", vec![compare(404, not_found)]),
        (
            "a merge base that is no commit id",
            vec![compare(200, json!({"merge_base_commit": {"sha": "main"}}))],
        ),
    ];
    for (what, reads) in no_merge_base {
        let mut all = pr_exchanges(json!([{"filename": REG}]));
        all.extend(reads);
        let n = all.len();
        let t = RecordedTransport::new(all);
        let input = audit_gate::collect("o", "r", 9, None, &t).unwrap();
        assert_eq!(t.remaining(), 0, "{what}: the reads it makes");
        assert_eq!(t.sent.borrow().len(), n, "{what}: no other read");
        assert_eq!(gate_tier(&input), 2, "{what}: the registry stays tier 2");
    }
    for (what, old_reads, new_reads) in cases {
        let mut all = pr_exchanges(json!([{"filename": REG}]));
        all.push(merge.clone());
        all.extend(old_reads);
        all.extend(new_reads);
        let n = all.len();
        let t = RecordedTransport::new(all);
        let input = audit_gate::collect("o", "r", 9, None, &t).unwrap();
        assert_eq!(t.remaining(), 0, "{what}: the reads it makes");
        assert_eq!(t.sent.borrow().len(), n, "{what}: no other read");
        assert_eq!(gate_tier(&input), 2, "{what}: the registry stays tier 2");
    }
}

/// The audit's B1 with the answer in reach: the head's registry entry is a symbolic link (mode
/// 120000) whose blob, were it read, answers a real registration. The text must not be read at
/// all (the blob is never requested), so the head stays tier 2 and the recorded answer stays
/// unused.
#[test]
fn a_linked_registry_whose_blob_would_read_as_a_registration_is_tier_2() {
    let mut grown = base_scans();
    grown[0]["rules"] = json!(["tests/alpha.rs", "tests/beta.rs", "tests/gamma.rs"]);
    let (old, new) = (base(), registry(grown));
    let (a, b, c, d) = (sha('a'), sha('b'), sha('c'), sha('d'));
    let mut all = pr_exchanges(json!([{"filename": REG}]));
    all.push(compare(200, json!({"merge_base_commit": {"sha": M}})));
    all.extend(side(M, &a, &c, &old));
    all.extend([
        tree(H, vec![entry("tools", "040000", "tree", &b)]),
        tree(&b, vec![entry("repo_scans.json", "120000", "blob", &d)]),
        // Never requested: the blob of the link, which holds the text of a registration.
        blob_text(&d, &new),
    ]);
    let before_blob = all.len() - 1;
    let t = RecordedTransport::new(all);
    let input = audit_gate::collect("o", "r", 9, None, &t).unwrap();
    assert_eq!(
        t.remaining(),
        1,
        "the answer for the link's blob stays unused"
    );
    assert_eq!(
        t.sent.borrow().len(),
        before_blob,
        "no read of the link's blob"
    );
    assert!(
        t.sent
            .borrow()
            .iter()
            .all(|r| !r.url.contains(&format!("/git/blobs/{d}"))),
        "the link's blob was requested"
    );
    assert_eq!(
        input.registry,
        Some(RegistryTexts {
            old: Some(old),
            new: None
        })
    );
    assert_eq!(gate_tier(&input), 2, "a link whose blob is a registration");
}

#[test]
fn collect_reads_an_executable_registry_as_a_regular_file() {
    let mut grown = base_scans();
    grown[0]["rules"] = json!(["tests/alpha.rs", "tests/beta.rs", "tests/gamma.rs"]);
    let (old, new) = (base(), registry(grown));
    let mut all = pr_exchanges(json!([{"filename": REG}]));
    all.push(compare(200, json!({"merge_base_commit": {"sha": M}})));
    all.extend(side(M, &sha('a'), &sha('c'), &old));
    all.extend([
        tree(H, vec![entry("tools", "040000", "tree", &sha('b'))]),
        tree(
            &sha('b'),
            vec![entry("repo_scans.json", "100755", "blob", &sha('d'))],
        ),
        blob_text(&sha('d'), &new),
    ]);
    let t = RecordedTransport::new(all);
    let input = audit_gate::collect("o", "r", 9, None, &t).unwrap();
    assert_eq!(t.remaining(), 0);
    assert_eq!(
        gate_tier(&input),
        1,
        "a regular file with the executable bit"
    );
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
