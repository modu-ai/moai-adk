//! GitLab forge contract declaration (Issue #2171, REQ-B4-01) — metadata only.
//!
//! Decided 2026-11-09 01:1x (`docs/decisions-2026-11-08.md`: GitLab contract only — no
//! account). This module declares the capabilities the official GitLab REST
//! documentation describes and NOTHING else: there is no GitLab transport,
//! client, URL or call path anywhere in this crate, and growing one is a
//! separate issue. A declaration is not a route (REQ-B4-01).

use serde::{Deserialize, Serialize};

/// Whether a declared capability is backed by anything measured. `Documented`
/// means the provider's official documentation describes it; `Unknown` means
/// nothing has been verified for it. Neither value creates a call path.
#[derive(Clone, Copy, Debug, PartialEq, Eq, Serialize, Deserialize)]
#[serde(rename_all = "snake_case")]
pub enum CapabilityStatus {
    Unknown,
    Documented,
}

/// One declared capability of a forge provider.
#[derive(Clone, Copy, Debug, PartialEq, Eq, Serialize, Deserialize)]
pub struct ProviderCapability {
    /// Stable capability name (e.g. `merge_request`, `pipeline`, `note`).
    pub name: &'static str,
    pub status: CapabilityStatus,
}

/// The declaration surface of a forge provider (REQ-B4-01). Implementing this
/// trait is documentation, never a route: a provider with declared
/// capabilities but no adapter module has no call path at all, and
/// `moai-cli` must not grow one without its own issue.
pub trait ForgeCapabilities {
    /// The provider's stable name.
    fn provider(&self) -> &'static str;
    /// Declared capabilities and their status.
    fn capabilities(&self) -> &'static [ProviderCapability];
}

/// GitLab contract (decisions 2026-11-09 01:1x: contract only — no account). The three
/// capabilities the SYNC procedure needs (merge request, pipeline, note) are
/// `documented` by the official GitLab REST documentation. Every other
/// capability is unknown by absence — it is not declared here at all, and a
/// caller must read that absence as [`CapabilityStatus::Unknown`].
pub struct GitLabContract;

/// The declared GitLab capabilities (fixture schema: `schemas/gitlab-fixture.schema.json`,
/// sample: `tests/fixtures/gitlab/sample-fixture.json`).
pub const GITLAB_CAPABILITIES: &[ProviderCapability] = &[
    ProviderCapability {
        name: "merge_request",
        status: CapabilityStatus::Documented,
    },
    ProviderCapability {
        name: "pipeline",
        status: CapabilityStatus::Documented,
    },
    ProviderCapability {
        name: "note",
        status: CapabilityStatus::Documented,
    },
];

impl ForgeCapabilities for GitLabContract {
    fn provider(&self) -> &'static str {
        "gitlab"
    }
    fn capabilities(&self) -> &'static [ProviderCapability] {
        GITLAB_CAPABILITIES
    }
}
