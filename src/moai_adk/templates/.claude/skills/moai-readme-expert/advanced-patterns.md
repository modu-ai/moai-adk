---
name: moai-readme-expert/advanced-patterns
description: Advanced README generation patterns, template systems, and documentation strategies
---

# Advanced README Generation Patterns (v4.0.0)

## Dynamic README Generation

### 1. Context-Aware README Builder

```typescript
interface ReadmeContext {
    projectName: string;
    projectType: 'library' | 'framework' | 'cli-tool' | 'web-app' | 'api';
    description: string;
    packageJson: PackageJson;
    gitHistory: GitCommit[];
    testCoverage: number;
    documentation: DocumentationFiles;
    contributors: Contributor[];
}

class DynamicReadmeGenerator {
    async generateReadme(context: ReadmeContext): Promise<string> {
        const sections: string[] = [];

        // 1. Title and badge section
        sections.push(this.generateTitle(context));
        sections.push(this.generateBadges(context));

        // 2. Description
        sections.push(this.generateDescription(context));

        // 3. Installation based on project type
        sections.push(await this.generateInstallation(context));

        // 4. Quick start / Usage examples
        sections.push(await this.generateQuickStart(context));

        // 5. Features (from package.json keywords)
        sections.push(this.generateFeatures(context));

        // 6. API Documentation
        if (context.projectType !== 'web-app') {
            sections.push(await this.generateAPIDocumentation(context));
        }

        // 7. Examples
        sections.push(await this.generateExamples(context));

        // 8. Configuration
        if (context.documentation.configFiles.length > 0) {
            sections.push(this.generateConfiguration(context));
        }

        // 9. Testing
        sections.push(this.generateTestingSection(context));

        // 10. Contributing
        sections.push(this.generateContributing(context));

        // 11. License
        sections.push(this.generateLicense(context));

        return sections.filter(s => s.length > 0).join('\n\n');
    }

    private generateTitle(context: ReadmeContext): string {
        return `# ${context.projectName}\n\n${context.description}`;
    }

    private generateBadges(context: ReadmeContext): string {
        const badges: string[] = [];

        // NPM/GitHub badges
        badges.push(`![npm version](https://img.shields.io/npm/v/${context.projectName})`);
        badges.push(`![license](https://img.shields.io/npm/l/${context.projectName})`);
        badges.push(`![npm downloads](https://img.shields.io/npm/dm/${context.projectName})`);

        // Test coverage badge
        if (context.testCoverage > 0) {
            const color = context.testCoverage >= 80 ? 'green' : context.testCoverage >= 60 ? 'yellow' : 'red';
            badges.push(
                `![coverage](https://img.shields.io/badge/coverage-${context.testCoverage}%25-${color})`
            );
        }

        return badges.join(' ');
    }

    private async generateInstallation(context: ReadmeContext): string {
        let installation = '## Installation\n\n';

        // Detect package manager
        const packageManager = this.detectPackageManager(context.packageJson);

        installation += `### ${packageManager.name}\n\n`;
        installation += `\`\`\`bash\n${packageManager.command} ${context.projectName}\n\`\`\`\n\n`;

        // Add yarn alternative
        if (packageManager.name !== 'yarn') {
            installation += `### Yarn\n\n`;
            installation += `\`\`\`bash\nyarn add ${context.projectName}\n\`\`\`\n`;
        }

        return installation;
    }

    private async generateQuickStart(context: ReadmeContext): Promise<string> {
        let quickStart = '## Quick Start\n\n';

        // Extract example from existing examples
        if (context.documentation.exampleFiles.length > 0) {
            const exampleFile = context.documentation.exampleFiles[0];
            const exampleCode = await fs.readFile(exampleFile, 'utf-8');

            quickStart += '```javascript\n';
            quickStart += exampleCode.split('\n').slice(0, 20).join('\n');
            quickStart += '\n```\n';
        } else {
            // Generate example based on project type
            quickStart += this.generateExampleForType(context.projectType);
        }

        return quickStart;
    }

    private generateFeatures(context: ReadmeContext): string {
        const features = context.packageJson.keywords || [];

        if (features.length === 0) return '';

        let featureSection = '## Features\n\n';
        featureSection += features.map(f => `- ${f}`).join('\n');

        return featureSection;
    }

    private async generateAPIDocumentation(context: ReadmeContext): Promise<string> {
        // Generate from JSDoc comments
        const apiDocs = await this.extractAPIFromJSDoc(context);

        let apiSection = '## API\n\n';
        for (const [methodName, documentation] of Object.entries(apiDocs)) {
            apiSection += `### ${methodName}\n\n`;
            apiSection += `${documentation.description}\n\n`;
            apiSection += `**Parameters:**\n\n`;
            for (const param of documentation.parameters) {
                apiSection += `- \`${param.name}\` (\`${param.type}\`): ${param.description}\n`;
            }
            apiSection += '\n';
        }

        return apiSection;
    }

    private async generateExamples(context: ReadmeContext): Promise<string> {
        const examples = await this.findExampleFiles(context);

        if (examples.length === 0) return '';

        let examplesSection = '## Examples\n\n';

        for (const example of examples.slice(0, 3)) {
            examplesSection += `### ${example.name}\n\n`;
            examplesSection += '```javascript\n';
            examplesSection += example.code.split('\n').slice(0, 15).join('\n');
            examplesSection += '\n```\n\n';
        }

        return examplesSection;
    }

    private generateTestingSection(context: ReadmeContext): string {
        let testSection = '## Testing\n\n';

        const testCommand = context.packageJson.scripts?.test || 'npm test';

        testSection += `\`\`\`bash\n${testCommand}\n\`\`\`\n\n`;

        if (context.testCoverage > 0) {
            testSection += `Test coverage: **${context.testCoverage}%**\n`;
        }

        return testSection;
    }
}
```

## Template System

### 1. Multi-Format README Templates

```typescript
interface ReadmeTemplate {
    name: string;
    projectType: string;
    sections: TemplateSection[];
    customization: Customization;
}

interface TemplateSection {
    title: string;
    content: string;
    optional: boolean;
    order: number;
}

class ReadmeTemplateEngine {
    private templates: Map<string, ReadmeTemplate> = new Map();

    registerTemplate(template: ReadmeTemplate) {
        this.templates.set(template.name, template);
    }

    async generateFromTemplate(
        templateName: string,
        data: ReadmeContext
    ): Promise<string> {
        const template = this.templates.get(templateName);
        if (!template) throw new Error(`Template not found: ${templateName}`);

        const sections: string[] = [];

        // Sort sections by order
        const sortedSections = template.sections.sort((a, b) => a.order - b.order);

        for (const section of sortedSections) {
            // Skip optional sections if data not available
            if (section.optional && !this.hasData(data, section.title)) {
                continue;
            }

            // Render section with data
            const rendered = this.renderSection(section, data);
            sections.push(rendered);
        }

        return sections.join('\n\n');
    }

    private renderSection(section: TemplateSection, data: ReadmeContext): string {
        // Template variable substitution
        let content = section.content;

        // Replace variables like {{projectName}}, {{description}}
        content = content.replace(/{{(\w+)}}/g, (match, key) => {
            return data[key as keyof ReadmeContext] || match;
        });

        return `# ${section.title}\n\n${content}`;
    }
}
```

## Table of Contents Generation

### 1. Intelligent TOC Builder

```typescript
class ReadmeTOCGenerator {
    generateTableOfContents(readmeContent: string): string {
        // Extract headings from markdown
        const headings = this.extractHeadings(readmeContent);

        // Create anchor links
        let toc = '## Table of Contents\n\n';

        for (const heading of headings) {
            const indent = '  '.repeat(heading.level - 1);
            const anchor = this.generateAnchor(heading.text);
            toc += `${indent}- [${heading.text}](#${anchor})\n`;
        }

        return toc;
    }

    private extractHeadings(content: string): Heading[] {
        const headings: Heading[] = [];
        const lines = content.split('\n');

        for (const line of lines) {
            const match = line.match(/^(#+)\s+(.+)$/);
            if (match) {
                headings.push({
                    level: match[1].length,
                    text: match[2],
                    anchor: this.generateAnchor(match[2])
                });
            }
        }

        return headings;
    }

    private generateAnchor(text: string): string {
        return text
            .toLowerCase()
            .replace(/[^\w\s-]/g, '')
            .replace(/\s+/g, '-');
    }
}
```

---

**Version**: 4.0.0 | **Last Updated**: 2025-11-22 | **Enterprise Ready**: ✓
