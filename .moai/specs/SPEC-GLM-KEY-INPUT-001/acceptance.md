---
id: SPEC-GLM-KEY-INPUT-001
title: "Acceptance Criteria — GLM API Key Input Surface in the moai web Settings Console"
version: "0.2.0"
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

# Acceptance Criteria

## §A. Conventions

- **Sentinel key.** Every criterion that writes a key uses the literal
  `NOT-A-REAL-KEY-glm-acceptance-sentinel`. It is distinctive enough that a content grep
  produces no false positives anywhere in the tree, which is what makes the
  absence assertions meaningful rather than vacuous.
- **Sandboxed home.** Credential tests resolve `~/.moai/` through the
  injectable home-dir seam pointed at `t.TempDir()`. No criterion may read or
  write the developer's real `~/.moai/.env.glm`.
- **Content, not filename.** Absence criteria grep file *contents*
  recursively. A key leaked into `llm.yaml` changes no filename.
- **Sentinel trailing four.** The sentinel is 38 characters and its final four
  are `inel`. These are the only characters REQ-GKI-004-002 permits a response
  body to disclose. Short-key criteria derive their keys from the sentinel's
  tail (`inel`, `nel`) rather than introducing another key-shaped literal.
- **Differential window scan.** Absence-of-disclosure criteria compare two
  renders of the same page — one with a key stored, one without — and treat as
  leaked only those substrings of the key that appear in the with-key render
  and not in the without-key render. This matters: the sentinel contains
  ordinary English fragments (`ccept`, `ance`, `sent`) that also occur in
  normal page vocabulary, and the console already renders the string
  `acceptEdits` as a permission-mode option. A naive substring scan would
  report those as leaks and the criterion would be quietly disabled to make it
  pass. The differential removes every such false positive by construction,
  because shared page vocabulary appears in both renders.
- **Test names are illustrative.** The run phase may name tests differently;
  what is binding is the assertion each criterion describes and the command
  that produces its evidence.
- **Non-vacuity obligation.** Every absence criterion (`AC-GKI-003-01`,
  `-03-02`, `-04-01`, `-04-02`, `-04-03`) must be shown to fail when the
  guarantee is removed. A grep that returns zero matches because the sentinel
  was never written proves nothing. §D records the required self-trip.

---

## §B. Criteria

### AC-GKI-001-01 — key field renders in the 3rd Party LLM section

Maps REQ-GKI-001-001.

Given the console is served, When the settings page is fetched, Then the
response contains a GLM API key control inside the 3rd Party LLM section.

```bash
go test ./internal/web/ -run 'TestGLMKeyField_Renders' -count=1 -v
```

Assertion: the rendered HTML contains an input whose `name` is the GLM API key
field name, and that input appears after the `SectionLLM` fieldset opening
marker and before the next fieldset marker.

### AC-GKI-001-02 — control is a masked secret input with autofill disabled

Maps REQ-GKI-001-002.

Given the settings page is rendered, When the GLM key control is inspected,
Then it carries `type="password"` and `autocomplete="off"`.

```bash
go test ./internal/web/ -run 'TestGLMKeyField_SecretClass' -count=1 -v
```

Assertion: the control's tag carries both attributes. A `type="text"` control
fails this criterion.

### AC-GKI-001-03 — a submitted key reaches the credential file

Maps REQ-GKI-001-003.

Given a sandboxed home, When a save is POSTed with the sentinel key, Then
`<home>/.moai/.env.glm` exists and reading it back yields the sentinel.

```bash
go test ./internal/web/ -run 'TestGLMKeySave_Persists' -count=1 -v
```

Assertion: the credential reader returns exactly `NOT-A-REAL-KEY-glm-acceptance-sentinel`
after the POST.

### AC-GKI-001-04 — success banner confirms without echoing

Maps REQ-GKI-001-004.

Given a successful key save, When the response is rendered, Then the banner
reports success and the response body does not contain the sentinel.

```bash
go test ./internal/web/ -run 'TestGLMKeySave_BannerNoEcho' -count=1 -v
```

Assertion: `strings.Contains(body, "NOT-A-REAL-KEY-glm-acceptance-sentinel") == false`
and the banner kind is the success kind.

### AC-GKI-002-01 — credential file is the only file holding the key

Maps REQ-GKI-002-001.

Given a sandboxed home and a sandboxed project root, When the sentinel is
saved, Then a recursive content grep across both trees matches exactly one
file, and that file is `<home>/.moai/.env.glm` in the expected dotenv form.

```bash
go test ./internal/web/ -run 'TestGLMKeySave_SoleLocation' -count=1 -v
```

Assertion: the walk over both trees collects exactly one path whose contents
contain the sentinel; that path ends in `.moai/.env.glm`; and its contents
match `GLM_API_KEY="NOT-A-REAL-KEY-glm-acceptance-sentinel"`.

### AC-GKI-002-02 — credential file mode is exactly 0600

Maps REQ-GKI-002-002. **Mechanical mode assertion.**

Given a sandboxed home, When the sentinel is saved through the console path,
Then `os.Stat` on the credential file reports permission bits exactly `0600`.

```bash
go test ./internal/glmcred/ ./internal/web/ -run 'TestCredentialFileMode|TestGLMKeySave_Mode' -count=1 -v
```

Assertion, verbatim shape:

```go
fi, err := os.Stat(credPath)
require.NoError(t, err)
require.Equal(t, os.FileMode(0o600), fi.Mode().Perm())
```

`Mode().Perm()` masks off type and setuid bits, so the comparison is exact
rather than a bitwise-superset check. An equivalent shell confirmation on the
sandbox path:

```bash
# macOS
stat -f '%Lp' "$CRED_PATH"   # expect: 600
# Linux
stat -c '%a' "$CRED_PATH"    # expect: 600
```

### AC-GKI-002-03 — exactly one credential-write implementation exists

Maps REQ-GKI-002-003.

Given the post-implementation tree, When credential-write implementations are
enumerated, Then exactly one function writes the credential file and both the
CLI and console paths call it.

```bash
# Exactly one write of the credential file across the tree.
grep -rn "\.env\.glm" internal/ --include='*.go' | grep -v '_test\.go'

# No WriteFile of a GLM credential outside the owning package.
grep -rn "os\.WriteFile" internal/cli/ internal/web/ --include='*.go' \
  | grep -i 'glm' | grep -v '_test\.go'
```

Assertion: the first command shows the credential path constructed in exactly
one non-test location (the extracted package). The second command returns no
matches — neither `internal/cli` nor `internal/web` writes the credential
file directly.

### AC-GKI-002-04 — validation failure elsewhere leaves the credential untouched

Maps REQ-GKI-002-004.

Given a credential file holding a prior key, When a save is POSTed carrying
the sentinel key *and* an invalid sibling field, Then the response is a
validation rejection and the credential file's contents and modification time
are unchanged.

```bash
go test ./internal/web/ -run 'TestGLMKeySave_AtomicRejectPreservesCredential' -count=1 -v
```

Assertion: response status is the rejection status; the credential file still
holds the prior key; `ModTime()` is byte-identical to the pre-POST snapshot.

### AC-GKI-002-05 — a write failure is reported as a failure

Maps REQ-GKI-002-005.

Given the credential write is forced to fail (unwritable credential
directory), When a key save is POSTed, Then the response reports failure and
does not render the success banner.

```bash
go test ./internal/web/ -run 'TestGLMKeySave_WriteFailureSurfaced' -count=1 -v
```

Assertion: the banner kind is the error kind; the success banner string is
absent from the body.

### AC-GKI-003-01 — no key material reaches the profile store

Maps REQ-GKI-003-001. **Mechanical profile-store absence assertion.**

Given a sandboxed home containing an initialised profile store, When the
sentinel key is saved through the console, Then a recursive content scan of
`<home>/.moai/claude-profiles/` matches zero files.

```bash
go test ./internal/web/ -run 'TestGLMKeySave_NoProfileStoreLeak' -count=1 -v
```

Assertion, verbatim shape:

```go
var hits []string
_ = filepath.WalkDir(profileStoreRoot, func(p string, d fs.DirEntry, err error) error {
    if err != nil || d.IsDir() {
        return nil
    }
    b, readErr := os.ReadFile(p)
    if readErr == nil && bytes.Contains(b, []byte(sentinelKey)) {
        hits = append(hits, p)
    }
    return nil
})
require.Empty(t, hits, "GLM key material leaked into the profile store: %v", hits)
```

The walk must run after a save that also wrote at least one legitimate
profile-store field, so the tree is non-empty and the assertion is not passing
merely because there were no files to scan. That co-write is part of the
criterion, not an incidental detail.

Shell confirmation on the sandbox path:

```bash
grep -rl 'NOT-A-REAL-KEY-glm-acceptance-sentinel' "$SANDBOX_HOME/.moai/claude-profiles/"
# expect: no output, exit status 1
```

### AC-GKI-003-02 — no key material reaches project config sections

Maps REQ-GKI-003-002.

Given a sandboxed project root with populated config sections, When the
sentinel key is saved, Then a recursive content scan of
`<root>/.moai/config/sections/` matches zero files — `llm.yaml` in particular.

```bash
go test ./internal/web/ -run 'TestGLMKeySave_NoConfigSectionLeak' -count=1 -v
```

Assertion: the same walk shape as AC-GKI-003-01, rooted at the config sections
directory, with an explicit additional read of `llm.yaml` asserting the
sentinel is absent from it. The save under test must also carry a legitimate
`glm.models.high` edit, so `llm.yaml` is provably written during the same
request and the absence result is not an artefact of the file never being
touched.

Shell confirmation:

```bash
grep -rl 'NOT-A-REAL-KEY-glm-acceptance-sentinel' "$SANDBOX_ROOT/.moai/config/sections/"
# expect: no output, exit status 1
```

### AC-GKI-003-03 — credential field is absent from the bulk value read

Maps REQ-GKI-003-003.

Given a stored key exists, When the bulk schema current-value read runs, Then
its result contains no entry whose value is the sentinel.

```bash
go test ./internal/settings/ -run 'TestSchemaCurrentValues_NoCredential' -count=1 -v
```

Assertion, per the resolved D-2 out-of-schema route: no `FieldDef` returned by
`settings.AllFields()` carries the credential field name, and no value returned
by `SchemaCurrentValues` equals the sentinel. The first half is the load-bearing
one — it asserts the structural guarantee that the credential is not a member
of the collection every generic loop iterates, so the second half cannot
regress without the first failing too.

### AC-GKI-004-01 — no response body carries the full stored key

Maps REQ-GKI-004-001.

Given a credential file holding the sentinel, When every console GET route is
fetched, Then no response body contains the full sentinel.

```bash
go test ./internal/web/ -run 'TestGLMKey_NoLeakAcrossRoutes' -count=1 -v
```

Assertion: the test enumerates the registered GET routes (`/`, `/specs`, and
any others registered on the mux in `internal/web/app.go`) and asserts
`strings.Contains(body, sentinel) == false` for each. Enumerating from the mux
rather than a hand-written list keeps the criterion honest as routes are added.
This criterion is unaffected by the trailing-four hint: the hint is four
characters, the full 38-character sentinel must never appear.

### AC-GKI-004-02 — only the trailing four characters are disclosed

Maps REQ-GKI-004-002.

Given a credential file holding the sentinel (38 characters), When the settings
page is rendered, Then the hint `inel` is present, and a differential window
scan finds no other run of the stored key in the body — no prefix, and no run
of five or more characters anywhere.

```bash
go test ./internal/web/ -run 'TestGLMKeyField_HintDisclosesOnlyTrailingFour' -count=1 -v
```

Assertion, verbatim shape:

```go
withKey := renderSettingsPage(t, app)      // sentinel stored
withoutKey := renderSettingsPage(t, appNoKey) // no credential file

hint := key[len(key)-4:] // "inel"
require.Contains(t, withKey, hint, "trailing-four hint must be rendered")
require.NotContains(t, withoutKey, hint, "hint must appear only when a key is stored")

leaked := map[string]bool{}
for l := 4; l <= len(key); l++ {
    for i := 0; i+l <= len(key); i++ {
        w := key[i : i+l]
        if strings.Contains(withKey, w) && !strings.Contains(withoutKey, w) {
            leaked[w] = true
        }
    }
}
delete(leaked, hint) // the ONLY sanctioned disclosure
require.Empty(t, leaked, "disclosed runs beyond the trailing four: %v", leaked)
```

Why this shape rather than a plain sliding-window absence check:

- It **permits exactly one** disclosure — the trailing four — and nothing else. Any prefix (`NOT-`, `NOT-A`, …), any middle run, and any run of length five or more that reaches the body lands in `leaked` and fails the criterion.
- It is **non-vacuous in both directions**. `require.Contains(withKey, hint)` fails an implementation that renders no hint at all, so the criterion cannot pass by simply omitting the feature. `require.NotContains(withoutKey, hint)` proves the hint is genuinely key-derived rather than a static string that happens to match.
- It is **false-positive-free**. The differential against the no-key render cancels every fragment that belongs to ordinary page vocabulary, including the `ccept` that `acceptEdits` would otherwise contribute. Without the differential this criterion would fail for reasons unrelated to the key and would end up weakened until it passed.

### AC-GKI-004-03 — no key material in logs or error renderings

Maps REQ-GKI-004-003.

Given a credential file holding the sentinel, When a save is forced to fail
and server log output is captured, Then neither the captured log nor the
rendered error page contains the sentinel.

```bash
go test ./internal/web/ -run 'TestGLMKey_NoLeakInLogsOrErrors' -count=1 -v
```

Assertion: the test installs a capturing writer as the log sink for the
request, triggers the failure path from AC-GKI-002-05, and asserts the
sentinel is absent from both the captured buffer and the response body.

### AC-GKI-004-04 — a key of four characters or fewer discloses nothing

Maps REQ-GKI-004-004.

Given a credential file holding a key of exactly four characters (`inel`, the
sentinel's tail) and again a key of three (`nel`), When the settings page is
rendered, Then the presence indicator is shown and the differential window scan
finds zero disclosed runs — with no permitted exception.

```bash
go test ./internal/web/ -run 'TestGLMKeyField_ShortKeyDisclosesNothing' -count=1 -v
```

Assertion: table-driven over `"inel"` (4), `"nel"` (3), and `"l"` (1). For each,
the presence-indicator marker is present in the body, and the differential
window scan from AC-GKI-004-02 is run with the window floor lowered to 1 and
**no `delete(leaked, hint)` exception granted**. `require.Empty(t, leaked)` must
hold.

The removed exception is the whole point of this criterion. The natural
implementation of the hint — "show the last four characters, or the whole key
if it is shorter" — passes AC-GKI-004-02 and discloses a short key in full.
Only an assertion that grants no exception catches it.

### AC-GKI-005-01 — empty submission preserves the stored key

Maps REQ-GKI-005-001.

Given a credential file holding the sentinel, When a save is POSTed with the
key field empty (and again with whitespace only), Then the stored key is
unchanged and the file's modification time is unchanged.

```bash
go test ./internal/web/ -run 'TestGLMKeySave_EmptyPreserves' -count=1 -v
```

Assertion: table-driven over `""`, `" "`, `"\t"`, `"  \t "`. For each, the
reader still returns the sentinel and `ModTime()` matches the pre-POST
snapshot. The mtime check is what proves the file was not rewritten with
identical content.

### AC-GKI-005-02 — surrounding whitespace is trimmed

Maps REQ-GKI-005-002.

Given a sandboxed home, When `"  NOT-A-REAL-KEY-glm-acceptance-sentinel\t"` is
submitted, Then the stored key reads back as the untrimmed-free sentinel.

```bash
go test ./internal/web/ -run 'TestGLMKeySave_Trims' -count=1 -v
```

Assertion: the reader returns exactly `NOT-A-REAL-KEY-glm-acceptance-sentinel` with no
leading or trailing whitespace.

### AC-GKI-005-03 — line-break input is rejected per-field

Maps REQ-GKI-005-003.

Given a credential file holding a prior key, When a value containing `\n` or
`\r` is submitted, Then the response carries a per-field validation error for
the GLM key field and the stored key is unchanged.

```bash
go test ./internal/web/ -run 'TestGLMKeySave_RejectsLineBreak' -count=1 -v
```

Assertion: table-driven over `"abc\ndef"`, `"abc\r\ndef"`, `"abc\rdef"`, and a
trailing `"abc\n"`. For each: the response `FieldErrors` map carries an entry
keyed by the GLM key field name, and the credential reader still returns the
prior key. This criterion is the regression guard for the newline-corruption
defect recorded in `plan.md` §A.3 — without it, a newline is written raw into
the quoted dotenv value and the reader silently returns a truncated key.

### AC-GKI-006-01 — an existing credential file is replaced in place

Maps REQ-GKI-006-001.

Given a credential file already holding a prior key, When a new key is saved,
Then the file holds exactly one `GLM_API_KEY` entry and its value is the new
key.

```bash
go test ./internal/glmcred/ -run 'TestSave_ReplacesInPlace' -count=1 -v
```

Assertion: `strings.Count(contents, "GLM_API_KEY=") == 1`, the parsed value is
the new key, the prior key is absent from the contents, and no sibling
credential file was created alongside it.

### AC-GKI-006-02 — a pre-existing wide-mode file is tightened to 0600

Maps REQ-GKI-006-002. **Mechanical mode assertion on the pre-existing-file path.**

Given a credential file that already exists at mode `0644`, When a key is
written **from the console save path**, and again when a key is written **from
the `moai glm` interactive save path**, Then the file's permission bits
afterwards are exactly `0600` in both cases.

Per the resolved D-3 decision the tightening reaches both surfaces, so both are
asserted. Testing only the console path would leave the CLI half of the shared
writer unverified.

```bash
go test ./internal/glmcred/ -run 'TestSave_TightensExistingMode' -count=1 -v
go test ./internal/cli/ -run 'TestGLMKeySave_TightensExistingMode' -count=1 -v
go test ./internal/web/ -run 'TestGLMKeySave_TightensExistingMode' -count=1 -v
```

Assertion, verbatim shape (applied at each of the three entry points):

```go
require.NoError(t, os.WriteFile(credPath, []byte("GLM_API_KEY=\"old\"\n"), 0o644))
require.NoError(t, os.Chmod(credPath, 0o644)) // defeat umask interference
require.NoError(t, saveUnderTest(sentinelKey)) // glmcred.Save / CLI path / console path
fi, err := os.Stat(credPath)
require.NoError(t, err)
require.Equal(t, os.FileMode(0o600), fi.Mode().Perm())
```

The explicit `os.Chmod` after creation matters: the process umask can narrow
the `0o644` argument at creation time, which would make the test pass without
the tightening logic ever running. Forcing the mode to `0644` first is what
makes this criterion a real test of REQ-GKI-006-002 rather than a
umask-dependent accident.

### AC-GKI-006-03 — no-key save leaves the file intact

Maps REQ-GKI-006-003.

Given a credential file holding the sentinel, When a settings save is POSTed
that changes an unrelated field and submits no key, Then the credential file
still exists with unchanged contents.

```bash
go test ./internal/web/ -run 'TestGLMKeySave_NoKeyLeavesFileIntact' -count=1 -v
```

Assertion: the file exists; its bytes are identical to the pre-POST snapshot;
and the unrelated field's change did land, proving the save actually executed
rather than being rejected before reaching persistence.

---

## §C. Quality Gates

- Full suite green: `go test ./internal/cli/... ./internal/web/... ./internal/settings/... ./internal/glmcred/...`
- Build and vet clean: `go build ./...` and `go vet ./...`
- Lint clean: `golangci-lint run --timeout=2m`
- SPEC lint clean: `go run ./cmd/moai spec lint .moai/specs/SPEC-GLM-KEY-INPUT-001/spec.md`
- No template surface touched: `git diff --name-only` shows no entry under `internal/template/templates/`. A hit there means the credential is being persisted in a distributed file, which contradicts REQ-GKI-002-001.

---

## §D. Non-Vacuity Self-Trip

Each absence criterion must be demonstrated to fail when its guarantee is
removed. Absence assertions are the criteria most likely to pass for the wrong
reason, so this section is a completion obligation, not an optional extra.

| Criterion | Temporary mutation that must turn it red |
|-----------|------------------------------------------|
| AC-GKI-003-01 | Route the key through `PersistProfileStore` instead of the credential writer |
| AC-GKI-003-02 | Route the key through `PersistTypedSection` into `llm.yaml` |
| AC-GKI-003-03 | Add the credential as a `FieldDef` in `AllFields()` |
| AC-GKI-004-01 | Populate the form control's `value` attribute with the stored key |
| AC-GKI-004-02 (widen) | Widen the hint from the final 4 characters to the final 8 |
| AC-GKI-004-02 (move) | Move the hint from the final 4 characters to the leading 4 |
| AC-GKI-004-02 (omit) | Render no hint at all — must fail the `require.Contains` positive control |
| AC-GKI-004-03 | Log the submitted key at the failure path |
| AC-GKI-004-04 | Change the hint derivation to "last 4, or the whole key if shorter" |
| AC-GKI-006-02 | Remove the explicit mode assertion, leaving only `os.WriteFile`'s `perm` argument |

The three AC-GKI-004-02 rows are listed separately because they falsify
different halves of that criterion. *Widen* and *move* must be caught by the
`leaked` set; *omit* must be caught by the `require.Contains` positive control.
A criterion that catches only the first two would pass an implementation that
silently dropped the feature, and one that catches only the third would pass an
implementation that disclosed half the key.

Record the observed failure output for each. A mutation that leaves its
criterion green means the criterion is not testing what it claims, and the
criterion must be strengthened before the run phase closes.

---

## §E. Definition of Done

- All 22 criteria in §B report PASS with quoted command output.
- All 10 mutations in §D were applied, observed red, and reverted.
- All gates in §C pass.
- Exactly one credential-write implementation exists in the tree, in `internal/glmcred`, and `moai glm` still saves a key through it.
- The credential field is absent from `settings.AllFields()`, preserving the D-2 structural guarantee.
- The trailing-four hint renders for a long key and discloses nothing for a key of four characters or fewer.
- The three plan-phase decisions recorded in `plan.md` §A.1 (D-1 `internal/glmcred`, D-2 out-of-schema, D-3 both-surface mode tightening) were resolved before implementation began and are recorded in the SPEC HISTORY.
