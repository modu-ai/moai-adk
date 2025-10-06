# CLI Core System Analysis Report

**ANALYSIS:CLI-001 | Chain: @SPEC:ANALYSIS-001 -> @SPEC:ANALYSIS-001 -> @CODE:ANALYSIS-001**

**Generated**: 2025-10-01
**Analyzed Path**: `/Users/goos/MoAI/MoAI-ADK/moai-adk-ts/src/cli/`
**Total Files**: 13 TypeScript files
**Total LOC**: 3,388 lines

---

## Executive Summary

### Complexity Analysis

| Metric | Status | Score | Notes |
|--------|--------|-------|-------|
| **Overall Complexity** | 🟡 Medium | 6.5/10 | Well-structured but some files exceed guidelines |
| **Maintainability** | 🟢 Good | 8/10 | Clear separation of concerns, good naming |
| **Test Coverage** | 🟡 Partial | 6/10 | 4 test files present, but coverage appears limited |
| **Code Organization** | 🟢 Excellent | 9/10 | Follows clear command pattern, good modularity |
| **Error Handling** | 🟢 Good | 8/10 | Consistent try-catch blocks across commands |

### Quick Stats

- **Commands**: 6 (init, doctor, status, update, restore, help)
- **Test Files**: 4 (help, restore, status, update)
- **Largest File**: `doctor.ts` (437 LOC) - **⚠️ Exceeds 300 LOC guideline**
- **Average File Size**: 260 LOC
- **@TAG Usage**: 65 occurrences - Good traceability
- **Error Handling**: 39 try-catch blocks - Comprehensive coverage

---

## File Hierarchy Structure

```
moai-adk-ts/src/cli/
├── index.ts (264 LOC)                      # CLI entry point, command setup
│   ├── @CODE:CLI-ENTRY-001
│   └── @CODE:CLI-APP-001
│
├── commands/
│   ├── doctor.ts (437 LOC) ⚠️              # System diagnostics
│   │   ├── @CODE:CLI-002
│   │   ├── @SPEC:DOCTOR-RESULT-001
│   │   └── 10+ helper methods
│   │
│   ├── init.ts (276 LOC)                   # Project initialization
│   │   ├── @CODE:CLI-001
│   │   └── Orchestrates setup workflow
│   │
│   ├── status.ts (368 LOC) ⚠️              # Project status display
│   │   ├── @CODE:CLI-003
│   │   └── Version & file counting
│   │
│   ├── update.ts (282 LOC)                 # Update management
│   │   ├── @CODE:CLI-004
│   │   └── Orchestrator integration
│   │
│   ├── restore.ts (279 LOC)                # Backup restoration
│   │   ├── @CODE:CLI-005
│   │   └── Validation & restore logic
│   │
│   └── help.ts (317 LOC) ⚠️                # Help system
│       ├── @CODE:CLI-006
│       └── Comprehensive command docs
│
├── prompts/
│   └── init-prompts.ts (379 LOC) ⚠️        # Interactive prompts
│       ├── @CODE:INTERACTIVE-INIT-019
│       └── 8+ prompt functions
│
└── config/
    └── config-builder.ts (202 LOC)         # Config generation
        └── @CODE:INTERACTIVE-INIT-019
```

**⚠️ Files exceeding 300 LOC guideline**: 4 files

---

## TRUST 5 Principles Compliance

### T - Test First (TDD)

**Status**: 🟡 Partial Compliance

**Findings**:
- ✅ 4 test files present (`help.test.ts`, `restore.test.ts`, `status.test.ts`, `update.test.ts`)
- ❌ Missing tests: `init.test.ts`, `doctor.test.ts`
- ❌ No tests for `init-prompts.ts` (379 LOC)
- ❌ No tests for `config-builder.ts` (202 LOC)

**Test Coverage Estimate**: ~40-50% (based on file presence)

**Recommendations**:
1. Add `doctor.test.ts` - critical for system diagnostics validation
2. Add `init.test.ts` - highest impact command needs comprehensive testing
3. Add `prompts.test.ts` - interactive flows need validation
4. Add `config-builder.test.ts` - configuration logic needs unit tests

### R - Readable Code

**Status**: 🟢 Good Compliance

**Strengths**:
- ✅ Consistent naming conventions (class names, method names)
- ✅ Clear function purposes with descriptive names
- ✅ Good use of TypeScript interfaces for type safety
- ✅ Proper @TAG usage for traceability (65 tags across files)
- ✅ JSDoc comments present on public APIs

**Concerns**:
- ⚠️ `doctor.ts` contains 10+ methods in single class (437 LOC)
- ⚠️ `init-prompts.ts` has 8+ standalone functions (could be class-based)
- ⚠️ Some methods exceed 50 LOC guideline (e.g., `runInteractive` in init.ts)

### U - Unified Architecture

**Status**: 🟢 Excellent Compliance

**Strengths**:
- ✅ Clear Command Pattern implementation
- ✅ Consistent structure across all commands:
  - Options interface
  - Result interface
  - Command class with `run()` method
- ✅ Proper separation of concerns:
  - Commands handle orchestration
  - Prompts handle user interaction
  - Config handles data transformation
- ✅ Dependency injection (e.g., `SystemDetector` in constructors)

**Architecture Pattern**:
```typescript
// Consistent pattern across all commands
interface CommandOptions { /* ... */ }
interface CommandResult { /* ... */ }

class CommandName {
  async run(options: CommandOptions): Promise<CommandResult> {
    // 1. Validation
    // 2. Business logic
    // 3. Return result
  }
}
```

### S - Secured by Design

**Status**: 🟢 Good Compliance

**Findings**:
- ✅ Input validation present (`InputValidator.validatePath`, `validateProjectName`)
- ✅ Path validation to prevent directory traversal (`validateProjectPath`)
- ✅ Proper error handling with try-catch blocks (39 occurrences)
- ✅ No hardcoded credentials found
- ✅ Backup creation before destructive operations

**Concerns**:
- ⚠️ `init.ts` creates directories without sanitization in some paths
- ⚠️ No explicit rate limiting or denial-of-service protection
- ℹ️ Relies on external validators - ensure they're comprehensive

### T - Trackable (@TAG System)

**Status**: 🟢 Good Compliance

**Findings**:
- ✅ 65 @TAG occurrences across 13 files
- ✅ Consistent TAG chain usage:
  - `@CODE` → `@SPEC` → `@SPEC` → `@CODE`
- ✅ Related tags properly referenced (`@CODE`, `@CODE`, `@CODE`)

**TAG Distribution**:
- `index.ts`: 4 tags
- `doctor.ts`: 7+ tags
- `init.ts`: 5+ tags
- `status.ts`: 12+ tags
- `update.ts`: 13+ tags
- `restore.ts`: 9+ tags
- `help.ts`: 9+ tags

**Missing TAG Chains**:
- Some utility methods lack @TAG documentation
- Test files could benefit from @TEST tags

---

## Technical Debt Identification

### @CODE:CLI-TD-001 - File Size Violations

**Priority**: High
**Files Affected**: 4

```
doctor.ts (437 LOC)          - Exceeds 300 LOC by 46%
init-prompts.ts (379 LOC)    - Exceeds 300 LOC by 26%
status.ts (368 LOC)          - Exceeds 300 LOC by 23%
help.ts (317 LOC)            - Exceeds 300 LOC by 6%
```

**Impact**: Reduced maintainability, harder to test, violates TRUST principles

**Suggested Remediation**:
1. **doctor.ts**: Extract backup listing logic to `BackupManager` class
2. **init-prompts.ts**: Group prompts into `PromptOrchestrator` class
3. **status.ts**: Extract file counting to separate utility class
4. **help.ts**: Extract command definitions to JSON/YAML configuration

**Effort**: Medium (2-3 days)

### @CODE:CLI-TD-002 - Missing Test Coverage

**Priority**: Critical
**Coverage Gap**: ~50%

**Missing Tests**:
- `doctor.ts` (437 LOC) - NO TESTS
- `init.ts` (276 LOC) - NO TESTS
- `init-prompts.ts` (379 LOC) - NO TESTS
- `config-builder.ts` (202 LOC) - NO TESTS

**Impact**: High risk of regression, difficult to refactor with confidence

**Suggested Remediation**:
1. Add `doctor.test.ts` with system check mocking
2. Add `init.test.ts` with orchestrator mocking
3. Add `prompts.test.ts` with inquirer mocking
4. Add `config-builder.test.ts` for configuration logic

**Effort**: High (4-5 days)

### @CODE:CLI-TD-003 - Method Complexity

**Priority**: Medium
**Affected Methods**: 3+

**Long Methods** (>50 LOC):
- `InitCommand.runInteractive()` - ~200 LOC
- `DoctorCommand.run()` - ~97 LOC
- `StatusCommand.run()` - ~95 LOC

**Impact**: Difficult to understand, test, and maintain

**Suggested Remediation**:
1. Extract validation logic to separate methods
2. Extract display logic to formatting utilities
3. Use step-based orchestration pattern

**Effort**: Low-Medium (1-2 days)

### @CODE:CLI-TD-004 - Console.log Usage

**Priority**: Low
**Occurrences**: 73

**Findings**:
- Mix of `console.log`, `logger.info`, `logger.error`
- Inconsistent logging approach across files
- Some files use `console.log` directly (28 occurrences in `init.ts`, `doctor.ts`)

**Impact**: Difficult to control logging levels, poor production debugging

**Suggested Remediation**:
1. Standardize on `logger` utility everywhere
2. Remove all direct `console.log` calls
3. Add log level configuration (debug, info, warn, error)

**Effort**: Low (1 day)

### @CODE:CLI-TD-005 - Hardcoded Strings

**Priority**: Low
**Occurrences**: Multiple

**Findings**:
- Error messages hardcoded in multiple languages
- Some i18n support exists (locale selection) but not fully utilized
- Help text hardcoded in `help.ts`

**Impact**: Difficult to maintain multiple languages, poor i18n support

**Suggested Remediation**:
1. Extract all user-facing strings to i18n resource files
2. Use consistent translation keys
3. Support dynamic locale switching

**Effort**: Medium (2-3 days)

---

## Top 5 Files - Detailed Analysis

### 1. doctor.ts (437 LOC)

**Purpose**: System diagnostics and backup management

**Strengths**:
- ✅ Comprehensive system checking
- ✅ Good error handling
- ✅ Clear method naming

**Issues**:
- ❌ File too large (437 LOC > 300 LOC limit)
- ❌ Mixes two concerns: diagnostics + backup listing
- ❌ No unit tests
- ❌ `listBackups()` method is 68 LOC (should be <50)

**Refactoring Suggestions**:
```typescript
// Extract backup functionality
class BackupManager {
  async listBackups(): Promise<DoctorResult>
  async findBackupDirectories(): Promise<string[]>
  async printBackupInfo(path: string): Promise<void>
  async getBackupContents(path: string): Promise<string[]>
}

// Keep diagnostics focused
class DoctorCommand {
  constructor(
    private detector: SystemDetector,
    private backupManager: BackupManager
  ) {}

  async run(options): Promise<DoctorResult>
  private printResults(summary): void
  private categorizeResults(results): CategorizedResults
}
```

**Priority**: High

### 2. init-prompts.ts (379 LOC)

**Purpose**: Interactive prompt orchestration

**Strengths**:
- ✅ Good UX with step indicators
- ✅ Comprehensive validation
- ✅ i18n support

**Issues**:
- ❌ File too large (379 LOC > 300 LOC limit)
- ❌ 8+ standalone functions (should be class-based)
- ❌ No unit tests
- ❌ Tight coupling to `inquirer` library

**Refactoring Suggestions**:
```typescript
class PromptOrchestrator {
  async promptLocale(): Promise<Partial<InitAnswers>>
  async promptBasicInfo(name?, mode?): Promise<Partial<InitAnswers>>
  async promptMode(): Promise<Partial<InitAnswers>>
  async promptGitConfig(): Promise<Partial<InitAnswers>>
  async promptGitHubConfig(mode): Promise<Partial<InitAnswers>>
  async promptSpecWorkflow(mode, github): Promise<Partial<InitAnswers>>
  async promptAutoPush(github): Promise<Partial<InitAnswers>>

  async runFullWorkflow(defaults?): Promise<InitAnswers>
}

// Makes testing easier with dependency injection
class PromptOrchestrator {
  constructor(
    private prompter: IPrompter = new InquirerAdapter()
  ) {}
}
```

**Priority**: High

### 3. status.ts (368 LOC)

**Purpose**: Project status reporting

**Strengths**:
- ✅ Comprehensive status information
- ✅ Good type definitions
- ✅ Has test coverage

**Issues**:
- ❌ File too large (368 LOC > 300 LOC limit)
- ❌ File counting logic could be extracted
- ⚠️ Some duplication in path checking

**Refactoring Suggestions**:
```typescript
// Extract file operations
class ProjectFileCounter {
  async countProjectFiles(path: string): Promise<FileCount>
  private async countFilesInDirectory(dir: string): Promise<number>
}

// Extract version operations
class VersionChecker {
  async getVersionInfo(path: string): Promise<VersionInfo>
  async checkForUpdates(path: string): Promise<boolean>
}

// Keep status orchestration
class StatusCommand {
  constructor(
    private fileCounter: ProjectFileCounter,
    private versionChecker: VersionChecker
  ) {}

  async run(options): Promise<StatusResult>
  async checkProjectStatus(path): Promise<ProjectStatus>
}
```

**Priority**: Medium

### 4. help.ts (317 LOC)

**Purpose**: Help system and documentation

**Strengths**:
- ✅ Comprehensive command documentation
- ✅ Well-structured help output
- ✅ Has test coverage

**Issues**:
- ⚠️ File slightly large (317 LOC, close to limit)
- ❌ Command definitions hardcoded (should be config)
- ❌ Help text formatting mixed with data

**Refactoring Suggestions**:
```typescript
// commands.config.ts
export const COMMAND_DEFINITIONS: CommandHelp[] = [
  {
    name: 'init',
    description: 'Initialize a new MoAI-ADK project',
    // ... rest of config
  }
];

// help-formatter.ts
class HelpFormatter {
  formatGeneralHelp(commands: CommandHelp[]): string
  formatCommandHelp(command: CommandHelp): string
}

// help.ts (simplified)
class HelpCommand {
  constructor(
    private commands: CommandHelp[],
    private formatter: HelpFormatter
  ) {}

  async run(options): Promise<HelpResult>
}
```

**Priority**: Low

### 5. init.ts (276 LOC)

**Purpose**: Project initialization orchestration

**Strengths**:
- ✅ Clear workflow steps
- ✅ Good user feedback
- ✅ Proper validation

**Issues**:
- ❌ No unit tests (critical gap)
- ❌ `runInteractive()` method is ~200 LOC (way over 50 LOC limit)
- ⚠️ Complex path logic could be extracted

**Refactoring Suggestions**:
```typescript
class InitCommand {
  async runInteractive(options): Promise<InitResult> {
    this.displayBanner();
    await this.verifySystem();
    const config = await this.collectConfiguration(options);
    const projectPath = this.determinePath(config, options);
    await this.validatePath(projectPath);
    await this.saveConfiguration(projectPath, config);
    return await this.executeInstallation(projectPath, config);
  }

  // Each step becomes a focused method < 50 LOC
  private async verifySystem(): Promise<void>
  private async collectConfiguration(options): Promise<MoAIConfig>
  private determinePath(config, options): string
  private async validatePath(path): Promise<void>
  private async saveConfiguration(path, config): Promise<void>
  private async executeInstallation(path, config): Promise<InitResult>
}
```

**Priority**: High

---

## Refactoring Priorities

### High Priority (Immediate Action)

1. **Add Missing Tests** (@CODE:CLI-TD-002)
   - Impact: Critical for reliability
   - Files: doctor.ts, init.ts, init-prompts.ts, config-builder.ts
   - Effort: 4-5 days

2. **Refactor doctor.ts** (@CODE:CLI-TD-001)
   - Impact: Largest file, violates multiple principles
   - Extract BackupManager class
   - Effort: 1-2 days

3. **Refactor init.ts.runInteractive()** (@CODE:CLI-TD-003)
   - Impact: Most complex method in CLI
   - Extract step methods
   - Effort: 1 day

### Medium Priority (Next Sprint)

4. **Refactor init-prompts.ts** (@CODE:CLI-TD-001)
   - Convert to class-based PromptOrchestrator
   - Effort: 2 days

5. **Refactor status.ts** (@CODE:CLI-TD-001)
   - Extract file counting utilities
   - Effort: 1 day

6. **Standardize Logging** (@CODE:CLI-TD-004)
   - Replace console.log with logger
   - Effort: 1 day

### Low Priority (Future Improvements)

7. **Extract help.ts command definitions** (@CODE:CLI-TD-005)
   - Move to configuration file
   - Effort: 1 day

8. **Improve i18n coverage** (@CODE:CLI-TD-005)
   - Extract all user-facing strings
   - Effort: 2-3 days

---

## Action Item Checklist

### Testing & Quality

- [ ] Add `doctor.test.ts` with system check mocking
- [ ] Add `init.test.ts` with orchestrator mocking
- [ ] Add `prompts.test.ts` with inquirer mocking
- [ ] Add `config-builder.test.ts` for configuration logic
- [ ] Measure actual test coverage (target: 85%+)
- [ ] Add integration tests for full CLI workflows

### Refactoring & Code Quality

- [ ] Extract `BackupManager` from `doctor.ts`
- [ ] Convert `init-prompts.ts` to class-based `PromptOrchestrator`
- [ ] Split `InitCommand.runInteractive()` into step methods
- [ ] Extract `ProjectFileCounter` from `status.ts`
- [ ] Extract `VersionChecker` from `status.ts`
- [ ] Move help command definitions to config file
- [ ] Replace all `console.log` with `logger` calls

### Architecture Improvements

- [ ] Add dependency injection container
- [ ] Create abstract `BaseCommand` class
- [ ] Implement command middleware pattern
- [ ] Add command result transformers
- [ ] Consider CQRS pattern for complex commands

### Documentation

- [ ] Add architectural decision records (ADRs)
- [ ] Document command pattern implementation
- [ ] Add contribution guide for new commands
- [ ] Document testing strategy
- [ ] Add troubleshooting guide

### Security & Validation

- [ ] Audit all input validation points
- [ ] Add path traversal attack tests
- [ ] Add malicious input fuzzing tests
- [ ] Review error messages for information leakage
- [ ] Add rate limiting for API calls

---

## Metrics Summary

### Code Quality Metrics

| Metric | Current | Target | Status |
|--------|---------|--------|--------|
| Test Coverage | ~40-50% | ≥85% | 🔴 Below Target |
| Avg File Size | 260 LOC | ≤300 LOC | 🟡 Close to Limit |
| Max File Size | 437 LOC | ≤300 LOC | 🔴 Over Limit |
| Files > 300 LOC | 4 | 0 | 🔴 Multiple Violations |
| @TAG Coverage | 65 tags | All public APIs | 🟢 Good |
| Error Handling | 39 blocks | Comprehensive | 🟢 Good |
| Console.log Usage | 73 | 0 | 🔴 Should Use Logger |

### Complexity Metrics

| File | LOC | Methods | Complexity | Status |
|------|-----|---------|------------|--------|
| doctor.ts | 437 | 15 | High | 🔴 Refactor |
| init-prompts.ts | 379 | 11 | Medium | 🟡 Review |
| status.ts | 368 | 7 | Medium | 🟡 Review |
| help.ts | 317 | 8 | Low | 🟢 OK |
| update.ts | 282 | 7 | Low | 🟢 OK |
| restore.ts | 279 | 5 | Low | 🟢 OK |
| init.ts | 276 | 3 | High | 🟡 Review |
| index.ts | 264 | 4 | Low | 🟢 OK |
| config-builder.ts | 202 | 7 | Low | 🟢 OK |

---

## Conclusion

The CLI Core System demonstrates **good architectural design** with clear separation of concerns and consistent command patterns. However, there are **significant gaps in test coverage** and **several files that exceed size guidelines**.

**Key Strengths**:
- ✅ Excellent unified architecture (Command Pattern)
- ✅ Good @TAG traceability
- ✅ Comprehensive error handling
- ✅ Strong type safety with TypeScript

**Critical Issues**:
- ❌ Only ~40-50% test coverage (target: 85%+)
- ❌ 4 files exceed 300 LOC guideline
- ❌ Some methods exceed 50 LOC guideline
- ❌ Inconsistent logging approach

**Immediate Next Steps**:
1. Add missing test files (4-5 days effort)
2. Refactor `doctor.ts` to extract BackupManager (1-2 days)
3. Split `InitCommand.runInteractive()` into step methods (1 day)

**Estimated Effort to Full Compliance**:
- **High Priority**: 6-8 days
- **Medium Priority**: 4-5 days
- **Low Priority**: 3-4 days
- **Total**: 13-17 days

---

**Report Generated**: 2025-10-01
**Analyzed By**: Claude Code Analysis System
**Next Review**: After refactoring implementation
