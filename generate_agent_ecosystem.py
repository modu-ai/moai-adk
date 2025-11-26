#!/usr/bin/env python3
"""
Generate MoAI-ADK Agent Orchestration Architecture Infographic
Using Nano Banana Pro (Gemini 3 Pro Image Preview)
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

    api_key = os.getenv("GEMINI_API_KEY")
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

    # Optimized prompt following Nano Banana Pro best practices
    prompt = """
A professional hand-drawn infographic diagram on cream-colored paper showing an AI agent orchestration architecture.

CENTRAL HUB - ALFRED LOGO:
In the center, draw the bold text "ALFRED" in large black letters with a distinctive bowtie symbol elegantly integrated into the letter 'F'. The bowtie should be a focal point, rendered in black with fine sketch lines. Below the logo, write in smaller handwritten text: "AI-powered Super Agent for Agentic Coding". Below that, a label reading "Orchestrator" in a hand-drawn box.

RADIAL AGENT LAYOUT (24 agents arranged in hub-and-spoke pattern):
Around the ALFRED logo, arrange 24 small rounded rectangles in five color-coded groups:

UPPER RIGHT QUADRANT (7 agents in green tones #7ED321):
- expert-backend
- expert-frontend
- expert-database
- expert-devops
- expert-security
- expert-uiux
- expert-debug
Group label above: "7 Domain Experts"

RIGHT SIDE (8 agents in orange tones #F5A623):
- manager-tdd
- manager-spec
- manager-docs
- manager-strategy
- manager-quality
- manager-git
- manager-project
- manager-claude-code
Group label: "8 Workflow Managers"

LOWER RIGHT QUADRANT (3 agents in purple tones #BD10E0):
- builder-agent
- builder-skill
- builder-command
Group label: "3 Meta-Builders"

LOWER LEFT QUADRANT (5 agents in teal tones #50E3C2):
- mcp-docs
- mcp-design
- mcp-notion
- mcp-browser
- mcp-ultrathink
Group label: "5 MCP Integrators"

UPPER LEFT QUADRANT (1 agent in pink tones #FF6B9D):
- ai-nano-banana
Group label: "1 AI Service"

USER INTERACTION (top-left corner):
Draw a simple person silhouette icon. Next to it, a hand-drawn speech bubble containing the text "Agentic Coding". A flowing wavy arrow connects from the user icon to the central ALFRED logo, showing the conversation flow.

CHAIN CONNECTIONS:
Draw organic, slightly wobbly brown-gray chain lines connecting from the ALFRED logo hub to each of the 24 agent boxes, creating a radial network pattern. The chains should have a hand-drawn quality with sketchy links.

STYLE SPECIFICATIONS:
Art style: Hand-drawn sketch notes aesthetic with colored pencils on cream paper.
Background color: Warm cream paper texture (#F5F1E8).
Line quality: Slightly wobbly, organic hand-drawn lines with visible pencil strokes.
Shading: Soft colored pencil shading for depth and dimension.
Typography: Handwritten style text throughout, clear and legible.
Color palette: Green (#7ED321), Orange (#F5A623), Purple (#BD10E0), Teal (#50E3C2), Pink (#FF6B9D), Brown-gray chains, Black ALFRED logo with bowtie.
Composition: Radial symmetry with ALFRED as the strong central focal point.
Mood: Professional yet approachable, educational, organized.

TECHNICAL SPECIFICATIONS:
Quality: Professional infographic quality, high-resolution detail.
Lighting: Soft, even lighting as if photographed from directly above.
Format: Clean digital scan aesthetic of hand-drawn artwork.
Aspect ratio: 16:9 landscape orientation.
Style reference: Sketchnotes, visual thinking, educational infographics.

The overall visual should communicate a sophisticated yet friendly AI orchestration system, with the ALFRED logo clearly positioned as the central coordinator managing all 24 specialized agents through a hub-and-spoke architecture pattern.
"""

    # Output path
    output_path = "/Users/goos/MoAI/MoAI-ADK/assets/images/readme/agent-skill-ecosystem.png"

    print("🎨 Generating MoAI-ADK Agent Orchestration Architecture Infographic...")
    print("📝 Using Nano Banana Pro (Gemini 3 Pro Image Preview)")
    print(f"📊 Resolution: 2K (optimized for web)")
    print(f"📐 Aspect Ratio: 16:9")
    print(f"💾 Output: {output_path}")
    print("\n⏳ Generation in progress (estimated 20-40 seconds)...\n")

    try:
        # Generate image with Nano Banana Pro
        result = generator.generate(
            prompt=prompt,
            model="pro",  # Use gemini-3-pro-image-preview
            aspect_ratio="16:9",
            save_path=output_path
        )

        image_data = result["image"]
        metadata = result["metadata"]

        print("✅ Image generation successful!")
        print(f"\n📸 Generation Details:")
        print(f"   - Model: {metadata.get('model', 'gemini-3-pro-image-preview')}")
        print(f"   - Processing Time: {metadata.get('processing_time_seconds', 'N/A')}s")
        print(f"   - Aspect Ratio: {metadata.get('aspect_ratio', '16:9')}")
        print(f"   - SynthID Watermark: Included (digital authentication)")
        print(f"   - Google Search Integration: {metadata.get('google_search_enabled', False)}")
        print(f"   - Thinking Process: {metadata.get('thinking_enabled', False)}")
        print(f"\n💾 Saved to: {output_path}")
        print(f"\n🎯 Optimized Prompt Length: {len(prompt)} characters")

        return True

    except Exception as e:
        print(f"❌ Error during image generation: {e}")
        print("\n🔧 Troubleshooting:")
        print("   1. Check API key in .env file")
        print("   2. Verify API quota (https://console.cloud.google.com/)")
        print("   3. Ensure network connectivity")
        print("   4. Try again in a few moments")
        return False

if __name__ == "__main__":
    success = generate_agent_ecosystem_infographic()
    sys.exit(0 if success else 1)
