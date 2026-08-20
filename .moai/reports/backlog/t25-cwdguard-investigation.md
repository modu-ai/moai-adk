# t25 — The worktree guard that rejects `env -u VAR …` wrappers

Read-only investigation. No behaviour was changed.

---

## Claim

1. **The guard is not MoAI code.** It lives in the Claude Code binary, not in this
   repository. `grep` across `internal/`, `.claude/hooks/`, and `.claude/rules/`
   finds no `env`-wrapper analysis and no matching message string. The card's name
   "CwdGuard" does not correspond to any MoAI symbol — the only `CwdGuard` identifier
   in the tree is `TestCwdGuard_DeletedDirectory` in `internal/cli/statusline_test.go:283`,
   an unrelated statusline test. Internally Claude Code calls the feature the
   **agent worktree isolation guard**, and emits the telemetry event
   `tengu_agent_worktree_cwd_escape_blocked`.

2. **It is inert outside a worktree-isolated session.** The guard only runs when the
   session or agent carries an isolation-worktree root, and only for the `bash` shell.
   In the primary checkout it never executes.

3. **Its purpose is git containment, not env hygiene.** The refusal message states it
   directly: *"… git operations must target its own worktree."* The hazard defended
   against is a worktree-isolated agent redirecting a git write into the **shared
   checkout** — via `git -C`, `--git-dir`, `--work-tree`, `--bare`, `GIT_DIR`/
   `GIT_WORK_TREE`/`GIT_CONFIG*`/`HOME`/`CDPATH`, a `cd` computed at runtime, or a
   payload hidden inside a wrapper.

4. **`env -u VAR cmd` trips it through an argument-boundary misparse — not because
   `-u` is forbidden.** `-u` is explicitly modelled and accepted. The failure is that
   when `argv[0]` is `env`, the guard scans *the entire remaining argv* as if it were
   env's own flags, because it has no notion of where env's options end and the wrapped
   command begins. Any dashed flag belonging to the **inner** command that is not one
   of `-i -v -C -u` is then classified as an unmodelled env flag, and the command is
   refused. The reported PR #1527 message named `-run` — that is `go test -run …`, the
   inner command's flag, misattributed to `env`.

5. **`unset VAR && cmd` passes because `unset` is not in the wrapper set.** The scan
   window is only opened for `argv[0] ∈ {export, declare, typeset, local, readonly,
   env, make}`. `unset` is absent, so the window closes at index 0 and no flag is
   inspected. `&&` lists are modelled as multiple simple commands, so the compound form
   survives intact. (`make` **is** in the set — `make <target> --some-flag` is exposed
   to the same misparse.)

6. **The guard's env-wrapper and shell-wrapper refusals are not gated on git being
   present.** The `envUnmodeled` check and the "runs a string through sh/bash" check
   both fire unconditionally per command, while the `-C`/`--git-dir`/`GIT_*` checks are
   gated on a git binary appearing in argv. So `env -u FOO go test -run X ./...`, which
   contains no git at all, is refused by a guard whose stated purpose is git containment.
   This is the over-broad edge the card is feeling.

All six claims are **code-derived** from a disassembled release binary, except where
the Evidence section records a probe.

---

## Evidence (verbatim)

### E1 — No MoAI implementation

```
$ grep -rln "CwdGuard\|cwd_guard\|cwdGuard\|cwd-guard" internal/ pkg/ cmd/ .claude/
internal/cli/statusline_test.go
… (remaining hits are copies of the same file inside .claude/worktrees/*)

$ grep -n "CwdGuard" internal/cli/statusline_test.go
278:// TestCwdGuard_DeletedDirectory tests AC-SF-006: Cwd Guard for deleted directories.
283:func TestCwdGuard_DeletedDirectory(t *testing.T) {
```

No `env`-wrapper logic exists in `internal/hook/pre_tool.go` (the MoAI PreToolUse
handler) or in `.claude/settings.json` permissions. `.claude/settings.json` sets
`"defaultMode": "bypassPermissions"` and carries no `env` deny entry.

### E2 — The message string is in the Claude Code binary

```
$ strings -n 20 /Users/goos/.local/share/claude/versions/2.1.233 \
    | grep -iE "Refusing to run|whose effect on the command"
, whose effect on the command it wraps can't be verified
). Refusing to run it. Retry the command; if this keeps failing, report that worktree isolation was lost.
```

### E3 — The guard body (function `hmp`), abridged to the load-bearing lines

```js
function hmp(e,t,r){
  let {noun:n, possessive:o} = CDn(r),
      i = (h) => `${n} is isolated in the worktree ${r}, but this command ${h}. `
                + `Refusing to run it — ${o} git operations must target its own worktree. `
                + `Run the equivalent from ${r} without the redirect.`;
  if (e.kind !== "simple")
    return i("is too complex to verify that it stays inside the worktree; break it into plain, separate commands");
  let s = e.commands,
      a = s.map(h => ymp(h.argv));          // index of a git binary in each command, or -1
  …
  for (let [h,g] of s.entries()) {
    …
    let S = pmp(g);
    if ("opaque" in S)          return i(`sets ${S.opaque}, injecting git configuration whose effect on where git writes can't be verified`);
    if (S.envUnmodeled !== null) return i(`runs env with ${S.envUnmodeled}, whose effect on the command it wraps can't be verified`);
    …
```

Note that neither the `opaque` nor the `envUnmodeled` return is guarded by
`a[h] !== -1` (git present) — unlike every `-C` / `--git-dir` / pin check below them.

### E4 — Why `env` swallows the inner command's flags (function `pmp`)

```js
function pmp(e){
  let t = ymp(e.argv),                                          // git index, or -1
      r = Hjb.has(rie.basename(e.argv[0] ?? "").toLowerCase()), // argv[0] is a wrapper
      n = t !== -1 ? t : (r ? e.argv.length : 0),               // ← scan window END
      …
      s = e.argv.slice(1, n),
      a = rie.basename(e.argv[0] ?? "").toLowerCase() === "env",
      l = null;                                                 // ← envUnmodeled
  for (let [d,p] of s.entries()) {
    if (rie.basename(p).toLowerCase() === "env") { a = true; continue }
    let f = Fjb.exec(p);                                        // NAME=value operand
    if (f) { o.push({name:f[1], value:f[2], fromOperand:true}); continue }
    if (a) {
      if (p === "--chdir")                { … i.push(next) }
      else if (p.startsWith("--chdir="))  { i.push(p.slice(8)) }
      else if (p === "--ignore-environment") continue;
      else if (/^-[a-zA-Z]/.test(p)) {
        let m = p.slice(1), h = false;
        for (let g = 0; g < m.length; g++) {
          let y = m[g];
          if (y === "i" || y === "v") continue;                 // modelled, harmless
          if (y === "C" || y === "u") {                         // modelled: -C dir, -u NAME
            let b = m.slice(g+1), v = b.length > 0 ? b : s[d+1];
            if (y === "C" && v !== undefined) i.push(v);
            break;
          }
          h = true; break;                                      // ← anything else: UNMODELLED
        }
        if (h) l ??= p;                                         // ← records the offending flag
      }
      else if (p.startsWith("-")) l ??= p;
    }
  }
  return { assignments:u, envChdirs:i, envUnmodeled:l };
}
```

with the wrapper set:

```js
Hjb = new Set(["export","declare","typeset","local","readonly","env","make"]);
```

Reading the two together: for `env -u A -u B go test -run X ./...` there is no git in
argv, so `t === -1`; `argv[0]` is `env`, so `r === true` and the scan window
`n = argv.length` — **the whole command**. `-u A` and `-u B` are accepted. Then
`-run` reaches the `/^-[a-zA-Z]/` branch, its first character `r` is not one of
`i v C u`, so `h = true` and `envUnmodeled = "-run"`. The guard refuses with exactly
the message recorded in PR #1527.

### E5 — Gating: worktree-isolated bash only

```js
let D, U = g ?? h;                     // h = agentWorktree isolation root
if (U) {
  let ee = Dpp(L, U);                  // cwd escape check
  if (ee) { … return Yut(ee) }
  let ce = r === "bash" ? hmp(D = await Iat(e), L, U) : null;
  if (ce) {
    w(`[worktree] blocked shell exec: command redirects git into the shared checkout: cwd=${L} isolationRoot=${U}`, {level:"warn"}),
    H("tengu_agent_worktree_cwd_escape_blocked", {reason: Ae("command_redirect")}),
    …
    return Yut(ce)
  }
}
```

`hmp` is reached only when `U` (the isolation-worktree root) is truthy.

### E6 — Probe (this session, primary checkout)

```
$ env -u MOAI_KANBAN -u MOAI_KANBAN_ID echo PROBE_OK
PROBE_OK

$ env -u MOAI_KANBAN -u MOAI_KANBAN_ID -u CLAUDE_CODE_STOP_HOOK_BLOCK_CAP go version
go version go1.26.4 darwin/arm64
```

Both accepted. This confirms claim 2 (inert outside a worktree) but **does not**
reproduce the rejection: this session is not worktree-isolated, so `U` is falsy and
`hmp` never ran. Note also that neither probe carries an inner dashed flag, so even
inside a worktree both would have passed. **The rejection itself is code-derived and
corroborated by the PR #1527 field observation; it was not reproduced here.**

### E7 — Full refusal catalogue (for reference)

Only inside a worktree-isolated bash session. Ordered as the guard evaluates them:

| Shape | Gated on git present? |
|---|---|
| command kind is not `simple` (pipelines, substitutions) | no |
| command name spelled as an unquoted glob | no |
| a string run through `sh`/`bash`/eval-like wrapper | no |
| `xargs` / `parallel` feeding git its arguments | yes |
| `find -execdir` / `-okdir` | yes |
| `env` with an unmodelled flag | **no** ← this card |
| sets `GIT_CONFIG*`, `HOME`, `CDPATH`, `XDG_CONFIG_HOME` | no |
| assigns `GIT_DIR`/`GIT_WORK_TREE`/`GIT_COMMON_DIR`/`GIT_OBJECT_DIRECTORY`/`GIT_INDEX_FILE`/`GIT_SHALLOW_FILE` twice | no |
| git named more than once in one command | yes |
| more than one `-C`/`--chdir` passed to `env` | yes |
| `git -C` / `--git-dir` / `--work-tree` / `--bare` resolving into the shared checkout | yes |
| `cd`/`pushd` to a runtime-computed path before git | yes |

---

## Baseline-attribution

| Item | Value |
|---|---|
| Repository | `/Users/goos/MoAI/moai-adk-go`, branch `main`, HEAD `3b9b3bf99` |
| Claude Code binary | `/Users/goos/.local/share/claude/versions/2.1.233` |
| Binary sha256 (first 24) | `bc466b6cde63edafc773f471` |
| Extraction commands | `strings -n 20` / `strings -n 200` piped to `grep`, plus a byte-offset window read via `python3` |
| Probe environment | primary checkout, **not** worktree-isolated |
| Governing MoAI rule | none for this guard. The nearest MoAI-side rules are `.claude/rules/moai/workflow/main-checkout-branch-guard.md` and `internal/hook/branch_guard.go`, which are a **separate, MoAI-owned** guard defending the same hazard from the opposite direction (branch-state mutation *in* the primary checkout). They do not implement, configure, or interact with the Claude Code worktree guard. |

---

## Gaps

- **The rejection was not reproduced.** No worktree-isolated session was available to
  probe from, and creating one would have mutated repository state, which this
  investigation was scoped against. Claims 4-6 rest on the decompiled guard body plus
  the PR #1527 field report.
- **`e.kind === "simple"` was not characterised.** Whether `A && B`, `A; B`, and `A | B`
  each qualify as `simple` was inferred from the guard iterating `e.commands` as a list
  and from the field observation that the `unset … && …` form passes. The parser
  `Iat`/`M$o` was not read.
- **No version sweep.** Only `2.1.233` was inspected. Whether the `env` argument-boundary
  behaviour is new, and which version introduced it, is unknown.
- **No upstream issue search.** Whether this misparse is already reported to Anthropic
  was not checked.
- **`make` exposure untested.** Claim 5's note that `make` sits in the same wrapper set
  is read off `Hjb` and was not exercised.

---

## Residual-risk

- The guard is an **opaque upstream dependency**. Its rules can change in any Claude Code
  release without notice, so any option below that depends on the current parsing
  behaviour can silently break — or silently start passing — on upgrade.
- The `envUnmodeled` refusal is **not a security boundary in the usual sense**: it fires
  on commands containing no git at all, and it is trivially bypassed by moving the
  command into a script file (which the guard cannot inspect). Treating it as a hard
  safety property would overstate what it provides.
- Conversely, the git-redirect checks it *does* gate on git presence are load-bearing.
  Any option that routes around the guard wholesale gives up those too, not just the
  env-parsing annoyance.

---

## Options

None of these were implemented. Each names the hazard it reopens.

### Option A — Standardise on the shell-builtin `unset … && cmd` form (recommended)

Bake the exact form into every kanban delegation prompt that runs an env-isolated
verification:

```
unset MOAI_KANBAN MOAI_KANBAN_ID MOAI_KANBAN_LABEL MOAI_KANBAN_LEAD_ADDR \
      MOAI_KANBAN_SETTINGS_INJECTED CLAUDE_CODE_STOP_HOOK_BLOCK_CAP \
  && go test ./...
```

`unset` is outside the guard's wrapper set (E4, `Hjb`), so no flag scan opens and the
inner command's flags are never inspected. This is the only option that keeps every
guard check intact.

- **Hazard reopened:** none. The guard's full check set continues to apply.
- **Cost:** `unset` removes variables; it cannot *set* one to a different value for a
  single command. The prefix form `VAR=value cmd` covers that case and is modelled by
  the guard — but only for names outside `GIT_*`, `HOME`, `CDPATH`, `XDG_CONFIG_HOME`,
  which the guard refuses outright.
- **Caveat:** the `unset` and the command must be one compound invocation. A separate
  Bash call does not inherit the scrub — each Bash tool call is a fresh process.
- **Do not** wrap it in a subshell (`( unset A; cmd )`): that likely changes the parsed
  kind away from `simple` and triggers the "too complex to verify" refusal instead.

### Option B — Keep `env`, but require the wrapped command to carry no dashed flags

`env -u A -u B go test ./...` passes the guard today; `env -u A go test -run X ./...`
does not (E4). A delegation prompt could mandate the flag-free shape.

- **Hazard reopened:** none directly, but it relies on an **undocumented parsing
  accident**. The constraint is invisible at the call site: whoever later adds `-run`,
  `-count=1`, or `-race` re-breaks it, and the resulting message blames `env` for a flag
  `env` never saw — the same context burn that killed the first PR #1527 delegation.
- **Assessment:** strictly worse than Option A for the same benefit. Worth recording
  only so the shape is recognised when it appears in existing prompts.

### Option C — Move env-isolated verification out from under the guard

Either run it from a session that is not worktree-isolated, or place the command in a
script inside the worktree and invoke the script (the PR #1527 workaround).

- **Hazard reopened:** the guard cannot see inside a script file, so **all** of its
  checks are bypassed for that payload — including the git-redirect checks that are the
  guard's actual purpose. A script that runs `git -C <shared-checkout> …` executes
  unexamined. Running from a non-isolated session gives up worktree isolation entirely,
  which is the hazard [[project-primary-checkout-orphan-work-2026-08-14]] already cost
  us once.
- **Assessment:** acceptable only as a deliberate, per-incident escape hatch with the
  script's contents reviewed — never as the standing pattern for routine kanban
  verification.

### Option D — Report the argument-boundary misparse upstream

`env`'s operand boundary is determinable without executing anything: env's own options
end at the first token that is neither a dashed flag it recognises nor a `NAME=value`
assignment. The guard could stop its scan there instead of consuming `argv.length`,
which would accept `env -u A go test -run X ./...` while preserving every check.

- **Hazard reopened:** none — this is a request, not a change we control.
- **Cost:** unbounded latency, and it does nothing for the currently-installed version.
  Pair it with Option A rather than waiting on it.

---

*Investigation scope: read-only. No files outside this report were modified, and no
worktree, branch, or configuration state was changed.*
