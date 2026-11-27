//! Small CLI gaps found in a real project run: `moai session-register --list` reads the whole
//! session registry of this clone (the other reads need a name), and a refused
//! `card finish --evidence-note` names the `ref` it wants. Every case runs in a temporary git
//! project with a temporary HOME; nothing of this machine is read, and the list writes nothing.
#![cfg(unix)]

use crate::common;

use common::CleanEnv;
use serde_json::Value;
use std::collections::BTreeMap;
use std::fs;
use std::path::{Path, PathBuf};
use std::process::{Command, Output};

struct Fixture {
    _root: tempfile::TempDir,
    repo: PathBuf,
    home: PathBuf,
}

impl Fixture {
    fn new() -> Self {
        let root = tempfile::tempdir().unwrap();
        let base = fs::canonicalize(root.path()).unwrap();
        let repo = base.join("project");
        let home = base.join("home");
        fs::create_dir_all(&repo).unwrap();
        fs::create_dir_all(&home).unwrap();
        assert!(common::git::command()
            .args(["init", "-q"])
            .current_dir(&repo)
            .status()
            .unwrap()
            .success());
        Fixture {
            _root: root,
            repo,
            home,
        }
    }

    fn sessions(&self) -> PathBuf {
        moai_cli::session::agent_seq_dir(&self.repo).unwrap()
    }

    fn run(&self, args: &[&str]) -> (Output, Value) {
        let out = Command::new(common::MOAI_BIN)
            .args(args)
            .current_dir(&self.repo)
            .clean_env()
            .env("HOME", &self.home)
            .env("PATH", "/usr/bin:/bin")
            .env("GIT_CONFIG_GLOBAL", "/dev/null")
            .output()
            .unwrap();
        let v = serde_json::from_slice(&out.stdout).unwrap_or(Value::Null);
        (out, v)
    }

    fn answer(&self, args: &[&str]) -> Value {
        let (out, v) = self.run(args);
        assert!(out.status.success(), "{args:?}: {}", text(&out));
        assert_eq!(v["ok"], true, "{args:?}: {v}");
        v["result"].clone()
    }

    fn register(&self, name: &str, now: &str, lease: &str) {
        let project = self.repo.to_str().unwrap();
        self.answer(&[
            "session-register",
            project,
            "--name",
            name,
            "--mode",
            "agent",
            "--lease-secs",
            lease,
            "--now",
            now,
        ]);
    }
}

fn text(out: &Output) -> String {
    format!(
        "{}{}",
        String::from_utf8_lossy(&out.stdout),
        String::from_utf8_lossy(&out.stderr)
    )
}

fn snapshot(dir: &Path) -> BTreeMap<String, Vec<u8>> {
    let mut seen = BTreeMap::new();
    if let Ok(entries) = fs::read_dir(dir) {
        for e in entries.flatten() {
            seen.insert(
                e.file_name().to_string_lossy().into_owned(),
                fs::read(e.path()).unwrap_or_default(),
            );
        }
    }
    seen
}

const T0: &str = "2026-11-27T00:00:00Z";
const T1: &str = "2026-11-27T00:30:00Z";

#[test]
fn list_of_a_clone_without_a_registry_is_empty_and_makes_no_folder() {
    let fx = Fixture::new();
    let answer = fx.answer(&["session-register", "--list"]);
    assert_eq!(answer["action"], "session.list", "{answer}");
    assert_eq!(answer["entries"], serde_json::json!([]), "{answer}");
    assert_eq!(answer["writes"], false, "{answer}");
    assert!(!fx.sessions().exists(), "a read makes no registry folder");
}

#[test]
fn list_shows_every_entry_by_name_with_its_state_and_never_the_instance() {
    let fx = Fixture::new();
    fx.register("agent-b", T0, "3600");
    fx.register("agent-a", T0, "60");
    let before = snapshot(&fx.sessions());
    let answer = fx.answer(&["session-register", "--list", "--now", T1]);
    assert_eq!(snapshot(&fx.sessions()), before, "the list writes nothing");
    let names: Vec<&str> = answer["entries"]
        .as_array()
        .unwrap()
        .iter()
        .map(|e| e["name"].as_str().unwrap())
        .collect();
    assert_eq!(names, ["agent-a", "agent-b"], "{answer}");
    // agent-a's lease (60 s from T0) ended before T1; agent-b's (an hour) still runs.
    assert_eq!(answer["entries"][0]["status"], "ended", "{answer}");
    assert_eq!(answer["entries"][1]["status"], "held", "{answer}");
    assert_eq!(answer["entries"][1]["mode"], "agent", "{answer}");
    assert_eq!(answer["count"], 2, "{answer}");
    // The instance is the key that releases a name: `--show` prints it, the list does not.
    assert!(
        !answer.to_string().contains("instance"),
        "the list must not print an instance: {answer}"
    );
}

#[test]
fn list_cannot_be_combined_with_a_name() {
    let fx = Fixture::new();
    let (out, _) = fx.run(&["session-register", "--list", "--name", "agent-1"]);
    assert!(!out.status.success(), "{}", text(&out));
    assert!(text(&out).contains("--list"), "{}", text(&out));
}

// ---- `card finish --evidence-note` ----------------------------------------------------------

mod finish_note {
    use super::common;
    use common::home::TempHome;
    use moai_cli::ledger;
    use serde_json::{json, Value};
    use std::fs;
    use std::io::Write;
    use std::path::PathBuf;
    use std::process::{Output, Stdio};

    const SESSION_VAR: &str = "CLAUDE_CODE_SESSION_ID";
    const S_A: &str = "aaaaaaaa-0000-4000-8000-00000000000a";

    struct World {
        home: TempHome,
        p: String,
    }

    impl World {
        fn new() -> Self {
            let home = TempHome::new();
            let project: PathBuf = home.project("demo");
            ledger::register(home.path(), "demo").unwrap();
            let git = |args: &[&str]| {
                let mut all = vec![
                    "-c",
                    "user.name=fixture",
                    "-c",
                    "user.email=fixture@example.invalid",
                ];
                all.extend_from_slice(args);
                common::git::git(&project, &all)
            };
            git(&["init", "-q", "-b", "main"]);
            fs::write(project.join("a.txt"), "a\n").unwrap();
            git(&["add", "a.txt"]);
            git(&["commit", "-q", "-m", "init"]);
            let cfg = moai_cli::state_root::path(&project, "project.yaml");
            fs::create_dir_all(cfg.parent().unwrap()).unwrap();
            fs::write(
                &cfg,
                "{\"schema_version\": 1, \"project_id\": \"demo\", \"purpose\": \"fixture\", \"tools\": [\"rust\"], \"development_methods\": [\"sdd\"], \"hosts\": [\"claude\"]}",
            )
            .unwrap();
            World {
                p: project.to_str().unwrap().to_string(),
                home,
            }
        }

        fn card(&self, stdin: Option<&str>, args: &[&str]) -> Output {
            let mut all = vec!["card"];
            all.extend_from_slice(args);
            all.extend_from_slice(&["--project", &self.p]);
            let mut cmd = self.home.command(&all);
            cmd.env("LC_ALL", "C")
                .env(SESSION_VAR, S_A)
                .stdin(Stdio::piped())
                .stdout(Stdio::piped())
                .stderr(Stdio::piped());
            let mut child = cmd.spawn().expect("run moai");
            let mut pipe = child.stdin.take().unwrap();
            let _ = pipe.write_all(stdin.unwrap_or("").as_bytes());
            drop(pipe);
            child.wait_with_output().unwrap()
        }

        fn ok(&self, stdin: Option<&str>, args: &[&str]) -> Value {
            let out = self.card(stdin, args);
            assert_eq!(
                out.status.code(),
                Some(0),
                "{args:?}: {}",
                super::text(&out)
            );
            let v: Value = serde_json::from_slice(&out.stdout).unwrap();
            assert_eq!(v["ok"], json!(true), "{v}");
            v["result"].clone()
        }

        /// A research card in review with one report note; returns the card, the note's number
        /// and its `ref`.
        fn research_card_with_report(&self) -> (String, i64, String) {
            let added = self.ok(
                None,
                &[
                    "add", "--title", "fixture", "--kind", "research", "--scope", "docs/", "--ac",
                    "one",
                ],
            );
            let card = added["name"].as_str().unwrap().to_string();
            self.ok(None, &["ready", &card]);
            self.ok(None, &["take", &card]);
            let note = self.ok(Some("the finding\n"), &["note", &card, "--kind", "report"]);
            self.ok(None, &["review", &card, "--generation", "1"]);
            (
                card,
                note["note_id"].as_i64().unwrap(),
                note["ref"].as_str().unwrap().to_string(),
            )
        }
    }

    #[test]
    fn a_note_number_closes_a_research_card_as_its_ref_does() {
        let w = World::new();
        let (card, number, reference) = w.research_card_with_report();
        assert!(reference.starts_with("note:"), "{reference}");
        let done = w.ok(
            None,
            &[
                "finish",
                &card,
                "--evidence-note",
                &number.to_string(),
                "--met",
                "1=the finding",
                "--by",
                "leader",
            ],
        );
        assert_eq!(done["closed"], json!(true), "{done}");
        assert_eq!(done["via"], "note", "{done}");
        assert_eq!(
            done["note"],
            json!(reference),
            "the answer names the ref: {done}"
        );
    }

    #[test]
    fn a_number_that_is_no_note_of_this_card_is_still_refused() {
        let w = World::new();
        let (card, number, _) = w.research_card_with_report();
        for wrong in [(number + 1000).to_string(), "0".to_string()] {
            let out = w.card(
                None,
                &[
                    "finish",
                    &card,
                    "--evidence-note",
                    &wrong,
                    "--met",
                    "1=x",
                    "--by",
                    "leader",
                ],
            );
            assert_eq!(out.status.code(), Some(1), "{}", super::text(&out));
            assert!(
                super::text(&out).contains("no note with ref"),
                "{}",
                super::text(&out)
            );
        }
    }
}
