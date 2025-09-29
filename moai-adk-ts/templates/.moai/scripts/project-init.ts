#!/usr/bin/env tsx
// @FEATURE-PROJECT-INIT-001: 프로젝트 초기화 스크립트
// 연결: @REQ-PROJECT-001 → @DESIGN-INIT-001 → @TASK-INIT-001

import { program } from 'commander';
import { promises as fs } from 'fs';
import path from 'path';
import chalk from 'chalk';

interface ProjectInitOptions {
  name?: string;
  type?: 'personal' | 'team';
  language?: string;
  template?: string;
}

interface ProjectConfig {
  project: {
    name: string;
    description: string;
    mode: 'personal' | 'team';
    version: string;
    created_at: string;
    initialized: boolean;
  };
  constitution: {
    enforce_tdd: boolean;
    require_tags: boolean;
    test_coverage_target: number;
    simplicity_threshold: number;
  };
  pipeline: {
    current_stage: string;
    available_commands: string[];
  };
}

async function initializeProject(options: ProjectInitOptions): Promise<{ success: boolean; message: string; config?: ProjectConfig }> {
  try {
    const projectName = options.name || path.basename(process.cwd());
    const projectType = options.type || 'personal';

    // MoAI 디렉토리 구조 생성
    const moaiDirs = [
      '.moai',
      '.moai/specs',
      '.moai/indexes',
      '.moai/reports',
      '.moai/checkpoints',
      '.moai/memory',
      '.moai/project',
      '.moai/scripts'
    ];

    for (const dir of moaiDirs) {
      await fs.mkdir(dir, { recursive: true });
    }

    // 프로젝트 설정 생성
    const config: ProjectConfig = {
      project: {
        name: projectName,
        description: `${projectName} - MoAI 프로젝트`,
        mode: projectType,
        version: '0.1.0',
        created_at: new Date().toISOString(),
        initialized: true
      },
      constitution: {
        enforce_tdd: true,
        require_tags: true,
        test_coverage_target: 85,
        simplicity_threshold: 5
      },
      pipeline: {
        current_stage: 'initialized',
        available_commands: [
          '/moai:0-project',
          '/moai:1-spec',
          '/moai:2-build',
          '/moai:3-sync'
        ]
      }
    };

    // config.json 작성
    await fs.writeFile(
      '.moai/config.json',
      JSON.stringify(config, null, 2)
    );

    // 기본 태그 인덱스 초기화
    const tagIndex = {
      version: '1.0.0',
      tags: {},
      indexes: {
        byType: {},
        byCategory: {},
        byStatus: {},
        byFile: {}
      },
      metadata: {
        totalTags: 0,
        lastUpdated: new Date().toISOString()
      }
    };

    await fs.writeFile(
      '.moai/indexes/tags.json',
      JSON.stringify(tagIndex, null, 2)
    );

    return {
      success: true,
      message: `프로젝트 '${projectName}' 초기화 완료`,
      config
    };

  } catch (error) {
    return {
      success: false,
      message: `프로젝트 초기화 실패: ${error.message}`
    };
  }
}

program
  .name('project-init')
  .description('MoAI 프로젝트 초기화')
  .option('-n, --name <name>', '프로젝트 이름')
  .option('-t, --type <type>', '프로젝트 타입 (personal|team)', 'personal')
  .option('-l, --language <language>', '주 사용 언어')
  .option('--template <template>', '프로젝트 템플릿')
  .action(async (options: ProjectInitOptions) => {
    try {
      console.log(chalk.blue('🗿 MoAI 프로젝트 초기화 시작...'));

      const result = await initializeProject(options);

      if (result.success) {
        console.log(chalk.green('✅'), result.message);
        console.log(JSON.stringify({
          success: true,
          project: result.config?.project,
          nextSteps: [
            'moai 1-spec로 첫 번째 SPEC 생성',
            'moai 2-build로 TDD 구현',
            'moai 3-sync로 문서 동기화'
          ]
        }, null, 2));
        process.exit(0);
      } else {
        console.error(chalk.red('❌'), result.message);
        console.log(JSON.stringify({
          success: false,
          error: result.message
        }, null, 2));
        process.exit(1);
      }
    } catch (error) {
      console.error(chalk.red('❌ 스크립트 실행 실패:'), error.message);
      console.log(JSON.stringify({
        success: false,
        error: error.message
      }, null, 2));
      process.exit(1);
    }
  });

program.parse();