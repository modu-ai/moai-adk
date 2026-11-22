//! The hook layer of `moai update` is the config-folder layer (SPEC-20261122-next-flow-hooks
//! REQ-14): the generic hook set and the guarded deny rules are offered into each Claude
//! config folder's settings.json the user picks, never into a project's .claude/settings.json
//! by default. The guards the leader named are kept: a hand edit stays a conflict, a file
//! changed after the preview refuses the confirmation, and a revert restores. Everything runs
//! in a temporary HOME and a temporary project.

#![cfg(unix)]

mod common;
use common::CleanEnv;

use serde_json::{json, Value};
use std::{
    fs,
    path::{Path, PathBuf},
    process::{Command, Output, Stdio},
};
use tempfile::TempDir;

const EXE: &str = crate::common::MOAI_BIN;

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

struct Fixture {
    _tmp: TempDir,
    home: PathBuf,
    exe: PathBuf,
    project: PathBuf,
    path: String,
}

/// A temporary HOME with the binary, the real package and the agent sources in a version
/// folder, and a git project beside it.
fn setup() -> Fixture {
    let tmp = TempDir::new().unwrap();
    let root = fs::canonicalize(tmp.path()).unwrap();
    let home = root.join("home");
    let folder = home.join(".moai-cli/versions/a");
    fs::create_dir_all(folder.join("bin")).unwrap();
    let exe = folder.join("bin/moai-cli");
    common::copy_executable(Path::new(EXE), &exe);
    let repo = Path::new(env!("CARGO_MANIFEST_DIR"));
    copy_tree(&repo.join("skills/moai"), &folder.join("skills/moai"));
    copy_tree(&repo.join("agents"), &folder.join("agents"));
    let project = root.join("project");
    fs::create_dir_all(&project).unwrap();
    common::git::git_ok(&project, &["init", "-q"]);
    let bin = root.join("bin");
    fs::create_dir_all(&bin).unwrap();
    std::os::unix::fs::symlink(&exe, bin.join("moai")).unwrap();
    Fixture {
        _tmp: tmp,
        home,
        exe,
        project,
        path: format!("{}:/usr/bin:/bin", bin.display()),
    }
}

fn run(f: &Fixture, cwd: &Path, args: &[&str]) -> Output {
    Command::new(&f.exe)
        .args(args)
        .clean_env()
        .env("PATH", &f.path)
        .env("HOME", &f.home)
        .env("LC_ALL", "C")
        .current_dir(cwd)
        .stdin(Stdio::null())
        .output()
        .unwrap()
}

fn result(out: &Output) -> Value {
    let v: Value = serde_json::from_slice(&out.stdout).unwrap();
    assert_eq!(v["ok"], json!(true), "{v}");
    v["result"].clone()
}

fn update_preview(f: &Fixture) -> Value {
    result(&run(f, &f.project, &["update"]))
}

fn update_confirm(f: &Fixture) -> Value {
    let p = update_preview(f);
    let digest = p["plan_digest"].as_str().unwrap().to_string();
    result(&run(f, &f.project, &["update", "--confirm", &digest]))
}

/// The one config folder this HOME resolves (the account-home variable is unset, so the
/// default under the temporary HOME is used).
fn config_folder(f: &Fixture) -> PathBuf {
    f.home.join(".claude")
}

/// The config folder's settings file path.
fn config_settings(f: &Fixture) -> PathBuf {
    config_folder(f).join("settings.json")
}

/// The hex key of the config folder's ownership record.
fn record_hex(f: &Fixture) -> String {
    moai_cli::digest(config_folder(f).to_string_lossy().as_bytes())
        .trim_start_matches("sha256:")
        .get(..16)
        .unwrap_or("o")
        .to_string()
}

/// Every receipts root the HOME may resolve (the first update adopts ~/.moai with a marker,
/// so the product reads there; the version folder still sits under .moai-cli) — the fixture
/// writes its record to each so whichever the layer reads, it finds.
fn receipts_roots(home: &Path) -> Vec<PathBuf> {
    vec![home.join(".moai-cli"), home.join(".moai")]
}

fn crate_dir_placeholder(home: &Path) -> PathBuf {
    let marked = home.join(".moai");
    if marked.join("moai.json").is_file() {
        marked
    } else {
        home.join(".moai-cli")
    }
}

/// The folder row of the hooks layer's preview.
fn hook_folder(preview: &Value) -> &Value {
    preview["hooks"]["folders"]
        .as_array()
        .and_then(|f| f.first())
        .unwrap_or_else(|| panic!("no folder row in {}", preview["hooks"]))
}

const OLD_COMMAND: &str = "moai-cli hook-receipt --host claude --from-an-older-release";

/// What an older release left in the config settings: the PostToolUse handler names another
/// command, in the settings file and in the ownership record alike (MoAI wrote it, the user
/// did not touch it).
fn make_hooks_old(f: &Fixture) {
    let settings_path = config_settings(f);
    let mut settings: Value = serde_json::from_slice(&fs::read(&settings_path).unwrap()).unwrap();
    settings["hooks"]["SessionStart"][0]["hooks"][0]["command"] = json!(OLD_COMMAND);
    let mut bytes = serde_json::to_vec_pretty(&settings).unwrap();
    bytes.push(b'\n');
    fs::write(&settings_path, &bytes).unwrap();
    write_record(f, &settings["hooks"]);
}

const RECEIPT_STOP: &str =
    r#"[{"hooks":[{"type":"command","command":"moai hook-receipt --host claude","timeout":10}]}]"#;

/// What a release before the redesign installed into the config settings: the four-event set,
/// with SessionStart and Stop as plain receipts. Written to the settings file and the
/// ownership record alike, so the user is taken not to have touched it.
fn make_hooks_four_event(f: &Fixture) {
    let settings_path = config_settings(f);
    let mut settings: Value = serde_json::from_slice(&fs::read(&settings_path).unwrap()).unwrap();
    let group: Value = serde_json::from_str(RECEIPT_STOP).unwrap();
    settings["hooks"]["SessionStart"] = group.clone();
    settings["hooks"]["Stop"] = group;
    let mut bytes = serde_json::to_vec_pretty(&settings).unwrap();
    bytes.push(b'\n');
    fs::write(&settings_path, &bytes).unwrap();
    write_record(f, &settings["hooks"]);
}

/// The canonical sorted-keys digest of a hooks value — the same bytes the layer records
/// and compares, spelled here in the test so a fixture record matches the product's rule.
fn canonical_hooks_sha(hooks: &Value) -> String {
    fn canonical(v: &Value, out: &mut String) {
        match v {
            Value::Object(m) => {
                let mut keys: Vec<&String> = m.keys().collect();
                keys.sort();
                out.push('{');
                for (i, k) in keys.iter().enumerate() {
                    if i > 0 {
                        out.push(',');
                    }
                    out.push_str(&serde_json::to_string(k).unwrap());
                    out.push(':');
                    canonical(&m[*k], out);
                }
                out.push('}');
            }
            Value::Array(a) => {
                out.push('[');
                for (i, v) in a.iter().enumerate() {
                    if i > 0 {
                        out.push(',');
                    }
                    canonical(v, out);
                }
                out.push(']');
            }
            other => out.push_str(&serde_json::to_string(other).unwrap()),
        }
    }
    let mut text = String::new();
    canonical(hooks, &mut text);
    moai_cli::digest(text.as_bytes())
}

/// The ownership record that says the hooks in the file are MoAI's.
fn write_record(f: &Fixture, hooks: &Value) {
    let name = format!("config-hooks-{}.json", record_hex(f));
    for root in receipts_roots(&f.home) {
        let dir = root.join("receipts");
        fs::create_dir_all(&dir).unwrap();
        fs::write(
            dir.join(&name),
            serde_json::to_vec_pretty(&json!({
                "schema_version": 1,
                "folder": config_folder(f).display().to_string(),
                "hooks_sha256": canonical_hooks_sha(hooks),
                "settings_sha256": moai_cli::digest(
                    &fs::read(config_settings(f)).unwrap())
            }))
            .unwrap(),
        )
        .unwrap();
    }
}

/// A user's hand edit of the current set: one timeout moves.
fn hand_edit_hooks(f: &Fixture) {
    let settings_path = config_settings(f);
    let mut s: Value = serde_json::from_slice(&fs::read(&settings_path).unwrap()).unwrap();
    s["hooks"]["SessionStart"][0]["hooks"][0]["timeout"] = json!(99);
    fs::write(&settings_path, serde_json::to_vec_pretty(&s).unwrap()).unwrap();
}

/// One update makes the whole user layer current (the hook layer writes the config settings).
fn layer_current(f: &Fixture) {
    update_confirm(f);
    assert!(
        config_settings(f).is_file(),
        "the hook layer wrote the config settings"
    );
    let p = update_preview(f);
    assert_eq!(hook_folder(&p)["action"], json!("unchanged"), "{p}");
}

/// The earlier four-event set in the config settings is an update: the confirmation drops
/// Stop, moves SessionStart to the session-check duty, and the next preview has nothing more
/// to apply from the hook layer.
#[test]
fn an_installed_four_event_set_is_replanned() {
    let f = setup();
    layer_current(&f);
    make_hooks_four_event(&f);

    let p = update_preview(&f);
    assert_eq!(hook_folder(&p)["action"], json!("update"), "{}", p["hooks"]);
    let digest = p["plan_digest"].as_str().unwrap().to_string();
    let r = result(&run(&f, &f.project, &["update", "--confirm", &digest]));
    assert_eq!(r["hooks_result"]["state"], json!("applied"), "{r}");
    let text = fs::read_to_string(config_settings(&f)).unwrap();
    assert!(!text.contains(OLD_COMMAND), "{text}");
    // The exec-form command names the binary by absolute path, which legitimately runs under
    // the version folder .moai-cli; the old command SHAPE is what must be gone.
    assert!(!text.contains("moai-cli hook-"), "{text}");
    let again = update_preview(&f);
    assert_eq!(hook_folder(&again)["action"], json!("unchanged"));
}

/// An earlier set the user edited is not rewritten: the folder is a conflict, nothing is
/// offered for it, and the confirmation leaves the file as it is.
#[test]
fn an_edited_four_event_set_is_kept() {
    let f = setup();
    layer_current(&f);
    make_hooks_four_event(&f);
    hand_edit_hooks(&f);
    let before = fs::read(config_settings(&f)).unwrap();

    let p = update_preview(&f);
    let row = hook_folder(&p);
    assert_eq!(row["action"], json!("conflict"), "{row}");
    assert_eq!(row["plan_digest"], Value::Null, "{row}");

    let digest = p["plan_digest"].as_str().unwrap().to_string();
    let r = result(&run(&f, &f.project, &["update", "--confirm", &digest]));
    assert_eq!(r["hooks_result"]["state"], json!("unchanged"), "{r}");
    // The style layer may add its own key to the same file; the stranger's hooks stay as
    // they were.
    let after: Value = serde_json::from_slice(&fs::read(config_settings(&f)).unwrap()).unwrap();
    let before_v: Value = serde_json::from_slice(&before).unwrap();
    assert_eq!(after["hooks"], before_v["hooks"], "{after}");
}

/// An old MoAI set is previewed as an update, applied, and reverted byte for byte through the
/// printed receipt; the receipt reverts once.
#[test]
fn an_old_hook_set_is_previewed_applied_and_reverted() {
    let f = setup();
    layer_current(&f);
    make_hooks_old(&f);
    let settings = config_settings(&f);
    let old_settings = fs::read(&settings).unwrap();

    let p = update_preview(&f);
    assert_eq!(p["nothing_to_apply"], json!(false), "{p}");
    assert_eq!(hook_folder(&p)["action"], json!("update"), "{}", p["hooks"]);
    assert_eq!(
        fs::read(&settings).unwrap(),
        old_settings,
        "a preview writes nothing"
    );

    let digest = p["plan_digest"].as_str().unwrap().to_string();
    let r = result(&run(&f, &f.project, &["update", "--confirm", &digest]));
    assert_eq!(r["hooks_result"]["state"], json!("applied"), "{r}");
    let text = fs::read_to_string(&settings).unwrap();
    assert!(!text.contains(OLD_COMMAND), "{text}");
    assert!(!text.contains("moai-cli hook-"), "{text}");
    let id = r["hooks_result"]["receipt"].as_str().unwrap().to_string();
    assert!(r["revert"]["hooks"].as_str().unwrap().contains(&id), "{r}");

    let again = update_preview(&f);
    assert_eq!(hook_folder(&again)["action"], json!("unchanged"));

    let reverted = result(&run(
        &f,
        &f.project,
        [
            "hooks-revert",
            f.project.to_str().unwrap(),
            "--receipt",
            id.as_str(),
            "--permission-reference",
            "test-approval",
        ]
        .as_slice(),
    ));
    assert_eq!(reverted["reverted"], json!(["claude-config"]), "{reverted}");
    assert_eq!(fs::read(&settings).unwrap(), old_settings);
    // A receipt reverts once.
    let twice = run(
        &f,
        &f.project,
        &[
            "hooks-revert",
            f.project.to_str().unwrap(),
            "--receipt",
            &id,
            "--permission-reference",
            "test-approval",
        ],
    );
    assert!(!twice.status.success());
}

/// `hooks-revert` skips a config settings file that changed after the apply and writes
/// nothing to it; once the user puts the file back, the receipt works.
#[test]
fn a_revert_skips_a_host_changed_after_the_apply() {
    let f = setup();
    layer_current(&f);
    make_hooks_old(&f);
    let r = update_confirm(&f);
    let id = r["hooks_result"]["receipt"].as_str().unwrap().to_string();
    let settings = config_settings(&f);
    let written = fs::read(&settings).unwrap();
    let mut edited = written.clone();
    edited.extend_from_slice(b"\n");
    fs::write(&settings, &edited).unwrap();
    let revert = || {
        let out = run(
            &f,
            &f.project,
            &[
                "hooks-revert",
                f.project.to_str().unwrap(),
                "--receipt",
                &id,
                "--permission-reference",
                "test-approval",
            ],
        );
        serde_json::from_slice::<Value>(&out.stdout).unwrap_or(Value::Null)
    };
    // The config-layer receipt names the claude-config host; a file changed after the apply
    // is skipped whole.
    let reverted = revert();
    let skipped_or_unchanged = reverted["result"]["skipped"]
        .as_array()
        .map(|s| !s.is_empty())
        .unwrap_or(false)
        || reverted["result"]["state"] == json!("unchanged");
    assert!(skipped_or_unchanged, "{reverted}");
    assert_eq!(fs::read(&settings).unwrap(), edited);
    // Nothing was restored, so the receipt still works once the user puts the file back.
    fs::write(&settings, &written).unwrap();
    let again = revert();
    assert_eq!(again["result"]["state"], json!("reverted"), "{again}");
    assert_eq!(
        again["result"]["reverted"],
        json!(["claude-config"]),
        "{again}"
    );
}

/// A user's edit of a current set is kept: the edit survives a confirmation untouched (the
/// leader's word).
#[test]
fn a_hand_edit_of_a_current_hook_set_is_kept() {
    let f = setup();
    layer_current(&f);
    hand_edit_hooks(&f);
    let settings = config_settings(&f);
    let edited = fs::read(&settings).unwrap();
    let p = update_preview(&f);
    assert_eq!(
        hook_folder(&p)["action"],
        json!("conflict"),
        "{}",
        p["hooks"]
    );
    let digest = p["plan_digest"].as_str().unwrap().to_string();
    result(&run(&f, &f.project, &["update", "--confirm", &digest]));
    assert_eq!(fs::read(&settings).unwrap(), edited);
}

/// A hooks key MoAI never wrote — a stranger's set in the config settings — is a conflict,
/// named in the preview, and never overwritten by a confirmation.
#[test]
fn hand_edited_hooks_are_conflicts_and_are_never_overwritten() {
    let f = setup();
    // No layer has run: the config settings hold a stranger's hooks key only.
    fs::create_dir_all(config_folder(&f)).unwrap();
    fs::write(
        config_settings(&f),
        "{\"hooks\": {\"SessionStart\": [{\"hooks\": [{\"command\": \"someone-else\", \"type\": \"command\", \"timeout\": 10}]}]}}\n",
    )
    .unwrap();
    let before = fs::read(config_settings(&f)).unwrap();

    let p = update_preview(&f);
    let row = hook_folder(&p);
    assert_eq!(row["action"], json!("conflict"), "{row}");
    assert!(row["reason"].as_str().is_some(), "{row}");
    let digest = p["plan_digest"].as_str().unwrap().to_string();
    let r = result(&run(&f, &f.project, &["update", "--confirm", &digest]));
    assert_eq!(r["hooks_result"]["state"], json!("unchanged"), "{r}");
    // The style layer may add its own key to the same file; the stranger's hooks stay as
    // they were.
    let after: Value = serde_json::from_slice(&fs::read(config_settings(&f)).unwrap()).unwrap();
    let before_v: Value = serde_json::from_slice(&before).unwrap();
    assert_eq!(after["hooks"], before_v["hooks"], "{after}");
}

/// A current config set adds nothing to apply from the hook layer; a project's own hook files
/// (installed by hooks-apply) are never touched by the update.
#[test]
fn a_current_or_absent_hook_set_leaves_nothing_to_apply() {
    let f = setup();
    layer_current(&f);
    let p = update_preview(&f);
    assert_eq!(hook_folder(&p)["action"], json!("unchanged"), "{p}");

    // A project's own MoAI hook set (hooks-apply) stays exactly as the lane installed it.
    let project = f.project.to_str().unwrap().to_string();
    let plan = result(&run(
        &f,
        &f.project,
        &["hooks-plan", &project, "--host", "claude"],
    ));
    let digest = plan["plan_digest"].as_str().unwrap().to_string();
    result(&run(
        &f,
        &f.project,
        &[
            "hooks-apply",
            &project,
            "--host",
            "claude",
            "--plan-digest",
            &digest,
            "--permission-reference",
            "test-approval",
        ],
    ));
    let project_settings = f.project.join(".claude/settings.json");
    let installed = fs::read(&project_settings).unwrap();
    update_confirm(&f);
    assert_eq!(fs::read(&project_settings).unwrap(), installed);
}

/// The update never touches a Codex project hook file: a lane's Stop hook installed by
/// hooks-apply survives a confirmation as it was.
#[test]
fn update_keeps_the_codex_lane_stop_hook() {
    let f = setup();
    let p = f.project.to_str().unwrap().to_string();
    let lane = ["--lane-loop", "agent-1", "--repo", "o/r"];
    let mut args = vec!["hooks-plan", p.as_str(), "--host", "codex"];
    args.extend(lane.iter().copied());
    let plan = result(&run(&f, &f.project, &args));
    let digest = plan["plan_digest"].as_str().unwrap().to_string();
    let mut apply = vec![
        "hooks-apply",
        p.as_str(),
        "--host",
        "codex",
        "--plan-digest",
        digest.as_str(),
        "--permission-reference",
        "test-approval",
    ];
    apply.extend(lane.iter().copied());
    result(&run(&f, &f.project, &apply));
    let file = f.project.join(".codex/hooks.json");
    let with_lane = fs::read(&file).unwrap();
    update_confirm(&f);
    assert_eq!(fs::read(&file).unwrap(), with_lane);
}

/// The confirmation binds the hook layer: a config settings file changed after the preview
/// refuses the whole update before any write.
#[test]
fn a_hook_file_changed_after_the_preview_refuses_the_confirmation() {
    let f = setup();
    layer_current(&f);
    make_hooks_old(&f);
    let p = update_preview(&f);
    let digest = p["plan_digest"].as_str().unwrap().to_string();
    let settings = config_settings(&f);
    let mut s = fs::read(&settings).unwrap();
    s.extend_from_slice(b"\n");
    fs::write(&settings, &s).unwrap();
    let out = run(&f, &f.project, &["update", "--confirm", &digest]);
    assert!(!out.status.success());
    assert!(
        String::from_utf8_lossy(&out.stdout).contains("stale or altered")
            || String::from_utf8_lossy(&out.stderr).contains("stale or altered")
            || String::from_utf8_lossy(&out.stdout).contains("changed since the preview"),
        "stdout {} stderr {}",
        String::from_utf8_lossy(&out.stdout),
        String::from_utf8_lossy(&out.stderr)
    );
    assert_eq!(fs::read(&settings).unwrap(), s);
}

/// The receipt id is the only thing that keeps `hooks-revert` inside the receipt folder: an id
/// with a path step, without the `hooks-` prefix, or longer than 64 characters is refused before
/// any path is built; a 64-character id of the right shape is looked up (and not found).
#[test]
fn a_receipt_id_of_another_shape_is_refused() {
    let f = setup();
    let p = f.project.to_str().unwrap();
    let revert = |id: &str| {
        run(
            &f,
            &f.project,
            &[
                "hooks-revert",
                p,
                "--receipt",
                id,
                "--permission-reference",
                "test-approval",
            ],
        )
    };
    let len65 = format!("hooks-{}", "a".repeat(59));
    let len64 = format!("hooks-{}", "a".repeat(58));
    assert_eq!((len64.len(), len65.len()), (64, 65));
    for id in [
        "hooks-../../x",
        "../hooks-1",
        "hooks-1/../../x",
        "agents-1",
        "hooks",
        len65.as_str(),
    ] {
        let out = revert(id);
        assert!(!out.status.success(), "{id}");
        let text = String::from_utf8_lossy(&out.stdout);
        assert!(text.contains("unknown hook receipt id"), "{id}: {text}");
    }
    let out = revert(&len64);
    assert!(!out.status.success());
    let text = String::from_utf8_lossy(&out.stdout);
    assert!(text.contains("hook receipt not found"), "{text}");
}

/// Edits the `PostToolUse` handler's timeout in the settings file only, as a user would.
fn hand_edit_claude_hooks(f: &Fixture) {
    let settings = f.project.join(".claude/settings.json");
    let mut s: Value = serde_json::from_slice(&fs::read(&settings).unwrap()).unwrap();
    s["hooks"]["PostToolUse"][0]["hooks"][0]["timeout"] = json!(99);
    fs::write(&settings, serde_json::to_vec_pretty(&s).unwrap()).unwrap();
}

/// Walks `dir`, returning every folder and every file below it with its permission bits.
fn modes_below(dir: &Path, out: &mut Vec<(PathBuf, bool, u32)>) {
    use std::os::unix::fs::PermissionsExt;
    for e in fs::read_dir(dir).unwrap() {
        let e = e.unwrap();
        let meta = fs::symlink_metadata(e.path()).unwrap();
        out.push((e.path(), meta.is_dir(), meta.permissions().mode() & 0o777));
        if meta.is_dir() {
            modes_below(e.path().as_path(), out);
        }
    }
}

/// Audit FAIL 6882751326 B1 kept: the config-folder apply's receipt keeps the bytes it
/// replaces under the HOME state root owner-only whatever the umask, and the kept copy of a
/// settings file holds what the file held.
#[test]
fn the_hook_receipt_keeps_a_protected_settings_file_privately_under_any_umask() {
    use std::os::unix::fs::PermissionsExt;
    for umask in ["000", "022", "077"] {
        let f = setup();
        layer_current(&f);
        let settings_path = config_settings(&f);
        let mut settings: Value =
            serde_json::from_slice(&fs::read(&settings_path).unwrap()).unwrap();
        settings["env"] = json!({"TOKEN_FOR_TEST": "TESTVALUE-not-secret"});
        fs::write(
            &settings_path,
            serde_json::to_vec_pretty(&settings).unwrap(),
        )
        .unwrap();
        make_hooks_old(&f);

        let digest = update_preview(&f)["plan_digest"]
            .as_str()
            .unwrap()
            .to_string();
        let out = Command::new("/bin/sh")
            .args(["-c", &format!("umask {umask}; exec \"$0\" \"$@\"")])
            .arg(&f.exe)
            .args(["update", "--confirm", &digest])
            .clean_env()
            .env("PATH", &f.path)
            .env("HOME", &f.home)
            .env("LC_ALL", "C")
            .current_dir(&f.project)
            .stdin(Stdio::null())
            .output()
            .unwrap();
        let r = result(&out);
        assert_eq!(
            r["hooks_result"]["state"],
            json!("applied"),
            "umask {umask}: {r}"
        );
        let id = r["hooks_result"]["receipt"].as_str().unwrap();

        let receipts = [f.home.join(".moai-cli"), f.home.join(".moai")]
            .into_iter()
            .map(|root| root.join("receipts").join(id))
            .find(|d| d.join("receipt.json").is_file())
            .unwrap_or_else(|| f.home.join(".moai").join("receipts").join(id));
        assert!(receipts.is_dir(), "umask {umask}: {receipts:?}");
        let mut seen = Vec::new();
        modes_below(&receipts, &mut seen);
        let own = fs::symlink_metadata(&receipts)
            .unwrap()
            .permissions()
            .mode()
            & 0o777;
        assert_eq!(own & 0o077, 0, "umask {umask}: receipts folder is {own:o}");
        let mut kept = 0;
        for (path, is_dir, mode) in &seen {
            if path.file_name().is_some_and(|n| n == ".gitignore") {
                continue;
            }
            assert_eq!(
                mode & 0o077,
                0,
                "umask {umask}: {} is {mode:o}",
                path.display()
            );
            if !is_dir && path.to_string_lossy().contains("/before/") {
                let text = fs::read_to_string(path).unwrap();
                kept += usize::from(text.contains("TESTVALUE-not-secret"));
            }
        }
        assert_eq!(
            kept, 1,
            "umask {umask}: the kept settings copy must exist: {seen:?}"
        );
        if let Ok(text) = fs::read_to_string(receipts.join(".gitignore")) {
            assert_eq!(text, "*\n", "umask {umask}");
        }
    }
}
