//! C-513 (S7 of the update redesign): one confirmed `moai update` leaves one run receipt, and the
//! end screen is drawn from it. Every run here uses a temporary HOME with the real `skills/moai`
//! package and agents beside a copy of the binary; the real HOME is never touched. The failure
//! behind the skills comes from the debug-build seam `MOAI_TEST_FAULT=update_part_agents`.

#![cfg(unix)]

mod common;
use common::CleanEnv;

use moai_cli::ui::Painter;
use moai_cli::ui_text::Lang;
use moai_cli::update_tux::{screen, Asked, View};
use serde_json::{json, Value};
use sha2::{Digest, Sha256};
use std::{
    fs,
    path::{Path, PathBuf},
    process::{Command, Output, Stdio},
};
use tempfile::TempDir;

const EXE: &str = crate::common::MOAI_BIN;
const AGENTS_FAULT: (&str, &str) = ("MOAI_TEST_FAULT", "update_part_agents");
const HTML_REPORT: &str = moai_cli::update::extra_skill::SKILLS[0];

fn repo() -> &'static Path {
    Path::new(env!("CARGO_MANIFEST_DIR"))
}

fn copy_tree(from: &Path, to: &Path) {
    fs::create_dir_all(to).unwrap();
    for entry in fs::read_dir(from).unwrap() {
        let p = entry.unwrap().path();
        let dest = to.join(p.file_name().unwrap());
        if p.is_dir() {
            copy_tree(&p, &dest);
        } else {
            fs::copy(&p, &dest).unwrap();
        }
    }
}

/// The folder of the installed versions under the HOME state folder.
fn versions(home: &Path) -> PathBuf {
    home.join(moai_cli::HOME_STATE_DIR).join("versions")
}

/// A temporary HOME with one installed version folder holding a copy of the binary, the real
/// package beside it and the agent sources next to `skills/`.
fn setup() -> (TempDir, PathBuf, PathBuf) {
    let tmp = TempDir::new().unwrap();
    let root = fs::canonicalize(tmp.path()).unwrap();
    let home = root.join("home");
    let folder = versions(&home).join("a");
    fs::create_dir_all(folder.join("bin")).unwrap();
    let exe = folder.join("bin/moai-cli");
    common::copy_executable(Path::new(EXE), &exe);
    copy_tree(&repo().join("skills/moai"), &folder.join("skills/moai"));
    copy_tree(&repo().join("agents"), &folder.join("agents"));
    (tmp, home, exe)
}

fn command(exe: &Path, home: &Path, args: &[&str]) -> Command {
    let mut c = Command::new(exe);
    c.args(args)
        .clean_env()
        .env("PATH", "/usr/bin:/bin")
        .env("HOME", home)
        .env("LC_ALL", "C")
        .current_dir(home)
        .stdin(Stdio::null());
    c
}

fn run(exe: &Path, home: &Path, args: &[&str]) -> Output {
    command(exe, home, args).output().unwrap()
}

fn run_with(exe: &Path, home: &Path, args: &[&str], var: (&str, &str)) -> Output {
    command(exe, home, args).env(var.0, var.1).output().unwrap()
}

fn result(out: &Output) -> Value {
    let v: Value = serde_json::from_slice(&out.stdout).unwrap();
    assert_eq!(v["ok"], json!(true), "{v}");
    v["result"].clone()
}

fn digest_of(preview: &Value) -> String {
    preview["plan_digest"].as_str().unwrap().to_string()
}

fn agents_dir(home: &Path) -> PathBuf {
    home.join(".claude/agents/moai")
}

fn sha256(bytes: &[u8]) -> String {
    format!("{:x}", Sha256::digest(bytes))
}

/// Every file under `dir` with its bytes, except the version folder of the fixture.
fn state(home: &Path) -> Vec<(PathBuf, Vec<u8>)> {
    fn walk(dir: &Path, skip: &Path, out: &mut Vec<(PathBuf, Vec<u8>)>) {
        let Ok(entries) = fs::read_dir(dir) else {
            return;
        };
        for e in entries {
            let p = e.unwrap().path();
            if p.starts_with(skip) {
                continue;
            }
            if p.is_dir() {
                walk(&p, skip, out);
            } else {
                out.push((p.clone(), fs::read(&p).unwrap()));
            }
        }
    }
    let mut out = Vec::new();
    walk(home, &versions(home), &mut out);
    out.sort();
    out
}

fn view(lang: Lang) -> View {
    View {
        painter: Painter {
            color: false,
            unicode: true,
            lang,
        },
        width: 100,
        home: "/home/u".into(),
        project: None,
    }
}

fn parts<'a>(run: &'a Value, name: &str) -> Vec<&'a Value> {
    run["parts"]
        .as_array()
        .unwrap()
        .iter()
        .filter(|p| p["part"] == json!(name))
        .collect()
}

/// AC-1: a failure injected behind the skills ends the run `partial`, names the part that is
/// left, and leaves the skill files written; the next run applies what is left and ends complete.
#[test]
fn a_failure_behind_the_skills_ends_partial_names_what_is_left_and_a_rerun_finishes() {
    let (_t, home, exe) = setup();
    let top = digest_of(&result(&run(&exe, &home, &["update"])));
    let r = result(&run_with(
        &exe,
        &home,
        &["update", "--confirm", &top],
        AGENTS_FAULT,
    ));
    assert_eq!(r["mode"], json!("applied"), "{r}");
    assert_eq!(r["agents_result"]["state"], json!("error"), "{r}");
    let first = &r["run"];
    assert_eq!(first["status"], json!("partial"), "{first}");
    assert_eq!(first["remaining"], json!(["agents"]), "{first}");
    assert_eq!(first["next"], json!("moai update"), "{first}");
    assert_eq!(first["wrote"], json!(true), "{first}");
    let agents = parts(first, "agents");
    assert_eq!(agents.len(), 1, "{first}");
    assert!(
        agents[0]["reason"]
            .as_str()
            .unwrap()
            .contains("update_part_agents"),
        "{first}"
    );
    // The skills landed and the part that failed wrote nothing.
    assert!(home.join(".claude/skills/moai/SKILL.md").is_file());
    assert!(!agents_dir(&home).exists(), "{:?}", state(&home));
    // The record is on disk and says the same.
    let path = PathBuf::from(first["receipt"].as_str().unwrap());
    assert!(path.starts_with(&home), "{}", path.display());
    assert_eq!(path.file_name().unwrap(), "update-run.json");
    let on_disk: Value = serde_json::from_slice(&fs::read(&path).unwrap()).unwrap();
    assert_eq!(on_disk["status"], json!("partial"));
    assert_eq!(on_disk["remaining"], json!(["agents"]));

    // The rerun finds the skills unchanged, applies the agents and ends complete.
    let p = result(&run(&exe, &home, &["update"]));
    assert_eq!(p["nothing_to_apply"], json!(false), "{p}");
    let r = result(&run(&exe, &home, &["update", "--confirm", &digest_of(&p)]));
    assert_eq!(r["agents_result"]["state"], json!("applied"), "{r}");
    let second = &r["run"];
    assert_eq!(second["status"], json!("complete"), "{second}");
    assert_eq!(second["remaining"], json!([]), "{second}");
    assert!(agents_dir(&home).is_dir());
    let on_disk: Value = serde_json::from_slice(&fs::read(&path).unwrap()).unwrap();
    assert_eq!(on_disk["status"], json!("complete"));
    // Nothing is left for a third run.
    let p = result(&run(&exe, &home, &["update"]));
    assert_eq!(p["nothing_to_apply"], json!(true), "{p}");
}

/// AC-2: every skill file the run wrote is read back from disk and its digest is compared with
/// the one the result named; the moai skill and the folder beside it are both covered, and the
/// digests the record shows are the ones of the bytes on disk.
#[test]
fn every_written_skill_file_is_read_back_and_its_digest_compared() {
    let (_t, home, exe) = setup();
    let pkg = versions(&home).join("a/skills");
    copy_tree(
        &repo().join("skills").join(HTML_REPORT),
        &pkg.join(HTML_REPORT),
    );
    let top = digest_of(&result(&run(&exe, &home, &["update"])));
    let r = result(&run(&exe, &home, &["update", "--confirm", &top]));
    let rec = &r["run"];
    assert_eq!(rec["status"], json!("complete"), "{rec}");
    let files = rec["files"].as_array().unwrap();
    let published: usize = r["result"]["hosts"]
        .as_array()
        .unwrap()
        .iter()
        .chain(r["extra_skills_result"]["parts"].as_array().unwrap())
        .flat_map(|h| h["files"].as_array().unwrap())
        .filter(|f| f["published"] == json!(true))
        .count();
    assert!(
        published > 10,
        "a fresh install writes the whole skill: {published}"
    );
    assert_eq!(files.len(), published, "one check per written file");
    assert_eq!(rec["files_checked"], json!(published));
    assert_eq!(rec["files_not_matching"], json!(0));
    let skills = home.join(".claude/skills");
    let mut seen = std::collections::BTreeSet::new();
    for f in files {
        assert_eq!(f["check"], json!("match"), "{f}");
        let root = match f["part"].as_str().unwrap() {
            "skills" => skills.join("moai"),
            "extra_skills" => skills.join(HTML_REPORT),
            other => panic!("unexpected part {other}"),
        };
        let on_disk = fs::read(root.join(f["path"].as_str().unwrap())).unwrap();
        let bare = |k: &str| {
            f[k].as_str()
                .unwrap()
                .trim_start_matches("sha256:")
                .to_string()
        };
        assert_eq!(bare("observed_sha256"), sha256(&on_disk), "{f}");
        assert_eq!(bare("expected_sha256"), sha256(&on_disk), "{f}");
        seen.insert(f["part"].as_str().unwrap().to_string());
    }
    assert_eq!(
        seen.into_iter().collect::<Vec<_>>(),
        ["extra_skills", "skills"]
    );
    // The parts that wrote files this record does not read back are named, not hidden.
    let names: Vec<&str> = rec["not_compared"]
        .as_array()
        .unwrap()
        .iter()
        .map(|p| p["part"].as_str().unwrap())
        .collect();
    assert!(names.contains(&"agents"), "{rec}");
}

/// A run that writes nothing and leaves nothing open writes no receipt and no other file.
#[test]
fn a_run_with_nothing_to_write_leaves_no_receipt() {
    let (_t, home, exe) = setup();
    let top = digest_of(&result(&run(&exe, &home, &["update"])));
    result(&run(&exe, &home, &["update", "--confirm", &top]));
    let again = result(&run(&exe, &home, &["update"]));
    assert_eq!(again["nothing_to_apply"], json!(true), "{again}");
    let before = state(&home);
    let r = result(&run(
        &exe,
        &home,
        &["update", "--confirm", &digest_of(&again)],
    ));
    assert_eq!(r["run"]["status"], json!("complete"), "{}", r["run"]);
    assert_eq!(r["run"]["wrote"], json!(false), "{}", r["run"]);
    assert!(r["run"].get("receipt").is_none(), "{}", r["run"]);
    assert_eq!(state(&home), before, "nothing was written");
}

/// AC-2: the end screen reads the run receipt. A partial run names what is left and the command
/// that continues it; a clean run says the files were read back; and the head follows the
/// receipt, not the parts: a receipt whose file does not read back as written turns a run whose
/// parts all look clean into a partly applied one.
#[test]
fn the_end_screen_is_drawn_from_the_run_receipt() {
    let (_t, home, exe) = setup();
    let top = digest_of(&result(&run(&exe, &home, &["update"])));
    let partial = result(&run_with(
        &exe,
        &home,
        &["update", "--confirm", &top],
        AGENTS_FAULT,
    ));
    let en = screen(&partial, Asked::Yes, &view(Lang::En));
    assert!(en.contains("Partly applied"), "{en}");
    assert!(en.contains("left to apply: agents"), "{en}");
    assert!(en.contains("run moai update again"), "{en}");
    assert!(en.contains("every hash matches"), "{en}");
    let ko = screen(&partial, Asked::Yes, &view(Lang::Ko));
    assert!(ko.contains("일부만 바꿨어요"), "{ko}");
    assert!(ko.contains("남은 부분:"), "{ko}");
    assert!(ko.contains("moai update 를 다시 실행하면"), "{ko}");

    let p = result(&run(&exe, &home, &["update"]));
    let done = result(&run(&exe, &home, &["update", "--confirm", &digest_of(&p)]));
    let en = screen(&done, Asked::Yes, &view(Lang::En));
    assert!(en.contains("Done"), "{en}");
    assert!(!en.contains("left to apply"), "{en}");
    assert!(!en.contains("Partly applied"), "{en}");

    // The same clean parts, a receipt that read one file back differently.
    // Every part looks clean, and only the receipt says a part is left: the head and the line
    // that names it come from the receipt alone.
    let mut open = done.clone();
    open["run"]["status"] = json!("partial");
    open["run"]["remaining"] = json!(["routing"]);
    let en = screen(&open, Asked::Yes, &view(Lang::En));
    assert!(en.contains("Partly applied"), "{en}");
    assert!(en.contains("left to apply: routing"), "{en}");
    // (The rerun wrote no skill file, so the receipt names none; one is added to it.)
    assert_eq!(done["run"]["files"], json!([]), "{}", done["run"]);
    let mut tampered = done.clone();
    tampered["run"]["files"] = json!([{
        "part": "skills", "host": "claude", "path": "SKILL.md", "check": "mismatch",
    }]);
    tampered["run"]["files_checked"] = json!(1);
    tampered["run"]["files_not_matching"] = json!(1);
    tampered["run"]["status"] = json!("partial");
    let en = screen(&tampered, Asked::Yes, &view(Lang::En));
    assert!(en.contains("Partly applied"), "{en}");
    assert!(en.contains("SKILL.md: read back differently"), "{en}");
    assert!(en.contains("(mismatch)"), "{en}");
    assert!(!en.contains("every hash matches"), "{en}");
}
