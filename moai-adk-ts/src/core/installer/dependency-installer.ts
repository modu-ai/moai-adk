// @CODE:INIT-001 | SPEC: SPEC-INIT-001.md | TEST: __tests__/core/installer/dependency-installer.test.ts
// Related: @CODE:CLI-INIT-001, @SPEC:INIT-001

/**
 * @file Automatic dependency installer for moai init
 * @author MoAI Team
 * @tags @CODE:INIT-001:INSTALLER
 */

import chalk from 'chalk';
import { execa } from 'execa';

/**
 * Automatic dependency installer for Git and Node.js
 * @description
 * Provides platform-specific automatic installation for missing dependencies:
 * - macOS: Uses Homebrew (brew install)
 * - Ubuntu/Linux: Uses apt (sudo apt install)
 * - Windows: Provides manual installation guide (winget or direct download)
 *
 * Features:
 * - nvm priority: Prefers nvm for Node.js installation (no sudo required)
 * - Timeout: 5 minutes per installation command
 * - Real-time output: Streams stdout/stderr during installation
 * - Graceful failure: Returns false on error with helpful messages
 */
export class DependencyInstaller {
  private readonly TIMEOUT_MS = 300000; // 5 minutes

  /**
   * Install Git on the current platform
   * @param platform - Platform identifier (darwin, linux, win32)
   * @returns True if installation succeeded, false otherwise
   */
  public async installGit(platform: string): Promise<boolean> {
    try {
      switch (platform) {
        case 'darwin':
          console.log(chalk.blue('Installing Git via Homebrew...'));
          await this.execWithTimeout('brew install git', this.TIMEOUT_MS);
          console.log(chalk.green('✅ Git installed successfully'));
          return true;

        case 'linux':
          console.log(chalk.blue('Installing Git via apt...'));
          await this.execWithTimeout(
            'sudo apt install git -y',
            this.TIMEOUT_MS
          );
          console.log(chalk.green('✅ Git installed successfully'));
          return true;

        case 'win32':
          console.log(
            chalk.yellow('⚠ Automatic installation not supported on Windows')
          );
          console.log(chalk.gray('Please install Git manually:'));
          console.log(chalk.white('  winget install Git.Git'));
          console.log(
            chalk.gray('Or download from: https://git-scm.com/download/win')
          );
          return false;

        default:
          console.log(chalk.yellow(`⚠ Unsupported platform: ${platform}`));
          console.log(chalk.gray('Please install Git manually'));
          return false;
      }
    } catch (error) {
      const errorMessage =
        error instanceof Error ? error.message : String(error);
      console.log(chalk.red(`❌ Git installation failed: ${errorMessage}`));
      console.log(chalk.gray('Please install Git manually'));
      return false;
    }
  }

  /**
   * Install Node.js on the current platform
   * @param platform - Platform identifier (darwin, linux, win32)
   * @returns True if installation succeeded, false otherwise
   * @description
   * Installation priority:
   * 1. Try nvm (if available) - no sudo required
   * 2. Fallback to platform package manager (brew/apt)
   * 3. Windows: Manual installation guide
   */
  public async installNodeJS(platform: string): Promise<boolean> {
    try {
      // Try nvm first (works on macOS, Linux, Windows with WSL)
      console.log(chalk.blue('Checking for nvm...'));
      const nvmInstalled = await this.checkNvmAvailable();

      if (nvmInstalled) {
        console.log(chalk.blue('Installing Node.js via nvm...'));
        await this.execWithTimeout('nvm install --lts', this.TIMEOUT_MS);
        console.log(chalk.green('✅ Node.js installed successfully via nvm'));

        // Verify installation
        const nodeVersion = await this.getNodeVersion();
        if (nodeVersion) {
          console.log(chalk.gray(`   Node.js version: ${nodeVersion}`));
          return true;
        }
      }

      // Fallback to platform package manager
      switch (platform) {
        case 'darwin':
          console.log(chalk.blue('Installing Node.js via Homebrew...'));
          await this.execWithTimeout('brew install node', this.TIMEOUT_MS);
          console.log(chalk.green('✅ Node.js installed successfully'));
          return true;

        case 'linux':
          console.log(chalk.blue('Installing Node.js via apt...'));
          await this.execWithTimeout(
            'sudo apt install nodejs -y',
            this.TIMEOUT_MS
          );
          console.log(chalk.green('✅ Node.js installed successfully'));
          return true;

        case 'win32':
          console.log(
            chalk.yellow('⚠ Automatic installation not supported on Windows')
          );
          console.log(chalk.gray('Please install Node.js manually:'));
          console.log(chalk.white('  winget install OpenJS.NodeJS.LTS'));
          console.log(chalk.gray('Or download from: https://nodejs.org/'));
          return false;

        default:
          console.log(chalk.yellow(`⚠ Unsupported platform: ${platform}`));
          console.log(chalk.gray('Please install Node.js manually'));
          return false;
      }
    } catch (error) {
      const errorMessage =
        error instanceof Error ? error.message : String(error);
      console.log(chalk.red(`❌ Node.js installation failed: ${errorMessage}`));
      console.log(chalk.gray('Please install Node.js manually'));
      return false;
    }
  }

  /**
   * Check if nvm is available
   * @returns True if nvm is installed, false otherwise
   */
  private async checkNvmAvailable(): Promise<boolean> {
    try {
      await execa('which', ['nvm'], { timeout: 5000 });
      return true;
    } catch {
      return false;
    }
  }

  /**
   * Get installed Node.js version
   * @returns Node.js version string or null if not installed
   */
  private async getNodeVersion(): Promise<string | null> {
    try {
      const { stdout } = await execa('node', ['--version'], { timeout: 5000 });
      return stdout.trim();
    } catch {
      return null;
    }
  }

  /**
   * Execute command with timeout and real-time output
   * @param command - Shell command to execute
   * @param timeout - Timeout in milliseconds
   * @throws Error if timeout is exceeded or command fails
   */
  private async execWithTimeout(
    command: string,
    timeout: number
  ): Promise<void> {
    const [cmd, ...args] = command.split(' ');

    try {
      const subprocess = execa(cmd, args, {
        timeout,
        reject: true,
        all: true,
      });

      // Stream stdout/stderr in real-time
      if (subprocess.stdout) {
        subprocess.stdout.on('data', data => {
          console.log(chalk.gray(data.toString().trim()));
        });
      }

      if (subprocess.stderr) {
        subprocess.stderr.on('data', data => {
          console.log(chalk.gray(data.toString().trim()));
        });
      }

      await subprocess;
    } catch (error: any) {
      if (error.timedOut) {
        throw new Error(`Installation timeout after ${timeout / 1000} seconds`);
      }
      throw error;
    }
  }
}
