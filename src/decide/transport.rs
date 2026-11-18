//! The one place of `moai decide` that sends (slice S4, SPEC-20261118-decide-round-trip):
//! the bytes of a [`Prepared`] request go to a loopback HTTP server and the response body
//! comes back for [`conclude`](super::conclude). Plain HTTP/1.1 over `std::net`, no new
//! dependency.
//!
//! In this slice nothing leaves the machine: [`post_loopback`] refuses any address that is not
//! a loopback address before it opens a socket, so the key it carries cannot go anywhere else.
//! The real endpoints are https; speaking to them needs a TLS client, which is a later slice.
//!
//! Every refusal is a [`Refusal`] with a fixed code and a fixed sentence: the response is never
//! quoted and the key is in no text. The caller treats a refusal like any other: the model does
//! not decide, the leader does.

use super::{Prepared, Record, Refusal};
use std::io::{ErrorKind, Read, Write};
use std::net::{SocketAddr, TcpStream};
use std::time::{Duration, Instant};

/// The longest key the transport accepts, in bytes.
pub const KEY_MAX: usize = 4096;

/// A provider key held for one request. It has no `Display`, and its `Debug` prints nothing of
/// the value.
pub struct ApiKey(String);

impl ApiKey {
    /// The key `text`, or a refusal when it is empty, longer than [`KEY_MAX`] or holds a byte
    /// that is not a visible ASCII character (a space, a control character and any
    /// non-ASCII byte could change the header the key goes into).
    pub fn new(text: &str) -> Result<Self, Refusal> {
        let visible = text.bytes().all(|b| (0x21..=0x7e).contains(&b));
        if text.is_empty() || text.len() > KEY_MAX || !visible {
            return Err(Refusal::new(
                "key_invalid",
                "the key is empty, too long or holds a byte that is not a visible ASCII character: the leader decides",
            ));
        }
        Ok(ApiKey(text.to_string()))
    }
}

impl std::fmt::Debug for ApiKey {
    fn fmt(&self, f: &mut std::fmt::Formatter<'_>) -> std::fmt::Result {
        f.write_str("ApiKey(..)")
    }
}

/// The bounds of one exchange.
#[derive(Debug, Clone, Copy, PartialEq, Eq)]
pub struct Limits {
    /// Longest wait for the connection.
    pub connect: Duration,
    /// Longest time for the whole exchange, connection included.
    pub total: Duration,
    /// Largest response body, in bytes.
    pub response_max: usize,
}

impl Default for Limits {
    fn default() -> Self {
        Limits {
            connect: Duration::from_secs(3),
            total: Duration::from_secs(20),
            response_max: 64 * 1024,
        }
    }
}

/// POST the body of `prepared` to the path of its provider endpoint on the loopback `addr` and
/// return the response body.
pub fn post_loopback(
    addr: SocketAddr,
    prepared: &Prepared,
    key: &ApiKey,
    limits: &Limits,
) -> Result<Vec<u8>, Refusal> {
    // Before any socket: the key and the request go nowhere but this machine.
    if !addr.ip().is_loopback() {
        return Err(fixed(
            "transport_address",
            "the address is not a loopback address: nothing is sent",
        ));
    }
    let started = Instant::now();
    let deadline = started + limits.total;
    let connect = limits.connect.min(limits.total);
    let mut stream = TcpStream::connect_timeout(&addr, connect).map_err(|e| match e.kind() {
        ErrorKind::TimedOut | ErrorKind::WouldBlock => timeout(),
        _ => fixed(
            "transport_connect",
            "the connection was refused: the leader decides",
        ),
    })?;
    stream
        .set_write_timeout(Some(left(deadline)?))
        .map_err(|_| io_refusal())?;
    let mut request = Vec::with_capacity(prepared.body().len() + 512);
    request.extend_from_slice(
        format!(
            "POST {} HTTP/1.1\r\nHost: {addr}\r\nUser-Agent: moai-decide\r\nAccept: application/json\r\nContent-Type: application/json\r\nAuthorization: Bearer {}\r\nContent-Length: {}\r\nConnection: close\r\n\r\n",
            path_of(prepared.endpoint()),
            key.0,
            prepared.body().len()
        )
        .as_bytes(),
    );
    request.extend_from_slice(prepared.body());
    stream.write_all(&request).map_err(write_refusal)?;
    drop(request);
    read_response(&mut stream, deadline, limits.response_max)
}

/// The path of an endpoint: what follows its host.
fn path_of(endpoint: &str) -> &str {
    endpoint
        .split_once("://")
        .and_then(|(_, rest)| rest.find('/').map(|at| &rest[at..]))
        .unwrap_or("/")
}

/// The most header bytes the transport reads before it gives up on a response.
const HEAD_MAX: usize = 16 * 1024;

fn fixed(code: &'static str, reason: &'static str) -> Refusal {
    Refusal::new(code, reason)
}

fn timeout() -> Refusal {
    fixed(
        "transport_timeout",
        "the server did not answer in time: the leader decides",
    )
}

fn io_refusal() -> Refusal {
    fixed("transport_io", "the connection failed: the leader decides")
}

fn unreadable(why: &'static str) -> Refusal {
    fixed("transport_response", why)
}

fn write_refusal(e: std::io::Error) -> Refusal {
    match e.kind() {
        ErrorKind::TimedOut | ErrorKind::WouldBlock => timeout(),
        _ => io_refusal(),
    }
}

/// The time left before `deadline`, or the timeout refusal.
fn left(deadline: Instant) -> Result<Duration, Refusal> {
    let left = deadline.saturating_duration_since(Instant::now());
    if left.is_zero() {
        Err(timeout())
    } else {
        Ok(left)
    }
}

/// One read within the deadline: the bytes read, 0 for the end of the stream.
fn read_some(
    stream: &mut TcpStream,
    deadline: Instant,
    chunk: &mut [u8],
) -> Result<usize, Refusal> {
    loop {
        stream
            .set_read_timeout(Some(left(deadline)?))
            .map_err(|_| io_refusal())?;
        match stream.read(chunk) {
            Ok(n) => return Ok(n),
            Err(e) if e.kind() == ErrorKind::Interrupted => {}
            Err(e) if matches!(e.kind(), ErrorKind::TimedOut | ErrorKind::WouldBlock) => {
                return Err(timeout())
            }
            // A reset after the request was sent is the server closing early.
            Err(_) => return Ok(0),
        }
    }
}

fn find(hay: &[u8], needle: &[u8]) -> Option<usize> {
    hay.windows(needle.len()).position(|w| w == needle)
}

/// The response body of one HTTP/1.1 answer, or the refusal for everything else: a status
/// other than 200, a response without a plain `Content-Length`, one larger than `max` or one
/// that stops short. Nothing of the response is quoted.
fn read_response(
    stream: &mut TcpStream,
    deadline: Instant,
    max: usize,
) -> Result<Vec<u8>, Refusal> {
    let mut got: Vec<u8> = Vec::new();
    let mut chunk = [0u8; 2048];
    let head_end = loop {
        if let Some(at) = find(&got, b"\r\n\r\n") {
            break at;
        }
        if got.len() > HEAD_MAX {
            return Err(unreadable(
                "the response headers are too long: the leader decides",
            ));
        }
        match read_some(stream, deadline, &mut chunk)? {
            0 => {
                return Err(unreadable(
                    "the server closed the connection before it answered: the leader decides",
                ))
            }
            n => got.extend_from_slice(&chunk[..n]),
        }
    };
    let head = std::str::from_utf8(&got[..head_end])
        .map_err(|_| unreadable("the response headers are not text: the leader decides"))?;
    let mut lines = head.split("\r\n");
    let status = lines
        .next()
        .and_then(|line| {
            let mut parts = line.split_whitespace();
            let version = parts.next()?;
            let code = parts.next()?;
            version.starts_with("HTTP/1.").then_some(())?;
            code.parse::<u16>().ok()
        })
        .ok_or_else(|| unreadable("the status line is not HTTP/1.x: the leader decides"))?;
    match status {
        200 => {}
        300..=399 => {
            return Err(fixed(
                "transport_redirect",
                "the server answered with a redirect, which is never followed: the leader decides",
            ))
        }
        401 | 403 => {
            return Err(fixed(
                "transport_unauthorized",
                "the server did not accept the key: the leader decides",
            ))
        }
        429 => {
            return Err(fixed(
                "transport_rate_limited",
                "the server asked to slow down: the leader decides",
            ))
        }
        500..=599 => {
            return Err(fixed(
                "transport_server_error",
                "the server failed: the leader decides",
            ))
        }
        _ => {
            return Err(fixed(
                "transport_status",
                "the server answered with a status other than 200: the leader decides",
            ))
        }
    }
    let mut length: Option<usize> = None;
    for line in lines {
        let Some((name, value)) = line.split_once(':') else {
            return Err(unreadable(
                "a response header has no name: the leader decides",
            ));
        };
        let name = name.trim();
        if name.eq_ignore_ascii_case("transfer-encoding") {
            return Err(unreadable(
                "the response uses a transfer encoding, which is not read: the leader decides",
            ));
        }
        if name.eq_ignore_ascii_case("content-length") {
            let value = value.trim();
            let plain =
                !value.is_empty() && value.len() <= 12 && value.bytes().all(|b| b.is_ascii_digit());
            let parsed = value.parse::<usize>().ok().filter(|_| plain);
            match (parsed, length) {
                (None, _) => {
                    return Err(unreadable(
                        "`Content-Length` is not a number: the leader decides",
                    ))
                }
                (Some(n), Some(before)) if n != before => {
                    return Err(unreadable(
                        "`Content-Length` is given twice: the leader decides",
                    ))
                }
                (Some(n), _) => length = Some(n),
            }
        }
    }
    let length = length
        .ok_or_else(|| unreadable("the response has no `Content-Length`: the leader decides"))?;
    if length > max {
        return Err(fixed(
            "transport_too_large",
            "the response is larger than the bound: the leader decides",
        ));
    }
    let mut body = got[head_end + 4..].to_vec();
    body.truncate(length);
    while body.len() < length {
        match read_some(stream, deadline, &mut chunk)? {
            0 => {
                return Err(unreadable(
                    "the response ended before its `Content-Length`: the leader decides",
                ))
            }
            n => {
                let take = n.min(length - body.len());
                body.extend_from_slice(&chunk[..take]);
            }
        }
    }
    Ok(body)
}

/// [`post_loopback`], then [`conclude`](super::conclude) on the body that came back.
pub fn round_trip_loopback(
    addr: SocketAddr,
    prepared: &Prepared,
    key: &ApiKey,
    limits: &Limits,
) -> Result<Record, Refusal> {
    let body = post_loopback(addr, prepared, key, limits)?;
    super::conclude(prepared, &body)
}

#[cfg(test)]
mod tests {
    use super::*;
    use crate::decide::prepare;
    use crate::decide::provider::Provider;
    use crate::decide::testing::{request_json, who};
    use std::io::{Read, Write};
    use std::net::{TcpListener, TcpStream};
    use std::sync::{Arc, Mutex};
    use std::thread::JoinHandle;
    use std::time::Instant;

    /// Not a key: made up for these tests, and the only key text they use.
    const KEY: &str = "FAKEKEY-not-a-real-key-0123456789";

    const TYPESAFE_ANSWER: &str = r#"{"model":"jev-test-1","answers":{"decision":{"type":"choice","choice":"nextest","confidence":0.9,"probabilities":{"nextest":0.8,"cargo-test":0.2}}}}"#;
    const OPENAI_ANSWER: &str = r#"{"answers":[{"type":"choice","name":"decision","choice":"nextest","confidence":0.8,"probabilities":[{"value":"nextest","probability":0.7},{"value":"cargo-test","probability":0.3}]}]}"#;

    /// What the fake server does after it has read the request.
    enum Script {
        Reply(Vec<u8>),
        /// Reads the request, then closes without a byte.
        Close,
        /// Waits, then replies.
        Delay(Duration, Vec<u8>),
    }

    struct Fake {
        addr: SocketAddr,
        seen: Arc<Mutex<Vec<u8>>>,
        handle: JoinHandle<()>,
    }

    impl Fake {
        /// The raw request the server read, once its thread has ended.
        fn finish(self) -> Vec<u8> {
            self.handle.join().unwrap();
            let seen = self.seen.lock().unwrap().clone();
            seen
        }
    }

    fn find(hay: &[u8], needle: &[u8]) -> Option<usize> {
        hay.windows(needle.len()).position(|w| w == needle)
    }

    /// One request, headers and the body its `Content-Length` names.
    fn read_request(stream: &mut TcpStream) -> Vec<u8> {
        stream
            .set_read_timeout(Some(Duration::from_secs(5)))
            .unwrap();
        let mut got = Vec::new();
        let mut chunk = [0u8; 1024];
        loop {
            if let Some(end) = find(&got, b"\r\n\r\n") {
                let head = String::from_utf8_lossy(&got[..end]).to_lowercase();
                let want = head
                    .lines()
                    .find_map(|l| l.strip_prefix("content-length:"))
                    .and_then(|v| v.trim().parse::<usize>().ok())
                    .unwrap_or(0);
                if got.len() >= end + 4 + want {
                    return got;
                }
            }
            match stream.read(&mut chunk) {
                Ok(0) | Err(_) => return got,
                Ok(n) => got.extend_from_slice(&chunk[..n]),
            }
        }
    }

    fn serve(script: Script) -> Fake {
        let listener = TcpListener::bind("127.0.0.1:0").unwrap();
        let addr = listener.local_addr().unwrap();
        let seen = Arc::new(Mutex::new(Vec::new()));
        let log = Arc::clone(&seen);
        let handle = std::thread::spawn(move || {
            // No connection within the wait: end, so a test whose transport never connects
            // fails instead of hanging in `join`.
            listener.set_nonblocking(true).unwrap();
            let until = Instant::now() + Duration::from_secs(3);
            let mut stream = loop {
                match listener.accept() {
                    Ok((stream, _)) => break stream,
                    Err(_) if Instant::now() < until => {
                        std::thread::sleep(Duration::from_millis(5));
                    }
                    Err(_) => return,
                }
            };
            stream.set_nonblocking(false).unwrap();
            *log.lock().unwrap() = read_request(&mut stream);
            match script {
                Script::Close => {}
                Script::Reply(bytes) => {
                    let _ = stream.write_all(&bytes);
                }
                Script::Delay(wait, bytes) => {
                    std::thread::sleep(wait);
                    let _ = stream.write_all(&bytes);
                }
            }
        });
        Fake { addr, seen, handle }
    }

    /// A response with the given status line, extra header lines and body.
    fn reply(status: &str, extra: &str, body: &str) -> Vec<u8> {
        format!(
            "HTTP/1.1 {status}\r\nContent-Type: application/json\r\n{extra}Content-Length: {}\r\nConnection: close\r\n\r\n{body}",
            body.len()
        )
        .into_bytes()
    }

    fn ok(body: &str) -> Script {
        Script::Reply(reply("200 OK", "", body))
    }

    fn prepared(provider: Provider) -> Prepared {
        prepare(&request_json(|_| {}), provider, &who()).unwrap()
    }

    fn key() -> ApiKey {
        ApiKey::new(KEY).unwrap()
    }

    fn quick() -> Limits {
        Limits {
            connect: Duration::from_secs(2),
            total: Duration::from_secs(3),
            response_max: 4096,
        }
    }

    /// The refusal of one exchange with a fake server that does `script`.
    fn refused(script: Script, limits: &Limits) -> Refusal {
        let fake = serve(script);
        let result = round_trip_loopback(fake.addr, &prepared(Provider::Typesafe), &key(), limits);
        let _ = fake.finish();
        result.unwrap_err()
    }

    #[test]
    fn a_prepared_request_makes_a_round_trip_for_each_provider() {
        for (provider, path, answer, model) in [
            (
                Provider::Typesafe,
                "/v1/systemone",
                TYPESAFE_ANSWER,
                "model:typesafe/jev-test-1",
            ),
            (
                Provider::Openai,
                "/v1/decisions",
                OPENAI_ANSWER,
                "model:openai/gpt-6-luna",
            ),
        ] {
            let sent = prepared(provider);
            let fake = serve(ok(answer));
            let record = round_trip_loopback(fake.addr, &sent, &key(), &quick()).unwrap();
            let seen = fake.finish();
            let text = String::from_utf8_lossy(&seen).to_string();
            assert!(
                text.starts_with(&format!("POST {path} HTTP/1.1\r\n")),
                "{provider:?}: {text}"
            );
            for header in [
                format!("\r\nAuthorization: Bearer {KEY}\r\n"),
                "\r\nContent-Type: application/json\r\n".to_string(),
                format!("\r\nContent-Length: {}\r\n", sent.body().len()),
                "\r\nConnection: close\r\n".to_string(),
            ] {
                assert!(text.contains(&header), "{provider:?} lacks {header:?}");
            }
            let split = find(&seen, b"\r\n\r\n").unwrap() + 4;
            assert_eq!(&seen[split..], sent.body(), "{provider:?}: the body bytes");
            assert_eq!(record.chosen, "nextest", "{provider:?}");
            assert_eq!(record.decided_by, model, "{provider:?}");
        }
    }

    #[test]
    fn a_non_loopback_address_or_a_bad_key_is_refused_before_any_socket() {
        // 192.0.2.1 is TEST-NET-1: a connect to it would wait for the whole connect bound.
        let started = Instant::now();
        let limits = Limits {
            connect: Duration::from_secs(5),
            total: Duration::from_secs(5),
            response_max: 4096,
        };
        for addr in ["192.0.2.1:9", "10.0.0.1:9", "[2001:db8::1]:9", "0.0.0.0:9"] {
            let addr: SocketAddr = addr.parse().unwrap();
            let r =
                post_loopback(addr, &prepared(Provider::Typesafe), &key(), &limits).unwrap_err();
            assert_eq!(r.code, "transport_address", "{addr}");
        }
        assert!(
            started.elapsed() < Duration::from_secs(2),
            "a refused address must not wait for a connection"
        );
        let long = "k".repeat(KEY_MAX + 1);
        for bad in [
            "",
            "has space",
            "tab\there",
            "new\nline",
            "carriage\rreturn",
            "nul\0byte",
            "ünïcode",
            long.as_str(),
        ] {
            let r = ApiKey::new(bad).unwrap_err();
            assert_eq!(r.code, "key_invalid", "{bad:?}");
        }
        assert!(ApiKey::new(&"k".repeat(KEY_MAX)).is_ok());
        assert!(ApiKey::new("sk-abc_DEF.123~+/=").is_ok());
    }

    /// The transport takes a `SocketAddr`, never a name: a host name cannot be parsed into one,
    /// and the file resolves nothing, so a name cannot lead a request off the machine.
    #[test]
    fn a_named_address_cannot_reach_the_transport() {
        for named in [
            "example.com:443",
            "localhost:80",
            "api.openai.com:443",
            "127.0.0.1.example.com:80",
            "http://127.0.0.1:80",
        ] {
            assert!(named.parse::<SocketAddr>().is_err(), "{named}");
        }
        let text = include_str!("transport.rs");
        let code = text.split("#[cfg(test)]").next().unwrap();
        for resolver in [
            "ToSocketAddrs",
            "to_socket_addrs",
            "lookup_host",
            "getaddrinfo",
        ] {
            assert!(
                !code.contains(resolver),
                "the transport resolves names: {resolver}"
            );
        }
    }

    /// `ApiKey` has no `Display`: a key cannot be formatted with `{}` by mistake.
    #[test]
    fn the_key_type_has_no_display() {
        use std::fmt::Display;
        use std::marker::PhantomData;
        struct Probe<T>(PhantomData<T>);
        trait Fallback {
            const HAS_DISPLAY: bool = false;
        }
        impl<T> Fallback for Probe<T> {}
        impl<T: Display> Probe<T> {
            const HAS_DISPLAY: bool = true;
        }
        // The probe sees `Display` on a type that has it, and not on the key.
        assert!(Probe::<String>::HAS_DISPLAY);
        assert!(!Probe::<ApiKey>::HAS_DISPLAY);
    }

    #[test]
    fn the_key_is_in_no_debug_output_and_no_refusal() {
        assert!(!format!("{:?}", key()).contains(KEY));
        let slow = Limits {
            connect: Duration::from_secs(2),
            total: Duration::from_millis(150),
            response_max: 4096,
        };
        let big = Limits {
            response_max: 8,
            ..quick()
        };
        let cases = [
            refused(Script::Close, &quick()),
            refused(Script::Reply(reply("401 Unauthorized", "", "{}")), &quick()),
            refused(Script::Reply(reply("500 Oops", "", "{}")), &quick()),
            refused(
                Script::Reply(reply("302 Found", "Location: /elsewhere\r\n", "")),
                &quick(),
            ),
            refused(ok(TYPESAFE_ANSWER), &big),
            refused(ok("not json"), &quick()),
            refused(
                Script::Delay(
                    Duration::from_millis(500),
                    reply("200 OK", "", TYPESAFE_ANSWER),
                ),
                &slow,
            ),
        ];
        // Each case is refused for the reason it was built for: a stub that refuses
        // everything does not pass this test.
        let codes: Vec<&str> = cases.iter().map(|r| r.code).collect();
        assert_eq!(
            codes,
            [
                "transport_response",
                "transport_unauthorized",
                "transport_server_error",
                "transport_redirect",
                "transport_too_large",
                "answer_invalid",
                "transport_timeout"
            ]
        );
        for r in cases {
            let shown = format!("{r:?} {} {}", r.reason, r.to_json());
            assert!(!shown.contains(KEY), "{shown}");
            assert!(!shown.contains("Bearer"), "{shown}");
        }
    }

    #[test]
    fn a_status_other_than_200_is_a_fixed_refusal_and_a_redirect_is_not_followed() {
        let echo = r#"{"detail":"SECRET-ECHO-OF-THE-SERVER"}"#;
        for (status, code) in [
            ("401 Unauthorized", "transport_unauthorized"),
            ("403 Forbidden", "transport_unauthorized"),
            ("429 Too Many Requests", "transport_rate_limited"),
            ("500 Internal Server Error", "transport_server_error"),
            ("503 Service Unavailable", "transport_server_error"),
            ("404 Not Found", "transport_status"),
            ("204 No Content", "transport_status"),
        ] {
            let r = refused(Script::Reply(reply(status, "", echo)), &quick());
            assert_eq!(r.code, code, "{status}");
            let shown = format!("{r:?} {}", r.to_json());
            assert!(!shown.contains("SECRET-ECHO"), "{status}: {shown}");
        }
        // A redirect is a refusal; the second server is never asked.
        let second = TcpListener::bind("127.0.0.1:0").unwrap();
        second.set_nonblocking(true).unwrap();
        // A network-path reference (no scheme): still an absolute redirect to another server.
        let location = format!(
            "Location: //{}/v1/systemone\r\n",
            second.local_addr().unwrap()
        );
        for status in [
            "301 Moved Permanently",
            "302 Found",
            "307 Temporary Redirect",
        ] {
            let r = refused(Script::Reply(reply(status, &location, "")), &quick());
            assert_eq!(r.code, "transport_redirect", "{status}");
        }
        assert!(
            second.accept().is_err(),
            "the transport followed a redirect to the second server"
        );
    }

    #[test]
    fn a_response_that_is_too_big_short_early_closed_or_slow_is_refused() {
        // Too big: the Content-Length alone is enough, the body is not read.
        let big = Limits {
            response_max: 16,
            ..quick()
        };
        assert_eq!(
            refused(ok(TYPESAFE_ANSWER), &big).code,
            "transport_too_large"
        );
        // Exactly at the bound is fine for the transport (the answer is then judged).
        let body = "x".repeat(16);
        let fake = serve(ok(&body));
        let at_bound = post_loopback(fake.addr, &prepared(Provider::Typesafe), &key(), &big);
        let _ = fake.finish();
        assert_eq!(at_bound.unwrap(), body.as_bytes());
        // Short: fewer bytes than Content-Length, then the server closes.
        let short =
            b"HTTP/1.1 200 OK\r\nContent-Length: 100\r\nConnection: close\r\n\r\n0123456789"
                .to_vec();
        assert_eq!(
            refused(Script::Reply(short), &quick()).code,
            "transport_response"
        );
        // Closed early: nothing at all.
        assert_eq!(refused(Script::Close, &quick()).code, "transport_response");
        // No Content-Length, a chunked body, a bad length, a bad status line.
        for raw in [
            "HTTP/1.1 200 OK\r\nConnection: close\r\n\r\n{}",
            "HTTP/1.1 200 OK\r\nTransfer-Encoding: chunked\r\n\r\n2\r\n{}\r\n0\r\n\r\n",
            // A transfer encoding with a good length too: still not read.
            "HTTP/1.1 200 OK\r\nTransfer-Encoding: chunked\r\nContent-Length: 2\r\n\r\n{}",
            "HTTP/1.1 200 OK\r\nContent-Length: abc\r\n\r\n{}",
            // Two lengths that disagree, with a body that fits the second one.
            "HTTP/1.1 200 OK\r\nContent-Length: 2\r\nContent-Length: 3\r\n\r\n{}x",
            "HTTP/1.1 200 OK\r\nContent-Length: -2\r\n\r\n{}",
            "garbage\r\n\r\n",
            "FTP/1.1 200 OK\r\nContent-Length: 2\r\n\r\n{}",
        ] {
            assert_eq!(
                refused(Script::Reply(raw.as_bytes().to_vec()), &quick()).code,
                "transport_response",
                "{raw:?}"
            );
        }
        // A header block that never ends is cut at its bound.
        let endless = format!("HTTP/1.1 200 OK\r\nX-Pad: {}", "p".repeat(40_000));
        assert_eq!(
            refused(Script::Reply(endless.into_bytes()), &quick()).code,
            "transport_response"
        );
        // Slow: the server answers after the total bound.
        let slow = Limits {
            connect: Duration::from_secs(2),
            total: Duration::from_millis(150),
            response_max: 4096,
        };
        let started = Instant::now();
        let r = refused(
            Script::Delay(
                Duration::from_millis(600),
                reply("200 OK", "", TYPESAFE_ANSWER),
            ),
            &slow,
        );
        assert_eq!(r.code, "transport_timeout");
        assert!(
            started.elapsed() < Duration::from_millis(590) + Duration::from_millis(300),
            "the refusal waited for the server"
        );
        // Nobody listens: a refused connection is a connect refusal.
        let closed = TcpListener::bind("127.0.0.1:0").unwrap();
        let addr = closed.local_addr().unwrap();
        drop(closed);
        let r = post_loopback(addr, &prepared(Provider::Typesafe), &key(), &quick()).unwrap_err();
        assert_eq!(r.code, "transport_connect");
    }
}
