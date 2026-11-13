//! 08-B (#2013) GitHub templates through `github-plan` / `github-apply`.
//! Every test works in its own temporary project; no network, no `gh`, and GitHub's own
//! form rendering is not exercised (default-branch only: not_run). The AC commands name
//! `product_cli github_`; the same six fixture names live in this new file.
mod common;

use serde_json::{json, Value};
use std::{
    collections::BTreeMap,
    fs,
    path::{Path, PathBuf},
    process::{Command, Stdio},
};

fn cli(args: &[&str], input: &Value) -> Value {
    let mut child = Command::new(env!("CARGO_BIN_EXE_moai-cli"))
        .args(args)
        .stdin(Stdio::piped())
        .stdout(Stdio::piped())
        .stderr(Stdio::piped())
        .spawn()
        .unwrap();
    common::feed_stdin(&mut child, input.to_string().as_bytes());
    let out = child.wait_with_output().unwrap();
    serde_json::from_slice(&out.stdout).unwrap_or_else(|e| {
        panic!(
            "stdout is not JSON: {e}\n{}",
            String::from_utf8_lossy(&out.stdout)
        )
    })
}

fn plan(root: &Path, opt_in: bool) -> Value {
    let env = cli(
        &["github-plan", root.to_str().unwrap()],
        &json!({"protocol_version": 1, "github_issues": opt_in}),
    );
    assert_eq!(env["ok"], true, "{env}");
    env["result"].clone()
}

fn apply(root: &Path, plan: &Value) -> Value {
    cli(
        &[
            "github-apply",
            root.to_str().unwrap(),
            "--plan-digest",
            plan["plan_digest"].as_str().unwrap(),
            "--permission-reference",
            "test:08-B",
        ],
        plan,
    )
}

/// sha256 of every file under `dir`, keyed by relative path.
fn hashes(dir: &Path) -> BTreeMap<String, Vec<u8>> {
    let mut out = BTreeMap::new();
    let mut stack = vec![dir.to_path_buf()];
    while let Some(d) = stack.pop() {
        let Ok(entries) = fs::read_dir(&d) else {
            continue;
        };
        for e in entries.flatten() {
            let p = e.path();
            if p.is_dir() {
                stack.push(p);
            } else {
                let rel = p.strip_prefix(dir).unwrap().to_string_lossy().to_string();
                out.insert(rel, fs::read(&p).unwrap());
            }
        }
    }
    out
}

fn template(rel: &str) -> Vec<u8> {
    fs::read(
        Path::new(env!("CARGO_MANIFEST_DIR"))
            .join("templates/github")
            .join(rel),
    )
    .unwrap()
}

const TARGETS: [(&str, &str); 5] = [
    (
        ".github/ISSUE_TEMPLATE/moai-feature.yml",
        "ISSUE_TEMPLATE/moai-feature.yml",
    ),
    // Decision 43 (#2434 6842518327): the follow-up form ships too.
    (
        ".github/ISSUE_TEMPLATE/moai-followup.yml",
        "ISSUE_TEMPLATE/moai-followup.yml",
    ),
    (
        ".github/ISSUE_TEMPLATE/moai-bugfix.yml",
        "ISSUE_TEMPLATE/moai-bugfix.yml",
    ),
    (
        ".github/ISSUE_TEMPLATE/moai-investigation.yml",
        "ISSUE_TEMPLATE/moai-investigation.yml",
    ),
    (".github/moai/labels.yml", "labels.yml"),
];

fn project() -> (tempfile::TempDir, PathBuf) {
    let tmp = tempfile::tempdir().unwrap();
    let root = tmp.path().join("project");
    fs::create_dir_all(&root).unwrap();
    (tmp, root)
}

fn actions(plan: &Value) -> Vec<(String, String)> {
    plan["files"]
        .as_array()
        .unwrap()
        .iter()
        .map(|f| {
            (
                f["path"].as_str().unwrap().to_string(),
                f["action"].as_str().unwrap().to_string(),
            )
        })
        .collect()
}

#[test]
fn github_not_generated_without_profile_opt_in() {
    let (_tmp, root) = project();
    fs::create_dir_all(root.join(".github/ISSUE_TEMPLATE")).unwrap();
    fs::write(root.join(".github/ISSUE_TEMPLATE/bug.md"), "user\n").unwrap();
    let before = hashes(&root);
    let p = plan(&root, false);
    assert_eq!(p["files"], json!([]));
    assert!(p["notes"][0].as_str().unwrap().contains("not generated"));
    let env = apply(&root, &p);
    assert_eq!(env["ok"], true, "{env}");
    assert_eq!(env["result"]["written"], json!([]));
    assert_eq!(
        before,
        hashes(&root),
        "only the receipt directory may appear"
    );
}

#[test]
fn github_clean_project_creates_only_selected_files() {
    let (_tmp, root) = project();
    let p = plan(&root, true);
    assert!(actions(&p).iter().all(|(_, a)| a == "create"), "{p}");
    let env = apply(&root, &p);
    assert_eq!(env["ok"], true, "{env}");
    let files = hashes(&root.join(".github"));
    assert_eq!(files.len(), TARGETS.len(), "{:?}", files.keys());
    for (target, tpl) in TARGETS {
        assert_eq!(
            fs::read(root.join(target)).unwrap(),
            template(tpl),
            "{target}"
        );
    }
    assert!(!root.join(".github/ISSUE_TEMPLATE/config.yml").exists());
    let manifest: Value =
        serde_json::from_slice(&fs::read(root.join(".moai-cli/manifests/github.json")).unwrap())
            .unwrap();
    assert_eq!(
        manifest["owned_files"].as_array().unwrap().len(),
        TARGETS.len()
    );
    assert!(env["result"]["rendering"]
        .as_str()
        .unwrap()
        .starts_with("not_run"));
}

#[test]
fn github_second_run_is_unchanged() {
    let (_tmp, root) = project();
    let first = plan(&root, true);
    assert_eq!(apply(&root, &first)["ok"], true);
    let before = hashes(&root.join(".github"));
    let manifest_before = fs::read(root.join(".moai-cli/manifests/github.json")).unwrap();
    let second = plan(&root, true);
    assert!(
        actions(&second).iter().all(|(_, a)| a == "unchanged"),
        "{second}"
    );
    assert!(second["files"]
        .as_array()
        .unwrap()
        .iter()
        .all(|f| f["adopt_baseline"] == false));
    let env = apply(&root, &second);
    assert_eq!(env["result"]["written"], json!([]));
    assert_eq!(before, hashes(&root.join(".github")));
    assert_eq!(
        manifest_before,
        fs::read(root.join(".moai-cli/manifests/github.json")).unwrap()
    );
    // The plan itself is byte-stable for equal input.
    assert_eq!(plan(&root, true), second);
}

#[test]
fn github_existing_user_templates_are_preserved_and_unowned_is_conflict() {
    let (_tmp, root) = project();
    let forms = root.join(".github/ISSUE_TEMPLATE");
    fs::create_dir_all(&forms).unwrap();
    fs::write(forms.join("bug.md"), "---\nname: Bug\n---\n").unwrap();
    fs::write(
        forms.join("config.yml"),
        "blank_issues_enabled: true\nx-unknown-key: kept\n",
    )
    .unwrap();
    // Without a clash, user files are listed as preserved and stay byte-identical.
    let user_before = hashes(&root.join(".github"));
    let p = plan(&root, true);
    let preserved: Vec<&str> = p["preserved"]
        .as_array()
        .unwrap()
        .iter()
        .map(|v| v.as_str().unwrap())
        .collect();
    assert_eq!(
        preserved,
        [
            ".github/ISSUE_TEMPLATE/bug.md",
            ".github/ISSUE_TEMPLATE/config.yml"
        ]
    );
    assert_eq!(apply(&root, &p)["ok"], true);
    let after = hashes(&root.join(".github"));
    for (k, v) in &user_before {
        assert_eq!(after.get(k), Some(v), "{k} changed");
    }
    // A user file at a target path without a baseline is a conflict; nothing is written.
    let (_tmp2, root2) = project();
    let forms2 = root2.join(".github/ISSUE_TEMPLATE");
    fs::create_dir_all(&forms2).unwrap();
    fs::write(forms2.join("moai-feature.yml"), "name: mine\n").unwrap();
    fs::write(forms2.join("bug.md"), "user\n").unwrap();
    let before = hashes(&root2);
    let p = plan(&root2, true);
    let conflict: Vec<_> = actions(&p)
        .into_iter()
        .filter(|(_, a)| a == "conflict")
        .collect();
    assert_eq!(
        conflict,
        [(
            ".github/ISSUE_TEMPLATE/moai-feature.yml".to_string(),
            "conflict".to_string()
        )]
    );
    let env = apply(&root2, &p);
    assert_eq!(env["ok"], false);
    assert_eq!(env["error"]["kind"], "conflict");
    assert_eq!(before, hashes(&root2), "nothing applied");
}

#[test]
fn github_conflicting_labels_manifest_is_reported_not_merged() {
    let (_tmp, root) = project();
    fs::create_dir_all(root.join(".github/moai")).unwrap();
    fs::write(
        root.join(".github/moai/labels.yml"),
        "schema_version: 1\nadded: []\n",
    )
    .unwrap();
    let before = hashes(&root);
    let p = plan(&root, true);
    let labels = p["files"]
        .as_array()
        .unwrap()
        .iter()
        .find(|f| f["path"] == ".github/moai/labels.yml")
        .unwrap();
    assert_eq!(labels["action"], "conflict");
    assert!(labels["detail"]
        .as_str()
        .unwrap()
        .contains("no recorded baseline"));
    let env = apply(&root, &p);
    assert_eq!(env["ok"], false);
    assert!(env["error"]["message"]
        .as_str()
        .unwrap()
        .contains(".github/moai/labels.yml"));
    assert_eq!(before, hashes(&root));
}

#[test]
fn github_interrupted_update_recovers_baseline() {
    let (_tmp, root) = project();
    // The files were written and the manifest was not (an apply interrupted between them).
    for (target, tpl) in TARGETS {
        let path = root.join(target);
        fs::create_dir_all(path.parent().unwrap()).unwrap();
        fs::write(&path, template(tpl)).unwrap();
    }
    let p = plan(&root, true);
    assert!(actions(&p).iter().all(|(_, a)| a == "unchanged"), "{p}");
    assert!(p["files"]
        .as_array()
        .unwrap()
        .iter()
        .all(|f| f["adopt_baseline"] == true));
    let env = apply(&root, &p);
    assert_eq!(env["ok"], true, "{env}");
    assert_eq!(
        env["result"]["adopted_baselines"].as_array().unwrap().len(),
        TARGETS.len()
    );
    assert!(root.join(".moai-cli/manifests/github.json").exists());
    let again = plan(&root, true);
    assert!(again["files"]
        .as_array()
        .unwrap()
        .iter()
        .all(|f| f["adopt_baseline"] == false && f["action"] == "unchanged"));
    // A later user edit of an owned file is kept: the template did not move, so the
    // 3-way decision keeps the current bytes and apply writes nothing.
    fs::write(root.join(".github/moai/labels.yml"), "edited\n").unwrap();
    let edited = plan(&root, true);
    let labels = edited["files"]
        .as_array()
        .unwrap()
        .iter()
        .find(|f| f["path"] == ".github/moai/labels.yml")
        .unwrap()
        .clone();
    assert_eq!(labels["action"], "unchanged");
    assert_eq!(apply(&root, &edited)["result"]["written"], json!([]));
    assert_eq!(
        fs::read(root.join(".github/moai/labels.yml")).unwrap(),
        b"edited\n"
    );
}

#[test]
fn github_apply_refuses_stale_or_unpermitted_plans() {
    let (_tmp, root) = project();
    let p = plan(&root, true);
    let no_permission = cli(
        &[
            "github-apply",
            root.to_str().unwrap(),
            "--plan-digest",
            p["plan_digest"].as_str().unwrap(),
            "--permission-reference",
            "placeholder",
        ],
        &p,
    );
    assert_eq!(no_permission["ok"], false);
    fs::create_dir_all(root.join(".github/moai")).unwrap();
    fs::write(root.join(".github/moai/labels.yml"), "late\n").unwrap();
    let stale = apply(&root, &p);
    assert_eq!(stale["ok"], false);
    assert_eq!(stale["error"]["kind"], "stale_plan");
    assert!(!root.join(".github/ISSUE_TEMPLATE").exists());
}
