---
name: moai-connector-nano-banana
description: Professional image generation with Google Nano Banana Pro (Gemini 3 Pro)
version: 1.0.1
modularized: true
tags:
  - enterprise
  - patterns
  - nano
  - banana
  - architecture
updated: 2025-11-24
status: active
---

## 🎯 Quick Reference (30 seconds)

**Purpose**: Professional image generation using Google's Nano Banana Pro (Gemini 3 Pro Image Preview).

**Key Features**:
- **Text-to-Image**: Detailed prompts → 1K/2K/4K resolution images
- **Image-to-Image**: Style transfer, object manipulation, editing
- **Real-time Grounding**: Google Search integration for factual content
- **Multi-Reference**: Up to 14 reference images (6 objects + 5 humans)
- **Advanced Text**: Sophisticated text rendering directly in images

**Two Models**:
1. **Nano Banana Pro** (gemini-3-pro-image-preview) - Professional quality, 10-60s
2. **Gemini 2.5 Flash** (gemini-2.5-flash-image) - Fast, ~5-15s

---

## 🍌 Agent Introduction (에이전트 소개)

### What is Nano Banana? (나노 바나나란?)

**Nano Banana** is MoAI-ADK's professional image generation connector, providing a unified interface to Google's Gemini Image Generation API. It acts as a bridge between natural language requests and high-quality AI-generated images.

**Core Capabilities**:
- **Text-to-Image Generation**: Transform detailed prompts into professional images (1K/2K/4K)
- **Image-to-Image Editing**: Apply style transfer, modify objects, and refine compositions
- **Multi-Reference Guidance**: Use up to 14 reference images for style consistency
- **Multi-turn Refinement**: Iteratively improve images through conversational feedback
- **Batch Processing**: Generate multiple images efficiently for campaigns or variations
- **Google Search Grounding**: Integrate real-time data for factual accuracy (Pro model)
- **Prompt Optimization**: Built-in prompt engineering assistance for best results
- **Model Selection Guide**: Intelligent recommendations for Pro vs Flash models

**Supported Use Cases**:
- **Marketing Materials**: Product photography, advertising visuals, promotional content
- **UI/UX Assets**: Mockups, design concepts, component illustrations
- **Documentation**: Technical diagrams, tutorials, visual guides
- **Content Creation**: Social media graphics, blog illustrations, presentations
- **Prototyping**: Rapid visual ideation, concept validation, design exploration

---

### Command Usage (커맨드 사용법)

Nano Banana integrates seamlessly with MoAI-ADK's command workflow for professional image generation projects.

#### `/moai:0-project` - Project Initialization

**Setup Nano Banana in New Projects**:

```bash
# Initialize MoAI-ADK project with image generation
/moai:0-project

# System automatically:
# 1. Creates .env file for API key storage
# 2. Configures moai-connector-nano-banana skill
# 3. Sets up outputs/ directory for generated images
# 4. Adds .gitignore rules for .env and outputs/
```

**Manual API Key Setup**:

```bash
# Create .env file
touch .env

# Add your Google Gemini API key
echo "GOOGLE_API_KEY=your_api_key_here" >> .env

# Secure permissions (owner read/write only)
chmod 600 .env

# Get API key from: https://aistudio.google.com/apikey
```

**Verification**:

```bash
# Verify setup
ls -la .env  # Should show -rw------- (600 permissions)
cat .env     # Should show GOOGLE_API_KEY=...
```

---

#### `/moai:1-plan` - Image Generation SPEC Creation

**Generate SPEC for Image Generation Projects**:

```bash
# Example 1: Marketing campaign imagery
/moai:1-plan "Create a SPEC for generating 10 product photos for a new smartphone launch campaign. Images should be professional studio quality, 4K resolution, various angles and lighting setups."

# System creates SPEC-001 with:
# - Image generation requirements (resolution, style, quantity)
# - Model selection recommendations (Pro for 4K quality)
# - Prompt engineering guidelines
# - Reference image requirements (if any)
# - Output organization structure
# - Quality validation criteria
```

**Example SPEC Structure**:

```markdown
# SPEC-001: Product Photography Campaign

## Requirements
- Generate 10 professional product photos
- Resolution: 4K (Pro model required)
- Aspect ratios: 1:1 (Instagram), 16:9 (web)
- Lighting: Studio setup with gradient backgrounds
- Reference images: 3 existing product shots

## Prompt Engineering Strategy
- Use structured prompts (Scene + Lighting + Camera + Style)
- Apply photographic elements (85mm lens, f/2.8, shallow DOF)
- Specify color palette (metallic, gradient backgrounds)
- Include quality indicators (studio-grade, 4K resolution)

## Model Selection
- Primary: Gemini 3 Pro (quality priority, 4K support)
- Fallback: Gemini 2.5 Flash (speed priority for variations)

## Output Organization
outputs/
├── product-front-1.png (4K, 1:1)
├── product-angle-2.png (4K, 16:9)
└── ...

## Validation Criteria
- Resolution check: Exactly 4K (3840x2160 or equivalent)
- Quality review: Professional studio-grade appearance
- Consistency check: Color palette alignment across all images
```

---

#### `/moai:2-run` - Image Generation Execution

**Execute Image Generation with TDD Workflow**:

```bash
# Run SPEC-001 for image generation
/moai:2-run SPEC-001

# System workflow:
# RED Phase: Define test criteria
#   - Resolution validation tests
#   - Image quality checks
#   - Color palette consistency
#   - Prompt optimization validation
#
# GREEN Phase: Generate images
#   - Load nano-banana skill
#   - Apply prompt engineering patterns
#   - Generate with selected model (Pro/Flash)
#   - Save to outputs/ with metadata
#
# REFACTOR Phase: Optimize results
#   - Review generation quality
#   - Refine prompts if needed
#   - Regenerate low-quality images
#   - Validate against test criteria
```

**Agent Delegation Pattern**:

```python
# Alfred delegates to ai-nano-banana agent
Task(
    subagent_type="ai-nano-banana",
    prompt="""
    Generate 10 professional product photos according to SPEC-001.

    Requirements:
    - Model: Gemini 3 Pro (4K quality)
    - Resolution: 4K for print quality
    - Style: Professional studio photography
    - Lighting: Gradient backgrounds, studio setup
    - Reference images: Use provided product shots for consistency

    Apply structured prompts:
    1. Scene description (product, angle, setting)
    2. Photographic elements (85mm lens, f/2.8, lighting)
    3. Color palette (metallic, gradients)
    4. Quality standards (studio-grade, 4K)

    Save all images to outputs/ with descriptive names and metadata.
    """
)
```

---

#### `/moai:3-sync` - Documentation & Asset Management

**Document Generated Images and Process**:

```bash
# Sync documentation for SPEC-001
/moai:3-sync SPEC-001

# System automatically:
# 1. Catalogs all generated images
# 2. Creates README.md with image previews
# 3. Documents prompts used for each image
# 4. Saves metadata (model, resolution, parameters)
# 5. Generates markdown gallery
# 6. Creates GitHub-ready asset documentation
```

**Generated Documentation Structure**:

```markdown
# Product Photography Campaign - SPEC-001

## Generated Images (10 total)

### Image 1: Front View
![Product Front](outputs/product-front-1.png)

**Generation Details**:
- Model: Gemini 3 Pro Image
- Resolution: 4K (3840x2160)
- Aspect Ratio: 1:1
- Prompt: "A professional studio photograph of a premium smartphone..."
- Generation Time: 28 seconds
- Tokens Used: 3,240

### Metadata
- Total generations: 10 images
- Total processing time: 4m 32s
- Average quality score: 4.8/5.0
- Model: Gemini 3 Pro (100%)
- Total cost estimate: $0.80
```

---

### Agent Usage (에이전트 사용법)

Nano Banana can be used through specialized agents for different workflows.

#### **ai-nano-banana** - Dedicated Image Generation Agent

**Primary Use Case**: Direct image generation with prompt optimization

**Capabilities**:
- Analyzes natural language requests ("cute cat eating banana")
- Transforms vague requests into optimized prompts
- Selects appropriate model (Pro vs Flash)
- Handles multi-turn refinement
- Manages API key configuration
- Saves images with metadata

**Invocation Pattern**:

```python
# Example 1: Simple text-to-image
Task(
    subagent_type="ai-nano-banana",
    prompt="""
    Generate a professional product photo of wireless headphones.

    Style: Studio photography
    Lighting: Soft diffused light with subtle gradient background
    Quality: 4K resolution for print
    Mood: Premium, elegant, minimalist
    """
)

# Example 2: Image editing
Task(
    subagent_type="ai-nano-banana",
    prompt="""
    Edit the image at outputs/product-original.png:
    - Change background to sunset outdoor scene
    - Keep product lighting consistent
    - Maintain photorealistic style
    - Resolution: 2K
    """
)

# Example 3: Batch generation with variations
Task(
    subagent_type="ai-nano-banana",
    prompt="""
    Generate 5 variations of a mountain landscape:
    1. Sunrise lighting (warm golden tones)
    2. Midday (bright, clear blue sky)
    3. Sunset (dramatic purple and orange)
    4. Twilight (deep blues, stars emerging)
    5. Night (moonlit, starry sky)

    Consistency: Same composition, camera angle, mountain features
    Model: Flash (speed priority for variations)
    Resolution: 2K
    Aspect ratio: 16:9
    """
)
```

**Agent Workflow**:

1. **Request Analysis**: Parse natural language, extract key elements
2. **Clarification**: Use AskUserQuestion if details missing
3. **Prompt Engineering**: Transform into structured Nano Banana prompt
4. **Model Selection**: Recommend Pro vs Flash based on requirements
5. **Generation**: Call Gemini API with optimized parameters
6. **Feedback Loop**: Present result, collect feedback, refine if needed

---

#### **expert-frontend** - UI/UX Asset Generation

**Primary Use Case**: Generate design assets for frontend implementations

**Capabilities**:
- UI mockup generation
- Component illustrations
- Icon and graphic creation
- Design system assets
- Interactive example visuals

**Invocation Pattern**:

```python
# Example 1: UI component illustrations
Task(
    subagent_type="expert-frontend",
    prompt="""
    Using moai-connector-nano-banana, generate UI component illustrations:

    Components needed:
    1. Login form mockup (modern, gradient background)
    2. Dashboard card designs (3 variations)
    3. Navigation menu concept (mobile-first)
    4. Button states (default, hover, active, disabled)

    Style: Modern, flat design, Material Design inspired
    Colors: Primary blue (#3B82F6), neutral grays
    Resolution: 2K, aspect ratio: 16:9
    Model: Flash (speed priority for rapid prototyping)

    Save to: outputs/ui-components/
    """
)

# Example 2: Hero section background images
Task(
    subagent_type="expert-frontend",
    prompt="""
    Generate 3 abstract background images for hero sections:

    Style: Geometric, modern, gradient overlays
    Colors: Blue-purple gradients
    Resolution: 4K (for retina displays)
    Aspect ratio: 21:9 (ultra-wide)
    Model: Pro (quality priority for hero visuals)

    Use cases:
    - Landing page hero
    - About page header
    - Contact section background
    """
)
```

**Integration Workflow**:

1. Frontend expert analyzes UI requirements
2. Delegates image generation to ai-nano-banana
3. Receives generated assets
4. Integrates into frontend codebase
5. Validates image optimization (WebP conversion, lazy loading)
6. Documents asset usage in component library

---

#### **expert-devops** - Infrastructure Visualization

**Primary Use Case**: Generate diagrams and visuals for deployment documentation

**Capabilities**:
- Architecture diagram illustrations
- System topology visuals
- Infrastructure process flows
- Deployment pipeline graphics
- Monitoring dashboard mockups

**Invocation Pattern**:

```python
# Example 1: Cloud architecture illustration
Task(
    subagent_type="expert-devops",
    prompt="""
    Using moai-connector-nano-banana, generate a visual illustration
    of our cloud architecture:

    Components:
    - Load balancer (AWS ALB)
    - EC2 instances (3 nodes)
    - RDS database (PostgreSQL)
    - S3 buckets (static assets)
    - CloudFront CDN
    - Route 53 DNS

    Style: Professional technical diagram aesthetic
    Layout: Left-to-right flow (user → CDN → ALB → EC2 → RDS)
    Resolution: 2K
    Aspect ratio: 16:9
    Model: Pro (clarity and detail priority)

    Include: Connection arrows, labels, AWS service icons (approximated)
    """
)

# Example 2: Deployment pipeline visualization
Task(
    subagent_type="expert-devops",
    prompt="""
    Generate CI/CD pipeline flow illustration:

    Stages:
    1. Developer commits → GitHub
    2. GitHub Actions triggered
    3. Build & Test (pytest, coverage)
    4. Docker image build
    5. Push to ECR
    6. Deploy to ECS (staging)
    7. Integration tests
    8. Deploy to ECS (production)

    Style: Flowchart-like, modern DevOps aesthetic
    Colors: GitHub dark theme colors
    Resolution: 2K
    Model: Flash (speed priority)
    """
)
```

---

#### **manager-docs** - Documentation Visual Assets

**Primary Use Case**: Generate illustrations for technical documentation

**Capabilities**:
- Tutorial step illustrations
- Concept diagrams
- Feature showcase visuals
- README banner images
- API documentation graphics

**Invocation Pattern**:

```python
# Example 1: Tutorial step illustrations
Task(
    subagent_type="manager-docs",
    prompt="""
    Using moai-connector-nano-banana, generate 5 tutorial step illustrations
    for "Getting Started with MoAI-ADK" guide:

    Step 1: Installation (command terminal visual)
    Step 2: Project initialization (folder structure)
    Step 3: First SPEC creation (editor with SPEC file)
    Step 4: TDD execution (test results output)
    Step 5: Documentation sync (GitHub repo view)

    Style: Modern, clean, screenshot-like but illustrated
    Colors: VSCode dark theme palette
    Resolution: 2K
    Aspect ratio: 16:9
    Model: Flash (batch generation efficiency)

    Save to: .moai/docs/getting-started/images/
    """
)

# Example 2: README banner image
Task(
    subagent_type="manager-docs",
    prompt="""
    Generate a professional README banner for GitHub repository:

    Content:
    - Project name: "MoAI-ADK"
    - Tagline: "AI-Powered Development Kit"
    - Visual elements: Abstract tech background, modern gradient

    Style: Professional, tech-forward, clean typography
    Colors: Blue gradients (#3B82F6 → #8B5CF6)
    Resolution: 4K (high-quality GitHub display)
    Aspect ratio: 21:9 (GitHub banner standard)
    Model: Pro (branding quality priority)

    Include: Logo placeholder, text overlays, gradient effects
    """
)

# Example 3: API documentation diagrams
Task(
    subagent_type="manager-docs",
    prompt="""
    Generate sequence diagram illustrations for API documentation:

    Scenario: User authentication flow

    Actors:
    - Client (web app)
    - API Gateway
    - Auth Service
    - User Database

    Flow:
    1. Client → API: POST /login (credentials)
    2. API → Auth: Validate credentials
    3. Auth → DB: Query user
    4. DB → Auth: User record
    5. Auth → API: JWT token
    6. API → Client: Token + user data

    Style: Technical sequence diagram, UML-inspired
    Resolution: 2K
    Model: Flash (documentation efficiency)
    """
)
```

---

### Skill Builder Agent Usage (스킬 빌더 에이전트 사용법)

The **builder-skill** agent enables extension and customization of the moai-connector-nano-banana skill.

#### What is Skill Builder? (스킬 빌더란?)

**Skill Builder** is a specialized meta-agent that creates, updates, and extends MoAI-ADK skills. It understands skill architecture, documentation patterns, and integration requirements.

**Capabilities**:
- Create new skills from scratch
- Extend existing skills with new features
- Update skill documentation
- Add examples and patterns
- Ensure MoAI-ADK standards compliance
- Generate skill tests and validation

---

#### Use Case 1: Add New Model Support

**Scenario**: Google releases a new image generation model, and you want to add support to nano-banana.

**Invocation**:

```python
Task(
    subagent_type="builder-skill",
    prompt="""
    Extend moai-connector-nano-banana skill to support the new
    Gemini 4 Ultra Image model (gemini-4-ultra-image).

    Requirements:
    1. Add new model to NanoBananaImageGenerator.MODELS dict
    2. Support 8K resolution (new capability)
    3. Add model-specific parameters (ultra_quality_mode)
    4. Update model selection guide in SKILL.md
    5. Add examples in examples.md
    6. Create tests for new model
    7. Update API reference documentation

    Maintain backward compatibility:
    - Keep existing Pro and Flash models
    - Preserve all existing parameters
    - Ensure seamless upgrade path

    Documentation updates needed:
    - SKILL.md: Add "Ultra" to model comparison table
    - modules/api-reference.md: Document 8K resolution
    - examples.md: Add ultra-quality example
    - modules/troubleshooting.md: Add common Ultra issues
    """
)
```

**Builder-Skill Workflow**:

1. **Analysis Phase**:
   - Read current skill structure
   - Identify integration points
   - Check compatibility requirements
   - Review documentation organization

2. **Design Phase**:
   - Design new model integration architecture
   - Plan parameter extensions
   - Design documentation updates
   - Create test strategy

3. **Implementation Phase**:
   - Update `modules/image_generator.py`:
     ```python
     MODELS = {
         "flash": "gemini-2.5-flash-image",
         "pro": "gemini-3-pro-image-preview",
         "ultra": "gemini-4-ultra-image"  # NEW
     }

     RESOLUTIONS = ["1K", "2K", "4K", "8K"]  # Add 8K
     ```
   - Add ultra_quality_mode parameter
   - Update generate() method logic

4. **Documentation Phase**:
   - Update SKILL.md model comparison
   - Add 8K resolution examples
   - Document ultra-specific features
   - Create migration guide

5. **Validation Phase**:
   - Test all models (flash, pro, ultra)
   - Verify backward compatibility
   - Check documentation accuracy
   - Validate example code

**Expected Outcome**:

```markdown
## Updated SKILL.md

**Three Models**:
1. **Nano Banana Pro** (gemini-3-pro-image-preview) - Professional quality, 4K max
2. **Gemini 2.5 Flash** (gemini-2.5-flash-image) - Fast, 1K max
3. **Gemini 4 Ultra** (gemini-4-ultra-image) - Ultra quality, 8K max ⭐ NEW

## Model Selection Guide

| Feature | Flash | Pro | Ultra |
|---------|-------|-----|-------|
| Max Resolution | 1K | 4K | 8K |
| Speed | Fast (5-15s) | Medium (10-60s) | Slow (30-90s) |
| Quality | Good | Professional | Ultra Premium |
| Cost | $ | $$ | $$$ |
```

---

#### Use Case 2: Add Video Generation Support

**Scenario**: Extend nano-banana to support video generation (future capability).

**Invocation**:

```python
Task(
    subagent_type="builder-skill",
    prompt="""
    Extend moai-connector-nano-banana with video generation capabilities.

    New Features:
    1. Text-to-Video: Generate short videos from prompts (5-15s)
    2. Image-to-Video: Animate static images
    3. Video-to-Video: Apply style transfer to videos
    4. Multi-clip composition: Combine multiple video segments

    Technical Requirements:
    - Add VideoGenerator class in modules/video_generator.py
    - Support MP4, WebM output formats
    - Frame rates: 24fps, 30fps, 60fps
    - Resolutions: 720p, 1080p, 4K
    - Duration: 5s to 60s clips

    API Integration:
    - Use Gemini Video Generation API (when available)
    - Maintain consistency with image generation patterns
    - Reuse prompt engineering utilities

    Documentation:
    - Add "Video Generation" section to SKILL.md
    - Create modules/video-generation.md with detailed guide
    - Add video examples to examples.md
    - Update ai-nano-banana agent to support video workflows

    Maintain Compatibility:
    - Keep all existing image generation features
    - Shared utilities (prompt engineering, model selection)
    - Unified error handling
    """
)
```

**Implementation Structure**:

```python
# New file: modules/video_generator.py

class NanoBananaVideoGenerator:
    """Video generation using Gemini Video API"""

    FORMATS = ["mp4", "webm"]
    FRAME_RATES = [24, 30, 60]
    RESOLUTIONS = ["720p", "1080p", "4K"]

    def generate_video(
        self,
        prompt: str,
        duration_seconds: int = 10,
        resolution: str = "1080p",
        fps: int = 30,
        format: str = "mp4"
    ) -> Tuple[bytes, Dict]:
        """Generate video from text prompt"""
        pass

    def animate_image(
        self,
        image_path: str,
        animation_prompt: str,
        duration_seconds: int = 5
    ) -> Tuple[bytes, Dict]:
        """Animate static image"""
        pass
```

**Documentation Updates**:

```markdown
## SKILL.md Update

### Video Generation (New Capability)

**Features**:
- Text-to-Video: Generate 5-60s videos from prompts
- Image-to-Video: Animate static images with motion
- Style Transfer: Apply artistic styles to existing videos
- Multi-clip Composition: Combine segments into sequences

**Usage**:
```python
from modules.video_generator import NanoBananaVideoGenerator

generator = NanoBananaVideoGenerator()

# Generate video from text
video, metadata = generator.generate_video(
    prompt="A serene mountain landscape, camera slowly panning right",
    duration_seconds=10,
    resolution="1080p",
    fps=30
)
```
```

---

#### Use Case 3: Create Community Example Gallery

**Scenario**: Build a comprehensive example gallery with real-world use cases.

**Invocation**:

```python
Task(
    subagent_type="builder-skill",
    prompt="""
    Create a comprehensive example gallery for moai-connector-nano-banana.

    Categories (20 examples total):

    1. E-commerce (5 examples)
       - Product photography (clothing, electronics, food)
       - Lifestyle scenes (products in use)
       - 360-degree product views

    2. Marketing (5 examples)
       - Social media graphics (Instagram, LinkedIn)
       - Ad campaign visuals
       - Email newsletter headers

    3. UI/UX Design (5 examples)
       - App mockups
       - Website hero sections
       - Icon sets and illustrations

    4. Documentation (5 examples)
       - Technical diagrams
       - Tutorial illustrations
       - README banners

    Structure:
    - Create examples/gallery/ directory
    - Each example: README.md + generated images + prompts.txt
    - Difficulty levels: Beginner, Intermediate, Advanced
    - Real-world context for each example

    Documentation:
    - Add "Community Gallery" section to SKILL.md
    - Create examples/gallery/INDEX.md with overview
    - Include prompt engineering breakdowns
    - Add "Copy-paste ready" prompts

    Metadata:
    - Model used (Flash/Pro)
    - Resolution and aspect ratio
    - Generation time
    - Cost estimate
    - Difficulty level
    - Real-world use case description
    """
)
```

**Gallery Structure**:

```
examples/gallery/
├── INDEX.md (gallery overview)
├── ecommerce/
│   ├── product-photography/
│   │   ├── README.md (step-by-step guide)
│   │   ├── prompt.txt (copy-paste ready)
│   │   ├── output.png (generated image)
│   │   └── metadata.json (generation details)
│   ├── lifestyle-scenes/
│   └── ...
├── marketing/
├── ui-ux/
└── documentation/
```

**Example Gallery Entry**:

```markdown
# Product Photography: Wireless Headphones

## Difficulty: Intermediate
## Use Case: E-commerce Product Listing

### Context
Generate professional product photos for an e-commerce listing without
a physical photography studio.

### Prompt (Copy-Paste Ready)

```
A professional studio photograph of premium wireless headphones in matte black finish.
The headphones are positioned at a 45-degree angle on a seamless white background.
Soft diffused lighting from the top-left creates subtle shadows and highlights
the texture of the matte material. The headphone cups show fine detail and
craftsmanship. Shot with an 85mm macro lens at f/5.6 for moderate depth of field.
Studio-grade product photography, 4K resolution, photorealistic quality.
```

### Generation Details
- **Model**: Gemini 3 Pro (quality priority)
- **Resolution**: 4K (3840x2160)
- **Aspect Ratio**: 1:1 (e-commerce standard)
- **Generation Time**: 32 seconds
- **Cost Estimate**: $0.06
- **Tokens Used**: 3,890

### Result
![Generated Product Photo](output.png)

### Prompt Engineering Breakdown

1. **Subject**: "premium wireless headphones in matte black finish"
   - Specific product description
   - Material details (matte finish)

2. **Composition**: "positioned at a 45-degree angle"
   - Clear positioning instruction
   - Professional product photography standard

3. **Background**: "seamless white background"
   - E-commerce standard
   - Clean, distraction-free

4. **Lighting**: "Soft diffused lighting from top-left"
   - Directional light specification
   - Creates depth with shadows

5. **Technical**: "85mm macro lens at f/5.6"
   - Photography technical details
   - Moderate DOF for product clarity

6. **Quality**: "Studio-grade, 4K, photorealistic"
   - Quality indicators
   - Resolution specification

### Variations

Try these modifications:
- **Different angles**: "top-down view", "side profile"
- **Backgrounds**: "gradient gray", "wooden surface", "lifestyle setting"
- **Lighting**: "dramatic side lighting", "soft morning light"
- **Colors**: "white finish", "rose gold", "navy blue"

### Common Issues

**Issue**: Background not fully white
**Solution**: Add "pure white background (#FFFFFF), no shadows"

**Issue**: Image too soft
**Solution**: Increase f-stop: "f/8.0 for sharper focus"

### Next Steps

1. Generate multiple angles (front, side, top)
2. Create lifestyle scenes (person wearing headphones)
3. Generate color variations
4. Batch process for entire product line
```

---

### Skill Builder Best Practices

**DO**:
- ✅ Maintain backward compatibility
- ✅ Follow MoAI-ADK skill standards
- ✅ Document all changes thoroughly
- ✅ Add examples for new features
- ✅ Create tests for new functionality
- ✅ Update SKILL.md and modules/
- ✅ Preserve existing file structure
- ✅ Use progressive disclosure pattern

**DON'T**:
- ❌ Break existing API contracts
- ❌ Remove deprecated features without migration path
- ❌ Skip documentation updates
- ❌ Ignore test coverage
- ❌ Violate skill architecture patterns
- ❌ Create circular dependencies
- ❌ Exceed 500-line SKILL.md limit

---

### Integrated Workflow Example (통합 워크플로우 예제)

**Scenario**: E-commerce product photography campaign for 20 products.

#### Phase 1: Planning (`/moai:1-plan`)

```bash
/moai:1-plan "Create professional product photos for 20 smartphone accessories. Need 3 angles per product (front, side, lifestyle), studio quality, 4K resolution for print catalogs."
```

**SPEC Output**:
- SPEC-001: Product Photography Campaign
- 60 total images (20 products × 3 angles)
- Model: Gemini 3 Pro (4K quality)
- Estimated time: 32 minutes
- Estimated cost: $3.60
- Prompt templates for consistency

#### Phase 2: Execution (`/moai:2-run`)

```bash
/moai:2-run SPEC-001
```

**Alfred delegates to ai-nano-banana**:

```python
# Agent workflow
result = Task(
    subagent_type="ai-nano-banana",
    prompt="""
    Execute SPEC-001 product photography campaign.

    Products (20):
    1. Wireless charging pad
    2. Phone case (clear)
    3. Screen protector
    ... (17 more)

    For each product, generate 3 images:

    Angle 1 - Front View:
    "Professional studio photograph of [product] on white background.
    Soft diffused lighting, 85mm lens, f/5.6, 4K resolution."

    Angle 2 - Side Profile:
    "Side view of [product] showing depth and detail.
    Gradient gray background, dramatic side lighting, 4K."

    Angle 3 - Lifestyle Scene:
    "Lifestyle shot: [product] in modern home office setting.
    Natural window lighting, shallow DOF, photorealistic, 4K."

    Batch processing:
    - Model: Pro (quality priority)
    - Resolution: 4K
    - Aspect ratios: 1:1 (e-commerce), 16:9 (lifestyle)
    - Output: outputs/product-campaign/[product-name]/
    - Metadata: Save all prompts and generation details

    Quality validation:
    - Check resolution exactness (4K)
    - Verify background consistency
    - Ensure lighting coherence
    """
)
```

**Agent executes**:

1. **Request Analysis**: Parse 20 products, 3 angles each
2. **Prompt Engineering**: Create 60 optimized prompts
3. **Model Selection**: Pro model (4K requirement)
4. **Batch Generation**: Generate all 60 images
5. **Quality Validation**: Check resolution, consistency
6. **Iterative Refinement**: Regenerate low-quality images
7. **Metadata Logging**: Save all generation details

#### Phase 3: Documentation (`/moai:3-sync`)

```bash
/moai:3-sync SPEC-001
```

**Manager-docs creates**:

```markdown
# Product Photography Campaign - SPEC-001

## Summary
- Total images: 60 (20 products × 3 angles)
- Success rate: 98.3% (59/60 on first generation)
- Total time: 31m 45s
- Total cost: $3.54
- Quality score: 4.9/5.0

## Gallery

### Wireless Charging Pad
| Front | Side | Lifestyle |
|-------|------|-----------|
| ![](outputs/charging-pad/front.png) | ![](outputs/charging-pad/side.png) | ![](outputs/charging-pad/lifestyle.png) |

### Phone Case (Clear)
| Front | Side | Lifestyle |
|-------|------|-----------|
| ![](outputs/phone-case/front.png) | ![](outputs/phone-case/side.png) | ![](outputs/phone-case/lifestyle.png) |

... (18 more products)

## Prompts Used

### Front View Template
```
Professional studio photograph of [product] on pure white background.
Soft diffused lighting from top-left creates subtle shadows.
85mm macro lens, f/5.6, moderate depth of field.
Studio-grade product photography, 4K resolution (3840x2160), photorealistic.
```

### Generation Metadata
- Model: Gemini 3 Pro Image
- Average generation time: 31.75s per image
- Average tokens: 3,450 per image
- Total tokens used: 207,000
```

---

### Best Practices (모범 사례)

Maximize image generation quality and efficiency with these proven patterns.

#### Prompt Engineering Best Practices

**Structure Your Prompts**:
```
✅ GOOD (Structured narrative):
"A fluffy orange tabby cat with bright green eyes, delicately holding
a peeled banana in its paws. The cat sits on a sunlit windowsill with
soft morning light. 85mm portrait lens, shallow depth of field (f/2.8),
warm color palette. Studio-grade photography, 2K resolution."

❌ BAD (Keyword list):
"cat, banana, eating, cute"
```

**Layer Your Details**:
1. **Scene Foundation**: Subject + action + setting
2. **Photographic Elements**: Lighting + camera + lens
3. **Color & Style**: Palette + art style + mood
4. **Quality Standards**: Resolution + technical specs

**Specify Technical Details**:
- Lighting: "soft diffused", "golden hour", "dramatic side lighting"
- Camera: "85mm lens", "wide-angle 35mm", "macro close-up"
- Depth of field: "f/2.8 shallow DOF", "f/8 sharp focus"
- Quality: "studio-grade", "4K resolution", "photorealistic"

---

#### Model Selection Best Practices

**Choose Pro When**:
- ✅ Quality is top priority (advertising, branding)
- ✅ Need 4K resolution for print or large displays
- ✅ Require Google Search grounding for factual content
- ✅ Complex compositions with detailed elements
- ✅ Professional client deliverables

**Choose Flash When**:
- ✅ Speed is priority (rapid prototyping, iterations)
- ✅ Batch generation (10+ images)
- ✅ Testing prompt variations
- ✅ Internal use or draft quality
- ✅ Budget constraints

**Cost Optimization**:
```python
# Estimate costs before generation
def estimate_cost(num_images: int, model: str, resolution: str) -> float:
    cost_per_image = {
        ("flash", "1K"): 0.02,
        ("pro", "2K"): 0.04,
        ("pro", "4K"): 0.06
    }
    return num_images * cost_per_image.get((model, resolution), 0.04)

# Example: 20 images at 4K with Pro
total_cost = estimate_cost(20, "pro", "4K")  # $1.20
```

---

#### Iterative Refinement Best Practices

**Multi-Turn Strategy**:
1. **Turn 1**: Generate base image with detailed prompt
2. **Turn 2**: Refine specific elements ("make sky more dramatic")
3. **Turn 3**: Fine-tune details ("add warmer tones")
4. **Turn 4**: Final adjustments ("increase contrast slightly")
5. **Turn 5**: Final polish (if needed)

**Maximum Efficiency**:
- Limit to 5 turns maximum
- Each turn should be specific and targeted
- Avoid complete regeneration unless necessary
- Save intermediate results for comparison

---

#### Documentation and Organization

**File Naming Convention**:
```bash
outputs/
├── project-name/
│   ├── product-front-4K-001.png
│   ├── product-side-2K-002.png
│   └── lifestyle-16x9-003.png
```

**Metadata Management**:
```json
{
  "image": "product-front-4K-001.png",
  "timestamp": "2025-11-27T10:30:00Z",
  "model": "pro",
  "resolution": "4K",
  "aspect_ratio": "1:1",
  "prompt": "Professional studio photograph...",
  "generation_time_seconds": 28,
  "tokens_used": 3240,
  "cost_estimate_usd": 0.06
}
```

**Project Organization**:
```
project-root/
├── .env (API key - NEVER commit)
├── outputs/ (generated images - gitignored)
│   ├── campaign-01/
│   ├── campaign-02/
│   └── metadata/
├── prompts/ (reusable prompt templates)
└── docs/ (generation documentation)
```

---

#### Security and Privacy

**API Key Management**:
```bash
# ✅ CORRECT: Environment variable
GOOGLE_API_KEY=your_key_here

# ❌ WRONG: Hardcoded in code
api_key = "AIzaSyD..." # NEVER DO THIS

# ✅ CORRECT: Gitignore .env
echo ".env" >> .gitignore

# ✅ CORRECT: Secure permissions
chmod 600 .env
```

**Data Privacy**:
- Never include sensitive information in prompts
- Review images before sharing externally
- Delete test/draft images regularly
- Use `.gitignore` for outputs/ directory

**API Quota Management**:
- Monitor daily usage via Google AI Studio
- Set up quota alerts
- Use Flash model for development/testing
- Reserve Pro model for production

---

### Frequently Asked Questions (FAQ)

#### Model Selection

**Q: Which model should I use for product photography?**

A: Use **Gemini 3 Pro** for professional product photography. It supports 4K resolution, better detail rendering, and "thinking" process for composition optimization. Cost: ~$0.04-0.06 per image.

**Q: When should I use Flash instead of Pro?**

A: Use **Flash** when:
- Testing prompt variations (need speed)
- Generating 10+ images in batch
- Creating draft/internal visuals
- Budget is constrained
- 1K resolution is sufficient

---

#### Prompts and Quality

**Q: How do I write effective prompts?**

A: Follow the 4-layer structure:
1. **Scene**: "A modern office desk with laptop and coffee"
2. **Photography**: "Soft window light, 50mm lens, f/4"
3. **Style**: "Minimalist aesthetic, warm tones"
4. **Quality**: "Photorealistic, 2K resolution"

**Q: My images are blurry. How to fix?**

A: Add technical camera details:
- Specify higher f-stop: "f/8 for sharp focus"
- Mention "sharp focus" explicitly
- Use "macro lens" for product details
- Request "high detail" or "crisp quality"

**Q: Background not consistent. What to do?**

A: Be explicit about background:
- "Pure white background (#FFFFFF), seamless"
- "Gradient gray from #E5E5E5 to #F5F5F5"
- "Soft bokeh blur, out of focus background"

---

#### API and Configuration

**Q: Where do I get the Google API key?**

A: Visit https://aistudio.google.com/apikey
1. Sign in with Google account
2. Click "Create API Key"
3. Copy key to `.env` file
4. Never commit `.env` to git

**Q: How much does it cost per image?**

A: Approximate costs:
- **Flash (1K)**: $0.02 per image
- **Pro (2K)**: $0.04 per image
- **Pro (4K)**: $0.06 per image

Example: 100 images at 2K with Pro = $4.00

**Q: What if I hit quota limits?**

A:
1. Check quota at Google Cloud Console
2. Wait for quota reset (usually daily)
3. Request quota increase if needed
4. Use Flash model to reduce quota usage
5. Batch process during off-peak hours

---

#### Technical Issues

**Q: "Permission denied" error. What's wrong?**

A:
1. Verify API key in `.env` file is correct
2. Check key from https://aistudio.google.com/apikey
3. Ensure no extra spaces or quotes in `.env`
4. Restart terminal to reload environment variables

**Q: Images taking too long to generate?**

A:
- Pro model (4K): 30-60 seconds normal
- Pro model (2K): 20-35 seconds normal
- Flash model (1K): 5-15 seconds normal

If longer:
- Simplify prompt (reduce complexity)
- Try Flash model for speed
- Check internet connection
- Retry during off-peak hours

**Q: Safety filter blocked my prompt. Why?**

A:
- Gemini has content safety filters
- Avoid explicit, violent, or controversial content
- Rephrase using neutral, descriptive language
- Focus on positive, creative descriptions
- Example: Instead of "scary monster", use "fantastical creature"

---

#### Workflow Integration

**Q: Can I use nano-banana in CI/CD pipelines?**

A: Yes, but consider:
- API keys in CI secrets (not hardcoded)
- Quota limits for automated workflows
- Cache generated images to avoid regeneration
- Use Flash model for speed in CI

**Q: How to integrate with existing projects?**

A:
1. Add `moai-connector-nano-banana` to skills
2. Create `.env` with API key
3. Use via agents (ai-nano-banana, expert-frontend, etc.)
4. Document generated assets in project docs

**Q: Can I batch generate 100+ images?**

A: Yes, using batch processing:
```python
from modules.image_generator import NanoBananaImageGenerator

generator = NanoBananaImageGenerator()
prompts = load_prompts_from_file("prompts.txt")

results = generator.batch_generate(
    prompts=prompts,
    output_dir="outputs/batch-campaign",
    model="flash"  # Speed priority for large batches
)
```

Monitor quota usage carefully.

---

#### Advanced Features

**Q: How to use reference images?**

A:
```python
image, metadata = generator.generate_with_references(
    prompt="Create artwork in the style of these references",
    reference_images=[
        "style-ref-1.png",
        "composition-ref-2.png"
    ],
    model="pro"
)
```

Supports up to 14 reference images (6 objects + 5 persons).

**Q: What is Google Search grounding?**

A: Enables real-time data integration:
```python
image, metadata = generator.generate(
    prompt="Infographic about latest AI breakthroughs in 2025",
    model="pro",
    enable_google_search=True  # Fetch current info
)
```

Only available with Pro model.

**Q: Can I edit existing images?**

A: Yes, using image-to-image:
```python
edited, metadata = generator.edit(
    image_path="original.png",
    instruction="Change background to sunset beach scene",
    model="pro"
)
```

---

#### Best Practices Summary

**Quality Checklist**:
- ✅ Use structured prompts (4-layer approach)
- ✅ Select appropriate model (Pro vs Flash)
- ✅ Specify technical details (lighting, camera, lens)
- ✅ Include quality indicators (4K, photorealistic)
- ✅ Secure API key in `.env` (never commit)
- ✅ Document all prompts and metadata
- ✅ Organize outputs systematically
- ✅ Monitor API quota usage

**Efficiency Checklist**:
- ✅ Use Flash for testing and iterations
- ✅ Batch process similar images
- ✅ Reuse optimized prompt templates
- ✅ Cache generated images
- ✅ Limit multi-turn refinement to 5 turns max
- ✅ Delete unnecessary draft images

**Security Checklist**:
- ✅ Store API key in `.env` only
- ✅ Add `.env` to `.gitignore`
- ✅ Set `.env` permissions to 600
- ✅ Never include sensitive data in prompts
- ✅ Review images before external sharing
- ✅ Rotate API keys every 90 days

---


## Implementation Guide (5 minutes)

### Features

- Text-to-Image generation with 1K/2K/4K resolutions
- Image-to-Image editing and style transfer
- Multi-turn refinement for iterative improvements
- Reference image guidance (up to 14 references)
- Real-time Google Search grounding for factual content
- Advanced text rendering directly in images

### When to Use

- Generating professional visual assets for documentation or marketing
- Creating UI mockups and design concepts quickly
- Producing social media graphics and promotional images
- Illustrating technical documentation with custom diagrams
- Rapid prototyping of visual ideas before final design work

### Core Patterns

**Pattern 1: Structured Prompt for Quality**
```python
prompt = """
A serene Japanese garden at golden hour.
Lighting: warm sunset light filtering through maple trees.
Camera: wide-angle 35mm lens, low angle shot.
Composition: Rule of thirds, stone path leading to pagoda.
Color palette: warm gold, jade green, soft cream.
Style: photorealistic with slight cinematic color grading.
Quality: 4K resolution. Final output: PNG.
"""
```

**Pattern 2: Multi-Turn Refinement**
1. Generate initial image with base prompt
2. Review output and identify areas for improvement
3. Provide targeted refinement: "Make sky more dramatic"
4. Iterate up to 5 turns for perfect result

**Pattern 3: Reference-Guided Generation**
```python
# Use reference images to guide style
generate_image(
    prompt="Mountain landscape in the style of reference",
    reference_images=["style_ref.png", "composition_ref.png"],
    resolution="2K",
    aspect_ratio="16:9"
)
```

## 📚 Core Patterns (5-10 minutes)

### Pattern 1: Prompt Writing Best Practices

**Key Concept**: Well-structured prompts generate better images

**Good Prompt Characteristics**:
- **Clear Subject**: "Female CEO with blue glasses"
- **Style Specification**: "Matte acrylic painting style"
- **Composition Details**: "Standing on the left, hands on table"
- **Lighting Description**: "Warm golden hour lighting"
- **Details**: "Sharp focus, blurred background, cinematic quality"
- **Color Palette**: "Warm tones, photorealistic"

**Prompt Template**:
```
[Subject & Action]
A [adjective] [subject] doing [action].

[Setting & Composition]
Setting: [location] with [environmental details].
Composition: [framing], [positioning].

[Photographic Elements]
Lighting: [type], creating [mood].
Camera: [angle] shot with [lens] lens.
Depth of field: [shallow/deep].

[Color & Style]
Color palette: [colors].
Style: [art_style], [quality level].
```

**Professional Generation Example**:
```
A serene Japanese garden in spring, with cherry blossoms reflected
in a still pond. A single wooden bridge spans the water.

Soft golden hour lighting, photorealistic, cinematic composition,
shallow depth of field.
```

**Image Editing Example**:
```
Change the background from office to a modern coffee shop.
Keep the person in the same pose and lighting.
Maintain photorealistic style, warm tones.
```

### Pattern 2: Text-to-Image Generation

**Key Concept**: Generate professional images from text prompts

**Basic Flow**:
1. Write detailed, structured prompt
2. Choose resolution (1K, 2K, 4K)
3. Select aspect ratio (1:1, 16:9, 3:2, etc.)
4. Enable Google Search for current information (optional)
5. Generate and retrieve Base64 PNG

**Execution**:
```python
image_data = generate_image(
    prompt="Your detailed prompt here",
    resolution="2K",
    aspect_ratio="16:9",
    enable_google_search=True,  # For current info
    thinking_process=True        # Auto-optimize
)
```

### Pattern 3: Image-to-Image Editing

**Key Concept**: Transform existing images with detailed instructions

**Common Tasks**:
- **Style Transfer**: Convert to art style (Van Gogh, anime, etc.)
- **Object Manipulation**: Add, remove, or modify elements
- **Composition Change**: Reframe, zoom, or reposition subjects
- **Quality Enhancement**: Upscale, improve detail, adjust colors

**Flow**:
1. Load original image
2. Write transformation instruction
3. Reference images (optional)
4. Apply edit maintaining coherence
5. Retrieve edited image

### Pattern 4: Multi-Turn Refinement

**Key Concept**: Iteratively improve images through conversation

**Workflow**:
1. Generate initial image
2. Review output
3. Provide refinement instruction
4. Regenerate with improvements
5. Repeat (max 5 turns)

**Example**:
```
Turn 1: "A mountain landscape at sunset"
Turn 2: "Make the sky more dramatic with purple clouds"
Turn 3: "Add a lone tree in foreground"
```

### Pattern 5: Reference Image Guidance

**Key Concept**: Use reference images to guide generation style

**Supported References**:
- Up to 6 object references
- Up to 5 human references
- Style influences
- Composition guides

**Usage**:
```python
generate_image(
    prompt="Similar style to reference",
    reference_images=[
        "path/to/style_reference.png",
        "path/to/composition_ref.png"
    ]
)
```

---

## 📖 Advanced Documentation

This Skill uses Progressive Disclosure. For detailed implementation:

- **[modules/prompt-engineering.md](modules/prompt-engineering.md)** - Professional prompt templates
- **[modules/api-reference.md](modules/api-reference.md)** - Complete API documentation
- **[modules/examples.md](modules/examples.md)** - Real-world usage examples
- **[modules/troubleshooting.md](modules/troubleshooting.md)** - Common issues and solutions

---

## 🎨 Model Selection Guide

**Gemini 3 Pro Image (Nano Banana Pro)**:
- **Use for**: Professional asset creation, complex guidance
- **Features**: Google Search grounding, "thinking" process, up to 4K resolution
- **Speed**: Slower (quality-first)
- **Cost**: Higher
- **Best for**: Advertising materials, high-end illustrations, real-time data integration

**Gemini 2.5 Flash Image (Nano Banana)**:
- **Use for**: Speed and efficiency
- **Features**: Optimized for batch work, low latency, 1K resolution
- **Speed**: Fast
- **Cost**: Lower
- **Best for**: Quick prototyping, batch processing, variation generation

**Decision Criteria**:
- Speed priority → Flash
- Quality priority → Pro
- Real-time grounding needed → Pro
- High resolution (4K) needed → Pro
- Batch generation → Flash
- Cost minimization → Flash

---

## Quick Reference (30 seconds)

**Core Purpose**: Professional AI image generation using Nano Banana Pro (Gemini 3 Pro) and Gemini 2.5 Flash.

**Key Features**: Text-to-image, image-to-image editing, multi-turn refinement, reference guidance, 4K resolution.

**When to Use**: Visual asset creation, prototyping, documentation, UI mockups, marketing materials.

---

## Works Well With

**Agents**:
- **design-uiux** - UI/UX design integration
- **code-frontend** - Frontend asset implementation
- **workflow-docs** - Visual documentation generation

**Skills**:
- **moai-lang-unified** - UI/UX implementation with generated assets
- **moai-docs-generation** - Create visual documentation
- **moai-cc-claude-md** - Embed generated images in markdown
- **moai-domain-frontend** - Frontend integration

**Commands**:
- `/moai:3-sync` - Documentation with visual assets
- `/moai:9-feedback` - Image generation improvements

---

## 🔗 Integration with Other Skills

**Typical Workflow**:
1. Use this Skill to generate visual assets
2. Use moai-domain-frontend to implement in UI
3. Use moai-docs-generation to document with images

---

## 📈 Version History

**1.0.1** (2025-11-23)
- 🔄 Refactored with Progressive Disclosure pattern
- 📚 Detailed prompts moved to modules/
- ✨ Core patterns highlighted in SKILL.md
- ✨ Added model selection guide

**1.0.0** (2025-11-12)
- ✨ Nano Banana Pro (Gemini 3 Pro) support
- ✨ Text-to-Image and Image-to-Image
- ✨ Multi-turn refinement capability
- ✨ Reference image guidance

---

**Maintained by**: alfred
**Domain**: Image Generation & Visual Creation
**Generated with**: MoAI-ADK Skill Factory
