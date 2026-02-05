---
name: moai-formats-data
description: >
  Data format specialist covering TOON encoding, JSON/YAML optimization,
  serialization patterns, and data validation for modern applications. Use when
  optimizing data for LLM transmission, implementing high-performance
  serialization, validating data schemas, or converting between data formats.
license: Apache-2.0
compatibility: Designed for Claude Code
allowed-tools: Read Write Edit Grep Glob mcp__context7__resolve-library-id mcp__context7__get-library-docs
user-invocable: false
metadata:
  version: "2.0.0"
  category: "library"
  status: "active"
  updated: "2026-01-11"
  modularized: "true"
  tags: "formats, data, toon, serialization, validation, optimization"
  author: "MoAI-ADK Team"

# MoAI Extension: Triggers
triggers:
  keywords: ["serialization", "data format", "json", "yaml", "toon", "validation", "schema", "optimization"]
---

# Data Format Specialist

## Quick Reference

Advanced Data Format Management - Comprehensive data handling covering TOON encoding, JSON/YAML optimization, serialization patterns, and data validation for performance-critical applications.

Core Capabilities:

- TOON Encoding: 40-60% token reduction vs JSON for LLM communication
- JSON/YAML Optimization: Efficient serialization and parsing patterns
- Data Validation: Schema validation, type checking, error handling
- Format Conversion: Seamless transformation between data formats
- Performance: Optimized data structures and caching strategies
- Schema Management: Dynamic schema generation and evolution

When to Use:

- Optimizing data transmission to LLMs within token budgets
- High-performance serialization/deserialization
- Schema validation and data integrity
- Format conversion and data transformation
- Large dataset processing and optimization

Quick Start:

Create a TOONEncoder instance and call encode with a dictionary containing user and age fields to compress the data. The encoded result achieves 40-60% token reduction. Call decode to restore the original data structure.

Create a JSONOptimizer instance and call serialize_fast with a large dataset to achieve ultra-fast JSON processing.

Create a DataValidator instance and call create_schema with a dictionary defining name as a required string type. Call validate with the data and schema to check validity.

---

## Implementation Guide

### Core Concepts

TOON (Token-Optimized Object Notation):

- Custom binary-compatible format optimized for LLM token usage
- Type markers: # for numbers, ! for booleans, @ for timestamps, ~ for null
- 40-60% size reduction vs JSON for typical data structures
- Lossless round-trip encoding/decoding

Performance Optimization:

- Ultra-fast JSON processing with orjson achieving 2-5x faster than standard json
- Streaming processing for large datasets using ijson
- Intelligent caching with LRU eviction and memory management
- Schema compression and validation optimization

Data Validation:

- Type-safe validation with custom rules and patterns
- Schema evolution and migration support
- Cross-field validation and dependency checking
- Performance-optimized batch validation

### Basic Implementation

TOON Encoding for LLM Optimization:

Create a TOONEncoder instance. Define data with user object containing id, name, active boolean, and created datetime, plus permissions array. Call encode to compress and decode to restore. Compare sizes to verify reduction.

Fast JSON Processing:

Create a JSONOptimizer instance. Call serialize_fast to get bytes and deserialize_fast to parse. Use compress_schema with a type object and properties definition to optimize repeated validation.

Data Validation:

Create a DataValidator instance. Define user_schema with username requiring string type, minimum length 3, email requiring email type, and age as optional integer with minimum value 13. Call validate with user_data and schema, then check result for valid status, sanitized_data, or errors list.

### Common Use Cases

API Response Optimization:

Create a function to optimize API responses for LLM consumption by encoding data with TOONEncoder. Create a corresponding function to parse optimized responses by decoding TOON data back to dictionary.

Configuration Management:

Create a YAMLOptimizer instance and call load_fast with a config file path. Call merge_configs with base_config, env_config, and user_config for multi-file merging.

Large Dataset Processing:

Create a StreamProcessor with chunk_size of 8192. Define a process_item function that handles each item. Call process_json_stream with the file path and callback to process large JSON files without loading into memory.

---

## Advanced Features Overview

### Advanced TOON Features

See modules/toon-encoding.md for custom type handlers (UUID, Decimal), streaming TOON processing, batch TOON encoding, and performance characteristics with benchmarks.

### Advanced Validation Patterns

See modules/data-validation.md for cross-field validation, schema evolution and migration, custom validation rules, and batch validation optimization.

### Performance Optimization

See modules/caching-performance.md for intelligent caching strategies, cache warming and invalidation, memory management, and performance monitoring.

### JSON/YAML Advanced Features

See modules/json-optimization.md for streaming JSON processing, memory-efficient parsing, schema compression, and format conversion utilities.

---

## Works Well With

- moai-domain-backend - Backend data serialization and API responses
- moai-domain-database - Database data format optimization
- moai-foundation-core - MCP data serialization and transmission patterns
- moai-workflow-docs - Documentation data formatting
- moai-foundation-context - Context optimization for token budgets

---

## Module References

Core Implementation Modules:

- modules/toon-encoding.md - TOON encoding implementation
- modules/json-optimization.md - High-performance JSON/YAML
- modules/data-validation.md - Advanced validation and schemas
- modules/caching-performance.md - Caching strategies

Supporting Files:

- modules/INDEX.md - Module overview and integration patterns
- reference.md - Extended reference documentation
- examples.md - Complete working examples

---

## Technology Stack

Core Libraries:

- orjson: Ultra-fast JSON parsing and serialization
- PyYAML: YAML processing with C-based loaders
- ijson: Streaming JSON parser for large files
- python-dateutil: Advanced datetime parsing
- regex: Advanced regular expression support

Performance Tools:

- lru_cache: Built-in memoization
- pickle: Object serialization
- hashlib: Hash generation for caching
- functools: Function decorators and utilities

Validation Libraries:

- jsonschema: JSON Schema validation
- cerberus: Lightweight data validation
- marshmallow: Object serialization/deserialization
- pydantic: Data validation using Python type hints

---

## Resources

For working code examples, see [examples.md](examples.md).

Status: Production Ready
Last Updated: 2026-01-11
Maintained by: MoAI-ADK Data Team
