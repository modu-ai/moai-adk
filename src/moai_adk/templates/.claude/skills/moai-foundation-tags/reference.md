# Foundation Tags - Reference Documentation

## Official Documentation

### Core References
- **MoAI-ADK TAG System**: [Internal Documentation]
- **TAG Management Guidelines**: `.moai/docs/tag-management.md`
- **@TAG Chain Documentation**: `CLAUDE.md` - Documentation Reference Map

### TAG Patterns and Standards

#### TAG Structure
```
@{DOMAIN}-{TOPIC}-{###}
├─ DOMAIN: Project area (SPEC, CODE, TEST, DOC)
├─ TOPIC: Specific subject area
└─ ###: Sequential ID (001-999)
```

#### TAG Chain Integrity
- **SPEC → CODE**: Requirements to implementation
- **CODE → TEST**: Implementation to verification
- **TEST → DOC**: Verification to documentation
- **DOC → SPEC**: Documentation to requirements

#### TAG Validation Rules
1. **Format Compliance**: Correct @TAG-DOMAIN-TOPIC-### structure
2. **Chain Completeness**: All four TAG types present
3. **Reference Accuracy**: Links point to existing files
4. **Traceability**: Clear lineage from SPEC to DOC

### Tools and Integration

#### TAG Management Tools
- **Git Integration**: Automatic TAG commit validation
- **Documentation Generation**: TAG-based doc assembly
- **Search and Discovery**: TAG-based content finding
- **Quality Gates**: TAG completeness validation

#### Best Practices
- **TAG First**: Always create TAGs before implementation
- **Consistent Naming**: Use established topic conventions
- **Complete Chains**: Maintain SPEC→CODE→TEST→DOC links
- **Regular Audits**: Periodic orphan detection and cleanup

---

## External References

### Version Control and Traceability
- **Git TAG Best Practices**: [Git Documentation](https://git-scm.com/book/en/v2/Git-Basics-Tagging)
- **Semantic Versioning**: [SemVer.org](https://semver.org/)
- **Conventional Commits**: [Conventional Commits](https://www.conventionalcommits.org/)

### Documentation Standards
- **Markdown Best Practices**: [CommonMark](https://commonmark.org/)
- **Technical Writing**: [Google Developer Documentation Style Guide](https://developers.google.com/tech-writing)

---

**Last Updated**: 2025-11-11
**Related Skills**: moai-foundation-specs, moai-foundation-trust
