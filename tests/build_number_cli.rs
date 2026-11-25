//! SPEC-20261125-install-build-number: `build-number next` records and prints the next install
//! build number, `build-number show` reads it. Every run uses a temporary HOME and a temporary
//! project in a clean environment; the real `~/.moai` is never read or written.
mod common;

use common::home::TempHome;
use serde_json::{json, Value};
use std::fs;
use std::path::{Path, PathBuf};
use std::process::Output;

const LOG: &str = "build-number.log";

struct World {
    home: TempHome,
    project: PathBuf,
    /// `<home>/.moai/projects/demo`, the folder the record lives in.
    state: PathBuf,
}

fn world() -> World {
    let home = TempHome::new();
    let project = home.project("demo");
    fs::write(project.join("HEADVER"), "4\n").unwrap();
    let state = moai_cli::ledger::location::project_dir(home.path(), "demo").unwrap();
    fs::create_dir_all(&state).unwrap();
    World {
        home,
        project,
        state,
    }
}

impl World {
    fn run(&self, action: &str) -> Output {
        self.home
            .command(&[
                "build-number",
                action,
                "--project",
                self.project.to_str().unwrap(),
            ])
            .output()
            .expect("run moai")
    }

    fn log(&self) -> String {
        fs::read_to_string(self.state.join(LOG)).unwrap_or_default()
    }

    /// The installed version folders the floor is read from.
    fn install(&self, name: &str) {
        fs::create_dir_all(self.home.path().join(".moai/versions").join(name)).unwrap();
    }
}

fn envelope(out: &Output) -> Value {
    serde_json::from_slice(&out.stdout).unwrap_or_else(|e| {
        panic!(
            "stdout is not JSON ({e}): {} / {}",
            String::from_utf8_lossy(&out.stdout),
            String::from_utf8_lossy(&out.stderr)
        )
    })
}

fn result(out: &Output) -> Value {
    let v = envelope(out);
    assert_eq!(out.status.code(), Some(0), "{v}");
    assert_eq!(v["ok"], json!(true), "{v}");
    v["result"].clone()
}

fn refused(out: &Output) -> String {
    let v = envelope(out);
    assert_eq!(out.status.code(), Some(1), "{v}");
    assert_eq!(v["ok"], json!(false), "{v}");
    v["error"]["message"].as_str().unwrap().to_string()
}

fn build(out: &Output) -> u64 {
    result(out)["build"].as_u64().unwrap()
}

#[test]
fn three_runs_hand_out_one_two_three_and_record_them() {
    let w = world();
    for want in 1..=3u64 {
        let r = result(&w.run("next"));
        assert_eq!(r["build"], json!(want), "{r}");
        assert_eq!(r["env"], json!(format!("MOAI_LOCAL_BUILD={want}")), "{r}");
        let version = r["version"].as_str().unwrap();
        assert!(
            version.starts_with("4.") && version.ends_with(&format!(".{want}-dev")),
            "{version}"
        );
        assert_eq!(r["recorded"], json!(true), "{r}");
    }
    let log = w.log();
    let numbers: Vec<&str> = log.lines().map(|l| l.split('\t').next().unwrap()).collect();
    assert_eq!(numbers, ["1", "2", "3"], "{log}");
    for line in log.lines() {
        let stamp = line.split('\t').nth(1).unwrap();
        assert!(moai_cli::state_root::utc_stamp(stamp), "{line}");
    }
}

#[test]
fn the_first_number_starts_above_the_highest_installed_folder_of_its_head() {
    let w = world();
    w.install("4.2648.36-dev");
    w.install("4.2640.7");
    w.install("5.2648.99-dev");
    w.install("4.2648.40-rc");
    w.install("notes");
    assert_eq!(build(&w.run("next")), 37);
    assert_eq!(build(&w.run("next")), 38);
    // An installed folder that appears later raises the floor, and a number never goes back.
    w.install("4.2649.50");
    assert_eq!(build(&w.run("next")), 51);
}

#[test]
fn a_record_that_is_not_whole_numbers_is_refused_and_left_as_it_is() {
    for (bad, why) in [
        ("garbage\n", "not a record line"),
        ("1\t20261125T000000Z\t-\n3", "a torn last line"),
        (
            "1\t20261125T000000Z\t-",
            "a whole last line without its line break",
        ),
        (
            "2\t20261125T000000Z\t-\n1\t20261125T000001Z\t-\n",
            "a number that goes down",
        ),
        ("0\t20261125T000000Z\t-\n", "zero"),
        ("01\t20261125T000000Z\t-\n", "a leading zero"),
        ("1\tyesterday\t-\n", "a time that is no UTC stamp"),
    ] {
        let w = world();
        fs::write(w.state.join(LOG), bad).unwrap();
        let message = refused(&w.run("next"));
        assert!(message.contains(LOG), "{why}: {message}");
        assert_eq!(w.log(), bad, "{why}: the record changed");
        let message = refused(&w.run("show"));
        assert!(message.contains(LOG), "{why}: {message}");
        assert_eq!(w.log(), bad, "{why}: show changed the record");
    }
}

#[test]
fn a_project_without_a_head_file_or_without_its_state_folder_is_refused() {
    let w = world();
    fs::remove_file(w.project.join("HEADVER")).unwrap();
    let message = refused(&w.run("next"));
    assert!(message.contains("HEADVER"), "{message}");
    assert_eq!(w.log(), "");

    let w = world();
    fs::remove_dir_all(&w.state).unwrap();
    let message = refused(&w.run("next"));
    assert!(message.contains("register"), "{message}");
    assert!(!w.state.exists(), "the command created the state folder");
}

#[cfg(unix)]
#[test]
fn a_state_folder_that_is_a_link_is_refused() {
    let w = world();
    let elsewhere = w.home.path().join("elsewhere");
    fs::create_dir_all(&elsewhere).unwrap();
    fs::remove_dir_all(&w.state).unwrap();
    std::os::unix::fs::symlink(&elsewhere, &w.state).unwrap();
    let message = refused(&w.run("next"));
    assert!(message.contains("link"), "{message}");
    assert!(
        !elsewhere.join(LOG).exists(),
        "the record was written through the link"
    );
}

#[test]
fn show_writes_nothing() {
    let w = world();
    let before = w.home.snapshot();
    let r = result(&w.run("show"));
    assert_eq!(r["last"], Value::Null, "{r}");
    assert_eq!(r["next"], json!(1), "{r}");
    assert_eq!(w.home.snapshot(), before, "show changed the HOME");

    assert_eq!(build(&w.run("next")), 1);
    assert_eq!(build(&w.run("next")), 2);
    let before = w.home.snapshot();
    let r = result(&w.run("show"));
    assert_eq!(r["last"]["build"], json!(2), "{r}");
    assert_eq!(r["next"], json!(3), "{r}");
    assert_eq!(w.home.snapshot(), before, "show changed the HOME");
}

#[test]
fn concurrent_runs_get_different_numbers() {
    let w = world();
    let children: Vec<_> = (0..8)
        .map(|_| {
            w.home
                .command(&[
                    "build-number",
                    "next",
                    "--project",
                    w.project.to_str().unwrap(),
                ])
                .stdout(std::process::Stdio::piped())
                .stderr(std::process::Stdio::piped())
                .spawn()
                .expect("spawn moai")
        })
        .collect();
    let mut numbers: Vec<u64> = children
        .into_iter()
        .map(|c| build(&c.wait_with_output().expect("wait")))
        .collect();
    numbers.sort_unstable();
    assert_eq!(numbers, (1..=8).collect::<Vec<u64>>());
    let recorded: Vec<String> = w
        .log()
        .lines()
        .map(|l| l.split('\t').next().unwrap().to_string())
        .collect();
    assert_eq!(recorded, ["1", "2", "3", "4", "5", "6", "7", "8"]);
}

#[test]
fn the_record_names_the_session_that_asked() {
    let w = world();
    let id = "428e8b95-c84b-44e2-8496-8c71d0864a2b";
    let out = w
        .home
        .command(&[
            "build-number",
            "next",
            "--project",
            w.project.to_str().unwrap(),
        ])
        .env("CLAUDE_CODE_SESSION_ID", id)
        .output()
        .expect("run moai");
    assert_eq!(build(&out), 1);
    assert!(
        w.log().trim_end().ends_with(&format!("\t{id}")),
        "{}",
        w.log()
    );
    // Without a session id the record keeps a dash, and the number is still handed out.
    assert_eq!(build(&w.run("next")), 2);
    assert!(w.log().trim_end().ends_with("\t-"), "{}", w.log());
}

#[test]
fn the_record_file_is_private_to_its_owner() {
    #[cfg(unix)]
    {
        use std::os::unix::fs::PermissionsExt;
        let w = world();
        assert_eq!(build(&w.run("next")), 1);
        let mode = fs::metadata(Path::new(&w.state).join(LOG))
            .unwrap()
            .permissions()
            .mode()
            & 0o777;
        assert_eq!(mode, 0o600);
    }
}
