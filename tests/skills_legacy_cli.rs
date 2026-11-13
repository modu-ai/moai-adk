//! #2592 Codex parity: `skills-legacy HOME` lists the Go generation's `moai-*` skills beside
//! `moai` in a temporary HOME and changes nothing.
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

/// An empty HOME says `0건` and reports every host folder as absent.
#[test]
fn an_empty_home_reports_zero_in_words() {
    let home = tempfile::tempdir().unwrap();
    let r = run(home.path());
    assert_eq!(r["count"], 0, "{r}");
    assert!(r["summary"].as_str().unwrap().starts_with("0건"), "{r}");
    assert_eq!(r["roots"].as_array().unwrap().len(), 5);
    for root in r["roots"].as_array().unwrap() {
        assert_eq!(root["state"], "absent", "{root}");
    }
    assert_eq!(r["writes"], false);
}

/// Legacy folders are listed per host folder; `moai` itself and other skills are not; a link
/// is reported as a link; the Codex disable entries cover only `.agents/skills`; nothing in HOME
/// changes.
#[cfg(unix)]
#[test]
fn legacy_moai_skills_are_listed_and_nothing_changes() {
    let home = tempfile::tempdir().unwrap();
    let h = fs::canonicalize(home.path()).unwrap();
    skill(&h.join(".agents/skills/moai"), "moai");
    skill(&h.join(".agents/skills/moai-plan"), "moai-plan");
    skill(&h.join(".agents/skills/moai-run"), "moai-run");
    skill(&h.join(".agents/skills/other"), "other");
    skill(&h.join(".claude/skills/moai-sync"), "moai-sync");
    skill(&h.join("elsewhere/moai-x"), "moai-x");
    std::os::unix::fs::symlink(
        h.join("elsewhere/moai-x"),
        h.join(".claude/skills/moai-linked"),
    )
    .unwrap();
    let before = snapshot(&h);

    let r = run(&h);
    assert_eq!(r["count"], 4, "{r}");
    assert!(r["summary"].as_str().unwrap().starts_with("4건"), "{r}");
    let root = |rel: &str| {
        r["roots"]
            .as_array()
            .unwrap()
            .iter()
            .find(|x| x["root"] == rel)
            .unwrap()
            .clone()
    };
    let agents = root(".agents/skills");
    let folders: Vec<&str> = agents["legacy"]
        .as_array()
        .unwrap()
        .iter()
        .map(|l| l["folder"].as_str().unwrap())
        .collect();
    assert_eq!(folders, ["moai-plan", "moai-run"]);
    assert_eq!(agents["legacy"][0]["name"], "moai-plan");
    assert_eq!(agents["legacy"][0]["skill_md"], "file");
    let claude = root(".claude/skills");
    let kinds: Vec<(&str, &str)> = claude["legacy"]
        .as_array()
        .unwrap()
        .iter()
        .map(|l| (l["folder"].as_str().unwrap(), l["kind"].as_str().unwrap()))
        .collect();
    assert_eq!(kinds, [("moai-linked", "link"), ("moai-sync", "folder")]);
    assert_eq!(
        claude["legacy"][0]["name"],
        Value::Null,
        "a link is not followed"
    );
    let entries = r["codex_disable_entries"].as_array().unwrap();
    assert_eq!(entries.len(), 2);
    assert_eq!(
        entries[0],
        format!(
            "[[skills.config]]\npath = \"{}/SKILL.md\"\nenabled = false",
            h.join(".agents/skills/moai-plan").display()
        )
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
    let dir = h.join(".agents/skills/moai-b");
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
    let entry = legacy_entry(&r, ".agents/skills", "moai-b");
    assert_eq!(entry["skill_md"], "not_a_file", "{entry}");
    assert_eq!(entry["name"], Value::Null);
    assert_eq!(r["codex_disable_entries"].as_array().unwrap().len(), 0);
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
    let dir = h.join(".agents/skills/moai-a");
    fs::create_dir_all(&dir).unwrap();
    std::os::unix::fs::symlink(&target, dir.join("SKILL.md")).unwrap();
    let r = run(h);
    let entry = legacy_entry(&r, ".agents/skills", "moai-a");
    assert_eq!(entry["skill_md"], "link", "{entry}");
    assert_eq!(entry["name"], Value::Null);
    assert!(!r.to_string().contains("secret-name-from-outside"), "{r}");
    assert_eq!(r["codex_disable_entries"].as_array().unwrap().len(), 0);
}

/// Pre-audit N2: a link between HOME and the skill folder (`.agents` itself) is not followed.
#[cfg(unix)]
#[test]
fn a_linked_folder_above_the_skill_root_is_not_followed() {
    let home = tempfile::tempdir().unwrap();
    let h = home.path();
    skill(&h.join("real/skills/moai-plan"), "moai-plan");
    std::os::unix::fs::symlink(h.join("real"), h.join(".agents")).unwrap();
    let r = run(h);
    assert_eq!(r["count"], 0, "{r}");
    let agents = r["roots"]
        .as_array()
        .unwrap()
        .iter()
        .find(|x| x["root"] == ".agents/skills")
        .unwrap()
        .clone();
    assert_eq!(agents["state"], "link_not_followed", "{agents}");
}

/// B1 repair bound, as literals: a 65536-byte `SKILL.md` is read, a 65537-byte one is not.
#[test]
fn skill_md_is_read_up_to_64_kib_exactly() {
    let home = tempfile::tempdir().unwrap();
    let h = home.path();
    for (folder, len) in [("moai-at", 65536usize), ("moai-over", 65537usize)] {
        let dir = h.join(".agents/skills").join(folder);
        fs::create_dir_all(&dir).unwrap();
        let mut bytes = format!("---\nname: {folder}\n---\n").into_bytes();
        bytes.resize(len, b'x');
        fs::write(dir.join("SKILL.md"), bytes).unwrap();
    }
    let r = run(h);
    let at = legacy_entry(&r, ".agents/skills", "moai-at");
    assert_eq!(at["skill_md"], "file", "{at}");
    assert_eq!(at["name"], "moai-at");
    let over = legacy_entry(&r, ".agents/skills", "moai-over");
    assert_eq!(over["skill_md"], "too_large", "{over}");
    assert_eq!(over["name"], Value::Null);
}

/// Pre-audit N1: the Codex entry escapes the path as a TOML basic string, so a `"` or `\`
/// in HOME cannot break or redirect it.
#[cfg(unix)]
#[test]
fn the_codex_entry_escapes_the_path() {
    let base = tempfile::tempdir().unwrap();
    let h = base.path().join("q\"uo\\te");
    skill(&h.join(".agents/skills/moai-plan"), "moai-plan");
    let r = run(&h);
    let entries = r["codex_disable_entries"].as_array().unwrap();
    assert_eq!(entries.len(), 1, "{r}");
    let escaped = format!("{}/.agents/skills/moai-plan/SKILL.md", h.display())
        .replace('\\', "\\\\")
        .replace('"', "\\\"");
    assert_eq!(
        entries[0],
        format!("[[skills.config]]\npath = \"{escaped}\"\nenabled = false")
    );
}

/// #2758 audit N1: a control character in the path (a newline in HOME) is escaped as `\uXXXX`,
/// so the pasted entry stays one TOML line.
#[cfg(unix)]
#[test]
fn the_codex_entry_escapes_control_characters() {
    let base = tempfile::tempdir().unwrap();
    let h = base.path().join("line\nbreak");
    skill(&h.join(".agents/skills/moai-plan"), "moai-plan");
    let r = run(&h);
    let entry = r["codex_disable_entries"][0].as_str().unwrap().to_string();
    assert!(entry.contains("line\\u000Abreak"), "{entry}");
    assert_eq!(entry.lines().count(), 3, "{entry}");
}

/// #2758 audit N2: a file where a folder between HOME and the skill root should be is reported
/// as `not_a_folder`, not as unreadable.
#[test]
fn a_file_in_place_of_a_folder_on_the_way_is_not_a_folder() {
    let home = tempfile::tempdir().unwrap();
    let h = home.path();
    fs::write(h.join(".agents"), "not a folder\n").unwrap();
    let r = run(h);
    let agents = r["roots"]
        .as_array()
        .unwrap()
        .iter()
        .find(|x| x["root"] == ".agents/skills")
        .unwrap()
        .clone();
    assert_eq!(agents["state"], "not_a_folder", "{agents}");
}

/// A linked host skill folder is not followed.
#[cfg(unix)]
#[test]
fn a_linked_skill_root_is_not_followed() {
    let home = tempfile::tempdir().unwrap();
    let h = home.path();
    skill(&h.join("real/moai-plan"), "moai-plan");
    fs::create_dir_all(h.join(".agents")).unwrap();
    std::os::unix::fs::symlink(h.join("real"), h.join(".agents/skills")).unwrap();
    let r = run(h);
    assert_eq!(r["count"], 0, "{r}");
    let agents = r["roots"]
        .as_array()
        .unwrap()
        .iter()
        .find(|x| x["root"] == ".agents/skills")
        .unwrap()
        .clone();
    assert_eq!(agents["state"], "link_not_followed");
}
