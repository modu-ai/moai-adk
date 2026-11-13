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

#[test]
fn the_python_claude_adapter_and_its_tests_are_gone() {
    let root = Path::new(ROOT);
    for rel in REMOVED {
        assert!(
            fs::symlink_metadata(root.join(rel)).is_err(),
            "{rel} should be removed (#2709)"
        );
    }
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
