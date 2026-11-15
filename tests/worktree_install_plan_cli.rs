//! #3010 P6a: `worktree-install-plan` reads the `scopes` key of the project profile and shows
//! the install command for one scope. It runs nothing and writes nothing. Real binary,
//! temporary projects only.

mod common;

use common::cli::{cli, create, ok_result, profile, tree_state, write_file};
use serde_json::{json, Value};
use tempfile::TempDir;

/// A project whose profile declares the scope `web` (two folders) and that holds a pnpm lockfile.
fn project() -> TempDir {
    let root = TempDir::new().unwrap();
    let mut v = profile(None);
    v["scopes"] = json!([{"name": "web", "paths": ["apps/web", "packages/ui"]}]);
    create(&root, &v);
    write_file(root.path(), "pnpm-lock.yaml", b"");
    write_file(root.path(), "apps/web/package.json", b"{}");
    write_file(root.path(), "packages/ui/package.json", b"{}");
    root
}

fn preview(root: &TempDir, scope: &str) -> std::process::Output {
    cli(
        &[
            "worktree-install-plan",
            root.path().to_str().unwrap(),
            "--scope",
            scope,
        ],
        &Value::Null,
    )
}

#[test]
fn a_declared_scope_gets_a_preview_and_nothing_is_written() {
    let root = project();
    let before = tree_state(root.path());
    let out = preview(&root, "web");
    assert!(
        out.status.success(),
        "{}",
        String::from_utf8_lossy(&out.stdout)
    );
    let result = ok_result(&out);
    assert_eq!(result["mode"], "scoped");
    assert_eq!(result["manager"], "pnpm");
    assert_eq!(result["ran"], false);
    assert_eq!(
        result["commands"],
        json!(["pnpm install --frozen-lockfile --filter 'apps/web' --filter 'packages/ui'"])
    );
    assert_eq!(tree_state(root.path()), before);
}

#[test]
fn an_undeclared_scope_or_a_profile_without_scopes_is_refused() {
    let root = project();
    let out = preview(&root, "api");
    assert!(!out.status.success());
    assert!(String::from_utf8_lossy(&out.stdout).contains("no scope with that name"));

    let bare = TempDir::new().unwrap();
    create(&bare, &profile(None));
    let out = preview(&bare, "web");
    assert!(!out.status.success());
    assert!(String::from_utf8_lossy(&out.stdout).contains("declares no scopes"));
}

#[test]
fn a_project_without_a_profile_is_refused() {
    let root = TempDir::new().unwrap();
    let out = preview(&root, "web");
    assert!(!out.status.success());
}
