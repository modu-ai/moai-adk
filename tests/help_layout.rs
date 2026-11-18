//! #2802 S1 (design #2799 D2 sections 9 and 11): the people-facing help, pinned as goldens, and
//! the machine envelope left untouched by it. #3004 P1: every name prints `moai`, and the root
//! help leads with the first steps.
//!
//! The goldens are the help a person sees when stdout is not a terminal (no logo): the root help
//! and `help --all`, in the default width, `NO_COLOR=1 TERM=dumb`, `COLUMNS=60` and
//! `COLUMNS=80`. Rewrite them with `MOAI_HELP_GOLDEN_WRITE=1 cargo test --test help_layout` and
//! review the diff.
mod common;

use common::CleanEnv;
use std::path::{Path, PathBuf};
use std::process::Command;

fn golden_dir() -> PathBuf {
    Path::new(env!("CARGO_MANIFEST_DIR")).join("tests/fixtures/help")
}

fn run(bin: &str, args: &[&str], env: &[(&str, &str)]) -> std::process::Output {
    let home = tempfile::TempDir::new().unwrap();
    let mut cmd = Command::new(bin);
    cmd.args(args)
        .clean_env()
        .env("HOME", home.path())
        .env("PATH", "/usr/bin:/bin");
    for (k, v) in env {
        cmd.env(k, v);
    }
    cmd.output().unwrap()
}

/// (golden name, arguments, environment).
type Case = (
    &'static str,
    &'static [&'static str],
    &'static [(&'static str, &'static str)],
);

const CASES: [Case; 8] = [
    ("root", &["--help"], &[]),
    (
        "root-dumb",
        &["--help"],
        &[("NO_COLOR", "1"), ("TERM", "dumb")],
    ),
    ("root-60", &["--help"], &[("COLUMNS", "60")]),
    ("root-80", &["--help"], &[("COLUMNS", "80")]),
    ("all", &["help", "--all"], &[]),
    (
        "all-dumb",
        &["help", "--all"],
        &[("NO_COLOR", "1"), ("TERM", "dumb")],
    ),
    ("all-60", &["help", "--all"], &[("COLUMNS", "60")]),
    ("all-80", &["help", "--all"], &[("COLUMNS", "80")]),
];

/// What the goldens hold in place of this build's version.
const VERSION_TOKEN: &str = "<version>";

#[test]
fn help_matches_its_goldens() {
    let write = std::env::var_os("MOAI_HELP_GOLDEN_WRITE").is_some();
    let mut failed = Vec::new();
    for (name, args, env) in CASES {
        let out = run(common::MOAI_BIN, args, env);
        assert!(out.status.success(), "{name}: {out:?}");
        let text = String::from_utf8(out.stdout).unwrap();
        // A pipe never carries a style code (#3004 AC-5).
        assert!(!text.contains('\x1b'), "{name}: escape in piped help");
        // Card C-67 S1: the build version is HeadVer, whose yearweek and build change with
        // every build, so the goldens hold `<version>` where the help names it.
        let first = text.lines().next().unwrap_or("");
        assert_eq!(
            first,
            format!(
                "MoAI-ADK {} — agentic development kit",
                moai_cli::BUILD_VERSION
            ),
            "{name}"
        );
        let text = text.replacen(moai_cli::BUILD_VERSION, VERSION_TOKEN, 1);
        let path = golden_dir().join(format!("{name}.txt"));
        if write {
            std::fs::create_dir_all(golden_dir()).unwrap();
            std::fs::write(&path, &text).unwrap();
            continue;
        }
        let want = std::fs::read_to_string(&path)
            .unwrap_or_else(|_| panic!("missing golden {}", path.display()));
        if text != want {
            failed.push(name);
        }
    }
    assert!(
        failed.is_empty(),
        "help differs from its golden: {failed:?}"
    );
}

/// The root help stays inside the width it is asked for: no line passes 60 columns at
/// `COLUMNS=60` or 80 at `COLUMNS=80`, under both names, and the whole screen stays short enough
/// to read (#3004 P1 grew it from D2 section 9's 45 lines by the first steps, the examples and
/// the pointers).
#[test]
fn the_root_help_is_short_and_fits_sixty_and_eighty_columns() {
    for bin in [common::MOAI_BIN, common::MOAI_CLI_BIN] {
        let wide = String::from_utf8(run(bin, &["--help"], &[]).stdout).unwrap();
        assert!(
            wide.lines().count() <= 60,
            "{bin}: {} lines",
            wide.lines().count()
        );
        for columns in [60usize, 80] {
            let text = String::from_utf8(
                run(bin, &["--help"], &[("COLUMNS", &columns.to_string())]).stdout,
            )
            .unwrap();
            for line in text.lines() {
                assert!(
                    unicode_width::UnicodeWidthStr::width(line) <= columns,
                    "{bin} COLUMNS={columns}: {line:?}"
                );
            }
        }
    }
}

/// Every name prints `moai` in help and usage text (#3004 AC-1): `moai` and `moai-cli` give the
/// same bytes, and the first `--version` line stays the signature line (D3 decision 2; the
/// second line names the kind of build, card C-67 S1).
#[test]
fn every_name_prints_moai_and_the_version_line_is_unchanged() {
    let reference = |args: &[&str]| run(common::MOAI_BIN, args, &[]);
    for bin in [common::MOAI_BIN, common::MOAI_CLI_BIN] {
        let help = String::from_utf8(run(bin, &["--help"], &[]).stdout).unwrap();
        assert!(
            help.contains("Usage: moai <command> [options]"),
            "{bin}: {help}"
        );
        assert!(help.contains("  moai help --all "), "{bin}: {help}");
        assert!(help.contains("  moai <command> --help "), "{bin}: {help}");
        let sub = String::from_utf8(run(bin, &["identity", "--help"], &[]).stdout).unwrap();
        assert!(sub.contains("Usage: moai identity"), "{bin}: {sub}");
        let all = String::from_utf8(run(bin, &["help", "--all"], &[]).stdout).unwrap();
        assert!(
            all.contains("Run 'moai <command> --help' for details and examples."),
            "{bin}: {all}"
        );
        for args in [
            &["--help"][..],
            &["-h"][..],
            &["help", "--all"][..],
            &["identity", "--help"][..],
            &["identity", "-h"][..],
            &["identity"][..],
            &["no-such-command"][..],
            &["--bogus"][..],
            &[][..],
        ] {
            let want = reference(args);
            let got = run(bin, args, &[]);
            assert_eq!(got.status.code(), want.status.code(), "{bin} {args:?}");
            assert_eq!(got.stdout, want.stdout, "{bin} {args:?} stdout");
            assert_eq!(got.stderr, want.stderr, "{bin} {args:?} stderr");
        }
        let version = String::from_utf8(run(bin, &["--version"], &[]).stdout).unwrap();
        let first = version.lines().next().unwrap_or("");
        assert!(
            first.starts_with("moai-cli ") && first.ends_with(" (rust)"),
            "{version}"
        );
        assert_eq!(version.lines().count(), 2, "{version}");
    }
}

/// `moai-cli -h` and every help text of it name no command `moai-cli` (#3004 AC-1). What may
/// stay is a folder or file name (`.moai-cli/…`, `~/.moai-cli`, `…/moai-cli/heavy-slots`) and
/// the sentences that describe the alias file itself (`ALIAS_FILE`): those are things on disk,
/// not the name a person types.
#[test]
fn moai_cli_help_names_no_command_moai_cli() {
    /// The two sentences of `bin-names-plan` that say which file the `moai` link is made beside
    /// and points at: the alias executable itself, which stays.
    const ALIAS_FILE: [&str; 2] = [
        "Preview the moai link beside moai-cli",
        "should be a hard link to `moai-cli`",
    ];
    let as_command = |text: &str| -> Vec<String> {
        let mut hits = Vec::new();
        for line in text.lines() {
            let line = ALIAS_FILE
                .iter()
                .fold(line.to_string(), |l, phrase| l.replace(phrase, ""));
            for (i, _) in line.match_indices("moai-cli") {
                let before = line[..i].chars().next_back();
                let after = line[i + "moai-cli".len()..].chars().next();
                if !matches!(before, Some('.' | '/' | '~'))
                    && !matches!(after, Some('/' | '-' | '.' | '_'))
                {
                    hits.push(line.trim().to_string());
                }
            }
        }
        hits
    };
    let mut found = Vec::new();
    for args in [&["-h"][..], &["--help"][..], &["help", "--all"][..]] {
        let out = run(common::MOAI_CLI_BIN, args, &[]);
        let text = String::from_utf8(out.stdout).unwrap();
        for hit in as_command(&text) {
            found.push(format!("{args:?}: {hit}"));
        }
    }
    for g in moai_cli::help::GROUPS {
        for (word, _) in g.commands {
            for flag in ["-h", "--help"] {
                let out = run(common::MOAI_CLI_BIN, &[*word, flag], &[]);
                let text = String::from_utf8(out.stdout).unwrap();
                for hit in as_command(&text) {
                    found.push(format!("{word} {flag}: {hit}"));
                }
            }
        }
    }
    // A usage error names `moai` too, on stderr and in the envelope.
    for args in [&[][..], &["no-such-command"][..], &["identity"][..]] {
        let out = run(common::MOAI_CLI_BIN, args, &[]);
        for text in [
            String::from_utf8_lossy(&out.stdout),
            String::from_utf8_lossy(&out.stderr),
        ] {
            for hit in as_command(&text) {
                found.push(format!("{args:?}: {hit}"));
            }
        }
    }
    assert!(
        found.is_empty(),
        "moai-cli as a command name:\n{}",
        found.join("\n")
    );
}

/// The first steps the root help names are real commands (#3004 AC-5): every
/// `moai <command>` line of "Getting started" and "Examples" starts with a command this build
/// has, and that command answers `--help`, so a removed command breaks this test. A step that
/// starts with `/` is typed inside a session, not to the shell.
#[test]
fn the_getting_started_and_example_commands_are_real() {
    let root =
        String::from_utf8(run(common::MOAI_BIN, &["--help"], &[("COLUMNS", "80")]).stdout).unwrap();
    let section = |name: &str| -> Vec<String> {
        let mut lines = root.lines().skip_while(|l| *l != name).skip(1);
        let mut out = Vec::new();
        for l in lines.by_ref() {
            if l.is_empty() {
                break;
            }
            out.push(l.to_string());
        }
        out
    };
    let mut checked = Vec::new();
    for name in ["Getting started", "Examples"] {
        let lines = section(name);
        assert!(!lines.is_empty(), "no {name} section:\n{root}");
        for l in lines {
            // `  1. moai cc   about`, `  moai cc -l   about`, `  3. /moai project   about`.
            let body = l.trim_start();
            let body = body
                .split_once(". ")
                .filter(|(n, _)| n.chars().all(|c| c.is_ascii_digit()))
                .map_or(body, |(_, rest)| rest);
            if body.starts_with('/') {
                continue;
            }
            let command = body
                .strip_prefix("moai ")
                .unwrap_or_else(|| panic!("{name}: not a moai command: {l:?}"));
            let words: Vec<&str> = command
                .split("  ")
                .next()
                .unwrap()
                .split_whitespace()
                .collect();
            let word = words[0];
            assert!(
                moai_cli::help::GROUPS
                    .iter()
                    .any(|g| g.commands.iter().any(|(w, _)| *w == word)),
                "{name}: `{word}` is no command: {l:?}"
            );
            let out = run(common::MOAI_BIN, &[word, "--help"], &[]);
            assert_eq!(out.status.code(), Some(0), "{name}: {word} --help");
            checked.push(word.to_string());
        }
    }
    assert!(checked.len() >= 5, "{checked:?}");
}

/// The help layer never reaches stdout's envelope (D2 section 1): a usage error prints the same
/// envelope bytes under both names.
#[test]
fn a_usage_error_envelope_is_byte_identical_under_both_names() {
    for args in [&["identity"][..], &["go-state-plan", "--nope", "."][..]] {
        let a = run(common::MOAI_BIN, args, &[]);
        let b = run(common::MOAI_CLI_BIN, args, &[]);
        assert_eq!(a.status.code(), Some(2), "{args:?}");
        assert_eq!(a.stdout, b.stdout, "{args:?}");
        let envelope: serde_json::Value = serde_json::from_slice(&a.stdout).unwrap();
        assert_eq!(envelope["ok"], serde_json::json!(false), "{args:?}");
        assert_eq!(
            envelope["error"]["kind"],
            serde_json::json!("usage"),
            "{args:?}"
        );
    }
}

/// Every command appears in `help --all`, and the root help lists only the root ones.
#[test]
fn help_all_lists_every_command_and_the_root_a_subset() {
    let all = String::from_utf8(run(common::MOAI_BIN, &["help", "--all"], &[]).stdout).unwrap();
    let root = String::from_utf8(run(common::MOAI_BIN, &["--help"], &[]).stdout).unwrap();
    // A row is `  <word>` then spaces and its about, or `  <word>` alone when it has none.
    // The root help sets the groups one level under "Commands" (4 spaces), `help --all` flush
    // (2 spaces).
    let listed = |text: &str, word: &str| {
        text.lines().any(|l| {
            l.strip_prefix("  ")
                .map(|r| r.strip_prefix("  ").unwrap_or(r))
                .and_then(|r| r.strip_prefix(word))
                .is_some_and(|r| r.is_empty() || r.starts_with(' '))
        })
    };
    for g in moai_cli::help::GROUPS {
        assert!(all.contains(&format!("\n{}\n", g.heading)), "{}", g.heading);
        for (word, on_root) in g.commands {
            assert!(listed(&all, word), "help --all lacks {word}");
            if !["cc", "codex", "glm", "omp"].contains(word) {
                assert_eq!(listed(&root, word), *on_root, "root help and {word}");
            }
        }
    }
    assert!(root.contains("  cc, codex, glm, omp "), "{root}");
}

/// #2824 re-audit B3: a wide `COLUMNS` never takes a help line past 100 columns — the grouped
/// list, clap's options section of `help --all`, and a command's own help alike.
#[test]
fn no_help_line_goes_past_one_hundred_columns() {
    for args in [
        &["help", "--all"][..],
        &["--help"][..],
        &["identity", "--help"][..],
    ] {
        for columns in ["101", "300"] {
            let out = run(common::MOAI_BIN, args, &[("COLUMNS", columns)]);
            assert!(out.status.success(), "{args:?}");
            let text = String::from_utf8(out.stdout).unwrap();
            for line in text.lines() {
                assert!(
                    unicode_width::UnicodeWidthStr::width(line) <= 100,
                    "{args:?} COLUMNS={columns}: {} columns: {line:?}",
                    unicode_width::UnicodeWidthStr::width(line)
                );
            }
        }
    }
}

/// Only the exact form `help --all` is answered before parsing; anything after it goes to clap,
/// which refuses it as a usage error instead of printing the list.
#[test]
fn help_all_with_more_arguments_is_not_the_grouped_list() {
    let out = run(common::MOAI_BIN, &["help", "--all", "extra"], &[]);
    assert_eq!(out.status.code(), Some(2), "{out:?}");
    let text = String::from_utf8_lossy(&out.stdout);
    assert!(!text.contains("Lane and leader"), "{text}");
}

/// The bytes `bin args…` writes to a pseudo-terminal as stdout, with `env` on top of a clean
/// environment (`TERM=xterm-256color`, a throwaway HOME).
#[cfg(unix)]
fn on_a_terminal(bin: &str, args: &[&str], env: &[(&str, &str)]) -> String {
    use std::io::Read;
    use std::os::fd::{FromRawFd, OwnedFd};
    use std::process::Stdio;
    let mut master: libc::c_int = -1;
    let mut slave: libc::c_int = -1;
    let rc = unsafe {
        libc::openpty(
            &mut master,
            &mut slave,
            // macOS takes `*mut` for the last two, Linux `*const`; a `*mut` null fits both.
            std::ptr::null_mut(),
            std::ptr::null_mut(),
            std::ptr::null_mut(),
        )
    };
    assert_eq!(rc, 0, "openpty");
    // Owned from here on: the parent's copy of the slave closes with `cmd`.
    let master = unsafe { OwnedFd::from_raw_fd(master) };
    let slave = unsafe { OwnedFd::from_raw_fd(slave) };
    let home = tempfile::TempDir::new().unwrap();
    let mut cmd = Command::new(bin);
    cmd.args(args)
        .clean_env()
        .env("HOME", home.path())
        .env("PATH", "/usr/bin:/bin")
        .env("TERM", "xterm-256color")
        .env("COLUMNS", "80")
        .stdin(Stdio::null())
        .stdout(Stdio::from(slave))
        .stderr(Stdio::null());
    for (k, v) in env {
        cmd.env(k, v);
    }
    let mut child = cmd.spawn().unwrap();
    drop(cmd);
    let mut bytes = Vec::new();
    let mut file = std::fs::File::from(master);
    let mut buf = [0u8; 8192];
    loop {
        match file.read(&mut buf) {
            Ok(0) | Err(_) => break,
            Ok(n) => bytes.extend_from_slice(&buf[..n]),
        }
    }
    assert!(child.wait().unwrap().success());
    String::from_utf8_lossy(&bytes).replace('\r', "")
}

/// `text` from the help's title line on: a terminal gets the startup logo above it, and the
/// logo is the banner's own colour, not help's.
#[cfg(unix)]
fn from_the_title(text: &str) -> &str {
    let at = text.find("MoAI-ADK ").expect("no help title");
    &text[at..]
}

/// `text` without its SGR style codes (`ESC [ … m`).
#[cfg(unix)]
fn without_styles(text: &str) -> String {
    let mut out = String::new();
    let mut chars = text.chars().peekable();
    while let Some(c) = chars.next() {
        if c == '\x1b' && chars.peek() == Some(&'[') {
            for d in chars.by_ref() {
                if d == 'm' {
                    break;
                }
            }
        } else {
            out.push(c);
        }
    }
    out
}

/// Colour and bold only on a terminal, and `NO_COLOR` wins over the terminal (#3004 AC-5):
/// a terminal gets style codes in the root help, the same words as the pipe once they are
/// removed; `NO_COLOR` and `TERM=dumb` get none even on a terminal.
#[cfg(unix)]
#[test]
fn the_root_help_is_styled_on_a_terminal_and_no_color_turns_it_off() {
    let piped =
        String::from_utf8(run(common::MOAI_BIN, &["--help"], &[("COLUMNS", "80")]).stdout).unwrap();
    assert!(!piped.contains('\x1b'));
    let styled = on_a_terminal(common::MOAI_BIN, &["--help"], &[]);
    let help = from_the_title(&styled);
    // clap styles its own `Usage:` heading on a terminal, so the style codes asked for are the
    // ones on the help's own headings and command names.
    for heading in ["Getting started", "Commands", "Examples", "More"] {
        let line = help
            .lines()
            .find(|l| without_styles(l) == heading)
            .unwrap_or_else(|| panic!("no {heading} line: {help:?}"));
        assert!(line.contains("\x1b["), "{heading} is not styled: {line:?}");
    }
    let step = help
        .lines()
        .find(|l| without_styles(l).starts_with("  1. moai install-inspect"))
        .unwrap();
    assert!(step.contains("\x1b["), "command not styled: {step:?}");
    assert_eq!(
        without_styles(help),
        piped,
        "same words with and without style"
    );
    for env in [&[("NO_COLOR", "1")][..], &[("TERM", "dumb")][..]] {
        let plain = on_a_terminal(common::MOAI_BIN, &["--help"], env);
        assert!(
            !from_the_title(&plain).contains('\x1b'),
            "{env:?}: style code on a terminal: {plain:?}"
        );
    }
}
