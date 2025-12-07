# Yoda System Usage Examples

Real-world scenarios demonstrating how to use the Yoda System templates.

## Example 1: Generate Theory-Focused Education Material

**Scenario**: Create a beginner-friendly guide on Python async programming

```bash
/yoda:generate \
  --topic "Python 비동기 프로그래밍" \
  --format "education" \
  --audience "초급 개발자" \
  --difficulty "basic" \
  --output "md,pdf,docx"
```

**Execution Process**:

1. **Load Template**: education.md
2. **Adapt for Difficulty**: Basic
   - Add more example code (13 blocks total)
   - Increase explanatory text
   - Simplify practice problems
   - Add more hints and guidance
3. **Fetch Documentation**: Context7 MCP
   - Get Python asyncio official docs
   - Get best practices from python.org
4. **Generate Markdown**:
   - Replace all `{{VARIABLES}}`
   - Inject Context7 documentation links
   - Create final content file
5. **Convert Formats**:
   - MD → PDF (via pandoc or wkhtmltopdf)
   - MD → DOCX (via pandoc)
6. **Save Outputs**:
   - `.moai/yoda/output/python-비동기-프로그래밍.md`
   - `.moai/yoda/output/python-비동기-프로그래밍.pdf`
   - `.moai/yoda/output/python-비동기-프로그래밍.docx`

**Output Structure**:
```
python-비동기-프로그래밍.md
├─ Overview (goals, prerequisites, time estimate)
├─ Core Concepts
│  ├─ Basic Principles (asyncio fundamentals)
│  ├─ Practical Application (coroutines, tasks)
│  └─ Advanced Topics (optimization, debugging)
├─ Practice Problems
│  ├─ Basic (2 problems with solutions)
│  └─ Advanced (2 problems with solutions)
├─ References (official docs, tutorials, resources)
└─ Summary (learning path, next steps)
```

**Generated Files**:
- **PDF**: Professional document for printing/sharing
- **DOCX**: Editable document for customization
- **MD**: Source format for updates

---

## Example 2: Generate Presentation with Context7

**Scenario**: Create a conference presentation on Next.js 14 features

```bash
/yoda:generate \
  --topic "Next.js 14 최신 기능" \
  --format "presentation" \
  --instructor "Alice Park" \
  --audience "개발자 커뮤니티" \
  --presentation-time "45분" \
  --output "md,pptx,pdf"
```

**Execution Process**:

1. **Load Template**: presentation.md (18 slides)
2. **Fetch Documentation**: Context7 MCP
   - Get Next.js 14 official release notes
   - Get new features documentation
   - Get best practices and examples
3. **Customize Slides**:
   - Slide 1: Title with instructor, audience, time
   - Slides 2-4: Overview, learning goals, context
   - Slides 5-7: Core concepts (3 major features)
   - Slides 8-10: Practical patterns & case studies
   - Slides 11-15: Best practices, trends, action plan
   - Slides 16-18: Q&A, summary, closing
4. **Inject Documentation**:
   - Replace `{{CONCEPT_1_DEFINITION}}` with actual docs
   - Add code examples from official repository
   - Include latest API information
5. **Generate Formats**:
   - MD with YAML frontmatter (presenter's copy)
   - PPTX (for presentation to audience)
   - PDF (for distribution after talk)
6. **Save Outputs**:
   - `.moai/yoda/output/nextjs-14-최신-기능.md`
   - `.moai/yoda/output/nextjs-14-최신-기능.pptx`
   - `.moai/yoda/output/nextjs-14-최신-기능.pdf`

**Slide Output Breakdown**:
```
Slide 1: Title (Next.js 14 최신 기능)
Slide 2: Overview (Journey outline - 45 min agenda)
Slide 3: Learning Goals (4 main objectives)
Slide 4: Background (Current trends in React ecosystem)
Slide 5: Concept 1 (Server Components)
Slide 6: Concept 2 (Incremental Static Regeneration)
Slide 7: Concept 3 (Edge Functions)
Slide 8: Practical Patterns (3 implementation examples)
Slide 9: Case Study 1 (Company A - migration story)
Slide 10: Case Study 2 (Company B - performance results)
Slide 11: Success/Failure Patterns (best practices)
Slide 12: Advanced Topics (optimization strategies)
Slide 13: Trends & Future (2025-2027 roadmap)
Slide 14: Action Plan (3-phase implementation)
Slide 15: FAQ (5 common questions)
Slide 16: Q&A Session (facilitation guide)
Slide 17: Key Takeaways (3 main messages)
Slide 18: Closing (Resources, contact, thanks)
```

---

## Example 3: Generate Hands-On Workshop with Notion

**Scenario**: Create a Docker optimization bootcamp

```bash
/yoda:generate \
  --topic "Docker 컨테이너 최적화" \
  --format "workshop" \
  --instructor "Bob Kim" \
  --audience "DevOps 팀" \
  --difficulty "advanced" \
  --total-duration "6시간" \
  --notion-enhanced \
  --output "md,pdf,notion"
```

**Execution Process**:

1. **Load Template**: workshop.md (928 lines)
2. **Adapt for Advanced Level**:
   - Create 2 hands-on labs (basic → advanced)
   - Design capstone team project
   - Include production patterns
   - Add performance optimization content
3. **Fetch Documentation**: Context7 MCP
   - Get official Docker documentation
   - Get latest best practices
   - Get security guidelines
4. **Generate Markdown Content**:
   - Section 1: Workshop goals & prerequisites
   - Section 2: Environment setup (30 min)
   - Section 3: Lab 1 - Basics (45 min)
   - Section 4: Lab 2 - Advanced (45 min)
   - Section 5: Team project - Production setup (4-6 hours)
   - Section 6: Troubleshooting
5. **Convert Formats**:
   - MD → PDF (printable workbook)
   - MD → DOCX (editable document)
   - MD → Notion (knowledge base)
6. **Publish to Notion** (if --notion-enhanced):
   - Create public database
   - Add workshop metadata
   - Link to resources
   - Enable team collaboration
7. **Save Outputs**:
   - `.moai/yoda/output/docker-컨테이너-최적화.md`
   - `.moai/yoda/output/docker-컨테이너-최적화.pdf`
   - `.moai/yoda/output/docker-컨테이너-최적화-notion-link.txt`

**Workshop Structure**:
```
Workshop: Docker Container Optimization (Advanced, 6 hours)

1. Setup & Prerequisites (30 min)
   - Docker installation & verification
   - Required tools setup
   - Environment validation

2. Hands-On Lab 1: Image Optimization (45 min)
   - Step 1: Multi-stage builds
   - Step 2: Layer caching
   - Step 3: Image size reduction
   - Validation checkpoint

3. Hands-On Lab 2: Production Deployment (45 min)
   - Step 1: Health checks & networking
   - Step 2: Secrets & configuration
   - Step 3: Monitoring & logging
   - Step 4: Performance optimization (advanced)

4. Team Capstone Project (4-6 hours)
   - Scenario: Deploy microservices application
   - Team roles: Leader, Developers (2), QA
   - Requirements:
     * Multi-container setup (3+ services)
     * Docker Compose orchestration
     * Health checks & restart policies
     * Performance benchmarks
   - Evaluation: 40 pts functionality, 30 pts code quality, 15 pts testing, 10 pts documentation
   - Timeline: Design (Day 1-2), Development (Day 3-5), Testing (Day 5), Submission (Day 6)

5. Troubleshooting Guide (5+ scenarios)
   - Network connectivity issues
   - Resource constraints
   - Image layer problems
   - Performance bottlenecks
   - Security best practices
```

**Notion Publication**:
- Database: "Workshops"
- Title: "Docker Container Optimization"
- Properties: instructor (Bob Kim), audience (DevOps), difficulty (Advanced)
- Cover: Docker logo
- Public URL: `https://notion.so/workshop-docker-optimization`
