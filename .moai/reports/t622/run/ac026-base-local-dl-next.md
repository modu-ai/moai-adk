#### Context-Aware Next Steps

Full-pipeline completion close: when this sync phase completed as the tail of a `full-pipeline` contract AND no genuine pending decision remains, close with the Completion Report as a clean completion statement — NO manufactured next-step question is emitted (the askuser-protocol § Completion-Report Next-Step Discipline "close with NO question" clause is the full-pipeline default). A genuine next-step decision, when one actually exists, still rides AskUserQuestion. The option lists below apply to `single-phase` contract completions (explicit `/moai sync` invocation), whose "(Recommended)" next-step chain question is unchanged.

Tool: AskUserQuestion with options tailored to delivery result (single-phase contract):

**If PR was created (github-flow feature branch, or a git-flow PR route):**
- Review PR on GitHub (Recommended)
- Auto-Merge PR (/moai sync --merge)
- Create Next SPEC (/moai plan)
- Start New Session (/clear)

**If direct push (github-flow main branch, git-flow develop, or a git-flow `WT-*` integration merge):**
- Create Next SPEC (/moai plan) (Recommended)
- Continue Development
- Start New Session (/clear)

**If worktree context:**
- Review PR in Browser (Recommended)
- Return to Main Directory
- Remove This Worktree

---

