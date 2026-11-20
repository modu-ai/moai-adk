# Design DNA Schema

The three-dimension field list and the enum vocabularies each field draws from.
Read this before Phase 1 (present the schema) or Phase 2 (extract a profile).

The JSON is a flat, human-readable document — three top-level keys, one per
dimension. Every field is populated; where a reference does not exercise a
field, the value says so explicitly rather than being left empty.

## Dimension 1 — `design_system` (measurable)

Concrete, token-level values. Extract exact values where visible; estimate from
visual inspection otherwise.

| Group | Fields |
|---|---|
| `color` | `palette_type`, `primary` / `secondary` / `accent` (each `{hex, role}`), `neutral_scale` (ordered ramp), `semantic` (success / warning / error / info), `surface` (background, raised, sunken), `contrast_strategy` |
| `typography` | `font_families` (heading / body / mono), `type_scale` — display, h1, h2, h3, body, body_small, caption, overline, each `{size, weight, line_height, tracking}` — plus `font_style_notes` |
| `spacing` | `base_unit`, `scale`, `content_density`, `section_rhythm` |
| `layout` | `grid_system`, `max_content_width`, `columns`, `gutter`, `breakpoints`, `alignment_tendency` |
| `shape` | `border_radius` (small / medium / large / pill), `border_usage`, `divider_style` |
| `elevation` | `shadow_style`, `levels` (low / medium / high), `depth_cues` |
| `iconography` | `style`, `stroke_weight`, `size_scale`, `preferred_set` |
| `motion` | `easing`, `duration_scale` (micro / normal / macro), `entrance` / `exit` pattern, `philosophy` |
| `components` | `button`, `input`, `card`, `nav`, `modal`, `list` — style plus observation notes |

Enum vocabularies:

- `color.palette_type` — monochromatic, complementary, analogous, triadic, split-complementary
- `color.contrast_strategy` — high contrast, subtle layers, dark-on-light dominant, light-on-dark dominant
- `spacing.content_density` — compact, comfortable, spacious
- `layout.alignment_tendency` — strict grid, centered, asymmetric, mixed
- `shape.border_usage` — none, subtle 1px, bold borders, inputs only
- `elevation.shadow_style` — none, soft diffused, hard drop, layered
- `elevation.depth_cues` — shadows, overlapping layers, blur / glass, colour intensity
- `motion.philosophy` — minimal functional, playful bouncy, cinematic, none

`components` takes observation prose rather than an enum, because the
distinguishing detail is usually specific: "ghost buttons with thick borders,
rounded inputs carrying an inner shadow" transfers; "primary button" does not.

## Dimension 2 — `design_style` (qualitative)

Subjective assessments in descriptive language. These fields guide the
judgement calls the token values leave open.

| Group | Fields |
|---|---|
| `aesthetic` | `mood` (3-5 words), `visual_metaphor`, `era_influence`, `genre`, `personality_traits`, `adjectives` |
| `visual_language` | `complexity`, `ornamentation`, `whitespace_usage`, `visual_weight_distribution`, `focal_strategy`, `contrast_level`, `texture_usage` |
| `composition` | `hierarchy_method`, `balance_type`, `flow_direction`, `grouping_strategy`, `negative_space_role` |
| `imagery` | `photo_treatment`, `illustration_style`, `graphic_elements`, `pattern_usage`, `image_shape` |
| `interaction_feel` | `feedback_style`, `hover_behavior`, `transition_personality`, `loading_style`, `microinteraction_density` |
| `brand_voice_in_ui` | `tone`, `formality`, `cta_style`, `empty_state_approach`, `error_tone` |

Enum vocabularies:

- `visual_language.complexity` — minimal, moderate, rich, maximal
- `visual_language.ornamentation` — none, subtle accents, decorative, heavily ornamented
- `visual_language.focal_strategy` — single hero element, distributed interest, progressive reveal
- `composition.hierarchy_method` — scale contrast, colour weight, spatial isolation, typographic hierarchy
- `composition.balance_type` — symmetric, asymmetric, radial, mosaic
- `interaction_feel.transition_personality` — snappy, smooth glide, bouncy elastic, fade-subtle
- `brand_voice_in_ui.cta_style` — direct imperative, friendly invitation, urgent scarcity, subtle suggestion

`aesthetic.genre` is free text against archetypes — corporate SaaS, indie
creative, luxury editorial, neo-brutalist, technical documentation, and so on.
Naming the archetype is what lets a generator make a hundred small decisions
consistently without each one being specified.

`aesthetic.personality_traits` is written as if the design were a person:
confident, approachable, meticulous. This reads as whimsy and is not — it is
the field that decides whether a hover state is eager or restrained.

## Dimension 3 — `visual_effects` (special rendering)

Each category is an object carrying at minimum `enabled`, `type`,
`description`, and a `params` map. A category the reference does not use gets
`enabled: false` and is not implemented.

| Category | `type` vocabulary |
|---|---|
| `overview` | not a category — see below |
| `background_effects` | gradient-animation, noise-field, mesh-gradient, video-bg, generative-art, none |
| `particle_systems` | floating-dots, confetti, snow, fireflies, connected-nodes, custom |
| `3d_elements` | hero-model, product-viewer, scene-bg, text-extrusion, abstract-geometry |
| `shader_effects` | noise-distortion, wave, morph, color-shift, custom-GLSL |
| `scroll_effects` | parallax (`layers`), scroll_triggered_animations (`animation_type`), scroll_morphing |
| `text_effects` | split-letter-animate, typewriter, glitch, gradient-fill, 3d-extrude, none |
| `cursor_effects` | custom-cursor, magnetic-buttons, spotlight, trail, none |
| `image_effects` | hover-distortion, reveal-clip, parallax-tilt, rgb-shift, none |
| `glassmorphism_neumorphism` | glass, neumorphic-light, neumorphic-dark, frosted-layers, none |
| `canvas_drawings` | generative-lines, interactive-blobs, data-visualization, pattern-fill, none |
| `svg_animations` | path-draw, morph-shapes, logo-reveal, decorative-loop, none |
| `composite_notes` | free text — layering, trade-offs, screenshot-only observations |

`overview` sits above the categories and drives implementation:

- `effect_intensity` — none, subtle-accent, moderate, heavy-immersive
- `performance_tier` — lightweight, medium, heavy (the technology selector; see `effects-implementation.md`)
- `fallback_strategy` — disable effects, reduce to CSS, static snapshot
- `primary_technology` — CSS only, Canvas 2D, WebGL, GSAP, Lottie, SVG SMIL, or another named runtime

Selected `params` vocabularies worth recording precisely, because they change
the implementation rather than merely describing it:

- `particle_systems.params.interaction` — mouse-repel, mouse-attract, click-burst, none
- `particle_systems.params.count` — decides whether a hand-written loop suffices
- `shader_effects.params.noise_type` — perlin, simplex, worley, fbm
- `scroll_effects.scroll_triggered_animations.animation_type` — fade-up, scale-in, clip-reveal, counter, draw-SVG
- `text_effects.params.split_strategy` — by-char, by-word, by-line
- `image_effects.params.distortion_type` — barrel, wave, liquid, glitch
- `3d_elements.params` — lighting, camera, materials, post_processing

`composite_notes` is where an honest profile shows its quality. Effects layer,
and a screenshot rarely reveals which of three plausible techniques produced a
given surface. Recording the ambiguity keeps the generator from committing to a
guess that the reference never made.
