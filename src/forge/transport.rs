//! HTTP exchange seam for the GitHub adapter (07-A #2008).
//!
//! RC1 interim decision (leader, 2026-11-08): the real transport is a `gh api` subprocess.
//! The token stays in gh's keyring or the caller's `GH_TOKEN`; this program never reads it.
//! Tests replay recorded exchanges through [`RecordedTransport`] and never touch the network.
use serde::{Deserialize, Serialize};
use std::{
    cell::RefCell,
    collections::BTreeMap,
    io::Write,
    process::{Command, Stdio},
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

impl HttpResponse {
    pub fn header(&self, name: &str) -> Option<&str> {
        self.headers
            .get(&name.to_ascii_lowercase())
            .map(String::as_str)
    }
}

pub trait Transport {
    fn send(&self, req: &HttpRequest) -> Result<HttpResponse>;
}

/// Removes token-shaped text from anything this program may surface (REQ-FORGE-010).
pub fn redact(text: &str) -> String {
    text.split_inclusive(char::is_whitespace)
        .map(|word| {
            let w = word.trim_end();
            let token = ["ghp_", "gho_", "ghu_", "ghs_", "ghr_", "github_pat_"]
                .iter()
                .any(|p| w.contains(p));
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
}

impl GhTransport {
    /// `gh` from PATH. `MOAI_FORGE_GH_BIN` is a test-only seam that swaps the executable; it
    /// must never be set on a product path. To keep it from silently picking another program
    /// through PATH it is accepted only as an absolute path to an existing regular file.
    pub fn from_env() -> Result<Self> {
        let Some(raw) = std::env::var_os("MOAI_FORGE_GH_BIN") else {
            return Ok(GhTransport { bin: "gh".into() });
        };
        let path = std::path::PathBuf::from(&raw);
        let regular = std::fs::metadata(&path)
            .map(|m| m.is_file())
            .unwrap_or(false);
        match path.to_str() {
            Some(s) if path.is_absolute() && regular => Ok(GhTransport { bin: s.into() }),
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
    Some(HttpResponse {
        status,
        headers,
        body,
    })
}

impl Transport for GhTransport {
    fn send(&self, req: &HttpRequest) -> Result<HttpResponse> {
        if !req.url.starts_with("https://api.github.com/") {
            return Err("refused: only https://api.github.com/ URLs are requested".into());
        }
        let mut cmd = Command::new(&self.bin);
        cmd.args(["api", "-X", &req.method, "-i"]);
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
        let mut child = cmd
            .spawn()
            .map_err(|e| format!("cannot start {}: {}", self.bin, e.kind()))?;
        if let (Some(body), Some(mut stdin)) = (&req.body, child.stdin.take()) {
            // Ignore a broken pipe: gh may exit before reading a body it rejects.
            let _ = stdin.write_all(body.to_string().as_bytes());
        }
        let out = child
            .wait_with_output()
            .map_err(|_| "gh did not finish".to_string())?;
        let stdout = String::from_utf8_lossy(&out.stdout);
        parse_gh_output(&stdout).ok_or_else(|| {
            let err = redact(String::from_utf8_lossy(&out.stderr).trim());
            format!(
                "gh api returned no HTTP response (exit {:?}): {}",
                out.status.code(),
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
