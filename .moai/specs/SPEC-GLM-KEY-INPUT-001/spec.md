---
id: SPEC-GLM-KEY-INPUT-001
title: "GLM API Key Input Surface in the moai web Settings Console"
version: "0.2.0"
status: completed
created: 2026-07-25
updated: 2026-07-30
author: manager-spec
priority: P1
phase: "v3.2.0 target"
module: "internal/web, internal/settings, internal/cli"
lifecycle: spec-anchored
tags: "web-console, glm, credential, secret-handling, file-mode, settings-schema"
tier: M
---

## HISTORY

| Version | Date       | Author       | Change                          |
|---------|------------|--------------|---------------------------------|
| 0.1.0   | 2026-07-25 | manager-spec | Initial plan-phase authoring.   |
| 0.2.0   | 2026-07-25 | manager-spec | Plan-phase decisions resolved. D-1: credential writer extracts to new leaf package `internal/glmcred`. D-2: key field is out-of-schema hand-built, not a fifth `PersistKind`. D-3: mode tightening applies to both the console and `moai glm`. Scope change: the trailing-4-character disambiguation hint is now IN scope — REQ-GKI-004-002 rewritten, REQ-GKI-004-004 added for the short-key guard, §D exclusion narrowed. |

---

## §A. Vision and Scope

### §A.1 Problem

A GLM API key can be supplied through exactly one path today: the interactive
`moai glm` terminal flow, which calls `saveGLMKey` (`internal/cli/glm.go`) and
writes `~/.moai/.env.glm` with mode `0600`.

The `moai web` settings console — the surface a user is directed to for
configuring MoAI — has no GLM key field at all. A user who configures the
project through the console can set the four GLM model tier mappings
(`glm.models.{high,medium,low,fable}`, rendered in the "3rd Party LLM"
section) but cannot supply the credential those mappings depend on. They must
drop to the terminal and run a different command to finish a task the console
otherwise presents as complete.

### §A.2 Goal

Add a GLM API key input to the settings console that writes through the
existing credential path, so the console becomes a complete configuration
surface for the GLM backend.

### §A.3 The central hazard this SPEC exists to prevent

Every field the settings console renders today persists through one of four
`PersistKind` routes declared in `internal/settings/schema.go`. All four
terminate in a YAML file that is **not credential-grade**:

| PersistKind            | Destination                                          | Why it is wrong for a secret          |
|------------------------|------------------------------------------------------|---------------------------------------|
| `PersistProfileStore`  | `~/.moai/claude-profiles/<name>/preferences.yaml`     | Ordinary preferences file; copied and synced as ordinary preferences; not written at `0600` |
| `PersistProjectConfig` | `.moai/config/sections/quality.yaml`, `git-convention.yaml` | Git-tracked project config           |
| `PersistTypedSection`  | `.moai/config/sections/llm.yaml` and siblings         | Git-tracked project config — this is where the sibling `glm.models.*` fields land |
| `PersistSeam`          | `.moai/config/sections/*.yaml` via `yamlpatch`        | Git-tracked project config           |

Routing an API key through any of them is a permission downgrade from `0600`
to a world-readable file, and for three of the four it additionally commits a
live secret to version control. The adjacency is the trap: the natural home
for the key field — beside `glm.models.*` in the "3rd Party LLM" section — is
precisely the route that would write it into `llm.yaml`.

Therefore the governing constraint of this SPEC: **the key input is rendered
alongside profile-store- and config-backed fields but must not persist like
them.** It persists only through the credential writer that owns
`~/.moai/.env.glm` at mode `0600`.

### §A.4 Structural constraint discovered during authoring

`saveGLMKey` and `loadGLMKey` are unexported members of package `cli`, and the
dependency runs one way: `internal/cli/web.go` imports `internal/web`, so
`internal/web` cannot import `internal/cli` without creating an import cycle.
The console therefore cannot reach the existing writer as the code stands.

This is not a reason to write a second implementation. Two writers means two
mode policies and two escaping rules, which is the defect class this SPEC is
written to avoid. The credential writer is therefore lifted into a new leaf
package, `internal/glmcred`, which both `internal/cli` and `internal/web`
import, with the CLI path delegating to it so exactly one implementation
survives. The package is stdlib-only, so it is cycle-free by construction.
See `plan.md` §A.1 D-1 for the rationale and `plan.md` §F M1 for the work.

---

## §B. Requirements

### §B.1 Input surface

- REQ-GKI-001-001: The settings console shall render a GLM API key input field within the "3rd Party LLM" section, adjacent to the existing GLM model tier fields.
- REQ-GKI-001-002: The GLM API key input field shall be rendered as a secret-class control that masks typed characters in the browser and opts out of browser autofill.
- REQ-GKI-001-003: When the operator submits the settings form carrying a non-empty GLM API key value, the console shall persist that value through the single credential writer that owns the GLM credential file.
- REQ-GKI-001-004: When the credential write succeeds, the console shall confirm the outcome in the save banner without reproducing the submitted key in the response.

### §B.2 Credential-grade persistence

- REQ-GKI-002-001: The console shall persist GLM API key material only to the GLM credential file at `~/.moai/.env.glm`, in the quoted `GLM_API_KEY` dotenv form the existing writer produces.
- REQ-GKI-002-002: The GLM credential file written on behalf of the console shall carry file mode `0600`.
- REQ-GKI-002-003: The console shall delegate the credential write to the shared writer rather than reimplementing dotenv escaping, path resolution, or file-mode selection.
- REQ-GKI-002-004: While any submitted settings field is failing validation, the console shall not write the GLM API key to disk.
- REQ-GKI-002-005: When the credential write fails, the console shall surface the failure to the operator and shall not present the save as successful.

### §B.3 Prohibited persistence routes

- REQ-GKI-003-001: The console shall not write GLM API key material to the profile store preferences file.
- REQ-GKI-003-002: The console shall not write GLM API key material to any file under `.moai/config/sections/`, including `llm.yaml`.
- REQ-GKI-003-003: The console shall exclude the GLM API key field from the bulk current-value read that repopulates schema form state from persisted YAML.

### §B.4 Redaction of stored key material

- REQ-GKI-004-001: The console shall not emit a stored GLM API key in full in any HTTP response body, and shall not populate any rendered form value attribute with stored key material.
- REQ-GKI-004-002: Where a GLM API key of more than four characters is already stored, the console shall render a presence indicator that discloses at most the final four characters of the stored key and no other part of it.
- REQ-GKI-004-003: The console shall not write GLM API key material to server logs or into rendered error messages.
- REQ-GKI-004-004: Where the stored GLM API key is four characters or fewer, the console shall render the presence indicator with no characters of the stored key disclosed.

### §B.5 Submitted-value validation

- REQ-GKI-005-001: When the operator submits an empty or whitespace-only GLM API key value, the console shall preserve the currently stored key and shall not rewrite the credential file.
- REQ-GKI-005-002: When the submitted GLM API key value carries surrounding whitespace around a non-empty body, the console shall trim the value before persisting it.
- REQ-GKI-005-003: When the submitted GLM API key value contains a line-break character, the console shall reject the submission with a per-field validation error and shall leave the stored key unchanged.

### §B.6 Pre-existing credential file

- REQ-GKI-006-001: When a GLM API key is submitted and the credential file already exists, the console shall replace the stored key value in place rather than creating a second credential location.
- REQ-GKI-006-002: When the credential file already exists with a mode wider than `0600`, the shared credential writer shall tighten the mode to `0600` as part of the write, on both the console save path and the `moai glm` interactive save path.
- REQ-GKI-006-003: The console shall not delete or truncate the credential file when no GLM API key value is submitted.

---

## §C. Constraints

- C-1: `~/.moai/.env.glm` is the single storage location for the GLM credential. No second credential file, no duplicate copy, no cache.
- C-2: Exactly one credential writer implementation may exist in the tree after this SPEC lands. The `moai glm` interactive flow and the console must call the same function.
- C-3: `loadGLMKey`'s existing consumers (`internal/cli/glm.go` and `internal/cli/glm_tools.go`) must keep reading the same file in the same format. The dotenv shape is a compatibility surface, not an implementation detail.
- C-4: The `MOAI_TEST_GLM_KEY` environment override that `loadGLMKey` honours is a test seam and must keep working after any extraction.
- C-5: The console's existing atomic-reject contract holds: a validation failure anywhere in the submitted form leaves all persisted state unchanged, credential file included.
- C-6: `os.WriteFile` does not narrow the mode of a file that already exists. Satisfying REQ-GKI-002-002 for a pre-existing wide-mode file therefore requires an explicit mode assertion, not a `perm` argument.
- C-7: The four sibling `glm.models.*` fields keep their current `PersistTypedSection` route into `llm.yaml`. This SPEC does not move them.

---

## §D. Exclusions

### Out of Scope — key lifecycle management

- Automatic key rotation, expiry tracking, or revocation.
- Storing more than one GLM credential at a time, or per-profile credentials.

### Out of Scope — other credential surfaces

- Anthropic / Claude API token entry in the console.
- OS keychain backends (macOS Keychain, libsecret, Windows Credential Manager).
- Encrypting the credential file at rest.

### Out of Scope — disclosure beyond the trailing-four hint

The trailing-four-character hint permitted by REQ-GKI-004-002 is the complete
extent of sanctioned disclosure. Everything below stays excluded:

- A reveal or "show key" toggle that renders the full stored key.
- A copy-the-stored-key-to-clipboard control.
- Disclosing a leading fragment, a middle fragment, or any run longer than the final four characters.
- Rendering the stored key's length, entropy, or checksum as a separate signal.

### Out of Scope — key validity verification

- Any network call that checks the submitted key against the z.ai endpoint.
- Reporting whether a stored key is expired or revoked.

### Out of Scope — CLI flow redesign

- Changes to the `moai glm` interactive prompt beyond routing it through the shared writer and inheriting the mode-tightening in REQ-GKI-006-002.
- Changes to `moai cg` / `moai cc` launcher behaviour.

---

## §E. Assumptions

- A-1: The operator running `moai web` owns the account whose home directory holds `~/.moai/.env.glm`. The console binds to loopback only, so the credential never crosses a network boundary.
- A-2: A single GLM credential per machine is sufficient. Profiles select models and effort, not credentials.
- A-3: An empty submitted key means "leave it alone", consistent with the console's established empty-is-preserve convention for schema fields.

---

## §F. Risks

- R-1: A future generic loop over schema fields — a bulk value read, a form-state dump, a diagnostics view — silently picks up the credential field and leaks it. Structurally mitigated by the resolved D-2 decision: the key field is not a `FieldDef`, so it is not in the set those loops iterate. REQ-GKI-003-003 and its anti-leak test are the standing guard that the field stays out of that set.
- R-2: Mode-tightening (REQ-GKI-006-002) changes behaviour on the `moai glm` path as well. Accepted per the resolved D-3 decision: permissions only ever narrow, never widen, and a world-readable credential file is itself the defect being corrected. Recorded as a silent correction, not a breaking change.
- R-3: Extracting the writer touches a live CLI path. A regression there breaks `moai glm` for every user, which is a worse outcome than the missing console field.
- R-4: The trailing-four hint (REQ-GKI-004-002) is a deliberate partial disclosure accepted for key disambiguation. Its risk is drift — a later change that widens the window, moves it to the leading characters, or drops the short-key guard turns a bounded hint into a real leak. Mitigated by AC-GKI-004-02, which permits exactly the final four characters and fails on any other disclosed run, and by AC-GKI-004-04 for keys of four characters or fewer.

---

## §G. Acceptance Criteria Summary

Full criteria with commands live in `acceptance.md`. One criterion per requirement:

- AC-GKI-001-01: Given the console is running, When the operator opens the settings page, Then the 3rd Party LLM section renders a GLM API key field (maps REQ-GKI-001-001)
- AC-GKI-001-02: Given the settings page is rendered, When the GLM key control is inspected, Then it is a masked secret-class input with autofill disabled (maps REQ-GKI-001-002)
- AC-GKI-001-03: Given a non-empty key is submitted, When the save completes, Then the credential file contains that key (maps REQ-GKI-001-003)
- AC-GKI-001-04: Given a non-empty key is submitted, When the success response is rendered, Then the banner confirms the save and the response body does not contain the key (maps REQ-GKI-001-004)
- AC-GKI-002-01: Given a key is submitted, When the write completes, Then the only file containing the key is the GLM credential file in the expected dotenv form (maps REQ-GKI-002-001)
- AC-GKI-002-02: Given a key is submitted, When the credential file is stat-ed, Then its permission bits are exactly 0600 (maps REQ-GKI-002-002)
- AC-GKI-002-03: Given the tree after implementation, When credential-write implementations are enumerated, Then exactly one exists and both callers delegate to it (maps REQ-GKI-002-003)
- AC-GKI-002-04: Given a submission with a key plus an invalid sibling field, When the save is rejected, Then the credential file is unchanged (maps REQ-GKI-002-004)
- AC-GKI-002-05: Given the credential write fails, When the response is rendered, Then it reports failure rather than success (maps REQ-GKI-002-005)
- AC-GKI-003-01: Given a key is submitted, When the profile store is scanned, Then no file under it contains the key material (maps REQ-GKI-003-001)
- AC-GKI-003-02: Given a key is submitted, When the project config sections are scanned, Then no file under them contains the key material (maps REQ-GKI-003-002)
- AC-GKI-003-03: Given a stored key exists, When the bulk schema current-value read runs, Then the credential field is absent from its result (maps REQ-GKI-003-003)
- AC-GKI-004-01: Given a stored key exists, When any console route is fetched, Then no response body contains the full key (maps REQ-GKI-004-001)
- AC-GKI-004-02: Given a stored key longer than four characters exists, When the settings page is rendered, Then the final four characters appear as the hint and a differential window scan shows no other run of the key disclosed (maps REQ-GKI-004-002)
- AC-GKI-004-03: Given a stored key exists, When logs and error renderings are captured across a failing save, Then none contains the key (maps REQ-GKI-004-003)
- AC-GKI-004-04: Given a stored key of four characters or fewer, When the settings page is rendered, Then the presence indicator appears and a differential window scan shows zero disclosed characters (maps REQ-GKI-004-004)
- AC-GKI-005-01: Given a stored key exists, When an empty key field is submitted, Then the stored key and the file mtime are unchanged (maps REQ-GKI-005-001)
- AC-GKI-005-02: Given a key with surrounding whitespace is submitted, When it is read back, Then the trimmed value is stored (maps REQ-GKI-005-002)
- AC-GKI-005-03: Given a key containing a newline is submitted, When the save is attempted, Then a per-field validation error is returned and the stored key is unchanged (maps REQ-GKI-005-003)
- AC-GKI-006-01: Given a credential file already holds a key, When a new key is submitted, Then the file holds exactly one key entry with the new value (maps REQ-GKI-006-001)
- AC-GKI-006-02: Given a credential file exists at mode 0644, When a key is written from the console path and again from the `moai glm` path, Then the file mode is 0600 after each (maps REQ-GKI-006-002)
- AC-GKI-006-03: Given a credential file exists, When a save is performed with no key submitted, Then the file still exists with its content intact (maps REQ-GKI-006-003)

---

## §H. Cross-References

- `internal/cli/glm.go` — `getGLMEnvPath`, `saveGLMKey`, `loadGLMKey`, `escapeDotenvValue`, `unescapeDotenvValue`: the current credential path.
- `internal/settings/schema.go` — `PersistKind` constants and the `PersistTarget` / `FieldDef` records that declare where a field's value lands.
- `internal/settings/schema_sections.go` — `llmFields()`: the four GLM tier fields the new field sits beside.
- `internal/settings/sectionapply.go` — `ApplySchemaEdits`: the persistence dispatcher whose `default` branch rejects unknown persist kinds.
- `internal/web/handlers.go` — `handleSave`: the console save path and its atomic-reject contract.
- `internal/web/schemaform.go` — `parseSchemaForm`: the generic form parser and its empty-is-preserve convention.
- `.claude/rules/moai/core/moai-constitution.md` § Quality Gates — TRUST 5 Secured.
- `CLAUDE.local.md` §2 — the `settings.local.json` separation rule, the closest existing precedent for keeping machine-local secret-adjacent state out of shared files.
