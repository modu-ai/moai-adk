//! The confirm prompt and static progress lines under a real terminal (Issue #2806, design
//! #2799 D2 §5·§7·§11, leader decision 6866110172).
//!
//! `key-remove` without `--yes` asks only when stdin and stderr are both terminals; a pipe keeps
//! the old non-interactive answer byte for byte. `index-build` prints one line per stage only
//! when stderr is a terminal, and its stdout envelope does not change with it. `forge-read`
//! names the page it reads from the second page on, on a terminal only. Every run uses a
//! temporary HOME. Moved here from the Python module `ui_progress_confirm_test.py`
//! (SPEC-20261120-python-binary-tests-to-rust).

#![cfg(unix)]

mod common;
use common::pty::{env_with, piped_with, Env, Seen, Terminal};
use common::screens::strip_colors;
use std::{
    collections::BTreeSet,
    fs,
    os::unix::fs::PermissionsExt as _,
    path::{Path, PathBuf},
};
use tempfile::TempDir;

const BIN: &str = common::MOAI_CLI_BIN;

struct Fx {
    tmp: TempDir,
    home: PathBuf,
    work: PathBuf,
}

fn fx() -> Fx {
    let tmp = TempDir::new().unwrap();
    let root = fs::canonicalize(tmp.path()).unwrap();
    let (home, work) = (root.join("home"), root.join("work"));
    fs::create_dir_all(&home).unwrap();
    fs::create_dir_all(&work).unwrap();
    Fx { tmp, home, work }
}

impl Fx {
    fn root(&self) -> PathBuf {
        fs::canonicalize(self.tmp.path()).unwrap()
    }

    fn env(&self, extra: &[(&str, &str)]) -> Env {
        env_with(
            &[
                ("PATH", "/usr/bin:/bin"),
                ("HOME", self.home.to_str().unwrap()),
                ("LANG", "C"),
                ("TERM", "xterm"),
            ],
            extra,
        )
    }

    /// stderr on a terminal, stdout a pipe, in `cwd`, with this fixture's environment.
    fn terminal(&self, cwd: &Path, args: &[&str], extra: &[(&str, &str)]) -> Terminal {
        let mut term = Terminal::new(BIN)
            .args(args)
            .cwd(cwd)
            .stdout_not_a_terminal();
        for (k, v) in self.env(extra) {
            term = term.env(k, v);
        }
        term
    }

    fn piped(&self, cwd: &Path, args: &[&str], stdin: &str, extra: &[(&str, &str)]) -> Seen {
        piped_with(BIN, cwd, &self.env(extra), args, Some(stdin))
    }

    fn set_key(&self) {
        let seen = self.piped(
            &self.work,
            &["key-set", "glm", "--shared"],
            "not-a-real-key\n",
            &[],
        );
        assert_eq!(seen.code, Some(0), "{}", seen.stderr);
    }

    /// A git project with one committed file.
    fn project(&self, name: &str) -> PathBuf {
        let root = self.root().join(name);
        fs::create_dir_all(&root).unwrap();
        fs::write(root.join("a.md"), "# a\n").unwrap();
        common::git::git_in(&root, &["init", "-q"]);
        common::git::git_in(&root, &["add", "a.md"]);
        common::git::git_in(&root, &["commit", "-qm", "a"]);
        root
    }
}

fn json(text: &str) -> serde_json::Value {
    serde_json::from_str(text).unwrap_or_else(|e| panic!("not JSON ({e}): {text:?}"))
}

fn arrow_lines(screen: &str) -> Vec<String> {
    screen
        .lines()
        .filter(|l| l.starts_with("->"))
        .map(str::to_string)
        .collect()
}

// ---- key-remove ----

#[test]
fn a_pipe_never_asks_and_keeps_the_old_answer() {
    let f = fx();
    f.set_key();
    let piped = f.piped(&f.work, &["key-remove", "glm", "--shared"], "", &[]);
    assert_eq!(piped.code, Some(0), "{}", piped.stderr);
    assert!(!piped.stderr.contains("[y/N]"), "{}", piped.stderr);
    assert_eq!(json(&piped.stdout)["result"]["removed"], false);
    // stdin not a terminal, stderr a terminal: still no question (both must be terminals).
    let s = f
        .terminal(&f.work, &["key-remove", "glm", "--shared"], &[])
        .stdin_not_a_terminal()
        .run();
    assert!(!s.text.contains("[y/N]"), "{}", s.text);
    assert_eq!(
        s.stdout, piped.stdout,
        "the stdout envelope is the same bytes"
    );
}

#[test]
fn no_on_the_terminal_removes_nothing_and_names_the_flag() {
    let f = fx();
    f.set_key();
    let piped = f.piped(&f.work, &["key-remove", "glm", "--shared"], "", &[]);
    let s = f
        .terminal(&f.work, &["key-remove", "glm", "--shared"], &[])
        .answer(b"n\n")
        .run();
    let term = strip_colors(&s.text);
    assert_eq!(s.code, Some(0), "{term}");
    assert!(term.contains("key-remove deletes 1 file"), "{term}");
    assert!(term.contains("glm key, shared scope:"), "{term}");
    assert!(term.contains("! Remove it? [y/N] "), "{term}");
    assert!(
        term.contains("next: run with --yes to remove it without asking"),
        "{term}"
    );
    assert_eq!(
        s.stdout, piped.stdout,
        "a refusal answers with the preview envelope as before"
    );
    assert!(!s.stdout.contains("\u{1b}["), "{:?}", s.stdout);
}

#[test]
fn yes_on_the_terminal_equals_the_yes_flag() {
    let f = fx();
    f.set_key();
    let s = f
        .terminal(&f.work, &["key-remove", "glm", "--shared"], &[])
        .answer(b"y\n")
        .run();
    assert_eq!(s.code, Some(0), "{}", s.text);
    assert_eq!(json(&s.stdout)["result"]["removed"], true);
    f.set_key();
    let flag = f.piped(
        &f.work,
        &["key-remove", "glm", "--shared", "--yes"],
        "",
        &[],
    );
    assert_eq!(
        s.stdout, flag.stdout,
        "an interactive yes answers like --yes, byte for byte"
    );
}

#[test]
fn key_set_refuses_a_terminal_stdin() {
    // The command left src/main.rs for src/ops/ (#2834 R3); a typed value would be echoed.
    let f = fx();
    let s = f
        .terminal(&f.work, &["key-set", "glm", "--shared"], &[])
        .run();
    assert_eq!(s.code, Some(1), "{}", s.text);
    assert_eq!(
        json(&s.stdout)["error"]["message"],
        "key-set reads the value from stdin and refuses a terminal (the value would be echoed); pipe it in instead"
    );
    assert!(!f.home.join(".moai-cli").join("vault").exists());
}

#[test]
fn nothing_to_remove_asks_nothing() {
    let f = fx();
    let s = f
        .terminal(&f.work, &["key-remove", "glm", "--shared"], &[])
        .answer(b"y\n")
        .run();
    assert_eq!(s.code, Some(0), "{}", s.text);
    assert!(!s.text.contains("[y/N]"), "{}", s.text);
    assert_eq!(json(&s.stdout)["result"]["exists"], false);
}

// ---- index-build ----

const STAGES: [&str; 3] = [
    "scanning tracked files (1/3)",
    "building the views (2/3)",
    "publishing the generation (3/3)",
];

const REQUEST: &str = "{\"protocol_version\":1}";

fn keys(text: &str) -> BTreeSet<String> {
    json(text).as_object().unwrap().keys().cloned().collect()
}

#[test]
fn stages_show_on_a_terminal_and_never_on_a_pipe() {
    let f = fx();
    let (a, b) = (f.project("a"), f.project("b"));
    let piped = f.piped(&a, &["index-build", a.to_str().unwrap()], REQUEST, &[]);
    assert_eq!(piped.code, Some(0), "{}", piped.stderr);
    assert_eq!(piped.stderr, "", "no progress byte on a pipe");
    // stdin carries the request, so only stderr is the terminal here.
    let s = f
        .terminal(&b, &["index-build", b.to_str().unwrap()], &[])
        .stdin_text(REQUEST)
        .run();
    let term = strip_colors(&s.text);
    assert_eq!(s.code, Some(0), "{term}");
    assert_eq!(
        arrow_lines(&term),
        vec![
            "-> scanning tracked files (1/3)",
            "-> building the views (2/3)",
            "-> publishing the generation (3/3)",
        ]
    );
    assert!(!term.replace('\n', "").contains('\r'), "{term:?}");
    assert_eq!(keys(&s.stdout), keys(&piped.stdout));
}

#[test]
fn moai_output_quiet_drops_the_stages_and_plain_drops_color_and_symbols() {
    let f = fx();
    let utf8 = [("LANG", "en_US.UTF-8"), ("LC_ALL", "en_US.UTF-8")];
    let shown = |name: &str, extra: &[(&str, &str)]| {
        let root = f.project(name);
        let mut env: Vec<(&str, &str)> = utf8.to_vec();
        env.extend_from_slice(extra);
        let s = f
            .terminal(&root, &["index-build", root.to_str().unwrap()], &env)
            .stdin_text(REQUEST)
            .run();
        let term = strip_colors(&s.text);
        assert_eq!(s.code, Some(0), "{term}");
        (s.stdout, term)
    };

    let (_, normal) = shown("n", &[]);
    assert!(
        normal.contains("→ scanning tracked files (1/3)"),
        "{normal}"
    );
    let (_, plain) = shown("p", &[("MOAI_OUTPUT", "plain")]);
    assert_eq!(
        arrow_lines(&plain),
        STAGES.iter().map(|s| format!("-> {s}")).collect::<Vec<_>>()
    );
    assert!(!plain.contains('\u{1b}'), "{plain:?}");
    assert!(!plain.contains('→'), "{plain:?}");
    let (out, quiet) = shown("q", &[("MOAI_OUTPUT", "quiet")]);
    assert!(!STAGES.iter().any(|s| quiet.contains(s)), "{quiet}");
    assert_eq!(json(&out)["ok"], true);
}

// ---- forge-read ----

const GH_STUB: &str = r#"#!/bin/sh
for a in "$@"; do last="$a"; done
case "$last" in
  *issues/comments*) printf 'HTTP/2.0 200 OK\r\nEtag: W/"c-1"\r\n\r\n[]';;
  *issues*page=2*) printf 'HTTP/2.0 200 OK\r\nEtag: W/"i-2"\r\n\r\n[]';;
  *issues*) printf 'HTTP/2.0 200 OK\r\nEtag: W/"i-1"\r\nLink: <https://api.github.com/repos/acme/widgets/issues?state=all&page=2>; rel="next"\r\n\r\n[]';;
  *) printf 'HTTP/2.0 200 OK\r\nEtag: "repo-1"\r\n\r\n{"id":101,"full_name":"acme/widgets"}';;
esac
"#;

// `forge-read` names the page it reads from the second page on, on a terminal only. Its command
// left `src/main.rs` for `src/ops/forge_read.rs` (#2834 R2b), so the terminal check moved from
// `shows_progress` into the operation; this pins that it still runs.
#[test]
fn the_second_page_shows_on_a_terminal_and_never_on_a_pipe() {
    let f = fx();
    let gh = f.root().join("gh-stub");
    fs::write(&gh, GH_STUB).unwrap();
    fs::set_permissions(&gh, fs::Permissions::from_mode(0o755)).unwrap();
    let extra = [
        ("MOAI_FORGE_GH_BIN", gh.to_str().unwrap()),
        ("GH_TOKEN", "not-a-real-token"),
    ];
    let req = "{\"schema_version\":1,\"owner\":\"acme\",\"repo\":\"widgets\"}";
    let make = |name: &str| {
        let dir = f.root().join(name);
        fs::create_dir_all(&dir).unwrap();
        dir
    };
    let a = make("a");
    let piped = f.piped(&a, &["forge-read", a.to_str().unwrap()], req, &extra);
    assert_eq!(piped.code, Some(0), "{}", piped.stderr);
    assert_eq!(piped.stderr, "", "no progress byte on a pipe");
    let b = make("b");
    let s = f
        .terminal(&b, &["forge-read", b.to_str().unwrap()], &extra)
        .stdin_text(req)
        .run();
    let term = strip_colors(&s.text);
    assert_eq!(s.code, Some(0), "{term}");
    assert_eq!(
        arrow_lines(&term),
        vec!["-> reading page 2 of /repos/acme/widgets/issues"],
        "{term}"
    );
    assert_eq!(json(&s.stdout)["ok"], true);
    assert_eq!(
        json(&s.stdout)["operation"],
        json(&piped.stdout)["operation"]
    );
}
