//! A command at a terminal, for the integration tests that need a person's screen.
//!
//! One pseudo terminal made with `libc::openpty` (the same call `tests/help_layout.rs` makes) is
//! the command's stdin, stdout and stderr, the way a person runs it, so the same cases run on
//! Linux and macOS (BSD `script` took other arguments than util-linux's and only macOS ran
//! them). The answer is typed only after a question (`[y/N] `) ends the screen, as a person
//! would, so the echo of the typed text never comes before the question.
#![allow(dead_code)] // every test crate includes `common`; only the terminal tests run a command here

use super::CleanEnv;
use std::ffi::OsStr;
use std::fs::File;
use std::io::{Read, Write};
use std::os::fd::{AsRawFd, FromRawFd, OwnedFd};
use std::path::{Path, PathBuf};
use std::process::{Command, Stdio};
use std::time::{Duration, Instant};

/// What a question ends with.
pub const QUESTION_END: &[u8] = b"[y/N] ";

/// How long a command may run before the helper kills it.
pub const LIMIT: Duration = Duration::from_secs(60);

/// End of input typed at a terminal (Ctrl-D).
pub const END_OF_INPUT: u8 = 0x04;

/// What the terminal showed and how the command ended.
pub struct Screen {
    /// Everything the command wrote to the terminal, with `\r\n` read as `\n`.
    pub text: String,
    /// What it wrote to stdout when stdout was a pipe (`stdout_not_a_terminal`), else empty.
    pub stdout: String,
    /// What it wrote to stderr when stderr was a pipe (`stderr_not_a_terminal`), else empty.
    pub stderr: String,
    /// The exit code, `None` when a signal ended the command.
    pub code: Option<i32>,
}

/// A command to run at a terminal.
pub struct Terminal {
    exe: PathBuf,
    args: Vec<String>,
    env: Vec<(String, String)>,
    cwd: Option<PathBuf>,
    answer: Option<Vec<u8>>,
    before_answer: Option<Box<dyn Fn()>>,
    stdin_is_terminal: bool,
    stdin_text: Option<Vec<u8>>,
    stdout_is_terminal: bool,
    stderr_is_terminal: bool,
    limit: Duration,
}

impl Terminal {
    pub fn new(exe: impl AsRef<Path>) -> Terminal {
        Terminal {
            exe: exe.as_ref().to_path_buf(),
            args: Vec::new(),
            env: Vec::new(),
            cwd: None,
            answer: None,
            before_answer: None,
            stdin_is_terminal: true,
            stdin_text: None,
            stdout_is_terminal: true,
            stderr_is_terminal: true,
            limit: LIMIT,
        }
    }

    pub fn args<I, S>(mut self, args: I) -> Terminal
    where
        I: IntoIterator<Item = S>,
        S: AsRef<OsStr>,
    {
        self.args.extend(
            args.into_iter()
                .map(|a| a.as_ref().to_string_lossy().into_owned()),
        );
        self
    }

    /// One variable on top of the clean environment the command starts with.
    pub fn env(mut self, key: impl AsRef<OsStr>, value: impl AsRef<OsStr>) -> Terminal {
        self.env.push((
            key.as_ref().to_string_lossy().into_owned(),
            value.as_ref().to_string_lossy().into_owned(),
        ));
        self
    }

    pub fn cwd(mut self, dir: impl AsRef<Path>) -> Terminal {
        self.cwd = Some(dir.as_ref().to_path_buf());
        self
    }

    /// Typed once the first question is on screen. Without it, the helper types end of input
    /// at a question, so a command that asks when it should not ends with the question in its
    /// text instead of waiting for the limit.
    pub fn answer(mut self, bytes: impl AsRef<[u8]>) -> Terminal {
        self.answer = Some(bytes.as_ref().to_vec());
        self
    }

    /// Runs once the question is on screen, before the answer is typed.
    pub fn before_answer(mut self, f: impl Fn() + 'static) -> Terminal {
        self.before_answer = Some(Box::new(f));
        self
    }

    /// stdin is `/dev/null`; stdout and stderr stay on the terminal.
    pub fn stdin_not_a_terminal(mut self) -> Terminal {
        self.stdin_is_terminal = false;
        self
    }

    /// stdin is a pipe that carries `text` and then ends; stdout and stderr stay on the terminal
    /// unless a case moves them. The text is written once the command runs.
    pub fn stdin_text(mut self, text: impl AsRef<[u8]>) -> Terminal {
        self.stdin_text = Some(text.as_ref().to_vec());
        self
    }

    /// stdout is a pipe the helper reads (`Screen::stdout`); stderr stays on the terminal.
    pub fn stdout_not_a_terminal(mut self) -> Terminal {
        self.stdout_is_terminal = false;
        self
    }

    /// stderr is a pipe the helper reads (`Screen::stderr`); stdout stays on the terminal.
    pub fn stderr_not_a_terminal(mut self) -> Terminal {
        self.stderr_is_terminal = false;
        self
    }

    pub fn limit(mut self, limit: Duration) -> Terminal {
        self.limit = limit;
        self
    }

    /// Runs the command and returns its screen. A command still running at the limit is killed
    /// and the call fails with the screen seen so far.
    pub fn run(self) -> Screen {
        let (master, slave) = open_pty();
        let mut cmd = Command::new(&self.exe);
        cmd.args(&self.args).clean_env();
        for (k, v) in &self.env {
            cmd.env(k, v);
        }
        if let Some(dir) = &self.cwd {
            cmd.current_dir(dir);
        }
        let stdin = if self.stdin_text.is_some() {
            Stdio::piped()
        } else if self.stdin_is_terminal {
            Stdio::from(slave.try_clone().unwrap())
        } else {
            Stdio::null()
        };
        let stdout = if self.stdout_is_terminal {
            Stdio::from(slave.try_clone().unwrap())
        } else {
            Stdio::piped()
        };
        let stderr = if self.stderr_is_terminal {
            Stdio::from(slave)
        } else {
            Stdio::piped()
        };
        cmd.stdin(stdin).stdout(stdout).stderr(stderr);
        let mut child = cmd.spawn().unwrap();
        if let (Some(text), Some(mut pipe)) = (&self.stdin_text, child.stdin.take()) {
            // A broken pipe means the command stopped reading; its exit says why.
            let _ = pipe.write_all(text);
        }
        // A pipe is read on its own thread so a full pipe never stops the command.
        let pipe_reader = |pipe: Option<Box<dyn Read + Send>>| {
            pipe.map(|mut p| {
                std::thread::spawn(move || {
                    let mut bytes = Vec::new();
                    let _ = p.read_to_end(&mut bytes);
                    bytes
                })
            })
        };
        let out_thread = pipe_reader(
            child
                .stdout
                .take()
                .map(|p| Box::new(p) as Box<dyn Read + Send>),
        );
        let err_thread = pipe_reader(
            child
                .stderr
                .take()
                .map(|p| Box::new(p) as Box<dyn Read + Send>),
        );
        // Our copies of the slave end go with `cmd`: the master reads to its end only when the
        // command and its children have closed theirs.
        drop(cmd);

        let fd = master.as_raw_fd();
        let mut master = File::from(master);
        let mut seen: Vec<u8> = Vec::new();
        let mut typed = false;
        let started = Instant::now();
        let mut buf = [0u8; 8192];
        loop {
            if started.elapsed() > self.limit {
                let _ = child.kill();
                let _ = child.wait();
                panic!(
                    "still running after {:?}, killed; screen so far:\n{}",
                    self.limit,
                    text_of(&seen)
                );
            }
            if !readable(fd, 200) {
                if child.try_wait().unwrap().is_some() {
                    // Exited without more output: take what is left, then stop.
                    while readable(fd, 0) {
                        match master.read(&mut buf) {
                            Ok(0) | Err(_) => break,
                            Ok(n) => seen.extend_from_slice(&buf[..n]),
                        }
                    }
                    break;
                }
                continue;
            }
            match master.read(&mut buf) {
                // Linux reports the end of a closed terminal as an error (EIO), macOS as 0.
                Ok(0) | Err(_) => break,
                Ok(n) => seen.extend_from_slice(&buf[..n]),
            }
            if !typed && seen.ends_with(QUESTION_END) {
                if let Some(f) = &self.before_answer {
                    f();
                }
                let bytes = self.answer.clone().unwrap_or_else(|| vec![END_OF_INPUT]);
                master.write_all(&bytes).unwrap();
                typed = true;
            }
        }
        let status = child.wait().unwrap();
        let joined = |t: Option<std::thread::JoinHandle<Vec<u8>>>| {
            t.map(|t| String::from_utf8_lossy(&t.join().unwrap()).into_owned())
                .unwrap_or_default()
        };
        Screen {
            text: text_of(&seen),
            stdout: joined(out_thread),
            stderr: joined(err_thread),
            code: status.code(),
        }
    }
}

/// What a command carried on stdout and stderr, whether each stream was a pipe or the terminal.
#[derive(Debug, PartialEq, Eq)]
pub struct Seen {
    pub code: Option<i32>,
    pub stdout: String,
    pub stderr: String,
}

/// An environment as name and value pairs.
pub type Env = Vec<(String, String)>;

/// `base` with each of `extra` set on top; an `extra` name replaces a `base` one.
pub fn env_with(base: &[(&str, &str)], extra: &[(&str, &str)]) -> Env {
    let mut env: Env = base
        .iter()
        .map(|(k, v)| (k.to_string(), v.to_string()))
        .collect();
    for (k, v) in extra {
        env.retain(|(name, _)| name != k);
        env.push((k.to_string(), v.to_string()));
    }
    env
}

/// Runs `exe` with stdin `/dev/null` and both streams as pipes, on a clean environment.
pub fn piped(exe: impl AsRef<Path>, cwd: &Path, env: &Env, args: &[&str]) -> Seen {
    piped_with(exe, cwd, env, args, None)
}

/// Like `piped`, with `input` on stdin when given (stdin ends after it).
pub fn piped_with(
    exe: impl AsRef<Path>,
    cwd: &Path,
    env: &Env,
    args: &[&str],
    input: Option<&str>,
) -> Seen {
    let mut cmd = Command::new(exe.as_ref());
    cmd.args(args)
        .clean_env()
        .current_dir(cwd)
        .stdin(if input.is_some() {
            Stdio::piped()
        } else {
            Stdio::null()
        })
        .stdout(Stdio::piped())
        .stderr(Stdio::piped());
    for (k, v) in env {
        cmd.env(k, v);
    }
    let mut child = cmd.spawn().unwrap();
    if let (Some(text), Some(mut pipe)) = (input, child.stdin.take()) {
        // A broken pipe means the command stopped reading; its exit says why.
        let _ = pipe.write_all(text.as_bytes());
    }
    let out = child.wait_with_output().unwrap();
    Seen {
        code: out.status.code(),
        stdout: String::from_utf8_lossy(&out.stdout).into_owned(),
        stderr: String::from_utf8_lossy(&out.stderr).into_owned(),
    }
}

fn terminal_of(exe: impl AsRef<Path>, cwd: &Path, env: &Env, args: &[&str]) -> Terminal {
    let mut term = Terminal::new(exe)
        .args(args)
        .cwd(cwd)
        .stdin_not_a_terminal();
    for (k, v) in env {
        term = term.env(k, v);
    }
    term
}

/// stderr on a terminal, stdin `/dev/null`, stdout a pipe; `Seen::stderr` is the screen.
pub fn stderr_on_terminal(exe: impl AsRef<Path>, cwd: &Path, env: &Env, args: &[&str]) -> Seen {
    let s = terminal_of(exe, cwd, env, args)
        .stdout_not_a_terminal()
        .run();
    Seen {
        code: s.code,
        stdout: s.stdout,
        stderr: s.text,
    }
}

/// stdout on a terminal, stdin `/dev/null`, stderr a pipe; `Seen::stdout` is the screen.
pub fn stdout_on_terminal(exe: impl AsRef<Path>, cwd: &Path, env: &Env, args: &[&str]) -> Seen {
    let s = terminal_of(exe, cwd, env, args)
        .stderr_not_a_terminal()
        .run();
    Seen {
        code: s.code,
        stdout: s.text,
        stderr: s.stderr,
    }
}

fn text_of(bytes: &[u8]) -> String {
    String::from_utf8_lossy(bytes).replace("\r\n", "\n")
}

/// The two ends of a new pseudo terminal, owned.
pub fn open_pty() -> (OwnedFd, OwnedFd) {
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
    unsafe { (OwnedFd::from_raw_fd(master), OwnedFd::from_raw_fd(slave)) }
}

/// Whether `fd` has something to read (or has ended) within `millis`.
fn readable(fd: libc::c_int, millis: libc::c_int) -> bool {
    let mut p = libc::pollfd {
        fd,
        events: libc::POLLIN,
        revents: 0,
    };
    let n = unsafe { libc::poll(&mut p, 1, millis) };
    n > 0 && p.revents != 0
}
