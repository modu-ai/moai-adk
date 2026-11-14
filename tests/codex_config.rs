//! `.codex/config.toml` plan/apply (Issue #2240, REQ-I3-01): user keys, tables and comments
//! survive; only managed keys change, three-way against the ownership receipt. Every limit
//! and branch here has a test that fails when it moves.
use moai_cli::hosts::codex_config::{config_apply, config_plan};
use serde_json::{json, Map, Value};
use std::fs;

/// The bounds as numbers, not the crate's constants, so a moved constant fails a test.
const CONFIG_LIMIT: u64 = 262_144;
const MAX_DOC_BYTES: i64 = 1_048_576;

const USER_CONFIG: &str = r#"# my codex settings
model = "gpt-6"  # keep me

[tui]
# footer
theme = "dark"

[mcp_servers.docs]
command = "docs-mcp"
args = ["--port", "9000"]
"#;

struct Fixture {
    _tmp: tempfile::TempDir,
    project: std::path::PathBuf,
    home: std::path::PathBuf,
}

fn fixture(config: Option<&str>) -> Fixture {
    let tmp = tempfile::tempdir().unwrap();
    let project = tmp.path().join("project");
    let home = tmp.path().join("home");
    fs::create_dir_all(&project).unwrap();
    fs::create_dir_all(&home).unwrap();
    if let Some(text) = config {
        fs::create_dir_all(project.join(".codex")).unwrap();
        fs::write(project.join(".codex/config.toml"), text).unwrap();
    }
    Fixture {
        _tmp: tmp,
        project,
        home,
    }
}

fn desired(v: Value) -> Map<String, Value> {
    v.as_object().unwrap().clone()
}

fn read(f: &Fixture) -> String {
    fs::read_to_string(f.project.join(".codex/config.toml")).unwrap()
}

/// Plans and applies; returns the plan.
fn apply(f: &Fixture, want: &Map<String, Value>) -> Value {
    let plan = config_plan(&f.project, &f.home, want).unwrap();
    let digest = plan["plan_digest"].as_str().unwrap().to_string();
    config_apply(&f.project, &f.home, want, &digest, "test-permission").unwrap();
    plan
}

fn all_managed() -> Map<String, Value> {
    desired(json!({
        "tui.status_line": ["model", "git-branch"],
        "skills.bundled.enabled": false,
        "project_doc_fallback_filenames": ["TEAM_GUIDE.md"],
        "project_doc_max_bytes": 65536
    }))
}

#[test]
fn round_trip_keeps_every_user_key_comment_and_table() {
    let f = fixture(Some(USER_CONFIG));
    let want = all_managed();
    let plan = apply(&f, &want);
    assert_eq!(plan["action"], "update");
    assert_eq!(plan["writes"], false);
    let text = read(&f);
    // Every user line is still there, in order.
    let mut at = 0;
    for line in USER_CONFIG.lines().filter(|l| !l.is_empty()) {
        let found = text[at..]
            .find(line)
            .unwrap_or_else(|| panic!("lost {line:?}:\n{text}"));
        at += found + line.len();
    }
    let doc: toml_edit::DocumentMut = text.parse().unwrap();
    assert_eq!(doc["model"].as_str(), Some("gpt-6"));
    assert_eq!(doc["tui"]["theme"].as_str(), Some("dark"));
    assert_eq!(doc["tui"]["status_line"].as_array().unwrap().len(), 2);
    assert_eq!(doc["skills"]["bundled"]["enabled"].as_bool(), Some(false));
    assert_eq!(doc["project_doc_max_bytes"].as_integer(), Some(65536));
    assert_eq!(
        doc["mcp_servers"]["docs"]["command"].as_str(),
        Some("docs-mcp")
    );
    // Root keys stay above the first table; only the holding table gets a header.
    assert!(text.find("project_doc_max_bytes").unwrap() < text.find("[tui]").unwrap());
    assert!(text.contains("[skills.bundled]") && !text.contains("[skills]\n"));

    // Idempotent: a second plan is unchanged and applying leaves the bytes alone.
    let again = config_plan(&f.project, &f.home, &want).unwrap();
    assert_eq!(again["action"], "unchanged");
    let before = read(&f);
    apply(&f, &want);
    assert_eq!(read(&f), before);
}

#[test]
fn creates_the_file_when_missing_and_never_writes_trust() {
    let f = fixture(None);
    let plan = apply(&f, &all_managed());
    assert_eq!(plan["action"], "create");
    assert_eq!(plan["trust"]["writes"], false);
    // #2928 S-5: both items are Codex 0.169.0 identifiers, so nothing is unverified.
    assert!(plan["status_line_unverified"]["reason"].is_null());
    let text = read(&f);
    assert!(!text.contains("trust"), "{text}");
    assert!(!text.contains("projects"), "{text}");
}

#[test]
fn a_managed_key_the_user_changed_is_a_conflict_and_nothing_is_applied() {
    let f = fixture(Some(USER_CONFIG));
    apply(&f, &all_managed());
    let edited = read(&f).replace("65536", "1000");
    fs::write(f.project.join(".codex/config.toml"), &edited).unwrap();
    let mut want = all_managed();
    want.insert("project_doc_max_bytes".into(), json!(131072));
    let plan = config_plan(&f.project, &f.home, &want).unwrap();
    assert_eq!(plan["action"], "conflict");
    assert_eq!(plan["conflicts"], json!(["project_doc_max_bytes"]));
    let err = config_apply(
        &f.project,
        &f.home,
        &want,
        plan["plan_digest"].as_str().unwrap(),
        "p",
    )
    .unwrap_err();
    assert!(err.starts_with("conflict in .codex/config.toml"), "{err}");
    assert_eq!(read(&f), edited);
}

#[test]
fn a_user_edit_beside_a_managed_key_is_kept() {
    let f = fixture(Some(USER_CONFIG));
    apply(&f, &all_managed());
    let edited = read(&f).replace("theme = \"dark\"", "theme = \"light\"");
    fs::write(f.project.join(".codex/config.toml"), &edited).unwrap();
    let mut want = all_managed();
    want.insert("tui.status_line".into(), json!(["model"]));
    let plan = apply(&f, &want);
    assert_eq!(plan["action"], "update");
    let text = read(&f);
    assert!(text.contains("theme = \"light\""), "{text}");
    assert!(text.contains("status_line = [\"model\"]"), "{text}");
}

#[test]
fn dropping_a_key_removes_it_only_when_moai_owned_it() {
    // The user already had max_bytes = 65536: MoAI does not own an equal value it found.
    let f = fixture(Some("project_doc_max_bytes = 65536\n"));
    apply(&f, &all_managed());
    let receipt: Value = serde_json::from_slice(
        &fs::read(f.project.join(".moai-cli/manifests/codex-config.json")).unwrap(),
    )
    .unwrap();
    assert!(receipt["owned"].get("project_doc_max_bytes").is_none());
    assert!(receipt["owned"].get("tui.status_line").is_some());
    apply(&f, &Map::new());
    let text = read(&f);
    assert!(text.contains("project_doc_max_bytes = 65536"), "{text}");
    assert!(!text.contains("status_line"), "{text}");
    assert!(!text.contains("enabled"), "{text}");
}

#[test]
fn refuses_trust_ignored_unknown_and_out_of_range_values() {
    let f = fixture(None);
    let refused = [
        (
            json!({"projects./x.trust_level": "trusted"}),
            "user config; refused",
        ),
        (
            json!({"model_provider": "x"}),
            "codex ignores model_provider",
        ),
        (json!({"notify": ["x"]}), "codex ignores notify"),
        (json!({"model": "x"}), "unmanaged codex config key refused"),
        (
            json!({"project_doc_max_bytes": 0}),
            "invalid codex config value",
        ),
        (
            json!({"project_doc_max_bytes": -1}),
            "invalid codex config value",
        ),
        (
            json!({"project_doc_max_bytes": MAX_DOC_BYTES + 1}),
            "invalid codex config value",
        ),
        (
            json!({"project_doc_max_bytes": "65536"}),
            "invalid codex config value",
        ),
        (
            json!({"skills.bundled.enabled": "false"}),
            "invalid codex config value",
        ),
        (
            json!({"tui.status_line": Value::Null}),
            "invalid codex config value",
        ),
        (
            json!({"tui.status_line": vec!["a"; 33]}),
            "invalid codex config value",
        ),
        (
            json!({"tui.status_line": [""]}),
            "invalid codex config value",
        ),
        (
            json!({"tui.status_line": ["x".repeat(129)]}),
            "invalid codex config value",
        ),
        (
            json!({"project_doc_fallback_filenames": ["docs/A.md"]}),
            "invalid codex config value",
        ),
        (
            json!({"project_doc_fallback_filenames": [".."]}),
            "invalid codex config value",
        ),
    ];
    for (want, phrase) in refused {
        let err = config_plan(&f.project, &f.home, &desired(want.clone())).unwrap_err();
        assert!(err.contains(phrase), "{want}: {err}");
    }
    // The bounds themselves are accepted.
    let edge = desired(json!({
        "project_doc_max_bytes": MAX_DOC_BYTES,
        "tui.status_line": vec!["a"; 32],
        "project_doc_fallback_filenames": ["y".repeat(128)]
    }));
    assert_eq!(
        config_plan(&f.project, &f.home, &edge).unwrap()["action"],
        "create"
    );
    let one = desired(json!({"project_doc_max_bytes": 1}));
    assert_eq!(
        config_plan(&f.project, &f.home, &one).unwrap()["action"],
        "create"
    );
    assert!(!f.project.join(".codex").exists(), "plan wrote nothing");
}

#[test]
fn refuses_the_home_directory_as_the_project() {
    let f = fixture(None);
    let err = config_plan(&f.home, &f.home, &all_managed()).unwrap_err();
    assert!(err.contains("home directory"), "{err}");
    assert!(!f.home.join(".codex").exists());
}

#[test]
fn the_size_limit_holds_for_the_file_and_the_result() {
    let padded = |n: u64| {
        let mut s = String::from("# ");
        s.push_str(&"x".repeat(n as usize - 3));
        s.push('\n');
        s
    };
    // Exactly at the limit: readable, and unchanged when nothing is asked.
    let f = fixture(Some(&padded(CONFIG_LIMIT)));
    assert_eq!(
        config_plan(&f.project, &f.home, &Map::new()).unwrap()["action"],
        "unchanged"
    );
    // A change that grows it past the limit is refused.
    let err = config_plan(&f.project, &f.home, &all_managed()).unwrap_err();
    assert!(err.contains("exceeds the 256 KiB limit"), "{err}");
    // One byte over: not read at all.
    let g = fixture(Some(&padded(CONFIG_LIMIT + 1)));
    let err = config_plan(&g.project, &g.home, &Map::new()).unwrap_err();
    assert!(err.contains("not a bounded regular file"), "{err}");
}

#[test]
fn updating_a_managed_value_keeps_the_comments_around_it() {
    let f = fixture(None);
    apply(&f, &desired(json!({"project_doc_max_bytes": 65536})));
    let commented = read(&f).replace(
        "project_doc_max_bytes = 65536",
        "# sized for our docs\nproject_doc_max_bytes = 65536  # tuned",
    );
    fs::write(f.project.join(".codex/config.toml"), &commented).unwrap();
    apply(&f, &desired(json!({"project_doc_max_bytes": 131072})));
    let text = read(&f);
    assert!(
        text.contains("# sized for our docs\nproject_doc_max_bytes = 131072  # tuned"),
        "{text}"
    );
}

#[test]
fn malformed_or_unexpected_shapes_need_review() {
    for text in ["tui = 1\n", "[[tui]]\nx = 1\n", "not toml = = \n"] {
        let f = fixture(Some(text));
        let err = config_plan(&f.project, &f.home, &all_managed()).unwrap_err();
        assert!(
            err.contains("requires preservation review"),
            "{text:?}: {err}"
        );
    }
    // A managed key holding a shape MoAI never writes is the user's value: a conflict.
    let f = fixture(Some("tui.status_line = { a = 1 }\n"));
    let plan = config_plan(&f.project, &f.home, &all_managed()).unwrap();
    assert_eq!(plan["conflicts"], json!(["tui.status_line"]));
    // An inline parent table is a table: the key goes inside it, the rest stays.
    let f = fixture(Some("tui = { theme = \"dark\" }\n"));
    apply(&f, &desired(json!({"tui.status_line": ["model"]})));
    let doc: toml_edit::DocumentMut = read(&f).parse().unwrap();
    assert_eq!(doc["tui"]["theme"].as_str(), Some("dark"));
    assert_eq!(doc["tui"]["status_line"][0].as_str(), Some("model"));
}

#[test]
fn stale_digest_and_missing_permission_are_refused() {
    let f = fixture(Some(USER_CONFIG));
    let want = all_managed();
    let plan = config_plan(&f.project, &f.home, &want).unwrap();
    let digest = plan["plan_digest"].as_str().unwrap();
    let err = config_apply(&f.project, &f.home, &want, digest, "").unwrap_err();
    assert_eq!(err, "permission reference required");
    fs::write(
        f.project.join(".codex/config.toml"),
        format!("{USER_CONFIG}x = 1\n"),
    )
    .unwrap();
    let err = config_apply(&f.project, &f.home, &want, digest, "p").unwrap_err();
    assert!(err.starts_with("stale or altered plan"), "{err}");
}

#[test]
fn a_pending_receipt_resolves_by_the_side_the_file_is_on() {
    let f = fixture(Some(USER_CONFIG));
    let want = all_managed();
    apply(&f, &want);
    let receipt_path = f.project.join(".moai-cli/manifests/codex-config.json");
    let config_path = f.project.join(".codex/config.toml");
    let applied = fs::read(&config_path).unwrap();
    let sha = |b: &[u8]| moai_cli::digest(b);
    let write_receipt = |config_sha: String, pending_sha: String| {
        let r = json!({"schema_version": 1, "config_path": ".codex/config.toml",
            "owned": {}, "config_sha256": config_sha,
            "pending": {"owned": want, "config_sha256": pending_sha}});
        fs::write(&receipt_path, serde_json::to_vec(&r).unwrap()).unwrap();
    };
    // The config write landed: the intended ownership holds, so the plan is unchanged.
    write_receipt(sha(USER_CONFIG.as_bytes()), sha(&applied));
    assert_eq!(
        config_plan(&f.project, &f.home, &want).unwrap()["action"],
        "unchanged"
    );
    // It did not land: the recorded (empty) ownership holds and the plan writes again.
    fs::write(&config_path, USER_CONFIG).unwrap();
    assert_eq!(
        config_plan(&f.project, &f.home, &want).unwrap()["action"],
        "update"
    );
    // Neither side: nobody recorded this change.
    fs::write(&config_path, "x = 1\n").unwrap();
    let err = config_plan(&f.project, &f.home, &want).unwrap_err();
    assert!(
        err.contains("receipt requires preservation review"),
        "{err}"
    );
}

#[cfg(unix)]
#[test]
fn refuses_a_symlinked_codex_directory() {
    let f = fixture(None);
    std::os::unix::fs::symlink(f.home.as_path(), f.project.join(".codex")).unwrap();
    let err = config_plan(&f.project, &f.home, &all_managed()).unwrap_err();
    assert!(err.contains("symlink"), "{err}");
}

// ---- #2928 S-5: Codex status line in Claude Code order -------------------------------------

/// The `moai` preset, literal: the Claude Code status-line elements Codex 0.169.0 can show,
/// in Claude Code order (Issue #2928 comment 6877416640).
const MOAI_PRESET: [&str; 10] = [
    "model",
    "reasoning",
    "codex-version",
    "project-name",
    "pull-request-number",
    "git-branch",
    "branch-changes",
    "thread-name",
    "context-remaining",
    "weekly-limit",
];

/// Every identifier Codex 0.169.0 parses, aliases included, literal (openai/codex
/// rust-v0.169.0 a41fe5b443652cbdbee57717fbb9fd78d8d8b32a,
/// codex-rs/tui/src/bottom_pane/status_line_setup.rs:57-169).
const CODEX_0169_IDS: [&str; 41] = [
    "model",
    "model-name",
    "model-with-reasoning",
    "reasoning",
    "current-dir",
    "project-name",
    "project",
    "project-root",
    "hostname",
    "git-branch",
    "pull-request-number",
    "branch-changes",
    "run-state",
    "status",
    "permissions",
    "approval-mode",
    "approval",
    "context-remaining",
    "context-used",
    "context-usage",
    "five-hour-limit",
    "weekly-limit",
    "codex-version",
    "context-window-size",
    "used-tokens",
    "total-input-tokens",
    "total-output-tokens",
    "thread-credits",
    "estimated-thread-cost",
    "thread-id",
    "session-id",
    "fast-mode",
    "daybreak",
    "raw-output",
    "thread-name",
    "thread-title",
    "workspace-headline",
    "task-progress",
    "profile",
    "provider",
    "model-provider",
];

const USER_STATUS: &str = r#"# my codex settings
model = "gpt-6"  # keep me

[tui]
# footer
theme = "dark"
status_line = ["model", "current-dir"]  # mine

[mcp_servers.docs]
command = "docs-mcp"
"#;

fn preset_line() -> String {
    let quoted: Vec<String> = MOAI_PRESET.iter().map(|s| format!("\"{s}\"")).collect();
    format!("status_line = [{}]", quoted.join(", "))
}

/// Every path under `dir` with its bytes, for a before/after comparison of a whole tree.
fn tree(dir: &std::path::Path) -> Vec<(std::path::PathBuf, Vec<u8>)> {
    let mut out = Vec::new();
    let mut stack = vec![dir.to_path_buf()];
    while let Some(d) = stack.pop() {
        for entry in fs::read_dir(&d).unwrap() {
            let path = entry.unwrap().path();
            if path.is_dir() {
                out.push((path.clone(), Vec::new()));
                stack.push(path);
            } else {
                out.push((path.clone(), fs::read(&path).unwrap()));
            }
        }
    }
    out.sort();
    out
}

fn receipt(f: &Fixture) -> Value {
    serde_json::from_slice(
        &fs::read(f.project.join(".moai-cli/manifests/codex-config.json")).unwrap(),
    )
    .unwrap()
}

#[test]
fn moai_preset_resolves_to_the_claude_code_order() {
    let f = fixture(None);
    let want = desired(json!({"tui.status_line": "moai"}));
    let plan = apply(&f, &want);
    assert_eq!(plan["status_line"]["preset"], "moai");
    assert_eq!(plan["status_line"]["after"], json!(MOAI_PRESET));
    assert_eq!(plan["status_line"]["before"], Value::Null);
    let doc: toml_edit::DocumentMut = read(&f).parse().unwrap();
    let written: Vec<&str> = doc["tui"]["status_line"]
        .as_array()
        .unwrap()
        .iter()
        .map(|v| v.as_str().unwrap())
        .collect();
    assert_eq!(written, MOAI_PRESET);
    // The same request again is unchanged; an explicit list still works beside the preset.
    assert_eq!(
        config_plan(&f.project, &f.home, &want).unwrap()["action"],
        "unchanged"
    );
    let explicit = desired(json!({"tui.status_line": ["model", "git-branch"]}));
    let plan = apply(&f, &explicit);
    assert_eq!(plan["status_line"]["preset"], Value::Null);
    assert_eq!(plan["status_line"]["before"], json!(MOAI_PRESET));
    assert!(read(&f).contains("status_line = [\"model\", \"git-branch\"]"));
    // An unknown preset name is refused, and so is a near miss.
    for name in ["Moai", "moai ", "claude-code", ""] {
        let err = config_plan(
            &f.project,
            &f.home,
            &desired(json!({"tui.status_line": name})),
        )
        .unwrap_err();
        assert_eq!(
            err, "invalid codex config value for tui.status_line",
            "{name:?}"
        );
    }
}

#[test]
fn plan_with_the_preset_writes_nothing() {
    for (text, action, before) in [
        (USER_CONFIG, "update", Value::Null),
        (USER_STATUS, "conflict", json!(["model", "current-dir"])),
    ] {
        let f = fixture(Some(text));
        let root = f.project.parent().unwrap().to_path_buf();
        let tree_before = tree(&root);
        let want = desired(json!({"tui.status_line": "moai"}));
        let plan = config_plan(&f.project, &f.home, &want).unwrap();
        assert_eq!(plan["action"], action);
        assert_eq!(plan["writes"], false);
        assert_eq!(plan["status_line"]["before"], before);
        // A conflict writes nothing, so the list after is the list before.
        let after = if action == "conflict" {
            before
        } else {
            json!(MOAI_PRESET)
        };
        assert_eq!(plan["status_line"]["after"], after);
        assert_eq!(tree(&root), tree_before);
    }
}

#[test]
fn apply_changes_only_the_status_line_line() {
    let f = fixture(Some(USER_CONFIG));
    apply(&f, &desired(json!({"tui.status_line": "moai"})));
    let expected = USER_CONFIG.replace(
        "theme = \"dark\"\n",
        &format!("theme = \"dark\"\n{}\n", preset_line()),
    );
    assert_ne!(expected, USER_CONFIG);
    assert_eq!(read(&f), expected);
    assert_eq!(receipt(&f)["owned"]["tui.status_line"], json!(MOAI_PRESET));

    // The user's own list is theirs: a conflict, nothing written, the bytes kept.
    let g = fixture(Some(USER_STATUS));
    let want = desired(json!({"tui.status_line": "moai"}));
    let plan = config_plan(&g.project, &g.home, &want).unwrap();
    let digest = plan["plan_digest"].as_str().unwrap();
    let err = config_apply(&g.project, &g.home, &want, digest, "p").unwrap_err();
    assert!(err.starts_with("conflict in .codex/config.toml"), "{err}");
    assert_eq!(read(&g), USER_STATUS);
}

#[test]
fn dropping_the_preset_restores_the_original_bytes() {
    let g = fixture(Some(USER_CONFIG));
    apply(&g, &desired(json!({"tui.status_line": "moai"})));
    assert_ne!(read(&g), USER_CONFIG);
    // Through an explicit list and back, then out: the receipt-owned key is removed.
    apply(&g, &desired(json!({"tui.status_line": ["model"]})));
    let plan = apply(&g, &Map::new());
    assert_eq!(plan["status_line"]["before"], json!(["model"]));
    assert_eq!(plan["status_line"]["after"], Value::Null);
    assert_eq!(
        fs::read(g.project.join(".codex/config.toml")).unwrap(),
        USER_CONFIG.as_bytes()
    );
    assert_eq!(receipt(&g)["owned"], json!({}));
}

#[test]
fn known_ids_are_not_flagged_and_unknown_ids_are_reported() {
    let f = fixture(None);
    // Every 0.169.0 identifier and alias is known (two requests: the list bound is 32).
    for chunk in CODEX_0169_IDS.chunks(21) {
        let want = desired(json!({ "tui.status_line": chunk }));
        let plan = config_plan(&f.project, &f.home, &want).unwrap();
        assert_eq!(plan["status_line_unverified"]["items"], json!([]), "{plan}");
        assert_eq!(plan["status_line_unverified"]["reason"], Value::Null);
        let items = plan["status_line"]["items"].as_array().unwrap();
        assert_eq!(items.len(), chunk.len());
        for item in items {
            assert_eq!(item["status"], "known", "{item}");
        }
    }
    // The picker's colour toggle is not a status-line item, and near misses are unknown.
    let asked = [
        "model",
        "moai-version",
        "status-line-use-theme-colors",
        "Model",
        "git_branch",
    ];
    let plan = apply(&f, &desired(json!({ "tui.status_line": asked })));
    assert_eq!(
        plan["status_line"]["items"],
        json!([
            {"id": "model", "status": "known"},
            {"id": "moai-version", "status": "unknown_item"},
            {"id": "status-line-use-theme-colors", "status": "unknown_item"},
            {"id": "Model", "status": "unknown_item"},
            {"id": "git_branch", "status": "unknown_item"},
        ])
    );
    assert_eq!(
        plan["status_line_unverified"]["items"],
        json!([
            "moai-version",
            "status-line-use-theme-colors",
            "Model",
            "git_branch"
        ])
    );
    assert!(plan["status_line_unverified"]["reason"]
        .as_str()
        .unwrap()
        .contains("a41fe5b443652cbdbee57717fbb9fd78d8d8b32a"));
    // Written as given, nothing silently dropped.
    let doc: toml_edit::DocumentMut = read(&f).parse().unwrap();
    let written = doc["tui"]["status_line"].as_array().unwrap();
    assert_eq!(written.len(), asked.len());
}
