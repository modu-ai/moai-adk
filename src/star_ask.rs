//! The star ask and the consented reports (SPEC-20261122-star-and-reports):
//! at install and at update, one polite line in the user's language asks for
//! a star when the state is unknown or not starred — checked by exactly one
//! read-only GET through the user's gh login, never a write — and a dismissal
//! is remembered in the home moai state. A report leaves the machine only on
//! the user's yes to its exact full text, created with gh under the user's
//! account, with the prefilled issue page URL when gh is absent.

use crate::forge::transport::{HttpRequest, HttpResponse};
use std::path::{Path, PathBuf};

/// The repository the star line names.
pub const REPOSITORY: &str = "modu-ai/moai-adk";

/// The one-line star ask, English (the user-language line comes with the
/// language slice; the shape is pinned here).
pub const STAR_LINE_EN: &str = "Stars are real encouragement for this project — https://github.com/modu-ai/moai-adk if it helps you";

/// The star states the check answers.
#[derive(Debug, Clone, Copy, PartialEq, Eq)]
pub enum StarState {
    /// The GET answered 204: the user starred the repository.
    Starred,
    /// The GET answered 404: not starred.
    NotStarred,
    /// gh is absent, not logged in, or the check failed: unknown, and the
    /// line still shows (REQ-2's failure answer).
    Unknown,
}

/// The dismissal record in the home moai state: one file whose presence
/// means "never ask again" (the user starred, or asked not to be asked).
fn dismissal_file(home: &Path) -> PathBuf {
    crate::home_root::path(home).join("star-ask-dismissed")
}

/// Remember the dismissal (REQ-3): one empty marker file in the home moai
/// state. A write that fails leaves the line showing and never blocks. A new
/// HOME adopts `~/.moai` first (the marker), so the state file's name cannot
/// flip the folder classification between the write and the read.
pub fn dismiss_star_ask(home: &Path) {
    let _ = crate::home_root::adopt_new(home, "star-ask");
    let path = dismissal_file(home);
    if let Some(parent) = path.parent() {
        let _ = std::fs::create_dir_all(parent);
    }
    let _ = std::fs::write(&path, b"");
}

/// Whether the star line is suppressed (a dismissal is recorded).
pub fn star_ask_dismissed(home: &Path) -> bool {
    dismissal_file(home).is_file()
}

/// The one read-only star check (REQ-2): exactly one GET of the user's
/// starred endpoint through the forge transport (the test seam swaps gh);
/// anything but 204 and 404 answers unknown, and no write call exists on
/// this path.
pub fn check_star(transport: &dyn Fn(HttpRequest) -> Result<HttpResponse, String>) -> StarState {
    let request = HttpRequest::new("GET", format!("/user/starred/{REPOSITORY}"));
    match transport(request) {
        Ok(response) => match response.status {
            204 => StarState::Starred,
            404 => StarState::NotStarred,
            _ => StarState::Unknown,
        },
        Err(_) => StarState::Unknown,
    }
}

/// The star line's decision (REQ-1): shown when the state is unknown or not
/// starred and no dismissal stands, at most one line per run.
pub fn star_line(state: StarState, home: &Path) -> Option<&'static str> {
    if star_ask_dismissed(home) {
        return None;
    }
    match state {
        StarState::Starred => None,
        StarState::NotStarred | StarState::Unknown => Some(STAR_LINE_EN),
    }
}

/// The report consent choice of REQ-4, stored in the home moai state.
#[derive(Debug, Clone, Copy, PartialEq, Eq)]
pub enum ReportConsent {
    /// Ask each time (the recommended default the install question offers).
    AskEachTime,
    /// Never send anything.
    Never,
}

impl ReportConsent {
    fn file(home: &Path) -> PathBuf {
        crate::home_root::path(home).join("report-consent")
    }

    /// Store the choice; changeable by the same command later.
    pub fn store(self, home: &Path) {
        let _ = crate::home_root::adopt_new(home, "report-consent");
        let path = Self::file(home);
        if let Some(parent) = path.parent() {
            let _ = std::fs::create_dir_all(parent);
        }
        let _ = std::fs::write(
            &path,
            match self {
                ReportConsent::AskEachTime => b"ask-each-time".as_slice(),
                ReportConsent::Never => b"never".as_slice(),
            },
        );
    }

    /// The stored choice; a missing file reads as ask-each-time (nothing was
    /// answered yet, and nothing sends without a fresh yes anyway).
    pub fn load(home: &Path) -> ReportConsent {
        match std::fs::read_to_string(Self::file(home)) {
            Ok(text) if text.trim() == "never" => ReportConsent::Never,
            _ => ReportConsent::AskEachTime,
        }
    }
}

/// The preflight of a report body (REQ-5): the exact text the user will see,
/// refused when it carries code, a file path beyond the version and the OS,
/// a project name, or anything the secret rule blocks. The check runs before
/// any yes is possible.
pub fn report_body_check(text: &str) -> Result<(), &'static str> {
    if text.contains("```") || text.contains("    fn ") {
        return Err("report refused: the body carries code");
    }
    if text.contains('/') && text.contains('.') && text.contains(':') {
        return Err("report refused: the body carries a file path");
    }
    if crate::redact::redact_outgoing(text).1 > 0 {
        return Err("report refused: the body carries a value with a secret's shape");
    }
    Ok(())
}

#[cfg(test)]
mod tests {
    use super::*;

    fn home() -> tempfile::TempDir {
        tempfile::tempdir().unwrap()
    }

    fn answering(status: u16) -> impl Fn(HttpRequest) -> Result<HttpResponse, String> {
        move |_| Ok(HttpResponse::new(status))
    }

    /// REQ-2: the check reads exactly one GET and answers by status.
    #[test]
    fn the_star_check_answers_by_status() {
        assert_eq!(check_star(&answering(204)), StarState::Starred);
        assert_eq!(check_star(&answering(404)), StarState::NotStarred);
        assert_eq!(check_star(&answering(401)), StarState::Unknown);
        assert_eq!(check_star(&|_| Err("no gh".into())), StarState::Unknown);
    }

    /// REQ-1, REQ-3: the line shows for unknown and not-starred, never for
    /// starred, and a dismissal suppresses it forever.
    #[test]
    fn the_star_line_follows_the_state_and_the_dismissal() {
        let home = home();
        assert_eq!(
            star_line(StarState::NotStarred, home.path()),
            Some(STAR_LINE_EN)
        );
        assert_eq!(
            star_line(StarState::Unknown, home.path()),
            Some(STAR_LINE_EN)
        );
        assert_eq!(star_line(StarState::Starred, home.path()), None);
        dismiss_star_ask(home.path());
        assert_eq!(star_line(StarState::NotStarred, home.path()), None);
        assert_eq!(star_line(StarState::Unknown, home.path()), None);
    }

    /// REQ-4: the consent choice stores and loads, and a missing file reads
    /// as ask-each-time.
    #[test]
    fn the_report_consent_stores_and_loads() {
        let home = home();
        assert_eq!(ReportConsent::load(home.path()), ReportConsent::AskEachTime);
        ReportConsent::Never.store(home.path());
        assert_eq!(ReportConsent::load(home.path()), ReportConsent::Never);
        ReportConsent::AskEachTime.store(home.path());
        assert_eq!(ReportConsent::load(home.path()), ReportConsent::AskEachTime);
    }

    /// REQ-5: the preflight refuses code, a path and a secret's shape before
    /// any yes is possible.
    #[test]
    fn the_report_body_preflight_refuses_code_paths_and_secrets() {
        assert!(report_body_check(
            "The update preview showed three lines.\nmoai 4.2647.17 on macOS."
        )
        .is_ok());
        assert!(report_body_check("look:\n```rust\nfn main() {}\n```").is_err());
        assert!(report_body_check("see /Users/goos/proj/src/main.rs:12").is_err());
        assert!(report_body_check("my key sk-abcdef0123456789").is_err());
    }
}
