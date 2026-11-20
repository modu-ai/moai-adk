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

    /// A commit on its own file `file` with the commit date pinned to `date`, so the order of
    /// `git log` does not depend on the clock; `verified` is the `Moai-Verified` line, if any.
    fn commit_at(&self, subject: &str, verified: Option<&str>, date: &str, file: &str) -> String {
        fs::write(self.project.join(file), format!("{subject}\n")).unwrap();
        let message = match verified {
            Some(v) => format!("{subject}\n\nMoai-Verified: {v}\n"),
            None => format!("{subject}\n"),
        };
        self.git(&["add", "-A"]);
        common::git::git_pinned(
            &self.project,
            &[
                "-c",
                "user.name=fixture",
                "-c",
                "user.email=fixture@example.invalid",
                "-c",
                "commit.gpgsign=false",
                "commit",
                "-q",
                "-m",
                &message,
            ],
            Some(date),
        );
        self.git(&["rev-parse", "HEAD"]).trim().to_string()
    }

    /// A merge commit of `branch` with the date pinned; merges carry no `Moai-Verified` line.
    fn merge_at(&self, branch: &str, date: &str) {
        common::git::git_pinned(
            &self.project,
            &[
                "-c",
                "user.name=fixture",
                "-c",
                "user.email=fixture@example.invalid",
                "-c",
                "commit.gpgsign=false",
                "merge",
                "-q",
                "--no-ff",
                "-m",
                "merge side",
                branch,
            ],
            Some(date),
        );
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

/// The shas of a result's `missing_verified` findings, sorted.
fn missing_shas(v: &Value) -> Vec<String> {
    let mut out: Vec<String> = v["findings"]
        .as_array()
        .unwrap()
        .iter()
        .filter(|f| f["kind"] == "missing_verified")
        .map(|f| f["sha"].as_str().unwrap().to_string())
        .collect();
    out.sort();
    out
}

/// SPEC-20261120-trailer-verified-ancestry AC-1 and AC-3: an old side branch whose commits lack
/// the line, merged into a range that has a carrying commit, is not "after the start" — the
/// carrying commit is listed before its commits but they do not descend from it. The dates are
/// pinned so the list order puts the side commits after the carrying one, the order that made
/// the rule fail them.
#[test]
fn trailer_check_old_side_branch() {
    let r = Repo::new();
    let base = r.git(&["rev-parse", "HEAD"]).trim().to_string();
    let m1 = r.commit_at("carrying", Some("tests"), "2026-11-20T10:00:01", "m1.txt");
    r.git(&["checkout", "-q", "-b", "side", &base]);
    let s1 = r.commit_at("old side one", None, "2026-11-20T10:00:02", "s1.txt");
    let s2 = r.commit_at("old side two", None, "2026-11-20T10:00:03", "s2.txt");
    r.git(&["checkout", "-q", "main"]);
    r.merge_at("side", "2026-11-20T10:00:04");
    let range = format!("{base}..HEAD");
    // The default start is the first carrying commit; the side commits are not its descendants.
    let v = ok(&r.run(&["trailer-check", &range, "--dir", r.p()]));
    assert_eq!(v["failing"], 0, "{v}");
    assert_eq!(v["commits"], 3, "{v}");
    // A named start inside the range covers its descendants only: still none here.
    let v = ok(&r.run(&[
        "trailer-check",
        &range,
        "--dir",
        r.p(),
        "--verified-since",
        &m1,
    ]));
    assert_eq!(v["failing"], 0, "{v}");
    // A named start outside the range (the base) keeps covering the whole range.
    let v = checked(&r.run(&[
        "trailer-check",
        &range,
        "--dir",
        r.p(),
        "--verified-since",
        &base,
    ]));
    let mut want = vec![s1, s2];
    want.sort();
    assert_eq!(missing_shas(&v), want, "{v}");
}

/// AC-2: a commit that descends from a carrying commit and lacks the line still fails — on the
/// main line, and on a side branch that forked after the carrying commit — while an old side
/// branch beside them does not.
#[test]
fn trailer_check_descendants_still_fail() {
    let r = Repo::new();
    let base = r.git(&["rev-parse", "HEAD"]).trim().to_string();
    let carrying = r.commit_at("carrying", Some("tests"), "2026-11-20T10:00:01", "c1.txt");
    let late = r.commit_at("main line", None, "2026-11-20T10:00:02", "c2.txt");
    r.git(&["checkout", "-q", "-b", "forked-after", &carrying]);
    let after = r.commit_at("forked after", None, "2026-11-20T10:00:03", "f1.txt");
    r.git(&["checkout", "-q", "-b", "old", &base]);
    r.commit_at("old side", None, "2026-11-20T10:00:04", "o1.txt");
    r.git(&["checkout", "-q", "main"]);
    r.merge_at("forked-after", "2026-11-20T10:00:05");
    r.merge_at("old", "2026-11-20T10:00:06");
    // A commit after both merges descends from the carrying commit only through merge commits.
    let tail = r.commit_at("after the merges", None, "2026-11-20T10:00:07", "t1.txt");
    let v = checked(&r.run(&["trailer-check", &format!("{base}..HEAD"), "--dir", r.p()]));
    let mut want = vec![late, after, tail];
    want.sort();
    assert_eq!(missing_shas(&v), want, "{v}");
    assert_eq!(v["failing"], 3, "{v}");
}

/// AC-3: `--verified-since` inside the range covers the descendants of the named commit — a
/// chain of them, even when a child's commit date is older than its parent's — and neither the
/// named commit nor an old side branch beside it, when nothing in the range carries the line.
#[test]
fn trailer_check_since_inside_the_range() {
    let r = Repo::new();
    let base = r.git(&["rev-parse", "HEAD"]).trim().to_string();
    let a1 = r.commit_at("named start", None, "2026-11-20T10:00:01", "a1.txt");
    let a2 = r.commit_at("child", None, "2026-11-20T10:00:05", "a2.txt");
    // The grandchild's date is older than its parent's (a skewed clock), and a2 has a second,
    // newer child on a side branch: a date-ordered walk reaches a2 through that child first and
    // lists a2 after the grandchild, so only a parents-first order marks the grandchild.
    let a3 = r.commit_at("grandchild", None, "2026-11-20T10:00:02", "a3.txt");
    r.git(&["checkout", "-q", "-b", "mid", &a2]);
    let a4 = r.commit_at("second child", None, "2026-11-20T10:00:06", "a4.txt");
    r.git(&["checkout", "-q", "-b", "old", &base]);
    r.commit_at("old side", None, "2026-11-20T10:00:08", "o1.txt");
    r.git(&["checkout", "-q", "main"]);
    r.merge_at("mid", "2026-11-20T10:00:07");
    r.merge_at("old", "2026-11-20T10:00:09");
    let range = format!("{base}..HEAD");
    // Nothing carries the line and no start is named: every commit reads as `none`.
    let v = ok(&r.run(&["trailer-check", &range, "--dir", r.p()]));
    assert_eq!(v["failing"], 0, "{v}");
    let v = checked(&r.run(&[
        "trailer-check",
        &range,
        "--dir",
        r.p(),
        "--verified-since",
        &a1,
    ]));
    let mut want = vec![a2, a3, a4];
    want.sort();
    assert_eq!(missing_shas(&v), want, "{v}");
    assert_eq!(v["failing"], 3, "{v}");
}

/// AC-3: `--verified-since` may name a merge commit of the range: a merge is no listed commit, so
/// only its descendants are required, not the commits that forked before it.
#[test]
fn trailer_check_since_names_a_merge_commit() {
    let r = Repo::new();
    let base = r.git(&["rev-parse", "HEAD"]).trim().to_string();
    let a1 = r.commit_at("first", None, "2026-11-20T10:00:01", "a1.txt");
    r.git(&["checkout", "-q", "-b", "side", &base]);
    r.commit_at("side", None, "2026-11-20T10:00:02", "s1.txt");
    r.git(&["checkout", "-q", "-b", "late", &a1]);
    r.commit_at(
        "forked before the merge",
        None,
        "2026-11-20T10:00:03",
        "x1.txt",
    );
    r.git(&["checkout", "-q", "main"]);
    r.merge_at("side", "2026-11-20T10:00:04");
    let m = r.git(&["rev-parse", "HEAD"]).trim().to_string();
    let d1 = r.commit_at("after the merge", None, "2026-11-20T10:00:05", "d1.txt");
    r.merge_at("late", "2026-11-20T10:00:06");
    let range = format!("{base}..HEAD");
    let v = checked(&r.run(&[
        "trailer-check",
        &range,
        "--dir",
        r.p(),
        "--verified-since",
        &m,
    ]));
    assert_eq!(missing_shas(&v), vec![d1], "{v}");
    assert_eq!(v["failing"], 1, "{v}");
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
