//! SDD-8 (SPEC-20261118-spec-close): `moai spec close <SPEC-ID>` closes an approved SPEC whose
//! work is merged: its status line becomes `completed`, a linked card is closed through
//! `card done`, and the commit is printed, never run. Every refusal leaves every byte under the
//! temporary HOME (project, `.git`, ledger) as it was. A temporary repository, a temporary
//! ledger and a stub `gh` only: never the network, the real HOME or the real ledger.

mod common;

use common::home::TempHome;
use moai_cli::ledger;
use serde_json::{json, Value};
use std::fs;
use std::path::PathBuf;
use std::process::Output;

const ID: &str = "SPEC-20261117-login-fix";
const BRANCH: &str = "spec/20261117-login-fix";
const SESSION_VAR: &str = "CLAUDE_CODE_SESSION_ID";
const S_A: &str = "aaaaaaaa-0000-4000-8000-00000000000a";
const MERGE: &str = "0123456789abcdef0123456789abcdef01234567";
const HEAD: &str = "89abcdef0123456789abcdef0123456789abcdef";
const MET: &str = "\n## Result\n\n- AC-1: met — the test passes\n";

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

fn refused(out: &Output) -> String {
    let v = envelope(out);
    assert_eq!(out.status.code(), Some(1), "{v}");
    v["error"]["message"]
        .as_str()
        .unwrap_or_default()
        .to_string()
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

/// A HOME with the registered ledger of project `demo`, a git repository there on `main`, and
/// a stub `gh` outside the HOME that answers pull request reads from files.
struct World {
    home: TempHome,
    project: PathBuf,
    _host: tempfile::TempDir,
    dir: PathBuf,
    gh: PathBuf,
}

impl World {
    fn new() -> Self {
        let home = TempHome::new();
        let project = home.project("demo");
        ledger::register(home.path(), "demo").unwrap();
        let host = tempfile::Builder::new()
            .prefix("moai-test-host-")
            .tempdir()
            .unwrap();
        let dir = host.path().to_path_buf();
        let script = format!(
            "#!/bin/sh\nurl=''\nfor a in \"$@\"; do case \"$a\" in https://*) url=\"$a\";; esac; done\n\
             case \"$url\" in\n\
               */pulls/*) n=${{url##*/pulls/}}; if [ -f '{dir}/pull-'$n'.json' ]; then printf 'HTTP/2.0 200 OK\\r\\n\\r\\n'; cat '{dir}/pull-'$n'.json'; exit 0; fi;;\n\
             esac\n\
             printf 'HTTP/2.0 404 Not Found\\r\\n\\r\\n{{}}'; exit 1\n",
            dir = dir.display()
        );
        let gh = dir.join("gh-stub");
        common::write_executable(&gh, &script);
        let w = World {
            home,
            project,
            _host: host,
            dir,
            gh,
        };
        w.git(&["init", "-q", "-b", "main"]);
        fs::write(w.project.join("a.txt"), "a\n").unwrap();
        w.git(&["add", "-A"]);
        w.commit("init");
        w
    }

    fn p(&self) -> &str {
        self.project.to_str().unwrap()
    }

    fn git(&self, args: &[&str]) -> String {
        common::git::git(&self.project, args)
    }

    fn commit(&self, message: &str) {
        self.git_as(&["commit", "-q", "-m", message]);
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

    fn run(&self, args: &[&str], session: Option<&str>) -> Output {
        let mut cmd = self.home.command(args);
        cmd.env("MOAI_FORGE_GH_BIN", &self.gh);
        if let Some(s) = session {
            cmd.env(SESSION_VAR, s);
        }
        cmd.output().expect("run moai")
    }

    /// `moai card <args> --project P` as `S_A`.
    fn card(&self, args: &[&str]) -> Output {
        let mut all = vec!["card"];
        all.extend_from_slice(args);
        all.extend_from_slice(&["--project", self.p()]);
        self.run(&all, Some(S_A))
    }

    /// `moai spec close <id> --project P <extra>` as `S_A`.
    fn close(&self, id: &str, extra: &[&str]) -> Output {
        let mut all = vec!["spec", "close", id, "--project", self.p()];
        all.extend_from_slice(extra);
        self.run(&all, Some(S_A))
    }

    fn spec_file(&self, id: &str) -> PathBuf {
        self.project.join(".moai/specs").join(id).join("spec.md")
    }

    /// Writes SPEC `SPEC-20261117-<slug>` with placeholders filled, approved when `approved`,
    /// with `result` appended. Returns its id.
    fn write_spec(&self, slug: &str, approved: bool, result: &str) -> String {
        let r = ok(&self.run(
            &[
                "spec",
                "new",
                slug,
                "--title",
                "Login fix",
                "--minimal",
                "--date",
                "2026-11-17",
                "--project",
                self.p(),
            ],
            None,
        ));
        let id = r["id"].as_str().unwrap().to_string();
        let path = self.spec_file(&id);
        fs::write(&path, filled(&fs::read_to_string(&path).unwrap())).unwrap();
        if approved {
            let shown = ok(&self.run(&["spec", "show", &id, "--project", self.p()], None));
            let digest = shown["spec"]["digest"].as_str().unwrap().to_string();
            let text = fs::read_to_string(&path).unwrap().replace(
                "status: draft\n",
                &format!(
                    "status: approved\napproval:\n  by: user\n  at: 2026-11-17\n  basis: test \
                     selection\n  digest: {digest}\n"
                ),
            );
            fs::write(&path, text).unwrap();
        }
        let mut text = fs::read_to_string(&path).unwrap();
        text.push_str(result);
        fs::write(&path, text).unwrap();
        ok(&self.run(&["spec", "check", &id, "--project", self.p()], None));
        id
    }

    /// The SPEC written and committed on its branch with its trailer, then merged into `main`.
    fn merged(&self, approved: bool, result: &str) {
        self.git(&["switch", "-q", "-c", BRANCH]);
        self.write_spec("login-fix", approved, result);
        self.git(&["add", "-A"]);
        self.commit(&format!("docs(spec): login fix\n\nMoai-Spec: {ID}"));
        self.git(&["switch", "-q", "main"]);
        self.git_as(&["merge", "-q", "--no-ff", "-m", "Merge login fix", BRANCH]);
    }

    /// What the host says about pull request `number` from now on.
    fn host_pr(&self, number: u64, merged: bool) {
        let pull = json!({
            "number": number, "state": if merged { "closed" } else { "open" }, "merged": merged,
            "merged_at": if merged { json!("2026-11-17T00:00:00Z") } else { Value::Null },
            "merge_commit_sha": if merged { json!(MERGE) } else { Value::Null },
            "head": {"sha": HEAD},
        });
        fs::write(
            self.dir.join(format!("pull-{number}.json")),
            pull.to_string(),
        )
        .unwrap();
    }

    /// Card C-1, taken by `S_A` at generation 1, its one criterion met, pull request 7 linked
    /// (merged when `merged`), in `review`, naming SPEC `spec` when given.
    fn review_card(&self, spec: Option<&str>, merged: bool) {
        ok(&self.card(&[
            "add", "--title", "t", "--kind", "fix", "--scope", "src/", "--ac", "it works",
        ]));
        ok(&self.card(&["ready", "C-1"]));
        ok(&self.card(&["take", "C-1"]));
        if let Some(id) = spec {
            ok(&self.card(&["spec", "C-1", id]));
        }
        ok(&self.card(&[
            "criterion",
            "C-1",
            "1",
            "--generation",
            "1",
            "--met",
            "--evidence",
            "test passed",
        ]));
        self.host_pr(7, merged);
        ok(&self.card(&["pr", "C-1", "7", "--generation", "1", "--repo", "octo/demo"]));
        ok(&self.card(&["review", "C-1", "--generation", "1"]));
    }

    /// Every path and byte under the temporary HOME: the project, its `.git` and the ledger.
    fn everything(&self) -> Vec<(String, Vec<u8>)> {
        common::cli::tree_state(self.home.path())
    }
}

#[test]
fn close_completes_a_merged_spec_and_prints_the_commit() {
    let w = World::new();
    w.merged(true, MET);
    let head = w.git(&["rev-parse", "HEAD"]);
    let before = fs::read_to_string(w.spec_file(ID)).unwrap();
    let r = ok(&w.close(ID, &[]));
    assert_eq!(r["action"], "close");
    assert_eq!(r["spec"], ID);
    assert_eq!(
        (r["from"].clone(), r["status"].clone()),
        ("approved".into(), "completed".into())
    );
    assert_eq!(r["base"], "main");
    assert_eq!(r["writes"], true);
    assert!(r["card"].is_null(), "{r}");
    // Only the status line changed, and the file now passes the closure check.
    let after = fs::read_to_string(w.spec_file(ID)).unwrap();
    assert_eq!(
        after,
        before.replacen("status: approved\n", "status: completed\n", 1)
    );
    ok(&w.run(
        &["spec", "check", ID, "--closure", "--project", w.p()],
        None,
    ));
    // Nothing was committed; the commit is a preview naming the file and the trailer.
    assert_eq!(w.git(&["rev-parse", "HEAD"]), head);
    let path = format!(".moai/specs/{ID}/spec.md");
    assert_eq!(w.git(&["status", "--porcelain"]), format!("M {path}"));
    let next: Vec<&str> = r["next"]
        .as_array()
        .unwrap()
        .iter()
        .map(|v| v.as_str().unwrap())
        .collect();
    assert_eq!(next[0], format!("moai trailers --spec {ID}"), "{r}");
    assert!(next[1].starts_with("git commit -m "), "{r}");
    assert!(next[1].contains(&format!("Moai-Spec: {ID}")), "{r}");
    assert!(next[1].ends_with(&format!(" -- {path}")), "{r}");
    // Closed once: a second close is refused and changes nothing.
    let before = w.everything();
    let m = refused(&w.close(ID, &[]));
    assert!(m.contains("completed"), "{m}");
    assert_eq!(w.everything(), before);
}

#[test]
fn close_refuses_and_changes_nothing() {
    // Each case: a world, the id, extra arguments and a word the message must hold.
    type Case = (
        &'static str,
        World,
        &'static str,
        Vec<&'static str>,
        &'static str,
    );
    let mut cases: Vec<Case> = Vec::new();
    let w = World::new();
    w.merged(true, MET);
    cases.push((
        "missing",
        w,
        "SPEC-20261117-nothing-here",
        vec![],
        "not found",
    ));
    let w = World::new();
    w.merged(true, MET);
    cases.push((
        "go id",
        w,
        "SPEC-AUTH-001",
        vec![],
        "SPEC-<yyyymmdd>-<slug>",
    ));
    let w = World::new();
    w.merged(false, MET);
    cases.push(("draft", w, ID, vec![], "approval is the user's"));
    let w = World::new();
    w.merged(true, "");
    cases.push(("no result", w, ID, vec![], "closure_result"));
    let w = World::new();
    w.merged(true, "\n## Result\n\n- AC-1: not yet\n");
    cases.push(("open AC", w, ID, vec![], "closure_open"));
    let w = World::new();
    w.merged(true, MET);
    cases.push(("base not a commit", w, ID, vec!["--base", "nope"], "nope"));
    let w = World::new();
    w.merged(true, MET);
    cases.push(("option base", w, ID, vec!["--base=--output=x"], "--base"));
    // The SPEC is on `main` without a trailer; its branch carries unmerged work.
    let w = World::new();
    w.write_spec("login-fix", true, MET);
    w.git(&["add", "-A"]);
    w.commit("docs(spec): login fix");
    w.git(&["switch", "-q", "-c", BRANCH]);
    fs::write(w.project.join("b.txt"), "b\n").unwrap();
    w.git(&["add", "-A"]);
    w.commit(&format!("work\n\nMoai-Spec: {ID}"));
    w.git(&["switch", "-q", "main"]);
    cases.push(("unmerged branch", w, ID, vec![], "not merged"));
    // No branch and no trailer commit on the base.
    let w = World::new();
    w.write_spec("login-fix", true, MET);
    w.git(&["add", "-A"]);
    // The message names the SPEC, but no trailer does; another SPEC's trailer is no evidence.
    w.commit(&format!(
        "docs(spec): add {ID}\n\nMoai-Spec: SPEC-20261117-other-fix"
    ));
    cases.push(("no trailer", w, ID, vec![], "Moai-Spec"));

    for (name, w, id, extra, says) in cases {
        let before = w.everything();
        let m = refused(&w.close(id, &extra));
        assert!(m.starts_with("spec close refused: "), "{name}: {m}");
        assert!(m.contains(says), "{name}: {m}");
        assert_eq!(w.everything(), before, "{name}");
    }
}

#[test]
fn close_with_a_card_marks_it_done() {
    let w = World::new();
    w.merged(true, MET);
    w.review_card(Some(ID), true);
    let r = ok(&w.close(ID, &["--card", "C-1", "--generation", "1"]));
    assert_eq!(r["card"]["action"], "done", "{r}");
    assert_eq!(r["card"]["state"], "done", "{r}");
    assert_eq!(r["status"], "completed");
    ok(&w.run(
        &["spec", "check", ID, "--closure", "--project", w.p()],
        None,
    ));
    let shown = ok(&w.card(&["show", "C-1"]));
    assert_eq!(shown["state"], "done", "{shown}");
    // No card number reaches a shared place: the printed commit names none.
    assert!(!r["next"].to_string().contains("C-1"), "{r}");
}

#[test]
fn a_card_refusal_changes_nothing() {
    // `card done` refuses: the pull request is not merged.
    let open_pr = World::new();
    open_pr.merged(true, MET);
    open_pr.review_card(Some(ID), false);
    // The card names no SPEC.
    let unlinked = World::new();
    unlinked.merged(true, MET);
    unlinked.review_card(None, true);
    // The card names another SPEC.
    let other = World::new();
    other.merged(true, MET);
    let other_id = other.write_spec("other-fix", false, "");
    other.review_card(Some(&other_id), true);
    for (name, w, says) in [
        ("open pr", &open_pr, "merged pull request"),
        ("unlinked", &unlinked, "names no SPEC"),
        ("other", &other, "SPEC-20261117-other-fix"),
    ] {
        let before = w.everything();
        let m = refused(&w.close(ID, &["--card", "C-1", "--generation", "1"]));
        assert!(m.starts_with("spec close refused: "), "{name}: {m}");
        assert!(m.contains(says), "{name}: {m}");
        assert_eq!(w.everything(), before, "{name}");
    }
    // No session id: no holder, refused before anything.
    let before = open_pr.everything();
    let out = open_pr.run(
        &[
            "spec",
            "close",
            ID,
            "--project",
            open_pr.p(),
            "--card",
            "C-1",
            "--generation",
            "1",
        ],
        None,
    );
    let m = refused(&out);
    assert!(m.contains("session"), "{m}");
    assert_eq!(open_pr.everything(), before);
}
