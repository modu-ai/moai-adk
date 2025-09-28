/**
 * @FEATURE:SYSTEM-DETECTOR-001 System tool detection and validation
 *
 * Detects installed system tools and validates their versions against
 * minimum requirements for MoAI-ADK functionality.
 */

import { exec } from 'child_process';
import { promisify } from 'util';
import semver from 'semver';
import os from 'os';
import chalk from 'chalk';
import which from 'which';
import type { SystemRequirement } from './requirements.js';

const execAsync = promisify(exec);

export interface DetectionResult {
  name: string;
  installed: boolean;
  version?: string;
  meetsMinimum: boolean;
  installSuggestion?: string;
  error?: string;
  required: boolean;
  category: SystemRequirement['category'];
}

export class SystemDetector {
  private platform = os.platform();
  private timeoutMs = 10000; // 10 second timeout for checks

  async detectRequirement(requirement: SystemRequirement): Promise<DetectionResult> {
    try {
      // First check if the command exists using 'which'
      const commandExists = await this.checkCommandExists(requirement.checkCommand);
      if (!commandExists) {
        return {
          name: requirement.name,
          installed: false,
          meetsMinimum: false,
          installSuggestion: this.getInstallCommand(requirement),
          error: 'Command not found',
          required: requirement.required,
          category: requirement.category
        };
      }

      // Execute version check command
      const result = await this.executeCheck(requirement.checkCommand);
      const version = this.extractVersion(result.stdout, requirement);
      const meetsMinimum = requirement.minVersion && version
        ? semver.gte(version, requirement.minVersion)
        : true;

      return {
        name: requirement.name,
        installed: true,
        version,
        meetsMinimum,
        installSuggestion: meetsMinimum ? undefined : this.getInstallCommand(requirement),
        required: requirement.required,
        category: requirement.category
      };
    } catch (error) {
      return {
        name: requirement.name,
        installed: false,
        meetsMinimum: false,
        installSuggestion: this.getInstallCommand(requirement),
        error: error instanceof Error ? error.message : 'Detection failed',
        required: requirement.required,
        category: requirement.category
      };
    }
  }

  private async checkCommandExists(command: string): Promise<boolean> {
    try {
      const commandName = command.split(' ')[0]; // Extract base command
      await which(commandName);
      return true;
    } catch {
      return false;
    }
  }

  private async executeCheck(command: string): Promise<{ stdout: string; stderr: string }> {
    return await execAsync(command, {
      timeout: this.timeoutMs,
      encoding: 'utf8'
    });
  }

  private extractVersion(output: string, requirement: SystemRequirement): string | undefined {
    // Use custom parser if provided
    if (requirement.versionParser) {
      return requirement.versionParser(output);
    }

    // Generic version extraction patterns
    const patterns = [
      /v?(\d+\.\d+\.\d+)/,           // Generic semver
      /version (\d+\.\d+\.\d+)/,     // "version X.Y.Z"
      /(\d+\.\d+\.\d+)/              // Bare semver
    ];

    for (const pattern of patterns) {
      const match = output.match(pattern);
      if (match) {
        return match[1];
      }
    }

    return undefined;
  }

  private getInstallCommand(requirement: SystemRequirement): string | undefined {
    return requirement.installCommand?.[this.platform];
  }

  async detectAll(requirements: SystemRequirement[]): Promise<DetectionResult[]> {
    const results: DetectionResult[] = [];

    console.log(chalk.cyan('🔍 Checking system requirements...'));

    for (const requirement of requirements) {
      process.stdout.write(chalk.gray(`  Checking ${requirement.name}... `));

      const result = await this.detectRequirement(requirement);

      if (result.installed && result.meetsMinimum) {
        console.log(chalk.green(`✅ v${result.version || 'unknown'}`));
      } else if (result.installed && !result.meetsMinimum) {
        console.log(chalk.yellow(`⚠️  v${result.version} (requires ${requirement.minVersion}+)`));
      } else {
        console.log(chalk.red('❌ Not found'));
      }

      results.push(result);
    }

    return results;
  }

  async detectQuick(requirements: SystemRequirement[]): Promise<DetectionResult[]> {
    // Parallel detection for faster results
    const promises = requirements.map(req => this.detectRequirement(req));
    return await Promise.all(promises);
  }

  generateSystemReport(results: DetectionResult[]): string {
    const report: string[] = [];

    report.push(chalk.cyan('📋 System Requirements Report'));
    report.push(chalk.cyan('=' .repeat(50)));

    const categories = ['runtime', 'development', 'optional'] as const;

    for (const category of categories) {
      const categoryResults = results.filter(r => r.category === category);
      if (categoryResults.length === 0) continue;

      report.push(chalk.white(`\n${category.toUpperCase()} TOOLS:`));

      for (const result of categoryResults) {
        const status = result.installed && result.meetsMinimum
          ? chalk.green('✅')
          : result.installed
            ? chalk.yellow('⚠️ ')
            : chalk.red('❌');

        const version = result.version ? ` (v${result.version})` : '';
        const required = result.required ? chalk.red(' [REQUIRED]') : chalk.gray(' [OPTIONAL]');

        report.push(`  ${status} ${result.name}${version}${required}`);

        if (result.installSuggestion && (!result.installed || !result.meetsMinimum)) {
          report.push(chalk.gray(`     Install: ${result.installSuggestion}`));
        }
      }
    }

    return report.join('\n');
  }
}