//! Helpers shared by integration tests.
use std::path::Path;

/// Writes an executable test script (mode 0755) through a short-lived `sh` child, so this
/// multi-threaded test process never holds a write fd on the script (#2283).
///
/// With an in-process `fs::write`, a parallel test thread that forks while the write fd is
/// open hands a copy of that fd to its child, which keeps it until its own exec. On Linux,
/// exec of a file that any process holds open for writing fails with ETXTBSY ("executable
/// file busy"), so the script's first spawn can fail at random (rc-ci run 42573077303).
/// macOS does not refuse such an exec, so the race only shows on Linux.
pub fn write_executable(path: &Path, content: &str) {
    #[cfg(unix)]
    {
        use std::io::Write;
        use std::process::{Command, Stdio};
        let mut child = Command::new("sh")
            .args(["-c", "cat > \"$1\" && chmod 755 \"$1\"", "sh"])
            .arg(path)
            .stdin(Stdio::piped())
            .spawn()
            .expect("spawn sh to write a test script");
        child
            .stdin
            .take()
            .expect("sh stdin")
            .write_all(content.as_bytes())
            .expect("write test script to sh");
        let status = child.wait().expect("wait for sh");
        assert!(status.success(), "writing {} failed", path.display());
    }
    #[cfg(not(unix))]
    std::fs::write(path, content).unwrap();
}
