# Prompt Engineering for Nano Banana

## Overview

Professional prompt engineering patterns specifically optimized for Google Gemini Image Generation API (Nano Banana Pro & Flash).

---

## Fundamental Principles

### 1. Specificity Over Generality

**Poor**:
```
A landscape
```

**Good**:
```
A breathtaking mountain landscape at golden hour with snow-capped peaks
reflected in a pristine alpine lake. Foreground features wildflowers.
Professional nature photography, 4K quality, photorealistic.
```

**Why**: Specific prompts provide clear direction for the AI model, resulting in higher-quality, more predictable outputs.

---

### 2. Technical Language for Precision

**Poor**:
```
A photo of a person
```

**Good**:
```
Professional portrait with 85mm lens, shallow depth of field (f/1.8),
soft natural lighting from 45-degree angle, medium close-up composition.
```

**Why**: Photography and cinematography terminology helps the model understand exact technical requirements.

---

### 3. Layered Prompt Structure

Build prompts in logical layers:

```
[Layer 1: Subject & Action]
A vintage bicycle leaning against a brick wall

[Layer 2: Environment]
in a charming European alleyway with cobblestones

[Layer 3: Lighting & Mood]
Morning light filtering through overhead vines, creating dappled shadows

[Layer 4: Technical Specs]
Photorealistic, 50mm lens, shallow depth of field (f/2.8)

[Layer 5: Style & Quality]
Warm color grading, cinematic quality, Instagram-worthy
```

---

## Core Prompt Patterns

### Pattern 1: Product Photography

**Template**:
```
Professional product photo of [PRODUCT] on [SURFACE].
[LIGHTING_TYPE] from [ANGLE], creating [EFFECT].
Camera: [SHOT_TYPE] with [LENS].
Depth of field: [SHALLOW/DEEP] (f/[NUMBER]).
Color palette: [COLORS].
Style: [PHOTOGRAPHY_STYLE].
Quality: [RESOLUTION], [QUALITY_DESCRIPTOR].
```

**Example**:
```
Professional product photo of wireless headphones on a clean white surface.
Soft studio lighting from 45-degree angle, creating subtle shadows.
Camera: Medium close-up with 85mm lens.
Depth of field: Shallow (f/2.8) with blurred background.
Color palette: Silver metallic, white, subtle blue accents.
Style: Commercial photography, high-end retail quality.
Quality: 4K, crystal clear focus.
```

---

### Pattern 2: Portrait Photography

**Template**:
```
[TYPE] portrait of [SUBJECT] wearing [CLOTHING/ACCESSORIES].
[SETTING] with [ENVIRONMENT_DETAILS].

Lighting: [LIGHT_TYPE] from [DIRECTION], creating [MOOD].
Camera: [SHOT_TYPE] with [LENS].
Composition: [POSITIONING_DETAILS].
Depth of field: [SHALLOW/DEEP] (f/[NUMBER]).
Color palette: [COLORS].
Style: [PHOTOGRAPHY_STYLE].
Quality: [RESOLUTION], [FOCUS_DETAILS].
```

**Example**:
```
Professional portrait of a confident female CEO in her 40s wearing blue glasses
and a dark blazer. Standing in a modern office with natural light.

Lighting: Soft natural window light from the left, creating gentle shadows.
Camera: Medium close-up portrait with 85mm lens.
Composition: Subject positioned slightly off-center, looking at camera.
Depth of field: Shallow (f/1.8) with beautifully blurred office background.
Color palette: Warm skin tones, cool blue from glasses, neutral office colors.
Style: Professional corporate photography, LinkedIn quality.
Quality: 2K resolution, sharp focus on eyes.
```

---

### Pattern 3: Landscape/Nature Photography

**Template**:
```
A [ADJECTIVE] [LANDSCAPE_TYPE] at [TIME_OF_DAY] with [KEY_FEATURES].
[FOREGROUND_DETAILS].

Lighting: [LIGHT_QUALITY] from [DIRECTION], creating [ATMOSPHERE].
Camera: [SHOT_TYPE] with [LENS].
Composition: [COMPOSITIONAL_RULE], [POSITIONING].
Depth of field: [SHALLOW/DEEP] (f/[NUMBER]).
Color palette: [COLORS].
Style: [PHOTOGRAPHY_STYLE].
Quality: [RESOLUTION], [QUALITY_DESCRIPTOR].
```

**Example**:
```
A breathtaking mountain landscape at golden hour with snow-capped peaks
reflected in a pristine alpine lake. Foreground features wildflowers.

Lighting: Warm golden hour light from the side, creating depth.
Camera: Wide-angle shot with 24mm lens.
Composition: Rule of thirds, lake in lower third.
Depth of field: Deep focus (f/16) for full sharpness.
Color palette: Warm golden tones, deep blue sky, jade green water.
Style: Professional nature photography, National Geographic quality.
Quality: 4K resolution, photorealistic.
```

---

## Image Editing Prompt Patterns

### Pattern 4: Style Transfer

**Template**:
```
Transform this [SUBJECT] into [TARGET_STYLE] style.
[STYLE_DETAILS].
Keep [PRESERVE_ELEMENTS].
Apply [ARTISTIC_EFFECTS].
Color palette: [COLORS].
```

**Example**:
```
Transform this city street photo into a Van Gogh oil painting style.
Swirling brushstrokes and vibrant colors.
Keep the composition the same.
Apply expressive artistic interpretation.
Color palette: Bold blues, yellows, and greens typical of Van Gogh.
```

---

### Pattern 5: Background Replacement

**Template**:
```
Change the background from [CURRENT] to [NEW_ENVIRONMENT].
Keep [PRESERVE_ELEMENTS] in the same [CHARACTERISTICS].
[NEW_ENVIRONMENT_DETAILS].
Maintain [CONSISTENCY_REQUIREMENTS].
```

**Example**:
```
Change the background from this office to a modern coffee shop interior.
Keep the person in the exact same pose and lighting.
Coffee shop should have warm wood tones, plants, and soft ambient lighting.
Maintain photorealistic style and natural color consistency.
```

---

### Pattern 6: Object Addition/Removal

**Template**:
```
[ADD/REMOVE] [OBJECTS] from this [SCENE].
[ADDITION_DETAILS or REMOVAL_DETAILS].
Keep [CONSISTENCY_REQUIREMENTS].
```

**Example (Addition)**:
```
Add a sleek laptop, white coffee mug, and small potted plant to this desk.
Arrange them naturally with proper shadows and reflections.
Keep the existing desk surface and lighting consistent.
```

**Example (Removal)**:
```
Remove the telephone poles and power lines from this landscape photo.
Fill the removed areas naturally with sky and background elements.
Maintain the natural look and lighting of the original scene.
```

---

## Advanced Techniques

### Technique 1: Compositional Language

Use photography composition terms:

- **Rule of thirds**: Subject positioned at intersection points
- **Leading lines**: Lines that guide the eye
- **Negative space**: Empty space around subject
- **Framing**: Using elements to frame the subject
- **Symmetry**: Balanced composition
- **Golden ratio**: Natural proportion in composition

**Example**:
```
A single red apple centered on a white marble surface with generous negative space.
Minimalist composition, soft top-down lighting, high-key photography style.
Rule of thirds positioning, clean and simple.
```

---

### Technique 2: Lighting Vocabulary

Precise lighting descriptions:

**Natural Light**:
- Golden hour: Warm, soft light during sunrise/sunset
- Blue hour: Cool, diffuse light before sunrise/after sunset
- Harsh midday: Direct, strong light creating hard shadows
- Window light: Soft, directional indoor natural light
- Overcast: Soft, even, diffused light

**Artificial Light**:
- Studio lighting: Controlled, professional setup
- Softbox: Large, diffused light source
- Rim lighting: Backlighting creating edge highlight
- Three-point lighting: Key, fill, and back lights
- Ambient lighting: General environmental light

**Example**:
```
Portrait with soft window light from the left (key light),
subtle fill from white reflector on right, and natural ambient lighting.
Creates gentle shadows, flattering skin tones, professional quality.
```

---

### Technique 3: Depth of Field Control

**Shallow DOF** (f/1.4 - f/2.8):
- Subject sharp, background blurred
- Best for portraits, product close-ups
- Creates bokeh effect
- Isolates subject

**Medium DOF** (f/4 - f/8):
- Subject and near background sharp
- Balanced depth
- Good for environmental portraits
- Versatile general use

**Deep DOF** (f/11 - f/22):
- Everything in focus
- Best for landscapes
- Maximum detail throughout
- Architectural photography

**Example**:
```
Portrait with shallow depth of field (f/1.8).
Subject's eyes tack-sharp, face fully in focus.
Background office environment beautifully blurred with creamy bokeh.
85mm portrait lens equivalent.
```

---

### Technique 4: Color Palette Design

**Warm Palettes**:
```
Color palette: Warm golden tones, amber, honey, cream, soft oranges.
Creates inviting, cozy, friendly atmosphere.
```

**Cool Palettes**:
```
Color palette: Cool blues, steel gray, ice white, mint green.
Creates professional, clean, modern atmosphere.
```

**Complementary**:
```
Color palette: Blue and orange complementary contrast.
Dynamic, energetic, eye-catching composition.
```

**Monochromatic**:
```
Color palette: Shades of blue from navy to pale sky.
Harmonious, cohesive, sophisticated look.
```

---

### Technique 5: Mood and Atmosphere

**Serene/Peaceful**:
```
Serene and peaceful atmosphere. Soft, gentle lighting.
Muted tones, calm composition, minimal distractions.
Evokes tranquility and relaxation.
```

**Dramatic/Intense**:
```
Dramatic and intense atmosphere. Strong contrast lighting.
Bold colors, dynamic composition, powerful presence.
Evokes emotion and energy.
```

**Professional/Clean**:
```
Professional and clean atmosphere. Bright, even lighting.
Neutral tones, organized composition, polished look.
Evokes trust and competence.
```

---

## Model-Specific Optimization

### For Gemini 3 Pro Image (Nano Banana Pro)

**Strengths**:
- Complex compositions
- High-resolution output (up to 4K)
- Real-time Google Search grounding
- Advanced text rendering
- "Thinking" process optimization

**Prompt Strategy**:
- Use detailed, layered prompts
- Specify high-quality requirements (4K, photorealistic)
- Include technical photography terms
- Request Google Search integration when needed
- Expect slower generation (10-60s)

**Example**:
```
Professional commercial photograph of luxury watch on black velvet.
Dramatic studio lighting with subtle highlights revealing intricate details.
Macro lens, ultra-sharp focus, 4K resolution.
Color palette: Platinum silver, deep black, subtle gold accents.
Style: High-end jewelry photography, magazine quality.
Enable Google Search for current luxury watch design trends.
```

---

### For Gemini 2.5 Flash Image (Nano Banana)

**Strengths**:
- Fast generation (5-15s)
- Batch processing efficiency
- Cost-effective
- Good for iterations
- 1K resolution

**Prompt Strategy**:
- Use concise, clear prompts
- Focus on essential details only
- Simpler compositions work best
- Great for quick prototyping
- Ideal for variations testing

**Example**:
```
Modern office illustration with people working at desks.
Colorful flat design style, clean and simple.
Bright colors, friendly atmosphere, professional setting.
```

---

## Common Mistakes to Avoid

### 1. Conflicting Instructions

❌ **Wrong**: "Photorealistic cartoon style with abstract elements"
✅ **Correct**: Choose one consistent style - either "Photorealistic" OR "Cartoon illustration"

### 2. Vague Subject Description

❌ **Wrong**: "A person"
✅ **Correct**: "A confident female CEO in her 40s wearing blue glasses and dark blazer"

### 3. Missing Technical Specifications

❌ **Wrong**: "A nice photo"
✅ **Correct**: "Professional portrait, 85mm lens, f/1.8, soft natural lighting, 2K resolution"

### 4. Overloading with Details

❌ **Wrong**: "A red car, blue sky, green trees, yellow flowers, white clouds, brown road, orange sunset, purple mountains, pink clouds..."
✅ **Correct**: "A red sports car on a winding mountain road at sunset. Warm golden hour lighting."

### 5. Unclear Editing Instructions

❌ **Wrong**: "Make it better"
✅ **Correct**: "Increase brightness by 20%, add warmer color tones, sharpen focus on main subject"

---

## Prompt Quality Checklist

**Essential Elements**:
- [ ] Clear subject definition
- [ ] Setting/environment description
- [ ] Lighting specification
- [ ] Camera angle/lens details
- [ ] Depth of field requirement
- [ ] Color palette description
- [ ] Style/quality requirements
- [ ] Resolution specification

**Optional Enhancements**:
- [ ] Mood/atmosphere description
- [ ] Compositional rules
- [ ] Specific artistic references
- [ ] Technical photography terms
- [ ] Google Search integration (Pro only)

---

## Testing and Iteration

### Iterative Refinement Process

**Step 1**: Start with base prompt
```
A mountain landscape at sunset
```

**Step 2**: Add technical details
```
A breathtaking mountain landscape at golden hour with snow-capped peaks.
Wide-angle 24mm lens, deep focus (f/16).
```

**Step 3**: Enhance with lighting and composition
```
A breathtaking mountain landscape at golden hour with snow-capped peaks.
Warm golden hour light from the side, creating depth.
Wide-angle 24mm lens, deep focus (f/16).
Composition using rule of thirds.
```

**Step 4**: Add quality and style
```
A breathtaking mountain landscape at golden hour with snow-capped peaks.
Warm golden hour light from the side, creating depth.
Wide-angle 24mm lens, deep focus (f/16).
Composition using rule of thirds.
Color palette: Warm golden tones, deep blue sky.
Style: Professional nature photography, National Geographic quality.
Quality: 4K resolution, photorealistic.
```

---

## Best Practices Summary

1. **Be Specific**: Provide clear, detailed subject descriptions
2. **Use Technical Terms**: Photography/cinematography vocabulary
3. **Layer Your Prompts**: Build from general to specific
4. **Specify Quality**: Resolution, style, technical requirements
5. **Describe Lighting**: Type, direction, mood
6. **Include Composition**: Camera angle, framing, positioning
7. **Choose Right Model**: Pro for quality, Flash for speed
8. **Test and Iterate**: Refine through multi-turn conversations
9. **Maintain Consistency**: Avoid conflicting instructions
10. **Learn from Examples**: Study successful prompts

---

**Last Updated**: 2025-11-27
**Status**: Production Ready
**Model Compatibility**: Gemini 3 Pro Image, Gemini 2.5 Flash Image
