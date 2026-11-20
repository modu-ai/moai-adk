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
use std::os::unix::process::CommandExt;
use std::path::{Path, PathBuf};
use std::process::{Child, ChildStdin, Command, ExitStatus, Output, Stdio};
use std::sync::mpsc;
use std::time::{Duration, Instant};

/// What a question ends with.
pub const QUESTION_END: &[u8] = b"[y/N] ";

/// How long a command may run before the helper kills it.
pub const LIMIT: Duration = Duration::from_secs(60);

/// How long the loop waits for more output before it looks at whether the command ended. A
/// short wait keeps a test that starts many commands one after another quick.
const POLL_MILLIS: libc::c_int = 20;

/// End of input typed at a terminal (Ctrl-D).
pub const END_OF_INPUT: u8 = 0x04;

/// How long the helper waits for a command that already ended to let go of its pipes.
const PIPES_CLOSE_WITHIN: Duration = Duration::from_secs(10);

/// A command the helper started, in a process group of its own. Dropping it before the command
/// was seen to end (a failed assertion, a panic, the limit) kills the whole group, so neither the
/// command nor a child it started outlives the test.
struct Running {
    child: Child,
    ended: Option<ExitStatus>,
}

impl Running {
    fn spawn(cmd: &mut Command) -> Running {
        cmd.process_group(0);
        Running {
            child: cmd.spawn().unwrap(),
            ended: None,
        }
    }

    fn kill_group(&mut self) {
        if self.ended.is_none() {
            // The group is the command's own (`process_group(0)`): its id is the command's pid.
            unsafe { libc::kill(-(self.child.id() as libc::pid_t), libc::SIGKILL) };
            let _ = self.child.kill();
            self.ended = self.child.wait().ok();
        }
    }

    /// The exit status once the command ends, `None` at `deadline`.
    fn wait_until(&mut self, deadline: Instant) -> Option<ExitStatus> {
        loop {
            if let Some(status) = self.child.try_wait().unwrap() {
                self.ended = Some(status);
                return Some(status);
            }
            if Instant::now() >= deadline {
                return None;
            }
            std::thread::sleep(Duration::from_millis(10));
        }
    }
}

impl Drop for Running {
    fn drop(&mut self) {
        self.kill_group();
    }
}

/// What a pipe carried, sent when it ends; the thread reads so a full pipe never stops the command.
fn read_thread(mut pipe: impl Read + Send + 'static) -> mpsc::Receiver<Vec<u8>> {
    let (tx, rx) = mpsc::channel();
    std::thread::spawn(move || {
        let mut bytes = Vec::new();
        let _ = pipe.read_to_end(&mut bytes);
        let _ = tx.send(bytes);
    });
    rx
}

/// Writes `text` to the command's stdin on a thread of its own, then ends the input, so a command
/// that writes a lot before it has read everything never waits on the helper. A broken pipe means
/// the command stopped reading; its exit says why.
fn write_thread(mut pipe: ChildStdin, text: Vec<u8>) {
    std::thread::spawn(move || {
        let _ = pipe.write_all(&text);
    });
}

/// What a pipe carried, once the command ended; a pipe still open at `deadline` (a child kept it)
/// kills the group and fails.
fn bytes_of(rx: Option<mpsc::Receiver<Vec<u8>>>, deadline: Instant, run: &mut Running) -> Vec<u8> {
    let Some(rx) = rx else {
        return Vec::new();
    };
    match rx.recv_timeout(deadline.saturating_duration_since(Instant::now())) {
        Ok(bytes) => bytes,
        Err(_) => {
            run.kill_group();
            panic!("the command ended but a stream it started stayed open; the group was killed");
        }
    }
}

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
        let mut run = Running::spawn(&mut cmd);
        let (child_in, child_out, child_err) = (
            run.child.stdin.take(),
            run.child.stdout.take(),
            run.child.stderr.take(),
        );
        // The readers start before the input is written, so output never waits on input.
        let out_rx = child_out.map(read_thread);
        let err_rx = child_err.map(read_thread);
        if let (Some(text), Some(pipe)) = (&self.stdin_text, child_in) {
            write_thread(pipe, text.clone());
        }
        // Our copies of the slave end go with `cmd`: the master reads to its end only when the
        // command and its children have closed theirs.
        drop(cmd);

        let fd = master.as_raw_fd();
        let mut master = File::from(master);
        let mut seen: Vec<u8> = Vec::new();
        let mut typed = false;
        // The limit counts from the start, and again from the answer: a slow command is judged
        // on each stretch, not on the sum.
        let mut deadline = Instant::now() + self.limit;
        let mut buf = [0u8; 8192];
        loop {
            if Instant::now() > deadline {
                run.kill_group();
                panic!(
                    "still running after {:?}, killed; screen so far:\n{}",
                    self.limit,
                    text_of(&seen)
                );
            }
            if !readable(fd, POLL_MILLIS) {
                if run.wait_until(Instant::now()).is_some() {
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
                deadline = Instant::now() + self.limit;
            }
        }
        // The terminal is closed, but the command may still run (it closed its ends and kept
        // going): wait for it within the limit too.
        let Some(status) =
            run.wait_until(deadline.max(Instant::now() + Duration::from_millis(500)))
        else {
            run.kill_group();
            panic!(
                "closed the terminal but was still running after {:?}, killed; screen so far:\n{}",
                self.limit,
                text_of(&seen)
            );
        };
        let pipes_by = Instant::now() + PIPES_CLOSE_WITHIN.min(self.limit);
        let stdout = bytes_of(out_rx, pipes_by, &mut run);
        let stderr = bytes_of(err_rx, pipes_by, &mut run);
        Screen {
            text: text_of(&seen),
            stdout: String::from_utf8_lossy(&stdout).into_owned(),
            stderr: String::from_utf8_lossy(&stderr).into_owned(),
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
    cmd.args(args).clean_env().current_dir(cwd);
    for (k, v) in env {
        cmd.env(k, v);
    }
    let out = output_within(&mut cmd, input.map(str::as_bytes));
    Seen {
        code: out.status.code(),
        stdout: String::from_utf8_lossy(&out.stdout).into_owned(),
        stderr: String::from_utf8_lossy(&out.stderr).into_owned(),
    }
}

/// `Command::output` with a time limit (`LIMIT`, as the Python tests had `timeout=60`): stdin is
/// `input` and then ends (or `/dev/null` without it), stdout and stderr are pipes. A command
/// still running at the limit is killed with its whole process group and the call fails.
pub fn output_within(cmd: &mut Command, input: Option<&[u8]>) -> Output {
    output_within_limit(cmd, input, LIMIT)
}

/// Like `output_within` with a limit of the caller's choosing.
pub fn output_within_limit(cmd: &mut Command, input: Option<&[u8]>, limit: Duration) -> Output {
    cmd.stdin(if input.is_some() {
        Stdio::piped()
    } else {
        Stdio::null()
    })
    .stdout(Stdio::piped())
    .stderr(Stdio::piped());
    let mut run = Running::spawn(cmd);
    let (child_in, child_out, child_err) = (
        run.child.stdin.take(),
        run.child.stdout.take(),
        run.child.stderr.take(),
    );
    let out_rx = child_out.map(read_thread);
    let err_rx = child_err.map(read_thread);
    if let (Some(text), Some(pipe)) = (input, child_in) {
        write_thread(pipe, text.to_vec());
    }
    let Some(status) = run.wait_until(Instant::now() + limit) else {
        run.kill_group();
        panic!("still running after {limit:?}, killed");
    };
    let pipes_by = Instant::now() + PIPES_CLOSE_WITHIN.min(limit);
    let stdout = bytes_of(out_rx, pipes_by, &mut run);
    let stderr = bytes_of(err_rx, pipes_by, &mut run);
    Output {
        status,
        stdout,
        stderr,
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
