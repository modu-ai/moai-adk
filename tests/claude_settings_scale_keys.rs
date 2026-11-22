//! The settings writer's checks for the large-codebase keys (SPEC-20261122-project-scale-settings,
//! REQ-9, REQ-10, AC-3): `claudeMdExcludes`, `enabledPlugins`, `worktree.sparsePaths`,
//! `worktree.symlinkDirectories` and `permissions.additionalDirectories`. Every value is checked
//! before any write; the plan functions write nothing. Temporary folders only, no HOME.

use moai_cli::claude::{plan_settings, settings_plan_in, user_settings_desired_in, SettingsScope};
use serde_json::{json, Map, Value};
use std::fs;
use tempfile::TempDir;

fn obj(v: Value) -> Map<String, Value> {
    v.as_object().unwrap().clone()
}

fn plan(desired: Value) -> Result<Map<String, Value>, String> {
    plan_settings(&Map::new(), &Map::new(), &obj(desired)).map(|(planned, conflicts)| {
        assert!(conflicts.is_empty());
        planned.unwrap()
    })
}

#[test]
fn the_new_keys_are_checked_before_any_write() {
    // accepted values: what the large-codebase plan writes.
    for desired in [
        json!({"claudeMdExcludes": ["**/packages/legacy/CLAUDE.md", "**/vendor/**/CLAUDE.md"]}),
        json!({"enabledPlugins": {"typescript-lsp@claude-plugins-official": true}}),
        json!({"enabledPlugins": {"rust-analyzer-lsp@claude-plugins-official": true,
                                  "pyright-lsp@claude-plugins-official": true}}),
        json!({"worktree": {"sparsePaths": [".claude", "packages/api", "packages/shared"]}}),
        json!({"worktree": {"baseRef": "head", "sparsePaths": [".claude", "apps/web"],
                            "symlinkDirectories": ["node_modules"]}}),
        json!({"worktree": {"symlinkDirectories": ["node_modules", ".venv"]}}),
        json!({"permissions": {"additionalDirectories": ["../shared", "../../libs/core"],
                               "deny": ["Read(./**/dist/**/*)"]}}),
    ] {
        let planned = plan(desired.clone()).unwrap_or_else(|e| panic!("{desired}: {e}"));
        let key = desired.as_object().unwrap().keys().next().unwrap();
        assert_eq!(planned[key], desired[key]);
    }
    // the m-flow value of the earlier card is still accepted unchanged.
    for base in ["head", "fresh"] {
        assert!(plan(json!({"worktree": {"baseRef": base}})).is_ok());
    }

    // refused values: each names its key and plans nothing.
    for (desired, key) in [
        // `..` that leaves the repository, roots and drives
        (
            json!({"claudeMdExcludes": ["../other/CLAUDE.md"]}),
            "claudeMdExcludes",
        ),
        (
            json!({"claudeMdExcludes": ["**/a/../../b/CLAUDE.md"]}),
            "claudeMdExcludes",
        ),
        (
            json!({"worktree": {"sparsePaths": ["../outside"]}}),
            "worktree",
        ),
        (
            json!({"worktree": {"sparsePaths": ["packages/../.."]}}),
            "worktree",
        ),
        (json!({"worktree": {"sparsePaths": ["/abs"]}}), "worktree"),
        (
            json!({"worktree": {"symlinkDirectories": ["C:\\deps"]}}),
            "worktree",
        ),
        (
            json!({"worktree": {"symlinkDirectories": ["\\\\server\\share"]}}),
            "worktree",
        ),
        (
            json!({"permissions": {"additionalDirectories": ["/"]}}),
            "permissions",
        ),
        (
            json!({"permissions": {"additionalDirectories": ["/Users/x/repo"]}}),
            "permissions",
        ),
        (
            json!({"permissions": {"additionalDirectories": ["~/code"]}}),
            "permissions",
        ),
        (
            json!({"permissions": {"additionalDirectories": ["D:/code"]}}),
            "permissions",
        ),
        (
            json!({"permissions": {"additionalDirectories": ["a/../../b"]}}),
            "permissions",
        ),
        // control characters and empty or long values
        (json!({"claudeMdExcludes": ["a\nb"]}), "claudeMdExcludes"),
        (json!({"claudeMdExcludes": [""]}), "claudeMdExcludes"),
        (
            json!({"claudeMdExcludes": ["x".repeat(257)]}),
            "claudeMdExcludes",
        ),
        (json!({"worktree": {"sparsePaths": ["a\tb"]}}), "worktree"),
        (
            json!({"permissions": {"additionalDirectories": ["a\u{7}"]}}),
            "permissions",
        ),
        // wrong shapes
        (
            json!({"claudeMdExcludes": "**/CLAUDE.md"}),
            "claudeMdExcludes",
        ),
        (json!({"claudeMdExcludes": [1]}), "claudeMdExcludes"),
        (
            json!({"enabledPlugins": ["typescript-lsp"]}),
            "enabledPlugins",
        ),
        (
            json!({"enabledPlugins": {"typescript-lsp": true}}),
            "enabledPlugins",
        ),
        (
            json!({"enabledPlugins": {"@market": true}}),
            "enabledPlugins",
        ),
        (json!({"enabledPlugins": {"name@": true}}), "enabledPlugins"),
        (
            json!({"enabledPlugins": {"a b@market": true}}),
            "enabledPlugins",
        ),
        (
            json!({"enabledPlugins": {"x@market": false}}),
            "enabledPlugins",
        ),
        (
            json!({"enabledPlugins": {"x@market": "true"}}),
            "enabledPlugins",
        ),
        (json!({"worktree": {}}), "worktree"),
        (json!({"worktree": {"sparsePaths": ".claude"}}), "worktree"),
        (json!({"worktree": {"baseRef": "main"}}), "worktree"),
        // any other sub-key than the documented ones
        (json!({"worktree": {"bgIsolation": "none"}}), "worktree"),
        (json!({"worktree": {"location": "/tmp"}}), "worktree"),
        (
            json!({"permissions": {"disableBypassPermissionsMode": "disable"}}),
            "permissions",
        ),
        // a value shaped like a credential
        (
            json!({"claudeMdExcludes": ["https://user:pw@example.com/CLAUDE.md"]}),
            "claudeMdExcludes",
        ),
        (
            json!({"enabledPlugins": {"x@https://u:p@host": true}}),
            "enabledPlugins",
        ),
    ] {
        let err = plan(desired.clone()).expect_err(&desired.to_string());
        assert!(err.contains(key), "{desired}: {err}");
    }
}

#[test]
fn the_lists_are_bounded_at_256() {
    let paths = |n: usize| (0..n).map(|i| format!("p{i}")).collect::<Vec<_>>();
    assert!(plan(json!({"worktree": {"sparsePaths": paths(256)}})).is_ok());
    assert!(plan(json!({"worktree": {"sparsePaths": paths(257)}})).is_err());
    assert!(plan(json!({"claudeMdExcludes": paths(256)})).is_ok());
    assert!(plan(json!({"claudeMdExcludes": paths(257)})).is_err());
    assert!(plan(json!({"permissions": {"additionalDirectories": paths(256)}})).is_ok());
    assert!(plan(json!({"permissions": {"additionalDirectories": paths(257)}})).is_err());
    let plugins: Map<String, Value> = (0..65).map(|i| (format!("p{i}@m"), json!(true))).collect();
    let mut sixty_four = plugins.clone();
    sixty_four.remove("p0@m");
    assert!(plan(json!({"enabledPlugins": sixty_four})).is_ok());
    assert!(plan(json!({"enabledPlugins": plugins})).is_err());
    // one value at the character limit is accepted.
    assert!(plan(json!({"claudeMdExcludes": ["x".repeat(256)]})).is_ok());
}

#[test]
fn a_caller_may_set_the_new_keys_in_either_scope() {
    let dir = TempDir::new().unwrap();
    let desired = obj(json!({
        "claudeMdExcludes": ["**/packages/legacy/CLAUDE.md"],
        "enabledPlugins": {"typescript-lsp@claude-plugins-official": true}
    }));
    for scope in [SettingsScope::Project, SettingsScope::Local] {
        let out = user_settings_desired_in(dir.path(), &desired, scope).unwrap();
        assert_eq!(out["claudeMdExcludes"], desired["claudeMdExcludes"]);
        assert_eq!(out["enabledPlugins"], desired["enabledPlugins"]);
    }
    // nothing was written by asking.
    assert_eq!(fs::read_dir(dir.path()).unwrap().count(), 0);
}

#[test]
fn an_additional_directory_stays_inside_the_repository() {
    let repo = TempDir::new().unwrap();
    fs::create_dir_all(repo.path().join(".git")).unwrap();
    let api = repo.path().join("packages/api");
    fs::create_dir_all(&api).unwrap();
    let ask = |dirs: Value| {
        user_settings_desired_in(
            &api,
            &obj(json!({"permissions": {"additionalDirectories": dirs}})),
            SettingsScope::Project,
        )
    };
    // a sibling package and the repository root are inside.
    assert!(ask(json!(["../web", "../shared"])).is_ok());
    assert!(ask(json!(["../.."])).is_ok());
    // one level more leaves the repository.
    let err = ask(json!(["../../../elsewhere"])).unwrap_err();
    assert!(err.contains("additionalDirectories"), "{err}");
    // a project with no repository above it accepts no `..` at all.
    let lone = TempDir::new().unwrap();
    let err = user_settings_desired_in(
        lone.path(),
        &obj(json!({"permissions": {"additionalDirectories": ["../x"]}})),
        SettingsScope::Project,
    )
    .unwrap_err();
    assert!(err.contains("additionalDirectories"), "{err}");
    assert!(user_settings_desired_in(
        lone.path(),
        &obj(json!({"permissions": {"additionalDirectories": ["sub"]}})),
        SettingsScope::Project,
    )
    .is_ok());
}

#[test]
fn a_token_shaped_plugin_key_is_refused_and_hidden() {
    // an obviously fake token, shaped like one.
    let token = format!("ghp_{}", "x".repeat(36));
    let mut plugins = Map::new();
    plugins.insert(format!("{token}@claude-plugins-official"), json!(true));
    let err = plan(json!({"enabledPlugins": plugins})).unwrap_err();
    assert!(err.contains("enabledPlugins"), "{err}");
    assert!(err.contains("looks like a credential"), "{err}");
    assert!(!err.contains(&token) && !err.contains("ghp_"), "{err}");
    // a plugin name that only starts like one is still a plugin name.
    assert!(plan(json!({"enabledPlugins": {"ghp_lsp@claude-plugins-official": true}})).is_ok());
    // an unmanaged key shaped like a token is refused without being echoed.
    let mut desired = Map::new();
    desired.insert(token.clone(), json!(true));
    let err = plan_settings(&Map::new(), &Map::new(), &desired).unwrap_err();
    assert!(
        err.contains("unmanaged claude settings key refused"),
        "{err}"
    );
    assert!(err.contains("looks like a credential"), "{err}");
    assert!(!err.contains(&token) && !err.contains("ghp_"), "{err}");
    // a plain unmanaged key is still named.
    let err = plan_settings(&Map::new(), &Map::new(), &obj(json!({"theme": "dark"}))).unwrap_err();
    assert!(err.contains("theme"), "{err}");
}

#[test]
fn a_token_split_by_a_control_character_or_named_as_an_env_entry_is_never_echoed() {
    // an obviously fake token, shaped like one.
    let token = format!("ghp_{}", "x".repeat(36));
    // a control character inside the token hides its shape until it is stripped; the stripped
    // text is what a message would print, so it is screened too.
    let split = format!("ghp_{}\u{7}{}", "x".repeat(18), "x".repeat(18));
    let mut desired = Map::new();
    desired.insert(split.clone(), json!(true));
    let err = plan_settings(&Map::new(), &Map::new(), &desired).unwrap_err();
    assert!(
        err.contains("unmanaged claude settings key refused"),
        "{err}"
    );
    assert!(err.contains("looks like a credential"), "{err}");
    assert!(!err.contains(&token) && !err.contains("ghp_"), "{err}");
    // an env entry named like a token is refused, and its name is not printed back, whether the
    // refusal is about the name or about the value.
    for value in [json!("1"), json!("two\nlines")] {
        let mut env = Map::new();
        env.insert(token.clone(), value);
        let err = plan(json!({ "env": env })).unwrap_err();
        assert!(err.contains("env entry"), "{err}");
        assert!(err.contains("looks like a credential"), "{err}");
        assert!(!err.contains(&token) && !err.contains("ghp_"), "{err}");
    }
    // a plain refused env name is still named.
    let err = plan(json!({"env": {"PLAIN_NAME": "two\nlines"}})).unwrap_err();
    assert!(err.contains("PLAIN_NAME"), "{err}");
}

#[test]
fn a_token_split_by_an_invisible_format_or_separator_character_is_never_echoed() {
    // an obviously fake token, shaped like one.
    let token = format!("ghp_{}", "x".repeat(36));
    for mark in [
        '\u{200B}',
        '\u{00AD}',
        '\u{202E}',
        '\u{2028}',
        '\u{2029}',
        '\u{FEFF}',
        '\u{2060}',
        '\u{E0041}',
    ] {
        let split = format!("ghp_{}{mark}{}", "x".repeat(18), "x".repeat(18));
        let mut desired = Map::new();
        desired.insert(split, json!(true));
        let err = plan_settings(&Map::new(), &Map::new(), &desired).unwrap_err();
        assert!(
            err.contains("unmanaged claude settings key refused"),
            "{mark:?}: {err}"
        );
        assert!(err.contains("looks like a credential"), "{mark:?}: {err}");
        assert!(
            !err.contains(&token) && !err.contains("ghp_"),
            "{mark:?}: {err}"
        );
    }
    // a plain name is still named, without the invisible character that could hide or reorder
    // what the reader sees.
    let mut desired = Map::new();
    desired.insert("the\u{202E}me".into(), json!("dark"));
    let err = plan_settings(&Map::new(), &Map::new(), &desired).unwrap_err();
    assert!(err.contains("theme"), "{err}");
    assert!(!err.contains('\u{202E}'), "{err:?}");
}

#[test]
fn the_writer_refuses_an_additional_directory_outside_the_repository() {
    let repo = TempDir::new().unwrap();
    fs::create_dir_all(repo.path().join(".git")).unwrap();
    let pkg = repo.path().join("packages");
    fs::create_dir_all(&pkg).unwrap();
    let dirs = |d: Value| obj(json!({"permissions": {"additionalDirectories": d}}));
    // the settings plan itself refuses, not only the caller's desired map: one `..` from the
    // repository root leaves it.
    let err = settings_plan_in(
        repo.path(),
        &dirs(json!(["../outside"])),
        SettingsScope::Project,
    )
    .unwrap_err();
    assert!(err.contains("additionalDirectories"), "{err}");
    // from a folder one level down, one `..` stays inside and two leave.
    assert!(settings_plan_in(&pkg, &dirs(json!(["../web"])), SettingsScope::Project).is_ok());
    let err =
        settings_plan_in(&pkg, &dirs(json!(["../../x"])), SettingsScope::Project).unwrap_err();
    assert!(err.contains("additionalDirectories"), "{err}");
    // the search for `.git` stops at the repository root: an outer repository above it does
    // not widen what may be named.
    let outer = TempDir::new().unwrap();
    fs::create_dir_all(outer.path().join(".git")).unwrap();
    let inner = outer.path().join("inner");
    fs::create_dir_all(inner.join(".git")).unwrap();
    fs::create_dir_all(inner.join("pkg")).unwrap();
    assert!(settings_plan_in(
        &inner.join("pkg"),
        &dirs(json!([".."])),
        SettingsScope::Project
    )
    .is_ok());
    let err = settings_plan_in(
        &inner.join("pkg"),
        &dirs(json!(["../../x"])),
        SettingsScope::Project,
    )
    .unwrap_err();
    assert!(err.contains("additionalDirectories"), "{err}");
    // planning wrote nothing.
    assert!(!repo.path().join(".claude").exists());
    assert!(!pkg.join(".claude").exists());
    assert!(!inner.join("pkg/.claude").exists());
}
