// @FEATURE:UTIL-007 | Chain: @REQ:UTIL-007 -> @DESIGN:UTIL-007 -> @TASK:UTIL-007 -> @TEST:UTIL-007
// Related: @DATA:I18N-MSG-001

/**
 * @file Internationalization system
 * @author MoAI Team
 */

export type Locale = 'en' | 'ko';

/**
 * Translation message structure
 */
export interface Messages {
  readonly common: {
    readonly success: string;
    readonly error: string;
    readonly warning: string;
    readonly info: string;
    readonly version: string;
  };
  readonly init: {
    readonly welcome: string;
    readonly selectLanguage: string;
    readonly languagePrompt: string;
    readonly projectName: string;
    readonly projectType: string;
    readonly creating: string;
    readonly completed: string;
    readonly failed: string;
    readonly prompts: {
      readonly projectInfo: string;
      readonly projectNameLabel: string;
      readonly projectNameTipNew: string;
      readonly projectNameTipCurrent: string;
      readonly devMode: string;
      readonly selectMode: string;
      readonly modePersonal: string;
      readonly modePersonalDesc: string;
      readonly modeTeam: string;
      readonly modeTeamDesc: string;
      readonly tipPersonal: string;
      readonly tipTeam: string;
      readonly versionControl: string;
      readonly initGit: string;
      readonly tipGitEnabled: string;
      readonly tipGitDisabled: string;
      readonly github: string;
      readonly useGithub: string;
      readonly tipGithubDisabled: string;
      readonly githubRepo: string;
      readonly githubUrl: string;
      readonly tipGithubUrl: string;
      readonly specWorkflow: string;
      readonly workflowBranch: string;
      readonly workflowBranchDesc: string;
      readonly workflowCommit: string;
      readonly workflowCommitDesc: string;
      readonly tipBranch: string;
      readonly tipCommit: string;
      readonly remoteSyn: string;
      readonly autoPush: string;
      readonly tipAutoPushEnabled: string;
      readonly tipAutoPushDisabled: string;
    };
  };
  readonly update: {
    readonly starting: string;
    readonly checking: string;
    readonly upToDate: string;
    readonly available: string;
    readonly analyzing: string;
    readonly backup: string;
    readonly applying: string;
    readonly completed: string;
    readonly failed: string;
    readonly filesChanged: string;
    readonly duration: string;
  };
  readonly doctor: {
    readonly checking: string;
    readonly allGood: string;
    readonly issuesFound: string;
    readonly fixSuggestion: string;
  };
}

/**
 * English translations
 */
const en: Messages = {
  common: {
    success: '✅ Success',
    error: '❌ Error',
    warning: '⚠️  Warning',
    info: 'ℹ️  Info',
    version: 'v{version}',
  },
  init: {
    welcome: '🗿 Welcome to MoAI-ADK',
    selectLanguage: 'Select your preferred language',
    languagePrompt: 'Choose CLI language:',
    projectName: 'Project name',
    projectType: 'Project type',
    creating: '🚀 Creating project...',
    completed: '✅ Project initialized successfully',
    failed: '❌ Project initialization failed',
    prompts: {
      projectInfo: 'Project Information',
      projectNameLabel: 'Project name:',
      projectNameTipNew: 'This will be used as the folder name and project identifier',
      projectNameTipCurrent: 'This will be used in configuration (current directory will NOT be renamed)',
      devMode: 'Development Mode',
      selectMode: 'Select mode:',
      modePersonal: '🧑 Personal',
      modePersonalDesc: 'Local development with .moai/specs/',
      modeTeam: '👥 Team',
      modeTeamDesc: 'GitHub Issues for SPEC management',
      tipPersonal: 'Personal mode: SPEC files stored locally, simpler workflow',
      tipTeam: 'Team mode: GitHub Issues for SPECs, PR-based workflow',
      versionControl: 'Version Control',
      initGit: 'Initialize local Git repository?',
      tipGitEnabled: 'Git will be initialized with initial commit',
      tipGitDisabled: 'You can initialize Git later with: git init',
      github: 'GitHub Integration',
      useGithub: 'Use GitHub for remote repository?',
      tipGithubDisabled: 'GitHub integration disabled - local Git only',
      githubRepo: 'GitHub Repository',
      githubUrl: 'GitHub repository URL:',
      tipGithubUrl: 'Example: https://github.com/username/project-name',
      specWorkflow: 'SPEC Workflow',
      workflowBranch: '🌿 Branch + Merge',
      workflowBranchDesc: 'GitHub PR workflow (recommended)',
      workflowCommit: '📝 Local commits',
      workflowCommitDesc: 'Direct commits to main',
      tipBranch: 'Branch workflow: feature/* branches + Pull Requests',
      tipCommit: 'Commit workflow: Direct commits to main branch',
      remoteSyn: 'Remote Synchronization',
      autoPush: 'Auto-push commits to remote repository?',
      tipAutoPushEnabled: 'Commits will be automatically pushed to GitHub',
      tipAutoPushDisabled: "You'll need to manually push with: git push",
    },
  },
  update: {
    starting: '🔄 Updating MoAI-ADK project...',
    checking: '🔍 Checking for updates...',
    upToDate: '✅ Project is up to date (v{version})',
    available: '⚡ Update available: v{from} → v{to}',
    analyzing: '📊 Analyzing {count} files...',
    backup: '💾 Creating backup...',
    applying: '🔧 Applying updates...',
    completed: '✅ Update completed successfully',
    failed: '❌ Update failed: {error}',
    filesChanged: '📝 {count} files updated',
    duration: '⏱️  Completed in {duration}s',
  },
  doctor: {
    checking: '🔍 Running system diagnostics...',
    allGood: '✅ All checks passed',
    issuesFound: '⚠️  {count} issue(s) found',
    fixSuggestion: '💡 Run with --fix to auto-repair',
  },
};

/**
 * Korean translations
 */
const ko: Messages = {
  common: {
    success: '✅ 성공',
    error: '❌ 오류',
    warning: '⚠️  경고',
    info: 'ℹ️  정보',
    version: 'v{version}',
  },
  init: {
    welcome: '🗿 MoAI-ADK에 오신 것을 환영합니다',
    selectLanguage: '사용할 언어를 선택하세요',
    languagePrompt: 'CLI 언어를 선택하세요:',
    projectName: '프로젝트 이름',
    projectType: '프로젝트 타입',
    creating: '🚀 프로젝트 생성 중...',
    completed: '✅ 프로젝트가 성공적으로 초기화되었습니다',
    failed: '❌ 프로젝트 초기화에 실패했습니다',
    prompts: {
      projectInfo: '프로젝트 정보',
      projectNameLabel: '프로젝트 이름:',
      projectNameTipNew: '폴더 이름과 프로젝트 식별자로 사용됩니다',
      projectNameTipCurrent: '설정에 사용될 이름입니다 (현재 디렉토리는 이름이 변경되지 않습니다)',
      devMode: '개발 모드',
      selectMode: '모드 선택:',
      modePersonal: '🧑 Personal',
      modePersonalDesc: '.moai/specs/를 사용한 로컬 개발',
      modeTeam: '👥 Team',
      modeTeamDesc: 'SPEC 관리를 위한 GitHub Issues',
      tipPersonal: 'Personal 모드: SPEC 파일이 로컬에 저장되며, 단순한 워크플로우',
      tipTeam: 'Team 모드: SPEC을 위한 GitHub Issues, PR 기반 워크플로우',
      versionControl: '버전 관리',
      initGit: '로컬 Git 저장소를 초기화하시겠습니까?',
      tipGitEnabled: 'Git이 초기 커밋과 함께 초기화됩니다',
      tipGitDisabled: '나중에 다음 명령어로 Git을 초기화할 수 있습니다: git init',
      github: 'GitHub 연동',
      useGithub: '원격 저장소로 GitHub를 사용하시겠습니까?',
      tipGithubDisabled: 'GitHub 연동 비활성화 - 로컬 Git만 사용',
      githubRepo: 'GitHub 저장소',
      githubUrl: 'GitHub 저장소 URL:',
      tipGithubUrl: '예시: https://github.com/username/project-name',
      specWorkflow: 'SPEC 워크플로우',
      workflowBranch: '🌿 브랜치 + 머지',
      workflowBranchDesc: 'GitHub PR 워크플로우 (권장)',
      workflowCommit: '📝 로컬 커밋',
      workflowCommitDesc: 'main에 직접 커밋',
      tipBranch: '브랜치 워크플로우: feature/* 브랜치 + Pull Request',
      tipCommit: '커밋 워크플로우: main 브랜치에 직접 커밋',
      remoteSyn: '원격 동기화',
      autoPush: '원격 저장소에 커밋을 자동으로 push하시겠습니까?',
      tipAutoPushEnabled: '커밋이 자동으로 GitHub에 push됩니다',
      tipAutoPushDisabled: 'git push 명령어로 수동으로 push해야 합니다',
    },
  },
  update: {
    starting: '🔄 MoAI-ADK 프로젝트 업데이트 중...',
    checking: '🔍 업데이트 확인 중...',
    upToDate: '✅ 프로젝트가 최신 상태입니다 (v{version})',
    available: '⚡ 업데이트 가능: v{from} → v{to}',
    analyzing: '📊 {count}개 파일 분석 중...',
    backup: '💾 백업 생성 중...',
    applying: '🔧 업데이트 적용 중...',
    completed: '✅ 업데이트가 완료되었습니다',
    failed: '❌ 업데이트 실패: {error}',
    filesChanged: '📝 {count}개 파일 업데이트됨',
    duration: '⏱️  {duration}초 소요',
  },
  doctor: {
    checking: '🔍 시스템 진단 실행 중...',
    allGood: '✅ 모든 검사 통과',
    issuesFound: '⚠️  {count}개 문제 발견',
    fixSuggestion: '💡 --fix 옵션으로 자동 수리 가능',
  },
};

/**
 * All available translations
 */
const translations: Record<Locale, Messages> = {
  en,
  ko,
};

/**
 * Current locale (default: ko)
 */
let currentLocale: Locale = 'ko';

/**
 * Set the current locale
 * @param locale - Locale to set
 */
export function setLocale(locale: Locale): void {
  if (!(locale in translations)) {
    throw new Error(`Unsupported locale: ${locale}`);
  }
  currentLocale = locale;
}

/**
 * Get the current locale
 * @returns Current locale
 */
export function getLocale(): Locale {
  return currentLocale;
}

/**
 * Get translation for a key with optional interpolation
 * @param key - Translation key (e.g., 'update.completed')
 * @param params - Parameters for interpolation
 * @returns Translated string
 */
export function t(key: string, params?: Record<string, string | number>): string {
  const keys = key.split('.');
  let value: any = translations[currentLocale];

  for (const k of keys) {
    if (value && typeof value === 'object' && k in value) {
      value = value[k];
    } else {
      return key; // Return key if translation not found
    }
  }

  if (typeof value !== 'string') {
    return key;
  }

  // Interpolate parameters
  if (params) {
    return value.replace(/\{(\w+)\}/g, (match, paramKey) => {
      return paramKey in params ? String(params[paramKey]) : match;
    });
  }

  return value;
}

/**
 * Initialize i18n system with locale from environment or config
 * @param locale - Optional locale to initialize with
 */
export function initI18n(locale?: Locale): void {
  if (locale) {
    setLocale(locale);
    return;
  }

  // Try to detect from environment
  const envLocale = process.env['MOAI_LOCALE'] || process.env['LANG'];
  if (envLocale) {
    const detected = envLocale.toLowerCase().startsWith('ko') ? 'ko' : 'en';
    setLocale(detected);
  }
}