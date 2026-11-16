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

struct World {
    home: TempHome,
    /// The stub `gh`, its call log and its answers live outside the HOME.
    _host: tempfile::TempDir,
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
    }]);
    fs::write(dir.join("issues.json"), issues.to_string()).unwrap();
    let posted_comments = json!([{
        "id": 9001, "body": format!("summary\n\n{MARKER}"), "user": {"login": "bot"},
        "author_association": "OWNER", "updated_at": "2026-11-03T00:00:00Z"
    }]);
    fs::write(
        dir.join("posted-comments.json"),
        posted_comments.to_string(),
    )
    .unwrap();
    let log = dir.join("gh-calls.log");
    let script = format!(
        "#!/bin/sh\nprintf '%s\\n' \"$*\" >> '{log}'\nurl=''\nfor a in \"$@\"; do case \"$a\" in https://*) url=\"$a\";; esac; done\n\
         if [ \"$3\" = POST ]; then cat > '{dir}/post-body.txt'; touch '{dir}/posted'; printf 'HTTP/2.0 201 Created\\r\\n\\r\\n{{\"id\": 9001}}'; exit 0; fi\n\
         case \"$url\" in\n\
           */issues/*/comments*) printf 'HTTP/2.0 200 OK\\r\\n\\r\\n'; if [ -f '{dir}/posted' ]; then cat '{dir}/posted-comments.json'; else printf '[]'; fi;;\n\
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
