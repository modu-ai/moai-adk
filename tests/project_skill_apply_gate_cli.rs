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
