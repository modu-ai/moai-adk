---
name: alfred:research
description: "Research management and coordination command"
argument-hint: [action] [topic] [options]
allowed-tools:
- Read
- Write
- Edit
- Grep
- Glob
- TodoWrite
- Bash(git:*)
- WebSearch
- mcp__context7__resolve-library-id
- mcp__context7__get-library-docs
- mcp__playwright__browser_navigate
- mcp__playwright__browser_snapshot
- mcp__sequential-thinking__sequentialthinking
---

# 🔬 MoAI-ADK Research Command - Research Management & Coordination

> **Critical Note**: ALWAYS invoke `Skill("moai-alfred-ask-user-questions")` before using `AskUserQuestion` tool. This skill provides up-to-date best practices, field specifications, and validation rules for interactive prompts.
>
> **Batched Design**: All AskUserQuestion calls follow batched design principles (1-4 questions per call) to minimize user interaction turns. See CLAUDE.md section "Alfred Command Completion Pattern" for details.

<!-- @CODE:ALF-WORKFLOW-001:CMD-RESEARCH -->

**Research Integration**: This command integrates research capabilities with the MoAI-ADK workflow, enabling systematic research management and coordination.

## 🎯 Command Purpose

**"Research → Analyze → Integrate"** - Comprehensive research management system that coordinates research activities, manages TAG-based research, and provides research dashboards.

**Research for**: $ARGUMENTS

## 🔬 Research Management Architecture

### Research Command Structure
```
/alfred:research [action] [topic] [options]

Actions:
- status: Show research dashboard and system status
- search: Search research topics across TAG system
- analyze: Deep analysis using research engines
- integrate: Integrate research findings into SPEC/TEST/CODE
- optimize: Research performance optimization
- validate: Research validation and verification

Topics:
- Any research topic, technology, or domain
- TAG-based research (@RESEARCH, @PATTERN, @SOLUTION)
- Multi-domain research patterns
- Performance and optimization research

Options:
--depth=shallow|medium|deep: Research depth level
--engines=python|context7|playwright|sequential: Specific research engines
--export=json|markdown|yaml: Export research results
--integrate: Auto-integrate findings into project
```

## 📋 Your Task

You are executing the `/alfred:research` command. Your job is to provide comprehensive research management and coordination using the enhanced research infrastructure.

The command has **FOUR execution phases**:

1. **PHASE 1**: Research Intent Analysis & Planning
2. **PHASE 2**: Research Execution & Data Collection
3. **PHASE 3**: Research Analysis & Integration
4. **PHASE 4**: Research Reporting & Documentation

Each phase contains explicit step-by-step instructions.

---

## 🔍 PHASE 1: Research Intent Analysis & Planning

### Step 1: Parse research arguments

Analyze the user's `$ARGUMENTS` to understand research intent:

**Required elements**:
- **action**: What type of research operation (status, search, analyze, integrate, optimize, validate)
- **topic**: Research subject area or domain
- **options**: Additional parameters for research execution

**Action mapping**:
- `status` → Research dashboard and system overview
- `search` → TAG-based research search
- `analyze` → Deep research analysis
- `integrate` → Research integration into project
- `optimize` → Performance and optimization research
- `validate` → Research validation and verification

**Research depth levels**:
- `shallow`: Quick overview (2-3 minutes)
- `medium`: Standard analysis (5-10 minutes)
- `deep`: Comprehensive research (15+ minutes)

### Step 2: Research Planning

Create research execution plan based on parsed arguments:

**Research Engine Selection**:
- **Python Engines**: `knowledge_integration_hub`, `cross_domain_analysis_engine`, `pattern_recognition_engine`
- **MCP Context7**: Documentation and API research
- **MCP Playwright**: Web interface and user experience research
- **MCP Sequential Thinking**: Complex reasoning and analysis
- **TAG System**: @RESEARCH, @PATTERN, @SOLUTION integration

**Resource Allocation**:
- Determine required skills and agents
- Estimate research time requirements
- Plan integration points with existing project

### Step 3: Research Scope Validation

**Ask user for research confirmation**:

```
"Research plan ready. Execute the following research?

**Research Scope**: [action] on [topic]
**Depth Level**: [depth]
**Engines**: [selected engines]
**Estimated Time**: [time estimate]
**Integration Plan**: [integration approach]"
```

**Present options**:
1. **Execute Research** - Start research execution as planned
2. **Modify Scope** - Adjust research parameters
3. **Change Engines** - Select different research engines
4. **Cancel** - Cancel research operation

---

## 🚀 PHASE 2: Research Execution & Data Collection

### Step 1: Engine Initialization

Based on user confirmation, initialize selected research engines:

**Python Research Engines**:
```python
# Knowledge Integration Hub
import knowledge_integration_hub
hub = knowledge_integration_hub.KnowledgeIntegrationHub()
hub.initialize_research_context(topic, depth)

# Cross-Domain Analysis Engine
import cross_domain_analysis_engine
analyzer = cross_domain_analysis_engine.CrossDomainAnalyzer()
analyzer.setup_analysis_parameters(topic, domains)

# Pattern Recognition Engine
import pattern_recognition_engine
recognizer = pattern_recognition_engine.PatternRecognizer()
recognizer.configure_pattern_detection(topic, patterns)
```

**MCP Research Engines**:
```python
# Context7 for documentation research
if 'context7' in engines:
    library_docs = mcp__context7__resolve-library_id(topic)
    research_data = mcp__context7__get-library-docs(library_docs, tokens=5000)

# Playwright for web research
if 'playwright' in engines:
    browser_data = mcp__playwright__browser_navigate(url)
    interface_analysis = mcp__playwright__browser_snapshot()

# Sequential Thinking for complex analysis
if 'sequential' in engines:
    thinking_process = mcp__sequential-thinking__sequentialthinking(
        thought="Initial research analysis for [topic]",
        nextThoughtNeeded=True,
        thoughtNumber=1,
        totalThoughts=estimated_thoughts
    )
```

### Step 2: Data Collection Strategy

**TAG-Based Research**:
```bash
# Search existing research TAGs
rg "@RESEARCH:" -n .moai/specs/ src/ tests/ docs/
rg "@PATTERN:" -n .moai/specs/ src/ tests/ docs/
rg "@SOLUTION:" -n .moai/specs/ src/ tests/ docs/
```

**Web Research Integration**:
```python
# Web search for latest information
web_results = WebSearch(query=f"{topic} latest research 2025")
```

**Project Context Integration**:
```python
# Analyze project-specific context
project_specs = Glob(pattern=".moai/specs/*/spec.md")
existing_code = Glob(pattern=f"src/**/*{topic}*")
test_coverage = Glob(pattern=f"tests/**/*{topic}*")
```

### Step 3: Research Data Processing

**Data Aggregation**:
- Collect results from all engines
- Normalize data formats
- Apply quality filters
- Remove duplicates and noise

**Data Validation**:
- Verify data integrity
- Check for contradictions
- Validate source credibility
- Ensure relevance to research topic

---

## 🔬 PHASE 3: Research Analysis & Integration

### Step 1: Analysis Execution

**Invoke Research Coordinator Agent**:
```
Tool: Task
Parameters:
- subagent_type: "research-coordinator"
- description: "Analyze research data and provide insights"
- prompt: """You are the research-coordinator agent.

Language settings:
- conversation_language: {{CONVERSATION_LANGUAGE}}
- language_name: {{CONVERSATION_LANGUAGE_NAME}}

TASK:
Analyze collected research data and provide comprehensive insights:

**Research Topic**: [topic]
**Research Action**: [action]
**Data Sources**: [list of engines and sources]
**Research Depth**: [depth]

**Analysis Requirements**:
1. **Pattern Recognition**: Identify key patterns and trends
2. **Cross-Domain Analysis**: Connect findings across domains
3. **Knowledge Integration**: Synthesize information from multiple sources
4. **Practical Applications**: Identify actionable insights
5. **Integration Points**: Determine how to integrate findings into project

**Output Format**:
- Executive Summary
- Key Findings
- Patterns and Trends
- Integration Recommendations
- Next Steps

Output language: {{CONVERSATION_LANGUAGE}}"""
```

### Step 2: Integration Planning

**Integration Strategy**:
- **SPEC Integration**: Update existing SPECs or create new ones
- **Code Integration**: Generate code implementations based on research
- **Test Integration**: Create test cases for research findings
- **Documentation**: Update project documentation

**TAG Assignment**:
```python
# Assign research TAGs
@RESEARCH:[topic]-[timestamp]: Research findings and insights
@PATTERN:[domain]-[pattern]: Identified patterns and solutions
@SOLUTION:[problem]-[solution]: Practical solutions and implementations
```

### Step 3: Validation & Verification

**Research Validation**:
- Cross-check findings with multiple sources
- Verify technical accuracy
- Validate against project requirements
- Check for conflicts with existing implementations

**Quality Assurance**:
- Apply TRUST 5 principles to research findings
- Ensure traceability of research sources
- Validate integration points
- Test proposed solutions

---

## 📊 PHASE 4: Research Reporting & Documentation

### Step 1: Research Dashboard Creation

**Dashboard Components**:
```markdown
# Research Dashboard: [topic]

## Research Summary
- **Topic**: [research topic]
- **Action**: [research action]
- **Depth**: [research depth]
- **Engines Used**: [list of engines]
- **Execution Time**: [duration]

## Key Findings
### Patterns Identified
- [Pattern 1 with @PATTERN tag]
- [Pattern 2 with @PATTERN tag]

### Solutions Discovered
- [Solution 1 with @SOLUTION tag]
- [Solution 2 with @SOLUTION tag]

### Research Insights
- [Insight 1 with @RESEARCH tag]
- [Insight 2 with @RESEARCH tag]

## Integration Recommendations
### SPEC Updates
- [SPEC-XXX: Update required]
- [New SPEC proposal: SPEC-YYY]

### Code Implementations
- [Module: implementation suggestions]
- [Function: code improvements]

### Documentation Updates
- [Document: updates needed]
- [New doc: creation suggested]

## Next Steps
1. [Immediate action item]
2. [Follow-up research needed]
3. [Long-term implementation plan]
```

### Step 2: Research Documentation

**Create Research Documentation**:
```bash
# Create research documentation directory
mkdir -p .moai/research/[topic]/

# Create research report
Write file: .moai/research/[topic]/research_report.md
```

**Export Research Data**:
```python
# Export in requested format
if export_format == 'json':
    export_research_json(research_data)
elif export_format == 'markdown':
    export_research_markdown(research_data)
elif export_format == 'yaml':
    export_research_yaml(research_data)
```

### Step 3: Integration Execution (if --integrate specified)

**Automatic Integration**:
```python
# Update SPEC documents
for spec_update in spec_updates:
    update_spec_document(spec_update['id'], spec_update['changes'])

# Generate code implementations
for code_suggestion in code_suggestions:
    generate_code_implementation(code_suggestion)

# Create test cases
for test_case in test_cases:
    create_test_case(test_case)

# Update documentation
for doc_update in doc_updates:
    update_documentation(doc_update)
```

---

## ✅ Command Completion & Next Steps

After all phases complete successfully, you MUST ask the user what to do next.

### Ask the user this question:

"Research execution is complete. What would you like to do next?"

### Present these options:

1. **Implement Findings** - Proceed with integration of research results
2. **Deep Dive Analysis** - Conduct deeper research on specific findings
3. **Export Results** - Export research data in preferred format
4. **Update Project** - Apply research findings to project components
5. **New Research** - Start new research on different topic
6. **Cancel** - End research session

### Wait for the user to answer

### Process user's answer:

**IF user selected "Implement Findings"**:
1. Execute integration procedures
2. Update relevant project components
3. Create integration commit
4. Report integration status

**IF user selected "Deep Dive Analysis"**:
1. Identify areas for deeper research
2. Execute focused research engines
3. Provide detailed analysis
4. Create specialized reports

**IF user selected "Export Results"**:
1. Export data in requested format
2. Create export files
3. Provide export summary
4. Share file locations

**IF user selected "Update Project"**:
1. Apply research findings to SPEC/TEST/CODE/DOC
2. Create appropriate TAGs
3. Generate commits
4. Update project status

**IF user selected "New Research"**:
1. Reset research context
2. Guide user to new research command
3. End current session

**IF user selected "Cancel"**:
1. Save research results for future reference
2. Create research summary
3. End session gracefully

---

## 📚 Reference Information

### Research Engine Capabilities

**Knowledge Integration Hub**:
- Multi-domain knowledge synthesis
- Pattern recognition across domains
- Knowledge graph construction
- Semantic analysis and integration

**Cross-Domain Analysis Engine**:
- Identify connections between different domains
- Transfer learning opportunities
- Cross-disciplinary pattern matching
- Interdisciplinary insights

**Pattern Recognition Engine**:
- Statistical pattern detection
- Behavioral pattern analysis
- Code pattern identification
- Solution pattern matching

**MCP Context7 Integration**:
- Latest documentation research
- API reference and examples
- Best practices and patterns
- Community knowledge integration

**MCP Playwright Integration**:
- Web interface analysis
- User experience research
- Performance testing
- Accessibility evaluation

**MCP Sequential Thinking Integration**:
- Complex reasoning chains
- Multi-step analysis
- Decision tree exploration
- Hypothesis testing

### TAG System Integration

**@RESEARCH Tags**:
- Research findings and insights
- Analysis results and conclusions
- Methodology and approaches
- Data sources and references

**@PATTERN Tags**:
- Identified patterns and trends
- Behavioral patterns
- Code patterns
- Solution patterns

**@SOLUTION Tags**:
- Practical solutions and implementations
- Best practices and recommendations
- Code examples and templates
- Implementation guides

### Research Quality Standards

**TRUST 5 Application**:
- **Test First**: Validate research findings with tests
- **Readable**: Clear documentation and communication
- **Unified**: Consistent integration with project standards
- **Secured**: Validate sources and security implications
- **Trackable**: Maintain research traceability and provenance

---

## 🎯 Summary: Your Execution Checklist

Before you consider this command complete, verify:

- [ ] **PHASE 1 executed**: Research intent analyzed and planned
- [ ] **User confirmation obtained**: User approved research plan
- [ ] **PHASE 2 executed**: Research engines initialized and data collected
- [ ] **Data processing complete**: Research data validated and normalized
- [ ] **PHASE 3 executed**: Research analysis and integration planning complete
- [ ] **TAGs assigned**: Appropriate @RESEARCH, @PATTERN, @SOLUTION tags created
- [ ] **PHASE 4 executed**: Research dashboard and documentation created
- [ ] **Integration complete**: Findings integrated into project (if requested)
- [ ] **Next steps presented**: User asked what to do next
- [ ] **Research data preserved**: Results saved for future reference

IF all checkboxes are checked → Command execution successful

IF any checkbox is unchecked → Identify missing step and complete it before ending

---

**End of command execution guide**