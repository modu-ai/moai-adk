#!/usr/bin/env python3
"""
Generate Agent Ecosystem Infographic
Hand-drawn diagram showing MoAI-ADK's 24-agent orchestration with ALFRED logo
"""

import os
import sys
from pathlib import Path
from datetime import datetime

# Add skill module to path
skill_path = Path("/Users/goos/MoAI/MoAI-ADK/.claude/skills/moai-connector-nano-banana/modules")
sys.path.insert(0, str(skill_path))

from image_generator import NanoBananaImageGenerator

def generate_agent_ecosystem_infographic():
    """Generate hand-drawn agent orchestration architecture infographic."""

    # Load API key from .env
    from dotenv import load_dotenv
    load_dotenv()

    api_key = os.getenv("GEMINI_API_KEY") or os.getenv("GOOGLE_API_KEY")
    if not api_key:
        raise EnvironmentError(
            "❌ Google API Key not found!\n\n"
            "Setup instructions:\n"
            "1. Check .env file in project root\n"
            "2. Verify: GEMINI_API_KEY=your_api_key\n"
            "3. Get key from: https://aistudio.google.com/apikey"
        )

    # Initialize generator
    generator = NanoBananaImageGenerator(api_key=api_key)

    # Optimized prompt following Nano Banana Pro narrative style best practices
    prompt = """
A hand-drawn infographic on cream-colored paper showing an agent orchestration architecture diagram in a radial hub-and-spoke pattern. The illustration has a warm, friendly, hand-drawn notebook aesthetic with slightly wobbly organic lines and colored pencil shading throughout.

At the center of the composition, the ALFRED logo is prominently displayed as bold black text with a distinctive bowtie graphic cleverly integrated into the letter 'F'. The bowtie is rendered as an elegant black bow tie shape that forms part of the horizontal bar of the letter F. Below the logo are two descriptive labels written in neat handwriting: "AI-powered Super Agent for Agentic Coding" and underneath that "Orchestrator". The central ALFRED logo serves as the hub of the entire system.

Surrounding the central ALFRED logo in a circular radial arrangement are 24 sub-agents, each represented as rounded rectangular cards with hand-drawn borders. These agents are organized into five distinct groups, positioned around the circle like spokes on a wheel:

In the upper right section, seven green-colored cards in warm grass green tones represent the Domain Experts. Each card contains an agent name in clear handwriting: expert-backend, expert-frontend, expert-database, expert-devops, expert-security, expert-uiux, and expert-debug. Above this group is a label that reads "7 Domain Experts" with a subtle green underline.

In the right section, eight orange-colored cards in warm tangerine tones represent the Workflow Managers. The cards are labeled: manager-tdd, manager-spec, manager-docs, manager-strategy, manager-quality, manager-git, manager-project, and manager-claude-code. A label reading "8 Workflow Managers" with an orange underline sits above this cluster.

In the lower right section, three purple-colored cards in rich violet tones show the Meta-Builders: builder-agent, builder-skill, and builder-command. The label "3 Meta-Builders" with purple underline marks this group.

In the lower left section, five teal-colored cards in bright turquoise tones display the MCP Integrators: mcp-context7, mcp-figma, mcp-notion, mcp-playwright, and mcp-sequential-thinking. The label "5 MCP Integrators" with teal underline identifies this section.

In the upper left section, a single pink-colored card in warm rose tones shows the AI Service: ai-nano-banana. It is labeled "1 AI Service" with a pink underline.

From the central ALFRED logo, brown and gray hand-drawn chain lines radiate outward like spokes to connect to all 24 surrounding agent cards, creating a visual representation of orchestration and control. The chain lines are sketchy and organic with hand-drawn link shapes, emphasizing the hand-drawn aesthetic.

In the top-left corner of the image, there is a simple stick figure silhouette representing a user. Next to the user is a speech bubble drawn with wobbly lines containing the text "Agentic Coding". A curvy arrow flows from the user icon toward the central ALFRED logo, showing the direction of interaction.

The entire composition is arranged on cream-colored paper with a flat lay, top-down perspective. The lighting is soft and even, showing the texture of colored pencil shading. The style maintains a professional yet approachable, friendly visual language with clear hierarchy: the ALFRED logo is the central authority from which all agent connections radiate.

Photographic elements: Flat lay composition, top-down view, soft natural lighting showing paper texture and colored pencil details. The entire infographic is captured as if it were a high-quality scan of hand-drawn sketch notes. Camera: overhead shot with 50mm lens, even depth of field across the entire surface.

Color palette: Warm cream background, vibrant grass green for experts, warm tangerine orange for managers, rich violet purple for builders, bright turquoise teal for MCP, warm rose pink for AI service, brown-gray for chain connections, and bold black for the ALFRED logo and text.

Style: Hand-drawn sketch notes infographic, colored pencil on cream paper, organic and approachable, professional educational design.

Quality: High-resolution professional scan, 16:9 aspect ratio, studio-grade lighting, crisp text legibility while maintaining hand-drawn character. The ALFRED logo should be the dominant focal point with the bowtie clearly visible as part of the letter F design.
"""

    # Output path
    output_path = "/Users/goos/MoAI/MoAI-ADK/assets/images/readme/agent-skill-ecosystem.png"

    print("\n" + "="*70)
    print("🍌 Generating Agent Ecosystem Infographic with ALFRED Logo")
    print("="*70)
    print("\n📋 Configuration:")
    print(f"   • Model: Nano Banana Pro (gemini-3-pro-image-preview)")
    print(f"   • Style: Hand-drawn sketch notes on cream paper")
    print(f"   • Aspect Ratio: 16:9 (1920x1080)")
    print(f"   • Output: {output_path}")
    print("\n⏳ Generating image (this may take 20-60 seconds)...\n")

    try:
        # Generate image with Nano Banana Pro
        image, metadata = generator.generate(
            prompt=prompt,
            model="pro",
            aspect_ratio="16:9",
            save_path=output_path
        )

        print("\n" + "="*70)
        print("✅ Agent Ecosystem Infographic Generated Successfully!")
        print("="*70)
        print(f"\n📊 Generation Details:")
        print(f"   • Model: {metadata['model_name']}")
        print(f"   • Aspect Ratio: {metadata['aspect_ratio']}")
        print(f"   • Tokens Used: {metadata['tokens_used']}")
        print(f"   • Timestamp: {metadata['timestamp']}")
        print(f"\n💾 Saved to: {output_path}")
        print("\n🎨 Visual Elements:")
        print("   ✓ Central ALFRED logo with bowtie")
        print("   ✓ 24 agents in hub-and-spoke pattern")
        print("   ✓ 5-tier color coding (green, orange, purple, teal, pink)")
        print("   ✓ User interaction (top-left corner)")
        print("   ✓ Chain connections from ALFRED to all agents")
        print("   ✓ Hand-drawn sketch aesthetic on cream paper")
        print("\n✨ Ready for README integration!")

        return True

    except Exception as e:
        print(f"\n❌ Error during image generation: {e}")
        print("\n🔧 Troubleshooting:")
        print("   1. Check API key in .env file")
        print("   2. Verify API quota (https://console.cloud.google.com/)")
        print("   3. Ensure network connectivity")
        print("   4. Try again in a few moments")
        return False

if __name__ == "__main__":
    success = generate_agent_ecosystem_infographic()
    sys.exit(0 if success else 1)
