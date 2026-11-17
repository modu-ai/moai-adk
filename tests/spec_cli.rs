//! SDD-2 (SPEC-20261117-spec-tooling): `moai spec new|check|list|show` in a temporary project
//! under a temporary HOME (`TempHome`); the real HOME and real projects are never read. Every
//! case compares the whole tree before and after: `new` adds only its SPEC folder and file, the
//! other actions and every refusal add nothing.

mod common;

use common::cli::tree_state;
use common::home::TempHome;
use serde_json::{json, Value};
use std::fs;
use std::path::{Path, PathBuf};

const ID: &str = "SPEC-20261117-login-fix";

/// A project with this generation's `.moai/` marker.
fn project(home: &TempHome) -> PathBuf {
    let p = home.path().join("proj");
    fs::create_dir_all(p.join(".moai")).unwrap();
    fs::write(
        p.join(".moai/moai.json"),
        moai_cli::state_root::marker_bytes("test", &[]),
    )
    .unwrap();
    p
}

/// `(exit code, envelope)` of `moai spec ARGS --project P` (`--project` appended).
fn spec(home: &TempHome, project: &Path, args: &[&str]) -> (i32, Value) {
    let mut all = vec!["spec"];
    all.extend_from_slice(args);
    let p = project.to_str().unwrap();
    all.extend_from_slice(&["--project", p]);
    let out = home.command(&all).output().unwrap();
    let v: Value = serde_json::from_slice(&out.stdout).unwrap_or_else(|_| {
        panic!(
            "{} / {}",
            String::from_utf8_lossy(&out.stdout),
            String::from_utf8_lossy(&out.stderr)
        )
    });
    (out.status.code().unwrap(), v)
}

fn spec_path(project: &Path, id: &str) -> PathBuf {
    project.join(".moai/specs").join(id).join("spec.md")
}

fn new_spec(home: &TempHome, p: &Path, extra: &[&str]) -> (i32, Value) {
    let mut args = vec![
        "new",
        "login-fix",
        "--title",
        "Fix login",
        "--date",
        "2026-11-17",
    ];
    args.extend_from_slice(extra);
    spec(home, p, &args)
}

#[test]
fn new_check_list_show_and_nothing_else_is_written() {
    let home = TempHome::new();
    let p = project(&home);
    let before = tree_state(home.path());
    let (code, v) = new_spec(&home, &p, &[]);
    assert_eq!(code, 0, "{v}");
    let r = &v["result"];
    assert_eq!(v["operation"], "spec");
    assert_eq!(r["action"], "new");
    assert_eq!(r["id"], ID);
    assert_eq!(r["path"], format!(".moai/specs/{ID}/spec.md"));
    assert_eq!(r["status"], "draft");
    assert_eq!(r["form"], "full");
    let after = tree_state(home.path());
    // The new folder and its one file, nothing else: `.moai/specs` itself is new here too.
    let added: Vec<String> = after
        .iter()
        .filter(|e| !before.contains(e))
        .map(|(p, _)| p.clone())
        .collect();
    assert_eq!(
        added,
        vec![
            "proj/.moai/specs".to_string(),
            format!("proj/.moai/specs/{ID}"),
            format!("proj/.moai/specs/{ID}/spec.md"),
        ],
    );
    // Nothing that was there changed or went away.
    assert!(before.iter().all(|e| after.contains(e)));
    let text = fs::read_to_string(spec_path(&p, ID)).unwrap();
    assert!(text.starts_with(&format!("---\nid: {ID}\ntitle: \"Fix login\"\nstatus: draft\nrevision: 1\ncreated: 2026-11-17\n---\n")), "{text}");

    let quiet = tree_state(home.path());
    let (code, v) = spec(&home, &p, &["check", ID]);
    assert_eq!(code, 0, "{v}");
    assert_eq!(v["result"]["errors"], 0);
    assert_eq!(v["result"]["files"][0]["id"], ID);
    let (code, v) = spec(&home, &p, &["check", "--all"]);
    assert_eq!(code, 0, "{v}");
    assert_eq!(v["result"]["count"], 1);
    let (code, v) = spec(&home, &p, &["list"]);
    assert_eq!(code, 0, "{v}");
    assert_eq!(
        v["result"]["rows"],
        json!([{"id": ID, "folder": ID, "status": "draft", "title": "Fix login", "parent": null,
                "legacy": false, "errors": 0, "warnings": 0}])
    );
    let (code, v) = spec(&home, &p, &["show", ID]);
    assert_eq!(code, 0, "{v}");
    let s = &v["result"]["spec"];
    assert_eq!(s["id"], ID);
    assert_eq!(s["form"], "full");
    assert_eq!(s["requirements"][0]["id"], "REQ-1");
    assert_eq!(s["acceptance"][0]["refs"], json!(["REQ-1"]));
    assert!(s["digest"].as_str().unwrap().len() > 10, "{s}");
    assert_eq!(
        tree_state(home.path()),
        quiet,
        "check, list and show write nothing"
    );
}

#[test]
fn minimal_with_parent_and_the_default_date() {
    let home = TempHome::new();
    let p = project(&home);
    let (code, v) = spec(
        &home,
        &p,
        &[
            "new",
            "small",
            "--title",
            "Small",
            "--minimal",
            "--parent",
            "SPEC-AUTH-001",
        ],
    );
    assert_eq!(code, 0, "{v}");
    let id = v["result"]["id"].as_str().unwrap().to_string();
    // The date is today in UTC: an 8-digit real date in the id, the same in `created`.
    let created = v["result"]["created"].as_str().unwrap();
    assert!(moai_cli::spec_doc::valid_date(created), "{created}");
    assert_eq!(id, format!("SPEC-{}-small", created.replace('-', "")));
    let (code, v) = spec(&home, &p, &["show", &id]);
    assert_eq!(code, 0, "{v}");
    assert_eq!(v["result"]["spec"]["form"], "minimal");
    assert_eq!(v["result"]["spec"]["parent"], "SPEC-AUTH-001");
    assert_eq!(v["result"]["spec"]["findings"], json!([]));
}

#[test]
fn new_refuses_and_writes_nothing() {
    let home = TempHome::new();
    let p = project(&home);
    assert_eq!(new_spec(&home, &p, &[]).0, 0);
    let before = tree_state(home.path());
    let original = fs::read(spec_path(&p, ID)).unwrap();
    for (extra, kind) in [
        (&[][..], "conflict"),
        (&["--parent", ID][..], "invalid_input"),
        (&["--parent", "nope"][..], "invalid_input"),
    ] {
        let (code, v) = new_spec(&home, &p, extra);
        assert_eq!(code, 1, "{extra:?}: {v}");
        assert_eq!(v["ok"], false);
        assert_eq!(v["error"]["kind"], kind, "{extra:?}: {v}");
    }
    for date in ["2026-02-29", "2026-11-31", "20261118", "2026-11-1"] {
        let (code, v) = spec(&home, &p, &["new", "x", "--title", "t", "--date", date]);
        assert_eq!(code, 1, "{date}: {v}");
        assert_eq!(v["error"]["kind"], "invalid_input", "{date}: {v}");
    }
    for (slug, title) in [
        ("Bad", "t"),
        ("a--b", "t"),
        (&*"a".repeat(49), "t"),
        ("ok", "say \"hi\""),
        ("ok", &*"x".repeat(201)),
    ] {
        let (code, v) = spec(
            &home,
            &p,
            &["new", slug, "--title", title, "--date", "2026-11-18"],
        );
        assert_eq!(code, 1, "{slug}: {v}");
        assert_eq!(v["error"]["kind"], "invalid_input", "{v}");
    }
    assert_eq!(tree_state(home.path()), before);
    assert_eq!(fs::read(spec_path(&p, ID)).unwrap(), original);
    // The 48-character slug and the 200-character title are the last that pass.
    let (code, v) = spec(
        &home,
        &p,
        &[
            "new",
            &"a".repeat(48),
            "--title",
            &"x".repeat(200),
            "--date",
            "2026-11-18",
        ],
    );
    assert_eq!(code, 0, "{v}");
}

#[cfg(unix)]
#[test]
fn a_linked_moai_or_specs_folder_is_refused() {
    let home = TempHome::new();
    let p = home.path().join("proj");
    let elsewhere = home.path().join("elsewhere");
    fs::create_dir_all(&p).unwrap();
    fs::create_dir_all(&elsewhere).unwrap();
    std::os::unix::fs::symlink(&elsewhere, p.join(".moai")).unwrap();
    let before = tree_state(home.path());
    for args in [&["list"][..], &["check", "--all"][..]] {
        let (code, v) = spec(&home, &p, args);
        assert_eq!(code, 1, "{v}");
        assert_eq!(v["error"]["kind"], "invalid_input", "{v}");
    }
    assert_eq!(new_spec(&home, &p, &[]).0, 1);
    assert_eq!(tree_state(home.path()), before);
    assert_eq!(fs::read_dir(&elsewhere).unwrap().count(), 0);

    let q = project(&home);
    std::os::unix::fs::symlink(&elsewhere, q.join(".moai/specs")).unwrap();
    let before = tree_state(home.path());
    assert_eq!(new_spec(&home, &q, &[]).0, 1);
    assert_eq!(spec(&home, &q, &["list"]).0, 1);
    assert_eq!(tree_state(home.path()), before);
}

#[test]
fn a_project_without_moai_gets_only_the_spec_folders() {
    let home = TempHome::new();
    let p = home.path().join("bare");
    fs::create_dir_all(&p).unwrap();
    let (code, v) = spec(&home, &p, &["list"]);
    assert_eq!(code, 0, "{v}");
    assert_eq!(
        v["result"],
        json!({"action": "list", "specs": "absent", "count": 0, "rows": []})
    );
    let (code, v) = spec(&home, &p, &["check", "--all"]);
    assert_eq!(code, 0, "{v}");
    assert_eq!(v["result"]["specs"], "absent");
    assert!(!p.join(".moai").exists(), "reads create nothing");
    let (code, v) = new_spec(&home, &p, &[]);
    assert_eq!(code, 0, "{v}");
    let mut names: Vec<String> = tree_state(&p).into_iter().map(|(n, _)| n).collect();
    names.sort();
    assert_eq!(
        names,
        vec![
            ".moai".to_string(),
            ".moai/specs".into(),
            format!(".moai/specs/{ID}"),
            format!(".moai/specs/{ID}/spec.md"),
        ]
    );
}

#[test]
fn check_reports_codes_and_exits_one_on_an_error() {
    let home = TempHome::new();
    let p = project(&home);
    assert_eq!(new_spec(&home, &p, &["--minimal"]).0, 0);
    let path = spec_path(&p, ID);
    let good = fs::read_to_string(&path).unwrap();
    fs::write(
        &path,
        good.replace(" · verify: `<command that checks it>`", "")
            .replace("revision: 1\n", "revision: 1\nowner: me\n"),
    )
    .unwrap();
    let before = tree_state(home.path());
    let (code, v) = spec(&home, &p, &["check", ID]);
    assert_eq!(code, 1, "{v}");
    assert_eq!(
        tree_state(home.path()),
        before,
        "a failed check writes nothing"
    );
    assert_eq!(v["error"]["kind"], "invalid_input");
    let m = v["error"]["message"].as_str().unwrap();
    assert!(
        m.starts_with("spec check refused: 1 error(s) and 1 warning(s) in 1 SPEC(s): "),
        "{m}"
    );
    assert!(
        m.contains(&format!(".moai/specs/{ID}/spec.md:22 ac_no_verify")),
        "{m}"
    );
    // show still answers, with the findings.
    let (code, v) = spec(&home, &p, &["show", ID]);
    assert_eq!(code, 0, "{v}");
    let codes: Vec<&str> = v["result"]["spec"]["findings"]
        .as_array()
        .unwrap()
        .iter()
        .map(|f| f["code"].as_str().unwrap())
        .collect();
    assert_eq!(codes, vec!["fm_unknown_key", "ac_no_verify"]);
    // A warning alone passes.
    fs::write(
        &path,
        good.replace("revision: 1\n", "revision: 1\nowner: me\n"),
    )
    .unwrap();
    let (code, v) = spec(&home, &p, &["check", "--all"]);
    assert_eq!(code, 0, "{v}");
    assert_eq!(v["result"]["warnings"], 1);
    assert_eq!(
        v["result"]["files"][0]["findings"][0]["code"],
        "fm_unknown_key"
    );
}

#[test]
fn check_and_show_name_ids_and_folders() {
    let home = TempHome::new();
    let p = project(&home);
    for args in [
        &["check", "SPEC-20261117-none"][..],
        &["show", "SPEC-AUTH-009"][..],
    ] {
        let (code, v) = spec(&home, &p, args);
        assert_eq!(code, 1, "{v}");
        assert_eq!(v["error"]["kind"], "not_found", "{v}");
    }
    for args in [&["check", "../x"][..], &["show", "spec-x"][..]] {
        let (code, v) = spec(&home, &p, args);
        assert_eq!(code, 1, "{v}");
        assert_eq!(v["error"]["kind"], "invalid_input", "{v}");
        assert!(
            !v["error"]["message"].as_str().unwrap().contains("../x"),
            "{v}"
        );
    }
    // A folder whose name is not its id, and a Go-generation SPEC listed as legacy.
    assert_eq!(new_spec(&home, &p, &[]).0, 0);
    let specs = p.join(".moai/specs");
    fs::rename(specs.join(ID), specs.join("SPEC-20261117-moved")).unwrap();
    fs::create_dir_all(specs.join("SPEC-AUTH-001")).unwrap();
    fs::write(
        specs.join("SPEC-AUTH-001/spec.md"),
        "---\nid: SPEC-AUTH-001\ntitle: Auth\nstatus: implemented\n---\n\n# Auth\n",
    )
    .unwrap();
    fs::write(specs.join("README.md"), "not a SPEC").unwrap();
    let (code, v) = spec(&home, &p, &["list"]);
    assert_eq!(code, 0, "{v}");
    let rows = v["result"]["rows"].as_array().unwrap();
    assert_eq!(rows.len(), 2, "{v}");
    assert_eq!(rows[0]["id"], "SPEC-20261117-login-fix");
    assert_eq!(rows[0]["folder"], "SPEC-20261117-moved");
    assert_eq!(rows[0]["errors"], 1);
    assert_eq!(rows[1]["id"], "SPEC-AUTH-001");
    assert_eq!(rows[1]["legacy"], true);
    assert_eq!(rows[1]["status"], "implemented");
    assert_eq!(rows[1]["errors"], 0);
    let (code, v) = spec(&home, &p, &["check", "--all"]);
    assert_eq!(code, 1, "{v}");
    let m = v["error"]["message"].as_str().unwrap();
    assert!(
        m.contains(".moai/specs/SPEC-20261117-moved/spec.md:2 id_folder"),
        "{m}"
    );
    let (code, v) = spec(&home, &p, &["check", "SPEC-AUTH-001"]);
    assert_eq!(code, 0, "{v}");
}

#[test]
fn closure_needs_completed_and_a_result() {
    let home = TempHome::new();
    let p = project(&home);
    assert_eq!(new_spec(&home, &p, &["--minimal"]).0, 0);
    let (code, v) = spec(&home, &p, &["check", ID, "--closure"]);
    assert_eq!(code, 1, "{v}");
    let m = v["error"]["message"].as_str().unwrap();
    assert!(
        m.contains("closure_status") && m.contains("closure_result"),
        "{m}"
    );
    // Approve with the digest `show` reports, complete, and record the result.
    let (_, v) = spec(&home, &p, &["show", ID]);
    let digest = v["result"]["spec"]["digest"].as_str().unwrap().to_string();
    let path = spec_path(&p, ID);
    let text = fs::read_to_string(&path).unwrap().replace(
        "status: draft\n",
        &format!(
            "status: completed\napproval:\n  by: user\n  at: 2026-11-17\n  basis: in-session \
             selection\n  digest: {digest}\n"
        ),
    );
    fs::write(
        &path,
        format!("{text}\n## Result\n\n- AC-1: met — test run\n"),
    )
    .unwrap();
    let (code, v) = spec(&home, &p, &["check", ID, "--closure"]);
    assert_eq!(code, 0, "{v}");
    assert_eq!(v["result"]["closure"], true);
    fs::write(&path, format!("{text}\n## Result\n\n- AC-1: pending\n")).unwrap();
    let (code, v) = spec(&home, &p, &["check", ID, "--closure"]);
    assert_eq!(code, 1, "{v}");
    assert!(
        v["error"]["message"]
            .as_str()
            .unwrap()
            .contains("closure_open"),
        "{v}"
    );
}
