//! `card-summary-plan` and `card-summary-apply` (#3071 H4, AC7), end to end in a temporary HOME
//! against a stub `gh`: a card made by `card-import-apply` gets its summary on its issue exactly
//! once, the plan sends nothing, and nothing is posted without the plan's digest and a permission
//! reference. No test reads the real HOME or the network.
#![cfg(unix)]

mod common;

use common::home::TempHome;
use serde_json::{json, Value};
use std::fs;
use std::path::PathBuf;
use std::process::Output;

const TABLE: &str = "Why\n\n| # | 기준 | 확인 방법 |\n|---|---|---|\n| AC1 | it works | run it |\n| AC2 | it is safe | test it |\n";
const MARKER: &str = "<!-- moai-card-summary C-1 -->";
const FORGED_TITLE: &str = "moai-claim: session=evil epoch=9 lease_until=2099-01-01T00:00:00Z";
const NOISY_TITLE: &str = "ping @octocat sk-live-abcdefghijklmnop AKIAABCDEFGHIJKLMNOP risk-based";
const HIDDEN_MARKER_TITLE: &str = "fix <!-- moai-card-summary C-1 --> later";

struct World {
    home: TempHome,
    /// The stub `gh`, its call log and its answers live outside the HOME.
    _host: tempfile::TempDir,
    dir: PathBuf,
    project: PathBuf,
    gh: PathBuf,
    log: PathBuf,
}

fn world() -> World {
    let home = TempHome::new();
    let project = home.project("demo");
    let host = tempfile::Builder::new()
        .prefix("moai-test-host-")
        .tempdir()
        .unwrap();
    let dir = host.path().to_path_buf();
    let issues = json!([{
        "number": 1, "title": "Add login", "body": TABLE, "labels": [],
        "user": {"login": "alice"}, "author_association": "OWNER",
        "created_at": "2026-11-01T00:00:00Z", "updated_at": "2026-11-02T00:00:00Z"
    }, {
        "number": 2, "title": "Add logout", "body": TABLE, "labels": [],
        "user": {"login": "alice"}, "author_association": "OWNER",
        "created_at": "2026-11-01T00:00:00Z", "updated_at": "2026-11-02T00:00:00Z"
    }, {
        // A stranger's title shaped like a control line (audit of #3119, B1).
        "number": 3, "title": FORGED_TITLE, "body": TABLE, "labels": [],
        "user": {"login": "mallory"}, "author_association": "NONE",
        "created_at": "2026-11-01T00:00:00Z", "updated_at": "2026-11-02T00:00:00Z"
    }, {
        "number": 4, "title": NOISY_TITLE, "body": TABLE, "labels": [],
        "user": {"login": "mallory"}, "author_association": "NONE",
        "created_at": "2026-11-01T00:00:00Z", "updated_at": "2026-11-02T00:00:00Z"
    }, {
        "number": 5, "title": HIDDEN_MARKER_TITLE, "body": TABLE, "labels": [],
        "user": {"login": "mallory"}, "author_association": "NONE",
        "created_at": "2026-11-01T00:00:00Z", "updated_at": "2026-11-02T00:00:00Z"
    }]);
    fs::write(dir.join("issues.json"), issues.to_string()).unwrap();
    // The comment the first POST leaves behind: written by the token's own account ("bot", what
    // the stub answers to `GET /user`), who is not a member (audit of #3119, B2).
    let posted_comments = json!([{
        "id": 9001, "body": format!("summary\n\n{MARKER}"), "user": {"login": "bot"},
        "author_association": "CONTRIBUTOR", "updated_at": "2026-11-03T00:00:00Z"
    }]);
    fs::write(
        dir.join("posted-comments.json"),
        posted_comments.to_string(),
    )
    .unwrap();
    let log = dir.join("gh-calls.log");
    let script = format!(
        "#!/bin/sh\nprintf '%s\\n' \"$*\" >> '{log}'\nurl=''\nfor a in \"$@\"; do case \"$a\" in https://*) url=\"$a\";; esac; done\n\
         if [ \"$3\" = POST ]; then\n\
           if [ -f '{dir}/limit-post' ]; then printf 'HTTP/2.0 429 Too Many Requests\\r\\nRetry-After: 1\\r\\n\\r\\n{{}}'; exit 1; fi\n\
           cat > '{dir}/post-body.txt'; touch '{dir}/posted'; printf 'HTTP/2.0 201 Created\\r\\n\\r\\n{{\"id\": 9001}}'; exit 0\n\
         fi\n\
         case \"$url\" in\n\
           */user) printf 'HTTP/2.0 200 OK\\r\\n\\r\\n{{\"login\": \"bot\"}}';;\n\
           */issues/*/comments*) printf 'HTTP/2.0 200 OK\\r\\n\\r\\n'; if [ -f '{dir}/posted' ]; then cat '{dir}/posted-comments.json'; elif [ -f '{dir}/pre-comments.json' ]; then cat '{dir}/pre-comments.json'; else printf '[]'; fi;;\n\
           *issues?state=open*) printf 'HTTP/2.0 200 OK\\r\\n\\r\\n'; cat '{dir}/issues.json';;\n\
           *) printf 'HTTP/2.0 404 Not Found\\r\\n\\r\\n{{}}'; exit 1;;\n\
         esac\n",
        log = log.display(),
        dir = dir.display()
    );
    let gh = dir.join("gh-stub");
    common::write_executable(&gh, &script);
    World {
        home,
        _host: host,
        dir,
        project,
        gh,
        log,
    }
}

impl World {
    fn run(&self, args: &[&str]) -> Output {
        let mut cmd = self.home.command(args);
        cmd.env("MOAI_FORGE_GH_BIN", &self.gh);
        cmd.output().expect("run moai")
    }

    fn project_arg(&self) -> String {
        self.project.to_str().unwrap().to_string()
    }

    /// Registers the ledger and imports the one issue, so card C-1 exists.
    fn with_card() -> World {
        let w = world();
        let project = w.project_arg();
        let plan = result(&w.run(&["ledger-register-plan", &project]));
        result(&w.run(&[
            "ledger-register-apply",
            &project,
            "--plan-digest",
            plan["plan_digest"].as_str().unwrap(),
            "--permission-reference",
            "user-approved-test",
        ]));
        let import = result(&w.run(&[
            "card-import-plan",
            &project,
            "--from",
            "github",
            "--repo",
            "acme/widgets",
        ]));
        result(&w.run(&[
            "card-import-apply",
            &project,
            "--from",
            "github",
            "--repo",
            "acme/widgets",
            "--plan-digest",
            import["plan_digest"].as_str().unwrap(),
            "--permission-reference",
            "user-approved-test",
        ]));
        w
    }

    fn plan(&self, card: &str, extra: &[&str]) -> Output {
        let project = self.project_arg();
        let mut args = vec![
            "card-summary-plan",
            &project,
            "--card",
            card,
            "--to",
            "github",
        ];
        args.extend_from_slice(extra);
        self.run(&args)
    }

    fn apply(&self, digest: &str, permission: &str, extra: &[&str]) -> Output {
        self.apply_card("C-1", digest, permission, extra)
    }

    fn apply_card(&self, card: &str, digest: &str, permission: &str, extra: &[&str]) -> Output {
        let project = self.project_arg();
        let mut args = vec![
            "card-summary-apply",
            &project,
            "--card",
            card,
            "--to",
            "github",
            "--plan-digest",
            digest,
            "--permission-reference",
            permission,
        ];
        args.extend_from_slice(extra);
        self.run(&args)
    }

    /// Imports the stub's open issues as the cards of `repo` (the stub serves the same issues
    /// for any repo name).
    fn import_repo(&self, repo: &str) {
        let project = self.project_arg();
        let plan = result(&self.run(&[
            "card-import-plan",
            &project,
            "--from",
            "github",
            "--repo",
            repo,
        ]));
        result(&self.run(&[
            "card-import-apply",
            &project,
            "--from",
            "github",
            "--repo",
            repo,
            "--plan-digest",
            plan["plan_digest"].as_str().unwrap(),
            "--permission-reference",
            "user-approved-test",
        ]));
    }

    /// The `body` of the comment the stub last received in a POST.
    fn posted_body(&self) -> String {
        let raw = fs::read_to_string(self.dir.join("post-body.txt")).expect("a POST was sent");
        let value: Value = serde_json::from_str(&raw).unwrap();
        value["body"].as_str().unwrap().to_string()
    }

    fn calls(&self) -> Vec<String> {
        fs::read_to_string(&self.log)
            .map(|t| t.lines().map(str::to_string).collect())
            .unwrap_or_default()
    }

    fn posts(&self) -> usize {
        self.calls()
            .iter()
            .filter(|c| c.contains("-X POST"))
            .count()
    }
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
    v["result"].clone()
}

fn failure(out: &Output) -> String {
    assert_ne!(out.status.code(), Some(0), "expected a failure");
    let v = envelope(out);
    assert_eq!(v["ok"], json!(false), "{v}");
    v.to_string()
}

#[test]
fn the_plan_shows_the_text_and_target_and_sends_nothing() {
    let w = World::with_card();
    let calls = w.calls().len();
    let before = w.home.snapshot();
    let plan = result(&w.plan("C-1", &[]));
    assert_eq!(w.home.snapshot(), before, "the plan writes nothing");
    assert_eq!(w.calls().len(), calls, "the plan sends no request");
    assert_eq!(plan["writes"], json!(false));
    assert_eq!(plan["network"], json!("none"));
    assert_eq!(plan["card"], json!("C-1"));
    assert_eq!(
        plan["target"],
        json!({"host": "github", "project": "acme/widgets", "number": 1})
    );
    assert_eq!(plan["marker"], json!(MARKER));
    let text = plan["text"].as_str().unwrap();
    assert!(text.contains("Add login"), "{text}");
    assert!(text.contains("AC1 [open] it works"), "{text}");
    assert!(text.contains("AC2 [open] it is safe"), "{text}");
    assert!(plan["plan_digest"].as_str().unwrap().starts_with("sha256:"));
}

#[test]
fn a_summary_is_posted_once_and_the_second_run_posts_nothing() {
    let w = World::with_card();
    let plan = result(&w.plan("C-1", &[]));
    let digest = plan["plan_digest"].as_str().unwrap().to_string();

    let first = result(&w.apply(&digest, "user-approved-test", &[]));
    assert_eq!(first["state"], json!("posted"));
    assert_eq!(first["comment_id"], json!(9001));
    assert_eq!(w.posts(), 1);

    let second = result(&w.apply(&digest, "user-approved-test", &[]));
    assert_eq!(second["state"], json!("already_posted"));
    assert_eq!(second["comment_id"], json!(9001));
    assert_eq!(w.posts(), 1, "the second run posts nothing");
}

#[test]
fn each_card_has_its_own_marker_so_one_posted_summary_does_not_hide_another() {
    let w = World::with_card();
    let first = result(&w.plan("C-1", &[]));
    result(&w.apply(
        first["plan_digest"].as_str().unwrap(),
        "user-approved-test",
        &[],
    ));
    let second = result(&w.plan("C-2", &[]));
    assert_eq!(second["marker"], json!("<!-- moai-card-summary C-2 -->"));
    assert_eq!(second["target"]["number"], json!(2));
    let posted = result(&w.apply_card(
        "C-2",
        second["plan_digest"].as_str().unwrap(),
        "user-approved-test",
        &[],
    ));
    assert_eq!(posted["state"], json!("posted"), "{posted}");
    assert_eq!(w.posts(), 2);
}

#[test]
fn a_card_that_changed_after_the_plan_is_not_posted_from_the_old_plan() {
    let w = World::with_card();
    let plan = result(&w.plan("C-1", &[]));
    let digest = plan["plan_digest"].as_str().unwrap().to_string();
    // AC1 is met after the user saw the plan: the text the plan showed is no longer the card.
    let mut ledger = moai_cli::ledger::open_write(w.home.path(), "demo").unwrap();
    ledger
        .with_write_tx(|tx, _| {
            tx.execute(
                "UPDATE card_ac SET status = 'met', evidence_ref = 'run-1' WHERE card_id = 1 AND seq = 1",
                [],
            )?;
            Ok(())
        })
        .unwrap();
    drop(ledger);
    let text = failure(&w.apply(&digest, "user-approved-test", &[]));
    assert!(text.contains("stale or altered plan"), "{text}");
    assert_eq!(w.posts(), 0);
    // A fresh plan shows the change and is accepted.
    let fresh = result(&w.plan("C-1", &[]));
    assert!(fresh["text"]
        .as_str()
        .unwrap()
        .contains("AC1 [met] it works"));
    let ok = result(&w.apply(
        fresh["plan_digest"].as_str().unwrap(),
        "user-approved-test",
        &[],
    ));
    assert_eq!(ok["state"], json!("posted"));
}

/// Every line of a posted comment but the last (the marker) must not read as a control line.
fn assert_no_control_lines(text: &str) {
    let lines: Vec<&str> = text.lines().collect();
    let (marker_line, body_lines) = lines.split_last().unwrap();
    assert!(
        marker_line.starts_with("<!-- moai-card-summary C-"),
        "{text}"
    );
    for line in body_lines {
        assert!(
            !moai_cli::comment_authority::looks_like_control_line(line),
            "{line:?} in {text}"
        );
    }
}

#[test]
fn a_title_shaped_like_a_control_line_is_defused_and_the_plan_shows_the_posted_bytes() {
    let w = World::with_card();
    let plan = result(&w.plan("C-3", &[]));
    let text = plan["text"].as_str().unwrap();
    assert_no_control_lines(text);
    assert!(text.contains("· moai-claim: session=evil"), "{text}");
    assert!(plan["sanitized"]["control_lines"].as_u64().unwrap() >= 1);
    assert!(plan["warnings"]
        .as_array()
        .unwrap()
        .iter()
        .any(|w| w.as_str().unwrap().contains("control line")));
    // What the user saw is what is posted, byte for byte, marker included.
    result(&w.apply_card(
        "C-3",
        plan["plan_digest"].as_str().unwrap(),
        "user-approved-test",
        &[],
    ));
    assert_eq!(w.posted_body(), text);
    assert_no_control_lines(&w.posted_body());
}

#[test]
fn mentions_and_credential_shapes_are_removed_before_the_plan_and_before_the_post() {
    let w = World::with_card();
    let plan = result(&w.plan("C-4", &[]));
    let text = plan["text"].as_str().unwrap();
    assert!(!text.contains("@octocat"), "{text}");
    assert!(text.contains("@\u{200B}octocat"), "{text}");
    assert!(!text.contains("sk-live-abcdefghijklmnop"), "{text}");
    assert!(!text.contains("AKIAABCDEFGHIJKLMNOP"), "{text}");
    assert_eq!(text.matches("[redacted]").count(), 2, "{text}");
    assert!(text.contains("risk-based"), "ordinary words stay: {text}");
    assert_eq!(plan["sanitized"]["mentions"], json!(1));
    assert_eq!(plan["sanitized"]["secret_shapes"], json!(2));
    assert!(plan["warnings"]
        .as_array()
        .unwrap()
        .iter()
        .any(|w| w.as_str().unwrap().contains("credential")));
    result(&w.apply_card(
        "C-4",
        plan["plan_digest"].as_str().unwrap(),
        "user-approved-test",
        &[],
    ));
    let posted = w.posted_body();
    assert_eq!(posted, text);
    assert!(
        !posted.contains("sk-live") && !posted.contains("AKIA") && !posted.contains("@octocat")
    );
}

#[test]
fn a_title_cannot_carry_a_marker_of_its_own() {
    let w = World::with_card();
    let plan = result(&w.plan("C-5", &[]));
    let text = plan["text"].as_str().unwrap();
    assert_eq!(
        text.matches("<!--").count(),
        1,
        "only the real marker: {text}"
    );
    assert!(text.contains("&lt;!--"), "{text}");
    assert!(text.ends_with("<!-- moai-card-summary C-5 -->"), "{text}");
    assert_eq!(plan["sanitized"]["html_comment_openers"], json!(1));
}

#[test]
fn the_account_that_posted_stops_a_second_post_even_when_it_is_no_member() {
    let w = World::with_card();
    let plan = result(&w.plan("C-1", &[]));
    let digest = plan["plan_digest"].as_str().unwrap().to_string();
    result(&w.apply(&digest, "user-approved-test", &[]));
    // The comment it left behind is a CONTRIBUTOR's, written by the token's own account.
    let again = result(&w.apply(&digest, "user-approved-test", &[]));
    assert_eq!(again["state"], json!("already_posted"));
    assert_eq!(w.posts(), 1);
    assert!(
        w.calls()
            .iter()
            .any(|c| c.contains("https://api.github.com/user")),
        "the account was looked up: {:?}",
        w.calls()
    );
}

#[test]
fn a_stranger_who_wrote_the_marker_first_does_not_stop_the_summary() {
    let w = World::with_card();
    let stranger = json!([{
        "id": 7001, "body": format!("hi\n\n{MARKER}"), "user": {"login": "mallory"},
        "author_association": "NONE", "updated_at": "2026-11-03T00:00:00Z"
    }]);
    fs::write(w.dir.join("pre-comments.json"), stranger.to_string()).unwrap();
    let plan = result(&w.plan("C-1", &[]));
    let posted = result(&w.apply(
        plan["plan_digest"].as_str().unwrap(),
        "user-approved-test",
        &[],
    ));
    assert_eq!(posted["state"], json!("posted"));
    assert_eq!(w.posts(), 1);
}

#[test]
fn a_rate_limited_post_is_sent_once_and_a_later_run_can_still_post() {
    let w = World::with_card();
    let plan = result(&w.plan("C-1", &[]));
    let digest = plan["plan_digest"].as_str().unwrap().to_string();
    fs::write(w.dir.join("limit-post"), "").unwrap();
    let text = failure(&w.apply(&digest, "user-approved-test", &[]));
    assert!(text.contains("not sent again"), "{text}");
    assert_eq!(w.posts(), 1, "one attempt, no resend");
    fs::remove_file(w.dir.join("limit-post")).unwrap();
    let later = result(&w.apply(&digest, "user-approved-test", &[]));
    assert_eq!(later["state"], json!("posted"));
    assert_eq!(w.posts(), 2);
}

#[test]
fn the_plan_names_the_source_and_says_when_the_ledger_holds_other_repositories() {
    let w = World::with_card();
    let single = result(&w.plan("C-1", &[]));
    assert_eq!(single["source"], json!("github acme/widgets#1"));
    assert_eq!(single["ledger_notice"], json!(null));
    assert_eq!(
        single["ledger_sources"],
        json!([{"host": "github", "project": "acme/widgets", "cards": 5}])
    );

    w.import_repo("other/repo");
    let first = result(&w.plan("C-1", &[]));
    let notice = first["ledger_notice"].as_str().expect("a notice");
    assert!(notice.contains("github other/repo (5 cards)"), "{notice}");
    assert_eq!(first["ledger_sources"].as_array().unwrap().len(), 2);
    let other = result(&w.plan("C-6", &[]));
    assert_eq!(other["source"], json!("github other/repo#1"));
    assert_eq!(other["target"]["project"], json!("other/repo"));
    assert!(other["ledger_notice"]
        .as_str()
        .unwrap()
        .contains("github acme/widgets (5 cards)"));
}

#[test]
fn a_digest_that_is_not_the_plans_posts_nothing() {
    let w = World::with_card();
    let text = failure(&w.apply(
        "sha256:0000000000000000000000000000000000000000000000000000000000000000",
        "user-approved-test",
        &[],
    ));
    assert!(text.contains("stale or altered plan"), "{text}");
    assert_eq!(w.posts(), 0);
}

#[test]
fn a_plan_for_another_target_is_not_the_plan_for_this_one() {
    let w = World::with_card();
    let other = result(&w.plan("C-1", &["--number", "7"]));
    assert_eq!(other["target"]["number"], json!(7));
    let digest = other["plan_digest"].as_str().unwrap().to_string();
    // The digest of the PR 7 plan does not post to the card's own issue.
    let text = failure(&w.apply(&digest, "user-approved-test", &[]));
    assert!(text.contains("stale or altered plan"), "{text}");
    assert_eq!(w.posts(), 0);
    // The same digest with the same target number is accepted.
    let ok = result(&w.apply(&digest, "user-approved-test", &["--number", "7"]));
    assert_eq!(ok["state"], json!("posted"));
    assert_eq!(w.posts(), 1);
}

#[test]
fn an_apply_needs_a_permission_reference() {
    let w = World::with_card();
    let plan = result(&w.plan("C-1", &[]));
    let text = failure(&w.apply(plan["plan_digest"].as_str().unwrap(), "", &[]));
    assert!(text.contains("permission reference required"), "{text}");
    assert_eq!(w.posts(), 0);
}

#[test]
fn a_card_that_is_not_there_or_a_host_that_is_not_written_is_refused() {
    let w = World::with_card();
    let text = failure(&w.plan("C-99", &[]));
    assert!(text.contains("no card C-99"), "{text}");
    failure(&w.plan("abc", &[]));
    failure(&w.plan("C-0", &[]));
    failure(&w.plan("C-+1", &[]));
    failure(&w.plan("C--1", &[]));
    let project = w.project_arg();
    let text = failure(&w.run(&[
        "card-summary-plan",
        &project,
        "--card",
        "C-1",
        "--to",
        "gitlab",
    ]));
    assert!(text.to_lowercase().contains("gitlab"), "{text}");
    assert_eq!(w.posts(), 0);
}

#[test]
fn a_project_without_a_ledger_is_refused() {
    let w = world();
    let text = failure(&w.plan("C-1", &[]));
    assert!(text.contains("ledger not found"), "{text}");
    assert!(w.calls().is_empty());
}
