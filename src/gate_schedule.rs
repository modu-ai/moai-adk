//! A scheduler entry for `gate-loop --once` (#2906 S2): a macOS LaunchAgent plist
//! (`StartInterval 90`) or a Linux crontab file (`*/2`, under `flock -n`), previewed by
//! `gate-schedule-plan` and written by `gate-schedule-apply`.
//!
//! The entry runs `gate-loop --once --dry-run` unless `--merge` was asked, and then carries the
//! operator's `--permission-reference` (decision Q3). It never carries `--review-confirmed`:
//! G7 stays a human stop. Its environment is PATH plus the named `--env` variables, and a
//! token-like variable (in either) is refused, so no credential lands in a file.
//!
//! `plan` writes nothing. `apply` re-plans, refuses any other digest, keeps the previous bytes
//! in a receipt and writes the one file; it never runs `launchctl` or `crontab` — the enable
//! and disable commands are printed as text for the operator (decision Q5). Windows is refused
//! with a named reason (decision Q1). A symlinked folder on the way to the target or the
//! receipt is refused.

use crate::apply::{publish_exclusive, publish_replace, read_opt};
use serde::{Deserialize, Serialize};
use serde_json::{json, Value};
use std::{
    fs,
    path::{Path, PathBuf},
};

type Result<T> = std::result::Result<T, String>;

const SCHEMA_VERSION: u32 = 1;
/// Seconds between launchd runs (the shell loop's 90 s).
pub const START_INTERVAL: u32 = 90;
/// The cron schedule (cron's finest step above one minute that does not overlap a 90 s pass).
pub const CRON_SCHEDULE: &str = "*/2 * * * *";
const SCHEDULE_DIR: &str = "schedule";
const CRON_FILE: &str = "gate-loop.crontab";
const LOCK_FILE: &str = "gate-loop.lock";
const LOG_FILE: &str = "logs/gate-loop.log";
const LAUNCH_AGENTS: &str = "Library/LaunchAgents";
/// Named variables a scheduler entry must never carry.
const TOKEN_NAMES: [&str; 3] = ["GH_TOKEN", "GITHUB_TOKEN", "MOAI_AUDIT_APP_TOKEN"];
const TOKEN_SUFFIXES: [&str; 3] = ["_TOKEN", "_KEY", "_SECRET"];
/// GitHub token value prefixes.
const TOKEN_VALUES: [&str; 5] = ["ghp_", "gho_", "ghs_", "ghu_", "github_pat_"];
const WINDOWS_REASON: &str = "windows_not_supported: Windows Task Scheduler entries are out of \
                              scope (#2906 Q1); use --os macos or --os linux";

/// What the entry is made from; every value is given, so the entry is deterministic.
#[derive(Clone, Debug)]
pub struct Request {
    pub home: PathBuf,
    /// `macos` | `linux` (`windows` is refused with a named reason).
    pub os: String,
    /// `OWNER/REPO`.
    pub repo: String,
    /// Absolute path of the `moai` binary the entry runs.
    pub moai_bin: PathBuf,
    /// Absolute repository folder the pass runs in (`--dir` and the working folder).
    pub dir: PathBuf,
    /// The entry's PATH.
    pub path: String,
    /// Extra `NAME=VALUE` variables for the entry.
    pub env: Vec<String>,
    /// A merging entry instead of the dry run.
    pub merge: bool,
    /// The operator's permission reference the merging entry carries.
    pub merge_permission_reference: Option<String>,
}

#[derive(Clone, Debug, Serialize, Deserialize, PartialEq)]
#[serde(deny_unknown_fields)]
pub struct Plan {
    pub schema_version: u32,
    pub os: String,
    pub home: String,
    pub label: String,
    pub target_path: String,
    /// create | replace | unchanged
    pub action: String,
    /// The sha256 of the target's bytes the plan was made from; none when absent.
    pub expected_sha256: Option<String>,
    pub entry: String,
    pub entry_sha256: String,
    pub mode: String,
    /// Commands the operator runs to turn the entry on (never run here).
    pub enable: Vec<String>,
    /// Commands the operator runs to turn it off and remove it (never run here).
    pub disable: Vec<String>,
    pub plan_digest: String,
}

#[derive(Clone, Debug, Serialize, Deserialize, PartialEq)]
#[serde(deny_unknown_fields)]
pub struct Receipt {
    pub schema_version: u32,
    pub id: String,
    pub plan_digest: String,
    pub permission_reference: String,
    /// pending | committed
    pub status: String,
    pub target_path: String,
    pub before_sha256: Option<String>,
    pub after_sha256: String,
}

#[derive(Clone, Copy, Debug, PartialEq, Eq)]
enum Os {
    Macos,
    Linux,
}

fn parse_os(os: &str) -> Result<Os> {
    match os {
        "macos" => Ok(Os::Macos),
        "linux" => Ok(Os::Linux),
        "windows" => Err(WINDOWS_REASON.into()),
        _ => Err(format!("unknown --os {os:?}; use macos or linux")),
    }
}

fn sha(bytes: &[u8]) -> String {
    crate::digest(bytes)
}

/// Whether a variable name is token-like: a named GitHub/MoAI token, or any name ending in
/// `_TOKEN`, `_KEY` or `_SECRET` (case-insensitive).
pub fn is_token_name(name: &str) -> bool {
    let up = name.trim().to_ascii_uppercase();
    TOKEN_NAMES.contains(&up.as_str()) || TOKEN_SUFFIXES.iter().any(|s| up.ends_with(s))
}

fn token_refusal(name: &str) -> String {
    format!("token_variable: {name} must not be in a scheduler entry; gh reads its own login")
}

/// A value that may go into the entry: no control character, and none of `%` (cron's newline)
/// or `'` (the cron line's quoting).
fn plain_value(what: &str, v: &str) -> Result<()> {
    if v.is_empty() || v.chars().any(|c| c.is_control() || c == '%' || c == '\'') {
        return Err(format!(
            "{what} is empty or has a control character, % or '"
        ));
    }
    if TOKEN_VALUES.iter().any(|p| v.contains(p)) {
        return Err(token_refusal(what));
    }
    Ok(())
}

fn absolute(what: &str, p: &Path) -> Result<String> {
    let s = p.to_str().ok_or(format!("{what} is not UTF-8"))?;
    if !p.is_absolute() {
        return Err(format!("{what} must be an absolute path"));
    }
    plain_value(what, s)?;
    Ok(s.to_string())
}

/// The entry's PATH: absolute folders only, and no `NAME=` assignment smuggled in.
fn check_path(path: &str) -> Result<()> {
    for part in path.split(':') {
        if let Some((name, _)) = part.split_once('=') {
            let name = name.rsplit('/').next().unwrap_or(name);
            if is_token_name(name) {
                return Err(token_refusal(name));
            }
            return Err(format!(
                "--path entry {part:?} is an assignment, not a folder"
            ));
        }
        if !part.starts_with('/') {
            return Err(format!("--path entry {part:?} is not an absolute folder"));
        }
    }
    plain_value("--path", path)
}

/// `NAME=VALUE` pairs: a plain name, not PATH, not token-like.
fn check_env(env: &[String]) -> Result<Vec<(String, String)>> {
    let mut out: Vec<(String, String)> = Vec::new();
    for pair in env {
        let (name, value) = pair
            .split_once('=')
            .ok_or(format!("--env {pair:?} is not NAME=VALUE"))?;
        if is_token_name(name) {
            return Err(token_refusal(name));
        }
        let ok_name = !name.is_empty()
            && !name.starts_with(|c: char| c.is_ascii_digit())
            && name.chars().all(|c| c.is_ascii_alphanumeric() || c == '_');
        if !ok_name || name == "PATH" || out.iter().any(|(n, _)| n == name) {
            return Err(format!(
                "--env name {name:?} is not a plain, new name (PATH is --path)"
            ));
        }
        plain_value(name, value)?;
        out.push((name.to_string(), value.to_string()));
    }
    Ok(out)
}

fn split_repo(repo: &str) -> Result<(&str, &str)> {
    let (owner, name) = repo.split_once('/').ok_or("--repo must be OWNER/REPO")?;
    let part_ok = |s: &str| {
        !s.is_empty()
            && !s.starts_with('.')
            && s.chars()
                .all(|c| c.is_ascii_alphanumeric() || "._-".contains(c))
    };
    if !part_ok(owner) || !part_ok(name) {
        return Err("--repo must be OWNER/REPO of letters, digits, '.', '_' or '-'".into());
    }
    Ok((owner, name))
}

/// The launchd label, also the plist's file name.
pub fn label(owner: &str, repo: &str) -> String {
    format!("moai.gate-loop.{owner}.{repo}")
}

/// The `gate-loop` arguments after the binary. Never `--review-confirmed`.
fn loop_args(repo: &str, dir: &str, merge_reference: Option<&str>) -> Vec<String> {
    let mut args = vec!["gate-loop".to_string(), "--once".into()];
    match merge_reference {
        Some(r) => args.extend(["--permission-reference".into(), r.to_string()]),
        None => args.push("--dry-run".into()),
    }
    args.extend([
        "--repo".into(),
        repo.to_string(),
        "--dir".into(),
        dir.to_string(),
    ]);
    args
}

fn xml(s: &str) -> String {
    s.replace('&', "&amp;")
        .replace('<', "&lt;")
        .replace('>', "&gt;")
}

/// The LaunchAgent plist.
fn render_plist(
    label: &str,
    program: &[String],
    dir: &str,
    env: &[(String, String)],
    log: &str,
) -> String {
    let mut o = String::new();
    o.push_str("<?xml version=\"1.0\" encoding=\"UTF-8\"?>\n");
    o.push_str("<!DOCTYPE plist PUBLIC \"-//Apple//DTD PLIST 1.0//EN\" \"http://www.apple.com/DTDs/PropertyList-1.0.dtd\">\n");
    o.push_str("<plist version=\"1.0\">\n<dict>\n");
    o.push_str(&format!(
        "\t<key>Label</key>\n\t<string>{}</string>\n",
        xml(label)
    ));
    o.push_str("\t<key>ProgramArguments</key>\n\t<array>\n");
    for a in program {
        o.push_str(&format!("\t\t<string>{}</string>\n", xml(a)));
    }
    o.push_str("\t</array>\n");
    o.push_str(&format!(
        "\t<key>WorkingDirectory</key>\n\t<string>{}</string>\n",
        xml(dir)
    ));
    o.push_str("\t<key>EnvironmentVariables</key>\n\t<dict>\n");
    for (k, v) in env {
        o.push_str(&format!(
            "\t\t<key>{}</key>\n\t\t<string>{}</string>\n",
            xml(k),
            xml(v)
        ));
    }
    o.push_str("\t</dict>\n");
    o.push_str(&format!(
        "\t<key>StartInterval</key>\n\t<integer>{START_INTERVAL}</integer>\n"
    ));
    for key in ["StandardOutPath", "StandardErrorPath"] {
        o.push_str(&format!(
            "\t<key>{key}</key>\n\t<string>{}</string>\n",
            xml(log)
        ));
    }
    o.push_str("</dict>\n</plist>\n");
    o
}

/// A word as the shell reads it: bare when it holds only safe characters, else single-quoted
/// (`'` itself is refused earlier).
fn sh(word: &str) -> String {
    if !word.is_empty()
        && word
            .chars()
            .all(|c| c.is_ascii_alphanumeric() || "/._-:,+=@".contains(c))
    {
        word.to_string()
    } else {
        format!("'{word}'")
    }
}

/// The crontab file: a comment naming the entry and the one line. The variables (PATH first)
/// prefix the command, so no `NAME=` line changes the user's other cron entries.
fn render_cron(
    label: &str,
    repo: &str,
    program: &[String],
    dir: &str,
    env: &[(String, String)],
    lock: &str,
    log: &str,
) -> String {
    let vars: Vec<String> = env.iter().map(|(k, v)| format!("{k}={}", sh(v))).collect();
    let words: Vec<String> = program.iter().map(|w| sh(w)).collect();
    format!(
        "# moai gate-loop {repo} ({label})\n{CRON_SCHEDULE} cd {} && {} flock -n {} {} >> {} 2>&1\n",
        sh(dir),
        vars.join(" "),
        sh(lock),
        words.join(" "),
        sh(log),
    )
}

/// A path that must be a real folder (or absent): never a symlink, never a file.
fn plain_dir_or_absent(path: &Path) -> Result<()> {
    match fs::symlink_metadata(path) {
        Ok(m) if m.file_type().is_symlink() => {
            Err(format!("{} is a symlink; refused", path.display()))
        }
        Ok(m) if !m.is_dir() => Err(format!("{} is not a folder", path.display())),
        Ok(_) => Ok(()),
        Err(e) if e.kind() == std::io::ErrorKind::NotFound => Ok(()),
        Err(_) => Err(format!("{} is unreadable", path.display())),
    }
}

/// Every folder from HOME (exclusive) down to `path`'s parent must be plain or absent.
fn folders_plain(home: &Path, path: &Path) -> Result<()> {
    let rel = path
        .parent()
        .and_then(|p| p.strip_prefix(home).ok())
        .ok_or("target is outside HOME")?;
    let mut dir = home.to_path_buf();
    for part in rel.components() {
        dir.push(part);
        plain_dir_or_absent(&dir)?;
    }
    Ok(())
}

/// The current bytes of the target: `Ok(None)` when absent; a symlink or a folder in its place
/// is refused, never followed.
fn current_file(path: &Path) -> Result<Option<Vec<u8>>> {
    match fs::symlink_metadata(path) {
        Ok(m) if m.file_type().is_symlink() => {
            Err(format!("{} is a symlink; refused", path.display()))
        }
        Ok(m) if !m.is_file() => Err(format!("{} is not a regular file", path.display())),
        Ok(_) => read_opt(path),
        Err(e) if e.kind() == std::io::ErrorKind::NotFound => Ok(None),
        Err(_) => Err(format!("{} is unreadable", path.display())),
    }
}

fn schedule_dir(home: &Path) -> PathBuf {
    crate::home_root::path(home).join(SCHEDULE_DIR)
}

/// Read-only preview of the entry; writes nothing.
pub fn plan(req: &Request) -> Result<Plan> {
    let os = parse_os(&req.os)?;
    let home = fs::canonicalize(&req.home).map_err(|_| "home directory unavailable")?;
    let home_s = home.to_str().ok_or("non UTF-8 home path")?.to_string();
    plain_value("home", &home_s)?;
    let (owner, name) = split_repo(&req.repo)?;
    let moai = absolute("--moai-bin", &req.moai_bin)?;
    let dir = absolute("--dir", &req.dir)?;
    check_path(&req.path)?;
    let mut env = vec![("PATH".to_string(), req.path.clone())];
    env.extend(check_env(&req.env)?);
    let merge_reference = match (req.merge, req.merge_permission_reference.as_deref()) {
        (true, Some(r)) if crate::contract::permission_reference_ok(r) => {
            plain_value("--merge-permission-reference", r)?;
            Some(r)
        }
        (true, _) => return Err("--merge needs --merge-permission-reference".into()),
        (false, Some(_)) => {
            return Err("--merge-permission-reference is only for --merge".into());
        }
        (false, None) => None,
    };
    let label = label(owner, name);
    let state = crate::home_root::path(&home);
    let log = state
        .join(LOG_FILE)
        .to_str()
        .ok_or("non UTF-8 state path")?
        .to_string();
    let mut program = vec![moai.clone()];
    program.extend(loop_args(&req.repo, &dir, merge_reference));
    let (target, entry, enable, disable) = match os {
        Os::Macos => {
            let target = home.join(LAUNCH_AGENTS).join(format!("{label}.plist"));
            let t = target.to_str().ok_or("non UTF-8 target")?.to_string();
            let entry = render_plist(&label, &program, &dir, &env, &log);
            let logs = sh(&state.join("logs").display().to_string());
            (
                target,
                entry,
                vec![
                    "# stop target/leader/auto_loop.sh (the shell loop) first".to_string(),
                    format!("mkdir -p {logs}"),
                    format!("launchctl bootstrap gui/$(id -u) {}", sh(&t)),
                ],
                vec![
                    format!("launchctl bootout gui/$(id -u)/{label}"),
                    format!("rm {}", sh(&t)),
                ],
            )
        }
        Os::Linux => {
            let target = schedule_dir(&home).join(CRON_FILE);
            let t = sh(target.to_str().ok_or("non UTF-8 target")?);
            let lock = schedule_dir(&home).join(LOCK_FILE).display().to_string();
            let entry = render_cron(&label, &req.repo, &program, &dir, &env, &lock, &log);
            let logs = sh(&state.join("logs").display().to_string());
            (
                target,
                entry,
                vec![
                    "# stop target/leader/auto_loop.sh (the shell loop) first".to_string(),
                    format!("mkdir -p {logs}"),
                    format!("(crontab -l 2>/dev/null; cat {t}) | crontab -"),
                ],
                vec![format!("crontab -l | grep -vxF -f {t} | crontab -")],
            )
        }
    };
    folders_plain(&home, &target)?;
    let receipts = schedule_dir(&home).join("receipts").join("x");
    folders_plain(&home, &receipts)?;
    let before = current_file(&target)?;
    let action = match &before {
        None => "create",
        Some(b) if b.as_slice() == entry.as_bytes() => "unchanged",
        Some(_) => "replace",
    };
    let mut plan = Plan {
        schema_version: SCHEMA_VERSION,
        os: req.os.clone(),
        home: home_s,
        label,
        target_path: target.to_str().ok_or("non UTF-8 target")?.to_string(),
        action: action.into(),
        expected_sha256: before.as_deref().map(sha),
        entry_sha256: sha(entry.as_bytes()),
        entry,
        mode: if merge_reference.is_some() {
            "merge"
        } else {
            "dry-run"
        }
        .into(),
        enable,
        disable,
        plan_digest: String::new(),
    };
    plan.plan_digest = sha(&serde_json::to_vec(&plan).map_err(|_| "encode failed")?);
    Ok(plan)
}

fn write_json<T: Serialize>(path: &Path, value: &T) -> Result<()> {
    let mut bytes = serde_json::to_vec_pretty(value).map_err(|_| "encode failed")?;
    bytes.push(b'\n');
    match read_opt(path)? {
        Some(b) => publish_replace(path, Some(&b), &bytes),
        None => publish_exclusive(path, &bytes),
    }
}

fn new_receipt_dir(home: &Path) -> Result<(String, PathBuf)> {
    let receipts = schedule_dir(home).join("receipts");
    folders_plain(home, &receipts.join("x"))?;
    fs::create_dir_all(&receipts).map_err(|_| "schedule receipt folder unavailable")?;
    folders_plain(home, &receipts.join("x"))?;
    let nanos = std::time::SystemTime::now()
        .duration_since(std::time::UNIX_EPOCH)
        .map(|d| d.as_nanos())
        .unwrap_or(0);
    for n in 0..16 {
        let id = format!("schedule-{nanos}-{n}");
        let dir = receipts.join(&id);
        if fs::create_dir(&dir).is_ok() {
            return Ok((id, dir));
        }
    }
    Err("schedule receipt id unavailable".into())
}

/// Writes a previewed entry: `plan_digest` must equal a fresh plan's digest. The previous
/// bytes go into the receipt first. Runs no scheduler command.
pub fn apply(req: &Request, plan_digest: &str, permission: &str) -> Result<Value> {
    if !crate::contract::permission_reference_ok(permission) {
        return Err("permission reference required".into());
    }
    let fresh = plan(req)?;
    if fresh.plan_digest != plan_digest {
        return Err("stale or altered plan; re-preview required".into());
    }
    let out = json!({
        "plan_digest": fresh.plan_digest,
        "target_path": fresh.target_path,
        "enable": fresh.enable,
        "disable": fresh.disable,
    });
    if fresh.action == "unchanged" {
        let mut out = out;
        out["state"] = json!("unchanged");
        out["written"] = json!([]);
        return Ok(out);
    }
    let home = PathBuf::from(&fresh.home);
    let target = PathBuf::from(&fresh.target_path);
    let folder = target.parent().ok_or("target folder")?;
    // Checked before creating (create_dir_all would follow a linked parent) and after.
    folders_plain(&home, &target)?;
    fs::create_dir_all(folder).map_err(|_| format!("{} unavailable", folder.display()))?;
    folders_plain(&home, &target)?;
    let before = current_file(&target)?;
    if before.as_deref().map(sha) != fresh.expected_sha256 {
        return Err(format!(
            "{} changed since the plan; re-preview required",
            target.display()
        ));
    }
    let (id, receipt_dir) = new_receipt_dir(&home)?;
    if let Some(b) = &before {
        publish_exclusive(&receipt_dir.join("target.before"), b)?;
    }
    let mut receipt = Receipt {
        schema_version: SCHEMA_VERSION,
        id: id.clone(),
        plan_digest: fresh.plan_digest.clone(),
        permission_reference: permission.to_string(),
        status: "pending".into(),
        target_path: fresh.target_path.clone(),
        before_sha256: before.as_deref().map(sha),
        after_sha256: fresh.entry_sha256.clone(),
    };
    let receipt_path = receipt_dir.join("receipt.json");
    write_json(&receipt_path, &receipt)?;
    match &before {
        Some(b) => publish_replace(&target, Some(b), fresh.entry.as_bytes())?,
        None => publish_exclusive(&target, fresh.entry.as_bytes())?,
    }
    receipt.status = "committed".into();
    write_json(&receipt_path, &receipt)?;
    let mut out = out;
    out["state"] = json!("applied");
    out["receipt"] = json!(id);
    out["written"] = json!([fresh.target_path]);
    Ok(out)
}

#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    fn token_names_at_their_edges() {
        for n in [
            "GH_TOKEN",
            "GITHUB_TOKEN",
            "MOAI_AUDIT_APP_TOKEN",
            "X_TOKEN",
            "a_key",
            "S_SECRET",
        ] {
            assert!(is_token_name(n), "{n}");
        }
        // The suffix needs its underscore; a near name passes.
        for n in [
            "TOKEN",
            "KEY",
            "MONKEY",
            "GH_CONFIG_DIR",
            "TOKENS",
            "SECRET_DIR",
            "PATH",
        ] {
            assert!(!is_token_name(n), "{n}");
        }
    }

    #[test]
    fn loop_args_never_confirm_a_review() {
        for r in [None, Some("ref-1")] {
            let a = loop_args("o/r", "/d", r);
            assert!(!a.iter().any(|w| w.contains("review")), "{a:?}");
            assert_eq!(&a[..2], ["gate-loop", "--once"]);
            assert_eq!(a.contains(&"--dry-run".to_string()), r.is_none());
        }
    }

    #[test]
    fn os_names() {
        assert_eq!(parse_os("macos"), Ok(Os::Macos));
        assert_eq!(parse_os("linux"), Ok(Os::Linux));
        assert!(parse_os("windows")
            .unwrap_err()
            .starts_with("windows_not_supported"));
        assert!(parse_os("Linux").is_err());
    }

    #[test]
    fn shell_words_are_quoted_when_needed() {
        assert_eq!(sh("/usr/bin"), "/usr/bin");
        assert_eq!(sh("/a b"), "'/a b'");
        assert_eq!(sh("$(x)"), "'$(x)'");
        assert!(plain_value("v", "100%").is_err());
        assert!(plain_value("v", "it's").is_err());
        assert!(plain_value("v", "ghp_abc").is_err());
    }

    #[test]
    fn path_needs_absolute_folders() {
        assert!(check_path("/usr/bin:/bin").is_ok());
        assert!(check_path("/usr/bin::/bin").is_err());
        assert!(check_path("bin:/bin").is_err());
        assert!(check_path("GH_TOKEN=x:/bin")
            .unwrap_err()
            .starts_with("token_variable"));
    }
}
