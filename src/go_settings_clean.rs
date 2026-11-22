//! The Go-generation `.moai/` references in `.claude/settings.json` and their clean-up
//! (Issue #2936 slice L2a, leader decisions D1 = A and D2 = C on the Issue).
//!
//! A Go-generation project's `.claude/settings.json` names its own `.moai/` folder in three keys:
//! `statusLine` and `subagentStatusLine` (a command that runs `.moai/status_line.sh`) and
//! `plansDirectory` (`.moai/plans`). While they stand, `go-state-plan` keeps the Go `.moai/`
//! folder in place ("settings reference .moai/"), and nothing else in the update changes them.
//!
//! This module recognises exactly the values the Go tool wrote — anything different by even one
//! character is the user's and is never touched — and plans, applies and reverts their removal
//! (`go-settings-plan`, `go-settings-apply`, `go-settings-revert`; the shape follows
//! [`crate::go_hooks`]). `plansDirectory` and `subagentStatusLine` go; `statusLine` becomes a
//! command that runs the Rust launcher by its direct path (`"$HOME/.moai/path/moai-cli"
//! statusline`, or `.moai-cli` for a HOME that keeps the earlier folder, [`launcher_shown`]; not
//! the PATH name `moai`, which a routing shim may send to the Go build). That
//! write is planned only after the launcher is shown to be the Rust build; when it cannot be
//! shown the plan is `held`, writes nothing and says what to check.
//!
//! Only a project whose state is the Go generation's (`state_root` verdict `go`) has Go-made
//! keys: once `.moai/` is this generation's folder, a `plansDirectory` naming it is not a
//! leftover. Only `.claude/settings.json` is written, never `.claude/settings.local.json`. The files under
//! `.moai/plans` are counted for the preview and never touched: they move with the whole `.moai/`
//! folder in the go-state backup. A preview names keys, never a value.

use crate::Result;
use serde_json::{json, Map, Value};
use std::fs;
use std::path::{Path, PathBuf};

const SETTINGS: &str = ".claude/settings.json";

/// The Go tool's plans folder, the only `plansDirectory` value taken for its own.
const PLANS_DIR: &str = ".moai/plans";

/// The launcher as a preview shows it when no HOME is known (the status line is then never
/// written: the plan is blocked on `home_unavailable`).
pub const LAUNCHER_SHOWN: &str = "$HOME/.moai-cli/path/moai-cli";

/// The Rust launcher as the status line command spells it (Claude Code's shell expands `$HOME`)
/// and as every preview shows it: under the folder the HOME resolves to
/// ([`crate::home_root::path`], SPEC-20261121-path-root-moai-default REQ-11).
pub fn launcher_shown(home: &Path) -> String {
    let root = crate::home_root::path(home);
    let rel = root
        .strip_prefix(home)
        .map(|r| r.to_string_lossy().into_owned())
        .unwrap_or_else(|_| crate::home_root::dir(home).to_string());
    format!("$HOME/{rel}/path/moai-cli")
}

/// The most directory entries the `.moai/plans` count looks at (#2936 L2a, D1 condition 2): the
/// walk stops at the next one and the count is reported as `more_than_limit`.
pub const PLANS_WALK_LIMIT: usize = 1000;

const RECEIPT_PREFIX: &str = "go-settings-";

/// What the preview and the apply say of the old file's kept copy (#3111 audit N2).
const KEPT_COPY_NOTE: &str =
    "the old .claude/settings.json is kept as before/settings.json in the \
    receipt folder (receipts/go-settings-<id>/ under the project state folder); it can hold \
    secrets, so it is owner-only (0600 in a 0700 folder) and a .gitignore in that folder keeps \
    it out of git";

/// Owner-only bits for the receipt folder and the kept copy of the settings file, which may hold
/// a token in its `env` block (#3111 audit B1).
const PRIVATE_DIR: u32 = 0o700;
const PRIVATE_FILE: u32 = 0o600;

/// Create one folder (not its parents); it is 0700 from its creation.
#[cfg(unix)]
fn create_private_dir(path: &Path) -> std::io::Result<()> {
    use std::os::unix::fs::DirBuilderExt;
    let mut builder = fs::DirBuilder::new();
    builder.mode(PRIVATE_DIR);
    builder.create(path)
}

#[cfg(not(unix))]
fn create_private_dir(path: &Path) -> std::io::Result<()> {
    fs::create_dir(path)
}

/// The kept copy lives in a folder whose own `.gitignore` is `*`: a nested ignore file outranks
/// the project's, so `git add -A` never offers the copy, whether or not the managed block of the
/// root `.gitignore` exists yet (#3111 audit B2: this is the first step of the backup phase).
fn keep_receipt_out_of_git(dir: &Path) -> Result<()> {
    crate::apply::publish_exclusive_mode(&dir.join(".gitignore"), b"*\n", Some(PRIVATE_FILE))
        .map_err(|_| "receipt .gitignore refused".into())
}

/// What a Go key becomes.
#[derive(Clone, Copy, Debug, PartialEq, Eq)]
pub enum Change {
    /// The key goes.
    Remove,
    /// The key stays with the Rust launcher command.
    Replace,
}

impl Change {
    fn name(self) -> &'static str {
        match self {
            Change::Remove => "remove",
            Change::Replace => "replace",
        }
    }
}

/// One Go-made key and what the apply does with it.
#[derive(Clone, Copy, Debug, PartialEq, Eq)]
pub struct GoKey {
    pub key: &'static str,
    pub change: Change,
}

/// The `command` values the Go settings template (`settings.json.tmpl`) wrote for the wrapper
/// over its history: the bare path (the first releases), behind `$CLAUDE_PROJECT_DIR`, quoted
/// whole (the October 2026 template, also for `subagentStatusLine`), and the Windows `bash "…"`.
const GO_STATUS_COMMANDS: [&str; 4] = [
    ".moai/status_line.sh",
    "$CLAUDE_PROJECT_DIR/.moai/status_line.sh",
    "\"$CLAUDE_PROJECT_DIR/.moai/status_line.sh\"",
    "bash \"$CLAUDE_PROJECT_DIR/.moai/status_line.sh\"",
];

/// The `refreshInterval` values that template wrote (10 until August 2026, 60 after).
const GO_REFRESH: [u64; 2] = [10, 60];

/// Whether `v` is exactly the status line object the Go tool wrote: `type` `command`, one of
/// [`GO_STATUS_COMMANDS`], and at most a `refreshInterval` of [`GO_REFRESH`] besides — nothing
/// else. Any other value, key or spelling is the user's.
fn is_go_status_line(v: &Value) -> bool {
    let Some(o) = v.as_object() else {
        return false;
    };
    o.get("type") == Some(&json!("command"))
        && o.get("command")
            .and_then(Value::as_str)
            .is_some_and(|c| GO_STATUS_COMMANDS.contains(&c))
        && o.keys()
            .all(|k| matches!(k.as_str(), "type" | "command" | "refreshInterval"))
        && o.get("refreshInterval")
            .is_none_or(|r| GO_REFRESH.iter().any(|n| *r == json!(n)))
}

/// Whether any object in `bytes` repeats a key. Parsing into a map keeps only the last one, so a
/// rewrite would merge them silently (#3111 audit N1).
fn repeats_a_key(bytes: &[u8]) -> bool {
    use serde::de::{Deserialize, Deserializer, Error, MapAccess, SeqAccess, Visitor};
    struct Strict;
    impl<'de> Deserialize<'de> for Strict {
        fn deserialize<D: Deserializer<'de>>(d: D) -> std::result::Result<Self, D::Error> {
            struct V;
            impl<'de> Visitor<'de> for V {
                type Value = Strict;
                fn expecting(&self, f: &mut std::fmt::Formatter) -> std::fmt::Result {
                    f.write_str("any JSON value")
                }
                fn visit_bool<E>(self, _: bool) -> std::result::Result<Strict, E> {
                    Ok(Strict)
                }
                fn visit_i64<E>(self, _: i64) -> std::result::Result<Strict, E> {
                    Ok(Strict)
                }
                fn visit_u64<E>(self, _: u64) -> std::result::Result<Strict, E> {
                    Ok(Strict)
                }
                fn visit_f64<E>(self, _: f64) -> std::result::Result<Strict, E> {
                    Ok(Strict)
                }
                fn visit_str<E>(self, _: &str) -> std::result::Result<Strict, E> {
                    Ok(Strict)
                }
                fn visit_unit<E>(self) -> std::result::Result<Strict, E> {
                    Ok(Strict)
                }
                fn visit_seq<A: SeqAccess<'de>>(
                    self,
                    mut a: A,
                ) -> std::result::Result<Strict, A::Error> {
                    while a.next_element::<Strict>()?.is_some() {}
                    Ok(Strict)
                }
                fn visit_map<A: MapAccess<'de>>(
                    self,
                    mut a: A,
                ) -> std::result::Result<Strict, A::Error> {
                    let mut seen = std::collections::HashSet::new();
                    while let Some(k) = a.next_key::<String>()? {
                        if !seen.insert(k) {
                            return Err(A::Error::custom("repeated key"));
                        }
                        a.next_value::<Strict>()?;
                    }
                    Ok(Strict)
                }
            }
            d.deserialize_any(V)
        }
    }
    serde_json::from_slice::<Strict>(bytes).is_err()
}

/// The names of the top-level keys that hold a number which writing the file again would spell
/// differently (#3111 audit B4): a number whose text is not what `serde_json` prints for the
/// value it reads (an integer beyond 64 bits that becomes a float, `-0`, `1E2`, `1.10`). Names
/// only. `bytes` is JSON that already parsed as an object.
fn numbers_that_change(bytes: &[u8]) -> Vec<String> {
    struct Scan<'a> {
        b: &'a [u8],
        i: usize,
        hits: Vec<String>,
    }
    impl Scan<'_> {
        fn ws(&mut self) {
            while matches!(self.b.get(self.i), Some(b' ' | b'\t' | b'\n' | b'\r')) {
                self.i += 1;
            }
        }
        fn string(&mut self) -> String {
            let start = self.i;
            self.i += 1;
            while let Some(c) = self.b.get(self.i) {
                self.i += 1;
                match c {
                    b'\\' => self.i += 1,
                    b'"' => break,
                    _ => {}
                }
            }
            serde_json::from_slice::<String>(self.b.get(start..self.i).unwrap_or_default())
                .unwrap_or_default()
        }
        fn value(&mut self, top: Option<&str>, depth: usize) {
            self.ws();
            match self.b.get(self.i) {
                Some(b'"') => {
                    self.string();
                }
                Some(&open @ (b'{' | b'[')) => {
                    self.i += 1;
                    loop {
                        self.ws();
                        match self.b.get(self.i) {
                            None => return,
                            Some(b'}' | b']') => {
                                self.i += 1;
                                return;
                            }
                            Some(b',') => self.i += 1,
                            Some(_) => {
                                let mut key = top.map(String::from);
                                if open == b'{' {
                                    let k = self.string();
                                    self.ws();
                                    self.i += 1; // the colon
                                    if depth == 0 {
                                        key = Some(k);
                                    }
                                }
                                self.value(key.as_deref(), depth + 1);
                            }
                        }
                    }
                }
                Some(b't' | b'f' | b'n') => {
                    while self.b.get(self.i).is_some_and(u8::is_ascii_alphabetic) {
                        self.i += 1;
                    }
                }
                Some(_) => {
                    let start = self.i;
                    while self.b.get(self.i).is_some_and(|c| {
                        c.is_ascii_digit() || matches!(c, b'.' | b'e' | b'E' | b'+' | b'-')
                    }) {
                        self.i += 1;
                    }
                    let token = std::str::from_utf8(self.b.get(start..self.i).unwrap_or_default())
                        .unwrap_or_default();
                    let same = serde_json::from_str::<Value>(token)
                        .map(|v| {
                            let printed = v.to_string();
                            printed == token
                        })
                        .unwrap_or(false);
                    if !same {
                        if let Some(name) = top {
                            if !self.hits.iter().any(|h| h == name) {
                                self.hits.push(name.to_string());
                            }
                        }
                    }
                }
                None => {}
            }
        }
    }
    let mut scan = Scan {
        b: bytes,
        i: 0,
        hits: Vec::new(),
    };
    scan.value(None, 0);
    scan.hits.sort();
    scan.hits
}

/// The Go-made keys of a settings object (sorted by key name) and, as `kept`, the names of the
/// other top-level keys whose value still names `.moai/` — the user's own values, never touched.
pub fn classify(map: &Map<String, Value>) -> (Vec<GoKey>, Vec<String>) {
    let mut go = Vec::new();
    if map.get("plansDirectory") == Some(&json!(PLANS_DIR)) {
        go.push(GoKey {
            key: "plansDirectory",
            change: Change::Remove,
        });
    }
    if map.get("statusLine").is_some_and(is_go_status_line) {
        go.push(GoKey {
            key: "statusLine",
            change: Change::Replace,
        });
    }
    if map.get("subagentStatusLine").is_some_and(is_go_status_line) {
        go.push(GoKey {
            key: "subagentStatusLine",
            change: Change::Remove,
        });
    }
    let mut kept: Vec<String> = map
        .iter()
        .filter(|(k, v)| {
            !go.iter().any(|g| g.key == k.as_str()) && v.to_string().contains(".moai/")
        })
        .map(|(k, _)| k.clone())
        .collect();
    kept.sort();
    (go, kept)
}

/// The status line object a project gets: the Rust launcher by its direct path.
pub fn rust_status_line(home: &Path) -> Value {
    json!({
        "type": "command",
        "command": format!("\"{}\" {}", launcher_shown(home), crate::status_line::SUBCOMMAND),
    })
}

/// What the Rust launcher under HOME is.
#[derive(Clone, Copy, Debug, PartialEq, Eq)]
pub enum LauncherState {
    /// `statusline --help` printed the Rust build's first line.
    Rust,
    /// Not an executable regular file (or a link that does not resolve to one).
    Missing,
    /// It ran and printed another first line.
    NotRust,
    /// It printed nothing within the probe limit, or it lies inside the project (never run).
    Unprobed,
    /// No Go-made `statusLine` is to be replaced, so the launcher was not looked at.
    NotNeeded,
}

impl LauncherState {
    pub fn name(self) -> &'static str {
        match self {
            LauncherState::Rust => "rust",
            LauncherState::Missing => "missing",
            LauncherState::NotRust => "not_rust",
            LauncherState::Unprobed => "unprobed",
            LauncherState::NotNeeded => "not_needed",
        }
    }
}

fn launcher_path(home: &Path) -> PathBuf {
    crate::home_root::path(home).join("path").join("moai-cli")
}

/// Looks at `path/moai-cli` under the folder the HOME resolves to (REQ-12): it must be a file, or a link that resolves to one,
/// with an execute bit; it is then run as `statusline --help` by the safe probe the status line
/// gate uses (HOME the only variable, a fixed PATH, the system temporary folder as working
/// folder, stdin closed, output capped, killed on time-out) and must print
/// [`crate::status_line::RUST_HELP_LINE`] as its first line. A launcher that resolves to a path
/// inside the project is never run.
pub fn launcher_state(root: &Path, home: &Path) -> LauncherState {
    let path = launcher_path(home);
    let Ok(meta) = fs::metadata(&path) else {
        return LauncherState::Missing;
    };
    if !meta.is_file() {
        return LauncherState::Missing;
    }
    #[cfg(unix)]
    {
        use std::os::unix::fs::PermissionsExt as _;
        if meta.permissions().mode() & 0o111 == 0 {
            return LauncherState::Missing;
        }
    }
    match fs::canonicalize(&path) {
        Ok(real) if !real.starts_with(root) => {}
        _ => return LauncherState::Unprobed,
    }
    match crate::install::probe_args_env(
        &path,
        &[crate::status_line::SUBCOMMAND, "--help"],
        &[("HOME", home)],
    ) {
        None => LauncherState::Unprobed,
        Some(line) if line == crate::status_line::RUST_HELP_LINE => LauncherState::Rust,
        Some(_) => LauncherState::NotRust,
    }
}

/// How many regular files `.moai/plans` holds, from a bounded walk that never follows a link.
#[derive(Clone, Copy, Debug, PartialEq, Eq)]
pub struct PlansCount {
    /// `.moai/plans` is a real folder under a real `.moai`.
    pub exists: bool,
    /// Regular files counted.
    pub files: usize,
    /// The walk reached [`PLANS_WALK_LIMIT`] entries and stopped; there are more than `files`.
    pub more_than_limit: bool,
}

/// Counts the regular files under the project's `.moai/plans`, looking at no more than
/// [`PLANS_WALK_LIMIT`] entries (files, folders and links alike). A link is neither followed nor
/// counted; a link at `.moai` or at `.moai/plans` counts as no folder.
pub fn count_plans(root: &Path) -> PlansCount {
    let none = PlansCount {
        exists: false,
        files: 0,
        more_than_limit: false,
    };
    let mut dir = root.to_path_buf();
    for part in PLANS_DIR.split('/') {
        dir.push(part);
        match fs::symlink_metadata(&dir) {
            Ok(m) if m.is_dir() && !m.file_type().is_symlink() => {}
            _ => return none,
        }
    }
    let (mut files, mut seen) = (0usize, 0usize);
    let mut stack = vec![dir];
    while let Some(d) = stack.pop() {
        let Ok(entries) = fs::read_dir(&d) else {
            continue;
        };
        for e in entries.flatten() {
            seen += 1;
            if seen > PLANS_WALK_LIMIT {
                return PlansCount {
                    exists: true,
                    files,
                    more_than_limit: true,
                };
            }
            let Ok(m) = fs::symlink_metadata(e.path()) else {
                continue;
            };
            if m.file_type().is_symlink() {
                continue;
            }
            if m.is_dir() {
                stack.push(e.path());
            } else if m.is_file() {
                files += 1;
            }
        }
    }
    PlansCount {
        exists: true,
        files,
        more_than_limit: false,
    }
}

/// What `go-settings-plan` worked out: the printed plan, the settings bytes it read and the
/// bytes the apply would write (`None` unless the action is `replace`).
struct Planned {
    json: Value,
    before: Option<Vec<u8>>,
    after: Option<Vec<u8>>,
    go: Vec<GoKey>,
}

/// The settings object with the Go-made keys removed or replaced; every other key keeps its
/// value and its place.
fn rewrite(map: &Map<String, Value>, go: &[GoKey], status_line: &Value) -> Result<Vec<u8>> {
    let mut out = map.clone();
    for g in go {
        match g.change {
            Change::Remove => {
                out.shift_remove(g.key);
            }
            Change::Replace => {
                out.insert(g.key.to_string(), status_line.clone());
            }
        }
    }
    let mut bytes = serde_json::to_vec_pretty(&out).map_err(|_| "encode failed")?;
    bytes.push(b'\n');
    Ok(bytes)
}

fn plan_inner(project: &Path, home: Option<&Path>) -> Result<Planned> {
    let root = crate::hooks::project_root(project, home)?;
    if fs::symlink_metadata(root.join(".claude")).is_ok_and(|m| m.file_type().is_symlink()) {
        return Err(".claude is a link; nothing is followed".into());
    }
    let before = crate::claude::read_existing(&root.join(SETTINGS))?;
    let map = match &before {
        None => None,
        Some(b) => match serde_json::from_slice::<Value>(b) {
            Ok(Value::Object(m)) => Some(m),
            _ => return Err(format!("{SETTINGS} is not a JSON object")),
        },
    };
    // Only a project whose state is the Go generation's has Go-made keys: after the move `.moai/`
    // is this generation's own folder and a `plansDirectory` that names it is not a leftover.
    let generation = crate::state_root::classify(&root).generation;
    let (go, kept) = match &map {
        Some(m) if generation == "go" => classify(m),
        _ => Default::default(),
    };
    let needs_launcher = go.iter().any(|g| g.change == Change::Replace);
    let repeats = !go.is_empty() && before.as_deref().is_some_and(repeats_a_key);
    let spelled_differently: Vec<String> = match (&before, go.is_empty()) {
        (Some(b), false) => numbers_that_change(b),
        _ => Vec::new(),
    };
    let shown = home.map_or_else(|| LAUNCHER_SHOWN.to_string(), launcher_shown);
    let launcher = match (needs_launcher, home) {
        (false, _) => LauncherState::NotNeeded,
        (true, Some(h)) => launcher_state(&root, h),
        (true, None) => LauncherState::Unprobed,
    };

    let mut blocks: Vec<&str> = Vec::new();
    let mut held: Option<(String, String)> = None;
    if needs_launcher && home.is_none() {
        blocks.push("home_unavailable");
    } else if repeats {
        held = Some((
            format!("{SETTINGS} repeats a key; writing it again would merge the repeats into one"),
            format!(
                "remove the repeated key from {SETTINGS} yourself, then run `moai go-settings-plan \
                 <project>` again; nothing was written and the Go keys stay"
            ),
        ));
    } else if !spelled_differently.is_empty() {
        held = Some((
            format!(
                "{SETTINGS} holds numbers that writing it again would spell differently or round \
                 (keys: {}); the values would change",
                spelled_differently.join(", ")
            ),
            format!(
                "change those numbers in {SETTINGS} yourself (write them as the plain value, for \
                 example a string for a very large id), then run `moai go-settings-plan \
                 <project>` again; nothing was written and the Go keys stay"
            ),
        ));
    } else if needs_launcher && launcher != LauncherState::Rust {
        let why = match launcher {
            LauncherState::Missing => {
                format!("the Rust launcher {shown} is missing or is not an executable file")
            }
            LauncherState::NotRust => format!(
                "{shown} answers `statusline --help` as something other than the Rust build"
            ),
            _ => format!(
                "{shown} did not answer `statusline --help` within the probe limit (or \
                 lies inside the project, where it is never run)"
            ),
        };
        let next = format!(
            "run `\"{shown}\" statusline --help`: its first line must read \"{}\"; \
             install or repair the Rust launcher if it does not, then run `moai go-settings-plan \
             <project>` again; nothing was written and the Go keys stay",
            crate::status_line::RUST_HELP_LINE
        );
        held = Some((why, next));
    }
    let action = if go.is_empty() {
        "none"
    } else if !blocks.is_empty() {
        "blocked"
    } else if held.is_some() {
        "held"
    } else {
        "replace"
    };

    let after = match (&map, action) {
        (Some(m), "replace") => match home {
            Some(h) => Some(rewrite(m, &go, &rust_status_line(h))?),
            None => None,
        },
        _ => None,
    };
    let go_json: Vec<Value> = go
        .iter()
        .map(|g| json!({"key": g.key, "change": g.change.name()}))
        .collect();
    let bound = json!({
        "settings_sha256": before.as_deref().map(crate::digest),
        "state_root": generation,
        "launcher": launcher.name(),
        "go_keys": go_json,
        "action": action,
        "result_sha256": after.as_deref().map(crate::digest),
    });

    let yours: Vec<&str> = ["plansDirectory", "statusLine", "subagentStatusLine"]
        .into_iter()
        .filter(|k| {
            map.as_ref().is_some_and(|m| m.contains_key(*k))
                && !go.iter().any(|g| g.key == *k)
                && !kept.iter().any(|n| n == k)
        })
        .collect();
    let plans_files = if go.iter().any(|g| g.key == "plansDirectory") {
        let c = count_plans(&root);
        json!({
            "folder": PLANS_DIR,
            "exists": c.exists,
            "count": c.files,
            "more_than_limit": c.more_than_limit,
            "limit": PLANS_WALK_LIMIT,
            "deleted": false,
            "note": "these files are not deleted; they stay in .moai/ and move with it into the \
                     go-state backup",
        })
    } else {
        Value::Null
    };
    let mut p = json!({
        "settings": SETTINGS,
        "settings_present": before.is_some(),
        "state_root": generation,
        "action": action,
        "go_keys": go_json,
        "kept": kept,
        "launcher": {"path": shown, "state": launcher.name()},
        "plans_files": plans_files,
        "local_settings": "never written",
        "kept_copy": if go.is_empty() {
            Value::Null
        } else {
            json!(KEPT_COPY_NOTE)
        },
        "formatting": if go.is_empty() {
            Value::Null
        } else {
            json!("the whole file is written again as JSON with 2-space indentation and LF line \
                   ends: every key keeps its value and its place, but spacing and escapes may change")
        },
        "yours": yours,
        "blocks": blocks,
        "project": root.display().to_string(),
        "writes": false,
    });
    if !kept.is_empty() {
        p["kept_note"] = json!(
            "these keys hold values that differ from what the Go tool wrote, so they are yours \
             and stay; they still name .moai/"
        );
    }
    if let Some((reason, next)) = &held {
        p["reason"] = json!(reason);
        p["next"] = json!(next);
    }
    p["plan_digest"] = json!(crate::digest(
        &serde_json::to_vec(&bound).map_err(|_| "encode failed")?
    ));
    p["next_actions"] = json!(match action {
        "replace" => vec![String::from(
            "review the keys that go and the one that changes, then run `moai go-settings-apply \
             --plan-digest <plan_digest> --permission-reference <ref> <project>`"
        )],
        "held" => held.iter().map(|(_, n)| n.clone()).collect(),
        _ => Vec::new(),
    });
    Ok(Planned {
        json: p,
        before,
        after,
        go,
    })
}

/// What cleaning the Go-made `.moai/` references out of `project`'s `.claude/settings.json`
/// would do. Reads only (and runs the launcher's `statusline --help` when a Go-made `statusLine`
/// is to be replaced). `action` is `replace`, `none` (no Go-made key), `held` (the Rust launcher
/// could not be shown; nothing would be written, `reason` and `next` say why and what to run)
/// or `blocked` (`blocks` says why). `plan_digest` binds the settings bytes, the launcher state
/// and the result.
pub fn plan(project: &Path, home: Option<&Path>) -> Result<Value> {
    plan_inner(project, home).map(|p| p.json)
}

/// The record of one clean-up, under `<state>/receipts/go-settings-<id>/`.
#[derive(serde::Serialize, serde::Deserialize)]
struct Receipt {
    schema_version: u32,
    id: String,
    plan_digest: String,
    permission_reference: String,
    status: String,
    /// sha256 of `.claude/settings.json` before and after.
    settings_before: Option<String>,
    settings_after: Option<String>,
    /// The keys the apply removed or replaced.
    keys: Vec<String>,
}

fn receipts_dir(root: &Path) -> PathBuf {
    crate::state_root::path(root, "receipts")
}

fn new_receipt_dir(root: &Path) -> Result<(String, PathBuf)> {
    let receipts = receipts_dir(root);
    crate::claude::refuse_symlink_dir(&crate::state_root::path(root, ""))?;
    fs::create_dir_all(&receipts).map_err(|_| "receipt folder unavailable")?;
    crate::claude::refuse_symlink_dir(&receipts)?;
    let nanos = std::time::SystemTime::now()
        .duration_since(std::time::UNIX_EPOCH)
        .map(|d| d.as_nanos())
        .unwrap_or(0);
    for n in 0..16 {
        let id = format!("{RECEIPT_PREFIX}{nanos}-{n}");
        let dir = receipts.join(&id);
        if create_private_dir(&dir).is_ok() {
            return Ok((id, dir));
        }
    }
    Err("receipt id unavailable".into())
}

fn write_receipt(dir: &Path, r: &Receipt) -> Result<()> {
    let mut bytes = serde_json::to_vec_pretty(r).map_err(|_| "encode failed")?;
    bytes.push(b'\n');
    let path = dir.join("receipt.json");
    let before = crate::claude::read_existing(&path)?;
    crate::apply::publish_replace(&path, before.as_deref(), &bytes)
}

/// Cleans the Go-made references out of `.claude/settings.json` after the user confirmed the
/// plan whose digest is `plan_digest`. The plan is made again first: a stale digest, a `held` or
/// `blocked` plan, or a missing permission reference writes nothing. The old bytes are kept as
/// `before/settings.json` in the receipt folder before the file is replaced; when the replace
/// fails the file is as it was and the receipt reads `rolled_back`.
pub fn apply(
    project: &Path,
    home: Option<&Path>,
    plan_digest: &str,
    permission: &str,
) -> Result<Value> {
    if !crate::contract::permission_reference_ok(permission) {
        return Err("a permission reference is required; nothing was written".into());
    }
    let planned = plan_inner(project, home)?;
    let p = &planned.json;
    if p["plan_digest"].as_str() != Some(plan_digest) {
        return Err("stale or altered plan; re-preview required".into());
    }
    match p["action"].as_str() {
        Some("replace") => {}
        Some("none") => return Ok(json!({"state": "unchanged", "writes": []})),
        Some("held") => {
            return Err(format!(
                "the plan is held: {}; nothing was written",
                p["reason"]
                    .as_str()
                    .unwrap_or("the Rust launcher is not shown")
            ))
        }
        _ => return Err(format!("the plan is blocked: {}", p["blocks"])),
    }
    let root = crate::hooks::project_root(project, home)?;
    let settings_path = root.join(SETTINGS);
    let (Some(before), Some(after)) = (planned.before.as_deref(), planned.after.as_deref()) else {
        return Err("the plan has no settings bytes; nothing was written".into());
    };

    let (id, dir) = new_receipt_dir(&root)?;
    keep_receipt_out_of_git(&dir)?;
    let keep = dir.join("before");
    create_private_dir(&keep).map_err(|_| "receipt folder unavailable")?;
    crate::apply::publish_exclusive_mode(&keep.join("settings.json"), before, Some(PRIVATE_FILE))
        .map_err(|_| "receipt write failed")?;
    let mut receipt = Receipt {
        schema_version: 1,
        id: id.clone(),
        plan_digest: plan_digest.to_string(),
        permission_reference: permission.to_string(),
        status: "pending".into(),
        settings_before: Some(crate::digest(before)),
        settings_after: None,
        keys: planned.go.iter().map(|g| g.key.to_string()).collect(),
    };
    write_receipt(&dir, &receipt)?;

    if let Err(e) = crate::apply::publish_replace(&settings_path, Some(before), after) {
        // A failed replace leaves the file as it was; if it is not, put the bytes back.
        let now = crate::claude::read_existing(&settings_path)?;
        if now.as_deref() != Some(before) {
            restore(&settings_path, now.as_deref(), before)?;
        }
        receipt.status = "rolled_back".into();
        write_receipt(&dir, &receipt)?;
        return Err(format!(
            "{SETTINGS} was not written ({e}); everything was put back"
        ));
    }
    receipt.settings_after = Some(crate::digest(after));
    receipt.status = "applied".into();
    write_receipt(&dir, &receipt)?;
    let names = |c: Change| -> Vec<&str> {
        planned
            .go
            .iter()
            .filter(|g| g.change == c)
            .map(|g| g.key)
            .collect()
    };
    Ok(json!({
        "state": "applied",
        "project": root.display().to_string(),
        "removed": names(Change::Remove),
        "replaced": names(Change::Replace),
        "plans_files": p["plans_files"],
        "receipt": id,
        "kept_copy": KEPT_COPY_NOTE,
        "host_loaded": "not_run",
        "next_actions": [format!("to undo: moai go-settings-revert --receipt {id} --permission-reference <ref> <project>")],
    }))
}

/// Puts the settings file back to `bytes` from `now`.
fn restore(path: &Path, now: Option<&[u8]>, bytes: &[u8]) -> Result<()> {
    crate::apply::publish_replace(path, now, bytes)
}

/// Undoes one clean-up while `.claude/settings.json` still holds the bytes the apply wrote: the
/// file gets its before bytes back. Any other state of the file writes nothing.
pub fn revert(project: &Path, receipt_id: &str, permission: &str) -> Result<Value> {
    if !crate::contract::permission_reference_ok(permission) {
        return Err("a permission reference is required; nothing was written".into());
    }
    if !(receipt_id.starts_with(RECEIPT_PREFIX)
        && receipt_id.len() <= 64
        && receipt_id
            .bytes()
            .all(|b| b.is_ascii_alphanumeric() || b == b'-'))
    {
        return Err("unknown receipt id".into());
    }
    let root = fs::canonicalize(project).map_err(|_| "project directory unavailable")?;
    let dir = receipts_dir(&root).join(receipt_id);
    for d in [
        crate::state_root::path(&root, ""),
        receipts_dir(&root),
        dir.clone(),
    ] {
        crate::claude::refuse_symlink_dir(&d)?;
    }
    let bytes =
        crate::claude::read_existing(&dir.join("receipt.json"))?.ok_or("receipt not found")?;
    let mut r: Receipt = serde_json::from_slice(&bytes).map_err(|_| "receipt is unreadable")?;
    if r.status == "pending" {
        return Err(
            "the receipt is pending: an apply was interrupted; compare .claude/settings.json with \
             before/settings.json in the receipt folder by hand; nothing was written"
                .into(),
        );
    }
    if r.status != "applied" {
        return Err(format!("the receipt is {}; nothing to revert", r.status));
    }
    // Nothing is written through a link: `.claude` must be a real folder.
    match fs::symlink_metadata(root.join(".claude")) {
        Ok(m) if m.is_dir() && !m.file_type().is_symlink() => {}
        _ => return Err(".claude is a link or not a folder; nothing was written".into()),
    }
    let settings_path = root.join(SETTINGS);
    let now = crate::claude::read_existing(&settings_path)?;
    if now.as_deref().map(crate::digest) != r.settings_after {
        return Err(format!(
            "{SETTINGS} changed since the apply; nothing was written"
        ));
    }
    let want = r.settings_before.clone().ok_or("receipt is unreadable")?;
    let old = fs::read(dir.join("before").join("settings.json"))
        .map_err(|_| "before bytes missing from the receipt")?;
    if crate::digest(&old) != want {
        return Err("before bytes do not match the receipt".into());
    }
    restore(&settings_path, now.as_deref(), &old)?;
    r.status = "reverted".into();
    write_receipt(&dir, &r)?;
    Ok(json!({"state": "reverted", "receipt": receipt_id, "keys_restored": r.keys}))
}

#[cfg(test)]
mod tests {
    use super::*;

    fn status(command: &str) -> Value {
        json!({"type": "command", "command": command})
    }

    fn map(v: Value) -> Map<String, Value> {
        v.as_object().cloned().unwrap()
    }

    /// Every spelling and shape the Go templates wrote is Go-made (#2936 L2a, template history).
    #[test]
    fn the_go_template_spellings_are_go_made() {
        let with = |c: &str, r: Option<u64>| {
            let mut v = status(c);
            if let Some(r) = r {
                v["refreshInterval"] = json!(r);
            }
            v
        };
        for c in GO_STATUS_COMMANDS {
            for r in [None, Some(10), Some(60)] {
                let m = map(json!({"statusLine": with(c, r), "subagentStatusLine": with(c, r)}));
                let (go, kept) = classify(&m);
                assert_eq!(go.len(), 2, "{c} {r:?}");
                assert!(kept.is_empty(), "{c} {r:?}");
            }
        }
        for c in [
            "$CLAUDE_PROJECT_DIR/.moai/status_line.sh",
            "\"$CLAUDE_PROJECT_DIR/.moai/status_line.sh\"",
            ".moai/status_line.sh",
        ] {
            let m = map(json!({"statusLine": status(c), "subagentStatusLine": status(c)}));
            let (go, kept) = classify(&m);
            assert_eq!(
                go.iter().map(|g| (g.key, g.change)).collect::<Vec<_>>(),
                [
                    ("statusLine", Change::Replace),
                    ("subagentStatusLine", Change::Remove)
                ],
                "{c}"
            );
            assert!(kept.is_empty(), "{c}");
        }
    }

    /// One character more, less or different, another shape, or another key is the user's.
    #[test]
    fn a_value_one_character_off_is_kept_by_key_name() {
        let ok = "$CLAUDE_PROJECT_DIR/.moai/status_line.sh";
        for off in [
            status(&format!("{ok} ")),
            status(&format!(" {ok}")),
            status(&format!("{ok}2")),
            status(&format!("bash {ok}")),
            status("$CLAUDE_PROJECT_DIR/.moai/status_line.sh --x"),
            status("./.moai/status_line.sh"),
            // Spellings the Go template never wrote for a status line.
            status("${CLAUDE_PROJECT_DIR}/.moai/status_line.sh"),
            status("\"$CLAUDE_PROJECT_DIR\"/.moai/status_line.sh"),
            status("bash $CLAUDE_PROJECT_DIR/.moai/status_line.sh"),
            // A refreshInterval the template never wrote is the user's edit.
            json!({"type": "command", "command": ok, "refreshInterval": 61}),
            json!({"type": "command", "command": ok, "refreshInterval": "60"}),
            json!({"type": "command", "command": ok, "refreshInterval": 0}),
            status("$CLAUDE_PROJECT_DIR/.moai/status-line.sh"),
            status("$CLAUDE_PROJECT_DIR/.moai/Status_line.sh"),
            status("$CLAUDE_PROJECT_DIR//.moai/status_line.sh"),
            status("CLAUDE_PROJECT_DIR/.moai/status_line.sh"),
            status("$HOME/.moai/status_line.sh"),
            json!({"type": "command", "command": ok, "padding": 1}),
            json!({"type": "command", "command": ok, "timeout": 5}),
            json!({"type": "Command", "command": ok}),
            json!({"command": ok}),
            json!(ok),
        ] {
            let m = map(json!({"statusLine": off, "subagentStatusLine": off}));
            let (go, kept) = classify(&m);
            assert!(go.is_empty(), "{off}");
            assert_eq!(kept, ["statusLine", "subagentStatusLine"], "{off}");
        }
        for off in [
            ".moai/plans/",
            ".moai/plan",
            ".moai/Plans",
            "./.moai/plans",
            " .moai/plans",
            ".moai/plans ",
            ".moai/plans/x",
        ] {
            let (go, kept) = classify(&map(json!({"plansDirectory": off})));
            assert!(go.is_empty(), "{off}");
            assert_eq!(kept, ["plansDirectory"], "{off}");
        }
        let (go, kept) = classify(&map(json!({"plansDirectory": ".moai/plans"})));
        assert_eq!(go.len(), 1);
        assert!(kept.is_empty());
        // The key order of the object does not matter for the equality.
        let swapped: Value =
            serde_json::from_str(&format!("{{\"command\":\"{ok}\",\"type\":\"command\"}}"))
                .unwrap();
        assert_eq!(classify(&map(json!({"statusLine": swapped}))).0.len(), 1);
    }

    /// A key that does not name `.moai/` is neither Go-made nor kept.
    #[test]
    fn keys_that_do_not_name_the_folder_are_not_listed() {
        let m = map(json!({
            "permissions": {"deny": ["Read(./.env)"]},
            "env": {"X": "1"},
            "statusLine": status("moai statusline"),
            "plansDirectory": "my-own-plans",
            "other": {"path": "x/.moai/y"},
        }));
        let (go, kept) = classify(&m);
        assert!(go.is_empty());
        assert_eq!(kept, ["other"]);
    }

    #[test]
    fn the_rewrite_keeps_every_other_key_in_place() {
        let m = map(json!({
            "a": 1,
            "plansDirectory": ".moai/plans",
            "b": {"x": [1, 2]},
            "statusLine": status("$CLAUDE_PROJECT_DIR/.moai/status_line.sh"),
            "c": "z",
            "subagentStatusLine": status("$CLAUDE_PROJECT_DIR/.moai/status_line.sh"),
            "d": null,
        }));
        let (go, _) = classify(&m);
        let home = tempfile::tempdir().unwrap();
        fs::create_dir(home.path().join(crate::HOME_STATE_DIR)).unwrap();
        let line = rust_status_line(home.path());
        let out: Value = serde_json::from_slice(&rewrite(&m, &go, &line).unwrap()).unwrap();
        let keys: Vec<&String> = out.as_object().unwrap().keys().collect();
        assert_eq!(keys, ["a", "b", "statusLine", "c", "d"]);
        assert_eq!(out["statusLine"], line);
        assert_eq!(
            line["command"],
            json!("\"$HOME/.moai-cli/path/moai-cli\" statusline")
        );
        assert_eq!(out["b"], m["b"]);
    }
}
