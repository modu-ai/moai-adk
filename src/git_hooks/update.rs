//! The git-hook part of `moai update` (SPEC-20261121-go-path-removal): the same shape as the
//! name-switch layer — a preview JSON part with its own `--confirm-git-hooks <token>`, an apply
//! that retires and installs, and a revert record. The repository is the project the update
//! runs in. The wiring lines in `src/update.rs` land through the pair with the card that holds
//! that file; until then the standalone `git-hooks-*` verbs reach the feature directly.

use super::{apply as feature_apply, plan as feature_plan, revert as feature_revert};
use serde_json::{json, Value};
use std::path::Path;

type Result<T> = std::result::Result<T, String>;

const NOT_REVIEWED: &str =
    "the git-hooks token differs from the preview's; nothing applied, preview again";

/// The git-hook part of one `moai update` preview.
#[derive(Debug)]
pub struct Layer {
    preview: Option<Value>,
    token: Option<String>,
}

/// The git-hook preview for the update flow: the repository is the project the update runs in
/// (`cwd`); `home` keeps the layer call's shape beside the other layers and is unused today.
/// A folder without `.git` carries no part — `moai update` runs outside repositories too — and
/// writes nothing; a repository with nothing to change carries no part either, like the switch
/// layer.
pub fn plan(_home: &Path, cwd: &Path) -> Result<Layer> {
    let none = Layer {
        preview: None,
        token: None,
    };
    if std::fs::symlink_metadata(cwd.join(".git")).is_err() {
        return Ok(none);
    }
    let preview = feature_plan(cwd)?;
    let writes = preview["writes"].as_bool().unwrap_or(false);
    Ok(if writes {
        Layer {
            token: preview["confirm_token"].as_str().map(String::from),
            preview: Some(preview),
        }
    } else {
        Layer {
            preview: None,
            token: None,
        }
    })
}

impl Layer {
    /// The part's preview, when there is one.
    pub fn to_json(&self) -> Option<&Value> {
        self.preview.as_ref()
    }

    /// The token `--confirm-git-hooks` takes, when confirming would write something.
    pub fn token(&self) -> Option<&str> {
        self.token.as_deref()
    }

    /// Whether confirming would write files.
    pub fn writes(&self) -> bool {
        self.token.is_some()
    }

    /// The terminal question's lines for this part; empty when there is no part.
    pub fn summary(&self) -> String {
        match &self.preview {
            Some(p) => {
                let names = |key: &str| -> Vec<String> {
                    p[key]
                        .as_array()
                        .map(|rows| {
                            rows.iter()
                                .filter_map(|r| r["name"].as_str().map(String::from))
                                .collect()
                        })
                        .unwrap_or_default()
                };
                let install: Vec<String> = p["install"]
                    .as_array()
                    .map(|rows| {
                        rows.iter()
                            .filter_map(|r| r.as_str().map(String::from))
                            .collect()
                    })
                    .unwrap_or_default();
                format!(
                    "git hooks\n  retire {} and install {}, keep {}\n",
                    names("go_hooks").join(", "),
                    install.join(", "),
                    names("kept").join(", ")
                )
            }
            None => String::new(),
        }
    }

    /// The parts the preview carries for the JSON report.
    pub fn parts(&self) -> Value {
        match &self.preview {
            Some(p) => p.clone(),
            None => json!(null),
        }
    }
}

/// Refuse a `--confirm-git-hooks` token that is not the preview's, before anything is written.
pub fn check_token<'a>(layer: &Layer, given: Option<&'a str>) -> Result<Option<&'a str>> {
    let Some(t) = given else {
        return Ok(None);
    };
    if layer.token() != Some(t) {
        return Err(NOT_REVIEWED.into());
    }
    Ok(Some(t))
}

/// Apply the previewed change to the project at `cwd`; `now` (`YYYY-MM-DDTHH:MM:SSZ`) names the
/// retire folder.
pub fn apply(cwd: &Path, token: &str, now: &str) -> Result<Value> {
    feature_apply(cwd, token, now)
}

/// Put the retired hooks of the project at `cwd` back.
pub fn revert(cwd: &Path, stamp: &str) -> Result<Value> {
    feature_revert(cwd, stamp)
}
