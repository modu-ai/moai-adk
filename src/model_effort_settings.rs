//! The model and effort a launcher passes to each coding tool (SPEC-20261118-model-effort-settings,
//! card C-47 slice 2): the table of what each tool supports, the defaults, and two pure functions
//! on the `launcher` object of the user launcher settings. The choice is saved as
//! `launcher.models.<tool>.<role>` = `{"model": <id>, "effort": <level>}`.
//!
//! The module reads and writes no file. Refusals name the key path and never quote saved text, so
//! a secret pasted into the file cannot reach a message. Model ids and efforts come from the
//! research file of SPEC-20261118-install-model-per-tool (read 2026-11-18); `ultracode` (Claude
//! Code) is not offered.
//!
//! Codex is not a tool any more (SPEC-20261121-codex-launcher-and-models-removed). A
//! `launcher.models.codex` value that an earlier version saved is ignored on every read and kept
//! as it was in every write: it is never validated, listed or resolved.
use serde_json::{json, Map, Value};

/// The tools that take a choice. `glm` is the `moai glm` launcher and shares the OMP table.
pub const TOOLS: [&str; 3] = ["claude", "omp", "glm"];
/// Tool names an earlier version saved that this version no longer supports: their
/// `launcher.models.<name>` value is skipped on read and left in the file.
const IGNORED_TOOLS: [&str; 1] = ["codex"];
/// The two roles: the leader entry (`-l`) and the agent entry (`-a`).
pub const ROLES: [&str; 2] = ["leader", "agent"];

/// One supported model and the efforts it offers.
pub struct ModelRow {
    pub id: &'static str,
    pub efforts: &'static [&'static str],
}

const FIVE_EFFORTS: &[&str] = &["low", "medium", "high", "xhigh", "max"];
const THREE_EFFORTS: &[&str] = &["low", "high", "max"];

const CLAUDE: &[ModelRow] = &[
    ModelRow {
        id: "claude-opus-5-5",
        efforts: FIVE_EFFORTS,
    },
    ModelRow {
        id: "claude-sonnet-5-5",
        efforts: FIVE_EFFORTS,
    },
    ModelRow {
        id: "claude-haiku-5-5",
        efforts: FIVE_EFFORTS,
    },
    ModelRow {
        id: "claude-fable-5-2",
        efforts: FIVE_EFFORTS,
    },
];
const GLM: &[ModelRow] = &[
    ModelRow {
        id: "zai/glm-5.3",
        efforts: THREE_EFFORTS,
    },
    ModelRow {
        id: "zai/glm-5.3-flash",
        efforts: THREE_EFFORTS,
    },
];

/// The models a tool supports; `None` for a name this version does not know.
pub fn supported(tool: &str) -> Option<&'static [ModelRow]> {
    match tool {
        "claude" => Some(CLAUDE),
        "omp" | "glm" => Some(GLM),
        _ => None,
    }
}

/// A model id and an effort level.
#[derive(Clone, Debug, PartialEq, Eq)]
pub struct Choice {
    pub model: String,
    pub effort: String,
}

/// Where a resolved choice came from.
#[derive(Clone, Copy, Debug, PartialEq, Eq)]
pub enum Source {
    Saved,
    Default,
}

/// The default of a tool and role (decision D-4 of SPEC-20261118-model-effort-settings; the
/// claude-leader effort changed by the user decision of 2026-11-19, from the research note of
/// SPEC-20261120-update-three-tool-models).
pub fn default_choice(tool: &str, role: &str) -> Option<Choice> {
    let (model, effort) = match (tool, role) {
        ("claude", "leader") => ("claude-opus-5-5", "high"),
        ("claude", "agent") => ("claude-sonnet-5-5", "high"),
        ("omp" | "glm", "leader") => ("zai/glm-5.3", "high"),
        ("omp" | "glm", "agent") => ("zai/glm-5.3-flash", "max"),
        _ => return None,
    };
    Some(Choice {
        model: model.to_string(),
        effort: effort.to_string(),
    })
}

fn offers(tool: &str, model: &str, effort: &str) -> bool {
    supported(tool).is_some_and(|rows| {
        rows.iter()
            .any(|row| row.id == model && row.efforts.contains(&effort))
    })
}

/// The `launcher.models` object, `None` when absent.
fn models_of(launcher: &Map<String, Value>) -> Result<Option<&Map<String, Value>>, String> {
    match launcher.get("models") {
        None => Ok(None),
        Some(Value::Object(models)) => Ok(Some(models)),
        Some(_) => Err("launcher.models must be an object".into()),
    }
}

/// The structure of `launcher.models`: known tool names, known role names, and objects at the
/// tool and role level. Names are printed only after they matched a known name.
fn check_structure(models: &Map<String, Value>) -> Result<(), String> {
    for (tool, per_tool) in models {
        if IGNORED_TOOLS.contains(&tool.as_str()) {
            continue;
        }
        if !TOOLS.contains(&tool.as_str()) {
            return Err("launcher.models names a tool this version does not know".into());
        }
        let Value::Object(roles) = per_tool else {
            return Err(format!("launcher.models.{tool} must be an object"));
        };
        for (role, entry) in roles {
            if !ROLES.contains(&role.as_str()) {
                return Err(format!(
                    "launcher.models.{tool} names a role this version does not know"
                ));
            }
            if !entry.is_object() {
                return Err(format!("launcher.models.{tool}.{role} must be an object"));
            }
        }
    }
    Ok(())
}

/// One saved entry (its tool and role names are known): exactly `model` and `effort` as text, a
/// model of the tool's table, an effort that model offers.
fn check_entry(tool: &str, role: &str, entry: &Value) -> Result<(), String> {
    let Value::Object(fields) = entry else {
        return Err(format!("launcher.models.{tool}.{role} must be an object"));
    };
    if fields.len() != 2 {
        return Err(format!(
            "launcher.models.{tool}.{role} must hold exactly model and effort"
        ));
    }
    let (Some(Value::String(model)), Some(Value::String(effort))) =
        (fields.get("model"), fields.get("effort"))
    else {
        return Err(format!(
            "launcher.models.{tool}.{role} must hold model and effort as text"
        ));
    };
    let rows = supported(tool).unwrap_or(&[]);
    if !rows.iter().any(|row| row.id == model) {
        return Err(format!(
            "launcher.models.{tool}.{role}.model is not a supported model of {tool}"
        ));
    }
    if !offers(tool, model, effort) {
        return Err(format!(
            "launcher.models.{tool}.{role}.effort is not an effort that model offers"
        ));
    }
    Ok(())
}

/// Checks every saved entry, not only the one asked for, so a broken file is never skipped
/// silently.
fn validate(models: &Map<String, Value>) -> Result<(), String> {
    check_structure(models)?;
    for (tool, per_tool) in models {
        if IGNORED_TOOLS.contains(&tool.as_str()) {
            continue;
        }
        if let Value::Object(roles) = per_tool {
            for (role, entry) in roles {
                check_entry(tool, role, entry)?;
            }
        }
    }
    Ok(())
}

/// The key paths (`launcher.models.<tool>.<role>`) of the saved entries the support table does
/// not accept: the launcher refuses such a file until they are replaced. Structural damage (an
/// unknown tool or role name, a level that is not an object, `models` not an object) is refused
/// instead of listed, because it cannot be named without quoting saved text.
pub fn unsupported_entries(launcher: &Map<String, Value>) -> Result<Vec<String>, String> {
    let Some(models) = models_of(launcher)? else {
        return Ok(Vec::new());
    };
    check_structure(models)?;
    let mut listed = Vec::new();
    for (tool, per_tool) in models {
        if IGNORED_TOOLS.contains(&tool.as_str()) {
            continue;
        }
        if let Value::Object(roles) = per_tool {
            for (role, entry) in roles {
                if check_entry(tool, role, entry).is_err() {
                    listed.push(format!("launcher.models.{tool}.{role}"));
                }
            }
        }
    }
    Ok(listed)
}

/// The saved choice of a tool and role when it is there and supported, else `None`. It does not
/// look at any other entry (see [`resolve`] and [`unsupported_entries`]).
pub fn saved_choice(launcher: &Map<String, Value>, tool: &str, role: &str) -> Option<Choice> {
    let entry = launcher.get("models")?.get(tool)?.get(role)?;
    check_entry(tool, role, entry).ok()?;
    Some(Choice {
        model: entry.get("model")?.as_str()?.to_string(),
        effort: entry.get("effort")?.as_str()?.to_string(),
    })
}

/// The model and effort of a tool and role: the saved choice when the launcher object holds one
/// the table supports, else the default. A saved entry that is not supported, anywhere under
/// `launcher.models`, is refused.
pub fn resolve(
    tool: &str,
    role: &str,
    launcher: Option<&Map<String, Value>>,
) -> Result<(Choice, Source), String> {
    let default = default_choice(tool, role)
        .ok_or_else(|| "model settings: unknown tool or role".to_string())?;
    let Some(models) = launcher.map(models_of).transpose()?.flatten() else {
        return Ok((default, Source::Default));
    };
    validate(models)?;
    let saved = models
        .get(tool)
        .and_then(|per_tool| per_tool.get(role))
        .and_then(|entry| {
            Some(Choice {
                model: entry.get("model")?.as_str()?.to_string(),
                effort: entry.get("effort")?.as_str()?.to_string(),
            })
        });
    Ok(match saved {
        Some(choice) => (choice, Source::Saved),
        None => (default, Source::Default),
    })
}

/// A copy of the launcher object with the choice of a tool and role set; every other key, entry
/// and sibling is kept and the input is not changed. Refused when the tool, role, model or
/// effort is not supported or when `launcher.models` is not an object.
pub fn with_choice(
    launcher: &Map<String, Value>,
    tool: &str,
    role: &str,
    model: &str,
    effort: &str,
) -> Result<Map<String, Value>, String> {
    if default_choice(tool, role).is_none() {
        return Err("model settings: unknown tool or role".into());
    }
    if !offers(tool, model, effort) {
        return Err(format!(
            "model settings: that model and effort are not supported for {tool}"
        ));
    }
    let mut models = models_of(launcher)?.cloned().unwrap_or_default();
    let per_tool = models
        .entry(tool.to_string())
        .or_insert_with(|| Value::Object(Map::new()));
    let Value::Object(roles) = per_tool else {
        return Err(format!("launcher.models.{tool} must be an object"));
    };
    roles.insert(role.to_string(), json!({"model": model, "effort": effort}));
    let mut out = launcher.clone();
    out.insert("models".to_string(), Value::Object(models));
    Ok(out)
}
