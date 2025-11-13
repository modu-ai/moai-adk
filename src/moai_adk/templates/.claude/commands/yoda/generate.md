---
name: yoda:generate
description: "Generate educational lecture materials automatically via Master Yoda agent"
argument-hint: "--topic 'topic name' [options] --instructor, --audience, --difficulty, --format, --output"
allowed-tools:
- Read
- Write
- Edit
- MultiEdit
- Grep
- Glob
- TodoWrite
- AskUserQuestion
- Task
- Skill
model: "sonnet"
---

# /yoda:generate Command

> Automatically generate educational lecture materials via Master Yoda agent.
> Markdown → PDF, PowerPoint, Word, Notion simultaneous generation

## Usage

### Basic Format

```bash
/yoda:generate --topic "topic name" [options]
```

### Required Arguments

- `--topic "topic name"` - Lecture topic (example: "FastAPI Fundamentals")

### Optional Arguments

```bash
--instructor "instructor name"   # Instructor name (default: "MoAI Instructor")
--audience "audience"            # Target audience (default: "developers")
--difficulty "difficulty"        # easy/medium/hard (default: "medium")
--format "format"                # education/presentation/workshop (default: "education")
--output "formats"               # pdf/pptx/docx/notion (default: "pdf,pptx")
--description "description"      # Lecture overview (default: "")
--duration "hours"               # Expected duration (default: "2 hours")

# Notion Enhancement Options (mcp-notion-integrator integration)
--notion-enhanced                # Enable AI-optimized Notion generation
--notion-database "DB name"      # Target Notion database (default: "Lectures")
--notion-template "template"     # enhanced/optimized/interactive (default: "enhanced")
--notion-workflow "workflow"     # auto-publish/draft/review (default: "auto-publish")
--notion-analytics true          # Enable student engagement analytics
--notion-tags "tags"             # Notion tags (comma-separated)
--notion-category "category"     # Lecture category (example: "programming", "design")
```

---

## 📝 Usage Examples

### Example 1: Generate Basic Lecture

```bash
/yoda:generate --topic "FastAPI Fundamentals"
```

**Result**:
- Format: education.md (default)
- Output: PDF, PPTX (default)
- Saved: `.moai/yoda/output/fastapi-fundamentals.md` etc.

---

### Example 2: Generate With Full Options

```bash
/yoda:generate \
  --topic "FastAPI Fundamentals" \
  --instructor "Alice Johnson" \
  --audience "Python Developers" \
  --difficulty "medium" \
  --format "education" \
  --output "pdf,pptx,docx,notion" \
  --description "From basic concepts to hands-on practice" \
  --duration "2 hours"
```

**Result**:
- Markdown: `.moai/yoda/output/fastapi-fundamentals.md`
- PDF: `.moai/yoda/output/fastapi-fundamentals.pdf`
- PPTX: `.moai/yoda/output/fastapi-fundamentals.pptx`
- DOCX: `.moai/yoda/output/fastapi-fundamentals.docx`
- Notion: Auto-generated public page + link provided

---

### Example 3: Generate Presentation

```bash
/yoda:generate \
  --topic "MoAI-ADK Architecture" \
  --format "presentation" \
  --output "pptx,notion"
```

**Result**:
- Format: presentation.md
- Output: PPTX, Notion
- Saved: `.moai/yoda/output/moai-adk-architecture.pptx` etc.

---

### Example 4: Generate Workshop Materials

```bash
/yoda:generate \
  --topic "React Component Design" \
  --format "workshop" \
  --output "docx,pdf" \
  --difficulty "hard" \
  --duration "4 hours"
```

**Result**:
- Format: workshop.md
- Output: DOCX, PDF
- Saved: `.moai/yoda/output/react-components-workshop.docx` etc.

---

### Example 5: Generate AI-Enhanced Notion Lecture (NEW)

```bash
/yoda:generate \
  --topic "Machine Learning Fundamentals" \
  --format "education" \
  --notion-enhanced \
  --notion-database "ML-Courses" \
  --notion-template "enhanced" \
  --notion-workflow "auto-publish" \
  --notion-analytics true \
  --notion-tags "ml,basics,ai" \
  --notion-category "programming" \
  --output "notion"
```

**Result**:
- AI-optimized Notion database page auto-generated
- Lecture structuring, tagging, analytics features included
- Intelligent optimization via mcp-notion-integrator agent

---

### Example 6: Interactive Notion Workshop (NEW)

```bash
/yoda:generate \
  --topic "UI/UX Design Principles" \
  --format "workshop" \
  --notion-enhanced \
  --notion-template "interactive" \
  --notion-database "Design-Workshops" \
  --notion-workflow "draft" \
  --notion-category "design" \
  --output "notion,pdf"
```

**Result**:
- Notion workshop page with interactive elements
- Student engagement tracking features
- Simultaneous PDF and Notion page generation

---

## 🎯 Format Guide

### education - Theory Lecture

**Structure**: Introduction → Theory → Examples → Practice → Summary

**Features**:
- ✅ Theory-focused lecture
- ✅ Step-by-step learning structure
- ✅ Examples and practice included
- ✅ Quizzes and validation

**Use Cases**:
- Concept explanation lectures
- Technology tutorials
- Online courses

---

### presentation - Presentation

**Structure**: Opening → Key Concepts → Visualization → Discussion → Closing

**Features**:
- ✅ Visual content-focused
- ✅ Interactive discussion
- ✅ Data and case studies
- ✅ Concise message delivery

**Use Cases**:
- Conference presentations
- Team meetings
- Special lectures

---

### workshop - Hands-on Workshop

**Structure**: Orientation → Live Practice → Team Projects → Review → Wrap-up

**Features**:
- ✅ 100% practice-focused
- ✅ Hands-on projects
- ✅ Team collaboration
- ✅ Code review

**Use Cases**:
- Hands-on workshops
- Bootcamps
- Skill development

---

## 📦 Output Format Guide

### pdf - PDF Document

**Library**: reportlab

**Features**:
- ✅ Classic print format
- ✅ Identical display across environments
- ✅ Small file size
- ✅ Easy distribution/sharing

**Use Cases**: Printing, email sharing, official documents

---

### pptx - PowerPoint Presentation

**Script**: html2pptx.js (reused)

**Features**:
- ✅ Slide format
- ✅ Animation support
- ✅ Editable
- ✅ Office compatibility

**Use Cases**: Presentations, live discussions, meetings

---

### docx - Word Document

**Script**: document.py (reused)

**Features**:
- ✅ Linear document format
- ✅ Editable
- ✅ Word compatible
- ✅ Table of contents/index capable

**Use Cases**: Detailed documents, textbooks, reference materials

---

### notion - Notion Page

**Tool**: Notion MCP integration + mcp-notion-integrator agent

**Basic Features**:
- ✅ Online publishing
- ✅ Collaborative editing enabled
- ✅ Auto-published
- ✅ Web link provided

**AI Enhancement Features** (--notion-enhanced):
- ✅ AI-optimized content structure
- ✅ Intelligent database field mapping
- ✅ Student engagement analytics
- ✅ Auto tagging and categorization
- ✅ Interactive element integration

**Use Cases**: Online sharing, team collaboration, web publishing, education management

**Template Options**:
- `enhanced`: Basic AI-optimized template
- `optimized`: Performance-optimized template
- `interactive`: Interactive elements included template

**Workflow Options**:
- `auto-publish`: Immediately public (default)
- `draft`: Save as draft
- `review`: Mark as review needed

---

## ⏱️ Difficulty Guide

### easy - Beginner

**Features**:
- Foundation concepts focus
- Many examples and explanations
- Step-by-step approach
- Simple practice

**Target**: Complete beginners

---

### medium - Intermediate (default)

**Features**:
- Balanced theory and practice
- Intermediate-level examples
- Real-world use cases
- Challenge tasks included

**Target**: Learners with basic knowledge

---

### hard - Advanced

**Features**:
- Advanced concepts
- Complex examples
- In-depth practice
- Professional use cases

**Target**: Experienced developers

---

## 📋 Generation Result Verification

### Generated File Location

All files are saved in `.moai/yoda/output/`:

```
.moai/yoda/output/
├── {topic}.md              # Original markdown
├── {topic}.pdf             # PDF document
├── {topic}.pptx            # PowerPoint
├── {topic}.docx            # Word document
└── {topic}-notion-link.txt # Notion link
```

### File Naming Convention

For topic "FastAPI Fundamentals":
- `fastapi-fundamentals.md`
- `fastapi-fundamentals.pdf`
- `fastapi-fundamentals.pptx`
- `fastapi-fundamentals.docx`

Spaces converted to hyphens, special characters removed

---

## 🔍 Advanced Usage

### Multiple Format Combinations

```bash
# Generate PDF only
/yoda:generate --topic "topic" --output "pdf"

# Generate all formats
/yoda:generate --topic "topic" --output "pdf,pptx,docx,notion"

# Generate presentation + Notion only
/yoda:generate --topic "topic" --format "presentation" --output "pptx,notion"
```

### Batch Generation (Multiple Lectures)

```bash
# Repeat execution for multiple lectures
/yoda:generate --topic "Lecture 1"
/yoda:generate --topic "Lecture 2"
/yoda:generate --topic "Lecture 3"
```

### Overwrite Existing Files

Regenerating with the same topic automatically overwrites:

```bash
# First generation
/yoda:generate --topic "FastAPI Fundamentals" --format "education"

# Regenerate as presentation (overwrites)
/yoda:generate --topic "FastAPI Fundamentals" --format "presentation"
```

---

## ⚡ Performance Tips

### Fast Generation

```bash
# Markdown only (fastest)
/yoda:generate --topic "topic" --output ""
```

### Final Output Generation

```bash
# Markdown + all formats (most time)
/yoda:generate --topic "topic" --output "pdf,pptx,docx,notion"
```

### Estimated Time

| Output Format | Estimated Time |
|---------------|----------------|
| Markdown | < 1s |
| Markdown + PDF | ~2s |
| Markdown + PPTX | ~2s |
| Markdown + DOCX | ~2s |
| Markdown + All | ~5s |
| Markdown + Notion | ~3s |

---

## ❓ Frequently Asked Questions

**Q1**: "Can I edit generated files?"

**A1**: Yes! Markdown, PDF, PPTX, DOCX are all editable. Regenerate as needed.

---

**Q2**: "Are Notion pages private when auto-published?"

**A2**: No, they're set to public and a share link is provided. Modify permissions in Notion as needed.

---

**Q3**: "Can I include images or videos?"

**A3**: Currently text-based generation only. Add images manually after generation.

---

**Q4**: "Can I use custom templates?"

**A4**: Three standard templates provided (education/presentation/workshop). Modify markdown after generation as needed.

---

## 🛠️ Troubleshooting

### Issue 1: Notion Publishing Failed

**Solution**:
1. Verify Notion MCP settings in `.claude/mcp.json`
2. Check Notion API key
3. Verify database access permissions

### Issue 2: File Generation Failed

**Solution**:
1. Verify `.moai/yoda/output/` directory exists
2. Check disk space
3. Verify required libraries installed (reportlab, python-pptx, etc.)

### Issue 3: Special Character Handling Error

**Solution**:
- Remove special characters from topic name
- Non-ASCII characters automatically converted

---

## 🚀 Next Steps

Recommended tasks after generation:

1. **Review Markdown**: Verify variable substitutions are correct
2. **Review PDF/PPTX**: Verify formatting as expected
3. **Share Notion**: Share link with team if needed
4. **Collect Feedback**: Gather learner feedback for improvements

---

*This command invokes the Master Yoda agent to automatically generate educational materials.*
