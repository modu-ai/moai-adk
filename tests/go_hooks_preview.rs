//! #2877 (I5): the preview of replacing a project's Go-generation hooks. Temporary projects
//! only; the preview reads and writes nothing.

#![cfg(unix)]

use moai_cli::go_hooks::preview;
use serde_json::{json, Value};
use std::fs;
use std::path::{Path, PathBuf};

fn go_handler(script: &str) -> Value {
    json!({
        "type": "command",
        "command": "bash",
        "args": ["-c", "[ -f \"$0\" ] && exec bash \"$0\"; exit 0",
                 format!("${{CLAUDE_PROJECT_DIR}}/.claude/hooks/moai/{script}")],
        "timeout": 10
    })
}

/// A project the Go generation set up: Go entries next to one handler of another tool, and the
/// script folder with one script no entry runs.
fn go_project() -> (tempfile::TempDir, PathBuf) {
    let dir = tempfile::tempdir().unwrap();
    let root = fs::canonicalize(dir.path()).unwrap();
    let settings = json!({
        "permissions": {"allow": ["Bash(ls:*)"]},
        "hooks": {
            "SessionStart": [{"matcher": "startup", "hooks": [go_handler("handle-session-start.sh")]}],
            "PreToolUse": [
                {"matcher": "Write|Edit|Bash", "hooks": [go_handler("handle-pre-tool.sh"),
                    {"type": "command", "command": "orca guard", "timeout": 5}]}
            ],
            "Stop": [{"hooks": [go_handler("handle-stop.sh")]}]
        }
    });
    fs::create_dir_all(root.join(".claude/hooks/moai")).unwrap();
    fs::write(
        root.join(".claude/settings.json"),
        serde_json::to_vec_pretty(&settings).unwrap(),
    )
    .unwrap();
    for s in [
        "handle-session-start.sh",
        "handle-pre-tool.sh",
        "handle-stop.sh",
        "handle-notification.sh",
    ] {
        fs::write(
            root.join(".claude/hooks/moai").join(s),
            "#!/bin/sh\nexit 0\n",
        )
        .unwrap();
    }
    (dir, root)
}

fn tree(root: &Path) -> Vec<(PathBuf, Vec<u8>)> {
    let mut out = Vec::new();
    let mut stack = vec![root.to_path_buf()];
    while let Some(d) = stack.pop() {
        for e in fs::read_dir(&d).unwrap().flatten() {
            let p = e.path();
            if p.is_dir() {
                stack.push(p);
            } else {
                out.push((p.clone(), fs::read(&p).unwrap()));
            }
        }
    }
    out.sort();
    out
}

#[test]
fn the_preview_lists_what_goes_what_stays_and_what_is_added_and_writes_nothing() {
    let (_d, root) = go_project();
    let before = tree(&root);
    let p = preview(&root).unwrap();
    assert_eq!(p["action"], json!("replace"), "{p}");
    assert_eq!(p["writes"], json!(false));
    let go: Vec<(String, String)> = p["go_entries"]
        .as_array()
        .unwrap()
        .iter()
        .map(|e| {
            (
                e["event"].as_str().unwrap().to_string(),
                e["script"].as_str().unwrap().to_string(),
            )
        })
        .collect();
    assert_eq!(
        go,
        // In the file's own order (serde_json keeps it).
        [
            (
                "SessionStart".to_string(),
                "handle-session-start.sh".to_string()
            ),
            ("PreToolUse".into(), "handle-pre-tool.sh".into()),
            ("Stop".into(), "handle-stop.sh".into()),
        ]
    );
    // The other tool's handler stays, in its own group.
    assert_eq!(
        p["kept_entries"],
        json!({"PreToolUse": [{"matcher": "Write|Edit|Bash",
            "hooks": [{"type": "command", "command": "orca guard", "timeout": 5}]}]})
    );
    // The MoAI set is added; nothing of the Go generation is left in the result.
    let after = p["hooks_after"].to_string();
    assert!(!after.contains(".claude/hooks/moai/"), "{after}");
    assert!(after.contains("moai-cli claim-verify"), "{after}");
    assert!(after.contains("orca guard"), "{after}");
    assert_eq!(
        p["scripts"],
        json!([
            "handle-notification.sh",
            "handle-pre-tool.sh",
            "handle-session-start.sh",
            "handle-stop.sh"
        ])
    );
    assert_eq!(p["scripts_unreferenced"], json!(["handle-notification.sh"]));
    assert_eq!(p["scripts_missing"], json!([]));
    assert_eq!(tree(&root), before, "the preview writes nothing");
}

#[test]
fn a_project_without_go_hooks_needs_nothing() {
    let dir = tempfile::tempdir().unwrap();
    let root = fs::canonicalize(dir.path()).unwrap();
    let p = preview(&root).unwrap();
    assert_eq!(p["action"], json!("none"));
    assert_eq!(p["settings_present"], json!(false));
    assert_eq!(p["go_entries"], json!([]));

    fs::create_dir_all(root.join(".claude")).unwrap();
    fs::write(
        root.join(".claude/settings.json"),
        r#"{"hooks":{"Stop":[{"hooks":[{"type":"command","command":"say done"}]}]}}"#,
    )
    .unwrap();
    let p = preview(&root).unwrap();
    assert_eq!(p["action"], json!("none"), "{p}");
    assert_eq!(
        p["kept_entries"],
        json!({"Stop":[{"hooks":[{"type":"command","command":"say done"}]}]})
    );
}

#[test]
fn an_entry_whose_script_is_gone_is_still_a_go_entry_and_is_named() {
    let (_d, root) = go_project();
    fs::remove_file(root.join(".claude/hooks/moai/handle-stop.sh")).unwrap();
    let p = preview(&root).unwrap();
    assert_eq!(p["scripts_missing"], json!(["handle-stop.sh"]));
    assert_eq!(p["go_entries"].as_array().unwrap().len(), 3);
}

#[test]
fn links_are_refused_and_nothing_is_followed() {
    // The settings file as a link.
    let (_d, root) = go_project();
    let real = root.join("real-settings.json");
    fs::rename(root.join(".claude/settings.json"), &real).unwrap();
    std::os::unix::fs::symlink(&real, root.join(".claude/settings.json")).unwrap();
    let e = preview(&root).unwrap_err().to_string();
    assert!(e.contains("is a link"), "{e}");

    // The script folder as a link.
    let (_d2, root) = go_project();
    let elsewhere = root.join("elsewhere");
    fs::rename(root.join(".claude/hooks/moai"), &elsewhere).unwrap();
    std::os::unix::fs::symlink(&elsewhere, root.join(".claude/hooks/moai")).unwrap();
    let e = preview(&root).unwrap_err().to_string();
    assert!(e.contains("is a link"), "{e}");

    // A link inside the script folder.
    let (_d3, root) = go_project();
    std::os::unix::fs::symlink(
        "/etc/hosts",
        root.join(".claude/hooks/moai/handle-extra.sh"),
    )
    .unwrap();
    let e = preview(&root).unwrap_err().to_string();
    assert!(e.contains("not a regular file"), "{e}");
}

#[test]
fn settings_that_are_not_a_json_object_are_refused() {
    let (_d, root) = go_project();
    fs::write(root.join(".claude/settings.json"), "[1, 2]").unwrap();
    let e = preview(&root).unwrap_err().to_string();
    assert!(e.contains("not a JSON object"), "{e}");
}
