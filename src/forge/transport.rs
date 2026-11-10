//! HTTP exchange seam for the GitHub adapter (07-A #2008).
//!
//! RC1 interim decision (leader, 2026-11-08): the real transport is a `gh api` subprocess.
//! The token stays in gh's keyring or the caller's `GH_TOKEN`; this program never reads it.
//! Tests replay recorded exchanges through [`RecordedTransport`] and never touch the network.
use serde::{Deserialize, Serialize};
use std::{
    cell::RefCell,
    collections::BTreeMap,
    io::{Read, Write},
    process::{Command, Stdio},
    time::{Duration, Instant},
};

type Result<T> = std::result::Result<T, String>;

#[derive(Clone, Debug, PartialEq, Eq, Serialize, Deserialize)]
pub struct HttpRequest {
    pub method: String,
    /// Full `https://api.github.com/...` URL. `Link` URLs are passed through byte for byte.
    pub url: String,
    #[serde(default)]
    pub headers: BTreeMap<String, String>,
    #[serde(default)]
    pub body: Option<serde_json::Value>,
}

#[derive(Clone, Debug, PartialEq, Eq, Serialize, Deserialize)]
pub struct HttpResponse {
    pub status: u16,
    /// Lower-cased header names.
    #[serde(default)]
    pub headers: BTreeMap<String, String>,
    #[serde(default)]
    pub body: Option<serde_json::Value>,
}

impl HttpRequest {
    /// A request with no headers and no body (#2256). Optional fields get their defaults
    /// here, so adding one does not touch call sites.
    pub fn new(method: impl Into<String>, url: impl Into<String>) -> Self {
        HttpRequest {
            method: method.into(),
            url: url.into(),
            headers: BTreeMap::new(),
            body: None,
        }
    }

    /// Adds one header, name kept as given.
    pub fn with_header(mut self, name: impl Into<String>, value: impl Into<String>) -> Self {
        self.headers.insert(name.into(), value.into());
        self
    }

    pub fn with_body(mut self, body: serde_json::Value) -> Self {
        self.body = Some(body);
        self
    }
}

impl HttpResponse {
    /// A response with no headers and no body (#2256).
    pub fn new(status: u16) -> Self {
        HttpResponse {
            status,
            headers: BTreeMap::new(),
            body: None,
        }
    }

    /// Adds one header, name kept as given; callers pass the lower-cased name this type
    /// documents, as the literals it replaces did.
    pub fn with_header(mut self, name: impl Into<String>, value: impl Into<String>) -> Self {
        self.headers.insert(name.into(), value.into());
        self
    }

    pub fn with_body(mut self, body: serde_json::Value) -> Self {
        self.body = Some(body);
        self
    }

    pub fn header(&self, name: &str) -> Option<&str> {
        self.headers
            .get(&name.to_ascii_lowercase())
            .map(String::as_str)
    }
}

pub trait Transport {
    fn send(&self, req: &HttpRequest) -> Result<HttpResponse>;
}

/// The merged token prefix list (REQ-FORGE-010). This is the single list the
/// shared redact module consumes too (#2351) — the prefixes stay spelled here,
/// because this is where they were already merged; nothing else restates them.
pub(crate) const TOKEN_PREFIXES: [&str; 6] =
    ["ghp_", "gho_", "ghu_", "ghs_", "ghr_", "github_pat_"];

/// Removes token-shaped text from anything this program may surface (REQ-FORGE-010).
pub fn redact(text: &str) -> String {
    text.split_inclusive(char::is_whitespace)
        .map(|word| {
            let w = word.trim_end();
            let token = TOKEN_PREFIXES.iter().any(|p| w.contains(p));
            if token {
                format!("[redacted]{}", &word[w.len()..])
            } else {
                word.to_string()
            }
        })
        .collect()
}

/// `gh api -X <METHOD> -i [-H ...] <URL> [--input -]`. gh exits non-zero on 304 and on 4xx/5xx
/// while still printing the status line and headers, so the parsed response is authoritative and
/// the exit code alone is never read as success or failure.
pub struct GhTransport {
    pub bin: String,
    /// Upper bound for one `gh api` call; past it the whole process group is ended and the
    /// call fails as `transport` (#2048).
    pub timeout: Duration,
    /// A token handed to gh as `GH_TOKEN` in the child's environment only — never argv, never
    /// disk (#2173). `None` leaves gh's own authentication in charge. No `Debug` is derived on
    /// this struct; one added later must skip this field.
    pub gh_token: Option<String>,
}

// @MX:NOTE: [AUTO] default bound for one gh call; not derived from a measurement (unmeasured)
/// Default upper bound for one `gh api` call (unmeasured; chosen above gh's own retry window).
pub const GH_TIMEOUT: Duration = Duration::from_secs(60);

/// How long the output readers may take after gh ended (or was ended).
const OUTPUT_GRACE: Duration = Duration::from_secs(2);

/// Ends every process left in gh's process group (unix: `kill -KILL -- -<pgid>`, as argv), the
/// same way `verify::run_check` does. An empty group makes `kill` fail, which is fine.
fn end_group(pid: u32) {
    #[cfg(unix)]
    {
        let _ = Command::new("kill")
            .args(["-KILL", "--", &format!("-{pid}")])
            .stdin(Stdio::null())
            .stdout(Stdio::null())
            .stderr(Stdio::null())
            .status();
    }
    #[cfg(not(unix))]
    let _ = pid;
}

/// Reads a pipe on its own thread and hands the bytes over a channel, so the wait is bounded.
fn drain<R: Read + Send + 'static>(pipe: Option<R>) -> std::sync::mpsc::Receiver<Vec<u8>> {
    let (tx, rx) = std::sync::mpsc::channel();
    std::thread::spawn(move || {
        let mut buf = Vec::new();
        if let Some(mut p) = pipe {
            let _ = p.read_to_end(&mut buf);
        }
        let _ = tx.send(buf);
    });
    rx
}

impl GhTransport {
    /// `gh` at `bin` with the default timeout and gh's own authentication (#2256). Optional
    /// fields get their defaults here, so adding one does not touch call sites.
    pub fn new(bin: impl Into<String>) -> Self {
        GhTransport {
            bin: bin.into(),
            timeout: GH_TIMEOUT,
            gh_token: None,
        }
    }

    pub fn timeout(mut self, timeout: Duration) -> Self {
        self.timeout = timeout;
        self
    }

    pub fn gh_token(mut self, gh_token: Option<String>) -> Self {
        self.gh_token = gh_token;
        self
    }

    /// `gh` from PATH. `MOAI_FORGE_GH_BIN` is a test-only seam that swaps the executable; it
    /// must never be set on a product path. To keep it from silently picking another program
    /// through PATH it is accepted only as an absolute path to an existing regular file.
    pub fn from_env() -> Result<Self> {
        let Some(raw) = std::env::var_os("MOAI_FORGE_GH_BIN") else {
            return Ok(GhTransport::new("gh"));
        };
        let path = std::path::PathBuf::from(&raw);
        let regular = std::fs::metadata(&path)
            .map(|m| m.is_file())
            .unwrap_or(false);
        match path.to_str() {
            Some(s) if path.is_absolute() && regular => Ok(GhTransport::new(s)),
            _ => Err("MOAI_FORGE_GH_BIN must be an absolute path to an existing file".into()),
        }
    }
}

fn parse_gh_output(stdout: &str) -> Option<HttpResponse> {
    let (head, body) = match stdout.find("\r\n\r\n") {
        Some(i) => (&stdout[..i], &stdout[i + 4..]),
        None => match stdout.find("\n\n") {
            Some(i) => (&stdout[..i], &stdout[i + 2..]),
            None => (stdout, ""),
        },
    };
    let mut lines = head.lines();
    let status_line = lines.next()?;
    if !status_line.starts_with("HTTP/") {
        return None;
    }
    let status: u16 = status_line.split_whitespace().nth(1)?.parse().ok()?;
    let mut headers = BTreeMap::new();
    for line in lines {
        if let Some((k, v)) = line.split_once(':') {
            headers.insert(k.trim().to_ascii_lowercase(), v.trim().to_string());
        }
    }
    let body = body.trim();
    let body = if body.is_empty() {
        None
    } else {
        serde_json::from_str(body).ok()
    };
    let mut resp = HttpResponse::new(status);
    resp.headers = headers;
    resp.body = body;
    Some(resp)
}

impl Transport for GhTransport {
    fn send(&self, req: &HttpRequest) -> Result<HttpResponse> {
        if !req.url.starts_with("https://api.github.com/") {
            return Err("refused: only https://api.github.com/ URLs are requested".into());
        }
        let mut cmd = Command::new(&self.bin);
        cmd.args(["api", "-X", &req.method, "-i"]);
        if let Some(token) = &self.gh_token {
            cmd.env("GH_TOKEN", token);
        }
        for (k, v) in &req.headers {
            cmd.arg("-H").arg(format!("{k}: {v}"));
        }
        cmd.arg(&req.url);
        if req.body.is_some() {
            cmd.args(["--input", "-"]);
        }
        cmd.stdin(Stdio::piped())
            .stdout(Stdio::piped())
            .stderr(Stdio::piped());
        // Its own process group, so anything gh starts ends with it.
        #[cfg(unix)]
        std::os::unix::process::CommandExt::process_group(&mut cmd, 0);
        let started = Instant::now();
        let mut child = cmd
            .spawn()
            .map_err(|e| format!("cannot start {}: {}", self.bin, e.kind()))?;
        let stdin = child.stdin.take();
        if let (Some(body), Some(mut stdin)) = (req.body.clone(), stdin) {
            // On its own thread: a gh that never reads must not block the bounded wait below.
            // A broken pipe is ignored: gh may exit before reading a body it rejects.
            std::thread::spawn(move || {
                let _ = stdin.write_all(body.to_string().as_bytes());
            });
        }
        let out = drain(child.stdout.take());
        let err = drain(child.stderr.take());
        let status = loop {
            match child.try_wait() {
                Ok(Some(s)) => break Some(s),
                Ok(None) if started.elapsed() >= self.timeout => break None,
                Ok(None) => std::thread::sleep(Duration::from_millis(10)),
                Err(_) => break None,
            }
        };
        if status.is_none() {
            let _ = child.kill();
            let _ = child.wait();
        }
        // Whatever gh left running (children holding the pipes) ends with it.
        end_group(child.id());
        if status.is_none() {
            return Err(format!(
                "gh did not answer: timed out after {} ms",
                self.timeout.as_millis()
            ));
        }
        let stdout = out
            .recv_timeout(OUTPUT_GRACE)
            .map_err(|_| "gh output did not close after gh exited".to_string())?;
        let stderr = err.recv_timeout(OUTPUT_GRACE).unwrap_or_default();
        let stdout = String::from_utf8_lossy(&stdout);
        parse_gh_output(&stdout).ok_or_else(|| {
            let err = redact(String::from_utf8_lossy(&stderr).trim());
            format!(
                "gh api returned no HTTP response (exit {:?}): {}",
                status.and_then(|s| s.code()),
                err.chars().take(200).collect::<String>()
            )
        })
    }
}

/// One recorded request/response pair. `request` lists what the adapter must send.
#[derive(Clone, Debug, Serialize, Deserialize)]
pub struct Exchange {
    pub request: HttpRequest,
    pub response: HttpResponse,
}

/// Replays exchanges in order and keeps every request actually sent, so tests can assert
/// headers, URLs, methods and request counts.
pub struct RecordedTransport {
    exchanges: RefCell<std::collections::VecDeque<Exchange>>,
    pub sent: RefCell<Vec<HttpRequest>>,
}

impl RecordedTransport {
    pub fn new(exchanges: Vec<Exchange>) -> Self {
        RecordedTransport {
            exchanges: RefCell::new(exchanges.into()),
            sent: RefCell::new(Vec::new()),
        }
    }
    pub fn remaining(&self) -> usize {
        self.exchanges.borrow().len()
    }
}

impl Transport for RecordedTransport {
    fn send(&self, req: &HttpRequest) -> Result<HttpResponse> {
        self.sent.borrow_mut().push(req.clone());
        let next = self
            .exchanges
            .borrow_mut()
            .pop_front()
            .ok_or_else(|| format!("unexpected request {} {}", req.method, req.url))?;
        if next.request.method != req.method || next.request.url != req.url {
            return Err(format!(
                "recorded {} {} but got {} {}",
                next.request.method, next.request.url, req.method, req.url
            ));
        }
        for (k, v) in &next.request.headers {
            if req.headers.get(k) != Some(v) {
                return Err(format!("expected header {k}: {v} on {}", req.url));
            }
        }
        Ok(next.response)
    }
}

#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    fn gh_output_with_304_is_a_response() {
        let r =
            parse_gh_output("HTTP/2.0 304 Not Modified\nEtag: \"x\"\nX-Ratelimit-Remaining: 9\n\n")
                .unwrap();
        assert_eq!(r.status, 304);
        assert_eq!(r.header("ETag"), Some("\"x\""));
        assert!(r.body.is_none());
        assert!(parse_gh_output("not http").is_none());
    }

    #[test]
    fn token_shapes_are_redacted() {
        let s = redact("bad credentials ghp_abcdef123 and github_pat_11ZZ end");
        assert!(!s.contains("ghp_") && !s.contains("github_pat_"));
        assert!(s.contains("bad credentials") && s.ends_with("end"));
    }
}
