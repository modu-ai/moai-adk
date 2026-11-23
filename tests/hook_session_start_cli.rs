//! SPEC-20261123-hook-session-start: the SessionStart handler (H1) of `moai hook session-context`
//! and the entry contract it closes (a link under HOME is not followed, a closed pipe exits 0, a
//! stdin over the limit is read to its end). Every case runs the real binary in a temporary
//! checkout and HOME; the ledger fixtures are rows inserted into a ledger registered in that HOME.
#![cfg(unix)]
mod common;

use moai_cli::ledger::{self, WriteLedger};
use serde_json::Value;
use std::collections::BTreeMap;
use std::fs;
use std::io::{Read, Write};
use std::os::unix::fs::symlink;
use std::path::{Path, PathBuf};
use std::process::{Command, Stdio};
use std::time::{Duration, Instant};

const PROJECT: &str = "demo";
const RECEIPT: &str = "hook-entry.jsonl";
const BRANCH: &str = "spec/demo";
const STDIN_CAP: usize = 4 * 1024 * 1024;

struct Run {
    /// `None` when the child had to be killed after the wait limit.
    code: Option<i32>,
    stdout: String,
    stderr: String,
    /// The result of writing the whole input to the child's stdin.
    wrote: std::io::Result<()>,
    elapsed: Duration,
}

#[derive(Clone, Copy, PartialEq)]
enum Out {
    Pipe,
    /// A pipe whose reader is gone.
    Closed,
}

struct Fx {
    _tmp: tempfile::TempDir,
    root: PathBuf,
    checkout: PathBuf,
    home: PathBuf,
}

fn project_dir_of(home: &Path) -> PathBuf {
    moai_cli::home_workdir::projects_root(home).join(PROJECT)
}

impl Fx {
    /// A MoAI project that is a git checkout on `BRANCH`, and a HOME holding a registered ledger
    /// (when `registered`).
    fn new(registered: bool) -> Fx {
        let tmp = tempfile::TempDir::new().unwrap();
        let root = fs::canonicalize(tmp.path()).unwrap();
        let checkout = root.join("checkout");
        let home = root.join("home");
        fs::create_dir_all(checkout.join(".moai")).unwrap();
        fs::create_dir_all(checkout.join(".git")).unwrap();
        fs::create_dir_all(&home).unwrap();
        fs::write(
            checkout.join(".moai/project.yaml"),
            format!("{{\"schema_version\": 2, \"project_id\": \"{PROJECT}\"}}\n"),
        )
        .unwrap();
        fs::write(
            checkout.join(".git/HEAD"),
            format!("ref: refs/heads/{BRANCH}\n"),
        )
        .unwrap();
        if registered {
            ledger::register(&home, PROJECT).unwrap();
        }
        Fx {
            _tmp: tmp,
            root,
            checkout,
            home,
        }
    }

    fn head(&self) -> PathBuf {
        self.checkout.join(".git/HEAD")
    }

    fn receipt(&self) -> PathBuf {
        project_dir_of(&self.home).join(RECEIPT)
    }

    fn receipt_lines(&self) -> Vec<Value> {
        match fs::read_to_string(self.receipt()) {
            Ok(text) => text
                .lines()
                .map(|l| serde_json::from_str(l).unwrap())
                .collect(),
            Err(_) => Vec::new(),
        }
    }

    fn write<T>(&self, f: impl FnOnce(&rusqlite::Transaction<'_>, i64) -> T) -> T {
        let mut w: WriteLedger = ledger::open_write(&self.home, PROJECT).unwrap();
        w.with_write_tx(|tx, now| Ok(f(tx, now))).unwrap()
    }

    fn add_card(&self, title: &str) -> i64 {
        self.write(|tx, now| {
            tx.execute(
                "INSERT INTO card (title, state, created_by, created_at, updated_at) \
                 VALUES (?1, 'active', 'test', ?2, ?2)",
                rusqlite::params![title, now],
            )
            .unwrap();
            tx.last_insert_rowid()
        })
    }

    fn reserve(&self, card: i64, status: &str, worktree: Option<&Path>, branch: Option<&str>) {
        let worktree = worktree.map(|p| p.to_str().unwrap().to_owned());
        self.write(|tx, now| {
            tx.execute(
                "INSERT INTO reservation (card_id, generation, holder, status, acquired_at, \
                 lease_until, updated_at, worktree, branch) \
                 VALUES (?1, 1, 'test', ?2, ?3, ?3 + 3600, ?3, ?4, ?5)",
                rusqlite::params![card, status, now, worktree, branch],
            )
            .unwrap();
        });
    }

    /// A card reserved on this checkout by its worktree and branch.
    fn reserved_card(&self, title: &str) -> i64 {
        let card = self.add_card(title);
        self.reserve(card, "held", Some(&self.checkout), Some(BRANCH));
        card
    }

    fn note(&self, card: i64, body: &str) -> i64 {
        self.write(|tx, now| {
            tx.execute(
                "INSERT INTO card_note (card_id, kind, author, body, created_at) \
                 VALUES (?1, 'report', 'test', ?2, ?3)",
                rusqlite::params![card, body, now],
            )
            .unwrap();
            tx.last_insert_rowid()
        })
    }

    fn input(&self, source: &str) -> Vec<u8> {
        serde_json::json!({
            "hook_event_name": "SessionStart",
            "session_id": "s1",
            "cwd": self.checkout.to_str().unwrap(),
            "source": source,
        })
        .to_string()
        .into_bytes()
    }

    /// Runs `moai <args>` with `stdin`, killing it after 20 seconds.
    fn exec(
        &self,
        args: &[&str],
        cwd: &Path,
        stdin: &[u8],
        env: &[(&str, &str)],
        out: Out,
        err: Out,
    ) -> Run {
        fn stdio(out: Out) -> Stdio {
            match out {
                Out::Pipe => Stdio::piped(),
                Out::Closed => {
                    let (reader, writer) = std::io::pipe().unwrap();
                    drop(reader);
                    Stdio::from(writer)
                }
            }
        }
        let mut cmd = Command::new(common::MOAI_BIN);
        cmd.args(args)
            .current_dir(cwd)
            .env("HOME", &self.home)
            .env_remove("CLAUDE_PROJECT_DIR")
            .env("GIT_CONFIG_GLOBAL", "/dev/null")
            .env("GIT_CONFIG_SYSTEM", "/dev/null")
            .stdin(Stdio::piped())
            .stdout(stdio(out))
            .stderr(stdio(err));
        for (k, v) in env {
            cmd.env(k, v);
        }
        let start = Instant::now();
        let mut child = cmd.spawn().unwrap();
        let mut pipe = child.stdin.take().unwrap();
        let bytes = stdin.to_vec();
        let writer = std::thread::spawn(move || pipe.write_all(&bytes));
        let code = loop {
            if let Some(status) = child.try_wait().unwrap() {
                break status.code();
            }
            if start.elapsed() > Duration::from_secs(20) {
                child.kill().unwrap();
                child.wait().unwrap();
                break None;
            }
            std::thread::sleep(Duration::from_millis(5));
        };
        let elapsed = start.elapsed();
        let read = |pipe: Option<&mut dyn Read>| {
            let mut text = String::new();
            if let Some(p) = pipe {
                let mut bytes = Vec::new();
                let _ = p.read_to_end(&mut bytes);
                text = String::from_utf8_lossy(&bytes).into_owned();
            }
            text
        };
        let stdout = read(child.stdout.as_mut().map(|p| p as &mut dyn Read));
        let stderr = read(child.stderr.as_mut().map(|p| p as &mut dyn Read));
        Run {
            code,
            stdout,
            stderr,
            wrote: writer.join().unwrap(),
            elapsed,
        }
    }

    fn hook(&self, name: &str, stdin: &[u8]) -> Run {
        self.exec(
            &["hook", name],
            &self.checkout,
            stdin,
            &[],
            Out::Pipe,
            Out::Pipe,
        )
    }

    /// The context text a `session-context` start with `source` answers; the run must be a clean
    /// exit 0 with one JSON line on stdout.
    fn context(&self, source: &str) -> String {
        let r = self.hook("session-context", &self.input(source));
        context_of(&r)
    }
}

fn context_of(r: &Run) -> String {
    assert_eq!(r.code, Some(0), "stderr: {}", r.stderr);
    assert!(r.stdout.ends_with('\n'), "{:?}", r.stdout);
    assert_eq!(r.stdout.trim_end().lines().count(), 1, "{:?}", r.stdout);
    let v: Value = serde_json::from_str(r.stdout.trim_end()).unwrap();
    assert_eq!(v["hookSpecificOutput"]["hookEventName"], "SessionStart");
    assert_eq!(v.as_object().unwrap().len(), 1, "{v}");
    assert_eq!(v["hookSpecificOutput"].as_object().unwrap().len(), 2, "{v}");
    v["hookSpecificOutput"]["additionalContext"]
        .as_str()
        .unwrap()
        .to_owned()
}

/// Every file under `root` with its bytes; a link is listed by its target text.
fn snapshot(root: &Path) -> BTreeMap<String, Vec<u8>> {
    fn walk(base: &Path, dir: &Path, out: &mut BTreeMap<String, Vec<u8>>) {
        let Ok(entries) = fs::read_dir(dir) else {
            return;
        };
        for entry in entries.flatten() {
            let path = entry.path();
            let rel = path
                .strip_prefix(base)
                .unwrap()
                .to_string_lossy()
                .into_owned();
            let meta = fs::symlink_metadata(&path).unwrap();
            if meta.file_type().is_symlink() {
                let target = fs::read_link(&path).unwrap();
                out.insert(
                    format!("{rel} -> link"),
                    target.to_string_lossy().as_bytes().to_vec(),
                );
            } else if meta.is_dir() {
                out.insert(format!("{rel}/"), Vec::new());
                walk(base, &path, out);
            } else if meta.is_file() {
                out.insert(rel, fs::read(&path).unwrap());
            } else {
                out.insert(format!("{rel} (special)"), Vec::new());
            }
        }
    }
    let mut out = BTreeMap::new();
    walk(root, root, &mut out);
    out
}

fn mkfifo(path: &Path) {
    let c = std::ffi::CString::new(path.to_str().unwrap()).unwrap();
    // SAFETY: `c` is a valid NUL-terminated path for the whole call.
    assert_eq!(unsafe { libc::mkfifo(c.as_ptr(), 0o600) }, 0);
}

const BRANCH_ONLY: &str = "MoAI: branch spec/demo";

/// AC-1: the first line of a start carries the branch and the reserved card, in the JSON shape the
/// host reads; the other two names stay silent.
#[test]
fn ac1_a_start_names_the_branch_and_the_reserved_card_in_one_json_line() {
    let fx = Fx::new(true);
    let card = fx.reserved_card("Hook H1");
    for source in ["startup", "resume", "clear"] {
        let r = fx.hook("session-context", &fx.input(source));
        assert_eq!(r.stderr, "", "{source}");
        let text = context_of(&r);
        assert_eq!(
            text,
            format!("{BRANCH_ONLY} · card C-{card} \"Hook H1\" (active)"),
            "{source}"
        );
    }
    for name in ["bash-boundary", "write-boundary"] {
        let r = fx.hook(name, &fx.input("startup"));
        assert_eq!(
            (r.code, r.stdout.as_str(), r.stderr.as_str()),
            (Some(0), "", ""),
            "{name}"
        );
    }
    assert!(fx.receipt_lines().is_empty());
}

/// AC-2: which reservation counts as "here".
#[test]
fn ac2_the_card_part_follows_the_reservation_of_this_checkout() {
    let other = Path::new("/somewhere/else");
    let none = format!("{BRANCH_ONLY} · no card reserved here");
    // (label, how the reservation is made, expected card part of the first line)
    type Setup = fn(&Fx) -> Option<i64>;
    let cases: [(&str, Setup, bool); 6] = [
        (
            "worktree only",
            |fx| {
                let c = fx.add_card("by root");
                fx.reserve(c, "held", Some(&fx.checkout), Some("other-branch"));
                Some(c)
            },
            true,
        ),
        (
            "branch only",
            |fx| {
                let c = fx.add_card("by branch");
                fx.reserve(c, "held", Some(Path::new("/somewhere/else")), Some(BRANCH));
                Some(c)
            },
            true,
        ),
        (
            "released",
            |fx| {
                let c = fx.add_card("released");
                fx.reserve(c, "released", Some(&fx.checkout), Some(BRANCH));
                Some(c)
            },
            false,
        ),
        (
            "another worktree and branch",
            |fx| {
                let c = fx.add_card("elsewhere");
                fx.reserve(c, "held", Some(Path::new("/somewhere/else")), Some("zz"));
                Some(c)
            },
            false,
        ),
        ("no reservation", |_| None, false),
        (
            "held_for_review counts",
            |fx| {
                let c = fx.add_card("in review");
                fx.reserve(c, "held_for_review", Some(&fx.checkout), None);
                Some(c)
            },
            true,
        ),
    ];
    let _ = other;
    for (label, setup, shown) in cases {
        let fx = Fx::new(true);
        let card = setup(&fx);
        let text = fx.context("startup");
        if shown {
            let c = card.unwrap();
            assert!(
                text.starts_with(&format!("{BRANCH_ONLY} · card C-{c} \"")),
                "{label}: {text}"
            );
        } else {
            assert_eq!(text, none, "{label}");
        }
    }
    // A reservation that names the root wins over one that names only the branch.
    let fx = Fx::new(true);
    let by_branch = fx.add_card("by branch");
    fx.reserve(
        by_branch,
        "held",
        Some(Path::new("/somewhere/else")),
        Some(BRANCH),
    );
    let by_root = fx.add_card("by root");
    fx.reserve(by_root, "held", Some(&fx.checkout), Some("other-branch"));
    assert_eq!(
        fx.context("startup"),
        format!("{BRANCH_ONLY} · card C-{by_root} \"by root\" (active)")
    );
}

/// AC-3: after a compaction the text names the card's newest RESUME note.
#[test]
fn ac3_a_compact_start_names_the_newest_resume_note() {
    let fx = Fx::new(true);
    let card = fx.reserved_card("Hook H1");
    let first = format!("{BRANCH_ONLY} · card C-{card} \"Hook H1\" (active)");
    // No RESUME note yet; notes that only mention it or start differently do not count.
    fx.note(card, "READY something");
    fx.note(card, "see the RESUME note later");
    fx.note(card, "resume lowercase");
    assert_eq!(
        fx.context("compact"),
        format!("{first}\nAfter compaction read the notes of C-{card}: no RESUME note yet")
    );
    let old = fx.note(card, "RESUME C-1 old");
    let newest = fx.note(card, "RESUME C-1 newest\nbody");
    fx.note(card, "LANE-NOTE later than the newest RESUME");
    assert!(newest > old);
    assert_eq!(
        fx.context("compact"),
        format!(
            "{first}\nAfter compaction first read RESUME note #{newest} of C-{card} \
             (moai card show C-{card})"
        )
    );
    for source in ["startup", "resume", "clear"] {
        assert_eq!(fx.context(source), first, "{source}");
    }
    // A compact start without a reserved card has one line.
    let bare = Fx::new(true);
    let other = bare.add_card("not reserved here");
    bare.note(other, "RESUME C-9 elsewhere");
    assert_eq!(
        bare.context("compact"),
        format!("{BRANCH_ONLY} · no card reserved here")
    );
}

/// AC-4: the limits at N and N+1, characters not bytes, no control character reaches the text.
#[test]
fn ac4_limits_hold_at_the_edge_and_control_characters_never_reach_the_text() {
    let cut = |s: &str, limit: usize| -> String {
        // The text is cut to `limit` characters, the trailing `…` included.
        s.chars().take(limit - 1).collect::<String>() + "…"
    };
    for ch in ['a', '가'] {
        let fx = Fx::new(true);
        let t80: String = std::iter::repeat_n(ch, 80).collect();
        let t81: String = std::iter::repeat_n(ch, 81).collect();
        let c80 = fx.add_card(&t80);
        fx.reserve(c80, "held", Some(&fx.checkout), None);
        let text = fx.context("startup");
        assert_eq!(
            text,
            format!("{BRANCH_ONLY} · card C-{c80} \"{t80}\" (active)"),
            "title of 80 {ch}"
        );
        let fx = Fx::new(true);
        let c81 = fx.add_card(&t81);
        fx.reserve(c81, "held", Some(&fx.checkout), None);
        assert_eq!(
            fx.context("startup"),
            format!(
                "{BRANCH_ONLY} · card C-{c81} \"{}\" (active)",
                cut(&t81, 80)
            ),
            "title of 81 {ch}"
        );
        // The branch: 100 characters whole, 101 cut.
        for (n, whole) in [(100usize, true), (101, false)] {
            let fx = Fx::new(true);
            let name: String = std::iter::repeat_n(ch, n).collect();
            fs::write(fx.head(), format!("ref: refs/heads/{name}\n")).unwrap();
            let shown = if whole { name.clone() } else { cut(&name, 100) };
            assert_eq!(
                fx.context("startup"),
                format!("MoAI: branch {shown} · no card reserved here"),
                "branch of {n} {ch}"
            );
        }
    }
    // Line feeds, tabs and escape characters: whitespace collapses, other controls are dropped.
    let fx = Fx::new(true);
    fs::write(fx.head(), "ref: refs/heads/ab\u{1b}[31mc\td\r\n").unwrap();
    let card = fx.add_card("x\ny\u{1b}z\t\tw\u{7}  v");
    fx.reserve(card, "held", Some(&fx.checkout), None);
    let text = fx.context("startup");
    assert_eq!(
        text,
        format!("MoAI: branch ab[31mc d · card C-{card} \"x yz w v\" (active)")
    );
    // Everything at its limit, on a compact start: well inside 600 characters, no control
    // character but the line feed between the two lines.
    let fx = Fx::new(true);
    let name: String = std::iter::repeat_n('가', 200).collect();
    fs::write(fx.head(), format!("ref: refs/heads/{name}\n")).unwrap();
    let card = fx.add_card(&"가".repeat(300));
    fx.reserve(card, "held", Some(&fx.checkout), None);
    fx.note(card, "RESUME x");
    let text = fx.context("compact");
    assert!(text.chars().count() <= 600, "{}", text.chars().count());
    assert_eq!(text.lines().count(), 2);
    assert!(text.chars().all(|c| !c.is_control() || c == '\n'));
}

/// AC-5: where the branch comes from.
#[test]
fn ac5_the_branch_is_read_from_head_files_only() {
    // A plain checkout, the input's cwd below the root, and no cwd in the input.
    let fx = Fx::new(true);
    let sub = fx.checkout.join("src/deep");
    fs::create_dir_all(&sub).unwrap();
    let below = serde_json::json!({"hook_event_name": "SessionStart", "cwd": sub.to_str().unwrap(),
        "source": "startup"})
    .to_string();
    let r = fx.hook("session-context", below.as_bytes());
    assert_eq!(
        context_of(&r),
        format!("{BRANCH_ONLY} · no card reserved here")
    );
    let bare = br#"{"hook_event_name":"SessionStart","source":"startup"}"#;
    // The host sets `CLAUDE_PROJECT_DIR`; the process works in a folder below it.
    let r = fx.exec(
        &["hook", "session-context"],
        &sub,
        bare,
        &[("CLAUDE_PROJECT_DIR", fx.checkout.to_str().unwrap())],
        Out::Pipe,
        Out::Pipe,
    );
    assert_eq!(
        context_of(&r),
        format!("{BRANCH_ONLY} · no card reserved here")
    );

    // A linked worktree: `.git` is a file naming the git folder, absolute or relative.
    for relative in [false, true] {
        let fx = Fx::new(true);
        fs::remove_dir_all(fx.checkout.join(".git")).unwrap();
        let gitdir = fx.root.join("gitdirs/wt");
        fs::create_dir_all(&gitdir).unwrap();
        fs::write(gitdir.join("HEAD"), "ref: refs/heads/lane/x\n").unwrap();
        let line = if relative {
            "gitdir: ../gitdirs/wt\n".to_owned()
        } else {
            format!("gitdir: {}\n", gitdir.display())
        };
        fs::write(fx.checkout.join(".git"), line).unwrap();
        assert_eq!(
            fx.context("startup"),
            "MoAI: branch lane/x · no card reserved here",
            "relative {relative}"
        );
    }

    // A detached head shows 12 hex digits.
    let fx = Fx::new(true);
    fs::write(fx.head(), "0123456789abcdef0123456789abcdef01234567\n").unwrap();
    assert_eq!(
        fx.context("startup"),
        "MoAI: detached 0123456789ab · no card reserved here"
    );

    // Unreadable heads read as unknown: missing, a folder, a FIFO (no hang), no `.git` at all.
    let unknown = "MoAI: branch unknown · no card reserved here";
    let fx = Fx::new(true);
    fs::remove_file(fx.head()).unwrap();
    assert_eq!(fx.context("startup"), unknown, "missing");
    fs::create_dir(fx.head()).unwrap();
    assert_eq!(fx.context("startup"), unknown, "folder");
    fs::remove_dir(fx.head()).unwrap();
    mkfifo(&fx.head());
    assert_eq!(fx.context("startup"), unknown, "FIFO");
    fs::remove_file(fx.head()).unwrap();
    fs::remove_dir(fx.checkout.join(".git")).unwrap();
    assert_eq!(fx.context("startup"), unknown, "no .git");

    // The `HEAD` file is read up to 1024 bytes: 1024 is read, 1025 is unknown.
    for (bytes, readable) in [(1024usize, true), (1025, false)] {
        let fx = Fx::new(true);
        fs::create_dir_all(fx.checkout.join(".git")).unwrap();
        let name = "b".repeat(bytes - "ref: refs/heads/".len() - 1);
        fs::write(fx.head(), format!("ref: refs/heads/{name}\n")).unwrap();
        assert_eq!(fs::metadata(fx.head()).unwrap().len() as usize, bytes);
        let text = fx.context("startup");
        if readable {
            let shown = format!("{}…", "b".repeat(99));
            assert_eq!(
                text,
                format!("MoAI: branch {shown} · no card reserved here")
            );
        } else {
            assert_eq!(text, unknown, "{bytes} bytes");
        }
    }
}

/// AC-6: the handler starts no process, writes nothing and reads only the two variables.
#[test]
fn ac6_no_process_no_write_and_two_variables() {
    let fx = Fx::new(true);
    let card = fx.reserved_card("Hook H1");
    fx.note(card, "RESUME C-1 note");
    let bin = fx.root.join("fakebin");
    fs::create_dir_all(&bin).unwrap();
    let marker = fx.root.join("git-was-started");
    common::write_executable(
        &bin.join("git"),
        &format!("#!/bin/sh\ntouch '{}'\nexit 0\n", marker.display()),
    );
    let before = snapshot(&fx.root);
    for source in ["startup", "compact"] {
        let r = fx.exec(
            &["hook", "session-context"],
            &fx.checkout,
            &fx.input(source),
            &[("PATH", bin.to_str().unwrap())],
            Out::Pipe,
            Out::Pipe,
        );
        let text = context_of(&r);
        assert!(text.contains(&format!("card C-{card}")), "{text}");
    }
    assert!(!marker.exists(), "a git process was started");
    let after = snapshot(&fx.root);
    assert_eq!(before.len(), after.len(), "a file appeared or went");
    assert!(before == after, "a file changed");

    let source = include_str!("../src/hook_session_start.rs");
    for word in [
        "Command::new",
        "std::process",
        "TcpStream",
        "UdpSocket",
        "reqwest",
        "ureq",
        "hyper",
        "settings.json",
        "settings.local.json",
        "ORCA",
        "std::env",
        "var_os(",
        "env::var",
        ".spawn(",
    ] {
        assert!(
            !source.contains(word),
            "src/hook_session_start.rs names `{word}`"
        );
    }
    let entry = include_str!("../src/hook_entry.rs");
    let mut variables = Vec::new();
    for marker in ["var_os(\"", "var(\""] {
        let mut rest = entry;
        while let Some(at) = rest.find(marker) {
            rest = &rest[at + marker.len()..];
            variables.push(rest[..rest.find('"').unwrap()].to_owned());
        }
    }
    variables.sort();
    variables.dedup();
    assert_eq!(variables, ["CLAUDE_PROJECT_DIR", "HOME"]);
}

/// AC-7: a ledger that cannot be read still gives the branch, one receipt line and exit 0.
#[test]
fn ac7_an_unreadable_ledger_gives_the_branch_and_one_receipt_line() {
    let ledger_file = |fx: &Fx| ledger::location::ledger_path(&fx.home, PROJECT).unwrap();
    let reason_of = |fx: &Fx| -> Vec<String> {
        fx.receipt_lines()
            .iter()
            .map(|v| v["reason"].as_str().unwrap().to_owned())
            .collect()
    };

    // Absent: the folder is registered, the file is not there.
    let fx = Fx::new(true);
    fs::remove_file(ledger_file(&fx)).unwrap();
    let r = fx.hook("session-context", &fx.input("startup"));
    assert_eq!(r.stderr, "");
    assert_eq!(context_of(&r), BRANCH_ONLY);
    assert_eq!(reason_of(&fx), ["ledger_absent"]);
    let line = &fx.receipt_lines()[0];
    assert_eq!(line["name"], "session-context");

    // Damaged: bytes that are no database, left as they were.
    let fx = Fx::new(true);
    let junk = vec![7u8; 8192];
    fs::write(ledger_file(&fx), &junk).unwrap();
    let r = fx.hook("session-context", &fx.input("startup"));
    assert_eq!(context_of(&r), BRANCH_ONLY);
    assert_eq!(reason_of(&fx), ["ledger_unreadable"]);
    assert_eq!(fs::read(ledger_file(&fx)).unwrap(), junk);

    // Busy: another connection holds the exclusive lock; the start answers within 3 seconds.
    let fx = Fx::new(true);
    fx.reserved_card("Hook H1");
    let holder = rusqlite::Connection::open(ledger_file(&fx)).unwrap();
    holder.execute_batch("BEGIN EXCLUSIVE").unwrap();
    let r = fx.hook("session-context", &fx.input("startup"));
    assert!(r.elapsed < Duration::from_secs(3), "{:?}", r.elapsed);
    assert_eq!(context_of(&r), BRANCH_ONLY);
    assert_eq!(reason_of(&fx), ["ledger_busy"]);
    holder.execute_batch("ROLLBACK").unwrap();
    drop(holder);
    // The same start after the lock is gone shows the card again, with no new receipt line.
    assert!(fx.context("startup").contains("card C-1"));
    assert_eq!(reason_of(&fx), ["ledger_busy"]);

    // No registered project folder: the branch, no folder made, one stderr line instead.
    let fx = Fx::new(false);
    let before = snapshot(&fx.home);
    let r = fx.hook("session-context", &fx.input("startup"));
    assert_eq!(r.stderr.lines().count(), 1, "{:?}", r.stderr);
    assert_eq!(context_of(&r), BRANCH_ONLY);
    assert_eq!(before, snapshot(&fx.home));
}

/// AC-8: a link at `.moai`, `projects` or the project folder is not followed, for the ledger or
/// for a receipt; a receipt file that is a link or a FIFO is refused.
#[test]
fn ac8_links_under_home_are_not_followed() {
    // A second HOME that holds a valid ledger with the card reserved on the checkout and a
    // receipt file: reading or writing through a link would show or change it.
    let make_target = |fx: &Fx| -> PathBuf {
        let target_home = fx.root.join("target-home");
        fs::create_dir_all(&target_home).unwrap();
        ledger::register(&target_home, PROJECT).unwrap();
        let mut w = ledger::open_write(&target_home, PROJECT).unwrap();
        w.with_write_tx(|tx, now| {
            tx.execute(
                "INSERT INTO card (title, state, created_by, created_at, updated_at) \
                 VALUES ('behind a link', 'active', 'test', ?1, ?1)",
                [now],
            )?;
            tx.execute(
                "INSERT INTO reservation (card_id, generation, holder, status, acquired_at, \
                 lease_until, updated_at, worktree, branch) \
                 VALUES (1, 1, 'test', 'held', ?1, ?1 + 3600, ?1, ?2, NULL)",
                rusqlite::params![now, fx.checkout.to_str().unwrap()],
            )?;
            Ok(())
        })
        .unwrap();
        fs::write(project_dir_of(&target_home).join(RECEIPT), "{\"old\":1}\n").unwrap();
        target_home
    };
    type Link = fn(&Fx, &Path);
    let links: [(&str, Link); 3] = [
        ("moai", |fx, target_home| {
            fs::remove_dir_all(fx.home.join(".moai")).unwrap();
            symlink(target_home.join(".moai"), fx.home.join(".moai")).unwrap();
        }),
        ("projects", |fx, target_home| {
            fs::remove_dir_all(fx.home.join(".moai/projects")).unwrap();
            symlink(
                target_home.join(".moai/projects"),
                fx.home.join(".moai/projects"),
            )
            .unwrap();
        }),
        ("project", |fx, target_home| {
            fs::remove_dir_all(project_dir_of(&fx.home)).unwrap();
            symlink(project_dir_of(target_home), project_dir_of(&fx.home)).unwrap();
        }),
    ];
    for (what, link) in links {
        let fx = Fx::new(true);
        let target_home = make_target(&fx);
        link(&fx, &target_home);
        let before = snapshot(&target_home);
        // The start does not read the ledger behind the link.
        let r = fx.hook("session-context", &fx.input("startup"));
        assert_eq!(r.stderr.lines().count(), 1, "{what}: {:?}", r.stderr);
        assert_eq!(context_of(&r), BRANCH_ONLY, "{what}");
        // A bad input does not write a receipt behind the link.
        let r = fx.hook("bash-boundary", b"not json");
        assert_eq!(r.code, Some(0), "{what}");
        assert_eq!(r.stdout, "", "{what}");
        assert_eq!(r.stderr.lines().count(), 1, "{what}: {:?}", r.stderr);
        assert_eq!(before, snapshot(&target_home), "{what}: the target changed");
    }

    // The receipt file itself: a link to a file outside, and a FIFO.
    let fx = Fx::new(true);
    let outside = fx.root.join("outside");
    fs::write(&outside, "original").unwrap();
    symlink(&outside, fx.receipt()).unwrap();
    let r = fx.hook("write-boundary", b"not json");
    assert_eq!(
        (r.code, r.stderr.lines().count()),
        (Some(0), 1),
        "{:?}",
        r.stderr
    );
    assert_eq!(fs::read_to_string(&outside).unwrap(), "original");
    fs::remove_file(fx.receipt()).unwrap();
    mkfifo(&fx.receipt());
    let r = fx.hook("write-boundary", b"not json");
    assert_eq!(
        (r.code, r.stderr.lines().count()),
        (Some(0), 1),
        "{:?}",
        r.stderr
    );
    assert!(r.elapsed < Duration::from_secs(10));
}

/// AC-9: a closed standard error or standard output never changes the exit code.
#[test]
fn ac9_a_closed_pipe_still_exits_zero() {
    // Standard error closed, and the run has a line to write there (no registered folder).
    let fx = Fx::new(false);
    let r = fx.exec(
        &["hook", "write-boundary"],
        &fx.checkout,
        b"not json",
        &[],
        Out::Pipe,
        Out::Closed,
    );
    assert_eq!(r.code, Some(0));
    // Standard output closed, and the run has a text to write there.
    let fx = Fx::new(true);
    fx.reserved_card("Hook H1");
    let r = fx.exec(
        &["hook", "session-context"],
        &fx.checkout,
        &fx.input("startup"),
        &[],
        Out::Closed,
        Out::Pipe,
    );
    assert_eq!(r.code, Some(0), "{}", r.stderr);
    // Both closed.
    let r = fx.exec(
        &["hook", "session-context"],
        &fx.checkout,
        b"not json",
        &[],
        Out::Closed,
        Out::Closed,
    );
    assert_eq!(r.code, Some(0));
}

/// AC-10: an input over the limit is read to its end, so the writer is never cut off.
#[test]
fn ac10_an_input_over_the_limit_is_read_to_its_end() {
    let fx = Fx::new(true);
    let mut big = br#"{"k":""#.to_vec();
    big.extend(std::iter::repeat_n(b'x', 16 * 1024 * 1024));
    big.extend_from_slice(br#""}"#);
    assert!(big.len() > STDIN_CAP);
    let r = fx.hook("session-context", &big);
    assert_eq!(r.code, Some(0), "{}", r.stderr);
    assert_eq!(r.stdout, "");
    r.wrote
        .expect("the writer was cut off before its input was read");
    let lines = fx.receipt_lines();
    assert_eq!(lines.len(), 1, "{lines:?}");
    assert_eq!(lines[0]["reason"], "stdin_too_large");
}
