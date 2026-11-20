//! `install-inspect`'s PATH note under a real terminal (#2867 R5a, PR #2881 audit B1).
//!
//! When the first `moai` on PATH is not this generation's, the note is printed once on stderr
//! only when stderr is a terminal; a lane or a script that pipes stderr reads no extra line. The
//! stdout JSON envelope is byte-identical either way and carries the note in
//! `path_first_moai.note`. Every run uses a temporary HOME. Moved here from the Python module
//! `install_inspect_tty_test.py` (SPEC-20261120-python-binary-tests-to-rust).

#![cfg(unix)]

mod common;
use common::pty::{output_within, Terminal};
use common::CleanEnv;
use std::{
    fs,
    os::unix::fs::PermissionsExt as _,
    path::{Path, PathBuf},
    process::{Command, Stdio},
};
use tempfile::TempDir;

const NOTE: &str = "moai: the first moai on PATH is not the Rust build; the generated hooks call moai, so hooks-apply writes them only after a Rust moai is first on PATH (nothing was changed)\n";

struct Fx {
    _tmp: TempDir,
    home: PathBuf,
    proj: PathBuf,
    /// A folder whose `moai` is a Go-generation shim.
    go: PathBuf,
    /// A folder whose `moai` is this build.
    rust: PathBuf,
}

fn fx() -> Fx {
    let tmp = TempDir::new().unwrap();
    let root = fs::canonicalize(tmp.path()).unwrap();
    let (home, proj, go, rust) = (
        root.join("home"),
        root.join("proj"),
        root.join("go-bin"),
        root.join("rust-bin"),
    );
    for dir in [&home, &proj, &go, &rust] {
        fs::create_dir_all(dir).unwrap();
    }
    let shim = go.join("moai");
    fs::write(&shim, "#!/bin/sh\necho 'moai 3.1.5 (go)'\n").unwrap();
    fs::set_permissions(&shim, fs::Permissions::from_mode(0o755)).unwrap();
    let target = rust.join("moai");
    if fs::hard_link(common::MOAI_CLI_BIN, &target).is_err() {
        common::copy_executable(Path::new(common::MOAI_CLI_BIN), &target);
    }
    Fx {
        _tmp: tmp,
        home,
        proj,
        go,
        rust,
    }
}

/// The environment of a run: `first` leads PATH, a temporary HOME, English locale and a terminal.
fn env_of(fx: &Fx, first: &Path, extra: &[(&str, &str)]) -> Vec<(String, String)> {
    let mut env = vec![
        (
            "PATH".to_string(),
            format!("{}:/usr/bin:/bin", first.display()),
        ),
        ("HOME".to_string(), fx.home.display().to_string()),
        ("LANG".to_string(), "en_US.UTF-8".to_string()),
        ("TERM".to_string(), "xterm-256color".to_string()),
    ];
    for (k, v) in extra {
        env.retain(|(name, _)| name != k);
        env.push((k.to_string(), v.to_string()));
    }
    env
}

struct Run {
    code: Option<i32>,
    stdout: String,
    stderr: String,
}

const ARGS: [&str; 2] = ["install-inspect", "--json"];

/// stdout and stderr both pipes.
fn piped(fx: &Fx, first: &Path, extra: &[(&str, &str)]) -> Run {
    let mut cmd = Command::new(common::MOAI_CLI_BIN);
    cmd.args(ARGS)
        .clean_env()
        .current_dir(&fx.proj)
        .stdin(Stdio::null());
    for (k, v) in env_of(fx, first, extra) {
        cmd.env(k, v);
    }
    let out = output_within(&mut cmd, None);
    Run {
        code: out.status.code(),
        stdout: String::from_utf8_lossy(&out.stdout).into_owned(),
        stderr: String::from_utf8_lossy(&out.stderr).into_owned(),
    }
}

/// stderr on a terminal, stdout a pipe.
fn stderr_on_terminal(fx: &Fx, first: &Path, extra: &[(&str, &str)]) -> Run {
    let mut term = Terminal::new(common::MOAI_CLI_BIN)
        .args(ARGS)
        .cwd(&fx.proj)
        .stdin_not_a_terminal()
        .stdout_not_a_terminal();
    for (k, v) in env_of(fx, first, extra) {
        term = term.env(k, v);
    }
    let s = term.run();
    Run {
        code: s.code,
        stdout: s.stdout,
        stderr: s.text,
    }
}

// The terminal line is English whatever the language settings say (#2825 R1); the envelope stays
// the same bytes and a pipe stays silent.
#[test]
fn a_korean_setting_keeps_the_english_note() {
    let f = fx();
    let english = piped(&f, &f.go, &[]).stdout;
    for extra in [[("MOAI_LANG", "ko")], [("LANG", "ko_KR.UTF-8")]] {
        let p = piped(&f, &f.go, &extra);
        let t = stderr_on_terminal(&f, &f.go, &extra);
        assert_eq!(
            (p.code, t.code),
            (Some(0), Some(0)),
            "{}{}",
            p.stderr,
            t.stderr
        );
        assert_eq!(p.stderr, "", "{extra:?}: a piped stderr carries no note");
        assert_eq!(t.stderr, NOTE, "{extra:?}: the English note exactly once");
        assert_eq!(
            t.stdout, english,
            "{extra:?}: the envelope bytes do not follow the language"
        );
        assert_eq!(p.stdout, english, "{extra:?}");
    }
}

#[test]
fn a_go_first_moai_prints_the_note_once_on_a_terminal_only() {
    let f = fx();
    let p = piped(&f, &f.go, &[]);
    let t = stderr_on_terminal(&f, &f.go, &[]);
    assert_eq!(
        (p.code, t.code),
        (Some(0), Some(0)),
        "{}{}",
        p.stderr,
        t.stderr
    );
    assert_eq!(p.stderr, "", "a piped stderr carries no note");
    assert_eq!(t.stderr, NOTE, "the terminal shows the note exactly once");
    assert_eq!(t.stdout, p.stdout, "the envelope is byte-identical");
    assert!(p.stdout.contains("\"generation\":\"go\""), "{}", p.stdout);
    let in_envelope = NOTE.strip_prefix("moai: ").unwrap().trim_end();
    assert!(p.stdout.contains(in_envelope), "{}", p.stdout);
}

#[test]
fn a_rust_first_moai_prints_nothing_on_a_terminal() {
    let f = fx();
    let p = piped(&f, &f.rust, &[]);
    let t = stderr_on_terminal(&f, &f.rust, &[]);
    assert_eq!(
        (p.code, t.code),
        (Some(0), Some(0)),
        "{}{}",
        p.stderr,
        t.stderr
    );
    assert_eq!(p.stderr, "");
    assert_eq!(
        t.stderr, "",
        "no note when the first moai is this generation's"
    );
    assert_eq!(t.stdout, p.stdout, "the envelope is byte-identical");
    assert!(p.stdout.contains("\"generation\":\"rust\""), "{}", p.stdout);
}
