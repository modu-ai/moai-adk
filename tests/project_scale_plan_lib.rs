//! The large-codebase planner (SPEC-20261122-project-scale REQ-6, REQ-7, REQ-10, REQ-11 with
//! AC-1 and AC-4; SPEC-20261122-project-scale-settings REQ-1 to REQ-7 with AC-1 and AC-2). The
//! plan is read-only: every fixture is compared byte for byte before and after. Temporary
//! folders only, no HOME, no program started.

use moai_cli::project_scale::{scale_plan, ScalePlan, ScaleRequest};
use serde_json::{json, Value};
use std::collections::BTreeMap;
use std::fs;
use std::path::Path;
use tempfile::TempDir;

const OFFICIAL_KEYS: [&str; 6] = [
    "claudeMdExcludes",
    "permissions.deny",
    "enabledPlugins",
    "worktree.sparsePaths",
    "worktree.symlinkDirectories",
    "permissions.additionalDirectories",
];

fn write(root: &Path, rel: &str, text: &str) {
    let path = root.join(rel);
    fs::create_dir_all(path.parent().unwrap()).unwrap();
    fs::write(path, text).unwrap();
}

/// A git index header that says the index holds `entries` files; the detector reads only it.
fn index_with(root: &Path, entries: u32) {
    let mut header = b"DIRC".to_vec();
    header.extend_from_slice(&2u32.to_be_bytes());
    header.extend_from_slice(&entries.to_be_bytes());
    write(root, ".git/HEAD", "ref: refs/heads/main\n");
    fs::write(root.join(".git/index"), header).unwrap();
}

fn small() -> TempDir {
    let dir = TempDir::new().unwrap();
    write(dir.path(), "Cargo.toml", "[package]\nname = \"one\"\n");
    write(dir.path(), "src/main.rs", "fn main() {}\n");
    index_with(dir.path(), 2);
    dir
}

fn large_tree() -> TempDir {
    let dir = TempDir::new().unwrap();
    write(dir.path(), "Cargo.toml", "[package]\nname = \"big\"\n");
    write(dir.path(), "src/main.rs", "fn main() {}\n");
    write(dir.path(), "generated/schema.rs", "// generated\n");
    index_with(dir.path(), 5_000);
    dir
}

fn monorepo() -> TempDir {
    let dir = TempDir::new().unwrap();
    let root = dir.path();
    write(
        root,
        "package.json",
        "{\"name\": \"mono\", \"private\": true}\n",
    );
    write(root, "pnpm-workspace.yaml", "packages:\n  - packages/*\n");
    for pkg in ["api", "web", "shared", "legacy"] {
        write(
            root,
            &format!("packages/{pkg}/package.json"),
            &format!("{{\"name\": \"{pkg}\"}}\n"),
        );
        write(root, &format!("packages/{pkg}/index.js"), "export {}\n");
    }
    write(root, "generated/api.js", "// generated\n");
    write(
        root,
        "node_modules/left-pad/index.js",
        "module.exports = 1\n",
    );
    index_with(root, 40);
    dir
}

/// Every file under `root` with its bytes, folders included as empty entries.
fn snapshot(root: &Path) -> BTreeMap<String, Option<Vec<u8>>> {
    let mut out = BTreeMap::new();
    let mut stack = vec![root.to_path_buf()];
    while let Some(dir) = stack.pop() {
        for entry in fs::read_dir(&dir).unwrap() {
            let path = entry.unwrap().path();
            let rel = path
                .strip_prefix(root)
                .unwrap()
                .to_string_lossy()
                .to_string();
            if path.is_dir() {
                out.insert(rel, None);
                stack.push(path);
            } else {
                out.insert(rel, Some(fs::read(&path).unwrap()));
            }
        }
    }
    out
}

fn request(v: Value) -> ScaleRequest {
    serde_json::from_value(v).unwrap()
}

fn plan(root: &Path, v: Value) -> Result<ScalePlan, String> {
    scale_plan(root, request(v))
}

fn row<'a>(
    p: &'a ScalePlan,
    file: &str,
    key: Option<&str>,
) -> &'a moai_cli::project_scale::ScaleRow {
    p.rows
        .iter()
        .find(|r| r.file == file && r.key.as_deref() == key)
        .unwrap_or_else(|| panic!("no row {file} {key:?} in {:#?}", p.rows))
}

#[test]
fn three_scales_give_three_previews_and_write_nothing() {
    let (s, l, m) = (small(), large_tree(), monorepo());
    let before = [snapshot(s.path()), snapshot(l.path()), snapshot(m.path())];

    let ps = plan(
        s.path(),
        json!({"protocol_version": 1, "kind": "small",
        "deny_read": ["src"]}),
    )
    .unwrap();
    let pl = plan(
        l.path(),
        json!({"protocol_version": 1, "kind": "large_tree",
        "deny_read": ["generated"], "sparse_paths": ["src"]}),
    )
    .unwrap();
    let pm = plan(
        m.path(),
        json!({"protocol_version": 1, "kind": "monorepo",
        "package_files": ["packages/api"], "exclude_packages": ["packages/legacy"]}),
    )
    .unwrap();

    assert_eq!(
        (
            ps.detected_kind.as_str(),
            pl.detected_kind.as_str(),
            pm.detected_kind.as_str()
        ),
        ("small", "large_tree", "monorepo")
    );
    // small plans nothing and says so, whatever else was picked.
    assert!(ps.rows.is_empty(), "{:#?}", ps.rows);
    assert!(
        ps.notes.iter().any(|n| n.contains("nothing is planned")),
        "{:?}",
        ps.notes
    );
    assert!(!pl.rows.is_empty() && !pm.rows.is_empty());
    let shape = |p: &ScalePlan| -> Vec<(String, Option<String>)> {
        p.rows
            .iter()
            .map(|r| (r.file.clone(), r.key.clone()))
            .collect()
    };
    assert_ne!(shape(&pl), shape(&pm));
    // every settings key the plan names is an official Claude Code key.
    for r in pl.rows.iter().chain(&pm.rows) {
        if let Some(k) = &r.key {
            assert!(OFFICIAL_KEYS.contains(&k.as_str()), "{k}");
        }
    }
    // the three digests differ and the plan wrote nothing anywhere.
    assert_ne!(ps.plan_digest, pl.plan_digest);
    assert_ne!(pl.plan_digest, pm.plan_digest);
    assert_eq!(
        [snapshot(s.path()), snapshot(l.path()), snapshot(m.path())],
        before
    );
    // the same request plans the same digest again.
    let again = plan(
        m.path(),
        json!({"protocol_version": 1, "kind": "monorepo",
        "package_files": ["packages/api"], "exclude_packages": ["packages/legacy"]}),
    )
    .unwrap();
    assert_eq!(again.plan_digest, pm.plan_digest);
}

#[test]
fn a_monorepo_preview_lists_every_native_setting_with_its_doc_line() {
    let m = monorepo();
    let root = m.path();
    write(root, "packages/api/dist/index.js", "// built\n");
    let before = snapshot(root);
    let p = plan(
        root,
        json!({
            "protocol_version": 1, "kind": "monorepo",
            "package_files": ["packages/api", "packages/web"],
            "exclude_packages": ["packages/legacy"],
            "deny_read": ["generated", "packages/api/dist"],
            "plugins": ["javascript"],
            "sparse_paths": ["packages/api", "packages/shared"],
            "ignore_worktrees": true,
            "start_packages": ["packages/api"]
        }),
    )
    .unwrap();

    // REQ-1: one CLAUDE.md per picked package, header and detected facts only.
    let api = row(&p, "packages/api/CLAUDE.md", None);
    assert_eq!(api.action, "create");
    let text = api.value.as_str().unwrap();
    assert!(text.starts_with("# packages/api\n"), "{text}");
    assert!(text.contains("`package.json`"), "{text}");
    assert_eq!(row(&p, "packages/web/CLAUDE.md", None).action, "create");
    assert!(!p.rows.iter().any(|r| r.file == "packages/shared/CLAUDE.md"));
    // REQ-3: the excluded package in the local settings file.
    let ex = row(&p, ".claude/settings.local.json", Some("claudeMdExcludes"));
    assert_eq!(ex.value, json!(["**/packages/legacy/**/CLAUDE.md"]));
    assert_eq!(ex.action, "create");
    // REQ-4: a Read deny rule in the root settings.
    let deny = row(&p, ".claude/settings.json", Some("permissions.deny"));
    assert_eq!(
        deny.value,
        json!(["Read(./generated/**/*)", "Read(./packages/api/dist/**/*)"])
    );
    // REQ-5: the plugin key, and the install command and binary as text only.
    let plug = row(&p, ".claude/settings.json", Some("enabledPlugins"));
    assert_eq!(
        plug.value,
        json!({"typescript-lsp@claude-plugins-official": true})
    );
    assert!(p.install.iter().any(|l| l
        .contains("/plugin install typescript-lsp@claude-plugins-official")
        && l.contains("typescript-language-server")));
    // REQ-6: sparse paths start with .claude; the root dependency folder is linked.
    assert_eq!(
        row(&p, ".claude/settings.json", Some("worktree.sparsePaths")).value,
        json!([".claude", "packages/api", "packages/shared"])
    );
    assert_eq!(
        row(
            &p,
            ".claude/settings.json",
            Some("worktree.symlinkDirectories")
        )
        .value,
        json!(["node_modules"])
    );
    let ignore = row(&p, ".git/info/exclude", None);
    assert_eq!(ignore.value, json!(".claude/worktrees/"));
    // REQ-7: a self-contained package settings file naming only siblings that exist. It carries
    // the Read deny rules of its own folders again, relative to the package, because a session
    // started there does not inherit the root file; the root file keeps them too, for a
    // worktree session.
    let own = row(&p, "packages/api/.claude/settings.json", None);
    assert_eq!(own.action, "create");
    assert_eq!(
        own.value,
        json!({"permissions": {
            "additionalDirectories": ["../legacy", "../shared", "../web"],
            "deny": ["Read(./dist/**/*)"]
        }})
    );
    // every row carries its documentation line.
    for r in &p.rows {
        assert!(
            r.doc_url.starts_with("https://code.claude.com/docs/en/"),
            "{} {:?}: {}",
            r.file,
            r.key,
            r.doc_url
        );
        assert!(r.doc_quote.len() > 20, "{} {:?}", r.file, r.key);
    }
    // REQ-10: one paragraph on where to start Claude.
    let notes = p.notes.join("\n");
    for words in ["aren't inherited", "plan mode", "start Claude"] {
        assert!(notes.contains(words), "{words}: {notes}");
    }
    assert!(p.docs_read_on.starts_with("2026-"), "{}", p.docs_read_on);
    assert_eq!(snapshot(root), before);
}

#[test]
fn the_plan_never_writes_an_unscoped_rule_or_a_long_file() {
    let m = monorepo();
    let root = m.path();
    // a package with many manifests still gets a short file.
    for i in 0..300 {
        write(root, &format!("packages/api/sub{i}/package.json"), "{}\n");
    }
    let p = plan(
        root,
        json!({
            "protocol_version": 1, "kind": "monorepo",
            "package_files": ["packages/api", "packages/web", "packages/shared"],
            "exclude_packages": ["packages/legacy"],
            "start_packages": ["packages/api", "packages/web"]
        }),
    )
    .unwrap();
    assert!(!p.rows.is_empty());
    for r in &p.rows {
        assert!(
            !r.file.contains(".claude/rules/"),
            "unscoped rule file {}",
            r.file
        );
        if let Some(text) = r.value.as_str() {
            assert!(
                text.lines().count() <= 200,
                "{} has {} lines",
                r.file,
                text.lines().count()
            );
        }
        if r.file.ends_with("SKILL.md") {
            let text = r.value.as_str().unwrap();
            let description = text
                .lines()
                .find_map(|l| l.strip_prefix("description: "))
                .unwrap();
            assert!(description.chars().count() <= 250, "{description}");
        }
    }
}

#[test]
fn a_credential_or_an_outside_path_refuses_the_whole_plan() {
    let m = monorepo();
    let root = m.path();
    let before = snapshot(root);
    for (picks, key) in [
        (
            json!({"deny_read": ["https://user:pw@example.com/x"]}),
            "deny_read",
        ),
        (json!({"deny_read": ["../outside"]}), "deny_read"),
        (json!({"sparse_paths": ["packages/../.."]}), "sparse_paths"),
        (json!({"sparse_paths": ["/etc"]}), "sparse_paths"),
        (json!({"package_files": ["C:\\code"]}), "package_files"),
        (json!({"package_files": ["packages/nope"]}), "package_files"),
        (
            json!({"exclude_packages": ["packages/a\nb"]}),
            "exclude_packages",
        ),
        (json!({"start_packages": ["~/code"]}), "start_packages"),
        (json!({"plugins": ["cobol"]}), "plugins"),
        (json!({"plugins": ["rust"]}), "plugins"),
        (
            json!({"deny_read": ["generated"], "sparse_paths": ["packages/api", "../x"]}),
            "sparse_paths",
        ),
    ] {
        let mut v = json!({"protocol_version": 1, "kind": "monorepo"});
        for (k, val) in picks.as_object().unwrap() {
            v[k] = val.clone();
        }
        let err = plan(root, v.clone()).expect_err(&v.to_string());
        assert!(err.contains(key), "{v}: {err}");
    }
    for bad in [
        json!({"protocol_version": 1, "kind": "huge"}),
        json!({"protocol_version": 2, "kind": "monorepo"}),
    ] {
        assert!(scale_plan(root, request(bad.clone())).is_err(), "{bad}");
    }
    // an unknown request field is refused when read.
    assert!(serde_json::from_value::<ScaleRequest>(
        json!({"protocol_version": 1, "kind": "monorepo", "rules": ["x"]})
    )
    .is_err());
    assert_eq!(snapshot(root), before);
}

#[test]
fn a_refusal_never_echoes_a_credential_shaped_pick() {
    let m = monorepo();
    let root = m.path();
    // obviously fake secrets: a token shape and a URL with user information.
    let token = format!("ghp_{}", "x".repeat(36));
    let url = "https://user:pass@example.com/x".to_string();
    for secret in [&token, &url] {
        // a path pick is refused for what it is, by the path check, and names no value.
        for key in [
            "deny_read",
            "sparse_paths",
            "package_files",
            "exclude_packages",
            "start_packages",
        ] {
            let mut v = json!({"protocol_version": 1, "kind": "monorepo"});
            v[key] = json!([secret]);
            let err = plan(root, v).unwrap_err();
            assert_eq!(
                err,
                format!("{key} refused: it looks like a credential; nothing is planned")
            );
        }
        // a plugin pick is not a path: its refusal hides the value.
        let err = plan(
            root,
            json!({"protocol_version": 1, "kind": "monorepo", "plugins": [secret]}),
        )
        .unwrap_err();
        assert!(err.contains("plugins refused"), "{err}");
        assert!(err.contains("looks like a credential"), "{err}");
        for part in [secret.as_str(), "ghp_", "user:pass"] {
            assert!(!err.contains(part), "{err}");
        }
    }
    // a plain unknown plugin is still named, so the user sees which pick was refused.
    let err = plan(
        root,
        json!({"protocol_version": 1, "kind": "monorepo", "plugins": ["cobol"]}),
    )
    .unwrap_err();
    assert!(err.contains("cobol"), "{err}");
}

#[test]
fn the_repository_root_itself_is_refused_as_a_deny_or_sparse_pick() {
    let m = monorepo();
    let root = m.path();
    for key in ["deny_read", "sparse_paths"] {
        for pick in [".", "./", "./.", "././"] {
            let mut v = json!({"protocol_version": 1, "kind": "monorepo"});
            v[key] = json!([pick]);
            let err = plan(root, v).expect_err(pick);
            assert!(err.contains(key), "{pick}: {err}");
            assert!(err.contains("repository root"), "{pick}: {err}");
        }
    }
    // a folder below the root, written with `./` and a trailing `/`, still passes.
    let p = plan(
        root,
        json!({"protocol_version": 1, "kind": "monorepo",
        "deny_read": ["./generated/"], "sparse_paths": ["./packages/api/"]}),
    )
    .unwrap();
    assert_eq!(
        row(&p, ".claude/settings.json", Some("permissions.deny")).value,
        json!(["Read(./generated/**/*)"])
    );
}

#[test]
fn a_merge_row_carries_the_merged_value_and_drops_nothing() {
    let m = monorepo();
    let root = m.path();
    let root_settings = json!({
        "permissions": {"allow": ["Bash(ls)"], "deny": ["Read(./secrets/**/*)"]},
        "enabledPlugins": {"other-lsp@claude-plugins-official": true},
        "worktree": {"baseRef": "head", "sparsePaths": [".claude", "docs"]}
    });
    let local_settings = json!({"claudeMdExcludes": ["**/old/**/CLAUDE.md"]});
    write(
        root,
        ".claude/settings.json",
        &serde_json::to_string_pretty(&root_settings).unwrap(),
    );
    write(
        root,
        ".claude/settings.local.json",
        &serde_json::to_string_pretty(&local_settings).unwrap(),
    );
    let before = snapshot(root);
    let p = plan(
        root,
        json!({
            "protocol_version": 1, "kind": "monorepo",
            "exclude_packages": ["packages/legacy"],
            "deny_read": ["generated"],
            "plugins": ["javascript"],
            "sparse_paths": ["packages/api"]
        }),
    )
    .unwrap();
    assert_eq!(snapshot(root), before);

    // arrays: the current entries first in their order, the new ones appended once.
    let deny = row(&p, ".claude/settings.json", Some("permissions.deny"));
    assert_eq!(deny.action, "merge");
    assert_eq!(
        deny.value,
        json!(["Read(./secrets/**/*)", "Read(./generated/**/*)"])
    );
    let ex = row(&p, ".claude/settings.local.json", Some("claudeMdExcludes"));
    assert_eq!(ex.action, "merge");
    assert_eq!(
        ex.value,
        json!(["**/old/**/CLAUDE.md", "**/packages/legacy/**/CLAUDE.md"])
    );
    let sparse = row(&p, ".claude/settings.json", Some("worktree.sparsePaths"));
    assert_eq!(sparse.action, "merge");
    assert_eq!(sparse.value, json!([".claude", "docs", "packages/api"]));
    // the plugin object: the plugins already there stay.
    let plug = row(&p, ".claude/settings.json", Some("enabledPlugins"));
    assert_eq!(plug.action, "merge");
    assert_eq!(
        plug.value,
        json!({"other-lsp@claude-plugins-official": true,
               "typescript-lsp@claude-plugins-official": true})
    );

    // a writer that sets each row's key to its value drops no entry and keeps the sub-keys it
    // does not name (`worktree.baseRef`, `permissions.allow`).
    let mut written = root_settings.clone();
    for r in p.rows.iter().filter(|r| r.file == ".claude/settings.json") {
        let key = r.key.as_deref().unwrap();
        match key.split_once('.') {
            Some((parent, child)) => written[parent][child] = r.value.clone(),
            None => written[key] = r.value.clone(),
        }
    }
    assert_eq!(written["worktree"]["baseRef"], json!("head"));
    assert_eq!(written["permissions"]["allow"], json!(["Bash(ls)"]));
    assert_eq!(
        written["permissions"]["deny"][0],
        json!("Read(./secrets/**/*)")
    );
    assert_eq!(written["worktree"]["sparsePaths"][1], json!("docs"));
    assert_eq!(
        written["enabledPlugins"]["other-lsp@claude-plugins-official"],
        json!(true)
    );

    // the same picks again on the merged files are unchanged rows holding the current value.
    write(
        root,
        ".claude/settings.json",
        &serde_json::to_string_pretty(&written).unwrap(),
    );
    let again = plan(
        root,
        json!({"protocol_version": 1, "kind": "monorepo", "deny_read": ["generated"]}),
    )
    .unwrap();
    let deny = row(&again, ".claude/settings.json", Some("permissions.deny"));
    assert_eq!(deny.action, "unchanged");
    assert_eq!(deny.value, written["permissions"]["deny"]);
}

#[test]
fn a_linked_worktree_reads_the_exclude_file_of_the_common_git_dir() {
    let main = large_tree();
    let wt = TempDir::new().unwrap();
    write(wt.path(), "src/main.rs", "fn main() {}\n");
    // a linked worktree: `.git` is a file naming its admin folder below the common git dir,
    // which names the worktree back and points at the common dir with `commondir`.
    let admin = main.path().join(".git/worktrees/wt");
    write(main.path(), ".git/worktrees/wt/commondir", "../..\n");
    write(
        main.path(),
        ".git/worktrees/wt/gitdir",
        &format!("{}\n", wt.path().join(".git").display()),
    );
    write(wt.path(), ".git", &format!("gitdir: {}\n", admin.display()));
    let request = json!({"protocol_version": 1, "kind": "large_tree", "ignore_worktrees": true});

    // the line in the common exclude file, which every worktree shares, is found.
    write(main.path(), ".git/info/exclude", ".claude/worktrees/\n");
    let p = plan(wt.path(), request.clone()).unwrap();
    assert_eq!(row(&p, ".git/info/exclude", None).action, "unchanged");

    // a line only in the per-worktree admin folder is not the one git reads.
    write(main.path(), ".git/info/exclude", "# nothing\n");
    write(
        main.path(),
        ".git/worktrees/wt/info/exclude",
        ".claude/worktrees/\n",
    );
    let p = plan(wt.path(), request.clone()).unwrap();
    assert_eq!(row(&p, ".git/info/exclude", None).action, "merge");

    // an admin folder that does not name this worktree back is not trusted.
    write(main.path(), ".git/info/exclude", ".claude/worktrees/\n");
    write(main.path(), ".git/worktrees/wt/gitdir", "/elsewhere/.git\n");
    let p = plan(wt.path(), request).unwrap();
    assert_eq!(row(&p, ".git/info/exclude", None).action, "merge");
}

#[test]
fn a_picked_plugin_already_set_the_same_way_is_kept_and_the_rest_are_added() {
    let m = monorepo();
    let root = m.path();
    write(root, "pyproject.toml", "[project]\nname = \"tools\"\n");
    write(root, "tools/gen.py", "print(1)\n");
    let request = json!({
        "protocol_version": 1, "kind": "monorepo", "plugins": ["javascript", "python"]
    });
    // one picked plugin is already on, the other is new: a merge, not a conflict.
    write(
        root,
        ".claude/settings.json",
        "{\"enabledPlugins\": {\"typescript-lsp@claude-plugins-official\": true}}\n",
    );
    let before = snapshot(root);
    let p = plan(root, request.clone()).unwrap();
    assert_eq!(snapshot(root), before);
    let plug = row(&p, ".claude/settings.json", Some("enabledPlugins"));
    assert_eq!(plug.action, "merge", "{plug:#?}");
    assert_eq!(
        plug.value,
        json!({"typescript-lsp@claude-plugins-official": true,
               "pyright-lsp@claude-plugins-official": true})
    );
    assert!(
        plug.detail
            .as_deref()
            .is_some_and(|d| d.contains("1 picked already set the same way")),
        "{plug:#?}"
    );
    // both already on: unchanged.
    write(
        root,
        ".claude/settings.json",
        "{\"enabledPlugins\": {\"typescript-lsp@claude-plugins-official\": true, \
         \"pyright-lsp@claude-plugins-official\": true}}\n",
    );
    let both = plan(root, request.clone()).unwrap();
    let plug = row(&both, ".claude/settings.json", Some("enabledPlugins"));
    assert_eq!(plug.action, "unchanged", "{plug:#?}");
    // one set to another value: a conflict that names that plugin and only that one.
    write(
        root,
        ".claude/settings.json",
        "{\"enabledPlugins\": {\"typescript-lsp@claude-plugins-official\": false}}\n",
    );
    let p = plan(root, request).unwrap();
    let plug = row(&p, ".claude/settings.json", Some("enabledPlugins"));
    assert_eq!(plug.action, "conflict", "{plug:#?}");
    let detail = plug.detail.as_deref().unwrap_or_default();
    assert!(
        detail.contains("typescript-lsp@claude-plugins-official already set to another value"),
        "{plug:#?}"
    );
    assert!(!detail.contains("pyright"), "{plug:#?}");
}

#[test]
fn a_linked_worktree_pointer_is_read_up_to_4096_bytes_and_not_one_more() {
    let main = large_tree();
    let wt = TempDir::new().unwrap();
    write(wt.path(), "src/main.rs", "fn main() {}\n");
    let admin = main.path().join(".git/worktrees/wt");
    write(main.path(), ".git/worktrees/wt/commondir", "../..\n");
    write(
        main.path(),
        ".git/worktrees/wt/gitdir",
        &format!("{}\n", wt.path().join(".git").display()),
    );
    write(main.path(), ".git/info/exclude", ".claude/worktrees/\n");
    let request = json!({"protocol_version": 1, "kind": "large_tree", "ignore_worktrees": true});
    // the pointer's first line, then padding lines up to `len` bytes in all.
    let pointer = |len: usize| {
        let line = format!("gitdir: {}\n", admin.display());
        format!("{line}{}", "#".repeat(len - line.len()))
    };
    write(wt.path(), ".git", &pointer(4096));
    assert_eq!(fs::metadata(wt.path().join(".git")).unwrap().len(), 4096);
    let p = plan(wt.path(), request.clone()).unwrap();
    assert_eq!(row(&p, ".git/info/exclude", None).action, "unchanged");
    // one byte over the bound: the pointer is not trusted, so the common line is not found.
    write(wt.path(), ".git", &pointer(4097));
    assert_eq!(fs::metadata(wt.path().join(".git")).unwrap().len(), 4097);
    let p = plan(wt.path(), request).unwrap();
    assert_eq!(row(&p, ".git/info/exclude", None).action, "merge");
}

#[test]
fn a_dotted_key_under_a_parent_that_is_not_an_object_is_a_conflict_and_kept() {
    let m = monorepo();
    let root = m.path();
    // `worktree` holds a string, so `worktree.sparsePaths` has no object to go in.
    write(root, ".claude/settings.json", "{\"worktree\": \"head\"}\n");
    let before = snapshot(root);
    let p = plan(
        root,
        json!({"protocol_version": 1, "kind": "monorepo", "sparse_paths": ["packages/api"]}),
    )
    .unwrap();
    assert_eq!(snapshot(root), before);
    let sparse = row(&p, ".claude/settings.json", Some("worktree.sparsePaths"));
    assert_eq!(sparse.action, "conflict", "{sparse:#?}");
    assert!(
        sparse
            .detail
            .as_deref()
            .is_some_and(|d| d.contains("worktree is not an object")),
        "{sparse:#?}"
    );
}
