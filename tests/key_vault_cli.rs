//! #2393: the encrypted key vault through the real binary (user decisions, Issue #2393
//! comments 6777274025 and 6778671862). Every case uses a temporary HOME, a temporary
//! working directory and a synthetic key made at run time; the real home, any OS key store and
//! any legacy key file are never read. Every answer (stdout and stderr) is checked to hold no
//! key text, and no stored file holds it in clear.
#![cfg(unix)]

mod common;
use serde_json::Value;
use std::fs;
use std::io::Write;
use std::os::unix::fs::PermissionsExt;
use std::path::{Path, PathBuf};
use std::process::{Command, Output, Stdio};

const BIN: &str = crate::common::MOAI_BIN;

/// A synthetic key that cannot be a real one.
fn synthetic() -> String {
    format!("moai-test-key-{}-{}", std::process::id(), line!())
}

/// A temporary HOME and, inside it, a working directory that is not a git work tree.
fn home() -> (tempfile::TempDir, PathBuf) {
    let dir = tempfile::TempDir::new().unwrap();
    let home = fs::canonicalize(dir.path()).unwrap();
    fs::create_dir(home.join("work")).unwrap();
    (dir, home)
}

fn run_in(home: &Path, cwd: &Path, args: &[&str], stdin: Option<&str>) -> Output {
    let mut child = Command::new(BIN)
        .args(args)
        .current_dir(cwd)
        .env("HOME", home)
        .env("GIT_CEILING_DIRECTORIES", home)
        .env_remove("ANTHROPIC_AUTH_TOKEN")
        .stdin(Stdio::piped())
        .stdout(Stdio::piped())
        .stderr(Stdio::piped())
        .spawn()
        .unwrap();
    {
        let mut pipe = child.stdin.take().unwrap();
        if let Some(text) = stdin {
            pipe.write_all(text.as_bytes()).unwrap();
        }
    }
    child.wait_with_output().unwrap()
}

/// Runs with `--shared` added, from the non-git working directory.
fn run(home: &Path, args: &[&str], stdin: Option<&str>) -> Output {
    let mut all = args.to_vec();
    all.push("--shared");
    run_in(home, &home.join("work"), &all, stdin)
}

/// The envelope, after checking that neither stream holds `secret`.
fn envelope(out: &Output, secret: &str) -> Value {
    let stdout = String::from_utf8_lossy(&out.stdout);
    let stderr = String::from_utf8_lossy(&out.stderr);
    assert!(!stdout.contains(secret), "key text in stdout");
    assert!(!stderr.contains(secret), "key text in stderr");
    serde_json::from_str(&stdout).unwrap_or_else(|e| panic!("{e}: {stdout} {stderr}"))
}

fn vault(home: &Path) -> PathBuf {
    home.join(".moai-cli/vault")
}

fn entry(home: &Path) -> PathBuf {
    vault(home).join("_shared/glm.key")
}

fn mode(path: &Path) -> u32 {
    fs::metadata(path).unwrap().permissions().mode() & 0o777
}

fn holds(path: &Path, text: &str) -> bool {
    let bytes = fs::read(path).unwrap();
    bytes.windows(text.len()).any(|w| w == text.as_bytes())
}

#[test]
fn set_then_status_then_remove_round_trip_without_echoing_the_key() {
    let (_d, home) = home();
    let key = synthetic();
    let s = envelope(&run(&home, &["key-status", "glm"], None), &key);
    assert_eq!(s["ok"], true, "{s}");
    assert_eq!(s["result"]["scope"], "shared");
    assert_eq!(s["result"]["entry"], "absent");

    let set = envelope(
        &run(&home, &["key-set", "glm"], Some(&format!("  {key}\n"))),
        &key,
    );
    assert_eq!(set["result"]["written"], true, "{set}");
    assert_eq!(set["result"]["encrypted"], true);
    assert_eq!(set["result"]["value_shown"], false);
    // Encrypted on disk: the header, and no key text anywhere in the vault.
    assert!(fs::read(entry(&home)).unwrap().starts_with(b"MOAIKV1\n"));
    assert!(!holds(&entry(&home), &key));
    assert_eq!(
        fs::metadata(vault(&home).join(".master.key"))
            .unwrap()
            .len(),
        32
    );
    for (path, want) in [
        (entry(&home), 0o600),
        (vault(&home).join(".master.key"), 0o600),
        (vault(&home).join("_shared"), 0o700),
        (vault(&home), 0o700),
    ] {
        assert_eq!(mode(&path), want, "{}", path.display());
    }
    // No temporary file is left next to the entry.
    let names: Vec<String> = fs::read_dir(vault(&home).join("_shared"))
        .unwrap()
        .map(|e| e.unwrap().file_name().to_string_lossy().into_owned())
        .collect();
    assert_eq!(names, ["glm.key"]);

    let s = envelope(&run(&home, &["key-status", "glm"], None), &key);
    assert_eq!(s["result"]["entry"], "present", "{s}");
    assert_eq!(s["result"]["blocked"], Value::Null);

    // Replacing keeps one entry, mode 0600 and the same machine key.
    let master = fs::read(vault(&home).join(".master.key")).unwrap();
    let key2 = format!("{key}-2");
    envelope(&run(&home, &["key-set", "glm"], Some(&key2)), &key2);
    assert!(!holds(&entry(&home), &key2));
    assert_eq!(mode(&entry(&home)), 0o600);
    assert_eq!(fs::read(vault(&home).join(".master.key")).unwrap(), master);

    // Without --yes nothing is removed; with it only the entry goes.
    let r = envelope(&run(&home, &["key-remove", "glm"], None), &key2);
    assert_eq!(r["result"]["removed"], false);
    assert!(entry(&home).exists());
    let r = envelope(&run(&home, &["key-remove", "glm", "--yes"], None), &key2);
    assert_eq!(r["result"]["removed"], true);
    assert!(!entry(&home).exists());
    assert!(vault(&home).join(".master.key").exists());
}

/// Without `--shared` the entry belongs to the project of the working directory:
/// `<name>-<8 hex of clone_id>`. Outside a git work tree only `--shared` works.
#[test]
fn the_project_folder_comes_from_the_working_directory() {
    let (_d, home) = home();
    let key = synthetic();
    let out = run_in(&home, &home.join("work"), &["key-set", "glm"], Some(&key));
    let e = envelope(&out, &key);
    assert_eq!(e["ok"], false, "{e}");
    assert!(!vault(&home).exists());

    let repo = home.join("My Repo");
    fs::create_dir(&repo).unwrap();
    let init = common::git::command()
        .args(["init", "-q"])
        .current_dir(&repo)
        .status()
        .unwrap();
    assert!(init.success());
    let set = envelope(&run_in(&home, &repo, &["key-set", "glm"], Some(&key)), &key);
    assert_eq!(set["ok"], true, "{set}");
    assert_eq!(set["result"]["scope"], "project");
    let folder = set["result"]["folder"].as_str().unwrap().to_string();
    let (name, hex) = folder.rsplit_once('-').unwrap();
    assert_eq!(name, "my-repo", "{folder}");
    assert!(
        hex.len() == 8 && hex.bytes().all(|b| b.is_ascii_hexdigit()),
        "{folder}"
    );
    let path = vault(&home).join(&folder).join("glm.key");
    assert!(path.exists() && !holds(&path, &key));
    assert_eq!(mode(&vault(&home).join(&folder)), 0o700);
    // A subfolder of the project is the same project.
    fs::create_dir(repo.join("sub")).unwrap();
    let s = envelope(
        &run_in(&home, &repo.join("sub"), &["key-status", "glm"], None),
        &key,
    );
    assert_eq!(s["result"]["folder"], folder.as_str());
    assert_eq!(s["result"]["entry"], "present", "{s}");
    // The shared folder is separate.
    let s = envelope(
        &run_in(&home, &repo, &["key-status", "glm", "--shared"], None),
        &key,
    );
    assert_eq!(s["result"]["entry"], "absent", "{s}");
}

#[test]
fn unusable_values_are_refused_and_nothing_is_written() {
    let (_d, home) = home();
    for bad in ["", "   \n", "a\u{1}b"] {
        let out = run(&home, &["key-set", "glm"], Some(bad));
        let env = envelope(&out, "never-written");
        assert_eq!(env["ok"], false, "{bad:?}: {env}");
        assert!(!vault(&home).exists(), "{bad:?}");
    }
}

#[test]
fn group_or_other_bits_make_the_vault_unsafe_with_the_chmod_fix() {
    let (_d, home) = home();
    let key = synthetic();
    envelope(&run(&home, &["key-set", "glm"], Some(&key)), &key);
    let unsafe_with = |fix: &str| {
        let s = envelope(&run(&home, &["key-status", "glm"], None), &key);
        assert_eq!(s["result"]["entry"], "unsafe", "{s}");
        assert_eq!(s["result"]["blocked"]["code"], "key_file_unsafe");
        assert!(
            s["result"]["blocked"]["next"]
                .as_str()
                .unwrap()
                .starts_with(fix),
            "{s}"
        );
    };
    // Each group/other bit alone on the entry; 0600 and 0400 pass.
    for bits in [0o640, 0o620, 0o610, 0o604, 0o602, 0o601] {
        fs::set_permissions(entry(&home), fs::Permissions::from_mode(bits)).unwrap();
        unsafe_with("chmod 600 ");
    }
    for bits in [0o600, 0o400] {
        fs::set_permissions(entry(&home), fs::Permissions::from_mode(bits)).unwrap();
        let s = envelope(&run(&home, &["key-status", "glm"], None), &key);
        assert_eq!(s["result"]["entry"], "present", "{bits:o}: {s}");
    }
    fs::set_permissions(entry(&home), fs::Permissions::from_mode(0o600)).unwrap();
    // The machine key, the shared folder and the vault folder too.
    let master = vault(&home).join(".master.key");
    fs::set_permissions(&master, fs::Permissions::from_mode(0o640)).unwrap();
    unsafe_with("chmod 600 ");
    fs::set_permissions(&master, fs::Permissions::from_mode(0o600)).unwrap();
    for dir in [vault(&home).join("_shared"), vault(&home)] {
        fs::set_permissions(&dir, fs::Permissions::from_mode(0o750)).unwrap();
        unsafe_with("chmod 700 ");
        // And key-set refuses to write into it.
        let out = envelope(&run(&home, &["key-set", "glm"], Some(&key)), &key);
        assert_eq!(out["ok"], false, "{out}");
        fs::set_permissions(&dir, fs::Permissions::from_mode(0o700)).unwrap();
    }
}

#[test]
fn a_symlinked_or_special_entry_is_not_followed() {
    let (_d, home) = home();
    let key = synthetic();
    envelope(&run(&home, &["key-set", "glm"], Some(&key)), &key);
    let sealed = fs::read(entry(&home)).unwrap();
    // A symlink to a sealed copy elsewhere.
    let outside = home.join("outside.key");
    fs::write(&outside, &sealed).unwrap();
    fs::set_permissions(&outside, fs::Permissions::from_mode(0o600)).unwrap();
    fs::remove_file(entry(&home)).unwrap();
    std::os::unix::fs::symlink(&outside, entry(&home)).unwrap();
    let s = envelope(&run(&home, &["key-status", "glm"], None), &key);
    assert_eq!(s["result"]["entry"], "unsafe", "{s}");
    // A folder in place of the entry.
    fs::remove_file(entry(&home)).unwrap();
    fs::create_dir(entry(&home)).unwrap();
    let s = envelope(&run(&home, &["key-status", "glm"], None), &key);
    assert_eq!(s["result"]["entry"], "unsafe", "{s}");
    let out = envelope(&run(&home, &["key-set", "glm"], Some(&key)), &key);
    assert_eq!(out["ok"], false, "a folder is not replaced: {out}");
    // A FIFO in place of the entry: opened without blocking, refused.
    fs::remove_dir(entry(&home)).unwrap();
    let made = Command::new("mkfifo").arg(entry(&home)).status().unwrap();
    assert!(made.success());
    let s = envelope(&run(&home, &["key-status", "glm"], None), &key);
    assert_eq!(s["result"]["entry"], "unsafe", "{s}");
    // A symlinked shared folder is refused as not a real folder.
    fs::remove_file(entry(&home)).unwrap();
    fs::remove_dir(vault(&home).join("_shared")).unwrap();
    fs::create_dir(home.join("elsewhere")).unwrap();
    std::os::unix::fs::symlink(home.join("elsewhere"), vault(&home).join("_shared")).unwrap();
    let s = envelope(&run(&home, &["key-status", "glm"], None), &key);
    assert_eq!(s["result"]["entry"], "unsafe", "{s}");
    // Pre-audit B1: a removal never follows the folder link out of the vault. The file at
    // the link's target stays, with or without --yes.
    let marker = home.join("elsewhere/glm.key");
    fs::write(&marker, "marker-not-a-key\n").unwrap();
    for args in [
        &["key-remove", "glm"][..],
        &["key-remove", "glm", "--yes"][..],
    ] {
        let r = envelope(&run(&home, args, None), &key);
        assert_eq!(r["ok"], false, "{args:?}: {r}");
        assert!(marker.exists(), "{args:?} removed a file outside the vault");
    }
    assert_eq!(fs::read_to_string(&marker).unwrap(), "marker-not-a-key\n");
}

/// Pre-audit B2: a value given as an argument by mistake is never echoed — not on stdout,
/// not on stderr — for any key command.
#[test]
fn a_stray_key_argument_is_not_echoed() {
    let (_d, home) = home();
    let marker = "AUDIT_ONLY_NOT_A_REAL_SECRET";
    for args in [
        &["key-set", "glm", marker][..],
        &["key-status", "glm", marker][..],
        &["key-remove", "glm", marker][..],
        &["key-set", marker, marker][..],
    ] {
        let out = run(&home, args, None);
        assert_eq!(out.status.code(), Some(2), "{args:?}");
        let e = envelope(&out, marker);
        assert_eq!(e["ok"], false, "{args:?}: {e}");
    }
    assert!(!vault(&home).exists());
}

/// Pre-audit B1 on the new format: whatever `key-set` accepts reads back as `present`.
/// 65,535 bytes of value (a 65,583-byte entry) round-trips; 65,536 bytes is refused and
/// nothing is written. Literal sizes.
#[test]
fn every_value_key_set_accepts_reads_back() {
    let (_d, home) = home();
    let max = "k".repeat(65_535);
    let set = envelope(&run(&home, &["key-set", "glm"], Some(&max)), &max);
    assert_eq!(set["ok"], true, "{set}");
    assert_eq!(fs::metadata(entry(&home)).unwrap().len(), 65_583);
    let s = envelope(&run(&home, &["key-status", "glm"], None), &max);
    assert_eq!(s["result"]["entry"], "present", "{s}");
    fs::remove_file(entry(&home)).unwrap();
    let over = "k".repeat(65_536);
    let set = envelope(&run(&home, &["key-set", "glm"], Some(&over)), &over);
    assert_eq!(set["ok"], false, "{set}");
    assert!(!entry(&home).exists());
}

/// A plain-text entry (the format before the decision) is refused, never used as a value.
#[test]
fn a_plain_text_entry_is_refused() {
    let (_d, home) = home();
    let key = synthetic();
    envelope(&run(&home, &["key-set", "glm"], Some(&key)), &key);
    fs::write(entry(&home), format!("{key}\n")).unwrap();
    let s = envelope(&run(&home, &["key-status", "glm"], None), &key);
    assert_eq!(s["result"]["entry"], "unsafe", "{s}");
}

#[test]
fn provider_names_cannot_leave_the_vault() {
    let (_d, home) = home();
    for bad in ["../x", "a/b", "GLM", ".key", ".master"] {
        let out = run(&home, &["key-set", bad], Some("v-never"));
        let env = envelope(&out, "v-never");
        assert_eq!(env["ok"], false, "{bad}");
    }
    assert!(!home.join(".moai-cli").exists());
}

#[test]
fn a_legacy_key_file_is_never_read_or_touched() {
    let (_d, home) = home();
    let key = synthetic();
    for legacy in [".moai/.env.glm", ".moai-cli/.env.glm"] {
        let p = home.join(legacy);
        fs::create_dir_all(p.parent().unwrap()).unwrap();
        fs::write(&p, format!("ANTHROPIC_AUTH_TOKEN={key}\n")).unwrap();
    }
    let before = fs::metadata(home.join(".moai/.env.glm"))
        .unwrap()
        .modified()
        .unwrap();
    let s = envelope(&run(&home, &["key-status", "glm"], None), &key);
    assert_eq!(s["result"]["entry"], "absent", "{s}");
    envelope(&run(&home, &["key-remove", "glm", "--yes"], None), &key);
    assert!(home.join(".moai/.env.glm").exists());
    assert!(home.join(".moai-cli/.env.glm").exists());
    assert_eq!(
        fs::metadata(home.join(".moai/.env.glm"))
            .unwrap()
            .modified()
            .unwrap(),
        before
    );
}

/// The key commands left `src/main.rs` for `src/ops/` (#2834 R3): each envelope still names
/// its own operation.
#[test]
fn each_key_command_answers_under_its_operation_name() {
    let (_d, home) = home();
    let key = synthetic();
    let names = [
        (vec!["key-status", "glm"], None, "key.status"),
        (vec!["key-set", "glm"], Some(key.as_str()), "key.set"),
        (vec!["key-remove", "glm"], None, "key.remove"),
        (vec!["key-import-plan", "glm"], None, "key.import_plan"),
        (
            vec![
                "key-import-apply",
                "glm",
                "--plan-digest",
                "x",
                "--permission-reference",
                "r",
            ],
            None,
            "key.import_apply",
        ),
    ];
    for (args, stdin, op) in names {
        let v = envelope(&run(&home, &args, stdin), &key);
        assert_eq!(v["operation"], op, "{v}");
    }
}
