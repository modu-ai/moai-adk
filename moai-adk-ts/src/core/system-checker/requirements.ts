/**
 * @FEATURE:SYSTEM-CHECKER-001 System requirements validation for MoAI-ADK
 *
 * Defines system requirements and platform-specific installation commands
 * for ensuring MoAI-ADK works properly across different environments.
 */

export interface SystemRequirement {
  name: string;
  required: boolean;
  minVersion?: string;
  installCommand?: Record<string, string>; // platform -> command
  checkCommand: string;
  description: string;
  category: 'runtime' | 'development' | 'optional';
  versionParser?: (output: string) => string | undefined;
}

export const SYSTEM_REQUIREMENTS: SystemRequirement[] = [
  // Runtime requirements (absolutely necessary)
  {
    name: 'Node.js',
    required: true,
    minVersion: '18.0.0',
    installCommand: {
      'darwin': 'brew install node',
      'linux': 'curl -fsSL https://deb.nodesource.com/setup_18.x | sudo -E bash - && sudo apt-get install -y nodejs',
      'win32': 'winget install OpenJS.NodeJS'
    },
    checkCommand: 'node --version',
    description: 'JavaScript runtime required for MoAI-ADK execution',
    category: 'runtime',
    versionParser: (output: string) => {
      const match = output.match(/v?(\d+\.\d+\.\d+)/);
      return match ? match[1] : undefined;
    }
  },
  {
    name: 'npm',
    required: true,
    minVersion: '9.0.0',
    installCommand: {
      'darwin': 'npm install -g npm@latest',
      'linux': 'npm install -g npm@latest',
      'win32': 'npm install -g npm@latest'
    },
    checkCommand: 'npm --version',
    description: 'Package manager for Node.js dependencies',
    category: 'runtime',
    versionParser: (output: string) => {
      const match = output.match(/(\d+\.\d+\.\d+)/);
      return match ? match[1] : undefined;
    }
  },

  // Development tools (required for full functionality)
  {
    name: 'Git',
    required: true,
    minVersion: '2.20.0',
    installCommand: {
      'darwin': 'brew install git',
      'linux': 'sudo apt-get update && sudo apt-get install git',
      'win32': 'winget install Git.Git'
    },
    checkCommand: 'git --version',
    description: 'Version control system for project management and workflow automation',
    category: 'development',
    versionParser: (output: string) => {
      const match = output.match(/git version (\d+\.\d+\.\d+)/);
      return match ? match[1] : undefined;
    }
  },
  {
    name: 'SQLite3',
    required: true,
    minVersion: '3.30.0',
    installCommand: {
      'darwin': 'brew install sqlite',
      'linux': 'sudo apt-get install sqlite3',
      'win32': 'winget install SQLite.SQLite'
    },
    checkCommand: 'sqlite3 --version',
    description: 'Database engine for TAG system and project metadata storage',
    category: 'development',
    versionParser: (output: string) => {
      const match = output.match(/(\d+\.\d+\.\d+)/);
      return match ? match[1] : undefined;
    }
  },

  // Claude Code integration (highly recommended)
  {
    name: 'Claude Code',
    required: true,
    minVersion: '1.0.0',
    installCommand: {
      'darwin': 'npm install -g @anthropic-ai/claude-code',
      'linux': 'npm install -g @anthropic-ai/claude-code',
      'win32': 'npm install -g @anthropic-ai/claude-code'
    },
    checkCommand: 'claude-code --version',
    description: 'Anthropic Claude Code IDE integration for enhanced development workflow',
    category: 'development',
    versionParser: (output: string) => {
      const match = output.match(/(\d+\.\d+\.\d+)/);
      return match ? match[1] : undefined;
    }
  },

  // Optional tools (enhances functionality but not required)
  {
    name: 'Python',
    required: false,
    minVersion: '3.10.0',
    installCommand: {
      'darwin': 'brew install python@3.11',
      'linux': 'sudo apt-get install python3.11',
      'win32': 'winget install Python.Python.3.11'
    },
    checkCommand: 'python3 --version',
    description: 'Python runtime for legacy Python-based projects and tools',
    category: 'optional',
    versionParser: (output: string) => {
      const match = output.match(/Python (\d+\.\d+\.\d+)/);
      return match ? match[1] : undefined;
    }
  },
  {
    name: 'Docker',
    required: false,
    minVersion: '20.0.0',
    installCommand: {
      'darwin': 'brew install --cask docker',
      'linux': 'curl -fsSL https://get.docker.com -o get-docker.sh && sh get-docker.sh',
      'win32': 'winget install Docker.DockerDesktop'
    },
    checkCommand: 'docker --version',
    description: 'Containerization platform for isolated development environments',
    category: 'optional',
    versionParser: (output: string) => {
      const match = output.match(/Docker version (\d+\.\d+\.\d+)/);
      return match ? match[1] : undefined;
    }
  },
  {
    name: 'VS Code',
    required: false,
    minVersion: '1.80.0',
    installCommand: {
      'darwin': 'brew install --cask visual-studio-code',
      'linux': 'sudo snap install --classic code',
      'win32': 'winget install Microsoft.VisualStudioCode'
    },
    checkCommand: 'code --version',
    description: 'Popular code editor with excellent TypeScript and Git support',
    category: 'optional',
    versionParser: (output: string) => {
      const lines = output.split('\n');
      const versionLine = lines[0];
      const match = versionLine.match(/(\d+\.\d+\.\d+)/);
      return match ? match[1] : undefined;
    }
  }
];

export function getRequiredSystemRequirements(): SystemRequirement[] {
  return SYSTEM_REQUIREMENTS.filter(req => req.required);
}

export function getOptionalSystemRequirements(): SystemRequirement[] {
  return SYSTEM_REQUIREMENTS.filter(req => !req.required);
}

export function getSystemRequirementsByCategory(category: SystemRequirement['category']): SystemRequirement[] {
  return SYSTEM_REQUIREMENTS.filter(req => req.category === category);
}