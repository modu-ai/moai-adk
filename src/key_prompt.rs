//! The hidden prompt of `key-set --prompt` (card C-54 S1, user decision 2026-11-16 "Key
//! entry": a hidden prompt command): one line typed at the terminal with echo off.
//!
//! Unix only, through `termios` of the `libc` dependency (no terminal crate). The approach is
//! the one `getpass(3)`, Python's `getpass` and the `rpassword` crate share (save the
//! terminal's modes, turn `ECHO` off, read one line, put the modes back), with two
//! differences, both on purpose:
//! - canonical mode is off as well, so the line is put together here ([`read_line`]). The
//!   kernel's canonical line holds about 1 KiB (macOS) or 4 KiB (Linux) and drops what does
//!   not fit without saying so; read byte by byte, the terminal waits instead, and the
//!   vault's own size limit is the one that answers.
//! - the saved modes are also put back from a signal handler (`SIGINT`, `SIGTERM`, `SIGHUP`,
//!   `SIGQUIT`), which then lets the signal end the process as it would have. A Ctrl-C at the
//!   prompt therefore never leaves the shell without echo. Suspending (Ctrl-Z) is switched
//!   off while the prompt waits, since a stopped process cannot put anything back.
//!
//! Nothing typed is ever written anywhere: not the value, its length or a fingerprint.

type Result<T> = std::result::Result<T, String>;

/// Erase keys: DEL (what Backspace sends on most terminals) and BS.
const ERASE: [u8; 2] = [0x7f, 0x08];
/// Ctrl-U: drop the whole line.
const KILL: u8 = 0x15;
/// Ctrl-D: end of input.
const END_OF_INPUT: u8 = 0x04;

/// One line from `input`, as a terminal with canonical mode off delivers it: ends at a line
/// feed, a carriage return, Ctrl-D or the end of `input`; Backspace removes the last
/// character and Ctrl-U all of them. A line of more than
/// [`crate::key_vault::VALUE_FILE_MAX`] bytes is refused, never cut: its rest is still read
/// and dropped, so it is not left for the shell to run, and an erase cannot bring it back
/// under the limit (what was dropped would be missing from the value).
pub fn read_line(input: impl std::io::BufRead) -> Result<Vec<u8>> {
    let keep = crate::key_vault::VALUE_FILE_MAX as usize;
    let mut line: Vec<u8> = Vec::new();
    let mut dropped = false;
    for byte in input.bytes() {
        let byte = byte.map_err(|_| "could not read the terminal")?;
        match byte {
            b'\n' | b'\r' | END_OF_INPUT => break,
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
    if dropped {
        return Err("the value exceeds 64 KiB".into());
    }
    Ok(line)
}

/// Asks for the `provider` key on the terminal stdin is, with echo off, and returns the one
/// usable value typed (the rule and the limits of [`crate::key_vault::read_value`]). The
/// terminal's modes are put back before this returns, and when a signal ends the process.
/// The caller checks first that stdin is a terminal.
#[cfg(unix)]
pub fn read_hidden(provider: &str) -> Result<String> {
    use std::io::Write;
    let lang = crate::ui::ui_lang(crate::ui::Stream::Stderr);
    let question = crate::ui_text::fill(
        crate::ui_text::Msg::KeySetPrompt,
        lang,
        &[("provider", provider)],
    );
    let line = {
        // Echo goes off before the question shows, so nothing typed after it is shown.
        let _quiet = unix::Quiet::start()?;
        let mut err = std::io::stderr();
        let _ = write!(err, "{question}: ");
        let _ = err.flush();
        let line = read_line(std::io::stdin().lock());
        // The Enter that ended the line was not echoed either.
        let _ = writeln!(err);
        line?
    };
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
    /// The signals that end a process waiting at a prompt: Ctrl-C, `kill`, a closed terminal
    /// window, Ctrl-\.
    const ENDING: [libc::c_int; 4] = [libc::SIGINT, libc::SIGTERM, libc::SIGHUP, libc::SIGQUIT];

    /// The terminal's modes as they were, for the signal handler. Written once by
    /// [`Quiet::start`] before `ARMED` is set and before the handler is installed; read by the
    /// handler only while `ARMED`.
    static mut SAVED: MaybeUninit<libc::termios> = MaybeUninit::uninit();
    static ARMED: AtomicBool = AtomicBool::new(false);

    /// Puts the saved modes back, then lets the signal end the process the default way (the
    /// exit status a shell sees is the signal's, as without the prompt). Only
    /// async-signal-safe calls: `tcsetattr`, `signal`, `raise`.
    extern "C" fn restore_then_end(signal: libc::c_int) {
        if ARMED.load(Ordering::SeqCst) {
            // SAFETY: `SAVED` was initialized before `ARMED` was set and is not written
            // while armed; tcsetattr only reads it.
            unsafe { libc::tcsetattr(FD, libc::TCSANOW, addr_of!(SAVED).cast()) };
        }
        // SAFETY: plain libc calls. The signal is blocked while its handler runs, so the
        // raised one is delivered, with the default action, when the handler returns.
        unsafe {
            libc::signal(signal, libc::SIG_DFL);
            libc::raise(signal);
        }
    }

    /// The terminal with echo off, until dropped.
    pub struct Quiet {
        saved: libc::termios,
        actions: [libc::sigaction; ENDING.len()],
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
            // SAFETY: a zeroed sigaction is a valid value; sigaction fills the old action.
            let mut actions: [libc::sigaction; ENDING.len()] = unsafe { std::mem::zeroed() };
            for (signal, old) in ENDING.iter().zip(actions.iter_mut()) {
                // SAFETY: the handler is an `extern "C" fn(c_int)` and stays valid for the
                // life of the process.
                unsafe {
                    let mut new: libc::sigaction = std::mem::zeroed();
                    new.sa_sigaction = restore_then_end as extern "C" fn(libc::c_int) as usize;
                    libc::sigemptyset(&mut new.sa_mask);
                    libc::sigaction(*signal, &new, old);
                }
            }
            // From here on a drop puts everything back, a failed tcsetattr included.
            let quiet = Quiet { saved, actions };
            let mut hidden = saved;
            hidden.c_lflag &= !(libc::ECHO | libc::ECHONL | libc::ICANON | libc::IEXTEN);
            hidden.c_cc[libc::VMIN] = 1;
            hidden.c_cc[libc::VTIME] = 0;
            hidden.c_cc[libc::VSUSP] = libc::_POSIX_VDISABLE;
            // TCSAFLUSH: what was typed before the question is dropped, not taken as the key.
            // SAFETY: tcsetattr reads the struct it is given.
            if unsafe { libc::tcsetattr(FD, libc::TCSAFLUSH, &hidden) } != 0 {
                return Err(failed.into());
            }
            Ok(quiet)
        }
    }

    impl Drop for Quiet {
        fn drop(&mut self) {
            // SAFETY: tcsetattr and sigaction read the structs saved by `start`.
            unsafe {
                libc::tcsetattr(FD, libc::TCSANOW, &self.saved);
                ARMED.store(false, Ordering::SeqCst);
                for (signal, old) in ENDING.iter().zip(self.actions.iter()) {
                    libc::sigaction(*signal, old, std::ptr::null_mut());
                }
            }
        }
    }
}

#[cfg(test)]
mod tests {
    use super::*;
    use crate::key_vault::{read_value, VALUE_FILE_MAX, VALUE_MAX};

    fn line(typed: &[u8]) -> Vec<u8> {
        read_line(typed).unwrap()
    }

    #[test]
    fn a_line_ends_at_enter_ctrl_d_or_the_end_of_input() {
        assert_eq!(line(b"abc\nrest"), b"abc");
        assert_eq!(line(b"abc\rrest"), b"abc");
        assert_eq!(line(b"abc\x04rest"), b"abc");
        assert_eq!(line(b"abc"), b"abc");
        assert_eq!(line(b""), b"");
        assert_eq!(line(b"\n"), b"");
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

    /// The vault's limit at its edge, through the line reader: 65,535 bytes are a value and
    /// 65,536 are refused, with or without the line end. Literal sizes.
    #[test]
    fn the_size_limit_is_the_vaults() {
        assert_eq!((VALUE_MAX, VALUE_FILE_MAX), (65_535, 65_536));
        let n = "k".repeat(65_535);
        assert_eq!(read_value(line(n.as_bytes()).as_slice()).unwrap(), n);
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
        assert_eq!(
            read_line(far.as_bytes()).unwrap_err(),
            "the value exceeds 64 KiB"
        );
        let cut = format!("{}\x7f\x7f\x7f\n", "k".repeat(65_537));
        assert_eq!(
            read_line(cut.as_bytes()).unwrap_err(),
            "the value exceeds 64 KiB"
        );
        let again = format!("{}\x15ok\n", "k".repeat(65_537));
        assert_eq!(line(again.as_bytes()), b"ok");
    }
}
