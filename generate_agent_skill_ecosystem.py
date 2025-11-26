#!/usr/bin/env python3
"""
Generate hand-drawn agent-skill ecosystem image for MoAI-ADK
"""
import sys
import os
from pathlib import Path

# Add skill modules to path
skill_path = Path("/Users/goos/MoAI/MoAI-ADK/.claude/skills/moai-connector-nano-banana/modules")
sys.path.insert(0, str(skill_path))

from image_generator import NanoBananaImageGenerator


def create_optimized_prompt() -> str:
    """Create Nano Banana Pro optimized prompt for agent-skill ecosystem."""

    prompt = """A hand-drawn organizational pyramid diagram on warm cream-colored paper (#F5F1E8), illustrating a 5-tier agent hierarchy system for MoAI-ADK software framework.

The pyramid structure displays five distinct tiers from top to bottom, with each tier progressively wider, creating a classic hierarchical visualization:

Tier 1 (Top, narrowest): Seven small hand-drawn stick figure person icons with circular specialty badges floating beside them. Each figure is labeled in neat handwriting: "backend", "frontend", "database", "devops", "security", "uiux", and "debug". These figures are accented with soft powder blue watercolor-style highlights, giving a technical specialist appearance.

Tier 2: Eight stick figure person icons, each wearing simple hand-drawn manager hats or carrying briefcase symbols. Labels in brown handwriting read: "project", "spec", "tdd", "docs", "strategy", "quality", "git", and "claude-code". Soft sage green watercolor accents highlight this management tier, creating a coordinating visual layer.

Tier 3 (Middle, central tier): Three stick figure person icons wearing tool belts with small wrench and hammer symbols. Labels read: "agent", "skill", "command" in the same warm brown handwriting. Orange watercolor accents distinguish this builder tier, suggesting construction and creation.

Tier 4: Five stick figure person icons with plug or connection symbols drawn beside them, representing integration. Labels read: "docs", "design", "notion", "browser", and "ultrathink". Purple watercolor accents create a technological, connected aesthetic for this integration layer.

Tier 5 (Bottom, widest): One larger, friendly robot icon with simple geometric shapes (rectangle body, circular head, antenna). Labeled "nano-banana" in prominent handwriting. Bright golden yellow watercolor accent makes this AI service layer stand out as the foundation.

Visual connections: Thin, organic wavy lines drawn in warm brown (#8B6F47) connect the tiers, flowing naturally between levels to show relationships and dependencies. The lines are slightly imperfect, maintaining the hand-drawn aesthetic. Small decorative elements scattered throughout include tiny hand-drawn stars, dots, and small spiral doodles, adding personality without clutter.

Each tier has a slightly different background shade, progressively getting warmer and slightly darker from top to bottom, creating subtle depth: lightest cream at top, deepening to a slightly warmer cream at the bottom.

All stick figures are drawn with simple, friendly features: circular heads, straight line bodies, stick arms and legs. Faces have minimal details - just simple dots for eyes and curved lines for smiles. The style is reminiscent of visual note-taking and sketchnote illustration, balancing professionalism with approachability.

Lighting: Soft, even natural daylight from above-left, as if photographed on a desk near a window. Gentle shadows cast by the paper texture create subtle depth. No harsh highlights, maintaining a soft, educational diagram aesthetic.

Camera: Straight-on view, perpendicular to the paper surface, with very slight perspective showing the pyramid's three-dimensional depth. Shot with a 50mm portrait lens at f/4 to maintain natural proportions and slight depth of field on the edges.

Composition: Pyramid perfectly centered in the 16:9 frame with balanced white space on all sides. The top tier (Tier 1) starts about 15% from the top, and the bottom tier (Tier 5) ends about 15% from the bottom. Clean, organized layout with clear visual hierarchy flowing downward. Rule of thirds applied for balanced composition.

Color palette: Cream paper background (#F5F1E8), warm brown lines and handwritten text (#8B6F47), soft powder blue (Tier 1 accents), sage green (Tier 2 accents), warm orange (Tier 3 accents), purple (Tier 4 accents), golden yellow (Tier 5 accent). All colors are muted and professional, watercolor-style with soft edges.

Art style: Hand-drawn sketch note aesthetic, inspired by visual facilitation and infographic design. Professional documentation illustration with friendly, organic qualities. Mix of structured organizational diagram with playful hand-drawn elements. Line quality is deliberately imperfect - slightly wobbly, organic lines that feel authentic and human-created. Think "professional whiteboard sketch" or "design thinking workshop visualization".

Mood: Professional yet approachable, clear and educational, organized yet creative, friendly and inviting. The image should feel like a helpful diagram from a well-designed technical book or design workshop documentation.

Typography: All text in consistent brown handwriting style. Letters are uppercase for tier labels, lowercase for agent names. Handwriting is clear, legible, slightly imperfect with natural human variation. Font style resembles careful marker or fine-tip pen handwriting.

Quality: Studio-grade professional documentation image, high-resolution output optimized for README markdown documentation. Clean, crisp lines with authentic hand-drawn texture. Professional photography of analog sketch artwork. 16:9 aspect ratio for wide documentation display. PNG format with cream background (#F5F1E8).

Technical specifications: 4K resolution for maximum detail preservation. Professional documentation photography standard. Suitable for technical documentation, README files, presentation slides, and educational materials."""

    return prompt


def main():
    """Generate the agent-skill ecosystem image."""

    # Check API key
    api_key = os.getenv("GEMINI_API_KEY") or os.getenv("GOOGLE_API_KEY")
    if not api_key:
        print("❌ Error: API key not found!")
        print("   Please set GEMINI_API_KEY or GOOGLE_API_KEY environment variable")
        sys.exit(1)

    # Initialize generator
    print("🎨 Initializing Nano Banana Pro Image Generator...")
    generator = NanoBananaImageGenerator(api_key=api_key)

    # Create optimized prompt
    print("📝 Creating optimized prompt...")
    prompt = create_optimized_prompt()

    # Output path
    output_path = "/Users/goos/MoAI/MoAI-ADK/assets/images/readme/agent-skill-ecosystem.png"

    print(f"\n{'='*70}")
    print("🍌 Generating Hand-Drawn Agent-Skill Ecosystem")
    print(f"{'='*70}")
    print(f"📍 Output: {output_path}")
    print(f"📐 Aspect Ratio: 16:9")
    print(f"🎯 Model: Nano Banana Pro (gemini-3-pro-image-preview)")
    print(f"{'='*70}\n")

    # Generate image
    try:
        image, metadata = generator.generate(
            prompt=prompt,
            model="pro",
            aspect_ratio="16:9",
            save_path=output_path
        )

        print(f"\n{'='*70}")
        print("✅ SUCCESS! Image Generated")
        print(f"{'='*70}")
        print(f"📁 Saved to: {output_path}")
        print(f"⚙️  Model: {metadata['model_name']}")
        print(f"📐 Aspect Ratio: {metadata['aspect_ratio']}")
        print(f"🔢 Tokens Used: {metadata['tokens_used']}")
        print(f"⏱️  Timestamp: {metadata['timestamp']}")
        print(f"{'='*70}\n")

        print("🎉 Hand-drawn agent-skill ecosystem image ready for documentation!")

    except Exception as e:
        print(f"\n❌ Error generating image: {e}")
        sys.exit(1)


if __name__ == "__main__":
    main()
