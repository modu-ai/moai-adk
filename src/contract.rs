//! Output contract of the internal CLI (Issue #1990, 01-C; user decision 2026-11-08).
//!
//! Every internal operation prints one envelope on stdout:
//! success `{protocol_version, operation, ok: true, result}`,
//! failure `{protocol_version, operation, ok: false, error{kind, message, retryable, next_actions?}}`.
//! stderr carries one human line. Exit codes stay 0 (ok), 1 (failure) and 2 (usage); the
//! failure is told apart by `kind`, from the 00-D vocabulary. The developer launchers
//! (`-l`, `-a`, `cc`, `glm`, `codex`, `session`) are outside this contract.

use serde_json::{json, Value};

pub const PROTOCOL_VERSION: u32 = 1;

/// Longest permission reference, in bytes, that a write operation accepts.
pub const PERMISSION_REFERENCE_MAX_BYTES: usize = 128;

/// Whether a caller-supplied permission reference is usable (WF-OBS #24, Issue #2401).
///
/// A reference is refused when it is empty, `null` or `placeholder` after trimming, longer than
/// [`PERMISSION_REFERENCE_MAX_BYTES`] bytes, or carries a control character. Caller
/// authorization remains host-owned; this check is not an authorization verifier.
pub fn permission_reference_ok(reference: &str) -> bool {
    !(["", "null", "placeholder"].contains(&reference.trim())
        || reference.len() > PERMISSION_REFERENCE_MAX_BYTES
        || reference.chars().any(char::is_control))
}

#[derive(Clone, Copy, Debug, PartialEq, Eq)]
pub enum Kind {
    Usage,
    InvalidInput,
    InputTooLarge,
    UnsupportedProtocolVersion,
    StalePlan,
    PermissionReferenceMissing,
    NotFound,
    Conflict,
    Partial,
    Io,
    Internal,
    /// git ran and failed for a reason other than "not a repository" (Issue #2033).
    GitFailed,
    /// 07-A (#2008): authentication or permission refused, or a rate limit outlived the retry
    /// budget. An addition to the 00-D vocabulary as #1990 HOLD-C2 / REQ-01C-02 (r27) allows
    /// when recorded in docs.
    Blocked,
    /// 07-A (#2008): the host CLI (`gh`) was missing, gave no HTTP response, or an unexpected
    /// status.
    Transport,
    /// #2345: another process holds an advisory lock this command needs — so far the claim
    /// journal lock (REQ-06A-05: the tool never waits). Resending the unchanged command can
    /// succeed once the holder drops it. Raised with an explicit kind at the lock-wait site,
    /// never left to [`classify`]'s message guessing.
    Unavailable,
}

impl Kind {
    pub fn as_str(self) -> &'static str {
        match self {
            Kind::Usage => "usage",
            Kind::InvalidInput => "invalid_input",
            Kind::InputTooLarge => "input_too_large",
            Kind::UnsupportedProtocolVersion => "unsupported_protocol_version",
            Kind::StalePlan => "stale_plan",
            Kind::PermissionReferenceMissing => "permission_reference_missing",
            Kind::NotFound => "not_found",
            Kind::Conflict => "conflict",
            Kind::Partial => "partial",
            Kind::Io => "io",
            Kind::Internal => "internal",
            Kind::GitFailed => "git_failed",
            Kind::Blocked => "blocked",
            Kind::Transport => "transport",
            Kind::Unavailable => "unavailable",
        }
    }
    /// Only an I/O or transport failure, or a lock another process holds (#2345), may succeed
    /// when the same request is sent again unchanged. `blocked` needs a change first
    /// (credentials, permission, or waiting until the rate-limit reset named in the message).
    pub fn retryable(self) -> bool {
        matches!(self, Kind::Io | Kind::Transport | Kind::Unavailable)
    }
}

/// An operation error that carries its 00-D kind from the raise site (Issue #2345). The
/// journal lock-wait path raises this directly; every other error stays a plain `String`
/// and keeps the [`classify`] fallback at the envelope boundary.
#[derive(Clone, Debug)]
pub struct OpError {
    pub kind: Kind,
    pub message: String,
}

impl OpError {
    pub fn new(kind: Kind, message: impl Into<String>) -> Self {
        Self {
            kind,
            message: message.into(),
        }
    }
}

impl From<String> for OpError {
    fn from(message: String) -> Self {
        Self {
            kind: classify(&message),
            message,
        }
    }
}

impl From<&str> for OpError {
    fn from(message: &str) -> Self {
        Self::from(message.to_string())
    }
}

impl std::fmt::Display for OpError {
    /// stderr carries one human line: the message, exactly as the envelope does.
    fn fmt(&self, f: &mut std::fmt::Formatter<'_>) -> std::fmt::Result {
        f.write_str(&self.message)
    }
}

// @MX:NOTE: [AUTO] Error messages stay the source of truth; this ordered table maps them to a
// kind. First match wins, so specific phrases precede the generic words at the end.
// The wrapping messages (`partial apply: ...; failed at X: <inner>`, `conflict in X: <detail>`)
// come first because they quote an inner message that would otherwise match.
const RULES: &[(&str, Kind)] = &[
    ("partial apply", Kind::Partial),
    ("partial creation possible", Kind::Partial),
    ("conflict in ", Kind::Conflict),
    // claim-verify (rc.2 A9 #2255): ahead of the generic words its messages quote.
    ("claim verify: not owner", Kind::Conflict),
    ("claim verify: closed", Kind::Conflict),
    ("claim verify: no claim", Kind::Conflict),
    // audit-claim (rc.2 #2403 item 1): ahead of the generic words its messages quote.
    ("audit claim refused: ", Kind::InvalidInput),
    ("audit claim conflict: ", Kind::Conflict),
    ("audit claim blocked: ", Kind::Blocked),
    // audit-delta (#2403): explicit prefixes precede generic Git/input phrases.
    ("audit delta refused: ", Kind::InvalidInput),
    ("audit delta blocked: ", Kind::Blocked),
    ("audit delta git failed: ", Kind::GitFailed),
    ("audit delta I/O failed: ", Kind::Io),
    // decision-append (rc.2 #2403): ahead of the generic words its messages may contain.
    ("decision append blocked: ", Kind::Blocked),
    ("decision append refused: ", Kind::InvalidInput),
    ("claim verify: stale", Kind::Blocked),
    ("claim verify: uncertain", Kind::Blocked),
    ("claim verify: trailer missing", Kind::InvalidInput),
    ("claim verify: trailer check unavailable", Kind::Blocked),
    // identity's `git failed: <git summary>`; ahead of the generic words the summary may hold.
    // The bare `git failed` of baseline (no colon) stays `io` below.
    ("git failed: ", Kind::GitFailed),
    // forge-read (07-A #2008): its messages carry these prefixes; they precede the generic
    // words because a transport message may quote `unavailable` or `refused`.
    ("forge blocked: ", Kind::Blocked),
    ("lane handoff blocked: ", Kind::Blocked),
    ("gh transport: ", Kind::Transport),
    ("forge resource not found", Kind::NotFound),
    (
        "unsupported request schema_version",
        Kind::UnsupportedProtocolVersion,
    ),
    // plan-admit (05-B #2002): the request's protocol_version is not 1.
    (
        "unsupported protocol_version",
        Kind::UnsupportedProtocolVersion,
    ),
    ("input exceeds 64 KiB", Kind::InputTooLarge),
    (
        "permission reference required",
        Kind::PermissionReferenceMissing,
    ),
    ("stale or altered plan", Kind::StalePlan),
    ("changed since plan", Kind::StalePlan),
    ("changed during apply", Kind::StalePlan),
    ("requires preservation review", Kind::Conflict),
    // #2727: reusing a receipt that was already reverted is a refusal on the target's state,
    // not an internal error (agents, routing and Go-state reverts).
    ("already reverted", Kind::Conflict),
    ("exclusive create refused", Kind::Conflict),
    ("malformed managed markers", Kind::Conflict),
    ("not uniquely locatable", Kind::Conflict),
    ("generation manifest", Kind::Conflict),
    ("digest mismatch", Kind::Conflict),
    // home layout (#2175): a failed integrity verify is the digest-mismatch family; the
    // copy failure is plain I/O.
    ("home layout verify failed", Kind::Conflict),
    ("SKILL.md missing", Kind::NotFound),
    // Generic "not found" suffix. First match wins, so the specific phrases before it
    // (the forge rule above) keep their kinds; the phrases after it do not contain
    // this text, so its position among them is harmless.
    ("not found", Kind::NotFound),
    ("project profile missing", Kind::NotFound),
    ("reference directory missing", Kind::NotFound),
    ("project directory unavailable", Kind::NotFound),
    ("project root unavailable", Kind::NotFound),
    ("package unavailable", Kind::NotFound),
    ("package directory unavailable", Kind::NotFound),
    ("git rev-parse exited", Kind::NotFound),
    ("path unavailable", Kind::NotFound),
    ("encode failed", Kind::Internal),
    ("rendered output", Kind::Internal),
    ("git output", Kind::Io),
    ("unexpected git", Kind::Io),
    ("git did not report", Kind::Io),
    ("git reported a path", Kind::Io),
    ("malformed ls-", Kind::Io),
    ("read failed", Kind::Io),
    ("write failed", Kind::Io),
    // C6 tracked-file count is an input refusal even inside the build wrapper.
    ("index file limit exceeded", Kind::InvalidInput),
    // incremental index (rc.2 C6 #2182): a failed build is plain I/O trouble; the
    // previous generation is kept and the query state turns partial.
    ("index build failed", Kind::Io),
    // search (rc.2 C7 #2183): every refusal is about the request itself.
    ("search refused: ", Kind::InvalidInput),
    ("home layout copy failed", Kind::Io),
    ("replace failed", Kind::Io),
    ("temporary", Kind::Io),
    ("unavailable", Kind::Io),
    ("unreadable", Kind::Io),
    ("disappeared", Kind::Io),
    ("git failed", Kind::Io),
    ("hash-object failed", Kind::Io),
    ("exited", Kind::Io),
    ("invalid", Kind::InvalidInput),
    ("nonportable", Kind::InvalidInput),
    ("UTF-8", Kind::InvalidInput),
    ("limit exceeded", Kind::InvalidInput),
    ("too large", Kind::InvalidInput),
    ("must be", Kind::InvalidInput),
    ("require", Kind::InvalidInput),
    // #2242: the project-skill copy ops refuse before any write without the
    // explicit opt-in flag.
    ("explicit opt-in required", Kind::InvalidInput),
    ("forbidden", Kind::InvalidInput),
    ("refused", Kind::InvalidInput),
    ("outside", Kind::InvalidInput),
    ("unknown", Kind::InvalidInput),
    ("missing", Kind::InvalidInput),
    ("not a ", Kind::InvalidInput),
];

/// Kind of an existing error message; an unlisted message is `internal`, never guessed.
pub fn classify(message: &str) -> Kind {
    RULES
        .iter()
        .find(|(phrase, _)| message.contains(phrase))
        .map_or(Kind::Internal, |&(_, kind)| kind)
}

pub fn success(operation: &str, result: Value) -> Value {
    json!({"protocol_version": PROTOCOL_VERSION, "operation": operation, "ok": true, "result": result})
}

/// A stale or partially applied plan names the plan operation to call next.
pub fn failure(operation: &str, kind: Kind, message: &str) -> Value {
    let mut error =
        json!({"kind": kind.as_str(), "message": message, "retryable": kind.retryable()});
    if matches!(kind, Kind::StalePlan | Kind::Partial) {
        if let Some(prefix) = operation.strip_suffix(".apply") {
            error["next_actions"] = json!([format!("{prefix}.plan")]);
        }
    }
    json!({"protocol_version": PROTOCOL_VERSION, "operation": operation, "ok": false, "error": error})
}

#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    fn permission_reference_boundaries() {
        // Accepted: an ordinary reference, exactly 128 bytes, multi-byte text at 128 bytes.
        assert!(permission_reference_ok("lane-approval-42"));
        assert!(permission_reference_ok(&"a".repeat(128)));
        let multi = format!("{}aa", "가".repeat(42));
        assert_eq!(multi.len(), 128);
        assert!(permission_reference_ok(&multi));
        assert!(permission_reference_ok(" nullable "));
        // Refused: empty or placeholder words after trimming.
        for word in [
            "",
            "   ",
            "null",
            " null ",
            "placeholder",
            "\u{a0}placeholder",
        ] {
            assert!(!permission_reference_ok(word), "{word:?}");
        }
        // Refused: one byte over the limit, also when the overflow is a multi-byte character.
        assert!(!permission_reference_ok(&"a".repeat(129)));
        assert!(!permission_reference_ok(&format!("{}가", "a".repeat(126))));
        // Refused: any control character, including inside the text.
        for bad in ["ref\n", "re\u{7}f", "\tref", "ref\u{85}"] {
            assert!(!permission_reference_ok(bad), "{bad:?}");
        }
    }

    #[test]
    fn permission_reference_rule_has_one_copy() {
        // WF-OBS #24 (#2401): the rule lives only in `permission_reference_ok`.
        let root = std::path::Path::new(env!("CARGO_MANIFEST_DIR")).join("src");
        let mut stack = vec![root];
        let mut copies = Vec::new();
        while let Some(dir) = stack.pop() {
            for entry in std::fs::read_dir(&dir).unwrap() {
                let path = entry.unwrap().path();
                if path.is_dir() {
                    stack.push(path);
                } else if path.extension().is_some_and(|e| e == "rs") {
                    let text = std::fs::read_to_string(&path).unwrap();
                    let n = text.matches("\"placeholder\"].contains(").count();
                    if n > 0 {
                        copies.push((path, n));
                    }
                }
            }
        }
        assert_eq!(copies.len(), 1, "{copies:?}");
        assert!(copies[0].0.ends_with("src/contract.rs"), "{copies:?}");
        assert_eq!(copies[0].1, 1, "{copies:?}");
    }

    #[test]
    fn every_current_message_has_a_deliberate_kind() {
        use Kind::*;
        let cases = [
            ("unsupported protocol_version; supported: 1", UnsupportedProtocolVersion),
            ("explicit opt-in required; nothing written", InvalidInput),
            ("input exceeds 64 KiB", InputTooLarge),
            ("invalid typed JSON input", InvalidInput),
            ("stdin read failed", Io),
            ("encode failed", Internal),
            ("permission reference required", PermissionReferenceMissing),
            ("stale or altered plan; re-preview required", StalePlan),
            ("target changed since plan", StalePlan),
            ("target changed since plan; re-preview required", StalePlan),
            ("target changed during apply; temporary file discarded", StalePlan),
            ("partial apply: applied [a]; failed at b: x", Partial),
            // The wrapper wins over the plan-admit version phrase it may quote.
            (
                "partial apply: applied [a]; failed at b: unsupported protocol_version; supported: 1",
                Partial,
            ),
            (
                "partial apply: applied [AGENTS.md]; failed at .gitignore: target changed since plan",
                Partial,
            ),
            (
                "conflict in AGENTS.md: managed block missing or removed; baseline mismatch; nothing applied",
                Conflict,
            ),
            ("partial creation possible; preserve and inspect target", Partial),
            ("conflict in AGENTS.md: managed block edited; nothing applied", Conflict),
            ("existing YAML requires preservation review", Conflict),
            ("exclusive create refused", Conflict),
            ("malformed managed markers", Conflict),
            ("rendered block not uniquely locatable; refused", Conflict),
            ("generation manifest unreadable; inspect before regenerating", Conflict),
            ("generation manifest belongs to another schema or project", Conflict),
            ("pinned package digest mismatch; fallback forbidden", Conflict),
            ("package SKILL.md missing", NotFound),
            ("receipt not found", NotFound),
            ("receipt backup not found; inspect before recovering", NotFound),
            ("project profile missing", NotFound),
            ("workflow reference directory missing; generate it first", NotFound),
            ("project directory unavailable", NotFound),
            ("project root unavailable", NotFound),
            ("package unavailable", NotFound),
            ("package directory unavailable", NotFound),
            ("git rev-parse exited 128", NotFound),
            ("rendered output is not UTF-8", Internal),
            ("non UTF-8 git output", Io),
            ("malformed ls-files record", Io),
            ("malformed ls-tree record", Io),
            ("target read failed", Io),
            ("temporary write failed", Io),
            ("temporary file refused", Io),
            ("atomic replace failed", Io),
            ("target metadata unavailable", Io),
            ("target parent unavailable", Io),
            ("target directory unavailable", Io),
            ("project state directory unavailable", Io),
            ("package metadata unavailable", Io),
            ("package entry unavailable", Io),
            ("directory entry unavailable", Io),
            ("git unavailable", Io),
            ("git stdin unavailable", Io),
            ("git stdin write failed", Io),
            ("git failed", Io),
            ("hash-object failed", Io),
            ("package file disappeared", Io),
            ("symlink unreadable", Io),
            ("unreadable file x", Io),
            ("git ls-files exited 1", Io),
            ("ls-remote exited 2", Io),
            ("invalid profile", InvalidInput),
            ("invalid or nonportable profile; no values were logged", InvalidInput),
            ("invalid agents request; no values were logged", InvalidInput),
            ("explicit null pin is invalid", InvalidInput),
            ("nonportable package path", InvalidInput),
            ("non UTF-8 project path", InvalidInput),
            ("target is not UTF-8", InvalidInput),
            ("package entry limit exceeded", InvalidInput),
            ("package nesting limit exceeded", InvalidInput),
            ("package too large", InvalidInput),
            ("project must be a directory", InvalidInput),
            ("go baseline commit must be 40 lowercase hex", InvalidInput),
            ("user package required without pin", InvalidInput),
            ("explicit pin requires repository package; fallback forbidden", InvalidInput),
            ("symlink target refused", InvalidInput),
            ("package symlink refused", InvalidInput),
            ("package special file refused", InvalidInput),
            ("special file refused: x", InvalidInput),
            ("profile value contains managed-marker syntax; refused", InvalidInput),
            ("path outside root", InvalidInput),
            ("unknown planned path", InvalidInput),
            ("manifest missing from plan", InvalidInput),
            ("target is not a bounded regular file", InvalidInput),
            ("HEAD is not a 40 hex commit", InvalidInput),
            ("package path invalid", InvalidInput),
            // queue-next (rc.2 A2 #2161)
            ("queue milestone must be a positive number", InvalidInput),
            ("queue issue page limit exceeded (50)", InvalidInput),
            // leader-tick and the profile `leader` key (rc.2 A3 #2162)
            ("leader tick page limit exceeded (50)", InvalidInput),
            ("leader tick base_branch must be a plain branch name", InvalidInput),
            ("leader pr_check_interval_secs must be a positive number", InvalidInput),
            ("leader tick state directory unavailable", Io),
            ("leader tick state write failed", Io),
            ("leader tick state replace failed", Io),
            // lane-status (rc.2 A11 #2265)
            ("lane status issues must be 1 to 50 Issue numbers", InvalidInput),
            ("lane status lanes must be at most 50", InvalidInput),
            ("lane status base_branch must be a plain branch name", InvalidInput),
            ("lane status evidence must be given with a session, a source and a time", InvalidInput),
            ("lane status page limit exceeded (50)", InvalidInput),
            // audit-claim (rc.2 #2403 item 1)
            ("audit claim refused: head must be 40 lowercase hex", InvalidInput),
            ("audit claim refused: lane must be a plain name (letters, digits, - _ .)", InvalidInput),
            ("audit claim conflict: PR #7 is not open", Conflict),
            ("audit claim conflict: head is not the current head of PR #7 (aaaa); re-read the PR and reserve its current head", Conflict),
            ("audit claim conflict: lane x is an author of PR #7 (claim comment 1); an author or repairer never audits its own PR", Conflict),
            ("audit claim blocked: PR #7 names no Refs #<issue>, so its authors cannot be read", Blocked),
            ("audit claim blocked: claim comment 1 named by a Moai-Claim trailer is not a claim line on Issue #3; authors cannot be ruled out", Blocked),
            ("audit claim blocked: the comment POST returned HTTP 403; nothing was posted", Blocked),
            // decision-append (rc.2 #2403)
            ("decision append blocked: the project has no HEAD commit", Blocked),
            ("decision append blocked: HEAD is not the given base (HEAD 0123); fetch the tip and work from a worktree at it", Blocked),
            ("decision append blocked: the log differs from HEAD in the worktree or the index; commit or discard local edits first", Blocked),
            ("decision append refused: give at least one row", InvalidInput),
            ("decision append refused: base must be a 40-hex commit", InvalidInput),
            ("decision append refused: --log must be a relative path inside the repository", InvalidInput),
            ("decision append refused: --log passes through a symlink", InvalidInput),
            ("decision append refused: --log must be UTF-8", InvalidInput),
            ("decision append refused: PROJECT must be the repository top directory", InvalidInput),
            ("decision append refused: --log is not tracked at HEAD", InvalidInput),
            ("decision append refused: the log is not UTF-8", InvalidInput),
            ("decision append refused: the log must hold exactly one table (found 2)", InvalidInput),
            ("decision append refused: the table's last line has no line terminator", InvalidInput),
            ("decision append refused: row 1 has 4 cells; the table has 5", InvalidInput),
            ("decision append refused: row 1 cell 3 holds a bare | or a line break", InvalidInput),
            ("decision append git failed to start", Io),
            ("decision append project unreadable", Io),
            ("decision append log unreadable", Io),
            ("decision append log unreadable after write", Io),
            ("decision append log replace failed; the log is unchanged", Io),
            ("decision append log replace failed: read-back differs", Io),
            // identity (01-A #1988)
            ("path unavailable", NotFound),
            ("path must be a directory", InvalidInput),
            ("path is a symlink; refused", InvalidInput),
            ("project state root is a symlink; refused", InvalidInput),
            ("home state root is a symlink; refused", InvalidInput),
            ("non UTF-8 path", InvalidInput),
            ("unexpected git rev-parse output", Io),
            ("git did not report a work tree", Io),
            ("git reported a path that cannot be resolved", Io),
            // forge-read (07-A #2008)
            ("forge blocked: authentication failed (401)", Blocked),
            ("forge blocked: access denied (403)", Blocked),
            ("forge blocked: rate limited after 3 retries", Blocked),
            ("gh transport: cannot start gh: NotFound", Transport),
            ("gh transport: gh api returned no HTTP response (exit Some(1)): x", Transport),
            ("gh transport: refused: only https://api.github.com/ URLs are requested", Transport),
            ("gh transport: GET returned HTTP 500", Transport),
            ("gh transport: 304 without a cached response", Transport),
            ("gh transport: repository response carried no id", Transport),
            ("forge resource not found or not visible to this account (404)", NotFound),
            ("unsupported request schema_version", UnsupportedProtocolVersion),
            ("owner and repo must be GitHub names", InvalidInput),
            ("state must be open, closed or all", InvalidInput),
            ("overlap_s must be >= 0 and max_pages > 0", InvalidInput),
            ("page limit exceeded (config.max_pages 50)", InvalidInput),
            ("poll state directory is a symlink or not a directory; refused", InvalidInput),
            ("poll state directory unavailable", Io),
            ("poll state write failed", Io),
            ("poll state replace failed", Io),
            ("MOAI_FORGE_GH_BIN must be an absolute path to an existing file", InvalidInput),
            // identity git failure (#2033); the last two would match `git output` without the rule
            ("git failed: detected dubious ownership in repository at", GitFailed),
            ("git failed: git output exceeded the size limit", GitFailed),
            ("git failed: git output was not UTF-8", GitFailed),
            // home layout (rc.2 B8 #2175); every current message of the module
            ("conflict in home layout: already switched; remove the home layout.json marker to revert first", Conflict),
            ("conflict in home layout: home target for receipts already exists; refusing overwrite; remove or merge it first; nothing applied", Conflict),
            ("conflict in home layout: claim journal busy; retry when no claimant holds it; nothing applied", Conflict),
            ("home layout verify failed; partial copy removed", Conflict),
            ("home layout copy failed", Io),
            ("home layout target unavailable", Io),
            ("layout source unreadable", Io),
            ("home layout marker write failed", Io),
            ("home root unavailable; HOME is not set", Io),
            // claim-verify (rc.2 A9 #2255)
            (
                "claim verify: not owner; the lowest open claim is comment 7 (REQ-A9-01), so no commit or push",
                Conflict,
            ),
            (
                "claim verify: stale; claim comment 7 lease expired without a handoff, and the operator decides (REQ-A9-03)",
                Blocked,
            ),
            ("claim verify: closed; claim comment 7 was released or handed off", Conflict),
            (
                "claim verify: uncertain; this session's claim comment is not visible on the issue",
                Blocked,
            ),
            ("claim verify: no claim of this session on the issue", Conflict),
            (
                "claim verify: trailer missing; commits abc123def lack `Moai-Claim: 7#200`, so no push",
                InvalidInput,
            ),
            (
                "claim verify: trailer check unavailable; git log of the claim base failed",
                Blocked,
            ),
            (
                "invalid claim verify request: owner, repo, issue_number and session are required",
                InvalidInput,
            ),
            (
                "claim verify: uncertain; the commit message carries no `Moai-Claim: 7#<id>`",
                Blocked,
            ),
            // archive move, restore and search (D2 #2188)
            ("readback digest mismatch; move cancelled, original kept", Conflict),
            ("source changed during apply; move cancelled, original kept", StalePlan),
            // #2727: a reverted receipt used again
            ("agent receipt already reverted", Conflict),
            ("routing receipt already reverted", Conflict),
            ("this apply was already reverted", Conflict),
            ("archive destination already exists; exclusive create refused", Conflict),
            ("archive index entry already exists; exclusive create refused", Conflict),
            ("restore destination already exists; exclusive create refused", Conflict),
            ("symlink in SPEC directory refused: notes/link.md", InvalidInput),
            ("special file in SPEC directory refused: fifo", InvalidInput),
            ("SPEC path is not a directory", InvalidInput),
            ("SPEC directory unavailable", Io),
            ("SPEC directory unreadable", Io),
            ("SPEC file name is not UTF-8", InvalidInput),
            ("SPEC file read failed: spec.md", Io),
            ("archive destination parent unavailable", Io),
            ("archive destination name invalid", InvalidInput),
            ("archive directory unavailable", Io),
            ("archive staging unavailable", Io),
            ("archive copy write failed: spec.md", Io),
            ("archive publish write failed", Io),
            ("archive index invalid", InvalidInput),
            ("archive index invalid: unknown schema_version", InvalidInput),
            ("archive index invalid: invalid spec id", InvalidInput),
            ("archive index invalid: spec path mismatch", InvalidInput),
            ("invalid spec id: ../x", InvalidInput),
            ("spec not found in archive index: SPEC-X", NotFound),
            ("archive manifest not found", NotFound),
            ("archive manifest invalid", InvalidInput),
            ("archive manifest digest mismatch; restore refused", Conflict),
            ("archived copy digest mismatch; restore refused", Conflict),
            (
                "partial apply: archived [SPEC-A]; failed at SPEC-B: source directory removal write failed; archived copy and index kept",
                Partial,
            ),
            (
                "partial apply: archived [SPEC-A]; failed at SPEC-A: source changed during apply; source removal refused; archived copy and index kept",
                Partial,
            ),
            (
                "partial apply: restored [SPEC-A]; failed at archive cleanup: archive manifest removal write failed",
                Partial,
            ),
            ("archive manifest removal write failed", Io),
            ("archived copy removal write failed", Io),
            ("search query must be 1 to 256 characters", InvalidInput),
            // hooks-plan / hooks-apply (rc.2 I1 #2238)
            ("unknown hook host; must be claude, codex or omp", InvalidInput),
            (
                "home directory refused; hooks are written to a project tree only",
                InvalidInput,
            ),
            ("existing .codex/hooks.json requires preservation review", Conflict),
            (".codex/hooks.json receipt requires preservation review", Conflict),
            // incremental index (rc.2 C6 #2182); every current message of the module
            ("index build failed; previous generation kept: a.txt read failed", Io),
            ("index source unavailable: no HEAD commit", Io),
            ("index generation staging write failed: x", Io),
            ("index file limit exceeded (16384)", InvalidInput),
            ("index manifest encode failed", Internal),
            ("index pointer write failed", Io),
            ("index pointer replace failed", Io),
            ("index pointer encode failed", Internal),
            // search (rc.2 C7 #2183); every current message of the module
            ("search refused: give at least one path or literal", InvalidInput),
            ("search refused: too many query terms", InvalidInput),
            (
                "search refused: a path must be relative inside the repository",
                InvalidInput,
            ),
            (
                "search refused: a literal must be non-empty, one line, at most 512 bytes",
                InvalidInput,
            ),
            ("search refused: max_files must be 1..=2000", InvalidInput),
            (
                "search refused: max_spans_per_file must be 1..=1000",
                InvalidInput,
            ),
        ];
        for (message, kind) in cases {
            assert_eq!(classify(message), kind, "{message}");
        }
        assert_eq!(classify("something new"), Internal);
    }

    #[test]
    fn journal_lock_busy_sets_its_kind_at_the_raise_site() {
        // #2345: the busy error carries `unavailable` explicitly; it never depends on
        // classify()'s message guessing, and it is retryable.
        let e = OpError::new(Kind::Unavailable, "journal busy: test");
        assert_eq!(e.kind, Kind::Unavailable);
        assert!(e.kind.retryable());
        assert_eq!(e.kind.as_str(), "unavailable");
        // A plain String error keeps the classified fallback at the boundary.
        let fallback = OpError::from("target read failed".to_string());
        assert_eq!(fallback.kind, Kind::Io);
        // AC-3: a message raised as conflict that happens to contain "unavailable" stays
        // conflict — the "conflict in " prefix rule wins over the generic word, so no
        // busy reclassification may steal a deliberate conflict message.
        assert_eq!(
            classify("conflict in claim journal: stale_epoch 2; the journal was unavailable"),
            Kind::Conflict
        );
        assert_eq!(
            classify(
                "conflict in home layout: claim journal busy; retry when no claimant holds it; nothing applied"
            ),
            Kind::Conflict
        );
        // The shared busy message, if it ever reached classify through a non-typed path,
        // degrades to `io` (still retryable) — but the typed path never lets it.
        assert_eq!(
            classify("journal busy: the claim journal lock is unavailable (another process holds it); the caller may retry the same command after a bounded wait, within its own cap (REQ-06A-05)"),
            Kind::Io
        );
    }

    #[test]
    fn unavailable_failure_envelope_is_retryable() {
        let busy = failure(
            "claim.acquire",
            Kind::Unavailable,
            "journal busy: the claim journal lock is unavailable",
        );
        assert_eq!(busy["ok"], json!(false));
        assert_eq!(busy["error"]["kind"], "unavailable");
        assert_eq!(busy["error"]["retryable"], json!(true));
        assert!(busy["error"].get("next_actions").is_none());
    }

    #[test]
    fn envelopes_carry_the_decided_fields() {
        let ok = success("project.plan", json!({"a": 1}));
        assert_eq!(
            ok,
            json!({"protocol_version":1,"operation":"project.plan","ok":true,"result":{"a":1}})
        );
        let stale = failure("agents.apply", Kind::StalePlan, "target changed since plan");
        assert_eq!(stale["ok"], json!(false));
        assert_eq!(stale["error"]["kind"], "stale_plan");
        assert_eq!(stale["error"]["retryable"], json!(false));
        assert_eq!(stale["error"]["next_actions"], json!(["agents.plan"]));
        let io = failure("package.inspect", Kind::Io, "package entry unavailable");
        assert_eq!(io["error"]["retryable"], json!(true));
        assert!(io["error"].get("next_actions").is_none());
    }
}
