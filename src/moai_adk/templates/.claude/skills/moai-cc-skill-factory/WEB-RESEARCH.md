# Web Research Strategy for Skill Creation

This guide covers how skill-factory uses **WebSearch** and **WebFetch** to gather latest information, official documentation, and best practices for creating current, accurate Skills.

---

## Why Web Research?

### The Problem: Outdated Information

```
❌ Static Skill (created 6 months ago):
   ├─ Framework version: 7.2
   ├─ Best practice: Use approach X
   ├─ Deprecated feature: Still recommended in examples
   └─ Result: Users follow outdated guidance
```

### The Solution: Research-Driven Skill Generation

```
✅ Research-backed Skill (created today):
   ├─ Framework version: 8.0 (current as of 2025-10-22)
   ├─ Best practice: Approach Y (official docs say so)
   ├─ Deprecated feature: Flagged as "legacy"
   ├─ Official resources: Linked for further reading
   └─ Result: Users follow proven, current patterns
```

---

## Research Query Construction

### Query Template

```
[Framework] [Version] [Topic] [Best Practice Keyword] [Year]

Examples:
- "Pytest 8.0 async testing best practices 2025"
- "TypeScript 5.3 strict typing patterns official guide"
- "Go 1.22 error handling idioms 2025"
- "React 19 hooks migration guide 2025"
```

### Search Query Prioritization

**Tier 1: Version-Specific + Official Keywords**
```
Query: "Python 3.12 testing unittest pytest official 2025"
Expect: Official Python docs, pytest releases, best practices
```

**Tier 2: Framework + Pattern + Docs**
```
Query: "TypeScript 5 strict compiler options official"
Expect: TypeScript handbook, release notes, examples
```

**Tier 3: Fallback to Broader Search**
```
Query: "Go error handling patterns best practices"
Expect: Blog posts, tutorials, Stack Overflow consensus
```

### Search Domain Filtering

```
Tier 1 (Authoritative - 60% weight):
├─ docs.python.org
├─ nodejs.org/docs
├─ rust-lang.org
├─ golang.org
├─ official-project.org
├─ RFC documents (tools.ietf.org)
└─ GitHub official repositories

Tier 2 (Reputable - 30% weight):
├─ MDN Web Docs (mozilla.org)
├─ Dev.to (technical articles)
├─ Medium engineering blogs
├─ Reputable YouTube channels
└─ Academic papers (arxiv.org)

Tier 3 (Supporting - 10% weight):
├─ Stack Overflow
├─ Blogs by known experts
├─ Tutorial sites
└─ Community forums
```

---

## Information Extraction & Validation

### Step 1: Comprehensive Search Execution

For each Skill requirement from Interactive Discovery:

```
User Requirements:
├─ Domain: Testing
├─ Tech: Python + Pytest
├─ Features: Fixtures, mocking, async testing
├─ Version: Latest stable
└─ Audience: Intermediate

↓ Generate search queries:

Query 1: "Pytest 8.0 fixtures best practices official 2025"
Query 2: "Python unittest.mock mocking patterns 2025"
Query 3: "Pytest async testing pytest-asyncio 2025"
Query 4: "Pytest breaking changes migration guide 8.0"
Query 5: "Python 3.12 testing performance profiling 2025"
```

### Step 2: Source Validation Checklist

For each search result, validate:

```
□ Source authority level (Tier 1/2/3)
□ Publication date (< 6 months for fast domains, < 2 years for stable)
□ Information freshness (mentions latest version?)
□ Conflicts with official docs? (None should contradict official)
□ Security implications mentioned? (if applicable)
□ Deprecation warnings present? (historical context important)
```

### Step 3: Information Categorization

Organize findings by category:

```
Latest Version Information:
├─ Current stable: Pytest 8.0.0
├─ Release date: 2025-10
├─ Previous version: 7.4.3 (EOL date: 2025-12)
└─ LTS variant: N/A

Official Recommendations:
├─ Test framework: pytest (official)
├─ Fixture pattern: Use @pytest.fixture decorator
├─ Async support: Use pytest.mark.asyncio
└─ Type checking: Use --strict-markers flag

Deprecated Features:
├─ Old syntax: class-based fixtures (deprecated)
├─ Legacy pattern: nose integration (removed in 8.0)
├─ Outdated tool: tox replacement (pytest-xdist recommended)
└─ Security: Removed unsafe plugin loading

Emerging Patterns:
├─ Type-hinted fixtures (new in 8.0)
├─ Async fixture generators
├─ Plugin discovery improvements
└─ Performance optimizations

Breaking Changes:
├─ Plugin API changes (5-7 plugins affected)
├─ Deprecation warnings now errors
├─ Some legacy fixtures removed
└─ Migration path: Documented in official guide

Common Pitfalls:
├─ Fixture scope confusion (session vs module vs function)
├─ Mocking external services (use monkeypatch or mock library)
├─ Async test timeout configuration
└─ Plugin conflict resolution
```

### Step 4: Authority & Accuracy Scoring

```
Score each finding:

High Confidence (9-10/10):
├─ Official docs (python.org, pytest.org, etc.)
├─ Official blog / release notes
├─ RFC / Standards documents
└─ Official examples in source code

Medium Confidence (6-8/10):
├─ Major tech publication (MDN, Dev.to)
├─ Well-maintained GitHub repo with stars
├─ Book by recognized expert
└─ Podcast from authority

Lower Confidence (<6/10):
├─ Blog post (single source)
├─ Tutorial site (may be outdated)
├─ Forum discussion (not reviewed)
└─ Anecdotal evidence

Usage Rule:
- Use only High & Medium confidence for Skill content
- Cite sources for transparency
- Flag if High confidence is unavailable
```

---

## Integration with Skill Content

### Mapping Research → Skill Sections

```
Research Finding → SKILL.md Section → reference.md → examples.md

Latest Version:
├─ SKILL.md: "Using Pytest 8.0+"
├─ reference.md: Version compatibility table
└─ examples.md: Updated code samples

Official Best Practice:
├─ SKILL.md: Recommended pattern
├─ reference.md: API details + link to official docs
└─ examples.md: Working example with explanation

Breaking Change:
├─ SKILL.md: Migration section (if major)
├─ reference.md: Detailed breaking change list
└─ examples.md: Before/after migration example

Deprecated Feature:
├─ SKILL.md: ❌ Avoid this approach
├─ reference.md: Why it's deprecated + replacement
└─ examples.md: Updated pattern

Common Pitfall:
├─ SKILL.md: Anti-pattern warning
├─ reference.md: Why it's problematic
└─ examples.md: Correct approach
```

---

## Official Documentation Integration

### Documentation Quality Hierarchy

```
Tier 1: Direct Official Resources
├─ docs.official-project.org
├─ GitHub official repository README
├─ Official blog / release notes
└─ Email from project maintainers

Tier 2: Derived Official Content
├─ Official examples in repo
├─ Official tutorials
├─ Official conference talks (YouTube)
└─ Official Discord/Slack community

Tier 3: Community Translation
├─ Well-maintained community docs
├─ Translated docs (official translation)
├─ Community translation efforts
└─ Multiple sources agreeing on interpretation
```

### Citation Standards

For every claim in SKILL.md, cite source:

```
✅ Good:
"According to the official Pytest documentation [1],
fixtures should use @pytest.fixture decorator.

[1] https://docs.pytest.org/fixtures/index.html
    (Accessed 2025-10-22)"

❌ Bad:
"You should use fixtures for testing"
(No source, no link, no date)
```

---

## Conflict Resolution

### When Official Sources Disagree

```
Situation:
├─ Official docs recommend Pattern A
├─ Official blog recommends Pattern B
└─ They conflict!

Resolution:
1. Check publication date (newer is more reliable)
2. Check version specificity (more specific wins)
3. Look for explicitly deprecated notice
4. Present BOTH with explanation:

   "Official docs (2025-10) recommend Pattern A.
    However, the migration guide (2025-09) suggests
    Pattern B for legacy code. Use Pattern A for new code."
```

### When Community Disagrees with Official

```
Situation:
├─ Official docs say "Never do X"
├─ 90% of Stack Overflow does X
└─ Users will be confused

Resolution:
1. Prioritize official guidance (100%)
2. Document the discrepancy
3. Explain why official recommends differently

   "Official docs recommend approach A (for [reason]).
    You may see approach B in legacy code or tutorials.
    We recommend A for [reason], but understand both."
```

---

## Deprecation & Security Research

### Deprecation Detection

Search specifically for deprecation information:

```
Query templates:
- "[Framework] [version] deprecated features"
- "[Framework] breaking changes migration"
- "[Feature] deprecation timeline"
- "[Framework] EOL end of life support"

Action: Flag each deprecated feature
Result: Include deprecation warnings in SKILL.md

Example finding:
  "nose testing framework deprecated, replaced by pytest"
  Action: Don't mention nose in examples
  Action: If mentioning legacy patterns, warn about nose
```

### Security Research

Search for security best practices:

```
Query templates:
- "[Framework] security best practices 2025"
- "[Feature] security vulnerability"
- "[Framework] OWASP compliance"
- "[Library] security audit results"

Action: Include security warnings if relevant
Result: Examples follow security guidelines

Example finding:
  "pickle module has security implications for untrusted data"
  Action: Add warning in serialization examples
  Action: Recommend alternative approaches
```

---

## Verification Workflow

### Pre-Publication Verification

Before finalizing Skill:

```
Verification Checklist:

□ All versions stated match current release date
□ All examples tested against stated version
□ All "official recommendation" claims have official source
□ All deprecations verified in official docs
□ All links are working (WebFetch each link)
□ No security vulnerabilities in example code
□ No outdated patterns presented as current
□ Conflicting advice explained and contextualized
□ Breaking changes prominently noted
□ Migration paths provided for deprecated features
```

### Currency Dating

Add generation metadata to Skill:

```
# Skill Metadata

**Created**: 2025-10-22
**Framework Versions**: Pytest 8.0.0 (current as of creation date)
**Based on Official Docs**: https://docs.pytest.org (2025-10-20)
**Update Recommended**: 2025-11-22 (quarterly check)
**Security Review Date**: 2025-10-22
**Deprecation Review**: No deprecated patterns found as of 2025-10-22
```

---

## Handling Edge Cases

### Framework in Rapid Development

```
Situation: Framework has weekly updates

Solution:
1. Target stable/LTS version, not bleeding edge
2. Note development frequency
3. Link to changelog
4. Recommend users check for updates

Example: "React releases new minor versions weekly.
This Skill targets React 19 LTS. Check react.dev/blog
for latest updates after reading this."
```

### Framework with Multiple Implementation Options

```
Situation: Testing framework available for multiple languages

Solution:
1. Clarify which implementation in title
2. Link to each implementation's official docs
3. Include examples for the specified implementation
4. Mention alternatives but keep focus narrow

Example:
"# Testing with Jest (JavaScript/TypeScript implementation)

Note: Jest is primarily for JavaScript/TypeScript.
For Python, see pytest (moai-skill-testing-python).
For Go, see testing package (moai-skill-testing-go)."
```

### Niche/Emerging Framework

```
Situation: Framework is 1-2 years old, limited documentation

Solution:
1. Acknowledge as emerging
2. Prioritize official resources (may be on GitHub)
3. Include community resources with caveats
4. Recommend checking official repo for updates
5. Suggest users check issues/discussions for latest

Example: "This framework is actively developed.
Check the official GitHub repository regularly
for breaking changes and new features."
```

---

## Search Workflow Automation

### Typical Search Sequence

```
Phase 1: Version & Release Info
├─ Query: "[Framework] latest version 2025"
├─ Goal: Find current version number
└─ Result: Exact version string (e.g., "8.0.0")

Phase 2: Official Best Practices
├─ Query: "[Framework] [version] best practices official guide"
├─ Goal: Official recommendations
└─ Result: Canonical patterns to recommend

Phase 3: Breaking Changes & Migration
├─ Query: "[Framework] [current] vs [previous] breaking changes"
├─ Goal: Upgrade path for users on older version
└─ Result: Migration steps and deprecation timeline

Phase 4: Security Audit
├─ Query: "[Framework] [version] security vulnerabilities"
├─ Goal: Identify security considerations
└─ Result: Security warnings (if any)

Phase 5: Real-World Examples
├─ Query: "[Framework] [version] example projects GitHub stars"
├─ Goal: Working code examples
└─ Result: Reference implementations
```

---

## Continuous Maintenance

### Skill Update Cycle

```
Created: 2025-10-22
├─ Next Review: 2025-11-22 (1 month for fast-moving)
├─ Quarterly Review: 2026-01-22
└─ Major Review: 2026-10-22 (anniversary)

Research on Schedule:
├─ Check latest version released?
├─ Any breaking changes?
├─ Any new best practices?
├─ Any deprecations?
├─ Any security updates?

Update Skill if any changes found!
```

---

## Related Resources

- [SKILL.md](SKILL.md) — Main Skill framework
- [INTERACTIVE-DISCOVERY.md](INTERACTIVE-DISCOVERY.md) — Gathering requirements
- [SKILL-UPDATE-ADVISOR.md](SKILL-UPDATE-ADVISOR.md) — Analyzing for updates
- [skill-factory.md](../../agents/alfred/skill-factory.md) — Orchestration

---

**Version**: 0.1.0
**Last Updated**: 2025-10-22
**Framework**: MoAI-ADK + Claude Skills + skill-factory
