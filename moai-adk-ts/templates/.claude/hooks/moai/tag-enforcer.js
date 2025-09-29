/**
 * @file TAG Enforcer Hook - Auto @TAG validation and insertion
 * @author MoAI Team
 * @tags @SECURITY:TAG-ENFORCER-001 @FEATURE:AUTO-TAG-001
 */

import fs from 'fs';
import path from 'path';

/**
 * TAG patterns for different file types
 */
const TAG_PATTERNS = {
  typescript: {
    fileHeader: /^\/\*\*\s*\n\s*\*\s*@file\s+.*\n\s*\*\s*@author\s+.*\n\s*\*\s*@tags\s+@\w+:\w+-\d+/m,
    functionDoc: /\/\*\*[\s\S]*?\*\s*@tags\s+@\w+:\w+-\d+[\s\S]*?\*\//g,
    classDoc: /\/\*\*[\s\S]*?\*\s*@tags\s+@\w+:\w+-\d+[\s\S]*?\*\/\s*export\s+(class|interface)/g
  },
  python: {
    fileHeader: /^"""\s*@file\s+.*\n@author\s+.*\n@tags\s+@\w+:\w+-\d+/m,
    functionDoc: /"""\s*[\s\S]*?@tags\s+@\w+:\w+-\d+[\s\S]*?"""/g
  },
  javascript: {
    fileHeader: /^\/\*\*\s*\n\s*\*\s*@file\s+.*\n\s*\*\s*@author\s+.*\n\s*\*\s*@tags\s+@\w+:\w+-\d+/m,
    functionDoc: /\/\*\*[\s\S]*?\*\s*@tags\s+@\w+:\w+-\d+[\s\S]*?\*\//g
  }
};

/**
 * 16-Core TAG categories
 */
const TAG_CATEGORIES = {
  primary: ['REQ', 'DESIGN', 'TASK', 'TEST'],
  implementation: ['FEATURE', 'API', 'UI', 'DATA', 'UTIL'],
  quality: ['PERF', 'SEC', 'DOCS', 'TAG', 'DEBT', 'TODO']
};

class TAGEnforcer {
  name = 'tag-enforcer';

  /**
   * Execute TAG enforcement
   */
  async execute(input) {
    try {
      // Check if this is a Write/Edit/MultiEdit operation
      if (!this.isWriteOperation(input.tool_name)) {
        return { success: true };
      }

      const filePath = this.extractFilePath(input.tool_input || {});
      if (!filePath || !this.shouldEnforceTags(filePath)) {
        return { success: true };
      }

      const content = this.extractFileContent(input.tool_input || {});
      const fileType = this.getFileType(filePath);

      const validation = this.validateTags(content, fileType);

      if (!validation.isValid) {
        return {
          success: false,
          blocked: true,
          message: `🏷️ @TAG 규칙 위반: ${validation.violations.join(', ')}`,
          suggestions: this.generateTagSuggestions(filePath, content, fileType),
          exitCode: 2
        };
      }

      // If validation passes but improvements can be made
      if (validation.warnings.length > 0) {
        console.error(`⚠️ TAG 개선 권장: ${validation.warnings.join(', ')}`);
      }

      return {
        success: true,
        message: `✅ @TAG 규칙 준수됨 (${fileType})`
      };

    } catch (error) {
      // Don't block on errors, just log
      console.error(`TAG Enforcer warning: ${error.message}`);
      return { success: true };
    }
  }

  /**
   * Check if this is a file write operation
   */
  isWriteOperation(toolName) {
    return ['Write', 'Edit', 'MultiEdit', 'NotebookEdit'].includes(toolName);
  }

  /**
   * Extract file path from tool input
   */
  extractFilePath(toolInput) {
    return toolInput.file_path || toolInput.filePath || toolInput.notebook_path || null;
  }

  /**
   * Extract file content from tool input
   */
  extractFileContent(toolInput) {
    if (toolInput.content) return toolInput.content;
    if (toolInput.new_string) return toolInput.new_string;
    if (toolInput.new_source) return toolInput.new_source;
    if (toolInput.edits && Array.isArray(toolInput.edits)) {
      return toolInput.edits.map(edit => edit.new_string).join('\n');
    }
    return '';
  }

  /**
   * Check if file should be enforced for @TAG rules
   */
  shouldEnforceTags(filePath) {
    const enforceExtensions = ['.ts', '.tsx', '.js', '.jsx', '.py', '.go', '.rs', '.java', '.cpp', '.hpp'];
    const ext = path.extname(filePath);

    // Skip test files for now (they have different TAG requirements)
    if (filePath.includes('test') || filePath.includes('spec')) {
      return false;
    }

    // Skip node_modules and similar
    if (filePath.includes('node_modules') || filePath.includes('.git')) {
      return false;
    }

    return enforceExtensions.includes(ext);
  }

  /**
   * Get file type from extension
   */
  getFileType(filePath) {
    const ext = path.extname(filePath);
    switch (ext) {
      case '.ts':
      case '.tsx':
        return 'typescript';
      case '.js':
      case '.jsx':
        return 'javascript';
      case '.py':
        return 'python';
      default:
        return 'typescript'; // Default to TypeScript patterns
    }
  }

  /**
   * Validate @TAG compliance in content
   */
  validateTags(content, fileType) {
    const violations = [];
    const warnings = [];
    let isValid = true;

    // Check for file header @TAG
    if (!this.hasFileHeaderTag(content, fileType)) {
      violations.push('파일 헤더에 @tags 누락');
      isValid = false;
    }

    // Check for function/class @TAGs (less strict)
    const missingTags = this.findMissingFunctionTags(content, fileType);
    if (missingTags.length > 0) {
      warnings.push(`함수/클래스 @tags 권장: ${missingTags.length}개`);
    }

    // Check @TAG format validity
    const invalidTags = this.findInvalidTagFormats(content);
    if (invalidTags.length > 0) {
      violations.push(`잘못된 @TAG 형식: ${invalidTags.join(', ')}`);
      isValid = false;
    }

    return { isValid, violations, warnings };
  }

  /**
   * Check if file has proper header @TAG
   */
  hasFileHeaderTag(content, fileType) {
    const pattern = TAG_PATTERNS[fileType]?.fileHeader;
    if (!pattern) return true; // Skip if no pattern defined

    return pattern.test(content);
  }

  /**
   * Find functions/classes missing @TAG documentation
   */
  findMissingFunctionTags(content, fileType) {
    const missing = [];

    if (fileType === 'typescript' || fileType === 'javascript') {
      // Find exported functions without @tags
      const functionMatches = content.match(/export\s+(function|class|interface|const)\s+\w+/g) || [];
      const taggedFunctions = content.match(/\/\*\*[\s\S]*?@tags[\s\S]*?\*\/[\s\S]*?export\s+(function|class|interface|const)/g) || [];

      if (functionMatches.length > taggedFunctions.length) {
        missing.push(`${functionMatches.length - taggedFunctions.length} exported items`);
      }
    }

    return missing;
  }

  /**
   * Find invalid @TAG formats
   */
  findInvalidTagFormats(content) {
    const invalid = [];

    // Find all @TAG patterns
    const tagMatches = content.match(/@\w+:\w+(-\d+)?/g) || [];

    for (const tag of tagMatches) {
      const [category, id] = tag.substring(1).split(':');

      // Check if category is valid
      const allCategories = [...TAG_CATEGORIES.primary, ...TAG_CATEGORIES.implementation, ...TAG_CATEGORIES.quality];
      if (!allCategories.includes(category)) {
        invalid.push(`Unknown category: ${category}`);
      }

      // Check if ID format is valid (should end with -NNN)
      if (id && !id.match(/.*-\d+$/)) {
        invalid.push(`Invalid ID format: ${tag}`);
      }
    }

    return invalid;
  }

  /**
   * Generate @TAG suggestions for the file
   */
  generateTagSuggestions(filePath, content, fileType) {
    const suggestions = [];
    const fileName = path.basename(filePath, path.extname(filePath));

    // Suggest file header
    suggestions.push('파일 헤더 예시:');
    suggestions.push('/**');
    suggestions.push(` * @file ${fileName} implementation`);
    suggestions.push(' * @author MoAI Team');

    // Suggest appropriate @TAG based on file content
    if (content.includes('class ') || content.includes('interface ')) {
      suggestions.push(' * @tags @FEATURE:' + fileName.toUpperCase() + '-001 @API:' + fileName.toUpperCase() + '-API-001');
    } else if (content.includes('export function') || content.includes('export const')) {
      suggestions.push(' * @tags @UTIL:' + fileName.toUpperCase() + '-001');
    } else {
      suggestions.push(' * @tags @FEATURE:' + fileName.toUpperCase() + '-001');
    }

    suggestions.push(' */');
    suggestions.push('');
    suggestions.push('함수/메서드 예시:');
    suggestions.push('/**');
    suggestions.push(' * Function description');
    suggestions.push(' * @param paramName Description');
    suggestions.push(' * @returns Return description');
    suggestions.push(' * @tags @API:FUNCTION-NAME-001');
    suggestions.push(' */');

    return suggestions.join('\n');
  }
}

/**
 * Main execution
 */
async function main() {
  try {
    let input = '';

    // Read input from stdin
    process.stdin.setEncoding('utf8');

    for await (const chunk of process.stdin) {
      input += chunk;
    }

    const parsedInput = input.trim() ? JSON.parse(input) : {};
    const enforcer = new TAGEnforcer();
    const result = await enforcer.execute(parsedInput);

    if (result.blocked) {
      console.error(`BLOCKED: ${result.message}`);
      if (result.suggestions) {
        console.error('\n📝 권장 @TAG 형식:\n' + result.suggestions);
      }
      process.exit(2);
    } else if (!result.success) {
      console.error(`ERROR: ${result.message}`);
      process.exit(result.exitCode || 1);
    } else if (result.message) {
      console.log(result.message);
    }

    process.exit(0);
  } catch (error) {
    console.error(`TAG Enforcer error: ${error.message}`);
    process.exit(0); // Don't block on errors
  }
}

// Export for testing
export { TAGEnforcer, main };

// Run if called directly (ES module equivalent)
if (import.meta.url === `file://${process.argv[1]}`) {
  main().catch(error => {
    console.error(`TAG Enforcer fatal error: ${error.message}`);
    process.exit(0);
  });
}