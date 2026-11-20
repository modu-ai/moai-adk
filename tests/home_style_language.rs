//! SPEC-20261120-reply-language-rule: the user layer's Claude Code settings also get the
//! `language` key when it is missing, from the conversation language the installer saved.
//! Every case runs against a temporary HOME; the real one is never read or written.

mod common;

use moai_cli::home_style::{apply, plan, revert, Plan, STYLE_NAME};
use serde_json::{json, Value};
use std::fs;
use std::path::{Path, PathBuf};

const PERM: &str = "test-permission";
const SETTINGS: &str = ".claude/settings.json";

fn package() -> PathBuf {
    Path::new(env!("CARGO_MANIFEST_DIR")).join("skills/moai")
}

fn temp_home() -> (tempfile::TempDir, PathBuf) {
    let dir = tempfile::tempdir().unwrap();
    let path = fs::canonicalize(dir.path()).unwrap();
    (dir, path)
}

fn target<'a>(p: &'a Plan, kind: &str) -> &'a moai_cli::home_style::TargetPlan {
    p.targets.iter().find(|t| t.kind == kind).unwrap()
}

fn settings(home: &Path) -> Value {
    serde_json::from_slice(&fs::read(home.join(SETTINGS)).unwrap()).unwrap()
}

fn write(home: &Path, rel: &str, bytes: &str) {
    let p = home.join(rel);
    fs::create_dir_all(p.parent().unwrap()).unwrap();
    fs::write(p, bytes).unwrap();
}

fn saved_language(home: &Path, code: &str) {
    // The installer's state folder exists before it records the language.
    fs::create_dir_all(moai_cli::home_root::path(home)).unwrap();
    let r = moai_cli::language::set_home(home, code).unwrap();
    assert_eq!(r["written"], json!(true), "{r}");
}

/// REQ-1, AC-1: Korean saved at install time, no settings file: the plan sets both keys and the
/// apply writes them.
#[test]
fn the_plan_sets_language_to_the_saved_conversation_language() {
    let (_t, home) = temp_home();
    saved_language(&home, "ko");
    let p = plan(&home, &package()).unwrap();
    let t = target(&p, "setting");
    assert_eq!(t.action, "set");
    assert!(
        t.detail.as_deref().unwrap_or("").contains("language"),
        "{:?}",
        t.detail
    );
    let r = apply(&home, &package(), &p.plan_digest, PERM).unwrap();
    assert_eq!(r["state"], "applied");
    assert_eq!(
        settings(&home),
        json!({"outputStyle": STYLE_NAME, "language": "korean"})
    );
    // Planned again, nothing is left to write.
    let again = plan(&home, &package()).unwrap();
    assert_eq!(target(&again, "setting").action, "unchanged");
}

/// REQ-2, AC-2: a `language` the user wrote stays, whatever it holds, and the plan says so.
#[test]
fn a_user_set_language_is_kept_and_said_so() {
    for value in ["\"japanese\"", "\"\""] {
        let (_t, home) = temp_home();
        saved_language(&home, "ko");
        let before = format!("{{\"language\": {value}}}\n");
        write(&home, SETTINGS, &before);
        let p = plan(&home, &package()).unwrap();
        let t = target(&p, "setting");
        // The output style key is still missing, so the file is written for it alone.
        assert_eq!(t.action, "set", "{value}");
        let d = t.detail.as_deref().unwrap_or("");
        assert!(d.contains("language") && d.contains("kept"), "{value}: {d}");
        apply(&home, &package(), &p.plan_digest, PERM).unwrap();
        let s = settings(&home);
        let want: Value = serde_json::from_str(value).unwrap();
        assert_eq!(s["language"], want, "{value}");
        assert_eq!(s["outputStyle"], json!(STYLE_NAME));
    }
}

/// REQ-3, AC-3: nothing saved, or a language MoAI has no name for: no `language` key is planned
/// and the detail says why. The old plan stays: the output style alone.
#[test]
fn without_a_known_conversation_language_no_key_is_planned_and_the_reason_is_named() {
    // Nothing saved.
    let (_t, home) = temp_home();
    let p = plan(&home, &package()).unwrap();
    let t = target(&p, "setting");
    assert_eq!(t.action, "set");
    let d = t.detail.as_deref().unwrap_or("");
    assert!(
        d.contains("language") && d.contains("no conversation language"),
        "{d}"
    );
    apply(&home, &package(), &p.plan_digest, PERM).unwrap();
    assert_eq!(settings(&home), json!({"outputStyle": STYLE_NAME}));

    // A well-formed code MoAI has no name for.
    let (_t, home) = temp_home();
    saved_language(&home, "fr");
    let p = plan(&home, &package()).unwrap();
    let d = target(&p, "setting").detail.clone().unwrap_or_default();
    assert!(d.contains("language") && d.contains("fr"), "{d}");
    apply(&home, &package(), &p.plan_digest, PERM).unwrap();
    assert_eq!(settings(&home), json!({"outputStyle": STYLE_NAME}));
}

/// REQ-4, REQ-5, REQ-9, AC-4: one apply writes both keys as one replacement with one receipt;
/// the revert removes those keys and nothing else, and leaves one changed since.
#[test]
fn one_apply_writes_both_keys_and_revert_removes_only_those() {
    let (_t, home) = temp_home();
    saved_language(&home, "ko");
    let before = "{\n  \"z\": 1.50,\n  \"hooks\": {\"Stop\": []}\n}\n";
    write(&home, SETTINGS, before);
    let p = plan(&home, &package()).unwrap();
    let r = apply(&home, &package(), &p.plan_digest, PERM).unwrap();
    let written: Vec<&str> = r["written"]
        .as_array()
        .unwrap()
        .iter()
        .map(|v| v.as_str().unwrap())
        .collect();
    assert_eq!(
        written.iter().filter(|w| **w == "claude:setting").count(),
        1,
        "{r}"
    );
    let s = settings(&home);
    assert_eq!(s["language"], json!("korean"));
    assert_eq!(s["outputStyle"], json!(STYLE_NAME));
    let receipt = r["receipt"].as_str().unwrap().to_string();

    // The user edits the file after the apply, then the revert runs: the keys MoAI wrote are
    // gone, every other byte the user has stays.
    let text = fs::read_to_string(home.join(SETTINGS)).unwrap();
    let edited = text.replacen("\"z\": 1.50", "\"z\": 2.50", 1);
    fs::write(home.join(SETTINGS), &edited).unwrap();
    let out = revert(&home, &receipt, PERM).unwrap();
    assert!(
        out["reverted"]
            .as_array()
            .unwrap()
            .iter()
            .any(|v| v == "claude:setting"),
        "{out}"
    );
    let after = settings(&home);
    assert_eq!(after, json!({"z": 2.50, "hooks": {"Stop": []}}));

    // A second round: the user changes the `language` MoAI wrote; that key stays and is reported,
    // the other key goes.
    let p = plan(&home, &package()).unwrap();
    let r = apply(&home, &package(), &p.plan_digest, PERM).unwrap();
    let receipt = r["receipt"].as_str().unwrap().to_string();
    let text = fs::read_to_string(home.join(SETTINGS)).unwrap();
    fs::write(
        home.join(SETTINGS),
        text.replacen("\"korean\"", "\"spanish\"", 1),
    )
    .unwrap();
    let out = revert(&home, &receipt, PERM).unwrap();
    let skipped = out["skipped"].to_string();
    assert!(
        skipped.contains("language") && skipped.contains("user_changed_since_apply"),
        "{out}"
    );
    let after = settings(&home);
    assert_eq!(after["language"], json!("spanish"));
    assert!(after.get("outputStyle").is_none(), "{after}");
}

/// A receipt written before this change (one `outputStyle` key, no key list) still reverts.
#[test]
fn a_receipt_from_before_the_language_key_still_reverts() {
    let (_t, home) = temp_home();
    let p = plan(&home, &package()).unwrap();
    let r = apply(&home, &package(), &p.plan_digest, PERM).unwrap();
    assert_eq!(settings(&home), json!({"outputStyle": STYLE_NAME}));
    let receipt = r["receipt"].as_str().unwrap().to_string();
    // The user adds a key after the apply, so the revert has to cut the text, not restore bytes.
    let text = fs::read_to_string(home.join(SETTINGS)).unwrap();
    fs::write(
        home.join(SETTINGS),
        text.replacen("}\n", ",\n  \"mine\": true\n}\n", 1),
    )
    .unwrap();
    // The receipt as the old code wrote it: no `inserted_keys` field.
    let path = moai_cli::home_root::path(&home)
        .join("home-style/receipts")
        .join(&receipt)
        .join("receipt.json");
    let mut v: Value = serde_json::from_slice(&fs::read(&path).unwrap()).unwrap();
    for e in v["entries"].as_array_mut().unwrap() {
        e.as_object_mut().unwrap().remove("inserted_keys");
    }
    fs::write(&path, serde_json::to_vec_pretty(&v).unwrap()).unwrap();
    let out = revert(&home, &receipt, PERM).unwrap();
    assert_eq!(out["skipped"], json!([]), "{out}");
    assert_eq!(settings(&home), json!({"mine": true}));
}

/// REQ-6, AC-5: a relocated Claude Code config folder (the account's own) takes the keys; the
/// default folder under HOME stays untouched.
#[test]
#[cfg(unix)]
fn a_config_dir_profile_gets_its_own_settings() {
    use common::home::TempHome;
    let th = TempHome::new();
    let base = fs::canonicalize(th.path()).unwrap();
    saved_language(&base, "ko");
    let conf = base.join("elsewhere/claude-conf");
    let run = |args: &[&str]| {
        let mut cmd = th.command(args);
        cmd.env("HOME", &base)
            .env("MOAI_ACCOUNT_HOME", &base)
            .env("CLAUDE_CONFIG_DIR", &conf);
        let out = cmd.output().expect("run moai");
        let v: Value = serde_json::from_slice(&out.stdout)
            .unwrap_or_else(|e| panic!("not JSON ({e}): {}", String::from_utf8_lossy(&out.stdout)));
        assert_eq!(v["ok"], json!(true), "{v}");
        v["result"].clone()
    };
    let h = base.to_str().unwrap();
    let pkg = package();
    let plan = run(&["home-style-plan", h, pkg.to_str().unwrap()]);
    run(&[
        "home-style-apply",
        h,
        pkg.to_str().unwrap(),
        "--plan-digest",
        plan["plan_digest"].as_str().unwrap(),
        "--permission-reference",
        PERM,
    ]);
    let s: Value = serde_json::from_slice(&fs::read(conf.join("settings.json")).unwrap()).unwrap();
    assert_eq!(s["language"], json!("korean"), "{s}");
    assert!(!base.join(".claude").exists());
}

/// REQ-8, AC-7: with a known conversation language the Codex block opens with one reply
/// sentence; without one the block is the old bytes.
#[test]
fn the_codex_block_carries_the_reply_sentence_when_the_language_is_known() {
    use moai_cli::home_style::{render_block, SOURCE};
    let source = fs::read_to_string(package().join(SOURCE)).unwrap();

    let (_t, home) = temp_home();
    saved_language(&home, "ko");
    let p = plan(&home, &package()).unwrap();
    apply(&home, &package(), &p.plan_digest, PERM).unwrap();
    let block = fs::read_to_string(home.join(".codex/AGENTS.md")).unwrap();
    let line = "Reply to the user, write reports and ask questions in Korean.";
    assert_eq!(block, render_block(&format!("{line}\n\n{source}")));
    assert_eq!(block.matches(line).count(), 1);
    assert_eq!(
        target(&plan(&home, &package()).unwrap(), "block").action,
        "unchanged"
    );

    // No language saved, English, or a code without a name: the bytes of the old block.
    for code in [None, Some("en"), Some("fr")] {
        let (_t, home) = temp_home();
        if let Some(c) = code {
            saved_language(&home, c);
        }
        let p = plan(&home, &package()).unwrap();
        apply(&home, &package(), &p.plan_digest, PERM).unwrap();
        let block = fs::read_to_string(home.join(".codex/AGENTS.md")).unwrap();
        assert_eq!(block, render_block(&source), "{code:?}");
    }
}
