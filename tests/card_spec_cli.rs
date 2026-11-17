//! SDD-4 (SPEC-20261117-card-spec-link): a card names the SPEC it works on. `moai card spec`
//! links and unlinks, `card show`/`list` print the link, `card next`/`take` print the linked
//! SPEC's status and warn when a card that changes the repository has no approved SPEC,
//! `take --require-approved-spec` refuses instead, and `moai spec show` lists the cards of this
//! machine's ledger that name the SPEC. The link lives only in the ledger, never in the SPEC.
//! Every command runs against a temporary HOME and a temporary project; the real `~/.moai` is
//! not read or written.

mod common;

use common::home::TempHome;
use moai_cli::ledger;
use serde_json::{json, Value};
use std::fs;
use std::path::PathBuf;
use std::process::Output;

const SESSION_VAR: &str = "CLAUDE_CODE_SESSION_ID";
const S_A: &str = "aaaaaaaa-0000-4000-8000-00000000000a";
const OTHER: &str = "SPEC-20261117-other-fix";

fn envelope(out: &Output) -> Value {
    serde_json::from_slice(&out.stdout).unwrap_or_else(|e| {
        panic!(
            "stdout is not JSON ({e}): {} {}",
            String::from_utf8_lossy(&out.stdout),
            String::from_utf8_lossy(&out.stderr)
        )
    })
}

fn ok(out: &Output) -> Value {
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

/// The refusal: exit 1, its `kind` and its message.
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

/// A HOME with a project `demo`, its registered ledger and the project folder.
struct World {
    home: TempHome,
    project: PathBuf,
}

impl World {
    fn new() -> Self {
        let home = TempHome::new();
        let project = home.project("demo");
        ledger::register(home.path(), "demo").unwrap();
        World { home, project }
    }

    fn p(&self) -> &str {
        self.project.to_str().unwrap()
    }

    /// `moai <args>` with no session id.
    fn run(&self, args: &[&str]) -> Output {
        self.home.command(args).output().expect("run moai")
    }

    /// `moai <args>` as the session `S_A`.
    fn run_as(&self, args: &[&str]) -> Output {
        self.home
            .command(args)
            .env(SESSION_VAR, S_A)
            .output()
            .expect("run moai")
    }

    /// `moai card <args> --project P`.
    fn card(&self, args: &[&str]) -> Output {
        let mut all = vec!["card"];
        all.extend_from_slice(args);
        all.extend_from_slice(&["--project", self.p()]);
        self.run(&all)
    }

    /// A ready card of `kind` with one write scope of its own; returns its name.
    fn ready_card(&self, kind: &str, scope: &str) -> String {
        let added = ok(&self.card(&[
            "add", "--title", "t", "--kind", kind, "--scope", scope,
        ]));
        let name = added["name"].as_str().unwrap().to_string();
        ok(&self.card(&["ready", &name]));
        name
    }

    /// A draft SPEC `SPEC-20261117-<slug>` with every template placeholder filled, so
    /// `spec check` passes it.
    fn draft_spec(&self, slug: &str) -> String {
        let r = ok(&self.run(&[
            "spec",
            "new",
            slug,
            "--title",
            "t",
            "--minimal",
            "--date",
            "2026-11-17",
            "--project",
            self.p(),
        ]));
        let id = r["id"].as_str().unwrap().to_string();
        let path = self.spec_file(&id);
        fs::write(&path, filled(&fs::read_to_string(&path).unwrap())).unwrap();
        id
    }

    /// Marks the SPEC approved with the digest `spec show` reports.
    fn approve(&self, id: &str) {
        let shown = ok(&self.run(&["spec", "show", id, "--project", self.p()]));
        let digest = shown["spec"]["digest"].as_str().unwrap().to_string();
        let path = self.spec_file(id);
        let text = fs::read_to_string(&path).unwrap().replace(
            "status: draft\n",
            &format!(
                "status: approved\napproval:\n  by: user\n  at: 2026-11-17\n  basis: test \
                 selection\n  digest: {digest}\n"
            ),
        );
        fs::write(&path, text).unwrap();
        ok(&self.run(&["spec", "check", id, "--project", self.p()]));
    }

    fn spec_file(&self, id: &str) -> PathBuf {
        self.project.join(".moai/specs").join(id).join("spec.md")
    }

    fn count(&self, sql: &str) -> i64 {
        let l = ledger::open_read_only(self.home.path(), "demo").unwrap();
        l.conn().query_row(sql, [], |r| r.get(0)).unwrap()
    }

    fn raw(&self) -> rusqlite::Connection {
        rusqlite::Connection::open(ledger::ledger_path(self.home.path(), "demo").unwrap()).unwrap()
    }
}

/// A template's text with every `<…>` placeholder filled with the word `filled`.
fn filled(text: &str) -> String {
    let mut out = String::new();
    let mut rest = text;
    while let Some(open) = rest.find('<') {
        let Some(close) = rest[open..].find('>') else {
            break;
        };
        out.push_str(&rest[..open]);
        out.push_str("filled");
        rest = &rest[open + close + 1..];
    }
    out.push_str(rest);
    out
}

fn warnings(v: &Value) -> Vec<String> {
    v["warnings"]
        .as_array()
        .unwrap_or_else(|| panic!("no warnings list: {v}"))
        .iter()
        .map(|w| w.as_str().unwrap().to_string())
        .collect()
}

/// REQ-1, REQ-3: link, link again (nothing new), relink, unlink; `show` and `list` follow.
#[test]
fn a_card_links_to_a_spec_unlinks_and_show_and_list_print_it() {
    let w = World::new();
    let login = w.draft_spec("login-fix");
    let other = w.draft_spec("other-fix");
    let card = w.ready_card("feature", "src/login");
    let spec_before = fs::read(w.spec_file(&login)).unwrap();

    let linked = ok(&w.card(&["spec", &card, &login, "--by", "leader"]));
    assert_eq!(linked["action"], json!("spec"), "{linked}");
    assert_eq!(linked["name"], json!(card));
    assert_eq!(linked["from"], Value::Null);
    assert_eq!(linked["spec"], json!(login));
    assert_eq!(linked["changed"], json!(true));
    assert_eq!(linked["events"], json!(1));
    assert_eq!(linked["writes"], json!(true));
    assert_eq!(linked["by"], json!("leader"));
    assert_eq!(ok(&w.card(&["show", &card]))["spec"], json!(login));
    let listed = ok(&w.card(&["list"]));
    assert_eq!(listed["cards"][0]["spec"], json!(login), "{listed}");

    // The same link again changes nothing and writes no event.
    let events = w.count("SELECT count(*) FROM event");
    let again = ok(&w.card(&["spec", &card, &login]));
    assert_eq!(again["changed"], json!(false), "{again}");
    assert_eq!(again["events"], json!(0));
    assert_eq!(again["writes"], json!(false));
    assert_eq!(w.count("SELECT count(*) FROM event"), events);

    // Another SPEC replaces it; the event keeps the old one.
    let moved = ok(&w.card(&["spec", &card, &other]));
    assert_eq!(moved["from"], json!(login));
    assert_eq!(moved["spec"], json!(other));

    let unlinked = ok(&w.card(&["spec", &card, "--unlink"]));
    assert_eq!(unlinked["from"], json!(other), "{unlinked}");
    assert_eq!(unlinked["spec"], Value::Null);
    assert_eq!(unlinked["changed"], json!(true));
    assert_eq!(ok(&w.card(&["show", &card]))["spec"], Value::Null);
    assert_eq!(ok(&w.card(&["list"]))["cards"][0]["spec"], Value::Null);
    let none = ok(&w.card(&["spec", &card, "--unlink"]));
    assert_eq!(none["changed"], json!(false), "{none}");
    assert_eq!(none["writes"], json!(false));

    assert_eq!(
        w.count("SELECT count(*) FROM event WHERE kind = 'card.spec'"),
        3,
        "link, relink and unlink; the repeats wrote nothing"
    );
    // The SPEC file never carries the card.
    assert_eq!(fs::read(w.spec_file(&login)).unwrap(), spec_before);
}

/// REQ-2: a SPEC that is not there, an id that is not a SPEC id, a SPEC with an error and a
/// card that is not there are refused, and the ledger is left as it was.
#[test]
fn a_missing_or_unparsable_spec_is_refused_and_nothing_is_written() {
    let w = World::new();
    let card = w.ready_card("fix", "src/a");
    let good = w.draft_spec("login-fix");
    let before = w.home.snapshot();

    let (kind, message) = refusal(&w.card(&["spec", &card, OTHER]));
    assert_eq!(kind, "not_found", "{message}");
    assert!(message.contains(OTHER), "{message}");

    for bad in ["C-3", "SPEC-2026-x", "spec-20261117-login-fix", "../SPEC-20261117-a"] {
        let (kind, message) = refusal(&w.card(&["spec", &card, bad]));
        assert_eq!(kind, "invalid_input", "{bad}: {message}");
        assert!(message.starts_with("card spec refused: "), "{message}");
    }

    // A SPEC with an error (no frontmatter) is not linked.
    let broken = w.project.join(".moai/specs/SPEC-20261117-broken");
    fs::create_dir_all(&broken).unwrap();
    fs::write(broken.join("spec.md"), "hello\n").unwrap();
    let before_broken = w.home.snapshot();
    let (kind, message) = refusal(&w.card(&["spec", &card, "SPEC-20261117-broken"]));
    assert_eq!(kind, "invalid_input", "{message}");
    assert!(
        message.contains("SPEC-20261117-broken") && message.contains("error"),
        "{message}"
    );
    assert_eq!(w.home.snapshot(), before_broken);
    fs::remove_dir_all(&broken).unwrap();

    let (kind, message) = refusal(&w.card(&["spec", "C-99", &good]));
    assert_eq!(kind, "not_found", "{message}");
    // A card is named by its C-number.
    let (kind, _) = refusal(&w.card(&["spec", "#12", &good]));
    assert_eq!(kind, "invalid_input");
    // `--unlink` and an id together are refused by the parser before anything is read.
    let out = w.card(&["spec", &card, &good, "--unlink"]);
    assert_ne!(out.status.code(), Some(0));

    assert_eq!(w.home.snapshot(), before, "no refusal writes anything");
    assert_eq!(
        w.count("SELECT count(*) FROM event WHERE kind = 'card.spec'"),
        0
    );
}

/// REQ-2 (busy): a ledger another connection holds for writing refuses the link after the
/// busy timeout, as `ledger busy`, and the holder's write is not lost.
#[test]
fn a_busy_ledger_refuses_the_link_and_keeps_the_other_write() {
    let w = World::new();
    let card = w.ready_card("feature", "src/a");
    let spec = w.draft_spec("login-fix");
    let holder = w.raw();
    holder.execute_batch("BEGIN IMMEDIATE").unwrap();
    holder
        .execute(
            "INSERT INTO event (event_key, at, actor, kind) VALUES ('held', 1, 't', 'test.held')",
            [],
        )
        .unwrap();
    let (kind, message) = refusal(&w.card(&["spec", &card, &spec]));
    assert!(message.contains("ledger busy"), "{message}");
    assert_eq!(kind, "io", "{message}");
    holder.execute_batch("COMMIT").unwrap();
    assert_eq!(
        w.count("SELECT count(*) FROM event WHERE kind = 'card.spec'"),
        0
    );
    assert_eq!(
        w.count("SELECT count(*) FROM event WHERE kind = 'test.held'"),
        1
    );
}

/// REQ-4, REQ-5: `next` and `take` print the linked SPEC's status and warn for a card that
/// changes the repository without an approved SPEC; `research` and `audit` cards are not
/// warned about.
#[test]
fn next_and_take_print_the_spec_status_and_warn_until_it_is_approved() {
    let w = World::new();
    let spec = w.draft_spec("login-fix");
    let linked = w.ready_card("feature", "src/a");
    let bare = w.ready_card("chore", "src/b");
    let research = w.ready_card("research", "notes/c");
    let audit = w.ready_card("audit", "notes/d");
    ok(&w.card(&["spec", &linked, &spec]));

    let next = ok(&w.run(&["card", "next", w.p()]));
    let ready = next["ready"].as_array().unwrap();
    let by_name = |name: &str| ready.iter().find(|c| c["name"] == json!(name)).unwrap();
    assert_eq!(
        by_name(&linked)["spec"],
        json!({"id": spec, "found": true, "status": "draft", "errors": 0, "approved": false}),
        "{next}"
    );
    assert_eq!(by_name(&bare)["spec"], Value::Null);
    assert_eq!(by_name(&research)["spec"], Value::Null);
    let warned = warnings(&next);
    assert_eq!(warned.len(), 2, "{warned:?}");
    assert!(
        warned[0].starts_with(&format!("{linked}: ")) && warned[0].contains("draft"),
        "{warned:?}"
    );
    assert!(
        warned[1].starts_with(&format!("{bare}: ")) && warned[1].contains("no SPEC"),
        "{warned:?}"
    );
    assert!(!warned.iter().any(|m| m.contains(&research) || m.contains(&audit)));

    // `take` warns too, and the card is taken.
    let taken = ok(&w.run_as(&["card", "take", &linked, "--project", w.p()]));
    assert_eq!(taken["state"], json!("active"), "{taken}");
    assert_eq!(taken["spec"]["status"], json!("draft"));
    assert_eq!(warnings(&taken).len(), 1, "{taken}");
    // A taken card is `mine` in `next`, with its SPEC and its warning.
    let next = ok(&w.run_as(&["card", "next", w.p()]));
    assert_eq!(next["mine"][0]["spec"]["id"], json!(spec), "{next}");
    assert!(warnings(&next).iter().any(|m| m.starts_with(&format!("{linked}: "))));

    // A research card is taken without a word.
    let taken = ok(&w.run_as(&["card", "take", &research, "--project", w.p()]));
    assert_eq!(warnings(&taken), Vec::<String>::new(), "{taken}");

    // Approved: no warning for the linked card any more.
    w.approve(&spec);
    let next = ok(&w.run_as(&["card", "next", w.p()]));
    assert_eq!(next["mine"][0]["spec"]["approved"], json!(true), "{next}");
    assert!(
        !warnings(&next).iter().any(|m| m.starts_with(&format!("{linked}: "))),
        "{next}"
    );
}

/// REQ-6: `--require-approved-spec` refuses a card that changes the repository unless its SPEC
/// is approved, and takes nothing; an approved SPEC and a research card pass.
#[test]
fn take_with_require_approved_spec_refuses_until_the_spec_is_approved() {
    let w = World::new();
    let spec = w.draft_spec("login-fix");
    let linked = w.ready_card("fix", "src/a");
    let bare = w.ready_card("docs", "docs/b");
    let research = w.ready_card("research", "notes/c");
    ok(&w.card(&["spec", &linked, &spec]));
    let take = |card: &str| {
        w.run_as(&[
            "card",
            "take",
            card,
            "--require-approved-spec",
            "--project",
            w.p(),
        ])
    };
    let reservations = || w.count("SELECT count(*) FROM reservation");

    for (card, why) in [(&linked, "draft"), (&bare, "no SPEC")] {
        let (kind, message) = refusal(&take(card));
        assert_eq!(kind, "invalid_input", "{message}");
        assert!(
            message.starts_with("card take refused: ") && message.contains(why),
            "{message}"
        );
        assert!(message.contains(card.as_str()), "{message}");
    }
    assert_eq!(reservations(), 0, "a refused take reserves nothing");
    // A SPEC that went missing after the link is not approved either.
    let gone = w.project.join(".moai/specs").join(&spec);
    let kept = w.project.join("kept-spec");
    fs::rename(&gone, &kept).unwrap();
    let (_, message) = refusal(&take(&linked));
    assert!(message.contains("not found"), "{message}");
    fs::rename(&kept, &gone).unwrap();

    ok(&take(&research));
    w.approve(&spec);
    let taken = ok(&take(&linked));
    assert_eq!(taken["spec"]["approved"], json!(true), "{taken}");
    assert_eq!(warnings(&taken), Vec::<String>::new());
    assert_eq!(reservations(), 2);
}

/// REQ-7: `spec show` lists the cards of this machine's ledger that name the SPEC, marked
/// local; no ledger, no profile and an unusable ledger still show the SPEC.
#[test]
fn spec_show_lists_the_linked_cards_of_this_machine_as_local() {
    let w = World::new();
    let login = w.draft_spec("login-fix");
    let other = w.draft_spec("other-fix");
    let one = w.ready_card("feature", "src/a");
    let two = w.ready_card("fix", "src/b");
    let three = w.ready_card("fix", "src/c");
    ok(&w.card(&["spec", &one, &login]));
    ok(&w.card(&["spec", &two, &login]));
    ok(&w.card(&["spec", &three, &login]));
    ok(&w.card(&["spec", &three, &other]));

    let shown = ok(&w.run(&["spec", "show", &login, "--project", w.p()]));
    let local = &shown["local_cards"];
    assert_eq!(local["local"], json!(true), "{shown}");
    assert_eq!(local["ledger"], json!("ready"));
    let names: Vec<&str> = local["cards"]
        .as_array()
        .unwrap()
        .iter()
        .map(|c| c["name"].as_str().unwrap())
        .collect();
    assert_eq!(names, [one.as_str(), two.as_str()], "{shown}");
    assert_eq!(local["cards"][0]["state"], json!("ready"));
    assert_eq!(local["cards"][0]["kind"], json!("feature"));
    let shown = ok(&w.run(&["spec", "show", &other, "--project", w.p()]));
    assert_eq!(shown["local_cards"]["cards"][0]["name"], json!(three));

    // `--home` names the HOME whose ledger is read: an empty one has none.
    let empty = TempHome::new();
    let shown = ok(&w.run(&[
        "spec",
        "show",
        &login,
        "--project",
        w.p(),
        "--home",
        empty.path().to_str().unwrap(),
    ]));
    assert_eq!(shown["local_cards"]["ledger"], json!("absent"), "{shown}");
    assert_eq!(shown["local_cards"]["cards"], json!([]));

    // A ledger that is not usable does not hide the SPEC.
    let path = ledger::ledger_path(w.home.path(), "demo").unwrap();
    fs::write(&path, b"this is not a database").unwrap();
    let shown = ok(&w.run(&["spec", "show", &login, "--project", w.p()]));
    assert_eq!(shown["spec"]["id"], json!(login));
    assert_eq!(
        shown["local_cards"]["ledger"],
        json!("ledger_unavailable"),
        "{shown}"
    );
    assert_eq!(shown["local_cards"]["cards"], json!([]));
}

/// REQ-7: a project without a profile has no ledger to read; `spec show` says so.
#[test]
fn spec_show_without_a_project_profile_has_no_local_cards() {
    let home = TempHome::new();
    let project = home.path().join("plain");
    fs::create_dir_all(&project).unwrap();
    let p = project.to_str().unwrap();
    let id = ok(&home
        .command(&[
            "spec", "new", "login-fix", "--title", "t", "--date", "2026-11-17", "--project", p,
        ])
        .output()
        .unwrap())["id"]
        .as_str()
        .unwrap()
        .to_string();
    let shown = ok(&home
        .command(&["spec", "show", &id, "--project", p])
        .output()
        .unwrap());
    assert_eq!(shown["local_cards"]["ledger"], json!("no_project"), "{shown}");
    assert_eq!(shown["local_cards"]["cards"], json!([]));
}

/// REQ-8: the link needs no schema change: a ledger of the current schema takes it as an
/// event, keeps its version, and `card export` carries the event.
#[test]
fn the_link_is_an_event_of_the_current_schema_and_the_export_keeps_it() {
    let w = World::new();
    let card = w.ready_card("feature", "src/a");
    let spec = w.draft_spec("login-fix");
    let version = |c: &rusqlite::Connection| -> i64 {
        c.query_row("PRAGMA user_version", [], |r| r.get(0)).unwrap()
    };
    let before = version(&w.raw());
    ok(&w.card(&["spec", &card, &spec]));
    assert_eq!(version(&w.raw()), before);
    assert_eq!(before, ledger::SCHEMA_VERSION);
    let payload: String = w
        .raw()
        .query_row(
            "SELECT payload FROM event WHERE kind = 'card.spec'",
            [],
            |r| r.get(0),
        )
        .unwrap();
    assert_eq!(
        serde_json::from_str::<Value>(&payload).unwrap(),
        json!({"from": null, "to": spec})
    );
    let out = w.home.path().join("export");
    ok(&w.card(&["export", "--out", out.to_str().unwrap()]));
    let events = fs::read_to_string(out.join("events.jsonl")).unwrap();
    assert!(
        events.lines().any(|l| l.contains("\"card.spec\"") && l.contains(&spec)),
        "{events}"
    );
}
