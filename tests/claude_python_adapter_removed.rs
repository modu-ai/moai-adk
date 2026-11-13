//! #2709 (leader decision 6852250517 on #2682): the provisional Python Claude adapter and its
//! mirror are gone, and nothing in the repository still tells anyone to run them. The Rust
//! `claude-settings-plan/apply` and `moai-cli statusline` replace them.

use std::{
    fs,
    path::{Path, PathBuf},
};

const ROOT: &str = env!("CARGO_MANIFEST_DIR");

const REMOVED: [&str; 9] = [
    "adapters/claude",
    "experiments/claude_settings_plan.py",
    "tests/claude_preview_test.py",
    "tests/claude_recovery_test.py",
    "tests/claude_settings_mirror_parity_test.py",
    "tests/claude_settings_plan_test.py",
    "tests/claude_transaction_boundary_test.py",
    "tests/claude_update_test.py",
    "adapters/claude/output-styles",
];

/// Command lines that ran the removed files.
const RUN_LINES: [&str; 3] = [
    "python3 adapters/claude/",
    "python3 tests/claude_preview_test.py",
    "python3 experiments/claude_settings_plan.py",
];

fn text_files(dir: &Path, out: &mut Vec<PathBuf>) {
    let Ok(entries) = fs::read_dir(dir) else {
        return;
    };
    for e in entries.flatten() {
        let p = e.path();
        let Ok(m) = fs::symlink_metadata(&p) else {
            continue;
        };
        if m.is_dir() {
            text_files(&p, out);
        } else if m.is_file()
            && p.extension().is_some_and(|x| {
                ["md", "rs", "py", "yml", "yaml", "toml", "sh"].contains(&x.to_str().unwrap_or(""))
            })
        {
            out.push(p);
        }
    }
}

/// Leader decision on #2709, option (나): the output styles and their licenses are kept for a
/// later user-layer install (#2638), not deleted.
#[test]
fn the_output_styles_and_their_licenses_are_kept_in_templates() {
    let dir = Path::new(ROOT).join("templates/claude/output-styles");
    for name in [
        "moai.md",
        "moai-easy.md",
        "moai-learn.md",
        "LICENSE",
        "NOTICE.md",
    ] {
        let text = fs::read_to_string(dir.join(name)).unwrap_or_default();
        assert!(
            !text.trim().is_empty(),
            "templates/claude/output-styles/{name} is missing"
        );
    }
    let license = fs::read_to_string(dir.join("LICENSE")).unwrap();
    assert!(
        license.contains("Apache License"),
        "the license text is kept"
    );
}

/// Whether `path` still holds something of the removed adapter. A checkout that once ran the
/// old Python tests keeps the ignored `__pycache__/*.pyc` files, and git never deletes ignored
/// files, so a folder holding only those counts as removed (#2718 audit N1). A symlink, a file
/// or any other content counts as present — a `.pyc` outside `__pycache__` too: Python imports
/// such a sourceless file from its old location (pre-audit 6857452324 B1).
fn still_present(path: &Path) -> bool {
    let Ok(m) = fs::symlink_metadata(path) else {
        return false;
    };
    if !m.is_dir() {
        return true;
    }
    let Ok(entries) = fs::read_dir(path) else {
        return true;
    };
    entries.flatten().any(|e| {
        let p = e.path();
        let is_cache_dir = p.file_name().is_some_and(|n| n == "__pycache__")
            && fs::symlink_metadata(&p).is_ok_and(|m| m.is_dir());
        if is_cache_dir {
            // Only compiled files may sit in a cache folder.
            fs::read_dir(&p).map_or(true, |inner| {
                inner.flatten().any(|f| {
                    !(f.path().extension().is_some_and(|x| x == "pyc")
                        && fs::symlink_metadata(f.path()).is_ok_and(|m| m.is_file()))
                })
            })
        } else {
            still_present(&p)
        }
    })
}

#[test]
fn the_python_claude_adapter_and_its_tests_are_gone() {
    let root = Path::new(ROOT);
    for rel in REMOVED {
        assert!(
            !still_present(&root.join(rel)),
            "{rel} should be removed (#2709)"
        );
    }
}

/// #2718 audit N1: leftover compiled caches do not count; a source file or a stray file does.
#[test]
fn only_leftover_python_caches_count_as_removed() {
    let tmp = tempfile::tempdir().unwrap();
    let dir = tmp.path().join("adapters/claude");
    assert!(!still_present(&dir), "a missing folder is removed");
    fs::create_dir_all(dir.join("__pycache__")).unwrap();
    fs::write(dir.join("__pycache__/install.cpython-313.pyc"), b"x").unwrap();
    assert!(!still_present(&dir), "only a compiled cache is left");
    fs::write(dir.join("__pycache__/notes.txt"), b"x").unwrap();
    assert!(
        still_present(&dir),
        "a non-cache file inside __pycache__ counts"
    );
    fs::remove_file(dir.join("__pycache__/notes.txt")).unwrap();
    fs::write(dir.join("install.py"), b"x").unwrap();
    assert!(still_present(&dir), "a source file counts");
    fs::remove_file(dir.join("install.py")).unwrap();
    fs::create_dir_all(dir.join("output-styles")).unwrap();
    fs::write(dir.join("output-styles/moai.md"), b"x").unwrap();
    assert!(still_present(&dir), "any other file counts");
    let file = tmp.path().join("experiments_plan.py");
    fs::write(&file, b"x").unwrap();
    assert!(still_present(&file), "a removed file that is back counts");
}

/// Pre-audit 6857452324 B1: only a plain `.pyc` file directly inside a real `__pycache__`
/// folder is ignored. A sourceless `.pyc` beside the sources is importable code, and a link in
/// place of a cache file or of the cache folder is not a cache Python wrote.
#[cfg(unix)]
#[test]
fn a_compiled_file_counts_as_removed_only_inside_a_real_cache_folder() {
    let tmp = tempfile::tempdir().unwrap();
    let dir = tmp.path().join("adapters/claude");
    fs::create_dir_all(&dir).unwrap();
    fs::write(dir.join("install.pyc"), b"x").unwrap();
    assert!(
        still_present(&dir),
        "a sourceless .pyc outside __pycache__ counts"
    );
    fs::remove_file(dir.join("install.pyc")).unwrap();

    let real = tmp.path().join("real.pyc");
    fs::write(&real, b"x").unwrap();
    fs::create_dir_all(dir.join("__pycache__")).unwrap();
    std::os::unix::fs::symlink(&real, dir.join("__pycache__/install.cpython-313.pyc")).unwrap();
    assert!(
        still_present(&dir),
        "a link in place of a cache file counts"
    );
    fs::remove_dir_all(dir.join("__pycache__")).unwrap();

    let elsewhere = tmp.path().join("elsewhere");
    fs::create_dir_all(&elsewhere).unwrap();
    fs::write(elsewhere.join("install.cpython-313.pyc"), b"x").unwrap();
    std::os::unix::fs::symlink(&elsewhere, dir.join("__pycache__")).unwrap();
    assert!(
        still_present(&dir),
        "a link in place of the cache folder counts"
    );
}

#[test]
fn nothing_tells_anyone_to_run_the_removed_files() {
    let root = Path::new(ROOT);
    let mut files = Vec::new();
    for dir in [
        "docs",
        "skills",
        "src",
        "tests",
        "tools",
        "templates",
        ".github",
    ] {
        text_files(&root.join(dir), &mut files);
    }
    for top in ["README.md", "AGENTS.md"] {
        let p = root.join(top);
        if p.is_file() {
            files.push(p);
        }
    }
    assert!(
        files.len() > 200,
        "scan found too few files: {}",
        files.len()
    );
    let this = Path::new(file!())
        .file_name()
        .and_then(|n| n.to_str())
        .unwrap_or("");
    for f in &files {
        if f.file_name().and_then(|n| n.to_str()) == Some(this) {
            continue;
        }
        let Ok(text) = fs::read_to_string(f) else {
            continue;
        };
        for line in RUN_LINES {
            assert!(
                !text.contains(line),
                "{} still says `{line}` (#2709 removed it)",
                f.display()
            );
        }
    }
}
