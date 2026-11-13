//! #2592 Codex parity: `skills-legacy HOME` lists the Go generation's `moai-*` skills beside
//! `moai` in a temporary HOME and changes nothing.
use serde_json::Value;
use std::{fs, path::Path, process::Command};

fn run(home: &Path) -> Value {
    let out = Command::new(env!("CARGO_BIN_EXE_moai-cli"))
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
    assert_eq!(agents["legacy"][0]["skill_md"], true);
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
