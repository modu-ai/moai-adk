# Alfred Reporting Reference

## API Specification

### Core Functions

#### `generate_report(report_type, data, format_type) -> Report`
**Description**: Generate standardized report based on type and data

**Parameters**:
- `report_type` (str): Type of report (completion, analysis, sync, etc.)
- `data` (dict): Report data and metrics
- `format_type` (str): "plain_text" or "markdown"

**Returns**: `Report` object with formatted content

**Report Types**:
```python
REPORT_TYPES = {
    "task_completion": {
        "sections": ["achievements", "metrics", "next_steps"],
        "format": "structured"
    },
    "analysis": {
        "sections": ["findings", "impact", "recommendations"],
        "format": "detailed"
    },
    "sync": {
        "sections": ["changes", "validation", "quality"],
        "format": "metrics_focused"
    },
    "error": {
        "sections": ["problem", "cause", "solution"],
        "format": "problem_solution"
    }
}
```

#### `format_output(content, output_type) -> str`
**Description**: Format content according to output type standards

**Parameters**:
- `content` (str): Raw content to format
- `output_type` (str): "screen" or "document"

**Returns**: `str` - Formatted content

**Output Type Rules**:
```python
SCREEN_FORMAT_RULES = {
    "no_markdown": True,
    "no_tables": True,
    "no_emoji_formatting": True,
    "plain_text_hierarchy": True,
    "indentation_based": True
}

DOCUMENT_FORMAT_RULES = {
    "markdown_required": True,
    "structured_sections": True,
    "tables_allowed": True,
    "emoji_status_indicators": True,
    "proper_headings": True
}
```

### Format Specifications

#### Plain Text Format (Screen Output)

**Structure**:
```
[SECTION TITLE]

[Content with indentation]
- Bullet point with dash
- Another bullet point

[Subsection]
[Indented content]
```

**Rules**:
- No markdown syntax (`#`, `##`, `**`, `*`)
- Use simple text hierarchy
- Indentation for structure (2 spaces per level)
- Dashes (`-`) for bullet points
- Clear line breaks between sections

#### Markdown Format (Documents)

**Structure**:
```markdown
## 🎯 Section Title

Content with proper markdown formatting.

### Subsection Title

- ✅ Status item with emoji
- ❌ Another status item

| Column 1 | Column 2 | Status |
|----------|----------|--------|
| Value 1 | Value 2 | ✅ |
```

**Rules**:
- Proper markdown headings (`##`, `###`)
- Tables for structured data
- Emojis for visual status indicators
- Bold/italic for emphasis where needed
- Consistent formatting throughout

### Document Location Rules

#### Allowed Document Locations

```python
DOCUMENT_LOCATIONS = {
    "internal_guides": {
        "directory": ".moai/docs/",
        "pattern": "implementation-{spec}.md"
    },
    "exploration_reports": {
        "directory": ".moai/docs/", 
        "pattern": "exploration-{topic}.md"
    },
    "spec_documents": {
        "directory": ".moai/specs/SPEC-*/",
        "pattern": "spec.md, plan.md, acceptance.md"
    },
    "sync_reports": {
        "directory": ".moai/reports/",
        "pattern": "sync-report-{date}.md"
    },
    "technical_analysis": {
        "directory": ".moai/analysis/",
        "pattern": "{topic}-analysis.md"
    }
}
```

#### Forbidden Root Directory Files

**NEVER create in project root**:
```
- ❌ IMPLEMENTATION_GUIDE.md
- ❌ EXPLORATION_REPORT.md  
- ❌ *_ANALYSIS.md
- ❌ *_GUIDE.md
- ❌ *_REPORT.md
```

**Allowed root files only**:
```
- ✅ README.md
- ✅ CHANGELOG.md
- ✅ CONTRIBUTING.md
- ✅ LICENSE
```

## Template System

### Report Templates

#### Task Completion Template
```python
TASK_COMPLETION_TEMPLATE = """
## 🎊 {task_type} Complete

### {results_section_title}
{results_list}

### {metrics_section_title}
{metrics_table}

### {next_steps_section_title}
{next_steps_list}
"""
```

#### Analysis Template
```python
ANALYSIS_TEMPLATE = """
## 🔍 {analysis_type} Analysis

### Key Findings
{findings_list}

### Impact Assessment
{impact_description}

### Recommendations
{recommendations_list}

### Next Steps
{action_items}
"""
```

### Sub-agent Report Templates

#### spec-builder Template Variables
```python
SPEC_BUILDER_TEMPLATE = {
    "title": "📋 SPEC Creation Complete",
    "sections": {
        "generated_documents": [
            "spec.md - Requirements specification",
            "plan.md - Implementation plan", 
            "acceptance.md - Acceptance criteria"
        ],
        "validation_results": [
            "EARS format compliance",
            "@TAG chain creation",
            "Measurability verification"
        ]
    },
    "metrics_table": {
        "headers": ["Aspect", "Status", "Details"],
        "rows": [
            ["Clarity", "✅", "Requirements unambiguous"],
            ["Completeness", "✅", "All needs captured"],
            ["Testability", "✅", "Each requirement testable"]
        ]
    }
}
```

#### tdd-implementer Template Variables
```python
TDD_IMPLEMENTER_TEMPLATE = {
    "title": "🚀 TDD Implementation Complete",
    "sections": {
        "implementation_files": [
            "src/feature.py - Core implementation",
            "tests/test_feature.py - Test suite"
        ],
        "tdd_phases": [
            "RED - Failure confirmed",
            "GREEN - Implementation successful", 
            "REFACTOR - Code optimized"
        ]
    },
    "quality_metrics": {
        "test_coverage": "95%",
        "code_quality": "0 issues",
        "performance": "+15% improvement"
    }
}
```

## Quality Standards

### Content Requirements

**Every report must contain**:
1. **Clear Title**: Descriptive, searchable
2. **Executive Summary**: Key findings (3-4 bullet points)
3. **Detailed Analysis**: Structured information
4. **Quantifiable Metrics**: Data where applicable
5. **Actionable Next Steps**: 2-5 specific items

**Formatting Requirements**:
1. **Consistent Structure**: Use standardized templates
2. **Visual Organization**: Emojis, tables, lists
3. **Language Consistency**: User's language for explanations
4. **Technical Accuracy**: Verify all details

### Quality Validation

```python
def validate_report(report: Report) -> ValidationResult:
    """Validate report against quality standards"""
    
    errors = []
    warnings = []
    
    # Content validation
    if not report.title:
        errors.append("Report title is required")
    
    if not report.summary:
        warnings.append("Executive summary recommended")
    
    if not report.next_steps:
        errors.append("Next steps are required")
    
    # Formatting validation
    if report.format_type == "markdown":
        if not has_proper_headings(report.content):
            errors.append("Markdown reports require proper headings")
        
        if not has_status_indicators(report.content):
            warnings.append("Status indicators (✅❌⚠️) recommended")
    
    elif report.format_type == "plain_text":
        if has_markdown_syntax(report.content):
            errors.append("Plain text reports must not contain markdown")
    
    # Location validation
    if not is_valid_location(report.file_path):
        errors.append(f"Invalid document location: {report.file_path}")
    
    return ValidationResult(
        valid=len(errors) == 0,
        errors=errors,
        warnings=warnings
    )
```

## Integration Points

### Alfred Workflow Integration

```python
WORKFLOW_REPORT_TRIGGERS = {
    "alfred:0-project": {
        "trigger": "command_completion",
        "report_type": "project_setup",
        "location": ".moai/docs/",
        "template": "project_completion"
    },
    "alfred:1-plan": {
        "trigger": "command_completion", 
        "report_type": "spec_creation",
        "location": ".moai/docs/",
        "template": "spec_completion"
    },
    "alfred:2-run": {
        "trigger": "command_completion",
        "report_type": "implementation",
        "location": ".moai/docs/",
        "template": "implementation_completion"
    },
    "alfred:3-sync": {
        "trigger": "command_completion",
        "report_type": "synchronization", 
        "location": ".moai/reports/",
        "template": "sync_completion"
    }
}
```

### Sub-agent Integration

```python
SUBAGENT_REPORT_MAPPING = {
    "spec-builder": {
        "report_type": "spec_creation",
        "template": "spec_builder_template",
        "location": ".moai/docs/",
        "triggers": ["task_completion"]
    },
    "tdd-implementer": {
        "report_type": "implementation", 
        "template": "tdd_implementer_template",
        "location": ".moai/docs/",
        "triggers": ["task_completion", "quality_verification"]
    },
    "doc-syncer": {
        "report_type": "synchronization",
        "template": "doc_syncer_template", 
        "location": ".moai/reports/",
        "triggers": ["task_completion", "tag_verification"]
    }
}
```

## Error Handling

### Common Formatting Errors

```python
def fix_formatting_errors(content: str, format_type: str) -> str:
    """Automatically fix common formatting errors"""
    
    if format_type == "plain_text":
        # Remove markdown syntax
        content = re.sub(r'#{1,6}\s*', '', content)  # Remove headers
        content = re.sub(r'\*\*(.*?)\*\*', r'\1', content)  # Remove bold
        content = re.sub(r'\*(.*?)\*', r'\1', content)  # Remove italic
        content = re.sub(r'\|.*\|', '', content)  # Remove tables
        
    elif format_type == "markdown":
        # Ensure proper markdown structure
        content = ensure_proper_headings(content)
        content = fix_table_formatting(content)
        content = add_status_indicators(content)
    
    return content.strip()
```

### Location Validation

```python
def validate_document_location(file_path: str) -> bool:
    """Validate document follows location rules"""
    
    # Forbidden root files
    forbidden_patterns = [
        r'^IMPLEMENTATION_.*\.md$',
        r'^.*_ANALYSIS\.md$',
        r'^.*_GUIDE\.md$',
        r'^.*_REPORT\.md$'
    ]
    
    # Check if in project root
    if '/' not in file_path:
        for pattern in forbidden_patterns:
            if re.match(pattern, file_path):
                return False
        
        # Allow only specific root files
        allowed_root_files = ['README.md', 'CHANGELOG.md', 'CONTRIBUTING.md', 'LICENSE']
        return file_path in allowed_root_files
    
    # Check if in allowed .moai/ locations
    if file_path.startswith('.moai/'):
        allowed_dirs = ['.moai/docs/', '.moai/reports/', '.moai/analysis/', '.moai/specs/']
        return any(file_path.startswith(allowed_dir) for allowed_dir in allowed_dirs)
    
    return False
```

## Performance Considerations

### Report Generation Efficiency

```python
class ReportGenerator:
    """Efficient report generation with caching"""
    
    def __init__(self):
        self.template_cache = {}
        self.validation_cache = {}
    
    def generate_report(self, report_type: str, data: dict) -> Report:
        """Generate report with template caching"""
        
        # Use cached template if available
        template = self.template_cache.get(report_type)
        if not template:
            template = self.load_template(report_type)
            self.template_cache[report_type] = template
        
        # Generate report
        content = template.render(data)
        
        # Validate with caching
        cache_key = f"{report_type}_{hash(content)}"
        validation = self.validation_cache.get(cache_key)
        if not validation:
            validation = validate_report_content(content, report_type)
            self.validation_cache[cache_key] = validation
        
        return Report(
            content=content,
            validation=validation,
            metadata=self.generate_metadata(report_type, data)
        )
```

### Memory Management

- **Template caching**: Cache frequently used templates in memory
- **Validation caching**: Cache validation results for identical content
- **Lazy loading**: Load templates only when needed
- **Cleanup**: Remove old cache entries periodically

## Troubleshooting

### Common Issues

**Issue**: Report contains markdown when plain text expected
**Solution**: Use `format_output(content, "screen")` for user responses

**Issue**: Document saved in wrong location
**Solution**: Use `validate_document_location()` before saving

**Issue**: Report missing required sections
**Solution**: Use appropriate template from `REPORT_TEMPLATES`

**Issue**: Quality validation fails
**Solution**: Check `validate_report()` output for specific issues

### Debug Commands

```python
# Debug report generation
debug_report_generation(report_type, data)

# Validate document location
validate_document_location(file_path)

# Test format conversion
test_format_conversion(content, from_format, to_format)

# Check template availability
list_available_templates()
```

### Configuration Validation

```python
def validate_reporting_config(config: dict) -> ValidationResult:
    """Validate reporting configuration"""
    
    errors = []
    
    # Check required locations exist
    required_dirs = ['.moai/docs/', '.moai/reports/', '.moai/analysis/']
    for dir_path in required_dirs:
        if not os.path.exists(dir_path):
            errors.append(f"Required directory missing: {dir_path}")
    
    # Check template availability
    required_templates = ['task_completion', 'analysis', 'sync', 'error']
    for template in required_templates:
        if template not in config['available_templates']:
            errors.append(f"Required template missing: {template}")
    
    return ValidationResult(valid=len(errors) == 0, errors=errors)
```
