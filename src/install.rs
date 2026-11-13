//! Installation receipts and read-only install inspection (Issue #1998, 04-A).
//!
//! A receipt records one installed moai-cli binary under
//! `~/.moai-cli/installations/<install_id>.json` (REQ-04A-01). `install_id` is the
//! lowercase hex sha256 of the canonical absolute binary path
//! (`sha256-canonical-path-hex-v0`); hex only, so the file name is valid on Windows too
//! (identity's `sha256-canonical-path-v0` id carries a `sha256:` prefix that NTFS
//! rejects in a file name). Self-identification (REQ-04A-03) matches BOTH the receipt's
//! `binary_path` (its `~` token expanded against the current home) and its
//! `binary_sha256` against `std::env::current_exe`; a path match with a digest mismatch
//! is reported as `stale_receipt` and never rewritten here (REQ-04A-06; the explicit
//! update path belongs to 04-B). Nothing in this module writes to the Go generation's
//! `.moai/`, `~/.moai/` or `~/go/bin/moai`: `inspect` reports existence booleans for the
//! folders (REQ-04A-04) and, since #2791, runs each `moai` candidate (PATH, the Go install
//! places) once with `--version` to tell Rust from Go.
//!
//! The Go generation's updater selection rules are characterized in
//! `tests/install_cli.rs` against the recorded feed `tests/fixtures/go-releases-sample.json`;
//! none of that logic is implemented here.

use crate::contract::PROTOCOL_VERSION;
use crate::SECRET_SHAPES;
use serde::{Deserialize, Serialize};
use serde_json::{json, Value};
use sha2::{Digest, Sha256};
use std::{
    fs::{self, OpenOptions},
    io::{Read as _, Write as _},
    path::{Path, PathBuf},
    time::{Duration, Instant, SystemTime, UNIX_EPOCH},
};

type Result<T> = std::result::Result<T, String>;

/// Same bound as the profile files in lib.rs; receipts are smaller still.
const RECEIPT_LIMIT: u64 = 65536;
/// A release binary may legitimately be tens of MiB; this only refuses absurd inputs.
const BINARY_HASH_LIMIT: u64 = 512 * 1024 * 1024;
const PROBE_TIMEOUT: Duration = Duration::from_secs(5);
/// After a timed-out probe is killed, how long its already printed output is still awaited.
const PROBE_GRACE: Duration = Duration::from_millis(200);
/// `--version` output beyond this is discarded (drained, not stored).
const PROBE_STDOUT_CAP: usize = 4096;
/// The only `PATH` a `--version` probe sees (#2800 audit B1).
const PROBE_PATH: &str = if cfg!(windows) {
    r"C:\Windows\System32"
} else {
    "/usr/bin:/bin"
};
/// The longest `--version` first line reported (characters); the rest is cut.
const PROBE_LINE_CAP: usize = 200;
/// Every `moai` beyond this many candidates is counted, not probed (#2791 AC-2).
const CANDIDATE_LIMIT: usize = 64;
/// How `install_id` is derived; recorded in `inspect` output and docs/rust-product-path.md.
const INSTALL_ID_DERIVATION: &str = "sha256-canonical-path-hex-v0";

/// Identity fields fixed by REQ-04A-01; the schema refuses anything else on read.
pub const RECEIPT_SCHEMA_VERSION: u32 = 1;
pub const PRODUCT: &str = "moai-cli";
pub const GENERATION: &str = "rust-rc";

/// One installed moai-cli binary (REQ-04A-01). Unknown keys are refused on read; no
/// field may carry a credential-shaped value (the shared `SECRET_SHAPES` token list of
/// the profile filter).
#[derive(Clone, Debug, Serialize, Deserialize, PartialEq)]
#[serde(deny_unknown_fields)]
pub struct InstallReceipt {
    pub schema_version: u32,
    pub product: String,
    pub generation: String,
    pub channel: String,
    pub version: String,
    /// Absolute path of the installed binary; the user home is normalized to the `~` token.
    pub binary_path: String,
    /// `sha256:<64 lowercase hex>` of the binary bytes (lib.rs `digest` form).
    pub binary_sha256: String,
    pub installer_owner: InstallerOwner,
    pub protocol_version: u32,
    pub created_at: String,
    pub updated_at: String,
}

#[derive(Clone, Copy, Debug, Serialize, Deserialize, PartialEq, Eq)]
#[serde(rename_all = "lowercase")]
pub enum InstallerOwner {
    Direct,
    Cargo,
    Homebrew,
    Msi,
    Unknown,
}

/// What `plan_receipt` hands to `record_receipt`: the receipt plus its file identity.
#[derive(Clone, Debug)]
pub struct ReceiptPlan {
    pub install_id: String,
    pub receipt: InstallReceipt,
}

fn valid_sha256_field(s: &str) -> bool {
    s.len() == 71
        && s.starts_with("sha256:")
        && s[7..]
            .bytes()
            .all(|b| b.is_ascii_digit() || (b'a'..=b'f').contains(&b))
}

fn secret_shaped(s: &str) -> bool {
    let lower = s.to_ascii_lowercase();
    SECRET_SHAPES.iter().any(|shape| lower.contains(shape))
}

fn no_control(s: &str) -> bool {
    !s.chars().any(char::is_control)
}

/// channel/version: short portable words, no separators, no credential shapes.
fn portable_word(s: &str) -> bool {
    !s.trim().is_empty() && s.len() <= 64 && no_control(s) && !secret_shaped(s)
}

fn is_windows_absolute(s: &str) -> bool {
    let b = s.as_bytes();
    b.len() >= 3 && b[0].is_ascii_alphabetic() && b[1] == b':' && (b[2] == b'\\' || b[2] == b'/')
}

/// `binary_path` is a path by design, so unlike `portable` it allows `/`, `~/` and a
/// Windows drive prefix — but still refuses credential shapes and control characters.
fn valid_binary_path(s: &str) -> bool {
    let rooted = s == "~" || s.starts_with("~/") || s.starts_with('/') || is_windows_absolute(s);
    rooted && s.len() <= 2048 && no_control(s) && !secret_shaped(s)
}

/// RFC3339 UTC second precision, `YYYY-MM-DDTHH:MM:SSZ` — the only form this module writes.
fn rfc3339(s: &str) -> bool {
    let b = s.as_bytes();
    b.len() == 20
        && b[0..4].iter().all(u8::is_ascii_digit)
        && b[4] == b'-'
        && b[5..7].iter().all(u8::is_ascii_digit)
        && b[7] == b'-'
        && b[8..10].iter().all(u8::is_ascii_digit)
        && (b[10] == b'T' || b[10] == b't')
        && b[11..13].iter().all(u8::is_ascii_digit)
        && b[13] == b':'
        && b[14..16].iter().all(u8::is_ascii_digit)
        && b[16] == b':'
        && b[17..19].iter().all(u8::is_ascii_digit)
        && (b[19] == b'Z' || b[19] == b'z')
}

/// Validates a receipt against REQ-04A-01; serde has already refused unknown keys.
pub fn validate_receipt(r: &InstallReceipt) -> Result<()> {
    if r.schema_version != RECEIPT_SCHEMA_VERSION
        || r.product != PRODUCT
        || r.generation != GENERATION
        || r.protocol_version != PROTOCOL_VERSION
        || !portable_word(&r.channel)
        || !portable_word(&r.version)
        || !valid_binary_path(&r.binary_path)
        || !valid_sha256_field(&r.binary_sha256)
        || !rfc3339(&r.created_at)
        || !rfc3339(&r.updated_at)
    {
        return Err("invalid installation receipt; no values were logged".into());
    }
    Ok(())
}

fn rfc3339_now() -> String {
    let secs = SystemTime::now()
        .duration_since(UNIX_EPOCH)
        .map(|d| d.as_secs())
        .unwrap_or(0);
    let days = (secs / 86400) as i64;
    let rem = secs % 86400;
    // Howard Hinnant's civil_from_days, the standard days->Y/M/D conversion.
    let z = days + 719_468;
    let era = z.div_euclid(146_097);
    let doe = z.rem_euclid(146_097);
    let yoe = (doe - doe / 1460 + doe / 36_524 - doe / 146_096) / 365;
    let y = yoe + era * 400;
    let doy = doe - (365 * yoe + yoe / 4 - yoe / 100);
    let mp = (5 * doy + 2) / 153;
    let d = doy - (153 * mp + 2) / 5 + 1;
    let m = if mp < 10 { mp + 3 } else { mp - 9 };
    let y = if m <= 2 { y + 1 } else { y };
    format!(
        "{y:04}-{m:02}-{d:02}T{:02}:{:02}:{:02}Z",
        rem / 3600,
        (rem % 3600) / 60,
        rem % 60
    )
}

/// Replaces a leading home prefix with the `~` token (REQ-04A-01). Both inputs must be
/// canonical; when the binary lives outside the home the absolute path is kept as is.
fn normalize_home(canonical_home: &Path, canonical_exe: &Path) -> Result<String> {
    let abs = canonical_exe
        .to_str()
        .ok_or("non UTF-8 binary path")?
        .to_string();
    let home = canonical_home.to_str().ok_or("non UTF-8 home path")?;
    if let Some(rest) = abs.strip_prefix(home) {
        if rest.is_empty() {
            return Ok("~".into());
        }
        if rest.starts_with('/') || rest.starts_with('\\') {
            return Ok(format!("~{rest}"));
        }
    }
    Ok(abs)
}

/// Expands a receipt's `~` token against `home`, for comparison only.
fn expand_home(home: &Path, stored: &str) -> PathBuf {
    if stored == "~" {
        home.to_path_buf()
    } else if let Some(rest) = stored.strip_prefix("~/") {
        home.join(rest)
    } else {
        PathBuf::from(stored)
    }
}

pub(crate) fn sha256_file(path: &Path) -> Result<String> {
    let meta = fs::metadata(path).map_err(|_| "binary metadata unavailable")?;
    if !meta.is_file() {
        return Err("binary path is not a regular file".into());
    }
    if meta.len() > BINARY_HASH_LIMIT {
        return Err("binary exceeds the hashing size limit".into());
    }
    let mut file = fs::File::open(path).map_err(|_| "binary read failed")?;
    let mut hasher = Sha256::new();
    let mut buf = vec![0u8; 64 * 1024];
    loop {
        let n = file.read(&mut buf).map_err(|_| "binary read failed")?;
        if n == 0 {
            break;
        }
        hasher.update(&buf[..n]);
    }
    Ok(format!("sha256:{:x}", hasher.finalize()))
}

/// Release channel derived from the version string; provisional until the P12 release
/// matrix decides channel naming (HOLD reported, not decided here).
fn channel_of(version: &str) -> String {
    if version.contains("-rc") {
        "rc".into()
    } else {
        "stable".into()
    }
}

/// The path an installation is identified by (#2790 R1, impact table B2): the binary under
/// its canonical name. `moai` and `moai-cli` are built from one `main`; a `moai` (or
/// `moai.exe`) whose sibling `moai-cli` (`moai-cli.exe`) is a regular file with the same bytes
/// — a hard link or a copy of the same build — is the same installation, so its receipt and
/// `install_id` are the sibling's. Any other path, a missing or different sibling included,
/// is its own identity, as before.
pub(crate) fn identity_exe(canonical_exe: PathBuf) -> PathBuf {
    let (Some(stem), ext) = (
        canonical_exe.file_stem().and_then(|s| s.to_str()),
        canonical_exe.extension().and_then(|e| e.to_str()),
    ) else {
        return canonical_exe;
    };
    if stem != "moai" || !matches!(ext, None | Some("exe")) {
        return canonical_exe;
    }
    let name = match ext {
        Some(ext) => format!("{PRODUCT}.{ext}"),
        None => PRODUCT.to_string(),
    };
    let sibling = canonical_exe.with_file_name(name);
    let same = fs::symlink_metadata(&sibling).is_ok_and(|m| m.file_type().is_file())
        && matches!(
            (sha256_file(&sibling), sha256_file(&canonical_exe)),
            (Ok(a), Ok(b)) if a == b
        );
    if same {
        sibling
    } else {
        canonical_exe
    }
}

pub(crate) fn install_id(canonical_exe: &Path) -> Result<String> {
    let bytes = canonical_exe
        .to_str()
        .ok_or("non UTF-8 binary path")?
        .as_bytes();
    Ok(format!("{:x}", Sha256::digest(bytes)))
}

/// Plans the receipt the current executable would be recorded under. Computes everything
/// from the binary itself; writes nothing.
///
/// The binary is identified by the regular file its path resolves to, so a call through a
/// link (the installed layout `~/.moai-cli/path/moai-cli`, decisions L165) plans the same
/// receipt and `install_id` as the target itself (#2495). Whether `std::env::current_exe`
/// reports the link is platform dependent (macOS keeps it, Linux resolves it); the result
/// is the same either way. A link to a non-file or a dangling link is refused.
pub fn plan_receipt(
    current_exe: &Path,
    home: &Path,
    installer_owner: InstallerOwner,
) -> Result<ReceiptPlan> {
    fs::symlink_metadata(current_exe).map_err(|_| "binary metadata unavailable")?;
    let canonical_exe = fs::canonicalize(current_exe).map_err(|_| "binary path unavailable")?;
    if !canonical_exe.is_file() {
        return Err("binary path is not a regular file".into());
    }
    let canonical_exe = identity_exe(canonical_exe);
    let canonical_home = fs::canonicalize(home).map_err(|_| "home directory unavailable")?;
    let version = env!("CARGO_PKG_VERSION").to_string();
    let receipt = InstallReceipt {
        schema_version: RECEIPT_SCHEMA_VERSION,
        product: PRODUCT.into(),
        generation: GENERATION.into(),
        channel: channel_of(&version),
        binary_path: normalize_home(&canonical_home, &canonical_exe)?,
        binary_sha256: sha256_file(&canonical_exe)?,
        version,
        installer_owner,
        protocol_version: PROTOCOL_VERSION,
        created_at: rfc3339_now(),
        updated_at: rfc3339_now(),
    };
    validate_receipt(&receipt)?;
    Ok(ReceiptPlan {
        install_id: install_id(&canonical_exe)?,
        receipt,
    })
}

/// The nine identity fields a rerun compares; timestamps are deliberately excluded so a
/// rerun with identical inputs is `unchanged` (AC-04A-05).
fn identity_fields_equal(a: &InstallReceipt, b: &InstallReceipt) -> bool {
    a.schema_version == b.schema_version
        && a.product == b.product
        && a.generation == b.generation
        && a.channel == b.channel
        && a.version == b.version
        && a.binary_path == b.binary_path
        && a.binary_sha256 == b.binary_sha256
        && a.installer_owner == b.installer_owner
        && a.protocol_version == b.protocol_version
}

fn installations_dir(home: &Path) -> Result<PathBuf> {
    let dir = home.join(crate::HOME_STATE_DIR).join("installations");
    match fs::symlink_metadata(&dir) {
        Ok(m) if m.file_type().is_symlink() => {
            Err("installations directory is a symlink; refused".into())
        }
        Ok(m) if !m.is_dir() => Err("installations path is not a directory".into()),
        Ok(_) => Ok(dir),
        Err(_) => {
            fs::create_dir_all(&dir).map_err(|_| "installations directory unavailable")?;
            match fs::symlink_metadata(&dir) {
                Ok(m) if m.is_dir() => Ok(dir),
                _ => Err("installations directory unavailable".into()),
            }
        }
    }
}

fn read_bounded(path: &Path, limit: u64, what: &str) -> Result<Vec<u8>> {
    let m = fs::metadata(path).map_err(|_| format!("{what} metadata unavailable"))?;
    if !m.is_file() || m.len() > limit {
        return Err(format!("{what} is not a bounded regular file"));
    }
    fs::read(path).map_err(|_| format!("{what} read failed"))
}

fn parse_receipt_file(path: &Path) -> Result<InstallReceipt> {
    let bytes = read_bounded(path, RECEIPT_LIMIT, "installations receipt")?;
    let receipt: InstallReceipt = serde_json::from_slice(&bytes)
        .map_err(|_| "invalid installation receipt; no values were logged".to_string())?;
    validate_receipt(&receipt)?;
    Ok(receipt)
}

/// Records the planned receipt as `<home>/.moai-cli/installations/<install_id>.json`.
/// Writes only inside `~/.moai-cli/`. A rerun with identical identity fields reports
/// `unchanged` and touches nothing (AC-04A-05); different fields for the same binary
/// path are refused — updating a receipt is the explicit 04-B apply path, never an
/// automatic rewrite (REQ-04A-06).
pub fn record_receipt(home: &Path, plan: &ReceiptPlan) -> Result<Value> {
    validate_receipt(&plan.receipt)?;
    let dir = installations_dir(home)?;
    let path = dir.join(format!("{}.json", plan.install_id));
    let mut bytes = serde_json::to_vec_pretty(&plan.receipt).map_err(|_| "encode failed")?;
    bytes.push(b'\n');
    match fs::symlink_metadata(&path) {
        Ok(m) if m.file_type().is_symlink() => {
            return Err("installations receipt is a symlink; refused".into())
        }
        Ok(_) => {
            let existing = parse_receipt_file(&path)?;
            if identity_fields_equal(&existing, &plan.receipt) {
                return Ok(json!({"state":"unchanged","install_id":plan.install_id}));
            }
            return Err("conflict in installations: a different receipt exists for this binary path; nothing written".into());
        }
        Err(e) if e.kind() == std::io::ErrorKind::NotFound => (),
        Err(_) => return Err("installations receipt metadata unavailable".into()),
    }
    let mut file = OpenOptions::new()
        .write(true)
        .create_new(true)
        .open(&path)
        .map_err(|_| "exclusive create refused")?;
    file.write_all(&bytes)
        .and_then(|_| file.sync_all())
        .map_err(|_| "partial creation possible; preserve and inspect target")?;
    Ok(json!({
        "state":"created",
        "install_id":plan.install_id,
        "receipt_digest": crate::digest(&bytes),
        "host_verified":false,
    }))
}

/// Reads every `*.json` receipt under `<home>/.moai-cli/installations/`, oldest-independent
/// sorted by id. A malformed sibling is reported as `invalid`, never fails the listing.
pub fn read_receipts(home: &Path) -> Result<Vec<Value>> {
    let dir = home.join(crate::HOME_STATE_DIR).join("installations");
    let entries = match fs::read_dir(&dir) {
        Ok(e) => e,
        Err(e) if e.kind() == std::io::ErrorKind::NotFound => return Ok(Vec::new()),
        Err(_) => return Err("installations directory unavailable".into()),
    };
    let mut out = Vec::new();
    for entry in entries {
        let entry = entry.map_err(|_| "directory entry unavailable".to_string())?;
        let name = entry.file_name().to_string_lossy().into_owned();
        if !name.ends_with(".json") {
            continue;
        }
        let install_id = name.trim_end_matches(".json").to_string();
        let path = entry.path();
        let item = match fs::symlink_metadata(&path) {
            Ok(m) if m.file_type().is_symlink() => {
                json!({"install_id":install_id,"status":"invalid","reason":"receipt is a symlink; refused"})
            }
            Ok(_) => match parse_receipt_file(&path) {
                Ok(receipt) => json!({"install_id":install_id,"status":"ok","receipt":receipt}),
                Err(reason) => json!({"install_id":install_id,"status":"invalid","reason":reason}),
            },
            Err(_) => {
                json!({"install_id":install_id,"status":"invalid","reason":"receipt metadata unavailable"})
            }
        };
        out.push((install_id, item));
    }
    out.sort_by(|a, b| a.0.cmp(&b.0));
    Ok(out.into_iter().map(|(_, v)| v).collect())
}

/// REQ-04A-03: a receipt is this installation only when BOTH the path (its `~` token
/// expanded) and the digest match the current executable. A path match with a digest
/// mismatch is `stale_receipt` (REQ-04A-06); anything else is `foreign`.
fn match_receipt(
    receipt: &InstallReceipt,
    canonical_home: &Path,
    canonical_exe: &Path,
    exe_sha256: &str,
) -> &'static str {
    let expanded = expand_home(canonical_home, &receipt.binary_path);
    let Ok(resolved) = fs::canonicalize(&expanded) else {
        return "foreign";
    };
    if resolved != canonical_exe {
        return "foreign";
    }
    if receipt.binary_sha256 == exe_sha256 {
        "self"
    } else {
        "stale_receipt"
    }
}

/// `--version` output of one candidate (REQ-04A-02, #2791). Spawns the binary with the
/// single argument `--version`, stdin closed, stdout capped, killed after
/// `PROBE_TIMEOUT`; returns `None` when nothing was printed, else the first line (at most
/// `PROBE_LINE_CAP` characters, trimmed). The child gets a cleared environment — only
/// `PROBE_PATH`, plus `SYSTEMROOT` on Windows, which process start needs — and the
/// system temporary folder as its working folder, so no token or key in this process's
/// environment reaches it and it does not start inside a project (#2800 audit B1).
///
/// The limit holds for the whole probe, not only the direct child (#2791 pre-audit
/// 6865400825 B1): a shell wrapper's own children may keep stdout open after the wrapper is
/// killed. On unix the child leads a process group of its own and the whole group is killed
/// when the time is up; on every platform the output is awaited only for the time left, so a
/// pipe some leftover process still holds never keeps this call waiting.
// @MX:WARN: [AUTO] executes whatever binary a PATH entry or a Go install place holds under the expected name
// @MX:REASON: REQ-04A-02 and #2791 AC-2 tell Rust from Go by the `--version` line; a hostile PATH entry runs at inspect time, without this process's environment; project-tree binaries are never run
pub(crate) fn probe_version(path: &Path) -> Option<String> {
    let mut cmd = std::process::Command::new(path);
    cmd.env_clear()
        .env("PATH", PROBE_PATH)
        .current_dir(std::env::temp_dir());
    if let Some(root) = std::env::var_os("SYSTEMROOT") {
        cmd.env("SYSTEMROOT", root);
    }
    #[cfg(unix)]
    {
        use std::os::unix::process::CommandExt as _;
        cmd.process_group(0);
    }
    let Ok(mut child) = cmd
        .arg("--version")
        .stdin(std::process::Stdio::null())
        .stdout(std::process::Stdio::piped())
        .stderr(std::process::Stdio::null())
        .spawn()
    else {
        return None;
    };
    let Some(stdout) = child.stdout.take() else {
        let _ = child.kill();
        return None;
    };
    let (tx, rx) = std::sync::mpsc::channel();
    std::thread::spawn(move || {
        let mut kept = Vec::new();
        let mut total = 0usize;
        let mut chunk = [0u8; 512];
        let mut handle = stdout;
        while total <= PROBE_STDOUT_CAP {
            match handle.read(&mut chunk) {
                Ok(0) => break,
                Ok(n) => {
                    if kept.len() < PROBE_STDOUT_CAP {
                        let room = PROBE_STDOUT_CAP - kept.len();
                        kept.extend_from_slice(&chunk[..n.min(room)]);
                    }
                    total += n;
                }
                Err(_) => break,
            }
        }
        let _ = tx.send((total > 0).then(|| first_line(&kept)));
    });
    let deadline = Instant::now() + PROBE_TIMEOUT;
    let exited = loop {
        match child.try_wait() {
            Ok(Some(_)) => break true,
            Ok(None) if Instant::now() >= deadline => break false,
            Ok(None) => std::thread::sleep(Duration::from_millis(50)),
            Err(_) => break false,
        }
    };
    // The output is complete once every holder of the pipe closed it; wait only the time left.
    let output = rx.recv_timeout(deadline.saturating_duration_since(Instant::now()));
    if !exited || output.is_err() {
        kill_probe(&mut child);
    }
    match output {
        Ok(line) => line,
        // Killed: what was printed before the limit, if the reader still finishes promptly.
        Err(_) => rx.recv_timeout(PROBE_GRACE).unwrap_or(None),
    }
}

/// Ends a probe that ran out of time, with the process group it leads on unix, and reaps it.
fn kill_probe(child: &mut std::process::Child) {
    #[cfg(unix)]
    if let Ok(pid) = i32::try_from(child.id()) {
        // SAFETY: `kill` takes plain integers; `-pid` names only the group this child leads
        // (`process_group(0)` at spawn), never this process's own group.
        unsafe {
            libc::kill(-pid, libc::SIGKILL);
        }
    }
    let _ = child.kill();
    let _ = child.wait();
}

/// The first line of `--version` output, trimmed, at most `PROBE_LINE_CAP` characters.
fn first_line(bytes: &[u8]) -> String {
    let text = String::from_utf8_lossy(bytes);
    let line = text.lines().next().unwrap_or("").trim();
    line.chars().take(PROBE_LINE_CAP).collect()
}

/// Which generation a `--version` first line names (#2791 AC-2): the Rust CLI prints
/// `moai-cli <version>` (R1 adds ` (rust)`), the Go `moai` prints `moai-adk <version>`
/// (moai-adk-go `internal/cli/root.go` version template; older Go builds `moai <version>`).
/// Anything else, or no output at all, is `unknown` — never guessed.
pub(crate) fn generation_of(first: Option<&str>) -> &'static str {
    match first {
        Some(l) if l.starts_with("moai-cli ") => "rust",
        Some(l) if l.starts_with("moai-adk ") || l.starts_with("moai ") => "go",
        _ => "unknown",
    }
}

fn is_executable(path: &Path) -> bool {
    let Ok(meta) = fs::metadata(path) else {
        return false;
    };
    if !meta.is_file() {
        return false;
    }
    #[cfg(unix)]
    {
        use std::os::unix::fs::PermissionsExt;
        meta.permissions().mode() & 0o111 != 0
    }
    #[cfg(not(unix))]
    {
        true
    }
}

/// First PATH entry holding an executable `<name>` (shell order); on Windows the
/// `.exe` spelling is tried after the bare name.
pub(crate) fn resolve_on_path(path_env: &str, name: &str) -> Option<PathBuf> {
    for dir in std::env::split_paths(path_env) {
        if dir.as_os_str().is_empty() {
            continue;
        }
        if let Some(found) = executable_in(&dir, name) {
            return Some(found);
        }
    }
    None
}

/// The executable `<dir>/<name>` (on Windows `<name>.exe` after the bare name), if any.
fn executable_in(dir: &Path, name: &str) -> Option<PathBuf> {
    let candidate = dir.join(name);
    if is_executable(&candidate) {
        return Some(candidate);
    }
    #[cfg(windows)]
    {
        let mut with_ext = candidate.clone();
        with_ext.set_extension("exe");
        if is_executable(&with_ext) {
            return Some(with_ext);
        }
    }
    None
}

/// Where a Go-generation `moai` may sit besides PATH (#2791 AC-2, design #2725 C2·C4).
pub struct CandidatePlaces<'a> {
    /// `$CLAUDE_PROJECT_DIR`; its `.moai/bin` holds a project-local Go `moai`.
    pub project_dir: Option<&'a Path>,
    /// `$GOBIN`.
    pub gobin: Option<&'a Path>,
}

/// Every `moai` on PATH and in the Go install places, told apart by `--version` (#2791
/// AC-2). Report only: nothing is written, linked or reordered. Candidates are listed in
/// PATH order first, then the fixed places; one file reached twice (same canonical path)
/// is one item with every source named. Beyond `CANDIDATE_LIMIT` items the rest are
/// counted in `more`, not probed.
fn moai_candidates(home: &Path, path_env: &str, places: &CandidatePlaces) -> Value {
    let mut found: Vec<(PathBuf, String, Option<usize>)> = Vec::new();
    for (i, dir) in std::env::split_paths(path_env).enumerate() {
        if dir.as_os_str().is_empty() {
            continue;
        }
        if let Some(p) = executable_in(&dir, "moai") {
            found.push((p, "path".into(), Some(i)));
        }
    }
    let fixed = [
        (
            "claude_project_dir",
            places.project_dir.map(|d| d.join(".moai").join("bin")),
        ),
        ("gobin", places.gobin.map(Path::to_path_buf)),
        ("go_bin", Some(home.join("go").join("bin"))),
        ("local_bin", Some(home.join(".local").join("bin"))),
    ];
    for (source, dir) in fixed {
        if let Some(p) = dir.as_deref().and_then(|d| executable_in(d, "moai")) {
            found.push((p, source.into(), None));
        }
    }
    let first_on_path = resolve_on_path(path_env, "moai").and_then(|p| fs::canonicalize(p).ok());
    let mut items: Vec<(PathBuf, Value)> = Vec::new();
    let mut more = 0usize;
    for (path, source, order) in found {
        let canonical = fs::canonicalize(&path).unwrap_or_else(|_| path.clone());
        if let Some((_, item)) = items.iter_mut().find(|(c, _)| *c == canonical) {
            let sources = item["sources"].as_array_mut().expect("sources array");
            if !sources.contains(&json!(source)) {
                sources.push(json!(source));
            }
            continue;
        }
        if items.len() >= CANDIDATE_LIMIT {
            more += 1;
            continue;
        }
        // A `moai` inside the project tree is repository content, not a place the user
        // chose: it is reported, never run (#2800 audit B1). Reached through PATH first, the
        // user's PATH chose it and it is probed like any PATH entry.
        let probed = source != "claude_project_dir";
        let line = if probed { probe_version(&path) } else { None };
        let generation = if probed {
            generation_of(line.as_deref())
        } else {
            "not_probed"
        };
        let item = json!({
            "path": path.to_string_lossy(),
            "sources": [source],
            "path_order": order,
            "first_on_path": first_on_path.as_ref() == Some(&canonical),
            "sha256": sha256_file(&path).ok(),
            "probed": probed,
            "version_line": line,
            "generation": generation,
        });
        items.push((canonical, item));
    }
    let first = items
        .iter()
        .find(|(_, i)| i["first_on_path"] == json!(true))
        .map(|(_, i)| i["generation"].clone())
        .unwrap_or(Value::Null);
    json!({
        "items": items.into_iter().map(|(_, i)| i).collect::<Vec<_>>(),
        "more": more,
        "first_on_path_generation": first,
    })
}

fn path_resolution(found: Option<PathBuf>) -> Value {
    match found {
        None => json!({"found":false,"path":null,"sha256":null,"version_output":null}),
        Some(path) => {
            let sha256 = sha256_file(&path).ok();
            let version_output = probe_version(&path).is_some();
            json!({
                "found":true,
                "path":path.to_string_lossy(),
                "sha256":sha256,
                "version_output":version_output,
            })
        }
    }
}

fn exists_dir(path: &Path) -> bool {
    fs::symlink_metadata(path)
        .map(|m| m.is_dir())
        .unwrap_or(false)
}

fn text_or_null(path: &Path) -> Value {
    match path.to_str() {
        Some(s) => json!(s),
        None => Value::Null,
    }
}

/// Read-only installation inventory (REQ-04A-02): PATH resolution of `moai` and
/// `moai-cli`, the receipts under `~/.moai-cli/installations/`, and existence of the Go
/// generation's paths. Creates and modifies nothing; the Go paths are only probed with
/// `symlink_metadata` existence checks (REQ-04A-04). Self-identification follows the
/// receipts, never the first PATH `moai` (REQ-04A-03).
pub fn inspect(home: &Path, path_env: &str, places: &CandidatePlaces) -> Result<Value> {
    let canonical_home = fs::canonicalize(home).map_err(|_| "home directory unavailable")?;
    let exe = std::env::current_exe().map_err(|_| "current executable unavailable".to_string())?;
    let canonical_exe =
        fs::canonicalize(&exe).map_err(|_| "current executable unavailable".to_string())?;
    let exe_sha256 = sha256_file(&canonical_exe)?;
    // A `moai` that is the same build as its sibling `moai-cli` matches that receipt (B2).
    let identity = identity_exe(canonical_exe.clone());

    let moai = resolve_on_path(path_env, "moai");
    let moai_cli = resolve_on_path(path_env, "moai-cli");

    let mut self_state = "unrecorded";
    let mut installations = Vec::new();
    for mut entry in read_receipts(&canonical_home)? {
        if entry["status"] == "ok" {
            let receipt: InstallReceipt =
                serde_json::from_value(entry["receipt"].clone()).map_err(|_| "encode failed")?;
            let state = match_receipt(&receipt, &canonical_home, &identity, &exe_sha256);
            if state == "self" {
                self_state = "self";
            } else if state == "stale_receipt" && self_state == "unrecorded" {
                self_state = "stale_receipt";
            }
            entry["match_state"] = json!(state);
        }
        installations.push(entry);
    }

    let cwd = std::env::current_dir().map_err(|_| "project directory unavailable")?;
    Ok(json!({
        "install_id_derivation": INSTALL_ID_DERIVATION,
        "home_state_root": text_or_null(&canonical_home.join(crate::HOME_STATE_DIR)),
        "path_resolution": {
            "moai": path_resolution(moai),
            "moai_cli": path_resolution(moai_cli),
        },
        "moai_candidates": moai_candidates(&canonical_home, path_env, places),
        "installations": installations,
        "self": {
            "current_exe": normalize_home(&canonical_home, &canonical_exe)?,
            "binary_sha256": exe_sha256,
            "state": self_state,
            // The path the platform reported before resolution, and whether it was a link
            // (macOS reports a link it was started through; Linux already resolves it).
            "invoked_exe": normalize_home(&canonical_home, &exe)?,
            "invoked_via_link": fs::symlink_metadata(&exe).is_ok_and(|m| m.file_type().is_symlink()),
        },
        "go_generation": {
            "home_moai_exists": exists_dir(&canonical_home.join(".moai")),
            "project_moai_exists": exists_dir(&cwd.join(".moai")),
            "project_root": text_or_null(&cwd),
        },
        "writes": false,
    }))
}

#[cfg(test)]
mod identity_tests {
    use super::identity_exe;
    use std::fs;
    use std::path::{Path, PathBuf};

    fn write(dir: &Path, name: &str, bytes: &[u8]) -> PathBuf {
        let p = dir.join(name);
        fs::write(&p, bytes).unwrap();
        p
    }

    /// #2790 R1 (B2): `moai` beside the same build's `moai-cli` is identified as `moai-cli`;
    /// a different or missing sibling, or any other name, keeps its own path.
    #[test]
    fn moai_takes_the_canonical_name_only_beside_the_same_build() {
        let d = tempfile::TempDir::new().unwrap();
        let dir = d.path();
        let moai = write(dir, "moai", b"build-1");
        assert_eq!(identity_exe(moai.clone()), moai, "no sibling");
        let cli = write(dir, "moai-cli", b"build-1");
        assert_eq!(identity_exe(moai.clone()), cli, "same bytes");
        assert_eq!(
            identity_exe(cli.clone()),
            cli,
            "the canonical name is itself"
        );
        write(dir, "moai-cli", b"build-2");
        assert_eq!(identity_exe(moai.clone()), moai, "a different build");

        let exe = write(dir, "moai.exe", b"w");
        let cli_exe = write(dir, "moai-cli.exe", b"w");
        assert_eq!(identity_exe(exe), cli_exe, "Windows names");
        for other in ["moaix", "moai.old", "xmoai"] {
            let p = write(dir, other, b"build-2");
            assert_eq!(identity_exe(p.clone()), p, "{other}");
        }
    }

    /// A sibling that is a symlink is not a file of this installation, even to the same bytes.
    #[cfg(unix)]
    #[test]
    fn a_linked_sibling_is_not_the_same_installation() {
        let d = tempfile::TempDir::new().unwrap();
        let dir = d.path();
        let moai = write(dir, "moai", b"build-1");
        let target = write(dir, "elsewhere", b"build-1");
        std::os::unix::fs::symlink(&target, dir.join("moai-cli")).unwrap();
        assert_eq!(identity_exe(moai.clone()), moai);
    }
}
