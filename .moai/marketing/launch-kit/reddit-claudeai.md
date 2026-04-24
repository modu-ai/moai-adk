# r/ClaudeAI Submission Kit

## Key Messaging Decisions

1. **Tone**: Builder's voice, personal experience, no hype. "I built this because X frustrated me."
2. **Angle**: Claude Code power users who want *methodology* layered on top of agents, not just chat.
3. **Proof Point**: Real workflow example (code block showing `/moai plan` + `/moai run`), not abstract claims.
4. **Community Signaling**: Explicitly invite feedback, acknowledge limitations, show vulnerability (early-stage, learning).
5. **Flair**: Choose "Discussion" or "Open Source" flair (not "News" or "Shilling").

---

## Post Title

**"I built MoAI-ADK: SPEC-first development framework for Claude Code. Tired of chat-based coding—wanted methodology instead."**

- Length: 122 characters (well under Reddit 300-char limit)
- Why it works: Personal ("I built", "I was tired") + concrete problem statement + tool intro.
- Flair: **Discussion** or **Open Source** (avoid "Announcement")

---

## Post Body

```
Over the last 7 months, I've been working on MoAI-ADK—a development framework for 
Claude Code that enforces a SPEC-first, quality-driven workflow. It's now at v2.12.0, 
and I wanted to share it with the community here and get your feedback.

## What This Is

Not just another "AI coding tool." It's a methodology + harness. You write a formal 
specification, 24 specialized AI agents implement with TDD or DDD (auto-selected per 
project), then docs and quality gates run automatically. Single Go binary, zero 
dependencies, works on any platform.

The central idea: **The engineer designs the environment. The AI executes.**

## Why I Built This

About a year ago, I was using Cursor heavily. Amazing IDE. But I kept hitting the same 
wall: without clear specs, agents would implement features that sort-of-worked but 
didn't fit the architecture. Code reviews became messy. Tests were often incomplete. 
Documentation lagged behind.

I started writing detailed specs before each feature. Suddenly, agents made better 
decisions. Code quality went up. Fewer surprises in review.

So I built a framework around that discipline.

## Real Example

Here's what using MoAI looks like in practice:

```bash
# Step 1: Write a specification
moai plan "Add JWT authentication with refresh token rotation"

# That generates a SPEC document in EARS format
# (Event, Always, State, Requirements). You fill in details:
# - When user logs in with email+password...
# - Always validate email format
# - State machine: unauthenticated → authenticated → token-expired
# - Requirement: 85%+ test coverage

# Step 2: Run the implementation
moai run SPEC-001

# Agents read the spec, choose TDD (new feature), and autonomously:
# - Create empty stubs
# - Write tests first (RED phase)
# - Implement to pass (GREEN phase)
# - Refactor for clarity (BLUE phase)
# - Generate documentation

# Step 3: Verify and sync
moai sync SPEC-001

# Validates LSP, linting, coverage. Generates PR.
```

No ping-pong chat. No manual code review loops. Just spec → agents → done.

## What Makes It Different

1. **Methodology baked in**: TDD for greenfield, DDD for legacy. Not a religious choice—
   determined by project type. Enforced, not optional.

2. **16 languages supported**: Auto-detects Go, Python, TypeScript, Rust, Java, Kotlin, 
   C#, Ruby, PHP, Elixir, C++, Scala, R, Flutter, Swift, and more. Picks the right 
   linter, test runner, coverage tool. Works the same way regardless of language.

3. **Multi-language docs**: English, 한국어, 日本語, 中文. Not common for early-stage tools.

4. **Built on Go**: Single binary, ~5ms startup, zero runtime dependencies. No venv, 
   no Docker, no npm install nightmare. Works on macOS, Linux, Windows (WSL).

5. **Opus 4.7 native**: Uses Adaptive Thinking with dynamic token allocation. No fixed 
   budgets. Cleaner prompts. Better cost efficiency.

## Numbers

- **937 GitHub stars** (steady growth over 7 months)
- **38K+ lines of Go code**, 85-100% test coverage
- **24 specialized agents**, 52 skills
- **TRUST 5 framework**: Tested (85%+ coverage), Readable, Unified, Secured (OWASP), 
  Trackable (conventional commits)
- **v2.12.0 just shipped**: Design system (Claude Design handoff parser) + DB schema 
  sync + Opus 4.7 full support

## Honest Limitations

- **Early-stage**: Not battle-tested on 10K+ LOC projects yet. I use it daily on 
  real work, but YMMV.
- **Opinionated**: Methodology is enforced, not optional. If you prefer "write code 
  first, docs later," this will feel prescriptive.
- **Learning curve**: Takes ~2h to learn the Plan/Run/Sync workflow. But once it 
  clicks, it becomes natural.
- **Claude Code exclusive**: Doesn't work with other editors (yet). Only Claude Code.

## What I Want Feedback On

1. **Is SPEC-first methodology valuable?** Or does it feel like overkill for smaller 
   projects?
2. **TDD vs DDD auto-selection**: I auto-select based on existing coverage. Does that 
   logic make sense to you?
3. **Pain points with Claude Code agents**: What's YOUR biggest friction point with 
   agent-assisted coding right now? I'm iterating based on feedback.
4. **Language support**: Any languages you'd want but don't see in the list?

## Get Started

- **GitHub**: https://github.com/modu-ai/moai-adk
- **Docs**: https://adk.mo.ai.kr
- **Discord**: https://discord.gg/moai-adk (small, friendly community—drop by and say hi)
- **License**: Apache 2.0 (open-source, commercial-friendly)

Happy to answer any questions here or in Discord. Really curious what the r/ClaudeAI 
community thinks about this.

Thanks for reading!
```

---

## Word Count
~850 words (ideal for Reddit discussion, not TL;DR territory but still readable).

---

## Engagement Strategy

### Respond to Top-Level Comments (Priority Order)

1. **Any technical questions** (1–5 min) → prioritize accuracy over speed
2. **Comparisons to other tools** (Cursor, Aider, etc.) → acknowledge strengths, explain differentiation
3. **Criticism of methodology** ("too opinionated", "overkill") → validate concern, explain target user
4. **Requests for features** → add to backlog, thank for idea
5. **Spam / low-effort trolling** → downvote, don't engage

### Pre-emptive Comment Seeding (Optional)

If you want to seed early engagement, post 2–3 comments 10–30 min after posting:
- "Would love to know if anyone here is already using Claude Code agents. What's your workflow?"
- "The GitHub link has examples if you want to try it. Happy to troubleshoot if something breaks."

But **do NOT** seed more than 2 comments—looks like artificial boosting and kills credibility on Reddit.

---

## Best Time to Post

**Tuesday–Thursday, 10 AM–2 PM EST** (peak engagement for r/ClaudeAI, when US-based developers are at desks).

Avoid Friday evenings and weekends (low traffic). Avoid Monday mornings (inbox overload from weekend).

---

## Flair Selection

- **Recommended**: "Discussion" or "Open Source"
- **Avoid**: "News" (sounds like corporate announcement), "Shilling" (self-explanatory)
- **Why**: "Discussion" signals you want feedback, not just visibility

---

## Expected Engagement

| Metric | Realistic Target | Notes |
|--------|------------------|-------|
| Upvotes | 300–800 | Niche but technical; good engagement if 500+ |
| Comments | 40–80 | Shows genuine interest, not dismissal |
| Reddit silver/gold | 0–2 | Measure of quality, not critical |
| Click-through to GitHub | ~5% of viewers | 15–40 new stars over 48h |
| Discord joins from Reddit | 5–10 | Smaller subreddit, less referral |

---

## Red Flags to Avoid

1. **Don't compare directly to Cursor/Windsurf** ("MoAI is better because..."). Instead: "Different philosophy. Cursor is IDE-first; MoAI is methodology-first."

2. **Don't oversell maturity**. "v2.12.0 just shipped" is OK. "Production-ready for all use cases" is dishonest.

3. **Don't post on multiple subreddits on the same day** (Reddit's spam filter flags simultaneous cross-posts).

4. **Don't ask for upvotes or comment** ("Please upvote if interested"). Violates Reddit rules, gets downvoted.

5. **Don't engage in heated debates** about programming methodologies. You'll lose. Just acknowledge and move on.

---

## If the Post Gains Traction

If you hit 500+ upvotes and 60+ comments within 24h:

1. **Post a follow-up comment** summarizing key feedback and what you're planning next.
2. **Monitor Discord** for new joiners; be extra responsive.
3. **Consider a AMA (Ask Me Anything)** in a follow-up thread if community interest is strong.
4. **Update your GitHub README** with Reddit feedback highlights (e.g., "Rust users asked for better support; added to v2.13 roadmap").

---

## If the Post Underperforms

If you don't hit 200 upvotes after 24h:

1. **Don't delete and repost** (spam flagged immediately).
2. **Wait 4–6 weeks**, prepare a follow-up milestone post ("v2.13 ships with X"), and resubmit.
3. **Leverage Discord/GitHub discussions** for community feedback instead of chasing Reddit momentum.
4. **Consider r/golang** or r/programming** as alternative forums (different audience, different reception).

---

## Sample Follow-Up Comments (Use if Relevant)

### If someone asks, "How is this different from Cursor + prompting?"

```
Great question. Cursor is a fantastic IDE—best editing experience on the market. 
MoAI-ADK is a methodology + agent orchestration layer that lives on top of 
Claude Code (Cursor isn't a Claude Code environment).

Main difference:
- Cursor: You write prompts, agent writes code, you review manually.
- MoAI: You write specs in EARS format, agents implement autonomously following 
  the spec, verification is automatic.

It's the difference between "describe what you want" (interactive) vs "design the 
system, delegate execution" (autonomous). Different philosophies for different 
contexts.

Use both? Sure. Use one or the other? That's fine too.
```

### If someone says, "Isn't this overkill for small projects?"

```
100% fair. If you're a solo dev hacking on side projects, MoAI is probably 
overkill. The target is:

- Teams > 3 engineers (need consistency)
- Projects with existing code quality standards
- Situations where deviation from spec is a liability (financial services, 
  healthcare, etc.)

For rapid prototyping or solo explorations, lighter tools (Aider, Cursor) are 
better.

That said, I'm always interested in lowering the onboarding barrier. If you try 
it and find it too heavy, tell me what felt burdensome—might inspire a 
lighter-weight workflow path.
```

### If someone asks, "Why not just use ChatGPT / Claude directly?"

```
Fair question! You absolutely can use Claude in a chat interface for coding tasks. 
The limitation I hit:

- Chat is great for brainstorming and one-off questions.
- But for complete features with architectural constraints, quality gates, and 
  documentation? Chat starts to feel fragile. Scope creep, missed requirements, 
  no audit trail.

MoAI adds structure: specs, methodology, automated verification. Think of it like 
the difference between "ask a colleague for advice" (chat) vs "write requirements, 
formal review process" (SPEC-driven).

Both have value. Different problems they solve.
```
