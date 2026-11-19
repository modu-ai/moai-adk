//! SPEC-20261119-verified-trailer: every commit says how it was verified. The `Moai-Verified`
//! trailer (`tests|mutants|e2e|smoke|none`, comma-separated) is printed by `moai trailers
//! --verified`, judged by `moai trailer-check` (a missing line is `none` only before the rule's
//! start commit: the first commit of the range that carries the line, or `--verified-since`),
//! and listed by kind by `moai trailers --list`, which also marks commits a standard revert
//! message names. A temporary repository and a temporary HOME only: never the network, the real
//! HOME or the real ledger.

mod common;

use common::home::TempHome;
use serde_json::Value;
use std::fs;
use std::path::PathBuf;
use std::process::Output;

const SESSION_VAR: &str = "CLAUDE_CODE_SESSION_ID";
const S_A: &str = "aaaaaaaa-0000-4000-8000-00000000000a";

fn envelope(out: &Output) -> Value {
    serde_json::from_slice(&out.stdout).unwrap_or_else(|_| {
        panic!(
            "{} / {}",
            String::from_utf8_lossy(&out.stdout),
            String::from_utf8_lossy(&out.stderr)
        )
    })
}

fn ok(out: &Output) -> Value {
    let v = envelope(out);
    assert_eq!(out.status.code(), Some(0), "{v}");
    v["result"].clone()
}

/// A `trailer-check` that answered with findings: the envelope is fine, the command exits 1
/// because `ok` in its result is false.
fn checked(out: &Output) -> Value {
    let v = envelope(out);
    assert_eq!(out.status.code(), Some(1), "{v}");
    assert_eq!(v["ok"], true, "{v}");
    v["result"].clone()
}

fn refused(out: &Output) -> String {
    let v = envelope(out);
    assert_eq!(out.status.code(), Some(1), "{v}");
    v["error"]["message"]
        .as_str()
        .unwrap_or_default()
        .to_string()
}

/// One repository under a temporary HOME, with a first commit already made.
struct Repo {
    home: TempHome,
    project: PathBuf,
}

impl Repo {
    fn new() -> Repo {
        let home = TempHome::new();
        let project = home.path().join("p");
        fs::create_dir_all(&project).unwrap();
        let r = Repo { home, project };
        r.git(&["init", "-q", "-b", "main"]);
        fs::write(r.project.join("a.txt"), "a\n").unwrap();
        r.git(&["add", "-A"]);
        r.commit("init", None);
        r
    }

    fn p(&self) -> &str {
        self.project.to_str().unwrap()
    }

    fn git(&self, args: &[&str]) -> String {
        common::git::git(&self.project, args)
    }

    /// A commit whose message ends with `Moai-Verified: <line>` when `verified` names one.
    fn commit(&self, subject: &str, verified: Option<&str>) -> String {
        let message = match verified {
            Some(v) => format!("{subject}\n\nMoai-Verified: {v}\n"),
            None => format!("{subject}\n"),
        };
        self.git(&["add", "-A"]);
        self.git_as(&["commit", "-q", "-m", &message]);
        self.git(&["rev-parse", "HEAD"]).trim().to_string()
    }

    /// A commit whose message is the standard revert of `of` (full body, no verified line);
    /// the tree changes too, as a real revert does.
    fn revert(&self, subject: &str, of: &str) -> String {
        self.touch();
        let message = format!("{subject}\n\nThis reverts commit {of}:\nold message\n");
        self.git(&["add", "-A"]);
        self.git_as(&["commit", "-q", "-m", &message]);
        self.git(&["rev-parse", "HEAD"]).trim().to_string()
    }

    fn git_as(&self, args: &[&str]) {
        let mut all = vec![
            "-c",
            "user.name=fixture",
            "-c",
            "user.email=fixture@example.invalid",
            "-c",
            "commit.gpgsign=false",
        ];
        all.extend_from_slice(args);
        common::git::git_ok(&self.project, &all);
    }

    /// `moai <args>` with the session id of `S_A`.
    fn run(&self, args: &[&str]) -> Output {
        let mut cmd = self.home.command(args);
        cmd.env(SESSION_VAR, S_A);
        cmd.output().expect("run moai")
    }

    fn touch(&self) {
        use std::sync::atomic::{AtomicUsize, Ordering};
        static N: AtomicUsize = AtomicUsize::new(0);
        let n = N.fetch_add(1, Ordering::Relaxed);
        fs::write(self.project.join("a.txt"), format!("{n}\n")).unwrap();
    }
}

/// AC-1: `--verified` puts the line in the printed trailers, several kinds keep their order,
/// and an unknown kind is refused with no text at all.
#[test]
fn trailers_prints_the_verified_line() {
    let r = Repo::new();
    let v = ok(&r.run(&[
        "trailers",
        "--spec",
        "SPEC-20261119-x",
        "--verified",
        "tests",
        "mutants",
    ]));
    let text = v["text"].as_str().unwrap();
    assert!(text.contains("Moai-Spec: SPEC-20261119-x\n"), "{text}");
    assert!(text.contains("Moai-Verified: tests, mutants\n"), "{text}");
    // `none` alone is an honest value.
    let v = ok(&r.run(&[
        "trailers",
        "--spec",
        "SPEC-20261119-x",
        "--verified",
        "none",
    ]));
    assert!(v["text"]
        .as_str()
        .unwrap()
        .contains("Moai-Verified: none\n"));
    // Without --verified the line is simply absent.
    let v = ok(&r.run(&["trailers", "--spec", "SPEC-20261119-x"]));
    assert!(!v["text"].as_str().unwrap().contains("Moai-Verified"));
    // A misspelled kind is refused, and no trailer text is printed.
    let why = refused(&r.run(&[
        "trailers",
        "--spec",
        "SPEC-20261119-x",
        "--verified",
        "tsets",
    ]));
    assert!(why.contains("tests"), "{why}");
}

/// AC-2: the check cuts at the rule's start — the range's first commit carrying the line.
/// Before it a missing line reads as `none`; after it, `missing_verified` fails; a misspelled
/// value fails anywhere; a range with no carrying commit passes as all-before; `--verified-since`
/// fixes the start itself.
#[test]
fn trailer_check_cuts_at_the_rule_start() {
    let r = Repo::new();
    r.touch();
    let c1 = r.commit("one", None);
    r.touch();
    let c2 = r.commit("two", Some("tests"));
    r.touch();
    let c3 = r.commit("three", None);
    r.touch();
    let c4 = r.commit("four", Some("mutants, e2e"));
    r.touch();
    let c5 = r.commit("five", Some("tset"));
    // The whole range: c1 before the start (c2) is fine; c3 is after it and misses the line;
    // c5 carries a misspelled value.
    let v = checked(&r.run(&["trailer-check", &format!("{c1}..HEAD"), "--dir", r.p()]));
    assert_eq!(v["failing"], 2, "{v}");
    let kinds: Vec<&str> = v["findings"]
        .as_array()
        .unwrap()
        .iter()
        .map(|f| f["kind"].as_str().unwrap())
        .collect();
    assert!(kinds.contains(&"missing_verified"), "{v}");
    assert!(kinds.contains(&"malformed_verified"), "{v}");
    let missing: Vec<&str> = v["findings"]
        .as_array()
        .unwrap()
        .iter()
        .filter(|f| f["kind"] == "missing_verified")
        .map(|f| f["sha"].as_str().unwrap())
        .collect();
    assert_eq!(missing, [c3.as_str()], "{v}");
    let malformed: Vec<&str> = v["findings"]
        .as_array()
        .unwrap()
        .iter()
        .filter(|f| f["kind"] == "malformed_verified")
        .map(|f| f["sha"].as_str().unwrap())
        .collect();
    assert_eq!(malformed, [c5.as_str()], "{v}");
    // Only the first commit of the range: no carrying commit, so all of it is before the start.
    let v = ok(&r.run(&["trailer-check", &format!("{c1}~1..{c1}"), "--dir", r.p()]));
    assert_eq!(v["failing"], 0, "{v}");
    assert_eq!(v["ok"], true, "{v}");
    // A range that begins past c2 (excluded by `c2..`) names its own start: c4, the range's
    // first carrying commit — c3 before it reads as `none` (the start is range-local, D-1),
    // and only the misspelled c5 fails.
    let v = checked(&r.run(&["trailer-check", &format!("{c2}..HEAD"), "--dir", r.p()]));
    assert_eq!(v["failing"], 1, "{v}");
    assert_eq!(v["verified_from"], 1, "{v}");
    // A repo with no carrying commit at all: everything reads as `none`.
    let b = Repo::new();
    b.touch();
    let b1 = b.commit("b-one", None);
    b.touch();
    b.commit("b-two", None);
    let v = ok(&b.run(&["trailer-check", &format!("{b1}..HEAD"), "--dir", b.p()]));
    assert_eq!(v["failing"], 0, "{v}");
    // `--verified-since` names the start itself: commits after it need the line even when
    // nothing in the range carries one.
    let v = checked(&b.run(&[
        "trailer-check",
        &format!("{b1}..HEAD"),
        "--dir",
        b.p(),
        "--verified-since",
        &b1,
    ]));
    assert_eq!(v["failing"], 1, "{v}");
    assert_eq!(
        v["findings"]
            .as_array()
            .unwrap()
            .iter()
            .filter(|f| f["fails"] == serde_json::json!(true))
            .count(),
        1,
        "{v}"
    );
    // The named start commit itself is not after itself: it stays exempt.
    let v = ok(&b.run(&[
        "trailer-check",
        &format!("{b1}~1..{b1}"),
        "--dir",
        b.p(),
        "--verified-since",
        &b1,
    ]));
    assert_eq!(v["failing"], 0, "{v}");
    assert_eq!(v["commits"], 1, "{v}");
    // c4 carries two accepted kinds: it is accepted.
    let v = ok(&r.run(&["trailer-check", &format!("{c4}~1..{c4}"), "--dir", r.p()]));
    assert_eq!(v["failing"], 0, "{v}");
    assert_eq!(v["commits"], 1, "{v}");
}

/// AC-3: `--list` groups the commits of a range by verification kind — a commit with several
/// kinds appears in each — and marks the commits a standard revert message names, so `none` and
/// reverted commits can be counted together.
#[test]
fn trailers_list_groups_and_marks_reverts() {
    let r = Repo::new();
    r.touch();
    let v1 = r.commit("v-one", Some("tests"));
    r.touch();
    let v2 = r.commit("v-two", None);
    r.touch();
    let v3 = r.commit("v-three", Some("tests, mutants"));
    let v4 = r.revert("Revert \"v-three\"", &v3);
    let v = ok(&r.run(&[
        "trailers",
        "--list",
        "--range",
        &format!("{v1}..HEAD"),
        "--dir",
        r.p(),
    ]));
    let by = |k: &str| -> Vec<&str> {
        v["by_kind"][k]
            .as_array()
            .unwrap()
            .iter()
            .map(|s| s.as_str().unwrap())
            .collect()
    };
    // `v1..HEAD` excludes v1 itself (git range semantics): v2, v3 and v4 are listed.
    assert_eq!(by("tests"), [v3.as_str()], "{v}");
    assert_eq!(by("mutants"), [v3.as_str()], "{v}");
    assert_eq!(by("none"), [v2.as_str(), v4.as_str()], "{v}");
    assert_eq!(v["reverted"].as_array().unwrap().len(), 1, "{v}");
    assert!(
        v["reverted"]
            .as_array()
            .unwrap()
            .contains(&serde_json::json!(v3)),
        "{v}"
    );
    let commits: Vec<&str> = v["commits"]
        .as_array()
        .unwrap()
        .iter()
        .map(|c| c["sha"].as_str().unwrap())
        .collect();
    assert_eq!(commits, [v2.as_str(), v3.as_str(), v4.as_str()], "{v}");
    assert_eq!(v["counts"]["tests"], 1, "{v}");
    assert_eq!(v["counts"]["none"], 2, "{v}");
    // A range that includes v1 groups it under its own kind too.
    let v = ok(&r.run(&[
        "trailers",
        "--list",
        "--range",
        &format!("{v1}~1..HEAD"),
        "--dir",
        r.p(),
    ]));
    assert_eq!(
        v["by_kind"]["tests"].as_array().unwrap(),
        &[serde_json::json!(v1), serde_json::json!(v3)],
        "{v}"
    );
}
