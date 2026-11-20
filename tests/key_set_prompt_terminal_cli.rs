//! `key-set PROVIDER --prompt` under a real terminal (card C-54 S1).
//!
//! The value is typed at a hidden prompt: the terminal's echo is off while it is read, and its
//! modes (echo and line mode) are back at every exit (a stored value, a refusal, Ctrl-C, SIGTERM,
//! SIGHUP, SIGQUIT). Neither the terminal nor stdout ever shows the value, and nothing typed or
//! pasted is left in the terminal for the shell. One line is one value: anything after the line
//! end refuses the whole input. Without a terminal on stdin and on stderr `--prompt` is refused
//! and nothing is read or stored.
//!
//! Every case uses a temporary HOME, a temporary working directory and a synthetic value made at
//! run time; the real home and any real key are never read. Moved here from the Python module
//! `key_set_prompt_tty_test.py` (SPEC-20261120-python-binary-tests-to-rust).
//!
//! The command runs under a stand-in for the shell a person runs `moai` from. That shell leads the
//! session that owns the terminal and outlives the command, so the terminal can be looked at after
//! the command ended (a terminal whose session leader exits is taken away from everyone). It
//! starts the command in its own foreground group, reports the command's pid and then its exit
//! code, and stays until told to go. The stand-in is this test program run again on one test
//! (`prompt_shell`), which does nothing unless the case that started it has set the variables it
//! reads.

#![cfg(unix)]

mod common;
use common::pty::{env_with, open_pty, piped_with, Env};
use common::CleanEnv;
use std::{
    collections::VecDeque,
    fs,
    io::Read,
    os::{
        fd::{AsRawFd, FromRawFd, OwnedFd},
        unix::process::{CommandExt, ExitStatusExt},
    },
    path::PathBuf,
    process::{Command, Stdio},
    time::{Duration, Instant},
};
use tempfile::TempDir;

const BIN: &str = common::MOAI_CLI_BIN;
const PROMPT_END: &[u8] = b": ";
/// The signals a terminal sends and the stand-in shell handles itself.
const TERMINAL_SIGNALS: [libc::c_int; 5] = [
    libc::SIGINT,
    libc::SIGQUIT,
    libc::SIGTERM,
    libc::SIGHUP,
    libc::SIGTSTP,
];

// ---- the stand-in shell ----

/// Runs once as a test; does nothing unless a case started this program as the stand-in shell.
#[test]
fn prompt_shell() {
    let Ok(ignored) = std::env::var("KEY_PROMPT_IGNORED") else {
        return;
    };
    let exe = std::env::var("KEY_PROMPT_EXE").unwrap();
    let argv = std::env::var("KEY_PROMPT_ARGV").unwrap();
    let ignored: Vec<libc::c_int> = ignored
        .split(',')
        .filter(|s| !s.is_empty())
        .map(|s| s.parse().unwrap())
        .collect();
    // A typed Ctrl-C goes to the whole foreground group: caught here (a handler, which the
    // command does not inherit), so only the command ends.
    extern "C" fn nothing(_: libc::c_int) {}
    for sig in TERMINAL_SIGNALS {
        unsafe {
            libc::signal(
                sig,
                nothing as extern "C" fn(libc::c_int) as libc::sighandler_t,
            )
        };
    }
    let mut cmd = Command::new(exe);
    cmd.args(argv.split('\u{1f}').filter(|a| !a.is_empty()))
        .clean_env();
    for (k, v) in std::env::vars() {
        if !k.starts_with("KEY_PROMPT_") {
            cmd.env(k, v);
        }
    }
    // The command's stdout is the pipe the case reads (descriptor 5); its stdin and stderr are
    // this program's, the terminal or nothing.
    cmd.stdout(unsafe { Stdio::from(OwnedFd::from_raw_fd(5)) });
    unsafe {
        cmd.pre_exec(move || {
            for sig in TERMINAL_SIGNALS {
                let how = if ignored.contains(&sig) {
                    libc::SIG_IGN
                } else {
                    libc::SIG_DFL
                };
                libc::signal(sig, how);
            }
            Ok(())
        });
    }
    let mut child = cmd.spawn().unwrap();
    write_fd(3, format!("{}\n", child.id()).as_bytes());
    let status = child.wait().unwrap();
    let code = status
        .code()
        .unwrap_or_else(|| -status.signal().expect("a code or a signal"));
    write_fd(3, format!("{code}\n").as_bytes());
    let mut byte = [0u8; 1];
    unsafe { libc::read(4, byte.as_mut_ptr().cast(), 1) };
}

fn write_fd(fd: libc::c_int, bytes: &[u8]) {
    unsafe { libc::write(fd, bytes.as_ptr().cast(), bytes.len()) };
}

// ---- looking at the terminal ----

const LFLAG_ECHO_AND_LINES: libc::tcflag_t = libc::ECHO | libc::ICANON;

fn attrs(fd: libc::c_int) -> libc::termios {
    let mut t: libc::termios = unsafe { std::mem::zeroed() };
    assert_eq!(unsafe { libc::tcgetattr(fd, &mut t) }, 0, "tcgetattr");
    t
}

fn echo_on(fd: libc::c_int) -> bool {
    attrs(fd).c_lflag & libc::ECHO != 0
}

/// What the terminal still holds as input once the command is gone: what a shell would read next.
/// Read without waiting, with the line mode off so a part of a line counts too.
fn left_in_queue(slave: libc::c_int) -> Vec<u8> {
    let mut t = attrs(slave);
    t.c_lflag &= !(libc::ICANON | libc::ECHO);
    t.c_cc[libc::VMIN] = 0;
    t.c_cc[libc::VTIME] = 0;
    assert_eq!(unsafe { libc::tcsetattr(slave, libc::TCSANOW, &t) }, 0);
    let mut buf = vec![0u8; 65536];
    let n = unsafe { libc::read(slave, buf.as_mut_ptr().cast(), buf.len()) };
    buf.truncate(n.max(0) as usize);
    buf
}

fn wait_for(fd: libc::c_int, write: bool, millis: libc::c_int) -> (bool, bool) {
    let mut p = libc::pollfd {
        fd,
        events: libc::POLLIN | if write { libc::POLLOUT } else { 0 },
        revents: 0,
    };
    let n = unsafe { libc::poll(&mut p, 1, millis) };
    (
        n > 0 && p.revents & (libc::POLLIN | libc::POLLHUP | libc::POLLERR) != 0,
        n > 0 && p.revents & libc::POLLOUT != 0,
    )
}

fn pipe_pair() -> (OwnedFd, OwnedFd) {
    let mut fds = [0 as libc::c_int; 2];
    assert_eq!(unsafe { libc::pipe(fds.as_mut_ptr()) }, 0, "pipe");
    for fd in fds {
        unsafe { libc::fcntl(fd, libc::F_SETFD, libc::FD_CLOEXEC) };
    }
    unsafe { (OwnedFd::from_raw_fd(fds[0]), OwnedFd::from_raw_fd(fds[1])) }
}

/// A copy of `fd` numbered 10 or more, so the descriptors the stand-in shell gets at 3, 4 and 5
/// never overwrite one another while they are set up.
fn high_copy(fd: &OwnedFd) -> OwnedFd {
    let copy = unsafe { libc::fcntl(fd.as_raw_fd(), libc::F_DUPFD_CLOEXEC, 10) };
    assert!(copy >= 10, "fcntl F_DUPFD_CLOEXEC");
    unsafe { OwnedFd::from_raw_fd(copy) }
}

/// The command the stand-in shell started: its pid and exit code as they are reported.
struct Reported {
    status: OwnedFd,
    text: Vec<u8>,
}

impl Reported {
    fn lines(&mut self, wait_millis: libc::c_int) -> Vec<String> {
        if wait_for(self.status.as_raw_fd(), false, wait_millis).0 {
            let mut buf = [0u8; 4096];
            let n = unsafe { libc::read(self.status.as_raw_fd(), buf.as_mut_ptr().cast(), 4096) };
            if n > 0 {
                self.text.extend_from_slice(&buf[..n as usize]);
            }
        }
        let text = String::from_utf8_lossy(&self.text).into_owned();
        let mut lines: Vec<String> = text.split('\n').map(str::to_string).collect();
        lines.pop();
        lines
    }

    fn pid(&mut self) -> libc::pid_t {
        let deadline = Instant::now() + Duration::from_secs(30);
        while Instant::now() < deadline {
            if let Some(first) = self.lines(100).first() {
                return first.parse().unwrap();
            }
        }
        panic!("the command did not start");
    }

    fn poll(&mut self) -> Option<i32> {
        self.lines(0).get(1).map(|l| l.parse().unwrap())
    }

    fn send_signal(&mut self, sig: libc::c_int) {
        // Already gone is fine; the exit code tells.
        unsafe { libc::kill(self.pid(), sig) };
    }
}

/// What a case does once the prompt is on screen: given the terminal's master end and the command.
type AtPrompt<'a> = &'a dyn Fn(libc::c_int, &mut Reported);

/// What a run of the prompt left behind.
struct Run {
    /// The command's stdout.
    out: String,
    /// The terminal bytes.
    term: Vec<u8>,
    /// The command's exit code, or minus the signal that ended it.
    code: Option<i32>,
    /// Echo while the prompt waited; `None` when the prompt never showed.
    echo_asked: Option<bool>,
    /// Echo and line mode on after the exit.
    modes_back: bool,
    /// Input the terminal still holds after the exit.
    left: Vec<u8>,
}

#[derive(Default)]
struct How<'a> {
    args: Option<Vec<&'a str>>,
    /// Pieces typed once the prompt shows: seconds to wait first, then the bytes.
    typed: Vec<(f64, Vec<u8>)>,
    /// Runs once the prompt shows, given the terminal's master end and the command.
    at_prompt: Option<AtPrompt<'a>>,
    stderr_not_a_terminal: bool,
    ignored: Vec<libc::c_int>,
}

fn has(haystack: &[u8], needle: &[u8]) -> bool {
    !needle.is_empty() && haystack.windows(needle.len()).any(|w| w == needle)
}

struct Fx {
    _tmp: TempDir,
    home: PathBuf,
    work: PathBuf,
    /// Test data, made at run time: not a real key and not shaped like one.
    value: String,
}

const ARGS: [&str; 4] = ["key-set", "typesafe", "--shared", "--prompt"];

fn fx() -> Fx {
    let tmp = TempDir::new().unwrap();
    let root = fs::canonicalize(tmp.path()).unwrap();
    let (home, work) = (root.join("home"), root.join("work"));
    fs::create_dir_all(&home).unwrap();
    fs::create_dir_all(&work).unwrap();
    let nanos = std::time::SystemTime::now()
        .duration_since(std::time::UNIX_EPOCH)
        .unwrap()
        .as_nanos();
    Fx {
        _tmp: tmp,
        home,
        work,
        value: format!("moai-prompt-test-{}-{nanos}", std::process::id()),
    }
}

impl Fx {
    fn env(&self) -> Env {
        env_with(
            &[
                ("PATH", "/usr/bin:/bin"),
                ("HOME", self.home.to_str().unwrap()),
                ("LANG", "C"),
                ("TERM", "xterm"),
            ],
            &[],
        )
    }

    fn vault(&self) -> PathBuf {
        self.home.join(".moai-cli").join("vault")
    }

    fn entry(&self) -> PathBuf {
        self.vault().join("_shared").join("typesafe.key")
    }

    fn piped(&self, args: &[&str], stdin: Option<&str>) -> common::pty::Seen {
        piped_with(BIN, &self.work, &self.env(), args, stdin)
    }

    fn status(&self) -> String {
        let s = self.piped(&["key-status", "typesafe", "--shared"], None);
        assert_eq!(s.code, Some(0), "{}", s.stderr);
        json(&s.stdout)["result"]["entry"]
            .as_str()
            .unwrap()
            .to_string()
    }

    /// The command at a hidden prompt on a terminal, under the stand-in shell.
    fn prompt(&self, how: How) -> Run {
        run_prompt(self, how)
    }

    fn assert_hidden(&self, value: &str, r: &Run) {
        assert!(
            !has(&r.term, value.as_bytes()),
            "the typed value reached the terminal"
        );
        assert!(!r.out.contains(value), "the typed value reached stdout");
    }
}

fn assert_terminal_as_before(r: &Run, why: &str) {
    assert!(r.modes_back, "echo and line mode are not back {why}");
    assert_eq!(
        r.left, b"",
        "input was left in the terminal for the shell {why}"
    );
}

fn json(text: &str) -> serde_json::Value {
    serde_json::from_str(text).unwrap_or_else(|e| panic!("not JSON ({e}): {text:?}"))
}

fn run_prompt(f: &Fx, how: How) -> Run {
    let (master, slave) = open_pty();
    let (status_r, status_w) = pipe_pair();
    let (hold_r, hold_w) = pipe_pair();
    let (out_r, out_w) = pipe_pair();
    let (sw, hr, ow) = (high_copy(&status_w), high_copy(&hold_r), high_copy(&out_w));
    let (sw_fd, hr_fd, ow_fd) = (sw.as_raw_fd(), hr.as_raw_fd(), ow.as_raw_fd());

    let args: Vec<&str> = how.args.clone().unwrap_or_else(|| ARGS.to_vec());
    let mut shell = Command::new(std::env::current_exe().unwrap());
    shell
        .args(["--exact", "prompt_shell", "--nocapture"])
        .clean_env()
        .current_dir(&f.work)
        .env("KEY_PROMPT_EXE", BIN)
        .env("KEY_PROMPT_ARGV", args.join("\u{1f}"))
        .env(
            "KEY_PROMPT_IGNORED",
            how.ignored
                .iter()
                .map(|s| s.to_string())
                .collect::<Vec<_>>()
                .join(","),
        )
        .stdin(Stdio::from(slave.try_clone().unwrap()))
        .stdout(Stdio::null())
        .stderr(if how.stderr_not_a_terminal {
            Stdio::null()
        } else {
            Stdio::from(slave.try_clone().unwrap())
        });
    for (k, v) in f.env() {
        shell.env(k, v);
    }
    unsafe {
        shell.pre_exec(move || {
            // The shell leads a new session whose terminal is the one on its stdin.
            libc::setsid();
            libc::ioctl(0, libc::TIOCSCTTY as _, 0);
            libc::dup2(sw_fd, 3);
            libc::dup2(hr_fd, 4);
            libc::dup2(ow_fd, 5);
            Ok(())
        });
    }
    let mut shell = shell.spawn().unwrap();
    drop((status_w, hold_r, out_w, sw, hr, ow));

    let mut command = Reported {
        status: status_r,
        text: Vec::new(),
    };
    let mut chunks: VecDeque<(f64, Vec<u8>)> = how.typed.into();
    let mfd = master.as_raw_fd();
    let mut seen: Vec<u8> = Vec::new();
    let mut asked = false;
    let mut echo_asked = None;
    let mut pending: Vec<u8> = Vec::new();
    let mut due: Option<Instant> = None;
    let deadline = Instant::now() + Duration::from_secs(60);
    while Instant::now() < deadline {
        // The typed bytes go in only as fast as the terminal takes them: a write it has no room
        // for waits its turn instead of blocking this loop.
        let (ready, writable) = wait_for(mfd, !pending.is_empty(), 100);
        if ready {
            let mut buf = vec![0u8; 65536];
            let n = unsafe { libc::read(mfd, buf.as_mut_ptr().cast(), buf.len()) };
            // Linux reports the end of a closed terminal as an error (EIO), macOS as 0.
            if n <= 0 {
                break;
            }
            seen.extend_from_slice(&buf[..n as usize]);
            if !asked && seen.ends_with(PROMPT_END) {
                asked = true;
                echo_asked = Some(echo_on(mfd));
                if let Some(at) = how.at_prompt {
                    at(mfd, &mut command);
                }
                due = Some(Instant::now());
            }
        }
        if !pending.is_empty() && writable {
            let piece = pending.len().min(256);
            let n = unsafe { libc::write(mfd, pending.as_ptr().cast(), piece) };
            if n > 0 {
                pending.drain(..n as usize);
            }
            if pending.is_empty() {
                due = Some(Instant::now());
            }
        }
        if let Some(since) = due.filter(|_| pending.is_empty() && !chunks.is_empty()) {
            // The next piece goes in after its wait, whether or not the command is still there:
            // what a paste sends late still reaches the terminal.
            if Instant::now() >= since + Duration::from_secs_f64(chunks[0].0) {
                pending = chunks.pop_front().unwrap().1;
            }
            continue;
        }
        if !ready && pending.is_empty() && command.poll().is_some() && (chunks.is_empty() || !asked)
        {
            break;
        }
    }
    if command.poll().is_none() {
        // Past the deadline a command still waiting on the terminal fails the test instead of
        // hanging it.
        unsafe { libc::kill(command.pid(), libc::SIGKILL) };
        while command.poll().is_none() {
            std::thread::sleep(Duration::from_millis(50));
        }
    }
    let code = command.poll();
    // The stand-in shell still holds the terminal: this is what the person's shell finds.
    let flags = attrs(slave.as_raw_fd()).c_lflag;
    let modes_back = flags & LFLAG_ECHO_AND_LINES == LFLAG_ECHO_AND_LINES;
    let left = left_in_queue(slave.as_raw_fd());
    // Tell the shell to go: one byte down the hold pipe.
    write_fd(hold_w.as_raw_fd(), b"x");
    // A session leader does not exit while the terminal still has output for a reader: keep
    // reading until the stand-in shell is gone.
    let until = deadline + Duration::from_secs(30);
    while shell.try_wait().unwrap().is_none() && Instant::now() < until {
        if wait_for(mfd, false, 50).0 {
            let mut buf = vec![0u8; 65536];
            let n = unsafe { libc::read(mfd, buf.as_mut_ptr().cast(), buf.len()) };
            if n <= 0 {
                break;
            }
            seen.extend_from_slice(&buf[..n as usize]);
        }
    }
    let _ = shell.wait();
    let mut out = String::new();
    fs::File::from(out_r).read_to_string(&mut out).unwrap();
    Run {
        out,
        term: seen,
        code,
        echo_asked,
        modes_back,
        left,
    }
}

fn typed(bytes: &[u8]) -> Vec<(f64, Vec<u8>)> {
    vec![(0.0, bytes.to_vec())]
}

// ---- the cases ----

#[test]
fn the_typed_value_is_stored_and_never_shown() {
    let f = fx();
    let r = f.prompt(How {
        typed: typed(&[f.value.as_bytes(), b"\n"].concat()),
        ..How::default()
    });
    assert_eq!(r.code, Some(0), "{:?}", String::from_utf8_lossy(&r.term));
    f.assert_hidden(&f.value, &r);
    assert_eq!(
        r.echo_asked,
        Some(false),
        "echo is off while the prompt waits"
    );
    assert_terminal_as_before(&r, "");
    assert!(has(&r.term, b"typesafe"), "the prompt names the provider");
    let result = json(&r.out)["result"].clone();
    // The answer of a piped `key-set`, field for field: no value, length or fingerprint.
    let mut keys: Vec<&str> = result
        .as_object()
        .unwrap()
        .keys()
        .map(String::as_str)
        .collect();
    keys.sort();
    assert_eq!(
        keys,
        [
            "encrypted",
            "folder",
            "path",
            "provider",
            "scope",
            "value_shown",
            "written"
        ]
    );
    assert_eq!(result["value_shown"], false);
    assert_eq!(result["written"], true);
    // The entry decrypts, and no stored file holds the value in clear.
    assert_eq!(f.status(), "present");
    assert!(!has(&fs::read(f.entry()).unwrap(), f.value.as_bytes()));
}

#[test]
fn the_prompt_answers_like_the_pipe() {
    let f = fx();
    let piped = f.piped(
        &["key-set", "typesafe", "--shared"],
        Some(&format!("{}\n", f.value)),
    );
    assert_eq!(piped.code, Some(0), "{}", piped.stderr);
    // A terminal that ends a line with CR LF sends one more line end: still one line.
    let r = f.prompt(How {
        typed: typed(&[f.value.as_bytes(), b"\r\n"].concat()),
        ..How::default()
    });
    assert_eq!(r.code, Some(0), "{:?}", String::from_utf8_lossy(&r.term));
    assert_eq!(r.out, piped.stdout, "the stdout envelope is the same bytes");
    assert_terminal_as_before(&r, "");
}

#[test]
fn without_a_terminal_the_prompt_is_refused_and_nothing_is_stored() {
    let f = fx();
    let with_value = format!("{}\n", f.value);
    for stdin in [Some(with_value.as_str()), None] {
        let s = f.piped(&ARGS, stdin);
        assert_eq!(s.code, Some(1), "{}", s.stderr);
        assert!(!format!("{}{}", s.stdout, s.stderr).contains(&f.value));
        assert_eq!(
            json(&s.stdout)["error"]["message"],
            "--prompt reads the value from a terminal with echo off, and stdin is not a terminal; nothing was read or stored (run it in your own terminal window, or drop --prompt and pipe the value in)"
        );
        assert!(!f.vault().exists());
    }
}

#[test]
fn without_a_terminal_to_ask_on_the_prompt_is_refused() {
    // stdin is a terminal, stderr is not: the question could not be seen.
    let f = fx();
    let r = f.prompt(How {
        typed: typed(&[f.value.as_bytes(), b"\n"].concat()),
        stderr_not_a_terminal: true,
        ..How::default()
    });
    assert_eq!(r.code, Some(1), "{}", r.out);
    assert_eq!(
        json(&r.out)["error"]["message"],
        "--prompt shows its question on stderr, and stderr is not a terminal; nothing was read or stored (run it in your own terminal window, or drop --prompt and pipe the value in)"
    );
    assert_eq!(r.echo_asked, None);
    assert!(!f.vault().exists());
}

#[test]
fn ctrl_c_leaves_the_terminal_as_it_was_and_stores_nothing() {
    let f = fx();
    let head = f.value.as_bytes()[..8].to_vec();
    let interrupt = |master: libc::c_int, _: &mut Reported| {
        let bytes = [head.as_slice(), b"\x03"].concat();
        write_fd(master, &bytes);
    };
    let r = f.prompt(How {
        at_prompt: Some(&interrupt),
        ..How::default()
    });
    assert_eq!(
        r.code,
        Some(-libc::SIGINT),
        "{:?}",
        String::from_utf8_lossy(&r.term)
    );
    assert_eq!(r.echo_asked, Some(false));
    assert_terminal_as_before(&r, "after Ctrl-C");
    assert!(!has(&r.term, &head));
    assert!(!f.vault().exists());
}

/// A signal sent at the prompt, with what was typed first, must end the command with `code`, the
/// terminal as it was, and nothing stored.
fn ends_with_signal(f: &Fx, sig: libc::c_int, code: i32) {
    let head = f.value.as_bytes()[..8].to_vec();
    let end = |master: libc::c_int, proc: &mut Reported| {
        write_fd(master, &head);
        proc.send_signal(sig);
    };
    let r = f.prompt(How {
        at_prompt: Some(&end),
        ..How::default()
    });
    assert_eq!(
        r.code,
        Some(code),
        "{sig}: {:?}",
        String::from_utf8_lossy(&r.term)
    );
    assert_eq!(r.echo_asked, Some(false));
    assert_terminal_as_before(&r, &format!("after signal {sig}"));
    assert!(!has(&r.term, &head));
    assert!(!f.vault().exists());
}

#[test]
fn ending_signals_leave_the_terminal_as_it_was_and_store_nothing() {
    // SIGQUIT ends the process without a core file (it could hold typed bytes): the exit status
    // is the conventional 128 + signal instead of the signal itself.
    let f = fx();
    for (sig, code) in [
        (libc::SIGTERM, -libc::SIGTERM),
        (libc::SIGHUP, -libc::SIGHUP),
        (libc::SIGQUIT, 128 + libc::SIGQUIT),
    ] {
        ends_with_signal(&f, sig, code);
    }
}

#[test]
fn the_other_ending_signals_leave_the_terminal_as_it_was_too() {
    // Every signal that ends a process by default, not only the four a terminal sends: a timer, a
    // user signal and a resource limit here. SIGXCPU could write a core file, so like SIGQUIT it
    // ends with 128 + signal.
    let f = fx();
    for (sig, code) in [
        (libc::SIGUSR1, -libc::SIGUSR1),
        (libc::SIGALRM, -libc::SIGALRM),
        (libc::SIGXCPU, 128 + libc::SIGXCPU),
    ] {
        ends_with_signal(&f, sig, code);
    }
}

#[test]
fn a_signal_drops_what_was_typed_and_not_read_yet() {
    // The command is stopped (SIGSTOP cannot be caught), so what is typed now stays in the
    // terminal unread; then it is told to end and let go. The handler runs before the command
    // reads again, and it must drop that input with the modes it puts back: otherwise the shell
    // would run it as its next command. The order matters: the signal is already waiting when
    // the input arrives, so the interrupted read hands over to the handler instead of taking the
    // input first (with the input sent first the read took it, and a handler that only put the
    // modes back passed).
    let f = fx();
    let end = |master: libc::c_int, proc: &mut Reported| {
        proc.send_signal(libc::SIGSTOP);
        std::thread::sleep(Duration::from_millis(300));
        proc.send_signal(libc::SIGTERM);
        std::thread::sleep(Duration::from_millis(100));
        write_fd(master, b"echo LEFT-FOR-THE-SHELL\n");
        std::thread::sleep(Duration::from_millis(100));
        proc.send_signal(libc::SIGCONT);
    };
    let r = f.prompt(How {
        at_prompt: Some(&end),
        ..How::default()
    });
    assert_eq!(
        r.code,
        Some(-libc::SIGTERM),
        "{:?}",
        String::from_utf8_lossy(&r.term)
    );
    assert_terminal_as_before(&r, "after a signal with input waiting");
    assert!(!has(&r.term, b"LEFT-FOR-THE-SHELL"));
    assert!(!f.vault().exists());
}

#[test]
fn an_ignored_hangup_stays_ignored_and_a_stop_signal_does_not_stop_the_prompt() {
    // Under `nohup` SIGHUP is ignored and stays so; SIGTSTP is ignored while echo is off, since a
    // stopped process could not put the terminal's modes back.
    let f = fx();
    let poke = |_: libc::c_int, proc: &mut Reported| {
        proc.send_signal(libc::SIGHUP);
        proc.send_signal(libc::SIGTSTP);
    };
    let r = f.prompt(How {
        at_prompt: Some(&poke),
        typed: vec![(0.3, [f.value.as_bytes(), b"\n"].concat())],
        ignored: vec![libc::SIGHUP],
        ..How::default()
    });
    assert_eq!(r.code, Some(0), "{:?}", String::from_utf8_lossy(&r.term));
    assert_terminal_as_before(&r, "");
    assert_eq!(f.status(), "present");
}

#[test]
fn empty_and_control_character_input_is_refused_with_the_terminal_as_it_was() {
    // Enter alone, spaces, a control character inside the value, and an arrow key (an escape
    // sequence).
    let f = fx();
    for input in [b"\n".as_slice(), b"   \n", b"ab\x01cd\n", b"ab\x1b[Dcd\n"] {
        let r = f.prompt(How {
            typed: typed(input),
            ..How::default()
        });
        let why = format!("{input:?}");
        assert_eq!(
            r.code,
            Some(1),
            "{why}: {:?}",
            String::from_utf8_lossy(&r.term)
        );
        assert_eq!(json(&r.out)["ok"], false, "{why}");
        assert_terminal_as_before(&r, &why);
        assert!(!has(&r.term, b"abcd"), "{why}");
        assert!(!f.vault().exists(), "{why}");
    }
}

#[test]
fn ctrl_d_cancels_and_stores_nothing() {
    // On an empty line and after a part of a value alike: Ctrl-D never ends a value.
    let f = fx();
    for input in [b"\x04".to_vec(), [f.value.as_bytes(), b"\x04"].concat()] {
        let r = f.prompt(How {
            typed: typed(&input),
            ..How::default()
        });
        let why = format!("{input:?}");
        assert_eq!(
            r.code,
            Some(1),
            "{why}: {:?}",
            String::from_utf8_lossy(&r.term)
        );
        assert!(
            json(&r.out)["error"]["message"]
                .as_str()
                .unwrap()
                .contains("cancelled"),
            "{}",
            r.out
        );
        f.assert_hidden(&f.value, &r);
        assert_terminal_as_before(&r, &why);
        assert!(!f.vault().exists(), "{why}");
    }
}

#[test]
fn a_second_line_refuses_the_whole_input() {
    // A paste of two lines, in one piece and with the second line a moment later: nothing is
    // stored, neither line is shown, and nothing is left for the shell to run.
    let f = fx();
    for input in [
        typed(b"AAAfirst\nBBBsecond\n"),
        vec![
            (0.0, b"AAAfirst\n".to_vec()),
            (0.3, b"BBBsecond\n".to_vec()),
        ],
    ] {
        let r = f.prompt(How {
            typed: input.clone(),
            ..How::default()
        });
        let why = format!("{input:?}");
        assert_eq!(
            r.code,
            Some(1),
            "{why}: {:?}",
            String::from_utf8_lossy(&r.term)
        );
        assert!(
            json(&r.out)["error"]["message"]
                .as_str()
                .unwrap()
                .contains("more than one line"),
            "{}",
            r.out
        );
        assert!(!has(&r.term, b"AAAfirst"), "{why}");
        assert!(!has(&r.term, b"BBBsecond"), "{why}");
        assert!(!r.out.contains("AAAfirst") && !r.out.contains("BBBsecond"));
        assert_terminal_as_before(&r, &why);
        assert!(!f.vault().exists(), "{why}");
    }
}

#[test]
fn backspace_edits_the_hidden_line() {
    // A wrong character and its backspace: stored as if never typed. A stray control byte would
    // be refused (the case above), so a stored entry shows the edit was read.
    let f = fx();
    let r = f.prompt(How {
        typed: typed(&[b"x\x7f".as_slice(), f.value.as_bytes(), b"\n"].concat()),
        ..How::default()
    });
    assert_eq!(r.code, Some(0), "{:?}", String::from_utf8_lossy(&r.term));
    f.assert_hidden(&f.value, &r);
    assert_eq!(f.status(), "present");
}

#[test]
fn the_size_limit_holds_at_the_prompt() {
    // 65,535 bytes is the largest value stored (a 65,583-byte entry); 65,536 is refused.
    let f = fx();
    let longest = vec![b'k'; 65_535];
    let r = f.prompt(How {
        typed: typed(&[longest.as_slice(), b"\n"].concat()),
        ..How::default()
    });
    let tail =
        |r: &Run| String::from_utf8_lossy(&r.term[r.term.len().saturating_sub(200)..]).into_owned();
    assert_eq!(r.code, Some(0), "{}", tail(&r));
    assert!(!has(&r.term, b"kkkk"));
    assert!(!r.out.replace(',', "").contains("65535"), "{}", r.out);
    assert_eq!(fs::metadata(f.entry()).unwrap().len(), 65_583);
    assert_eq!(f.status(), "present");
    fs::remove_file(f.entry()).unwrap();
    let r = f.prompt(How {
        typed: typed(&[longest.as_slice(), b"k\n"].concat()),
        ..How::default()
    });
    assert_eq!(r.code, Some(1), "{}", tail(&r));
    assert!(!has(&r.term, b"kkkk"));
    assert_terminal_as_before(&r, "");
    assert!(!f.entry().exists());
}

#[test]
fn a_bad_provider_name_is_refused_before_anything_is_asked() {
    let f = fx();
    let r = f.prompt(How {
        args: Some(vec!["key-set", "Not_A_Name", "--shared", "--prompt"]),
        typed: typed(&[f.value.as_bytes(), b"\n"].concat()),
        ..How::default()
    });
    assert_eq!(r.code, Some(1), "{:?}", String::from_utf8_lossy(&r.term));
    assert_eq!(
        r.echo_asked, None,
        "no prompt for a name that cannot be stored"
    );
    assert!(r.modes_back);
    assert!(!f.vault().exists());
}
