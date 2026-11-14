//! #2871 I2 S3: the `keys` layer of `moai update`. A legacy `.env.glm` (the Go `~/.moai/` one or
//! the earlier `~/.moai-cli/` one) is offered as an import into the vault's `_shared` folder,
//! inside the same preview and `plan_digest` as the rest of the user layer. The confirmed apply
//! copies the key, keeps the original file as it was and writes a receipt; `keys-revert` removes
//! only the entry the layer imported. Unsafe sources (symlink, FIFO, folder, over 64 KiB, mode
//! looser than 0600, unknown provider) are kept and never read. The value is never shown.
//! Everything runs in a temporary HOME with dummy values (leader survey R8, PLAN Q5–Q7).

#![cfg(unix)]

mod common;
use common::CleanEnv;

use serde_json::{json, Value};
use std::{
    fs,
    os::unix::fs::{MetadataExt, PermissionsExt},
    path::{Path, PathBuf},
    process::{Command, Output, Stdio},
};
use tempfile::TempDir;

const EXE: &str = crate::common::MOAI_BIN;
const KEY: &str = "DUMMY-GLM-KEY-123";
const OTHER_KEY: &str = "DUMMY-OPENAI-KEY-456";
const NEW_KEY: &str = "DUMMY-GLM-KEY-SET-LATER-789";
const GO_GLM: &str = ".moai/.env.glm";

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
    /// A plain folder outside HOME the commands run from (no git work tree: no hook layer).
    work: PathBuf,
}

/// A temporary HOME with the binary, the real package and the agent sources in a version folder.
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
    let work = root.join("work");
    fs::create_dir_all(&work).unwrap();
    Fixture {
        _tmp: tmp,
        home,
        exe,
        work,
    }
}

/// A private key file under HOME (`0600`, the folders made as needed).
fn key_file(f: &Fixture, rel: &str, text: &str) -> PathBuf {
    let path = f.home.join(rel);
    fs::create_dir_all(path.parent().unwrap()).unwrap();
    fs::write(&path, text).unwrap();
    fs::set_permissions(&path, fs::Permissions::from_mode(0o600)).unwrap();
    path
}

fn glm_text(key: &str) -> String {
    format!("GLM_API_KEY={key}\n")
}

fn command(f: &Fixture, cwd: &Path, args: &[&str]) -> Command {
    let mut c = Command::new(&f.exe);
    c.args(args)
        .clean_env()
        .env("PATH", "/usr/bin:/bin")
        .env("HOME", &f.home)
        .env("GIT_CONFIG_GLOBAL", "/dev/null")
        .env("GIT_CONFIG_SYSTEM", "/dev/null")
        .env("GIT_CEILING_DIRECTORIES", cwd.parent().unwrap())
        .env("LC_ALL", "C")
        .current_dir(cwd);
    c
}

fn run(f: &Fixture, args: &[&str]) -> Output {
    command(f, &f.work, args)
        .stdin(Stdio::null())
        .output()
        .unwrap()
}

fn text_of(out: &Output) -> String {
    format!(
        "{}{}",
        String::from_utf8_lossy(&out.stdout),
        String::from_utf8_lossy(&out.stderr)
    )
}

fn result(out: &Output) -> Value {
    let v: Value = serde_json::from_slice(&out.stdout)
        .unwrap_or_else(|_| panic!("not JSON: {}", text_of(out)));
    assert_eq!(v["ok"], json!(true), "{v}");
    v["result"].clone()
}

fn preview(f: &Fixture) -> Value {
    result(&run(f, &["update"]))
}

fn confirm(f: &Fixture, digest: &str) -> Value {
    result(&run(f, &["update", "--confirm", digest]))
}

fn rows(p: &Value) -> Vec<Value> {
    p["keys"]["rows"]
        .as_array()
        .unwrap_or_else(|| panic!("no keys rows in {p}"))
        .clone()
}

fn row_for(p: &Value, source: &Path) -> Value {
    rows(p)
        .into_iter()
        .find(|r| r["source"] == json!(source.display().to_string()))
        .unwrap_or_else(|| panic!("no row for {} in {}", source.display(), p["keys"]))
}

fn shared_entry(f: &Fixture) -> PathBuf {
    f.home.join(".moai-cli/vault/_shared/glm.key")
}

fn revert(f: &Fixture, receipt: &str) -> Output {
    run(
        f,
        &[
            "keys-revert",
            f.home.to_str().unwrap(),
            "--receipt",
            receipt,
            "--permission-reference",
            "test-approval",
        ],
    )
}

/// Every file under `dir` with its bytes and mode (a snapshot that shows nothing changed).
fn snapshot(dir: &Path) -> Vec<(PathBuf, Vec<u8>, u32)> {
    let mut out = Vec::new();
    let mut stack = vec![dir.to_path_buf()];
    while let Some(p) = stack.pop() {
        let Ok(m) = fs::symlink_metadata(&p) else {
            continue;
        };
        if m.is_dir() {
            out.push((p.clone(), Vec::new(), m.mode()));
            for e in fs::read_dir(&p).unwrap() {
                stack.push(e.unwrap().path());
            }
        } else if m.is_file() {
            out.push((p.clone(), fs::read(&p).unwrap(), m.mode()));
        } else {
            out.push((p.clone(), Vec::new(), m.mode()));
        }
    }
    out.sort();
    out
}

/// Every receipt file the keys layer wrote, as text.
fn receipts_text(f: &Fixture) -> String {
    let dir = f.home.join(".moai-cli/receipts");
    let mut text = String::new();
    for e in fs::read_dir(&dir).into_iter().flatten() {
        let p = e.unwrap().path();
        text.push_str(&fs::read_to_string(&p).unwrap());
    }
    text
}

/// AC: a Go-era `~/.moai/.env.glm` is offered as `import` into the vault, and another provider's
/// file is reported as `keep` with its reason; the layer's digest is in the update plan.
#[test]
fn keys_layer_offers_import_for_legacy_glm() {
    let f = setup();
    let glm = key_file(&f, GO_GLM, &glm_text(KEY));
    let other = key_file(
        &f,
        ".moai/.env.openai",
        &format!("OPENAI_API_KEY={OTHER_KEY}\n"),
    );
    let p = preview(&f);
    assert_eq!(p["nothing_to_apply"], json!(false), "{p}");
    let r = row_for(&p, &glm);
    assert_eq!(r["provider"], json!("glm"), "{r}");
    assert_eq!(r["action"], json!("import"), "{r}");
    let o = row_for(&p, &other);
    assert_eq!(o["provider"], json!("openai"), "{o}");
    assert_eq!(o["action"], json!("keep"), "{o}");
    assert!(
        o["reason"].as_str().unwrap().contains("unknown provider"),
        "{o}"
    );
    assert!(p["keys"]["plan_digest"].is_string(), "{p}");
    assert_eq!(
        p["keys"]["destination"],
        json!(shared_entry(&f).display().to_string())
    );
    assert_eq!(p["keys"]["value_shown"], json!(false));
}

/// AC: the value appears nowhere — not on stdout or stderr of the preview, the apply, the revert
/// or the human summary, not in the JSON, not in the receipt.
#[test]
fn keys_layer_never_prints_the_value() {
    let f = setup();
    key_file(&f, GO_GLM, &glm_text(KEY));
    key_file(
        &f,
        ".moai/.env.openai",
        &format!("OPENAI_API_KEY={OTHER_KEY}\n"),
    );
    let pre = run(&f, &["update"]);
    let p = result(&pre);
    let digest = p["plan_digest"].as_str().unwrap().to_string();
    let app = run(&f, &["update", "--confirm", &digest]);
    let r = result(&app);
    assert_eq!(r["keys_result"]["state"], json!("applied"), "{r}");
    let id = r["keys_result"]["receipt"].as_str().unwrap().to_string();
    let rev = revert(&f, &id);
    result(&rev);
    let receipts = receipts_text(&f);
    assert!(receipts.contains("update.keys"), "{receipts}");
    for (what, text) in [
        ("preview", text_of(&pre)),
        ("apply", text_of(&app)),
        ("revert", text_of(&rev)),
        ("receipt", receipts),
    ] {
        for value in [KEY, OTHER_KEY] {
            assert!(!text.contains(value), "{what} shows the value: {text}");
        }
    }
    let sealed = fs::read(shared_entry(&f)).unwrap_or_default();
    assert!(!sealed.windows(KEY.len()).any(|w| w == KEY.as_bytes()));
}

/// AC: a preview writes nothing anywhere under HOME, the vault included.
#[test]
fn keys_preview_writes_nothing() {
    let f = setup();
    key_file(&f, GO_GLM, &glm_text(KEY));
    key_file(&f, ".moai-cli/.env.glm", &glm_text(OTHER_KEY));
    let before = snapshot(&f.home);
    let p = preview(&f);
    assert_eq!(row_for(&p, &f.home.join(GO_GLM))["action"], json!("import"));
    assert_eq!(snapshot(&f.home), before, "the preview wrote under HOME");
    assert!(!f.home.join(".moai-cli/vault").exists());
}

/// AC: the confirmed apply copies the key into the vault (sealed, `0600` in `0700` folders,
/// readable back), keeps the original byte-, mode- and time-identical, and records a receipt.
/// The second legacy file is kept, not imported over the first.
#[test]
fn keys_apply_copies_and_keeps_original() {
    let f = setup();
    let glm = key_file(&f, GO_GLM, &glm_text(KEY));
    let rust = key_file(&f, ".moai-cli/.env.glm", &glm_text(OTHER_KEY));
    let before = fs::symlink_metadata(&glm).unwrap();
    let p = preview(&f);
    assert_eq!(row_for(&p, &glm)["action"], json!("import"), "{p}");
    let second = row_for(&p, &rust);
    assert_eq!(second["action"], json!("keep"), "{second}");
    let r = confirm(&f, p["plan_digest"].as_str().unwrap());
    assert_eq!(r["applied"], json!(true), "{r}");
    let k = &r["keys_result"];
    assert_eq!(k["state"], json!("applied"), "{r}");
    assert_eq!(k["original_kept"], json!(true), "{k}");
    assert_eq!(
        k["written"],
        json!([shared_entry(&f).display().to_string()])
    );
    assert!(r["revert"]["keys"]
        .as_str()
        .unwrap()
        .contains(k["receipt"].as_str().unwrap()));
    // The original is untouched.
    assert_eq!(fs::read_to_string(&glm).unwrap(), glm_text(KEY));
    let after = fs::symlink_metadata(&glm).unwrap();
    assert_eq!(
        (after.ino(), after.mode(), after.mtime(), after.mtime_nsec()),
        (
            before.ino(),
            before.mode(),
            before.mtime(),
            before.mtime_nsec()
        )
    );
    assert_eq!(fs::read_to_string(&rust).unwrap(), glm_text(OTHER_KEY));
    // The entry is sealed, private, and holds the first file's key.
    let entry = shared_entry(&f);
    assert!(fs::read(&entry).unwrap().starts_with(b"MOAIKV1\n"));
    assert_eq!(fs::metadata(&entry).unwrap().mode() & 0o777, 0o600);
    for dir in [
        entry.parent().unwrap(),
        entry.parent().unwrap().parent().unwrap(),
    ] {
        assert_eq!(fs::metadata(dir).unwrap().mode() & 0o777, 0o700, "{dir:?}");
    }
    let status = result(&run(&f, &["key-status", "glm", "--shared"]));
    assert_eq!(status["entry"], json!("present"), "{status}");
}

/// AC: `keys-revert` removes the entry the layer imported and never the original; an entry
/// replaced after the update (a later `key-set`) is not the layer's and stays; a receipt reverts
/// once.
#[test]
fn keys_revert_removes_imported_entry() {
    let f = setup();
    let glm = key_file(&f, GO_GLM, &glm_text(KEY));
    let p = preview(&f);
    let r = confirm(&f, p["plan_digest"].as_str().unwrap());
    let id = r["keys_result"]["receipt"].as_str().unwrap().to_string();
    assert!(shared_entry(&f).is_file());
    let v = result(&revert(&f, &id));
    assert_eq!(v["state"], json!("reverted"), "{v}");
    assert_eq!(
        v["reverted"],
        json!([shared_entry(&f).display().to_string()]),
        "{v}"
    );
    assert!(!shared_entry(&f).exists(), "the imported entry stayed");
    assert_eq!(fs::read_to_string(&glm).unwrap(), glm_text(KEY));
    assert!(
        !revert(&f, &id).status.success(),
        "a receipt reverted twice"
    );

    // Import again, then replace the entry with `key-set`: the revert leaves it.
    let p = preview(&f);
    assert_eq!(row_for(&p, &glm)["action"], json!("import"), "{p}");
    let r = confirm(&f, p["plan_digest"].as_str().unwrap());
    let id = r["keys_result"]["receipt"].as_str().unwrap().to_string();
    let mut child = command(&f, &f.work, &["key-set", "glm", "--shared"])
        .stdin(Stdio::piped())
        .stdout(Stdio::piped())
        .stderr(Stdio::piped())
        .spawn()
        .unwrap();
    common::feed_stdin(&mut child, NEW_KEY.as_bytes());
    let set = child.wait_with_output().unwrap();
    result(&set);
    let replaced = fs::read(shared_entry(&f)).unwrap();
    let v = result(&revert(&f, &id));
    assert_eq!(v["reverted"], json!([]), "{v}");
    assert_eq!(
        v["skipped"][0]["path"],
        json!(shared_entry(&f).display().to_string()),
        "{v}"
    );
    assert_eq!(fs::read(shared_entry(&f)).unwrap(), replaced);
    assert_eq!(fs::read_to_string(&glm).unwrap(), glm_text(KEY));
}

/// AC: each unsafe source is `keep` with its reason and is never imported: a symlink, a FIFO, a
/// folder, a file over 64 KiB, a file with a mode looser than `0600`, a file of an unknown
/// provider. The edges pass: exactly 64 KiB and mode `0600`.
#[test]
fn keys_layer_refuses_unsafe_sources() {
    let mut big = glm_text(KEY).into_bytes();
    big.push(b'#');
    big.resize(65_537, b'x');
    let mut edge = big.clone();
    edge.truncate(65_536);
    type Make = Box<dyn Fn(&Fixture, &Path)>;
    let cases: Vec<(&str, &str, Make)> = vec![
        (
            GO_GLM,
            "symlink",
            Box::new(|f: &Fixture, p: &Path| {
                let target = key_file(f, "elsewhere/glm", &glm_text(KEY));
                std::os::unix::fs::symlink(target, p).unwrap();
            }),
        ),
        (
            GO_GLM,
            "not a regular file",
            Box::new(|_: &Fixture, p: &Path| {
                let ok = Command::new("mkfifo").arg(p).status().unwrap();
                assert!(ok.success());
            }),
        ),
        (
            GO_GLM,
            "directory",
            Box::new(|_: &Fixture, p: &Path| fs::create_dir(p).unwrap()),
        ),
        (
            GO_GLM,
            "64 KiB",
            Box::new(move |_: &Fixture, p: &Path| {
                fs::write(p, &big).unwrap();
                fs::set_permissions(p, fs::Permissions::from_mode(0o600)).unwrap();
            }),
        ),
        (
            GO_GLM,
            "0600",
            Box::new(|_: &Fixture, p: &Path| {
                fs::write(p, glm_text(KEY)).unwrap();
                fs::set_permissions(p, fs::Permissions::from_mode(0o640)).unwrap();
            }),
        ),
        (
            GO_GLM,
            "0600",
            Box::new(|_: &Fixture, p: &Path| {
                fs::write(p, glm_text(KEY)).unwrap();
                fs::set_permissions(p, fs::Permissions::from_mode(0o604)).unwrap();
            }),
        ),
        (
            // Not `.env.GLM`: on a case-insensitive file system that is `.env.glm` itself.
            ".moai/.env.openai",
            "unknown provider",
            Box::new(|_: &Fixture, p: &Path| {
                fs::write(p, glm_text(KEY)).unwrap();
                fs::set_permissions(p, fs::Permissions::from_mode(0o600)).unwrap();
            }),
        ),
    ];
    for (rel, reason, make) in &cases {
        let f = setup();
        let path = f.home.join(rel);
        fs::create_dir_all(path.parent().unwrap()).unwrap();
        make(&f, &path);
        let p = preview(&f);
        let r = row_for(&p, &path);
        assert_eq!(r["action"], json!("keep"), "{rel} {reason}: {r}");
        assert!(
            r["reason"].as_str().unwrap().contains(reason),
            "{rel}: want {reason}: {r}"
        );
        assert!(
            rows(&p).iter().all(|r| r["action"] != json!("import")),
            "{rel} {reason}: {p}"
        );
        assert!(!text_of(&run(&f, &["update"])).contains(KEY));
    }
    // A kept source is not imported by a confirmed run either.
    let f = setup();
    let loose = f.home.join(GO_GLM);
    fs::create_dir_all(loose.parent().unwrap()).unwrap();
    fs::write(&loose, glm_text(KEY)).unwrap();
    fs::set_permissions(&loose, fs::Permissions::from_mode(0o644)).unwrap();
    let p = preview(&f);
    let r = confirm(&f, p["plan_digest"].as_str().unwrap());
    assert!(!shared_entry(&f).exists(), "{r}");
    // The edges: exactly 64 KiB with mode 0600 is imported.
    let f = setup();
    let at = f.home.join(GO_GLM);
    fs::create_dir_all(at.parent().unwrap()).unwrap();
    fs::write(&at, &edge).unwrap();
    fs::set_permissions(&at, fs::Permissions::from_mode(0o600)).unwrap();
    assert_eq!(fs::metadata(&at).unwrap().len(), 65_536);
    let p = preview(&f);
    assert_eq!(row_for(&p, &at)["action"], json!("import"), "{p}");
}

/// AC: after the apply, the next preview has nothing for the keys layer to do: every glm source
/// is `already_in_vault` and the run is up to date.
#[test]
fn keys_layer_rerun_is_empty() {
    let f = setup();
    let glm = key_file(&f, GO_GLM, &glm_text(KEY));
    let p = preview(&f);
    confirm(&f, p["plan_digest"].as_str().unwrap());
    let again = preview(&f);
    assert_eq!(again["nothing_to_apply"], json!(true), "{again}");
    assert_eq!(
        row_for(&again, &glm)["action"],
        json!("already_in_vault"),
        "{again}"
    );
    assert!(rows(&again).iter().all(|r| r["action"] != json!("import")));
    // Without any legacy file the layer has no rows and the preview has no `keys` part.
    let empty = setup();
    let p = preview(&empty);
    assert!(p.get("keys").is_none(), "{p}");
}

/// AC: the keys layer is inside the update `plan_digest`: a source changed after the preview
/// makes the old digest stale, and nothing is imported.
#[test]
fn keys_layer_is_inside_the_plan_digest() {
    let f = setup();
    let glm = key_file(&f, GO_GLM, &glm_text(KEY));
    let p = preview(&f);
    let digest = p["plan_digest"].as_str().unwrap().to_string();
    fs::write(&glm, glm_text("DUMMY-GLM-KEY-123-CHANGED")).unwrap();
    let out = run(&f, &["update", "--confirm", &digest]);
    assert!(!out.status.success(), "{}", text_of(&out));
    assert!(text_of(&out).contains("stale"), "{}", text_of(&out));
    assert!(!shared_entry(&f).exists());
}

/// AC: with only the legacy file `glm` stops with `legacy_key_only`; after the keys layer the
/// launch takes the key from the vault's shared folder.
#[test]
fn glm_launch_unblocked_after_keys_layer() {
    let f = setup();
    key_file(&f, GO_GLM, &glm_text(KEY));
    let project = f.work.join("project");
    fs::create_dir_all(&project).unwrap();
    common::git::git_ok(&project, &["init", "-q"]);
    let dry = |f: &Fixture| {
        command(f, &project, &["glm", "-a", "--dry-run"])
            .stdin(Stdio::null())
            .output()
            .unwrap()
    };
    let before = dry(&f);
    assert!(!before.status.success(), "{}", text_of(&before));
    assert!(
        text_of(&before).contains("legacy_key_only"),
        "{}",
        text_of(&before)
    );
    let p = preview(&f);
    confirm(&f, p["plan_digest"].as_str().unwrap());
    let after = dry(&f);
    assert!(after.status.success(), "{}", text_of(&after));
    let v: Value = serde_json::from_slice(&after.stdout).unwrap();
    assert_eq!(v["key_source"], json!("vault:shared:glm"), "{v}");
    assert!(!text_of(&after).contains(KEY));
}
