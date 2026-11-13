//! #2679 standing goal: `goal-show`, `goal-plan`, `goal-apply` against a temporary repository
//! with a linked worktree. Runs the real binary with a temporary HOME; never touches the
//! developer's repository.
mod common;

use common::git::git_ok;
use serde_json::{json, Value};
use std::{
    fs,
    path::{Path, PathBuf},
    process::Command,
};

struct Fx {
    _tmp: tempfile::TempDir,
    root: PathBuf,
    primary: PathBuf,
    worktree: PathBuf,
}

fn fixture() -> Fx {
    let tmp = tempfile::tempdir().unwrap();
    let root = fs::canonicalize(tmp.path()).unwrap();
    let primary = root.join("primary");
    fs::create_dir_all(&primary).unwrap();
    git_ok(&primary, &["init", "-q", "-b", "main"]);
    fs::write(primary.join("README.md"), "x\n").unwrap();
    git_ok(&primary, &["add", "README.md"]);
    git_ok(
        &primary,
        &[
            "-c",
            "user.name=t",
            "-c",
            "user.email=t@example.invalid",
            "commit",
            "-q",
            "-m",
            "init",
        ],
    );
    let worktree = root.join("lane");
    let wt = worktree.to_string_lossy().to_string();
    git_ok(&primary, &["worktree", "add", "-q", "-b", "lane", &wt]);
    Fx {
        _tmp: tmp,
        root,
        primary,
        worktree,
    }
}

fn run(f: &Fx, args: &[&str]) -> Value {
    let out = Command::new(env!("CARGO_BIN_EXE_moai-cli"))
        .args(args)
        .current_dir(&f.root)
        .env("HOME", f.root.join("home"))
        .env("GIT_CONFIG_GLOBAL", "/dev/null")
        .env("GIT_CONFIG_SYSTEM", "/dev/null")
        .output()
        .unwrap();
    serde_json::from_slice(&out.stdout).unwrap_or_else(|e| {
        panic!(
            "{e}: {}\n{}",
            String::from_utf8_lossy(&out.stdout),
            String::from_utf8_lossy(&out.stderr)
        )
    })
}

fn request(f: &Fx, name: &str, body: Value) -> String {
    let path = f.root.join(name);
    fs::write(&path, serde_json::to_vec(&body).unwrap()).unwrap();
    path.to_string_lossy().into()
}

fn set(text: &str, exclusions: Value) -> Value {
    json!({
        "schema_version": 1,
        "action": "set",
        "text": text,
        "end_condition": "no open pull request and no startable Issue",
        "exclusions": exclusions,
        "source": "user 2026-11-13T08:40Z"
    })
}

fn p(path: &Path) -> String {
    path.to_string_lossy().into()
}

fn plan(f: &Fx, from: &Path, req: &str) -> Value {
    let env = run(f, &["goal-plan", &p(from), "--request", req]);
    assert_eq!(env["ok"], true, "{env}");
    env["result"].clone()
}

fn apply(f: &Fx, from: &Path, req: &str, digest: &str, now: &str) -> Value {
    run(
        f,
        &[
            "goal-apply",
            &p(from),
            "--request",
            req,
            "--plan-digest",
            digest,
            "--now",
            now,
        ],
    )
}

fn show(f: &Fx, from: &Path) -> Value {
    let env = run(f, &["goal-show", &p(from)]);
    assert_eq!(env["ok"], true, "{env}");
    env["result"].clone()
}

/// AC-1/AC-2: no goal reads as `none`; a goal set from the primary worktree is the one a lane
/// in a linked worktree reads, with its text, start time, end condition and exclusions.
#[test]
fn a_goal_set_once_is_read_from_every_worktree() {
    let f = fixture();
    let none = show(&f, &f.worktree);
    assert_eq!(none["state"], "none", "{none}");
    assert_eq!(none["goal"], Value::Null);

    let req = request(
        &f,
        "set.json",
        set(
            "Keep going until no Issue is open",
            json!([{"number": 2383, "reason": "user HOLD"}]),
        ),
    );
    let preview = plan(&f, &f.primary, &req);
    assert_eq!(preview["change"], "create");
    assert_eq!(preview["asks_person"], true);
    assert_eq!(preview["writes"], false);
    assert_eq!(
        show(&f, &f.primary)["state"],
        "none",
        "a plan writes nothing"
    );

    let applied = apply(
        &f,
        &f.primary,
        &req,
        preview["plan_digest"].as_str().unwrap(),
        "2026-11-13T09:30:00Z",
    );
    assert_eq!(applied["result"]["written"], true, "{applied}");

    let seen = show(&f, &f.worktree);
    assert_eq!(seen["state"], "active", "{seen}");
    let g = &seen["goal"];
    assert_eq!(g["text"], "Keep going until no Issue is open");
    assert_eq!(
        g["end_condition"],
        "no open pull request and no startable Issue"
    );
    assert_eq!(
        g["exclusions"],
        json!([{"number": 2383, "reason": "user HOLD"}])
    );
    assert_eq!(g["started_at"], "2026-11-13T09:30:00Z");
    assert_eq!(g["revision"], 1);
    assert_eq!(
        seen["store"],
        show(&f, &f.primary)["store"],
        "one store per clone"
    );
    assert!(
        seen["store"]
            .as_str()
            .unwrap()
            .ends_with(".git/.moai-cli/goal/goal.json"),
        "{seen}"
    );
    // The store is in the common git dir: neither checkout shows a change.
    for dir in [&f.primary, &f.worktree] {
        let status = common::git::git(dir, &["status", "--porcelain"]);
        assert_eq!(status, "", "{}", dir.display());
    }
}

/// AC-3: only a change asks the person. The same request again is `unchanged`
/// (`asks_person: false`) and writes nothing; a changed end condition keeps the start time;
/// new wording restarts it; every applied change raises the revision.
#[test]
fn only_a_change_asks_and_the_start_time_follows_the_wording() {
    let f = fixture();
    let first = request(&f, "a.json", set("Keep going", json!([])));
    let d = plan(&f, &f.primary, &first)["plan_digest"].clone();
    apply(
        &f,
        &f.primary,
        &first,
        d.as_str().unwrap(),
        "2026-11-13T09:00:00Z",
    );

    let again = plan(&f, &f.worktree, &first);
    assert_eq!(again["change"], "unchanged");
    assert_eq!(again["asks_person"], false);
    let out = apply(
        &f,
        &f.worktree,
        &first,
        again["plan_digest"].as_str().unwrap(),
        "2026-11-13T09:05:00Z",
    );
    assert_eq!(out["result"]["written"], false, "{out}");
    assert_eq!(show(&f, &f.primary)["goal"]["revision"], 1);

    let mut narrower = set("Keep going", json!([{"number": 9, "reason": "parent"}]));
    narrower["end_condition"] = json!("no open Issue in milestone rc.2");
    let narrower = request(&f, "b.json", narrower);
    let pv = plan(&f, &f.primary, &narrower);
    assert_eq!(pv["change"], "update");
    assert_eq!(pv["changed_fields"], json!(["end_condition", "exclusions"]));
    apply(
        &f,
        &f.primary,
        &narrower,
        pv["plan_digest"].as_str().unwrap(),
        "2026-11-13T10:00:00Z",
    );
    let g = show(&f, &f.primary)["goal"].clone();
    assert_eq!(
        g["started_at"], "2026-11-13T09:00:00Z",
        "same wording keeps the start"
    );
    assert_eq!(g["updated_at"], "2026-11-13T10:00:00Z");
    assert_eq!(g["revision"], 2);

    let reworded = request(&f, "c.json", set("Stop after rc.2 ships", json!([])));
    let pv = plan(&f, &f.primary, &reworded);
    apply(
        &f,
        &f.primary,
        &reworded,
        pv["plan_digest"].as_str().unwrap(),
        "2026-11-13T11:00:00Z",
    );
    let g = show(&f, &f.primary)["goal"].clone();
    assert_eq!(
        g["started_at"], "2026-11-13T11:00:00Z",
        "new wording restarts"
    );
    assert_eq!(g["revision"], 3);
}

/// The digest binds the request and the file as it was previewed: another request, or a file
/// changed after the preview, is refused and nothing is written.
#[test]
fn a_stale_or_altered_plan_writes_nothing() {
    let f = fixture();
    let a = request(&f, "a.json", set("Keep going", json!([])));
    let b = request(&f, "b.json", set("Something else", json!([])));
    let da = plan(&f, &f.primary, &a)["plan_digest"].clone();
    let refused = apply(
        &f,
        &f.primary,
        &b,
        da.as_str().unwrap(),
        "2026-11-13T09:00:00Z",
    );
    assert_eq!(refused["ok"], false, "{refused}");
    assert_eq!(show(&f, &f.primary)["state"], "none");

    apply(
        &f,
        &f.primary,
        &a,
        da.as_str().unwrap(),
        "2026-11-13T09:00:00Z",
    );
    let db = plan(&f, &f.primary, &b)["plan_digest"].clone();
    // Another session changes the goal between this preview and its apply.
    let c = request(&f, "c.json", set("Third wording", json!([])));
    let dc = plan(&f, &f.worktree, &c)["plan_digest"].clone();
    apply(
        &f,
        &f.worktree,
        &c,
        dc.as_str().unwrap(),
        "2026-11-13T09:10:00Z",
    );
    let late = apply(
        &f,
        &f.primary,
        &b,
        db.as_str().unwrap(),
        "2026-11-13T09:20:00Z",
    );
    assert_eq!(late["ok"], false, "{late}");
    assert_eq!(show(&f, &f.primary)["goal"]["text"], "Third wording");
}

/// A cleared goal stays readable as `cleared`; clearing twice changes nothing.
#[test]
fn clear_keeps_the_last_goal_readable() {
    let f = fixture();
    let a = request(&f, "a.json", set("Keep going", json!([])));
    let d = plan(&f, &f.primary, &a)["plan_digest"].clone();
    apply(
        &f,
        &f.primary,
        &a,
        d.as_str().unwrap(),
        "2026-11-13T09:00:00Z",
    );
    let clear = request(
        &f,
        "clear.json",
        json!({"schema_version": 1, "action": "clear", "source": "user 2026-11-13T12:00Z"}),
    );
    let pv = plan(&f, &f.primary, &clear);
    assert_eq!(pv["change"], "clear");
    apply(
        &f,
        &f.primary,
        &clear,
        pv["plan_digest"].as_str().unwrap(),
        "2026-11-13T12:00:00Z",
    );
    let seen = show(&f, &f.worktree);
    assert_eq!(seen["state"], "cleared", "{seen}");
    assert_eq!(seen["goal"]["text"], "Keep going");
    assert_eq!(seen["goal"]["active"], false);
    assert_eq!(plan(&f, &f.primary, &clear)["change"], "none");
}

/// Malformed requests are refused before anything is read or written.
#[test]
fn bad_requests_are_refused() {
    let f = fixture();
    for (name, body) in [
        ("empty-text", set("", json!([]))),
        (
            "dup",
            set(
                "x",
                json!([{"number": 1, "reason": "a"}, {"number": 1, "reason": "b"}]),
            ),
        ),
        (
            "unknown",
            json!({"schema_version": 1, "action": "set", "text": "x", "end_condition": "y", "source": "z", "extra": 1}),
        ),
        (
            "secret",
            set("token ghp_0123456789abcdefghijklmnopqrstuvwxyz", json!([])),
        ),
    ] {
        let req = request(&f, &format!("{name}.json"), body);
        let env = run(&f, &["goal-plan", &p(&f.primary), "--request", &req]);
        assert_eq!(env["ok"], false, "{name}: {env}");
    }
    assert_eq!(show(&f, &f.primary)["state"], "none");
}
