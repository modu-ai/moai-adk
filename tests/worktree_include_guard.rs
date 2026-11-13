//! #2670 AC-2 (Rust): nothing this product ships or embeds is a `.worktreeinclude`, and no
//! source writes one. The Go generation shipped one listing `.env` and `.env.*`; the Rust
//! product only reports such a file (`go-state-plan`), so a template carrying it would
//! reintroduce the copy of secrets into every worktree.

use std::{
    fs,
    path::{Path, PathBuf},
};

const ROOT: &str = env!("CARGO_MANIFEST_DIR");

fn files(dir: &Path, out: &mut Vec<PathBuf>) {
    for e in fs::read_dir(dir).unwrap() {
        let p = e.unwrap().path();
        let m = fs::symlink_metadata(&p).unwrap();
        if m.is_dir() {
            files(&p, out);
        } else {
            out.push(p);
        }
    }
}

/// AC-3: the UPDATE route tells the user about the warning and never edits the file.
#[test]
fn the_update_workflow_reports_the_warning_and_never_edits_the_file() {
    let text = fs::read_to_string(Path::new(ROOT).join("skills/moai/workflows/update.md")).unwrap();
    for line in [
        "run the read-only `moai go-state-plan <project>` and show the user every warning it gives about `.worktreeinclude`",
        "selects secret files such as `.env`",
        "copy them into every new worktree",
        "`moai-impl` and `moai-audit`",
        "never edit that file for them",
    ] {
        assert!(text.contains(line), "update.md lacks: {line}");
    }
}

#[test]
fn no_shipped_or_embedded_file_is_a_worktreeinclude() {
    let root = Path::new(ROOT);
    // `templates/` holds every `include_str!` target the binary embeds; `skills/moai/` is the
    // installed skill package.
    let mut shipped = Vec::new();
    for dir in ["templates", "skills/moai"] {
        files(&root.join(dir), &mut shipped);
    }
    assert!(
        shipped.len() > 30,
        "scan found too few files: {}",
        shipped.len()
    );
    let named: Vec<_> = shipped
        .iter()
        .filter(|p| p.file_name().is_some_and(|n| n == ".worktreeinclude"))
        .collect();
    assert!(named.is_empty(), "shipped .worktreeinclude: {named:?}");

    // Embedded files are named in `include_str!`; no source may embed or name the file.
    let mut sources = Vec::new();
    files(&root.join("src"), &mut sources);
    assert!(
        sources.len() > 50,
        "scan found too few sources: {}",
        sources.len()
    );
    for src in sources
        .iter()
        .filter(|p| p.extension().is_some_and(|e| e == "rs"))
    {
        let text = fs::read_to_string(src).unwrap();
        for (i, line) in text.lines().enumerate() {
            if let Some(at) = line.find("include_str!(") {
                let arg = &line[at..];
                assert!(
                    !arg.contains(".worktreeinclude"),
                    "{}:{} embeds a .worktreeinclude",
                    src.display(),
                    i + 1
                );
            }
        }
        // Only the read-only checker may name the file.
        if src.file_name().is_some_and(|n| n != "worktree_include.rs") {
            assert!(
                !text.contains(".worktreeinclude"),
                "{} names .worktreeinclude; only src/worktree_include.rs reads it",
                src.display()
            );
        }
    }
}
