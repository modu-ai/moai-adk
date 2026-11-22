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
    let before = snapshot(root);
    let p = plan(
        root,
        json!({
            "protocol_version": 1, "kind": "monorepo",
            "package_files": ["packages/api", "packages/web"],
            "exclude_packages": ["packages/legacy"],
            "deny_read": ["generated"],
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
    assert_eq!(deny.value, json!(["Read(./generated/**/*)"]));
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
    // REQ-7: a self-contained package settings file naming only siblings that exist; the deny
    // rules stay in the root file, which a worktree session reads.
    let own = row(&p, "packages/api/.claude/settings.json", None);
    assert_eq!(own.action, "create");
    assert_eq!(
        own.value,
        json!({"permissions": {"additionalDirectories": ["../legacy", "../shared", "../web"]}})
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
