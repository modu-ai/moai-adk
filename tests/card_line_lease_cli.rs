//! `moai card review --keep`, `card give-back` and `card line-lease`
//! (SPEC-20261120-card-shared-line-lease): a card gives back the paths it no longer writes, and a
//! short lease adds one line to a shared file without the card's main reservation ever holding
//! that file.
//!
//! Every command runs against a temporary HOME; the real `~/.moai` is not read or written. The
//! holder is the session id in the environment (`CLAUDE_CODE_SESSION_ID`).
#![cfg(unix)]

mod common;

use common::home::TempHome;
use moai_cli::ledger;
use rusqlite::Connection;
use serde_json::{json, Value};
use std::fs;
use std::path::Path;
use std::process::{Output, Stdio};
use std::time::{Duration, Instant};

const SESSION_VAR: &str = "CLAUDE_CODE_SESSION_ID";
const S_A: &str = "aaaaaaaa-0000-4000-8000-00000000000a";
const S_B: &str = "bbbbbbbb-0000-4000-8000-00000000000b";
const S_C: &str = "cccccccc-0000-4000-8000-00000000000c";
const SPEC_PATH: &str = ".moai/specs/SPEC-20261120-demo/spec.md";

struct World {
    home: TempHome,
    project: String,
}

fn world() -> World {
    let home = TempHome::new();
    let project = home.project("demo");
    ledger::register(home.path(), "demo").unwrap();
    World {
        home,
        project: project.to_str().unwrap().to_string(),
    }
}

impl World {
    /// `moai card <args> --project <project>`, as `session`.
    fn card(&self, session: Option<&str>, args: &[&str]) -> Output {
        self.command(session, args).output().expect("run moai")
    }

    fn command(&self, session: Option<&str>, args: &[&str]) -> std::process::Command {
        let mut all = vec!["card"];
        all.extend_from_slice(args);
        all.extend_from_slice(&["--project", &self.project]);
        let mut cmd = self.home.command(&all);
        if let Some(id) = session {
            cmd.env(SESSION_VAR, id);
        }
        cmd.stdin(Stdio::null());
        cmd
    }

    /// A ready card with one open criterion and `scope`, taken by `session` at generation 1.
    fn held(&self, session: &str, title: &str, scope: &[&str]) -> String {
        let mut args = vec!["add", "--title", title, "--ac", "one"];
        for entry in scope {
            args.push("--scope");
            args.push(entry);
        }
        let name = result(&self.card(None, &args))["name"]
            .as_str()
            .unwrap()
            .to_string();
        result(&self.card(None, &["ready", &name]));
        result(&self.card(
            Some(session),
            &[
                "take",
                &name,
                "--worktree",
                "/work/tree",
                "--branch",
                "card/x",
            ],
        ));
        name
    }

    fn reservation(&self, card: &str) -> Value {
        result(&self.card(None, &["show", card]))["reservation"].clone()
    }

    fn scope_of(&self, card: &str) -> Vec<String> {
        self.reservation(card)["write_scope"]
            .as_array()
            .unwrap()
            .iter()
            .map(|v| v.as_str().unwrap().to_string())
            .collect()
    }

    fn rows(&self, sql: &str) -> Vec<Vec<Value>> {
        let l = ledger::open_read_only(self.home.path(), "demo").unwrap();
        let mut st = l.conn().prepare(sql).unwrap();
        let n = st.column_count();
        st.query_map([], |r| {
            Ok((0..n)
                .map(|i| match r.get_ref(i).unwrap() {
                    rusqlite::types::ValueRef::Integer(v) => json!(v),
                    rusqlite::types::ValueRef::Text(t) => {
                        json!(String::from_utf8_lossy(t).to_string())
                    }
                    _ => Value::Null,
                })
                .collect())
        })
        .unwrap()
        .map(Result::unwrap)
        .collect()
    }

    /// The `card.scope` events as payloads, oldest first.
    fn scope_events(&self) -> Vec<Value> {
        self.rows("SELECT payload FROM event WHERE kind = 'card.scope' ORDER BY event_id")
            .into_iter()
            .map(|row| serde_json::from_str(row[0].as_str().unwrap()).unwrap())
            .collect()
    }

    fn dump(&self) -> Vec<String> {
        let l = ledger::open_read_only(self.home.path(), "demo").unwrap();
        dump_tables(
            l.conn(),
            &[
                "card",
                "card_note",
                "reservation",
                "reservation_scope",
                "event",
            ],
        )
    }

    /// `card line-lease <card> --generation 1 --path <path> [--wait-secs n] -- <command…>`.
    fn lease(
        &self,
        session: &str,
        card: &str,
        path: &str,
        wait: Option<&str>,
        command: &[&str],
    ) -> Output {
        self.lease_command(session, card, path, wait, command)
            .output()
            .expect("run moai")
    }

    fn lease_command(
        &self,
        session: &str,
        card: &str,
        path: &str,
        wait: Option<&str>,
        command: &[&str],
    ) -> std::process::Command {
        let mut args = vec![
            "line-lease",
            card,
            "--generation",
            "1",
            "--path",
            path,
        ];
        if let Some(wait) = wait {
            args.extend_from_slice(&["--wait-secs", wait]);
        }
        args.push("--");
        args.extend_from_slice(command);
        self.command(Some(session), &args)
    }
}

fn dump_tables(conn: &Connection, tables: &[&str]) -> Vec<String> {
    let mut out = Vec::new();
    for table in tables {
        let mut st = conn
            .prepare(&format!("SELECT * FROM {table} ORDER BY 1, 2"))
            .unwrap();
        let n = st.column_count();
        let lines: Vec<String> = st
            .query_map([], |r| {
                let mut line = format!("{table}:");
                for i in 0..n {
                    line.push_str(&format!("{:?}|", r.get_ref(i)?));
                }
                Ok(line)
            })
            .unwrap()
            .map(Result::unwrap)
            .collect();
        out.extend(lines);
    }
    out
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
    assert_eq!(v["operation"], json!("card"), "{v}");
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

fn strings(v: &Value) -> Vec<String> {
    v.as_array()
        .unwrap_or_else(|| panic!("not an array: {v}"))
        .iter()
        .map(|s| s.as_str().unwrap().to_string())
        .collect()
}

// ---- review ------------------------------------------------------------------------------------

#[test]
fn review_gives_back_every_path_that_no_keep_names() {
    let w = world();
    let a = w.held(S_A, "writer", &["src/a.rs", "tests/a.rs", SPEC_PATH]);
    let done = result(&w.card(
        Some(S_A),
        &["review", &a, "--generation", "1", "--keep", SPEC_PATH],
    ));
    assert_eq!(done["state"], json!("review"), "{done}");
    assert_eq!(
        strings(&done["given_back"]),
        ["src/a.rs", "tests/a.rs"],
        "{done}"
    );
    assert_eq!(strings(&done["kept"]), [SPEC_PATH], "{done}");
    assert_eq!(w.scope_of(&a), [SPEC_PATH]);
    // The event names both lists.
    let events = w.rows("SELECT payload FROM event WHERE kind = 'card.review'");
    let payload: Value = serde_json::from_str(events[0][0].as_str().unwrap()).unwrap();
    assert_eq!(
        strings(&payload["given_back"]),
        ["src/a.rs", "tests/a.rs"],
        "{payload}"
    );
    // Another card takes a given-back path at once.
    let b = w.held(S_B, "next", &["src/a.rs"]);
    assert_eq!(w.scope_of(&b), ["src/a.rs"]);
}

#[test]
fn review_without_keep_gives_back_the_whole_scope() {
    let w = world();
    let a = w.held(S_A, "writer", &["src/a.rs", "tests/a.rs"]);
    let done = result(&w.card(Some(S_A), &["review", &a, "--generation", "1"]));
    assert_eq!(
        strings(&done["given_back"]),
        ["src/a.rs", "tests/a.rs"],
        "{done}"
    );
    assert_eq!(strings(&done["kept"]), Vec::<String>::new(), "{done}");
    assert_eq!(w.scope_of(&a), Vec::<String>::new());
    // The reservation is still held: a repair widens it again with a reason.
    result(&w.card(Some(S_A), &["reopen", &a, "--generation", "1"]));
    let widened = result(&w.card(
        None,
        &[
            "scope",
            &a,
            "--scope",
            "src/a.rs",
            "--by",
            "leader",
            "--reason",
            "the audit asked for a repair in this file",
        ],
    ));
    assert_eq!(strings(&widened["write_scope"]), ["src/a.rs"], "{widened}");
}

#[test]
fn review_refuses_a_keep_path_the_card_does_not_hold() {
    let w = world();
    let a = w.held(S_A, "writer", &["src/a.rs"]);
    let before = w.dump();
    let (kind, message) = refusal(&w.card(
        Some(S_A),
        &["review", &a, "--generation", "1", "--keep", "src/other.rs"],
    ));
    assert_eq!(kind, "conflict", "{message}");
    assert!(message.contains("src/other.rs"), "{message}");
    assert_eq!(w.dump(), before, "a refused review writes nothing");
    assert_eq!(
        result(&w.card(None, &["show", &a]))["state"],
        json!("active")
    );
}

// ---- give-back ---------------------------------------------------------------------------------

#[test]
fn give_back_removes_the_named_paths_and_writes_one_event() {
    let w = world();
    let a = w.held(S_A, "writer", &["src/a.rs", "tests/a.rs", "docs/a.md"]);
    let events_before = w.scope_events().len();
    let done = result(&w.card(
        Some(S_A),
        &[
            "give-back",
            &a,
            "--generation",
            "1",
            "--path",
            "src/a.rs",
            "--path",
            "docs/a.md",
        ],
    ));
    assert_eq!(done["action"], json!("give-back"), "{done}");
    assert_eq!(
        strings(&done["given_back"]),
        ["docs/a.md", "src/a.rs"],
        "{done}"
    );
    assert_eq!(strings(&done["write_scope"]), ["tests/a.rs"], "{done}");
    assert_eq!(w.scope_of(&a), ["tests/a.rs"]);
    let events = w.scope_events();
    assert_eq!(events.len(), events_before + 1, "one event: {events:?}");
    let last = events.last().unwrap();
    assert_eq!(
        strings(&last["given_back"]),
        ["docs/a.md", "src/a.rs"],
        "{last}"
    );
    assert_eq!(last["generation"], json!(1), "{last}");
}

#[test]
fn give_back_refuses_an_unknown_path_and_a_stale_generation_and_writes_nothing() {
    let w = world();
    let a = w.held(S_A, "writer", &["src/a.rs"]);
    let before = w.dump();
    let (kind, message) = refusal(&w.card(
        Some(S_A),
        &[
            "give-back",
            &a,
            "--generation",
            "1",
            "--path",
            "src/never.rs",
        ],
    ));
    assert_eq!(kind, "conflict", "{message}");
    assert!(message.contains("src/never.rs"), "{message}");
    let (kind, message) = refusal(&w.card(
        Some(S_A),
        &["give-back", &a, "--generation", "2", "--path", "src/a.rs"],
    ));
    assert_eq!(kind, "conflict", "{message}");
    // Another session is not the holder.
    let (kind, _) = refusal(&w.card(
        Some(S_B),
        &["give-back", &a, "--generation", "1", "--path", "src/a.rs"],
    ));
    assert_eq!(kind, "conflict");
    assert_eq!(w.dump(), before, "refusals write nothing");
}

// ---- line-lease --------------------------------------------------------------------------------

fn append_command(file: &Path, text: &str) -> Vec<String> {
    vec![
        "sh".into(),
        "-c".into(),
        format!("printf '%s\\n' '{text}' >> '{}'", file.display()),
    ]
}

fn as_refs(v: &[String]) -> Vec<&str> {
    v.iter().map(String::as_str).collect()
}

#[test]
fn line_lease_adds_runs_and_gives_the_path_back_in_one_call() {
    let w = world();
    let a = w.held(S_A, "writer", &["src/a.rs"]);
    let file = Path::new(&w.project).join("lib.rs");
    let cmd = append_command(&file, "pub mod a;");
    let done = result(&w.lease(S_A, &a, "src/lib.rs", None, &as_refs(&cmd)));
    assert_eq!(done["action"], json!("line-lease"), "{done}");
    assert_eq!(done["exit_code"], json!(0), "{done}");
    assert_eq!(strings(&done["added"]), ["src/lib.rs"], "{done}");
    assert_eq!(strings(&done["given_back"]), ["src/lib.rs"], "{done}");
    assert_eq!(fs::read_to_string(&file).unwrap(), "pub mod a;\n");
    // The main reservation never holds the shared file afterwards.
    assert_eq!(w.scope_of(&a), ["src/a.rs"]);
    // The history keeps both steps.
    let events = w.scope_events();
    let widened = &events[events.len() - 2];
    let narrowed = &events[events.len() - 1];
    assert!(
        strings(&widened["to"]).contains(&"src/lib.rs".to_string()),
        "{widened}"
    );
    assert_eq!(strings(&narrowed["given_back"]), ["src/lib.rs"], "{narrowed}");
}

#[test]
fn line_lease_lets_two_cards_each_add_a_line_to_the_same_file() {
    let w = world();
    let a = w.held(S_A, "first", &["src/a.rs"]);
    let b = w.held(S_B, "second", &["src/b.rs"]);
    let file = Path::new(&w.project).join("lib.rs");
    let started = Path::new(&w.project).join("a-started");
    // A's command holds the lease for a while.
    let slow = vec![
        "sh".to_string(),
        "-c".to_string(),
        format!(
            "touch '{}'; sleep 2; printf '%s\\n' 'pub mod a;' >> '{}'",
            started.display(),
            file.display()
        ),
    ];
    let mut first = w
        .lease_command(S_A, &a, "src/lib.rs", None, &as_refs(&slow))
        .stdout(Stdio::piped())
        .stderr(Stdio::piped())
        .spawn()
        .expect("run moai");
    let deadline = Instant::now() + Duration::from_secs(60);
    while !started.exists() {
        assert!(Instant::now() < deadline, "the first lease never started");
        std::thread::sleep(Duration::from_millis(50));
    }
    // While A holds it, B waits and then adds its own line.
    let cmd = append_command(&file, "pub mod b;");
    let second = result(&w.lease(S_B, &b, "src/lib.rs", Some("30"), &as_refs(&cmd)));
    assert!(
        second["waited_ms"].as_i64().unwrap() >= 500,
        "B should have waited for A: {second}"
    );
    let first = first.wait_with_output().unwrap();
    result(&first);
    assert_eq!(
        fs::read_to_string(&file).unwrap(),
        "pub mod a;\npub mod b;\n"
    );
    assert_eq!(w.scope_of(&a), ["src/a.rs"]);
    assert_eq!(w.scope_of(&b), ["src/b.rs"]);
}

#[test]
fn line_lease_refuses_after_the_wait_limit_naming_the_other_card() {
    let w = world();
    let owner = w.held(S_C, "long reservation", &["src/lib.rs"]);
    let b = w.held(S_B, "needs one line", &["src/b.rs"]);
    let file = Path::new(&w.project).join("lib.rs");
    let cmd = append_command(&file, "pub mod b;");
    let before = w.dump();
    let started = Instant::now();
    let (kind, message) = refusal(&w.lease(S_B, &b, "src/lib.rs", Some("1"), &as_refs(&cmd)));
    assert_eq!(kind, "conflict", "{message}");
    assert!(message.contains(&owner), "names the other card: {message}");
    assert!(
        started.elapsed() >= Duration::from_millis(900),
        "it waited for the limit"
    );
    assert!(!file.exists(), "the command did not run");
    assert_eq!(w.dump(), before, "a refused lease writes nothing");
}

#[test]
fn line_lease_gives_the_path_back_when_the_command_fails_or_cannot_start() {
    let w = world();
    let a = w.held(S_A, "writer", &["src/a.rs"]);
    let (_, message) = refusal(&w.lease(S_A, &a, "src/lib.rs", None, &["sh", "-c", "exit 7"]));
    assert!(
        message.contains("status 7"),
        "the command's status is named: {message}"
    );
    assert_eq!(w.scope_of(&a), ["src/a.rs"], "given back after a failure");
    let (_, message) = refusal(&w.lease(S_A, &a, "src/lib.rs", None, &["/no/such/command-here"]));
    assert!(
        message.contains("could not start"),
        "a start failure is named: {message}"
    );
    assert_eq!(w.scope_of(&a), ["src/a.rs"], "given back after a start failure");
    // The ledger is usable afterwards: the same path can be leased again.
    let file = Path::new(&w.project).join("lib.rs");
    let cmd = append_command(&file, "pub mod a;");
    result(&w.lease(S_A, &a, "src/lib.rs", None, &as_refs(&cmd)));
}

#[test]
fn line_lease_leaves_a_path_the_card_already_holds_in_its_scope() {
    let w = world();
    let a = w.held(S_A, "holds lib", &["src/lib.rs"]);
    let file = Path::new(&w.project).join("lib.rs");
    let cmd = append_command(&file, "pub mod a;");
    let done = result(&w.lease(S_A, &a, "src/lib.rs", None, &as_refs(&cmd)));
    assert_eq!(strings(&done["added"]), Vec::<String>::new(), "{done}");
    assert_eq!(strings(&done["given_back"]), Vec::<String>::new(), "{done}");
    assert_eq!(w.scope_of(&a), ["src/lib.rs"], "a held path stays held");
}

// ---- help --------------------------------------------------------------------------------------

fn help(w: &World, sub: &str) -> String {
    let out = w.card(None, &[sub, "--help"]);
    String::from_utf8_lossy(&out.stdout).to_string()
}

#[test]
fn help_names_the_rules_of_the_three_commands() {
    let w = world();
    let review = help(&w, "review");
    for phrase in ["--keep", "given back", "free for any card", "card scope"] {
        assert!(review.contains(phrase), "review help lacks `{phrase}`: {review}");
    }
    let give = help(&w, "give-back");
    for phrase in ["--path", "given back", "free for any card"] {
        assert!(give.contains(phrase), "give-back help lacks `{phrase}`: {give}");
    }
    let lease = help(&w, "line-lease");
    for phrase in [
        "--wait-secs",
        "gives the paths back",
        "standard error",
        "status",
    ] {
        assert!(lease.contains(phrase), "line-lease help lacks `{phrase}`: {lease}");
    }
    let doc = fs::read_to_string(
        Path::new(env!("CARGO_MANIFEST_DIR")).join("docs/design-review/appendix/operations/card.md"),
    )
    .unwrap();
    for phrase in ["line-lease", "give-back", "--keep"] {
        assert!(doc.contains(phrase), "the card operations document lacks `{phrase}`");
    }
}
