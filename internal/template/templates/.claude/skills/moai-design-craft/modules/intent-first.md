# Intent-First Design Process

The Intent-First process ensures every design decision traces back to a clear statement of purpose. It prevents premature visual decisions by anchoring work in domain understanding.

## Principle: Why Before What

Design problems have two layers:
1. **Surface layer**: What does it look like? What components are used?
2. **Intent layer**: Why does this exist? What does the user need to accomplish? What would success feel like?

Intent-First starts at layer 2 and works up to layer 1. Skipping layer 2 produces technically correct UI that feels wrong.

## Domain Exploration

Before any component or layout decision, explore the domain:

### Step 1: User Context
- Who is the user in this specific moment? (Not a persona — the actual task they are doing)
- What do they already know? What are they uncertain about?
- What is the cost of a mistake for them?

### Step 2: Mental Model Audit
- What real-world model does the user bring to this interaction?
- Does the interface reinforce or contradict that model?
- Where is the vocabulary mismatch between domain and UI?

### Step 3: Interaction Contract
- What is the user committing to when they interact with this element?
- What feedback do they need to feel confident?
- What does an empty state, error state, and success state mean in this domain?

## Design Direction Statement

After domain exploration, write a Design Direction — 1 to 3 sentences that capture:

1. **User intent**: What the user is trying to accomplish
2. **Design intent**: How the interface supports that goal
3. **Craft principle**: The single most important quality constraint (e.g., "must feel instant", "must signal permanence", "must not overwhelm")

Example:
> The user is reconciling a financial transaction they did not initiate. The interface must surface all relevant context without requiring navigation. Trust is the craft principle — every element must communicate reversibility or confirmation before action.

Write this to `.moai/design/system.md` under the relevant section.

## Domain Vocabulary

Collect 3–5 terms that belong to the domain (not the UI layer). These become the shared vocabulary for components, copy, and code comments.

Example for a payments domain:
- "Reconciliation" (not "match" or "sync")
- "Initiating party" (not "sender" or "payer")
- "Settlement window" (not "processing time")
- "Dispute" (not "issue" or "problem")

Bad vocabulary choices create misaligned UI. When developers name things after UI patterns instead of domain concepts, the implementation drifts from intent.

## When to Trigger Design Direction

Trigger this process when a SPEC contains any of:
- New user-facing flows (not just modifying existing components)
- Features that require the user to make a consequential decision
- Features in a domain the codebase has not handled before
- Any feature where the copy or labels are part of the design (not just styling)

## Output

Write findings to `.moai/specs/SPEC-{ID}/design-direction.md` (when inside a SPEC workflow) containing:
- Intent statement (who, what, feel)
- Domain concepts and vocabulary (5+ entries)
- Color world exploration (5+ entries)
- Signature element definition
- Defaults to avoid (3+ generic patterns to reject)

Additionally, offer to persist project-level decisions to `.moai/design/system.md` via the write protocol in `modules/design-memory.md`.
