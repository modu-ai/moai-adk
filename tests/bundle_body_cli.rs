//! `moai bundle body` (SPEC-20261119-bundle-ready-check, card C-119): the bundle pull request body
//! from heads, a bundle branch and a run record file, against a temporary ledger and repository.
//! Every command runs against a temporary HOME; the real `~/.moai` and the real repository are not
//! read or written.
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

/// Two heads merged into `bundle`; cards with their READY lines and the verdict the first needs.
struct Bundle {
    w: World,
    cards: [String; 2],
    heads: [String; 2],
    branch_sha: String,
    record: PathBuf,
}

fn record_json(skip_fmt: bool, failed_lib: bool, skip_host: bool) -> Value {
    let mut commands = Vec::new();
    let mut not_run = Vec::new();
    let keys = [
        ("build", "cargo build --locked --offline --bin moai-cli"),
        ("lib_tests", "cargo test --locked --offline --lib"),
        (
            "home_isolation_scan",
            "cargo test --locked --offline --test home_isolation_scan",
        ),
        ("fmt", "cargo fmt --check"),
        (
            "clippy",
            "cargo clippy --locked --offline --all-targets -- -D warnings",
        ),
        ("trailer_check", "moai-cli trailer-check origin/trunk..HEAD"),
        ("spec_check", "moai-cli spec check --all"),
        ("result_coverage", "moai-cli bundle check C-1 C-2"),
    ];
    for (key, cmd) in keys {
        if key == "fmt" && skip_fmt {
            continue;
        }
        let exit = if key == "lib_tests" && failed_lib {
            1
        } else {
            0
        };
        let mut c = serde_json::json!({"check": key, "command": cmd, "exit": exit});
        if key == "lib_tests" {
            c["counts"] = "1453 passed, 0 failed".into();
        }
        commands.push(c);
    }
    not_run.push(serde_json::json!({"check": "product_cli", "reason": "the integration changes no product path"}));
    not_run.push(serde_json::json!({"check": "unittest", "reason": "no Python file changed"}));
    let mut audit = serde_json::json!({
        "plan": "card PLAN note 2447001 of each lane",
        "red": "recorded by each lane in its SPEC Result",
        "mutation": "not run by the leader: merge commits only, no new code to turn a test FAILED",
        "clone": "not_run — no clone path changes",
        "host": "macOS arm64; temp HOME only",
        "receipt": "not_run — no receipt path changes",
    });
    if skip_host {
        audit.as_object_mut().unwrap().remove("host");
    }
    serde_json::json!({"commands": commands, "not_run": not_run, "self_audit": audit})
}

fn bundle_world(merge_second: bool) -> Bundle {
    let w = World::new();
    let h1 = w.head(
        "spec/h1",
        TIP,
        &[("src/h1.rs", "fn h1() {}\n")],
        "SPEC-20261119-demo1",
        &[1, 2],
        true,
    );
    let h2 = w.head(
        "spec/h2",
        TIP,
        &[("docs/h2.md", "h2\n")],
        "SPEC-20261119-demo2",
        &[1, 2],
        true,
    );
    let (c1, c2) = (w.new_card(), w.new_card());
    w.ready_note(&c1, &h1);
    w.pass_note(&c1, &h1);
    w.ready_note(&c2, &h2);
    w.git(&["checkout", "-q", "-B", "bundle", TIP]);
    w.git(&["merge", "-q", "--no-ff", "-m", "Merge h1", &h1]);
    if merge_second {
        w.git(&["merge", "-q", "--no-ff", "-m", "Merge h2", &h2]);
    }
    let branch_sha = w.git(&["rev-parse", "bundle"]);
    let record = w.home.path().join("run-record.json");
    fs::write(&record, record_json(false, false, false).to_string()).unwrap();
    Bundle {
        w,
        cards: [c1, c2],
        heads: [h1, h2],
        branch_sha,
        record,
    }
}

impl Bundle {
    fn body(&self, extra: &[&str]) -> (Value, i32) {
        let mut args = vec![
            "body",
            &self.cards[0],
            &self.cards[1],
            "--branch",
            "bundle",
            "--run",
            self.record.to_str().unwrap(),
        ];
        args.extend_from_slice(extra);
        self.w.bundle(&args)
    }
}

#[test]
fn the_body_carries_the_facts_and_passes_the_self_audit_checker() {
    let b = bundle_world(true);
    let before = b.w.repo_state();
    let (v, code) = b.body(&[]);
    assert_eq!(code, 0, "{v}");
    let r = &v["result"];
    assert_eq!(r["writes"], false);
    assert_eq!(r["pinned"], b.branch_sha.as_str());
    let body = r["body"].as_str().unwrap();
    let first = body.lines().next().unwrap();
    assert!(first.starts_with("audit-tier: 1 "), "{first}");
    assert!(
        first.contains("src/h1.rs") && first.contains("docs/h2.md"),
        "{first}"
    );
    assert!(
        body.contains("Moai-Spec: SPEC-20261119-demo1\n")
            && body.contains("Moai-Spec: SPEC-20261119-demo2\n")
    );
    let row1 = body
        .lines()
        .find(|l| l.contains(&b.heads[0]))
        .expect("a table row for the first head");
    assert!(
        row1.contains("SPEC-20261119-demo1")
            && row1.contains(&b.cards[0])
            && row1.contains("`PASS` card note"),
        "{row1}"
    );
    let row2 = body
        .lines()
        .find(|l| l.contains(&b.heads[1]))
        .expect("a table row for the second head");
    assert!(row2.contains("audit floor 0: no verdict needed"), "{row2}");
    let tip = b.w.tip_sha();
    assert!(
        body.contains(&format!("Integration: `bundle` built from the tip `{tip}` with 2 merge commits → pinned SHA `{}`", b.branch_sha)),
        "{body}"
    );
    assert!(
        body.contains("\n## Self-Audit\n")
            && body.contains("- plan: card PLAN note 2447001 of each lane")
    );
    assert!(
        body.contains("- `cargo test --locked --offline --lib` → exit 0, 1453 passed, 0 failed.")
    );
    assert!(body.contains("- not run: product_cli — the integration changes no product path."));
    assert!(body.contains("SPEC-20261119-demo1 (") && body.contains("2 AC, 2 met, 0 withdrawn"));
    assert!(body
        .trim_end()
        .ends_with("[Claude Code](https://claude.com/claude-code)"));
    // The text passes the product's own Self-Audit checker.
    let file = b.w.home.path().join("body.md");
    fs::write(&file, body).unwrap();
    let checked =
        b.w.home
            .command(&["pr-self-audit-check", "--file", file.to_str().unwrap()])
            .output()
            .unwrap();
    assert!(
        checked.status.success(),
        "{}{}",
        String::from_utf8_lossy(&checked.stdout),
        String::from_utf8_lossy(&checked.stderr)
    );
    assert_eq!(
        b.w.repo_state(),
        before,
        "refs, status and HEAD are as before"
    );
}

#[test]
fn a_body_with_missing_facts_is_refused_naming_all_of_them() {
    // The branch lacks the second head; the record lacks `fmt` and the `host` line, and a command failed.
    let b = bundle_world(false);
    fs::write(&b.record, record_json(true, true, true).to_string()).unwrap();
    let out = b.w.home.path().join("never.md");
    let (v, _) = b.body(&["--out", out.to_str().unwrap()]);
    assert_eq!(v["ok"], false, "{v}");
    assert_eq!(v["error"]["kind"], "invalid_input");
    let m = v["error"]["message"].as_str().unwrap();
    assert!(m.contains("4 missing fact(s)"), "{m}");
    assert!(
        m.contains("the check `fmt` neither ran nor is named not run"),
        "{m}"
    );
    assert!(
        m.contains("the command `cargo test --locked --offline --lib` ended with status 1"),
        "{m}"
    );
    assert!(m.contains("the Self-Audit line `host` has no text"), "{m}");
    assert!(
        m.contains(&format!(
            "{}: the head {} is not in the branch `bundle`",
            b.cards[1],
            &b.heads[1][..10]
        )),
        "{m}"
    );
    assert!(!out.exists(), "a refused body writes no file");
}

#[test]
fn out_writes_a_new_file_only() {
    let b = bundle_world(true);
    let out = b.w.home.path().join("pr-body.md");
    let (v, code) = b.body(&["--out", out.to_str().unwrap()]);
    assert_eq!(code, 0, "{v}");
    assert_eq!(v["result"]["writes"], true);
    let written = fs::read_to_string(&out).unwrap();
    assert_eq!(written, v["result"]["body"].as_str().unwrap());
    // Again to the same path: refused, the file is as it was.
    let (v2, _) = b.body(&["--out", out.to_str().unwrap()]);
    assert_eq!(v2["ok"], false, "{v2}");
    assert!(
        v2["error"]["message"]
            .as_str()
            .unwrap()
            .contains("already exists"),
        "{v2}"
    );
    assert_eq!(fs::read_to_string(&out).unwrap(), written);
    // A folder is an existing path too.
    let dir = b.w.home.path().join("a-folder");
    fs::create_dir(&dir).unwrap();
    let (v3, _) = b.body(&["--out", dir.to_str().unwrap()]);
    assert_eq!(v3["ok"], false, "{v3}");
}

#[test]
fn the_help_and_the_ops_doc_say_what_each_action_does() {
    let w = World::new();
    for action in ["ready", "check", "body"] {
        let out = w
            .home
            .command(&["bundle", action, "--help"])
            .output()
            .unwrap();
        let text = String::from_utf8_lossy(&out.stdout).to_string();
        assert!(
            text.starts_with("Reads only."),
            "bundle {action} --help starts with: {}",
            text.lines().next().unwrap_or("")
        );
    }
    let top = w.home.command(&["bundle", "--help"]).output().unwrap();
    let top = String::from_utf8_lossy(&top.stdout).to_string();
    for word in ["ready", "check", "body", "fetches", "pull request"] {
        assert!(
            top.to_lowercase().contains(&word.to_lowercase()),
            "bundle --help lacks `{word}`: {top}"
        );
    }
    let doc = fs::read_to_string(
        Path::new(env!("CARGO_MANIFEST_DIR"))
            .join("docs/design-review/appendix/operations/bundle.md"),
    )
    .expect("the operations page for bundle");
    for word in [
        "bundle ready",
        "bundle check",
        "bundle body",
        "no_ready_line",
        "head_missing",
        "not_on_origin",
        "already_merged",
        "audit_missing",
        "trailers",
        "result_missing",
        "conflict",
        "unreadable_verdict",
        "commands",
        "not_run",
        "self_audit",
        "lib_tests",
        "result_coverage",
        "Reads only.",
    ] {
        assert!(doc.contains(word), "the operations page lacks `{word}`");
    }
}
