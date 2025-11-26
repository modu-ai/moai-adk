#!/usr/bin/env python3
"""Generate command cheatsheet card using Nano Banana Pro."""

import os
import sys
from pathlib import Path
from datetime import datetime
from dotenv import load_dotenv

# Load environment variables
load_dotenv()

# Add skill module to path
skill_path = Path(".claude/skills/moai-connector-nano-banana/modules")
sys.path.insert(0, str(skill_path))

from image_generator import NanoBananaImageGenerator

def main():
    """Generate command cheatsheet card."""

    # Load API key (try both GOOGLE_API_KEY and GEMINI_API_KEY)
    api_key = os.getenv("GOOGLE_API_KEY") or os.getenv("GEMINI_API_KEY")
    if not api_key:
        print("❌ Error: API key not found in environment")
        print("\nSetup instructions:")
        print("1. Create .env file in project root")
        print("2. Add: GOOGLE_API_KEY=your_api_key or GEMINI_API_KEY=your_api_key")
        print("3. Get key from: https://aistudio.google.com/apikey")
        return 1

    # Initialize generator
    generator = NanoBananaImageGenerator(api_key=api_key)

    # Optimized prompt for command cheatsheet
    prompt = """
A hand-drawn terminal command reference card on textured cream paper.
The cheatsheet displays four MoAI-ADK commands in a clean, organized layout.

Top row (large header in brown): "MoAI Command Reference"

Four command rows displayed vertically:
1. "/moai:0-project" with rocket icon - Project initialization
2. "/moai:1-plan" with document icon - Specification planning
3. "/moai:2-run" with gear icon - TDD implementation
4. "/moai:3-sync" with sync arrows icon - Documentation sync

Visual style:
- Hand-drawn aesthetic with slight sketch texture
- Cream paper background (#F5F1E8) with subtle fiber texture
- Icons in soft blue (#6B9BD1) with hand-drawn outlines
- Command text and descriptions in warm brown (#8B6F47)
- Terminal-style monospace font for commands
- Friendly, approachable reference card design
- Clean spacing between elements
- Subtle drop shadow for depth

Technical specifications:
- 16:9 aspect ratio (landscape orientation)
- High-resolution output suitable for README documentation
- Professional quality with hand-crafted feel
- Clear readability at various sizes
- Warm, inviting color palette
- Clean composition with balanced whitespace

Photography style: Top-down flat lay shot with soft natural lighting.
Mood: Friendly, professional, educational reference material.
"""

    # Generate image
    print("🎨 Generating command cheatsheet card...")
    print(f"📝 Prompt length: {len(prompt)} characters")
    print(f"🎯 Resolution: 16:9 aspect ratio")
    print(f"⏱️  Estimated time: 25-35 seconds\n")

    try:
        image, metadata = generator.generate(
            prompt=prompt,
            model="pro",  # gemini-3-pro-image-preview
            aspect_ratio="16:9",
            save_path="assets/images/readme/command-cheatsheet-card.png"
        )

        print("✅ Image generation completed!")
        print(f"\n📸 Generation Settings:")
        print(f"   - Model: Nano Banana Pro (gemini-3-pro-image-preview)")
        print(f"   - Aspect Ratio: 16:9")
        print(f"   - Style: Hand-drawn terminal cheatsheet")

        print(f"\n✨ Technical Specifications:")
        print(f"   - SynthID Watermark: Included")
        print(f"   - Google Search: Enabled")
        print(f"   - Thinking Process: Enabled")
        print(f"   - Generation Time: {metadata.get('generation_time', 'N/A')}s")

        print(f"\n💾 Saved Location:")
        output_path = Path("assets/images/readme/command-cheatsheet-card.png").absolute()
        print(f"   {output_path}")

        print(f"\n🎯 Optimized Prompt Used:")
        print(f"   '{prompt[:100]}...'")

        return 0

    except Exception as e:
        print(f"❌ Error during generation: {e}")
        import traceback
        traceback.print_exc()
        return 1

if __name__ == "__main__":
    sys.exit(main())
