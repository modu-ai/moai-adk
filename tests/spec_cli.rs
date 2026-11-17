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

/// A template's text with every `<…>` placeholder filled with the word `filled`.
fn filled(text: &str) -> String {
    let mut out = String::new();
    let mut rest = text;
    while let Some(open) = rest.find('<') {
        let Some(close) = rest[open..].find('>') else {
            break;
        };
        out.push_str(&rest[..open]);
        out.push_str("filled");
        rest = &rest[open + close + 1..];
    }
    out.push_str(rest);
    out
}

/// The codes of `show`'s findings.
fn finding_codes(v: &Value) -> Vec<String> {
    v["result"]["spec"]["findings"]
        .as_array()
        .unwrap()
        .iter()
        .map(|f| f["code"].as_str().unwrap().to_string())
        .collect()
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
    let row = &v["result"]["rows"][0];
    assert_eq!(v["result"]["count"], 1);
    for (key, want) in [
        ("id", json!(ID)),
        ("folder", json!(ID)),
        ("status", json!("draft")),
        ("title", json!("Fix login")),
        ("parent", Value::Null),
        ("legacy", json!(false)),
        ("errors", json!(0)),
    ] {
        assert_eq!(row[key], want, "{key}: {row}");
    }
    let (code, v) = spec(&home, &p, &["show", ID]);
    assert_eq!(code, 0, "{v}");
    // A fresh draft only warns about the template text still to fill.
    let codes = finding_codes(&v);
    assert!(
        !codes.is_empty() && codes.iter().all(|c| c == "placeholder"),
        "{codes:?}"
    );
    assert_eq!(row["warnings"], json!(codes.len()));
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
    let codes = finding_codes(&v);
    assert!(
        !codes.is_empty() && codes.iter().all(|c| c == "placeholder"),
        "{codes:?}"
    );
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
    let template = fs::read_to_string(&path).unwrap();
    let good = filled(&template);
    fs::write(
        &path,
        filled(&template.replace(" · verify: `<command that checks it>`", ""))
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
        m.contains(&format!(".moai/specs/{ID}/spec.md:23 ac_no_verify")),
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
    let path = spec_path(&p, ID);
    fs::write(&path, filled(&fs::read_to_string(&path).unwrap())).unwrap();
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

/// The refusal names at most 50 errors (N) and says how many more there are from 51 (N+1).
#[test]
fn the_refusal_lists_fifty_errors_and_counts_the_rest() {
    for (count, more) in [
        (50, None),
        (51, Some("; and 1 more (spec show <id> lists them)")),
    ] {
        let home = TempHome::new();
        let p = project(&home);
        for n in 0..count {
            let id = format!("SPEC-20261117-s{n:02}");
            let dir = p.join(".moai/specs").join(&id);
            fs::create_dir_all(&dir).unwrap();
            // One error each: no `revision`.
            let text = format!(
                "---\nid: {id}\ntitle: t\nstatus: draft\ncreated: 2026-11-17\n---\n\n## Intent\n\nx\n\n\
                 ## Requirements\n\n- REQ-1: The tool shall work.\n\n## Acceptance\n\n\
                 - AC-1 (REQ-1): works · verify: true\n\n## Scope\n\n- write: `a.rs`\n\n## Plan\n\n1. x\n"
            );
            fs::write(dir.join("spec.md"), text).unwrap();
        }
        let (code, v) = spec(&home, &p, &["check", "--all"]);
        assert_eq!(code, 1, "{v}");
        let m = v["error"]["message"].as_str().unwrap();
        assert!(
            m.starts_with(&format!(
                "spec check refused: {count} error(s) and 0 warning(s)"
            )),
            "{m}"
        );
        assert_eq!(m.matches(" revision_invalid").count(), 50, "{count}");
        match more {
            Some(tail) => assert!(m.ends_with(tail), "{m}"),
            None => assert!(!m.contains("more"), "{m}"),
        }
    }
}

/// The repository's own SPECs pass the checker (SDD B2): a SPEC merged with an error turns
/// this red.
#[test]
fn the_repository_specs_pass_check() {
    let home = TempHome::new();
    let root = Path::new(env!("CARGO_MANIFEST_DIR"));
    let (code, v) = spec(&home, root, &["check", "--all"]);
    assert_eq!(code, 0, "{v}");
    assert_eq!(v["result"]["specs"], "present");
    assert!(v["result"]["count"].as_u64().unwrap() >= 2, "{v}");
}

/// `(exit code, stdout)` of a read-only command run with the temporary HOME.
fn raw(home: &TempHome, args: &[&str]) -> (Option<i32>, String) {
    let out = home.command(args).output().unwrap();
    (
        out.status.code(),
        String::from_utf8_lossy(&out.stdout).into_owned(),
    )
}

/// SDD B1 (leader decision 2026-11-17): a `.moai/` that holds only `specs/` is what a fresh
/// clone of a repository that tracks its SPECs has, and it reads like no state folder at all.
/// `spec new` in a new project and in a `.moai-cli` project changes neither the state
/// verdict nor what `project-detect` and `project-update-plan` answer, and writes no marker.
#[test]
fn spec_new_keeps_the_state_verdict_and_the_project_answers() {
    for legacy in [false, true] {
        let home = TempHome::new();
        let p = home.path().join("proj");
        fs::create_dir_all(&p).unwrap();
        if legacy {
            fs::create_dir_all(p.join(".moai-cli")).unwrap();
            fs::write(p.join(".moai-cli/project.yaml"), "schema_version: 2\n").unwrap();
        }
        let ps = p.to_str().unwrap();
        let look = || {
            let c = moai_cli::state_root::classify(&p);
            (
                c.generation,
                moai_cli::state_root::dir(&p),
                moai_cli::state_root::is_new_project(&p),
                raw(&home, &["project-detect", ps]),
                raw(&home, &["project-update-plan", ps]),
            )
        };
        let before = look();
        assert_eq!(before.0, "none", "{legacy}");
        assert_eq!(new_spec(&home, &p, &[]).0, 0);
        assert!(spec_path(&p, ID).is_file());
        assert_eq!(look(), before, "legacy={legacy}");
        assert!(
            !p.join(".moai/moai.json").exists(),
            "spec new writes no marker"
        );
        assert!(!p.join(".moai/moai-cli.json").exists());
        // Anything beside specs/ is state again: the old verdict for an unmarked folder.
        fs::write(p.join(".moai/notes.txt"), "x").unwrap();
        assert_eq!(moai_cli::state_root::classify(&p).generation, "unknown");
    }
}

/// A `.moai-cli` project whose clone already holds `.moai/specs/` moves its earlier state into
/// that `.moai/` and back; the SPECs stay where git tracks them in both directions.
#[test]
fn an_earlier_root_moves_into_a_moai_that_holds_specs_and_back() {
    let home = TempHome::new();
    let p = home.path().join("proj");
    fs::create_dir_all(p.join(".moai-cli/index")).unwrap();
    fs::write(p.join(".moai-cli/project.yaml"), "schema_version: 2\n").unwrap();
    fs::write(p.join(".moai-cli/index/a.json"), "{\"a\":1}").unwrap();
    assert_eq!(new_spec(&home, &p, &[]).0, 0);
    let spec_bytes = fs::read(spec_path(&p, ID)).unwrap();
    let ps = p.to_str().unwrap();
    let call = |args: &[&str]| -> Value {
        let (_, out) = raw(&home, args);
        serde_json::from_str(&out).unwrap_or_else(|_| panic!("{args:?}: {out}"))
    };
    let plan = call(&["go-state-plan", ps]);
    assert_eq!(plan["ok"], true, "{plan}");
    let digest = plan["result"]["plan_digest"].as_str().unwrap().to_string();
    let v = call(&[
        "go-state-apply",
        ps,
        "--plan-digest",
        &digest,
        "--permission-reference",
        "test-permission-ref",
    ]);
    assert_eq!(v["ok"], true, "{v}");
    assert_eq!(v["result"]["legacy_move"], true, "{v}");
    assert_eq!(
        fs::read_to_string(p.join(".moai/project.yaml")).unwrap(),
        "schema_version: 2\n"
    );
    assert_eq!(fs::read(spec_path(&p, ID)).unwrap(), spec_bytes);
    assert_eq!(moai_cli::state_root::classify(&p).generation, "rust");
    let backup = v["result"]["backup"].as_str().unwrap().to_string();
    let rp = call(&["go-state-revert-plan", ps, "--backup", &backup]);
    assert_eq!(rp["ok"], true, "{rp}");
    let rd = rp["result"]["plan_digest"].as_str().unwrap().to_string();
    let v = call(&[
        "go-state-revert",
        ps,
        "--backup",
        &backup,
        "--plan-digest",
        &rd,
        "--permission-reference",
        "test-permission-ref",
    ]);
    assert_eq!(v["ok"], true, "{v}");
    assert_eq!(
        fs::read_to_string(p.join(".moai-cli/project.yaml")).unwrap(),
        "schema_version: 2\n"
    );
    assert_eq!(fs::read(spec_path(&p, ID)).unwrap(), spec_bytes);
    let mut top: Vec<String> = fs::read_dir(p.join(".moai"))
        .unwrap()
        .map(|e| e.unwrap().file_name().to_string_lossy().into_owned())
        .collect();
    top.sort();
    assert_eq!(
        top,
        vec!["specs".to_string()],
        "only the SPECs stay in .moai/"
    );
    assert_eq!(moai_cli::state_root::classify(&p).generation, "none");
}

/// A `.moai-cli` project with earlier SPECs and the tracked `.moai/specs/` (RB1).
fn earlier_with_specs(home: &TempHome, earlier_ids: &[&str]) -> PathBuf {
    let p = home.path().join("proj");
    fs::create_dir_all(p.join(".moai-cli/index")).unwrap();
    fs::write(p.join(".moai-cli/project.yaml"), "schema_version: 2\n").unwrap();
    fs::write(p.join(".moai-cli/index/a.json"), "{\"a\":1}").unwrap();
    for id in earlier_ids {
        let dir = p.join(".moai-cli/specs").join(id);
        fs::create_dir_all(&dir).unwrap();
        fs::write(dir.join("spec.md"), format!("earlier {id}\n")).unwrap();
    }
    assert_eq!(new_spec(home, &p, &[]).0, 0);
    p
}

/// `go-state-plan` → `go-state-apply` → `go-state-revert-plan` → `go-state-revert`, each
/// envelope.
fn json_call(home: &TempHome, args: &[&str]) -> Value {
    let (_, out) = raw(home, args);
    serde_json::from_str(&out).unwrap_or_else(|_| panic!("{args:?}: {out}"))
}

/// RB1: SPEC folders with different ids merge one by one into `.moai/specs/`; the revert puts
/// each back on its side.
#[test]
fn earlier_specs_with_other_ids_merge_and_the_revert_splits_them_again() {
    let home = TempHome::new();
    let p = earlier_with_specs(&home, &["SPEC-OLD-001", "SPEC-20261101-old"]);
    let before = tree_state(&p);
    let ps = p.to_str().unwrap();
    let plan = json_call(&home, &["go-state-plan", ps]);
    assert_eq!(plan["ok"], true, "{plan}");
    let moves: Vec<(String, String, String)> = plan["result"]["state_moves"]
        .as_array()
        .unwrap()
        .iter()
        .map(|m| {
            (
                m["from"].as_str().unwrap().to_string(),
                m["to"].as_str().unwrap().to_string(),
                m["action"].as_str().unwrap().to_string(),
            )
        })
        .collect();
    for id in ["SPEC-20261101-old", "SPEC-OLD-001"] {
        let want = (
            format!(".moai-cli/specs/{id}"),
            format!(".moai/specs/{id}"),
            "move".to_string(),
        );
        assert!(moves.contains(&want), "{id}: {moves:?}");
    }
    assert!(
        !moves.iter().any(|(from, _, _)| from == ".moai-cli/specs"),
        "the folder never moves whole: {moves:?}"
    );
    let digest = plan["result"]["plan_digest"].as_str().unwrap().to_string();
    let v = json_call(
        &home,
        &[
            "go-state-apply",
            ps,
            "--plan-digest",
            &digest,
            "--permission-reference",
            "test-permission-ref",
        ],
    );
    assert_eq!(v["ok"], true, "{v}");
    assert_eq!(v["result"]["state"], "committed", "{v}");
    let mut merged: Vec<String> = fs::read_dir(p.join(".moai/specs"))
        .unwrap()
        .map(|e| e.unwrap().file_name().to_string_lossy().into_owned())
        .collect();
    merged.sort();
    assert_eq!(
        merged,
        vec![
            "SPEC-20261101-old",
            "SPEC-20261117-login-fix",
            "SPEC-OLD-001"
        ]
    );
    assert_eq!(
        fs::read_to_string(p.join(".moai/specs/SPEC-OLD-001/spec.md")).unwrap(),
        "earlier SPEC-OLD-001\n"
    );
    let backup = v["result"]["backup"].as_str().unwrap().to_string();
    let rp = json_call(&home, &["go-state-revert-plan", ps, "--backup", &backup]);
    assert_eq!(rp["ok"], true, "{rp}");
    let items = rp["result"]["items"].as_array().unwrap();
    assert!(
        items
            .iter()
            .any(|i| i["path"] == ".moai/specs" && i["action"] == "restored_to_moai_specs"),
        "N-e: {rp}"
    );
    let rd = rp["result"]["plan_digest"].as_str().unwrap().to_string();
    let v = json_call(
        &home,
        &[
            "go-state-revert",
            ps,
            "--backup",
            &backup,
            "--plan-digest",
            &rd,
            "--permission-reference",
            "test-permission-ref",
        ],
    );
    assert_eq!(v["ok"], true, "{v}");
    // Both sides as they were; the set-aside `.moai.rust-<UTC>` keeps the receipt.
    let after: Vec<(String, Vec<u8>)> = tree_state(&p)
        .into_iter()
        .filter(|(n, _)| !n.starts_with(".moai.rust-") && !n.starts_with(".moai-cli/backups"))
        .collect();
    let before: Vec<(String, Vec<u8>)> = before
        .into_iter()
        .filter(|(n, _)| !n.starts_with(".moai-cli/backups"))
        .collect();
    assert_eq!(after, before);
}

/// RB1: the same SPEC id on both sides is a conflict; the whole move waits and nothing moves.
#[test]
fn an_earlier_spec_with_the_same_id_blocks_the_move() {
    let home = TempHome::new();
    let p = earlier_with_specs(&home, &[ID, "SPEC-OLD-001"]);
    let before = tree_state(&p);
    let ps = p.to_str().unwrap();
    let plan = json_call(&home, &["go-state-plan", ps]);
    assert_eq!(plan["ok"], true, "{plan}");
    let item = plan["result"]["items"]
        .as_array()
        .unwrap()
        .iter()
        .find(|i| i["path"] == ".moai-cli")
        .unwrap_or_else(|| panic!("{plan}"))
        .clone();
    assert_eq!(item["action"], "leave", "{item}");
    let reason = item["reason"].as_str().unwrap();
    assert!(
        reason.starts_with(&format!(
            "conflict: {ID} is in both .moai-cli/specs/ and .moai/specs/"
        )),
        "{reason}"
    );
    let digest = plan["result"]["plan_digest"].as_str().unwrap().to_string();
    let v = json_call(
        &home,
        &[
            "go-state-apply",
            ps,
            "--plan-digest",
            &digest,
            "--permission-reference",
            "test-permission-ref",
        ],
    );
    assert_eq!(v["ok"], true, "{v}");
    assert_ne!(v["result"]["state"], "committed", "{v}");
    assert_eq!(tree_state(&p), before, "nothing moved");
    assert_eq!(moai_cli::state_root::classify(&p).generation, "none");
}

/// N-a, N-c: only a real `specs/` folder (and a Finder `.DS_Store` beside it) reads as no
/// state; a linked or plain-file `specs` does not.
#[test]
fn only_a_real_specs_folder_reads_as_no_state() {
    let home = TempHome::new();
    let verdict = |setup: &dyn Fn(&Path)| {
        let p = home
            .path()
            .join(format!("p{}", fs::read_dir(home.path()).unwrap().count()));
        fs::create_dir_all(p.join(".moai")).unwrap();
        setup(&p);
        moai_cli::state_root::classify(&p).generation
    };
    assert_eq!(
        verdict(&|p| fs::create_dir(p.join(".moai/specs")).unwrap()),
        "none"
    );
    assert_eq!(
        verdict(&|p| {
            fs::create_dir(p.join(".moai/specs")).unwrap();
            fs::write(p.join(".moai/.DS_Store"), "x").unwrap();
        }),
        "none"
    );
    assert_eq!(
        verdict(&|p| fs::write(p.join(".moai/specs"), "a file").unwrap()),
        "unknown"
    );
    // A `.DS_Store` that is a folder is something else.
    assert_eq!(
        verdict(&|p| {
            fs::create_dir(p.join(".moai/specs")).unwrap();
            fs::create_dir(p.join(".moai/.DS_Store")).unwrap();
        }),
        "unknown"
    );
    #[cfg(unix)]
    {
        let elsewhere = home.path().join("elsewhere-specs");
        fs::create_dir_all(&elsewhere).unwrap();
        assert_eq!(
            verdict(&|p| std::os::unix::fs::symlink(&elsewhere, p.join(".moai/specs")).unwrap()),
            "unknown"
        );
    }
}

/// Third audit N1: a `.moai-cli` that holds only `specs/<id>`, with the same id in
/// `.moai/specs/`, has nothing else to move; the plan still shows the conflict and the apply
/// moves nothing.
#[test]
fn a_same_id_conflict_shows_even_when_nothing_else_would_move() {
    let home = TempHome::new();
    let p = home.path().join("proj");
    let earlier = p.join(".moai-cli/specs").join(ID);
    fs::create_dir_all(&earlier).unwrap();
    fs::write(earlier.join("spec.md"), "earlier copy\n").unwrap();
    assert_eq!(new_spec(&home, &p, &[]).0, 0);
    let before = tree_state(&p);
    let ps = p.to_str().unwrap();
    let plan = json_call(&home, &["go-state-plan", ps]);
    assert_eq!(plan["ok"], true, "{plan}");
    let item = plan["result"]["items"]
        .as_array()
        .unwrap()
        .iter()
        .find(|i| i["path"] == ".moai-cli")
        .unwrap_or_else(|| panic!("the conflict is not shown: {plan}"))
        .clone();
    assert_eq!(item["action"], "leave", "{item}");
    assert!(
        item["reason"]
            .as_str()
            .unwrap()
            .starts_with(&format!("conflict: {ID} is in both")),
        "{item}"
    );
    let digest = plan["result"]["plan_digest"].as_str().unwrap().to_string();
    let v = json_call(
        &home,
        &[
            "go-state-apply",
            ps,
            "--plan-digest",
            &digest,
            "--permission-reference",
            "test-permission-ref",
        ],
    );
    assert_eq!(v["ok"], true, "{v}");
    assert_ne!(v["result"]["state"], "committed", "{v}");
    assert_eq!(tree_state(&p), before, "nothing moved");
}

/// Third audit N2: the receipt is no authority over SPEC entries. A receipt edited after the
/// merge to name `specs/..`, a nested or absolute path, or a path under `.moai/` is refused by
/// the revert plan and by the revert, and nothing moves.
#[test]
fn a_receipt_that_forges_a_spec_entry_is_refused() {
    let home = TempHome::new();
    let p = earlier_with_specs(&home, &["SPEC-OLD-001"]);
    let ps = p.to_str().unwrap();
    let plan = json_call(&home, &["go-state-plan", ps]);
    let digest = plan["result"]["plan_digest"].as_str().unwrap().to_string();
    let v = json_call(
        &home,
        &[
            "go-state-apply",
            ps,
            "--plan-digest",
            &digest,
            "--permission-reference",
            "test-permission-ref",
        ],
    );
    assert_eq!(v["result"]["state"], "committed", "{v}");
    let backup = v["result"]["backup"].as_str().unwrap().to_string();
    let receipt_path = p
        .join(".moai/backups")
        .join(&backup)
        .join("go-state-receipt.json");
    let original = fs::read(&receipt_path).unwrap();
    let receipt: Value = serde_json::from_slice(&original).unwrap();
    let at = receipt["moves"]
        .as_array()
        .unwrap()
        .iter()
        .position(|m| m["path"] == ".moai-cli/specs/SPEC-OLD-001")
        .unwrap_or_else(|| panic!("{receipt}"));
    assert_eq!(receipt["moves"][at]["kind"], "state", "{receipt}");
    // The revert plan of the true receipt is accepted.
    assert_eq!(
        json_call(&home, &["go-state-revert-plan", ps, "--backup", &backup])["ok"],
        true
    );
    for forged in [
        ".moai-cli/specs/..",
        ".moai-cli/specs/.",
        ".moai-cli/specs/SPEC-OLD-001/spec.md",
        ".moai-cli/specs//etc",
        ".moai-cli/specs/a\\b",
        "/etc/specs/x",
        ".moai/specs/SPEC-OLD-001",
    ] {
        let mut r = receipt.clone();
        r["moves"][at]["path"] = json!(forged);
        fs::write(&receipt_path, serde_json::to_vec(&r).unwrap()).unwrap();
        let before = tree_state(&p);
        let rp = json_call(&home, &["go-state-revert-plan", ps, "--backup", &backup]);
        assert_eq!(rp["ok"], false, "{forged}: {rp}");
        let v = json_call(
            &home,
            &[
                "go-state-revert",
                ps,
                "--backup",
                &backup,
                "--plan-digest",
                "sha256:0",
                "--permission-reference",
                "test-permission-ref",
            ],
        );
        assert_eq!(v["ok"], false, "{forged}: {v}");
        assert_eq!(tree_state(&p), before, "{forged}: nothing moved");
    }
    fs::write(&receipt_path, &original).unwrap();
}
