//! SPEC-20261119-flaky-followups (card C-99, REQ-4 and REQ-5): `moai card list --title-prefix
//! <text>` lists only the cards whose title starts with that text, case-sensitively, in the order
//! and shape it uses without the option; an empty prefix or one with a control character is
//! refused before the ledger is read. Every command runs against a temporary HOME; the real
//! `~/.moai` is not read or written.

mod common;

use common::home::TempHome;
use moai_cli::ledger;
use serde_json::{json, Value};
use std::process::Output;

fn run(home: &TempHome, args: &[&str]) -> Output {
    home.command(args).output().expect("run moai")
}

fn envelope(out: &Output) -> Value {
    serde_json::from_slice(&out.stdout).unwrap_or_else(|e| {
        panic!(
            "stdout is not JSON ({e}): {} {}",
            String::from_utf8_lossy(&out.stdout),
            String::from_utf8_lossy(&out.stderr)
        )
    })
}

fn result(out: &Output) -> Value {
    assert_eq!(
        out.status.code(),
        Some(0),
        "{} {}",
        String::from_utf8_lossy(&out.stdout),
        String::from_utf8_lossy(&out.stderr)
    );
    let v = envelope(out);
    assert_eq!(v["ok"], json!(true), "{v}");
    v["result"].clone()
}

fn refusal(out: &Output) -> (String, String) {
    assert_eq!(
        out.status.code(),
        Some(1),
        "{} {}",
        String::from_utf8_lossy(&out.stdout),
        String::from_utf8_lossy(&out.stderr)
    );
    let v = envelope(out);
    assert_eq!(v["ok"], json!(false), "{v}");
    (
        v["error"]["kind"].as_str().unwrap().to_string(),
        v["error"]["message"].as_str().unwrap().to_string(),
    )
}

/// A HOME with a project `demo`, its ledger and four cards: two start with `flaky: `, one with
/// `Flaky: ` (capital F) and one holds the text only in the middle.
fn four_cards() -> (TempHome, String) {
    let home = TempHome::new();
    let project = home.project("demo");
    ledger::register(home.path(), "demo").unwrap();
    let project = project.to_str().unwrap().to_string();
    for title in [
        "flaky: ci_gate_times_out",
        "Flaky: capital letter",
        "flaky: second_one",
        "not flaky: text in the middle",
    ] {
        result(&run(
            &home,
            &["card", "add", "--title", title, "--project", &project],
        ));
    }
    (home, project)
}

fn names(r: &Value) -> Vec<String> {
    r["cards"]
        .as_array()
        .unwrap()
        .iter()
        .map(|c| c["name"].as_str().unwrap().to_string())
        .collect()
}

#[test]
fn the_prefix_keeps_only_cards_whose_title_starts_with_it_case_sensitively() {
    let (home, project) = four_cards();
    let r = result(&run(
        &home,
        &["card", "list", &project, "--title-prefix", "flaky: "],
    ));
    assert_eq!(names(&r), vec!["C-1", "C-3"], "{r}");
    assert_eq!(r["count"], json!(2));
    assert_eq!(r["writes"], json!(false));
    // Case-sensitive: the capital F is a different prefix.
    let r = result(&run(
        &home,
        &["card", "list", &project, "--title-prefix", "Flaky: "],
    ));
    assert_eq!(names(&r), vec!["C-2"], "{r}");
    // A prefix nothing starts with lists nothing.
    let r = result(&run(
        &home,
        &["card", "list", &project, "--title-prefix", "middle"],
    ));
    assert_eq!(r["count"], json!(0), "{r}");
}

#[test]
fn without_the_option_the_output_is_the_same_as_before_and_each_card_keeps_its_shape() {
    let (home, project) = four_cards();
    let all = result(&run(&home, &["card", "list", &project]));
    assert_eq!(names(&all), vec!["C-1", "C-2", "C-3", "C-4"], "{all}");
    let one = result(&run(
        &home,
        &["card", "list", &project, "--title-prefix", "flaky: "],
    ));
    // The filtered card is the unfiltered card, field for field, in the same order.
    assert_eq!(one["cards"][0], all["cards"][0]);
    assert_eq!(one["cards"][1], all["cards"][2]);
    let keys = |v: &Value| -> Vec<String> { v.as_object().unwrap().keys().cloned().collect() };
    assert_eq!(keys(&one), keys(&all), "the envelope keeps its shape");
}

#[test]
fn it_combines_with_state_and_both_filters_must_match() {
    let (home, project) = four_cards();
    let r = result(&run(
        &home,
        &[
            "card",
            "list",
            &project,
            "--title-prefix",
            "flaky: ",
            "--state",
            "draft",
        ],
    ));
    assert_eq!(names(&r), vec!["C-1", "C-3"], "{r}");
    let r = result(&run(
        &home,
        &[
            "card",
            "list",
            &project,
            "--title-prefix",
            "flaky: ",
            "--state",
            "done",
        ],
    ));
    assert_eq!(r["count"], json!(0), "{r}");
}

#[test]
fn an_empty_prefix_or_one_with_a_control_character_is_refused_before_the_ledger_is_read() {
    // No ledger at all: a valid prefix still answers "absent" (exit 0), so a refusal below can
    // only come from checking the prefix first.
    let home = TempHome::new();
    let project = home.project("demo");
    let project = project.to_str().unwrap().to_string();
    let absent = result(&run(
        &home,
        &["card", "list", &project, "--title-prefix", "flaky: "],
    ));
    assert_eq!(absent["ledger"], json!("absent"), "{absent}");
    for bad in ["", "a\nb", "a\tb", "a\u{7f}b"] {
        let (kind, message) = refusal(&run(
            &home,
            &["card", "list", &project, "--title-prefix", bad],
        ));
        assert_eq!(kind, "invalid_input", "{bad:?}: {message}");
        assert!(message.contains("title prefix"), "{bad:?}: {message}");
    }
}

/// The prefix is taken as typed (a trailing space is part of it, nothing is trimmed) and its length
/// counts characters, not bytes: Hangul titles match a Hangul prefix.
#[test]
fn the_prefix_is_taken_as_typed_and_counts_characters() {
    let home = TempHome::new();
    let project = home.project("demo");
    ledger::register(home.path(), "demo").unwrap();
    let project = project.to_str().unwrap().to_string();
    for title in ["flaky: one", "flaky:tight", "간헐: a", "간헐 b"] {
        result(&run(
            &home,
            &["card", "add", "--title", title, "--project", &project],
        ));
    }
    let with = |prefix: &str| {
        let r = result(&run(
            &home,
            &["card", "list", &project, "--title-prefix", prefix],
        ));
        names(&r)
    };
    // The trailing space is kept: `flaky:tight` does not start with `flaky: `.
    assert_eq!(with("flaky: "), vec!["C-1"]);
    assert_eq!(with("flaky:"), vec!["C-1", "C-2"]);
    // Characters, not bytes.
    assert_eq!(with("간헐: "), vec!["C-3"]);
    assert_eq!(with("간헐"), vec!["C-3", "C-4"]);
}
