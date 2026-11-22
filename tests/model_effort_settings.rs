//! SPEC-20261118-model-effort-settings (card C-47, slice 2): the support table, the defaults and
//! the `launcher.models.<tool>.<role>` key of the user launcher settings. The module is pure
//! (no file access), so every test builds a launcher object in memory.
//!
//! SPEC-20261121-codex-launcher-and-models-removed: Codex is no longer a tool. A `codex` value
//! saved by an earlier version is ignored on read (it never refuses a Claude, OMP or GLM launch)
//! and is left out of the file when another tool's choice is written (leader decision D2).
use serde_json::{json, Map, Value};

use moai_cli::model_effort_settings::{
    default_choice, resolve, saved_choice, supported, unsupported_entries, with_choice, Choice,
    Source,
};
use moai_cli::model_effort_settings::{ROLES, TOOLS};

/// A secret-looking marker: it is put into every refused value, and no refusal may contain it.
const SECRET: &str = "SECRET-saved-text";

fn object(value: Value) -> Map<String, Value> {
    match value {
        Value::Object(map) => map,
        other => panic!("test fixture must be an object, got {other}"),
    }
}

fn choice(model: &str, effort: &str) -> Choice {
    Choice {
        model: model.to_string(),
        effort: effort.to_string(),
    }
}

/// The shapes a `launcher.models.codex` value of an earlier version can have in a real file: a
/// valid entry, damaged entries, and text that looks like a secret. All are ignored on read.
fn old_codex_shapes() -> Vec<(&'static str, Value)> {
    vec![
        (
            "a valid old entry",
            json!({"agent": {"model": "gpt-6.2-sol", "effort": "max"}}),
        ),
        ("not an object", json!(SECRET)),
        ("a role that is not an object", json!({"agent": SECRET})),
        (
            "an unknown role",
            json!({SECRET: {"model": "gpt-6.1-sol", "effort": "high"}}),
        ),
        (
            "an unsupported model and effort",
            json!({"leader": {"model": SECRET, "effort": SECRET}}),
        ),
        (
            "an extra field",
            json!({"agent": {"model": "gpt-6.1-sol", "effort": "high", SECRET: 1}}),
        ),
        ("an empty object", json!({})),
    ]
}

#[test]
fn no_saved_value_answers_the_default_for_every_tool_and_role() {
    let expected = [
        ("claude", "leader", "claude-opus-5-5", "high"),
        ("claude", "agent", "claude-sonnet-5-5", "high"),
        ("omp", "leader", "zai/glm-5.3", "high"),
        ("omp", "agent", "zai/glm-5.3-flash", "max"),
        ("glm", "leader", "zai/glm-5.3", "high"),
        ("glm", "agent", "zai/glm-5.3-flash", "max"),
    ];
    assert_eq!(expected.len(), TOOLS.len() * ROLES.len());
    let other_keys = object(json!({"role_access": "prompt", "advisor": "fable"}));
    for (tool, role, model, effort) in expected {
        let want = (choice(model, effort), Source::Default);
        assert_eq!(
            resolve(tool, role, None).unwrap(),
            want,
            "{tool} {role}: no launcher"
        );
        assert_eq!(
            resolve(tool, role, Some(&other_keys)).unwrap(),
            want,
            "{tool} {role}: launcher without models"
        );
        assert_eq!(default_choice(tool, role), Some(choice(model, effort)));
    }
}

#[test]
fn a_saved_supported_value_wins_and_is_marked_saved() {
    let launcher = object(json!({
        "models": {"glm": {"agent": {"model": "zai/glm-5.3", "effort": "max"}}}
    }));
    assert_eq!(
        resolve("glm", "agent", Some(&launcher)).unwrap(),
        (choice("zai/glm-5.3", "max"), Source::Saved)
    );
    // The role and the tool next to it keep their defaults.
    assert_eq!(
        resolve("glm", "leader", Some(&launcher)).unwrap(),
        (choice("zai/glm-5.3", "high"), Source::Default)
    );
    assert_eq!(
        resolve("claude", "agent", Some(&launcher)).unwrap(),
        (choice("claude-sonnet-5-5", "high"), Source::Default)
    );
}

#[test]
fn an_unsupported_or_malformed_saved_value_is_refused_without_quoting_it() {
    let entry = |model: Value, effort: Value| json!({"model": model, "effort": effort});
    let bad: Vec<(&str, Value)> = vec![
        ("models is not an object", json!({"models": SECRET})),
        (
            "tool entry is not an object",
            json!({"models": {"claude": SECRET}}),
        ),
        (
            "role entry is not an object",
            json!({"models": {"glm": {"agent": SECRET}}}),
        ),
        (
            "unsupported model",
            json!({"models": {"glm": {"agent": entry(json!(SECRET), json!("high"))}}}),
        ),
        (
            "unsupported effort",
            json!({"models": {"glm": {"agent": entry(json!("zai/glm-5.3"), json!(SECRET))}}}),
        ),
        (
            "effort the model does not list",
            json!({"models": {"glm": {"agent": entry(json!("zai/glm-5.3-flash"), json!("medium"))}}}),
        ),
        (
            "ultra is not offered",
            json!({"models": {"claude": {"agent": entry(json!("claude-opus-5-5"), json!("ultra"))}}}),
        ),
        (
            "an alias is not an id",
            json!({"models": {"claude": {"leader": entry(json!("opus"), json!("high"))}}}),
        ),
        (
            "a model of another tool",
            json!({"models": {"claude": {"leader": entry(json!("zai/glm-5.3"), json!("high"))}}}),
        ),
        (
            "an old Codex model under Claude",
            json!({"models": {"claude": {"leader": entry(json!("gpt-6.1-sol"), json!("high"))}}}),
        ),
        (
            "model has the wrong type",
            json!({"models": {"omp": {"agent": entry(json!(7), json!("high"))}}}),
        ),
        (
            "effort field missing",
            json!({"models": {"glm": {"agent": {"model": "zai/glm-5.3"}}}}),
        ),
        (
            "extra field",
            json!({"models": {"glm": {"agent": {"model": "zai/glm-5.3", "effort": "high", SECRET: 1}}}}),
        ),
        (
            "unknown tool",
            json!({"models": {SECRET: {"agent": entry(json!("zai/glm-5.3"), json!("high"))}}}),
        ),
        (
            "unknown role",
            json!({"models": {"glm": {SECRET: entry(json!("zai/glm-5.3"), json!("high"))}}}),
        ),
    ];
    for (name, launcher) in bad {
        let launcher = object(launcher);
        // A broken entry is refused for every question, also for a tool it does not belong to,
        // so a broken file is never skipped silently.
        for (tool, role) in [("glm", "agent"), ("claude", "leader")] {
            let message = resolve(tool, role, Some(&launcher))
                .expect_err(&format!("{name}: must be refused for {tool} {role}"));
            assert!(
                message.contains("launcher.models"),
                "{name}: the message must name the key path, got: {message}"
            );
            assert!(
                !message.contains(SECRET),
                "{name}: the message quotes saved text: {message}"
            );
            assert!(
                !message.contains("ultra"),
                "{name}: the message quotes saved text: {message}"
            );
        }
    }
    // Unknown tool or role names asked for by the caller are refused too, and `codex` is one.
    for (tool, role) in [
        ("nope", "agent"),
        ("glm", "nope"),
        ("", ""),
        ("Codex", "agent"),
        ("codex", "agent"),
        ("codex", "leader"),
    ] {
        assert!(resolve(tool, role, None).is_err(), "{tool:?} {role:?}");
        assert_eq!(default_choice(tool, role), None);
    }
}

#[test]
fn setting_a_choice_changes_one_entry_and_keeps_every_other_key() {
    let base = object(json!({
        "role_access": "prompt",
        "advisor": "fable",
        "unrelated": [1, 2, {"a": null}],
        "models": {
            "codex": {"agent": {"model": "gpt-6.2-sol", "effort": "max"}},
            "claude": {"leader": {"model": "claude-fable-5-2", "effort": "low"}}
        }
    }));
    let before = base.clone();
    let after = with_choice(&base, "glm", "leader", "zai/glm-5.3", "low").unwrap();
    assert_eq!(base, before, "the input object must not change");

    // The retired `codex` key is the one thing the write leaves out (leader decision D2).
    let mut expected = before.clone();
    expected["models"]["glm"] = json!({"leader": {"model": "zai/glm-5.3", "effort": "low"}});
    expected["models"].as_object_mut().unwrap().remove("codex");
    assert_eq!(after, expected);
    assert_eq!(
        resolve("glm", "leader", Some(&after)).unwrap(),
        (choice("zai/glm-5.3", "low"), Source::Saved)
    );

    // Replacing an existing entry keeps its sibling role and every other key.
    let again = with_choice(&after, "glm", "leader", "zai/glm-5.3-flash", "high").unwrap();
    assert_eq!(
        again["models"]["glm"]["leader"],
        json!({"model": "zai/glm-5.3-flash", "effort": "high"})
    );
    assert_eq!(again["models"]["claude"], before["models"]["claude"]);
    assert!(again["models"].get("codex").is_none(), "{again:?}");
    assert_eq!(again["role_access"], json!("prompt"));
    assert_eq!(again["unrelated"], before["unrelated"]);

    // A launcher object with no models key gets the key, and nothing else changes.
    let plain = object(json!({"advisor": "none"}));
    let made = with_choice(&plain, "claude", "agent", "claude-sonnet-5-5", "max").unwrap();
    assert_eq!(made["advisor"], json!("none"));
    assert_eq!(
        made["models"],
        json!({"claude": {"agent": {"model": "claude-sonnet-5-5", "effort": "max"}}})
    );
}

#[test]
fn an_unsupported_choice_is_refused_and_the_input_is_unchanged() {
    let base = object(
        json!({"advisor": "fable", "models": {"codex": {"agent": {"model": "gpt-6.2-sol", "effort": "max"}}}}),
    );
    let before = base.clone();
    let cases = [
        ("nope", "leader", "zai/glm-5.3", "high"),
        ("claude", "nope", "claude-opus-5-5", "high"),
        ("claude", "leader", "zai/glm-5.3", "high"),
        ("claude", "leader", "claude-opus-5-5", "ultra"),
        ("claude", "leader", "claude-opus-5-5", ""),
        ("glm", "agent", "zai/glm-5.3-flash", "xhigh"),
        ("glm", "agent", "zai/glm-5.3-flash", "medium"),
        ("claude", "leader", "opus", "high"),
        ("claude", "leader", "claude-opus-5-5", "ultracode"),
        // Codex is not a tool any more: a choice written for it is refused, whatever it holds.
        ("codex", "agent", "gpt-6.2-sol", "max"),
        ("codex", "leader", "gpt-6.1-sol", "high"),
        ("codex", "agent", "claude-opus-5-5", "high"),
    ];
    for (tool, role, model, effort) in cases {
        let message = with_choice(&base, tool, role, model, effort)
            .expect_err(&format!("{tool} {role} {model} {effort} must be refused"));
        assert!(!message.is_empty());
        assert_eq!(
            base, before,
            "{tool} {role} {model} {effort}: the input changed"
        );
    }
    // A models key that is not an object cannot be merged into.
    let broken = object(json!({"models": SECRET}));
    let message = with_choice(&broken, "claude", "leader", "claude-opus-5-5", "high").unwrap_err();
    assert!(
        message.contains("launcher.models") && !message.contains(SECRET),
        "{message}"
    );
}

#[test]
fn every_default_is_inside_the_support_table() {
    assert!(supported("nope").is_none());
    for tool in TOOLS {
        let rows = supported(tool).unwrap_or_else(|| panic!("{tool}: no table"));
        assert!(!rows.is_empty(), "{tool}");
        for row in rows {
            assert!(!row.efforts.is_empty(), "{tool} {}", row.id);
            for effort in row.efforts {
                assert!(
                    !["ultra", "ultracode"].contains(effort),
                    "{tool} {} lists {effort}",
                    row.id
                );
            }
        }
        for role in ROLES {
            let default =
                default_choice(tool, role).unwrap_or_else(|| panic!("{tool} {role}: no default"));
            let row = rows
                .iter()
                .find(|r| r.id == default.model)
                .unwrap_or_else(|| {
                    panic!(
                        "{tool} {role}: default model {} not in the table",
                        default.model
                    )
                });
            assert!(
                row.efforts.contains(&default.effort.as_str()),
                "{tool} {role}: default effort {} not offered by {}",
                default.effort,
                row.id
            );
        }
    }
}

#[test]
fn unsupported_saved_entries_are_listed_by_key_path_without_quoting_them() {
    let launcher = object(json!({
        "models": {
            "claude": {"leader": {"model": SECRET, "effort": "high"}},
            "codex": {"agent": {"model": "gpt-6.2-sol", "effort": "max"}},
            "glm": {"agent": {"model": "zai/glm-5.3-flash"}, "leader": {"model": "zai/glm-5.3", "effort": SECRET}}
        }
    }));
    let mut listed = unsupported_entries(&launcher).unwrap();
    listed.sort();
    assert_eq!(
        listed,
        [
            "launcher.models.claude.leader",
            "launcher.models.glm.agent",
            "launcher.models.glm.leader"
        ]
    );
    assert!(!format!("{listed:?}").contains(SECRET));
    // No models key and an all-supported models key list nothing.
    assert!(unsupported_entries(&Map::new()).unwrap().is_empty());
    let fine =
        object(json!({"models": {"glm": {"agent": {"model": "zai/glm-5.3", "effort": "max"}}}}));
    assert!(unsupported_entries(&fine).unwrap().is_empty());
    // Structural damage is refused by key path, never listed.
    for broken in [
        json!({"models": SECRET}),
        json!({"models": {SECRET: {}}}),
        json!({"models": {"claude": SECRET}}),
        json!({"models": {"claude": {SECRET: {}}}}),
        json!({"models": {"claude": {"agent": SECRET}}}),
    ] {
        let message = unsupported_entries(&object(broken)).unwrap_err();
        assert!(
            message.contains("launcher.models") && !message.contains(SECRET),
            "{message}"
        );
    }
}

#[test]
fn the_slice_one_audit_cases_are_pinned() {
    // A non-object tool entry cannot be merged into; the refusal names the key path only.
    let base = object(json!({"models": {"claude": SECRET}}));
    let before = base.clone();
    let message = with_choice(&base, "claude", "leader", "claude-opus-5-5", "high").unwrap_err();
    assert!(
        message.contains("launcher.models.claude") && !message.contains(SECRET),
        "{message}"
    );
    assert_eq!(base, before);

    // Replacing an entry drops its extra fields.
    let base = object(json!({
        "models": {"claude": {"agent": {"model": "claude-sonnet-5-5", "effort": "max", "extra": 1}}}
    }));
    let after = with_choice(&base, "claude", "agent", "claude-haiku-5-5", "low").unwrap();
    assert_eq!(
        after["models"]["claude"]["agent"],
        json!({"model": "claude-haiku-5-5", "effort": "low"})
    );

    // The omp tool with a saved value and with a set; glm is a separate entry.
    let saved =
        object(json!({"models": {"omp": {"leader": {"model": "zai/glm-5.3", "effort": "low"}}}}));
    assert_eq!(
        resolve("omp", "leader", Some(&saved)).unwrap(),
        (choice("zai/glm-5.3", "low"), Source::Saved)
    );
    assert_eq!(
        resolve("glm", "leader", Some(&saved)).unwrap().1,
        Source::Default
    );
    let set = with_choice(&saved, "omp", "agent", "zai/glm-5.3-flash", "high").unwrap();
    assert_eq!(
        set["models"]["omp"]["leader"],
        saved["models"]["omp"]["leader"]
    );
    assert_eq!(
        resolve("omp", "agent", Some(&set)).unwrap(),
        (choice("zai/glm-5.3-flash", "high"), Source::Saved)
    );

    // A refused set quotes no caller text.
    let message = with_choice(&Map::new(), "claude", "leader", SECRET, "high").unwrap_err();
    assert!(!message.contains(SECRET), "{message}");
}

#[test]
fn codex_is_not_a_tool_any_more() {
    assert!(!TOOLS.contains(&"codex"), "TOOLS still lists codex");
    assert!(
        supported("codex").is_none(),
        "the table still has a codex row"
    );
    for role in ROLES {
        assert_eq!(default_choice("codex", role), None, "codex {role}");
    }
    // Not even a valid-looking old value saved for it is read as a choice.
    let launcher = object(json!({
        "models": {"codex": {"agent": {"model": "gpt-6.2-sol", "effort": "max"}}}
    }));
    assert_eq!(saved_choice(&launcher, "codex", "agent"), None);
    assert!(resolve("codex", "agent", Some(&launcher)).is_err());
}

#[test]
fn a_saved_codex_value_never_refuses_another_tool_and_is_not_listed() {
    for (shape, codex) in old_codex_shapes() {
        let launcher = object(json!({"models": {"codex": codex.clone()}}));
        for tool in TOOLS {
            for role in ROLES {
                let (got, source) = resolve(tool, role, Some(&launcher))
                    .unwrap_or_else(|e| panic!("{shape}: {tool} {role} refused: {e}"));
                assert_eq!(source, Source::Default, "{shape}: {tool} {role}");
                assert_eq!(
                    Some(got),
                    default_choice(tool, role),
                    "{shape}: {tool} {role}"
                );
            }
        }
        assert_eq!(
            unsupported_entries(&launcher).unwrap(),
            Vec::<String>::new(),
            "{shape}: listed as unsupported"
        );
        // A saved choice of another tool next to it still wins.
        let with_glm = object(json!({"models": {
            "codex": codex,
            "glm": {"agent": {"model": "zai/glm-5.3", "effort": "low"}}
        }}));
        assert_eq!(
            resolve("glm", "agent", Some(&with_glm)).unwrap(),
            (choice("zai/glm-5.3", "low"), Source::Saved),
            "{shape}"
        );
        // A broken entry of a known tool is still refused with a Codex value beside it.
        let broken = object(json!({"models": {
            "codex": json!(SECRET),
            "glm": {"agent": {"model": SECRET, "effort": "low"}}
        }}));
        let message = resolve("claude", "leader", Some(&broken)).unwrap_err();
        assert!(
            message.contains("launcher.models.glm.agent") || message.contains("launcher.models"),
            "{message}"
        );
        assert!(!message.contains(SECRET), "{message}");
    }
}

#[test]
fn a_saved_codex_value_is_left_out_when_another_tool_is_written() {
    for (shape, codex) in old_codex_shapes() {
        let base = object(json!({"advisor": "fable", "models": {"codex": codex.clone()}}));
        let before = base.clone();
        let after = with_choice(&base, "claude", "agent", "claude-haiku-5-5", "low")
            .unwrap_or_else(|e| panic!("{shape}: the write was refused: {e}"));
        assert_eq!(base, before, "{shape}: the input object changed");
        assert!(
            after["models"].get("codex").is_none(),
            "{shape}: the Codex value stayed in the written file"
        );
        assert_eq!(after["advisor"], json!("fable"), "{shape}");
        assert_eq!(
            after["models"]["claude"]["agent"],
            json!({"model": "claude-haiku-5-5", "effort": "low"}),
            "{shape}"
        );
    }
}
