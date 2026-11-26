//! SPEC-20261124-launcher-retire-role-flags (REQ-6, AC-9): the stage asks one model and one effort
//! per host, saves the flat `launcher.models.<host>` entry and lists a saved role entry as
//! replaced.
//!
//! SPEC-20261118-update-model-pick (card C-47 slice 3): the model stage of `moai update`. The
//! stage tests feed a captured input and read a captured output; the command test runs copies of
//! the built binary inside a temporary HOME. The real HOME is never read or written.

#![cfg(unix)]

mod common;
use common::CleanEnv;

use moai_cli::model_effort_questions::tool_question;
use moai_cli::update_models::{
    installed_tools, installed_tools_with, report, settings_path, stage_with,
};
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
    // `false`: the glm key is one line of the input, so a run on a developer's terminal never
    // opens the hidden prompt of that terminal.
    let result = stage_with(home, &mut Cursor::new(input.to_string()), &mut out, false);
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
    let body = r#"{"launcher":{"role_access":"prompt","models":{"glm":{"leader":{"model":"SECRET-MODEL","effort":"ultra"}}}}}"#;
    let path = write_settings(&h.path, body);
    // Tool pick Enter (claude is installed), Enter for the model and the effort, then no save answer.
    let (result, out) = run_stage(&h.path, &enters(3));
    assert_eq!(result["state"], json!("declined"), "{result}");
    assert!(out.contains("claude-opus-5-5") && out.contains("claude-sonnet-5-5"));
    assert!(
        !out.contains("gpt-6.1") && !out.contains("zai/glm"),
        "{out}"
    );
    assert!(
        out.contains("launcher.models.glm.leader"),
        "the unsupported entry is named by key path: {out}"
    );
    assert!(
        !out.contains("SECRET-MODEL") && !result.to_string().contains("SECRET-MODEL"),
        "saved text is never quoted"
    );
    assert_eq!(fs::read_to_string(&path).unwrap(), body, "nothing written");

    // Saved after a yes, the unsupported entry and the other keys are still there.
    let (result, _) = run_stage(&h.path, &format!("{}y\n", enters(3)));
    assert_eq!(result["state"], json!("applied"), "{result}");
    let saved: Value = serde_json::from_str(&fs::read_to_string(&path).unwrap()).unwrap();
    assert_eq!(saved["launcher"]["role_access"], json!("prompt"));
    assert_eq!(
        saved["launcher"]["models"]["glm"]["leader"],
        json!({"model": "SECRET-MODEL", "effort": "ultra"})
    );
    assert_eq!(
        saved["launcher"]["models"]["claude"],
        json!({"model": "claude-opus-5-5", "effort": "medium"})
    );
}

#[test]
fn enter_selects_the_defaults_and_the_preview_shows_them() {
    let h = home();
    fs::create_dir_all(h.path.join(".claude")).unwrap();
    // A `.codex` folder is no installed tool any more: the pick offers `claude` only.
    fs::create_dir_all(h.path.join(".codex")).unwrap();
    // Tool pick, then the model and the effort question of the one tool, then no save answer.
    let (result, out) = run_stage(&h.path, &enters(3));
    assert_eq!(result["state"], json!("declined"), "{result}");
    assert!(
        out.contains("claude: nothing saved -> claude-opus-5-5 medium"),
        "the preview shows the one pick of the host: {out}"
    );
    assert!(
        !out.contains("gpt-6"),
        "no Codex model in the preview: {out}"
    );
    assert_eq!(result["changes"].as_array().unwrap().len(), 1, "{result}");
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
    let (result, _) = run_stage(&h.path, &format!("{}y\n", enters(3)));
    assert_eq!(result["state"], json!("applied"), "{result}");
    let saved: Value = serde_json::from_str(&fs::read_to_string(&path).unwrap()).unwrap();
    assert_eq!(saved["keep"], json!(1));
    assert_eq!(saved["launcher"]["advisor"], json!("x"));
    assert_eq!(
        saved["launcher"]["models"]["claude"],
        json!({"model": "claude-opus-5-5", "effort": "medium"})
    );
    // The report reads the same file back.
    let reported = report(&h.path);
    let claude = reported["values"]
        .as_array()
        .unwrap()
        .iter()
        .find(|v| v["tool"] == json!("claude"))
        .unwrap();
    assert_eq!(claude["source"], json!("saved"), "{reported}");
}

/// SPEC-20261123-glm-54-followups AC-1 and AC-2 (REQ-3, REQ-4), as changed by the one-pick-per-host
/// screen: with the glm tool picked, Enter everywhere saves the 5.4 id at effort high as the one
/// glm entry, and the report reads the same value back as saved.
#[test]
fn enter_saves_glm_54_as_the_one_glm_entry() {
    let h = home();
    let path = write_settings(&h.path, r#"{"launcher":{"advisor":"x"}}"#);
    // Tool pick "2" is glm; then its model and effort; then Enter at the glm key question
    // (skip); then the yes.
    let (result, out) = run_stage(&h.path, &format!("2\n{}\ny\n", enters(2)));
    assert_eq!(result["state"], json!("applied"), "{result}");
    assert!(
        out.contains("zai/glm-5.4[1m]"),
        "the 5.4 option is offered: {out}"
    );
    let saved: Value = serde_json::from_str(&fs::read_to_string(&path).unwrap()).unwrap();
    assert_eq!(saved["launcher"]["advisor"], json!("x"));
    assert_eq!(
        saved["launcher"]["models"]["glm"],
        json!({"model": "zai/glm-5.4[1m]", "effort": "high"})
    );
    let reported = report(&h.path);
    let glm = reported["values"]
        .as_array()
        .unwrap()
        .iter()
        .find(|v| v["tool"] == json!("glm"))
        .unwrap();
    assert_eq!(glm["source"], json!("saved"), "{reported}");
}

/// SPEC-20261121-glm-key-at-pick AC-1, AC-2: a glm pick asks for the key between the model
/// questions and the save question, the typed key lands in the vault of the temporary home, and
/// the plain value is in no output, no launcher settings and no receipt under that home.
#[test]
fn a_typed_glm_key_is_vaulted_and_appears_nowhere_else() {
    const FAKE: &str = "FAKE-glm-key-for-the-stage-test";
    let h = home();
    let path = write_settings(&h.path, r#"{"launcher":{"advisor":"x"}}"#);
    let (result, out) = run_stage(&h.path, &format!("2\n{}{FAKE}\ny\n", enters(2)));
    assert_eq!(result["state"], json!("applied"), "{result}");
    assert!(
        out.contains("GLM API key"),
        "the key question is asked: {out}"
    );
    assert!(!out.contains(FAKE), "no echo of the key: {out}");
    assert!(
        !result.to_string().contains(FAKE),
        "the result names no key: {result}"
    );
    let scope = moai_cli::key_vault::project_scope_in(
        &std::env::current_dir().unwrap(),
        Some(&h.path),
    )
    .unwrap()
    .expect("the tests run inside a git work tree");
    assert_eq!(
        moai_cli::key_vault::read_entry(&h.path, &scope, "glm")
            .unwrap()
            .as_deref(),
        Some(FAKE),
        "the vault of the temporary home holds the key"
    );
    assert!(
        !fs::read_to_string(&path).unwrap().contains(FAKE),
        "the launcher settings hold no key"
    );
    let mut folders = vec![h.path.clone()];
    while let Some(folder) = folders.pop() {
        for entry in fs::read_dir(&folder).unwrap().flatten() {
            let kind = entry.file_type().unwrap();
            if kind.is_dir() {
                folders.push(entry.path());
            } else if kind.is_file() {
                let bytes = fs::read(entry.path()).unwrap();
                assert!(
                    !bytes.windows(FAKE.len()).any(|w| w == FAKE.as_bytes()),
                    "{} holds the plain key",
                    entry.path().display()
                );
            }
        }
    }
}

/// SPEC-20261121-glm-key-at-pick: a key vault that cannot be written is told on the screen, and
/// the stage goes on to the preview and the save question with the model pick kept.
#[test]
fn a_vault_that_cannot_be_written_does_not_lose_the_model_pick() {
    const FAKE: &str = "FAKE-glm-key-for-the-refused-write";
    let h = home();
    let path = write_settings(&h.path, r#"{"launcher":{"advisor":"x"}}"#);
    // A plain file where the vault folder belongs: every write into the vault is refused.
    let vault =
        moai_cli::key_vault::entry_path(&h.path, &moai_cli::key_vault::Scope::Shared, "glm")
            .parent()
            .and_then(Path::parent)
            .unwrap()
            .to_path_buf();
    fs::write(&vault, b"in the way").unwrap();
    let (result, out) = run_stage(&h.path, &format!("2\n{}{FAKE}\ny\n", enters(2)));
    assert_eq!(result["state"], json!("applied"), "{result}\n{out}");
    assert!(
        out.contains("The key could not be stored:"),
        "the screen tells the write failed: {out}"
    );
    assert!(!out.contains(FAKE), "no echo of the key: {out}");
    assert!(
        !result.to_string().contains(FAKE),
        "the result names no key: {result}"
    );
    let saved: Value = serde_json::from_str(&fs::read_to_string(&path).unwrap()).unwrap();
    assert_eq!(
        saved["launcher"]["models"]["glm"]["model"],
        json!("zai/glm-5.4[1m]"),
        "the model pick was saved: {saved}"
    );
}

/// SPEC-20261123-glm-54-followups AC-2 (REQ-2): a `zai/glm-5.3` entry saved before 5.4 is not
/// reported as unsupported, is the default of the next model step, and stays as it is when the
/// person presses Enter everywhere: nothing is asked to save and the bytes stay.
#[test]
fn a_saved_glm_53_entry_stays_through_enter_everywhere() {
    let h = home();
    let body = r#"{"launcher":{"models":{"glm":{"model":"zai/glm-5.3","effort":"low"}}}}"#;
    let path = write_settings(&h.path, body);
    let (result, out) = run_stage(&h.path, &format!("2\n{}", enters(2)));
    assert_eq!(result["state"], json!("unchanged"), "{result}");
    assert!(
        !out.contains("launcher.models.glm"),
        "a supported saved value is not listed as unsupported: {out}"
    );
    assert_eq!(fs::read_to_string(&path).unwrap(), body, "nothing written");
    let (result, _) = run_stage(&h.path, &format!("2\n{}\ny\n", enters(2)));
    assert_eq!(result["state"], json!("unchanged"), "{result}");
    assert_eq!(fs::read_to_string(&path).unwrap(), body, "the bytes stay");
}

/// SPEC-20261124-launcher-retire-role-flags AC-9 (REQ-6): a saved `leader` or `agent` entry is
/// listed among the keys the write replaces, in the preview and in the result, and the written
/// entry of that host is the flat pair alone. Nothing is written before the yes.
#[test]
fn a_saved_role_entry_is_listed_as_replaced_and_dropped_by_the_write() {
    let h = home();
    fs::create_dir_all(h.path.join(".claude")).unwrap();
    let body = r#"{"launcher":{"models":{"claude":{"leader":{"model":"claude-opus-5-5","effort":"high"},"agent":{"model":"SECRET-MODEL","effort":"ultra"}}}}}"#;
    let path = write_settings(&h.path, body);
    let (result, out) = run_stage(&h.path, &enters(3));
    assert_eq!(result["state"], json!("declined"), "{result}");
    for key in [
        "launcher.models.claude.leader",
        "launcher.models.claude.agent",
    ] {
        assert!(out.contains(key), "{key} in the preview: {out}");
        assert!(result["replaced"].to_string().contains(key), "{result}");
    }
    assert!(
        !out.contains("SECRET-MODEL") && !result.to_string().contains("SECRET-MODEL"),
        "saved text is never quoted"
    );
    assert_eq!(
        fs::read_to_string(&path).unwrap(),
        body,
        "nothing written before the yes"
    );
    let (result, _) = run_stage(&h.path, &format!("{}y\n", enters(3)));
    assert_eq!(result["state"], json!("applied"), "{result}");
    assert_eq!(
        result["replaced"],
        json!([
            "launcher.models.claude.leader",
            "launcher.models.claude.agent"
        ])
    );
    let saved: Value = serde_json::from_str(&fs::read_to_string(&path).unwrap()).unwrap();
    assert_eq!(
        saved["launcher"]["models"]["claude"],
        json!({"model": "claude-opus-5-5", "effort": "medium"}),
        "no role entry is left beside the flat pair"
    );
}

/// SPEC-20261124-launcher-retire-role-flags AC-9 (REQ-6), the case a pick that saves the value the
/// flat entry already has leaves open: the flat entry equals the picked pair and a `leader` and an
/// `agent` entry sit beside it. The role entries are still replaced, so the result is `applied`
/// (not `unchanged`), both keys are listed, and the file keeps the flat pair alone.
#[test]
fn a_flat_entry_equal_to_the_pick_still_replaces_the_role_entries_beside_it() {
    let h = home();
    fs::create_dir_all(h.path.join(".claude")).unwrap();
    // The first run saves the default pair as the flat entry.
    let (result, _) = run_stage(&h.path, &format!("{}y\n", enters(3)));
    assert_eq!(result["state"], json!("applied"), "{result}");
    let path = settings_path(&h.path);
    let mut saved: Value = serde_json::from_str(&fs::read_to_string(&path).unwrap()).unwrap();
    let flat = saved["launcher"]["models"]["claude"].clone();
    assert!(
        flat["model"].is_string() && flat["effort"].is_string(),
        "{flat}"
    );
    // Old role entries come back beside the same flat pair.
    saved["launcher"]["models"]["claude"]["leader"] =
        json!({"model": "OLD-LEADER", "effort": "low"});
    saved["launcher"]["models"]["claude"]["agent"] = json!({"model": "OLD-AGENT", "effort": "low"});
    fs::write(&path, saved.to_string()).unwrap();
    let before = fs::read_to_string(&path).unwrap();
    let (result, out) = run_stage(&h.path, &format!("{}n\n", enters(3)));
    assert_eq!(result["state"], json!("declined"), "{result}");
    assert_eq!(
        fs::read_to_string(&path).unwrap(),
        before,
        "nothing is written before the yes"
    );
    for key in [
        "launcher.models.claude.leader",
        "launcher.models.claude.agent",
    ] {
        assert!(out.contains(key), "{key} is listed as replaced: {out}");
    }
    let (result, _) = run_stage(&h.path, &format!("{}y\n", enters(3)));
    assert_eq!(
        result["state"],
        json!("applied"),
        "the role entries make the pick a change even though the flat pair is the same: {result}"
    );
    assert_eq!(
        result["replaced"],
        json!([
            "launcher.models.claude.leader",
            "launcher.models.claude.agent"
        ])
    );
    let after: Value = serde_json::from_str(&fs::read_to_string(&path).unwrap()).unwrap();
    assert_eq!(
        after["launcher"]["models"]["claude"], flat,
        "the role entries are gone and the flat pair is as it was"
    );
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
        let (result, out) = run_stage(&h.path, &enters(3));
        assert_eq!(result["state"], json!("declined"), "{codex}: {result}");
        assert!(
            !out.contains("launcher.models.codex") && !out.contains("SECRET-CODEX-MODEL"),
            "{codex}: the saved Codex value is not listed or quoted: {out}"
        );
        assert_eq!(fs::read_to_string(&path).unwrap(), body, "nothing written");
        let (result, _) = run_stage(&h.path, &format!("{}y\n", enters(3)));
        assert_eq!(result["state"], json!("applied"), "{codex}: {result}");
        let saved: Value = serde_json::from_str(&fs::read_to_string(&path).unwrap()).unwrap();
        assert!(
            saved["launcher"]["models"].get("codex").is_none(),
            "{codex}: the Codex value stayed: {saved}"
        );
        assert!(saved["launcher"]["models"]["claude"].is_object(), "{saved}");
    }
}

/// SPEC-20261122-remove-omp-launcher AC-7: a saved OMP value of any shape is neither listed nor
/// quoted by the models step, and the next save drops it.
#[test]
fn a_saved_omp_value_is_not_reported_and_leaves_the_file_on_the_next_save() {
    for omp in [
        r#"{"agent":{"model":"zai/glm-5.3","effort":"low"}}"#,
        r#"{"agent":"SECRET-OMP-MODEL"}"#,
        r#""SECRET-OMP-MODEL""#,
    ] {
        let h = home();
        fs::create_dir_all(h.path.join(".claude")).unwrap();
        fs::create_dir_all(h.path.join(".omp")).unwrap();
        let body = format!(r#"{{"launcher":{{"models":{{"omp":{omp}}}}}}}"#);
        let path = write_settings(&h.path, &body);
        let (result, out) = run_stage(&h.path, &enters(3));
        assert_eq!(result["state"], json!("declined"), "{omp}: {result}");
        assert!(
            !out.contains("launcher.models.omp") && !out.contains("SECRET-OMP-MODEL"),
            "{omp}: the saved OMP value is not listed or quoted: {out}"
        );
        assert!(
            !out.contains("OMP"),
            "{omp}: the models step offers OMP: {out}"
        );
        assert_eq!(fs::read_to_string(&path).unwrap(), body, "nothing written");
        let (result, _) = run_stage(&h.path, &format!("{}y\n", enters(3)));
        assert_eq!(result["state"], json!("applied"), "{omp}: {result}");
        let saved: Value = serde_json::from_str(&fs::read_to_string(&path).unwrap()).unwrap();
        assert!(
            saved["launcher"]["models"].get("omp").is_none(),
            "{omp}: the OMP value stayed: {saved}"
        );
        assert!(saved["launcher"]["models"]["claude"].is_object(), "{saved}");
    }
}

#[test]
fn a_no_and_an_unchanged_pick_write_nothing() {
    let h = home();
    fs::create_dir_all(h.path.join(".claude")).unwrap();
    let (result, _) = run_stage(&h.path, &format!("{}n\n", enters(3)));
    assert_eq!(result["state"], json!("declined"), "{result}");
    assert!(!settings_path(&h.path).exists());

    let (result, _) = run_stage(&h.path, &format!("{}y\n", enters(3)));
    assert_eq!(result["state"], json!("applied"), "{result}");
    let path = settings_path(&h.path);
    let before = fs::read(&path).unwrap();
    // The same picks again: nothing changes, nothing is asked to save, the bytes stay.
    let (result, _) = run_stage(&h.path, &format!("{}y\n", enters(3)));
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
    let (result, out) = run_stage(&h.path, &format!("{}y\n", enters(3)));
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
    let mut lines: Vec<String> = (0..3).map(|_| "\n".to_string()).collect();
    lines.push("y\n".to_string());
    let hooked = Hooked {
        lines,
        next: 0,
        hook_at: 3,
        hook: Some(Box::new(move || fs::write(&target, changed).unwrap())),
    };
    let mut out = Vec::new();
    let result = stage_with(
        &h.path,
        &mut BufReader::with_capacity(4096, hooked),
        &mut out,
        false,
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

/// The environment of the account that runs the host: only the Claude Code variable, as a map.
fn config_dir_env(value: Option<&Path>) -> impl Fn(&str) -> Option<std::ffi::OsString> {
    let value = value.map(|p| p.as_os_str().to_os_string());
    move |name: &str| {
        if name == "CLAUDE_CONFIG_DIR" {
            value.clone()
        } else {
            None
        }
    }
}

// SPEC-20261126-install-docs-match-code (AC-5, REQ-9): the tool search follows the folder the
// host home resolver gives, `CLAUDE_CONFIG_DIR` first.
#[test]
fn the_installed_tools_follow_the_claude_config_dir() {
    let h = home();
    let moved = h.path.join("moved-claude");
    fs::create_dir_all(&moved).unwrap();
    // The variable relocates the folder: no `~/.claude`, yet Claude Code is installed there.
    assert_eq!(
        installed_tools_with(&h.path, Some(&h.path), config_dir_env(Some(&moved))),
        ["claude"],
        "the folder CLAUDE_CONFIG_DIR names counts as the Claude Code folder"
    );
    // The variable unset and no `~/.claude`: nothing is installed.
    assert!(
        installed_tools_with(&h.path, Some(&h.path), config_dir_env(None)).is_empty(),
        "no variable and no `.claude` folder finds nothing"
    );
    // Only `~/.claude`, the variable unset: the default folder.
    fs::remove_dir_all(&moved).unwrap();
    fs::create_dir_all(h.path.join(".claude")).unwrap();
    assert_eq!(
        installed_tools_with(&h.path, Some(&h.path), config_dir_env(None)),
        ["claude"],
        "with the variable unset the HOME's `.claude` folder counts"
    );
}

#[test]
fn the_installed_tools_fall_back_to_the_home_folder_when_the_variable_is_refused() {
    let h = home();
    fs::create_dir_all(h.path.join(".claude")).unwrap();
    // A relative value and an empty value are the two the resolver refuses.
    for refused in ["moved-claude", ""] {
        assert_eq!(
            installed_tools_with(
                &h.path,
                Some(&h.path),
                config_dir_env(Some(Path::new(refused)))
            ),
            ["claude"],
            "a refused variable ({refused:?}) leaves the HOME's `.claude` folder as the answer"
        );
    }
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
    let body = r#"{"launcher":{"models":{"claude":{"model":"claude-sonnet-5-5","effort":"max"},"glm":{"leader":{"model":"SECRET-MODEL","effort":"ultra"}}}}}"#;
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
    assert_eq!(values.len(), 2, "one value per host: {models}");
    assert!(
        values
            .iter()
            .all(|v| v["tool"] != json!("codex") && v["tool"] != json!("omp")),
        "no Codex or OMP row in the report: {models}"
    );
    let find = |tool: &str| -> Value {
        values
            .iter()
            .find(|v| v["tool"] == json!(tool))
            .unwrap()
            .clone()
    };
    assert_eq!(
        find("claude"),
        json!({"tool": "claude", "model": "claude-sonnet-5-5", "effort": "max", "source": "saved"})
    );
    // Nothing saved for glm: the report shows what a launch starts with (the alias and the
    // host effort), not a role row.
    assert_eq!(
        find("glm"),
        json!({"tool": "glm", "model": "opus", "effort": "high", "source": "default"})
    );
    assert_eq!(models["unsupported"], json!(["launcher.models.glm.leader"]));
    assert_eq!(
        models["role_entries"],
        json!(["launcher.models.glm.leader"])
    );
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
    let body = r#"{"launcher":{"models":{"claude":{"model":"claude-sonnet-5-5","effort":"max"}}}}"#;
    let path = write_settings(&h.path, body);
    // Enter at every question: the saved entry stays, so there is nothing to save and the bytes
    // stay.
    let (result, out) = run_stage(&h.path, &format!("{}y\n", enters(3)));
    assert_eq!(result["state"], json!("unchanged"), "{result}");
    assert!(
        out.contains("claude-sonnet-5-5"),
        "the saved model is offered first: {out}"
    );
    assert_eq!(fs::read_to_string(&path).unwrap(), body, "nothing written");
    // A different effort is a change: the preview shows saved -> new.
    let (result, out) = run_stage(&h.path, "\n\n2\ny\n");
    assert_eq!(result["state"], json!("applied"), "{result}");
    assert!(
        out.contains("claude: claude-sonnet-5-5 max -> claude-sonnet-5-5"),
        "the preview shows the saved entry replaced: {out}"
    );
    let saved: Value = serde_json::from_str(&fs::read_to_string(&path).unwrap()).unwrap();
    assert_eq!(
        saved["launcher"]["models"]["claude"]["model"],
        json!("claude-sonnet-5-5")
    );
    assert_ne!(
        saved["launcher"]["models"]["claude"]["effort"],
        json!("max")
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
    let plan = launcher_dry_run(&h.path, &["cc"]);
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
