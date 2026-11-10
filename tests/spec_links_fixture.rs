//! C5 #2181 fixture: explicit REQ ↔ code ↔ test links and commit → issue mentions.
//!
//! A temporary git repository holds every explicit source next to decoys that look
//! like links but are not explicit. The extracted edge set must EQUAL the
//! hand-enumerated ground truth — precision and recall both 100% (AC-1; REQ-C5-01:
//! no inferred link) — and every non-link must surface as a `missing` row (AC-2).

use moai_cli::spec_links::extract;
use serde_json::Value;
use std::{collections::BTreeSet, fs, path::Path, process::Command};
use tempfile::TempDir;

const OBSERVED_AT: &str = "2026-11-10T00:00:00Z";

fn git(dir: &Path, args: &[&str]) -> String {
    let out = Command::new("git")
        .arg("-C")
        .arg(dir)
        .args([
            "-c",
            "user.name=fixture",
            "-c",
            "user.email=fixture@example.invalid",
            "-c",
            "commit.gpgsign=false",
        ])
        .args(args)
        .env("GIT_AUTHOR_DATE", "1735689600 +0000")
        .env("GIT_COMMITTER_DATE", "1735689600 +0000")
        .env("GIT_CONFIG_GLOBAL", "/dev/null")
        .env("GIT_CONFIG_NOSYSTEM", "1")
        .output()
        .unwrap();
    assert!(
        out.status.success(),
        "git {args:?}: {}",
        String::from_utf8_lossy(&out.stderr)
    );
    String::from_utf8(out.stdout).unwrap().trim().to_string()
}

fn write(root: &Path, rel: &str, body: &[u8]) {
    let p = root.join(rel);
    fs::create_dir_all(p.parent().unwrap()).unwrap();
    fs::write(p, body).unwrap();
}

/// Commits everything with `message`; returns the new HEAD sha.
fn commit(root: &Path, message: &str) -> String {
    let msg = root.join(".git/fixture-msg");
    fs::write(&msg, message).unwrap();
    git(root, &["add", "-A"]);
    git(
        root,
        &[
            "commit",
            "-q",
            "--allow-empty",
            "--cleanup=verbatim",
            "-F",
            msg.to_str().unwrap(),
        ],
    );
    git(root, &["rev-parse", "HEAD"])
}

struct Fixture {
    _dir: TempDir,
    root: std::path::PathBuf,
    commits: Vec<String>,
}

fn fixture() -> Fixture {
    let dir = TempDir::new().unwrap();
    let root = dir.path().join("repo");
    fs::create_dir_all(&root).unwrap();
    git(&root, &["init", "-q", "-b", "main"]);

    write(
        &root,
        "src/payment.rs",
        br#"// @MOAI:REQ REQ-PAY-01
pub fn pay() {}
// @MX:REQ REQ-PAY-02, REQ-PAY-03
const NOTE: &str = "REQ-PAY-09 appears in a string without a marker";
// see REQ-PAY-08 in a plain comment, no marker
// @MX:NOTE REQ-PAY-07 is under the wrong tag kind
// @MOAI:REQ req-pay-1
// @MOAI:REQ
// @MOAI:req lowercase kind is not grammar v1
fn req_pay_06_helper() {}
"#,
    );
    write(
        &root,
        "tests/payment_test.rs",
        br#"#[test]
fn req_pay_01_charges_once() {}

#[cfg(unix)]
#[test]
fn checks_req_pay_02_and_req_pay_04() {}

/// A feature name next to tagged code is not an explicit link.
#[test]
fn payment_flow_works() {}
"#,
    );
    write(
        &root,
        "tests/flow_test.py",
        b"def test_req_pay_05_python():\n    pass\n\ndef helper_req_pay_01():\n    pass\n",
    );
    write(
        &root,
        "docs/spec.md",
        b"@MOAI:REQ REQ-PAY-10 in a document, not code\n",
    );
    write(&root, "src/blob.rs", &[0xff, 0xfe, 0x00, 0x41]);

    let c1 = commit(&root, "fixture: init\n\nMoai-Claim: 2181#6753145385\n");
    let c2 = commit(
        &root,
        "fixture: body reference\n\nRefs #77\n\nThis closing paragraph is prose, not a trailer.\n",
    );
    let c3 = commit(&root, "fixture: refs trailer\n\nRefs: #88\n");
    let c4 = commit(&root, "fixture: malformed claim\n\nMoai-Claim: abc#1\n");
    Fixture {
        _dir: dir,
        root,
        commits: vec![c1, c2, c3, c4],
    }
}

fn edge_set(doc: &Value) -> BTreeSet<(String, String, String)> {
    doc["edges"]
        .as_array()
        .unwrap()
        .iter()
        .map(|e| {
            (
                e["kind"].as_str().unwrap().to_string(),
                e["source_id"].as_str().unwrap().to_string(),
                e["target_id"].as_str().unwrap().to_string(),
            )
        })
        .collect()
}

fn missing_set(doc: &Value) -> BTreeSet<String> {
    doc["missing"]
        .as_array()
        .unwrap()
        .iter()
        .map(|m| {
            let kind = m["kind"].as_str().unwrap();
            let key = ["token", "requirement", "issue", "path"]
                .iter()
                .find_map(|k| m[*k].as_str())
                .unwrap_or("");
            format!("{kind} {key}")
        })
        .collect()
}

fn t(kind: &str, source: &str, target: &str) -> (String, String, String) {
    (kind.into(), source.into(), target.into())
}

/// AC-1 / REQ-C5-01: the extracted set equals the ground truth exactly — every
/// explicit link is found and no decoy becomes an edge.
#[test]
fn extracted_links_equal_the_explicit_ground_truth() {
    let f = fixture();
    let head = f.commits.last().unwrap();
    let doc = extract(&f.root, head, OBSERVED_AT).unwrap();
    let [c1, _, c3, _] = &f.commits[..] else {
        unreachable!()
    };
    let expected: BTreeSet<_> = [
        t(
            "references_requirement",
            "file:src/payment.rs",
            "req:REQ-PAY-01",
        ),
        t(
            "references_requirement",
            "file:src/payment.rs",
            "req:REQ-PAY-02",
        ),
        t(
            "references_requirement",
            "file:src/payment.rs",
            "req:REQ-PAY-03",
        ),
        t(
            "verifies",
            "test:tests/payment_test.rs::req_pay_01_charges_once",
            "req:REQ-PAY-01",
        ),
        t(
            "verifies",
            "test:tests/payment_test.rs::checks_req_pay_02_and_req_pay_04",
            "req:REQ-PAY-02",
        ),
        t(
            "verifies",
            "test:tests/payment_test.rs::checks_req_pay_02_and_req_pay_04",
            "req:REQ-PAY-04",
        ),
        t(
            "verifies",
            "test:tests/flow_test.py::test_req_pay_05_python",
            "req:REQ-PAY-05",
        ),
        t("mentions_issue", &format!("commit:{c1}"), "issue:2181"),
        t("mentions_issue", &format!("commit:{c3}"), "issue:88"),
    ]
    .into_iter()
    .collect();
    let got = edge_set(&doc);
    let extra: Vec<_> = got.difference(&expected).collect();
    let lost: Vec<_> = expected.difference(&got).collect();
    assert!(
        extra.is_empty() && lost.is_empty(),
        "extra (precision): {extra:?}\nlost (recall): {lost:?}"
    );
    for e in doc["edges"].as_array().unwrap() {
        assert_eq!(e["basis"], "explicit", "{e}");
        assert_eq!(e["verification"], "unreviewed", "{e}");
        assert_eq!(e["extractor"]["name"], "spec-links");
    }
    assert_eq!(doc["basis_inferred"], 0);
}

/// AC-2: everything that resembles a link without being explicit is reported.
#[test]
fn non_explicit_references_are_reported_as_missing() {
    let f = fixture();
    let head = f.commits.last().unwrap();
    let doc = extract(&f.root, head, OBSERVED_AT).unwrap();
    let expected: BTreeSet<String> = [
        "malformed_req_ref req-pay-1",
        "req_tag_without_id src/payment.rs",
        "invalid_marker src/payment.rs",
        "unreadable_file src/blob.rs",
        "issue_ref_not_trailer 77",
        "malformed_issue_trailer ",
        "req_without_test req:REQ-PAY-03",
        "req_without_code req:REQ-PAY-04",
        "req_without_code req:REQ-PAY-05",
    ]
    .into_iter()
    .map(String::from)
    .collect();
    assert_eq!(missing_set(&doc), expected, "{:#}", doc["missing"]);
}

/// The output is C4 schema-valid and a pure function of tree, history and inputs.
#[test]
fn output_is_schema_valid_and_deterministic() {
    let f = fixture();
    let head = f.commits.last().unwrap();
    let a = extract(&f.root, head, OBSERVED_AT).unwrap();
    let b = extract(&f.root, head, OBSERVED_AT).unwrap();
    assert_eq!(
        serde_json::to_string(&a).unwrap(),
        serde_json::to_string(&b).unwrap()
    );
    assert_eq!(
        moai_cli::ontology::validate_document(&a).unwrap(),
        (
            a["nodes"].as_array().unwrap().len(),
            a["edges"].as_array().unwrap().len()
        )
    );
    // Evidence points at the exact line and content.
    let refs: Vec<&str> = a["edges"]
        .as_array()
        .unwrap()
        .iter()
        .flat_map(|e| e["evidence_refs"].as_array().unwrap())
        .map(|r| r.as_str().unwrap())
        .collect();
    assert!(
        refs.iter()
            .any(|r| r.starts_with("file:src/payment.rs#L1@sha256:")),
        "{refs:?}"
    );
    assert!(
        refs.iter()
            .any(|r| r.starts_with("file:src/payment.rs#L3@sha256:")),
        "{refs:?}"
    );
    assert!(refs.iter().any(|r| r.starts_with(&format!(
        "commit:{}#trailer:Moai-Claim@sha256:",
        f.commits[0]
    ))));
}

#[test]
fn a_non_rfc3339_time_is_refused() {
    let f = fixture();
    let err = extract(&f.root, "x", "yesterday").unwrap_err();
    assert!(err.contains("RFC3339"), "{err}");
}

/// Real lane messages keep explicit protocol lines outside Git's trailer block.
#[test]
fn protocol_lines_survive_non_trailer_footer_and_separate_coauthor() {
    let f = fixture();
    let footer = commit(
        &f.root,
        "fixture: footer\n\nMoai-Claim: 901#123\nRefs: #902\n\n🗿 MoAI\n",
    );
    let coauthor = commit(
        &f.root,
        "fixture: coauthor\n\nMoai-Claim: 903#456\nRefs: #904\n\nCo-Authored-By: Fixture <fixture@example.invalid>\n",
    );
    for sha in [&footer, &coauthor] {
        let parsed = git(
            &f.root,
            &["show", "-s", "--format=%(trailers:only=true)", sha],
        );
        assert!(
            !parsed.contains("Moai-Claim:"),
            "fixture must expose the gap"
        );
    }
    let before = extract(&f.root, &coauthor, OBSERVED_AT).unwrap();
    let baseline = edge_set(&before);
    let expected: BTreeSet<_> = [
        t("mentions_issue", &format!("commit:{footer}"), "issue:901"),
        t("mentions_issue", &format!("commit:{footer}"), "issue:902"),
        t("mentions_issue", &format!("commit:{coauthor}"), "issue:903"),
        t("mentions_issue", &format!("commit:{coauthor}"), "issue:904"),
    ]
    .into_iter()
    .collect();
    let got: BTreeSet<_> = baseline
        .into_iter()
        .filter(|(_, source, _)| {
            source == &format!("commit:{footer}") || source == &format!("commit:{coauthor}")
        })
        .collect();
    assert_eq!(got, expected, "every explicit body line must survive");
    assert_eq!(before["basis_inferred"], 0);
    for edge in before["edges"].as_array().unwrap() {
        assert_eq!(edge["basis"], "explicit");
        assert_eq!(edge["verification"], "unreviewed");
        if edge["source_id"] == format!("commit:{footer}")
            || edge["source_id"] == format!("commit:{coauthor}")
        {
            let sha = edge["source_id"]
                .as_str()
                .unwrap()
                .strip_prefix("commit:")
                .unwrap();
            let body = git(&f.root, &["show", "-s", "--format=%B", sha]);
            let digest = moai_cli::digest(format!("{body}\n").as_bytes());
            let key = if edge["target_id"] == "issue:901" || edge["target_id"] == "issue:903" {
                "Moai-Claim"
            } else {
                "Refs"
            };
            assert_eq!(
                edge["evidence_refs"],
                serde_json::json!([format!("commit:{sha}#trailer:{key}@{digest}")])
            );
        }
    }
    assert_eq!(before, extract(&f.root, &coauthor, OBSERVED_AT).unwrap());
}

/// Only exact protocol lines count; prose, malformed ids and folded values do not.
#[test]
fn protocol_body_lines_refuse_near_matches_and_merge_duplicate_evidence() {
    let f = fixture();
    let sha = commit(
        &f.root,
        "fixture: protocol grammar\n\nMoai-Claim: 910#100\nMoai-Claim: 910#100\nRefs: #910\nRefs: #910\nMoai-Claim: 911#\nMoai-Claim: 912#abc\nMoai-Claim: 913#123#4\nMoai-Claim: 914#123 extra\nMoai-Claim: 915#123 \nMoai-Claim:  916#123\nMoai-Claim: 917#123\t\nMoai-Claim: ９１８#123\nRefs: #919 extra\nRefs: #920 \nRefs: #９２１\nRefs: #\nRefs:#922\nMoai-Claim:923#123\nMoai-Claim: 924#\n 123\nprose Moai-Claim: 925#123\n> Moai-Claim: 926#123\n Moai-Claim: 927#123\nmoai-claim: 928#123\nRefs #929\n\n🗿 MoAI\n",
    );
    let doc = extract(&f.root, &sha, OBSERVED_AT).unwrap();
    let got: BTreeSet<_> = edge_set(&doc)
        .into_iter()
        .filter(|(_, source, _)| source == &format!("commit:{sha}"))
        .collect();
    assert_eq!(
        got,
        BTreeSet::from([t("mentions_issue", &format!("commit:{sha}"), "issue:910")])
    );
    let edge = doc["edges"]
        .as_array()
        .unwrap()
        .iter()
        .find(|e| e["source_id"] == format!("commit:{sha}"))
        .unwrap();
    assert_eq!(
        edge["evidence_refs"].as_array().unwrap().len(),
        2,
        "one evidence per explicit key, not per duplicate line"
    );
    let malformed: BTreeSet<_> = doc["missing"]
        .as_array()
        .unwrap()
        .iter()
        .filter(|row| row["commit"] == sha && row["kind"] == "malformed_issue_trailer")
        .map(|row| row["trailer"].as_str().unwrap())
        .collect();
    for line in [
        "Moai-Claim: 911#",
        "Moai-Claim: 912#abc",
        "Moai-Claim: 913#123#4",
        "Moai-Claim: ９１８#123",
        "Refs:#922",
        "Moai-Claim:923#123",
        "Moai-Claim: 924#",
    ] {
        assert!(
            malformed.contains(line),
            "unreported malformed protocol: {line}"
        );
    }
    assert!(doc["missing"]
        .as_array()
        .unwrap()
        .iter()
        .any(|row| row["commit"] == sha
            && row["kind"] == "issue_ref_not_trailer"
            && row["issue"] == "929"));
    assert_eq!(doc["basis_inferred"], 0);
}

#[test]
fn review_test_declarations_are_lexical_scoped_and_keep_attribute_blocks() {
    let dir = TempDir::new().unwrap();
    let root = dir.path().join("repo");
    fs::create_dir(&root).unwrap();
    git(&root, &["init", "-q", "-b", "main"]);
    write(
        &root,
        "tests/scoped.rs",
        br####"const EXAMPLE: &str = r###"
#[test]
fn req_decoy_01() {}
"###;
/* outer /* nested */ #[test]
fn req_decoy_02() {}
*/
macro_rules! example {
    () => { #[test] fn req_decoy_04() {} };
}
// An in-progress file may carry an incomplete raw-identifier prefix.
r##
mod first {
    #[test]
    #[cfg_attr(
        any(),
        ignore
    )]

    fn req_real_01() {}
}
mod second {
    #[tokio::test]
    async fn req_real_01() {}
}
"####,
    );
    write(
        &root,
        "tests/scoped_test.py",
        b"EXAMPLE = '''\ndef test_req_decoy_03():\n    pass\n'''\nclass First:\n    def test_req_real_02(self):\n        pass\nclass Second:\n    def test_req_real_02(self):\n        pass\n",
    );
    let head = commit(&root, "fixture: scoped declarations");
    let doc = extract(&root, &head, OBSERVED_AT).unwrap();
    let expected = BTreeSet::from([
        t(
            "verifies",
            "test:tests/scoped.rs::first::req_real_01",
            "req:REQ-REAL-01",
        ),
        t(
            "verifies",
            "test:tests/scoped.rs::second::req_real_01",
            "req:REQ-REAL-01",
        ),
        t(
            "verifies",
            "test:tests/scoped_test.py::First::test_req_real_02",
            "req:REQ-REAL-02",
        ),
        t(
            "verifies",
            "test:tests/scoped_test.py::Second::test_req_real_02",
            "req:REQ-REAL-02",
        ),
    ]);
    assert_eq!(edge_set(&doc), expected);
}

#[cfg(unix)]
#[test]
fn review_external_symlink_never_supplies_verification_evidence() {
    let dir = TempDir::new().unwrap();
    let root = dir.path().join("repo");
    fs::create_dir(&root).unwrap();
    git(&root, &["init", "-q", "-b", "main"]);
    fs::create_dir(root.join("tests")).unwrap();
    let external = dir.path().join("external.py");
    fs::write(&external, "def test_req_external_01():\n    pass\n").unwrap();
    std::os::unix::fs::symlink(&external, root.join("tests/link_test.py")).unwrap();
    let head = commit(&root, "fixture: tracked external symlink");
    let before = extract(&root, &head, OBSERVED_AT).unwrap();
    assert_eq!(edge_set(&before), BTreeSet::new());
    assert!(before["missing"]
        .as_array()
        .unwrap()
        .iter()
        .any(|row| { row["kind"] == "unreadable_file" && row["path"] == "tests/link_test.py" }));
    fs::write(&external, "def test_req_external_99():\n    pass\n").unwrap();
    assert_eq!(extract(&root, &head, OBSERVED_AT).unwrap(), before);
}
