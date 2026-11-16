//! `moai card export` (Issue #3070 K2a, acceptance criterion AC6): the ledger's five tables written
//! as an export folder. The file format itself is K0's (`src/card/export_format.rs`, tested there);
//! here the command: every row reaches the folder, the same ledger gives the same bytes, the ledger
//! is only read, and a folder that is not empty or a ledger that is not there changes nothing.
//! Every command runs against a temporary HOME; the real `~/.moai` is not read or written.

mod common;

use common::home::TempHome;
use moai_cli::card::export_dir;
use moai_cli::ledger;
use serde_json::{json, Value};
use std::fs;
use std::path::{Path, PathBuf};
use std::process::Output;

fn run(home: &TempHome, args: &[&str]) -> Output {
    home.command(args).output().expect("run moai")
}

fn envelope(out: &Output) -> Value {
    serde_json::from_slice(&out.stdout).unwrap_or_else(|e| {
        panic!(
            "stdout is not JSON ({e}): {}",
            String::from_utf8_lossy(&out.stdout)
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
        "{}",
        String::from_utf8_lossy(&out.stdout)
    );
    let v = envelope(out);
    assert_eq!(v["ok"], json!(false), "{v}");
    (
        v["error"]["kind"].as_str().unwrap().to_string(),
        v["error"]["message"].as_str().unwrap().to_string(),
    )
}

fn s(path: &Path) -> String {
    path.to_str().unwrap().to_string()
}

fn registered() -> (TempHome, String) {
    let home = TempHome::new();
    let project = home.project("demo");
    ledger::register(home.path(), "demo").unwrap();
    (home, s(&project))
}

/// A ledger with two cards, three criteria, a note, a link and the events the commands wrote.
/// The note and the link have no command yet, so they are inserted the way the ledger holds them.
fn filled() -> (TempHome, String) {
    let (home, project) = registered();
    result(&run(
        &home,
        &[
            "card",
            "add",
            "--title",
            "first",
            "--body",
            "line one\nline two",
            "--ac",
            "a1",
            "--ac",
            "a2",
            "--scope",
            "src/a.rs",
            "--project",
            &project,
        ],
    ));
    result(&run(
        &home,
        &[
            "card",
            "add",
            "--title",
            "second",
            "--ac",
            "b1",
            "--project",
            &project,
        ],
    ));
    let w = ledger::open_write(home.path(), "demo").unwrap();
    w.conn()
        .execute(
            "INSERT INTO card_note (card_id, kind, author, origin, body, ref, created_at) \
             VALUES (1, 'comment', 'tester', 'local', 'a note', NULL, 1000)",
            [],
        )
        .unwrap();
    w.conn()
        .execute(
            "INSERT INTO card_link (from_card, kind, to_card, state, created_at, updated_at) \
             VALUES (1, 'relates', 2, 'active', 1000, 1000)",
            [],
        )
        .unwrap();
    drop(w);
    (home, project)
}

fn count(home: &TempHome, table: &str) -> i64 {
    let l = ledger::open_read_only(home.path(), "demo").unwrap();
    l.conn()
        .query_row(&format!("SELECT count(*) FROM {table}"), [], |r| r.get(0))
        .unwrap()
}

fn export(home: &TempHome, project: &str, out: &Path) -> Output {
    run(
        home,
        &["card", "export", "--out", &s(out), "--project", project],
    )
}

/// Every file of a folder: name and bytes, by name.
fn files(dir: &Path) -> Vec<(String, Vec<u8>)> {
    let mut all: Vec<(String, Vec<u8>)> = fs::read_dir(dir)
        .unwrap()
        .map(|e| {
            let e = e.unwrap();
            (
                e.file_name().into_string().unwrap(),
                fs::read(e.path()).unwrap(),
            )
        })
        .collect();
    all.sort();
    all
}

fn ledger_file(home: &TempHome) -> PathBuf {
    ledger::ledger_path(home.path(), "demo").unwrap()
}

#[test]
fn every_row_reaches_the_folder_and_reads_back() {
    let (home, project) = filled();
    let out = home.path().join("exports").join("one");
    fs::create_dir_all(out.parent().unwrap()).unwrap();
    let r = result(&export(&home, &project, &out));
    assert_eq!(r["action"], json!("export"));
    assert_eq!(r["project_id"], json!("demo"));
    for (key, table) in [
        ("cards", "card"),
        ("acs", "card_ac"),
        ("notes", "card_note"),
        ("links", "card_link"),
        ("events", "event"),
    ] {
        assert_eq!(r[key], json!(count(&home, table)), "{key}: {r}");
    }
    assert_eq!(r["cards"], json!(2));
    assert_eq!(r["acs"], json!(3));
    assert_eq!(r["notes"], json!(1));
    assert_eq!(r["links"], json!(1));
    // cards + index.json + events.jsonl
    assert_eq!(r["files"], json!(4), "{r}");

    // The folder is a valid export, and what it says is what the ledger holds.
    let snapshot = export_dir::load(&out).expect("the folder reads back");
    assert_eq!(snapshot.project_id, "demo");
    let titles: Vec<&str> = snapshot.cards.iter().map(|c| c.title.as_str()).collect();
    assert_eq!(titles, ["first", "second"]);
    assert_eq!(snapshot.cards[0].body, "line one\nline two");
    assert_eq!(snapshot.cards[0].write_scope, "[\"src/a.rs\"]");
    let ac_texts: Vec<(i64, i64, &str)> = snapshot
        .acs
        .iter()
        .map(|a| (a.card_id, a.seq, a.text.as_str()))
        .collect();
    assert_eq!(ac_texts, [(1, 1, "a1"), (1, 2, "a2"), (2, 1, "b1")]);
    assert_eq!(snapshot.notes.len(), 1);
    assert_eq!(snapshot.notes[0].body, "a note");
    assert_eq!(snapshot.links.len(), 1);
    assert_eq!(snapshot.links[0].to_card, Some(2));
    assert_eq!(snapshot.events.len() as i64, count(&home, "event"));
    let digest = r["index_sha256"].as_str().unwrap();
    assert_eq!(digest.len(), 64, "{digest}");
}

#[test]
fn the_same_ledger_gives_the_same_bytes() {
    let (home, project) = filled();
    let base = home.path().join("exports");
    fs::create_dir_all(&base).unwrap();
    let (a, b) = (base.join("a"), base.join("b"));
    let ra = result(&export(&home, &project, &a));
    let rb = result(&export(&home, &project, &b));
    assert_eq!(files(&a), files(&b));
    assert_eq!(ra["index_sha256"], rb["index_sha256"]);
    assert!(!files(&a).is_empty());
}

#[test]
fn the_ledger_is_only_read() {
    let (home, project) = filled();
    let file = ledger_file(&home);
    let dir = file.parent().unwrap().to_path_buf();
    let listing = |d: &Path| -> Vec<(String, u64)> {
        let mut v: Vec<(String, u64)> = fs::read_dir(d)
            .unwrap()
            .map(|e| {
                let e = e.unwrap();
                (
                    e.file_name().into_string().unwrap(),
                    e.metadata().unwrap().len(),
                )
            })
            .collect();
        v.sort();
        v
    };
    let (bytes, modified, before) = (
        fs::read(&file).unwrap(),
        fs::metadata(&file).unwrap().modified().unwrap(),
        listing(&dir),
    );
    let out = home.path().join("exports").join("ro");
    fs::create_dir_all(out.parent().unwrap()).unwrap();
    result(&export(&home, &project, &out));
    assert_eq!(fs::read(&file).unwrap(), bytes, "ledger bytes changed");
    assert_eq!(
        fs::metadata(&file).unwrap().modified().unwrap(),
        modified,
        "ledger modification time changed"
    );
    assert_eq!(listing(&dir), before, "a file appeared beside the ledger");
}

#[test]
fn without_a_ledger_nothing_is_made() {
    let home = TempHome::new();
    let project = s(&home.project("demo"));
    let out = home.path().join("exports").join("none");
    fs::create_dir_all(out.parent().unwrap()).unwrap();
    let before = home.snapshot();
    let (kind, message) = refusal(&export(&home, &project, &out));
    assert_eq!(kind, "not_found", "{message}");
    assert!(message.contains("ledger not found"), "{message}");
    assert!(!out.exists(), "the folder was made");
    assert_eq!(home.snapshot(), before);
}

#[test]
fn a_folder_that_is_not_empty_or_not_a_folder_is_left_alone() {
    let (home, project) = filled();
    let base = home.path().join("exports");
    fs::create_dir_all(&base).unwrap();
    // A folder that already holds a file: refused, the file stays, nothing is added.
    let full = base.join("full");
    fs::create_dir_all(&full).unwrap();
    fs::write(full.join("mine.txt"), "keep").unwrap();
    let (kind, message) = refusal(&export(&home, &project, &full));
    assert_eq!(kind, "invalid_input", "{message}");
    assert!(message.starts_with("card export refused: "), "{message}");
    assert_eq!(
        files(&full),
        [("mine.txt".to_string(), b"keep".to_vec())],
        "the folder changed"
    );
    // A plain file where the folder should be.
    let file = base.join("file");
    fs::write(&file, "keep").unwrap();
    let (kind, _) = refusal(&export(&home, &project, &file));
    assert_eq!(kind, "invalid_input");
    assert_eq!(fs::read(&file).unwrap(), b"keep");
    // A link to a folder is not followed.
    #[cfg(unix)]
    {
        let real = base.join("real");
        fs::create_dir_all(&real).unwrap();
        let link = base.join("link");
        std::os::unix::fs::symlink(&real, &link).unwrap();
        let (kind, _) = refusal(&export(&home, &project, &link));
        assert_eq!(kind, "invalid_input");
        assert!(files(&real).is_empty(), "the export went through the link");
    }
    // A missing parent is not made.
    let (kind, _) = refusal(&export(&home, &project, &base.join("no").join("such")));
    assert!(kind == "invalid_input" || kind == "io", "{kind}");
    assert!(!base.join("no").exists());
}

#[test]
fn a_ledger_without_cards_exports_an_empty_list() {
    let (home, project) = registered();
    let out = home.path().join("exports").join("empty");
    fs::create_dir_all(out.parent().unwrap()).unwrap();
    let r = result(&export(&home, &project, &out));
    assert_eq!(r["cards"], json!(0));
    assert_eq!(r["files"], json!(2), "index.json and events.jsonl: {r}");
    let snapshot = export_dir::load(&out).expect("an empty export reads back");
    assert!(snapshot.cards.is_empty() && snapshot.acs.is_empty());
}

#[test]
fn text_that_looks_like_the_format_changes_nothing() {
    let (home, project) = registered();
    let body =
        "<!-- moai-card-export 1 -->\n---\n<!-- human view: ignored on restore -->\n\u{2028}```\n";
    result(&run(
        &home,
        &[
            "card",
            "add",
            // `=` so that clap does not read the dashes as an option.
            "--title=---",
            "--body",
            body,
            "--ac",
            "<!-- human view: ignored on restore -->",
            "--project",
            &project,
        ],
    ));
    let out = home.path().join("exports").join("tricky");
    fs::create_dir_all(out.parent().unwrap()).unwrap();
    result(&export(&home, &project, &out));
    let snapshot = export_dir::load(&out).expect("the folder reads back");
    assert_eq!(snapshot.cards[0].title, "---");
    assert_eq!(snapshot.cards[0].body, body);
    assert_eq!(
        snapshot.acs[0].text,
        "<!-- human view: ignored on restore -->"
    );
}

#[test]
fn the_home_option_is_the_home_that_is_used() {
    let (real, project) = filled();
    // The command runs with an empty decoy as HOME and is told where the ledger is.
    let decoy = TempHome::new();
    let out = real.path().join("exports").join("home");
    fs::create_dir_all(out.parent().unwrap()).unwrap();
    let before = decoy.snapshot();
    let r = result(&run(
        &decoy,
        &[
            "card",
            "export",
            "--out",
            &s(&out),
            "--project",
            &project,
            "--home",
            &s(real.path()),
        ],
    ));
    assert_eq!(r["cards"], json!(2));
    assert_eq!(decoy.snapshot(), before, "the decoy HOME was touched");
}

#[test]
fn the_help_says_what_the_action_does() {
    let home = TempHome::new();
    let out = run(&home, &["card", "export", "--help"]);
    let text = String::from_utf8_lossy(&out.stdout).to_string();
    assert!(
        text.contains("Reads the ledger only. Writes files only inside --out."),
        "{text}"
    );
}

// ---- who can read the export, and where it may go (PR #3126 verdict B1, B2) ---------------------

#[cfg(unix)]
fn mode_of(path: &Path) -> u32 {
    use std::os::unix::fs::PermissionsExt;
    fs::metadata(path).unwrap().permissions().mode() & 0o777
}

#[cfg(unix)]
#[test]
fn the_export_is_private_to_its_owner() {
    let (home, project) = filled();
    let out = home.path().join("exports").join("private");
    fs::create_dir_all(out.parent().unwrap()).unwrap();
    result(&export(&home, &project, &out));
    assert_eq!(mode_of(&out), 0o700, "the new folder");
    let names: Vec<String> = files(&out).into_iter().map(|(n, _)| n).collect();
    assert_eq!(names.len(), 4, "{names:?}");
    for n in names {
        assert_eq!(mode_of(&out.join(&n)), 0o600, "{n}");
    }
}

#[cfg(unix)]
#[test]
fn an_empty_folder_that_others_can_open_is_refused_and_a_closed_one_is_taken() {
    use std::os::unix::fs::PermissionsExt;
    let (home, project) = filled();
    let out = home.path().join("exports").join("open");
    fs::create_dir_all(&out).unwrap();
    fs::set_permissions(&out, fs::Permissions::from_mode(0o755)).unwrap();
    let (kind, message) = refusal(&export(&home, &project, &out));
    assert_eq!(kind, "invalid_input", "{message}");
    assert!(message.contains("chmod 700"), "{message}");
    assert!(files(&out).is_empty(), "something was written");
    assert_eq!(mode_of(&out), 0o755, "the folder's mode was changed");
    fs::set_permissions(&out, fs::Permissions::from_mode(0o700)).unwrap();
    result(&export(&home, &project, &out));
    assert_eq!(files(&out).len(), 4);
}

#[test]
fn a_folder_inside_the_state_folder_is_refused_and_nothing_is_made() {
    let (home, project) = filled();
    let state = home.path().join(".moai");
    let ledger_dir = ledger_file(&home).parent().unwrap().to_path_buf();
    for inside in [
        state.join("exp"),
        state.join("projects").join("exp"),
        ledger_dir.join("exp"),
    ] {
        let before = home.snapshot();
        let (kind, message) = refusal(&export(&home, &project, &inside));
        assert_eq!(kind, "invalid_input", "{message}");
        assert!(message.contains("state folder"), "{message}");
        assert!(!inside.exists(), "{inside:?} was made");
        assert_eq!(home.snapshot(), before, "{inside:?}");
    }
    // The ledger itself is still readable afterwards.
    assert_eq!(count(&home, "card"), 2);
}

#[cfg(unix)]
#[test]
fn a_link_in_the_path_is_followed_and_the_real_folder_is_what_is_reported_and_checked() {
    let (home, project) = filled();
    let base = home.path().join("exports");
    let real = base.join("real");
    fs::create_dir_all(&real).unwrap();
    let link = base.join("link");
    std::os::unix::fs::symlink(&real, &link).unwrap();
    let typed = link.join("new");
    let r = result(&export(&home, &project, &typed));
    let real_new = fs::canonicalize(&real).unwrap().join("new");
    assert_eq!(r["out"], json!(real_new.display().to_string()), "{r}");
    assert_eq!(r["out_given"], json!(typed.display().to_string()), "{r}");
    assert_eq!(
        files(&real_new).len(),
        4,
        "the files went to the real folder"
    );

    // A link that leads into the state folder is judged by where it leads.
    let sneaky = base.join("sneaky");
    std::os::unix::fs::symlink(home.path().join(".moai").join("projects"), &sneaky).unwrap();
    let before = home.snapshot();
    let (kind, message) = refusal(&export(&home, &project, &sneaky.join("exp")));
    assert_eq!(kind, "invalid_input", "{message}");
    assert!(message.contains("state folder"), "{message}");
    assert_eq!(home.snapshot(), before);
}

// On a case-insensitive volume (the macOS default) `.MOAI` is the state folder too.
#[cfg(target_os = "macos")]
#[test]
fn another_letter_case_of_the_state_folder_is_still_the_state_folder() {
    let (home, project) = filled();
    let upper = home.path().join(".MOAI");
    if !upper.exists() {
        return; // a case-sensitive volume: `.MOAI` is another folder
    }
    let out = upper.join("projects").join("exp");
    let before = home.snapshot();
    let (kind, message) = refusal(&export(&home, &project, &out));
    assert_eq!(kind, "invalid_input", "{message}");
    assert!(message.contains("state folder"), "{message}");
    assert_eq!(home.snapshot(), before);
}

// ---- the ledger is only read (PR #3126 verdict B3) ----------------------------------------------

/// A leftover rollback journal is what a read-write connection would roll back, which writes to
/// the ledger. A read-only connection must refuse it instead, and export nothing.
#[test]
fn a_ledger_with_a_leftover_journal_is_not_rolled_back_and_nothing_is_exported() {
    let (home, project) = filled();
    let file = ledger_file(&home);
    let mut journal = file.clone().into_os_string();
    journal.push("-journal");
    let journal = PathBuf::from(journal);
    fs::write(&journal, b"left behind").unwrap();
    let (bytes, journal_bytes) = (fs::read(&file).unwrap(), fs::read(&journal).unwrap());
    let out = home.path().join("exports").join("hot");
    fs::create_dir_all(out.parent().unwrap()).unwrap();
    let before = home.snapshot();
    let (_, message) = refusal(&export(&home, &project, &out));
    assert!(message.contains("needs_recovery"), "{message}");
    assert_eq!(fs::read(&file).unwrap(), bytes, "the ledger was written");
    assert_eq!(
        fs::read(&journal).unwrap(),
        journal_bytes,
        "the journal was touched"
    );
    assert!(!out.exists(), "a folder was made");
    assert_eq!(home.snapshot(), before);
}

#[test]
fn a_file_that_is_not_a_ledger_is_refused_unchanged_and_nothing_is_exported() {
    let (home, project) = filled();
    let file = ledger_file(&home);
    fs::write(&file, vec![0x5au8; 5000]).unwrap();
    let out = home.path().join("exports").join("junk");
    fs::create_dir_all(out.parent().unwrap()).unwrap();
    let before = home.snapshot();
    let (_, message) = refusal(&export(&home, &project, &out));
    assert!(!message.is_empty());
    assert_eq!(
        fs::read(&file).unwrap(),
        vec![0x5au8; 5000],
        "the file was written"
    );
    assert!(!out.exists(), "a folder was made");
    assert_eq!(home.snapshot(), before);
}

#[test]
fn the_documents_and_the_help_say_what_the_export_holds() {
    let home = TempHome::new();
    let out = run(&home, &["card", "export", "--help"]);
    // The help wraps lines: compare with the white space folded.
    let text = String::from_utf8_lossy(&out.stdout)
        .split_whitespace()
        .collect::<Vec<_>>()
        .join(" ");
    for needle in [
        "reservation holders (session ids)",
        "lease times",
        "Look at it before you share it",
        "chmod 700",
        "may not be inside `<HOME>/.moai`",
    ] {
        assert!(text.contains(needle), "{needle}: {text}");
    }
}
