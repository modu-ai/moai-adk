//! SPEC-20261119-card-precheck (card C-96, workflow W3): `moai card precheck` is the read-only
//! check the leader runs before readying or dispatching a card. It lists the merged commits and
//! pull requests whose message names the card (`C-<n>`, `card/<n>`, `card <n>`, or a `Moai-Spec`
//! trailer naming the card's linked SPEC), the SPEC linked to the card and the SPEC folders that
//! name it, and every decision note that shares a path with the card's write scope or a phrase
//! from its criteria — so a card already done on the tip (C-88 #3151, C-92 #3187) or a criterion
//! that contradicts a recorded decision (C-63 N1, C-87 AC 2) is seen before a lane starts.
//! Nothing is written: the ledger's bytes and mtimes and the repository's HEAD are unchanged.
//! Every case runs the real binary against a temporary HOME ledger and a temporary repository.

mod common;

use common::git::{git, git_in};
use common::home::TempHome;
use moai_cli::ledger;
use serde_json::{json, Value};
use std::fs;
use std::path::PathBuf;
use std::process::Output;

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

/// A temporary HOME with a registered ledger and a git repository on `main` whose first commit
/// is `fixture`.
struct World {
    home: TempHome,
    project: PathBuf,
}

impl World {
    fn new() -> Self {
        let home = TempHome::new();
        let project = home.project("demo");
        ledger::register(home.path(), "demo").unwrap();
        git_in(&project, &["init", "-q", "-b", "main"]);
        git_in(
            &project,
            &["commit", "--allow-empty", "-q", "-m", "fixture"],
        );
        World { home, project }
    }

    fn p(&self) -> &str {
        self.project.to_str().unwrap()
    }

    fn run(&self, args: &[&str]) -> Output {
        self.home.command(args).output().expect("run moai")
    }

    /// `moai card <args> --project P`.
    fn card(&self, args: &[&str]) -> Output {
        let mut all = vec!["card"];
        all.extend_from_slice(args);
        all.extend_from_slice(&["--project", self.p()]);
        self.run(&all)
    }

    /// `moai card precheck <card> --integration-ref main`.
    fn precheck(&self, card: &str) -> Output {
        self.card(&["precheck", card, "--integration-ref", "main"])
    }

    fn head(&self) -> String {
        git(&self.project, &["rev-parse", "HEAD"])
    }

    /// One empty commit whose message is `subject` and then each paragraph, in order.
    fn commit(&self, subject: &str, paragraphs: &[&str]) {
        let mut args = vec!["commit", "--allow-empty", "-q", "-m", subject];
        for p in paragraphs {
            args.extend_from_slice(&["-m", p]);
        }
        git_in(&self.project, &args);
    }

    /// A real merge commit — a side branch with one commit of its own, merged with `--no-ff` —
    /// whose message is `subject` and then each paragraph.
    fn merge(&self, subject: &str, paragraphs: &[&str]) {
        git_in(&self.project, &["checkout", "-q", "-b", "side"]);
        git_in(
            &self.project,
            &["commit", "--allow-empty", "-q", "-m", "side work"],
        );
        git_in(&self.project, &["checkout", "-q", "main"]);
        let mut args = vec!["merge", "--no-ff", "-q", "--no-edit", "-m", subject];
        for p in paragraphs {
            args.extend_from_slice(&["-m", p]);
        }
        args.push("side");
        git_in(&self.project, &args);
    }

    /// Adds a `feature` card with `scope` and `criteria`; returns its name (`C-1`).
    fn card_with(&self, scope: &[&str], criteria: &[&str]) -> String {
        let mut args = vec!["add", "--title", "precheck demo", "--kind", "feature"];
        for s in scope {
            args.extend_from_slice(&["--scope", s]);
        }
        for c in criteria {
            args.extend_from_slice(&["--ac", c]);
        }
        let added = ok(&self.card(&args));
        added["name"].as_str().unwrap().to_string()
    }

    /// A note on `card_id` of `kind` with `body` and, when `ref` is given, that ref; returns the
    /// note id. Written straight into the ledger, the way note fixtures are built elsewhere.
    fn note(&self, card_id: i64, kind: &str, body: &str, ref_: Option<&str>) -> i64 {
        {
            let mut l = ledger::open_write(self.home.path(), "demo").unwrap();
            l.with_write_tx(|tx, now| {
                tx.execute(
                    "INSERT INTO card_note (card_id, kind, author, origin, body, ref, created_at) \
                     VALUES (?1, ?2, 'leader', 'local', ?3, ?4, ?5)",
                    rusqlite::params![card_id, kind, body, ref_, now],
                )?;
                Ok(())
            })
            .unwrap();
        }
        let l = ledger::open_read_only(self.home.path(), "demo").unwrap();
        l.conn()
            .query_row("SELECT max(note_id) FROM card_note", [], |r| r.get(0))
            .unwrap()
    }

    /// A minimal SPEC that passes `moai spec check`, through the real commands.
    fn draft_spec(&self, slug: &str) -> String {
        let r = ok(&self.run(&[
            "spec",
            "new",
            slug,
            "--title",
            "t",
            "--minimal",
            "--date",
            "2026-11-19",
            "--project",
            self.p(),
        ]));
        let id = r["id"].as_str().unwrap().to_string();
        let path = self.project.join(".moai/specs").join(&id).join("spec.md");
        let text = fs::read_to_string(&path).unwrap();
        fs::write(&path, filled(&text)).unwrap();
        id
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

/// REQ-1: the commits, PRs, linked SPEC and SPEC folders that name the card — and the noise
/// (`C-10`, `C-19`, `card/10`, `AC-100`) that must stay out. REQ-5: nothing is written.
#[test]
fn lists_merged_prs_commits_and_spec_folders_naming_the_card() {
    let w = World::new();
    let linked = w.draft_spec("demo-linked");
    let card = w.card_with(
        &["src/alpha.rs", "tests/beta/**"],
        &[
            "N1: the hooks offer reads ready when #2936 names it",
            "N2: user_chat lines follow Lang::En",
        ],
    );
    assert_eq!(card, "C-1");
    ok(&w.card(&["spec", &card, &linked, "--by", "leader"]));

    // A SPEC folder whose text names the card (the folder scan reads the text only).
    let two = w.project.join(".moai/specs/SPEC-20261119-demo-two/spec.md");
    fs::create_dir_all(two.parent().unwrap()).unwrap();
    fs::write(&two, "# two\n\nThe card C-1 names this SPEC.\n").unwrap();

    // Five commits name the card, three are noise; the merge brings one side commit of its own.
    w.commit("fix(demo): noise for another card (C-10)", &[]);
    w.commit(
        "chore(demo): unrelated subject",
        &["Notes: C-1 covered here."],
    );
    w.merge(
        "Merge pull request #31 from modu-ai/card/1-demo",
        &["feat(demo): the PR title (card 1)"],
    );
    w.commit(
        "chore(demo): trailer one",
        &["Moai-Spec: SPEC-20261119-demo-linked"],
    );
    w.commit("fix(demo): lowercase (c-1) too", &[]);
    w.commit("fix(demo): the thing lands (C-1)", &[]);
    w.commit("chore: AC-100 cleanup", &["see C-19 and card/10"]);

    let before = w.home.snapshot();
    let head_before = w.head();
    let v = ok(&w.precheck(&card));
    assert_eq!(v["action"], json!("precheck"));
    assert_eq!(v["card"], json!("C-1"));
    assert_eq!(v["ledger"], json!("ready"));
    assert_eq!(v["git"]["ref"], json!("main"));
    assert_eq!(v["git"]["head"], json!(head_before));
    assert_eq!(v["git"]["scanned"], json!(9));
    assert_eq!(v["git"]["truncated"], json!(false));

    let m = v["merged_mentions"].as_array().unwrap();
    let rows: Vec<String> = m
        .iter()
        .map(|e| format!("{}|{}|{}|{}", e["kind"], e["pr"], e["where"], e["matched"]))
        .collect();
    assert_eq!(
        rows,
        vec![
            "commit|null|subject|C-1".to_string(),
            "commit|null|subject|C-1".to_string(),
            "commit|null|trailer|spec:SPEC-20261119-demo-linked".to_string(),
            "pr_merge|31|subject|card/1".to_string(),
            "commit|null|body|C-1".to_string(),
        ],
        "{m:?}"
    );

    assert_eq!(v["spec"]["linked"], json!("SPEC-20261119-demo-linked"));
    assert_eq!(
        v["spec"]["folders_naming_the_card"],
        json!(["SPEC-20261119-demo-two"])
    );
    assert_eq!(v["decision_notes"], json!([]));

    // REQ-5: the HOME tree and the repository's HEAD are exactly as they were.
    assert_eq!(w.home.snapshot(), before);
    assert_eq!(w.head(), head_before);
}

/// REQ-3: decision notes that share a scope path or a criteria phrase are listed with their
/// card, note ref, author and every key that matched; a note of another kind and an unrelated
/// decision stay out.
#[test]
fn lists_decision_notes_sharing_a_path_or_a_criteria_phrase() {
    let w = World::new();
    let card = w.card_with(
        &["src/alpha.rs", "tests/beta/**"],
        &[
            "N1: the hooks offer reads ready when #2936 names it",
            "N2: user_chat lines follow Lang::En",
        ],
    );
    let other = w.card_with(&["src/elsewhere.rs"], &["N9: elsewhere only"]);
    assert_eq!((card.as_str(), other.as_str()), ("C-1", "C-2"));

    let n1 = w.note(
        1,
        "decision",
        "DECISION leader — src/alpha.rs 줄은 C-77 이 관리해요",
        None,
    );
    let n2 = w.note(
        2,
        "decision",
        "DECISION user — #2936 S-CYCLE AC 13: hooks 제안이 없어요",
        Some("note:2:deadbeef"),
    );
    let _n3 = w.note(
        2,
        "decision",
        "DECISION leader — 배포 스크립트는 deploy/ 안에만 써요",
        None,
    );
    let _n4 = w.note(
        2,
        "comment",
        "LANE-DONE — 같은 경로 src/alpha.rs 를 언급하지만 다른 종류",
        None,
    );

    let v = ok(&w.precheck(&card));
    let notes = v["decision_notes"].as_array().unwrap();
    assert_eq!(notes.len(), 2, "{notes:?}");
    // Sorted by card, then note id: the card's own note first.
    assert_eq!(notes[0]["card"], json!("C-1"));
    assert_eq!(notes[0]["note_id"], json!(n1));
    assert_eq!(notes[0]["ref"], Value::Null);
    assert_eq!(notes[0]["author"], json!("leader"));
    assert_eq!(notes[0]["matched_by"], json!(["path:src/alpha.rs"]));
    assert_eq!(notes[1]["card"], json!("C-2"));
    assert_eq!(notes[1]["note_id"], json!(n2));
    assert_eq!(notes[1]["ref"], json!("note:2:deadbeef"));
    assert_eq!(
        notes[1]["matched_by"],
        json!(["phrase:hooks", "phrase:#2936"])
    );
}

/// REQ-2, REQ-4, REQ-6: a bad ref and an absent ledger are advisories inside an `ok` report
/// (the git half still runs; the decision half stays empty), a malformed name and a missing
/// card are refusals like the other read-only actions, and `--max-commits` bounds the scan.
#[test]
fn advisory_on_failures_refuses_bad_names_and_bounds_the_scan() {
    let w = World::new();
    let card = w.card_with(&["src/alpha.rs"], &["N1: the hooks offer reads ready"]);
    w.commit("fix(demo): the thing lands (C-1)", &[]);
    w.commit("chore: noise one", &[]);
    w.commit("chore: noise two", &[]);
    w.commit("chore: noise three", &[]);

    // A malformed name and a missing card refuse like `card show` does.
    let (kind, message) = refusal(&w.card(&["precheck", "hello", "--integration-ref", "main"]));
    assert_eq!(kind, "invalid_input");
    assert!(message.contains("not a card"), "{message}");
    let (kind, message) = refusal(&w.card(&["precheck", "C-99", "--integration-ref", "main"]));
    assert_eq!(kind, "not_found");
    assert!(message.contains("C-99"), "{message}");

    // A ref that is not there: an advisory inside an ok report, no mentions.
    let v = ok(&w.card(&["precheck", &card, "--integration-ref", "no-such-ref"]));
    assert_eq!(v["ledger"], json!("ready"));
    assert!(!v["git"]["error"].as_str().unwrap().is_empty());
    assert_eq!(v["merged_mentions"], json!([]));

    // A project with a repository but no ledger: the git half still answers.
    let bare = TempHome::new();
    let project = bare.project("demo");
    git_in(&project, &["init", "-q", "-b", "main"]);
    git_in(
        &project,
        &["commit", "--allow-empty", "-q", "-m", "feat: (C-1) lands"],
    );
    let out = bare
        .command(&[
            "card",
            "precheck",
            "C-1",
            "--integration-ref",
            "main",
            "--project",
            project.to_str().unwrap(),
        ])
        .output()
        .expect("run moai");
    let v = ok(&out);
    assert_eq!(v["ledger"], json!("absent"));
    assert_eq!(v["spec"]["linked"], Value::Null);
    assert_eq!(v["decision_notes"], json!([]));
    assert_eq!(v["merged_mentions"].as_array().unwrap().len(), 1);

    // The scan bound: the newest three commits only, and the bound is named.
    let v = ok(&w.card(&[
        "precheck",
        &card,
        "--integration-ref",
        "main",
        "--max-commits",
        "3",
    ]));
    assert_eq!(v["git"]["scanned"], json!(3));
    assert_eq!(v["git"]["truncated"], json!(true));
    assert_eq!(v["merged_mentions"].as_array().unwrap().len(), 0);
}
