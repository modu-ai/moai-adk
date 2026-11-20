//! Help is English on a terminal whatever the language settings say (Issue #2825 R1, user
//! decision 6878265192: the CLI's human output is English only; the user's language belongs to
//! the LLM session).
//!
//! The pipe bytes are pinned by `tests/help_layout.rs`; here a terminal, with no language setting
//! and with `MOAI_LANG=ko` or a Korean locale, must give those same bytes — for the root help,
//! `help --all`, a subcommand's `-h` and `--help`, and the stderr of a usage error. Every run
//! uses a temporary HOME. Moved here from the Python module `help_english_tty_test.py`
//! (SPEC-20261120-python-binary-tests-to-rust).

#![cfg(unix)]

mod common;
use common::pty::{output_within, Terminal};
use common::CleanEnv;
use std::{
    path::{Path, PathBuf},
    process::{Command, Stdio},
};
use tempfile::TempDir;

/// The Korean language settings: the app's own, and each of two locale variables.
const KOREAN_SETTINGS: [(&str, &str); 3] = [
    ("MOAI_LANG", "ko"),
    ("LANG", "ko_KR.UTF-8"),
    ("LC_ALL", "ko_KR.UTF-8"),
];

struct Fx {
    _tmp: TempDir,
    home: PathBuf,
    /// The binary as `<home>/bin/moai`, so help names `moai` as a person types it.
    moai: PathBuf,
}

fn fx() -> Fx {
    let tmp = TempDir::new().unwrap();
    let home = std::fs::canonicalize(tmp.path()).unwrap();
    let moai = home.join("bin/moai");
    std::fs::create_dir_all(moai.parent().unwrap()).unwrap();
    common::copy_executable(Path::new(common::MOAI_BIN), &moai);
    Fx {
        _tmp: tmp,
        home,
        moai,
    }
}

fn env_of(fx: &Fx, extra: &[(&str, &str)]) -> Vec<(String, String)> {
    let mut env: Vec<(String, String)> = [
        ("PATH", "/usr/bin:/bin"),
        ("HOME", fx.home.to_str().unwrap()),
        ("LANG", "en_US.UTF-8"),
        ("TERM", "xterm-256color"),
        ("COLUMNS", "60"),
        ("NO_COLOR", "1"),
    ]
    .iter()
    .map(|(k, v)| (k.to_string(), v.to_string()))
    .collect();
    for (k, v) in extra {
        env.retain(|(name, _)| name != k);
        env.push((k.to_string(), v.to_string()));
    }
    env
}

/// `(stdout, stderr, exit)` with both streams as pipes.
fn pipe(fx: &Fx, args: &[&str], extra: &[(&str, &str)]) -> (String, String, Option<i32>) {
    let mut cmd = Command::new(&fx.moai);
    cmd.args(args)
        .clean_env()
        .current_dir(&fx.home)
        .stdin(Stdio::null());
    for (k, v) in env_of(fx, extra) {
        cmd.env(k, v);
    }
    let out = output_within(&mut cmd, None);
    (
        String::from_utf8_lossy(&out.stdout).into_owned(),
        String::from_utf8_lossy(&out.stderr).into_owned(),
        out.status.code(),
    )
}

/// `(stdout on a terminal, stderr pipe, exit)`. The terminal turns each `\n` into `\r\n`, and
/// under a full output queue macOS has been seen to send one `\r` twice (`\r\r\n` in a 15 KB
/// `help --all`); help text carries no `\r` of its own, so every `\r` goes.
fn on_a_terminal(fx: &Fx, args: &[&str], extra: &[(&str, &str)]) -> (String, String, Option<i32>) {
    let mut term = Terminal::new(&fx.moai)
        .args(args)
        .cwd(&fx.home)
        .stdin_not_a_terminal()
        .stderr_not_a_terminal();
    for (k, v) in env_of(fx, extra) {
        term = term.env(k, v);
    }
    let s = term.run();
    // The pty turns each `\n` into `\r\n`, and under a full output queue macOS has been seen to
    // send one `\r` twice (`\r\r\n`, in a 15 KB `help --all`); `Screen::text` reads `\r\n` as `\n`,
    // so the odd `\r` is left over. Help text carries no `\r` of its own.
    (s.text.replace('\r', ""), s.stderr, s.code)
}

/// `text` from its help title on — the startup banner a terminal gets before the root help is
/// the banner's own (#2803), not help — with the version of each title line as `<version>`.
fn normal(text: &str) -> String {
    let title = |line: &str| -> Option<usize> {
        // `MoAI-ADK <version> — `: the length of that prefix.
        let rest = line.strip_prefix("MoAI-ADK ")?;
        let (version, _) = rest.split_once(" — ")?;
        (!version.is_empty() && !version.contains(char::is_whitespace))
            .then(|| "MoAI-ADK ".len() + version.len() + " — ".len())
    };
    let mut out = String::new();
    let mut started = false;
    for line in text.split_inclusive('\n') {
        match title(line) {
            Some(prefix) => {
                started = true;
                out.push_str("MoAI-ADK <version> — ");
                out.push_str(&line[prefix..]);
            }
            None if started => out.push_str(line),
            None => {}
        }
    }
    if started {
        out
    } else {
        text.to_string()
    }
}

fn hangul(text: &str) -> bool {
    text.chars().any(|c| ('가'..='힣').contains(&c))
}

fn help_on_a_terminal(fx: &Fx, args: &[&str], extra: &[(&str, &str)]) -> String {
    let (out, err, code) = on_a_terminal(fx, args, extra);
    assert_eq!(code, Some(0), "{args:?}: {err}");
    normal(&out)
}

#[test]
fn a_terminal_gets_the_pipe_bytes_in_every_language_setting() {
    let f = fx();
    for args in [
        vec!["--help"],
        vec!["help", "--all"],
        vec!["project-detect", "--help"],
        vec!["key-set", "-h"],
        vec!["cc", "--help"],
        vec!["update", "-h"],
    ] {
        let (piped, _, code) = pipe(&f, &args, &[]);
        assert_eq!(code, Some(0), "{args:?}");
        let piped = normal(&piped);
        assert_eq!(help_on_a_terminal(&f, &args, &[]), piped, "{args:?}");
        for setting in KOREAN_SETTINGS {
            assert_eq!(
                help_on_a_terminal(&f, &args, &[setting]),
                piped,
                "{args:?} {setting:?}"
            );
        }
    }
}

#[test]
fn no_command_help_carries_korean_under_a_korean_setting() {
    let f = fx();
    let (listing, _, _) = pipe(&f, &["help", "--all"], &[]);
    let mut words: Vec<String> = listing
        .lines()
        .filter(|l| l.starts_with("  ") && !l.starts_with("   "))
        .filter_map(|l| l.split_whitespace().next())
        .filter(|w| !w.starts_with('-'))
        .map(str::to_string)
        .collect();
    words.sort();
    words.dedup();
    assert!(words.len() > 100, "{} commands", words.len());
    let mut korean = Vec::new();
    for word in &words {
        let args = [word.as_str(), "--help"];
        let (english, _, _) = pipe(&f, &args, &[]);
        let text = help_on_a_terminal(&f, &args, &[("MOAI_LANG", "ko")]);
        // A command's own English about may quote a Korean phrase; it is in the pipe too.
        if hangul(&text) && text != normal(&english) {
            korean.push(word.clone());
        }
    }
    assert!(korean.is_empty(), "{korean:?}");
}

#[test]
fn a_usage_error_is_english_on_a_terminal_too() {
    let f = fx();
    for setting in KOREAN_SETTINGS {
        for args in [vec!["--bogus"], vec!["bogus"], vec!["project-detect"]] {
            let (_, piped_err, piped_code) = pipe(&f, &args, &[]);
            let (_, err, code) = on_a_terminal(&f, &args, &[setting]);
            assert_eq!((code, err), (piped_code, piped_err), "{args:?} {setting:?}");
        }
    }
}
