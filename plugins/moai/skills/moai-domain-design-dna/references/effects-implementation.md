# Effects Implementation

How a `visual_effects` entry becomes running code. Read this before
implementing any effect in Phase 3.

## Technology follows the performance tier

`overview.performance_tier` is the selector. It is recorded during extraction
precisely so the generator does not reach for a 600 KB renderer to draw a
gradient.

| Tier | Reach for | Signals it in the reference |
|---|---|---|
| lightweight | CSS animations, SVG (SMIL or CSS-driven), plain JS | Effects expressible as transform, opacity, filter, gradient |
| medium | Canvas 2D, a timeline animation library, vector-animation playback | Per-pixel or per-particle drawing, sequenced multi-element choreography |
| heavy | WebGL, a 3D scene library, custom GLSL | Real 3D geometry, lighting, per-fragment shading, thousands of particles |

Two directions of error, both common. Reaching one tier too high pays a large
bundle and a class of device failures for something CSS already does. Reaching
one tier too low produces an approximation that reads as a worse version of the
reference — which is more damaging than omitting the effect, because it draws
attention to what it fails to be.

Heavy-tier libraries load from a CDN in a self-contained artifact. Pin a
version rather than tracking a moving `latest`: a self-contained file is often
opened months later, and a renderer that silently majored underneath it fails
with a blank canvas and no diagnostic.

## Per-category patterns

**Background effects.** Gradient animation is CSS keyframes over a linear or
conic gradient — no canvas needed. A noise field is Canvas 2D with a coherent
noise function. Mesh gradients interpolate on canvas. Video backgrounds need
`muted` for autoplay to be permitted, plus a poster frame for the interval
before the first decoded frame.

**Particle systems.** Below roughly a hundred particles with simple motion, a
hand-written Canvas 2D loop is smaller and clearer than a library. Above that,
or with inter-particle interaction, a GPU-backed renderer earns its weight.
Map `interaction` to pointer handlers, and register teardown — a particle loop
that outlives its container is a leak that only shows up after navigation.

**3D elements.** Apply the recorded lighting, camera, and material params
before adding post-processing; a bloom pass over a badly lit scene reads as
haze. Handle resize through `ResizeObserver` rather than the window resize
event, so the canvas tracks its container rather than the viewport.

**Shaders.** Build vertex and fragment programs from the recorded `type`, pass
time / resolution / pointer as uniforms, and advance time inside the animation
frame callback. Match the noise implementation to the recorded `noise_type` —
the four named functions produce visibly different textures and are not
interchangeable.

**Scroll effects.** Parallax is a translate driven by scroll offset times a
per-layer speed. Trigger-once reveals belong to `IntersectionObserver` with an
explicit threshold, not to a scroll listener. Scrubbed animation maps progress
directly to scroll position and must remain correct when the user jumps — a
scrubbed timeline that only advances on incremental scroll breaks on anchor
navigation.

**Text effects.** Splitting text into per-character or per-word elements
destroys its accessible reading order unless the original string is preserved
for assistive technology. Gradient fill is `background-clip: text`. Glitch is
layered clip paths with colour offset. Typewriter reveal is either stepped CSS
or a JS interval over character count.

**Cursor effects.** Hiding the system cursor removes a primary affordance, so
the replacement must track without perceptible lag and must never apply to
coarse pointers. Magnetic buttons transform on proximity; spotlight is a radial
mask following the pointer; trails spawn short-lived fading elements.

**Glass and neumorphism.** Glass is `backdrop-filter` blur over a translucent
background — verify text contrast against the *actual* backdrop, since a value
that passes over one background fails over another. Neumorphism is a paired
light and dark shadow offset, and it is the treatment most likely to fail
contrast outright.

**Canvas drawings and SVG animation.** Size the canvas to its container in
device pixels, not CSS pixels, or the output is soft on high-density displays.
Path drawing animates `stroke-dashoffset` from the measured path length to
zero. Shape morphing interpolates path data between forms with matching node
counts.

## Fallback is part of the effect

The recorded `fallback_strategy` is implemented, not merely stored. Two
conditions gate it: a reduced-motion preference, and a device that cannot carry
the tier. Both resolve before the effect initialises — an effect that starts
and is then torn down has already cost the frame drop it was meant to avoid.

Honouring reduced motion means the interface still works and still communicates
its hierarchy with the motion removed. Freezing an animation mid-state, or
removing the element the animation was revealing, is not honouring the
preference.

## Hygiene

- `requestAnimationFrame` drives every animation loop. `setInterval` drifts
  against the display refresh and keeps running in background tabs.
- Cancel the frame and release the context on teardown.
- One loop per surface. Several independent loops on one page compete for the
  same frame budget and produce jitter that profiles as "slow" without any
  single culprit.
- Contrast is verified against the rendered result, not against the token
  values, wherever an effect sits between the text and its background.

Component-level finish — easing curve choice, hover-state timing, hit-area
sizing — belongs to `moai-ref-ui-polish`. This file covers only what those
rules do not reach: effects that need a renderer.
