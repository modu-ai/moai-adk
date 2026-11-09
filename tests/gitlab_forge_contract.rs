//! GitLab forge contract (Issue #2171, REQ-B4-01) — AC-1: the trait compiles
//! as a declaration surface and the fixture loads and conforms to the schema's
//! required structure. The declaration is metadata only: no GitLab transport,
//! client or call path exists in this crate (REQ-B4-01), so this test uses
//! nothing beyond the declaration surface and the fixture file.

use moai_cli::forge::gitlab::{CapabilityStatus, ForgeCapabilities, GitLabContract};

#[test]
fn gitlab_contract_declares_documented_capabilities() {
    let contract = GitLabContract;
    assert_eq!(ForgeCapabilities::provider(&contract), "gitlab");
    let caps = ForgeCapabilities::capabilities(&contract);
    assert_eq!(caps.len(), 3, "merge_request, pipeline, note");
    for c in caps {
        assert_eq!(c.status, CapabilityStatus::Documented);
    }
    let names: Vec<&str> = caps.iter().map(|c| c.name).collect();
    assert_eq!(names, ["merge_request", "pipeline", "note"]);
}

#[test]
fn fixture_loads_and_conforms_to_the_schema_structure() {
    let fixture_path = std::path::Path::new(env!("CARGO_MANIFEST_DIR"))
        .join("tests/fixtures/gitlab/sample-fixture.json");
    let fixture: serde_json::Value =
        serde_json::from_slice(&std::fs::read(&fixture_path).unwrap()).unwrap();
    assert_eq!(fixture["provider"], "gitlab");

    // Required top-level structure (schemas/gitlab-fixture.schema.json).
    let caps = fixture["capabilities"].as_array().unwrap();
    assert_eq!(
        caps.len(),
        3,
        "the fixture declares the same three capabilities"
    );
    for c in caps {
        assert!(
            c["status"] == "documented" || c["status"] == "unknown",
            "capability status must be one of the schema's enum values"
        );
    }
    for name in ["merge_request", "pipeline", "note"] {
        let res = &fixture["resources"][name];
        assert!(
            res["documented_in"].is_string(),
            "{name} must cite where its shape is documented"
        );
        assert!(
            res["fixture_example"].is_object(),
            "{name} must carry a fixture example object"
        );
    }

    // The fixture's declared capabilities match the trait declaration.
    let declared = ForgeCapabilities::capabilities(&GitLabContract);
    for (fixture_cap, trait_cap) in caps.iter().zip(declared.iter()) {
        assert_eq!(
            fixture_cap["name"].as_str().unwrap(),
            trait_cap.name,
            "fixture and trait must declare the same capability names"
        );
        match trait_cap.status {
            CapabilityStatus::Documented => {
                assert_eq!(fixture_cap["status"], "documented")
            }
            CapabilityStatus::Unknown => {
                assert_eq!(fixture_cap["status"], "unknown")
            }
        }
    }
}
