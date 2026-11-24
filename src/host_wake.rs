//! The wake path of the host adapter (SPEC-20261122-session-messages): one
//! cross-session message to a flow session by name. The message names only the
//! ledger change to read — the ledger row is the record, the message only
//! wakes — and the send result's shapes are read: queued (the line left this
//! process for the addressee's inbox socket), held with a notice, or refused;
//! a refused or held wake is reported, never re-sent.
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

/// The result shapes of a send (REQ-3, the docs' shapes): queued to the
/// addressed session's inbox, held with a delivery notice, or refused.
#[derive(Debug, Clone, PartialEq, Eq)]
pub enum SendShape {
    /// The line left this process for the addressee's inbox socket — queued,
    /// not delivered: the held-or-delivered status goes to a sender socket
    /// this command does not have (N3 of the first audit).
    Queued,
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
    /// The name answers to more than one live session; the wake refuses to
    /// pick (N10 of the first audit).
    Ambiguous,
    /// The wake asked for the idle notice, which the minimum line does not
    /// carry (N7 of the first audit).
    IdleUnsupported,
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
            Refusal::Ambiguous => {
                "wake refused: the name answers to more than one live session; address one from its own terminal or rename it first"
            }
            Refusal::IdleUnsupported => {
                "wake refused: the idle notice rides the model-session tool; the minimum line carries none"
            }
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
            Invalid::EmptyText => "wake refused: the text is empty or only a newline",
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
        // N1 of the audit: a text of only a newline (or any whitespace) is
        // empty — trim before the line checks.
        if self.text.trim().is_empty() {
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
        if self.notify_when_idle {
            return Ok(SendShape::Refused(Refusal::IdleUnsupported));
        }
        let peers = match transport.list() {
            Ok(peers) => peers,
            Err(_) => {
                return Ok(SendShape::Refused(Refusal::NoChannel));
            }
        };
        let matches: Vec<&Peer> = peers
            .iter()
            .filter(|peer| peer.name == self.addressee)
            .collect();
        match matches.as_slice() {
            [] => return Ok(SendShape::Refused(Refusal::DifferentRoot)),
            [peer] => {
                if !peer.channel {
                    return Ok(SendShape::Refused(Refusal::NoChannel));
                }
                if !peer.inbox() {
                    return Ok(SendShape::Refused(Refusal::NoInbox));
                }
            }
            _ => return Ok(SendShape::Refused(Refusal::Ambiguous)),
        }
        let peer = matches[0];
        transport.send(peer, &self.text, self.notify_when_idle)
    }
}

/// The registry listing of the sessions this config root can see (measured
/// 2026-11-22: one JSON per live session under `<config dir>/sessions/`, with
/// `name` and `messagingSocketPath`). A session with a socket path binds an
/// inbox; the listing itself is the channel. A symlinked sessions folder is
/// followed — the listing only reads, and the socket-path guard checks every
/// component between the socket and /tmp before any connect (follow-up 3 of
/// the landed follow-ups: decided, read-only).
#[derive(Debug)]
pub struct RegistryListing {
    peers: Vec<Peer>,
}

impl RegistryListing {
    /// Read the listing of one config root. A folder that cannot be read is
    /// an error the operation answers with the sessions folder named.
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

/// Check the registry's socket path before any connect (the audit's B1 and
/// the landed slice's follow-up): the path must be absolute, no component of
/// it may be a symbolic link (a linked socket or a linked parent folder is
/// refused, so a moved registry entry cannot redirect the wake), its parent
/// must be the registry's own socket folder — a directory directly under
/// `/tmp` whose name starts with `cc-socks`, owned by this user and closed
/// against group and world writes (/tmp itself is world-writable) — and the
/// entry itself must be a socket file, so a dead path or a regular file never
/// reaches `connect`.
#[cfg(unix)]
fn checked_socket_path(raw: &str) -> Result<std::path::PathBuf, String> {
    use std::os::unix::fs::{FileTypeExt as _, MetadataExt as _};
    let refuse = |why: &str| Err(format!("wake refused: the socket path {why}"));
    let path = std::path::Path::new(raw);
    if !path.is_absolute() {
        return refuse("is not absolute");
    }
    // Every component between the socket and /tmp must be real; /tmp itself
    // is the system's own link on macOS (/private/tmp) and the root is /.
    for ancestor in path.ancestors().skip(1) {
        if ancestor == std::path::Path::new("/tmp") || ancestor == std::path::Path::new("/") {
            continue;
        }
        match std::fs::symlink_metadata(ancestor) {
            Ok(data) if !data.file_type().is_symlink() => {}
            _ => return refuse("has a missing or linked component"),
        }
    }
    let parent = path
        .parent()
        .filter(|p| p.parent() == Some(std::path::Path::new("/tmp")));
    let folder = parent
        .and_then(|p| p.file_name())
        .map(|n| n.to_string_lossy().into_owned());
    match folder {
        Some(name) if name.starts_with("cc-socks") => {}
        _ => return refuse("is outside the registry's socket folder"),
    }
    let parent = parent.expect("the folder check above passed with this parent");
    // 9 of the N-slice audit: the owner and mode of the folder itself, read
    // with symlink_metadata (the folder is already link-checked above, so the
    // lstat cannot follow a link); the pure predicate below is what the
    // mutation-pinned tests drive.
    match std::fs::symlink_metadata(parent).map(|data| (data.uid(), data.mode())) {
        Ok((folder_uid, mode)) if folder_safe(folder_uid, self_uid(), mode) => {}
        Ok((folder_uid, _)) if folder_uid != self_uid() => {
            return refuse("is in a socket folder another user owns")
        }
        Ok(_) => return refuse("is in a socket folder open to group or world writes"),
        Err(_) => return refuse("is a dead path"),
    }
    match std::fs::symlink_metadata(path) {
        Ok(data) if data.file_type().is_symlink() => refuse("is a linked socket"),
        Ok(data) if data.file_type().is_socket() => Ok(path.to_path_buf()),
        Ok(_) => refuse("is not a socket file"),
        Err(_) => refuse("is a dead path"),
    }
}

/// The pure folder predicate (pinned by its own tests, N-slice mutants 1
/// and 2): the socket folder belongs to this user and carries neither the
/// group-write nor the world-write bit — each bit alone refuses.
#[cfg(unix)]
fn folder_safe(folder_uid: u32, own_uid: u32, mode: u32) -> bool {
    folder_uid == own_uid && (mode & 0o022) == 0
}

/// The current process's user id (the folder another account could have
/// created in world-writable /tmp is what the owner check exists for).
#[cfg(unix)]
fn self_uid() -> u32 {
    static UID: std::sync::OnceLock<u32> = std::sync::OnceLock::new();
    *UID.get_or_init(|| {
        // Safety: getuid reads the process's real user id and touches no
        // memory of ours; it is async-signal-safe by POSIX.
        unsafe { libc::getuid() }
    })
}

/// Write one complete line to the addressee's inbox socket and close. The
/// inbox socket is a Unix domain socket on macOS and Linux (a named pipe on
/// native Windows, whose auth line is required — that port stays open).
#[cfg(unix)]
fn write_line_to_socket(socket: &str, line: &str) -> Result<SendShape, String> {
    use std::io::Write as _;
    let path = checked_socket_path(socket)?;
    let mut stream = std::os::unix::net::UnixStream::connect(&path)
        .map_err(|e| format!("wake send could not reach {}: {e}", path.display()))?;
    stream
        .write_all(format!("{line}\n").as_bytes())
        .and_then(|_| stream.flush())
        .map_err(|e| format!("wake send failed: {e}"))?;
    drop(stream);
    Ok(SendShape::Queued)
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

    /// A test-owned socket folder under /tmp with the registry folder's name
    /// shape, its mode set explicitly (the audit's N8 and item 12: never the
    /// real /tmp/cc-socks folder, never the umask's default). The Drop
    /// removes the folder, panic or not.
    struct SockFolder(std::path::PathBuf);
    impl SockFolder {
        fn new(tag: &str) -> Self {
            let path = std::path::Path::new("/tmp")
                .join(format!("cc-socks-moai-test-{tag}-{}", std::process::id()));
            let _ = std::fs::remove_dir_all(&path);
            std::fs::create_dir_all(&path).unwrap();
            #[cfg(unix)]
            {
                use std::os::unix::fs::PermissionsExt as _;
                std::fs::set_permissions(&path, std::fs::Permissions::from_mode(0o755)).unwrap();
            }
            Self(path)
        }
        fn socket(&self, name: &str) -> String {
            self.0.join(name).display().to_string()
        }
    }
    impl Drop for SockFolder {
        fn drop(&mut self) {
            let _ = std::fs::remove_dir_all(&self.0);
        }
    }

    /// One config root whose single session names the given socket path.
    fn one_peer_config(socket: &str) -> tempfile::TempDir {
        let dir = tempfile::tempdir().unwrap();
        let sessions = dir.path().join("sessions");
        std::fs::create_dir_all(&sessions).unwrap();
        std::fs::write(
            sessions.join("1.json"),
            format!(r#"{{"name":"leader","messagingSocketPath":"{socket}"}}"#),
        )
        .unwrap();
        dir
    }

    /// The registry listing reads one JSON per live session (measured shape):
    /// the name answers, a socket path binds an inbox, a session without one
    /// does not, and an unreadable folder answers with the folder named.
    #[test]
    fn the_registry_listing_reads_names_and_inboxes() {
        let socks = SockFolder::new("listing");
        let dir = tempfile::tempdir().unwrap();
        let sessions = dir.path().join("sessions");
        std::fs::create_dir_all(&sessions).unwrap();
        std::fs::write(
            sessions.join("1.json"),
            format!(
                r#"{{"name":"leader","messagingSocketPath":"{}"}}"#,
                socks.socket("1.sock")
            ),
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
                    socket: Some(socks.socket("1.sock")),
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
    /// here stands for the addressee's inbox exactly as a session binds one,
    /// in a test-owned folder with the registry folder's name shape.
    #[cfg(unix)]
    #[test]
    fn the_send_writes_the_minimum_user_line_to_the_socket() {
        let socks = SockFolder::new("wire");
        let socket_path = std::path::PathBuf::from(socks.socket("wake.sock"));
        let listener = std::os::unix::net::UnixListener::bind(&socket_path).unwrap();
        let dir = one_peer_config(&socket_path.display().to_string());
        let listing = RegistryListing::read(dir.path()).unwrap();
        let reader = std::thread::spawn(move || {
            let (mut stream, _) = listener.accept().unwrap();
            use std::io::Read as _;
            let mut line = String::new();
            stream.read_to_string(&mut line).unwrap();
            line
        });
        let shape = wake("leader", "a verdict landed; read the card timeline").send_once(&listing);
        assert_eq!(shape, Ok(SendShape::Queued));
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

    /// The audit's B1: a symlinked socket is refused — the registry entry may
    /// not redirect the wake to another path.
    #[cfg(unix)]
    #[test]
    fn a_linked_socket_is_refused() {
        let socks = SockFolder::new("link");
        let real = socks.socket("real.sock");
        let _listener = std::os::unix::net::UnixListener::bind(&real).unwrap();
        let link = socks.socket("link.sock");
        std::os::unix::fs::symlink(&real, &link).unwrap();
        let dir = one_peer_config(&link);
        let listing = RegistryListing::read(dir.path()).unwrap();
        let answer = wake("leader", "wake").send_once(&listing);
        assert!(
            answer.as_ref().is_err_and(|line| line.contains("linked")),
            "{answer:?}"
        );
    }

    /// The landed slice's follow-up: a linked parent FOLDER is refused — the
    /// registry entry may not route the wake through a symlinked directory
    /// even when both ends carry the socket folder's name. The guard removes
    /// the link on every path out (the audit's item 11: a failing run must
    /// not leak it into /tmp).
    #[cfg(unix)]
    #[test]
    fn a_linked_parent_folder_is_refused() {
        struct LinkGuard(std::path::PathBuf);
        impl Drop for LinkGuard {
            fn drop(&mut self) {
                let _ = std::fs::remove_file(&self.0);
            }
        }
        let real_folder = SockFolder::new("dirlink-real");
        let real = real_folder.socket("wake.sock");
        let _listener = std::os::unix::net::UnixListener::bind(&real).unwrap();
        let linked_folder = std::path::Path::new("/tmp")
            .join(format!("cc-socks-moai-test-dirlink-{}", std::process::id()));
        let _ = std::fs::remove_file(&linked_folder);
        std::os::unix::fs::symlink(
            real_folder
                .0
                .parent()
                .unwrap()
                .join(real_folder.0.file_name().unwrap()),
            &linked_folder,
        )
        .unwrap();
        let _guard = LinkGuard(linked_folder.clone());
        let dir = one_peer_config(&linked_folder.join("wake.sock").display().to_string());
        let listing = RegistryListing::read(dir.path()).unwrap();
        let answer = wake("leader", "wake").send_once(&listing);
        assert!(
            answer
                .as_ref()
                .is_err_and(|line| line.contains("linked component")),
            "{answer:?}"
        );
    }

    /// The pure folder predicate (the audit's mutants 1 and 2): each bit
    /// alone refuses and the owner must be this user.
    #[cfg(unix)]
    #[test]
    fn the_folder_predicate_pins_each_bit_and_the_owner() {
        let me = self_uid();
        assert!(folder_safe(me, me, 0o755));
        assert!(folder_safe(me, me, 0o700));
        // Group-write alone (0o775, 0o720) and world-write alone (0o702).
        assert!(!folder_safe(me, me, 0o775));
        assert!(!folder_safe(me, me, 0o720));
        assert!(!folder_safe(me, me, 0o702));
        assert!(!folder_safe(me, me, 0o777));
        // Another owner refuses at every safe mode.
        assert!(!folder_safe(me.wrapping_add(1), me, 0o755));
    }

    /// The audit's B1: a dead path (nothing bound, no socket file) never
    /// reaches connect.
    #[cfg(unix)]
    #[test]
    fn a_dead_socket_is_refused() {
        let socks = SockFolder::new("dead");
        let dir = one_peer_config(&socks.socket("dead.sock"));
        let listing = RegistryListing::read(dir.path()).unwrap();
        let answer = wake("leader", "wake").send_once(&listing);
        assert!(
            answer
                .as_ref()
                .is_err_and(|line| line.contains("dead path")),
            "{answer:?}"
        );
    }

    /// The audit's B1: a regular file where a socket should be is refused.
    #[cfg(unix)]
    #[test]
    fn a_non_socket_file_is_refused() {
        let socks = SockFolder::new("plain");
        let plain = socks.socket("plain.sock");
        std::fs::write(&plain, "not a socket").unwrap();
        let dir = one_peer_config(&plain);
        let listing = RegistryListing::read(dir.path()).unwrap();
        let answer = wake("leader", "wake").send_once(&listing);
        assert!(
            answer
                .as_ref()
                .is_err_and(|line| line.contains("not a socket file")),
            "{answer:?}"
        );
    }

    /// The audit's B1: a relative socket path resolves against the caller's
    /// working directory and is refused outright.
    #[cfg(unix)]
    #[test]
    fn a_relative_socket_path_is_refused() {
        let dir = one_peer_config("sessions/1.sock");
        let listing = RegistryListing::read(dir.path()).unwrap();
        let answer = wake("leader", "wake").send_once(&listing);
        assert!(
            answer
                .as_ref()
                .is_err_and(|line| line.contains("not absolute")),
            "{answer:?}"
        );
    }

    /// The audit's N1: a text of only a newline is empty and is refused
    /// before any transport.
    #[test]
    fn a_text_of_only_a_newline_is_refused() {
        let fake = Fake {
            peers: vec![Peer {
                name: "leader".into(),
                socket: Some("/tmp/cc-socks-moai-test-fake/none.sock".into()),
                channel: true,
            }],
            shape: SendShape::Queued,
            sends: std::cell::RefCell::new(vec![]),
        };
        let answer = wake("leader", "\n").send_once(&fake);
        assert!(
            answer
                .as_ref()
                .is_err_and(|line| line.contains("only a newline")),
            "{answer:?}"
        );
        assert!(fake.sends.borrow().is_empty());
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
                socket: Some("/tmp/cc-socks-moai-test-fake/none.sock".into()),
                channel: true,
            }],
            shape: SendShape::Queued,
            sends: std::cell::RefCell::new(vec![]),
        };
        assert_eq!(
            wake("leader", "a verdict landed; read the card timeline").send_once(&fake),
            Ok(SendShape::Queued)
        );
        assert_eq!(fake.sends.borrow().len(), 1, "exactly one send");
    }

    #[test]
    fn an_unknown_addressee_is_the_different_root_refusal() {
        let fake = Fake {
            peers: vec![],
            shape: SendShape::Queued,
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
            shape: SendShape::Queued,
            sends: std::cell::RefCell::new(vec![]),
        };
        assert_eq!(
            wake("worker", "wake").send_once(&fake),
            Ok(SendShape::Refused(Refusal::NoInbox))
        );
        assert!(fake.sends.borrow().is_empty(), "a refusal sends nothing");
    }

    #[test]
    fn a_held_shape_is_reported_and_never_re_sent() {
        let fake = Fake {
            peers: vec![Peer {
                name: "leader".into(),
                socket: Some("/tmp/cc-socks-moai-test-fake/none.sock".into()),
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
                    socket: Some("/tmp/cc-socks-moai-test-fake/none.sock".into()),
                    channel: true,
                }],
                shape: SendShape::Queued,
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
