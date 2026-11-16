//! The hidden prompt of `key-set --prompt` (card C-54 S1, user decision 2026-11-16 "Key
//! entry": a hidden prompt command): one line typed at the terminal with echo off.
//!
//! Unix only, through `termios` of the `libc` dependency (no terminal crate). The approach is
//! the one `getpass(3)`, Python's `getpass` and the `rpassword` crate share (save the
//! terminal's modes, turn `ECHO` off, read one line, put the modes back), with these
//! differences, all on purpose:
//! - canonical mode is off as well, so the line is put together here ([`read_line`]). The
//!   kernel's canonical line holds about 1 KiB (macOS) or 4 KiB (Linux) and drops what does
//!   not fit without saying so; read byte by byte, the terminal waits instead, and the
//!   vault's own size limit is the one that answers.
//! - one line is one value. After the line end the prompt keeps the terminal quiet until one
//!   second passes without input ([`QUIET_TENTHS`]) and counts what still arrives: anything
//!   but a single further line end (a terminal that sends CR LF) refuses the whole input, as
//!   the stdin path refuses a value with a line break in it. A paste of two lines is thus
//!   neither stored in part, nor shown, nor left for the shell to run. The second is a bound,
//!   not a proof: a second line that starts more than a second after the first ended is
//!   caught only if it is already in the terminal when the modes go back (the unread count,
//!   `FIONREAD`, is asked right before; anything there refuses the value). Later than that
//!   it is the shell's, like anything typed after any command ended.
//! - the modes are put back with `TCSAFLUSH`, which also drops input not read yet, at every
//!   exit and from a signal handler, which then lets the signal end the process as it would
//!   have. The handler is there for every signal that ends a process by default and is not a
//!   program fault: `SIGINT`, `SIGTERM`, `SIGHUP`, `SIGQUIT`, `SIGUSR1`, `SIGUSR2`,
//!   `SIGALRM`, `SIGPIPE`, `SIGXCPU`, `SIGXFSZ`, `SIGVTALRM`, `SIGPROF`. `SIGQUIT`, `SIGXCPU`
//!   and `SIGXFSZ` end it with the usual status
//!   (128 + signal) but without a core file, which could hold typed bytes. A signal that was
//!   ignored when the prompt started (`nohup`) stays ignored. `SIGTSTP` is ignored while echo
//!   is off (and Ctrl-Z is switched off), since a stopped process cannot put anything back.
//! - Ctrl-D never ends a value: it cancels, on an empty line and after a part of one.
//!
//! Nothing typed is ever written anywhere: not the value, its length or a fingerprint.

type Result<T> = std::result::Result<T, String>;

/// Erase keys: DEL (what Backspace sends on most terminals) and BS.
const ERASE: [u8; 2] = [0x7f, 0x08];
/// Ctrl-U: drop the whole line.
const KILL: u8 = 0x15;
/// Ctrl-D: cancel.
const CANCEL: u8 = 0x04;
/// How long the terminal must stay silent after the line end before the prompt lets go of
/// it, in tenths of a second (the unit of `VTIME`).
pub const QUIET_TENTHS: u8 = 10;

/// One line from `input`, as a terminal with canonical mode off delivers it: it ends at a
/// line feed or a carriage return, and nothing after that is taken from `input`'s source
/// beyond what it had already buffered. Backspace removes the last character and Ctrl-U all
/// of them. Ctrl-D, or the end of `input` before a line end, cancels: an error, whatever was
/// typed before it. A line of more than [`crate::key_vault::VALUE_FILE_MAX`] bytes is
/// refused, never cut: its rest is still read up to the line end, and an erase cannot bring
/// it back under the limit (what was dropped would be missing from the value).
pub fn read_line(input: &mut impl std::io::BufRead) -> Result<Vec<u8>> {
    use std::io::Read;
    let keep = crate::key_vault::VALUE_FILE_MAX as usize;
    let mut line: Vec<u8> = Vec::new();
    let mut dropped = false;
    let mut ended = false;
    for byte in input.by_ref().bytes() {
        let byte = byte.map_err(|_| "could not read the terminal; nothing was stored")?;
        match byte {
            b'\n' | b'\r' => {
                ended = true;
                break;
            }
            CANCEL => return Err("cancelled with Ctrl-D; nothing was stored".into()),
            b if ERASE.contains(&b) => {
                // One character, not one byte: drop the continuation bytes of a UTF-8
                // sequence, then its first byte.
                while line.last().is_some_and(|l| l & 0xc0 == 0x80) {
                    line.pop();
                }
                line.pop();
            }
            KILL => {
                line.clear();
                dropped = false;
            }
            b if line.len() < keep => line.push(b),
            _ => dropped = true,
        }
    }
    if !ended {
        return Err("cancelled: the input ended before Enter; nothing was stored".into());
    }
    if dropped {
        return Err("the value exceeds 64 KiB".into());
    }
    Ok(line)
}

/// What arrived after the line end: how much, and its first byte.
#[derive(Debug, Default, Clone, Copy, PartialEq, Eq)]
pub struct Extra {
    count: usize,
    first: Option<u8>,
    /// Bytes still in the terminal, unread, when it was given back.
    queued: usize,
}

impl Extra {
    pub fn note(&mut self, bytes: &[u8]) {
        if self.first.is_none() {
            self.first = bytes.first().copied();
        }
        self.count += bytes.len();
    }

    /// Notes `queued` bytes the terminal still held, unread, when the wait for silence was
    /// over: they came after it. Their content is unknown, so one byte is already too much.
    pub fn note_queued(&mut self, queued: usize) {
        self.queued += queued;
    }

    /// True when it is more than the one further line end a CR LF terminal sends, or when
    /// anything was still queued at the end.
    pub fn is_more_input(&self) -> bool {
        self.queued > 0
            || !(self.count == 0 || (self.count == 1 && matches!(self.first, Some(b'\n' | b'\r'))))
    }
}

/// The refusal of input that went on after the line end.
pub const MORE_THAN_ONE_LINE: &str = "more than one line was typed or pasted; nothing was \
     stored (the prompt takes one value on one line)";

/// Asks for the `provider` key on the terminal stdin is, with echo off, and returns the one
/// usable value typed (the rule and the limits of [`crate::key_vault::read_value`]). The
/// terminal's modes are put back before this returns, and when a signal ends the process.
/// The caller checks first that stdin and stderr are terminals.
#[cfg(unix)]
pub fn read_hidden(provider: &str) -> Result<String> {
    use std::io::{BufRead, Write};
    let lang = crate::ui::ui_lang(crate::ui::Stream::Stderr);
    let question = crate::ui_text::fill(
        crate::ui_text::Msg::KeySetPrompt,
        lang,
        &[("provider", provider)],
    );
    let (line, extra) = {
        // Echo goes off before the question shows, so nothing typed after it is shown.
        let quiet = unix::Quiet::start()?;
        let mut err = std::io::stderr();
        let _ = write!(err, "{question}: ");
        let _ = err.flush();
        let mut input = std::io::BufReader::new(unix::Terminal);
        let line = read_line(&mut input);
        // What one read brought along with the line, then what still comes in until the
        // terminal has been silent for a moment. Only then do the modes go back.
        let mut extra = Extra::default();
        extra.note(input.buffer());
        let buffered = input.buffer().len();
        input.consume(buffered);
        quiet.drain(&mut extra);
        // The Enter that ended the line was not echoed either.
        let _ = writeln!(err);
        // Last thing before the modes go back (the drop of `quiet`, which also discards
        // what is queued): input that came in after the silence refuses the value as well.
        extra.note_queued(quiet.queued());
        (line, extra)
    };
    let line = line?;
    if extra.is_more_input() {
        return Err(MORE_THAN_ONE_LINE.into());
    }
    crate::key_vault::read_value(line.as_slice())
}

/// Without `termios` there is no way here to turn echo off, so nothing is asked.
#[cfg(not(unix))]
pub fn read_hidden(_provider: &str) -> Result<String> {
    Err(UNSUPPORTED.into())
}

/// The refusal of `--prompt` where it cannot hide what is typed.
pub const UNSUPPORTED: &str = "--prompt needs a Unix terminal (it turns echo off through \
     termios), and this system has none; nothing was read or stored (pipe the value in instead)";

#[cfg(unix)]
mod unix {
    use std::mem::MaybeUninit;
    use std::ptr::{addr_of, addr_of_mut};
    use std::sync::atomic::{AtomicBool, Ordering};

    const FD: libc::c_int = libc::STDIN_FILENO;
    /// The signals that end a process by default, that can be caught and that nothing else
    /// here handles: Ctrl-C, `kill`, a closed terminal window, Ctrl-\, then the user signals,
    /// the timers, a closed pipe and the resource limits (second audit, N4). Not in it: the
    /// signals of a fault in the program itself (`SIGSEGV`, `SIGBUS`, `SIGABRT` and the
    /// like) and the ones only some systems have (`SIGIO`, `SIGPWR`, `SIGSTKFLT` on Linux).
    const ENDING: [libc::c_int; 12] = [
        libc::SIGINT,
        libc::SIGTERM,
        libc::SIGHUP,
        libc::SIGQUIT,
        libc::SIGUSR1,
        libc::SIGUSR2,
        libc::SIGALRM,
        libc::SIGPIPE,
        libc::SIGXCPU,
        libc::SIGXFSZ,
        libc::SIGVTALRM,
        libc::SIGPROF,
    ];
    /// The ones of [`ENDING`] whose default action may also write a core file.
    const WITH_CORE: [libc::c_int; 3] = [libc::SIGQUIT, libc::SIGXCPU, libc::SIGXFSZ];
    /// Most bytes read and dropped after the line end before the prompt stops waiting for
    /// silence; what is still queued then goes with `TCSAFLUSH`.
    const DRAIN_MAX: usize = 1 << 20;

    /// The terminal's modes as they were, for the signal handler. Written once by
    /// [`Quiet::start`] before `ARMED` is set and before the handler is installed; read by the
    /// handler only while `ARMED`.
    static mut SAVED: MaybeUninit<libc::termios> = MaybeUninit::uninit();
    static ARMED: AtomicBool = AtomicBool::new(false);

    /// Puts the saved modes back and drops the input not read yet (`TCSAFLUSH`), then lets
    /// the signal end the process the default way (the exit status a shell sees is the
    /// signal's, as without the prompt). A signal of [`WITH_CORE`] would also write a core
    /// file, which could hold what was typed: for those the process exits with the status a
    /// shell reports for the signal, 128 + its number. Only async-signal-safe calls:
    /// `tcsetattr`, `_exit`, `signal`, `raise`.
    extern "C" fn restore_then_end(signal: libc::c_int) {
        if ARMED.load(Ordering::SeqCst) {
            // SAFETY: `SAVED` was initialized before `ARMED` was set and is not written
            // while armed; tcsetattr only reads it.
            unsafe { libc::tcsetattr(FD, libc::TCSAFLUSH, addr_of!(SAVED).cast()) };
        }
        // SAFETY: plain libc calls. The signal is blocked while its handler runs, so the
        // raised one is delivered, with the default action, when the handler returns.
        unsafe {
            // A plain loop over a constant: nothing here allocates or locks.
            let mut at = 0;
            while at < WITH_CORE.len() {
                if signal == WITH_CORE[at] {
                    libc::_exit(128 + signal);
                }
                at += 1;
            }
            libc::signal(signal, libc::SIG_DFL);
            libc::raise(signal);
        }
    }

    /// Standard input, read without a buffer of its own and through an interrupted call.
    pub struct Terminal;

    impl std::io::Read for Terminal {
        fn read(&mut self, buf: &mut [u8]) -> std::io::Result<usize> {
            loop {
                // SAFETY: read fills at most `buf.len()` bytes of the buffer it is given.
                let n = unsafe { libc::read(FD, buf.as_mut_ptr().cast(), buf.len()) };
                if n >= 0 {
                    return Ok(n as usize);
                }
                let error = std::io::Error::last_os_error();
                if error.kind() != std::io::ErrorKind::Interrupted {
                    return Err(error);
                }
            }
        }
    }

    /// The terminal with echo off, until dropped.
    pub struct Quiet {
        saved: libc::termios,
        hidden: libc::termios,
        /// The action each [`ENDING`] signal had, where the handler replaced it.
        actions: [Option<libc::sigaction>; ENDING.len()],
        /// What `SIGTSTP` did before it was ignored.
        stop: libc::sighandler_t,
    }

    impl Quiet {
        pub fn start() -> super::Result<Self> {
            let failed = "could not turn the terminal's echo off; nothing was read or stored";
            // SAFETY: tcgetattr fills the struct it is given for an open descriptor.
            let mut saved: libc::termios = unsafe { std::mem::zeroed() };
            if unsafe { libc::tcgetattr(FD, &mut saved) } != 0 {
                return Err(failed.into());
            }
            // SAFETY: one prompt per process and no handler installed yet, so nothing reads
            // `SAVED` during this write.
            unsafe { addr_of_mut!(SAVED).write(MaybeUninit::new(saved)) };
            ARMED.store(true, Ordering::SeqCst);
            let mut actions: [Option<libc::sigaction>; ENDING.len()] = [None; ENDING.len()];
            for (signal, kept) in ENDING.iter().zip(actions.iter_mut()) {
                // SAFETY: a zeroed sigaction is a valid value; sigaction fills the old
                // action; the handler is an `extern "C" fn(c_int)` that stays valid for the
                // life of the process.
                unsafe {
                    let mut old: libc::sigaction = std::mem::zeroed();
                    libc::sigaction(*signal, std::ptr::null(), &mut old);
                    // Ignored on purpose by whoever started this process (`nohup`): the
                    // signal does not end it, so there is nothing to put back for it.
                    if old.sa_sigaction == libc::SIG_IGN {
                        continue;
                    }
                    let mut new: libc::sigaction = std::mem::zeroed();
                    new.sa_sigaction = restore_then_end as extern "C" fn(libc::c_int) as usize;
                    libc::sigemptyset(&mut new.sa_mask);
                    libc::sigaction(*signal, &new, std::ptr::null_mut());
                    *kept = Some(old);
                }
            }
            // A stopped process cannot put the modes back: no stop while echo is off.
            // SAFETY: signal only changes the disposition and returns the old one.
            let stop = unsafe { libc::signal(libc::SIGTSTP, libc::SIG_IGN) };
            let mut hidden = saved;
            hidden.c_lflag &= !(libc::ECHO | libc::ECHONL | libc::ICANON | libc::IEXTEN);
            hidden.c_cc[libc::VMIN] = 1;
            hidden.c_cc[libc::VTIME] = 0;
            hidden.c_cc[libc::VSUSP] = libc::_POSIX_VDISABLE;
            // From here on a drop puts everything back, a failed tcsetattr included.
            let quiet = Quiet {
                saved,
                hidden,
                actions,
                stop,
            };
            // TCSAFLUSH: what was typed before the question is dropped, not taken as the key.
            // SAFETY: tcsetattr reads the struct it is given.
            if unsafe { libc::tcsetattr(FD, libc::TCSAFLUSH, &quiet.hidden) } != 0 {
                return Err(failed.into());
            }
            Ok(quiet)
        }

        /// How many bytes the terminal holds that nobody has read (`FIONREAD`); 0 when it
        /// cannot be asked.
        pub fn queued(&self) -> usize {
            let mut count: libc::c_int = 0;
            // SAFETY: FIONREAD writes one `int` into the pointer it is given.
            if unsafe { libc::ioctl(FD, libc::FIONREAD, &mut count) } != 0 {
                return 0;
            }
            usize::try_from(count).unwrap_or(0)
        }

        /// Reads and drops what still arrives, echo off, until the terminal has been silent
        /// for [`super::QUIET_TENTHS`]; `extra` counts it.
        pub fn drain(&self, extra: &mut super::Extra) {
            let mut waiting = self.hidden;
            waiting.c_cc[libc::VMIN] = 0;
            waiting.c_cc[libc::VTIME] = super::QUIET_TENTHS;
            // SAFETY: tcsetattr reads the struct it is given.
            if unsafe { libc::tcsetattr(FD, libc::TCSANOW, &waiting) } != 0 {
                return;
            }
            let mut buf = [0u8; 4096];
            let mut total = 0;
            // A read of 0 bytes is the silence; an error ends the wait as well.
            while let Ok(n @ 1..) = std::io::Read::read(&mut Terminal, &mut buf) {
                extra.note(&buf[..n]);
                total += n;
                if total > DRAIN_MAX {
                    break;
                }
            }
        }
    }

    impl Drop for Quiet {
        fn drop(&mut self) {
            // SAFETY: tcsetattr, sigaction and signal read what `start` saved.
            unsafe {
                libc::tcsetattr(FD, libc::TCSAFLUSH, &self.saved);
                ARMED.store(false, Ordering::SeqCst);
                for (signal, old) in ENDING.iter().zip(self.actions.iter()) {
                    if let Some(old) = old {
                        libc::sigaction(*signal, old, std::ptr::null_mut());
                    }
                }
                libc::signal(libc::SIGTSTP, self.stop);
            }
        }
    }
}

#[cfg(test)]
mod tests {
    use super::*;
    use crate::key_vault::{read_value, VALUE_FILE_MAX, VALUE_MAX};

    fn line(mut typed: &[u8]) -> Vec<u8> {
        read_line(&mut typed).unwrap()
    }

    fn refused(mut typed: &[u8]) -> String {
        read_line(&mut typed).unwrap_err()
    }

    #[test]
    fn a_line_ends_at_enter_and_only_there() {
        assert_eq!(line(b"abc\nrest"), b"abc");
        assert_eq!(line(b"abc\rrest"), b"abc");
        assert_eq!(line(b"\n"), b"");
        // What follows the line end stays with the reader, for the caller to count.
        let mut typed: &[u8] = b"abc\nrest\n";
        assert_eq!(read_line(&mut typed).unwrap(), b"abc");
        assert_eq!(typed, b"rest\n");
    }

    /// Ctrl-D and an input that ends without Enter cancel, with or without a part of a value.
    #[test]
    fn ctrl_d_and_the_end_of_input_cancel() {
        for typed in [&b"\x04"[..], b"abc\x04", b"abc\x04\n", b"abc\x04def\n"] {
            assert_eq!(
                refused(typed),
                "cancelled with Ctrl-D; nothing was stored",
                "{typed:?}"
            );
        }
        for typed in [&b""[..], b"abc"] {
            assert_eq!(
                refused(typed),
                "cancelled: the input ended before Enter; nothing was stored",
                "{typed:?}"
            );
        }
    }

    #[test]
    fn backspace_removes_one_character_and_ctrl_u_the_line() {
        assert_eq!(line(b"abx\x7fc\n"), b"abc");
        assert_eq!(line(b"abx\x08c\n"), b"abc");
        // One character of two and of three bytes, and more erases than characters.
        assert_eq!(line("a\u{e9}\x7fb\n".as_bytes()), b"ab");
        assert_eq!(line("a\u{d0a4}\x7fb\n".as_bytes()), b"ab");
        assert_eq!(line(b"a\x7f\x7f\x7fb\n"), b"b");
        assert_eq!(line(b"wrong\x15right\n"), b"right");
    }

    /// Any other control byte stays in the line, so the value rule refuses the whole input
    /// instead of storing something other than what was typed.
    #[test]
    fn other_control_bytes_reach_the_value_rule_and_are_refused() {
        for typed in [
            &b"ab\x01cd\n"[..],
            b"ab\x1b[Dcd\n",
            b"ab\tcd\n",
            b"\n",
            b"  \n",
        ] {
            assert!(read_value(line(typed).as_slice()).is_err(), "{typed:?}");
        }
        assert_eq!(read_value(line(b" ok \n").as_slice()).unwrap(), "ok");
    }

    /// After the line end: nothing, or the one further line end of a CR LF terminal. Anything
    /// else is a second line.
    #[test]
    fn only_one_further_line_end_may_follow_the_line() {
        let after = |pieces: &[&[u8]]| {
            let mut extra = Extra::default();
            for piece in pieces {
                extra.note(piece);
            }
            extra.is_more_input()
        };
        assert!(!after(&[]));
        assert!(!after(&[b""]));
        assert!(!after(&[b"\n"]));
        assert!(!after(&[b"\r"]));
        assert!(!after(&[b"", b"\n"]));
        assert!(after(&[b"\n\n"]));
        assert!(after(&[b"\n", b"\n"]));
        assert!(after(&[b"B"]));
        assert!(after(&[b"BBBsecond\n"]));
        assert!(after(&[b"\n", b"BBBsecond\n"]));
        assert!(after(&[b" "]));
    }

    /// Second audit, N2: input still queued when the terminal is given back came after the
    /// silence. What it is was never read, so any of it refuses, a lone line end too.
    #[test]
    fn input_queued_at_the_end_refuses_whatever_it_is() {
        let late = |pieces: &[&[u8]], queued: usize| {
            let mut extra = Extra::default();
            for piece in pieces {
                extra.note(piece);
            }
            extra.note_queued(queued);
            extra.is_more_input()
        };
        assert!(!late(&[], 0));
        assert!(!late(&[b"\n"], 0));
        assert!(late(&[], 1));
        assert!(late(&[b"\n"], 1));
        assert!(late(&[], 4096));
        assert!(late(&[b"B"], 0));
    }

    /// The vault's limit at its edge, through the line reader: 65,535 bytes are a value and
    /// 65,536 are refused. Literal sizes.
    #[test]
    fn the_size_limit_is_the_vaults() {
        assert_eq!((VALUE_MAX, VALUE_FILE_MAX), (65_535, 65_536));
        let n = "k".repeat(65_535);
        let typed = format!("{n}\n");
        assert_eq!(read_value(line(typed.as_bytes()).as_slice()).unwrap(), n);
        let over = format!("{n}k\n");
        assert_eq!(
            read_value(line(over.as_bytes()).as_slice()).unwrap_err(),
            "the value exceeds 65,535 bytes"
        );
        // Far over: refused, not cut, and an erase after the limit was passed cannot bring
        // the line back under it. A line dropped whole (Ctrl-U) starts again.
        let far = format!("{}\n", "k".repeat(200_000));
        assert_eq!(refused(far.as_bytes()), "the value exceeds 64 KiB");
        let cut = format!("{}\x7f\x7f\x7f\n", "k".repeat(65_537));
        assert_eq!(refused(cut.as_bytes()), "the value exceeds 64 KiB");
        let again = format!("{}\x15ok\n", "k".repeat(65_537));
        assert_eq!(line(again.as_bytes()), b"ok");
        // The rest of an over-long line is read up to its end, not left behind.
        let mut typed: &[u8] = far.as_bytes();
        assert!(read_line(&mut typed).is_err());
        assert_eq!(typed, b"");
    }
}
