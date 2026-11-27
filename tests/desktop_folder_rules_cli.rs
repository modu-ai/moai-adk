//! SPEC-20261124-desktop-install-agents: the flow rule sits directly under `rules/` of the
//! default Claude folder (`rules/moai-flow.md`), an earlier `rules/moai/flow.md` is left as it
//! is and named by the preview, and a receipt reverts the path its kind names. Every test runs
//! against a temporary HOME only, never the real one.

#[cfg(unix)]
mod common;

use moai_cli::home_routing::{apply, plan, revert, Plan};
use std::fs;
use std::path::{Path, PathBuf};

const PERM: &str = "test-permission";
const NEW_FLOW: &str = ".claude/rules/moai-flow.md";
const OLD_FLOW: &str = ".claude/rules/moai/flow.md";

fn package() -> PathBuf {
    Path::new(env!("CARGO_MANIFEST_DIR")).join("skills/moai")
}

fn temp_home() -> (tempfile::TempDir, PathBuf) {
    let dir = tempfile::tempdir().unwrap();
    let path = fs::canonicalize(dir.path()).unwrap();
    (dir, path)
}

fn packaged_flow() -> String {
    fs::read_to_string(package().join("references/home/flow.md")).unwrap()
}

fn flow_row(p: &Plan) -> &moai_cli::agents::FilePlan {
    p.hosts[0]
        .flow
        .as_ref()
        .expect("the package has a flow rule")
}

fn write(path: &Path, bytes: &str) {
    fs::create_dir_all(path.parent().unwrap()).unwrap();
    fs::write(path, bytes).unwrap();
}

#[test]
fn flow_rule_is_written_directly_under_rules() {
    let (_t, home) = temp_home();
    let p = plan(&home, &package()).unwrap();
    let row = flow_row(&p);
    assert_eq!(row.action, "create");
    assert!(
        Path::new(&row.path).ends_with("rules/moai-flow.md"),
        "the plan names {}",
        row.path
    );
    let r = apply(&home, &package(), &p.plan_digest, PERM).unwrap();
    assert_eq!(r["state"], "applied", "{r}");
    assert_eq!(
        fs::read_to_string(home.join(NEW_FLOW)).unwrap(),
        packaged_flow()
    );
    assert!(
        !home.join(".claude/rules/moai").exists(),
        "no rules/moai folder is made"
    );
    // The routing file stays where it was.
    assert!(home.join(".claude/rules/moai.md").is_file());
}

#[test]
fn earlier_flow_file_is_left_in_place() {
    let (_t, home) = temp_home();
    let old = home.join(OLD_FLOW);
    write(&old, "an earlier build wrote this\n");
    let before = fs::metadata(&old).unwrap().modified().unwrap();
    let p = plan(&home, &package()).unwrap();
    let row = flow_row(&p);
    assert_eq!(row.action, "create", "the new file is still planned");
    let detail = row.detail.as_deref().unwrap_or("");
    assert!(
        detail.contains("rules/moai/flow.md") && detail.contains("left in place"),
        "the preview names the earlier file: {detail:?}"
    );
    let r = apply(&home, &package(), &p.plan_digest, PERM).unwrap();
    assert_eq!(r["state"], "applied", "{r}");
    assert_eq!(
        fs::read_to_string(&old).unwrap(),
        "an earlier build wrote this\n"
    );
    assert_eq!(fs::metadata(&old).unwrap().modified().unwrap(), before);
    assert_eq!(
        fs::read_to_string(home.join(NEW_FLOW)).unwrap(),
        packaged_flow()
    );
    let again = plan(&home, &package()).unwrap();
    assert_eq!(flow_row(&again).action, "unchanged");
    assert_eq!(again.hosts[0].section.as_ref().unwrap().action, "unchanged");
    let r = apply(&home, &package(), &again.plan_digest, PERM).unwrap();
    assert_eq!(r["state"], "unchanged");
}

#[test]
fn earlier_flow_note_is_added_after_an_existing_detail() {
    // The flow row already carries a detail (it waits behind a conflicting routing file); the
    // note about the earlier file joins that detail and must not replace it.
    let (_t, home) = temp_home();
    write(&home.join(".claude/rules/moai.md"), "the user's own rule\n");
    write(&home.join(OLD_FLOW), "an earlier build wrote this\n");
    let p = plan(&home, &package()).unwrap();
    let row = flow_row(&p);
    assert_eq!(
        row.action, "held",
        "the flow file waits behind the conflict"
    );
    let detail = row.detail.as_deref().unwrap_or("");
    assert!(
        detail.contains("is a conflict"),
        "the earlier reason stays in the detail: {detail:?}"
    );
    assert!(
        detail.contains("rules/moai/flow.md") && detail.contains("left in place"),
        "the note about the earlier file follows it: {detail:?}"
    );
}

#[test]
fn no_earlier_flow_file_means_no_note() {
    // Without an earlier `rules/moai/flow.md` the preview says nothing about one, both for a
    // plain create and for a row that already carries a detail.
    let (_t, home) = temp_home();
    let p = plan(&home, &package()).unwrap();
    let row = flow_row(&p);
    assert_eq!(row.action, "create");
    let detail = row.detail.as_deref().unwrap_or("");
    assert!(
        !detail.contains("rules/moai/flow.md") && !detail.contains("earlier build"),
        "no note without an earlier file: {detail:?}"
    );

    let (_t2, home) = temp_home();
    write(&home.join(".claude/rules/moai.md"), "the user's own rule\n");
    let p = plan(&home, &package()).unwrap();
    let row = flow_row(&p);
    assert_eq!(row.action, "held");
    let detail = row.detail.as_deref().unwrap_or("");
    assert!(
        detail.contains("is a conflict"),
        "the held reason is kept: {detail:?}"
    );
    assert!(
        !detail.contains("rules/moai/flow.md") && !detail.contains("earlier build"),
        "no note without an earlier file: {detail:?}"
    );
}

#[test]
fn revert_follows_the_kind() {
    // A receipt of the new apply reverts the new path.
    let (_t, home) = temp_home();
    let p = plan(&home, &package()).unwrap();
    let r = apply(&home, &package(), &p.plan_digest, PERM).unwrap();
    let id = r["receipt"].as_str().unwrap().to_string();
    assert!(home.join(NEW_FLOW).is_file());
    let r = revert(&home, &id, PERM).unwrap();
    assert_eq!(r["state"], "reverted", "{r}");
    assert!(!home.join(NEW_FLOW).exists(), "the new file goes");

    // An entry of kind `flow`, as the earlier layout wrote it, reverts `rules/moai/flow.md`
    // and never touches `rules/moai-flow.md`.
    let (_t2, home) = temp_home();
    let bytes = "the flow rule an earlier build wrote\n";
    write(&home.join(OLD_FLOW), bytes);
    write(&home.join(NEW_FLOW), bytes);
    let claude = home.join(".claude");
    let id = "routing-1000-0";
    let dir = moai_cli::home_root::path(&home)
        .join("home-routing/receipts")
        .join(id);
    fs::create_dir_all(&dir).unwrap();
    let receipt = serde_json::json!({
        "schema_version": 1,
        "id": id,
        "plan_digest": "sha256:unused",
        "permission_reference": PERM,
        "status": "committed",
        "entries": [{
            "host": "claude",
            "kind": "flow",
            "config_dir": claude.to_string_lossy(),
            "action": "create",
            "before_sha256": null,
            "after_sha256": moai_cli::digest(bytes.as_bytes()),
            "baseline_before": null
        }]
    });
    fs::write(dir.join("receipt.json"), receipt.to_string()).unwrap();
    let r = revert(&home, id, PERM).unwrap();
    assert_eq!(r["state"], "reverted", "{r}");
    assert!(
        !home.join(OLD_FLOW).exists(),
        "the earlier path is what an entry of kind flow reverts: {r}"
    );
    assert_eq!(
        fs::read_to_string(home.join(NEW_FLOW)).unwrap(),
        bytes,
        "the new path is not touched by an earlier receipt"
    );
}

/// The `moai update` preview names the Claude folders its parts write (REQ-5 to REQ-9). It runs
/// the real binary in a temporary HOME beside a copy of the package; the real HOME is never read.
#[cfg(unix)]
mod preview {
    use serde_json::{json, Value};
    use std::fs;
    use std::path::{Path, PathBuf};
    use std::process::{Command, Stdio};

    use crate::common::{self, CleanEnv};

    const EXE: &str = common::MOAI_BIN;
    const ENV_UNSET: &str = "env -u CLAUDE_CONFIG_DIR moai update";

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

    /// A temporary HOME with `versions/a/bin/moai-cli` under its state folder (found through
    /// `home_root`, before anything is created there) and the real package and agent sources
    /// beside it.
    fn setup() -> (tempfile::TempDir, PathBuf, PathBuf) {
        let tmp = tempfile::TempDir::new().unwrap();
        let root = fs::canonicalize(tmp.path()).unwrap();
        let home = root.join("home");
        let folder = moai_cli::home_root::path(&home).join("versions/a");
        fs::create_dir_all(folder.join("bin")).unwrap();
        let exe = folder.join("bin/moai-cli");
        common::copy_executable(Path::new(EXE), &exe);
        let repo = Path::new(env!("CARGO_MANIFEST_DIR"));
        copy_tree(&repo.join("skills/moai"), &folder.join("skills/moai"));
        copy_tree(&repo.join("agents"), &folder.join("agents"));
        (tmp, home, exe)
    }

    /// `moai update` (a preview) with `vars` set; `account` makes HOME the account home, so the
    /// Claude variable counts.
    fn preview(exe: &Path, home: &Path, account: bool, vars: &[(&str, &Path)]) -> Value {
        let mut cmd = Command::new(exe);
        cmd.arg("update")
            .clean_env()
            .env("PATH", "/usr/bin:/bin")
            .env("HOME", home)
            .env("LC_ALL", "C")
            .current_dir(home)
            .stdin(Stdio::null());
        if account {
            cmd.env("MOAI_ACCOUNT_HOME", home);
        }
        for (k, v) in vars {
            cmd.env(k, v);
        }
        let out = cmd.output().unwrap();
        let v: Value = serde_json::from_slice(&out.stdout).unwrap_or_else(|e| {
            panic!(
                "stdout is not JSON ({e}): {} {}",
                String::from_utf8_lossy(&out.stdout),
                String::from_utf8_lossy(&out.stderr)
            )
        });
        assert_eq!(v["ok"], json!(true), "{v}");
        v["result"].clone()
    }

    /// The one `claude_folders` notice of a preview.
    fn folders_notice(p: &Value) -> Value {
        let all: Vec<&Value> = p["notices"]
            .as_array()
            .unwrap()
            .iter()
            .filter(|n| n["kind"] == json!("claude_folders"))
            .collect();
        assert_eq!(all.len(), 1, "one claude_folders notice: {p}");
        all[0].clone()
    }

    fn folder(n: &Value, part: &str) -> PathBuf {
        PathBuf::from(
            n["folders"][part]
                .as_str()
                .unwrap_or_else(|| panic!("no folder for {part}: {n}")),
        )
    }

    /// Every file under `dir` with its bytes, leaving out the `versions` folder that holds the
    /// installed binary `exe` (`<state>/versions/a/bin/<exe>`).
    fn files(dir: &Path, exe: &Path) -> Vec<(PathBuf, Vec<u8>)> {
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
        let versions = exe.ancestors().nth(3).expect("exe sits in versions/a/bin");
        walk(dir, versions, &mut out);
        out.sort();
        out
    }

    #[test]
    fn preview_names_each_claude_folder() {
        let (_t, home, exe) = setup();
        let default = home.join(".claude");
        let conf = home.join("elsewhere/claude-conf");

        // The variable is unset: one folder for every part, no warning.
        let n = folders_notice(&preview(&exe, &home, true, &[]));
        for part in ["skills", "agents", "style", "rules"] {
            assert_eq!(folder(&n, part), default, "{part}: {n}");
        }
        assert_eq!(n["warning"], json!(false), "{n}");
        assert!(!n["reason"].as_str().unwrap().contains(ENV_UNSET), "{n}");

        // Another absolute folder under this account: both folders, the warning and the command.
        let n = folders_notice(&preview(&exe, &home, true, &[("CLAUDE_CONFIG_DIR", &conf)]));
        for part in ["skills", "agents", "style"] {
            assert_eq!(folder(&n, part), conf, "{part}: {n}");
        }
        assert_eq!(folder(&n, "rules"), default, "the routing copy: {n}");
        assert_eq!(n["warning"], json!(true), "{n}");
        let reason = n["reason"].as_str().unwrap();
        for want in [
            conf.to_str().unwrap(),
            default.to_str().unwrap(),
            "Desktop",
            "login shell",
            ENV_UNSET,
        ] {
            assert!(reason.contains(want), "missing {want:?}: {reason}");
        }

        // The variable names the default folder: one folder, no warning.
        let n = folders_notice(&preview(
            &exe,
            &home,
            true,
            &[("CLAUDE_CONFIG_DIR", &default)],
        ));
        for part in ["skills", "agents", "style", "rules"] {
            assert_eq!(folder(&n, part), default, "{part}: {n}");
        }
        assert_eq!(n["warning"], json!(false), "{n}");

        // The variable does not count for this account (HOME is not the account home).
        let n = folders_notice(&preview(
            &exe,
            &home,
            false,
            &[("CLAUDE_CONFIG_DIR", &conf)],
        ));
        for part in ["skills", "agents", "style", "rules"] {
            assert_eq!(folder(&n, part), default, "{part}: {n}");
        }
        assert_eq!(n["warning"], json!(false), "{n}");
    }

    /// SPEC-20261127-update-linked-profile-display AC-4 (REQ-7, REQ-8): when every entry of the
    /// profile that the update writes is a link into the default folder, the notice names the
    /// default folder's paths and does not warn that the update writes into the profile; one
    /// entry that is a real folder in the profile names that place and keeps the warning.
    #[test]
    fn a_fully_linked_profile_is_shown_as_the_default_folder() {
        use std::os::unix::fs::symlink;
        let (_t, home, exe) = setup();
        let default = home.join(".claude");
        let profile = home.join("profiles/p");
        fs::create_dir_all(&profile).unwrap();
        for entry in ["skills", "agents", "output-styles", "rules"] {
            fs::create_dir_all(default.join(entry)).unwrap();
            symlink(default.join(entry), profile.join(entry)).unwrap();
        }
        fs::write(default.join("settings.json"), "{\"a\": 1}\n").unwrap();
        symlink(default.join("settings.json"), profile.join("settings.json")).unwrap();
        let vars = [("CLAUDE_CONFIG_DIR", profile.as_path())];

        let n = folders_notice(&preview(&exe, &home, true, &vars));
        assert_eq!(n["warning"], json!(false), "{n}");
        for part in ["skills", "agents", "style", "rules"] {
            assert_eq!(folder(&n, part), default, "{part}: {n}");
        }
        assert!(
            !n.to_string().contains(profile.to_str().unwrap()),
            "no profile path: {n}"
        );
        let reason = n["reason"].as_str().unwrap();
        assert!(reason.contains(default.to_str().unwrap()), "{reason}");
        assert!(!reason.contains("CLAUDE_CONFIG_DIR names"), "{reason}");

        // One entry that is a real folder in the profile: that place and the warning.
        fs::remove_file(profile.join("skills")).unwrap();
        fs::create_dir_all(profile.join("skills")).unwrap();
        let n = folders_notice(&preview(&exe, &home, true, &vars));
        assert_eq!(n["warning"], json!(true), "{n}");
        assert_eq!(folder(&n, "skills"), profile, "{n}");
        let reason = n["reason"].as_str().unwrap();
        assert!(reason.contains("CLAUDE_CONFIG_DIR names"), "{reason}");
        assert!(reason.contains(profile.to_str().unwrap()), "{reason}");
    }

    /// Every entry of the profile is a link into the default folder except `settings.json`, a real
    /// file there: the update writes the output style setting into that file, so the notice keeps
    /// its warning and the profile's paths (the real-file side of the `settings.json` rule).
    #[test]
    fn a_real_settings_file_in_a_linked_profile_keeps_the_warning() {
        use std::os::unix::fs::symlink;
        let (_t, home, exe) = setup();
        let default = home.join(".claude");
        let profile = home.join("profiles/p");
        fs::create_dir_all(&profile).unwrap();
        for entry in ["skills", "agents", "output-styles", "rules"] {
            fs::create_dir_all(default.join(entry)).unwrap();
            symlink(default.join(entry), profile.join(entry)).unwrap();
        }
        fs::write(default.join("settings.json"), "{\"a\": 1}\n").unwrap();
        fs::write(profile.join("settings.json"), "{\"a\": 1}\n").unwrap();
        let vars = [("CLAUDE_CONFIG_DIR", profile.as_path())];

        let n = folders_notice(&preview(&exe, &home, true, &vars));
        assert_eq!(n["warning"], json!(true), "{n}");
        let reason = n["reason"].as_str().unwrap();
        assert!(reason.contains("CLAUDE_CONFIG_DIR names"), "{reason}");
        assert!(reason.contains(profile.to_str().unwrap()), "{reason}");
    }

    #[test]
    fn notice_is_read_only() {
        let (_t, home, exe) = setup();
        let conf = home.join("elsewhere/claude-conf");
        let before = files(&home, &exe);
        let a = preview(&exe, &home, true, &[("CLAUDE_CONFIG_DIR", &conf)]);
        assert_eq!(folders_notice(&a)["warning"], json!(true), "{a}");
        let b = preview(&exe, &home, true, &[("CLAUDE_CONFIG_DIR", &conf)]);
        assert_eq!(
            a["plan_digest"], b["plan_digest"],
            "the same preview, the same digest"
        );
        assert_eq!(files(&home, &exe), before, "the preview wrote nothing");
        assert!(!conf.exists(), "the named folder was not made");
        assert!(!home.join(".claude").exists(), "no default folder was made");
        // The confirm line a preview prints carries the digest of the parts, not of the notice.
        assert!(
            a["confirm_with"]
                .as_str()
                .unwrap()
                .ends_with(a["plan_digest"].as_str().unwrap()),
            "{a}"
        );
    }
}
