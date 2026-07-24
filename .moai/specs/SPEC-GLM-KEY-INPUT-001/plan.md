---
id: SPEC-GLM-KEY-INPUT-001
title: "Implementation Plan — GLM API Key Input Surface in the moai web Settings Console"
version: "0.2.0"
status: draft
created: 2026-07-25
updated: 2026-07-25
author: manager-spec
priority: P1
phase: "v3.2.0 target"
module: "internal/web, internal/settings, internal/cli"
lifecycle: spec-anchored
tags: "web-console, glm, credential, secret-handling, file-mode, settings-schema"
tier: M
---

# Implementation Plan

Milestones below are ordered by decision-reversibility: the structural and
type-interface decisions that are expensive to undo come first, the mechanical
work last. Review attention belongs at the top.

---

## §A. Context

### §A.1 Resolved decisions

All three plan-phase decisions are RESOLVED. No open clarifications remain and
no clarification markers are outstanding in this plan.

**D-1 — Host package for the extracted credential writer. RESOLVED: new leaf package `internal/glmcred`.**

`internal/web` cannot import `internal/cli`; the dependency already runs the
other way (`internal/cli/web.go:12` imports `internal/web`). The writer moves
to `internal/glmcred`, a new stdlib-only leaf package that both `internal/cli`
and `internal/web` import.

`internal/config` was considered and rejected. It is already imported by both
callers and already owns `EnvTestGLMKey`, so it would have worked mechanically
— but it is the YAML-config owner, and a credential writer belongs to a
different concern with a different file format (dotenv, not YAML), a different
lifetime, and a different permission policy. A package that owns project
configuration should not also own secret storage. `internal/glmcred` has a
single responsibility and, importing nothing beyond stdlib, cannot participate
in any future import cycle.

**D-2 — Persist route for the key field. RESOLVED: out-of-schema hand-built field.**

The field is rendered directly in `fieldsets.templ` following the
`fieldsetLaunch` precedent (which hand-builds the quality-extras toggle), and
parsed by a dedicated branch in `handleSave`. It does not become a `FieldDef`
and never enters `AllFields()`. The rejected alternative was a fifth
`PersistKind` (`PersistCredential`) routed through `ApplySchemaEdits`.

*This deliberately places the key field outside the `schema.go` SSOT*, and
that is a real cost worth naming. `schema.go` is the single record set from
which both the `moai init` wizard and the web console derive widget type,
validation, i18n labels, and persistence target. A field outside it loses that
machinery and must be hand-wired, and it introduces a second place a reader
must look to enumerate everything the console can write.

The trade is worth it here for one reason: REQ-GKI-003-003 and risk R-1 both
describe the same failure mode — a future generic loop over `AllFields()`
picking up the credential and writing or rendering it. Keeping the field out
of that set makes the guarantee **structural**: the leak cannot happen because
the field is not in the collection being iterated. The fifth-`PersistKind`
alternative would instead make it a **standing audit obligation** — every
existing and future `switch f.Persist.Kind` and every `range AllFields()` in
the tree would acquire a new case that must be handled correctly, and a missed
case falls through to a YAML writer, which is exactly the outcome this SPEC
exists to prevent. A one-time cost in wiring is preferable to a permanent
obligation on every future schema-walking code path.

The SSOT divergence is bounded and documented: the credential is the only
console-writable value that is not a `FieldDef`, and M6 carries a regression
test asserting it stays absent from `AllFields()` so the exception cannot
silently grow into a pattern.

**D-3 — Mode tightening on the shared path. RESOLVED: it applies to both surfaces.**

Because M1 makes `moai glm` and the console share one writer, the tightening
required by REQ-GKI-006-002 reaches the CLI path too. A user whose
`~/.moai/.env.glm` is currently `0644` will find it `0600` after their next
save from **either** surface. This is accepted and intended.

Rationale to carry into the changelog: the permission only ever narrows and
never widens, and the current `0644` state is itself the defect — a
world-readable file holding a live API credential. `0600` is already the mode
the writer intends for a newly created file, so the change makes existing
files match the policy that new files have always followed. It is recorded as
a **silent correction, not a breaking change**: nothing a user can legitimately
depend on is removed, since no supported workflow requires the credential file
to be readable by other accounts.

### §A.2 Measured starting state

Every path and line reference below was read directly from the tree at
`spec/glm-key-input` before this plan was written.

| Fact | Location |
|------|----------|
| Credential path resolver | `internal/cli/glm.go` — `getGLMEnvPath`, returns `~/.moai/.env.glm` |
| Credential writer | `internal/cli/glm.go` — `saveGLMKey`, `os.MkdirAll(dir, 0o755)` then `os.WriteFile(path, content, 0o600)` |
| Credential reader | `internal/cli/glm.go` — `loadGLMKey`, honours the `MOAI_TEST_GLM_KEY` override via `config.EnvTestGLMKey` |
| Dotenv escaping | `internal/cli/glm.go` — `escapeDotenvValue` escapes `\`, `"`, `$`; **does not escape newlines** |
| Reader consumers | `internal/cli/glm.go` (×2), `internal/cli/glm_tools.go` (×2) |
| Only writer caller today | the interactive `moai glm` key prompt in `internal/cli/glm.go` |
| Package direction | `internal/cli/web.go:12` imports `internal/web`; `internal/web` imports no `internal/cli` |
| Console has no key field | grep for `GLM_API_KEY` / `apiKey` across `internal/settings/` and `internal/web/` returns zero non-test matches |
| Sibling GLM fields | `internal/settings/schema_sections.go` — `llmFields()` returns four `glm.models.{high,medium,low,fable}` typed fields |
| Their destination | `PersistTypedSection` → `.moai/config/sections/llm.yaml`, applied in `internal/settings/sectionapply.go` |
| Section label | `internal/web/schemaform.go` — `{settings.SectionLLM, "rocket", "3rd Party LLM", ...}` |
| Save handler | `internal/web/handlers.go` — `handleSave`, runs all validators, merges `FieldErrors`, atomic-rejects before any write |
| Generic form parser | `internal/web/schemaform.go` — `parseSchemaForm`, empty submitted value means preserve (EC-1) |
| Persist dispatcher | `internal/settings/sectionapply.go` — `ApplySchemaEdits`, `default` branch errors on an unhandled `PersistKind` |
| Bulk value read | `internal/settings/sectionvalues.go` — `SchemaCurrentValues` |

### §A.3 Two latent defects this SPEC closes

Both were found by reading the current writer, and both exist today on the
`moai glm` path independent of the console work.

- **Newline corruption.** `escapeDotenvValue` escapes `\`, `"` and `$` but not
  `\n`. A key containing a newline is written raw inside the quoted value;
  `loadGLMKey` scans line by line and would return a truncated key. REQ-GKI-005-003
  closes this by rejecting the input at the boundary rather than by extending
  the escaper, because a newline in a GLM API key is malformed input, not a
  value to be faithfully round-tripped.
- **Mode not narrowed on rewrite.** `os.WriteFile`'s `perm` argument applies
  only when the file is created. A `~/.moai/.env.glm` that already exists at
  `0644` stays `0644` after `saveGLMKey` rewrites it. REQ-GKI-006-002 closes
  this with an explicit mode assertion.

---

## §B. Known Issues and Prior Art

- `internal/web/validate.go:23` documents that the one-way `internal/cli` →
  `internal/web` dependency already blocked a cross-package reference once
  before; the resolution there was to duplicate a list. Duplication is
  acceptable for a list of option strings and unacceptable for a credential
  writer, because two writers means two mode policies. This is the precedent
  that makes extraction (M1) the right answer rather than a second
  implementation.
- `handleSave` already carries a hard rule against the web layer marshalling
  YAML directly (`@MX:REASON` at `internal/web/handlers.go`). The credential
  write is the same class of boundary rule and should be expressed the same
  way: the web layer calls a writer, it does not format the file.
- `CLAUDE.local.md` §2 keeps machine-local, secret-adjacent state out of
  git-tracked and template-managed files. The credential file follows that
  same doctrine and lives in `~/.moai/`, never in the project tree.

---

## §C. Pre-flight

Run before the first edit:

1. `go build ./...` — establish a green baseline.
2. `go test ./internal/cli/... ./internal/web/... ./internal/settings/...` — record the pre-change pass set, so an M1 regression on the shared CLI path is attributable.
3. `grep -rn "saveGLMKey\|loadGLMKey\|getGLMEnvPath\|escapeDotenvValue\|unescapeDotenvValue" internal/ --include='*.go'` — enumerate every call site the extraction must update, including tests.
4. Confirm `~/.moai/.env.glm` on the development machine is backed up or absent, since M1 and M5 exercise real credential writes if a test escapes its `t.TempDir()` sandbox.

---

## §D. Constraints

Inherits §C of `spec.md`. Additional implementation constraints:

- All credential tests must sandbox the home directory. `saveGLMKey` resolves
  its path through the `userHomeDirFn` indirection in `internal/cli/glm.go`;
  the extracted package must keep an equivalent seam so tests never touch the
  developer's real `~/.moai/.env.glm`.
- Per `CLAUDE.local.md` §6, temp directories come from `t.TempDir()`. Do not
  set `HOME` via `t.Setenv` in parallel tests — use the injectable home-dir
  function instead.
- Anti-leak tests assert on a sentinel key value distinctive enough to grep
  for without false positives, and must scan file *contents*, not filenames.
- Per `CLAUDE.local.md` §2 Template-First: no template files are touched by
  this SPEC, since the credential lives outside the project tree. Confirm this
  holds — if any milestone starts editing `internal/template/templates/`, that
  is a signal the credential is being persisted in the wrong place.

---

## §E. Self-Verification

Before declaring the run phase complete:

- E1: Every AC in `acceptance.md` has a PASS/FAIL result backed by an actually-executed command whose output is quoted.
- E2: `go build ./...` and `go vet ./...` clean.
- E3: `go test ./internal/cli/... ./internal/web/... ./internal/settings/...` green, compared against the §C baseline.
- E4: The anti-leak battery (M6) is run against a real save, not a mocked one.
- E5: `golangci-lint run` shows no new findings.
- E6: `go run ./cmd/moai spec lint .moai/specs/SPEC-GLM-KEY-INPUT-001/spec.md` still reports no findings.

---

## §F. Milestones

### M1 — Extract the credential writer to `internal/glmcred`

Implements D-1 and D-3. Least reversible: it creates a package boundary,
changes a live CLI path, and every later milestone depends on its interface.

Files:

- `internal/glmcred/glmcred.go` (new, stdlib-only) — receives `Path()`, `Save(key string) error`, `Load() string`, the dotenv escape/unescape helpers, and the injectable home-dir seam.
- `internal/cli/glm.go` — `saveGLMKey` / `loadGLMKey` / `getGLMEnvPath` / `escapeDotenvValue` / `unescapeDotenvValue` become thin delegations, or are deleted with call sites repointed. Retain the `MOAI_TEST_GLM_KEY` override behaviour (C-4).
- `internal/cli/glm_tools.go` — two `loadGLMKey` call sites repointed.
- New package test file — mode-tightening, escaping, round-trip, and the pre-existing-`0644` case.

Behaviour change in this milestone: `Save` explicitly asserts mode `0600`
after writing, closing the §A.3 defect (REQ-GKI-006-002). Per D-3 this reaches
the `moai glm` path as well as the console, by design.

Requirements: REQ-GKI-002-001, REQ-GKI-002-002, REQ-GKI-002-003, REQ-GKI-006-001, REQ-GKI-006-002.

Exit: `moai glm` key save still works end to end; the new package's tests
cover the `0644` → `0600` narrowing from both call paths; exactly one
credential-write implementation remains in the tree.

### M2 — Route the console's key field to the credential writer

Implements D-2, the type-interface decision. Second-least reversible: it
determines whether a credential ever enters the `FieldDef` record type. It
does not — the field stays out of schema.

Files:

- `internal/web/handlers.go` — a dedicated parse-and-persist branch in `handleSave`, calling `glmcred.Save`, ordered after validation so the atomic-reject contract holds.
- `internal/web/schemaform.go` — assert by construction that `parseSchemaForm` cannot see the field, since it iterates `settings.AllFields()` and the credential is not a member.

Explicitly NOT touched, and this is the point of D-2: `internal/settings/schema.go`
gains no `PersistCredential` constant, `llmFields()` gains no credential field,
`ApplySchemaEdits` gains no routing branch, and `SchemaCurrentValues` needs no
skip clause. Every one of those would have been a new leak site to audit.

Requirements: REQ-GKI-001-003, REQ-GKI-002-004, REQ-GKI-003-001, REQ-GKI-003-002, REQ-GKI-003-003.

Exit: a submitted key lands in `~/.moai/.env.glm` and nowhere else; a
validation failure elsewhere in the form leaves the credential file untouched.

### M3 — Render the input surface with redaction

User-facing; changes what the operator sees.

Files:

- `internal/web/fieldsets.templ` — the GLM API key control in the 3rd Party LLM section: `type="password"`, `autocomplete="off"`, empty `value` attribute unconditionally (the control is never pre-filled), and a presence indicator when a key is already stored.
- `internal/web/fieldsets_templ.go` — regenerated (`templ generate`; per project memory, `templ` is go.mod-pinned).
- `internal/web/handlers.go` — compute the view's hint value: a "configured" boolean plus, per D-2's scope change, at most the final four characters of the stored key.

The hint derivation is the security-sensitive part of this milestone and
belongs in one function, not inline in the template:

- For a stored key longer than four characters, the hint is exactly `key[len(key)-4:]`. Nothing else about the key crosses into the view model — not its length, not a prefix, not a middle run.
- For a stored key of four characters or fewer, the hint is empty and only the "configured" boolean is surfaced (REQ-GKI-004-004). A naive `key[len(key)-4:]` panics below length 4 and, worse, a naive "last N or the whole thing" fallback discloses the entire key. The guard is the requirement.
- The view model must carry the hint, never the key. If the full key is passed into the template and truncated there, every future template edit becomes a potential leak.

Requirements: REQ-GKI-001-001, REQ-GKI-001-002, REQ-GKI-004-001, REQ-GKI-004-002, REQ-GKI-004-004.

Exit: the field renders; the full key never appears in a response body; the
only disclosed run is the final four characters, and nothing is disclosed for
a key of four characters or fewer.

### M4 — Validate the submitted value

Files:

- `internal/web/validate.go` — trim, empty-is-preserve, and line-break rejection.
- `internal/web/validate_test.go` — table-driven cases including the newline defect from §A.3.

Requirements: REQ-GKI-005-001, REQ-GKI-005-002, REQ-GKI-005-003.

Exit: an empty submission preserves the stored key; a newline-bearing
submission is rejected per-field with the stored key unchanged.

### M5 — Wire the save outcome and failure surfacing

Files:

- `internal/web/handlers.go` — success banner, credential-write failure surfacing, and the no-key-submitted path that leaves the file alone.

Requirements: REQ-GKI-001-004, REQ-GKI-002-005, REQ-GKI-004-003, REQ-GKI-006-003.

Exit: a write failure is reported as a failure; no response or log line
carries key material.

### M6 — Anti-leak regression battery

Mechanical, and last: it encodes guarantees the earlier milestones establish.

Files:

- `internal/web/security_test.go` — extend the existing security suite: submit a sentinel key, then assert the full key is absent from every response body, from the profile store tree, from `.moai/config/sections/`, and from captured log output.
- Regression guard asserting the credential field is absent from `settings.AllFields()`. This is the test that keeps D-2's structural guarantee true: if a later change adds the credential to the schema, this fails rather than silently reopening every generic-loop leak site.
- Differential window guard for the trailing-four hint (AC-GKI-004-02 / AC-GKI-004-04), which fails if any run of the key other than its final four characters reaches a response body.

Requirements: covered by the AC battery in `acceptance.md`; no new requirements.

Exit: the battery fails if a future change routes the key into any YAML
surface, adds it to the schema, or widens the disclosed hint.

---

## §G. Anti-Patterns

- **AP-1 — Second writer.** Adding a `saveGLMKey` equivalent inside `internal/web` to dodge the import-cycle problem. Two writers means two mode policies and two escaping rules; the next divergence is silent.
- **AP-2 — Echoing the stored key into the form value.** Populating `value="..."` with the stored key so the field "round-trips" like every other field. Every other field is not a secret. The empty-is-preserve convention exists precisely so a secret field never needs to be re-rendered. The trailing-four hint is a separate read-only display element, never the input's value.
- **AP-2b — Truncating the key inside the template.** Passing the full key into the view model and slicing it in the template. The hint must be computed once, before the view model is built, so the full key never reaches rendering scope. A template that receives the key is one careless edit away from printing it.
- **AP-2c — "Last four, or the whole key if shorter".** The obvious fallback for a short key discloses it entirely, which is the exact inverse of the requirement. REQ-GKI-004-004 exists because this is the natural thing to write.
- **AP-3 — Extending the dotenv escaper to survive newlines.** A newline in an API key is malformed input. Escaping it faithfully preserves a corrupt value; rejecting it at the boundary is correct.
- **AP-4 — Trusting `os.WriteFile`'s perm argument on an existing file.** It applies at creation only. A test asserting `0600` on a freshly created temp file passes while the real-world `0644` case silently fails.
- **AP-5 — Asserting absence by grepping filenames.** The anti-leak criteria must scan file contents. A key written into `llm.yaml` leaves the filename unchanged.
- **AP-6 — Testing against the developer's real home directory.** Use the injectable home-dir seam with `t.TempDir()`; a test that writes the real `~/.moai/.env.glm` destroys the developer's working credential.

---

## §H. Cross-References

- `spec.md` §A.3 — the persistence-route hazard table this plan implements against.
- `spec.md` §A.4 — the package-boundary constraint that makes M1 mandatory.
- `acceptance.md` — the mechanical criteria each milestone is measured by.
- `internal/web/validate.go:23` — documented precedent for the one-way package dependency.
- `CLAUDE.local.md` §6 — Go test isolation rules (`t.TempDir()`, no `t.Setenv("HOME")` in parallel tests).
- `CLAUDE.local.md` §14 — hardcoding prevention; the credential path belongs behind a resolver, not inline.
