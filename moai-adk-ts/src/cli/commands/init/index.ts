// @CODE:CLI-001 |
// Related: @CODE:INST-001:API, @CODE:PROMPT-001:UI, @CODE:CFG-001:DATA

/**
 * @file CLI init command orchestrator (delegates to handlers)
 * @author MoAI Team
 * @tags @CODE:INIT-001:ORCHESTRATOR
 */

import type { SystemDetector } from '@/core/system-checker/detector';
import type { InitResult, InitOptions } from '@/types/project';
import { DoctorCommand } from '../doctor';
import { handleInteractiveInit } from './interactive-handler';
import { handleNonInteractiveInit } from './non-interactive-handler';

/**
 * Initialize command for project setup
 * Delegates to interactive or non-interactive handlers
 */
export class InitCommand {
  private readonly doctorCommand: DoctorCommand;

  constructor(detector: SystemDetector) {
    this.doctorCommand = new DoctorCommand(detector);
  }

  /**
   * Run project initialization in interactive mode
   * @param options Init command options
   * @returns Complete initialization result
   */
  public async runInteractive(options?: InitOptions): Promise<InitResult> {
    return handleInteractiveInit(this.doctorCommand, options);
  }

  /**
   * Run project initialization in non-interactive mode
   * @param options Init command options
   * @returns Complete initialization result
   */
  public async runNonInteractive(options?: InitOptions): Promise<InitResult> {
    return handleNonInteractiveInit(this.doctorCommand, options);
  }

  /**
   * Run project initialization (legacy method)
   * @param projectName - Name of the project to initialize
   * @returns Success status
   */
  public async run(projectName?: string): Promise<boolean> {
    const options = projectName
      ? {
          name: projectName,
          mode: 'personal' as const,
        }
      : undefined;

    const result = await this.runInteractive(options);
    return result.success;
  }
}
