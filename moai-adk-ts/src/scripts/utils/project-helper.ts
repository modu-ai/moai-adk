/**
 * @FEATURE-PROJECT-HELPER-001: TypeScript ProjectHelper 구현
 * @연결: @TASK:PROJECT-UTILS-001 → @FEATURE:PROJECT-MANAGEMENT-001 → @API:PROJECT-HELPER
 */

import * as fs from 'node:fs';
import * as path from 'node:path';

export interface ProjectConfig {
  mode?: string;
  project_type?: string;
  languages?: {
    backend?: string;
    frontend?: string;
  };
  project_language?: string;
  test_framework?: string;
  linter?: string;
  formatter?: string;
  [key: string]: any;
}

export class ProjectHelper {
  static loadConfig(projectRoot: string): ProjectConfig {
    const configPath = path.join(projectRoot, '.moai', 'config.json');

    try {
      if (fs.existsSync(configPath)) {
        const configContent = fs.readFileSync(configPath, 'utf-8');
        return JSON.parse(configContent);
      }
    } catch (error) {
      console.warn(
        `Config 로드 실패: ${error instanceof Error ? error.message : String(error)}`
      );
    }

    // 기본 설정 반환
    return {
      mode: 'personal',
      project_type: 'single',
      project_language: 'typescript',
    };
  }

  static findProjectRoot(startPath?: string): string {
    let currentDir = startPath || process.cwd();
    const root = path.parse(currentDir).root;

    while (currentDir !== root) {
      try {
        const gitDir = path.join(currentDir, '.git');
        if (fs.existsSync(gitDir)) {
          return currentDir;
        }

        const moaiDir = path.join(currentDir, '.moai');
        if (fs.existsSync(moaiDir)) {
          return currentDir;
        }
      } catch (_error) {
        // 무시하고 계속
      }
      currentDir = path.dirname(currentDir);
    }

    // 프로젝트 루트를 찾지 못하면 현재 디렉토리 반환
    return startPath || process.cwd();
  }

  static saveConfig(projectRoot: string, config: ProjectConfig): void {
    const configPath = path.join(projectRoot, '.moai', 'config.json');
    const moaiDir = path.dirname(configPath);

    try {
      // .moai 디렉토리 생성
      if (!fs.existsSync(moaiDir)) {
        fs.mkdirSync(moaiDir, { recursive: true });
      }

      // 설정 저장
      fs.writeFileSync(configPath, JSON.stringify(config, null, 2), 'utf-8');
    } catch (error) {
      throw new Error(
        `Config 저장 실패: ${error instanceof Error ? error.message : String(error)}`
      );
    }
  }

  static detectProjectType(projectRoot: string): string {
    const packageJsonPath = path.join(projectRoot, 'package.json');
    const hasPyprojectToml = fs.existsSync(
      path.join(projectRoot, 'pyproject.toml')
    );
    const hasSetupPy = fs.existsSync(path.join(projectRoot, 'setup.py'));

    let hasUserPackageJson = false;

    // Check if package.json exists and is not a MoAI-ADK package
    if (fs.existsSync(packageJsonPath)) {
      try {
        const packageData = JSON.parse(
          fs.readFileSync(packageJsonPath, 'utf-8')
        );
        // Exclude MoAI-ADK package.json from user project detection
        if (
          !(
            packageData.name === 'moai-adk' ||
            packageData.description?.includes('MoAI-ADK')
          )
        ) {
          hasUserPackageJson = true;
          console.log('Detected user package.json (not MoAI-ADK)');
        } else {
          console.log(
            'Skipping MoAI-ADK package.json in project type detection'
          );
        }
      } catch (error) {
        console.warn(
          'Could not parse package.json for project type detection:',
          error
        );
      }
    }

    if (hasUserPackageJson && (hasPyprojectToml || hasSetupPy)) {
      return 'fullstack';
    }

    if (hasUserPackageJson) {
      return 'typescript';
    }

    if (hasPyprojectToml || hasSetupPy) {
      return 'python';
    }

    return 'unknown';
  }
}
