#!/usr/bin/env python3
"""
Generate TRUST 5 Pentagon diagram for MoAI-ADK README
Using Nano Banana Pro (Gemini 3 Pro Image Preview)
"""

import os
import sys
from pathlib import Path
from datetime import datetime

# Add skill modules to path
skill_path = Path("/Users/goos/MoAI/MoAI-ADK/.claude/skills/moai-connector-nano-banana/modules")
sys.path.insert(0, str(skill_path))

from image_generator import NanoBananaImageGenerator


def main():
    """Generate trust5-pentagon.png with professional quality badge style."""

    # API key from .env
    api_key = os.getenv("GEMINI_API_KEY")
    if not api_key:
        print("Error: GEMINI_API_KEY not found in environment")
        sys.exit(1)

    # Initialize generator
    generator = NanoBananaImageGenerator(api_key=api_key)

    # Optimized prompt for TRUST 5 Pentagon
    prompt = """
A professional quality badge style pentagon diagram on cream-colored paper texture.
The pentagon has 5 vertices, each labeled with one letter: T, R, U, S, T.
Each vertex has a distinctive icon representing its principle:

T (Testability): Small test tube or checkmark icon
R (Reliability): Solid shield icon
U (Usability): Hand holding mobile phone icon
S (Security): Lock or padlock icon
T (Traceability): Linked chain or document trail icon

The pentagon is hand-drawn style with clean, professional lines in dark navy blue.
Each vertex is connected to form a complete pentagon shape.
The icons are simple, modern, and minimalist in design, rendered in complementary colors:
orange for Testability, blue for Reliability, green for Usability,
red for Security, and purple for Traceability.

The background is a warm cream-colored paper texture with subtle grain.
The overall aesthetic is professional, clean, and suitable for documentation.
Quality: high-resolution infographic style, sharp details.
Composition: centered pentagon with balanced spacing.
Style: modern technical documentation quality badge.
Format: 16:9 aspect ratio, horizontal orientation.
"""

    # Output path
    output_dir = Path("/Users/goos/MoAI/MoAI-ADK/assets/images/readme")
    output_dir.mkdir(parents=True, exist_ok=True)
    output_path = output_dir / "trust5-pentagon.png"

    print("🎨 Generating TRUST 5 Pentagon image...")
    print(f"📸 Resolution: 2K (2048px)")
    print(f"📐 Aspect Ratio: 16:9")
    print(f"🎯 Style: Professional quality badge on cream paper")
    print()

    try:
        # Generate image
        start_time = datetime.now()

        image, metadata = generator.generate(
            prompt=prompt,
            model="pro",  # Nano Banana Pro for best quality
            aspect_ratio="16:9",
            save_path=str(output_path)
        )

        duration = (datetime.now() - start_time).total_seconds()

        print("✅ Image generation completed!")
        print()
        print("📊 Generation Details:")
        print(f"   - Processing Time: {duration:.1f} seconds")
        print(f"   - Model: Nano Banana Pro (gemini-3-pro-image-preview)")
        print(f"   - SynthID Watermark: Included")
        print(f"   - Thinking Process: Enabled")
        print()
        print(f"💾 Saved to: {output_path}")
        print()
        print("🎯 Image Features:")
        print("   - Hand-drawn pentagon with 5 vertices (T-R-U-S-T)")
        print("   - Distinctive icons for each principle")
        print("   - Cream paper texture background")
        print("   - Professional quality badge style")
        print("   - 16:9 aspect ratio (horizontal)")

    except Exception as e:
        print(f"❌ Error generating image: {e}")
        import traceback
        traceback.print_exc()
        sys.exit(1)


if __name__ == "__main__":
    main()
