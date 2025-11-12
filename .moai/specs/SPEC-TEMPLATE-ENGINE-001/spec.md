---
id: TEMPLATE-ENGINE-001
domain: TEMPLATE-ENGINE
title: "Template Engine"
version: "1.0.0"
status: "completed"
created: "2025-11-10"
author: "GoosLab"
---


## SPEC Overview

This SPEC defines the template engine system for MoAI-ADK, which provides Jinja2-style templating with variable substitution and conditional sections for GitHub templates and configuration files.

## Requirements

- **Jinja2 Integration**: Support Jinja2-style templating with variable substitution
- **Conditional Sections**: Support conditional sections with syntax like {{#ENABLE_TRUST_5}}...{{/ENABLE_TRUST_5}}
- **Variable Management**: Extract template variables from project configuration
- **File Processing**: Support file-based and string-based template rendering

## Implementation Files


## Acceptance Criteria

- ✅ Jinja2 template rendering with variable substitution
- ✅ Conditional section support with proper syntax
- ✅ Template variable extraction from configuration
- ✅ File-based and string-based rendering support
- ✅ Error handling and validation for templates
- ✅ Performance optimization for large template sets

## Traceability Chain

```
```
