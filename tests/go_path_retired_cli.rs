//! SPEC-20261121-go-path-removal revision 6 (REQ-18, REQ-21, REQ-22): what a fresh shell gets
//! when `moai` is reached only through the two links an install lays, and that nothing the
//! product starts is `go` or `moai.old`. Temporary HOME folders and a scrubbed `PATH` only: the
//! real HOME is never read or written.
#![cfg(unix)]

mod common;

use common::home::TempHome;
use serde_json::{json, Value};
use std::ffi::{OsStr, OsString};
use std::fs;
use std::io::Write as _;
use std::os::unix::fs::symlink;
use std::path::{Path, PathBuf};
use std::process::{Command, Output, Stdio};

/// A pack of the user-asset catalog, the one the bundle tests add and remove.
const PACK: &str = "backend";
/// The statusline event of a session in a folder that is no project.
const STATUSLINE_EVENT: &str = r#"{"session_id":"s-fresh-shell","cwd":"/nowhere","model":{"display_name":"Sonnet 5.5"},"effort":{"level":"high"},"workspace":{"current_dir":"/nowhere","project_dir":"/nowhere"}}"#;

/// A temporary HOME whose `moai` is reached through links, a folder outside it for the programs
/// a test plants, and one log where every planted program records its starts.
struct Shell {
    home: TempHome,
    outside: tempfile::TempDir,
    log: PathBuf,
}

impl Shell {
    fn new() -> Shell {
        let home = TempHome::new();
        let outside = tempfile::TempDir::new().unwrap();
        let log = outside.path().join("starts");
        Shell { home, outside, log }
    }

    /// The state folder's `path/`, holding the two links an install lays: `moai` and
    /// `moai-cli`, each to this build.
    fn lay_links(&self) -> PathBuf {
        let dir = self
            .home
            .path()
            .join(moai_cli::home_root::dir(self.home.path()))
            .join("path");
        fs::create_dir_all(&dir).unwrap();
        for name in ["moai", "moai-cli"] {
            symlink(common::MOAI_BIN, dir.join(name)).unwrap();
        }
        dir
    }

    /// A program that records every start with its arguments and fails.
    fn plant(&self, dir: &Path, name: &str, log: &Path) -> PathBuf {
        fs::create_dir_all(dir).unwrap();
        let p = dir.join(name);
        common::write_executable(
            &p,
            &format!("#!/bin/sh\necho \"$*\" >> '{}'\nexit 99\n", log.display()),
        );
        p
    }

    /// `PATH` as the test gives it: the first folder, then the links, then the system folders.
    fn path_env(first: &Path, links: &Path) -> OsString {
        let mut path = OsString::new();
        for dir in [first, links, Path::new("/usr/bin"), Path::new("/bin")] {
            path.push(dir);
            path.push(":");
        }
        path
    }

    /// A command with the clean environment of the HOME and the given `PATH`, run from the
    /// folder outside the HOME.
    fn command_for(&self, program: impl AsRef<OsStr>, args: &[&str], path: &OsStr) -> Command {
        let mut cmd = self.home.command_for(program, args);
        cmd.env("PATH", path)
            .env("LC_ALL", "C")
            .current_dir(self.outside.path());
        cmd
    }

    /// `moai`, looked up on `PATH` the way a shell starts it.
    fn moai(&self, path: &OsStr, args: &[&str], stdin: Option<&str>) -> Output {
        run(self.command_for("moai", args, path), stdin)
    }

    /// Every start the planted programs of this shell recorded, in order.
    fn starts(&self) -> Vec<String> {
        lines(&self.log)
    }
}

fn run(mut cmd: Command, stdin: Option<&str>) -> Output {
    cmd.stdin(if stdin.is_some() {
        Stdio::piped()
    } else {
        Stdio::null()
    })
    .stdout(Stdio::piped())
    .stderr(Stdio::piped());
    let mut child = cmd.spawn().unwrap();
    if let Some(text) = stdin {
        child
            .stdin
            .take()
            .unwrap()
            .write_all(text.as_bytes())
            .unwrap();
    }
    child.wait_with_output().unwrap()
}

fn text(out: &Output) -> String {
    format!(
        "exit {:?}, stdout {:?}, stderr {:?}",
        out.status.code(),
        String::from_utf8_lossy(&out.stdout),
        String::from_utf8_lossy(&out.stderr)
    )
}

fn first_line(out: &Output) -> String {
    String::from_utf8_lossy(&out.stdout)
        .lines()
        .next()
        .unwrap_or("")
        .to_string()
}

/// The JSON envelope a command printed on stdout.
fn envelope(out: &Output) -> Value {
    serde_json::from_slice(&out.stdout)
        .unwrap_or_else(|e| panic!("not an envelope ({e}): {}", text(out)))
}

fn lines(path: &Path) -> Vec<String> {
    fs::read_to_string(path)
        .map(|t| t.lines().map(str::to_string).collect())
        .unwrap_or_default()
}

/// REQ-18, REQ-22: through the links every call reaches the Rust build, and `go` and `moai.old`
/// are never started, though both sit first on `PATH` and one sits in `go/bin` of the HOME.
#[test]
fn a_fresh_shell_reaches_the_rust_build_through_the_links_and_starts_nothing_planted() {
    let sh = Shell::new();
    let links = sh.lay_links();
    // The Go programs a machine of the Go generation has: on `PATH`, ahead of everything else,
    // and in `go/bin` of the HOME, where a routing script looks for them.
    let planted = sh.outside.path().join("planted");
    sh.plant(&planted, "go", &sh.log);
    sh.plant(&planted, "moai.old", &sh.log);
    sh.plant(&sh.home.path().join("go/bin"), "moai.old", &sh.log);
    let path = Shell::path_env(&planted, &links);

    // The version line is this build's.
    let out = sh.moai(&path, &["--version"], None);
    assert_eq!(out.status.code(), Some(0), "{}", text(&out));
    assert_eq!(
        first_line(&out),
        format!("moai {} (rust)", moai_cli::BUILD_VERSION)
    );

    // The bundle verbs that need no project.
    for args in [
        &["bundle", "add", PACK][..],
        &["bundle", "remove", PACK][..],
        &["bundle", "list"][..],
    ] {
        let out = sh.moai(&path, args, None);
        assert_eq!(out.status.code(), Some(0), "{args:?}: {}", text(&out));
        assert_eq!(envelope(&out)["ok"], json!(true), "{args:?}");
    }

    // A word only the Go build had: the Rust build answers and blocks nothing.
    let out = sh.hook_call(&path);
    assert_eq!(out.status.code(), Some(0), "{}", text(&out));
    assert!(out.stdout.is_empty(), "{}", text(&out));

    // The status line.
    let out = sh.moai(&path, &["statusline"], Some(STATUSLINE_EVENT));
    assert_eq!(out.status.code(), Some(0), "{}", text(&out));

    // The two bundle verbs that need a project answer with this build's envelope.
    for verb in ["build", "verify"] {
        let out = sh.moai(&path, &["bundle", verb], None);
        let v = envelope(&out);
        assert_eq!(v["protocol_version"], json!(1), "{verb}: {}", text(&out));
        assert_eq!(v["ok"], json!(false), "{verb}: {}", text(&out));
        assert!(
            v["error"]["message"]
                .as_str()
                .is_some_and(|m| !m.is_empty()),
            "{verb}: {}",
            text(&out)
        );
    }

    // None of the Go programs was started by any of it.
    assert!(sh.starts().is_empty(), "{:?}", sh.starts());
}

impl Shell {
    /// One hook call of the host, with the host's input on stdin.
    fn hook_call(&self, path: &OsStr) -> Output {
        self.moai(path, &["hook", "PreToolUse"], Some("{}"))
    }
}

/// REQ-21: the one place the product starts a `moai` it did not build asks for its version and
/// for nothing else.
#[test]
fn install_inspect_starts_a_moai_it_did_not_build_with_version_alone() {
    let sh = Shell::new();
    let links = sh.lay_links();
    // A `moai` of the Go generation, first on `PATH`: it records every start.
    let fake = sh.outside.path().join("fake");
    let log = sh.outside.path().join("fake-calls");
    fs::create_dir_all(&fake).unwrap();
    common::write_executable(
        &fake.join("moai"),
        &format!(
            "#!/bin/sh\necho \"$*\" >> '{}'\necho 'moai-adk v3.2.0-rc.23'\n",
            log.display()
        ),
    );
    let path = Shell::path_env(&fake, &links);

    // This build by its path, with that `PATH`.
    let out = run(
        sh.command_for(common::MOAI_BIN, &["install-inspect"], &path),
        None,
    );
    let v = envelope(&out);
    assert_eq!(v["ok"], json!(true), "{v}");
    assert_eq!(
        v["result"]["path_first_moai"]["generation"],
        json!("go"),
        "{v}"
    );
    // Every start of the fake asks for its version and for nothing else. One `install-inspect`
    // run asks twice today (its candidates list and its PATH resolution), each with `--version`.
    let calls = lines(&log);
    assert!(!calls.is_empty(), "the fake was never started");
    assert!(
        calls.iter().all(|call| call == "--version"),
        "every start asks for the version alone: {calls:?}"
    );
    assert!(sh.starts().is_empty());
}
