//! #2434 decision 40: `moai-cli` syncs a project's labels from `.github/moai/labels.yml`.
//! Library tests replay synthetic exchanges through `RecordedTransport` with a fake clock: no
//! network, no sleeping, no `gh`. Nothing here talks to a real remote (host: not_run).
use moai_cli::forge::{
    labels::{self, LabelsRequest, RemoteLabel},
    policy::Clock,
    transport::{Exchange, HttpRequest, HttpResponse, RecordedTransport},
    Client, Config,
};
mod common;

use serde_json::{json, Value};
use std::{
    cell::{Cell, RefCell},
    collections::BTreeMap,
    fs,
    path::{Path, PathBuf},
    process::{Command, Stdio},
};

struct FakeClock {
    now: Cell<i64>,
    slept: RefCell<Vec<u64>>,
}
impl Clock for FakeClock {
    fn now(&self) -> i64 {
        self.now.get()
    }
    fn sleep(&self, secs: u64) {
        self.slept.borrow_mut().push(secs);
        self.now.set(self.now.get() + secs as i64);
    }
}
fn clock() -> FakeClock {
    FakeClock {
        now: Cell::new(1_800_000_000),
        slept: RefCell::new(Vec::new()),
    }
}

const BASE: &str = "https://api.github.com/repos/acme/widgets/labels";

fn request() -> LabelsRequest {
    serde_json::from_value(json!({"schema_version": 1, "owner": "acme", "repo": "widgets"}))
        .unwrap()
}

fn shipped() -> String {
    fs::read_to_string(Path::new(env!("CARGO_MANIFEST_DIR")).join("templates/github/labels.yml"))
        .unwrap()
}

const SMALL: &str = "# project labels\nschema_version: 2\nlabels:\n- name: type:feature\n  color: a2eeef\n  description: New feature or behaviour\n- name: priority:P0\n  color: b60205\n  description: Drop everything\n- name: area:docs\n  color: 0075ca\n  description: Documentation\n";

fn remote(name: &str, color: &str, description: Option<&str>) -> Value {
    json!({"id": 1, "name": name, "color": color, "description": description, "default": false})
}

fn list_page(url: &str, labels: Vec<Value>, next: Option<&str>) -> Exchange {
    let mut response = HttpResponse::new(200).with_body(Value::Array(labels));
    if let Some(next) = next {
        response = response.with_header("link", format!("<{next}>; rel=\"next\""));
    }
    Exchange {
        request: HttpRequest::new("GET", url),
        response,
    }
}

fn first_page() -> String {
    format!("{BASE}?per_page=100")
}

fn write(method: &str, url: &str, status: u16) -> Exchange {
    Exchange {
        request: HttpRequest::new(method, url),
        response: HttpResponse::new(status).with_body(json!({})),
    }
}

fn project(text: &str) -> (tempfile::TempDir, PathBuf) {
    let tmp = tempfile::tempdir().unwrap();
    let root = tmp.path().join("project");
    fs::create_dir_all(root.join(".github/moai")).unwrap();
    fs::write(root.join(".github/moai/labels.yml"), text).unwrap();
    (tmp, root)
}

#[test]
fn the_shipped_label_list_parses() {
    let specs = labels::parse_labels_yml(&shipped()).unwrap();
    assert_eq!(specs.len(), 15);
    assert!(specs.iter().any(|s| s.name == "needs:decision"));
    assert!(specs
        .iter()
        .all(|s| s.color.len() == 6 && !s.description.is_empty()));
}

#[test]
fn other_label_file_shapes_are_refused_not_skipped() {
    // This repository's own list (reused/added) is not the product shape.
    let repo =
        fs::read_to_string(Path::new(env!("CARGO_MANIFEST_DIR")).join(".github/moai/labels.yml"))
            .unwrap();
    assert!(labels::parse_labels_yml(&repo).is_err());
    for bad in [
        SMALL.replace("schema_version: 2", "schema_version: 1"),
        SMALL.replace("  color: a2eeef\n", ""),
        SMALL.replace("color: a2eeef", "color: '#a2eeef'"),
        SMALL.replace("color: a2eeef", "color: blue"),
        SMALL.replace(
            "  description: Drop everything\n",
            "  description: Drop everything\n  default: true\n",
        ),
        SMALL.replace("area:docs", "Type:Feature"),
        SMALL.replace("- name: priority:P0", "-   name: priority:P0"),
        SMALL.replace("  color: b60205", "\tcolor: b60205"),
        SMALL.replace("labels:\n", "labels: []\n"),
        SMALL.replace("- name: area:docs", "- {name: area:docs}"),
        format!("{SMALL}extra: 1\n"),
        String::new(),
    ] {
        assert!(labels::parse_labels_yml(&bad).is_err(), "accepted:\n{bad}");
    }
}

#[test]
fn quoted_scalars_parse_like_plain_ones() {
    let text = SMALL.replace(
        "  description: Drop everything",
        "  description: 'Drop everything, it''s urgent'",
    );
    let specs = labels::parse_labels_yml(&text).unwrap();
    assert_eq!(specs[1].description, "Drop everything, it's urgent");
}

#[test]
fn plan_compares_names_case_insensitively_and_never_deletes() {
    let specs = labels::parse_labels_yml(SMALL).unwrap();
    let remotes: Vec<RemoteLabel> = vec![
        // same label, different case and an upper-case colour: unchanged
        serde_json::from_value(remote(
            "Type:Feature",
            "A2EEEF",
            Some("New feature or behaviour"),
        ))
        .unwrap(),
        // listed, colour differs: update
        serde_json::from_value(remote("priority:P0", "000000", None)).unwrap(),
        // not listed: reported, never deleted
        serde_json::from_value(remote("wontfix", "ffffff", None)).unwrap(),
    ];
    let plan = labels::plan_labels(&request(), &specs, &remotes);
    let actions: BTreeMap<String, String> = plan
        .changes
        .iter()
        .map(|c| (c.name.clone(), c.action.clone()))
        .collect();
    assert_eq!(actions["type:feature"], "unchanged");
    assert_eq!(actions["priority:P0"], "update");
    assert_eq!(actions["area:docs"], "create");
    assert_eq!(plan.unlisted, vec!["wontfix".to_string()]);
    let again = labels::plan_labels(&request(), &specs, &remotes);
    assert_eq!(plan.plan_digest, again.plan_digest);
    // A null remote description equals an empty one.
    let same = vec![serde_json::from_value::<RemoteLabel>(remote("x", "aaaaaa", None)).unwrap()];
    let text = "schema_version: 2\nlabels:\n- name: x\n  color: aaaaaa\n  description: ''\n";
    let p = labels::plan_labels(&request(), &labels::parse_labels_yml(text).unwrap(), &same);
    assert_eq!(p.changes[0].action, "unchanged");
}

#[test]
fn remote_labels_are_read_in_pages_of_100() {
    for total in [100usize, 101] {
        let all: Vec<Value> = (0..total)
            .map(|i| remote(&format!("l{i}"), "aaaaaa", None))
            .collect();
        let second = format!("{BASE}?per_page=100&page=2");
        let mut ex = vec![list_page(
            &first_page(),
            all[..100].to_vec(),
            (total > 100).then_some(second.as_str()),
        )];
        if total > 100 {
            ex.push(list_page(&second, all[100..].to_vec(), None));
        }
        let t = RecordedTransport::new(ex);
        let c = clock();
        let client = Client::new(&t, &c, Config::default(), BTreeMap::new());
        let read = labels::read_remote(&client, &request()).unwrap();
        assert_eq!(read.len(), total);
        assert_eq!(t.remaining(), 0);
    }
}

fn plan_with(root: &Path, remote_labels: Vec<Value>) -> Value {
    let t = RecordedTransport::new(vec![list_page(&first_page(), remote_labels, None)]);
    let c = clock();
    labels::labels_plan_cli(root, request(), &t, &c).unwrap()
}

#[test]
fn apply_creates_and_updates_with_encoded_names_and_sends_no_delete() {
    let (_tmp, root) = project(SMALL);
    let before = vec![
        remote("priority:P0", "000000", None),
        remote("wontfix", "ffffff", None),
    ];
    let plan = plan_with(&root, before.clone());
    let t = RecordedTransport::new(vec![
        list_page(&first_page(), before, None),
        write("POST", BASE, 201),
        write("PATCH", &format!("{BASE}/priority%3AP0"), 200),
        write("POST", BASE, 201),
    ]);
    let c = clock();
    let digest = plan["plan_digest"].as_str().unwrap().to_string();
    let out = labels::labels_apply_cli(&root, plan, &digest, "test:2434", &t, &c).unwrap();
    assert_eq!(out["created"], json!(["type:feature", "area:docs"]));
    assert_eq!(out["updated"], json!(["priority:P0"]));
    let sent = t.sent.borrow();
    assert!(sent.iter().all(|r| r.method != "DELETE"));
    let posts: Vec<&HttpRequest> = sent.iter().filter(|r| r.method == "POST").collect();
    assert_eq!(
        posts[0].body,
        Some(
            json!({"name": "type:feature", "color": "a2eeef", "description": "New feature or behaviour"})
        )
    );
    // Writes are spaced at least one second apart.
    assert!(
        c.slept.borrow().iter().sum::<u64>() >= 2,
        "{:?}",
        c.slept.borrow()
    );
}

#[test]
fn apply_refuses_a_stale_remote_and_a_missing_permission() {
    let (_tmp, root) = project(SMALL);
    let plan = plan_with(&root, vec![]);
    let digest = plan["plan_digest"].as_str().unwrap().to_string();
    // The remote gained a label after the plan: stale, nothing written.
    let t = RecordedTransport::new(vec![list_page(
        &first_page(),
        vec![remote("area:docs", "0075ca", Some("Documentation"))],
        None,
    )]);
    let c = clock();
    let err =
        labels::labels_apply_cli(&root, plan.clone(), &digest, "test:2434", &t, &c).unwrap_err();
    assert!(err.contains("stale"), "{err}");
    assert!(t.sent.borrow().iter().all(|r| r.method == "GET"));
    // A changed file is stale too.
    let t = RecordedTransport::new(vec![list_page(&first_page(), vec![], None)]);
    fs::write(
        root.join(".github/moai/labels.yml"),
        SMALL.replace("0075ca", "0075cb"),
    )
    .unwrap();
    let err = labels::labels_apply_cli(&root, plan.clone(), &digest, "test:2434", &t, &clock())
        .unwrap_err();
    assert!(err.contains("stale"), "{err}");
    // No permission reference: refused before any request.
    let t = RecordedTransport::new(vec![]);
    let err = labels::labels_apply_cli(&root, plan, &digest, "", &t, &clock()).unwrap_err();
    assert!(err.contains("permission"), "{err}");
    assert!(t.sent.borrow().is_empty());
}

#[test]
fn a_failed_write_reports_what_was_applied_and_a_rerun_does_the_rest() {
    let (_tmp, root) = project(SMALL);
    let plan = plan_with(&root, vec![]);
    let digest = plan["plan_digest"].as_str().unwrap().to_string();
    let t = RecordedTransport::new(vec![
        list_page(&first_page(), vec![], None),
        write("POST", BASE, 201),
        write("POST", BASE, 422),
    ]);
    let err =
        labels::labels_apply_cli(&root, plan, &digest, "test:2434", &t, &clock()).unwrap_err();
    assert!(err.contains("priority:P0"), "{err}");
    assert!(err.contains("type:feature"), "{err}");
    // The rerun plans only what is still missing.
    let plan = plan_with(
        &root,
        vec![remote(
            "type:feature",
            "a2eeef",
            Some("New feature or behaviour"),
        )],
    );
    let todo: Vec<&str> = plan["changes"]
        .as_array()
        .unwrap()
        .iter()
        .filter(|c| c["action"] != "unchanged")
        .map(|c| c["name"].as_str().unwrap())
        .collect();
    assert_eq!(todo, vec!["priority:P0", "area:docs"]);
}

/// Runs the CLI with a stub `gh` that only records that it was called.
fn cli(gh: &Path, args: &[&str], input: &Value) -> Value {
    let mut child = Command::new(env!("CARGO_BIN_EXE_moai-cli"))
        .args(args)
        .env("MOAI_FORGE_GH_BIN", gh)
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

#[test]
fn the_commands_exist_and_refuse_a_bad_request_without_a_network_call() {
    let (tmp, root) = project(SMALL);
    let marker = tmp.path().join("gh-was-called");
    let gh = tmp.path().join("gh");
    fs::write(
        &gh,
        format!("#!/bin/sh\ntouch '{}'\nexit 1\n", marker.display()),
    )
    .unwrap();
    #[cfg(unix)]
    {
        use std::os::unix::fs::PermissionsExt;
        fs::set_permissions(&gh, fs::Permissions::from_mode(0o755)).unwrap();
    }
    let env = cli(
        &gh,
        &["labels-plan", root.to_str().unwrap()],
        &json!({"schema_version": 1, "owner": "acme", "repo": "../x"}),
    );
    assert_eq!(env["ok"], false, "{env}");
    assert!(env.to_string().contains("owner and repo"), "{env}");
    let env = cli(
        &gh,
        &[
            "labels-apply",
            root.to_str().unwrap(),
            "--plan-digest",
            "sha256:00",
            "--permission-reference",
            "",
        ],
        &json!({}),
    );
    assert_eq!(env["ok"], false, "{env}");
    assert!(env.to_string().contains("permission"), "{env}");
    assert!(!marker.exists(), "gh must not run for a refused request");
}
