# Session Handoff — Appendix (examples and the goal-first variant)

> Reference appendix relocated from `.claude/rules/moai/workflow/session-handoff-examples.md` (instruction-budget reduction): the two illustrative examples and the documented goal-first bootstrap variant. This location sits outside the loaded-instruction budget, so it costs no trigger budget; the rules file keeps a pointer.

## Example (Illustrative; substitute project-specific values when adapting)

```
✂──── 여기부터 복사 ────✂

ultrathink. SPEC-MYPROJ-001 implementation 진입.
applied lessons: <lesson-id-1>, <lesson-id-2>.
source_session_id: <not-available — environment-fallback, next session will backfill via /moai session register on activation>

전제 검증:
1) git log --oneline -1 → <commit-sha> 확인
2) ls .moai/specs/SPEC-MYPROJ-001/ → N files

실행: /moai run SPEC-MYPROJ-001

머지 후: SPEC-MYPROJ-002 → SPEC-MYPROJ-003

✂──── 여기까지 복사 ────✂
```

> Block 5 carries the work-starting action. Where the next SPEC declares a machine-verifiable end-state, the orchestrator arms `/moai goal "<condition>"` alongside it after Implementation Kickoff Approval — arm-only, so it never replaces the `실행:` action (§ Canonical Format, Field-by-Field Block 5).


## Example with Block 0 (Illustrative)

```
✂──── 여기부터 복사 ────✂

[New Terminal — START IN WORKTREE]
$ moai cc -w ~/.moai/worktrees/<project>/SPEC-MYPROJ-001
   # (launcher -w accepts L2 absolute paths; or moai glm -w ...)

ultrathink. SPEC-MYPROJ-001 Epic N 진입.
applied lessons: <lesson-id-1>, <lesson-id-2>.

전제 검증:
0) git rev-parse --show-toplevel → ~/.moai/worktrees/<project>/SPEC-MYPROJ-001 (★ critical)
1) gh pr view <PR-number> → MERGED

실행: /moai run SPEC-MYPROJ-001 --team

후속: Milestone M<N+1> (single-SPEC next step) 또는 Epic N+1 (multi-SPEC next grouping)

✂──── 여기까지 복사 ────✂
```

---


## Goal-first bootstrap variant (documented alternative — NOT the default)

[ZONE:Evolvable] An explicit alternative single-paste form exists: the **goal-first bootstrap** — a one-line `/moai goal` message whose condition text carries both a resume pointer and the compact completion condition. Illustrative:

```text
/moai goal "resume SPEC-X run: read <handoff-file> from memory and progress.md, then continue. Completion: <machine-verifiable end-state>, or stop after N turns."
```

(The condition text follows the user's `conversation_language`; shown above in English-canonical form. The `/moai goal` token itself is locale-verbatim.)

Normative content:

- **(a) Selection criterion**: choose goal-first bootstrap when the user wants one-paste + autonomous continuation; the standard 6-block paste (§ Canonical Format) remains the DEFAULT.
- **(b) Caveats**: effort keywords (`ultrathink` / `ultracode`) placed inside a command argument are NOT documented to fire — the session may run at default effort; and precondition verification shifts from paste-time structure (the Block 4 verifiable commands) to **model discretion** via the directive text.
- **(c) Invariants preserved**: the condition must stay compact (one measurable end state); the Implementation Kickoff Approval gate is unaffected — arming never authorizes autonomous run-phase entry; the `/moai goal` token stays locale-verbatim (never translated).

