// @CODE:SYS-REQ-001 | 
// Related: @CODE:SYS-001

/**
 * @file System requirements definition
 * @author MoAI Team
 */

/**
 * System requirement definition interface
 * @tags @SPEC:REQUIREMENT-INTERFACE-001
 */
export interface SystemRequirement {
  readonly name: string;
  readonly category: 'runtime' | 'development' | 'optional';
  readonly minVersion?: string;
  readonly installCommands: Record<string, string>;
  readonly checkCommand: string;
  readonly versionCommand?: string;
}

/**
 * System requirements registry for centralized management
 * @tags @CODE:REQUIREMENT-REGISTRY-001
 */
export class RequirementRegistry {
  private readonly requirements: Map<string, SystemRequirement> = new Map();

  constructor() {
    this.initializeDefaultRequirements();
  }

  /**
   * Add language-specific development requirements
   * @param language - Programming language detected
   * @tags @CODE:ADD-LANGUAGE-REQUIREMENTS-001:API
   */
  public addLanguageRequirements(language: string): void {
    switch (language.toLowerCase()) {
      case 'typescript':
      case 'javascript':
        this.addIfNotExists({
          name: 'TypeScript',
          category: 'development',
          minVersion: '5.0.0',
          installCommands: {
            darwin: 'npm install -g typescript',
            linux: 'npm install -g typescript',
            win32: 'npm install -g typescript',
          },
          checkCommand: 'tsc --version',
          versionCommand: 'tsc --version',
        });
        break;

      case 'python':
        this.addIfNotExists({
          name: 'Python',
          category: 'development',
          minVersion: '3.8.0',
          installCommands: {
            darwin: 'brew install python@3.11',
            linux: 'sudo apt-get install python3.11',
            win32: 'winget install Python.Python.3.11',
          },
          checkCommand: 'python3 --version',
          versionCommand: 'python3 --version',
        });
        break;

      case 'java':
        this.addIfNotExists({
          name: 'Java',
          category: 'development',
          minVersion: '17.0.0',
          installCommands: {
            darwin: 'brew install openjdk@17',
            linux: 'sudo apt-get install openjdk-17-jdk',
            win32: 'winget install Eclipse.Temurin.17.JDK',
          },
          checkCommand: 'java --version',
          versionCommand: 'java --version',
        });
        break;

      case 'go':
        this.addIfNotExists({
          name: 'Go',
          category: 'development',
          minVersion: '1.21.0',
          installCommands: {
            darwin: 'brew install go',
            linux: 'sudo apt-get install golang-go',
            win32: 'winget install GoLang.Go',
          },
          checkCommand: 'go version',
          versionCommand: 'go version',
        });
        break;
    }
  }

  /**
   * Add requirement only if it doesn't already exist
   * @param requirement - System requirement to add
   * @tags @UTIL:ADD-IF-NOT-EXISTS-001
   */
  private addIfNotExists(requirement: SystemRequirement): void {
    if (!this.requirements.has(requirement.name)) {
      this.addRequirement(requirement);
    }
  }

  /**
   * Initialize default system requirements
   * @tags @CODE:DEFAULT-REQUIREMENTS-001
   */
  private initializeDefaultRequirements(): void {
    // Git requirement
    this.addRequirement({
      name: 'Git',
      category: 'runtime',
      minVersion: '2.30.0',
      installCommands: {
        darwin: 'brew install git',
        linux: 'sudo apt-get install git',
        win32: 'winget install Git.Git',
      },
      checkCommand: 'git --version',
      versionCommand: 'git --version',
    });

    // Node.js requirement
    this.addRequirement({
      name: 'Node.js',
      category: 'runtime',
      minVersion: '18.0.0',
      installCommands: {
        darwin: 'brew install node',
        linux:
          'curl -fsSL https://deb.nodesource.com/setup_18.x | sudo -E bash - && sudo apt-get install -y nodejs',
        win32: 'winget install OpenJS.NodeJS',
      },
      checkCommand: 'node --version',
      versionCommand: 'node --version',
    });

    // npm requirement (development)
    this.addRequirement({
      name: 'npm',
      category: 'development',
      minVersion: '8.0.0',
      installCommands: {
        darwin: 'brew install npm',
        linux: 'sudo apt-get install npm',
        win32: 'winget install OpenJS.NodeJS', // npm comes with Node.js
      },
      checkCommand: 'npm --version',
      versionCommand: 'npm --version',
    });

    // Git LFS requirement (optional)
    this.addRequirement({
      name: 'Git LFS',
      category: 'optional',
      minVersion: '3.0.0',
      installCommands: {
        darwin: 'brew install git-lfs',
        linux: 'sudo apt-get install git-lfs',
        win32: 'winget install GitHub.GitLFS',
      },
      checkCommand: 'git lfs version',
      versionCommand: 'git lfs version',
    });
  }

  /**
   * Add new requirement to registry
   * @param requirement - System requirement to add
   * @tags @CODE:ADD-REQUIREMENT-001:API
   */
  public addRequirement(requirement: SystemRequirement): void {
    this.requirements.set(requirement.name, requirement);
  }

  /**
   * Get specific requirement by name
   * @param name - Requirement name
   * @returns System requirement or undefined
   * @tags @CODE:GET-REQUIREMENT-001:API
   */
  public getRequirement(name: string): SystemRequirement | undefined {
    return this.requirements.get(name);
  }

  /**
   * Get requirements by category
   * @param category - Requirement category
   * @returns Array of matching requirements
   * @tags @CODE:GET-BY-CATEGORY-001:API
   */
  public getByCategory(
    category: SystemRequirement['category']
  ): SystemRequirement[] {
    return Array.from(this.requirements.values()).filter(
      req => req.category === category
    );
  }

  /**
   * Get all requirements
   * @returns Array of all requirements
   * @tags @CODE:GET-ALL-REQUIREMENTS-001:API
   */
  public getAllRequirements(): SystemRequirement[] {
    return Array.from(this.requirements.values());
  }
}

/**
 * Global requirement registry instance
 * @tags @SINGLETON:REQUIREMENT-REGISTRY-001
 */
export const requirementRegistry = new RequirementRegistry();
