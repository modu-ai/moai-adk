---
source: https://opencode.ai/docs/gitlab/
fetched: 2026-01-08
title: GitLab Integration
---

# GitLab Integration for OpenCode

## Overview

OpenCode provides two integration methods for GitLab workflows:

1. GitLab CI/CD Pipeline
2. GitLab Duo Integration

## GitLab CI/CD Pipeline

OpenCode functions as a standard GitLab pipeline component using the community-maintained [nagyv/gitlab-opencode](https://gitlab.com/nagyv/gitlab-opencode) component.

### Key Capabilities

- Custom configuration per job invocation
- Streamlined setup with minimal configuration needs
- Flexible input parameters for behavior customization

### Implementation

1. Store authentication JSON as masked CI variables
2. Add component reference to `.gitlab-ci.yml`
3. Specify configuration directory and prompts

### Example Configuration

```yaml
include:
  - component: gitlab.com/nagyv/gitlab-opencode/opencode@main
    inputs:
      config_dir: .opencode
      prompt: "Review this merge request"

variables:
  OPENCODE_AUTH: $OPENCODE_AUTH_JSON
```

## GitLab Duo Integration

OpenCode operates within GitLab CI/CD pipelines, activated by mentioning `@opencode` in comments.

### Supported Functions

- Issue analysis and explanation
- Automated fixes and feature implementation via merge requests
- Code review capabilities

### Setup Requirements

1. **GitLab Environment Configuration**
   - Configure GitLab CI/CD environment
   - Set up pipeline triggers

2. **API Key Acquisition**
   - Obtain API key from an AI model provider

3. **Service Account Creation**
   - Create a dedicated service account for OpenCode

4. **CI Variable Configuration**
   - Store credentials as protected CI variables

5. **Flow Configuration**
   - Create configuration file with authentication steps

### Usage Examples

#### Request Issue Explanation

```
@opencode Explain this issue and identify the root cause
```

#### Implement Fixes

```
@opencode Implement a fix for this issue
```

The assistant handles:
- Branch creation
- Code implementation
- Merge request generation

#### Review Merge Requests

```
@opencode Review this merge request for security and performance
```

## Configuration File Example

```yaml
# .gitlab-ci.yml
stages:
  - opencode

opencode:
  stage: opencode
  image: node:20
  script:
    - npm install -g opencode
    - opencode run "$PROMPT"
  variables:
    ANTHROPIC_API_KEY: $ANTHROPIC_API_KEY
  rules:
    - if: $CI_PIPELINE_SOURCE == "merge_request_event"
```

## Best Practices

1. **Use protected variables:** Store API keys as protected CI variables
2. **Limit branch access:** Restrict which branches can trigger OpenCode
3. **Review generated code:** Always review AI-generated merge requests
4. **Set appropriate permissions:** Use minimal required GitLab permissions
