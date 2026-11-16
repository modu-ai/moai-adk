//! `card-import-plan` and `card-import-apply` (#3071 H3b), end to end in a temporary HOME against
//! a stub `gh` that serves two open issues and one pull request from files: the preview writes
//! nothing, the apply needs the ledger and the preview's digest, a second apply adds nothing,
//! and only GET requests are sent. No test reads the real HOME or the network.
#![cfg(unix)]

mod common;

use common::home::TempHome;
use serde_json::{json, Value};
use std::fs;
use std::path::PathBuf;
use std::process::Output;

const TABLE: &str = "Why\n\n| # | 기준 | 확인 방법 |\n|---|---|---|\n| AC1 | it works | run it |\n| AC2 | it is safe | test it |\n";

struct World {
    home: TempHome,
    /// The stub `gh` and its call log live outside the HOME, so the HOME snapshots stay exact.
    _host: tempfile::TempDir,
    project: PathBuf,
    gh: PathBuf,
    log: PathBuf,
}

fn issue(number: u64, title: &str, body: &str) -> Value {
    json!({"number": number, "title": title, "body": body, "labels": [],
           "user": {"login": "alice"}, "author_association": "OWNER",
           "created_at": "2026-11-01T00:00:00Z", "updated_at": "2026-11-02T00:00:00Z"})
}

fn comment(id: u64, body: &str, association: &str) -> Value {
    json!({"id": id, "body": body, "user": {"login": "carol"},
           "author_association": association, "updated_at": "2026-11-03T00:00:00Z"})
}

fn world() -> World {
    let home = TempHome::new();
    let project = home.project("demo");
    let host = tempfile::Builder::new()
        .prefix("moai-test-host-")
        .tempdir()
        .unwrap();
    let dir = host.path().to_path_buf();
    let issues = json!([
        issue(1, "Add login", TABLE),
        issue(2, "No table here", "just words"),
        {"number": 3, "title": "A pull request", "pull_request": {}, "labels": [],
         "user": {"login": "bob"}, "author_association": "OWNER"}
    ]);
    fs::write(dir.join("issues.json"), issues.to_string()).unwrap();
    fs::write(
        dir.join("comments-1.json"),
        json!([
            comment(501, "a member note", "MEMBER"),
            comment(502, "a stranger note", "NONE")
        ])
        .to_string(),
    )
    .unwrap();
    fs::write(dir.join("comments-2.json"), "[]").unwrap();
    let log = dir.join("gh-calls.log");
    let script = format!(
        "#!/bin/sh\nprintf '%s\\n' \"$*\" >> '{log}'\nfor a in \"$@\"; do last=\"$a\"; done\n\
         case \"$last\" in\n\
           */issues/*/comments*) n=${{last#*/issues/}}; n=${{n%%/*}}; printf 'HTTP/2.0 200 OK\\r\\n\\r\\n'; cat '{dir}/comments-'$n'.json';;\n\
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

    fn plan(&self) -> Value {
        let project = self.project_arg();
        result(&self.run(&[
            "card-import-plan",
            &project,
            "--from",
            "github",
            "--repo",
            "acme/widgets",
        ]))
    }

    fn apply(&self, digest: &str) -> Output {
        let project = self.project_arg();
        self.run(&[
            "card-import-apply",
            &project,
            "--from",
            "github",
            "--repo",
            "acme/widgets",
            "--plan-digest",
            digest,
            "--permission-reference",
            "user-approved-test",
        ])
    }

    fn register(&self) {
        let project = self.project_arg();
        let plan = result(&self.run(&["ledger-register-plan", &project]));
        let digest = plan["plan_digest"].as_str().unwrap().to_string();
        result(&self.run(&[
            "ledger-register-apply",
            &project,
            "--plan-digest",
            &digest,
            "--permission-reference",
            "user-approved-test",
        ]));
    }

    fn cards_total(&self) -> u64 {
        let project = self.project_arg();
        let status = result(&self.run(&["ledger-status", &project]));
        status["cards"]["total"].as_u64().unwrap()
    }

    fn calls(&self) -> Vec<String> {
        fs::read_to_string(&self.log)
            .map(|t| t.lines().map(str::to_string).collect())
            .unwrap_or_default()
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

fn only_gets(calls: &[String]) {
    for call in calls {
        assert!(call.starts_with("api -X GET "), "{call}");
    }
}

#[test]
fn the_preview_lists_the_cards_and_writes_nothing() {
    let w = world();
    let before = w.home.snapshot();
    let plan = w.plan();
    assert_eq!(
        w.home.snapshot(),
        before,
        "no folder, file or ledger is made"
    );
    assert_eq!(plan["writes"], json!(false));
    assert_eq!(plan["project_id"], json!("demo"));
    assert_eq!(
        plan["source"],
        json!({"host": "github", "project": "acme/widgets"})
    );
    assert_eq!(plan["card_count"], json!(2));
    assert_eq!(plan["pull_requests_skipped"], json!(1));
    assert_eq!(plan["criteria_unreadable"], json!([2]));
    assert_eq!(plan["ledger"]["state"], json!("absent"));
    assert!(plan["plan_digest"].as_str().unwrap().starts_with("sha256:"));
    let first = &plan["cards"][0];
    assert_eq!(first["number"], json!(1));
    assert_eq!(first["title"], json!("Add login"));
    assert_eq!(first["criteria"]["rows"], json!(2));
    assert_eq!(
        first["notes"]["comment"],
        json!(1),
        "the member's note is kept"
    );
    assert_eq!(
        first["dropped"]["other"].as_u64().unwrap()
            + first["dropped"]["control_lines"].as_u64().unwrap(),
        1,
        "the stranger's comment is dropped"
    );
    assert!(!w.calls().is_empty());
    only_gets(&w.calls());
}

#[test]
fn an_apply_without_a_ledger_stops_before_it_reads_the_host() {
    let w = world();
    let before = w.home.snapshot();
    let plan = w.plan();
    let calls_after_plan = w.calls().len();
    let text = failure(&w.apply(plan["plan_digest"].as_str().unwrap()));
    assert!(text.contains("ledger not found"), "{text}");
    assert_eq!(
        w.calls().len(),
        calls_after_plan,
        "the host was not read again"
    );
    assert_eq!(w.home.snapshot(), before);
}

#[test]
fn an_apply_writes_the_previewed_cards_and_a_second_apply_adds_nothing() {
    let w = world();
    w.register();
    let plan = w.plan();
    assert_eq!(plan["ledger"]["state"], json!("ready"));
    let digest = plan["plan_digest"].as_str().unwrap().to_string();
    assert_eq!(w.cards_total(), 0);

    let first = result(&w.apply(&digest));
    assert_eq!(first["state"], json!("applied"));
    assert_eq!(first["created"], json!([1, 2]));
    assert_eq!(first["already_there"], json!([]));
    assert_eq!(first["created_without_criteria"], json!([2]));
    assert_eq!(first["notes_written"]["comment"], json!(1));
    assert_eq!(w.cards_total(), 2);

    let again = result(&w.apply(&digest));
    assert_eq!(again["state"], json!("unchanged"));
    assert_eq!(again["created"], json!([]));
    assert_eq!(again["already_there"], json!([1, 2]));
    assert_eq!(w.cards_total(), 2);
    only_gets(&w.calls());
}

#[test]
fn a_digest_that_is_not_the_previews_writes_nothing() {
    let w = world();
    w.register();
    let text = failure(
        &w.apply("sha256:0000000000000000000000000000000000000000000000000000000000000000"),
    );
    assert!(text.contains("changed since the preview"), "{text}");
    assert_eq!(w.cards_total(), 0);
}

#[test]
fn an_apply_needs_a_permission_reference() {
    let w = world();
    w.register();
    let plan = w.plan();
    let project = w.project_arg();
    let out = w.run(&[
        "card-import-apply",
        &project,
        "--from",
        "github",
        "--repo",
        "acme/widgets",
        "--plan-digest",
        plan["plan_digest"].as_str().unwrap(),
        "--permission-reference",
        "",
    ]);
    let text = failure(&out);
    assert!(text.contains("permission reference required"), "{text}");
    assert_eq!(w.cards_total(), 0);
}

#[test]
fn gitlab_is_refused_with_a_reason_and_no_request_is_sent() {
    let w = world();
    let before = w.home.snapshot();
    let project = w.project_arg();
    let out = w.run(&[
        "card-import-plan",
        &project,
        "--from",
        "gitlab",
        "--repo",
        "acme/widgets",
    ]);
    let text = failure(&out);
    assert!(text.to_lowercase().contains("gitlab"), "{text}");
    assert!(w.calls().is_empty(), "no request: {:?}", w.calls());
    assert_eq!(w.home.snapshot(), before);
}

#[test]
fn a_repo_that_is_not_owner_slash_name_is_refused_before_any_request() {
    let w = world();
    let project = w.project_arg();
    for repo in ["widgets", "a/b/c", ""] {
        let out = w.run(&[
            "card-import-plan",
            &project,
            "--from",
            "github",
            "--repo",
            repo,
        ]);
        failure(&out);
    }
    assert!(w.calls().is_empty());
}

#[test]
fn a_project_without_a_profile_is_refused() {
    let w = world();
    let bare = w.home.path().join("work").join("bare");
    fs::create_dir_all(&bare).unwrap();
    let out = w.run(&[
        "card-import-plan",
        bare.to_str().unwrap(),
        "--from",
        "github",
        "--repo",
        "acme/widgets",
    ]);
    let text = failure(&out);
    assert!(text.contains("profile"), "{text}");
    assert!(w.calls().is_empty());
}
