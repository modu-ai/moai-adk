"""
Generate MoAI-ADK Agent Orchestration Diagram
Hand-drawn style showing Alfred as central orchestrator with 24 sub-agents
"""

import sys
import os
from pathlib import Path

# Add skill modules to path
skill_path = Path("/Users/goos/MoAI/MoAI-ADK/.claude/skills/moai-connector-nano-banana/modules")
sys.path.insert(0, str(skill_path))

from image_generator import NanoBananaImageGenerator

def main():
    """Generate agent orchestration diagram"""

    # Load API key from environment
    api_key = os.getenv("GOOGLE_API_KEY") or os.getenv("GEMINI_API_KEY")
    if not api_key:
        print("❌ Error: GOOGLE_API_KEY or GEMINI_API_KEY environment variable not set")
        print("Please set your API key: export GOOGLE_API_KEY='your-key-here'")
        sys.exit(1)

    # Initialize generator
    generator = NanoBananaImageGenerator(api_key=api_key)

    # Optimized structured prompt for Nano Banana Pro
    prompt = """A hand-drawn organizational diagram showing MoAI-ADK's agentic coding ecosystem with hub-and-spoke architecture.

Central Hub - Mr. Alfred:
In the center of the composition, a large friendly stick figure character wearing a top hat, holding a conductor's baton in a maestro pose. Label clearly reads "🎩 Mr. Alfred" with subtitle "Super Agent Orchestrator" below. The character is highlighted with a blue glow or accent color (#6B9BD1). Alfred acts as the central control point with wavy chain lines radiating outward to all surrounding agents.

Five Groups of Sub-Agents (24 total) arranged radially around Alfred:

Top Section - Expert Agents (7 agents, green theme #4CAF50):
Small stick figures with specialty tool icons, arranged in an arc at the top.
Labels: "expert-backend", "expert-frontend", "expert-database", "expert-devops", "expert-security", "expert-uiux", "expert-debug"
Each connected to Alfred by wavy chain lines. Green accent highlights.

Right Section - Manager Agents (8 agents, orange theme #FF9800):
Small stick figures holding briefcases, arranged vertically on the right side.
Labels: "manager-project", "manager-spec", "manager-tdd", "manager-docs", "manager-strategy", "manager-quality", "manager-git", "manager-claude-code"
Connected to Alfred by wavy chain lines. Orange accent highlights.

Bottom Section - Builder Agents (3 agents, purple theme #9C27B0):
Small stick figures with construction helmets and tools, arranged at the bottom.
Labels: "builder-agent", "builder-skill", "builder-command"
Connected to Alfred by wavy chain lines. Purple accent highlights.

Left Section - MCP Integrators (5 agents, blue theme #2196F3):
Small stick figures with electrical plug symbols, arranged vertically on the left side.
Labels: "mcp-docs", "mcp-design", "mcp-notion", "mcp-browser", "mcp-ultrathink"
Connected to Alfred by wavy chain lines. Blue accent highlights.

Bottom-Left Corner - AI Services (1 agent, yellow theme #FFC107):
Small robot-style stick figure with antenna.
Label: "ai-nano-banana"
Connected to Alfred by wavy chain line. Yellow accent highlight.

User Interaction Section (Top-Left):
A simple stick figure labeled "👤 User" in the top-left corner.
Speech bubble above user with text "Agentic Coding".
Bidirectional wavy arrow connecting User to Alfred, labeled "Natural language conversation".

Visual Elements:
- Background: Cream-colored paper texture (#F5F1E8) giving a warm, organic feel
- All lines and text: Hand-drawn brown ink (#8B6F47) with organic, slightly imperfect lines
- Chain connections: Wavy, hand-drawn lines showing orchestration flow
- Dotted lines: Lighter dotted connections between some sub-agents
- Arrows: Hand-drawn directional arrows showing workflow
- Decorative elements: Small stars, asterisks, and organic embellishments scattered tastefully
- Annotations: "Orchestration" label near Alfred, "Control Flow" along some arrows, "24 Sub-Agents" at bottom

Text Labels:
All agent names should be clearly legible in hand-written style font.
Group labels should indicate the category (Experts, Managers, Builders, MCP, AI).

Composition and Style:
Hub-and-spoke pattern with perfect radial symmetry.
Hand-drawn sketch notes aesthetic, similar to whiteboard diagrams or visual facilitation graphics.
Friendly, approachable stick figures with personality.
Professional yet warm and inviting visual style.
Clear information hierarchy with Alfred as the obvious focal point.
Organic, slightly imperfect hand-drawn quality that feels authentic and human.
Educational and documentation-friendly design.

Mood and Atmosphere:
Professional but approachable, educational, clear, friendly, well-organized, inspiring confidence.

Technical Specifications:
High-resolution output suitable for professional documentation.
Aspect ratio: 16:9 (wide format perfect for README and presentations).
Output: PNG format with transparent or cream background.
All text legible at various sizes (documentation, presentations, web).
Color coding clear and consistent for each agent group.
Visual clarity prioritized for information architecture understanding.

Quality: Studio-grade illustration quality, professional documentation standard, high-resolution rendering."""

    # Output path
    output_path = "/Users/goos/MoAI/MoAI-ADK/assets/images/readme/agent-skill-ecosystem.png"

    print("\n" + "="*80)
    print("🎨 Generating MoAI-ADK Agent Orchestration Diagram")
    print("="*80)
    print(f"📍 Output: {output_path}")
    print(f"🎯 Model: Nano Banana Pro (gemini-3-pro-image-preview)")
    print(f"📐 Aspect Ratio: 16:9")
    print(f"⏳ Estimated time: 30-45 seconds")
    print("="*80 + "\n")

    # Generate image
    try:
        image, metadata = generator.generate(
            prompt=prompt,
            model="pro",  # Use Nano Banana Pro for highest quality
            aspect_ratio="16:9",
            save_path=output_path
        )

        print("\n" + "="*80)
        print("✅ SUCCESS: Agent Orchestration Diagram Generated")
        print("="*80)
        print(f"📁 Saved to: {output_path}")
        print(f"📊 Image size: {image.size[0]} x {image.size[1]} pixels")
        print(f"🎯 Tokens used: {metadata['tokens_used']}")
        print(f"⏱️  Timestamp: {metadata['timestamp']}")
        print("="*80 + "\n")

        # Display prompt summary
        print("📝 Prompt Summary:")
        print("   • Central orchestrator: Mr. Alfred (top hat, conductor)")
        print("   • 7 Expert agents (top, green)")
        print("   • 8 Manager agents (right, orange)")
        print("   • 3 Builder agents (bottom, purple)")
        print("   • 5 MCP Integrator agents (left, blue)")
        print("   • 1 AI Service agent (bottom-left, yellow)")
        print("   • User interaction (top-left)")
        print("   • Hub-and-spoke architecture with wavy chain connections")
        print("   • Hand-drawn sketch note style on cream paper")
        print("\n✨ Image ready for documentation use!\n")

    except Exception as e:
        print(f"\n❌ ERROR: {e}\n")
        sys.exit(1)

if __name__ == "__main__":
    main()
