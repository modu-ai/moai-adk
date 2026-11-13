//! Helpers shared by integration tests.
pub mod git;
use moai_cli::forge::transport::{Exchange, HttpRequest, HttpResponse};
use std::path::Path;

/// #2699: `Command::env_clear` for a test's child that keeps the developer folder the test
/// process got from `.cargo/config.toml` (`DEVELOPER_DIR`). Without it the macOS `/usr/bin/git`
/// shim of every git call under the child's fresh temporary HOME starts
/// `xcodebuild -license check`. The value is passed on only when it names a folder that exists,
/// so a Mac without that folder keeps its shim's own choice. Tests call `.clean_env()` and never
/// clear a child's environment themselves (`tests/cargo_env_developer_dir.rs` scans for it).
#[allow(dead_code)] // test crates that never clear a child's environment include this module too
pub trait CleanEnv {
    fn clean_env(&mut self) -> &mut Self;
}

impl CleanEnv for std::process::Command {
    fn clean_env(&mut self) -> &mut Self {
        self.env_clear();
        if let Some(dir) = std::env::var_os("DEVELOPER_DIR").filter(|d| Path::new(d).is_dir()) {
            self.env("DEVELOPER_DIR", dir);
        }
        self
    }
}

/// A recorded response with a JSON body and no headers (#2256). Built through the
/// constructor, so a new optional `HttpResponse` field never touches the fixtures.
#[allow(dead_code)] // only the forge/claim test crates replay exchanges
pub fn json_response(status: u16, body: serde_json::Value) -> HttpResponse {
    HttpResponse::new(status).with_body(body)
}

/// One recorded exchange: a body-less request answered by `response` (#2256).
#[allow(dead_code)] // only the forge/claim test crates replay exchanges
pub fn exchange(method: &str, url: &str, response: HttpResponse) -> Exchange {
    Exchange {
        request: HttpRequest::new(method, url),
        response,
    }
}

/// Writes an executable test script (mode 0755) through a short-lived `sh` child, so this
/// multi-threaded test process never holds a write fd on the script (#2283).
///
/// With an in-process `fs::write`, a parallel test thread that forks while the write fd is
/// open hands a copy of that fd to its child, which keeps it until its own exec. On Linux,
/// exec of a file that any process holds open for writing fails with ETXTBSY ("executable
/// file busy"), so the script's first spawn can fail at random (rc-ci run 42573077303).
/// macOS does not refuse such an exec, so the race only shows on Linux.
#[allow(dead_code)] // test crates that only replay exchanges include this module too (#2256)
pub fn write_executable(path: &Path, content: &str) {
    #[cfg(unix)]
    {
        use std::process::{Command, Stdio};
        let mut child = Command::new("sh")
            .args(["-c", "cat > \"$1\" && chmod 755 \"$1\"", "sh"])
            .arg(path)
            .stdin(Stdio::piped())
            .spawn()
            .expect("spawn sh to write a test script");
        // If the redirection fails, sh exits without reading; the status check below then
        // reports the real failure instead of an EPIPE panic.
        feed_stdin(&mut child, content.as_bytes());
        let status = child.wait().expect("wait for sh");
        assert!(status.success(), "writing {} failed", path.display());
    }
    #[cfg(not(unix))]
    std::fs::write(path, content).unwrap();
}

/// Copies an executable (mode 0755) through a short-lived `sh` child, for the same reason as
/// [`write_executable`]: an in-process `fs::copy` holds a write fd on `dest` that a concurrent
/// fork can inherit. A temp-name write plus rename does not help — the inherited fd still
/// refers to the renamed inode.
#[allow(dead_code)] // only some test crates copy a binary
pub fn copy_executable(src: &Path, dest: &Path) {
    #[cfg(unix)]
    {
        let status = std::process::Command::new("sh")
            .args(["-c", "cp \"$1\" \"$2\" && chmod 755 \"$2\"", "sh"])
            .arg(src)
            .arg(dest)
            .status()
            .expect("spawn sh to copy a test executable");
        assert!(
            status.success(),
            "copying {} to {} failed",
            src.display(),
            dest.display()
        );
    }
    #[cfg(not(unix))]
    {
        std::fs::copy(src, dest).unwrap();
    }
}

/// Writes `bytes` to a spawned child's stdin and closes it (#2348, PR #2382 CI).
///
/// A child may exit before reading its stdin (a usage error, a refused request), and the
/// write then races that exit: once the read end is gone the write fails with EPIPE. The
/// tests assert the child's exit status and output, never that the write completed, so a
/// broken pipe is tolerated and any other write error still fails the test.
#[allow(dead_code)] // not every test crate feeds a child's stdin
pub fn feed_stdin(child: &mut std::process::Child, bytes: &[u8]) {
    let pipe = child.stdin.take().expect("child stdin is piped");
    write_stdin(pipe, bytes);
}

/// The write half of [`feed_stdin`] for a pipe already taken from the child; the pipe is
/// closed on return.
#[allow(dead_code)] // not every test crate feeds a child's stdin
pub fn write_stdin(mut pipe: std::process::ChildStdin, bytes: &[u8]) {
    use std::io::Write;
    stdin_write_outcome(pipe.write_all(bytes));
}

/// Accepts a completed write or a broken pipe; panics on any other write error.
#[allow(dead_code)] // not every test crate feeds a child's stdin
pub fn stdin_write_outcome(result: std::io::Result<()>) {
    match result {
        Ok(()) => {}
        Err(e) if e.kind() == std::io::ErrorKind::BrokenPipe => {}
        Err(e) => panic!("stdin write failed: {e}"),
    }
}
