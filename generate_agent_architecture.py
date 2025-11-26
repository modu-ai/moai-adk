#!/usr/bin/env python3
"""
Generate MoAI-ADK Agent Architecture Infographic
Hand-drawn style with ALFRED logo and 24 sub-agents
"""

import sys
import os
from pathlib import Path

# Add skill modules to path
skill_path = Path(__file__).parent / ".claude/skills/moai-connector-nano-banana/modules"
sys.path.insert(0, str(skill_path))

from image_generator import NanoBananaImageGenerator
from dotenv import load_dotenv

# Load environment variables
load_dotenv()

# Optimized prompt for Nano Banana Pro
PROMPT = """A professional hand-drawn infographic on cream-colored textured paper showing MoAI-ADK's agent orchestration architecture in a radial hub-and-spoke layout.

Central Hub: The ALFRED logo displayed prominently in the center - bold black uppercase letters "ALFRED" with a distinctive black bowtie symbol elegantly integrated into the letter 'F'. The bowtie should be clearly visible as part of the typography design. Below the logo: "AI-powered Super Agent for Agentic Coding" written in smaller hand-lettered text. Below that: "Orchestrator" label inside a hand-drawn rounded badge.

Radial Hub-and-Spoke Layout with 24 Agent Boxes:

Upper Right Sector (Green Zone): Hand-drawn header "7 Domain Experts" with 7 rounded rectangular boxes below arranged in a gentle arc, each labeled: expert-backend, expert-frontend, expert-database, expert-devops, expert-security, expert-uiux, expert-debug. All boxes filled with green colored pencil shading.

Right Sector (Orange Zone): Hand-drawn header "8 Workflow Managers" with 8 rounded rectangular boxes arranged in a vertical column, each labeled: manager-tdd, manager-spec, manager-docs, manager-strategy, manager-quality, manager-git, manager-project, manager-claude-code. All boxes filled with orange colored pencil shading.

Lower Right Sector (Purple Zone): Hand-drawn header "3 Meta-Builders" with 3 rounded rectangular boxes arranged in a small cluster, each labeled: builder-agent, builder-skill, builder-command. All boxes filled with purple colored pencil shading.

Lower Left Sector (Teal Zone): Hand-drawn header "5 MCP Integrators" with 5 rounded rectangular boxes arranged in a gentle arc, each labeled: mcp-context7, mcp-figma, mcp-notion, mcp-playwright, mcp-sequential-thinking. All boxes filled with teal colored pencil shading.

Upper Left Sector (Pink Zone): Hand-drawn header "1 AI Service" with 1 rounded rectangular box labeled: ai-nano-banana. Box filled with pink colored pencil shading.

User Interaction (Top-Left Corner): Simple hand-drawn person silhouette icon with "User" label underneath. Next to it, a hand-drawn speech bubble containing the text "Agentic Coding". A wavy brown arrow drawn from the user icon to the ALFRED logo center, showing conversation flow.

Connection Network: Multiple wavy chain-style lines drawn in brown and gray sketch pencil connecting from the ALFRED logo at center radiating outward to each of the 24 agent boxes like a spider web, emphasizing the orchestrator pattern. The chains should look hand-drawn with slight wobble.

Visual Style: Hand-drawn sketch notes aesthetic with colored pencil technique. All lines are slightly wobbly and organic. Visible colored pencil strokes for shading within boxes. Hand-lettered typography throughout. Friendly and approachable visual language. The ALFRED logo is significantly larger than other text, making it the clear focal point.

Lighting: Flat illustration style with subtle colored pencil shading creating gentle depth within boxes.

Color Palette: Cream textured paper background, ALFRED logo in bold black with black bowtie, Expert agents in green tones, Manager agents in orange tones, Builder agents in purple tones, MCP agents in teal tones, AI agent in pink tones, connection chains in brown and gray sketch.

Composition: Radial symmetrical layout with clear visual hierarchy - ALFRED logo prominently sized at center, 24 agent boxes distributed in 5 colored sectors radiating outward, user interaction in top-left corner.

Quality: High-resolution professional infographic suitable for documentation and presentations, 16:9 widescreen aspect ratio, all 24 agent names clearly legible, tier labels visible, ALFRED logo prominent with bowtie detail integrated into the F.

Mood: Professional yet approachable, educational and clear, warm and friendly color palette, inviting visual language."""


def main():
    """Generate the agent architecture infographic"""

    # Check for API key
    api_key = os.getenv("GEMINI_API_KEY") or os.getenv("GOOGLE_API_KEY")
    if not api_key:
        print("❌ Error: Google API Key not found!")
        print("\nPlease set your API key:")
        print("  export GOOGLE_API_KEY='your-api-key'")
        print("  or")
        print("  export GEMINI_API_KEY='your-api-key'")
        print("\nGet your API key from: https://aistudio.google.com/apikey")
        sys.exit(1)

    # Output path
    output_path = Path(__file__).parent / "assets/images/readme/agent-skill-ecosystem.png"

    print("\n" + "="*70)
    print("🎨 Generating MoAI-ADK Agent Architecture Infographic")
    print("="*70)
    print(f"\n📝 Prompt: Hand-drawn ALFRED orchestration with 24 agents")
    print(f"🎯 Style: Sketch notes on cream paper with colored pencils")
    print(f"📐 Aspect Ratio: 16:9 (1920x1080)")
    print(f"🖼️  Output: {output_path}")
    print(f"\n⏳ Generating with Nano Banana Pro (gemini-3-pro-image-preview)...")
    print(f"   This may take 20-60 seconds for high quality...\n")

    try:
        # Initialize generator
        generator = NanoBananaImageGenerator(api_key=api_key)

        # Generate image
        image, metadata = generator.generate(
            prompt=PROMPT,
            model="pro",  # gemini-3-pro-image-preview (Nano Banana Pro)
            aspect_ratio="16:9",
            save_path=str(output_path)
        )

        print(f"\n" + "="*70)
        print(f"✅ Image generation successful!")
        print(f"="*70)
        print(f"\n📊 Generation Details:")
        print(f"   • Model: {metadata['model_name']}")
        print(f"   • Aspect Ratio: {metadata['aspect_ratio']}")
        print(f"   • Tokens Used: {metadata['tokens_used']}")
        print(f"   • Timestamp: {metadata['timestamp']}")
        print(f"\n💾 Saved to: {output_path}")
        print(f"\n🎉 Done! Your hand-drawn agent architecture infographic is ready!")

    except Exception as e:
        print(f"\n❌ Error during generation: {e}")
        print(f"\nTroubleshooting:")
        print(f"  1. Check your API key is valid")
        print(f"  2. Ensure you have quota remaining")
        print(f"  3. Verify network connectivity")
        sys.exit(1)


if __name__ == "__main__":
    main()
