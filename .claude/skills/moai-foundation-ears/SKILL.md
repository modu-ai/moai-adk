---
name: moai-foundation-ears
description: EARS requirement authoring guide (Ubiquitous/Event/State/Optional/Constraints)
allowed-tools:
  - Read
  - Bash
  - Write
  - Edit
  - TodoWrite
---

# Alfred EARS Authoring Guide

## What it does

EARS (Easy Approach to Requirements Syntax) authoring guide for writing clear, testable requirements using 5 statement patterns.

## When to use

- “Writing SPEC”, “Requirements summary”, “EARS syntax”
- Automatically invoked by `/alfred:1-plan`
- When writing or refining SPEC documents

## How it works

EARS provides 5 statement patterns for structured requirements:

### 1. Ubiquitous (Basic Requirements)
**Format**: The system must provide [function]
**Example**: The system must provide user authentication function

### 2. Event-driven (event-based)
**Format**: WHEN If [condition], the system must [operate]
**Example**: WHEN When the user logs in, the system must issue a JWT token

### 3. State-driven
**Format**: WHILE When in [state], the system must [operate]
**Example**: WHILE When the user is authenticated, the system must allow access to protected resources

### 4. Optional (Optional function)
**Format**: If WHERE [condition], the system can [operate]
**Example**: If WHERE refresh token is provided, the system can issue a new access token

### 5. Constraints
**Format**: IF [condition], the system SHOULD [constrain]
**Example**: IF an invalid token is provided, the system SHOULD deny access

## Writing Tips

✅ Be specific and measurable
✅ Avoid vague terms (“adequate”, “sufficient”, “fast”)
✅ One requirement per statement
✅ Make it testable

## Examples

User: "Please write JWT authentication SPEC"
Claude: (applies EARS patterns to structure authentication requirements)
## Works well with

- moai-foundation-specs
