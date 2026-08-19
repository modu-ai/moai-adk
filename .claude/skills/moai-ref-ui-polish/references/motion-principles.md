# Motion Principles (L3)

Depth layer behind the Motion table in `SKILL.md`. The Motion table holds the
implementation gotchas an agent gets wrong most often; this file holds the reasoning
that decides *what* the motion should be before those gotchas apply.

Load this when motion is the substance of the work — designing an animation system,
choosing a project-wide motion character, choreographing a multi-element scene — not
for a single hover state, which the Motion table already covers.

Principles are platform-neutral. Values are stated in milliseconds and easing families
so they carry to native mobile, desktop, and web alike.

---

## 1. Decide in Three Passes

Answer all three before naming a property or a duration. Skipping straight to the
property is the most common cause of motion that is technically smooth and emotionally
wrong.

| Pass | Question | Output |
|------|----------|--------|
| Emotional intent | What should the user feel? | An emotion, which fixes the duration band + easing family |
| Visual narrative | What is the setup, the action, the resolution? | A time budget split across the three |
| Motion craft | Do the curve, distance, and secondary motion serve the above? | Concrete values |

**Emotional intent → timing**

| Emotion | Character | Duration | Easing |
|---------|-----------|----------|--------|
| Trust | Smooth, predictable | 300-400ms | Gentle curves |
| Delight | Bouncy, surprising | 200-300ms | Overshoot |
| Urgency | Sharp, direct | 100-200ms | Snappy decelerate |
| Calm | Slow, flowing | 500-1000ms | Sine curves |
| Surprise | Sudden, explosive | 150-300ms | Exponential |
| Confidence | Direct, decisive | 200-400ms | Strong decelerate |

**Visual narrative → time budget.** Even a 200ms tooltip fade has an implicit
setup → action → resolution.

| Phase | Share | Purpose |
|-------|-------|---------|
| Setup | 20-30% | Establish context, prepare the viewer |
| Action | 30-40% | Primary motion, the moment being communicated |
| Resolution | 30-40% | Settle, breathe, confirm |

Leave 100-200ms of stillness after resolution before any new motion starts.

---

## 2. Three Motion Layers

Motion that animates only the thing being acted on reads flat. Richness comes from
layering, not from making the primary motion larger.

| Layer | Role | Amplitude | Timing |
|-------|------|-----------|--------|
| Primary | The action the viewer is following | 100% | Leads |
| Secondary | Supporting reaction (shadow, sibling, ripple) | 30-50% | 50-100ms after primary, *different* easing |
| Ambient | Background life | 10-20% | Continuous, slow, never demands attention |

A card that enters while its shadow grows behind it is one primary plus one secondary.
Adding a second simultaneous primary does not add polish; adding the secondary does.

---

## 3. The Two 1/3 Rules

**Distance.** No motion travels more than 1/3 of the viewport without an intermediate
keyframe. Break the span with a direction change, a speed shift, or an arc adjustment —
an unbroken long translate reads as a slide projector, not as movement.

**Simultaneity.** With 3+ animated elements, at most 1/3 may be in active motion at
once. Stagger so element 1 is settling as element 3 begins.

**Attention budget.** One hero motion per moment; 2-3 elements in active motion maximum.
Ambient motion does not count against the budget. When in doubt, stagger rather than
synchronize.

---

## 4. Stagger Budgets

Vary only the start time across a staggered set. Every element keeps the same easing
family — varying the curve as well reads as inconsistency, not as richness.

| Pattern | Per-element delay | Total budget | Use for |
|---------|-------------------|--------------|---------|
| Micro cascade | 20-40ms | <200ms | List items, grid cells |
| Standard | 50-100ms | <400ms | Cards, panels, nav |
| Wave | 30-60ms | <500ms | Data visualization bars |
| Dramatic | 100-200ms | <600ms | Hero sections only |

**Ceiling: total stagger stays under 500ms.** The Dramatic row is the source taxonomy's
own stated exception to that ceiling — treat any stagger past 500ms as needing
justification, and never apply one to a routine or high-frequency interaction.

Stagger ordering: sequential (reading order) for lists and grids; center-out for hero
content; reverse (bottom-to-top) for exits and backward navigation.

---

## 5. Motion Personality

Pick one archetype per project and hold it. Consistency across the interface is what
makes motion read as identity rather than as decoration.

| Archetype | Duration | Easing | Overshoot | Paths |
|-----------|----------|--------|-----------|-------|
| Playful | 150-300ms | back / bouncy spring | 10-20% | Arcs, never straight |
| Premium | 350-600ms | `cubic-bezier(0.4, 0, 0.2, 1)` | 0% | Smooth curves, subtle parallax |
| Corporate | 200-400ms | `cubic-bezier(0.2, 0, 0, 1)` | 0-3% | Mostly straight, small arcs for emphasis |
| Energetic | 100-250ms | expo / elastic | 15-30% | Dramatic arcs, diagonal, large displacement |

**Defaults when unspecified**: Corporate for UI surfaces, Playful for illustration.
Dashboards, admin, and data tools are Corporate — and the `cubic-bezier(0.2, 0, 0, 1)`
curve pinned in `SKILL.md` is the Corporate signature, which is why it is the house
default.

**Mixing**: hold ~90% to the primary archetype. A single moment may borrow another
(a Corporate dashboard borrowing Playful for a success state); ease into the shift
rather than snapping between characters.

---

## 6. Disney Principles, UI-Adapted

The twelve principles reduced to the ranges that apply to interface motion.

| Principle | UI application |
|-----------|----------------|
| Squash and stretch | Scale ~[1.2, 0.8] on impact (30-65ms), recover over 65-130ms; preserve volume. Skip entirely for Premium |
| Anticipation | Small counter-motion before the action: 100-200ms at 10-20% of the main magnitude. Skip below 150ms total |
| Staging | Dim non-hero elements to 40-60% opacity; hero enters 100-200ms after its supporting cast |
| Pose to pose | Plan keyframes for UI state changes; straight-ahead only for particles and generative motion |
| Follow through | Child elements trail the parent by 50-150ms; trailing parts get lower spring stiffness |
| Slow in / slow out | Entrance decelerates, exit accelerates, on-screen movement eases in and out, ambient loops on a sine curve. **Never linear for spatial movement** — linear is for rotation, progress bars, and timers only |
| Arcs | 10-20px perpendicular offset at the path midpoint; 5px for Corporate, 20px+ for Playful |
| Secondary action | 30-50% of primary amplitude, 50-100ms later, different easing |
| Timing | Heavy surfaces (modals, pages) 400-800ms; light ones (tooltips, toggles) 100-250ms. **Entrances run 30-50% longer than exits** — users care more about what arrives |
| Exaggeration | Overshoot 10-30% past target, by archetype: Playful 15-25%, Energetic 20-30%, Corporate 0-5%, Premium 0% |
| Solid drawing | Keep proportions consistent across keyframes; shadow behavior follows the implied light source |
| Appeal | Smooth curves over sharp angles. Appeal killers: jerky motion, inconsistent timing, abrupt stops, everything animating identically |

The slow-in/slow-out row is the principle behind the enter and exit rows of the Motion
table in `SKILL.md`: an entrance decelerates because the element is arriving and should
settle; an exit accelerates because it is leaving and should get out of the way.

---

## 7. Emotion to Motion

| Emotion | Path | Easing | Duration |
|---------|------|--------|----------|
| Joy / delight | Curved, upward | back (overshoot) | 200-400ms |
| Calm | Gentle curves | sine ease-in-out | 500-1000ms |
| Urgency | Straight lines | decelerate | 100-200ms |
| Sadness / weight | Drooping curves | cubic ease-in-out | 600-1200ms |
| Surprise / impact | Radial outward | expo decelerate | 150-300ms |
| Elegance | Long smooth arcs | `cubic-bezier(0.4, 0, 0.2, 1)` | 400-700ms |
| Confidence | Straight, horizontal | decelerate | 200-400ms |
| Tenderness | Very subtle curves | soft ease-in-out | 600-1000ms |

**Path shape carries meaning independently of timing.**

| Path | Reads as |
|------|----------|
| Angular / sharp | Tense, urgent, mechanical |
| Curved / smooth | Relaxed, friendly, organic |
| Straight diagonal | Dynamic, purposeful |
| Vertical up / down | Growth / settling |
| Radial outward / inward | Release / focus |

**Context defaults**: form success → joy + confidence; validation error → mild urgency;
page load and dashboards → calm + confidence; navigation → confidence; notification →
mild surprise; onboarding → curiosity + delight; delete → calm (a respectful departure,
never a celebratory one).

**Color pairing**: transition *to* success green rather than starting there; flash error
red and settle rather than sustaining it; pulse amber for warnings; prefer opacity over
a color change for neutral state shifts.

---

## Cross-references

- `SKILL.md` § Motion — the implementation rules and anti-patterns these principles sit above
- `SKILL.md` § Motion Easing Values — the concrete curve constants for web targets
- `SKILL.md` § Motion Accessibility and Cost — the reduced-motion escape hatch and the compositor-cost rule, both of which bind every value on this page
