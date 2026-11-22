//! The wake path of the host adapter (SPEC-20261122-session-messages): one
//! cross-session message to a flow session by name. The message names only the
//! ledger change to read — the ledger row is the record, the message only
//! wakes — and the send result's three shapes are read, never re-sent.
//!
//! This module is the decision layer: it validates the wake, classifies the
//! refusal reasons of REQ-2, and reports the result shape of REQ-3. The
//! transport itself is the [`PeerMessaging`] trait, so the same rules govern
//! the real socket transport and the test fake; no host-terminal path is ever
//! taken (REQ-4).

/// Longest wake text, in characters, matching the lane-send guard's limit.
pub const MAX_TEXT_CHARS: usize = 2000;
/// Longest addressee name, in characters.
pub const MAX_NAME_CHARS: usize = 64;

/// One reachable peer session as the listing answers it.
#[derive(Debug, Clone, PartialEq, Eq)]
pub struct Peer {
    /// The name other sessions message the peer by.
    pub name: String,
    /// The peer's inbox socket path, when it binds one the wake can reach.
    pub socket: Option<String>,
    /// True when the peer's channel is on (the listing could see it at all).
    pub channel: bool,
}

impl Peer {
    /// True when the peer binds an inbox the wake can reach.
    pub fn inbox(&self) -> bool {
        self.socket.as_deref().is_some_and(|s| !s.is_empty())
    }
}

/// The transport: which peers are reachable and how one send answers. The real
/// socket transport and the test fake implement this with the same rules.
pub trait PeerMessaging {
    /// The listing of reachable peers of this session (REQ-2's first look).
    fn list(&self) -> Result<Vec<Peer>, String>;
    /// One send to a peer; the answer is the result shape, never a retry.
    fn send(&self, peer: &Peer, text: &str, notify_when_idle: bool) -> Result<SendShape, String>;
}

/// The three result shapes of a send (REQ-3, the docs' three shapes): delivered
/// to the inbox, held with a delivery notice, or refused.
#[derive(Debug, Clone, PartialEq, Eq)]
pub enum SendShape {
    /// The text reached the addressee's inbox.
    Delivered,
    /// The addressee's inbound control holds the message and a delivery notice
    /// follows; the wake reports it and sends nothing more.
    HeldWithNotice,
    /// The send was refused outright; the reason names the constraint.
    Refused(Refusal),
}

/// The refusal reasons of REQ-2, one per constraint, each a single line.
#[derive(Debug, Clone, Copy, PartialEq, Eq)]
pub enum Refusal {
    /// The addressee lives under a different config root the listing cannot see.
    DifferentRoot,
    /// This session has no channel at all (the shared flag slot, the version).
    NoChannel,
    /// The addressee is listed but binds no inbox (a bare worker).
    NoInbox,
}

impl Refusal {
    /// The one-line reason, quoted from the SPEC's REQ-2.
    pub fn line(self) -> &'static str {
        match self {
            Refusal::DifferentRoot => "wake refused: the addressee is in a different config root",
            Refusal::NoChannel => {
                "wake refused: this session has no message channel (the flag slot or the version)"
            }
            Refusal::NoInbox => "wake refused: the addressee binds no inbox",
        }
    }
}

/// A wake to send: the addressee by name, the text, and whether the one-shot
/// idle notice is attached instead of a polling loop (REQ-9).
#[derive(Debug, Clone, PartialEq, Eq)]
pub struct Wake {
    pub addressee: String,
    pub text: String,
    pub notify_when_idle: bool,
}

/// The refusal of an invalid wake, before any transport is asked.
#[derive(Debug, Clone, PartialEq, Eq)]
pub enum Invalid {
    EmptyName,
    LongName,
    BadNameChar,
    EmptyText,
    LongText,
    MultiLineText,
}

impl Invalid {
    pub fn line(&self) -> &'static str {
        match self {
            Invalid::EmptyName => "wake refused: the addressee name is empty",
            Invalid::LongName => "wake refused: the addressee name is too long",
            Invalid::BadNameChar => {
                "wake refused: the addressee name carries a character outside letters, digits, - and _"
            }
            Invalid::EmptyText => "wake refused: the text is empty",
            Invalid::LongText => "wake refused: the text is too long",
            Invalid::MultiLineText => "wake refused: the text is one line, the ledger holds the rest",
        }
    }
}

/// A name answers to letters, digits, hyphen and underscore (the runtime's own
/// rule for a settable name), so a wake never types into a terminal.
fn valid_name(name: &str) -> bool {
    !name.is_empty()
        && name.chars().count() <= MAX_NAME_CHARS
        && name
            .chars()
            .all(|c| c.is_ascii_alphanumeric() || matches!(c, '-' | '_'))
}

impl Wake {
    /// The shape checks of REQ-1's text: one line, not empty, within the cap.
    pub fn validate(&self) -> Result<(), Invalid> {
        if self.addressee.is_empty() {
            return Err(Invalid::EmptyName);
        }
        if self.addressee.chars().count() > MAX_NAME_CHARS {
            return Err(Invalid::LongName);
        }
        if !valid_name(&self.addressee) {
            return Err(Invalid::BadNameChar);
        }
        if self.text.is_empty() {
            return Err(Invalid::EmptyText);
        }
        if self.text.chars().count() > MAX_TEXT_CHARS {
            return Err(Invalid::LongText);
        }
        if self.text.lines().count() > 1 {
            return Err(Invalid::MultiLineText);
        }
        Ok(())
    }

    /// Resolve the addressee against the listing and send once (REQ-1, REQ-2,
    /// REQ-3): a listing failure is the no-channel refusal, an addressee the
    /// listing cannot see is the different-root refusal, a listed peer without
    /// an inbox is the no-inbox refusal, and one send's shape is the answer —
    /// a held or refused wake is reported, never re-sent.
    pub fn send_once(&self, transport: &dyn PeerMessaging) -> Result<SendShape, String> {
        self.validate()
            .map_err(|invalid| invalid.line().to_string())?;
        let peers = match transport.list() {
            Ok(peers) => peers,
            Err(_) => {
                return Ok(SendShape::Refused(Refusal::NoChannel));
            }
        };
        let Some(peer) = peers.iter().find(|peer| peer.name == self.addressee) else {
            return Ok(SendShape::Refused(Refusal::DifferentRoot));
        };
        if !peer.channel {
            return Ok(SendShape::Refused(Refusal::NoChannel));
        }
        if !peer.inbox() {
            return Ok(SendShape::Refused(Refusal::NoInbox));
        }
        transport.send(peer, &self.text, self.notify_when_idle)
    }
}

/// The registry listing of the sessions this config root can see (measured
/// 2026-11-22: one JSON per live session under `<config dir>/sessions/`, with
/// `name` and `messagingSocketPath`). A session with a socket path binds an
/// inbox; the listing itself is the channel.
#[derive(Debug)]
pub struct RegistryListing {
    peers: Vec<Peer>,
}

impl RegistryListing {
    /// Read the listing of one config root. A folder that cannot be read is an
    /// error the wake reports as the no-channel refusal.
    pub fn read(config_dir: &std::path::Path) -> Result<Self, String> {
        let sessions = config_dir.join("sessions");
        let entries = std::fs::read_dir(&sessions).map_err(|e| e.to_string())?;
        let mut peers = Vec::new();
        for entry in entries.flatten() {
            let path = entry.path();
            if path.extension().is_none_or(|e| e != "json") {
                continue;
            }
            let Ok(text) = std::fs::read_to_string(&path) else {
                continue;
            };
            let Ok(value) = serde_json::from_str::<serde_json::Value>(&text) else {
                continue;
            };
            let Some(name) = value["name"].as_str() else {
                continue;
            };
            let socket = value["messagingSocketPath"].as_str().map(str::to_string);
            peers.push(Peer {
                name: name.to_string(),
                socket: socket.filter(|s| !s.is_empty()),
                channel: true,
            });
        }
        peers.sort_by(|a, b| a.name.cmp(&b.name));
        Ok(Self { peers })
    }
}

impl PeerMessaging for RegistryListing {
    fn list(&self) -> Result<Vec<Peer>, String> {
        Ok(self.peers.clone())
    }
    fn send(&self, peer: &Peer, text: &str, _idle: bool) -> Result<SendShape, String> {
        // The wire format the leader's research measured (2026-11-22): one
        // line of JSON ending in a newline, one direction, close after; the
        // minimum line a receiver accepts is a user message. moai sends
        // exactly that line — no auth token (none is read), no from-mode
        // (never faked). A bypass receiver holds the message unless accept is
        // set (moai's install default), and the status notice goes to a sender
        // socket moai does not have, so the answer here is the queued shape:
        // the line left this process for the addressee's inbox.
        let socket = peer
            .socket
            .clone()
            .ok_or_else(|| "the addressee binds no inbox".to_string())?;
        let line = serde_json::json!({
            "type": "user",
            "message": {"role": "user", "content": text},
        })
        .to_string();
        write_line_to_socket(&socket, &line)
    }
}

/// Write one complete line to the addressee's inbox socket and close. The
/// inbox socket is a Unix domain socket on macOS and Linux (a named pipe on
/// native Windows, whose auth line is required — that port stays open).
#[cfg(unix)]
fn write_line_to_socket(socket: &str, line: &str) -> Result<SendShape, String> {
    use std::io::Write as _;
    let mut stream = std::os::unix::net::UnixStream::connect(socket)
        .map_err(|e| format!("wake send could not reach {socket}: {e}"))?;
    stream
        .write_all(format!("{line}\n").as_bytes())
        .and_then(|_| stream.flush())
        .map_err(|e| format!("wake send failed: {e}"))?;
    drop(stream);
    Ok(SendShape::Delivered)
}

/// On a platform without the Unix domain inbox socket the wake answers the
/// open port: the documented wake path there is a Claude session's own tool.
#[cfg(not(unix))]
fn write_line_to_socket(_socket: &str, _line: &str) -> Result<SendShape, String> {
    Err("wake send is unix-only: on this platform the wake path is a Claude session's SendMessage tool".into())
}

#[cfg(test)]
mod tests {
    use super::*;

    /// The registry listing reads one JSON per live session (measured shape):
    /// the name answers, a socket path binds an inbox, a session without one
    /// does not, and an unreadable folder is the no-channel answer.
    #[test]
    fn the_registry_listing_reads_names_and_inboxes() {
        let dir = tempfile::tempdir().unwrap();
        let sessions = dir.path().join("sessions");
        std::fs::create_dir_all(&sessions).unwrap();
        std::fs::write(
            sessions.join("1.json"),
            r#"{"name":"leader","messagingSocketPath":"/tmp/cc-socks/1.sock"}"#,
        )
        .unwrap();
        std::fs::write(sessions.join("2.json"), r#"{"name":"worker"}"#).unwrap();
        std::fs::write(sessions.join("note.txt"), "not a session").unwrap();
        let listing = RegistryListing::read(dir.path()).unwrap();
        assert_eq!(
            listing.list().unwrap(),
            vec![
                Peer {
                    name: "leader".into(),
                    socket: Some("/tmp/cc-socks/1.sock".into()),
                    channel: true
                },
                Peer {
                    name: "worker".into(),
                    socket: None,
                    channel: true
                },
            ]
        );
        let missing = RegistryListing::read(&dir.path().join("nope")).unwrap_err();
        assert!(!missing.is_empty());
    }

    /// The wire write the leader's research measured: one line of JSON ending
    /// in a newline, one direction, no auth line, no from-mode. The listener
    /// here stands for the addressee's inbox exactly as a session binds one.
    #[cfg(unix)]
    #[test]
    fn the_send_writes_the_minimum_user_line_to_the_socket() {
        let dir = tempfile::tempdir().unwrap();
        let socket_path = dir.path().join("wake.sock");
        let listener = std::os::unix::net::UnixListener::bind(&socket_path).unwrap();
        let sessions = dir.path().join("sessions");
        std::fs::create_dir_all(&sessions).unwrap();
        std::fs::write(
            sessions.join("1.json"),
            format!(
                r#"{{"name":"leader","messagingSocketPath":"{}"}}"#,
                socket_path.display()
            ),
        )
        .unwrap();
        let listing = RegistryListing::read(dir.path()).unwrap();
        let reader = std::thread::spawn(move || {
            let (mut stream, _) = listener.accept().unwrap();
            use std::io::Read as _;
            let mut line = String::new();
            stream.read_to_string(&mut line).unwrap();
            line
        });
        let shape = wake("leader", "a verdict landed; read the card timeline").send_once(&listing);
        assert_eq!(shape, Ok(SendShape::Delivered));
        let line = reader.join().unwrap();
        assert!(
            line.ends_with('\n') && !line.ends_with("\n\n"),
            "one newline: {line:?}"
        );
        let value: serde_json::Value = serde_json::from_str(line.trim()).unwrap();
        assert_eq!(
            value,
            serde_json::json!({
                "type": "user",
                "message": {"role": "user", "content": "a verdict landed; read the card timeline"},
            }),
            "exactly the minimum user line, nothing else"
        );
    }

    /// A fake transport answering from a fixed listing and recording the one
    /// send it allows (REQ-3: no path here sends twice).
    struct Fake {
        peers: Vec<Peer>,
        shape: SendShape,
        sends: std::cell::RefCell<Vec<(String, bool)>>,
    }

    impl PeerMessaging for Fake {
        fn list(&self) -> Result<Vec<Peer>, String> {
            Ok(self.peers.clone())
        }
        fn send(&self, peer: &Peer, _text: &str, idle: bool) -> Result<SendShape, String> {
            self.sends.borrow_mut().push((peer.name.clone(), idle));
            Ok(self.shape.clone())
        }
    }

    fn wake(name: &str, text: &str) -> Wake {
        Wake {
            addressee: name.into(),
            text: text.into(),
            notify_when_idle: false,
        }
    }

    #[test]
    fn a_valid_wake_sends_once_and_reports_its_shape() {
        let fake = Fake {
            peers: vec![Peer {
                name: "leader".into(),
                socket: Some("/tmp/none.sock".into()),
                channel: true,
            }],
            shape: SendShape::Delivered,
            sends: std::cell::RefCell::new(vec![]),
        };
        assert_eq!(
            wake("leader", "a verdict landed; read the card timeline").send_once(&fake),
            Ok(SendShape::Delivered)
        );
        assert_eq!(fake.sends.borrow().len(), 1, "exactly one send");
    }

    #[test]
    fn an_unknown_addressee_is_the_different_root_refusal() {
        let fake = Fake {
            peers: vec![],
            shape: SendShape::Delivered,
            sends: std::cell::RefCell::new(vec![]),
        };
        let answer = wake("leader", "wake").send_once(&fake);
        assert_eq!(answer, Ok(SendShape::Refused(Refusal::DifferentRoot)));
        assert!(fake.sends.borrow().is_empty(), "nothing was sent");
    }

    #[test]
    fn a_listing_failure_is_the_no_channel_refusal() {
        struct None;
        impl PeerMessaging for None {
            fn list(&self) -> Result<Vec<Peer>, String> {
                Err("no channel".into())
            }
            fn send(&self, _: &Peer, _: &str, _: bool) -> Result<SendShape, String> {
                unreachable!("a listing failure never sends")
            }
        }
        assert_eq!(
            wake("leader", "wake").send_once(&None),
            Ok(SendShape::Refused(Refusal::NoChannel))
        );
    }

    #[test]
    fn a_listed_peer_without_inbox_is_the_no_inbox_refusal() {
        let fake = Fake {
            peers: vec![Peer {
                name: "worker".into(),
                socket: None,
                channel: true,
            }],
            shape: SendShape::Delivered,
            sends: std::cell::RefCell::new(vec![]),
        };
        assert_eq!(
            wake("worker", "wake").send_once(&fake),
            Ok(SendShape::Refused(Refusal::NoInbox))
        );
    }

    #[test]
    fn a_held_shape_is_reported_and_never_re_sent() {
        let fake = Fake {
            peers: vec![Peer {
                name: "leader".into(),
                socket: Some("/tmp/none.sock".into()),
                channel: true,
            }],
            shape: SendShape::HeldWithNotice,
            sends: std::cell::RefCell::new(vec![]),
        };
        assert_eq!(
            wake("leader", "wake").send_once(&fake),
            Ok(SendShape::HeldWithNotice)
        );
        assert_eq!(fake.sends.borrow().len(), 1, "held is not re-sent");
    }

    #[test]
    fn invalid_wakes_are_refused_before_any_transport() {
        for (name, text, why) in [
            ("", "wake", "empty name"),
            ("bad name", "wake", "name with a space"),
            ("x".repeat(65).as_str(), "wake", "long name"),
            ("leader", "", "empty text"),
            ("leader", &"x".repeat(2001), "long text"),
            ("leader", "two\nlines", "multi-line text"),
        ] {
            let fake = Fake {
                peers: vec![Peer {
                    name: "leader".into(),
                    socket: Some("/tmp/none.sock".into()),
                    channel: true,
                }],
                shape: SendShape::Delivered,
                sends: std::cell::RefCell::new(vec![]),
            };
            let answer = Wake {
                addressee: name.into(),
                text: text.into(),
                notify_when_idle: false,
            }
            .send_once(&fake);
            assert!(answer.is_err(), "{why} must be refused");
            assert!(fake.sends.borrow().is_empty(), "{why} reached no transport");
        }
    }
}
