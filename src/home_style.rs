//! The MoAI-Style output style in Claude Code's user layer (#2875, #2871 I3).
//!
//! The style text is one package file, `references/home/output-style.md`: Claude Code reads it
//! as the output style file `~/.claude/output-styles/MoAI-Style.md` (frontmatter added) and
//! selects it by default through `outputStyle` in `~/.claude/settings.json`. The style was named
//! MoAI-Easy before SPEC-20261127-output-style-moai-style: an installed `MoAI-Easy.md` that is
//! MoAI's own is retired (its bytes kept in the receipt) and an `outputStyle` of exactly
//! `MoAI-Easy` is replaced as text, in the same apply; see [`OLD_STYLE_NAME`]. Hosts without a
//! user-level file this module can write are reported in `not_supported`, as for the routing
//! section. A receipt entry that names a host this module does not write is skipped.
//!
//! `plan` writes nothing. Every target is three-way: the style file is ours while it holds the
//! bytes the record names (`<state folder>/home-style/<host>.json`; a config folder other than
//! the HOME's default has its own record, see [`record_name`]), the current text (adopted into the
//! record) or a text MoAI shipped earlier ([`EARLIER_RENDERS`]); anything else is a `conflict`,
//! never overwritten. `outputStyle` is set only when the key is absent — another
//! value is the user's choice and a `conflict`. The same settings target also plans Claude Code's
//! `language` key (the saved conversation language's lower-case English name) when that key is
//! absent, apart from the style file's state: both keys go in one file replacement with one
//! receipt entry (SPEC-20261120-reply-language-rule). A symlinked folder or file blocks the host.
//! `apply` re-plans, refuses any other digest and keeps the before bytes in a receipt; `revert`
//! restores a file only while it still holds the bytes the apply wrote, and removes each key
//! the apply inserted only while it still holds the value the apply wrote.

use crate::apply::{publish_exclusive, publish_replace, read_opt};
use crate::retired_hosts::retired;
use serde::{Deserialize, Serialize};
use serde_json::{json, Map, Value};
use std::{
    fs,
    path::{Path, PathBuf},
};

type Result<T> = std::result::Result<T, String>;

const SCHEMA_VERSION: u32 = 1;
const STATE_DIR: &str = "home-style";
/// The style source inside the skill package.
pub const SOURCE: &str = "references/home/output-style.md";
/// The style name Claude Code selects (`outputStyle` is case-sensitive).
pub const STYLE_NAME: &str = "MoAI-Style";
/// The style name before SPEC-20261127-output-style-moai-style: the source of the move, never
/// written. Its file is [`OLD_STYLE_FILE`].
pub const OLD_STYLE_NAME: &str = "MoAI-Easy";
/// Hosts written.
pub const HOSTS: [&str; 1] = ["claude"];
/// The targets, each below its host folder ([`crate::host_home`]: `~/.claude` or
/// `CLAUDE_CONFIG_DIR`; card C-79).
const CLAUDE_STYLE_FILE: &str = "output-styles/MoAI-Style.md";
/// The earlier style file: a target only to retire, and the path of a receipt entry that names
/// none (a receipt written before the rename).
const OLD_STYLE_FILE: &str = "output-styles/MoAI-Easy.md";
/// The words of a setting detail that says the old value could not be swapped as text; the old
/// file is then not retired either.
const SWAP_REFUSED: &str = "could not be swapped as text";
const CLAUDE_SETTINGS: &str = "settings.json";
const SOURCE_LIMIT: usize = 64 * 1024;
const SETTINGS_LIMIT: usize = 256 * 1024;
/// The frontmatter description. No double quote inside: [`render_style_file`] wraps it in two.
const DESCRIPTION: &str = "Plain words, four steps (understand, plan, do, check), evidence before done, MoAI 🗿 banners and a progress board.";
/// The sha256 (64 hex digits) of every style file text moai-cli rendered before the current one,
/// oldest first: the first text with an unquoted description and a star banner, the YAML repair
/// of its description, the text with the `MoAI 🗿` banner, and the text of the picked sections
/// before its progress board pointed to the flow rule, and the text whose board pointed to the
/// flow rule and listed seven icons, before the board became the one progress display with
/// eleven flow icons, and that text before its board took the one icon set of the status line
/// and the lane reports, and that text before its board took the card-flow icons of
/// `statusline::icons` (SPEC-20261124-card-flow-icons), and that text, named MoAI-Easy, before
/// the rename and the wording principle (SPEC-20261127-output-style-moai-style). A style file
/// equal to one of them is a text MoAI shipped, so it is updated and not left as a conflict, and
/// an old `MoAI-Easy.md` equal to one of them is MoAI's own and is retired. When the template or
/// the frontmatter changes, the render this module made until then is appended here (the test
/// `the_earlier_renders_are_pinned_and_never_hold_the_current_one` fails until it is).
pub const EARLIER_RENDERS: [&str; 8] = [
    "93e53607dd194da86abfbe7ea993303fff1d17af1a1a1ae6868788938ed785d2",
    "d4e1851a65eb0ed21d16de888925dc6ae8a7bcebcae45973e1b8cce884d75e67",
    "690c534220f9bb1614e27d608e56306da973e2fe7b53e2bc1a4d42d7d2b9c65e",
    "a31d966b0486f501839121832eb6a99d3fc718996b22a901e9e17884a0ab369e",
    "80e189d27dd7be365724bbba2f38f1eddb3c0e96801da4caf82535ce80181a65",
    "5f6b5fd1d93f3b0a4553ba74df113bba99bb018ca819120e1a25926fe242d006",
    "8468a4c95c48ab62975d13bac3041e60a393b2ba368f36c09544ef96df810847",
    "0c66f9d0662a8a2d6bb7aa73f6e2c623c33d43d27fe800abc51a230ae35389f3",
];

/// One write target of one host.
#[derive(Clone, Debug, Serialize, Deserialize, PartialEq)]
#[serde(deny_unknown_fields)]
pub struct TargetPlan {
    pub host: String,
    /// style_file | setting | old_style_file
    pub kind: String,
    pub path: String,
    /// create | merge | append | set | replace | retire | unchanged | conflict | skipped | blocked
    /// (`replace`: the setting target swaps the exact old `outputStyle` value; `retire`: the
    /// old style file is removed after its bytes are kept in the receipt)
    pub action: String,
    /// The sha256 of the file's bytes the plan was made from; none when absent.
    pub expected_sha256: Option<String>,
    /// What this module owned before (the style file's sha256).
    pub baseline_sha256: Option<String>,
    pub detail: Option<String>,
    /// The link in the config folder that the file at `path` is reached through (the link stays as
    /// it is); none for a file that is not behind a link. Left out of the plan then, so a plan
    /// without a link keeps its digest.
    #[serde(default, skip_serializing_if = "Option::is_none")]
    pub link: Option<String>,
    /// The `settings.json` keys an apply writes (the `setting` target only), in file order. Left
    /// out of the plan while empty, so a plan that writes no key keeps its digest.
    #[serde(default, skip_serializing_if = "Vec::is_empty")]
    pub keys: Vec<SettingKey>,
}

/// One key the settings target would add, with the value it would hold.
#[derive(Clone, Debug, Serialize, Deserialize, PartialEq)]
#[serde(deny_unknown_fields)]
pub struct SettingKey {
    pub key: String,
    pub value: Value,
    /// The value the key holds now and the apply swaps for `value` as text; none for a key
    /// that is added. Left out of the plan then, so a plan that adds keys keeps its digest.
    #[serde(default, skip_serializing_if = "Option::is_none")]
    pub replaces: Option<Value>,
}

/// A key an apply added or swapped: its value and the exact text put into the file, so a revert
/// can take back this key alone.
#[derive(Clone, Debug, Serialize, Deserialize, PartialEq)]
#[serde(deny_unknown_fields)]
pub struct InsertedKey {
    pub key: String,
    pub value: Value,
    pub text: String,
    /// The value and the exact text the swap took away; a revert puts them back. None for an
    /// added key (a revert removes `text`).
    #[serde(default, skip_serializing_if = "Option::is_none")]
    pub replaced: Option<ReplacedText>,
}

/// What a swapped key held before: its value and the exact text of `"key": value` in the file.
#[derive(Clone, Debug, Serialize, Deserialize, PartialEq)]
#[serde(deny_unknown_fields)]
pub struct ReplacedText {
    pub value: Value,
    pub text: String,
}

#[derive(Clone, Debug, Serialize, Deserialize, PartialEq)]
#[serde(deny_unknown_fields)]
pub struct Unsupported {
    pub host: String,
    pub reason: String,
}

#[derive(Clone, Debug, Serialize, Deserialize, PartialEq)]
#[serde(deny_unknown_fields)]
pub struct Plan {
    pub schema_version: u32,
    pub home: String,
    pub package_digest: String,
    pub source_sha256: String,
    pub targets: Vec<TargetPlan>,
    pub not_supported: Vec<Unsupported>,
    pub plan_digest: String,
}

/// The style file we last wrote for one host.
#[derive(Clone, Debug, Default, Serialize, Deserialize, PartialEq)]
#[serde(deny_unknown_fields)]
struct HostState {
    schema_version: u32,
    owned_sha256: Option<String>,
}

#[derive(Clone, Debug, Serialize, Deserialize, PartialEq)]
#[serde(deny_unknown_fields)]
pub struct ReceiptEntry {
    pub host: String,
    pub kind: String,
    /// The target's path below the config folder. A receipt written before the rename names
    /// none: its `style_file` entry is the earlier file (`output-styles/MoAI-Easy.md`).
    #[serde(default, skip_serializing_if = "Option::is_none")]
    pub path: Option<String>,
    pub action: String,
    pub before_sha256: Option<String>,
    pub after_sha256: String,
    /// What this module owned before the apply; a revert restores it.
    pub baseline_before: Option<String>,
    /// The exact text the apply put into `settings.json` (#2875 audit B1): a revert of a file
    /// changed since removes only this text, never re-encoding the user's values.
    #[serde(default, skip_serializing_if = "Option::is_none")]
    pub inserted: Option<String>,
    /// Each key the apply added with its own text, so a revert takes back the keys that still
    /// hold what was written and leaves a key changed since. A receipt written before the
    /// `language` key has none: it added `outputStyle` alone, and `inserted` says how.
    #[serde(default, skip_serializing_if = "Vec::is_empty")]
    pub inserted_keys: Vec<InsertedKey>,
}

#[derive(Clone, Debug, Serialize, Deserialize, PartialEq)]
#[serde(deny_unknown_fields)]
pub struct Receipt {
    pub schema_version: u32,
    pub id: String,
    pub plan_digest: String,
    pub permission_reference: String,
    /// pending | committed | reverted
    pub status: String,
    pub entries: Vec<ReceiptEntry>,
    /// The Claude config folder the apply wrote in (D-5 of
    /// SPEC-20261122-output-style-install-checks): a revert runs only where the environment
    /// picks this folder again. A receipt written before it has none and reverts in the
    /// default folder only.
    #[serde(default, skip_serializing_if = "Option::is_none")]
    pub config_folder: Option<String>,
}

/// The Claude config folder the environment picks for `home` ([`crate::host_home::resolve`]).
fn claude_folder(home: &Path) -> Result<PathBuf> {
    crate::host_home::resolve(crate::host_home::Host::Claude, home)
}

/// Two spellings of one folder: equal real paths, or equal text when either cannot be resolved.
fn same_folder(a: &Path, b: &Path) -> bool {
    match (fs::canonicalize(a), fs::canonicalize(b)) {
        (Ok(x), Ok(y)) => x == y,
        _ => a == b,
    }
}

/// D-5: a revert changes nothing unless the environment picks the folder the receipt was
/// applied in; a receipt without a folder belongs to the default folder.
fn receipt_folder_matches(home: &Path, receipt: &Receipt) -> Result<()> {
    let here = claude_folder(home)?;
    let default = home.join(crate::host_home::Host::Claude.default_dir());
    let theirs = receipt
        .config_folder
        .as_ref()
        .map_or_else(|| default.clone(), PathBuf::from);
    if same_folder(&here, &theirs) {
        return Ok(());
    }
    Err(match &receipt.config_folder {
        Some(f) => format!(
            "this style receipt belongs to the Claude config folder {f}, not {}; run the revert with CLAUDE_CONFIG_DIR={f}; nothing was changed",
            here.display()
        ),
        None => format!(
            "this style receipt names no config folder, so it reverts only in the default folder {}, not {}; nothing was changed",
            default.display(),
            here.display()
        ),
    })
}

fn sha(bytes: &[u8]) -> String {
    crate::digest(bytes)
}

fn rel_path(host: &str, kind: &str) -> &'static str {
    match (host, kind) {
        ("claude", "style_file") => CLAUDE_STYLE_FILE,
        ("claude", "setting") => CLAUDE_SETTINGS,
        ("claude", "old_style_file") => OLD_STYLE_FILE,
        _ => unreachable!("unknown style target {host}/{kind}"),
    }
}

/// The path a receipt entry names, as one of the paths this module writes: an entry that names
/// none is an entry of the earlier format (its style file is the old file); a path that is none of
/// the three is refused, so a receipt never makes a revert write anywhere else.
fn entry_rel(e: &ReceiptEntry) -> std::result::Result<&'static str, String> {
    const UNKNOWN: &str = "style receipt names an unknown target";
    if e.host != "claude" {
        return Err(UNKNOWN.into());
    }
    // The paths each kind may name; the first is the one a receipt without a path means.
    let allowed: &[&'static str] = match e.kind.as_str() {
        // An earlier receipt's style file entry is the file the earlier build wrote.
        "style_file" => &[OLD_STYLE_FILE, CLAUDE_STYLE_FILE],
        "setting" => &[CLAUDE_SETTINGS],
        "old_style_file" => &[OLD_STYLE_FILE],
        _ => return Err(UNKNOWN.into()),
    };
    match e.path.as_deref() {
        None => Ok(allowed[0]),
        Some(p) => allowed
            .iter()
            .copied()
            .find(|known| *known == p)
            .ok_or_else(|| UNKNOWN.to_string()),
    }
}

/// The target's host folder and its path below that folder.
/// A relocated `output-styles` link to the default folder is followed there (decision
/// "기본 폴더 링크만 허용"); the returned folder is where the link checks start.
fn place(home: &Path, host: &str, kind: &str) -> Result<(PathBuf, &'static str)> {
    let rel = rel_path(host, kind);
    Ok((place_rel(home, host, rel)?, rel))
}

fn place_rel(home: &Path, host: &str, rel: &str) -> Result<PathBuf> {
    let h = crate::host_home::Host::from_name(host)
        .unwrap_or_else(|| unreachable!("unknown style host {host}"));
    Ok(crate::host_home::place(h, home, rel)?.base)
}

/// Where one target really is.
struct Spot {
    /// The file that is read and written.
    path: PathBuf,
    /// The link in the config folder the file is reached through, when it is behind one.
    link: Option<PathBuf>,
    /// The file lies in the host's default folder (a followed link of either kind, or the config
    /// folder is the default folder).
    shared: bool,
}

/// The host's default folder below `home`.
fn default_folder(home: &Path, host: &str) -> PathBuf {
    let h = crate::host_home::Host::from_name(host)
        .unwrap_or_else(|| unreachable!("unknown style host {host}"));
    home.join(h.default_dir())
}

/// The target's real place: its host folder and every folder down to the file's parent found
/// plain or absent, and the file itself either a plain file, absent, or (in a config folder other
/// than the default folder) a link straight to a regular file inside the default folder.
fn locate(home: &Path, host: &str, kind: &str) -> Result<Spot> {
    locate_labeled(home, host, kind).map_err(|(_, why)| why)
}

/// [`locate`], with the word that says what was refused: `folder_not_plain` for a folder on the
/// way, `file_not_plain` for the file itself (a link that is not followed).
fn locate_labeled(
    home: &Path,
    host: &str,
    kind: &str,
) -> std::result::Result<Spot, (&'static str, String)> {
    locate_rel_labeled(home, host, rel_path(host, kind))
}

/// [`locate_labeled`] for the path `rel` below the host folder (a receipt entry names its own).
fn locate_rel_labeled(
    home: &Path,
    host: &str,
    rel: &str,
) -> std::result::Result<Spot, (&'static str, String)> {
    let folder = |why: String| ("folder_not_plain", why);
    let root = place_rel(home, host, rel).map_err(folder)?;
    folders_plain(&root, rel).map_err(folder)?;
    let own = root.join(rel);
    let default = default_folder(home, host);
    let link = linked_file(&root, &own, &default).map_err(|why| ("file_not_plain", why))?;
    let shared = link.is_some() || same_folder(&root, &default);
    Ok(match link {
        Some(real) => Spot {
            path: real,
            link: Some(own),
            shared,
        },
        None => Spot {
            path: own,
            link: None,
            shared,
        },
    })
}

/// The next step for a link that is refused.
fn link_next(default: &Path) -> String {
    format!(
        "next: replace the link with a real file, or link it straight to a regular file inside {}",
        default.display()
    )
}

/// The real file behind `own` when `own` is a link of a config folder to a regular file inside
/// the default folder; `None` when `own` is no link. Every other link is refused: one in the
/// default folder itself, one that points at another link, at nothing, at a folder or at a place
/// outside the default folder. The link is never changed.
fn linked_file(root: &Path, own: &Path, default: &Path) -> Result<Option<PathBuf>> {
    if !fs::symlink_metadata(own).is_ok_and(|m| m.file_type().is_symlink()) {
        return Ok(None);
    }
    let refuse = |why: &str| {
        format!(
            "{} is a symlink; refused ({why}); {}",
            own.display(),
            link_next(default)
        )
    };
    if same_folder(root, default) {
        return Err(refuse(
            "a link inside the default folder itself is not followed",
        ));
    }
    plain_dir_or_absent(default).map_err(|e| refuse(&e))?;
    let target = fs::read_link(own).map_err(|_| refuse("the link cannot be read"))?;
    let dir = own.parent().ok_or_else(|| refuse("it has no folder"))?;
    let at = if target.is_absolute() {
        target
    } else {
        dir.join(target)
    };
    match fs::symlink_metadata(&at) {
        Err(_) => return Err(refuse("it points at nothing")),
        Ok(m) if m.file_type().is_symlink() => return Err(refuse("it points at another link")),
        Ok(m) if !m.is_file() => return Err(refuse("it does not point at a regular file")),
        Ok(_) => {}
    }
    let name = at
        .file_name()
        .ok_or_else(|| refuse("it has no file name"))?;
    let parent = at.parent().ok_or_else(|| refuse("it has no folder"))?;
    let (parent, default) = match (fs::canonicalize(parent), fs::canonicalize(default)) {
        (Ok(p), Ok(d)) => (p, d),
        _ => return Err(refuse("the default folder is not there")),
    };
    if !parent.starts_with(&default) {
        return Err(refuse(&format!("it points outside {}", default.display())));
    }
    Ok(Some(parent.join(name)))
}

/// The target file, after its host folder and every folder down to the file's parent were
/// found plain or absent; its bytes, `None` when absent.
fn read_target(home: &Path, host: &str, kind: &str) -> Result<(Spot, Option<Vec<u8>>)> {
    read_target_rel(home, host, rel_path(host, kind))
}

/// [`read_target`] for the path `rel` below the host folder.
fn read_target_rel(home: &Path, host: &str, rel: &str) -> Result<(Spot, Option<Vec<u8>>)> {
    let spot = locate_rel_labeled(home, host, rel).map_err(|(_, why)| why)?;
    let cur = current_file(&spot.path)?;
    Ok((spot, cur))
}

fn state_root(home: &Path) -> PathBuf {
    crate::home_root::path(home).join(STATE_DIR)
}

/// A path that must be a real folder (or absent): never a symlink, never a file.
fn plain_dir_or_absent(path: &Path) -> Result<()> {
    match fs::symlink_metadata(path) {
        Ok(m) if m.file_type().is_symlink() => Err(format!(
            "{} is a symlink; refused; next: replace the link with a real folder",
            path.display()
        )),
        Ok(m) if !m.is_dir() => Err(format!(
            "{} is not a folder; next: move that file away so a folder can be used",
            path.display()
        )),
        Ok(_) => Ok(()),
        Err(e) if e.kind() == std::io::ErrorKind::NotFound => Ok(()),
        Err(_) => Err(format!(
            "{} is unreadable; next: fix its permissions",
            path.display()
        )),
    }
}

/// The host folder and every folder below it down to the target's parent must be plain or
/// absent.
fn folders_plain(root: &Path, rel: &str) -> Result<()> {
    plain_dir_or_absent(root)?;
    let mut dir = root.to_path_buf();
    let parts: Vec<&str> = rel.split('/').collect();
    for part in &parts[..parts.len() - 1] {
        dir.push(part);
        plain_dir_or_absent(&dir)?;
    }
    Ok(())
}

/// The current bytes of a target file: `Ok(None)` when absent; a symlink or a folder in its
/// place is an error (the target is blocked, never followed).
fn current_file(path: &Path) -> Result<Option<Vec<u8>>> {
    match fs::symlink_metadata(path) {
        Ok(m) if m.file_type().is_symlink() => Err(format!(
            "{} is a symlink; refused; next: replace the link with a real file",
            path.display()
        )),
        Ok(m) if !m.is_file() => Err(format!(
            "{} is not a regular file; next: replace it with a regular file",
            path.display()
        )),
        Ok(_) => read_opt(path),
        Err(e) if e.kind() == std::io::ErrorKind::NotFound => Ok(None),
        Err(_) => Err(format!(
            "{} is unreadable; next: fix its permissions",
            path.display()
        )),
    }
}

/// The name of the owned-text record of `host`'s folder in the state folder (D-1 of
/// SPEC-20261122-output-style-install-checks). The HOME's default folder keeps `<host>.json`
/// where it is today. Another config folder (`CLAUDE_CONFIG_DIR`) gets `<host>-<12 hex digits of
/// the sha256 of its path>.json`, and only on a HOME whose state root is this generation's: the
/// record is never made below the earlier state folder, so any other HOME blocks it.
/// The folder is the one the style file really lies in: a style file that a link of the config
/// folder brings to the default folder is the default folder's file and has its record. A style
/// file whose place cannot be found is the config folder's (its own refusal comes first).
fn record_name(home: &Path, host: &str) -> Result<String> {
    let Some(h) = crate::host_home::Host::from_name(host) else {
        return Ok(format!("{host}.json"));
    };
    let shared = locate(home, host, "style_file").is_ok_and(|s| s.shared);
    let folder = if shared {
        home.join(h.default_dir())
    } else {
        crate::host_home::resolve(h, home)?
    };
    if folder == home.join(h.default_dir()) {
        return Ok(format!("{host}.json"));
    }
    if !matches!(
        crate::state_root::classify_home(home).generation,
        "rust" | "rust-first"
    ) {
        return Err(format!(
            "{} is not the default {host} folder, so moai writes nothing to its output style here; next: run env -u CLAUDE_CONFIG_DIR moai update to update the default folder, or link {} to {} and run moai update again",
            folder.display(),
            folder.join("output-styles").display(),
            home.join(h.default_dir()).join("output-styles").display()
        ));
    }
    let digest = sha(folder.to_string_lossy().as_bytes());
    let hex = digest.strip_prefix("sha256:").unwrap_or(&digest);
    Ok(format!("{host}-{}.json", &hex[..12]))
}

fn read_state(home: &Path, host: &str) -> Result<HostState> {
    let root = state_root(home);
    plain_dir_or_absent(&crate::home_root::path(home))?;
    plain_dir_or_absent(&root)?;
    match read_opt(&root.join(record_name(home, host)?))? {
        None => Ok(HostState {
            schema_version: SCHEMA_VERSION,
            owned_sha256: None,
        }),
        Some(bytes) => {
            let state: HostState = serde_json::from_slice(&bytes)
                .map_err(|_| format!("style state for {host} is unreadable; inspect it first"))?;
            if state.schema_version != SCHEMA_VERSION {
                return Err(format!("style state for {host} has an unknown schema"));
            }
            Ok(state)
        }
    }
}

fn write_state(home: &Path, host: &str, owned: Option<String>) -> Result<()> {
    write_json(
        &state_root(home).join(record_name(home, host)?),
        &HostState {
            schema_version: SCHEMA_VERSION,
            owned_sha256: owned,
        },
    )
}

fn write_json<T: Serialize>(path: &Path, value: &T) -> Result<()> {
    let mut bytes = serde_json::to_vec_pretty(value).map_err(|_| "encode failed")?;
    bytes.push(b'\n');
    match read_opt(path)? {
        Some(b) => publish_replace(path, Some(&b), &bytes),
        None => publish_exclusive(path, &bytes),
    }
}

/// The style source, checked: UTF-8, at most 64 KiB, no marker tokens, no frontmatter of its
/// own, ends with a newline.
fn source(package: &Path) -> Result<String> {
    let path = package.join(SOURCE);
    let text = match fs::symlink_metadata(&path) {
        Ok(m) if m.is_file() => {
            fs::read_to_string(&path).map_err(|_| format!("{SOURCE} is not readable UTF-8"))?
        }
        _ => return Err(format!("{SOURCE} missing from the package")),
    };
    if text.len() > SOURCE_LIMIT
        || text.starts_with("---")
        || ["<!--", "-->", "moai:managed:"]
            .iter()
            .any(|t| text.contains(t))
        || !text.ends_with('\n')
    {
        return Err(format!(
            "{SOURCE} must be at most 64 KiB, start without frontmatter, hold no marker tokens and end with a newline"
        ));
    }
    Ok(text)
}

/// Claude Code's output style file: frontmatter, then the source.
pub fn render_style_file(source: &str) -> String {
    format!(
        "---\nname: {STYLE_NAME}\ndescription: \"{DESCRIPTION}\"\nkeep-coding-instructions: true\n---\n\n{source}"
    )
}

fn target(home: &Path, host: &str, kind: &str) -> TargetPlan {
    // The file that is written, and the link it is reached through. A target whose place cannot
    // be found shows the file in the config folder (the link itself for a refused link); an
    // unnamed host folder shows the path under HOME. Reading it blocks the target.
    let (path, link) = match locate(home, host, kind) {
        Ok(spot) => (spot.path, spot.link),
        Err(_) => match place(home, host, kind) {
            Ok((root, rel)) => (root.join(rel), None),
            Err(_) => {
                let default =
                    crate::host_home::Host::from_name(host).map_or(".", |h| h.default_dir());
                (home.join(default).join(rel_path(host, kind)), None)
            }
        },
    };
    TargetPlan {
        host: host.into(),
        kind: kind.into(),
        path: path.to_string_lossy().into_owned(),
        action: String::new(),
        expected_sha256: None,
        baseline_sha256: None,
        detail: None,
        link: link.map(|l| l.to_string_lossy().into_owned()),
        keys: Vec::new(),
    }
}

fn blocked(mut t: TargetPlan, reason: String) -> TargetPlan {
    t.action = "blocked".into();
    t.detail = Some(if reason.contains("next:") {
        reason
    } else {
        format!("{reason}; next: fix what this names, then run moai update again")
    });
    t
}

fn style_file_plan(home: &Path, rendered: &str) -> TargetPlan {
    let mut t = target(home, "claude", "style_file");
    let read = read_target(home, "claude", "style_file")
        .and_then(|(_, cur)| read_state(home, "claude").map(|s| (cur, s)));
    let (cur, state) = match read {
        Ok(v) => v,
        Err(e) => return blocked(t, e),
    };
    t.expected_sha256 = cur.as_deref().map(sha);
    t.baseline_sha256 = state.owned_sha256.clone();
    let record = state.owned_sha256.as_deref();
    t.action = match &cur {
        None => "create",
        // The current text: owned when the record says so; otherwise the same bytes are written
        // again and the record is brought to this text (a hand-copied install is adopted).
        Some(b) if b.as_slice() == rendered.as_bytes() => {
            if record == Some(sha(b).as_str()) {
                "unchanged"
            } else {
                t.detail =
                    Some("the file equals the current text; recorded as MoAI's own".into());
                "merge"
            }
        }
        // A text this module wrote earlier (the record names it) or shipped earlier (the list).
        Some(b) if record == Some(sha(b).as_str()) => "merge",
        Some(b) if is_earlier_render(&sha(b)) => {
            t.detail = Some("the file equals a text MoAI shipped earlier; updated".into());
            "merge"
        }
        Some(_) => {
            t.detail = Some(
                "the style file is not the one MoAI last wrote (edited or someone else's); preserve and review".into(),
            );
            "conflict"
        }
    }
    .into();
    t
}

/// The target of the earlier style file `output-styles/MoAI-Easy.md`, or none when it is known to
/// be absent (a plan without an old file has the targets it had before the rename). The file
/// is MoAI's own while it holds the bytes the record names or a text MoAI shipped
/// ([`EARLIER_RENDERS`]): then it is retired, unless the new style file will not be written, when
/// it is skipped. Any other bytes are the user's: `conflict`, left as they are. A folder or file
/// the link rules refuse blocks it with the style file's own reason, whether or not a file is
/// there, because nothing behind a refused link is read.
fn old_file_plan(home: &Path, style: &TargetPlan) -> Option<TargetPlan> {
    let mut t = target(home, "claude", "old_style_file");
    let cur = match read_target_rel(home, "claude", OLD_STYLE_FILE) {
        Ok((_, Some(cur))) => cur,
        Ok((_, None)) => return None,
        Err(e) => return Some(blocked(t, e)),
    };
    // The record that says which bytes are MoAI's cannot be read (its own place is refused): the
    // style file target already carries that reason, so the old file adds no target of its own.
    let Ok(state) = read_state(home, "claude") else {
        return None;
    };
    t.expected_sha256 = Some(sha(&cur));
    t.baseline_sha256 = state.owned_sha256.clone();
    let digest = sha(&cur);
    let own = state.owned_sha256.as_deref() == Some(digest.as_str()) || is_earlier_render(&digest);
    t.action = if !own {
        t.detail = Some(format!(
            "{OLD_STYLE_FILE} is not the one MoAI last wrote (edited or someone else's); left as it is, and so is `outputStyle`; to use {STYLE_NAME} set \"outputStyle\": \"{STYLE_NAME}\" in settings.json, or move that file away and run moai update again"
        ));
        "conflict"
    } else if !["create", "merge", "unchanged"].contains(&style.action.as_str()) {
        t.detail = Some(format!(
            "the new style file is {}; the old file is left as it is",
            style.action
        ));
        "skipped"
    } else {
        t.detail = Some(format!(
            "MoAI's own earlier style file; retired for {STYLE_NAME}, its bytes kept in the receipt"
        ));
        "retire"
    }
    .into();
    Some(t)
}

/// True when `digest` (as [`sha`] writes it, `sha256:` and 64 hex digits) is one of the texts
/// in [`EARLIER_RENDERS`].
fn is_earlier_render(digest: &str) -> bool {
    digest
        .strip_prefix("sha256:")
        .is_some_and(|hex| EARLIER_RENDERS.contains(&hex))
}

/// The settings object, or why it cannot be read as one.
fn settings_object(bytes: &[u8]) -> std::result::Result<Map<String, Value>, String> {
    if bytes.len() > SETTINGS_LIMIT {
        return Err("settings.json is larger than 256 KiB; not changed".into());
    }
    match serde_json::from_slice::<Value>(bytes) {
        Ok(Value::Object(map)) => Ok(map),
        _ => Err("settings.json is not a JSON object; not changed".into()),
    }
}

fn setting_plan(home: &Path, style: &TargetPlan, old: Option<&TargetPlan>) -> TargetPlan {
    let mut t = target(home, "claude", "setting");
    let (spot, cur) = match read_target(home, "claude", "setting") {
        Ok(v) => v,
        Err(e) => return blocked(t, e),
    };
    // A folder whose record cannot be made blocks both of its targets (REQ-7); a folder that is
    // refused for its own reason (a link) keeps that reason first.
    if let Err(e) = record_name(home, "claude") {
        return blocked(t, e);
    }
    t.expected_sha256 = cur.as_deref().map(sha);
    // Why `outputStyle` is not pointed at the style (it is the only key these guards hold back):
    // a style file MoAI does not own, or the default folder's own settings with the style file
    // written in another folder. The default folder's `settings.json` is read by the default
    // folder's Claude Code too, so it is pointed at the style only when the style file is written
    // in the default folder as well.
    let style_stop = if !["create", "merge", "unchanged"].contains(&style.action.as_str()) {
        Some(format!(
            "the style file is {}; the default is not set to a file MoAI does not own",
            style.action
        ))
    } else if spot.shared && !locate(home, "claude", "style_file").is_ok_and(|s| s.shared) {
        Some(
            "this settings file is the default folder's own; the style file is written in another folder, so the default folder would name a style it does not have; link output-styles to the default folder first"
                .into(),
        )
    } else {
        None
    };
    let map = match cur.as_deref().map(settings_object) {
        None => Map::new(),
        Some(Ok(map)) => map,
        // A file that is no JSON object stops both keys: a style stop is told first, as before.
        Some(Err(e)) => {
            match style_stop {
                Some(why) => {
                    t.action = "skipped".into();
                    t.detail = Some(why);
                }
                None => {
                    t.action = "conflict".into();
                    t.detail = Some(e);
                }
            }
            return t;
        }
    };
    let mut keys = Vec::new();
    let mut notes = Vec::new();
    let mut style_conflict = false;
    match &style_stop {
        Some(why) => notes.push(why.clone()),
        None => match map.get("outputStyle") {
            None => keys.push(SettingKey {
                key: "outputStyle".into(),
                value: json!(STYLE_NAME),
                replaces: None,
            }),
            Some(Value::String(s)) if s == STYLE_NAME => {}
            // Exactly the earlier name: swapped once the old file is retired or known absent. An
            // old file that stays (edited, or the new file not written) keeps the value too, so
            // the two never point apart.
            Some(Value::String(s)) if s == OLD_STYLE_NAME => match old.map(|o| o.action.as_str()) {
                None | Some("retire") => keys.push(SettingKey {
                    key: "outputStyle".into(),
                    value: json!(STYLE_NAME),
                    replaces: Some(json!(OLD_STYLE_NAME)),
                }),
                Some(action) => {
                    style_conflict = true;
                    notes.push(format!(
                            "outputStyle is \"{s}\": the old style file {OLD_STYLE_FILE} is {action}, so the value is not changed"
                        ));
                }
            },
            // Claude Code compares the name with case: a near miss silently selects its default.
            Some(Value::String(s)) if s.eq_ignore_ascii_case(STYLE_NAME) => {
                style_conflict = true;
                notes.push(format!(
                    "outputStyle is \"{s}\": the name is case-sensitive, so Claude Code uses its default style, not {STYLE_NAME}; not changed"
                ));
            }
            Some(other) => {
                style_conflict = true;
                notes.push(format!(
                    "outputStyle is {other}: the user's choice; not changed"
                ));
            }
        },
    }
    // `language` is planned apart from the style file: only a key that is absent is written.
    match language_key(home) {
        Ok(k) => match map.get("language") {
            None => keys.push(k),
            // The value already there is the one MoAI would write: nothing to say.
            Some(v) if *v == k.value => {}
            Some(v) => notes.push(format!("language is {v}: the user's value; kept")),
        },
        Err(reason) => notes.push(reason),
    }
    // A text the swap cannot find exactly once is no swap: the value is left as it is.
    if keys.iter().any(|k| k.replaces.is_some()) {
        let tried = cur
            .as_deref()
            .ok_or_else(|| "settings.json is absent".to_string())
            .and_then(|b| insert_settings(Some(b), &keys).map(|_| ()));
        if let Err(why) = tried {
            keys.retain(|k| k.replaces.is_none());
            style_conflict = true;
            notes.push(format!(
                "outputStyle is \"{OLD_STYLE_NAME}\" but {SWAP_REFUSED} ({why}); not changed"
            ));
        }
    }
    t.action = if keys.iter().any(|k| k.replaces.is_some()) {
        "replace"
    } else if !keys.is_empty() {
        "set"
    } else if style_conflict {
        "conflict"
    } else if style_stop.is_some() {
        "skipped"
    } else {
        "unchanged"
    }
    .into();
    t.detail = (!notes.is_empty()).then(|| notes.join("; "));
    t.keys = keys;
    t
}

/// The `language` key to plan: Claude Code passes the value to the model verbatim as the language
/// to answer in, so it is the lower-case English name MoAI's own table gives the conversation
/// language the installer saved. `Err` names why no key is planned.
fn language_key(home: &Path) -> std::result::Result<SettingKey, String> {
    let Some(code) = crate::language::home_conversation(home) else {
        return Err("language: no conversation language saved; not set".into());
    };
    match crate::language::english_name(&code) {
        Some(name) => Ok(SettingKey {
            key: "language".into(),
            value: json!(name.to_lowercase()),
            replaces: None,
        }),
        None => Err(format!(
            "language: no name for the saved conversation language `{code}`; not set"
        )),
    }
}

/// The `outputStyle` string a project settings file holds when it is another name than ours; none
/// for a file that is absent, linked, not a regular file, too large, not a JSON object, or without
/// a string value (REQ-3). Reads only.
fn hiding_style(path: &Path) -> Option<String> {
    let meta = fs::symlink_metadata(path).ok()?;
    if !meta.is_file() || meta.len() > SETTINGS_LIMIT as u64 {
        return None;
    }
    let map = settings_object(&fs::read(path).ok()?).ok()?;
    match map.get("outputStyle") {
        Some(Value::String(s)) if s != STYLE_NAME => Some(s.clone()),
        _ => None,
    }
}

/// The notices of a style plan for `moai update` (SPEC-20261122-output-style-install-checks):
/// one entry per project settings file whose `outputStyle` hides ours in `project` (REQ-2, read
/// only), and one `/status` pointer when the plan writes the style file (REQ-4). Entries have the
/// shape of the update's `skipped` entries but go to its own `notices` list, which the update
/// screens print with their own word; they are not part of any plan digest and nothing is
/// written for them.
pub fn notices(plan: &Plan, project: &Path) -> Vec<Value> {
    let mut out = Vec::new();
    for rel in [".claude/settings.local.json", ".claude/settings.json"] {
        if let Some(value) = hiding_style(&project.join(rel)) {
            out.push(json!({
                "part": "style",
                "host": "claude",
                "reason": format!("{rel} sets outputStyle \"{value}\": this project uses it, not {STYLE_NAME}; moai does not change project settings"),
            }));
        }
    }
    for t in plan.targets.iter().filter(|t| t.action == "blocked") {
        out.push(json!({
            "part": "style",
            "host": "claude",
            "reason": format!("{} was not written: {}", t.path, t.detail.as_deref().unwrap_or("blocked")),
        }));
    }
    // An old style file the user edited stays, and so does `outputStyle`: say how to choose.
    for t in plan
        .targets
        .iter()
        .filter(|t| t.kind == "old_style_file" && t.action == "conflict")
    {
        out.push(json!({
            "part": "style",
            "host": "claude",
            "reason": format!("{} is not the file MoAI installed (edited or someone else's), so it and `outputStyle` stay as they are; to use {STYLE_NAME}, set \"outputStyle\": \"{STYLE_NAME}\" in settings.json, or move that file away and run moai update again", t.path),
        }));
    }
    let writes_style = plan
        .targets
        .iter()
        .any(|t| t.kind == "style_file" && matches!(t.action.as_str(), "create" | "merge"));
    if writes_style {
        out.push(json!({
            "part": "style",
            "host": "claude",
            "reason": "if the style does not show up in Claude Code, `/status` lists a style file it could not load",
        }));
    }
    out
}

/// Read-only preview of every host; writes nothing.
pub fn plan(home: &Path, package: &Path) -> Result<Plan> {
    plan_for(home, package, &HOSTS)
}

/// Read-only preview of the named hosts only (`moai update` leaves out a host whose skill copy
/// is blocked); writes nothing. An unknown host is refused.
pub fn plan_for(home: &Path, package: &Path, only: &[&str]) -> Result<Plan> {
    if let Some(h) = only.iter().find(|h| !HOSTS.contains(h)) {
        return Err(format!("style host not supported: {h:?}"));
    }
    let home = fs::canonicalize(home).map_err(|_| "home directory unavailable")?;
    // No host to write: the source is not needed (a package from before #2875 has none).
    let src = if only.is_empty() {
        String::new()
    } else {
        source(package)?
    };
    let mut targets = Vec::new();
    if only.contains(&"claude") {
        let style = style_file_plan(&home, &render_style_file(&src));
        let mut old = old_file_plan(&home, &style);
        let setting = setting_plan(&home, &style, old.as_ref());
        // The old file and the value move together: a value that cannot be swapped keeps the
        // old file too.
        if let Some(o) = old.as_mut().filter(|o| o.action == "retire") {
            if setting
                .detail
                .as_deref()
                .is_some_and(|d| d.contains(SWAP_REFUSED))
            {
                o.action = "skipped".into();
                o.detail = Some(format!(
                    "outputStyle could not be swapped, so {OLD_STYLE_FILE} is left as it is"
                ));
            }
        }
        targets.push(style);
        targets.push(setting);
        targets.extend(old);
    }
    let mut plan = Plan {
        schema_version: SCHEMA_VERSION,
        home: home.to_str().ok_or("non UTF-8 home path")?.to_string(),
        package_digest: crate::package_digest(package)?,
        source_sha256: sha(src.as_bytes()),
        targets,
        not_supported: crate::home_routing::NOT_SUPPORTED
            .iter()
            .map(|(h, r)| Unsupported {
                host: (*h).into(),
                reason: (*r).into(),
            })
            .collect(),
        plan_digest: String::new(),
    };
    plan.plan_digest = sha(&serde_json::to_vec(&plan).map_err(|_| "encode failed")?);
    Ok(plan)
}

/// True when applying `plan` would write at least one file.
pub fn writes_anything(plan: &Plan) -> bool {
    plan.targets.iter().any(|t| is_write(&t.action))
}

/// Whether two plans write the same targets (#2875): `moai update` re-plans the style after the
/// other parts of the layer were applied, so a file's sha256 may differ from the confirmed
/// preview; which targets are written, and which settings keys with which values, must not.
pub fn same_writes(a: &Plan, b: &Plan) -> bool {
    type Write = (String, String, bool, Vec<SettingKey>);
    let writes = |p: &Plan| -> Vec<Write> {
        p.targets
            .iter()
            .map(|t| {
                let write = is_write(&t.action);
                let keys = if write { t.keys.clone() } else { Vec::new() };
                (t.host.clone(), t.kind.clone(), write, keys)
            })
            .collect()
    };
    a.source_sha256 == b.source_sha256 && writes(a) == writes(b)
}

fn is_write(action: &str) -> bool {
    ["create", "append", "merge", "set", "replace", "retire"].contains(&action)
}

fn validate_permission(permission: &str) -> Result<()> {
    if !crate::contract::permission_reference_ok(permission) {
        return Err("permission reference required".into());
    }
    Ok(())
}

fn new_receipt_dir(home: &Path) -> Result<(String, PathBuf)> {
    plain_dir_or_absent(&crate::home_root::path(home))?;
    let receipts = state_root(home).join("receipts");
    fs::create_dir_all(&receipts).map_err(|_| "style receipt folder unavailable")?;
    plain_dir_or_absent(&state_root(home))?;
    plain_dir_or_absent(&receipts)?;
    let nanos = std::time::SystemTime::now()
        .duration_since(std::time::UNIX_EPOCH)
        .map(|d| d.as_nanos())
        .unwrap_or(0);
    for n in 0..16 {
        let id = format!("style-{nanos}-{n}");
        let dir = receipts.join(&id);
        if fs::create_dir(&dir).is_ok() {
            return Ok((id, dir));
        }
    }
    Err("style receipt id unavailable".into())
}

/// The bytes a write leaves in the target file.
fn new_bytes(t: &TargetPlan, before: Option<&[u8]>, src: &str) -> Result<Vec<u8>> {
    match t.kind.as_str() {
        "style_file" => Ok(render_style_file(src).into_bytes()),
        "setting" => insert_settings(before, &t.keys).map(|(out, _)| out),
        other => unreachable!("unknown style target kind {other}"),
    }
}

/// `text` with the value of `"key": <old>` swapped for `new`, keeping every other byte: the
/// spacing between key and value stays. Returns the new text, the exact `"key": <old>` text taken
/// away and the exact `"key": <new>` text put in. The old pair must occur exactly once.
fn swap_value_text(
    text: &str,
    key: &str,
    old: &Value,
    new: &Value,
) -> std::result::Result<(String, String, String), String> {
    let quoted = format!("\"{key}\"");
    let old_json = old.to_string();
    let mut found = Vec::new();
    for (at, _) in text.match_indices(&quoted) {
        let after_key = at + quoted.len();
        let rest = &text[after_key..];
        let Some(rest) = rest.trim_start().strip_prefix(':') else {
            continue;
        };
        let value_at = text.len() - rest.trim_start().len();
        if text[value_at..].starts_with(&old_json) {
            found.push((at, value_at, value_at + old_json.len()));
        }
    }
    let [(at, value_at, end)] = found[..] else {
        return Err(format!(
            "`\"{key}\": {old_json}` is in the file {} times, not once",
            found.len()
        ));
    };
    let old_text = text[at..end].to_string();
    let new_text = format!("{}{new}", &text[at..value_at]);
    let out = format!("{}{new_text}{}", &text[..at], &text[end..]);
    Ok((out, old_text, new_text))
}

/// `settings.json` with each planned key added as text before the object's closing brace, or
/// swapped as text where it holds `replaces`, and what was put in per key (#2875 audit B1): every
/// other byte stays as the user wrote it — no re-encoding of numbers, indentation or order. The
/// result must parse to the same object plus exactly those changes, or nothing is written.
fn insert_settings(
    before: Option<&[u8]>,
    keys: &[SettingKey],
) -> Result<(Vec<u8>, Vec<InsertedKey>)> {
    if keys.is_empty() {
        return Err("no settings key to add".into());
    }
    let pair = |k: &SettingKey| format!("\"{}\": {}", k.key, k.value);
    let kept = |k: &SettingKey, text: String| InsertedKey {
        key: k.key.clone(),
        value: k.value.clone(),
        text,
        replaced: None,
    };
    let Some(bytes) = before else {
        if keys.iter().any(|k| k.replaces.is_some()) {
            return Err("a value cannot be swapped in a file that is not there".into());
        }
        let mut inserted = Vec::new();
        let mut body = String::new();
        for (i, k) in keys.iter().enumerate() {
            let text = if i == 0 {
                pair(k)
            } else {
                format!(",\n  {}", pair(k))
            };
            body.push_str(&text);
            inserted.push(kept(k, text));
        }
        return Ok((format!("{{\n  {body}\n}}\n").into_bytes(), inserted));
    };
    let old = settings_object(bytes)?;
    let mut text = std::str::from_utf8(bytes)
        .map_err(|_| "settings.json is not UTF-8; not changed")?
        .to_string();
    let mut inserted = Vec::new();
    let mut want = old;
    // The swaps first, on the user's text; then the added keys go before the closing brace.
    for k in keys.iter().filter(|k| k.replaces.is_some()) {
        let was = k.replaces.as_ref().expect("filtered");
        if want.get(&k.key) != Some(was) {
            return Err(format!("{} no longer holds the value to swap", k.key));
        }
        let (next, old_text, new_text) = swap_value_text(&text, &k.key, was, &k.value)?;
        text = next;
        want.insert(k.key.clone(), k.value.clone());
        inserted.push(InsertedKey {
            key: k.key.clone(),
            value: k.value.clone(),
            text: new_text,
            replaced: Some(ReplacedText {
                value: was.clone(),
                text: old_text,
            }),
        });
    }
    let adds: Vec<&SettingKey> = keys.iter().filter(|k| k.replaces.is_none()).collect();
    if !adds.is_empty() {
        let close = text
            .trim_end()
            .strip_suffix('}')
            .map(str::len)
            .ok_or("settings.json does not end with `}`; not changed")?;
        let body_end = text[..close].trim_end().len();
        let open = text
            .find('{')
            .ok_or("settings.json has no `{`; not changed")?;
        let inner = &text[open + 1..close];
        let empty = inner.trim().is_empty();
        let separator = if inner.contains('\n') {
            let indent: String = inner
                .lines()
                .find(|l| !l.trim().is_empty())
                .map(|l| l.chars().take_while(|c| c.is_whitespace()).collect())
                .unwrap_or_else(|| "  ".into());
            format!(",\n{indent}")
        } else {
            ", ".to_string()
        };
        let mut add = String::new();
        for (i, k) in adds.iter().enumerate() {
            let piece = if empty && i == 0 {
                pair(k)
            } else {
                format!("{separator}{}", pair(k))
            };
            add.push_str(&piece);
            want.insert(k.key.clone(), k.value.clone());
            inserted.push(kept(k, piece));
        }
        text = format!("{}{add}{}", &text[..body_end], &text[body_end..]);
    }
    if settings_object(text.as_bytes()).ok() != Some(want) {
        return Err("settings.json could not take the keys as text; not changed".into());
    }
    Ok((text.into_bytes(), inserted))
}

/// Applies a previewed plan. `plan_digest` must equal a fresh plan's digest; only `create`,
/// `append`, `merge` and `set` are written, blocked targets and conflicts are skipped.
pub fn apply(home: &Path, package: &Path, plan_digest: &str, permission: &str) -> Result<Value> {
    apply_for(home, package, &HOSTS, plan_digest, permission)
}

/// [`apply`] for the named hosts only; `plan_digest` must equal a fresh [`plan_for`] of the
/// same hosts.
pub fn apply_for(
    home: &Path,
    package: &Path,
    only: &[&str],
    plan_digest: &str,
    permission: &str,
) -> Result<Value> {
    validate_permission(permission)?;
    let fresh = plan_for(home, package, only)?;
    if fresh.plan_digest != plan_digest {
        return Err("stale or altered plan; re-preview required".into());
    }
    let home = PathBuf::from(&fresh.home);
    let writes: Vec<&TargetPlan> = fresh
        .targets
        .iter()
        .filter(|t| is_write(&t.action))
        .collect();
    if writes.is_empty() {
        return Ok(json!({"state": "unchanged", "plan_digest": fresh.plan_digest, "written": []}));
    }
    let src = source(package)?;
    let (id, receipt_dir) = new_receipt_dir(&home)?;
    let mut receipt = Receipt {
        schema_version: SCHEMA_VERSION,
        id: id.clone(),
        plan_digest: fresh.plan_digest.clone(),
        permission_reference: permission.to_string(),
        status: "pending".into(),
        entries: Vec::new(),
        config_folder: Some(claude_folder(&home)?.to_string_lossy().into_owned()),
    };
    let receipt_path = receipt_dir.join("receipt.json");
    write_json(&receipt_path, &receipt)?;
    let mut written = Vec::new();
    for t in writes {
        // The real file, found again now (a link may have changed since the plan): checked before
        // creating (create_dir_all would follow a linked parent) and after.
        let spot = locate(&home, &t.host, &t.kind)?;
        let path = spot.path.clone();
        let folder = path.parent().expect("target folder");
        fs::create_dir_all(folder).map_err(|_| format!("{} unavailable", folder.display()))?;
        if locate(&home, &t.host, &t.kind)?.path != path {
            return Err(format!(
                "{} changed while it was written; re-preview required",
                spot.link.as_deref().unwrap_or(&path).display()
            ));
        }
        let before = current_file(&path)?;
        if before.as_deref().map(sha) != t.expected_sha256 {
            return Err(format!(
                "{} changed since the plan; re-preview required",
                path.display()
            ));
        }
        let retire = t.kind == "old_style_file";
        let (out, inserted_keys) = if retire {
            // The file is going away: nothing is written, and the entry records the hash of
            // nothing as what the apply left.
            if before.is_none() {
                return Err(format!(
                    "{} is gone since the plan; re-preview required",
                    path.display()
                ));
            }
            (Vec::new(), Vec::new())
        } else if t.kind == "setting" {
            insert_settings(before.as_deref(), &t.keys)?
        } else {
            (new_bytes(t, before.as_deref(), &src)?, Vec::new())
        };
        // The texts of the keys that were added together, as a receipt before the key list
        // recorded them; a swapped key is told by its `replaced` text only.
        let inserted = Some(
            inserted_keys
                .iter()
                .filter(|k| k.replaced.is_none())
                .map(|k| k.text.as_str())
                .collect::<String>(),
        )
        .filter(|s| !s.is_empty());
        if let Some(b) = &before {
            publish_exclusive(
                &receipt_dir.join(format!("{}-{}.before", t.host, t.kind)),
                b,
            )?;
        }
        // The entry is recorded before the write (#2875 audit N3): an apply cut after the write
        // still leaves what a revert needs.
        receipt.entries.push(ReceiptEntry {
            host: t.host.clone(),
            kind: t.kind.clone(),
            path: Some(rel_path(&t.host, &t.kind).to_string()),
            action: t.action.clone(),
            before_sha256: before.as_deref().map(sha),
            after_sha256: sha(&out),
            baseline_before: t.baseline_sha256.clone(),
            inserted,
            inserted_keys,
        });
        write_json(&receipt_path, &receipt)?;
        if retire {
            // The same bytes the bytes were kept from, read again right before the removal.
            if current_file(&path)?.as_deref().map(sha) != t.expected_sha256 {
                return Err(format!(
                    "{} changed since the plan; re-preview required",
                    path.display()
                ));
            }
            fs::remove_file(&path).map_err(|_| format!("{} not removed", path.display()))?;
            written.push(format!("{}:{}", t.host, t.kind));
            continue;
        }
        match &before {
            Some(b) => publish_replace(&path, Some(b), &out)?,
            None => publish_exclusive(&path, &out)?,
        }
        if t.kind == "style_file" {
            write_state(&home, &t.host, Some(sha(&out)))?;
        }
        written.push(format!("{}:{}", t.host, t.kind));
    }
    receipt.status = "committed".into();
    write_json(&receipt_path, &receipt)?;
    Ok(json!({
        "state": "applied",
        "plan_digest": fresh.plan_digest,
        "receipt": id,
        "written": written,
    }))
}

fn receipt_id_ok(id: &str) -> bool {
    id.starts_with("style-")
        && id.len() <= 64
        && id.bytes().all(|b| b.is_ascii_alphanumeric() || b == b'-')
}

/// The keys an entry added: the receipt's own list, or for a receipt from before the key list
/// the one `outputStyle` text it recorded in `inserted`.
fn entry_keys(e: &ReceiptEntry) -> Option<Vec<InsertedKey>> {
    if !e.inserted_keys.is_empty() {
        return Some(e.inserted_keys.clone());
    }
    // The value is read from the text the receipt kept: a receipt from before the key list was
    // written by the build that named the style MoAI-Easy, so it holds that name, not the
    // current one; a text that cannot be read falls back to the earlier name.
    e.inserted.as_ref().map(|text| {
        let body = text.trim_start_matches(|c: char| c == ',' || c.is_whitespace());
        let value = serde_json::from_str::<Value>(&format!("{{{body}}}"))
            .ok()
            .and_then(|v| v.get("outputStyle").cloned())
            .unwrap_or_else(|| json!(OLD_STYLE_NAME));
        vec![InsertedKey {
            key: "outputStyle".into(),
            value,
            text: text.clone(),
            replaced: None,
        }]
    })
}

/// `text` without the comma a removed first pair left in front of the next one.
fn drop_leading_comma(text: String) -> String {
    let Some(open) = text.find('{') else {
        return text;
    };
    let rest = &text[open + 1..];
    let skipped = rest.len() - rest.trim_start().len();
    if rest[skipped..].starts_with(',') {
        let at = open + 1 + skipped;
        return format!("{}{}", &text[..at], &text[at + 1..]);
    }
    text
}

/// Undoes the keys the apply added (#2875 audit B1). A file still holding the bytes the apply
/// wrote gets its before bytes back byte for byte (a created file is removed). A file changed
/// since loses, key by key, only the exact text the apply inserted for a key that still holds the
/// value it wrote, while that text is there once and the rest parses to the same object without
/// the key. The user's other values are never re-encoded. Returns the keys left in place; a file
/// where no key could go is an error.
fn revert_setting(
    path: &Path,
    dir: &Path,
    e: &ReceiptEntry,
) -> std::result::Result<Vec<String>, String> {
    let current = current_file(path)
        .map_err(|r| format!("file_not_plain: {r}"))?
        .ok_or("user_changed_since_apply")?;
    if sha(&current) == e.after_sha256 {
        return match &e.before_sha256 {
            None => fs::remove_file(path)
                .map(|()| Vec::new())
                .map_err(|_| format!("{} not removed", path.display())),
            Some(before_sha) => {
                let before = read_opt(&dir.join(format!("{}-{}.before", e.host, e.kind)))?
                    .ok_or("before bytes missing from the receipt")?;
                if &sha(&before) != before_sha {
                    return Err("before bytes do not match the receipt".into());
                }
                publish_replace(path, Some(&current), &before).map(|()| Vec::new())
            }
        };
    }
    let keys = entry_keys(e).ok_or("user_changed_since_apply")?;
    let mut text = std::str::from_utf8(&current)
        .map_err(|_| "user_changed_since_apply")?
        .to_string();
    let mut object = settings_object(&current).map_err(|_| "user_changed_since_apply")?;
    let mut kept = Vec::new();
    let mut removed = 0;
    for k in keys.iter().rev() {
        if let Some(r) = &k.replaced {
            // A swapped key goes back to the text it had, while it still holds what the apply
            // wrote, that text is there once, and the result differs only in that value.
            let mut want = object.clone();
            let holds = want.insert(k.key.clone(), r.value.clone()).as_ref() == Some(&k.value);
            if holds && text.matches(&k.text).count() == 1 {
                let out = text.replacen(&k.text, &r.text, 1);
                if settings_object(out.as_bytes()).ok().as_ref() == Some(&want) {
                    text = out;
                    object = want;
                    removed += 1;
                    continue;
                }
            }
            kept.push(k.key.clone());
            continue;
        }
        let mut rest = object.clone();
        let removable = rest.shift_remove(&k.key).as_ref() == Some(&k.value)
            && text.matches(&k.text).count() == 1;
        if removable {
            let out = drop_leading_comma(text.replacen(&k.text, "", 1));
            if settings_object(out.as_bytes()).ok().as_ref() == Some(&rest) {
                text = out;
                object = rest;
                removed += 1;
                continue;
            }
        }
        kept.push(k.key.clone());
    }
    if removed == 0 {
        return Err("user_changed_since_apply".into());
    }
    publish_replace(path, Some(&current), text.as_bytes())?;
    kept.reverse();
    Ok(kept)
}

/// Undoes one apply for the targets that still hold what it wrote: a created file is removed, a
/// changed one gets its before bytes back, the keys the apply added are removed (one changed
/// since stays and is named in `skipped`). A target changed since is skipped.
pub fn revert(home: &Path, receipt_id: &str, permission: &str) -> Result<Value> {
    revert_in(&HOSTS, home, receipt_id, permission)
}

/// [`revert`] for a place whose active hosts are `active`: a receipt entry of a retired host is
/// skipped and reported (reason `host_retired`), the other entries are reverted.
pub fn revert_in(
    active: &[&str],
    home: &Path,
    receipt_id: &str,
    permission: &str,
) -> Result<Value> {
    validate_permission(permission)?;
    if !receipt_id_ok(receipt_id) {
        return Err("unknown style receipt id".into());
    }
    let home = fs::canonicalize(home).map_err(|_| "home directory unavailable")?;
    plain_dir_or_absent(&crate::home_root::path(&home))?;
    plain_dir_or_absent(&state_root(&home))?;
    // Every folder on the way, `receipts` included (#2875 audit B3): a link is never followed.
    plain_dir_or_absent(&state_root(&home).join("receipts"))?;
    let dir = state_root(&home).join("receipts").join(receipt_id);
    plain_dir_or_absent(&dir)?;
    let path = dir.join("receipt.json");
    let bytes = read_opt(&path)?.ok_or("style receipt not found")?;
    let mut receipt: Receipt =
        serde_json::from_slice(&bytes).map_err(|_| "style receipt is unreadable")?;
    if receipt.status == "reverted" {
        return Err("style receipt already reverted".into());
    }
    receipt_folder_matches(&home, &receipt)?;
    let mut reverted = Vec::new();
    let mut skipped = Vec::new();
    for e in receipt.entries.iter().rev() {
        if retired(&e.host, active) {
            skipped.push(
                json!({"target": format!("{}:{}", e.host, e.kind), "reason": "host_retired"}),
            );
            continue;
        }
        // The file the entry names (a receipt of the earlier format names none: its style file
        // is the old file); a path outside the three this module writes is refused.
        let rel = entry_rel(e)?;
        let name = format!("{}:{}", e.host, e.kind);
        let spot = match locate_rel_labeled(&home, &e.host, rel) {
            Ok(spot) => spot,
            Err((label, reason)) => {
                skipped.push(json!({"target": name, "reason": format!("{label}: {reason}")}));
                continue;
            }
        };
        let target = spot.path;
        if e.kind == "old_style_file" {
            // The retired file comes back only where nothing is now: a file the user put there
            // since stays untouched.
            match current_file(&target) {
                Ok(None) => {}
                Ok(Some(_)) => {
                    skipped.push(json!({"target": name, "reason": "user_changed_since_apply"}));
                    continue;
                }
                Err(reason) => {
                    skipped.push(
                        json!({"target": name, "reason": format!("file_not_plain: {reason}")}),
                    );
                    continue;
                }
            }
            let before_sha = e
                .before_sha256
                .as_ref()
                .ok_or("before bytes missing from the receipt")?;
            let before = read_opt(&dir.join(format!("{}-{}.before", e.host, e.kind)))?
                .ok_or("before bytes missing from the receipt")?;
            if &sha(&before) != before_sha {
                return Err("before bytes do not match the receipt".into());
            }
            let folder = target.parent().expect("target folder");
            fs::create_dir_all(folder).map_err(|_| format!("{} unavailable", folder.display()))?;
            if locate_rel_labeled(&home, &e.host, rel)
                .map(|s| s.path)
                .ok()
                .as_ref()
                != Some(&target)
            {
                return Err(format!(
                    "{} changed while it was restored; run the revert again",
                    target.display()
                ));
            }
            publish_exclusive(&target, &before)?;
            reverted.push(name);
            continue;
        }
        if e.kind == "setting" {
            match revert_setting(&target, &dir, e) {
                Ok(kept) => {
                    // A key changed since the apply stays; it is named, not hidden.
                    if !kept.is_empty() {
                        skipped.push(json!({
                            "target": name,
                            "reason": "user_changed_since_apply",
                            "keys": kept,
                        }));
                    }
                    reverted.push(name);
                }
                Err(reason) => skipped.push(json!({"target": name, "reason": reason})),
            }
            continue;
        }
        let current = match current_file(&target) {
            Ok(c) => c,
            Err(reason) => {
                skipped
                    .push(json!({"target": name, "reason": format!("file_not_plain: {reason}")}));
                continue;
            }
        };
        if current.as_deref().map(sha) != Some(e.after_sha256.clone()) {
            skipped.push(json!({"target": name, "reason": "user_changed_since_apply"}));
            continue;
        }
        let current = current.expect("checked above");
        match &e.before_sha256 {
            None => {
                fs::remove_file(&target)
                    .map_err(|_| format!("{} not removed", target.display()))?;
            }
            Some(before_sha) => {
                let before = read_opt(&dir.join(format!("{}-{}.before", e.host, e.kind)))?
                    .ok_or("before bytes missing from the receipt")?;
                if &sha(&before) != before_sha {
                    return Err("before bytes do not match the receipt".into());
                }
                publish_replace(&target, Some(&current), &before)?;
            }
        }
        write_state(&home, &e.host, e.baseline_before.clone())?;
        reverted.push(name);
    }
    receipt.status = "reverted".into();
    write_json(&path, &receipt)?;
    Ok(
        json!({"state": "reverted", "receipt": receipt_id, "reverted": reverted, "skipped": skipped}),
    )
}

#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    fn the_style_file_names_the_style_claude_selects() {
        let out = render_style_file("body\n");
        assert!(out.starts_with("---\nname: MoAI-Style\n"));
        assert!(out.contains("\nkeep-coding-instructions: true\n---\n\nbody\n"));
    }

    #[test]
    fn a_settings_file_that_is_not_an_object_is_refused() {
        assert!(settings_object(b"[]").is_err());
        assert!(settings_object(b"{").is_err());
        assert!(settings_object(&vec![b' '; SETTINGS_LIMIT + 1]).is_err());
        assert!(settings_object(b"{\"a\":1}").is_ok());
    }

    fn plan_of(source: &str, targets: &[(&str, &str, &str)]) -> Plan {
        Plan {
            schema_version: SCHEMA_VERSION,
            home: "/h".into(),
            package_digest: "p".into(),
            source_sha256: source.into(),
            targets: targets
                .iter()
                .map(|(host, kind, action)| TargetPlan {
                    host: (*host).into(),
                    kind: (*kind).into(),
                    path: String::new(),
                    action: (*action).into(),
                    expected_sha256: None,
                    baseline_sha256: None,
                    detail: None,
                    link: None,
                    keys: Vec::new(),
                })
                .collect(),
            not_supported: Vec::new(),
            plan_digest: String::new(),
        }
    }

    /// #2875 audit N1: `moai update` accepts a re-plan that writes the same targets from the same
    /// source — the style file went from `create` to `merge` — and refuses any other change.
    #[test]
    fn same_writes_accepts_only_the_same_written_targets() {
        let confirmed = plan_of(
            "s",
            &[
                ("claude", "style_file", "create"),
                ("claude", "setting", "set"),
            ],
        );
        let merged = plan_of(
            "s",
            &[
                ("claude", "style_file", "merge"),
                ("claude", "setting", "set"),
            ],
        );
        assert!(same_writes(&confirmed, &merged));
        let conflict = plan_of(
            "s",
            &[
                ("claude", "style_file", "conflict"),
                ("claude", "setting", "set"),
            ],
        );
        assert!(!same_writes(&confirmed, &conflict));
        let other_source = plan_of(
            "t",
            &[
                ("claude", "style_file", "merge"),
                ("claude", "setting", "set"),
            ],
        );
        assert!(!same_writes(&confirmed, &other_source));
        let fewer = plan_of("s", &[("claude", "setting", "set")]);
        assert!(!same_writes(&confirmed, &fewer));
    }

    fn key(name: &str, value: &str) -> SettingKey {
        SettingKey {
            key: name.into(),
            value: json!(value),
            replaces: None,
        }
    }

    /// A language that changed between the preview and the apply is a different write.
    #[test]
    fn same_writes_compares_the_keys_of_a_written_target() {
        let mut korean = plan_of("s", &[("claude", "setting", "set")]);
        korean.targets[0].keys = vec![key("language", "korean")];
        let mut same = korean.clone();
        // Another file hash or detail is no different write.
        same.targets[0].expected_sha256 = Some("x".into());
        same.targets[0].detail = Some("d".into());
        assert!(same_writes(&korean, &same));
        let mut english = korean.clone();
        english.targets[0].keys = vec![key("language", "english")];
        assert!(!same_writes(&korean, &english));
        let mut more = korean.clone();
        more.targets[0]
            .keys
            .insert(0, key("outputStyle", STYLE_NAME));
        assert!(!same_writes(&korean, &more));
        // The keys of a target that writes nothing are not compared.
        let mut idle_a = plan_of("s", &[("claude", "setting", "unchanged")]);
        let idle_b = idle_a.clone();
        idle_a.targets[0].keys = vec![key("language", "korean")];
        assert!(same_writes(&idle_a, &idle_b));
    }

    #[test]
    fn insert_settings_adds_exactly_the_planned_keys_as_text() {
        let keys = [key("outputStyle", STYLE_NAME), key("language", "korean")];
        let (out, inserted) = insert_settings(Some(b"{\"a\": 1}"), &keys).unwrap();
        assert_eq!(
            String::from_utf8(out).unwrap(),
            "{\"a\": 1, \"outputStyle\": \"MoAI-Style\", \"language\": \"korean\"}"
        );
        let texts: Vec<&str> = inserted.iter().map(|k| k.text.as_str()).collect();
        assert_eq!(
            texts,
            [
                ", \"outputStyle\": \"MoAI-Style\"",
                ", \"language\": \"korean\""
            ]
        );
        let (out, inserted) = insert_settings(Some(b"{ }\n"), &keys).unwrap();
        assert_eq!(
            String::from_utf8(out).unwrap(),
            "{\"outputStyle\": \"MoAI-Style\", \"language\": \"korean\" }\n"
        );
        assert_eq!(inserted[0].text, "\"outputStyle\": \"MoAI-Style\"");
        let (out, _) = insert_settings(None, &keys[1..]).unwrap();
        assert_eq!(
            String::from_utf8(out).unwrap(),
            "{\n  \"language\": \"korean\"\n}\n"
        );
        assert!(insert_settings(Some(b"{}"), &[]).is_err());
        assert!(insert_settings(Some(b"[1]"), &keys).is_err());
    }

    /// A first pair removed from an object that was empty leaves the comma of the next one; it
    /// goes, and an object that does not start with a comma is left alone.
    #[test]
    fn drop_leading_comma_removes_only_a_comma_that_opens_the_object() {
        assert_eq!(
            drop_leading_comma("{\n  ,\n  \"b\": 2\n}".into()),
            "{\n  \n  \"b\": 2\n}"
        );
        assert_eq!(
            drop_leading_comma("{\"a\": 1, \"b\": 2}".into()),
            "{\"a\": 1, \"b\": 2}"
        );
    }
}
