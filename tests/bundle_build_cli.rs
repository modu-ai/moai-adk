//! `moai bundle build` (SPEC-20261119-bundle-build, card C-119): one new branch and worktree, one
//! merge commit per head, a conflicting head dropped and named, against a temporary ledger and
//! repository. Every command runs against a temporary HOME; the real `~/.moai` and the real
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
            "user.email=test@example.invalid",
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

const DATE: &str = "20261119";

#[allow(dead_code)]
impl World {
    /// The clone carries an author identity, as the real one does through its global config
    /// (the tool runs git with the user's own settings).
    fn identity(&self) {
        self.git(&["config", "user.name", "Test"]);
        self.git(&["config", "user.email", "test@example.invalid"]);
    }

    /// A card whose newest READY line names `head`.
    fn ready_card(&self, head: &str) -> String {
        let c = self.new_card();
        self.ready_note(&c, head);
        c
    }

    /// `moai bundle build <args> --date 20261119`.
    fn build(&self, args: &[&str]) -> (Value, i32) {
        let mut all = vec!["build"];
        all.extend_from_slice(args);
        all.extend_from_slice(&["--date", DATE]);
        self.bundle(&all)
    }

    fn worktree_of(&self, n: u32) -> PathBuf {
        self.work
            .join(".claude")
            .join("worktrees")
            .join(format!("work-bundle-{n}"))
    }

    fn refs(&self) -> Vec<String> {
        self.git(&["for-each-ref"])
            .lines()
            .map(String::from)
            .collect()
    }

    fn origin_refs(&self) -> String {
        let bare = self.work.parent().unwrap().join("origin.git");
        self.git_in(&bare, &["for-each-ref"])
    }

    /// Whatever the ledger shows: the card list and every card with its notes.
    fn ledger_view(&self) -> Value {
        // `card list` takes the project as a positional argument, not `--project`.
        let out = self
            .home
            .command(&["card", "list", &self.project])
            .stdin(Stdio::null())
            .output()
            .expect("run moai");
        let line = String::from_utf8_lossy(&out.stdout)
            .lines()
            .next()
            .unwrap_or("")
            .to_string();
        let list: Value = serde_json::from_str(&line).expect("card list prints JSON");
        let list = list["result"].clone();
        let mut shown = Vec::new();
        for c in list["cards"].as_array().into_iter().flatten() {
            shown.push(self.card(&["show", c["name"].as_str().unwrap()], None));
        }
        serde_json::json!({"list": list, "cards": shown})
    }
}

fn names(v: &Value, key: &str) -> Vec<String> {
    v["result"][key]
        .as_array()
        .unwrap()
        .iter()
        .map(|e| e["card"].as_str().unwrap().to_string())
        .collect()
}

#[test]
fn two_heads_make_one_branch_and_worktree_with_a_merge_commit_each() {
    let w = World::new();
    w.identity();
    let tip = w.tip_sha();
    let a = w.head("spec/a", TIP, &[("notes/a.md", "a\n")], SPEC, &[1, 2], true);
    let b = w.head("spec/b", TIP, &[("notes/b.md", "b\n")], SPEC, &[1, 2], true);
    let (c1, c2) = (w.ready_card(&a), w.ready_card(&b));

    let (v, code) = w.build(&[&c1, &c2]);
    assert_eq!(code, 0, "{v}");
    let r = &v["result"];
    assert_eq!(r["action"], "build");
    assert_eq!(r["ok"], true, "{v}");
    assert_eq!(r["number"], 1);
    assert_eq!(r["rebuild"], "");
    let branch = format!("m-flow/bundle-{DATE}-1");
    assert_eq!(r["branch"], branch.as_str());
    let wt = w.worktree_of(1);
    assert_eq!(
        fs::canonicalize(r["worktree"].as_str().unwrap()).unwrap(),
        fs::canonicalize(&wt).unwrap()
    );
    assert!(wt.join(".git").exists(), "the worktree exists");
    assert_eq!(names(&v, "merged"), vec![c1.clone(), c2.clone()]);
    assert_eq!(r["tip"]["sha"], tip.as_str());

    // `pinned` is what git says for the branch; the worktree is on it and clean.
    let pinned = w.git(&["rev-parse", &format!("refs/heads/{branch}")]);
    assert_eq!(r["pinned"], pinned.as_str());
    assert_eq!(w.git_in(&wt, &["rev-parse", "HEAD"]), pinned);
    assert_eq!(
        w.git_in(&wt, &["rev-parse", "--abbrev-ref", "HEAD"]),
        branch
    );
    assert_eq!(w.git_in(&wt, &["status", "--porcelain"]), "");
    // One merge commit per head, in the given order, from the tip.
    assert_eq!(
        w.git(&[
            "rev-list",
            "--merges",
            "--count",
            &format!("{tip}..{pinned}")
        ]),
        "2"
    );
    assert_eq!(w.git(&["rev-parse", &format!("{pinned}^2")]), b);
    assert_eq!(w.git(&["rev-parse", &format!("{pinned}^1^2")]), a);
    assert_eq!(w.git(&["rev-parse", &format!("{pinned}^1^1")]), tip);
    assert_eq!(
        w.git(&["log", "-1", "--format=%s", &pinned]),
        format!("Merge {c2} head {} into {branch}", &b[..10])
    );
    assert_eq!(r["merged"][1]["merge"], pinned.as_str());
}

#[test]
fn heads_that_cannot_be_used_are_refused_before_anything_is_created() {
    let w = World::new();
    w.identity();
    let tip = w.tip_sha();
    let none = w.new_card(); // no READY line at all
    let gone = w.ready_card(&"e".repeat(40)); // a head that is not here
    let in_tip = w.ready_card(&tip);
    let ledger = w.ledger_view();
    let before = w.repo_state();
    let nothing_made = |w: &World| {
        assert_eq!(w.git(&["branch", "--list", "m-flow/*"]), "");
        assert!(
            !w.home.path().join(".moai").join("worktrees").exists(),
            "no worktrees folder was made"
        );
        assert!(
            !w.work.join(".claude").join("worktrees").exists(),
            "no worktrees folder was made in the project either"
        );
        assert_eq!(w.repo_state(), before, "no ref, no status change");
    };

    // Two unusable heads: both named in one answer.
    let (v, code) = w.build(&[&none, &gone]);
    assert_eq!(code, 1, "{v}");
    assert_eq!(v["ok"], false);
    let m = v["error"]["message"].as_str().unwrap();
    assert!(m.contains("2 problem(s), nothing was created"), "{m}");
    assert!(
        m.contains(&format!("{none}: no readable LANE-READY-INTEGRATION line")),
        "{m}"
    );
    assert!(
        m.contains(&format!(
            "{gone}: the head eeeeeeeeee is not a commit here (fetch first)"
        )),
        "{m}"
    );
    nothing_made(&w);
    // A card named twice.
    let (v, _) = w.build(&[&in_tip, &in_tip]);
    assert!(
        v["error"]["message"]
            .as_str()
            .unwrap()
            .contains("named twice"),
        "{v}"
    );
    nothing_made(&w);
    // Every head already in the tip: nothing to merge.
    let (v, _) = w.build(&[&in_tip]);
    assert!(
        v["error"]["message"]
            .as_str()
            .unwrap()
            .contains("every head is already in the tip"),
        "{v}"
    );
    nothing_made(&w);
    assert_eq!(w.ledger_view(), ledger);
}

#[test]
fn a_conflicting_head_is_dropped_and_named_and_the_rest_merge() {
    let w = World::new();
    w.identity();
    let tip = w.tip_sha();
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
    let r_ = w.head("spec/r", TIP, &[("notes/r.md", "r\n")], SPEC, &[1, 2], true);
    // The premise: the driver registered in this clone hides the conflict from a plain merge.
    let hidden = w.git_raw(&w.work, &["merge-tree", "--write-tree", &p, &q]);
    assert_eq!(
        hidden.status.code(),
        Some(0),
        "the registered driver hides it"
    );
    let (cp, cq, cr) = (w.ready_card(&p), w.ready_card(&q), w.ready_card(&r_));

    let (v, code) = w.build(&[&cp, &cq, &cr]);
    assert_eq!(code, 0, "{v}");
    let r = &v["result"];
    assert_eq!(r["ok"], false, "{v}");
    assert_eq!(names(&v, "merged"), vec![cp.clone(), cr.clone()]);
    assert_eq!(
        r["dropped"],
        serde_json::json!([{"card": cq, "head": q, "files": [MANIFEST]}])
    );
    assert_eq!(
        r["rebuild"],
        format!("moai bundle build {cp}={p} {cr}={r_} --tip {TIP}").as_str()
    );
    // The branch holds the first and third merges and nothing of the dropped head.
    let pinned = r["pinned"].as_str().unwrap().to_string();
    assert_eq!(
        w.git(&[
            "rev-list",
            "--merges",
            "--count",
            &format!("{tip}..{pinned}")
        ]),
        "2"
    );
    assert_eq!(w.git(&["rev-parse", &format!("{pinned}^2")]), r_);
    assert_eq!(w.git(&["rev-parse", &format!("{pinned}^1^2")]), p);
    let wt = w.worktree_of(1);
    assert_eq!(
        w.git_in(&wt, &["status", "--porcelain"]),
        "",
        "the aborted merge left nothing"
    );
    assert!(
        !w.git_raw(&wt, &["rev-parse", "-q", "--verify", "MERGE_HEAD"])
            .status
            .success(),
        "no merge in progress"
    );
    // The command that `rebuild` names gives a new branch without the dropped head.
    let (v2, _) = w.build(&[&format!("{cp}={p}"), &format!("{cr}={r_}")]);
    assert_eq!(v2["result"]["ok"], true, "{v2}");
    assert_eq!(
        v2["result"]["branch"],
        format!("m-flow/bundle-{DATE}-2").as_str()
    );
    // The same heads merged in the same order give the same tree (the commits differ in time).
    let pinned2 = v2["result"]["pinned"].as_str().unwrap();
    assert_eq!(
        w.git(&["rev-parse", &format!("{pinned2}^{{tree}}")]),
        w.git(&["rev-parse", &format!("{pinned}^{{tree}}")]),
        "the same heads, the same merged tree"
    );
}

#[test]
fn a_head_in_the_tip_or_in_the_branch_is_skipped() {
    let w = World::new();
    w.identity();
    let tip = w.tip_sha();
    let a = w.head("spec/a", TIP, &[("notes/a.md", "a\n")], SPEC, &[1, 2], true);
    let ab = w.head(
        "spec/ab",
        &a,
        &[("notes/ab.md", "ab\n")],
        SPEC,
        &[1, 2],
        true,
    );
    let (c_tip, c_ab, c_a) = (w.ready_card(&tip), w.ready_card(&ab), w.ready_card(&a));

    let (v, code) = w.build(&[&c_tip, &c_ab, &c_a]);
    assert_eq!(code, 0, "{v}");
    let r = &v["result"];
    assert_eq!(r["ok"], true, "skipping is not a failure: {v}");
    assert_eq!(names(&v, "merged"), vec![c_ab.clone()]);
    assert_eq!(
        r["skipped"],
        serde_json::json!([
            {"card": c_tip, "head": tip, "reason": "in_tip"},
            {"card": c_a, "head": a, "reason": "in_branch"},
        ])
    );
    let pinned = r["pinned"].as_str().unwrap();
    assert_eq!(
        w.git(&[
            "rev-list",
            "--merges",
            "--count",
            &format!("{tip}..{pinned}")
        ]),
        "1"
    );
}

#[test]
fn the_number_is_the_lowest_unused_and_existing_names_are_never_touched() {
    let w = World::new();
    w.identity();
    let a = w.head("spec/a", TIP, &[("notes/a.md", "a\n")], SPEC, &[1, 2], true);
    let c = w.ready_card(&a);
    let name = |n: u32| format!("m-flow/bundle-{DATE}-{n}");
    // -1: a local branch; -2: an `origin` branch; -3: a folder; -4: a registered worktree.
    w.git(&["branch", &name(1), TIP]);
    w.git(&[
        "push",
        "-q",
        "origin",
        &format!("trunk:refs/heads/{}", name(2)),
    ]);
    w.git(&["fetch", "-q", "origin"]);
    fs::create_dir_all(w.worktree_of(3)).unwrap();
    let other = w.worktree_of(4);
    fs::create_dir_all(other.parent().unwrap()).unwrap();
    w.git(&[
        "worktree",
        "add",
        "-q",
        "-b",
        "elsewhere",
        other.to_str().unwrap(),
        TIP,
    ]);
    let refs_before = w.refs();
    let worktrees_before = w.git(&["worktree", "list", "--porcelain"]);

    let (v, code) = w.build(&[&c]);
    assert_eq!(code, 0, "{v}");
    assert_eq!(v["result"]["number"], 5, "{v}");
    assert_eq!(v["result"]["branch"], name(5).as_str());
    // Every earlier ref is byte-equal; the only new ref is the new branch.
    let refs_after = w.refs();
    for line in &refs_before {
        assert!(refs_after.contains(line), "{line} changed");
    }
    assert_eq!(refs_after.len(), refs_before.len() + 1);
    assert!(refs_after
        .iter()
        .any(|l| l.ends_with(&format!("refs/heads/{}", name(5)))));
    assert!(w.worktree_of(3).is_dir(), "the folder of -3 is still there");
    assert!(w
        .git(&["worktree", "list", "--porcelain"])
        .starts_with(&worktrees_before.split("\n\n").next().unwrap().to_string()));
    // A second build goes on to the next number.
    let (v2, _) = w.build(&[&c]);
    assert_eq!(v2["result"]["number"], 6, "{v2}");
    // The folder of -3 was never written into: `remove_dir` succeeds only on an empty folder.
    fs::remove_dir(w.worktree_of(3)).expect("the folder of -3 is still empty");
}

#[test]
fn creation_refuses_a_taken_branch_or_folder() {
    let w = World::new();
    let tip = w.tip_sha();
    let folder = w
        .home
        .path()
        .join(".moai")
        .join("worktrees")
        .join("work-bundle-9");
    let other = w
        .home
        .path()
        .join(".moai")
        .join("worktrees")
        .join("work-bundle-10");
    moai_cli::bundle::build::create(&w.work, "m-flow/x-1", &folder, &tip).unwrap();
    let refs = w.refs();
    // A taken branch with a new folder: git refuses `-b`, nothing changes.
    let e = moai_cli::bundle::build::create(&w.work, "m-flow/x-1", &other, &tip).unwrap_err();
    assert!(e.contains("nothing was changed"), "{e}");
    assert!(!other.exists());
    assert_eq!(w.refs(), refs);
    // A taken folder with a new branch: refused before git runs.
    let e = moai_cli::bundle::build::create(&w.work, "m-flow/x-2", &folder, &tip).unwrap_err();
    assert!(e.contains("already exists"), "{e}");
    assert_eq!(w.refs(), refs);
    // A taken branch that is checked out nowhere and points elsewhere: `-b` refuses it, where
    // `-B` would move it to the tip (audit N3 of slice 3).
    let a = w.head(
        "spec/a",
        TIP,
        &[("notes/a.md", "a\n")],
        SPEC,
        &[1, 2],
        false,
    );
    w.git(&["branch", "m-flow/x-3", &a]);
    let third = w
        .home
        .path()
        .join(".moai")
        .join("worktrees")
        .join("work-bundle-11");
    let refs = w.refs();
    let e = moai_cli::bundle::build::create(&w.work, "m-flow/x-3", &third, &tip).unwrap_err();
    assert!(e.contains("nothing was changed"), "{e}");
    assert!(!third.exists());
    assert_eq!(w.refs(), refs);
    assert_eq!(
        w.git(&["rev-parse", "m-flow/x-3"]),
        a,
        "the branch did not move"
    );
}

#[test]
fn a_second_registered_merge_driver_does_not_decide() {
    let w = World::new();
    w.identity();
    // A driver named `ours` (it always succeeds and keeps our side) chosen by a .gitattributes
    // line of the tip: a plain merge hides the conflict (audit N1 of slice 3).
    w.git(&["config", "merge.ours.driver", "true"]);
    w.put(
        ".gitattributes",
        &format!("{MANIFEST} merge=moai-manifest\nq.txt merge=ours\n"),
    );
    w.put("q.txt", "base\n");
    w.git(&["add", "-A"]);
    w.git(&["commit", "-q", "-m", "chore: q.txt with its own driver"]);
    w.git(&["push", "-q", "origin", "trunk"]);
    w.git(&["fetch", "-q", "origin"]);
    let p = w.head("spec/p", TIP, &[("q.txt", "P\n")], SPEC, &[1, 2], true);
    let q = w.head("spec/q", TIP, &[("q.txt", "Q\n")], SPEC, &[1, 2], true);
    let hidden = w.git_raw(&w.work, &["merge-tree", "--write-tree", &p, &q]);
    assert_eq!(hidden.status.code(), Some(0), "the driver `ours` hides it");
    let (cp, cq) = (w.ready_card(&p), w.ready_card(&q));

    let (v, code) = w.build(&[&cp, &cq]);
    assert_eq!(code, 0, "{v}");
    assert_eq!(v["result"]["ok"], false, "{v}");
    assert_eq!(names(&v, "merged"), vec![cp.clone()]);
    assert_eq!(
        v["result"]["dropped"],
        serde_json::json!([{"card": cq, "head": q, "files": ["q.txt"]}])
    );
    // `bundle check` overrides the clone's drivers too.
    let (v, _) = w.bundle(&["check", &cp, &cq]);
    let conflicts: Vec<&Value> = v["result"]["findings"]
        .as_array()
        .unwrap()
        .iter()
        .filter(|f| f["code"] == "conflict")
        .collect();
    assert_eq!(conflicts.len(), 1, "{v}");
    assert_eq!(
        conflicts[0]["detail"]["files"],
        serde_json::json!(["q.txt"])
    );
}

#[test]
fn a_date_that_is_not_a_date_is_refused_and_a_tag_head_is_the_commit() {
    let w = World::new();
    w.identity();
    let a = w.head("spec/a", TIP, &[("notes/a.md", "a\n")], SPEC, &[1, 2], true);
    let c = w.ready_card(&a);
    let before = w.repo_state();
    for d in [
        "20261340",
        "20260132",
        "00000000",
        "20260000",
        "2026111a",
        "2026111",
        "202611190",
        "",
    ] {
        let (v, code) = w.bundle(&["build", &c, "--date", d]);
        assert_eq!(code, 1, "{d}: {v}");
        let m = v["error"]["message"].as_str().unwrap_or("");
        assert!(m.contains("--date"), "{d}: {v}");
    }
    assert_eq!(w.repo_state(), before, "nothing was created");
    assert!(!w.home.path().join(".moai").join("worktrees").exists());
    // The edges are dates: 12 and 31 are accepted. The worktree folder has no date in its name,
    // so the second build takes the number 2.
    for (d, n) in [("20261231", 1), ("20260131", 2)] {
        let (v, code) = w.bundle(&["build", &c, "--date", d]);
        assert_eq!(code, 0, "{d}: {v}");
        assert_eq!(v["result"]["number"], n, "{d}: {v}");
        assert_eq!(
            v["result"]["branch"],
            format!("m-flow/bundle-{d}-{n}").as_str(),
            "{v}"
        );
    }
    // A tag object's SHA as a head: the commit the tag points to is merged and reported.
    w.git(&["tag", "-a", "-m", "a tag", "v-a", &a]);
    let tag = w.git(&["rev-parse", "v-a"]);
    assert_ne!(tag, a);
    let c2 = w.new_card();
    let (v, code) = w.build(&[&format!("{c2}={tag}")]);
    assert_eq!(code, 0, "{v}");
    assert_eq!(v["result"]["merged"][0]["head"], a.as_str(), "{v}");
    let pinned = v["result"]["pinned"].as_str().unwrap();
    assert_eq!(w.git(&["rev-parse", &format!("{pinned}^2")]), a);
}

#[test]
fn a_registered_worktree_whose_folder_is_gone_still_takes_its_number() {
    let w = World::new();
    w.identity();
    let a = w.head("spec/a", TIP, &[("notes/a.md", "a\n")], SPEC, &[1, 2], true);
    let c = w.ready_card(&a);
    let gone = w.worktree_of(1);
    fs::create_dir_all(gone.parent().unwrap()).unwrap();
    w.git(&[
        "worktree",
        "add",
        "-q",
        "-b",
        "other",
        gone.to_str().unwrap(),
        TIP,
    ]);
    fs::remove_dir_all(&gone).unwrap();
    assert!(w
        .git(&["worktree", "list", "--porcelain"])
        .contains("prunable"));
    // The branch `-1` does not exist and the folder is gone; only the registration holds the
    // number, and `git worktree add` would refuse it.
    let (v, code) = w.build(&[&c]);
    assert_eq!(code, 0, "{v}");
    assert_eq!(v["result"]["number"], 2, "{v}");
}

#[test]
fn all_999_numbers_used_is_refused_before_anything_is_created() {
    let w = World::new();
    w.identity();
    let tip = w.tip_sha();
    let a = w.head("spec/a", TIP, &[("notes/a.md", "a\n")], SPEC, &[1, 2], true);
    let c = w.ready_card(&a);
    // 998 loose branch refs written as files: much cheaper than 998 `git branch` runs.
    let dir = w.work.join(".git/refs/heads/m-flow");
    fs::create_dir_all(&dir).unwrap();
    for n in 1..=998 {
        fs::write(dir.join(format!("bundle-{DATE}-{n}")), format!("{tip}\n")).unwrap();
    }
    // The 999th number is still free and is taken (the edge: `1..=999`, not `1..=998`).
    let (v, code) = w.build(&[&c]);
    assert_eq!(code, 0, "{v}");
    assert_eq!(v["result"]["number"], 999, "{v}");
    // Now all 999 are used: refused, nothing created.
    let before = w.refs();
    let (v, code) = w.build(&[&c]);
    assert_eq!(code, 1, "{v}");
    assert!(
        v["error"]["message"]
            .as_str()
            .unwrap()
            .contains("the numbers 1 to 999 of this date are all used"),
        "{v}"
    );
    assert_eq!(w.refs(), before);
    assert!(!w.worktree_of(1).exists());
    assert!(w.worktree_of(999).exists());
    w.git(&[
        "worktree",
        "remove",
        "--force",
        w.worktree_of(999).to_str().unwrap(),
    ]);
}

#[test]
fn the_rebuild_line_is_quoted_and_empty_when_no_head_is_left() {
    let w = World::new();
    w.identity();
    // A tip with a shell character in its name, and a manifest change of its own.
    w.git(&["checkout", "-q", "-B", "odd;tip", TIP]);
    w.put(MANIFEST, "{\n  \"line\": \"tip-side\"\n}\n");
    w.git(&["add", "-A"]);
    w.git(&[
        "commit",
        "-q",
        "-m",
        "chore: the odd tip changes the manifest",
    ]);
    let b = w.head(
        "spec/b",
        TIP,
        &[(MANIFEST, "{\n  \"line\": \"b-side\"\n}\n")],
        SPEC,
        &[1, 2],
        true,
    );
    let e = w.head("spec/e", TIP, &[("notes/e.md", "e\n")], SPEC, &[1, 2], true);
    let (cb, ce) = (w.ready_card(&b), w.ready_card(&e));
    let build = |heads: &[&str]| -> Value {
        let mut all = vec!["bundle", "build"];
        all.extend_from_slice(heads);
        all.extend_from_slice(&[
            "--date",
            DATE,
            "--project",
            &w.project,
            "--dir",
            w.work.to_str().unwrap(),
            "--tip",
            "odd;tip",
        ]);
        let out = w.home.command(&all).stdin(Stdio::null()).output().unwrap();
        let line = String::from_utf8_lossy(&out.stdout)
            .lines()
            .next()
            .unwrap_or("")
            .to_string();
        serde_json::from_str(&line).unwrap_or_else(|e| panic!("{e}: {line}"))
    };
    // b conflicts with the tip, e does not: `rebuild` is e alone, the tip word quoted.
    let v = build(&[&cb, &ce]);
    assert_eq!(v["result"]["ok"], false, "{v}");
    assert_eq!(
        v["result"]["rebuild"],
        format!("moai bundle build {ce}={e} --tip 'odd;tip'").as_str(),
        "{v}"
    );
    // The conflicting head and a head already in the tip: only the in-tip one would be left, and
    // that build is refused ("no head left"), so `rebuild` is empty too.
    let cx = w.new_card();
    let trunk = w.tip_sha();
    let v = build(&[&cb, &format!("{cx}={trunk}")]);
    assert_eq!(v["result"]["ok"], false, "{v}");
    assert_eq!(v["result"]["rebuild"], "", "{v}");
    // Only the conflicting head: nothing would be left, so `rebuild` is empty.
    let v = build(&[&cb]);
    assert_eq!(v["result"]["ok"], false, "{v}");
    assert_eq!(v["result"]["merged"], serde_json::json!([]));
    assert_eq!(v["result"]["rebuild"], "", "{v}");
}

#[test]
fn a_git_failure_after_creation_names_what_is_left_behind() {
    let w = World::new();
    // No author identity anywhere: git cannot write the merge commit.
    w.git(&["config", "user.useConfigOnly", "true"]);
    let tip = w.tip_sha();
    let a = w.head("spec/a", TIP, &[("notes/a.md", "a\n")], SPEC, &[1, 2], true);
    let c = w.ready_card(&a);

    let (v, code) = w.build(&[&c]);
    assert_eq!(code, 1, "{v}");
    let m = v["error"]["message"].as_str().unwrap();
    let branch = format!("m-flow/bundle-{DATE}-1");
    assert!(m.contains("bundle build stopped"), "{m}");
    assert!(m.contains(&format!("the branch `{branch}`")), "{m}");
    assert!(m.contains("work-bundle-1"), "{m}");
    assert!(m.contains("nothing was deleted"), "{m}");
    // Both still exist, the branch at the tip, and the worktree is clean (the merge was aborted).
    let wt = w.worktree_of(1);
    assert!(wt.join(".git").exists());
    assert_eq!(w.git(&["rev-parse", &format!("refs/heads/{branch}")]), tip);
    assert_eq!(w.git_in(&wt, &["status", "--porcelain"]), "");
}

#[test]
fn build_changes_only_the_new_branch_and_worktree() {
    let w = World::new();
    w.identity();
    let a = w.head("spec/a", TIP, &[("notes/a.md", "a\n")], SPEC, &[1, 2], true);
    let b = w.head("spec/b", TIP, &[("notes/b.md", "b\n")], SPEC, &[1, 2], true);
    let (c1, c2) = (w.ready_card(&a), w.ready_card(&b));
    w.git(&["checkout", "-q", "trunk"]);
    let refs = w.refs();
    let origin = w.origin_refs();
    let status = w.git(&["status", "--porcelain"]);
    let head = w.git(&["rev-parse", "HEAD"]);
    let ledger = w.ledger_view();

    let (v, code) = w.build(&[&c1, &c2]);
    assert_eq!(code, 0, "{v}");
    let after = w.refs();
    let new: Vec<&String> = after.iter().filter(|l| !refs.contains(l)).collect();
    assert_eq!(new.len(), 1, "{new:?}");
    assert!(
        new[0].ends_with(&format!("refs/heads/m-flow/bundle-{DATE}-1")),
        "{new:?}"
    );
    assert_eq!(after.len(), refs.len() + 1, "no ref was removed or changed");
    assert_eq!(w.origin_refs(), origin, "nothing was pushed");
    assert_eq!(w.git(&["status", "--porcelain"]), status);
    assert_eq!(
        w.git(&["rev-parse", "HEAD"]),
        head,
        "the clone's HEAD did not move"
    );
    assert_eq!(w.ledger_view(), ledger, "no card note, no ledger row");
}

/// REQ-11, audit B2: an exclude file that cannot be written leaves the build done, names the
/// error under `place_exclude_error`, and the file the link points at keeps its bytes.
#[cfg(unix)]
#[test]
fn worktree_place_status_the_bundle_names_an_unwritable_exclude_file() {
    let w = World::new();
    w.identity();
    let a = w.head("spec/a", TIP, &[("notes/a.md", "a\n")], SPEC, &[1, 2], true);
    let c1 = w.ready_card(&a);
    w.git(&["checkout", "-q", "trunk"]);
    let outside = w.work.parent().unwrap().join("outside-exclude");
    std::fs::write(&outside, "keep\n").unwrap();
    let exclude = w.work.join(".git/info/exclude");
    let _ = std::fs::remove_file(&exclude);
    std::os::unix::fs::symlink(&outside, &exclude).unwrap();
    let (v, code) = w.build(&[&c1]);
    assert_eq!(code, 0, "{v}");
    let err = v["result"]["place_exclude_error"]
        .as_str()
        .unwrap_or_default();
    assert!(err.contains("exclude"), "{v}");
    assert!(w.worktree_of(1).join("notes/a.md").is_file(), "{v}");
    assert_eq!(std::fs::read_to_string(&outside).unwrap(), "keep\n");
}

/// A bundle build that works writes one exclude line, and an answer without an error has no
/// `place_exclude_error` key.
#[test]
fn worktree_place_status_the_bundle_writes_the_exclude_line_once() {
    let w = World::new();
    w.identity();
    let a = w.head("spec/a", TIP, &[("notes/a.md", "a\n")], SPEC, &[1, 2], true);
    let b = w.head("spec/b", TIP, &[("notes/b.md", "b\n")], SPEC, &[1, 2], true);
    let (c1, c2) = (w.ready_card(&a), w.ready_card(&b));
    w.git(&["checkout", "-q", "trunk"]);
    let (v, code) = w.build(&[&c1]);
    assert_eq!(code, 0, "{v}");
    assert!(v["result"].get("place_exclude_error").is_none(), "{v}");
    let (v2, code2) = w.build(&[&c2]);
    assert_eq!(code2, 0, "{v2}");
    let lines = std::fs::read_to_string(w.work.join(".git/info/exclude"))
        .unwrap_or_default()
        .lines()
        .filter(|l| *l == "/.claude/worktrees/")
        .count();
    assert_eq!(lines, 1);
}

/// REQ-13: a `.claude` that is a link makes the build refuse before anything is made: no new
/// ref, nothing behind the link, and the refusal names the link.
#[cfg(unix)]
#[test]
fn worktree_place_links_the_bundle_refuses_a_linked_claude_folder() {
    let w = World::new();
    w.identity();
    let a = w.head("spec/a", TIP, &[("notes/a.md", "a\n")], SPEC, &[1, 2], true);
    let c1 = w.ready_card(&a);
    w.git(&["checkout", "-q", "trunk"]);
    let outside = w.work.parent().unwrap().join("outside-place");
    std::fs::create_dir_all(&outside).unwrap();
    std::os::unix::fs::symlink(&outside, w.work.join(".claude")).unwrap();
    let refs = w.refs();
    let (v, code) = w.build(&[&c1]);
    assert_ne!(code, 0, "{v}");
    assert!(v.to_string().contains(".claude"), "{v}");
    assert_eq!(std::fs::read_dir(&outside).unwrap().count(), 0);
    assert_eq!(w.refs(), refs, "no ref was made");
}

#[test]
fn the_help_and_the_ops_doc_say_what_build_writes() {
    let w = World::new();
    let out = w
        .home
        .command(&["bundle", "build", "--help"])
        .output()
        .unwrap();
    let help = String::from_utf8_lossy(&out.stdout).to_string();
    assert!(
        help.lines()
            .next()
            .unwrap_or("")
            .starts_with("Writes a new branch and worktree."),
        "{help}"
    );
    let doc = fs::read_to_string(
        Path::new(env!("CARGO_MANIFEST_DIR"))
            .join("docs/design-review/appendix/operations/bundle.md"),
    )
    .unwrap();
    for needle in [
        "bundle build HEADS",
        "m-flow/bundle-",
        "`dropped`",
        "`rebuild`",
        "`skipped`",
        "`pinned`",
    ] {
        assert!(doc.contains(needle), "the operations page lacks {needle}");
    }
}
