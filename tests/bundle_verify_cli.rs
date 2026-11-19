//! `moai bundle verify` (SPEC-20261119-bundle-verify, card C-119 slice 4): the check list run on
//! the pinned bundle branch under one heavy slot, the run record `bundle body` reads, the heads
//! named for a failure. Every command runs against a temporary HOME, a temporary repository with a
//! bare `origin` and a temporary slot folder; `cargo`, `python3` and `moai-cli` are stub scripts
//! that log their folder and arguments, so nothing real is built or tested and the real `~/.moai`
//! and the real repository are not read or written.
#![cfg(unix)]

mod common;

use common::home::TempHome;
use moai_cli::ledger;
use serde_json::{json, Value};
use std::cell::Cell;
use std::fs;
use std::io::Write;
use std::os::unix::fs::PermissionsExt;
use std::path::{Path, PathBuf};
use std::process::Stdio;

const TIP: &str = "origin/trunk";
const SPEC: &str = "SPEC-20261119-demo";
const SPEC2: &str = "SPEC-20261119-other";
const MANIFEST: &str = "skills/moai/references/source-manifest.json";
const DATE: &str = "20261119";

/// The ten keys of the check list, in order, as `bundle::CHECKS` pins them.
const KEYS: [&str; 10] = [
    "build",
    "lib_tests",
    "home_isolation_scan",
    "product_cli",
    "fmt",
    "clippy",
    "trailer_check",
    "spec_check",
    "result_coverage",
    "unittest",
];

/// A `spec.md` that passes `spec check`, with `## Result` lines for AC-1 and AC-2; its first AC
/// names `tests` as `--test` targets of its `verify:` line when there are any.
fn spec_text(id: &str, tests: &[&str]) -> String {
    let verify = if tests.is_empty() {
        "manual: run it.".to_string()
    } else {
        let cmds: Vec<String> = tests
            .iter()
            .map(|t| format!("`cargo test --locked --offline --test {t}`"))
            .collect();
        format!("verify: {}", cmds.join("; "))
    };
    format!(
        "---\nid: {id}\ntitle: \"demo: a SPEC the bundle tests carry\"\nstatus: draft\nrevision: 1\ncreated: 2026-11-19\n---\n\n\
         # demo: a SPEC the bundle tests carry\n\n## Intent\n\nA fixture.\n\n## Non-goals\n\n- Nothing else.\n\n\
         ## Requirements\n\n- REQ-1: When `demo run` is run, `demo run` shall print one line.\n\
         - REQ-2 (REQ-1): IF the input is empty, THEN `demo run` shall print nothing.\n\n\
         ## Acceptance\n\n- AC-1 (REQ-1): it prints one line.\n  {verify}\n\
         - AC-2 (REQ-2): it prints nothing for empty input.\n  manual: run it with empty input.\n\n\
         ## Scope\n\n- write: `src/demo.rs`\n\n## Plan\n\nOne step.\n\n## Risks\n\n- None.\n\n\
         ## Decisions\n\n| Id | Decision | Basis | Undo |\n| --- | --- | --- | --- |\n| D-1 | None | fixture | none |\n\n\
         ## Result\n\n- AC-1: met — ran it.\n- AC-2: met — ran it.\n"
    )
}

/// What a head adds: files, and a SPEC (its id and the `--test` targets of its `verify:` line).
struct Head<'a> {
    files: Vec<(&'a str, &'a str)>,
    spec: Option<(&'a str, Vec<&'a str>)>,
}

/// A bundle branch built from two heads by `bundle build`.
#[allow(dead_code)]
struct Built {
    cards: [String; 2],
    heads: [String; 2],
    branch: String,
    worktree: PathBuf,
    pinned: String,
}

struct World {
    home: TempHome,
    root: tempfile::TempDir,
    project: String,
    work: PathBuf,
    cards: Cell<u32>,
    runs: Cell<u32>,
    /// `--min-free-gb` given to `verify`: 0 asks for no free disk; a huge value refuses the slot.
    min_free: Cell<u64>,
}

#[allow(dead_code)]
impl World {
    /// A temporary HOME with a ledger, a bare `origin`, a work clone whose tip is `origin/trunk`
    /// (with `shared.txt` of twenty lines and `tests/existing.rs`), the stub programs and a slot
    /// folder whose `limits.json` asks for no free disk and no memory.
    fn new() -> World {
        let home = TempHome::new();
        let project = home.project("demo");
        ledger::register(home.path(), "demo").unwrap();
        let root = tempfile::Builder::new()
            .prefix("moai-test-verify-")
            .tempdir()
            .unwrap();
        let bare = root.path().join("origin.git");
        let work = root.path().join("work");
        fs::create_dir_all(&work).unwrap();
        let w = World {
            home,
            root,
            project: project.to_str().unwrap().to_string(),
            work,
            cards: Cell::new(0),
            runs: Cell::new(0),
            min_free: Cell::new(0),
        };
        let bare_s = bare.to_str().unwrap().to_string();
        w.git_in(
            w.work.parent().unwrap(),
            &["init", "-q", "--bare", "-b", "trunk", &bare_s],
        );
        w.git(&["init", "-q", "-b", "trunk"]);
        w.git(&["remote", "add", "origin", &bare_s]);
        w.git(&["config", "user.name", "Test"]);
        w.git(&["config", "user.email", "test@example.invalid"]);
        w.git(&["config", "merge.moai-manifest.driver", "true"]);
        w.put("README.md", "tip\n");
        w.put(
            ".gitattributes",
            &format!("{MANIFEST} merge=moai-manifest\n"),
        );
        w.put(MANIFEST, "{\n  \"line\": \"base\"\n}\n");
        let shared: String = (1..=20).map(|n| format!("line {n}\n")).collect();
        w.put("shared.txt", &shared);
        w.put("tests/existing.rs", "// the tip's own test\n");
        w.git(&["add", "-A"]);
        w.git(&["commit", "-q", "-m", "chore: the tip"]);
        w.git(&["push", "-q", "origin", "trunk"]);
        w.git(&["fetch", "-q", "origin"]);
        w.stubs();
        w
    }

    fn stub_dir(&self) -> PathBuf {
        self.root.path().join("stubs")
    }

    /// `cargo`, `python3` and `moai-cli`: each logs `<folder>|<name> <arguments>` to `STUB_LOG`,
    /// then answers by the first line of `STUB_RULES` (`pattern|exit|output`, `\n` in the output
    /// is a line break) whose pattern is in the arguments, and otherwise prints its own success
    /// line and exits 0.
    fn stubs(&self) {
        let dir = self.stub_dir();
        fs::create_dir_all(&dir).unwrap();
        for (name, ok) in [
            ("cargo", "test result: ok. 3 passed; 0 failed; 0 ignored"),
            ("python3", "Ran 4 tests in 0.010s\\n\\nOK"),
            ("moai-cli", "ok"),
        ] {
            let script = format!(
                "#!/bin/sh\necho \"$(pwd -P)|$(basename \"$0\") $*\" >> \"$STUB_LOG\"\n\
                 if [ -f \"$STUB_RULES\" ]; then\n  while IFS='|' read -r pat code out; do\n    \
                 [ -z \"$pat\" ] && continue\n    case \"$*\" in *\"$pat\"*) printf '%b\\n' \"$out\"; exit \"$code\";; esac\n  \
                 done < \"$STUB_RULES\"\nfi\nprintf '%b\\n' '{ok}'\nexit 0\n"
            );
            let path = dir.join(name);
            fs::write(&path, script).unwrap();
            fs::set_permissions(&path, fs::Permissions::from_mode(0o755)).unwrap();
        }
        let slots = self.root.path().join("slots");
        fs::create_dir_all(&slots).unwrap();
        fs::write(
            slots.join("limits.json"),
            r#"{"max":2,"min_free_gb":0,"max_swap_pct":100,"min_mem_free_pct":0}"#,
        )
        .unwrap();
        fs::write(self.root.path().join("rules.txt"), "").unwrap();
        fs::write(self.root.path().join("stub.log"), "").unwrap();
    }

    /// Replaces the stub rules (`pattern|exit|output` lines).
    fn rules(&self, lines: &[&str]) {
        let mut text = lines.join("\n");
        text.push('\n');
        fs::write(self.root.path().join("rules.txt"), text).unwrap();
    }

    /// `(folder, command)` of every stub call so far, in order; the folder is as the stub saw it.
    fn log(&self) -> Vec<(String, String)> {
        fs::read_to_string(self.root.path().join("stub.log"))
            .unwrap()
            .lines()
            .filter_map(|l| l.split_once('|'))
            .map(|(a, b)| (a.to_string(), b.to_string()))
            .collect()
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

    /// A branch of one commit on the tip with these files (and a SPEC, with its trailer), pushed.
    fn head(&self, branch: &str, h: &Head) -> String {
        self.git(&["checkout", "-q", "-B", branch, TIP]);
        for (p, c) in &h.files {
            self.put(p, c);
        }
        let mut msg = format!("feat: {branch}");
        if let Some((id, tests)) = &h.spec {
            self.put(&format!(".moai/specs/{id}/spec.md"), &spec_text(id, tests));
            msg = format!("{msg}\n\nMoai-Spec: {id}");
        }
        self.git(&["add", "-A"]);
        self.git(&["commit", "-q", "-m", &msg]);
        let sha = self.git(&["rev-parse", "HEAD"]);
        self.git(&[
            "push",
            "-q",
            "origin",
            &format!("{branch}:refs/heads/{branch}"),
        ]);
        self.git(&["fetch", "-q", "origin"]);
        sha
    }

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

    fn ready_card(&self, head: &str) -> String {
        let n = self.cards.get() + 1;
        self.cards.set(n);
        let scope = format!("area{n}/");
        let c = self.card(
            &["add", "--title", "work", "--scope", &scope, "--ac", "done"],
            None,
        )["name"]
            .as_str()
            .unwrap()
            .to_string();
        self.card(
            &["note", &c, "--kind", "report", "--by", "lane"],
            Some(&format!("LANE-READY-INTEGRATION {c} head={head}")),
        );
        c
    }

    /// `moai bundle <args> --project … --dir … --tip origin/trunk`: the first stdout line as JSON
    /// and the exit status; `env` is added to the clean environment.
    fn bundle_env(&self, args: &[&str], env: &[(&str, String)]) -> (Value, i32) {
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
        let mut cmd = self.home.command(&all);
        for (k, v) in env {
            cmd.env(k, v);
        }
        // Run from the folder that holds the stubs and outputs, so a relative path means that.
        let out = cmd
            .current_dir(self.root.path())
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

    fn bundle(&self, args: &[&str]) -> (Value, i32) {
        self.bundle_env(args, &[])
    }

    /// A new, not yet existing output folder for a verify run.
    fn out(&self) -> PathBuf {
        let n = self.runs.get() + 1;
        self.runs.set(n);
        self.root.path().join("outs").join(format!("run{n}"))
    }

    /// `moai bundle verify <args>` with the stub programs first on `PATH`, the temporary slot
    /// folder and the stub `moai-cli` unless `cli` names another path.
    fn verify_cli(&self, args: &[&str], cli: Option<&Path>) -> (Value, i32) {
        let stub_cli = self.stub_dir().join("moai-cli");
        let cli = cli.unwrap_or(&stub_cli).to_str().unwrap().to_string();
        let slots = self.root.path().join("slots");
        let free = self.min_free.get().to_string();
        let mut all = vec!["verify"];
        all.extend_from_slice(args);
        all.extend_from_slice(&[
            "--moai-cli",
            &cli,
            "--slot-dir",
            slots.to_str().unwrap(),
            "--min-free-gb",
            &free,
            "--max",
            "2",
        ]);
        let path = format!("{}:/usr/bin:/bin", self.stub_dir().display());
        self.bundle_env(
            &all,
            &[
                ("PATH", path),
                (
                    "STUB_LOG",
                    self.root.path().join("stub.log").display().to_string(),
                ),
                (
                    "STUB_RULES",
                    self.root.path().join("rules.txt").display().to_string(),
                ),
            ],
        )
    }

    fn verify(&self, args: &[&str]) -> (Value, i32) {
        self.verify_cli(args, None)
    }

    /// Two heads merged into a new bundle branch by `bundle build`.
    fn built(&self, a: &Head, b: &Head) -> Built {
        let ha = self.head("spec/a", a);
        let hb = self.head("spec/b", b);
        let (ca, cb) = (self.ready_card(&ha), self.ready_card(&hb));
        let (v, code) = self.bundle(&["build", &ca, &cb, "--date", DATE]);
        assert_eq!(code, 0, "{v}");
        assert_eq!(v["result"]["ok"], true, "{v}");
        Built {
            cards: [ca, cb],
            heads: [ha, hb],
            branch: v["result"]["branch"].as_str().unwrap().to_string(),
            worktree: PathBuf::from(v["result"]["worktree"].as_str().unwrap()),
            pinned: v["result"]["pinned"].as_str().unwrap().to_string(),
        }
    }

    /// The refs of the clone and of `origin`, the status of the clone, and the cards with their
    /// notes: what a verify run must leave as it was.
    fn state(&self) -> Value {
        let bare = self.work.parent().unwrap().join("origin.git");
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
        json!({
            "refs": self.git(&["for-each-ref"]),
            "status": self.git(&["status", "--porcelain"]),
            "head": self.git(&["rev-parse", "HEAD"]),
            "origin": self.git_in(&bare, &["for-each-ref"]),
            "worktrees": self.git(&["worktree", "list", "--porcelain"]),
            "ledger": {"list": list, "cards": shown},
        })
    }
}

fn commands_of(v: &Value) -> Vec<(String, String, i64)> {
    v["result"]["commands"]
        .as_array()
        .unwrap()
        .iter()
        .map(|c| {
            (
                c["check"].as_str().unwrap_or("").to_string(),
                c["command"].as_str().unwrap().to_string(),
                c["exit"].as_i64().unwrap(),
            )
        })
        .collect()
}

fn not_run_of(v: &Value) -> Vec<(String, String)> {
    v["result"]["not_run"]
        .as_array()
        .unwrap()
        .iter()
        .map(|c| {
            (
                c["check"].as_str().unwrap().to_string(),
                c["reason"].as_str().unwrap().to_string(),
            )
        })
        .collect()
}

#[test]
fn every_command_of_the_list_runs_in_the_branch_worktree() {
    let w = World::new();
    let b = w.built(
        &Head {
            files: vec![("notes/a.md", "a\n")],
            spec: Some((SPEC, vec![])),
        },
        &Head {
            files: vec![("notes/b.md", "b\n")],
            spec: Some((SPEC, vec![])),
        },
    );
    let audit = w.root.path().join("self-audit.json");
    fs::write(
        &audit,
        json!({
            "plan": "card PLAN note 2447001 of each lane",
            "red": "recorded by each lane in its SPEC Result",
            "mutation": "not run by the leader: merge commits only, no new code to turn a test FAILED",
            "clone": "not_run - no clone path changes",
            "host": "macOS arm64; temp HOME only",
            "receipt": "not_run - no receipt path changes",
        })
        .to_string(),
    )
    .unwrap();
    let out = w.out();
    let (v, code) = w.verify(&[
        &b.cards[0],
        &b.cards[1],
        "--branch",
        &b.branch,
        "--out",
        out.to_str().unwrap(),
        "--self-audit",
        audit.to_str().unwrap(),
    ]);
    assert_eq!(code, 0, "{v}");
    let r = &v["result"];
    assert_eq!(r["action"], "verify");
    assert_eq!(r["ok"], true, "{v}");
    assert_eq!(r["pinned"], b.pinned.as_str());
    assert_eq!(r["writes"], true);

    // One slot hold for the whole list, the commands in the table's order, in the worktree.
    let cli = w.stub_dir().join("moai-cli").display().to_string();
    let wanted = [
        (
            "build",
            "cargo build --locked --offline --bin moai-cli".to_string(),
        ),
        (
            "lib_tests",
            "cargo test --locked --offline --lib".to_string(),
        ),
        (
            "home_isolation_scan",
            "cargo test --locked --offline --test home_isolation_scan".to_string(),
        ),
        (
            "product_cli",
            "cargo test --locked --offline --test product_cli".to_string(),
        ),
        ("fmt", "cargo fmt --check".to_string()),
        (
            "clippy",
            "cargo clippy --locked --offline --all-targets -- -D warnings".to_string(),
        ),
        ("trailer_check", format!("{cli} trailer-check {TIP}..HEAD")),
        ("spec_check", format!("{cli} spec check --all")),
        (
            "result_coverage",
            format!("{cli} spec check {SPEC} --closure"),
        ),
        (
            "unittest",
            "python3 -m unittest discover -s tests -p '*_test.py'".to_string(),
        ),
    ];
    let got = commands_of(&v);
    let got_pairs: Vec<(String, String)> = got.iter().map(|c| (c.0.clone(), c.1.clone())).collect();
    let want_pairs: Vec<(String, String)> = wanted
        .iter()
        .map(|(k, c)| (k.to_string(), c.clone()))
        .collect();
    assert_eq!(got_pairs, want_pairs);
    assert!(got.iter().all(|c| c.2 == 0));
    let wt = fs::canonicalize(&b.worktree).unwrap().display().to_string();
    let log = w.log();
    assert_eq!(log.len(), 10, "{log:?}");
    for (folder, _) in &log {
        assert_eq!(folder, &wt, "every command runs in the branch's worktree");
    }
    assert!(
        log[9]
            .1
            .ends_with("-m unittest discover -s tests -p *_test.py"),
        "{:?}",
        log[9]
    );

    // The folder: the record and one log per command; `body` accepts the record.
    let run = out.join("run.json");
    let record: Value = serde_json::from_str(&fs::read_to_string(&run).unwrap()).unwrap();
    assert_eq!(record["commands"].as_array().unwrap().len(), 10);
    assert_eq!(record["not_run"], json!([]));
    assert_eq!(
        record["self_audit"]["plan"],
        "card PLAN note 2447001 of each lane"
    );
    for c in record["commands"].as_array().unwrap() {
        assert_eq!(c["exit"], 0);
        assert!(c["started"].as_i64().unwrap() <= c["ended"].as_i64().unwrap());
        let log_file = out.join(c["log"].as_str().unwrap());
        assert!(log_file.is_file(), "{log_file:?}");
    }
    assert_eq!(record["commands"][1]["counts"], "3 passed");
    assert_eq!(record["commands"][9]["counts"], "Ran 4 tests");
    let mut log_names: Vec<&str> = record["commands"]
        .as_array()
        .unwrap()
        .iter()
        .map(|c| c["log"].as_str().unwrap())
        .collect();
    log_names.sort_unstable();
    log_names.dedup();
    assert_eq!(log_names.len(), 10, "one log per command");
    assert!(out.join("run.sh").is_file());
    let (body, code) = w.bundle(&[
        "body",
        &b.cards[0],
        &b.cards[1],
        "--branch",
        &b.branch,
        "--run",
        run.to_str().unwrap(),
    ]);
    assert_eq!(code, 0, "{body}");
    assert_eq!(body["result"]["pinned"], b.pinned.as_str());
}

#[test]
fn refusals_run_nothing_and_write_nothing() {
    let w = World::new();
    let b = w.built(
        &Head {
            files: vec![("notes/a.md", "a\n")],
            spec: Some((SPEC, vec![])),
        },
        &Head {
            files: vec![("notes/b.md", "b\n")],
            spec: Some((SPEC, vec![])),
        },
    );
    let stray = w.head(
        "spec/stray",
        &Head {
            files: vec![("notes/stray.md", "s\n")],
            spec: Some((SPEC, vec![])),
        },
    );
    let c_stray = w.ready_card(&stray);
    w.git(&["branch", "m-flow/no-worktree", &b.pinned]);
    let before = w.state();
    let heads = [b.cards[0].as_str(), b.cards[1].as_str()];
    let refused = |w: &World, args: &[&str], must: &[&str], out: &Path| {
        let (v, code) = w.verify(args);
        assert_eq!(code, 1, "{v}");
        assert_eq!(v["ok"], false);
        let message = v["error"]["message"].as_str().unwrap();
        for m in must {
            assert!(message.contains(m), "`{m}` missing in: {message}");
        }
        assert!(w.log().is_empty(), "no stub ran: {:?}", w.log());
        assert!(!out.exists(), "no output folder was made");
    };

    // A branch that has no worktree.
    let out = w.out();
    refused(
        &w,
        &[
            heads[0],
            heads[1],
            "--branch",
            "m-flow/no-worktree",
            "--out",
            out.to_str().unwrap(),
        ],
        &["no worktree"],
        &out,
    );

    // A worktree whose HEAD is not the branch's SHA: detached one commit back.
    w.git_in(&b.worktree, &["checkout", "-q", "--detach", "HEAD~1"]);
    let out = w.out();
    refused(
        &w,
        &[
            heads[0],
            heads[1],
            "--branch",
            &b.branch,
            "--out",
            out.to_str().unwrap(),
        ],
        &["HEAD"],
        &out,
    );
    w.git_in(&b.worktree, &["checkout", "-q", &b.branch]);

    // A modified tracked file, and a head in neither the tip nor the branch: both named at once.
    fs::write(b.worktree.join("README.md"), "changed\n").unwrap();
    let out = w.out();
    refused(
        &w,
        &[
            heads[0],
            heads[1],
            &c_stray,
            "--branch",
            &b.branch,
            "--out",
            out.to_str().unwrap(),
        ],
        &["uncommitted", &c_stray],
        &out,
    );
    w.git_in(&b.worktree, &["checkout", "-q", "--", "README.md"]);

    // An output folder that exists; a self-audit file that is not an object, and one that is missing.
    let taken = w.out();
    fs::create_dir_all(&taken).unwrap();
    let (v, code) = w.verify(&[
        heads[0],
        heads[1],
        "--branch",
        &b.branch,
        "--out",
        taken.to_str().unwrap(),
    ]);
    assert_eq!(code, 1, "{v}");
    assert!(
        v["error"]["message"].as_str().unwrap().contains("exists"),
        "{v}"
    );
    assert!(w.log().is_empty());
    assert!(
        !taken.join("run.json").exists() && !taken.join("run.sh").exists(),
        "left as it was"
    );
    let array = w.root.path().join("array.json");
    fs::write(&array, "[1, 2]").unwrap();
    let out = w.out();
    refused(
        &w,
        &[
            heads[0],
            heads[1],
            "--branch",
            &b.branch,
            "--out",
            out.to_str().unwrap(),
            "--self-audit",
            array.to_str().unwrap(),
        ],
        &["--self-audit"],
        &out,
    );
    let missing = w.root.path().join("missing.json");
    let out = w.out();
    refused(
        &w,
        &[
            heads[0],
            heads[1],
            "--branch",
            &b.branch,
            "--out",
            out.to_str().unwrap(),
            "--self-audit",
            missing.to_str().unwrap(),
        ],
        &["--self-audit"],
        &out,
    );
    assert_eq!(
        w.state(),
        before,
        "no ref, status, worktree or ledger row changed"
    );
}

#[test]
fn a_failed_command_does_not_stop_the_list_and_a_missing_binary_is_not_run() {
    let w = World::new();
    let b = w.built(
        &Head {
            files: vec![("notes/a.md", "a\n")],
            spec: Some((SPEC, vec![])),
        },
        &Head {
            files: vec![("notes/b.md", "b\n")],
            spec: Some((SPEC, vec![])),
        },
    );
    let heads = [b.cards[0].as_str(), b.cards[1].as_str()];

    // `fmt` fails: everything after it still runs. `cargo fmt --check` names the file by its
    // absolute path (`Diff in <worktree>/notes/a.md:1:`), which only the first head adds.
    let wt = fs::canonicalize(&b.worktree).unwrap().display().to_string();
    let fmt_rule = format!("fmt --check|1|Diff in {wt}/notes/a.md:1:");
    w.rules(&[fmt_rule.as_str()]);
    let out = w.out();
    let (v, code) = w.verify(&[
        heads[0],
        heads[1],
        "--branch",
        &b.branch,
        "--out",
        out.to_str().unwrap(),
    ]);
    assert_eq!(code, 0, "{v}");
    assert_eq!(v["result"]["ok"], false);
    let got = commands_of(&v);
    assert_eq!(got.len(), 10);
    assert_eq!(
        got.iter()
            .filter(|c| c.2 != 0)
            .map(|c| c.0.as_str())
            .collect::<Vec<_>>(),
        vec!["fmt"]
    );
    assert_eq!(w.log().len(), 10, "the list went on after the failure");
    assert_eq!(v["result"]["failed"].as_array().unwrap().len(), 1);
    assert_eq!(v["result"]["failed"][0]["check"], "fmt");
    assert_eq!(
        v["result"]["failed"][0]["heads"],
        json!([b.cards[0].clone()])
    );
    assert_eq!(not_run_of(&v), vec![]);

    // `build` fails: the three checks that need the built binary are not run, with the reason,
    // and are not among the failed commands.
    fs::write(w.root.path().join("stub.log"), "").unwrap();
    w.rules(&["build --locked|101|error: boom"]);
    let out = w.out();
    let (v, code) = w.verify(&[
        heads[0],
        heads[1],
        "--branch",
        &b.branch,
        "--out",
        out.to_str().unwrap(),
    ]);
    assert_eq!(code, 0, "{v}");
    let ran: Vec<String> = commands_of(&v).into_iter().map(|c| c.0).collect();
    assert_eq!(
        ran,
        vec![
            "build",
            "lib_tests",
            "home_isolation_scan",
            "product_cli",
            "fmt",
            "clippy",
            "unittest"
        ]
    );
    let not_run = not_run_of(&v);
    assert_eq!(
        not_run.iter().map(|n| n.0.as_str()).collect::<Vec<_>>(),
        vec!["trailer_check", "spec_check", "result_coverage"]
    );
    assert!(not_run.iter().all(|n| n.1.contains("build")), "{not_run:?}");
    let failed: Vec<&str> = v["result"]["failed"]
        .as_array()
        .unwrap()
        .iter()
        .map(|f| f["check"].as_str().unwrap())
        .collect();
    assert_eq!(failed, vec!["build"]);
    let record: Value =
        serde_json::from_str(&fs::read_to_string(out.join("run.json")).unwrap()).unwrap();
    assert_eq!(record["not_run"].as_array().unwrap().len(), 3);

    // A `moai-cli` path with nothing at it: the same three are not run.
    fs::write(w.root.path().join("stub.log"), "").unwrap();
    w.rules(&[]);
    let nothing = w.root.path().join("no-such-moai-cli");
    let out = w.out();
    let (v, code) = w.verify_cli(
        &[
            heads[0],
            heads[1],
            "--branch",
            &b.branch,
            "--out",
            out.to_str().unwrap(),
        ],
        Some(&nothing),
    );
    assert_eq!(code, 0, "{v}");
    assert_eq!(
        not_run_of(&v)
            .iter()
            .map(|n| n.0.clone())
            .collect::<Vec<_>>(),
        vec!["trailer_check", "spec_check", "result_coverage"]
    );
    assert_eq!(
        v["result"]["ok"], true,
        "a command that cannot run is not a failure: {v}"
    );
}

#[test]
fn a_bundle_that_names_no_spec_has_no_result_coverage_to_run() {
    let w = World::new();
    let b = w.built(
        &Head {
            files: vec![("notes/a.md", "a\n")],
            spec: None,
        },
        &Head {
            files: vec![("notes/b.md", "b\n")],
            spec: None,
        },
    );
    let out = w.out();
    let (v, code) = w.verify(&[
        &b.cards[0],
        &b.cards[1],
        "--branch",
        &b.branch,
        "--out",
        out.to_str().unwrap(),
    ]);
    assert_eq!(code, 0, "{v}");
    assert_eq!(
        not_run_of(&v)
            .iter()
            .map(|n| n.0.clone())
            .collect::<Vec<_>>(),
        vec!["result_coverage"]
    );
    assert!(not_run_of(&v)[0].1.contains("SPEC"), "{v}");
    assert_eq!(commands_of(&v).len(), 9);
}

#[test]
fn a_failure_names_the_head_that_touches_its_file() {
    let w = World::new();
    let shared_a: String = std::iter::once("line 1 changed by a\n".to_string())
        .chain((2..=20).map(|n| format!("line {n}\n")))
        .collect();
    let shared_b: String = (1..=19)
        .map(|n| format!("line {n}\n"))
        .chain(std::iter::once("line 20 changed by b\n".to_string()))
        .collect();
    let b = w.built(
        &Head {
            files: vec![
                ("tests/alpha.rs", "// alpha\n"),
                ("shared.txt", shared_a.as_str()),
            ],
            spec: Some((SPEC, vec![])),
        },
        &Head {
            files: vec![("notes/b.md", "b\n"), ("shared.txt", shared_b.as_str())],
            spec: Some((SPEC, vec![])),
        },
    );
    w.rules(&[
        "--test alpha|101|test alpha ... FAILED",
        "clippy|1|error: unused\\n --> shared.txt:3:1",
        "unittest|1|FAILED (failures=1)",
    ]);
    let out = w.out();
    let (v, code) = w.verify(&[
        &b.cards[0],
        &b.cards[1],
        "--branch",
        &b.branch,
        "--out",
        out.to_str().unwrap(),
    ]);
    assert_eq!(code, 0, "{v}");
    assert_eq!(v["result"]["ok"], false);
    let failed = v["result"]["failed"].as_array().unwrap();
    let heads_of = |check: &str| -> Vec<String> {
        failed
            .iter()
            .find(|f| f["check"] == check)
            .unwrap_or_else(|| panic!("`{check}` is not among the failed: {failed:?}"))["heads"]
            .as_array()
            .unwrap()
            .iter()
            .map(|h| h.as_str().unwrap().to_string())
            .collect()
    };
    // The target of a failing `--test alpha` is tests/alpha.rs, which only the first head adds.
    assert_eq!(heads_of("test:alpha"), vec![b.cards[0].clone()]);
    // A path in the output that both heads touch names both.
    assert_eq!(
        heads_of("clippy"),
        vec![b.cards[0].clone(), b.cards[1].clone()]
    );
    // An output that names no changed file is unattributed.
    assert_eq!(heads_of("unittest"), vec!["unattributed".to_string()]);
    assert_eq!(failed.len(), 3);
}

#[test]
fn the_list_is_the_documented_data() {
    let w = World::new();
    // Head one adds tests/alpha.rs; head two changes tests/existing.rs and its SPEC names
    // alpha again, existing again and beta (which has no file).
    let b = w.built(
        &Head {
            files: vec![("tests/alpha.rs", "// alpha\n")],
            spec: Some((SPEC, vec![])),
        },
        &Head {
            files: vec![("tests/existing.rs", "// changed\n")],
            spec: Some((SPEC2, vec!["alpha", "existing", "beta"])),
        },
    );
    let out = w.out();
    let (v, code) = w.verify(&[
        &b.cards[0],
        &b.cards[1],
        "--branch",
        &b.branch,
        "--out",
        out.to_str().unwrap(),
    ]);
    assert_eq!(code, 0, "{v}");
    let cli = w.stub_dir().join("moai-cli").display().to_string();
    let texts: Vec<String> = commands_of(&v).into_iter().map(|c| c.1).collect();
    let want = vec![
        "cargo build --locked --offline --bin moai-cli".to_string(),
        "cargo test --locked --offline --lib".to_string(),
        "cargo test --locked --offline --test home_isolation_scan".to_string(),
        "cargo test --locked --offline --test product_cli".to_string(),
        "cargo test --locked --offline --test alpha".to_string(),
        "cargo test --locked --offline --test existing".to_string(),
        "cargo fmt --check".to_string(),
        "cargo clippy --locked --offline --all-targets -- -D warnings".to_string(),
        format!("{cli} trailer-check {TIP}..HEAD"),
        format!("{cli} spec check --all"),
        format!("{cli} spec check {SPEC} --closure"),
        format!("{cli} spec check {SPEC2} --closure"),
        "python3 -m unittest discover -s tests -p '*_test.py'".to_string(),
    ];
    assert_eq!(texts, want);
    // A key for each of the ten checks and none for the extras; `beta` has no file.
    let keys: Vec<String> = commands_of(&v).into_iter().map(|c| c.0).collect();
    let mut seen: Vec<&str> = Vec::new();
    for k in &keys {
        if !k.is_empty() && !seen.contains(&k.as_str()) {
            seen.push(k);
        }
    }
    assert_eq!(seen, KEYS.to_vec());
    assert_eq!(keys[4], "");
    assert_eq!(keys[5], "");
    let not_run = not_run_of(&v);
    assert_eq!(not_run.len(), 1, "{not_run:?}");
    assert_eq!(not_run[0].0, "test:beta");
    assert!(not_run[0].1.contains("tests/beta.rs"), "{not_run:?}");
}

#[test]
fn verify_writes_no_ref_and_no_ledger_row() {
    let w = World::new();
    let b = w.built(
        &Head {
            files: vec![("notes/a.md", "a\n")],
            spec: Some((SPEC, vec![])),
        },
        &Head {
            files: vec![("notes/b.md", "b\n")],
            spec: Some((SPEC, vec![])),
        },
    );
    let before = w.state();
    let out = w.out();
    let (v, code) = w.verify(&[
        &b.cards[0],
        &b.cards[1],
        "--branch",
        &b.branch,
        "--out",
        out.to_str().unwrap(),
    ]);
    assert_eq!(code, 0, "{v}");
    assert_eq!(w.state(), before);
    assert_eq!(
        w.git_in(
            &b.worktree,
            &["status", "--porcelain", "--untracked-files=no"]
        ),
        ""
    );
}

#[test]
fn the_help_and_the_ops_doc_name_what_verify_does() {
    let w = World::new();
    let out = w
        .home
        .command(&["bundle", "verify", "--help"])
        .output()
        .unwrap();
    let text = String::from_utf8_lossy(&out.stdout).to_string();
    assert!(
        text.starts_with("Runs commands."),
        "bundle verify --help starts with: {}",
        text.lines().next().unwrap_or("")
    );
    for word in [
        "heavy slot",
        "run.json",
        "target/",
        "not_run",
        "--self-audit",
        "--out",
    ] {
        assert!(
            text.contains(word),
            "the verify help lacks `{word}`: {text}"
        );
    }
    let doc = fs::read_to_string(
        Path::new(env!("CARGO_MANIFEST_DIR"))
            .join("docs/design-review/appendix/operations/bundle.md"),
    )
    .expect("the operations page for bundle");
    for word in [
        "bundle verify",
        "Runs commands.",
        "run.json",
        "not_run",
        "unattributed",
        "heavy-slot-acquire",
        "cargo build --locked --offline --bin moai-cli",
        "python3 -m unittest discover -s tests -p '*_test.py'",
        "result_coverage",
        "--self-audit",
    ] {
        assert!(doc.contains(word), "the operations page lacks `{word}`");
    }
}

#[test]
fn a_relative_path_is_the_callers_own_not_the_worktrees() {
    let w = World::new();
    let b = w.built(
        &Head {
            files: vec![("notes/a.md", "a\n")],
            spec: Some((SPEC, vec![])),
        },
        &Head {
            files: vec![("notes/b.md", "b\n")],
            spec: Some((SPEC, vec![])),
        },
    );
    // The command runs in the folder holding `stubs/` and `outs/`; both paths are relative to it.
    let (v, code) = w.verify_cli(
        &[
            &b.cards[0],
            &b.cards[1],
            "--branch",
            &b.branch,
            "--out",
            "outs/relative",
        ],
        Some(Path::new("stubs/moai-cli")),
    );
    assert_eq!(code, 0, "{v}");
    assert_eq!(v["result"]["ok"], true, "{v}");
    assert!(w.root.path().join("outs/relative/run.json").is_file());
    assert_eq!(not_run_of(&v), vec![]);
    // The caller's folder is `/private/var/...` where the temporary folder says `/var/...`: compare
    // the canonical spelling.
    let stubs = fs::canonicalize(w.stub_dir())
        .unwrap()
        .display()
        .to_string();
    assert!(
        commands_of(&v)
            .iter()
            .any(|c| c.0 == "trailer_check" && c.1.starts_with(&stubs)),
        "the binary is the absolute path of the stub: {v}"
    );
}

#[test]
fn the_default_folder_is_under_the_worktree_target_and_untracked_files_are_allowed() {
    let w = World::new();
    let b = w.built(
        &Head {
            files: vec![("notes/a.md", "a\n")],
            spec: Some((SPEC, vec![])),
        },
        &Head {
            files: vec![("notes/b.md", "b\n")],
            spec: Some((SPEC, vec![])),
        },
    );
    // An untracked file is not a change to a tracked file (REQ-2).
    fs::write(b.worktree.join("untracked.txt"), "x\n").unwrap();
    let (v, code) = w.verify(&[&b.cards[0], &b.cards[1], "--branch", &b.branch]);
    assert_eq!(code, 0, "{v}");
    assert_eq!(v["result"]["ok"], true, "{v}");
    let want = fs::canonicalize(&b.worktree)
        .unwrap()
        .join("target")
        .join("bundle-verify")
        .join(&b.pinned[..10]);
    let got = fs::canonicalize(v["result"]["out"].as_str().unwrap()).unwrap();
    assert_eq!(got, want);
    assert!(want.join("run.json").is_file());
}

#[test]
fn a_missing_branch_and_a_busy_slot_are_refused_with_nothing_left() {
    let w = World::new();
    let b = w.built(
        &Head {
            files: vec![("notes/a.md", "a\n")],
            spec: Some((SPEC, vec![])),
        },
        &Head {
            files: vec![("notes/b.md", "b\n")],
            spec: Some((SPEC, vec![])),
        },
    );
    let out = w.out();
    let (v, code) = w.verify(&[
        &b.cards[0],
        &b.cards[1],
        "--branch",
        "m-flow/nope",
        "--out",
        out.to_str().unwrap(),
    ]);
    assert_eq!(code, 1, "{v}");
    assert!(
        v["error"]["message"]
            .as_str()
            .unwrap()
            .contains("is not here"),
        "{v}"
    );
    assert!(!out.exists());

    // A machine that has less free disk than asked: the slot is refused (exit 75), the script
    // never starts, and the folder this run made is removed again.
    w.min_free.set(999_999_999);
    let (v, code) = w.verify(&[
        &b.cards[0],
        &b.cards[1],
        "--branch",
        &b.branch,
        "--out",
        out.to_str().unwrap(),
    ]);
    assert_eq!(code, 1, "{v}");
    assert!(
        v["error"]["message"]
            .as_str()
            .unwrap()
            .contains("heavy slot"),
        "{v}"
    );
    assert!(!out.exists(), "the folder of a refused run is gone");
    assert!(w.log().is_empty(), "no command ran: {:?}", w.log());
}

#[test]
fn a_command_that_never_finished_is_not_run_not_a_failure() {
    let w = World::new();
    let b = w.built(
        &Head {
            files: vec![("notes/a.md", "a\n")],
            spec: Some((SPEC, vec![])),
        },
        &Head {
            files: vec![("notes/b.md", "b\n")],
            spec: Some((SPEC, vec![])),
        },
    );
    // The last command kills the script that runs it: no exit status is ever written for it.
    let kill = w.stub_dir().join("python3");
    fs::write(&kill, "#!/bin/sh\nkill -9 $PPID\nsleep 5\n").unwrap();
    fs::set_permissions(&kill, fs::Permissions::from_mode(0o755)).unwrap();
    let out = w.out();
    let (v, code) = w.verify(&[
        &b.cards[0],
        &b.cards[1],
        "--branch",
        &b.branch,
        "--out",
        out.to_str().unwrap(),
    ]);
    assert_eq!(code, 0, "{v}");
    assert_eq!(
        v["result"]["ok"], true,
        "no exit status is not a failure: {v}"
    );
    assert_eq!(commands_of(&v).len(), 9);
    let not_run = not_run_of(&v);
    assert_eq!(not_run.len(), 1, "{not_run:?}");
    assert_eq!(not_run[0].0, "unittest");
    assert!(not_run[0].1.contains("did not run"), "{not_run:?}");
}
