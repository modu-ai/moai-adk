//! Helpers shared by integration tests.
use moai_cli::forge::transport::{Exchange, HttpRequest, HttpResponse};
use std::path::Path;

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
