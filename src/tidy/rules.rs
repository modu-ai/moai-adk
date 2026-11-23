//! The tables of the tidy operations (SPEC-20261124-home-tidy-slice2): the class of a top-level
//! entry of the home moai folder, the folder kinds that grow and their thresholds.
//!
//! A class is decided from facts, never from a name alone (parent D-2): the file type comes
//! first (a link, a pipe or a device is `protected`), then a secret-shaped name, then the
//! tables below. A name in no table is `unknown` and never offered.

use crate::home_state::{self, EntryClass};
use serde_json::{json, Map, Value};
use std::collections::BTreeMap;

/// Days without any change after which an owned entry is `owned_stale`.
pub const STALE_DAYS: u64 = 30;

/// Folders of the product itself, with no tidy rule in this slice.
pub const OWNED: &[&str] = &[
    "reports",
    "logs",
    "evidence",
    "worktrees",
    "build",
    "receipts",
    "handoffs",
    "backup",
    "archive",
    "holding",
    "projects",
];

/// Entries of other products in the same folder: listed, never changed (parent Non-goals).
pub const OTHER_PRODUCT: &[&str] = &[
    "claude-profiles",
    "gpt-appserver",
    "local-plugins",
    "mcp",
    "studio",
];

/// Entries whose writer nobody knows (parent D-10): `unknown`, listed, never offered.
pub const UNKNOWN_WRITER: &[&str] = &["state", "logs", "dev-login.sh"];
pub const UNKNOWN_WRITER_PREFIX: &[&str] = &["thread-media"];

/// Go-era names that `home_state::SELECTIVE_TABLE` lacks (the leader's inventory, note 5902).
pub const RETIRED_EXTRA: &[&str] = &["db"];

/// What a top-level entry is, before its age is read.
#[derive(Clone, Copy, PartialEq, Eq, Debug)]
pub enum Class {
    Owned,
    RetiredTool,
    OtherProduct,
    Unknown,
    Protected,
}

/// The class of a directory entry from its name; `reason` names the table row.
pub fn class_of_name(name: &str) -> (Class, &'static str) {
    let lower = name.to_ascii_lowercase();
    if home_state::is_secret(name) {
        return (Class::Protected, "secret_name");
    }
    let (table, reason) = home_state::table_class(name);
    if table == EntryClass::SecretManual {
        return (Class::Protected, "secret_name");
    }
    if OTHER_PRODUCT.contains(&lower.as_str()) {
        return (Class::OtherProduct, "other_product");
    }
    if UNKNOWN_WRITER.contains(&lower.as_str())
        || UNKNOWN_WRITER_PREFIX.iter().any(|p| lower.starts_with(p))
    {
        return (Class::Unknown, "unknown_writer");
    }
    if table == EntryClass::MoveToBackup || RETIRED_EXTRA.contains(&lower.as_str()) {
        let r = if table == EntryClass::MoveToBackup {
            reason
        } else {
            "go_run_data"
        };
        return (Class::RetiredTool, r);
    }
    if table == EntryClass::KeepInPlace || OWNED.contains(&lower.as_str()) {
        return (Class::Owned, "owned_folder");
    }
    (Class::Unknown, "not_in_table")
}

/// The folder kinds that grow, in the order their rows are listed.
pub const KINDS: &[&str] = &[
    "reports",
    "logs",
    "evidence",
    "worktrees",
    "build",
    "receipts",
];

/// How a kind's items are found below its folder.
#[derive(Clone, Copy, PartialEq, Eq, Debug)]
pub enum ItemMode {
    /// Every regular file at any depth is an item.
    Files,
    /// The children at this depth are items (a file above that depth is an item itself).
    Depth(usize),
}

pub fn item_mode(kind: &str) -> ItemMode {
    match kind {
        "reports" | "logs" => ItemMode::Files,
        "worktrees" => ItemMode::Depth(2),
        _ => ItemMode::Depth(1),
    }
}

/// `(field, default)` of every kind: the proposed values of the parent D-5 table.
fn fields(kind: &str) -> &'static [(&'static str, f64)] {
    match kind {
        "reports" => &[
            ("max_mib", 500.0),
            ("max_files", 2000.0),
            ("archive_after_days", 90.0),
        ],
        "evidence" => &[("max_mib", 2048.0), ("landed_after_days", 30.0)],
        "logs" => &[("file_mib", 4.0), ("keep_copies", 3.0)],
        "build" => &[("max_gib", 50.0), ("free_gib", 40.0)],
        "worktrees" => &[
            ("max_count", 8.0),
            ("max_gib", 30.0),
            ("quiet_minutes", 120.0),
        ],
        "receipts" => &[("max_mib", 200.0)],
        _ => &[],
    }
}

const MIB: f64 = 1_048_576.0;
const GIB: f64 = 1_073_741_824.0;

/// The thresholds of every kind: the defaults, with the caller's `--threshold` values over them.
#[derive(Clone, Debug)]
pub struct Thresholds {
    values: BTreeMap<&'static str, BTreeMap<&'static str, f64>>,
}

impl Thresholds {
    pub fn defaults() -> Self {
        let mut values = BTreeMap::new();
        for kind in KINDS {
            let mut m = BTreeMap::new();
            for (f, d) in fields(kind) {
                m.insert(*f, *d);
            }
            values.insert(*kind, m);
        }
        Thresholds { values }
    }

    /// The defaults with each `<kind>.<field>=<value>` applied; an argument that is not valid
    /// is an error naming it, before anything is read.
    pub fn parse(args: &[String]) -> Result<Self, String> {
        let mut t = Self::defaults();
        for arg in args {
            let bad = |why: &str| format!("--threshold `{arg}`: {why}");
            let (key, value) = arg
                .split_once('=')
                .ok_or_else(|| bad("expected <kind>.<field>=<value>"))?;
            let (kind, field) = key
                .split_once('.')
                .ok_or_else(|| bad("expected <kind>.<field>=<value>"))?;
            let kind = KINDS
                .iter()
                .find(|k| **k == kind)
                .ok_or_else(|| bad(&format!("unknown kind (use {})", KINDS.join(", "))))?;
            let field = fields(kind)
                .iter()
                .map(|(f, _)| *f)
                .find(|f| *f == field)
                .ok_or_else(|| bad(&format!("unknown field of `{kind}`")))?;
            let v: f64 = value
                .parse()
                .map_err(|_| bad("the value is not a number"))?;
            if !v.is_finite() || v < 0.0 {
                return Err(bad("the value must be a number of zero or more"));
            }
            t.values.get_mut(kind).unwrap().insert(field, v);
        }
        Ok(t)
    }

    pub fn get(&self, kind: &str, field: &str) -> f64 {
        self.values
            .get(kind)
            .and_then(|m| m.get(field))
            .copied()
            .unwrap_or(0.0)
    }

    /// The values of one kind as JSON; a whole number is written as an integer.
    pub fn json(&self, kind: &str) -> Value {
        let mut o = Map::new();
        if let Some(m) = self.values.get(kind) {
            for (f, v) in m {
                o.insert((*f).to_string(), number(*v));
            }
        }
        Value::Object(o)
    }

    /// The fields of `kind` that the measures pass, in table order. `free_gib` needs the free
    /// disk space and is not measured in this slice (it is shown, never passed).
    pub fn passed(&self, kind: &str, m: &Measure) -> Vec<&'static str> {
        let mut out = Vec::new();
        for (f, _) in fields(kind) {
            let limit = self.get(kind, f);
            let hit = match (kind, *f) {
                ("reports", "max_mib") | ("evidence", "max_mib") | ("receipts", "max_mib") => {
                    m.bytes as f64 > limit * MIB
                }
                ("reports", "max_files") | ("worktrees", "max_count") => m.count as f64 > limit,
                ("build", "max_gib") | ("worktrees", "max_gib") => m.bytes as f64 > limit * GIB,
                ("logs", "file_mib") => m.largest_item as f64 > limit * MIB,
                _ => false,
            };
            if hit {
                out.push(*f);
            }
        }
        out
    }
}

/// What a kind's items add up to.
#[derive(Clone, Copy, Default, Debug)]
pub struct Measure {
    pub count: u64,
    pub bytes: u64,
    pub largest_item: u64,
}

/// A whole number as an integer, anything else as a float.
pub fn number(v: f64) -> Value {
    if v.fract() == 0.0 && (0.0..9.0e15).contains(&v) {
        json!(v as u64)
    } else {
        json!(v)
    }
}
