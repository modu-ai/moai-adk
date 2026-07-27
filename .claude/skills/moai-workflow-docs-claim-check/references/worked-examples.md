# Worked Examples

Level-3 reference for `moai-workflow-docs-claim-check`. Each example runs a
document claim through Preflight, Claim Triage, and Validation, then shows the
three-section output.

The examples deliberately span several language ecosystems — Python,
JavaScript, Rust, and a language-neutral CLI — because the skill ships to
projects in any supported language. Read the one nearest your subject; the
procedure is identical across all of them.

---

## Example A — Python library README (composite decomposition)

**Document sentence.** "Works on Python 3.10+ across Linux and macOS, and
installs with a single `pip install` — no compiler required."

**Preflight.**

- Public-facing: yes, project README.
- Evidence supplied:
  - `pyproject.toml`, version `2.4.0`, timestamp available.
  - CI run summary, Linux job only, timestamp available.
- Secrets scan: clean.

**Claim Triage.** One sentence, four atomic claims plus one exclusion:

| # | Atomic claim | Origin |
|---|--------------|--------|
| A1 | Supports Python 3.10 and newer | conjunction split |
| A2 | Supports Linux | list member |
| A3 | Supports macOS | list member |
| A4 | Installs with a single `pip install` and needs no compiler | qualifier with its own truth value → split further |

A4 is still composite (install shape + no-compiler condition), so it splits into
A4a "installs with a single `pip install`" and A4b "requires no compiler".

**Validation.**

- A1 → `verified`. `pyproject.toml` declares `requires-python = ">=3.10"`.
- A2 → `verified`. CI run summary shows a passing Linux job.
- A3 → `unsupported` / `insufficient-coverage`. The CI evidence is on-topic
  (platform support) but covers only Linux.
- A4a → `needs-human`. Settling it requires running the install; name the
  command `pip install <package>` in a clean environment.
- A4b → `unsupported` / `missing-evidence`. Nothing supplied speaks to build
  requirements. A wheel listing would settle it.

**Lesson.** One marketing sentence produced five rows and three different
labels. Had it been assessed whole, a single verdict would have hidden the
macOS gap entirely.

---

## Example B — JavaScript package release note (staleness)

**Document sentence.** "The latest release is 3.1.0 and ships 12 built-in
plugins."

**Preflight.**

- Public-facing: yes, published release note.
- Evidence supplied:
  - Tag listing, newest entry `4.0.2`, timestamp available.
  - `plugins/` directory listing, 15 entries, timestamp available.
- Secrets scan: clean.

**Claim Triage.** Two atomic claims: B1 "latest release is 3.1.0", B2 "ships 12
built-in plugins".

**Validation.**

- B1 → `stale-suspected`. Gate 2 fires before gate 4: the claim was accurate
  when 3.1.0 was newest, and the evidence disagrees only on a version field.
  Record both values (document `3.1.0`, evidence `4.0.2`).
- B2 → `stale-suspected`. A count mismatch (12 vs 15) with a plausible earlier
  moment when 12 was correct.

**Lesson.** Neither row is `contradicted`. Contradiction would require evidence
that the claim never held. Reporting drift as contradiction misroutes a
maintenance task into a defect queue.

---

## Example C — Rust CLI install guide (coverage and contradiction)

**Document sentences.** "Prebuilt binaries are available for every major
platform. The `--offline` flag is enabled by default."

**Preflight.**

- Public-facing: yes, install guide page.
- Evidence supplied:
  - Release artifact listing: two artifacts, `linux-x86_64` and `darwin-arm64`.
  - CLI argument definition source showing `offline` with `default_value = "false"`.
- Secrets scan: clean.

**Claim Triage.** C1 "prebuilt binaries exist for every major platform",
C2 "`--offline` is enabled by default".

**Validation.**

- C1 → `unsupported` / `insufficient-coverage`. The artifact listing is exactly
  on-topic but narrower than "every major platform" — Windows is absent. Anchor
  the two artifacts present and name the gap.
- C2 → `unsupported` / `contradicted`. The argument definition asserts the
  opposite. No earlier moment is visible in the evidence, so gate 2 does not
  fire and the reason is contradiction rather than staleness.

**Lesson.** Two `unsupported` rows, two different reasons. The reason field is
what makes the row actionable: one needs a build-matrix change, the other needs
a one-word doc correction.

---

## Example D — Language-neutral CLI quickstart (subjective exclusion)

**Document sentence.** "Blazing-fast setup: one command, and the daemon starts
automatically on boot."

**Preflight.**

- Public-facing: yes, quickstart page.
- Evidence supplied: none.
- Secrets scan: not applicable, no evidence supplied.

**Claim Triage.**

- Excluded as subjective: "blazing-fast" — a speed adjective with no stated
  threshold; nothing could confirm or refute it.
- D1 "setup is one command".
- D2 "the daemon starts automatically on boot".

**Validation.** With zero evidence supplied, no claim can reach `verified`.

- D1 → `unsupported` / `missing-evidence`. Name what would settle it: the
  install section of the packaging manifest, or a recorded setup transcript.
- D2 → `needs-human`. Boot behavior requires exercising a machine restart; no
  document artifact settles it. Name the service-unit definition as the closest
  readable proxy and the reboot check as the human step.

**Lesson.** An empty evidence set is a valid input and must never produce a
fabricated pass. It produces a fully populated report whose value is the named
list of evidence the maintainer still has to supply.

---

## Composite output — Example A rendered

The three-section contract as it appears for Example A:

### 1. Input Scope Reviewed

- Document: project README, version not stated in document.
- Evidence: `pyproject.toml` (version `2.4.0`, timestamped); CI run summary
  (Linux job only, timestamped).
- Excluded as subjective: none in this sentence.
- Needed but not supplied: a macOS CI job record; a clean-environment install
  transcript; a wheel or build-requirement listing.

### 2. Claim Assessments

| # | Atomic claim | Label | Reason | Evidence anchor or what is missing |
|---|--------------|-------|--------|------------------------------------|
| A1 | Supports Python 3.10+ | `verified` | — | `pyproject.toml`, `requires-python = ">=3.10"` |
| A2 | Supports Linux | `verified` | — | CI run summary, Linux job passing |
| A3 | Supports macOS | `unsupported` | `insufficient-coverage` | CI evidence covers Linux only; macOS job record missing |
| A4a | Single-command `pip install` | `needs-human` | — | Run `pip install <package>` in a clean environment |
| A4b | Requires no compiler | `unsupported` | `missing-evidence` | No build-requirement evidence supplied; a wheel listing would settle it |

### 3. Boundary Notes

- Certification: **no commands executed** during this assessment.
- Declined: nothing requested beyond the assessment.
- Residual risk: the two `verified` rows rest on a manifest declaration and a
  single CI job. Neither proves runtime behavior on an end-user machine, and
  the CI job reflects one runner image rather than the platform generally.
- Human steps: for A4a, run `pip install <package>` in a clean environment and
  record the transcript.

---

## Cross-cutting lessons

1. **Split before labeling.** Every example above produced more rows than the
   document produced sentences. That expansion is the point.
2. **A named anchor is part of the label.** A `verified` row without an anchor
   is an assertion, not a finding.
3. **Drift and error are different labels.** Version and count mismatches are
   `stale-suspected`; only an outright reversal is `contradicted`.
4. **Adjacent is not covering.** Evidence on the right topic but narrower than
   the claim is `insufficient-coverage`, never a hedged pass.
5. **No evidence means no pass.** An empty evidence set yields a complete report
   of what is missing — never a fabricated `verified`.
