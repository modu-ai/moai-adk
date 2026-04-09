# Claude Code Skills Best Practices Checklist

Comprehensive checklist for creating, validating, and maintaining Claude Code Skills that comply with official standards and deliver maximum value to users.

Purpose: Complete validation guide for skill quality and compliance
Target: Skill creators, maintainers, and reviewers
Last Updated: 2025-11-25
Version: 2.0.0

---

## Quick Reference (30 seconds)

Essential Validation: Official standards compliance + user value delivery. Key Areas: Frontmatter accuracy, content structure, code quality, integration patterns. Quality Gates: Technical validation + user experience testing + standards compliance.

---

## Pre-Creation Checklist

### Planning and Research

Problem Definition:
- [ ] Clearly identified specific problem or need
- [ ] Defined target user personas and use cases
- [ ] Researched existing skills to avoid duplication
- [ ] Scoped functionality to single responsibility

Requirements Analysis:
- [ ] Documented all trigger scenarios and use cases
- [ ] Identified required tools and permissions
- [ ] Planned integration with existing skills
- [ ] Defined success metrics and acceptance criteria

Standards Research:
- [ ] Reviewed latest Claude Code official documentation
- [ ] Understood current skill formatting standards
- [ ] Checked for recent changes in best practices
- [ ] Identified relevant examples and templates

### Technical Planning

Tool Selection:
- [ ] Applied principle of least privilege
- [ ] Selected minimal necessary tool set
- [ ] Considered MCP integration requirements
- [ ] Planned for security constraints

Architecture Design:
- [ ] Designed progressive disclosure structure
- [ ] Planned supporting file organization
- [ ] Considered performance and loading speed
- [ ] Designed for maintainability and updates

---

## Frontmatter Validation Checklist

### Required Fields Validation

name Field:
- [ ] Uses kebab-case format (lowercase, hyphens only)
- [ ] Maximum 64 characters in length
- [ ] Follows official naming convention (`[prefix]-[domain]-[function]`)
- [ ] No special characters other than hyphens and numbers
- [ ] Unique within the project/organization

description Field:
- [ ] Clearly describes what the skill does
- [ ] Includes specific trigger scenarios
- [ ] Maximum 1024 characters in length
- [ ] Avoids vague or generic language
- [ ] Includes context for when to use the skill

Optional Fields Validation:

allowed-tools (if present):
- [ ] Follows comma-separated format (no brackets)
- [ ] Uses minimal tool set required for functionality
- [ ] No deprecated or invalid tool names
- [ ] Considers security implications of each tool

version (if present):
- [ ] Follows semantic versioning (X.Y.Z)
- [ ] Incremented appropriately for changes
- [ ] Documented in changelog for major changes

tags (if present):
- [ ] Relevant to skill functionality
- [ ] Uses consistent categorization
- [ ] Facilitates skill discovery
- [ ] Follows organizational tag standards

updated (if present):
- [ ] Format: YYYY-MM-DD
- [ ] Reflects actual last modification date
- [ ] Updated with each content change

status (if present):
- [ ] Uses valid values: active, deprecated, experimental
- [ ] Accurately reflects skill state
- [ ] Provides migration guidance for deprecated skills

### YAML Syntax Validation

Structure Validation:
- [ ] Valid YAML syntax (no parsing errors)
- [ ] Proper indentation (2 spaces standard)
- [ ] No trailing whitespace or extra spaces
- [ ] Proper quoting for special characters

Content Validation:
- [ ] No forbidden characters or encoding issues
- [ ] Consistent quoting style
- [ ] Proper escaping of special characters
- [ ] No duplicate field names

---

## Content Structure Validation

### Required Sections

Quick Reference Section:
- [ ] Present and properly formatted (H2 heading)
- [ ] 2-4 sentences maximum for quick overview
- [ ] Focuses on core functionality and immediate value
- [ ] Uses clear, concise language
- [ ] Avoids technical jargon where possible

Implementation Guide Section:
- [ ] Present and properly formatted (H2 heading)
- [ ] Contains Core Capabilities subsection (H3)
- [ ] Contains When to Use subsection (H3)
- [ ] Contains Essential Patterns subsection (H3)
- [ ] Logical flow from simple to complex

Best Practices Section:
- [ ] Present and properly formatted (H2 heading)
- [ ] Uses DO format for positive recommendations
- [ ] Uses DON'T format for anti-patterns
- [ ] Each point includes clear rationale or explanation
- [ ] Covers security, performance, and maintainability

Works Well With Section (Optional but Recommended):
- [ ] Present if skill integrates with others
- [ ] Uses proper markdown link formatting
- [ ] Includes brief relationship description
- [ ] Links are valid and functional

### Content Quality Validation

Clarity and Specificity:
- [ ] Language is clear and unambiguous
- [ ] Examples are specific and actionable
- [ ] Technical terms are defined or explained
- [ ] No vague or generic descriptions

Technical Accuracy:
- [ ] Code examples are syntactically correct
- [ ] Technical details are current and accurate
- [ ] Examples follow language conventions
- [ ] Security considerations are appropriate

User Experience:
- [ ] Progressive disclosure structure (simple to complex)
- [ ] Examples are immediately usable
- [ ] Error conditions are documented
- [ ] Troubleshooting information is provided

---

## Code Example Validation

### Code Quality Standards

Syntax and Style:
- [ ] All code examples are syntactically correct
- [ ] Follow language-specific conventions and style guides
- [ ] Proper indentation and formatting
- [ ] Consistent coding style throughout examples

Documentation and Comments:
- [ ] Code includes appropriate comments and documentation
- [ ] Complex logic is explained
- [ ] Function and variable names are descriptive
- [ ] Docstrings follow language conventions

Error Handling:
- [ ] Examples include proper error handling where applicable
- [ ] Edge cases are considered and documented
- [ ] Exception handling follows best practices
- [ ] Resource cleanup is demonstrated

Security Considerations:
- [ ] No hardcoded credentials or sensitive data
- [ ] Examples follow security best practices
- [ ] Input validation is demonstrated
- [ ] Appropriate permission levels are shown

### Multi-language Support

Language Identification:
- [ ] All code blocks include language identifiers
- [ ] Examples cover relevant programming languages
- [ ] Language-specific conventions are followed
- [ ] Cross-language compatibility is considered

Integration Examples:
- [ ] Examples show how to integrate with other tools/services
- [ ] API integration patterns are demonstrated
- [ ] Configuration examples are provided
- [ ] Testing and validation approaches are shown

---

## Integration and Compatibility

### Skill Integration

Works Well With Section:
- [ ] Identifies complementary skills
- [ ] Describes integration patterns
- [ ] Links are valid and functional
- [ ] Integration examples are provided

MCP Integration (if applicable):
- [ ] MCP tools properly declared in allowed-tools
- [ ] Two-step Context7 pattern used where appropriate
- [ ] Proper error handling for MCP calls
- [ ] Fallback strategies are documented

Tool Dependencies:
- [ ] All required tools are properly declared
- [ ] Optional dependencies are documented
- [ ] Version requirements are specified
- [ ] Installation instructions are provided

### Compatibility Validation

Claude Code Version:
- [ ] Compatible with current Claude Code version
- [ ] No deprecated features or APIs used
- [ ] Future compatibility considered
- [ ] Migration plans for breaking changes

Platform Compatibility:
- [ ] Works across different operating systems
- [ ] Browser compatibility considered for web-related skills
- [ ] Cross-platform dependencies handled
- [ ] Platform-specific limitations documented

---

## Performance and Scalability

### Performance Considerations

Token Usage Optimization:
- [ ] SKILL.md under 500 lines (strict requirement)
- [ ] Progressive disclosure implemented effectively
- [ ] Large content moved to supporting files
- [ ] Cache-friendly structure implemented

Loading Speed:
- [ ] Supporting files organized for efficient loading
- [ ] Internal links use relative paths
- [ ] No circular references or deep nesting
- [ ] File sizes are reasonable for quick loading

Resource Management:
- [ ] Minimal external dependencies
- [ ] Efficient file organization
- [ ] Appropriate use of caching strategies
- [ ] Memory-efficient patterns demonstrated

### Scalability Design

Maintainability:
- [ ] Modular structure for easy updates
- [ ] Clear separation of concerns
- [ ] Consistent patterns and conventions
- [ ] Documentation for future maintainers

Extensibility:
- [ ] Extension points identified and documented
- [ ] Plugin architecture considered if applicable
- [ ] Version compatibility maintained
- [ ] Backward compatibility preserved where possible

---

## Security and Privacy

### Security Validation

Tool Permissions:
- [ ] Principle of least privilege applied
- [ ] No unnecessary permissions granted
- [ ] Security implications documented
- [ ] Safe defaults provided

Data Handling:
- [ ] No sensitive data in examples or comments
- [ ] Proper data sanitization demonstrated
- [ ] Privacy considerations adddessed
- [ ] Secure data storage patterns shown

Input Validation:
- [ ] Input validation demonstrated where applicable
- [ ] Sanitization patterns are included
- [ ] Edge cases and boundary conditions considered
- [ ] Error handling prevents information disclosure

### Compliance Standards

OWASP Compliance (if applicable):
- [ ] Security best practices followed
- [ ] Common vulnerabilities adddessed
- [ ] Security headers and configurations shown
- [ ] Secure coding practices demonstrated

Industry Standards:
- [ ] Industry-specific regulations considered
- [ ] Compliance requirements documented
- [ ] Audit trails demonstrated where applicable
- [ ] Documentation meets organizational standards

---

## Documentation Quality

### Content Organization

Logical Structure:
- [ ] Content flows logically from simple to complex
- [ ] Sections are clearly defined and labeled
- [ ] Navigation between sections is intuitive
- [ ] Information architecture supports different user needs

Writing Quality:
- [ ] Language is clear and concise
- [ ] Technical writing standards followed
- [ ] Consistent terminology throughout
- [ ] Grammar and spelling are correct

User Experience:
- [ ] Learning curve is appropriate for target audience
- [ ] Examples are immediately actionable
- [ ] Troubleshooting information is comprehensive
- [ ] Help and support resources are identified

### Visual Formatting

Markdown Standards:
- [ ] Proper heading hierarchy (H1 → H2 → H3)
- [ ] Consistent use of emphasis and formatting
- [ ] Code blocks use proper syntax highlighting
- [ ] Lists and tables are properly formatted

Accessibility:
- [ ] Content is accessible to screen readers
- [ ] Color contrast meets accessibility standards
- [ ] Alternative text provided for images
- [ ] Structure supports navigation without visual formatting

---

## Testing and Validation

### Functional Testing

Example Validation:
- [ ] All code examples tested and verified working
- [ ] Test cases cover main functionality
- [ ] Edge cases are tested and documented
- [ ] Integration examples are tested in context

Cross-platform Testing:
- [ ] Examples work on different operating systems
- [ ] Browser compatibility verified for web-related skills
- [ ] Version compatibility tested
- [ ] Environment-specific variations documented

### Quality Assurance

Automated Validation:
- [ ] YAML syntax validation automated
- [ ] Link checking automated
- [ ] Code linting and formatting validation
- [ ] Performance metrics monitored

Manual Review:
- [ ] Content reviewed by subject matter experts
- [ ] User experience tested with target audience
- [ ] Peer review process completed
- [ ] Documentation accuracy verified

---

## Publication and Deployment

### File Structure Validation

Required Structure:
```
skill-name/
 SKILL.md (REQUIRED, ≤500 lines)
 reference.md (OPTIONAL)
 examples.md (OPTIONAL)
 scripts/ (OPTIONAL)
 helper.sh
 templates/ (OPTIONAL)
 template.md
```

File Naming:
- [ ] Directory name matches skill name (kebab-case)
- [ ] SKILL.md is uppercase (required)
- [ ] Supporting files follow naming conventions
- [ ] No prohibited characters in file names

Content Distribution:
- [ ] Core content in SKILL.md (≤500 lines)
- [ ] Additional documentation in reference.md
- [ ] Extended examples in examples.md
- [ ] Utility scripts in scripts/ directory

### Version Control

Semantic Versioning:
- [ ] Version follows X.Y.Z format
- [ ] Major version indicates breaking changes
- [ ] Minor version indicates new features
- [ ] Patch version indicates bug fixes

Change Documentation:
- [ ] Changelog maintained with version history
- [ ] Breaking changes clearly documented
- [ ] Migration paths provided for major changes
- [ ] Deprecation notices with timelines

Release Process:
- [ ] Pre-release validation completed
- [ ] Release notes prepared
- [ ] Version tags properly applied
- [ ] Distribution channels updated

---

## Post-Publication Monitoring

### Success Metrics

Usage Analytics:
- [ ] Skill loading and usage tracked
- [ ] User feedback collected and analyzed
- [ ] Performance metrics monitored
- [ ] Error rates tracked and adddessed

Quality Indicators:
- [ ] User satisfaction measured
- [ ] Support requests analyzed
- [ ] Community adoption tracked
- [ ] Documentation quality assessed

### Maintenance Planning

Regular Updates:
- [ ] Update schedule established
- [ ] Deprecation timeline planned
- [ ] Succession planning for maintainers
- [ ] Community contribution process defined

Continuous Improvement:
- [ ] User feedback incorporation process
- [ ] Performance optimization ongoing
- [ ] Standards compliance monitoring
- [ ] Technology trends monitoring

---

## Comprehensive Validation Checklist

### Final Validation Gates

Technical Compliance:
- [ ] All YAML frontmatter fields are correct and complete
- [ ] Content structure follows official standards
- [ ] Code examples are tested and functional
- [ ] File organization is optimal

Quality Standards:
- [ ] Content is clear, specific, and actionable
- [ ] Examples demonstrate best practices
- [ ] Security considerations are adddessed
- [ ] Performance optimization is implemented

User Experience:
- [ ] Learning curve is appropriate for target audience
- [ ] Documentation supports different use cases
- [ ] Troubleshooting information is comprehensive
- [ ] Integration patterns are clear

Standards Compliance:
- [ ] Official Claude Code standards followed
- [ ] Organization guidelines met
- [ ] Industry best practices implemented
- [ ] Accessibility standards met

### Publication Approval Criteria

Ready for Publication:
- [ ] All required sections present and complete
- [ ] Technical validation passed with no critical issues
- [ ] Quality standards met with high confidence
- [ ] User testing shows positive results

Conditional Publication:
- [ ] Minor issues identified but don't block publication
- [ ] Improvements planned for next version
- [ ] Monitoring and feedback collection established
- [ ] Update timeline defined

Not Ready for Publication:
- [ ] Critical issues blocking functionality
- [ ] Major standards compliance violations
- [ ] Incomplete or missing required sections
- [ ] User testing shows significant problems

---

## Troubleshooting Common Issues

### Validation Failures

YAML Parsing Errors:
```yaml
# Common issue: Invalid array format
# WRONG
allowed-tools: [Read, Write, Bash]

# CORRECT
allowed-tools: Read, Write, Bash
```

Line Count Exceeded:
- Move detailed examples to examples.md
- Transfer advanced patterns to reference.md
- Consolidate related content
- Use progressive disclosure effectively

Link Validation Failures:
- Check relative path formats
- Verify target files exist
- Update broken external links
- Test all internal navigation

### Quality Improvement

Content Clarity Issues:
- Add specific examples for abstract concepts
- Define technical terms and jargon
- Include context and rationale for recommendations
- Use consistent terminology throughout

User Experience Problems:
- Simplify complex explanations
- Add more step-by-step examples
- Improve navigation and organization
- Enhance troubleshooting section

---

## Example Validation Process

### Step-by-Step Validation

1. Automated Checks:
```bash
# YAML syntax validation
yamllint .claude/skills/skill-name/SKILL.md

# Link checking
markdown-link-check .claude/skills/skill-name/

# Line count verification
wc -l .claude/skills/skill-name/SKILL.md
```

2. Manual Review:
- Read through entire skill content
- Test all code examples
- Verify all links and references
- Assess user experience and flow

3. User Testing:
- Have target users test the skill
- Collect feedback on clarity and usefulness
- Validate examples work in real scenarios
- Assess learning curve and documentation

4. Final Validation:
- Complete comprehensive checklist
- Adddess any identified issues
- Document any known limitations
- Prepare for publication

---

Version: 2.0.0
Compliance: Claude Code Official Standards
Last Updated: 2025-11-25
Checklist Items: 200+ validation points
Quality Gates: Technical + User Experience + Standards

Generated with Claude Code using official documentation and best practices.
