# Implementation Plan — SPEC-OWASP-CHECKLIST-GAP-001

## §A. Context

- **Working directory**: `/Users/goos/MoAI/moai-adk-go`
- **Target files** (exactly 2 content files + 1 metadata file, both content files currently byte-identical):
  - `internal/template/templates/.claude/skills/moai-ref-owasp-checklist/SKILL.md` — **template source, single source of truth** (edit FIRST)
  - `.claude/skills/moai-ref-owasp-checklist/SKILL.md` — local deployed copy (edit SECOND, to match template source byte-for-byte)
  - `internal/template/catalog.yaml` — the `moai-ref-owasp-checklist` entry's `hash:` field (lines ~205-208 at plan-time; regenerate via tooling, do not hand-edit the hash value)
- **Tier**: S (single skill file, < 300 LOC changed, no code, no test suite) — this plan uses the Tier S minimal delegation form when a future run-phase delegates to `manager-develop`; the full Section A-E `manager-develop-prompt-template.md` structure is optional. Per `.claude/rules/moai/workflow/spec-workflow.md` § SPEC Complexity Tier, this Tier S SPEC uses the canonical 2-file artifact set (spec.md + plan.md, AC inline in spec.md §3) — `acceptance.md` was removed in this revision (plan-auditor D7 remediation; missing `acceptance.md` is explicitly acceptable for Tier S per `internal/spec/closer.go`'s LEAN-workflow comment).
- **Plan-phase baseline commit**: `366e701af60bb789714efbe6068cac59788fb6bb` — recorded at plan-phase authoring time, before any M1-M5 implementation edit begins. Referenced by spec.md §3 AC-OCG-006 (no-deletion-outside-scope check) and AC-OCG-008 (post-commit no-new-skill check).
- **Sync mechanism verified this session**: `make build` (`Makefile` target `build`) runs `go run ./internal/template/scripts/gen-catalog-hashes.go --all` followed by `go build`. It regenerates `catalog.yaml` hash fields and recompiles the binary (which re-embeds template content via `go:embed` at compile time). **`make build` does NOT copy files from `internal/template/templates/` into the local `.claude/` directory** — that copy only happens via `moai update` (full sync, may touch unrelated files) or a manual mirrored edit. Given this SPEC touches exactly one file, a manual mirrored edit (apply the identical diff to both files) is lower-risk than invoking `moai update`.

## §B. Known Issues / Risks (Tier S filtered subset)

- **Vendor-name leakage risk**: the 5 new principles are *generalized from* vendor-specific naming (Supabase `getSession()`/`getUser()`, Vercel `CRON_SECRET`, Next.js `productionBrowserSourceMaps`, Supabase `x-supabase-signature`). The implementer MUST write the abstracted principle only — no vendor/framework/host name may appear in the committed text. Verify with a case-insensitive grep sweep (§D Constraints) before considering the milestone done.
- **Frontmatter description/when_to_use overlap**: `expert-security` and `expert-backend` appear in BOTH the frontmatter block (lines 6, 12) and the body (`## Target Agents`, lines 34-35) of the *current* file. All 4 occurrences must be corrected, not just the body section.
- **catalog.yaml hash drift**: if the hash field is left stale after editing `SKILL.md`, a future `moai spec audit` or catalog-integrity check will flag a mismatch. Regenerate immediately after content edits (M4 below).
- **Scope creep risk**: `expert-security`/`expert-backend` also appear in `moai-ref-api-patterns/SKILL.md`, `moai-foundation-cc/SKILL.md`, `moai-workflow-spec/references/reference.md`, and `moai-meta-harness/references/*.md`. These are explicitly OUT OF SCOPE (spec.md §D) — do not touch them in this SPEC's commits.

## §C. Pre-flight Checklist (for the future run-phase delegate)

```bash
# 1. Confirm current byte-identity baseline (should report no differences before edits begin)
diff .claude/skills/moai-ref-owasp-checklist/SKILL.md \
     internal/template/templates/.claude/skills/moai-ref-owasp-checklist/SKILL.md

# 2. Confirm no other skill is touched (scope containment baseline)
git status --porcelain -- '.claude/skills/' 'internal/template/templates/.claude/skills/'

# 3. Confirm the catalog.yaml entry exists and note current hash
grep -n -A5 "name: moai-ref-owasp-checklist" internal/template/catalog.yaml
```

## §D. Technical Approach

### D.1 Milestones (priority-ordered, no time estimates)

- **M1 (Priority High) — Add the 5 generalized principles to the template source.**
  Placement: a new `## Trust Boundary Verification Principles` H2 section, inserted after the existing `## Security Review Severity Levels` table and before the `<!-- moai:evolvable-start id="rationalizations" -->` block (i.e., appended to the end of the static reference content, ahead of the evolvable blocks). Recommended table form (5 rows, `Principle | Applies To | Defense`):

  | Principle | Applies To | Defense |
  |-----------|------------|---------|
  | Cached/client-supplied session state is not proof of current identity | Any framework caching or locally decoding a session/JWT value | Re-verify identity against the server-side source of truth (session store, token introspection, identity provider) before every authorization decision |
  | Edge/gateway/middleware auth checks are a UX convenience, not a security boundary | Reverse proxies, framework middleware, API gateways, serverless edge functions | Every mutation-handling endpoint independently re-checks authentication AND resource-ownership authorization |
  | Scheduled/cron-triggered HTTP endpoints are still public URLs | Any scheduler that invokes an HTTP endpoint (cron, Lambda scheduled events, Cloud Scheduler, k8s CronJob) | Require a shared-secret bearer check (constant-time compare) on every scheduled-endpoint invocation |
  | Production builds must not expose source maps or equivalent debug artifacts | Any bundler/build tool | Disable production source maps / verbose stack traces / build manifests in production configuration |
  | Webhook receivers must verify a signature/HMAC header | Any webhook provider (Stripe, GitHub, Slack, or a custom sender) | Verify signature/HMAC against a shared secret before trusting the payload as legitimate business data |

  The exact wording above is a starting draft, not a frozen requirement — the implementer may adjust phrasing, but MUST preserve vendor-neutrality (REQ-OCG-002) and MUST cover all 5 concepts (REQ-OCG-001).

- **M2 (Priority High) — Fix the stale agent references in the template source.**
  - Frontmatter `description` (line 6) and `when_to_use` (line 12): replace "Agent-extending skill that amplifies expert-security and expert-backend expertise..." / "Amplifies expert-security and expert-backend expertise..." with vendor/agent-neutral phrasing, e.g. "Agent-extending skill that amplifies backend-implementation and security-audit workflows with production-grade security patterns."
  - Body `## Target Agents` section (lines 32-35): replace with:
    ```markdown
    ## Target Agents

    - `manager-develop` - Applies checklist during backend API implementation (`cycle_type=tdd` or `cycle_type=ddd` context)
    - `/moai security` workflow - Primary security-audit invocation surface; equivalently available as a per-spawn `Agent(general-purpose)` security specialist per `archived-agent-rejection.md` §C
    ```

- **M3 (Priority High) — Propagate to the local deployed copy.**
  Apply the identical M1 + M2 edits to `.claude/skills/moai-ref-owasp-checklist/SKILL.md` so the two files remain byte-identical. Verify with `diff` (must report no output).

- **M4 (Priority Medium) — Regenerate the catalog.yaml hash.**
  ```bash
  go run internal/template/scripts/gen-catalog-hashes.go --entry moai-ref-owasp-checklist
  # or, to be safe against unrelated stale hashes:
  go run internal/template/scripts/gen-catalog-hashes.go --all
  ```
  Verify the `hash:` field for the `moai-ref-owasp-checklist` entry in `internal/template/catalog.yaml` changed (it MUST, since file content changed) and that a `--dry-run` re-run reports the same hash (idempotent / stable).

- **M5 (Priority Medium) — Verification sweep.**
  Run the grep/diff battery in spec.md §3.1 Quality Gate Criteria (vendor-name absence, archived-agent-name AND-semantics check, byte-identity, no-deletion-outside-scope, baseline-anchored scope containment via `git diff --name-only <baseline-SHA>`, pre- AND post-commit no-new-skill check) and confirm all pass before considering the SPEC implementation-ready.

### D.2 Explicitly not planned (per spec.md §D Out of Scope)

- No new skill file/directory.
- No Next.js/Vercel/Supabase-specific content anywhere in the committed diff.
- No edits to `moai-ref-api-patterns`, `moai-foundation-cc`, `moai-workflow-spec/references/reference.md`, `moai-meta-harness`, or any agent/workflow/rule file.

## §E. Self-Verification Deliverables (for the future run-phase delegate)

When a future `manager-develop` (or orchestrator-direct, given Tier S) delegation reports completion, it MUST include:

1. **Byte-identity proof**: `diff` output between the template source and the local deployed copy (expect empty).
2. **Vendor-neutrality grep**: case-insensitive grep for `supabase|next\.?js|vercel` across the diff (expect zero matches).
3. **Archived-agent-name AND-semantics proof**: grep for `expert-security|expert-backend` (expect zero matches) AND grep for `manager-develop` (expect >=1) AND grep for `/moai security|Agent\(general-purpose` (expect >=1) — all three sub-checks must hold simultaneously (spec.md §3 AC-OCG-002).
4. **Scope-containment proof**: `git diff --name-only <plan-phase-baseline-SHA> -- . ':!.moai/specs/SPEC-OWASP-CHECKLIST-GAP-001/'` filtered against the 3-path allowlist (expect no output) — use `--name-only`, NOT `--stat` (the `--stat` trailing summary line is not path-shaped and always leaks past a path-exclusion grep — D2). The command MUST be anchored to the plan-phase baseline SHA recorded in §A, NOT a bare unanchored `git diff` — a bare `git diff --name-only` compares only worktree-vs-index and vacuously passes once a violation is staged or committed (D9; empirically reproduced: bare form loses the violation after `git add`/`git commit`, the baseline-anchored form retains it). See spec.md §3 AC-OCG-007.
5. **Catalog hash proof**: `gen-catalog-hashes.go --dry-run --entry moai-ref-owasp-checklist` output matching the committed `catalog.yaml` hash.
6. **No-deletion-outside-scope proof**: `git diff <plan-phase-baseline-SHA> -- <file>` filtered to deleted lines (`^-[^-]`) minus stale-reference matches (expect no output) — run BEFORE the implementation commit, using the plan-phase baseline SHA recorded in §A (spec.md §3 AC-OCG-006).
7. **No-new-skill proof (pre- AND post-commit)**: `git status --porcelain` (pre-commit) AND `git diff --name-only --diff-filter=A <plan-phase-baseline-SHA>` (post-commit) both scoped to the skills directories (expect no output from either) — running only the pre-commit check leaves a post-commit blind spot (spec.md §3 AC-OCG-008).

## §F. Milestone Summary Table

| Milestone | Priority | Deliverable |
|-----------|----------|-------------|
| M1 | High | 5 generalized principles added to template source |
| M2 | High | Stale `expert-security`/`expert-backend` references corrected (frontmatter + body) in template source |
| M3 | High | Local deployed copy mirrored to byte-identity |
| M4 | Medium | `catalog.yaml` hash regenerated for the entry |
| M5 | Medium | Full verification sweep (vendor-neutrality, archived-agent-absence, byte-identity, scope containment) |

## §G. Anti-Patterns (to avoid during implementation)

- Editing only the local deployed copy and forgetting the template source (violates the Template-First Rule, `CLAUDE.local.md` §2).
- Editing only the template source and forgetting to mirror the local copy (leaves the two files diverged, contrary to their current byte-identical baseline).
- Running `moai update` as the sync mechanism for this single-file change (broader blast radius than necessary for a Tier S change; risks touching unrelated local customizations).
- Copying vendor-specific example code or prose verbatim from the external repository (violates REQ-OCG-002 and the 16-language-neutrality doctrine).
- "While I'm in this file" drive-by edits to unrelated sections (`OWASP API Security Top 10` table, `Authentication Checklist`, etc.) — scope discipline applies (REQ-OCG-008).

## §H. Cross-References

- `.claude/rules/moai/development/manager-develop-prompt-template.md` § Applicability (Tier S minimal form)
- `CLAUDE.local.md` §2 (Template-First Rule, File Synchronization)
- `internal/template/scripts/gen-catalog-hashes.go` (hash regeneration tooling, usage documented in its header comment)
