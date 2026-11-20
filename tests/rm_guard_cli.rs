//! Unit checks for `src/delete_guard.rs` — the pure classification of recursive-delete
//! targets. The module is included by path so the checks run before the wiring piece
//! (the `lib.rs` declaration and the `hook_guards.rs` delegation land with the C-234
//! follow-up slice, per the leader decision on the card).
//!
//! CLI-level checks that drive the real hook path are added by the wiring slice.

#[path = "../src/delete_guard.rs"]
mod delete_guard;

use delete_guard::check;
use serde_json::Value;
use std::path::{Path, PathBuf};

// REQ-5, REQ-6: the deny offer names its entries, the official docs and the Codex note;
// the apply writes only on a pick, merges without duplicating, keeps sibling lists and
// leaves the file byte-identical when there is nothing new.
mod deny_offer {
    // The library build provides `update_hooks`; the include above is the pre-wiring
    // classification module only (lib.rs does not declare it yet).
}

#[test]
fn deny_offer_previews_entries_docs_and_codex_note() {
    let entries = moai_cli::update_hooks::deny_entries();
    assert!(entries.iter().any(|e| e.contains("rm -rf /")));
    assert!(entries.iter().any(|e| e.starts_with("Bash(sudo")));
    assert!(entries.iter().any(|e| e.contains("$HOME")));
    let preview = moai_cli::update_hooks::deny_preview();
    assert!(preview.contains("permissions.deny"));
    assert!(preview.contains("code.claude.com/docs/en/settings"));
    assert!(preview.contains("hook"));
    assert!(preview.to_ascii_lowercase().contains("codex"));
    assert!(preview.contains("learn.chatgpt.com/docs/agent-configuration/rules"));
}

#[test]
fn deny_apply_writes_only_on_a_pick_and_merges_idempotently() {
    let dir = tempfile::TempDir::new().unwrap();
    let settings = dir.path().join("settings.json");
    std::fs::write(
        &settings,
        r#"{"permissions": {"deny": ["Bash(existing)"], "allow": []}}"#,
    )
    .unwrap();
    let before = std::fs::read(&settings).unwrap();

    // No pick: byte-identical.
    let out = moai_cli::update_hooks::deny_apply(&settings, false).unwrap();
    assert_eq!(out["state"], "not-picked");
    assert_eq!(std::fs::read(&settings).unwrap(), before);

    // With the pick: entries merge, the existing entry and the allow list stay.
    let out = moai_cli::update_hooks::deny_apply(&settings, true).unwrap();
    assert_eq!(out["state"], "applied");
    let v: Value = serde_json::from_slice(&std::fs::read(&settings).unwrap()).unwrap();
    let deny = v["permissions"]["deny"].as_array().unwrap();
    assert!(deny.iter().any(|d| d == "Bash(existing)"));
    assert!(deny.iter().any(|d| d == "Bash(rm -rf /)"));
    assert_eq!(
        v["permissions"]["allow"].as_array().unwrap().len(),
        0,
        "the sibling allow list stays as it was"
    );

    // Applying again adds nothing new.
    let out = moai_cli::update_hooks::deny_apply(&settings, true).unwrap();
    assert_eq!(out["written"], 0);

    // A settings file with no permissions block grows one.
    let fresh = dir.path().join("fresh.json");
    std::fs::write(&fresh, "{}").unwrap();
    moai_cli::update_hooks::deny_apply(&fresh, true).unwrap();
    let v: Value = serde_json::from_slice(&std::fs::read(&fresh).unwrap()).unwrap();
    assert_eq!(
        v["permissions"]["deny"].as_array().unwrap().len(),
        moai_cli::update_hooks::deny_entries().len(),
    );
}

fn refused(name: &str, args: &[&str]) -> String {
    let args: Vec<String> = args.iter().map(|s| s.to_string()).collect();
    match check(name, &args, None) {
        Some(reason) => reason,
        None => panic!("expected a refusal for {name} {args:?}, got none"),
    }
}

fn allowed(name: &str, args: &[&str]) {
    let args: Vec<String> = args.iter().map(|s| s.to_string()).collect();
    if let Some(reason) = check(name, &args, None) {
        panic!("expected {name} {args:?} to be allowed, refused with: {reason}");
    }
}

fn root_of() -> PathBuf {
    Path::new("/Users/goos/MoAI/proj").to_path_buf()
}

fn refused_at_root(name: &str, args: &[&str]) -> String {
    let args: Vec<String> = args.iter().map(|s| s.to_string()).collect();
    match check(name, &args, Some(&root_of())) {
        Some(reason) => reason,
        None => panic!("expected a refusal (at project root) for {name} {args:?}"),
    }
}

fn allowed_at_root(name: &str, args: &[&str]) {
    let args: Vec<String> = args.iter().map(|s| s.to_string()).collect();
    if let Some(reason) = check(name, &args, Some(&root_of())) {
        panic!("expected {name} {args:?} allowed at project root, refused: {reason}");
    }
}

// REQ-1: every recursive spelling of rm refuses the literal catastrophic targets.
#[test]
fn recursive_rm_spellings_refuse_the_root() {
    for argv in [
        vec!["-rf", "/"],
        vec!["-fr", "/"],
        vec!["-r", "/"],
        vec!["-R", "/"],
        vec!["-rRf", "/"],
        vec!["--recursive", "/"],
        vec!["-rf", "/*"],
        vec!["-rf", "///"],
        vec!["-rf", "/."],
        vec!["-rf", "/./"],
        vec!["-rf", "--", "/"],
        vec!["-rf", "~"],
        vec!["-rf", "~/"],
        vec!["-rf", "$HOME"],
        vec!["-rf", "${HOME}"],
        vec!["-rf", "$HOME/*"],
        vec!["-rf", "${HOME}/*"],
        vec!["-rf", "*"],
        vec!["-rf", "."],
        vec!["-rf", ".."],
        vec!["-rf", "/usr"],
        vec!["-rf", ".git"],
        vec!["-rf", "repo/.git/"],
    ] {
        let reason = refused("rm", &argv);
        assert!(
            reason.starts_with("MoAI delete-guard:"),
            "reason line must be the guard prefix, got: {reason}"
        );
    }
}

// REQ-2: a target built from a possibly-empty variable is refused; guarded forms are not.
#[test]
fn empty_variable_paths_are_refused_and_guarded_forms_allowed() {
    for target in ["$M", "${M}", "$M/", "$M/*", "${M}/*", "$S/$1"] {
        let reason = refused("rm", &["-rf", target]);
        assert!(
            reason.contains("variable"),
            "empty-variable reason must name the variable, got: {reason}"
        );
    }
    for target in ["${M:?}/x", "${M:-default}/x"] {
        allowed("rm", &["-rf", target]);
    }
}

// REQ-2 boundary: a `$` that is not a leading variable segment is data.
#[test]
fn mid_word_dollar_is_not_a_variable_path() {
    allowed("rm", &["-rf", "price$2"]);
    allowed("rm", &["-rf", "weird$name"]);
}

// REQ-3: project root, its parent, and their slash forms are refused; children are not.
#[test]
fn project_root_and_parent_refused_children_allowed() {
    refused_at_root("rm", &["-rf", "/Users/goos/MoAI/proj"]);
    refused_at_root("rm", &["-rf", "/Users/goos/MoAI/proj/"]);
    refused_at_root("rm", &["-rf", "/Users/goos/MoAI/proj/."]);
    refused_at_root("rm", &["-rf", "//Users//goos//MoAI//proj"]);
    refused_at_root("rm", &["-rf", "/Users/goos/MoAI"]);
    let parent_reason = refused_at_root("rm", &["-rf", "/Users/goos/MoAI"]);
    assert!(
        parent_reason.contains("parent"),
        "parent refusal must say parent, got: {parent_reason}"
    );
    allowed_at_root("rm", &["-rf", "build"]);
    allowed_at_root("rm", &["-rf", "/Users/goos/MoAI/proj/build"]);
    allowed_at_root("rm", &["-rf", "target/debug"]);
}

// REQ-3 boundary: with no root passed, root-based checks are skipped but the rest run.
#[test]
fn no_root_still_refuses_the_literals() {
    refused("rm", &["-rf", "/"]);
    allowed("rm", &["-rf", "/Users/goos/MoAI/proj"]);
}

// REQ-4: find -delete, find -exec rm, xargs rm and sudo wrapping follow the same targets.
#[test]
fn other_recursive_shapes_follow_the_same_targets() {
    refused("find", &["/", "-delete"]);
    refused("find", &["${M}", "-name", "*.o", "-delete"]);
    refused("find", &["/", "-exec", "rm", "-rf", "{}", "+"]);
    allowed("find", &["build", "-name", "*.o", "-delete"]);
    refused("xargs", &["rm", "-rf", "/"]);
    allowed("xargs", &["rm", "-rf", "build"]);
    refused("sudo", &["rm", "-rf", "/"]);
    refused("sudo", &["find", "/", "-delete"]);
}

// REQ-7: safe deletes keep working.
#[test]
fn safe_deletes_stay_allowed() {
    allowed("rm", &["-rf", "build"]);
    allowed("rm", &["-rf", "./build"]);
    allowed("rm", &["-rf", "node_modules"]);
    allowed("rm", &["-rf", "/tmp/x"]);
    allowed("rm", &["-f", ".gitignore"]);
    allowed("rm", &["-rf", "src/.github"]);
    allowed("rm", &["--", "-weird-name"]);
    allowed("echo", &["rm", "-rf", "/"]);
    allowed("grep", &["-n", "rm -rf /", "notes.md"]);
}

// The reason message names what was hit, so a user can see why.
#[test]
fn reasons_name_the_hit() {
    assert!(refused("rm", &["-rf", "/"]).contains("root"));
    assert!(refused("rm", &["-rf", "~"]).contains("home"));
    assert!(refused("rm", &["-rf", "$M"]).contains("variable"));
    assert!(refused_at_root("rm", &["-rf", "/Users/goos/MoAI/proj"]).contains("project root"));
}
