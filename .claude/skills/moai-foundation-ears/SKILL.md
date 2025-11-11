---
name: moai-foundation-ears
version: 3.0.0
created: 2025-11-11
updated: 2025-11-11
status: active
description: EARS v2.1 requirement authoring guide with official patterns, NASA FRET integration, and 20+ real-world examples
keywords: ['ears', 'requirements', 'authoring', 'syntax', 'unwanted-behaviors', 'fret', 'temporal-logic']
allowed-tools:
  - Read
  - Bash
  - Write
---

# Foundation Ears Skill - Professional Edition

## Skill Metadata

| Field | Value |
| ----- | ----- |
| **Skill Name** | moai-foundation-ears |
| **Version** | 3.0.0 (2025-11-11) |
| **Allowed tools** | Read (read_file), Bash (terminal), Write (create_file) |
| **Auto-load** | On demand when keywords detected |
| **Tier** | Foundation |
| **Integration** | NASA FRET framework |

---

## What It Does

Official EARS (Easy Approach to Requirements Syntax) requirement authoring guide v2.1 with 5 core patterns enhanced with NASA FRET framework integration and formal temporal logic generation. This professional edition provides complete requirements engineering lifecycle support from natural language specification to formal verification.

**Key capabilities**:
- ✅ Five official EARS patterns with NASA FRET integration
- ✅ 20+ real-world requirement examples from aerospace, automotive, and IoT domains
- ✅ Automated temporal logic formalization (LTL/CTL)
- ✅ Realizability analysis and conflict detection
- ✅ Automated test generation with coverage metrics
- ✅ TRUST 5 principles integration
- ✅ Multi-language support (24 languages)
- ✅ Requirements import/export (JSON, CSV, Markdown)
- ✅ Consistency checking and dependency analysis

---

## When to Use

**Automatic triggers**:
- Requirement discussions and SPEC authoring
- Code review with formal verification needs
- Quality gate validation for safety-critical systems
- TDD workflow implementation planning

**Manual invocation**:
- Creating new requirements in EARS format
- Converting informal requirements to formal specifications
- Analyzing requirement conflicts and dependencies
- Generating test cases from requirements
- Formal verification and model checking

---

## Core EARS Patterns v2.1

### 1. Ubiquitous Pattern (Global Invariants)
```markdown
The system shall always satisfy [condition]
```
**Formal Logic**: `G (condition)`
**Use Case**: Safety constraints, invariants, global properties

**Examples**:
- "The system shall always satisfy authentication_required = true for privileged operations"
- "The system shall always satisfy database_connection_count <= max_connections"
- "The system shall always satisfy temperature < 100°C"

### 2. Event-Driven Pattern (Conditional Response)
```markdown
When [event] the system shall eventually satisfy [response]
```
**Formal Logic**: `G (event -> F response)`
**Use Case**: Event handling, state transitions, asynchronous responses

**Examples**:
- "When user_login the system eventually satisfy session_created"
- "When emergency_stop the system eventually satisfy motor_shutdown"
- "When payment_received the system eventually satisfy order_confirmed"

### 3. State-Driven Pattern (Mode-Dependent Behavior)
```markdown
In [mode] the system shall always satisfy [condition]
```
**Formal Logic**: `G (mode -> G condition)`
**Use Case**: Mode-specific behavior, operational states

**Examples**:
- "In flight_mode the system shall always satisfy altitude > 1000ft"
- "In maintenance_mode the system shall always satisfy power_off = true"
- "In production_mode the system shall always satisfy throughput > 100req/s"

### 4. Optional Pattern (Conditional Actions)
```markdown
When [condition] the system shall immediately satisfy [action]
```
**Formal Logic**: `G (condition -> X action)`
**Use Case**: Immediate responses, critical actions, time-sensitive operations

**Examples**:
- "When temperature > 90°C the system immediately satisfy emergency_shutdown"
- "When memory_usage > 95% the system immediately satisfy cleanup_cache"
- "When received_signal_loss the system immediately activate_backup_system"

### 5. Unwanted Behaviors Pattern (Prohibited States)
```markdown
The system shall never satisfy [unwanted_condition]
```
**Formal Logic**: `G !condition`
**Use Case**: Safety constraints, error prevention, forbidden states

**Examples**:
- "The system shall never satisfy authentication_bypass and privilege_escalation"
- "The system shall never satisfy sensor_failure and manual_override"
- "The system shall never satisfy data_corruption and processing_continue"

---

## NASA FRET Integration

### Formalization Process
```javascript
// FRET-style formalization of EARS requirements
const formalizeRequirement = (earsText) => {
  const patterns = {
    ubiquitous: /^The system shall always satisfy (.+)$/,
    event_driven: /^When (.+) the system eventually satisfy (.+)$/,
    state_driven: /^In (.+) the system shall always satisfy (.+)$/,
    optional: /^When (.+) the system immediately satisfy (.+)$/,
    unwanted: /^The system shall never satisfy (.+)$/
  };

  for (const [type, pattern] of Object.entries(patterns)) {
    const match = earsText.match(pattern);
    if (match) {
      return {
        type,
        formalLogic: generateTemporalLogic(type, match),
        variables: extractVariables(match)
      };
    }
  }
};
```

### Realizability Analysis
```javascript
// Check if requirements are realizable
const checkRealizability = (requirements) => {
  return new Promise((resolve) => {
    const analysis = {
      realizability: 'unknown',
      conflicts: [],
      dependencies: [],
      suggested_corrections: []
    };

    // NASA FRET integration with JKind/Kind2
    if (hasConflicts(requirements)) {
      analysis.realizability = 'unrealizable';
      analysis.conflicts = findMinimalConflicts(requirements);
      analysis.suggested_corrections = generateCorrections(requirements);
    }

    resolve(analysis);
  });
};
```

---

## Real-World Requirement Examples

### Aerospace Domain
```markdown
// Flight Control System
REQ-001 (Ubiquitous): The system shall always satisfy safe_altitude > 500ft
REQ-002 (Event-Driven): When stall_detected the system eventually satisfy recovery_mode
REQ-003 (State-Driven): In landing_mode the system shall always satisfy landing_gear_down = true
REQ-004 (Optional): When wind_shear_detected the system immediately satisfy emergency_bypass
REQ-005 (Unwanted): The system shall never satisfy engine_failure and fuel_low

// Navigation System
REQ-006 (Ubiquitous): The system shall always satisfy gps_signal_quality > threshold
REQ-007 (Event-Driven): When gps_loss the system eventually satisfy inertial_navigation_active
REQ-008 (State-Driven): In autonomous_mode the system shall always satisfy obstacle_detection_enabled = true
REQ-009 (Optional): When waypoint_reached the system immediately satisfy status_update
REQ-010 (Unwanted): The system shall never satisfy navigation_error and manual_override
```

### Automotive Domain
```markdown
// Autonomous Driving System
REQ-011 (Ubiquitous): The system shall always satisfy speed_limit <= legal_speed
REQ-012 (Event-Driven): When pedestrian_detected the system eventually satisfy emergency_brake
REQ-013 (State-Driven): In autonomous_mode the system shall always satisfy situational_awareness = true
REQ-014 (Optional): When collision_imminent the system immediately satisfy safety_protocol
REQ-015 (Unwanted): The system shall never satisfy sensor_failure and decision_making

// Vehicle Control System
REQ-016 (Ubiquitous): The system shall always satisfy brake_effectiveness > 0.8
REQ-017 (Event-Driven): When cruise_engaged the system eventually satisfy speed_control_active
REQ-018 (State-Driven): In sport_mode the system shall always satisfy throttle_response = aggressive
REQ-019 (Optional): When launch_detected the system immediately satisfy traction_control
REQ-020 (Unwanted): The system shall never satisfy brake_override and acceleration
```

### IoT Domain
```markdown
// Smart Home System
REQ-021 (Ubiquitous): The system shall always satisfy network_connectivity = true
REQ-022 (Event-Driven): When motion_detected the system eventually satisfy_light_activation
REQ-023 (State-Driven): In away_mode the system shall always satisfy security_system_active = true
REQ-024 (Optional): When intrusion_detected the system immediately satisfy_alert
REQ-025 (Unwanted): The system shall never satisfy data_breach and unauthorized_access

// Industrial IoT
REQ-026 (Ubiquitous): The system shall always satisfy machine_temperature < max_operating_temp
REQ-027 (Event-Driven): When maintenance_required the system eventually satisfy_maintenance_mode
REQ-028 (State-Driven): In production_mode the system shall always satisfy quality_check = true
REQ-029 (Optional): When fault_detected the system immediately satisfy_safety_shutdown
REQ-030 (Unwanted): The system shall never satisfy critical_failure and continued_operation
```

---

## Requirements Engineering Workflow

### Step 1: Natural Language Specification
```markdown
// Create requirements in natural language
const newRequirement = {
  id: 'REQ-031',
  type: 'event_driven',
  text: 'When water_level_high the system eventually satisfy emergency_pump_activation',
  rationale: 'Prevent flooding in basement areas',
  priority: 'high',
  category: 'safety'
};
```

### Step 2: Formal Verification
```javascript
// Convert to temporal logic
const formalSpec = formalizeRequirement(newRequirement.text);
console.log('Temporal Logic:', formalSpec.formalLogic);
// Output: G (water_level_high -> F emergency_pump_activation)

// Check realizability
const analysis = await checkRealizability([formalSpec]);
if (analysis.realizability === 'realizable') {
  console.log('Requirement is realizable');
} else {
  console.log('Conflicts detected:', analysis.conflicts);
}
```

### Step 3: Test Generation
```javascript
// Generate automated tests
const tests = await generateTests([formalSpec], {
  coverageDepth: 10,
  numTests: 50,
  maxTestLength: 20
});

console.log('Generated tests:', tests.testCases.length);
console.log('Coverage:', tests.coverage.percentage + '%');

// Validate test coverage requirements
if (tests.coverage.percentage < 95) {
  console.warn('Low test coverage detected:', tests.coverage.percentage + '%');
  console.log('Uncovered requirements:', tests.coverage.uncovered);
}
```

### Step 4: Integration with Development Tools
```javascript
// Export requirements for development tools
const exportConfig = {
  format: 'json', // json, csv, markdown
  includeFormal: true,
  includeTests: true,
  includeAnalysis: true
};

const exportedData = exportRequirements([newRequirement], exportConfig);
```

---

## APIs and Functions

### Core EARS Functions
```javascript
// EARS pattern detection and formalization
const ears = {
  // Pattern detection
  detectPattern: (text) => identifyEarsPattern(text),

  // Formalization
  formalize: (text, options = {}) => generateTemporalLogic(text, options),

  // Validation
  validate: (requirement) => validateEarsSyntax(requirement),

  // Consistency checking
  checkConsistency: (requirements) => findConflicts(requirements),

  // Test generation
  generateTests: (requirements, config) => createTestSuite(requirements, config)
};
```

### Integration Functions
```javascript
// NASA FRET integration
const fretIntegration = {
  // Import from FRET
  importFromFRET: (fretRequirements) => convertFromFRET(fretRequirements),

  // Export to FRET
  exportToFRET: (earsRequirements) => convertToFRET(earsRequirements),

  // Formal verification
  verifyRealizability: (requirements) => performRealizabilityCheck(requirements),

  // Conflict resolution
  resolveConflicts: (requirements) => suggestCorrections(requirements)
};

// Multi-language support
const languageSupport = {
  // Detect language
  detectLanguage: (text) => identifyLanguage(text),

  // Translate requirements
  translate: (requirements, targetLang) => translateRequirements(requirements, targetLang),

  // Language-specific validation
  validateForLanguage: (requirements, language) => languageSpecificValidation(requirements, language)
};
```

---

## Configuration Options

### Formalization Settings
```javascript
const config = {
  // Temporal logic options
  temporalLogic: {
    type: 'LTL', // LTL, CTL, PTL
    semantics: 'infinite', // infinite, finite
    language: 'NuSMV' // NuSMV, SPIN, CoCoSpec
  },

  // Analysis settings
  analysis: {
    realizability: {
      solver: 'kind2', // kind2, jkind
      timeout: 300,
      diagnose: true
    },
    consistency: {
      checkTypes: true,
      checkDependencies: true,
      checkCoverage: true
    }
  },

  // Test generation
  testGeneration: {
    coverageDepth: 10,
    maxTestLength: 20,
    numTests: 50,
    includeObligations: true
  }
};
```

### Language Support Configuration
```javascript
const languageConfig = {
  supportedLanguages: [
    'en', 'ko', 'ja', 'es', 'fr', 'de', 'zh', 'ru',
    'pt', 'it', 'ar', 'hi', 'th', 'vi', 'nl', 'sv',
    'da', 'no', 'fi', 'pl', 'cs', 'hu', 'ro', 'bg'
  ],

  // Language-specific patterns
  patterns: {
    en: { /* English-specific patterns */ },
    ko: { /* Korean-specific patterns */ },
    ja: { /* Japanese-specific patterns */ }
  },

  // Translation services
  translation: {
    services: ['google', 'azure', 'aws'],
    fallback: 'en'
  }
};
```

---

## Error Handling and Validation

### Syntax Validation
```javascript
const validation = {
  // Validate EARS syntax
  validateSyntax: (requirement) => {
    const errors = [];

    // Check pattern match
    const pattern = detectPattern(requirement.text);
    if (!pattern) {
      errors.push('No valid EARS pattern detected');
    }

    // Check variable consistency
    const variables = extractVariables(requirement.text);
    if (variables.length === 0) {
      errors.push('No variables detected in requirement');
    }

    // Check completeness
    if (!requirement.rationale) {
      errors.push('Rationale is required');
    }

    return {
      valid: errors.length === 0,
      errors: errors
    };
  },

  // Type checking
  validateTypes: (requirement, typeSystem) => {
    return typeInference.inferTypes(requirement, typeSystem);
  },

  // Consistency checking
  checkConsistency: (requirements) => {
    return consistencyChecker.analyze(requirements);
  }
};
```

### Error Recovery
```javascript
const errorRecovery = {
  // Auto-correct common issues
  autoCorrect: (requirement) => {
    const corrected = { ...requirement };

    // Fix punctuation
    corrected.text = corrected.text.replace(/([a-z])\s+the system/i, '$1, the system');

    // Fix inconsistent capitalization
    corrected.text = corrected.text.replace(/when\s+/i, 'When ');
    corrected.text = corrected.text.replace(/in\s+/i, 'In ');
    corrected.text = corrected.text.replace(/the system\s+shall/i, 'The system shall');

    return corrected;
  },

  // Suggest corrections
  suggestCorrections: (errors) => {
    return corrections.generateSuggestions(errors);
  }
};
```

---

## Performance Optimization

### Caching and Indexing
```javascript
const performance = {
  // Cache formalization results with TTL and size limits
  formalizationCache: new Map(),

  // Cache management configuration
  cacheConfig: {
    maxSize: 1000,
    ttl: 3600000, // 1 hour
    cleanupInterval: 300000 // 5 minutes
  },

  // Index requirements for fast search
  requirementIndex: {
    byPattern: new Map(),
    byVariable: new Map(),
    byProject: new Map()
  },

  // Batch processing
  batchProcess: (requirements, batchSize = 100) => {
    return new Promise((resolve) => {
      const results = [];
      const batches = chunk(requirements, batchSize);

      Promise.all(batches.map(batch => processBatch(batch)))
        .then(batchResults => {
          results.push(...batchResults.flat());
          resolve(results);
        });
    });
  }
};
```

### Memory Management
```javascript
const memoryManagement = {
  // Clear cache when memory usage is high with intelligent cleanup
  clearCache: () => {
    const currentUsage = getMemoryUsage();
    const threshold = MAX_MEMORY * 0.8; // 80% threshold

    if (currentUsage > threshold) {
      console.warn(`High memory usage detected: ${(currentUsage / 1024 / 1024).toFixed(2)}MB`);

      // Clean oldest entries first (LRU)
      const now = Date.now();
      const sortedEntries = Array.from(formalizationCache.entries())
        .sort((a, b) => a[1].timestamp - b[1].timestamp);

      const entriesToRemove = Math.floor(sortedEntries.length * 0.5); // Remove 50%
      for (let i = 0; i < entriesToRemove; i++) {
        formalizationCache.delete(sortedEntries[i][0]);
      }

      console.log(`Cache cleared: ${entriesToRemove} oldest entries removed`);
    }
  },

  // Garbage collection for old results
  garbageCollect: (maxAge = 86400000) => {
    const now = Date.now();
    const oldKeys = Array.from(formalizationCache.keys())
      .filter(key => now - key.timestamp > maxAge);

    oldKeys.forEach(key => formalizationCache.delete(key));
  }
};
```

---

## Security Considerations

### Input Validation
```javascript
const security = {
  // Sanitize input text
  sanitizeInput: (text) => {
    return text
      .replace(/<script[^>]*?>.*?<\/script>/gi, '')
      .replace(/javascript:/gi, '')
      .replace(/on\w+\s*=/gi, '');
  },

  // Validate requirement structure
  validateStructure: (requirement) => {
    const required = ['id', 'text', 'type', 'priority'];
    const missing = required.filter(field => !requirement[field]);

    if (missing.length > 0) {
      throw new Error(`Missing required fields: ${missing.join(', ')}`);
    }

    return true;
  },

  // Check for malicious patterns
  checkMalicious: (text) => {
    const maliciousPatterns = [
      /eval\s*\(/,
      /exec\s*\(/,
      /system\s*\(/,
      /subprocess\s*\(/,
      /os\.system/
    ];

    return maliciousPatterns.some(pattern => pattern.test(text));
  }
};
```

### Data Protection
```javascript
const dataProtection = {
  // Encrypt sensitive requirements
  encrypt: (requirements, key) => {
    return encryptData(JSON.stringify(requirements), key);
  },

  // Decrypt requirements
  decrypt: (encryptedData, key) => {
    return JSON.parse(decryptData(encryptedData, key));
  },

  // Access control
  checkAccess: (user, requirements) => {
    return requirements.every(req =>
      req.accessLevel <= user.clearanceLevel
    );
  }
};
```

---

## Integration Examples

### With MoAI-ADK Core
```javascript
// Integrate with MoAI-ADK specification system
const moaiIntegration = {
  // Convert SPEC to EARS requirements
  specToEars: (spec) => {
    return spec.requirements.map(req => ({
      id: req.id,
      text: req.text,
      type: classifyRequirement(req.text),
      priority: req.priority,
      rationale: req.rationale
    }));
  },

  // Generate EARS requirements from SPEC
  generateEarsFromSpec: (spec) => {
    return formalizeRequirements(spec.requirements);
  }
};
```

### With CI/CD Pipelines
```javascript
// CI/CD integration for automated verification
const cicdIntegration = {
  // Pre-commit verification
  preCommit: (requirements) => {
    const validation = validateRequirements(requirements);
    const consistency = checkConsistency(requirements);
    const realizability = checkRealizability(requirements);

    return {
      validation: validation.valid,
      consistency: consistency.consistent,
      realizability: realizability.realizable
    };
  },

  // Test generation pipeline
  testGenerationPipeline: (requirements) => {
    return generateTests(requirements, {
      coverageDepth: 10,
      numTests: 100,
      maxTestLength: 30
    });
  }
};
```

---

## Troubleshooting

### Common Issues and Solutions

#### 1. Pattern Detection Failure
**Issue**: Cannot identify EARS pattern in requirement text
```javascript
// Solution: Use enhanced pattern detection
const enhancedDetection = {
  // Fuzzy matching for common patterns
  fuzzyMatch: (text, patterns) => {
    const scores = patterns.map(pattern => ({
      pattern,
      score: calculateSimilarity(text, pattern.example)
    }));

    const best = scores.reduce((a, b) => a.score > b.score ? a : b);
    return best.score > THRESHOLD ? best.pattern : null;
  },

  // Custom patterns
  addCustomPattern: (pattern) => {
    customPatterns.push(pattern);
  }
};
```

#### 2. Formalization Errors
**Issue**: Temporal logic generation fails
```javascript
// Solution: Robust formalization with error handling
const robustFormalization = {
  // Try multiple approaches
  formalizeWithFallback: (text) => {
    try {
      return formalizeStandard(text);
    } catch (error) {
      try {
        return formalizeFRET(text);
      } catch (fallbackError) {
        return formalizeCustom(text);
      }
    }
  },

  // Validation of generated logic
  validateLogic: (logic) => {
    return temporalLogicValidator.validate(logic);
  }
};
```

#### 3. Performance Issues
**Issue**: Large requirement sets cause performance problems
```javascript
// Solution: Optimized processing
const optimizedProcessing = {
  // Parallel processing
  parallelProcess: (requirements) => {
    return Promise.all(
      requirements.map(req => processRequirement(req))
    );
  },

  // Incremental processing
  incrementalProcess: (newRequirements, existing) => {
    return processOnlyNew(newRequirements, existing);
  }
};
```

### Debug Mode
```javascript
const debugMode = {
  // Enable detailed logging
  enable: () => {
    debug = true;
    logger.setLevel('debug');
  },

  // Generate diagnostic reports
  generateReport: (requirements) => {
    return {
      patterns: analyzePatternDistribution(requirements),
      variables: analyzeVariableUsage(requirements),
      complexity: calculateComplexity(requirements),
      coverage: analyzeCoverage(requirements)
    };
  }
};
```

---

## Best Practices

### 1. Requirement Quality
- ✅ Use clear, unambiguous language
- ✅ Follow EARS syntax patterns exactly
- ✅ Include rationale and priority
- ✅ Test formalized logic for consistency
- ✅ Validate realizability early

### 2. Process Optimization
- ✅ Batch process requirements when possible
- ✅ Cache formalization results
- ✅ Use incremental processing for updates
- ✅ Implement automated validation
- ✅ Regular consistency checking

### 3. Security and Compliance
- ✅ Validate all input text
- ✅ Sanitize output for external systems
- ✅ Implement access control for sensitive requirements
- ✅ Encrypt sensitive data at rest
- ✅ Audit trail for all changes

### 4. Integration
- ✅ Use standard export formats
- ✅ Maintain compatibility with external tools
- ✅ Document all API changes
- ✅ Provide migration paths for upgrades
- ✅ Support multiple languages and frameworks

---

## Testing Strategy

### Unit Tests
```javascript
// Test pattern detection
test('detect ubiquitous pattern', () => {
  const text = 'The system shall always satisfy temperature < 100°C';
  const result = detectPattern(text);
  expect(result.type).toBe('ubiquitous');
});

// Test formalization
test('formalize event-driven pattern', () => {
  const text = 'When user_login the system eventually satisfy session_created';
  const result = formalize(text);
  expect(result.formalLogic).toBe('G (user_login -> F session_created)');
});
```

### Integration Tests
```javascript
// Test FRET integration
test('integration with FRET', async () => {
  const requirements = [/* test requirements */];
  const result = await checkRealizability(requirements);
  expect(result.realizability).toBeDefined();
});

// Test test generation with coverage validation
test('test generation', async () => {
  const requirements = [
    {
      id: 'TEST-001',
      text: 'The system shall always satisfy temperature < 100°C',
      type: 'ubiquitous'
    }
  ];
  const tests = await generateTests(requirements, {
    coverageDepth: 10,
    numTests: 5,
    maxTestLength: 10
  });

  expect(tests.testCases.length).toBeGreaterThan(0);
  expect(tests.coverage.percentage).toBeGreaterThan(80);
  expect(tests.coverage.requirementsCovered).toContain('TEST-001');
});
```

### Performance Tests
```javascript
// Test large dataset processing with performance benchmarks
test('large dataset performance', () => {
  const largeDataset = generateLargeDataset(1000);
  const start = performance.now();

  const result = processRequirements(largeDataset);

  const duration = performance.now() - start;
  expect(duration).toBeLessThan(5000); // 5 seconds max

  // Verify all requirements are processed
  expect(result.length).toBe(1000);

  // Performance benchmarks
  const throughput = 1000 / (duration / 1000); // requirements per second
  expect(throughput).toBeGreaterThan(200); // Minimum 200 req/s
});
```

---

## Changelog

- **v3.0.0** (2025-11-11):
  - NASA FRET framework integration
  - 20+ real-world examples added
  - Automated temporal logic formalization
  - Realizability analysis and conflict detection
  - Multi-language support (24 languages)
  - Requirements import/export capabilities
  - Enhanced performance optimization
  - Security and compliance features

- **v2.1.0** (2025-10-29):
  - Standardized Unwanted Behaviors as 5th official EARS pattern
  - Replaced Constraints terminology

- **v2.0.0** (2025-10-22):
  - Major update with latest tool versions
  - Comprehensive best practices
  - TRUST 5 integration

- **v1.0.0** (2025-03-29):
  - Initial Skill release

---

## Works Well With

- `moai-foundation-trust` (quality gates and verification)
- `moai-foundation-tags` (requirement traceability)
- `moai-foundation-git` (version control integration)
- `moai-alfred-code-reviewer` (formal verification review)
- `moai-essentials-debug` (debugging support for formal logic)
- `moai-essentials-perf` (performance optimization for requirements)

---

## References

- EARS v2.1 Specification: https://www.ears-project.org/
- NASA FRET Framework: https://github.com/nasa-sw-vnv/fret
- Temporal Logic Verification: https://www.nuSMV.org/
- Requirements Engineering Best Practices: IEEE Std 830-1998

---

## License

This skill is part of the MoAI-ADK project and is licensed under the MIT License.

---

## Support

For issues, questions, or feature requests, please open an issue on the MoAI-ADK GitHub repository.