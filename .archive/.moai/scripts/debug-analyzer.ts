#!/usr/bin/env tsx
// CODE-DEBUG-ANALYZER-001: 디버깅 분석 및 문제 진단 스크립트
// 연결: SPEC-DEBUG-001 → SPEC-DEBUG-ANALYZER-001 → CODE-DEBUG-001

import { program } from 'commander';
import { promises as fs } from 'fs';
import path from 'path';
import chalk from 'chalk';
import { execa } from 'execa';

interface DebugAnalyzerOptions {
  error?: string;
  logs?: string;
  system?: boolean;
  performance?: boolean;
  dependencies?: boolean;
  interactive?: boolean;
  fix?: boolean;
}

interface DebugIssue {
  category: 'error' | 'warning' | 'performance' | 'dependency' | 'system';
  severity: 'critical' | 'high' | 'medium' | 'low';
  title: string;
  description: string;
  context?: string;
  solution: string;
  autofix?: boolean;
  commands?: string[];
}

interface SystemDiagnostic {
  nodejs: { version: string; compatible: boolean };
  git: { version: string; available: boolean };
  packageManager: { type: string; version: string; available: boolean };
  disk: { available: string; usage: string };
  memory: { total: string; available: string };
}

interface PerformanceMetric {
  metric: string;
  current: number;
  threshold: number;
  unit: string;
  status: 'ok' | 'warning' | 'critical';
}

interface DependencyIssue {
  name: string;
  current?: string;
  required: string;
  status: 'missing' | 'outdated' | 'incompatible' | 'vulnerable';
  severity: 'low' | 'medium' | 'high' | 'critical';
}

interface DebugAnalysisResult {
  issues: DebugIssue[];
  diagnostics: {
    system?: SystemDiagnostic;
    performance?: PerformanceMetric[];
    dependencies?: DependencyIssue[];
  };
  recommendations: string[];
  summary: {
    criticalIssues: number;
    fixableIssues: number;
    totalIssues: number;
  };
}

async function analyzeErrorMessage(errorMessage: string): Promise<DebugIssue[]> {
  const issues: DebugIssue[] = [];

  // 일반적인 에러 패턴들
  const errorPatterns = [
    {
      pattern: /ENOENT.*no such file or directory/i,
      category: 'error' as const,
      severity: 'high' as const,
      title: '파일 또는 디렉토리를 찾을 수 없음',
      solution: '파일 경로를 확인하고 필요한 파일이 존재하는지 확인하세요',
      autofix: false
    },
    {
      pattern: /EACCES.*permission denied/i,
      category: 'error' as const,
      severity: 'high' as const,
      title: '권한 부족',
      solution: '파일 권한을 확인하거나 sudo 권한이 필요할 수 있습니다',
      commands: ['chmod +x filename', 'sudo command'],
      autofix: false
    },
    {
      pattern: /Cannot find module ['"]([^'"]+)['"]/i,
      category: 'dependency' as const,
      severity: 'high' as const,
      title: '모듈을 찾을 수 없음',
      solution: 'npm install 또는 yarn install을 실행하여 의존성을 설치하세요',
      commands: ['npm install', 'yarn install', 'bun install'],
      autofix: true
    },
    {
      pattern: /Port \d+ is already in use/i,
      category: 'system' as const,
      severity: 'medium' as const,
      title: '포트가 이미 사용 중',
      solution: '다른 포트를 사용하거나 해당 포트를 사용하는 프로세스를 종료하세요',
      commands: ['lsof -ti:PORT | xargs kill -9', 'netstat -tulpn | grep :PORT'],
      autofix: false
    },
    {
      pattern: /SyntaxError.*Unexpected token/i,
      category: 'error' as const,
      severity: 'critical' as const,
      title: '구문 오류',
      solution: '코드 구문을 확인하고 누락된 괄호, 세미콜론 등을 찾아 수정하세요',
      autofix: false
    },
    {
      pattern: /TypeError.*is not a function/i,
      category: 'error' as const,
      severity: 'high' as const,
      title: '타입 오류 - 함수가 아님',
      solution: '변수가 올바른 타입인지 확인하고 함수 호출 전에 타입 검사를 추가하세요',
      autofix: false
    },
    {
      pattern: /ReferenceError.*is not defined/i,
      category: 'error' as const,
      severity: 'high' as const,
      title: '참조 오류 - 정의되지 않은 변수',
      solution: '변수가 올바르게 선언되었는지 확인하고 스코프를 검토하세요',
      autofix: false
    },
    {
      pattern: /TimeoutError.*exceeded/i,
      category: 'performance' as const,
      severity: 'medium' as const,
      title: '타임아웃 발생',
      solution: '네트워크 연결을 확인하거나 타임아웃 값을 증가시키세요',
      autofix: false
    },
    {
      pattern: /EMFILE.*too many open files/i,
      category: 'system' as const,
      severity: 'high' as const,
      title: '열린 파일 수 제한 초과',
      solution: 'ulimit을 증가시키거나 파일 핸들 누수를 확인하세요',
      commands: ['ulimit -n 65536'],
      autofix: false
    },
    {
      pattern: /npm ERR!.*ERESOLVE/i,
      category: 'dependency' as const,
      severity: 'medium' as const,
      title: '의존성 해결 충돌',
      solution: 'npm install --legacy-peer-deps 또는 yarn resolutions을 사용하세요',
      commands: ['npm install --legacy-peer-deps', 'npm audit fix'],
      autofix: true
    }
  ];

  for (const { pattern, category, severity, title, solution, commands, autofix } of errorPatterns) {
    if (pattern.test(errorMessage)) {
      const match = errorMessage.match(pattern);
      issues.push({
        category,
        severity,
        title,
        description: `감지된 패턴: ${match?.[0] || '알 수 없음'}`,
        context: errorMessage.slice(0, 200),
        solution,
        commands,
        autofix: autofix || false
      });
    }
  }

  // 스택 트레이스 분석
  const stackTrace = errorMessage.match(/at\s+(.+)\s+\((.+):(\d+):(\d+)\)/g);
  if (stackTrace && stackTrace.length > 0) {
    issues.push({
      category: 'error',
      severity: 'medium',
      title: '스택 트레이스 분석',
      description: `${stackTrace.length}개의 스택 프레임 발견`,
      context: stackTrace.slice(0, 3).join('\n'),
      solution: '스택 트레이스를 따라 오류 발생 지점을 추적하세요',
      autofix: false
    });
  }

  return issues;
}

async function runSystemDiagnostics(): Promise<SystemDiagnostic> {
  const diagnostic: SystemDiagnostic = {
    nodejs: { version: '', compatible: false },
    git: { version: '', available: false },
    packageManager: { type: '', version: '', available: false },
    disk: { available: '', usage: '' },
    memory: { total: '', available: '' }
  };

  // Node.js 버전 확인
  try {
    const result = await execa('node', ['--version'], { stdio: 'pipe' });
    diagnostic.nodejs.version = result.stdout.trim();
    const majorVersion = parseInt(diagnostic.nodejs.version.slice(1).split('.')[0]);
    diagnostic.nodejs.compatible = majorVersion >= 18;
  } catch {
    diagnostic.nodejs.version = 'Not installed';
  }

  // Git 확인
  try {
    const result = await execa('git', ['--version'], { stdio: 'pipe' });
    diagnostic.git.version = result.stdout.trim();
    diagnostic.git.available = true;
  } catch {
    diagnostic.git.version = 'Not installed';
  }

  // 패키지 매니저 확인
  const packageManagers = ['bun', 'npm', 'yarn'];
  for (const pm of packageManagers) {
    try {
      const result = await execa(pm, ['--version'], { stdio: 'pipe' });
      diagnostic.packageManager = {
        type: pm,
        version: result.stdout.trim(),
        available: true
      };
      break;
    } catch {
      // 다음 패키지 매니저 시도
    }
  }

  // 디스크 사용량 확인
  try {
    const result = await execa('df', ['-h', '.'], { stdio: 'pipe' });
    const lines = result.stdout.split('\n');
    if (lines.length > 1) {
      const fields = lines[1].split(/\s+/);
      diagnostic.disk.available = fields[3] || 'Unknown';
      diagnostic.disk.usage = fields[4] || 'Unknown';
    }
  } catch {
    // Windows 또는 df 명령어 없음
    diagnostic.disk.available = 'Unknown';
    diagnostic.disk.usage = 'Unknown';
  }

  // 메모리 사용량 확인
  try {
    if (process.platform === 'darwin' || process.platform === 'linux') {
      const result = await execa('free', ['-h'], { stdio: 'pipe' });
      const lines = result.stdout.split('\n');
      if (lines.length > 1) {
        const fields = lines[1].split(/\s+/);
        diagnostic.memory.total = fields[1] || 'Unknown';
        diagnostic.memory.available = fields[6] || 'Unknown';
      }
    } else {
      // Node.js 메모리 정보 사용
      const memUsage = process.memoryUsage();
      diagnostic.memory.total = `${Math.round(memUsage.heapTotal / 1024 / 1024)}MB`;
      diagnostic.memory.available = `${Math.round((memUsage.heapTotal - memUsage.heapUsed) / 1024 / 1024)}MB`;
    }
  } catch {
    diagnostic.memory.total = 'Unknown';
    diagnostic.memory.available = 'Unknown';
  }

  return diagnostic;
}

async function analyzePerformance(): Promise<PerformanceMetric[]> {
  const metrics: PerformanceMetric[] = [];

  // 메모리 사용량
  const memUsage = process.memoryUsage();
  metrics.push({
    metric: 'Heap Used',
    current: Math.round(memUsage.heapUsed / 1024 / 1024),
    threshold: 512,
    unit: 'MB',
    status: memUsage.heapUsed / 1024 / 1024 > 512 ? 'warning' : 'ok'
  });

  // 힙 사용률
  const heapUsagePercent = (memUsage.heapUsed / memUsage.heapTotal) * 100;
  metrics.push({
    metric: 'Heap Usage',
    current: Math.round(heapUsagePercent),
    threshold: 80,
    unit: '%',
    status: heapUsagePercent > 90 ? 'critical' : heapUsagePercent > 80 ? 'warning' : 'ok'
  });

  // 프로세스 업타임
  const uptimeMinutes = Math.round(process.uptime() / 60);
  metrics.push({
    metric: 'Process Uptime',
    current: uptimeMinutes,
    threshold: 60,
    unit: 'minutes',
    status: 'ok'
  });

  // CPU 사용량 (근사값)
  const startTime = process.hrtime();
  await new Promise(resolve => setTimeout(resolve, 100));
  const endTime = process.hrtime(startTime);
  const cpuUsage = process.cpuUsage();
  const cpuPercent = (cpuUsage.user + cpuUsage.system) / 1000000 / (endTime[0] + endTime[1] / 1e9) * 100;

  metrics.push({
    metric: 'CPU Usage',
    current: Math.round(cpuPercent),
    threshold: 80,
    unit: '%',
    status: cpuPercent > 90 ? 'critical' : cpuPercent > 80 ? 'warning' : 'ok'
  });

  return metrics;
}

async function analyzeDependencies(): Promise<DependencyIssue[]> {
  const issues: DependencyIssue[] = [];

  try {
    // package.json 읽기
    const packageJson = JSON.parse(await fs.readFile('package.json', 'utf-8'));
    const dependencies = { ...packageJson.dependencies, ...packageJson.devDependencies };

    // npm audit 실행 (가능한 경우)
    try {
      const auditResult = await execa('npm', ['audit', '--json'], { stdio: 'pipe' });
      const auditData = JSON.parse(auditResult.stdout);

      if (auditData.vulnerabilities) {
        for (const [name, vuln] of Object.entries(auditData.vulnerabilities as any)) {
          issues.push({
            name,
            current: dependencies[name],
            required: vuln.via?.[0]?.range || 'Latest',
            status: 'vulnerable',
            severity: vuln.severity || 'medium'
          });
        }
      }
    } catch {
      // npm audit 실행 실패는 무시
    }

    // 주요 의존성 버전 확인
    const criticalDeps = {
      'typescript': '>=5.0.0',
      'node': '>=18.0.0',
      '@types/node': '>=18.0.0',
      'vitest': '>=1.0.0',
      'jest': '>=29.0.0'
    };

    for (const [depName, requiredVersion] of Object.entries(criticalDeps)) {
      if (dependencies[depName]) {
        const currentVersion = dependencies[depName];
        // 간단한 버전 비교 (실제로는 semver 라이브러리 사용 권장)
        if (currentVersion.includes('*') || currentVersion.includes('^') || currentVersion.includes('~')) {
          // 동적 버전은 OK로 간주
          continue;
        }

        issues.push({
          name: depName,
          current: currentVersion,
          required: requiredVersion,
          status: 'outdated',
          severity: 'medium'
        });
      }
    }

    // 필수 개발 도구 확인
    const requiredDevTools = ['typescript', '@types/node'];
    for (const tool of requiredDevTools) {
      if (!dependencies[tool] && !packageJson.dependencies?.[tool]) {
        issues.push({
          name: tool,
          required: 'latest',
          status: 'missing',
          severity: 'high'
        });
      }
    }

  } catch (error) {
    issues.push({
      name: 'package.json',
      required: 'valid package.json',
      status: 'missing',
      severity: 'critical'
    });
  }

  return issues;
}

async function parseLogFile(logPath: string): Promise<DebugIssue[]> {
  const issues: DebugIssue[] = [];

  try {
    const logContent = await fs.readFile(logPath, 'utf-8');
    const lines = logContent.split('\n');

    let errorCount = 0;
    let warningCount = 0;

    for (let i = 0; i < lines.length; i++) {
      const line = lines[i];

      // 에러 로그 패턴
      if (/ERROR|FATAL|CRITICAL/i.test(line)) {
        errorCount++;
        issues.push({
          category: 'error',
          severity: 'high',
          title: `로그 에러 (라인 ${i + 1})`,
          description: line.trim(),
          solution: '로그에서 에러 원인을 분석하고 해당 코드를 수정하세요',
          autofix: false
        });
      }

      // 경고 로그 패턴
      if (/WARN|WARNING/i.test(line)) {
        warningCount++;
        if (warningCount <= 5) { // 최대 5개만 표시
          issues.push({
            category: 'warning',
            severity: 'medium',
            title: `로그 경고 (라인 ${i + 1})`,
            description: line.trim(),
            solution: '경고 메시지를 검토하고 필요시 수정하세요',
            autofix: false
          });
        }
      }

      // 성능 관련 로그
      if (/slow|timeout|took.*ms|exceeded/i.test(line)) {
        issues.push({
          category: 'performance',
          severity: 'medium',
          title: `성능 이슈 (라인 ${i + 1})`,
          description: line.trim(),
          solution: '성능 병목 지점을 분석하고 최적화하세요',
          autofix: false
        });
      }
    }

    // 요약 정보 추가
    if (errorCount > 0 || warningCount > 0) {
      issues.unshift({
        category: 'warning',
        severity: 'medium',
        title: '로그 파일 요약',
        description: `총 ${errorCount}개 에러, ${warningCount}개 경고 발견`,
        solution: '로그 파일의 모든 에러와 경고를 검토하세요',
        autofix: false
      });
    }

  } catch (error) {
    issues.push({
      category: 'error',
      severity: 'high',
      title: '로그 파일 읽기 실패',
      description: `로그 파일을 읽을 수 없습니다: ${error.message}`,
      solution: '로그 파일 경로와 권한을 확인하세요',
      autofix: false
    });
  }

  return issues;
}

async function generateRecommendations(result: DebugAnalysisResult): Promise<string[]> {
  const recommendations: string[] = [];

  // 시스템 진단 기반 권장사항
  if (result.diagnostics.system) {
    const sys = result.diagnostics.system;

    if (!sys.nodejs.compatible) {
      recommendations.push('Node.js를 18.0.0 이상으로 업그레이드하세요');
    }

    if (!sys.git.available) {
      recommendations.push('Git을 설치하세요');
    }

    if (!sys.packageManager.available) {
      recommendations.push('npm, yarn, 또는 bun 패키지 매니저를 설치하세요');
    }
  }

  // 성능 메트릭 기반 권장사항
  if (result.diagnostics.performance) {
    const criticalMetrics = result.diagnostics.performance.filter(m => m.status === 'critical');
    if (criticalMetrics.length > 0) {
      recommendations.push('심각한 성능 문제를 즉시 해결하세요');
    }

    const memoryIssues = result.diagnostics.performance.filter(m =>
      m.metric.includes('Memory') || m.metric.includes('Heap')
    ).filter(m => m.status !== 'ok');

    if (memoryIssues.length > 0) {
      recommendations.push('메모리 사용량을 최적화하세요 (메모리 누수 확인)');
    }
  }

  // 의존성 이슈 기반 권장사항
  if (result.diagnostics.dependencies) {
    const criticalDeps = result.diagnostics.dependencies.filter(d => d.severity === 'critical');
    if (criticalDeps.length > 0) {
      recommendations.push('중요한 의존성 문제를 즉시 해결하세요');
    }

    const vulnerableDeps = result.diagnostics.dependencies.filter(d => d.status === 'vulnerable');
    if (vulnerableDeps.length > 0) {
      recommendations.push('보안 취약점이 있는 패키지를 업데이트하세요');
    }
  }

  // 일반적인 권장사항
  if (result.summary.criticalIssues > 0) {
    recommendations.push('심각한 문제를 우선적으로 해결하세요');
  }

  if (result.summary.fixableIssues > 0) {
    recommendations.push('자동 수정 가능한 문제들을 --fix 옵션으로 해결하세요');
  }

  if (recommendations.length === 0) {
    recommendations.push('현재 감지된 주요 문제가 없습니다');
    recommendations.push('정기적인 의존성 업데이트와 코드 리뷰를 유지하세요');
  }

  return recommendations;
}

async function runDebugAnalysis(options: DebugAnalyzerOptions): Promise<DebugAnalysisResult> {
  const result: DebugAnalysisResult = {
    issues: [],
    diagnostics: {},
    recommendations: [],
    summary: {
      criticalIssues: 0,
      fixableIssues: 0,
      totalIssues: 0
    }
  };

  // 에러 메시지 분석
  if (options.error) {
    console.log(chalk.blue('🔍 에러 메시지 분석 중...'));
    const errorIssues = await analyzeErrorMessage(options.error);
    result.issues.push(...errorIssues);
  }

  // 로그 파일 분석
  if (options.logs) {
    console.log(chalk.blue('📄 로그 파일 분석 중...'));
    const logIssues = await parseLogFile(options.logs);
    result.issues.push(...logIssues);
  }

  // 시스템 진단
  if (options.system) {
    console.log(chalk.blue('💻 시스템 진단 중...'));
    result.diagnostics.system = await runSystemDiagnostics();
  }

  // 성능 분석
  if (options.performance) {
    console.log(chalk.blue('⚡ 성능 분석 중...'));
    result.diagnostics.performance = await analyzePerformance();
  }

  // 의존성 분석
  if (options.dependencies) {
    console.log(chalk.blue('📦 의존성 분석 중...'));
    result.diagnostics.dependencies = await analyzeDependencies();
  }

  // 기본 진단 (옵션이 없으면 모든 진단 실행)
  if (!options.error && !options.logs && !options.system && !options.performance && !options.dependencies) {
    console.log(chalk.blue('🔄 전체 진단 실행 중...'));
    result.diagnostics.system = await runSystemDiagnostics();
    result.diagnostics.performance = await analyzePerformance();
    result.diagnostics.dependencies = await analyzeDependencies();
  }

  // 권장사항 생성
  result.recommendations = await generateRecommendations(result);

  // 요약 통계 계산
  result.summary.totalIssues = result.issues.length;
  result.summary.criticalIssues = result.issues.filter(i => i.severity === 'critical').length;
  result.summary.fixableIssues = result.issues.filter(i => i.autofix).length;

  return result;
}

async function autoFixIssues(issues: DebugIssue[]): Promise<{ fixed: number; failed: number }> {
  let fixed = 0;
  let failed = 0;

  for (const issue of issues.filter(i => i.autofix && i.commands)) {
    try {
      console.log(chalk.yellow(`🔧 자동 수정 시도: ${issue.title}`));

      for (const command of issue.commands!) {
        const [cmd, ...args] = command.split(' ');
        await execa(cmd, args, { stdio: 'inherit' });
      }

      console.log(chalk.green(`✅ 수정 완료: ${issue.title}`));
      fixed++;
    } catch (error) {
      console.log(chalk.red(`❌ 수정 실패: ${issue.title}`));
      failed++;
    }
  }

  return { fixed, failed };
}

program
  .name('debug-analyzer')
  .description('MoAI 디버깅 분석 및 문제 진단')
  .option('-e, --error <message>', '분석할 에러 메시지')
  .option('-l, --logs <path>', '분석할 로그 파일 경로')
  .option('-s, --system', '시스템 진단 실행')
  .option('-p, --performance', '성능 분석 실행')
  .option('-d, --dependencies', '의존성 분석 실행')
  .option('-i, --interactive', '대화형 모드')
  .option('-f, --fix', '자동 수정 시도')
  .action(async (options: DebugAnalyzerOptions) => {
    try {
      console.log(chalk.blue('🔍 디버그 분석 시작...'));

      const result = await runDebugAnalysis(options);

      // 결과 출력
      if (result.issues.length > 0) {
        console.log(chalk.cyan('\n🐛 발견된 문제들:'));
        for (const issue of result.issues) {
          const severityColor = issue.severity === 'critical' ? chalk.red :
                               issue.severity === 'high' ? chalk.yellow :
                               issue.severity === 'medium' ? chalk.blue : chalk.gray;

          console.log(severityColor(`  ${issue.severity.toUpperCase()}: ${issue.title}`));
          console.log(chalk.gray(`    ${issue.description}`));
          console.log(chalk.cyan(`    해결방법: ${issue.solution}`));

          if (issue.commands) {
            console.log(chalk.gray(`    명령어: ${issue.commands.join(', ')}`));
          }
          console.log();
        }
      }

      // 시스템 진단 결과
      if (result.diagnostics.system) {
        console.log(chalk.cyan('\n💻 시스템 진단:'));
        const sys = result.diagnostics.system;
        console.log(`  Node.js: ${sys.nodejs.version} ${sys.nodejs.compatible ? '✅' : '❌'}`);
        console.log(`  Git: ${sys.git.version} ${sys.git.available ? '✅' : '❌'}`);
        console.log(`  패키지 매니저: ${sys.packageManager.type} ${sys.packageManager.version} ${sys.packageManager.available ? '✅' : '❌'}`);
        console.log(`  디스크: ${sys.disk.available} 사용가능 (${sys.disk.usage} 사용중)`);
        console.log(`  메모리: ${sys.memory.available} 사용가능 / ${sys.memory.total} 총량`);
      }

      // 성능 메트릭
      if (result.diagnostics.performance) {
        console.log(chalk.cyan('\n⚡ 성능 메트릭:'));
        for (const metric of result.diagnostics.performance) {
          const statusColor = metric.status === 'critical' ? chalk.red :
                             metric.status === 'warning' ? chalk.yellow : chalk.green;
          console.log(statusColor(`  ${metric.metric}: ${metric.current}${metric.unit} (임계값: ${metric.threshold}${metric.unit})`));
        }
      }

      // 의존성 이슈
      if (result.diagnostics.dependencies && result.diagnostics.dependencies.length > 0) {
        console.log(chalk.cyan('\n📦 의존성 문제:'));
        for (const dep of result.diagnostics.dependencies) {
          const severityColor = dep.severity === 'critical' ? chalk.red :
                               dep.severity === 'high' ? chalk.yellow : chalk.blue;
          console.log(severityColor(`  ${dep.name}: ${dep.status} (현재: ${dep.current || 'N/A'}, 필요: ${dep.required})`));
        }
      }

      // 권장사항
      if (result.recommendations.length > 0) {
        console.log(chalk.cyan('\n💡 권장사항:'));
        for (const rec of result.recommendations) {
          console.log(chalk.green(`  • ${rec}`));
        }
      }

      // 자동 수정 실행
      if (options.fix && result.summary.fixableIssues > 0) {
        console.log(chalk.blue('\n🔧 자동 수정 시도 중...'));
        const fixResult = await autoFixIssues(result.issues);
        console.log(chalk.green(`✅ ${fixResult.fixed}개 문제 수정 완료`));
        if (fixResult.failed > 0) {
          console.log(chalk.red(`❌ ${fixResult.failed}개 문제 수정 실패`));
        }
      }

      // JSON 출력
      console.log(JSON.stringify({
        success: result.summary.criticalIssues === 0,
        summary: result.summary,
        issues: result.issues.map(i => ({
          category: i.category,
          severity: i.severity,
          title: i.title,
          autofix: i.autofix
        })),
        diagnostics: {
          systemOk: result.diagnostics.system ?
            result.diagnostics.system.nodejs.compatible &&
            result.diagnostics.system.git.available : true,
          performanceOk: result.diagnostics.performance ?
            !result.diagnostics.performance.some(m => m.status === 'critical') : true,
          dependenciesOk: result.diagnostics.dependencies ?
            !result.diagnostics.dependencies.some(d => d.severity === 'critical') : true
        },
        nextSteps: result.summary.criticalIssues > 0 ? [
          '심각한 문제를 즉시 해결하세요',
          '자동 수정 가능한 문제는 --fix 옵션을 사용하세요'
        ] : [
          '주요 문제가 감지되지 않았습니다',
          '정기적인 시스템 점검을 유지하세요'
        ]
      }, null, 2));

      process.exit(result.summary.criticalIssues === 0 ? 0 : 1);
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