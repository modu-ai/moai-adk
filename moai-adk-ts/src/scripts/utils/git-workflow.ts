/**
 * @FEATURE-GIT-WORKFLOW-001: TypeScript Git 워크플로우 구현
 * @연결: @REQ:GIT-WORKFLOW-001 → @DESIGN:UNIFIED-GIT-001 → @TASK:GIT-TS-PORT-001
 */

import { exec } from 'node:child_process';
import { promisify } from 'node:util';

const execAsync = promisify(exec);

export interface CommandResult {
  stdout: string;
  stderr: string;
  exitCode: number;
}

export class GitWorkflowError extends Error {
  constructor(message: string) {
    super(message);
    this.name = 'GitWorkflowError';
  }
}

export class GitWorkflow {
  private projectRoot: string;

  constructor(projectRoot?: string) {
    this.projectRoot = projectRoot || process.cwd();
  }

  async runCommand(command: string[]): Promise<CommandResult> {
    try {
      const { stdout, stderr } = await execAsync(command.join(' '), {
        cwd: this.projectRoot,
      });

      return {
        stdout: stdout || '',
        stderr: stderr || '',
        exitCode: 0,
      };
    } catch (error: any) {
      if (error.code !== undefined) {
        return {
          stdout: error.stdout || '',
          stderr: error.stderr || '',
          exitCode: error.code,
        };
      }
      throw new GitWorkflowError(error.message);
    }
  }

  async createConstitutionCommit(
    message: string,
    files?: string[]
  ): Promise<string> {
    try {
      // 파일 추가
      if (files && files.length > 0) {
        for (const file of files) {
          await this.runCommand(['git', 'add', file]);
        }
      } else {
        await this.runCommand(['git', 'add', '.']);
      }

      // 커밋 생성
      await this.runCommand(['git', 'commit', '-m', `"${message}"`]);

      // 커밋 해시 추출
      const hashResult = await this.runCommand(['git', 'rev-parse', 'HEAD']);
      return hashResult.stdout.trim();
    } catch (error) {
      throw new GitWorkflowError(
        `커밋 생성 실패: ${error instanceof Error ? error.message : String(error)}`
      );
    }
  }
}
