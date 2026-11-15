//! #2936 R9 S5c: `project-skill-apply` takes an optional plan digest and permission reference
//! (the shape `project-update-plan` prints). With them it applies only the previewed plan to
//! the previewed project; without them it behaves as before. Temporary folders only; the
//! real HOME is never read or written.

mod common;
use common::cli::cli;
use serde_json::{json, Value};
use std::{fs, path::Path};
use tempfile::TempDir;

const PERMISSION: &str = "fixture-approved";

fn write(path: &Path, body: &str) {
    fs::create_dir_all(path.parent().unwrap()).unwrap();
    fs::write(path, body).unwrap();
}

struct Fx {
    _tmp: TempDir,
    a: String,
    b: String,
    source: String,
}

fn fx() -> Fx {
    let tmp = TempDir::new().unwrap();
    let root = fs::canonicalize(tmp.path()).unwrap();
    let (a, b, source) = (root.join("a"), root.join("b"), root.join("source"));
    fs::create_dir_all(&a).unwrap();
    fs::create_dir_all(&b).unwrap();
    write(
        &source.join("SKILL.md"),
        "---\nname: moai\ndescription: fixture\n---\n# MoAI\n",
    );
    write(&source.join("references/routes.json"), "{\"routes\":1}\n");
    let s = |p: &Path| p.to_str().unwrap().to_string();
    Fx {
        a: s(&a),
        b: s(&b),
        source: s(&source),
        _tmp: tmp,
    }
}

impl Fx {
    fn request(&self) -> Value {
        json!({
            "protocol_version": 1, "opt_in": true, "project_root": "ignored",
            "home": null, "hosts": ["codex"], "source_dir": self.source,
        })
    }

    fn plan(&self, project: &str) -> Value {
        let out = cli(&["project-skill-plan", project], &self.request());
        let v: Value = serde_json::from_slice(&out.stdout).unwrap();
        assert_eq!(v["ok"], json!(true), "{v}");
        v["result"].clone()
    }

    fn apply(&self, project: &str, extra: &[&str]) -> Value {
        let mut args = vec!["project-skill-apply", project];
        args.extend_from_slice(extra);
        let out = cli(&args, &self.request());
        serde_json::from_slice(&out.stdout)
            .unwrap_or_else(|_| panic!("{args:?}: {}", String::from_utf8_lossy(&out.stderr)))
    }
}

fn digest(plan: &Value) -> String {
    plan["plan_digest"]
        .as_str()
        .unwrap_or_else(|| panic!("no plan_digest: {plan}"))
        .to_string()
}

fn copy_of(project: &str) -> std::path::PathBuf {
    Path::new(project).join(".agents/skills/moai/SKILL.md")
}

#[test]
fn the_plan_prints_a_digest_and_writes_nothing() {
    let f = fx();
    let plan = f.plan(&f.a);
    assert!(digest(&plan).starts_with("sha256:"), "{plan}");
    assert!(!Path::new(&f.a).join(".agents").exists());
}

#[test]
fn the_previewed_plan_applies_with_its_digest_and_a_permission_reference() {
    let f = fx();
    let d = digest(&f.plan(&f.a));
    let out = f.apply(
        &f.a,
        &["--plan-digest", &d, "--permission-reference", PERMISSION],
    );
    assert_eq!(out["ok"], json!(true), "{out}");
    assert!(copy_of(&f.a).is_file());
    let again = f.plan(&f.a);
    assert_eq!(again["entries"][0]["action"], "unchanged", "{again}");
}

#[test]
fn a_wrong_digest_or_an_empty_permission_writes_nothing() {
    let f = fx();
    let d = digest(&f.plan(&f.a));
    let stale = f.apply(
        &f.a,
        &[
            "--plan-digest",
            &format!("sha256:{}", "0".repeat(64)),
            "--permission-reference",
            PERMISSION,
        ],
    );
    assert_eq!(stale["ok"], json!(false), "{stale}");
    let empty = f.apply(&f.a, &["--plan-digest", &d, "--permission-reference", ""]);
    assert_eq!(empty["ok"], json!(false), "{empty}");
    assert!(!Path::new(&f.a).join(".agents").exists());
}

#[test]
fn one_flag_without_the_other_is_a_usage_error_and_writes_nothing() {
    let f = fx();
    let d = digest(&f.plan(&f.a));
    let only_digest = f.apply(&f.a, &["--plan-digest", &d]);
    assert_eq!(only_digest["ok"], json!(false), "{only_digest}");
    let only_ref = f.apply(&f.a, &["--permission-reference", PERMISSION]);
    assert_eq!(only_ref["ok"], json!(false), "{only_ref}");
    assert!(!Path::new(&f.a).join(".agents").exists());
}

#[test]
fn a_plan_of_one_project_is_not_a_plan_of_another() {
    let f = fx();
    let d = digest(&f.plan(&f.a));
    let out = f.apply(
        &f.b,
        &["--plan-digest", &d, "--permission-reference", PERMISSION],
    );
    assert_eq!(out["ok"], json!(false), "{out}");
    assert!(!Path::new(&f.b).join(".agents").exists());
}

#[test]
fn a_package_changed_after_the_preview_is_refused() {
    let f = fx();
    let d = digest(&f.plan(&f.a));
    write(
        &Path::new(&f.source).join("references/routes.json"),
        "{\"routes\":2}\n",
    );
    let out = f.apply(
        &f.a,
        &["--plan-digest", &d, "--permission-reference", PERMISSION],
    );
    assert_eq!(out["ok"], json!(false), "{out}");
    assert!(!Path::new(&f.a).join(".agents").exists());
}

#[test]
fn the_old_call_without_the_flags_still_applies() {
    let f = fx();
    let out = f.apply(&f.a, &[]);
    assert_eq!(out["ok"], json!(true), "{out}");
    assert!(copy_of(&f.a).is_file());
}

/// Every entry under `root` (relative path, and the bytes of a file); nothing is followed.
fn tree(root: &str) -> Vec<(String, Option<Vec<u8>>)> {
    fn walk(base: &Path, dir: &Path, out: &mut Vec<(String, Option<Vec<u8>>)>) {
        let mut names: Vec<_> = fs::read_dir(dir).unwrap().map(|e| e.unwrap()).collect();
        names.sort_by_key(|e| e.file_name());
        for e in names {
            let rel = e.path().strip_prefix(base).unwrap().display().to_string();
            let meta = fs::symlink_metadata(e.path()).unwrap();
            if meta.is_dir() {
                out.push((rel, None));
                walk(base, &e.path(), out);
            } else {
                out.push((rel, Some(fs::read(e.path()).unwrap_or_default())));
            }
        }
    }
    let mut out = Vec::new();
    walk(Path::new(root), Path::new(root), &mut out);
    out
}

fn kind(out: &Value) -> String {
    out["error"]["kind"]
        .as_str()
        .unwrap_or_default()
        .to_string()
}

/// #2998 audit follow-up 3: a digest that is almost right is not right. Each variant keeps the
/// real digest's first characters, or its letters, or only adds blank space, or is the package
/// digest instead of the plan digest; every one is a stale plan, and the project tree is the
/// same afterwards (no rollback folder, no state folder).
#[test]
fn a_digest_that_is_only_almost_right_is_refused_and_the_project_is_untouched() {
    let f = fx();
    let plan = f.plan(&f.a);
    let d = digest(&plan);
    let package = plan["package_digest"].as_str().unwrap().to_string();
    let mut tail = d.clone();
    tail.pop();
    tail.push(if d.ends_with('0') { '1' } else { '0' });
    let variants = [
        ("last character changed", tail),
        ("upper case", d.to_ascii_uppercase()),
        ("trailing newline", format!("{d}\n")),
        ("leading space", format!(" {d}")),
        ("the package digest", package),
        ("first 20 characters only", d[..20].to_string()),
    ];
    let before = tree(&f.a);
    for (what, bad) in variants {
        let out = f.apply(
            &f.a,
            &["--plan-digest", &bad, "--permission-reference", PERMISSION],
        );
        assert_eq!(out["ok"], json!(false), "{what}: {out}");
        assert_eq!(kind(&out), "stale_plan", "{what}: {out}");
        assert_eq!(before, tree(&f.a), "{what}: the project changed");
    }
}

/// A reference that says nothing is refused with its own kind, whatever the digest, and nothing
/// is written anywhere in the project.
#[test]
fn an_empty_null_or_placeholder_reference_is_refused_with_its_own_kind() {
    let f = fx();
    let d = digest(&f.plan(&f.a));
    let before = tree(&f.a);
    for bad in ["", "null", "placeholder", " null ", "a\u{7}b"] {
        let out = f.apply(&f.a, &["--plan-digest", &d, "--permission-reference", bad]);
        assert_eq!(out["ok"], json!(false), "{bad:?}: {out}");
        assert_eq!(kind(&out), "permission_reference_missing", "{bad:?}: {out}");
        assert_eq!(before, tree(&f.a), "{bad:?}: the project changed");
    }
}

/// The digest binds the project by its real path: the same project reached through a link has
/// the same plan, so a plan made through one name applies through the other.
#[cfg(unix)]
#[test]
fn a_plan_made_through_a_link_to_the_project_applies_to_the_project() {
    let f = fx();
    let alias = Path::new(&f.b).join("alias");
    std::os::unix::fs::symlink(&f.a, &alias).unwrap();
    let alias = alias.to_str().unwrap().to_string();
    let via_link = digest(&f.plan(&alias));
    assert_eq!(via_link, digest(&f.plan(&f.a)));
    let out = f.apply(
        &f.a,
        &[
            "--plan-digest",
            &via_link,
            "--permission-reference",
            PERMISSION,
        ],
    );
    assert_eq!(out["ok"], json!(true), "{out}");
    assert!(copy_of(&f.a).is_file());
}
