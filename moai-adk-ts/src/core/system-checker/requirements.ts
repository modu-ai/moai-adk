/**
 * @file System requirements definition and registry
 * @author MoAI Team
 * @tags @FEATURE:SYSTEM-REQUIREMENTS-001 @REQ:AUTO-VERIFY-012
 */

/**
 * System requirement definition interface
 * @tags @DESIGN:REQUIREMENT-INTERFACE-001
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
 * @tags @FEATURE:REQUIREMENT-REGISTRY-001
 */
export class RequirementRegistry {
  private readonly requirements: Map<string, SystemRequirement> = new Map();

  constructor() {
    this.initializeDefaultRequirements();
  }

  /**
   * Initialize default system requirements
   * @tags @TASK:DEFAULT-REQUIREMENTS-001
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

    // SQLite3 requirement (development)
    this.addRequirement({
      name: 'SQLite3',
      category: 'development',
      minVersion: '3.35.0',
      installCommands: {
        darwin: 'brew install sqlite3',
        linux: 'sudo apt-get install sqlite3',
        win32: 'winget install SQLite.SQLite',
      },
      checkCommand: 'sqlite3 --version',
      versionCommand: 'sqlite3 --version',
    });
  }

  /**
   * Add new requirement to registry
   * @param requirement - System requirement to add
   * @tags @API:ADD-REQUIREMENT-001
   */
  public addRequirement(requirement: SystemRequirement): void {
    this.requirements.set(requirement.name, requirement);
  }

  /**
   * Get specific requirement by name
   * @param name - Requirement name
   * @returns System requirement or undefined
   * @tags @API:GET-REQUIREMENT-001
   */
  public getRequirement(name: string): SystemRequirement | undefined {
    return this.requirements.get(name);
  }

  /**
   * Get requirements by category
   * @param category - Requirement category
   * @returns Array of matching requirements
   * @tags @API:GET-BY-CATEGORY-001
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
   * @tags @API:GET-ALL-REQUIREMENTS-001
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
