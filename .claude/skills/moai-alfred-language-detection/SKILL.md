---
name: moai-alfred-language-detection
version: 2.0.0
created: 2025-10-22
updated: 2025-11-11
status: active
description: "Auto-detects project language and framework from package.json, pyproject.toml, etc. Enhanced with research capabilities for detection algorithm optimization and language pattern analysis. Use when analyzing project stacks, optimizing tool selection, or researching language detection patterns."
keywords: ['language', 'detection', 'framework', 'auto', 'research', 'pattern-analysis', 'optimization', 'algorithm-development']
allowed-tools:
  - Read
  - Bash
  - AskUserQuestion
  - TodoWrite
---

# Alfred Language Detection Skill

## Skill Metadata

| Field | Value |
| ----- | ----- |
| **Skill Name** | moai-alfred-language-detection |
| **Version** | 2.0.0 (2025-10-22) |
| **Allowed tools** | Read (read_file), Bash (terminal) |
| **Auto-load** | On demand when keywords detected |
| **Tier** | Alfred |

---

## What It Does

Auto-detects project language and framework from package.json, pyproject.toml, etc.

**Key capabilities**:
- ✅ Best practices enforcement for alfred domain
- ✅ TRUST 5 principles integration
- ✅ Latest tool versions (2025-10-22)
- ✅ TDD workflow support

---

## When to Use

**Automatic triggers**:
- Related code discussions and file patterns
- SPEC implementation (`/alfred:2-run`)
- Code review requests

**Manual invocation**:
- Review code for TRUST 5 compliance
- Design new features
- Troubleshoot issues

---

## Inputs

- Language-specific source directories
- Configuration files
- Test suites and sample data

## Outputs

- Test/lint execution plan
- TRUST 5 review checkpoints
- Migration guidance

## Failure Modes

- When required tools are not installed
- When dependencies are missing
- When test coverage falls below 85%

## Dependencies

- Access to project files via Read/Bash tools
- Integration with `moai-foundation-langs` for language detection
- Integration with `moai-foundation-trust` for quality gates

---

## References (Latest Documentation)

_Documentation links updated 2025-10-22_

---

## Changelog

- **v2.0.0** (2025-10-22): Major update with latest tool versions, comprehensive best practices, TRUST 5 integration
- **v1.0.0** (2025-03-29): Initial Skill release

---

## Research Integration

### Language Detection Research Capabilities

**Detection Algorithm Research**:
- **Pattern recognition optimization**: Research on improving file pattern recognition for different languages and frameworks
- **Framework detection accuracy**: Analyze and improve detection accuracy for complex multi-framework projects
- **Configuration file analysis**: Research optimal methods for analyzing configuration files across different ecosystems
- **Language evolution tracking**: Study how language detection adapts to new framework versions and patterns

**Language Pattern Research**:
- **Multi-language project analysis**: Research detection patterns for polyglot projects and framework coexistence
- **Framework migration patterns**: Study language and framework evolution patterns in real-world projects
- **Toolchain optimization research**: Analyze the relationship between detected stacks and optimal tool selection
- **Community adoption patterns**: Research language and framework adoption trends and their impact on detection algorithms

**Research Methodology**:
- **Detection accuracy measurement**: Monitor detection success rates across different project types and complexities
- **False positive/negative analysis**: Study and reduce detection errors through pattern refinement
- **Performance optimization research**: Analyze detection speed and resource usage patterns
- **Algorithm effectiveness benchmarking**: Compare different detection approaches and their effectiveness

### Detection Research Framework

#### 1. Pattern Recognition Research
- **File pattern optimization**: Research on optimal file patterns for language and framework detection
- **Configuration analysis enhancement**: Study advanced configuration file parsing and analysis techniques
- **Dependency graph analysis**: Research dependency-based language detection methods
- **Machine learning integration**: Explore ML approaches for improved detection accuracy

#### 2. Multi-Language Project Research
- **Polyglot project analysis**: Research detection strategies for projects using multiple languages
- **Framework coexistence patterns**: Study how different frameworks coexist and affect detection
- **Language dominance identification**: Research methods for identifying primary languages in complex projects
- **Cross-framework compatibility**: Analyze compatibility patterns between different frameworks and tools

#### 3. Detection Optimization Research
```
Language Detection Research Framework:
├── Algorithm Development
│   ├── Pattern recognition enhancement
│   ├── Configuration file analysis
│   ├── Dependency graph research
│   └── Machine learning integration
├── Accuracy Analysis
│   ├── Detection success rate measurement
│   ├── False positive/negative reduction
│   ├── Multi-language project handling
│   └── Framework coexistence patterns
└── Performance Optimization
        ├── Detection speed optimization
        ├── Resource usage analysis
        ├── Algorithm efficiency benchmarking
        └── Scalability research
```

**Current Research Focus Areas**:
- Multi-language project detection optimization
- Framework coexistence pattern analysis
- Configuration file parsing enhancement
- Detection algorithm performance optimization
- Language evolution adaptation strategies

---

## Integration with Research System

The language detection system integrates with MoAI-ADK's research framework by:

1. **Collecting detection data**: Track detection accuracy, patterns, and performance across different project types
2. **Validating detection algorithms**: Provide real-world testing ground for new detection approaches and pattern recognition methods
3. **Documenting language patterns**: Capture successful detection patterns and share them across different ecosystems
4. **Benchmarking detection approaches**: Measure the effectiveness of different detection strategies and identify improvements

**Research Collaboration**:
- **Pattern recognition team**: Share data on file pattern effectiveness and optimization opportunities
- **Framework analysis team**: Provide insights on framework evolution and detection pattern adaptation
- **Tool optimization team**: Collaborate on tool selection optimization based on detected stacks
- **Performance research team**: Study detection algorithm performance and efficiency improvements

---

## Works Well With

- `moai-foundation-trust` (quality gates)
- `moai-alfred-code-reviewer` (code review)
- `moai-essentials-debug` (debugging support)
- `moai-foundation-langs` (language-specific guidance)

---

## Best Practices

✅ **DO**:
- Follow alfred best practices
- Use latest stable tool versions
- Maintain test coverage ≥85%
- Document all public APIs
- Validate detection results with user confirmation

❌ **DON'T**:
- Skip quality gates
- Use deprecated tools
- Ignore security warnings
- Mix testing frameworks
- Assume detection is always correct without validation
