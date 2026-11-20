//! `check-ready` (SPEC-20261120-shipped-workflows-commands, slice 2): the declared pre-ready
//! checks of a project, in one word. It reads only what the project declares — the SPEC
//! documents under its `.moai/specs/` folder and the `test_paths` its profile declares — and
//! never a git range nor a test run: selecting checks from a change range is `check`
//! (`moai_cli::check_affected`)'s job, so the two do not overlap. The SPEC parser and the
//! checks are `moai_cli::spec_doc`; the profile and its declared-paths validation are
//! `moai_cli::read_profile` (which already refuses a malformed declaration).

use moai_cli::spec_doc::{self as sd, SPECS_DIR};
use serde_json::{json, Value};
use std::fs;
use std::path::PathBuf;

/// Runs the declared pre-ready checks of a project
///
/// `moai check-ready` checks what the project declares, in the folder `--project` names
/// (default the current one): every SPEC document under its `.moai/specs/` folder — a
/// project without that folder reports `specs absent` and checks nothing — and every path
/// its profile declares under `test_paths` (a declared path that does not exist in the
/// project is a problem; a profile that declares one is itself invalid). It reads no git
/// range and runs no test, so it does not overlap `moai check affected`, which selects its
/// checks from a change range. A project with neither a profile nor a SPEC folder answers
/// `nothing declared` and passes. The answer is one JSON envelope; exit 0 only when nothing
/// is wrong, and the refusal names each problem on stderr.
#[derive(clap::Args)]
pub struct Args {
    /// The project folder.
    #[arg(long, default_value = ".")]
    project: PathBuf,
}

pub const OPERATION: &str = "check-ready";

pub fn run(args: Args) -> Result<Value, String> {
    let root = args.project;
    let mut problems: Vec<String> = Vec::new();

    // The profile's declared test paths. `read_profile` validates the declaration itself, so
    // a declaration that walks out of the project is an invalid profile, not a missing path.
    let (declared, profile) = match moai_cli::read_profile(&root) {
        Ok((p, _)) => (p.test_paths.unwrap_or_default(), "present"),
        Err(e) if e == "project profile missing" => (Vec::new(), "absent"),
        Err(e) => {
            problems.push(format!("profile: invalid profile ({e})"));
            (Vec::new(), "invalid")
        }
    };
    let mut missing: Vec<String> = Vec::new();
    for d in &declared {
        let path = root.join(d.trim());
        match fs::symlink_metadata(&path) {
            Ok(m) if m.is_dir() || m.is_file() => {}
            _ => {
                missing.push(d.clone());
                problems.push(format!("declared test path missing: {d}"));
            }
        }
    }

    // The SPEC documents.
    let (specs_root, present) =
        sd::specs_dir(&root).map_err(|e| format!("check-ready refused: {e}"))?;
    let mut spec_folders = 0usize;
    let mut spec_errors = 0usize;
    let mut spec_warnings = 0usize;
    if present {
        let specs = specs_root.join(SPECS_DIR);
        let mut names: Vec<(String, PathBuf)> = Vec::new();
        let entries = fs::read_dir(&specs)
            .map_err(|_| "check-ready refused: the SPEC folder could not be listed".to_string())?;
        for e in entries {
            let e = e.map_err(|_| {
                "check-ready refused: the SPEC folder could not be listed".to_string()
            })?;
            let path = e.path();
            let meta = fs::symlink_metadata(&path).map_err(|_| {
                "check-ready refused: a SPEC folder could not be examined".to_string()
            })?;
            if meta.is_dir() || meta.file_type().is_symlink() {
                names.push((e.file_name().to_string_lossy().into_owned(), path));
            }
        }
        names.sort();
        spec_folders = names.len();
        for (name, dir) in &names {
            let doc = sd::load(dir, false);
            spec_errors += doc.errors();
            spec_warnings += doc.warnings();
            for f in doc
                .findings
                .iter()
                .filter(|f| f.severity == sd::Severity::Error)
            {
                let line = f.line.map(|l| format!(":{l}")).unwrap_or_default();
                problems.push(format!("{SPECS_DIR}/{name}/spec.md{line} {}", f.code));
            }
        }
    }

    let nothing_declared = profile == "absent" && !present;
    if !problems.is_empty() {
        return Err(format!(
            "check-ready refused: {} problem(s): {}",
            problems.len(),
            problems.join("; ")
        ));
    }
    Ok(json!({
        "action": OPERATION,
        "project": root,
        "profile": profile,
        "declared_paths": declared.len(),
        "missing_paths": missing.len(),
        "specs": if present { "present" } else { "absent" },
        "spec_folders": spec_folders,
        "spec_errors": spec_errors,
        "spec_warnings": spec_warnings,
        "nothing_declared": nothing_declared,
        "result": "PASS",
    }))
}
