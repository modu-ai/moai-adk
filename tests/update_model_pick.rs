//! SPEC-20261118-update-model-pick (card C-47 slice 3): the model stage of `moai update`. The
//! stage tests feed a captured input and read a captured output; the command test runs copies of
//! the built binary inside a temporary HOME. The real HOME is never read or written.

#![cfg(unix)]

mod common;
use common::CleanEnv;

use moai_cli::model_effort_questions::tool_question;
use moai_cli::update_models::{installed_tools, report, settings_path, stage};
use serde_json::{json, Value};
use std::fs;
use std::io::{BufReader, Cursor, Read};
use std::path::{Path, PathBuf};
use std::process::{Command, Output, Stdio};
use tempfile::TempDir;

const EXE: &str = crate::common::MOAI_BIN;

struct Home {
    _tmp: TempDir,
    path: PathBuf,
}

fn home() -> Home {
    let tmp = TempDir::new().unwrap();
    let path = fs::canonicalize(tmp.path()).unwrap().join("home");
    fs::create_dir_all(&path).unwrap();
    Home { _tmp: tmp, path }
}

fn write_settings(home: &Path, body: &str) -> PathBuf {
    let path = settings_path(home);
    fs::create_dir_all(path.parent().unwrap()).unwrap();
    fs::write(&path, body).unwrap();
    path
}

fn run_stage(home: &Path, input: &str) -> (Value, String) {
    let mut out = Vec::new();
    let result = stage(home, &mut Cursor::new(input.to_string()), &mut out);
    (result, String::from_utf8(out).unwrap())
}

fn enters(count: usize) -> String {
    "\n".repeat(count)
}

#[test]
fn the_tool_pick_comes_first_and_an_empty_pick_asks_nothing() {
    // Nothing installed: Enter picks no tool, so no model question follows and nothing is written.
    let h = home();
    let (result, out) = run_stage(&h.path, "\n");
    assert_eq!(result["state"], json!("nothing_picked"), "{result}");
    assert!(out.starts_with(&tool_question(&[]).prompt), "{out}");
    assert!(
        !out.contains("claude-opus-5-5"),
        "no model question after an empty pick: {out}"
    );
    assert!(!settings_path(&h.path).exists());

    // Three answers that are no choice end as cancelled, still without a write.
    let (result, out) = run_stage(&h.path, "q\nq\nq\n");
    assert_eq!(result["state"], json!("cancelled"), "{result}");
    assert!(out.starts_with(&tool_question(&[]).prompt), "{out}");
    assert!(!settings_path(&h.path).exists());
}

#[test]
fn only_supported_values_are_offered_and_an_unknown_saved_value_is_kept() {
    let h = home();
    fs::create_dir_all(h.path.join(".claude")).unwrap();
    let body = r#"{"launcher":{"role_access":"prompt","models":{"omp":{"leader":{"model":"SECRET-MODEL","effort":"ultra"}}}}}"#;
    let path = write_settings(&h.path, body);
    // Tool pick Enter (claude is installed), four Enters for the two roles, then no save answer.
    let (result, out) = run_stage(&h.path, &enters(5));
    assert_eq!(result["state"], json!("declined"), "{result}");
    assert!(out.contains("claude-opus-5-5") && out.contains("claude-sonnet-5-5"));
    assert!(
        !out.contains("gpt-6.1") && !out.contains("zai/glm"),
        "{out}"
    );
    assert!(
        out.contains("launcher.models.omp.leader"),
        "the unsupported entry is named by key path: {out}"
    );
    assert!(
        !out.contains("SECRET-MODEL") && !result.to_string().contains("SECRET-MODEL"),
        "saved text is never quoted"
    );
    assert_eq!(fs::read_to_string(&path).unwrap(), body, "nothing written");

    // Saved after a yes, the unsupported entry and the other keys are still there.
    let (result, _) = run_stage(&h.path, &format!("{}y\n", enters(5)));
    assert_eq!(result["state"], json!("applied"), "{result}");
    let saved: Value = serde_json::from_str(&fs::read_to_string(&path).unwrap()).unwrap();
    assert_eq!(saved["launcher"]["role_access"], json!("prompt"));
    assert_eq!(
        saved["launcher"]["models"]["omp"]["leader"],
        json!({"model": "SECRET-MODEL", "effort": "ultra"})
    );
    assert_eq!(
        saved["launcher"]["models"]["claude"]["leader"],
        json!({"model": "claude-opus-5-5", "effort": "high"})
    );
}

#[test]
fn enter_selects_the_defaults_and_the_preview_shows_them() {
    let h = home();
    fs::create_dir_all(h.path.join(".claude")).unwrap();
    // A `.codex` folder is no installed tool any more: the pick offers `claude` only.
    fs::create_dir_all(h.path.join(".codex")).unwrap();
    // Tool pick, then four questions for the one tool, then no save answer.
    let (result, out) = run_stage(&h.path, &enters(5));
    assert_eq!(result["state"], json!("declined"), "{result}");
    for expected in ["claude-opus-5-5", "claude-sonnet-5-5"] {
        assert!(out.contains(expected), "{expected} in the preview: {out}");
    }
    assert!(
        !out.contains("gpt-6"),
        "no Codex model in the preview: {out}"
    );
    assert_eq!(result["changes"].as_array().unwrap().len(), 2, "{result}");
    assert!(
        !settings_path(&h.path).exists(),
        "the preview was shown and nothing is written yet"
    );
}

#[test]
fn the_choice_is_written_through_the_digest_bound_apply_and_read_back() {
    let h = home();
    fs::create_dir_all(h.path.join(".claude")).unwrap();
    let path = write_settings(&h.path, r#"{"keep":1,"launcher":{"advisor":"x"}}"#);
    let (result, _) = run_stage(&h.path, &format!("{}y\n", enters(5)));
    assert_eq!(result["state"], json!("applied"), "{result}");
    let saved: Value = serde_json::from_str(&fs::read_to_string(&path).unwrap()).unwrap();
    assert_eq!(saved["keep"], json!(1));
    assert_eq!(saved["launcher"]["advisor"], json!("x"));
    assert_eq!(
        saved["launcher"]["models"]["claude"]["agent"],
        json!({"model": "claude-sonnet-5-5", "effort": "high"})
    );
    // The report reads the same file back.
    let reported = report(&h.path);
    let claude_agent = reported["values"]
        .as_array()
        .unwrap()
        .iter()
        .find(|v| v["tool"] == json!("claude") && v["role"] == json!("agent"))
        .unwrap();
    assert_eq!(claude_agent["source"], json!("saved"), "{reported}");
}

/// SPEC-20261121-codex-launcher-and-models-removed REQ-5, REQ-6, REQ-8: a saved
/// `launcher.models.codex` value, whatever its shape, is no error of the model step, is not
/// listed as unsupported, and is left out of the file once a Claude choice is written (leader
/// decision D2).
#[test]
fn a_saved_codex_value_is_ignored_and_dropped_on_the_next_save() {
    for codex in [
        r#"{"agent":{"model":"gpt-6.1-terra","effort":"xhigh"}}"#,
        r#"{"agent":"SECRET-CODEX-MODEL"}"#,
        r#""SECRET-CODEX-MODEL""#,
    ] {
        let h = home();
        fs::create_dir_all(h.path.join(".claude")).unwrap();
        fs::create_dir_all(h.path.join(".codex")).unwrap();
        let body = format!(r#"{{"launcher":{{"models":{{"codex":{codex}}}}}}}"#);
        let path = write_settings(&h.path, &body);
        let (result, out) = run_stage(&h.path, &enters(5));
        assert_eq!(result["state"], json!("declined"), "{codex}: {result}");
        assert!(
            !out.contains("launcher.models.codex") && !out.contains("SECRET-CODEX-MODEL"),
            "{codex}: the saved Codex value is not listed or quoted: {out}"
        );
        assert_eq!(fs::read_to_string(&path).unwrap(), body, "nothing written");
        let (result, _) = run_stage(&h.path, &format!("{}y\n", enters(5)));
        assert_eq!(result["state"], json!("applied"), "{codex}: {result}");
        let saved: Value = serde_json::from_str(&fs::read_to_string(&path).unwrap()).unwrap();
        assert!(
            saved["launcher"]["models"].get("codex").is_none(),
            "{codex}: the Codex value stayed: {saved}"
        );
        assert!(saved["launcher"]["models"]["claude"].is_object(), "{saved}");
    }
}

#[test]
fn a_no_and_an_unchanged_pick_write_nothing() {
    let h = home();
    fs::create_dir_all(h.path.join(".claude")).unwrap();
    let (result, _) = run_stage(&h.path, &format!("{}n\n", enters(5)));
    assert_eq!(result["state"], json!("declined"), "{result}");
    assert!(!settings_path(&h.path).exists());

    let (result, _) = run_stage(&h.path, &format!("{}y\n", enters(5)));
    assert_eq!(result["state"], json!("applied"), "{result}");
    let path = settings_path(&h.path);
    let before = fs::read(&path).unwrap();
    // The same picks again: nothing changes, nothing is asked to save, the bytes stay.
    let (result, _) = run_stage(&h.path, &format!("{}y\n", enters(5)));
    assert_eq!(result["state"], json!("unchanged"), "{result}");
    assert_eq!(fs::read(&path).unwrap(), before);
}

/// Reads one line per `read` call and runs a hook just before the line with the given index.
struct Hooked {
    lines: Vec<String>,
    next: usize,
    hook_at: usize,
    hook: Option<Box<dyn FnOnce()>>,
}

impl Read for Hooked {
    fn read(&mut self, buf: &mut [u8]) -> std::io::Result<usize> {
        let Some(line) = self.lines.get(self.next) else {
            return Ok(0);
        };
        if self.next == self.hook_at {
            if let Some(hook) = self.hook.take() {
                hook();
            }
        }
        let bytes = line.as_bytes();
        buf[..bytes.len()].copy_from_slice(bytes);
        self.next += 1;
        Ok(bytes.len())
    }
}

#[test]
fn a_broken_or_changed_settings_file_is_an_error_and_stays_as_it_was() {
    // A file that is no JSON-subset: reported as an error, never quoted, and left as it was.
    let h = home();
    fs::create_dir_all(h.path.join(".claude")).unwrap();
    let path = write_settings(&h.path, "SECRET-MARKER: [not json");
    let (result, out) = run_stage(&h.path, &format!("{}y\n", enters(5)));
    assert_eq!(result["state"], json!("error"), "{result}");
    assert!(!result.to_string().contains("SECRET-MARKER"));
    assert!(!out.contains("SECRET-MARKER"));
    assert_eq!(
        fs::read_to_string(&path).unwrap(),
        "SECRET-MARKER: [not json"
    );

    // A file that changes between the preview and the save answer: the digest no longer matches.
    let h = home();
    fs::create_dir_all(h.path.join(".claude")).unwrap();
    let path = write_settings(&h.path, r#"{"launcher":{}}"#);
    let changed = r#"{"launcher":{"role_access":"prompt"}}"#;
    let target = path.clone();
    let mut lines: Vec<String> = (0..5).map(|_| "\n".to_string()).collect();
    lines.push("y\n".to_string());
    let hooked = Hooked {
        lines,
        next: 0,
        hook_at: 5,
        hook: Some(Box::new(move || fs::write(&target, changed).unwrap())),
    };
    let mut out = Vec::new();
    let result = stage(
        &h.path,
        &mut BufReader::with_capacity(4096, hooked),
        &mut out,
    );
    assert_eq!(result["state"], json!("error"), "{result}");
    assert_eq!(
        fs::read_to_string(&path).unwrap(),
        changed,
        "left as it was"
    );
}

#[test]
fn the_installed_tools_follow_the_home_folders() {
    let h = home();
    assert!(installed_tools(&h.path).is_empty());
    fs::create_dir_all(h.path.join(".claude")).unwrap();
    assert_eq!(installed_tools(&h.path), ["claude"]);
    fs::write(h.path.join(".codex"), "a file, not a folder").unwrap();
    assert_eq!(
        installed_tools(&h.path),
        ["claude"],
        "a file is not a folder"
    );
    fs::remove_file(h.path.join(".codex")).unwrap();
    fs::create_dir_all(h.path.join(".codex")).unwrap();
    assert_eq!(
        installed_tools(&h.path),
        ["claude"],
        "a `.codex` folder is no installed tool"
    );
}

// The command: a run without a terminal asks nothing and reports the values.

fn skill_md(body: &str) -> String {
    format!("---\nname: moai\ndescription: fixture package\n---\n# MoAI\n{body}\n")
}

fn install_version(home: &Path) -> PathBuf {
    let folder = home.join(".moai-cli/versions/a");
    let bin = folder.join("bin");
    fs::create_dir_all(&bin).unwrap();
    fs::create_dir_all(folder.join("agents")).unwrap();
    let exe = bin.join("moai-cli");
    common::copy_executable(Path::new(EXE), &exe);
    fs::hard_link(&exe, bin.join("moai")).unwrap();
    let p = folder.join("skills/moai/SKILL.md");
    fs::create_dir_all(p.parent().unwrap()).unwrap();
    fs::write(p, skill_md("body v1")).unwrap();
    exe
}

fn run(exe: &Path, home: &Path, args: &[&str]) -> Output {
    let mut child = Command::new(exe)
        .args(args)
        .clean_env()
        .env("PATH", "/usr/bin:/bin")
        .env("HOME", home)
        .env("LC_ALL", "C")
        .current_dir(home)
        .stdin(Stdio::null())
        .stdout(Stdio::piped())
        .stderr(Stdio::piped())
        .spawn()
        .unwrap();
    let _ = child.stdin.take();
    child.wait_with_output().unwrap()
}

fn result_of(out: &Output) -> Value {
    let v: Value = serde_json::from_slice(&out.stdout).unwrap_or_else(|_| {
        panic!(
            "envelope JSON on stdout: {}{}",
            String::from_utf8_lossy(&out.stdout),
            String::from_utf8_lossy(&out.stderr)
        )
    });
    assert_eq!(v["ok"], json!(true), "{v}");
    v["result"].clone()
}

#[test]
fn a_non_terminal_run_asks_nothing_and_reports_the_values() {
    let h = home();
    let exe = install_version(&h.path);
    let body = r#"{"launcher":{"models":{"claude":{"agent":{"model":"claude-sonnet-5-5","effort":"max"}},"omp":{"leader":{"model":"SECRET-MODEL","effort":"ultra"}}}}}"#;
    let path = write_settings(&h.path, body);
    // A preview keeps exactly the keys it had: no `models`.
    let preview = result_of(&run(&exe, &h.path, &["update"]));
    assert!(preview.get("models").is_none(), "{preview}");
    let digest = preview["plan_digest"].as_str().unwrap().to_string();
    // The applied result carries the earlier keys and the new one.
    let out = run(&exe, &h.path, &["update", "--confirm", &digest]);
    let r = result_of(&out);
    for key in ["mode", "applied", "plan_digest", "result", "revert"] {
        assert!(r.get(key).is_some(), "{key} in {r}");
    }
    let models = &r["models"];
    assert_eq!(models["state"], json!("ok"), "{r}");
    let values = models["values"].as_array().unwrap();
    assert_eq!(values.len(), 6, "{models}");
    assert!(
        values.iter().all(|v| v["tool"] != json!("codex")),
        "no Codex row in the report: {models}"
    );
    let find = |tool: &str, role: &str| -> Value {
        values
            .iter()
            .find(|v| v["tool"] == json!(tool) && v["role"] == json!(role))
            .unwrap()
            .clone()
    };
    assert_eq!(
        find("claude", "agent"),
        json!({"tool": "claude", "role": "agent", "model": "claude-sonnet-5-5", "effort": "max", "source": "saved"})
    );
    assert_eq!(
        find("claude", "leader"),
        json!({"tool": "claude", "role": "leader", "model": "claude-opus-5-5", "effort": "high", "source": "default"})
    );
    assert_eq!(models["unsupported"], json!(["launcher.models.omp.leader"]));
    assert!(!models.to_string().contains("SECRET-MODEL"));
    assert_eq!(
        fs::read_to_string(&path).unwrap(),
        body,
        "a run without a terminal never writes the settings"
    );
    assert!(
        String::from_utf8_lossy(&out.stderr).trim().is_empty()
            || !String::from_utf8_lossy(&out.stderr).contains("[y/N]"),
        "nothing is asked"
    );
}

#[test]
fn enter_keeps_a_saved_value_instead_of_the_table_default() {
    let h = home();
    fs::create_dir_all(h.path.join(".claude")).unwrap();
    let path = write_settings(
        &h.path,
        r#"{"launcher":{"models":{"claude":{"leader":{"model":"claude-sonnet-5-5","effort":"max"}}}}}"#,
    );
    // Enter at every question, then a yes: the saved leader stays, only the unsaved agent is new.
    let (result, out) = run_stage(&h.path, &format!("{}y\n", enters(5)));
    assert_eq!(result["state"], json!("applied"), "{result}");
    assert!(
        out.contains("claude leader: claude-sonnet-5-5 max -> claude-sonnet-5-5 max"),
        "the preview shows the saved leader kept: {out}"
    );
    let saved: Value = serde_json::from_str(&fs::read_to_string(&path).unwrap()).unwrap();
    assert_eq!(
        saved["launcher"]["models"]["claude"]["leader"],
        json!({"model": "claude-sonnet-5-5", "effort": "max"})
    );
    assert_eq!(
        saved["launcher"]["models"]["claude"]["agent"],
        json!({"model": "claude-sonnet-5-5", "effort": "high"})
    );
}

/// A project folder with `git init` and a stub `claude` on the path, for a launcher dry run.
fn launcher_dry_run(home: &Path, args: &[&str]) -> Value {
    let project = TempDir::new().unwrap();
    assert!(common::git::command()
        .args(["init", "-q"])
        .current_dir(project.path())
        .status()
        .unwrap()
        .success());
    let bin = TempDir::new().unwrap();
    let stub = bin.path().join("claude");
    fs::write(&stub, "#!/bin/sh\nexit 0\n").unwrap();
    fs::set_permissions(&stub, std::os::unix::fs::PermissionsExt::from_mode(0o755)).unwrap();
    let out = Command::new(EXE)
        .args(args)
        .arg("--dry-run")
        .current_dir(project.path())
        .env("PATH", format!("{}:/usr/bin:/bin", bin.path().display()))
        .env("HOME", home)
        .env("USERPROFILE", home)
        .env("ANTHROPIC_AUTH_TOKEN", "moai-test-dummy-not-real")
        .output()
        .unwrap();
    assert!(
        out.status.success(),
        "{}",
        String::from_utf8_lossy(&out.stderr)
    );
    serde_json::from_slice(&out.stdout).unwrap()
}

#[test]
fn the_settings_path_is_the_file_the_launcher_reads() {
    let h = home();
    // The launcher reads `launcher.advisor` from the user file; a value written at the path of
    // this module reaches the launcher command line, so both name the same file.
    write_settings(&h.path, r#"{"launcher":{"advisor":"fable"}}"#);
    let plan = launcher_dry_run(&h.path, &["cc", "-l"]);
    let argv: Vec<&str> = plan["argv"]
        .as_array()
        .unwrap()
        .iter()
        .map(|v| v.as_str().unwrap())
        .collect();
    assert!(
        argv.windows(2).any(|w| w == ["--advisor", "fable"]),
        "{argv:?}"
    );
}
