// @CODE:UTIL-007 |
// Related: @CODE:I18N-MSG-001

/**
 * @file Internationalization system
 * @author MoAI Team
 */

export type Locale = 'en' | 'ko' | 'ja' | 'zh';

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
      projectNameTipNew:
        'This will be used as the folder name and project identifier',
      projectNameTipCurrent:
        'This will be used in configuration (current directory will NOT be renamed)',
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
      projectNameTipCurrent:
        '설정에 사용될 이름입니다 (현재 디렉토리는 이름이 변경되지 않습니다)',
      devMode: '개발 모드',
      selectMode: '모드 선택:',
      modePersonal: '🧑 Personal',
      modePersonalDesc: '.moai/specs/를 사용한 로컬 개발',
      modeTeam: '👥 Team',
      modeTeamDesc: 'SPEC 관리를 위한 GitHub Issues',
      tipPersonal:
        'Personal 모드: SPEC 파일이 로컬에 저장되며, 단순한 워크플로우',
      tipTeam: 'Team 모드: SPEC을 위한 GitHub Issues, PR 기반 워크플로우',
      versionControl: '버전 관리',
      initGit: '로컬 Git 저장소를 초기화하시겠습니까?',
      tipGitEnabled: 'Git이 초기 커밋과 함께 초기화됩니다',
      tipGitDisabled:
        '나중에 다음 명령어로 Git을 초기화할 수 있습니다: git init',
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
 * Japanese translations
 */
const ja: Messages = {
  common: {
    success: '✅ 成功',
    error: '❌ エラー',
    warning: '⚠️  警告',
    info: 'ℹ️  情報',
    version: 'v{version}',
  },
  init: {
    welcome: '🗿 MoAI-ADKへようこそ',
    selectLanguage: '言語を選択してください',
    languagePrompt: 'CLI言語を選択:',
    projectName: 'プロジェクト名',
    projectType: 'プロジェクトタイプ',
    creating: '🚀 プロジェクトを作成中...',
    completed: '✅ プロジェクトが正常に初期化されました',
    failed: '❌ プロジェクトの初期化に失敗しました',
    prompts: {
      projectInfo: 'プロジェクト情報',
      projectNameLabel: 'プロジェクト名:',
      projectNameTipNew: 'フォルダ名とプロジェクト識別子として使用されます',
      projectNameTipCurrent:
        '設定で使用される名前です（現在のディレクトリは名前が変更されません）',
      devMode: '開発モード',
      selectMode: 'モードを選択:',
      modePersonal: '🧑 Personal',
      modePersonalDesc: '.moai/specs/を使用したローカル開発',
      modeTeam: '👥 Team',
      modeTeamDesc: 'SPEC管理用GitHub Issues',
      tipPersonal:
        'Personalモード: SPECファイルがローカルに保存され、シンプルなワークフロー',
      tipTeam: 'Teamモード: SPEC用GitHub Issues、PRベースのワークフロー',
      versionControl: 'バージョン管理',
      initGit: 'ローカルGitリポジトリを初期化しますか？',
      tipGitEnabled: 'Gitが初期コミットとともに初期化されます',
      tipGitDisabled: '後で次のコマンドでGitを初期化できます: git init',
      github: 'GitHub連携',
      useGithub: 'リモートリポジトリにGitHubを使用しますか？',
      tipGithubDisabled: 'GitHub連携無効 - ローカルGitのみ使用',
      githubRepo: 'GitHubリポジトリ',
      githubUrl: 'GitHubリポジトリURL:',
      tipGithubUrl: '例: https://github.com/username/project-name',
      specWorkflow: 'SPECワークフロー',
      workflowBranch: '🌿 ブランチ + マージ',
      workflowBranchDesc: 'GitHub PRワークフロー（推奨）',
      workflowCommit: '📝 ローカルコミット',
      workflowCommitDesc: 'mainに直接コミット',
      tipBranch: 'ブランチワークフロー: feature/*ブランチ + Pull Request',
      tipCommit: 'コミットワークフロー: mainブランチに直接コミット',
      remoteSyn: 'リモート同期',
      autoPush: 'リモートリポジトリに自動pushしますか？',
      tipAutoPushEnabled: 'コミットが自動的にGitHubにpushされます',
      tipAutoPushDisabled: 'git pushコマンドで手動でpushする必要があります',
    },
  },
  update: {
    starting: '🔄 MoAI-ADKプロジェクトを更新中...',
    checking: '🔍 更新を確認中...',
    upToDate: '✅ プロジェクトは最新です (v{version})',
    available: '⚡ 更新可能: v{from} → v{to}',
    analyzing: '📊 {count}ファイルを分析中...',
    backup: '💾 バックアップを作成中...',
    applying: '🔧 更新を適用中...',
    completed: '✅ 更新が完了しました',
    failed: '❌ 更新に失敗しました: {error}',
    filesChanged: '📝 {count}ファイルが更新されました',
    duration: '⏱️  {duration}秒で完了',
  },
  doctor: {
    checking: '🔍 システム診断を実行中...',
    allGood: '✅ すべてのチェックに合格',
    issuesFound: '⚠️  {count}件の問題が見つかりました',
    fixSuggestion: '💡 --fixオプションで自動修復可能',
  },
};

/**
 * Chinese translations
 */
const zh: Messages = {
  common: {
    success: '✅ 成功',
    error: '❌ 错误',
    warning: '⚠️  警告',
    info: 'ℹ️  信息',
    version: 'v{version}',
  },
  init: {
    welcome: '🗿 欢迎使用MoAI-ADK',
    selectLanguage: '选择您的首选语言',
    languagePrompt: '选择CLI语言:',
    projectName: '项目名称',
    projectType: '项目类型',
    creating: '🚀 正在创建项目...',
    completed: '✅ 项目初始化成功',
    failed: '❌ 项目初始化失败',
    prompts: {
      projectInfo: '项目信息',
      projectNameLabel: '项目名称:',
      projectNameTipNew: '将用作文件夹名称和项目标识符',
      projectNameTipCurrent: '将在配置中使用（当前目录不会重命名）',
      devMode: '开发模式',
      selectMode: '选择模式:',
      modePersonal: '🧑 Personal',
      modePersonalDesc: '使用.moai/specs/进行本地开发',
      modeTeam: '👥 Team',
      modeTeamDesc: '使用GitHub Issues进行SPEC管理',
      tipPersonal: 'Personal模式: SPEC文件存储在本地，工作流程更简单',
      tipTeam: 'Team模式: 使用GitHub Issues管理SPEC，基于PR的工作流程',
      versionControl: '版本控制',
      initGit: '初始化本地Git仓库？',
      tipGitEnabled: 'Git将随初始提交一起初始化',
      tipGitDisabled: '您可以稍后使用以下命令初始化Git: git init',
      github: 'GitHub集成',
      useGithub: '使用GitHub作为远程仓库？',
      tipGithubDisabled: 'GitHub集成已禁用 - 仅使用本地Git',
      githubRepo: 'GitHub仓库',
      githubUrl: 'GitHub仓库URL:',
      tipGithubUrl: '示例: https://github.com/username/project-name',
      specWorkflow: 'SPEC工作流程',
      workflowBranch: '🌿 分支 + 合并',
      workflowBranchDesc: 'GitHub PR工作流程（推荐）',
      workflowCommit: '📝 本地提交',
      workflowCommitDesc: '直接提交到main',
      tipBranch: '分支工作流程: feature/*分支 + Pull Request',
      tipCommit: '提交工作流程: 直接提交到main分支',
      remoteSyn: '远程同步',
      autoPush: '自动推送提交到远程仓库？',
      tipAutoPushEnabled: '提交将自动推送到GitHub',
      tipAutoPushDisabled: '您需要使用git push手动推送',
    },
  },
  update: {
    starting: '🔄 正在更新MoAI-ADK项目...',
    checking: '🔍 正在检查更新...',
    upToDate: '✅ 项目已是最新版本 (v{version})',
    available: '⚡ 可用更新: v{from} → v{to}',
    analyzing: '📊 正在分析{count}个文件...',
    backup: '💾 正在创建备份...',
    applying: '🔧 正在应用更新...',
    completed: '✅ 更新完成',
    failed: '❌ 更新失败: {error}',
    filesChanged: '📝 已更新{count}个文件',
    duration: '⏱️  耗时{duration}秒',
  },
  doctor: {
    checking: '🔍 正在运行系统诊断...',
    allGood: '✅ 所有检查通过',
    issuesFound: '⚠️  发现{count}个问题',
    fixSuggestion: '💡 使用--fix选项自动修复',
  },
};

/**
 * All available translations
 */
const translations: Record<Locale, Messages> = {
  en,
  ko,
  ja,
  zh,
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
export function t(
  key: string,
  params?: Record<string, string | number>
): string {
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
  const envLocale = process.env.MOAI_LOCALE || process.env.LANG;
  if (envLocale) {
    const detected = envLocale.toLowerCase().startsWith('ko') ? 'ko' : 'en';
    setLocale(detected);
  }
}
