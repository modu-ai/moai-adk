// @CODE:GIT-004 | SPEC: SPEC-GIT-001.md | TEST: src/core/git/workflow/__tests__/workflow-automation.test.ts
// Related: @CODE:GIT-004:DATA

/**
 * @file SPEC Structure Step
 * @author MoAI Team
 *
 * @fileoverview SPEC 디렉토리 구조 생성 단계
 */

import path from 'node:path';
import fs from 'fs-extra';

/**
 * SPEC 구조 생성 단계
 */
export class SpecStructureStep {
  /**
   * SPEC 디렉토리 구조 생성
   */
  async createSpecStructure(
    specId: string,
    description: string
  ): Promise<void> {
    const specDir = path.join(process.cwd(), '.moai', 'specs', specId);
    await fs.ensureDir(specDir);

    await this.createSpecFiles(specDir, specId, description);
  }

  /**
   * SPEC 파일들 생성
   */
  private async createSpecFiles(
    specDir: string,
    specId: string,
    description: string
  ): Promise<void> {
    const specContent = this.generateSpecContent(specId, description);
    const planContent = '# Implementation Plan\n\nTBD\n';
    const acceptanceContent = '# Acceptance Criteria\n\nTBD\n';

    await fs.writeFile(path.join(specDir, 'spec.md'), specContent);
    await fs.writeFile(path.join(specDir, 'plan.md'), planContent);
    await fs.writeFile(path.join(specDir, 'acceptance.md'), acceptanceContent);
  }

  /**
   * SPEC 내용 생성
   */
  private generateSpecContent(specId: string, description: string): string {
    return `# ${specId} Specification

## Description
${description}

## Requirements
- [ ] Requirement 1
- [ ] Requirement 2

## Implementation Plan
- [ ] Step 1
- [ ] Step 2

## Test Plan
- [ ] Test case 1
- [ ] Test case 2
`;
  }
}
