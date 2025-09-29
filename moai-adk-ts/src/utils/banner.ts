/**
 * @file MoAI-ADK Banner Module
 * @author MoAI Team
 * @tags @FEATURE:BANNER-001 @REQ:CLI-UX-001
 */

import chalk from 'chalk';
import { getCurrentVersion } from './version';

/**
 * Check if terminal supports color output
 * @returns True if colors are supported
 * @tags @UTIL:COLOR-SUPPORT-001
 */
function supportsColor(): boolean {
  return (
    process.stdout.isTTY &&
    process.env.TERM !== 'dumb' &&
    process.env.NO_COLOR === undefined
  );
}

/**
 * Apply Claude AI brand color to ASCII art line
 * @param line - Line to colorize
 * @returns Colorized line
 * @tags @UTIL:COLORIZE-001
 */
function applyClaudeBrandColor(line: string): string {
  if (!supportsColor() || !line.trim()) {
    return line;
  }

  // Claude AI Official Brand Color - #da7756 (RGB: 218, 119, 86)
  return chalk.rgb(218, 119, 86)(line);
}

/**
 * Get MoAI logo ASCII art
 * @returns Array of logo lines
 * @tags @UTIL:LOGO-ASCII-001
 */
function getMoaiLogo(): string[] {
  return [
    '███╗   ███╗          █████╗ ██╗      █████╗ ██████╗ ██╗  ██╗',
    '████╗ ████║ ██████╗ ██╔══██╗██║     ██╔══██╗██╔══██╗██║ ██╔╝',
    '██╔████╔██║██╔═══██ ███████║██║ ███ ███████║██║  ██║█████╔╝ ',
    '██║╚██╔╝██║██║   ██║██╔══██║██║ ╚══ ██╔══██║██║  ██║██╔═██╗ ',
    '██║ ╚═╝ ██║╚██████╔╝██║  ██║██║     ██║  ██║██████╔╝██║  ██╗',
    '╚═╝     ╚═╝ ╚═════╝ ╚═╝  ╚═╝╚═╝     ╚═╝  ╚═╝╚═════╝ ╚═╝  ╚═╝',
  ];
}

/**
 * Create the complete MoAI banner
 * @param options Banner options
 * @returns Complete banner string
 * @tags @API:CREATE-BANNER-001
 */
export function createBanner(
  options: { version?: string; showUsage?: boolean } = {}
): string {
  const { version = getCurrentVersion(), showUsage = false } = options;

  const moaiLines = getMoaiLogo();
  const bannerLines: string[] = [];

  // Empty line at top
  bannerLines.push('');

  // MoAI-ADK logo with Claude brand color
  for (const line of moaiLines) {
    const coloredLine = applyClaudeBrandColor(line);
    bannerLines.push(coloredLine);
  }

  // Bottom border matching logo width (60 chars for MoAI-ADK)
  const border = supportsColor()
    ? applyClaudeBrandColor('═'.repeat(60))
    : '═'.repeat(60);
  bannerLines.push(border);
  bannerLines.push('');

  // Description with version in one line
  const description = `🗿 MoAI-ADK: Modu-AI's Agentic Development kit (v${version}) 🚀`;

  bannerLines.push(
    supportsColor() ? applyClaudeBrandColor(description) : description
  );
  bannerLines.push('');

  // Usage info if requested
  if (showUsage) {
    bannerLines.push('');
    bannerLines.push('Usage: moai [options] [command]');
    bannerLines.push('');
    bannerLines.push('Commands:');
    bannerLines.push(
      '  doctor [options]                 Run system diagnostics'
    );
    bannerLines.push(
      '  init [options] [project]         Initialize a new MoAI-ADK project'
    );
    bannerLines.push(
      '  restore [options] <backup-path>  Restore MoAI-ADK from a backup directory'
    );
    bannerLines.push(
      '  status [options]                 Show MoAI-ADK project status'
    );
    bannerLines.push(
      '  update [options]                 Update MoAI-ADK to the latest version'
    );
    bannerLines.push(
      '  help [command]                   Show help for MoAI-ADK commands'
    );
  }

  // Footer
  bannerLines.push('');
  const footer = 'copyleft 2024, Modu-AI / 모두의AI (https://mo.ai.kr)';
  bannerLines.push(supportsColor() ? chalk.gray(footer) : footer);
  bannerLines.push('');

  return bannerLines.join('\n');
}

/**
 * Print the MoAI-ADK banner to stdout
 * @param options Banner options
 * @tags @API:PRINT-BANNER-001
 */
export function printBanner(
  options: { version?: string; showUsage?: boolean } = {}
): void {
  console.log(createBanner(options));
}

/**
 * Create a simple header for commands
 * @param title Command title
 * @param subtitle Optional subtitle
 * @returns Formatted header
 * @tags @API:CREATE-HEADER-001
 */
export function createHeader(title: string, subtitle?: string): string {
  const lines: string[] = [];

  const formattedTitle = supportsColor()
    ? chalk.blue.bold(`🚀 ${title}`)
    : `🚀 ${title}`;

  lines.push(formattedTitle);

  if (subtitle) {
    const formattedSubtitle = supportsColor() ? chalk.blue(subtitle) : subtitle;
    lines.push(formattedSubtitle);
  }

  lines.push('');

  return lines.join('\n');
}
