//! #2592: `skills-legacy HOME` lists the Go generation's `moai-*` skills beside `moai` in the
//! Claude skill folder of a temporary HOME and changes nothing. The Codex, OMP and Cursor skill
//! folders are not read (SPEC-20261122-codex-home-update-removed).
mod common;
use serde_json::Value;
use std::{fs, path::Path, process::Command};

fn run(home: &Path) -> Value {
    let out = Command::new(crate::common::MOAI_BIN)
        .args(["skills-legacy", &home.to_string_lossy()])
        .env("HOME", home)
        .output()
        .unwrap();
    let env: Value = serde_json::from_slice(&out.stdout).unwrap();
    assert_eq!(env["ok"], true, "{env}");
    env["result"].clone()
}

fn skill(dir: &Path, name: &str) {
    fs::create_dir_all(dir).unwrap();
    fs::write(
        dir.join("SKILL.md"),
        format!("---\nname: {name}\ndescription: x\n---\nbody\n"),
    )
    .unwrap();
}

/// A tree snapshot (path, bytes) to prove nothing was written.
fn snapshot(root: &Path) -> Vec<(String, Vec<u8>)> {
    let mut out = Vec::new();
    let mut stack = vec![root.to_path_buf()];
    while let Some(dir) = stack.pop() {
        for e in fs::read_dir(&dir).unwrap().flatten() {
            let p = e.path();
            let m = fs::symlink_metadata(&p).unwrap();
            if m.is_dir() {
                stack.push(p.clone());
            }
            let bytes = if m.is_file() {
                fs::read(&p).unwrap()
            } else {
                vec![]
            };
            out.push((p.to_string_lossy().to_string(), bytes));
        }
    }
    out.sort();
    out
}

/// An empty HOME says `0건` and reports the Claude skill folder as absent.
#[test]
fn an_empty_home_reports_zero_in_words() {
    let home = tempfile::tempdir().unwrap();
    let r = run(home.path());
    assert_eq!(r["count"], 0, "{r}");
    assert!(r["summary"].as_str().unwrap().starts_with("0건"), "{r}");
    assert_eq!(r["roots"].as_array().unwrap().len(), 1);
    for root in r["roots"].as_array().unwrap() {
        assert_eq!(root["state"], "absent", "{root}");
    }
    assert_eq!(r["writes"], false);
}

/// Legacy folders are listed in the Claude skill folder; `moai` itself and other skills are
/// not; a link is reported as a link; the folders of the hosts moai no longer manages are not
/// read; nothing in HOME changes.
#[cfg(unix)]
#[test]
fn legacy_moai_skills_are_listed_and_nothing_changes() {
    let home = tempfile::tempdir().unwrap();
    let h = fs::canonicalize(home.path()).unwrap();
    skill(&h.join(".claude/skills/moai"), "moai");
    skill(&h.join(".claude/skills/moai-plan"), "moai-plan");
    skill(&h.join(".claude/skills/moai-sync"), "moai-sync");
    skill(&h.join(".claude/skills/other"), "other");
    skill(&h.join("elsewhere/moai-x"), "moai-x");
    std::os::unix::fs::symlink(
        h.join("elsewhere/moai-x"),
        h.join(".claude/skills/moai-linked"),
    )
    .unwrap();
    // Folders of the hosts moai no longer manages hold legacy skills too; they are not read.
    skill(&h.join(".agents/skills/moai-run"), "moai-run");
    skill(&h.join(".codex/skills/moai-run"), "moai-run");
    skill(&h.join(".omp/agent/skills/moai-run"), "moai-run");
    skill(&h.join(".cursor/skills/moai-run"), "moai-run");
    let before = snapshot(&h);

    let r = run(&h);
    assert_eq!(r["count"], 3, "{r}");
    assert!(r["summary"].as_str().unwrap().starts_with("3건"), "{r}");
    let roots: Vec<&str> = r["roots"]
        .as_array()
        .unwrap()
        .iter()
        .map(|x| x["root"].as_str().unwrap())
        .collect();
    assert_eq!(roots, [".claude/skills"], "{r}");
    let claude = r["roots"][0].clone();
    assert_eq!(claude["read_by"], "claude");
    let kinds: Vec<(&str, &str)> = claude["legacy"]
        .as_array()
        .unwrap()
        .iter()
        .map(|l| (l["folder"].as_str().unwrap(), l["kind"].as_str().unwrap()))
        .collect();
    assert_eq!(
        kinds,
        [
            ("moai-linked", "link"),
            ("moai-plan", "folder"),
            ("moai-sync", "folder")
        ]
    );
    assert_eq!(claude["legacy"][1]["name"], "moai-plan");
    assert_eq!(claude["legacy"][1]["skill_md"], "file");
    assert_eq!(
        claude["legacy"][0]["name"],
        Value::Null,
        "a link is not followed"
    );
    assert!(
        r.get("codex_disable_entries").is_none(),
        "no Codex entry is printed: {r}"
    );
    assert_eq!(before, snapshot(&h), "skills-legacy writes nothing");
}

fn legacy_entry(r: &Value, root: &str, folder: &str) -> Value {
    r["roots"]
        .as_array()
        .unwrap()
        .iter()
        .find(|x| x["root"] == root)
        .unwrap()["legacy"]
        .as_array()
        .unwrap()
        .iter()
        .find(|l| l["folder"] == folder)
        .unwrap_or_else(|| panic!("{folder} not listed: {r}"))
        .clone()
}

/// Pre-audit 6852075845 B1: a FIFO `SKILL.md` is reported and never read, so the check ends
/// at once instead of waiting for a writer.
#[cfg(unix)]
#[test]
fn a_fifo_skill_md_is_reported_and_never_read() {
    let home = tempfile::tempdir().unwrap();
    let h = home.path();
    let dir = h.join(".claude/skills/moai-b");
    fs::create_dir_all(&dir).unwrap();
    let fifo = std::ffi::CString::new(dir.join("SKILL.md").to_string_lossy().as_bytes()).unwrap();
    assert_eq!(unsafe { libc::mkfifo(fifo.as_ptr(), 0o600) }, 0);
    let mut child = Command::new(crate::common::MOAI_BIN)
        .args(["skills-legacy", &h.to_string_lossy()])
        .env("HOME", h)
        .stdout(std::process::Stdio::piped())
        .spawn()
        .unwrap();
    let start = std::time::Instant::now();
    let status = loop {
        if let Some(s) = child.try_wait().unwrap() {
            break s;
        }
        if start.elapsed() > std::time::Duration::from_secs(10) {
            child.kill().unwrap();
            panic!("skills-legacy blocked on a FIFO SKILL.md");
        }
        std::thread::sleep(std::time::Duration::from_millis(20));
    };
    assert!(status.success());
    let mut out = Vec::new();
    std::io::Read::read_to_end(&mut child.stdout.take().unwrap(), &mut out).unwrap();
    let r = serde_json::from_slice::<Value>(&out).unwrap()["result"].clone();
    let entry = legacy_entry(&r, ".claude/skills", "moai-b");
    assert_eq!(entry["skill_md"], "not_a_file", "{entry}");
    assert_eq!(entry["name"], Value::Null);
}

/// Pre-audit 6852075845 B2: a linked `SKILL.md` is reported as a link and never read, so a
/// file outside HOME never reaches the output.
#[cfg(unix)]
#[test]
fn a_linked_skill_md_is_never_followed() {
    let home = tempfile::tempdir().unwrap();
    let outside = tempfile::tempdir().unwrap();
    let h = home.path();
    let target = outside.path().join("target.md");
    fs::write(&target, "---\nname: secret-name-from-outside\n---\n").unwrap();
    let dir = h.join(".claude/skills/moai-a");
    fs::create_dir_all(&dir).unwrap();
    std::os::unix::fs::symlink(&target, dir.join("SKILL.md")).unwrap();
    let r = run(h);
    let entry = legacy_entry(&r, ".claude/skills", "moai-a");
    assert_eq!(entry["skill_md"], "link", "{entry}");
    assert_eq!(entry["name"], Value::Null);
    assert!(!r.to_string().contains("secret-name-from-outside"), "{r}");
}

/// Pre-audit N2: a link between HOME and the skill folder (`.claude` itself) is not followed.
#[cfg(unix)]
#[test]
fn a_linked_folder_above_the_skill_root_is_not_followed() {
    let home = tempfile::tempdir().unwrap();
    let h = home.path();
    skill(&h.join("real/skills/moai-plan"), "moai-plan");
    std::os::unix::fs::symlink(h.join("real"), h.join(".claude")).unwrap();
    let r = run(h);
    assert_eq!(r["count"], 0, "{r}");
    let claude = r["roots"]
        .as_array()
        .unwrap()
        .iter()
        .find(|x| x["root"] == ".claude/skills")
        .unwrap()
        .clone();
    assert_eq!(claude["state"], "link_not_followed", "{claude}");
}

/// B1 repair bound, as literals: a 65536-byte `SKILL.md` is read, a 65537-byte one is not.
#[test]
fn skill_md_is_read_up_to_64_kib_exactly() {
    let home = tempfile::tempdir().unwrap();
    let h = home.path();
    for (folder, len) in [("moai-at", 65536usize), ("moai-over", 65537usize)] {
        let dir = h.join(".claude/skills").join(folder);
        fs::create_dir_all(&dir).unwrap();
        let mut bytes = format!("---\nname: {folder}\n---\n").into_bytes();
        bytes.resize(len, b'x');
        fs::write(dir.join("SKILL.md"), bytes).unwrap();
    }
    let r = run(h);
    let at = legacy_entry(&r, ".claude/skills", "moai-at");
    assert_eq!(at["skill_md"], "file", "{at}");
    assert_eq!(at["name"], "moai-at");
    let over = legacy_entry(&r, ".claude/skills", "moai-over");
    assert_eq!(over["skill_md"], "too_large", "{over}");
    assert_eq!(over["name"], Value::Null);
}

/// #2758 audit N2: a file where a folder between HOME and the skill root should be is reported
/// as `not_a_folder`, not as unreadable.
#[test]
fn a_file_in_place_of_a_folder_on_the_way_is_not_a_folder() {
    let home = tempfile::tempdir().unwrap();
    let h = home.path();
    fs::write(h.join(".claude"), "not a folder\n").unwrap();
    let r = run(h);
    let claude = r["roots"]
        .as_array()
        .unwrap()
        .iter()
        .find(|x| x["root"] == ".claude/skills")
        .unwrap()
        .clone();
    assert_eq!(claude["state"], "not_a_folder", "{claude}");
}

/// A linked Claude skill folder is not followed.
#[cfg(unix)]
#[test]
fn a_linked_skill_root_is_not_followed() {
    let home = tempfile::tempdir().unwrap();
    let h = home.path();
    skill(&h.join("real/moai-plan"), "moai-plan");
    fs::create_dir_all(h.join(".claude")).unwrap();
    std::os::unix::fs::symlink(h.join("real"), h.join(".claude/skills")).unwrap();
    let r = run(h);
    assert_eq!(r["count"], 0, "{r}");
    let claude = r["roots"]
        .as_array()
        .unwrap()
        .iter()
        .find(|x| x["root"] == ".claude/skills")
        .unwrap()
        .clone();
    assert_eq!(claude["state"], "link_not_followed");
}
