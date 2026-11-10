//! OMP host adapter fixtures (#2362). Pure composition only: no omp binary is started and no
//! user setting is read. Limits are tested with literal values, not the module constants.
use moai_cli::hosts::omp::{compose, OmpRequest, Role};
use serde_json::Value;

fn argv(v: &Value) -> Vec<String> {
    serde_json::from_value(v["argv"].clone()).unwrap()
}

fn notes(v: &Value) -> Vec<String> {
    serde_json::from_value(v["notes"].clone()).unwrap()
}

fn req() -> OmpRequest {
    OmpRequest::default()
}

const FORBIDDEN: [&str; 3] = ["--auto-approve", "--approval-mode", "--plan-yolo"];

fn assert_no_approval_flag(v: &Value) {
    for t in argv(v) {
        for f in FORBIDDEN {
            assert!(!t.starts_with(f), "approval flag {t:?} in argv");
        }
    }
}

#[test]
fn bare_request_is_just_the_binary_with_disclosed_approval_default() {
    let v = compose(&req()).unwrap();
    assert_eq!(argv(&v), ["omp"]);
    assert_eq!(v["host"], "omp");
    assert_eq!(v["env_keys"], serde_json::json!([]));
    assert_eq!(v["executed"], false);
    assert_eq!(v["host_loaded"], false);
    let n = notes(&v).join("\n");
    assert!(
        n.contains("no role: the launcher passes no permission flag"),
        "{n}"
    );
    assert!(n.contains("default in omp 18.18.6 is yolo"), "{n}");
    assert_eq!(v["role"], Value::Null);
    assert!(n.contains("no --model"), "{n}");
    assert!(n.contains("no --thinking"), "{n}");
    assert_no_approval_flag(&v);
}

#[test]
fn full_request_composes_single_token_flags_in_order() {
    let r = OmpRequest {
        model: Some("anthropic/claude-opus-5-5".into()),
        thinking: Some("high".into()),
        prompt: Some("Read AGENTS.md first.\nThen report.".into()),
        omp_bin: "/opt/omp/bin/omp".into(),
        ..req()
    };
    let v = compose(&r).unwrap();
    assert_eq!(
        argv(&v),
        [
            "/opt/omp/bin/omp",
            "--model=anthropic/claude-opus-5-5",
            "--thinking=high",
            "Read AGENTS.md first.\nThen report.",
        ]
    );
    assert_no_approval_flag(&v);
}

#[test]
fn name_is_validated_and_reported_but_never_passed() {
    let r = OmpRequest {
        name: Some("agent-7".into()),
        ..req()
    };
    let v = compose(&r).unwrap();
    assert_eq!(v["name"], "agent-7");
    assert_eq!(argv(&v), ["omp"]);
    assert!(notes(&v)
        .join("\n")
        .contains("name delivery to omp is unverified"));
}

// Launcher rule (#2359 comment 6760976883): -l / -a open with full access, no role adds no
// permission flag. Role sessions are compared byte for byte.
#[test]
fn leader_role_adds_exactly_one_auto_approve() {
    let r = OmpRequest {
        role: Some(Role::Leader),
        name: Some("leader".into()),
        ..req()
    };
    let v = compose(&r).unwrap();
    assert_eq!(argv(&v), ["omp", "--auto-approve"]);
    assert_eq!(v["role"], "leader");
    assert_eq!(v["name"], "leader");
    assert!(notes(&v)
        .join("\n")
        .contains("full access via --auto-approve"));
}

#[test]
fn agent_role_adds_exactly_one_auto_approve_before_other_flags() {
    let r = OmpRequest {
        role: Some(Role::Agent),
        name: Some("agent-12".into()),
        model: Some("opus".into()),
        thinking: Some("high".into()),
        prompt: Some("Read AGENTS.md first.".into()),
        ..req()
    };
    let v = compose(&r).unwrap();
    assert_eq!(
        argv(&v),
        [
            "omp",
            "--auto-approve",
            "--model=opus",
            "--thinking=high",
            "Read AGENTS.md first.",
        ]
    );
    assert_eq!(v["role"], "agent");
}

#[test]
fn role_with_keep_prompts_is_refused_until_decided() {
    for role in [Role::Leader, Role::Agent] {
        let r = OmpRequest {
            role: Some(role),
            keep_prompts: true,
            name: Some("leader".into()),
            ..req()
        };
        let e = compose(&r).unwrap_err();
        assert!(e.contains("HOLD"), "{e}");
    }
}

#[test]
fn keep_prompts_without_role_adds_nothing() {
    let r = OmpRequest {
        keep_prompts: true,
        ..req()
    };
    let v = compose(&r).unwrap();
    assert_eq!(argv(&v), ["omp"]);
}

#[test]
fn role_without_name_is_refused() {
    for role in [Role::Leader, Role::Agent] {
        let r = OmpRequest {
            role: Some(role),
            ..req()
        };
        assert!(compose(&r).is_err(), "{role:?} without name accepted");
    }
}

#[test]
fn role_still_validates_name() {
    let r = OmpRequest {
        role: Some(Role::Agent),
        name: Some("--approval-mode=yolo".into()),
        ..req()
    };
    assert!(compose(&r).is_err());
}

#[test]
fn role_json_is_lowercase_and_unknown_role_is_rejected() {
    let r: OmpRequest = serde_json::from_str(r#"{"role":"agent","name":"agent-3"}"#).unwrap();
    assert_eq!(r.role, Some(Role::Agent));
    assert!(serde_json::from_str::<OmpRequest>(r#"{"role":"Leader"}"#).is_err());
    assert!(serde_json::from_str::<OmpRequest>(r#"{"role":"admin"}"#).is_err());
}

#[test]
fn every_listed_thinking_level_is_accepted() {
    for t in [
        "off", "minimal", "low", "medium", "high", "xhigh", "max", "auto",
    ] {
        let r = OmpRequest {
            thinking: Some(t.into()),
            ..req()
        };
        assert_eq!(argv(&compose(&r).unwrap())[1], format!("--thinking={t}"));
    }
}

#[test]
fn thinking_near_misses_are_rejected() {
    for t in [
        "",
        "Medium",
        "HIGH",
        "ultracode",
        "ultra",
        "none",
        "high ",
        " high",
        "xhigh\n",
        "-high",
    ] {
        let r = OmpRequest {
            thinking: Some(t.into()),
            ..req()
        };
        assert!(compose(&r).is_err(), "thinking {t:?} accepted");
    }
}

#[test]
fn field_length_boundary_128_accepted_129_rejected() {
    let ok = "m".repeat(128);
    let long = "m".repeat(129);
    for (field, value, accepted) in [
        ("model", &ok, true),
        ("model", &long, false),
        ("name", &ok, true),
        ("name", &long, false),
        ("bin", &ok, true),
        ("bin", &long, false),
    ] {
        let mut r = req();
        match field {
            "model" => r.model = Some(value.clone()),
            "name" => r.name = Some(value.clone()),
            _ => r.omp_bin = value.clone(),
        }
        assert_eq!(compose(&r).is_ok(), accepted, "{field} len {}", value.len());
    }
}

#[test]
fn malformed_model_name_and_bin_are_rejected() {
    for bad in [
        "",
        "-x",
        "--auto-approve",
        "a b",
        "a\tb",
        "a\u{1b}b",
        "a\nb",
    ] {
        let model = OmpRequest {
            model: Some(bad.into()),
            ..req()
        };
        assert!(compose(&model).is_err(), "model {bad:?} accepted");
        let name = OmpRequest {
            name: Some(bad.into()),
            ..req()
        };
        assert!(compose(&name).is_err(), "name {bad:?} accepted");
        let bin = OmpRequest {
            omp_bin: bad.into(),
            ..req()
        };
        assert!(compose(&bin).is_err(), "bin {bad:?} accepted");
    }
}

#[test]
fn prompt_cannot_smuggle_a_flag_or_file_include() {
    for bad in [
        "--auto-approve",
        "--approval-mode=yolo",
        "--plan-yolo",
        " --auto-approve",
        "\t--plan-yolo",
        "\n-p",
        "-c",
        "@/etc/passwd",
        " @secret.md",
        "",
        "   ",
        "ok\u{1b}[2J",
        "ok\u{0}",
    ] {
        let r = OmpRequest {
            prompt: Some(bad.into()),
            ..req()
        };
        assert!(compose(&r).is_err(), "prompt {bad:?} accepted");
    }
}

#[test]
fn prompt_length_boundary_4096_accepted_4097_rejected() {
    let ok = OmpRequest {
        prompt: Some("p".repeat(4096)),
        ..req()
    };
    assert!(compose(&ok).is_ok());
    let long = OmpRequest {
        prompt: Some("p".repeat(4097)),
        ..req()
    };
    assert!(compose(&long).is_err());
}

#[test]
fn no_valid_combination_carries_an_approval_flag() {
    let models = [None, Some("opus"), Some("openai/gpt-5.2")];
    let levels = [None, Some("off"), Some("max"), Some("auto")];
    let prompts = [None, Some("approve all tools"), Some("use yolo mode")];
    let names = [None, Some("leader")];
    for m in models {
        for t in levels {
            for p in prompts {
                for n in names {
                    let r = OmpRequest {
                        model: m.map(Into::into),
                        thinking: t.map(Into::into),
                        prompt: p.map(Into::into),
                        name: n.map(Into::into),
                        ..req()
                    };
                    assert_no_approval_flag(&compose(&r).unwrap());
                }
            }
        }
    }
}

#[test]
fn empty_json_request_defaults_to_omp_binary() {
    let r: OmpRequest = serde_json::from_str("{}").unwrap();
    assert_eq!(r, OmpRequest::default());
    assert_eq!(argv(&compose(&r).unwrap()), ["omp"]);
}
