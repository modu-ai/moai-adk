//! SPEC-20261123-take-scope-stored: `moai card take` keeps the scope it was given on the card,
//! fills an empty scope from the SPEC the card names or from the card's last reservation, names
//! the two ways out when it has no source, and `moai card next` lists what a take could fill.
//!
//! Every command runs against a temporary HOME; the real home state is not read or written. The
//! holder is the session id in the environment (`CLAUDE_CODE_SESSION_ID`).
#![cfg(unix)]

mod common;

use common::home::TempHome;
use moai_cli::ledger;
use serde_json::{json, Value};
use std::fs;
use std::path::PathBuf;
use std::process::{Output, Stdio};

const SESSION_VAR: &str = "CLAUDE_CODE_SESSION_ID";
const S_A: &str = "aaaaaaaa-0000-4000-8000-00000000000a";
const S_B: &str = "bbbbbbbb-0000-4000-8000-00000000000b";

struct World {
    home: TempHome,
    root: PathBuf,
    project: String,
}

fn world() -> World {
    let home = TempHome::new();
    let root = home.project("demo");
    ledger::register(home.path(), "demo").unwrap();
    World {
        project: root.to_str().unwrap().to_string(),
        home,
        root,
    }
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

/// The message of a refused command (exit 1, `ok: false`).
fn refusal(out: &Output) -> String {
    assert_eq!(
        out.status.code(),
        Some(1),
        "{} {}",
        String::from_utf8_lossy(&out.stdout),
        String::from_utf8_lossy(&out.stderr)
    );
    let v = envelope(out);
    assert_eq!(v["ok"], json!(false), "{v}");
    v["error"]["message"].as_str().unwrap().to_string()
}

impl World {
    /// `moai card <args> --project <project>`, as `session`.
    fn card(&self, session: Option<&str>, args: &[&str]) -> Output {
        let mut all = vec!["card"];
        all.extend_from_slice(args);
        all.extend_from_slice(&["--project", &self.project]);
        self.run(session, &all)
    }

    /// `moai card next <project>` (the project is a positional argument there).
    fn next(&self, session: Option<&str>) -> Value {
        result(&self.run(session, &["card", "next", &self.project]))
    }

    fn run(&self, session: Option<&str>, args: &[&str]) -> Output {
        let mut cmd = self.home.command(args);
        if let Some(id) = session {
            cmd.env(SESSION_VAR, id);
        }
        cmd.stdin(Stdio::null()).output().expect("run moai")
    }

    /// Runs `statement` against the ledger directly.
    fn sql(&self, statement: &str) {
        let mut l = ledger::open_write(self.home.path(), "demo").unwrap();
        l.with_write_tx(|tx, _| {
            tx.execute_batch(statement)?;
            Ok(())
        })
        .unwrap();
    }

    fn count(&self, sql: &str) -> i64 {
        let l = ledger::open_read_only(self.home.path(), "demo").unwrap();
        l.conn().query_row(sql, [], |r| r.get(0)).unwrap()
    }

    /// A `ready` card without a write scope (`--scope` is left out of `card add`).
    fn empty_card(&self, title: &str) -> String {
        let added = result(&self.card(None, &["add", "--title", title, "--ac", "fixture"]));
        let name = added["name"].as_str().unwrap().to_string();
        result(&self.card(None, &["ready", &name]));
        name
    }

    /// A SPEC in the project with the `write:` paths given, light form; returns its id.
    fn spec(&self, slug: &str, paths: &[String]) -> String {
        let id = format!("SPEC-20261123-{slug}");
        let list = paths
            .iter()
            .map(|p| format!("`{p}`"))
            .collect::<Vec<_>>()
            .join(", ");
        let body = format!(
            "---\nid: {id}\ntitle: \"a demo\"\nstatus: draft\nrevision: 1\ncreated: 2026-11-23\n---\n\n\
             # a demo\n\n## Intent\n\nA demo.\n\n## Requirements\n\n- REQ-1: The tool shall do a thing.\n\n\
             ## Acceptance\n\n- AC-1 (REQ-1): it does the thing.\n  verify: `true`\n\n\
             ## Scope\n\n- write: {list}\n- read-only: every other file\n\n## Plan\n\n1. Do it.\n"
        );
        let dir = self.root.join(moai_cli::spec_doc::SPECS_DIR).join(&id);
        fs::create_dir_all(&dir).unwrap();
        fs::write(dir.join("spec.md"), body).unwrap();
        id
    }

    /// A SPEC whose `## Scope` section is the text given, verbatim; returns its id.
    fn spec_with_scope(&self, slug: &str, scope: &str) -> String {
        let id = format!("SPEC-20261123-{slug}");
        let body = format!(
            "---\nid: {id}\ntitle: \"a demo\"\nstatus: draft\nrevision: 1\ncreated: 2026-11-23\n---\n\n\
             # a demo\n\n## Intent\n\nA demo.\n\n## Requirements\n\n- REQ-1: The tool shall do a thing.\n\n\
             ## Acceptance\n\n- AC-1 (REQ-1): it does the thing.\n  verify: `true`\n\n\
             ## Scope\n\n{scope}\n\n## Plan\n\n1. Do it.\n"
        );
        let dir = self.root.join(moai_cli::spec_doc::SPECS_DIR).join(&id);
        fs::create_dir_all(&dir).unwrap();
        fs::write(dir.join("spec.md"), body).unwrap();
        id
    }

    fn link(&self, card: &str, spec: &str) {
        result(&self.card(None, &["spec", card, spec]));
    }

    fn show(&self, card: &str) -> Value {
        result(&self.card(None, &["show", card]))
    }

    /// The card's own scope as the `card show` answer reads it.
    fn own_scope(&self, card: &str) -> Value {
        self.show(card)["write_scope"].clone()
    }
}

fn paths(n: usize) -> Vec<String> {
    (0..n).map(|i| format!("src/part_{i:03}.rs")).collect()
}

/// AC-1 (REQ-1, REQ-5): the scope given to a take is on the card after the release, and the card
/// is on the list `card next` makes.
#[test]
fn a_given_scope_is_kept_on_the_card() {
    let w = world();
    let name = w.empty_card("given");
    let took = result(&w.card(
        Some(S_A),
        &["take", &name, "--scope", "src/a.rs", "--scope", "docs/"],
    ));
    assert_eq!(took["scope_source"], json!("given"), "{took}");
    assert_eq!(
        took["reservation"]["write_scope"],
        json!(["docs/", "src/a.rs"]),
        "{took}"
    );
    result(&w.card(Some(S_A), &["release", &name, "--generation", "1"]));
    assert_eq!(w.own_scope(&name), json!(["docs/", "src/a.rs"]));
    let next = w.next(Some(S_B));
    let ready: Vec<&str> = next["ready"]
        .as_array()
        .unwrap()
        .iter()
        .map(|c| c["name"].as_str().unwrap())
        .collect();
    assert_eq!(ready, [name.as_str()], "{next}");
}

/// AC-2 (REQ-2): a card that has its own scope keeps it; only the reservation takes the given one.
#[test]
fn an_own_scope_is_never_changed_by_take() {
    let w = world();
    let added = result(&w.card(
        None,
        &[
            "add", "--title", "own", "--scope", "src/a.rs", "--ac", "fixture",
        ],
    ));
    let name = added["name"].as_str().unwrap().to_string();
    result(&w.card(None, &["ready", &name]));
    let took = result(&w.card(Some(S_A), &["take", &name, "--scope", "src/b.rs"]));
    assert_eq!(took["reservation"]["write_scope"], json!(["src/b.rs"]));
    assert_eq!(took["scope_source"], json!("given"), "{took}");
    assert_eq!(w.own_scope(&name), json!(["src/a.rs"]));
}

/// AC-3 (REQ-3, REQ-5): an empty card that names a SPEC with a `write:` list is filled from it.
#[test]
fn an_empty_card_is_filled_from_its_spec() {
    let w = world();
    let name = w.empty_card("spec");
    let spec = w.spec("fill-demo", &["src/x.rs".into(), "tests/x.rs".into()]);
    w.link(&name, &spec);
    let took = result(&w.card(Some(S_A), &["take", &name]));
    assert_eq!(took["scope_source"], json!("spec"), "{took}");
    assert_eq!(
        took["reservation"]["write_scope"],
        json!(["src/x.rs", "tests/x.rs"]),
        "{took}"
    );
    assert_eq!(w.own_scope(&name), json!(["src/x.rs", "tests/x.rs"]));
    assert_eq!(
        w.count("SELECT count(*) FROM event WHERE kind = 'card.scope'"),
        1
    );
}

/// AC-4 (REQ-4, REQ-5): an empty card with no SPEC list gets the scope of its last reservation.
#[test]
fn an_empty_card_is_filled_from_its_last_reservation() {
    let w = world();
    let name = w.empty_card("again");
    result(&w.card(Some(S_A), &["take", &name, "--scope", "src/old.rs"]));
    result(&w.card(Some(S_A), &["release", &name, "--generation", "1"]));
    // A card released before this change has no scope of its own: the shape of the ready cards
    // that `card next` skips as `empty_scope`.
    w.sql("UPDATE card SET write_scope = '[]' WHERE card_id = 1");
    assert_eq!(w.own_scope(&name), json!([]));
    let took = result(&w.card(Some(S_B), &["take", &name]));
    assert_eq!(took["scope_source"], json!("last_reservation"), "{took}");
    assert_eq!(took["reservation"]["write_scope"], json!(["src/old.rs"]));
    assert_eq!(w.own_scope(&name), json!(["src/old.rs"]));
}

/// AC-5 (REQ-6): with no source the refusal names both ways out and nothing is written.
#[test]
fn no_source_names_the_two_ways_out() {
    let w = world();
    let name = w.empty_card("nothing");
    let events = w.count("SELECT count(*) FROM event");
    let message = refusal(&w.card(Some(S_A), &["take", &name]));
    assert!(message.contains("--scope"), "{message}");
    assert!(message.contains("card scope"), "{message}");
    assert_eq!(w.count("SELECT count(*) FROM event"), events);
    assert_eq!(w.count("SELECT count(*) FROM reservation"), 0);
    assert_eq!(w.own_scope(&name), json!([]));
}

/// AC-6 (REQ-7): the fill meets the overlap check in the same transaction; a refusal leaves the
/// card's scope empty and writes no `card.scope` event.
#[test]
fn a_fill_keeps_the_overlap_check() {
    let w = world();
    let holder = w.empty_card("holder");
    result(&w.card(Some(S_A), &["take", &holder, "--scope", "src/x.rs"]));
    let name = w.empty_card("filled");
    let spec = w.spec("fill-overlap", &["src/x.rs".into()]);
    w.link(&name, &spec);
    let events = w.count("SELECT count(*) FROM event WHERE kind = 'card.scope'");
    let message = refusal(&w.card(Some(S_B), &["take", &name]));
    // The reservation's own refusal names the write scope in conflict and the card that holds it.
    assert!(message.contains("conflict in write scope"), "{message}");
    assert!(message.contains(&holder), "{message}");
    assert_eq!(w.own_scope(&name), json!([]));
    assert_eq!(
        w.count("SELECT count(*) FROM event WHERE kind = 'card.scope'"),
        events
    );
}

/// AC-7 (REQ-8): a SPEC list of 64 paths fills the scope, a list of 65 is refused with the limit.
#[test]
fn the_spec_list_has_the_same_limit() {
    let w = world();
    let fits = w.empty_card("sixty four");
    let spec = w.spec("fill-sixty-four", &paths(64));
    w.link(&fits, &spec);
    let took = result(&w.card(Some(S_A), &["take", &fits]));
    assert_eq!(took["scope_source"], json!("spec"), "{took}");
    assert_eq!(
        took["reservation"]["write_scope"].as_array().unwrap().len(),
        64
    );

    let too_many = w.empty_card("sixty five");
    let spec = w.spec("fill-sixty-five", &paths(65));
    w.link(&too_many, &spec);
    let reservations = w.count("SELECT count(*) FROM reservation");
    let message = refusal(&w.card(Some(S_A), &["take", &too_many]));
    assert!(message.contains("64"), "{message}");
    assert_eq!(w.count("SELECT count(*) FROM reservation"), reservations);
    assert_eq!(w.own_scope(&too_many), json!([]));
}

/// AC-11 (REQ-8, REQ-9): `card next` offers no fill that a take would refuse; a SPEC list of 65
/// paths leaves the card skipped as `empty_scope`.
#[test]
fn next_skips_a_fill_a_take_would_refuse() {
    let w = world();
    let name = w.empty_card("sixty five");
    let spec = w.spec("fill-next-limit", &paths(65));
    w.link(&name, &spec);
    let next = w.next(Some(S_B));
    assert!(next["ready"].as_array().unwrap().is_empty(), "{next}");
    let skipped = next["skipped"].as_array().unwrap();
    let card = skipped
        .iter()
        .find(|c| c["name"] == json!(name))
        .unwrap_or_else(|| panic!("no {name} in {next}"));
    assert_eq!(card["reason"], json!("empty_scope"), "{card}");
}

/// AC-8 (REQ-9): `card next` lists the cards a take could fill, with the scope and its source,
/// and still skips a card with no source as `empty_scope`.
#[test]
fn next_lists_what_take_can_fill() {
    let w = world();
    let from_spec = w.empty_card("from spec");
    let spec = w.spec("fill-next", &["src/n.rs".into()]);
    w.link(&from_spec, &spec);
    let from_reservation = w.empty_card("from reservation");
    result(&w.card(
        Some(S_A),
        &["take", &from_reservation, "--scope", "src/r.rs"],
    ));
    result(&w.card(
        Some(S_A),
        &["release", &from_reservation, "--generation", "1"],
    ));
    w.sql("UPDATE card SET write_scope = '[]' WHERE title = 'from reservation'");
    let nothing = w.empty_card("nothing");

    let next = w.next(Some(S_B));
    let ready = next["ready"].as_array().unwrap();
    let find = |name: &str| ready.iter().find(|c| c["name"] == json!(name)).cloned();
    let a = find(&from_spec).unwrap_or_else(|| panic!("no {from_spec} in {next}"));
    assert_eq!(a["write_scope"], json!(["src/n.rs"]), "{a}");
    assert_eq!(a["scope_source"], json!("spec"), "{a}");
    let b = find(&from_reservation).unwrap_or_else(|| panic!("no {from_reservation} in {next}"));
    assert_eq!(b["write_scope"], json!(["src/r.rs"]), "{b}");
    assert_eq!(b["scope_source"], json!("last_reservation"), "{b}");
    let skipped = next["skipped"].as_array().unwrap();
    let empty = skipped
        .iter()
        .find(|c| c["name"] == json!(nothing))
        .unwrap_or_else(|| panic!("no {nothing} in {next}"));
    assert_eq!(empty["reason"], json!("empty_scope"), "{empty}");
    // `next` only reads: the cards still have no scope of their own.
    assert_eq!(w.own_scope(&from_spec), json!([]));
}

// Scope sections copied from real SPECs of this repository (the kinds a fill met on the real
// SPEC set, audit of the first head): a `write:` list whose spans are not paths, one that wraps
// into prose, one that repeats a path. Only the lines of `## Scope` that matter are kept.

/// `SPEC-20261119-bundle-build`: backtick spans that are code, not paths (`pub mod build;`).
const SCOPE_CODE_SPAN: &str = "- write: `src/bundle.rs` (the line `pub mod build;`, the merge-driver list, and what `build` takes from the module), `src/bundle/build.rs`, `src/ops/bundle.rs`, `tests/bundle_build_cli.rs`.\n- read-only: `src/home_workdir.rs`, `src/worktree.rs`.";

/// `SPEC-20261121-remove-codex-project-layer`: a `write:` line that is prose, no backticks (the
/// words `the`, `and`, `listed` were read as paths).
const SCOPE_PROSE_LINE: &str = "- write: `docs/remove-codex-project-layer/**`\n- write: `src/project_hosts.rs`, `src/project_update.rs`\n- write: the tests and fixtures that call an API this slice removes, and the Codex-only test files of the project layer and the hook adapters, listed in the impact report (sections 2.3 and 2.5)\n- read-only: every other file";

/// `SPEC-20261119-card-scope-widen`: an option (`--reason`) in the middle of a path's note.
const SCOPE_OPTION_SPAN: &str = "- write: `src/card/dispatch.rs`, `src/ledger/reservation.rs`, `src/ops/card.rs` (the new `--reason` option and the help only), `tests/card_scope_widen_cli.rs`\n- read-only: `src/ledger/schema.rs`";

/// `SPEC-20261119-flaky-policy`: a command with its argument inside one span (a `/` and a space).
const SCOPE_SPACED_SPAN: &str =
    "- write: `src/flaky.rs`, `tools/source_manifest.py write`, `tests/flaky_cli.rs`";

/// `SPEC-20261119-headver-release-workflow`: a call with a dot (`WorkflowTest.workflows()`): no
/// `/`, and what follows the dot is no extension.
const SCOPE_CALL_SPAN: &str =
    "- write: `.github/workflows/release.yml`, `WorkflowTest.workflows()`, `tests/headver_cli.rs`";

/// `SPEC-20261122-worktree-place`: an untagged list whose items name the same file first.
const SCOPE_REPEATED_PATH: &str = "- The path keystone (the functions the others call): `src/home_workdir.rs` (worktree_dir, worktrees_root), `src/lib.rs` (the WORKTREES_DIR constant).\n- The local exclude helper (one function every creator calls): `src/home_workdir.rs`; it reads `git rev-parse --git-common-dir`.\n- Creators of new worktrees: `src/worktree.rs` (the worktree command), `src/claude.rs` (the launcher).\n- Checkers and guards: `src/hooks.rs` (the hook path check).";

/// An option written with a path (`--out/dir.rs`): it has a `/` and an extension, so only its
/// leading `-` makes it no path (audit B4: removing that rule left every other fixture green).
const SCOPE_DASH_PATH: &str = "- write: `src/dash.rs`, `--out/dir.rs`, `tests/dash_cli.rs`";

/// `SPEC-20261123-audit-slot`: a clean list that wraps paths and notes the way real SPECs do.
const SCOPE_CLEAN: &str = "- write:\n  - `docs/audit_slot.md`\n  - `src/audit_slot.rs` (new)\n  - `src/lib.rs` (one `pub mod` line)\n  - `tests/audit_slot_cli.rs` (new)\n- read-only: everything else";

/// B1 (REQ-10): a SPEC list with an entry that is not a path shape fills nothing: the take is
/// refused with the SPEC file and both ways out, the card stays empty, `card next` skips it, and
/// nothing is reserved or written.
#[test]
fn a_spec_list_with_a_word_is_refused_not_stored() {
    let kinds: [(&str, &str, &str); 6] = [
        ("code-span", SCOPE_CODE_SPAN, "pub mod build;"),
        ("prose-line", SCOPE_PROSE_LINE, "the"),
        ("option-span", SCOPE_OPTION_SPAN, "--reason"),
        (
            "spaced-span",
            SCOPE_SPACED_SPAN,
            "tools/source_manifest.py write",
        ),
        ("call-span", SCOPE_CALL_SPAN, "WorkflowTest.workflows()"),
        ("dash-path", SCOPE_DASH_PATH, "--out/dir.rs"),
    ];
    for (slug, scope, word) in kinds {
        let w = world();
        let name = w.empty_card(slug);
        let spec = w.spec_with_scope(slug, scope);
        w.link(&name, &spec);
        let events = w.count("SELECT count(*) FROM event");
        let message = refusal(&w.card(Some(S_A), &["take", &name]));
        let spec_file = format!("{}/{spec}/spec.md", moai_cli::spec_doc::SPECS_DIR);
        assert!(message.contains(&spec_file), "{slug}: {message}");
        assert!(message.contains(&format!("`{word}`")), "{slug}: {message}");
        assert!(
            message.contains(&format!("card take {name} --scope <path>")),
            "{slug}: {message}"
        );
        assert!(
            message.contains(&format!("card scope {name} --scope <path>")),
            "{slug}: {message}"
        );
        assert_eq!(w.count("SELECT count(*) FROM event"), events, "{slug}");
        assert_eq!(w.count("SELECT count(*) FROM reservation"), 0, "{slug}");
        assert_eq!(w.own_scope(&name), json!([]), "{slug}");
        let next = w.next(Some(S_B));
        assert!(
            next["ready"].as_array().unwrap().is_empty(),
            "{slug}: {next}"
        );
        let skipped = next["skipped"].as_array().unwrap();
        let card = skipped.iter().find(|c| c["name"] == json!(name));
        assert_eq!(
            card.map(|c| c["reason"].clone()),
            Some(json!("empty_scope")),
            "{slug}: {next}"
        );
    }
}

/// B1 (REQ-10): a real list of paths with notes and wrapped list items still fills.
#[test]
fn a_clean_real_spec_list_still_fills() {
    let w = world();
    let name = w.empty_card("clean");
    let spec = w.spec_with_scope("clean", SCOPE_CLEAN);
    w.link(&name, &spec);
    let took = result(&w.card(Some(S_A), &["take", &name]));
    assert_eq!(took["scope_source"], json!("spec"), "{took}");
    assert_eq!(
        took["reservation"]["write_scope"],
        json!([
            "docs/audit_slot.md",
            "src/audit_slot.rs",
            "src/lib.rs",
            "tests/audit_slot_cli.rs"
        ]),
        "{took}"
    );
}

/// B1 (REQ-11): a path the list names twice is reserved once; the answer says what was merged.
#[test]
fn a_spec_list_that_repeats_a_path_is_merged() {
    let w = world();
    let name = w.empty_card("repeat");
    let spec = w.spec_with_scope("repeat", SCOPE_REPEATED_PATH);
    w.link(&name, &spec);
    let took = result(&w.card(Some(S_A), &["take", &name]));
    assert_eq!(took["scope_source"], json!("spec"), "{took}");
    assert_eq!(
        took["reservation"]["write_scope"],
        // An untagged list item gives its first path only: four items, three different paths.
        json!(["src/home_workdir.rs", "src/hooks.rs", "src/worktree.rs"]),
        "{took}"
    );
    let warnings = took["warnings"].to_string();
    assert!(warnings.contains("src/home_workdir.rs"), "{took}");
    assert!(warnings.contains("merged"), "{took}");
    assert_eq!(w.own_scope(&name), took["reservation"]["write_scope"]);
}

/// B2 (REQ-4): the fill takes the newest reservation, not the oldest.
#[test]
fn the_last_reservation_is_the_newest_one() {
    let w = world();
    let name = w.empty_card("two takes");
    result(&w.card(Some(S_A), &["take", &name, "--scope", "src/first.rs"]));
    result(&w.card(Some(S_A), &["release", &name, "--generation", "1"]));
    // The card keeps `src/first.rs`; the second reservation takes a different scope for it.
    result(&w.card(Some(S_A), &["take", &name, "--scope", "src/second.rs"]));
    result(&w.card(Some(S_A), &["release", &name, "--generation", "2"]));
    w.sql("UPDATE card SET write_scope = '[]' WHERE card_id = 1");
    let took = result(&w.card(Some(S_B), &["take", &name]));
    assert_eq!(took["scope_source"], json!("last_reservation"), "{took}");
    assert_eq!(took["reservation"]["write_scope"], json!(["src/second.rs"]));
    assert_eq!(w.own_scope(&name), json!(["src/second.rs"]));
}

/// B3 (REQ-9, REQ-7): `card next` runs the overlap check on the scope it fills in, so a card
/// whose fill meets another card's active reservation is skipped as `scope_overlap`.
#[test]
fn next_skips_a_fill_that_overlaps_a_held_scope() {
    let w = world();
    let holder = w.empty_card("holder");
    result(&w.card(Some(S_A), &["take", &holder, "--scope", "src/x.rs"]));
    let name = w.empty_card("filled");
    let spec = w.spec("fill-next-overlap", &["src/x.rs".into()]);
    w.link(&name, &spec);
    let next = w.next(Some(S_B));
    let ready = next["ready"].as_array().unwrap();
    assert!(
        ready.iter().all(|c| c["name"] != json!(name)),
        "{name} is offered although its fill overlaps {holder}: {next}"
    );
    let skipped = next["skipped"].as_array().unwrap();
    let card = skipped
        .iter()
        .find(|c| c["name"] == json!(name))
        .unwrap_or_else(|| panic!("no {name} in {next}"));
    assert_eq!(card["reason"], json!("scope_overlap"), "{card}");
}

/// N2 (REQ-12): a comma inside a `--scope` value is refused, for a take and for `card scope`; one
/// `--scope` names one path.
#[test]
fn a_comma_in_a_scope_value_is_refused() {
    let w = world();
    let name = w.empty_card("commas");
    let message = refusal(&w.card(Some(S_A), &["take", &name, "--scope", "src/a.rs,src/b.rs"]));
    assert!(message.contains("comma"), "{message}");
    assert!(message.contains("one `--scope` per path"), "{message}");
    assert_eq!(w.count("SELECT count(*) FROM reservation"), 0);
    let message = refusal(&w.card(
        None,
        &[
            "scope",
            &name,
            "--scope",
            "src/a.rs,src/b.rs",
            "--by",
            "leader",
        ],
    ));
    assert!(message.contains("comma"), "{message}");
    assert_eq!(w.own_scope(&name), json!([]));
}

/// N6 (REQ-12): the comma refusal names the option the caller typed: `--scope` for take, scope and
/// add, `--path` for give-back and line-lease, `--keep` for review.
#[test]
fn the_comma_refusal_names_the_option_of_the_caller() {
    let w = world();
    let name = w.empty_card("option names");
    result(&w.card(Some(S_A), &["take", &name, "--scope", "src/a.rs"]));
    let cases: [(&[&str], &str); 3] = [
        (
            &[
                "give-back",
                &name,
                "--generation",
                "1",
                "--path",
                "src/a.rs,src/b.rs",
            ],
            "one `--path` per path",
        ),
        (
            &[
                "line-lease",
                &name,
                "--generation",
                "1",
                "--path",
                "src/a.rs,src/b.rs",
                "--",
                "true",
            ],
            "one `--path` per path",
        ),
        (
            &[
                "review",
                &name,
                "--generation",
                "1",
                "--keep",
                "src/a.rs,src/b.rs",
            ],
            "one `--keep` per path",
        ),
    ];
    for (args, expected) in cases {
        let message = refusal(&w.card(Some(S_A), args));
        assert!(message.contains("comma"), "{args:?}: {message}");
        assert!(message.contains(expected), "{args:?}: {message}");
        assert!(!message.contains("`--scope`"), "{args:?}: {message}");
    }
}

// ---- shared files (SPEC-20261124-shared-files-profile) ------------------------------------------

impl World {
    /// Writes the profile with a `shared_files` list (`None` leaves the key out; the test project's
    /// profile is the minimal one of `TempHome::project`).
    fn shared(&self, list: Option<Value>) {
        let mut profile = json!({"schema_version": 2, "project_id": "demo"});
        if let Some(list) = list {
            profile["shared_files"] = list;
        }
        fs::write(
            self.root
                .join(moai_cli::PROJECT_STATE_DIR)
                .join("project.yaml"),
            format!("{profile}\n"),
        )
        .unwrap();
    }

    /// A `ready` card with the scope given (`card add --scope`, so no take has checked it yet).
    fn scoped_card(&self, title: &str, scope: &[&str]) -> String {
        let mut args = vec!["add", "--title", title, "--ac", "fixture"];
        for entry in scope {
            args.push("--scope");
            args.push(entry);
        }
        let name = result(&self.card(None, &args))["name"]
            .as_str()
            .unwrap()
            .to_string();
        result(&self.card(None, &["ready", &name]));
        name
    }
}

/// AC-2 (REQ-3): a given `--scope` and a card's own scope that overlap a shared file are refused
/// with the entry, the shared entry and the line-lease form; nothing is written.
#[test]
fn shared_files_refuse_a_given_scope() {
    let w = world();
    w.shared(Some(json!(["src/lib.rs", "tests/fixtures/help/*.txt"])));
    let name = w.empty_card("shared given");
    for (entry, shared) in [
        ("src/lib.rs", "src/lib.rs"),
        ("src/**", "src/lib.rs"),
        ("tests/fixtures/help/a.txt", "tests/fixtures/help/*.txt"),
    ] {
        let message = refusal(&w.card(Some(S_A), &["take", &name, "--scope", entry]));
        assert!(message.contains(&format!("`{entry}`")), "{message}");
        assert!(message.contains(&format!("`{shared}`")), "{message}");
        assert!(message.contains("card line-lease"), "{message}");
    }
    // A card whose own scope holds a shared file is refused without any `--scope`.
    let own = w.scoped_card("shared own", &["src/lib.rs"]);
    let message = refusal(&w.card(Some(S_A), &["take", &own]));
    assert!(message.contains("src/lib.rs"), "{message}");
    assert!(message.contains("card line-lease"), "{message}");
    assert_eq!(w.count("SELECT count(*) FROM reservation"), 0);
    assert_eq!(w.own_scope(&name), json!([]));
    // A path that overlaps no shared entry is taken.
    let took = result(&w.card(Some(S_A), &["take", &name, "--scope", "src/other.rs"]));
    assert_eq!(took["scope_source"], json!("given"), "{took}");
}

/// AC-3 (REQ-4): a fill from a SPEC list or from the last reservation leaves the shared entries
/// out and says so in `warnings`; a list that holds only shared entries refuses.
#[test]
fn shared_files_are_left_out_of_a_fill() {
    let w = world();
    w.shared(Some(json!(["src/lib.rs"])));
    let from_spec = w.empty_card("shared spec");
    let spec = w.spec("shared-fill", &["src/other.rs".into(), "src/lib.rs".into()]);
    w.link(&from_spec, &spec);
    let took = result(&w.card(Some(S_A), &["take", &from_spec]));
    assert_eq!(took["scope_source"], json!("spec"), "{took}");
    assert_eq!(
        took["reservation"]["write_scope"],
        json!(["src/other.rs"]),
        "{took}"
    );
    let warnings = took["warnings"].to_string();
    assert!(warnings.contains("src/lib.rs"), "{warnings}");
    assert!(warnings.contains("card line-lease"), "{warnings}");
    assert_eq!(w.own_scope(&from_spec), json!(["src/other.rs"]));

    // The last reservation of a card that has no scope of its own.
    w.shared(None);
    let again = w.empty_card("shared last");
    result(&w.card(
        Some(S_B),
        &[
            "take",
            &again,
            "--scope",
            "src/a.rs",
            "--scope",
            "src/lib.rs",
        ],
    ));
    result(&w.card(Some(S_B), &["release", &again, "--generation", "1"]));
    w.sql("UPDATE card SET write_scope = '[]' WHERE title = 'shared last'");
    w.shared(Some(json!(["src/lib.rs"])));
    let took = result(&w.card(Some(S_A), &["take", &again]));
    assert_eq!(took["scope_source"], json!("last_reservation"), "{took}");
    assert_eq!(
        took["reservation"]["write_scope"],
        json!(["src/a.rs"]),
        "{took}"
    );
    assert!(
        took["warnings"].to_string().contains("src/lib.rs"),
        "{took}"
    );

    // A list that names only shared files cannot fill anything.
    let only = w.empty_card("shared only");
    let spec = w.spec("shared-only", &["src/lib.rs".into()]);
    w.link(&only, &spec);
    let reservations = w.count("SELECT count(*) FROM reservation");
    let message = refusal(&w.card(Some(S_A), &["take", &only]));
    assert!(message.contains("card line-lease"), "{message}");
    assert_eq!(w.count("SELECT count(*) FROM reservation"), reservations);
    assert_eq!(w.own_scope(&only), json!([]));
}

/// AC-4 (REQ-5): `card scope` refuses a shared entry it would add, for a card nobody holds and for
/// a held card; an entry the card's scope already holds passes.
#[test]
fn shared_files_refuse_a_scope_change() {
    let w = world();
    w.shared(Some(json!(["src/lib.rs", "src/help.rs"])));
    let free = w.empty_card("shared free");
    let message = refusal(&w.card(
        None,
        &["scope", &free, "--scope", "src/help.rs", "--by", "leader"],
    ));
    assert!(message.contains("src/help.rs"), "{message}");
    assert!(message.contains("card line-lease"), "{message}");
    assert_eq!(w.own_scope(&free), json!([]));

    // A held card: widening with a shared file is refused, with another path it works.
    let held = w.empty_card("shared held");
    result(&w.card(Some(S_A), &["take", &held, "--scope", "src/a.rs"]));
    let widen = |card: &str, entries: &[&str]| {
        let mut args = vec!["scope", card];
        for e in entries {
            args.push("--scope");
            args.push(e);
        }
        args.extend_from_slice(&["--by", "leader", "--reason", "fixture", "--generation", "1"]);
        w.card(None, &args)
    };
    let message = refusal(&widen(&held, &["src/a.rs", "src/help.rs"]));
    assert!(message.contains("card line-lease"), "{message}");
    let widened = result(&widen(&held, &["src/a.rs", "src/other.rs"]));
    assert_eq!(
        widened["write_scope"],
        json!(["src/a.rs", "src/other.rs"]),
        "{widened}"
    );

    // A scope that held the shared file before the list named it is still widened by other paths.
    w.shared(None);
    let old = w.empty_card("shared old");
    result(&w.card(Some(S_B), &["take", &old, "--scope", "src/lib.rs"]));
    w.shared(Some(json!(["src/lib.rs"])));
    let widened = result(&widen(&old, &["src/lib.rs", "src/more.rs"]));
    assert_eq!(
        widened["write_scope"],
        json!(["src/lib.rs", "src/more.rs"]),
        "{widened}"
    );
}

/// AC-5 (REQ-6): `card check` of a shared file the reservation does not hold ends its refusal with
/// the line-lease form for that path; a path that is not shared keeps the old refusal.
#[test]
fn shared_files_check_names_line_lease() {
    let w = world();
    w.shared(Some(json!(["src/lib.rs"])));
    let name = w.empty_card("shared check");
    result(&w.card(Some(S_A), &["take", &name, "--scope", "src/a.rs"]));
    let message = refusal(&w.card(
        Some(S_A),
        &["check", &name, "--generation", "1", "--path", "src/lib.rs"],
    ));
    assert!(
        message.starts_with("reservation conflict in write scope"),
        "{message}"
    );
    assert!(
        message.ends_with("card line-lease C-1 --generation 1 --path src/lib.rs"),
        "{message}"
    );
    let message = refusal(&w.card(
        Some(S_A),
        &[
            "check",
            &name,
            "--generation",
            "1",
            "--path",
            "src/other.rs",
        ],
    ));
    assert!(!message.contains("line-lease"), "{message}");
}

/// AC-7 (REQ-8): without the key, and with a list of the wrong type, take, scope and check answer
/// as before this SPEC.
#[test]
fn shared_files_absent_changes_nothing() {
    for list in [None, Some(json!("src/lib.rs"))] {
        let w = world();
        w.shared(list);
        let name = w.empty_card("shared absent");
        let took = result(&w.card(Some(S_A), &["take", &name, "--scope", "src/lib.rs"]));
        assert_eq!(
            took["reservation"]["write_scope"],
            json!(["src/lib.rs"]),
            "{took}"
        );
        let widened = result(&w.card(
            None,
            &[
                "scope",
                &name,
                "--scope",
                "src/lib.rs",
                "--scope",
                "src/help.rs",
                "--by",
                "leader",
                "--reason",
                "fixture",
                "--generation",
                "1",
            ],
        ));
        assert_eq!(
            widened["write_scope"],
            json!(["src/help.rs", "src/lib.rs"]),
            "{widened}"
        );
        let message = refusal(&w.card(
            Some(S_A),
            &[
                "check",
                &name,
                "--generation",
                "1",
                "--path",
                "src/other.rs",
            ],
        ));
        assert!(!message.contains("line-lease"), "{message}");
    }
}
