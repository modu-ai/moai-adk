//! #2935 L1: `language-set --home <code>` records the conversation language the installer
//! asked for in the HOME state folder, once and never over a value that is already there, and
//! `language-setup` reads that value as the conversation default after `--conversation` and the
//! profile's `user_chat`. Every case runs the built binary inside a temporary HOME; the real
//! HOME is never read or written.

#![cfg(unix)]

mod common;

use common::CleanEnv as _;
use serde_json::{json, Value};
use std::{
    fs,
    os::unix::fs::symlink,
    path::{Path, PathBuf},
    process::{Command, Stdio},
};
use tempfile::TempDir;

const EXE: &str = common::MOAI_BIN;
/// `language::HOME_FILE_LIMIT` as a literal, so a changed bound fails here.
const LIMIT: usize = 4096;
const FILE: &str = "config/language.yaml";

fn home() -> (TempDir, PathBuf) {
    let d = TempDir::new().unwrap();
    let h = fs::canonicalize(d.path()).unwrap();
    (d, h)
}

/// A HOME that already has the earlier state folder `~/.moai-cli`.
fn home_with_state() -> (TempDir, PathBuf) {
    let (d, h) = home();
    fs::create_dir(h.join(".moai-cli")).unwrap();
    (d, h)
}

fn call(home: &Path, args: &[&str]) -> (i32, Value, String) {
    let out = Command::new(EXE)
        .args(args)
        .clean_env()
        .env("HOME", home)
        .env("PATH", "/usr/bin:/bin")
        .env("GIT_CONFIG_GLOBAL", "/dev/null")
        .env("GIT_CONFIG_SYSTEM", "/dev/null")
        .current_dir(home)
        .stdin(Stdio::null())
        .output()
        .unwrap();
    let json = serde_json::from_slice(&out.stdout).unwrap_or(Value::Null);
    (
        out.status.code().unwrap(),
        json,
        String::from_utf8_lossy(&out.stderr).into_owned(),
    )
}

fn set(home: &Path, code: &str) -> (i32, Value) {
    let (status, json, _) = call(home, &["language-set", "--home", code]);
    (status, json)
}

/// Every entry below `root` with its kind and bytes (a link with its target).
fn tree(root: &Path) -> Vec<(String, Vec<u8>)> {
    let mut out = Vec::new();
    let mut stack = vec![root.to_path_buf()];
    while let Some(d) = stack.pop() {
        for e in fs::read_dir(&d).unwrap().flatten() {
            let p = e.path();
            let rel = p.strip_prefix(root).unwrap().display().to_string();
            let m = fs::symlink_metadata(&p).unwrap();
            if m.file_type().is_symlink() {
                let target = fs::read_link(&p).unwrap().to_string_lossy().into_owned();
                out.push((format!("{rel}@"), target.into_bytes()));
            } else if m.is_dir() {
                out.push((format!("{rel}/"), Vec::new()));
                stack.push(p);
            } else {
                out.push((rel, fs::read(&p).unwrap()));
            }
        }
    }
    out.sort();
    out
}

fn result(json: &Value) -> &Value {
    assert_eq!(json["ok"], json!(true), "{json}");
    assert_eq!(json["operation"], json!("language.set"), "{json}");
    &json["result"]
}

/// First write: the folder `config/` and one file with exactly the two lines, nothing else.
#[test]
fn the_first_write_creates_config_and_one_file() {
    let (_d, h) = home_with_state();
    let (status, json) = set(&h, "ko");
    assert_eq!(status, 0, "{json}");
    let r = result(&json);
    assert_eq!(r["written"], json!(true), "{json}");
    assert_eq!(r["code"], json!("ko"));
    assert_eq!(r["reason"], Value::Null);
    let file = h.join(".moai-cli").join(FILE);
    assert_eq!(r["path"], json!(file.to_string_lossy()));
    assert_eq!(
        fs::read(&file).unwrap(),
        b"language:\n  conversation_language: ko\n"
    );
    let names: Vec<String> = tree(&h).into_iter().map(|(n, _)| n).collect();
    assert_eq!(
        names,
        vec![
            ".moai-cli/".to_string(),
            ".moai-cli/config/".to_string(),
            format!(".moai-cli/{FILE}")
        ],
        "only config/ and the file are new; no staging file is left"
    );
}

/// A second run, with the same or another code, writes nothing and keeps the bytes.
#[test]
fn a_second_run_writes_nothing() {
    let (_d, h) = home_with_state();
    assert_eq!(set(&h, "ko").0, 0);
    let before = tree(&h);
    for code in ["ko", "ja", "en"] {
        let (status, json) = set(&h, code);
        assert_eq!(status, 0, "{json}");
        let r = result(&json);
        assert_eq!(r["written"], json!(false), "{json}");
        assert_eq!(r["reason"], json!("exists"), "{json}");
        assert_eq!(r["code"], json!(code));
        assert_eq!(
            tree(&h),
            before,
            "bytes and tree are untouched after {code}"
        );
    }
}

/// A file the user wrote by hand (even in another shape) is never replaced.
#[test]
fn a_file_the_user_already_has_is_kept_byte_for_byte() {
    let (_d, h) = home_with_state();
    fs::create_dir(h.join(".moai-cli/config")).unwrap();
    let mine = b"# mine\nlanguage:\n  conversation_language: zh\n  code_comments: en\n";
    fs::write(h.join(".moai-cli").join(FILE), mine).unwrap();
    let before = tree(&h);
    let (status, json) = set(&h, "ja");
    assert_eq!(status, 0, "{json}");
    assert_eq!(result(&json)["reason"], json!("exists"), "{json}");
    assert_eq!(tree(&h), before);
    // An empty file is a file too.
    fs::write(h.join(".moai-cli").join(FILE), b"").unwrap();
    let (status, json) = set(&h, "ja");
    assert_eq!(status, 0, "{json}");
    assert_eq!(result(&json)["written"], json!(false), "{json}");
    assert_eq!(fs::read(h.join(".moai-cli").join(FILE)).unwrap(), b"");
}

/// A HOME whose state folder is `~/.moai` (valid marker) is written there, not in `~/.moai-cli`.
#[test]
fn a_moai_state_folder_with_the_marker_is_the_one_written() {
    let (_d, h) = home();
    fs::create_dir(h.join(".moai")).unwrap();
    fs::write(
        h.join(".moai/moai-cli.json"),
        r#"{"schema_version":1,"generation":"moai-cli"}"#,
    )
    .unwrap();
    let (status, json) = set(&h, "ja");
    assert_eq!(status, 0, "{json}");
    assert_eq!(result(&json)["written"], json!(true), "{json}");
    assert_eq!(
        fs::read(h.join(".moai").join(FILE)).unwrap(),
        b"language:\n  conversation_language: ja\n"
    );
    assert!(!h.join(".moai-cli").exists());
}

/// No state folder: nothing is written and no folder is created (the installer only informs).
#[test]
fn a_home_without_a_state_folder_is_not_written() {
    let (_d, h) = home();
    let (status, json) = set(&h, "ko");
    assert_eq!(status, 0, "{json}");
    let r = result(&json);
    assert_eq!(r["written"], json!(false), "{json}");
    assert_eq!(r["reason"], json!("no_home_state"), "{json}");
    assert_eq!(r["code"], json!("ko"));
    assert!(tree(&h).is_empty(), "no folder was created: {:?}", tree(&h));
}

/// The Go generation's `~/.moai` is not this generation's: it is never written into.
#[test]
fn the_go_generations_moai_folder_is_never_written() {
    let (_d, h) = home();
    fs::create_dir_all(h.join(".moai/state")).unwrap();
    fs::write(h.join(".moai/state/active-sessions.json"), "{}").unwrap();
    let before = tree(&h);
    let (status, json) = set(&h, "ko");
    assert_eq!(status, 0, "{json}");
    assert_eq!(result(&json)["reason"], json!("no_home_state"), "{json}");
    assert_eq!(tree(&h), before);
}

/// A dangling link where the state folder would be is not a folder either: nothing is created
/// through it.
#[test]
fn a_dangling_state_folder_link_creates_nothing() {
    let (_d, h) = home();
    symlink(h.join("nowhere"), h.join(".moai-cli")).unwrap();
    let before = tree(&h);
    let (status, json) = set(&h, "ko");
    assert_eq!(status, 0, "{json}");
    assert_eq!(result(&json)["written"], json!(false), "{json}");
    assert_eq!(tree(&h), before);
    assert!(!h.join("nowhere").exists());
}

/// The longest code (35 bytes) and the first too long (36).
fn longest() -> String {
    format!("ko{}-abcde", "-abcdefgh".repeat(3))
}

/// Any language-code-shaped text is stored, exactly as given: the grammar's last passing value
/// at each edge (2 and 3 letters, a 2- and an 8-character group, 35 bytes in all) and the
/// codes MoAI has names for. The file holds the two lines and nothing else.
#[test]
fn a_code_of_the_right_shape_is_stored_exactly_as_given() {
    let longest = longest();
    assert_eq!(longest.len(), 35);
    for code in [
        "en",
        "ko",
        "ja",
        "zh",
        "fr",
        "kor",
        "KO",
        "pt-BR",
        "es-419",
        "ko-ab",
        "ko-abcdefgh",
        "zh-Hant-TW",
        longest.as_str(),
    ] {
        let (_d, h) = home_with_state();
        let (status, json) = set(&h, code);
        assert_eq!(status, 0, "{code}: {json}");
        let r = result(&json);
        assert_eq!(r["written"], json!(true), "{code}");
        assert_eq!(r["code"], json!(code));
        assert_eq!(
            fs::read(h.join(".moai-cli").join(FILE)).unwrap(),
            format!("language:\n  conversation_language: {code}\n").into_bytes(),
            "{code}: written as given"
        );
    }
}

/// The first value past each edge is a usage error (exit 2) and the whole tree stays as it was:
/// one letter, four letters, a trailing or doubled `-`, a one- or nine-character group, 36
/// bytes in all, and anything outside ASCII letters, digits and `-`.
#[test]
fn a_code_of_the_wrong_shape_changes_nothing() {
    let too_long = format!("{}f", longest());
    assert_eq!(too_long.len(), 36);
    let bad = [
        "",
        "e",
        "k",
        "abcd",
        "ko-",
        "-ko",
        "ko--KR",
        "ko-a",
        "ko-abcdefghi",
        "ko_KR",
        "ko_",
        "ko ",
        " ko",
        "ko\n",
        "ko\t",
        "ko'",
        "ko\"",
        "ko\0",
        "k\u{6f}\u{301}",
        "한",
        "ko-한국",
        "../ko",
        "ko/../..",
        "ko;",
        "$HOME",
        "~",
        "-",
        "--home",
        too_long.as_str(),
    ];
    for code in bad {
        let (_d, h) = home_with_state();
        let before = tree(&h);
        // `ko\0` cannot be an argument at all; the others run.
        if code.contains('\0') {
            continue;
        }
        let (status, json, _) = call(&h, &["language-set", "--home", code]);
        assert_eq!(status, 2, "{code:?}: {json}");
        assert_eq!(json["ok"], json!(false), "{code:?}: {json}");
        assert_eq!(json["error"]["kind"], json!("usage"), "{code:?}: {json}");
        assert_eq!(tree(&h), before, "{code:?}: the tree is unchanged");
        // A HOME with no state folder stays untouched too.
        let (_e, empty) = home();
        let (status, _, _) = call(&empty, &["language-set", "--home", code]);
        assert_eq!(status, 2, "{code:?}");
        assert!(tree(&empty).is_empty());
    }
}

/// The scope flag and the code are both required; a missing code is a usage error (exit 2), a
/// missing scope flag a refusal (exit 1); neither writes.
#[test]
fn the_scope_flag_and_the_code_are_required() {
    let (_d, h) = home_with_state();
    let before = tree(&h);
    for args in [
        vec!["language-set", "--home"],
        vec!["language-set"],
        vec!["language-set", "--home", "ko", "ja"],
    ] {
        let (status, json, _) = call(&h, &args);
        assert_eq!(status, 2, "{args:?}: {json}");
        assert_eq!(tree(&h), before, "{args:?}");
    }
    let (status, json, _) = call(&h, &["language-set", "ko"]);
    assert_eq!(status, 1, "{json}");
    assert_eq!(json["ok"], json!(false), "{json}");
    assert_eq!(json["error"]["kind"], json!("invalid_input"), "{json}");
    assert_eq!(tree(&h), before, "no --home, no write");
}

/// `config` that is a link (to a real folder) or a plain file is refused; nothing is written
/// through the link or over the file.
#[test]
fn a_config_that_is_a_link_or_a_file_is_refused() {
    // A link to a real folder elsewhere.
    let (_d, h) = home_with_state();
    let elsewhere = h.join("elsewhere");
    fs::create_dir(&elsewhere).unwrap();
    symlink(&elsewhere, h.join(".moai-cli/config")).unwrap();
    let before = tree(&h);
    let (status, json) = set(&h, "ko");
    assert_eq!(status, 1, "{json}");
    assert_eq!(json["ok"], json!(false), "{json}");
    assert_eq!(json["error"]["kind"], json!("invalid_input"), "{json}");
    assert_eq!(tree(&h), before);
    assert!(!elsewhere.join("language.yaml").exists());
    // A dangling link.
    let (_d, h) = home_with_state();
    symlink(h.join("nowhere"), h.join(".moai-cli/config")).unwrap();
    let before = tree(&h);
    assert_eq!(set(&h, "ko").0, 1);
    assert_eq!(tree(&h), before);
    // A plain file.
    let (_d, h) = home_with_state();
    fs::write(h.join(".moai-cli/config"), "not a folder").unwrap();
    let before = tree(&h);
    let (status, json) = set(&h, "ko");
    assert_eq!(status, 1, "{json}");
    assert_eq!(tree(&h), before);
}

/// The target itself as a link (to a file, dangling) or a folder is refused and left alone.
#[test]
fn a_target_that_is_a_link_or_a_folder_is_refused() {
    for shape in ["file-link", "dangling-link", "folder"] {
        let (_d, h) = home_with_state();
        let config = h.join(".moai-cli/config");
        fs::create_dir(&config).unwrap();
        fs::write(
            h.join("real.yaml"),
            "language:\n  conversation_language: zh\n",
        )
        .unwrap();
        match shape {
            "file-link" => symlink(h.join("real.yaml"), config.join("language.yaml")).unwrap(),
            "dangling-link" => symlink(h.join("nowhere"), config.join("language.yaml")).unwrap(),
            _ => fs::create_dir(config.join("language.yaml")).unwrap(),
        }
        let before = tree(&h);
        let (status, json) = set(&h, "ko");
        assert_eq!(status, 1, "{shape}: {json}");
        assert_eq!(json["ok"], json!(false), "{shape}: {json}");
        assert_eq!(tree(&h), before, "{shape}");
        assert!(!h.join("nowhere").exists(), "{shape}");
    }
}

/// A state folder that is a plain file is refused.
#[test]
fn a_state_folder_that_is_a_file_is_refused() {
    let (_d, h) = home();
    fs::write(h.join(".moai-cli"), "x").unwrap();
    let before = tree(&h);
    let (status, json) = set(&h, "ko");
    assert_eq!(status, 1, "{json}");
    assert_eq!(tree(&h), before);
}

// ---- the read side: `language-setup` --------------------------------------------------

/// A project with a profile; `user_chat` optional.
fn project(user_chat: Option<&str>) -> (TempDir, PathBuf) {
    let d = TempDir::new().unwrap();
    let root = fs::canonicalize(d.path()).unwrap();
    fs::create_dir_all(root.join(".moai-cli")).unwrap();
    let mut profile = json!({"schema_version":2,"project_id":"demo","purpose":"Build a small editor","tools":["Rust"],"development_methods":["TDD"],"hosts":["claude","codex"]});
    if let Some(code) = user_chat {
        profile["language"] = json!({"user_chat": code});
    }
    fs::write(
        root.join(".moai-cli/project.yaml"),
        serde_json::to_vec_pretty(&profile).unwrap(),
    )
    .unwrap();
    (d, root)
}

/// The conversation language `language-setup --questions` resolves in `project` under `home`.
fn conversation(home: &Path, project: &Path, extra: &[&str]) -> String {
    let mut args = vec!["language-setup", project.to_str().unwrap(), "--questions"];
    args.extend_from_slice(extra);
    let (status, json, err) = call(home, &args);
    assert_eq!(status, 0, "{json}{err}");
    json["result"]["questions"]["conversation_language"]
        .as_str()
        .unwrap()
        .to_owned()
}

fn home_file(h: &Path, bytes: &[u8]) {
    fs::create_dir_all(h.join(".moai-cli/config")).unwrap();
    fs::write(h.join(".moai-cli").join(FILE), bytes).unwrap();
}

/// Order: `--conversation`, the profile's `user_chat`, the HOME value, `en`.
#[test]
fn the_home_value_comes_after_the_flag_and_the_profile_and_before_english() {
    let (_h, h) = home_with_state();
    home_file(&h, b"language:\n  conversation_language: ja\n");
    let (_p, none) = project(None);
    let (_q, chat) = project(Some("zh"));
    assert_eq!(
        conversation(&h, &none, &[]),
        "ja",
        "HOME beats the built-in en"
    );
    assert_eq!(conversation(&h, &chat, &[]), "zh", "the profile beats HOME");
    assert_eq!(
        conversation(&h, &chat, &["--conversation", "ko"]),
        "ko",
        "the flag beats the profile"
    );
    assert_eq!(
        conversation(&h, &none, &["--conversation", "ko"]),
        "ko",
        "the flag beats HOME"
    );
    // With no HOME value the built-in English stays.
    let (_e, empty) = home_with_state();
    assert_eq!(conversation(&empty, &none, &[]), "en");
    // The file `language-set` itself wrote is read back.
    let (_w, written) = home_with_state();
    assert_eq!(set(&written, "ko").0, 0);
    assert_eq!(conversation(&written, &none, &[]), "ko");
}

/// A HOME value beats the Go backup's conversation language (the user chose it at install).
#[test]
fn the_home_value_beats_the_go_backup() {
    let (_h, h) = home_with_state();
    home_file(&h, b"language:\n  conversation_language: ja\n");
    let (_p, root) = project(None);
    let sections = root.join(".moai-backup/config/sections");
    fs::create_dir_all(&sections).unwrap();
    fs::write(
        sections.join("language.yaml"),
        "language:\n  conversation_language: zh\n",
    )
    .unwrap();
    assert_eq!(conversation(&h, &root, &[]), "ja");
    // Without a HOME value the Go backup still applies, as before.
    let (_e, empty) = home_with_state();
    assert_eq!(conversation(&empty, &root, &[]), "zh");
}

/// A stored code that is well-formed but is not one MoAI has a name for (`fr`, `pt-BR`, `KO`) is
/// not used by `language-setup`: no panic, and it falls through to the next source, the Go
/// backup's value, and then `en`.
#[test]
fn a_stored_code_without_a_name_falls_through_without_a_panic() {
    let (_p, none) = project(None);
    let (_b, backup) = project(None);
    let sections = backup.join(".moai-backup/config/sections");
    fs::create_dir_all(&sections).unwrap();
    fs::write(
        sections.join("language.yaml"),
        "language:\n  conversation_language: zh\n",
    )
    .unwrap();
    for code in ["fr", "pt-BR", "es-419", "KO", "kor", "zh-Hant-TW"] {
        let (_h, h) = home_with_state();
        let (status, json) = set(&h, code);
        assert_eq!(status, 0, "{code}: {json}");
        assert_eq!(conversation(&h, &none, &[]), "en", "{code}: next is en");
        assert_eq!(conversation(&h, &backup, &[]), "zh", "{code}: next is Go");
        // The explicit flag still wins over any stored value.
        assert_eq!(conversation(&h, &none, &["--conversation", "ja"]), "ja");
    }
}

/// A missing, garbled, oversized, linked or wrong-typed HOME file silently means `en`.
#[test]
fn an_unusable_home_file_means_english() {
    let (_p, none) = project(None);
    let cases: Vec<(&str, Vec<u8>)> = vec![
        ("empty", b"".to_vec()),
        ("garbage", b"\x00\xff\xfe not yaml {{{".to_vec()),
        (
            "not utf8",
            b"language:\n  conversation_language: \xff\xfe\n".to_vec(),
        ),
        ("no section", b"conversation_language: ja\n".to_vec()),
        ("no key", b"language:\n  code_comments: ja\n".to_vec()),
        (
            "not a code",
            b"language:\n  conversation_language: fr\n".to_vec(),
        ),
        (
            "upper case",
            b"language:\n  conversation_language: JA\n".to_vec(),
        ),
        (
            "nested",
            b"language:\n  conversation_language:\n    - ja\n".to_vec(),
        ),
        ("tab", b"language:\n\tconversation_language: ja\n".to_vec()),
    ];
    for (name, bytes) in cases {
        let (_h, h) = home_with_state();
        home_file(&h, &bytes);
        assert_eq!(conversation(&h, &none, &[]), "en", "{name}");
    }
    // A link to a good file, a folder in its place, and a link as `config`.
    let good = b"language:\n  conversation_language: ja\n";
    let (_h, h) = home_with_state();
    fs::write(h.join("good.yaml"), good).unwrap();
    fs::create_dir(h.join(".moai-cli/config")).unwrap();
    symlink(h.join("good.yaml"), h.join(".moai-cli").join(FILE)).unwrap();
    assert_eq!(conversation(&h, &none, &[]), "en", "linked file");
    let (_h, h) = home_with_state();
    fs::create_dir_all(h.join(".moai-cli").join(FILE)).unwrap();
    assert_eq!(conversation(&h, &none, &[]), "en", "folder");
    let (_h, h) = home_with_state();
    fs::create_dir(h.join("real-config")).unwrap();
    fs::write(h.join("real-config/language.yaml"), good).unwrap();
    symlink(h.join("real-config"), h.join(".moai-cli/config")).unwrap();
    assert_eq!(conversation(&h, &none, &[]), "en", "linked config");
}

/// The read bound: a valid file of exactly LIMIT bytes is read; one byte more is not.
#[test]
fn the_home_file_is_read_up_to_the_limit_and_not_past_it() {
    let (_p, none) = project(None);
    let head = "language:\n  conversation_language: ja\n";
    let pad = |total: usize| {
        let mut s = String::from(head);
        s.push('#');
        while s.len() < total - 1 {
            s.push('x');
        }
        s.push('\n');
        assert_eq!(s.len(), total);
        s.into_bytes()
    };
    let (_h, h) = home_with_state();
    home_file(&h, &pad(LIMIT));
    assert_eq!(conversation(&h, &none, &[]), "ja", "N bytes are read");
    let (_h, h) = home_with_state();
    home_file(&h, &pad(LIMIT + 1));
    assert_eq!(conversation(&h, &none, &[]), "en", "N+1 bytes are not");
}
