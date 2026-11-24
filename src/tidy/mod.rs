//! `tidy-inventory` and `tidy-plan` (SPEC-20261124-home-tidy-slice2, parent
//! SPEC-20261122-home-tidy): the read-only half of the home moai folder tidy. The inventory
//! lists every top-level entry of `<HOME>/.moai` with a class, size, newest change and reason;
//! the plan names what could be archived or deleted and why. Neither writes anything, opens a
//! regular file (but the one memory index the caller names), follows a link or starts a command.

pub mod inventory;
pub mod memory;
pub mod plan;
pub mod rules;

use inventory::Limits;
use rules::Thresholds;
use serde_json::{json, Value};
use std::path::{Path, PathBuf};

/// Default entries visited below one top-level entry.
pub const DEFAULT_ENTRY_LIMIT: usize = 200_000;
/// Default seconds for one run.
pub const DEFAULT_TIME_LIMIT_SECS: u64 = 30;

/// What the caller asked: thresholds, memory index paths and the walk limits.
pub struct Options {
    pub thresholds: Thresholds,
    pub memory: Vec<PathBuf>,
    pub limits: Limits,
}

impl Options {
    /// Reads the arguments before any file is read; a bad `--threshold` is an error naming it.
    pub fn new(
        threshold_args: &[String],
        memory: Vec<PathBuf>,
        entry_limit: usize,
        time_limit_secs: u64,
    ) -> Result<Self, String> {
        Ok(Options {
            thresholds: Thresholds::parse(threshold_args)?,
            memory,
            limits: Limits {
                entries: entry_limit,
                seconds: time_limit_secs,
            },
        })
    }
}

fn scan(home: &Path, opts: &Options) -> Result<inventory::Scan, String> {
    let moai = home.join(crate::home_root::MOAI_DIR);
    // The folder itself is read by its own directory data first: a link is refused, never followed.
    match std::fs::symlink_metadata(&moai) {
        Ok(md) if md.file_type().is_symlink() => {
            return Err(format!(
                "`{}` is a link; the tidy operations never follow a link, so nothing was read",
                moai.display()
            ));
        }
        Ok(md) if md.is_dir() => {}
        _ => return Err(format!("no `.moai` folder under `{}`", home.display())),
    }
    inventory::scan(&moai, opts.limits, &opts.thresholds)
}

/// The inventory of `<home>/.moai`.
pub fn run_inventory(home: &Path, opts: &Options) -> Result<Value, String> {
    let scan = scan(home, opts)?;
    Ok(json!({
        "home": home.display().to_string(),
        "moai": scan.moai.display().to_string(),
        "now": scan.now,
        "entries": scan.entries.iter().map(inventory::entry_json).collect::<Vec<_>>(),
        "kinds": scan.kinds.iter().map(|r| inventory::kind_json(r, &opts.thresholds)).collect::<Vec<_>>(),
        "memory_index": memory::rows(&opts.memory),
        "writes": false,
    }))
}

/// The tidy plan of `<home>/.moai`.
pub fn run_plan(home: &Path, opts: &Options) -> Result<Value, String> {
    let scan = scan(home, opts)?;
    let (items, digest) = plan::build(&scan, &opts.thresholds);
    let memory_rows = memory::rows(&opts.memory);
    let tidy_text = memory::tidy_text(&memory_rows);
    let mut out = json!({
        "home": home.display().to_string(),
        "moai": scan.moai.display().to_string(),
        "now": scan.now,
        "items": items,
        "kinds": scan.kinds.iter().map(|r| inventory::kind_json(r, &opts.thresholds)).collect::<Vec<_>>(),
        "digest": digest,
        "memory_index": memory_rows,
        "writes": false,
    });
    if let Some(text) = tidy_text {
        out["memory_tidy"] = text;
    }
    Ok(out)
}
