# Design Audit Suite (L3)

Mechanical detection for the rules `SKILL.md` states in prose. The checklists in
`SKILL.md` § Verification and § Red Flags say what good looks like; this file says how
to find the places where a codebase is not that.

Load this when auditing an existing codebase — a `/moai review` UI pass, a pre-ship
gate, or a "is our motion consistent?" question. Do not load it to write a single
component; the rules themselves live in `SKILL.md`.

## How to run these

Every pattern below is a regex plus a file scope. Two ways to run them:

- **Preferred inside a MoAI agent**: the `Grep` tool, per `agent-common-protocol.md`
  § Tool Selection by Task. Pass the pattern, set `glob` to the file scope, use
  `output_mode: "files_with_matches"` first and only then re-run for content.
- **Bash form** (shown here for exactness): the commands below are literal. Replace
  `<scope>` with the project's source root — `src/`, `app/`, `lib/`, `internal/`, or
  `.` — determined by reading the project layout first. There is no universal `src/`.

**Bound the output.** Several patterns are deliberately broad and can return hundreds
of lines on a large codebase. Pipe through `| wc -l` first; only read the matches when
the count is small enough to be worth reading. Per `cache-aware-execution.md` § 7, a
grep that dumps thousands of lines into context has cost more than it found.

**A match is a candidate, not a defect.** Every pattern here over-matches by
construction — see the Signal column. Confirm the finding by reading the site before
reporting it. A grep hit reported as a defect without that read is an unobserved
defect claim (`verification-claim-integrity.md` §1.1 surface 3).

---

## 1. Motion Gaps

| Check | Signal | What a real hit means |
|-------|--------|-----------------------|
| Conditional render without exit animation | Low — over-matches badly | A mount/unmount with no exit path; the element vanishes instantly |
| Hover rule without transition | High | Instant state flip; reads as broken rather than as feedback |
| Enter declared without exit | High | Asymmetric by omission, not by intent |
| List render without stagger | Low — most `.map()` calls are not animated | Simultaneous pop-in on a staged entrance |
| Inline style change without transition | Low | Dynamic colour/position jumps with no tween |

```bash
# Conditional renders lacking an exit-animation wrapper (React/JSX projects)
grep -rn '{.*&&\s*<' --include='*.tsx' --include='*.jsx' <scope> | grep -v 'AnimatePresence'

# :hover rules with no transition on the same file
grep -rn ':hover' --include='*.css' --include='*.scss' --include='*.module.css' <scope> \
  | grep -vE 'transition|animation'

# Enter declared without a matching exit
grep -rn 'initial=' --include='*.tsx' --include='*.jsx' <scope> | grep -v 'exit='

# Lists rendered without any stagger mechanism
grep -rn '\.map(' --include='*.tsx' --include='*.jsx' <scope> \
  | grep -vE 'stagger|delay.*index|variants|transition.*delay'

# Inline style mutation with no transition
grep -rn 'style={{' --include='*.tsx' --include='*.jsx' <scope> \
  | grep -vE 'transition|transform|opacity'
```

The first, fourth, and fifth patterns are JSX-shaped and only meaningful in a
React-family project. Skip them entirely elsewhere rather than reporting zero matches
as a pass — a check that could not apply is a gap, not a clean result.

---

## 2. Accessibility

### Reduced motion — the one non-negotiable check

`SKILL.md` § Motion Accessibility and Cost requires a `prefers-reduced-motion` branch on
every non-decorative animation. This detects whether the project has *any* handler at
all. Run all three; the project only needs the one matching its stack.

```bash
# Web
grep -rn 'prefers-reduced-motion' --include='*.css' --include='*.scss' --include='*.ts' \
  --include='*.tsx' --include='*.js' --include='*.jsx' <scope>

# Apple
grep -rn 'accessibilityReduceMotion\|isReduceMotionEnabled\|reduceMotionStatusDidChangeNotification' \
  --include='*.swift' <scope>

# Android
grep -rn 'LocalAccessibilityManager\|isReduceTransitions\|TRANSITION_ANIMATION_SCALE\|ANIMATOR_DURATION_SCALE\|areAnimatorsEnabled' \
  --include='*.kt' <scope>
```

**Zero matches across all three, in a project that animates, is a HIGH finding.** It is
the only check in this file that is decisive from the grep alone: absence of any handler
anywhere cannot be a false negative in the way a per-site check can.

### Remaining accessibility checks

| Check | Pattern | Severity when confirmed |
|-------|---------|-------------------------|
| Focus ring removed without replacement | `outline:\s*none\|outline:\s*0` — each hit needs a `:focus-visible` style nearby | HIGH |
| Click handler on a non-interactive element | `onClick` on `<div`/`<span` with no `role=` | HIGH |
| Decorative animation exposed to screen readers | animated component tags without `aria-hidden` | MEDIUM |

```bash
grep -rn 'outline:\s*none\|outline:\s*0' --include='*.css' --include='*.scss' --include='*.module.css' <scope>

grep -rn 'onClick' --include='*.tsx' --include='*.jsx' <scope> | grep -E '<div|<span' | grep -v 'role='

grep -rn '<motion\.\|<animated\.\|<Canvas' --include='*.tsx' --include='*.jsx' <scope> | grep -v 'aria-hidden'
```

Contrast ratio (4.5:1) is not greppable. Check it with the browser's own contrast
inspector, or an accessibility auditor already present in the project's toolchain —
do not install one for the audit. Check animated text mid-transition too: text fading
through low opacity must stay readable at every step above roughly 0.4.

---

## 3. Performance

| Check | Pattern | Severity when confirmed |
|-------|---------|-------------------------|
| Animating layout properties | `transition` naming `width`/`height`/`top`/`left`/`right`/`bottom`/`margin`/`padding` | HIGH |
| Animation loop on a timer instead of a frame callback | `setTimeout`/`setInterval` near motion code | MEDIUM |
| `will-change` sprawl | more than ~5 permanent declarations | LOW |

```bash
grep -rnE 'transition[^;]*(width|height|top|left|right|bottom|margin|padding)' \
  --include='*.css' --include='*.scss' --include='*.module.css' <scope>

grep -rn 'setTimeout\|setInterval' --include='*.ts' --include='*.tsx' --include='*.js' --include='*.jsx' <scope> \
  | grep -iE 'anim|motion|scroll|position|style|transform'

grep -rn 'will-change' --include='*.css' --include='*.scss' --include='*.module.css' <scope> | wc -l
```

The first is the mechanical form of the compositor-vs-layout rule in `SKILL.md`
§ Motion Accessibility and Cost. The third is a count, not a list: `will-change` is
correct in small numbers and wrong as a habit, so the finding is the total, and it
should be applied dynamically (set on interaction, removed when the animation ends)
rather than declared permanently.

**Library weight.** Where the project only does fades, slides, and springs, the
platform's own animation primitives cost nothing extra, while an animation library —
web or native — is a real payload measured in tens of KB on web and in hundreds of KB
to megabytes in a mobile binary. Treat a library as justified only by genuinely complex
pre-designed animation, not by fade-plus-slide. Measure the actual figure with whatever
bundle analyzer the project already has rather than quoting a remembered size; published
library sizes move between releases.

---

## 4. Consistency

This section is where an audit finds problems a per-component review cannot see: each
site is defensible on its own, and the defect is only visible across the whole set.

### Duration inventory

```bash
grep -rnoE 'duration[:"'\''= ]+[0-9.]+' --include='*.tsx' --include='*.jsx' --include='*.ts' \
  --include='*.css' --include='*.scss' <scope> | sed 's/.*duration//' | sort | uniq -c | sort -rn
```

### Easing inventory

```bash
# Web
grep -rnoE 'ease[A-Za-z-]*|cubic-bezier\([^)]+\)' --include='*.tsx' --include='*.jsx' \
  --include='*.ts' --include='*.css' --include='*.scss' <scope> | sed 's/^[^:]*:[0-9]*://' | sort | uniq -c | sort -rn

# Apple
grep -rnoE '\.spring\([^)]*\)|\.snappy|\.bouncy|\.smooth|\.easeIn|\.easeOut|\.interpolatingSpring' \
  --include='*.swift' <scope> | sed 's/^[^:]*:[0-9]*://' | sort | uniq -c | sort -rn

# Android
grep -rnoE 'spring\([^)]*\)|tween\([^)]*\)|FastOutSlowInEasing|LinearOutSlowInEasing|CubicBezierEasing' \
  --include='*.kt' <scope> | sed 's/^[^:]*:[0-9]*://' | sort | uniq -c | sort -rn
```

**Reading the inventory**: a coherent system uses 3-5 distinct durations and 3-5 named
easings. Beyond roughly 8 of either, the values are being chosen per-site rather than
drawn from a system — a MEDIUM finding whose fix is to centralize into named tokens, not
to adjust individual call sites.

This check pairs with `motion-principles.md` § 5: the archetype fixes *which* few values
the system should have; this inventory shows whether the codebase actually holds to them.

### Enter/exit symmetry

```bash
grep -rn -A5 'exit=' --include='*.tsx' --include='*.jsx' <scope>
```

Read each pair against the rule in `SKILL.md` § Motion and `motion-principles.md` § 6:

- Enter duration is longer than exit — entrances run 30-50% longer. The reverse is wrong.
- Enter decelerates (`ease-out`); exit accelerates (`ease-in`).
- Enter may be fully choreographed (translate + opacity + scale); exit should be simpler.

---

## 5. Reporting Findings

Findings from this suite use the severity and verdict model already defined in
`SKILL.md` § Review Modes — HIGH / MEDIUM / LOW, resolving to Block / Needs changes /
Approve. Do not introduce a second scale.

| Severity | Audit findings that land here |
|----------|-------------------------------|
| HIGH | No reduced-motion handler anywhere; `outline: none` with no `:focus-visible`; click handler on a non-interactive element; layout properties being animated |
| MEDIUM | Conditional render with no exit; hover with no transition; decorative animation missing `aria-hidden`; timer-driven animation loop; duration or easing sprawl past ~8 values |
| LOW | Staged list with no stagger; inline style change with no transition; `will-change` sprawl; enter/exit asymmetric in the wrong direction |

The finding caps in § Review Modes still apply: `quick` reports at most 5 and covers
HIGH and MEDIUM only; `full` reports at most 15. An audit that produces 60 raw grep hits
reports the most severe ones up to the cap and states the total it saw — never a
truncated list presented as the whole result.

### Native profiling entry points

Frame-level performance is not greppable. Where the project ships a native UI and the
audit needs real numbers, the platform's own profilers are the instrument — Android's
Layout Inspector and macrobenchmark tooling for recomposition and frame timing, Apple's
Instruments (time profiler, hitch detection) for the same. Target one frame budget:
16.67ms at 60fps, 8.33ms at 120fps. Web has the browser's own performance panel.

Report these as measurements taken or as a stated gap — never as an assumed pass.

---

## Cross-references

- `SKILL.md` § Review Modes — the severity scale and finding caps this suite reports into
- `SKILL.md` § Motion Accessibility and Cost — the reduced-motion and compositor rules §2 and §3 detect
- `SKILL.md` § Verification, § Red Flags — the prose checklists this file mechanizes
- `motion-principles.md` § 4, § 5, § 6 — the stagger budgets, archetype value sets, and timing ranges the §4 inventories are read against
