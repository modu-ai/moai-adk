//! The profile keys of the CI design (SPEC-20261122-project-ci-design, slice A): `ci`, the
//! saved answers of the plain questions, and `checks`, the ordered commands the project
//! declares.
//!
//! This module only states and checks the two declarations. It runs no command and writes
//! nothing; the profile they live in goes through `project-plan` and `project-apply` like any
//! other change. The declared list is the one source of the generated workflow's run steps and
//! of the project's local check, so it keeps its declared order and each command is one line a
//! person confirmed. The limits are stated values chosen for the product, not measured ones.

use serde::{Deserialize, Serialize};

type Result<T> = std::result::Result<T, String>;

/// What a project is (`ci_kind`), in the order the question lists them.
pub const KINDS: &[&str] = &["web", "api", "cli", "library", "mobile"];
/// The operating systems a user may name (`ci_users`), in the one order a list keeps.
pub const SYSTEMS: &[&str] = &["linux", "macos", "windows"];
/// Whether the repository is public or private (`ci_money`).
pub const VISIBILITIES: &[&str] = &["public", "private"];
/// Whether failed checks block a merge or only report (`ci_strict`).
pub const STRICTNESS: &[&str] = &["report", "block"];
/// Largest monthly budget in US dollars the answers accept.
pub const BUDGET_MAX_USD: u32 = 10_000;
/// Longest deploy place or check id, in bytes.
pub const NAME_BYTES: usize = 40;
/// Most declared checks.
pub const CHECK_LIMIT: usize = 16;
/// Longest declared command, in bytes.
pub const COMMAND_BYTES: usize = 200;

/// Where the project deploys (`ci_deploy`): a named place only when it does.
#[derive(Clone, Debug, Serialize, Deserialize, PartialEq, Eq)]
#[serde(deny_unknown_fields)]
pub struct CiDeploy {
    pub enabled: bool,
    #[serde(default, skip_serializing_if = "Option::is_none")]
    pub target: Option<String>,
}

/// The saved answers of the plain CI questions.
#[derive(Clone, Debug, Serialize, Deserialize, PartialEq, Eq)]
#[serde(deny_unknown_fields)]
pub struct CiAnswers {
    pub kind: String,
    /// Sorted, with `linux` always present: Linux runs the tests, the others only build.
    pub systems: Vec<String>,
    pub deploy: CiDeploy,
    pub visibility: String,
    /// Private repositories only: the monthly Actions budget in whole US dollars.
    #[serde(default, skip_serializing_if = "Option::is_none")]
    pub monthly_budget_usd: Option<u32>,
    /// Present, and `true`, exactly when `macos` is named: the user saw its per-minute price.
    #[serde(default, skip_serializing_if = "Option::is_none")]
    pub macos_price_ack: Option<bool>,
    /// The user picked the `github-actions` update file in the preview.
    #[serde(default, skip_serializing_if = "Option::is_none")]
    pub dependabot: Option<bool>,
    pub strict: String,
}

/// One declared check: a short id and the one-line command a person confirmed.
#[derive(Clone, Debug, Serialize, Deserialize, PartialEq, Eq)]
#[serde(deny_unknown_fields)]
pub struct CheckEntry {
    pub id: String,
    pub command: String,
}

const REFUSED_ANSWERS: &str = "invalid ci answers; no values were logged";
const REFUSED_CHECKS: &str = "invalid declared checks; no values were logged";

/// Lowercase letters, digits, `-` and `_`, starting with a letter or digit, 1..=[`NAME_BYTES`].
fn name_ok(name: &str) -> bool {
    !name.is_empty()
        && name.len() <= NAME_BYTES
        && name
            .bytes()
            .next()
            .is_some_and(|b| b.is_ascii_lowercase() || b.is_ascii_digit())
        && name
            .bytes()
            .all(|b| b.is_ascii_lowercase() || b.is_ascii_digit() || b == b'-' || b == b'_')
}

/// Accepts a `ci` value as the interview writes it.
pub fn validate_answers(a: &CiAnswers) -> Result<()> {
    let macos = a.systems.iter().any(|s| s == "macos");
    let systems_ok = a.systems.first().is_some_and(|s| s == "linux")
        && a.systems.iter().all(|s| SYSTEMS.contains(&s.as_str()))
        && a.systems.windows(2).all(|w| w[0] < w[1]);
    let deploy_ok = match (&a.deploy.enabled, &a.deploy.target) {
        (true, Some(target)) => name_ok(target),
        (false, None) => true,
        _ => false,
    };
    let budget_ok = match a.monthly_budget_usd {
        None => true,
        Some(usd) => a.visibility == "private" && usd <= BUDGET_MAX_USD,
    };
    let ack_ok = a.macos_price_ack == macos.then_some(true);
    if KINDS.contains(&a.kind.as_str())
        && systems_ok
        && deploy_ok
        && VISIBILITIES.contains(&a.visibility.as_str())
        && budget_ok
        && ack_ok
        && STRICTNESS.contains(&a.strict.as_str())
    {
        Ok(())
    } else {
        Err(REFUSED_ANSWERS.into())
    }
}

/// Whether `command` is one plain line that is safe to put in a workflow run step: no control
/// character (so no second line), no leading or trailing space, no expression opener `${{`
/// (event data must not reach a shell through it), no secret-shaped text, at most
/// [`COMMAND_BYTES`].
pub fn command_ok(command: &str) -> bool {
    !command.is_empty()
        && command.len() <= COMMAND_BYTES
        && command == command.trim()
        && !command.chars().any(char::is_control)
        && !command.contains("${{")
        && !crate::secret_shaped(command)
}

/// Accepts a `checks` value: 1..=[`CHECK_LIMIT`] entries with distinct ids, each command one
/// plain line. The order is the declared order and is never sorted.
pub fn validate_checks(checks: &[CheckEntry]) -> Result<()> {
    let ids_distinct = checks
        .iter()
        .enumerate()
        .all(|(i, c)| checks[..i].iter().all(|earlier| earlier.id != c.id));
    if !checks.is_empty()
        && checks.len() <= CHECK_LIMIT
        && ids_distinct
        && checks
            .iter()
            .all(|c| name_ok(&c.id) && command_ok(&c.command))
    {
        Ok(())
    } else {
        Err(REFUSED_CHECKS.into())
    }
}

#[cfg(test)]
mod tests {
    use super::*;

    fn answers() -> CiAnswers {
        CiAnswers {
            kind: "cli".into(),
            systems: vec!["linux".into()],
            deploy: CiDeploy {
                enabled: false,
                target: None,
            },
            visibility: "public".into(),
            monthly_budget_usd: None,
            macos_price_ack: None,
            dependabot: None,
            strict: "report".into(),
        }
    }

    fn check(id: &str, command: &str) -> CheckEntry {
        CheckEntry {
            id: id.into(),
            command: command.into(),
        }
    }

    /// The exit codes of the answer rules at their edges, one value each side of the line.
    #[test]
    fn answers_accept_the_edge_and_refuse_one_past_it() {
        assert!(validate_answers(&answers()).is_ok());
        let mut a = answers();
        a.visibility = "private".into();
        a.monthly_budget_usd = Some(BUDGET_MAX_USD);
        assert!(validate_answers(&a).is_ok());
        a.monthly_budget_usd = Some(BUDGET_MAX_USD + 1);
        assert!(validate_answers(&a).is_err());
        a.monthly_budget_usd = Some(0);
        assert!(validate_answers(&a).is_ok());
        a.visibility = "public".into();
        assert!(
            validate_answers(&a).is_err(),
            "a budget needs a private repository"
        );

        let mut d = answers();
        d.deploy = CiDeploy {
            enabled: true,
            target: Some("p".repeat(NAME_BYTES)),
        };
        assert!(validate_answers(&d).is_ok());
        d.deploy.target = Some("p".repeat(NAME_BYTES + 1));
        assert!(validate_answers(&d).is_err());
        d.deploy.target = None;
        assert!(validate_answers(&d).is_err());
        d.deploy = CiDeploy {
            enabled: false,
            target: Some("production".into()),
        };
        assert!(validate_answers(&d).is_err());
    }

    #[test]
    fn systems_need_linux_first_sorted_and_the_macos_yes() {
        let with = |systems: &[&str], ack: Option<bool>| {
            let mut a = answers();
            a.systems = systems.iter().map(|s| s.to_string()).collect();
            a.macos_price_ack = ack;
            validate_answers(&a).is_ok()
        };
        assert!(with(&["linux"], None));
        assert!(with(&["linux", "windows"], None));
        assert!(with(&["linux", "macos", "windows"], Some(true)));
        assert!(!with(&["linux", "macos"], None));
        assert!(!with(&["linux", "macos"], Some(false)));
        assert!(
            !with(&["linux"], Some(true)),
            "a yes with no macos names nothing"
        );
        assert!(!with(&["windows"], None));
        assert!(!with(&["windows", "linux"], None));
        assert!(!with(&["linux", "linux"], None));
        assert!(!with(&["linux", "beos"], None));
        assert!(!with(&[], None));
    }

    #[test]
    fn declared_checks_keep_their_order_and_refuse_a_command_that_is_not_one_plain_line() {
        let ok = vec![check("lint", "cargo clippy"), check("build", "cargo build")];
        assert!(validate_checks(&ok).is_ok());
        for bad in [
            "",
            " x",
            "x ",
            "a\nb",
            "a\tb",
            "a\rb",
            "echo ${{ github.event.issue.title }}",
            "echo ${{x}}",
            "run --key sk-abc",
            "run token=abc",
        ] {
            assert!(validate_checks(&[check("t", bad)]).is_err(), "{bad:?}");
        }
        // A `$(…)` substitution and a `{{` that is not an expression opener are plain text.
        assert!(command_ok("test -z \"$(gofmt -l .)\""));
        assert!(command_ok("echo {{x}}"));
        assert!(command_ok(&"x".repeat(COMMAND_BYTES)));
        assert!(!command_ok(&"x".repeat(COMMAND_BYTES + 1)));
    }

    #[test]
    fn declared_check_ids_are_distinct_short_names_and_the_list_is_bounded() {
        assert!(validate_checks(&[]).is_err());
        assert!(validate_checks(&[check("a", "x"), check("a", "y")]).is_err());
        for bad in ["", "A", "a b", "a/b", "-a", &"i".repeat(NAME_BYTES + 1)] {
            assert!(validate_checks(&[check(bad, "x")]).is_err(), "{bad:?}");
        }
        assert!(validate_checks(&[check(&"i".repeat(NAME_BYTES), "x")]).is_ok());
        let many = |n: usize| -> Vec<CheckEntry> {
            (0..n).map(|i| check(&format!("c{i}"), "true")).collect()
        };
        assert!(validate_checks(&many(CHECK_LIMIT)).is_ok());
        assert!(validate_checks(&many(CHECK_LIMIT + 1)).is_err());
    }
}
