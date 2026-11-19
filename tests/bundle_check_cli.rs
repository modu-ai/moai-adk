//! `moai bundle check` (SPEC-20261119-bundle-ready-check, card C-119): every check of every head,
//! all findings in one answer, against a temporary ledger and a temporary repository with a bare
//! `origin`. Every command runs against a temporary HOME; the real `~/.moai` and the real
//! repository are not read or written.
#![cfg(unix)]

mod common;

use common::home::TempHome;
use moai_cli::ledger;
use serde_json::Value;
use std::cell::Cell;
use std::fs;
use std::io::Write;
use std::path::{Path, PathBuf};
use std::process::Stdio;

const TIP: &str = "origin/trunk";
const SPEC: &str = "SPEC-20261119-demo";
const MANIFEST: &str = "skills/moai/references/source-manifest.json";

/// A `spec.md` that passes `spec check`, with `## Result` lines for the given AC numbers.
fn spec_text(id: &str, result_acs: &[u32]) -> String {
    let mut results = String::new();
    for n in result_acs {
        results.push_str(&format!("- AC-{n}: met — ran it.\n"));
    }
    format!(
        "---\nid: {id}\ntitle: \"demo: a SPEC the bundle tests carry\"\nstatus: draft\nrevision: 1\ncreated: 2026-11-19\n---\n\n\
         # demo: a SPEC the bundle tests carry\n\n## Intent\n\nA fixture.\n\n## Non-goals\n\n- Nothing else.\n\n\
         ## Requirements\n\n- REQ-1: When `demo run` is run, `demo run` shall print one line.\n\
         - REQ-2 (REQ-1): IF the input is empty, THEN `demo run` shall print nothing.\n\n\
         ## Acceptance\n\n- AC-1 (REQ-1): it prints one line.\n  manual: run it.\n\
         - AC-2 (REQ-2): it prints nothing for empty input.\n  manual: run it with empty input.\n\n\
         ## Scope\n\n- write: `src/demo.rs`\n\n## Plan\n\nOne step.\n\n## Risks\n\n- None.\n\n\
         ## Decisions\n\n| Id | Decision | Basis | Undo |\n| --- | --- | --- | --- |\n| D-1 | None | fixture | none |\n\n\
         ## Result\n\n{results}"
    )
}

struct World {
    home: TempHome,
    _root: tempfile::TempDir,
    project: String,
    work: PathBuf,
    cards: Cell<u32>,
}

#[allow(dead_code)]
impl World {
    /// A temporary HOME with a ledger, a bare `origin` and a work clone whose tip is
    /// `origin/trunk`. A merge driver that always succeeds is registered for the manifest, as in
    /// the real clone; the tool must not let it decide.
    fn new() -> World {
        let home = TempHome::new();
        let project = home.project("demo");
        ledger::register(home.path(), "demo").unwrap();
        let root = tempfile::Builder::new()
            .prefix("moai-test-bundle-")
            .tempdir()
            .unwrap();
        let bare = root.path().join("origin.git");
        let work = root.path().join("work");
        fs::create_dir_all(&work).unwrap();
        let w = World {
            home,
            _root: root,
            project: project.to_str().unwrap().to_string(),
            work,
            cards: Cell::new(0),
        };
        let bare_s = bare.to_str().unwrap().to_string();
        w.git_in(
            w.work.parent().unwrap(),
            &["init", "-q", "--bare", "-b", "trunk", &bare_s],
        );
        w.git(&["init", "-q", "-b", "trunk"]);
        w.git(&["remote", "add", "origin", &bare_s]);
        w.git(&["config", "merge.moai-manifest.driver", "true"]);
        w.put("README.md", "tip\n");
        w.put(
            ".gitattributes",
            &format!("{MANIFEST} merge=moai-manifest\n"),
        );
        w.put(MANIFEST, "{\n  \"line\": \"base\"\n}\n");
        w.git(&["add", "-A"]);
        w.git(&["commit", "-q", "-m", "chore: the tip"]);
        w.git(&["push", "-q", "origin", "trunk"]);
        w.git(&["fetch", "-q", "origin"]);
        w
    }

    fn put(&self, path: &str, text: &str) {
        let f = self.work.join(path);
        fs::create_dir_all(f.parent().unwrap()).unwrap();
        fs::write(f, text).unwrap();
    }

    fn git_raw(&self, dir: &Path, args: &[&str]) -> std::process::Output {
        let mut all = vec![
            "-c",
            "user.name=Test",
            "-c",
            "user.email=test@example.test",
            "-c",
            "commit.gpgsign=false",
            "-c",
            "protocol.file.allow=always",
        ];
        all.extend_from_slice(args);
        self.home
            .command_for("git", &all)
            .current_dir(dir)
            .stdin(Stdio::null())
            .output()
            .expect("run git")
    }

    fn git_in(&self, dir: &Path, args: &[&str]) -> String {
        let out = self.git_raw(dir, args);
        assert!(
            out.status.success(),
            "git {args:?}: {}",
            String::from_utf8_lossy(&out.stderr)
        );
        String::from_utf8_lossy(&out.stdout).trim().to_string()
    }

    fn git(&self, args: &[&str]) -> String {
        self.git_in(&self.work, args)
    }

    fn tip_sha(&self) -> String {
        self.git(&["rev-parse", TIP])
    }

    /// A branch of one commit on `from`, with these files and this `Moai-Spec` trailer (none when
    /// `None`); pushed to `origin` when `push`. Returns the commit.
    fn head_raw(
        &self,
        branch: &str,
        from: &str,
        files: &[(&str, &str)],
        trailer: Option<&str>,
        push: bool,
    ) -> String {
        self.git(&["checkout", "-q", "-B", branch, from]);
        for (p, c) in files {
            self.put(p, c);
        }
        self.git(&["add", "-A"]);
        let msg = match trailer {
            Some(t) => format!("feat: {branch}\n\nMoai-Spec: {t}"),
            None => format!("feat: {branch}"),
        };
        self.git(&["commit", "-q", "-m", &msg]);
        let sha = self.git(&["rev-parse", "HEAD"]);
        if push {
            self.git(&[
                "push",
                "-q",
                "origin",
                &format!("{branch}:refs/heads/{branch}"),
            ]);
            self.git(&["fetch", "-q", "origin"]);
        }
        sha
    }

    /// The same with a valid SPEC file for `spec_id` whose `## Result` covers `result_acs`.
    fn head(
        &self,
        branch: &str,
        from: &str,
        files: &[(&str, &str)],
        spec_id: &str,
        result_acs: &[u32],
        push: bool,
    ) -> String {
        let path = format!(".moai/specs/{spec_id}/spec.md");
        let text = spec_text(spec_id, result_acs);
        let mut all: Vec<(&str, &str)> = files.to_vec();
        all.push((&path, &text));
        self.head_raw(branch, from, &all, Some(spec_id), push)
    }

    /// `moai card <args>` with an optional stdin; the answer's `result`.
    fn card(&self, args: &[&str], stdin: Option<&str>) -> Value {
        let mut all = vec!["card"];
        all.extend_from_slice(args);
        all.extend_from_slice(&["--project", &self.project]);
        let mut cmd = self.home.command(&all);
        cmd.stdin(if stdin.is_some() {
            Stdio::piped()
        } else {
            Stdio::null()
        })
        .stdout(Stdio::piped())
        .stderr(Stdio::piped());
        let mut child = cmd.spawn().expect("run moai");
        if let Some(t) = stdin {
            child.stdin.take().unwrap().write_all(t.as_bytes()).unwrap();
        }
        let out = child.wait_with_output().unwrap();
        let line = String::from_utf8_lossy(&out.stdout)
            .lines()
            .next()
            .unwrap_or("")
            .to_string();
        let v: Value = serde_json::from_str(&line).unwrap_or_else(|e| {
            panic!(
                "card {args:?}: {e}: {}",
                String::from_utf8_lossy(&out.stderr)
            )
        });
        assert_eq!(v["ok"], true, "card {args:?}: {v}");
        v["result"].clone()
    }

    fn new_card(&self) -> String {
        let n = self.cards.get() + 1;
        self.cards.set(n);
        let scope = format!("area{n}/");
        self.card(
            &["add", "--title", "work", "--scope", &scope, "--ac", "done"],
            None,
        )["name"]
            .as_str()
            .unwrap()
            .to_string()
    }

    /// Adds a note and returns its number.
    fn note(&self, card: &str, text: &str) -> i64 {
        self.card(
            &["note", card, "--kind", "report", "--by", "lane"],
            Some(text),
        );
        let shown = self.card(&["show", card], None);
        shown["notes"]
            .as_array()
            .unwrap()
            .iter()
            .map(|n| n["note_id"].as_i64().unwrap())
            .max()
            .unwrap()
    }

    fn ready_note(&self, card: &str, head: &str) -> i64 {
        self.note(
            card,
            &format!("LANE-READY-INTEGRATION {card} head={head} spec={SPEC}"),
        )
    }

    fn pass_note(&self, card: &str, head: &str) -> i64 {
        let tip = self.tip_sha();
        self.note(card, &format!("AUDIT {card} ({SPEC}, no PR — M-Flow) head {head} base {tip} — verdict: PASS\nbody"))
    }

    /// `moai bundle <args> --project … --dir … --tip origin/trunk`: the first stdout line as JSON
    /// and the exit status.
    fn bundle(&self, args: &[&str]) -> (Value, i32) {
        let mut all = vec!["bundle"];
        all.extend_from_slice(args);
        all.extend_from_slice(&[
            "--project",
            &self.project,
            "--dir",
            self.work.to_str().unwrap(),
            "--tip",
            TIP,
        ]);
        let out = self
            .home
            .command(&all)
            .stdin(Stdio::null())
            .output()
            .expect("run moai");
        let line = String::from_utf8_lossy(&out.stdout)
            .lines()
            .next()
            .unwrap_or("")
            .to_string();
        let v: Value = serde_json::from_str(&line).unwrap_or_else(|e| {
            panic!(
                "bundle {args:?}: {e}: {}",
                String::from_utf8_lossy(&out.stderr)
            )
        });
        (v, out.status.code().unwrap_or(-1))
    }

    /// Every ref, the work tree status and the HEAD of the work clone, and the bare `origin`'s refs.
    fn repo_state(&self) -> String {
        let bare = self.work.parent().unwrap().join("origin.git");
        format!(
            "{}\n--\n{}\n--\n{}\n--\n{}",
            self.git(&["for-each-ref"]),
            self.git(&["status", "--porcelain"]),
            self.git(&["rev-parse", "HEAD"]),
            self.git_in(&bare, &["for-each-ref"]),
        )
    }
}

#[allow(dead_code)]
fn findings(v: &Value) -> Vec<(String, String)> {
    v["result"]["findings"]
        .as_array()
        .unwrap()
        .iter()
        .map(|f| {
            (
                f["card"].as_str().unwrap().to_string(),
                f["code"].as_str().unwrap().to_string(),
            )
        })
        .collect()
}

#[test]
fn a_clean_head_has_no_finding() {
    let w = World::new();
    let h = w.head(
        "spec/clean",
        TIP,
        &[("docs/clean.md", "c\n")],
        SPEC,
        &[1, 2],
        true,
    );
    let c = w.new_card();
    w.ready_note(&c, &h);
    for arg in [c.clone(), format!("{c}={h}")] {
        let (v, code) = w.bundle(&["check", &arg]);
        assert_eq!(code, 0, "{v}");
        let r = &v["result"];
        assert_eq!(r["ok"], true, "{v}");
        assert_eq!(r["findings"], serde_json::json!([]));
        assert_eq!(r["heads"][0]["audit_floor"], 0);
        assert_eq!(r["heads"][0]["head"], h.as_str());
        assert_eq!(r["heads"][0]["specs"], serde_json::json!([SPEC]));
        assert_eq!(r["writes"], false);
    }
}

#[test]
fn findings_of_all_heads_come_together_with_stable_codes() {
    let w = World::new();
    // C-1: only here, a trailer that is not a SPEC id.
    let a = w.head_raw("spec/a", TIP, &[("docs/a.md", "a\n")], Some("bogus"), false);
    // C-2: the tip itself.
    let b = w.tip_sha();
    // C-3: pushed, a SPEC whose Result lacks AC-2, a change under src/ without a verdict.
    let c = w.head(
        "spec/c",
        TIP,
        &[("src/c.rs", "fn c() {}\n")],
        SPEC,
        &[1],
        true,
    );
    let (c1, c2, c3) = (w.new_card(), w.new_card(), w.new_card());
    w.ready_note(&c1, &a);
    w.ready_note(&c2, &b);
    w.ready_note(&c3, &c);

    let (v, _) = w.bundle(&["check", &c1, &c2, &c3]);
    assert_eq!(v["result"]["ok"], false, "{v}");
    let mut got = findings(&v);
    got.sort();
    let mut want: Vec<(String, String)> = [
        (&c1, "not_on_origin"),
        (&c1, "trailers"),
        (&c1, "spec"),
        (&c2, "already_merged"),
        (&c3, "audit_missing"),
        (&c3, "result_missing"),
    ]
    .iter()
    .map(|(card, code)| (card.to_string(), code.to_string()))
    .collect();
    want.sort();
    assert_eq!(got, want, "{v}");
    let missing = v["result"]["findings"]
        .as_array()
        .unwrap()
        .iter()
        .find(|f| f["code"] == "result_missing")
        .unwrap();
    assert_eq!(missing["detail"]["acs"], serde_json::json!(["AC-2"]));
    assert_eq!(missing["head"], c.as_str());
    assert_eq!(v["result"]["unreadable_verdict"], serde_json::json!([]));
}

#[test]
fn the_audit_floor_decides_which_heads_need_a_pass() {
    let w = World::new();
    // C-1: only docs change: floor 0, no verdict needed.
    let x = w.head("spec/x", TIP, &[("docs/x.md", "x\n")], SPEC, &[1, 2], true);
    // C-2: src changes and a PASS names exactly the head.
    let y = w.head(
        "spec/y",
        TIP,
        &[("src/y.rs", "fn y() {}\n")],
        SPEC,
        &[1, 2],
        true,
    );
    // C-3: on top of C-2's head; its PASS names the parent, its own verdict is FAIL.
    let z = w.head(
        "spec/z",
        &y,
        &[("src/z.rs", "fn z() {}\n")],
        SPEC,
        &[1, 2],
        true,
    );
    let (c1, c2, c3) = (w.new_card(), w.new_card(), w.new_card());
    w.ready_note(&c1, &x);
    w.ready_note(&c2, &y);
    let pass = w.pass_note(&c2, &y);
    w.ready_note(&c3, &z);
    w.pass_note(&c3, &y);
    let tip = w.tip_sha();
    w.note(
        &c3,
        &format!("AUDIT {c3} (x) head {z} base {tip} — verdict: FAIL"),
    );
    let unreadable = w.note(&c3, "AUDIT free text without the head fields");

    let (v, _) = w.bundle(&["check", &c1, &c2, &c3]);
    let r = &v["result"];
    assert_eq!(
        findings(&v),
        vec![(c3.clone(), "audit_missing".to_string())],
        "{v}"
    );
    let heads = r["heads"].as_array().unwrap();
    assert_eq!(
        (
            heads[0]["audit_floor"].clone(),
            heads[1]["audit_floor"].clone()
        ),
        (0.into(), 1.into())
    );
    assert_eq!(heads[1]["pass_note"], pass);
    assert_eq!(heads[2]["pass_note"], Value::Null);
    let f = &r["findings"][0];
    assert_eq!(f["detail"]["floor"], 1);
    assert_eq!(f["detail"]["ancestor_pass"]["head"], y.as_str());
    assert_eq!(f["detail"]["ancestor_pass"]["commits_behind"], 1);
    // The note that does not read is listed and does not fail the check by itself.
    assert_eq!(
        r["unreadable_verdict"],
        serde_json::json!([{"card": c3, "note_id": unreadable}])
    );
    assert!(r["audit_note"].as_str().unwrap().contains("cannot tell"));
}

#[test]
fn a_conflict_is_found_without_the_merge_driver() {
    let w = World::new();
    let p = w.head(
        "spec/p",
        TIP,
        &[(MANIFEST, "{\n  \"line\": \"p\"\n}\n")],
        SPEC,
        &[1, 2],
        true,
    );
    let q = w.head(
        "spec/q",
        TIP,
        &[(MANIFEST, "{\n  \"line\": \"q\"\n}\n")],
        SPEC,
        &[1, 2],
        true,
    );
    // The premise: the driver registered in this clone (`true`, always succeeds) hides it.
    let hidden = w.git_raw(&w.work, &["merge-tree", "--write-tree", &p, &q]);
    assert_eq!(
        hidden.status.code(),
        Some(0),
        "the registered driver must hide the conflict"
    );
    let (c1, c2) = (w.new_card(), w.new_card());
    w.ready_note(&c1, &p);
    w.ready_note(&c2, &q);

    let (v, _) = w.bundle(&["check", &c1, &c2]);
    let conflicts: Vec<&Value> = v["result"]["findings"]
        .as_array()
        .unwrap()
        .iter()
        .filter(|f| f["code"] == "conflict")
        .collect();
    assert_eq!(conflicts.len(), 1, "{v}");
    assert_eq!(conflicts[0]["card"], c2.as_str());
    assert_eq!(conflicts[0]["detail"]["with"], c1.as_str());
    assert_eq!(
        conflicts[0]["detail"]["files"],
        serde_json::json!([MANIFEST])
    );
}

#[test]
fn a_head_that_is_not_here_is_reported_and_the_rest_are_judged() {
    let w = World::new();
    let gone = "e".repeat(40);
    let ok = w.head(
        "spec/ok",
        TIP,
        &[("docs/ok.md", "ok\n")],
        SPEC,
        &[1, 2],
        true,
    );
    let (c1, c2) = (w.new_card(), w.new_card());
    w.ready_note(&c1, &gone);
    w.ready_note(&c2, &ok);
    let (v, _) = w.bundle(&["check", &c1, &c2]);
    assert_eq!(v["result"]["ok"], false);
    assert_eq!(
        findings(&v),
        vec![(c1.clone(), "head_missing".to_string())],
        "{v}"
    );
    assert_eq!(v["result"]["findings"][0]["detail"], "fetch first");
    assert_eq!(v["result"]["heads"][1]["findings"], 0);
    // A card without a readable line is named, not guessed.
    let c3 = w.new_card();
    let (v, _) = w.bundle(&["check", &c3]);
    assert_eq!(findings(&v), vec![(c3, "no_ready_line".to_string())]);
}

#[test]
fn the_commands_change_no_ledger_row_and_no_ref() {
    let w = World::new();
    let h = w.head(
        "spec/h",
        TIP,
        &[("src/h.rs", "fn h() {}\n")],
        SPEC,
        &[1],
        true,
    );
    let c = w.new_card();
    w.ready_note(&c, &h);
    let missing_record = w.home.path().join("no-such-record.json");
    let ledger_before = w.home.snapshot();
    let repo_before = w.repo_state();
    let _ = w.bundle(&["ready"]);
    let (v, _) = w.bundle(&["check", &c]);
    assert_eq!(
        v["result"]["ok"], false,
        "the head has findings, so this run is not trivial"
    );
    let (refused, _) = w.bundle(&[
        "body",
        &c,
        "--branch",
        &h,
        "--run",
        missing_record.to_str().unwrap(),
    ]);
    assert_eq!(refused["ok"], false);
    assert_eq!(
        w.home.snapshot(),
        ledger_before,
        "the ledger and the HOME are as before"
    );
    assert_eq!(
        w.repo_state(),
        repo_before,
        "refs, status and HEAD are as before"
    );
}

#[allow(dead_code)]
impl World {
    /// The tip moves on: a new commit on `trunk` that adds `path`, pushed and fetched.
    fn advance_tip(&self, path: &str) {
        self.git(&["checkout", "-q", "trunk"]);
        self.put(path, "fn moved() {}\n");
        self.git(&["add", "-A"]);
        self.git(&["commit", "-q", "-m", "chore: the tip moves on"]);
        self.git(&["push", "-q", "origin", "trunk"]);
        self.git(&["fetch", "-q", "origin"]);
    }
}

#[test]
fn a_head_behind_the_tip_does_not_count_the_tips_own_changes() {
    let w = World::new();
    // A head that only changes docs, cut from the old tip.
    let h = w.head(
        "spec/behind",
        TIP,
        &[("docs/behind.md", "b\n")],
        SPEC,
        &[1, 2],
        true,
    );
    // The tip then gains a change under src/. Two dots would count it against the head.
    w.advance_tip("src/moved.rs");
    let c = w.new_card();
    w.ready_note(&c, &h);
    let (v, _) = w.bundle(&["check", &c]);
    assert_eq!(v["result"]["heads"][0]["audit_floor"], 0, "{v}");
    assert_eq!(v["result"]["ok"], true, "{v}");
    assert_eq!(v["result"]["findings"], serde_json::json!([]));
}
