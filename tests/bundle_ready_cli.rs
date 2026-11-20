//! `moai bundle ready` (SPEC-20261119-bundle-ready-check, card C-119): the READY heads of a
//! temporary ledger, read against a temporary repository with a bare `origin`. Every command runs
//! against a temporary HOME; the real `~/.moai` and the real repository are not read or written.
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

    /// Overwrites the project profile with a valid one carrying the work flow the
    /// scene wants; `home.project` writes a minimal file that `read_profile`
    /// refuses, so every scene that needs a readable profile calls this.
    fn write_profile(&self, work_flow: Option<Value>) {
        let mut profile = serde_json::json!({
            "schema_version": 2, "project_id": "demo", "purpose": "Fixture",
            "tools": ["Rust"], "development_methods": ["TDD"], "hosts": ["claude"],
            "decisions": [{"id": "d1", "topic": "tools", "value": "Rust",
                "source": "user", "scope": "project"}],
        });
        if let Some(wf) = work_flow {
            profile["work_flow"] = wf;
        }
        let path = Path::new(&self.project)
            .join(".moai-cli")
            .join("project.yaml");
        fs::write(path, profile.to_string()).unwrap();
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
        self.bundle_args(args, true)
    }

    /// `with_tip=false` leaves `--tip` out, so the action resolves the tip itself —
    /// the profile-resolved default the slice pins.
    fn bundle_args(&self, args: &[&str], with_tip: bool) -> (Value, i32) {
        let mut all = vec!["bundle"];
        all.extend_from_slice(args);
        all.extend_from_slice(&[
            "--project",
            &self.project,
            "--dir",
            self.work.to_str().unwrap(),
        ]);
        if with_tip {
            all.extend_from_slice(&["--tip", TIP]);
        }
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
fn ready_lists_the_newest_readable_line_per_card_and_names_the_unreadable_ones() {
    let w = World::new();
    let h1 = w.head("spec/one", TIP, &[], SPEC, &[1, 2], true);
    let h2 = w.head(
        "spec/two",
        TIP,
        &[("notes/two.md", "two\n")],
        SPEC,
        &[1, 2],
        true,
    );
    let h3 = w.head(
        "spec/three",
        TIP,
        &[("notes/three.md", "three\n")],
        SPEC,
        &[1, 2],
        false,
    );
    let absent = "f".repeat(40);
    let cards: Vec<String> = (0..6).map(|_| w.new_card()).collect();
    let [c1, c2, c3, c4, c5, c6] = &cards[..] else {
        unreachable!()
    };

    // C-1: an old line, then a repair line with another head: the newest wins.
    w.ready_note(c1, &h1);
    let newest = w.note(
        c1,
        &format!("LANE-READY-INTEGRATION {c1} head={h2} — repaired"),
    );
    // C-2: a head that is only here, with a spec= value.
    w.ready_note(c2, &h3);
    // C-3: a readable line, then a newer one that does not read: no head.
    w.ready_note(c3, &h1);
    let bad3 = w.note(
        c3,
        &format!("LANE-READY-INTEGRATION agent-1 {c3} branch x head {h1}"),
    );
    // C-4: the marker is not the first line: not a READY line.
    w.note(
        c4,
        &format!("see below\nLANE-READY-INTEGRATION {c4} head={h1}"),
    );
    // C-5: the line names another card.
    let bad5 = w.note(c5, &format!("LANE-READY-INTEGRATION {c1} head={h1}"));
    // C-6: a head this repository does not have.
    w.note(c6, &format!("LANE-READY-INTEGRATION {c6} head={absent}"));

    let (v, code) = w.bundle(&["ready"]);
    assert_eq!(code, 0, "{v}");
    let r = &v["result"];
    assert_eq!(r["action"], "ready");
    assert_eq!(r["writes"], false);
    assert_eq!(r["tip"]["sha"], w.tip_sha());
    let heads = r["heads"].as_array().unwrap();
    let of = |card: &str| heads.iter().find(|h| h["card"] == card);

    let e1 = of(c1).expect("C-1 listed");
    assert_eq!(e1["head"], h2.as_str());
    assert_eq!(e1["note_id"], newest);
    assert_eq!(e1["ready_lines"], 2);
    assert_eq!(e1["specs"], serde_json::json!([SPEC]));
    assert_eq!(e1["spec"], Value::Null);
    assert_eq!(
        (e1["present"].clone(), e1["on_origin"].clone()),
        (true.into(), true.into())
    );
    assert_eq!(e1["already_in_tip"], false);

    let e2 = of(c2).expect("C-2 listed");
    assert_eq!(e2["head"], h3.as_str());
    assert_eq!(e2["spec"], SPEC);
    assert_eq!(
        (e2["present"].clone(), e2["on_origin"].clone()),
        (true.into(), false.into())
    );

    let e6 = of(c6).expect("C-6 listed");
    assert_eq!(e6["head"], absent.as_str());
    assert_eq!(
        (e6["present"].clone(), e6["on_origin"].clone()),
        (false.into(), false.into())
    );
    assert_eq!(e6["specs"], serde_json::json!([]));

    for gone in [c3, c4, c5] {
        assert!(of(gone).is_none(), "{gone} must give no head: {heads:?}");
    }
    assert_eq!(
        r["unreadable"],
        serde_json::json!([{"card": c3, "note_id": bad3}, {"card": c5, "note_id": bad5}])
    );
    assert_eq!(r["count"], 3);
}

#[test]
fn a_ledger_without_ready_lines_lists_nothing() {
    let w = World::new();
    let c = w.new_card();
    w.note(&c, "a note that is not a READY line");
    let (v, code) = w.bundle(&["ready"]);
    assert_eq!(code, 0, "{v}");
    assert_eq!(v["result"]["heads"], serde_json::json!([]));
    assert_eq!(v["result"]["unreadable"], serde_json::json!([]));
    assert_eq!(v["result"]["count"], 0);
}

/// SPEC-20261120-mflow-integration-branch-choice: with no `--tip`, the action
/// resolves the tip from the project profile's m-flow integration branch —
/// the stored value drives the integration ref, so a repository that named
/// its branch at first setup needs no flag at all.
#[test]
fn the_tip_comes_from_the_profile_integration_branch() {
    let w = World::new();
    w.write_profile(Some(serde_json::json!({
        "flow": "m-flow",
        "integration_branch": "trunk",
    })));
    let (v, code) = w.bundle_args(&["ready"], false);
    assert_eq!(code, 0, "{v}");
    assert_eq!(
        v["result"]["tip"]["name"], "origin/trunk",
        "the stored branch drives the tip: {v}"
    );
}

/// With no recorded integration branch the old default stays — visible here as
/// the refusal naming it, because the fixture repository has no such branch.
#[test]
fn without_a_recorded_branch_the_old_tip_default_stays() {
    let w = World::new();
    w.write_profile(None);
    let (v, code) = w.bundle_args(&["ready"], false);
    assert_eq!(code, 1, "no recorded branch keeps the old default: {v}");
    let message = v["error"]["message"].as_str().unwrap();
    assert!(
        message.contains("origin/v2026.11.0"),
        "the refusal names the old default: {v}"
    );
}

/// The tip resolution is an m-flow rule: a github-flow profile does not point the
/// tip at origin/<integration_branch> even when one is recorded — without --tip it
/// keeps the old default (audit B1: the non-m-flow branch of the resolution had no
/// scene of its own).
#[test]
fn a_github_flow_profile_keeps_the_old_tip_default() {
    let w = World::new();
    w.write_profile(Some(
        serde_json::json!({"flow": "github-flow", "integration_branch": "trunk"}),
    ));
    let (v, code) = w.bundle_args(&["ready"], false);
    assert_eq!(code, 1, "github-flow keeps the old default: {v}");
    let message = v["error"]["message"].as_str().unwrap();
    assert!(
        message.contains("origin/v2026.11.0"),
        "the refusal names the old default, not the recorded branch: {v}"
    );
}
