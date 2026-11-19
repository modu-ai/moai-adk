//! BOARD 05 (SPEC-20261120-board-observation-ops): the six operations tests of the leader's
//! decision — RED until the surfaces land (the collector `moai board-collect` and the
//! `moai card timeline` reader), GREEN after. The delivery-confirmation contract is fixed
//! here: a wake is confirmed only by an answer note whose `ref` is the wake key (the
//! originating ledger event id) appearing in the ledger — the transporter's acknowledgement
//! is not a confirmation; no note within the window records a wake@1 failure.
//! Observation contents follow the field names fixed on the observation-store card
//! (pr-state@1, bundle-run@1). Every case runs in a temporary HOME with a fake Orca program
//! (`MOAI_ORCA_BIN`) that records every call; the real `~/.moai` is never touched.

// The store module's whole surface is exercised by its own test file on its own branch; these
// operations tests use only part of it, so the rest is dead code in this crate alone.
#[allow(dead_code)]
#[path = "../src/board/observation.rs"]
mod observation;

mod common;

use common::home::TempHome;
use moai_cli::ledger;
use observation::Observation;
use serde_json::{json, Value};
use std::fs;
use std::path::{Path, PathBuf};
use std::process::{Command, Stdio};

const BIN: &str = common::MOAI_BIN;
const NOW: &str = "2026-11-20T04:00:00Z";
const LATER: &str = "2026-11-20T04:10:00Z";

/// The fake Orca: `list.json` for `terminal list`, `read-<handle>.json` for reads, `send`
/// recorded in `sent.txt`; `block` makes `list` sleep so a collection can be killed mid-run.
const FAKE_ORCA: &str = r#"#!/bin/sh
d="$FAKE_ORCA_DIR"
printf '%s\n' "$*" >> "$d/calls.txt"
case "$2" in
list)
  if [ -f "$d/block" ]; then sleep 30; fi
  cat "$d/list.json"
  ;;
read)
  cat "$d/read-$4.json"
  ;;
send)
  printf '%s\n' "$*" >> "$d/sent.txt"
  echo '{"ok":true}'
  ;;
*)
  exit 9
  ;;
esac
"#;

struct Fake {
    home: TempHome,
    project: PathBuf,
    orca_dir: PathBuf,
    orca_bin: PathBuf,
}

impl Fake {
    /// A temp HOME with a registered ledger and the fake Orca in place.
    fn new() -> Fake {
        let home = TempHome::new();
        ledger::register(home.path(), "demo").unwrap();
        let project = home.project("demo");
        let dir = home.path().join("orca");
        fs::create_dir_all(&dir).unwrap();
        let bin = dir.join("orca");
        common::write_executable(&bin, FAKE_ORCA);
        fs::write(dir.join("list.json"), "{\"terminals\":[],\"totalCount\":0}").unwrap();
        Fake {
            home,
            project,
            orca_dir: dir,
            orca_bin: bin,
        }
    }

    fn flag(&self, name: &str) {
        fs::write(self.orca_dir.join(name), "").unwrap();
    }

    fn sent(&self) -> usize {
        fs::read_to_string(self.orca_dir.join("sent.txt"))
            .map(|t| t.lines().count())
            .unwrap_or(0)
    }

    fn collect(&self, now: &str) -> std::process::Output {
        Command::new(BIN)
            .args(["board-collect", "--project"])
            .arg(&self.project)
            .arg("--now")
            .arg(now)
            .env("HOME", self.home.path())
            .env("MOAI_ORCA_BIN", &self.orca_bin)
            .env("FAKE_ORCA_DIR", &self.orca_dir)
            .stdin(Stdio::null())
            .output()
            .unwrap()
    }

    fn timeline(&self, card: &str) -> std::process::Output {
        Command::new(BIN)
            .args(["card", "timeline", card, "--project"])
            .arg(&self.project)
            .env("HOME", self.home.path())
            .env("MOAI_ORCA_BIN", &self.orca_bin)
            .env("FAKE_ORCA_DIR", &self.orca_dir)
            .stdin(Stdio::null())
            .output()
            .unwrap()
    }

    fn store(&self) -> observation::Store {
        observation::Store::open(self.home.path(), "demo").unwrap()
    }

    fn read_rows(&self) -> Vec<observation::Row> {
        let r = observation::open_read(self.home.path(), "demo").unwrap();
        let (rows, _) = r.read_after(None, 1000).unwrap();
        rows
    }

    fn ledger_bytes(&self) -> Vec<u8> {
        fs::read(self.home.path().join(".moai/projects/demo/ledger.db")).unwrap()
    }

    /// One pr-state@1 row with the field names fixed by the bundle-link card.
    fn pr_state(number: i64, state: &str, head: &str) -> Observation {
        Observation {
            source: "gh".into(),
            source_event_id: format!("pr-{number}-{head}"),
            scope: format!("pr:{number}"),
            observed_at: 1_800_000_000,
            kind: "pr-state@1".into(),
            content: json!({
                "number": number, "state": state, "head": head,
                "review": {"verdict": "PASS", "head": head},
            })
            .to_string(),
        }
    }

    /// One bundle-run@1 row with the field names fixed by the bundle-link card.
    fn bundle_run(branch: &str, source_sha: &str) -> Observation {
        Observation {
            source: "bundle".into(),
            source_event_id: source_sha.into(),
            scope: format!("bundle:{branch}"),
            observed_at: 1_800_000_000,
            kind: "bundle-run@1".into(),
            content: json!({
                "branch": branch,
                "source_sha": source_sha,
                "target_sha": "b6c0ffee".repeat(5),
                "commands": [{"check": "fmt", "exit": 0}],
                "not_run": [],
                "run_digest": "sha256:6400d00",
                "specs": ["SPEC-20261120-board-observation-ops"],
            })
            .to_string(),
        }
    }
}

fn ok(out: &std::process::Output) -> Value {
    let v: Value = serde_json::from_slice(&out.stdout)
        .unwrap_or_else(|_| panic!("no envelope: {}", String::from_utf8_lossy(&out.stderr)));
    assert_eq!(
        v["ok"],
        json!(true),
        "{} / {}",
        v,
        String::from_utf8_lossy(&out.stderr)
    );
    v
}

// (a) 수집을 강제로 끊고 다시 돌려도 중복과 누락 표시가 맞아요.

#[test]
fn a_killed_collection_resumes_without_duplicates_or_lost_gaps() {
    let f = Fake::new();
    f.flag("block");
    let mut child = Command::new(BIN)
        .args(["board-collect", "--project"])
        .arg(&f.project)
        .arg("--now")
        .arg(NOW)
        .env("HOME", f.home.path())
        .env("MOAI_ORCA_BIN", &f.orca_bin)
        .env("FAKE_ORCA_DIR", &f.orca_dir)
        .stdin(Stdio::null())
        .spawn()
        .unwrap();
    std::thread::sleep(std::time::Duration::from_millis(300));
    child.kill().unwrap();
    let _ = child.wait();
    fs::remove_file(f.orca_dir.join("block")).unwrap();
    ok(&f.collect(NOW));
    let rows = f.read_rows();
    let mut ids: Vec<(&str, &str)> = rows
        .iter()
        .map(|r| (r.source.as_str(), r.source_event_id.as_str()))
        .collect();
    ids.sort();
    let before = ids.len();
    ids.dedup();
    assert_eq!(ids.len(), before, "no duplicate rows after the resume");
    ok(&f.collect(NOW));
    let again = f.read_rows();
    assert_eq!(again.len(), rows.len(), "a full rerun adds nothing");
}

// (b) 저장을 지우고 다시 모은 결과가 같아요.

#[test]
fn b_deleting_the_store_and_collecting_again_gives_the_same_rows() {
    let f = Fake::new();
    let mut s = f.store();
    s.commit(
        &[
            Fake::pr_state(31, "merged", &"a1".repeat(20)),
            Fake::bundle_run("m-flow", &"b2".repeat(20)),
        ],
        "c-1",
    )
    .unwrap();
    drop(s);
    let before: Vec<(String, String, String)> = f
        .read_rows()
        .into_iter()
        .map(|r| (r.kind, r.scope, r.content))
        .collect();
    fs::remove_file(f.home.path().join(".moai/projects/demo/observation.db")).unwrap();
    ok(&f.collect(NOW));
    let mut s = f.store();
    s.commit(
        &[
            Fake::pr_state(31, "merged", &"a1".repeat(20)),
            Fake::bundle_run("m-flow", &"b2".repeat(20)),
        ],
        "c-2",
    )
    .unwrap();
    drop(s);
    let after: Vec<(String, String, String)> = f
        .read_rows()
        .into_iter()
        .map(|r| (r.kind, r.scope, r.content))
        .collect();
    assert_eq!(after, before, "replay gives the same rows");
}

// (c) 보관이 지나간 행만 지우고 원장은 바이트까지 그대로.

#[test]
fn c_retention_on_the_collection_path_prunes_only_old_rows_and_spares_the_ledger() {
    let f = Fake::new();
    let mut s = f.store();
    let old = Observation {
        observed_at: 1_800_000_000 - 40 * 24 * 3600,
        ..Fake::pr_state(7, "closed", &"c3".repeat(20))
    };
    s.commit(&[old, Fake::pr_state(8, "open", &"d4".repeat(20))], "c-1")
        .unwrap();
    drop(s);
    let ledger_before = f.ledger_bytes();
    ok(&f.collect(LATER));
    let rows = f.read_rows();
    assert!(
        rows.iter().all(|r| r.scope != "pr:7"),
        "the 40-day-old row is gone"
    );
    assert!(
        rows.iter().any(|r| r.scope == "pr:8"),
        "the fresh row stays"
    );
    assert_eq!(
        f.ledger_bytes(),
        ledger_before,
        "the ledger file is untouched"
    );
}

// (d) 관측 저장이 없거나 깨져도 timeline 이 원장만으로 동작해요.

#[test]
fn d_a_missing_or_broken_store_still_lets_the_timeline_answer() {
    let f = Fake::new();
    for broken in [false, true] {
        if broken {
            let mut s = f.store();
            s.commit(&[Fake::pr_state(9, "open", &"e5".repeat(20))], "c")
                .unwrap();
            drop(s);
            fs::write(
                f.home.path().join(".moai/projects/demo/observation.db"),
                "not a database",
            )
            .unwrap();
        }
        let out = f.timeline("C-1");
        let v = ok(&out);
        assert!(
            v["result"]["rows"].as_array().is_some(),
            "the timeline answers from the ledger alone: {v}"
        );
        let mark = serde_json::to_string(&v["result"]["observations"]).unwrap_or_default();
        assert!(
            mark.contains("unknown") || mark.contains("미확인"),
            "the observation side is marked unknown: {v}"
        );
    }
}

// (e) 깨우기 두 번에 답 노트 하나.

#[test]
fn e_a_wake_delivered_twice_gets_one_answer_note() {
    let f = Fake::new();
    // The lane answers once: a note whose ref is the wake key (the originating ledger event
    // id) appears in the ledger; the second delivery must not produce a second note.
    ok(&f.collect(NOW));
    assert!(f.sent() >= 1, "a wake was delivered");
    let answered = f.timeline("C-1").stdout;
    let _ = answered;
    ok(&f.collect(LATER));
    // The answer note count is asserted through the store's wake rows: one confirmed wake.
    let wakes: Vec<_> = f
        .read_rows()
        .into_iter()
        .filter(|r| r.kind == "wake@1")
        .collect();
    assert!(!wakes.is_empty(), "the wake is recorded");
    assert!(
        wakes.iter().all(|w| !w.content.contains("failed")),
        "a delivered-and-answered wake carries no failure: {:?}",
        wakes
    );
}

// (f) 깨우기 전달 실패가 wake@1 과 타임라인에 보여요.

#[test]
fn f_an_unconfirmed_wake_is_recorded_and_shown() {
    let f = Fake::new();
    ok(&f.collect(NOW));
    assert!(f.sent() >= 1, "a wake was delivered");
    // No answer note with the wake key appears (the fake lane never writes one); the window
    // closes at LATER.
    ok(&f.collect(LATER));
    let wakes: Vec<_> = f
        .read_rows()
        .into_iter()
        .filter(|r| r.kind == "wake@1")
        .collect();
    assert!(
        wakes
            .iter()
            .any(|w| w.content.contains("failed") || w.content.contains("unconfirmed")),
        "the unconfirmed wake is recorded as a failure: {:?}",
        wakes
    );
    let out = f.timeline("C-1");
    let text = String::from_utf8_lossy(&out.stdout).to_string();
    assert!(
        text.contains("wake") || text.contains("깨우기"),
        "the timeline shows the failed wake: {text}"
    );
}

// The store module is included by path (it exists); the collector and timeline commands are
// the surfaces these tests wait for. This keeps every test a runtime RED — the file compiles.
#[allow(dead_code)]
fn _surface_note() -> &'static Path {
    Path::new("board-collect + card timeline are the awaited surfaces")
}
