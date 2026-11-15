//! `worktree-install-plan` (Issue #3010, P6a): a read-only preview of the command that would
//! install only the packages one profile scope names.
//!
//! The command is never run: no process is started and nothing is written, because installing
//! packages fetches them over the network. The text shown is built from the package manager's
//! own documentation (read 2026-11-15, `target/lane/plan3010_p6.md`). What a document does not
//! say is listed under `unverified` instead of being filled in, and a manager with no way to
//! install a subset (yarn 1, go) is reported as a full install, not invented.

use crate::scopes::ScopeEntry;
use serde_json::{json, Value};
use std::fs;
use std::io::Read;
use std::path::Path;

type Result<T> = std::result::Result<T, String>;

/// Longest manifest or lockfile prefix read, in bytes.
const READ_BYTES: u64 = 256 * 1024;

#[derive(Clone, Copy, PartialEq, Eq, Debug)]
enum Manager {
    Npm,
    Pnpm,
    YarnClassic,
    YarnBerry,
    Bun,
    Cargo,
    Go,
}

impl Manager {
    fn name(self) -> &'static str {
        match self {
            Manager::Npm => "npm",
            Manager::Pnpm => "pnpm",
            Manager::YarnClassic => "yarn1",
            Manager::YarnBerry => "yarn",
            Manager::Bun => "bun",
            Manager::Cargo => "cargo",
            Manager::Go => "go",
        }
    }
}

/// The text of `path` when it is a plain file (not a link), at most [`READ_BYTES`] of it.
fn plain_file_text(path: &Path) -> Option<String> {
    let meta = fs::symlink_metadata(path).ok()?;
    if !meta.is_file() {
        return None;
    }
    let mut buf = Vec::new();
    fs::File::open(path)
        .ok()?
        .take(READ_BYTES)
        .read_to_end(&mut buf)
        .ok()?;
    Some(String::from_utf8_lossy(&buf).into_owned())
}

fn is_plain_file(path: &Path) -> bool {
    fs::symlink_metadata(path).is_ok_and(|m| m.is_file())
}

fn is_plain_dir(path: &Path) -> bool {
    fs::symlink_metadata(path).is_ok_and(|m| m.is_dir())
}

/// The lockfiles at the project root, as `(file name, manager)`. `None` for the manager means
/// the file is a `yarn.lock` that is neither generation. `bun.lock` and `bun.lockb` are one
/// manager and are listed once.
fn lockfiles(root: &Path) -> Vec<(&'static str, Option<Manager>)> {
    let mut found = Vec::new();
    for (file, manager) in [
        ("package-lock.json", Manager::Npm),
        ("pnpm-lock.yaml", Manager::Pnpm),
        ("Cargo.lock", Manager::Cargo),
        ("go.sum", Manager::Go),
    ] {
        if is_plain_file(&root.join(file)) {
            found.push((file, Some(manager)));
        }
    }
    for file in ["bun.lock", "bun.lockb"] {
        if is_plain_file(&root.join(file)) {
            found.push((file, Some(Manager::Bun)));
            break;
        }
    }
    if let Some(text) = plain_file_text(&root.join("yarn.lock")) {
        let head: String = text.chars().take(4096).collect();
        let manager = if head.contains("__metadata:") {
            Some(Manager::YarnBerry)
        } else if head.contains("yarn lockfile v1") {
            Some(Manager::YarnClassic)
        } else {
            None
        };
        found.push(("yarn.lock", manager));
    }
    found
}

/// Whether `name` is a package name safe to show in a command: letters, digits and `@ / . _ -`.
fn name_ok(name: &str) -> bool {
    !name.is_empty()
        && name.len() <= 214
        && !name.starts_with('-')
        && name
            .bytes()
            .all(|b| b.is_ascii_alphanumeric() || matches!(b, b'@' | b'/' | b'.' | b'_' | b'-'))
}

/// Single-quotes `text` for a shell. Folder paths cannot hold `'` (the scopes check refuses
/// it) and package names are checked by [`name_ok`], so no quote ever needs escaping here.
fn quote(text: &str) -> String {
    format!("'{text}'")
}

/// The `name` of `<dir>/package.json`.
fn npm_name(dir: &Path) -> Option<String> {
    let text = plain_file_text(&dir.join("package.json"))?;
    let name = serde_json::from_str::<Value>(&text)
        .ok()?
        .get("name")?
        .as_str()?
        .to_string();
    name_ok(&name).then_some(name)
}

/// The `[package] name` of `<dir>/Cargo.toml`.
fn cargo_name(dir: &Path) -> Option<String> {
    let text = plain_file_text(&dir.join("Cargo.toml"))?;
    let doc = text.parse::<toml_edit::DocumentMut>().ok()?;
    let name = doc.get("package")?.get("name")?.as_str()?.to_string();
    name_ok(&name).then_some(name)
}

/// The preview for one scope in `root`. Reads only.
pub fn preview(root: &Path, scope: &ScopeEntry) -> Result<Value> {
    let mut problems: Vec<&str> = Vec::new();
    for p in &scope.paths {
        if !is_plain_dir(&root.join(p)) {
            problems.push(p);
        }
    }
    let found = lockfiles(root);
    let base = |mode: &str, manager: Option<&str>, lock: Option<&str>| {
        json!({
            "scope": scope.name,
            "mode": mode,
            "manager": manager,
            "lockfile": lock,
            "paths": scope.paths,
            "commands": [],
            "unverified": [],
            "ran": false,
        })
    };
    if !problems.is_empty() {
        let mut v = base("refused", None, None);
        v["reason"] = json!("a scope folder is missing or is not a plain folder");
        v["problem_paths"] = json!(problems);
        return Ok(v);
    }
    let names: Vec<&str> = found.iter().map(|(f, _)| *f).collect();
    let [(lock, Some(manager))] = found[..] else {
        let mut v = base("undecidable", None, None);
        v["reason"] = json!(if found.is_empty() {
            "no known lockfile at the project root"
        } else if found.len() > 1 {
            "more than one lockfile at the project root"
        } else {
            "yarn.lock is neither a yarn 1 nor a yarn 2+ lockfile"
        });
        v["lockfiles"] = json!(names);
        return Ok(v);
    };

    let mut v = base("scoped", Some(manager.name()), Some(lock));
    let paths: Vec<&String> = scope.paths.iter().collect();
    let by_name =
        |lookup: fn(&Path) -> Option<String>| -> std::result::Result<Vec<String>, Vec<String>> {
            let mut names = Vec::new();
            let mut missing = Vec::new();
            for p in &paths {
                match lookup(&root.join(p)) {
                    Some(n) => names.push(n),
                    None => missing.push((*p).clone()),
                }
            }
            if missing.is_empty() {
                Ok(names)
            } else {
                Err(missing)
            }
        };
    let flags = |flag: &str| {
        paths
            .iter()
            .map(|p| format!("{flag} {}", quote(p)))
            .collect::<Vec<_>>()
            .join(" ")
    };
    let mut unverified: Vec<&str> = Vec::new();
    let mut commands: Vec<String> = Vec::new();
    let mut unnamed: Vec<String> = Vec::new();
    match manager {
        Manager::Npm => {
            commands.push(format!("npm ci {}", flags("-w")));
            unverified.extend([
                "whether the -w selection on npm ci installs only the selected workspaces",
                "the lowest npm version that accepts -w on npm ci",
            ]);
        }
        Manager::Pnpm => {
            commands.push(format!(
                "pnpm install --frozen-lockfile {}",
                flags("--filter")
            ));
            unverified.extend([
                "whether pnpm install --filter also installs the root package",
                "the lowest pnpm version that accepts this form",
            ]);
        }
        Manager::YarnBerry => match by_name(npm_name) {
            Ok(names) => {
                let list: Vec<String> = names.iter().map(|n| quote(n)).collect();
                commands.push(format!("yarn workspaces focus {}", list.join(" ")));
                unverified.extend([
                    "whether yarn workspaces focus accepts --immutable",
                    "the lowest yarn version that has workspaces focus",
                ]);
            }
            Err(missing) => unnamed = missing,
        },
        Manager::Bun => {
            commands.push(format!(
                "bun install --frozen-lockfile {}",
                flags("--filter")
            ));
            unverified.extend([
                "an example that combines --filter with --frozen-lockfile",
                "the lowest bun version that accepts this form",
            ]);
        }
        Manager::Cargo => match by_name(cargo_name) {
            Ok(names) => {
                let list: Vec<String> = names.iter().map(|n| format!("-p {}", quote(n))).collect();
                commands.push(format!("cargo build --locked {}", list.join(" ")));
                unverified.push("whether cargo fetch downloads only the selected members");
            }
            Err(missing) => unnamed = missing,
        },
        Manager::YarnClassic => {
            v["mode"] = json!("full_only");
            v["reason"] = json!("yarn 1 has no way to install a subset from the root");
            commands.push("yarn install --frozen-lockfile".into());
        }
        Manager::Go => {
            v["mode"] = json!("full_only");
            v["reason"] = json!("go cannot select workspace members for go mod download");
            commands.push("go mod download".into());
        }
    }
    if !unnamed.is_empty() {
        v["mode"] = json!("needs_package_names");
        v["reason"] = json!("this manager selects by package name and a scope folder has none");
        v["unnamed_paths"] = json!(unnamed);
    }
    v["commands"] = json!(commands);
    v["unverified"] = json!(unverified);
    v["note"] = json!("preview only: no command was run and nothing was written");
    Ok(v)
}

/// The preview for the scope `name` declared in the project profile. Reads only.
pub fn plan(root: &Path, name: &str) -> Result<Value> {
    let (profile, _) = crate::profile::read_profile(root)?;
    let scopes = profile
        .scopes
        .ok_or("the profile declares no scopes (profile key `scopes`)")?;
    let scope = scopes
        .iter()
        .find(|s| s.name == name)
        .ok_or("the profile declares no scope with that name")?;
    preview(root, scope)
}

#[cfg(test)]
mod tests {
    use super::*;
    use tempfile::TempDir;

    fn scope(paths: &[&str]) -> ScopeEntry {
        ScopeEntry {
            name: "web".into(),
            paths: paths.iter().map(|p| p.to_string()).collect(),
        }
    }

    fn project(files: &[(&str, &str)]) -> TempDir {
        let root = TempDir::new().unwrap();
        for (path, body) in files {
            let p = root.path().join(path);
            fs::create_dir_all(p.parent().unwrap()).unwrap();
            fs::write(p, body).unwrap();
        }
        root
    }

    fn commands(v: &Value) -> Vec<String> {
        v["commands"]
            .as_array()
            .unwrap()
            .iter()
            .map(|c| c.as_str().unwrap().to_string())
            .collect()
    }

    #[test]
    fn every_manager_gets_its_documented_command_shape() {
        let two = scope(&["apps/web", "packages/ui"]);
        let dirs = [
            ("apps/web/package.json", r#"{"name":"@acme/web"}"#),
            ("packages/ui/package.json", r#"{"name":"ui"}"#),
        ];
        let cases: [(&str, &str, &str); 4] = [
            (
                "package-lock.json",
                "{}",
                "npm ci -w 'apps/web' -w 'packages/ui'",
            ),
            (
                "pnpm-lock.yaml",
                "",
                "pnpm install --frozen-lockfile --filter 'apps/web' --filter 'packages/ui'",
            ),
            (
                "bun.lock",
                "",
                "bun install --frozen-lockfile --filter 'apps/web' --filter 'packages/ui'",
            ),
            (
                "yarn.lock",
                "__metadata:\n  version: 8\n",
                "yarn workspaces focus '@acme/web' 'ui'",
            ),
        ];
        for (lock, body, want) in cases {
            let mut files = dirs.to_vec();
            files.push((lock, body));
            let root = project(&files);
            let v = preview(root.path(), &two).unwrap();
            assert_eq!(v["mode"], "scoped", "{lock}");
            assert_eq!(commands(&v), [want], "{lock}");
            assert_eq!(v["ran"], false);
            assert!(!v["unverified"].as_array().unwrap().is_empty(), "{lock}");
        }
    }

    #[test]
    fn cargo_selects_members_by_package_name() {
        let root = project(&[
            ("Cargo.lock", ""),
            ("crates/a/Cargo.toml", "[package]\nname = \"a-core\"\n"),
        ]);
        let v = preview(root.path(), &scope(&["crates/a"])).unwrap();
        assert_eq!(commands(&v), ["cargo build --locked -p 'a-core'"]);
    }

    /// Mutation survivor: every selected member is named, in the order of the scope's folders.
    #[test]
    fn cargo_names_every_selected_member() {
        let root = project(&[
            ("Cargo.lock", ""),
            ("crates/a/Cargo.toml", "[package]\nname = \"a-core\"\n"),
            ("crates/b/Cargo.toml", "[package]\nname = \"b-core\"\n"),
        ]);
        let v = preview(root.path(), &scope(&["crates/a", "crates/b"])).unwrap();
        assert_eq!(
            commands(&v),
            ["cargo build --locked -p 'a-core' -p 'b-core'"]
        );
    }

    /// Mutation survivor, and the edges of the length limit: only letters, digits and
    /// `@ / . _ -` are shown in a command, so no shell word can come out of a manifest.
    #[test]
    fn a_shown_package_name_is_plain_text_within_its_limit() {
        for ok in ["a", "@scope/pkg.name_x-1", "A9"] {
            assert!(name_ok(ok), "{ok}");
        }
        for no in [
            "", "-x", "a b", "a;b", "a$b", "a`b", "a'b", "a\"b", "a&b", "a|b", "a\nb", "a(b",
            "a>b", "a\\b", "a*b",
        ] {
            assert!(!name_ok(no), "{no:?}");
        }
        assert!(name_ok(&"a".repeat(214)));
        assert!(!name_ok(&"a".repeat(215)));
    }

    #[test]
    fn yarn_one_and_go_are_reported_as_full_installs() {
        let root = project(&[("yarn.lock", "# yarn lockfile v1\n"), ("apps/web/x", "")]);
        let v = preview(root.path(), &scope(&["apps/web"])).unwrap();
        assert_eq!(v["mode"], "full_only");
        assert_eq!(commands(&v), ["yarn install --frozen-lockfile"]);
        let root = project(&[("go.sum", ""), ("apps/web/x", "")]);
        let v = preview(root.path(), &scope(&["apps/web"])).unwrap();
        assert_eq!(v["mode"], "full_only");
        assert_eq!(commands(&v), ["go mod download"]);
    }

    #[test]
    fn no_lockfile_two_lockfiles_or_an_unknown_yarn_lock_decide_nothing() {
        let dir = ("apps/web/x", "");
        for (files, why) in [
            (vec![dir], "no known lockfile"),
            (
                vec![dir, ("package-lock.json", ""), ("pnpm-lock.yaml", "")],
                "more than one",
            ),
            (vec![dir, ("yarn.lock", "something else")], "neither"),
        ] {
            let root = project(&files);
            let v = preview(root.path(), &scope(&["apps/web"])).unwrap();
            assert_eq!(v["mode"], "undecidable");
            assert!(commands(&v).is_empty());
            assert!(v["reason"].as_str().unwrap().contains(why), "{why}");
        }
    }

    #[test]
    fn bun_text_and_binary_lockfiles_are_one_manager() {
        let root = project(&[("bun.lock", ""), ("bun.lockb", ""), ("apps/web/x", "")]);
        let v = preview(root.path(), &scope(&["apps/web"])).unwrap();
        assert_eq!(v["manager"], "bun");
        assert_eq!(v["mode"], "scoped");
    }

    #[test]
    fn a_missing_folder_or_a_folder_without_a_package_name_makes_no_command() {
        let root = project(&[("pnpm-lock.yaml", "")]);
        let v = preview(root.path(), &scope(&["apps/web"])).unwrap();
        assert_eq!(v["mode"], "refused");
        assert_eq!(v["problem_paths"], json!(["apps/web"]));
        assert!(commands(&v).is_empty());

        let root = project(&[
            ("yarn.lock", "__metadata:\n"),
            ("apps/web/package.json", "{}"),
        ]);
        let v = preview(root.path(), &scope(&["apps/web"])).unwrap();
        assert_eq!(v["mode"], "needs_package_names");
        assert_eq!(v["unnamed_paths"], json!(["apps/web"]));
        assert!(commands(&v).is_empty());
    }

    #[test]
    fn a_package_name_that_is_not_plain_text_is_not_shown() {
        let root = project(&[
            ("yarn.lock", "__metadata:\n"),
            ("apps/web/package.json", r#"{"name":"x; rm -rf ~"}"#),
        ]);
        let v = preview(root.path(), &scope(&["apps/web"])).unwrap();
        assert_eq!(v["mode"], "needs_package_names");
        assert!(commands(&v).is_empty());
    }

    #[cfg(unix)]
    #[test]
    fn a_linked_folder_or_lockfile_is_not_followed() {
        let root = project(&[("elsewhere/package-lock.json", "{}"), ("real/x", "")]);
        std::os::unix::fs::symlink(
            root.path().join("elsewhere/package-lock.json"),
            root.path().join("package-lock.json"),
        )
        .unwrap();
        std::os::unix::fs::symlink(root.path().join("real"), root.path().join("apps")).unwrap();
        let v = preview(root.path(), &scope(&["apps"])).unwrap();
        assert_eq!(v["mode"], "refused");
        let ok = project(&[("real/x", "")]);
        std::os::unix::fs::symlink(
            ok.path().join("real/x"),
            ok.path().join("package-lock.json"),
        )
        .unwrap();
        let v = preview(ok.path(), &scope(&["real"])).unwrap();
        assert_eq!(v["mode"], "undecidable");
    }

    #[test]
    fn a_manifest_is_read_only_up_to_its_limit() {
        let big = format!("{{\"name\":\"ok\",\"pad\":\"{}\"}}", "x".repeat(300_000));
        let root = project(&[
            ("yarn.lock", "__metadata:\n"),
            ("apps/web/package.json", &big),
        ]);
        let v = preview(root.path(), &scope(&["apps/web"])).unwrap();
        assert_eq!(v["mode"], "needs_package_names");
    }
}
