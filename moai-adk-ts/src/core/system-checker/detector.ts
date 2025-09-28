/**
 * @file System requirement detector and validator
 * @author MoAI Team
 * @tags @FEATURE:SYSTEM-DETECTOR-001 @REQ:AUTO-VERIFY-012
 */

import execa from 'execa';
import * as semver from 'semver';
import * as os from 'os';
import { SystemRequirement } from './requirements';

/**
 * Detection result for a single requirement
 * @tags @DESIGN:DETECTION-RESULT-001
 */
export interface DetectionResult {
  readonly isInstalled: boolean;
  readonly detectedVersion?: string;
  readonly versionSatisfied: boolean;
  readonly error?: string;
}

/**
 * Combined requirement and detection result
 * @tags @DESIGN:CHECK-RESULT-001
 */
export interface RequirementCheckResult {
  readonly requirement: SystemRequirement;
  readonly result: DetectionResult;
}

/**
 * System requirement detector
 * Automatically detects installed tools and validates versions
 * @tags @FEATURE:SYSTEM-DETECTOR-001
 */
export class SystemDetector {
  /**
   * Check single system requirement
   * @param requirement - System requirement to check
   * @returns Detection result
   * @tags @API:CHECK-REQUIREMENT-001
   */
  public async checkRequirement(
    requirement: SystemRequirement
  ): Promise<DetectionResult> {
    try {
      // Execute check command
      const execResult = await execa(
        requirement.checkCommand.split(' ')[0]!,
        requirement.checkCommand.split(' ').slice(1),
        {
          timeout: 10000,
          reject: false,
        }
      );

      if (execResult.exitCode !== 0) {
        return {
          isInstalled: false,
          versionSatisfied: false,
          error:
            execResult.stderr ||
            `Command failed with exit code ${execResult.exitCode}`,
        };
      }

      // Extract version from output
      const detectedVersion = this.extractVersion(execResult.stdout);

      // Check version satisfaction
      const versionSatisfied = requirement.minVersion
        ? detectedVersion
          ? semver.gte(detectedVersion, requirement.minVersion)
          : false
        : true;

      const detectionResult: DetectionResult = {
        isInstalled: true,
        versionSatisfied,
      };

      if (detectedVersion) {
        (
          detectionResult as DetectionResult & { detectedVersion: string }
        ).detectedVersion = detectedVersion;
      }

      return detectionResult;
    } catch (error) {
      return {
        isInstalled: false,
        versionSatisfied: false,
        error: error instanceof Error ? error.message : 'Unknown error',
      };
    }
  }

  /**
   * Check multiple requirements concurrently
   * @param requirements - Array of system requirements
   * @returns Array of check results
   * @tags @API:CHECK-MULTIPLE-001
   */
  public async checkMultipleRequirements(
    requirements: SystemRequirement[]
  ): Promise<RequirementCheckResult[]> {
    if (requirements.length === 0) {
      return [];
    }

    const checkPromises = requirements.map(async requirement => ({
      requirement,
      result: await this.checkRequirement(requirement),
    }));

    return Promise.all(checkPromises);
  }

  /**
   * Get current platform
   * @returns Platform identifier
   * @tags @API:GET-PLATFORM-001
   */
  public getCurrentPlatform(): 'darwin' | 'linux' | 'win32' {
    const platform = os.platform();
    if (platform === 'darwin' || platform === 'linux' || platform === 'win32') {
      return platform;
    }
    // Default to linux for unknown platforms
    return 'linux';
  }

  /**
   * Get install command for current platform
   * @param requirement - System requirement
   * @returns Install command or undefined
   * @tags @API:GET-INSTALL-COMMAND-001
   */
  public getInstallCommandForCurrentPlatform(
    requirement: SystemRequirement
  ): string | undefined {
    const platform = this.getCurrentPlatform();
    return requirement.installCommands[platform];
  }

  /**
   * Extract version from command output
   * @param output - Command output string
   * @returns Extracted version or undefined
   * @tags @UTIL:EXTRACT-VERSION-001
   */
  private extractVersion(output: string): string | undefined {
    // Common version patterns
    const patterns = [
      /v?(\d+\.\d+\.\d+(?:\.\d+)?)/, // Standard semver
      /version\s+v?(\d+\.\d+\.\d+(?:\.\d+)?)/i, // "version X.Y.Z"
      /v?(\d+\.\d+\.\d+(?:-[a-zA-Z0-9]+)?)/, // With pre-release
    ];

    for (const pattern of patterns) {
      const match = output.match(pattern);
      if (match?.[1]) {
        const version = match[1];
        // Validate with semver
        if (semver.valid(version)) {
          return version;
        }
      }
    }

    return undefined;
  }
}
