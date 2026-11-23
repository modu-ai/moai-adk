//! The one bounded walk of the home moai folder (SPEC-20261124-home-tidy-slice2, REQ-1 to
//! REQ-3). It reads directory entries and their own metadata only: a regular file is never
//! opened, a link is never followed, and an entry that is secret-shaped, a link or a special
//! file is not even listed below (it is `protected`). Nothing is written.

use super::rules::{self, Class, ItemMode, Measure, Thresholds};
use serde_json::{json, Value};
use std::fs::{self, Metadata};
use std::os::unix::fs::FileTypeExt;
use std::path::{Path, PathBuf};
use std::time::{Instant, SystemTime, UNIX_EPOCH};

/// The limits of one walk.
#[derive(Clone, Copy, Debug)]
pub struct Limits {
    /// Nodes visited below one top-level entry.
    pub entries: usize,
    /// Seconds for the whole run.
    pub seconds: u64,
}

/// What a subtree adds up to.
#[derive(Clone, Copy, Default, Debug)]
pub struct Stats {
    pub bytes: u64,
    pub files: u64,
    pub newest: u64,
}

/// One item of a folder kind (a report file, a build folder, a worktree …).
#[derive(Clone, Debug)]
pub struct Item {
    pub path: String,
    pub bytes: u64,
    pub newest: u64,
}

/// One top-level entry with everything the plan needs.
#[derive(Clone, Debug)]
pub struct Entry {
    pub name: String,
    pub class: String,
    pub reason: String,
    pub file_type: &'static str,
    pub stats: Stats,
    pub over_limit: bool,
    /// The entry is a folder with nothing in it.
    pub empty_dir: bool,
    /// A fact could not be read: the plan keeps the entry.
    pub unreadable: bool,
    pub items: Vec<Item>,
    pub kind: Option<&'static str>,
}

/// One row per folder kind that exists (REQ-3).
#[derive(Clone, Debug)]
pub struct KindRow {
    pub kind: &'static str,
    pub count: u64,
    pub bytes: u64,
    pub oldest: Option<u64>,
    pub newest: Option<u64>,
    pub passed: Vec<&'static str>,
}

pub struct Scan {
    pub moai: PathBuf,
    pub now: u64,
    pub entries: Vec<Entry>,
    pub kinds: Vec<KindRow>,
}

fn secs(t: std::io::Result<SystemTime>) -> u64 {
    t.ok()
        .and_then(|t| t.duration_since(UNIX_EPOCH).ok())
        .map(|d| d.as_secs())
        .unwrap_or(0)
}

fn file_type_name(md: &Metadata) -> &'static str {
    let t = md.file_type();
    if t.is_symlink() {
        "symlink"
    } else if t.is_dir() {
        "dir"
    } else if t.is_file() {
        "file"
    } else if t.is_fifo() {
        "fifo"
    } else if t.is_socket() {
        "socket"
    } else {
        "device"
    }
}

struct Ctx {
    left: usize,
    deadline: Instant,
    over: bool,
    partial: bool,
    mode: Option<ItemMode>,
    items: Vec<Item>,
    root: PathBuf,
}

impl Ctx {
    fn spend(&mut self) -> bool {
        if self.left == 0 || Instant::now() > self.deadline {
            self.over = true;
            return false;
        }
        self.left -= 1;
        true
    }
}

fn merge(a: &mut Stats, b: Stats) {
    a.bytes += b.bytes;
    a.files += b.files;
    a.newest = a.newest.max(b.newest);
}

/// Adds the subtree at `path` (already known as `md`) and returns its stats. `depth` is the
/// distance from the top-level entry.
fn walk(path: &Path, md: &Metadata, depth: usize, ctx: &mut Ctx) -> Stats {
    let mut st = Stats {
        newest: secs(md.modified()),
        ..Stats::default()
    };
    let t = md.file_type();
    if t.is_file() {
        st.bytes = md.len();
        st.files = 1;
    } else if t.is_dir() {
        match fs::read_dir(path) {
            Ok(rd) => {
                let mut kids: Vec<PathBuf> = rd.flatten().map(|e| e.path()).collect();
                kids.sort();
                for kid in kids {
                    if !ctx.spend() {
                        break;
                    }
                    let Ok(kmd) = fs::symlink_metadata(&kid) else {
                        ctx.partial = true;
                        continue;
                    };
                    let s = walk(&kid, &kmd, depth + 1, ctx);
                    merge(&mut st, s);
                }
            }
            Err(_) => ctx.partial = true,
        }
    }
    if let Some(mode) = ctx.mode {
        let is_item = match mode {
            ItemMode::Files => t.is_file(),
            ItemMode::Depth(d) => depth == d || (t.is_file() && depth < d && depth > 0),
        };
        if is_item && depth > 0 {
            let rel = path.strip_prefix(&ctx.root).unwrap_or(path);
            ctx.items.push(Item {
                path: rel.display().to_string(),
                bytes: st.bytes,
                newest: st.newest,
            });
        }
    }
    st
}

/// Walks the home moai folder `moai`.
pub fn scan(moai: &Path, limits: Limits, thresholds: &Thresholds) -> Result<Scan, String> {
    let now = secs(Ok(SystemTime::now()));
    let deadline = Instant::now() + std::time::Duration::from_secs(limits.seconds);
    let mut names: Vec<(String, PathBuf)> = fs::read_dir(moai)
        .map_err(|e| format!("cannot list `{}`: {}", moai.display(), e.kind()))?
        .flatten()
        .map(|e| (e.file_name().to_string_lossy().into_owned(), e.path()))
        .collect();
    names.sort();
    let mut entries = Vec::new();
    for (name, path) in names {
        let Ok(md) = fs::symlink_metadata(&path) else {
            entries.push(unreadable_entry(&name, "unreadable: no metadata"));
            continue;
        };
        let file_type = file_type_name(&md);
        let own_newest = secs(md.modified());
        let (class, why) = rules::class_of_name(&name);
        // Protected by what it is, or by its name: only its own directory data is used.
        if file_type != "file" && file_type != "dir" || class == Class::Protected {
            let reason = if file_type != "file" && file_type != "dir" {
                format!("{file_type}_not_followed")
            } else {
                why.to_string()
            };
            entries.push(Entry {
                name,
                class: "protected".into(),
                reason,
                file_type,
                stats: Stats {
                    bytes: md.len(),
                    files: 0,
                    newest: own_newest,
                },
                over_limit: false,
                empty_dir: false,
                unreadable: false,
                items: Vec::new(),
                kind: None,
            });
            continue;
        }
        let kind = rules::KINDS.iter().copied().find(|k| *k == name);
        let mut ctx = Ctx {
            left: limits.entries,
            deadline,
            over: false,
            partial: false,
            mode: kind.map(rules::item_mode),
            items: Vec::new(),
            root: moai.to_path_buf(),
        };
        let unreadable_top = file_type == "dir" && fs::read_dir(&path).is_err();
        let stats = walk(&path, &md, 0, &mut ctx);
        let empty_dir = file_type == "dir"
            && !unreadable_top
            && fs::read_dir(&path)
                .map(|mut r| r.next().is_none())
                .unwrap_or(false);
        let (class_name, reason) = if unreadable_top {
            (
                "unknown".to_string(),
                "unreadable: cannot list the folder".to_string(),
            )
        } else {
            match class {
                Class::Owned => {
                    let age_days = now.saturating_sub(stats.newest) / 86_400;
                    if age_days <= rules::STALE_DAYS {
                        ("owned_live".into(), format!("changed {age_days} days ago"))
                    } else {
                        (
                            "owned_stale".into(),
                            format!("no change for {age_days} days"),
                        )
                    }
                }
                Class::RetiredTool => ("retired_tool".into(), why.to_string()),
                Class::OtherProduct => ("other_product".into(), why.to_string()),
                Class::Unknown => ("unknown".into(), why.to_string()),
                Class::Protected => unreachable!("handled above"),
            }
        };
        let reason = if ctx.over {
            format!("{reason}; over_limit: the walk stopped, the size is what was seen")
        } else if ctx.partial && !unreadable_top {
            format!("{reason}; part of it is unreadable")
        } else {
            reason
        };
        entries.push(Entry {
            name,
            class: class_name,
            reason,
            file_type,
            stats,
            over_limit: ctx.over,
            empty_dir,
            unreadable: unreadable_top || ctx.partial,
            items: ctx.items,
            kind,
        });
    }
    let kinds = rows(&entries, thresholds);
    Ok(Scan {
        moai: moai.to_path_buf(),
        now,
        entries,
        kinds,
    })
}

fn unreadable_entry(name: &str, reason: &str) -> Entry {
    Entry {
        name: name.to_string(),
        class: "unknown".into(),
        reason: reason.to_string(),
        file_type: "file",
        stats: Stats::default(),
        over_limit: false,
        empty_dir: false,
        unreadable: true,
        items: Vec::new(),
        kind: None,
    }
}

fn rows(entries: &[Entry], thresholds: &Thresholds) -> Vec<KindRow> {
    let mut out = Vec::new();
    for kind in rules::KINDS {
        let Some(e) = entries
            .iter()
            .find(|e| e.kind == Some(*kind) && e.file_type == "dir")
        else {
            continue;
        };
        let m = Measure {
            count: e.items.len() as u64,
            bytes: e.items.iter().map(|i| i.bytes).sum(),
            largest_item: e.items.iter().map(|i| i.bytes).max().unwrap_or(0),
        };
        out.push(KindRow {
            kind,
            count: m.count,
            bytes: m.bytes,
            oldest: e.items.iter().map(|i| i.newest).min(),
            newest: e.items.iter().map(|i| i.newest).max(),
            passed: thresholds.passed(kind, &m),
        });
    }
    out
}

pub fn entry_json(e: &Entry) -> Value {
    let mut v = json!({
        "name": e.name,
        "class": e.class,
        "reason": e.reason,
        "type": e.file_type,
        "bytes": e.stats.bytes,
        "newest": e.stats.newest,
        "over_limit": e.over_limit,
    });
    if e.class != "protected" {
        v["files"] = json!(e.stats.files);
    }
    v
}

pub fn kind_json(r: &KindRow, thresholds: &Thresholds) -> Value {
    json!({
        "kind": r.kind,
        "count": r.count,
        "bytes": r.bytes,
        "oldest": r.oldest,
        "newest": r.newest,
        "thresholds": thresholds.json(r.kind),
        "passed": r.passed,
    })
}
